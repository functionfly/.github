package registry

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Total   int  `json:"total"`
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	HasMore bool `json:"has_more"`
}

// ChangelogResponse represents a changelog entry in API responses
type ChangelogResponse struct {
	ID                uuid.UUID                          `json:"id"`
	FunctionID        uuid.UUID                          `json:"function_id"`
	FunctionVersionID uuid.UUID                          `json:"function_version_id"`
	Version           string                             `json:"version"`
	PreviousVersion   *string                            `json:"previous_version,omitempty"`
	ChangeType        registry.ChangeType                `json:"change_type"`
	Category          registry.ChangeCategory            `json:"category"`
	Title             string                             `json:"title"`
	Description       string                             `json:"description"`
	Changes           []registry.FunctionChangelogChange `json:"changes"`
	Author            string                             `json:"author"`
	AuthorID          *uuid.UUID                         `json:"author_id,omitempty"`
	CreatedAt         string                             `json:"created_at"`
}

// ChangelogListResponse represents a paginated list of changelogs
type ChangelogListResponse struct {
	Data       []ChangelogResponse `json:"data"`
	Pagination PaginationInfo      `json:"pagination"`
}

// HandleGetChangelogs handles GET /registry/functions/{author}/{name}/changelogs
// Returns all changelogs for a function
func (h *Handler) HandleGetChangelogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function")
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	changelogs, err := h.repo.GetFunctionVersionChangelogs(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get changelogs")
		apierror.WriteError(w, apierror.NewInternal("Failed to get changelogs"))
		return
	}

	response := ChangelogListResponse{
		Data:       make([]ChangelogResponse, len(changelogs)),
		Pagination: PaginationInfo{Total: len(changelogs)},
	}

	for i, c := range changelogs {
		response.Data[i] = convertChangelogToResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetChangelogByVersion handles GET /registry/functions/{author}/{name}/changelogs/{version}
// Returns changelog for a specific version
func (h *Handler) HandleGetChangelogByVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	version := vars["version"]

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function")
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	changelogs, err := h.repo.GetChangelogsByVersion(fn.ID, version)
	if err != nil {
		logrus.WithError(err).Error("Failed to get changelog")
		apierror.WriteError(w, apierror.NewInternal("Failed to get changelog"))
		return
	}

	if len(changelogs) == 0 {
		apierror.WriteError(w, apierror.NewNotFound("Changelog not found for this version"))
		return
	}

	response := ChangelogListResponse{
		Data:       make([]ChangelogResponse, len(changelogs)),
		Pagination: PaginationInfo{Total: len(changelogs)},
	}

	for i, c := range changelogs {
		response.Data[i] = convertChangelogToResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetChangelogByCategory handles GET /registry/functions/{author}/{name}/changelogs/category/{category}
// Returns changelogs filtered by category
func (h *Handler) HandleGetChangelogByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	category := registry.ChangeCategory(vars["category"])

	// Validate category
	validCategories := map[registry.ChangeCategory]bool{
		registry.ChangeCategoryFeature:       true,
		registry.ChangeCategoryBugFix:        true,
		registry.ChangeCategoryPerformance:   true,
		registry.ChangeCategorySecurity:      true,
		registry.ChangeCategoryDocumentation: true,
		registry.ChangeCategoryDependency:    true,
		registry.ChangeCategoryBreaking:      true,
		registry.ChangeCategoryOther:         true,
	}

	if !validCategories[category] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid category"))
		return
	}

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function")
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	changelogs, err := h.repo.GetChangelogsByCategory(fn.ID, category)
	if err != nil {
		logrus.WithError(err).Error("Failed to get changelogs by category")
		apierror.WriteError(w, apierror.NewInternal("Failed to get changelogs"))
		return
	}

	response := ChangelogListResponse{
		Data:       make([]ChangelogResponse, len(changelogs)),
		Pagination: PaginationInfo{Total: len(changelogs)},
	}

	for i, c := range changelogs {
		response.Data[i] = convertChangelogToResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetVersionHistory handles GET /registry/functions/{author}/{name}/history
// Returns complete version history with changelogs
func (h *Handler) HandleGetVersionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function")
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	// Get all versions
	versions, err := h.repo.ListFunctionVersions(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get versions")
		apierror.WriteError(w, apierror.NewInternal("Failed to get versions"))
		return
	}

	// Get all changelogs
	changelogs, err := h.repo.GetFunctionVersionChangelogs(fn.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get changelogs")
		apierror.WriteError(w, apierror.NewInternal("Failed to get changelogs"))
		return
	}

	// Create a map of version to changelog
	changelogMap := make(map[string]registry.FunctionVersionChangelog)
	for _, c := range changelogs {
		changelogMap[c.Version] = c
	}

	// Build version history response
	type VersionHistoryEntry struct {
		Version     string             `json:"version"`
		PublishedAt string             `json:"published_at"`
		Changelog   *ChangelogResponse `json:"changelog,omitempty"`
		Runtime     string             `json:"runtime"`
	}

	history := make([]VersionHistoryEntry, len(versions))
	for i, v := range versions {
		history[i] = VersionHistoryEntry{
			Version:     v.Version,
			PublishedAt: v.PublishedAt.Format("2006-01-02T15:04:05Z"),
			Runtime:     v.Runtime,
		}

		if c, ok := changelogMap[v.Version]; ok {
			cr := convertChangelogToResponse(c)
			history[i].Changelog = &cr
		}
	}

	response := map[string]interface{}{
		"function_id":    fn.ID,
		"function_name":  fn.Name,
		"author":         fn.Author,
		"total_versions": len(versions),
		"history":        history,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// convertChangelogToResponse converts a FunctionVersionChangelog to a ChangelogResponse
func convertChangelogToResponse(c registry.FunctionVersionChangelog) ChangelogResponse {
	var changes []registry.FunctionChangelogChange
	if len(c.Changes) > 0 {
		json.Unmarshal(c.Changes, &changes)
	}

	return ChangelogResponse{
		ID:                c.ID,
		FunctionID:        c.FunctionID,
		FunctionVersionID: c.FunctionVersionID,
		Version:           c.Version,
		PreviousVersion:   c.PreviousVersion,
		ChangeType:        c.ChangeType,
		Category:          c.Category,
		Title:             c.Title,
		Description:       c.Description,
		Changes:           changes,
		Author:            c.Author,
		AuthorID:          c.AuthorID,
		CreatedAt:         c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
