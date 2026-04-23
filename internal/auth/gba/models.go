package gba

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the GoBetterAuth system
type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index"`
	Email    string    `gorm:"uniqueIndex:idx_users_email_tenant;size:255;not null"`
	Username string    `gorm:"uniqueIndex:idx_users_username_tenant;size:255"`
	Password string    `gorm:"size:255"` // Hashed password
	Name     string    `gorm:"size:255"`
	Image    string    `gorm:"size:512"`
	Role     string    `gorm:"size:50;default:'user'"`

	// Email verification
	EmailVerified bool `gorm:"default:false"`
	VerifiedAt    *time.Time

	// MFA
	MFAEnabled bool   `gorm:"default:false"`
	MFASecret  string `gorm:"size:255"`

	// OAuth
	Provider   string `gorm:"size:50"`
	ProviderID string `gorm:"size:255;index"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "gba_users"
}

// Account represents an OAuth provider account
type Account struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;index"`
	TenantID          uuid.UUID `gorm:"type:uuid;not null;index"`
	Provider          string    `gorm:"size:50;not null"`
	ProviderAccountID string    `gorm:"size:255;not null"`

	// OAuth tokens (encrypted at application level)
	AccessToken  string `gorm:"type:text"`
	RefreshToken string `gorm:"type:text"`
	TokenType    string `gorm:"size:50"`
	Scope        string `gorm:"size:512"`

	// Token expiry
	ExpiresAt *time.Time

	// Session state
	SessionState string `gorm:"size:255"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for the Account model
func (Account) TableName() string {
	return "gba_accounts"
}

// Session represents an active user session
type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index"`
	SessionToken string    `gorm:"size:255;uniqueIndex;not null"`

	// MFA
	MFAVerified bool `gorm:"default:false"`

	// Trusted device tokens — when set, the session is bound to a trusted device
	// and gets extended expiry (30 days vs the default session max age)
	TrustedDeviceToken string `gorm:"size:64;index"`

	// Security
	IPAddress string `gorm:"size:45"` // IPv6 max length
	UserAgent string `gorm:"type:text"`

	// Expiry
	ExpiresAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for the Session model
func (Session) TableName() string {
	return "gba_sessions"
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// VerificationToken represents an email verification or password reset token
type VerificationToken struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Identifier string    `gorm:"size:255;not null;index"` // Usually email
	Token      string    `gorm:"size:255;uniqueIndex;not null"`
	ExpiresAt  time.Time `gorm:"not null"`
	TenantID   uuid.UUID `gorm:"type:uuid"`
	CreatedAt  time.Time
}

// TableName returns the table name for the VerificationToken model
func (VerificationToken) TableName() string {
	return "gba_verification_tokens"
}

// IsExpired returns true if the token has expired
func (v *VerificationToken) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// Tenant represents a tenant/organization
type Tenant struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string    `gorm:"size:255;not null"`
	Subdomain string    `gorm:"size:255;uniqueIndex"`
	Status    string    `gorm:"size:50;default:'active'"`

	// Authentication policies
	MFAPolicy               string `gorm:"size:50;default:'optional'"` // optional, required, suspended
	SessionMaxDuration      int    `gorm:"default:604800"`             // 7 days in seconds
	SessionIdleTimeout      int    `gorm:"default:1800"`               // 30 minutes in seconds
	ConcurrentSessionsLimit int    `gorm:"default:5"`

	// Domain restrictions
	AllowedEmailDomains []string `gorm:"type:text[]"`

	// OAuth configuration
	OAuthProvidersEnabled []string `gorm:"type:text[]"`

	// Security
	IPAllowlistEnabled bool `gorm:"default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for the Tenant model
func (Tenant) TableName() string {
	return "gba_tenants"
}

// TenantIPAllowlist represents an IP allowlist entry for a tenant
type TenantIPAllowlist struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	IPAddress   string    `gorm:"size:45;not null"`
	Description string    `gorm:"size:255"`
	ExpiresAt   *time.Time
	CreatedAt   time.Time
}

// TableName returns the table name for the TenantIPAllowlist model
func (TenantIPAllowlist) TableName() string {
	return "gba_tenant_ip_allowlist"
}

// AuthAuditLog represents an authentication audit log entry
type AuthAuditLog struct {
	ID           uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     uuid.UUID              `gorm:"type:uuid;not null;index"`
	UserID       *uuid.UUID             `gorm:"type:uuid;index"`
	Email        string                 `gorm:"size:255;index"`
	IPAddress    string                 `gorm:"size:45;index"`
	UserAgent    string                 `gorm:"type:text"`
	Action       string                 `gorm:"size:50;not null;index"` // login, logout, password_reset, etc.
	Success      bool                   `gorm:"default:true"`
	ErrorMessage string                 `gorm:"type:text"`
	Metadata     map[string]interface{} `gorm:"serializer:json"`
	CreatedAt    time.Time
}

// TableName returns the table name for the AuthAuditLog model
func (AuthAuditLog) TableName() string {
	return "gba_auth_audit_logs"
}

// Role represents a role in the unified role-based system
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `gorm:"size:50;uniqueIndex;not null"`
	Description string    `gorm:"size:255"`
	Level       int       `gorm:"default:0"` // Hierarchy level for inheritance

	// Permissions associated with this role
	Permissions []Permission `gorm:"many2many:gba_role_permissions;"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "gba_roles"
}

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"size:255"`
	Resource    string    `gorm:"size:50;index"` // e.g., "apps", "functions", "users"
	Action      string    `gorm:"size:50;index"` // e.g., "read", "write", "delete"

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the table name for the Permission model
func (Permission) TableName() string {
	return "gba_permissions"
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primary_key"`
	PermissionID uuid.UUID `gorm:"type:uuid;primary_key"`
}

// TableName returns the table name for the RolePermission model
func (RolePermission) TableName() string {
	return "gba_role_permissions"
}

// Migrate performs database migrations for GoBetterAuth tables
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Account{},
		&Session{},
		&VerificationToken{},
		&Tenant{},
		&TenantIPAllowlist{},
		&AuthAuditLog{},
		&Role{},
		&Permission{},
		&RolePermission{},
	)
}
