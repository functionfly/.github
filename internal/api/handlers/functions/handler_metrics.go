package functions

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetFunctionMetrics handles GET /v1/functions/{id}/metrics
// Returns real metrics derived from deployment history and log entries.
func (h *Handler) HandleGetFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}
	if function.TenantID != user.TenantID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Fetch deployment history (all)
	deployments, err := h.repo.ListFunctionDeployments(r.Context(), functionID, 0)
	if err != nil {
		logrus.WithError(err).Error("Failed to list deployments for metrics")
		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	// Fetch log entries
	logs, err := h.repo.GetFunctionLogs(r.Context(), &functionID, nil, 500, nil, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch logs for metrics")
		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	// Compute deployment stats
	var totalDeploys, successDeploys, failedDeploys int
	var lastDeployedAt *string
	for _, d := range deployments {
		totalDeploys++
		switch d.Status {
		case "success":
			successDeploys++
			if lastDeployedAt == nil {
				ts := d.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
				lastDeployedAt = &ts
			}
		case "failed":
			failedDeploys++
		}
	}
	deploySuccessRate := 0.0
	if totalDeploys > 0 {
		deploySuccessRate = float64(successDeploys) / float64(totalDeploys) * 100
	}

	// Compute log stats
	var errorLogs, warnLogs, infoLogs, debugLogs int
	for _, l := range logs {
		switch l.Level {
		case "error":
			errorLogs++
		case "warn":
			warnLogs++
		case "debug":
			debugLogs++
		default:
			infoLogs++
		}
	}

	type deploymentMetrics struct {
		Total          int     `json:"total"`
		Successful     int     `json:"successful"`
		Failed         int     `json:"failed"`
		SuccessRatePct float64 `json:"success_rate_pct"`
		LastDeployedAt *string `json:"last_deployed_at,omitempty"`
	}
	type logMetrics struct {
		Total int `json:"total"`
		Error int `json:"error"`
		Warn  int `json:"warn"`
		Info  int `json:"info"`
		Debug int `json:"debug"`
	}
	type metricsResponse struct {
		FunctionID  string            `json:"function_id"`
		Status      string            `json:"status"`
		Deployments deploymentMetrics `json:"deployments"`
		Logs        logMetrics        `json:"logs"`
	}

	resp := metricsResponse{
		FunctionID: function.ID.String(),
		Status:     function.Status,
		Deployments: deploymentMetrics{
			Total:          totalDeploys,
			Successful:     successDeploys,
			Failed:         failedDeploys,
			SuccessRatePct: deploySuccessRate,
			LastDeployedAt: lastDeployedAt,
		},
		Logs: logMetrics{
			Total: len(logs),
			Error: errorLogs,
			Warn:  warnLogs,
			Info:  infoLogs,
			Debug: debugLogs,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
