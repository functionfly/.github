package vault

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/gorilla/mux"
)

// HandleGenerateToken handles POST /v1/vault/secrets/{id}/tokens
// Generates a secure random token, stores its hash, and returns the plaintext once
func (h *Handler) HandleGenerateToken(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	secretID := parseUUID(vars["id"])
	if secretID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid secret ID"))
		return
	}

	var req GenerateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Verify secret exists and belongs to tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to verify secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}

	// Generate secure random token (32 bytes = 256 bits)
	rawToken, err := generateSecureToken()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate secure token")
		apierror.WriteError(w, apierror.NewInternal("Failed to generate token"))
		return
	}

	// Hash the token for storage (only hash is kept)
	tokenHash := hashToken(rawToken)

	// Calculate expiration
	expiresInHours := req.ExpiresInHours
	if expiresInHours <= 0 {
		expiresInHours = 24 // Default 24 hours
	}
	if expiresInHours > 8760 { // Max 1 year
		expiresInHours = 8760
	}
	expiresAt := time.Now().Add(time.Duration(expiresInHours) * time.Hour)

	// Create access token
	token := &vault.AccessToken{
		SecretID:  *secretID,
		TenantID:  claims.TenantID,
		TokenHash: tokenHash,
		Name:      req.Name,
		Scopes:    scopesToJSONMap(req.Scopes),
		ExpiresAt: expiresAt,
		CreatedBy: claims.UserID.String(),
	}

	if err := h.repo.CreateAccessToken(r.Context(), token); err != nil {
		h.logger.WithError(err).Error("Failed to create access token")
		apierror.WriteError(w, apierror.NewInternal("Failed to create token"))
		return
	}

	// Log audit event
	auditLog := &vault.AuditLog{
		SecretID:  secretID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionCreate,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"token_id":      token.ID.String(),
			"secret_name":   secret.Name,
			"token_name":    req.Name,
			"expires_hours": expiresInHours,
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	// Return token (plaintext shown only once!)
	h.respondJSON(w, http.StatusCreated, GenerateTokenResponse{
		TokenID:   token.ID,
		Token:     rawToken,
		SecretID:  *secretID,
		Name:      req.Name,
		ExpiresAt: expiresAt,
		Scopes:    req.Scopes,
		CreatedAt: token.CreatedAt,
	})
}

// HandleRevokeToken handles DELETE /v1/vault/tokens/{id}
// Revokes an access token by ID
func (h *Handler) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	tokenID := parseUUID(vars["id"])
	if tokenID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid token ID"))
		return
	}

	// Get token to verify ownership
	token, err := h.repo.GetAccessTokenByID(r.Context(), *tokenID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get token")
		apierror.WriteError(w, apierror.NewInternal("Failed to get token"))
		return
	}
	if token == nil {
		apierror.WriteError(w, apierror.NewNotFound("Token not found"))
		return
	}

	// Parse optional reason from body
	var req struct {
		Reason string `json:"reason,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "manual_revocation"
	}

	// Revoke the token
	if err := h.repo.RevokeAccessToken(r.Context(), *tokenID, req.Reason); err != nil {
		h.logger.WithError(err).Error("Failed to revoke token")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke token"))
		return
	}

	// Log audit event
	auditLog := &vault.AuditLog{
		SecretID:  &token.SecretID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionRevoke,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"token_id":   tokenID.String(),
			"token_name": token.Name,
			"reason":     req.Reason,
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Token revoked successfully",
		"token_id": tokenID.String(),
		"revoked":  true,
	})
}

// HandleListTokens handles GET /v1/vault/secrets/{id}/tokens
// Lists all access tokens for a secret (without the actual token values)
func (h *Handler) HandleListTokens(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	secretID := parseUUID(vars["id"])
	if secretID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid secret ID"))
		return
	}

	// Verify secret exists and belongs to tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to verify secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}

	// Get tokens
	tokens, err := h.repo.ListAccessTokensBySecret(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tokens")
		apierror.WriteError(w, apierror.NewInternal("Failed to list tokens"))
		return
	}

	// Convert to response (without token values)
	responses := make([]TokenResponse, len(tokens))
	for i, token := range tokens {
		responses[i] = tokenToResponse(&token)
	}

	h.respondJSON(w, http.StatusOK, ListTokensResponse{
		Tokens: responses,
		Total:  int64(len(tokens)),
	})
}

// ValidateTokenForRuntime validates a token for function execution
// This is a middleware function that can be used to protect runtime endpoints
func (h *Handler) ValidateTokenForRuntime(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			apierror.WriteError(w, apierror.NewUnauthorized("Authorization header required"))
			return
		}

		// Expected format: "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			apierror.WriteError(w, apierror.NewUnauthorized("Invalid authorization format"))
			return
		}
		rawToken := authHeader[7:]

		// Hash the provided token
		tokenHash := hashToken(rawToken)

		// Look up token by hash
		token, err := h.repo.GetAccessTokenByHash(r.Context(), tokenHash)
		if err != nil {
			h.logger.WithError(err).Error("Failed to validate token")
			apierror.WriteError(w, apierror.NewInternal("Token validation failed"))
			return
		}
		if token == nil {
			apierror.WriteError(w, apierror.NewUnauthorized("Invalid or revoked token"))
			return
		}
		// Constant-time comparison to prevent timing leakage
		if !constantTimeCompare(token.TokenHash, tokenHash) {
			apierror.WriteError(w, apierror.NewUnauthorized("Invalid or revoked token"))
			return
		}

		// Check if token is valid (not expired, not revoked)
		if !token.IsValid() {
			if token.IsRevoked {
				apierror.WriteError(w, apierror.NewUnauthorized("Token has been revoked"))
			} else {
				apierror.WriteError(w, apierror.NewUnauthorized("Token has expired"))
			}
			return
		}

		// Record token use
		if err := h.repo.RecordTokenUse(r.Context(), token.ID); err != nil {
			h.logger.WithError(err).Warn("Failed to record token use")
		}

		// Log audit event
		auditLog := &vault.AuditLog{
			SecretID:  &token.SecretID,
			TenantID:  token.TenantID,
			Action:    vault.AuditActionUse,
			ActorID:   token.ID.String(),
			ActorType: vault.ActorTypeToken,
			RequestID: r.Header.Get("X-Request-ID"),
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
			Metadata: vault.JSONMap{
				"token_id": token.ID.String(),
			},
			Success: true,
		}
		if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
			h.logger.WithError(err).Warn("Failed to create audit log")
		}

		// Token is valid, continue to next handler
		next.ServeHTTP(w, r)
	}
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
