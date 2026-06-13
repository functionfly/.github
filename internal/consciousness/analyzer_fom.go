package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FOMAnalyzer struct {
	db *sql.DB
}

func NewFOMAnalyzer(db *sql.DB) *FOMAnalyzer {
	return &FOMAnalyzer{db: db}
}

func (a *FOMAnalyzer) Name() string {
	return "fom"
}

func (a *FOMAnalyzer) Category() InsightCategory {
	return CategoryFOM
}

const CategoryFOM InsightCategory = "fom"

func (a *FOMAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	insights := []*Insight{}

	efficiency, err := a.analyzeWorkflowEfficiency(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	insights = append(insights, efficiency...)

	failures, err := a.analyzeFailurePatterns(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	insights = append(insights, failures...)

	costs, err := a.analyzeCostOptimization(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	insights = append(insights, costs...)

	workflows, err := a.analyzeWorkflowOptimization(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	insights = append(insights, workflows...)

	return insights, nil
}

func (a *FOMAnalyzer) analyzeWorkflowEfficiency(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		SELECT
			g.goal_type,
			AVG(r.total_time_ms) as avg_time,
			AVG(r.total_cost) as avg_cost,
			AVG(r.outcome_score) as avg_score,
			COUNT(*) as execution_count,
			SUM(CASE WHEN r.success THEN 1 ELSE 0 END)::float / COUNT(*) as success_rate
		FROM fom_results r
		JOIN fom_plans p ON r.plan_id = p.id
		JOIN fom_goals g ON p.goal_id = g.id
		WHERE g.tenant_id = $1
		AND r.created_at > NOW() - INTERVAL '7 days'
		GROUP BY g.goal_type
		HAVING COUNT(*) >= 10
		ORDER BY avg_time DESC
	`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var goalType string
		var avgTime, avgCost, avgScore float64
		var executionCount int
		var successRate float64

		if err := rows.Scan(&goalType, &avgTime, &avgCost, &avgScore, &executionCount, &successRate); err != nil {
			continue
		}

		if avgTime > 30000 && successRate < 0.8 {
			insight := &Insight{
				ID:         uuid.New(),
				TenantID:   tenantID,
				Category:   CategoryFOM,
				Severity:   SeverityWarning,
				Title:      "Slow workflow detected: " + goalType,
				Message:    "Workflow type '" + goalType + "' averages " + formatDuration(int(avgTime)) + " with only " + formatPercent(successRate) + " success rate",
				Confidence: floatPtr(0.85),
				Status:     StatusActive,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				InsightData: JSONMap{
					"goal_type":     goalType,
					"avg_time_ms":   int(avgTime),
					"avg_cost":      avgCost,
					"avg_score":     avgScore,
					"success_rate":  successRate,
					"executions":    executionCount,
				},
				ActionType: ActionOptimize,
				ActionData: JSONMap{
					"suggested_action": "consider_workflow_optimization",
					"goal_type":        goalType,
				},
			}
			insights = append(insights, insight)
		}
	}

	return insights, rows.Err()
}

func (a *FOMAnalyzer) analyzeFailurePatterns(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		SELECT
			r.failure_code,
			COUNT(*) as failure_count,
			COUNT(*)::float / SUM(COUNT(*)) OVER () as failure_ratio
		FROM fom_results r
		JOIN fom_plans p ON r.plan_id = p.id
		JOIN fom_goals g ON p.goal_id = g.id
		WHERE g.tenant_id = $1
		AND r.success = false
		AND r.created_at > NOW() - INTERVAL '7 days'
		GROUP BY r.failure_code
		HAVING COUNT(*) >= 5
		ORDER BY failure_count DESC
	`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var failureCode string
		var failureCount int
		var failureRatio float64

		if err := rows.Scan(&failureCode, &failureCount, &failureRatio); err != nil {
			continue
		}

		if failureRatio > 0.15 {
			insight := &Insight{
				ID:         uuid.New(),
				TenantID:   tenantID,
				Category:   CategoryFOM,
				Severity:   SeverityOpportunity,
				Title:      "Frequent failure pattern: " + failureCode,
				Message:    "Failure code '" + failureCode + "' accounts for " + formatPercent(failureRatio) + " of all failures",
				Confidence: floatPtr(0.80),
				Status:     StatusActive,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				InsightData: JSONMap{
					"failure_code":    failureCode,
					"failure_count":   failureCount,
					"failure_ratio":   failureRatio,
				},
				ActionType: ActionOptimize,
				ActionData: JSONMap{
					"suggested_action": "investigate_failure_pattern",
					"failure_code":     failureCode,
				},
			}
			insights = append(insights, insight)
		}
	}

	return insights, rows.Err()
}

func (a *FOMAnalyzer) analyzeCostOptimization(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		SELECT
			g.goal_type,
			AVG(r.total_cost) as avg_cost,
			MAX(r.total_cost) as max_cost,
			COUNT(*) as execution_count,
			SUM(r.total_cost) as total_cost
		FROM fom_results r
		JOIN fom_plans p ON r.plan_id = p.id
		JOIN fom_goals g ON p.goal_id = g.id
		WHERE g.tenant_id = $1
		AND r.created_at > NOW() - INTERVAL '7 days'
		GROUP BY g.goal_type
		HAVING COUNT(*) >= 10 AND SUM(r.total_cost) > 10
		ORDER BY total_cost DESC
	`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var goalType string
		var avgCost, maxCost, totalCost float64
		var executionCount int

		if err := rows.Scan(&goalType, &avgCost, &maxCost, &executionCount, &totalCost); err != nil {
			continue
		}

		if avgCost > 0.50 {
			insight := &Insight{
				ID:         uuid.New(),
				TenantID:   tenantID,
				Category:   CategoryFOM,
				Severity:   SeverityOpportunity,
				Title:      "High-cost workflow: " + goalType,
				Message:    "Workflow type '" + goalType + "' costs $" + formatMoney(avgCost) + " on average",
				Confidence: floatPtr(0.75),
				Status:     StatusActive,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				InsightData: JSONMap{
					"goal_type":    goalType,
					"avg_cost":     avgCost,
					"max_cost":     maxCost,
					"total_cost":   totalCost,
					"executions":   executionCount,
				},
				ActionType: ActionOptimize,
				ActionData: JSONMap{
					"suggested_action": "cost_optimization",
					"goal_type":        goalType,
				},
			}
			insights = append(insights, insight)
		}
	}

	return insights, rows.Err()
}

func (a *FOMAnalyzer) analyzeWorkflowOptimization(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		SELECT
			w.pattern_name,
			w.usage_count,
			w.avg_success_rate,
			w.avg_cost,
			w.avg_time_ms
		FROM fom_workflow_patterns w
		JOIN fom_goals g ON w.goal_type = g.goal_type
		WHERE g.tenant_id = $1
		AND w.usage_count >= 20
		AND w.avg_success_rate < 0.85
		ORDER BY w.usage_count DESC
		LIMIT 5
	`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var patternName string
		var usageCount int
		var avgSuccessRate, avgCost float64
		var avgTimeMs int

		if err := rows.Scan(&patternName, &usageCount, &avgSuccessRate, &avgCost, &avgTimeMs); err != nil {
			continue
		}

		insight := &Insight{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Category:   CategoryFOM,
			Severity:   SeverityWarning,
			Title:      "Low-performing workflow pattern",
			Message:    "Pattern '" + patternName + "' succeeds only " + formatPercent(avgSuccessRate) + " of the time across " + formatInt(usageCount) + " uses",
			Confidence: floatPtr(0.90),
			Status:     StatusActive,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			InsightData: JSONMap{
				"pattern_name":     patternName,
				"usage_count":      usageCount,
				"success_rate":     avgSuccessRate,
				"avg_cost":         avgCost,
				"avg_time_ms":      avgTimeMs,
			},
			ActionType: ActionOptimize,
			ActionData: JSONMap{
				"suggested_action": "workflow_replacement",
				"pattern_name":     patternName,
			},
		}
		insights = append(insights, insight)
	}

	return insights, rows.Err()
}

func floatPtr(f float64) *float64 {
	return &f
}

func formatDuration(ms int) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000.0)
}

func formatPercent(f float64) string {
	return fmt.Sprintf("%.0f%%", f*100)
}

func formatMoney(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}