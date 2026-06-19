package deployments

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/apputil"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/apierror"
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()
	app, resolveErr := apputil.ResolveAppForRequest(ctx, h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	var req types.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate required fields
	if req.Provider == "" || req.Region == "" || req.Artifact == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Provider, region, and artifact are required"))
		return
	}

	// Decode base64 artifact
	artifactData, err := base64.StdEncoding.DecodeString(req.Artifact)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid artifact data"))
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
		apierror.WriteError(w, apierror.NewInternal("deployment failed"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	ctx := r.Context()

	app, resolveErr := apputil.ResolveAppForRequest(ctx, h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

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
		apierror.WriteError(w, apierror.NewInternal("Failed to list deployments"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	ctx := r.Context()

	vars := mux.Vars(r)
	deploymentIDStr := vars["deploymentId"]
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid deployment ID"))
		return
	}

	deployment, err := h.deploySvc.GetDeploymentStatus(r.Context(), deploymentID)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to get deployment")
		apierror.WriteError(w, apierror.NewInternal("Failed to get deployment"))
		return
	}
	if deployment == nil {
		apierror.WriteError(w, apierror.NewNotFound("Deployment not found"))
		return
	}

	// Verify user has access to the app
	app, err := h.repo.GetAppByID(ctx, deployment.AppID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", deployment.AppID).Error("Failed to get app")
		apierror.WriteError(w, apierror.NewInternal("Failed to get app"))
		return
	}
	if app == nil || app.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Access denied"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

// HandleRollback handles deployment rollback
func (h *Handler) HandleRollback(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	ctx := r.Context()

	vars := mux.Vars(r)
	deploymentIDStr := vars["deploymentId"]
	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid deployment ID"))
		return
	}

	// Get the deployment to verify access
	deployment, err := h.repo.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to get deployment")
		apierror.WriteError(w, apierror.NewInternal("Failed to get deployment"))
		return
	}
	if deployment == nil {
		apierror.WriteError(w, apierror.NewNotFound("Deployment not found"))
		return
	}

	// Verify user has access to the app
	app, err := h.repo.GetAppByID(ctx, deployment.AppID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", deployment.AppID).Error("Failed to get app")
		apierror.WriteError(w, apierror.NewInternal("Failed to get app"))
		return
	}
	if app == nil || app.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Access denied"))
		return
	}

	// Execute rollback
	result, err := h.deploySvc.Rollback(r.Context(), deployment.AppID, deploymentID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"app_id":        deployment.AppID,
			"deployment_id": deploymentID,
		}).Error("Failed to rollback")
		apierror.WriteError(w, apierror.NewInternal("rollback failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}
