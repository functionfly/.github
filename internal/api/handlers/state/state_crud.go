package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/monitoring"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/functionfly/functionfly/internal/apierror"
)

// HandleCreateState handles POST /v1/state
func (h *Handler) HandleCreateState(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		monitoring.RecordStateOperation("", "create", "unauthorized")
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	var req CreateStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		monitoring.RecordStateOperation(claims.TenantID.String(), "create", "bad_request")
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	// Get tenant ID from user context
	tenantID := claims.TenantID

	// Security by default: encryption is enabled unless explicitly opted out
	// This ensures state values are encrypted at rest
	isEncrypted := true
	if req.IsEncryptedSet {
		isEncrypted = req.IsEncrypted
	}

	state := &staterepo.State{
		TenantID:    tenantID,
		Name:        req.Name,
		FullPath:    fmt.Sprintf("%s/%s", tenantID.String()[:8], req.Name),
		StorageType: req.StorageType,
		TTLDays:     req.TTLDays,
		MaxSizeMB:   req.MaxSizeMB,
		IsVersioned: req.IsVersioned,
		IsEncrypted: isEncrypted,
		IsPublic:    req.IsPublic,
		Description: strPtr(req.Description),
		Tags:        req.Tags,
	}

	created, err := h.stateRepo.CreateState(r.Context(), state)
	if err != nil {
		logrus.Errorf("failed to create state: %v", err)
		monitoring.RecordStateOperation(tenantID.String(), "create", "error")
		apierror.WriteError(w, apierror.NewInternal("failed to create state"))
		return
	}

	monitoring.RecordStateOperation(tenantID.String(), "create", "success")
	monitoring.RecordStateOperationDuration(tenantID.String(), "create", time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

// HandleGetState handles GET /v1/state/{path}
func (h *Handler) HandleGetState(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	path := vars["path"]

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		monitoring.RecordStateOperation("", "read", "unauthorized")
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		logrus.Errorf("failed to get state: %v", err)
		monitoring.RecordStateOperation(tenantID.String(), "read", "not_found")
		apierror.WriteError(w, apierror.NewNotFound("state not found"))
		return
	}

	// Check read permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_read") {
		monitoring.RecordStateOperation(tenantID.String(), "read", "forbidden")
		return
	}

	monitoring.RecordStateOperation(tenantID.String(), "read", "success")
	monitoring.RecordStateOperationDuration(tenantID.String(), "read", time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

// HandleListStates handles GET /v1/state
func (h *Handler) HandleListStates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset == 0 {
		offset = 0
	}

	states, total, err := h.stateRepo.ListStatesByTenant(r.Context(), tenantID, limit, offset)
	if err != nil {
		logrus.Errorf("failed to list states: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to list states"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"states": states,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleDeleteState handles DELETE /v1/state/{path}
func (h *Handler) HandleDeleteState(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	path := vars["path"]

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		monitoring.RecordStateOperation("", "delete", "unauthorized")
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		monitoring.RecordStateOperation(tenantID.String(), "delete", "not_found")
		apierror.WriteError(w, apierror.NewNotFound("state not found"))
		return
	}

	// Check delete permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_delete") {
		monitoring.RecordStateOperation(tenantID.String(), "delete", "forbidden")
		return
	}

	err = h.stateRepo.DeleteState(r.Context(), state.ID)
	if err != nil {
		logrus.Errorf("failed to delete state: %v", err)
		monitoring.RecordStateOperation(tenantID.String(), "delete", "error")
		apierror.WriteError(w, apierror.NewInternal("failed to delete state"))
		return
	}

	monitoring.RecordStateOperation(tenantID.String(), "delete", "success")
	monitoring.RecordStateOperationDuration(tenantID.String(), "delete", time.Since(start))

	w.WriteHeader(http.StatusNoContent)
}

// HandleUpdateState handles PUT /v1/state/{path}
func (h *Handler) HandleUpdateState(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	path := vars["path"]

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		monitoring.RecordStateOperation("", "update", "unauthorized")
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	var req UpdateStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		monitoring.RecordStateOperation(claims.TenantID.String(), "update", "bad_request")
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	tenantID := claims.TenantID

	state, err := h.stateRepo.GetStateByPath(r.Context(), tenantID, path)
	if err != nil {
		monitoring.RecordStateOperation(tenantID.String(), "update", "not_found")
		apierror.WriteError(w, apierror.NewNotFound("state not found"))
		return
	}

	// Check admin permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_admin") {
		monitoring.RecordStateOperation(tenantID.String(), "update", "forbidden")
		return
	}

	// Update only provided fields
	if req.Name != "" {
		state.Name = req.Name
	}
	if req.Description != "" {
		state.Description = &req.Description
	}
	if req.Tags != nil {
		state.Tags = req.Tags
	}
	if req.TTLDays != nil {
		state.TTLDays = *req.TTLDays
	}
	if req.MaxSizeMB != nil {
		state.MaxSizeMB = *req.MaxSizeMB
	}
	if req.IsPublic != nil {
		state.IsPublic = *req.IsPublic
	}
	if req.IsEncrypted != nil {
		state.IsEncrypted = *req.IsEncrypted
	}

	updated, err := h.stateRepo.UpdateState(r.Context(), state)
	if err != nil {
		logrus.Errorf("failed to update state: %v", err)
		monitoring.RecordStateOperation(tenantID.String(), "update", "error")
		apierror.WriteError(w, apierror.NewInternal("failed to update state"))
		return
	}

	monitoring.RecordStateOperation(tenantID.String(), "update", "success")
	monitoring.RecordStateOperationDuration(tenantID.String(), "update", time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
