package payment

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
)

// CreateCheckoutSessionRequest contains the parameters for creating a checkout session.
type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// IsValidStripePriceID validates that the ID is a valid Stripe price ID (starts with "price_").
// This prevents common errors like using product IDs (prod_*) or subscription IDs (sub_*) instead of price IDs.
func IsValidStripePriceID(priceID string) bool {
	if priceID == "" {
		return false
	}
	// Stripe price IDs must start with "price_"
	return strings.HasPrefix(priceID, "price_")
}

// ValidatePriceID returns an error if the price ID is not a valid Stripe price ID.
// The error message includes helpful details if a product ID or other invalid ID is detected.
func ValidatePriceID(priceID string) error {
	if priceID == "" {
		return fmt.Errorf("price_id is required")
	}

	// Check for common mistakes
	if strings.HasPrefix(priceID, "prod_") {
		return fmt.Errorf("invalid price_id: received product ID (%s) instead of price ID. Product IDs (prod_*) cannot be used for checkout - use the associated price ID (price_*) from Stripe Dashboard", priceID)
	}
	if strings.HasPrefix(priceID, "sub_") {
		return fmt.Errorf("invalid price_id: received subscription ID (%s) instead of price ID. Subscription IDs (sub_*) cannot be used for checkout", priceID)
	}
	if strings.HasPrefix(priceID, "plan_") {
		return fmt.Errorf("invalid price_id: received plan ID (%s) instead of price ID. Plan IDs (plan_*) are deprecated, use price IDs (price_*) instead", priceID)
	}

	if !strings.HasPrefix(priceID, "price_") {
		return fmt.Errorf("invalid price_id: must start with 'price_', got: %s", priceID)
	}

	return nil
}

// CreateAddonCheckoutSessionRequest creates a subscription checkout for a State Fabric add-on.
type CreateAddonCheckoutSessionRequest struct {
	PriceID    string
	SuccessURL string
	CancelURL  string
	TenantID   uuid.UUID
	AddonID    string
}

// CreateCheckoutSessionResponse contains the response from creating a checkout session.
type CreateCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// IsValidReturnURL validates that a return URL is safe to use.
func IsValidReturnURL(returnURL string) bool {
	if returnURL == "" {
		return false
	}

	parsed, err := url.Parse(returnURL)
	if err != nil {
		return false
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	allowedURL, err := url.Parse(appURL)
	if err != nil {
		return false
	}

	return parsed.Host == allowedURL.Host
}

// SanitizeReturnURL ensures the return URL is valid and safe.
func SanitizeReturnURL(returnURL, defaultURL string) string {
	if IsValidReturnURL(returnURL) {
		return returnURL
	}
	return defaultURL
}

// CreateCheckoutSession creates a Stripe Checkout session for subscription checkout.
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

	// Validate the price ID to prevent common mistakes
	if err := ValidatePriceID(req.PriceID); err != nil {
		return nil, err
	}

	customerID, err := CreateOrGetStripeCustomer(ctx, repo, tenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	successURL := req.SuccessURL
	cancelURL := req.CancelURL

	if successURL == "" {
		successURL = appURL + "/settings?tab=billing&subscription=success"
	}
	if cancelURL == "" {
		cancelURL = appURL + "/pricing?subscription=cancel"
	}

	successURL = SanitizeReturnURL(successURL, appURL+"/settings?tab=billing&subscription=success")
	cancelURL = SanitizeReturnURL(cancelURL, appURL+"/pricing?subscription=cancel")

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
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

// CreateStateFabricAddonCheckoutSession creates a checkout session tied to one add-on entitlement.
func CreateStateFabricAddonCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	email, name string,
	req CreateAddonCheckoutSessionRequest,
) (*CreateCheckoutSessionResponse, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	// Validate the price ID to prevent common mistakes
	if err := ValidatePriceID(req.PriceID); err != nil {
		return nil, err
	}
	if req.AddonID == "" {
		return nil, fmt.Errorf("addon_id is required")
	}

	customerID, err := CreateOrGetStripeCustomer(ctx, repo, tenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	successURL := req.SuccessURL
	cancelURL := req.CancelURL
	if successURL == "" {
		successURL = appURL + "/pricing?stateFabricAddOn=success"
	}
	if cancelURL == "" {
		cancelURL = appURL + "/pricing?stateFabricAddOn=cancel"
	}

	successURL = SanitizeReturnURL(successURL, appURL+"/pricing?stateFabricAddOn=success")
	cancelURL = SanitizeReturnURL(cancelURL, appURL+"/pricing?stateFabricAddOn=cancel")

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
		Metadata: map[string]string{
			"purpose":   "state_fabric_addon",
			"tenant_id": req.TenantID.String(),
			"addon_id":  req.AddonID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"purpose":   "state_fabric_addon",
				"tenant_id": req.TenantID.String(),
				"addon_id":  req.AddonID,
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
	return &CreateCheckoutSessionResponse{SessionID: sess.ID, URL: sess.URL}, nil
}
