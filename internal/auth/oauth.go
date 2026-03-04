package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

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

// GetOAuthURL generates OAuth authorization URL for a provider
func (a *AuthService) GetOAuthURL(provider string) (string, error) {
	p, exists := a.oauthProviders[provider]
	if !exists {
		return "", fmt.Errorf("OAuth provider '%s' not configured", provider)
	}

	// Generate state for CSRF protection
	state, err := a.generateVerificationToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate OAuth state: %w", err)
	}

	// Store state server-side (persisted for multi-instance) so we can validate it on callback
	expiresAt := time.Now().Add(10 * time.Minute)
	if err := a.repo.StoreOAuthState(context.Background(), state, expiresAt); err != nil {
		return "", fmt.Errorf("failed to store OAuth state: %w", err)
	}

	return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// HandleOAuthCallback processes OAuth callback and returns user authentication
func (a *AuthService) HandleOAuthCallback(provider, code, state string) (*OAuthCallbackResponse, error) {
	// Validate CSRF state token — must match one issued by GetOAuthURL (persisted, one-time use)
	if state == "" {
		return nil, &OAuthError{
			Type:        "invalid_state",
			Message:     "Invalid or expired OAuth state",
			Description: "The OAuth state parameter is invalid or has expired. Please try signing in again.",
		}
	}
	valid, err := a.repo.ValidateAndConsumeOAuthState(context.Background(), state)
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

	// Validate provider
	p, exists := a.oauthProviders[provider]
	if !exists {
		return nil, &OAuthError{
			Type:        "invalid_provider",
			Message:     "OAuth provider not configured",
			Description: fmt.Sprintf("The OAuth provider '%s' is not configured on this server", provider),
		}
	}

	// Exchange code for token
	token, err := p.Config.Exchange(context.Background(), code)
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
			// Link social account to existing user
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
			tenants, err := a.repo.ListTenants()
			if err != nil {
				return nil, &OAuthError{
					Type:        "database_error",
					Message:     "Database error occurred",
					Description: "A database error occurred while creating your account. Please try again.",
				}
			}

			var tenantID uuid.UUID
			if len(tenants) > 0 {
				tenantID = tenants[0].ID
			} else {
				tenant, err := a.repo.CreateTenant(context.Background(), "Default Tenant")
				if err != nil {
					return nil, &OAuthError{
						Type:        "tenant_creation_failed",
						Message:     "Failed to create tenant",
						Description: "Could not create a tenant for your account. Please contact support.",
					}
				}
				tenantID = tenant.ID
			}

			user, err = a.repo.CreateUserWithSocialAuth(userInfo.Email, tenantID, provider, userInfo.ID, map[string]interface{}{
				"name":           userInfo.Name,
				"avatar_url":     userInfo.AvatarURL,
				"verified_email": userInfo.VerifiedEmail,
				"created_via":    "oauth",
				"provider":       provider,
			})
			if err != nil {
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
