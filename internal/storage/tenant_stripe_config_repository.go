package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TenantStripeConfig holds per-tenant Stripe configuration for isolated payment processing.
// This enables each tenant to have their own Stripe customer and payment configuration.
type TenantStripeConfig struct {
	ID                    uuid.UUID  `json:"id"`
	TenantID              uuid.UUID  `json:"tenant_id"`
	StripeCustomerID      string     `json:"stripe_customer_id"`       // Tenant's Stripe Customer ID
	StripeConnectAccountID *string    `json:"stripe_connect_account_id,omitempty"` // For marketplace/gateway usage
	IsolatedPaymentEnabled bool       `json:"isolated_payment_enabled"` // Whether tenant uses isolated payment flow
	PaymentMode           string     `json:"payment_mode"`            // "platform" (default), "isolated"
	AllowedPaymentMethods  string     `json:"allowed_payment_methods"`  // JSON array of allowed payment method types
	DefaultPaymentMethod  *string    `json:"default_payment_method,omitempty"`
	BillingAddressRequired bool      `json:"billing_address_required"`
	TaxCalculationMode     string     `json:"tax_calculation_mode"`     // "automatic", "manual"
	Metadata              JSONMap    `json:"metadata,omitempty"`       // Additional tenant-specific config
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// TenantPaymentConfig represents the payment configuration for a bundle/tenant
type TenantPaymentConfig struct {
	TenantID              uuid.UUID `json:"tenant_id"`
	BundleSlug            string    `json:"bundle_slug"`
	StripeCustomerID      string    `json:"stripe_customer_id"`
	IsolatedPaymentFlow   bool      `json:"isolated_payment_flow"`   // True if tenant has isolated payments
	UseTenantStripeAccount bool     `json:"use_tenant_stripe_account"` // Route payments through tenant's Stripe
	WebhookEndpoint       string    `json:"webhook_endpoint"`        // Tenant-specific webhook URL
	AllowedPaymentMethods []string  `json:"allowed_payment_methods"`  // e.g., ["card", "us_bank_account"]
}

// TenantStripeConfigRepository handles database operations for tenant Stripe configurations.
type TenantStripeConfigRepository struct {
	db *sql.DB
}

// NewTenantStripeConfigRepository creates a new tenant Stripe config repository.
func NewTenantStripeConfigRepository(db *sql.DB) *TenantStripeConfigRepository {
	return &TenantStripeConfigRepository{db: db}
}

// Create inserts a new tenant Stripe configuration.
func (r *TenantStripeConfigRepository) Create(ctx context.Context, config *TenantStripeConfig) error {
	query := `
		INSERT INTO tenant_stripe_configs (
			id, tenant_id, stripe_customer_id, stripe_connect_account_id,
			isolated_payment_enabled, payment_mode, allowed_payment_methods,
			default_payment_method, billing_address_required, tax_calculation_mode,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	metadataJSON, err := json.Marshal(config.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = r.db.ExecContext(ctx, query,
		config.ID,
		config.TenantID,
		config.StripeCustomerID,
		config.StripeConnectAccountID,
		config.IsolatedPaymentEnabled,
		config.PaymentMode,
		config.AllowedPaymentMethods,
		config.DefaultPaymentMethod,
		config.BillingAddressRequired,
		config.TaxCalculationMode,
		metadataJSON,
		config.CreatedAt,
		config.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create tenant stripe config: %w", err)
	}
	return nil
}

// GetByTenantID retrieves the Stripe configuration for a tenant.
func (r *TenantStripeConfigRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*TenantStripeConfig, error) {
	query := `
		SELECT id, tenant_id, stripe_customer_id, stripe_connect_account_id,
			   isolated_payment_enabled, payment_mode, allowed_payment_methods,
			   default_payment_method, billing_address_required, tax_calculation_mode,
			   metadata, created_at, updated_at
		FROM tenant_stripe_configs
		WHERE tenant_id = $1
	`
	var config TenantStripeConfig
	var metadataJSON []byte
	var defaultPaymentMethod sql.NullString
	var stripeConnectAccountID sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&config.ID,
		&config.TenantID,
		&config.StripeCustomerID,
		&stripeConnectAccountID,
		&config.IsolatedPaymentEnabled,
		&config.PaymentMode,
		&config.AllowedPaymentMethods,
		&defaultPaymentMethod,
		&config.BillingAddressRequired,
		&config.TaxCalculationMode,
		&metadataJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant stripe config: %w", err)
	}

	if defaultPaymentMethod.Valid {
		config.DefaultPaymentMethod = &defaultPaymentMethod.String
	}
	if stripeConnectAccountID.Valid {
		config.StripeConnectAccountID = &stripeConnectAccountID.String
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &config.Metadata)
	}

	return &config, nil
}

// Update updates an existing tenant Stripe configuration.
func (r *TenantStripeConfigRepository) Update(ctx context.Context, config *TenantStripeConfig) error {
	query := `
		UPDATE tenant_stripe_configs SET
			stripe_customer_id = $1,
			stripe_connect_account_id = $2,
			isolated_payment_enabled = $3,
			payment_mode = $4,
			allowed_payment_methods = $5,
			default_payment_method = $6,
			billing_address_required = $7,
			tax_calculation_mode = $8,
			metadata = $9,
			updated_at = $10
		WHERE id = $11 AND tenant_id = $12
	`
	metadataJSON, err := json.Marshal(config.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	result, err := r.db.ExecContext(ctx, query,
		config.StripeCustomerID,
		config.StripeConnectAccountID,
		config.IsolatedPaymentEnabled,
		config.PaymentMode,
		config.AllowedPaymentMethods,
		config.DefaultPaymentMethod,
		config.BillingAddressRequired,
		config.TaxCalculationMode,
		metadataJSON,
		time.Now(),
		config.ID,
		config.TenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant stripe config: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant stripe config not found")
	}
	return nil
}

// Delete removes a tenant Stripe configuration.
func (r *TenantStripeConfigRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	query := `DELETE FROM tenant_stripe_configs WHERE tenant_id = $1`
	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant stripe config: %w", err)
	}
	return nil
}

// ListIsolatedPaymentTenants returns all tenants with isolated payment flow enabled.
func (r *TenantStripeConfigRepository) ListIsolatedPaymentTenants(ctx context.Context) ([]*TenantStripeConfig, error) {
	query := `
		SELECT id, tenant_id, stripe_customer_id, stripe_connect_account_id,
			   isolated_payment_enabled, payment_mode, allowed_payment_methods,
			   default_payment_method, billing_address_required, tax_calculation_mode,
			   metadata, created_at, updated_at
		FROM tenant_stripe_configs
		WHERE isolated_payment_enabled = true
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list isolated payment tenants: %w", err)
	}
	defer rows.Close()

	var configs []*TenantStripeConfig
	for rows.Next() {
		var config TenantStripeConfig
		var metadataJSON []byte
		var defaultPaymentMethod, stripeConnectAccountID sql.NullString

		err := rows.Scan(
			&config.ID,
			&config.TenantID,
			&config.StripeCustomerID,
			&stripeConnectAccountID,
			&config.IsolatedPaymentEnabled,
			&config.PaymentMode,
			&config.AllowedPaymentMethods,
			&defaultPaymentMethod,
			&config.BillingAddressRequired,
			&config.TaxCalculationMode,
			&metadataJSON,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant stripe config: %w", err)
		}

		if defaultPaymentMethod.Valid {
			config.DefaultPaymentMethod = &defaultPaymentMethod.String
		}
		if stripeConnectAccountID.Valid {
			config.StripeConnectAccountID = &stripeConnectAccountID.String
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &config.Metadata)
		}

		configs = append(configs, &config)
	}
	return configs, nil
}

// UpdateStripeCustomerID updates the Stripe customer ID for a tenant.
func (r *TenantStripeConfigRepository) UpdateStripeCustomerID(ctx context.Context, tenantID uuid.UUID, customerID string) error {
	query := `
		UPDATE tenant_stripe_configs SET
			stripe_customer_id = $1,
			updated_at = $2
		WHERE tenant_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, customerID, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to update stripe customer id: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant stripe config not found")
	}
	return nil
}

// EnableIsolatedPayment enables isolated payment flow for a tenant.
func (r *TenantStripeConfigRepository) EnableIsolatedPayment(ctx context.Context, tenantID uuid.UUID, useConnectAccount bool) error {
	paymentMode := "isolated"
	if useConnectAccount {
		paymentMode = "connect"
	}

	query := `
		UPDATE tenant_stripe_configs SET
			isolated_payment_enabled = true,
			payment_mode = $1,
			updated_at = $2
		WHERE tenant_id = $3
	`
	result, err := r.db.ExecContext(ctx, query, paymentMode, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to enable isolated payment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant stripe config not found")
	}
	return nil
}

// GetOrCreate retrieves or creates a tenant Stripe configuration.
func (r *TenantStripeConfigRepository) GetOrCreate(ctx context.Context, tenantID uuid.UUID, stripeCustomerID string) (*TenantStripeConfig, error) {
	existing, err := r.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// Create new config with defaults
	config := &TenantStripeConfig{
		ID:                    uuid.New(),
		TenantID:              tenantID,
		StripeCustomerID:      stripeCustomerID,
		IsolatedPaymentEnabled: false,
		PaymentMode:           "platform",
		AllowedPaymentMethods: `["card"]`,
		BillingAddressRequired: false,
		TaxCalculationMode:    "automatic",
		Metadata:              make(JSONMap),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := r.Create(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}