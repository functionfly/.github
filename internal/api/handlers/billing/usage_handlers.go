package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// UsageHandler handles real-time usage API endpoints
type UsageHandler struct {
	tracker   services.RealtimeUsageTrackerInterface
	repo      storage.Repository
	logger    *logrus.Logger
}

// NewUsageHandler creates a new usage handler
func NewUsageHandler(tracker services.RealtimeUsageTrackerInterface, repo storage.Repository) *UsageHandler {
	return &UsageHandler{
		tracker: tracker,
		repo:    repo,
		logger:  logrus.New(),
	}
}

// GetRealtimeUsage returns the current real-time quota status for the authenticated tenant
//
// GET /api/v1/usage/realtime
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "executions_used": 500,
//	  "executions_limit": 1000,
//	  "executions_percent": 50.0,
//	  "compute_ms_used": 1800000,
//	  "compute_ms_limit": 3600000,
//	  "compute_ms_percent": 50.0,
//	  "functions_used": 3,
//	  "functions_limit": 5,
//	  "status": "ok",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-30T23:59:59Z",
//	  "last_updated": "2026-04-09T12:34:56Z"
//	}
func (h *UsageHandler) GetRealtimeUsage(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	// Get real-time quota status
	ctx := r.Context()
	status, err := h.tracker.GetQuotaStatus(ctx, tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get quota status")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
		return
	}

	// If tracker is disabled, fall back to database query
	if status == nil {
		status, err = h.getUsageFromDB(ctx, tenantID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get usage from database")
			h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// CheckQuota checks if the tenant has available quota without consuming it
//
// GET /api/v1/usage/check
//
// Response:
//
//	{
//	  "allowed": true,
//	  "status": {
//	    "executions_used": 500,
//	    "executions_limit": 1000,
//	    "executions_percent": 50.0,
//	    ...
//	  }
//	}
func (h *UsageHandler) CheckQuota(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	ctx := r.Context()
	result, err := h.tracker.CheckQuota(ctx, tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to check quota")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to check quota")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetUsageHistory returns historical usage data for the tenant
//
// GET /api/v1/usage/history?start=2026-04-01&end=2026-04-09
//
// Response:
//
//	{
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "total_executions": 500,
//	  "total_compute_ms": 1800000,
//	  "daily_breakdown": [...]
//	}
func (h *UsageHandler) GetUsageHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	// Parse date range
	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	// Get usage rollups from database
	execRollups, err := h.repo.GetUsageByTenant(tenantID, "function_execution", start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get execution rollups")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
		return
	}

	computeRollups, err := h.repo.GetUsageByTenant(tenantID, "compute_time_ms", start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get compute rollups")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
		return
	}

	// Calculate totals
	totalExecutions := 0
	for _, r := range execRollups {
		totalExecutions += r.TotalQuantity
	}

	totalComputeMs := 0
	for _, r := range computeRollups {
		totalComputeMs += r.TotalQuantity
	}

	// Build response
	response := map[string]interface{}{
		"tenant_id":        tenantID.String(),
		"period_start":     start,
		"period_end":       end,
		"total_executions": totalExecutions,
		"total_compute_ms": totalComputeMs,
		"execution_rollups": execRollups,
		"compute_rollups":  computeRollups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUsageByFunction returns usage breakdown by function
//
// GET /api/v1/usage/by-function
//
// Response:
//
//	{
//	  "functions": [
//	    {
//	      "function_id": "uuid",
//	      "function_name": "my-function",
//	      "total_executions": 100,
//	      "total_compute_ms": 500000
//	    }
//	  ]
//	}
func (h *UsageHandler) GetUsageByFunction(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	// This would require a new repository method for function-level aggregation
	// For now, return a placeholder
	response := map[string]interface{}{
		"tenant_id": tenantID.String(),
		"functions": []interface{}{},
		"note":      "Function-level usage aggregation coming soon",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentPeriodUsage returns usage for the current billing period
//
// GET /api/v1/usage/current-period
//
// Response: Same as GetRealtimeUsage but guaranteed to be current period
func (h *UsageHandler) GetCurrentPeriodUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	// Get subscription to determine period
	sub, err := h.repo.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subscription")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve subscription")
		return
	}

	var periodStart, periodEnd time.Time
	if sub != nil {
		periodStart = sub.CurrentPeriodStart
		periodEnd = sub.CurrentPeriodEnd
	} else {
		// Default to calendar month
		now := time.Now().UTC()
		periodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		nextMonth := now.AddDate(0, 1, 0)
		periodEnd = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)
	}

	// Get usage from database for the period
	execRollups, err := h.repo.GetUsageByTenant(tenantID, "function_execution", periodStart, periodEnd)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get execution rollups")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
		return
	}

	computeRollups, err := h.repo.GetUsageByTenant(tenantID, "compute_time_ms", periodStart, periodEnd)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get compute rollups")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
		return
	}

	totalExecutions := 0
	for _, rollup := range execRollups {
		totalExecutions += rollup.TotalQuantity
	}

	totalComputeMs := 0
	for _, rollup := range computeRollups {
		totalComputeMs += rollup.TotalQuantity
	}

	// Get limits from subscription
	limits := h.extractLimits(sub)

	response := map[string]interface{}{
		"tenant_id":         tenantID.String(),
		"period_start":      periodStart,
		"period_end":        periodEnd,
		"executions_used":   totalExecutions,
		"executions_limit":  limits.ExecutionsLimit,
		"compute_ms_used":   totalComputeMs,
		"compute_ms_limit":  limits.ComputeMsLimit,
		"functions_limit":   limits.FunctionsLimit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminGetTenantUsage returns usage for any tenant (admin only)
//
// GET /api/v1/admin/usage/{tenant_id}
func (h *UsageHandler) AdminGetTenantUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid tenant ID")
		return
	}

	ctx := r.Context()
	status, err := h.tracker.GetQuotaStatus(ctx, tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get quota status")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
		return
	}

	if status == nil {
		status, err = h.getUsageFromDB(ctx, tenantID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get usage from database")
			h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve usage data")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetUsageMetrics returns system-level usage metrics (admin only)
//
// GET /api/v1/admin/usage/metrics
func (h *UsageHandler) GetUsageMetrics(w http.ResponseWriter, r *http.Request) {
	// This endpoint returns metrics about the usage tracking system itself
	// For now, return basic info about whether the tracker is enabled
	metrics := map[string]interface{}{
		"realtime_tracking_enabled": h.tracker.IsEnabled(),
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// Helper methods

func (h *UsageHandler) extractTenantID(r *http.Request) uuid.UUID {
	// Try context values
	if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}

	if user, ok := r.Context().Value("user").(*storage.User); ok && user.TenantID != uuid.Nil {
		return user.TenantID
	}

	return uuid.Nil
}

func (h *UsageHandler) parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		// Default to 30 days ago
		start = time.Now().UTC().AddDate(0, 0, -30)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// Set to end of day
		end = end.Add(24*time.Hour - time.Second)
	} else {
		end = time.Now().UTC()
	}

	return start, end, nil
}

func (h *UsageHandler) getUsageFromDB(ctx context.Context, tenantID uuid.UUID) (*services.RealtimeQuotaStatus, error) {
	// Get subscription for limits
	sub, err := h.repo.GetSubscriptionByTenantID(tenantID)
	if err != nil {
		return nil, err
	}

	// Calculate period
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := now.AddDate(0, 1, 0)
	periodEnd := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	// Get usage from database
	execRollups, err := h.repo.GetUsageByTenant(tenantID, "function_execution", periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	computeRollups, err := h.repo.GetUsageByTenant(tenantID, "compute_time_ms", periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	totalExecutions := 0
	for _, r := range execRollups {
		totalExecutions += r.TotalQuantity
	}

	totalComputeMs := 0
	for _, r := range computeRollups {
		totalComputeMs += r.TotalQuantity
	}

	limits := h.extractLimits(sub)

	return &services.RealtimeQuotaStatus{
		TenantID:        tenantID,
		ExecutionsUsed:  totalExecutions,
		ExecutionsLimit: limits.ExecutionsLimit,
		ComputeMsUsed:   totalComputeMs,
		ComputeMsLimit:  limits.ComputeMsLimit,
		FunctionsLimit:  limits.FunctionsLimit,
		Status:          "ok",
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		LastUpdated:     time.Now().UTC(),
	}, nil
}

func (h *UsageHandler) extractLimits(sub *storage.Subscription) services.TenantLimits {
	limits := services.TenantLimits{
		ExecutionsLimit: 1000,
		ComputeMsLimit:  3600000,
		FunctionsLimit:  5,
	}

	if sub == nil || sub.PricingTier == nil || sub.PricingTier.Features == nil {
		return limits
	}

	features, ok := sub.PricingTier.Features.(map[string]interface{})
	if !ok {
		return limits
	}

	if v, ok := features["requests"].(float64); ok {
		limits.ExecutionsLimit = int(v)
	}
	if v, ok := features["included_compute_ms"].(float64); ok {
		limits.ComputeMsLimit = int(v)
	}
	if v, ok := features["included_compute_hours"].(float64); ok {
		limits.ComputeMsLimit = int(v * 3600000)
	}
	if v, ok := features["functions"].(float64); ok {
		limits.FunctionsLimit = int(v)
	}

	return limits
}

func (h *UsageHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

// TenantLimits re-export from services package
type TenantLimits = services.TenantLimits
