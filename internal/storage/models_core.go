package storage

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant        *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Username      *string   `json:"username,omitempty" gorm:"uniqueIndex;size:255"`
	Email         string    `json:"email" gorm:"uniqueIndex;not null"`
	Name          string    `json:"name,omitempty" gorm:"size:255"` // Display name (separate from OAuth provider name)
	PasswordHash  string    `json:"password_hash" gorm:"column:password_hash"`
	Role          string    `json:"role,omitempty" gorm:"size:50"` // Platform role for admin users
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	// TokenVersion is used for JWT revocation - incrementing this invalidates all existing tokens
	// Set via application code on password change/logout all
	TokenVersion  int        `json:"token_version,omitempty" gorm:"default:0"`
	CompanyName   *string   `json:"company_name,omitempty" gorm:"size:255"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty" gorm:"column:date_of_birth;type:date"`
	Bio           *string   `json:"bio,omitempty" gorm:"type:text"`
	// ProfileNumber is a sequential number assigned to users based on registration order
	// Used to identify and reward early adopters (e.g., "Member #123")
	ProfileNumber *int      `json:"profile_number,omitempty" gorm:"column:profile_number;uniqueIndex"`
	// Extended profile fields
	Location              *string                `json:"location,omitempty" gorm:"size:255"`
	Website               *string                `json:"website,omitempty" gorm:"size:500"`
	JobTitle              *string                `json:"job_title,omitempty" gorm:"size:255"`
	SocialLinks           map[string]interface{} `json:"social_links,omitempty" gorm:"type:jsonb;default:'{}'"`
	TwitterURL            *string                `json:"twitter_url,omitempty" gorm:"column:twitter_url;size:500"`
	GithubURL             *string                `json:"github_url,omitempty" gorm:"column:github_url;size:500"`
	LinkedInURL           *string                `json:"linkedin_url,omitempty" gorm:"column:linkedin_url;size:500"`
	CoverImageURL         *string                `json:"cover_image_url,omitempty" gorm:"column:cover_image_url;size:500"`
	VerificationToken     *string                `json:"verification_token,omitempty"`
	VerificationExpiresAt *time.Time             `json:"verification_expires_at,omitempty"`
	// Social authentication fields
	Provider     *string                `json:"provider,omitempty"`                        // 'google', 'github', etc.
	ProviderID   *string                `json:"provider_id,omitempty"`                     // External user ID from OAuth provider
	ProviderData map[string]interface{} `json:"provider_data,omitempty" gorm:"type:jsonb"` // Additional provider-specific data
	// MFA fields
	MFASecret      *string    `json:"mfa_secret,omitempty"`                         // TOTP secret for MFA
	MFAEnabled     bool       `json:"mfa_enabled" gorm:"default:false"`             // Whether MFA is enabled for this user
	MFABackupCodes []string   `json:"mfa_backup_codes,omitempty" gorm:"type:jsonb"` // Backup codes for MFA recovery
	MFALastUsed    *time.Time `json:"mfa_last_used,omitempty"`                      // Last time MFA was used
	// Deactivation fields (for user management instead of hard delete)
	DeactivatedAt  *time.Time `json:"deactivated_at,omitempty"`
	DeactivatedBy  *uuid.UUID `json:"deactivated_by,omitempty"`
	// Team collaboration fields
	Teams []TeamMembership `json:"teams,omitempty" gorm:"foreignKey:UserID"`
	// Profile-related associations
	Skills       []UserSkill            `json:"skills,omitempty" gorm:"foreignKey:UserID"`
	Achievements []UserAchievement      `json:"achievements,omitempty" gorm:"foreignKey:UserID"`
	Activity     []UserActivity         `json:"activity,omitempty" gorm:"foreignKey:UserID"`
	Settings     map[string]interface{} `json:"settings,omitempty" gorm:"type:jsonb;default:'{}'"` // Profile settings (visibility, notifications, privacy)
	// Online status tracking
	LastActiveAt *time.Time             `json:"last_active_at,omitempty" gorm:"column:last_active_at;index"` // Last time user was active (for online status)
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserSearchHit is a public-safe row for username autocomplete (no email or tenant).
type UserSearchHit struct {
	ID       uuid.UUID
	Username string
	Name     string
}

// UserProfileSettings represents the user's profile settings
// Stored in the User.Settings JSONB field
type UserProfileSettings struct {
	// Visibility settings
	ProfileVisibility string `json:"profileVisibility"` // "public", "followers", "private"
	ShowEmail         bool   `json:"showEmail"`
	ShowLocation      bool   `json:"showLocation"`
	ShowCompany       bool   `json:"showCompany"`
	ShowActivity      bool   `json:"showActivity"`
	ShowAnalytics     bool   `json:"showAnalytics"`

	// Notification settings
	EmailNotifications    bool `json:"emailNotifications"`
	PushNotifications     bool `json:"pushNotifications"`
	NotifyOnFollow        bool `json:"notifyOnFollow"`
	NotifyOnMention       bool `json:"notifyOnMention"`
	NotifyOnFunctionUsage bool `json:"notifyOnFunctionUsage"`
	NotifyOnReviews       bool `json:"notifyOnReviews"`
	WeeklyDigest          bool `json:"weeklyDigest"`

	// Privacy settings
	AllowTagging   bool `json:"allowTagging"`
	AllowIndexing  bool `json:"allowIndexing"`
	ShowLastActive bool `json:"showLastActive"`
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
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name             string    `json:"name" gorm:"not null"`
	Plan             string    `json:"plan" gorm:"not null"`
	Status           string    `json:"status" gorm:"not null;default:'active'"` // 'active', 'suspended'
	StripeCustomerID *string   `json:"stripe_customer_id,omitempty" gorm:"column:stripe_customer_id;size:255"`
	// Session policy fields
	SessionMaxDuration *time.Duration `json:"session_max_duration,omitempty"`          // Maximum session duration
	SessionIdleTimeout *time.Duration `json:"session_idle_timeout,omitempty"`          // Idle timeout before session expires
	ConcurrentSessions *int           `json:"concurrent_sessions,omitempty"`           // Maximum concurrent sessions per user
	MFAPolicy          string         `json:"mfa_policy" gorm:"default:'optional'"`    // 'disabled', 'optional', 'required'
	SessionPersistence bool           `json:"session_persistence" gorm:"default:true"` // Whether sessions persist across browser restarts
	// Seat management fields
	SeatGracePeriodEnd *time.Time    `json:"seat_grace_period_end,omitempty"`   // Grace period end after downgrade
	SeatWarningSentAt *time.Time    `json:"seat_warning_sent_at,omitempty"`    // Last seat warning notification
	Users              []User         `json:"users,omitempty" gorm:"foreignKey:TenantID"`
	Apps               []App          `json:"apps,omitempty" gorm:"foreignKey:TenantID"`
	Teams              []Team         `json:"teams,omitempty" gorm:"foreignKey:TenantID"`
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
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

// UserSkill represents a skill or expertise area for a user
type UserSkill struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Level     string    `json:"level" gorm:"size:20;not null;default:'intermediate'"` // beginner, intermediate, advanced, expert
	Category  string    `json:"category" gorm:"size:50"`                              // language, framework, tool, platform, soft
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (UserSkill) TableName() string {
	return "user_skills"
}

// Achievement represents a badge/achievement definition
type Achievement struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug             string    `json:"slug" gorm:"uniqueIndex;size:100;not null"`
	Name             string    `json:"name" gorm:"size:255;not null"`
	Description      string    `json:"description" gorm:"type:text;not null"`
	Icon             string    `json:"icon" gorm:"size:100"`
	Color            string    `json:"color" gorm:"size:20"`
	Category         string    `json:"category" gorm:"size:50;not null"`         // publisher, community, usage, milestone
	RequirementType  string    `json:"requirement_type" gorm:"size:50;not null"` // function_count, execution_count, rating, days_active
	RequirementValue int       `json:"requirement_value" gorm:"not null"`
	Points           int       `json:"points" gorm:"default:0"`
	IsHidden         bool      `json:"is_hidden" gorm:"default:false"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Achievement) TableName() string {
	return "achievements"
}

// UserAchievement represents an achievement earned by a user
type UserAchievement struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	User          *User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AchievementID uuid.UUID              `json:"achievement_id" gorm:"type:uuid;not null;index"`
	Achievement   *Achievement           `json:"achievement,omitempty" gorm:"foreignKey:AchievementID"`
	EarnedAt      time.Time              `json:"earned_at" gorm:"autoCreateTime"`
	Progress      int                    `json:"progress" gorm:"default:0"`
	IsCompleted   bool                   `json:"is_completed" gorm:"default:false"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (UserAchievement) TableName() string {
	return "user_achievements"
}

// UserActivity represents an activity feed item for a user
type UserActivity struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	User         *User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ActivityType string                 `json:"activity_type" gorm:"size:50;not null"` // function_published, function_updated, badge_earned, profile_updated
	Title        string                 `json:"title" gorm:"size:255;not null"`
	Description  string                 `json:"description" gorm:"type:text"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	IsPublic     bool                   `json:"is_public" gorm:"default:true"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

func (UserActivity) TableName() string {
	return "user_activity"
}

// SocialLinks represents social media links for a user (stored as JSON)
type SocialLinks struct {
	Twitter  string `json:"twitter,omitempty"`
	Github   string `json:"github,omitempty"`
	LinkedIn string `json:"linkedin,omitempty"`
}

// EmailEvent represents an email delivery event from Resend webhooks
type EmailEvent struct {
	ID           int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       *uuid.UUID             `json:"user_id,omitempty" gorm:"type:uuid;index"`
	User         *User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	UserEmail    string                 `json:"user_email" gorm:"size:255;index"`
	EmailID      string                 `json:"email_id" gorm:"size:255;index"`          // Resend email ID for deduplication
	EventType    string                 `json:"event_type" gorm:"size:50;index"`         // email.sent, email.delivered, email.bounced, email.complained, etc.
	EventData    map[string]interface{} `json:"event_data,omitempty" gorm:"type:jsonb"`  // Raw webhook data
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`    // Additional metadata (alias for EventData for compatibility)
	BounceReason string                 `json:"bounce_reason,omitempty" gorm:"size:255"` // Bounce reason for bounce events
	Timestamp    time.Time              `json:"timestamp" gorm:"index"`
	Reviewed     bool                   `json:"reviewed" gorm:"default:false;index"`
	ReviewedBy   *uuid.UUID             `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt   *time.Time             `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EmailEvent) TableName() string {
	return "email_events"
}
