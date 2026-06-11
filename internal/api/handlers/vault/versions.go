package vault

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/pagination"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleListSecretVersions handles GET /v1/vault/secrets/{id}/versions
// Lists all versions for a secret (metadata only, no encrypted data)
func (h *Handler) HandleListSecretVersions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondErrorStandard(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	if err := middleware.ValidateUUIDParam(vars, "id"); err != nil {
		h.respondErrorStandard(w, err)
		return
	}
	secretID, _ := uuid.Parse(vars["id"])

	params, err := pagination.ParseParams(r)
	if err != nil {
		h.respondErrorStandard(w, err)
		return
	}

	secret, err := h.repo.GetSecretByID(r.Context(), secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("Secret not found"))
		return
	}

	total, err := h.repo.CountSecretVersions(r.Context(), secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count versions")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to count versions"))
		return
	}

	versions, err := h.repo.GetSecretVersions(r.Context(), secretID, claims.TenantID, params.Limit, params.Offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list versions")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to list versions"))
		return
	}

	responses := make([]SecretVersionMetadataResponse, len(versions))
	for i, v := range versions {
		responses[i] = secretVersionToMetadataResponse(&v)
	}

	h.respondPaginated(w, responses, total, params)
}

// HandleGetSecretVersion handles GET /v1/vault/secrets/{id}/versions/{version}
// Gets a specific version of a secret
func (h *Handler) HandleGetSecretVersion(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondErrorStandard(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	if err := middleware.ValidateUUIDParam(vars, "id"); err != nil {
		h.respondErrorStandard(w, err)
		return
	}
	secretID, _ := uuid.Parse(vars["id"])

	versionNumber, err := strconv.Atoi(vars["version"])
	if err != nil || versionNumber < 1 {
		h.respondErrorStandard(w, apierror.ValidationFieldError("version", "Invalid version number"))
		return
	}

	secret, err := h.repo.GetSecretByID(r.Context(), secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("Secret not found"))
		return
	}

	version, err := h.repo.GetSecretVersionByNumber(r.Context(), secretID, claims.TenantID, versionNumber)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get version")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get version"))
		return
	}
	if version == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("Version not found"))
		return
	}

	includeEncrypted := r.URL.Query().Get("include_encrypted") == "true"

	h.respondJSON(w, http.StatusOK, secretVersionToResponse(version, includeEncrypted))
}

// HandleDiffSecretVersions handles GET /v1/vault/secrets/{id}/versions/diff
// Compares two versions of a secret and returns the differences
// Query params: from_version (int), to_version (int, optional, defaults to current)
func (h *Handler) HandleDiffSecretVersions(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondErrorStandard(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	if err := middleware.ValidateUUIDParam(vars, "id"); err != nil {
		h.respondErrorStandard(w, err)
		return
	}
	secretID, _ := uuid.Parse(vars["id"])

	fromVersionStr := r.URL.Query().Get("from_version")
	if fromVersionStr == "" {
		h.respondErrorStandard(w, apierror.ValidationFieldError("from_version", "This parameter is required"))
		return
	}

	fromVersion, err := strconv.Atoi(fromVersionStr)
	if err != nil || fromVersion < 1 {
		h.respondErrorStandard(w, apierror.ValidationFieldError("from_version", "Invalid from_version"))
		return
	}

	toVersionStr := r.URL.Query().Get("to_version")

	secret, err := h.repo.GetSecretByID(r.Context(), secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("Secret not found"))
		return
	}

	toVersion := fromVersion + 1
	if toVersionStr != "" {
		toVersion, err = strconv.Atoi(toVersionStr)
		if err != nil || toVersion < 1 {
			h.respondErrorStandard(w, apierror.ValidationFieldError("to_version", "Invalid to_version"))
			return
		}
	} else if secret.CurrentVersion != nil {
		toVersion = *secret.CurrentVersion
	}

	if fromVersion >= toVersion {
		h.respondErrorStandard(w, apierror.NewBadRequest("from_version must be less than to_version"))
		return
	}

	fromVer, err := h.repo.GetSecretVersionByNumber(r.Context(), secretID, claims.TenantID, fromVersion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get from version")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get version"))
		return
	}
	if fromVer == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("From version not found"))
		return
	}

	toVer, err := h.repo.GetSecretVersionByNumber(r.Context(), secretID, claims.TenantID, toVersion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get to version")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get version"))
		return
	}
	if toVer == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("To version not found"))
		return
	}

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
		h.respondErrorStandard(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	vars := mux.Vars(r)
	if err := middleware.ValidateUUIDParam(vars, "id"); err != nil {
		h.respondErrorStandard(w, err)
		return
	}
	secretID, _ := uuid.Parse(vars["id"])

	var req RollbackSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondErrorStandard(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.TargetVersion < 1 {
		h.respondErrorStandard(w, apierror.ValidationFieldError("target_version", "Target version must be at least 1"))
		return
	}

	secret, err := h.repo.GetSecretByID(r.Context(), secretID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get secret")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get secret"))
		return
	}
	if secret == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("Secret not found"))
		return
	}

	targetVer, err := h.repo.GetSecretVersionByNumber(r.Context(), secretID, claims.TenantID, req.TargetVersion)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get target version")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to get target version"))
		return
	}
	if targetVer == nil {
		h.respondErrorStandard(w, apierror.NewNotFound("Target version not found"))
		return
	}

	newVersion, err := h.repo.RollbackSecret(r.Context(), secretID, claims.TenantID, req.TargetVersion, claims.UserID, vault.ActorTypeUser)
	if err != nil {
		h.logger.WithError(err).Error("Failed to rollback secret")
		h.respondErrorStandard(w, apierror.NewInternal("Failed to rollback secret"))
		return
	}

	changeSummary := "Rolled back to version " + strconv.Itoa(req.TargetVersion)
	if req.Reason != "" {
		changeSummary += ": " + req.Reason
	}

	auditLog := &vault.AuditLog{
		SecretID:  &secretID,
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

	updatedSecret, err := h.repo.GetSecretByID(r.Context(), secretID, claims.TenantID)
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
