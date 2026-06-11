package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentFunctionRepository handles database operations for agent functions
type AgentFunctionRepository struct {
	db *gorm.DB
}

// NewAgentFunctionRepository creates a new agent function repository
func NewAgentFunctionRepository(db *gorm.DB) *AgentFunctionRepository {
	return &AgentFunctionRepository{db: db}
}

// Create creates a new agent function
func (r *AgentFunctionRepository) Create(ctx context.Context, af *AgentFunction) error {
	return r.db.WithContext(ctx).Create(af).Error
}

// GetByID retrieves an agent function by ID
func (r *AgentFunctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*AgentFunction, error) {
	var af AgentFunction
	err := r.db.WithContext(ctx).
		Preload("Function").
		First(&af, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &af, nil
}

// GetByFunctionID retrieves an agent function by function ID
func (r *AgentFunctionRepository) GetByFunctionID(ctx context.Context, functionID uuid.UUID) (*AgentFunction, error) {
	var af AgentFunction
	err := r.db.WithContext(ctx).
		Preload("Function").
		First(&af, "function_id = ?", functionID).Error
	if err != nil {
		return nil, err
	}
	return &af, nil
}

// Update updates an existing agent function
func (r *AgentFunctionRepository) Update(ctx context.Context, af *AgentFunction) error {
	af.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(af).Error
}

// Delete deletes an agent function
func (r *AgentFunctionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AgentFunction{}, "id = ?", id).Error
}

// ListByCategory lists agent functions by category
func (r *AgentFunctionRepository) ListByCategory(ctx context.Context, category AgentFunctionCategory, limit, offset int) ([]AgentFunction, int64, error) {
	var functions []AgentFunction
	var total int64

	query := r.db.WithContext(ctx).
		Model(&AgentFunction{}).
		Where("category = ?", category)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Function").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&functions).Error
	if err != nil {
		return nil, 0, err
	}

	return functions, total, nil
}

// ListExclusive lists all exclusive agent functions
func (r *AgentFunctionRepository) ListExclusive(ctx context.Context, limit, offset int) ([]AgentFunction, int64, error) {
	var functions []AgentFunction
	var total int64

	query := r.db.WithContext(ctx).
		Model(&AgentFunction{}).
		Where("is_exclusive = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Function").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&functions).Error
	if err != nil {
		return nil, 0, err
	}

	return functions, total, nil
}

// ListVerified lists all verified agent functions
func (r *AgentFunctionRepository) ListVerified(ctx context.Context, limit, offset int) ([]AgentFunction, int64, error) {
	var functions []AgentFunction
	var total int64

	query := r.db.WithContext(ctx).
		Model(&AgentFunction{}).
		Where("is_verified = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Function").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&functions).Error
	if err != nil {
		return nil, 0, err
	}

	return functions, total, nil
}

// ListByCapabilities lists agent functions matching required capabilities
func (r *AgentFunctionRepository) ListByCapabilities(ctx context.Context, capabilities []string, limit, offset int) ([]AgentFunction, int64, error) {
	var functions []AgentFunction
	var total int64

	query := r.db.WithContext(ctx).
		Model(&AgentFunction{}).
		Where("capabilities @> ?", capabilities)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Function").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&functions).Error
	if err != nil {
		return nil, 0, err
	}

	return functions, total, nil
}

// ListAll lists all agent functions with optional filters
func (r *AgentFunctionRepository) ListAll(ctx context.Context, category *string, exclusive *bool, verified *bool, limit, offset int) ([]AgentFunction, int64, error) {
	var functions []AgentFunction
	var total int64

	query := r.db.WithContext(ctx).Model(&AgentFunction{})

	if category != nil && *category != "" {
		query = query.Where("category = ?", *category)
	}
	if exclusive != nil {
		query = query.Where("is_exclusive = ?", *exclusive)
	}
	if verified != nil {
		query = query.Where("is_verified = ?", *verified)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Function").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&functions).Error
	if err != nil {
		return nil, 0, err
	}

	return functions, total, nil
}

// GetFunctionPolicy retrieves the policy for an agent-function pair
func (r *AgentFunctionRepository) GetFunctionPolicy(ctx context.Context, agentID, functionID uuid.UUID) (*AgentFunctionPolicy, error) {
	var policy AgentFunctionPolicy
	err := r.db.WithContext(ctx).
		First(&policy, "agent_id = ? AND function_id = ?", agentID, functionID).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// UpsertFunctionPolicy creates or updates a function policy
func (r *AgentFunctionRepository) UpsertFunctionPolicy(ctx context.Context, policy *AgentFunctionPolicy) error {
	policy.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(policy).Error
}

// RecordExecution records a function execution
func (r *AgentFunctionRepository) RecordExecution(ctx context.Context, execution *AgentFunctionExecution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

// GetExecution retrieves a specific execution
func (r *AgentFunctionRepository) GetExecution(ctx context.Context, id uuid.UUID) (*AgentFunctionExecution, error) {
	var execution AgentFunctionExecution
	err := r.db.WithContext(ctx).
		Preload("Function").
		First(&execution, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// ListExecutions lists executions with filters
func (r *AgentFunctionRepository) ListExecutions(ctx context.Context, filter *AgentFunctionExecutionsFilter) ([]AgentFunctionExecution, int64, error) {
	var executions []AgentFunctionExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&AgentFunctionExecution{})

	if filter.AgentID != nil {
		query = query.Where("agent_id = ?", *filter.AgentID)
	}
	if filter.FunctionID != nil {
		query = query.Where("function_id = ?", *filter.FunctionID)
	}
	if filter.SessionID != nil {
		query = query.Where("session_id = ?", *filter.SessionID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := 50
	offset := 0
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	err := query.
		Preload("Function").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions).Error
	if err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

// GetExecutionStats returns execution statistics for an agent
func (r *AgentFunctionRepository) GetExecutionStats(ctx context.Context, agentID uuid.UUID, since time.Time) (*AgentExecutionStats, error) {
	var stats AgentExecutionStats

	// Total executions
	err := r.db.WithContext(ctx).
		Model(&AgentFunctionExecution{}).
		Where("agent_id = ? AND created_at >= ?", agentID, since).
		Count(&stats.TotalExecutions).Error
	if err != nil {
		return nil, err
	}

	// Successful executions
	err = r.db.WithContext(ctx).
		Model(&AgentFunctionExecution{}).
		Where("agent_id = ? AND created_at >= ? AND error IS NULL", agentID, since).
		Count(&stats.SuccessfulExecutions).Error
	if err != nil {
		return nil, err
	}

	// Failed executions
	stats.FailedExecutions = stats.TotalExecutions - stats.SuccessfulExecutions

	// Total cost
	var totalCost float64
	err = r.db.WithContext(ctx).
		Model(&AgentFunctionExecution{}).
		Select("COALESCE(SUM(cost_usd), 0)").
		Where("agent_id = ? AND created_at >= ?", agentID, since).
		Scan(&totalCost).Error
	if err != nil {
		return nil, err
	}
	stats.TotalCostUSD = totalCost

	// Average duration
	var avgDuration float64
	err = r.db.WithContext(ctx).
		Model(&AgentFunctionExecution{}).
		Select("COALESCE(AVG(duration_ms), 0)").
		Where("agent_id = ? AND created_at >= ?", agentID, since).
		Scan(&avgDuration).Error
	if err != nil {
		return nil, err
	}
	stats.AvgDurationMs = avgDuration

	return &stats, nil
}

// AgentExecutionStats holds execution statistics
type AgentExecutionStats struct {
	TotalExecutions     int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions   int64   `json:"failed_executions"`
	TotalCostUSD        float64 `json:"total_cost_usd"`
	AvgDurationMs       float64 `json:"avg_duration_ms"`
}

// ToDefinition converts an AgentFunction to API response format
func (af *AgentFunction) ToDefinition() *AgentFunctionDefinition {
	if af.Function == nil {
		return nil
	}

	var pricing PricingModel
	if af.PricingModel != nil {
		json.Unmarshal(af.PricingModel, &pricing)
	}

	version := ""
	if af.Function.LatestVersion.Valid {
		version = af.Function.LatestVersion.String
	}

	description := ""
	if af.Function.Description.Valid {
		description = af.Function.Description.String
	}

	return &AgentFunctionDefinition{
		Author:        af.Function.Author,
		Name:          af.Function.Name,
		Version:       version,
		Description:   description,
		Category:      string(af.Category),
		InputSchema:   af.InputSchema,
		OutputSchema:  af.OutputSchema,
		Pricing:       pricing,
		Capabilities:  af.Capabilities,
		IsVerified:    af.IsVerified,
		IsExclusive:   af.IsExclusive,
		MaxConcurrency: af.MaxConcurrency,
		RateLimitRPM:  af.RateLimitRPM,
	}
}