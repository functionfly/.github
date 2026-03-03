package learning

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Optimizer provides self-optimization capabilities for agents
type Optimizer struct {
	db *gorm.DB
}

// NewOptimizer creates a new learning optimizer
func NewOptimizer(db *gorm.DB) *Optimizer {
	return &Optimizer{db: db}
}

// Optimization represents a suggested optimization
type Optimization struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID         string                 `json:"agent_id" gorm:"not null"`
	PatternID       *uuid.UUID            `json:"pattern_id" gorm:"type:uuid"`
	OptimizationType string                `json:"optimization_type" gorm:"not null"` // timeout_adjustment, caching, batch_processing, etc.
	Description     string                 `json:"description"`
	ExpectedImpact  map[string]any         `json:"expected_impact" gorm:"type:jsonb"` // latency_reduction, cost_reduction, etc.
	Implementation  string                 `json:"implementation"` // low, medium, high
	Status          string                 `json:"status" gorm:"not null;default:'pending'"` // pending, approved, rejected, applied
	ApprovedBy      *string               `json:"approved_by"`
	AppliedAt       *time.Time            `json:"applied_at"`
	CreatedAt       time.Time             `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time             `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (Optimization) TableName() string {
	return "agent_optimizations"
}

// Optimization types
const (
	OptimizationTypeTimeoutAdjustment = "timeout_adjustment"
	OptimizationTypeCaching           = "caching"
	OptimizationTypeBatchProcessing  = "batch_processing"
	OptimizationTypeResourceUpgrade   = "resource_upgrade"
	OptimizationTypePolicyChange      = "policy_change"
	OptimizationTypeRetryStrategy     = "retry_strategy"
	OptimizationTypeQueryOptimization = "query_optimization"
)

// Implementation levels
const (
	ImplementationLow    = "low"
	ImplementationMedium = "medium"
	ImplementationHigh   = "high"
)

// GenerateOptimizations generates optimization suggestions based on patterns
func (o *Optimizer) GenerateOptimizations(ctx context.Context, agentID string, patterns []ExecutionPattern) ([]Optimization, error) {
	optimizations := []Optimization{}

	for _, pattern := range patterns {
		if !pattern.IsActive {
			continue
		}

		var opt Optimization
		switch pattern.PatternType {
		case PatternTypeFrequentFailure:
			opt = o.generateFailureOptimization(agentID, pattern)
		case PatternTypeSlowExecution:
			opt = o.generateLatencyOptimization(agentID, pattern)
		case PatternTypeCostInefficient:
			opt = o.generateCostOptimization(agentID, pattern)
		case PatternTypeHighRetryRate:
			opt = o.generateRetryOptimization(agentID, pattern)
		default:
			continue
		}

		if opt.ID == uuid.Nil {
			opt.ID = uuid.New()
		}
		opt.AgentID = agentID
		opt.PatternID = &pattern.ID
		opt.CreatedAt = time.Now()
		opt.UpdatedAt = time.Now()

		optimizations = append(optimizations, opt)

		// Save optimization to database
		if err := o.db.WithContext(ctx).Create(&opt).Error; err != nil {
			return nil, fmt.Errorf("failed to save optimization: %w", err)
		}
	}

	// Also analyze recent executions directly for quick wins
	execOpt, err := o.analyzeQuickWins(ctx, agentID)
	if err == nil && execOpt != nil {
		optimizations = append(optimizations, *execOpt)
	}

	return optimizations, nil
}

func (o *Optimizer) generateFailureOptimization(agentID string, pattern ExecutionPattern) Optimization {
	data := pattern.PatternData
	_ = data // Suppress unused warning

	return Optimization{
		AgentID:           agentID,
		OptimizationType:  OptimizationTypePolicyChange,
		Description:       "Add error handling improvements to reduce failure rate",
		ExpectedImpact:    map[string]any{"failure_reduction": 0.3, "success_rate_increase": 15.0},
		Implementation:    ImplementationMedium,
		Status:            "pending",
	}
}

func (o *Optimizer) generateLatencyOptimization(agentID string, pattern ExecutionPattern) Optimization {
	avgLatency := pattern.PatternData["avg_latency_ms"].(float64)

	optType := OptimizationTypeCaching
	impl := ImplementationMedium
	desc := "Implement caching to reduce latency"

	if avgLatency > 20000 {
		optType = OptimizationTypeResourceUpgrade
		impl = ImplementationHigh
		desc = "Upgrade to higher resource tier for better performance"
	}

	return Optimization{
		AgentID:           agentID,
		OptimizationType:  optType,
		Description:       desc,
		ExpectedImpact:    map[string]any{"latency_reduction": 0.4},
		Implementation:    impl,
		Status:            "pending",
	}
}

func (o *Optimizer) generateCostOptimization(agentID string, pattern ExecutionPattern) Optimization {
	return Optimization{
		AgentID:           agentID,
		OptimizationType:  OptimizationTypeQueryOptimization,
		Description:       "Optimize database queries and reduce computation to lower costs",
		ExpectedImpact:    map[string]any{"cost_reduction": 0.5},
		Implementation:    ImplementationMedium,
		Status:            "pending",
	}
}

func (o *Optimizer) generateRetryOptimization(agentID string, pattern ExecutionPattern) Optimization {
	retryRate := pattern.PatternData["retry_rate"].(float64)

	impl := ImplementationLow
	desc := "Improve retry strategy with exponential backoff"

	if retryRate > 0.3 {
		impl = ImplementationMedium
		desc = "Add circuit breaker pattern to prevent cascade failures"
	}

	return Optimization{
		AgentID:           agentID,
		OptimizationType:  OptimizationTypeRetryStrategy,
		Description:       desc,
		ExpectedImpact:    map[string]any{"retry_reduction": 0.6},
		Implementation:    impl,
		Status:            "pending",
	}
}

func (o *Optimizer) analyzeQuickWins(ctx context.Context, agentID string) (*Optimization, error) {
	since := time.Now().Add(-24 * time.Hour) // Last 24 hours

	var records []attribution.AgentExecutionRecord
	err := o.db.WithContext(ctx).
		Where("agent_id = ? AND timestamp >= ?", agentID, since).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	if len(records) < 10 {
		return nil, nil // Not enough data
	}

	// Check for timeout patterns
	var timeoutCount int
	for _, r := range records {
		if r.Outcome == "timeout" {
			timeoutCount++
		}
	}

	if timeoutCount > len(records)/10 { // More than 10% timeouts
		opt := Optimization{
			AgentID:           agentID,
			OptimizationType:  OptimizationTypeTimeoutAdjustment,
			Description:       "Increase timeout threshold based on recent timeout patterns",
			ExpectedImpact:    map[string]any{"timeout_reduction": 0.5},
			Implementation:    ImplementationLow,
			Status:            "pending",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		if err := o.db.WithContext(ctx).Create(&opt).Error; err != nil {
			return nil, err
		}
		return &opt, nil
	}

	return nil, nil
}

// ApproveOptimization approves an optimization for application
func (o *Optimizer) ApproveOptimization(ctx context.Context, optimizationID uuid.UUID, approverID string) (*Optimization, error) {
	var opt Optimization
	if err := o.db.WithContext(ctx).Where("id = ?", optimizationID).First(&opt).Error; err != nil {
		return nil, fmt.Errorf("optimization not found: %w", err)
	}

	if opt.Status != "pending" {
		return nil, fmt.Errorf("optimization is not pending")
	}

	opt.Status = "approved"
	opt.ApprovedBy = &approverID
	opt.UpdatedAt = time.Now()

	if err := o.db.WithContext(ctx).Save(&opt).Error; err != nil {
		return nil, fmt.Errorf("failed to approve optimization: %w", err)
	}

	return &opt, nil
}

// RejectOptimization rejects an optimization
func (o *Optimizer) RejectOptimization(ctx context.Context, optimizationID uuid.UUID, reason string) error {
	var opt Optimization
	if err := o.db.WithContext(ctx).Where("id = ?", optimizationID).First(&opt).Error; err != nil {
		return fmt.Errorf("optimization not found: %w", err)
	}

	opt.Status = "rejected"
	opt.Description = reason + " | " + opt.Description
	opt.UpdatedAt = time.Now()

	return o.db.WithContext(ctx).Save(&opt).Error
}

// ApplyOptimization marks an optimization as applied
func (o *Optimizer) ApplyOptimization(ctx context.Context, optimizationID uuid.UUID) (*Optimization, error) {
	var opt Optimization
	if err := o.db.WithContext(ctx).Where("id = ?", optimizationID).First(&opt).Error; err != nil {
		return nil, fmt.Errorf("optimization not found: %w", err)
	}

	if opt.Status != "approved" {
		return nil, fmt.Errorf("optimization must be approved before application")
	}

	now := time.Now()
	opt.Status = "applied"
	opt.AppliedAt = &now
	opt.UpdatedAt = time.Now()

	if err := o.db.WithContext(ctx).Save(&opt).Error; err != nil {
		return nil, fmt.Errorf("failed to apply optimization: %w", err)
	}

	// Trigger the actual optimization based on type
	go o.implementOptimization(opt)

	return &opt, nil
}

func (o *Optimizer) implementOptimization(opt Optimization) {
	// This would trigger the actual implementation
	// For now, we just log what would happen
	fmt.Printf("Implementing optimization %s for agent %s: %s\n", opt.ID, opt.AgentID, opt.Description)

	// In a real implementation, this would:
	// - For timeout adjustments: update agent quota config
	// - For caching: configure caching middleware
	// - For policy changes: update agent policy
	// etc.
}

// GetOptimizations retrieves optimizations for an agent
func (o *Optimizer) GetOptimizations(ctx context.Context, agentID string, status string) ([]Optimization, error) {
	query := o.db.WithContext(ctx).Where("agent_id = ?", agentID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var optimizations []Optimization
	err := query.Order("created_at DESC").Find(&optimizations).Error
	return optimizations, err
}

// GetOptimization gets a specific optimization
func (o *Optimizer) GetOptimization(ctx context.Context, optimizationID uuid.UUID) (*Optimization, error) {
	var opt Optimization
	err := o.db.WithContext(ctx).Where("id = ?", optimizationID).First(&opt).Error
	return &opt, err
}

// AutoOptimize automatically runs optimization analysis and generates recommendations
func (o *Optimizer) AutoOptimize(ctx context.Context, agentID string) (*OptimizationResult, error) {
	analyzer := NewAnalyzer(o.db)

	// Get active patterns
	patterns, err := analyzer.GetActivePatterns(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get patterns: %w", err)
	}

	if len(patterns) == 0 {
		// Analyze fresh patterns
		analysis, err := analyzer.AnalyzePatterns(ctx, agentID, 7*24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze patterns: %w", err)
		}

		// Save new patterns
		for i := range analysis.Patterns {
			if err := analyzer.SavePattern(ctx, &analysis.Patterns[i]); err != nil {
				return nil, fmt.Errorf("failed to save pattern: %w", err)
			}
			patterns = append(patterns, analysis.Patterns[i])
		}
	}

	// Generate optimizations
	optimizations, err := o.GenerateOptimizations(ctx, agentID, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimizations: %w", err)
	}

	result := &OptimizationResult{
		AgentID:        agentID,
		PatternsFound:  len(patterns),
		Optimizations:  optimizations,
		RecommendedCnt: len(optimizations),
	}

	return result, nil
}

// OptimizationResult contains the results of auto-optimization
type OptimizationResult struct {
	AgentID         string         `json:"agent_id"`
	PatternsFound   int            `json:"patterns_found"`
	Optimizations   []Optimization `json:"optimizations"`
	RecommendedCnt  int            `json:"recommended_count"`
}
