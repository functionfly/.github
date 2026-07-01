package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/apputil"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
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

// deployURLForSlug returns the public URL for an app based on environment.
func deployURLForSlug(slug string) string {
	if config.IsProduction() {
		return fmt.Sprintf("https://%s.functionfly.com", slug)
	}
	return fmt.Sprintf("http://%s.localhost:8082", slug)
}

// HandleListApps handles GET /v1/apps - list apps for the authenticated user's tenant
func (h *Handler) HandleListApps(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	apps, err := h.repo.ListAppsByTenant(context.Background(), user.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", user.TenantID).Error("Failed to list apps")
		apierror.WriteError(w, apierror.NewInternal("Failed to list apps"))
		return
	}

	responses := make([]*types.AppResponse, 0, len(apps))
	for _, app := range apps {
		responses = append(responses, &types.AppResponse{
			ID:        app.ID.String(),
			Name:      app.Name,
			Slug:      app.Slug,
			TenantID:  app.TenantID.String(),
			DeployUrl: deployURLForSlug(app.Slug),
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req types.CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" || req.Slug == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Name and slug are required"))
		return
	}

	// Validate slug format (lowercase, no spaces, etc.)
	if strings.Contains(req.Slug, " ") || strings.ToLower(req.Slug) != req.Slug {
		apierror.WriteError(w, apierror.NewBadRequest("Slug must be lowercase with no spaces"))
		return
	}

	// Check app limit for the plan
	plan := middleware.GetTenantPlan(r)
	maxApps := plans.MaxApps(plan)
	if maxApps != -1 {
		apps, err := h.repo.ListAppsByTenant(context.Background(), user.TenantID)
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", user.TenantID).Error("Failed to list apps")
			apierror.WriteError(w, apierror.NewInternal("Failed to check app limit"))
			return
		}
		if len(apps) >= maxApps {
			http.Error(w, fmt.Sprintf("App limit reached for your plan (%d). Upgrade to create more.", maxApps), http.StatusForbidden)
			return
		}
	}

	app, err := h.repo.CreateApp(context.Background(), req.Name, req.Slug, user.TenantID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"name":      req.Name,
			"slug":      req.Slug,
			"tenant_id": user.TenantID,
		}).Error("Failed to create app")
		apierror.WriteError(w, apierror.NewInternal("Failed to create app"))
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
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(context.Background(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// HandleGetAppStatus handles retrieving app status with backends
func (h *Handler) HandleGetAppStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(context.Background(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	appID := app.ID

	// Get backend status data
	backendStatuses, err := h.repo.GetBackendStatusByAppID(context.Background(), appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get backend status")
		apierror.WriteError(w, apierror.NewInternal("Failed to get backend status"))
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
		TenantID:  app.TenantID.String(),
		DeployUrl: deployURLForSlug(app.Slug),
		CreatedAt: app.CreatedAt,
	}

	response := &types.StatusResponse{
		App:      appResp,
		Backends: backendStatusResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetAppAnalytics handles GET /v1/apps/{appId}/analytics
func (h *Handler) HandleGetAppAnalytics(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(context.Background(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	appID := app.ID

	// Parse days parameter (default 7)
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	since := time.Now().AddDate(0, 0, -days)

	// Determine interval based on time range
	interval := "1h"
	if days > 7 {
		interval = "1d"
	}

	ctx := context.Background()

	summary, err := h.repo.GetAppAnalyticsSummary(ctx, appID, since)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get app analytics summary")
		apierror.WriteError(w, apierror.NewInternal("Failed to get analytics"))
		return
	}

	requestsOverTime, err := h.repo.GetAppRequestTimeseries(ctx, appID, since, interval)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get request timeseries")
		requestsOverTime = nil
	}

	latencyOverTime, err := h.repo.GetAppLatencyTimeseries(ctx, appID, since, interval)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get latency timeseries")
		latencyOverTime = nil
	}

	topErrors, err := h.repo.GetAppTopErrors(ctx, appID, since)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get top errors")
		topErrors = nil
	}

	backendBreakdown, err := h.repo.GetAppBackendBreakdown(ctx, appID, since)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get backend breakdown")
		backendBreakdown = nil
	}

	response := &storage.AppAnalyticsResponse{
		Summary:          summary,
		RequestsOverTime: requestsOverTime,
		LatencyOverTime:  latencyOverTime,
		TopErrors:        topErrors,
		BackendBreakdown: backendBreakdown,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateApp handles PATCH /v1/apps/{appId}
func (h *Handler) HandleUpdateApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(context.Background(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	var req types.UpdateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	updated, err := h.repo.UpdateApp(context.Background(), app.ID, name)
	if err != nil {
		logrus.WithError(err).WithField("app_id", app.ID).Error("Failed to update app")
		apierror.WriteError(w, apierror.NewInternal("Failed to update app"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&types.AppResponse{
		ID:        updated.ID.String(),
		Name:      updated.Name,
		Slug:      updated.Slug,
		TenantID:  updated.TenantID.String(),
		DeployUrl: deployURLForSlug(updated.Slug),
		CreatedAt: updated.CreatedAt,
	})
}

// HandleDeleteApp handles DELETE /v1/apps/{appId}
func (h *Handler) HandleDeleteApp(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(context.Background(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	if err := h.repo.DeleteApp(context.Background(), app.ID); err != nil {
		logrus.WithError(err).WithField("app_id", app.ID).Error("Failed to delete app")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete app"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
