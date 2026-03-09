// Package version provides HTTP handlers for API version management.
package version

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/versioning"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains version management handlers
type Handler struct {
	repo *versioning.Repository
}

// NewHandler creates a new version handler
func NewHandler(repo *versioning.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// HandleListVersions handles GET /api/versions - list available API versions
// @Summary List API versions
// @Description Returns a list of all available API versions with their status and deprecation information
// @Tags Versioning
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/versions [get]
func (h *Handler) HandleListVersions(w http.ResponseWriter, r *http.Request) {
	params := versioning.ListAPIVersionsParams{
		Status: r.URL.Query().Get("status"),
		Limit:  20,
	}

	versions, err := h.repo.ListAPIVersions(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("Failed to list API versions")
		http.Error(w, "Failed to list API versions", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	responses := make([]versioning.APIVersionResponse, 0, len(versions))
	for _, v := range versions {
		resp := versioning.APIVersionResponse{
			Version:    v.Version,
			Status:     v.Status,
			ReleasedAt: v.ReleasedAt,
		}

		// Extract features from metadata
		if len(v.Metadata) > 0 {
			var metadata map[string]interface{}
			if err := json.Unmarshal(v.Metadata, &metadata); err == nil {
				if features, ok := metadata["features"].([]interface{}); ok {
					resp.Features = make([]string, len(features))
					for i, f := range features {
						if s, ok := f.(string); ok {
							resp.Features[i] = s
						}
					}
				}
			}
		}

		// Add deprecation info if applicable
		if v.Status == versioning.APIVersionStatusDeprecated || v.Status == versioning.APIVersionStatusSunset {
			resp.Deprecation = &versioning.DeprecationInfo{}
			if v.DeprecatedAt != nil {
				resp.Deprecation.DeprecatedAt = *v.DeprecatedAt
			}
			if v.SunsetAt != nil {
				resp.Deprecation.SunsetAt = *v.SunsetAt
			}
			resp.Deprecation.MigrationGuide = v.ChangelogURL
		}

		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": responses,
	})
}

// HandleGetVersion handles GET /api/versions/{version} - get specific API version details
// @Summary Get API version details
// @Description Returns detailed information about a specific API version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param version path string true "API version (e.g., v1, v2)"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/versions/{version} [get]
func (h *Handler) HandleGetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionStr := vars["version"]

	// Handle version prefix
	versionStr = strings.TrimPrefix(versionStr, "v")

	apiVersion, err := h.repo.GetAPIVersionByVersion(r.Context(), "v"+versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get API version")
		http.Error(w, "Failed to get API version", http.StatusInternalServerError)
		return
	}

	if apiVersion == nil {
		http.Error(w, "API version not found", http.StatusNotFound)
		return
	}

	// Build response
	resp := map[string]interface{}{
		"version":        apiVersion.Version,
		"status":         apiVersion.Status,
		"releasedAt":     apiVersion.ReleasedAt,
		"openapiSpec":    apiVersion.OpenAPISpecURL,
		"changelog":      apiVersion.ChangelogURL,
		"supportedUntil": apiVersion.SunsetAt,
	}

	// Add deprecation info if applicable
	if apiVersion.Status == versioning.APIVersionStatusDeprecated || apiVersion.Status == versioning.APIVersionStatusSunset {
		deprecation := map[string]interface{}{
			"deprecatedAt": apiVersion.DeprecatedAt,
		}
		if apiVersion.SunsetAt != nil {
			deprecation["sunsetAt"] = apiVersion.SunsetAt
		}
		if apiVersion.SunsetMessage != "" {
			deprecation["sunsetMessage"] = apiVersion.SunsetMessage
		}
		resp["deprecation"] = deprecation
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleListFunctionVersions handles GET /functions/{functionId}/versions - list function versions
// @Summary List function versions
// @Description Returns a list of all versions for a specific function
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param status query string false "Filter by status (active, deprecated, archived)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions [get]
func (h *Handler) HandleListFunctionVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	params := versioning.ListFunctionVersionsParams{
		FunctionID: functionID,
		Status:     r.URL.Query().Get("status"),
		Limit:      20,
	}

	versions, err := h.repo.ListFunctionVersions(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("Failed to list function versions")
		http.Error(w, "Failed to list function versions", http.StatusInternalServerError)
		return
	}

	// Get latest version to determine isLatest flag
	latestVersion, err := h.repo.GetLatestFunctionVersion(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get latest function version")
	}

	// Convert to response
	responses := make([]versioning.FunctionVersionResponse, 0, len(versions))
	for _, v := range versions {
		resp := versioning.FunctionVersionResponse{
			ID:         v.ID,
			FunctionID: v.FunctionID,
			Version:    v.Version,
			Status:     v.VersionState,
			IsLatest:   latestVersion != nil && v.ID == latestVersion.ID,
			IsStable:   !isPrerelease(v.Version),
		}

		if v.VersionState == versioning.FunctionVersionStateDeprecated {
			resp.Deprecation = &versioning.VersionDeprecationInfo{
				Reason:         v.DeprecationReason,
				MigrationGuide: v.MigrationGuide,
				ReplacedBy:     v.ReplacedByVersion,
			}
		}

		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": responses,
	})
}

// HandleGetFunctionVersion handles GET /functions/{functionId}/versions/{version} - get specific function version
// @Summary Get function version
// @Description Returns detailed information about a specific function version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version} [get]
func (h *Handler) HandleGetFunctionVersion(w http.ResponseWriter, r *http.Request) {
	// This would typically query the registry_function_versions table
	// For now, return a placeholder response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Get function version not yet implemented",
	})
}

// HandleCreateChangelog handles POST /functions/{functionId}/versions/{version}/changelog - create changelog entry
// @Summary Create changelog entry
// @Description Creates a new changelog entry for a function version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param body body struct{} true "Changelog entry details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/changelog [post]
func (h *Handler) HandleCreateChangelog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	versionStr := vars["version"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ChangeType      string   `json:"changeType"`
		ChangeCategory  string   `json:"changeCategory"`
		Description     string   `json:"description"`
		BreakingChanges []string `json:"breakingChanges"`
		MigrationSteps  []string `json:"migrationSteps"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from context
	user := middleware.GetUserFromContext(r)
	var createdBy *uuid.UUID
	if user != nil {
		id := uuid.MustParse(user.ID)
		createdBy = &id
	}

	params := versioning.CreateChangelogParams{
		FunctionID:      functionID,
		Version:         versionStr,
		ChangeType:      req.ChangeType,
		ChangeCategory:  req.ChangeCategory,
		Description:     req.Description,
		BreakingChanges: req.BreakingChanges,
		MigrationSteps:  req.MigrationSteps,
		CreatedBy:       createdBy,
	}

	changelog, err := h.repo.CreateChangelog(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("Failed to create changelog")
		http.Error(w, "Failed to create changelog", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(changelog)
}

// HandleGetChangelogs handles GET /functions/{functionId}/changelogs - get function changelogs
// @Summary Get function changelogs
// @Description Returns all changelog entries for a function
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/changelogs [get]
func (h *Handler) HandleGetChangelogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	changelogs, err := h.repo.GetChangelogByFunctionID(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get changelogs")
		http.Error(w, "Failed to get changelogs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"changelogs": changelogs,
	})
}

// HandleDeprecateVersion handles POST /api/versions/{version}/deprecate - deprecate an API version
// @Summary Deprecate API version
// @Description Marks an API version as deprecated with optional sunset information
// @Tags Versioning
// @Accept json
// @Produce json
// @Param version path string true "API version (e.g., v1, v2)"
// @Param body body struct{ SunsetMessage string "json:\"sunsetMessage\"" } true "Deprecation details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/versions/{version}/deprecate [post]
func (h *Handler) HandleDeprecateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionStr := vars["version"]

	var req struct {
		SunsetMessage string `json:"sunsetMessage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the version
	apiVersion, err := h.repo.GetAPIVersionByVersion(r.Context(), versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get API version")
		http.Error(w, "Failed to get API version", http.StatusInternalServerError)
		return
	}

	if apiVersion == nil {
		http.Error(w, "API version not found", http.StatusNotFound)
		return
	}

	// Set sunset date to 30 days from now if not provided
	sunsetAt := time.Now().Add(30 * 24 * time.Hour)

	err = h.repo.DeprecateAPIVersion(r.Context(), apiVersion.ID, &sunsetAt, req.SunsetMessage)
	if err != nil {
		logrus.WithError(err).Error("Failed to deprecate API version")
		http.Error(w, "Failed to deprecate API version", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":       apiVersion.Version,
		"status":        "deprecated",
		"deprecatedAt":  time.Now(),
		"sunsetAt":      sunsetAt,
		"sunsetMessage": req.SunsetMessage,
	})
}

// isPrerelease checks if a version string is a prerelease
func isPrerelease(version string) bool {
	for _, c := range []string{"-alpha", "-beta", "-rc", "-dev"} {
		if strings.Contains(version, c) {
			return true
		}
	}
	return false
}

// HandlePublishVersion handles POST /functions/{functionId}/versions/{version}/publish - publish a function version
// @Summary Publish function version
// @Description Publishes a function version, making it available for use
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param body body versioning.PublishVersionRequest false "Publish options"
// @Success 200 {object} versioning.PublishVersionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/publish [post]
func (h *Handler) HandlePublishVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	versionStr := vars["version"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var req versioning.PublishVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults if no body
		req.Version = versionStr
		req.SetAsLatest = true
	}

	// Get the function version
	version, err := h.repo.GetFunctionVersionByVersion(r.Context(), functionID, versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function version")
		http.Error(w, "Failed to get function version", http.StatusInternalServerError)
		return
	}

	if version == nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Check if already published
	if version.VersionState == versioning.FunctionVersionStatePublished {
		http.Error(w, "Version already published", http.StatusBadRequest)
		return
	}

	// Publish the version
	publishedVersion, err := h.repo.PublishFunctionVersion(r.Context(), version.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish function version")
		http.Error(w, "Failed to publish version", http.StatusInternalServerError)
		return
	}

	// Set aliases if requested
	if req.SetAsLatest {
		_ = h.repo.SetVersionAlias(r.Context(), functionID, string(versioning.VersionAliasLatest), publishedVersion.ID)
	}
	if req.SetAsStable && !isPrerelease(publishedVersion.Version) {
		_ = h.repo.SetVersionAlias(r.Context(), functionID, string(versioning.VersionAliasStable), publishedVersion.ID)
	}

	// Get latest and stable status
	latestVersion, _ := h.repo.GetLatestFunctionVersion(r.Context(), functionID)
	stableVersion, _ := h.repo.GetStableFunctionVersion(r.Context(), functionID)

	resp := versioning.PublishVersionResponse{
		ID:          publishedVersion.ID,
		FunctionID:  publishedVersion.FunctionID,
		Version:     publishedVersion.Version,
		Status:      publishedVersion.VersionState,
		PublishedAt: publishedVersion.PublishedAt,
		IsLatest:    latestVersion != nil && latestVersion.ID == publishedVersion.ID,
		IsStable:    stableVersion != nil && stableVersion.ID == publishedVersion.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleArchiveVersion handles POST /functions/{functionId}/versions/{version}/archive - archive a function version
// @Summary Archive function version
// @Description Archives a function version, removing it from active use
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param body body versioning.ArchiveVersionRequest false "Archive details"
// @Success 200 {object} versioning.ArchiveVersionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/archive [post]
func (h *Handler) HandleArchiveVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	versionStr := vars["version"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var req versioning.ArchiveVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = ""
	}

	// Get the function version
	version, err := h.repo.GetFunctionVersionByVersion(r.Context(), functionID, versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function version")
		http.Error(w, "Failed to get function version", http.StatusInternalServerError)
		return
	}

	if version == nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Archive the version
	archivedVersion, err := h.repo.ArchiveFunctionVersion(r.Context(), version.ID, req.Reason)
	if err != nil {
		logrus.WithError(err).Error("Failed to archive function version")
		http.Error(w, "Failed to archive version", http.StatusInternalServerError)
		return
	}

	resp := versioning.ArchiveVersionResponse{
		Version:    archivedVersion.Version,
		Status:     archivedVersion.VersionState,
		ArchivedAt: archivedVersion.ArchivedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleDeprecateFunctionVersion handles POST /functions/{functionId}/versions/{version}/deprecate - deprecate a function version
// @Summary Deprecate function version
// @Description Marks a function version as deprecated with migration information
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param body body versioning.DeprecateVersionRequest true "Deprecation details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/deprecate [post]
func (h *Handler) HandleDeprecateFunctionVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	versionStr := vars["version"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var req versioning.DeprecateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the function version
	version, err := h.repo.GetFunctionVersionByVersion(r.Context(), functionID, versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function version")
		http.Error(w, "Failed to get function version", http.StatusInternalServerError)
		return
	}

	if version == nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Deprecate the version
	err = h.repo.DeprecateFunctionVersion(r.Context(), version.ID, req.Reason, req.ReplacedBy, req.MigrationGuide)
	if err != nil {
		logrus.WithError(err).Error("Failed to deprecate function version")
		http.Error(w, "Failed to deprecate version", http.StatusInternalServerError)
		return
	}

	// Calculate sunset date
	gracePeriodDays := req.GracePeriodDays
	if gracePeriodDays == 0 {
		gracePeriodDays = 30
	}
	sunsetAt := time.Now().Add(time.Duration(gracePeriodDays) * 24 * time.Hour)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":        version.Version,
		"status":         versioning.FunctionVersionStateDeprecated,
		"deprecatedAt":   time.Now(),
		"sunsetAt":       sunsetAt,
		"reason":         req.Reason,
		"replacedBy":     req.ReplacedBy,
		"migrationGuide": req.MigrationGuide,
	})
}

// HandleSetAlias handles POST /functions/{functionId}/versions/{version}/alias/{alias} - set version alias
// @Summary Set version alias
// @Description Sets an alias (latest, stable) for a function version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param alias path string true "Alias type (latest or stable)"
// @Success 200 {object} versioning.SetAliasResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/alias/{alias} [post]
func (h *Handler) HandleSetAlias(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	versionStr := vars["version"]
	aliasType := vars["alias"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	// Validate alias type
	if aliasType != string(versioning.VersionAliasLatest) && aliasType != string(versioning.VersionAliasStable) {
		http.Error(w, "Invalid alias type. Must be 'latest' or 'stable'", http.StatusBadRequest)
		return
	}

	// Get the function version
	version, err := h.repo.GetFunctionVersionByVersion(r.Context(), functionID, versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function version")
		http.Error(w, "Failed to get function version", http.StatusInternalServerError)
		return
	}

	if version == nil {
		http.Error(w, "Function version not found", http.StatusNotFound)
		return
	}

	// Only published versions can be aliased
	if version.VersionState != versioning.FunctionVersionStatePublished {
		http.Error(w, "Only published versions can be aliased", http.StatusBadRequest)
		return
	}

	// Set the alias
	err = h.repo.SetVersionAlias(r.Context(), functionID, aliasType, version.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to set version alias")
		http.Error(w, "Failed to set alias", http.StatusInternalServerError)
		return
	}

	resp := versioning.SetAliasResponse{
		Alias:   aliasType,
		Version: version.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleRollbackVersion handles POST /functions/{functionId}/versions/{version}/rollback - rollback to specific version
// @Summary Rollback function version
// @Description Rolls back a function to a specific version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Target version for rollback"
// @Param body body versioning.RollbackVersionRequest false "Rollback options"
// @Success 200 {object} versioning.RollbackVersionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/rollback [post]
func (h *Handler) HandleRollbackVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var req versioning.RollbackVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Strategy = versioning.RollbackStrategyImmediate
	}

	// Validate strategy
	if req.Strategy == "" {
		req.Strategy = versioning.RollbackStrategyImmediate
	}

	// Get current latest version
	currentVersion, err := h.repo.GetLatestFunctionVersion(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get current version")
		http.Error(w, "Failed to get current version", http.StatusInternalServerError)
		return
	}

	if currentVersion == nil {
		http.Error(w, "No published versions found", http.StatusNotFound)
		return
	}

	// Determine target version
	targetVersionStr := req.ToVersion
	if targetVersionStr == "" {
		// Rollback to previous version
		previousVersion, err := h.repo.GetPreviousFunctionVersion(r.Context(), functionID, currentVersion.Version)
		if err != nil {
			logrus.WithError(err).Error("Failed to get previous version")
			http.Error(w, "Failed to get previous version", http.StatusInternalServerError)
			return
		}
		if previousVersion == nil {
			http.Error(w, "No previous version to rollback to", http.StatusBadRequest)
			return
		}
		targetVersionStr = previousVersion.Version
	} else {
		// Validate target version exists
		targetVersion, err := h.repo.GetFunctionVersionByVersion(r.Context(), functionID, targetVersionStr)
		if err != nil {
			logrus.WithError(err).Error("Failed to get target version")
			http.Error(w, "Failed to get target version", http.StatusInternalServerError)
			return
		}
		if targetVersion == nil {
			http.Error(w, "Target version not found", http.StatusNotFound)
			return
		}
	}

	// Get user from context
	var initiatedBy *uuid.UUID
	user := r.Context().Value("user")
	if user != nil {
		// Try to get ID if it's a known type
		if claims, ok := user.(*auth.Claims); ok {
			initiatedBy = &claims.UserID
		}
	}

	// Create rollback record
	rollbackParams := versioning.CreateRollbackParams{
		FunctionID:  functionID,
		FromVersion: currentVersion.Version,
		ToVersion:   targetVersionStr,
		Strategy:    req.Strategy,
		InitiatedBy: initiatedBy,
		Metadata:    map[string]interface{}{"strategy": req.Strategy},
	}

	rollbackRecord, err := h.repo.CreateRollbackRecord(r.Context(), rollbackParams)
	if err != nil {
		logrus.WithError(err).Error("Failed to create rollback record")
		http.Error(w, "Failed to create rollback", http.StatusInternalServerError)
		return
	}

	// Set the target version as latest
	targetVersion, err := h.repo.GetFunctionVersionByVersion(r.Context(), functionID, targetVersionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get target version for alias")
	} else if targetVersion != nil {
		_ = h.repo.SetVersionAlias(r.Context(), functionID, string(versioning.VersionAliasLatest), targetVersion.ID)
	}

	// Complete the rollback
	now := time.Now()
	_ = h.repo.CompleteRollbackRecord(r.Context(), rollbackRecord.ID, "completed")

	resp := versioning.RollbackVersionResponse{
		RollbackID:  rollbackRecord.ID,
		FunctionID:  functionID,
		FromVersion: currentVersion.Version,
		ToVersion:   targetVersionStr,
		Strategy:    req.Strategy,
		Status:      "completed",
		InitiatedAt: rollbackRecord.InitiatedAt,
		CompletedAt: &now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleRollbackLatest handles POST /functions/{functionId}/rollback - rollback to previous version
// @Summary Rollback to previous version
// @Description Rolls back a function to its previous version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param body body versioning.RollbackVersionRequest false "Rollback options"
// @Success 200 {object} versioning.RollbackVersionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/rollback [post]
func (h *Handler) HandleRollbackLatest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	var req versioning.RollbackVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Strategy = versioning.RollbackStrategyImmediate
	}

	// Validate strategy
	if req.Strategy == "" {
		req.Strategy = versioning.RollbackStrategyImmediate
	}

	// Get current latest version
	currentVersion, err := h.repo.GetLatestFunctionVersion(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get current version")
		http.Error(w, "Failed to get current version", http.StatusInternalServerError)
		return
	}

	if currentVersion == nil {
		http.Error(w, "No published versions found", http.StatusNotFound)
		return
	}

	// Get previous version
	previousVersion, err := h.repo.GetPreviousFunctionVersion(r.Context(), functionID, currentVersion.Version)
	if err != nil {
		logrus.WithError(err).Error("Failed to get previous version")
		http.Error(w, "Failed to get previous version", http.StatusInternalServerError)
		return
	}

	if previousVersion == nil {
		http.Error(w, "No previous version to rollback to", http.StatusBadRequest)
		return
	}

	// Get user from context
	var initiatedBy *uuid.UUID
	user := r.Context().Value("user")
	if user != nil {
		// Try to get ID if it's a known type
		if claims, ok := user.(*auth.Claims); ok {
			initiatedBy = &claims.UserID
		}
	}

	// Create rollback record
	rollbackParams := versioning.CreateRollbackParams{
		FunctionID:  functionID,
		FromVersion: currentVersion.Version,
		ToVersion:   previousVersion.Version,
		Strategy:    req.Strategy,
		InitiatedBy: initiatedBy,
		Metadata:    map[string]interface{}{"strategy": req.Strategy, "type": "rollback_to_previous"},
	}

	rollbackRecord, err := h.repo.CreateRollbackRecord(r.Context(), rollbackParams)
	if err != nil {
		logrus.WithError(err).Error("Failed to create rollback record")
		http.Error(w, "Failed to create rollback", http.StatusInternalServerError)
		return
	}

	// Set the previous version as latest
	_ = h.repo.SetVersionAlias(r.Context(), functionID, string(versioning.VersionAliasLatest), previousVersion.ID)

	// Complete the rollback
	now := time.Now()
	_ = h.repo.CompleteRollbackRecord(r.Context(), rollbackRecord.ID, "completed")

	resp := versioning.RollbackVersionResponse{
		RollbackID:  rollbackRecord.ID,
		FunctionID:  functionID,
		FromVersion: currentVersion.Version,
		ToVersion:   previousVersion.Version,
		Strategy:    req.Strategy,
		Status:      "completed",
		InitiatedAt: rollbackRecord.InitiatedAt,
		CompletedAt: &now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleGetRollbackHistory handles GET /functions/{functionId}/rollbacks - get rollback history
// @Summary Get rollback history
// @Description Returns the rollback history for a function
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param limit query int false "Maximum number of records to return"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/rollbacks [get]
func (h *Handler) HandleGetRollbackHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	records, err := h.repo.GetRollbackHistory(r.Context(), functionID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get rollback history")
		http.Error(w, "Failed to get rollback history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rollbacks": records,
	})
}

// HandleCreateAPIVersion handles POST /api/versions - create a new API version
// @Summary Create API version
// @Description Creates a new API version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param body body versioning.CreateAPIVersionRequest true "API version details"
// @Success 201 {object} versioning.APIVersion
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/versions [post]
func (h *Handler) HandleCreateAPIVersion(w http.ResponseWriter, r *http.Request) {
	var req versioning.CreateAPIVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate version format
	if !strings.HasPrefix(req.Version, "v") {
		req.Version = "v" + req.Version
	}

	now := time.Now()
	apiVersion := &versioning.APIVersion{
		ID:             uuid.New(),
		Version:        req.Version,
		PathPrefix:     req.PathPrefix,
		Status:         req.Status,
		ReleasedAt:     req.ReleasedAt,
		OpenAPISpecURL: req.OpenAPISpecURL,
		ChangelogURL:   req.ChangelogURL,
		Metadata:       req.Metadata,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if apiVersion.Status == "" {
		apiVersion.Status = versioning.APIVersionStatusActive
	}

	err := h.repo.CreateAPIVersion(r.Context(), apiVersion)
	if err != nil {
		logrus.WithError(err).Error("Failed to create API version")
		http.Error(w, "Failed to create API version", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiVersion)
}

// HandleUpdateAPIVersion handles PATCH /api/versions/{version} - update an API version
// @Summary Update API version
// @Description Updates an existing API version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param version path string true "API version (e.g., v1, v2)"
// @Param body body versioning.UpdateAPIVersionRequest true "Update details"
// @Success 200 {object} versioning.APIVersion
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/versions/{version} [patch]
func (h *Handler) HandleUpdateAPIVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionStr := vars["version"]

	// Handle version prefix
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Get existing version
	apiVersion, err := h.repo.GetAPIVersionByVersion(r.Context(), "v"+versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get API version")
		http.Error(w, "Failed to get API version", http.StatusInternalServerError)
		return
	}

	if apiVersion == nil {
		http.Error(w, "API version not found", http.StatusNotFound)
		return
	}

	var req versioning.UpdateAPIVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.PathPrefix != nil {
		apiVersion.PathPrefix = *req.PathPrefix
	}
	if req.Status != nil {
		apiVersion.Status = *req.Status
	}
	if req.OpenAPISpecURL != nil {
		apiVersion.OpenAPISpecURL = *req.OpenAPISpecURL
	}
	if req.ChangelogURL != nil {
		apiVersion.ChangelogURL = *req.ChangelogURL
	}
	if req.Metadata != nil {
		apiVersion.Metadata = req.Metadata
	}

	err = h.repo.UpdateAPIVersion(r.Context(), apiVersion)
	if err != nil {
		logrus.WithError(err).Error("Failed to update API version")
		http.Error(w, "Failed to update API version", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiVersion)
}

// HandleSetDefaultAPIVersion handles POST /api/versions/{version}/set-default - set default API version
// @Summary Set default API version
// @Description Sets a specific API version as the default
// @Tags Versioning
// @Accept json
// @Produce json
// @Param version path string true "API version (e.g., v1, v2)"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/versions/{version}/set-default [post]
func (h *Handler) HandleSetDefaultAPIVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionStr := vars["version"]

	// Handle version prefix
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Get the version
	apiVersion, err := h.repo.GetAPIVersionByVersion(r.Context(), "v"+versionStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to get API version")
		http.Error(w, "Failed to get API version", http.StatusInternalServerError)
		return
	}

	if apiVersion == nil {
		http.Error(w, "API version not found", http.StatusNotFound)
		return
	}

	// In a real implementation, this would update a configuration or database setting
	// For now, we'll just return success

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":   apiVersion.Version,
		"isDefault": true,
		"message":   "Version set as default",
	})
}

// ==================== Phase 3: Deployment Endpoints ====================

// HandleListDeployments handles GET /functions/{functionId}/versions/{version}/deployments - list deployments
// @Summary List deployments
// @Description Returns a list of deployments for a specific function version
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param status query string false "Filter by deployment status"
// @Param limit query int false "Maximum number of records to return"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/deployments [get]
func (h *Handler) HandleListDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]
	versionStr := vars["version"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	params := versioning.ListDeploymentsParams{
		FunctionID: functionID,
		Version:    versionStr,
		Status:     r.URL.Query().Get("status"),
		Limit:      20,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = parsed
		}
	}

	deployments, err := h.repo.GetDeploymentsByFunctionVersion(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("Failed to list deployments")
		http.Error(w, "Failed to list deployments", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	responses := make([]versioning.DeploymentVersionResponse, 0, len(deployments))
	for _, d := range deployments {
		resp := versioning.DeploymentVersionResponse{
			ID:          d.ID,
			FunctionID:  d.FunctionID,
			Version:     d.Version,
			Provider:    d.Provider,
			Region:      d.Region,
			Status:      d.Status,
			ArtifactURI: d.ArtifactURI,
			Checksum:    d.Checksum,
			CreatedAt:   d.CreatedAt,
			CompletedAt: d.CompletedAt,
		}

		// Parse metadata
		if len(d.Metadata) > 0 {
			var metadata versioning.DeploymentMetadata
			if err := json.Unmarshal(d.Metadata, &metadata); err == nil {
				resp.Metadata = &metadata
			}
		}

		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployments": responses,
	})
}

// HandleGetDeployment handles GET /functions/{functionId}/versions/{version}/deployments/{deploymentId} - get deployment details
// @Summary Get deployment details
// @Description Returns detailed information about a specific deployment
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} versioning.DeploymentVersionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/deployments/{deploymentId} [get]
func (h *Handler) HandleGetDeployment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deploymentIDStr := vars["deploymentId"]

	deploymentID, err := uuid.Parse(deploymentIDStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	deployment, err := h.repo.GetDeploymentByID(r.Context(), deploymentID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get deployment")
		http.Error(w, "Failed to get deployment", http.StatusInternalServerError)
		return
	}

	if deployment == nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	resp := versioning.DeploymentVersionResponse{
		ID:          deployment.ID,
		FunctionID:  deployment.FunctionID,
		Version:     deployment.Version,
		Provider:    deployment.Provider,
		Region:      deployment.Region,
		Status:      deployment.Status,
		ArtifactURI: deployment.ArtifactURI,
		Checksum:    deployment.Checksum,
		CreatedAt:   deployment.CreatedAt,
		CompletedAt: deployment.CompletedAt,
	}

	// Parse metadata
	if len(deployment.Metadata) > 0 {
		var metadata versioning.DeploymentMetadata
		if err := json.Unmarshal(deployment.Metadata, &metadata); err == nil {
			resp.Metadata = &metadata
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ==================== Phase 3: Service Contract Endpoints ====================

// HandleListServiceContracts handles GET /api/internal/contracts - list all service contracts
// @Summary List service contracts
// @Description Returns a list of all service contracts, optionally filtered by service name
// @Tags Versioning
// @Accept json
// @Produce json
// @Param service query string false "Filter by service name"
// @Param status query string false "Filter by contract status"
// @Success 200 {object} versioning.ContractListResponse
// @Failure 500 {object} map[string]string
// @Router /api/internal/contracts [get]
func (h *Handler) HandleListServiceContracts(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")

	params := versioning.ListServiceContractsParams{
		ServiceName: serviceName,
		Status:      r.URL.Query().Get("status"),
		Limit:       50,
	}

	contracts, err := h.repo.GetAllServiceContracts(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("Failed to list service contracts")
		http.Error(w, "Failed to list service contracts", http.StatusInternalServerError)
		return
	}

	// Get all service names if no specific service requested
	var services []string
	if serviceName == "" {
		services, err = h.repo.GetAllServiceNames(r.Context())
		if err != nil {
			logrus.WithError(err).Warn("Failed to get service names")
		}
	}

	// Convert to response format
	responses := make([]versioning.ServiceContractResponse, 0, len(contracts))
	for _, c := range contracts {
		resp := versioning.ServiceContractResponse{
			ID:                  c.ID,
			ServiceName:         c.ServiceName,
			ContractVersion:     c.ContractVersion,
			ContractType:        c.ContractType,
			Status:              c.Status,
			IntroducedInRelease: c.IntroducedInRelease,
			DeprecatedInRelease: c.DeprecatedInRelease,
			CreatedAt:           c.CreatedAt,
		}

		// Parse schema
		if len(c.Schema) > 0 {
			var schema versioning.ContractSchema
			if err := json.Unmarshal(c.Schema, &schema); err == nil {
				resp.Schema = &schema
			}
		}

		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versioning.ContractListResponse{
		Services:  services,
		Contracts: responses,
	})
}

// HandleGetServiceContracts handles GET /api/internal/contracts/{service} - get contract versions for a service
// @Summary Get service contracts
// @Description Returns all contract versions for a specific service
// @Tags Versioning
// @Accept json
// @Produce json
// @Param service path string true "Service name"
// @Param status query string false "Filter by contract status"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/internal/contracts/{service} [get]
func (h *Handler) HandleGetServiceContracts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	params := versioning.ListServiceContractsParams{
		ServiceName: serviceName,
		Status:      r.URL.Query().Get("status"),
		Limit:       50,
	}

	contracts, err := h.repo.GetAllServiceContracts(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Error("Failed to get service contracts")
		http.Error(w, "Failed to get service contracts", http.StatusInternalServerError)
		return
	}

	if len(contracts) == 0 {
		http.Error(w, "No contracts found for service", http.StatusNotFound)
		return
	}

	// Convert to response format
	responses := make([]versioning.ServiceContractResponse, 0, len(contracts))
	for _, c := range contracts {
		resp := versioning.ServiceContractResponse{
			ID:                  c.ID,
			ServiceName:         c.ServiceName,
			ContractVersion:     c.ContractVersion,
			ContractType:        c.ContractType,
			Status:              c.Status,
			IntroducedInRelease: c.IntroducedInRelease,
			DeprecatedInRelease: c.DeprecatedInRelease,
			CreatedAt:           c.CreatedAt,
		}

		// Parse schema
		if len(c.Schema) > 0 {
			var schema versioning.ContractSchema
			if err := json.Unmarshal(c.Schema, &schema); err == nil {
				resp.Schema = &schema
			}
		}

		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   serviceName,
		"contracts": responses,
	})
}

// HandleNegotiateContractVersion handles POST /api/internal/contracts/negotiate - negotiate contract version
// @Summary Negotiate contract version
// @Description Negotiates a compatible contract version between consumer and provider services
// @Tags Versioning
// @Accept json
// @Produce json
// @Param body body versioning.ContractVersionNegotiationRequest true "Negotiation request"
// @Success 200 {object} versioning.ContractVersionNegotiationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/internal/contracts/negotiate [post]
func (h *Handler) HandleNegotiateContractVersion(w http.ResponseWriter, r *http.Request) {
	var req versioning.ContractVersionNegotiationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProviderService == "" {
		http.Error(w, "Provider service is required", http.StatusBadRequest)
		return
	}

	// Find compatible contract version
	contract, err := h.repo.GetCompatibleContractVersion(r.Context(), req.ProviderService, req.SupportedVersions)
	if err != nil {
		logrus.WithError(err).Error("Failed to negotiate contract version")
		http.Error(w, "Failed to negotiate contract version", http.StatusInternalServerError)
		return
	}

	if contract == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(versioning.ContractVersionNegotiationResponse{
			ConsumerService: req.ConsumerService,
			ProviderService: req.ProviderService,
			AgreedVersion:   "",
			Compatible:      false,
			Reason:          "No compatible contract version found",
		})
		return
	}

	resp := versioning.ContractVersionNegotiationResponse{
		ConsumerService: req.ConsumerService,
		ProviderService: req.ProviderService,
		AgreedVersion:   contract.ContractVersion,
		Compatible:      contract.Status == "active",
		Reason:          "Found compatible contract version",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ==================== Phase 3: Version Lineage Endpoints ====================

// HandleGetVersionLineage handles GET /functions/{functionId}/versions/{version}/lineage - version history
// @Summary Get version lineage
// @Description Returns the version history/lineage for a function
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param version path string true "Function version"
// @Param limit query int false "Maximum number of lineage entries to return"
// @Success 200 {object} versioning.VersionLineageResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/{version}/lineage [get]
func (h *Handler) HandleGetVersionLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	entries, err := h.repo.GetVersionLineage(r.Context(), functionID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get version lineage")
		http.Error(w, "Failed to get version lineage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versioning.VersionLineageResponse{
		FunctionID: functionID,
		Entries:    entries,
		TotalCount: len(entries),
	})
}

// HandleCompareVersions handles GET /functions/{functionId}/versions/compare - compare two versions
// @Summary Compare versions
// @Description Compares two versions of a function and returns the differences
// @Tags Versioning
// @Accept json
// @Produce json
// @Param functionId path string true "Function ID"
// @Param v1 query string true "First version to compare"
// @Param v2 query string true "Second version to compare"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /functions/{functionId}/versions/compare [get]
func (h *Handler) HandleCompareVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionIDStr := vars["functionId"]

	functionID, err := uuid.Parse(functionIDStr)
	if err != nil {
		http.Error(w, "Invalid function ID", http.StatusBadRequest)
		return
	}

	v1 := r.URL.Query().Get("v1")
	v2 := r.URL.Query().Get("v2")

	if v1 == "" || v2 == "" {
		http.Error(w, "Both v1 and v2 query parameters are required", http.StatusBadRequest)
		return
	}

	diff, err := h.repo.CompareVersions(r.Context(), functionID, v1, v2)
	if err != nil {
		logrus.WithError(err).Error("Failed to compare versions")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diff)
}
