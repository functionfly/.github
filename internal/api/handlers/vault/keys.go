package vault

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/google/uuid"
)

// ============================================================================
// Per-tenant DEK endpoints (2026-06-16 design)
// ============================================================================

// HandleGetWrappedDEK returns the current user's wrapped DEK or 404 if
// not yet initialized. Authentication: JWT only.
func (h *Handler) HandleGetWrappedDEK(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	k, err := h.repo.GetTenantKey(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load DEK"))
		return
	}
	if k == nil {
		apierror.WriteError(w, apierror.NewNotFound("DEK not initialized"))
		return
	}
	h.respondJSON(w, http.StatusOK, wrappedDEKResponse(k))
}

// HandleUpsertWrappedDEK creates or rotates the current user's
// wrapped DEK. The body carries the ciphertext produced by the
// client; the server stores it opaquely. Authenticated + MFA-gated
// when the tenant has VaultMFAConfig.EnforceForAPI.
func (h *Handler) HandleUpsertWrappedDEK(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req UpsertWrappedDEKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	wrapped, err := base64.StdEncoding.DecodeString(req.WrappedDEK)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("wrapped_dek must be base64"))
		return
	}
	iv, err := base64.StdEncoding.DecodeString(req.DEKIV)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("dek_iv must be base64"))
		return
	}
	tag, err := base64.StdEncoding.DecodeString(req.DEKAuthTag)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("dek_auth_tag must be base64"))
		return
	}
	salt, err := base64.StdEncoding.DecodeString(req.DEKSalt)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("dek_salt must be base64"))
		return
	}
	if req.KeyVersion == 0 {
		req.KeyVersion = vault.ClientWrapKeyVersion
	}
	if err := vault.ValidateClientWrapEnvelope(wrapped, iv, tag); err != nil {
		apierror.LogAndBadRequest(w, r, err, "vault keys handler")
		return
	}
	k := &vault.VaultTenantKey{
		TenantID:   claims.TenantID,
		UserID:     claims.UserID,
		WrappedDEK: wrapped,
		DEKIV:      iv,
		DEKAuthTag: tag,
		DEKSalt:    salt,
		KeyVersion: req.KeyVersion,
		KDFParams:  vault.JSONMap{"params": req.KDFParams},
	}
	if k.KDFParams == nil {
		k.KDFParams = vault.JSONMap{}
	}
	if err := h.repo.UpsertTenantKey(r.Context(), k); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to persist DEK"))
		return
	}
	action := vault.AuditActionClientDekInit
	existing, _ := h.repo.GetTenantKey(r.Context(), claims.TenantID, claims.UserID)
	if existing != nil && !existing.CreatedAt.IsZero() {
		action = vault.AuditActionClientWrapRotate
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		action, true, "", vault.JSONMap{
			"operation":   string(action),
			"key_version": req.KeyVersion,
		})
	h.respondJSON(w, http.StatusOK, wrappedDEKResponse(k))
}

// HandleRotateWrappedDEK is the explicit rotate endpoint. The client
// generates a fresh DEK, re-wraps all client-mode targets, and POSTs
// the new wrapped DEK in the same shape as Upsert.
func (h *Handler) HandleRotateWrappedDEK(w http.ResponseWriter, r *http.Request) {
	h.HandleUpsertWrappedDEK(w, r)
}

// HandleShareDEK lets an admin re-wrap the per-tenant DEK for a
// target user in the same tenant. RBAC: rbac:manage (admin).
func (h *Handler) HandleShareDEK(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	if h.RBAC == nil {
		apierror.WriteError(w, apierror.NewInternal("RBAC not configured"))
		return
	}
	dec, err := h.RBAC.Check(r.Context(), claims.TenantID, claims.UserID,
		"dynamic_credentials:client_manage_keys", "/")
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to check permission"))
		return
	}
	if !dec.Allowed {
		apierror.WriteError(w, apierror.NewForbidden("Permission denied"))
		return
	}
	var req ShareDEKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.TargetUserID == uuid.Nil {
		apierror.WriteError(w, apierror.NewBadRequest("target_user_id is required"))
		return
	}
	wrapped, err := base64.StdEncoding.DecodeString(req.WrappedDEK)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("wrapped_dek must be base64"))
		return
	}
	iv, err := base64.StdEncoding.DecodeString(req.DEKIV)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("dek_iv must be base64"))
		return
	}
	tag, err := base64.StdEncoding.DecodeString(req.DEKAuthTag)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("dek_auth_tag must be base64"))
		return
	}
	salt, err := base64.StdEncoding.DecodeString(req.DEKSalt)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("dek_salt must be base64"))
		return
	}
	if err := vault.ValidateClientWrapEnvelope(wrapped, iv, tag); err != nil {
		apierror.LogAndBadRequest(w, r, err, "vault keys handler")
		return
	}
	k := &vault.VaultTenantKey{
		TenantID:   claims.TenantID,
		UserID:     req.TargetUserID,
		WrappedDEK: wrapped,
		DEKIV:      iv,
		DEKAuthTag: tag,
		DEKSalt:    salt,
		KeyVersion: req.KeyVersion,
		KDFParams:  vault.JSONMap{"params": req.KDFParams},
	}
	if k.KeyVersion == 0 {
		k.KeyVersion = vault.ClientWrapKeyVersion
	}
	if k.KDFParams == nil {
		k.KDFParams = vault.JSONMap{}
	}
	if err := h.repo.UpsertTenantKey(r.Context(), k); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to persist shared DEK"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionClientDekShare, true, "", vault.JSONMap{
			"operation":      "client_dek_share",
			"target_user_id": req.TargetUserID.String(),
		})
	h.respondJSON(w, http.StatusOK, wrappedDEKResponse(k))
}

func wrappedDEKResponse(k *vault.VaultTenantKey) WrappedDEKResponse {
	return WrappedDEKResponse{
		TenantID:   k.TenantID,
		UserID:     k.UserID,
		WrappedDEK: base64.StdEncoding.EncodeToString(k.WrappedDEK),
		DEKIV:      base64.StdEncoding.EncodeToString(k.DEKIV),
		DEKAuthTag: base64.StdEncoding.EncodeToString(k.DEKAuthTag),
		DEKSalt:    base64.StdEncoding.EncodeToString(k.DEKSalt),
		KeyVersion: k.KeyVersion,
		KDFParams:  "",
		CreatedAt:  k.CreatedAt,
		RotatedAt:  k.RotatedAt,
	}
}
