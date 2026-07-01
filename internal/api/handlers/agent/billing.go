package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ============================================================
// Billing & Economic Controls
// ============================================================

// HandleGetBillingSummary returns the billing summary for an agent
// GET /v1/agent/{agent_id}/billing/summary
func (h *Handler) HandleGetBillingSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	period := r.URL.Query().Get("period")
	summary, err := h.billingCtrl.GetAgentSpend(r.Context(), agentID, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_FAILED", "failed to get billing summary")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"summary": summary,
	})
}

// HandleGetWallet returns a wallet summary (balance + lifetime stats) for an agent.
// GET /v1/agent/{agent_id}/wallet
func (h *Handler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}
	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	controls, err := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_FAILED", "failed to get wallet")
		return
	}

	summary, err := h.financialTxRepo.GetAgentWalletSummary(r.Context(), claims.TenantID, agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_FAILED", "failed to compute wallet summary")
		return
	}

	balanceUSD := controls.CreditBalanceUSD
	totalEarnedUSD := 0.0
	totalSpentUSD := 0.0
	var lastEarningAt string
	var lastSpendingAt string
	if summary != nil {
		balanceUSD = summary.BalanceUSD
		totalEarnedUSD = summary.TotalEarnedUSD
		totalSpentUSD = summary.TotalSpentUSD
		if summary.LastEarningAt != nil {
			lastEarningAt = summary.LastEarningAt.UTC().Format(time.RFC3339Nano)
		}
		if summary.LastSpendingAt != nil {
			lastSpendingAt = summary.LastSpendingAt.UTC().Format(time.RFC3339Nano)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"wallet": map[string]interface{}{
			"agent_id":           agentID,
			"balance_usd":        balanceUSD,
			"escrow_balance_usd":  0,
			"total_earned_usd":    totalEarnedUSD,
			"total_spent_usd":     totalSpentUSD,
			"last_earning_at":     lastEarningAt,
			"last_spending_at":    lastSpendingAt,
		},
	})
}

// HandleGetCreditBalance returns the credit balance for an agent
// GET /v1/agent/{agent_id}/credits/balance
func (h *Handler) HandleGetCreditBalance(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	controls, err := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_FAILED", "failed to get credit balance")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                 true,
		"agent_id":           agentID,
		"credit_balance_usd": controls.CreditBalanceUSD,
	})
}

// HandleUpdateSpendCap updates the spend cap for an agent
// PUT /v1/agent/{agent_id}/billing/spend-cap
func (h *Handler) HandleUpdateSpendCap(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		SpendCapDailyUSD   *float64 `json:"spend_cap_daily_usd"`
		SpendCapMonthlyUSD *float64 `json:"spend_cap_monthly_usd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.SpendCapDailyUSD != nil && *req.SpendCapDailyUSD < 0 {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "spend_cap_daily_usd must be non-negative")
		return
	}
	if req.SpendCapMonthlyUSD != nil {
		if *req.SpendCapMonthlyUSD < 0 {
			writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "spend_cap_monthly_usd must be non-negative")
			return
		}
		const maxMonthlyCap = 100000.0
		if *req.SpendCapMonthlyUSD > maxMonthlyCap {
			writeError(w, http.StatusBadRequest, "CAP_TOO_HIGH", fmt.Sprintf("spend_cap_monthly_usd cannot exceed $%.2f", maxMonthlyCap))
			return
		}
	}

	if err := h.billingCtrl.UpdateSpendCap(r.Context(), agentID, req.SpendCapDailyUSD, req.SpendCapMonthlyUSD); err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update spend cap")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "spend cap updated",
	})
}

// HandleGetCostBreakdown returns cost breakdown by function for an agent
// GET /v1/agent/{agent_id}/cost-breakdown
func (h *Handler) HandleGetCostBreakdown(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	breakdown, err := h.attributionRepo.GetCostBreakdown(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "COST_BREAKDOWN_FAILED", "failed to get cost breakdown")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"agent_id": agentID,
		"breakdown": breakdown,
	})
}

// HandleGetModelBreakdown returns cost and token breakdown by model for an agent
// GET /v1/agent/{agent_id}/model-breakdown
func (h *Handler) HandleGetModelBreakdown(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	breakdown, err := h.attributionRepo.GetModelBreakdown(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "MODEL_BREAKDOWN_FAILED", "failed to get model breakdown")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"agent_id": agentID,
		"models":   breakdown,
	})
}

// HandlePurchaseCredits purchases execution credits for an agent
// POST /v1/agent/{agent_id}/credits/purchase
func (h *Handler) HandlePurchaseCredits(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		AmountUSD       float64 `json:"amount_usd"`
		PaymentMethodID string  `json:"payment_method_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.AmountUSD <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive")
		return
	}

	if req.PaymentMethodID == "" {
		// Simulate credit purchase (no Stripe)
		_, err := h.billingCtrl.CheckSpendCap(r.Context(), agentID, req.AmountUSD)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                true,
			"message":           "simulated credit purchase (no Stripe configured)",
			"agent_id":          agentID,
			"credits_added_usd": req.AmountUSD,
		})
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"agent_id":          agentID,
				"credits_added_usd": req.AmountUSD,
			}).Warn("credit purchase payment failed")
			writeErrorFromErr(r, w, http.StatusPaymentRequired, "PAYMENT_FAILED", "credit purchase payment", err)
			return
		}
	} else if req.PaymentMethodID != "" {
		writeError(w, http.StatusBadRequest, "PAYMENTS_NOT_CONFIGURED", "Stripe is not configured; omit payment_method_id for simulated credit purchase")
		return
	}

	if err := h.billingCtrl.AddCredits(r.Context(), agentID, req.AmountUSD); err != nil {
		writeError(w, http.StatusInternalServerError, "PURCHASE_FAILED", "failed to add credits")
		return
	}

	controls, _ := h.billingCtrl.GetOrCreateControls(r.Context(), agentID)
	if h.notificationSvc != nil {
		if err := h.notificationSvc.SendWalletTopUp(r.Context(), claims.UserID, agentID, req.AmountUSD, controls.CreditBalanceUSD); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"user_id":  claims.UserID,
				"agent_id": agentID,
			}).Warn("failed to send wallet top-up notification")
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"message":           "credits purchased successfully",
		"agent_id":          agentID,
		"credits_added_usd": req.AmountUSD,
		"new_balance_usd":   controls.CreditBalanceUSD,
	})
}

// HandleCreateCreditsCheckout creates a Stripe Checkout session for purchasing agent credits.
// POST /v1/agent/{agent_id}/credits/checkout
func (h *Handler) HandleCreateCreditsCheckout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing agent id")
		return
	}

	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		if strings.Contains(err.Error(), "agent not found") {
			logrus.WithField("agent_id", agentID).Warn("credits checkout: agent not found")
			writeError(w, http.StatusNotFound, "AGENT_NOT_FOUND", "agent not found")
			return
		}
		logrus.WithError(err).WithField("agent_id", agentID).Error("credits checkout: agent lookup failed")
		writeError(w, http.StatusInternalServerError, "AGENT_LOOKUP_FAILED", "failed to load agent")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		AmountUSD  float64 `json:"amount_usd"`
		SuccessURL string  `json:"success_url,omitempty"`
		CancelURL  string  `json:"cancel_url,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.AmountUSD <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_AMOUNT", "amount must be positive")
		return
	}

	if !payment.IsConfigured() {
		writeError(w, http.StatusServiceUnavailable, "PAYMENTS_NOT_CONFIGURED", "Stripe is not configured")
		return
	}

	user, err := h.userRepo.GetUserByID(context.Background(), claims.UserID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"agent_id": agentID,
		}).Error("credits checkout: user lookup failed")
		writeError(w, http.StatusInternalServerError, "USER_LOOKUP_FAILED", "failed to load user")
		return
	}
	if user == nil {
		logrus.WithFields(logrus.Fields{
			"user_id":  claims.UserID,
			"agent_id": agentID,
		}).Warn("credits checkout: user missing for JWT (stale token or DB mismatch)")
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	result, err := payment.CreateAgentCreditsCheckoutSession(
		r.Context(),
		h.userRepo,
		claims.TenantID,
		claims.UserID,
		user.Email,
		user.Name,
		agentID,
		req.AmountUSD,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		logrus.WithError(err).Warn("failed to create credits checkout session")
		writeError(w, http.StatusInternalServerError, "CHECKOUT_FAILED", "failed to create checkout session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"session_id": result.SessionID,
		"url":        result.URL,
	})
}

// HandleListAgentTransactions returns the financial transaction ledger for an agent.
// GET /v1/agent/{agent_id}/transactions
func (h *Handler) HandleListAgentTransactions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	transactions, total, err := h.financialTxRepo.ListByAgent(r.Context(), claims.TenantID, agentID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("failed to list agent transactions")
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to retrieve transactions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"agent_id":     agentID,
		"transactions": transactions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}