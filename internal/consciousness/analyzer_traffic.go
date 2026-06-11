package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TrafficAnalyzer detects traffic pattern changes.
type TrafficAnalyzer struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewTrafficAnalyzer(db *sql.DB, logger *logrus.Logger) *TrafficAnalyzer {
	return &TrafficAnalyzer{db: db, logger: logger}
}

func (a *TrafficAnalyzer) Name() string              { return "traffic" }
func (a *TrafficAnalyzer) Category() InsightCategory { return CategoryTraffic }

func (a *TrafficAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	var insights []*Insight

	query := `
		WITH current_period AS (
			SELECT function_id, COUNT(*) as current_count
			FROM usage_events
			WHERE tenant_id = $1
			AND created_at > NOW() - INTERVAL '24 hours'
			GROUP BY function_id
		),
		historical_period AS (
			SELECT function_id, COUNT(*) / 7.0 as daily_avg
			FROM usage_events
			WHERE tenant_id = $1
			AND created_at > NOW() - INTERVAL '7 days'
			AND created_at <= NOW() - INTERVAL '24 hours'
			GROUP BY function_id
		)
		SELECT c.function_id, c.current_count, h.daily_avg
		FROM current_period c
		JOIN historical_period h ON c.function_id = h.function_id
		WHERE h.daily_avg > 10
		AND c.current_count > h.daily_avg * 3`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query traffic patterns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var functionID string
		var currentCount int64
		var dailyAvg float64

		if err := rows.Scan(&functionID, &currentCount, &dailyAvg); err != nil {
			a.logger.WithError(err).Error("Failed to scan traffic pattern row")
			continue
		}

		fid, _ := uuid.Parse(functionID)
		if fid == uuid.Nil {
			continue
		}

		multiplier := float64(currentCount) / dailyAvg
		severity := SeverityWarning
		if multiplier > 5 {
			severity = SeverityCritical
		}

		confidence := 0.85
		trajectory := TrajectoryDegrading

		insights = append(insights, &Insight{
			TenantID:   tenantID,
			Category:   CategoryTraffic,
			Severity:   severity,
			Priority:   SeverityWeight(severity) * 10,
			Title:      fmt.Sprintf("Traffic spike: %.1fx normal volume", multiplier),
			Message:    fmt.Sprintf("Your function is handling %.1fx more traffic than the 7-day average (%d vs %.0f daily). Based on this trajectory you may hit scaling issues soon.", multiplier, currentCount, dailyAvg),
			Summary:    strPtr(fmt.Sprintf("%.1fx traffic spike", multiplier)),
			FunctionID: &fid,
			InsightData: JSONMap{
				"current_count": currentCount,
				"daily_average": dailyAvg,
				"multiplier":    multiplier,
			},
			Trajectory: &trajectory,
			Confidence: &confidence,
			Status:     StatusActive,
			ExpiresAt:  timePtr(time.Now().Add(3 * 24 * time.Hour)),
		})
	}

	return insights, rows.Err()
}
