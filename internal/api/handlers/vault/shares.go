package vault

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/gorilla/mux"
)

// HandleCreateTargetShare handles POST /v1/vault/dynamic-secret-targets/{id}/shares.
// It shares a dynamic secret target with another tenant.
func (h *Handler) HandleCreateTargetShare(w http.ResponseWriter, r *http.Request) {
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
		h.logger.WithError(err).Error("Failed to load target")
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if target == nil {
		apierror.WriteError(w, apierror.NewNotFound("Target not found"))
		return
	}

	if target.Status != "active" {
		apierror.WriteError(w, apierror.NewConflict("Target is not active"))
		return
	}

	var req DynamicTargetShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	granteeTenantID := parseUUID(req.GranteeTenantID)
	if granteeTenantID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("grantee_tenant_id is required"))
		return
	}

	if *granteeTenantID == claims.TenantID {
		apierror.WriteError(w, apierror.NewBadRequest("cannot share a target with the same tenant"))
		return
	}

	perms := req.Permissions
	if perms == "" {
		perms = "read"
	}
	if perms != "read" && perms != "admin" {
		apierror.WriteError(w, apierror.NewBadRequest("permissions must be 'read' or 'admin'"))
		return
	}

	if !h.requireDynamicPerm(w, r, claims, "dynamic_credentials:create", target.Namespace) {
		return
	}

	share := &vault.DynamicTargetShare{
		TargetID:          *targetID,
		SourceTenantID:    claims.TenantID,
		GrantedToTenantID: *granteeTenantID,
		GrantedByUser:     claims.UserID,
		Permissions:       perms,
		ExpiresAt:         req.ExpiresAt,
	}

	if err := h.repo.CreateDynamicTargetShare(r.Context(), share); err != nil {
		h.logger.WithError(err).Error("Failed to create target share")
		apierror.WriteError(w, apierror.NewConflict("Share already exists or invalid"))
		return
	}

	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation":       "target_share_create",
			"target_id":      targetID.String(),
			"grantee_tenant": granteeTenantID.String(),
			"permissions":    perms,
		})

	h.respondJSON(w, http.StatusCreated, dynamicTargetShareResponse(share))
}

// HandleListTargetShares handles GET /v1/vault/dynamic-secret-targets/{id}/shares.
// It lists outbound shares for a specific target (shares created by this tenant for this target).
func (h *Handler) HandleListTargetShares(w http.ResponseWriter, r *http.Request) {
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
		h.logger.WithError(err).Error("Failed to load target")
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if target == nil {
		apierror.WriteError(w, apierror.NewNotFound("Target not found"))
		return
	}

	if !h.requireDynamicPerm(w, r, claims, "dynamic_credentials:read", target.Namespace) {
		return
	}

	shares, err := h.repo.ListTargetSharesBySource(r.Context(), *targetID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list target shares")
		apierror.WriteError(w, apierror.NewInternal("Failed to list shares"))
		return
	}

	resp := make([]DynamicTargetShareResponse, len(shares))
	for i := range shares {
		resp[i] = dynamicTargetShareResponse(&shares[i])
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"shares": resp,
		"total":  len(resp),
	})
}

// HandleRevokeTargetShare handles DELETE /v1/vault/dynamic-target-shares/{id}.
// It revokes a share by its ID. Only the source tenant can revoke a share.
func (h *Handler) HandleRevokeTargetShare(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	shareID := parseUUID(vars["id"])
	if shareID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid share ID"))
		return
	}

	share, err := h.repo.GetDynamicTargetShare(r.Context(), *shareID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load share")
		apierror.WriteError(w, apierror.NewInternal("Failed to load share"))
		return
	}
	if share == nil {
		apierror.WriteError(w, apierror.NewNotFound("Share not found"))
		return
	}

	if share.SourceTenantID != claims.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("You do not have permission to revoke this share"))
		return
	}

	if share.RevokedAt != nil {
		apierror.WriteError(w, apierror.NewConflict("Share already revoked"))
		return
	}

	target, err := h.repo.GetTarget(r.Context(), share.TargetID, claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to load target")
		apierror.WriteError(w, apierror.NewInternal("Failed to load target"))
		return
	}
	if target == nil {
		apierror.WriteError(w, apierror.NewNotFound("Target not found"))
		return
	}

	if !h.requireDynamicPerm(w, r, claims, "dynamic_credentials:create", target.Namespace) {
		return
	}

	if err := h.repo.RevokeDynamicTargetShare(r.Context(), *shareID, claims.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to revoke target share")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke share"))
		return
	}

	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionRevoke, true, "", vault.JSONMap{
			"operation": "target_share_revoke",
			"share_id":  shareID.String(),
			"target_id": share.TargetID.String(),
		})

	w.WriteHeader(http.StatusNoContent)
}

// dynamicTargetShareResponse converts a DynamicTargetShare to its API response form.
func dynamicTargetShareResponse(s *vault.DynamicTargetShare) DynamicTargetShareResponse {
	return DynamicTargetShareResponse{
		ID:                s.ID,
		TargetID:          s.TargetID,
		SourceTenantID:    s.SourceTenantID,
		GrantedToTenantID: s.GrantedToTenantID,
		GrantedByUser:     s.GrantedByUser,
		Permissions:       s.Permissions,
		ExpiresAt:         s.ExpiresAt,
		RevokedAt:         s.RevokedAt,
		CreatedAt:         s.CreatedAt,
	}
}

// Ensure these are used
var _ = strings.TrimSpace
var _ = time.Time{}
