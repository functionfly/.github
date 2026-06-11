package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ScalingAnalyzer predicts scaling bottlenecks.
type ScalingAnalyzer struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewScalingAnalyzer(db *sql.DB, logger *logrus.Logger) *ScalingAnalyzer {
	return &ScalingAnalyzer{db: db, logger: logger}
}

func (a *ScalingAnalyzer) Name() string          { return "scaling" }
func (a *ScalingAnalyzer) Category() InsightCategory { return CategoryScaling }

func (a *ScalingAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	var insights []*Insight

	var plan string
	err := a.db.QueryRowContext(ctx, "SELECT plan FROM tenants WHERE id = $1", tenantID).Scan(&plan)
	if err != nil {
		return nil, fmt.Errorf("get tenant plan: %w", err)
	}

	limit := planRequestLimit(plan)
	if limit <= 0 {
		return nil, nil
	}

	query := `
		SELECT COUNT(*) as daily_executions
		FROM usage_events
		WHERE tenant_id = $1
		AND created_at > NOW() - INTERVAL '7 days'`

	var weeklyExecutions int64
	if err := a.db.QueryRowContext(ctx, query, tenantID).Scan(&weeklyExecutions); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("query usage: %w", err)
	}

	dailyRate := float64(weeklyExecutions) / 7.0
	if dailyRate < 1 {
		return nil, nil
	}

	var currentUsage int64
	currentQuery := `
		SELECT COUNT(*) FROM usage_events
		WHERE tenant_id = $1
		AND created_at > DATE_TRUNC('month', NOW())`
	if err := a.db.QueryRowContext(ctx, currentQuery, tenantID).Scan(&currentUsage); err != nil && err != sql.ErrNoRows {
		a.logger.WithError(err).Error("Failed to scan current usage")
		return nil, err
	}

	remaining := int64(limit) - currentUsage
	if remaining <= 0 {
		remaining = 0
		confidence := 1.0
		trajectory := TrajectoryCritical
		insights = append(insights, &Insight{
			TenantID:  tenantID,
			Category:  CategoryScaling,
			Severity:  SeverityCritical,
			Priority:  40,
			Title:     "Monthly request limit reached",
			Message:   fmt.Sprintf("You've used %d of your %d monthly requests. Consider upgrading your plan or optimizing your function calls.", currentUsage, limit),
			Summary:   strPtr("Request limit reached"),
			InsightData: JSONMap{
				"current_usage": currentUsage,
				"limit":         limit,
				"daily_rate":    dailyRate,
			},
			Trajectory: &trajectory,
			Confidence: &confidence,
			Status:     StatusActive,
			ExpiresAt:  timePtr(time.Now().Add(24 * time.Hour)),
		})
		return insights, nil
	}

	daysUntilLimit := math.Ceil(float64(remaining) / dailyRate)

	if daysUntilLimit <= 14 {
		severity := SeverityWarning
		if daysUntilLimit <= 3 {
			severity = SeverityCritical
		}

		trajectory := TrajectoryDegrading
		confidence := 0.80
		projectedDays := int(daysUntilLimit)

		insights = append(insights, &Insight{
			TenantID:  tenantID,
			Category:  CategoryScaling,
			Severity:  severity,
			Priority:  SeverityWeight(severity) * 10,
			Title:     fmt.Sprintf("You'll hit your %d request limit in ~%d days", limit, projectedDays),
			Message:   fmt.Sprintf("At your current rate of %.0f requests/day, you'll exhaust your %d monthly limit in roughly %d days. Consider upgrading to Enterprise or optimizing your top functions.", dailyRate, limit, projectedDays),
			Summary:   strPtr(fmt.Sprintf("Limit hit in ~%d days", projectedDays)),
			InsightData: JSONMap{
				"current_usage":  currentUsage,
				"limit":          limit,
				"daily_rate":     dailyRate,
				"remaining":      remaining,
				"projected_days": projectedDays,
			},
			ActionType:    ActionScaleConfig,
			Trajectory:    &trajectory,
			ProjectedDays: &projectedDays,
			Confidence:    &confidence,
			Status:        StatusActive,
			ExpiresAt:     timePtr(time.Now().Add(3 * 24 * time.Hour)),
		})
	}

	return insights, nil
}
