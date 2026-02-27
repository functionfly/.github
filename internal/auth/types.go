package auth

import (
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// OAuthProvider represents an OAuth provider configuration
type OAuthProvider struct {
	Name         string
	Config       *oauth2.Config
	UserInfoURL  string
	Scopes       []string
}

// OAuthUserInfo represents user information from OAuth providers
type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	AvatarURL     string `json:"avatar_url"`
	VerifiedEmail bool   `json:"verified_email"`
}

// OAuthCallbackRequest represents an OAuth callback request
type OAuthCallbackRequest struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

// OAuthCallbackResponse represents an OAuth callback response
type OAuthCallbackResponse struct {
	Token   string             `json:"token"`
	User    *storage.User      `json:"user"`
	NewUser bool               `json:"new_user"`
}

// Claims represents JWT claims
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Role        string    `json:"role,omitempty"`        // Platform role (for admin users)
	Permissions []string  `json:"permissions,omitempty"` // Explicit permissions
	jwt.RegisteredClaims
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	TermsAccepted  bool   `json:"termsAccepted"`
	RecaptchaToken string `json:"recaptchaToken,omitempty"`
}

// SignupResponse represents a signup response (no token until verified)
type SignupResponse struct {
	Message              string `json:"message"`
	EmailSent            bool   `json:"emailSent"`
	RequiresVerification bool   `json:"requiresVerification"`
}

// LoginUser is the safe user subset returned on login (no password, MFA secrets, etc.)
type LoginUser struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Email       string                 `json:"email"`
	Name        string                 `json:"name,omitempty"`
	Avatar      string                 `json:"avatar,omitempty"`
	Role        string                 `json:"role,omitempty"`
	EmailVerified bool                 `json:"email_verified"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string     `json:"token"`
	User  *LoginUser `json:"user"`
}

// OAuthError represents an OAuth-specific error with user-friendly messages
type OAuthError struct {
	Type        string // Error type for programmatic handling
	Message     string // Short error message
	Description string // Detailed error description for users
}

// Error implements the error interface
func (e *OAuthError) Error() string {
	return e.Message + ": " + e.Description
}

// GitHubEmail represents a GitHub email
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}