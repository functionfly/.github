package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Billing handler functions

// HandleBillingSummary returns summary stats for the admin billing page.
// GET /v1/admin/billing/summary
func (h *Handler) HandleBillingSummary(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.repo.ListAllInvoices(10000, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list invoices for billing summary")
		http.Error(w, "Failed to get billing summary", http.StatusInternalServerError)
		return
	}

	subs, err := h.repo.ListAllSubscriptions(10000, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list subscriptions for billing summary")
		http.Error(w, "Failed to get billing summary", http.StatusInternalServerError)
		return
	}

	var totalRevenueCents int
	var pendingInvoices int
	var overdueCents int
	now := time.Now()

	for _, inv := range invoices {
		totalRevenueCents += inv.AmountPaidCents
		switch strings.ToLower(inv.Status) {
		case "pending", "open":
			pendingInvoices++
			if inv.DueDate != nil && inv.DueDate.Before(now) {
				overdueCents += inv.AmountDueCents - inv.AmountPaidCents
			}
		case "overdue":
			pendingInvoices++
			overdueCents += inv.AmountDueCents - inv.AmountPaidCents
		}
	}

	activeSubscriptions := 0
	for _, sub := range subs {
		if strings.ToLower(sub.Status) == "active" {
			activeSubscriptions++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": map[string]interface{}{
			"totalRevenue":        totalRevenueCents / 100,
			"activeSubscriptions": activeSubscriptions,
			"pendingInvoices":     pendingInvoices,
			"overdue":             overdueCents / 100,
		},
		"success": true,
	})
}

// HandleListPricingTiers lists all pricing tiers
func (h *Handler) HandleListPricingTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.repo.ListPricingTiers()
	if err != nil {
		logrus.WithError(err).Error("Failed to list pricing tiers")
		http.Error(w, "Failed to list pricing tiers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
	})
}

// HandleCreatePricingTier creates a new pricing tier
func (h *Handler) HandleCreatePricingTier(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		PriceCents  int         `json:"price_cents"`
		Currency    string      `json:"currency"`
		Features    interface{} `json:"features"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tier := &storage.PricingTier{
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Features:    req.Features,
		IsActive:    true,
	}

	createdTier, err := h.repo.CreatePricingTier(r.Context(), tier)
	if err != nil {
		logrus.WithError(err).Error("Failed to create pricing tier")
		http.Error(w, "Failed to create pricing tier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTier)
}

// HandleGetPricingTier gets a specific pricing tier
func (h *Handler) HandleGetPricingTier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tierIDStr := vars["tierId"]
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	tier, err := h.repo.GetPricingTierByID(tierID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get pricing tier")
		http.Error(w, "Failed to get pricing tier", http.StatusInternalServerError)
		return
	}
	if tier == nil {
		http.Error(w, "Pricing tier not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tier)
}

// HandleUpdatePricingTier updates a pricing tier
func (h *Handler) HandleUpdatePricingTier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tierIDStr := vars["tierId"]
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	tier, err := h.repo.UpdatePricingTier(r.Context(), tierID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update pricing tier")
		http.Error(w, "Failed to update pricing tier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tier)
}

// HandleDeletePricingTier deletes a pricing tier
func (h *Handler) HandleDeletePricingTier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tierIDStr := vars["tierId"]
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	err = h.repo.DeletePricingTier(r.Context(), tierID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete pricing tier")
		http.Error(w, "Failed to delete pricing tier", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleListSubscriptions lists all subscriptions with pagination.
func (h *Handler) HandleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	limit, offset := 100, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	subs, err := h.repo.ListAllSubscriptions(limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list subscriptions")
		http.Error(w, "Failed to list subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subs,
	})
}

// HandleCreateSubscription creates a new subscription
func (h *Handler) HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID      uuid.UUID  `json:"tenant_id"`
		PricingTierID uuid.UUID  `json:"pricing_tier_id"`
		TrialEnd      *time.Time `json:"trial_end,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if tenant already has an active subscription
	existing, err := h.repo.GetSubscriptionByTenantID(req.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check existing subscription")
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		http.Error(w, "Tenant already has an active subscription", http.StatusConflict)
		return
	}

	now := time.Now()
	sub := &storage.Subscription{
		TenantID:           req.TenantID,
		PricingTierID:      req.PricingTierID,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0), // 1 month
		TrialEnd:           req.TrialEnd,
	}

	createdSub, err := h.repo.CreateSubscription(r.Context(), sub)
	if err != nil {
		logrus.WithError(err).Error("Failed to create subscription")
		http.Error(w, "Failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSub)
}

// HandleGetSubscription gets a subscription by its ID.
func (h *Handler) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subIDStr := vars["subscriptionId"]
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	subscription, err := h.repo.GetSubscriptionByID(r.Context(), subID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get subscription")
		http.Error(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}
	if subscription == nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// HandleUpdateSubscription updates a subscription
func (h *Handler) HandleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subIDStr := vars["subscriptionId"]
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	subscription, err := h.repo.UpdateSubscription(r.Context(), subID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update subscription")
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscription)
}

// HandleCancelSubscription cancels a subscription
func (h *Handler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subIDStr := vars["subscriptionId"]
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	err = h.repo.CancelSubscription(r.Context(), subID)
	if err != nil {
		logrus.WithError(err).Error("Failed to cancel subscription")
		http.Error(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "canceled"})
}

// HandleListInvoices lists all invoices (admin)
func (h *Handler) HandleListInvoices(w http.ResponseWriter, r *http.Request) {
	limit, offset := 100, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	invoices, err := h.repo.ListAllInvoices(limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list invoices")
		http.Error(w, "Failed to list invoices", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invoices": invoices,
	})
}

// HandleCreateInvoice creates a new invoice
func (h *Handler) HandleCreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID       uuid.UUID  `json:"tenant_id"`
		SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
		AmountDueCents int        `json:"amount_due_cents"`
		PeriodStart    *time.Time `json:"period_start,omitempty"`
		PeriodEnd      *time.Time `json:"period_end,omitempty"`
		DueDate        *time.Time `json:"due_date,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	invoice := &storage.Invoice{
		TenantID:        req.TenantID,
		SubscriptionID:  req.SubscriptionID,
		AmountDueCents:  req.AmountDueCents,
		AmountPaidCents: 0,
		Currency:        "USD",
		PeriodStart:     req.PeriodStart,
		PeriodEnd:       req.PeriodEnd,
		DueDate:         req.DueDate,
	}

	createdInvoice, err := h.repo.CreateInvoice(r.Context(), invoice)
	if err != nil {
		logrus.WithError(err).Error("Failed to create invoice")
		http.Error(w, "Failed to create invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdInvoice)
}

// HandleGetInvoice gets a specific invoice
func (h *Handler) HandleGetInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceIDStr := vars["invoiceId"]
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	invoice, err := h.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get invoice")
		http.Error(w, "Failed to get invoice", http.StatusInternalServerError)
		return
	}
	if invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

// HandleUpdateInvoice updates an invoice
func (h *Handler) HandleUpdateInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceIDStr := vars["invoiceId"]
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	invoice, err := h.repo.UpdateInvoice(r.Context(), invoiceID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update invoice")
		http.Error(w, "Failed to update invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

// HandleGetUsage gets usage data
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	tenantIDStr := r.URL.Query().Get("tenant_id")
	eventType := r.URL.Query().Get("event_type")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if tenantIDStr == "" || eventType == "" {
		http.Error(w, "tenant_id and event_type are required", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		start = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		end = time.Now() // Default to now
	}

	usage, err := h.repo.GetUsageByTenant(tenantID, eventType, start, end)
	if err != nil {
		logrus.WithError(err).Error("Failed to get usage")
		http.Error(w, "Failed to get usage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"usage": usage,
	})
}

// HandleRecordUsage records usage event
func (h *Handler) HandleRecordUsage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID  uuid.UUID   `json:"tenant_id"`
		EventType string      `json:"event_type"`
		Quantity  int         `json:"quantity"`
		Metadata  interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	event := &storage.UsageEvent{
		TenantID:  req.TenantID,
		EventType: req.EventType,
		Quantity:  req.Quantity,
		Metadata:  req.Metadata,
	}

	err := h.repo.RecordUsageEvent(r.Context(), event)
	if err != nil {
		logrus.WithError(err).Error("Failed to record usage event")
		http.Error(w, "Failed to record usage event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

// HandleListCoupons lists coupons
func (h *Handler) HandleListCoupons(w http.ResponseWriter, r *http.Request) {
	coupons, err := h.repo.ListCoupons()
	if err != nil {
		logrus.WithError(err).Error("Failed to list coupons")
		http.Error(w, "Failed to list coupons", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"coupons": coupons,
	})
}

// HandleCreateCoupon creates a new coupon
func (h *Handler) HandleCreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code           string     `json:"code"`
		Name           string     `json:"name"`
		Description    string     `json:"description"`
		DiscountType   string     `json:"discount_type"`
		DiscountValue  int        `json:"discount_value"`
		MaxRedemptions *int       `json:"max_redemptions,omitempty"`
		ValidFrom      *time.Time `json:"valid_from,omitempty"`
		ValidUntil     *time.Time `json:"valid_until,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	coupon := &storage.Coupon{
		Code:           strings.ToUpper(req.Code),
		Name:           req.Name,
		Description:    req.Description,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		MaxRedemptions: req.MaxRedemptions,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		IsActive:       true,
	}

	createdCoupon, err := h.repo.CreateCoupon(r.Context(), coupon)
	if err != nil {
		logrus.WithError(err).Error("Failed to create coupon")
		http.Error(w, "Failed to create coupon", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCoupon)
}

// HandleRedeemCoupon redeems a coupon
func (h *Handler) HandleRedeemCoupon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	couponIDStr := vars["couponId"]
	couponID, err := uuid.Parse(couponIDStr)
	if err != nil {
		http.Error(w, "Invalid coupon ID", http.StatusBadRequest)
		return
	}

	var req struct {
		TenantID       uuid.UUID  `json:"tenant_id"`
		SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	redemption, err := h.repo.RedeemCoupon(r.Context(), couponID, req.TenantID, req.SubscriptionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to redeem coupon")
		http.Error(w, "Failed to redeem coupon", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redemption)
}

// =============================================================================
// Affiliate / Referral Commission handlers
// =============================================================================

// HandleListAffiliateCodes lists all affiliate codes
// GET /v1/admin/billing/affiliate-codes
func (h *Handler) HandleListAffiliateCodes(w http.ResponseWriter, r *http.Request) {
	codes, err := h.repo.ListAffiliateCodes()
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate codes")
		http.Error(w, "Failed to list affiliate codes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"affiliate_codes": codes,
	})
}

// HandleCreateAffiliateCode creates a new affiliate code
// POST /v1/admin/billing/affiliate-codes
func (h *Handler) HandleCreateAffiliateCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code           string     `json:"code"`
		PublisherID    uuid.UUID  `json:"publisher_id"`
		TenantID       *uuid.UUID `json:"tenant_id,omitempty"`
		Name           string     `json:"name"`
		Description    string     `json:"description,omitempty"`
		CommissionType string     `json:"commission_type"`
		CommissionValue float64   `json:"commission_value"`
		MaxCommissions *int       `json:"max_commissions,omitempty"`
		MaxReferrals   *int       `json:"max_referrals,omitempty"`
		ValidFrom      *time.Time `json:"valid_from,omitempty"`
		ValidUntil     *time.Time `json:"valid_until,omitempty"`
		UTMSource      string     `json:"utm_source,omitempty"`
		UTMCampaign    string     `json:"utm_campaign,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	code := &storage.AffiliateCode{
		Code:            strings.ToUpper(req.Code),
		PublisherID:     req.PublisherID,
		TenantID:        req.TenantID,
		Name:            req.Name,
		Description:     req.Description,
		CommissionType:  req.CommissionType,
		CommissionValue: req.CommissionValue,
		MaxCommissions:  req.MaxCommissions,
		MaxReferrals:    req.MaxReferrals,
		ValidFrom:       req.ValidFrom,
		ValidUntil:      req.ValidUntil,
		UTMSource:       req.UTMSource,
		UTMCampaign:     req.UTMCampaign,
		IsActive:       true,
	}

	createdCode, err := h.repo.CreateAffiliateCode(r.Context(), code)
	if err != nil {
		logrus.WithError(err).Error("Failed to create affiliate code")
		http.Error(w, "Failed to create affiliate code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCode)
}

// HandleGetAffiliateCode retrieves an affiliate code by ID
// GET /v1/admin/billing/affiliate-codes/{codeId}
func (h *Handler) HandleGetAffiliateCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeID, err := uuid.Parse(vars["codeId"])
	if err != nil {
		http.Error(w, "Invalid code ID", http.StatusBadRequest)
		return
	}

	code, err := h.repo.GetAffiliateCodeByID(codeID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get affiliate code")
		http.Error(w, "Failed to get affiliate code", http.StatusInternalServerError)
		return
	}
	if code == nil {
		http.Error(w, "Affiliate code not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(code)
}

// HandleUpdateAffiliateCode updates an affiliate code
// PUT /v1/admin/billing/affiliate-codes/{codeId}
func (h *Handler) HandleUpdateAffiliateCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeID, err := uuid.Parse(vars["codeId"])
	if err != nil {
		http.Error(w, "Invalid code ID", http.StatusBadRequest)
		return
	}

	code, err := h.repo.GetAffiliateCodeByID(codeID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get affiliate code")
		http.Error(w, "Failed to get affiliate code", http.StatusInternalServerError)
		return
	}
	if code == nil {
		http.Error(w, "Affiliate code not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name            string     `json:"name,omitempty"`
		Description     string     `json:"description,omitempty"`
		CommissionType  string     `json:"commission_type,omitempty"`
		CommissionValue float64    `json:"commission_value,omitempty"`
		MaxCommissions  *int       `json:"max_commissions,omitempty"`
		MaxReferrals    *int       `json:"max_referrals,omitempty"`
		ValidFrom       *time.Time `json:"valid_from,omitempty"`
		ValidUntil      *time.Time `json:"valid_until,omitempty"`
		IsActive        *bool     `json:"is_active,omitempty"`
		UTMSource       string     `json:"utm_source,omitempty"`
		UTMCampaign     string     `json:"utm_campaign,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		code.Name = req.Name
	}
	if req.Description != "" {
		code.Description = req.Description
	}
	if req.CommissionType != "" {
		code.CommissionType = req.CommissionType
	}
	if req.CommissionValue > 0 {
		code.CommissionValue = req.CommissionValue
	}
	code.MaxCommissions = req.MaxCommissions
	code.MaxReferrals = req.MaxReferrals
	code.ValidFrom = req.ValidFrom
	code.ValidUntil = req.ValidUntil
	if req.IsActive != nil {
		code.IsActive = *req.IsActive
	}
	if req.UTMSource != "" {
		code.UTMSource = req.UTMSource
	}
	if req.UTMCampaign != "" {
		code.UTMCampaign = req.UTMCampaign
	}

	if err := h.repo.UpdateAffiliateCode(r.Context(), code); err != nil {
		logrus.WithError(err).Error("Failed to update affiliate code")
		http.Error(w, "Failed to update affiliate code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(code)
}

// HandleListAffiliateReferrals lists referrals for an affiliate code
// GET /v1/admin/billing/affiliate-codes/{codeId}/referrals
func (h *Handler) HandleListAffiliateReferrals(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeID, err := uuid.Parse(vars["codeId"])
	if err != nil {
		http.Error(w, "Invalid code ID", http.StatusBadRequest)
		return
	}

	referrals, err := h.repo.ListAffiliateReferralsByCode(codeID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate referrals")
		http.Error(w, "Failed to list affiliate referrals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"referrals": referrals,
	})
}

// HandleListAffiliateCommissions lists commissions for an affiliate code
// GET /v1/admin/billing/affiliate-codes/{codeId}/commissions
func (h *Handler) HandleListAffiliateCommissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeID, err := uuid.Parse(vars["codeId"])
	if err != nil {
		http.Error(w, "Invalid code ID", http.StatusBadRequest)
		return
	}

	commissions, err := h.repo.ListAffiliateCommissionsByCode(codeID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate commissions")
		http.Error(w, "Failed to list affiliate commissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"commissions": commissions,
	})
}

// HandleRecordAffiliateReferral records a new affiliate referral
// POST /v1/admin/billing/affiliate-referrals
func (h *Handler) HandleRecordAffiliateReferral(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AffiliateCode string     `json:"affiliate_code"`
		TenantID     uuid.UUID  `json:"tenant_id"`
		SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
		UTMSource    string     `json:"utm_source,omitempty"`
		UTMCampaign  string     `json:"utm_campaign,omitempty"`
		UTContent    string     `json:"utm_content,omitempty"`
		UTMTerm      string     `json:"utm_term,omitempty"`
		IPAddress    string     `json:"ip_address,omitempty"`
		UserAgent    string     `json:"user_agent,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	code, err := h.repo.GetAffiliateCodeByCode(req.AffiliateCode)
	if err != nil {
		logrus.WithError(err).Error("Failed to get affiliate code")
		http.Error(w, "Failed to get affiliate code", http.StatusInternalServerError)
		return
	}
	if code == nil {
		http.Error(w, "Affiliate code not found", http.StatusNotFound)
		return
	}

	referral := &storage.AffiliateReferral{
		AffiliateCodeID: code.ID,
		ReferredTenantID: req.TenantID,
		SubscriptionID: req.SubscriptionID,
		UTMSource: req.UTMSource,
		UTMCampaign: req.UTMCampaign,
		UTContent: req.UTContent,
		UTMTerm: req.UTMTerm,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Status: storage.ReferralStatusPending,
		ReferredAt: time.Now(),
	}

	createdReferral, err := h.repo.CreateAffiliateReferral(r.Context(), referral)
	if err != nil {
		logrus.WithError(err).Error("Failed to create affiliate referral")
		http.Error(w, "Failed to create affiliate referral", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdReferral)
}

// HandleUpdateAffiliateReferralStatus updates the status of an affiliate referral
// PATCH /v1/admin/billing/affiliate-referrals/{referralId}/status
func (h *Handler) HandleUpdateAffiliateReferralStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	referralID, err := uuid.Parse(vars["referralId"])
	if err != nil {
		http.Error(w, "Invalid referral ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateAffiliateReferralStatus(r.Context(), referralID, req.Status); err != nil {
		logrus.WithError(err).Error("Failed to update referral status")
		http.Error(w, "Failed to update referral status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// HandleApproveAffiliateCommission approves a pending commission
// POST /v1/admin/billing/affiliate-commissions/{commissionId}/approve
func (h *Handler) HandleApproveAffiliateCommission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commissionID, err := uuid.Parse(vars["commissionId"])
	if err != nil {
		http.Error(w, "Invalid commission ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateAffiliateCommissionStatus(r.Context(), commissionID, storage.CommissionStatusApproved); err != nil {
		logrus.WithError(err).Error("Failed to approve commission")
		http.Error(w, "Failed to approve commission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
}

// HandleMarkAffiliateCommissionPaid marks a commission as paid
// POST /v1/admin/billing/affiliate-commissions/{commissionId}/paid
func (h *Handler) HandleMarkAffiliateCommissionPaid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commissionID, err := uuid.Parse(vars["commissionId"])
	if err != nil {
		http.Error(w, "Invalid commission ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateAffiliateCommissionStatus(r.Context(), commissionID, storage.CommissionStatusPaid); err != nil {
		logrus.WithError(err).Error("Failed to mark commission as paid")
		http.Error(w, "Failed to mark commission as paid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "paid"})
}

// HandleCalculateAffiliateCommission calculates commission for a given base amount
// POST /v1/admin/billing/affiliate-commissions/calculate
func (h *Handler) HandleCalculateAffiliateCommission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AffiliateCode   string  `json:"affiliate_code"`
		BaseAmountCents int64   `json:"base_amount_cents"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	code, err := h.repo.GetAffiliateCodeByCode(req.AffiliateCode)
	if err != nil {
		logrus.WithError(err).Error("Failed to get affiliate code")
		http.Error(w, "Failed to get affiliate code", http.StatusInternalServerError)
		return
	}
	if code == nil {
		http.Error(w, "Affiliate code not found", http.StatusNotFound)
		return
	}

	baseAmountUSD := float64(req.BaseAmountCents) / 100.0
	commissionCents, commissionUSD := h.repo.CalculateCommission(code.CommissionType, code.CommissionValue, baseAmountUSD)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"affiliate_code":   code.Code,
		"commission_type": code.CommissionType,
		"commission_value": code.CommissionValue,
		"base_amount_cents": req.BaseAmountCents,
		"base_amount_usd":  baseAmountUSD,
		"commission_cents": commissionCents,
		"commission_usd":   commissionUSD,
	})
}

// =============================================================================
// Credit Note handlers (for refund accounting / SOX compliance)
// =============================================================================

// HandleCreateCreditNote creates a new credit note
// POST /v1/admin/billing/credit-notes
func (h *Handler) HandleCreateCreditNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID      uuid.UUID  `json:"tenant_id"`
		InvoiceID     *uuid.UUID `json:"invoice_id,omitempty"`
		Status        string     `json:"status,omitempty"`
		SubtotalCents int        `json:"subtotal_cents"`
		TaxCents      int        `json:"tax_cents"`
		TotalCents    int        `json:"total_cents"`
		Currency      string     `json:"currency"`
		Reason        string     `json:"reason"`
		Description   string     `json:"description,omitempty"`
		IssuedBy      uuid.UUID  `json:"issued_by"`
		Notes         string     `json:"notes,omitempty"`
		LineItems     []struct {
			Description    string `json:"description"`
			Quantity       int    `json:"quantity"`
			UnitPriceCents int    `json:"unit_price_cents"`
			TaxCents       int    `json:"tax_cents"`
			AmountCents    int    `json:"amount_cents"`
			TotalCents     int    `json:"total_cents"`
		} `json:"line_items,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TenantID == uuid.Nil {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	creditNote := &storage.CreditNote{
		TenantID:      req.TenantID,
		InvoiceID:     req.InvoiceID,
		Status:        req.Status,
		SubtotalCents: req.SubtotalCents,
		TaxCents:      req.TaxCents,
		TotalCents:    req.TotalCents,
		Currency:      req.Currency,
		Reason:        req.Reason,
		Description:   req.Description,
		IssuedBy:      req.IssuedBy,
		Notes:         req.Notes,
	}

	created, err := h.repo.CreateCreditNote(r.Context(), creditNote)
	if err != nil {
		logrus.WithError(err).Error("Failed to create credit note")
		http.Error(w, "Failed to create credit note", http.StatusInternalServerError)
		return
	}

	for _, item := range req.LineItems {
		lineItem := &storage.CreditNoteLineItem{
			CreditNoteID:   created.ID,
			Description:    item.Description,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			TaxCents:       item.TaxCents,
			AmountCents:    item.AmountCents,
			TotalCents:     item.TotalCents,
		}
		if err := h.repo.CreateCreditNoteLineItem(r.Context(), lineItem); err != nil {
			logrus.WithError(err).WithField("credit_note_id", created.ID).Error("Failed to create credit note line item")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleListCreditNotes lists credit notes with optional filtering
// GET /v1/admin/billing/credit-notes
func (h *Handler) HandleListCreditNotes(w http.ResponseWriter, r *http.Request) {
	filter := &storage.CreditNoteFilter{}

	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tenantID
		}
	}
	if invoiceIDStr := r.URL.Query().Get("invoice_id"); invoiceIDStr != "" {
		if invoiceID, err := uuid.Parse(invoiceIDStr); err == nil {
			filter.InvoiceID = &invoiceID
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = status
	}
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	limit := 50
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

	creditNotes, total, err := h.repo.ListCreditNotes(r.Context(), filter, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list credit notes")
		http.Error(w, "Failed to list credit notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credit_notes": creditNotes,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

// HandleGetCreditNote retrieves a credit note by ID
// GET /v1/admin/billing/credit-notes/{creditNoteId}
func (h *Handler) HandleGetCreditNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditNoteIDStr := vars["creditNoteId"]
	creditNoteID, err := uuid.Parse(creditNoteIDStr)
	if err != nil {
		http.Error(w, "Invalid credit note ID", http.StatusBadRequest)
		return
	}

	creditNote, err := h.repo.GetCreditNoteWithRelations(r.Context(), creditNoteID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get credit note")
		http.Error(w, "Failed to get credit note", http.StatusInternalServerError)
		return
	}
	if creditNote == nil {
		http.Error(w, "Credit note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creditNote)
}

// HandleUpdateCreditNote updates a credit note
// PATCH /v1/admin/billing/credit-notes/{creditNoteId}
func (h *Handler) HandleUpdateCreditNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditNoteIDStr := vars["creditNoteId"]
	creditNoteID, err := uuid.Parse(creditNoteIDStr)
	if err != nil {
		http.Error(w, "Invalid credit note ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status      string `json:"status,omitempty"`
		Description string `json:"description,omitempty"`
		Notes       string `json:"notes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	creditNote, err := h.repo.GetCreditNoteByID(r.Context(), creditNoteID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get credit note")
		http.Error(w, "Failed to get credit note", http.StatusInternalServerError)
		return
	}
	if creditNote == nil {
		http.Error(w, "Credit note not found", http.StatusNotFound)
		return
	}

	if req.Status != "" {
		creditNote.Status = req.Status
	}
	if req.Description != "" {
		creditNote.Description = req.Description
	}
	if req.Notes != "" {
		creditNote.Notes = req.Notes
	}

	if err := h.repo.UpdateCreditNote(r.Context(), creditNote); err != nil {
		logrus.WithError(err).Error("Failed to update credit note")
		http.Error(w, "Failed to update credit note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creditNote)
}

// HandleVoidCreditNote voids a credit note
// POST /v1/admin/billing/credit-notes/{creditNoteId}/void
func (h *Handler) HandleVoidCreditNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditNoteIDStr := vars["creditNoteId"]
	creditNoteID, err := uuid.Parse(creditNoteIDStr)
	if err != nil {
		http.Error(w, "Invalid credit note ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.VoidCreditNote(r.Context(), creditNoteID); err != nil {
		logrus.WithError(err).Error("Failed to void credit note")
		http.Error(w, "Failed to void credit note", http.StatusInternalServerError)
		return
	}

	creditNote, _ := h.repo.GetCreditNoteByID(r.Context(), creditNoteID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"credit_note": creditNote,
	})
}

// HandleApplyCreditNote marks a credit note as applied
// POST /v1/admin/billing/credit-notes/{creditNoteId}/apply
func (h *Handler) HandleApplyCreditNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditNoteIDStr := vars["creditNoteId"]
	creditNoteID, err := uuid.Parse(creditNoteIDStr)
	if err != nil {
		http.Error(w, "Invalid credit note ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.ApplyCreditNote(r.Context(), creditNoteID); err != nil {
		logrus.WithError(err).Error("Failed to apply credit note")
		http.Error(w, "Failed to apply credit note", http.StatusInternalServerError)
		return
	}

	creditNote, _ := h.repo.GetCreditNoteByID(r.Context(), creditNoteID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"credit_note": creditNote,
	})
}

// HandleGetCreditNoteStats returns credit note statistics
// GET /v1/admin/billing/credit-notes/stats
func (h *Handler) HandleGetCreditNoteStats(w http.ResponseWriter, r *http.Request) {
	var tenantID *uuid.UUID
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if parsed, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &parsed
		}
	}

	stats, err := h.repo.GetCreditNoteStats(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get credit note stats")
		http.Error(w, "Failed to get credit note stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats": stats,
	})
}