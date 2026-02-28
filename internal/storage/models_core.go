package storage

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                    uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant                *Tenant    `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Username              *string    `json:"username,omitempty" gorm:"uniqueIndex;size:255"`
	Email                 string     `json:"email" gorm:"uniqueIndex;not null"`
	Name                  string     `json:"name,omitempty" gorm:"size:255"` // Display name (separate from OAuth provider name)
	PasswordHash          string     `json:"password_hash" gorm:"column:password_hash"`
	Role                  string     `json:"role,omitempty" gorm:"size:50"` // Platform role for admin users
	EmailVerified         bool       `json:"email_verified" gorm:"default:false"`
	CompanyName           *string    `json:"company_name,omitempty" gorm:"size:255"`
	Bio                   *string    `json:"bio,omitempty" gorm:"type:text"`
	VerificationToken     *string    `json:"verification_token,omitempty"`
	VerificationExpiresAt *time.Time `json:"verification_expires_at,omitempty"`
	// Social authentication fields
	Provider     *string                `json:"provider,omitempty"`                        // 'google', 'github', etc.
	ProviderID   *string                `json:"provider_id,omitempty"`                     // External user ID from OAuth provider
	ProviderData map[string]interface{} `json:"provider_data,omitempty" gorm:"type:jsonb"` // Additional provider-specific data
	// MFA fields
	MFASecret      *string    `json:"mfa_secret,omitempty"`                         // TOTP secret for MFA
	MFAEnabled     bool       `json:"mfa_enabled" gorm:"default:false"`             // Whether MFA is enabled for this user
	MFABackupCodes []string   `json:"mfa_backup_codes,omitempty" gorm:"type:jsonb"` // Backup codes for MFA recovery
	MFALastUsed    *time.Time `json:"mfa_last_used,omitempty"`                      // Last time MFA was used
	// Team collaboration fields
	Teams     []TeamMembership `json:"teams,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID           uuid.UUID   `json:"id"`
	ActorUserID  *uuid.UUID  `json:"actor_user_id,omitempty"`
	ActorEmail   string      `json:"actor_email,omitempty"`
	TenantID     *uuid.UUID  `json:"tenant_id,omitempty"`
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	RequestID    string      `json:"request_id,omitempty"`
	BeforeState  interface{} `json:"before_state,omitempty"`
	AfterState   interface{} `json:"after_state,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
	Success      bool        `json:"success"`
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Plan      string    `json:"plan" gorm:"not null"`
	Status    string    `json:"status" gorm:"not null;default:'active'"` // 'active', 'suspended'
	Users     []User    `json:"users,omitempty" gorm:"foreignKey:TenantID"`
	Apps      []App     `json:"apps,omitempty" gorm:"foreignKey:TenantID"`
	Teams     []Team    `json:"teams,omitempty" gorm:"foreignKey:TenantID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Team represents a team within a tenant
type Team struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID        `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant      *Tenant          `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Name        string           `json:"name" gorm:"not null"`
	Description string           `json:"description"`
	CreatedBy   uuid.UUID        `json:"created_by" gorm:"type:uuid;not null"`
	Members     []TeamMembership `json:"members,omitempty" gorm:"foreignKey:TeamID"`
	Permissions []TeamPermission `json:"permissions,omitempty" gorm:"foreignKey:TeamID"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`
}

// TeamMembership represents a user's membership in a team
type TeamMembership struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID  uuid.UUID `json:"team_id" gorm:"type:uuid;not null"`
	Team    *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	UserID  uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	User    *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Role    string    `json:"role" gorm:"not null"` // 'owner', 'admin', 'member', 'viewer'
	AddedBy uuid.UUID `json:"added_by" gorm:"type:uuid;not null"`
	AddedAt time.Time `json:"added_at" gorm:"autoCreateTime"`
}

// TeamPermission represents permissions granted to a team for specific resources
type TeamPermission struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID       uuid.UUID `json:"team_id" gorm:"type:uuid;not null"`
	Team         *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	ResourceType string    `json:"resource_type" gorm:"not null"` // 'app', 'function', 'backend', 'deployment'
	ResourceID   uuid.UUID `json:"resource_id" gorm:"type:uuid;not null"`
	Permissions  string    `json:"permissions" gorm:"not null"` // JSON array of permissions like ["read", "write", "delete"]
	GrantedBy    uuid.UUID `json:"granted_by" gorm:"type:uuid;not null"`
	GrantedAt    time.Time `json:"granted_at" gorm:"autoCreateTime"`
}

// TeamInvite represents a team invitation during onboarding
type TeamInvite struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID     uuid.UUID  `json:"team_id" gorm:"type:uuid;not null;index"`
	Email      string     `json:"email" gorm:"not null"`
	Token      string     `json:"token" gorm:"uniqueIndex;not null"`
	Role       string     `json:"role" gorm:"not null"` // "admin", "member", "viewer"
	InvitedBy  uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`
	Message    string     `json:"message,omitempty"`
	Status     string     `json:"status" gorm:"default:'pending'"` // "pending", "accepted", "expired", "cancelled"
	ExpiresAt  time.Time  `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (TeamInvite) TableName() string {
	return "team_invites"
}

// Provider represents a cloud provider configuration
type Provider struct {
	ID        string    `json:"id" gorm:"type:varchar(255);primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Provider  string    `json:"provider" gorm:"not null;index"` // "cloudflare", "vercel", "fly"
	Token     string    `json:"token" gorm:"not null"`          // Encrypted API token
	Status    string    `json:"status" gorm:"not null"`         // "active", "inactive", "error"
	IsShared  bool      `json:"is_shared" gorm:"default:false"` // Shared with team
	TeamID    *string   `json:"team_id,omitempty" gorm:"type:varchar(255);index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Provider) TableName() string {
	return "providers"
}
