package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
	"github.com/google/uuid"
)

// HandleWebhook receives GitHub webhook events.
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "read_error", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-Hub-Signature-256")
	event := r.Header.Get("X-GitHub-Event")
	deliveryID := r.Header.Get("X-GitHub-Delivery")

	if event == "" {
		h.respondError(w, http.StatusBadRequest, "missing_event", "Missing X-GitHub-Event header")
		return
	}

	if event == "ping" {
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "pong"})
		return
	}

	var repoID int64
	var repoFullName string
	var pushEvent githubsvc.WebhookPushEvent

	switch event {
	case "push":
		if err := json.Unmarshal(body, &pushEvent); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid push webhook payload")
			return
		}
		repoID = pushEvent.Repository.ID
		repoFullName = pushEvent.Repository.FullName
	case "pull_request":
		var prEvent githubsvc.WebhookPROvent
		if err := json.Unmarshal(body, &prEvent); err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid PR webhook payload")
			return
		}
		repoID = prEvent.Repository.ID
		repoFullName = prEvent.Repository.FullName
	default:
		h.respondJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	go h.processWebhookEvent(r.Context(), event, deliveryID, sig, repoID, repoFullName, body, pushEvent)

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func (h *Handler) processWebhookEvent(
	ctx context.Context,
	event, deliveryID, sig string,
	repoID int64,
	repoFullName string,
	body []byte,
	pushEvent githubsvc.WebhookPushEvent,
) {
	log := h.logger.WithFields(map[string]interface{}{
		"event":       event,
		"delivery_id": deliveryID,
		"repo":        repoFullName,
	})

	repoRecord, err := h.findRepoByGitHubID(ctx, repoID)
	if err != nil || repoRecord == nil {
		log.Debug("No matching repo found for webhook event")
		return
	}

	webhook, err := h.githubRepo.GetWebhookByRepoID(ctx, repoRecord.ID)
	if err != nil || webhook == nil || !webhook.IsActive {
		log.Debug("No active webhook found for repo")
		return
	}

	if sig != "" && !verifyHMAC(body, sig, webhook.WebhookSecret) {
		log.Warn("Webhook signature verification failed")
		_ = h.githubRepo.UpdateWebhookDelivery(ctx, webhook.ID, event, false, "signature verification failed")
		return
	}

	_ = h.githubRepo.UpdateWebhookDelivery(ctx, webhook.ID, event, true, "")

	switch event {
	case "push":
		h.handlePushEvent(ctx, repoRecord, pushEvent)
	case "pull_request":
		h.handlePREvent(ctx, repoRecord, body)
	}
}

func (h *Handler) handlePushEvent(ctx context.Context, repo *storage.GitHubRepo, event githubsvc.WebhookPushEvent) {
	branch := strings.TrimPrefix(event.Ref, "refs/heads/")

	logger := h.logger.WithFields(map[string]interface{}{
		"branch": branch,
		"after":  event.After,
	})

	conn, err := h.githubRepo.GetConnectionByID(ctx, repo.ConnectionID)
	if err != nil || conn == nil {
		return
	}

	imports, _, err := h.githubRepo.ListImportsByUser(ctx, conn.UserID, storage.ListImportsParams{
		PerPage: 100,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to list imports for webhook")
		return
	}

	for _, imp := range imports {
		if imp.RepoID != repo.ID {
			continue
		}
		if imp.Status != "completed" {
			continue
		}
		if !imp.AutoSyncEnabled {
			continue
		}

		var syncBranches []string
		if err := json.Unmarshal(imp.SyncBranches, &syncBranches); err != nil {
			syncBranches = []string{"main"}
		}

		matched := false
		for _, sb := range syncBranches {
			if sb == branch {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		logger.WithField("import_id", imp.ID).Info("Triggering auto-sync for push event")

		syncLog := &storage.GitHubSyncLog{
			ImportID:         imp.ID,
			FunctionID:       imp.FunctionID,
			TriggerType:      "webhook",
			TriggerBranch:    &branch,
			TriggerCommitSHA: &event.After,
			Status:           "running",
		}

		createdLog, err := h.githubRepo.CreateSyncLog(ctx, syncLog)
		if err != nil {
			logger.WithError(err).Error("Failed to create sync log")
			continue
		}

		_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 0, "")

		go h.runSyncPipeline(ctx, imp, createdLog, event.After, branch)
	}
}

func (h *Handler) handlePREvent(ctx context.Context, repo *storage.GitHubRepo, body []byte) {
	var prEvent githubsvc.WebhookPROvent
	if err := json.Unmarshal(body, &prEvent); err != nil {
		return
	}

	if prEvent.Action != "opened" && prEvent.Action != "synchronize" {
		return
	}

	headSHA := prEvent.PullRequest.Head.SHA
	branch := prEvent.PullRequest.Head.Ref
	prNumber := prEvent.PullRequest.Number

	conn, err := h.githubRepo.GetConnectionByID(ctx, repo.ConnectionID)
	if err != nil || conn == nil {
		return
	}

	imports, _, err := h.githubRepo.ListImportsByUser(ctx, conn.UserID, storage.ListImportsParams{
		PerPage: 100,
	})
	if err != nil {
		return
	}

	for _, imp := range imports {
		if imp.RepoID != repo.ID || imp.Status != "completed" || !imp.AutoSyncEnabled {
			continue
		}

		syncLog := &storage.GitHubSyncLog{
			ImportID:         imp.ID,
			FunctionID:       imp.FunctionID,
			TriggerType:      "pull_request",
			TriggerBranch:    &branch,
			TriggerCommitSHA: &headSHA,
			TriggerPRNumber:  &prNumber,
			Status:           "running",
		}

		createdLog, err := h.githubRepo.CreateSyncLog(ctx, syncLog)
		if err != nil {
			continue
		}

		go h.runSyncPipeline(ctx, imp, createdLog, headSHA, branch)
	}
}

func (h *Handler) findRepoByGitHubID(ctx context.Context, githubRepoID int64) (*storage.GitHubRepo, error) {
	return h.githubRepo.GetRepoByGitHubRepoID(ctx, githubRepoID)
}

func (h *Handler) runSyncPipeline(ctx context.Context, imp *storage.GitHubImport, syncLog *storage.GitHubSyncLog, commitSHA, branch string) {
	startTime := time.Now().UTC()

	h.logger.WithFields(map[string]interface{}{
		"import_id": imp.ID,
		"log_id":    syncLog.ID,
		"commit":    commitSHA,
		"branch":    branch,
	}).Info("Starting sync pipeline")

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 10, "Fetching latest code")

	repo, err := h.githubRepo.GetRepoByID(ctx, imp.RepoID)
	if err != nil || repo == nil {
		h.finishSyncWithError(ctx, syncLog, imp, "Repository not found")
		return
	}

	ghClient, err := h.getGitHubClient(ctx, imp.UserID)
	if err != nil {
		h.finishSyncWithError(ctx, syncLog, imp, fmt.Sprintf("Failed to create GitHub client: %v", err))
		return
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 30, "Fetching file tree")

	tree, err := ghClient.GetTree(ctx, repo.Owner, repo.Name, commitSHA, true)
	if err != nil {
		h.finishSyncWithError(ctx, syncLog, imp, fmt.Sprintf("Failed to fetch tree: %v", err))
		return
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 50, "Analyzing changes")

	contentHash := computeContentHash(tree.Tree)

	if imp.FunctionID == nil {
		h.finishSyncWithError(ctx, syncLog, imp, "No function associated with import")
		return
	}

	if h.registryRepo == nil {
		h.finishSyncWithError(ctx, syncLog, imp, "Registry repository not configured")
		return
	}

	fn, err := h.registryRepo.GetFunctionByID(*imp.FunctionID)
	if err != nil || fn == nil {
		h.finishSyncWithError(ctx, syncLog, imp, "Function not found in registry")
		return
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 60, "Creating new version")

	latestVersion, err := h.registryRepo.GetLatestFunctionVersion(*imp.FunctionID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get latest version, proceeding with default")
	}

	nextVersion := "1.0.0"
	if latestVersion != nil {
		nextVersion = incrementVersion(latestVersion.Version)
	}

	manifestConfig := map[string]interface{}{
		"name":       imp.FunctionName,
		"runtime":    detectRuntimeFromTreeForSync(tree.Tree, repo),
		"source":     fmt.Sprintf("github:%s", repo.FullName),
		"branch":     branch,
		"commit":     commitSHA,
		"visibility": imp.Visibility,
		"synced_from": "webhook",
	}

	if imp.ManifestOverrides != nil {
		var overrides map[string]interface{}
		if err := json.Unmarshal(imp.ManifestOverrides, &overrides); err == nil {
			for k, v := range overrides {
				manifestConfig[k] = v
			}
		}
	}

	manifestJSON, _ := json.Marshal(manifestConfig)

	runtimeStr := "node18"
	if rt, ok := manifestConfig["runtime"].(string); ok && rt != "" {
		runtimeStr = rt
	}

	versionID := uuid.New()
	version := &storageregistry.RegistryFunctionVersion{
		ID:          versionID,
		FunctionID:  *imp.FunctionID,
		Version:     nextVersion,
		Manifest:    manifestJSON,
		Runtime:     runtimeStr,
		MemoryMB:    256,
		TimeoutMs:   30000,
		ContentHash: sql.NullString{String: contentHash, Valid: true},
	}

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
			_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 70, "Fetching source code")

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

				content, err := ghClient.GetFileContent(ctx, repo.Owner, repo.Name, filePath, commitSHA)
				if err != nil {
					h.logger.WithError(err).WithField("file", filePath).Warn("Failed to fetch source file during sync")
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
			}
		}
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 80, "Publishing version")

	if err := h.registryRepo.CreateFunctionVersion(version); err != nil {
		h.finishSyncWithError(ctx, syncLog, imp, fmt.Sprintf("Failed to create function version: %v", err))
		return
	}

	filesImported := len(tree.Tree)
	var totalSize int64
	for _, entry := range tree.Tree {
		totalSize += int64(entry.Size)
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 90, "Updating import record")

	if err := h.githubRepo.UpdateImportResult(ctx, imp.ID, *imp.FunctionID, versionID, commitSHA, contentHash, filesImported, totalSize); err != nil {
		h.logger.WithError(err).Warn("Failed to update import result during sync")
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "completed", 100, "")

	durationMs := int(time.Since(startTime).Milliseconds())
	_ = h.githubRepo.UpdateSyncLogStatus(ctx, syncLog.ID, "success", nextVersion, "", durationMs)

	if err := ghClient.CreateCommitStatus(ctx, repo.Owner, repo.Name, commitSHA, &githubsvc.CommitStatusRequest{
		State:       "success",
		Description: fmt.Sprintf("Synced %s to v%s", imp.FunctionName, nextVersion),
		Context:     "functionfly/sync",
	}); err != nil {
		h.logger.WithError(err).Warn("Failed to set commit status on GitHub")
	}

	h.logger.WithFields(map[string]interface{}{
		"import_id":     imp.ID,
		"function_id":   *imp.FunctionID,
		"version_id":    versionID,
		"version":       nextVersion,
		"commit":        commitSHA,
		"duration_ms":   durationMs,
	}).Info("Sync pipeline completed successfully")
}

func detectRuntimeFromTreeForSync(tree []githubsvc.GitHubTreeEntry, repo *storage.GitHubRepo) string {
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

func incrementVersion(current string) string {
	parts := strings.Split(current, ".")
	if len(parts) != 3 {
		return "1.0.0"
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	patch++

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func (h *Handler) finishSyncWithError(ctx context.Context, syncLog *storage.GitHubSyncLog, imp *storage.GitHubImport, errMsg string) {
	h.logger.WithError(fmt.Errorf("%s", errMsg)).WithField("import_id", imp.ID).Error("Sync failed")

	_ = h.githubRepo.UpdateSyncLogStatus(ctx, syncLog.ID, "failed", "", errMsg, 0)
	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "completed", imp.Progress, "")

	if syncLog.TriggerCommitSHA != nil {
		ghClient, err := h.getGitHubClient(ctx, imp.UserID)
		if err == nil {
			repo, repoErr := h.githubRepo.GetRepoByID(ctx, imp.RepoID)
			if repoErr == nil && repo != nil {
				status := &githubsvc.CommitStatusRequest{
					State:       "failure",
					Description: errMsg,
					Context:     "functionfly/sync",
				}
				_ = ghClient.CreateCommitStatus(ctx, repo.Owner, repo.Name, *syncLog.TriggerCommitSHA, status)
			}
		}
	}
}

func verifyHMAC(payload []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)

	return hmac.Equal(sig, expected)
}
