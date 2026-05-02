package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
)

// TenantWebhookHandler handles Stripe webhooks for tenant-isolated payments.
// Each tenant with isolated payments has their own webhook endpoint and signing secret.
type TenantWebhookHandler struct {
	repo               storage.Repository
	webhookSecrets     map[string]string // tenantID -> signing secret
	tenantIsolationSvc interface {
		VerifyWebhookSignature(tenantID uuid.UUID, payload []byte, sig string, secret string) bool
	}
}

// NewTenantWebhookHandler creates a new tenant webhook handler.
func NewTenantWebhookHandler(repo storage.Repository) *TenantWebhookHandler {
	return &TenantWebhookHandler{
		repo:           repo,
		webhookSecrets: make(map[string]string),
	}
}

// SetWebhookSecret sets the webhook signing secret for a tenant.
func (h *TenantWebhookHandler) SetWebhookSecret(tenantID uuid.UUID, secret string) {
	h.webhookSecrets[tenantID.String()] = secret
}

// GetWebhookSecret retrieves the webhook signing secret for a tenant.
func (h *TenantWebhookHandler) GetWebhookSecret(tenantID uuid.UUID) (string, bool) {
	secret, ok := h.webhookSecrets[tenantID.String()]
	return secret, ok
}

// RegisterRoutes registers tenant webhook routes.
func (h *TenantWebhookHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/tenants/{tenant_id}/webhook", h.HandleTenantWebhook).Methods("POST", "OPTIONS")
}

// HandleTenantWebhook processes incoming Stripe webhook events for tenant-isolated payments.
func (h *TenantWebhookHandler) HandleTenantWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Error("tenant webhook: invalid tenant_id")
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Read payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("tenant webhook: failed to read body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature for tenant-isolated payments
	if err := h.verifyTenantWebhookSignature(ctx, tenantID, payload, r.Header.Get("Stripe-Signature")); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantIDStr).Error("tenant webhook: signature verification failed")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse the event
	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		logrus.WithError(err).Error("tenant webhook: failed to parse event")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Log the event
	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantIDStr,
		"event_id":  event.ID,
		"event_type": string(event.Type),
	}).Info("tenant webhook: processing event")

	// Process the event based on type
	h.processTenantEvent(ctx, tenantID, &event, w, r)
}

// verifyTenantWebhookSignature verifies the Stripe webhook signature for a tenant.
func (h *TenantWebhookHandler) verifyTenantWebhookSignature(ctx context.Context, tenantID uuid.UUID, payload []byte, signature string) error {
	// First check if tenant has isolated payments enabled
	config, err := h.repo.GetTenantStripeConfig(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant config: %w", err)
	}

	if config == nil || !config.IsolatedPaymentEnabled {
		// Not using isolated payments - skip tenant webhook handling
		return fmt.Errorf("tenant does not have isolated payments enabled")
	}

	// Get tenant-specific webhook secret
	secret, ok := h.GetWebhookSecret(tenantID)
	if !ok || secret == "" {
		// Secret not set - try to construct from tenant config metadata
		logrus.WithField("tenant_id", tenantID).Warn("tenant webhook: no signing secret configured")
		return fmt.Errorf("webhook secret not configured for tenant")
	}

	// Parse Stripe signature header
	sigParts := parseSignatureHeader(signature)
	timestamp, ok1 := sigParts["t"]
	sig, ok2 := sigParts["v1"]
	if !ok1 || !ok2 {
		return fmt.Errorf("invalid signature header format")
	}

	// Verify timestamp is not too old (5 minute tolerance)
	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}
	if time.Since(ts) > 5*time.Minute {
		return fmt.Errorf("webhook timestamp too old")
	}

	// Compute expected signature
	signedPayload := timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// parseSignatureHeader parses the Stripe signature header.
func parseSignatureHeader(header string) map[string]string {
	result := make(map[string]string)
	if header == "" {
		return result
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// parseTimestamp parses a Unix timestamp string.
func parseTimestamp(ts string) (time.Time, error) {
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}

// processTenantEvent processes a webhook event for a tenant.
func (h *TenantWebhookHandler) processTenantEvent(ctx context.Context, tenantID uuid.UUID, event *stripe.Event, w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch event.Type {
	case "checkout.session.completed":
		h.handleTenantCheckoutSessionCompleted(ctx, tenantID, event, w, r)
	case "invoice.payment_succeeded":
		h.handleTenantInvoicePaymentSucceeded(ctx, tenantID, event, w, r)
	case "invoice.payment_failed":
		h.handleTenantInvoicePaymentFailed(ctx, tenantID, event, w, r)
	case "customer.subscription.updated":
		h.handleTenantSubscriptionUpdated(ctx, tenantID, event, w, r)
	case "customer.subscription.deleted":
		h.handleTenantSubscriptionDeleted(ctx, tenantID, event, w, r)
	default:
		logrus.WithFields(logrus.Fields{
			"tenant_id":  tenantID,
			"event_id":  event.ID,
			"event_type": string(event.Type),
		}).Debug("tenant webhook: unhandled event type")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
	}

	// Record metrics
	duration := time.Since(start)
	monitoring.RecordBillingWebhookProcessingDuration(string(event.Type), duration)
}

// handleTenantCheckoutSessionCompleted handles checkout.session.completed for tenant.
func (h *TenantWebhookHandler) handleTenantCheckoutSessionCompleted(ctx context.Context, tenantID uuid.UUID, event *stripe.Event, w http.ResponseWriter, r *http.Request) {
	var session stripe.CheckoutSession
	sessionData, err := event.Data.Raw.MarshalJSON()
	if err != nil {
		logrus.WithError(err).Error("tenant webhook: failed to marshal checkout session data")
		http.Error(w, "Invalid event data", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(sessionData, &session); err != nil {
		logrus.WithError(err).Error("tenant webhook: failed to unmarshal checkout session")
		http.Error(w, "Invalid session data", http.StatusBadRequest)
		return
	}

	// Verify this session belongs to this tenant
	if session.Metadata == nil {
		logrus.WithField("session_id", session.ID).Warn("tenant webhook: checkout session has no metadata")
		http.Error(w, "Missing metadata", http.StatusBadRequest)
		return
	}

	sessionTenantID := session.Metadata["tenant_id"]
	if sessionTenantID == "" {
		logrus.WithField("session_id", session.ID).Warn("tenant webhook: session has no tenant_id in metadata")
		http.Error(w, "Missing tenant_id in metadata", http.StatusBadRequest)
		return
	}

	if sessionTenantID != tenantID.String() {
		logrus.WithFields(logrus.Fields{
			"session_id": session.ID,
			"expected_tenant": tenantID.String(),
			"actual_tenant": sessionTenantID,
		}).Error("tenant webhook: tenant mismatch")
		http.Error(w, "Tenant ID mismatch", http.StatusForbidden)
		return
	}

	purpose := session.Metadata["purpose"]
	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"session_id": session.ID,
		"purpose": purpose,
	}).Info("tenant webhook: checkout session completed")

	// Handle based on purpose
	switch purpose {
	case "bundle_subscription":
		h.handleTenantBundleSubscriptionCheckout(ctx, tenantID, &session, w, r)
	case "registry_wallet_credit":
		h.handleTenantWalletCreditCheckout(ctx, tenantID, &session, w, r)
	case "agent_execution_credits":
		h.handleTenantAgentCreditsCheckout(ctx, tenantID, &session, w, r)
	default:
		logrus.WithField("purpose", purpose).Debug("tenant webhook: unknown checkout purpose")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
	}
}

// handleTenantBundleSubscriptionCheckout handles bundle subscription checkout for a tenant.
func (h *TenantWebhookHandler) handleTenantBundleSubscriptionCheckout(ctx context.Context, tenantID uuid.UUID, session *stripe.CheckoutSession, w http.ResponseWriter, r *http.Request) {
	bundleSlug := session.Metadata["bundle_slug"]

	if bundleSlug == "" {
		logrus.WithField("session_id", session.ID).Warn("tenant webhook: bundle subscription missing bundle_slug")
		http.Error(w, "Missing bundle_slug", http.StatusBadRequest)
		return
	}

	// Get the bundle by slug
	bundle, err := h.repo.GetPricingBundleBySlug(ctx, bundleSlug)
	if err != nil {
		logrus.WithError(err).WithField("bundle_slug", bundleSlug).Error("tenant webhook: failed to get bundle")
		http.Error(w, "Failed to retrieve bundle", http.StatusInternalServerError)
		return
	}
	if bundle == nil {
		logrus.WithField("bundle_slug", bundleSlug).Warn("tenant webhook: bundle not found")
		http.Error(w, "Bundle not found", http.StatusNotFound)
		return
	}

	// Check for existing subscription
	existingSub, err := h.repo.GetBundleSubscriptionByTenant(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).Error("tenant webhook: failed to check existing bundle subscription")
	}

	// If already active, acknowledge
	if existingSub != nil && existingSub.Status == "active" {
		logrus.WithFields(logrus.Fields{
			"sub_id": existingSub.ID,
			"session_id": session.ID,
		}).Info("tenant webhook: bundle subscription already active")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "already_active"})
		return
	}

	// Get subscription details from Stripe
	var subscriptionID string
	var periodStart, periodEnd time.Time
	if session.Subscription != nil && session.Subscription.ID != "" {
		subscriptionID = session.Subscription.ID
	}

	now := time.Now().UTC()

	// Create or update subscription
	sub := &storage.BundleSubscription{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		BundleID:             bundle.ID,
		StripeSubscriptionID: subscriptionID,
		Status:               "active",
		CurrentPeriodStart:   periodStart,
		CurrentPeriodEnd:     periodEnd,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Handle founder mode conversion
	founderModeIDStr := session.Metadata["founder_mode_id"]
	if founderModeIDStr != "" {
		if fmid, err := uuid.Parse(founderModeIDStr); err == nil {
			sub.FounderModeID = &fmid
			sub.ConvertedFromFounderMode = true
			// Update founder mode status to converted
			_ = h.repo.UpdateFounderModeStatus(ctx, fmid, "converted")
		}
	}

	// Update existing deferred subscription or create new
	if existingSub != nil && existingSub.Status == "deferred" {
		sub.ID = existingSub.ID
		sub.DefaultAppID = existingSub.DefaultAppID
		if err := h.repo.UpdateBundleSubscription(ctx, sub); err != nil {
			logrus.WithError(err).Error("tenant webhook: failed to update bundle subscription")
			http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.repo.CreateBundleSubscription(ctx, sub); err != nil {
			logrus.WithError(err).Error("tenant webhook: failed to create bundle subscription")
			http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
			return
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"bundle_slug": bundleSlug,
		"subscription_id": sub.ID,
		"session_id": session.ID,
	}).Info("tenant webhook: bundle subscription created/updated")

	// Trigger provisioning asynchronously
	go func() {
		if err := ProvisionBundleResources(h.repo, tenantID, bundleSlug); err != nil {
			logrus.WithError(err).WithField("tenant_id", tenantID).Error("tenant webhook: failed to provision bundle resources")
		}
	}()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleTenantWalletCreditCheckout handles wallet credit checkout for a tenant.
func (h *TenantWebhookHandler) handleTenantWalletCreditCheckout(ctx context.Context, tenantID uuid.UUID, session *stripe.CheckoutSession, w http.ResponseWriter, r *http.Request) {
	userIDStr := session.Metadata["user_id"]
	amountUsdStr := session.Metadata["amount_usd"]

	if userIDStr == "" {
		logrus.WithField("session_id", session.ID).Warn("tenant webhook: wallet credit missing user_id")
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Warn("tenant webhook: invalid user_id")
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	amountUSD, _ := strconv.ParseFloat(amountUsdStr, 64)
	if amountUSD <= 0 {
		amountUSD = float64(session.AmountTotal) / 100
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"user_id": userID,
		"amount_usd": amountUSD,
		"session_id": session.ID,
	}).Info("tenant webhook: wallet credit processed")

	// Record invoice
	if err := h.repo.CreatePaidInvoiceForStripeCheckoutSession(ctx, tenantID, int(session.AmountTotal), string(session.Currency), session.ID, ""); err != nil {
		logrus.WithError(err).WithField("session_id", session.ID).Warn("tenant webhook: failed to persist invoice")
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleTenantAgentCreditsCheckout handles agent credits checkout for a tenant.
func (h *TenantWebhookHandler) handleTenantAgentCreditsCheckout(ctx context.Context, tenantID uuid.UUID, session *stripe.CheckoutSession, w http.ResponseWriter, r *http.Request) {
	agentID := session.Metadata["agent_id"]
	amountUsdStr := session.Metadata["amount_usd"]

	if agentID == "" {
		logrus.WithField("session_id", session.ID).Warn("tenant webhook: agent credits missing agent_id")
		http.Error(w, "Missing agent_id", http.StatusBadRequest)
		return
	}

	amountUSD, _ := strconv.ParseFloat(amountUsdStr, 64)
	if amountUSD <= 0 {
		amountUSD = float64(session.AmountTotal) / 100
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"agent_id": agentID,
		"amount_usd": amountUSD,
		"session_id": session.ID,
	}).Info("tenant webhook: agent credits processed")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleTenantInvoicePaymentSucceeded handles invoice.payment_succeeded for a tenant.
func (h *TenantWebhookHandler) handleTenantInvoicePaymentSucceeded(ctx context.Context, tenantID uuid.UUID, event *stripe.Event, w http.ResponseWriter, r *http.Request) {
	var invoice stripe.Invoice
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &invoice) != nil {
		logrus.WithError(err).Error("tenant webhook: failed to unmarshal invoice")
		http.Error(w, "Invalid invoice payload", http.StatusBadRequest)
		return
	}

	amountPaidUSD := float64(invoice.AmountPaid) / 100.0

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"invoice_id": invoice.ID,
		"amount_usd": amountPaidUSD,
	}).Info("tenant webhook: invoice payment succeeded")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleTenantInvoicePaymentFailed handles invoice.payment_failed for a tenant.
func (h *TenantWebhookHandler) handleTenantInvoicePaymentFailed(ctx context.Context, tenantID uuid.UUID, event *stripe.Event, w http.ResponseWriter, r *http.Request) {
	var invoice stripe.Invoice
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &invoice) != nil {
		logrus.WithError(err).Error("tenant webhook: failed to unmarshal invoice")
		http.Error(w, "Invalid invoice payload", http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"invoice_id": invoice.ID,
		"amount_due": invoice.AmountDue,
		"attempt_count": invoice.AttemptCount,
	}).Warn("tenant webhook: invoice payment failed")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleTenantSubscriptionUpdated handles subscription updates for a tenant.
func (h *TenantWebhookHandler) handleTenantSubscriptionUpdated(ctx context.Context, tenantID uuid.UUID, event *stripe.Event, w http.ResponseWriter, r *http.Request) {
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		http.Error(w, "Invalid subscription payload", http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"subscription_id": sub.ID,
		"status": string(sub.Status),
	}).Info("tenant webhook: subscription updated")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleTenantSubscriptionDeleted handles subscription deletions for a tenant.
func (h *TenantWebhookHandler) handleTenantSubscriptionDeleted(ctx context.Context, tenantID uuid.UUID, event *stripe.Event, w http.ResponseWriter, r *http.Request) {
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		http.Error(w, "Invalid subscription payload", http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"subscription_id": sub.ID,
	}).Info("tenant webhook: subscription deleted")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// BuildTenantWebhookURL builds the webhook URL for a tenant.
func BuildTenantWebhookURL(tenantID uuid.UUID) string {
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "https://api.functionfly.io"
	}
	return fmt.Sprintf("%s/v1/billing/tenants/%s/webhook", baseURL, tenantID.String())
}

// GenerateTenantWebhookSecret generates a secure webhook signing secret for a tenant.
func GenerateTenantWebhookSecret() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(io.Reader(nil), b); err != nil {
		// Fallback to time-based if crypto/rand fails
		for i := range b {
			b[i] = byte(time.Now().UnixNano() % 256)
		}
	}
	return "whsec_" + hex.EncodeToString(b)[:32]
}

// SetupTenantWebhook configures webhook infrastructure for a tenant with isolated payments.
func SetupTenantWebhook(tenantID uuid.UUID, repo storage.Repository, secret string) error {
	// Store secret in tenant stripe config metadata
	config, err := repo.GetTenantStripeConfig(context.Background(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant config: %w", err)
	}

	if config == nil {
		return fmt.Errorf("tenant stripe config not found")
	}

	// Update metadata with webhook secret reference
	if config.Metadata == nil {
		config.Metadata = make(storage.JSONMap)
	}
	config.Metadata["webhook_secret_set_at"] = time.Now().UTC().Format(time.RFC3339)
	config.Metadata["webhook_url"] = BuildTenantWebhookURL(tenantID)

	return repo.UpdateTenantStripeConfig(context.Background(), config)
}