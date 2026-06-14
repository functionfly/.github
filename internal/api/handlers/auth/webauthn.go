package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WebAuthnHandler handles WebAuthn/Passkey-related API endpoints
type WebAuthnHandler struct {
	webAuthnSvc  *auth.WebAuthnService
	authSvc      *auth.AuthService
	sessionStore *auth.WebAuthnSessionStore
	repo         storage.Repository
	logger       *logrus.Logger
}

// NewWebAuthnHandler creates a new WebAuthn handler
func NewWebAuthnHandler(webAuthnSvc *auth.WebAuthnService, authSvc *auth.AuthService, sessionStore *auth.WebAuthnSessionStore, repo storage.Repository) *WebAuthnHandler {
	return &WebAuthnHandler{
		webAuthnSvc:  webAuthnSvc,
		authSvc:      authSvc,
		sessionStore: sessionStore,
		repo:         repo,
		logger:       logrus.New(),
	}
}

// HandleWebAuthnRegisterBegin starts the WebAuthn registration ceremony
// POST /auth/webauthn/register/begin
func (h *WebAuthnHandler) HandleWebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// For now, use email as display name
	displayName := claims.Email

	// Begin registration
	options, sessionData, err := h.webAuthnSvc.BeginRegistration(claims.UserID, displayName, claims.Email)
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin WebAuthn registration")
		http.Error(w, "Failed to begin registration", http.StatusInternalServerError)
		return
	}

	// Store session in Redis
	ctx := context.Background()
	session := &auth.WebAuthnSession{
		Challenge:   "", // Will be extracted from sessionData
		UserHandle:  claims.UserID.String(),
		UserID:      claims.UserID.String(),
		Operation:   "registration",
		SessionData: sessionData,
	}

	// Extract challenge from sessionData for reference
	var sessionDataMap map[string]interface{}
	if err := json.Unmarshal(sessionData, &sessionDataMap); err == nil {
		if challenge, ok := sessionDataMap["challenge"].(string); ok {
			session.Challenge = challenge
		}
	}

	sessionID, err := h.sessionStore.Create(ctx, session)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create WebAuthn session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Return the options and session ID to the client
	response := map[string]interface{}{
		"options":   json.RawMessage(options),
		"sessionID": sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleWebAuthnRegisterComplete completes the WebAuthn registration ceremony
// POST /auth/webauthn/register/complete
func (h *WebAuthnHandler) HandleWebAuthnRegisterComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req struct {
		SessionID string          `json:"sessionId"`
		Response  json.RawMessage `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Retrieve session from Redis
	ctx := context.Background()
	session, err := h.sessionStore.Get(ctx, req.SessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to retrieve WebAuthn session")
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "Session not found or expired", http.StatusBadRequest)
		return
	}

	// Verify the session belongs to this user
	if session.UserID != claims.UserID.String() {
		h.logger.Warn("Session user ID mismatch")
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	// Verify this is a registration session
	if session.Operation != "registration" {
		http.Error(w, "Invalid session operation", http.StatusBadRequest)
		return
	}

	// Create a fake http.Request with the response data for the WebAuthn library
	// The library expects the response in the request body
	r.Body = http.NoBody
	// We need to reconstruct the request with the response data
	// The webauthn library reads from r.Form and r.PostForm for the attestation response

	// Instead of using the http.Request directly, we'll use a helper to parse the response
	// The FinishRegistration expects the response in a specific format

	// Call FinishRegistration with the stored session data
	credential, err := h.webAuthnSvc.FinishRegistration(claims.UserID, session.SessionData, req.Response)
	if err != nil {
		h.logger.WithError(err).Error("Failed to complete WebAuthn registration")
		// Delete the session
		h.sessionStore.Delete(ctx, req.SessionID)
		http.Error(w, "Failed to complete registration: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Delete the session after use
	h.sessionStore.Delete(ctx, req.SessionID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "Registration successful",
		"credential": credential.ID.String(),
	})
}

// HandleWebAuthnLoginBegin starts the WebAuthn login ceremony
// POST /auth/webauthn/login/begin
func (h *WebAuthnHandler) HandleWebAuthnLoginBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request to get user ID (or get from authenticated user for re-auth)
	var req struct {
		UserID *string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var userID uuid.UUID
	var err error

	if req.UserID != nil {
		// If user ID provided, use it (for initial login)
		userID, err = uuid.Parse(*req.UserID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
	} else {
		// Get user from context (set by auth middleware for re-authentication)
		claims := middleware.GetUserFromContext(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID = claims.UserID
	}

	// Begin login
	options, sessionData, err := h.webAuthnSvc.BeginLogin(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to begin WebAuthn login")
		http.Error(w, "Failed to begin login: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store session in Redis
	ctx := context.Background()
	session := &auth.WebAuthnSession{
		Challenge:   "", // Will be extracted from sessionData
		UserHandle:  userID.String(),
		UserID:      userID.String(),
		Operation:   "authentication",
		SessionData: sessionData,
	}

	// Extract challenge from sessionData for reference
	var sessionDataMap map[string]interface{}
	if err := json.Unmarshal(sessionData, &sessionDataMap); err == nil {
		if challenge, ok := sessionDataMap["challenge"].(string); ok {
			session.Challenge = challenge
		}
	}

	sessionID, err := h.sessionStore.Create(ctx, session)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create WebAuthn session")
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Return the options and session ID to the client
	response := map[string]interface{}{
		"options":   json.RawMessage(options),
		"sessionID": sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleWebAuthnLoginComplete completes the WebAuthn login ceremony
// POST /auth/webauthn/login/complete
func (h *WebAuthnHandler) HandleWebAuthnLoginComplete(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	ctx := r.Context()

	if r.Method != http.MethodPost {
		failureReason := "Method not allowed"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "validation"})
		http.Error(w, failureReason, http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		SessionID string          `json:"sessionId"`
		Response  json.RawMessage `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		failureReason := "Invalid request body"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "validation"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		failureReason := "Session ID required"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "validation"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	// Retrieve session from Redis
	ctx = context.Background()
	session, err := h.sessionStore.Get(ctx, req.SessionID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to retrieve WebAuthn session")
		failureReason := "Failed to retrieve session"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "session_retrieval"})
		http.Error(w, failureReason, http.StatusInternalServerError)
		return
	}

	if session == nil {
		failureReason := "Session not found or expired"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "session_validation"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	// Verify this is a login session
	if session.Operation != "authentication" {
		failureReason := "Invalid session operation"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "session_validation"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	// Parse user ID from session
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse user ID from session")
		failureReason := "Invalid session"
		h.logWebAuthnAuthEvent(ctx, nil, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "session_validation"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	// Complete login
	success, err := h.webAuthnSvc.FinishLogin(userID, session.SessionData, req.Response)
	if err != nil {
		h.logger.WithError(err).Error("Failed to complete WebAuthn login")
		// Delete the session
		h.sessionStore.Delete(ctx, req.SessionID)
		failureReason := "Failed to complete login: " + err.Error()
		h.logWebAuthnAuthEvent(ctx, &userID, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "authentication"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	if !success {
		h.logger.Warn("WebAuthn login failed verification")
		h.sessionStore.Delete(ctx, req.SessionID)
		failureReason := "Login verification failed"
		h.logWebAuthnAuthEvent(ctx, &userID, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "verification"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	// Delete the session after use
	h.sessionStore.Delete(ctx, req.SessionID)

	// Get user from database to generate JWT tokens
	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user after WebAuthn login")
		failureReason := "Failed to get user information"
		h.logWebAuthnAuthEvent(ctx, &userID, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "user_retrieval"})
		http.Error(w, failureReason, http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.logger.Error("User not found after WebAuthn login")
		failureReason := "User not found"
		h.logWebAuthnAuthEvent(ctx, &userID, nil, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "user_retrieval"})
		http.Error(w, failureReason, http.StatusInternalServerError)
		return
	}

	// Check if email is verified
	if !user.EmailVerified {
		failureReason := "Email not verified"
		h.logWebAuthnAuthEvent(ctx, &userID, &user.TenantID, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "verification"})
		http.Error(w, failureReason, http.StatusBadRequest)
		return
	}

	// Generate JWT token
	token, err := h.authSvc.GenerateToken(user)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token after WebAuthn login")
		failureReason := "Failed to generate token"
		h.logWebAuthnAuthEvent(ctx, &userID, &user.TenantID, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "token_generation"})
		http.Error(w, failureReason, http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, refreshTokenHash, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate refresh token after WebAuthn login")
		failureReason := "Failed to generate refresh token"
		h.logWebAuthnAuthEvent(ctx, &userID, &user.TenantID, false, "webauthn_login", clientIP, userAgent, &failureReason, time.Since(startTime), map[string]interface{}{"error_phase": "token_generation"})
		http.Error(w, failureReason, http.StatusInternalServerError)
		return
	}

	// Store refresh token in database (expires in 30 days)
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = h.repo.CreateRefreshToken(r.Context(), user.ID, refreshTokenHash, "webauthn", "webauthn-callback", refreshExpiresAt)
	if err != nil {
		h.logger.WithError(err).WithField("userID", user.ID).Warn("Failed to store WebAuthn refresh token")
		// Continue without refresh token - access token is still valid
		refreshToken = ""
	}

	// Log successful WebAuthn login with comprehensive metadata
	h.logWebAuthnAuthEvent(ctx, &user.ID, &user.TenantID, true, "webauthn_login", clientIP, userAgent, nil, time.Since(startTime), map[string]interface{}{
		"refresh_issued": refreshToken != "",
	})

	// Build login user response (handling pointer fields from storage.User)
	loginUser := &auth.LoginUser{
		ID:            user.ID.String(),
		TenantID:      user.TenantID.String(),
		Email:         user.Email,
		Role:          user.Role,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Name:          user.Name, // Name is not a pointer
	}
	// Handle pointer fields
	if user.Username != nil && *user.Username != "" {
		loginUser.Username = *user.Username
	} else if user.Email != "" {
		// Fallback: use email local part
		if at := strings.Index(user.Email, "@"); at > 0 {
			loginUser.Username = user.Email[:at]
		} else {
			loginUser.Username = user.Email
		}
	}
	if user.CompanyName != nil {
		loginUser.CompanyName = *user.CompanyName
	}

	// Return success response with tokens
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         loginUser,
	})
}

// HandleListWebAuthnCredentials returns all credentials for the authenticated user
// GET /auth/webauthn/credentials
func (h *WebAuthnHandler) HandleListWebAuthnCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get credentials
	credentials, err := h.webAuthnSvc.GetCredentialsForUser(claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get WebAuthn credentials")
		http.Error(w, "Failed to get credentials", http.StatusInternalServerError)
		return
	}

	// Convert to response format (hide sensitive data)
	type CredentialResponse struct {
		ID             string  `json:"id"`
		CreatedAt      string  `json:"createdAt"`
		LastUsedAt     *string `json:"lastUsedAt,omitempty"`
		BackupEligible bool    `json:"backupEligible"`
	}

	credResponses := make([]CredentialResponse, len(credentials))
	for i, cred := range credentials {
		credResponses[i] = CredentialResponse{
			ID:             cred.ID.String(),
			CreatedAt:      cred.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastUsedAt:     nil,
			BackupEligible: cred.BackupEligible,
		}
		if cred.LastUsedAt != nil {
			lastUsed := cred.LastUsedAt.Format("2006-01-02T15:04:05Z07:00")
			credResponses[i].LastUsedAt = &lastUsed
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"credentials": credResponses,
	})
}

// HandleDeleteWebAuthnCredential deletes a credential
// DELETE /auth/webauthn/credentials/{id}
func (h *WebAuthnHandler) HandleDeleteWebAuthnCredential(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get credential ID from URL
	credentialIDStr := r.URL.Query().Get("id")
	if credentialIDStr == "" {
		http.Error(w, "Credential ID required", http.StatusBadRequest)
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		http.Error(w, "Invalid credential ID", http.StatusBadRequest)
		return
	}

	// Get credentials to verify ownership
	credentials, err := h.webAuthnSvc.GetCredentialsForUser(claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get WebAuthn credentials")
		http.Error(w, "Failed to verify credentials", http.StatusInternalServerError)
		return
	}

	// Verify the credential belongs to the user
	belongsToUser := false
	for _, cred := range credentials {
		if cred.ID == credentialID {
			belongsToUser = true
			break
		}
	}

	if !belongsToUser {
		http.Error(w, "Credential not found", http.StatusNotFound)
		return
	}

	// Delete the credential
	if err := h.webAuthnSvc.DeleteCredential(credentialID); err != nil {
		h.logger.WithError(err).Error("Failed to delete WebAuthn credential")
		http.Error(w, "Failed to delete credential", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Credential deleted successfully",
	})
}

// logWebAuthnAuthEvent logs a WebAuthn authentication event for security auditing.
// Records success/failure, IP + user agent, user ID (on success),
// failure reason (on failure), and time taken (for latency monitoring).
func (h *WebAuthnHandler) logWebAuthnAuthEvent(ctx context.Context, userID, tenantID *uuid.UUID, success bool, eventType, clientIP, userAgent string, failureReason *string, duration time.Duration, metadata map[string]interface{}) {
	// Add latency information to metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["duration_ms"] = duration.Milliseconds()

	provider := "webauthn"
	authEvent := &storage.AuthEvent{
		UserID:    userID,
		TenantID:  tenantID,
		EventType: eventType,
		Success:   success,
		IPAddress: clientIP,
		UserAgent: userAgent,
		Provider:  &provider,
		Metadata:  metadata,
	}

	if failureReason != nil {
		authEvent.FailureReason = failureReason
	}

	if logErr := h.repo.LogAuthEvent(ctx, authEvent); logErr != nil {
		fields := logrus.Fields{
			"event_type":  eventType,
			"success":     success,
			"duration_ms": duration.Milliseconds(),
			"provider":    provider,
		}
		if userID != nil {
			fields["user_id"] = userID.String()
		}
		if failureReason != nil {
			fields["failure_reason"] = *failureReason
		}
		h.logger.WithError(logErr).WithFields(fields).Warn("Failed to log WebAuthn auth event")
	}
}
