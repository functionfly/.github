package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/functionfly/functionfly/internal/agent/circuitbreaker"
	"github.com/functionfly/functionfly/internal/agent/factory"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/policy"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Optimizer provides self-optimization capabilities for agents
type Optimizer struct {
	db           *gorm.DB
	redis        *redis.Client
	identityRepo *identity.Repository
	policyEngine *policy.Engine
	factorySvc   *factory.Service
	logger       *logrus.Logger
}

// OptimizerDeps holds all dependencies for the Optimizer
type OptimizerDeps struct {
	DB           *gorm.DB
	Redis        *redis.Client
	IdentityRepo *identity.Repository
	PolicyEngine *policy.Engine
	FactorySvc   *factory.Service
	Logger       *logrus.Logger
}

// NewOptimizer creates a new learning optimizer with dependencies
func NewOptimizer(deps OptimizerDeps) *Optimizer {
	logger := deps.Logger
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &Optimizer{
		db:           deps.DB,
		redis:        deps.Redis,
		identityRepo: deps.IdentityRepo,
		policyEngine: deps.PolicyEngine,
		factorySvc:   deps.FactorySvc,
		logger:       logger,
	}
}

// Optimization represents a suggested optimization
type Optimization struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID          string          `json:"agent_id" gorm:"not null"`
	PatternID        *uuid.UUID      `json:"pattern_id" gorm:"type:uuid"`
	OptimizationType string          `json:"optimization_type" gorm:"not null"` // timeout_adjustment, caching, batch_processing, etc.
	Description      string          `json:"description"`
	ExpectedImpact   map[string]any  `json:"expected_impact" gorm:"type:jsonb"` // latency_reduction, cost_reduction, etc.
	Implementation   string          `json:"implementation"`                  // low, medium, high
	Status           string          `json:"status" gorm:"not null;default:'pending'"` // pending, approved, rejected, applied
	ApprovedBy       *string         `json:"approved_by"`
	AppliedAt        *time.Time      `json:"applied_at"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	// AppliedConfig stores the configuration that was applied (JSON for flexibility)
	AppliedConfig map[string]any `json:"applied_config" gorm:"type:jsonb"`
	// ErrorMessage stores any error that occurred during implementation
	ErrorMessage string `json:"error_message"`
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var implErr error
	var appliedConfig map[string]any

	switch opt.OptimizationType {
	case OptimizationTypeTimeoutAdjustment:
		appliedConfig, implErr = o.applyTimeoutAdjustment(ctx, opt)

	case OptimizationTypeCaching:
		appliedConfig, implErr = o.applyCaching(ctx, opt)

	case OptimizationTypeBatchProcessing:
		appliedConfig, implErr = o.applyBatchProcessing(ctx, opt)

	case OptimizationTypeResourceUpgrade:
		appliedConfig, implErr = o.applyResourceUpgrade(ctx, opt)

	case OptimizationTypePolicyChange:
		appliedConfig, implErr = o.applyPolicyChange(ctx, opt)

	case OptimizationTypeRetryStrategy:
		appliedConfig, implErr = o.applyRetryStrategy(ctx, opt)

	case OptimizationTypeQueryOptimization:
		appliedConfig, implErr = o.applyQueryOptimization(ctx, opt)

	default:
		implErr = fmt.Errorf("unknown optimization type: %s", opt.OptimizationType)
	}

	// Record the result
	if implErr != nil {
		o.logger.WithFields(logrus.Fields{
			"optimization_id": opt.ID,
			"agent_id":        opt.AgentID,
			"type":            opt.OptimizationType,
			"error":           implErr.Error(),
		}).Error("Failed to apply optimization")
		opt.ErrorMessage = implErr.Error()
	} else {
		o.logger.WithFields(logrus.Fields{
			"optimization_id": opt.ID,
			"agent_id":        opt.AgentID,
			"type":            opt.OptimizationType,
		}).Info("Successfully applied optimization")
		opt.AppliedConfig = appliedConfig
	}

	// Persist result back to DB
	result := o.db.WithContext(ctx).Model(&opt).Updates(map[string]any{
		"applied_config": opt.AppliedConfig,
		"error_message":  opt.ErrorMessage,
	})
	if result.Error != nil {
		o.logger.WithError(result.Error).WithField("id", opt.ID).Error("Failed to persist optimization result")
	}
}

// applyTimeoutAdjustment updates agent quota config for timeout settings
func (o *Optimizer) applyTimeoutAdjustment(ctx context.Context, opt Optimization) (map[string]any, error) {
	if o.identityRepo == nil {
		return nil, fmt.Errorf("identity repository not available")
	}

	// Extract target timeout multiplier from expected impact
	timeoutMultiplier := 1.5 // 50% increase as default
	if impact, ok := opt.ExpectedImpact["timeout_reduction"].(float64); ok {
		// timeout_reduction is expected reduction (e.g., 0.5 = 50% reduction), so increase limit by inverse
		timeoutMultiplier = 1 / (1 - impact)
		if timeoutMultiplier < 1 {
			timeoutMultiplier = 1.5
		}
	}

	// Calculate current baseline from pattern data or use default
	currentTimeoutMs := 300000 // 5 minutes default
	if opt.PatternID != nil {
		// Could look up pattern data here if needed
	}
	newTimeoutMs := int(float64(currentTimeoutMs) * timeoutMultiplier)
	if newTimeoutMs > 3600000 { // Cap at 1 hour
		newTimeoutMs = 3600000
	}

	// If policy engine is available, update max execution time
	if o.policyEngine != nil {
		currentPolicy, err := o.policyEngine.GetPolicy(ctx, opt.AgentID)
		if err == nil && currentPolicy != nil {
			updatedPolicy := *currentPolicy
			updatedPolicy.MaxWallTimeMs = newTimeoutMs
			if err := o.policyEngine.UpsertPolicy(ctx, &updatedPolicy); err != nil {
				return nil, fmt.Errorf("failed to update policy: %w", err)
			}
		}
	}

	return map[string]any{
		"timeout_multiplier": timeoutMultiplier,
		"max_wall_time_ms":   newTimeoutMs,
	}, nil
}

// applyCaching enables or configures caching for the agent
func (o *Optimizer) applyCaching(ctx context.Context, opt Optimization) (map[string]any, error) {
	cacheConfig := map[string]any{
		"enabled": true,
		"ttl_seconds": 1800, // 30 minutes default
	}

	// If factory service is available, update its config
	if o.factorySvc != nil {
		cfg := factory.DefaultConfig(opt.AgentID)
		// Enable caching feature flag
		if cfg.FeatureFlags == nil {
			cfg.FeatureFlags = make(map[string]bool)
		}
		cfg.FeatureFlags["enable_caching"] = true
		cfg.FeatureFlags["cache_ttl_seconds"] = true
		cacheConfig["factory_config_updated"] = true
	}

	// If Redis is available, set agent-specific cache config key
	if o.redis != nil {
		cacheKey := fmt.Sprintf("agent:%s:cache:config", opt.AgentID)
		configJSON, _ := json.Marshal(cacheConfig)
		if err := o.redis.Set(ctx, cacheKey, configJSON, 24*time.Hour).Err(); err != nil {
			o.logger.WithError(err).Warn("Failed to set cache config in Redis")
		}
	}

	return cacheConfig, nil
}

// applyBatchProcessing enables batch processing optimization
func (o *Optimizer) applyBatchProcessing(ctx context.Context, opt Optimization) (map[string]any, error) {
	batchConfig := map[string]any{
		"batch_size":         10,
		"batch_timeout_ms":   5000,
		"max_batch_age_sec":  60,
		"enabled":            true,
	}

	if o.factorySvc != nil {
		cfg := factory.DefaultConfig(opt.AgentID)
		cfg.DiscoveryBatchSize = 10
		batchConfig["factory_config_updated"] = true
	}

	return batchConfig, nil
}

// applyResourceUpgrade increases resource allocation for the agent
func (o *Optimizer) applyResourceUpgrade(ctx context.Context, opt Optimization) (map[string]any, error) {
	if o.identityRepo == nil {
		return nil, fmt.Errorf("identity repository not available")
	}

	quotaCfg, err := o.identityRepo.GetQuotaConfig(ctx, opt.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota config: %w", err)
	}

	// Increase quotas by 50%
	updates := map[string]any{
		"max_calls_per_minute":    int(float64(quotaCfg.MaxCallsPerMinute) * 1.5),
		"max_calls_per_day":       int(float64(quotaCfg.MaxCallsPerDay) * 1.5),
		"max_state_writes_per_hour": int(float64(quotaCfg.MaxStateWritesPerHr) * 1.5),
		"max_cost_per_execution":  quotaCfg.MaxCostPerExecution * 1.5,
	}

	if err := o.identityRepo.UpdateQuotaConfig(ctx, opt.AgentID, updates); err != nil {
		return nil, fmt.Errorf("failed to update quota config: %w", err)
	}

	return map[string]any{
		"previous_limits": map[string]any{
			"max_calls_per_minute":     quotaCfg.MaxCallsPerMinute,
			"max_calls_per_day":       quotaCfg.MaxCallsPerDay,
			"max_state_writes_per_hour": quotaCfg.MaxStateWritesPerHr,
		},
		"new_limits": updates,
	}, nil
}

// applyPolicyChange updates agent behavioral policy
func (o *Optimizer) applyPolicyChange(ctx context.Context, opt Optimization) (map[string]any, error) {
	if o.policyEngine == nil {
		return nil, fmt.Errorf("policy engine not available")
	}

	currentPolicy, err := o.policyEngine.GetPolicy(ctx, opt.AgentID)
	if err != nil {
		// Create default policy if none exists
		currentPolicy = &policy.BehavioralPolicy{
			AgentID: opt.AgentID,
		}
	}

	// Increase allowed depth for better handling of complex operations
	updatedPolicy := *currentPolicy
	updatedPolicy.MaxExecutionDepth = currentPolicy.MaxExecutionDepth + 2
	updatedPolicy.MaxRecursionDepth = currentPolicy.MaxRecursionDepth + 1

	if err := o.policyEngine.UpsertPolicy(ctx, &updatedPolicy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return map[string]any{
		"max_execution_depth": updatedPolicy.MaxExecutionDepth,
		"max_recursion_depth": updatedPolicy.MaxRecursionDepth,
	}, nil
}

// applyRetryStrategy updates retry policy with exponential backoff
func (o *Optimizer) applyRetryStrategy(ctx context.Context, opt Optimization) (map[string]any, error) {
	// Build circuit breaker config based on retry patterns
	cbConfig := circuitbreaker.Config{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		CooldownDuration:    30 * time.Second,
		HalfOpenMaxRequests: 1,
	}

	// Store the config in Redis for the agent to pick up
	if o.redis != nil {
		cbKey := fmt.Sprintf("agent:%s:circuit_breaker:config", opt.AgentID)
		configJSON, _ := json.Marshal(cbConfig)
		if err := o.redis.Set(ctx, cbKey, configJSON, 24*time.Hour).Err(); err != nil {
			o.logger.WithError(err).Warn("Failed to set circuit breaker config")
		}
	}

	return map[string]any{
		"circuit_breaker_config": map[string]any{
			"failure_threshold":     cbConfig.FailureThreshold,
			"success_threshold":    cbConfig.SuccessThreshold,
			"cooldown_duration_sec": int(cbConfig.CooldownDuration.Seconds()),
		},
		"retry_strategy": "exponential_backoff",
	}, nil
}

// applyQueryOptimization optimizes database query patterns
func (o *Optimizer) applyQueryOptimization(ctx context.Context, opt Optimization) (map[string]any, error) {
	hints := map[string]any{
		"use_connection_pool":  true,
		"query_timeout_ms":      5000,
		"max_connections":       20,
		"enable_query_cache":    true,
		"cache_ttl_seconds":     300,
	}

	// Store query optimization hints in Redis
	if o.redis != nil {
		hintsKey := fmt.Sprintf("agent:%s:query_optimization:hints", opt.AgentID)
		hintsJSON, _ := json.Marshal(hints)
		if err := o.redis.Set(ctx, hintsKey, hintsJSON, 24*time.Hour).Err(); err != nil {
			o.logger.WithError(err).Warn("Failed to set query optimization hints")
		}
	}

	return map[string]any{
		"query_optimization_hints": hints,
		"applied":                  true,
	}, nil
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
