package vault

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage/vault"
	"github.com/gorilla/mux"
)

// ============================================================================
// Phase 4.3: Namespaces
// ============================================================================

func (h *Handler) HandleCreateNamespace(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	plan := middleware.GetTenantPlan(r)
	if !plans.SupportsVaultNamespaces(plan) {
		apierror.WriteError(w, apierror.NewForbidden("Namespaces require Professional plan or higher"))
		return
	}
	var req CreateNamespaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	req.Path = strings.TrimSpace(req.Path)
	if !vault.IsValidNamespacePath(req.Path) {
		apierror.WriteError(w, apierror.NewBadRequest("path must match [a-z0-9/_-]+ with no empty segments"))
		return
	}
	segments := vault.SplitNamespacePath(req.Path)
	if len(segments) > 5 {
		apierror.WriteError(w, apierror.NewBadRequest("namespace path exceeds maximum depth of 5 segments"))
		return
	}
	switch req.Path {
	case "default", "shared", "system":
		apierror.WriteError(w, apierror.NewBadRequest("path is reserved and cannot be used"))
		return
	}
	n := &vault.VaultNamespace{
		TenantID:    claims.TenantID,
		Path:        req.Path,
		Description: req.Description,
		CreatedBy:   claims.UserID,
	}
	if pid := parseUUID(req.ParentID); pid != nil {
		parent, err := h.repo.GetNamespace(r.Context(), *pid, claims.TenantID)
		if err != nil || parent == nil {
			apierror.WriteError(w, apierror.NewBadRequest("parent namespace not found"))
			return
		}
		expectedPrefix := parent.Path + "/"
		if !strings.HasPrefix(req.Path, expectedPrefix) {
			apierror.WriteError(w, apierror.NewBadRequest("path must be a child of the specified parent namespace"))
			return
		}
		n.ParentID = pid
	}
	if err := h.repo.CreateNamespace(r.Context(), n); err != nil {
		apierror.WriteError(w, apierror.NewConflict("namespace already exists or invalid parent"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation": "namespace_create",
			"path":      n.Path,
		})
	h.respondJSON(w, http.StatusCreated, namespaceResponse(n))
}

func (h *Handler) HandleListNamespaces(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	plan := middleware.GetTenantPlan(r)
	if !plans.SupportsVaultNamespaces(plan) {
		apierror.WriteError(w, apierror.NewForbidden("Namespaces require Professional plan or higher"))
		return
	}
	limit := parseLimit(r, 100)
	offset := parseOffset(r)
	ns, err := h.repo.ListNamespaces(r.Context(), claims.TenantID, limit, offset)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list namespaces"))
		return
	}
	resp := make([]NamespaceResponse, len(ns))
	for i := range ns {
		resp[i] = namespaceResponse(&ns[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"namespaces": resp, "total": len(resp)})
}

func (h *Handler) HandleDeleteNamespace(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	plan := middleware.GetTenantPlan(r)
	if !plans.SupportsVaultNamespaces(plan) {
		apierror.WriteError(w, apierror.NewForbidden("Namespaces require Professional plan or higher"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid namespace ID"))
		return
	}
	if err := h.repo.DeleteNamespace(r.Context(), *id, claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Namespace not found"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionDelete, true, "", vault.JSONMap{"operation": "namespace_delete"})
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Phase 4.1: RBAC
// ============================================================================

func (h *Handler) HandleListRoles(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	// Lazy seed of built-in roles.
	engine := h.RBAC
	if engine == nil {
		engine = vault.NewRBACEngine(h.repo)
	}
	if err := engine.EnsureBuiltinRoles(r.Context(), claims.TenantID); err != nil {
		h.logger.WithError(err).Warn("Failed to seed built-in roles")
	}
	roles, err := h.repo.ListRoles(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list roles"))
		return
	}
	resp := make([]RoleResponse, len(roles))
	for i := range roles {
		resp[i] = roleResponse(&roles[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"roles": resp, "total": len(resp)})
}

func (h *Handler) HandleCreateRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name is required"))
		return
	}
	role := &vault.VaultRole{
		TenantID:    claims.TenantID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Permissions: vault.JSONMap(req.Permissions),
		IsBuiltin:   false,
		CreatedBy:   &claims.UserID,
	}
	if err := h.repo.CreateRole(r.Context(), role); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create role"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation": "role_create",
			"role_name": role.Name,
		})
	h.respondJSON(w, http.StatusCreated, roleResponse(role))
}

func (h *Handler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid role ID"))
		return
	}
	role, err := h.repo.GetRole(r.Context(), *id, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load role"))
		return
	}
	if role == nil {
		apierror.WriteError(w, apierror.NewNotFound("Role not found"))
		return
	}
	if role.IsBuiltin {
		apierror.WriteError(w, apierror.NewForbidden("Built-in roles cannot be modified"))
		return
	}
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Permissions != nil {
		role.Permissions = vault.JSONMap(req.Permissions)
	}
	if err := h.repo.UpdateRole(r.Context(), role); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update role"))
		return
	}
	h.respondJSON(w, http.StatusOK, roleResponse(role))
}

func (h *Handler) HandleDeleteRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid role ID"))
		return
	}
	role, err := h.repo.GetRole(r.Context(), *id, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load role"))
		return
	}
	if role != nil && role.IsBuiltin {
		apierror.WriteError(w, apierror.NewForbidden("Built-in roles cannot be deleted"))
		return
	}
	if err := h.repo.DeleteRole(r.Context(), *id, claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Role not found"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionDelete, true, "", vault.JSONMap{"operation": "role_delete"})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleAssignRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid role ID"))
		return
	}
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	userID := parseUUID(req.UserID)
	if userID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("user_id is required"))
		return
	}
	a := &vault.VaultRoleAssignment{
		TenantID:  claims.TenantID,
		RoleID:    *id,
		UserID:    userID,
		Scope:     req.Scope,
		CreatedBy: claims.UserID,
	}
	if a.Scope == "" {
		a.Scope = "all"
	}
	if err := h.repo.AssignRole(r.Context(), a); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to assign role"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation": "role_assign",
			"role_id":   id.String(),
			"user_id":   userID.String(),
			"scope":     a.Scope,
		})
	h.respondJSON(w, http.StatusCreated, assignmentResponse(a))
}

func (h *Handler) HandleUnassignRole(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["assignment_id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid assignment ID"))
		return
	}
	if err := h.repo.DeleteAssignment(r.Context(), *id, claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Assignment not found"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionDelete, true, "", vault.JSONMap{"operation": "role_unassign"})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleListMyAssignments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	as, err := h.repo.ListAssignmentsForUser(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list assignments"))
		return
	}
	resp := make([]AssignmentResponse, len(as))
	for i := range as {
		resp[i] = assignmentResponse(&as[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"assignments": resp, "total": len(resp)})
}

// ============================================================================
// Phase 4.4: Secret sharing
// ============================================================================

func (h *Handler) HandleShareSecret(w http.ResponseWriter, r *http.Request) {
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
	var req ShareSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	granteeID := parseUUID(req.GranteeTenantID)
	if granteeID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("grantee_tenant_id is required"))
		return
	}
	if *granteeID == claims.TenantID {
		apierror.WriteError(w, apierror.NewBadRequest("cannot share with the same tenant"))
		return
	}
	// Verify the secret belongs to the source tenant.
	secret, err := h.repo.GetSecretByID(r.Context(), *secretID, claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load secret"))
		return
	}
	if secret == nil {
		apierror.WriteError(w, apierror.NewNotFound("Secret not found"))
		return
	}
	perms := req.Permissions
	if perms == "" {
		perms = "read"
	}
	if perms != "read" && perms != "read-write" {
		apierror.WriteError(w, apierror.NewBadRequest("permissions must be 'read' or 'read-write'"))
		return
	}
	share := &vault.VaultShare{
		SecretID:          *secretID,
		SourceTenantID:    claims.TenantID,
		GrantedToTenantID: *granteeID,
		GrantedByUser:     claims.UserID,
		Permissions:       perms,
		ExpiresAt:         req.ExpiresAt,
	}
	if err := h.repo.CreateShare(r.Context(), share); err != nil {
		apierror.WriteError(w, apierror.NewConflict("Share already exists or invalid"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, secretID, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{
			"operation":   "secret_share",
			"grantee_id":  granteeID.String(),
			"permissions": perms,
		})
	h.respondJSON(w, http.StatusCreated, shareResponse(share))
}

func (h *Handler) HandleListSharedWithMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	shares, err := h.repo.ListSharesForGrantee(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list shares"))
		return
	}
	resp := make([]ShareResponse, len(shares))
	for i := range shares {
		resp[i] = shareResponse(&shares[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"shares": resp, "total": len(resp)})
}

func (h *Handler) HandleRevokeShare(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	shareID := parseUUID(vars["share_id"])
	if shareID == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid share ID"))
		return
	}
	if err := h.repo.RevokeShare(r.Context(), *shareID, claims.UserID); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Share not found"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionRevoke, true, "", vault.JSONMap{"operation": "share_revoke"})
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Phase 4.5: SSO
// ============================================================================

func (h *Handler) HandleGetSSOConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	cfg, err := h.repo.GetSSOConfig(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load SSO config"))
		return
	}
	h.respondJSON(w, http.StatusOK, ssoConfigResponse(cfg))
}

func (h *Handler) HandleUpdateSSOConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req UpdateSSORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	cfg, err := h.repo.GetSSOConfig(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to load SSO config"))
		return
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.SAMLMetadataURL != "" {
		cfg.SAMLMetadataURL = req.SAMLMetadataURL
	}
	if req.SAMLEntityID != "" {
		cfg.SAMLEntityID = req.SAMLEntityID
	}
	if req.SAMLSSOURL != "" {
		cfg.SAMLSSOURL = req.SAMLSSOURL
	}
	if req.SAMLSLOURL != "" {
		cfg.SAMLSLOURL = req.SAMLSLOURL
	}
	if req.SAMLX509Cert != "" {
		cfg.SAMLX509Cert = req.SAMLX509Cert
	}
	if req.JITProvisioningEnabled != nil {
		cfg.JITProvisioningEnabled = *req.JITProvisioningEnabled
	}
	if req.AttributeRoleMapping != nil {
		cfg.AttributeRoleMapping = vault.JSONMap(req.AttributeRoleMapping)
	}
	if err := h.repo.UpdateSSOConfig(r.Context(), cfg); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update SSO config"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionUpdate, true, "", vault.JSONMap{"operation": "sso_config_update"})
	h.respondJSON(w, http.StatusOK, ssoConfigResponse(cfg))
}

// ============================================================================
// Phase 4.2: SIEM webhooks
// ============================================================================

func (h *Handler) HandleCreateSIEMWebhook(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	var req CreateSIEMWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.URL) == "" {
		apierror.WriteError(w, apierror.NewBadRequest("name and url are required"))
		return
	}
	if req.Format != "" && req.Format != "json" && req.Format != "cef" {
		apierror.WriteError(w, apierror.NewBadRequest("format must be 'json' or 'cef'"))
		return
	}
	if req.Format == "" {
		req.Format = "json"
	}
	hmacKey, err := vault.GenerateSalt(32)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to generate webhook secret"))
		return
	}
	hook := &vault.VaultSIEMWebhook{
		TenantID:   claims.TenantID,
		Name:       strings.TrimSpace(req.Name),
		URL:        strings.TrimSpace(req.URL),
		SecretHMAC: hmacKey,
		Format:     req.Format,
		Enabled:    true,
		CreatedBy:  claims.UserID,
	}
	if err := h.repo.CreateSIEMWebhook(r.Context(), hook); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to create webhook"))
		return
	}
	h.logAudit(r.Context(), claims.TenantID, nil, claims.UserID.String(),
		vault.AuditActionCreate, true, "", vault.JSONMap{"operation": "siem_webhook_create"})
	resp := siemWebhookResponse(hook)
	resp.SecretHMAC = base64Bytes(hmacKey)
	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *Handler) HandleListSIEMWebhooks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	hooks, err := h.repo.ListSIEMWebhooks(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to list webhooks"))
		return
	}
	resp := make([]SIEMWebhookResponse, len(hooks))
	for i := range hooks {
		resp[i] = siemWebhookResponse(&hooks[i])
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"webhooks": resp, "total": len(resp)})
}

func (h *Handler) HandleDeleteSIEMWebhook(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}
	vars := mux.Vars(r)
	id := parseUUID(vars["id"])
	if id == nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid webhook ID"))
		return
	}
	if err := h.repo.DeleteSIEMWebhook(r.Context(), *id, claims.TenantID); err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Webhook not found"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// helpers / response projections
// ============================================================================

func namespaceResponse(n *vault.VaultNamespace) NamespaceResponse {
	return NamespaceResponse{
		ID:          n.ID,
		TenantID:    n.TenantID,
		Path:        n.Path,
		Description: n.Description,
		ParentID:    n.ParentID,
		CreatedBy:   n.CreatedBy,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

func roleResponse(r *vault.VaultRole) RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		Description: r.Description,
		Permissions: r.Permissions,
		IsBuiltin:   r.IsBuiltin,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func assignmentResponse(a *vault.VaultRoleAssignment) AssignmentResponse {
	return AssignmentResponse{
		ID:        a.ID,
		TenantID:  a.TenantID,
		RoleID:    a.RoleID,
		UserID:    a.UserID,
		Scope:     a.Scope,
		CreatedBy: a.CreatedBy,
		CreatedAt: a.CreatedAt,
	}
}

func shareResponse(s *vault.VaultShare) ShareResponse {
	return ShareResponse{
		ID:                s.ID,
		SecretID:          s.SecretID,
		SourceTenantID:    s.SourceTenantID,
		GrantedToTenantID: s.GrantedToTenantID,
		GrantedByUser:     s.GrantedByUser,
		Permissions:       s.Permissions,
		ExpiresAt:         s.ExpiresAt,
		RevokedAt:         s.RevokedAt,
		CreatedAt:         s.CreatedAt,
	}
}

func ssoConfigResponse(c *vault.VaultSSOConfig) SSOConfigResponse {
	return SSOConfigResponse{
		TenantID:               c.TenantID,
		Enabled:                c.Enabled,
		SAMLMetadataURL:        c.SAMLMetadataURL,
		SAMLEntityID:           c.SAMLEntityID,
		SAMLSSOURL:             c.SAMLSSOURL,
		SAMLSLOURL:             c.SAMLSLOURL,
		JITProvisioningEnabled: c.JITProvisioningEnabled,
		AttributeRoleMapping:   c.AttributeRoleMapping,
		UpdatedAt:              c.UpdatedAt,
	}
}

func siemWebhookResponse(w *vault.VaultSIEMWebhook) SIEMWebhookResponse {
	return SIEMWebhookResponse{
		ID:                 w.ID,
		TenantID:           w.TenantID,
		Name:               w.Name,
		URL:                w.URL,
		Format:             w.Format,
		Enabled:            w.Enabled,
		LastDeliveryAt:     w.LastDeliveryAt,
		LastDeliveryStatus: w.LastDeliveryStatus,
		LastDeliveryError:  w.LastDeliveryError,
		CreatedAt:          w.CreatedAt,
	}
}

func base64Bytes(b []byte) string {
	return encodingBase64(b)
}

// encodingBase64 is a tiny indirection so the helper stays readable
// and we don't have to import encoding/base64 in the handler file.
func encodingBase64(b []byte) string {
	const tab = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	out := make([]byte, 0, ((len(b)+2)/3)*4)
	for i := 0; i < len(b); i += 3 {
		var n uint32
		switch len(b) - i {
		case 1:
			n = uint32(b[i]) << 16
			out = append(out, tab[(n>>18)&0x3F], tab[(n>>12)&0x3F], '=', '=')
		case 2:
			n = uint32(b[i])<<16 | uint32(b[i+1])<<8
			out = append(out, tab[(n>>18)&0x3F], tab[(n>>12)&0x3F], tab[(n>>6)&0x3F], '=')
		default:
			n = uint32(b[i])<<16 | uint32(b[i+1])<<8 | uint32(b[i+2])
			out = append(out, tab[(n>>18)&0x3F], tab[(n>>12)&0x3F], tab[(n>>6)&0x3F], tab[n&0x3F])
		}
	}
	return string(out)
}

// Time-tracked stub so the import isn't dropped on rebuilds.
var _ = time.Now
