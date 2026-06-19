package gba

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// OAuthManager handles OAuth authentication
type OAuthManager struct {
	config *OAuthConfig
	logger *logrus.Logger
}

// NewOAuthManager creates a new OAuth manager
func NewOAuthManager(config *OAuthConfig) *OAuthManager {
	return &OAuthManager{
		config: config,
		logger: logrus.New(),
	}
}

// GetAuthURL returns the OAuth authorization URL for a provider
func (om *OAuthManager) GetAuthURL(provider, state string) (string, error) {
	switch provider {
	case "github":
		if !om.config.GitHubEnabled {
			return "", fmt.Errorf("GitHub OAuth is not enabled")
		}
		return om.config.GitHub.AuthCodeURL(state), nil
	case "google":
		if !om.config.GoogleEnabled {
			return "", fmt.Errorf("Google OAuth is not enabled")
		}
		return om.config.Google.AuthCodeURL(state), nil
	default:
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// ExchangeCode exchanges an authorization code for an access token
func (om *OAuthManager) ExchangeCode(ctx context.Context, provider, code string) (*oauth2.Token, error) {
	switch provider {
	case "github":
		if !om.config.GitHubEnabled {
			return nil, fmt.Errorf("GitHub OAuth is not enabled")
		}
		return om.config.GitHub.Exchange(ctx, code)
	case "google":
		if !om.config.GoogleEnabled {
			return nil, fmt.Errorf("Google OAuth is not enabled")
		}
		return om.config.Google.Exchange(ctx, code)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// GetUserInfo retrieves user information from the OAuth provider
func (om *OAuthManager) GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*OAuthUserInfo, error) {
	switch provider {
	case "github":
		return om.getGitHubUserInfo(ctx, token)
	case "google":
		return om.getGoogleUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// OAuthUserInfo represents user information from an OAuth provider
type OAuthUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	Verified bool   `json:"verified"`
	Provider string `json:"provider"`
}

// getGitHubUserInfo retrieves user info from GitHub
func (om *OAuthManager) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := om.config.GitHub.Client(ctx, token)

	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub user info: %w", err)
	}

	// Get primary email if not in user info
	if githubUser.Email == "" {
		email, err := om.getGitHubPrimaryEmail(ctx, token)
		if err != nil {
			om.logger.WithError(err).Warn("Failed to get GitHub primary email")
		} else {
			githubUser.Email = email
		}
	}

	return &OAuthUserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Email:    githubUser.Email,
		Name:     githubUser.Name,
		Image:    githubUser.AvatarURL,
		Verified: true, // GitHub emails are verified
		Provider: "github",
	}, nil
}

// getGitHubPrimaryEmail retrieves the primary email from GitHub
func (om *OAuthManager) getGitHubPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	client := om.config.GitHub.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// Return first verified email if no primary
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}

// getGoogleUserInfo retrieves user info from Google
func (om *OAuthManager) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := om.config.Google.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get Google user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google API error: %s", string(body))
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode Google user info: %w", err)
	}

	return &OAuthUserInfo{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Image:    googleUser.Picture,
		Verified: googleUser.VerifiedEmail,
		Provider: "google",
	}, nil
}

// FindOrCreateUser finds an existing user by OAuth provider or creates a new one
func (om *OAuthManager) FindOrCreateUser(db *gorm.DB, tenantID uuid.UUID, info *OAuthUserInfo) (*User, error) {
	var account Account

	// Try to find existing account
	err := db.Where("tenant_id = ? AND provider = ? AND provider_account_id = ?",
		tenantID, info.Provider, info.ID).First(&account).Error

	if err == nil {
		// Account exists, get user
		var user User
		if err := db.First(&user, account.UserID).Error; err != nil {
			return nil, fmt.Errorf("user not found for account: %w", err)
		}
		return &user, nil
	}

	// Account doesn't exist, check if user exists by email
	var user User
	err = db.Where("tenant_id = ? AND email = ?", tenantID, info.Email).First(&user).Error

	if err != nil {
		// Create new user
		user = User{
			TenantID:      tenantID,
			Email:         info.Email,
			Name:          info.Name,
			Image:         info.Image,
			EmailVerified: info.Verified,
			Provider:      info.Provider,
			ProviderID:    info.ID,
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Create account link
	account = Account{
		UserID:            user.ID,
		TenantID:          tenantID,
		Provider:          info.Provider,
		ProviderAccountID: info.ID,
	}
	if err := db.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &user, nil
}
