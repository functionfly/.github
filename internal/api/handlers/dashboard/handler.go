package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Failed to get usage", http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		logrus.WithError(err).Error("Failed to get dashboard execution rate")
		http.Error(w, "Failed to get execution rate", http.StatusInternalServerError)
		return
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Failed to get activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"activities": items})
}
