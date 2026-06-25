package payment

import (
	"testing"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetBundlePaymentRequirements(t *testing.T) {
	tests := []struct {
		name       string
		bundleSlug string
		wantSlug   string
		wantMethods []string
	}{
		{
			name:       "saas-starter bundle",
			bundleSlug: "saas-starter",
			wantSlug:   "saas-starter",
			wantMethods: []string{"card"},
		},
		{
			name:       "marketplace bundle",
			bundleSlug: "marketplace",
			wantSlug:   "marketplace",
			wantMethods: []string{"card", "us_bank_account", "eu_bank_transfer"},
		},
		{
			name:       "ai-app bundle",
			bundleSlug: "ai-app",
			wantSlug:   "ai-app",
			wantMethods: []string{"card", "us_bank_account"},
		},
		{
			name:       "unknown bundle returns defaults",
			bundleSlug: "unknown-bundle",
			wantSlug:   "unknown-bundle",
			wantMethods: []string{"card"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := GetBundlePaymentRequirements(tt.bundleSlug)
			assert.NotNil(t, reqs)
			assert.Equal(t, tt.wantSlug, reqs.BundleSlug)
			assert.True(t, reqs.IsolatedPayments)
			assert.Equal(t, tt.wantMethods, reqs.PaymentMethodTypes)
		})
	}
}

func TestGetBundlePaymentRequirements_MarketplaceConnectAccount(t *testing.T) {
	reqs := GetBundlePaymentRequirements("marketplace")
	assert.NotNil(t, reqs)
	assert.True(t, reqs.RequiresOnboarding)
	assert.True(t, reqs.ConnectAccountEnabled)
	assert.Contains(t, reqs.PaymentMethodTypes, "us_bank_account")
	assert.Contains(t, reqs.PaymentMethodTypes, "eu_bank_transfer")
}

func TestGetBundlePaymentRequirements_SaaSStarter(t *testing.T) {
	reqs := GetBundlePaymentRequirements("saas-starter")
	assert.NotNil(t, reqs)
	assert.False(t, reqs.RequiresOnboarding)
	assert.False(t, reqs.ConnectAccountEnabled)
	assert.Contains(t, reqs.PaymentMethodTypes, "card")
	assert.Equal(t, "automatic", reqs.TaxCalculation)
}

func TestGetBundlePaymentRequirements_AIApp(t *testing.T) {
	reqs := GetBundlePaymentRequirements("ai-app")
	assert.NotNil(t, reqs)
	assert.False(t, reqs.RequiresOnboarding)
	assert.Contains(t, reqs.PaymentMethodTypes, "card")
	assert.Contains(t, reqs.PaymentMethodTypes, "us_bank_account")
}

func TestValidateBundlePaymentRequirements(t *testing.T) {
	tests := []struct {
		name       string
		req        *BundlePaymentRequirements
		config     *storage.TenantStripeConfig
		wantErr    bool
		errContains string
	}{
		{
			name: "nil requirements",
			req:  nil,
			config: &storage.TenantStripeConfig{
				IsolatedPaymentEnabled: true,
			},
			wantErr:    true,
			errContains: "bundle payment requirements not provided",
		},
		{
			name: "nil config",
			req: &BundlePaymentRequirements{
				BundleSlug:       "saas-starter",
				IsolatedPayments: true,
			},
			config:      nil,
			wantErr:    true,
			errContains: "tenant stripe config is required",
		},
		{
			name: "isolated payments required but not enabled",
			req: &BundlePaymentRequirements{
				BundleSlug:       "saas-starter",
				IsolatedPayments: true,
			},
			config: &storage.TenantStripeConfig{
				IsolatedPaymentEnabled: false,
			},
			wantErr:    true,
			errContains: "isolated payments but tenant does not have them enabled",
		},
		{
			name: "requires onboarding but not connect mode",
			req: &BundlePaymentRequirements{
				BundleSlug:            "marketplace",
				IsolatedPayments:      true,
				RequiresOnboarding:    true,
				ConnectAccountEnabled: true,
			},
			config: &storage.TenantStripeConfig{
				IsolatedPaymentEnabled: true,
				PaymentMode:          "isolated",
				StripeConnectAccountID: nil,
			},
			wantErr:    true,
			errContains: "requires Stripe Connect onboarding",
		},
		{
			name: "valid config",
			req: &BundlePaymentRequirements{
				BundleSlug:           "saas-starter",
				IsolatedPayments:     true,
				PaymentMethodTypes:    []string{"card"},
				BillingAddress:       true,
			},
			config: &storage.TenantStripeConfig{
				IsolatedPaymentEnabled: true,
				PaymentMode:           "isolated",
				AllowedPaymentMethods: `["card"]`,
				BillingAddressRequired: true,
			},
			wantErr: false,
		},
		{
			name: "payment method not allowed",
			req: &BundlePaymentRequirements{
				BundleSlug:        "marketplace",
				IsolatedPayments:  true,
				PaymentMethodTypes: []string{"us_bank_account", "eu_bank_transfer"},
			},
			config: &storage.TenantStripeConfig{
				IsolatedPaymentEnabled: true,
				AllowedPaymentMethods: `["card"]`,
			},
			wantErr:    true,
			errContains: "payment method us_bank_account but tenant only allows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBundlePaymentRequirements(tt.req, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBundlePaymentRequirements_ConnectAccount(t *testing.T) {
	connectAccountID := "acct_123"
	req := &BundlePaymentRequirements{
		BundleSlug:            "marketplace",
		IsolatedPayments:      true,
		RequiresOnboarding:    true,
		ConnectAccountEnabled: true,
		PaymentMethodTypes:    []string{"card", "us_bank_account"},
	}
	config := &storage.TenantStripeConfig{
		IsolatedPaymentEnabled: true,
		PaymentMode:           "connect",
		StripeConnectAccountID: &connectAccountID,
		AllowedPaymentMethods: `["card", "us_bank_account"]`,
		BillingAddressRequired: true,
	}

	err := ValidateBundlePaymentRequirements(req, config)
	assert.NoError(t, err)
}

func TestValidateTenantCheckoutRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     TenantCheckoutSessionRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.New(),
				Email:      "test@example.com",
				PriceID:    "price_123",
				BundleSlug: "saas-starter",
			},
			wantErr: false,
		},
		{
			name: "nil tenant ID",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.Nil,
				Email:      "test@example.com",
				PriceID:    "price_123",
				BundleSlug: "saas-starter",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing email",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.New(),
				Email:      "",
				PriceID:    "price_123",
				BundleSlug: "saas-starter",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "missing price ID",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.New(),
				Email:      "test@example.com",
				PriceID:    "",
				BundleSlug: "saas-starter",
			},
			wantErr: true,
			errMsg:  "price_id is required",
		},
		{
			name: "missing bundle slug",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.New(),
				Email:      "test@example.com",
				PriceID:    "price_123",
				BundleSlug: "",
			},
			wantErr: true,
			errMsg:  "bundle_slug is required",
		},
		{
			name: "invalid bundle slug",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.New(),
				Email:      "test@example.com",
				PriceID:    "price_123",
				BundleSlug: "invalid-bundle",
			},
			wantErr: true,
			errMsg:  "invalid bundle_slug",
		},
		{
			name: "invalid price ID format",
			req: TenantCheckoutSessionRequest{
				TenantID:   uuid.New(),
				Email:      "test@example.com",
				PriceID:    "prod_123",
				BundleSlug: "saas-starter",
			},
			wantErr: true,
			errMsg:  "invalid price_id format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTenantCheckoutRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseAllowedPaymentMethods(t *testing.T) {
	tests := []struct {
		name           string
		allowedMethods string
		want           []string
	}{
		{
			name:           "valid JSON array",
			allowedMethods: `["card", "us_bank_account"]`,
			want:           []string{"card", "us_bank_account"},
		},
		{
			name:           "empty string",
			allowedMethods: "",
			want:           []string{"card"},
		},
		{
			name:           "invalid JSON",
			allowedMethods: `{card}`,
			want:           []string{"card"},
		},
		{
			name:           "empty array",
			allowedMethods: `[]`,
			want:           []string{"card"},
		},
		{
			name:           "card only",
			allowedMethods: `["card"]`,
			want:           []string{"card"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAllowedPaymentMethods(tt.allowedMethods)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBuildTenantMetadata(t *testing.T) {
	tenantID := uuid.New()
	founderModeID := uuid.New()

	tests := []struct {
		name         string
		tenantID     uuid.UUID
		bundleSlug  string
		founderMode *uuid.UUID
		paymentMode string
	}{
		{
			name:         "with founder mode",
			tenantID:     tenantID,
			bundleSlug:   "saas-starter",
			founderMode:  &founderModeID,
			paymentMode:  "isolated",
		},
		{
			name:         "without founder mode",
			tenantID:     tenantID,
			bundleSlug:   "marketplace",
			founderMode:  nil,
			paymentMode:  "connect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := buildTenantMetadata(tt.tenantID, tt.bundleSlug, tt.founderMode, tt.paymentMode)

			assert.Equal(t, tt.tenantID.String(), meta["tenant_id"])
			assert.Equal(t, tt.bundleSlug, meta["bundle_slug"])
			assert.Equal(t, tt.paymentMode, meta["payment_mode"])
			assert.Equal(t, "true", meta["isolated"])

			if tt.founderMode != nil {
				assert.Equal(t, tt.founderMode.String(), meta["founder_mode_id"])
			} else {
				_, hasFounderMode := meta["founder_mode_id"]
				assert.False(t, hasFounderMode)
			}
		})
	}
}
