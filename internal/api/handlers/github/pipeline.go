package github

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/functionfly/functionfly/internal/services/github"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// pipelineStep tracks which steps have been completed for saga rollback.
type pipelineStep string

const (
	stepConnectionLoaded  pipelineStep = "connection_loaded"
	stepBranchResolved    pipelineStep = "branch_resolved"
	stepTreeFetched       pipelineStep = "tree_fetched"
	stepScanCompleted     pipelineStep = "scan_completed"
	stepFunctionCreated   pipelineStep = "function_created"
	stepWebhookRegistered pipelineStep = "webhook_registered"
)

// runImportPipeline executes the full import pipeline for a GitHub import job.
// Uses a saga pattern: each step is tracked, and on failure completed steps are rolled back.
func (h *Handler) runImportPipeline(ctx context.Context, importID uuid.UUID) {
	log := h.logger.WithField("import_id", importID)

	log.Info("Starting import pipeline")

	var completedSteps []pipelineStep
	var imp *storage.GitHubImport
	var functionID uuid.UUID
	var functionCreatedByThisImport bool

	var rollbackFn = func() {
		if imp == nil {
			return
		}
		log.WithField("completed_steps", completedSteps).Info("Rolling back import pipeline")
		for i := len(completedSteps) - 1; i >= 0; i-- {
			step := completedSteps[i]
			switch step {
			case stepWebhookRegistered:
				if wh, whErr := h.githubRepo.GetWebhookByRepoID(ctx, imp.RepoID); whErr == nil && wh != nil {
					if wh.GithubWebhookID != nil {
						if ghClient, clientErr := h.getGitHubClient(ctx, imp.UserID); clientErr == nil {
							repo, repoErr := h.githubRepo.GetRepoByID(ctx, imp.RepoID)
							if repoErr == nil && repo != nil {
								_ = ghClient.DeleteWebhook(ctx, repo.Owner, repo.Name, *wh.GithubWebhookID)
							}
						}
					}
					_ = h.githubRepo.DeleteWebhook(ctx, wh.ID)
					log.Info("Rollback: removed webhook")
				}
			case stepFunctionCreated:
				if functionID != uuid.Nil && functionCreatedByThisImport {
					if fn, getErr := h.registryRepo.GetFunctionByID(ctx, functionID); getErr == nil {
						if delErr := h.registryRepo.DeleteFunction(ctx, fn.Author, fn.Name); delErr != nil {
							log.WithError(delErr).WithField("function_id", functionID).Warn("Rollback: failed to delete function")
						} else {
							log.WithField("function_id", functionID).Info("Rollback: deleted function")
						}
					} else {
						log.WithError(getErr).WithField("function_id", functionID).Warn("Rollback: failed to get function for deletion")
					}
				}
				_ = h.githubRepo.UpdateRepoImportStatus(ctx, imp.RepoID, "not_imported")
				log.Info("Rollback: reset repo import status")
			}
		}
	}

	h.sendProgress(ctx, importID, "fetching", 5, "Loading import record")

	var err error
	imp, err = h.githubRepo.GetImportByID(ctx, importID)
	if err != nil || imp == nil {
		h.failImport(ctx, importID, "Import record not found")
		return
	}

	if imp.Status == "cancelled" {
		log.Info("Import was cancelled, aborting pipeline")
		h.completeProgress(importID)
		return
	}

	// Check tenant rate limit
	if !h.checkImportRateLimit(ctx, imp.TenantID) {
		h.failImport(ctx, importID, "Import rate limit exceeded. Maximum 10 concurrent imports per tenant.")
		return
	}
	defer h.releaseImportSlot(imp.TenantID)

	h.sendProgress(ctx, importID, "fetching", 10, "Loading repository and connection")

	repo, err := h.githubRepo.GetRepoByID(ctx, imp.RepoID)
	if err != nil || repo == nil {
		h.failImport(ctx, importID, "Repository not found")
		return
	}

	conn, err := h.githubRepo.GetConnectionByID(ctx, imp.ConnectionID)
	if err != nil || conn == nil {
		h.failImport(ctx, importID, "GitHub connection not found")
		return
	}

	vault, err := newVault(h.vaultKey)
	if err != nil {
		h.failImport(ctx, importID, "Internal encryption error")
		return
	}

	token, err := vault.Decrypt(conn.EncryptedToken, conn.TokenIV, conn.TokenTag)
	if err != nil {
		h.failImport(ctx, importID, "Failed to decrypt GitHub token")
		return
	}
	completedSteps = append(completedSteps, stepConnectionLoaded)

	ghClient := github.NewClient(token, github.WithLogger(h.logger))

	h.sendProgress(ctx, importID, "fetching", 20, "Fetching source files from GitHub")

	branch := imp.SourceBranch
	if branch == "" {
		branch = repo.DefaultBranch
	}

	branches, err := ghClient.ListBranches(ctx, repo.Owner, repo.Name)
	if err != nil {
		rollbackFn()
		h.failImport(ctx, importID, fmt.Sprintf("Failed to list branches: %v", err))
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
		rollbackFn()
		h.failImport(ctx, importID, fmt.Sprintf("Branch %q not found", branch))
		return
	}
	completedSteps = append(completedSteps, stepBranchResolved)

	h.sendProgress(ctx, importID, "fetching", 30, "Fetching file tree")

	tree, err := ghClient.GetTree(ctx, repo.Owner, repo.Name, branchSHA, true)
	if err != nil {
		rollbackFn()
		h.failImport(ctx, importID, fmt.Sprintf("Failed to fetch tree: %v", err))
		return
	}
	completedSteps = append(completedSteps, stepTreeFetched)

	h.sendProgress(ctx, importID, "analyzing", 40, "Detecting runtime")

	runtime := h.detectRuntime(tree.Tree, repo)

	h.sendProgress(ctx, importID, "analyzing", 50, "Scanning for functions")

	scanner := github.NewScanner(ghClient, h.logger)
	scanResult, err := scanner.ScanRepo(ctx, repo.Owner, repo.Name, branch)
	if err != nil {
		log.WithError(err).Warn("Scanner failed, using defaults")
		scanResult = &github.ScanResult{
			Functions:      []github.DetectedFunction{},
			PrimaryRuntime: runtime,
		}
	}
	completedSteps = append(completedSteps, stepScanCompleted)

	if len(scanResult.Functions) == 0 && imp.RuntimeOverride != nil {
		runtime = *imp.RuntimeOverride
	} else if scanResult.PrimaryRuntime != "" {
		runtime = scanResult.PrimaryRuntime
	}

	h.sendProgress(ctx, importID, "building", 60, "Preparing function metadata")

	manifestConfig := map[string]interface{}{
		"name":       imp.FunctionName,
		"runtime":    runtime,
		"source":     fmt.Sprintf("github:%s", repo.FullName),
		"branch":     branch,
		"commit":     branchSHA,
		"visibility": imp.Visibility,
	}

	if imp.ManifestOverrides != nil {
		var overrides map[string]interface{}
		if err := json.Unmarshal(imp.ManifestOverrides, &overrides); err == nil {
			for k, v := range overrides {
				manifestConfig[k] = v
			}
		}
	}

	h.sendProgress(ctx, importID, "publishing", 70, "Creating function in registry")

	// Resolve the author's username from user ID for proper registry authorship
	authorUsername, err := h.resolveAuthorUsername(ctx, imp.UserID)
	if err != nil {
		log.WithError(err).Warn("Failed to resolve author username, using fallback")
		authorUsername = "imported"
	}
	log.WithField("author_username", authorUsername).Info("Resolved author username")

	// Compute content hash for version tracking
	contentHash := computeContentHash(tree.Tree)

	// Check if function with same author/name already exists (idempotent import)
	log.WithFields(logrus.Fields{
		"author": authorUsername,
		"name":   imp.FunctionName,
	}).Info("Looking up existing function by author/name")
	existingFn, err := h.registryRepo.GetFunctionByAuthorName(ctx, authorUsername, imp.FunctionName)
	if err != nil {
		log.WithError(err).Warn("Failed to check for existing function")
	} else if existingFn != nil {
		log.WithFields(logrus.Fields{
			"existing_fn_id": existingFn.ID,
			"author":         existingFn.Author,
			"name":           existingFn.Name,
		}).Info("Found existing function")
	}

	if existingFn != nil {
		functionID = existingFn.ID
		functionCreatedByThisImport = false
		log.WithField("function_id", functionID).Info("Using existing function, will create new version")
	} else {
		// Create new function in registry
		fn := &storageregistry.RegistryFunction{
			Author:          authorUsername,
			Name:            imp.FunctionName,
			Title:           sql.NullString{String: imp.FunctionName, Valid: true},
			Visibility:      imp.Visibility,
			PopularityScore: 0,
			TenantID:        &imp.TenantID,
			OwnerUserID:     &imp.UserID,
			Tags:            json.RawMessage(`[]`),
		}

		// Apply manifest overrides for title, description, category
		if title, ok := manifestConfig["title"].(string); ok && title != "" {
			fn.Title = sql.NullString{String: title, Valid: true}
		}
		if desc, ok := manifestConfig["description"].(string); ok && desc != "" {
			fn.Description = sql.NullString{String: desc, Valid: true}
		}
		if cat, ok := manifestConfig["category"].(string); ok && cat != "" {
			fn.Category = sql.NullString{String: cat, Valid: true}
		}

		if err := h.registryRepo.CreateFunction(ctx, fn); err != nil {
			errStr := err.Error()
			if !strings.Contains(errStr, "duplicate key") && !strings.Contains(errStr, "23505") {
				rollbackFn()
				h.failImport(ctx, importID, fmt.Sprintf("Failed to create function in registry: %v", err))
				return
			}

			log.Info("Function already exists (concurrent or duplicate), fetching existing")
			existingFn, fetchErr := h.registryRepo.GetFunctionByAuthorName(ctx, authorUsername, imp.FunctionName)
			if fetchErr == nil && existingFn != nil {
				functionID = existingFn.ID
				functionCreatedByThisImport = false
				log.WithField("function_id", functionID).Info("Using existing function, will create new version")
			} else if errors.Is(fetchErr, gorm.ErrRecordNotFound) {
				rollbackFn()
				h.failImport(ctx, importID, fmt.Sprintf("Failed to create function: function already exists but could not be fetched: %v", err))
				return
			} else {
				rollbackFn()
				h.failImport(ctx, importID, fmt.Sprintf("Failed to create function (concurrent): %v", err))
				return
			}
		} else {
			functionID = fn.ID
			functionCreatedByThisImport = true
			log.WithField("function_id", functionID).Info("Created new function in registry")
		}
	}

	log.WithField("function_id", functionID).Info("Function ID to use for version")

	completedSteps = append(completedSteps, stepFunctionCreated)

	// Create function version
	versionID := uuid.New()
	manifestJSON, _ := json.Marshal(manifestConfig)

	runtimeStr := "node18"
	if rt, ok := manifestConfig["runtime"].(string); ok && rt != "" {
		runtimeStr = rt
	}

	version := &storageregistry.RegistryFunctionVersion{
		ID:          versionID,
		FunctionID:  functionID,
		Version:     "1.0.0",
		Manifest:    manifestJSON,
		Runtime:     runtimeStr,
		MemoryMB:    256,
		TimeoutMs:   30000,
		ContentHash: sql.NullString{String: contentHash, Valid: true},
	}

	// Set source code if available from tree
	if len(tree.Tree) > 0 {
		var sourceFiles []string
		for _, entry := range tree.Tree {
			if entry.Type != "blob" {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Path))
			if ext == ".js" || ext == ".ts" || ext == ".mjs" || ext == ".cjs" ||
				ext == ".jsx" || ext == ".tsx" || ext == ".py" || ext == ".go" ||
				ext == ".rs" || ext == ".java" || ext == ".rb" || ext == ".php" {
				sourceFiles = append(sourceFiles, entry.Path)
			}
		}

		if len(sourceFiles) > 0 {
			h.sendProgress(ctx, importID, "building", 73, "Fetching source code")

			var sourceCode strings.Builder
			maxFiles := 20
			maxFileSize := 50 * 1024
			totalFetched := 0
			totalSize := 0

			for _, filePath := range sourceFiles {
				if totalFetched >= maxFiles {
					fmt.Fprintf(&sourceCode, "\n// ... and %d more files\n", len(sourceFiles)-maxFiles)
					break
				}

				content, err := ghClient.GetFileContent(ctx, repo.Owner, repo.Name, filePath, branchSHA)
				if err != nil {
					log.WithError(err).WithField("file", filePath).Warn("Failed to fetch source file")
					fmt.Fprintf(&sourceCode, "// %s (failed to fetch)\n", filePath)
					continue
				}

				if totalSize+len(content) > maxFileSize {
					fmt.Fprintf(&sourceCode, "\n// %s (truncated, total size exceeded)\n", filePath)
					continue
				}

				totalSize += len(content)
				totalFetched++
				fmt.Fprintf(&sourceCode, "// %s\n%s\n\n", filePath, string(content))
			}

			if sourceCode.Len() > 0 {
				version.SourceCode = sql.NullString{String: sourceCode.String(), Valid: true}
				log.WithFields(logrus.Fields{
					"files_fetched": totalFetched,
					"total_size":    totalSize,
				}).Info("Source code fetched for function version")
			}
		}
	}

	if err := h.registryRepo.CreateFunctionVersion(version); err != nil {
		log.WithError(err).Error("Failed to create function version")
		rollbackFn()
		h.failImport(ctx, importID, fmt.Sprintf("Failed to create function version: %v", err))
		return
	}

	log.WithFields(logrus.Fields{
		"function_id": functionID,
		"version_id":  version.ID,
	}).Info("Created function and version in registry")

	h.sendProgress(ctx, importID, "publishing", 80, "Publishing function version")

	filesImported := len(tree.Tree)
	var totalSize int64
	for _, entry := range tree.Tree {
		totalSize += int64(entry.Size)
	}

	h.sendProgress(ctx, importID, "publishing", 90, "Finalizing import")

	if err := h.githubRepo.UpdateImportResult(ctx, importID, functionID, version.ID, branchSHA, contentHash, filesImported, totalSize); err != nil {
		rollbackFn()
		h.failImport(ctx, importID, fmt.Sprintf("Failed to save import result: %v", err))
		return
	}

	if err := h.githubRepo.UpdateRepoImportStatus(ctx, imp.RepoID, "imported"); err != nil {
		log.WithError(err).Warn("Failed to update repo import status")
	}

	if imp.AutoSyncEnabled {
		h.sendProgress(ctx, importID, "configuring", 95, "Registering webhook for auto-sync")

		webhookErr := h.ensureWebhook(ctx, imp)
		if webhookErr != nil {
			log.WithError(webhookErr).Warn("Webhook registration failed, auto-sync will not work")
			h.sendProgress(ctx, importID, "configuring", 97, "Webhook registration failed: "+webhookErr.Error())
		} else {
			completedSteps = append(completedSteps, stepWebhookRegistered)
		}
	}

	if err := ghClient.CreateCommitStatus(ctx, repo.Owner, repo.Name, branchSHA, &github.CommitStatusRequest{
		State:       "success",
		Description: fmt.Sprintf("Imported as %s", imp.FunctionName),
		Context:     "functionfly/import",
	}); err != nil {
		log.WithError(err).Warn("Failed to set commit status on GitHub")
	}

	h.sendProgress(ctx, importID, "completed", 100, "Import completed successfully")
	h.completeProgress(importID)

	log.WithFields(map[string]interface{}{
		"function_id":      functionID,
		"version_id":       versionID,
		"files_imported":   filesImported,
		"total_size_bytes": totalSize,
	}).Info("Import pipeline completed")
}

func (h *Handler) failImport(ctx context.Context, importID uuid.UUID, errMsg string) {
	h.logger.WithError(fmt.Errorf("%s", errMsg)).WithField("import_id", importID).Error("Import failed")

	if err := h.githubRepo.UpdateImportStatus(ctx, importID, "failed", 0, errMsg); err != nil {
		h.logger.WithError(err).Error("Failed to update import status to failed")
	}

	if v, ok := h.progressCh.Load(importID); ok {
		ch := v.(chan ProgressEvent)
		select {
		case ch <- ProgressEvent{Stage: "error", Progress: 0, Message: errMsg}:
		default:
		}
	}

	h.completeProgress(importID)
}

func (h *Handler) detectRuntime(tree []github.GitHubTreeEntry, repo *storage.GitHubRepo) string {
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

	return "node18"
}

func computeContentHash(tree []github.GitHubTreeEntry) string {
	h := sha256.New()
	for _, entry := range tree {
		if entry.Type == "blob" {
			fmt.Fprintf(h, "%s:%s:%d\n", entry.Path, entry.SHA, entry.Size)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func newVault(key string) (*github.TokenVault, error) {
	return github.NewTokenVault(key)
}
