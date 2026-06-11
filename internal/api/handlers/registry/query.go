package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/functionregistry"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	TrustedAuthorTrustScore    = 0.9
	TrustedAuthorTrustScorePct = 90.0
	TrustedAuthorDriftScore    = 1.0
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

func (h *Handler) HandleGetFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "record not found") || strings.Contains(errStr, "sql: no rows in result set") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
		return
	}

	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)
	info := fn.ToInfoWithRating(fnVersion, rating)

	if strings.EqualFold(fn.Author, "functionfly") {
		go func() {
			localRating := rating
			if localRating == nil {
				localRating, _ = h.repo.GetOrCreateRating(fn.ID)
			}
			if localRating != nil && localRating.TrustScore == 0 {
				localRating.TrustScore = TrustedAuthorTrustScore
				localRating.ReliabilityScore = TrustedAuthorTrustScore
				localRating.SuccessRate = TrustedAuthorTrustScore
				if err := h.repo.UpdateTrustScore(localRating); err != nil {
					logrus.WithError(err).WithField("function_id", fn.ID).Debug("Failed to backfill rating trust score")
				} else {
					dreScores := &storageregistry.DREScores{
						DeterminismScore:          TrustedAuthorTrustScore,
						ReplayIntegrityScore:      TrustedAuthorTrustScore,
						PerformanceStabilityScore: TrustedAuthorTrustScore,
						DriftScore:                TrustedAuthorDriftScore,
					}
					_ = h.repo.UpdateTrustScoreV2(fn.ID, dreScores, TrustedAuthorTrustScore)
				}
			}
			if fn.ReliabilityScore == 0 && fn.DeterministicScore == 0 {
				_, _ = h.repo.UpdateRegistryFunction(fn.ID, map[string]interface{}{
					"reliability_score":   TrustedAuthorTrustScorePct,
					"deterministic_score": TrustedAuthorTrustScorePct,
				})
			}
		}()

		info["trust_score"] = int(TrustedAuthorTrustScorePct)
		info["trust_level"] = "high"
		info["success_rate"] = TrustedAuthorTrustScore
		info["reliability"] = int(TrustedAuthorTrustScorePct)
	}

	verStatus, _ := h.repo.GetVerificationStatus(fnVersion.ID)
	verified := verStatus != nil && verStatus.OverallStatus == "verified"
	if !verified && strings.EqualFold(fn.Author, "functionfly") {
		verified = true
		go func() {
			now := time.Now()
			status := &storageregistry.RegistryFunctionVerificationStatus{
				ID:                  uuid.New(),
				FunctionVersionID:   fnVersion.ID,
				ContentHashVerified: true,
				SignatureVerified:   true,
				MalwareScanned:      true,
				MalwareStatus:       "clean",
				OverallStatus:       "verified",
				LastVerifiedAt:      &now,
				CreatedAt:           now,
				UpdatedAt:           now,
			}
			if err := h.repo.CreateOrUpdateVerificationStatus(status); err != nil {
				logrus.WithError(err).WithField("function_version_id", fnVersion.ID).Debug("Failed to backfill verification status")
			}
		}()
	}
	info["verified"] = verified

	likeCount, _ := h.repo.CountLikesForFunction(fn.ID)
	info["like_count"] = likeCount
	remixCount, _ := h.repo.CountRemixesForFunction(fn.ID)
	info["remix_count"] = remixCount

	if r.URL.Query().Get("expand") == "manifest" {
		var manifest functionregistry.FunctionManifest
		if err := json.Unmarshal(fnVersion.Manifest, &manifest); err == nil {
			transformedManifest := transformManifestForFrontend(manifest)
			info["manifest"] = transformedManifest
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *Handler) HandleGetFunctionByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["functionId"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	fn, err := h.repo.GetFunctionByID(id)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "record not found") || strings.Contains(errStr, "sql: no rows in result set") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function by ID")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
		return
	}

	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)
	info := fn.ToInfoWithRating(fnVersion, rating)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (h *Handler) buildRegistryFunctionInfos(functions []storageregistry.RegistryFunction) ([]map[string]interface{}, error) {
	if len(functions) == 0 {
		return make([]map[string]interface{}, 0), nil
	}
	ids := make([]uuid.UUID, len(functions))
	for i := range functions {
		ids[i] = functions[i].ID
	}
	versions, err := h.repo.ListLatestVersionsForFunctions(ids)
	if err != nil {
		return nil, err
	}
	ratings, err := h.repo.GetRatingsByFunctionIDs(ids)
	if err != nil {
		return nil, err
	}
	likeCounts, _ := h.repo.CountLikesForFunctions(ids)
	remixCounts, _ := h.repo.CountRemixesForFunctions(ids)
	out := make([]map[string]interface{}, len(functions))
	for i, fn := range functions {
		v := versions[fn.ID]
		rating := ratings[fn.ID]
		info := fn.ToInfoWithRating(v, rating)
		info["like_count"] = likeCounts[fn.ID]
		info["remix_count"] = remixCounts[fn.ID]
		out[i] = info
	}
	return out, nil
}

func (h *Handler) HandleListFunctions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := r.URL.Query().Get("author")
	if author == "" {
		author = vars["author"]
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")
	visibility := r.URL.Query().Get("visibility")

	functions, total, err := h.repo.ListFunctions(author, category, nil, visibility, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	funcInfos, err := h.buildRegistryFunctionInfos(functions)
	if err != nil {
		logrus.WithError(err).Error("Failed to enrich function list")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	response := functionregistry.ListFunctionsResponse{
		Functions: convertToFunctionInfos(funcInfos),
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleListFunctionsWithMCP serves GET /v1/functions/mcp.
// Returns all functions that have MCP settings configured (enabled or disabled).
func (h *Handler) HandleListFunctionsWithMCP(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	functions, total, err := h.repo.ListFunctionsWithMCPSettings(r.Context(), limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions with MCP settings")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	// Transform to the response format expected by the frontend
	type MCPFunctionInfo struct {
		ID         string `json:"id"`
		Author     string `json:"author"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		MCPSettings struct {
			Enabled            bool     `json:"enabled"`
			Transports         []string `json:"transports"`
			ExposeInputSchema  bool     `json:"expose_input_schema"`
			ExposeOutputSchema bool     `json:"expose_output_schema"`
			ToolNameOverride   string   `json:"tool_name_override,omitempty"`
			RateLimitPerMin    int      `json:"rate_limit_per_min"`
			AllowlistOrigins   []string `json:"allowlist_origins"`
			VerifiedMCP        bool     `json:"verified_mcp"`
			InvocationCount    int64    `json:"invocation_count"`
			LastInvokedAt      *string  `json:"last_invoked_at,omitempty"`
		} `json:"mcp"`
	}

	funcs := make([]MCPFunctionInfo, 0, len(functions))
	for _, f := range functions {
		info := MCPFunctionInfo{
			ID:     f.ID.String(),
			Author: f.Author,
			Name:   f.Name,
			Status: f.Status,
		}
		info.MCPSettings.Enabled = f.Enabled
		info.MCPSettings.Transports = f.Transports
		info.MCPSettings.ExposeInputSchema = f.ExposeInputSchema
		info.MCPSettings.ExposeOutputSchema = f.ExposeOutputSchema
		info.MCPSettings.RateLimitPerMin = f.RateLimitPerMin
		info.MCPSettings.AllowlistOrigins = f.AllowlistOrigins
		info.MCPSettings.VerifiedMCP = f.VerifiedMCP
		info.MCPSettings.InvocationCount = f.InvocationCount
		if f.ToolNameOverride.Valid {
			info.MCPSettings.ToolNameOverride = f.ToolNameOverride.String
		}
		if f.LastInvokedAt != nil {
			s := f.LastInvokedAt.Format(time.RFC3339)
			info.MCPSettings.LastInvokedAt = &s
		}
		funcs = append(funcs, info)
	}

	response := map[string]interface{}{
		"functions": funcs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleListMyFunctions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	functions, total, err := h.repo.ListFunctionsByOwner(user.UserID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list user's registry functions")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	funcInfos, err := h.buildRegistryFunctionInfos(functions)
	if err != nil {
		logrus.WithError(err).Error("Failed to enrich user's function list")
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	response := functionregistry.ListFunctionsResponse{
		Functions: convertToFunctionInfos(funcInfos),
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleSearchFunctions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")
	runtime := r.URL.Query().Get("runtime")
	minRating, _ := strconv.ParseFloat(r.URL.Query().Get("min_rating"), 64)
	sortBy := r.URL.Query().Get("sort_by")

	functions, total, err := h.repo.SearchFunctionsWithSort(query, category, runtime, minRating, limit, offset, sortBy)
	if err != nil {
		logrus.WithError(err).Error("Failed to search functions")
		http.Error(w, "Failed to search functions", http.StatusInternalServerError)
		return
	}

	funcInfos, err := h.buildRegistryFunctionInfos(functions)
	if err != nil {
		logrus.WithError(err).Error("Failed to enrich search results")
		http.Error(w, "Failed to search functions", http.StatusInternalServerError)
		return
	}

	response := functionregistry.ListFunctionsResponse{
		Functions: convertToFunctionInfos(funcInfos),
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleListVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	versions, err := h.repo.ListFunctionVersions(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list versions")
		http.Error(w, "Failed to list versions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

func (h *Handler) HandleGetFunctionSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "latest"
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	var fnVersion *storageregistry.RegistryFunctionVersion
	if version == "latest" {
		fnVersion, err = h.repo.GetLatestFunctionVersion(fn.ID)
	} else {
		fnVersion, err = h.repo.GetFunctionVersion(fn.ID, version)
	}
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"source_code": fnVersion.SourceCode.String,
		"version":     fnVersion.Version,
		"runtime":     fnVersion.Runtime,
	})
}

func (h *Handler) HandleListVersionsAt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	functionName := vars["functionName"]

	if len(username) > 0 && username[0] == '@' {
		username = username[1:]
	}

	fn, err := h.repo.GetFunctionByAuthorName(username, functionName)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	versions, err := h.repo.ListFunctionVersions(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list versions")
		http.Error(w, "Failed to list versions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

func (h *Handler) HandleDeleteFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "failed to find function") {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get function")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if fn.OwnerUserID == nil || *fn.OwnerUserID != user.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.repo.DeleteFunction(author, name)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete function")
		if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "failed to find function") {
			response := map[string]string{
				"message": "Function not found",
				"author":  author,
				"name":    name,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		http.Error(w, "Failed to delete function", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Function deleted successfully",
		"author":  author,
		"name":    name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleDeleteAllFunctions(w http.ResponseWriter, r *http.Request) {
	err := h.repo.DeleteAllFunctions()
	if err != nil {
		logrus.WithError(err).Error("Failed to delete all functions")
		http.Error(w, "Failed to delete all functions", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "All functions deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) buildSimilarFunctionInfos(functions []storageregistry.RegistryFunction, excludeAuthor, excludeName string, maxResults int) []map[string]interface{} {
	if len(functions) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(functions))
	for _, f := range functions {
		if f.Author == excludeAuthor && f.Name == excludeName {
			continue
		}
		ids = append(ids, f.ID)
	}

	if len(ids) == 0 {
		return nil
	}

	versions, err := h.repo.ListLatestVersionsForFunctions(ids)
	if err != nil {
		logrus.WithError(err).Debug("Failed to batch load latest versions for similar functions")
		return nil
	}

	out := make([]map[string]interface{}, 0, maxResults)
	for _, f := range functions {
		if f.Author == excludeAuthor && f.Name == excludeName {
			continue
		}
		v := versions[f.ID]
		out = append(out, f.ToInfo(v))
		if len(out) >= maxResults {
			break
		}
	}

	return out
}

func (h *Handler) HandleGetSimilarFunctions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	var similar []map[string]interface{}
	if fn.Category.Valid && fn.Category.String != "" {
		functions, _, err := h.repo.SearchFunctions("", fn.Category.String, "", 0, 5, 0)
		if err == nil {
			similar = h.buildSimilarFunctionInfos(functions, author, name, 5)
		}
	}

	if len(similar) == 0 {
		functions, _, err := h.repo.SearchFunctions("", "", "", 50, 5, 0)
		if err == nil {
			similar = h.buildSimilarFunctionInfos(functions, author, name, 5)
		}
	}

	response := map[string]interface{}{
		"function": fmt.Sprintf("%s/%s", author, name),
		"similar":  similar,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func transformManifestForFrontend(m functionregistry.FunctionManifest) map[string]interface{} {
	result := make(map[string]interface{})

	if m.Input != nil {
		result["input"] = transformIOTypeForFrontend(*m.Input)
	}
	if m.Output != nil {
		result["output"] = transformIOTypeForFrontend(*m.Output)
	}
	if len(m.Examples) > 0 {
		result["examples"] = m.Examples
	}

	return result
}

func transformIOTypeForFrontend(io functionregistry.IOType) map[string]interface{} {
	result := map[string]interface{}{
		"type": io.Type,
	}

	if len(io.Properties) > 0 {
		var props map[string]interface{}
		if err := json.Unmarshal(io.Properties, &props); err == nil {
			result["schema"] = map[string]interface{}{
				"type":       io.Type,
				"properties": props,
			}
		}
	}

	if len(io.Schema) > 0 {
		var schema map[string]interface{}
		if err := json.Unmarshal(io.Schema, &schema); err == nil {
			if _, ok := result["schema"]; !ok {
				result["schema"] = schema
			}
		}
	}

	if len(io.Example) > 0 {
		var example interface{}
		if err := json.Unmarshal(io.Example, &example); err == nil {
			result["example"] = example
		}
	}

	if io.Required.IsRequired() {
		if len(io.Required.Array) > 0 {
			result["required"] = io.Required.Array
		} else if io.Required.Bool {
			result["required"] = true
		}
	}

	return result
}
