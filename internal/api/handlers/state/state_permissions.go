package state

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleGrantPermission handles POST /v1/state/{path}/permissions
func (h *Handler) HandleGrantPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	var req GrantPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("state not found"))
		return
	}

	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_admin") {
		return
	}

	perm := &staterepo.StatePermission{
		StateID:       state.ID,
		PrincipalType: req.PrincipalType,
		PrincipalID:   &req.PrincipalID,
		CanRead:       req.CanRead,
		CanWrite:      req.CanWrite,
		CanDelete:     req.CanDelete,
		CanAdmin:      req.CanAdmin,
		CanTrigger:    req.CanTrigger,
	}

	created, err := h.stateRepo.GrantPermission(r.Context(), perm)
	if err != nil {
		logrus.Errorf("failed to grant permission: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to grant permission"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

// HandleGetPermissions handles GET /v1/state/{path}/permissions
func (h *Handler) HandleGetPermissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("state not found"))
		return
	}

	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_admin") {
		return
	}

	permissions, err := h.stateRepo.GetPermissions(r.Context(), state.ID)
	if err != nil {
		logrus.Errorf("failed to get permissions: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to get permissions"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}
