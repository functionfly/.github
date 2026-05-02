package github

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/storage"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
)

// HandleUpdateSync updates auto-sync settings for an import.
func (h *Handler) HandleUpdateSync(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	importID, err := h.parseUUID(r, "importId")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Invalid import ID")
		return
	}

	var req UpdateSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
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

	syncBranches := req.SyncBranches
	if syncBranches == nil {
		syncBranches = imp.SyncBranches
	}

	if err := h.githubRepo.UpdateImportSyncSettings(r.Context(), importID, req.AutoSyncEnabled, syncBranches); err != nil {
		h.logger.WithError(err).Error("Failed to update import sync settings")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to update sync settings")
		return
	}

	if req.AutoSyncEnabled && !imp.AutoSyncEnabled {
		go h.ensureWebhook(context.Background(), imp)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"auto_sync_enabled": req.AutoSyncEnabled,
		"sync_branches":     syncBranches,
	})
}

// HandleGetSyncLogs returns paginated sync logs for an import.
func (h *Handler) HandleGetSyncLogs(w http.ResponseWriter, r *http.Request) {
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

	params := storage.ListSyncLogsParams{
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

	logs, total, err := h.githubRepo.ListSyncLogsByImport(r.Context(), importID, params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list sync logs")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to list sync logs")
		return
	}

	results := make([]*SyncLogResponse, len(logs))
	for i, log := range logs {
		results[i] = &SyncLogResponse{
			ID:               log.ID,
			TriggerType:      log.TriggerType,
			TriggerBranch:    log.TriggerBranch,
			TriggerCommitSHA: log.TriggerCommitSHA,
			TriggerPRNumber:  log.TriggerPRNumber,
			Status:           log.Status,
			VersionPublished: log.VersionPublished,
			DurationMs:       log.DurationMs,
			ErrorMessage:     log.ErrorMessage,
			CreatedAt:        log.CreatedAt,
			CompletedAt:      log.CompletedAt,
		}
	}

	h.respondJSON(w, http.StatusOK, ListSyncLogsResponse{
		Logs:    results,
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
	})
}

// ensureWebhook creates a GitHub webhook for auto-sync if one doesn't exist.
func (h *Handler) ensureWebhook(ctx context.Context, imp *storage.GitHubImport) {
	log := h.logger.WithField("import_id", imp.ID)

	conn, err := h.githubRepo.GetConnectionByID(ctx, imp.ConnectionID)
	if err != nil || conn == nil {
		log.Warn("Cannot ensure webhook: connection not found")
		return
	}

	repo, err := h.githubRepo.GetRepoByID(ctx, imp.RepoID)
	if err != nil || repo == nil {
		log.Warn("Cannot ensure webhook: repo not found")
		return
	}

	existing, err := h.githubRepo.GetWebhookByRepoID(ctx, imp.RepoID)
	if err == nil && existing != nil {
		return
	}

	vault, err := githubsvc.NewTokenVault(h.vaultKey)
	if err != nil {
		log.WithError(err).Warn("Cannot create vault for webhook")
		return
	}

	token, err := vault.Decrypt(conn.EncryptedToken, conn.TokenIV, conn.TokenTag)
	if err != nil {
		log.WithError(err).Warn("Cannot decrypt token for webhook")
		return
	}

	ghClient := githubsvc.NewClient(token, githubsvc.WithLogger(h.logger))

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		log.WithError(err).Warn("Cannot generate webhook secret")
		return
	}
	webhookSecret := fmt.Sprintf("%x", secretBytes)

	webhookURL := h.baseURL + "/api/v1/github/webhook"
	whReq := &githubsvc.GitHubWebhookRequest{
		Name:   "web",
		Active: true,
		Events: []string{"push", "pull_request"},
	}
	whReq.Config.URL = webhookURL
	whReq.Config.ContentType = "json"
	whReq.Config.Secret = webhookSecret

	whResp, err := ghClient.CreateWebhook(ctx, repo.Owner, repo.Name, whReq)
	if err != nil {
		log.WithError(err).Warn("Failed to create webhook on GitHub")
		return
	}

	eventsJSON, _ := json.Marshal([]string{"push", "pull_request"})
	wh := &storage.GitHubWebhook{
		ConnectionID:    conn.ID,
		RepoID:          imp.RepoID,
		GithubWebhookID: &whResp.ID,
		WebhookSecret:   webhookSecret,
		Events:          eventsJSON,
		IsActive:        true,
	}

	if _, err := h.githubRepo.CreateWebhook(ctx, wh); err != nil {
		log.WithError(err).Warn("Failed to store webhook record")
	}
}
