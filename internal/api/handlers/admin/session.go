package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/sirupsen/logrus"
)

// HandleGetAdminSession returns normalized session + user payload for admin SPA bootstrap.
func (h *Handler) HandleGetAdminSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("Failed to resolve admin user for session bootstrap")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	plan := ""
	if tenant, terr := h.repo.GetTenantByID(user.TenantID); terr == nil && tenant != nil {
		plan = tenant.Plan
	}

	token := ""
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		token = parts[1]
	}

	now := time.Now().UTC()
	expiresAt := now.Add(30 * time.Minute)
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time.UTC()
	}

	name := ""
	avatar := ""
	if user.ProviderData != nil {
		if v, ok := user.ProviderData["name"].(string); ok {
			name = v
		}
		if v, ok := user.ProviderData["avatar_url"].(string); ok {
			avatar = v
		}
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	session := map[string]interface{}{
		"id":                 fmt.Sprintf("jwt-%s", claims.UserID.String()),
		"user_id":            claims.UserID.String(),
		"session_token_hash": "jwt",
		"access_token":       token,
		"ip_address":         extractClientIP(r),
		"user_agent":         r.UserAgent(),
		"created_at":         now.Format(time.RFC3339),
		"last_activity_at":   now.Format(time.RFC3339),
		"expires_at":         expiresAt.Format(time.RFC3339),
	}

	respUser := map[string]interface{}{
		"id":          user.ID.String(),
		"email":       user.Email,
		"name":        name,
		"avatar":      avatar,
		"username":    username,
		"tenant_id":   user.TenantID.String(),
		"plan":        plan,
		"role":        claims.Role,
		"permissions": claims.Permissions,
		"mfa_enabled": user.MFAEnabled,
		"created_at":  user.CreatedAt.Format(time.RFC3339),
		"updated_at":  user.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session": session,
		"user":    respUser,
	})
}

// HandleGetAdminLastLogin returns the most recent successful login for the authenticated admin user.
// The response shape matches what the admin SPA's login page expects: ip_address, device_name, timestamp, suspicious.
func (h *Handler) HandleGetAdminLastLogin(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lastAttempt, err := h.loginAttemptRepo.GetLastSuccessfulLogin(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("Failed to fetch last login")
		http.Error(w, "Failed to retrieve last login", http.StatusInternalServerError)
		return
	}

	if lastAttempt == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"last_login": nil})
		return
	}

	// Determine if the login is "suspicious" — flag if the IP or user-agent differs from the current request.
	currentIP := extractClientIP(r)
	currentUA := r.UserAgent()
	suspicious := lastAttempt.IPAddress != currentIP || lastAttempt.UserAgent != currentUA

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"last_login": map[string]interface{}{
			"ip_address":  lastAttempt.IPAddress,
			"device_name": lastAttempt.UserAgent,
			"timestamp":   lastAttempt.AttemptedAt.Format(time.RFC3339),
			"suspicious":  suspicious,
		},
	})
}

// HandleExtendAdminSession issues a new JWT with extended expiry and returns session + user (same shape as GET session).
// Called when the user clicks "Extend Session" so the countdown resets and the session is actually extended.
func (h *Handler) HandleExtendAdminSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("Failed to resolve admin user for session extend")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	newToken, err := h.authSvc.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Error("Failed to generate token for session extend")
		http.Error(w, "Failed to extend session", http.StatusInternalServerError)
		return
	}

	plan := ""
	if tenant, terr := h.repo.GetTenantByID(user.TenantID); terr == nil && tenant != nil {
		plan = tenant.Plan
	}

	now := time.Now().UTC()
	expiresAt := now.Add(30 * time.Minute)

	session := map[string]interface{}{
		"id":                 fmt.Sprintf("jwt-%s", claims.UserID.String()),
		"user_id":            claims.UserID.String(),
		"session_token_hash": "jwt",
		"access_token":       newToken,
		"ip_address":         extractClientIP(r),
		"user_agent":         r.UserAgent(),
		"created_at":         now.Format(time.RFC3339),
		"last_activity_at":   now.Format(time.RFC3339),
		"expires_at":         expiresAt.Format(time.RFC3339),
	}

	name := ""
	avatar := ""
	if user.ProviderData != nil {
		if v, ok := user.ProviderData["name"].(string); ok {
			name = v
		}
		if v, ok := user.ProviderData["avatar_url"].(string); ok {
			avatar = v
		}
	}
	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	respUser := map[string]interface{}{
		"id":          user.ID.String(),
		"email":       user.Email,
		"name":        name,
		"avatar":      avatar,
		"username":    username,
		"tenant_id":   user.TenantID.String(),
		"plan":        plan,
		"role":        claims.Role,
		"permissions": claims.Permissions,
		"mfa_enabled": user.MFAEnabled,
		"created_at":  user.CreatedAt.Format(time.RFC3339),
		"updated_at":  user.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session": session,
		"user":    respUser,
	})
}
