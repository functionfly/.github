package identity

import (
	"time"

	"github.com/google/uuid"
)

// AgentIdentity represents a registered agent in the system
type AgentIdentity struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	AgentID     string    `json:"agent_id" gorm:"uniqueIndex;not null"` // "org/agent-name"
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	PlanTier    string    `json:"plan_tier" gorm:"not null;default:'agent_starter'"`
	Status      string    `json:"status" gorm:"not null;default:'active'"` // active | suspended | deleted
	APIKeyHash  string    `json:"-" gorm:"column:api_key_hash"`             // hashed API key, never returned
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentIdentity) TableName() string {
	return "agent_identities"
}

// AgentQuotaConfig holds per-agent quota configuration
type AgentQuotaConfig struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID              string    `json:"agent_id" gorm:"uniqueIndex;not null"`
	MaxCallsPerMinute    int       `json:"max_calls_per_minute" gorm:"not null;default:100"`
	MaxCallsPerDay       int       `json:"max_calls_per_day" gorm:"not null;default:16667"`
	MaxStateWritesPerHr  int       `json:"max_state_writes_per_hr" gorm:"not null;default:1000"`
	MaxCostPerExecution  float64   `json:"max_cost_per_execution" gorm:"type:decimal(10,6);not null;default:0.01"`
	MaxDailySpendUSD     float64   `json:"max_daily_spend_usd" gorm:"type:decimal(10,2);not null;default:5.00"`
	AllowedFunctions     []string  `json:"allowed_functions,omitempty" gorm:"type:text[]"`
	ForbiddenFunctions   []string  `json:"forbidden_functions,omitempty" gorm:"type:text[]"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentQuotaConfig) TableName() string {
	return "agent_quota_configs"
}

// RegisterAgentRequest is the request body for registering a new agent
type RegisterAgentRequest struct {
	AgentID     string `json:"agent_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	PlanTier    string `json:"plan_tier,omitempty"`
}

// RegisterAgentResponse is the response after registering an agent
type RegisterAgentResponse struct {
	OK      bool           `json:"ok"`
	Agent   *AgentIdentity `json:"agent"`
	APIKey  string         `json:"api_key,omitempty"` // Only returned on creation
	Message string         `json:"message,omitempty"`
}

// AgentStatus constants
const (
	AgentStatusActive    = "active"
	AgentStatusSuspended = "suspended"
	AgentStatusDeleted   = "deleted"
)
