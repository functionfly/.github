package payouts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ExtendedHandler provides production-ready payout API endpoints.
type ExtendedHandler struct {
	payoutService *payment.PayoutServiceExtended
	payoutRepo    *storage.PayoutRepository
	repo          storage.Repository
}

// NewExtendedHandler creates a new extended payout handler.
func NewExtendedHandler(
	payoutService *payment.PayoutServiceExtended,
	payoutRepo *storage.PayoutRepository,
	repo storage.Repository,
) *ExtendedHandler {
	return &ExtendedHandler{
		payoutService: payoutService,
		payoutRepo:    payoutRepo,
		repo:          repo,
	}
}

// ─── User-Facing Endpoints ──────────────────────────────────────────────────

// HandleRequestPayoutWithFees creates a payout request with fee calculation and velocity checks.
// POST /v1/payouts/request
func (h *ExtendedHandler) HandleRequestPayoutWithFees(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Payouts are not configured")
		return
	}

	var payload struct {
		AmountCents    int    `json:"amount_cents"`
		IdempotencyKey string `json:"idempotency_key"`
		FeeType        string `json:"fee_type"` // standard, instant
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.AmountCents <= 0 {
		writeJSONError(w, http.StatusBadRequest, "amount_cents must be positive")
		return
	}
	if payload.IdempotencyKey == "" {
		writeJSONError(w, http.StatusBadRequest, "idempotency_key is required")
		return
	}

	result, fee, err := h.payoutService.RequestPayoutWithChecks(
		r.Context(), claims.UserID, payload.AmountCents, payload.IdempotencyKey, payload.FeeType,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id":      claims.UserID,
			"amount_cents": payload.AmountCents,
		}).Error("payouts: failed to request payout")
		writeJSONError(w, http.StatusBadRequest, "Payout request failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"payout": result,
		"fee":    fee,
	})
}

// HandleCancelPayout cancels a pending payout request.
// POST /v1/payouts/{id}/cancel
func (h *ExtendedHandler) HandleCancelPayout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	payoutID, err := uuid.Parse(vars["id"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid payout ID")
		return
	}

	var payload struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	if err := h.payoutService.CancelPayout(r.Context(), claims.UserID, payoutID, payload.Reason); err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Error("payouts: failed to cancel")
		writeJSONError(w, http.StatusBadRequest, "Payout cancellation failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"payout_id":  payoutID.String(),
		"cancelled_at": time.Now().Format(time.RFC3339),
	})
}

// HandleGetSchedulePreference returns the user's payout schedule preference.
// GET /v1/payouts/schedule
func (h *ExtendedHandler) HandleGetSchedulePreference(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pref, err := h.payoutRepo.GetSchedulePreference(r.Context(), claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("payouts: failed to get schedule preference")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve schedule preference")
		return
	}

	if pref == nil {
		// Return default
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"schedule_enabled":     false,
			"frequency":            "weekly",
			"minimum_amount_cents": 5000,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pref)
}

// HandleUpdateSchedulePreference updates the user's payout schedule preference.
// PUT /v1/payouts/schedule
func (h *ExtendedHandler) HandleUpdateSchedulePreference(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var payload struct {
		ScheduleEnabled    bool   `json:"schedule_enabled"`
		Frequency          string `json:"frequency"`
		MinimumAmountCents int    `json:"minimum_amount_cents"`
		DayOfWeek          *int   `json:"day_of_week"`
		DayOfMonth         *int   `json:"day_of_month"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate frequency
	validFreq := map[string]bool{"weekly": true, "biweekly": true, "monthly": true}
	if !validFreq[payload.Frequency] {
		writeJSONError(w, http.StatusBadRequest, "frequency must be weekly, biweekly, or monthly")
		return
	}

	if payload.MinimumAmountCents < 0 {
		writeJSONError(w, http.StatusBadRequest, "minimum_amount_cents cannot be negative")
		return
	}

	// Compute next scheduled time
	now := time.Now()
	var nextScheduled *time.Time
	if payload.ScheduleEnabled {
		var next time.Time
		switch payload.Frequency {
		case "weekly":
			day := time.Sunday
			if payload.DayOfWeek != nil {
				day = time.Weekday(*payload.DayOfWeek)
			}
			daysUntil := int(day - now.Weekday())
			if daysUntil <= 0 {
				daysUntil += 7
			}
			next = now.AddDate(0, 0, daysUntil)
		case "biweekly":
			next = now.AddDate(0, 0, 14)
		case "monthly":
			day := 1
			if payload.DayOfMonth != nil {
				day = *payload.DayOfMonth
			}
			if day > 28 {
				day = 28
			}
			nextMonth := now.AddDate(0, 1, 0)
			next = time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, time.UTC)
		}
		nextScheduled = &next
	}

	pref := &storage.PayoutSchedulePreference{
		UserID:             claims.UserID,
		ScheduleEnabled:    payload.ScheduleEnabled,
		Frequency:          payload.Frequency,
		MinimumAmountCents: payload.MinimumAmountCents,
		DayOfWeek:          payload.DayOfWeek,
		DayOfMonth:         payload.DayOfMonth,
		Currency:           "usd",
		NextScheduledAt:    nextScheduled,
	}

	if err := h.payoutRepo.UpsertSchedulePreference(r.Context(), pref); err != nil {
		logrus.WithError(err).Error("payouts: failed to update schedule preference")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update schedule")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"preference": pref,
	})
}

// HandleGetPayoutHistory returns enhanced payout history with fee info.
// GET /v1/payouts/history
func (h *ExtendedHandler) HandleGetPayoutHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	status := r.URL.Query().Get("status")

	var requests []*storage.PayoutRequest
	var total int
	var err error

	if status != "" {
		requests, total, err = h.payoutRepo.GetPayoutRequestsByStatus(r.Context(), status, limit, offset)
	} else {
		requests, total, err = h.payoutRepo.GetPayoutRequestsByUserID(r.Context(), claims.UserID, limit, offset)
	}

	if err != nil {
		logrus.WithError(err).Error("payouts: failed to list history")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve payout history")
		return
	}

	type historyItem struct {
		*storage.PayoutRequest
		Fee *storage.PayoutFeeDeduction `json:"fee,omitempty"`
	}

	out := make([]historyItem, 0, len(requests))
	for _, req := range requests {
		item := historyItem{PayoutRequest: req}
		fee, _ := h.payoutRepo.GetFeeDedByPayoutRequestID(r.Context(), req.ID)
		if fee != nil {
			item.Fee = fee
		}
		out = append(out, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"payouts": out,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// ─── Admin Endpoints ────────────────────────────────────────────────────────

// HandleAdminListAllPayouts lists all payout requests with filtering.
// GET /v1/admin/payouts
func (h *ExtendedHandler) HandleAdminListAllPayouts(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	status := r.URL.Query().Get("status")

	requests, total, err := h.payoutRepo.GetAllPayoutRequests(r.Context(), status, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("admin payouts: failed to list")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list payouts")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"payouts": requests,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleAdminGetPayoutSummary returns aggregated payout statistics.
// GET /v1/admin/payouts/summary
func (h *ExtendedHandler) HandleAdminGetPayoutSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period") // 7d, 30d, 90d, all
	since := time.Now().AddDate(0, 0, -30) // default 30 days

	switch period {
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "90d":
		since = time.Now().AddDate(0, 0, -90)
	case "all":
		since = time.Time{}
	}

	summary, err := h.payoutRepo.GetPayoutSummary(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("admin payouts: failed to get summary")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get summary")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

// HandleAdminForceProcessPayout force-processes a pending or failed payout.
// POST /v1/admin/payouts/{payoutId}/force-process
func (h *ExtendedHandler) HandleAdminForceProcessPayout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payoutID, err := uuid.Parse(vars["payoutId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid payout ID")
		return
	}

	if err := h.payoutRepo.ForceProcessPayout(r.Context(), payoutID); err != nil {
		logrus.WithError(err).WithField("payout_id", payoutID).Error("admin: failed to force process")
		writeJSONError(w, http.StatusInternalServerError, "Failed to force process payout")
		return
	}

	logrus.WithFields(logrus.Fields{
		"payout_id": payoutID,
		"admin_id":  middleware.GetUserFromContext(r).UserID,
	}).Warn("Payout force-processed by admin")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"payout_id": payoutID.String(),
		"status":    "processing",
	})
}

// HandleAdminGetUserBalance returns a specific user's payout balance.
// GET /v1/admin/payouts/users/{userId}/balance
func (h *ExtendedHandler) HandleAdminGetUserBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	balance, err := h.payoutRepo.GetPayoutBalance(r.Context(), userID)
	if err != nil {
		logrus.WithError(err).Error("admin: failed to get user balance")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get balance")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(balance)
}

// HandleAdminAdjustBalance performs a manual balance adjustment (credit/debit).
// POST /v1/admin/payouts/users/{userId}/adjust
func (h *ExtendedHandler) HandleAdminAdjustBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	adminUser := middleware.GetUserFromContext(r)
	if adminUser == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var payload struct {
		AmountCents int    `json:"amount_cents"` // positive = credit, negative = debit
		Reason      string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if payload.AmountCents == 0 {
		writeJSONError(w, http.StatusBadRequest, "amount_cents cannot be zero")
		return
	}
	if payload.Reason == "" {
		writeJSONError(w, http.StatusBadRequest, "reason is required")
		return
	}

	if payload.AmountCents > 0 {
		err = h.payoutRepo.CreditEarning(r.Context(), userID, payload.AmountCents,
			"admin_adjustment", adminUser.UserID,
			fmt.Sprintf("Admin credit by %s: %s", adminUser.UserID, payload.Reason))
	} else {
		// For debits, we use the reversal mechanism with negative amount
		err = fmt.Errorf("admin debit not supported via this endpoint; use payout request flow")
	}

	if err != nil {
		logrus.WithError(err).Error("admin: failed to adjust balance")
		writeJSONError(w, http.StatusBadRequest, "Balance adjustment failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"admin_id":      adminUser.UserID,
		"amount_cents":  payload.AmountCents,
		"reason":        payload.Reason,
	}).Warn("Admin balance adjustment")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"user_id":      userID.String(),
		"amount_cents": payload.AmountCents,
	})
}

// HandleAdminGetFeeConfigs returns all fee configurations.
// GET /v1/admin/payouts/fee-configs
func (h *ExtendedHandler) HandleAdminGetFeeConfigs(w http.ResponseWriter, r *http.Request) {
	// Fetch known configs
	configs := make([]*storage.PayoutFeeConfig, 0)
	for _, name := range []string{"standard", "instant"} {
		cfg, _ := h.payoutRepo.GetActiveFeeConfig(r.Context(), name)
		if cfg != nil {
			configs = append(configs, cfg)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"fee_configs": configs,
	})
}

// HandleAdminGetPayoutAuditLog returns the ledger entries for a specific payout request.
// GET /v1/admin/payouts/{payoutId}/audit
func (h *ExtendedHandler) HandleAdminGetPayoutAuditLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payoutID, err := uuid.Parse(vars["payoutId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid payout ID")
		return
	}

	payout, err := h.payoutRepo.GetPayoutRequestByID(r.Context(), payoutID)
	if err != nil || payout == nil {
		writeJSONError(w, http.StatusNotFound, "Payout request not found")
		return
	}

	// Get ledger entries for this user related to this payout
	entries, _, err := h.payoutRepo.ListLedgerEntries(r.Context(), payout.UserID, 100, 0)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get audit log")
		return
	}

	// Filter to entries referencing this payout
	var related []*storage.PayoutLedgerEntry
	for _, entry := range entries {
		if entry.ReferenceID != nil && *entry.ReferenceID == payoutID {
			related = append(related, entry)
		}
	}

	// Get fee deduction
	fee, _ := h.payoutRepo.GetFeeDedByPayoutRequestID(r.Context(), payoutID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"payout":      payout,
		"ledger":      related,
		"fee":         fee,
	})
}
