package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/functionregistry"
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
		if err.Error() == "sql: no rows in result set" {
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

	info := fn.ToInfo(fnVersion)

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

	var funcInfos []map[string]interface{}
	for _, fn := range functions {
		fnVersion, _ := h.repo.GetLatestFunctionVersion(fn.ID)
		// Get rating for trust score
		rating, _ := h.repo.GetRatingByFunctionID(fn.ID)
		funcInfos = append(funcInfos, fn.ToInfoWithRating(fnVersion, rating))
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

	var funcInfos []map[string]interface{}
	for _, fn := range functions {
		fnVersion, _ := h.repo.GetLatestFunctionVersion(fn.ID)
		// Get rating for trust score
		rating, _ := h.repo.GetRatingByFunctionID(fn.ID)
		funcInfos = append(funcInfos, fn.ToInfoWithRating(fnVersion, rating))
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
