package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
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
}

func NewHandler(repo HandlerRepo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleListExtensions(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	featured := r.URL.Query().Get("featured") == "true"
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

	params := ListParams{
		Category: categoryPtr,
		Status:  statusPtr,
		Search:  searchPtr,
		Featured: &featured,
		Limit:   limit,
		Offset:  offset,
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

	if err := h.repo.IncrementInstallCount(r.Context(), extID); err != nil {
		logrus.WithError(err).Warn("marketplace: failed to increment install count")
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Extension installed",
		"extension_id":  extID,
		"plugin_manifest": ext.Manifest,
	})
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
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Description string                  `json:"description"`
	Category    string                  `json:"category"`
	IconURL     string                  `json:"icon_url"`
	Manifest    map[string]interface{}  `json:"manifest"`
	Tags        []string               `json:"tags"`
}

type UpdateExtensionRequest struct {
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Description string                  `json:"description"`
	Category    string                  `json:"category"`
	IconURL     string                  `json:"icon_url"`
	Manifest    map[string]interface{}  `json:"manifest"`
	Status      string                  `json:"status"`
	Featured    *bool                   `json:"featured"`
	Tags        []string               `json:"tags"`
}