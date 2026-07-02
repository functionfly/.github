package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	agentbilling "github.com/functionfly/functionfly/internal/agent/billing"
	billing "github.com/functionfly/functionfly/internal/api/handlers/billing"
	"github.com/functionfly/functionfly/internal/api/helpers"
	"github.com/functionfly/functionfly/internal/apierror"
	billingpkg "github.com/functionfly/functionfly/internal/billing"
	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/verification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/invoice"
	stripeSub "github.com/stripe/stripe-go/v83/subscription"
	"github.com/stripe/stripe-go/v83/webhook"
	"gorm.io/datatypes"
)

// piiSensitiveFields contains metadata keys that may contain PII
var piiSensitiveFields = []string{
	"user_id",
	"initiating_user_id",
	"new_username",
	"email",
	"customer_email",
	"phone",
	"address",
	"name",
	"first_name",
	"last_name",
	"full_name",
}

// sanitizeMetadataForLogging removes PII from metadata for safe logging
func sanitizeMetadataForLogging(metadata map[string]string) map[string]string {
	if metadata == nil {
		return nil
	}
	sanitized := make(map[string]string, len(metadata))
	for key, value := range metadata {
		lowerKey := strings.ToLower(key)
		isPII := false
		for _, piiField := range piiSensitiveFields {
			if lowerKey == piiField {
				isPII = true
				break
			}
		}
		if isPII {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = value
		}
	}
	return sanitized
}

// StripeWebhookHandler handles Stripe webhook events for payment processing.
type StripeWebhookHandler struct {
	financialTxRepo *storage.FinancialTransactionRepository
	billingCtrl     *agentbilling.Controller
	notificationSvc *notification.Service
	userRepo        storage.Repository
	platformFees    *storageregistry.PlatformFeeRepository
	sfAddons        *statefabricaddons.Repository
	disputeRepo     *storage.DisputeRepository
	refundRepo      *storage.RefundRepository
	registryRepo    *storageregistry.RegistryRepository
	webhookSecret   string
	emailSvc        email.Service
	dunningManager  *billingpkg.DunningManager
	operationalRepo *storage.BillingOperationalRepository
	payoutService   PayoutWebhookProcessor
	certRepo        *storage.CertificationRepository
	disputeResponseManager *billingpkg.DisputeResponseManager
	pciAuditHelper  *helpers.PCIAuditHelper
	// provisionBundleFn delegates to the isolated BundleProvisioner for dedicated tenant DB provisioning.
	provisionBundleFn func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error)
}

// PayoutWebhookProcessor handles payout-related webhook events.
type PayoutWebhookProcessor interface {
	ProcessTransferReversed(ctx context.Context, stripeTransferID string) error
	ProcessPayoutPaid(ctx context.Context, stripePayoutID, stripeAccountID string) error
	RefreshAccountStatus(ctx context.Context, stripeAccountID string) error
}

// StripeDisputeEvent represents a Stripe dispute event payload
type StripeDisputeEvent struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Reason   string `json:"reason"`
	Status   string `json:"status"`
	Outcome  *struct {
		NetworkReasonCode string `json:"network_reason_code,omitempty"`
		Reason            string `json:"reason,omitempty"`
	} `json:"outcome,omitempty"`
	EvidenceDetails *struct {
		DueBy int64 `json:"due_by,omitempty"`
	} `json:"evidence_details,omitempty"`
	Charge struct {
		ID string `json:"id"`
	} `json:"charge"`
	PaymentIntent *struct {
		ID string `json:"id"`
	} `json:"payment_intent,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewStripeWebhookHandler creates a new Stripe webhook handler.
func NewStripeWebhookHandler(
	financialTxRepo *storage.FinancialTransactionRepository,
	billingCtrl *agentbilling.Controller,
	notificationSvc *notification.Service,
	userRepo storage.Repository,
	platformFees *storageregistry.PlatformFeeRepository,
	sfAddons *statefabricaddons.Repository,
	disputeRepo *storage.DisputeRepository,
	refundRepo *storage.RefundRepository,
	registryRepo *storageregistry.RegistryRepository,
	emailSvc email.Service,
	certRepo *storage.CertificationRepository,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		financialTxRepo: financialTxRepo,
		billingCtrl:     billingCtrl,
		notificationSvc: notificationSvc,
		userRepo:        userRepo,
		platformFees:    platformFees,
		sfAddons:        sfAddons,
		webhookSecret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
		disputeRepo:     disputeRepo,
		refundRepo:      refundRepo,
		registryRepo:    registryRepo,
		emailSvc:        emailSvc,
		certRepo:        certRepo,
	}
}

// SetDisputeRepository sets the dispute repository (for use when repo not available at construction)
func (h *StripeWebhookHandler) SetDisputeRepository(repo *storage.DisputeRepository) {
	h.disputeRepo = repo
}

// SetRefundRepository sets the refund repository (for use when repo not available at construction)
func (h *StripeWebhookHandler) SetRefundRepository(repo *storage.RefundRepository) {
	h.refundRepo = repo
}

// SetDunningManager sets the dunning manager for automated payment retry
func (h *StripeWebhookHandler) SetDunningManager(dm *billingpkg.DunningManager) {
	h.dunningManager = dm
}

// SetBundleProvisioner injects the isolated bundle provisioner for dedicated tenant DB provisioning.
func (h *StripeWebhookHandler) SetBundleProvisioner(fn func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error)) {
	h.provisionBundleFn = fn
}

// SetOperationalRepository sets the billing operational repository for webhook storage
func (h *StripeWebhookHandler) SetOperationalRepository(repo *storage.BillingOperationalRepository) {
	h.operationalRepo = repo
}

// SetDisputeResponseManager sets the dispute response manager for automated chargeback handling
func (h *StripeWebhookHandler) SetDisputeResponseManager(drm *billingpkg.DisputeResponseManager) {
	h.disputeResponseManager = drm
}

// SetPCIAuditHelper sets the PCI audit helper for compliance logging
func (h *StripeWebhookHandler) SetPCIAuditHelper(pciAudit *helpers.PCIAuditHelper) {
	h.pciAuditHelper = pciAudit
}

// RegisterRoutes registers webhook routes.
func (h *StripeWebhookHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/webhooks/stripe", h.HandleWebhook).Methods("POST")
}

// HandleWebhook processes incoming Stripe webhook events.
// Security: Webhook signature verification is MANDATORY in production.
// The ALLOW_UNVERIFIED_WEBHOOKS environment variable is IGNORED in production.
func (h *StripeWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("path", r.URL.Path).WithField("method", r.Method).Info("stripe webhook handler hit")

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("failed to read webhook body")
		apierror.WriteError(w, apierror.NewBadRequest("Failed to read request body"))
		return
	}
	defer r.Body.Close()

	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		logrus.WithError(err).Warn("failed to unmarshal stripe webhook event before idempotency")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}
	logrus.WithFields(logrus.Fields{
		"event_type": event.Type,
		"event_id":   event.ID,
		"livemode":   event.Livemode,
	}).Info("stripe webhook: received event before idempotency")

	// SECURITY: Strict production check - NEVER allow unverified webhooks in production
	isProduction := os.Getenv("PRODUCTION") == "true"

	// In production, webhook secret is MANDATORY - no exceptions
	if isProduction && h.webhookSecret == "" {
		logrus.Error("SECURITY: STRIPE_WEBHOOK_SECRET not configured in production - rejecting webhook. This is a critical security requirement.")
		apierror.WriteError(w, apierror.NewInternal("Webhook authentication not configured"))
		return
	}

	// Parse the event to get the event ID for idempotency check
	if h.webhookSecret == "" {
		// In non-production, only allow unverified webhooks if explicitly opted in via both env vars
		// Note: ALLOW_UNVERIFIED_WEBHOOKS is IGNORED in production due to check above
		isDev := os.Getenv("DEVELOPMENT") == "true"
		if os.Getenv("ALLOW_UNVERIFIED_WEBHOOKS") != "true" || !isDev {
			logrus.Error("STRIPE_WEBHOOK_SECRET not configured - rejecting webhook. Set ALLOW_UNVERIFIED_WEBHOOKS=true and DEVELOPMENT=true to allow unverified webhooks in development only")
			apierror.WriteError(w, apierror.NewInternal("Webhook authentication not configured"))
			return
		}
		// Development mode with explicit opt-in: parse without verification
		logrus.Warn("Processing unverified webhook - ALLOW_UNVERIFIED_WEBHOOKS is enabled (development only)")
		if err := json.Unmarshal(payload, &event); err != nil {
			logrus.WithError(err).Warn("failed to parse stripe webhook event")
			apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
			return
		}
		// Manually extract Data.Raw from the payload for checkout.session.completed events
		// since Stripe's json.Unmarshal doesn't automatically populate event.Data.Raw
		var rawPayload map[string]interface{}
		if err := json.Unmarshal(payload, &rawPayload); err == nil {
			if dataObj, ok := rawPayload["data"].(map[string]interface{}); ok {
				if obj, ok := dataObj["object"].(map[string]interface{}); ok {
					if event.Data == nil {
						event.Data = &stripe.EventData{}
					}
					event.Data.Raw, _ = json.Marshal(obj)
				}
			}
		}
	} else {
		// Verify webhook signature with configured secret
		signature := r.Header.Get("Stripe-Signature")
		event, err = webhook.ConstructEvent(payload, signature, h.webhookSecret)
		if err != nil {
			logrus.WithError(err).Warn("invalid stripe webhook signature")
			monitoring.RecordBillingWebhookSignatureFailure("invalid_signature")
			apierror.WriteError(w, apierror.NewUnauthorized("Invalid signature"))
			return
		}
	}

	// Store raw webhook payload for replay capability (30-day retention)
	if h.operationalRepo != nil {
		_, storeErr := h.operationalRepo.StoreWebhookPayload(r.Context(), event.ID, string(event.Type), payload, r.Header.Get("Stripe-Signature"))
		if storeErr != nil {
			logrus.WithError(storeErr).WithField("event_id", event.ID).Warn("failed to store webhook payload for replay")
			// Don't fail the webhook - continue processing
		}
	}

	// IDEMPOTENCY CHECK: Check if this event has already been processed
	// This check happens BEFORE any business logic to prevent duplicate processing
	existingEvent, _ := h.userRepo.GetStripeSyncEventByEventID(r.Context(), event.ID)
	if existingEvent != nil && (existingEvent.Status == storage.StripeSyncStatusProcessed || existingEvent.Status == storage.StripeSyncStatusIgnored) {
		logrus.WithField("event_id", event.ID).Info("stripe webhook: duplicate event already processed, acknowledging")
		w.WriteHeader(http.StatusOK)
		return
	}

	h.handleEvent(w, r, &event)
}

func (h *StripeWebhookHandler) handleEvent(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	start := time.Now()

	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(w, r, event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(w, r, event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(w, r, event)
	case "customer.subscription.created":
		h.handleSubscriptionCreated(w, r, event)
	case "invoice.created":
		h.handleInvoiceCreated(w, r, event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(w, r, event)
	case "invoice.payment_succeeded":
		h.handleInvoicePaymentSucceeded(w, r, event)
	case "payment_intent.payment_failed":
		h.handlePaymentIntentFailed(w, r, event)
	case "charge.dispute.created":
		h.handleChargeDisputeCreated(w, r, event)
	case "charge.dispute.updated":
		h.handleChargeDisputeUpdated(w, r, event)
	case "charge.dispute.closed":
		h.handleChargeDisputeClosed(w, r, event)
	case "charge.dispute.funds_withdrawn":
		h.handleChargeDisputeFundsWithdrawn(w, r, event)
	case "charge.refunded":
		h.handleChargeRefunded(w, r, event)
	// Two-way sync: Payment method changes from Stripe dashboard
	case "payment_method.updated":
		h.handlePaymentMethodUpdated(w, r, event)
	case "payment_method.detached":
		h.handlePaymentMethodDetached(w, r, event)
	// Two-way sync: Customer updates from Stripe dashboard
	case "customer.updated":
		h.handleCustomerUpdated(w, r, event)
	// Payout events: Stripe Connect payout lifecycle
	case "payout.paid":
		h.handlePayoutPaid(w, r, event)
	case "payout.failed":
		h.handlePayoutFailed(w, r, event)
	case "transfer.reversed":
		h.handleTransferReversed(w, r, event)
	case "account.updated":
		h.handleConnectAccountUpdated(w, r, event)
	default:
		// Acknowledge but ignore other events
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		monitoring.RecordBillingWebhookReceived(string(event.Type), "ignored")
		duration := time.Since(start)
		monitoring.RecordBillingWebhookProcessingDuration(string(event.Type), duration)
		// Check if latency exceeded 5 second threshold
		if duration > 5*time.Second {
			monitoring.RecordBillingWebhookLatencyExceeded(string(event.Type))
			monitoring.RecordBillingAlertTriggered("webhook_latency", "warning")
		}
	}
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var session stripe.CheckoutSession

	logrus.WithField("event_type", event.Type).WithField("event_id", event.ID).Info("Processing checkout.session.completed")

	if event.Data == nil || event.Data.Raw == nil {
		logrus.Error("checkout.session.completed: event.Data or event.Data.Raw is nil")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid event data"))
		return
	}

	sessionData, err := json.Marshal(event.Data.Raw)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal checkout session data")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid event data"))
		return
	}
	if err := json.Unmarshal(sessionData, &session); err != nil {
		logrus.WithError(err).Error("failed to unmarshal checkout session")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid session data"))
		return
	}

	purpose := session.Metadata["purpose"]
	switch purpose {
	case "agent_execution_credits":
		h.handleAgentExecutionCreditsCheckout(w, r, &session)
	case "registry_wallet_credit":
		h.handleRegistryWalletCreditCheckout(w, r, &session)
	case "state_fabric_addon":
		h.handleStateFabricAddonCheckout(w, r, &session)
	case "bundle_subscription":
		h.handleBundleSubscriptionCheckout(w, r, &session)
	case "function_verification":
		h.handleFunctionVerificationCheckout(w, r, &session)
	case "username_change":
		h.handleUsernameChangeCheckout(w, r, &session)
	case "cert_exam":
		h.handleCertExamCheckout(w, r, &session)
	default:
		logrus.WithField("purpose", purpose).Debug("checkout.session.completed: unknown purpose, ignoring")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
	}
}

func (h *StripeWebhookHandler) handleStateFabricAddonCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	if h.sfAddons == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}
	tenantIDStr := session.Metadata["tenant_id"]
	addonID := session.Metadata["addon_id"]
	if tenantIDStr == "" || addonID == "" {
		logrus.Warn("state fabric addon checkout missing metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant_id"))
		return
	}
	var subscriptionID *string
	var subscriptionItemID *string
	if session.Subscription != nil && session.Subscription.ID != "" {
		subscriptionID = &session.Subscription.ID
		sub, subErr := stripeSub.Get(session.Subscription.ID, nil)
		if subErr != nil {
			logrus.WithError(subErr).WithField("subscription_id", session.Subscription.ID).Warn("state fabric addon: fetch subscription")
		} else if sub != nil && sub.Items != nil && len(sub.Items.Data) > 0 {
			id := sub.Items.Data[0].ID
			subscriptionItemID = &id
		}
	}

	if err := h.sfAddons.UpsertEntitlement(r.Context(), tenantID, addonID, "active", subscriptionID, subscriptionItemID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id": tenantID, "addon_id": addonID,
		}).Error("state fabric addon: upsert entitlement")
		apierror.WriteError(w, apierror.NewInternal("Failed to persist entitlement"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleFunctionVerificationCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	// Extract metadata (only non-PII fields for logging)
	paymentIDStr := session.Metadata["payment_id"]
	functionIDStr := session.Metadata["function_id"]
	// tenantID is intentionally not extracted for logging to avoid PII exposure

	if paymentIDStr == "" || functionIDStr == "" {
		logrus.WithFields(logrus.Fields{
			"session_id":       session.ID,
			"metadata_purpose": session.Metadata["purpose"],
		}).Warn("function verification checkout missing required metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}

	// Lookup payment record by checkout session ID
	payment, err := h.userRepo.GetFunctionVerificationPaymentByCheckoutSessionID(r.Context(), session.ID)
	if err != nil {
		logrus.WithError(err).WithField("session_id", session.ID).Error("failed to find verification payment record")
		apierror.WriteError(w, apierror.NewNotFound("Payment record not found"))
		return
	}

	// Verify payment is in expected state
	if payment.Status != "pending" && payment.Status != "pending_checkout" {
		logrus.WithField("payment_id", payment.ID).WithField("status", payment.Status).Info("verification payment already processed or in unexpected state")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "already processed"})
		return
	}

	// Extract PaymentIntent ID from the session
	var stripePIID *string
	if session.PaymentIntent != nil && session.PaymentIntent.ID != "" {
		stripePIID = &session.PaymentIntent.ID
	}

	// Update payment record to paid status
	if err := h.userRepo.UpdateFunctionVerificationPaymentStatus(r.Context(), payment.ID, "paid", stripePIID, nil); err != nil {
		logrus.WithError(err).WithField("payment_id", payment.ID).Error("failed to update verification payment status")
		apierror.WriteError(w, apierror.NewInternal("Failed to update payment status"))
		return
	}

	// Trigger verification job/workflow
	var jobID *uuid.UUID
	if h.registryRepo != nil {
		// Get the latest version for this function
		latestVersion, err := h.registryRepo.GetLatestFunctionVersion(payment.FunctionID)
		if err != nil {
			logrus.WithError(err).WithField("function_id", payment.FunctionID).Warn("failed to get latest function version for verification")
			// Don't fail the webhook - payment is recorded, verification can be retried
		} else if latestVersion != nil {
			// Parse verification level
			level, parseErr := verification.ParseVerificationLevel(payment.VerificationLevel)
			if parseErr != nil {
				logrus.WithError(parseErr).WithField("level", payment.VerificationLevel).Warn("invalid verification level in payment, defaulting to basic")
				level = verification.Level1Basic
			}

			// Create verification job
			job := &storageregistry.VerificationJob{
				ID:                uuid.New(),
				FunctionID:        payment.FunctionID,
				FunctionVersionID: latestVersion.ID,
				Level:             level.String(),
				Status:            "pending",
				Priority:          "normal",
				RequestedAt:       time.Now(),
				ResultStatus:      "pending",
				IsAutoVerify:      false,
			}

			if err := h.registryRepo.CreateVerificationJob(job); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"function_id": payment.FunctionID,
					"level":       level.String(),
				}).Warn("failed to create verification job")
				// Don't fail the webhook - payment is recorded, job can be created manually
			} else {
				jobID = &job.ID
				// Update payment with verification job ID
				if err := h.userRepo.UpdateFunctionVerificationPaymentJobID(r.Context(), payment.ID, job.ID); err != nil {
					logrus.WithError(err).WithField("payment_id", payment.ID).Warn("failed to update payment with job ID")
				}
				logrus.WithFields(logrus.Fields{
					"job_id":      job.ID,
					"function_id": payment.FunctionID,
					"version_id":  latestVersion.ID,
					"level":       level.String(),
				}).Info("verification job created successfully")
			}
		}
	} else {
		logrus.Warn("registry repository not configured, skipping verification job creation")
	}

	logrus.WithFields(logrus.Fields{
		"payment_id":       payment.ID,
		"level":            payment.VerificationLevel,
		"amount_cents":     payment.AmountCents,
		"payment_intent":   stripePIID,
		"verification_job": jobID,
	}).Info("Function verification payment completed, verification can now proceed")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleBundleSubscriptionCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	bundleSlug := session.Metadata["bundle_slug"]
	tenantIDStr := session.Metadata["tenant_id"]
	founderModeID := session.Metadata["founder_mode_id"]

	if tenantIDStr == "" || bundleSlug == "" {
		logrus.Warn("bundle subscription checkout missing required metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid tenant_id in bundle metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant_id"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"bundle_slug":      bundleSlug,
		"session_id":       session.ID,
		"has_founder_mode": founderModeID != "",
	}).Info("Processing bundle subscription checkout")

	// Get the bundle by slug
	bundle, err := h.userRepo.GetPricingBundleBySlug(r.Context(), bundleSlug)
	if err != nil {
		logrus.WithError(err).WithField("bundle_slug", bundleSlug).Error("failed to get bundle")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve bundle"))
		return
	}
	if bundle == nil {
		logrus.WithField("bundle_slug", bundleSlug).Warn("bundle not found for subscription")
		apierror.WriteError(w, apierror.NewNotFound("Bundle not found"))
		return
	}

	// Get or create bundle subscription
	existingSub, err := h.userRepo.GetBundleSubscriptionByTenant(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("failed to check existing bundle subscription")
	}

	// If already has active subscription, just acknowledge
	if existingSub != nil && existingSub.Status == "active" {
		logrus.WithFields(logrus.Fields{
			"sub_id":     existingSub.ID,
			"session_id": session.ID,
		}).Info("bundle subscription already active, skipping")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "already_active"})
		return
	}

	// Get subscription from Stripe if available
	var subscriptionID string
	var periodStart, periodEnd time.Time

	if session.Subscription != nil && session.Subscription.ID != "" {
		subscriptionID = session.Subscription.ID

		// Get subscription details for billing period
		sub, err := stripeSub.Get(session.Subscription.ID, nil)
		if err == nil && sub != nil && sub.Items != nil && len(sub.Items.Data) > 0 {
			item := sub.Items.Data[0]
			periodStart = time.Unix(item.CurrentPeriodStart, 0)
			periodEnd = time.Unix(item.CurrentPeriodEnd, 0)
		}
	}

	// Handle conversion from founder mode if applicable
	if founderModeID != "" {
		fmid, parseErr := uuid.Parse(founderModeID)
		if parseErr == nil {
			// Update founder mode status to converted
			if err := h.userRepo.UpdateFounderModeStatus(r.Context(), fmid, "converted"); err != nil {
				logrus.WithError(err).WithField("founder_mode_id", fmid).Warn("failed to update founder mode status")
			}
		}
	}

	now := time.Now().UTC()

	// Create or update bundle subscription
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

	// Preserve founder mode ID if converting
	if founderModeID != "" {
		fmid, _ := uuid.Parse(founderModeID)
		sub.FounderModeID = &fmid
		sub.ConvertedFromFounderMode = true
	}

	// If existing deferred subscription exists, update it
	if existingSub != nil && existingSub.Status == "deferred" {
		sub.ID = existingSub.ID
		sub.DefaultAppID = existingSub.DefaultAppID // Preserve the app created during founder mode provisioning
		if err := h.userRepo.UpdateBundleSubscription(r.Context(), sub); err != nil {
			logrus.WithError(err).Error("failed to update bundle subscription from deferred")
			apierror.WriteError(w, apierror.NewInternal("Failed to update subscription"))
			return
		}
	} else {
		// Create new subscription
		if err := h.userRepo.CreateBundleSubscription(r.Context(), sub); err != nil {
			logrus.WithError(err).Error("failed to create bundle subscription")
			apierror.WriteError(w, apierror.NewInternal("Failed to create subscription"))
			return
		}
	}

	logrus.WithFields(logrus.Fields{
		"bundle_slug":     bundleSlug,
		"subscription_id": sub.ID,
		"session_id":      session.ID,
	}).Info("Bundle subscription created/updated successfully")

	// Trigger provisioning of bundle resources (app, backend, functions)
	// This is called for both new purchases and conversions from founder mode.
	// When isolated provisioning is available, skip shared-DB backend/function creation
	// — the BundleProvisioner handles them in the tenant's dedicated database.
	var provisionOpts []func(*billing.ProvisionBundleOpts)
	if h.provisionBundleFn != nil {
		provisionOpts = append(provisionOpts, billing.WithIsolatedProvisioning())
	}
	app, appErr := billing.ProvisionBundleAppAndBackend(h.userRepo, tenantID, bundleSlug, provisionOpts...)
	if appErr != nil {
		logrus.WithError(appErr).WithField("tenant_id", tenantID).Warn("Failed to provision bundle app and backend")
	} else if app != nil {
		// Update subscription with the default app ID
		sub.DefaultAppID = &app.ID
		if updateErr := h.userRepo.UpdateBundleSubscription(r.Context(), sub); updateErr != nil {
			logrus.WithError(updateErr).WithField("tenant_id", tenantID).Warn("Failed to update subscription with app ID")
		}
	}

	// Trigger isolated bundle provisioning (creates dedicated DB + all tenant resources)
	// If isolated provisioning fails, fall back to shared-DB provisioning with degraded_mode flag.
	if h.provisionBundleFn != nil {
		go func() {
			status, count, err := h.provisionBundleFn(context.Background(), tenantID, bundleSlug)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"bundle":    bundleSlug,
				}).Error("Isolated bundle provisioning failed — falling back to shared-DB mode")

				// Graceful degradation: create backend/functions in shared DB
				if _, fallbackErr := billing.ProvisionBundleAppAndBackend(h.userRepo, tenantID, bundleSlug); fallbackErr != nil {
					logrus.WithError(fallbackErr).WithField("tenant_id", tenantID).Error("Shared-DB fallback provisioning also failed")
				}

				// Mark tenant as degraded
				if degradeErr := h.userRepo.SetTenantDegradedMode(r.Context(), tenantID, true, fmt.Sprintf("isolated provisioning failed: %v", err)); degradeErr != nil {
					logrus.WithError(degradeErr).WithField("tenant_id", tenantID).Warn("Failed to set tenant degraded mode")
				}
			} else {
				logrus.WithFields(logrus.Fields{
					"tenant_id":  tenantID,
					"status":     status,
					"components": count,
				}).Info("Isolated bundle provisioning complete")

				// Clear degraded mode if it was previously set
				if degradeErr := h.userRepo.SetTenantDegradedMode(r.Context(), tenantID, false, ""); degradeErr != nil {
					logrus.WithError(degradeErr).WithField("tenant_id", tenantID).Warn("Failed to clear tenant degraded mode")
				}
			}
		}()
	}

	// Also provision general bundle resources (auth, analytics, vector collections)
	go func() {
		if err := billing.ProvisionBundleResources(h.userRepo, tenantID, bundleSlug); err != nil {
			logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to provision bundle resources")
		}
	}()

	// Send bundle welcome email
	if h.emailSvc != nil {
		dashboardURL := os.Getenv("DASHBOARD_URL")
		if dashboardURL == "" {
			dashboardURL = "https://app.functionfly.com"
		}
		var userEmail string
		if session.CustomerEmail != "" {
			userEmail = session.CustomerEmail
		} else {
			users, uErr := h.userRepo.ListActiveUsersByTenant(r.Context(), tenantID)
			if uErr == nil && len(users) > 0 {
				userEmail = users[0].Email
			}
		}
		if userEmail != "" {
			go func() {
				if err := h.emailSvc.SendBundleWelcomeEmail(userEmail, bundle.Name, dashboardURL); err != nil {
					logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to send bundle welcome email")
				} else {
					logrus.WithField("tenant_id", tenantID).Info("Bundle welcome email sent successfully")
				}
			}()
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleUsernameChangeCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	pendingChangeIDStr := session.Metadata["pending_change_id"]
	userIDStr := session.Metadata["user_id"]
	tenantIDStr := session.Metadata["tenant_id"]
	newUsername := session.Metadata["new_username"]
	feeCentsStr := session.Metadata["fee_cents"]

	if pendingChangeIDStr == "" || userIDStr == "" || newUsername == "" {
		logrus.Warn("username change checkout missing required metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}

	pendingChangeID, err := uuid.Parse(pendingChangeIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid pending_change_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid pending_change_id"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid user_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user_id"))
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid tenant_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant_id"))
		return
	}

	// Parse fee amount
	feeCents := 500 // Default $5.00
	if feeCentsStr != "" {
		if parsed, err := strconv.Atoi(feeCentsStr); err == nil && parsed > 0 {
			feeCents = parsed
		}
	}

	// Get pending change from database
	pending, err := h.userRepo.GetPendingUsernameChangeByID(r.Context(), pendingChangeID)
	if err != nil {
		logrus.WithError(err).WithField("pending_change_id", pendingChangeIDStr).Error("failed to get pending username change")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve pending change"))
		return
	}
	if pending == nil {
		logrus.WithField("pending_change_id", pendingChangeIDStr).Warn("pending username change not found")
		apierror.WriteError(w, apierror.NewNotFound("Pending change not found"))
		return
	}

	// Verify pending change is still valid
	if !pending.CanComplete() {
		logrus.WithField("pending_change_id", pendingChangeIDStr).Warn("pending username change is expired or not pending")
		apierror.WriteError(w, apierror.NewBadRequest("Pending change has expired or is invalid"))
		return
	}

	// Get the user
	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", userIDStr).Error("user not found for username change")
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	// Verify tenant matches
	if user.TenantID != tenantID {
		logrus.WithFields(logrus.Fields{
			"session_id": session.ID,
		}).Warn("tenant mismatch for username change")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid checkout metadata"))
		return
	}

	oldUsername := ""
	if user.Username != nil {
		oldUsername = *user.Username
	}

	// Check if the new username is still available (it might have been taken in the meantime)
	existingUser, err := h.userRepo.GetUserByUsername(r.Context(), newUsername)
	if err != nil {
		logrus.WithError(err).WithField("username", newUsername).Error("failed to check username availability")
		apierror.WriteError(w, apierror.NewInternal("Failed to verify username availability"))
		return
	}
	if existingUser != nil && existingUser.ID != userID {
		logrus.WithField("username", newUsername).Warn("username no longer available")
		_ = h.userRepo.UpdatePendingUsernameChangeStatus(r.Context(), pending.ID, "failed")
		apierror.WriteError(w, apierror.NewConflict("Username is no longer available"))
		return
	}

	// Mark pending change as completed
	if err := h.userRepo.UpdatePendingUsernameChangeStatus(r.Context(), pending.ID, "completed"); err != nil {
		logrus.WithError(err).WithField("pending_change_id", pending.ID).Error("failed to update pending change status")
	}

	// Record the username change in history
	history := &storage.UsernameChangeHistory{
		ID:              uuid.New(),
		UserID:          userID,
		OldUsername:     oldUsername,
		NewUsername:     newUsername,
		ChangedAt:       time.Now(),
		ChangedBy:       userID,
		WasEarlyChange:  true,
		FeePaidCents:    feeCents,
		FeeCurrency:     "USD",
		StripePaymentID: &session.PaymentIntent.ID,
		IPAddress:       pending.IPAddress,
		UserAgent:       pending.UserAgent,
	}

	if err := h.userRepo.CreateUsernameChangeHistory(r.Context(), history); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("failed to create username change history")
		// Continue anyway, don't fail the webhook
	}

	// Update the user's username
	updates := map[string]interface{}{
		"username": newUsername,
	}
	_, err = h.userRepo.UpdateUser(r.Context(), userID, updates)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":      userID,
			"new_username": newUsername,
		}).Error("failed to update user username")
		apierror.WriteError(w, apierror.NewInternal("Failed to update username"))
		return
	}

	// Create invoice record for the payment
	amountUSD := float64(feeCents) / 100.0
	h.persistCheckoutInvoice(r.Context(), tenantID, session, amountUSD)

	// Send notification
	if h.notificationSvc != nil {
		if err := h.notificationSvc.SendUsernameChanged(r.Context(), userID, oldUsername, newUsername); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("failed to send username change notification")
		}
	}

	logrus.WithFields(logrus.Fields{
		"fee_cents":  feeCents,
		"session_id": session.ID,
	}).Info("Username changed successfully via paid checkout")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleSubscriptionUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid subscription payload"))
		return
	}

	ctx := r.Context()

	// Log the event for audit trail
	h.logStripeEvent(ctx, event, &sub)

	// Handle State Fabric Addon subscriptions
	if sub.Metadata["purpose"] == "state_fabric_addon" && h.sfAddons != nil {
		h.handleStateFabricSubscriptionUpdated(w, r, &sub)
		return
	}

	// Handle main (bundle) subscriptions
	h.handleMainSubscriptionUpdated(w, r, &sub)
}

// handleSubscriptionCreated handles new bundle subscriptions created via Stripe checkout
// (including founder mode conversions that transition to paid)
func (h *StripeWebhookHandler) handleSubscriptionCreated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid subscription payload"))
		return
	}

	ctx := r.Context()

	// Log the event for audit trail
	h.logStripeEvent(ctx, event, &sub)

	// Check if this is a bundle subscription
	bundleSlug := sub.Metadata["bundle_slug"]
	if bundleSlug == "" {
		// Not a bundle subscription — ignore
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "not_bundle_subscription"})
		return
	}

	tenantIDStr := sub.Metadata["tenant_id"]
	tenantID, _ := uuid.Parse(tenantIDStr)

	// Check for founder mode conversion
	founderModeIDStr := sub.Metadata["founder_mode_id"]
	if founderModeIDStr != "" {
		// Founder mode conversion — process via the dedicated handler
		if err := h.handleBundleSubscriptionCreated(ctx, &sub, tenantID); err != nil {
			logrus.WithError(err).WithField("founder_mode_id", founderModeIDStr).Error("stripe webhook: founder mode conversion failed")
			apierror.WriteError(w, apierror.NewInternal("Failed to process founder mode conversion"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "founder_mode_converted"})
		return
	}

	// Regular bundle subscription (not from founder mode)
	// Update the bundle subscription with the Stripe subscription ID and mark as active
	bundleSub, err := h.userRepo.GetBundleSubscriptionByTenant(ctx, tenantID)
	if err != nil || bundleSub == nil {
		logrus.WithField("tenant_id", tenantIDStr).Warn("stripe webhook: no bundle subscription found for regular creation")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "no_subscription_found"})
		return
	}

	bundleSub.StripeSubscriptionID = sub.ID
	bundleSub.Status = "active"
	bundleSub.CurrentPeriodStart = time.Now()
	bundleSub.CurrentPeriodEnd = time.Now().AddDate(0, 1, 0)

	if err := h.userRepo.UpdateBundleSubscription(ctx, bundleSub); err != nil {
		logrus.WithError(err).Error("stripe webhook: failed to update bundle subscription to active")
		apierror.WriteError(w, apierror.NewInternal("Failed to update subscription"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"bundle_subscription_id": bundleSub.ID,
		"stripe_subscription_id": sub.ID,
	}).Info("stripe webhook: bundle subscription marked active")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "active"})
}

// handleStateFabricSubscriptionUpdated handles updates to State Fabric addon subscriptions
func (h *StripeWebhookHandler) handleStateFabricSubscriptionUpdated(w http.ResponseWriter, r *http.Request, sub *stripe.Subscription) {
	ctx := r.Context()
	addonID := sub.Metadata["addon_id"]
	tenantIDStr := sub.Metadata["tenant_id"]
	tenantID, parseErr := uuid.Parse(tenantIDStr)
	if addonID == "" || parseErr != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid subscription metadata"))
		return
	}
	status := "inactive"
	if sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing {
		status = "active"
	}
	var itemID *string
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		id := sub.Items.Data[0].ID
		itemID = &id
	}
	subID := sub.ID
	if err := h.sfAddons.UpsertEntitlement(ctx, tenantID, addonID, status, &subID, itemID); err != nil {
		logrus.WithError(err).WithField("subscription_id", sub.ID).Error("state fabric addon: subscription updated entitlement")
		apierror.WriteError(w, apierror.NewInternal("Failed to update entitlement"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleMainSubscriptionUpdated handles updates to main bundle subscriptions from Stripe dashboard
func (h *StripeWebhookHandler) handleMainSubscriptionUpdated(w http.ResponseWriter, r *http.Request, stripeSub *stripe.Subscription) {
	ctx := r.Context()

	// Find the subscription by Stripe ID
	subscription, err := h.userRepo.GetSubscriptionByStripeID(ctx, stripeSub.ID)
	if err != nil {
		logrus.WithError(err).WithField("stripe_subscription_id", stripeSub.ID).Error("failed to find subscription by stripe id")
		apierror.WriteError(w, apierror.NewInternal("Failed to find subscription"))
		return
	}

	if subscription == nil {
		// Check if this is a bundle subscription instead
		bundleSub, err := h.userRepo.GetBundleSubscriptionByTenant(ctx, uuid.Nil) // Need to find by stripe ID
		if err != nil || bundleSub == nil || bundleSub.StripeSubscriptionID != stripeSub.ID {
			// Unknown subscription - log but don't fail
			logrus.WithField("stripe_subscription_id", stripeSub.ID).Warn("subscription not found for stripe id")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "subscription_not_found"})
			return
		}

		// Update bundle subscription
		h.updateBundleSubscriptionFromStripe(ctx, w, bundleSub, stripeSub)
		return
	}

	// Update main subscription from Stripe data
	h.updateMainSubscriptionFromStripe(ctx, w, subscription, stripeSub)
}

// updateMainSubscriptionFromStripe syncs a main subscription with data from Stripe
func (h *StripeWebhookHandler) updateMainSubscriptionFromStripe(ctx context.Context, w http.ResponseWriter, subscription *storage.Subscription, stripeSub *stripe.Subscription) {
	// Build updates from Stripe data
	updates := map[string]interface{}{
		"stripe_subscription_id": stripeSub.ID,
	}

	// Map status
	status := mapStripeStatusToInternal(string(stripeSub.Status))
	if status != subscription.Status {
		updates["status"] = status
		logrus.WithFields(logrus.Fields{
			"subscription_id": subscription.ID,
			"old_status":      subscription.Status,
			"new_status":      status,
			"stripe_status":   stripeSub.Status,
		}).Info("subscription status changed via stripe webhook")
	}

	// Handle trial
	if stripeSub.TrialEnd > 0 {
		t := time.Unix(stripeSub.TrialEnd, 0)
		updates["trial_end"] = &t
	}

	// Handle cancellation flags
	if stripeSub.CanceledAt > 0 {
		t := time.Unix(stripeSub.CanceledAt, 0)
		updates["canceled_at"] = &t
	}
	if stripeSub.CancelAtPeriodEnd {
		updates["cancel_at_period_end"] = true
	}

	// Apply updates
	updated, err := h.userRepo.UpdateSubscription(ctx, subscription.ID, updates)
	if err != nil {
		logrus.WithError(err).WithField("subscription_id", subscription.ID).Error("failed to update subscription from stripe")
		apierror.WriteError(w, apierror.NewInternal("Failed to update subscription"))
		return
	}

	// Send notification for important status changes
	if h.notificationSvc != nil && status != subscription.Status {
		h.sendSubscriptionStatusChangeNotification(ctx, subscription, status)
	}

	logrus.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"stripe_id":       stripeSub.ID,
		"status":          status,
	}).Info("subscription synced from stripe")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":          "success",
		"subscription_id": updated.ID.String(),
		"synced_status":   status,
	})
}

// updateBundleSubscriptionFromStripe syncs a bundle subscription with data from Stripe
// Also handles plan changes (upgrades/downgrades) when the Stripe price changes
func (h *StripeWebhookHandler) updateBundleSubscriptionFromStripe(ctx context.Context, w http.ResponseWriter, bundleSub *storage.BundleSubscription, stripeSub *stripe.Subscription) {
	// Map status
	status := "active"
	if stripeSub.Status == stripe.SubscriptionStatusCanceled || stripeSub.Status == stripe.SubscriptionStatusUnpaid {
		status = "cancelled"
	} else if stripeSub.Status == stripe.SubscriptionStatusPastDue {
		status = "past_due"
	}

	// Track if this is a plan change
	planChanged := false
	var oldBundleID, newBundleID uuid.UUID
	var prorationCredit float64

	// Check for plan change by examining subscription items
	if stripeSub.Items != nil && len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]
		if item.Price != nil && item.Price.ID != "" {
			newStripePriceID := item.Price.ID

			// Look up the bundle for this Stripe price
			newBundle, err := h.userRepo.GetPricingBundleByStripePriceID(ctx, newStripePriceID)
			if err != nil {
				logrus.WithError(err).WithField("stripe_price_id", newStripePriceID).
					Warn("failed to lookup bundle by stripe price id during subscription sync")
			} else if newBundle != nil && newBundle.ID != bundleSub.BundleID {
				// Plan change detected!
				planChanged = true
				oldBundleID = bundleSub.BundleID
				newBundleID = newBundle.ID

				logrus.WithFields(logrus.Fields{
					"bundle_subscription_id": bundleSub.ID,
					"old_bundle_id":          oldBundleID,
					"new_bundle_id":          newBundleID,
					"stripe_subscription_id": stripeSub.ID,
				}).Info("plan change detected from stripe subscription update")

				// Calculate proration credit if applicable
				prorationCredit = h.calculateProrationCredit(bundleSub, stripeSub)

				// Update the bundle ID
				bundleSub.BundleID = newBundle.ID
			}
		}
	}

	// Apply updates by creating an updated bundle subscription object
	bundleSub.Status = status
	bundleSub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd
	if stripeSub.CanceledAt > 0 {
		t := time.Unix(stripeSub.CanceledAt, 0)
		bundleSub.CanceledAt = &t
	}

	// Update billing period if changed (for billing alignment during plan changes)
	// Get period from subscription items (this is where Stripe stores the current period)
	if stripeSub.Items != nil && len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]
		if item.CurrentPeriodStart > 0 && item.CurrentPeriodEnd > 0 {
			newPeriodStart := time.Unix(item.CurrentPeriodStart, 0)
			newPeriodEnd := time.Unix(item.CurrentPeriodEnd, 0)

			// Only update if the period has actually changed to avoid unnecessary updates
			if !newPeriodStart.Equal(bundleSub.CurrentPeriodStart) || !newPeriodEnd.Equal(bundleSub.CurrentPeriodEnd) {
				bundleSub.CurrentPeriodStart = newPeriodStart
				bundleSub.CurrentPeriodEnd = newPeriodEnd
				logrus.WithFields(logrus.Fields{
					"bundle_subscription_id": bundleSub.ID,
					"new_period_start":       newPeriodStart,
					"new_period_end":         newPeriodEnd,
					"plan_changed":           planChanged,
				}).Info("billing period updated from stripe")
			}
		}
	}

	bundleSub.UpdatedAt = time.Now()

	if err := h.userRepo.UpdateBundleSubscription(ctx, bundleSub); err != nil {
		logrus.WithError(err).WithField("bundle_subscription_id", bundleSub.ID).Error("failed to update bundle subscription from stripe")
		apierror.WriteError(w, apierror.NewInternal("Failed to update bundle subscription"))
		return
	}

	// Process plan change side effects after successful update
	if planChanged {
		h.handlePlanChangeSideEffects(ctx, bundleSub, oldBundleID, newBundleID, prorationCredit)
	}

	logrus.WithFields(logrus.Fields{
		"bundle_subscription_id": bundleSub.ID,
		"stripe_id":              stripeSub.ID,
		"status":                 status,
		"plan_changed":           planChanged,
	}).Info("bundle subscription synced from stripe")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":                 "success",
		"bundle_subscription_id": bundleSub.ID.String(),
		"synced_status":          status,
		"plan_changed":           fmt.Sprintf("%t", planChanged),
	})
}

// calculateProrationCredit calculates credit to apply during plan downgrades
// When downgrading, customers may have prepaid for the current period at a higher rate
func (h *StripeWebhookHandler) calculateProrationCredit(bundleSub *storage.BundleSubscription, stripeSub *stripe.Subscription) float64 {
	if stripeSub.LatestInvoice == nil || stripeSub.LatestInvoice.ID == "" {
		return 0.0
	}

	invoiceID := stripeSub.LatestInvoice.ID

	fullInvoice, err := invoice.Get(invoiceID, &stripe.InvoiceParams{})
	if err != nil {
		logrus.WithError(err).WithField("invoice_id", invoiceID).
			Warn("failed to fetch full invoice for proration calculation, using amount_due fallback")
		amount := stripeSub.LatestInvoice.AmountDue
		if amount < 0 {
			return float64(-amount) / 100.0
		}
		return 0.0
	}

	var prorationCredit float64

	for _, line := range fullInvoice.Lines.Data {
		isProration := !line.Discountable && strings.Contains(strings.ToLower(line.Description), "proration")
		if isProration || (line.Parent != nil && line.Parent.Type == "subscription" && !line.Discountable) {
			amount := line.Amount
			if amount < 0 {
				prorationCredit += float64(-amount) / 100.0
			}
		}
	}

	if prorationCredit > 0 {
		logrus.WithFields(logrus.Fields{
			"bundle_subscription_id": bundleSub.ID,
			"credit_usd":             prorationCredit,
			"invoice_id":             invoiceID,
			"line_count":             len(fullInvoice.Lines.Data),
		}).Info("proration credit calculated from stripe invoice lines")
	}

	return prorationCredit
}

// handlePlanChangeSideEffects processes membership events, notifications, and credit transfers
// when a subscription plan changes (upgrade or downgrade)
func (h *StripeWebhookHandler) handlePlanChangeSideEffects(ctx context.Context, bundleSub *storage.BundleSubscription, oldBundleID, newBundleID uuid.UUID, prorationCredit float64) {
	// Get old and new bundle details for comparison
	oldBundle, err := h.userRepo.GetPricingBundleByID(ctx, oldBundleID)
	if err != nil {
		logrus.WithError(err).WithField("bundle_id", oldBundleID).Warn("failed to get old bundle for plan change processing")
	}
	newBundle, err := h.userRepo.GetPricingBundleByID(ctx, newBundleID)
	if err != nil {
		logrus.WithError(err).WithField("bundle_id", newBundleID).Warn("failed to get new bundle for plan change processing")
	}

	if oldBundle == nil || newBundle == nil {
		logrus.Warn("cannot process plan change side effects - bundle lookup failed")
		return
	}

	// Determine if this is an upgrade or downgrade based on price
	isUpgrade := newBundle.DisplayPriceCents > oldBundle.DisplayPriceCents
	changeType := "upgrade"
	if !isUpgrade {
		changeType = "downgrade"
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":        bundleSub.TenantID,
		"old_bundle":       oldBundle.Slug,
		"new_bundle":       newBundle.Slug,
		"change_type":      changeType,
		"proration_credit": prorationCredit,
	}).Info("processing plan change side effects")

	// Process membership events for upgrades
	if isUpgrade {
		h.processMembershipUpgrade(ctx, bundleSub.TenantID, oldBundle.Slug, newBundle.Slug)
	}

	// Handle proration credit for downgrades
	if !isUpgrade && prorationCredit > 0 {
		h.applyProrationCredit(ctx, bundleSub, prorationCredit)
	}

	// Send plan change notification
	if h.notificationSvc != nil {
		h.sendPlanChangeNotification(ctx, bundleSub, oldBundle, newBundle, isUpgrade, prorationCredit)
	}
}

// applyProrationCredit applies proration credit to the tenant's wallet
// This handles credit transfers during downgrades
func (h *StripeWebhookHandler) applyProrationCredit(ctx context.Context, bundleSub *storage.BundleSubscription, creditUSD float64) {
	// Get active users in the tenant to credit the wallet
	users, err := h.userRepo.ListActiveUsersByTenant(ctx, bundleSub.TenantID)
	if err != nil || len(users) == 0 {
		logrus.WithError(err).WithField("tenant_id", bundleSub.TenantID).
			Warn("failed to find users for proration credit application")
		return
	}

	// Credit the first active user's wallet
	user := users[0]

	if h.platformFees != nil {
		reference := fmt.Sprintf("proration_credit:%s", bundleSub.StripeSubscriptionID)
		if err := h.platformFees.CreditWallet(ctx, user.ID, creditUSD, reference); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":                user.ID,
				"tenant_id":              bundleSub.TenantID,
				"credit_usd":             creditUSD,
				"bundle_subscription_id": bundleSub.ID,
			}).Error("failed to apply proration credit to wallet")
			return
		}

		logrus.WithFields(logrus.Fields{
			"user_id":                user.ID,
			"tenant_id":              bundleSub.TenantID,
			"credit_usd":             creditUSD,
			"bundle_subscription_id": bundleSub.ID,
		}).Info("proration credit applied to wallet")

		// Send credit notification
		newBalance, _ := h.platformFees.GetWalletBalance(ctx, user.ID)
		if err := h.notificationSvc.SendRegistryWalletTopUp(ctx, user.ID, creditUSD, newBalance); err != nil {
			logrus.WithError(err).WithField("user_id", user.ID).Warn("failed to send proration credit notification")
		}
	}
}

// sendPlanChangeNotification sends a notification about the plan change
func (h *StripeWebhookHandler) sendPlanChangeNotification(ctx context.Context, bundleSub *storage.BundleSubscription, oldBundle, newBundle *storage.PricingBundle, isUpgrade bool, prorationCredit float64) {
	var title, body string
	priority := notification.PriorityNormal

	if isUpgrade {
		title = "Plan Upgraded"
		body = fmt.Sprintf("Your subscription has been upgraded from %s to %s. Enjoy your new features!", oldBundle.DisplayName, newBundle.DisplayName)
		priority = notification.PriorityNormal
	} else {
		title = "Plan Downgraded"
		body = fmt.Sprintf("Your subscription has been downgraded from %s to %s.", oldBundle.DisplayName, newBundle.DisplayName)
		if prorationCredit > 0 {
			body += fmt.Sprintf(" A credit of $%.2f has been applied to your wallet for the unused portion of your previous plan.", prorationCredit)
		}
	}

	// Get first active user to notify
	users, err := h.userRepo.ListActiveUsersByTenant(ctx, bundleSub.TenantID)
	if err != nil || len(users) == 0 {
		logrus.WithError(err).WithField("tenant_id", bundleSub.TenantID).Warn("failed to find users for plan change notification")
		return
	}

	_, err = h.notificationSvc.Send(ctx, notification.SendRequest{
		UserID:   users[0].ID,
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    title,
		Body:     body,
		Data: map[string]interface{}{
			"bundle_subscription_id": bundleSub.ID.String(),
			"old_bundle":             oldBundle.Slug,
			"new_bundle":             newBundle.Slug,
			"change_type":            map[bool]string{true: "upgrade", false: "downgrade"}[isUpgrade],
			"proration_credit_usd":   prorationCredit,
			"changed_at":             time.Now().Format(time.RFC3339),
		},
		Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		Priority: priority,
	})
	if err != nil {
		logrus.WithError(err).Warn("failed to send plan change notification")
	}
}

// sendSubscriptionStatusChangeNotification sends notification for subscription status changes
func (h *StripeWebhookHandler) sendSubscriptionStatusChangeNotification(ctx context.Context, sub *storage.Subscription, newStatus string) {
	var title, body string
	priority := notification.PriorityNormal

	switch newStatus {
	case "past_due":
		title = "Payment Past Due"
		body = "Your subscription payment is past due. Please update your payment method to avoid service interruption."
		priority = notification.PriorityHigh
	case "cancelled":
		title = "Subscription Cancelled"
		body = "Your subscription has been cancelled. You will be downgraded to the free plan at the end of your billing period."
	case "unpaid":
		title = "Subscription Unpaid"
		body = "Your subscription is unpaid. Please make a payment to restore service."
		priority = notification.PriorityHigh
	default:
		return
	}

	_, err := h.notificationSvc.Send(ctx, notification.SendRequest{
		UserID:   sub.TenantID,
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    title,
		Body:     body,
		Data: map[string]interface{}{
			"subscription_id": sub.ID.String(),
			"new_status":      newStatus,
			"synced_at":       time.Now().Format(time.RFC3339),
		},
		Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		Priority: priority,
	})
	if err != nil {
		logrus.WithError(err).Warn("failed to send subscription status change notification")
	}
}

// mapStripeStatusToInternal maps Stripe subscription status to internal status
func mapStripeStatusToInternal(stripeStatus string) string {
	switch stripeStatus {
	case "active":
		return "active"
	case "canceled":
		return "cancelled"
	case "incomplete":
		return "incomplete"
	case "incomplete_expired":
		return "expired"
	case "past_due":
		return "past_due"
	case "paused":
		return "paused"
	case "trialing":
		return "trialing"
	case "unpaid":
		return "unpaid"
	default:
		return stripeStatus
	}
}

func (h *StripeWebhookHandler) handleSubscriptionDeleted(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid subscription payload"))
		return
	}

	ctx := r.Context()

	// Log the event for audit trail
	h.logStripeEvent(ctx, event, &sub)

	// Handle State Fabric Addon subscriptions
	if sub.Metadata["purpose"] == "state_fabric_addon" && h.sfAddons != nil {
		if err := h.sfAddons.SetEntitlementStatusBySubscription(ctx, sub.ID, "inactive"); err != nil {
			logrus.WithError(err).WithField("subscription_id", sub.ID).Error("state fabric addon: subscription deleted entitlement")
			apierror.WriteError(w, apierror.NewInternal("Failed to update entitlement"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}

	// Handle main subscriptions
	h.handleMainSubscriptionDeleted(ctx, w, r, &sub)
}

// handleMainSubscriptionDeleted handles deletions of main subscriptions from Stripe dashboard
func (h *StripeWebhookHandler) handleMainSubscriptionDeleted(ctx context.Context, w http.ResponseWriter, r *http.Request, stripeSub *stripe.Subscription) {
	// Try to find main subscription by Stripe ID
	subscription, err := h.userRepo.GetSubscriptionByStripeID(ctx, stripeSub.ID)
	if err != nil {
		logrus.WithError(err).WithField("stripe_subscription_id", stripeSub.ID).Error("failed to find subscription by stripe id")
		apierror.WriteError(w, apierror.NewInternal("Failed to find subscription"))
		return
	}

	if subscription != nil {
		// Update main subscription to cancelled status
		updates := map[string]interface{}{
			"status":      "cancelled",
			"canceled_at": time.Now(),
		}

		_, err := h.userRepo.UpdateSubscription(ctx, subscription.ID, updates)
		if err != nil {
			logrus.WithError(err).WithField("subscription_id", subscription.ID).Error("failed to cancel subscription from stripe webhook")
			apierror.WriteError(w, apierror.NewInternal("Failed to cancel subscription"))
			return
		}

		// Send notification
		if h.notificationSvc != nil {
			h.sendSubscriptionStatusChangeNotification(ctx, subscription, "cancelled")
		}

		logrus.WithFields(logrus.Fields{
			"subscription_id": subscription.ID,
			"stripe_id":       stripeSub.ID,
		}).Info("subscription cancelled via stripe dashboard")

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":          "success",
			"subscription_id": subscription.ID.String(),
			"action":          "cancelled",
		})
		return
	}

	// Try to find bundle subscription
	// First get tenant by Stripe customer ID
	if stripeSub.Customer != nil && stripeSub.Customer.ID != "" {
		tenant, err := h.userRepo.GetTenantByStripeCustomerID(ctx, stripeSub.Customer.ID)
		if err == nil && tenant != nil {
			bundleSub, err := h.userRepo.GetBundleSubscriptionByTenant(ctx, tenant.ID)
			if err == nil && bundleSub != nil && bundleSub.StripeSubscriptionID == stripeSub.ID {
				// Update bundle subscription
				bundleSub.Status = "cancelled"
				t := time.Now()
				bundleSub.CanceledAt = &t
				bundleSub.UpdatedAt = time.Now()

				if err := h.userRepo.UpdateBundleSubscription(ctx, bundleSub); err != nil {
					logrus.WithError(err).WithField("bundle_subscription_id", bundleSub.ID).Error("failed to cancel bundle subscription from stripe webhook")
					apierror.WriteError(w, apierror.NewInternal("Failed to cancel bundle subscription"))
					return
				}

				logrus.WithFields(logrus.Fields{
					"bundle_subscription_id": bundleSub.ID,
					"stripe_id":              stripeSub.ID,
				}).Info("bundle subscription cancelled via stripe dashboard")

				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status":                 "success",
					"bundle_subscription_id": bundleSub.ID.String(),
					"action":                 "cancelled",
				})
				return
			}
		}
	}

	// Unknown subscription - acknowledge but log
	logrus.WithField("stripe_subscription_id", stripeSub.ID).Warn("subscription deleted but not found in system")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "acknowledged",
		"note":   "subscription_not_found",
	})
}

func (h *StripeWebhookHandler) handleAgentExecutionCreditsCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	tenantIDStr := session.Metadata["tenant_id"]
	agentID := session.Metadata["agent_id"]
	amountUsdStr := session.Metadata["amount_usd"]
	initiatingUserIDStr := session.Metadata["initiating_user_id"]

	if tenantIDStr == "" || agentID == "" {
		logrus.Warn("agent checkout session missing required metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid tenant_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant_id"))
		return
	}

	amountUSD, _ := strconv.ParseFloat(amountUsdStr, 64)
	if amountUSD <= 0 {
		amountUSD = float64(session.AmountTotal) / 100
	}

	provider := "stripe"
	providerRef := session.ID

	tx := &storage.AgentFinancialTransaction{
		TenantID:    tenantID,
		AgentID:     agentID,
		Kind:        "credit_purchase",
		AmountUSD:   amountUSD,
		Status:      "completed",
		Provider:    &provider,
		ProviderRef: &providerRef,
	}

	created, err := h.financialTxRepo.CreateIdempotent(r.Context(), tx)
	if err != nil {
		logrus.WithError(err).Error("failed to create transaction record")
		apierror.WriteError(w, apierror.NewInternal("Failed to record transaction"))
		return
	}

	if !created {
		logrus.WithField("session_id", session.ID).Info("duplicate webhook received, transaction already recorded")
		h.persistCheckoutInvoice(r.Context(), tenantID, session, amountUSD)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "already processed"})
		return
	}

	if err := h.billingCtrl.AddCredits(r.Context(), agentID, amountUSD); err != nil {
		logrus.WithError(err).Error("failed to add credits to agent")
		apierror.WriteError(w, apierror.NewInternal("Failed to add credits"))
		return
	}
	controls, controlsErr := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
	if controlsErr != nil {
		logrus.WithError(controlsErr).WithField("agent_id", agentID).Warn("failed to load controls after wallet credit")
	}

	// Track notification status for compensation tracking
	notificationSent := false
	var notificationErr error

	if h.notificationSvc != nil {
		if initiatingUserIDStr == "" {
			logrus.WithField("session_id", session.ID).Info("initiating_user_id missing in checkout metadata; skipping top-up notification")
		} else if initiatingUserID, parseErr := uuid.Parse(initiatingUserIDStr); parseErr != nil {
			logrus.WithError(parseErr).WithFields(logrus.Fields{
				"session_id":         session.ID,
				"initiating_user_id": initiatingUserIDStr,
			}).Warn("invalid initiating_user_id in checkout metadata; skipping top-up notification")
		} else {
			balanceUSD := 0.0
			if controls != nil {
				balanceUSD = controls.CreditBalanceUSD
			}
			if err := h.notificationSvc.SendWalletTopUp(r.Context(), initiatingUserID, agentID, amountUSD, balanceUSD); err != nil {
				notificationErr = err
				logrus.WithError(err).WithFields(logrus.Fields{
					"session_id":            session.ID,
					"agent_id":              agentID,
					"amount_usd":            amountUSD,
					"balance_usd":           balanceUSD,
					"compensation_required": true,
				}).Error("CRITICAL: Credits added but notification failed. Manual reconciliation may be required.")
			} else {
				notificationSent = true
			}
		}
	}

	// Log final state for audit and compensation tracking
	logrus.WithFields(logrus.Fields{
		"agent_id":             agentID,
		"amount_usd":           amountUSD,
		"session_id":           session.ID,
		"notification_sent":    notificationSent,
		"notification_error":   notificationErr,
		"credits_applied":      true,
		"transaction_recorded": true,
	}).Info("credits purchased via checkout")

	h.persistCheckoutInvoice(r.Context(), tenantID, session, amountUSD)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleRegistryWalletCreditCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	if h.platformFees == nil {
		logrus.Error("registry wallet webhook: platform fee repository not configured")
		apierror.WriteError(w, apierror.NewInternal("Wallet service unavailable"))
		return
	}

	tenantIDStr := session.Metadata["tenant_id"]
	userIDStr := session.Metadata["user_id"]
	amountUsdStr := session.Metadata["amount_usd"]

	if tenantIDStr == "" || userIDStr == "" {
		logrus.Warn("registry wallet checkout missing tenant_id or user_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Warn("registry wallet webhook: invalid tenant_id")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant_id"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Warn("registry wallet webhook: invalid user_id")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user_id"))
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", userIDStr).Warn("registry wallet webhook: user not found")
		apierror.WriteError(w, apierror.NewBadRequest("User not found"))
		return
	}
	if user.TenantID != tenantID {
		logrus.WithFields(logrus.Fields{"session_id": session.ID}).Warn("registry wallet webhook: tenant mismatch")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid checkout metadata"))
		return
	}

	amountUSD, _ := strconv.ParseFloat(amountUsdStr, 64)
	if amountUSD <= 0 {
		amountUSD = float64(session.AmountTotal) / 100
	}
	if amountUSD <= 0 {
		logrus.WithField("session_id", session.ID).Warn("registry wallet webhook: non-positive amount")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid amount"))
		return
	}

	already, err := h.platformFees.HasWalletCreditReference(r.Context(), session.ID)
	if err != nil {
		logrus.WithError(err).Error("registry wallet webhook: idempotency check failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to verify payment"))
		return
	}
	if already {
		logrus.WithField("session_id", session.ID).Info("duplicate registry wallet webhook, credit already applied")
		h.persistCheckoutInvoice(r.Context(), tenantID, session, amountUSD)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "already processed"})
		return
	}

	if err := h.platformFees.CreditWallet(r.Context(), userID, amountUSD, session.ID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session_id": session.ID,
		}).Error("registry wallet webhook: credit failed")
		apierror.WriteError(w, apierror.NewInternal("Failed to credit wallet"))
		return
	}

	// Get updated wallet balance for notification
	newBalance, _ := h.platformFees.GetWalletBalance(r.Context(), userID)

	// Send notification about the wallet top-up
	if h.notificationSvc != nil {
		if err := h.notificationSvc.SendRegistryWalletTopUp(r.Context(), userID, amountUSD, newBalance); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"session_id":  session.ID,
				"user_id":     userID,
				"amount_usd":  amountUSD,
				"balance_usd": newBalance,
			}).Warn("failed to send registry wallet top-up notification after checkout")
		}
	}

	logrus.WithFields(logrus.Fields{
		"user_id":     userID,
		"tenant_id":   tenantID,
		"amount_usd":  amountUSD,
		"balance_usd": newBalance,
		"session_id":  session.ID,
	}).Info("registry wallet topped up via Stripe checkout")

	h.persistCheckoutInvoice(r.Context(), tenantID, session, amountUSD)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func checkoutReceiptURL(session *stripe.CheckoutSession) string {
	if session == nil || session.PaymentIntent == nil {
		return ""
	}
	pi := session.PaymentIntent
	if pi.LatestCharge != nil && pi.LatestCharge.ReceiptURL != "" {
		return pi.LatestCharge.ReceiptURL
	}
	return ""
}

func (h *StripeWebhookHandler) persistCheckoutInvoice(ctx context.Context, tenantID uuid.UUID, session *stripe.CheckoutSession, amountUSD float64) {
	if h.userRepo == nil || session == nil {
		return
	}
	amountCents := int(session.AmountTotal)
	if amountCents <= 0 {
		amountCents = int(math.Round(amountUSD * 100))
	}
	if amountCents <= 0 {
		return
	}
	curr := string(session.Currency)
	if curr == "" {
		curr = "usd"
	}
	rec := checkoutReceiptURL(session)
	if err := h.userRepo.CreatePaidInvoiceForStripeCheckoutSession(ctx, tenantID, amountCents, curr, session.ID, rec); err != nil {
		logrus.WithError(err).WithField("session_id", session.ID).Warn("stripe: persist checkout invoice failed")
	}
}

// processMembershipUpgrade creates activity feed items and awards achievements
// when a tenant's plan is upgraded via Stripe subscription.
func (h *StripeWebhookHandler) processMembershipUpgrade(ctx context.Context, tenantID uuid.UUID, oldPlan, newPlan string) {
	// Get all active users in the tenant
	users, err := h.userRepo.ListActiveUsersByTenant(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("stripe webhook: failed to list users for plan upgrade")
		return
	}

	// Determine upgrade description based on plan
	description := getUpgradeDescription(newPlan)
	isEnterprise := isEnterprisePlan(newPlan)
	wasEnterprise := isEnterprisePlan(oldPlan)

	for _, user := range users {
		// Create activity feed item
		activity := &storage.UserActivity{
			UserID:       user.ID,
			ActivityType: "membership_upgraded",
			Title:        fmt.Sprintf("Upgraded to %s", formatPlanName(newPlan)),
			Description:  description,
			Metadata: map[string]interface{}{
				"plan":         newPlan,
				"previousPlan": oldPlan,
				"upgradedAt":   time.Now().Format(time.RFC3339),
			},
			IsPublic: true,
		}

		if err := h.userRepo.CreateUserActivity(ctx, activity); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":   user.ID,
				"tenant_id": tenantID,
			}).Warn("stripe webhook: failed to create membership upgrade activity")
		}

		// Award enterprise achievement if upgrading to enterprise
		if isEnterprise && !wasEnterprise {
			if err := h.awardEnterpriseAchievement(ctx, user.ID); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"user_id":   user.ID,
					"tenant_id": tenantID,
				}).Warn("stripe webhook: failed to award enterprise achievement")
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":  tenantID,
		"old_plan":   oldPlan,
		"new_plan":   newPlan,
		"user_count": len(users),
	}).Info("stripe webhook: processed membership upgrade")
}

// awardEnterpriseAchievement awards the "Enterprise Pioneer" achievement to a user.
func (h *StripeWebhookHandler) awardEnterpriseAchievement(ctx context.Context, userID uuid.UUID) error {
	achievement, err := h.userRepo.GetAchievementBySlug(ctx, "enterprise_pioneer")
	if err != nil {
		return fmt.Errorf("failed to get enterprise pioneer achievement: %w", err)
	}

	if achievement == nil {
		logrus.Debug("Enterprise Pioneer achievement not found - skipping award")
		return nil
	}

	// Check if user already has this achievement
	existingAchievements, err := h.userRepo.GetUserAchievements(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing achievements: %w", err)
	}

	for _, ea := range existingAchievements {
		if ea.AchievementID == achievement.ID {
			return nil // User already has this achievement
		}
	}

	metadata := map[string]interface{}{
		"awarded_at": time.Now().Format(time.RFC3339),
		"reason":     "Upgraded to Enterprise plan",
	}

	if err := h.userRepo.AwardAchievement(userID, achievement.ID, metadata); err != nil {
		return fmt.Errorf("failed to award achievement: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":        userID,
		"achievement_id": achievement.ID,
	}).Info("stripe webhook: awarded Enterprise Pioneer achievement")

	return nil
}

// isEnterprisePlan checks if a plan name represents an enterprise tier.
func isEnterprisePlan(plan string) bool {
	return plan == "enterprise" || plan == "Enterprise"
}

// formatPlanName formats a plan name for display.
func formatPlanName(plan string) string {
	switch plan {
	case "enterprise", "Enterprise":
		return "Enterprise"
	case "professional", "pro", "Pro":
		return "Professional"
	case "starter", "Starter":
		return "Starter"
	case "free", "Free":
		return "Free"
	default:
		if len(plan) > 0 {
			return string(plan[0]-32) + plan[1:]
		}
		return plan
	}
}

// getUpgradeDescription returns a description based on the plan tier.
func getUpgradeDescription(plan string) string {
	switch plan {
	case "enterprise":
		return "Unlimited functions, dedicated support, and premium enterprise features"
	case "professional", "pro":
		return "Advanced features, priority support, and increased limits"
	case "starter":
		return "Expanded features and higher execution limits"
	default:
		return "Membership upgraded with new features and benefits"
	}
}

// handleInvoicePaymentFailed processes failed invoice payments and sends notifications.
func (h *StripeWebhookHandler) handleInvoicePaymentFailed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var invoice stripe.Invoice
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &invoice) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal invoice payment failed event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid invoice payload"))
		return
	}

	// Extract customer and subscription info for logging and potential notifications
	customerID := invoice.Customer.ID
	subscriptionID := ""
	if invoice.Parent != nil && invoice.Parent.SubscriptionDetails != nil && invoice.Parent.SubscriptionDetails.Subscription != nil {
		subscriptionID = invoice.Parent.SubscriptionDetails.Subscription.ID
	}

	logrus.WithFields(logrus.Fields{
		"customer_id":     customerID,
		"subscription_id": subscriptionID,
		"invoice_id":      invoice.ID,
		"amount_due":      invoice.AmountDue,
		"currency":        invoice.Currency,
		"attempt_count":   invoice.AttemptCount,
	}).Warn("stripe webhook: invoice payment failed")

	// Send notification if notification service is available
	if h.notificationSvc != nil && invoice.Customer != nil && invoice.Customer.Email != "" {
		// Create a billing alert notification
		ctx := r.Context()
		if err := h.notificationSvc.SendBillingAlert(ctx, invoice.Customer.Email, "payment_failed", map[string]interface{}{
			"invoice_id":    invoice.ID,
			"amount_due":    float64(invoice.AmountDue) / 100,
			"currency":      invoice.Currency,
			"attempt_count": invoice.AttemptCount,
			"next_attempt":  invoice.NextPaymentAttempt,
		}); err != nil {
			logrus.WithError(err).WithField("customer_email", invoice.Customer.Email).Warn("stripe webhook: failed to send payment failure notification")
		}
	}

	// Trigger dunning workflow for automated retry schedule
	if h.dunningManager != nil {
		ctx := r.Context()

		// Find tenant by Stripe customer ID
		tenant, err := h.userRepo.GetTenantByStripeCustomerID(ctx, customerID)
		if err != nil {
			logrus.WithError(err).WithField("customer_id", customerID).Warn("stripe webhook: failed to find tenant for dunning")
		} else if tenant != nil {
			// Convert subscription ID to UUID if present
			var subIDPtr *uuid.UUID
			if subscriptionID != "" {
				// Try to parse - this might fail if it's a Stripe ID not our UUID
				// In that case, we pass nil and let dunning manager work without it
				subUUID, err := uuid.Parse(subscriptionID)
				if err == nil {
					subIDPtr = &subUUID
				}
			}

			params := billingpkg.DunningInitiationParams{
				TenantID:         tenant.ID,
				SubscriptionID:   subIDPtr,
				InvoiceID:        invoice.ID,
				StripeCustomerID: customerID,
				CustomerEmail:    invoice.Customer.Email,
				AmountDueCents:   int(invoice.AmountDue),
				Currency:         string(invoice.Currency),
				FailureCode:      "payment_failed",
				FailureMessage:   "Invoice payment failed",
			}

			if _, err := h.dunningManager.InitiateDunningWorkflow(ctx, params); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"tenant_id":  tenant.ID,
					"invoice_id": invoice.ID,
				}).Warn("stripe webhook: failed to initiate dunning workflow")
			} else {
				logrus.WithFields(logrus.Fields{
					"tenant_id":  tenant.ID,
					"invoice_id": invoice.ID,
				}).Info("stripe webhook: dunning workflow initiated for failed payment")
			}
		} else {
			logrus.WithField("customer_id", customerID).Warn("stripe webhook: no tenant found for customer, skipping dunning")
		}
	}

	if h.pciAuditHelper != nil {
		ctx := r.Context()
		tenant, _ := h.userRepo.GetTenantByStripeCustomerID(ctx, customerID)
		if tenant != nil {
			actor := helpers.ActorContext{
				TenantID: &tenant.ID,
			}
			failureReason := "Invoice payment failed"
			h.pciAuditHelper.LogPaymentFlowAsync(ctx, actor, helpers.PaymentFlowParams{
				EventType:     "failed",
				TransactionID: invoice.ID,
				StripeEventID: &invoice.ID,
				AmountCents:   int(invoice.AmountDue),
				Currency:      string(invoice.Currency),
				PaymentMethod: "stripe_invoice",
				Details:       fmt.Sprintf("Invoice payment failed: %s", invoice.ID),
				Success:       false,
				FailureReason: &failureReason,
			})
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleInvoiceCreated processes newly created invoices and sends invoice ready emails.
func (h *StripeWebhookHandler) handleInvoiceCreated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var invoice stripe.Invoice
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &invoice) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal invoice created event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid invoice payload"))
		return
	}

	ctx := r.Context()

	// Log invoice creation
	logrus.WithFields(logrus.Fields{
		"customer_id":    invoice.Customer.ID,
		"invoice_id":     invoice.ID,
		"amount_due":     invoice.AmountDue,
		"currency":       invoice.Currency,
		"billing_reason": invoice.BillingReason,
		"period_start":   invoice.PeriodStart,
		"period_end":     invoice.PeriodEnd,
		"hosted_invoice": invoice.HostedInvoiceURL,
	}).Info("stripe webhook: invoice created")

	// Only send invoice emails for subscription invoices (not one-off payments)
	if invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCreate &&
		invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionCycle &&
		invoice.BillingReason != stripe.InvoiceBillingReasonSubscriptionUpdate {
		logrus.WithField("invoice_id", invoice.ID).Debug("skipping invoice email - not a subscription invoice")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "non_subscription_invoice"})
		return
	}

	// Skip if no amount due (free subscriptions, etc.)
	if invoice.AmountDue <= 0 {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "zero_amount"})
		return
	}

	// Send invoice email notification
	if h.notificationSvc != nil && invoice.Customer != nil && invoice.Customer.Email != "" {
		amountDueUSD := float64(invoice.AmountDue) / 100.0

		// Format period string
		period := "Subscription"
		if invoice.PeriodStart > 0 && invoice.PeriodEnd > 0 {
			start := time.Unix(invoice.PeriodStart, 0).Format("Jan 2, 2006")
			end := time.Unix(invoice.PeriodEnd, 0).Format("Jan 2, 2006")
			period = fmt.Sprintf("%s - %s", start, end)
		}

		invoiceURL := invoice.HostedInvoiceURL
		if invoiceURL == "" {
			invoiceURL = fmt.Sprintf("https://dashboard.functionfly.com/billing/invoices/%s", invoice.ID)
		}

		if err := h.notificationSvc.SendBillingInvoiceGenerated(ctx, invoice.Customer.Email, period, amountDueUSD, invoiceURL, invoice.ID); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"customer_email": invoice.Customer.Email,
				"invoice_id":     invoice.ID,
			}).Warn("stripe webhook: failed to send invoice generated notification")
		} else {
			logrus.WithFields(logrus.Fields{
				"customer_email": invoice.Customer.Email,
				"invoice_id":     invoice.ID,
				"amount_usd":     amountDueUSD,
			}).Info("stripe webhook: invoice generated notification sent")
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleInvoicePaymentSucceeded processes successful invoice payments and credits user wallet.
func (h *StripeWebhookHandler) handleInvoicePaymentSucceeded(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var invoice stripe.Invoice
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &invoice) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal invoice payment succeeded event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid invoice payload"))
		return
	}

	ctx := r.Context()
	amountPaidUSD := float64(invoice.AmountPaid) / 100.0

	// Log successful payment
	logrus.WithFields(logrus.Fields{
		"customer_id":    invoice.Customer.ID,
		"invoice_id":     invoice.ID,
		"amount_paid":    invoice.AmountPaid,
		"amount_usd":     amountPaidUSD,
		"currency":       invoice.Currency,
		"billing_reason": invoice.BillingReason,
	}).Info("stripe webhook: invoice payment succeeded")

	// Skip if no amount paid
	if invoice.AmountPaid <= 0 {
		logrus.Info("stripe webhook: invoice payment succeeded with zero amount, skipping wallet credit")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "zero_amount"})
		return
	}

	// Find tenant by Stripe customer ID
	tenant, err := h.userRepo.GetTenantByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", invoice.Customer.ID).Error("stripe webhook: failed to find tenant for invoice payment")
		// Don't fail the webhook - just log the error
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "tenant_not_found"})
		return
	}

	if tenant == nil {
		logrus.WithField("customer_id", invoice.Customer.ID).Warn("stripe webhook: no tenant found for customer ID")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "tenant_not_found"})
		return
	}

	// Get first active user in the tenant to credit wallet
	users, err := h.userRepo.ListActiveUsersByTenant(ctx, tenant.ID)
	if err != nil || len(users) == 0 {
		logrus.WithError(err).WithField("tenant_id", tenant.ID).Warn("stripe webhook: no active users found for tenant")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "no_active_users"})
		return
	}
	user := users[0]

	// Credit the user's wallet using platform fees repository
	if h.platformFees != nil {
		// Check if this invoice has already been processed
		reference := fmt.Sprintf("stripe_invoice:%s", invoice.ID)
		alreadyProcessed, err := h.platformFees.HasWalletCreditReference(ctx, reference)
		if err != nil {
			logrus.WithError(err).Error("stripe webhook: failed to check if invoice already processed")
			// Continue processing anyway
		} else if alreadyProcessed {
			logrus.WithField("invoice_id", invoice.ID).Info("stripe webhook: invoice already processed, skipping")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "already_processed"})
			return
		}

		// Credit the wallet
		if err := h.platformFees.CreditWallet(ctx, user.ID, amountPaidUSD, reference); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":    user.ID,
				"invoice_id": invoice.ID,
			}).Error("stripe webhook: failed to credit wallet for invoice payment")
			// Don't fail the webhook - the payment succeeded, we just couldn't credit the wallet
			// This should be handled by reconciliation later
		} else {
			logrus.WithFields(logrus.Fields{
				"user_id":    user.ID,
				"invoice_id": invoice.ID,
				"amount_usd": amountPaidUSD,
				"reference":  reference,
			}).Info("stripe webhook: credited wallet for invoice payment")

			// Send notification
			if h.notificationSvc != nil && user.Email != "" {
				// Get updated wallet balance for notification
				newBalance, _ := h.platformFees.GetWalletBalance(ctx, user.ID)
				if err := h.notificationSvc.SendRegistryWalletTopUp(ctx, user.ID, amountPaidUSD, newBalance); err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"user_id":     user.ID,
						"amount_usd":  amountPaidUSD,
						"balance_usd": newBalance,
					}).Warn("stripe webhook: failed to send wallet top-up notification")
				}
			}
		}
	} else {
		logrus.Warn("stripe webhook: platform fees repository not available, skipping wallet credit")
	}

	// Update subscription status if this is a subscription invoice
	if invoice.Parent != nil && invoice.Parent.SubscriptionDetails != nil &&
		invoice.Parent.SubscriptionDetails.Subscription != nil &&
		invoice.Parent.SubscriptionDetails.Subscription.ID != "" {
		subscriptionID := invoice.Parent.SubscriptionDetails.Subscription.ID
		if err := h.processSubscriptionInvoice(ctx, &invoice, user.ID, subscriptionID); err != nil {
			logrus.WithError(err).WithField("subscription_id", subscriptionID).Error("stripe webhook: failed to process subscription invoice")
		}

		// Send payment confirmation email for subscription payments
		if h.notificationSvc != nil && invoice.Customer != nil && invoice.Customer.Email != "" {
			period := "Subscription Period"
			if invoice.PeriodStart > 0 && invoice.PeriodEnd > 0 {
				start := time.Unix(invoice.PeriodStart, 0).Format("Jan 2, 2006")
				end := time.Unix(invoice.PeriodEnd, 0).Format("Jan 2, 2006")
				period = fmt.Sprintf("%s - %s", start, end)
			}

			if err := h.notificationSvc.SendBillingPaymentSuccess(ctx, invoice.Customer.Email, period, amountPaidUSD, invoice.ID); err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"customer_email": invoice.Customer.Email,
					"invoice_id":     invoice.ID,
				}).Warn("stripe webhook: failed to send payment success notification")
			} else {
				logrus.WithFields(logrus.Fields{
					"customer_email": invoice.Customer.Email,
					"invoice_id":     invoice.ID,
					"amount_usd":     amountPaidUSD,
				}).Info("stripe webhook: payment success notification sent")
			}
		}
	}

	// Record metrics
	monitoring.RecordStripeEventProcessed("invoice.payment_succeeded")

	if h.pciAuditHelper != nil {
		actor := helpers.ActorContext{
			TenantID: &tenant.ID,
			Email:    user.Email,
		}
		h.pciAuditHelper.LogPaymentFlowAsync(ctx, actor, helpers.PaymentFlowParams{
			EventType:     "processed",
			TransactionID: invoice.ID,
			StripeEventID: &invoice.ID,
			AmountCents:   int(invoice.AmountPaid),
			Currency:      string(invoice.Currency),
			PaymentMethod: "stripe_invoice",
			Details:       fmt.Sprintf("Invoice payment succeeded: %s", invoice.ID),
			Success:       true,
		})
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// processSubscriptionInvoice updates subscription records for subscription invoices
func (h *StripeWebhookHandler) processSubscriptionInvoice(ctx context.Context, invoice *stripe.Invoice, userID uuid.UUID, subscriptionID string) error {
	if subscriptionID == "" {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"subscription_id": subscriptionID,
		"invoice_id":      invoice.ID,
		"user_id":         userID,
	}).Info("stripe webhook: processing subscription invoice")

	// Update the subscription record in the database
	updates := map[string]interface{}{
		"last_payment_status": "succeeded",
		"last_payment_at":     time.Now(),
	}

	if invoice.PeriodStart > 0 {
		updates["current_period_start"] = time.Unix(invoice.PeriodStart, 0)
	}
	if invoice.PeriodEnd > 0 {
		updates["current_period_end"] = time.Unix(invoice.PeriodEnd, 0)
	}

	// Get subscription by Stripe ID
	subscription, err := h.userRepo.GetSubscriptionByStripeID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription != nil {
		if _, err := h.userRepo.UpdateSubscription(ctx, subscription.ID, updates); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// handlePaymentIntentFailed processes failed payment intents for immediate feedback.
func (h *StripeWebhookHandler) handlePaymentIntentFailed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var pi stripe.PaymentIntent
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &pi) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal payment intent failed event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payment intent payload"))
		return
	}

	// Log payment failure with detailed error information
	declineCode := ""
	if pi.LastPaymentError != nil && pi.LastPaymentError.DeclineCode != "" {
		declineCode = string(pi.LastPaymentError.DeclineCode)
	}

	logrus.WithFields(logrus.Fields{
		"payment_intent_id": pi.ID,
		"customer_id":       pi.Customer.ID,
		"amount":            pi.Amount,
		"currency":          pi.Currency,
		"status":            pi.Status,
		"decline_code":      declineCode,
		"error_message":     getPaymentErrorMessage(&pi),
	}).Warn("stripe webhook: payment intent failed")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// getPaymentErrorMessage extracts a user-friendly error message from a failed payment intent.
func getPaymentErrorMessage(pi *stripe.PaymentIntent) string {
	if pi.LastPaymentError == nil {
		return "Payment failed"
	}

	code := string(pi.LastPaymentError.DeclineCode)
	message := pi.LastPaymentError.Msg

	// Map common decline codes to user-friendly messages
	declineMessages := map[string]string{
		"insufficient_funds":     "Your card has insufficient funds.",
		"lost_card":              "This card has been reported lost. Please use a different card.",
		"stolen_card":            "This card has been reported stolen. Please use a different card.",
		"expired_card":           "Your card has expired. Please check the expiration date or use a different card.",
		"incorrect_cvc":          "Your card's security code is incorrect. Please check and try again.",
		"processing_error":       "An error occurred while processing your card. Please try again.",
		"issuer_not_available":   "Your card issuer is temporarily unavailable. Please try again later.",
		"try_again_later":        "A temporary error occurred. Please try again later.",
		"fraudulent":             "This payment was flagged as potentially fraudulent. Please contact your bank or use a different payment method.",
		"card_not_supported":     "This card does not support this type of purchase. Please use a different card.",
		"currency_not_supported": "This card does not support the selected currency. Please use a different card.",
	}

	if msg, ok := declineMessages[code]; ok {
		return msg
	}

	if message != "" {
		return message
	}

	return "Payment failed. Please try again or use a different payment method."
}

// handleBundleSubscriptionCreated processes bundle subscriptions when created from founder mode conversion
func (h *StripeWebhookHandler) handleBundleSubscriptionCreated(ctx context.Context, subscription *stripe.Subscription, tenantID uuid.UUID) error {
	// Check if this is a bundle subscription by looking at the metadata
	bundleSlug := subscription.Metadata["bundle_slug"]
	if bundleSlug == "" {
		// Not a bundle subscription
		return nil
	}

	founderModeID := subscription.Metadata["founder_mode_id"]
	if founderModeID == "" {
		// Not a conversion from founder mode
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":       tenantID,
		"subscription_id": subscription.ID,
		"bundle_slug":     bundleSlug,
		"founder_mode_id": founderModeID,
	}).Info("stripe webhook: processing bundle subscription from founder mode conversion")

	// Parse founder mode ID
	fmid, err := uuid.Parse(founderModeID)
	if err != nil {
		return fmt.Errorf("invalid founder_mode_id: %w", err)
	}

	now := time.Now().UTC()

	// Create bundle subscription record
	bundleSub := &storage.BundleSubscription{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		StripeSubscriptionID: subscription.ID,
		Status:               "active",
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     now.AddDate(0, 1, 0), // Default 1 month, will be updated by subscription webhooks
	}

	// Link to founder mode if provided
	bundleSub.FounderModeID = &fmid
	bundleSub.ConvertedFromFounderMode = true

	// Get bundle by slug to set BundleID
	bundle, err := h.userRepo.GetPricingBundleBySlug(ctx, bundleSlug)
	if err != nil {
		logrus.WithError(err).WithField("bundle_slug", bundleSlug).Warn("failed to get bundle for subscription")
	} else if bundle != nil {
		bundleSub.BundleID = bundle.ID
	}

	// Create the bundle subscription record
	if err := h.userRepo.CreateBundleSubscription(ctx, bundleSub); err != nil {
		return fmt.Errorf("failed to create bundle subscription: %w", err)
	}

	// Update founder mode status to converted
	if err := h.userRepo.UpdateFounderModeStatus(ctx, fmid, "converted"); err != nil {
		logrus.WithError(err).WithField("founder_mode_id", fmid).Warn("failed to update founder mode status")
		// Don't fail - the subscription was created
	}

	// Send notification
	if h.notificationSvc != nil {
		users, err := h.userRepo.ListActiveUsersByTenant(ctx, tenantID)
		if err == nil && len(users) > 0 && users[0].Email != "" {
			if err := h.notificationSvc.SendBillingAlert(ctx, users[0].Email, "bundle_converted", map[string]interface{}{
				"bundle_slug":     bundleSlug,
				"subscription_id": subscription.ID,
			}); err != nil {
				logrus.WithError(err).Warn("failed to send bundle conversion notification")
			}
		}
	}

	return nil
}

// handleChargeDisputeCreated handles when a dispute is created (customer files chargeback)
func (h *StripeWebhookHandler) handleChargeDisputeCreated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	if h.disputeRepo == nil {
		logrus.Warn("stripe webhook: dispute repository not configured, skipping dispute handling")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "dispute_repo_not_configured"})
		return
	}

	var dispute stripe.Dispute
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &dispute) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal dispute created event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid dispute payload"))
		return
	}

	ctx := r.Context()
	amountUSD := float64(dispute.Amount) / 100.0

	// Log the dispute
	logrus.WithFields(logrus.Fields{
		"dispute_id":   dispute.ID,
		"charge_id":    dispute.Charge.ID,
		"amount_cents": dispute.Amount,
		"currency":     dispute.Currency,
		"reason":       dispute.Reason,
		"status":       dispute.Status,
	}).Warn("stripe webhook: chargeback dispute created")

	// Extract tenant/user info from charge metadata if available
	var tenantID, userID *uuid.UUID
	var metadata map[string]interface{}

	// Try to find the original charge to get metadata
	if dispute.Charge != nil && dispute.Charge.PaymentIntent != nil && dispute.Charge.PaymentIntent.ID != "" {
		// Use the payment intent ID to find tenant/user if available
		piID := dispute.Charge.PaymentIntent.ID
		metadata = map[string]interface{}{
			"payment_intent_id": piID,
			"charge_id":         dispute.Charge.ID,
		}
	}

	// Build evidence details
	var evidenceDueBy *time.Time
	if dispute.EvidenceDetails != nil && dispute.EvidenceDetails.DueBy > 0 {
		due := time.Unix(dispute.EvidenceDetails.DueBy, 0)
		evidenceDueBy = &due
	}

	// Create or update dispute record
	paymentDispute := &storage.PaymentDispute{
		StripeDisputeID: dispute.ID,
		StripeChargeID:  dispute.Charge.ID,
		StripePaymentID: "", // Will be populated from charge lookup if available
		AmountCents:     int(dispute.Amount),
		Currency:        string(dispute.Currency),
		Reason:          string(dispute.Reason),
		Status:          string(dispute.Status),
		EvidenceDueBy:   evidenceDueBy,
		Metadata:        h.toDatatypesJSON(metadata),
	}

	// Link to tenant/user if available in charge metadata
	if dispute.Charge != nil && len(dispute.Charge.Metadata) > 0 {
		if tidStr := dispute.Charge.Metadata["tenant_id"]; tidStr != "" {
			if tid, err := uuid.Parse(tidStr); err == nil {
				tenantID = &tid
				paymentDispute.TenantID = tenantID
			}
		}
		if uidStr := dispute.Charge.Metadata["user_id"]; uidStr != "" {
			if uid, err := uuid.Parse(uidStr); err == nil {
				userID = &uid
				paymentDispute.UserID = userID
			}
		}
		if piID := dispute.Charge.PaymentIntent; piID != nil && piID.ID != "" {
			paymentDispute.StripePaymentID = piID.ID
		}
	}

	// Save to database
	if err := h.disputeRepo.UpsertDispute(ctx, paymentDispute); err != nil {
		logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("stripe webhook: failed to save dispute")
		apierror.WriteError(w, apierror.NewInternal("Failed to save dispute"))
		return
	}

	// Send notifications to admin users
	if h.notificationSvc != nil {
		adminUsers, _ := h.getAdminUsers(ctx)
		if len(adminUsers) > 0 {
			evidenceDueByStr := "unknown"
			if evidenceDueBy != nil {
				evidenceDueByStr = evidenceDueBy.Format("2006-01-02")
			}
			h.notificationSvc.SendDisputeCreated(ctx, adminUsers, dispute.ID, fmt.Sprintf("%.2f", amountUSD),
				string(dispute.Currency), string(dispute.Reason), evidenceDueByStr)
		}
	}

	// Trigger automated dispute response workflow if configured
	if h.disputeResponseManager != nil {
		if err := h.disputeResponseManager.HandleDisputeCreated(ctx, &dispute, paymentDispute); err != nil {
			logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("stripe webhook: dispute response manager failed")
		}
	}

	if h.pciAuditHelper != nil {
		actor := helpers.ActorContext{
			TenantID: tenantID,
		}
		h.pciAuditHelper.LogPaymentFlowAsync(ctx, actor, helpers.PaymentFlowParams{
			EventType:     "chargeback",
			TransactionID: dispute.Charge.ID,
			StripeEventID: &dispute.ID,
			AmountCents:   int(dispute.Amount),
			Currency:      string(dispute.Currency),
			PaymentMethod: "stripe_dispute",
			Details:       fmt.Sprintf("Chargeback received: %s, reason: %s", dispute.ID, dispute.Reason),
			Success:       false,
		})
	}

	monitoring.RecordStripeEventProcessed("charge.dispute.created")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleChargeDisputeUpdated handles updates to a dispute
func (h *StripeWebhookHandler) handleChargeDisputeUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	if h.disputeRepo == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "dispute_repo_not_configured"})
		return
	}

	var dispute stripe.Dispute
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &dispute) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal dispute updated event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid dispute payload"))
		return
	}

	ctx := r.Context()

	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"status":     dispute.Status,
		"reason":     dispute.Reason,
	}).Info("stripe webhook: dispute updated")

	// Get existing dispute
	existingDispute, err := h.disputeRepo.GetDisputeByStripeID(ctx, dispute.ID)
	if err != nil {
		logrus.WithError(err).WithField("dispute_id", dispute.ID).Warn("stripe webhook: failed to find existing dispute")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "dispute_not_found"})
		return
	}

	if existingDispute == nil {
		// Create it if it doesn't exist
		h.handleChargeDisputeCreated(w, r, event)
		return
	}

	// Update status if changed
	newStatus := string(dispute.Status)
	if existingDispute.Status != newStatus {
		if err := h.disputeRepo.UpdateDisputeStatus(ctx, existingDispute.ID, newStatus, "", ""); err != nil {
			logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("stripe webhook: failed to update dispute status")
		}
	}

	// Trigger automated dispute response workflow if configured
	if h.disputeResponseManager != nil {
		if err := h.disputeResponseManager.HandleDisputeUpdated(ctx, &dispute, existingDispute); err != nil {
			logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("stripe webhook: dispute response manager update failed")
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleChargeDisputeClosed handles when a dispute is closed (won, lost, or withdrawn)
func (h *StripeWebhookHandler) handleChargeDisputeClosed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	if h.disputeRepo == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "dispute_repo_not_configured"})
		return
	}

	var dispute stripe.Dispute
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &dispute) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal dispute closed event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid dispute payload"))
		return
	}

	ctx := r.Context()
	amountUSD := float64(dispute.Amount) / 100.0

	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"status":     dispute.Status,
		"amount_usd": amountUSD,
	}).Info("stripe webhook: dispute closed")

	// Get existing dispute
	existingDispute, err := h.disputeRepo.GetDisputeByStripeID(ctx, dispute.ID)
	if err != nil {
		logrus.WithError(err).WithField("dispute_id", dispute.ID).Warn("stripe webhook: failed to find existing dispute")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed", "note": "dispute_not_found"})
		return
	}

	if existingDispute != nil {
		status := string(dispute.Status)

		if err := h.disputeRepo.UpdateDisputeStatus(ctx, existingDispute.ID, status, status, ""); err != nil {
			logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("stripe webhook: failed to update dispute status")
		}

		// Send notification about resolution
		if h.notificationSvc != nil {
			adminUsers, _ := h.getAdminUsers(ctx)
			if len(adminUsers) > 0 {
				// Won if status indicates we won
				won := status == "won" || status == "charge_refunded"
				h.notificationSvc.SendDisputeResolved(ctx, adminUsers, dispute.ID, status, amountUSD, won)
			}
		}

		// Trigger automated dispute response workflow if configured
		if h.disputeResponseManager != nil {
			if err := h.disputeResponseManager.HandleDisputeClosed(ctx, &dispute, existingDispute); err != nil {
				logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("stripe webhook: dispute response manager closed failed")
			}
		}
	}

	monitoring.RecordStripeEventProcessed("charge.dispute.closed")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleChargeDisputeFundsWithdrawn handles when funds are withdrawn from the account after losing a dispute
func (h *StripeWebhookHandler) handleChargeDisputeFundsWithdrawn(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	if h.disputeRepo == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "dispute_repo_not_configured"})
		return
	}

	var dispute stripe.Dispute
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &dispute) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal funds withdrawn event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid dispute payload"))
		return
	}

	ctx := r.Context()
	amountUSD := float64(dispute.Amount) / 100.0

	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"amount_usd": amountUSD,
	}).Warn("stripe webhook: chargeback funds withdrawn from account")

	// Send notification about funds withdrawal
	if h.notificationSvc != nil {
		adminUsers, _ := h.getAdminUsers(ctx)
		if len(adminUsers) > 0 {
			h.notificationSvc.SendChargebackFundsWithdrawn(ctx, adminUsers, dispute.ID, amountUSD)
		}
	}

	monitoring.RecordStripeEventProcessed("charge.dispute.funds_withdrawn")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleChargeRefunded handles when a charge is refunded
func (h *StripeWebhookHandler) handleChargeRefunded(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var charge stripe.Charge
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &charge) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal charge refunded event")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid charge payload"))
		return
	}

	ctx := r.Context()

	// Process each refund on the charge
	for _, refund := range charge.Refunds.Data {
		amountUSD := float64(refund.Amount) / 100.0

		logrus.WithFields(logrus.Fields{
			"refund_id":  refund.ID,
			"charge_id":  charge.ID,
			"amount_usd": amountUSD,
			"reason":     refund.Reason,
			"status":     refund.Status,
		}).Info("stripe webhook: charge refunded")

		// Save refund record if repository is available
		if h.refundRepo != nil {
			paymentRefund := &storage.PaymentRefund{
				StripeRefundID:  refund.ID,
				StripeChargeID:  charge.ID,
				StripePaymentID: "", // Would need to map from charge
				AmountCents:     int(refund.Amount),
				Currency:        string(charge.Currency),
				Status:          string(refund.Status),
				Reason:          string(refund.Reason),
				ReceiptNumber:   refund.ReceiptNumber,
			}

			// Try to extract tenant/user from charge metadata
			if len(charge.Metadata) > 0 {
				if tidStr := charge.Metadata["tenant_id"]; tidStr != "" {
					if tid, err := uuid.Parse(tidStr); err == nil {
						paymentRefund.TenantID = &tid
					}
				}
				if uidStr := charge.Metadata["user_id"]; uidStr != "" {
					if uid, err := uuid.Parse(uidStr); err == nil {
						paymentRefund.UserID = &uid
					}
				}
				if piID := charge.PaymentIntent; piID != nil && piID.ID != "" {
					paymentRefund.StripePaymentID = piID.ID
				}
			}

			// Get failure reason if failed
			if refund.FailureReason != "" {
				paymentRefund.FailureReason = string(refund.FailureReason)
			}

			if err := h.refundRepo.UpsertRefund(ctx, paymentRefund); err != nil {
				logrus.WithError(err).WithField("refund_id", refund.ID).Error("stripe webhook: failed to save refund")
			}

			// Send notification
			if h.notificationSvc != nil {
				adminUsers, _ := h.getAdminUsers(ctx)
				if len(adminUsers) > 0 {
					var tenantIDStr *string
					if paymentRefund.TenantID != nil {
						s := paymentRefund.TenantID.String()
						tenantIDStr = &s
					}
					h.notificationSvc.SendRefundProcessed(ctx, adminUsers, refund.ID, amountUSD,
						string(refund.Reason), tenantIDStr)
				}
			}

			if h.pciAuditHelper != nil {
				var tenantID *uuid.UUID
				if paymentRefund.TenantID != nil {
					tenantID = paymentRefund.TenantID
				}
				actor := helpers.ActorContext{
					TenantID: tenantID,
				}
				h.pciAuditHelper.LogPaymentFlowAsync(ctx, actor, helpers.PaymentFlowParams{
					EventType:     "refunded",
					TransactionID: charge.ID,
					StripeEventID: &refund.ID,
					AmountCents:   int(refund.Amount),
					Currency:      string(charge.Currency),
					PaymentMethod: "stripe_refund",
					Details:       fmt.Sprintf("Charge refunded: %s, reason: %s", refund.ID, refund.Reason),
					Success:       refund.Status == "succeeded",
				})
			}
		}

		// Note: If the refund is related to a dispute, it would be handled by
		// the dispute webhook events. The refund record can be linked later
		// by querying for the payment intent or charge ID.
	}

	monitoring.RecordStripeEventProcessed("charge.refunded")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// getAdminUsers retrieves a list of admin user IDs for notifications
func (h *StripeWebhookHandler) getAdminUsers(ctx context.Context) ([]uuid.UUID, error) {
	// Query for users with admin role
	// This is a simplified implementation - in production you'd want to
	// use a more sophisticated query or caching mechanism
	return []uuid.UUID{}, nil // Return empty for now - notification service handles nil gracefully
}

// toDatatypesJSON converts a map to GORM datatypes.JSON
func (h *StripeWebhookHandler) toDatatypesJSON(m map[string]interface{}) datatypes.JSON {
	if m == nil {
		return datatypes.JSON([]byte("{}"))
	}
	data, _ := json.Marshal(m)
	return datatypes.JSON(data)
}

// logStripeEvent creates an audit log entry for Stripe webhook events
func (h *StripeWebhookHandler) logStripeEvent(ctx context.Context, event *stripe.Event, obj interface{}) *storage.StripeSyncEvent {
	// Determine the object ID based on the event type
	var objectID string
	switch v := obj.(type) {
	case *stripe.Subscription:
		objectID = v.ID
	case *stripe.PaymentMethod:
		objectID = v.ID
	case *stripe.Customer:
		objectID = v.ID
	default:
		objectID = ""
	}

	// Create the sync event record
	syncEvent := &storage.StripeSyncEvent{
		StripeEventID:  event.ID,
		StripeObjectID: objectID,
		EventType:      string(event.Type),
		EventData:      event.Data.Raw,
		Status:         storage.StripeSyncStatusPending,
	}

	// Add idempotency key if available
	if event.Request != nil && event.Request.IdempotencyKey != "" {
		syncEvent.IdempotencyKey = &event.Request.IdempotencyKey
	}

	// Try to determine the tenant from the object
	switch v := obj.(type) {
	case *stripe.Subscription:
		if v.Customer != nil && v.Customer.ID != "" {
			tenant, err := h.userRepo.GetTenantByStripeCustomerID(ctx, v.Customer.ID)
			if err == nil && tenant != nil {
				syncEvent.TenantID = &tenant.ID
			}
		}
	}

	// Save to database for audit trail
	if created, err := h.userRepo.CreateStripeSyncEvent(ctx, syncEvent); err == nil {
		return created
	}

	// If we couldn't save, just return the unsaved event
	return syncEvent
}

// handlePaymentMethodUpdated processes payment method update events from Stripe
func (h *StripeWebhookHandler) handlePaymentMethodUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var pm stripe.PaymentMethod
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &pm) != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payment method payload"))
		return
	}

	ctx := r.Context()

	// Log the event
	h.logStripeEvent(ctx, event, &pm)

	// Find tenant by customer ID
	if pm.Customer == nil || pm.Customer.ID == "" {
		logrus.WithField("payment_method_id", pm.ID).Warn("payment method update: no customer associated")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "no_customer"})
		return
	}

	tenant, err := h.userRepo.GetTenantByStripeCustomerID(ctx, pm.Customer.ID)
	if err != nil || tenant == nil {
		logrus.WithField("customer_id", pm.Customer.ID).Warn("payment method update: tenant not found")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "tenant_not_found"})
		return
	}

	// Only handle card payment methods for now
	if pm.Card == nil {
		logrus.WithField("payment_method_id", pm.ID).Debug("payment method update: non-card type, skipping")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "non_card_type"})
		return
	}

	// Build billing details JSON
	billingDetails := map[string]interface{}{}
	if pm.BillingDetails.Name != "" {
		billingDetails["name"] = pm.BillingDetails.Name
	}
	if pm.BillingDetails.Email != "" {
		billingDetails["email"] = pm.BillingDetails.Email
	}
	if pm.BillingDetails.Address != nil {
		billingDetails["address"] = map[string]interface{}{
			"line1":       pm.BillingDetails.Address.Line1,
			"line2":       pm.BillingDetails.Address.Line2,
			"city":        pm.BillingDetails.Address.City,
			"state":       pm.BillingDetails.Address.State,
			"postal_code": pm.BillingDetails.Address.PostalCode,
			"country":     pm.BillingDetails.Address.Country,
		}
	}
	billingDetailsJSON, _ := json.Marshal(billingDetails)

	// Update tenant payment method info
	paymentMethod := &storage.PaymentMethodInfoExtended{
		StripePaymentMethodID: pm.ID,
		Brand:                 string(pm.Card.Brand),
		Last4:                 pm.Card.Last4,
		ExpMonth:              int(pm.Card.ExpMonth),
		ExpYear:               int(pm.Card.ExpYear),
		BillingDetails:        billingDetailsJSON,
	}

	if err := h.userRepo.UpdateTenantPaymentMethod(ctx, tenant.ID, paymentMethod); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id":      tenant.ID,
			"payment_method": pm.ID,
		}).Error("failed to update tenant payment method")
		apierror.WriteError(w, apierror.NewInternal("Failed to update payment method"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":         tenant.ID,
		"payment_method_id": pm.ID,
		"brand":             pm.Card.Brand,
		"last4":             pm.Card.Last4,
	}).Info("payment method updated from stripe")

	if h.pciAuditHelper != nil {
		pmID, _ := uuid.Parse(pm.ID)
		brand := string(pm.Card.Brand)
		last4 := pm.Card.Last4
		actor := helpers.ActorContext{
			TenantID: &tenant.ID,
		}
		h.pciAuditHelper.LogCardDataAccessAsync(ctx, actor, helpers.CardDataAccessParams{
			AccessType:      "write",
			DataType:        "card_details",
			PaymentMethodID: &pmID,
			CardLastFour:    &last4,
			CardBrand:       &brand,
			CardExpiryMonth: func() *int { m := int(pm.Card.ExpMonth); return &m }(),
			CardExpiryYear:  func() *int { y := int(pm.Card.ExpYear); return &y }(),
			Purpose:         "Payment method updated via Stripe webhook",
			CDESection:      "cardholder_data",
			Success:         true,
		})
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":            "success",
		"tenant_id":         tenant.ID.String(),
		"payment_method_id": pm.ID,
	})
}

// handlePaymentMethodDetached processes payment method detachment events
func (h *StripeWebhookHandler) handlePaymentMethodDetached(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var pm stripe.PaymentMethod
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &pm) != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payment method payload"))
		return
	}

	ctx := r.Context()

	// Log the event
	h.logStripeEvent(ctx, event, &pm)

	// Mark the payment method as inactive in our records
	// We keep it for audit purposes but mark it as non-default
	existingPM, err := h.userRepo.GetPaymentMethodByStripeID(ctx, pm.ID)
	if err != nil || existingPM == nil {
		// Payment method not found - it was never synced
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "not_found"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":         existingPM.TenantID,
		"payment_method_id": pm.ID,
	}).Info("payment method detached in stripe")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":            "success",
		"action":            "detached",
		"payment_method_id": pm.ID,
	})
}

// handleCustomerUpdated processes customer update events from Stripe
func (h *StripeWebhookHandler) handleCustomerUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var customer stripe.Customer
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &customer) != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid customer payload"))
		return
	}

	ctx := r.Context()

	// Log the event
	h.logStripeEvent(ctx, event, &customer)

	// Find tenant by customer ID
	tenant, err := h.userRepo.GetTenantByStripeCustomerID(ctx, customer.ID)
	if err != nil || tenant == nil {
		logrus.WithField("customer_id", customer.ID).Warn("customer update: tenant not found")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "tenant_not_found"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":   tenant.ID,
		"customer_id": customer.ID,
		"email":       customer.Email,
	}).Debug("customer updated from stripe")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"tenant_id":   tenant.ID.String(),
		"customer_id": customer.ID,
	})
}

// SetPayoutService sets the payout service for handling payout-related webhook events.
func (h *StripeWebhookHandler) SetPayoutService(ps PayoutWebhookProcessor) {
	h.payoutService = ps
}

// handlePayoutPaid processes payout.paid events from Stripe.
// This fires when a Stripe payout to a connected account's bank succeeds.
func (h *StripeWebhookHandler) handlePayoutPaid(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payload"))
		return
	}

	var payout struct {
		ID      string `json:"id"`
		Amount  int64  `json:"amount"`
		Status  string `json:"status"`
		Account string `json:"account"`
	}
	if err := json.Unmarshal(raw, &payout); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payout payload"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"payout_id":  payout.ID,
		"account_id": payout.Account,
		"amount":     payout.Amount,
		"status":     payout.Status,
	}).Info("Stripe payout.paid event received")

	if h.payoutService != nil {
		if err := h.payoutService.ProcessPayoutPaid(r.Context(), payout.ID, payout.Account); err != nil {
			logrus.WithError(err).WithField("payout_id", payout.ID).Error("failed to process payout.paid event")
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handlePayoutFailed processes payout.failed events from Stripe.
// This fires when a Stripe payout to a connected account's bank fails.
func (h *StripeWebhookHandler) handlePayoutFailed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payload"))
		return
	}

	var payout struct {
		ID             string `json:"id"`
		Amount         int64  `json:"amount"`
		Status         string `json:"status"`
		Account        string `json:"account"`
		FailureCode    string `json:"failure_code"`
		FailureMessage string `json:"failure_message"`
	}
	if err := json.Unmarshal(raw, &payout); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payout payload"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"payout_id":       payout.ID,
		"account_id":      payout.Account,
		"amount":          payout.Amount,
		"failure_code":    payout.FailureCode,
		"failure_message": payout.FailureMessage,
	}).Warn("Stripe payout.failed event received")

	if h.payoutService != nil {
		if err := h.payoutService.RefreshAccountStatus(r.Context(), payout.Account); err != nil {
			logrus.WithError(err).WithField("account_id", payout.Account).Warn("failed to refresh account status after payout failure")
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleTransferReversed processes transfer.reversed events from Stripe.
// This fires when a transfer to a connected account is reversed (e.g., due to a dispute).
func (h *StripeWebhookHandler) handleTransferReversed(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payload"))
		return
	}

	var txfr struct {
		ID       string `json:"id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
		Dest     string `json:"destination"`
	}
	if err := json.Unmarshal(raw, &txfr); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid transfer payload"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"transfer_id": txfr.ID,
		"destination": txfr.Dest,
		"amount":      txfr.Amount,
	}).Warn("Stripe transfer.reversed event received")

	if h.payoutService != nil {
		if err := h.payoutService.ProcessTransferReversed(r.Context(), txfr.ID); err != nil {
			logrus.WithError(err).WithField("transfer_id", txfr.ID).Error("failed to process transfer.reversed")
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleConnectAccountUpdated processes account.updated events for Stripe Connect accounts.
func (h *StripeWebhookHandler) handleConnectAccountUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid payload"))
		return
	}

	var account struct {
		ID               string `json:"id"`
		PayoutsEnabled   bool   `json:"payouts_enabled"`
		DetailsSubmitted bool   `json:"details_submitted"`
		ChargesEnabled   bool   `json:"charges_enabled"`
	}
	if err := json.Unmarshal(raw, &account); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid account payload"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"account_id":        account.ID,
		"payouts_enabled":   account.PayoutsEnabled,
		"details_submitted": account.DetailsSubmitted,
		"charges_enabled":   account.ChargesEnabled,
	}).Info("Stripe Connect account.updated event received")

	if h.payoutService != nil {
		if err := h.payoutService.RefreshAccountStatus(r.Context(), account.ID); err != nil {
			logrus.WithError(err).WithField("account_id", account.ID).Warn("failed to refresh connect account status")
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

func (h *StripeWebhookHandler) handleCertExamCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	examIDStr := session.Metadata["exam_id"]
	tierSlug := session.Metadata["tier_slug"]
	userIDStr := session.Metadata["user_id"]

	logrus.WithFields(logrus.Fields{
		"exam_id":         examIDStr,
		"tier_slug":       tierSlug,
		"session_id":      session.ID,
		"payment_intent":  session.PaymentIntent,
		"payment_status":  session.PaymentStatus,
		"webhook_user_id": userIDStr,
	}).Info("handleCertExamCheckout called")

	if examIDStr == "" || userIDStr == "" {
		logrus.Warn("cert exam checkout missing required metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Missing metadata"))
		return
	}

	examID, err := uuid.Parse(examIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid exam_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid exam_id"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid user_id in metadata")
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user_id"))
		return
	}

	var paymentIntentID string
	if session.PaymentIntent != nil {
		paymentIntentID = session.PaymentIntent.ID
	}

	logrus.WithFields(logrus.Fields{
		"exam_id":       examIDStr,
		"user_id":       userIDStr,
		"pi_id":         paymentIntentID,
		"cert_repo_nil": h.certRepo == nil,
	}).Info("cert exam checkout: before update")

	// Update exam with Stripe payment ID and activate it
	if err := h.certRepo.UpdateExamStripePaymentID(r.Context(), examID, paymentIntentID); err != nil {
		logrus.WithError(err).WithField("exam_id", examIDStr).Error("failed to update exam with payment ID")
		apierror.WriteError(w, apierror.NewInternal("Failed to update exam"))
		return
	}

	logrus.WithField("exam_id", examIDStr).Info("cert exam checkout: stripe payment ID updated")

	// Activate the exam (select questions and set status to in_progress) — validates ownership
	if err := h.certRepo.ActivateExamFromPaymentWithUser(r.Context(), examID, userID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"exam_id": examIDStr,
			"user_id": userIDStr,
		}).Error("failed to activate exam from payment")
		apierror.WriteError(w, apierror.NewInternal(fmt.Sprintf("Failed to activate exam: %v", err)))
		return
	}

	logrus.WithFields(logrus.Fields{
		"exam_id":    examIDStr,
		"user_id":    userIDStr,
		"session_id": session.ID,
		"tier_slug":  tierSlug,
	}).Info("cert exam payment completed and exam activated")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}
