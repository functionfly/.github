package webhooks

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/functionfly/functionfly/internal/agent/billing"
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

	// Verify webhook signature if secret is configured
	if h.webhookSecret != "" {
		signature := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, signature, h.webhookSecret)
		if err != nil {
			logrus.WithError(err).Warn("invalid stripe webhook signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
		h.handleEvent(w, r, &event)
		return
	}

	// No secret configured - parse without verification (development mode only)
	var event stripe.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		logrus.WithError(err).Warn("failed to parse stripe webhook event")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	h.handleEvent(w, r, &event)
}

func (h *StripeWebhookHandler) handleEvent(w http.ResponseWriter, r *http.Request, event *stripe.Event) {
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(w, r, event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(w, r, event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(w, r, event)
	default:
		// Acknowledge but ignore other events
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
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
				logrus.WithError(err).WithFields(logrus.Fields{
					"session_id": session.ID,
					"user_id":    initiatingUserID,
					"agent_id":   agentID,
				}).Warn("failed to send wallet top-up notification after checkout")
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"agent_id":   agentID,
		"tenant_id":  tenantID,
		"amount_usd": amountUSD,
		"session_id": session.ID,
	}).Info("credits purchased via checkout")

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

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"amount_usd": amountUSD,
		"session_id": session.ID,
	}).Info("registry wallet topped up via Stripe checkout")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
