package types

import (
	"encoding/json"
	"time"
)

// ExecutionResult represents a cached function execution result
type ExecutionResult struct {
	FunctionID    string          `json:"function_id"`
	Version       string          `json:"version"`
	InputHash     string          `json:"input_hash"`
	Output        json.RawMessage `json:"output"`
	Error         string          `json:"error,omitempty"`
	DurationMs    int             `json:"duration_ms"`
	StatusCode    int             `json:"status_code"`
	ExecutedAt    time.Time       `json:"executed_at"`
	TTL           time.Duration   `json:"ttl"`
	ExpiresAt     time.Time       `json:"expires_at"`
	HitCount      int64           `json:"hit_count"`
	LastAccessed  time.Time       `json:"last_accessed"`
}