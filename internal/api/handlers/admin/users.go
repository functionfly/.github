package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/utils"
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
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the current admin user (who is performing the deactivation)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	deactivatedBy := claims.UserID

	// Get user before deactivation for audit
	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if user.DeactivatedAt != nil {
		http.Error(w, "User is already deactivated", http.StatusConflict)
		return
	}

	// Deactivate the user
	if err := h.repo.DeactivateUser(r.Context(), userID, deactivatedBy); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to deactivate user")
		http.Error(w, "Failed to deactivate user", http.StatusInternalServerError)
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
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user before reactivation for audit
	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if user.DeactivatedAt == nil {
		http.Error(w, "User is not deactivated", http.StatusConflict)
		return
	}

	// Reactivate the user
	if err := h.repo.ReactivateUser(r.Context(), userID); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to reactivate user")
		http.Error(w, "Failed to reactivate user", http.StatusInternalServerError)
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
	users, err := h.repo.ListUsers()
	if err != nil {
		logrus.WithError(err).Error("Failed to list users")
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
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
	users, err := h.repo.ListUsers()
	if err != nil {
		logrus.WithError(err).Error("Failed to get user stats")
		http.Error(w, "Failed to get user stats", http.StatusInternalServerError)
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
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
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
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.TenantID == uuid.Nil {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, err := h.repo.GetUserByEmail(req.Email)
	if err != nil {
		logrus.WithError(err).Error("Failed to check existing user")
		http.Error(w, "Failed to invite user", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
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
		http.Error(w, "Failed to invite user", http.StatusInternalServerError)
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
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.repo.UpdateUser(r.Context(), userID, updates)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user")
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleDeleteUser deletes a user
func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get current user to check if it's not deleting themselves
	currentUser := middleware.GetUserFromContext(r)
	if currentUser != nil && currentUser.UserID == userID {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	// For now, we'll implement soft delete by clearing sensitive data
	updates := map[string]interface{}{
		"password_hash": "", // Clear password
	}

	_, err = h.repo.UpdateUser(r.Context(), userID, updates)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to delete user")
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
