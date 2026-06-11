package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RegistryHandler handles admin registry API (stats, list, get, update, delete, visibility, pricing).
// Uses the same registry repo as the public registry handler.
type RegistryHandler struct {
	registryRepo *registryrepo.RegistryRepository
	cacheService *cache.CacheService
}

// NewRegistryHandler creates a new admin registry handler.
func NewRegistryHandler(registryRepo *registryrepo.RegistryRepository, cacheService *cache.CacheService) *RegistryHandler {
	return &RegistryHandler{
		registryRepo: registryRepo,
		cacheService: cacheService,
	}
}

// HandleGetRegistryStats returns GET /v1/admin/registry/stats
func (h *RegistryHandler) HandleGetRegistryStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.registryRepo.GetAdminRegistryStats()
	if err != nil {
		logrus.WithError(err).Error("Failed to get admin registry stats")
		apierror.WriteError(w, apierror.NewInternal("Failed to get registry stats"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_functions":    stats.TotalFunctions,
		"public_functions":   stats.PublicFunctions,
		"private_functions":  stats.PrivateFunctions,
		"unlisted_functions": stats.UnlistedFunctions,
		"flagged_functions":  stats.FlaggedFunctions,
		"total_calls":        stats.TotalCalls,
		"total_revenue":      stats.TotalRevenueUSD,
		"avg_rating":         stats.AvgRating,
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
		apierror.WriteError(w, apierror.NewInternal("Failed to list functions"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to get registry function (admin)")
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
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
	tags := fn.Tags
	capabilities := fn.Capabilities
	embedConfig := fn.EmbedConfig
	tenantID := ""
	if fn.TenantID != nil {
		tenantID = fn.TenantID.String()
	}
	ownerUserID := ""
	if fn.OwnerUserID != nil {
		ownerUserID = fn.OwnerUserID.String()
	}
	overallScore := 0.0
	totalRatings := 0
	if fn.Rating != nil {
		overallScore = fn.Rating.OverallScore
		totalRatings = fn.Rating.TotalRatings
	}

	versionMaps := make([]map[string]interface{}, 0, len(versions))
	for _, v := range versions {
		deploymentID := ""
		if v.DeploymentID != nil {
			deploymentID = v.DeploymentID.String()
		}
		backendID := ""
		if v.BackendID != nil {
			backendID = v.BackendID.String()
		}
		contentHash := ""
		if v.ContentHash.Valid {
			contentHash = v.ContentHash.String
		}
		sourceHash := ""
		if v.SourceHash.Valid {
			sourceHash = v.SourceHash.String
		}
		sourceCode := ""
		if v.SourceCode.Valid {
			sourceCode = v.SourceCode.String
		}
		bundleSize := 0
		if v.BundleSize.Valid {
			bundleSize = int(v.BundleSize.Int32)
		}
		versionMaps = append(versionMaps, map[string]interface{}{
			"id":            v.ID.String(),
			"function_id":   v.FunctionID.String(),
			"version":       v.Version,
			"manifest":      v.Manifest,
			"runtime":       v.Runtime,
			"timeout_ms":    v.TimeoutMs,
			"memory_mb":     v.MemoryMB,
			"deterministic": v.Deterministic,
			"cache_ttl":     v.CacheTTL,
			"capabilities":  v.Capabilities,
			"side_effects":  v.SideEffects,
			"idempotent":    v.Idempotent,
			"deployment_id": deploymentID,
			"backend_id":    backendID,
			"content_hash":  contentHash,
			"source_hash":   sourceHash,
			"source_code":   sourceCode,
			"bundle_size":   bundleSize,
			"published_at":  v.PublishedAt,
			"updated_at":    v.UpdatedAt,
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
			"tags":                tags,
			"visibility":          fn.Visibility,
			"price_per_call":      fn.PricePerCall,
			"popularity_score":    fn.PopularityScore,
			"reliability_score":   fn.ReliabilityScore,
			"deterministic_score": fn.DeterministicScore,
			"capabilities":        capabilities,
			"embed_config":        embedConfig,
			"tenant_id":           tenantID,
			"owner_user_id":       ownerUserID,
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	updated, err := h.registryRepo.UpdateRegistryFunction(functionID, updates)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update registry function (admin)")
		apierror.WriteError(w, apierror.NewInternal("Failed to update function"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if err := h.registryRepo.DeleteFunction(fn.Author, fn.Name); err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to delete registry function (admin)")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete function"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	var body struct {
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}
	if body.Visibility != "public" && body.Visibility != "private" && body.Visibility != "unlisted" {
		apierror.WriteError(w, apierror.NewBadRequest("visibility must be public, private, or unlisted"))
		return
	}

	updated, err := h.registryRepo.UpdateRegistryFunction(functionID, map[string]interface{}{"visibility": body.Visibility})
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update visibility (admin)")
		apierror.WriteError(w, apierror.NewInternal("Failed to update visibility"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	var body struct {
		PricePerCall float64 `json:"price_per_call"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	updated, err := h.registryRepo.UpdateRegistryFunction(functionID, map[string]interface{}{"price_per_call": body.PricePerCall})
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to update pricing (admin)")
		apierror.WriteError(w, apierror.NewInternal("Failed to update pricing"))
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

// HandleFlagRegistryFunction returns POST /v1/admin/registry/functions/{functionId}/flag
func (h *RegistryHandler) HandleFlagRegistryFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	var body struct {
		Reason  string `json:"reason"`
		Notes   string `json:"notes"`
		AdminID string `json:"admin_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}
	if body.Reason == "" {
		apierror.WriteError(w, apierror.NewBadRequest("reason is required (spam, malware, ip_infringement, abuse, policy_violation)"))
		return
	}

	flagReason := registryrepo.FlagFunctionFlags(body.Reason)
	var reviewerID uuid.UUID
	if body.AdminID != "" {
		reviewerID, _ = uuid.Parse(body.AdminID)
	}

	if err := h.registryRepo.FlagFunction(r.Context(), functionID, flagReason, reviewerID, body.Notes); err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to flag registry function")
		apierror.WriteError(w, apierror.NewInternal("Failed to flag function"))
		return
	}

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
		"is_flagged":     true,
		"flag_reason":    body.Reason,
	})
}

// HandleListRegistryFunctionVersions returns GET /v1/admin/registry/functions/{functionId}/versions
func (h *RegistryHandler) HandleListRegistryFunctionVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	versions, err := h.registryRepo.ListFunctionVersions(functionID)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to list versions (admin)")
		apierror.WriteError(w, apierror.NewInternal("Failed to list versions"))
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

// HandleDeactivateRegistryVersion returns POST /v1/admin/registry/functions/{functionId}/versions/{versionId}/deactivate
func (h *RegistryHandler) HandleDeactivateRegistryVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["functionId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}
	versionID, err := uuid.Parse(vars["versionId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid version ID"))
		return
	}

	// Verify the version belongs to the function
	version, err := h.registryRepo.GetVersionByID(versionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Version not found"))
		return
	}
	if version.FunctionID != functionID {
		apierror.WriteError(w, apierror.NewBadRequest("Version does not belong to the specified function"))
		return
	}

	if err := h.registryRepo.DeactivateFunctionVersion(versionID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": functionID,
			"version_id":  versionID,
		}).Error("Failed to deactivate registry version (admin)")
		apierror.WriteError(w, apierror.NewInternal("Failed to deactivate version"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           versionID.String(),
		"function_id":  functionID.String(),
		"version":      version.Version,
		"is_active":    false,
		"published_at": version.PublishedAt,
	})
}

// HandleGetRegistryFunctionMetrics returns GET /v1/admin/registry/functions/{functionId}/metrics
func (h *RegistryHandler) HandleGetRegistryFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Default to last 30 days of metrics
	since := time.Now().AddDate(0, 0, -30)
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}

	totalCalls, successRate, avgLatency, p95Latency, err := h.registryRepo.GetFunctionStats(functionID, since)
	if err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Warn("HandleGetRegistryFunctionMetrics: failed to get function stats")
		// Return zeros rather than error to avoid breaking dashboards
		totalCalls, successRate, avgLatency, p95Latency = 0, 0, 0, 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"function_id":    functionIDStr,
		"invocations":    totalCalls,
		"success_rate":  successRate,
		"errors":        int(float64(totalCalls) * (100 - successRate) / 100),
		"latency_p50_ms": avgLatency,
		"latency_p99_ms": p95Latency,
		"since":          since.Format(time.RFC3339),
	})
}

const openRouterFreeModel = "arcee-ai/trinity-large-preview:free"
const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

// HandleGenerateRegistryDescription returns POST /v1/admin/registry/generate-description
// Uses Open Router free models to generate a short description from function name/title/category.
func (h *RegistryHandler) HandleGenerateRegistryDescription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apierror.WriteError(w, apierror.NewBadRequest("Method not allowed"))
		return
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Open Router API key not configured (OPENROUTER_API_KEY)"))
		return
	}

	var body struct {
		Name     string `json:"name"`
		Title    string `json:"title"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name is required"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to create request"))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Open Router request failed")
		apierror.WriteError(w, apierror.NewInternal("Open Router request failed"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("status", resp.StatusCode).Error("Open Router returned non-200")
		apierror.WriteError(w, apierror.NewInternal("Open Router returned error"))
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
		apierror.WriteError(w, apierror.NewInternal("Failed to parse response"))
		return
	}
	description := ""
	if len(openResp.Choices) > 0 {
		description = strings.TrimSpace(openResp.Choices[0].Message.Content)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"description": description})
}

// HandlePurgeAllCache returns DELETE /v1/admin/cache - Purge all cache entries
func (h *RegistryHandler) HandlePurgeAllCache(w http.ResponseWriter, r *http.Request) {
	if h.cacheService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Cache service not available"))
		return
	}

	if err := h.cacheService.PurgeAll(); err != nil {
		logrus.WithError(err).Error("Failed to purge all cache entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to purge cache"))
		return
	}

	logrus.Info("Admin purged all cache entries")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": "All cache entries purged successfully",
	})
}

// HandlePurgeFunctionCache returns DELETE /v1/admin/cache/{functionId} - Purge all cache for a function
func (h *RegistryHandler) HandlePurgeFunctionCache(w http.ResponseWriter, r *http.Request) {
	if h.cacheService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Cache service not available"))
		return
	}

	vars := mux.Vars(r)
	functionID := vars["functionId"]
	if functionID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Function ID is required"))
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(functionID); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID format"))
		return
	}

	if err := h.cacheService.InvalidateFunction(functionID); err != nil {
		logrus.WithError(err).WithField("function_id", functionID).Error("Failed to purge function cache")
		apierror.WriteError(w, apierror.NewInternal("Failed to purge function cache"))
		return
	}

	logrus.WithField("function_id", functionID).Info("Admin purged function cache")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"message":     "Function cache purged successfully",
		"function_id": functionID,
	})
}

// HandlePurgeVersionCache returns DELETE /v1/admin/cache/{functionId}/{version} - Purge specific version cache
func (h *RegistryHandler) HandlePurgeVersionCache(w http.ResponseWriter, r *http.Request) {
	if h.cacheService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Cache service not available"))
		return
	}

	vars := mux.Vars(r)
	functionID := vars["functionId"]
	version := vars["version"]

	if functionID == "" || version == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Function ID and version are required"))
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(functionID); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID format"))
		return
	}

	if err := h.cacheService.InvalidateVersion(functionID, version); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": functionID,
			"version":     version,
		}).Error("Failed to purge version cache")
		apierror.WriteError(w, apierror.NewInternal("Failed to purge version cache"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"function_id": functionID,
		"version":     version,
	}).Info("Admin purged version cache")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"message":     "Version cache purged successfully",
		"function_id": functionID,
		"version":     version,
	})
}

// HandleGetCacheStats returns GET /v1/admin/cache/stats - Get comprehensive cache statistics
func (h *RegistryHandler) HandleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	if h.cacheService == nil {
		apierror.WriteError(w, apierror.NewServiceUnavailable("Cache service not available"))
		return
	}

	stats := map[string]interface{}{
		"cache_service_enabled": true,
	}

	// Get memory cache stats
	if memStats := h.cacheService.GetMemoryStats(); memStats != nil {
		stats["memory_cache"] = map[string]interface{}{
			"hits":       memStats.Hits,
			"misses":     memStats.Misses,
			"hit_ratio":  memStats.Ratio,
			"size_bytes": memStats.SizeBytes,
			"evictions":  memStats.Evictions,
		}
	}

	// Get disk cache stats
	if diskStats, err := h.cacheService.GetDiskStats(); err == nil && diskStats != nil {
		stats["disk_cache"] = map[string]interface{}{
			"total_entries":    diskStats.TotalEntries,
			"total_size_bytes": diskStats.TotalSizeBytes,
			"total_hits":       diskStats.TotalHits,
			"expired_entries":  diskStats.ExpiredEntries,
		}
	}

	// Check Redis status
	stats["redis_enabled"] = h.cacheService.IsRedisCacheEnabled()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
