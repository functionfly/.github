package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AgentFunctionCategory represents the category of an agent function
type AgentFunctionCategory string

const (
	AgentFunctionCategorySearch        AgentFunctionCategory = "search"
	AgentFunctionCategoryBrowser       AgentFunctionCategory = "browser"
	AgentFunctionCategoryFile          AgentFunctionCategory = "file"
	AgentFunctionCategoryData          AgentFunctionCategory = "data"
	AgentFunctionCategoryCompute        AgentFunctionCategory = "compute"
	AgentFunctionCategoryCommunication  AgentFunctionCategory = "communication"
	AgentFunctionCategoryWorkflow      AgentFunctionCategory = "workflow"
	AgentFunctionCategoryMemory        AgentFunctionCategory = "memory"
	AgentFunctionCategoryAssure         AgentFunctionCategory = "assure"
	AgentFunctionCategoryValidate      AgentFunctionCategory = "validate"
	AgentFunctionCategorySimulate      AgentFunctionCategory = "simulate"
	AgentFunctionCategoryObserve       AgentFunctionCategory = "observe"
	AgentFunctionCategoryLearn         AgentFunctionCategory = "learn"
	AgentFunctionCategoryAgentMgmt     AgentFunctionCategory = "agent_mgmt"
	AgentFunctionCategoryCapability   AgentFunctionCategory = "capability"
)

// AgentFunction represents an agent function registered in the platform
type AgentFunction struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID    uuid.UUID              `json:"function_id" gorm:"type:uuid;not null;uniqueIndex"`
	Category      AgentFunctionCategory  `json:"category" gorm:"type:varchar(50);not null;index"`
	Capabilities  []string              `json:"capabilities" gorm:"type:text[];not null;default:'{}'"`
	InputSchema   json.RawMessage       `json:"input_schema" gorm:"type:jsonb;not null;default:'{}'"`
	OutputSchema  json.RawMessage       `json:"output_schema" gorm:"type:jsonb;not null;default:'{}'"`
	IsVerified    bool                  `json:"is_verified" gorm:"default:false"`
	IsExclusive   bool                  `json:"is_exclusive" gorm:"default:false;index"`
	MaxConcurrency int                   `json:"max_concurrency" gorm:"default:10"`
	RateLimitRPM  int                    `json:"rate_limit_rpm" gorm:"default:60"`
	PricingModel  json.RawMessage       `json:"pricing_model" gorm:"type:jsonb;not null;default:'{\"type\": \"per_call\", \"price\": 0.001}'"`
	CreatedAt     time.Time             `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time             `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Function *RegistryFunction `json:"function,omitempty" gorm:"foreignKey:FunctionID"`
}

// TableName returns the table name for AgentFunction
func (AgentFunction) TableName() string {
	return "agent_functions"
}

// AgentFunctionExecution represents a single execution of an agent function
type AgentFunctionExecution struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID     uuid.UUID       `json:"agent_id" gorm:"type:uuid;not null;index"`
	FunctionID  uuid.UUID       `json:"function_id" gorm:"type:uuid;not null;index"`
	SessionID   *uuid.UUID      `json:"session_id,omitempty" gorm:"type:uuid;index"`
	Input       json.RawMessage `json:"input" gorm:"type:jsonb;not null;default:'{}'"`
	Output      json.RawMessage `json:"output,omitempty" gorm:"type:jsonb"`
	Error       string          `json:"error,omitempty" gorm:"type:text"`
	DurationMs  int             `json:"duration_ms"`
	CostUSD     float64         `json:"cost_usd" gorm:"type:decimal(10,6);default:0"`
	TraceID     string          `json:"trace_id,omitempty" gorm:"size:255"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime;index"`

	// Relationships
	Function *RegistryFunction `json:"function,omitempty" gorm:"foreignKey:FunctionID"`
}

// TableName returns the table name for AgentFunctionExecution
func (AgentFunctionExecution) TableName() string {
	return "agent_function_executions"
}

// AgentFunctionPolicy represents per-agent per-function policy
type AgentFunctionPolicy struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID          uuid.UUID  `json:"agent_id" gorm:"type:uuid;not null;index"`
	FunctionID       uuid.UUID  `json:"function_id" gorm:"type:uuid;not null"`
	Allowed          bool       `json:"allowed" gorm:"default:true"`
	MaxCallsPerDay   *int       `json:"max_calls_per_day,omitempty"`
	MaxCostPerCallUSD *float64  `json:"max_cost_per_call_usd,omitempty"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Function *RegistryFunction `json:"function,omitempty" gorm:"foreignKey:FunctionID"`
}

// TableName returns the table name for AgentFunctionPolicy
func (AgentFunctionPolicy) TableName() string {
	return "agent_function_policies"
}

// PricingModel represents the pricing model for a function
type PricingModel struct {
	Type  string  `json:"type"`  // 'per_call', 'per_second', 'per_unit'
	Price float64 `json:"price"`
}

// AgentFunctionDefinition is the API response format for function discovery
type AgentFunctionDefinition struct {
	Author        string          `json:"author"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	Description   string          `json:"description"`
	Category      string          `json:"category"`
	InputSchema   json.RawMessage `json:"input_schema"`
	OutputSchema  json.RawMessage `json:"output_schema"`
	Pricing       PricingModel    `json:"pricing"`
	Capabilities  []string       `json:"capabilities"`
	IsVerified    bool            `json:"is_verified"`
	IsExclusive   bool            `json:"is_exclusive"`
	MaxConcurrency int            `json:"max_concurrency"`
	RateLimitRPM  int             `json:"rate_limit_rpm"`
}

// AgentFunctionExecutionsFilter provides filtering options for execution queries
type AgentFunctionExecutionsFilter struct {
	AgentID    *uuid.UUID
	FunctionID *uuid.UUID
	SessionID  *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}