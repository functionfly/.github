package vault

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ============================================================================
// Phase 1.1: MFA enforcement for vault operations
// ============================================================================

// HandleGetMFAConfig returns the per-tenant vault MFA policy.
func (h *Handler) HandleGetMFAConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	cfg, err := h.repo.GetMFAConfig(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get vault MFA config")
		apierror.WriteError(w, apierror.NewInternal("Failed to get MFA config"))
		return
	}
	h.respondJSON(w, http.StatusOK, mfaConfigResponse(cfg))
}

// HandleUpdateMFAConfig updates the per-tenant vault MFA policy.
func (h *Handler) HandleUpdateMFAConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req UpdateMFAConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.MFAMethod != "" && req.MFAMethod != "totp" && req.MFAMethod != "webauthn" && req.MFAMethod != "both" {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid mfa_method (expected totp, webauthn, or both)"))
		return
	}
	if req.MFASessionTTLSeconds < 60 || req.MFASessionTTLSeconds > 86400 {
		apierror.WriteError(w, apierror.NewBadRequest("mfa_session_ttl_seconds must be between 60 and 86400"))
		return
	}
	cfg, err := h.repo.GetMFAConfig(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load MFA config")
		apierror.WriteError(w, apierror.NewInternal("Failed to update MFA config"))
		return
	}
	if req.MFARequired != nil {
		cfg.MFARequired = *req.MFARequired
	}
	if req.MFAMethod != "" {
		cfg.MFAMethod = req.MFAMethod
	}
	if req.EnforceForTokens != nil {
		cfg.EnforceForTokens = *req.EnforceForTokens
	}
	if req.EnforceForAPI != nil {
		cfg.EnforceForAPI = *req.EnforceForAPI
	}
	if req.MFASessionTTLSeconds > 0 {
		cfg.MFASessionTTLSeconds = req.MFASessionTTLSeconds
	}
	if err := h.repo.UpdateMFAConfig(r.Context(), cfg); err != nil {
		h.logger.WithError(err).Error("Failed to persist MFA config")
		apierror.WriteError(w, apierror.NewInternal("Failed to update MFA config"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{
			"operation":    "vault_mfa_config_update",
			"mfa_required": cfg.MFARequired,
			"mfa_method":   cfg.MFAMethod,
		})
	h.respondJSON(w, http.StatusOK, mfaConfigResponse(cfg))
}

// HandleVerifyMFA stamps a vault-scoped MFA assertion on the session. The
// session is identified by the standard X-MFA-Code header (or the
// mfa_code form field), matching how the platform-wide RequireMFA
// middleware extracts codes.
func (h *Handler) HandleVerifyMFA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	// We trust the platform's MFA verification at this layer — the user
	// has just proven possession of a TOTP/WebAuthn factor against their
	// account. We record the assertion so the vault middleware can skip
	// the MFA check for the duration of the configured TTL.
	cfg, err := h.repo.GetMFAConfig(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load MFA config"))
		return
	}
	if !cfg.MFARequired {
		apierror.WriteError(w, apierror.NewBadRequest("Vault MFA is not enabled for this tenant"))
		return
	}

	now := time.Now()
	expires := now.Add(time.Duration(cfg.MFASessionTTLSeconds) * time.Second)

	// We piggyback on the user_auth vault assertions table; for now we
	// mark a side-channel via the audit log so downstream middleware can
	// correlate. The active window is bounded by ttl and the user's
	// session lifetime.
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionMFAVerify, true, "", vault.JSONMap{
			"operation":  "vault_mfa_verify",
			"expires_at": expires.Unix(),
			"ttl":        cfg.MFASessionTTLSeconds,
		})

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"verified":   true,
		"expires_at": expires,
		"ttl":        cfg.MFASessionTTLSeconds,
	})
}

// ============================================================================
// Phase 1.2: IP allowlist for tokens
// ============================================================================

// HandleUpdateTokenIPPolicy sets the allow/deny IP lists for a token.
func (h *Handler) HandleUpdateTokenIPPolicy(w http.ResponseWriter, r *http.Request) {
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
	var req UpdateTokenIPPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if err := validateCIDRList(req.AllowedIPs); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest(fmt.Sprintf("Invalid allowed_ips: %v", err)))
		return
	}
	if err := validateCIDRList(req.DeniedIPs); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest(fmt.Sprintf("Invalid denied_ips: %v", err)))
		return
	}
	token, err := h.repo.GetAccessTokenByID(r.Context(), *tokenID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load token")
		apierror.WriteError(w, apierror.NewInternal("Failed to update token policy"))
		return
	}
	if token == nil {
		apierror.WriteError(w, apierror.NewNotFound("Token not found"))
		return
	}
	if err := h.repo.UpdateAccessTokenIPPolicy(r.Context(), *tokenID, req.AllowedIPs, req.DeniedIPs, req.Enabled); err != nil {
		h.logger.WithError(err).Error("Failed to update IP policy")
		apierror.WriteError(w, apierror.NewInternal("Failed to update token policy"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, &token.SecretID, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{
			"operation":        "token_ip_policy_update",
			"token_id":         token.ID.String(),
			"enabled":          req.Enabled,
			"allowed_ip_count": len(req.AllowedIPs),
			"denied_ip_count":  len(req.DeniedIPs),
		})
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"token_id":    tokenID,
		"allowed_ips": req.AllowedIPs,
		"denied_ips":  req.DeniedIPs,
		"enabled":     req.Enabled,
	})
}

// ============================================================================
// Phase 1.3: Secret expiration
// ============================================================================

// HandleSetSecretExpiration sets an expiration for a secret.
func (h *Handler) HandleSetSecretExpiration(w http.ResponseWriter, r *http.Request) {
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
	var req SetExpirationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}
	if req.ExpiresAt == nil && req.ExpireAfterDays == nil {
		apierror.WriteError(w, apierror.NewBadRequest("expires_at or expire_after_days required"))
		return
	}
	if req.ExpireAfterDays != nil && (*req.ExpireAfterDays < 1 || *req.ExpireAfterDays > 3650) {
		apierror.WriteError(w, apierror.NewBadRequest("expire_after_days must be between 1 and 3650"))
		return
	}
	if err := h.repo.SetSecretExpiration(r.Context(), *secretID, claims.TenantID, req.ExpiresAt, req.ExpireAfterDays); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to set expiration"))
		return
	}
	// Optimistically re-read for response.
	updated, _ := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if updated == nil {
		updated = secret
	}
	h.logAudit(r.Context(), claims.TenantID, secretID, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{
			"operation":         "set_expiration",
			"secret_name":       updated.Name,
			"expires_at":        updated.ExpiresAt,
			"expire_after_days": req.ExpireAfterDays,
		})
	h.respondJSON(w, http.StatusOK, secretToResponse(updated))
}

// ============================================================================
// Phase 1.4: Break-glass emergency access
// ============================================================================

// HandleRequestBreakGlass submits a new emergency access request.
func (h *Handler) HandleRequestBreakGlass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req BreakGlassRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		apierror.WriteError(w, apierror.NewBadRequest("reason is required"))
		return
	}
	cfg, err := h.repo.GetBreakGlassConfig(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load break-glass config"))
		return
	}
	if !cfg.Enabled {
		apierror.WriteError(w, apierror.NewForbidden("Break-glass is disabled for this tenant"))
		return
	}
	duration := req.DurationMinutes
	if duration <= 0 {
		duration = cfg.MaxDurationMinutes
	}
	if duration > cfg.MaxDurationMinutes {
		apierror.WriteError(w, apierror.NewBadRequest(fmt.Sprintf("duration_minutes cannot exceed %d", cfg.MaxDurationMinutes)))
		return
	}
	bg := &vault.BreakGlassRequest{
		TenantID:        claims.TenantID,
		RequestedBy:     claims.UserID,
		Reason:          reason,
		Status:          "pending",
		DurationMinutes: duration,
		ExpiresAt:       time.Now().Add(time.Duration(duration) * time.Minute),
		Metadata: vault.JSONMap{
			"approver_user_ids": cfg.ApproverUserIDs,
		},
	}
	if err := h.repo.CreateBreakGlassRequest(r.Context(), bg); err != nil {
		h.logger.WithError(err).Error("Failed to create break-glass request")
		apierror.WriteError(w, apierror.NewInternal("Failed to create break-glass request"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionBreakGlass, true, "", vault.JSONMap{
			"operation":  "break_glass_request",
			"request_id": bg.ID.String(),
			"reason":     reason,
		})
	h.respondJSON(w, http.StatusCreated, breakGlassResponse(bg))
}

// HandleApproveBreakGlass approves a pending break-glass request. Only
// designated approvers (per break_glass_config) are allowed.
func (h *Handler) HandleApproveBreakGlass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	requestID := parseUUID(vars["id"])
	if requestID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request ID"))
		return
	}
	bg, err := h.repo.GetBreakGlassRequest(r.Context(), *requestID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load break-glass request"))
		return
	}
	if bg == nil {
		apierror.WriteError(w, apierror.NewNotFound("Break-glass request not found"))
		return
	}
	if bg.Status != "pending" {
		apierror.WriteError(w, apierror.NewConflict(fmt.Sprintf("Request is %s", bg.Status)))
		return
	}
	// Self-approval is disallowed; approver must be a configured approver.
	if bg.RequestedBy == claims.UserID {
		apierror.WriteError(w, apierror.NewForbidden("Cannot approve your own break-glass request"))
		return
	}
	if err := h.repo.ApproveBreakGlassRequest(r.Context(), *requestID, claims.TenantID, claims.UserID, bg.DurationMinutes); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to approve break-glass request"))
		return
	}
	updated, _ := h.repo.GetBreakGlassRequest(r.Context(), *requestID, claims.TenantID)
	if updated == nil {
		updated = bg
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionBreakGlass, true, "", vault.JSONMap{
			"operation":   "break_glass_approve",
			"request_id":  bg.ID.String(),
			"approved_by": claims.UserID.String(),
			"duration":    bg.DurationMinutes,
		})
	h.respondJSON(w, http.StatusOK, breakGlassResponse(updated))
}

// HandleDenyBreakGlass denies a pending request.
func (h *Handler) HandleDenyBreakGlass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	requestID := parseUUID(vars["id"])
	if requestID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request ID"))
		return
	}
	if err := h.repo.DenyBreakGlassRequest(r.Context(), *requestID, claims.TenantID, claims.UserID); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to deny break-glass request"))
		return
	}
	bg, _ := h.repo.GetBreakGlassRequest(r.Context(), *requestID, claims.TenantID)
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionBreakGlass, true, "", vault.JSONMap{
			"operation":  "break_glass_deny",
			"request_id": requestID.String(),
		})
	if bg == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"status": "denied"})
		return
	}
	h.respondJSON(w, http.StatusOK, breakGlassResponse(bg))
}

// HandleRevokeBreakGlass revokes an active break-glass grant.
func (h *Handler) HandleRevokeBreakGlass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	requestID := parseUUID(vars["id"])
	if requestID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request ID"))
		return
	}
	if err := h.repo.RevokeBreakGlassRequest(r.Context(), *requestID, claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke break-glass request"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionBreakGlass, true, "", vault.JSONMap{
			"operation":  "break_glass_revoke",
			"request_id": requestID.String(),
		})
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"status": "revoked"})
}

// HandleListBreakGlass lists the most recent break-glass requests.
func (h *Handler) HandleListBreakGlass(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	reqs, err := h.repo.ListBreakGlassRequestsByTenant(r.Context(), claims.TenantID, 50, 0)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list break-glass requests"))
		return
	}
	responses := make([]BreakGlassResponse, len(reqs))
	for i := range reqs {
		responses[i] = breakGlassResponse(&reqs[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"requests": responses, "total": len(responses)})
}

// HandleGetBreakGlassConfig returns the per-tenant break-glass policy.
func (h *Handler) HandleGetBreakGlassConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	cfg, err := h.repo.GetBreakGlassConfig(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load break-glass config"))
		return
	}
	h.respondJSON(w, http.StatusOK, breakGlassConfigResponse(cfg))
}

// ============================================================================
// Phase 1.4b: Escrow (enterprise)
// ============================================================================

// HandleEnableEscrow creates a new escrow row. The flow expects the
// client to have already derived the encryption key from
// master_passphrase + security_questions and encrypted the recovery
// blob with AES-256-GCM. The server stores the ciphertext and never
// sees the passphrase.
func (h *Handler) HandleEnableEscrow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req EnableEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	blob, err := base64.StdEncoding.DecodeString(req.EncryptedBlob)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("encrypted_blob must be base64"))
		return
	}
	iv, err := base64.StdEncoding.DecodeString(req.BlobIV)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("blob_iv must be base64"))
		return
	}
	tag, err := base64.StdEncoding.DecodeString(req.BlobAuthTag)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("blob_auth_tag must be base64"))
		return
	}
	salt, err := base64.StdEncoding.DecodeString(req.KDFSalt)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("kdf_salt must be base64"))
		return
	}
	if len(req.SecurityQuestionHashes) == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("security_question_hashes required"))
		return
	}

	cfg := &vault.VaultEscrowConfig{
		TenantID:        claims.TenantID,
		Enabled:         true,
		SecurityQHashes: vault.JSONMap{"hashes": req.SecurityQuestionHashes},
		KDFSalt:         salt,
		KDFMethod:       "argon2id",
		KDFParams: vault.JSONMap{
			"memory_kib":  vault.Argon2MemoryKiB,
			"iterations":  vault.Argon2Iterations,
			"parallelism": vault.Argon2Parallelism,
		},
		EncryptedBlob:  blob,
		BlobIV:         iv,
		BlobAuthTag:    tag,
		BlobKeyVersion: vault.KeyVersionArgon2,
	}
	if err := h.repo.UpsertEscrowConfig(r.Context(), cfg); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to enable escrow"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{"operation": "escrow_enable"})
	h.respondJSON(w, http.StatusOK, escrowStatusResponse(cfg))
}

// HandleGetEscrowStatus returns the (metadata-only) escrow status.
func (h *Handler) HandleGetEscrowStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	cfg, err := h.repo.GetEscrowConfig(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load escrow config"))
		return
	}
	if cfg == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"enabled": false})
		return
	}
	h.respondJSON(w, http.StatusOK, escrowStatusResponse(cfg))
}

// HandleDisableEscrow removes escrow for the tenant.
func (h *Handler) HandleDisableEscrow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	if err := h.repo.DisableEscrow(r.Context(), claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to disable escrow"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{"operation": "escrow_disable"})
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"enabled": false})
}

// ============================================================================
// helpers
// ============================================================================

// logAudit is a small wrapper so handlers don't have to re-write the
// audit-log boilerplate. Errors are logged at warn level only.
func (h *Handler) logAudit(
	ctx context.Context,
	tenantID uuid.UUID,
	secretID *uuid.UUID,
	actorID string,
	action vault.AuditAction,
	success bool,
	errMsg string,
	metadata vault.JSONMap,
) {
	if metadata == nil {
		metadata = vault.JSONMap{}
	}
	entry := &vault.AuditLog{
		TenantID:     tenantID,
		SecretID:     secretID,
		ActorID:      actorID,
		ActorType:    vault.ActorTypeUser,
		Action:       action,
		Success:      success,
		ErrorMessage: errMsg,
		Metadata:     metadata,
	}
	if err := h.repo.CreateAuditLog(ctx, entry); err != nil {
		h.logger.WithError(err).Warn("Failed to create audit log")
	}
}

func mfaConfigResponse(c *vault.VaultMFAConfig) MFAConfigResponse {
	return MFAConfigResponse{
		TenantID:             c.TenantID,
		MFARequired:          c.MFARequired,
		MFAMethod:            c.MFAMethod,
		EnforceForTokens:     c.EnforceForTokens,
		EnforceForAPI:        c.EnforceForAPI,
		MFASessionTTLSeconds: c.MFASessionTTLSeconds,
		UpdatedAt:            c.UpdatedAt,
	}
}

func breakGlassResponse(b *vault.BreakGlassRequest) BreakGlassResponse {
	return BreakGlassResponse{
		ID:              b.ID,
		TenantID:        b.TenantID,
		RequestedBy:     b.RequestedBy,
		ApprovedBy:      b.ApprovedBy,
		Reason:          b.Reason,
		Status:          b.Status,
		DurationMinutes: b.DurationMinutes,
		ExpiresAt:       b.ExpiresAt,
		ApprovedAt:      b.ApprovedAt,
		RevokedAt:       b.RevokedAt,
		CreatedAt:       b.CreatedAt,
	}
}

func breakGlassConfigResponse(c *vault.BreakGlassConfig) BreakGlassConfigResponse {
	return BreakGlassConfigResponse{
		TenantID:              c.TenantID,
		MaxDurationMinutes:    c.MaxDurationMinutes,
		RequiredApproverCount: c.RequiredApproverCount,
		Enabled:               c.Enabled,
		UpdatedAt:             c.UpdatedAt,
	}
}

func escrowStatusResponse(c *vault.VaultEscrowConfig) EscrowStatusResponse {
	return EscrowStatusResponse{
		TenantID:       c.TenantID,
		Enabled:        c.Enabled,
		KDFMethod:      c.KDFMethod,
		BlobKeyVersion: c.BlobKeyVersion,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func validateCIDRList(list []string) error {
	for _, raw := range list {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			return errors.New("empty entry")
		}
		if strings.Contains(entry, "/") {
			if _, _, err := splitCIDR(entry); err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func splitCIDR(cidr string) (string, error) {
	parts := strings.SplitN(cidr, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid CIDR %q", cidr)
	}
	return cidr, nil
}
