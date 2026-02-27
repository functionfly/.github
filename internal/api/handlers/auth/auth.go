package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
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

// handleLogin handles user authentication
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

	// #region agent log
	payloadA := map[string]interface{}{"sessionId": "c3cea7", "hypothesisId": "A", "location": "auth.go:HandleLogin", "message": "handler entered, calling Login", "data": map[string]interface{}{"email": req.Email}, "timestamp": time.Now().UnixMilli()}
	if f, _ := os.OpenFile("/home/micro/projects/functionfly/.cursor/debug-c3cea7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
		b, _ := json.Marshal(payloadA)
		f.Write(append(b, '\n'))
		f.Close()
	}
	logrus.WithField("debug_c3cea7", payloadA).Info("agent log")
	// #endregion

	response, err := h.authSvc.Login(req.Email, req.Password)

	// #region agent log
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	payloadB := map[string]interface{}{"sessionId": "c3cea7", "hypothesisId": "B", "location": "auth.go:HandleLogin after Login", "message": "Login returned", "data": map[string]interface{}{"hasErr": err != nil, "errMsg": errStr, "isInternal": err != nil && isLoginInternalError(err), "hasResponse": response != nil}, "timestamp": time.Now().UnixMilli()}
	if f, _ := os.OpenFile("/home/micro/projects/functionfly/.cursor/debug-c3cea7.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
		b, _ := json.Marshal(payloadB)
		f.Write(append(b, '\n'))
		f.Close()
	}
	logrus.WithField("debug_c3cea7", payloadB).Info("agent log")
	// #endregion

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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	if req.Password != req.ConfirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	if !req.TermsAccepted {
		http.Error(w, "Terms must be accepted", http.StatusBadRequest)
		return
	}

	response, err := h.authSvc.Signup(req)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Signup failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(response)
}

// HandleVerifyEmail handles email verification
func (h *Handler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}

	err := h.authSvc.VerifyEmail(token)
	if err != nil {
		logrus.WithError(err).WithField("token", token).Warn("Email verification failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"message": "Email verified successfully! You can now log in to your account.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleResendVerification handles resending verification emails
func (h *Handler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	err := h.authSvc.ResendVerificationEmail(req.Email)
	if err != nil {
		logrus.WithError(err).WithField("email", req.Email).Warn("Resend verification failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"message": "Verification email sent successfully. Please check your email.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetOAuthURL generates OAuth authorization URL for a provider
func (h *Handler) HandleGetOAuthURL(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	url, err := h.authSvc.GetOAuthURL(provider)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Warn("Failed to get OAuth URL")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"url": url,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleOAuthCallback processes OAuth callback from providers
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code is required", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "State parameter is required", http.StatusBadRequest)
		return
	}

	response, err := h.authSvc.HandleOAuthCallback(provider, code, state)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Warn("OAuth callback failed")

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

	// Redirect to frontend with success and token
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	successRedirectURL := fmt.Sprintf("%s/auth/oauth/callback?token=%s&new_user=%t",
		frontendURL, response.Token, response.NewUser)
	http.Redirect(w, r, successRedirectURL, http.StatusFound)
}

// HandleGetOAuthProviders returns list of configured OAuth providers
func (h *Handler) HandleGetOAuthProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.authSvc.GetConfiguredOAuthProviders()

	response := map[string]interface{}{
		"providers": providers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
	json.NewEncoder(w).Encode(response)
}

// HandleValidateToken validates a JWT token and returns user information
func (h *Handler) HandleValidateToken(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	tokenString := parts[1]

	// Validate token
	claims, err := h.authSvc.ValidateToken(tokenString)
	if err != nil {
		logrus.WithError(err).Warn("Token validation failed")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := h.authSvc.Repo().GetUserByID(claims.UserID)
	if err != nil {
		logrus.WithError(err).WithField("userID", claims.UserID).Warn("Failed to get user by ID")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"token": tokenString,
		"user":  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
