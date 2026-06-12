package auth

import (
	"context"
	"crypto/rand"
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

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// allowedCallbackHosts defines the allowlist for valid redirect URI hosts.
// Includes localhost for CLI flows and production domains for production.
var allowedCallbackHosts = []string{
	"localhost",
	"127.0.0.1",
	"app.functionfly.com",     // production dashboard
	"staging.functionfly.com", // staging dashboard
}

// allowedCallbackPaths defines the allowlist for valid callback paths per host
var allowedCallbackPaths = map[string][]string{
	"app.functionfly.com":     {"/auth/callback", "/auth/oauth/callback", "/callback", "/oauth/callback"},
	"staging.functionfly.com": {"/auth/callback", "/auth/oauth/callback", "/callback", "/oauth/callback"},
	"localhost":               {"/auth/callback", "/auth/oauth/callback", "/callback", "/oauth/callback", "/"},
	"127.0.0.1":              {"/auth/callback", "/auth/oauth/callback", "/callback", "/oauth/callback", "/"},
}

// IsAllowedRedirectURI validates redirect_uri against a strict allowlist.
// For production, validates against known hosts AND paths. For CLI flows, allows localhost/127.0.0.1.
func IsAllowedRedirectURI(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}

	// Only allow https for production domains, http for localhost
	isHTTPS := u.Scheme == "https"
	isHTTP := u.Scheme == "http"
	if !isHTTPS && !isHTTP {
		return false
	}

	host := strings.ToLower(u.Hostname())
	path := u.Path

	// Check if host is in allowlist
	hostAllowed := false
	isLocalhost := false
	for _, allowed := range allowedCallbackHosts {
		if host == allowed {
			hostAllowed = true
			isLocalhost = (allowed == "localhost" || allowed == "127.0.0.1")
			break
		}
		// Subdomain match - ensure it's a proper subdomain boundary (dot prefix)
		if strings.HasSuffix(host, "."+allowed) {
			hostAllowed = true
			break
		}
	}

	if !hostAllowed {
		return false
	}

	// Require HTTPS for production domains, allow HTTP for localhost
	if !isLocalhost && !isHTTPS {
		return false
	}

	// For production domains, validate the path
	if !isLocalhost {
		allowedPaths, ok := allowedCallbackPaths[host]
		if !ok {
			// Unknown production host - deny
			return false
		}
		pathAllowed := false
		for _, allowedPath := range allowedPaths {
			if path == allowedPath {
				pathAllowed = true
				break
			}
		}
		if !pathAllowed {
			return false
		}
	}

	return true
}

// isAllowedRedirectURI is an alias for IsAllowedRedirectURI for internal use.
// Kept for backward compatibility within the auth package.
func isAllowedRedirectURI(raw string) bool {
	return IsAllowedRedirectURI(raw)
}

// PKCE (Proof Key for Code Exchange) helpers
// PKCE is required/recommended for public clients to prevent authorization code interception attacks.

// generateCodeVerifier generates a cryptographically secure PKCE code verifier.
// The code verifier is a high-entropy random string using unreserved URL characters.
// PKCE spec (RFC 7636) requires 43-128 characters of unreserved URL-safe chars.
// We use 32 bytes (256 bits) of entropy which encodes to ~43 base64url chars.
func generateCodeVerifier() (string, error) {
	// PKCE spec requires minimum 43 chars, we use 256 bits of entropy
	b := make([]byte, 32) // 32 bytes = 256 bits of entropy
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate code verifier: %w", err)
	}
	// Base64URL encode without padding (PKCE spec requirement)
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge generates the S256 code challenge from a code verifier.
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// SetBaseURL sets the base URL for OAuth redirects
func (a *AuthService) SetBaseURL(baseURL string) {
	a.baseURL = baseURL
	a.initOAuthProviders() // Re-initialize with new base URL
}

// initOAuthProviders initializes OAuth provider configurations
func (a *AuthService) initOAuthProviders() {
	// Google OAuth
	if googleClientID := a.getEnvVar("GOOGLE_CLIENT_ID"); googleClientID != "" {
		if googleClientSecret := a.getEnvVar("GOOGLE_CLIENT_SECRET"); googleClientSecret != "" {
			a.oauthProviders["google"] = &OAuthProvider{
				Name: "google",
				Config: &oauth2.Config{
					ClientID:     googleClientID,
					ClientSecret: googleClientSecret,
					RedirectURL:  a.baseURL + "/v1/auth/oauth/google/callback",
					Scopes:       []string{"openid", "profile", "email"},
					Endpoint:     google.Endpoint,
				},
				UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
				Scopes:      []string{"openid", "profile", "email"},
			}
		}
	}

	// GitHub OAuth
	if githubClientID := a.getEnvVar("GITHUB_CLIENT_ID"); githubClientID != "" {
		if githubClientSecret := a.getEnvVar("GITHUB_CLIENT_SECRET"); githubClientSecret != "" {
			a.oauthProviders["github"] = &OAuthProvider{
				Name: "github",
				Config: &oauth2.Config{
					ClientID:     githubClientID,
					ClientSecret: githubClientSecret,
					RedirectURL:  a.baseURL + "/v1/auth/oauth/github/callback",
					Scopes:       []string{"read:user", "user:email"},
					Endpoint:     github.Endpoint,
				},
				UserInfoURL: "https://api.github.com/user",
				Scopes:      []string{"read:user", "user:email"},
			}
		}
	}
}

// getEnvVar is a helper to get environment variables
func (a *AuthService) getEnvVar(key string) string {
	// Read from environment variables
	// This allows OAuth providers to be configured via environment
	return os.Getenv(key)
}

// getTenantOAuthProvider retrieves a decrypted OAuth provider configuration for a tenant.
// It looks up the provider in the database, decrypts the client secret, and returns
// an OAuthProvider ready for use in the OAuth flow.
func (a *AuthService) getTenantOAuthProvider(ctx context.Context, provider string, tenantID uuid.UUID) (*OAuthProvider, error) {
	// Get the tenant provider configuration from DB
	tenantProvider, err := a.repo.GetOAuthProvider(ctx, tenantID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth provider from DB: %w", err)
	}
	if tenantProvider == nil {
		return nil, nil // Not found, will fall back to global provider
	}
	if !tenantProvider.Enabled {
		return nil, fmt.Errorf("OAuth provider '%s' is disabled for this tenant", provider)
	}

	// Decrypt the client secret
	clientSecret, err := a.repo.DecryptField(tenantProvider.EncryptedClientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	// Build the oauth2 config for this provider
	// Note: We need to set the correct endpoint based on the provider
	endpoint := a.getOAuthEndpoint(provider)
	callbackURL := a.baseURL + "/v1/auth/oauth/" + provider + "/callback"

	// Parse scopes - default to common scopes if not stored
	scopes := []string{"openid", "profile", "email"}
	if len(tenantProvider.Scopes) > 0 {
		if err := json.Unmarshal(tenantProvider.Scopes, &scopes); err != nil {
			logrus.WithError(err).Warn("Failed to parse OAuth scopes, using defaults")
		}
	}

	oauth2Config := &oauth2.Config{
		ClientID:     tenantProvider.ClientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Scopes:       scopes,
		Endpoint:     endpoint,
	}

	return &OAuthProvider{
		Name:        provider,
		Config:      oauth2Config,
		UserInfoURL: a.getUserInfoURL(provider),
		Scopes:      scopes,
	}, nil
}

// getOAuthEndpoint returns the correct OAuth2 endpoint for a provider
func (a *AuthService) getOAuthEndpoint(provider string) oauth2.Endpoint {
	switch provider {
	case "google":
		return google.Endpoint
	case "github":
		return github.Endpoint
	default:
		// Default to Google for unknown providers
		return google.Endpoint
	}
}

// getUserInfoURL returns the user info endpoint for a provider
func (a *AuthService) getUserInfoURL(provider string) string {
	switch provider {
	case "google":
		return "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		return "https://api.github.com/user"
	case "microsoft":
		return "https://graph.microsoft.com/v1.0/me"
	default:
		return ""
	}
}

// GetOAuthURL generates OAuth authorization URL for a provider.
// redirectURI is optional; when set (e.g. CLI callback URL), the callback will redirect there with the token.
// Only http://127.0.0.1 and http://localhost (any port) are allowed for redirectURI.
// inviteCode is optional unless SignupInviteRequired(); then it must be valid (read-only check) and is stored in OAuth state for callback redemption.
// loginHint is optional; preserves tenant subdomain or email context through the OAuth flow for better UX (redirect back to origin post-auth).
// deviceFingerprint is optional; stores a hash of device characteristics for session binding validation on callback (prevents session fixation attacks).
// tenantID is optional; when provided, checks for per-tenant OAuth provider configuration first.
func (a *AuthService) GetOAuthURL(provider, redirectURI, inviteCode, loginHint, deviceFingerprint string, tenantID *uuid.UUID) (string, error) {
	var p *OAuthProvider
	var err error

	// Check for per-tenant OAuth provider first if tenantID is provided
	if tenantID != nil && *tenantID != uuid.Nil {
		p, err = a.getTenantOAuthProvider(context.Background(), provider, *tenantID)
		if err != nil {
			return "", fmt.Errorf("failed to get tenant OAuth provider: %w", err)
		}
	}

	// Fall back to global OAuth provider if not found in tenant
	if p == nil {
		var exists bool
		p, exists = a.oauthProviders[provider]
		if !exists {
			return "", fmt.Errorf("OAuth provider '%s' not configured", provider)
		}
	}

	if redirectURI != "" && !isAllowedRedirectURI(redirectURI) {
		return "", fmt.Errorf("redirect_uri must be http://127.0.0.1 or http://localhost (with optional port)")
	}

	ic := strings.TrimSpace(inviteCode)
	if SignupInviteRequired() {
		if ic == "" {
			return "", fmt.Errorf("invite code is required to sign up with OAuth")
		}
		if err := a.repo.ValidateSignupInviteReadOnly(context.Background(), ic); err != nil {
			if errors.Is(err, storage.ErrSignupInviteExhausted) {
				return "", fmt.Errorf("this invite code has no uses remaining")
			}
			if errors.Is(err, storage.ErrSignupInviteRevoked) {
				return "", fmt.Errorf("this invite code is no longer valid")
			}
			return "", fmt.Errorf("invalid or expired invite code")
		}
	}

	// Generate state for CSRF protection
	state, err := a.generateVerificationToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate OAuth state: %w", err)
	}

	// Generate PKCE code verifier and challenge (S256 method)
	// PKCE is required/recommended for public clients to prevent authorization code interception attacks
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE code verifier: %w", err)
	}
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Store state server-side (persisted for multi-instance) so we can validate it on callback
	// Store code_verifier for PKCE validation on callback
	// Store login_hint for redirecting back to tenant subdomain post-auth
	// Store device_fingerprint for session binding validation (prevents session fixation attacks)
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := a.repo.StoreOAuthState(context.Background(), state, expiresAt, redirectURI, ic, codeVerifier, loginHint, deviceFingerprint); err != nil {
		return "", fmt.Errorf("failed to store OAuth state: %w", err)
	}

	// Build auth URL with PKCE parameters
	authOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}

	return p.Config.AuthCodeURL(state, authOpts...), nil
}

// HandleOAuthCallback processes OAuth callback and returns user authentication.
// The returned RedirectURI (if set) should be used to redirect the user with the token (e.g. CLI callback).
// The returned LoginHint can be used to redirect back to tenant subdomain post-auth.
// deviceFingerprint must match the fingerprint stored when the OAuth flow was initiated.
func (a *AuthService) HandleOAuthCallback(provider, code, state, deviceFingerprint string) (*OAuthCallbackResponse, error) {
	// Validate CSRF state token — must match one issued by GetOAuthURL (persisted, one-time use)
	if state == "" {
		return nil, &OAuthError{
			Type:        "invalid_state",
			Message:     "Invalid or expired OAuth state",
			Description: "The OAuth state parameter is invalid or has expired. Please try signing in again.",
		}
	}
	valid, redirectURI, storedInvite, codeVerifier, loginHint, storedDeviceFingerprint, err := a.repo.ValidateAndConsumeOAuthState(context.Background(), state)
	if err != nil {
		return nil, &OAuthError{
			Type:        "invalid_state",
			Message:     "Invalid or expired OAuth state",
			Description: "The OAuth state could not be validated. Please try signing in again.",
		}
	}
	if !valid {
		return nil, &OAuthError{
			Type:        "invalid_state",
			Message:     "Invalid or expired OAuth state",
			Description: "The OAuth state parameter is invalid or has expired. Please try signing in again.",
		}
	}

	// Validate device fingerprint if stored (session binding - prevents session fixation attacks)
	// If device fingerprint was stored in state, the callback must come from the same device
	if storedDeviceFingerprint != "" {
		if deviceFingerprint == "" {
			return nil, &OAuthError{
				Type:        "device_verification_required",
				Message:     "Device verification required",
				Description: "This authentication requires device fingerprint verification. Please initiate the login from the same device.",
			}
		}
		if deviceFingerprint != storedDeviceFingerprint {
			logrus.WithFields(logrus.Fields{
				"stored_fp":   storedDeviceFingerprint[:8] + "...",
				"received_fp": deviceFingerprint[:8] + "...",
			}).Warn("OAuth device fingerprint mismatch - possible session fixation attack")
			return nil, &OAuthError{
				Type:        "device_verification_failed",
				Message:     "Device verification failed",
				Description: "The device fingerprint does not match the initial request. This could indicate a session fixation attack.",
			}
		}
		logrus.Debug("OAuth device fingerprint validated successfully")
	}

	// Validate provider
	p, exists := a.oauthProviders[provider]
	if !exists {
		return nil, &OAuthError{
			Type:        "invalid_provider",
			Message:     "OAuth provider not configured",
			Description: fmt.Sprintf("The OAuth provider '%s' is not configured on this server", provider),
		}
	}

	// Exchange code for token with PKCE code_verifier
	// PKCE code_verifier is required for public clients to prevent authorization code interception attacks
	exchangeOpts := []oauth2.AuthCodeOption{}
	if codeVerifier != "" {
		exchangeOpts = append(exchangeOpts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}
	token, err := p.Config.Exchange(context.Background(), code, exchangeOpts...)
	if err != nil {
		return nil, &OAuthError{
			Type:        "token_exchange_failed",
			Message:     "Failed to exchange authorization code",
			Description: "The authorization code could not be exchanged for an access token. It may be expired or invalid.",
		}
	}

	// Get user info from provider
	userInfo, err := a.getUserInfoFromProvider(provider, token.AccessToken)
	if err != nil {
		return nil, &OAuthError{
			Type:        "user_info_failed",
			Message:     "Failed to retrieve user information",
			Description: fmt.Sprintf("Could not get user information from %s. The access token may be invalid.", provider),
		}
	}

	// Validate required user info
	if userInfo.Email == "" {
		return nil, &OAuthError{
			Type:        "missing_email",
			Message:     "Email address required",
			Description: "Your account must have a verified email address to sign in. Please ensure your social account has an email address.",
		}
	}

	// Check if user exists by social provider
	existingUser, err := a.repo.GetUserBySocialProvider(provider, userInfo.ID)
	if err != nil {
		return nil, &OAuthError{
			Type:        "database_error",
			Message:     "Database error occurred",
			Description: "A database error occurred while checking your account. Please try again.",
		}
	}

	var user *storage.User
	var newUser bool

	if existingUser != nil {
		// User exists, update provider data and return
		user = existingUser
		newUser = false

		// Update provider data with latest info
		providerData := map[string]interface{}{
			"name":           userInfo.Name,
			"avatar_url":     userInfo.AvatarURL,
			"verified_email": userInfo.VerifiedEmail,
			"last_login":     time.Now(),
		}
		if err := a.repo.UpdateUserProviderData(user.ID, providerData); err != nil {
			// Log error but don't fail - this is not critical
			fmt.Printf("Warning: failed to update provider data: %v\n", err)
		}
	} else {
		// Check if user exists by email
		existingUserByEmail, err := a.repo.GetUserByEmail(userInfo.Email)
		if err != nil {
			return nil, &OAuthError{
				Type:        "database_error",
				Message:     "Database error occurred",
				Description: "A database error occurred while checking your account. Please try again.",
			}
		}

		if existingUserByEmail != nil {
			// Check if the existing user was created with a different provider
			// If so, require explicit confirmation before linking (security measure)
			if existingUserByEmail.Provider != nil && *existingUserByEmail.Provider != "" && *existingUserByEmail.Provider != provider {
				// User exists with a different provider - require explicit confirmation
				// Generate a temporary link token for confirmation
				linkToken, err := a.generateLinkToken(existingUserByEmail.ID, provider, userInfo.ID, userInfo)
				if err != nil {
					return nil, &OAuthError{
						Type:        "link_token_failed",
						Message:     "Failed to generate link token",
						Description: "Could not generate account linking token. Please try again.",
					}
				}

				return &OAuthCallbackResponse{
					User:         existingUserByEmail,
					LinkToken:    linkToken,
					LinkRequired: true,
					RedirectURI:  redirectURI,
					LoginHint:    loginHint,
				}, nil
			}

			// Link social account to existing user (same provider or no prior provider)
			user = existingUserByEmail
			newUser = false

			// Update user with social provider info
			err = a.linkSocialAccountToUser(user.ID, provider, userInfo.ID, map[string]interface{}{
				"name":           userInfo.Name,
				"avatar_url":     userInfo.AvatarURL,
				"verified_email": userInfo.VerifiedEmail,
				"linked_at":      time.Now(),
			})
			if err != nil {
				return nil, &OAuthError{
					Type:        "account_link_failed",
					Message:     "Failed to link social account",
					Description: "Your social account could not be linked to your existing account. Please contact support.",
				}
			}
		} else {
			// Create new user
			var oauthInviteReserved uuid.UUID
			if SignupInviteRequired() {
				si := strings.TrimSpace(storedInvite)
				if si == "" {
					return nil, &OAuthError{
						Type:        "invite_required",
						Message:     "Invite code required",
						Description: "Sign up requires a valid invite code. Start again from the sign-up page with your code, then use social login.",
					}
				}
				id, err := a.repo.ReserveSignupInvite(context.Background(), si)
				if err != nil {
					if errors.Is(err, storage.ErrSignupInviteExhausted) {
						return nil, &OAuthError{
							Type:        "invalid_invite",
							Message:     "Invite code exhausted",
							Description: "This invite code has no uses remaining.",
						}
					}
					return nil, &OAuthError{
						Type:        "invalid_invite",
						Message:     "Invalid invite code",
						Description: "The invite code is invalid or expired. Check the code and try again.",
					}
				}
				oauthInviteReserved = id
			}

			// Create a new tenant for each OAuth user (don't reuse existing tenants)
			// This ensures each user gets their own tenant with the correct plan
			tenant, err := a.repo.CreateTenant(context.Background(), "Default Tenant")
			if err != nil {
				if oauthInviteReserved != uuid.Nil {
					_ = a.repo.ReleaseSignupInviteReservation(context.Background(), oauthInviteReserved)
				}
				return nil, &OAuthError{
					Type:        "tenant_creation_failed",
					Message:     "Failed to create tenant",
					Description: "Could not create a tenant for your account. Please contact support.",
				}
			}
			tenantID := tenant.ID

			user, err = a.repo.CreateUserWithSocialAuth(userInfo.Email, tenantID, provider, userInfo.ID, map[string]interface{}{
				"name":           userInfo.Name,
				"avatar_url":     userInfo.AvatarURL,
				"verified_email": userInfo.VerifiedEmail,
				"created_via":    "oauth",
				"provider":       provider,
			})
			if err != nil {
				if oauthInviteReserved != uuid.Nil {
					_ = a.repo.ReleaseSignupInviteReservation(context.Background(), oauthInviteReserved)
				}
				return nil, &OAuthError{
					Type:        "user_creation_failed",
					Message:     "Failed to create account",
					Description: "Your account could not be created. Please try again or contact support.",
				}
			}
			newUser = true
		}
	}

	// Generate JWT token
	jwtToken, err := a.generateToken(user)
	if err != nil {
		return nil, &OAuthError{
			Type:        "token_generation_failed",
			Message:     "Authentication token error",
			Description: "Could not generate authentication token. Please try again.",
		}
	}

	// Generate refresh token for OAuth login
	refreshToken, refreshTokenHash, err := a.generateRefreshToken()
	if err != nil {
		return nil, &OAuthError{
			Type:        "token_generation_failed",
			Message:     "Authentication token error",
			Description: "Could not generate refresh token. Please try again.",
		}
	}

	// Store refresh token in database (expires in 30 days)
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = a.repo.CreateRefreshToken(user.ID, refreshTokenHash, "oauth", "oauth-callback", refreshExpiresAt)
	if err != nil {
		return nil, &OAuthError{
			Type:        "token_storage_failed",
			Message:     "Authentication token error",
			Description: "Could not store refresh token. Please try again.",
		}
	}

	if newUser {
		a.sendWelcomeNotification(context.Background(), user.ID)
	}

	return &OAuthCallbackResponse{
		Token:        jwtToken,
		RefreshToken: refreshToken,
		User:         user,
		NewUser:      newUser,
		RedirectURI:  redirectURI,
		LoginHint:    loginHint,
	}, nil
}

// getUserInfoFromProvider fetches user info from OAuth provider
func (a *AuthService) getUserInfoFromProvider(provider, accessToken string) (*OAuthUserInfo, error) {
	p, exists := a.oauthProviders[provider]
	if !exists {
		return nil, fmt.Errorf("provider not configured")
	}

	// Create HTTP client with token
	client := p.Config.Client(context.Background(), &oauth2.Token{AccessToken: accessToken})

	// Get user info
	resp, err := client.Get(p.UserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo OAuthUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// For GitHub, the /user endpoint may not include email if it's private.
	// Always fetch the emails list and prefer a primary+verified address.
	if provider == "github" {
		emails, err := a.getGitHubEmails(client)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub emails: %w", err)
		}
		// Selection priority: primary+verified → primary → first verified → first
		var chosen *GitHubEmail
		for i := range emails {
			e := &emails[i]
			if chosen == nil {
				chosen = e
				continue
			}
			// Prefer primary+verified over anything else
			if e.Primary && e.Verified && !(chosen.Primary && chosen.Verified) {
				chosen = e
			} else if e.Primary && !chosen.Primary {
				chosen = e
			} else if e.Verified && !chosen.Verified && !chosen.Primary {
				chosen = e
			}
		}
		if chosen != nil {
			userInfo.Email = chosen.Email
			userInfo.VerifiedEmail = chosen.Verified
		}
	}

	return &userInfo, nil
}

// getGitHubEmails gets email addresses from GitHub API
func (a *AuthService) getGitHubEmails(client *http.Client) ([]GitHubEmail, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, err
	}

	return emails, nil
}

// linkSocialAccountToUser links a social account to an existing user
func (a *AuthService) linkSocialAccountToUser(userID uuid.UUID, provider, providerID string, providerData map[string]interface{}) error {
	// Update the user record with social provider information
	ctx := context.Background()
	updates := map[string]interface{}{
		"provider":      &provider,
		"provider_id":   &providerID,
		"provider_data": providerData,
	}

	_, err := a.repo.UpdateUser(ctx, userID, updates)
	if err != nil {
		return fmt.Errorf("failed to link social account: %w", err)
	}

	return nil
}

// IsOAuthProviderConfigured checks if an OAuth provider is configured
func (a *AuthService) IsOAuthProviderConfigured(provider string) bool {
	_, exists := a.oauthProviders[provider]
	return exists
}

// GetConfiguredOAuthProviders returns list of configured OAuth providers
func (a *AuthService) GetConfiguredOAuthProviders() []string {
	providers := make([]string, 0, len(a.oauthProviders))
	for provider := range a.oauthProviders {
		providers = append(providers, provider)
	}
	return providers
}

// LinkTokenClaims represents the JWT claims for account linking tokens
type LinkTokenClaims struct {
	UserID       uuid.UUID `json:"user_id"`
	Provider     string    `json:"provider"`
	ProviderID   string    `json:"provider_id"`
	Email        string    `json:"email"`
	LinkType     string    `json:"link_type"` // "account_linking"
	jwt.RegisteredClaims
}

// generateLinkToken creates a temporary token for account linking confirmation.
// This token is short-lived and can only be used to confirm linking a specific
// social account to an existing user account.
func (a *AuthService) generateLinkToken(userID uuid.UUID, provider, providerID string, userInfo *OAuthUserInfo) (string, error) {
	if len(a.jwtSecret) == 0 {
		return "", fmt.Errorf("JWT secret not configured")
	}

	now := time.Now()
	claims := LinkTokenClaims{
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
		Email:      userInfo.Email,
		LinkType:   "account_linking",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)), // Short-lived: 15 minutes
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%s-%s-%s", provider, providerID, userInfo.Email), // Unique ID for this link attempt
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// ConfirmAccountLinking confirms linking a social account to an existing user.
// This should be called after the user has explicitly confirmed they want to link accounts.
func (a *AuthService) ConfirmAccountLinking(linkToken, provider, providerID string, userInfo *OAuthUserInfo) (*OAuthAccountLinkResponse, error) {
	// Validate the link token
	if linkToken == "" {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Invalid link token",
			Description: "The account linking token is missing. Please try signing in again.",
		}
	}

	// Parse and validate the JWT link token
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(Issuer),
	)

	token, err := parser.ParseWithClaims(linkToken, &LinkTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})
	if err != nil {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Invalid link token",
			Description: "The account linking token is invalid or has expired. Please sign in again.",
		}
	}

	claims, ok := token.Claims.(*LinkTokenClaims)
	if !ok || !token.Valid {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Invalid link token claims",
			Description: "The account linking token has invalid claims. Please sign in again.",
		}
	}

	// Verify the token is for account linking
	if claims.LinkType != "account_linking" {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Invalid link token type",
			Description: "The token is not an account linking token. Please sign in again.",
		}
	}

	// Verify the provider and providerID match
	if claims.Provider != provider || claims.ProviderID != providerID {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Link token provider mismatch",
			Description: "The account linking token does not match the provider. Please sign in again.",
		}
	}

	// Verify the email matches
	if claims.Email != userInfo.Email {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Link token email mismatch",
			Description: "The account linking token does not match your email. Please sign in again.",
		}
	}

	// Look up the user by email to get their ID
	existingUser, err := a.repo.GetUserByEmail(userInfo.Email)
	if err != nil {
		return nil, &OAuthError{
			Type:        "database_error",
			Message:     "Database error occurred",
			Description: "A database error occurred while checking your account. Please try again.",
		}
	}
	if existingUser == nil {
		return nil, &OAuthError{
			Type:        "user_not_found",
			Message:     "User not found",
			Description: "The user account could not be found. Please try signing in again.",
		}
	}

	// Verify the user ID in the token matches the existing user
	if claims.UserID != existingUser.ID {
		return nil, &OAuthError{
			Type:        "invalid_link_token",
			Message:     "Link token user mismatch",
			Description: "The account linking token does not match your user account. Please sign in again.",
		}
	}

	// Link the social account
	err = a.linkSocialAccountToUser(existingUser.ID, provider, providerID, map[string]interface{}{
		"name":           userInfo.Name,
		"avatar_url":     userInfo.AvatarURL,
		"verified_email": userInfo.VerifiedEmail,
		"linked_at":      time.Now(),
		"link_token":     linkToken,
	})
	if err != nil {
		return nil, &OAuthError{
			Type:        "account_link_failed",
			Message:     "Failed to link social account",
			Description: "Your social account could not be linked. Please contact support.",
		}
	}

	// Generate tokens
	accessToken, err := a.generateToken(existingUser)
	if err != nil {
		return nil, &OAuthError{
			Type:        "token_generation_failed",
			Message:     "Token generation failed",
			Description: "Could not generate authentication token. Please try again.",
		}
	}

	refreshToken, refreshTokenHash, err := a.generateRefreshToken()
	if err != nil {
		return nil, &OAuthError{
			Type:        "token_generation_failed",
			Message:     "Refresh token generation failed",
			Description: "Could not generate refresh token. Please try again.",
		}
	}

	// Store refresh token
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = a.repo.CreateRefreshToken(existingUser.ID, refreshTokenHash, "oauth_link", "link-confirmation", refreshExpiresAt)
	if err != nil {
		// Log but don't fail - access token is still valid
		fmt.Printf("Warning: failed to store refresh token: %v\n", err)
		refreshToken = ""
	}

	return &OAuthAccountLinkResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         existingUser,
	}, nil
}
