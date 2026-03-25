package trustapi

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PartnerTier represents the partner subscription tier
type PartnerTier string

const (
	PartnerTierDeveloper  PartnerTier = "developer"
	PartnerTierStartup   PartnerTier = "startup"
	PartnerTierBusiness  PartnerTier = "business"
	PartnerTierEnterprise PartnerTier = "enterprise"
)

// PartnerStatus represents the partner account status
type PartnerStatus string

const (
	PartnerStatusPending   PartnerStatus = "pending"
	PartnerStatusActive   PartnerStatus = "active"
	PartnerStatusSuspended PartnerStatus = "suspended"
	PartnerStatusCancelled PartnerStatus = "cancelled"
)

// RateLimitTier defines rate limits per partner tier (requests per minute)
var RateLimitsPerTier = map[PartnerTier]RateLimitConfig{
	PartnerTierDeveloper: {
		PerMinute:       60,
		PerDay:          10000,
		MonthlyRequestLimit: 50000,
	},
	PartnerTierStartup: {
		PerMinute:       300,
		PerDay:          100000,
		MonthlyRequestLimit: 500000,
	},
	PartnerTierBusiness: {
		PerMinute:       1000,
		PerDay:          500000,
		MonthlyRequestLimit: 2000000,
	},
	PartnerTierEnterprise: {
		PerMinute:       10000,
		PerDay:          10000000,
		MonthlyRequestLimit: 100000000,
	},
}

// RateLimitConfig holds rate limit configuration for a tier
type RateLimitConfig struct {
	PerMinute          int
	PerDay             int
	MonthlyRequestLimit int
}

// TrustAPIPartner represents an external platform partner
type TrustAPIPartner struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name     string    `json:"name" gorm:"size:255;not null"`
	Slug     string    `json:"slug" gorm:"size:100;not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`

	// Contact information
	ContactEmail string `json:"contact_email" gorm:"size:255;not null"`
	ContactName  string `json:"contact_name" gorm:"size:255"`
	WebsiteURL   string `json:"website_url" gorm:"size:500"`

	// Partner tier
	Tier string `json:"tier" gorm:"size:50;not null;default:'developer'"`

	// Rate limits
	RateLimitPerMinute int `json:"rate_limit_per_minute" gorm:"default:60"`
	RateLimitPerDay    int `json:"rate_limit_per_day" gorm:"default:10000"`

	// Usage quotas
	MonthlyRequestLimit int `json:"monthly_request_limit" gorm:"default:50000"`
	CurrentMonthUsage   int `json:"current_month_usage" gorm:"default:0"`

	// Billing
	BillingEmail      string `json:"billing_email" gorm:"size:255"`
	BillingAccountID  string `json:"billing_account_id" gorm:"size:255"`

	// Status
	Status string `json:"status" gorm:"size:50;not null;default:'pending'"`

	// SSO
	SSOEnabled   bool   `json:"sso_enabled" gorm:"default:false"`
	SSOProvider  string `json:"sso_provider" gorm:"size:50"`

	// Webhook
	WebhookURL        string `json:"webhook_url" gorm:"size:500"`
	WebhookSecretHash string `json:"webhook_secret_hash" gorm:"size:255"`

	// Metadata
	Metadata json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Timestamps
	ActivatedAt *time.Time `json:"activated_at"`
	SuspendedAt *time.Time `json:"suspended_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for TrustAPIPartner
func (TrustAPIPartner) TableName() string {
	return "trust_api_partners"
}

// TrustAPIKey represents an API key for partner authentication
type TrustAPIKey struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PartnerID uuid.UUID `json:"partner_id" gorm:"type:uuid;not null"`

	// Key identification
	KeyID     string `json:"key_id" gorm:"size:32;not null;uniqueIndex"` // Public key ID (e.g., "tak_abc123...")
	KeyPrefix string `json:"key_prefix" gorm:"size:10;not null"`          // First 8 chars for display
	KeyHash   string `json:"-" gorm:"size:255;not null;uniqueIndex"`      // SHA-256 hash, never expose

	// Key metadata
	Name        string `json:"name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`

	// Scope and permissions
	Scopes json.RawMessage `json:"scopes" gorm:"type:jsonb;default:'["trust:read"]'::jsonb"`

	// IP allowlist
	AllowedIPs json.RawMessage `json:"allowed_ips" gorm:"type:jsonb;default:'[]'::jsonb"`

	// Expiration and revocation
	ExpiresAt    *time.Time `json:"expires_at"`
	IsRevoked    bool      `json:"is_revoked" gorm:"default:false"`
	RevokedAt    *time.Time `json:"revoked_at"`
	RevokedReason string   `json:"revoked_reason,omitempty" gorm:"type:text"`

	// Usage tracking
	LastUsedAt *time.Time `json:"last_used_at"`
	UseCount   int       `json:"use_count" gorm:"default:0"`

	// Created by
	CreatedBy string `json:"created_by" gorm:"size:255;not null"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Partner *TrustAPIPartner `json:"partner,omitempty" gorm:"foreignKey:PartnerID"`
}

// TableName returns the table name for TrustAPIKey
func (TrustAPIKey) TableName() string {
	return "trust_api_keys"
}

// DefaultAPIKeyScopes returns the default scopes for a new API key
func DefaultAPIKeyScopes() []string {
	return []string{"trust:read"}
}

// AllAPIKeyScopes lists all available scopes
var AllAPIKeyScopes = []string{
	"trust:read",         // Read trust scores and history
	"trust:write",        // Submit trust reports
	"verification:request", // Request function verification
	"reports:submit",     // Submit trust issue reports
	"partners:manage",    // Manage partner account (admin only)
}

// HasScope checks if the key has a specific scope
func (k *TrustAPIKey) HasScope(scope string) bool {
	var scopes []string
	if err := json.Unmarshal(k.Scopes, &scopes); err != nil {
		return false
	}
	for _, s := range scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// TrustAPIUsage represents detailed API usage tracking
type TrustAPIUsage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PartnerID uuid.UUID `json:"partner_id" gorm:"type:uuid;not null"`
	APIKeyID  *uuid.UUID `json:"api_key_id,omitempty" gorm:"type:uuid"`

	// Endpoint and method
	Endpoint string `json:"endpoint" gorm:"size:255;not null"`
	Method   string `json:"method" gorm:"size:10;not null"`

	// Request details
	RequestID string    `json:"request_id" gorm:"size:255;not null;uniqueIndex"`
	FunctionID *uuid.UUID `json:"function_id,omitempty" gorm:"type:uuid"`

	// Response details
	StatusCode      int    `json:"status_code" gorm:"not null"`
	ResponseTimeMs int    `json:"response_time_ms" gorm:"not null"`

	// Rate limit tracking
	RateLimitRemaining *int        `json:"rate_limit_remaining,omitempty"`
	RateLimitResetAt    *time.Time `json:"rate_limit_reset_at,omitempty"`

	// Request context
	IPAddress string `json:"ip_address" gorm:"size:45"`
	UserAgent string `json:"user_agent" gorm:"type:text"`

	// Error tracking
	ErrorCode    string `json:"error_code,omitempty" gorm:"size:50"`
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	Partner *TrustAPIPartner `json:"partner,omitempty" gorm:"foreignKey:PartnerID"`
	APIKey  *TrustAPIKey    `json:"api_key,omitempty" gorm:"foreignKey:APIKeyID"`
}

// TableName returns the table name for TrustAPIUsage
func (TrustAPIUsage) TableName() string {
	return "trust_api_usage"
}

// RateLimitType represents the type of rate limit window
type RateLimitType string

const (
	RateLimitMinute RateLimitType = "minute"
	RateLimitHour   RateLimitType = "hour"
	RateLimitDay    RateLimitType = "day"
	RateLimitMonth  RateLimitType = "month"
)

// TrustAPIRateLimit tracks sliding window rate limits
type TrustAPIRateLimit struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PartnerID uuid.UUID `json:"partner_id" gorm:"type:uuid;not null"`

	// Rate limit type
	LimitType string `json:"limit_type" gorm:"size:50;not null"`

	// Window tracking
	WindowStart time.Time `json:"window_start" gorm:"not null"`
	WindowEnd   time.Time `json:"window_end" gorm:"not null"`

	// Request count
	RequestCount int `json:"request_count" gorm:"default:0"`

	// Unique constraint
	// CONSTRAINT uq_trust_api_rate_limits_partner_type_window UNIQUE (partner_id, limit_type, window_start)
}

// TableName returns the table name for TrustAPIRateLimit
func (TrustAPIRateLimit) TableName() string {
	return "trust_api_rate_limits"
}

// ReportType represents the type of trust issue report
type ReportType string

const (
	ReportTypeMalware         ReportType = "malware"
	ReportTypePhishing       ReportType = "phishing"
	ReportTypeDataLeak       ReportType = "data_leak"
	ReportTypeAbuse          ReportType = "abuse"
	ReportTypeMisinformation ReportType = "misinformation"
	ReportTypeOther          ReportType = "other"
)

// ReportStatus represents the status of a trust report
type ReportStatus string

const (
	ReportStatusPending      ReportStatus = "pending"
	ReportStatusInvestigating ReportStatus = "investigating"
	ReportStatusResolved     ReportStatus = "resolved"
	ReportStatusDismissed    ReportStatus = "dismissed"
	ReportStatusEscalated    ReportStatus = "escalated"
)

// TrustAPIReport represents a trust issue report from a partner
type TrustAPIReport struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PartnerID uuid.UUID `json:"partner_id" gorm:"type:uuid;not null"`
	APIKeyID  *uuid.UUID `json:"api_key_id,omitempty" gorm:"type:uuid"`

	// Report identification
	ReportID string `json:"report_id" gorm:"size:32;not null;uniqueIndex"`

	// Function being reported
	FunctionID      uuid.UUID `json:"function_id" gorm:"type:uuid;not null"`
	FunctionAuthor  string    `json:"function_author,omitempty" gorm:"size:255"`
	FunctionName    string    `json:"function_name,omitempty" gorm:"size:255"`

	// Report details
	ReportType  string `json:"report_type" gorm:"size:50;not null"`
	Severity   string `json:"severity" gorm:"size:20;not null;default:'medium'"`
	Title      string `json:"title" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text;not null"`
	Evidence   json.RawMessage `json:"evidence" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Status tracking
	Status          string     `json:"status" gorm:"size:50;not null;default:'pending'"`
	ResolutionNotes string     `json:"resolution_notes,omitempty" gorm:"type:text"`
	ResolvedBy      *uuid.UUID `json:"resolved_by,omitempty" gorm:"type:uuid"`
	ResolvedAt      *time.Time `json:"resolved_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Partner *TrustAPIPartner `json:"partner,omitempty" gorm:"foreignKey:PartnerID"`
	APIKey  *TrustAPIKey    `json:"api_key,omitempty" gorm:"foreignKey:APIKeyID"`
}

// TableName returns the table name for TrustAPIReport
func (TrustAPIReport) TableName() string {
	return "trust_api_reports"
}

// VerificationLevel represents the level of verification
type VerificationLevel string

const (
	VerificationLevelBasic     VerificationLevel = "basic"
	VerificationLevelStandard  VerificationLevel = "standard"
	VerificationLevelAdvanced  VerificationLevel = "advanced"
	VerificationLevelEnterprise VerificationLevel = "enterprise"
)

// VerificationStatus represents the status of a verification request
type VerificationStatus string

const (
	VerificationStatusPending     VerificationStatus = "pending"
	VerificationStatusInProgress  VerificationStatus = "in_progress"
	VerificationStatusCompleted   VerificationStatus = "completed"
	VerificationStatusFailed      VerificationStatus = "failed"
	VerificationStatusCancelled    VerificationStatus = "cancelled"
)

// TrustAPIVerification represents a function verification request from a partner
type TrustAPIVerification struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PartnerID uuid.UUID `json:"partner_id" gorm:"type:uuid;not null"`
	APIKeyID  *uuid.UUID `json:"api_key_id,omitempty" gorm:"type:uuid"`

	// Verification identification
	VerificationID string `json:"verification_id" gorm:"size:32;not null;uniqueIndex"`

	// Function to verify
	FunctionID      uuid.UUID `json:"function_id" gorm:"type:uuid;not null"`
	FunctionAuthor  string    `json:"function_author,omitempty" gorm:"size:255"`
	FunctionName    string    `json:"function_name,omitempty" gorm:"size:255"`
	FunctionVersion string    `json:"function_version,omitempty" gorm:"size:50"`

	// Verification details
	VerificationLevel string `json:"verification_level" gorm:"size:50;not null;default:'standard'"`

	// Request content
	Metadata json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Status tracking
	Status string `json:"status" gorm:"size:50;not null;default:'pending'"`

	// Result
	TrustScore          *float64   `json:"trust_score,omitempty"`
	TrustTier           string     `json:"trust_tier,omitempty" gorm:"size:50"`
	VerificationBadgeURL string    `json:"verification_badge_url,omitempty" gorm:"size:500"`
	CompletionNotes     string     `json:"completion_notes,omitempty" gorm:"type:text"`
	CompletedBy         *uuid.UUID `json:"completed_by,omitempty" gorm:"type:uuid"`
	CompletedAt         *time.Time `json:"completed_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Partner *TrustAPIPartner `json:"partner,omitempty" gorm:"foreignKey:PartnerID"`
	APIKey  *TrustAPIKey    `json:"api_key,omitempty" gorm:"foreignKey:APIKeyID"`
}

// TableName returns the table name for TrustAPIVerification
func (TrustAPIVerification) TableName() string {
	return "trust_api_verifications"
}

// ============================================
// Request/Response DTOs
// ============================================

// PartnerCreateRequest represents a request to create a new partner
type PartnerCreateRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=255"`
	Slug        string `json:"slug" binding:"required,min=2,max=100,alphanum"`
	Description string `json:"description"`
	ContactEmail string `json:"contact_email" binding:"required,email"`
	ContactName string `json:"contact_name"`
	WebsiteURL  string `json:"website_url" binding:"omitempty,url"`
	Tier        string `json:"tier" binding:"omitempty,oneof=developer startup business enterprise"`
}

// PartnerResponse represents a partner in API responses
type PartnerResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Slug                string     `json:"slug"`
	Description         string     `json:"description,omitempty"`
	ContactEmail        string     `json:"contact_email"`
	ContactName         string     `json:"contact_name,omitempty"`
	WebsiteURL          string     `json:"website_url,omitempty"`
	Tier                string     `json:"tier"`
	RateLimitPerMinute  int        `json:"rate_limit_per_minute"`
	RateLimitPerDay     int        `json:"rate_limit_per_day"`
	MonthlyRequestLimit int        `json:"monthly_request_limit"`
	CurrentMonthUsage   int        `json:"current_month_usage"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	ActivatedAt         *time.Time `json:"activated_at,omitempty"`
}

// APIKeyCreateRequest represents a request to create a new API key
type APIKeyCreateRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=255"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes" binding:"required,min=1"`
	AllowedIPs  []string `json:"allowed_ips"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// APIKeyResponse represents an API key in responses (without the actual key)
type APIKeyResponse struct {
	ID          uuid.UUID  `json:"id"`
	KeyID       string     `json:"key_id"`  // Public key ID (e.g., "tak_abc123...")
	KeyPrefix   string     `json:"key_prefix"` // First 8 chars for display
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Scopes      []string   `json:"scopes"`
	AllowedIPs  []string   `json:"allowed_ips"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsRevoked   bool       `json:"is_revoked"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UseCount    int        `json:"use_count"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   string     `json:"created_by"`
}

// APIKeyCreatedResponse represents the response when a new API key is created (includes the actual key)
type APIKeyCreatedResponse struct {
	APIKeyResponse
	Key        string     `json:"key"` // The actual API key, only shown once
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// UsageResponse represents API usage statistics
type UsageResponse struct {
	PartnerID         uuid.UUID `json:"partner_id"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	TotalRequests     int64     `json:"total_requests"`
	SuccessfulRequests int64    `json:"successful_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	AverageLatencyMs  float64   `json:"average_latency_ms"`
	RateLimitHits     int64     `json:"rate_limit_hits"`
	TopEndpoints      []EndpointUsage `json:"top_endpoints"`
}

// EndpointUsage represents usage stats for a specific endpoint
type EndpointUsage struct {
	Endpoint string `json:"endpoint"`
	Count    int64  `json:"count"`
}

// TrustScoreResponse represents trust score data from the API
type TrustScoreResponse struct {
	FunctionID         uuid.UUID `json:"function_id"`
	TrustScore         float64   `json:"trust_score"`
	TrustTier          string    `json:"trust_tier"`
	IsVerified        bool      `json:"is_verified"`
	VerificationLevel  string    `json:"verification_level,omitempty"`
	LastUpdated       time.Time `json:"last_updated"`
	Components        TrustComponents `json:"components"`
	Metrics           TrustMetrics `json:"metrics"`
}

// TrustComponents represents individual trust score components
type TrustComponents struct {
	Reliability    float64 `json:"reliability"`
	Latency        float64 `json:"latency"`
	ErrorRate      float64 `json:"error_rate"`
	UserRating     float64 `json:"user_rating"`
	Verification   float64 `json:"verification"`
}

// TrustMetrics represents trust-related metrics
type TrustMetrics struct {
	TotalCalls    int64   `json:"total_calls"`
	SuccessRate   float64 `json:"success_rate"`
	P50LatencyMs  float64 `json:"p50_latency_ms"`
	P95LatencyMs  float64 `json:"p95_latency_ms"`
	P99LatencyMs  float64 `json:"p99_latency_ms"`
	ErrorRate     float64 `json:"error_rate"`
	TimeoutRate   float64 `json:"timeout_rate"`
}

// TrustHistoryResponse represents trust score history
type TrustHistoryResponse struct {
	FunctionID  uuid.UUID        `json:"function_id"`
	History     []TrustHistoryItem `json:"history"`
	TotalCount  int64             `json:"total_count"`
	Page        int               `json:"page"`
	PageSize    int               `json:"page_size"`
}

// TrustHistoryItem represents a single trust history entry
type TrustHistoryItem struct {
	TrustScore     float64   `json:"trust_score"`
	TrustTier      string    `json:"trust_tier"`
	IsVerified     bool      `json:"is_verified"`
	CalculatedAt   time.Time `json:"calculated_at"`
	WindowStart    time.Time `json:"window_start"`
	WindowEnd      time.Time `json:"window_end"`
}

// ReportCreateRequest represents a trust report submission
type ReportCreateRequest struct {
	FunctionID   uuid.UUID `json:"function_id" binding:"required"`
	ReportType   string    `json:"report_type" binding:"required,oneof=malware phishing data_leak abuse misinformation other"`
	Severity     string    `json:"severity" binding:"required,oneof=low medium high critical"`
	Title        string    `json:"title" binding:"required,min=5,max=255"`
	Description  string    `json:"description" binding:"required,min=10"`
	Evidence     map[string]interface{} `json:"evidence"`
}

// ReportResponse represents a trust report in responses
type ReportResponse struct {
	ID           uuid.UUID  `json:"id"`
	ReportID     string     `json:"report_id"`
	FunctionID   uuid.UUID  `json:"function_id"`
	FunctionAuthor string   `json:"function_author,omitempty"`
	FunctionName string     `json:"function_name,omitempty"`
	ReportType   string     `json:"report_type"`
	Severity     string     `json:"severity"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

// VerificationRequest represents a verification request submission
type VerificationRequest struct {
	FunctionID         uuid.UUID `json:"function_id" binding:"required"`
	FunctionVersion    string    `json:"function_version,omitempty"`
	VerificationLevel string    `json:"verification_level" binding:"required,oneof=basic standard advanced enterprise"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationResponse represents a verification request in responses
type VerificationResponse struct {
	ID                 uuid.UUID  `json:"id"`
	VerificationID     string     `json:"verification_id"`
	FunctionID         uuid.UUID  `json:"function_id"`
	FunctionAuthor     string     `json:"function_author,omitempty"`
	FunctionName       string     `json:"function_name,omitempty"`
	FunctionVersion    string     `json:"function_version,omitempty"`
	VerificationLevel  string     `json:"verification_level"`
	Status             string     `json:"status"`
	TrustScore         *float64   `json:"trust_score,omitempty"`
	TrustTier          string     `json:"trust_tier,omitempty"`
	VerificationBadgeURL string   `json:"verification_badge_url,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

// BatchTrustScoreRequest represents a batch trust score lookup request
type BatchTrustScoreRequest struct {
	FunctionIDs []uuid.UUID `json:"function_ids" binding:"required,min=1,max=100"`
}

// BatchTrustScoreResponse represents batch trust score results
type BatchTrustScoreResponse struct {
	Scores []TrustScoreResponse `json:"scores"`
	Errors []BatchTrustScoreError `json:"errors,omitempty"`
}

// BatchTrustScoreError represents an error for a specific function in batch lookup
type BatchTrustScoreError struct {
	FunctionID uuid.UUID `json:"function_id"`
	Error      string    `json:"error"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// RateLimitResponse represents rate limit status in responses
type RateLimitResponse struct {
	Limit     int        `json:"limit"`
	Remaining int        `json:"remaining"`
	ResetAt   time.Time  `json:"reset_at"`
	Tier      string     `json:"tier"`
}
