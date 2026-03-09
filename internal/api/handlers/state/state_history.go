package state

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
)

// HandleGetHistory handles GET /v1/state/{path}/history
func (h *Handler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	key := r.URL.Query().Get("key")

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

	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_read") {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset == 0 {
		offset = 0
	}

	events, total, err := h.stateRepo.GetStateHistory(r.Context(), state.ID, key, limit, offset)
	if err != nil {
		logrus.Errorf("failed to get history: %v", err)
		http.Error(w, "failed to get history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleCreateSnapshot handles POST /v1/state/{path}/snapshot
func (h *Handler) HandleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	var req CreateSnapshotRequest
	json.NewDecoder(r.Body).Decode(&req)

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

	snapshot, err := h.stateRepo.CreateSnapshot(r.Context(), state.ID, req.Label)
	if err != nil {
		logrus.Errorf("failed to create snapshot: %v", err)
		http.Error(w, "failed to create snapshot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// HandleListSnapshots handles GET /v1/state/{path}/snapshots
func (h *Handler) HandleListSnapshots(w http.ResponseWriter, r *http.Request) {
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

	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_read") {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset == 0 {
		offset = 0
	}

	snapshots, total, err := h.stateRepo.ListSnapshots(r.Context(), state.ID, limit, offset)
	if err != nil {
		logrus.Errorf("failed to list snapshots: %v", err)
		http.Error(w, "failed to list snapshots", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"snapshots": snapshots,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// HandleRestoreSnapshot handles POST /v1/state/{path}/restore
func (h *Handler) HandleRestoreSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	var req RestoreSnapshotRequest
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

	err = h.stateRepo.RestoreSnapshot(r.Context(), state.ID, req.SnapshotVersion, "user", claims.UserID.String())
	if err != nil {
		logrus.Errorf("failed to restore snapshot: %v", err)
		http.Error(w, "failed to restore snapshot", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "restored"})
}

// HandleTimeTravel handles GET /v1/state/{path}/time-travel
func (h *Handler) HandleTimeTravel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]

	timestampStr := r.URL.Query().Get("at")
	if timestampStr == "" {
		http.Error(w, "timestamp required", http.StatusBadRequest)
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		http.Error(w, "invalid timestamp format", http.StatusBadRequest)
		return
	}

	user := r.Context().Value("user")
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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

	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_read") {
		return
	}

	data, err := h.stateRepo.TimeTravelQuery(r.Context(), state.ID, timestamp)
	if err != nil {
		logrus.Errorf("failed to time travel: %v", err)
		http.Error(w, "failed to time travel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": timestamp,
		"data":      data,
	})
}
