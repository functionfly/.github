package learning

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/attribution"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Analyzer analyzes execution patterns and provides insights
type Analyzer struct {
	db *gorm.DB
}

// NewAnalyzer creates a new learning analyzer
func NewAnalyzer(db *gorm.DB) *Analyzer {
	return &Analyzer{db: db}
}

// ExecutionPattern represents a recognized pattern in agent executions
type ExecutionPattern struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID       string                 `json:"agent_id" gorm:"not null"`
	PatternType   string                 `json:"pattern_type" gorm:"not null"` // frequent_failure, slow_execution, cost_inefficient, etc.
	PatternData   map[string]any         `json:"pattern_data" gorm:"type:jsonb"`
	Confidence    float64                `json:"confidence" gorm:"type:decimal(5,2)"`
	OccurrenceCnt int                    `json:"occurrence_count" gorm:"not null;default:1"`
	FirstSeenAt   time.Time              `json:"first_seen_at" gorm:"autoCreateTime"`
	LastSeenAt    time.Time              `json:"last_seen_at" gorm:"autoUpdateTime"`
	Recommendations []string            `json:"recommendations" gorm:"type:text[]"`
	IsActive      bool                   `json:"is_active" gorm:"not null;default:true"`
}

// TableName returns the GORM table name
func (ExecutionPattern) TableName() string {
	return "agent_execution_patterns"
}

// Pattern types
const (
	PatternTypeFrequentFailure   = "frequent_failure"
	PatternTypeSlowExecution     = "slow_execution"
	PatternTypeCostInefficient   = "cost_inefficient"
	PatternTypeHighRetryRate     = "high_retry_rate"
	PatternTypeResourceContention = "resource_contention"
	PatternTypeSuccessful         = "successful"
	PatternTypeOptimal           = "optimal"
)

// AnalyzePatterns analyzes execution patterns for an agent
func (a *Analyzer) AnalyzePatterns(ctx context.Context, agentID string, timeWindow time.Duration) (*AnalysisResult, error) {
	since := time.Now().Add(-timeWindow)

	// Get execution records
	var records []attribution.AgentExecutionRecord
	err := a.db.WithContext(ctx).
		Where("agent_id = ? AND timestamp >= ?", agentID, since).
		Order("timestamp DESC").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get execution records: %w", err)
	}

	result := &AnalysisResult{
		AgentID:        agentID,
		AnalysisWindow: timeWindow,
		TotalExecutions: len(records),
		Patterns:       []ExecutionPattern{},
		Insights:       []string{},
	}

	if len(records) == 0 {
		result.Insights = append(result.Insights, "No execution data available for analysis")
		return result, nil
	}

	// Analyze failure patterns
	failurePattern := a.analyzeFailurePatterns(records)
	if failurePattern != nil {
		result.Patterns = append(result.Patterns, *failurePattern)
	}

	// Analyze latency patterns
	latencyPattern := a.analyzeLatencyPatterns(records)
	if latencyPattern != nil {
		result.Patterns = append(result.Patterns, *latencyPattern)
	}

	// Analyze cost patterns
	costPattern := a.analyzeCostPatterns(records)
	if costPattern != nil {
		result.Patterns = append(result.Patterns, *costPattern)
	}

	// Analyze retry patterns
	retryPattern := a.analyzeRetryPatterns(records)
	if retryPattern != nil {
		result.Patterns = append(result.Patterns, *retryPattern)
	}

	// Generate insights
	result.Insights = a.generateInsights(result.Patterns, records)

	return result, nil
}

// AnalysisResult contains the results of pattern analysis
type AnalysisResult struct {
	AgentID          string              `json:"agent_id"`
	AnalysisWindow   time.Duration       `json:"analysis_window"`
	TotalExecutions  int                 `json:"total_executions"`
	Patterns         []ExecutionPattern  `json:"patterns"`
	Insights         []string             `json:"insights"`
	SuccessRate      float64             `json:"success_rate"`
	AvgLatencyMs     float64             `json:"avg_latency_ms"`
	AvgCostUSD       float64             `json:"avg_cost_usd"`
}

func (a *Analyzer) analyzeFailurePatterns(records []attribution.AgentExecutionRecord) *ExecutionPattern {
	var failureCount int
	failureCategories := map[string]int{}

	for _, r := range records {
		if r.Outcome == "error" || r.Outcome == "timeout" || r.Outcome == "policy_violation" {
			failureCount++
			if r.ErrorCode != nil {
				failureCategories[*r.ErrorCode]++
			} else {
				failureCategories["unknown"]++
			}
		}
	}

	if failureCount == 0 {
		return nil
	}

	failureRate := float64(failureCount) / float64(len(records))
	if failureRate < 0.1 {
		return nil // Not significant
	}

	// Find most common failure category
	var maxCategory string
	maxCount := 0
	for cat, count := range failureCategories {
		if count > maxCount {
			maxCount = count
			maxCategory = cat
		}
	}

	pattern := ExecutionPattern{
		ID:            uuid.New(),
		AgentID:       records[0].AgentID,
		PatternType:   PatternTypeFrequentFailure,
		PatternData:   map[string]any{"failure_rate": failureRate, "categories": failureCategories},
		Confidence:    failureRate * 100,
		OccurrenceCnt: failureCount,
		FirstSeenAt:   time.Now(),
		LastSeenAt:    time.Now(),
	}

	if maxCategory != "" {
		pattern.Recommendations = []string{
			fmt.Sprintf("Address %s failures (%d occurrences)", maxCategory, maxCount),
			"Consider adding error handling logic",
			"Review function implementation for edge cases",
		}
	}

	return &pattern
}

func (a *Analyzer) analyzeLatencyPatterns(records []attribution.AgentExecutionRecord) *ExecutionPattern {
	var totalLatency int64
	var slowExecutions int

	for _, r := range records {
		totalLatency += int64(r.LatencyMs)
		if r.LatencyMs > 10000 {
			slowExecutions++
		}
	}

	avgLatency := float64(totalLatency) / float64(len(records))
	slowRate := float64(slowExecutions) / float64(len(records))

	if avgLatency < 5000 && slowRate < 0.1 {
		return nil // Performance is acceptable
	}

	pattern := ExecutionPattern{
		ID:            uuid.New(),
		AgentID:       records[0].AgentID,
		PatternType:   PatternTypeSlowExecution,
		PatternData:   map[string]any{"avg_latency_ms": avgLatency, "slow_executions": slowExecutions, "slow_rate": slowRate},
		Confidence:    min(100, avgLatency/100),
		OccurrenceCnt: slowExecutions,
		FirstSeenAt:   time.Now(),
		LastSeenAt:    time.Now(),
		Recommendations: []string{
			fmt.Sprintf("Average latency %.0fms is above target", avgLatency),
			"Consider implementing caching",
			"Review database queries for optimization",
			"Consider async processing for long-running tasks",
		},
	}

	return &pattern
}

func (a *Analyzer) analyzeCostPatterns(records []attribution.AgentExecutionRecord) *ExecutionPattern {
	var totalCost float64
	var highCostExecutions int

	for _, r := range records {
		totalCost += r.CostUSD
		if r.CostUSD > 0.10 {
			highCostExecutions++
		}
	}

	avgCost := totalCost / float64(len(records))
	highCostRate := float64(highCostExecutions) / float64(len(records))

	if avgCost < 0.05 && highCostRate < 0.1 {
		return nil // Cost is acceptable
	}

	pattern := ExecutionPattern{
		ID:            uuid.New(),
		AgentID:       records[0].AgentID,
		PatternType:   PatternTypeCostInefficient,
		PatternData:   map[string]any{"avg_cost_usd": avgCost, "high_cost_executions": highCostExecutions, "high_cost_rate": highCostRate},
		Confidence:    min(100, avgCost*1000),
		OccurrenceCnt: highCostExecutions,
		FirstSeenAt:   time.Now(),
		LastSeenAt:    time.Now(),
		Recommendations: []string{
			fmt.Sprintf("Average cost $%.4f is above target", avgCost),
			"Optimize function to reduce computation",
			"Consider caching expensive operations",
			"Review resource usage",
		},
	}

	return &pattern
}

func (a *Analyzer) analyzeRetryPatterns(records []attribution.AgentExecutionRecord) *ExecutionPattern {
	return nil // Retry tracking not available in current implementation
}

func (a *Analyzer) generateInsights(patterns []ExecutionPattern, records []attribution.AgentExecutionRecord) []string {
	insights := []string{}

	// Calculate aggregate metrics
	var successCount int
	var totalLatency int64
	var totalCost float64

	for _, r := range records {
		if r.Outcome == "success" {
			successCount++
		}
		totalLatency += int64(r.LatencyMs)
		totalCost += r.CostUSD
	}

	successRate := float64(successCount) / float64(len(records)) * 100
	avgLatency := float64(totalLatency) / float64(len(records))
	avgCost := totalCost / float64(len(records))

	// Add aggregate insights
	insights = append(insights, fmt.Sprintf("Success rate: %.1f%%", successRate))
	insights = append(insights, fmt.Sprintf("Average latency: %.0fms", avgLatency))
	insights = append(insights, fmt.Sprintf("Average cost: $%.4f", avgCost))

	// Add pattern-specific insights
	for _, p := range patterns {
		if !p.IsActive {
			continue
		}
		switch p.PatternType {
		case PatternTypeFrequentFailure:
			insights = append(insights, fmt.Sprintf("⚠️ High failure rate detected (%.0f%% confidence)", p.Confidence))
		case PatternTypeSlowExecution:
			insights = append(insights, fmt.Sprintf("🐢 Slow execution pattern detected (%.0fms avg)", p.PatternData["avg_latency_ms"]))
		case PatternTypeCostInefficient:
			insights = append(insights, fmt.Sprintf("💰 High cost pattern detected ($%.4f avg)", p.PatternData["avg_cost_usd"]))
		case PatternTypeHighRetryRate:
			insights = append(insights, fmt.Sprintf("🔄 High retry rate detected (%.0f%%)", p.PatternData["retry_rate"].(float64)*100))
		}
	}

	return insights
}

// SavePattern saves or updates an execution pattern
func (a *Analyzer) SavePattern(ctx context.Context, pattern *ExecutionPattern) error {
	// Check if similar pattern exists
	var existing ExecutionPattern
	err := a.db.WithContext(ctx).
		Where("agent_id = ? AND pattern_type = ? AND is_active = ?", pattern.AgentID, pattern.PatternType, true).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new pattern
		return a.db.WithContext(ctx).Create(pattern).Error
	}

	if err != nil {
		return fmt.Errorf("failed to check existing pattern: %w", err)
	}

	// Update existing pattern
	existing.OccurrenceCnt += pattern.OccurrenceCnt
	existing.LastSeenAt = time.Now()
	// Update confidence if new pattern has higher confidence
	if pattern.Confidence > existing.Confidence {
		existing.Confidence = pattern.Confidence
		existing.PatternData = pattern.PatternData
		existing.Recommendations = pattern.Recommendations
	}

	return a.db.WithContext(ctx).Save(&existing).Error
}

// GetActivePatterns retrieves active patterns for an agent
func (a *Analyzer) GetActivePatterns(ctx context.Context, agentID string) ([]ExecutionPattern, error) {
	var patterns []ExecutionPattern
	err := a.db.WithContext(ctx).
		Where("agent_id = ? AND is_active = ?", agentID, true).
		Order("confidence DESC").
		Find(&patterns).Error
	return patterns, err
}

// AutoEnrichMemory automatically stores execution outcomes as memories
func (a *Analyzer) AutoEnrichMemory(ctx context.Context, record *attribution.AgentExecutionRecord) error {
	// Store the execution as a learning memory
	memory := AgentMemory{
		ID:          uuid.New(),
		AgentID:     record.AgentID,
		MemoryType:  MemoryTypeExecution,
		Content:     map[string]any{"execution_id": record.ExecutionID, "outcome": record.Outcome, "latency_ms": record.LatencyMs, "cost_usd": record.CostUSD},
		Importance:  a.calculateMemoryImportance(record),
		IsLearned:   false,
		CreatedAt:   time.Now(),
	}

	return a.db.WithContext(ctx).Create(&memory).Error
}

func (a *Analyzer) calculateMemoryImportance(record *attribution.AgentExecutionRecord) float64 {
	importance := 0.5 // Base importance

	// Failures are more important to remember
	if record.Outcome == "error" || record.Outcome == "timeout" {
		importance += 0.3
	}

	// High latency is important
	if record.LatencyMs > 10000 {
		importance += 0.1
	}

	// High cost is important
	if record.CostUSD > 0.10 {
		importance += 0.1
	}

	return min(1.0, importance)
}

// min returns the minimum of two floats
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
