package storage

import (
	"time"

	"github.com/google/uuid"
)

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

// OAuthState stores OAuth CSRF state tokens for validation on callback (persisted for multi-instance).
type OAuthState struct {
	State     string    `gorm:"column:state;primaryKey;size:512"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
}

// TableName overrides the default table name for GORM.
func (OAuthState) TableName() string {
	return "oauth_states"
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

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"token_hash" db:"token_hash"` // SHA-256 hash of the refresh token
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`     // Whether this token has been revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
