package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Billing handler functions

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

// HandleListSubscriptions lists all subscriptions
func (h *Handler) HandleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	// For now, return empty list - in a real implementation you'd have a ListSubscriptions method
	subscriptions := []*storage.Subscription{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subscriptions,
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

// HandleGetSubscription gets a subscription by tenant ID
func (h *Handler) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subIDStr := vars["subscriptionId"]
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		http.Error(w, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	// For now, we need to get subscription by tenant ID, not by subscription ID
	// This is a simplified implementation
	subscription := &storage.Subscription{
		ID: subID,
		// This would need a proper GetSubscriptionByID method in the repository
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

// HandleListInvoices lists invoices
func (h *Handler) HandleListInvoices(w http.ResponseWriter, r *http.Request) {
	// For now, return empty list - in a real implementation you'd filter by tenant
	invoices := []*storage.Invoice{}

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