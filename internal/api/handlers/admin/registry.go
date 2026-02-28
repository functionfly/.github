package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RegistryHandler handles admin registry API (stats, list, get, update, delete, visibility, pricing).
// Uses the same registry repo as the public registry handler.
type RegistryHandler struct {
	registryRepo *registry.RegistryRepository
}

// NewRegistryHandler creates a new admin registry handler.
func NewRegistryHandler(registryRepo *registry.RegistryRepository) *RegistryHandler {
	return &RegistryHandler{registryRepo: registryRepo}
}

// HandleGetRegistryStats returns GET /v1/admin/registry/stats
func (h *RegistryHandler) HandleGetRegistryStats(w http.ResponseWriter, r *http.Request) {
	total, byVisibility, err := h.registryRepo.GetRegistryStats()
	if err != nil {
		logrus.WithError(err).Error("Failed to get registry stats")
		http.Error(w, "Failed to get registry stats", http.StatusInternalServerError)
		return
	}
	publicCount := byVisibility["public"]
	privateCount := byVisibility["private"]
	unlistedCount := byVisibility["unlisted"]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_functions":    total,
		"public_functions":   publicCount,
		"private_functions":  privateCount,
		"unlisted_functions": unlistedCount,
		"flagged_functions":  0,
		"total_calls":        0,
		"total_revenue":      0,
		"avg_rating":         0,
	})
}

// HandleListRegistryFunctions returns GET /v1/admin/registry/functions
func (h *RegistryHandler) HandleListRegistryFunctions(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0
	visibility := r.URL.Query().Get("visibility")
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	functions, total, err := h.registryRepo.ListFunctionsForAdmin(visibility, category, search, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list registry functions (admin)")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	// Build response in dashboard shape (id, author, name, visibility, price_per_call, etc.)
	out := make([]map[string]interface{}, 0, len(functions))
	for _, fn := range functions {
		latestVersion := ""
		if fn.LatestVersion.Valid {
			latestVersion = fn.LatestVersion.String
		}
		title := ""
		if fn.Title.Valid {
			title = fn.Title.String
		}
		desc := ""
		if fn.Description.Valid {
			desc = fn.Description.String
		}
		cat := ""
		if fn.Category.Valid {
			cat = fn.Category.String
		}
		overallScore := 0.0
		totalRatings := 0
		if fn.Rating != nil {
			overallScore = fn.Rating.OverallScore
			totalRatings = fn.Rating.TotalRatings
		}
		out = append(out, map[string]interface{}{
			"id":                  fn.ID.String(),
			"author":              fn.Author,
			"name":                fn.Name,
			"title":               title,
			"description":         desc,
			"category":            cat,
			"visibility":          fn.Visibility,
			"price_per_call":      fn.PricePerCall,
			"popularity_score":    fn.PopularityScore,
			"reliability_score":   fn.ReliabilityScore,
			"deterministic_score": fn.DeterministicScore,
			"latest_version":      latestVersion,
			"total_ratings":       totalRatings,
			"overall_score":       overallScore,
			"created_at":          fn.CreatedAt,
			"updated_at":          fn.UpdatedAt,
			"is_flagged":          false,
			"flag_reason":         nil,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"functions": out,
		"total":     total,
	})
}

// HandleGetRegistryFunction returns GET /v1/admin/registry/functions/{functionId}
func (h *RegistryHandler) HandleGetRegistryFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to get registry function (admin)")
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	versions, _ := h.registryRepo.ListFunctionVersions(fn.ID)
	latestVersion := ""
	if fn.LatestVersion.Valid {
		latestVersion = fn.LatestVersion.String
	}
	title := ""
	if fn.Title.Valid {
		title = fn.Title.String
	}
	desc := ""
	if fn.Description.Valid {
		desc = fn.Description.String
	}
	cat := ""
	if fn.Category.Valid {
		cat = fn.Category.String
	}
	overallScore := 0.0
	totalRatings := 0
	if fn.Rating != nil {
		overallScore = fn.Rating.OverallScore
		totalRatings = fn.Rating.TotalRatings
	}

	versionMaps := make([]map[string]interface{}, 0, len(versions))
	for _, v := range versions {
		versionMaps = append(versionMaps, map[string]interface{}{
			"id":            v.ID.String(),
			"function_id":   v.FunctionID.String(),
			"version":       v.Version,
			"runtime":       v.Runtime,
			"timeout_ms":    v.TimeoutMs,
			"memory_mb":     v.MemoryMB,
			"deterministic": v.Deterministic,
			"cache_ttl":     v.CacheTTL,
			"published_at":  v.PublishedAt,
			"is_active":     true,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"function": map[string]interface{}{
			"id":                  fn.ID.String(),
			"author":              fn.Author,
			"name":                fn.Name,
			"title":               title,
			"description":         desc,
			"category":            cat,
			"visibility":          fn.Visibility,
			"price_per_call":      fn.PricePerCall,
			"popularity_score":    fn.PopularityScore,
			"reliability_score":   fn.ReliabilityScore,
			"deterministic_score": fn.DeterministicScore,
			"latest_version":      latestVersion,
			"total_ratings":       totalRatings,
			"overall_score":       overallScore,
			"created_at":          fn.CreatedAt,
			"updated_at":          fn.UpdatedAt,
			"is_flagged":          false,
			"flag_reason":         nil,
		},
		"versions": versionMaps,
	})
}

// HandleUpdateRegistryFunction returns PATCH /v1/admin/registry/functions/{functionId}
func (h *RegistryHandler) HandleUpdateRegistryFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updated, err := h.registryRepo.UpdateRegistryFunction(functionID, updates)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update registry function (admin)")
		http.Error(w, "Failed to update function", http.StatusInternalServerError)
		return
	}

	// Return same shape as get
	latestVersion := ""
	if updated.LatestVersion.Valid {
		latestVersion = updated.LatestVersion.String
	}
	title := ""
	if updated.Title.Valid {
		title = updated.Title.String
	}
	desc := ""
	if updated.Description.Valid {
		desc = updated.Description.String
	}
	cat := ""
	if updated.Category.Valid {
		cat = updated.Category.String
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                updated.ID.String(),
		"author":            updated.Author,
		"name":              updated.Name,
		"title":             title,
		"description":       desc,
		"category":          cat,
		"visibility":        updated.Visibility,
		"price_per_call":    updated.PricePerCall,
		"popularity_score":  updated.PopularityScore,
		"reliability_score": updated.ReliabilityScore,
		"latest_version":    latestVersion,
		"created_at":        updated.CreatedAt,
		"updated_at":        updated.UpdatedAt,
		"is_flagged":        false,
		"flag_reason":       nil,
	})
}

// HandleDeleteRegistryFunction returns DELETE /v1/admin/registry/functions/{functionId}
func (h *RegistryHandler) HandleDeleteRegistryFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	if err := h.registryRepo.DeleteFunction(fn.Author, fn.Name); err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to delete registry function (admin)")
		http.Error(w, "Failed to delete function", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "deleted"})
}

// HandleUpdateRegistryVisibility returns PATCH /v1/admin/registry/functions/{functionId}/visibility
func (h *RegistryHandler) HandleUpdateRegistryVisibility(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Visibility != "public" && body.Visibility != "private" && body.Visibility != "unlisted" {
		http.Error(w, "visibility must be public, private, or unlisted", http.StatusBadRequest)
		return
	}

	updated, err := h.registryRepo.UpdateRegistryFunction(functionID, map[string]interface{}{"visibility": body.Visibility})
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update visibility (admin)")
		http.Error(w, "Failed to update visibility", http.StatusInternalServerError)
		return
	}

	latestVersion := ""
	if updated.LatestVersion.Valid {
		latestVersion = updated.LatestVersion.String
	}
	title := ""
	if updated.Title.Valid {
		title = updated.Title.String
	}
	desc := ""
	if updated.Description.Valid {
		desc = updated.Description.String
	}
	cat := ""
	if updated.Category.Valid {
		cat = updated.Category.String
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             updated.ID.String(),
		"author":         updated.Author,
		"name":           updated.Name,
		"title":          title,
		"description":    desc,
		"category":       cat,
		"visibility":     updated.Visibility,
		"price_per_call": updated.PricePerCall,
		"latest_version": latestVersion,
		"created_at":     updated.CreatedAt,
		"updated_at":     updated.UpdatedAt,
		"is_flagged":     false,
	})
}

// HandleUpdateRegistryPricing returns PATCH /v1/admin/registry/functions/{functionId}/pricing
func (h *RegistryHandler) HandleUpdateRegistryPricing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var body struct {
		PricePerCall float64 `json:"price_per_call"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updated, err := h.registryRepo.UpdateRegistryFunction(functionID, map[string]interface{}{"price_per_call": body.PricePerCall})
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update pricing (admin)")
		http.Error(w, "Failed to update pricing", http.StatusInternalServerError)
		return
	}

	latestVersion := ""
	if updated.LatestVersion.Valid {
		latestVersion = updated.LatestVersion.String
	}
	title := ""
	if updated.Title.Valid {
		title = updated.Title.String
	}
	desc := ""
	if updated.Description.Valid {
		desc = updated.Description.String
	}
	cat := ""
	if updated.Category.Valid {
		cat = updated.Category.String
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             updated.ID.String(),
		"author":         updated.Author,
		"name":           updated.Name,
		"title":          title,
		"description":    desc,
		"category":       cat,
		"visibility":     updated.Visibility,
		"price_per_call": updated.PricePerCall,
		"latest_version": latestVersion,
		"created_at":     updated.CreatedAt,
		"updated_at":     updated.UpdatedAt,
		"is_flagged":     false,
	})
}

// HandleFlagRegistryFunction returns POST /v1/admin/registry/functions/{functionId}/flag (stub: no flagged column yet)
func (h *RegistryHandler) HandleFlagRegistryFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Stub: flag storage not implemented yet; return current function
	latestVersion := ""
	if fn.LatestVersion.Valid {
		latestVersion = fn.LatestVersion.String
	}
	title := ""
	if fn.Title.Valid {
		title = fn.Title.String
	}
	desc := ""
	if fn.Description.Valid {
		desc = fn.Description.String
	}
	cat := ""
	if fn.Category.Valid {
		cat = fn.Category.String
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             fn.ID.String(),
		"author":         fn.Author,
		"name":           fn.Name,
		"title":          title,
		"description":    desc,
		"category":       cat,
		"visibility":     fn.Visibility,
		"price_per_call": fn.PricePerCall,
		"latest_version": latestVersion,
		"created_at":     fn.CreatedAt,
		"updated_at":     fn.UpdatedAt,
		"is_flagged":     false,
	})
}

// HandleListRegistryFunctionVersions returns GET /v1/admin/registry/functions/{functionId}/versions
func (h *RegistryHandler) HandleListRegistryFunctionVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	versions, err := h.registryRepo.ListFunctionVersions(functionID)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to list versions (admin)")
		http.Error(w, "Failed to list versions", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(versions))
	for _, v := range versions {
		out = append(out, map[string]interface{}{
			"id":            v.ID.String(),
			"function_id":   v.FunctionID.String(),
			"version":       v.Version,
			"runtime":       v.Runtime,
			"timeout_ms":    v.TimeoutMs,
			"memory_mb":     v.MemoryMB,
			"deterministic": v.Deterministic,
			"cache_ttl":     v.CacheTTL,
			"published_at":  v.PublishedAt,
			"is_active":     true,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"versions": out})
}

// HandleDeactivateRegistryVersion returns POST /v1/admin/registry/functions/{functionId}/versions/{versionId}/deactivate (stub)
func (h *RegistryHandler) HandleDeactivateRegistryVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, _ = uuid.Parse(vars["functionId"])
	versionID, err := uuid.Parse(vars["versionId"])
	if err != nil {
		http.Error(w, "Invalid version ID", http.StatusBadRequest)
		return
	}

	// Stub: no is_active on versions in schema; return 200 with placeholder
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           versionID.String(),
		"is_active":    false,
		"version":      "",
		"published_at": nil,
	})
}

// HandleGetRegistryFunctionMetrics returns GET /v1/admin/registry/functions/{functionId}/metrics (stub)
func (h *RegistryHandler) HandleGetRegistryFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invocations":    0,
		"errors":         0,
		"latency_p50_ms": 0,
		"latency_p99_ms": 0,
	})
}

const openRouterFreeModel = "arcee-ai/trinity-large-preview:free"
const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

// HandleGenerateRegistryDescription returns POST /v1/admin/registry/generate-description
// Uses Open Router free models to generate a short description from function name/title/category.
func (h *RegistryHandler) HandleGenerateRegistryDescription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		http.Error(w, "Open Router API key not configured (OPENROUTER_API_KEY)", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Name     string `json:"name"`
		Title    string `json:"title"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	prompt := "Write a short, clear one or two sentence description for a registry function."
	if body.Name != "" {
		prompt += " Function name: " + body.Name + "."
	}
	if body.Title != "" {
		prompt += " Display title: " + body.Title + "."
	}
	if body.Category != "" {
		prompt += " Category: " + body.Category + "."
	}
	prompt += " Output only the description text, no quotes or prefix."

	reqBody := map[string]interface{}{
		"model": openRouterFreeModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 150,
	}
	encoded, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(r.Context(), "POST", openRouterURL, bytes.NewReader(encoded))
	if err != nil {
		logrus.WithError(err).Error("Failed to create Open Router request")
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Open Router request failed")
		http.Error(w, "Open Router request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("status", resp.StatusCode).Error("Open Router returned non-200")
		http.Error(w, "Open Router returned error", http.StatusBadGateway)
		return
	}

	var openResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openResp); err != nil {
		logrus.WithError(err).Error("Failed to decode Open Router response")
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}
	description := ""
	if len(openResp.Choices) > 0 {
		description = strings.TrimSpace(openResp.Choices[0].Message.Content)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"description": description})
}
