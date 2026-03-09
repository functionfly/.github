package categorization

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/agent/categorization"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handler handles categorization API requests
type Handler struct {
	db  *gorm.DB
	svc *categorization.Service
}

// NewHandler creates a new categorization handler
func NewHandler(db *gorm.DB, svc *categorization.Service) *Handler {
	return &Handler{db: db, svc: svc}
}

// HandleGetTaxonomy returns the full category and tag taxonomy
func (h *Handler) HandleGetTaxonomy(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"categories":      categorization.CategoryTaxonomy,
		"tags":            categorization.TagTaxonomy,
		"root_categories": categorization.GetRootCategories(),
		"auto_apply_tags": categorization.GetAutoApplyTags(),
	}
	writeJSON(w, http.StatusOK, response)
}

// HandleGetCategories returns all categories
func (h *Handler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	parentID := r.URL.Query().Get("parent")

	var categories []categorization.Category
	if parentID != "" {
		categories = categorization.GetSubCategories(parentID)
	} else {
		categories = categorization.CategoryTaxonomy
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"categories": categories,
		"total":      len(categories),
	})
}

// HandleGetCategory returns a specific category by ID
func (h *Handler) HandleGetCategory(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/categorization/categories/")
	if id == "" || id == "/" {
		writeError(w, http.StatusBadRequest, "category ID required")
		return
	}

	category := categorization.GetCategoryByID(id)
	if category == nil {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	// Get sub-categories
	subCategories := categorization.GetSubCategories(id)

	writeJSON(w, http.StatusOK, map[string]any{
		"category":       category,
		"sub_categories": subCategories,
	})
}

// HandleGetTags returns all tags
func (h *Handler) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	autoApply := r.URL.Query().Get("auto_apply") == "true"

	var tags []categorization.Tag
	if category != "" {
		tags = categorization.GetTagsByCategory(category)
	} else if autoApply {
		tags = categorization.GetAutoApplyTags()
	} else {
		tags = categorization.TagTaxonomy
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tags":  tags,
		"total": len(tags),
	})
}

// CategorizeRequest is the request body for categorizing a function
type CategorizeRequest struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	InputSchema  map[string]any `json:"input_schema"`
	OutputSchema map[string]any `json:"output_schema"`
	Code         string         `json:"code"`
	Runtime      string         `json:"runtime"`
}

// HandleCategorize categorizes a function spec without storing
func (h *Handler) HandleCategorize(w http.ResponseWriter, r *http.Request) {
	var req CategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "function name is required")
		return
	}

	spec := &categorization.FunctionSpec{
		Name:         req.Name,
		Description:  req.Description,
		InputSchema:  req.InputSchema,
		OutputSchema: req.OutputSchema,
		Code:         req.Code,
		Runtime:      req.Runtime,
	}

	result, err := h.svc.CategorizeFunction(r.Context(), spec)
	if err != nil {
		logrus.WithError(err).Error("failed to categorize function")
		writeError(w, http.StatusInternalServerError, "failed to categorize function")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"categorization": result,
	})
}

// HandleGetFunctionCategory returns the categorization for a specific function
func (h *Handler) HandleGetFunctionCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/categorization/functions/")
	if idStr == "" || strings.Contains(idStr, "/") {
		writeError(w, http.StatusBadRequest, "function ID required")
		return
	}

	functionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	fc, err := h.svc.GetCategorization(r.Context(), functionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "categorization not found")
			return
		}
		logrus.WithError(err).Error("failed to get categorization")
		writeError(w, http.StatusInternalServerError, "failed to get categorization")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"categorization": fc,
	})
}

// UpdateCategorizationRequest is the request body for updating categorization
type UpdateCategorizationRequest struct {
	PrimaryCategory   string   `json:"primary_category"`
	SecondaryCategory string   `json:"secondary_category"`
	Tags              []string `json:"tags"`
}

// HandleUpdateFunctionCategory manually updates a function's categorization
func (h *Handler) HandleUpdateFunctionCategory(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	if middleware.GetUserFromContext(r) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/categorization/functions/")
	if idStr == "" || strings.Contains(idStr, "/") {
		writeError(w, http.StatusBadRequest, "function ID required")
		return
	}

	functionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	var req UpdateCategorizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PrimaryCategory == "" {
		writeError(w, http.StatusBadRequest, "primary_category is required")
		return
	}

	// Validate category exists
	if cat := categorization.GetCategoryByID(req.PrimaryCategory); cat == nil {
		writeError(w, http.StatusBadRequest, "invalid primary category")
		return
	}

	fc, err := h.svc.UpdateCategorization(r.Context(), functionID, req.PrimaryCategory, req.SecondaryCategory, req.Tags)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "categorization not found")
			return
		}
		logrus.WithError(err).Error("failed to update categorization")
		writeError(w, http.StatusInternalServerError, "failed to update categorization")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"categorization": fc,
		"message":        "categorization updated successfully",
	})
}

// HandleGetFunctionsByCategory returns functions in a category
func (h *Handler) HandleGetFunctionsByCategory(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimPrefix(r.URL.Path, "/categorization/category/")
	if category == "" || category == "/" {
		writeError(w, http.StatusBadRequest, "category required")
		return
	}

	// Remove trailing slash
	category = strings.TrimSuffix(category, "/")

	limit, offset := parseLimitOffset(r, 20, 100)

	fcs, total, err := h.svc.GetFunctionsByCategory(r.Context(), category, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("failed to get functions by category")
		writeError(w, http.StatusInternalServerError, "failed to get functions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"functions": fcs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"category":  category,
	})
}

// HandleGetFunctionsByTag returns functions with a specific tag
func (h *Handler) HandleGetFunctionsByTag(w http.ResponseWriter, r *http.Request) {
	tag := strings.TrimPrefix(r.URL.Path, "/categorization/tag/")
	if tag == "" || tag == "/" {
		writeError(w, http.StatusBadRequest, "tag required")
		return
	}

	// Remove trailing slash
	tag = strings.TrimSuffix(tag, "/")

	limit, offset := parseLimitOffset(r, 20, 100)

	fcs, total, err := h.svc.GetFunctionsByTag(r.Context(), tag, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("failed to get functions by tag")
		writeError(w, http.StatusInternalServerError, "failed to get functions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"functions": fcs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"tag":       tag,
	})
}

// HandleAnalyzeCode analyzes code and returns detected patterns
func (h *Handler) HandleAnalyzeCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	// Create a temp service to analyze patterns
	svc := categorization.NewService(nil)
	spec := &categorization.FunctionSpec{Code: req.Code}
	patterns := svc.AnalyzeCodePatterns(req.Code)

	// Also get suggested categories
	result, _ := svc.CategorizeFunction(r.Context(), spec)

	writeJSON(w, http.StatusOK, map[string]any{
		"patterns":           patterns,
		"suggested_category": result.PrimaryCategory,
		"suggested_tags":     result.Tags,
		"confidence":         result.Confidence,
	})
}

// HandleReCategorize triggers re-categorization of a function
func (h *Handler) HandleReCategorize(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	if middleware.GetUserFromContext(r) == nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/categorization/functions/")
	idStr = strings.TrimSuffix(idStr, "/recategorize")

	functionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid function ID")
		return
	}

	// Get function details from registry
	var funcData struct {
		ID          string
		Name        string
		Description string
		Category    string
		Tags        []string
	}

	err = h.db.WithContext(r.Context()).Table("registry_functions").
		Select("id, name, description, category, tags").
		Where("id = ?", functionID).
		First(&funcData).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, http.StatusNotFound, "function not found")
			return
		}
		logrus.WithError(err).Error("failed to get function")
		writeError(w, http.StatusInternalServerError, "failed to get function")
		return
	}

	// Get generated code if available
	var code string
	h.db.WithContext(r.Context()).Table("generated_code").
		Select("generated_code").
		Where("id = ?", functionID).
		Scan(&code)

	spec := &categorization.FunctionSpec{
		Name:        funcData.Name,
		Description: funcData.Description,
		Code:        code,
	}

	fc, err := h.svc.CategorizeAndStore(r.Context(), functionID, spec)
	if err != nil {
		logrus.WithError(err).Error("failed to re-categorize function")
		writeError(w, http.StatusInternalServerError, "failed to re-categorize function")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"categorization": fc,
		"message":        "function re-categorized successfully",
	})
}

// Helper functions

func parseLimitOffset(r *http.Request, defaultLimit, maxLimit int) (int, int) {
	limit := defaultLimit
	offset := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}
