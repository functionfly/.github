package aimodels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/ai/modelprofiles"
	"github.com/functionfly/functionfly/internal/aikeys"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

const catalogCacheKey = "ai:catalog:v1"

type Handler struct {
	repo         storage.Repository
	prefsRepo    *storage.AIModelPreferencesRepository
	byokRepo     *aikeys.Repository
	redisClient  *redis.Client
	aiServiceURL string
	httpClient   *http.Client
}

func NewHandler(
	repo storage.Repository,
	prefsRepo *storage.AIModelPreferencesRepository,
	redisClient *redis.Client,
	aiServiceURL string,
	byokRepo *aikeys.Repository,
) *Handler {
	baseURL := strings.TrimRight(aiServiceURL, "/")
	if baseURL == "" && os.Getenv("DEVELOPMENT") == "true" {
		baseURL = "http://localhost:18081"
	}
	return &Handler{
		repo:         repo,
		prefsRepo:    prefsRepo,
		byokRepo:     byokRepo,
		redisClient:  redisClient,
		aiServiceURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type CatalogModel struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"display_name"`
	Provider          string   `json:"provider"`
	ProviderLabel     string   `json:"provider_label,omitempty"`
	Tier              string   `json:"tier,omitempty"`
	ContextWindow     int      `json:"context_window,omitempty"`
	CostHint          string   `json:"cost_hint,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"`
	ProviderAvailable *bool    `json:"provider_available,omitempty"`
	KeySource         string   `json:"key_source,omitempty"`
}

func (h *Handler) HandleGetCatalog(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	models, err := h.getCatalog(r.Context(), r.URL.Query())
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI model catalog")
		apierror.WriteError(w, apierror.NewInternal("Failed to retrieve model catalog"))
		return
	}

	prefs, err := h.prefsRepo.GetTenantAIPreferences(r.Context(), claims.TenantID)
	if err == nil {
		if len(prefs.EnabledProviders) > 0 {
			allowProv := make(map[string]struct{}, len(prefs.EnabledProviders))
			for _, p := range prefs.EnabledProviders {
				allowProv[p] = struct{}{}
			}
			filtered := make([]CatalogModel, 0, len(models))
			for _, model := range models {
				if _, ok := allowProv[model.Provider]; ok {
					filtered = append(filtered, model)
				}
			}
			models = filtered
		}
		if len(prefs.EnabledModels) > 0 {
			allow := make(map[string]struct{}, len(prefs.EnabledModels))
			for _, item := range prefs.EnabledModels {
				allow[item.Provider+":"+item.ModelID] = struct{}{}
			}
			filtered := make([]CatalogModel, 0, len(models))
			for _, model := range models {
				if _, ok := allow[model.Provider+":"+model.ID]; ok {
					filtered = append(filtered, model)
				}
			}
			models = filtered
		}
	}

	// Annotate models with BYOK key source
	if h.byokRepo != nil {
		keys, err := h.byokRepo.ListByTenant(r.Context(), claims.TenantID)
		if err == nil {
			providerKeySource := make(map[string]string)
			for _, k := range keys {
				if k.Status == "active" {
					switch k.Provider {
					case "mimo-token-plan":
						providerKeySource["mimo"] = "token-plan"
					case "minimax-token-plan":
						providerKeySource["minimax"] = "token-plan"
					default:
						if _, exists := providerKeySource[k.Provider]; !exists {
							providerKeySource[k.Provider] = "byok"
						}
					}
				}
			}
			providerLabels := map[string]string{
				"mimo":    "MiMo",
				"minimax": "MiniMax",
			}
			for i := range models {
				prov := models[i].Provider
				cost := models[i].CostHint
				if src, ok := providerKeySource[prov]; ok {
					models[i].KeySource = src
					label := providerLabels[prov]
					if label == "" {
						label = prov
					}
					if src == "token-plan" {
						models[i].ProviderLabel = label + " Token Plan"
					} else if src == "byok" {
						models[i].ProviderLabel = label + " (API Key)"
					}
				} else if prov == "openrouter" && cost == "free" {
					models[i].ProviderLabel = "Free (OpenRouter)"
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"models": models})
}

func (h *Handler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	prefs, err := h.prefsRepo.GetTenantAIPreferences(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load preferences"))
		return
	}
	prefs.Defaults = modelprofiles.EffectiveDefaults(prefs.Profile, prefs.Defaults)
	writeJSON(w, http.StatusOK, prefs)
}

func (h *Handler) HandlePutPreferences(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	if !h.isTenantAdmin(r.Context(), claims.TenantID, claims.UserID) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var req storage.TenantAIPreferencesUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Profile == "" {
		req.Profile = "balanced"
	}
	if req.RoutingStrategy == "" {
		req.RoutingStrategy = "quality_first"
	}
	if req.Defaults == nil {
		req.Defaults = map[string]storage.ModelSelection{}
	}
	if req.EnabledModels == nil {
		req.EnabledModels = []storage.ModelSelection{}
	}
	if req.Profile != "" && req.Profile != "custom" {
		for feature, selection := range modelprofiles.Expand(req.Profile) {
			if _, ok := req.Defaults[feature]; !ok {
				req.Defaults[feature] = selection
			}
		}
	}

	updated, err := h.prefsRepo.UpsertTenantAIPreferences(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		logrus.WithError(err).Error("Failed to upsert tenant ai preferences")
		apierror.WriteError(w, apierror.NewInternal("Failed to save preferences"))
		return
	}

	_ = h.repo.LogAuditEvent(r.Context(), &storage.AuditEvent{
		ID:         uuid.New(),
		ActorUserID: &claims.UserID,
		TenantID:   &claims.TenantID,
		Action:     "tenant.ai_preferences.updated",
		ResourceType: "tenant_ai_preferences",
		ResourceID: &claims.TenantID,
		AfterState: map[string]interface{}{
			"profile": req.Profile,
		},
		Timestamp: time.Now(),
		Success:   true,
	})

	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) HandleRefreshCatalog(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	if !h.isTenantAdmin(r.Context(), claims.TenantID, claims.UserID) {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}
	if h.redisClient != nil {
		_ = h.redisClient.Del(r.Context(), catalogCacheKey).Err()
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Catalog cache refreshed"})
}

func (h *Handler) getCatalog(ctx context.Context, query url.Values) ([]CatalogModel, error) {
	if h.redisClient != nil {
		if cached, err := h.redisClient.Get(ctx, catalogCacheKey).Result(); err == nil {
			var models []CatalogModel
			if json.Unmarshal([]byte(cached), &models) == nil {
				return filterCatalog(models, query), nil
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.aiServiceURL+"/api/models/catalog", nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("catalog request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("catalog request failed with status %d", resp.StatusCode)
	}

	var payload struct {
		Models []CatalogModel `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if h.redisClient != nil {
		if encoded, err := json.Marshal(payload.Models); err == nil {
			_ = h.redisClient.Set(ctx, catalogCacheKey, encoded, 30*time.Minute).Err()
		}
	}
	return filterCatalog(payload.Models, query), nil
}

func filterCatalog(models []CatalogModel, query url.Values) []CatalogModel {
	capability := strings.TrimSpace(query.Get("capability"))
	if capability == "" {
		return models
	}
	filtered := make([]CatalogModel, 0, len(models))
	for _, model := range models {
		for _, c := range model.Capabilities {
			if strings.EqualFold(c, capability) {
				filtered = append(filtered, model)
				break
			}
		}
	}
	return filtered
}

func (h *Handler) isTenantAdmin(ctx context.Context, tenantID, userID uuid.UUID) bool {
	membership, err := h.repo.GetMembership(ctx, tenantID, userID)
	if err != nil || membership == nil {
		return false
	}
	return membership.Role == "owner" || membership.Role == "admin"
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// HandleCheckModel tests if a model is still available on its provider.
// POST /v1/ai/models/check  { "provider": "openai", "model_id": "gpt-4o" }
func (h *Handler) HandleCheckModel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req struct {
		Provider string `json:"provider"`
		ModelID  string `json:"model_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Provider == "" || req.ModelID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("provider and model_id are required"))
		return
	}

	ctx := r.Context()
	start := time.Now()

	// Try the AI service's model check endpoint
	checkURL := fmt.Sprintf("%s/api/models/check", h.aiServiceURL)
	body, _ := json.Marshal(map[string]string{
		"provider":  req.Provider,
		"model_id":  req.ModelID,
		"tenant_id": claims.TenantID.String(),
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", checkURL, strings.NewReader(string(body)))
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create check request"))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		// AI service unreachable — fall back to provider-direct check
		result := h.checkModelDirect(ctx, req.Provider, req.ModelID)
		writeJSON(w, http.StatusOK, result)
		return
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Milliseconds()

	if resp.StatusCode >= 300 {
		result := h.checkModelDirect(ctx, req.Provider, req.ModelID)
		writeJSON(w, http.StatusOK, result)
		return
	}

	var aiResp struct {
		Available bool   `json:"available"`
		Message   string `json:"message"`
		Deprecated bool  `json:"deprecated"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		result := h.checkModelDirect(ctx, req.Provider, req.ModelID)
		writeJSON(w, http.StatusOK, result)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"available":   aiResp.Available,
		"deprecated":  aiResp.Deprecated,
		"message":     aiResp.Message,
		"latency_ms":  elapsed,
		"provider":    req.Provider,
		"model_id":    req.ModelID,
	})
}

// checkModelDirect tests a model by making a minimal API call directly to the provider.
func (h *Handler) checkModelDirect(ctx context.Context, provider, modelID string) map[string]interface{} {
	start := time.Now()

	endpoint, headers, body := modelCheckConfig(provider, modelID)
	if endpoint == "" {
		return map[string]interface{}{
			"available":  false,
			"deprecated": true,
			"message":    fmt.Sprintf("No check endpoint for provider '%s'", provider),
			"provider":   provider,
			"model_id":   modelID,
		}
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(body))
	if err != nil {
		return map[string]interface{}{
			"available": false,
			"message":   fmt.Sprintf("Request error: %v", err),
			"provider":  provider,
			"model_id":  modelID,
		}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"available": false,
			"message":   fmt.Sprintf("Connection failed: %v", err),
			"provider":  provider,
			"model_id":  modelID,
		}
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Milliseconds()

	respBody, _ := io.ReadAll(resp.Body)
	respStr := string(respBody)

	// Check for model-not-found / deprecated signals
	deprecated := false
	available := resp.StatusCode >= 200 && resp.StatusCode < 300

	if !available {
		lower := strings.ToLower(respStr)
		if strings.Contains(lower, "model_not_found") ||
			strings.Contains(lower, "model not found") ||
			strings.Contains(lower, "does not exist") ||
			strings.Contains(lower, "deprecated") ||
			strings.Contains(lower, "invalid_model") {
			deprecated = true
		}
	}

	// Extract error message if failed
	msg := ""
	if !available {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
			BaseResp *struct {
				StatusMsg string `json:"status_msg"`
			} `json:"base_resp"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			msg = errResp.Error.Message
			if msg == "" && errResp.BaseResp != nil {
				msg = errResp.BaseResp.StatusMsg
			}
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	return map[string]interface{}{
		"available":   available,
		"deprecated":  deprecated,
		"message":     msg,
		"latency_ms":  elapsed,
		"provider":    provider,
		"model_id":    modelID,
	}
}

// modelCheckConfig returns the endpoint, headers, and body for testing a model.
func modelCheckConfig(provider, modelID string) (endpoint string, headers map[string]string, body string) {
	headers = map[string]string{"Content-Type": "application/json"}

	switch provider {
	case "openrouter":
		endpoint = "https://openrouter.ai/api/v1/chat/completions"
		headers["Authorization"] = "Bearer " + os.Getenv("OPENROUTER_API_KEY")
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	case "openai":
		endpoint = "https://api.openai.com/v1/chat/completions"
		headers["Authorization"] = "Bearer " + os.Getenv("OPENAI_API_KEY")
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	case "anthropic":
		endpoint = "https://api.anthropic.com/v1/messages"
		headers["x-api-key"] = os.Getenv("ANTHROPIC_API_KEY")
		headers["anthropic-version"] = "2023-06-01"
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	case "groq":
		endpoint = "https://api.groq.com/openai/v1/chat/completions"
		headers["Authorization"] = "Bearer " + os.Getenv("GROQ_API_KEY")
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	case "mimo":
		endpoint = "https://api.mimo.ai/v1/chat/completions"
		headers["Authorization"] = "Bearer " + os.Getenv("MIMO_API_KEY")
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	case "minimax":
		endpoint = "https://api.minimaxi.com/v1/chat/completions"
		headers["Authorization"] = "Bearer " + os.Getenv("MINIMAX_API_KEY")
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	case "stepfun":
		endpoint = "https://api.stepfun.com/v1/chat/completions"
		headers["Authorization"] = "Bearer " + os.Getenv("STEPFUN_API_KEY")
		body = fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"ping"}],"max_tokens":1}`, modelID)
	default:
		return "", nil, ""
	}
	return endpoint, headers, body
}
