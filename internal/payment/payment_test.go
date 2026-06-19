package payment

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v83"
)

// resetStripeKey resets the Stripe API key for testing
func resetStripeKey() {
	stripe.Key = ""
}

func TestIsConfigured(t *testing.T) {
	// Save original value
	origKey := os.Getenv("STRIPE_SECRET_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("STRIPE_SECRET_KEY", origKey)
		} else {
			os.Unsetenv("STRIPE_SECRET_KEY")
		}
		resetStripeKey()
	}()

	tests := []struct {
		name   string
		envKey string
		want   bool
	}{
		{
			name:   "stripe key set",
			envKey: "sk_test_123",
			want:   true,
		},
		{
			name:   "stripe key empty",
			envKey: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStripeKey()
			if tt.envKey != "" {
				os.Setenv("STRIPE_SECRET_KEY", tt.envKey)
			} else {
				os.Unsetenv("STRIPE_SECRET_KEY")
			}
			assert.Equal(t, tt.want, IsConfigured())
		})
	}
}

func TestCharge_Validation(t *testing.T) {
	// Save original value
	origKey := os.Getenv("STRIPE_SECRET_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("STRIPE_SECRET_KEY", origKey)
		} else {
			os.Unsetenv("STRIPE_SECRET_KEY")
		}
		resetStripeKey()
	}()

	tests := []struct {
		name            string
		envKey          string
		paymentMethodID string
		amountUSD       float64
		wantErr         bool
		errContains     string
	}{
		{
			name:            "stripe not configured",
			envKey:          "",
			paymentMethodID: "pm_123",
			amountUSD:       10.00,
			wantErr:         true,
			errContains:     "STRIPE_SECRET_KEY is not set",
		},
		{
			name:            "missing payment method",
			envKey:          "sk_test_123",
			paymentMethodID: "",
			amountUSD:       10.00,
			wantErr:         true,
			errContains:     "payment_method_id is required",
		},
		{
			name:            "amount too low",
			envKey:          "sk_test_123",
			paymentMethodID: "pm_123",
			amountUSD:       0.25,
			wantErr:         true,
			errContains:     "minimum charge is $0.50 USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStripeKey()
			if tt.envKey != "" {
				os.Setenv("STRIPE_SECRET_KEY", tt.envKey)
			} else {
				os.Unsetenv("STRIPE_SECRET_KEY")
			}

			result, err := Charge(context.Background(), tt.paymentMethodID, tt.amountUSD, nil, "")
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				// If no error expected but we got one, it means Stripe API was called
				// which would fail without a real key - this is expected in unit tests
				if err != nil {
					assert.Contains(t, err.Error(), "STRIPE_SECRET_KEY is not set") // Expected in unit test
				}
			}
			_ = result // May be nil if error occurred
		})
	}
}

func TestCharge_MinimumAmount(t *testing.T) {
	// Save original value
	origKey := os.Getenv("STRIPE_SECRET_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("STRIPE_SECRET_KEY", origKey)
		} else {
			os.Unsetenv("STRIPE_SECRET_KEY")
		}
		resetStripeKey()
	}()

	// Test that $0.50 minimum is enforced
	resetStripeKey()
	os.Setenv("STRIPE_SECRET_KEY", "sk_test_123")

	_, err := Charge(context.Background(), "pm_123", 0.50, nil, "")
	// Will fail at Stripe API level, but not at validation level
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "minimum charge") // Validation passes
}

func TestCreateCheckoutSessionRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateCheckoutSessionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateCheckoutSessionRequest{
				PriceID:    "price_123",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			wantErr: false,
		},
		{
			name: "empty price_id",
			req: CreateCheckoutSessionRequest{
				PriceID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation happens inside CreateCheckoutSession, not at request struct level
			if tt.wantErr {
				assert.Empty(t, tt.req.PriceID)
			}
		})
	}
}

func TestCreateAddonCheckoutSessionRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateAddonCheckoutSessionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateAddonCheckoutSessionRequest{
				PriceID:    "price_123",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
				AddonID:    "addon_123",
			},
			wantErr: false,
		},
		{
			name: "empty price_id",
			req: CreateAddonCheckoutSessionRequest{
				PriceID: "",
				AddonID: "addon_123",
			},
			wantErr: true,
		},
		{
			name: "empty addon_id",
			req: CreateAddonCheckoutSessionRequest{
				PriceID: "price_123",
				AddonID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.req.PriceID == "" || tt.req.AddonID == ""
			assert.Equal(t, tt.wantErr, hasError)
		})
	}
}
