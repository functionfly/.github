package state

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/api/middleware"
)

// HandleSetValue handles PUT /v1/state/{path}
func (h *Handler) HandleSetValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	key := r.URL.Query().Get("key")

	if key == "" {
		key = "" // Default to root key
	}

	var req SetValueRequest
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

	// Check write permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_write") {
		return
	}

	value, err := h.stateRepo.SetStateValue(r.Context(), state.ID, key, req.Value, "user", claims.UserID.String())
	if err != nil {
		logrus.Errorf("failed to set value: %v", err)
		http.Error(w, "failed to set value", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

// HandleGetValue handles GET /v1/state/{path}/value
func (h *Handler) HandleGetValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	key := r.URL.Query().Get("key")

	if key == "" {
		key = ""
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

	// Check read permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_read") {
		return
	}

	value, err := h.stateRepo.GetStateValue(r.Context(), state.ID, key)
	if err != nil {
		http.Error(w, "value not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}

// HandleDeleteValue handles DELETE /v1/state/{path}/value
func (h *Handler) HandleDeleteValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	key := r.URL.Query().Get("key")

	if key == "" {
		key = ""
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

	// Check delete permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_delete") {
		return
	}

	err = h.stateRepo.DeleteStateValue(r.Context(), state.ID, key, "user", claims.UserID.String())
	if err != nil {
		logrus.Errorf("failed to delete value: %v", err)
		http.Error(w, "failed to delete value", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandlePatchValue handles PATCH /v1/state/{path}/value - JSON Patch (RFC 6902)
func (h *Handler) HandlePatchValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	key := r.URL.Query().Get("key")

	if key == "" {
		key = ""
	}

	var req PatchValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Patch) == 0 {
		http.Error(w, "patch operations required", http.StatusBadRequest)
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

	// Check write permission
	if !h.requirePermission(w, r, state.ID, claims.UserID, "can_write") {
		return
	}

	// Get current value for patching
	currentValue, err := h.stateRepo.GetStateValue(r.Context(), state.ID, key)
	var currentMap map[string]interface{}
	if err == nil {
		currentMap = map[string]interface{}(currentValue.Value)
	} else {
		currentMap = make(map[string]interface{})
	}

	// Apply patch operations
	previousValue := currentMap
	for _, op := range req.Patch {
		opType, _ := op["op"].(string)
		patchPath, _ := op["path"].(string)

		switch opType {
		case "add", "replace":
			value := op["value"]
			if patchPath == "" || patchPath == "/" {
				// Replace entire value
				if valMap, ok := value.(map[string]interface{}); ok {
					currentMap = valMap
				}
			} else {
				// Handle nested paths like "/field/subfield"
				nestedPath := patchPath[1:] // Remove leading /
				setNestedValue(currentMap, nestedPath, value)
			}
		case "remove":
			if patchPath == "" || patchPath == "/" {
				currentMap = make(map[string]interface{})
			} else {
				nestedPath := patchPath[1:]
				deleteNestedValue(currentMap, nestedPath)
			}
		case "test":
			// Test operation - verify value matches
			value := op["value"]
			if !testNestedValue(currentMap, patchPath, value) {
				http.Error(w, "test operation failed: value mismatch", http.StatusBadRequest)
				return
			}
		}
	}

	// Set the patched value
	value, err := h.stateRepo.SetStateValue(r.Context(), state.ID, key, currentMap, "user", claims.UserID.String())
	if err != nil {
		logrus.Errorf("failed to patch value: %v", err)
		http.Error(w, "failed to patch value", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"value":       value,
		"previous":    previousValue,
		"applied_ops": len(req.Patch),
	})
}