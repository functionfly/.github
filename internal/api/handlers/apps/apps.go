package apps

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains app management handlers
type Handler struct {
	repo storage.Repository
}

// NewHandler creates a new apps handler
func NewHandler(repo storage.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// HandleListApps handles GET /v1/apps - list apps for the authenticated user's tenant
func (h *Handler) HandleListApps(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	apps, err := h.repo.ListAppsByTenant(user.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", user.TenantID).Error("Failed to list apps")
		http.Error(w, "Failed to list apps", http.StatusInternalServerError)
		return
	}

	responses := make([]*types.AppResponse, 0, len(apps))
	for _, app := range apps {
		responses = append(responses, &types.AppResponse{
			ID:        app.ID.String(),
			Name:      app.Name,
			Slug:      app.Slug,
			CreatedAt: app.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"apps": responses})
}

// HandleCreateApp handles app creation
func (h *Handler) HandleCreateApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req types.CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Slug == "" {
		http.Error(w, "Name and slug are required", http.StatusBadRequest)
		return
	}

	// Validate slug format (lowercase, no spaces, etc.)
	if strings.Contains(req.Slug, " ") || strings.ToLower(req.Slug) != req.Slug {
		http.Error(w, "Slug must be lowercase with no spaces", http.StatusBadRequest)
		return
	}

	app, err := h.repo.CreateApp(req.Name, req.Slug, user.TenantID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"name":      req.Name,
			"slug":      req.Slug,
			"tenant_id": user.TenantID,
		}).Error("Failed to create app")
		http.Error(w, "Failed to create app", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// HandleGetApp handles retrieving a specific app
func (h *Handler) HandleGetApp(w http.ResponseWriter, r *http.Request) {
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

	// Check if user belongs to the same tenant as the app
	if app.TenantID != user.TenantID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// HandleGetAppStatus handles retrieving app status with backends
func (h *Handler) HandleGetAppStatus(w http.ResponseWriter, r *http.Request) {
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

	// Get backend status data
	backendStatuses, err := h.repo.GetBackendStatusByAppID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get backend status")
		http.Error(w, "Failed to get backend status", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	backendStatusResponses := make([]*types.BackendStatusResponse, len(backendStatuses))
	for i, status := range backendStatuses {
		backendResp := &types.BackendStatusResponse{
			Backend: &types.BackendResponse{
				ID:           status.Backend.ID.String(),
				Provider:     status.Backend.Provider,
				Region:       status.Backend.Region,
				URL:          status.Backend.URL,
				SharedSecret: status.Backend.SharedSecret,
				CreatedAt:    status.Backend.CreatedAt,
			},
		}

		// Add circuit state if it exists
		if status.CircuitState != nil {
			backendResp.CircuitState = &types.CircuitStateResponse{
				State:         status.CircuitState.State,
				SinceTs:       status.CircuitState.SinceTs,
				FailCount:     status.CircuitState.FailCount,
				SuccessCount:  status.CircuitState.SuccessCount,
				LastFailureTs: status.CircuitState.LastFailureTs,
			}
		}

		// Add latest health check if it exists
		if status.LatestHealthCheck != nil {
			backendResp.LatestHealthCheck = &types.HealthCheckResponse{
				Timestamp:    status.LatestHealthCheck.Timestamp,
				OK:           status.LatestHealthCheck.OK,
				StatusCode:   status.LatestHealthCheck.StatusCode,
				LatencyMs:    status.LatestHealthCheck.LatencyMs,
				ErrorMessage: status.LatestHealthCheck.ErrorMessage,
			}
		}

		backendStatusResponses[i] = backendResp
	}

	appResp := &types.AppResponse{
		ID:        app.ID.String(),
		Name:      app.Name,
		Slug:      app.Slug,
		CreatedAt: app.CreatedAt,
	}

	response := &types.StatusResponse{
		App:      appResp,
		Backends: backendStatusResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
