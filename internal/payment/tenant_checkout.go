package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/customer"
)

// TenantPaymentConfig contains configuration for tenant-isolated payment flows.
type TenantPaymentConfig struct {
	TenantID                uuid.UUID
	StripeCustomerID        string
	IsolatedPaymentEnabled  bool
	PaymentMode             string // "platform", "isolated", "connect"
	AllowedPaymentMethods   []string
	BillingAddressRequired  bool
	TaxCalculationMode      string // "automatic", "manual"
	BundleSlug              string
	FounderModeID           *uuid.UUID
}

// TenantCheckoutSessionRequest contains parameters for creating an isolated checkout session.
type TenantCheckoutSessionRequest struct {
	TenantID       uuid.UUID
	Email          string
	Name           string
	PriceID        string
	SuccessURL     string
	CancelURL      string
	BundleSlug     string
	FounderModeID  *uuid.UUID
	PaymentConfig  *TenantPaymentConfig
	Origin         string // Original domain for tenant branding
}

// TenantCheckoutSessionResponse contains the response from creating a tenant checkout session.
type TenantCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	TenantID  string `json:"tenant_id"`
	PaymentMode string `json:"payment_mode"`
}

// IsolatedPaymentResult holds the result of an isolated payment operation.
type IsolatedPaymentResult struct {
	Success           bool
	PaymentIntentID   string
	TenantID          uuid.UUID
	StripeCustomerID  string
	Error             string
}

// CreateTenantCheckoutSession creates a Stripe Checkout session with tenant isolation.
// This ensures each tenant's payments are processed through their own Stripe customer
// and can be optionally isolated via Stripe Connect accounts.
func CreateTenantCheckoutSession(
	ctx context.Context,
	repo storage.Repository,
	req TenantCheckoutSessionRequest,
) (*TenantCheckoutSessionResponse, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}

	// Validate price ID
	if err := ValidatePriceID(req.PriceID); err != nil {
		return nil, err
	}

	// Get or create tenant Stripe configuration
	tenantStripeConfig, err := getOrCreateTenantStripeConfig(ctx, repo, req.TenantID, req.Email, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant Stripe config: %w", err)
	}

	// Determine payment mode
	paymentMode := tenantStripeConfig.PaymentMode
	if req.PaymentConfig != nil && req.PaymentConfig.PaymentMode != "" {
		paymentMode = req.PaymentConfig.PaymentMode
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = config.GetFrontendURL()
	}

	// Use origin for return URLs if provided (for tenant branding)
	origin := req.Origin
	if origin == "" {
		origin = appURL
	}

	successURL := req.SuccessURL
	cancelURL := req.CancelURL
	if successURL == "" {
		successURL = origin + "/dashboard?bundle=" + req.BundleSlug + "&success=true"
	}
	if cancelURL == "" {
		cancelURL = origin + "/pricing/bundles"
	}

	// Sanitize return URLs
	successURL = sanitizeTenantReturnURL(successURL, origin)
	cancelURL = sanitizeTenantReturnURL(cancelURL, origin)

	// Build metadata for webhook processing
	metadata := buildTenantMetadata(req.TenantID, req.BundleSlug, req.FounderModeID, paymentMode)

	// Create checkout session based on payment mode
	switch paymentMode {
	case "isolated":
		return createIsolatedCheckoutSession(ctx, tenantStripeConfig, req, successURL, cancelURL, metadata)
	case "connect":
		return createConnectCheckoutSession(ctx, tenantStripeConfig, req, successURL, cancelURL, metadata)
	default:
		return createPlatformCheckoutSession(ctx, tenantStripeConfig, req, successURL, cancelURL, metadata)
	}
}

// createIsolatedCheckoutSession creates a checkout session with full tenant isolation.
// Each tenant gets their own PaymentIntent and customer, with no platform involvement.
func createIsolatedCheckoutSession(
	ctx context.Context,
	tenantConfig *storage.TenantStripeConfig,
	req TenantCheckoutSessionRequest,
	successURL, cancelURL string,
	metadata map[string]string,
) (*TenantCheckoutSessionResponse, error) {
	// Ensure tenant has a Stripe customer
	customerID, err := ensureTenantStripeCustomer(ctx, tenantConfig, req.Email, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure tenant customer: %w", err)
	}

	// Parse allowed payment methods
	allowedMethods := parseAllowedPaymentMethods(tenantConfig.AllowedPaymentMethods)

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		PaymentMethodTypes: stripe.StringSlice(allowedMethods),
		BillingAddressCollection: stripe.String("required"),
		Metadata: metadata,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create isolated checkout session: %w", err)
	}

	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &TenantCheckoutSessionResponse{
		SessionID:   sess.ID,
		URL:         sess.URL,
		TenantID:    req.TenantID.String(),
		PaymentMode: "isolated",
	}, nil
}

// createConnectCheckoutSession creates a checkout session using Stripe Connect
// for full payment isolation (marketplace-style).
func createConnectCheckoutSession(
	ctx context.Context,
	tenantConfig *storage.TenantStripeConfig,
	req TenantCheckoutSessionRequest,
	successURL, cancelURL string,
	metadata map[string]string,
) (*TenantCheckoutSessionResponse, error) {
	if tenantConfig.StripeConnectAccountID == nil || *tenantConfig.StripeConnectAccountID == "" {
		return nil, fmt.Errorf("tenant does not have a connected Stripe account")
	}

	connectedAccountID := *tenantConfig.StripeConnectAccountID

	// For Connect, we create the session on behalf of the connected account
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(tenantConfig.StripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		AllowPromotionCodes: stripe.Bool(true),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: metadata,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: metadata,
		},
	}

	// Set the connected account ID so the session is created on behalf of that account
	params.SetStripeAccount(connectedAccountID)

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create connect checkout session: %w", err)
	}

	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &TenantCheckoutSessionResponse{
		SessionID:   sess.ID,
		URL:         sess.URL,
		TenantID:    req.TenantID.String(),
		PaymentMode: "connect",
	}, nil
}

// createPlatformCheckoutSession creates a standard checkout session on the platform account.
// This is the default mode where payments go through the platform's Stripe account.
func createPlatformCheckoutSession(
	ctx context.Context,
	tenantConfig *storage.TenantStripeConfig,
	req TenantCheckoutSessionRequest,
	successURL, cancelURL string,
	metadata map[string]string,
) (*TenantCheckoutSessionResponse, error) {
	customerID, err := ensureTenantStripeCustomer(ctx, tenantConfig, req.Email, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure tenant customer: %w", err)
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
		return nil, fmt.Errorf("failed to create platform checkout session: %w", err)
	}

	if sess.URL == "" {
		return nil, fmt.Errorf("checkout session has no URL")
	}

	return &TenantCheckoutSessionResponse{
		SessionID:   sess.ID,
		URL:         sess.URL,
		TenantID:    req.TenantID.String(),
		PaymentMode: "platform",
	}, nil
}

// getOrCreateTenantStripeConfig retrieves or creates tenant Stripe configuration.
func getOrCreateTenantStripeConfig(ctx context.Context, repo storage.Repository, tenantID uuid.UUID, email, name string) (*storage.TenantStripeConfig, error) {
	// Try to get existing tenant stripe config
	tenantConfig, err := repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Create default config if not exists
	if tenantConfig == nil {
		// Create a Stripe customer for this tenant
		stripeCustomerID, err := createStripeCustomer(ctx, email, name, map[string]string{
			"tenant_id": tenantID.String(),
			"source":    "tenant_isolation",
		})
		if err != nil {
			logrus.WithError(err).Warn("failed to create Stripe customer for tenant, will use platform customer")
			// Fall back to platform customer
			return &storage.TenantStripeConfig{
				ID:                    uuid.New(),
				TenantID:              tenantID,
				StripeCustomerID:      "",
				IsolatedPaymentEnabled: false,
				PaymentMode:           "platform",
				AllowedPaymentMethods: `["card"]`,
				BillingAddressRequired: true,
				TaxCalculationMode:    "automatic",
				Metadata:              make(storage.JSONMap),
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			}, nil
		}

		// Save tenant stripe config
		tenantConfig = &storage.TenantStripeConfig{
			ID:                    uuid.New(),
			TenantID:              tenantID,
			StripeCustomerID:      stripeCustomerID,
			IsolatedPaymentEnabled: true,
			PaymentMode:           "isolated",
			AllowedPaymentMethods: `["card"]`,
			BillingAddressRequired: true,
			TaxCalculationMode:    "automatic",
			Metadata:              make(storage.JSONMap),
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		if err := repo.CreateTenantStripeConfig(ctx, tenantConfig); err != nil {
			logrus.WithError(err).Warn("failed to save tenant stripe config")
		}
	}

	return tenantConfig, nil
}

// createStripeCustomer creates a new Stripe customer with metadata.
func createStripeCustomer(ctx context.Context, email, name string, metadata map[string]string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	params.Metadata = metadata

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	return cust.ID, nil
}

// ensureTenantStripeCustomer ensures the tenant has a valid Stripe customer ID.
func ensureTenantStripeCustomer(ctx context.Context, tenantConfig *storage.TenantStripeConfig, email, name string) (string, error) {
	if tenantConfig.StripeCustomerID != "" {
		return tenantConfig.StripeCustomerID, nil
	}

	// Create one if missing
	return createStripeCustomer(ctx, email, name, map[string]string{
		"tenant_id": tenantConfig.TenantID.String(),
		"source":    "tenant_isolation",
	})
}

// buildTenantMetadata builds metadata for checkout and subscription.
func buildTenantMetadata(tenantID uuid.UUID, bundleSlug string, founderModeID *uuid.UUID, paymentMode string) map[string]string {
	metadata := map[string]string{
		"tenant_id":    tenantID.String(),
		"bundle_slug":  bundleSlug,
		"payment_mode": paymentMode,
		"isolated":     "true",
	}

	if founderModeID != nil {
		metadata["founder_mode_id"] = founderModeID.String()
	}

	return metadata
}

// parseAllowedPaymentMethods parses the allowed payment methods JSON.
func parseAllowedPaymentMethods(allowedMethodsJSON string) []string {
	if allowedMethodsJSON == "" {
		return []string{"card"}
	}

	var methods []string
	if err := json.Unmarshal([]byte(allowedMethodsJSON), &methods); err != nil {
		logrus.WithError(err).Warn("failed to parse allowed payment methods, defaulting to card")
		return []string{"card"}
	}

	if len(methods) == 0 {
		return []string{"card"}
	}

	return methods
}

// sanitizeTenantReturnURL ensures return URL is safe for the tenant's origin.
func sanitizeTenantReturnURL(returnURL, origin string) string {
	if returnURL == "" {
		return origin
	}

	// Parse the URL using standard library
	parsed, err := url.Parse(returnURL)
	if err != nil {
		return origin
	}

	// Validate scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return origin
	}

	// For SaaS Starter isolated payments, we validate against the origin
	// but allow subpaths
	originParsed, err := url.Parse(origin)
	if err != nil {
		return origin
	}

	// Allow if same host
	if parsed.Host == originParsed.Host {
		return returnURL
	}

	// Default to origin if validation fails
	return origin
}

// GetTenantPaymentStatus returns the payment status for a tenant.
func GetTenantPaymentStatus(ctx context.Context, repo storage.Repository, tenantID uuid.UUID) (*TenantPaymentStatus, error) {
	config, err := repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	status := &TenantPaymentStatus{
		TenantID:              tenantID,
		IsolatedPaymentEnabled: false,
		PaymentMode:          "platform",
	}

	if config != nil {
		status.IsolatedPaymentEnabled = config.IsolatedPaymentEnabled
		status.PaymentMode = config.PaymentMode
		status.StripeCustomerID = config.StripeCustomerID
	}

	return status, nil
}

// TenantPaymentStatus represents the payment status for a tenant.
type TenantPaymentStatus struct {
	TenantID               uuid.UUID `json:"tenant_id"`
	IsolatedPaymentEnabled bool      `json:"isolated_payment_enabled"`
	PaymentMode            string    `json:"payment_mode"`
	StripeCustomerID       string    `json:"stripe_customer_id,omitempty"`
}

// SetTenantPaymentMode sets the payment mode for a tenant.
func SetTenantPaymentMode(ctx context.Context, repo storage.Repository, tenantID uuid.UUID, mode string) error {
	allowedModes := map[string]bool{"platform": true, "isolated": true, "connect": true}
	if !allowedModes[mode] {
		return fmt.Errorf("invalid payment mode: %s", mode)
	}

	config, err := repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return err
	}

	if config == nil {
		return fmt.Errorf("tenant stripe config not found")
	}

	config.PaymentMode = mode
	config.IsolatedPaymentEnabled = mode != "platform"

	return repo.UpdateTenantStripeConfig(ctx, config)
}

// ValidateTenantCheckoutRequest validates a tenant checkout request.
func ValidateTenantCheckoutRequest(req TenantCheckoutSessionRequest) error {
	if req.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.PriceID == "" {
		return fmt.Errorf("price_id is required")
	}
	if req.BundleSlug == "" {
		return fmt.Errorf("bundle_slug is required")
	}

	// Validate bundle slug
	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[req.BundleSlug] {
		return fmt.Errorf("invalid bundle_slug: %s", req.BundleSlug)
	}

	// Validate price ID format
	if !strings.HasPrefix(req.PriceID, "price_") {
		return fmt.Errorf("invalid price_id format: must start with 'price_'")
	}

	return nil
}