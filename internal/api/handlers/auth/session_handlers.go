package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HandleGetSession returns session information (compatible with Supabase auth flow)
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).Error("GetSession handler panic")
			writeJSONError(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.")
		}
	}()

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	tokenString := parts[1]

	claims, err := h.authSvc.ValidateToken(tokenString)
	if err != nil {
		logrus.WithError(err).Warn("Token validation failed")
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	user, err := h.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user by ID")
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}
	if user == nil {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	name := ""
	avatar := ""

	if user.ProviderData != nil {
		if providerName, ok := user.ProviderData["name"].(string); ok {
			name = providerName
		}
		if providerFullName, ok := user.ProviderData["full_name"].(string); ok && name == "" {
			name = providerFullName
		}
		if providerAvatar, ok := user.ProviderData["avatar_url"].(string); ok {
			avatar = providerAvatar
		}
		if providerPicture, ok := user.ProviderData["picture"].(string); ok && avatar == "" {
			avatar = providerPicture
		}
	}

	session := map[string]interface{}{
		"access_token":  tokenString,
		"refresh_token": "",
		"expires_at":    claims.ExpiresAt.Unix(),
		"token_type":    "bearer",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"username": func() interface{} {
				if user.Username != nil && *user.Username != "" {
					return *user.Username
				}
				return nil
			}(),
			"company_name": func() interface{} {
				if user.CompanyName != nil && *user.CompanyName != "" {
					return *user.CompanyName
				}
				return nil
			}(),
			"user_metadata": map[string]interface{}{
				"name":       name,
				"avatar_url": avatar,
			},
			"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at": user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"session": session,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleValidateToken validates a JWT token and returns safe user information
func (h *Handler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	tokenString := parts[1]

	claims, err := h.authSvc.ValidateToken(tokenString)
	if err != nil {
		logrus.WithError(err).Warn("Token validation failed")
		writeJSONError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	user, err := h.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user by ID")
		writeJSONError(w, http.StatusUnauthorized, "User not found")
		return
	}
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "User not found")
		return
	}

	plan := ""
	if tenant, err := h.authSvc.Repo().GetTenantByID(user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
		plan = tenant.Plan
	}

	safeUser := map[string]interface{}{
		"id":             user.ID,
		"tenant_id":      user.TenantID,
		"email":          user.Email,
		"role":           user.Role,
		"plan":           plan,
		"email_verified": user.EmailVerified,
		"mfa_enabled":    user.MFAEnabled,
		"created_at":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if user.Username != nil && *user.Username != "" {
		safeUser["username"] = *user.Username
	}
	if user.CompanyName != nil && *user.CompanyName != "" {
		safeUser["company_name"] = *user.CompanyName
	}

	if user.ProviderData != nil {
		if name, ok := user.ProviderData["name"].(string); ok {
			safeUser["name"] = name
		}
		if avatar, ok := user.ProviderData["avatar_url"].(string); ok {
			safeUser["avatar"] = avatar
		}
	}

	response := map[string]interface{}{
		"token": tokenString,
		"user":  safeUser,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleLogout invalidates the current session and refresh tokens server-side
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	var userID *uuid.UUID

	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			if claims, err := h.authSvc.ValidateToken(tokenString); err == nil {
				userID = &claims.UserID
			}
			if err := h.authSvc.Repo().DeleteSession(tokenString); err != nil {
				logrus.WithError(err).Debug("Logout: failed to delete session (may already be expired)")
			}
		}
	}

	if userID != nil {
		if err := h.authSvc.Repo().RevokeUserRefreshTokens(*userID); err != nil {
			logrus.WithError(err).WithField("userID", userID).Warn("Logout: failed to revoke refresh tokens")
		}

		authEvent := &storage.AuthEvent{
			UserID:    userID,
			EventType: "logout",
			Success:   true,
			IPAddress: getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", userID).Warn("Failed to log logout event")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// HandleRefreshToken handles refresh token requests to get new access tokens
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeJSONError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	tokenHash := storage.HashRefreshToken(req.RefreshToken)

	refreshToken, err := h.authSvc.Repo().GetRefreshTokenByHash(tokenHash)
	if err != nil {
		logrus.WithError(err).Warn("Refresh token lookup failed")
		writeJSONError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if refreshToken == nil {
		writeJSONError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if refreshToken.Revoked {
		logrus.WithField("tokenID", refreshToken.ID).Warn("Attempted to use revoked refresh token")
		writeJSONError(w, http.StatusUnauthorized, "Refresh token has been revoked")
		return
	}

	user, err := h.authSvc.Repo().GetUserByID(refreshToken.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", refreshToken.UserID).Warn("Failed to get user for refresh token")
		writeJSONError(w, http.StatusUnauthorized, "User not found")
		return
	}

	newAccessToken, err := h.authSvc.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to generate new access token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	newRefreshToken, newRefreshTokenHash, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to generate new refresh token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	if err := h.authSvc.Repo().RevokeRefreshToken(refreshToken.ID); err != nil {
		logrus.WithError(err).WithField("tokenID", refreshToken.ID).Warn("Failed to revoke old refresh token")
	}

	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	newExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = h.authSvc.Repo().CreateRefreshToken(user.ID, newRefreshTokenHash, ipAddress, userAgent, newExpiresAt)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to store new refresh token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	safeUser := map[string]interface{}{
		"id":             user.ID,
		"tenant_id":      user.TenantID,
		"email":          user.Email,
		"role":           user.Role,
		"email_verified": user.EmailVerified,
		"mfa_enabled":    user.MFAEnabled,
		"created_at":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at":     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if user.Username != nil && *user.Username != "" {
		safeUser["username"] = *user.Username
	}
	if user.CompanyName != nil && *user.CompanyName != "" {
		safeUser["company_name"] = *user.CompanyName
	}

	if tenant, err := h.authSvc.Repo().GetTenantByID(user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
		safeUser["plan"] = tenant.Plan
	}

	response := map[string]interface{}{
		"token":         newAccessToken,
		"refresh_token": newRefreshToken,
		"user":          safeUser,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleTrustedDeviceRequest generates a trusted device token for the authenticated user.
// POST /auth/trusted-device
// Client stores this token and sends it on future logins via X-Trusted-Device-Token header
// to get a 30-day session instead of the default session duration.
func (h *Handler) HandleTrustedDeviceRequest(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	claims, err := h.authSvc.ValidateToken(parts[1])
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Verify rememberDevices is enabled for this user
	settings, err := h.authSvc.Repo().GetUserSettings(claims.UserID)
	if err == nil && settings != nil {
		if val, ok := settings["rememberDevices"]; ok {
			if b, ok := val.(bool); ok && !b {
				writeJSONError(w, http.StatusForbidden, "Remember devices is disabled")
				return
			}
		}
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logrus.WithError(err).Error("Failed to generate trusted device token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	trustedToken := base64.URLEncoding.EncodeToString(tokenBytes)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"trustedToken": trustedToken,
		"expiresIn":    "30d",
	})
}
