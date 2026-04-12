package trustapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	storagetrustapi "github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
)

// StripeWebhookHandler handles Stripe webhook events for Trust API billing
type StripeWebhookHandler struct {
	service *BillingService
	repo    *storagetrustapi.BillingRepository
}

// NewStripeWebhookHandler creates a new Stripe webhook handler
func NewStripeWebhookHandler(service *BillingService, repo *storagetrustapi.BillingRepository) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		service: service,
		repo:    repo,
	}
}

// HandleStripeWebhook handles POST /v1/webhooks/stripe
// Processes Stripe webhook events for billing
func (h *StripeWebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read webhook body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Get webhook secret from environment
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		logrus.Warn("STRIPE_WEBHOOK_SECRET not set, skipping signature verification")
		// In production, you should always verify signatures
		// For now, we'll process without verification if secret is not set
		// (only in development)
	}

	// Verify and construct event
	var event stripe.Event
	if webhookSecret != "" {
		event, err = webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
		if err != nil {
			logrus.WithError(err).Error("Failed to verify webhook signature")
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}
	} else {
		// Parse without verification (dev only)
		if err := json.Unmarshal(payload, &event); err != nil {
			logrus.WithError(err).Error("Failed to parse webhook event")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	// Process the event
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(w, r, event)
	case "invoice.paid":
		h.handleInvoicePaid(w, r, event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(w, r, event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(w, r, event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(w, r, event)
	default:
		// Acknowledge unhandled events
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "acknowledged",
			"type":   "unhandled",
		})
	}
}

// handleCheckoutSessionCompleted processes completed checkout sessions
func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		logrus.WithError(err).Error("Failed to parse checkout session")
		http.Error(w, "Failed to parse session", http.StatusBadRequest)
		return
	}

	// Check if this is a Trust API subscription
	if session.Metadata["type"] != "trust_api_subscription" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process the successful checkout
	if err := h.service.HandleCheckoutSuccess(r.Context(), session.ID); err != nil {
		logrus.WithError(err).WithField("session_id", session.ID).Error("Failed to process checkout success")
		http.Error(w, "Failed to process checkout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "checkout.session.completed",
	})
}

// handleInvoicePaid processes paid invoices
func (h *StripeWebhookHandler) handleInvoicePaid(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		logrus.WithError(err).Error("Failed to parse invoice")
		http.Error(w, "Failed to parse invoice", http.StatusBadRequest)
		return
	}

	// Update billing records for this invoice
	// This would update the status from "draft" to "paid"
	ctx := context.Background()
	records, err := h.repo.GetBillingRecordsByStripeInvoice(ctx, invoice.ID)
	if err != nil {
		logrus.WithError(err).WithField("invoice_id", invoice.ID).Error("Failed to find billing records")
		http.Error(w, "Failed to process invoice", http.StatusInternalServerError)
		return
	}

	for _, record := range records {
		if err := h.repo.UpdateBillingRecordStatus(ctx, record.ID, "paid"); err != nil {
			logrus.WithError(err).WithField("record_id", record.ID).Warn("Failed to update billing record status")
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "invoice.paid",
	})
}

// handleInvoicePaymentFailed processes failed invoice payments
func (h *StripeWebhookHandler) handleInvoicePaymentFailed(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		logrus.WithError(err).Error("Failed to parse invoice")
		http.Error(w, "Failed to parse invoice", http.StatusBadRequest)
		return
	}

	// Mark billing records as failed
	ctx := context.Background()
	records, err := h.repo.GetBillingRecordsByStripeInvoice(ctx, invoice.ID)
	if err != nil {
		logrus.WithError(err).WithField("invoice_id", invoice.ID).Error("Failed to find billing records")
		http.Error(w, "Failed to process invoice", http.StatusInternalServerError)
		return
	}

	for _, record := range records {
		if err := h.repo.UpdateBillingRecordStatus(ctx, record.ID, "payment_failed"); err != nil {
			logrus.WithError(err).WithField("record_id", record.ID).Warn("Failed to update billing record status")
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "invoice.payment_failed",
	})
}

// handleSubscriptionUpdated processes subscription updates
func (h *StripeWebhookHandler) handleSubscriptionUpdated(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		logrus.WithError(err).Error("Failed to parse subscription")
		http.Error(w, "Failed to parse subscription", http.StatusBadRequest)
		return
	}

	// Update partner subscription info
	if partnerID := sub.Metadata["partner_id"]; partnerID != "" {
		// Could update subscription status here if needed
		logrus.WithFields(logrus.Fields{
			"partner_id":      partnerID,
			"subscription_id": sub.ID,
			"status":          sub.Status,
		}).Info("Subscription updated")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "customer.subscription.updated",
	})
}

// handleSubscriptionDeleted processes subscription cancellations
func (h *StripeWebhookHandler) handleSubscriptionDeleted(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		logrus.WithError(err).Error("Failed to parse subscription")
		http.Error(w, "Failed to parse subscription", http.StatusBadRequest)
		return
	}

	// Update partner subscription status
	if partnerID := sub.Metadata["partner_id"]; partnerID != "" {
		ctx := context.Background()
		// Find partner by subscription ID
		partners, err := h.repo.ListPartnersWithActiveBilling(ctx)
		if err == nil {
			for _, p := range partners {
				if p.StripeSubscriptionID == sub.ID {
					p.BillingStatus = "cancelled"
					p.StripeSubscriptionID = ""
					h.repo.UpdatePartner(&p)
					break
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "customer.subscription.deleted",
	})
}
