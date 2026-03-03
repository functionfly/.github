package enterprise

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

const (
	defaultSLATargetPercent = 99.99
	defaultPeriodDays       = 30
	maxPeriodDays           = 90
	maxIncidentsLimit       = 50
)

// SLAHandler handles Enterprise SLA dashboard API (overview, uptime history, incidents).
// All endpoints require authentication and an enterprise plan.
type SLAHandler struct {
	repo storage.Repository
}

// NewSLAHandler creates a new Enterprise SLA handler.
func NewSLAHandler(repo storage.Repository) *SLAHandler {
	return &SLAHandler{repo: repo}
}

// requireEnterprisePlan returns a 403 if the current user's tenant is not on enterprise plan.
func (h *SLAHandler) requireEnterprisePlan(w http.ResponseWriter, r *http.Request) bool {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	tenant, err := h.repo.GetTenantByID(user.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", user.TenantID).Error("Failed to get tenant for SLA")
		http.Error(w, "Failed to verify plan", http.StatusInternalServerError)
		return false
	}
	if tenant == nil || !plans.IsEnterpriseTier(tenant.Plan) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "enterprise_required",
			"message": "SLA Dashboard is available only for Enterprise plan customers.",
		})
		return false
	}
	return true
}

// HandleGetSLAOverview returns current uptime, SLA target, and incident count for the period.
// GET /v1/enterprise/sla/overview?days=30
func (h *SLAHandler) HandleGetSLAOverview(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	days := defaultPeriodDays
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= maxPeriodDays {
			days = n
		}
	}

	since := time.Now().AddDate(0, 0, -days)

	count, err := h.repo.CountIncidentsSince(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to count incidents for SLA overview")
		http.Error(w, "Failed to load SLA overview", http.StatusInternalServerError)
		return
	}

	// Current uptime: 100% minus impact from incidents. For MVP we use 99.99% if no incidents, else a simple formula.
	currentUptime := 99.99
	if count > 0 {
		// Simple degradation: each incident in the period reduces by 0.01% (placeholder until we have duration data).
		currentUptime = 99.99 - float64(count)*0.01
		if currentUptime < 99.0 {
			currentUptime = 99.0
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_uptime_percent": currentUptime,
		"sla_target_percent":     defaultSLATargetPercent,
		"incident_count":         count,
		"period_days":            days,
	})
}

// HandleGetUptimeHistory returns daily uptime percentages for the period (for charts).
// GET /v1/enterprise/sla/uptime-history?days=30
func (h *SLAHandler) HandleGetUptimeHistory(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	days := defaultPeriodDays
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= maxPeriodDays {
			days = n
		}
	}

	since := time.Now().UTC().AddDate(0, 0, -days)
	dayCounts, err := h.repo.CountIncidentsGroupedByDay(r.Context(), since)
	if err != nil {
		logrus.WithError(err).Error("Failed to get incident counts by day for uptime history")
		http.Error(w, "Failed to load uptime history", http.StatusInternalServerError)
		return
	}

	countByDate := make(map[string]int)
	for _, dc := range dayCounts {
		countByDate[dc.Date] = dc.Count
	}

	// Build one point per day with real uptime from per-day incident counts.
	points := make([]map[string]interface{}, 0, days)
	now := time.Now().UTC()
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		count := countByDate[dateStr]
		uptime := 99.99
		if count > 0 {
			uptime = 99.99 - float64(count)*0.01
			if uptime < 99.0 {
				uptime = 99.0
			}
		}
		points = append(points, map[string]interface{}{
			"date":           dateStr,
			"uptime_percent": uptime,
			"incident_count": count,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"period_days": days,
		"points":      points,
	})
}

// HandleGetIncidents returns recent incidents for the SLA dashboard.
// GET /v1/enterprise/sla/incidents?limit=20&days=30
func (h *SLAHandler) HandleGetIncidents(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= maxIncidentsLimit {
			limit = n
		}
	}

	days := defaultPeriodDays
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= maxPeriodDays {
			days = n
		}
	}

	since := time.Now().AddDate(0, 0, -days)
	incidents, err := h.repo.ListIncidentsSince(r.Context(), since, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to list incidents for SLA")
		http.Error(w, "Failed to load incidents", http.StatusInternalServerError)
		return
	}

	// Map to API response shape
	list := make([]map[string]interface{}, 0, len(incidents))
	for _, inc := range incidents {
		resolvedAt := ""
		if inc.ResolvedAt != nil {
			resolvedAt = inc.ResolvedAt.Format(time.RFC3339)
		}
		list = append(list, map[string]interface{}{
			"id":          inc.ID.String(),
			"title":       inc.Title,
			"severity":    inc.Severity,
			"status":      inc.Status,
			"description": inc.Description,
			"created_at":  inc.CreatedAt.Format(time.RFC3339),
			"resolved_at": resolvedAt,
			"updated_at":  inc.UpdatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"incidents":   list,
		"period_days": days,
	})
}
