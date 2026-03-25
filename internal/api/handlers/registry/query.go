package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetFunction handles getting function info
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

	// Get latest version
	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
		return
	}

	// Include rating for trust_score and trust_level on function profile
	rating, _ := h.repo.GetRatingByFunctionID(fn.ID)
	info := fn.ToInfoWithRating(fnVersion, rating)

	// Backfill high trust for functionfly when rating is missing or still at default 0 (e.g. published before we set defaults)
	if strings.EqualFold(fn.Author, "functionfly") {
		if rating == nil {
			rating, _ = h.repo.GetOrCreateRating(fn.ID)
		}
		if rating != nil && rating.TrustScore == 0 {
			rating.TrustScore = 0.9 // DB stores 0-1; API response will send 90 for frontend
			rating.ReliabilityScore = 0.9
			rating.SuccessRate = 0.9
			if err := h.repo.UpdateTrustScore(rating); err != nil {
				logrus.WithError(err).WithField("function_id", fn.ID).Debug("Failed to backfill rating trust score")
			} else {
				dreScores := &storageregistry.DREScores{
					DeterminismScore:          0.9,
					ReplayIntegrityScore:      0.9,
					PerformanceStabilityScore: 0.9,
					DriftScore:                1.0,
				}
				_ = h.repo.UpdateTrustScoreV2(fn.ID, dreScores, 0.9)
			}
			info["trust_score"] = 90 // 0-100 scale for frontend
			info["trust_level"] = "high"
			info["success_rate"] = 0.9
			info["reliability"] = 90
		}
		// Ensure function-level scores are high for display (reliability_score, deterministic_score)
		if fn.ReliabilityScore == 0 && fn.DeterministicScore == 0 {
			_, _ = h.repo.UpdateRegistryFunction(fn.ID, map[string]interface{}{
				"reliability_score":   90.0,
				"deterministic_score": 90.0,
			})
			info["reliability"] = 90
		}
	}

	// Include verification status so UI can show Verified for functionfly / approved functions
	verStatus, errVer := h.repo.GetVerificationStatus(fnVersion.ID)
	verified := verStatus != nil && verStatus.OverallStatus == "verified"
	if !verified && strings.EqualFold(fn.Author, "functionfly") {
		// Backfill verification row for trusted author so future requests don't hit "record not found"
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
		verified = true
	}
	info["verified"] = verified
	_ = errVer

	// Expand manifest if requested
	if r.URL.Query().Get("expand") == "manifest" {
		var manifest functionregistry.FunctionManifest
		if err := json.Unmarshal(fnVersion.Manifest, &manifest); err == nil {
			info["manifest"] = manifest
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// buildRegistryFunctionInfos loads latest versions and ratings in two queries instead of 2N per-function calls.
func (h *Handler) buildRegistryFunctionInfos(functions []storageregistry.RegistryFunction) ([]map[string]interface{}, error) {
	if len(functions) == 0 {
		return nil, nil
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
	out := make([]map[string]interface{}, len(functions))
	for i, fn := range functions {
		v := versions[fn.ID]
		rating := ratings[fn.ID]
		out[i] = fn.ToInfoWithRating(v, rating)
	}
	return out, nil
}

// HandleListFunctions handles listing functions
func (h *Handler) HandleListFunctions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
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

// HandleSearchFunctions handles searching functions
func (h *Handler) HandleSearchFunctions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")
	runtime := r.URL.Query().Get("runtime")
	minRating, _ := strconv.ParseFloat(r.URL.Query().Get("min_rating"), 64)

	functions, total, err := h.repo.SearchFunctions(query, category, runtime, minRating, limit, offset)
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

// HandleListVersions handles listing function versions
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

// HandleListVersionsAt handles listing function versions using @username URL structure
// URL: /@/{username}/v1/fx/{functionName}/versions
func (h *Handler) HandleListVersionsAt(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	functionName := vars["functionName"]

	// Remove @ prefix if present
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

// HandleDeleteFunction handles deleting a function
func (h *Handler) HandleDeleteFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	err := h.repo.DeleteFunction(author, name)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete function")
		// Check if it's a "not found" type error
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

// HandleDeleteAllFunctions handles deleting all functions (for reset)
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

// HandleGetSimilarFunctions handles getting similar functions
func (h *Handler) HandleGetSimilarFunctions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Search for functions in the same category
	var similar []map[string]interface{}
	if fn.Category.Valid && fn.Category.String != "" {
		functions, _, err := h.repo.SearchFunctions("", fn.Category.String, "", 0, 5, 0)
		if err == nil {
			for _, f := range functions {
				// Skip the original function
				if f.Author == author && f.Name == name {
					continue
				}
				fnVersion, _ := h.repo.GetLatestFunctionVersion(f.ID)
				similar = append(similar, f.ToInfo(fnVersion))
			}
		}
	}

	// If no similar found by category, return popular functions
	if len(similar) == 0 {
		functions, _, err := h.repo.SearchFunctions("", "", "", 50, 5, 0)
		if err == nil {
			for _, f := range functions {
				if f.Author == author && f.Name == name {
					continue
				}
				fnVersion, _ := h.repo.GetLatestFunctionVersion(f.ID)
				similar = append(similar, f.ToInfo(fnVersion))
				if len(similar) >= 5 {
					break
				}
			}
		}
	}

	response := map[string]interface{}{
		"function": fmt.Sprintf("%s/%s", author, name),
		"similar":  similar,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
