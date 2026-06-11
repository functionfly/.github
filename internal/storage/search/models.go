package search

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Execution represents a search tool execution record
type Execution struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ToolName        string          `json:"tool_name" gorm:"type:varchar(50);not null;index:idx_search_executions_tool"`
	Query           string          `json:"query" gorm:"type:text;not null"`
	Parameters      json.RawMessage `json:"parameters" gorm:"type:jsonb"`
	ResultsCount    int             `json:"results_count" gorm:"default:0"`
	CreditsUsed     float64         `json:"credits_used" gorm:"type:decimal(10,4);not null;default:0"`
	ExecutionTimeMs int             `json:"execution_time_ms" gorm:"not null;default:0"`
	AgentID         *uuid.UUID      `json:"agent_id" gorm:"type:uuid;index:idx_search_executions_agent"`
	CreatedAt       time.Time       `json:"created_at" gorm:"type:timestamp with time zone;default:now();index:idx_search_executions_created"`
}

// TableName returns the table name for Execution
func (Execution) TableName() string {
	return "search_executions"
}

// CacheEntry represents a cached search result
type CacheEntry struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CacheKey   string          `json:"cache_key" gorm:"type:varchar(255);unique;not null;index:idx_search_cache_key"`
	ToolName   string          `json:"tool_name" gorm:"type:varchar(50);not null;index:idx_search_cache_tool"`
	QueryHash  string          `json:"query_hash" gorm:"type:varchar(64);not null"`
	Parameters json.RawMessage `json:"parameters" gorm:"type:jsonb"`
	Results    json.RawMessage `json:"results" gorm:"type:jsonb;not null"`
	CachedAt   time.Time       `json:"cached_at" gorm:"type:timestamp with time zone;default:now()"`
	ExpiresAt  time.Time       `json:"expires_at" gorm:"type:timestamp with time zone;not null;index:idx_search_cache_expiry"`
}

// TableName returns the table name for CacheEntry
func (CacheEntry) TableName() string {
	return "search_result_cache"
}