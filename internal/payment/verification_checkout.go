package payment

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

// VerificationCheckoutResult is returned when creating a checkout session for function verification.
type VerificationCheckoutResult struct {
	SessionID string
	URL       string
}

// CreateVerificationCheckoutSession creates a Stripe Checkout Session (payment mode) for
// function verification fees. This uses inline PriceData for dynamic amounts rather than
// pre-configured Stripe products/prices.
func CreateVerificationCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	userID uuid.UUID,
	email, name string,
	paymentID uuid.UUID,
	functionID uuid.UUID,
	level string,
	amountCents int,
	successURL, cancelURL string,
) (*VerificationCheckoutResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	if amountCents <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0 cents")
	}

	customerID, err := CreateOrGetStripeCustomer(ctx, repo, tenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	if successURL == "" {
		successURL = fmt.Sprintf("%s/functions/%s/verification?success=true&payment_id=%s",
			strings.TrimSuffix(appURL, "/"), functionID.String(), paymentID.String())
	}
	if cancelURL == "" {
		cancelURL = fmt.Sprintf("%s/functions/%s/verification?canceled=true&payment_id=%s",
			strings.TrimSuffix(appURL, "/"), functionID.String(), paymentID.String())
	}

	md := map[string]string{
		"tenant_id":    tenantID.String(),
		"user_id":      userID.String(),
		"payment_id":   paymentID.String(),
		"function_id":  functionID.String(),
		"level":        level,
		"purpose":      "function_verification",
		"amount_cents": fmt.Sprintf("%d", amountCents),
	}

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		Customer:   stripe.String(customerID),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(fmt.Sprintf("Function Verification - %s", strings.Title(level))),
						Description: stripe.String(fmt.Sprintf("Verification fee for function %s at %s level", functionID.String()[:8], level)),
					},
					UnitAmount: stripe.Int64(int64(amountCents)),
				},
			},
		},
		// Enable automatic tax calculation via Stripe Tax
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: md,
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: md,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}
	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &VerificationCheckoutResult{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}
