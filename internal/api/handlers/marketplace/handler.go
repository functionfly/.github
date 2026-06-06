package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/security"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo HandlerRepo
}

type HandlerRepo interface {
	List(ctx context.Context, params ListParams) ([]Extension, error)
	Get(ctx context.Context, id string) (*Extension, error)
	Create(ctx context.Context, ext *Extension) error
	Update(ctx context.Context, ext *Extension) error
	Delete(ctx context.Context, id string) error
	IncrementInstallCount(ctx context.Context, id string) error
	GetInstallCounts(ctx context.Context, ids []string) (map[string]int, error)
	GetCategories(ctx context.Context) ([]CategoryCount, error)
	UpsertRating(ctx context.Context, rating *Rating) error
	GetRating(ctx context.Context, extensionID, tenantID string) (*Rating, error)
	ListRatings(ctx context.Context, extensionID string, limit int) ([]Rating, error)
	FindUpdates(ctx context.Context, installed []InstalledPlugin) ([]ExtensionUpdate, error)
	CreatePluginFromExtension(ctx context.Context, tenantID, extensionID string) (*Extension, error)
}

func NewHandler(repo HandlerRepo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleListExtensions(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	featured := r.URL.Query().Get("featured") == "true"
	tagsParam := r.URL.Query().Get("tags")
	sortBy := r.URL.Query().Get("sort")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	var tags []string
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
	}

	params := ListParams{
		Category: categoryPtr,
		Status:   statusPtr,
		Search:   searchPtr,
		Featured: &featured,
		Tags:     tags,
		SortBy:   sortBy,
		Limit:    limit,
		Offset:   offset,
	}

	extensions, err := h.repo.List(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Warn("marketplace: failed to list extensions")
		writeJSON(w, http.StatusOK, map[string]interface{}{"extensions": []Extension{}})
		return
	}

	if extensions == nil {
		extensions = []Extension{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extensions": extensions})
}

func (h *Handler) HandleGetExtension(w http.ResponseWriter, r *http.Request) {
	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extension": ext})
}

func (h *Handler) HandleCreateExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateExtensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateExtensionRequest(req.Name, req.Version, req.Manifest); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	ext := &Extension{
		CreatorID:     tenantID,
		Name:          req.Name,
		Version:       req.Version,
		Description:   req.Description,
		Category:      req.Category,
		IconURL:       req.IconURL,
		Manifest:      req.Manifest,
		Status:        "draft",
		Featured:      false,
		InstallCount:  0,
		RatingAverage: 0,
		RatingCount:   0,
		TrustScore:    0,
	}

	if err := h.repo.Create(r.Context(), ext); err != nil {
		logrus.WithError(err).Error("marketplace: failed to create extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create extension")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"extension": ext})
}

func (h *Handler) HandleUpdateExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	var req UpdateExtensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name != "" {
		ext.Name = req.Name
	}
	if req.Version != "" {
		ext.Version = req.Version
	}
	if req.Description != "" {
		ext.Description = req.Description
	}
	if req.Category != "" {
		ext.Category = req.Category
	}
	if req.IconURL != "" {
		ext.IconURL = req.IconURL
	}
	if req.Manifest != nil {
		ext.Manifest = req.Manifest
	}
	if req.Status != "" {
		ext.Status = req.Status
	}
	if req.Featured != nil {
		ext.Featured = *req.Featured
	}

	if err := h.repo.Update(r.Context(), ext); err != nil {
		logrus.WithError(err).Error("marketplace: failed to update extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"extension": ext})
}

func (h *Handler) HandleDeleteExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	if ext.CreatorID != tenantID {
		writeJSONError(w, http.StatusForbidden, "Not authorized to delete this extension")
		return
	}

	if err := h.repo.Delete(r.Context(), extID); err != nil {
		logrus.WithError(err).Error("marketplace: failed to delete extension")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete extension")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Extension deleted"})
}

func (h *Handler) HandleInstallExtension(w http.ResponseWriter, r *http.Request) {
	h.HandleInstallExtensionWithPlugin(w, r)
}

func (h *Handler) HandleGetInstallCounts(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		writeJSONError(w, http.StatusBadRequest, "ids parameter is required")
		return
	}

	ids := strings.Split(idsParam, ",")
	if len(ids) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"install_counts": map[string]int{}})
		return
	}

	counts, err := h.repo.GetInstallCounts(r.Context(), ids)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get install counts")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get install counts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"install_counts": counts})
}

func (h *Handler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetCategories(r.Context())
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get categories")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get categories")
		return
	}

	if categories == nil {
		categories = []CategoryCount{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"categories": categories})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getTenantID(r *http.Request) string {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		return ""
	}
	return claims.TenantID.String()
}

type CreateExtensionRequest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	IconURL     string                 `json:"icon_url"`
	Manifest    map[string]interface{} `json:"manifest"`
	Tags        []string               `json:"tags"`
}

type UpdateExtensionRequest struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	IconURL     string                 `json:"icon_url"`
	Manifest    map[string]interface{} `json:"manifest"`
	Status      string                 `json:"status"`
	Featured    *bool                  `json:"featured"`
	Tags        []string               `json:"tags"`
}

type Rating struct {
	ID          string    `json:"id"`
	ExtensionID string    `json:"extension_id"`
	TenantID    string    `json:"tenant_id"`
	Rating      int       `json:"rating"`
	Review      string    `json:"review,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RateRequest struct {
	Rating int    `json:"rating"`
	Review string `json:"review"`
}

type InstalledPlugin struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ExtensionUpdate struct {
	InstalledPluginID string                 `json:"installed_plugin_id"`
	InstalledVersion  string                 `json:"installed_version"`
	ExtensionID       string                 `json:"extension_id"`
	LatestVersion     string                 `json:"latest_version"`
	Changelog         string                 `json:"changelog"`
	Manifest          map[string]interface{} `json:"manifest"`
}

func (h *Handler) HandleRateExtension(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to rate extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		writeJSONError(w, http.StatusBadRequest, "rating must be between 1 and 5")
		return
	}

	rating := &Rating{
		ExtensionID: extID,
		TenantID:    tenantID,
		Rating:      req.Rating,
		Review:      req.Review,
	}
	if err := h.repo.UpsertRating(r.Context(), rating); err != nil {
		logrus.WithError(err).Error("marketplace: failed to upsert rating")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save rating")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Rating saved", "rating": rating})
}

func (h *Handler) HandleGetMyRating(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	rating, err := h.repo.GetRating(r.Context(), extID, tenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get rating")
		return
	}

	if rating == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"rating": nil})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"rating": rating})
}

func (h *Handler) HandleListRatings(w http.ResponseWriter, r *http.Request) {
	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	ratings, err := h.repo.ListRatings(r.Context(), extID, limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list ratings")
		return
	}

	if ratings == nil {
		ratings = []Rating{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"ratings": ratings})
}

func (h *Handler) HandleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var installed []InstalledPlugin
	if err := json.NewDecoder(r.Body).Decode(&installed); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updates, err := h.repo.FindUpdates(r.Context(), installed)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to find updates")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check for updates")
		return
	}

	if updates == nil {
		updates = []ExtensionUpdate{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"updates": updates})
}

func (h *Handler) HandleInstallExtensionWithPlugin(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	extID := mux.Vars(r)["id"]
	if extID == "" {
		writeJSONError(w, http.StatusBadRequest, "extension id is required")
		return
	}

	ext, err := h.repo.Get(r.Context(), extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to get extension for install")
		writeJSONError(w, http.StatusInternalServerError, "Failed to install extension")
		return
	}
	if ext == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	analyzer := security.NewPluginAnalyzer()
	manifest := &security.PluginManifest{}
	if err := mapToStruct(ext.Manifest, manifest); err == nil && manifest.Name != "" {
		if result, err := analyzer.AnalyzeManifest(r.Context(), manifest); err == nil {
			ext.TrustScore = result.Score
			ext.SecurityScore = result.Score
		}
	}

	created, err := h.repo.CreatePluginFromExtension(r.Context(), tenantID, extID)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to install extension as plugin")
		writeJSONError(w, http.StatusInternalServerError, "Failed to install extension")
		return
	}
	if created == nil {
		writeJSONError(w, http.StatusNotFound, "Extension not found")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":         "Extension installed",
		"extension":       created,
		"extension_id":    extID,
		"plugin_manifest": created.Manifest,
	})
}

func validateExtensionRequest(name, version string, manifest map[string]interface{}) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if version == "" {
		return fmt.Errorf("version is required")
	}
	if !isValidSemver(version) {
		return fmt.Errorf("version must be valid semver (e.g. 1.0.0)")
	}
	if len(name) > 255 {
		return fmt.Errorf("name must be 255 characters or less")
	}

	analyzer := security.NewPluginAnalyzer()
	parsed := &security.PluginManifest{}
	_ = mapToStruct(manifest, parsed)
	parsed.Name = name
	parsed.Version = version

	result, err := analyzer.AnalyzeManifest(context.Background(), parsed)
	if err != nil {
		return fmt.Errorf("manifest analysis failed: %w", err)
	}
	if !result.Safe {
		return fmt.Errorf("manifest failed security analysis (score: %.1f): %v", result.Score, result.Issues)
	}
	return nil
}

func isValidSemver(v string) bool {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

func mapToStruct(m map[string]interface{}, dst interface{}) error {
	if m == nil {
		return fmt.Errorf("nil map")
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
