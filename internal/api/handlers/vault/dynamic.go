package vault

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ============================================================================
// Phase 2.3: Dynamic secret targets
// ============================================================================

// HandleCreateTarget handles POST /v1/vault/dynamic-secret-targets.
func (h *Handler) HandleCreateTarget(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req CreateTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if err := validateCreateTarget(&req); err != nil {
		apierror.LogAndBadRequest(w, r, err, "vault dynamic credential")
		return
	}

	encrypted, nonce, keyVersion, err := vault.EncryptAdminPasswordForTarget(
		[]byte(req.AdminPassword), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt admin password")
		apierror.WriteError(w, apierror.NewInternal("Failed to encrypt target credentials"))
		return
	}

	target := &vault.DynamicSecretTarget{
		TenantID:               claims.TenantID,
		Name:                   strings.TrimSpace(req.Name),
		Description:            req.Description,
		DBType:                 vault.DynamicSecretDBType(req.DBType),
		Host:                   strings.TrimSpace(req.Host),
		Port:                   req.Port,
		DatabaseName:           strings.TrimSpace(req.DatabaseName),
		AdminUsername:          strings.TrimSpace(req.AdminUsername),
		EncryptedAdminPassword: encrypted,
		PasswordNonce:          nonce,
		PasswordKeyVersion:     keyVersion,
		SSLMode:                req.SSLMode,
		AllowedRoles:           vault.StringArray(req.AllowedRoles),
		DefaultTTLSeconds:      req.DefaultTTLSeconds,
		MaxTTLSeconds:          req.MaxTTLSeconds,
		Status:                 "active",
		CreatedBy:              claims.UserID,
	}
	if target.SSLMode == "" {
		target.SSLMode = "require"
	}
	if target.DefaultTTLSeconds == 0 {
		target.DefaultTTLSeconds = 3600
	}
	if target.MaxTTLSeconds == 0 {
		target.MaxTTLSeconds = 86400
	}
	if err := h.repo.CreateTarget(r.Context(), target); err != nil {
		h.logger.WithError(err).Error("Failed to create target")
		apierror.WriteError(w, apierror.NewInternal("Failed to create target"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation":   "dynamic_target_create",
			"target_id":   target.ID.String(),
			"target_name": target.Name,
			"db_type":     string(target.DBType),
		})
	h.respondJSON(w, http.StatusCreated, targetResponse(target))
}

// HandleListTargets handles GET /v1/vault/dynamic-secret-targets.
func (h *Handler) HandleListTargets(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	limit := parseLimit(r, 50)
	offset := parseOffset(r)
	rows, err := h.repo.ListTargets(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list targets"))
		return
	}
	resp := make([]TargetResponse, len(rows))
	for i := range rows {
		resp[i] = targetResponse(&rows[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"targets": resp, "total": len(resp)})
}

// HandleDeleteTarget handles DELETE /v1/vault/dynamic-secret-targets/{id}.
func (h *Handler) HandleDeleteTarget(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	targetID := parseUUID(vars["id"])
	if targetID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid target ID"))
		return
	}
	if err := h.repo.DeleteTarget(r.Context(), *targetID, claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Target not found"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionDelete, true, "", vault.JSONMap{
			"operation": "dynamic_target_delete",
			"target_id": targetID.String(),
		})
	w.WriteHeader(http.StatusNoContent)
}

// HandleTestTarget handles POST /v1/vault/dynamic-secret-targets/{id}/test.
// It tries a one-shot admin login; no user is created.
func (h *Handler) HandleTestTarget(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	targetID := parseUUID(vars["id"])
	if targetID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid target ID"))
		return
	}
	target, err := h.repo.GetTarget(r.Context(), *targetID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if target == nil {
		apierror.WriteError(w, apierror.NewNotFound("Target not found"))
		return
	}
	svc := h.DynamicService
	if svc == nil {
		apierror.WriteError(w, apierror.NewInternal("Dynamic service not configured"))
		return
	}
	// Issue + immediate revoke is overkill for a test; we just try a
	// no-op ping by issuing a 60s credential and revoking it. This
	// exercises the same code path a real issuance would.
	cred := &vault.DynamicCredential{
		ID:            uuid.Nil,
		TenantID:      target.TenantID,
		TargetID:      target.ID,
		Name:          "__test__",
		TTLSeconds:    60,
		MaxTTLSeconds: 60,
		Status:        "active",
		CreatedBy:     claims.UserID,
	}
	lease, material, err := svc.Issue(r.Context(), cred, target, 60*time.Second, &claims.UserID, getClientIP(r))
	if err != nil {
		apierror.LogAndBadRequest(w, r, err, "vault target connection test")
		return
	}
	if err := svc.Revoke(r.Context(), lease, target, "test"); err != nil {
		h.logger.WithError(err).Warn("Failed to revoke test credential")
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"username":   material.Username,
		"expires_at": material.ExpiresAt,
	})
}

// ============================================================================
// Phase 2.1: Dynamic credentials (templates)
// ============================================================================

// HandleCreateCredential handles POST /v1/vault/dynamic-credentials.
func (h *Handler) HandleCreateCredential(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req CreateCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	targetID := parseUUID(req.TargetID)
	if targetID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("target_id is required"))
		return
	}
	target, err := h.repo.GetTarget(r.Context(), *targetID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if target == nil {
		apierror.WriteError(w, apierror.NewNotFound("Target not found"))
		return
	}

	if h.quotaEnforcer != nil {
		decision, err := h.quotaEnforcer.CheckDynamicCreds(r.Context(), claims.TenantID)
		if err != nil {
			h.logger.WithError(err).Warn("Failed to check dynamic creds quota")
		} else if !decision.Allowed {
			h.respondErrorCode(w, http.StatusForbidden, apierror.ErrorCode("DYNAMIC_CREDS_QUOTA_EXCEEDED"),
				"Dynamic credentials quota exceeded. Upgrade your plan to create more credentials.")
			return
		}
	}

	cred := &vault.DynamicCredential{
		TenantID:      claims.TenantID,
		TargetID:      *targetID,
		Name:          strings.TrimSpace(req.Name),
		Description:   req.Description,
		RoleTemplate:  req.RoleTemplate,
		TTLSeconds:    req.TTLSeconds,
		MaxTTLSeconds: req.MaxTTLSeconds,
		Status:        "active",
		CreatedBy:     claims.UserID,
	}
	if cred.TTLSeconds == 0 {
		cred.TTLSeconds = target.DefaultTTLSeconds
	}
	if cred.MaxTTLSeconds == 0 {
		cred.MaxTTLSeconds = target.MaxTTLSeconds
	}
	if cred.TTLSeconds > cred.MaxTTLSeconds {
		apierror.WriteError(w, apierror.NewBadRequest("ttl_seconds cannot exceed max_ttl_seconds"))
		return
	}
	if err := h.repo.CreateCredential(r.Context(), cred); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create credential template"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation":       "dynamic_credential_create",
			"credential_id":   cred.ID.String(),
			"target_id":       targetID.String(),
			"credential_name": cred.Name,
		})
	h.respondJSON(w, http.StatusCreated, credentialResponse(cred))
}

// HandleListCredentials handles GET /v1/vault/dynamic-credentials.
func (h *Handler) HandleListCredentials(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	limit := parseLimit(r, 50)
	offset := parseOffset(r)
	rows, err := h.repo.ListCredentials(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list credentials"))
		return
	}
	resp := make([]CredentialResponse, len(rows))
	for i := range rows {
		resp[i] = credentialResponse(&rows[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"credentials": resp, "total": len(resp)})
}

// HandleGenerateCredential handles POST /v1/vault/dynamic-credentials/{id}/generate.
// Returns a fresh lease + credential material.
func (h *Handler) HandleGenerateCredential(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	credID := parseUUID(vars["id"])
	if credID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid credential ID"))
		return
	}
	cred, err := h.repo.GetCredential(r.Context(), *credID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load credential"))
		return
	}
	if cred == nil {
		apierror.WriteError(w, apierror.NewNotFound("Credential not found"))
		return
	}
	if cred.Status != "active" {
		apierror.WriteError(w, apierror.NewConflict("Credential is disabled"))
		return
	}
	target, err := h.repo.GetTarget(r.Context(), cred.TargetID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if target == nil || target.Status != "active" {
		apierror.WriteError(w, apierror.NewConflict("Target is disabled"))
		return
	}
	if h.DynamicService == nil {
		apierror.WriteError(w, apierror.NewInternal("Dynamic service not configured"))
		return
	}

	// Optional body: { "ttl_seconds": 1800 } overrides default.
	var body struct {
		TTLSeconds int `json:"ttl_seconds,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	ttl := time.Duration(cred.TTLSeconds) * time.Second
	if body.TTLSeconds > 0 {
		ttl = time.Duration(body.TTLSeconds) * time.Second
		if ttl > time.Duration(cred.MaxTTLSeconds)*time.Second {
			apierror.WriteError(w, apierror.NewBadRequest("ttl_seconds exceeds max_ttl_seconds"))
			return
		}
	}

	lease, material, err := h.DynamicService.Issue(
		r.Context(), cred, target, ttl, &claims.UserID, getClientIP(r),
	)
	if err != nil {
		apierror.LogAndInternal(w, r, err, "vault generate credential")
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation":     "dynamic_credential_generate",
			"credential_id": cred.ID.String(),
			"lease_id":      lease.LeaseID,
			"target_id":     target.ID.String(),
			"username":      material.Username,
		})
	h.respondJSON(w, http.StatusCreated, generatedCredentialResponse(lease, material, cred, target))
}

// HandleRevokeCredential handles DELETE /v1/vault/dynamic-credentials/{id}/revoke.
// Revokes the *most recent active lease* for the credential.
func (h *Handler) HandleRevokeCredential(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	credID := parseUUID(vars["id"])
	if credID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid credential ID"))
		return
	}
	cred, err := h.repo.GetCredential(r.Context(), *credID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load credential"))
		return
	}
	if cred == nil {
		apierror.WriteError(w, apierror.NewNotFound("Credential not found"))
		return
	}
	target, err := h.repo.GetTarget(r.Context(), cred.TargetID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if h.DynamicService == nil {
		apierror.WriteError(w, apierror.NewInternal("Dynamic service not configured"))
		return
	}
	leases, err := h.repo.ListLeasesByCredential(r.Context(), cred.ID, 50)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list leases"))
		return
	}
	revoked := 0
	for i := range leases {
		lease := leases[i]
		if lease.RevokedAt != nil {
			continue
		}
		if !lease.IsActive(time.Now()) {
			continue
		}
		if err := h.DynamicService.Revoke(r.Context(), &lease, target, "manual_revocation"); err != nil {
			h.logger.WithError(err).Warn("Failed to revoke lease")
			continue
		}
		revoked++
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionRevoke, true, "", vault.JSONMap{
			"operation":     "dynamic_credential_revoke",
			"credential_id": cred.ID.String(),
			"revoked":       revoked,
		})
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"credential_id": cred.ID,
		"revoked":       revoked,
	})
}

// ============================================================================
// Phase 2.2: Leases
// ============================================================================

// HandleRenewLease handles POST /v1/vault/dynamic-credentials/{id}/leases/{lease_id}/renew.
func (h *Handler) HandleRenewLease(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	leaseID := vars["lease_id"]
	if leaseID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("lease_id is required"))
		return
	}
	lease, err := h.repo.GetLeaseByLeaseID(r.Context(), leaseID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load lease"))
		return
	}
	if lease == nil || lease.TenantID != claims.TenantID {
		apierror.WriteError(w, apierror.NewNotFound("Lease not found"))
		return
	}
	cred, err := h.repo.GetCredential(r.Context(), lease.CredentialID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load credential"))
		return
	}
	target, err := h.repo.GetTarget(r.Context(), lease.TargetID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if h.DynamicService == nil {
		apierror.WriteError(w, apierror.NewInternal("Dynamic service not configured"))
		return
	}

	var body struct {
		TTLSeconds int `json:"ttl_seconds,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	ttl := time.Duration(cred.TTLSeconds) * time.Second
	if body.TTLSeconds > 0 {
		ttl = time.Duration(body.TTLSeconds) * time.Second
	}
	newExpires, err := h.DynamicService.Renew(r.Context(), lease, target, ttl)
	if err != nil {
		apierror.LogAndBadRequest(w, r, err, "vault renew credential")
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{
			"operation":  "dynamic_lease_renew",
			"lease_id":   lease.LeaseID,
			"expires_at": newExpires,
		})
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"lease_id":   lease.LeaseID,
		"expires_at": newExpires,
	})
}

// HandleRevokeLease handles POST /v1/vault/dynamic-credentials/{id}/leases/{lease_id}/revoke.
func (h *Handler) HandleRevokeLease(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	leaseID := vars["lease_id"]
	if leaseID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("lease_id is required"))
		return
	}
	lease, err := h.repo.GetLeaseByLeaseID(r.Context(), leaseID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load lease"))
		return
	}
	if lease == nil || lease.TenantID != claims.TenantID {
		apierror.WriteError(w, apierror.NewNotFound("Lease not found"))
		return
	}
	target, err := h.repo.GetTarget(r.Context(), lease.TargetID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if h.DynamicService == nil {
		apierror.WriteError(w, apierror.NewInternal("Dynamic service not configured"))
		return
	}
	if err := h.DynamicService.Revoke(r.Context(), lease, target, "manual_lease_revoke"); err != nil {
		apierror.LogAndInternal(w, r, err, "vault revoke credential")
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionRevoke, true, "", vault.JSONMap{
			"operation": "dynamic_lease_revoke",
			"lease_id":  lease.LeaseID,
		})
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"lease_id": lease.LeaseID,
		"revoked":  true,
	})
}

// ============================================================================
// helpers
// ============================================================================

func validateCreateTarget(req *CreateTargetRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errorStr("name is required")
	}
	dbType := vault.DynamicSecretDBType(req.DBType)
	if !dbType.Valid() {
		return errorStr("db_type must be 'postgres' or 'mysql'")
	}
	if strings.TrimSpace(req.Host) == "" {
		return errorStr("host is required")
	}
	if req.Port < 1 || req.Port > 65535 {
		return errorStr("port must be between 1 and 65535")
	}
	if strings.TrimSpace(req.DatabaseName) == "" {
		return errorStr("database_name is required")
	}
	if strings.TrimSpace(req.AdminUsername) == "" {
		return errorStr("admin_username is required")
	}
	if req.AdminPassword == "" {
		return errorStr("admin_password is required")
	}
	if req.DefaultTTLSeconds < 0 || req.DefaultTTLSeconds > 86400*7 {
		return errorStr("default_ttl_seconds must be between 0 and 604800")
	}
	if req.MaxTTLSeconds < 0 || req.MaxTTLSeconds > 86400*30 {
		return errorStr("max_ttl_seconds must be between 0 and 2592000")
	}
	if req.DefaultTTLSeconds > req.MaxTTLSeconds {
		return errorStr("default_ttl_seconds cannot exceed max_ttl_seconds")
	}
	return nil
}

func parseLimit(r *http.Request, defaultLimit int) int {
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 200 {
				return 200
			}
			return n
		}
	}
	return defaultLimit
}

func parseOffset(r *http.Request) int {
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 0
}

type stringError string

func (e stringError) Error() string { return string(e) }
func errorStr(s string) error       { return stringError(s) }

func targetResponse(t *vault.DynamicSecretTarget) TargetResponse {
	return TargetResponse{
		ID:                t.ID,
		TenantID:          t.TenantID,
		Name:              t.Name,
		Description:       t.Description,
		DBType:            string(t.DBType),
		Host:              t.Host,
		Port:              t.Port,
		DatabaseName:      t.DatabaseName,
		AdminUsername:     t.AdminUsername,
		SSLMode:           t.SSLMode,
		AllowedRoles:      []string(t.AllowedRoles),
		DefaultTTLSeconds: t.DefaultTTLSeconds,
		MaxTTLSeconds:     t.MaxTTLSeconds,
		Status:            t.Status,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
		LastUsedAt:        t.LastUsedAt,
	}
}

func credentialResponse(c *vault.DynamicCredential) CredentialResponse {
	return CredentialResponse{
		ID:            c.ID,
		TenantID:      c.TenantID,
		TargetID:      c.TargetID,
		Name:          c.Name,
		Description:   c.Description,
		RoleTemplate:  c.RoleTemplate,
		TTLSeconds:    c.TTLSeconds,
		MaxTTLSeconds: c.MaxTTLSeconds,
		Status:        c.Status,
		CreatedBy:     c.CreatedBy,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func generatedCredentialResponse(
	lease *vault.DynamicCredentialLease,
	material *vault.DynamicCredentialMaterial,
	cred *vault.DynamicCredential,
	target *vault.DynamicSecretTarget,
) GeneratedCredentialResponse {
	return GeneratedCredentialResponse{
		LeaseID:    lease.LeaseID,
		Username:   material.Username,
		Password:   material.Password,
		Host:       material.Host,
		Port:       material.Port,
		Database:   material.Database,
		ExpiresAt:  material.ExpiresAt,
		Credential: credentialResponse(cred),
		Target:     targetResponse(target),
	}
}
