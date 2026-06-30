package trustapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// BillingHandler handles Trust API partner billing HTTP endpoints
type BillingHandler struct {
	service *BillingService
	repo    *trustapi.BillingRepository
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(service *BillingService, repo *trustapi.BillingRepository) *BillingHandler {
	return &BillingHandler{
		service: service,
		repo:    repo,
	}
}

// resolvePartner resolves the partner for the current request.
// It first checks API key context, then falls back to JWT user context.
func (h *BillingHandler) resolvePartner(r *http.Request) *trustapi.TrustAPIPartner {
	// First try API key context
	if partner := getPartnerFromContext(r); partner != nil {
		return partner
	}

	// Fall back to JWT user context - look up partner by user email
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.Email == "" {
		return nil
	}

	partner, err := h.repo.GetPartnerByContactEmail(r.Context(), claims.Email)
	if err != nil {
		return nil
	}
	return partner
}

// ============================================
// Tier Pricing Endpoints
// ============================================

// HandleGetTierPricing handles GET /v1/partners/tiers
// Returns pricing information for all partner tiers
func (h *BillingHandler) HandleGetTierPricing(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.service.ListTierPricing(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list tier pricing")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve pricing tiers"))
		return
	}

	response := make([]TierPricingResponse, 0, len(tiers))
	for _, t := range tiers {
		response = append(response, TierPricingResponse{
			Tier:                t.Tier,
			MonthlyPriceUSD:     float64(t.MonthlyPriceCents) / 100.0,
			IncludedRequests:    t.IncludedRequests,
			OveragePricePer1000: float64(t.OveragePricePer1000) / 100.0,
			HasOverageBilling:   t.HasOverageBilling,
			RateLimitPerMinute:  t.RateLimitPerMinute,
			RateLimitPerDay:     t.RateLimitPerDay,
			MonthlyRequestLimit: t.MonthlyRequestLimit,
			Description:         t.Description,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": response,
	})
}

// TierPricingResponse represents pricing tier information
type TierPricingResponse struct {
	Tier                string  `json:"tier"`
	MonthlyPriceUSD     float64 `json:"monthly_price_usd"`
	IncludedRequests    int     `json:"included_requests"`
	OveragePricePer1000 float64 `json:"overage_price_per_1000,omitempty"`
	HasOverageBilling   bool    `json:"has_overage_billing"`
	RateLimitPerMinute  int     `json:"rate_limit_per_minute"`
	RateLimitPerDay     int     `json:"rate_limit_per_day"`
	MonthlyRequestLimit int     `json:"monthly_request_limit"`
	Description         string  `json:"description"`
}

// ============================================
// Partner Billing Endpoints
// ============================================

// HandleGetBillingStatus handles GET /v1/partners/{partner_id}/billing
// Returns current billing status for a partner
func (h *BillingHandler) HandleGetBillingStatus(w http.ResponseWriter, r *http.Request) {
	partner := h.resolvePartner(r)
	if partner == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	usage, err := h.service.GetCurrentUsage(r.Context(), partner.ID)
	if err != nil {
		logrus.WithError(err).WithField("partner_id", partner.ID).Error("Failed to get billing usage")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve billing status"))
		return
	}

	response := BillingStatusResponse{
		PartnerID:          partner.ID,
		Tier:               usage.Tier,
		BillingStatus:      partner.BillingStatus,
		IsFounderMode:      partner.IsFounderMode,
		MonthlyPriceUSD:    float64(usage.MonthlyPriceCents) / 100.0,
		IncludedRequests:   usage.IncludedRequests,
		CurrentUsage:       usage.CurrentUsage,
		RemainingRequests:  usage.RemainingRequests,
		OverageRequests:    usage.OverageRequests,
		OverageChargeUSD:   float64(usage.OverageChargeCents) / 100.0,
		BillingPeriodStart: usage.BillingPeriodStart,
		BillingPeriodEnd:   usage.BillingPeriodEnd,
		IsHardLimit:        usage.IsHardLimit,
	}

	// Add founder mode info if applicable
	if partner.IsFounderMode {
		response.FounderModeStartedAt = partner.FounderModeStartedAt
		response.FounderModeEndsAt = partner.FounderModeEndsAt
		response.UsageThreshold = partner.UsageThreshold
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// BillingStatusResponse represents a partner's billing status
type BillingStatusResponse struct {
	PartnerID            uuid.UUID  `json:"partner_id"`
	Tier                 string     `json:"tier"`
	BillingStatus        string     `json:"billing_status"`
	IsFounderMode        bool       `json:"is_founder_mode"`
	MonthlyPriceUSD      float64    `json:"monthly_price_usd"`
	IncludedRequests     int        `json:"included_requests"`
	CurrentUsage         int        `json:"current_usage"`
	RemainingRequests    int        `json:"remaining_requests"`
	OverageRequests      int        `json:"overage_requests"`
	OverageChargeUSD     float64    `json:"overage_charge_usd"`
	BillingPeriodStart   time.Time  `json:"billing_period_start"`
	BillingPeriodEnd     time.Time  `json:"billing_period_end"`
	IsHardLimit          bool       `json:"is_hard_limit"`
	FounderModeStartedAt *time.Time `json:"founder_mode_started_at,omitempty"`
	FounderModeEndsAt    *time.Time `json:"founder_mode_ends_at,omitempty"`
	UsageThreshold       int        `json:"usage_threshold,omitempty"`
}

// HandleCreateCheckout handles POST /v1/partners/{partner_id}/billing/checkout
// Creates a Stripe checkout session for tier upgrade
func (h *BillingHandler) HandleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	partner := h.resolvePartner(r)
	if partner == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Tier == "" {
		apierror.WriteError(w, apierror.NewValidation("Tier is required"))
		return
	}

	// Validate tier
	if req.Tier == "developer" {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, "INVALID_TIER", "Cannot checkout for free tier")
		return
	}

	result, err := h.service.CreateCheckoutSession(r.Context(), partner.ID, req.Tier, req.SuccessURL, req.CancelURL)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"partner_id": partner.ID,
			"tier":       req.Tier,
		}).Error("Failed to create checkout session")
		apierror.WriteError(w, apierror.NewBillingError("Failed to create checkout session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CheckoutResponse{
		SessionID: result.SessionID,
		URL:       result.URL,
		Status:    result.Status,
	})
}

// CheckoutRequest represents a request to create a checkout session
type CheckoutRequest struct {
	Tier       string `json:"tier"`
	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
}

// CheckoutResponse represents a checkout session response
type CheckoutResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	Status    string `json:"status"`
}

// HandleEnrollFounderMode handles POST /v1/partners/{partner_id}/founder
// Enrolls a partner in founder mode (free tier with limits)
func (h *BillingHandler) HandleEnrollFounderMode(w http.ResponseWriter, r *http.Request) {
	partner := h.resolvePartner(r)
	if partner == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req FounderModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Set defaults
	if req.UsageThreshold <= 0 {
		req.UsageThreshold = 100000 // 100K requests default
	}
	if req.FreeDays <= 0 {
		req.FreeDays = 90 // 90 days default
	}

	if err := h.service.EnrollFounderMode(r.Context(), partner.ID, req.UsageThreshold, req.FreeDays); err != nil {
		logrus.WithError(err).WithField("partner_id", partner.ID).Error("Failed to enroll in founder mode")
		if err.Error() == "partner is already in founder mode" {
			apierror.WriteError(w, apierror.NewConflict("Partner is already enrolled in founder mode"))
			return
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to enroll in founder mode"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"message":         "Enrolled in founder mode",
		"usage_threshold": req.UsageThreshold,
		"free_days":       req.FreeDays,
	})
}

// FounderModeRequest represents a request to enroll in founder mode
type FounderModeRequest struct {
	UsageThreshold int `json:"usage_threshold,omitempty"`
	FreeDays       int `json:"free_days,omitempty"`
}

// HandleGetUsageReport handles GET /v1/partners/{partner_id}/billing/usage
// Returns detailed usage report for the current billing period
func (h *BillingHandler) HandleGetUsageReport(w http.ResponseWriter, r *http.Request) {
	partner := h.resolvePartner(r)
	if partner == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	report, err := h.repo.GetUsageSummaryForPartner(r.Context(), partner.ID)
	if err != nil {
		logrus.WithError(err).WithField("partner_id", partner.ID).Error("Failed to get usage report")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve usage report"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

// HandleGetInvoices handles GET /v1/partners/{partner_id}/billing/invoices
// Returns billing history for a partner
func (h *BillingHandler) HandleGetInvoices(w http.ResponseWriter, r *http.Request) {
	partner := h.resolvePartner(r)
	if partner == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Parse pagination
	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	records, total, err := h.repo.ListBillingRecordsByPartner(r.Context(), partner.ID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("partner_id", partner.ID).Error("Failed to list invoices")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve invoices"))
		return
	}

	invoices := make([]InvoiceResponse, 0, len(records))
	for _, r := range records {
		invoices = append(invoices, InvoiceResponse{
			ID:               r.ID,
			PeriodStart:      r.PeriodStart,
			PeriodEnd:        r.PeriodEnd,
			BaseRequests:     r.BaseRequests,
			OverageRequests:  r.OverageRequests,
			TotalRequests:    r.TotalRequests,
			BaseChargeUSD:    float64(r.BaseChargeCents) / 100.0,
			OverageChargeUSD: float64(r.OverageChargeCents) / 100.0,
			TotalChargeUSD:   float64(r.TotalChargeCents) / 100.0,
			Status:           r.Status,
			StripeInvoiceID:  r.StripeInvoiceID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invoices": invoices,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// InvoiceResponse represents a billing invoice
type InvoiceResponse struct {
	ID               uuid.UUID `json:"id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	BaseRequests     int       `json:"base_requests"`
	OverageRequests  int       `json:"overage_requests"`
	TotalRequests    int       `json:"total_requests"`
	BaseChargeUSD    float64   `json:"base_charge_usd"`
	OverageChargeUSD float64   `json:"overage_charge_usd"`
	TotalChargeUSD   float64   `json:"total_charge_usd"`
	Status           string    `json:"status"`
	StripeInvoiceID  string    `json:"stripe_invoice_id,omitempty"`
}

// ============================================
// Admin Endpoints
// ============================================

// HandleListPartnerBilling handles GET /v1/admin/partners/billing
// Admin endpoint to list all partners with billing info
func (h *BillingHandler) HandleListPartnerBilling(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	tier := r.URL.Query().Get("tier")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	partners, total, err := h.repo.ListPartnersWithFilters(status, tier, pageSize, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list partners")
		apierror.WriteError(w, apierror.NewInternal("Failed to list partners"))
		return
	}

	response := make([]PartnerBillingSummary, 0, len(partners))
	for _, p := range partners {
		usage, _ := h.service.GetCurrentUsage(r.Context(), p.ID)
		var currentUsage, overageUsage int
		if usage != nil {
			currentUsage = usage.CurrentUsage
			overageUsage = usage.OverageRequests
		}

		response = append(response, PartnerBillingSummary{
			ID:                   p.ID,
			Name:                 p.Name,
			Slug:                 p.Slug,
			Tier:                 p.Tier,
			BillingStatus:        p.BillingStatus,
			IsFounderMode:        p.IsFounderMode,
			CurrentUsage:         currentUsage,
			OverageUsage:         overageUsage,
			StripeCustomerID:     p.StripeCustomerID,
			StripeSubscriptionID: p.StripeSubscriptionID,
			CreatedAt:            p.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"partners":  response,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// PartnerBillingSummary represents billing summary for a partner
type PartnerBillingSummary struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	Slug                 string    `json:"slug"`
	Tier                 string    `json:"tier"`
	BillingStatus        string    `json:"billing_status"`
	IsFounderMode        bool      `json:"is_founder_mode"`
	CurrentUsage         int       `json:"current_usage"`
	OverageUsage         int       `json:"overage_usage"`
	StripeCustomerID     string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}


