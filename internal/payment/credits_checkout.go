package payment

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	checkoutsession "github.com/stripe/stripe-go/v83/checkout/session"
)

const (
	// MinCreditsCheckoutUSD is Stripe's practical minimum for card charges (aligned with payment.Charge).
	MinCreditsCheckoutUSD = 0.50
	// MaxCreditsCheckoutUSD caps a single top-up to reduce fraud exposure.
	MaxCreditsCheckoutUSD = 10_000.0
)

// AgentCreditsCheckoutResult is returned when creating a one-time Checkout session for agent credits.
type AgentCreditsCheckoutResult struct {
	SessionID string
	URL       string
}

// CreateAgentCreditsCheckoutSession creates a Stripe Checkout Session (payment mode) for purchasing
// execution credits for a specific agent. Metadata is used by the webhook to credit the correct agent.
func CreateAgentCreditsCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	initiatingUserID uuid.UUID,
	email, name string,
	agentID string,
	amountUSD float64,
	successURL, cancelURL string,
) (*AgentCreditsCheckoutResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}
	if amountUSD < MinCreditsCheckoutUSD {
		return nil, fmt.Errorf("minimum amount is $%.2f USD", MinCreditsCheckoutUSD)
	}
	if amountUSD > MaxCreditsCheckoutUSD {
		return nil, fmt.Errorf("maximum amount is $%.2f USD", MaxCreditsCheckoutUSD)
	}

	customerID, err := CreateOrGetStripeCustomer(ctx, repo, tenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	if successURL == "" {
		successURL = os.Getenv("APP_URL")
		if successURL == "" {
			successURL = "http://localhost:3000"
		}
		successURL = strings.TrimSuffix(successURL, "/") + "/agents/" + agentID + "/wallet?credits=success"
	}
	if cancelURL == "" {
		cancelURL = os.Getenv("APP_URL")
		if cancelURL == "" {
			cancelURL = "http://localhost:3000"
		}
		cancelURL = strings.TrimSuffix(cancelURL, "/") + "/agents/" + agentID + "/wallet?credits=cancel"
	}

	amountCents := int64(amountUSD * 100)
	if amountCents < 50 {
		amountCents = 50
	}

	md := map[string]string{
		"tenant_id":          tenantID.String(),
		"initiating_user_id": initiatingUserID.String(),
		"agent_id":           agentID,
		"purpose":            "agent_execution_credits",
		"amount_usd":         fmt.Sprintf("%.4f", amountUSD),
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
						Name:        stripe.String("Agent execution credits"),
						Description: stripe.String("Prepaid credits for agent " + agentID),
					},
					UnitAmount: stripe.Int64(amountCents),
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

	sess, err := checkoutsession.New(params)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}
	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &AgentCreditsCheckoutResult{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}
