package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	githubsvc "github.com/functionfly/functionfly/internal/services/github"
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

	if req.RepoID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "missing_repo", "repo_id is required")
		return
	}
	if len(req.FunctionNames) == 0 && req.FunctionName == "" {
		h.respondError(w, http.StatusBadRequest, "missing_name", "function_name or function_names is required")
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}

	functionNames := req.FunctionNames
	if len(functionNames) == 0 {
		functionNames = []string{req.FunctionName}
	}

	var detectedFuncs []githubsvc.DetectedFunction
	if repo, err := h.githubRepo.GetRepoByID(r.Context(), req.RepoID); err == nil && repo != nil && repo.DetectedFunctions != nil {
		json.Unmarshal(repo.DetectedFunctions, &detectedFuncs)
	}

	previewFunctions := make([]map[string]interface{}, 0, len(functionNames))
	for _, fnName := range functionNames {
		var detected githubsvc.DetectedFunction
		for _, df := range detectedFuncs {
			if df.Name == fnName {
				detected = df
				break
			}
		}
		if detected.Name == "" {
			detected = githubsvc.DetectedFunction{
				Name:       fnName,
				EntryPoint: fnName,
				Runtime:    "auto-detect",
				Confidence: 0.85,
				Strategy:   "single",
			}
		}

		previewFunctions = append(previewFunctions, map[string]interface{}{
			"name":                 detected.Name,
			"entry_point":          detected.EntryPoint,
			"confidence":           detected.Confidence,
			"strategy":             detected.Strategy,
			"sub_directory":        detected.SubDirectory,
			"file_count":           0,
			"branch":               req.Branch,
			"runtime":              detected.Runtime,
			"visibility":           req.Visibility,
			"estimated_size_bytes": 0,
			"estimated_cost_usd":   0.0,
			"has_conflict":         false,
			"conflict_type":        "none",
		})
	}

	preview := map[string]interface{}{
		"repo_id":                  req.RepoID.String(),
		"repo_full_name":           "",
		"functions":                previewFunctions,
		"total_file_count":         0,
		"total_size_bytes":         0,
		"total_estimated_cost_usd": 0.0,
		"warnings":                 []string{},
		"conflicts":                []map[string]interface{}{},
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	repo, err := h.githubRepo.GetRepoByID(r.Context(), req.RepoID)
	if err != nil || repo == nil || repo.ConnectionID != conn.ID {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	vault, err := githubsvc.NewTokenVault(h.vaultKey)
	if err != nil {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	token, err := vault.Decrypt(conn.EncryptedToken, conn.TokenIV, conn.TokenTag)
	if err != nil {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	ghClient := githubsvc.NewClient(token, githubsvc.WithLogger(h.logger))

	branch := req.Branch
	if branch == "" {
		branch = repo.DefaultBranch
	}

	branches, err := ghClient.ListBranches(r.Context(), repo.Owner, repo.Name)
	if err != nil {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	var branchSHA string
	for _, b := range branches {
		if b.Name == branch {
			branchSHA = b.Commit.SHA
			break
		}
	}
	if branchSHA == "" {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	tree, err := ghClient.GetTree(r.Context(), repo.Owner, repo.Name, branchSHA, true)
	if err != nil {
		h.respondJSON(w, http.StatusOK, preview)
		return
	}

	var totalSize int64
	var sourceFiles []string
	for _, entry := range tree.Tree {
		if entry.Type == "blob" {
			totalSize += int64(entry.Size)
			ext := strings.ToLower(filepath.Ext(entry.Path))
			if ext == ".js" || ext == ".ts" || ext == ".mjs" || ext == ".cjs" ||
				ext == ".jsx" || ext == ".tsx" || ext == ".py" || ext == ".go" ||
				ext == ".rs" || ext == ".java" || ext == ".rb" || ext == ".php" {
				sourceFiles = append(sourceFiles, entry.Path)
			}
		}
	}

	runtime := detectRuntimeFromTree(tree.Tree, repo)
	if req.RuntimeOverride != nil && *req.RuntimeOverride != "" {
		runtime = *req.RuntimeOverride
	}

	estimatedCost := calculateImportCost(len(sourceFiles), totalSize, runtime)

	preview["repo_full_name"] = repo.FullName
	funcSlice := preview["functions"].([]map[string]interface{})
	perFunctionCost := estimatedCost / float64(len(funcSlice))
	for i := range funcSlice {
		funcSlice[i]["runtime"] = runtime
		funcSlice[i]["file_count"] = len(sourceFiles)
		funcSlice[i]["estimated_size_bytes"] = totalSize
		funcSlice[i]["estimated_cost_usd"] = perFunctionCost
	}
	preview["total_file_count"] = len(sourceFiles)
	preview["total_size_bytes"] = totalSize
	preview["total_estimated_cost_usd"] = estimatedCost

	if len(sourceFiles) == 0 {
		warnings := preview["warnings"].([]string)
		preview["warnings"] = append(warnings, "No source files detected in repository")
	}

	h.respondJSON(w, http.StatusOK, preview)
}

// HandlePreviewBulkImport performs a dry-run for multiple import requests.
func (h *Handler) HandlePreviewBulkImport(w http.ResponseWriter, r *http.Request) {
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
		h.respondError(w, http.StatusBadRequest, "too_many_imports", "Maximum 20 imports per bulk preview")
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil || conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	previews := make([]map[string]interface{}, 0, len(req.Imports))

	for _, impReq := range req.Imports {
		preview := h.buildPreviewForImport(r.Context(), conn, &impReq)
		previews = append(previews, preview)
	}

	h.respondJSON(w, http.StatusOK, previews)
}

func (h *Handler) buildPreviewForImport(ctx context.Context, conn *storage.GitHubConnection, req *ImportRequest) map[string]interface{} {
	preview := map[string]interface{}{
		"repo_id":                  req.RepoID.String(),
		"repo_full_name":           "",
		"functions":                []map[string]interface{}{},
		"total_file_count":         0,
		"total_size_bytes":         0,
		"total_estimated_cost_usd": 0.0,
		"warnings":                 []string{},
		"conflicts":                []map[string]interface{}{},
	}

	repo, err := h.githubRepo.GetRepoByID(ctx, req.RepoID)
	if err != nil || repo == nil || repo.ConnectionID != conn.ID {
		preview["warnings"] = append(preview["warnings"].([]string), "Repository not found or access denied")
		return preview
	}

	preview["repo_full_name"] = repo.FullName

	vault, err := githubsvc.NewTokenVault(h.vaultKey)
	if err != nil {
		preview["warnings"] = append(preview["warnings"].([]string), "Failed to decrypt token")
		return preview
	}

	token, err := vault.Decrypt(conn.EncryptedToken, conn.TokenIV, conn.TokenTag)
	if err != nil {
		preview["warnings"] = append(preview["warnings"].([]string), "Failed to decrypt token")
		return preview
	}

	ghClient := githubsvc.NewClient(token, githubsvc.WithLogger(h.logger))

	branch := req.Branch
	if branch == "" {
		branch = repo.DefaultBranch
	}

	branches, err := ghClient.ListBranches(ctx, repo.Owner, repo.Name)
	if err != nil {
		preview["warnings"] = append(preview["warnings"].([]string), fmt.Sprintf("Failed to list branches: %v", err))
		return preview
	}

	var branchSHA string
	for _, b := range branches {
		if b.Name == branch {
			branchSHA = b.Commit.SHA
			break
		}
	}
	if branchSHA == "" {
		preview["warnings"] = append(preview["warnings"].([]string), fmt.Sprintf("Branch %q not found", branch))
		return preview
	}

	tree, err := ghClient.GetTree(ctx, repo.Owner, repo.Name, branchSHA, true)
	if err != nil {
		preview["warnings"] = append(preview["warnings"].([]string), fmt.Sprintf("Failed to fetch tree: %v", err))
		return preview
	}

	var totalSize int64
	var sourceFiles []string
	for _, entry := range tree.Tree {
		if entry.Type == "blob" {
			totalSize += int64(entry.Size)
			ext := strings.ToLower(filepath.Ext(entry.Path))
			if ext == ".js" || ext == ".ts" || ext == ".mjs" || ext == ".cjs" ||
				ext == ".jsx" || ext == ".tsx" || ext == ".py" || ext == ".go" ||
				ext == ".rs" || ext == ".java" || ext == ".rb" || ext == ".php" {
				sourceFiles = append(sourceFiles, entry.Path)
			}
		}
	}

	runtime := detectRuntimeFromTree(tree.Tree, repo)
	if req.RuntimeOverride != nil && *req.RuntimeOverride != "" {
		runtime = *req.RuntimeOverride
	}

	estimatedCost := calculateImportCost(len(sourceFiles), totalSize, runtime)

	var detectedFuncs []githubsvc.DetectedFunction
	if repo.DetectedFunctions != nil {
		json.Unmarshal(repo.DetectedFunctions, &detectedFuncs)
	}

	functionNames := req.FunctionNames
	if len(functionNames) == 0 && req.FunctionName != "" {
		functionNames = []string{req.FunctionName}
	}

	previewFunctions := make([]map[string]interface{}, 0, len(functionNames))
	for _, fnName := range functionNames {
		var detected githubsvc.DetectedFunction
		for _, df := range detectedFuncs {
			if df.Name == fnName {
				detected = df
				break
			}
		}
		if detected.Name == "" {
			detected = githubsvc.DetectedFunction{
				Name:       fnName,
				EntryPoint: fnName,
				Runtime:    "auto-detect",
				Confidence: 0.85,
				Strategy:   "single",
			}
		}

		previewFunctions = append(previewFunctions, map[string]interface{}{
			"name":                 detected.Name,
			"entry_point":          detected.EntryPoint,
			"confidence":           detected.Confidence,
			"strategy":             detected.Strategy,
			"sub_directory":        detected.SubDirectory,
			"file_count":           len(sourceFiles),
			"branch":               branch,
			"runtime":              runtime,
			"visibility":           req.Visibility,
			"estimated_size_bytes": totalSize,
			"estimated_cost_usd":   estimatedCost / float64(len(functionNames)),
			"has_conflict":         false,
			"conflict_type":        "none",
		})
	}

	preview["functions"] = previewFunctions
	preview["total_file_count"] = len(sourceFiles)
	preview["total_size_bytes"] = totalSize
	preview["total_estimated_cost_usd"] = estimatedCost

	if len(sourceFiles) == 0 {
		preview["warnings"] = append(preview["warnings"].([]string), "No source files detected in repository")
	}

	return preview
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
		completeEvent := map[string]interface{}{
			"stage":          "completed",
			"progress":       100,
			"function_id":    imp.FunctionID,
			"function_name":  imp.FunctionName,
			"commit_sha":     imp.CommitSHA,
			"files_imported": imp.FilesImported,
		}
		if imp.FunctionID != nil && h.registryRepo != nil {
			if fn, err := h.registryRepo.GetFunctionByID(*imp.FunctionID); err == nil && fn != nil {
				completeEvent["author"] = fn.Author
			}
		}
		fmt.Fprintf(w, "event: complete\ndata: %s\n\n", mustJSON(completeEvent))
		flusher.Flush()
		return
	}
	if imp.Status == "failed" {
		errMsg := ""
		if imp.ErrorMessage != nil {
			errMsg = *imp.ErrorMessage
		}
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", mustJSON(map[string]interface{}{
			"stage":    "error",
			"progress": imp.Progress,
			"message":  errMsg,
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
					"stage":    "error",
					"progress": event.Progress,
					"message":  event.Message,
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
	resp := &ImportDetailResponse{
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
	if imp.FunctionID != nil && h.registryRepo != nil {
		if fn, err := h.registryRepo.GetFunctionByID(*imp.FunctionID); err == nil && fn != nil {
			resp.FunctionAuthor = fn.Author
		}
	}
	return resp
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func detectRuntimeFromTree(tree []githubsvc.GitHubTreeEntry, repo *storage.GitHubRepo) string {
	hasFile := func(names ...string) bool {
		for _, entry := range tree {
			if entry.Type != "blob" {
				continue
			}
			for _, name := range names {
				if strings.HasSuffix(entry.Path, "/"+name) || entry.Path == name {
					return true
				}
			}
		}
		return false
	}

	if hasFile("package.json", "tsconfig.json") {
		if hasFile("tsconfig.json") {
			return "node18-typescript"
		}
		return "node18"
	}
	if hasFile("requirements.txt", "pyproject.toml", "setup.py", "Pipfile") {
		return "python3.11"
	}
	if hasFile("go.mod") {
		return "go1.22"
	}
	if hasFile("Cargo.toml") {
		return "rust1.75"
	}

	if repo != nil && repo.DetectedRuntime != nil && *repo.DetectedRuntime != "" {
		return *repo.DetectedRuntime
	}

	return "node18"
}

func calculateImportCost(sourceFileCount int, totalSizeBytes int64, runtime string) float64 {
	baseCost := 0.01

	perFileCost := 0.002
	fileCost := float64(sourceFileCount) * perFileCost

	sizeMB := float64(totalSizeBytes) / (1024 * 1024)
	var sizeCost float64
	switch {
	case sizeMB <= 1:
		sizeCost = 0.0
	case sizeMB <= 10:
		sizeCost = 0.005
	case sizeMB <= 50:
		sizeCost = 0.015
	default:
		sizeCost = 0.03
	}

	var runtimeCost float64
	switch runtime {
	case "python3.11", "node18", "node18-typescript":
		runtimeCost = 0.0
	case "go1.22":
		runtimeCost = 0.002
	case "rust1.75":
		runtimeCost = 0.003
	default:
		runtimeCost = 0.001
	}

	return baseCost + fileCost + sizeCost + runtimeCost
}
