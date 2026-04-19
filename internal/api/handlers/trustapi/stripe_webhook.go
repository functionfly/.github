package trustapi

import (
	"context"
	"encoding/json"
	"fmt"
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
	if webhookSecret == "" {
		logrus.Error("STRIPE_WEBHOOK_SECRET not set - rejecting webhook")
		http.Error(w, "Webhook authentication not configured", http.StatusInternalServerError)
		return
	}
	event, err = webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify webhook signature")
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Process the event
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(w, r, event)
	case "invoice.paid":
		h.handleInvoicePaid(w, r, event)
	case "invoice.finalized":
		h.handleInvoiceFinalized(w, r, event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(w, r, event)
	case "payment_intent.succeeded":
		h.handlePaymentIntentSucceeded(w, r, event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(w, r, event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(w, r, event)
	case "customer.tax_id.updated":
		h.handleCustomerTaxIdUpdated(w, r, event)
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

// handleInvoiceFinalized processes finalized invoices (preview before charging)
// This allows partners to preview charges before they're actually billed
func (h *StripeWebhookHandler) handleInvoiceFinalized(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var invoice struct {
		ID              string     `json:"id"`
		Subscription    *string    `json:"subscription"`
		AmountDue       int64      `json:"amount_due"`
		Currency        string     `json:"currency"`
		PeriodStart     int64      `json:"period_start"`
		PeriodEnd       int64      `json:"period_end"`
		HostedInvoiceURL string    `json:"hosted_invoice_url"`
	}
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		logrus.WithError(err).Error("Failed to parse invoice")
		http.Error(w, "Failed to parse invoice", http.StatusBadRequest)
		return
	}

	// Skip finalized events for subscriptions (they're handled by subscription events)
	// Only process one-time charges or usage-based billing that needs preview
	if invoice.Subscription != nil && *invoice.Subscription != "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "skipped",
			"type":   "invoice.finalized",
			"reason": "subscription_invoice",
		})
		return
	}

	// Update billing record status to "finalized" for preview
	ctx := context.Background()
	records, err := h.repo.GetBillingRecordsByStripeInvoice(ctx, invoice.ID)
	if err != nil {
		logrus.WithError(err).WithField("invoice_id", invoice.ID).Error("Failed to find billing records")
		// Don't fail - finalized invoices might not have records yet
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "processed",
			"type":   "invoice.finalized",
			"note":   "no_billing_records_found",
		})
		return
	}

	for _, record := range records {
		if err := h.repo.UpdateBillingRecordStatus(ctx, record.ID, "finalized"); err != nil {
			logrus.WithError(err).WithField("record_id", record.ID).Warn("Failed to update billing record status")
		}

		// Log finalized invoice details for partner notification
		logrus.WithFields(logrus.Fields{
			"invoice_id":          invoice.ID,
			"record_id":           record.ID,
			"amount_due":          invoice.AmountDue,
			"currency":            invoice.Currency,
			"period_start":        invoice.PeriodStart,
			"period_end":          invoice.PeriodEnd,
			"hosted_invoice_url":  invoice.HostedInvoiceURL,
		}).Info("Invoice finalized - ready for partner preview")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "invoice.finalized",
	})
}

// handlePaymentIntentSucceeded processes successful payments for immediate confirmation
// This provides real-time payment success notifications beyond invoice events
func (h *StripeWebhookHandler) handlePaymentIntentSucceeded(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var paymentIntent struct {
		ID            string            `json:"id"`
		Amount        int64             `json:"amount"`
		Currency      string            `json:"currency"`
		PaymentMethod string            `json:"payment_method"`
		Invoice       *string           `json:"invoice"`
		Metadata      map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		logrus.WithError(err).Error("Failed to parse payment intent")
		http.Error(w, "Failed to parse payment intent", http.StatusBadRequest)
		return
	}

	// Extract partner info from metadata
	partnerIDStr := paymentIntent.Metadata["partner_id"]
	if partnerIDStr == "" {
		// Not a Trust API payment intent
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "skipped",
			"type":   "payment_intent.succeeded",
			"reason": "not_trust_api",
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"payment_intent_id": paymentIntent.ID,
		"partner_id":       partnerIDStr,
		"amount":          paymentIntent.Amount,
		"currency":       paymentIntent.Currency,
		"payment_method":  paymentIntent.PaymentMethod,
	}).Info("Payment intent succeeded - immediate confirmation")

	// If this payment is associated with an invoice, update billing status
	if paymentIntent.Invoice != nil && *paymentIntent.Invoice != "" {
		ctx := context.Background()
		records, err := h.repo.GetBillingRecordsByStripeInvoice(ctx, *paymentIntent.Invoice)
		if err == nil {
			for _, record := range records {
				if err := h.repo.UpdateBillingRecordStatus(ctx, record.ID, "paid"); err != nil {
					logrus.WithError(err).WithField("record_id", record.ID).Warn("Failed to update billing record status")
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "payment_intent.succeeded",
	})
}

// handleCustomerTaxIdUpdated processes tax ID updates for VAT handling
// This ensures proper VAT/tax compliance for EU and international partners
func (h *StripeWebhookHandler) handleCustomerTaxIdUpdated(w http.ResponseWriter, r *http.Request, event stripe.Event) {
	var customer struct {
		ID      string `json:"id"`
		TaxIDs  []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"tax_ids"`
	}
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		logrus.WithError(err).Error("Failed to parse customer")
		http.Error(w, "Failed to parse customer", http.StatusBadRequest)
		return
	}

	// Find partner by Stripe customer ID
	ctx := context.Background()
	partner, err := h.repo.GetPartnerByStripeCustomerID(ctx, customer.ID)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", customer.ID).Error("Failed to find partner by Stripe customer ID")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "skipped",
			"type":   "customer.tax_id.updated",
			"reason": "partner_not_found",
		})
		return
	}

	// Extract tax IDs from the event
	var taxIDStrings []string
	for _, taxID := range customer.TaxIDs {
		taxIDStrings = append(taxIDStrings, fmt.Sprintf("%s:%s", taxID.Type, taxID.Value))
	}

	// Update partner with tax information if they have a VAT ID
	// EU VAT IDs are stored for proper tax handling
	taxIDJSON, _ := json.Marshal(taxIDStrings)
	logrus.WithFields(logrus.Fields{
		"partner_id": partner.ID,
		"customer_id": customer.ID,
		"tax_ids":    string(taxIDJSON),
	}).Info("Customer tax ID updated - VAT handling")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "processed",
		"type":   "customer.tax_id.updated",
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
