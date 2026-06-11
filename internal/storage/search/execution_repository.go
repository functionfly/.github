package search

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExecutionRepository handles database operations for search executions
type ExecutionRepository struct {
	db *gorm.DB
}

// NewExecutionRepository creates a new execution repository
func NewExecutionRepository(db *gorm.DB) *ExecutionRepository {
	return &ExecutionRepository{db: db}
}

// Create records a new execution
func (r *ExecutionRepository) Create(ctx context.Context, execution *Execution) error {
	if execution.ID == uuid.Nil {
		execution.ID = uuid.New()
	}
	if execution.CreatedAt.IsZero() {
		execution.CreatedAt = time.Now()
	}

	return r.db.WithContext(ctx).Create(execution).Error
}

// GetByID retrieves an execution by ID
func (r *ExecutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Execution, error) {
	var execution Execution
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&execution).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution not found: %s", id)
		}
		return nil, err
	}
	return &execution, nil
}

// ListByAgent retrieves executions for a specific agent with pagination
func (r *ExecutionRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]Execution, error) {
	var executions []Execution

	query := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&executions).Error; err != nil {
		return nil, err
	}

	return executions, nil
}

// ListByTool retrieves executions for a specific tool with pagination
func (r *ExecutionRepository) ListByTool(ctx context.Context, toolName string, limit, offset int) ([]Execution, error) {
	var executions []Execution

	query := r.db.WithContext(ctx).
		Where("tool_name = ?", toolName).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&executions).Error; err != nil {
		return nil, err
	}

	return executions, nil
}

// List retrieves all executions with pagination
func (r *ExecutionRepository) List(ctx context.Context, limit, offset int) ([]Execution, error) {
	var executions []Execution

	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&executions).Error; err != nil {
		return nil, err
	}

	return executions, nil
}

// GetStats retrieves aggregate statistics for a tool
func (r *ExecutionRepository) GetStats(ctx context.Context, toolName string, since time.Time) (*ExecutionStats, error) {
	var stats ExecutionStats

	// Get total executions
	if err := r.db.WithContext(ctx).
		Model(&Execution{}).
		Where("tool_name = ? AND created_at >= ?", toolName, since).
		Count(&stats.TotalExecutions).Error; err != nil {
		return nil, err
	}

	// Get aggregate metrics
	type aggregateResult struct {
		TotalCredits float64
		TotalResults int64
		AvgTimeMs    float64
	}

	var result aggregateResult
	if err := r.db.WithContext(ctx).
		Model(&Execution{}).
		Select(`SUM(credits_used) as total_credits,
			SUM(results_count) as total_results,
			COALESCE(AVG(execution_time_ms), 0) as avg_time_ms`).
		Where("tool_name = ? AND created_at >= ?", toolName, since).
		Scan(&result).Error; err != nil {
		return nil, err
	}

	stats.TotalCredits = result.TotalCredits
	stats.TotalResults = result.TotalResults
	stats.AverageExecutionTimeMs = result.AvgTimeMs

	return &stats, nil
}

// GetStatsByAgent retrieves statistics for a specific agent
func (r *ExecutionRepository) GetStatsByAgent(ctx context.Context, agentID uuid.UUID, since time.Time) (*AgentExecutionStats, error) {
	var stats AgentExecutionStats

	// Get total executions for this agent
	if err := r.db.WithContext(ctx).
		Model(&Execution{}).
		Where("agent_id = ? AND created_at >= ?", agentID, since).
		Count(&stats.TotalExecutions).Error; err != nil {
		return nil, err
	}

	// Get credits used
	type result struct {
		TotalCredits float64
	}

	var res result
	if err := r.db.WithContext(ctx).
		Model(&Execution{}).
		Select("COALESCE(SUM(credits_used), 0) as total_credits").
		Where("agent_id = ? AND created_at >= ?", agentID, since).
		Scan(&res).Error; err != nil {
		return nil, err
	}
	stats.TotalCreditsUsed = res.TotalCredits

	// Get usage by tool
	type toolUsage struct {
		ToolName      string
		ExecutionCount int64
		CreditsUsed   float64
	}

	var toolUsages []toolUsage
	if err := r.db.WithContext(ctx).
		Model(&Execution{}).
		Select("tool_name, COUNT(*) as execution_count, SUM(credits_used) as credits_used").
		Where("agent_id = ? AND created_at >= ?", agentID, since).
		Group("tool_name").
		Scan(&toolUsages).Error; err != nil {
		return nil, err
	}

	stats.UsageByTool = make(map[string]ToolUsage, len(toolUsages))
	for _, tu := range toolUsages {
		stats.UsageByTool[tu.ToolName] = ToolUsage{
			ExecutionCount: tu.ExecutionCount,
			CreditsUsed:    tu.CreditsUsed,
		}
	}

	return &stats, nil
}

// ExecutionStats holds aggregate statistics
type ExecutionStats struct {
	TotalExecutions        int64
	TotalCredits           float64
	TotalResults           int64
	AverageExecutionTimeMs float64
}

// AgentExecutionStats holds agent-specific statistics
type AgentExecutionStats struct {
	TotalExecutions  int64
	TotalCreditsUsed float64
	UsageByTool      map[string]ToolUsage
}

// ToolUsage holds usage statistics for a specific tool
type ToolUsage struct {
	ExecutionCount int64
	CreditsUsed    float64
}