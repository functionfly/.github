package payment

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	stripeSub "github.com/stripe/stripe-go/v83/subscription"
)

// CreateCheckoutSessionRequest contains the parameters for creating a checkout session.
type CreateCheckoutSessionRequest struct {
	PriceID      string `json:"price_id"`
	SuccessURL   string `json:"success_url"`
	CancelURL    string `json:"cancel_url"`
	FounderModeID string `json:"founder_mode_id"` // Passed through to Stripe metadata for conversion tracking
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

// CreateBundleCheckoutSessionRequest contains parameters for creating a bundle checkout session.
type CreateBundleCheckoutSessionRequest struct {
	PriceID       string
	SuccessURL    string
	CancelURL     string
	TenantID      uuid.UUID
	BundleSlug    string
	FounderModeID string
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
		appURL = config.GetFrontendURL()
	}

	allowedURL, err := url.Parse(appURL)
	if err != nil {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host != allowedURL.Host {
		return false
	}
	if strings.HasPrefix(parsed.Path, "//") || strings.HasPrefix(parsed.Path, "/\\") {
		return false
	}
	return true
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
		appURL = config.GetFrontendURL()
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
		// Enable automatic tax calculation via Stripe Tax
		// This handles EU VAT, US sales tax, and global tax compliance
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
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
		appURL = config.GetFrontendURL()
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
		// Enable automatic tax calculation via Stripe Tax
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
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

// CreateBundleCheckoutSession creates a Stripe Checkout session for a bundle subscription.
// It includes bundle metadata in both session and subscription metadata for proper webhook processing.
func CreateBundleCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	tenantID uuid.UUID,
	email, name string,
	req CreateBundleCheckoutSessionRequest,
) (*CreateCheckoutSessionResponse, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	if err := ValidatePriceID(req.PriceID); err != nil {
		return nil, err
	}

	customerID, err := CreateOrGetStripeCustomer(ctx, repo, tenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = config.GetFrontendURL()
	}

	successURL := req.SuccessURL
	cancelURL := req.CancelURL
	if successURL == "" {
		successURL = appURL + "/dashboard?bundle=" + req.BundleSlug + "&success=true"
	}
	if cancelURL == "" {
		cancelURL = appURL + "/pricing/bundles"
	}

	successURL = SanitizeReturnURL(successURL, appURL+"/dashboard?bundle="+req.BundleSlug+"&success=true")
	cancelURL = SanitizeReturnURL(cancelURL, appURL+"/pricing/bundles")

	// Build metadata map
	metadata := map[string]string{
		"purpose":     "bundle_subscription",
		"tenant_id":   tenantID.String(),
		"bundle_slug": req.BundleSlug,
	}
	if req.FounderModeID != "" {
		metadata["founder_mode_id"] = req.FounderModeID
	}

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
		// Enable automatic tax calculation via Stripe Tax
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: metadata,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("create bundle checkout session: %w", err)
	}
	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &CreateCheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

// CreateUsernameChangeCheckoutSessionRequest contains parameters for username change fee checkout.
type CreateUsernameChangeCheckoutSessionRequest struct {
	SuccessURL      string
	CancelURL       string
	TenantID        uuid.UUID
	UserID          uuid.UUID
	PendingChangeID uuid.UUID
	NewUsername     string
	FeeCents        int
}

// UpdateBundleSubscriptionRequest contains parameters for changing bundle subscription
type UpdateBundleSubscriptionRequest struct {
	SubscriptionID string
	NewPriceID    string
	Prorate       bool
}

// CreateUsernameChangeCheckoutSession creates a Stripe Checkout session for username change fee.
// This is a one-time payment (not a subscription) for users who want to change their username
// before the 6-month free window expires.
func CreateUsernameChangeCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	email, name string,
	req CreateUsernameChangeCheckoutSessionRequest,
) (*CreateCheckoutSessionResponse, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	if req.FeeCents <= 0 {
		return nil, fmt.Errorf("fee must be greater than 0")
	}

	if req.PendingChangeID == uuid.Nil {
		return nil, fmt.Errorf("pending_change_id is required")
	}

	customerID, err := CreateOrGetStripeCustomer(ctx, repo, req.TenantID, email, name)
	if err != nil {
		return nil, fmt.Errorf("create or get stripe customer: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = config.GetFrontendURL()
	}

	successURL := req.SuccessURL
	cancelURL := req.CancelURL
	if successURL == "" {
		successURL = appURL + "/settings?usernameChange=success"
	}
	if cancelURL == "" {
		cancelURL = appURL + "/settings?usernameChange=cancel"
	}

	successURL = SanitizeReturnURL(successURL, appURL+"/settings?usernameChange=success")
	cancelURL = SanitizeReturnURL(cancelURL, appURL+"/settings?usernameChange=cancel")

	// Build metadata for webhook processing
	metadata := map[string]string{
		"purpose":           "username_change",
		"tenant_id":         req.TenantID.String(),
		"user_id":           req.UserID.String(),
		"pending_change_id": req.PendingChangeID.String(),
		"new_username":      req.NewUsername,
		"fee_cents":         fmt.Sprintf("%d", req.FeeCents),
	}

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)), // One-time payment, not subscription
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("usd"),
					UnitAmount: stripe.Int64(int64(req.FeeCents)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("Username Change Fee"),
						Description: stripe.String(fmt.Sprintf("Early username change to @%s", req.NewUsername)),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		// Enable automatic tax calculation via Stripe Tax
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: metadata,
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("create username change checkout session: %w", err)
	}
	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &CreateCheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

// UpdateBundleSubscription updates a Stripe subscription to a new price (for bundle changes)
func UpdateBundleSubscription(ctx context.Context, req UpdateBundleSubscriptionRequest) error {
	if stripeKey() == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	if err := ValidatePriceID(req.NewPriceID); err != nil {
		return err
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(req.SubscriptionID),
				Price: stripe.String(req.NewPriceID),
			},
		},
		ProrationBehavior: stripe.String("always_invoice"),
	}

	if !req.Prorate {
		params.ProrationBehavior = stripe.String("none")
	}

	_, err := stripeSub.Update(req.SubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}
