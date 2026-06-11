package auth

import (
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// OAuthProvider represents an OAuth provider configuration
type OAuthProvider struct {
	Name        string
	Config      *oauth2.Config
	UserInfoURL string
	Scopes      []string
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
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	User         *storage.User `json:"user"`
	NewUser      bool          `json:"new_user"`
	// RedirectURI is set when the client requested a CLI/local redirect; redirect the user here with token params
	RedirectURI string `json:"-"`
	// LoginHint preserves tenant subdomain or email context for redirecting back to origin post-auth
	LoginHint string `json:"-"`
	// LinkToken is set when account linking is required (user exists with different provider)
	// Client must call ConfirmAccountLinking with this token to complete the link
	LinkToken string `json:"link_token,omitempty"`
	// LinkRequired indicates that the user must confirm linking their social account
	LinkRequired bool `json:"link_required,omitempty"`
}

// OAuthAccountLinkRequest represents a request to confirm linking a social account
type OAuthAccountLinkRequest struct {
	LinkToken string `json:"link_token"`
	// Confirm indicates the user has confirmed they want to link the accounts
	Confirm bool `json:"confirm"`
}

// OAuthAccountLinkResponse represents the response after confirming account linking
type OAuthAccountLinkResponse struct {
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token"`
	User         *storage.User `json:"user"`
}

// Claims represents JWT claims
type Claims struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username,omitempty"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Role         string    `json:"role,omitempty"`                // Platform role (for admin users)
	Permissions  []string  `json:"permissions,omitempty"`        // Explicit permissions
	TokenVersion int       `json:"token_version,omitempty"`       // For token revocation - incremented on password change/logout all
	jwt.RegisteredClaims
}

// HasPermission checks if the claims contain the given permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// LoginRequest represents a login request
type LoginRequest struct {
	// Identifier can be either an email address or a username
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	TermsAccepted   bool   `json:"termsAccepted"`
	Username        string `json:"username,omitempty"`
	Name            string `json:"name,omitempty"`
	CompanyName     string `json:"companyName,omitempty"`
	DateOfBirth     string `json:"dateOfBirth"` // YYYY-MM-DD
	RecaptchaToken  string `json:"recaptchaToken,omitempty"`
	InviteCode      string `json:"inviteCode,omitempty"`
}

// SignupConfigResponse is returned by the public signup-config endpoint.
type SignupConfigResponse struct {
	InviteRequired bool `json:"inviteRequired"`
}

// SignupResponse represents a signup response (no token until verified)
type SignupResponse struct {
	Message              string `json:"message"`
	EmailSent            bool   `json:"emailSent"`
	RequiresVerification bool   `json:"requiresVerification"`
}

// LoginUser is the safe user subset returned on login (no password, MFA secrets, etc.)
type LoginUser struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	Username      string                 `json:"username,omitempty"`
	Email         string                 `json:"email"`
	CompanyName   string                 `json:"company_name,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Avatar        string                 `json:"avatar,omitempty"`
	Role          string                 `json:"role,omitempty"`
	Plan          string                 `json:"plan,omitempty"` // Tenant's plan (e.g. free, starter, enterprise)
	EmailVerified bool                   `json:"email_verified"`
	ProviderData  map[string]interface{} `json:"provider_data,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
	User         *LoginUser `json:"user"`
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

// AuthCallbackResult provides a consistent callback response pattern across all auth methods
// (OAuth, SAML, etc.) for both browser and CLI flows.
type AuthCallbackResult struct {
	Success      bool   `json:"success"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error,omitempty"`             // machine-readable error code
	ErrorDesc    string `json:"error_description,omitempty"` // human-readable error description
	NewUser      bool   `json:"new_user,omitempty"`
	NameID       string `json:"name_id,omitempty"` // SAML-specific
}

// AuthCallbackErrorCode defines standard error codes for auth callbacks
// These should be used consistently across OAuth, SAML, and other auth methods.
type AuthCallbackErrorCode string

const (
	// Configuration errors
	AuthErrProviderNotConfigured AuthCallbackErrorCode = "provider_not_configured"
	AuthErrSAMLEnabled           AuthCallbackErrorCode = "saml_not_enabled"

	// Request validation errors
	AuthErrInvalidState     AuthCallbackErrorCode = "invalid_state"
	AuthErrInvalidRequest   AuthCallbackErrorCode = "invalid_request"
	AuthErrMissingParameter AuthCallbackErrorCode = "missing_parameter"

	// Authentication errors
	AuthErrTokenExchangeFailed AuthCallbackErrorCode = "token_exchange_failed"
	AuthErrUserInfoFailed      AuthCallbackErrorCode = "user_info_failed"
	AuthErrMissingEmail        AuthCallbackErrorCode = "missing_email"
	AuthErrAccountLinkFailed   AuthCallbackErrorCode = "account_link_failed"

	// User/Account errors
	AuthErrUserCreationFailed   AuthCallbackErrorCode = "user_creation_failed"
	AuthErrTenantCreationFailed AuthCallbackErrorCode = "tenant_creation_failed"
	AuthErrInvalidInvite        AuthCallbackErrorCode = "invalid_invite"
	AuthErrInviteRequired       AuthCallbackErrorCode = "invite_required"

	// Token errors
	AuthErrTokenGenerationFailed AuthCallbackErrorCode = "token_generation_failed"
	AuthErrTokenStorageFailed    AuthCallbackErrorCode = "token_storage_failed"

	// SAML-specific errors
	AuthErrSAMLInvalidResponse  AuthCallbackErrorCode = "saml_invalid_response"
	AuthErrSAMLInvalidSignature AuthCallbackErrorCode = "saml_invalid_signature"
	AuthErrSAMLInvalidTenant    AuthCallbackErrorCode = "saml_invalid_tenant"
	AuthErrSAMLInvalidConfig    AuthCallbackErrorCode = "saml_invalid_config"
	AuthErrSAMLNoAssertion      AuthCallbackErrorCode = "saml_no_assertion"
	AuthErrSAMLNoEmail          AuthCallbackErrorCode = "saml_no_email"

	// Database errors
	AuthErrDatabaseError AuthCallbackErrorCode = "database_error"

	// Catch-all
	AuthErrUnknown AuthCallbackErrorCode = "unknown_error"
)

// String returns the string representation of the error code
func (c AuthCallbackErrorCode) String() string {
	return string(c)
}

// ErrorDescriptions maps error codes to user-friendly descriptions
var ErrorDescriptions = map[AuthCallbackErrorCode]string{
	AuthErrProviderNotConfigured: "The authentication provider is not configured on this server.",
	AuthErrSAMLEnabled:           "SAML authentication is not enabled for this tenant.",
	AuthErrInvalidState:          "The authentication state is invalid or has expired. Please try signing in again.",
	AuthErrInvalidRequest:        "The authentication request is invalid.",
	AuthErrMissingParameter:      "A required parameter is missing from the authentication request.",
	AuthErrTokenExchangeFailed:   "The authorization code could not be exchanged for an access token.",
	AuthErrUserInfoFailed:        "Could not retrieve user information from the provider.",
	AuthErrMissingEmail:          "Your account must have a verified email address to sign in.",
	AuthErrAccountLinkFailed:     "Your social account could not be linked to your existing account.",
	AuthErrUserCreationFailed:    "Your account could not be created. Please try again or contact support.",
	AuthErrTenantCreationFailed:  "Could not create a tenant for your account. Please contact support.",
	AuthErrInvalidInvite:         "The invite code is invalid or expired.",
	AuthErrInviteRequired:        "Sign up requires a valid invite code.",
	AuthErrTokenGenerationFailed: "Could not generate authentication token. Please try again.",
	AuthErrTokenStorageFailed:    "Could not store authentication token. Please try again.",
	AuthErrSAMLInvalidResponse:   "The SAML response could not be processed.",
	AuthErrSAMLInvalidSignature:  "The SAML response signature verification failed.",
	AuthErrSAMLInvalidTenant:     "Invalid tenant specified for SAML authentication.",
	AuthErrSAMLInvalidConfig:     "SAML configuration is invalid or incomplete.",
	AuthErrSAMLNoAssertion:       "No authentication assertion found in SAML response.",
	AuthErrSAMLNoEmail:           "No email address found in SAML response.",
	AuthErrDatabaseError:         "A database error occurred. Please try again.",
	AuthErrUnknown:               "An unexpected error occurred during authentication.",
}

// GetErrorDescription returns a user-friendly description for an error code
func GetErrorDescription(code AuthCallbackErrorCode) string {
	if desc, ok := ErrorDescriptions[code]; ok {
		return desc
	}
	return ErrorDescriptions[AuthErrUnknown]
}
