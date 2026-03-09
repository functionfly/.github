package state

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
)

// HandleGrantPermission handles POST /v1/state/{path}/permissions
func (h *Handler) HandleGrantPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	var req GrantPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		http.Error(w, "state not found", http.StatusNotFound)
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
		http.Error(w, "failed to grant permission", http.StatusInternalServerError)
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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		http.Error(w, "state not found", http.StatusNotFound)
		return
	}

	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_admin") {
		return
	}

	permissions, err := h.stateRepo.GetPermissions(r.Context(), state.ID)
	if err != nil {
		logrus.Errorf("failed to get permissions: %v", err)
		http.Error(w, "failed to get permissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}
