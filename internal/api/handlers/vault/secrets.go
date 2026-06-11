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
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse optional secret type filter
	secretType := vault.SecretType(r.URL.Query().Get("secret_type"))
	if secretType != "" && !secretType.Valid() {
		h.respondError(w, http.StatusBadRequest, "INVALID_SECRET_TYPE", "Invalid secret type")
		return
	}

	// Get total count and page of secrets (DB-level pagination)
	var total int64
	var err error
	if secretType != "" {
		total, err = h.repo.CountSecretsByTenantFiltered(r.Context(), claims.TenantID, secretType)
	} else {
		total, err = h.repo.CountSecretsByTenant(r.Context(), claims.TenantID)
	}
	if err != nil {
		h.logger.WithError(err).Error("Failed to count secrets")
		h.respondError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list secrets")
		return
	}

	var secrets []vault.Secret
	if secretType != "" {
		secrets, err = h.repo.GetSecretsByTenantPaginatedFiltered(r.Context(), claims.TenantID, limit, offset, secretType)
	} else {
		secrets, err = h.repo.GetSecretsByTenantPaginated(r.Context(), claims.TenantID, limit, offset)
	}
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

// HandleRotateSecret handles PATCH /v1/vault/secrets/{id}/rotate
// Rotates a secret's encrypted value by replacing it with new ciphertext.
// Creates a version snapshot before updating.
func (h *Handler) HandleRotateSecret(w http.ResponseWriter, r *http.Request) {
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

	var req RotateSecretRequest
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

	// Decode new encrypted data
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
		authTagBytes, err = base64.StdEncoding.DecodeString(req.EncryptedData.Tag)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "INVALID_TAG", "Tag must be base64 encoded")
			return
		}
	}
	if authTagBytes == nil {
		authTagBytes = []byte{}
	}

	// Reject oversized payloads
	totalPayloadSize := len(ciphertext) + len(saltBytes) + len(ivBytes) + len(authTagBytes)
	if totalPayloadSize > maxEncryptedPayloadBytes {
		h.respondError(w, http.StatusBadRequest, "PAYLOAD_TOO_LARGE",
			fmt.Sprintf("Encrypted payload must not exceed %d bytes", maxEncryptedPayloadBytes))
		return
	}

	// Create version snapshot before rotating
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
		ChangeType:        "rotate",
		ChangeSummary:     "Rotated encrypted value",
		ActorID:           claims.UserID,
		ActorType:         vault.ActorTypeUser,
	}

	if err := h.repo.CreateSecretVersion(r.Context(), version); err != nil {
		h.logger.WithError(err).Error("Failed to create version snapshot before rotation")
	}

	// Update secret with new encrypted data
	secret.EncryptedValue = ciphertext
	secret.EncryptionSalt = saltBytes
	secret.IV = ivBytes
	secret.EncryptionAuthTag = authTagBytes
	secret.KeyVersion = req.EncryptedData.KeyVersion

	if err := h.repo.UpdateSecret(r.Context(), secret); err != nil {
		h.logger.WithError(err).Error("Failed to rotate secret")
		h.respondError(w, http.StatusInternalServerError, "ROTATE_FAILED", "Failed to rotate secret")
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
			"secret_name": secret.Name,
			"operation":   "rotate",
			"reason":      req.Reason,
			"new_version": version.VersionNumber + 1,
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	// Get updated secret
	updatedSecret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil || updatedSecret == nil {
		updatedSecret = secret
	}

	h.respondJSON(w, http.StatusOK, secretToResponse(updatedSecret))
}

// HandleBulkDeleteSecrets handles DELETE /v1/vault/secrets/bulk
// Bulk deletes multiple secrets with dependency checking
func (h *Handler) HandleBulkDeleteSecrets(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if len(req.SecretIDs) == 0 {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "At least one secret ID is required")
		return
	}

	if len(req.SecretIDs) > 100 {
		h.respondError(w, http.StatusBadRequest, "TOO_MANY", "Maximum 100 secrets can be deleted at once")
		return
	}

	if req.DryRun {
		previews, err := h.repo.BulkDeleteSecretsDryRun(r.Context(), req.SecretIDs, claims.TenantID)
		if err != nil {
			h.logger.WithError(err).Error("Failed to get bulk delete previews")
			h.respondError(w, http.StatusInternalServerError, "PREVIEW_FAILED", "Failed to preview bulk delete")
			return
		}

		bulkPreviews := make(map[uuid.UUID]BulkDeletePreview)
		for id, p := range previews {
			depInfos := make([]DependencyInfo, len(p.Dependencies))
			for i, d := range p.Dependencies {
				depInfos[i] = DependencyInfo{
					ID:          d.ID,
					Type:        d.Type,
					Name:        d.Name,
					Criticality: d.Criticality,
				}
			}
			bulkPreviews[id] = BulkDeletePreview{
				SecretID:     p.SecretID,
				SecretName:   p.SecretName,
				Found:        p.Found,
				TokensCount:  p.TokensCount,
				Dependencies: depInfos,
			}
		}

		h.respondJSON(w, http.StatusOK, BulkDeleteResponse{
			Deleted:  0,
			Failed:   0,
			Previews: bulkPreviews,
		})
		return
	}

	deleted, errs := h.repo.BulkDeleteSecrets(r.Context(), req.SecretIDs, claims.TenantID)

	response := BulkDeleteResponse{
		Deleted: deleted,
		Failed:  int64(len(errs)),
	}

	if len(errs) > 0 {
		bulkErrors := make([]BulkDeleteError, len(errs))
		for i, e := range errs {
			var id uuid.UUID
			if se, ok := e.(interface{ SecretID() uuid.UUID }); ok {
				id = se.SecretID()
			}
			bulkErrors[i] = BulkDeleteError{
				SecretID: id,
				Error:    e.Error(),
			}
		}
		response.Errors = bulkErrors
	}

	auditLog := &vault.AuditLog{
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionDelete,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"operation":   "bulk_delete",
			"secret_ids":  len(req.SecretIDs),
			"deleted":     deleted,
			"failed":      len(errs),
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	h.respondJSON(w, http.StatusOK, response)
}

// HandleExportSecrets handles GET /v1/vault/secrets/export
// Exports all secrets for the tenant (metadata only, no encrypted values)
func (h *Handler) HandleExportSecrets(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	exports, err := h.repo.ExportSecrets(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to export secrets")
		h.respondError(w, http.StatusInternalServerError, "EXPORT_FAILED", "Failed to export secrets")
		return
	}

	secretExports := make([]SecretExport, len(exports))
	for i, e := range exports {
		secretExports[i] = SecretExport{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			SecretType:  e.SecretType,
			KeyVersion:  e.KeyVersion,
			Scopes:      jsonMapToScopes(e.Scopes),
			Metadata:    e.Metadata,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		}
	}

	auditLog := &vault.AuditLog{
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionRead,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"operation": "export",
			"count":     len(secretExports),
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	h.respondJSON(w, http.StatusOK, ExportSecretsResponse{
		Secrets:    secretExports,
		Total:      len(secretExports),
		ExportedAt: time.Now(),
	})
}

// HandleGetSecretDependencies handles GET /v1/vault/secrets/{id}/dependencies
// Gets all services/functions that depend on a specific secret
func (h *Handler) HandleGetSecretDependencies(w http.ResponseWriter, r *http.Request) {
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

	deps, err := h.repo.GetSecretDependencies(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret dependencies")
		h.respondError(w, http.StatusInternalServerError, "GET_DEPS_FAILED", "Failed to get dependencies")
		return
	}

	dependencies := make([]SecretDependency, len(deps))
	for i, d := range deps {
		depMeta := make(map[string]interface{})
		for k, v := range d.Metadata {
			depMeta[k] = v
		}
		dependencies[i] = SecretDependency{
			ID:            d.ID,
			DependentID:   d.DependentID,
			DependentType: d.DependentType,
			DependentName: d.DependentName,
			Criticality:   d.Criticality,
			Metadata:      depMeta,
			CreatedAt:     d.CreatedAt,
		}
	}

	h.respondJSON(w, http.StatusOK, SecretDependenciesResponse{
		SecretID:     *secretID,
		Dependencies: dependencies,
		Total:        int64(len(dependencies)),
	})
}

// HandleCreateSecretDependency handles POST /v1/vault/secrets/{id}/dependencies
// Creates a new dependency tracking record
func (h *Handler) HandleCreateSecretDependency(w http.ResponseWriter, r *http.Request) {
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

	var req CreateSecretDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.DependentType != "function" && req.DependentType != "service" &&
		req.DependentType != "integration" && req.DependentType != "workflow" {
		h.respondError(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid dependent type")
		return
	}

	criticality := req.Criticality
	if criticality == "" {
		criticality = "medium"
	}
	if criticality != "low" && criticality != "medium" && criticality != "high" && criticality != "critical" {
		h.respondError(w, http.StatusBadRequest, "INVALID_CRITICALITY", "Invalid criticality level")
		return
	}

	dep := &vault.SecretDependency{
		SecretID:      *secretID,
		TenantID:      claims.TenantID,
		DependentID:   req.DependentID,
		DependentType: req.DependentType,
		DependentName: req.DependentName,
		Criticality:   criticality,
		Metadata:      req.Metadata,
	}

	if err := h.repo.CreateSecretDependency(r.Context(), dep); err != nil {
		h.logger.WithError(err).Error("Failed to create secret dependency")
		h.respondError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create dependency")
		return
	}

	h.respondJSON(w, http.StatusCreated, SecretDependency{
		ID:            dep.ID,
		DependentID:   dep.DependentID,
		DependentType: dep.DependentType,
		DependentName: dep.DependentName,
		Criticality:   dep.Criticality,
		Metadata:      dep.Metadata,
		CreatedAt:     dep.CreatedAt,
	})
}

// HandleDeleteSecretDependency handles DELETE /v1/vault/secrets/{id}/dependencies/{dep_id}
// Deletes a secret dependency record
func (h *Handler) HandleDeleteSecretDependency(w http.ResponseWriter, r *http.Request) {
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

	depID := parseUUID(vars["dep_id"])
	if depID == nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid dependency ID")
		return
	}

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

	if err := h.repo.DeleteSecretDependency(r.Context(), *depID, claims.TenantID); err != nil {
		h.logger.WithError(err).Error("Failed to delete secret dependency")
		h.respondError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete dependency")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
