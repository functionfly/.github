package vault

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/gorilla/mux"
)

// HandleListSecretVersions handles GET /v1/vault/secrets/{id}/versions
// Lists all versions for a secret (metadata only, no encrypted data)
func (h *Handler) HandleListSecretVersions(w http.ResponseWriter, r *http.Request) {
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

	// Parse pagination params
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	// Verify the secret exists and belongs to this tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}

	// Get total count and versions
	total, err := h.repo.CountSecretVersions(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count versions")
		apierror.WriteError(w, apierror.NewInternal("Failed to count versions"))
		return
	}

	versions, err := h.repo.GetSecretVersions(r.Context(), *secretID, claims.TenantID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list versions")
		apierror.WriteError(w, apierror.NewInternal("Failed to list versions"))
		return
	}

	// Convert to response (metadata only)
	responses := make([]SecretVersionMetadataResponse, len(versions))
	for i, v := range versions {
		responses[i] = secretVersionToMetadataResponse(&v)
	}

	h.respondJSON(w, http.StatusOK, ListSecretVersionsResponse{
		Versions: responses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// HandleGetSecretVersion handles GET /v1/vault/secrets/{id}/versions/{version}
// Gets a specific version of a secret
func (h *Handler) HandleGetSecretVersion(w http.ResponseWriter, r *http.Request) {
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

	versionNumber, err := strconv.Atoi(vars["version"])
	if err != nil || versionNumber < 1 {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid version number"))
		return
	}

	// Verify the secret exists and belongs to this tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}

	// Get the specific version
	version, err := h.repo.GetSecretVersionByNumber(r.Context(), *secretID, claims.TenantID, versionNumber)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get version")
		apierror.WriteError(w, apierror.NewInternal("Failed to get version"))
		return
	}
	if version == nil {
		apierror.WriteError(w, apierror.NewNotFound("Version not found"))
		return
	}

	// Include encrypted data only if explicitly requested with ?include_encrypted=true
	includeEncrypted := r.URL.Query().Get("include_encrypted") == "true"

	h.respondJSON(w, http.StatusOK, secretVersionToResponse(version, includeEncrypted))
}

// HandleDiffSecretVersions handles GET /v1/vault/secrets/{id}/versions/diff
// Compares two versions of a secret and returns the differences
// Query params: from_version (int), to_version (int, optional, defaults to current)
func (h *Handler) HandleDiffSecretVersions(w http.ResponseWriter, r *http.Request) {
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

	// Parse version params
	fromVersionStr := r.URL.Query().Get("from_version")
	if fromVersionStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("from_version query parameter is required"))
		return
	}

	fromVersion, err := strconv.Atoi(fromVersionStr)
	if err != nil || fromVersion < 1 {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid from_version"))
		return
	}

	toVersionStr := r.URL.Query().Get("to_version")

	// Verify the secret exists and belongs to this tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}

	// Determine the to_version (default to current version if not specified)
	toVersion := fromVersion + 1
	if toVersionStr != "" {
		toVersion, err = strconv.Atoi(toVersionStr)
		if err != nil || toVersion < 1 {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid to_version"))
			return
		}
	} else if secret.CurrentVersion != nil {
		toVersion = *secret.CurrentVersion
	}

	// Ensure from < to
	if fromVersion >= toVersion {
		apierror.WriteError(w, apierror.NewBadRequest("from_version must be less than to_version"))
		return
	}

	// Get both versions
	fromVer, err := h.repo.GetSecretVersionByNumber(r.Context(), *secretID, claims.TenantID, fromVersion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get from version")
		apierror.WriteError(w, apierror.NewInternal("Failed to get version"))
		return
	}
	if fromVer == nil {
		apierror.WriteError(w, apierror.NewNotFound("From version not found"))
		return
	}

	toVer, err := h.repo.GetSecretVersionByNumber(r.Context(), *secretID, claims.TenantID, toVersion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get to version")
		apierror.WriteError(w, apierror.NewInternal("Failed to get version"))
		return
	}
	if toVer == nil {
		apierror.WriteError(w, apierror.NewNotFound("To version not found"))
		return
	}

	// Compare the versions
	diff := SecretVersionDiffResponse{
		FromVersion:   fromVersion,
		ToVersion:     toVersion,
		NameChanged:   fromVer.Name != toVer.Name,
		NameFrom:      fromVer.Name,
		NameTo:        toVer.Name,
		DescChanged:   fromVer.Description != toVer.Description,
		DescFrom:      fromVer.Description,
		DescTo:        toVer.Description,
		ScopesChanged: !scopesEqual(jsonMapToScopes(fromVer.Scopes), jsonMapToScopes(toVer.Scopes)),
		ScopesFrom:    jsonMapToScopes(fromVer.Scopes),
		ScopesTo:      jsonMapToScopes(toVer.Scopes),
		// Note: We can't compare encrypted values directly, but we can tell if they changed
		// by checking if the ciphertext bytes differ
		EncryptedChanged: !bytesEqual(fromVer.EncryptedValue, toVer.EncryptedValue),
		ActorID:          toVer.ActorID,
		ActorType:        toVer.ActorType,
		CreatedAt:        toVer.CreatedAt,
	}

	diff.HasChanges = diff.NameChanged || diff.DescChanged || diff.ScopesChanged || diff.EncryptedChanged
	if toVer.ChangeSummary != "" {
		diff.ChangeSummary = toVer.ChangeSummary
	}

	h.respondJSON(w, http.StatusOK, diff)
}

// HandleRollbackSecret handles POST /v1/vault/secrets/{id}/rollback
// Rolls back a secret to a previous version
func (h *Handler) HandleRollbackSecret(w http.ResponseWriter, r *http.Request) {
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

	var req RollbackSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.TargetVersion < 1 {
		apierror.WriteError(w, apierror.NewBadRequest("Target version must be at least 1"))
		return
	}

	// Verify the secret exists and belongs to this tenant
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}

	// Check if target version exists
	targetVer, err := h.repo.GetSecretVersionByNumber(r.Context(), *secretID, claims.TenantID, req.TargetVersion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get target version")
		apierror.WriteError(w, apierror.NewInternal("Failed to get target version"))
		return
	}
	if targetVer == nil {
		apierror.WriteError(w, apierror.NewNotFound("Target version not found"))
		return
	}

	// Perform the rollback
	newVersion, err := h.repo.RollbackSecret(r.Context(), *secretID, claims.TenantID, req.TargetVersion, claims.UserID, vault.ActorTypeUser)
	if err != nil {
		h.logger.WithError(err).Error("Failed to rollback secret")
		apierror.WriteError(w, apierror.NewInternal("Failed to rollback secret"))
		return
	}

	// Log audit event
	changeSummary := "Rolled back to version " + strconv.Itoa(req.TargetVersion)
	if req.Reason != "" {
		changeSummary += ": " + req.Reason
	}

	auditLog := &vault.AuditLog{
		SecretID:  secretID,
		TenantID:  claims.TenantID,
		Action:    vault.AuditActionRollback,
		ActorID:   claims.UserID.String(),
		ActorType: vault.ActorTypeUser,
		RequestID: r.Header.Get("X-Request-ID"),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		Metadata: vault.JSONMap{
			"secret_name":    secret.Name,
			"target_version": req.TargetVersion,
			"new_version":    newVersion.VersionNumber,
			"reason":         req.Reason,
		},
		Success: true,
	}
	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}

	// Get the updated secret
	updatedSecret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get updated secret after rollback")
	}

	resp := RollbackSecretResponse{
		NewVersion:   secretVersionToResponse(newVersion, false),
		RolledBackTo: req.TargetVersion,
		Message:      "Secret rolled back to version " + strconv.Itoa(req.TargetVersion) + " (created version " + strconv.Itoa(newVersion.VersionNumber) + ")",
	}

	if updatedSecret != nil {
		resp.Secret = secretToResponse(updatedSecret)
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// scopesEqual compares two string slices for equality
func scopesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
