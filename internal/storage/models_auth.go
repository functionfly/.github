package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TenantAuthSettings represents per-tenant authentication configuration
type TenantAuthSettings struct {
	ID                     uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID               uuid.UUID       `json:"tenant_id" gorm:"type:uuid;uniqueIndex;not null"`
	MFARequired            bool            `json:"mfa_required" gorm:"not null;default:false"`
	MFAMode                string          `json:"mfa_mode" gorm:"type:varchar(20);not null;default:'optional'"`
	PasswordPolicy         json.RawMessage `json:"password_policy" gorm:"type:jsonb;not null;default:'{}'"`
	SessionTimeoutMinutes  int             `json:"session_timeout_minutes" gorm:"not null;default:480"`
	IPAllowlistEnabled     bool            `json:"ip_allowlist_enabled" gorm:"not null;default:false"`
	IPAllowlist            json.RawMessage `json:"ip_allowlist" gorm:"type:jsonb;not null;default:'[]'"`
	AllowedDomains         json.RawMessage `json:"allowed_domains" gorm:"type:jsonb;not null;default:'[]'"`
	SSOProvider            string          `json:"sso_provider" gorm:"type:varchar(20);not null;default:'none'"`
	SAMLMetadataURL        *string         `json:"saml_metadata_url,omitempty" gorm:"type:text"`
	SAMLEntityID          *string         `json:"saml_entity_id,omitempty" gorm:"type:text"`
	SAMLCertificate        *string         `json:"saml_certificate,omitempty" gorm:"type:text"`
	SAMLPrivateKey         *string         `json:"saml_private_key,omitempty" gorm:"type:text"`
	UseCustomBranding     bool            `json:"use_custom_branding" gorm:"not null;default:false"`
	EmailFromName          string          `json:"email_from_name" gorm:"type:varchar(100);not null;default:'FunctionFly'"`
	EmailFromAddress       string          `json:"email_from_address" gorm:"type:varchar(255);not null;default:'noreply@functionfly.com'"`
	RequireEmailVerification bool          `json:"require_email_verification" gorm:"not null;default:true"`
	AllowPasswordLogin     bool            `json:"allow_password_login" gorm:"not null;default:true"`
	AllowMagicLink         bool            `json:"allow_magic_link" gorm:"not null;default:true"`
	MaxLoginAttempts       int             `json:"max_login_attempts" gorm:"not null;default:5"`
	LockoutDurationMinutes int             `json:"lockout_duration_minutes" gorm:"not null;default:15"`
	CreatedAt              time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Tenant            *Tenant               `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	OAuthProviders    []*TenantOAuthProvider `json:"oauth_providers,omitempty" gorm:"foreignKey:TenantID"`
	InviteCodes       []*TenantInviteCode   `json:"invite_codes,omitempty" gorm:"foreignKey:TenantID"`
	Memberships       []*TenantMembership   `json:"memberships,omitempty" gorm:"foreignKey:TenantID"`
	AuthAuditLog      []*TenantAuthAuditLog `json:"auth_audit_log,omitempty" gorm:"foreignKey:TenantID"`
}

// TableName returns the table name for TenantAuthSettings
func (TenantAuthSettings) TableName() string {
	return "tenant_auth_settings"
}

// PasswordPolicy represents password requirements
type PasswordPolicy struct {
	MinLength         int  `json:"min_length"`
	RequireUppercase  bool `json:"require_uppercase"`
	RequireLowercase  bool `json:"require_lowercase"`
	RequireDigit      bool `json:"require_digit"`
	RequireSpecial    bool `json:"require_special"`
	MaxAgeDays        int  `json:"max_age_days,omitempty"`       // 0 = no expiry
	PreventReuseCount int  `json:"prevent_reuse_count,omitempty"` // 0 = allow reuse
}

// DefaultPasswordPolicy returns the default strict password policy
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:         8,
		RequireUppercase:  true,
		RequireLowercase:  true,
		RequireDigit:      true,
		RequireSpecial:    true,
		MaxAgeDays:        0,
		PreventReuseCount: 5,
	}
}

// TenantOAuthProvider represents OAuth credentials for a tenant
type TenantOAuthProvider struct {
	ID                       uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID                 uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_provider"`
	Provider                 string         `json:"provider" gorm:"type:varchar(50);not null;uniqueIndex:idx_tenant_provider"`
	ClientID                 string         `json:"client_id" gorm:"type:varchar(255);not null"`
	EncryptedClientSecret    string         `json:"-" gorm:"type:text;not null"`
	EncryptedClientSecretIV  *string        `json:"-" gorm:"type:text"`
	EncryptedClientSecretTag *string       `json:"-" gorm:"type:text"`
	Enabled                  bool           `json:"enabled" gorm:"not null;default:true"`
	CallbackURL             *string        `json:"callback_url,omitempty" gorm:"type:text"`
	Scopes                  json.RawMessage `json:"scopes" gorm:"type:jsonb;not null;default:'[\"user:email\", \"read:user\"]'`
	CreatedAt               time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Decrypted client secret (transient, never stored)
	ClientSecret string `json:"client_secret,omitempty" gorm:"-"`
}

// TableName returns the table name for TenantOAuthProvider
func (TenantOAuthProvider) TableName() string {
	return "tenant_oauth_providers"
}

// OAuthProvider constants
const (
	OAuthProviderGitHub     = "github"
	OAuthProviderGoogle    = "google"
	OAuthProviderMicrosoft = "microsoft"
	OAuthProviderApple     = "apple"
	OAuthProviderGitLab    = "gitlab"
	OAuthProviderBitbucket = "bitbucket"
)

// ValidOAuthProviders returns valid OAuth provider names
func ValidOAuthProviders() []string {
	return []string{OAuthProviderGitHub, OAuthProviderGoogle, OAuthProviderMicrosoft, OAuthProviderApple, OAuthProviderGitLab, OAuthProviderBitbucket}
}

// IsValidOAuthProvider checks if a provider name is valid
func IsValidOAuthProvider(provider string) bool {
	for _, p := range ValidOAuthProviders() {
		if p == provider {
			return true
		}
	}
	return false
}

// TenantInviteCode represents an invite code for team member invitation
type TenantInviteCode struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID  uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Code      string     `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	Email     string     `json:"email" gorm:"type:varchar(255);not null"`
	Role      string     `json:"role" gorm:"type:varchar(50);not null;default:'team_member'"`
	InvitedBy uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty" gorm:"index"`
	AcceptedBy *uuid.UUID `json:"accepted_by,omitempty" gorm:"type:uuid"`
	MaxUses   int        `json:"max_uses" gorm:"not null;default:1"`
	Uses      int        `json:"uses" gorm:"not null;default:0"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Tenant    *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Inviter   *User   `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
	Acceptor  *User   `json:"acceptor,omitempty" gorm:"foreignKey:AcceptedBy"`
}

// TableName returns the table name for TenantInviteCode
func (TenantInviteCode) TableName() string {
	return "tenant_invite_codes"
}

// Role constants for team membership
const (
	RoleTeamOwner   = "team_owner"
	RoleTeamAdmin   = "team_admin"
	RoleTeamMember  = "team_member"
	RoleTeamViewer  = "team_viewer"
)

// ValidRoles returns valid role names
func ValidRoles() []string {
	return []string{RoleTeamOwner, RoleTeamAdmin, RoleTeamMember, RoleTeamViewer}
}

// IsValidRole checks if a role name is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// TenantMembership represents a user's membership in a tenant team
type TenantMembership struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_user"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_user"`
	Role       string     `json:"role" gorm:"type:varchar(50);not null;default:'team_member'"`
	InvitedBy  *uuid.UUID `json:"invited_by,omitempty" gorm:"type:uuid"`
	InvitedAt  *time.Time `json:"invited_at,omitempty"`
	JoinedAt   time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	Status     string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`

	// Relationships
	Tenant  *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	User    *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Inviter *User   `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

// TableName returns the table name for TenantMembership
func (TenantMembership) TableName() string {
	return "tenant_memberships"
}

// MembershipStatus constants
const (
	MembershipStatusActive    = "active"
	MembershipStatusSuspended = "suspended"
	MembershipStatusInvited  = "invited"
)

// TenantAuthAuditLog represents an auth event audit entry
type TenantAuthAuditLog struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID      uuid.UUID       `json:"tenant_id" gorm:"type:uuid;not null;index"`
	UserID        *uuid.UUID      `json:"user_id,omitempty" gorm:"type:uuid"`
	Action        string          `json:"action" gorm:"type:varchar(100);not null;index"`
	ResourceType  *string         `json:"resource_type,omitempty" gorm:"type:varchar(100)"`
	ResourceID    *uuid.UUID      `json:"resource_id,omitempty" gorm:"type:uuid"`
	IPAddress     *string         `json:"ip_address,omitempty" gorm:"type:inet"`
	UserAgent     *string         `json:"user_agent,omitempty" gorm:"type:text"`
	Metadata      json.RawMessage `json:"metadata,omitempty" gorm:"type:jsonb"`
	Success       bool            `json:"success" gorm:"not null;default:true"`
	ErrorMessage  *string         `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime;index"`

	// Relationships
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for TenantAuthAuditLog
func (TenantAuthAuditLog) TableName() string {
	return "tenant_auth_audit_log"
}

// AuthAction constants for audit logging
const (
	AuthActionLoginSuccess       = "login.success"
	AuthActionLoginFailed        = "login.failed"
	AuthActionLogout             = "logout"
	AuthActionPasswordChanged    = "password.changed"
	AuthActionPasswordReset     = "password.reset"
	AuthActionEmailVerified     = "email.verified"
	AuthActionMFASetup          = "mfa.setup"
	AuthActionMFAEnabled        = "mfa.enabled"
	AuthActionMFADisabled       = "mfa.disabled"
	AuthActionMFAVerified       = "mfa.verified"
	AuthActionMFALogin          = "mfa.login"
	AuthActionOAuthConnected    = "oauth.connected"
	AuthActionOAuthDisconnected = "oauth.disconnected"
	AuthActionOAuthLogin        = "oauth.login"
	AuthActionInviteSent        = "invite.sent"
	AuthActionInviteAccepted   = "invite.accepted"
	AuthActionInviteRevoked    = "invite.revoked"
	AuthActionUserInvited      = "user.invited"
	AuthActionUserRemoved      = "user.removed"
	AuthActionRoleChanged      = "role.changed"
	AuthActionSessionCreated   = "session.created"
	AuthActionSessionRevoked   = "session.revoked"
	AuthActionAccountLocked    = "account.locked"
	AuthActionAccountUnlocked  = "account.unlocked"
)

// MFAMode constants
const (
	MFAModeOptional  = "optional"
	MFAModeRequired = "required"
	MFAModeEnforced = "enforced"
)

// SSOProvider constants
const (
	SSOProviderNone  = "none"
	SSOProviderSAML  = "saml"
	SSOProviderOIDC  = "oidc"
)