// Package analytics provides API handlers for the analytics dashboard
package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/analytics"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles analytics API requests
type Handler struct {
	svc     *analytics.Service
	authSvc *auth.AuthService
}

// NewHandler creates a new analytics handler
func NewHandler(svc *analytics.Service, authSvc *auth.AuthService) *Handler {
	return &Handler{
		svc:     svc,
		authSvc: authSvc,
	}
}

// RegisterRoutes registers analytics routes on the provided router
func (h *Handler) RegisterRoutes(r *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// Dashboard stats
	r.HandleFunc("/analytics/dashboard", authMiddleware.RequireAuth(h.HandleDashboardStats)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/dashboard/{agent_id}", authMiddleware.RequireAuth(h.HandleDashboardStatsForAgent)).Methods("GET", "OPTIONS")

	// Time series data
	r.HandleFunc("/analytics/timeseries", authMiddleware.RequireAuth(h.HandleTimeSeries)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/timeseries/{agent_id}", authMiddleware.RequireAuth(h.HandleTimeSeriesForAgent)).Methods("GET", "OPTIONS")

	// Metrics listing
	r.HandleFunc("/analytics/metrics", authMiddleware.RequireAuth(h.HandleListMetrics)).Methods("GET", "OPTIONS")

	// Aggregated metrics
	r.HandleFunc("/analytics/aggregated", authMiddleware.RequireAuth(h.HandleAggregatedMetrics)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/aggregated/hourly", authMiddleware.RequireAuth(h.HandleHourlyStats)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/aggregated/daily", authMiddleware.RequireAuth(h.HandleDailyStats)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/aggregated/weekly", authMiddleware.RequireAuth(h.HandleWeeklyStats)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/aggregated/monthly", authMiddleware.RequireAuth(h.HandleMonthlyStats)).Methods("GET", "OPTIONS")

	// Run-specific metrics
	r.HandleFunc("/analytics/runs/{run_id}", authMiddleware.RequireAuth(h.HandleRunMetrics)).Methods("GET", "OPTIONS")
	r.HandleFunc("/analytics/runs", authMiddleware.RequireAuth(h.HandleRecentRuns)).Methods("GET", "OPTIONS")

	// Percentiles
	r.HandleFunc("/analytics/percentiles", authMiddleware.RequireAuth(h.HandlePercentiles)).Methods("GET", "OPTIONS")

	// Success rate
	r.HandleFunc("/analytics/success-rate", authMiddleware.RequireAuth(h.HandleSuccessRate)).Methods("GET", "OPTIONS")

	// Error rate
	r.HandleFunc("/analytics/error-rate", authMiddleware.RequireAuth(h.HandleErrorRate)).Methods("GET", "OPTIONS")

	// Throughput
	r.HandleFunc("/analytics/throughput", authMiddleware.RequireAuth(h.HandleThroughput)).Methods("GET", "OPTIONS")

	// Latency
	r.HandleFunc("/analytics/latency", authMiddleware.RequireAuth(h.HandleLatency)).Methods("GET", "OPTIONS")

	// Quality trend
	r.HandleFunc("/analytics/quality-trend", authMiddleware.RequireAuth(h.HandleQualityTrend)).Methods("GET", "OPTIONS")

	// Admin endpoints (require admin role)
	r.HandleFunc("/analytics/admin/aggregate", authMiddleware.RequirePermission("admin:analytics:aggregate")(h.HandleRunAggregation)).Methods("POST", "OPTIONS")
	r.HandleFunc("/analytics/admin/cleanup", authMiddleware.RequirePermission("admin:analytics:cleanup")(h.HandleCleanup)).Methods("POST", "OPTIONS")
}

// HandleDashboardStats returns dashboard statistics for the default agent
func (h *Handler) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	period := parsePeriod(r.URL.Query().Get("period"), 24*time.Hour)

	stats, err := h.svc.GetDashboardStats(r.Context(), period)
	if err != nil {
		logrus.WithError(err).Error("failed to get dashboard stats")
		writeError(w, http.StatusInternalServerError, "failed to get dashboard stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// HandleDashboardStatsForAgent returns dashboard statistics for a specific agent
func (h *Handler) HandleDashboardStatsForAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "agent_id is required")
		return
	}

	period := parsePeriod(r.URL.Query().Get("period"), 24*time.Hour)

	stats, err := h.svc.GetDashboardStatsForAgent(r.Context(), agentID, period)
	if err != nil {
		logrus.WithError(err).Error("failed to get dashboard stats for agent")
		writeError(w, http.StatusInternalServerError, "failed to get dashboard stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// HandleTimeSeries returns time series data for a metric
func (h *Handler) HandleTimeSeries(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metric_type")
	if metricType == "" {
		writeError(w, http.StatusBadRequest, "metric_type is required")
		return
	}

	period := parseAggregationPeriod(r.URL.Query().Get("period"), analytics.AggregationPeriodHourly)
	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	data, err := h.svc.GetTimeSeries(r.Context(), analytics.MetricType(metricType), period, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get time series")
		writeError(w, http.StatusInternalServerError, "failed to get time series data")
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// HandleTimeSeriesForAgent returns time series data for a specific agent
func (h *Handler) HandleTimeSeriesForAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "agent_id is required")
		return
	}

	metricType := r.URL.Query().Get("metric_type")
	if metricType == "" {
		writeError(w, http.StatusBadRequest, "metric_type is required")
		return
	}

	period := parseAggregationPeriod(r.URL.Query().Get("period"), analytics.AggregationPeriodHourly)
	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	data, err := h.svc.GetTimeSeriesForAgent(r.Context(), agentID, analytics.MetricType(metricType), period, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get time series for agent")
		writeError(w, http.StatusInternalServerError, "failed to get time series data")
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// HandleListMetrics lists metrics based on filter criteria
func (h *Handler) HandleListMetrics(w http.ResponseWriter, r *http.Request) {
	filter := analytics.MetricFilter{
		MetricType: analytics.MetricType(r.URL.Query().Get("metric_type")),
		Limit:      parseInt(r.URL.Query().Get("limit"), 100),
		Offset:     parseInt(r.URL.Query().Get("offset"), 0),
	}

	if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		filter.AgentID = agentID
	}

	if start := r.URL.Query().Get("start"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = &t
		}
	}
	if end := r.URL.Query().Get("end"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndTime = &t
		}
	}

	metrics, total, err := h.svc.GetMetrics(r.Context(), filter)
	if err != nil {
		logrus.WithError(err).Error("failed to list metrics")
		writeError(w, http.StatusInternalServerError, "failed to list metrics")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"metrics": metrics,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// HandleAggregatedMetrics returns pre-computed aggregated metrics
func (h *Handler) HandleAggregatedMetrics(w http.ResponseWriter, r *http.Request) {
	filter := analytics.MetricFilter{
		MetricType: analytics.MetricType(r.URL.Query().Get("metric_type")),
		Period:     parseAggregationPeriod(r.URL.Query().Get("period"), analytics.AggregationPeriodDaily),
		Limit:      parseInt(r.URL.Query().Get("limit"), 100),
	}

	if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		filter.AgentID = agentID
	}

	if start := r.URL.Query().Get("start"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = &t
		}
	}

	metrics, err := h.svc.GetAggregatedMetrics(r.Context(), filter)
	if err != nil {
		logrus.WithError(err).Error("failed to get aggregated metrics")
		writeError(w, http.StatusInternalServerError, "failed to get aggregated metrics")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"metrics": metrics,
		"period":  filter.Period,
	})
}

// HandleHourlyStats returns hourly statistics
func (h *Handler) HandleHourlyStats(w http.ResponseWriter, r *http.Request) {
	hours := parseInt(r.URL.Query().Get("hours"), 24)

	stats, err := h.svc.GetHourlyStats(r.Context(), hours)
	if err != nil {
		logrus.WithError(err).Error("failed to get hourly stats")
		writeError(w, http.StatusInternalServerError, "failed to get hourly stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats": stats,
		"hours": hours,
	})
}

// HandleDailyStats returns daily statistics
func (h *Handler) HandleDailyStats(w http.ResponseWriter, r *http.Request) {
	days := parseInt(r.URL.Query().Get("days"), 7)

	stats, err := h.svc.GetDailyStats(r.Context(), days)
	if err != nil {
		logrus.WithError(err).Error("failed to get daily stats")
		writeError(w, http.StatusInternalServerError, "failed to get daily stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats": stats,
		"days":  days,
	})
}

// HandleWeeklyStats returns weekly statistics
func (h *Handler) HandleWeeklyStats(w http.ResponseWriter, r *http.Request) {
	weeks := parseInt(r.URL.Query().Get("weeks"), 4)

	stats, err := h.svc.GetWeeklyStats(r.Context(), weeks)
	if err != nil {
		logrus.WithError(err).Error("failed to get weekly stats")
		writeError(w, http.StatusInternalServerError, "failed to get weekly stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats": stats,
		"weeks": weeks,
	})
}

// HandleMonthlyStats returns monthly statistics
func (h *Handler) HandleMonthlyStats(w http.ResponseWriter, r *http.Request) {
	months := parseInt(r.URL.Query().Get("months"), 6)

	stats, err := h.svc.GetMonthlyStats(r.Context(), months)
	if err != nil {
		logrus.WithError(err).Error("failed to get monthly stats")
		writeError(w, http.StatusInternalServerError, "failed to get monthly stats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats":  stats,
		"months": months,
	})
}

// HandleRunMetrics returns metrics for a specific run
func (h *Handler) HandleRunMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	runIDStr := vars["run_id"]
	if runIDStr == "" {
		writeError(w, http.StatusBadRequest, "run_id is required")
		return
	}

	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run_id format")
		return
	}

	summary, err := h.svc.GetRunMetricsSummary(r.Context(), runID)
	if err != nil {
		logrus.WithError(err).Error("failed to get run metrics")
		writeError(w, http.StatusInternalServerError, "failed to get run metrics")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// HandleRecentRuns returns recent factory runs with metrics
func (h *Handler) HandleRecentRuns(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 10)

	runs, err := h.svc.GetRecentRuns(r.Context(), limit)
	if err != nil {
		logrus.WithError(err).Error("failed to get recent runs")
		writeError(w, http.StatusInternalServerError, "failed to get recent runs")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"runs":  runs,
		"limit": limit,
	})
}

// HandlePercentiles returns percentile statistics for a metric
func (h *Handler) HandlePercentiles(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metric_type")
	if metricType == "" {
		writeError(w, http.StatusBadRequest, "metric_type is required")
		return
	}

	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	p50, p95, p99, err := h.svc.GetMetricPercentiles(r.Context(), analytics.MetricType(metricType), startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get percentiles")
		writeError(w, http.StatusInternalServerError, "failed to get percentiles")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"metric_type": metricType,
		"p50":         p50,
		"p95":         p95,
		"p99":         p99,
		"start":       startTime,
		"end":         endTime,
	})
}

// HandleSuccessRate returns the success rate for a time period
func (h *Handler) HandleSuccessRate(w http.ResponseWriter, r *http.Request) {
	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	rate, err := h.svc.GetSuccessRate(r.Context(), startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get success rate")
		writeError(w, http.StatusInternalServerError, "failed to get success rate")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success_rate": rate,
		"start":        startTime,
		"end":          endTime,
	})
}

// HandleErrorRate returns the error rate for a time period
func (h *Handler) HandleErrorRate(w http.ResponseWriter, r *http.Request) {
	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	rate, err := h.svc.GetErrorRate(r.Context(), startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get error rate")
		writeError(w, http.StatusInternalServerError, "failed to get error rate")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"error_rate": rate,
		"start":      startTime,
		"end":        endTime,
	})
}

// HandleThroughput returns throughput for a time period
func (h *Handler) HandleThroughput(w http.ResponseWriter, r *http.Request) {
	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	throughput, err := h.svc.GetThroughput(r.Context(), startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get throughput")
		writeError(w, http.StatusInternalServerError, "failed to get throughput")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"throughput_per_hour": throughput,
		"start":               startTime,
		"end":                 endTime,
	})
}

// HandleLatency returns latency statistics for a time period
func (h *Handler) HandleLatency(w http.ResponseWriter, r *http.Request) {
	latencyType := r.URL.Query().Get("type")
	if latencyType == "" {
		latencyType = "latency_total"
	}

	startTime, endTime := parseTimeRange(r.URL.Query().Get("start"), r.URL.Query().Get("end"))

	avg, err := h.svc.GetAverageLatency(r.Context(), analytics.MetricType(latencyType), startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("failed to get latency")
		writeError(w, http.StatusInternalServerError, "failed to get latency")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"latency_type": latencyType,
		"avg_ms":       avg,
		"start":        startTime,
		"end":          endTime,
	})
}

// HandleQualityTrend returns the quality trend
func (h *Handler) HandleQualityTrend(w http.ResponseWriter, r *http.Request) {
	period := parsePeriod(r.URL.Query().Get("period"), 24*time.Hour)

	trend, err := h.svc.GetQualityTrend(r.Context(), period)
	if err != nil {
		logrus.WithError(err).Error("failed to get quality trend")
		writeError(w, http.StatusInternalServerError, "failed to get quality trend")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"quality_trend": trend,
		"period":        period.String(),
	})
}

// HandleRunAggregation triggers aggregation job (admin only)
func (h *Handler) HandleRunAggregation(w http.ResponseWriter, r *http.Request) {
	err := h.svc.RunAggregationJob(r.Context())
	if err != nil {
		logrus.WithError(err).Error("failed to run aggregation job")
		writeError(w, http.StatusInternalServerError, "failed to run aggregation job")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "completed",
	})
}

// HandleCleanup triggers cleanup of old metrics (admin only)
func (h *Handler) HandleCleanup(w http.ResponseWriter, r *http.Request) {
	retentionDays := parseInt(r.URL.Query().Get("retention_days"), 90)

	count, err := h.svc.CleanupOldMetrics(r.Context(), retentionDays)
	if err != nil {
		logrus.WithError(err).Error("failed to cleanup old metrics")
		writeError(w, http.StatusInternalServerError, "failed to cleanup old metrics")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"deleted_count":  count,
		"retention_days": retentionDays,
	})
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parsePeriod(s string, defaultPeriod time.Duration) time.Duration {
	if s == "" {
		return defaultPeriod
	}

	// Parse duration strings like "24h", "7d", "30d"
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(daysStr)
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultPeriod
	}
	return d
}

func parseAggregationPeriod(s string, defaultPeriod analytics.AggregationPeriod) analytics.AggregationPeriod {
	switch s {
	case "hourly":
		return analytics.AggregationPeriodHourly
	case "daily":
		return analytics.AggregationPeriodDaily
	case "weekly":
		return analytics.AggregationPeriodWeekly
	case "monthly":
		return analytics.AggregationPeriodMonthly
	default:
		return defaultPeriod
	}
}

func parseTimeRange(startStr, endStr string) (time.Time, time.Time) {
	now := time.Now().UTC()
	var startTime, endTime time.Time

	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		} else {
			startTime = now.Add(-24 * time.Hour)
		}
	} else {
		startTime = now.Add(-24 * time.Hour)
	}

	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		} else {
			endTime = now
		}
	} else {
		endTime = now
	}

	return startTime, endTime
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
