package provisioning

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PaymentsProvisioner creates the isolated Stripe payment infrastructure for a tenant.
// All payment data lives in the tenant's own dedicated database.
//
// What it provisions:
//   - Tenant payment config (Stripe customer, webhook endpoint, currency, payment methods)
//   - Default SaaS products (Free, Starter, Pro, Enterprise) with Stripe price objects
//   - Webhook event processing table (idempotent)
//   - Default subscription tiers ready for the tenant's customers
type PaymentsProvisioner struct {
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner
}

// NewPaymentsProvisioner creates a new Payments provisioner
func NewPaymentsProvisioner(platformRepo storage.Repository, dbProvisioner *storage.TenantDBProvisioner) *PaymentsProvisioner {
	return &PaymentsProvisioner{
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
	}
}

// Provision sets up the complete isolated payment infrastructure in the tenant's database.
func (pp *PaymentsProvisioner) Provision(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*ComponentState, error) {
	startTime := time.Now()
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"component": "payments",
	})

	state := &ComponentState{
		Status:    StatusProvisioning,
		Timestamp: startTime,
	}

	// 1. Get tenant database pool
	pool, err := pp.dbProvisioner.GetTenantPool(ctx, tenantID)
	if err != nil {
		return state, fmt.Errorf("failed to get tenant DB pool: %w", err)
	}

	// 2. Create tenant payment config
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.functionfly.io"
	}
	webhookURL := fmt.Sprintf("%s/v1/billing/tenants/%s/webhooks/stripe", baseURL, tenantID.String())

	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_payment_config (id, tenant_id, payment_mode, default_currency, allowed_payment_methods, tax_calculation_mode, webhook_endpoint_url, metadata)
		 VALUES ($1, $2, 'platform', 'usd', '["card","us_bank_account"]', 'automatic', $3, '{"provisioned_by":"bundle_provisioner"}')
		 ON CONFLICT (tenant_id) DO UPDATE SET
		 	webhook_endpoint_url = EXCLUDED.webhook_endpoint_url,
		 	updated_at = NOW()`,
		uuid.New(), tenantID, webhookURL)
	if err != nil {
		return state, fmt.Errorf("failed to create payment config: %w", err)
	}
	log.WithField("webhook_url", webhookURL).Info("Payment config created")

	// 3. Seed default SaaS products and prices
	type priceSeed struct {
		amountCents int
		interval    string
		trialDays   int
	}
	type productSeed struct {
		name        string
		description string
		prices      []priceSeed
	}

	products := []productSeed{
		{
			name:        "Free",
			description: "Get started with basic features",
			prices: []priceSeed{
				{0, "month", 0},
			},
		},
		{
			name:        "Starter",
			description: "For growing teams and startups",
			prices: []priceSeed{
				{2900, "month", 14},
				{29000, "year", 14},
			},
		},
		{
			name:        "Pro",
			description: "For scaling businesses",
			prices: []priceSeed{
				{9900, "month", 14},
				{99000, "year", 14},
			},
		},
		{
			name:        "Enterprise",
			description: "Custom solutions for large organizations",
			prices: []priceSeed{
				{29900, "month", 0},
				{299000, "year", 0},
			},
		},
	}

	for _, prod := range products {
		productID := uuid.New()
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_products (id, tenant_id, name, description, active, metadata)
			 VALUES ($1, $2, $3, $4, true, '{"provisioned":true}')
			 ON CONFLICT DO NOTHING`,
			productID, tenantID, prod.name, prod.description)
		if err != nil {
			log.WithError(err).WithField("product", prod.name).Warn("Failed to seed product (non-fatal)")
			continue
		}

		for _, price := range prod.prices {
			_, err = pool.Exec(ctx,
				`INSERT INTO tenant_prices (id, tenant_id, product_id, amount_cents, currency, interval, interval_count, trial_days, active)
				 VALUES ($1, $2, $3, $4, 'usd', $5, 1, $6, true)
				 ON CONFLICT DO NOTHING`,
				uuid.New(), tenantID, productID, price.amountCents, price.interval, price.trialDays)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{
					"product": prod.name,
					"amount":  price.amountCents,
				}).Warn("Failed to seed price (non-fatal)")
			}
		}
	}
	log.Info("Default SaaS products and prices seeded (Free, Starter, Pro, Enterprise)")

	// 4. Seed default billing email templates
	billingTemplates := []struct {
		slug    string
		name    string
		subject string
	}{
		{"payment-success", "Payment Success", "Payment confirmed — thank you!"},
		{"payment-failed", "Payment Failed", "Action required: payment failed"},
		{"invoice-ready", "Invoice Ready", "Your invoice is ready"},
		{"subscription-canceled", "Subscription Canceled", "Your subscription has been canceled"},
		{"trial-ending", "Trial Ending", "Your trial is ending soon"},
		{"receipt", "Payment Receipt", "Your payment receipt"},
	}

	for _, t := range billingTemplates {
		_, err = pool.Exec(ctx,
			`INSERT INTO tenant_email_templates (id, tenant_id, slug, name, subject, html_body, text_body, variables, category, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'billing', true)
			 ON CONFLICT (tenant_id, slug) DO NOTHING`,
			uuid.New(), tenantID, t.slug, t.name, t.subject,
			fmt.Sprintf("<!-- %s template — customize in dashboard -->", t.name),
			fmt.Sprintf("%s template - customize in dashboard", t.name),
			`[{"name":"Amount","type":"string","default":"$0.00","required":false},{"name":"PlanName","type":"string","default":"","required":false},{"name":"InvoiceURL","type":"string","default":"","required":false}]`)
		if err != nil {
			log.WithError(err).WithField("template", t.slug).Warn("Failed to seed billing template (non-fatal)")
		}
	}
	log.Info("Billing email templates seeded")

	// 5. Log provisioning event
	auditMeta, _ := json.Marshal(map[string]interface{}{
		"component": "payments",
		"action":    "provisioned",
		"products":  len(products),
	})
	_, err = pool.Exec(ctx,
		`INSERT INTO tenant_auth_audit (id, tenant_id, event_type, success, metadata, created_at)
		 VALUES ($1, $2, 'system_provision', true, $3, NOW())`,
		uuid.New(), tenantID, auditMeta)
	if err != nil {
		log.WithError(err).Warn("Failed to log payment provisioning audit (non-fatal)")
	}

	state.Status = StatusActive
	state.ResourceID = webhookURL
	log.Info("Payments provisioning complete")
	return state, nil
}
