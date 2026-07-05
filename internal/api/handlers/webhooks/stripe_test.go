package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
)

func TestMain(m *testing.M) {
	os.Setenv("DEVELOPMENT", "true")
	os.Setenv("WALLET_AUDIT_HMAC_KEY", "test-key-for-testing")
	os.Setenv("WALLET_ENCRYPTION_KEY", "01234567890123456789012345678901")
	os.Setenv("ALLOW_UNVERIFIED_WEBHOOKS", "true")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "functionfly")
	os.Setenv("DB_SSLMODE", "disable")
	os.Exit(m.Run())
}

type mockRepo struct {
	storage.Repository
	bundleSlug string
	bundleID   uuid.UUID
	db        *storage.PostgresDB
}

func (m *mockRepo) GetPricingBundleBySlug(ctx context.Context, slug string) (*storage.PricingBundle, error) {
	return &storage.PricingBundle{
		ID:                m.bundleID,
		Slug:             m.bundleSlug,
		Name:             "SaaS Starter",
		DisplayName:      "SaaS Starter Bundle",
		DisplayPriceCents: 9900,
		IsActive:         true,
	}, nil
}

func (m *mockRepo) GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*storage.BundleSubscription, error) {
	return nil, nil
}

func (m *mockRepo) CreateBundleSubscription(ctx context.Context, sub *storage.BundleSubscription) error {
	if m.db == nil {
		return nil
	}
	_, err := m.db.Exec(`INSERT INTO bundle_subscriptions (id, tenant_id, bundle_id, status, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		sub.ID, sub.TenantID, sub.BundleID, sub.Status)
	return err
}

func (m *mockRepo) UpdateBundleSubscription(ctx context.Context, sub *storage.BundleSubscription) error {
	return nil
}

func (m *mockRepo) SetTenantDegradedMode(ctx context.Context, tenantID uuid.UUID, degraded bool, reason string) error {
	return nil
}

func (m *mockRepo) CreatePaidInvoiceForStripeCheckoutSession(ctx context.Context, tenantID uuid.UUID, amountCents int, currency, checkoutSessionID, receiptURL string) error {
	if m.db == nil {
		return nil
	}
	_, err := m.db.Exec(`INSERT INTO invoices (id, tenant_id, status, amount_due_cents, amount_paid_cents, currency, external_reference, stripe_invoice_id, hosted_invoice_url, invoice_pdf_url, paid_at, created_at, updated_at) VALUES ($1, $2, 'paid', $3, $3, $4, $5, $5, $6, $6, NOW(), NOW(), NOW())`,
		uuid.New(), tenantID, amountCents, currency, checkoutSessionID, receiptURL)
	return err
}

func TestBundleCheckout_CreatesInvoice(t *testing.T) {
	db, err := storage.NewPostgresDB()
	require.NoError(t, err)
	defer db.Close()

	tenantID := uuid.New()
	bundleID := uuid.New()
	bundleSlug := "saas-starter-" + uuid.New().String()[:8]
	sessionID := "cs_bundle_invoice_" + uuid.New().String()

	_, err = db.Exec("INSERT INTO tenants (id, name, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())", tenantID, "Test Tenant")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO pricing_bundles (id, slug, name, display_name, description, short_description, display_price_cents, billing_interval, stripe_price_id, icon, color, features_included, feature_limits, provisioning_templates, sort_order, is_active, is_popular, created_at, updated_at) VALUES ($1, $2, 'SaaS Starter', 'SaaS Starter Bundle', 'Description', 'Short', 9900, 'monthly', 'price_123', 'icon', 'color', '[]', '{}', '[]', 1, true, false, NOW(), NOW())", bundleID, bundleSlug)
	require.NoError(t, err)

	repo := db.Repository()
	mock := &mockRepo{Repository: repo, bundleSlug: bundleSlug, bundleID: bundleID, db: db}

	handler := NewStripeWebhookHandler(nil, nil, nil, mock, nil, nil, nil, nil, nil, nil, nil)
	handler.webhookSecret = ""

	router := http.NewServeMux()
	router.HandleFunc("/webhooks/stripe", handler.HandleWebhook)

	event := stripe.Event{
		Type: "checkout.session.completed",
		Data: &stripe.EventData{
			Raw: json.RawMessage(fmt.Sprintf(`{
				"id": "%s",
				"amount_total": 9900,
				"currency": "usd",
				"subscription": {"id": "sub_123"},
				"metadata": {
					"tenant_id": "%s",
					"bundle_slug": "%s",
					"purpose": "bundle_subscription"
				}
			}`, sessionID, tenantID, bundleSlug)),
		},
	}

	payload, _ := json.Marshal(event)
	req := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "webhook should return 200")

	var bundleSubCount int
	err = db.QueryRow("SELECT COUNT(*) FROM bundle_subscriptions WHERE tenant_id = $1", tenantID).Scan(&bundleSubCount)
	require.NoError(t, err)
	assert.Equal(t, 1, bundleSubCount, "bundle subscription should be created")

	var invoiceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM invoices WHERE external_reference = $1", sessionID).Scan(&invoiceCount)
	require.NoError(t, err)
	assert.Equal(t, 1, invoiceCount, "invoice should be created with session ID as external_reference")

	var invoiceAmount int
	err = db.QueryRow("SELECT amount_due_cents FROM invoices WHERE external_reference = $1", sessionID).Scan(&invoiceAmount)
	require.NoError(t, err)
	assert.Equal(t, 9900, invoiceAmount, "invoice amount should match session amount_total")
}
