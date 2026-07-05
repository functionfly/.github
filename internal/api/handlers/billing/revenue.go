package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/subscription"
)

// =============================================================================
// Response Types
// =============================================================================

// PlanResponse represents a pricing plan response
type PlanResponse struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	PriceCents            int       `json:"price_cents"`
	PriceMonthly          float64   `json:"price_monthly"`
	PriceAnnual           float64   `json:"price_annual,omitempty"`
	AnnualSavingsPercent  float64   `json:"annual_savings_percent,omitempty"`
	Currency              string    `json:"currency"`
	BillingCycle          string    `json:"billing_cycle"`
	TierType              string    `json:"tier_type"`
	Features              any       `json:"features"`
	TrialDays             int       `json:"trial_days"`
	MaxAgents             int       `json:"max_agents"`
	MaxFunctions          int       `json:"max_functions"`
	MaxExecutionsPerMonth int       `json:"max_executions_per_month"`
	IsActive              bool      `json:"is_active"`
}

// PlansListResponse is the response for listing pricing plans
type PlansListResponse struct {
	Plans       []PlanResponse `json:"plans"`
	Description string         `json:"description"`
}

// VerificationCostResponse represents verification cost for a level
type VerificationCostResponse struct {
	Level       string  `json:"level"`
	PriceCents  int     `json:"price_cents"`
	PriceUSD    float64 `json:"price_usd"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
	MinPlan     *string `json:"min_plan,omitempty"`
}

// VerificationCostsResponse is the response for listing verification costs
type VerificationCostsResponse struct {
	VerificationLevels []VerificationCostResponse `json:"verification_levels"`
	Message            string                     `json:"message"`
}

// VerifyFunctionRequest is the request body for paying for function verification
type VerifyFunctionRequest struct {
	FunctionID uuid.UUID `json:"function_id"`
	Level      string    `json:"level"` // 'basic', 'standard', 'full'
	SuccessURL string    `json:"success_url"`
	CancelURL  string    `json:"cancel_url"`
}

// VerifyFunctionResponse is the response for starting function verification payment
type VerifyFunctionResponse struct {
	SessionID   string    `json:"session_id,omitempty"`
	URL         string    `json:"url,omitempty"`
	PaymentID   uuid.UUID `json:"payment_id,omitempty"`
	AmountCents int       `json:"amount_cents"`
	Currency    string    `json:"currency"`
	Level       string    `json:"level"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
}

// EarningsResponse represents publisher earnings
type EarningsResponse struct {
	TotalPendingCents   int              `json:"total_pending_cents"`
	TotalAvailableCents int              `json:"total_available_cents"`
	TotalWithdrawnCents int              `json:"total_withdrawn_cents"`
	RecentEarnings      []EarningItem    `json:"recent_earnings"`
	PlatformFeePercent  float64          `json:"platform_fee_percent"`
	Summary             *EarningsSummary `json:"summary,omitempty"`
}

// EarningItem represents a single earning entry
type EarningItem struct {
	ID               uuid.UUID  `json:"id"`
	FunctionID       *uuid.UUID `json:"function_id,omitempty"`
	FunctionName     string     `json:"function_name,omitempty"`
	TransactionType  string     `json:"transaction_type"`
	NetAmountCents   int        `json:"net_amount_cents"`
	GrossAmountCents int        `json:"gross_amount_cents"`
	PlatformFeeCents int        `json:"platform_fee_cents"`
	Status           string     `json:"status"`
	EarnedAt         time.Time  `json:"earned_at"`
}

// EarningsSummary provides a monthly breakdown
type EarningsSummary struct {
	Year             int              `json:"year"`
	MonthlyBreakdown []MonthlyEarning `json:"monthly_breakdown"`
}

// MonthlyEarning represents monthly earnings breakdown
type MonthlyEarning struct {
	Month            int `json:"month"`
	TotalCents       int `json:"total_cents"`
	TransactionCount int `json:"transaction_count"`
}

// AgentUsageResponse represents agent usage stats
type AgentUsageResponse struct {
	AgentID             uuid.UUID           `json:"agent_id"`
	TotalCalls          int                 `json:"total_calls"`
	BillableCalls       int                 `json:"billable_calls"`
	OverageCalls        int                 `json:"overage_calls"`
	EstimatedCostCents  int                 `json:"estimated_cost_cents"`
	EstimatedCostUSD    float64             `json:"estimated_cost_usd"`
	ActiveSubscriptions []AgentSubscription `json:"active_subscriptions,omitempty"`
	RecentUsage         []AgentUsageItem    `json:"recent_usage,omitempty"`
}

// AgentUsageItem represents a single usage entry
type AgentUsageItem struct {
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	TotalCalls         int       `json:"total_calls"`
	BillableCalls      int       `json:"billable_calls"`
	OverageCalls       int       `json:"overage_calls"`
	EstimatedCostCents int       `json:"estimated_cost_cents"`
	Status             string    `json:"status"`
}

// AgentSubscription represents an agent subscription
type AgentSubscription struct {
	ID                 uuid.UUID `json:"id"`
	PlanName           string    `json:"plan_name"`
	PricePerAgentCents int       `json:"price_per_agent_cents"`
	MaxAgents          int       `json:"max_agents"`
	Status             string    `json:"status"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
}

// StripeMeterVerificationResponse is the response for verifying Stripe meter integration
type StripeMeterVerificationResponse struct {
	Status               string    `json:"status"`
	StripeConfigured     bool      `json:"stripe_configured"`
	SubscriptionID       string    `json:"subscription_id,omitempty"`
	MeterEventName      string    `json:"meter_event_name,omitempty"`
	EventID             string    `json:"event_id,omitempty"`
	Timestamp           int64     `json:"timestamp"`
	Message             string    `json:"message"`
	TestQuantity        int       `json:"test_quantity"`
	CustomerID          string    `json:"customer_id,omitempty"`
	Error               string    `json:"error,omitempty"`
	SubscriptionChecked bool      `json:"subscription_checked"`
	MeteredItemsFound   int       `json:"metered_items_found"`
}

// ReportUsageRequest is the request body for reporting usage to Stripe
type ReportUsageRequest struct {
	Quantity     int               `json:"quantity"`
	EventType   string            `json:"event_type"` // function_execution, ai_call, etc.
	Metadata    map[string]string `json:"metadata,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

// ReportUsageResponse is the response for reporting usage to Stripe
type ReportUsageResponse struct {
	Success        bool      `json:"success"`
	MeterEventID  string    `json:"meter_event_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	Quantity      int       `json:"quantity"`
	EventType     string    `json:"event_type"`
	Timestamp     int64     `json:"timestamp"`
	Error         string    `json:"error,omitempty"`
	IdempotencyKey string   `json:"idempotency_key"`
}

// MeteredBillingStatusResponse shows metered billing status for a tenant
type MeteredBillingStatusResponse struct {
	TenantID           uuid.UUID `json:"tenant_id"`
	StripeCustomerID   string    `json:"stripe_customer_id,omitempty"`
	HasSubscription    bool      `json:"has_subscription"`
	SubscriptionID     string    `json:"subscription_id,omitempty"`
	MeteredItems       []MeteredItemStatus `json:"metered_items"`
	StripeMeterEnabled bool      `json:"stripe_meter_enabled"`
	LastReportedAt     *time.Time `json:"last_reported_at,omitempty"`
	TotalReportedUsage int       `json:"total_reported_usage"`
}

// MeteredItemStatus shows the status of a metered billing item
type MeteredItemStatus struct {
	SubscriptionItemID string `json:"subscription_item_id"`
	PriceID           string `json:"price_id"`
	MeterEventName    string `json:"meter_event_name,omitempty"`
	IsMetered        bool   `json:"is_metered"`
	LastReportedUsage int    `json:"last_reported_usage"`
}

// =============================================================================
// HandleGetPlans returns the list of available pricing plans
// GET /v1/billing/plans
// =============================================================================

func (h *Handler) HandleGetPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.repo.ListPricingTiersExtended(r.Context())
	if err != nil {
		logrus.WithError(err).Error("billing: failed to list pricing tiers")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve pricing plans")
		return
	}

	response := PlansListResponse{
		Description: "FunctionFly pricing plans - Free, Starter ($29/mo, $278/yr), Professional ($99/mo, $950/yr), Enterprise (custom)",
		Plans:       make([]PlanResponse, 0, len(plans)),
	}

	for _, plan := range plans {
		var features any
		if plan.Features != nil {
			_ = json.Unmarshal(plan.Features, &features)
		}

		planResp := PlanResponse{
			ID:                    plan.ID,
			Name:                  plan.Name,
			Description:           plan.Description,
			PriceCents:            plan.PriceCents,
			PriceMonthly:          float64(plan.PriceCents) / 100.0,
			Currency:              plan.Currency,
			BillingCycle:          plan.BillingCycle,
			TierType:               plan.TierType,
			Features:              features,
			TrialDays:             plan.TrialDays,
			MaxAgents:             plan.MaxAgents,
			MaxFunctions:          plan.MaxFunctions,
			MaxExecutionsPerMonth: plan.MaxExecutionsPerMonth,
			IsActive:              plan.IsActive,
		}

		if plan.AnnualPriceCents != nil && *plan.AnnualPriceCents > 0 {
			planResp.PriceAnnual = float64(*plan.AnnualPriceCents) / 100.0
			monthlyAnnual := float64(plan.PriceCents) * 12 / 100.0
			if monthlyAnnual > 0 {
				planResp.AnnualSavingsPercent = ((monthlyAnnual - planResp.PriceAnnual) / monthlyAnnual) * 100
			}
		}

		response.Plans = append(response.Plans, planResp)
	}

	encodeJSON(w, http.StatusOK, response)
}

// =============================================================================
// HandleGetVerificationCost returns the cost for each verification level
// GET /v1/billing/verification-cost
// =============================================================================

func (h *Handler) HandleGetVerificationCost(w http.ResponseWriter, r *http.Request) {
	fees, err := h.repo.ListVerificationFees(r.Context())
	if err != nil {
		logrus.WithError(err).Error("billing: failed to list verification fees")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve verification costs")
		return
	}

	response := VerificationCostsResponse{
		Message:            "Function verification costs - Level 1 (basic) is free to bootstrap the marketplace",
		VerificationLevels: make([]VerificationCostResponse, 0, len(fees)),
	}

	for _, fee := range fees {
		response.VerificationLevels = append(response.VerificationLevels, VerificationCostResponse{
			Level:       fee.Level,
			PriceCents:  fee.PriceCents,
			PriceUSD:    float64(fee.PriceCents) / 100.0,
			Currency:    fee.Currency,
			Description: fee.Description,
			MinPlan:     fee.MinPlan,
		})
	}

	encodeJSON(w, http.StatusOK, response)
}

// =============================================================================
// HandleVerifyFunction initiates payment for function verification
// POST /v1/billing/verify-function
// =============================================================================

func (h *Handler) HandleVerifyFunction(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req VerifyFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FunctionID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "function_id is required")
		return
	}

	// Get verification fee for the requested level
	fee, err := h.repo.GetVerificationFeeByLevel(r.Context(), req.Level)
	if err != nil {
		logrus.WithError(err).WithField("level", req.Level).Warn("billing: verification fee not found")
		writeJSONError(w, http.StatusBadRequest, "Invalid verification level")
		return
	}

	// Free verification (basic level)
	if fee.PriceCents == 0 {
		// Create a free verification payment record
		payment := &storage.FunctionVerificationPayment{
			FunctionID:        req.FunctionID,
			VerificationLevel: req.Level,
			AmountCents:       0,
			Currency:          "USD",
			Status:            "paid",
			TenantID:          claims.TenantID,
			PaidBy:            &claims.UserID,
		}

		if err := h.repo.CreateFunctionVerificationPayment(r.Context(), payment); err != nil {
			logrus.WithError(err).Error("billing: failed to create verification payment record")
			writeJSONError(w, http.StatusInternalServerError, "Failed to record free verification")
			return
		}

		encodeJSON(w, http.StatusOK, VerifyFunctionResponse{
			PaymentID:   payment.ID,
			AmountCents: 0,
			Currency:    "USD",
			Level:       req.Level,
			Status:      "paid",
			Message:     "Free verification initiated",
		})
		return
	}

	// Check if Stripe is configured for paid verification
	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Payment processing not configured")
		return
	}

	// Get user for Stripe customer creation
	user, err := h.repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	// Build display name for Stripe customer
	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	// Create or get Stripe customer
	customerID, err := payment.CreateOrGetStripeCustomer(r.Context(), h.repo, claims.TenantID, user.Email, user.Name)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing: create stripe customer for verification")
		writeJSONError(w, http.StatusInternalServerError, "Failed to prepare payment")
		return
	}

	// Build success and cancel URLs
	successURL := req.SuccessURL
	cancelURL := req.CancelURL
	if successURL == "" {
		successURL = config.GetFrontendURL() + "/functions/" + req.FunctionID.String() + "/verification?success=true"
	}
	if cancelURL == "" {
		cancelURL = config.GetFrontendURL() + "/functions/" + req.FunctionID.String() + "/verification?canceled=true"
	}

	// Create a pending payment record first
	paymentRecord := &storage.FunctionVerificationPayment{
		FunctionID:        req.FunctionID,
		VerificationLevel: req.Level,
		AmountCents:       fee.PriceCents,
		Currency:          fee.Currency,
		Status:            "pending",
		TenantID:          claims.TenantID,
		PaidBy:            &claims.UserID,
	}

	if err := h.repo.CreateFunctionVerificationPayment(r.Context(), paymentRecord); err != nil {
		logrus.WithError(err).Error("billing: failed to create verification payment record")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create payment record")
		return
	}

	// Create Stripe checkout session for verification payment
	checkoutResult, err := payment.CreateVerificationCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		claims.UserID,
		user.Email,
		name,
		paymentRecord.ID,
		req.FunctionID,
		req.Level,
		fee.PriceCents,
		successURL,
		cancelURL,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"payment_id":  paymentRecord.ID,
			"function_id": req.FunctionID,
		}).Error("billing: failed to create verification checkout session")

		// Mark payment as failed since checkout creation failed
		_ = h.repo.UpdateFunctionVerificationPaymentStatus(r.Context(), paymentRecord.ID, "failed", nil, nil)

		writeJSONError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	// Update payment record with checkout session ID
	sessionIDStr := checkoutResult.SessionID
	if err := h.repo.UpdateFunctionVerificationPaymentStatus(r.Context(), paymentRecord.ID, "pending_checkout", nil, &sessionIDStr); err != nil {
		logrus.WithError(err).WithField("payment_id", paymentRecord.ID).Warn("billing: failed to update payment with checkout session ID")
		// Continue - this is not fatal, webhook can still match by metadata
	}

	logrus.WithFields(logrus.Fields{
		"payment_id":   paymentRecord.ID,
		"function_id":  req.FunctionID,
		"level":        req.Level,
		"amount_cents": fee.PriceCents,
		"customer_id":  customerID,
		"checkout_url": checkoutResult.URL,
	}).Info("Verification checkout session created")

	encodeJSON(w, http.StatusOK, VerifyFunctionResponse{
		PaymentID:   paymentRecord.ID,
		AmountCents: fee.PriceCents,
		Currency:    fee.Currency,
		Level:       req.Level,
		Status:      "pending_checkout",
		URL:         checkoutResult.URL,
		Message:     "Complete payment at the provided checkout URL to verify function",
	})
}

// =============================================================================
// HandleGetEarnings returns publisher earnings
// GET /v1/billing/earnings
// =============================================================================

func (h *Handler) HandleGetEarnings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
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

	// Get earnings summary
	pending, available, withdrawn, err := h.repo.GetPublisherEarningsSummary(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get earnings summary")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve earnings summary")
		return
	}

	// Get recent earnings
	earnings, err := h.repo.GetPublisherEarningsByTenant(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get earnings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve earnings")
		return
	}

	// Build response
	recentEarnings := make([]EarningItem, 0, len(earnings))
	for _, e := range earnings {
		recentEarnings = append(recentEarnings, EarningItem{
			ID:               e.ID,
			FunctionID:       e.FunctionID,
			FunctionName:     e.FunctionName,
			TransactionType:  e.TransactionType,
			NetAmountCents:   e.NetAmountCents,
			GrossAmountCents: e.GrossAmountCents,
			PlatformFeeCents: e.PlatformFeeCents,
			Status:           e.Status,
			EarnedAt:         e.EarnedAt,
		})
	}

	response := EarningsResponse{
		TotalPendingCents:   pending,
		TotalAvailableCents: available,
		TotalWithdrawnCents: withdrawn,
		RecentEarnings:      recentEarnings,
		PlatformFeePercent:  storage.DefaultPlatformFeePercent,
	}

	encodeJSON(w, http.StatusOK, response)
}

// =============================================================================
// HandleGetAgentUsage returns agent usage stats
// GET /v1/billing/agent-usage
// =============================================================================

func (h *Handler) HandleGetAgentUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get agent ID from query parameter
	agentIDStr := r.URL.Query().Get("agent_id")
	if agentIDStr == "" {
		writeJSONError(w, http.StatusBadRequest, "agent_id query parameter is required")
		return
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid agent_id format")
		return
	}

	// Tenant verification: Check if agent belongs to this tenant via subscriptions
	// This prevents IDOR - users cannot query usage for agents outside their tenant
	subscriptions, err := h.repo.GetAgentSubscriptionsByTenant(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get agent subscriptions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to verify agent ownership")
		return
	}

	agentBelongsToTenant := false
	for _, s := range subscriptions {
		if s.AgentID == agentID {
			agentBelongsToTenant = true
			break
		}
	}

	if !agentBelongsToTenant {
		writeJSONError(w, http.StatusForbidden, "Access denied: agent does not belong to your tenant")
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

	// Get usage summary with tenant_id for additional security
	totalCalls, billableCalls, overageCalls, estimatedCost, err := h.repo.GetAgentUsageSummary(r.Context(), agentID, claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Warn("billing: failed to get agent usage summary")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve usage summary")
		return
	}

	// Get recent usage records with tenant_id for additional security
	usageRecords, err := h.repo.GetAgentUsageByAgentID(r.Context(), agentID, claims.TenantID, limit, offset)
	if err != nil {
		logrus.WithError(err).WithField("agent_id", agentID).Warn("billing: failed to get agent usage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve usage records")
		return
	}

	// Build response
	recentUsage := make([]AgentUsageItem, 0, len(usageRecords))
	for _, u := range usageRecords {
		recentUsage = append(recentUsage, AgentUsageItem{
			PeriodStart:        u.PeriodStart,
			PeriodEnd:          u.PeriodEnd,
			TotalCalls:         u.TotalCalls,
			BillableCalls:      u.BillableCalls,
			OverageCalls:       u.OverageCalls,
			EstimatedCostCents: u.EstimatedCostCents,
			Status:             u.Status,
		})
	}

	activeSubs := make([]AgentSubscription, 0)
	for _, s := range subscriptions {
		if s.AgentID == agentID && s.Status == "active" {
			activeSubs = append(activeSubs, AgentSubscription{
				ID:                 s.ID,
				PlanName:           s.PlanName,
				PricePerAgentCents: s.PricePerAgentCents,
				MaxAgents:          s.MaxAgents,
				Status:             s.Status,
				CurrentPeriodEnd:   s.CurrentPeriodEnd,
			})
		}
	}

	response := AgentUsageResponse{
		AgentID:             agentID,
		TotalCalls:          totalCalls,
		BillableCalls:       billableCalls,
		OverageCalls:        overageCalls,
		EstimatedCostCents:  estimatedCost,
		EstimatedCostUSD:    float64(estimatedCost) / 100.0,
		ActiveSubscriptions: activeSubs,
		RecentUsage:         recentUsage,
	}

	encodeJSON(w, http.StatusOK, response)
}

// =============================================================================
// HandleSubscribe subscribes to a pricing plan
// POST /v1/billing/subscribe
// =============================================================================

func (h *Handler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateCheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PriceID == "" {
		writeJSONError(w, http.StatusBadRequest, "price_id is required")
		return
	}

	// Delegate to existing HandleCreateCheckoutSession which creates Stripe checkout
	h.HandleCreateCheckoutSession(w, r)
}

// =============================================================================
// HandleVerifyStripeMeterIntegration verifies that Stripe meter integration is working
// GET /v1/billing/meter/verify
// =============================================================================

func (h *Handler) HandleVerifyStripeMeterIntegration(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		writeJSON(w, http.StatusOK, StripeMeterVerificationResponse{
			Status:           "not_configured",
			StripeConfigured: false,
			Message:          "Stripe is not configured (STRIPE_SECRET_KEY not set)",
			Timestamp:        time.Now().UTC().Unix(),
		})
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	customerID := ""
	if tenant.StripeCustomerID != nil {
		customerID = *tenant.StripeCustomerID
	}

	sub, err := h.repo.GetSubscriptionByTenantID(r.Context(), claims.TenantID)
	hasSub := err == nil && sub != nil && sub.StripeSubscriptionID != ""

	meterEventName := os.Getenv("STRIPE_OVERAGE_METER_NAME")
	if meterEventName == "" {
		meterEventName = "functionfly_overage"
	}

	response := StripeMeterVerificationResponse{
		Status:               "configured",
		StripeConfigured:     true,
		MeterEventName:       meterEventName,
		Timestamp:            time.Now().UTC().Unix(),
		Message:              "Stripe is configured",
		TestQuantity:          1,
		CustomerID:           customerID,
		SubscriptionChecked:  true,
		MeteredItemsFound:    0,
	}

	if !hasSub {
		response.Status = "no_subscription"
		response.Message = "No active Stripe subscription found for tenant"
		writeJSON(w, http.StatusOK, response)
		return
	}

	response.SubscriptionID = sub.StripeSubscriptionID

	var meteredItems int
	if sub.StripeSubscriptionID != "" {
		meteredItems = h.countMeteredItemsInSubscription(r.Context(), sub.StripeSubscriptionID)
	}
	response.MeteredItemsFound = meteredItems

	if meteredItems == 0 {
		response.Status = "no_metered_items"
		response.Message = "Subscription exists but no metered billing items found"
		writeJSON(w, http.StatusOK, response)
		return
	}

	now := time.Now().UTC()
	idempotencyKey := fmt.Sprintf("verify_%s_%d", claims.TenantID.String(), now.Unix())

	meterPayload := map[string]interface{}{
		"event_name": meterEventName,
		"timestamp":  now.Unix(),
		"identifier": idempotencyKey,
		"payload": map[string]string{
			"value":              "1",
			"stripe_customer_id": customerID,
			"action":            "verification_test",
		},
	}

	eventID, err := h.createMeterEvent(r.Context(), meterPayload)
	if err != nil {
		response.Status = "api_error"
		response.Error = err.Error()
		response.Message = "Failed to create test meter event"
		writeJSON(w, http.StatusInternalServerError, response)
		return
	}

	response.Status = "success"
	response.EventID = eventID
	response.Message = "Stripe meter integration is working correctly"

	logrus.WithFields(logrus.Fields{
		"tenant_id":     claims.TenantID,
		"customer_id":   customerID,
		"event_id":      eventID,
		"meter_event":   meterEventName,
		"metered_items": meteredItems,
	}).Info("Stripe meter integration verification successful")

	writeJSON(w, http.StatusOK, response)
}

// countMeteredItemsInSubscription counts metered billing items in a Stripe subscription
func (h *Handler) countMeteredItemsInSubscription(ctx context.Context, subscriptionID string) int {
	if !payment.IsConfigured() {
		return 0
	}

	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil || sub == nil {
		return 0
	}

	count := 0
	for _, item := range sub.Items.Data {
		if item.Price != nil && item.Price.Recurring != nil &&
			item.Price.Recurring.UsageType == stripe.PriceRecurringUsageTypeMetered {
			count++
		}
	}
	return count
}

// =============================================================================
// HandleReportUsage reports usage to Stripe for metered billing
// POST /v1/billing/meter/report
// =============================================================================

func (h *Handler) HandleReportUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		writeJSON(w, http.StatusOK, ReportUsageResponse{
			Success:  false,
			TenantID: claims.TenantID,
			Error:    "Stripe is not configured",
		})
		return
	}

	var req ReportUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Quantity <= 0 {
		writeJSONError(w, http.StatusBadRequest, "quantity must be positive")
		return
	}

	if req.EventType == "" {
		req.EventType = "function_execution"
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	customerID := ""
	if tenant.StripeCustomerID != nil {
		customerID = *tenant.StripeCustomerID
	}

	sub, err := h.repo.GetSubscriptionByTenantID(r.Context(), claims.TenantID)
	if err != nil || sub == nil || sub.StripeSubscriptionID == "" {
		writeJSON(w, http.StatusOK, ReportUsageResponse{
			Success:  false,
			TenantID: claims.TenantID,
			Quantity: req.Quantity,
			EventType: req.EventType,
			Error:    "No active subscription found",
		})
		return
	}

	meterEventName := os.Getenv("STRIPE_OVERAGE_METER_NAME")
	if meterEventName == "" {
		meterEventName = "functionfly_overage"
	}

	now := time.Now().UTC()
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("%s_%s_%d", claims.TenantID.String(), req.EventType, now.Unix())
	}

	meterPayload := map[string]interface{}{
		"event_name": meterEventName,
		"timestamp":  now.Unix(),
		"identifier": idempotencyKey,
		"payload": map[string]string{
			"value":              strconv.Itoa(req.Quantity),
			"stripe_customer_id": customerID,
			"event_type":         req.EventType,
		},
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			meterPayload["payload"].(map[string]string)[k] = v
		}
	}

	eventID, err := h.createMeterEvent(r.Context(), meterPayload)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id":  claims.TenantID,
			"quantity":   req.Quantity,
			"event_type": req.EventType,
		}).Error("Failed to report usage to Stripe")

		writeJSON(w, http.StatusOK, ReportUsageResponse{
			Success:        false,
			TenantID:       claims.TenantID,
			Quantity:       req.Quantity,
			EventType:      req.EventType,
			Timestamp:      now.Unix(),
			Error:          err.Error(),
			IdempotencyKey: idempotencyKey,
		})
		return
	}

	logrus.WithFields(logrus.Fields{
		"tenant_id":   claims.TenantID,
		"event_id":    eventID,
		"quantity":    req.Quantity,
		"event_type":  req.EventType,
		"meter_event": meterEventName,
	}).Info("Successfully reported usage to Stripe meter events")

	writeJSON(w, http.StatusOK, ReportUsageResponse{
		Success:        true,
		MeterEventID:   eventID,
		TenantID:       claims.TenantID,
		Quantity:       req.Quantity,
		EventType:      req.EventType,
		Timestamp:      now.Unix(),
		IdempotencyKey: idempotencyKey,
	})
}

// =============================================================================
// HandleGetMeteredBillingStatus returns the metered billing status for a tenant
// GET /v1/billing/meter/status
// =============================================================================

func (h *Handler) HandleGetMeteredBillingStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tenant, err := h.repo.GetTenantByID(r.Context(), claims.TenantID)
	if err != nil || tenant == nil {
		writeJSONError(w, http.StatusNotFound, "Tenant not found")
		return
	}

	customerID := ""
	if tenant.StripeCustomerID != nil {
		customerID = *tenant.StripeCustomerID
	}

	sub, err := h.repo.GetSubscriptionByTenantID(r.Context(), claims.TenantID)
	hasSub := err == nil && sub != nil && sub.StripeSubscriptionID != ""

	response := MeteredBillingStatusResponse{
		TenantID:         claims.TenantID,
		StripeCustomerID: customerID,
		HasSubscription:  hasSub,
		MeteredItems:     []MeteredItemStatus{},
	}

	if hasSub {
		response.SubscriptionID = sub.StripeSubscriptionID

		if payment.IsConfigured() && sub.StripeSubscriptionID != "" {
			stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
			if err == nil && stripeSub != nil {
				for _, item := range stripeSub.Items.Data {
					if item.Price != nil && item.Price.Recurring != nil {
						isMetered := item.Price.Recurring.UsageType == stripe.PriceRecurringUsageTypeMetered
						response.MeteredItems = append(response.MeteredItems, MeteredItemStatus{
							SubscriptionItemID: item.ID,
							PriceID:           item.Price.ID,
							IsMetered:        isMetered,
						})
					}
				}
			}
		}
	}

	if len(response.MeteredItems) > 0 {
		response.StripeMeterEnabled = true
	}

	writeJSON(w, http.StatusOK, response)
}

// =============================================================================
// createMeterEvent creates a meter event in Stripe using the Meter Events API
// This uses a separate endpoint (meter-events.stripe.com) for high-volume usage reporting
// =============================================================================

func (h *Handler) createMeterEvent(ctx context.Context, payload map[string]interface{}) (string, error) {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return "", fmt.Errorf("stripe API key not configured")
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://meter-events.stripe.com/v1/billing/meter_events", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+stripeKey)
	req.Header.Set("Stripe-Version", "2024-04-15")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Stripe: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("stripe API error: %s (code: %s)", result.Error.Message, result.Error.Code)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return result.ID, nil
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}
