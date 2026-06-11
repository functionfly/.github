package aimodels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/ai/modelprofiles"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const catalogCacheKey = "ai:catalog:v1"

type Handler struct {
	repo         storage.Repository
	prefsRepo    *storage.AIModelPreferencesRepository
	redisClient  *redis.Client
	aiServiceURL string
	httpClient   *http.Client
}

func NewHandler(
	repo storage.Repository,
	prefsRepo *storage.AIModelPreferencesRepository,
	redisClient *redis.Client,
	aiServiceURL string,
) *Handler {
	baseURL := strings.TrimRight(aiServiceURL, "/")
	if baseURL == "" && os.Getenv("DEVELOPMENT") == "true" {
		baseURL = "http://localhost:18081"
	}
	return &Handler{
		repo:         repo,
		prefsRepo:    prefsRepo,
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
	Tier              string   `json:"tier,omitempty"`
	ContextWindow     int      `json:"context_window,omitempty"`
	CostHint          string   `json:"cost_hint,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"`
	ProviderAvailable *bool    `json:"provider_available,omitempty"`
}

func (h *Handler) HandleGetCatalog(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	models, err := h.getCatalog(r.Context(), r.URL.Query())
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI model catalog")
		http.Error(w, "Failed to retrieve model catalog", http.StatusBadGateway)
		return
	}

	prefs, err := h.prefsRepo.GetTenantAIPreferences(r.Context(), claims.TenantID)
	if err == nil && len(prefs.EnabledModels) > 0 {
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

	writeJSON(w, http.StatusOK, map[string]interface{}{"models": models})
}

func (h *Handler) HandleGetPreferences(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	prefs, err := h.prefsRepo.GetTenantAIPreferences(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to load preferences", http.StatusInternalServerError)
		return
	}
	prefs.Defaults = modelprofiles.EffectiveDefaults(prefs.Profile, prefs.Defaults)
	writeJSON(w, http.StatusOK, prefs)
}

func (h *Handler) HandlePutPreferences(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.isTenantAdmin(r.Context(), claims.TenantID, claims.UserID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req storage.TenantAIPreferencesUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
		http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !h.isTenantAdmin(r.Context(), claims.TenantID, claims.UserID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
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
