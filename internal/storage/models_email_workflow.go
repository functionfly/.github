package storage

import (
	"time"

	"github.com/google/uuid"
)

// EmailWorkflowConfig represents an email workflow configuration for a tenant.
// Each bundle type has pre-configured email workflows that are auto-provisioned
// when a bundle is deployed. Workflows are tenant-isolated and can be customized.
type EmailWorkflowConfig struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;index"`
	BundleSlug  string    `json:"bundle_slug" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Trigger     string    `json:"trigger" gorm:"not null"`    // "on_signup", "on_payment", "on_milestone", "on_inactivity", "manual"
	Category    string    `json:"category" gorm:"not null"`   // "onboarding", "billing", "engagement", "retention", "security"
	DelayDays   int       `json:"delay_days" gorm:"default:0"` // 0 = immediate, positive = delay in days
	Active      bool      `json:"active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for EmailWorkflowConfig
func (EmailWorkflowConfig) TableName() string {
	return "email_workflow_configs"
}

// EmailWorkflowExecution represents a single execution of an email workflow
type EmailWorkflowExecution struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID      uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	WorkflowID    uuid.UUID  `json:"workflow_id" gorm:"type:uuid;not null;index"`
	Recipient     string     `json:"recipient" gorm:"not null"`
	Status        string     `json:"status" gorm:"not null;default:'pending'"` // "pending", "sent", "failed", "cancelled"
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

// TableName returns the table name for EmailWorkflowExecution
func (EmailWorkflowExecution) TableName() string {
	return "email_workflow_executions"
}
