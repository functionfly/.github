package payment

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/account"
	"github.com/stripe/stripe-go/v83/accountlink"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/paymentintent"
)

var (
	stripeKeyOnce sync.Once
	stripeKeyVal  string
)

func stripeKey() string {
	stripeKeyOnce.Do(func() {
		stripeKeyVal = os.Getenv("STRIPE_SECRET_KEY")
	})
	if stripeKeyVal != "" {
		stripe.Key = stripeKeyVal
	}
	return stripe.Key
}

// ChargeResult holds the result of a successful charge (e.g. PaymentIntent ID for idempotency).
type ChargeResult struct {
	PaymentIntentID string
	IdempotencyKey  string
}

// Charge charges the given payment method the specified amount in USD.
// amountUSD is in dollars (e.g. 10.50); it is converted to cents for Stripe.
// Metadata is attached to the PaymentIntent for reference (e.g. agent_id, tenant_id).
// IdempotencyKey is recommended for production use to prevent duplicate charges.
// The Stripe SDK automatically retries on network errors with the same idempotency key.
func Charge(ctx context.Context, paymentMethodID string, amountUSD float64, metadata map[string]string, idempotencyKey string) (*ChargeResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if paymentMethodID == "" {
		return nil, fmt.Errorf("payment_method_id is required")
	}
	if amountUSD < 0.50 {
		return nil, fmt.Errorf("minimum charge is $0.50 USD")
	}

	amountCents := int64(amountUSD * 100)
	if amountCents < 50 {
		amountCents = 50
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountCents),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		PaymentMethod: stripe.String(paymentMethodID),
		Confirm:       stripe.Bool(true),
	}
	if metadata != nil {
		params.Metadata = metadata
	}
	params.SetStripeAccount("") // use default account

	// Apply idempotency key if provided to prevent duplicate charges
	// Stripe ignores duplicate requests within 24 hours and returns the original result
	if idempotencyKey != "" {
		params.Params.IdempotencyKey = stripe.String(idempotencyKey)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe payment failed: %w", err)
	}

	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded, stripe.PaymentIntentStatusRequiresCapture:
		return &ChargeResult{PaymentIntentID: pi.ID, IdempotencyKey: idempotencyKey}, nil
	case stripe.PaymentIntentStatusRequiresAction:
		return nil, fmt.Errorf("payment requires additional authentication (e.g. 3D Secure)")
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		return nil, fmt.Errorf("payment failed; try a different payment method")
	case stripe.PaymentIntentStatusCanceled:
		return nil, fmt.Errorf("payment was canceled")
	default:
		return nil, fmt.Errorf("unexpected payment status: %s", pi.Status)
	}
}

// IsConfigured reports whether Stripe is configured (secret key set).
func IsConfigured() bool {
	return stripeKey() != ""
}

// CreateStripeCustomer creates a new Stripe customer with metadata.
// This is used for tenant-isolated payment processing.
func CreateStripeCustomer(ctx context.Context, email, name string, metadata map[string]string) (*stripe.Customer, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	if metadata != nil {
		params.Metadata = metadata
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	return cust, nil
}

// CreateConnectAccount creates a new Stripe Connect Express account for marketplace tenants.
// This enables the tenant to receive payments directly to their own Stripe account.
func CreateConnectAccount(ctx context.Context, email string, metadata map[string]string) (string, error) {
	if stripeKey() == "" {
		return "", fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	params := &stripe.AccountParams{
		Type:  stripe.String(string(stripe.AccountTypeExpress)),
		Email: stripe.String(email),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		BusinessType: stripe.String(string(stripe.AccountBusinessTypeIndividual)),
	}
	if metadata != nil {
		params.Metadata = metadata
	}

	acct, err := account.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe Connect account: %w", err)
	}

	return acct.ID, nil
}

// CreateConnectAccountOnboardingLink creates an onboarding link for a Stripe Connect account.
// This URL allows the tenant to complete their Stripe Connect onboarding.
func CreateConnectAccountOnboardingLink(ctx context.Context, accountID string) (string, error) {
	if stripeKey() == "" {
		return "", fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	returnURL := os.Getenv("APP_URL")
	if returnURL == "" {
		returnURL = "https://functionfly.com"
	}

	params := &stripe.AccountLinkParams{
		Account:    stripe.String(accountID),
		RefreshURL: stripe.String(returnURL + "/settings/payouts?refresh=true"),
		ReturnURL:  stripe.String(returnURL + "/settings/payouts?connected=true"),
		Type:       stripe.String(string(stripe.AccountLinkTypeAccountOnboarding)),
	}

	link, err := accountlink.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create account link: %w", err)
	}

	return link.URL, nil
}
