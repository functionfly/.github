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
	// MinRegistryWalletTopUpUSD matches the dashboard copy (minimum $1.00 top-up).
	MinRegistryWalletTopUpUSD = 1.0
	// MaxRegistryWalletTopUpUSD caps a single top-up (aligned with agent credits checkout).
	MaxRegistryWalletTopUpUSD = 10_000.0
)

// RegistryWalletCheckoutResult is returned when creating a Stripe Checkout session for registry wallet balance.
type RegistryWalletCheckoutResult struct {
	SessionID string
	URL       string
}

// CreateRegistryWalletCheckoutSession creates a one-time Checkout session to add funds to the user's registry wallet
// (platform fee balance). Metadata is consumed by the Stripe webhook (purpose: registry_wallet_credit).
func CreateRegistryWalletCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	userID uuid.UUID,
	email, name string,
	amountUSD float64,
	successURL, cancelURL string,
) (*RegistryWalletCheckoutResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	if amountUSD < MinRegistryWalletTopUpUSD {
		return nil, fmt.Errorf("minimum amount is $%.2f USD", MinRegistryWalletTopUpUSD)
	}
	if amountUSD > MaxRegistryWalletTopUpUSD {
		return nil, fmt.Errorf("maximum amount is $%.2f USD", MaxRegistryWalletTopUpUSD)
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
		successURL = strings.TrimSuffix(successURL, "/") + "/settings?walletTopUp=success"
	}
	if cancelURL == "" {
		cancelURL = os.Getenv("APP_URL")
		if cancelURL == "" {
			cancelURL = "http://localhost:3000"
		}
		cancelURL = strings.TrimSuffix(cancelURL, "/") + "/settings?walletTopUp=cancel"
	}

	amountCents := int64(amountUSD * 100)
	if amountCents < 100 {
		amountCents = 100
	}

	md := map[string]string{
		"tenant_id":          tenantID.String(),
		"user_id":            userID.String(),
		"purpose":            "registry_wallet_credit",
		"amount_usd":         fmt.Sprintf("%.4f", amountUSD),
		"initiating_user_id": userID.String(),
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
						Name:        stripe.String("Registry wallet balance"),
						Description: stripe.String("Prepaid balance for FunctionFly registry fees"),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
			},
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

	return &RegistryWalletCheckoutResult{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}
