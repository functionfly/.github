package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains authentication handlers
type Handler struct {
	authSvc *auth.AuthService
}

// NewHandler creates a new auth handler
func NewHandler(authSvc *auth.AuthService) *Handler {
	return &Handler{
		authSvc: authSvc,
	}
}

// writeJSONError writes a JSON error body and status code so the frontend can parse it
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONErrorDetail(w, status, message, "")
}

// writeJSONErrorDetail writes a JSON error with optional detail (e.g. root cause for debugging)
func writeJSONErrorDetail(w http.ResponseWriter, status int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := map[string]string{"message": message}
	if detail != "" {
		body["detail"] = detail
	}
	_ = json.NewEncoder(w).Encode(body)
}

// isLoginInternalError returns true if the error is a server-side failure (DB, token, etc.) rather than bad credentials.
func isLoginInternalError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Credential/user-not-found cases -> 401
	if msg == "invalid credentials" || strings.Contains(msg, "invalid credentials") {
		return false
	}
	if strings.Contains(msg, "sql: no rows in result set") {
		return false
	}
	if strings.Contains(msg, "email not verified") {
		return false
	}
	// Server-side failures -> 500
	if msg == "internal error" {
		return true
	}
	if strings.Contains(msg, "failed to generate token") || strings.Contains(msg, "failed to verify password") {
		return true
	}
	if strings.Contains(msg, "failed to get user:") {
		return true // DB/query error, not "no rows"
	}
	return false
}

// HandleLogin handles user authentication
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Identify responses from this handler (vs proxy 500) for debugging
	w.Header().Set("X-FunctionFly-Auth", "1")
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).WithField("stack", string(debug.Stack())).Error("Login handler panic")
			writeJSONError(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.")
		}
	}()

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logrus.WithField("email", req.Email).Info("Login attempt")
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get client IP and user agent for refresh token tracking
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Check if the account is currently locked out
	user, _ := h.authSvc.Repo().GetUserByEmail(req.Email)
	if user != nil {
		lockoutUntil, err := h.authSvc.Repo().GetUserLockoutStatus(user.ID)
		if err != nil {
			logrus.WithError(err).WithField("userID", user.ID).Warn("Failed to check lockout status")
		} else if lockoutUntil != nil && time.Now().Before(*lockoutUntil) {
			remaining := time.Until(*lockoutUntil)
			minutes := int(remaining.Minutes()) + 1 // Round up
			message := fmt.Sprintf("Account is temporarily locked due to too many failed login attempts. Try again in %d minutes.", minutes)

			// Log account lockout event
			failureReason := "account_locked"
			authEvent := &storage.AuthEvent{
				UserID:        &user.ID,
				TenantID:      &user.TenantID,
				EventType:     "login_failed",
				Success:       false,
				FailureReason: &failureReason,
				IPAddress:     ipAddress,
				UserAgent:     userAgent,
			}
			if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
				logrus.WithError(logErr).WithField("userID", user.ID).Warn("Failed to log lockout auth event")
			}

			writeJSONError(w, http.StatusTooManyRequests, message)
			return
		}
	}

	response, err := h.authSvc.Login(req.Email, req.Password, ipAddress, userAgent)

	// Record login attempt and log auth event
	if user != nil {
		// Record the attempt (success will be determined below)
		_, recordErr := h.authSvc.Repo().CreateLoginAttempt(user.ID, ipAddress, userAgent, err == nil, nil)
		if recordErr != nil {
			logrus.WithError(recordErr).WithField("userID", user.ID).Warn("Failed to record login attempt")
		}

		// Log auth event
		eventType := "login"
		failureReason := ""
		if err != nil {
			eventType = "login_failed"
			failureReason = "invalid_credentials" // Default, will be overridden for lockouts
		}

		authEvent := &storage.AuthEvent{
			UserID:        &user.ID,
			TenantID:      &user.TenantID,
			EventType:     eventType,
			Success:       err == nil,
			FailureReason: &failureReason,
			IPAddress:     ipAddress,
			UserAgent:     userAgent,
		}

		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", user.ID).Warn("Failed to log auth event")
		}

		// Check if we need to implement lockout for failed attempts
		if err != nil {
			h.handleFailedLoginAttempt(user.ID, ipAddress, userAgent)
		}
	}

	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Login failed")
		if isLoginInternalError(err) {
			// Log root cause for debugging (e.g. DB connection, missing table, JWT secret)
			root := err
			for u := errors.Unwrap(root); u != nil; u = errors.Unwrap(root) {
				root = u
			}
			rootMsg := root.Error()
			logrus.WithError(err).WithField("email", req.Email).WithField("root_cause", rootMsg).Error("Login internal error (500)")
			// In development, include root cause in response so the UI can show it without checking server logs
			devDetail := ""
			if os.Getenv("DEVELOPMENT") == "true" {
				devDetail = rootMsg
			}
			// Return 503 with clear message when DB/schema is not ready so operators know to run migrations
			if strings.Contains(rootMsg, "does not exist") || strings.Contains(rootMsg, "connection refused") ||
				strings.Contains(rootMsg, "no such table") || strings.Contains(rootMsg, "JWT secret not configured") {
				msg := "Service temporarily unavailable. Ensure the database is running and migrations have been applied."
				writeJSONErrorDetail(w, http.StatusServiceUnavailable, msg, devDetail)
				return
			}
			writeJSONErrorDetail(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.", devDetail)
			return
		}
		// Safe to show to client: invalid credentials, email not verified, etc.
		msg := err.Error()
		if msg == "invalid credentials" || strings.Contains(msg, "invalid credentials") ||
			strings.Contains(msg, "sql: no rows in result set") {
			msg = "Invalid credentials"
		}
		writeJSONError(w, http.StatusUnauthorized, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleSignup handles user registration
func (h *Handler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req auth.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		writeJSONError(w, http.StatusBadRequest, "Email, password, and username are required")
		return
	}

	if req.Password != req.ConfirmPassword {
		writeJSONError(w, http.StatusBadRequest, "Passwords do not match")
		return
	}

	if !req.TermsAccepted {
		writeJSONError(w, http.StatusBadRequest, "Terms must be accepted")
		return
	}

	response, err := h.authSvc.Signup(req)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Signup failed")

		// Log signup failure
		authEvent := &storage.AuthEvent{
			EventType:     "signup",
			Success:       false,
			FailureReason: stringPtr("signup_failed"),
			IPAddress:     getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
			Metadata: map[string]interface{}{
				"email":    req.Email,
				"username": req.Username,
			},
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).Warn("Failed to log signup failure")
		}

		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Log successful signup
	authEvent := &storage.AuthEvent{
		EventType: "signup",
		Success:   true,
		IPAddress: getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Metadata: map[string]interface{}{
			"email":                 req.Email,
			"username":              req.Username,
			"email_sent":            response.EmailSent,
			"requires_verification": response.RequiresVerification,
		},
	}
	if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to log signup success")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(response)
}

// HandleSignupConfig returns public signup UI flags (e.g. invite-only mode).
func (h *Handler) HandleSignupConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(auth.SignupConfigResponse{
		InviteRequired: auth.SignupInviteRequired(),
	})
}

// HandleVerifyEmail handles email verification
func (h *Handler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSONError(w, http.StatusBadRequest, "Verification token is required")
		return
	}

	err := h.authSvc.VerifyEmail(token)
	if err != nil {
		logrus.WithError(err).Warn("Email verification failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"message": "Email verified successfully! You can now log in to your account.",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleResendVerification handles resending verification emails
func (h *Handler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	err := h.authSvc.ResendVerificationEmail(req.Email)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Resend verification failed")
		// Always return a generic success message to avoid leaking whether the email exists
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"message": "If that email address is registered and unverified, a verification email has been sent.",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleGetOAuthURL generates OAuth authorization URL for a provider.
// Optional query: redirect_uri (e.g. http://127.0.0.1:port/callback) for CLI — callback will redirect there with token.
func (h *Handler) HandleGetOAuthURL(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeJSONError(w, http.StatusBadRequest, "Provider is required")
		return
	}
	redirectURI := r.URL.Query().Get("redirect_uri")
	inviteCode := r.URL.Query().Get("invite_code")

	url, err := h.authSvc.GetOAuthURL(provider, redirectURI, inviteCode)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Warn("Failed to get OAuth URL")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]string{
		"url": url,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleOAuthCallback processes OAuth callback from providers
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		if v := mux.Vars(r); v != nil {
			provider = v["provider"]
		}
	}
	if provider == "" {
		writeJSONError(w, http.StatusBadRequest, "Provider is required")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "Authorization code is required")
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		writeJSONError(w, http.StatusBadRequest, "State parameter is required")
		return
	}

	response, err := h.authSvc.HandleOAuthCallback(provider, code, state)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Warn("OAuth callback failed")

		// Log OAuth login failure
		authEvent := &storage.AuthEvent{
			EventType:     "oauth_login",
			Success:       false,
			FailureReason: stringPtr("oauth_callback_failed"),
			IPAddress:     getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
			Provider:      &provider,
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).Warn("Failed to log OAuth login failure")
		}

		// Check if it's an OAuthError with structured error handling
		var oauthErr *auth.OAuthError
		if errors.As(err, &oauthErr) {
			// Structured OAuth error
			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:3000"
			}
			errorRedirectURL := fmt.Sprintf("%s/auth/oauth/callback?error=%s&error_description=%s",
				frontendURL, oauthErr.Type, url.QueryEscape(oauthErr.Description))
			http.Redirect(w, r, errorRedirectURL, http.StatusFound)
		} else {
			// Generic error
			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:3000"
			}
			errorRedirectURL := fmt.Sprintf("%s/auth/oauth/callback?error=unknown_error&error_description=%s",
				frontendURL, url.QueryEscape("An unexpected error occurred during authentication"))
			http.Redirect(w, r, errorRedirectURL, http.StatusFound)
		}
		return
	}

	// Log successful OAuth login
	if response.User != nil {
		authEvent := &storage.AuthEvent{
			UserID:    &response.User.ID,
			TenantID:  &response.User.TenantID,
			EventType: "oauth_login",
			Success:   true,
			IPAddress: getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
			Provider:  &provider,
			Metadata: map[string]interface{}{
				"new_user": response.NewUser,
			},
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", response.User.ID).Warn("Failed to log OAuth login success")
		}
	}

	// Redirect: if client requested a redirect_uri (e.g. CLI), use it; otherwise frontend
	if response.RedirectURI != "" {
		successRedirectURL := fmt.Sprintf("%s?token=%s&refresh_token=%s&new_user=%t",
			response.RedirectURI, response.Token, response.RefreshToken, response.NewUser)
		http.Redirect(w, r, successRedirectURL, http.StatusFound)
		return
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	successRedirectURL := fmt.Sprintf("%s/auth/oauth/callback?token=%s&refresh_token=%s&new_user=%t",
		frontendURL, response.Token, response.RefreshToken, response.NewUser)
	http.Redirect(w, r, successRedirectURL, http.StatusFound)
}

// HandleGetOAuthProviders returns list of configured OAuth providers
func (h *Handler) HandleGetOAuthProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.authSvc.GetConfiguredOAuthProviders()

	response := map[string]interface{}{
		"providers": providers,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleGetSession returns session information (compatible with Supabase auth flow)
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			logrus.WithField("panic", rec).Error("GetSession handler panic")
			writeJSONError(w, http.StatusInternalServerError, "An unexpected error occurred. Please try again.")
		}
	}()

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// No token provided - return empty session
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		// Invalid format - return empty session
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

	// Validate token
	claims, err := h.authSvc.ValidateToken(tokenString)
	if err != nil {
		logrus.WithError(err).Warn("Token validation failed")
		// Invalid token - return empty session
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	// Get user details
	user, err := h.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user by ID")
		// User not found - return empty session
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
		// User not found - return empty session
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"session": nil,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	// Extract name and avatar from provider data or use defaults
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

	// Create session-like response compatible with Supabase format
	session := map[string]interface{}{
		"access_token":  tokenString,
		"refresh_token": "", // FunctionFly doesn't use refresh tokens in the same way
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

// HandleValidateToken validates a JWT token and returns safe user information (no sensitive fields)
func (h *Handler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	tokenString := parts[1]

	// Validate token
	claims, err := h.authSvc.ValidateToken(tokenString)
	if err != nil {
		logrus.WithError(err).Warn("Token validation failed")
		writeJSONError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Get user details
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

	// Load tenant for plan (billing/UI)
	plan := ""
	if tenant, err := h.authSvc.Repo().GetTenantByID(user.TenantID); err == nil && tenant != nil && tenant.Plan != "" {
		plan = tenant.Plan
	}

	// Return only the safe subset of user data — never expose password hash, MFA secrets, or tokens
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

	// Include provider display fields if present
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

// HandlePasswordResetRequest initiates a password reset by sending an email with a reset token.
// Always returns 200 to avoid leaking whether the email exists.
func (h *Handler) HandlePasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Best-effort — don't reveal whether the email exists
	err := h.authSvc.RequestPasswordReset(req.Email)

	// Log password reset request (regardless of success/failure to avoid leaking user existence)
	authEvent := &storage.AuthEvent{
		EventType: "password_reset_request",
		Success:   err == nil,
		IPAddress: getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Metadata: map[string]interface{}{
			"email": req.Email,
		},
	}
	if err != nil {
		authEvent.FailureReason = stringPtr("reset_request_failed")
	}
	if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to log password reset request")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "If that email is registered, a password reset link has been sent.",
	})
}

// HandlePasswordResetConfirm completes a password reset using the token from the email.
func (h *Handler) HandlePasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Token == "" || req.NewPassword == "" {
		writeJSONError(w, http.StatusBadRequest, "token and newPassword are required")
		return
	}

	err := h.authSvc.ConfirmPasswordReset(req.Token, req.NewPassword)
	if err != nil {
		logrus.WithError(err).Debug("Password reset confirm failed")

		// Log password reset confirmation failure
		authEvent := &storage.AuthEvent{
			EventType:     "password_reset_confirm",
			Success:       false,
			FailureReason: stringPtr("invalid_token"),
			IPAddress:     getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).Warn("Failed to log password reset confirmation failure")
		}

		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Log successful password reset
	authEvent := &storage.AuthEvent{
		EventType: "password_reset_confirm",
		Success:   true,
		IPAddress: getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
	}
	if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
		logrus.WithError(logErr).Warn("Failed to log password reset confirmation success")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

// HandleCheckUsernameAvailability checks if a username is available for registration
func (h *Handler) HandleCheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		writeJSONError(w, http.StatusBadRequest, "Username parameter is required")
		return
	}

	available, err := h.authSvc.CheckUsernameAvailability(username)
	if err != nil {
		logrus.WithError(err).WithField("username", username).Warn("Username availability check failed")
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]interface{}{
		"available": available,
		"username":  username,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleLogout invalidates the current session and refresh tokens server-side
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the JWT token
	authHeader := r.Header.Get("Authorization")
	var userID *uuid.UUID

	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			// Validate token to get user ID
			if claims, err := h.authSvc.ValidateToken(tokenString); err == nil {
				userID = &claims.UserID
			}
			// Best-effort: delete the session; ignore errors (token may already be expired)
			if err := h.authSvc.Repo().DeleteSession(tokenString); err != nil {
				logrus.WithError(err).Debug("Logout: failed to delete session (may already be expired)")
			}
		}
	}

	// Revoke all refresh tokens for the user
	if userID != nil {
		if err := h.authSvc.Repo().RevokeUserRefreshTokens(*userID); err != nil {
			logrus.WithError(err).WithField("userID", userID).Warn("Logout: failed to revoke refresh tokens")
		}

		// Log logout event
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

// VerifyPasswordRequest is the body for POST /auth/verify-password (re-auth for sensitive actions).
type VerifyPasswordRequest struct {
	Password string `json:"password"`
}

// HandleVerifyPassword verifies the current user's password. Used for re-authentication (e.g. reveal secret).
// Requires valid JWT. Returns 200 on success, 401 on wrong password, 400 if account has no password (e.g. SSO).
func (h *Handler) HandleVerifyPassword(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSONError(w, http.StatusUnauthorized, "Authorization required")
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeJSONError(w, http.StatusUnauthorized, "Invalid authorization header")
		return
	}
	claims, err := h.authSvc.ValidateToken(parts[1])
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	var req VerifyPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Password is required")
		return
	}

	user, err := h.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil || user == nil {
		writeJSONError(w, http.StatusUnauthorized, "User not found")
		return
	}
	if user.PasswordHash == "" {
		writeJSONError(w, http.StatusBadRequest, "This account uses sign-in with a provider; password verification is not available")
		return
	}

	valid, err := h.authSvc.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Debug("Verify password error")
		writeJSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}
	if !valid {
		writeJSONError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Password verified"})
}

// HandleRefreshToken handles refresh token requests to get new access tokens
func (h *Handler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
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

	// Hash the provided refresh token for lookup
	tokenHash := storage.HashRefreshToken(req.RefreshToken)

	// Get refresh token from database
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

	// Get user associated with the refresh token
	user, err := h.authSvc.Repo().GetUserByID(refreshToken.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", refreshToken.UserID).Warn("Failed to get user for refresh token")
		writeJSONError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Generate new access token
	newAccessToken, err := h.authSvc.GenerateToken(user)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to generate new access token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	// Optionally generate new refresh token (token rotation)
	newRefreshToken, newRefreshTokenHash, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to generate new refresh token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Revoke old refresh token and create new one
	if err := h.authSvc.Repo().RevokeRefreshToken(refreshToken.ID); err != nil {
		logrus.WithError(err).WithField("tokenID", refreshToken.ID).Warn("Failed to revoke old refresh token")
	}

	// Get client info for new refresh token
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Store new refresh token
	newExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	_, err = h.authSvc.Repo().CreateRefreshToken(user.ID, newRefreshTokenHash, ipAddress, userAgent, newExpiresAt)
	if err != nil {
		logrus.WithError(err).WithField("userID", user.ID).Error("Failed to store new refresh token")
		writeJSONError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	// Build safe user response
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

	// Load tenant for plan
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

// handleFailedLoginAttempt checks if a user should be locked out after failed attempts
func (h *Handler) handleFailedLoginAttempt(userID uuid.UUID, ipAddress, userAgent string) {
	const maxFailedAttempts = 5
	const lockoutDuration = 15 * time.Minute // 15 minutes lockout
	const failureWindow = 15 * time.Minute   // Check failures in last 15 minutes

	// Count recent failed attempts
	failedCount, err := h.authSvc.Repo().GetRecentFailedLoginAttempts(userID, time.Now().Add(-failureWindow))
	if err != nil {
		logrus.WithError(err).WithField("userID", userID).Warn("Failed to count recent failed login attempts")
		return
	}

	// If we've exceeded the threshold, lock the account
	if failedCount >= maxFailedAttempts {
		lockoutUntil := time.Now().Add(lockoutDuration)

		// Record the lockout attempt
		_, err = h.authSvc.Repo().CreateLoginAttempt(userID, ipAddress, userAgent, false, &lockoutUntil)
		if err != nil {
			logrus.WithError(err).WithField("userID", userID).Warn("Failed to record lockout")
			return
		}

		// Log account lockout event
		eventType := "account_locked"
		authEvent := &storage.AuthEvent{
			UserID:    &userID,
			EventType: eventType,
			Success:   false,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Metadata: map[string]interface{}{
				"failed_attempts":          failedCount,
				"lockout_duration_minutes": 15,
			},
		}
		if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
			logrus.WithError(logErr).WithField("userID", userID).Warn("Failed to log account lockout event")
		}

		logrus.WithFields(logrus.Fields{
			"userID":         userID,
			"failedAttempts": failedCount,
			"lockoutUntil":   lockoutUntil,
		}).Warn("Account locked due to too many failed login attempts")
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// getClientIP extracts the real client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (common with proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
