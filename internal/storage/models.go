package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                    uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Tenant                *Tenant    `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Email                 string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash          string     `json:"password_hash" gorm:"column:password_hash"`
	Role                  string     `json:"role,omitempty" gorm:"size:50"` // Platform role for admin users
	EmailVerified         bool       `json:"email_verified" gorm:"default:false"`
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

// PricingTier represents a billing pricing tier
type PricingTier struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PriceCents  int         `json:"price_cents"`
	Currency    string      `json:"currency"`
	Features    interface{} `json:"features"` // JSON features/limits
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Subscription represents a tenant's subscription
type Subscription struct {
	ID                 uuid.UUID    `json:"id"`
	TenantID           uuid.UUID    `json:"tenant_id"`
	PricingTierID      uuid.UUID    `json:"pricing_tier_id"`
	Status             string       `json:"status"`
	CurrentPeriodStart time.Time    `json:"current_period_start"`
	CurrentPeriodEnd   time.Time    `json:"current_period_end"`
	TrialEnd           *time.Time   `json:"trial_end,omitempty"`
	CancelAtPeriodEnd  bool         `json:"cancel_at_period_end"`
	CanceledAt         *time.Time   `json:"canceled_at,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	PricingTier        *PricingTier `json:"pricing_tier,omitempty"` // Populated in queries
}

// Invoice represents a billing invoice
type Invoice struct {
	ID               uuid.UUID  `json:"id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	SubscriptionID   *uuid.UUID `json:"subscription_id,omitempty"`
	Status           string     `json:"status"`
	AmountDueCents   int        `json:"amount_due_cents"`
	AmountPaidCents  int        `json:"amount_paid_cents"`
	Currency         string     `json:"currency"`
	InvoicePdfURL    string     `json:"invoice_pdf_url,omitempty"`
	HostedInvoiceURL string     `json:"hosted_invoice_url,omitempty"`
	PeriodStart      *time.Time `json:"period_start,omitempty"`
	PeriodEnd        *time.Time `json:"period_end,omitempty"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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

// Coupon represents a discount coupon
type Coupon struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	DiscountType   string     `json:"discount_type"` // 'percent' or 'amount'
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
	Coupon         *Coupon    `json:"coupon,omitempty"` // Populated in queries
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
	State         string     `json:"state"` // "closed", "open", "half-open"
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
	Outcome   string    `json:"outcome"` // "success", "failure", "timeout"
	RequestID string    `json:"request_id"`
}

// Deployment represents a deployment of an app
type Deployment struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AppID        uuid.UUID `json:"app_id" gorm:"type:uuid;not null"`
	App          *App      `json:"app,omitempty" gorm:"foreignKey:AppID"`
	Provider     string    `json:"provider" gorm:"not null"`
	Region       string    `json:"region" gorm:"not null"`
	DeploymentID string    `json:"deployment_id" gorm:"not null"`            // Provider-specific deployment ID
	Status       string    `json:"status" gorm:"not null;default:'pending'"` // "pending", "deploying", "success", "failed", "rollback"
	ArtifactKey  string    `json:"artifact_key" gorm:"not null"`             // Reference to stored artifact
	Routes       []string  `json:"routes" gorm:"type:jsonb"`                 // Route patterns bound to this deployment
	Message      string    `json:"message"`                                  // Status message or error details
	Metadata     string    `json:"metadata" gorm:"type:json"`                // JSON metadata from provider
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

// ChangelogEntry represents a changelog entry
type ChangelogEntry struct {
	ID          uuid.UUID         `json:"id"`
	Version     string            `json:"version"`
	Date        time.Time         `json:"date"`
	Type        string            `json:"type"` // "major", "minor", "patch"
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
	Icon      string    `json:"icon"` // Icon name from lucide-react
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

// Feedback represents a user feedback submission
type Feedback struct {
	ID           uuid.UUID            `json:"id"`
	UserID       *uuid.UUID           `json:"user_id,omitempty"`    // Anonymous users can submit feedback
	UserEmail    *string              `json:"user_email,omitempty"` // For anonymous feedback
	FeedbackType string               `json:"feedback_type"`        // "bug", "feature", "improvement", "general"
	Subject      string               `json:"subject"`
	Message      string               `json:"message"`
	Priority     string               `json:"priority"`               // "low", "medium", "high", "critical"
	BrowserInfo  string               `json:"browser_info,omitempty"` // Browser/OS information
	Status       string               `json:"status"`                 // "submitted", "in-review", "resolved", "closed"
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
	S3Key       string    `json:"s3_key"` // S3 object key for the file
	S3Bucket    string    `json:"s3_bucket"`
	CreatedAt   time.Time `json:"created_at"`
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	ID          uuid.UUID              `json:"id"`
	MetricType  string                 `json:"metric_type"` // 'response_time', 'error_rate', 'throughput', 'health_score', 'circuit_state'
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID       *uuid.UUID             `json:"app_id,omitempty"`
	BackendID   *uuid.UUID             `json:"backend_id,omitempty"`
	Value       float64                `json:"value"`
	StringValue string                 `json:"string_value,omitempty"` // For string-based metrics like circuit states
	Unit        string                 `json:"unit"`                   // 'ms', 'percent', 'requests_per_second', 'score', 'state'
	Labels      map[string]interface{} `json:"labels,omitempty"`       // Additional metadata
	Timestamp   time.Time              `json:"timestamp"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Alert represents a monitoring alert or incident
type Alert struct {
	ID         uuid.UUID              `json:"id"`
	AlertType  string                 `json:"alert_type"` // 'health_degraded', 'backend_down', 'high_error_rate', 'circuit_open'
	Severity   string                 `json:"severity"`   // 'info', 'warning', 'error', 'critical'
	TenantID   *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID      *uuid.UUID             `json:"app_id,omitempty"`
	BackendID  *uuid.UUID             `json:"backend_id,omitempty"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message,omitempty"`
	Status     string                 `json:"status"` // 'active', 'acknowledged', 'resolved'
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy *uuid.UUID             `json:"resolved_by,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Additional alert-specific data
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SystemHealthCheck represents a system health check result
type SystemHealthCheck struct {
	ID             uuid.UUID              `json:"id"`
	CheckType      string                 `json:"check_type"` // 'database', 'api', 'external_service', 'disk_space', 'memory'
	ComponentName  string                 `json:"component_name"`
	Status         string                 `json:"status"` // 'healthy', 'degraded', 'unhealthy', 'unknown'
	ResponseTimeMs *int                   `json:"response_time_ms,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CheckedAt      time.Time              `json:"checked_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

// DatabaseMetric represents a historical database performance metric
type DatabaseMetric struct {
	ID         uuid.UUID              `json:"id"`
	MetricType string                 `json:"metric_type"` // 'connections', 'size_gb', 'query_time', 'cache_hit_ratio', 'throughput'
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`               // 'count', 'gb', 'ms', 'ratio', 'qps'
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Additional context
	RecordedAt time.Time              `json:"recorded_at"`
	CreatedAt  time.Time              `json:"created_at"`
}

// MonitoringEvent represents a real-time monitoring event
type MonitoringEvent struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"` // 'request_completed', 'backend_failover', 'circuit_breaker_transition'
	TenantID  *uuid.UUID             `json:"tenant_id,omitempty"`
	AppID     *uuid.UUID             `json:"app_id,omitempty"`
	BackendID *uuid.UUID             `json:"backend_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"` // Event-specific payload
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`
}

// DashboardConfig represents a dashboard configuration
type DashboardConfig struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"` // NULL for tenant-wide configs
	ConfigType string                 `json:"config_type"`       // 'metric_panel', 'alert_rule', 'chart_config'
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"` // Configuration data specific to the type
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SecurityScan represents a security scan stored in the database
type SecurityScan struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty" db:"tenant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	ScanType    string                 `json:"scan_type" db:"scan_type"` // "penetration_test", "vulnerability_scan", "compliance_check"
	Status      string                 `json:"status" db:"status"`       // "running", "completed", "failed"
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
	Severity      string                 `json:"severity" db:"severity"` // "critical", "high", "medium", "low", "info"
	CVSS          *float64               `json:"cvss_score,omitempty" db:"cvss_score"`
	CVE           *string                `json:"cve,omitempty" db:"cve"`
	Category      string                 `json:"category" db:"category"` // "injection", "auth", "crypto", "config", "network"
	Component     string                 `json:"component" db:"component"`
	Location      *string                `json:"location,omitempty" db:"location"`
	Status        string                 `json:"status" db:"status"` // "open", "fixed", "accepted", "false_positive"
	Remediation   *string                `json:"remediation,omitempty" db:"remediation"`
	ReferenceUrls []string               `json:"reference_urls,omitempty" db:"reference_urls"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	DiscoveredAt  time.Time              `json:"discovered_at" db:"discovered_at"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// Session represents a user session with MFA verification status
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	SessionToken string     `json:"session_token" db:"session_token"`           // JWT token or session ID
	MFAVerified  bool       `json:"mfa_verified" db:"mfa_verified"`             // Whether MFA has been verified for this session
	MFALastUsed  *time.Time `json:"mfa_last_used,omitempty" db:"mfa_last_used"` // Last time MFA was verified
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	LastActivity time.Time  `json:"last_activity" db:"last_activity"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// LocalRuntimeInstance represents a registered local runtime instance
type LocalRuntimeInstance struct {
	ID            uuid.UUID `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeID     string    `json:"runtime_id" gorm:"column:runtime_id;uniqueIndex;not null"` // Unique identifier for this runtime instance
	RuntimeType   string    `json:"runtime_type" gorm:"column:runtime_type;not null"`         // "node18", "node20", "python3.11", etc.
	FunctionName  string    `json:"function_name" gorm:"column:function_name;not null"`
	ManifestPath  string    `json:"manifest_path" gorm:"column:manifest_path;not null"`
	Host          string    `json:"host" gorm:"column:host;not null"`     // Hostname/IP
	Port          int       `json:"port" gorm:"column:port;not null"`     // Port number
	PID           int       `json:"pid" gorm:"column:pid;not null"`       // Process ID
	Status        string    `json:"status" gorm:"column:status;not null"` // "running", "stopped", "error"
	LastHeartbeat time.Time `json:"last_heartbeat" gorm:"column:last_heartbeat;not null"`
	Uptime        int64     `json:"uptime" gorm:"column:uptime;not null"` // Uptime in seconds
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// LocalRuntimeMetric represents a metric snapshot from a local runtime instance
type LocalRuntimeMetric struct {
	ID                uuid.UUID             `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	RuntimeInstanceID uuid.UUID             `json:"runtime_instance_id" gorm:"column:runtime_instance_id;type:uuid;not null"`
	RuntimeInstance   *LocalRuntimeInstance `json:"runtime_instance,omitempty" gorm:"foreignKey:RuntimeInstanceID"`
	Timestamp         time.Time             `json:"timestamp" gorm:"column:timestamp;not null"`

	// Performance metrics
	MemoryUsage       MemoryStats `json:"memory_usage" gorm:"column:memory_usage;type:jsonb"`
	CPUUsage          float64     `json:"cpu_usage" gorm:"column:cpu_usage;not null"`
	ActiveConnections int         `json:"active_connections" gorm:"column:active_connections;not null"`
	RequestThroughput float64     `json:"request_throughput" gorm:"column:request_throughput;not null"`
	TotalRequests     int64       `json:"total_requests" gorm:"column:total_requests;not null"`
	ErrorRate         float64     `json:"error_rate" gorm:"column:error_rate;not null"`

	// Function execution metrics
	ExecutionCount int64         `json:"execution_count" gorm:"column:execution_count;not null"`
	AverageLatency time.Duration `json:"average_latency" gorm:"column:average_latency;not null;type:bigint"`
	ErrorCount     int64         `json:"error_count" gorm:"column:error_count;not null"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Heap   uint64 `json:"heap" gorm:"column:heap"`
	Stack  uint64 `json:"stack" gorm:"column:stack"`
	System uint64 `json:"system" gorm:"column:system"`
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

// LocalRuntimeHealth represents the health status of a local runtime instance
type LocalRuntimeHealth struct {
	RuntimeInstanceID uuid.UUID             `json:"runtime_instance_id" gorm:"column:runtime_instance_id;type:uuid;primaryKey"`
	RuntimeInstance   *LocalRuntimeInstance `json:"runtime_instance,omitempty" gorm:"foreignKey:RuntimeInstanceID"`
	Timestamp         time.Time             `json:"timestamp" gorm:"column:timestamp;not null"`
	Status            string                `json:"status" gorm:"column:status;not null"` // "healthy", "degraded", "unhealthy"
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
	Status            string                 `json:"status" db:"status"` // "draft", "deploying", "deployed", "failed"
	PlaygroundEnabled bool                   `json:"playground_enabled" db:"playground_enabled"`
	PlaygroundConfig  map[string]interface{} `json:"playground_config" db:"playground_config"`
	Capabilities      []string               `json:"capabilities" db:"capabilities"` // Declared capabilities for sandbox enforcement
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// FunctionDeployment represents a deployment of a function
type FunctionDeployment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	FunctionID   uuid.UUID `json:"function_id" db:"function_id"`
	Version      string    `json:"version" db:"version"`
	Status       string    `json:"status" db:"status"` // "pending", "deploying", "success", "failed"
	Provider     string    `json:"provider" db:"provider"`
	Region       string    `json:"region" db:"region"`
	DeployedURL  *string   `json:"deployed_url,omitempty" db:"deployed_url"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// FunctionLog represents a log entry for function operations
type FunctionLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	FunctionID   *uuid.UUID             `json:"function_id,omitempty" db:"function_id"`
	DeploymentID *uuid.UUID             `json:"deployment_id,omitempty" db:"deployment_id"`
	Level        string                 `json:"level" db:"level"` // "info", "warn", "error", "debug"
	Message      string                 `json:"message" db:"message"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	Source       string                 `json:"source" db:"source"` // "deployment", "runtime", "monitoring"
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// Incident represents a system incident or operational issue
type Incident struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Severity    string     `json:"severity" db:"severity"` // "critical", "high", "medium", "low"
	Status      string     `json:"status" db:"status"`     // "resolved", "investigating", "monitoring"
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// ============================================
// StateFabric Models - Composable Durable State
// ============================================

// State represents a durable state container bound to a function identity
type State struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Name       string     `json:"name" gorm:"not null;size:255"`
	FullPath   string     `json:"full_path" gorm:"uniqueIndex;not null;size:500"` // "acme/cart"
	FunctionID *uuid.UUID `json:"function_id,omitempty" gorm:"type:uuid;index"`   // Optional bound function

	// State Configuration
	StorageType string `json:"storage_type" gorm:"not null;default:'keyvalue';size:50"` // "keyvalue" | "document" | "timeseries" | "graph"

	// Retention
	TTLDays   int `json:"ttl_days" gorm:"not null;default:0"`     // 0 = forever
	MaxSizeMB int `json:"max_size_mb" gorm:"not null;default:100"`

	// Versioning
	CurrentVersion int  `json:"current_version" gorm:"not null;default:1"`
	IsVersioned    bool `json:"is_versioned" gorm:"not null;default:true"`

	// Permissions
	IsPublic         bool `json:"is_public" gorm:"not null;default:false"`
	AllowCrossTenant bool `json:"allow_cross_tenant" gorm:"not null;default:false"`

	// Metadata
	Description *string `json:"description,omitempty" gorm:"size:1000"`
	Tags        JSONMap `json:"tags" gorm:"type:jsonb;not null;default:'{}'"`

	// Billing
	StorageUsedMB int64 `json:"storage_used_mb" gorm:"not null;default:0"`
	WriteOpsMonth int64 `json:"write_ops_month" gorm:"not null;default:0"`
	ReadOpsMonth  int64 `json:"read_ops_month" gorm:"not null;default:0"`

	// Timestamps
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`
	LastAccessedAt time.Time `json:"last_accessed_at" gorm:"not null"`
}

func (State) TableName() string {
	return "states"
}

// BeforeCreate hook for State
func (s *State) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.StorageType == "" {
		s.StorageType = "keyvalue"
	}
	if s.TTLDays == 0 {
		s.TTLDays = 0 // Explicitly set to 0 for forever
	}
	if s.MaxSizeMB == 0 {
		s.MaxSizeMB = 100
	}
	if s.CurrentVersion == 0 {
		s.CurrentVersion = 1
	}
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	if s.LastAccessedAt.IsZero() {
		s.LastAccessedAt = now
	}
	return nil
}

// BeforeUpdate hook for State
func (s *State) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

// StateValue represents a single key-value entry in state
type StateValue struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Key (supports hierarchical keys like "user/123/profile")
	Key string `json:"key" gorm:"not null;index;size:500"`

	// Value (JSON for flexibility)
	Value JSONMap `json:"value" gorm:"type:jsonb;not null"`

	// Versioning
	Version       int      `json:"version" gorm:"not null;index"`
	PreviousValue *JSONMap `json:"previous_value,omitempty" gorm:"type:jsonb"`

	// Content Addressing (for deduplication)
	ContentHash string `json:"content_hash" gorm:"index;size:64"`

	// TTL
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index"`

	// Metadata
	CreatedBy string    `json:"created_by" gorm:"not null;size:255"` // function_id or user_id
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;not null"`
}

func (StateValue) TableName() string {
	return "state_values"
}

// BeforeCreate hook for StateValue
func (sv *StateValue) BeforeCreate(tx *gorm.DB) error {
	if sv.ID == uuid.Nil {
		sv.ID = uuid.New()
	}
	if sv.CreatedAt.IsZero() {
		sv.CreatedAt = time.Now()
	}
	if sv.ContentHash == "" && sv.Value != nil {
		// Generate content hash for deduplication (simplified)
		// In production, you'd use a proper hashing algorithm
		sv.ContentHash = fmt.Sprintf("%d", len(sv.Value))
	}
	return nil
}

// StateEvent represents an immutable event in state history
type StateEvent struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Event Types: "set" | "delete" | "snapshot" | "restore" | "merge"
	EventType string `json:"event_type" gorm:"not null;index;size:50"`

	// Key affected (null for state-level events)
	Key *string `json:"key,omitempty" gorm:"index;size:500"`

	// Event Data
	PreviousValue *JSONMap `json:"previous_value,omitempty" gorm:"type:jsonb"`
	NewValue      *JSONMap `json:"new_value,omitempty" gorm:"type:jsonb"`

	// Causality
	CausationID   *uuid.UUID `json:"causation_id,omitempty" gorm:"type:uuid;index"`
	CorrelationID string     `json:"correlation_id" gorm:"not null;size:255;index"` // For distributed tracing

	// Source
	SourceType string `json:"source_type" gorm:"not null;size:50;index"` // "function" | "user" | "system" | "trigger"
	SourceID   string `json:"source_id" gorm:"not null;size:255;index"`  // function_id or user_id

	// Determinism Proof (for replay verification)
	InputHash     string `json:"input_hash" gorm:"size:128"`
	OutputHash    string `json:"output_hash" gorm:"size:128"`
	Deterministic bool   `json:"deterministic" gorm:"not null;default:false"`

	// Sequence (for ordering)
	SequenceNum int64 `json:"sequence_num" gorm:"not null;uniqueIndex"`

	Timestamp time.Time `json:"timestamp" gorm:"autoCreateTime;not null;index"`
}

func (StateEvent) TableName() string {
	return "state_events"
}

// BeforeCreate hook for StateEvent
func (se *StateEvent) BeforeCreate(tx *gorm.DB) error {
	if se.ID == uuid.Nil {
		se.ID = uuid.New()
	}
	if se.Timestamp.IsZero() {
		se.Timestamp = time.Now()
	}
	if se.SequenceNum == 0 {
		// Get next sequence number for this state
		var maxSeq int64
		tx.Model(&StateEvent{}).Where("state_id = ?", se.StateID).Select("COALESCE(MAX(sequence_num), 0)").Scan(&maxSeq)
		se.SequenceNum = maxSeq + 1
	}
	return nil
}

// StateSnapshot represents a point-in-time snapshot of state
type StateSnapshot struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Snapshot Identification
	SnapshotVersion int     `json:"snapshot_version" gorm:"not null"`
	Label           *string `json:"label,omitempty"` // Optional human-readable label

	// Content
	StateData      JSONMap `json:"state_data" gorm:"type:jsonb;not null"`
	StateSizeBytes int64   `json:"state_size_bytes"`

	// Coverage
	KeyCount      int   `json:"key_count"`
	FirstSequence int64 `json:"first_sequence"`
	LastSequence  int64 `json:"last_sequence"`

	// Determinism
	RootEventID uuid.UUID `json:"root_event_id" gorm:"type:uuid"` // First event in snapshot

	// Compression
	IsCompressed    bool   `json:"is_compressed" gorm:"default:false"`
	CompressionAlgo string `json:"compression_algo"` // "lz4", "zstd", ""

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (StateSnapshot) TableName() string {
	return "state_snapshots"
}

// StatePermission defines access control for state
type StatePermission struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StateID uuid.UUID `json:"state_id" gorm:"type:uuid;not null;index"`

	// Principal
	PrincipalType string     `json:"principal_type"` // "user" | "team" | "function" | "tenant"
	PrincipalID   *uuid.UUID `json:"principal_id,omitempty" gorm:"type:uuid"`

	// Permissions
	CanRead    bool `json:"can_read" gorm:"default:false"`
	CanWrite   bool `json:"can_write" gorm:"default:false"`
	CanDelete  bool `json:"can_delete" gorm:"default:false"`
	CanAdmin   bool `json:"can_admin" gorm:"default:false"`
	CanTrigger bool `json:"can_trigger" gorm:"default:false"` // For function triggers

	// Constraints
	IPWhitelist      JSONMap `json:"ip_whitelist" gorm:"type:jsonb;default:'[]'"`
	TimeRestrictions JSONMap `json:"time_restrictions" gorm:"type:jsonb"`
	RateLimit        *int    `json:"rate_limit,omitempty"` // Requests per minute

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (StatePermission) TableName() string {
	return "state_permissions"
}

// StateTrigger defines automatic function invocation on state changes
type StateTrigger struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`

	// Source
	SourceStateID *uuid.UUID `json:"source_state_id,omitempty" gorm:"type:uuid"`

	// Trigger Configuration
	TriggerType string  `json:"trigger_type" gorm:"not null"` // "on_write" | "on_read" | "on_delete" | "on_condition"
	KeyPattern  *string `json:"key_pattern,omitempty"`         // Glob pattern for keys

	// Condition (for advanced triggers)
	Condition JSONMap `json:"condition" gorm:"type:jsonb"`

	// Target Function
	TargetFunctionID *uuid.UUID `json:"target_function_id,omitempty" gorm:"type:uuid"`
	TargetFunction   string     `json:"target_function"` // "org/function:version"

	// Payload
	IncludePrevious bool `json:"include_previous" gorm:"default:false"`
	IncludeNew      bool `json:"include_new" gorm:"default:true"`

	// Rate Limiting
	MaxInvocationsPerMinute int `json:"max_invocations_per_minute" gorm:"default:60"`

	// Status
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (StateTrigger) TableName() string {
	return "state_triggers"
}

// AgentMemory represents AI agent memory with embeddings
type AgentMemory struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	AgentID  string    `json:"agent_id" gorm:"not null;index"`

	// Memory Type: "working" | "longterm" | "context" | "episodic"
	MemoryType string `json:"memory_type" gorm:"not null;index"`

	// Content
	Content        *string `json:"content,omitempty"`
	StructuredData JSONMap `json:"structured_data" gorm:"type:jsonb"`

	// Embedding (pgvector) - stored as byte array for PostgreSQL vector type
	Embedding []float32 `json:"embedding" gorm:"type:vector(1536)"`

	// Metadata
	ImportanceScore float32    `json:"importance_score" gorm:"default:0.5"` // 0.0-1.0 for retention
	AccessCount     int        `json:"access_count" gorm:"default:0"`
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty"`

	// Retention
	TTLDays   int        `json:"ttl_days"` // 0 = forever
	ExpiresAt *time.Time `json:"expires_at,omitempty" gorm:"index"`

	// Causality
	SourceEventID *uuid.UUID `json:"source_event_id,omitempty" gorm:"type:uuid"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (AgentMemory) TableName() string {
	return "agent_memories"
}

// AgentMemoryIndex for vector similarity search
type AgentMemoryIndex struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	AgentID  string    `json:"agent_id" gorm:"not null;index"`

	// Index Configuration
	MemoryType       string `json:"memory_type"`
	Dimension        int    `json:"dimension" gorm:"default:1536"`
	SimilarityMetric string `json:"similarity_metric" gorm:"default:'cosine'"`

	// Index Stats
	MemoryCount   int        `json:"memory_count" gorm:"default:0"`
	LastIndexedAt *time.Time `json:"last_indexed_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (AgentMemoryIndex) TableName() string {
	return "agent_memory_indexes"
}

// StateUsageMetric represents usage data for billing and analytics
type StateUsageMetric struct {
	ID       uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	StateID  *uuid.UUID `json:"state_id,omitempty" gorm:"type:uuid;index"`

	// Metric type
	MetricType string `json:"metric_type" gorm:"not null"` // "storage" | "write_ops" | "read_ops"

	// Value
	Value int64  `json:"value" gorm:"not null"`
	Unit  string `json:"unit"` // "bytes" | "ops" | "mb"

	// Time period
	PeriodStart time.Time `json:"period_start" gorm:"not null"`
	PeriodEnd   time.Time `json:"period_end" gorm:"not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (StateUsageMetric) TableName() string {
	return "state_usage_metrics"
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

// TeamInvite represents a team invitation during onboarding
type TeamInvite struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamID     uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index"`
	Email      string    `json:"email" gorm:"not null"`
	Token      string    `json:"token" gorm:"uniqueIndex;not null"`
	Role       string    `json:"role" gorm:"not null"` // "admin", "member", "viewer"
	InvitedBy  uuid.UUID `json:"invited_by" gorm:"type:uuid;not null"`
	Message    string    `json:"message,omitempty"`
	Status     string    `json:"status" gorm:"default:'pending'"` // "pending", "accepted", "expired", "cancelled"
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (TeamInvite) TableName() string {
	return "team_invites"
}
