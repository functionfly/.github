package trustapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Webhook Event Types
// ============================================

// WebhookEventType represents the type of webhook event
type WebhookEventType string

const (
	// Revocation events
	WebhookEventRevocationCreated  WebhookEventType = "trust.revocation.created"
	WebhookEventRevocationLifted   WebhookEventType = "trust.revocation.lifted"
	WebhookEventRevocationExpired  WebhookEventType = "trust.revocation.expired"
	WebhookEventRevocationAppealed WebhookEventType = "trust.revocation.appealed"

	// Attestation events
	WebhookEventAttestationCreated WebhookEventType = "trust.attestation.created"
	WebhookEventAttestationRevoked WebhookEventType = "trust.attestation.revoked"

	// Policy events
	WebhookEventPolicyCreated   WebhookEventType = "trust.policy.created"
	WebhookEventPolicyUpdated   WebhookEventType = "trust.policy.updated"
	WebhookEventPolicyDeleted   WebhookEventType = "trust.policy.deleted"
	WebhookEventPolicyEvaluated WebhookEventType = "trust.policy.evaluated"

	// Trust score events
	WebhookEventTrustScoreUpdated WebhookEventType = "trust.score.updated"
	WebhookEventTrustTierChanged  WebhookEventType = "trust.tier.changed"

	// Verification events
	WebhookEventVerificationCompleted WebhookEventType = "trust.verification.completed"
	WebhookEventVerificationFailed    WebhookEventType = "trust.verification.failed"

	// Report events
	WebhookEventReportSubmitted WebhookEventType = "trust.report.submitted"
	WebhookEventReportResolved  WebhookEventType = "trust.report.resolved"
	WebhookEventReportEscalated WebhookEventType = "trust.report.escalated"
)

// WebhookStatus represents the status of a webhook
type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
	WebhookStatusFailed   WebhookStatus = "failed"
	WebhookStatusDisabled WebhookStatus = "disabled"
)

// ============================================
// Webhook Configuration
// ============================================

// TrustWebhook represents a webhook configuration for trust events
type TrustWebhook struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Webhook identification
	WebhookID   string `json:"webhook_id" gorm:"size:32;not null;uniqueIndex"` // Public ID (e.g., "whk_abc123...")
	Name        string `json:"name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`

	// Owner
	OwnerID        uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null;index:idx_webhook_owner"`
	OwnerType      string     `json:"owner_type" gorm:"size:20;not null;default:'user'"` // user, partner, team
	OwnerPartnerID *uuid.UUID `json:"owner_partner_id,omitempty" gorm:"type:uuid"`

	// Endpoint configuration
	URL    string `json:"url" gorm:"size:500;not null"`
	Method string `json:"method" gorm:"size:10;not null;default:'POST'"`
	Secret string `json:"-" gorm:"size:255;not null"` // HMAC secret for signing

	// Event filtering
	Events      json.RawMessage `json:"events" gorm:"type:jsonb;not null"`                   // Array of event types to subscribe to
	EventFilter string          `json:"event_filter,omitempty" gorm:"size:50;default:'all'"` // all, specific

	// Function filtering (optional - only receive events for specific functions)
	FunctionFilter json.RawMessage `json:"function_filter,omitempty" gorm:"type:jsonb;default:'[]'::jsonb"` // Array of function IDs

	// Status
	Status      string     `json:"status" gorm:"size:20;not null;default:'active'"`
	FailCount   int        `json:"fail_count" gorm:"default:0"`
	LastFailure *time.Time `json:"last_failure,omitempty"`
	LastSuccess *time.Time `json:"last_success,omitempty"`

	// Retry configuration
	MaxRetries     int `json:"max_retries" gorm:"default:3"`
	RetryDelaySecs int `json:"retry_delay_secs" gorm:"default:60"`
	TimeoutSecs    int `json:"timeout_secs" gorm:"default:30"`

	// Payload customization
	IncludePayload bool            `json:"include_payload" gorm:"default:true"`
	CustomHeaders  json.RawMessage `json:"custom_headers,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for TrustWebhook
func (TrustWebhook) TableName() string {
	return "trust_webhooks"
}

// IsSubscribedToEvent checks if webhook is subscribed to a specific event type
func (w *TrustWebhook) IsSubscribedToEvent(eventType WebhookEventType) bool {
	var events []string
	if err := json.Unmarshal(w.Events, &events); err != nil {
		return false
	}

	for _, e := range events {
		if e == string(eventType) || e == "*" {
			return true
		}
	}
	return false
}

// IsFilteredForFunction checks if webhook is filtered for a specific function
func (w *TrustWebhook) IsFilteredForFunction(functionID uuid.UUID) bool {
	var functions []string
	if err := json.Unmarshal(w.FunctionFilter, &functions); err != nil {
		return false
	}

	if len(functions) == 0 {
		return true // No filter means all functions
	}

	for _, f := range functions {
		if f == functionID.String() {
			return true
		}
	}
	return false
}

// GenerateSignature generates HMAC signature for webhook payload
func (w *TrustWebhook) GenerateSignature(payload []byte) string {
	h := hmac.New(sha256.New, []byte(w.Secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================
// Webhook Delivery
// ============================================

// WebhookDeliveryStatus represents the status of a webhook delivery
type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatusPending   WebhookDeliveryStatus = "pending"
	WebhookDeliveryStatusSent      WebhookDeliveryStatus = "sent"
	WebhookDeliveryStatusDelivered WebhookDeliveryStatus = "delivered"
	WebhookDeliveryStatusFailed    WebhookDeliveryStatus = "failed"
	WebhookDeliveryStatusRetrying  WebhookDeliveryStatus = "retrying"
)

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Delivery identification
	DeliveryID string `json:"delivery_id" gorm:"size:32;not null;uniqueIndex"` // Public ID (e.g., "del_abc123...")

	// References
	WebhookID uuid.UUID `json:"webhook_id" gorm:"type:uuid;not null;index:idx_delivery_webhook"`
	EventType string    `json:"event_type" gorm:"size:50;not null"`
	EntityID  string    `json:"entity_id,omitempty" gorm:"size:255"` // ID of the entity that triggered the event

	// Payload
	Payload     json.RawMessage `json:"payload" gorm:"type:jsonb;not null"`
	PayloadHash string          `json:"payload_hash" gorm:"size:64;not null"` // SHA-256 of payload for integrity

	// Delivery attempt
	AttemptNumber int    `json:"attempt_number" gorm:"default:1"`
	MaxAttempts   int    `json:"max_attempts" gorm:"default:3"`
	Status        string `json:"status" gorm:"size:20;not null;default:'pending'"`

	// Response tracking
	ResponseStatusCode *int            `json:"response_status_code,omitempty"`
	ResponseHeaders    json.RawMessage `json:"response_headers,omitempty" gorm:"type:jsonb"`
	ResponseBody       string          `json:"response_body,omitempty" gorm:"type:text"`
	ResponseTimeMs     int             `json:"response_time_ms,omitempty"`

	// Error tracking
	ErrorMessage string `json:"error_message,omitempty" gorm:"type:text"`

	// Timing
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for WebhookDelivery
func (WebhookDelivery) TableName() string {
	return "trust_webhook_deliveries"
}

// ============================================
// Webhook Event Payload
// ============================================

// WebhookPayload represents the standard webhook event payload
type WebhookPayload struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	Timestamp  time.Time `json:"timestamp"`
	APIVersion string    `json:"api_version"`

	Data map[string]interface{} `json:"data"`
}

// NewWebhookPayload creates a new webhook payload
func NewWebhookPayload(eventType WebhookEventType, data map[string]interface{}) *WebhookPayload {
	return &WebhookPayload{
		EventID:    generateEventID(),
		EventType:  string(eventType),
		Timestamp:  time.Now(),
		APIVersion: "2024-04-12",
		Data:       data,
	}
}

func generateEventID() string {
	return "evt_" + uuid.New().String()[:24]
}

// ============================================
// Webhook DTOs
// ============================================

// WebhookCreateRequest represents a request to create a webhook
type WebhookCreateRequest struct {
	Name           string            `json:"name" binding:"required,min=2,max=255"`
	Description    string            `json:"description" binding:"max=1000"`
	URL            string            `json:"url" binding:"required,url,max=500"`
	Events         []string          `json:"events" binding:"required,min=1"`
	Secret         string            `json:"secret,omitempty" binding:"min=32,max=255"`
	FunctionFilter []string          `json:"function_filter,omitempty"`
	MaxRetries     int               `json:"max_retries,omitempty" binding:"min=0,max=10"`
	CustomHeaders  map[string]string `json:"custom_headers,omitempty"`
}

// WebhookUpdateRequest represents a request to update a webhook
type WebhookUpdateRequest struct {
	Name           string            `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Description    string            `json:"description,omitempty" binding:"omitempty,max=1000"`
	URL            string            `json:"url,omitempty" binding:"omitempty,url,max=500"`
	Events         []string          `json:"events,omitempty" binding:"omitempty,min=1"`
	Secret         string            `json:"secret,omitempty" binding:"omitempty,min=32,max=255"`
	FunctionFilter []string          `json:"function_filter,omitempty"`
	Status         string            `json:"status,omitempty" binding:"omitempty,oneof=active inactive disabled"`
	MaxRetries     int               `json:"max_retries,omitempty" binding:"omitempty,min=0,max=10"`
	CustomHeaders  map[string]string `json:"custom_headers,omitempty"`
}

// WebhookResponse represents a webhook in API responses
type WebhookResponse struct {
	ID             uuid.UUID         `json:"id"`
	WebhookID      string            `json:"webhook_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Events         []string          `json:"events"`
	EventFilter    string            `json:"event_filter"`
	Status         string            `json:"status"`
	FailCount      int               `json:"fail_count"`
	LastFailure    *time.Time        `json:"last_failure,omitempty"`
	LastSuccess    *time.Time        `json:"last_success,omitempty"`
	MaxRetries     int               `json:"max_retries"`
	IncludePayload bool              `json:"include_payload"`
	CustomHeaders  map[string]string `json:"custom_headers,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// WebhookListResponse represents a list of webhooks
type WebhookListResponse struct {
	Webhooks   []WebhookResponse `json:"webhooks"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// WebhookDeliveryResponse represents a webhook delivery in API responses
type WebhookDeliveryResponse struct {
	ID                 uuid.UUID  `json:"id"`
	DeliveryID         string     `json:"delivery_id"`
	WebhookID          uuid.UUID  `json:"webhook_id"`
	EventType          string     `json:"event_type"`
	EntityID           string     `json:"entity_id,omitempty"`
	AttemptNumber      int        `json:"attempt_number"`
	MaxAttempts        int        `json:"max_attempts"`
	Status             string     `json:"status"`
	ResponseStatusCode *int       `json:"response_status_code,omitempty"`
	ResponseTimeMs     int        `json:"response_time_ms,omitempty"`
	ErrorMessage       string     `json:"error_message,omitempty"`
	SentAt             *time.Time `json:"sent_at,omitempty"`
	DeliveredAt        *time.Time `json:"delivered_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// WebhookDeliveryListResponse represents a list of webhook deliveries
type WebhookDeliveryListResponse struct {
	Deliveries []WebhookDeliveryResponse `json:"deliveries"`
	TotalCount int64                     `json:"total_count"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
}

// WebhookTestRequest represents a request to test a webhook
type WebhookTestRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	TestData  map[string]interface{} `json:"test_data,omitempty"`
}

// WebhookTestResponse represents the response from testing a webhook
type WebhookTestResponse struct {
	Success        bool   `json:"success"`
	StatusCode     int    `json:"status_code,omitempty"`
	ResponseTimeMs int    `json:"response_time_ms,omitempty"`
	ResponseBody   string `json:"response_body,omitempty"`
	Error          string `json:"error,omitempty"`
	DeliveryID     string `json:"delivery_id"`
}
