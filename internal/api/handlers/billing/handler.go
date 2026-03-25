package billing

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/statefabricaddons"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PaymentMethodInfo represents the payment method details for display
type PaymentMethodInfo struct {
	Brand    string `json:"brand"`
	Last4    string `json:"last4"`
	ExpMonth int    `json:"exp_month"`
	ExpYear  int    `json:"exp_year"`
}

// SubscriptionResponse is the response for subscription details
type SubscriptionResponse struct {
	ID                   uuid.UUID          `json:"id"`
	TenantID             uuid.UUID          `json:"tenant_id"`
	Plan                 string             `json:"plan"`
	Status               string             `json:"status"`
	StripeSubscriptionID string             `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   *time.Time         `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time         `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	PaymentMethod        *PaymentMethodInfo `json:"payment_method,omitempty"`
}

// Handler handles billing portal and subscription management (Stripe).
type Handler struct {
	repo storage.Repository
	// Platform-fee wallet (registry credits balance for publish fees, etc.)
	platformFees *storageregistry.PlatformFeeRepository
	// State Fabric add-on entitlements (optional; nil returns empty entitlements).
	sfAddons *statefabricaddons.Repository
}

// NewHandler creates a new billing handler.
func NewHandler(repo storage.Repository, platformFees *storageregistry.PlatformFeeRepository, sfAddons *statefabricaddons.Repository) *Handler {
	return &Handler{repo: repo, platformFees: platformFees, sfAddons: sfAddons}
}

// CreatePortalSessionRequest is the request body for creating a billing portal session.
type CreatePortalSessionRequest struct {
	ReturnURL string `json:"return_url"`
}

// CreateCheckoutSessionRequest is the request body for creating a checkout session.
type CreateCheckoutSessionRequest struct {
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// CreatePortalSessionResponse is the response with the Stripe portal URL.
type CreatePortalSessionResponse struct {
	URL string `json:"url"`
}

// CreateCheckoutSessionResponse is the response with the Stripe checkout URL.
type CreateCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// HandleCreatePortalSession creates a Stripe Customer Billing Portal session and returns the URL.
// POST /v1/billing/portal-session
func (h *Handler) HandleCreatePortalSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing portal: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	var req CreatePortalSessionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = "/settings"
	}
	// Stripe requires a full URL; build from request if path-only
	if strings.HasPrefix(returnURL, "/") {
		scheme := "https"
		if r.TLS == nil && (r.URL == nil || r.URL.Scheme == "") {
			scheme = "http"
		}
		if r.URL != nil && r.URL.Scheme != "" {
			scheme = r.URL.Scheme
		}
		returnURL = scheme + "://" + r.Host + returnURL
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	customerID, err := payment.CreateOrGetStripeCustomer(r.Context(), h.repo, claims.TenantID, user.Email, name)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing portal: create or get stripe customer")
		msg := "Failed to prepare billing session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	url, err := payment.CreateBillingPortalSession(r.Context(), customerID, returnURL)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", customerID).Error("billing portal: create session")
		msg := "Failed to create billing session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CreatePortalSessionResponse{URL: url})
}

// HandleCreateCheckoutSession creates a Stripe Checkout session for subscription checkout.
// POST /v1/billing/checkout
func (h *Handler) HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing checkout: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	var req CreateCheckoutSessionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.PriceID == "" {
		writeJSONError(w, http.StatusBadRequest, "price_id is required")
		return
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	resp, err := payment.CreateCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateCheckoutSessionRequest{
			PriceID:    req.PriceID,
			SuccessURL: req.SuccessURL,
			CancelURL:  req.CancelURL,
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing checkout: create checkout session")
		msg := "Failed to create checkout session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// HandleGetSubscription returns the current user's subscription details.
// GET /v1/billing/subscription
func (h *Handler) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	subscription, err := h.repo.GetSubscriptionByTenantID(claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get subscription")
		writeJSONError(w, http.StatusNotFound, "No subscription found")
		return
	}
	if subscription == nil {
		writeJSONError(w, http.StatusNotFound, "No subscription found")
		return
	}

	// Get tenant to access Stripe customer ID for payment method info
	tenant, err := h.repo.GetTenantByID(claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get tenant")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve subscription details")
		return
	}

	// Build extended subscription response with payment method info
	var paymentMethod *PaymentMethodInfo
	if tenant.StripeCustomerID != nil && *tenant.StripeCustomerID != "" {
		pm, err := payment.GetPaymentMethodForCustomer(r.Context(), *tenant.StripeCustomerID)
		if err != nil {
			logrus.WithError(err).Warn("billing: failed to get payment method")
			// Don't fail the request, just don't include payment method
		} else if pm != nil {
			paymentMethod = &PaymentMethodInfo{
				Brand:    pm.Brand,
				Last4:    pm.Last4,
				ExpMonth: pm.ExpMonth,
				ExpYear:  pm.ExpYear,
			}
		}
	}

	// Convert storage.Subscription to response format (PricingTier may be nil if tier was deleted or not loaded)
	plan := ""
	if subscription.PricingTier != nil {
		plan = subscription.PricingTier.Name
	}
	response := SubscriptionResponse{
		ID:                   subscription.ID,
		TenantID:             subscription.TenantID,
		Plan:                 plan,
		Status:               subscription.Status,
		StripeSubscriptionID: subscription.ID.String(), // Use subscription ID as reference
		CurrentPeriodStart:   &subscription.CurrentPeriodStart,
		CurrentPeriodEnd:     &subscription.CurrentPeriodEnd,
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
		CanceledAt:           subscription.CanceledAt,
		CreatedAt:            subscription.CreatedAt,
		UpdatedAt:            subscription.UpdatedAt,
		PaymentMethod:        paymentMethod,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleListInvoices returns the current user's invoices.
// GET /v1/billing/invoices
func (h *Handler) HandleListInvoices(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	invoices, err := h.repo.ListInvoicesByTenant(claims.TenantID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to list invoices")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve invoices")
		return
	}

	// Convert []*storage.Invoice to []storage.Invoice
	invoiceList := make([]storage.Invoice, len(invoices))
	for i, invoice := range invoices {
		invoiceList[i] = *invoice
	}

	response := struct {
		Invoices []storage.Invoice `json:"invoices"`
		Limit    int               `json:"limit"`
		Offset   int               `json:"offset"`
	}{
		Invoices: invoiceList,
		Limit:    limit,
		Offset:   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetUsage returns the current user's usage details.
// GET /v1/billing/usage
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get usage from the past 30 days by default
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -30)

	if s := r.URL.Query().Get("start"); s != "" {
		if parsed, err := time.Parse("2006-01-02", s); err == nil {
			start = parsed
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if parsed, err := time.Parse("2006-01-02", e); err == nil {
			end = parsed
		}
	}

	usage, err := h.repo.GetUsageByTenant(claims.TenantID, "", start, end)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get usage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve usage")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"usage": usage,
		"start": start.Format("2006-01-02"),
		"end":   end.Format("2006-01-02"),
	})
}

// CancelSubscriptionRequest is the request body for cancelling a subscription
type CancelSubscriptionRequest struct {
	Immediately bool `json:"immediately"`
}

// HandleCancelSubscription cancels the current user's subscription.
// POST /v1/billing/subscription/cancel
func (h *Handler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	// Get subscription to find Stripe subscription ID
	subscription, err := h.repo.GetSubscriptionByTenantID(claims.TenantID)
	if err != nil || subscription == nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing cancel: no subscription found")
		writeJSONError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	// Parse request body
	var req CancelSubscriptionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	// Cancel the subscription via the repository
	err = h.repo.CancelSubscription(r.Context(), subscription.ID)
	if err != nil {
		logrus.WithError(err).WithField("subscription_id", subscription.ID).Error("billing cancel: failed to cancel subscription")
		msg := "Failed to cancel subscription"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subscription cancelled successfully",
	})
}
