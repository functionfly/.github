package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SCIMHandler handles SCIM 2.0 endpoints
type SCIMHandler struct {
	scimSvc *auth.SCIMService
}

// NewSCIMHandler creates a new SCIM handler
func NewSCIMHandler(scimSvc *auth.SCIMService) *SCIMHandler {
	return &SCIMHandler{
		scimSvc: scimSvc,
	}
}

// getStartIndex extracts startIndex from query params (SCIM uses 1-based indexing)
func getStartIndex(r *http.Request) int {
	startIndex := r.URL.Query().Get("startIndex")
	if startIndex == "" {
		return 1
	}
	index, err := strconv.Atoi(startIndex)
	if err != nil || index < 1 {
		return 1
	}
	return index
}

// getCount extracts count from query params
func getCount(r *http.Request) int {
	count := r.URL.Query().Get("count")
	if count == "" {
		return 100
	}
	c, err := strconv.Atoi(count)
	if err != nil || c < 1 {
		return 100
	}
	if c > 1000 {
		return 1000
	}
	return c
}

// writeSCIMError writes a SCIM error response
func writeSCIMError(w http.ResponseWriter, status int, detail string) {
	scimErr := auth.NewSCIMError(status, detail)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(scimErr)
}

// HandleListUsers handles GET /v1/scim/Users - List users
func (h *SCIMHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startIndex := getStartIndex(r)
	count := getCount(r)

	result, err := h.scimSvc.ListUsers(r.Context(), tenantID, startIndex, count)
	if err != nil {
		logrus.WithError(err).Error("Failed to list SCIM users")
		writeSCIMError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleCreateUser handles POST /v1/scim/Users - Create user
func (h *SCIMHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var scimUser auth.SCIMUser
	if err := json.Unmarshal(body, &scimUser); err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid SCIM user payload")
		return
	}

	result, err := h.scimSvc.CreateUser(r.Context(), tenantID, &scimUser)
	if err != nil {
		logrus.WithError(err).Error("Failed to create SCIM user")
		if scimErr, ok := err.(*auth.SCIMError); ok {
			writeSCIMError(w, 400, scimErr.Detail)
			return
		}
		writeSCIMError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// HandleGetUser handles GET /v1/scim/Users/{id} - Get user
func (h *SCIMHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	result, err := h.scimSvc.GetUser(r.Context(), tenantID, userID)
	if err != nil {
		writeSCIMError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleUpdateUser handles PUT /v1/scim/Users/{id} - Replace user
func (h *SCIMHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var scimUser auth.SCIMUser
	if err := json.Unmarshal(body, &scimUser); err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid SCIM user payload")
		return
	}

	result, err := h.scimSvc.UpdateUser(r.Context(), tenantID, userID, &scimUser)
	if err != nil {
		logrus.WithError(err).Error("Failed to update SCIM user")
		if scimErr, ok := err.(*auth.SCIMError); ok {
			writeSCIMError(w, 400, scimErr.Detail)
			return
		}
		writeSCIMError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandlePatchUser handles PATCH /v1/scim/Users/{id} - Update user
func (h *SCIMHandler) HandlePatchUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse SCIM patch operations
	var operations []auth.SCIMPatchOperation
	if err := json.Unmarshal(body, &operations); err != nil {
		// Try parsing as single operation
		var op auth.SCIMPatchOperation
		if err := json.Unmarshal(body, &op); err != nil {
			writeSCIMError(w, http.StatusBadRequest, "Invalid SCIM patch payload")
			return
		}
		operations = []auth.SCIMPatchOperation{op}
	}

	result, err := h.scimSvc.PatchUser(r.Context(), tenantID, userID, operations)
	if err != nil {
		logrus.WithError(err).Error("Failed to patch SCIM user")
		if scimErr, ok := err.(*auth.SCIMError); ok {
			writeSCIMError(w, 400, scimErr.Detail)
			return
		}
		writeSCIMError(w, http.StatusInternalServerError, "Failed to patch user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleDeleteUser handles DELETE /v1/scim/Users/{id} - Delete user
func (h *SCIMHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.scimSvc.DeleteUser(r.Context(), tenantID, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete SCIM user")
		writeSCIMError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleListGroups handles GET /v1/scim/Groups - List groups
func (h *SCIMHandler) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startIndex := getStartIndex(r)
	count := getCount(r)

	result, err := h.scimSvc.ListGroups(r.Context(), tenantID, startIndex, count)
	if err != nil {
		logrus.WithError(err).Error("Failed to list SCIM groups")
		writeSCIMError(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleCreateGroup handles POST /v1/scim/Groups - Create group
func (h *SCIMHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var scimGroup auth.SCIMGroup
	if err := json.Unmarshal(body, &scimGroup); err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid SCIM group payload")
		return
	}

	result, err := h.scimSvc.CreateGroup(r.Context(), tenantID, &scimGroup)
	if err != nil {
		logrus.WithError(err).Error("Failed to create SCIM group")
		if scimErr, ok := err.(*auth.SCIMError); ok {
			writeSCIMError(w, 400, scimErr.Detail)
			return
		}
		writeSCIMError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// HandleGetGroup handles GET /v1/scim/Groups/{id} - Get group
func (h *SCIMHandler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	result, err := h.scimSvc.GetGroup(r.Context(), tenantID, groupID)
	if err != nil {
		writeSCIMError(w, http.StatusNotFound, "Group not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleUpdateGroup handles PATCH /v1/scim/Groups/{id} - Update group
func (h *SCIMHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse SCIM patch operations
	var operations []auth.SCIMPatchOperation
	if err := json.Unmarshal(body, &operations); err != nil {
		// Try parsing as single operation
		var op auth.SCIMPatchOperation
		if err := json.Unmarshal(body, &op); err != nil {
			writeSCIMError(w, http.StatusBadRequest, "Invalid SCIM patch payload")
			return
		}
		operations = []auth.SCIMPatchOperation{op}
	}

	result, err := h.scimSvc.PatchGroup(r.Context(), tenantID, groupID, operations)
	if err != nil {
		logrus.WithError(err).Error("Failed to patch SCIM group")
		if scimErr, ok := err.(*auth.SCIMError); ok {
			writeSCIMError(w, 400, scimErr.Detail)
			return
		}
		writeSCIMError(w, http.StatusInternalServerError, "Failed to patch group")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleDeleteGroup handles DELETE /v1/scim/Groups/{id} - Delete group
func (h *SCIMHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		writeSCIMError(w, http.StatusBadRequest, "Invalid group ID")
		return
	}

	err = h.scimSvc.DeleteGroup(r.Context(), tenantID, groupID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete SCIM group")
		writeSCIMError(w, http.StatusInternalServerError, "Failed to delete group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetConfig handles GET /v1/scim/Config - Get SCIM config
func (h *SCIMHandler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		writeSCIMError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	config, err := h.scimSvc.GetConfig(r.Context(), tenantID)
	if err != nil {
		// Return empty config if not found
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(config)
}

// getTenantID extracts tenant ID from request context
func getTenantID(r *http.Request) (uuid.UUID, error) {
	// Try to get from context (set by auth middleware)
	tenantIDVal := r.Context().Value("tenant_id")
	if tenantIDVal != nil {
		if idStr, ok := tenantIDVal.(string); ok {
			return uuid.Parse(idStr)
		}
	}

	// Fall back to X-Tenant-ID header (for SCIM clients)
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr != "" {
		return uuid.Parse(tenantIDStr)
	}

	return uuid.Nil, http.ErrNoCookie
}
