package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// Handler handles dashboard API requests (tenant-scoped metrics and activity).
type Handler struct {
	repo storage.Repository
}

// NewHandler creates a new dashboard handler.
func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
}

// HandleGetUsage returns usage by day for the current tenant.
// GET /v1/dashboard/usage?days=14
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	days := 14
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 90 {
			days = n
		}
	}

	data, err := h.repo.GetUsageByDay(r.Context(), user.TenantID, days)
	if err != nil {
		logrus.WithError(err).Error("Failed to get dashboard usage")
		apierror.WriteError(w, apierror.NewInternal("Failed to get usage"))
		return
	}

	// Frontend expects { time, value } per point
	out := make([]map[string]interface{}, len(data))
	for i := range data {
		out[i] = map[string]interface{}{
			"time":  data[i].Time,
			"value": data[i].Value,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": out})
}

// HandleGetExecutionRate returns execution rate by hour for the current tenant.
// GET /v1/dashboard/execution-rate?hours=24
func (h *Handler) HandleGetExecutionRate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	hours := 24
	if hr := r.URL.Query().Get("hours"); hr != "" {
		if n, err := strconv.Atoi(hr); err == nil && n > 0 && n <= 168 {
			hours = n
		}
	}

	data, err := h.repo.GetExecutionRateByHour(r.Context(), user.TenantID, hours)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get dashboard execution rate, returning empty data")
		// Return empty data so dashboard still loads (e.g. missing function_logs table or migration)
		data = nil
	}

	out := make([]map[string]interface{}, len(data))
	for i := range data {
		out[i] = map[string]interface{}{
			"time": data[i].Time,
			"rate": data[i].Rate,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": out})
}

// HandleGetActivity returns recent activity (deployments + logs) for the current tenant.
// GET /v1/dashboard/activity?limit=20
func (h *Handler) HandleGetActivity(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	items, err := h.repo.GetRecentActivityForTenant(r.Context(), user.TenantID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get dashboard activity")
		apierror.WriteError(w, apierror.NewInternal("Failed to get activity"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"activities": items})
}

// HandleGetMetrics returns aggregated dashboard metrics for the current tenant.
// GET /v1/dashboard/metrics
func (h *Handler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	metrics, err := h.repo.GetDashboardMetrics(r.Context(), user.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get dashboard metrics")
		apierror.WriteError(w, apierror.NewInternal("Failed to get metrics"))
		return
	}

	out := map[string]interface{}{
		"requests_this_month": metrics.RequestsThisMonth,
		"requests_prev_month": metrics.RequestsPrevMonth,
		"uptime_sparkline":    metrics.UptimeSparkline,
		"requests_sparkline":  metrics.RequestsSparkline,
	}
	if metrics.AvgLatencyMs != nil {
		out["avg_latency_ms"] = *metrics.AvgLatencyMs
	}
	if metrics.UptimePct != nil {
		out["uptime_pct"] = *metrics.UptimePct
	}
	if metrics.UptimePrevPct != nil {
		out["uptime_prev_pct"] = *metrics.UptimePrevPct
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
