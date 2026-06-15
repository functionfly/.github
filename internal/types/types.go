// Package types contains shared types used across the storage and domain packages.
// This package exists to break import cycles between storage and domain packages.
// Types here are used by domain interfaces and storage models.
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// JSONMap represents a JSON map that can be stored in JSONB columns
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database reading
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}

	*j = JSONMap(m)
	return nil
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Heap   uint64 `json:"heap"`
	Stack  uint64 `json:"stack"`
	System uint64 `json:"system"`
}

// Value implements the driver.Valuer interface for database storage
func (m MemoryStats) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database reading
func (m *MemoryStats) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into MemoryStats", value)
	}

	return json.Unmarshal(bytes, m)
}

// PgNotification represents a PostgreSQL notification
type PgNotification struct {
	PID     int    `json:"pid"`
	Channel string `json:"channel"`
	Payload string `json:"payload"`
}

// User represents a user in the system
type User struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant        *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Username      *string   `json:"username,omitempty" gorm:"uniqueIndex;size:255"`
	Email         string    `json:"email" gorm:"uniqueIndex;not null"`
	Name          string    `json:"name,omitempty" gorm:"size:255"`
	PasswordHash  string    `json:"password_hash" gorm:"column:password_hash"`
	Role          string    `json:"role,omitempty" gorm:"size:50"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	TokenVersion  int        `json:"token_version,omitempty" gorm:"default:0"`
	CompanyName   *string    `json:"company_name,omitempty" gorm:"size:255"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty" gorm:"column:date_of_birth;type:date"`
	Bio           *string    `json:"bio,omitempty" gorm:"type:text"`
	ProfileNumber *int `json:"profile_number,omitempty" gorm:"column:profile_number;uniqueIndex"`
	Location      *string    `json:"location,omitempty" gorm:"size:255"`
	Website       *string    `json:"website,omitempty" gorm:"size:500"`
	JobTitle      *string    `json:"job_title,omitempty" gorm:"size:255"`
	SocialLinks   JSONMap    `json:"social_links,omitempty" gorm:"type:jsonb;default:'{}'"`
	TwitterURL    *string    `json:"twitter_url,omitempty" gorm:"column:twitter_url;size:500"`
	GithubURL     *string    `json:"github_url,omitempty" gorm:"column:github_url;size:500"`
	LinkedInURL   *string    `json:"linkedin_url,omitempty" gorm:"column:linkedin_url;size:500"`
	CoverImageURL *string    `json:"cover_image_url,omitempty" gorm:"column:cover_image_url;size:500"`
	VerificationToken     *string    `json:"verification_token,omitempty"`
	VerificationExpiresAt *time.Time `json:"verification_expires_at,omitempty"`
	Provider      *string `json:"provider,omitempty"`
	ProviderID    *string `json:"provider_id,omitempty"`
	ProviderData  JSONMap `json:"provider_data,omitempty" gorm:"type:jsonb"`
	MFASecret     *string    `json:"mfa_secret,omitempty"`
	MFAEnabled    bool       `json:"mfa_enabled" gorm:"default:false"`
	MFABackupCodes []string   `json:"mfa_backup_codes,omitempty" gorm:"type:jsonb"`
	MFALastUsed   *time.Time `json:"mfa_last_used,omitempty"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
	DeactivatedBy *uuid.UUID `json:"deactivated_by,omitempty"`
	Teams         []TeamMembership `json:"teams,omitempty" gorm:"foreignKey:UserID"`
	Skills        []UserSkill       `json:"skills,omitempty" gorm:"foreignKey:UserID"`
	Achievements  []UserAchievement `json:"achievements,omitempty" gorm:"foreignKey:UserID"`
	Activity      []UserActivity    `json:"activity,omitempty" gorm:"foreignKey:UserID"`
	Settings      JSONMap           `json:"settings,omitempty" gorm:"type:jsonb;default:'{}'"`
	LastActiveAt  *time.Time `json:"last_active_at,omitempty" gorm:"column:last_active_at;index"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserSearchHit is a public safe row for username autocomplete
type UserSearchHit struct {
	ID       uuid.UUID
	Username string
	Name     string
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
	Status           string    `json:"status" gorm:"not null;default:'active'"`
	StripeCustomerID *string   `json:"stripe_customer_id,omitempty" gorm:"column:stripe_customer_id;size:255"`
	SessionMaxDuration *time.Duration `json:"session_max_duration,omitempty"`
	SessionIdleTimeout *time.Duration `json:"session_idle_timeout,omitempty"`
	ConcurrentSessions *int           `json:"concurrent_sessions,omitempty"`
	MFAPolicy         string           `json:"mfa_policy" gorm:"default:'optional'"`
	SessionPersistence bool           `json:"session_persistence" gorm:"default:true"`
	SeatGracePeriodEnd *time.Time `json:"seat_grace_period_end,omitempty"`
	SeatWarningSentAt  *time.Time `json:"seat_warning_sent_at,omitempty"`
	BillingCountry      *string   `json:"billing_country,omitempty" gorm:"column:billing_country;size:2"`
	BillingState        *string   `json:"billing_state,omitempty" gorm:"column:billing_state;size:50"`
	BillingPostalCode   *string   `json:"billing_postal_code,omitempty" gorm:"column:billing_postal_code;size:20"`
	TaxID               *string   `json:"tax_id,omitempty" gorm:"column:tax_id;size:50"`
	TaxIDType           *string   `json:"tax_id_type,omitempty" gorm:"column:tax_id_type;size:20"`
	TaxStatus           string    `json:"tax_status" gorm:"column:tax_status;default:'pending';size:20"`
	TaxExempt           bool      `json:"tax_exempt" gorm:"column:tax_exempt;default:false"`
	StripeTaxLocationID *string   `json:"stripe_tax_location_id,omitempty" gorm:"column:stripe_tax_location_id;size:255"`
	StripeCustomerTaxID *string   `json:"stripe_customer_tax_id,omitempty" gorm:"column:stripe_customer_tax_id;size:255"`
	Users               []User    `json:"users,omitempty" gorm:"foreignKey:TenantID"`
	Apps                []App     `json:"apps,omitempty" gorm:"foreignKey:TenantID"`
	Teams               []Team    `json:"teams,omitempty" gorm:"foreignKey:TenantID"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Team represents a team within a tenant
type Team struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID        `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant      *Tenant          `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Name        string           `json:"name" gorm:"not null"`
	Slug        string           `json:"slug" gorm:"not null;size:100;uniqueIndex:idx_team_tenant_slug,priority:2"`
	Description string           `json:"description"`
	Visibility  string           `json:"visibility" gorm:"default:'private';size:20"`
	CreatedBy   uuid.UUID       `json:"created_by" gorm:"type:uuid;not null"`
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
	Role    string    `json:"role" gorm:"not null"`
	AddedBy uuid.UUID `json:"added_by" gorm:"type:uuid;not null"`
	AddedAt time.Time `json:"added_at" gorm:"autoCreateTime"`
}

// TeamPermission represents permissions granted to a team for specific resources
type TeamPermission struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID       uuid.UUID `json:"team_id" gorm:"type:uuid;not null"`
	Team         *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	ResourceType string    `json:"resource_type" gorm:"not null"`
	ResourceID   uuid.UUID `json:"resource_id" gorm:"type:uuid;not null"`
	Permissions  string    `json:"permissions" gorm:"not null"`
	GrantedBy    uuid.UUID `json:"granted_by" gorm:"type:uuid;not null"`
	GrantedAt    time.Time `json:"granted_at" gorm:"autoCreateTime"`
}

// TeamInvite represents a team invitation during onboarding
type TeamInvite struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID     uuid.UUID  `json:"team_id" gorm:"type:uuid;not null;index"`
	Email      string     `json:"email" gorm:"not null"`
	Token      string     `json:"token" gorm:"uniqueIndex;not null"`
	Role       string     `json:"role" gorm:"not null"`
	InvitedBy  uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`
	Message    string     `json:"message,omitempty"`
	Status     string     `json:"status" gorm:"default:'pending'"`
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
	ID         string     `json:"id" gorm:"type:varchar(255);primaryKey"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Provider   string     `json:"provider" gorm:"not null;index"`
	Token      string     `json:"token" gorm:"not null"`
	Status     string     `json:"status" gorm:"not null"`
	IsShared   bool       `json:"is_shared" gorm:"default:false"`
	TeamID     *string    `json:"team_id,omitempty" gorm:"type:varchar(255);index"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" gorm:"index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Provider) TableName() string {
	return "providers"
}

// ProviderSettings represents platform-wide provider maintenance settings
type ProviderSettings struct {
	Provider       string     `json:"provider" gorm:"primaryKey"`
	Disabled       bool       `json:"disabled" gorm:"default:false"`
	DisabledReason *string    `json:"disabled_reason,omitempty"`
	DisabledAt     *time.Time `json:"disabled_at,omitempty"`
	DisabledBy     *string    `json:"disabled_by,omitempty"`
}

func (ProviderSettings) TableName() string {
	return "provider_settings"
}

// UserSkill represents a skill or expertise area for a user
type UserSkill struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User      *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Level     string    `json:"level" gorm:"size:20;not null;default:'intermediate'"`
	Category  string    `json:"category" gorm:"size:50"`
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
	Category         string    `json:"category" gorm:"size:50;not null"`
	RequirementType  string    `json:"requirement_type" gorm:"size:50;not null"`
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
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index"`
	User          *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AchievementID uuid.UUID    `json:"achievement_id" gorm:"type:uuid;not null;index"`
	Achievement   *Achievement `json:"achievement,omitempty" gorm:"foreignKey:AchievementID"`
	EarnedAt      time.Time    `json:"earned_at" gorm:"autoCreateTime"`
	Progress      int          `json:"progress" gorm:"default:0"`
	IsCompleted   bool         `json:"is_completed" gorm:"default:false"`
	Metadata      JSONMap      `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (UserAchievement) TableName() string {
	return "user_achievements"
}

// UserActivity represents an activity feed item for a user
type UserActivity struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	User         *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ActivityType string    `json:"activity_type" gorm:"size:50;not null"`
	Title        string    `json:"title" gorm:"size:255;not null"`
	Description  string    `json:"description" gorm:"type:text"`
	Metadata     JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb"`
	IsPublic     bool      `json:"is_public" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserActivity) TableName() string {
	return "user_activity"
}

// EmailEvent represents an email delivery event from Resend webhooks
type EmailEvent struct {
	ID           int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	User         *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	UserEmail    string     `json:"user_email" gorm:"size:255;index"`
	EmailID      string     `json:"email_id" gorm:"size:255;index"`
	EventType    string     `json:"event_type" gorm:"size:50;index"`
	EventData    JSONMap    `json:"event_data,omitempty" gorm:"type:jsonb"`
	Metadata     JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb"`
	BounceReason string     `json:"bounce_reason,omitempty" gorm:"size:255"`
	Timestamp    time.Time  `json:"timestamp" gorm:"index"`
	Reviewed     bool       `json:"reviewed" gorm:"default:false;index"`
	ReviewedBy   *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EmailEvent) TableName() string {
	return "email_events"
}

// NewsletterSubscriber represents a newsletter subscription
type NewsletterSubscriber struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email              string     `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Name               string     `json:"name,omitempty" gorm:"size:255"`
	Status             string     `json:"status" gorm:"size:20;not null;default:'active'"`
	Source             string     `json:"source,omitempty" gorm:"size:50"`
	IPAddress          string     `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent          string     `json:"user_agent,omitempty" gorm:"size:500"`
	ConfirmationToken  *string    `json:"confirmation_token,omitempty" gorm:"size:255;index"`
	SubscribedAt       time.Time  `json:"subscribed_at" gorm:"autoCreateTime"`
	UnsubscribedAt     *time.Time `json:"unsubscribed_at,omitempty"`
	ConfirmedAt        *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (NewsletterSubscriber) TableName() string {
	return "newsletter_subscribers"
}

// NewsletterCampaign represents a newsletter campaign
type NewsletterCampaign struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Subject      string     `json:"subject" gorm:"not null;size:255"`
	PreviewText  string     `json:"preview_text,omitempty" gorm:"size:255"`
	Content      string     `json:"content" gorm:"type:text;not null"`
	HTMLContent  string     `json:"html_content" gorm:"type:text"`
	Status       string     `json:"status" gorm:"size:20;not null;default:'draft'"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
	SentCount    int        `json:"sent_count" gorm:"default:0"`
	OpenCount    int        `json:"open_count" gorm:"default:0"`
	ClickCount   int        `json:"click_count" gorm:"default:0"`
	BounceCount  int        `json:"bounce_count" gorm:"default:0"`
	ErrorMessage string     `json:"error_message,omitempty" gorm:"size:500"`
	Metadata     JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (NewsletterCampaign) TableName() string {
	return "newsletter_campaigns"
}

// NewsletterCampaignEmail tracks which subscribers received which campaign
type NewsletterCampaignEmail struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CampaignID   uuid.UUID  `json:"campaign_id" gorm:"type:uuid;not null;index"`
	SubscriberID uuid.UUID  `json:"subscriber_id" gorm:"type:uuid;not null;index"`
	EmailID      string     `json:"email_id,omitempty" gorm:"size:255"`
	Status       string     `json:"status" gorm:"size:20;not null;default:'pending'"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	OpenedAt     *time.Time `json:"opened_at,omitempty"`
	ClickedAt    *time.Time `json:"clicked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (NewsletterCampaignEmail) TableName() string {
	return "newsletter_campaign_emails"
}

// UsernameChangeHistory tracks all username changes for a user
type UsernameChangeHistory struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	OldUsername string    `json:"old_username" gorm:"size:255;not null"`
	NewUsername string    `json:"new_username" gorm:"size:255;not null"`
	ChangedAt   time.Time `json:"changed_at" gorm:"not null;index"`
	ChangedBy   uuid.UUID `json:"changed_by" gorm:"type:uuid"`
	WasEarlyChange  bool    `json:"was_early_change" gorm:"default:false"`
	FeePaidCents    int     `json:"fee_paid_cents" gorm:"default:0"`
	FeeCurrency     string  `json:"fee_currency" gorm:"size:3;default:'USD'"`
	StripePaymentID *string `json:"stripe_payment_id,omitempty" gorm:"size:255"`
	IPAddress string `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent string `json:"user_agent,omitempty" gorm:"size:500"`
}

func (UsernameChangeHistory) TableName() string {
	return "username_change_history"
}

// ExecutionRetentionSettings represents the database model for retention configuration
type ExecutionRetentionSettings struct {
	ID                            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionRetentionDays        int        `gorm:"column:execution_retention_days;not null;default:90"`
	PublicExecutionRetentionDays  int        `gorm:"column:public_execution_retention_days;not null;default:30"`
	ResourceUsageRetentionDays    int        `gorm:"column:resource_usage_retention_days;not null;default:90"`
	MEGRecordRetentionDays        int        `gorm:"column:meg_record_retention_days;not null;default:365"`
	DriftReportRetentionDays      int        `gorm:"column:drift_report_retention_days;not null;default:365"`
	ExecutionCertRetentionDays    int        `gorm:"column:execution_cert_retention_days;not null;default:365"`
	CleanupIntervalMinutes        int        `gorm:"column:cleanup_interval_minutes;not null;default:1440"`
	BatchSize                     int        `gorm:"column:batch_size;not null;default:1000"`
	VerboseLogging                bool       `gorm:"column:verbose_logging;not null;default:false"`
	CreatedAt                     time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt                     time.Time  `gorm:"column:updated_at;not null;default:now()"`
	UpdatedBy                     *uuid.UUID `gorm:"column:updated_by;type:uuid"`
	IsActive                      bool       `gorm:"column:is_active;not null;default:true;unique"`
}

func (ExecutionRetentionSettings) TableName() string {
	return "execution_retention_settings"
}

// ExecutionRetentionSettingsUpdate represents updateable fields for retention settings
type ExecutionRetentionSettingsUpdate struct {
	ExecutionRetentionDays        *int       `json:"execution_retention_days,omitempty"`
	PublicExecutionRetentionDays *int       `json:"public_execution_retention_days,omitempty"`
	ResourceUsageRetentionDays   *int       `json:"resource_usage_retention_days,omitempty"`
	MEGRecordRetentionDays       *int       `json:"meg_record_retention_days,omitempty"`
	DriftReportRetentionDays     *int       `json:"drift_report_retention_days,omitempty"`
	ExecutionCertRetentionDays   *int       `json:"execution_cert_retention_days,omitempty"`
	CleanupIntervalMinutes       *int       `json:"cleanup_interval_minutes,omitempty"`
	BatchSize                    *int       `json:"batch_size,omitempty"`
	VerboseLogging               *bool      `json:"verbose_logging,omitempty"`
	UpdatedBy                    *uuid.UUID `json:"updated_by,omitempty"`
}

// MagicLink represents a magic link token for passwordless authentication
type MagicLink struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Token        string     `json:"token" gorm:"uniqueIndex;not null;size:255"`
	Email        string     `json:"email" gorm:"not null;index;size:255"`
	UserID       *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	User         *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Used         bool       `json:"used" gorm:"default:false;not null"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null;index"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	IPCreated    string     `json:"ip_created,omitempty" gorm:"size:45"`
	UserAgent    string     `json:"user_agent,omitempty" gorm:"size:500"`
	RedirectPath string     `json:"redirect_path,omitempty" gorm:"size:255"`
}

func (MagicLink) TableName() string {
	return "magic_links"
}

func (m *MagicLink) IsExpired() bool {
	return time.Now().After(m.ExpiresAt)
}

func (m *MagicLink) CanUse() bool {
	return !m.Used && !m.IsExpired()
}

// PendingUsernameChange represents a username change waiting for payment completion
type PendingUsernameChange struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	OldUsername       string     `json:"old_username" gorm:"size:255;not null"`
	NewUsername       string     `json:"new_username" gorm:"size:255;not null"`
	Status            string     `json:"status" gorm:"size:20;default:'pending'"`
	CheckoutSessionID string     `json:"checkout_session_id,omitempty" gorm:"size:255;index"`
	FeeCents          int        `json:"fee_cents" gorm:"default:0"`
	FeeCurrency       string     `json:"fee_currency" gorm:"size:3;default:'USD'"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	IPAddress         string     `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent         string     `json:"user_agent,omitempty" gorm:"size:500"`
}

func (PendingUsernameChange) TableName() string {
	return "pending_username_changes"
}

func (p *PendingUsernameChange) IsExpired() bool {
	return time.Since(p.CreatedAt) > 24*time.Hour
}

func (p *PendingUsernameChange) CanComplete() bool {
	return p.Status == "pending" && !p.IsExpired()
}

// Session represents a user session with MFA verification status
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	SessionToken string     `json:"session_token" db:"session_token"`
	MFAVerified  bool       `json:"mfa_verified" db:"mfa_verified"`
	MFALastUsed  *time.Time `json:"mfa_last_used,omitempty" db:"mfa_last_used"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	LastActivity time.Time  `json:"last_activity" db:"last_activity"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash string     `json:"token_hash" db:"token_hash"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	Revoked   bool       `json:"revoked" db:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// LoginAttempt represents a login attempt (successful or failed) for account lockout protection
type LoginAttempt struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	Successful   bool       `json:"successful" db:"successful"`
	AttemptedAt  time.Time  `json:"attempted_at" db:"attempted_at"`
	LockoutUntil *time.Time `json:"lockout_until,omitempty" db:"lockout_until"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// AuthEvent represents an authentication event for security auditing
type AuthEvent struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	UserID        *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	TenantID      *uuid.UUID             `json:"tenant_id,omitempty" db:"tenant_id"`
	EventType     string                 `json:"event_type" db:"event_type"`
	Success       bool                   `json:"success" db:"success"`
	FailureReason *string                `json:"failure_reason,omitempty" db:"failure_reason"`
	IPAddress     string                 `json:"ip_address" db:"ip_address"`
	UserAgent     string                 `json:"user_agent" db:"user_agent"`
	LocationInfo  map[string]interface{} `json:"location_info,omitempty" db:"location_info"`
	SessionID     *uuid.UUID             `json:"session_id,omitempty" db:"session_id"`
	Provider      *string                `json:"provider,omitempty" db:"provider"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	SecurityFlags map[string]interface{} `json:"security_flags,omitempty" db:"security_flags"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// TenantAuthSettings represents per-tenant authentication configuration
type TenantAuthSettings struct {
	ID                       uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID                 uuid.UUID       `json:"tenant_id" gorm:"type:uuid;uniqueIndex;not null"`
	MFARequired              bool            `json:"mfa_required" gorm:"not null;default:false"`
	MFAMode                  string          `json:"mfa_mode" gorm:"type:varchar(20);not null;default:'optional'"`
	PasswordPolicy           json.RawMessage `json:"password_policy" gorm:"type:jsonb;not null;default:'{}'"`
	SessionTimeoutMinutes    int             `json:"session_timeout_minutes" gorm:"not null;default:480"`
	IPAllowlistEnabled       bool            `json:"ip_allowlist_enabled" gorm:"not null;default:false"`
	IPAllowlist              json.RawMessage `json:"ip_allowlist" gorm:"type:jsonb;not null;default:'[]'"`
	AllowedDomains           json.RawMessage `json:"allowed_domains" gorm:"type:jsonb;not null;default:'[]'"`
	SSOProvider              string          `json:"sso_provider" gorm:"type:varchar(20);not null;default:'none'"`
	SAMLMetadataURL           *string         `json:"saml_metadata_url,omitempty" gorm:"type:text"`
	SAMLEntityID             *string         `json:"saml_entity_id,omitempty" gorm:"type:text"`
	SAMLCertificate          *string         `json:"saml_certificate,omitempty" gorm:"type:text"`
	SAMLPrivateKey           *string         `json:"saml_private_key,omitempty" gorm:"type:text"`
	UseCustomBranding        bool            `json:"use_custom_branding" gorm:"not null;default:false"`
	EmailFromName            string          `json:"email_from_name" gorm:"type:varchar(100);not null;default:'FunctionFly'"`
	EmailFromAddress         string          `json:"email_from_address" gorm:"type:varchar(255);not null;default:'noreply@functionfly.com'"`
	RequireEmailVerification bool            `json:"require_email_verification" gorm:"not null;default:true"`
	AllowPasswordLogin       bool            `json:"allow_password_login" gorm:"not null;default:true"`
	AllowMagicLink           bool            `json:"allow_magic_link" gorm:"not null;default:true"`
	MaxLoginAttempts         int             `json:"max_login_attempts" gorm:"not null;default:5"`
	LockoutDurationMinutes   int             `json:"lockout_duration_minutes" gorm:"not null;default:15"`
	CreatedAt                time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time       `json:"updated_at" gorm:"autoUpdateTime"`

	Tenant            *Tenant               `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	OAuthProviders    []*TenantOAuthProvider `json:"oauth_providers,omitempty" gorm:"foreignKey:TenantID"`
	InviteCodes      []*TenantInviteCode   `json:"invite_codes,omitempty" gorm:"foreignKey:TenantID"`
	Memberships      []*TenantMembership   `json:"memberships,omitempty" gorm:"foreignKey:TenantID"`
	AuthAuditLog     []*TenantAuthAuditLog `json:"auth_audit_log,omitempty" gorm:"foreignKey:TenantID"`
}

func (TenantAuthSettings) TableName() string {
	return "tenant_auth_settings"
}

// PasswordPolicy represents password requirements
type PasswordPolicy struct {
	MinLength          int  `json:"min_length"`
	RequireUppercase   bool `json:"require_uppercase"`
	RequireLowercase   bool `json:"require_lowercase"`
	RequireDigit       bool `json:"require_digit"`
	RequireSpecial     bool `json:"require_special"`
	MaxAgeDays         int  `json:"max_age_days,omitempty"`
	PreventReuseCount int  `json:"prevent_reuse_count,omitempty"`
}

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
	ID                        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID                 uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex:idx_tenant_provider"`
	Provider                 string         `json:"provider" gorm:"type:varchar(50);not null;uniqueIndex:idx_tenant_provider"`
	ClientID                 string         `json:"client_id" gorm:"type:varchar(255);not null"`
	EncryptedClientSecret    string         `json:"-" gorm:"type:text;not null"`
	EncryptedClientSecretIV  *string        `json:"-" gorm:"type:text"`
	EncryptedClientSecretTag *string        `json:"-" gorm:"type:text"`
	Enabled                  bool           `json:"enabled" gorm:"not null;default:true"`
	CallbackURL              *string        `json:"callback_url,omitempty" gorm:"type:text"`
	Scopes                   json.RawMessage `json:"scopes" gorm:"type:jsonb;not null;default:'[user:email read:user]'"`
	CreatedAt                time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	ClientSecret string `json:"client_secret,omitempty" gorm:"-"`
}

func (TenantOAuthProvider) TableName() string {
	return "tenant_oauth_providers"
}

// OAuthProvider constants
const (
	OAuthProviderGitHub     = "github"
	OAuthProviderGoogle     = "google"
	OAuthProviderMicrosoft  = "microsoft"
	OAuthProviderApple      = "apple"
	OAuthProviderGitLab     = "gitlab"
	OAuthProviderBitbucket  = "bitbucket"
)

func ValidOAuthProviders() []string {
	return []string{OAuthProviderGitHub, OAuthProviderGoogle, OAuthProviderMicrosoft, OAuthProviderApple, OAuthProviderGitLab, OAuthProviderBitbucket}
}

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

	Tenant    *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Inviter   *User   `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
	Acceptor  *User   `json:"acceptor,omitempty" gorm:"foreignKey:AcceptedBy"`
}

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

// Certification credential lifecycle status
const (
	CertCredentialStatusActive    = "active"
	CertCredentialStatusExpired   = "expired"
	CertCredentialStatusRevoked   = "revoked"
	CertCredentialStatusSuspended = "suspended"
)

// Certification exam lifecycle status
const (
	CertExamStatusDraft      = "draft"
	CertExamStatusScheduled  = "scheduled"
	CertExamStatusSubmitted  = "submitted"
	CertExamStatusGrading    = "grading"
)

// Affiliate referral lifecycle status
const (
	ReferralStatusPending   = "pending"
	ReferralStatusConverted = "converted"
	ReferralStatusQualified = "qualified"
	ReferralStatusCanceled  = "canceled"
)

// Affiliate commission lifecycle status
const (
	CommissionStatusPending  = "pending"
	CommissionStatusApproved = "approved"
)

func ValidRoles() []string {
	return []string{RoleTeamOwner, RoleTeamAdmin, RoleTeamMember, RoleTeamViewer}
}

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
	Status    string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`

	Tenant  *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	User    *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Inviter *User   `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

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

	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	User   *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (TenantAuthAuditLog) TableName() string {
	return "tenant_auth_audit_log"
}

// AuthAction constants for audit logging
const (
	AuthActionLoginSuccess       = "login.success"
	AuthActionLoginFailed       = "login.failed"
	AuthActionLogout            = "logout"
	AuthActionPasswordChanged   = "password.changed"
	AuthActionPasswordReset    = "password.reset"
	AuthActionEmailVerified    = "email.verified"
	AuthActionMFASetup         = "mfa.setup"
	AuthActionMFAEnabled       = "mfa.enabled"
	AuthActionMFADisabled      = "mfa.disabled"
	AuthActionMFAVerified      = "mfa.verified"
	AuthActionMFALogin         = "mfa.login"
	AuthActionOAuthConnected   = "oauth.connected"
	AuthActionOAuthDisconnected = "oauth.disconnected"
	AuthActionOAuthLogin       = "oauth.login"
	AuthActionInviteSent       = "invite.sent"
	AuthActionInviteAccepted   = "invite.accepted"
	AuthActionInviteRevoked    = "invite.revoked"
	AuthActionUserInvited     = "user.invited"
	AuthActionUserRemoved      = "user.removed"
	AuthActionRoleChanged     = "role.changed"
	AuthActionSessionCreated  = "session.created"
	AuthActionSessionRevoked  = "session.revoked"
	AuthActionAccountLocked   = "account.locked"
	AuthActionAccountUnlocked = "account.unlocked"
)

// MFAMode constants
const (
	MFAModeOptional  = "optional"
	MFAModeRequired  = "required"
	MFAModeEnforced  = "enforced"
)

// SSOProvider constants
const (
	SSOProviderNone = "none"
	SSOProviderSAML = "saml"
	SSOProviderOIDC = "oidc"
)

// PricingTier represents a billing pricing tier
type PricingTier struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	PriceCents       int         `json:"price_cents"`
	AnnualPriceCents *int        `json:"annual_price_cents,omitempty"`
	Currency         string      `json:"currency"`
	BillingCycle     string      `json:"billing_cycle"`
	Features         interface{} `json:"features"`
	IsActive         bool        `json:"is_active"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// Subscription represents a tenant's subscription
type Subscription struct {
	ID                   uuid.UUID    `json:"id"`
	TenantID             uuid.UUID    `json:"tenant_id"`
	PricingTierID        uuid.UUID    `json:"pricing_tier_id"`
	Status               string       `json:"status"`
	BillingCycle         string       `json:"billing_cycle"`
	StripeSubscriptionID string       `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   time.Time    `json:"current_period_start"`
	CurrentPeriodEnd     time.Time    `json:"current_period_end"`
	TrialEnd             *time.Time   `json:"trial_end,omitempty"`
	CancelAtPeriodEnd    bool         `json:"cancel_at_period_end"`
	CanceledAt           *time.Time   `json:"canceled_at,omitempty"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	PricingTier          *PricingTier `json:"pricing_tier,omitempty"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	SubscriptionID    *uuid.UUID `json:"subscription_id,omitempty"`
	Status            string     `json:"status"`
	AmountDueCents    int        `json:"amount_due_cents"`
	AmountPaidCents   int        `json:"amount_paid_cents"`
	Currency          string     `json:"currency"`
	StripeInvoiceID   *string    `json:"stripe_invoice_id,omitempty"`
	ExternalReference *string    `json:"external_reference,omitempty"`
	InvoicePdfURL     string     `json:"invoice_pdf_url,omitempty"`
	HostedInvoiceURL  string     `json:"hosted_invoice_url,omitempty"`
	PeriodStart       *time.Time `json:"period_start,omitempty"`
	PeriodEnd         *time.Time `json:"period_end,omitempty"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UsageEvent represents a usage event for billing
type UsageEvent struct {
	ID             uuid.UUID   `json:"id"`
	TenantID       uuid.UUID   `json:"tenant_id"`
	EventType      string      `json:"event_type"`
	Quantity       int         `json:"quantity"`
	UnitPriceCents *int        `json:"unit_price_cents,omitempty"`
	Metadata       interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time   `json:"timestamp"`
}

// UsageRollup represents aggregated usage data
type UsageRollup struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	EventType     string    `json:"event_type"`
	PeriodDate    time.Time `json:"period_date"`
	TotalQuantity int       `json:"total_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AggregatedBillingUsage represents aggregated usage for billing
type AggregatedBillingUsage struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	EventType      string    `json:"event_type"`
	FunctionID     uuid.UUID `json:"function_id"`
	FunctionName   string    `json:"function_name"`
	Author         string    `json:"author"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	TotalQuantity  int       `json:"total_quantity"`
	TotalCostCents int64     `json:"total_cost_cents"`
	// Legacy execution-aggregation fields preserved for billing_repository_aggregation.go
	TotalCalls     int     `json:"total_calls"`
	SuccessCalls   int     `json:"success_calls"`
	ErrorCalls     int     `json:"error_calls"`
	CachedCalls    int     `json:"cached_calls"`
	TotalDuration  int64   `json:"total_duration_ms"`
	AvgDuration    float64 `json:"avg_duration_ms"`
	TotalCPUTimeMs int64   `json:"total_cpu_time_ms"`
	TotalMemoryMB  int64   `json:"total_memory_mb"`
}

// Coupon represents a discount coupon
type Coupon struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	DiscountType   string     `json:"discount_type"`
	DiscountValue  int        `json:"discount_value"`
	MaxRedemptions *int       `json:"max_redemptions,omitempty"`
	TimesRedeemed  int        `json:"times_redeemed"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CouponRedemption represents a coupon redemption
type CouponRedemption struct {
	ID             uuid.UUID  `json:"id"`
	CouponID       uuid.UUID  `json:"coupon_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
	RedeemedAt     time.Time  `json:"redeemed_at"`
	Coupon         *Coupon    `json:"coupon,omitempty"`
}

// CostAllocationEntry represents a detailed cost allocation record for a function execution
type CostAllocationEntry struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	APIKeyID         uuid.UUID `json:"api_key_id,omitempty"`
	FunctionID       uuid.UUID `json:"function_id"`
	FunctionName     string    `json:"function_name"`
	FunctionAuthor   string    `json:"function_author"`
	ExecutionID      uuid.UUID `json:"execution_id"`
	ExecutionOutcome string    `json:"execution_outcome"`
	Cached           bool      `json:"cached"`

	DurationMs   int64 `json:"duration_ms"`
	CPUTimeMs    int64 `json:"cpu_time_ms"`
	MemoryUsedMB int64 `json:"memory_used_mb"`
	WallTimeMs   int64 `json:"wall_time_ms"`

	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	Region      string                 `json:"region"`
	Timestamp   time.Time              `json:"timestamp"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// CostAllocationSummary provides aggregated cost data by function
type CostAllocationSummary struct {
	FunctionID     uuid.UUID `json:"function_id"`
	FunctionName   string    `json:"function_name"`
	FunctionAuthor string    `json:"function_author"`

	TotalExecutions   int64 `json:"total_executions"`
	SuccessExecutions int64 `json:"success_executions"`
	ErrorExecutions   int64 `json:"error_executions"`
	CachedExecutions  int64 `json:"cached_executions"`

	TotalDurationMs   int64 `json:"total_duration_ms"`
	TotalCPUTimeMs    int64 `json:"total_cpu_time_ms"`
	TotalMemoryUsedMB int64 `json:"total_memory_used_mb"`

	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	AvgDurationMs float64 `json:"avg_duration_ms"`
	AvgCostCents  float64 `json:"avg_cost_cents"`
}

// TenantCostSummary provides tenant-level cost aggregation
type TenantCostSummary struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name,omitempty"`

	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	TotalExecutions int64 `json:"total_executions"`
	UniqueFunctions int   `json:"unique_functions"`

	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`
	TotalCostCents     int64 `json:"total_cost_cents"`

	FunctionSummaries []CostAllocationSummary `json:"function_summaries,omitempty"`
	DailyBreakdown     []DailyCostBreakdown     `json:"daily_breakdown,omitempty"`
}

// DailyCostBreakdown provides daily cost aggregation
type DailyCostBreakdown struct {
	Date       time.Time `json:"date"`
	Executions int64     `json:"executions"`
	CostCents  int64     `json:"cost_cents"`
}

// CostAllocationFilter provides filtering options for cost queries
type CostAllocationFilter struct {
	TenantID     *uuid.UUID
	FunctionID   *uuid.UUID
	FunctionName *string
	Author       *string
	StartDate    *time.Time
	EndDate      *time.Time
	Outcome      *string
	Cached       *bool
	Region       *string
	MinCostCents *int64
	MaxCostCents *int64
	Tags         map[string]string
}

// CostAllocationReport represents a comprehensive cost report
type CostAllocationReport struct {
	ReportID    uuid.UUID `json:"report_id"`
	GeneratedAt time.Time `json:"generated_at"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	TenantCount     int   `json:"tenant_count"`
	FunctionCount   int   `json:"function_count"`
	TotalExecutions int64 `json:"total_executions"`
	TotalCostCents  int64 `json:"total_cost_cents"`

	ChargebackEntries []CostAllocationChargeback `json:"chargeback_entries"`
	TenantSummaries    []TenantCostSummary        `json:"tenant_summaries"`
}

// CostAllocationChargeback represents a chargeback entry for internal billing
type CostAllocationChargeback struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	TenantName     string    `json:"tenant_name"`
	CostCenter     string    `json:"cost_center,omitempty"`
	Department     string    `json:"department,omitempty"`
	Project        string    `json:"project,omitempty"`
	TotalCostCents int64     `json:"total_cost_cents"`

	ExecutionCostCents int64 `json:"execution_cost_cents"`
	ComputeCostCents   int64 `json:"compute_cost_cents"`
	PlatformFeeCents   int64 `json:"platform_fee_cents"`
	DataTransferCents  int64 `json:"data_transfer_cents"`

	InvoicePeriod string    `json:"invoice_period"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// PricingBundle represents a Backend-in-a-Box pricing bundle
type PricingBundle struct {
	ID                    uuid.UUID      `json:"id"`
	Slug                  string         `json:"slug"`
	Name                  string         `json:"name"`
	DisplayName           string         `json:"display_name"`
	Description           string         `json:"description"`
	ShortDescription      string         `json:"short_description"`
	DisplayPriceCents     int            `json:"display_price_cents"`
	BillingInterval       string         `json:"billing_interval"`
	StripePriceID         string         `json:"stripe_price_id"`
	Icon                  string         `json:"icon"`
	Color                 string         `json:"color"`
	FeaturesIncluded      []string       `json:"features_included"`
	FeatureLimits         map[string]int `json:"feature_limits"`
	ProvisioningTemplates []string       `json:"provisioning_templates"`
	SortOrder             int            `json:"sort_order"`
	IsActive              bool           `json:"is_active"`
	IsPopular             bool           `json:"is_popular"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// FounderModeRegistration represents a founder mode (free until trigger) registration
type FounderModeRegistration struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	BundleID             uuid.UUID  `json:"bundle_id"`
	ModeType             string     `json:"mode_type"`
	StartedAt            time.Time  `json:"started_at"`
	EndsAt               *time.Time `json:"ends_at,omitempty"`
	FreeDays             int        `json:"free_days"`
	MRRThresholdCents   int        `json:"mrr_threshold_cents"`
	Status               string     `json:"status"`
	ConvertedToBundleID  *uuid.UUID `json:"converted_to_bundle_id,omitempty"`
	ConvertedAt          *time.Time `json:"converted_at,omitempty"`
	StripeSubscriptionID string     `json:"stripe_subscription_id,omitempty"`
	GracePeriodStartedAt *time.Time `json:"grace_period_started_at,omitempty"`
	GracePeriodEndsAt    *time.Time `json:"grace_period_ends_at,omitempty"`
	MaxUsersSeen         int        `json:"max_users_seen"`
	MaxMRRSeenCents      int        `json:"max_mrr_seen_cents"`
	MaxAPICallsMonthly   int        `json:"max_api_calls_monthly"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// BundleSubscription represents a subscription to a Backend-in-a-Box bundle
type BundleSubscription struct {
	ID                       uuid.UUID  `json:"id"`
	TenantID                 uuid.UUID  `json:"tenant_id"`
	BundleID                 uuid.UUID  `json:"bundle_id"`
	FounderModeID            *uuid.UUID `json:"founder_mode_id,omitempty"`
	ConvertedFromFounderMode bool       `json:"converted_from_founder_mode"`
	Status                   string     `json:"status"`
	StripeSubscriptionID     string     `json:"stripe_subscription_id,omitempty"`
	DefaultAppID             *uuid.UUID `json:"default_app_id,omitempty"`
	CurrentPeriodStart       time.Time  `json:"current_period_start"`
	CurrentPeriodEnd         time.Time  `json:"current_period_end"`
	CancelAtPeriodEnd        bool       `json:"cancel_at_period_end"`
	CanceledAt               *time.Time `json:"canceled_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// DeferredBillingConfig represents configuration for deferred billing triggers
type DeferredBillingConfig struct {
	ID                      uuid.UUID  `json:"id"`
	BundleID                uuid.UUID  `json:"bundle_id"`
	IsDefault               bool       `json:"is_default"`
	TriggerUserCount        *int       `json:"trigger_user_count,omitempty"`
	TriggerRevenueCents     *int       `json:"trigger_revenue_cents,omitempty"`
	TriggerAPICalls         *int       `json:"trigger_api_calls,omitempty"`
	TriggerDaysElapsed      *int       `json:"trigger_days_elapsed,omitempty"`
	GracePeriodDays         int        `json:"grace_period_days"`
	ConvertToBundleID       *uuid.UUID `json:"convert_to_bundle_id,omitempty"`
	WarningEmailTemplate    string     `json:"warning_email_template"`
	TriggerEmailTemplate    string     `json:"trigger_email_template"`
	ConversionEmailTemplate string     `json:"conversion_email_template"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// VerificationFee represents the fee structure for function verification
type VerificationFee struct {
	ID          uuid.UUID  `json:"id"`
	Level       string    `json:"level"`
	PriceCents  int       `json:"price_cents"`
	Currency    string    `json:"currency"`
	IsActive    bool      `json:"is_active"`
	MinPlan     *string   `json:"min_plan,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FunctionVerificationPayment represents a payment for function verification
type FunctionVerificationPayment struct {
	ID                      uuid.UUID  `json:"id"`
	FunctionID              uuid.UUID  `json:"function_id"`
	VerificationLevel      string     `json:"verification_level"`
	AmountCents            int        `json:"amount_cents"`
	Currency                string     `json:"currency"`
	Status                  string     `json:"status"`
	StripePaymentIntentID   *string    `json:"stripe_payment_intent_id,omitempty"`
	StripeCheckoutSessionID *string    `json:"stripe_checkout_session_id,omitempty"`
	TenantID                uuid.UUID  `json:"tenant_id"`
	PaidBy                  *uuid.UUID `json:"paid_by,omitempty"`
	VerificationJobID       *uuid.UUID `json:"verification_job_id,omitempty"`
	PaidAt                  *time.Time `json:"paid_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// PublisherEarning represents earnings from function sales
type PublisherEarning struct {
	ID                  uuid.UUID        `json:"id"`
	TenantID            uuid.UUID        `json:"tenant_id"`
	PublisherUserID     uuid.UUID        `json:"publisher_user_id"`
	FunctionID          *uuid.UUID       `json:"function_id,omitempty"`
	FunctionName        string           `json:"function_name,omitempty"`
	TransactionType     string           `json:"transaction_type"`
	AmountCents         int              `json:"amount_cents"`
	Currency            string           `json:"currency"`
	GrossAmountCents    int              `json:"gross_amount_cents"`
	PlatformFeeCents    int              `json:"platform_fee_cents"`
	NetAmountCents      int              `json:"net_amount_cents"`
	PlatformFeePercent  float64          `json:"platform_fee_percent"`
	Status              string           `json:"status"`
	StripePayoutID      *string          `json:"stripe_payout_id,omitempty"`
	PeriodMonth         *int             `json:"period_month,omitempty"`
	PeriodYear          *int             `json:"period_year,omitempty"`
	Metadata            json.RawMessage  `json:"metadata,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	EarnedAt            time.Time        `json:"earned_at"`
}

// AgentSubscription represents an agent-based subscription
type AgentSubscription struct {
	ID                   uuid.UUID  `json:"id"`
	AgentID              uuid.UUID  `json:"agent_id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	PlanName             string     `json:"plan_name"`
	PricePerAgentCents   int        `json:"price_per_agent_cents"`
	Currency             string     `json:"currency"`
	MaxAgents            int        `json:"max_agents"`
	Status               string     `json:"status"`
	CurrentPeriodStart   time.Time  `json:"current_period_start"`
	CurrentPeriodEnd     time.Time  `json:"current_period_end"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	LastPaymentStatus    *string    `json:"last_payment_status,omitempty"`
	LastPaymentAt        *time.Time `json:"last_payment_at,omitempty"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	CancelledAt          *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// AgentUsage represents usage tracking for agent billing
type AgentUsage struct {
	ID                 uuid.UUID  `json:"id"`
	AgentID            uuid.UUID  `json:"agent_id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	SubscriptionID     *uuid.UUID `json:"subscription_id,omitempty"`
	PeriodStart        time.Time  `json:"period_start"`
	PeriodEnd          time.Time  `json:"period_end"`
	TotalCalls         int        `json:"total_calls"`
	TotalExecutions    int        `json:"total_executions"`
	TotalErrors        int        `json:"total_errors"`
	TotalLatencyMs     int64      `json:"total_latency_ms"`
	BillableCalls      int        `json:"billable_calls"`
	OverageCalls       int        `json:"overage_calls"`
	EstimatedCostCents int        `json:"estimated_cost_cents"`
	Status             string     `json:"status"`
	StripeInvoiceID    *string    `json:"stripe_invoice_id,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// PlatformFee represents platform fees collected from marketplace transactions
type PlatformFee struct {
	ID                   uuid.UUID       `json:"id"`
	FeeType              string          `json:"fee_type"`
	SourceTransactionID  *uuid.UUID      `json:"source_transaction_id,omitempty"`
	SourceType           string          `json:"source_type"`
	GrossAmountCents     int             `json:"gross_amount_cents"`
	PlatformFeeCents    int             `json:"platform_fee_cents"`
	NetAmountCents      int             `json:"net_amount_cents"`
	PlatformFeePercent  float64         `json:"platform_fee_percent"`
	Currency            string          `json:"currency"`
	TenantID            *uuid.UUID      `json:"tenant_id,omitempty"`
	UserID              *uuid.UUID      `json:"user_id,omitempty"`
	FunctionID          *uuid.UUID      `json:"function_id,omitempty"`
	AgentID             *uuid.UUID      `json:"agent_id,omitempty"`
	Status              string          `json:"status"`
	StripeTransferID    *string         `json:"stripe_transfer_id,omitempty"`
	PaidOutAt           *time.Time      `json:"paid_out_at,omitempty"`
	PeriodMonth         *int            `json:"period_month,omitempty"`
	PeriodYear          *int            `json:"period_year,omitempty"`
	Metadata            json.RawMessage `json:"metadata,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

// PricingTierExtended extends PricingTier with new Moat fields
type PricingTierExtended struct {
	ID                    uuid.UUID       `json:"id"`
	Name                  string          `json:"name"`
	Description           string          `json:"description"`
	PriceCents            int             `json:"price_cents"`
	AnnualPriceCents      *int            `json:"annual_price_cents,omitempty"`
	Currency              string          `json:"currency"`
	BillingCycle          string          `json:"billing_cycle"`
	Features              json.RawMessage `json:"features"`
	IsActive              bool            `json:"is_active"`
	TierType              string          `json:"tier_type"`
	StripePriceID         *string         `json:"stripe_price_id,omitempty"`
	StripePriceIDAnnual   *string         `json:"stripe_price_id_annual,omitempty"`
	TrialDays             int             `json:"trial_days"`
	MaxAgents             int             `json:"max_agents"`
	MaxFunctions          int             `json:"max_functions"`
	MaxExecutionsPerMonth int             `json:"max_executions_per_month"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// StripeSyncEvent represents an event received from Stripe webhooks
type StripeSyncEvent struct {
	ID             uuid.UUID       `json:"id"`
	StripeEventID  string          `json:"stripe_event_id"`
	StripeObjectID string          `json:"stripe_object_id"`
	EventType      string          `json:"event_type"`
	EventData      json.RawMessage `json:"event_data"`
	TenantID       *uuid.UUID      `json:"tenant_id,omitempty"`
	Status         string          `json:"status"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	ProcessedAt    *time.Time      `json:"processed_at,omitempty"`
	RetryCount     int             `json:"retry_count"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// PaymentMethodInfoExtended represents stored payment method information
type PaymentMethodInfoExtended struct {
	ID                    uuid.UUID       `json:"id"`
	TenantID              uuid.UUID       `json:"tenant_id"`
	StripePaymentMethodID string          `json:"stripe_payment_method_id"`
	Brand                 string          `json:"brand"`
	Last4                 string          `json:"last4"`
	ExpMonth              int             `json:"exp_month"`
	ExpYear               int             `json:"exp_year"`
	IsDefault             bool            `json:"is_default"`
	BillingDetails        json.RawMessage `json:"billing_details,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// AgentTierPricing represents a database-driven agent subscription tier
type AgentTierPricing struct {
	ID                       uuid.UUID              `json:"id"`
	TierSlug                 string                 `json:"tier_slug"`
	DisplayName              string                 `json:"display_name"`
	Description              string                 `json:"description"`
	MonthlyPriceCents        int                    `json:"monthly_price_cents"`
	AnnualPriceCents         *int                   `json:"annual_price_cents,omitempty"`
	BaseCurrency             string                 `json:"base_currency"`
	RegionPricing            map[string]interface{} `json:"region_pricing,omitempty"`
	MaxAgents                int                    `json:"max_agents"`
	IncludedAICalls          int                    `json:"included_ai_calls"`
	IncludedExecutions       int                    `json:"included_executions"`
	IncludedStorageGB        int                    `json:"included_storage_gb"`
	OveragePricePer1000Cents int                    `json:"overage_price_per_1000_cents"`
	StripePriceIDMonthly     *string                `json:"stripe_price_id_monthly,omitempty"`
	StripePriceIDAnnual      *string                `json:"stripe_price_id_annual,omitempty"`
	FeaturesIncluded         []string               `json:"features_included"`
	IsActive                 bool                   `json:"is_active"`
	SortOrder                int                    `json:"sort_order"`
	PricingVariant           string                 `json:"pricing_variant"`
	ValidFrom                time.Time              `json:"valid_from"`
	ValidUntil               *time.Time             `json:"valid_until,omitempty"`
	CreatedAt                time.Time              `json:"created_at"`
	UpdatedAt                time.Time              `json:"updated_at"`
}

func (t *AgentTierPricing) GetMonthlyPrice(currencyCode string) int {
	if currencyCode == t.BaseCurrency {
		return t.MonthlyPriceCents
	}
	if t.RegionPricing != nil {
		if regionPrice, ok := t.RegionPricing[currencyCode]; ok {
			if priceMap, ok := regionPrice.(map[string]interface{}); ok {
				if monthly, ok := priceMap["monthly"].(float64); ok {
					return int(monthly)
				}
			}
		}
	}
	return t.MonthlyPriceCents
}

func (t *AgentTierPricing) GetAnnualPrice(currencyCode string) *int {
	if t.AnnualPriceCents == nil {
		return nil
	}
	if currencyCode == t.BaseCurrency {
		return t.AnnualPriceCents
	}
	if t.RegionPricing != nil {
		if regionPrice, ok := t.RegionPricing[currencyCode]; ok {
			if priceMap, ok := regionPrice.(map[string]interface{}); ok {
				if annual, ok := priceMap["annual"].(float64); ok {
					annualInt := int(annual)
					return &annualInt
				}
			}
		}
	}
	return t.AnnualPriceCents
}

func (t *AgentTierPricing) IsUnlimited(field string) bool {
	switch field {
	case "agents":
		return t.MaxAgents < 0
	case "ai_calls":
		return t.IncludedAICalls < 0
	case "executions":
		return t.IncludedExecutions < 0
	case "storage":
		return t.IncludedStorageGB < 0
	default:
		return false
	}
}

func (t *AgentTierPricing) IsValidAt(checkTime time.Time) bool {
	if checkTime.Before(t.ValidFrom) {
		return false
	}
	if t.ValidUntil != nil && checkTime.After(*t.ValidUntil) {
		return false
	}
	return t.IsActive
}

// CurrencyExchangeRate represents an exchange rate for currency conversion
type CurrencyExchangeRate struct {
	ID               uuid.UUID  `json:"id"`
	BaseCurrency     string     `json:"base_currency"`
	QuoteCurrency    string     `json:"quote_currency"`
	Rate             float64    `json:"rate"`
	RateNumerator    int64      `json:"rate_numerator"`
	RateDenominator  int64      `json:"rate_denominator"`
	Source           string     `json:"source"`
	SourceURL        *string    `json:"source_url,omitempty"`
	EffectiveDate    string     `json:"effective_date"`
	FetchedAt        *time.Time `json:"fetched_at,omitempty"`
	IsManualOverride bool       `json:"is_manual_override"`
	OverrideReason   *string    `json:"override_reason,omitempty"`
	IsStripeRate     bool       `json:"is_stripe_rate"`
	StripePrecision  string     `json:"stripe_precision"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

const DefaultRateDenominator int64 = 1_000_000

func (r *CurrencyExchangeRate) GetRateDenominator() int64 {
	if r.RateDenominator <= 0 {
		return DefaultRateDenominator
	}
	return r.RateDenominator
}

func (r *CurrencyExchangeRate) GetRateNumerator() int64 {
	if r.RateNumerator > 0 {
		return r.RateNumerator
	}
	denom := r.GetRateDenominator()
	return int64(r.Rate * float64(denom))
}

func (r *CurrencyExchangeRate) Convert(amountCents int) int {
	numerator := r.GetRateNumerator()
	denominator := r.GetRateDenominator()
	converted := (int64(amountCents) * numerator) / denominator
	return int(converted)
}

func (r *CurrencyExchangeRate) ConvertTo(amountCents int) int {
	return r.Convert(amountCents)
}

func (r *CurrencyExchangeRate) SetRateFromFloat(rate float64) {
	r.Rate = rate
	r.RateDenominator = DefaultRateDenominator
	r.RateNumerator = int64(rate * float64(DefaultRateDenominator))
}

// SupportedCurrency represents a currency that can be used in the system
type SupportedCurrency struct {
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Symbol             string    `json:"symbol"`
	SymbolPosition     string    `json:"symbol_position"`
	DecimalPlaces      int       `json:"decimal_places"`
	ThousandsSeparator string    `json:"thousands_separator"`
	DecimalSeparator   string    `json:"decimal_separator"`
	IsActive           bool      `json:"is_active"`
	IsStablecoin       bool      `json:"is_stripecoin"`
	ContractAddress    *string   `json:"contract_address,omitempty"`
	ChainID            *int      `json:"chain_id,omitempty"`
	DefaultCountry     *string   `json:"default_country,omitempty"`
	SupportedCountries []string  `json:"supported_countries,omitempty"`
	RoundingMode       string    `json:"rounding_mode"`
	MinimumChargeCents int       `json:"minimum_charge_cents"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (c *SupportedCurrency) FormatAmount(amountCents int) string {
	amount := float64(amountCents) / math.Pow(10, float64(c.DecimalPlaces))

	parts := strings.Split(fmt.Sprintf("%.*f", c.DecimalPlaces, amount), ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = c.DecimalSeparator + parts[1]
	}

	var result strings.Builder
	for i, ch := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteString(c.ThousandsSeparator)
		}
		result.WriteRune(ch)
	}
	result.WriteString(decPart)

	if c.SymbolPosition == "before" {
		return c.Symbol + result.String()
	}
	return result.String() + " " + c.Symbol
}

func (c *SupportedCurrency) ConvertToStripeAmount(cents int) int64 {
	if c.DecimalPlaces == 0 {
		return int64(cents / 100)
	}
	return int64(cents)
}

func (c *SupportedCurrency) ConvertFromStripeAmount(stripeAmount int64) int {
	if c.DecimalPlaces == 0 {
		return int(stripeAmount * 100)
	}
	return int(stripeAmount)
}

// SignupInviteCodeAdminList is a safe subset for admin API responses
type SignupInviteCodeAdminList struct {
	ID        uuid.UUID  `json:"id"`
	Label     string     `json:"label"`
	MaxUses   *int       `json:"maxUses,omitempty"`
	UsesCount int        `json:"usesCount"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedBy *uuid.UUID `json:"createdBy,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// CreditNote represents an accounting credit note for refunds
type CreditNote struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	InvoiceID       *uuid.UUID `json:"invoice_id,omitempty" gorm:"type:uuid;index"`
	ReferenceNumber string     `json:"reference_number" gorm:"not null;size:50;uniqueIndex"`
	Status          string     `json:"status" gorm:"not null;size:20;default:draft"`

	SubtotalCents   int `json:"subtotal_cents" gorm:"not null;default:0"`
	TaxCents        int `json:"tax_cents" gorm:"not null;default:0"`
	TotalCents      int `json:"total_cents" gorm:"not null;default:0"`

	Currency string `json:"currency" gorm:"not null;size:3;default:'USD'"`

	Reason          string `json:"reason" gorm:"size:255"`
	Description     string `json:"description,omitempty" gorm:"size:500"`

	PaymentRefundID *uuid.UUID `json:"payment_refund_id,omitempty" gorm:"type:uuid;index"`

	IssuedAt  *time.Time `json:"issued_at,omitempty"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
	VoidedAt  *time.Time `json:"voided_at,omitempty"`

	IssuedBy uuid.UUID `json:"issued_by" gorm:"type:uuid;not null"`
	Notes    string    `json:"notes,omitempty" gorm:"size:1000"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Invoice        *Invoice        `json:"invoice,omitempty" gorm:"-"`
	LineItems      []*CreditNoteLineItem `json:"line_items,omitempty" gorm:"-"`
	PaymentRefund  *PaymentRefund  `json:"payment_refund,omitempty" gorm:"-"`
}

// CreditNoteLineItem represents a line item on a credit note
type CreditNoteLineItem struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreditNoteID   uuid.UUID  `json:"credit_note_id" gorm:"type:uuid;not null;index"`
	Description    string     `json:"description" gorm:"not null;size:500"`
	Quantity       int        `json:"quantity" gorm:"not null;default:1"`
	UnitPriceCents int        `json:"unit_price_cents" gorm:"not null;default:0"`
	TaxCents       int        `json:"tax_cents" gorm:"not null;default:0"`
	AmountCents    int        `json:"amount_cents" gorm:"not null;default:0"`
	TotalCents     int        `json:"total_cents" gorm:"not null;default:0"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CreditNote) TableName() string {
	return "credit_notes"
}

func (CreditNoteLineItem) TableName() string {
	return "credit_note_line_items"
}

// CreditNoteStats represents credit note statistics
type CreditNoteStats struct {
	TotalCount         int             `json:"total_count"`
	TotalValueCents    int64           `json:"total_value_cents"`
	TotalCreditNotes   int64           `json:"total_credit_notes"`
	TotalCreditedCents int64           `json:"total_credited_cents"`
	ByStatus           map[string]int64 `json:"by_status"`
}

// UsageSummary represents a summary of usage for a period
type UsageSummary struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	TotalExecutions int64     `json:"total_executions"`
	TotalComputeMs  int64     `json:"total_compute_ms"`
	EstimatedCost   int       `json:"estimated_cost"`
	TotalCostCents  int64     `json:"total_cost_cents"`
	UniqueFunctions int       `json:"unique_functions"`
}

// FunctionUsageRollup represents function-specific usage rollup
type FunctionUsageRollup struct {
	FunctionID      uuid.UUID `json:"function_id"`
	FunctionName    string    `json:"function_name"`
	TotalCalls      int64     `json:"total_calls"`
	TotalExecutions int64     `json:"total_executions"`
	TotalDurationMs int64     `json:"total_duration_ms"`
	TotalCostCents  int64     `json:"total_cost_cents"`
}

// UsageByDay represents daily usage for dashboard
type UsageByDay struct {
	Time       string    `json:"time"`
	Date       time.Time `json:"date"`
	Value      int       `json:"value"`
	Executions int       `json:"executions"`
	CostCents  int       `json:"cost_cents"`
}

// ExecutionRateByHour represents hourly execution rate for dashboard
type ExecutionRateByHour struct {
	Time        string  `json:"time"`
	Hour        int     `json:"hour"`
	Rate        int     `json:"rate"`
	Value       int     `json:"value"`
	Executions  int     `json:"executions"`
	SuccessRate float64 `json:"success_rate"`
}

// DashboardActivityItem represents an activity item for dashboard
type DashboardActivityItem struct {
	ID           uuid.UUID `json:"id"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	FunctionID   string    `json:"function_id,omitempty"`
	FunctionName string    `json:"function_name,omitempty"`
	Metadata     JSONMap   `json:"metadata,omitempty"`
}

// DashboardMetrics represents dashboard metrics
type DashboardMetrics struct {
	TenantID             uuid.UUID `json:"tenant_id"`
	TotalExecutions      int64     `json:"total_executions"`
	TotalCostCents       int64     `json:"total_cost_cents"`
	ActiveFunctions      int       `json:"active_functions"`
	SuccessRate          float64   `json:"success_rate"`
	RequestsThisMonth    int64     `json:"requests_this_month"`
	RequestsPrevMonth    int64     `json:"requests_prev_month"`
	AvgLatencyMs         *float64  `json:"avg_latency_ms"`
	UptimePct            *float64  `json:"uptime_pct"`
	UptimePrevPct        *float64  `json:"uptime_prev_pct"`
	UptimeSparkline      []float64 `json:"uptime_sparkline,omitempty"`
	RequestsSparkline    []int64   `json:"requests_sparkline,omitempty"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// CreditNoteFilter represents filter options for credit note queries
type CreditNoteFilter struct {
	TenantID        *uuid.UUID
	InvoiceID       *uuid.UUID
	PaymentRefundID *uuid.UUID
	Status          *string
	StartDate       *time.Time
	EndDate         *time.Time
	Limit           int
	Offset          int
}

// PaymentRefund represents a refund for a payment
type PaymentRefund struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	InvoiceID       *uuid.UUID `json:"invoice_id,omitempty" gorm:"type:uuid;index"`
	AmountCents     int        `json:"amount_cents" gorm:"not null"`
	Currency       string     `json:"currency" gorm:"size:3;default:'USD'"`
	Reason         string     `json:"reason" gorm:"size:255"`
	Status         string     `json:"status" gorm:"size:20;not null;default:'pending'"`
	StripeRefundID *string    `json:"stripe_refund_id,omitempty" gorm:"size:255"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (PaymentRefund) TableName() string {
	return "payment_refunds"
}

// DailyIncidentCount represents incident count by day
type DailyIncidentCount struct {
	Date     time.Time `json:"date"`
	Count    int       `json:"count"`
	Critical int      `json:"critical"`
	High     int      `json:"high"`
	Medium   int      `json:"medium"`
	Low      int      `json:"low"`
}

// App represents an application
type App struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Name      string    `json:"name" gorm:"not null"`
	Slug      string    `json:"slug" gorm:"uniqueIndex;not null"`
	Backends  []Backend `json:"backends,omitempty" gorm:"foreignKey:AppID"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Backend represents a backend server for an app
type Backend struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AppID        uuid.UUID `json:"app_id" gorm:"type:uuid;not null"`
	App          *App      `json:"app,omitempty" gorm:"foreignKey:AppID"`
	Provider     string    `json:"provider" gorm:"not null"`
	Region       string    `json:"region" gorm:"not null"`
	URL          string    `json:"url" gorm:"not null"`
	SharedSecret string    `json:"shared_secret" gorm:"column:shared_secret;not null"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	Priority     *int      `json:"priority,omitempty"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID           uuid.UUID `json:"id"`
	BackendID    uuid.UUID `json:"backend_id"`
	Timestamp    time.Time `json:"timestamp"`
	OK           bool      `json:"ok"`
	StatusCode   int       `json:"status_code"`
	LatencyMs    int       `json:"latency_ms"`
	ErrorMessage string    `json:"error_message"`
}

// CircuitState represents circuit breaker state
type CircuitState struct {
	BackendID     uuid.UUID  `json:"backend_id"`
	State         string     `json:"state"`
	SinceTs       time.Time  `json:"since_ts"`
	FailCount     int        `json:"fail_count"`
	SuccessCount  int        `json:"success_count"`
	LastFailureTs *time.Time `json:"last_failure_ts"`
	LastSuccessTs *time.Time `json:"last_success_ts"`
}

// BackendStatus represents the combined status of a backend
type BackendStatus struct {
	Backend           *Backend      `json:"backend"`
	CircuitState      *CircuitState `json:"circuit_state"`
	LatestHealthCheck *HealthCheck  `json:"latest_health_check"`
}

// RoutingEvent represents a routing decision and its outcome
type RoutingEvent struct {
	ID        uuid.UUID `json:"id"`
	AppID     uuid.UUID `json:"app_id"`
	BackendID uuid.UUID `json:"backend_id"`
	Timestamp time.Time `json:"timestamp"`
	LatencyMs int       `json:"latency_ms"`
	Outcome   string    `json:"outcome"`
	RequestID string    `json:"request_id"`
}

// Deployment represents a deployment of an app
type Deployment struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AppID        uuid.UUID `json:"app_id" gorm:"type:uuid;not null"`
	App          *App      `json:"app,omitempty" gorm:"foreignKey:AppID"`
	Provider     string    `json:"provider" gorm:"not null"`
	Region       string    `json:"region" gorm:"not null"`
	DeploymentID string    `json:"deployment_id" gorm:"not null"`
	Status       string    `json:"status" gorm:"not null;default:'pending'"`
	ArtifactKey  string    `json:"artifact_key" gorm:"not null"`
	Routes       []string  `json:"routes" gorm:"type:jsonb"`
	Message      string    `json:"message"`
	Metadata     string    `json:"metadata" gorm:"type:json"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DeploymentArtifact represents a stored deployment artifact
type DeploymentArtifact struct {
	Key         string    `json:"key"`
	AppID       uuid.UUID `json:"app_id"`
	Provider    string    `json:"provider"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
}

// LocalRuntimeInstance represents a registered local runtime instance
type LocalRuntimeInstance struct {
	ID            uuid.UUID `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeID     string    `json:"runtime_id" gorm:"column:runtime_id;uniqueIndex;not null"`
	RuntimeType   string    `json:"runtime_type" gorm:"column:runtime_type;not null"`
	FunctionName  string    `json:"function_name" gorm:"column:function_name;not null"`
	ManifestPath  string    `json:"manifest_path" gorm:"column:manifest_path;not null"`
	Host          string    `json:"host" gorm:"column:host;not null"`
	Port          int       `json:"port" gorm:"column:port;not null"`
	PID           int       `json:"pid" gorm:"column:pid;not null"`
	Status        string    `json:"status" gorm:"column:status;not null"`
	LastHeartbeat time.Time `json:"last_heartbeat" gorm:"column:last_heartbeat;not null"`
	Uptime        int64     `json:"uptime" gorm:"column:uptime;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// LocalRuntimeMetric represents a metric snapshot from a local runtime instance
type LocalRuntimeMetric struct {
	ID                uuid.UUID             `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeInstanceID uuid.UUID             `json:"runtime_instance_id" gorm:"column:runtime_instance_id;type:uuid;not null"`
	RuntimeInstance   *LocalRuntimeInstance `json:"runtime_instance,omitempty" gorm:"foreignKey:RuntimeInstanceID"`
	Timestamp         time.Time             `json:"timestamp" gorm:"column:timestamp;not null"`

	MemoryUsage       MemoryStats `json:"memory_usage" gorm:"column:memory_usage;type:jsonb"`
	CPUUsage          float64     `json:"cpu_usage" gorm:"column:cpu_usage;not null"`
	ActiveConnections int         `json:"active_connections" gorm:"column:active_connections;not null"`
	RequestThroughput float64     `json:"request_throughput" gorm:"column:request_throughput;not null"`
	TotalRequests     int64       `json:"total_requests" gorm:"column:total_requests;not null"`
	ErrorRate         float64     `json:"error_rate" gorm:"column:error_rate;not null"`

	ExecutionCount int64         `json:"execution_count" gorm:"column:execution_count;not null"`
	AverageLatency time.Duration `json:"average_latency" gorm:"column:average_latency;not null;type:bigint"`
	ErrorCount     int64         `json:"error_count" gorm:"column:error_count;not null"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// LocalRuntimeHealth represents the health status of a local runtime instance
type LocalRuntimeHealth struct {
	RuntimeInstanceID uuid.UUID             `json:"runtime_instance_id" gorm:"column:runtime_instance_id;type:uuid;primaryKey"`
	RuntimeInstance   *LocalRuntimeInstance `json:"runtime_instance,omitempty" gorm:"foreignKey:RuntimeInstanceID"`
	Timestamp         time.Time             `json:"timestamp" gorm:"column:timestamp;not null"`
	Status            string                `json:"status" gorm:"column:status;not null"`
	ResponseTime      time.Duration         `json:"response_time" gorm:"column:response_time;not null;type:bigint"`
	Checks            JSONMap               `json:"checks" gorm:"column:checks;type:jsonb"`
	Error             *string               `json:"error,omitempty" gorm:"column:error"`
}

func (LocalRuntimeHealth) TableName() string {
	return "local_runtime_health"
}

// EnvironmentVariable represents an environment variable for a function
type EnvironmentVariable struct {
	Key      string `json:"key" db:"key"`
	Value    string `json:"value" db:"value"`
	IsSecret bool   `json:"is_secret" db:"is_secret"`
}

// ScheduleConfig represents a function schedule configuration
type ScheduleConfig struct {
	ID           int64     `json:"id" db:"-"`
	Cron         string    `json:"cron" db:"cron"`
	Timezone     string    `json:"timezone" db:"timezone"`
	Enabled      bool      `json:"enabled" db:"enabled"`
	LastRun      time.Time `json:"last_run" db:"last_run"`
	NextRun      time.Time `json:"next_run" db:"next_run"`
	RunOnDeploy  bool      `json:"run_on_deploy" db:"run_on_deploy"`
}

// FunctionConfig represents a user-created function configuration
type FunctionConfig struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	TenantID          uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	AppID             *uuid.UUID             `json:"app_id,omitempty" db:"app_id"`
	Name              string                 `json:"name" db:"name"`
	Providers         []string               `json:"providers" db:"providers"`
	Region            string                 `json:"region" db:"region"`
	Code              string                 `json:"code" db:"code"`
	EnvVars           []EnvironmentVariable  `json:"env_vars" db:"env_vars"`
	Version           string                 `json:"version" db:"version"`
	Status            string                 `json:"status" db:"status"`
	Schedule          *ScheduleConfig        `json:"schedule,omitempty" db:"schedule"`
	PlaygroundEnabled bool                   `json:"playground_enabled" db:"playground_enabled"`
	PlaygroundConfig  map[string]interface{} `json:"playground_config" db:"playground_config"`
	Capabilities      []string               `json:"capabilities" db:"capabilities"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// FunctionDeployment represents a deployment of a function
type FunctionDeployment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	FunctionID   uuid.UUID `json:"function_id" db:"function_id"`
	Version      string    `json:"version" db:"version"`
	Status       string    `json:"status" db:"status"`
	Provider     string    `json:"provider" db:"provider"`
	Region       string    `json:"region" db:"region"`
	DeployedURL  *string   `json:"deployed_url,omitempty" db:"deployed_url"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	DurationMs   *int      `json:"duration_ms,omitempty" db:"duration_ms"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// FunctionLog represents a log entry for function operations
type FunctionLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	FunctionID   *uuid.UUID             `json:"function_id,omitempty" db:"function_id"`
	DeploymentID *uuid.UUID             `json:"deployment_id,omitempty" db:"deployment_id"`
	Level        string                 `json:"level" db:"level"`
	Message      string                 `json:"message" db:"message"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	Source       string                 `json:"source" db:"source"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// ChangelogEntry represents a changelog entry
type ChangelogEntry struct {
	ID          uuid.UUID         `json:"id"`
	Version     string            `json:"version"`
	Date        time.Time         `json:"date"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Changes     []ChangelogChange `json:"changes"`
	ReleaseURL  *string           `json:"release_url,omitempty"`
	GitHubID    *string           `json:"github_id,omitempty"`
	IsPublished bool              `json:"is_published"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ChangelogChange represents a category of changes in a changelog entry
type ChangelogChange struct {
	ID        uuid.UUID `json:"id"`
	EntryID   uuid.UUID `json:"entry_id"`
	Category  string    `json:"category"`
	Icon      string    `json:"icon"`
	Items     []string  `json:"items"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BlogPost represents a blog post
type BlogPost struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Content       string     `json:"content"`
	Excerpt       string     `json:"excerpt"`
	Author        string     `json:"author"`
	Tags          []string   `json:"tags"`
	FeaturedImage *string    `json:"featured_image,omitempty"`
	SanityID      *string    `json:"sanity_id,omitempty"`
	IsPublished   bool       `json:"is_published"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// BlogCategory represents a blog category
type BlogCategory struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	Order       int       `json:"order"`
	PostCount   int       `json:"postCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// BlogAuthor represents a blog author
type BlogAuthor struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Bio         string                 `json:"bio"`
	Photo       map[string]interface{} `json:"photo,omitempty"`
	Email       string                 `json:"email"`
	Website     string                 `json:"website"`
	SocialLinks map[string]interface{} `json:"socialLinks,omitempty"`
	Role        string                 `json:"role"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// BlogSettings represents platform-wide blog configuration
type BlogSettings struct {
	ID              uuid.UUID `json:"id"`
	BlogTitle       string    `json:"blog_title"`
	PostsPerPage    int       `json:"posts_per_page"`
	MetaDescription string    `json:"meta_description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BlogPageView represents a single page view on a blog post
type BlogPageView struct {
	ID          uuid.UUID  `json:"id"`
	PostID      uuid.UUID  `json:"post_id"`
	VisitorID   string     `json:"visitor_id,omitempty"`
	Referrer    string     `json:"referrer,omitempty"`
	UserAgent   string     `json:"user_agent,omitempty"`
	IPAddress   string     `json:"ip_address,omitempty"`
	Country     string     `json:"country,omitempty"`
	City        string     `json:"city,omitempty"`
	DeviceType  string     `json:"device_type,omitempty"`
	Browser     string     `json:"browser,omitempty"`
	OS          string     `json:"os,omitempty"`
	ViewedAt    time.Time  `json:"viewed_at"`
}

// BlogAnalyticsSummary represents the overall analytics summary
type BlogAnalyticsSummary struct {
	TotalViews      int64   `json:"total_views"`
	TotalPosts      int     `json:"total_posts"`
	PublishedPosts  int     `json:"published_posts"`
	TopPostID       *string `json:"top_post_id,omitempty"`
	TopPostTitle    string  `json:"top_post_title,omitempty"`
	TopPostViews    int64   `json:"top_post_views"`
}

// BlogViewsTimeSeries represents views over time
type BlogViewsTimeSeries struct {
	Date           time.Time `json:"date"`
	Views          int       `json:"views"`
	UniqueVisitors int       `json:"unique_visitors"`
}

// TopBlogPost represents a top performing blog post
type TopBlogPost struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	Author        string    `json:"author"`
	PublishedAt   *string   `json:"published_at,omitempty"`
	TotalViews    int64     `json:"total_views"`
	UniqueViews   int64     `json:"unique_visitors"`
	LastViewedAt  *string   `json:"last_viewed_at,omitempty"`
}

// Feedback represents a user feedback submission
type Feedback struct {
	ID           uuid.UUID            `json:"id"`
	UserID       *uuid.UUID           `json:"user_id,omitempty"`
	UserEmail    *string              `json:"user_email,omitempty"`
	FeedbackType string               `json:"feedback_type"`
	Subject      string               `json:"subject"`
	Message      string               `json:"message"`
	Priority     string               `json:"priority"`
	BrowserInfo  string               `json:"browser_info,omitempty"`
	Status       string               `json:"status"`
	IPAddress    string               `json:"ip_address,omitempty"`
	UserAgent    string               `json:"user_agent,omitempty"`
	Attachments  []FeedbackAttachment `json:"attachments,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// FeedbackAttachment represents a file attachment for feedback
type FeedbackAttachment struct {
	ID          uuid.UUID `json:"id"`
	FeedbackID  uuid.UUID `json:"feedback_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	S3Key       string    `json:"s3_key"`
	S3Bucket    string    `json:"s3_bucket"`
	CreatedAt   time.Time `json:"created_at"`
}

// WaitlistEntry represents a user who has joined the waitlist
type WaitlistEntry struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	Name         string     `json:"name,omitempty" gorm:"size:255"`
	Company      string     `json:"company,omitempty" gorm:"size:255"`
	UseCase      string     `json:"use_case,omitempty" gorm:"column:use_case;type:text"`
	Source       string     `json:"source,omitempty" gorm:"size:100"`
	Status       string     `json:"status" gorm:"size:50;default:'pending'"`
	InviteCodeID *uuid.UUID `json:"invite_code_id,omitempty" gorm:"column:invite_code_id;type:uuid"`
	InvitedAt    *time.Time `json:"invited_at,omitempty" gorm:"column:invited_at"`
	Notes        string     `json:"notes,omitempty" gorm:"type:text"`
	IPAddress    string     `json:"ip_address,omitempty" gorm:"column:ip;size:45"`
	UserAgent    string     `json:"user_agent,omitempty" gorm:"column:user_agent;size:500"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (WaitlistEntry) TableName() string {
	return "waitlist"
}

// WaitlistEntryAdminList is a stripped-down version for safe API responses to the admin dashboard
type WaitlistEntryAdminList struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Company      string     `json:"company"`
	UseCase      string     `json:"use_case"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	InviteCodeID *uuid.UUID `json:"invite_code_id,omitempty"`
	InvitedAt    *time.Time `json:"invited_at,omitempty"`
	Notes        string     `json:"notes"`
	IPAddress    string     `json:"ip_address,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WaitlistStats contains aggregate statistics about the waitlist
type WaitlistStats struct {
	Total       int64 `json:"total"`
	Pending     int64 `json:"pending"`
	Approved    int64 `json:"approved"`
	Invited     int64 `json:"invited"`
	Rejected    int64 `json:"rejected"`
	NewToday    int64 `json:"new_today"`
	NewThisWeek int64 `json:"new_this_week"`
}

// EmailWorkflowConfig represents an email workflow configuration for a tenant
type EmailWorkflowConfig struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	BundleSlug  string    `json:"bundle_slug" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Trigger     string    `json:"trigger" gorm:"not null"`
	Category    string    `json:"category" gorm:"not null"`
	DelayDays   int       `json:"delay_days" gorm:"default:0"`
	Active      bool      `json:"active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EmailWorkflowConfig) TableName() string {
	return "email_workflow_configs"
}

// EmailWorkflowExecution represents a single execution of an email workflow
type EmailWorkflowExecution struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	WorkflowID    uuid.UUID  `json:"workflow_id" gorm:"type:uuid;not null;index"`
	Recipient     string     `json:"recipient" gorm:"not null"`
	Status        string     `json:"status" gorm:"not null;default:'pending'"`
	ScheduledAt   time.Time  `json:"scheduled_at"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	Error         string     `json:"error,omitempty"`
	RetryCount    int        `json:"retry_count" gorm:"default:0"`
	LastRetryAt   *time.Time `json:"last_retry_at,omitempty"`
	EmailSubject  string     `json:"email_subject"`
	EmailTemplate string     `json:"email_template"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EmailWorkflowExecution) TableName() string {
	return "email_workflow_executions"
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	ID           uuid.UUID              `json:"id"`
	MetricType   string                 `json:"metric_type"`
	TenantID     *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID        *uuid.UUID             `json:"app_id,omitempty"`
	BackendID   *uuid.UUID             `json:"backend_id,omitempty"`
	Value        float64                `json:"value"`
	StringValue  string                 `json:"string_value,omitempty"`
	Unit         string                 `json:"unit"`
	Labels       map[string]interface{} `json:"labels,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Alert represents a monitoring alert or incident
type Alert struct {
	ID         uuid.UUID              `json:"id"`
	AlertType  string                 `json:"alert_type"`
	Severity   string                 `json:"severity"`
	TenantID   *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID      *uuid.UUID             `json:"app_id,omitempty"`
	BackendID  *uuid.UUID             `json:"backend_id,omitempty"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message,omitempty"`
	Status     string                 `json:"status"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy *uuid.UUID             `json:"resolved_by,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SystemHealthCheck represents a system health check result
type SystemHealthCheck struct {
	ID             uuid.UUID              `json:"id"`
	CheckType      string                 `json:"check_type"`
	ComponentName  string                 `json:"component_name"`
	Status         string                 `json:"status"`
	ResponseTimeMs *int                   `json:"response_time_ms,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CheckedAt      time.Time              `json:"checked_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

// DatabaseMetric represents a historical database performance metric
type DatabaseMetric struct {
	ID         uuid.UUID              `json:"id"`
	MetricType string                 `json:"metric_type"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	RecordedAt time.Time              `json:"recorded_at"`
	CreatedAt  time.Time              `json:"created_at"`
}

// MonitoringEvent represents a real-time monitoring event
type MonitoringEvent struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"`
	TenantID  *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID     *uuid.UUID             `json:"app_id,omitempty"`
	BackendID *uuid.UUID             `json:"backend_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`
}

// DashboardConfig represents a dashboard configuration
type DashboardConfig struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	ConfigType string                 `json:"config_type"`
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"`
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Incident represents a system incident or operational issue
type Incident struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Severity    string     `json:"severity" db:"severity"`
	Status      string     `json:"status" db:"status"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// FeatureMeasure represents a platform security/feature measure
type FeatureMeasure struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	Icon        string    `json:"icon" db:"icon"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SecurityScan represents a security scan stored in the database
type SecurityScan struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty" db:"tenant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	ScanType    string                 `json:"scan_type" db:"scan_type"`
	Status      string                 `json:"status" db:"status"`
	Target      string                 `json:"target" db:"target"`
	Config      map[string]interface{} `json:"config,omitempty" db:"config"`
	Summary     map[string]interface{} `json:"summary,omitempty" db:"summary"`
	StartedAt   time.Time              `json:"started_at" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	DurationMs  *int                   `json:"duration_ms,omitempty" db:"duration_ms"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Vulnerability represents a security vulnerability stored in the database
type Vulnerability struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	ScanID        uuid.UUID              `json:"scan_id" db:"scan_id"`
	Title         string                 `json:"title" db:"title"`
	Description   string                 `json:"description" db:"description"`
	Severity      string                 `json:"severity" db:"severity"`
	CVSS          *float64               `json:"cvss_score,omitempty" db:"cvss_score"`
	CVE           *string                `json:"cve,omitempty" db:"cve"`
	Category      string                 `json:"category" db:"category"`
	Component     string                 `json:"component" db:"component"`
	Location      *string                `json:"location,omitempty" db:"location"`
	Status        string                 `json:"status" db:"status"`
	Remediation   *string                `json:"remediation,omitempty" db:"remediation"`
	ReferenceUrls []string               `json:"reference_urls,omitempty" db:"reference_urls"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	DiscoveredAt  time.Time              `json:"discovered_at" db:"discovered_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// UserFollow represents a user-to-user follow relationship
type UserFollow struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FollowerID             uuid.UUID `json:"follower_id" gorm:"type:uuid;not null;index"`
	FollowedUserID         uuid.UUID `json:"followed_user_id" gorm:"type:uuid;not null;index"`
	FollowReason           *string   `json:"follow_reason,omitempty" gorm:"size:255"`
	NotifyOnNewFunction    bool      `json:"notify_on_new_function" gorm:"default:true"`
	NotifyOnFunctionUpdate bool      `json:"notify_on_function_update" gorm:"default:true"`
	NotifyOnNewVersion     bool      `json:"notify_on_new_version" gorm:"default:true"`
	CreatedAt              time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Follower     *User `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	FollowedUser *User `json:"followed_user,omitempty" gorm:"foreignKey:FollowedUserID"`
}

// FunctionFollow represents a user-to-function follow relationship
type FunctionFollow struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	FunctionID           uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	FollowReason         *string   `json:"follow_reason,omitempty" gorm:"size:255"`
	NotifyOnNewVersion   bool      `json:"notify_on_new_version" gorm:"default:true"`
	NotifyOnRatingChange bool      `json:"notify_on_rating_change" gorm:"default:false"`
	NotifyOnTrustChange  bool      `json:"notify_on_trust_change" gorm:"default:true"`
	NotifyOnVerification bool      `json:"notify_on_verification" gorm:"default:true"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Function *Function `json:"function,omitempty" gorm:"foreignKey:FunctionID"`
}

// FunctionFavorite represents a user's favorite function
type FunctionFavorite struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	FunctionID uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index"`
	Position   int       `json:"position" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (FunctionFavorite) TableName() string {
	return "function_favorites"
}

// TeamMemory represents a shared team memory
type TeamMemory struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	TeamID   uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	Team     *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID"`

	MemoryType string  `json:"memory_type" gorm:"not null;size:50;index"`
	Category   *string `json:"category,omitempty" gorm:"size:100;index"`

	Content   JSONMap   `json:"content,omitempty" gorm:"type:jsonb"`
	Summary   *string   `json:"summary,omitempty"`
	CreatedBy uuid.UUID `json:"created_by" gorm:"type:uuid;not null"`
	Creator   *User     `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`

	SourceConversationID *uuid.UUID `json:"source_conversation_id,omitempty" gorm:"type:uuid"`
	SourceEventID        *uuid.UUID `json:"source_event_id,omitempty" gorm:"type:uuid"`

	Embedding []float32 `json:"-" gorm:"type:vector(1536)"`

	ConfidenceScore float64    `json:"confidence_score" gorm:"default:0.9"`
	IsValidated     bool       `json:"is_validated" gorm:"default:false;index"`
	ValidatedBy     *uuid.UUID `json:"validated_by,omitempty" gorm:"type:uuid"`
	Validator       *User      `json:"validator,omitempty" gorm:"foreignKey:ValidatedBy"`
	ValidatedAt     *time.Time `json:"validated_at,omitempty"`

	ImportanceScore float64    `json:"importance_score" gorm:"default:0.5"`
	TTLDays         int        `json:"ttl_days" gorm:"default:0"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" gorm:"index"`

	AccessCount    int        `json:"access_count" gorm:"default:0"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`

	AutoUpdateEnabled    bool       `json:"auto_update_enabled" gorm:"default:true;index"`
	LastAutoUpdatedAt    *time.Time `json:"last_auto_updated_at,omitempty"`
	ExtractionConfidence *float64   `json:"extraction_confidence,omitempty"`

	IsEncrypted      bool   `json:"is_encrypted" gorm:"default:false;index"`
	EncryptedContent []byte `json:"-" gorm:"type:bytea"`
	EncryptionIV     []byte `json:"-" gorm:"type:bytea"`
	EncryptionTag    []byte `json:"-" gorm:"type:bytea"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (TeamMemory) TableName() string {
	return "team_memories"
}

func (tm *TeamMemory) IsActive() bool {
	if tm.ExpiresAt == nil {
		return true
	}
	return tm.ExpiresAt.After(time.Now())
}

func (tm *TeamMemory) GetContent() JSONMap {
	if tm.IsEncrypted {
		return nil
	}
	return tm.Content
}

func (tm *TeamMemory) GetSearchableText() string {
	text := ""
	if tm.Summary != nil {
		text += *tm.Summary + " "
	}
	if !tm.IsEncrypted && tm.Content != nil {
		text += fmt.Sprintf("%v", tm.Content)
	}
	return text
}

// TeamMemoryFilter provides filtering options for memory queries
type TeamMemoryFilter struct {
	MemoryType    *string
	Category      *string
	IsValidated   *bool
	MinConfidence *float64
	IsEncrypted   *bool
	CreatedAfter  *time.Time
	Limit         int
	Offset        int
}

// TeamMemorySearchResult represents a search result with relevance score
type TeamMemorySearchResult struct {
	TeamMemory
	RelevanceScore float64 `json:"relevance_score"`
}

// MemoryExtraction represents an AI-extracted memory pending validation
type MemoryExtraction struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID         uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	ConversationID uuid.UUID `json:"conversation_id" gorm:"type:uuid;not null;index"`

	MemoryType string  `json:"memory_type" gorm:"not null;size:50"`
	Category   *string `json:"category,omitempty" gorm:"size:100"`
	Content    JSONMap `json:"content" gorm:"type:jsonb;not null"`
	Summary    string  `json:"summary" gorm:"not null"`
	Confidence float64 `json:"confidence" gorm:"not null"`
	Rationale  string  `json:"rationale"`

	Status          string     `json:"status" gorm:"default:'pending';size:20"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" gorm:"type:uuid"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`

	AutoApplyThreshold float64 `json:"auto_apply_threshold" gorm:"default:0.9"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (MemoryExtraction) TableName() string {
	return "memory_extractions"
}

func (me *MemoryExtraction) ShouldAutoApply() bool {
	return me.Confidence >= me.AutoApplyThreshold && me.Status == "pending"
}

func (me *MemoryExtraction) ToTeamMemory(createdBy uuid.UUID) *TeamMemory {
	now := time.Now()
	return &TeamMemory{
		TeamID:               me.TeamID,
		MemoryType:           me.MemoryType,
		Category:             me.Category,
		Content:              me.Content,
		Summary:              &me.Summary,
		CreatedBy:            createdBy,
		SourceConversationID: &me.ConversationID,
		ConfidenceScore:      me.Confidence,
		IsValidated:          true,
		ValidatedAt:          &now,
		AutoUpdateEnabled:    true,
		ExtractionConfidence: &me.Confidence,
	}
}

// MemoryShare represents a shared memory between teams
type MemoryShare struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MemoryID       uuid.UUID  `json:"memory_id" gorm:"type:uuid;not null;index"`
	SourceTeamID   uuid.UUID  `json:"source_team_id" gorm:"type:uuid;not null;index"`
	TargetTeamID   uuid.UUID  `json:"target_team_id" gorm:"type:uuid;not null;index"`
	SharedBy       uuid.UUID  `json:"shared_by" gorm:"type:uuid;not null"`
	TargetMemoryID *uuid.UUID `json:"target_memory_id,omitempty" gorm:"type:uuid;index"`
	ShareType      string     `json:"share_type" gorm:"size:20;not null;default:'reference'"`
	Permission     string     `json:"permission" gorm:"size:20;not null;default:'read'"`
	Status         string     `json:"status" gorm:"size:20;not null;default:'pending'"`
	Message        *string    `json:"message,omitempty" gorm:"type:text"`
	AcceptedBy     *uuid.UUID `json:"accepted_by,omitempty" gorm:"type:uuid"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty" gorm:"type:timestamptz"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" gorm:"type:timestamptz"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (MemoryShare) TableName() string {
	return "memory_shares"
}

// UsageAlert represents a usage alert configuration for a tenant
type UsageAlert struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Name              string     `json:"name"`
	AlertType         string     `json:"alert_type"`
	ThresholdValue    float64    `json:"threshold_value"`
	ThresholdOperator string     `json:"threshold_operator"`
	PeriodType        string     `json:"period_type"`
	NotificationChannels []string `json:"notification_channels"`
	IsEnabled         bool       `json:"is_enabled"`
	LastTriggeredAt   *time.Time `json:"last_triggered_at,omitempty"`
	TriggerCount      int        `json:"trigger_count"`
	CooldownMinutes   int        `json:"cooldown_minutes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UsageAlertHistory tracks when alerts were triggered
type UsageAlertHistory struct {
	ID             uuid.UUID              `json:"id"`
	AlertID        uuid.UUID              `json:"alert_id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	TriggeredAt    time.Time              `json:"triggered_at"`
	TriggeredValue float64               `json:"triggered_value"`
	ThresholdValue float64               `json:"threshold_value"`
	Message        string                 `json:"message"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	AcknowledgedAt *time.Time            `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uuid.UUID            `json:"acknowledged_by,omitempty"`
}

// SpendCap represents a spending limit for a tenant
type SpendCap struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	CapAmountCents   int        `json:"cap_amount_cents"`
	WarningThresholds []int     `json:"warning_thresholds"`
	CurrentSpendCents int       `json:"current_spend_cents"`
	PeriodStart      time.Time  `json:"period_start"`
	PeriodEnd        time.Time  `json:"period_end"`
	ActionOnCap      string     `json:"action_on_cap"`
	IsHardCap        bool       `json:"is_hard_cap"`
	IsEnabled        bool       `json:"is_enabled"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UsageForecast represents a predicted usage forecast
type UsageForecast struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	ForecastType    string                 `json:"forecast_type"`
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	CurrentValue    float64                `json:"current_value"`
	PredictedValue  float64                `json:"predicted_value"`
	LowerBound      float64                `json:"lower_bound"`
	UpperBound      float64                `json:"upper_bound"`
	Confidence      float64                `json:"confidence"`
	MethodUsed      string                 `json:"method_used"`
	GrowthRate      float64                `json:"growth_rate"`
	DaysOfHistory   int                    `json:"days_of_history"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// DailyUsagePoint represents a single day's usage for time series analysis
type DailyUsagePoint struct {
	Date     time.Time `json:"date"`
	Value    float64   `json:"value"`
	IsAnomaly bool     `json:"is_anomaly"`
}

// UsageExportConfiguration represents a saved export configuration
type UsageExportConfiguration struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description,omitempty"`
	Format             UsageExportFormat      `json:"format"`
	DataTypes          []string               `json:"data_types"`
	Granularity        string                 `json:"granularity"`
	IncludeMetadata    bool                   `json:"include_metadata"`
	IncludeBreakdown   bool                   `json:"include_breakdown"`
	DateRangeType      string                 `json:"date_range_type"`
	FunctionFilter     []uuid.UUID            `json:"function_filter,omitempty"`
	RegionFilter       []string               `json:"region_filter,omitempty"`
	OutcomeFilter      []string               `json:"outcome_filter,omitempty"`
	IsScheduled        bool                   `json:"is_scheduled"`
	ScheduleFrequency  string                 `json:"schedule_frequency,omitempty"`
	ScheduleDayOfMonth *int                   `json:"schedule_day_of_month,omitempty"`
	ScheduleDayOfWeek  *int                   `json:"schedule_day_of_week,omitempty"`
	ScheduleHour       *int                   `json:"schedule_hour,omitempty"`
	DeliveryMethod     string                 `json:"delivery_method"`
	EmailRecipients    []string               `json:"email_recipients,omitempty"`
	WebhookURL         string                 `json:"webhook_url,omitempty"`
	S3Bucket           string                 `json:"s3_bucket,omitempty"`
	S3Prefix           string                 `json:"s3_prefix,omitempty"`
	ExternalSystemID   *uuid.UUID             `json:"external_system_id,omitempty"`
	FieldMapping       map[string]string      `json:"field_mapping,omitempty"`
	TransformConfig    map[string]interface{} `json:"transform_config,omitempty"`
	IsActive           bool                   `json:"is_active"`
	CreatedBy          uuid.UUID              `json:"created_by"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	LastExecutedAt     *time.Time             `json:"last_executed_at,omitempty"`
	LastExportID       *uuid.UUID             `json:"last_export_id,omitempty"`
}

// UsageExportJob represents an instance of an export job
type UsageExportJob struct {
	ID              uuid.UUID         `json:"id"`
	ConfigurationID uuid.UUID         `json:"configuration_id"`
	TenantID        uuid.UUID         `json:"tenant_id"`
	Status          UsageExportStatus `json:"status"`
	Format          UsageExportFormat `json:"format"`
	DataTypes       []string          `json:"data_types"`
	PeriodStart     time.Time         `json:"period_start"`
	PeriodEnd       time.Time         `json:"period_end"`
	RecordCount     int64             `json:"record_count"`
	FileSizeBytes   int64             `json:"file_size_bytes"`
	StorageProvider string            `json:"storage_provider"`
	StoragePath     string            `json:"storage_path"`
	StorageURL      string            `json:"storage_url,omitempty"`
	Checksum        string            `json:"checksum,omitempty"`
	StartedAt       *time.Time        `json:"started_at,omitempty"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	RetryCount      int               `json:"retry_count"`
	DeliveredAt     *time.Time        `json:"delivered_at,omitempty"`
	DeliveryMethod  string            `json:"delivery_method"`
	DeliveryStatus  string            `json:"delivery_status"`
	DeliveryError   string            `json:"delivery_error,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	TriggeredBy     string            `json:"triggered_by"`
}

// UsageExportFormat represents the supported export formats
type UsageExportFormat string

const (
	ExportFormatCSV     UsageExportFormat = "csv"
	ExportFormatJSON    UsageExportFormat = "json"
	ExportFormatParquet UsageExportFormat = "parquet"
	ExportFormatExcel   UsageExportFormat = "excel"
)

// UsageExportStatus represents the status of an export job
type UsageExportStatus string

const (
	ExportStatusPending    UsageExportStatus = "pending"
	ExportStatusProcessing UsageExportStatus = "processing"
	ExportStatusCompleted  UsageExportStatus = "completed"
	ExportStatusFailed     UsageExportStatus = "failed"
	ExportStatusExpired    UsageExportStatus = "expired"
)

// BillingSystemType represents supported external billing system types.
type BillingSystemType string

const (
	BillingSystemQuickBooks BillingSystemType = "quickbooks"
	BillingSystemXero       BillingSystemType = "xero"
)

// ExternalBillingSystem represents a configured external billing integration
type ExternalBillingSystem struct {
	ID                  uuid.UUID              `json:"id"`
	TenantID            uuid.UUID              `json:"tenant_id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	SystemType          string                 `json:"system_type"`
	APIEndpoint         string                 `json:"api_endpoint,omitempty"`
	AuthType            string                 `json:"auth_type"`
	APICredentialKey    string                 `json:"-"`
	APICredentialSecret string                 `json:"-"`
	OAuthToken          string                 `json:"-"`
	OAuthRefreshToken   string                 `json:"-"`
	OAuthExpiresAt      *time.Time             `json:"oauth_expires_at,omitempty"`
	IsActive            bool                   `json:"is_active"`
	LastTestedAt        *time.Time             `json:"last_tested_at,omitempty"`
	LastTestStatus      string                 `json:"last_test_status,omitempty"`
	LastTestError       string                 `json:"last_test_error,omitempty"`
	SyncEnabled         bool                   `json:"sync_enabled"`
	SyncFrequency       string                 `json:"sync_frequency,omitempty"`
	SyncDirection       string                 `json:"sync_direction,omitempty"`
	LastSyncAt          *time.Time             `json:"last_sync_at,omitempty"`
	LastSyncStatus      string                 `json:"last_sync_status,omitempty"`
	FieldMappings       map[string]string      `json:"field_mappings"`
	ValueMappings       map[string]interface{} `json:"value_mappings,omitempty"`
	TransformRules      []TransformRule        `json:"transform_rules,omitempty"`
	WebhookSecret       string                 `json:"-"`
	WebhookURL          string                 `json:"webhook_url,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	CreatedBy           uuid.UUID              `json:"created_by"`
}

// BillingIntegrationSync represents a sync operation with an external billing system
type BillingIntegrationSync struct {
	ID               uuid.UUID         `json:"id"`
	ExternalSystemID uuid.UUID         `json:"external_system_id"`
	TenantID         uuid.UUID         `json:"tenant_id"`
	SyncType         string            `json:"sync_type"`
	Direction        string            `json:"direction"`
	Status           string            `json:"status"`
	StartedAt        *time.Time        `json:"started_at,omitempty"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty"`
	RecordsProcessed int64             `json:"records_processed"`
	RecordsCreated   int64             `json:"records_created"`
	RecordsUpdated   int64             `json:"records_updated"`
	RecordsFailed    int64             `json:"records_failed"`
	RecordsSkipped   int64             `json:"records_skipped"`
	ErrorMessage     string            `json:"error_message,omitempty"`
	ErrorDetails     []SyncErrorDetail `json:"error_details,omitempty"`
	ExternalBatchID  string            `json:"external_batch_id,omitempty"`
	ExternalRefs     map[string]string `json:"external_references,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	TriggeredBy      string            `json:"triggered_by"`
}

// TransformRule represents a data transformation rule for external integrations
type TransformRule struct {
	Field      string               `json:"field"`
	Operation  string               `json:"operation"`
	Value      interface{}          `json:"value,omitempty"`
	Format     string               `json:"format,omitempty"`
	Conditions []TransformCondition `json:"conditions,omitempty"`
}

// TransformCondition represents a conditional transformation
type TransformCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Then     interface{} `json:"then"`
	Else     interface{} `json:"else,omitempty"`
}

// SyncErrorDetail represents an individual error during sync
type SyncErrorDetail struct {
	RecordID     string                 `json:"record_id"`
	RecordType   string                 `json:"record_type"`
	ErrorCode    string                 `json:"error_code"`
	ErrorMessage string                 `json:"error_message"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// UsageExportTemplate represents a predefined export template
type UsageExportTemplate struct {
	ID               uuid.UUID         `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Category         string            `json:"category"`
	Format           UsageExportFormat `json:"format"`
	DataTypes        []string          `json:"data_types"`
	Granularity      string            `json:"granularity"`
	IncludeMetadata  bool              `json:"include_metadata"`
	IncludeBreakdown bool              `json:"include_breakdown"`
	DefaultFields    []string          `json:"default_fields"`
	FieldOrder       []string          `json:"field_order,omitempty"`
	ColumnHeaders    map[string]string `json:"column_headers,omitempty"`
	DataTransforms   []TransformRule   `json:"data_transforms,omitempty"`
	IsActive         bool              `json:"is_active"`
	IsSystem         bool              `json:"is_system"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
