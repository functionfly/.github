package gba

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
)

// Handler contains HTTP handlers for GoBetterAuth
type Handler struct {
	auth     *Auth
	db       *gorm.DB
	logger   *logrus.Logger
	userRepo *storage.UserRepository // for checking user settings (rememberDevices)
}

// NewHandler creates a new GoBetterAuth handler
func NewHandler(auth *Auth, userRepo *storage.UserRepository) *Handler {
	return &Handler{
		auth:     auth,
		db:       auth.GetDB(),
		logger:   auth.Logger(),
		userRepo: userRepo,
	}
}

// SignUpRequest represents a sign-up request
type SignUpRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// SignUpResponse represents a sign-up response
type SignUpResponse struct {
	User    *User  `json:"user"`
	Message string `json:"message,omitempty"`
}

// HandleSignUp handles user registration
func (h *Handler) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	if !h.auth.IsEnabled("register") {
		http.Error(w, `{"message": "Registration is disabled"}`, http.StatusServiceUnavailable)
		return
	}

	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"message": "Email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Extract tenant context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, `{"message": "Tenant context is required"}`, http.StatusBadRequest)
		return
	}

	// Run before:signup hooks
	hookReq := &HookRequest{
		Email:     req.Email,
		TenantID:  tenantID,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Host:      r.Host,
		Headers:   h.extractHeaders(r),
	}

	if err := h.auth.hooks.Execute(r.Context(), "before:signup", hookReq); err != nil {
		h.logger.WithError(err).Warn("Signup hook failed")
		http.Error(w, `{"message": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser User
	if err := h.db.Where("tenant_id = ? AND email = ?", tenantID, req.Email).First(&existingUser).Error; err == nil {
		http.Error(w, `{"message": "User already exists"}`, http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Create user
	user := &User{
		TenantID: tenantID,
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	if err := h.db.Create(user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		http.Error(w, `{"message": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Generate verification token
	verificationToken, err := generateSessionToken()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate verification token")
		http.Error(w, `{"message": "Failed to create verification token"}`, http.StatusInternalServerError)
		return
	}

	// Create verification token record
	tokenRecord := &VerificationToken{
		Identifier: user.Email,
		Token:      verificationToken,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24 hours
		TenantID:   tenantID,
	}

	if err := h.db.Create(tokenRecord).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create verification token record")
		http.Error(w, `{"message": "Failed to create verification token"}`, http.StatusInternalServerError)
		return
	}

	// Run after:signup hooks
	hookReq.UserID = user.ID
	if err := h.auth.hooks.Execute(r.Context(), "after:signup", hookReq); err != nil {
		h.logger.WithError(err).Warn("After signup hook failed")
	}

	// Return user (without password) and verification info
	user.Password = ""
	response := SignUpResponse{
		User:    user,
		Message: "User created successfully. Please check your email for verification instructions.",
	}

	if svc := h.auth.EmailService(); svc != nil {
		storageUser := &storage.User{Email: user.Email, VerificationToken: &verificationToken}
		if err := svc.SendVerificationEmail(storageUser, verificationToken); err != nil {
			h.logger.WithError(err).WithField("user_id", user.ID).Warn("Failed to send verification email")
		}
	} else {
		verificationURL := fmt.Sprintf("https://%s/verify-email?token=%s", r.Host, verificationToken)
		h.logger.WithField("user_id", user.ID).WithField("verification_url", verificationURL).Info("Verification email not sent (no email service); URL for development")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SignInRequest represents a sign-in request
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignInResponse represents a sign-in response
type SignInResponse struct {
	User    *User  `json:"user"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
	Trusted bool   `json:"trusted,omitempty"` // true when session is a trusted 30-day device session
}

// HandleSignIn handles user login
func (h *Handler) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	if !h.auth.IsEnabled("login") {
		http.Error(w, `{"message": "Login is disabled"}`, http.StatusServiceUnavailable)
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"message": "Email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Extract tenant context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, `{"message": "Tenant context is required"}`, http.StatusBadRequest)
		return
	}

	// Run before:signin hooks
	hookReq := &HookRequest{
		Email:     req.Email,
		TenantID:  tenantID,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Host:      r.Host,
		Headers:   h.extractHeaders(r),
	}

	if err := h.auth.hooks.Execute(r.Context(), "before:signin", hookReq); err != nil {
		h.logger.WithError(err).Warn("Signin hook failed")
		http.Error(w, `{"message": "`+err.Error()+`"}`, http.StatusForbidden)
		return
	}

	// Find user
	var user User
	if err := h.db.Where("tenant_id = ? AND email = ?", tenantID, req.Email).First(&user).Error; err != nil {
		http.Error(w, `{"message": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"message": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Check rememberDevices setting and trusted device token
	trustedToken := r.Header.Get("X-Trusted-Device-Token")
	rememberDevices := h.shouldRememberDevice(user.ID)

	var session *Session
	var err error

	if trustedToken != "" && rememberDevices {
		// Use trusted session path for 30-day expiry
		session, err = h.auth.sessions.GetOrCreateTrustedSession(h.db, user.ID, tenantID, r.RemoteAddr, r.UserAgent(), trustedToken)
	} else {
		// Standard session with default expiry
		session, err = h.auth.sessions.CreateSession(h.db, user.ID, tenantID, r.RemoteAddr, r.UserAgent())
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to create session")
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Generate token with role and permissions
	token, err := h.auth.GenerateTokenWithRole(user.ID, tenantID, user.Role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Set session cookie
	h.auth.sessions.SetSessionCookie(w, session.SessionToken)

	// Run after:signin hooks
	hookReq.UserID = user.ID
	if err := h.auth.hooks.Execute(r.Context(), "after:signin", hookReq); err != nil {
		h.logger.WithError(err).Warn("After signin hook failed")
	}

	// Return user (without password)
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignInResponse{
		User:    &user,
		Token:   token,
		Message: "Signed in successfully",
		Trusted: session.TrustedDeviceToken != "",
	})
}

// HandleSignOut handles user logout
func (h *Handler) HandleSignOut(w http.ResponseWriter, r *http.Request) {
	token := h.auth.sessions.GetSessionTokenFromRequest(r)
	if token != "" {
		if err := h.auth.sessions.InvalidateSession(h.db, token); err != nil {
			h.logger.WithError(err).Warn("Failed to invalidate session")
		}
	}

	// Clear session cookie
	h.auth.sessions.ClearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Signed out successfully",
	})
}

// HandleGetSession returns the current session
func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	if !h.auth.IsEnabled("session") {
		http.Error(w, `{"message": "Session validation is disabled"}`, http.StatusServiceUnavailable)
		return
	}

	token := h.auth.sessions.GetSessionTokenFromRequest(r)
	if token == "" {
		http.Error(w, `{"message": "No session found"}`, http.StatusUnauthorized)
		return
	}

	session, err := h.auth.sessions.ValidateSession(h.db, token)
	if err != nil {
		http.Error(w, `{"message": "Invalid or expired session"}`, http.StatusUnauthorized)
		return
	}

	// Get user
	var user User
	if err := h.db.First(&user, session.UserID).Error; err != nil {
		http.Error(w, `{"message": "User not found"}`, http.StatusInternalServerError)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":    user,
		"session": session,
	})
}

// HandleOAuthInit initiates OAuth flow
func (h *Handler) HandleOAuthInit(w http.ResponseWriter, r *http.Request) {
	if !h.auth.IsEnabled("oauth") {
		http.Error(w, `{"message": "OAuth is disabled"}`, http.StatusServiceUnavailable)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, `{"message": "Provider is required"}`, http.StatusBadRequest)
		return
	}

	// Generate state (should be stored and validated in callback)
	state, err := generateSessionToken()
	if err != nil {
		http.Error(w, `{"message": "Failed to generate state"}`, http.StatusInternalServerError)
		return
	}

	authURL, err := h.auth.oauth.GetAuthURL(provider, state)
	if err != nil {
		http.Error(w, `{"message": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Store state in cookie for validation
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   h.auth.config.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleOAuthCallback handles OAuth callback
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !h.auth.IsEnabled("oauth") {
		http.Error(w, `{"message": "OAuth is disabled"}`, http.StatusServiceUnavailable)
		return
	}

	// Get provider from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		http.Error(w, `{"message": "Invalid callback URL"}`, http.StatusBadRequest)
		return
	}
	provider := pathParts[len(pathParts)-1]

	// Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, `{"message": "Invalid state"}`, http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"message": "Authorization code not provided"}`, http.StatusBadRequest)
		return
	}

	oauthToken, err := h.auth.oauth.ExchangeCode(r.Context(), provider, code)
	if err != nil {
		h.logger.WithError(err).Error("Failed to exchange OAuth code")
		http.Error(w, `{"message": "Failed to complete OAuth"}`, http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := h.auth.oauth.GetUserInfo(r.Context(), provider, oauthToken)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user info")
		http.Error(w, `{"message": "Failed to get user info"}`, http.StatusInternalServerError)
		return
	}

	// Extract tenant context - but do NOT trust X-Tenant-ID header for new signups
	// This prevents a security vulnerability where an attacker could inject a tenant ID
	// via the X-Tenant-ID header to get assigned to an enterprise tenant
	tenantID := h.extractTenantIDForExistingUser(r, userInfo, provider)
	if tenantID == uuid.Nil {
	// No existing user found - this is a new signup, create a fresh tenant
	// We create a new tenant directly with the default "free" plan
	newTenant := &Tenant{
		ID:     uuid.New(),
		Name:   "Default Tenant",
		Status: "active",
	}
	if err := h.db.Create(newTenant).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create tenant for new OAuth user")
		http.Error(w, `{"message": "Failed to create tenant"}`, http.StatusInternalServerError)
		return
	}
	tenantID = newTenant.ID
	}

	// Find or create user
	user, err := h.auth.oauth.FindOrCreateUser(h.db, tenantID, userInfo)
	if err != nil {
		h.logger.WithError(err).Error("Failed to find or create user")
		http.Error(w, `{"message": "Failed to process user"}`, http.StatusInternalServerError)
		return
	}

	// Create session
	session, err := h.auth.sessions.CreateSession(h.db, user.ID, tenantID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		h.logger.WithError(err).Error("Failed to create session")
		http.Error(w, `{"message": "Failed to create session"}`, http.StatusInternalServerError)
		return
	}

	// Generate token with role and permissions
	_, err = h.auth.GenerateTokenWithRole(user.ID, tenantID, user.Role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		http.Error(w, `{"message": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Set session cookie
	h.auth.sessions.SetSessionCookie(w, session.SessionToken)

	// Redirect to frontend
	frontendURL := r.URL.Query().Get("redirect")
	if frontendURL == "" {
		frontendURL = "/"
	}
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

// Helper methods

func (h *Handler) extractTenantID(r *http.Request) uuid.UUID {
	// Check header first
	tenantHeader := r.Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		if id, err := uuid.Parse(tenantHeader); err == nil {
			return id
		}
	}

	// Check subdomain
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		subdomain := parts[0]
		// Skip reserved subdomains
		reserved := []string{"www", "api", "auth", "admin", "app", "staging", "dev"}
		for _, r := range reserved {
			if strings.EqualFold(subdomain, r) {
				return uuid.Nil
			}
		}
		// Look up active tenant by subdomain
		var tenant Tenant
		if err := h.db.Where("subdomain = ? AND status = ?", subdomain, "active").First(&tenant).Error; err == nil {
			return tenant.ID
		}
	}

	return uuid.Nil
}

// extractTenantIDForExistingUser safely extracts tenant ID for OAuth callbacks
// It NEVER allows X-Tenant-ID header to assign a new user to an existing tenant.
// Instead, it checks if the user already exists via OAuth provider+email, and if so,
// uses their existing tenant. If the user doesn't exist, it returns uuid.Nil so
// a fresh tenant will be created for them.
func (h *Handler) extractTenantIDForExistingUser(r *http.Request, userInfo *OAuthUserInfo, provider string) uuid.UUID {
	// First, check if user already exists with this OAuth provider
	var account Account
	err := h.db.Where("provider = ? AND provider_account_id = ?", provider, userInfo.ID).First(&account).Error
	if err == nil {
		// User exists via OAuth - return their existing tenant
		return account.TenantID
	}

	// Check if user exists by email
	var existingUser User
	err = h.db.Where("email = ?", userInfo.Email).First(&existingUser).Error
	if err == nil {
		// User exists with this email - return their tenant
		return existingUser.TenantID
	}

	// User doesn't exist - return uuid.Nil to signal a new tenant should be created
	return uuid.Nil
}

func (h *Handler) extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}
	return headers
}

// HandleCheckEmailAvailability checks if an email is available for registration
func (h *Handler) HandleCheckEmailAvailability(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, `{"message": "Email parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Extract tenant context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, `{"message": "Tenant context is required"}`, http.StatusBadRequest)
		return
	}

	// Check if user exists with this email
	var existingUser User
	err := h.db.Where("tenant_id = ? AND email = ?", tenantID, email).First(&existingUser).Error
	available := err == gorm.ErrRecordNotFound

	response := map[string]interface{}{
		"available": available,
		"email":     email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleCheckUsernameAvailability checks if a username is available for registration in real-time
func (h *Handler) HandleCheckUsernameAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, `{"message": "Username parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Extract tenant context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, `{"message": "Tenant context is required"}`, http.StatusBadRequest)
		return
	}

	// Check if user exists with this username
	var existingUser User
	err := h.db.Where("tenant_id = ? AND username = ?", tenantID, username).First(&existingUser).Error
	available := err == gorm.ErrRecordNotFound

	response := map[string]interface{}{
		"available": available,
		"username":  username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// HandleVerifyEmail handles email verification with a token
func (h *Handler) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"message": "Verification token is required"}`, http.StatusBadRequest)
		return
	}

	// Extract tenant context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, `{"message": "Tenant context is required"}`, http.StatusBadRequest)
		return
	}

	// Clean up expired verification tokens for this tenant
	if err := h.db.Where("tenant_id = ? AND expires_at < ?", tenantID, time.Now()).Delete(&VerificationToken{}).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to clean up expired verification tokens")
		// Don't fail the request for this
	}

	// Find the verification token in the database
	var verificationToken VerificationToken
	err := h.db.Where("token = ? AND tenant_id = ? AND expires_at > ?", token, tenantID, time.Now()).First(&verificationToken).Error
	if err == gorm.ErrRecordNotFound {
		http.Error(w, `{"message": "Invalid or expired verification token"}`, http.StatusBadRequest)
		return
	}
	if err != nil {
		h.logger.WithError(err).Error("Failed to find verification token")
		http.Error(w, `{"message": "Failed to verify token"}`, http.StatusInternalServerError)
		return
	}

	// Find and update the user
	var user User
	if err := h.db.Where("tenant_id = ? AND email = ?", tenantID, verificationToken.Identifier).First(&user).Error; err != nil {
		http.Error(w, `{"message": "User not found"}`, http.StatusNotFound)
		return
	}

	// Check if already verified
	if user.EmailVerified {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Email is already verified",
		})
		return
	}

	// Mark email as verified
	now := time.Now()
	user.EmailVerified = true
	user.VerifiedAt = &now

	if err := h.db.Save(&user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to update user verification status")
		http.Error(w, `{"message": "Failed to verify email"}`, http.StatusInternalServerError)
		return
	}

	// Delete the used verification token
	if err := h.db.Delete(&verificationToken).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to delete used verification token")
		// Don't fail the request for this
	}

	// Run after:emailVerified hooks
	hookReq := &HookRequest{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Host:      r.Host,
		Headers:   h.extractHeaders(r),
	}

	if err := h.auth.hooks.Execute(r.Context(), "after:emailVerified", hookReq); err != nil {
		h.logger.WithError(err).Warn("Email verified hook failed")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully! You can now log in to your account.",
	})
}

// ResendVerificationRequest represents a resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// HandleResendVerification handles resending verification emails
func (h *Handler) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, `{"message": "Email is required"}`, http.StatusBadRequest)
		return
	}

	// Extract tenant context
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		http.Error(w, `{"message": "Tenant context is required"}`, http.StatusBadRequest)
		return
	}

	// Find user by email
	var user User
	err := h.db.Where("tenant_id = ? AND email = ?", tenantID, req.Email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Don't leak whether user exists - always return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If that email address is registered and unverified, a verification email has been sent.",
		})
		return
	}
	if err != nil {
		h.logger.WithError(err).Error("Failed to find user for verification resend")
		http.Error(w, `{"message": "Failed to process request"}`, http.StatusInternalServerError)
		return
	}

	// Check if already verified
	if user.EmailVerified {
		// Don't leak verification status
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "If that email address is registered and unverified, a verification email has been sent.",
		})
		return
	}

	// Run before:resendVerification hooks
	hookReq := &HookRequest{
		UserID:    user.ID,
		TenantID:  tenantID,
		Email:     req.Email,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Host:      r.Host,
		Headers:   h.extractHeaders(r),
	}

	if err := h.auth.hooks.Execute(r.Context(), "before:resendVerification", hookReq); err != nil {
		h.logger.WithError(err).Warn("Resend verification hook failed")
		http.Error(w, `{"message": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	verificationToken, err := generateSessionToken()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate verification token")
		http.Error(w, `{"message": "Failed to generate verification token"}`, http.StatusInternalServerError)
		return
	}

	// Delete any existing verification tokens for this user
	if err := h.db.Where("identifier = ? AND tenant_id = ?", req.Email, tenantID).Delete(&VerificationToken{}).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to delete existing verification tokens")
		// Continue anyway
	}

	// Create new verification token record
	tokenRecord := &VerificationToken{
		Identifier: req.Email,
		Token:      verificationToken,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24 hours
		TenantID:   tenantID,
	}

	if err := h.db.Create(tokenRecord).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create verification token record")
		http.Error(w, `{"message": "Failed to create verification token"}`, http.StatusInternalServerError)
		return
	}

	if svc := h.auth.EmailService(); svc != nil {
		storageUser := &storage.User{Email: user.Email, VerificationToken: &verificationToken}
		if err := svc.SendVerificationEmail(storageUser, verificationToken); err != nil {
			h.logger.WithError(err).WithField("user_id", user.ID).Warn("Failed to send resend verification email")
		}
	} else {
		verificationURL := fmt.Sprintf("https://%s/verify-email?token=%s", r.Host, verificationToken)
		h.logger.WithField("user_id", user.ID).WithField("verification_url", verificationURL).Info("Resend verification email not sent (no email service); URL for development")
	}

	// Run after:resendVerification hooks
	if err := h.auth.hooks.Execute(r.Context(), "after:resendVerification", hookReq); err != nil {
		h.logger.WithError(err).Warn("After resend verification hook failed")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If that email address is registered and unverified, a verification email has been sent.",
	})
}

// shouldRememberDevice returns true if the user has enabled "remember trusted devices" in their settings.
func (h *Handler) shouldRememberDevice(userID uuid.UUID) bool {
	settings, err := h.userRepo.GetUserSettings(userID)
	if err != nil || settings == nil {
		return true // default to enabled
	}
	if val, ok := settings["rememberDevices"]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return true
}

// HandleTrustedDeviceRequest generates a trusted device token for the current user.
// POST /auth/trusted-device
// Returns a token the client stores as a cookie/localStorage and sends on future logins via X-Trusted-Device-Token header.
func (h *Handler) HandleTrustedDeviceRequest(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, `{"message": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	trustedToken, err := generateTrustedDeviceToken()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate trusted device token")
		http.Error(w, `{"message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trustedToken": trustedToken,
		"expiresIn":    "30d",
	})
}

// generateTrustedDeviceToken creates a cryptographically secure token for device trust.
func generateTrustedDeviceToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
