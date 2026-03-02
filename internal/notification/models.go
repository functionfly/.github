package notification

import (
	"time"

	"github.com/google/uuid"
)

// JSONMap is a helper type for JSONB fields
type JSONMap map[string]interface{}

// StringArray is a helper type for text[] fields
type StringArray []string

// Notification stores individual notifications
type Notification struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Type      string    `json:"type" gorm:"not null;index"`      // e.g., "deployment.success"
	Category  string    `json:"category" gorm:"not null;index"`  // e.g., "deployments"
	Title     string    `json:"title" gorm:"not null"`
	Body      string    `json:"body" gorm:"type:text"`
	Data      JSONMap   `json:"data" gorm:"type:jsonb"`
	Channels  StringArray `json:"channels" gorm:"type:text[]"`
	Priority  string    `json:"priority" gorm:"not null;default:'normal'"` // low, normal, high, urgent
	Status    string    `json:"status" gorm:"not null;default:'pending'"`  // pending, processing, sent, failed, read
	ReadAt    *time.Time `json:"read_at,omitempty"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NotificationPreference stores user preferences per channel/category
type NotificationPreference struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_channel_category"`
	Channel         string    `json:"channel" gorm:"not null;uniqueIndex:idx_user_channel_category"` // email, in_app, webhook, push
	Category        string    `json:"category" gorm:"not null;uniqueIndex:idx_user_channel_category"` // system, security, billing, etc.
	Enabled         bool      `json:"enabled" gorm:"default:true"`
	Frequency       string    `json:"frequency" gorm:"default:'immediate'"` // immediate, digest_daily, digest_weekly
	QuietHoursStart *string   `json:"quiet_hours_start,omitempty"` // HH:MM format
	QuietHoursEnd   *string   `json:"quiet_hours_end,omitempty"` // HH:MM format
	Timezone        string    `json:"timezone" gorm:"default:'UTC'"`
	WebhookURL      *string   `json:"webhook_url,omitempty"`
	WebhookSecret   *string   `json:"webhook_secret,omitempty"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NotificationTemplate stores templates for each notification type/channel
type NotificationTemplate struct {
	ID       uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type     string      `json:"type" gorm:"not null;uniqueIndex:idx_type_channel"` // e.g., "deployment.success"
	Channel  string      `json:"channel" gorm:"not null;uniqueIndex:idx_type_channel"` // email, in_app, webhook
	Subject  string      `json:"subject"` // For email notifications
	BodyHTML string      `json:"body_html" gorm:"type:text"`
	BodyText string      `json:"body_text" gorm:"type:text"`
	Variables StringArray `json:"variables" gorm:"type:text[]"`
	IsActive bool        `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// NotificationAnalytics tracks delivery and engagement
type NotificationAnalytics struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	NotificationID uuid.UUID  `json:"notification_id" gorm:"type:uuid;not null;index"`
	Channel        string     `json:"channel" gorm:"not null"`
	Status         string     `json:"status" gorm:"not null"` // delivered, failed, bounced, opened, clicked
	ErrorMessage   *string    `json:"error_message,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ClickedAt      *time.Time `json:"clicked_at,omitempty"`
	IPAddress      *string    `json:"ip_address,omitempty"`
	UserAgent      *string    `json:"user_agent,omitempty"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// ListOptions provides pagination and filtering for listing notifications
type ListOptions struct {
	Limit      int
	Offset     int
	Status     string
	Category   string
	UnreadOnly bool
}

// SendRequest represents a request to send a notification
type SendRequest struct {
	UserID   uuid.UUID
	Type     string
	Category string
	Title    string
	Body     string
	Data     JSONMap
	Channels []string
	Priority string
}

// BroadcastRequest represents a request to broadcast to multiple users
type BroadcastRequest struct {
	UserIDs  []uuid.UUID
	Type     string
	Category string
	Title    string
	Body     string
	Data     JSONMap
	Channels []string
	Priority string
}
