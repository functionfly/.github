package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// extractToken extracts token from Authorization header or httpOnly cookie
func extractToken(r *http.Request) string {
	// Try Authorization header first (API clients)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Fall back to httpOnly cookie (browser clients)
	if cookie, err := r.Cookie(auth.CookieNameAccessToken); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// extractRefreshToken extracts refresh token from request body or httpOnly cookie
func extractRefreshToken(r *http.Request) string {
	// First try to read from request body (API clients)
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}

	// Fall back to httpOnly cookie (browser clients)
	if cookie, err := r.Cookie(auth.CookieNameRefreshToken); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// HandleGetSession returns session information (compatible with Supabase auth flow)
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).Error("GetSession handler panic")
			writeJSONError(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.")
		}
	}()

	tokenString := extractToken(r)
	if tokenString == "" {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	claims, err := h.authSvc.ValidateToken(r.Context(), tokenString)
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

	user, err := h.authSvc.Repo().GetUserByID(r.Context(), claims.UserID)
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
	tokenString := extractToken(r)
	logrus.WithFields(logrus.Fields{
		"hasToken": tokenString != "",
		"path":     r.URL.Path,
		"method":  r.Method,
		"ip":      getClientIP(r),
	}).Debug("HandleValidateToken called")

	if tokenString == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization header or cookie required")
		return
	}

	claims, err := h.authSvc.ValidateToken(r.Context(), tokenString)
	if err != nil {
		logrus.WithError(err).Warn("Token validation failed")
		writeJSONError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	user, err := h.authSvc.Repo().GetUserByID(r.Context(), claims.UserID)
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
	if tenant, err := h.authSvc.Repo().GetTenantByID(r.Context(), user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
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
		if name, ok := user.ProviderData["name"].(string); ok && name != "" {
			safeUser["name"] = name
		}
		if avatar, ok := user.ProviderData["avatar_url"].(string); ok && avatar != "" {
			safeUser["avatar"] = avatar
		}
	}
	if _, ok := safeUser["name"]; !ok && user.Name != "" {
		safeUser["name"] = user.Name
	}
	if user.DateOfBirth != nil {
		safeUser["date_of_birth"] = user.DateOfBirth.Format("2006-01-02")
	}

	response := map[string]interface{}{
		"token": tokenString,
		"user":  safeUser,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleLogout invalidates the current session and refresh tokens server-side
// Also clears httpOnly cookies for browser clients
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	tokenString := extractToken(r)

	if tokenString != "" {
		if claims, err := h.authSvc.ValidateToken(r.Context(), tokenString); err == nil {
			userID = &claims.UserID
		}
		if err := h.authSvc.Repo().DeleteSession(r.Context(), tokenString); err != nil {
			logrus.WithError(err).Debug("Logout: failed to delete session (may already be expired)")
		}
	}

	if userID != nil {
		if err := h.authSvc.Repo().RevokeUserRefreshTokens(r.Context(), *userID); err != nil {
			logrus.WithError(err).WithField("userID", userID).Warn("Logout: failed to revoke refresh tokens")
		}

		authEvent := &storage.AuthEvent{
			UserID:    userID,
			EventType: "logout",
			Success:   true,
			IPAddress: getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(r.Context(), authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", userID).Warn("Failed to log logout event")
		}
	}

	// Clear httpOnly cookies for browser clients
	clearAuthCookies(w)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// refreshTokenRateLimiter implements token refresh rate limiting to prevent enumeration attacks
type refreshTokenRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// newRefreshTokenRateLimiter creates a rate limiter for token refresh endpoint
func newRefreshTokenRateLimiter() *refreshTokenRateLimiter {
	rl := &refreshTokenRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    10,           // Max 10 refresh attempts
		window:   time.Minute,  // Per minute window
	}
	go rl.cleanupLoop()
	return rl
}

// global rate limiter instance for token refresh endpoint
var refreshRateLimiter = newRefreshTokenRateLimiter()

// Allow checks if a request from the given IP should be allowed
func (rl *refreshTokenRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Filter to only requests within the window
	var validRequests []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[ip] = validRequests
		return false
	}

	rl.requests[ip] = append(validRequests, now)
	return true
}

// cleanupLoop removes expired entries periodically
func (rl *refreshTokenRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *refreshTokenRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, times := range rl.requests {
		var validRequests []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}

// HandleRefreshToken handles refresh token requests to get new access tokens
// Supports both JSON body (API clients) and httpOnly cookies (browser clients)
// Rate limited to prevent token enumeration attacks
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract refresh token from body or cookie
	refreshTokenStr := extractRefreshToken(r)
	if refreshTokenStr == "" {
		writeJSONError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Rate limit by client IP for token refresh endpoint to prevent enumeration
	// This is a stricter limit than general auth rate limiting
	clientIP := getClientIP(r)
	if !refreshRateLimiter.Allow(clientIP) {
		logrus.WithField("clientIP", clientIP).Warn("Refresh token rate limit exceeded")
		writeJSONError(w, http.StatusTooManyRequests, "Too many refresh attempts. Please try again later.")
		return
	}

	tokenHash := storage.HashRefreshToken(refreshTokenStr)

	refreshToken, err := h.authSvc.Repo().GetRefreshTokenByHash(r.Context(), tokenHash)
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

	user, err := h.authSvc.Repo().GetUserByID(r.Context(), refreshToken.UserID)
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

	if err := h.authSvc.Repo().RevokeRefreshToken(r.Context(), refreshToken.ID); err != nil {
		logrus.WithError(err).WithField("tokenID", refreshToken.ID).Warn("Failed to revoke old refresh token")
	}

	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	newExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = h.authSvc.Repo().CreateRefreshToken(r.Context(), user.ID, newRefreshTokenHash, ipAddress, userAgent, newExpiresAt)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to store new refresh token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	// Set new httpOnly cookies for browser clients
	// This ensures cookie-based clients always have valid cookies
	setAuthCookies(w, newAccessToken, newRefreshToken)

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

	if tenant, err := h.authSvc.Repo().GetTenantByID(r.Context(), user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
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
	tokenString := extractToken(r)
	if tokenString == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}

	claims, err := h.authSvc.ValidateToken(r.Context(), tokenString)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Verify rememberDevices is enabled for this user
	settings, err := h.authSvc.Repo().GetUserSettings(r.Context(), claims.UserID)
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

	// Set trusted device cookie with httpOnly and SameSite=Strict
	trustedCookie := &http.Cookie{
		Name:     auth.CookieNameTrustedDevice,
		Value:    trustedToken,
		MaxAge:   auth.CookieMaxAgeTrustedDevice,
		HttpOnly: true,
		Secure:   auth.IsProduction(),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	http.SetCookie(w, trustedCookie)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"trustedToken": trustedToken,
		"expiresIn":    "30d",
	})
}
