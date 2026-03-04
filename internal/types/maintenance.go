package types

import (
	"time"

	"github.com/google/uuid"
)

// PlatformMaintenance represents the platform-wide maintenance mode configuration
type PlatformMaintenance struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Enabled           bool       `json:"enabled" db:"enabled"`
	Name              string     `json:"name" db:"name"`
	Description       *string    `json:"description" db:"description"`
	Message           *string    `json:"message" db:"message"`
	PageTemplate      string     `json:"page_template" db:"page_template"`
	RetryAfterSeconds int        `json:"retry_after_seconds" db:"retry_after_seconds"`
	RolloutPercentage int        `json:"rollout_percentage" db:"rollout_percentage"`
	RolloutSeed       *string    `json:"rollout_seed" db:"rollout_seed"`
	ScheduledStart    *time.Time `json:"scheduled_start" db:"scheduled_start"`
	ScheduledEnd      *time.Time `json:"scheduled_end" db:"scheduled_end"`
	IsScheduled       bool       `json:"is_scheduled" db:"is_scheduled"`
	RecurrenceRule    *string    `json:"recurrence_rule" db:"recurrence_rule"`
	Timezone          string     `json:"timezone" db:"timezone"`
	CreatedBy         *uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the database table name for PlatformMaintenance
func (PlatformMaintenance) TableName() string {
	return "platform_maintenance"
}

// MaintenancePageTemplate represents a customizable maintenance page template
type MaintenancePageTemplate struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Title           *string   `json:"title" db:"title"`
	MessageHTML     *string   `json:"message_html" db:"message_html"`
	LogoURL         *string   `json:"logo_url" db:"logo_url"`
	BackgroundColor string    `json:"background_color" db:"background_color"`
	TextColor       string    `json:"text_color" db:"text_color"`
	AccentColor     string    `json:"accent_color" db:"accent_color"`
	ShowContactInfo bool      `json:"show_contact_info" db:"show_contact_info"`
	ContactEmail    *string   `json:"contact_email" db:"contact_email"`
	ShowSocialLinks bool      `json:"show_social_links" db:"show_social_links"`
	IsDefault       bool      `json:"is_default" db:"is_default"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// MaintenanceAuditLog represents an audit log entry for maintenance mode changes
type MaintenanceAuditLog struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	MaintenanceID *uuid.UUID `json:"maintenance_id" db:"maintenance_id"`
	Action        string     `json:"action" db:"action"`
	OldValues     *string    `json:"old_values" db:"old_values"`
	NewValues     *string    `json:"new_values" db:"new_values"`
	ChangedBy     *uuid.UUID `json:"changed_by" db:"changed_by"`
	ChangedAt     time.Time  `json:"changed_at" db:"changed_at"`
	IPAddress     *string    `json:"ip_address" db:"ip_address"`
	UserAgent     *string    `json:"user_agent" db:"user_agent"`
}

// MaintenanceStatus represents the current maintenance status (for public API)
type MaintenanceStatus struct {
	MaintenanceMode      bool                   `json:"maintenance_mode"`
	ScheduledMaintenance []ScheduledMaintenance `json:"scheduled_maintenance,omitempty"`
}

// ScheduledMaintenance represents upcoming scheduled maintenance
type ScheduledMaintenance struct {
	Name           string    `json:"name"`
	ScheduledStart time.Time `json:"scheduled_start"`
	ScheduledEnd   time.Time `json:"scheduled_end"`
	Status         string    `json:"status"`
}

// MaintenanceConfig is the full configuration used by the middleware
type MaintenanceConfig struct {
	PlatformMaintenance
	Template *MaintenancePageTemplate `json:"template,omitempty"`
}

// MaintenanceAction represents the type of maintenance action
type MaintenanceAction string

const (
	MaintenanceActionEnabled   MaintenanceAction = "enabled"
	MaintenanceActionDisabled  MaintenanceAction = "disabled"
	MaintenanceActionUpdated   MaintenanceAction = "updated"
	MaintenanceActionScheduled MaintenanceAction = "scheduled"
	MaintenanceActionCancelled MaintenanceAction = "cancelled"
)
