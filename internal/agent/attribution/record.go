package attribution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentExecutionRecord is the full attribution record for a single agent tool call.
// Hot metadata is stored in Postgres; the full record is written to object storage.
type AgentExecutionRecord struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID          string     `json:"agent_id" gorm:"not null"`
	TenantID         uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	FunctionID       uuid.UUID  `json:"function_id" gorm:"type:uuid;not null"`
	FunctionURI      string     `json:"function_uri"`        // fx://org/name@version
	ExecutionID      string     `json:"execution_id"`
	SessionID        string     `json:"session_id,omitempty"`
	CallDepth        int        `json:"call_depth" gorm:"not null;default:0"`
	InputHash        string     `json:"input_hash"`          // SHA-256 of input
	OutputHash       string     `json:"output_hash"`         // SHA-256 of output
	MemoryBeforeHash string     `json:"memory_before_hash"`  // SHA-256 of agent context before
	MemoryAfterHash  string     `json:"memory_after_hash"`   // SHA-256 of agent context after
	CostUSD          float64    `json:"cost_usd" gorm:"type:decimal(10,6);not null;default:0"`
	LatencyMs        int        `json:"latency_ms" gorm:"not null"`
	Outcome          string     `json:"outcome" gorm:"not null"` // success | error | timeout | policy_violation
	ErrorCode        *string    `json:"error_code,omitempty"`
	PolicyViolation  *string    `json:"policy_violation,omitempty"`
	ObjectKey        string     `json:"object_key,omitempty"` // pointer to full record in object storage
	Timestamp        time.Time  `json:"timestamp" gorm:"not null;default:now()"`
	RetentionDays    int        `json:"retention_days" gorm:"-"` // not stored in DB, computed from plan
}

// TableName returns the GORM table name
func (AgentExecutionRecord) TableName() string {
	return "agent_execution_records"
}

// AgentSession tracks a multi-step agent session
type AgentSession struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SessionID    string     `json:"session_id" gorm:"uniqueIndex;not null"`
	AgentID      string     `json:"agent_id" gorm:"not null"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Status       string     `json:"status" gorm:"not null;default:'active'"` // active | completed | terminated
	CallCount    int        `json:"call_count" gorm:"not null;default:0"`
	TotalCostUSD float64    `json:"total_cost_usd" gorm:"type:decimal(10,6);not null;default:0"`
	CallGraph    []byte     `json:"call_graph,omitempty" gorm:"type:jsonb"`
	StartedAt    time.Time  `json:"started_at" gorm:"not null;default:now()"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	ObjectKey    string     `json:"object_key,omitempty"`
}

// TableName returns the GORM table name
func (AgentSession) TableName() string {
	return "agent_sessions"
}

// SessionStatus constants
const (
	SessionStatusActive     = "active"
	SessionStatusCompleted  = "completed"
	SessionStatusTerminated = "terminated"
)

// ExecutionOutcome constants
const (
	OutcomeSuccess         = "success"
	OutcomeError           = "error"
	OutcomeTimeout         = "timeout"
	OutcomePolicyViolation = "policy_violation"
)

// Repository handles persistence for execution records and sessions
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new attribution repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// RecordExecution writes an execution record to Postgres (hot metadata only)
func (r *Repository) RecordExecution(ctx context.Context, record *AgentExecutionRecord) error {
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	return r.db.WithContext(ctx).Create(record).Error
}

// GetExecution retrieves a single execution record by ID
func (r *Repository) GetExecution(ctx context.Context, executionID string) (*AgentExecutionRecord, error) {
	var record AgentExecutionRecord
	err := r.db.WithContext(ctx).Where("execution_id = ?", executionID).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution record not found: %s", executionID)
		}
		return nil, fmt.Errorf("failed to get execution record: %w", err)
	}
	return &record, nil
}

// ListExecutions lists execution records for an agent with pagination
func (r *Repository) ListExecutions(ctx context.Context, agentID string, limit, offset int) ([]*AgentExecutionRecord, int64, error) {
	var total int64
	var records []*AgentExecutionRecord

	query := r.db.WithContext(ctx).Model(&AgentExecutionRecord{}).Where("agent_id = ?", agentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count execution records: %w", err)
	}

	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list execution records: %w", err)
	}

	return records, total, nil
}

// StartSession creates a new agent session
func (r *Repository) StartSession(ctx context.Context, agentID string, tenantID uuid.UUID, sessionID string) (*AgentSession, error) {
	session := &AgentSession{
		ID:        uuid.New(),
		SessionID: sessionID,
		AgentID:   agentID,
		TenantID:  tenantID,
		Status:    SessionStatusActive,
		StartedAt: time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

// EndSession marks a session as completed or terminated
func (r *Repository) EndSession(ctx context.Context, sessionID string, status string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&AgentSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":   status,
			"ended_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to end session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	return nil
}

// GetSession retrieves a session by session_id
func (r *Repository) GetSession(ctx context.Context, sessionID string) (*AgentSession, error) {
	var session AgentSession
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// IncrementSessionCost atomically adds cost to a session's total
func (r *Repository) IncrementSessionCost(ctx context.Context, sessionID string, costUSD float64) error {
	return r.db.WithContext(ctx).Model(&AgentSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"call_count":     gorm.Expr("call_count + 1"),
			"total_cost_usd": gorm.Expr("total_cost_usd + ?", costUSD),
		}).Error
}

// GetAnalytics returns aggregated analytics for an agent
func (r *Repository) GetAnalytics(ctx context.Context, agentID string, since time.Time) (*AgentAnalytics, error) {
	var analytics AgentAnalytics
	analytics.AgentID = agentID

	err := r.db.WithContext(ctx).Model(&AgentExecutionRecord{}).
		Where("agent_id = ? AND timestamp >= ?", agentID, since).
		Select(`
			COUNT(*) as total_calls,
			SUM(cost_usd) as total_cost_usd,
			AVG(latency_ms) as avg_latency_ms,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY latency_ms) as p50_latency_ms,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms) as p95_latency_ms,
			SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN outcome = 'error' THEN 1 ELSE 0 END) as error_count,
			SUM(CASE WHEN outcome = 'timeout' THEN 1 ELSE 0 END) as timeout_count,
			SUM(CASE WHEN outcome = 'policy_violation' THEN 1 ELSE 0 END) as policy_violation_count
		`).
		Scan(&analytics).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	if analytics.TotalCalls > 0 {
		analytics.SuccessRate = float64(analytics.SuccessCount) / float64(analytics.TotalCalls)
	}

	return &analytics, nil
}

// AgentAnalytics holds aggregated analytics for an agent
type AgentAnalytics struct {
	AgentID              string  `json:"agent_id"`
	TotalCalls           int64   `json:"total_calls"`
	TotalCostUSD         float64 `json:"total_cost_usd"`
	AvgLatencyMs         float64 `json:"avg_latency_ms"`
	P50LatencyMs         float64 `json:"p50_latency_ms"`
	P95LatencyMs         float64 `json:"p95_latency_ms"`
	SuccessCount         int64   `json:"success_count"`
	ErrorCount           int64   `json:"error_count"`
	TimeoutCount         int64   `json:"timeout_count"`
	PolicyViolationCount int64   `json:"policy_violation_count"`
	SuccessRate          float64 `json:"success_rate"`
}

// HashInput computes a SHA-256 hash of the input for attribution
func HashInput(input json.RawMessage) string {
	h := sha256.Sum256(input)
	return hex.EncodeToString(h[:])
}

// HashOutput computes a SHA-256 hash of the output for attribution
func HashOutput(output json.RawMessage) string {
	h := sha256.Sum256(output)
	return hex.EncodeToString(h[:])
}
