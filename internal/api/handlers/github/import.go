package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// HandleImport starts a single import job.
func (h *Handler) HandleImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if req.RepoID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "missing_repo", "repo_id is required")
		return
	}
	if req.FunctionName == "" {
		h.respondError(w, http.StatusBadRequest, "missing_name", "function_name is required")
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.Visibility == "" {
		req.Visibility = "private"
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	repo, err := h.githubRepo.GetRepoByID(r.Context(), req.RepoID)
	if err != nil || repo == nil {
		h.respondError(w, http.StatusNotFound, "repo_not_found", "Repository not found")
		return
	}
	if repo.ConnectionID != conn.ID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied to this repository")
		return
	}

	syncBranches := req.SyncBranches
	if syncBranches == nil {
		syncBranches = json.RawMessage(`["main"]`)
	}
	manifestOverrides := req.ManifestOverrides
	if manifestOverrides == nil {
		manifestOverrides = json.RawMessage(`{}`)
	}

	imp := &storage.GitHubImport{
		UserID:            claims.UserID,
		TenantID:          claims.TenantID,
		ConnectionID:      conn.ID,
		RepoID:            req.RepoID,
		SourceBranch:      req.Branch,
		SourcePath:        req.SourcePath,
		FunctionName:      req.FunctionName,
		Visibility:        req.Visibility,
		RuntimeOverride:   req.RuntimeOverride,
		ManifestOverrides: manifestOverrides,
		AutoSyncEnabled:   req.AutoSync,
		SyncBranches:      syncBranches,
		Status:            "pending",
		Progress:          0,
	}

	created, err := h.githubRepo.CreateImport(r.Context(), imp)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create import")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to create import")
		return
	}

	go h.runImportPipeline(created.ID)

	h.respondJSON(w, http.StatusAccepted, h.mapImportResponse(created))
}

// HandlePreviewImport performs a dry-run of the import (no actual creation).
func (h *Handler) HandlePreviewImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// Build a preview response showing what would be created
	preview := map[string]interface{}{
		"functions": []map[string]interface{}{
			{
				"name":               req.FunctionName,
				"runtime":            "auto-detect",
				"visibility":         req.Visibility,
				"estimated_size_bytes": 0,
				"estimated_cost_usd": 0.0,
				"has_conflict":       false,
				"conflict_type":      "none",
			},
		},
		"total_estimated_cost_usd": 0.0,
		"warnings":                 []string{},
		"conflicts":                []map[string]interface{}{},
	}

	// Check for conflicts
	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err == nil && conn != nil {
		repo, err := h.githubRepo.GetRepoByID(r.Context(), req.RepoID)
		if err == nil && repo != nil {
			preview["functions"].([]map[string]interface{})[0]["runtime"] = repo.DetectedRuntime
		}
	}

	h.respondJSON(w, http.StatusOK, preview)
}

// HandleBulkImport starts multiple import jobs.
func (h *Handler) HandleBulkImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	var req BulkImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if len(req.Imports) == 0 {
		h.respondError(w, http.StatusBadRequest, "empty_imports", "At least one import is required")
		return
	}
	if len(req.Imports) > 20 {
		h.respondError(w, http.StatusBadRequest, "too_many_imports", "Maximum 20 imports per bulk request")
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	var results []ImportResponse
	for _, impReq := range req.Imports {
		if impReq.RepoID == uuid.Nil || impReq.FunctionName == "" {
			continue
		}
		if impReq.Branch == "" {
			impReq.Branch = "main"
		}
		if impReq.Visibility == "" {
			impReq.Visibility = "private"
		}

		repo, err := h.githubRepo.GetRepoByID(r.Context(), impReq.RepoID)
		if err != nil || repo == nil || repo.ConnectionID != conn.ID {
			continue
		}

		syncBranches := impReq.SyncBranches
		if syncBranches == nil {
			syncBranches = json.RawMessage(`["main"]`)
		}
		manifestOverrides := impReq.ManifestOverrides
		if manifestOverrides == nil {
			manifestOverrides = json.RawMessage(`{}`)
		}

		imp := &storage.GitHubImport{
			UserID:            claims.UserID,
			TenantID:          claims.TenantID,
			ConnectionID:      conn.ID,
			RepoID:            impReq.RepoID,
			SourceBranch:      impReq.Branch,
			SourcePath:        impReq.SourcePath,
			FunctionName:      impReq.FunctionName,
			Visibility:        impReq.Visibility,
			RuntimeOverride:   impReq.RuntimeOverride,
			ManifestOverrides: manifestOverrides,
			AutoSyncEnabled:   impReq.AutoSync,
			SyncBranches:      syncBranches,
			Status:            "pending",
			Progress:          0,
		}

		created, err := h.githubRepo.CreateImport(r.Context(), imp)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create bulk import")
			continue
		}

		go h.runImportPipeline(created.ID)

		results = append(results, ImportResponse{
			ImportID: created.ID,
			Status:   "pending",
		})
	}

	h.respondJSON(w, http.StatusAccepted, BulkImportResponse{Imports: results})
}

// HandleListImports lists all imports for the authenticated user.
func (h *Handler) HandleListImports(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	params := storage.ListImportsParams{
		Page:    1,
		PerPage: 20,
		Status:  r.URL.Query().Get("status"),
	}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			params.Page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			params.PerPage = n
		}
	}
	if v := r.URL.Query().Get("repo_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			params.RepoID = &id
		}
	}

	imports, total, err := h.githubRepo.ListImportsByUser(r.Context(), claims.UserID, params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list imports")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to list imports")
		return
	}

	results := make([]*ImportDetailResponse, len(imports))
	for i, imp := range imports {
		results[i] = h.mapImportResponse(imp)
	}

	h.respondJSON(w, http.StatusOK, ListImportsResponse{
		Imports: results,
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
	})
}

// HandleGetImport returns a single import by ID.
func (h *Handler) HandleGetImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	importID, err := h.parseUUID(r, "importId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid import ID")
		return
	}

	imp, err := h.githubRepo.GetImportByID(r.Context(), importID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get import")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to get import")
		return
	}
	if imp == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Import not found")
		return
	}
	if imp.UserID != claims.UserID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	h.respondJSON(w, http.StatusOK, h.mapImportResponse(imp))
}

// HandleCancelImport cancels a running import.
func (h *Handler) HandleCancelImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	importID, err := h.parseUUID(r, "importId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid import ID")
		return
	}

	imp, err := h.githubRepo.GetImportByID(r.Context(), importID)
	if err != nil || imp == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Import not found")
		return
	}
	if imp.UserID != claims.UserID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	if imp.Status == "completed" || imp.Status == "failed" || imp.Status == "cancelled" {
		h.respondError(w, http.StatusBadRequest, "invalid_status", fmt.Sprintf("Cannot cancel import in %s state", imp.Status))
		return
	}

	if err := h.githubRepo.UpdateImportStatus(r.Context(), importID, "cancelled", imp.Progress, "Cancelled by user"); err != nil {
		h.logger.WithError(err).Error("Failed to cancel import")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to cancel import")
		return
	}

	h.completeProgress(importID)

	// Re-fetch to get updated status
	updated, _ := h.githubRepo.GetImportByID(r.Context(), importID)
	if updated != nil {
		h.respondJSON(w, http.StatusOK, h.mapImportResponse(updated))
	} else {
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
	}
}

// HandleRetryImport retries a failed import.
func (h *Handler) HandleRetryImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	importID, err := h.parseUUID(r, "importId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid import ID")
		return
	}

	imp, err := h.githubRepo.GetImportByID(r.Context(), importID)
	if err != nil || imp == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Import not found")
		return
	}
	if imp.UserID != claims.UserID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	if imp.Status != "failed" {
		h.respondError(w, http.StatusBadRequest, "invalid_status", "Only failed imports can be retried")
		return
	}

	if err := h.githubRepo.UpdateImportStatus(r.Context(), importID, "pending", 0, ""); err != nil {
		h.logger.WithError(err).Error("Failed to reset import status")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to retry import")
		return
	}

	go h.runImportPipeline(importID)

	updated, _ := h.githubRepo.GetImportByID(r.Context(), importID)
	if updated != nil {
		h.respondJSON(w, http.StatusOK, h.mapImportResponse(updated))
	} else {
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "pending"})
	}
}

// HandleResyncImport re-runs the import pipeline for a completed import.
func (h *Handler) HandleResyncImport(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	importID, err := h.parseUUID(r, "importId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid import ID")
		return
	}

	imp, err := h.githubRepo.GetImportByID(r.Context(), importID)
	if err != nil || imp == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Import not found")
		return
	}
	if imp.UserID != claims.UserID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	if err := h.githubRepo.UpdateImportStatus(r.Context(), importID, "pending", 0, ""); err != nil {
		h.logger.WithError(err).Error("Failed to reset import status")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to resync import")
		return
	}

	go h.runImportPipeline(importID)

	updated, _ := h.githubRepo.GetImportByID(r.Context(), importID)
	if updated != nil {
		h.respondJSON(w, http.StatusOK, h.mapImportResponse(updated))
	} else {
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "pending"})
	}
}

// HandleImportProgress is an SSE endpoint for real-time progress.
// Since EventSource doesn't support custom headers, auth is via ?token= query param.
func (h *Handler) HandleImportProgress(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuthOrToken(w, r)
	if claims == nil {
		return
	}

	importID, err := h.parseUUID(r, "importId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid import ID")
		return
	}

	imp, err := h.githubRepo.GetImportByID(r.Context(), importID)
	if err != nil || imp == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "Import not found")
		return
	}
	if imp.UserID != claims.UserID {
		h.respondError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.respondError(w, http.StatusInternalServerError, "sse_error", "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := h.getProgressChan(importID)

	if imp.Status == "completed" {
		fmt.Fprintf(w, "event: complete\ndata: %s\n\n", mustJSON(map[string]interface{}{
			"stage":        "completed",
			"progress":     100,
			"function_id":  imp.FunctionID,
			"function_name": imp.FunctionName,
			"commit_sha":   imp.CommitSHA,
			"files_imported": imp.FilesImported,
		}))
		flusher.Flush()
		return
	}
	if imp.Status == "failed" {
		errMsg := ""
		if imp.ErrorMessage != nil {
			errMsg = *imp.ErrorMessage
		}
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", mustJSON(map[string]interface{}{
			"stage":     "error",
			"progress":  imp.Progress,
			"message":   errMsg,
		}))
		flusher.Flush()
		return
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}

			if event.Stage == "error" {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", mustJSON(map[string]interface{}{
					"stage":   "error",
					"progress": event.Progress,
					"message": event.Message,
				}))
				flusher.Flush()
				return
			}

			data := mustJSON(event)
			fmt.Fprintf(w, "event: progress\ndata: %s\n\n", data)
			flusher.Flush()

			if event.Progress >= 100 {
				completedImp, _ := h.githubRepo.GetImportByID(ctx, importID)
				completeData := map[string]interface{}{
					"stage":    "completed",
					"progress": 100,
				}
				if completedImp != nil {
					completeData["function_id"] = completedImp.FunctionID
					completeData["function_name"] = completedImp.FunctionName
					completeData["commit_sha"] = completedImp.CommitSHA
					completeData["files_imported"] = completedImp.FilesImported
				}
				fmt.Fprintf(w, "event: complete\ndata: %s\n\n", mustJSON(completeData))
				flusher.Flush()
				return
			}
		}
	}
}

func (h *Handler) mapImportResponse(imp *storage.GitHubImport) *ImportDetailResponse {
	return &ImportDetailResponse{
		ID:                imp.ID,
		RepoID:            imp.RepoID,
		FunctionName:      imp.FunctionName,
		SourceBranch:      imp.SourceBranch,
		SourcePath:        imp.SourcePath,
		Visibility:        imp.Visibility,
		AutoSyncEnabled:   imp.AutoSyncEnabled,
		SyncBranches:      imp.SyncBranches,
		Status:            imp.Status,
		Progress:          imp.Progress,
		ErrorMessage:      imp.ErrorMessage,
		FunctionID:        imp.FunctionID,
		FunctionVersionID: imp.FunctionVersionID,
		CommitSHA:         imp.CommitSHA,
		FilesImported:     imp.FilesImported,
		TotalSizeBytes:    imp.TotalSizeBytes,
		CreatedAt:         imp.CreatedAt,
		UpdatedAt:         imp.UpdatedAt,
		CompletedAt:       imp.CompletedAt,
	}
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
