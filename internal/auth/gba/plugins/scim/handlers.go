// Package scim provides SCIM 2.0 provisioning support for GoBetterAuth
package scim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler provides HTTP handlers for SCIM endpoints
type Handler struct {
	plugin *SCIMPlugin
	logger *logrus.Logger
}

// NewHandler creates a new SCIM handler
func NewHandler(plugin *SCIMPlugin, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		plugin: plugin,
		logger: logger,
	}
}

// SetupRoutes registers all SCIM routes with the provided mux
// Base path should be /v1/scim
func (h *Handler) SetupRoutes(mux *http.ServeMux, basePath string) {
	// User endpoints (RFC 7644)
	mux.HandleFunc("GET "+basePath+"/Users", h.HandleListUsers)
	mux.HandleFunc("POST "+basePath+"/Users", h.HandleCreateUser)
	mux.HandleFunc("GET "+basePath+"/Users/{id}", h.HandleGetUser)
	mux.HandleFunc("PUT "+basePath+"/Users/{id}", h.HandleUpdateUser)
	mux.HandleFunc("PATCH "+basePath+"/Users/{id}", h.HandlePatchUser)
	mux.HandleFunc("DELETE "+basePath+"/Users/{id}", h.HandleDeleteUser)

	// Group endpoints (RFC 7644)
	mux.HandleFunc("GET "+basePath+"/Groups", h.HandleListGroups)
	mux.HandleFunc("POST "+basePath+"/Groups", h.HandleCreateGroup)
	mux.HandleFunc("GET "+basePath+"/Groups/{id}", h.HandleGetGroup)
	mux.HandleFunc("PUT "+basePath+"/Groups/{id}", h.HandleUpdateGroup)
	mux.HandleFunc("PATCH "+basePath+"/Groups/{id}", h.HandlePatchGroup)
	mux.HandleFunc("DELETE "+basePath+"/Groups/{id}", h.HandleDeleteGroup)

	// Service Provider Config endpoint
	mux.HandleFunc("GET "+basePath+"/ServiceProviderConfig", h.HandleServiceProviderConfig)

	// Resource Types endpoint
	mux.HandleFunc("GET "+basePath+"/ResourceTypes", h.HandleResourceTypes)

	// Schemas endpoint
	mux.HandleFunc("GET "+basePath+"/Schemas", h.HandleSchemas)

	h.logger.WithField("path", basePath).Info("SCIM routes registered")
}

// Authentication middleware for SCIM
func (h *Handler) authenticate(r *http.Request) (uuid.UUID, error) {
	// Get tenant ID from header
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return uuid.Nil, fmt.Errorf("missing tenant ID")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid tenant ID")
	}

	// Get bearer token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return uuid.Nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return uuid.Nil, fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]

	// Verify token
	valid, err := h.plugin.service.VerifyToken(r.Context(), tenantID, token)
	if err != nil || !valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	return tenantID, nil
}

// HandleListUsers handles GET /v1/scim/Users
// RFC 7644 Section 3.4.2 - Query Resources
func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Parse query parameters
	filter := r.URL.Query().Get("filter")
	startIndexStr := r.URL.Query().Get("startIndex")
	countStr := r.URL.Query().Get("count")

	startIndex, count := ParsePagination(startIndexStr, countStr)

	// List users
	users, total, err := h.plugin.service.ListUsers(r.Context(), tenantID, filter, startIndex, count)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		h.respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	// Convert to response format
	resources := make([]map[string]interface{}, len(users))
	for i, user := range users {
		resources[i] = user.ToSCIMResponse(h.plugin.config.BaseURL)
	}

	response := NewSCIMListResponse(startIndex, count, total, resources)
	h.respondJSON(w, http.StatusOK, response)
}

// HandleCreateUser handles POST /v1/scim/Users
// RFC 7644 Section 3.3 - Creating Resources
func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req SCIMUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondSCIMError(w, http.StatusBadRequest, "invalidSyntax", "Invalid JSON")
		return
	}

	user, err := h.plugin.service.CreateUser(r.Context(), tenantID, &req)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with username %s already exists", req.UserName) {
			h.respondSCIMError(w, http.StatusConflict, "uniqueness", err.Error())
			return
		}
		h.respondSCIMError(w, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"tenant_id": tenantID,
	}).Info("SCIM user created")

	h.respondJSON(w, http.StatusCreated, user.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandleGetUser handles GET /v1/scim/Users/{id}
// RFC 7644 Section 3.4.1 - Retrieve a Known Resource
func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID := r.PathValue("id")

	user, err := h.plugin.service.GetUser(r.Context(), tenantID, userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		h.respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		h.respondSCIMError(w, http.StatusNotFound, "", "User not found")
		return
	}

	h.respondJSON(w, http.StatusOK, user.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandleUpdateUser handles PUT /v1/scim/Users/{id}
// RFC 7644 Section 3.5.1 - Update with PUT
func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID := r.PathValue("id")

	var req SCIMUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondSCIMError(w, http.StatusBadRequest, "invalidSyntax", "Invalid JSON")
		return
	}

	user, err := h.plugin.service.UpdateUser(r.Context(), tenantID, userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			h.respondSCIMError(w, http.StatusNotFound, "", "User not found")
			return
		}
		h.respondSCIMError(w, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Info("SCIM user updated")

	h.respondJSON(w, http.StatusOK, user.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandlePatchUser handles PATCH /v1/scim/Users/{id}
// RFC 7644 Section 3.5.2 - Update with PATCH
func (h *Handler) HandlePatchUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID := r.PathValue("id")

	var req SCIMPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondSCIMError(w, http.StatusBadRequest, "invalidSyntax", "Invalid JSON")
		return
	}

	user, err := h.plugin.service.PatchUser(r.Context(), tenantID, userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			h.respondSCIMError(w, http.StatusNotFound, "", "User not found")
			return
		}
		h.respondSCIMError(w, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Info("SCIM user patched")

	h.respondJSON(w, http.StatusOK, user.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandleDeleteUser handles DELETE /v1/scim/Users/{id}
// RFC 7644 Section 3.6 - Delete a Resource
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID := r.PathValue("id")

	if err := h.plugin.service.DeleteUser(r.Context(), tenantID, userID); err != nil {
		if err.Error() == "user not found" {
			h.respondSCIMError(w, http.StatusNotFound, "", "User not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Info("SCIM user deleted")

	w.WriteHeader(http.StatusNoContent)
}

// HandleListGroups handles GET /v1/scim/Groups
func (h *Handler) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	filter := r.URL.Query().Get("filter")
	startIndexStr := r.URL.Query().Get("startIndex")
	countStr := r.URL.Query().Get("count")

	startIndex, count := ParsePagination(startIndexStr, countStr)

	groups, total, err := h.plugin.service.ListGroups(r.Context(), tenantID, filter, startIndex, count)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list groups")
		h.respondError(w, http.StatusInternalServerError, "Failed to list groups")
		return
	}

	resources := make([]map[string]interface{}, len(groups))
	for i, group := range groups {
		resources[i] = group.ToSCIMResponse(h.plugin.config.BaseURL)
	}

	response := NewSCIMListResponse(startIndex, count, total, resources)
	h.respondJSON(w, http.StatusOK, response)
}

// HandleCreateGroup handles POST /v1/scim/Groups
func (h *Handler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req SCIMGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondSCIMError(w, http.StatusBadRequest, "invalidSyntax", "Invalid JSON")
		return
	}

	group, err := h.plugin.service.CreateGroup(r.Context(), tenantID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.respondSCIMError(w, http.StatusConflict, "uniqueness", err.Error())
			return
		}
		h.respondSCIMError(w, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"group_id":  group.ID,
		"tenant_id": tenantID,
	}).Info("SCIM group created")

	h.respondJSON(w, http.StatusCreated, group.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandleGetGroup handles GET /v1/scim/Groups/{id}
func (h *Handler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	groupID := r.PathValue("id")

	group, err := h.plugin.service.GetGroup(r.Context(), tenantID, groupID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get group")
		h.respondError(w, http.StatusInternalServerError, "Failed to get group")
		return
	}

	if group == nil {
		h.respondSCIMError(w, http.StatusNotFound, "", "Group not found")
		return
	}

	h.respondJSON(w, http.StatusOK, group.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandleUpdateGroup handles PUT /v1/scim/Groups/{id}
func (h *Handler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	groupID := r.PathValue("id")

	var req SCIMGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondSCIMError(w, http.StatusBadRequest, "invalidSyntax", "Invalid JSON")
		return
	}

	group, err := h.plugin.service.UpdateGroup(r.Context(), tenantID, groupID, &req)
	if err != nil {
		if err.Error() == "group not found" {
			h.respondSCIMError(w, http.StatusNotFound, "", "Group not found")
			return
		}
		h.respondSCIMError(w, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"group_id":  groupID,
		"tenant_id": tenantID,
	}).Info("SCIM group updated")

	h.respondJSON(w, http.StatusOK, group.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandlePatchGroup handles PATCH /v1/scim/Groups/{id}
func (h *Handler) HandlePatchGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	groupID := r.PathValue("id")

	var req SCIMPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondSCIMError(w, http.StatusBadRequest, "invalidSyntax", "Invalid JSON")
		return
	}

	group, err := h.plugin.service.PatchGroup(r.Context(), tenantID, groupID, &req)
	if err != nil {
		if err.Error() == "group not found" {
			h.respondSCIMError(w, http.StatusNotFound, "", "Group not found")
			return
		}
		h.respondSCIMError(w, http.StatusBadRequest, "invalidValue", err.Error())
		return
	}

	h.logger.WithFields(logrus.Fields{
		"group_id":  groupID,
		"tenant_id": tenantID,
	}).Info("SCIM group patched")

	h.respondJSON(w, http.StatusOK, group.ToSCIMResponse(h.plugin.config.BaseURL))
}

// HandleDeleteGroup handles DELETE /v1/scim/Groups/{id}
func (h *Handler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, err := h.authenticate(r)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	groupID := r.PathValue("id")

	if err := h.plugin.service.DeleteGroup(r.Context(), tenantID, groupID); err != nil {
		if err.Error() == "group not found" {
			h.respondSCIMError(w, http.StatusNotFound, "", "Group not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "Failed to delete group")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"group_id":  groupID,
		"tenant_id": tenantID,
	}).Info("SCIM group deleted")

	w.WriteHeader(http.StatusNoContent)
}

// HandleServiceProviderConfig handles GET /v1/scim/ServiceProviderConfig
// RFC 7644 Section 4 - Service Provider Configuration Endpoints
func (h *Handler) HandleServiceProviderConfig(w http.ResponseWriter, r *http.Request) {
	config := DefaultServiceProviderConfig(h.plugin.config.BaseURL)
	h.respondJSON(w, http.StatusOK, config)
}

// HandleResourceTypes handles GET /v1/scim/ResourceTypes
func (h *Handler) HandleResourceTypes(w http.ResponseWriter, r *http.Request) {
	resourceTypes := SCIMResourceTypes(h.plugin.config.BaseURL)
	h.respondJSON(w, http.StatusOK, resourceTypes)
}

// HandleSchemas handles GET /v1/scim/Schemas
func (h *Handler) HandleSchemas(w http.ResponseWriter, r *http.Request) {
	schemas := SCIMSchemas(h.plugin.config.BaseURL)
	h.respondJSON(w, http.StatusOK, schemas)
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends a simple error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// respondSCIMError sends a SCIM-compliant error response
func (h *Handler) respondSCIMError(w http.ResponseWriter, status int, scimType, detail string) {
	err := NewSCIMError(status, detail)
	if scimType != "" {
		err.ScimType = scimType
	}
	h.respondJSON(w, status, err)
}
