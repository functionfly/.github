package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleGetOAuthURL generates OAuth authorization URL for a provider.
// Optional query params:
//   - redirect_uri: e.g. http://127.0.0.1:port/callback for CLI — callback will redirect there with token
//   - invite_code: for invite-only signup validation
//   - login_hint: preserves tenant subdomain or email context through the OAuth flow
//     (e.g., if user was on tenant1.functionfly.com, this is stored and restored post-auth for redirect)
//   - tenant_id: tenant UUID to use per-tenant OAuth provider (optional, falls back to global)
func (h *Handler) HandleGetOAuthURL(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeJSONError(w, http.StatusBadRequest, "Provider is required")
		return
	}
	redirectURI := r.URL.Query().Get("redirect_uri")
	inviteCode := r.URL.Query().Get("invite_code")
	loginHint := r.URL.Query().Get("login_hint")
	tenantIDStr := r.URL.Query().Get("tenant_id")

	deviceFingerprint := generateDeviceFingerprint(r)

	// Parse tenant ID if provided
	var tenantID *uuid.UUID
	if tenantIDStr != "" {
		if parsed, err := uuid.Parse(tenantIDStr); err == nil {
			tenantID = &parsed
		}
	}

	url, err := h.authSvc.GetOAuthURL(provider, redirectURI, inviteCode, loginHint, deviceFingerprint, tenantID)
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
	startTime := time.Now()

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		if v := mux.Vars(r); v != nil {
			provider = v["provider"]
		}
	}

	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	if provider == "" {
		failureReason := "Provider is required"
		h.logOAuthAuthEvent(nil, nil, false, "oauth_login", clientIP, userAgent, &failureReason, time.Since(startTime), provider, map[string]interface{}{"error_phase": "validation"})
		writeJSONError(w, http.StatusBadRequest, failureReason)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		failureReason := "Authorization code is required"
		h.logOAuthAuthEvent(nil, nil, false, "oauth_login", clientIP, userAgent, &failureReason, time.Since(startTime), provider, map[string]interface{}{"error_phase": "validation"})
		writeJSONError(w, http.StatusBadRequest, failureReason)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		failureReason := "State parameter is required"
		h.logOAuthAuthEvent(nil, nil, false, "oauth_login", clientIP, userAgent, &failureReason, time.Since(startTime), provider, map[string]interface{}{"error_phase": "validation"})
		writeJSONError(w, http.StatusBadRequest, failureReason)
		return
	}

	// Generate device fingerprint from current request for session binding validation
	deviceFingerprint := generateDeviceFingerprint(r)

	response, err := h.authSvc.HandleOAuthCallback(provider, code, state, deviceFingerprint)
	if err != nil {
		logrus.WithError(err).WithField("provider", provider).Warn("OAuth callback failed")

		var errorCode auth.AuthCallbackErrorCode
		var errorDesc string

		var oauthErr *auth.OAuthError
		if errors.As(err, &oauthErr) {
			errorCode = mapOAuthErrorTypeToCode(oauthErr.Type)
			errorDesc = oauthErr.Description
		} else {
			errorCode = auth.AuthErrUnknown
			errorDesc = "An unexpected error occurred during authentication"
		}

		h.logOAuthAuthEvent(nil, nil, false, "oauth_login", clientIP, userAgent, &errorDesc, time.Since(startTime), provider, map[string]interface{}{
			"error_code":  string(errorCode),
			"error_type":  err.Error(),
			"error_phase": "callback_processing",
		})

		result := &auth.AuthCallbackResult{
			Success:   false,
			Error:     string(errorCode),
			ErrorDesc: errorDesc,
		}
		http.Redirect(w, r, buildOAuthRedirectURL(getOAuthFrontendURL(), result), http.StatusFound)
		return
	}

	h.logOAuthAuthEvent(&response.User.ID, &response.User.TenantID, true, "oauth_login", clientIP, userAgent, nil, time.Since(startTime), provider, map[string]interface{}{
		"new_user":       response.NewUser,
		"redirect_uri":   response.RedirectURI,
		"login_hint":     response.LoginHint,
		"refresh_issued": response.RefreshToken != "",
	})

	result := &auth.AuthCallbackResult{
		Success:      true,
		Token:        response.Token,
		RefreshToken: response.RefreshToken,
		NewUser:      response.NewUser,
	}

	if response.RedirectURI != "" {
		http.Redirect(w, r, buildOAuthRedirectURL(response.RedirectURI, result), http.StatusFound)
		return
	}

	redirectURL := getOAuthFrontendURL()
	if response.LoginHint != "" {
		redirectURL = buildTenantRedirectURL(response.LoginHint)
	}
	http.Redirect(w, r, buildOAuthRedirectURL(redirectURL, result), http.StatusFound)
}

// HandleGetOAuthProviders returns list of configured OAuth providers.
// Optional query params:
//   - tenant_id: tenant UUID to get per-tenant OAuth providers (merges with global if both exist)
func (h *Handler) HandleGetOAuthProviders(w http.ResponseWriter, r *http.Request) {
	// Get global providers
	globalProviders := h.authSvc.GetConfiguredOAuthProviders()
	providerSet := make(map[string]bool)
	for _, p := range globalProviders {
		providerSet[p] = true
	}

	// Check for per-tenant providers
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			tenantProviders, err := h.authSvc.GetTenantOAuthProviders(r.Context(), tenantID)
			if err == nil {
				for _, p := range tenantProviders {
					providerSet[p] = true
				}
			}
		}
	}

	// Convert set to list
	providers := make([]string, 0, len(providerSet))
	for p := range providerSet {
		providers = append(providers, p)
	}

	response := map[string]interface{}{
		"providers": providers,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleConfirmAccountLinking confirms linking a social account to an existing user account.
func (h *Handler) HandleConfirmAccountLinking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		LinkToken  string `json:"link_token"`
		Provider   string `json:"provider"`
		ProviderID string `json:"provider_id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		AvatarURL  string `json:"avatar_url"`
		Confirm    bool   `json:"confirm"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !req.Confirm {
		writeJSONError(w, http.StatusBadRequest, "User must confirm account linking")
		return
	}

	if req.LinkToken == "" || req.Provider == "" || req.ProviderID == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing required fields: link_token, provider, provider_id")
		return
	}

	userInfo := &auth.OAuthUserInfo{
		Email:     req.Email,
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	}

	response, err := h.authSvc.ConfirmAccountLinking(req.LinkToken, req.Provider, req.ProviderID, userInfo)
	if err != nil {
		logrus.WithError(err).Warn("Account linking confirmation failed")

		var oauthErr *auth.OAuthError
		if errors.As(err, &oauthErr) {
			writeJSONError(w, http.StatusBadRequest, oauthErr.Description)
		} else {
			writeJSONError(w, http.StatusInternalServerError, "Failed to link account")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth.LoginResponse{
		Token:        response.Token,
		RefreshToken: response.RefreshToken,
		User: &auth.LoginUser{
			ID:       response.User.ID.String(),
			TenantID: response.User.TenantID.String(),
			Email:    response.User.Email,
			Role:     response.User.Role,
		},
	})
}

// getOAuthFrontendURL returns the frontend URL for OAuth redirects
func getOAuthFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return frontendURL + "/auth/oauth/callback"
}

// buildOAuthRedirectURL builds the redirect URL for OAuth callbacks with unified result structure
func buildOAuthRedirectURL(baseURL string, result *auth.AuthCallbackResult) string {
	u, _ := url.Parse(baseURL)
	q := u.Query()

	if result.Success {
		q.Set("token", result.Token)
		q.Set("refresh_token", result.RefreshToken)
		q.Set("new_user", fmt.Sprintf("%t", result.NewUser))
	} else {
		q.Set("error", result.Error)
		if result.ErrorDesc != "" {
			q.Set("error_description", url.QueryEscape(result.ErrorDesc))
		}
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// mapOAuthErrorTypeToCode maps OAuthError.Type to standardized AuthCallbackErrorCode
func mapOAuthErrorTypeToCode(oauthErrType string) auth.AuthCallbackErrorCode {
	switch oauthErrType {
	case "invalid_state":
		return auth.AuthErrInvalidState
	case "invalid_provider":
		return auth.AuthErrProviderNotConfigured
	case "token_exchange_failed":
		return auth.AuthErrTokenExchangeFailed
	case "user_info_failed":
		return auth.AuthErrUserInfoFailed
	case "missing_email":
		return auth.AuthErrMissingEmail
	case "database_error":
		return auth.AuthErrDatabaseError
	case "account_link_failed":
		return auth.AuthErrAccountLinkFailed
	case "invalid_invite", "invite_required":
		return auth.AuthErrInvalidInvite
	case "tenant_creation_failed":
		return auth.AuthErrTenantCreationFailed
	case "user_creation_failed":
		return auth.AuthErrUserCreationFailed
	case "token_generation_failed":
		return auth.AuthErrTokenGenerationFailed
	case "token_storage_failed":
		return auth.AuthErrTokenStorageFailed
	default:
		return auth.AuthErrUnknown
	}
}

// buildTenantRedirectURL builds a redirect URL for tenant subdomain based on login_hint.
func buildTenantRedirectURL(loginHint string) string {
	if loginHint == "" {
		return getOAuthFrontendURL()
	}

	baseFrontend := os.Getenv("FRONTEND_URL")
	if baseFrontend == "" {
		baseFrontend = "http://localhost:3000"
	}

	baseURL, err := url.Parse(baseFrontend)
	if err != nil {
		return getOAuthFrontendURL()
	}

	host := baseURL.Hostname()
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		baseDomain := strings.Join(parts[len(parts)-2:], ".")
		tenantHost := loginHint + "." + baseDomain
		tenantURL := &url.URL{
			Scheme: baseURL.Scheme,
			Host:   tenantHost,
			Path:   "/auth/oauth/callback",
		}
		return tenantURL.String()
	}

	tenantURL := &url.URL{
		Scheme: baseURL.Scheme,
		Host:   loginHint + "." + host,
		Path:   "/auth/oauth/callback",
	}
	return tenantURL.String()
}

// logOAuthAuthEvent logs an OAuth authentication event for security auditing.
func (h *Handler) logOAuthAuthEvent(userID, tenantID *uuid.UUID, success bool, eventType, clientIP, userAgent string, failureReason *string, duration time.Duration, provider string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["duration_ms"] = duration.Milliseconds()
	metadata["provider"] = provider

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

	if logErr := h.authSvc.Repo().LogAuthEvent(authEvent); logErr != nil {
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
		logrus.WithError(logErr).WithFields(fields).Warn("Failed to log OAuth auth event")
	}
}

// generateDeviceFingerprint creates a device fingerprint from request characteristics.
func generateDeviceFingerprint(r *http.Request) string {
	components := []string{
		getClientIP(r),
		r.Header.Get("User-Agent"),
		r.Header.Get("Accept-Language"),
		r.Header.Get("Accept-Encoding"),
	}

	hashInput := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(hashInput))
	return base64.URLEncoding.EncodeToString(hash[:])
}
