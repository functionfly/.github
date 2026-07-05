package backends

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/api/apputil"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/routing"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains backend management handlers
type Handler struct {
	repo       storage.Repository
	routingSvc *routing.Router
}

// NewHandler creates a new backends handler
func NewHandler(repo storage.Repository, routingSvc *routing.Router) *Handler {
	return &Handler{
		repo:       repo,
		routingSvc: routingSvc,
	}
}

// HandleCreateBackend handles backend creation
func (h *Handler) HandleCreateBackend(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Get tenant to check plan limits
	tenant, err := h.repo.GetTenantByID(r.Context(), app.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", app.TenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Check provider limit for the plan
	maxProviders := plans.MaxProviders(tenant.Plan)
	backends, err := h.repo.ListBackendsByAppID(r.Context(), appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
		apierror.WriteError(w, apierror.NewInternal("Failed to list backends"))
		return
	}

	if len(backends) >= maxProviders {
		planName := "Starter"
		if tenant.Plan == "pro" {
			planName = "Pro"
		}
		http.Error(w, fmt.Sprintf("Provider limit reached for your plan (%s: %d). Upgrade to add more.", planName, maxProviders), http.StatusForbidden)
		return
	}

	// Check bundle provider limit if tenant has a bundle subscription
	if err := h.checkBundleProviderLimit(r.Context(), tenant.ID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var req types.CreateBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate required fields
	if req.Provider == "" || req.Region == "" || req.URL == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Provider, region, and URL are required"))
		return
	}

	// Get adapter for validation
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid provider. Must be one of: workers, vercel, fly, deno-deploy, functionfly-edge"))
		return
	}

	// Validate region and URL using provider-specific logic
	if err := adapter.ValidateConfig(req.Region, req.URL); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeValidation, "invalid backend configuration")
		return
	}

	// Reachability check: verify the backend URL is actually serving traffic
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	probeBackend := &storage.Backend{
		ID:           uuid.Nil,
		Provider:     req.Provider,
		Region:       req.Region,
		URL:          req.URL,
		SharedSecret: req.SharedSecret,
	}
	healthResult, healthErr := adapter.HealthCheck(ctx, probeBackend)
	if healthErr != nil {
		logrus.WithError(healthErr).WithField("url", req.URL).Warn("Backend reachability check failed")
		// Don't block creation — just warn. The URL might be temporarily down.
	} else if healthResult != nil && !healthResult.OK && healthResult.StatusCode == http.StatusNotFound {
		// 404 from the backend likely means no function is deployed yet.
		// Return a warning in the response but still create the backend.
		logrus.WithFields(logrus.Fields{
			"url":         req.URL,
			"status_code": healthResult.StatusCode,
		}).Warn("Backend URL returned 404 — no function may be deployed")
	}

	// Generate shared secret if not provided
	sharedSecret := req.SharedSecret
	if sharedSecret == "" {
		// Generate cryptographically secure random string
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		const length = 32
		bytes := make([]byte, length)
		if _, err := rand.Read(bytes); err != nil {
			apierror.WriteError(w, apierror.NewInternal("Failed to generate secure secret"))
			return
		}
		for i, b := range bytes {
			bytes[i] = charset[b%byte(len(charset))]
		}
		sharedSecret = string(bytes)
	}

	backend, err := h.repo.CreateBackend(r.Context(), appID, req.Provider, req.Region, req.URL, sharedSecret, req.Priority)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"app_id":   appID,
			"provider": req.Provider,
			"region":   req.Region,
		}).Error("Failed to create backend")
		apierror.WriteError(w, apierror.NewInternal("Failed to create backend"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(backend)
}

func (h *Handler) checkBundleProviderLimit(ctx context.Context, tenantID uuid.UUID) error {
	sub, err := h.repo.GetBundleSubscriptionByTenant(ctx, tenantID)
	if err != nil || sub == nil {
		return nil
	}

	bundle, err := h.repo.GetPricingBundleByID(ctx, sub.BundleID)
	if err != nil || bundle == nil {
		return nil
	}

	limit, exists := bundle.FeatureLimits["providers"]
	if !exists || limit <= 0 {
		return nil
	}

	count, err := h.repo.CountBackendsByTenant(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to count backends for bundle quota check")
		return nil
	}

	if count >= limit {
		return fmt.Errorf("provider limit reached for your bundle (%d/%d). Upgrade your bundle to add more providers", count, limit)
	}

	return nil
}

// HandleListBackends handles listing backends for an app
func (h *Handler) HandleListBackends(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	backends, err := h.repo.ListBackendsByAppID(r.Context(), appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
		apierror.WriteError(w, apierror.NewInternal("Failed to list backends"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"backends": backends,
	})
}

// HandleGetRoute handles getting routing decisions
func (h *Handler) HandleGetRoute(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Get tenant for plan information
	tenant, err := h.repo.GetTenantByID(r.Context(), app.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", app.TenantID).Error("Failed to get tenant")
		apierror.WriteError(w, apierror.NewInternal("Failed to get tenant"))
		return
	}
	if tenant == nil {
		apierror.WriteError(w, apierror.NewNotFound("Tenant not found"))
		return
	}

	// Get query parameters for method and request ID
	method := r.URL.Query().Get("method")
	if method == "" {
		method = "GET" // Default to GET
	}

	requestID := r.URL.Query().Get("request_id")
	if requestID == "" {
		requestID = fmt.Sprintf("debug-%d", time.Now().UnixNano())
	}

	// Get routing decision
	decision, err := h.routingSvc.SelectBackend(appID, method, requestID, tenant.Plan)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to get routing decision")
		apierror.WriteError(w, apierror.NewInternal("Failed to get routing decision"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// HandleDeployBlueGreen handles blue/green deployment for Cloudflare Workers
func (h *Handler) HandleDeployBlueGreen(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Parse request body
	var req struct {
		Artifact       []byte                 `json:"artifact"`
		Provider       string                 `json:"provider"`
		ProviderConfig map[string]interface{} `json:"provider_config"`
		ZoneID         string                 `json:"zone_id"`
		Domain         string                 `json:"domain"`
		EnableProxied  bool                   `json:"enable_proxied"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Provider != "workers" {
		apierror.WriteError(w, apierror.NewBadRequest("Blue/green deployment is only supported for Cloudflare Workers"))
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid provider"))
		return
	}

	// Check if adapter supports blue/green deployment
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		apierror.WriteError(w, apierror.NewBadRequest("Provider does not support blue/green deployment"))
		return
	}

	// Create deployment spec
	spec := &common.DeploymentSpec{
		Artifact:       req.Artifact,
		AppName:        app.Name,
		ProviderConfig: req.ProviderConfig,
	}

	// Perform blue/green deployment
	result, err := extendedAdapter.DeployBlueGreen(r.Context(), spec, req.ZoneID, req.Domain, req.EnableProxied)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Blue/green deployment failed")
		apierror.WriteError(w, apierror.NewInternal("blue/green deployment failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleLinkProject handles Vercel project linking
func (h *Handler) HandleLinkProject(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Parse request body
	var req struct {
		Provider       string                 `json:"provider"`
		ProviderConfig map[string]interface{} `json:"provider_config"`
		Environment    string                 `json:"environment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Provider != "vercel" {
		apierror.WriteError(w, apierror.NewBadRequest("Project linking is only supported for Vercel"))
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid provider"))
		return
	}

	// Check if adapter supports project linking
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		apierror.WriteError(w, apierror.NewBadRequest("Provider does not support project linking"))
		return
	}

	// Perform project linking
	result, err := extendedAdapter.LinkProject(r.Context(), req.ProviderConfig, appID.String(), req.Environment)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Project linking failed")
		apierror.WriteError(w, apierror.NewInternal("project linking failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleSetSecrets handles setting secrets for Fly.io
func (h *Handler) HandleSetSecrets(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Parse request body
	var req struct {
		Provider       string                 `json:"provider"`
		ProviderConfig map[string]interface{} `json:"provider_config"`
		Secrets        map[string]string      `json:"secrets"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Provider != "fly" {
		apierror.WriteError(w, apierror.NewBadRequest("Secret management is only supported for Fly.io"))
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid provider"))
		return
	}

	// Check if adapter supports secret management
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		apierror.WriteError(w, apierror.NewBadRequest("Provider does not support secret management"))
		return
	}

	// Set secrets
	result, err := extendedAdapter.SetSecrets(r.Context(), req.ProviderConfig, req.Secrets)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to set secrets")
		apierror.WriteError(w, apierror.NewInternal("failed to set secrets"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleListSecrets handles listing secrets for an app's provider (e.g. Fly.io).
// Auto-detects the provider from the app's backends. Returns an empty list when
// provider credentials (api_token, app_name) are not available in the query params.
func (h *Handler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		backends, err := h.repo.ListBackendsByAppID(r.Context(), appID)
		if err != nil {
			logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
			apierror.WriteError(w, apierror.NewInternal("Failed to list backends"))
			return
		}
		for _, b := range backends {
			if b.Enabled {
				provider = b.Provider
				break
			}
		}
	}

	if provider == "" || provider != "fly" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"secrets": []interface{}{}})
		return
	}

	adapter := utils.GetAdapterForProvider(provider)
	if adapter == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"secrets": []interface{}{}})
		return
	}

	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"secrets": []interface{}{}})
		return
	}

	providerConfig := make(map[string]interface{})
	if token := r.URL.Query().Get("api_token"); token != "" {
		providerConfig["api_token"] = token
	}
	if appName := r.URL.Query().Get("app_name"); appName != "" {
		providerConfig["app_name"] = appName
	}

	if providerConfig["api_token"] == nil || providerConfig["app_name"] == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"secrets": []interface{}{}})
		return
	}

	result, err := extendedAdapter.ListSecrets(r.Context(), providerConfig)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list secrets")
		apierror.WriteError(w, apierror.NewInternal("failed to list secrets"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleDeleteBackend handles deleting a backend
func (h *Handler) HandleDeleteBackend(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(r.Context(), h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}

	vars := mux.Vars(r)
	backendIDStr := vars["backendId"]
	if backendIDStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Backend ID is required"))
		return
	}

	backendID, err := uuid.Parse(backendIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid backend ID"))
		return
	}

	// Verify backend belongs to this app
	backend, err := h.repo.GetBackendByID(r.Context(), backendID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Backend not found"))
		return
	}

	if backend.AppID != app.ID {
		apierror.WriteError(w, apierror.NewNotFound("Backend not found"))
		return
	}

	// Delete using GORM directly
	if err := h.repo.DeleteBackend(r.Context(), backendID); err != nil {
		logrus.WithError(err).WithField("backend_id", backendID).Error("Failed to delete backend")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete backend"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
