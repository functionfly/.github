package storage

import (
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

// Type aliases for registry types
type (
	RegistryFunction                      = registry.RegistryFunction
	RegistryFunctionVersion              = registry.RegistryFunctionVersion
	RegistryFunctionExecution            = registry.RegistryFunctionExecution
	RegistryExecutionPublic              = registry.RegistryExecutionPublic
	ExecutionResourceUsage               = registry.ExecutionResourceUsage
	RegistryFunctionRating               = registry.RegistryFunctionRating
	RegistryFunctionSignature            = registry.RegistryFunctionSignature
	RegistryFunctionMalwareScan          = registry.RegistryFunctionMalwareScan
	RegistryFunctionApproval             = registry.RegistryFunctionApproval
	RegistryFunctionApprovalComment      = registry.RegistryFunctionApprovalComment
	RegistryFunctionVerificationStatus   = registry.RegistryFunctionVerificationStatus

	// DRE 2.0 type aliases
	MEGRecord             = registry.MEGRecord
	ExecutionCertificate  = registry.ExecutionCertificate
	DriftReportRecord     = registry.DriftReportRecord
	ExecutionPassport     = registry.ExecutionPassport
	PassportUpdate        = registry.PassportUpdate
	DREScores             = registry.DREScores
)

// ============================================
// Execution Security Models
// ============================================

// UserExecutionQuota represents per-user execution quotas and limits
type UserExecutionQuota struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"` // NULL for anonymous users
	TenantID    *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	IPAddress   string     `json:"ip_address" gorm:"size:45;index"` // IPv4/IPv6 address for anonymous users

	// Quota limits
	DailyExecutionLimit     int       `json:"daily_execution_limit" gorm:"default:1000"`
	HourlyExecutionLimit    int       `json:"hourly_execution_limit" gorm:"default:100"`
	MinuteExecutionLimit    int       `json:"minute_execution_limit" gorm:"default:10"`

	// Current usage counters
	DailyExecutions         int       `json:"daily_executions" gorm:"default:0"`
	HourlyExecutions        int       `json:"hourly_executions" gorm:"default:0"`
	MinuteExecutions        int       `json:"minute_executions" gorm:"default:0"`

	// Reset timestamps
	DailyResetAt            time.Time `json:"daily_reset_at" gorm:"index"`
	HourlyResetAt           time.Time `json:"hourly_reset_at" gorm:"index"`
	MinuteResetAt           time.Time `json:"minute_reset_at" gorm:"index"`

	// Abuse detection
	SuspiciousActivityScore int       `json:"suspicious_activity_score" gorm:"default:0"`
	LastSuspiciousActivity  *time.Time `json:"last_suspicious_activity,omitempty"`
	IsThrottled             bool      `json:"is_throttled" gorm:"default:false"`
	ThrottleUntil           *time.Time `json:"throttle_until,omitempty"`
	BlockUntil              *time.Time `json:"block_until,omitempty"`

	// CAPTCHA requirements
	CaptchaRequired         bool      `json:"captcha_required" gorm:"default:false"`
	LastCaptchaCompleted    *time.Time `json:"last_captcha_completed,omitempty"`

	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// AbusePattern represents detected abuse patterns
type AbusePattern struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PatternType   string    `json:"pattern_type" gorm:"size:50;index"` // "rate_spike", "error_rate", "suspicious_input", "resource_abuse"
	Severity      string    `json:"severity" gorm:"size:20"` // "low", "medium", "high", "critical"

	// Pattern identification
	UserID        *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	IPAddress     string     `json:"ip_address" gorm:"size:45;index"`
	FunctionID    *uuid.UUID `json:"function_id,omitempty" gorm:"type:uuid;index"`

	// Pattern data
	PatternData   json.RawMessage `json:"pattern_data" gorm:"type:jsonb"` // JSON data about the pattern
	Description   string          `json:"description" gorm:"type:text"`

	// Action taken
	ActionTaken   string          `json:"action_taken" gorm:"size:50"` // "none", "throttled", "blocked", "captcha_required"
	ActionData    json.RawMessage `json:"action_data" gorm:"type:jsonb"` // JSON data about the action

	// Timing
	DetectedAt    time.Time       `json:"detected_at" gorm:"autoCreateTime;index"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty"`

	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// ExecutionSecurityEvent represents security events during execution
type ExecutionSecurityEvent struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID   *uuid.UUID `json:"execution_id,omitempty" gorm:"type:uuid;index"` // Links to RegistryFunctionExecution
	UserID        *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	IPAddress     string     `json:"ip_address" gorm:"size:45;index"`

	EventType     string     `json:"event_type" gorm:"size:50;index"` // "quota_exceeded", "captcha_failed", "input_validation_failed", "timeout", "resource_limit_exceeded"
	Severity      string     `json:"severity" gorm:"size:20"` // "info", "warning", "error", "critical"

	Message       string     `json:"message" gorm:"type:text"`
	EventData     json.RawMessage `json:"event_data" gorm:"type:jsonb"` // Additional event data

	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// FunctionInputSchema represents JSON schema for input validation
type FunctionInputSchema struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionVersionID uuid.UUID       `json:"function_version_id" gorm:"type:uuid;not null;uniqueIndex"`
	Schema            json.RawMessage `json:"schema" gorm:"type:jsonb;not null"` // JSON Schema for input validation
	IsStrict          bool            `json:"is_strict" gorm:"default:false"` // Whether to enforce strict schema validation
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

