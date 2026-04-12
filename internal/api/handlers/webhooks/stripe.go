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
	"time"

	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	stripeSub "github.com/stripe/stripe-go/v83/subscription"
	"github.com/stripe/stripe-go/v83/webhook"
)

// StripeWebhookHandler handles Stripe webhook events for payment processing.
type StripeWebhookHandler struct {
	financialTxRepo *storage.FinancialTransactionRepository
	billingCtrl     *billing.Controller
	notificationSvc *notification.Service
	userRepo        storage.Repository
	platformFees    *storageregistry.PlatformFeeRepository
	sfAddons        *statefabricaddons.Repository
	webhookSecret   string
}

// NewStripeWebhookHandler creates a new Stripe webhook handler.
func NewStripeWebhookHandler(
	financialTxRepo *storage.FinancialTransactionRepository,
	billingCtrl *billing.Controller,
	notificationSvc *notification.Service,
	userRepo storage.Repository,
	platformFees *storageregistry.PlatformFeeRepository,
	sfAddons *statefabricaddons.Repository,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		financialTxRepo: financialTxRepo,
		billingCtrl:     billingCtrl,
		notificationSvc: notificationSvc,
		userRepo:        userRepo,
		platformFees:    platformFees,
		sfAddons:        sfAddons,
		webhookSecret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}
}

// RegisterRoutes registers webhook routes.
func (h *StripeWebhookHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/webhooks/stripe", h.HandleWebhook).Methods("POST")
}

// HandleWebhook processes incoming Stripe webhook events.
func (h *StripeWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("failed to read webhook body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Enforce webhook secret by default in all environments
	// Only allow unverified webhooks in development if explicitly opted in via ALLOW_UNVERIFIED_WEBHOOKS=true
	if h.webhookSecret == "" {
		if os.Getenv("PRODUCTION") == "true" {
			logrus.Error("STRIPE_WEBHOOK_SECRET not configured in production - rejecting webhook")
			http.Error(w, "Webhook authentication not configured", http.StatusInternalServerError)
			return
		}
		// In non-production, only allow unverified webhooks if explicitly opted in
		if os.Getenv("ALLOW_UNVERIFIED_WEBHOOKS") != "true" {
			logrus.Error("STRIPE_WEBHOOK_SECRET not configured - rejecting webhook. Set ALLOW_UNVERIFIED_WEBHOOKS=true to allow unverified webhooks in development (not recommended)")
			http.Error(w, "Webhook authentication not configured", http.StatusInternalServerError)
			return
		}
		// Development mode with explicit opt-in: parse without verification
		logrus.Warn("Processing unverified webhook - ALLOW_UNVERIFIED_WEBHOOKS is enabled (development only)")
		var event stripe.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			logrus.WithError(err).Warn("failed to parse stripe webhook event")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		h.handleEvent(w, r, &event)
		return
	}

	// Verify webhook signature with configured secret
	signature := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signature, h.webhookSecret)
	if err != nil {
		logrus.WithError(err).Warn("invalid stripe webhook signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
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
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(w, r, event)
	case "invoice.payment_succeeded":
		h.handleInvoicePaymentSucceeded(w, r, event)
	case "payment_intent.payment_failed":
		h.handlePaymentIntentFailed(w, r, event)
	default:
		// Acknowledge but ignore other events
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		monitoring.RecordBillingWebhookReceived(string(event.Type), "ignored")
		monitoring.RecordBillingWebhookProcessingDuration(string(event.Type), time.Since(start))
	}
}

func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var session stripe.CheckoutSession
	sessionData, err := event.Data.Raw.MarshalJSON()
	if err != nil {
		logrus.WithError(err).Error("failed to marshal checkout session data")
		http.Error(w, "Invalid event data", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(sessionData, &session); err != nil {
		logrus.WithError(err).Error("failed to unmarshal checkout session")
		http.Error(w, "Invalid session data", http.StatusBadRequest)
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
	case "function_verification":
		h.handleFunctionVerificationCheckout(w, r, &session)
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
		http.Error(w, "Missing metadata", http.StatusBadRequest)
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
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
		http.Error(w, "Failed to persist entitlement", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleFunctionVerificationCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	// Extract metadata
	paymentIDStr := session.Metadata["payment_id"]
	functionIDStr := session.Metadata["function_id"]
	tenantIDStr := session.Metadata["tenant_id"]

	if paymentIDStr == "" || functionIDStr == "" {
		logrus.Warn("function verification checkout missing required metadata")
		http.Error(w, "Missing metadata", http.StatusBadRequest)
		return
	}

	// Lookup payment record by checkout session ID
	payment, err := h.userRepo.GetFunctionVerificationPaymentByCheckoutSessionID(r.Context(), session.ID)
	if err != nil {
		logrus.WithError(err).WithField("session_id", session.ID).Error("failed to find verification payment record")
		http.Error(w, "Payment record not found", http.StatusNotFound)
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
		http.Error(w, "Failed to update payment status", http.StatusInternalServerError)
		return
	}

	// TODO: Trigger verification job/workflow here
	// This would typically enqueue a job to start the actual verification process
	// for the function at the purchased level

	logrus.WithFields(logrus.Fields{
		"payment_id":     payment.ID,
		"function_id":    functionIDStr,
		"tenant_id":      tenantIDStr,
		"level":          payment.VerificationLevel,
		"amount_cents":   payment.AmountCents,
		"payment_intent": stripePIID,
	}).Info("Function verification payment completed, verification can now proceed")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleSubscriptionUpdated(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	if h.sfAddons == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		http.Error(w, "Invalid subscription payload", http.StatusBadRequest)
		return
	}
	if sub.Metadata["purpose"] != "state_fabric_addon" {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}
	addonID := sub.Metadata["addon_id"]
	tenantIDStr := sub.Metadata["tenant_id"]
	tenantID, parseErr := uuid.Parse(tenantIDStr)
	if addonID == "" || parseErr != nil {
		http.Error(w, "Invalid subscription metadata", http.StatusBadRequest)
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
	if err := h.sfAddons.UpsertEntitlement(r.Context(), tenantID, addonID, status, &subID, itemID); err != nil {
		logrus.WithError(err).WithField("subscription_id", sub.ID).Error("state fabric addon: subscription updated entitlement")
		http.Error(w, "Failed to update entitlement", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleSubscriptionDeleted(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	if h.sfAddons == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}
	var sub stripe.Subscription
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &sub) != nil {
		http.Error(w, "Invalid subscription payload", http.StatusBadRequest)
		return
	}
	if sub.Metadata["purpose"] != "state_fabric_addon" {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}
	if err := h.sfAddons.SetEntitlementStatusBySubscription(r.Context(), sub.ID, "inactive"); err != nil {
		logrus.WithError(err).WithField("subscription_id", sub.ID).Error("state fabric addon: subscription deleted entitlement")
		http.Error(w, "Failed to update entitlement", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleAgentExecutionCreditsCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	tenantIDStr := session.Metadata["tenant_id"]
	agentID := session.Metadata["agent_id"]
	amountUsdStr := session.Metadata["amount_usd"]
	initiatingUserIDStr := session.Metadata["initiating_user_id"]

	if tenantIDStr == "" || agentID == "" {
		logrus.Warn("agent checkout session missing required metadata")
		http.Error(w, "Missing metadata", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Error("invalid tenant_id in metadata")
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
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
		http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
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
		http.Error(w, "Failed to add credits", http.StatusInternalServerError)
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
					"user_id":               initiatingUserID,
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
		"tenant_id":            tenantID,
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
		http.Error(w, "Wallet service unavailable", http.StatusInternalServerError)
		return
	}

	tenantIDStr := session.Metadata["tenant_id"]
	userIDStr := session.Metadata["user_id"]
	amountUsdStr := session.Metadata["amount_usd"]

	if tenantIDStr == "" || userIDStr == "" {
		logrus.Warn("registry wallet checkout missing tenant_id or user_id in metadata")
		http.Error(w, "Missing metadata", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		logrus.WithError(err).Warn("registry wallet webhook: invalid tenant_id")
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Warn("registry wallet webhook: invalid user_id")
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", userIDStr).Warn("registry wallet webhook: user not found")
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}
	if user.TenantID != tenantID {
		logrus.WithFields(logrus.Fields{"user_id": userID, "tenant_id": tenantID}).Warn("registry wallet webhook: tenant mismatch")
		http.Error(w, "Invalid checkout metadata", http.StatusBadRequest)
		return
	}

	amountUSD, _ := strconv.ParseFloat(amountUsdStr, 64)
	if amountUSD <= 0 {
		amountUSD = float64(session.AmountTotal) / 100
	}
	if amountUSD <= 0 {
		logrus.WithField("session_id", session.ID).Warn("registry wallet webhook: non-positive amount")
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	already, err := h.platformFees.HasWalletCreditReference(r.Context(), session.ID)
	if err != nil {
		logrus.WithError(err).Error("registry wallet webhook: idempotency check failed")
		http.Error(w, "Failed to verify payment", http.StatusInternalServerError)
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
			"user_id":    userID,
		}).Error("registry wallet webhook: credit failed")
		http.Error(w, "Failed to credit wallet", http.StatusInternalServerError)
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

		if err := h.userRepo.CreateUserActivity(activity); err != nil {
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
	achievement, err := h.userRepo.GetAchievementBySlug("enterprise_pioneer")
	if err != nil {
		return fmt.Errorf("failed to get enterprise pioneer achievement: %w", err)
	}

	if achievement == nil {
		logrus.Debug("Enterprise Pioneer achievement not found - skipping award")
		return nil
	}

	// Check if user already has this achievement
	existingAchievements, err := h.userRepo.GetUserAchievements(userID)
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
		http.Error(w, "Invalid invoice payload", http.StatusBadRequest)
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

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// handleInvoicePaymentSucceeded processes successful invoice payments and credits user wallet.
func (h *StripeWebhookHandler) handleInvoicePaymentSucceeded(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	var invoice stripe.Invoice
	raw, err := event.Data.Raw.MarshalJSON()
	if err != nil || json.Unmarshal(raw, &invoice) != nil {
		logrus.WithError(err).Error("stripe webhook: failed to unmarshal invoice payment succeeded event")
		http.Error(w, "Invalid invoice payload", http.StatusBadRequest)
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
	tenant, err := h.userRepo.GetTenantByStripeCustomerID(invoice.Customer.ID)
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
	}

	// Record metrics
	monitoring.RecordStripeEventProcessed("invoice.payment_succeeded")

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
		http.Error(w, "Invalid payment intent payload", http.StatusBadRequest)
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
