package deployments

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/deployment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains deployment handlers
type Handler struct {
	repo       storage.Repository
	deploySvc  *deployment.Orchestrator
}

// NewHandler creates a new deployments handler
func NewHandler(repo storage.Repository, deploySvc *deployment.Orchestrator) *Handler {
	return &Handler{
		repo:      repo,
		deploySvc: deploySvc,
	}
}

// HandleDeploy handles deployment creation
func (h *Handler) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appIDStr := vars["appId"]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	// Verify app exists and user has access
	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get app")
		http.Error(w, "Failed to get app", http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}
	if app.TenantID != user.TenantID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req types.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Provider == "" || req.Region == "" || req.Artifact == "" {
		http.Error(w, "Provider, region, and artifact are required", http.StatusBadRequest)
		return
	}

	// Decode base64 artifact
	artifactData, err := base64.StdEncoding.DecodeString(req.Artifact)
	if err != nil {
		http.Error(w, "Invalid artifact data", http.StatusBadRequest)
		return
	}

	// Execute deployment
	deploySpec := &deployment.DeploySpec{
		AppID:          appID,
		Provider:       req.Provider,
		Region:         req.Region,
		Artifact:       artifactData,
		Routes:         req.Routes,
		EnvVars:        req.EnvVars,
		Secrets:        req.Secrets,
		ProviderConfig: req.ProviderConfig,
	}

	result, err := h.deploySvc.Deploy(r.Context(), deploySpec)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"app_id":   appID,
			"provider": req.Provider,
		}).Error("Failed to deploy")
		http.Error(w, fmt.Sprintf("Deployment failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// HandleListDeployments handles listing deployments for an app
func (h *Handler) HandleListDeployments(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appIDStr := vars["appId"]
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		http.Error(w, "Invalid app ID", http.StatusBadRequest)
		return
	}

	// Verify app exists and user has access
	app, err := h.repo.GetAppByID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get app")
		http.Error(w, "Failed to get app", http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "App not found", http.StatusNotFound)
		return
	}
	if app.TenantID != user.TenantID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get limit from query params
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	deployments, err := h.deploySvc.ListDeployments(r.Context(), appID, limit)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list deployments")
		http.Error(w, "Failed to list deployments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployments": deployments,
	})
}

// HandleGetDeployment handles getting a specific deployment
func (h *Handler) HandleGetDeployment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	deploymentIDStr := vars["deploymentId"]
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	deployment, err := h.deploySvc.GetDeploymentStatus(r.Context(), deploymentID)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to get deployment")
		http.Error(w, "Failed to get deployment", http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Verify user has access to the app
	app, err := h.repo.GetAppByID(deployment.AppID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", deployment.AppID).Error("Failed to get app")
		http.Error(w, "Failed to get app", http.StatusInternalServerError)
		return
	}
	if app == nil || app.TenantID != user.TenantID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

// HandleRollback handles deployment rollback
func (h *Handler) HandleRollback(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	deploymentIDStr := vars["deploymentId"]
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	// Get the deployment to verify access
	deployment, err := h.repo.GetDeploymentByID(deploymentID)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to get deployment")
		http.Error(w, "Failed to get deployment", http.StatusInternalServerError)
		return
	}
	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Verify user has access to the app
	app, err := h.repo.GetAppByID(deployment.AppID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", deployment.AppID).Error("Failed to get app")
		http.Error(w, "Failed to get app", http.StatusInternalServerError)
		return
	}
	if app == nil || app.TenantID != user.TenantID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Execute rollback
	result, err := h.deploySvc.Rollback(r.Context(), deployment.AppID, deploymentID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"app_id":        deployment.AppID,
			"deployment_id": deploymentID,
		}).Error("Failed to rollback")
		http.Error(w, fmt.Sprintf("Rollback failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}