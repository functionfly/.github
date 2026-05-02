package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
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

	_, err := h.getGitHubClient(ctx, imp.UserID)
	if err != nil {
		h.finishSyncWithError(ctx, syncLog, imp, fmt.Sprintf("Failed to create GitHub client: %v", err))
		return
	}

	_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "syncing", 50, "Building and publishing")

	if imp.FunctionID != nil {
		_ = h.githubRepo.UpdateImportStatus(ctx, imp.ID, "completed", 100, "")

		durationMs := int(time.Since(startTime).Milliseconds())
		_ = h.githubRepo.UpdateSyncLogStatus(ctx, syncLog.ID, "success", "", "", durationMs)
	} else {
		h.finishSyncWithError(ctx, syncLog, imp, "No function associated with import")
	}
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
