package vault

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Max secret name length (matches DB column)
const maxSecretNameLen = 255

// Max total size for encrypted payload (ciphertext + salt + iv + auth tag) to prevent DoS
const maxEncryptedPayloadBytes = 64 * 1024 // 64 KB

// HandleCreateSecret handles POST /v1/vault/secrets
// Creates a new encrypted secret in the vault
func (h *Handler) HandleCreateSecret(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate secret type
	if !req.SecretType.Valid() {
		h.respondError(w, http.StatusBadRequest, "INVALID_SECRET_TYPE", "Invalid secret type")
		return
	}

	// Validate name: trim, non-empty, max length
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_NAME", "Secret name is required")
		return
	}
	if len(req.Name) > maxSecretNameLen {
		h.respondError(w, http.StatusBadRequest, "INVALID_NAME", fmt.Sprintf("Secret name must be at most %d characters", maxSecretNameLen))
		return
	}

	// Get tenant plan and check secrets limit
	plan := middleware.GetTenantPlan(r)
	maxSecrets := plans.GetMaxSecrets(plan)

	// Get current secret count for tenant
	count, err := h.repo.CountSecretsByTenant(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count secrets")
		h.respondError(w, http.StatusInternalServerError, "COUNT_FAILED", "Failed to verify secret limit")
		return
	}

	if count >= int64(maxSecrets) {
		h.respondError(w, http.StatusForbidden, "SECRET_LIMIT_EXCEEDED",
			fmt.Sprintf("Maximum number of secrets (%d) exceeded for your plan", maxSecrets))
		return
	}

	// Decode base64 encrypted data
	ciphertext, err := base64.StdEncoding.DecodeString(req.EncryptedData.Ciphertext)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_CIPHERTEXT", "Ciphertext must be base64 encoded")
		return
	}
	saltBytes, err := base64.StdEncoding.DecodeString(req.EncryptedData.Salt)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_SALT", "Salt must be base64 encoded")
		return
	}
	ivBytes, err := base64.StdEncoding.DecodeString(req.EncryptedData.IV)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_IV", "IV must be base64 encoded")
		return
	}
	var authTagBytes []byte
	if req.EncryptedData.Tag != "" {
		var errTag error
		authTagBytes, errTag = base64.StdEncoding.DecodeString(req.EncryptedData.Tag)
		if errTag != nil {
			h.respondError(w, http.StatusBadRequest, "INVALID_TAG", "Tag must be base64 encoded")
			return
		}
	}
	if authTagBytes == nil {
		authTagBytes = []byte{}
	}

	// Reject oversized payloads to prevent DoS and storage blow-up
	totalPayloadSize := len(ciphertext) + len(saltBytes) + len(ivBytes) + len(authTagBytes)
	if totalPayloadSize > maxEncryptedPayloadBytes {
		h.respondError(w, http.StatusBadRequest, "PAYLOAD_TOO_LARGE",
			fmt.Sprintf("Encrypted payload must not exceed %d bytes", maxEncryptedPayloadBytes))
		return
	}

	// Create secret model
	secret := &vault.Secret{
		TenantID:          claims.TenantID,
		UserID:            claims.UserID,
		Name:              req.Name,
		Description:       req.Description,
		SecretType:        req.SecretType,
		EncryptedValue:    ciphertext,
		EncryptionSalt:    saltBytes,
		IV:                ivBytes,
		EncryptionAuthTag: authTagBytes,
		KeyVersion:        req.EncryptedData.KeyVersion,
		Scopes:            scopesToJSONMap(req.Scopes),
		Metadata:          req.Metadata,
	}

	// Create secret in database
	if err := h.repo.CreateSecret(r.Context(), secret); err != nil {
		h.logger.WithError(err).Error("Failed to create secret")
		h.respondError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create secret")
		return
	}

	// Log audit event
	auditLog := &vault.AuditLog{
		SecretID:  &secret.ID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionCreate,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"secret_name": req.Name,
			"secret_type": req.SecretType.String(),
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	// Return created secret
	h.respondJSON(w, http.StatusCreated, secretToResponse(secret))
}

// HandleListSecrets handles GET /v1/vault/secrets
// Lists all secrets for the tenant (metadata only, no encrypted values)
func (h *Handler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Get total count and page of secrets (DB-level pagination)
	total, err := h.repo.CountSecretsByTenant(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count secrets")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list secrets")
		return
	}
	secrets, err := h.repo.GetSecretsByTenantPaginated(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list secrets")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list secrets")
		return
	}

	// Convert to metadata responses (no encrypted data)
	responses := make([]SecretMetadataResponse, len(secrets))
	for i, secret := range secrets {
		responses[i] = secretToMetadataResponse(&secret)
	}

	h.respondJSON(w, http.StatusOK, ListSecretsResponse{
		Secrets: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// HandleGetSecret handles GET /v1/vault/secrets/{id}
// Gets a single secret with encrypted data by ID
func (h *Handler) HandleGetSecret(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	secretID := parseUUID(vars["id"])
	if secretID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid secret ID")
		return
	}

	// Get secret with tenant isolation
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get secret")
		return
	}
	if secret == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
		return
	}

	// Record access
	if err := h.repo.RecordSecretAccess(r.Context(), *secretID, claims.TenantID); err != nil {
		h.logger.WithError(err).Warn("Failed to record secret access")
	}

	// Log audit event
	auditLog := &vault.AuditLog{
		SecretID:  secretID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionRead,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"secret_name": secret.Name,
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	h.respondJSON(w, http.StatusOK, secretToResponse(secret))
}

// HandleUpdateSecret handles PATCH /v1/vault/secrets/{id}
// Partially updates a secret (name, description, scopes)
func (h *Handler) HandleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	secretID := parseUUID(vars["id"])
	if secretID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid secret ID")
		return
	}

	var req UpdateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get existing secret
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get secret")
		return
	}
	if secret == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
		return
	}

	// Apply partial updates with validation
	updated := false
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			h.respondError(w, http.StatusBadRequest, "INVALID_NAME", "Secret name cannot be empty")
			return
		}
		if len(trimmed) > maxSecretNameLen {
			h.respondError(w, http.StatusBadRequest, "INVALID_NAME", fmt.Sprintf("Secret name must be at most %d characters", maxSecretNameLen))
			return
		}
		secret.Name = trimmed
		updated = true
	}
	if req.Description != nil {
		secret.Description = *req.Description
		updated = true
	}
	if req.Scopes != nil {
		secret.Scopes = scopesToJSONMap(*req.Scopes)
		updated = true
	}

	if !updated {
		h.respondError(w, http.StatusBadRequest, "NO_UPDATES", "No valid fields to update")
		return
	}

	// Build change summary
	var changeSummary string
	var updatedFields []string
	if req.Name != nil {
		updatedFields = append(updatedFields, "name")
	}
	if req.Description != nil {
		updatedFields = append(updatedFields, "description")
	}
	if req.Scopes != nil {
		updatedFields = append(updatedFields, "scopes")
	}
	if len(updatedFields) > 0 {
		changeSummary = "Updated: " + strings.Join(updatedFields, ", ")
	}

	// Create version snapshot before updating (for history tracking)
	version := &vault.SecretVersion{
		SecretID:          secret.ID,
		TenantID:          secret.TenantID,
		Name:              secret.Name,
		Description:       secret.Description,
		SecretType:        secret.SecretType,
		EncryptedValue:    secret.EncryptedValue,
		EncryptionSalt:    secret.EncryptionSalt,
		IV:                secret.IV,
		EncryptionAuthTag: secret.EncryptionAuthTag,
		KeyVersion:        secret.KeyVersion,
		Scopes:            secret.Scopes,
		Metadata:          secret.Metadata,
		ChangeType:        "update",
		ChangeSummary:     changeSummary,
		ActorID:           claims.UserID,
		ActorType:         vault.ActorTypeUser,
	}

	if err := h.repo.CreateSecretVersion(r.Context(), version); err != nil {
		h.logger.WithError(err).Error("Failed to create version snapshot")
		// Continue with update even if versioning fails - don't block the operation
	}

	// Update secret
	if err := h.repo.UpdateSecret(r.Context(), secret); err != nil {
		h.logger.WithError(err).Error("Failed to update secret")
		h.respondError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update secret")
		return
	}

	// Log audit event
	auditLog := &vault.AuditLog{
		SecretID:  secretID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionUpdate,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"secret_name":    secret.Name,
			"updated_fields": updatedFields,
			"new_version":    version.VersionNumber,
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	// Log version creation audit event
	if version.VersionNumber > 0 {
		versionAuditLog := &vault.AuditLog{
			SecretID:  secretID,
			TenantID:  claims.TenantID,
			Action:    vault.AuditActionVersion,
			ActorID:   claims.UserID.String(),
			ActorType: vault.ActorTypeUser,
			RequestID: r.Header.Get("X-Request-ID"),
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
			Metadata: vault.JSONMap{
				"secret_name":    secret.Name,
				"version_number": version.VersionNumber,
				"change_type":    "update",
			},
			Success: true,
		}
		if err := h.repo.CreateAuditLog(r.Context(), versionAuditLog); err != nil {
			h.logger.WithError(err).Warn("Failed to create version audit log")
		}
	}

	h.respondJSON(w, http.StatusOK, secretToResponse(secret))
}

// HandleDeleteSecret handles DELETE /v1/vault/secrets/{id}
// Soft-deletes a secret and revokes associated tokens
func (h *Handler) HandleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	secretID := parseUUID(vars["id"])
	if secretID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid secret ID")
		return
	}

	// Get existing secret to verify it exists and get name for audit
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
			return
		}
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondError(w, http.StatusInternalServerError, "GET_FAILED", "Failed to get secret")
		return
	}
	if secret == nil {
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
		return
	}

	secretName := secret.Name

	// Revoke associated tokens
	tokens, err := h.repo.ListAccessTokensBySecret(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to list tokens for revocation")
	} else {
		for _, token := range tokens {
			if !token.IsRevoked && token.RevokedAt == nil {
				if err := h.repo.RevokeAccessToken(r.Context(), token.ID, "secret_deleted"); err != nil {
					h.logger.WithError(err).WithField("token_id", token.ID).Warn("Failed to revoke token")
				}
			}
		}
	}

	// Soft-delete the secret
	if err := h.repo.DeleteSecret(r.Context(), *secretID, claims.TenantID); err != nil {
		h.logger.WithError(err).Error("Failed to delete secret")
		h.respondError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete secret")
		return
	}

	// Log audit event
	auditLog := &vault.AuditLog{
		SecretID:  secretID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionDelete,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"secret_name":    secretName,
			"tokens_revoked": len(tokens),
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	w.WriteHeader(http.StatusNoContent)
}

// logAuditError logs an audit event for an error that occurred
func (h *Handler) logAuditError(ctx context.Context, tenantID uuid.UUID, action vault.AuditAction, actorID string, secretID *uuid.UUID, errMsg string, metadata vault.JSONMap) {
	auditLog := &vault.AuditLog{
		SecretID:     secretID,
		TenantID:     tenantID,
		Action:       action,
		ActorID:      actorID,
		ActorType:    vault.ActorTypeUser,
		Metadata:     metadata,
		Success:      false,
		ErrorMessage: errMsg,
		CreatedAt:    time.Now(),
	}
	if logErr := h.repo.CreateAuditLog(ctx, auditLog); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to create audit log for error")
	}
}
