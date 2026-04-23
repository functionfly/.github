package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type FunctionWebhookSubscription struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID      `json:"tenant_id" gorm:"type:uuid;not null;index"`
	FunctionID *uuid.UUID     `json:"function_id,omitempty" gorm:"type:uuid;index"`
	URL        string         `json:"url" gorm:"type:text;not null"`
	Secret     string         `json:"-" gorm:"size:255;not null"`
	EventTypes pq.StringArray `json:"event_types" gorm:"type:text[];not null"`
	Active     bool           `json:"active" gorm:"not null;default:true"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime;not null"`
	CreatedBy  *uuid.UUID     `json:"created_by,omitempty" gorm:"type:uuid"`
}

func (FunctionWebhookSubscription) TableName() string {
	return "function_webhook_subscriptions"
}

func (s *FunctionWebhookSubscription) GenerateSignature(payload []byte) string {
	h := hmac.New(sha256.New, []byte(s.Secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

type FunctionWebhookDelivery struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SubscriptionID uuid.UUID       `json:"subscription_id" gorm:"type:uuid;not null;index"`
	EventType      string          `json:"event_type" gorm:"size:50;not null"`
	Payload        json.RawMessage `json:"payload" gorm:"type:jsonb;not null"`
	ResponseStatus *int            `json:"response_status,omitempty"`
	ResponseBody   *string         `json:"response_body,omitempty" gorm:"type:text"`
	AttemptedAt    time.Time       `json:"attempted_at" gorm:"autoCreateTime;not null"`
	Success        bool            `json:"success" gorm:"not null;default:false"`
	ErrorMessage   *string         `json:"error_message,omitempty" gorm:"type:text"`
}

func (FunctionWebhookDelivery) TableName() string {
	return "function_webhook_deliveries"
}

type FunctionWebhookCreateRequest struct {
	URL        string   `json:"url" binding:"required,url"`
	EventTypes []string `json:"event_types" binding:"required,min=1"`
	FunctionID *string  `json:"function_id,omitempty"`
	Secret     string   `json:"secret,omitempty"`
}

type FunctionWebhookUpdateRequest struct {
	URL        string   `json:"url,omitempty" binding:"omitempty,url"`
	EventTypes []string `json:"event_types,omitempty"`
	Active     *bool    `json:"active,omitempty"`
}

type FunctionWebhookResponse struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	FunctionID *uuid.UUID `json:"function_id,omitempty"`
	URL        string     `json:"url"`
	EventTypes []string   `json:"event_types"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
}

type FunctionWebhookListResponse struct {
	Subscriptions []FunctionWebhookResponse `json:"subscriptions"`
	TotalCount    int64                     `json:"total_count"`
	Page          int                       `json:"page"`
	PageSize      int                       `json:"page_size"`
}

type FunctionWebhookDeliveryResponse struct {
	ID             uuid.UUID       `json:"id"`
	SubscriptionID uuid.UUID       `json:"subscription_id"`
	EventType      string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload"`
	ResponseStatus *int            `json:"response_status,omitempty"`
	ResponseBody   *string         `json:"response_body,omitempty"`
	AttemptedAt    time.Time       `json:"attempted_at"`
	Success        bool            `json:"success"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
}

type FunctionWebhookDeliveryListResponse struct {
	Deliveries []FunctionWebhookDeliveryResponse `json:"deliveries"`
	TotalCount int64                             `json:"total_count"`
	Page       int                               `json:"page"`
	PageSize   int                               `json:"page_size"`
}

type FunctionWebhookTestRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	TestData  map[string]interface{} `json:"test_data,omitempty"`
}

type FunctionWebhookTestResponse struct {
	Success        bool   `json:"success"`
	StatusCode     int    `json:"status_code,omitempty"`
	ResponseTimeMs int    `json:"response_time_ms,omitempty"`
	ResponseBody   string `json:"response_body,omitempty"`
	Error          string `json:"error,omitempty"`
}
