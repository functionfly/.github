// Package analytics: unified analytics handlers (tenant summary and time series).
package analytics

import (
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/analytics/unified"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// UnifiedHandler handles unified analytics API (tenant summary, time series).
type UnifiedHandler struct {
	svc *unified.Service
}

// NewUnifiedHandler creates a new unified analytics handler.
func NewUnifiedHandler(svc *unified.Service) *UnifiedHandler {
	return &UnifiedHandler{svc: svc}
}

// RegisterUnifiedRoutes registers unified analytics routes (tenant-scoped).
func (h *UnifiedHandler) RegisterUnifiedRoutes(r *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	r.HandleFunc("/analytics/tenant/summary", authMiddleware.RequireAuth(h.HandleTenantSummary)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/tenant/timeseries", authMiddleware.RequireAuth(h.HandleTenantTimeSeries)).Methods("GET", "OPTIONS")
}

// HandleTenantSummary returns the unified summary for the authenticated tenant.
// GET /v1/analytics/tenant/summary?start=...&end=...
func (h *UnifiedHandler) HandleTenantSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	start, end := parseStartEnd(r, 30*24*time.Hour)

	sum, err := h.svc.TenantSummary(r.Context(), claims.TenantID, start, end)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("unified analytics: tenant summary failed")
		writeError(w, http.StatusInternalServerError, "failed to get tenant summary")
		return
	}

	writeJSON(w, http.StatusOK, sum)
}

// HandleTenantTimeSeries returns time series for the authenticated tenant.
// GET /v1/analytics/tenant/timeseries?metric=executions|state_ops|billing|agent_calls|registry_runs&granularity=day&start=...&end=...
func (h *UnifiedHandler) HandleTenantTimeSeries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	metricStr := r.URL.Query().Get("metric")
	if metricStr == "" {
		writeError(w, http.StatusBadRequest, "metric is required (executions, state_ops, billing, agent_calls, registry_runs)")
		return
	}
	kind := unified.MetricKind(metricStr)
	switch kind {
	case unified.MetricKindExecutions, unified.MetricKindStateOps, unified.MetricKindBilling,
		unified.MetricKindAgentCalls, unified.MetricKindRegistryRuns:
		// valid
	default:
		writeError(w, http.StatusBadRequest, "invalid metric")
		return
	}

	granularityStr := r.URL.Query().Get("granularity")
	if granularityStr == "" {
		granularityStr = "day"
	}
	granularity := unified.Granularity(granularityStr)
	switch granularity {
	case unified.GranularityHour, unified.GranularityDay, unified.GranularityWeek, unified.GranularityMonth:
		// valid
	default:
		granularity = unified.GranularityDay
	}

	start, end := parseStartEnd(r, 30*24*time.Hour)

	data, err := h.svc.TenantTimeSeries(r.Context(), claims.TenantID, kind, granularity, start, end)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("unified analytics: tenant timeseries failed")
		writeError(w, http.StatusInternalServerError, "failed to get time series")
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func parseStartEnd(r *http.Request, defaultRange time.Duration) (start, end time.Time) {
	now := time.Now().UTC()
	end = now
	start = now.Add(-defaultRange)

	if s := r.URL.Query().Get("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			start = t.UTC()
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t.UTC()
		}
	}
	if e := r.URL.Query().Get("end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			end = t.UTC()
		} else if t, err := time.Parse("2006-01-02", e); err == nil {
			end = t.UTC()
		}
	}

	if end.Before(start) {
		start, end = end, start
	}
	return start, end
}
