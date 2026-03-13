package payment

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

// CreateCheckoutSessionRequest contains the parameters for creating a checkout session.
type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// CreateCheckoutSessionResponse contains the response from creating a checkout session.
type CreateCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// CreateCheckoutSession creates a Stripe Checkout session for subscription checkout.
// It creates or retrieves the Stripe customer, then creates a checkout session with the given price ID.
func CreateCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	email, name string,
	req CreateCheckoutSessionRequest,
) (*CreateCheckoutSessionResponse, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	if req.PriceID == "" {
		return nil, fmt.Errorf("price_id is required")
	}

	// Get or create Stripe customer
	customerID, err := CreateOrGetStripeCustomer(ctx, repo, tenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	// Build success and cancel URLs
	successURL := req.SuccessURL
	cancelURL := req.CancelURL
	if successURL == "" {
		successURL = os.Getenv("APP_URL")
		if successURL == "" {
			successURL = "http://localhost:3000/settings?tab=billing&success=true"
		}
	}
	if cancelURL == "" {
		cancelURL = os.Getenv("APP_URL")
		if cancelURL == "" {
			cancelURL = "http://localhost:3000/pricing"
		}
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer:  stripe.String(customerID),
		Mode:      stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"tenant_id": tenantID.String(),
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &CreateCheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}
