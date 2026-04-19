package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleListAdminFunctions handles GET /v1/admin/functions (admin list all functions)
func (h *Handler) HandleListAdminFunctions(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0
	var tenantID *uuid.UUID
	var status *string

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := r.URL.Query().Get("tenant_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			tenantID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}

	functions, total, err := h.repo.ListAllFunctions(r.Context(), limit, offset, tenantID, status)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions (admin)")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"functions": functions,
		"total":     total,
	})
}

// HandleGetAdminFunction handles GET /v1/admin/functions/{functionId}
func (h *Handler) HandleGetAdminFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to get function (admin)")
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// HandleUpdateAdminFunction handles PATCH /v1/admin/functions/{functionId}
func (h *Handler) HandleUpdateAdminFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Only allow certain fields for admin update
	allowed := map[string]interface{}{}
	for _, k := range []string{"name", "region", "version", "status", "providers", "env_vars", "code"} {
		if v, ok := updates[k]; ok {
			allowed[k] = v
		}
	}
	if len(allowed) == 0 {
		http.Error(w, "No allowed fields to update", http.StatusBadRequest)
		return
	}

	function, err := h.repo.UpdateFunction(r.Context(), functionID, allowed)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update function (admin)")
		http.Error(w, "Failed to update function", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// HandleDeleteAdminFunction handles DELETE /v1/admin/functions/{functionId}
func (h *Handler) HandleDeleteAdminFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteFunction(r.Context(), functionID); err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to delete function (admin)")
		http.Error(w, "Failed to delete function", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "deleted"})
}

// HandleToggleAdminFunction handles POST /v1/admin/functions/{functionId}/toggle
func (h *Handler) HandleToggleAdminFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	status := "draft"
	if body.Enabled {
		status = "deployed"
	}
	function, err := h.repo.UpdateFunction(r.Context(), functionID, map[string]interface{}{"status": status})
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to toggle function (admin)")
		http.Error(w, "Failed to toggle function", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// HandleListAdminFunctionDeployments handles GET /v1/admin/functions/{functionId}/deployments
func (h *Handler) HandleListAdminFunctionDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	deployments, err := h.repo.ListFunctionDeployments(r.Context(), functionID, limit)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to list deployments (admin)")
		http.Error(w, "Failed to list deployments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deployments": deployments})
}

// HandleListAdminFunctionLogs handles GET /v1/admin/functions/{functionId}/logs
func (h *Handler) HandleListAdminFunctionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	id := functionID
	logs, err := h.repo.GetFunctionLogs(r.Context(), &id, nil, limit, nil, nil)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to list logs (admin)")
		http.Error(w, "Failed to list logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"logs": logs})
}

// HandleGetAdminFunctionMetrics handles GET /v1/admin/functions/{functionId}/metrics
func (h *Handler) HandleGetAdminFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	// Fetch deployments to derive invocation count and latency estimates
	deployments, err := h.repo.ListFunctionDeployments(r.Context(), functionID, 0)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to list deployments for metrics")
		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	// Fetch log entries to compute invocation and error counts
	logs, err := h.repo.GetFunctionLogs(r.Context(), &functionID, nil, 500, nil, nil)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to fetch logs for metrics")
		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	var invocations, errors int
	var totalLatencyMs int64
	var latencyCount int

	for _, l := range logs {
		switch l.Level {
		case "invocation":
			invocations++
			if dur, ok := l.Metadata["duration_ms"].(float64); ok {
				totalLatencyMs += int64(dur)
				latencyCount++
			}
		case "error":
			errors++
		}
	}

	// If no latency data from logs, fall back to deployment durations
	if latencyCount == 0 && len(deployments) > 0 {
		for _, d := range deployments {
			if d.DurationMs != nil && *d.DurationMs > 0 {
				totalLatencyMs += int64(*d.DurationMs)
				latencyCount++
			}
		}
	}

	avgLatency := 0
	if latencyCount > 0 {
		avgLatency = int(totalLatencyMs / int64(latencyCount))
	}

	// Estimate p50 ≈ avg, p99 ≈ 2× avg (without full distribution)
	latencyP50 := avgLatency
	latencyP99 := avgLatency * 2
	if latencyP99 == 0 {
		latencyP99 = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invocations":    invocations,
		"errors":         errors,
		"latency_p50_ms": latencyP50,
		"latency_p99_ms": latencyP99,
	})
}
