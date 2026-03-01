package payment

import (
	"context"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/paymentintent"
)

func stripeKey() string {
	if stripe.Key != "" {
		return stripe.Key
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return stripe.Key
}

// ChargeResult holds the result of a successful charge (e.g. PaymentIntent ID for idempotency).
type ChargeResult struct {
	PaymentIntentID string
}

// Charge charges the given payment method the specified amount in USD.
// amountUSD is in dollars (e.g. 10.50); it is converted to cents for Stripe.
// Metadata is attached to the PaymentIntent for reference (e.g. agent_id, tenant_id).
func Charge(ctx context.Context, paymentMethodID string, amountUSD float64, metadata map[string]string) (*ChargeResult, error) {
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

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe payment failed: %w", err)
	}

	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded, stripe.PaymentIntentStatusRequiresCapture:
		return &ChargeResult{PaymentIntentID: pi.ID}, nil
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
