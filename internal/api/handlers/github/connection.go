package github

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	githubsvc "github.com/functionfly/functionfly/internal/services/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	ghOAuth "golang.org/x/oauth2/github"
)

var (
	oauthStates   = &stateStore{entries: make(map[string]stateEntry)}
	oauthStatesMu sync.Mutex
)

type stateEntry struct {
	userID   uuid.UUID
	tenantID uuid.UUID
	expires  time.Time
}

type stateStore struct {
	entries map[string]stateEntry
}

func (s *stateStore) Set(state string, entry stateEntry) {
	s.entries[state] = entry
}

func (s *stateStore) Consume(state string) (stateEntry, bool) {
	entry, ok := s.entries[state]
	if ok {
		delete(s.entries, state)
	}
	return entry, ok
}

func (s *stateStore) Cleanup() {
	now := time.Now()
	for k, v := range s.entries {
		if now.After(v.expires) {
			delete(s.entries, k)
		}
	}
}

func (h *Handler) getOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = h.baseURL + "/api/v1/github/callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     ghOAuth.Endpoint,
		Scopes:       []string{"repo", "read:user", "user:email"},
		RedirectURL:  redirectURL,
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HandleConnect generates a GitHub OAuth URL and returns it.
func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	state, err := generateState()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate OAuth state")
		h.respondError(w, http.StatusInternalServerError, "state_error", "Failed to generate state")
		return
	}

	oauthStatesMu.Lock()
	oauthStates.Set(state, stateEntry{
		userID:   claims.UserID,
		tenantID: claims.TenantID,
		expires:  time.Now().Add(10 * time.Minute),
	})
	oauthStatesMu.Unlock()

	conf := h.getOAuthConfig()
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)

	h.respondJSON(w, http.StatusOK, ConnectResponse{URL: url})
}

// HandleCallback exchanges the OAuth code for a token, encrypts it, and stores the connection.
func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		h.respondError(w, http.StatusBadRequest, "missing_params", "Missing code or state parameter")
		return
	}

	oauthStatesMu.Lock()
	entry, ok := oauthStates.Consume(state)
	oauthStatesMu.Unlock()

	if !ok {
		h.respondError(w, http.StatusBadRequest, "invalid_state", "Invalid or expired OAuth state")
		return
	}

	if time.Now().After(entry.expires) {
		h.respondError(w, http.StatusBadRequest, "state_expired", "OAuth state has expired")
		return
	}

	conf := h.getOAuthConfig()
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		h.logger.WithError(err).Error("Failed to exchange OAuth code")
		h.respondError(w, http.StatusBadRequest, "exchange_failed", "Failed to exchange authorization code")
		return
	}

	ghClient := githubsvc.NewClient(token.AccessToken, githubsvc.WithLogger(h.logger))
	ghUser, err := ghClient.GetAuthenticatedUser(context.Background())
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch GitHub user info")
		h.respondError(w, http.StatusInternalServerError, "user_info_failed", "Failed to fetch GitHub user information")
		return
	}

	vault, err := githubsvc.NewTokenVault(h.vaultKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create token vault")
		h.respondError(w, http.StatusInternalServerError, "vault_error", "Internal server error")
		return
	}

	ciphertext, iv, tag, err := vault.Encrypt(token.AccessToken)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt token")
		h.respondError(w, http.StatusInternalServerError, "encrypt_error", "Internal server error")
		return
	}

	var encryptedRefresh, refreshIV, refreshTag *string
	if token.RefreshToken != "" {
		rc, ri, rt, err := vault.Encrypt(token.RefreshToken)
		if err != nil {
			h.logger.WithError(err).Error("Failed to encrypt refresh token")
			h.respondError(w, http.StatusInternalServerError, "encrypt_error", "Internal server error")
			return
		}
		encryptedRefresh = &rc
		refreshIV = &ri
		refreshTag = &rt
	}

	var scope *string
	if s, ok := token.Extra("scope").(string); ok && s != "" {
		scope = &s
	}

	var expiresAt *time.Time
	if !token.Expiry.IsZero() {
		expiresAt = &token.Expiry
	}

	existing, err := h.githubRepo.GetConnectionByUserAndGitHubID(r.Context(), entry.userID, int64(ghUser.ID))
	if err != nil {
		h.logger.WithError(err).Error("Failed to check existing connection")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Internal server error")
		return
	}

	avatarURL := ghUser.AvatarURL
	profileURL := ghUser.HTMLURL

	if existing != nil {
		updates := map[string]interface{}{
			"encrypted_token":   ciphertext,
			"token_iv":          iv,
			"token_tag":         tag,
			"encrypted_refresh": encryptedRefresh,
			"refresh_iv":        refreshIV,
			"refresh_tag":       refreshTag,
			"token_scope":       scope,
			"token_expires_at":  expiresAt,
			"github_username":   ghUser.Login,
			"github_avatar_url": avatarURL,
			"github_profile_url": profileURL,
			"status":            "active",
			"last_synced_at":    time.Now().UTC(),
		}
		if err := h.githubRepo.UpdateConnection(r.Context(), existing.ID, updates); err != nil {
			h.logger.WithError(err).Error("Failed to update connection")
			h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to update connection")
			return
		}
	} else {
		conn := &storage.GitHubConnection{
			UserID:           entry.userID,
			TenantID:         entry.tenantID,
			GithubUserID:     int64(ghUser.ID),
			GithubUsername:    ghUser.Login,
			GithubAvatarURL:  &avatarURL,
			GithubProfileURL: &profileURL,
			EncryptedToken:   ciphertext,
			TokenIV:          iv,
			TokenTag:         tag,
			EncryptedRefresh: encryptedRefresh,
			RefreshIV:        refreshIV,
			RefreshTag:       refreshTag,
			TokenScope:       scope,
			TokenExpiresAt:   expiresAt,
			Status:           "active",
		}
		if _, err := h.githubRepo.CreateConnection(r.Context(), conn); err != nil {
			h.logger.WithError(err).Error("Failed to create connection")
			h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to create connection")
			return
		}
	}

	redirectURL := h.baseURL + "/dashboard/github?connected=true"
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// HandleGetConnection returns the current GitHub connection (no secrets).
func (h *Handler) HandleGetConnection(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get connection")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to get connection")
		return
	}
	if conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	h.respondJSON(w, http.StatusOK, ConnectionResponse{
		ID:               conn.ID,
		GithubUsername:   conn.GithubUsername,
		GithubAvatarURL:  conn.GithubAvatarURL,
		GithubProfileURL: conn.GithubProfileURL,
		TokenScope:       conn.TokenScope,
		TokenExpiresAt:   conn.TokenExpiresAt,
		Status:           conn.Status,
		ConnectedAt:      conn.CreatedAt,
		LastSyncedAt:     conn.LastSyncedAt,
	})
}

// HandleDisconnect deletes the GitHub connection after revoking webhooks.
func (h *Handler) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get connection")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to get connection")
		return
	}
	if conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	ghClient, ghErr := h.getGitHubClient(r.Context(), claims.UserID)
	if ghErr == nil {
		repos, _, repoErr := h.githubRepo.ListReposByConnection(r.Context(), conn.ID, storage.ListReposParams{PerPage: 100})
		if repoErr == nil {
			for _, repo := range repos {
				webhooks, whErr := h.githubRepo.GetActiveWebhooksByRepoID(r.Context(), repo.ID)
				if whErr != nil {
					continue
				}
				for _, wh := range webhooks {
					if wh.GithubWebhookID != nil {
						if delErr := ghClient.DeleteWebhook(r.Context(), repo.Owner, repo.Name, *wh.GithubWebhookID); delErr != nil {
							h.logger.WithError(delErr).WithField("webhook_id", *wh.GithubWebhookID).Warn("Failed to revoke webhook on GitHub")
						}
					}
					_ = h.githubRepo.DeleteWebhook(r.Context(), wh.ID)
				}
			}
		}
	}

	if err := h.githubRepo.DeleteConnection(r.Context(), conn.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete connection")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to delete connection")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

// HandleRefreshToken refreshes the OAuth token for the current connection.
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	conn, err := h.githubRepo.GetConnectionByUserID(r.Context(), claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get connection")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to get connection")
		return
	}
	if conn == nil {
		h.respondError(w, http.StatusNotFound, "not_found", "No GitHub connection found")
		return
	}

	if conn.EncryptedRefresh == nil || *conn.EncryptedRefresh == "" {
		h.respondError(w, http.StatusBadRequest, "no_refresh_token", "No refresh token available")
		return
	}

	vault, err := githubsvc.NewTokenVault(h.vaultKey)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create vault")
		h.respondError(w, http.StatusInternalServerError, "vault_error", "Internal server error")
		return
	}

	refreshToken, err := vault.Decrypt(*conn.EncryptedRefresh, *conn.RefreshIV, *conn.RefreshTag)
	if err != nil {
		h.logger.WithError(err).Error("Failed to decrypt refresh token")
		h.respondError(w, http.StatusInternalServerError, "decrypt_error", "Internal server error")
		return
	}

	conf := h.getOAuthConfig()
	ts := conf.TokenSource(r.Context(), &oauth2.Token{
		RefreshToken: refreshToken,
	})
	newToken, err := ts.Token()
	if err != nil {
		h.logger.WithError(err).Error("Failed to refresh token")
		h.respondError(w, http.StatusBadGateway, "refresh_failed", "Failed to refresh GitHub token")
		return
	}

	ct, iv, tag, err := vault.Encrypt(newToken.AccessToken)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt new token")
		h.respondError(w, http.StatusInternalServerError, "encrypt_error", "Internal server error")
		return
	}

	updates := map[string]interface{}{
		"encrypted_token": ct,
		"token_iv":        iv,
		"token_tag":       tag,
		"status":          "active",
	}

	if newToken.RefreshToken != "" {
		rc, ri, rt, encErr := vault.Encrypt(newToken.RefreshToken)
		if encErr == nil {
			updates["encrypted_refresh"] = rc
			updates["refresh_iv"] = ri
			updates["refresh_tag"] = rt
		}
	}

	if !newToken.Expiry.IsZero() {
		updates["token_expires_at"] = newToken.Expiry
	}

	if err := h.githubRepo.UpdateConnection(r.Context(), conn.ID, updates); err != nil {
		h.logger.WithError(err).Error("Failed to update token")
		h.respondError(w, http.StatusInternalServerError, "db_error", "Failed to update token")
		return
	}

	var expiresAt *time.Time
	if !newToken.Expiry.IsZero() {
		expiresAt = &newToken.Expiry
	}

	h.respondJSON(w, http.StatusOK, RefreshTokenResponse{ExpiresAt: expiresAt})
}

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			oauthStatesMu.Lock()
			oauthStates.Cleanup()
			oauthStatesMu.Unlock()
		}
	}()
}
