package payment

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/paymentintent"
	"github.com/stripe/stripe-go/v83/price"
	"github.com/stripe/stripe-go/v83/product"
)

// TenantPaymentSecurity provides security features for isolated tenant payments.
type TenantPaymentSecurity struct {
	// Webhook signing key for tenant-specific webhooks
	webhookSigningSecrets map[string]string // tenantID -> signing secret
	mu                    sync.RWMutex

	// Rate limiting per tenant
	rateLimits map[string]*tenantRateLimit
	rateMu     sync.Mutex
}

// tenantRateLimit tracks rate limiting for a tenant.
type tenantRateLimit struct {
	Count     int
	ResetAt   time.Time
	MaxCalls  int
	WindowSec int
}

// NewTenantPaymentSecurity creates a new tenant payment security manager.
func NewTenantPaymentSecurity() *TenantPaymentSecurity {
	return &TenantPaymentSecurity{
		webhookSigningSecrets: make(map[string]string),
		rateLimits:            make(map[string]*tenantRateLimit),
	}
}

// GenerateWebhookSigningSecret generates a unique signing secret for a tenant's webhook.
// This is used to verify webhook authenticity for tenant-isolated payment events.
func (s *TenantPaymentSecurity) GenerateWebhookSigningSecret(tenantID uuid.UUID) string {
	secret := generateSecureSecret(32)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhookSigningSecrets[tenantID.String()] = secret
	return secret
}

// GetWebhookSigningSecret retrieves the signing secret for a tenant.
func (s *TenantPaymentSecurity) GetWebhookSigningSecret(tenantID uuid.UUID) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	secret, ok := s.webhookSigningSecrets[tenantID.String()]
	return secret, ok
}

// VerifyWebhookSignature verifies a Stripe webhook signature for tenant-isolated payments.
func (s *TenantPaymentSecurity) VerifyWebhookSignature(tenantID uuid.UUID, payload []byte, sig string, secret string) bool {
	if secret == "" {
		return false
	}

	// Parse the signature header
	// Format: "t=timestamp,v1=signature"
	sigParts := parseSignatureHeader(sig)
	timestamp, ok1 := sigParts["t"]
	signature, ok2 := sigParts["v1"]
	if !ok1 || !ok2 {
		return false
	}

	// Verify timestamp is not too old (5 minute tolerance)
	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return false
	}
	if time.Since(ts) > 5*time.Minute {
		return false
	}

	// Compute expected signature
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// parseSignatureHeader parses the Stripe signature header.
func parseSignatureHeader(header string) map[string]string {
	result := make(map[string]string)
	parts := splitSignatureHeader(header)
	for _, part := range parts {
		kv := splitKV(part)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// splitSignatureHeader splits the signature header by commas.
func splitSignatureHeader(header string) []string {
	var result []string
	var current []byte
	for i := 0; i < len(header); i++ {
		if header[i] == ',' {
			result = append(result, string(current))
			current = nil
		} else if header[i] == '=' && len(current) > 0 && current[len(current)-1] != '\\' {
			// Skip escaped equals
			current = append(current, header[i])
		} else {
			current = append(current, header[i])
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

// splitKV splits a key=value pair.
func splitKV(kv string) []string {
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			return []string{kv[:i], kv[i+1:]}
		}
	}
	return nil
}

// parseTimestamp parses a Unix timestamp string.
func parseTimestamp(ts string) (time.Time, error) {
	var sec int64
	fmt.Sscanf(ts, "%d", &sec)
	return time.Unix(sec, 0), nil
}

// generateSecureSecret generates a cryptographically secure secret.
func generateSecureSecret(length int) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based if crypto/rand fails
		for i := range b {
			b[i] = byte(time.Now().UnixNano() % 256)
		}
	}
	return "whsec_" + hex.EncodeToString(b)[:length]
}

// CheckRateLimit checks if a tenant is within their rate limit.
func (s *TenantPaymentSecurity) CheckRateLimit(tenantID uuid.UUID, operation string) error {
	key := fmt.Sprintf("%s:%s", tenantID.String(), operation)

	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	now := time.Now()
	limit := s.rateLimits[key]

	if limit == nil || now.After(limit.ResetAt) {
		// Create new window
		s.rateLimits[key] = &tenantRateLimit{
			Count:     1,
			ResetAt:   now.Add(60 * time.Second), // 1 minute window
			MaxCalls:  100,                        // 100 calls per minute
			WindowSec: 60,
		}
		return nil
	}

	limit.Count++
	if limit.Count > limit.MaxCalls {
		return fmt.Errorf("rate limit exceeded: %d calls per %d seconds", limit.MaxCalls, limit.WindowSec)
	}

	return nil
}

// TenantPaymentIsolation provides payment isolation for multi-tenant SaaS.
type TenantPaymentIsolation struct {
	security    *TenantPaymentSecurity
	repo        storage.Repository
	webhookURLs map[string]string // tenantID -> webhook URL
	mu          sync.RWMutex
}

// NewTenantPaymentIsolation creates a new tenant payment isolation manager.
func NewTenantPaymentIsolation(repo storage.Repository) *TenantPaymentIsolation {
	return &TenantPaymentIsolation{
		security:    NewTenantPaymentSecurity(),
		repo:        repo,
		webhookURLs: make(map[string]string),
	}
}

// ProvisionTenantPayment provisions isolated payment infrastructure for a tenant.
// This creates a tenant-specific Stripe configuration and webhook endpoint.
func (p *TenantPaymentIsolation) ProvisionTenantPayment(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (*TenantPaymentProvision, error) {
	// Check if already provisioned
	existing, err := p.repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing config: %w", err)
	}
	if existing != nil && existing.IsolatedPaymentEnabled {
		return &TenantPaymentProvision{
			TenantID:               tenantID,
			StripeCustomerID:       existing.StripeCustomerID,
			IsolatedPaymentEnabled: existing.IsolatedPaymentEnabled,
			PaymentMode:            existing.PaymentMode,
			WebhookURL:             p.getWebhookURL(tenantID),
			WebhookSigningSecret:   p.getWebhookSigningSecret(tenantID),
		}, nil
	}

	// Determine payment mode based on bundle
	paymentMode := determinePaymentMode(bundleSlug)

	// Create tenant Stripe customer
	stripeCustomerID, err := p.createTenantStripeCustomer(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant Stripe customer: %w", err)
	}

	// Generate webhook signing secret
	webhookSecret := p.security.GenerateWebhookSigningSecret(tenantID)
	webhookURL := p.generateWebhookURL(tenantID)

	// Create tenant Stripe config
	tenantConfig := &storage.TenantStripeConfig{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		StripeCustomerID:       stripeCustomerID,
		IsolatedPaymentEnabled: paymentMode != "platform",
		PaymentMode:            paymentMode,
		AllowedPaymentMethods:  getAllowedPaymentMethods(bundleSlug),
		BillingAddressRequired: true,
		TaxCalculationMode:     "automatic",
		Metadata:               p.buildTenantMetadata(bundleSlug, webhookURL),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if err := p.repo.CreateTenantStripeConfig(ctx, tenantConfig); err != nil {
		return nil, fmt.Errorf("failed to save tenant stripe config: %w", err)
	}

	// Store webhook URL
	p.mu.Lock()
	p.webhookURLs[tenantID.String()] = webhookURL
	p.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"bundle":    bundleSlug,
		"mode":      paymentMode,
	}).Info("Tenant payment infrastructure provisioned")

	return &TenantPaymentProvision{
		TenantID:               tenantID,
		StripeCustomerID:       stripeCustomerID,
		IsolatedPaymentEnabled: paymentMode != "platform",
		PaymentMode:            paymentMode,
		WebhookURL:             webhookURL,
		WebhookSigningSecret:   webhookSecret,
	}, nil
}

// TenantPaymentProvision contains the provisioned payment infrastructure for a tenant.
type TenantPaymentProvision struct {
	TenantID               uuid.UUID `json:"tenant_id"`
	StripeCustomerID       string    `json:"stripe_customer_id"`
	IsolatedPaymentEnabled  bool      `json:"isolated_payment_enabled"`
	PaymentMode            string    `json:"payment_mode"`
	WebhookURL             string    `json:"webhook_url"`
	WebhookSigningSecret   string    `json:"webhook_signing_secret"`
}

// createTenantStripeCustomer creates a Stripe customer for a tenant.
func (p *TenantPaymentIsolation) createTenantStripeCustomer(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// Get tenant info
	tenant, err := p.repo.GetTenantByID(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant: %w", err)
	}

	// Get first user for email
	users, err := p.repo.ListActiveUsersByTenant(ctx, tenantID)
	if err != nil || len(users) == 0 {
		return "", fmt.Errorf("no active users found for tenant")
	}

	user := users[0]
	email := user.Email
	name := user.Name
	if name == "" {
		name = tenant.Name
	}

	cust, err := CreateStripeCustomer(ctx, email, name, map[string]string{
		"tenant_id": tenantID.String(),
		"source":    "functionfly_isolation",
		"tier":      "isolated",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	return cust.ID, nil
}

// generateWebhookURL generates a unique webhook URL for a tenant.
func (p *TenantPaymentIsolation) generateWebhookURL(tenantID uuid.UUID) string {
	// In production, this would be a unique URL per tenant
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "https://api.functionfly.io"
	}
	return fmt.Sprintf("%s/v1/billing/tenants/%s/webhook", baseURL, tenantID.String())
}

// getWebhookURL retrieves the webhook URL for a tenant.
func (p *TenantPaymentIsolation) getWebhookURL(tenantID uuid.UUID) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if url, ok := p.webhookURLs[tenantID.String()]; ok {
		return url
	}
	return p.generateWebhookURL(tenantID)
}

// getWebhookSigningSecret retrieves the signing secret for a tenant.
func (p *TenantPaymentIsolation) getWebhookSigningSecret(tenantID uuid.UUID) string {
	secret, _ := p.security.GetWebhookSigningSecret(tenantID)
	return secret
}

// determinePaymentMode determines the payment mode based on bundle type.
func determinePaymentMode(bundleSlug string) string {
	switch bundleSlug {
	case "marketplace":
		return "connect" // Full Stripe Connect for marketplace
	case "saas-starter":
		return "isolated" // Isolated but platform-managed
	case "ai-app":
		return "isolated" // Isolated with AI-specific config
	default:
		return "platform" // Default to platform mode
	}
}

// getAllowedPaymentMethods returns allowed payment methods based on bundle type.
func getAllowedPaymentMethods(bundleSlug string) string {
	switch bundleSlug {
	case "marketplace":
		return `["card", "us_bank_account", "eu_bank_transfer"]`
	case "saas-starter":
		return `["card"]`
	case "ai-app":
		return `["card", "us_bank_account"]`
	default:
		return `["card"]`
	}
}

// buildTenantMetadata builds metadata for tenant Stripe configuration.
func (p *TenantPaymentIsolation) buildTenantMetadata(bundleSlug, webhookURL string) storage.JSONMap {
	return storage.JSONMap{
		"bundle_slug":    bundleSlug,
		"webhook_url":    webhookURL,
		"provisioned_at": time.Now().UTC().Format(time.RFC3339),
		"version":        "1.0",
	}
}

// CreateIsolatedPaymentIntent creates a PaymentIntent for tenant-isolated payments.
// This is used for one-time payments within the tenant's isolated environment.
func (p *TenantPaymentIsolation) CreateIsolatedPaymentIntent(
	ctx context.Context,
	tenantID uuid.UUID,
	amountCents int,
	currency string,
	metadata map[string]string,
) (*IsolatedPaymentIntent, error) {
	// Check rate limit
	if err := p.security.CheckRateLimit(tenantID, "create_payment_intent"); err != nil {
		return nil, err
	}

	// Get tenant Stripe config
	config, err := p.repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("tenant payment not provisioned")
	}

	// Build metadata with tenant isolation markers
	isoMetadata := map[string]string{
		"tenant_id":          tenantID.String(),
		"isolated_payment":    "true",
		"payment_mode":        config.PaymentMode,
		"tenant_customer_id":  config.StripeCustomerID,
	}
	for k, v := range metadata {
		isoMetadata[k] = v
	}

	// Create PaymentIntent
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(int64(amountCents)),
		Currency:      stripe.String(currency),
		Customer:      stripe.String(config.StripeCustomerID),
		Metadata:      isoMetadata,
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	return &IsolatedPaymentIntent{
		ID:               pi.ID,
		ClientSecret:     pi.ClientSecret,
		Amount:           int(pi.Amount),
		Currency:         string(pi.Currency),
		Status:           string(pi.Status),
		TenantID:         tenantID,
		StripeCustomerID: config.StripeCustomerID,
		PaymentMode:      config.PaymentMode,
	}, nil
}

// IsolatedPaymentIntent represents a payment intent for isolated tenant payments.
type IsolatedPaymentIntent struct {
	ID               string    `json:"id"`
	ClientSecret     string    `json:"client_secret"`
	Amount           int       `json:"amount"`
	Currency         string    `json:"currency"`
	Status           string    `json:"status"`
	TenantID         uuid.UUID `json:"tenant_id"`
	StripeCustomerID string    `json:"stripe_customer_id"`
	PaymentMode      string    `json:"payment_mode"`
}

// GetTenantPaymentConfig retrieves the payment configuration for a tenant.
func (p *TenantPaymentIsolation) GetTenantPaymentConfig(ctx context.Context, tenantID uuid.UUID) (*TenantPaymentConfigResponse, error) {
	config, err := p.repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	response := &TenantPaymentConfigResponse{
		TenantID:               tenantID,
		IsolatedPaymentEnabled: false,
		PaymentMode:           "platform",
	}

	if config != nil {
		response.IsolatedPaymentEnabled = config.IsolatedPaymentEnabled
		response.PaymentMode = config.PaymentMode
		response.StripeCustomerID = config.StripeCustomerID
		response.WebhookURL = p.getWebhookURL(tenantID)
		response.AllowedPaymentMethods = parseAllowedPaymentMethods(config.AllowedPaymentMethods)
	}

	return response, nil
}

// TenantPaymentConfigResponse represents a tenant's payment configuration.
type TenantPaymentConfigResponse struct {
	TenantID               uuid.UUID `json:"tenant_id"`
	IsolatedPaymentEnabled  bool      `json:"isolated_payment_enabled"`
	PaymentMode            string    `json:"payment_mode"`
	StripeCustomerID       string    `json:"stripe_customer_id,omitempty"`
	WebhookURL             string    `json:"webhook_url,omitempty"`
	AllowedPaymentMethods  []string  `json:"allowed_payment_methods,omitempty"`
}

// DisableIsolatedPayment disables isolated payments for a tenant and reverts to platform mode.
func (p *TenantPaymentIsolation) DisableIsolatedPayment(ctx context.Context, tenantID uuid.UUID) error {
	config, err := p.repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("tenant stripe config not found")
	}

	config.IsolatedPaymentEnabled = false
	config.PaymentMode = "platform"

	return p.repo.UpdateTenantStripeConfig(ctx, config)
}

// ValidateIsolatedPayment validates that a payment belongs to a tenant.
func (p *TenantPaymentIsolation) ValidateIsolatedPayment(tenantID uuid.UUID, paymentMeta map[string]string) bool {
	// Verify tenant ID in metadata
	if tid, ok := paymentMeta["tenant_id"]; ok {
		return tid == tenantID.String()
	}
	return false
}

// GetBundlePaymentRequirements returns payment requirements for a bundle.
func GetBundlePaymentRequirements(bundleSlug string) *BundlePaymentRequirements {
	reqs := &BundlePaymentRequirements{
		BundleSlug:            bundleSlug,
		IsolatedPayments:      true,
		TaxCalculation:        "automatic",
		BillingAddress:        true,
		PaymentMethodTypes:    []string{"card"},
		RequiresOnboarding:    false,
		ConnectAccountEnabled: false,
	}

	switch bundleSlug {
	case "marketplace":
		reqs.IsolatedPayments = true
		reqs.RequiresOnboarding = true
		reqs.ConnectAccountEnabled = true
		reqs.PaymentMethodTypes = []string{"card", "us_bank_account", "eu_bank_transfer"}
	case "saas-starter":
		reqs.IsolatedPayments = true
		reqs.TaxCalculation = "automatic"
		reqs.PaymentMethodTypes = []string{"card"}
	case "ai-app":
		reqs.IsolatedPayments = true
		reqs.TaxCalculation = "automatic"
		reqs.PaymentMethodTypes = []string{"card", "us_bank_account"}
	}

	return reqs
}

// BundlePaymentRequirements defines payment requirements for a bundle type.
type BundlePaymentRequirements struct {
	BundleSlug            string   `json:"bundle_slug"`
	IsolatedPayments      bool     `json:"isolated_payments"`
	TaxCalculation        string   `json:"tax_calculation"` // "automatic" or "manual"
	BillingAddress        bool     `json:"billing_address"`
	PaymentMethodTypes    []string `json:"payment_method_types"`
	RequiresOnboarding    bool     `json:"requires_onboarding"`
	ConnectAccountEnabled bool     `json:"connect_account_enabled"`
}

// GetStripePrice retrieves a Stripe price by ID with tenant isolation context.
func GetStripePrice(ctx context.Context, priceID string) (*stripe.Price, error) {
	if !strings.HasPrefix(priceID, "price_") {
		return nil, fmt.Errorf("invalid price ID format: must start with 'price_'")
	}

	p, err := price.Get(priceID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe price: %w", err)
	}

	return p, nil
}

// GetStripeProduct retrieves a Stripe product by ID.
func GetStripeProduct(ctx context.Context, productID string) (*stripe.Product, error) {
	if !strings.HasPrefix(productID, "prod_") {
		return nil, fmt.Errorf("invalid product ID format: must start with 'prod_'")
	}

	prod, err := product.Get(productID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe product: %w", err)
	}

	return prod, nil
}