package teams

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles team-related API requests
type Handler struct {
	repo            storage.Repository
	notify          *notification.Service
	realtimeMonitor interface{} // Will be used for broadcasting team updates
}

// TeamCreateRequest represents a request to create a team
type TeamCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TeamUpdateRequest represents a request to update a team
type TeamUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// TeamMemberRequest represents a request to add/update a team member
type TeamMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}

// TeamPermissionRequest represents a request to grant permissions to a team
type TeamPermissionRequest struct {
	ResourceType string    `json:"resource_type"`
	ResourceID   uuid.UUID `json:"resource_id"`
	Permissions  []string  `json:"permissions"`
}

// NewHandler creates a new team handler
func NewHandler(repo storage.Repository, notify *notification.Service, realtimeMonitor interface{}) *Handler {
	return &Handler{
		repo:            repo,
		notify:          notify,
		realtimeMonitor: realtimeMonitor,
	}
}

// HandleCreateTeam handles team creation
func (h *Handler) HandleCreateTeam(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req TeamCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Team name is required"))
		return
	}

	slug := uuid.New().String()[:8]
	team := &storage.Team{
		TenantID:    user.TenantID,
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		CreatedBy:   user.UserID,
	}

	if err := h.repo.CreateTeam(team); err != nil {
		logrus.WithError(err).Error("Failed to create team")
		apierror.WriteError(w, apierror.NewInternal("Failed to create team"))
		return
	}

	// Add creator as team owner
	membership := &storage.TeamMembership{
		TeamID:  team.ID,
		UserID:  user.UserID,
		Role:    auth.TeamRoleOwner,
		AddedBy: user.UserID,
	}

	if err := h.repo.AddTeamMember(membership); err != nil {
		logrus.WithError(err).Error("Failed to add team creator as owner")
		// Don't fail the request, but log the error
	}

	// Notify the creator that team was created
	if h.notify != nil {
		if err := h.notify.SendTeamCreated(r.Context(), user.UserID, team.ID, team.Name); err != nil {
			logrus.WithError(err).Warn("Failed to send team creation notification")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

// HandleGetTeam handles getting a team by ID
func (h *Handler) HandleGetTeam(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	team, err := h.repo.GetTeamByID(teamID)
	if err != nil {
		logrus.WithError(err).WithField("team_id", teamID).Error("Failed to get team")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team"))
		return
	}
	if team == nil {
		apierror.WriteError(w, apierror.NewNotFound("Team not found"))
		return
	}

	// Check if user has access to this team (must be team member or same tenant)
	membership, err := h.repo.GetTeamMembership(teamID, user.UserID)
	if err != nil && err.Error() != "record not found" {
		logrus.WithError(err).Error("Failed to check team membership")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if membership == nil && team.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Access denied"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

// HandleListTeams handles listing teams for the current user's tenant
func (h *Handler) HandleListTeams(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	teams, err := h.repo.GetTeamsByTenantID(user.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list teams")
		apierror.WriteError(w, apierror.NewInternal("Failed to list teams"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"teams": teams,
	})
}

// HandleUpdateTeam handles team updates
func (h *Handler) HandleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	team, err := h.repo.GetTeamByID(teamID)
	if err != nil {
		logrus.WithError(err).WithField("team_id", teamID).Error("Failed to get team")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team"))
		return
	}
	if team == nil {
		apierror.WriteError(w, apierror.NewNotFound("Team not found"))
		return
	}

	// Check if user is team owner or admin
	isOwner, err := h.repo.IsUserTeamOwner(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team ownership")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	isAdmin, err := h.repo.IsUserTeamAdmin(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team admin status")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isOwner && !isAdmin {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team owners and admins can update teams"))
		return
	}

	var req TeamUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name != nil {
		team.Name = *req.Name
	}
	if req.Description != nil {
		team.Description = *req.Description
	}

	if err := h.repo.UpdateTeam(team); err != nil {
		logrus.WithError(err).Error("Failed to update team")
		apierror.WriteError(w, apierror.NewInternal("Failed to update team"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

// HandleDeleteTeam handles team deletion
func (h *Handler) HandleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	team, err := h.repo.GetTeamByID(teamID)
	if err != nil {
		logrus.WithError(err).WithField("team_id", teamID).Error("Failed to get team")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team"))
		return
	}
	if team == nil {
		apierror.WriteError(w, apierror.NewNotFound("Team not found"))
		return
	}

	// Check if user is team owner
	isOwner, err := h.repo.IsUserTeamOwner(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team ownership")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isOwner {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team owners can delete teams"))
		return
	}

	// Get all team members to notify them before deletion
	var memberUserIDs []uuid.UUID
	for _, member := range team.Members {
		memberUserIDs = append(memberUserIDs, member.UserID)
	}

	teamName := team.Name
	deletedByName := user.Username
	if deletedByName == "" {
		deletedByName = user.Email
	}

	if err := h.repo.DeleteTeam(teamID); err != nil {
		logrus.WithError(err).Error("Failed to delete team")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete team"))
		return
	}

	// Notify all team members about deletion
	if h.notify != nil && len(memberUserIDs) > 0 {
		if err := h.notify.SendTeamDeleted(r.Context(), memberUserIDs, teamName, deletedByName); err != nil {
			logrus.WithError(err).Warn("Failed to send team deletion notifications")
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleAddTeamMember handles adding a member to a team
func (h *Handler) HandleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	team, err := h.repo.GetTeamByID(teamID)
	if err != nil {
		logrus.WithError(err).WithField("team_id", teamID).Error("Failed to get team")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team"))
		return
	}
	if team == nil {
		apierror.WriteError(w, apierror.NewNotFound("Team not found"))
		return
	}

	// Check if user is team admin or owner
	isAdmin, err := h.repo.IsUserTeamAdmin(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team admin status")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isAdmin {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team admins can manage members"))
		return
	}

	var req TeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate role
	validRoles := map[string]bool{
		auth.TeamRoleOwner:  true,
		auth.TeamRoleAdmin:  true,
		auth.TeamRoleMember: true,
		auth.TeamRoleViewer: true,
	}
	if !validRoles[req.Role] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid role"))
		return
	}

	// Check if user to add exists and is in the same tenant
	userToAdd, err := h.repo.GetUserByID(req.UserID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}
	if userToAdd.TenantID != team.TenantID {
		apierror.WriteError(w, apierror.NewBadRequest("Cannot add users from different tenants"))
		return
	}

	// Check if user is already a member
	existingMembership, err := h.repo.GetTeamMembership(teamID, req.UserID)
	if err != nil && err.Error() != "record not found" {
		logrus.WithError(err).Error("Failed to check existing membership")
		apierror.WriteError(w, apierror.NewInternal("Failed to check membership"))
		return
	}

	if existingMembership != nil {
		apierror.WriteError(w, apierror.NewConflict("User is already a team member"))
		return
	}

	membership := &storage.TeamMembership{
		TeamID:  teamID,
		UserID:  req.UserID,
		Role:    req.Role,
		AddedBy: user.UserID,
	}

	if err := h.repo.AddTeamMember(membership); err != nil {
		logrus.WithError(err).Error("Failed to add team member")
		apierror.WriteError(w, apierror.NewInternal("Failed to add team member"))
		return
	}

	// Notify the added user
	if h.notify != nil {
		addedByName := user.Username
		if addedByName == "" {
			addedByName = user.Email
		}
		if err := h.notify.SendTeamMemberAdded(r.Context(), req.UserID, teamID, team.Name, addedByName, req.Role); err != nil {
			logrus.WithError(err).Warn("Failed to send team member added notification")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(membership)
}

// HandleUpdateTeamMember handles updating a team member's role
func (h *Handler) HandleUpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	userIDStr := vars["userId"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	// Check if user is team admin or owner
	isAdmin, err := h.repo.IsUserTeamAdmin(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team admin status")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isAdmin {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team admins can manage members"))
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate role
	validRoles := map[string]bool{
		auth.TeamRoleOwner:  true,
		auth.TeamRoleAdmin:  true,
		auth.TeamRoleMember: true,
		auth.TeamRoleViewer: true,
	}
	if !validRoles[req.Role] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid role"))
		return
	}

	if err := h.repo.UpdateTeamMember(teamID, userID, req.Role); err != nil {
		logrus.WithError(err).Error("Failed to update team member")
		apierror.WriteError(w, apierror.NewInternal("Failed to update team member"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleRemoveTeamMember handles removing a member from a team
func (h *Handler) HandleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	userIDStr := vars["userId"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	// Check if user is team admin or owner
	isAdmin, err := h.repo.IsUserTeamAdmin(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team admin status")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isAdmin {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team admins can manage members"))
		return
	}

	// Prevent removing the last owner
	if userID == user.UserID {
		// Check if this user is the only owner
		team, err := h.repo.GetTeamByID(teamID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get team")
			apierror.WriteError(w, apierror.NewInternal("Failed to get team"))
			return
		}

		ownerCount := 0
		for _, member := range team.Members {
			if member.Role == auth.TeamRoleOwner {
				ownerCount++
			}
		}

		if ownerCount <= 1 {
			apierror.WriteError(w, apierror.NewBadRequest("Cannot remove the last team owner"))
			return
		}
	}

	// Get team details for notification
	team, err := h.repo.GetTeamByID(teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get team for removal notification")
	}

	if err := h.repo.RemoveTeamMember(teamID, userID); err != nil {
		logrus.WithError(err).Error("Failed to remove team member")
		apierror.WriteError(w, apierror.NewInternal("Failed to remove team member"))
		return
	}

	// Notify the removed user
	if h.notify != nil && team != nil {
		removedByName := user.Username
		if removedByName == "" {
			removedByName = user.Email
		}
		if err := h.notify.SendTeamMemberRemoved(r.Context(), userID, teamID, team.Name, removedByName); err != nil {
			logrus.WithError(err).Warn("Failed to send team member removal notification")
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetUserTeams handles getting teams for the current user
func (h *Handler) HandleGetUserTeams(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	teams, err := h.repo.GetUserTeams(user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user teams")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user teams"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"teams": teams,
	})
}

// HandleGrantTeamPermission handles granting permissions to a team for a resource
func (h *Handler) HandleGrantTeamPermission(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	// Check if user is team admin or owner
	isAdmin, err := h.repo.IsUserTeamAdmin(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team admin status")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isAdmin {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team admins can manage permissions"))
		return
	}

	var req TeamPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate resource type
	validResourceTypes := map[string]bool{
		"app":        true,
		"function":   true,
		"backend":    true,
		"deployment": true,
	}
	if !validResourceTypes[req.ResourceType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid resource type"))
		return
	}

	// Validate permissions
	validPermissions := map[string]bool{
		auth.PermAppsRead:          true,
		auth.PermAppsWrite:         true,
		auth.PermAppsDelete:        true,
		auth.PermFunctionsRead:     true,
		auth.PermFunctionsWrite:    true,
		auth.PermFunctionsDelete:   true,
		auth.PermBackendsRead:      true,
		auth.PermBackendsWrite:     true,
		auth.PermBackendsDelete:    true,
		auth.PermDeploymentsRead:   true,
		auth.PermDeploymentsWrite:  true,
		auth.PermDeploymentsDelete: true,
	}

	for _, perm := range req.Permissions {
		if !validPermissions[perm] {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid permission: "+perm))
			return
		}
	}

	// Convert permissions array to JSON string for storage
	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to encode permissions"))
		return
	}

	permission := &storage.TeamPermission{
		TeamID:       teamID,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Permissions:  string(permissionsJSON),
		GrantedBy:    user.UserID,
	}

	if err := h.repo.GrantTeamPermission(permission); err != nil {
		logrus.WithError(err).Error("Failed to grant team permission")
		apierror.WriteError(w, apierror.NewInternal("Failed to grant permission"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(permission)
}

// HandleRevokeTeamPermission handles revoking permissions from a team for a resource
func (h *Handler) HandleRevokeTeamPermission(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	resourceType := vars["resourceType"]
	resourceIDStr := vars["resourceId"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid resource ID"))
		return
	}

	// Check if user is team admin or owner
	isAdmin, err := h.repo.IsUserTeamAdmin(user.UserID, teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check team admin status")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if !isAdmin {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - only team admins can manage permissions"))
		return
	}

	if err := h.repo.RevokeTeamPermission(teamID, resourceType, resourceID); err != nil {
		logrus.WithError(err).Error("Failed to revoke team permission")
		apierror.WriteError(w, apierror.NewInternal("Failed to revoke permission"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetTeamPermissions handles getting all permissions for a team
func (h *Handler) HandleGetTeamPermissions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	teamIDStr := vars["teamId"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid team ID"))
		return
	}

	// Check if user is team member
	membership, err := h.repo.GetTeamMembership(teamID, user.UserID)
	if err != nil && err.Error() != "record not found" {
		logrus.WithError(err).Error("Failed to check team membership")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permissions"))
		return
	}

	if membership == nil {
		apierror.WriteError(w, apierror.NewForbidden("Access denied - not a team member"))
		return
	}

	permissions, err := h.repo.GetTeamPermissions(teamID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get team permissions")
		apierror.WriteError(w, apierror.NewInternal("Failed to get team permissions"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"permissions": permissions,
	})
}

// HandleCheckUserResourcePermission handles checking if a user has permission for a resource
func (h *Handler) HandleCheckUserResourcePermission(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	resourceType := vars["resourceType"]
	resourceIDStr := vars["resourceId"]
	permission := r.URL.Query().Get("permission")

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid resource ID"))
		return
	}

	hasPermission, err := h.repo.CheckUserResourcePermission(user.UserID, resourceType, resourceID, permission)
	if err != nil {
		logrus.WithError(err).Error("Failed to check user resource permission")
		apierror.WriteError(w, apierror.NewInternal("Failed to check permission"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"has_permission": hasPermission,
	})
}
