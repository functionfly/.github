package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	stripeSub "github.com/stripe/stripe-go/v83/subscription"
)

// StripeWebhookHandlerV2 is an updated webhook handler that uses the unified wallet system.
// It maintains backward compatibility with the old system during migration.
type StripeWebhookHandlerV2 struct {
	// Legacy repositories (for backward compatibility)
	financialTxRepo *storage.FinancialTransactionRepository
	billingCtrl     *billing.Controller
	userRepo        storage.Repository
	platformFees    *storageregistry.PlatformFeeRepository
	sfAddons        *statefabricaddons.Repository

	// New unified wallet service
	walletService *wallet.Service

	notificationSvc *notification.Service
	webhookSecret   string

	// Feature flags for gradual migration
	useUnifiedWalletForUsers  bool
	useUnifiedWalletForAgents bool
}

// StripeWebhookHandlerV2Config contains configuration for the new handler
type StripeWebhookHandlerV2Config struct {
	FinancialTxRepo *storage.FinancialTransactionRepository
	BillingCtrl     *billing.Controller
	UserRepo        storage.Repository
	PlatformFees    *storageregistry.PlatformFeeRepository
	SFAddons        *statefabricaddons.Repository
	WalletService   *wallet.Service
	NotificationSvc *notification.Service
	WebhookSecret   string

	// Migration flags
	UseUnifiedWalletForUsers  bool
	UseUnifiedWalletForAgents bool
}

// NewStripeWebhookHandlerV2 creates a new Stripe webhook handler with unified wallet support.
func NewStripeWebhookHandlerV2(cfg StripeWebhookHandlerV2Config) *StripeWebhookHandlerV2 {
	return &StripeWebhookHandlerV2{
		financialTxRepo:           cfg.FinancialTxRepo,
		billingCtrl:               cfg.BillingCtrl,
		userRepo:                  cfg.UserRepo,
		platformFees:              cfg.PlatformFees,
		sfAddons:                  cfg.SFAddons,
		walletService:             cfg.WalletService,
		notificationSvc:           cfg.NotificationSvc,
		webhookSecret:             cfg.WebhookSecret,
		useUnifiedWalletForUsers:  cfg.UseUnifiedWalletForUsers,
		useUnifiedWalletForAgents: cfg.UseUnifiedWalletForAgents,
	}
}

// handleAgentExecutionCreditsCheckoutUnified handles agent credit top-ups using the unified wallet system.
func (h *StripeWebhookHandlerV2) handleAgentExecutionCreditsCheckoutUnified(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
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

	// Parse initiating user ID for notifications
	var initiatingUserID *uuid.UUID
	if initiatingUserIDStr != "" {
		if uid, err := uuid.Parse(initiatingUserIDStr); err == nil {
			initiatingUserID = &uid
		}
	}

	// Record transaction in legacy system for audit trail (if available)
	if h.financialTxRepo != nil {
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
			logrus.WithError(err).Error("failed to create transaction record in legacy system")
			// Don't fail - continue with unified wallet
		}

		if !created {
			logrus.WithField("session_id", session.ID).Info("duplicate webhook received, transaction already recorded")
			h.persistCheckoutInvoiceV2(r.Context(), tenantID, session, amountUSD)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "already processed"})
			return
		}
	}

	// Use unified wallet system
	if h.walletService != nil && h.useUnifiedWalletForAgents {
		update, err := h.walletService.CreditAgentWallet(r.Context(), agentID, amountUSD, session.ID, initiatingUserID)
		if err != nil {
			// Check if already credited
			if hasRef, _ := h.walletService.HasUserWalletCreditReference(r.Context(), session.ID); hasRef {
				logrus.WithField("session_id", session.ID).Info("duplicate webhook - credits already applied to unified wallet")
				h.persistCheckoutInvoiceV2(r.Context(), tenantID, session, amountUSD)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "already processed"})
				return
			}

			logrus.WithError(err).WithFields(logrus.Fields{
				"agent_id":   agentID,
				"session_id": session.ID,
			}).Error("failed to add credits to unified wallet")

			// Fall back to legacy system if unified fails
			if h.billingCtrl != nil {
				if err := h.billingCtrl.AddCredits(r.Context(), agentID, amountUSD); err != nil {
					logrus.WithError(err).Error("fallback to legacy billing also failed")
					http.Error(w, "Failed to add credits", http.StatusInternalServerError)
					return
				}
				logrus.Warn("Successfully fell back to legacy billing system")
			} else {
				http.Error(w, "Failed to add credits", http.StatusInternalServerError)
				return
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"agent_id":        agentID,
				"tenant_id":       tenantID,
				"amount_usd":      amountUSD,
				"session_id":      session.ID,
				"new_balance":     update.CurrentBalance,
				"transaction_id":  update.TransactionID,
			}).Info("credits added to unified wallet via Stripe checkout")

			// Send notification via unified wallet service
			if initiatingUserID != nil {
				h.notificationSvc.SendWalletTopUp(r.Context(), *initiatingUserID, agentID, amountUSD, update.CurrentBalance)
			}
		}
	} else {
		// Use legacy system
		if h.billingCtrl == nil {
			logrus.Error("billing controller not configured and unified wallet disabled")
			http.Error(w, "Wallet service unavailable", http.StatusInternalServerError)
			return
		}

		if err := h.billingCtrl.AddCredits(r.Context(), agentID, amountUSD); err != nil {
			logrus.WithError(err).Error("failed to add credits via legacy billing controller")
			http.Error(w, "Failed to add credits", http.StatusInternalServerError)
			return
		}

		controls, err := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
		if err != nil {
			logrus.WithError(err).WithField("agent_id", agentID).Warn("failed to load controls after credit")
		}

		// Send notification via legacy system
		if h.notificationSvc != nil && initiatingUserID != nil {
			balanceUSD := 0.0
			if controls != nil {
				balanceUSD = controls.CreditBalanceUSD
			}
			if err := h.notificationSvc.SendWalletTopUp(r.Context(), *initiatingUserID, agentID, amountUSD, balanceUSD); err != nil {
				logrus.WithError(err).Warn("failed to send wallet top-up notification")
			}
		}

		logrus.WithFields(logrus.Fields{
			"agent_id":   agentID,
			"tenant_id":  tenantID,
			"amount_usd": amountUSD,
			"session_id": session.ID,
		}).Info("credits added to agent via legacy billing controller")
	}

	h.persistCheckoutInvoiceV2(r.Context(), tenantID, session, amountUSD)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleRegistryWalletCreditCheckoutUnified handles registry wallet top-ups using the unified wallet system.
func (h *StripeWebhookHandlerV2) handleRegistryWalletCreditCheckoutUnified(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
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

	// Verify user exists and belongs to tenant
	if h.userRepo != nil {
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

	// Use unified wallet system if enabled
	if h.walletService != nil && h.useUnifiedWalletForUsers {
		// Check for duplicate using unified wallet system
		hasRef, err := h.walletService.HasUserWalletCreditReference(r.Context(), session.ID)
		if err != nil {
			logrus.WithError(err).Error("registry wallet webhook: idempotency check failed")
			http.Error(w, "Failed to verify payment", http.StatusInternalServerError)
			return
		}
		if hasRef {
			logrus.WithField("session_id", session.ID).Info("duplicate registry wallet webhook, credit already applied")
			h.persistCheckoutInvoiceV2(r.Context(), tenantID, session, amountUSD)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "already processed"})
			return
		}

		// Credit the unified wallet
		_, err = h.walletService.CreditUserWallet(r.Context(), userID, amountUSD, session.ID)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"session_id": session.ID,
				"user_id":    userID,
			}).Error("registry wallet webhook: failed to credit unified wallet")

			// Fall back to legacy system
			if h.platformFees != nil {
				if err := h.platformFees.CreditWallet(r.Context(), userID, amountUSD, session.ID); err != nil {
					logrus.WithError(err).Error("fallback to legacy platform fees also failed")
					http.Error(w, "Failed to credit wallet", http.StatusInternalServerError)
					return
				}
				logrus.Warn("Successfully fell back to legacy platform fees system")
			} else {
				http.Error(w, "Failed to credit wallet", http.StatusInternalServerError)
				return
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"user_id":    userID,
				"tenant_id":  tenantID,
				"amount_usd": amountUSD,
				"session_id": session.ID,
			}).Info("registry wallet topped up via unified wallet system")
		}
	} else {
		// Use legacy system
		if h.platformFees == nil {
			logrus.Error("registry wallet webhook: platform fee repository not configured and unified wallet disabled")
			http.Error(w, "Wallet service unavailable", http.StatusInternalServerError)
			return
		}

		// Check for duplicate using legacy system
		already, err := h.platformFees.HasWalletCreditReference(r.Context(), session.ID)
		if err != nil {
			logrus.WithError(err).Error("registry wallet webhook: idempotency check failed")
			http.Error(w, "Failed to verify payment", http.StatusInternalServerError)
			return
		}
		if already {
			logrus.WithField("session_id", session.ID).Info("duplicate registry wallet webhook, credit already applied")
			h.persistCheckoutInvoiceV2(r.Context(), tenantID, session, amountUSD)
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
		}).Info("registry wallet topped up via legacy system")
	}

	h.persistCheckoutInvoiceV2(r.Context(), tenantID, session, amountUSD)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleStateFabricAddonCheckout handles state fabric addon purchases (unchanged).
func (h *StripeWebhookHandlerV2) handleStateFabricAddonCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
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

// persistCheckoutInvoiceV2 persists checkout invoice information.
func (h *StripeWebhookHandlerV2) persistCheckoutInvoiceV2(ctx context.Context, tenantID uuid.UUID, session *stripe.CheckoutSession, amountUSD float64) {
	if h.userRepo == nil || session == nil {
		return
	}

	amountCents := int(session.AmountTotal)
	if amountCents <= 0 && amountUSD > 0 {
		amountCents = int(amountUSD * 100)
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

// EnableUnifiedWalletForUsers enables the unified wallet for user transactions.
func (h *StripeWebhookHandlerV2) EnableUnifiedWalletForUsers() {
	h.useUnifiedWalletForUsers = true
	logrus.Info("Unified wallet enabled for users (registry fees)")
}

// EnableUnifiedWalletForAgents enables the unified wallet for agent transactions.
func (h *StripeWebhookHandlerV2) EnableUnifiedWalletForAgents() {
	h.useUnifiedWalletForAgents = true
	logrus.Info("Unified wallet enabled for agents (execution credits)")
}

// DisableUnifiedWallet disables the unified wallet and falls back to legacy systems.
func (h *StripeWebhookHandlerV2) DisableUnifiedWallet() {
	h.useUnifiedWalletForUsers = false
	h.useUnifiedWalletForAgents = false
	logrus.Info("Unified wallet disabled - using legacy systems")
}
