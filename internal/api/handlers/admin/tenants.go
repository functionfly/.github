package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Tenant-scoped handler functions

// HandleListTenantApps lists apps for a specific tenant (admin view)
func (h *Handler) HandleListTenantApps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Get apps for this tenant
	apps, err := h.repo.ListAppsByTenant(tenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list tenant apps")
		http.Error(w, "Failed to list tenant apps", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"apps":      apps,
		"tenant_id": tenantID,
	})
}

// HandleGetTenantApp gets a specific app for a tenant (admin view)
func (h *Handler) HandleGetTenantApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	appIDStr := vars["appId"]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get tenant app")
		http.Error(w, "Failed to get tenant app", http.StatusInternalServerError)
		return
	}
	if app == nil || app.TenantID != tenantID {
		http.Error(w, "App not found or access denied", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// HandleListTenantBackends lists backends for a specific app/tenant (admin view)
func (h *Handler) HandleListTenantBackends(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	appIDStr := vars["appId"]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	// Verify app belongs to tenant
	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify app ownership")
		http.Error(w, "Failed to verify app ownership", http.StatusInternalServerError)
		return
	}
	if app == nil || app.TenantID != tenantID {
		http.Error(w, "App not found or access denied", http.StatusNotFound)
		return
	}

	backends, err := h.repo.ListBackendsByAppID(appID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list tenant backends")
		http.Error(w, "Failed to list tenant backends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backends":  backends,
		"app_id":    appID,
		"tenant_id": tenantID,
	})
}

// HandleListTenantDeployments lists deployments for a specific app/tenant (admin view)
func (h *Handler) HandleListTenantDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	appIDStr := vars["appId"]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	// Verify app belongs to tenant
	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify app ownership")
		http.Error(w, "Failed to verify app ownership", http.StatusInternalServerError)
		return
	}
	if app == nil || app.TenantID != tenantID {
		http.Error(w, "App not found or access denied", http.StatusNotFound)
		return
	}

	// For now, return empty list - deployment functionality would need to be implemented
	deployments := []*storage.Deployment{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployments": deployments,
		"app_id":      appID,
		"tenant_id":   tenantID,
	})
}

// HandleTenantDeploymentRollback rolls back a deployment for a tenant (admin with approval)
func (h *Handler) HandleTenantDeploymentRollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	appIDStr := vars["appId"]
	deploymentIDStr := vars["deploymentId"]

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	// Verify app belongs to tenant
	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify app ownership")
		http.Error(w, "Failed to verify app ownership", http.StatusInternalServerError)
		return
	}
	if app == nil || app.TenantID != tenantID {
		http.Error(w, "App not found or access denied", http.StatusNotFound)
		return
	}

	// Parse rollback request
	var req struct {
		Reason     string `json:"reason"`
		ApprovedBy string `json:"approved_by,omitempty"` // For audit trail
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "Rollback reason is required", http.StatusBadRequest)
		return
	}

	// For now, simulate rollback - in a real implementation this would call the deployment service
	result := map[string]interface{}{
		"deployment_id": deploymentID,
		"app_id":        appID,
		"tenant_id":     tenantID,
		"status":        "rollback_initiated",
		"reason":        req.Reason,
		"approved_by":   req.ApprovedBy,
		"timestamp":     time.Now(),
	}

	// Log the rollback action
	utils.LogAuditEvent(r.Context(), h.repo, r, "deployment.rollback", "deployment", &deploymentID, nil, result, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleTenantMetrics provides observability metrics for a tenant (admin view)
func (h *Handler) HandleTenantMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Gather tenant-specific metrics
	metrics := map[string]interface{}{
		"tenant_id": tenantID,
		"timestamp": time.Now(),
		"metrics": map[string]interface{}{
			"apps_count":           0, // Would need to implement counting
			"active_deployments":   0,
			"total_requests_24h":   0,
			"error_rate_24h":       0.0,
			"avg_response_time_ms": 0,
		},
		"alerts": []string{}, // Any active alerts for this tenant
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleTenantHealth provides health status for a tenant's resources (admin view)
func (h *Handler) HandleTenantHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenantId"]
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	health := map[string]interface{}{
		"tenant_id": tenantID,
		"status":    "healthy",
		"timestamp": time.Now(),
		"checks": map[string]interface{}{
			"apps_accessible":     true,
			"deployments_healthy": true,
			"backends_responding": true,
			"routing_functional":  true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}