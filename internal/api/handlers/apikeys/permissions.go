package apikeys

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AddPermissionRequest represents a request to add a permission
type AddPermissionRequest struct {
	Permission   apikey.Permission   `json:"permission"`
	ResourceType apikey.ResourceType `json:"resource_type"`
	ResourceID   uuid.UUID           `json:"resource_id"`
}

// HandleListPermissions handles GET /api/v1/api-keys/:id/permissions
func (h *Handler) HandleListPermissions(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Get API key ID from path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid API key ID format")
		return
	}

	// Verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Fetch permissions
	permissions, err := h.repo.GetPermissions(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch permissions")
		return
	}

	h.writeSuccess(w, permissions)
}

// HandleAddPermission handles POST /api/v1/api-keys/:id/permissions
func (h *Handler) HandleAddPermission(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Get API key ID from path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid API key ID format")
		return
	}

	// Parse request body
	var req AddPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate request
	if req.Permission == "" || req.ResourceType == "" {
		h.writeError(w, http.StatusBadRequest, "validation_error", "permission and resource_type are required")
		return
	}

	// Verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Add permission
	perm := &apikey.PermissionGrant{
		Permission:   req.Permission,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
	}
	err = h.repo.AddPermission(ctx, id, perm)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to add permission")
		return
	}

	// Fetch updated permissions
	permissions, err := h.repo.GetPermissions(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch permissions")
		return
	}

	h.writeSuccess(w, permissions)
}

// HandleRemovePermission handles DELETE /api/v1/api-keys/:id/permissions/:perm_id
func (h *Handler) HandleRemovePermission(w http.ResponseWriter, r *http.Request) {
	// Get user claims
	claims, ok := getUserClaims(r)
	if !ok {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Get API key ID and permission ID from path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "API key ID is required")
		return
	}

	permIDStr, ok := vars["perm_id"]
	if !ok {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Permission ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid API key ID format")
		return
	}

	permID, err := uuid.Parse(permIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid permission ID format")
		return
	}

	// Verify ownership
	ctx := context.Background()
	apiKey, err := h.repo.GetByID(ctx, id)
	if err != nil || apiKey.TenantID != claims.TenantID {
		h.writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Get the permission to find its details
	permissions, err := h.repo.GetPermissions(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch permissions")
		return
	}

	// Find the permission
	var permToRemove *apikey.APIKeyPermission
	for _, p := range permissions {
		if p.ID == permID {
			permToRemove = p
			break
		}
	}

	if permToRemove == nil {
		h.writeError(w, http.StatusNotFound, "not_found", "Permission not found")
		return
	}

	// Remove permission
	err = h.repo.RemovePermission(ctx, id, permToRemove.Permission, permToRemove.ResourceType, permToRemove.ResourceID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to remove permission")
		return
	}

	// Fetch updated permissions
	permissions, err = h.repo.GetPermissions(ctx, id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch permissions")
		return
	}

	h.writeSuccess(w, permissions)
}

// RegisterPermissionRoutes registers the permission routes
func RegisterPermissionRoutes(router *mux.Router, repo *apikey.Repository) {
	h := NewHandler(repo, nil)
	router.HandleFunc("/api-keys/{id}/permissions", h.HandleListPermissions).Methods("GET", "OPTIONS")
	router.HandleFunc("/api-keys/{id}/permissions", h.HandleAddPermission).Methods("POST", "OPTIONS")
	router.HandleFunc("/api-keys/{id}/permissions/{perm_id}", h.HandleRemovePermission).Methods("DELETE", "OPTIONS")
}
