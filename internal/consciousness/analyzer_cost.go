package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CostAnalyzer detects cost inefficiencies.
type CostAnalyzer struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewCostAnalyzer(db *sql.DB, logger *logrus.Logger) *CostAnalyzer {
	return &CostAnalyzer{db: db, logger: logger}
}

func (a *CostAnalyzer) Name() string          { return "cost" }
func (a *CostAnalyzer) Category() InsightCategory { return CategoryCost }

func (a *CostAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	var insights []*Insight

	query := `
		SELECT function_id,
			SUM(total_cost_cents) as total_cost,
			COUNT(*) as exec_count,
			AVG(duration_ms) as avg_duration
		FROM cost_allocation_entries
		WHERE tenant_id = $1
		AND timestamp > $2
		GROUP BY function_id
		HAVING SUM(total_cost_cents) > 100
		ORDER BY total_cost DESC
		LIMIT 20`

	since := time.Now().Add(-time.Duration(params.LookbackDays) * 24 * time.Hour)
	if params.LookbackDays <= 0 {
		since = time.Now().Add(-7 * 24 * time.Hour)
	}

	rows, err := a.db.QueryContext(ctx, query, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("query cost allocation: %w", err)
	}
	defer rows.Close()

	type funcCost struct {
		FunctionID  string
		TotalCost   int64
		ExecCount   int64
		AvgDuration float64
	}

	var costs []funcCost
	for rows.Next() {
		var fc funcCost
		if err := rows.Scan(&fc.FunctionID, &fc.TotalCost, &fc.ExecCount, &fc.AvgDuration); err != nil {
			continue
		}
		costs = append(costs, fc)
	}

	var totalCost int64
	for _, c := range costs {
		totalCost += c.TotalCost
	}

	for _, c := range costs {
		if totalCost > 0 && float64(c.TotalCost)/float64(totalCost) > 0.40 {
			fid, _ := uuid.Parse(c.FunctionID)
			if fid == uuid.Nil {
				continue
			}

			pct := float64(c.TotalCost) / float64(totalCost) * 100
			costDollars := float64(c.TotalCost) / 100.0

			severity := SeverityWarning
			if pct > 60 {
				severity = SeverityCritical
			}

			confidence := 0.90
			insights = append(insights, &Insight{
				TenantID:   tenantID,
				Category:   CategoryCost,
				Severity:   severity,
				Priority:   SeverityWeight(severity) * 10,
				Title:      fmt.Sprintf("Cost concentration: %.0f%% of spend on one function", pct),
				Message:    fmt.Sprintf("Your function is consuming $%.2f/mo (%.0f%% of your total cost) with %d executions averaging %.0fms. Consider optimizing or caching results.", costDollars, pct, c.ExecCount, c.AvgDuration),
				Summary:    strPtr(fmt.Sprintf("$%.2f — %.0f%% of spend", costDollars, pct)),
				FunctionID: &fid,
				InsightData: JSONMap{
					"total_cost_cents": c.TotalCost,
					"exec_count":       c.ExecCount,
					"avg_duration_ms":  c.AvgDuration,
					"cost_percentage":  pct,
				},
				ActionType: ActionOptimize,
				Confidence: &confidence,
				Status:     StatusActive,
				ExpiresAt:  timePtr(time.Now().Add(7 * 24 * time.Hour)),
			})
		}
	}

	return insights, nil
}
