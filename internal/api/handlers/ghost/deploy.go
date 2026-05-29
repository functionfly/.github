package ghost

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) HandleDeployStaging(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	deployID := "deploy_" + uuid.New().String()[:8]
	url := "https://staging-" + deployID + ".functionfly.dev"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"deploy_id": deployID,
		"url":       url,
		"status":    "deployed",
		"phase":     "staging",
	})
}

func (h *Handler) HandleDeployProduction(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	buildID := req.BuildID
	h.mu.RLock()
	build, ok := h.builds[buildID]
	h.mu.RUnlock()

	if ok && build.HumanApprovalRequired {
		writeError(w, http.StatusForbidden, "APPROVAL_REQUIRED", "human approval required before production deployment")
		return
	}

	deployID := "deploy_" + uuid.New().String()[:8]
	url := "https://" + deployID + ".functionfly.app"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"deploy_id": deployID,
		"url":       url,
		"status":    "deployed",
		"phase":     "production",
	})
}

func (h *Handler) HandleSetupMonitoring(w http.ResponseWriter, r *http.Request) {
	var req SetupMonitoringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	config := map[string]interface{}{
		"prometheus_url":        "http://prometheus:9090",
		"grafana_url":           "http://grafana:3000",
		"alert_channels":        []string{"email", "slack"},
		"health_check_interval": "30s",
		"services":              req.Services,
		"dashboards":            req.Dashboards,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"config":  config,
		"message": "Monitoring configured",
	})
}
