package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleDeactivateUser deactivates a user (soft delete)
// POST /v1/admin/users/{userId}/deactivate
func (h *Handler) HandleDeactivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	// Get the current admin user (who is performing the deactivation)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}
	deactivatedBy := claims.UserID

	// Get user before deactivation for audit
	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user"))
		return
	}
	if user == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}
	if user.DeactivatedAt != nil {
		apierror.WriteError(w, apierror.NewConflict("User is already deactivated"))
		return
	}

	// Deactivate the user
	if err := h.repo.DeactivateUser(r.Context(), userID, deactivatedBy); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to deactivate user")
		apierror.WriteError(w, apierror.NewInternal("Failed to deactivate user"))
		return
	}

	// Log audit event
	utils.LogAuditEvent(r.Context(), h.repo, r, "user.deactivated", "user", &userID, user, nil, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User deactivated successfully",
		"user_id": userID,
	})
}

// HandleReactivateUser reactivates a previously deactivated user
// POST /v1/admin/users/{userId}/reactivate
func (h *Handler) HandleReactivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	// Get user before reactivation for audit
	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user"))
		return
	}
	if user == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}
	if user.DeactivatedAt == nil {
		apierror.WriteError(w, apierror.NewConflict("User is not deactivated"))
		return
	}

	// Reactivate the user
	if err := h.repo.ReactivateUser(r.Context(), userID); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to reactivate user")
		apierror.WriteError(w, apierror.NewInternal("Failed to reactivate user"))
		return
	}

	// Log audit event
	utils.LogAuditEvent(r.Context(), h.repo, r, "user.reactivated", "user", &userID, user, nil, true)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User reactivated successfully",
		"user_id": userID,
	})
}

// HandleListUsers lists all platform users (all tenants) for admin management.
func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListUsers(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list users")
		apierror.WriteError(w, apierror.NewInternal("Failed to list users"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	payload := map[string]interface{}{
		"data":      users,
		"users":     users,
		"total":     len(users),
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(payload)
}

// HandleGetUserStats returns user statistics
func (h *Handler) HandleGetUserStats(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListUsers(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to get user stats")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user stats"))
		return
	}

	totalUsers := len(users)
	adminUsers := 0
	activeUsers := 0

	for _, user := range users {
		if user.Role == "super_admin" || user.Role == "admin" || user.Role == "support" || user.Role == "billing_admin" || user.Role == "developer_admin" {
			adminUsers++
		}
		// Consider users created in the last 30 days as active (simplified)
		if user.CreatedAt.After(time.Now().AddDate(0, 0, -30)) {
			activeUsers++
		}
	}

	stats := map[string]interface{}{
		"total_users":  totalUsers,
		"admin_users":  adminUsers,
		"active_users": activeUsers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetUser gets a specific user
func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user"))
		return
	}
	if user == nil {
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	// Fetch tenant to get plan and attach to response
	tenant, _ := h.repo.GetTenantByID(r.Context(), user.TenantID)
	response := map[string]interface{}{}
	bytes, _ := json.Marshal(user)
	json.Unmarshal(bytes, &response)
	if tenant != nil {
		response["plan"] = tenant.Plan
		response["tenant_name"] = tenant.Name
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      response,
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HandleCreateUser creates a new user
func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string    `json:"email"`
		Password string    `json:"password"`
		TenantID uuid.UUID `json:"tenant_id"`
		Role     string    `json:"role,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Hash the password
	hashedPassword, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		apierror.WriteError(w, apierror.NewInternal("Failed to create user"))
		return
	}

	user := &storage.User{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdUser, err := h.repo.CreateUserWithRole(r.Context(), user)
	if err != nil {
		logrus.WithError(err).Error("Failed to create user")
		apierror.WriteError(w, apierror.NewInternal("Failed to create user"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

// HandleInviteUser invites a user
func (h *Handler) HandleInviteUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string    `json:"email"`
		TenantID uuid.UUID `json:"tenant_id"`
		Role     string    `json:"role,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	if req.Email == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Email is required"))
		return
	}

	if req.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewBadRequest("Tenant ID is required"))
		return
	}

	// Check if user already exists
	existingUser, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		logrus.WithError(err).Error("Failed to check existing user")
		apierror.WriteError(w, apierror.NewInternal("Failed to invite user"))
		return
	}

	if existingUser != nil {
		apierror.WriteError(w, apierror.NewConflict("User already exists"))
		return
	}

	// Create user without password (SSO-based authentication)
	user := &storage.User{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		Email:        req.Email,
		PasswordHash: "", // No password for SSO users
		Role:         req.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	invitedUser, err := h.repo.CreateUserWithRole(r.Context(), user)
	if err != nil {
		logrus.WithError(err).Error("Failed to invite user")
		apierror.WriteError(w, apierror.NewInternal("Failed to invite user"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invitedUser)
}

// HandleUpdateUser updates a user
func (h *Handler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Handle plan update separately since plan is a tenant field, not a user field
	if newPlan, ok := updates["plan"].(string); ok {
		logrus.WithFields(logrus.Fields{"user_id": userID, "new_plan": newPlan}).Info("Plan update request received")
		user, err := h.repo.GetUserByID(r.Context(), userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user for plan update")
			apierror.WriteError(w, apierror.NewInternal("Failed to get user"))
			return
		}
		if user == nil {
			apierror.WriteError(w, apierror.NewNotFound("User not found"))
			return
		}
		logrus.WithFields(logrus.Fields{"tenant_id": user.TenantID, "new_plan": newPlan}).Info("Updating tenant plan")
		// Update the tenant's plan
		tenantUpdates := map[string]interface{}{"plan": newPlan}
		updatedTenant, err := h.repo.UpdateTenant(r.Context(), user.TenantID, tenantUpdates)
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", user.TenantID).Error("Failed to update tenant plan")
			apierror.WriteError(w, apierror.NewInternal("Failed to update plan"))
			return
		}
		logrus.WithFields(logrus.Fields{"tenant_id": user.TenantID, "new_plan": updatedTenant.Plan}).Info("Tenant plan updated successfully")
		// Remove plan from user updates since we handled it via tenant
		delete(updates, "plan")
	}

	// Only call UpdateUser if there are remaining user-level fields to update
	if len(updates) > 0 {
		user, err := h.repo.UpdateUser(r.Context(), userID, updates)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user")
			apierror.WriteError(w, apierror.NewInternal("Failed to update user"))
			return
		}
		// Fetch tenant to get current plan and attach to response
		tenant, _ := h.repo.GetTenantByID(r.Context(), user.TenantID)
		response := map[string]interface{}{}
		bytes, _ := json.Marshal(user)
		json.Unmarshal(bytes, &response)
		if tenant != nil {
			response["plan"] = tenant.Plan
			response["tenant_name"] = tenant.Name
		}
		// Wrap in AdminAPIResponse format
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":      response,
			"success":   true,
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	// If only plan was updated, fetch and return the updated user with tenant plan
	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user after plan update")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user"))
		return
	}
	// Fetch the tenant to get the updated plan
	tenant, err := h.repo.GetTenantByID(r.Context(), user.TenantID)
	if err == nil && tenant != nil {
		user.Tenant = tenant
	}
	// Build response with plan flattened from tenant (frontend expects user.plan)
	response := map[string]interface{}{}
	bytes, _ := json.Marshal(user)
	json.Unmarshal(bytes, &response)
	if tenant != nil {
		response["plan"] = tenant.Plan
		response["tenant_name"] = tenant.Name
	}
	// Wrap in AdminAPIResponse format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      response,
		"success":   true,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HandleDeleteUser deletes a user
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid user ID"))
		return
	}

	// Get current user to check if it's not deleting themselves
	currentUser := middleware.GetUserFromContext(r)
	if currentUser != nil && currentUser.UserID == userID {
		apierror.WriteError(w, apierror.NewBadRequest("Cannot delete your own account"))
		return
	}

	// For now, we'll implement soft delete by clearing sensitive data
	updates := map[string]interface{}{
		"password_hash": "", // Clear password
	}

	_, err = h.repo.UpdateUser(r.Context(), userID, updates)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to delete user")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete user"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
