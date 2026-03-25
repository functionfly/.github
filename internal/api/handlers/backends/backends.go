package backends

import (
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
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/routing"
	"github.com/functionfly/functionfly/internal/storage"
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Get tenant to check plan limits
	tenant, err := h.repo.GetTenantByID(app.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", app.TenantID).Error("Failed to get tenant")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	// Check provider limit for the plan
	maxProviders := plans.MaxProviders(tenant.Plan)
	backends, err := h.repo.ListBackendsByAppID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
		http.Error(w, "Failed to list backends", http.StatusInternalServerError)
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

	var req types.CreateBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Provider == "" || req.Region == "" || req.URL == "" {
		http.Error(w, "Provider, region, and URL are required", http.StatusBadRequest)
		return
	}

	// Get adapter for validation
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		http.Error(w, "Invalid provider. Must be one of: workers, vercel, fly, deno-deploy, functionfly-edge", http.StatusBadRequest)
		return
	}

	// Validate region and URL using provider-specific logic
	if err := adapter.ValidateConfig(req.Region, req.URL); err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
	}

	// Generate shared secret if not provided
	sharedSecret := req.SharedSecret
	if sharedSecret == "" {
		// Generate cryptographically secure random string
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		const length = 32
		bytes := make([]byte, length)
		if _, err := rand.Read(bytes); err != nil {
			http.Error(w, "Failed to generate secure secret", http.StatusInternalServerError)
			return
		}
		for i, b := range bytes {
			bytes[i] = charset[b%byte(len(charset))]
		}
		sharedSecret = string(bytes)
	}

	backend, err := h.repo.CreateBackend(appID, req.Provider, req.Region, req.URL, sharedSecret, req.Priority)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"app_id":   appID,
			"provider": req.Provider,
			"region":   req.Region,
		}).Error("Failed to create backend")
		http.Error(w, "Failed to create backend", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(backend)
}

// HandleListBackends handles listing backends for an app
func (h *Handler) HandleListBackends(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	backends, err := h.repo.ListBackendsByAppID(appID)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list backends")
		http.Error(w, "Failed to list backends", http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Get tenant for plan information
	tenant, err := h.repo.GetTenantByID(app.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", app.TenantID).Error("Failed to get tenant")
		http.Error(w, "Failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
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
		http.Error(w, "Failed to get routing decision", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// HandleDeployBlueGreen handles blue/green deployment for Cloudflare Workers
func (h *Handler) HandleDeployBlueGreen(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider != "workers" {
		http.Error(w, "Blue/green deployment is only supported for Cloudflare Workers", http.StatusBadRequest)
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Check if adapter supports blue/green deployment
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		http.Error(w, "Provider does not support blue/green deployment", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Blue/green deployment failed: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider != "vercel" {
		http.Error(w, "Project linking is only supported for Vercel", http.StatusBadRequest)
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Check if adapter supports project linking
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		http.Error(w, "Provider does not support project linking", http.StatusBadRequest)
		return
	}

	// Perform project linking
	result, err := extendedAdapter.LinkProject(r.Context(), req.ProviderConfig, appID.String(), req.Environment)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Project linking failed")
		http.Error(w, fmt.Sprintf("Project linking failed: %v", err), http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider != "fly" {
		http.Error(w, "Secret management is only supported for Fly.io", http.StatusBadRequest)
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Check if adapter supports secret management
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		http.Error(w, "Provider does not support secret management", http.StatusBadRequest)
		return
	}

	// Set secrets
	result, err := extendedAdapter.SetSecrets(r.Context(), req.ProviderConfig, req.Secrets)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to set secrets")
		http.Error(w, fmt.Sprintf("Failed to set secrets: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleListSecrets handles listing secrets for Fly.io
func (h *Handler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	app, resolveErr := apputil.ResolveAppForRequest(h.repo, user, r)
	if resolveErr != nil {
		http.Error(w, resolveErr.Message, resolveErr.Status)
		return
	}
	appID := app.ID

	// Parse request body
	var req struct {
		Provider       string                 `json:"provider"`
		ProviderConfig map[string]interface{} `json:"provider_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider != "fly" {
		http.Error(w, "Secret management is only supported for Fly.io", http.StatusBadRequest)
		return
	}

	// Get adapter
	adapter := utils.GetAdapterForProvider(req.Provider)
	if adapter == nil {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	// Check if adapter supports secret management
	extendedAdapter, ok := adapter.(common.ExtendedDeploymentAdapter)
	if !ok {
		http.Error(w, "Provider does not support secret management", http.StatusBadRequest)
		return
	}

	// List secrets
	result, err := extendedAdapter.ListSecrets(r.Context(), req.ProviderConfig)
	if err != nil {
		logrus.WithError(err).WithField("app_id", appID).Error("Failed to list secrets")
		http.Error(w, fmt.Sprintf("Failed to list secrets: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
