package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HealthAnalyzer assesses function health via DNA fitness scores.
type HealthAnalyzer struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewHealthAnalyzer(db *sql.DB, logger *logrus.Logger) *HealthAnalyzer {
	return &HealthAnalyzer{db: db, logger: logger}
}

func (a *HealthAnalyzer) Name() string          { return "health" }
func (a *HealthAnalyzer) Category() InsightCategory { return CategoryHealth }

func (a *HealthAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	var insights []*Insight

	query := `
		SELECT p.function_id, p.function_type, p.fitness_score, p.total_executions,
			p.avg_latency_ms, p.p99_latency_ms, p.success_rate, p.cold_start_rate,
			p.error_distribution, p.bottleneck_signature
		FROM function_dna_profiles p
		WHERE p.tenant_id = $1
		AND p.fitness_score < 60
		AND p.total_executions > 10
		ORDER BY p.fitness_score ASC
		LIMIT 10`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query dna profiles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			functionID, functionType        string
			fitnessScore                    float64
			totalExecutions                 int64
			avgLatency, p99Latency          float64
			successRate, coldStartRate      float64
			errorDist, bottleneck           sql.NullString
		)

		if err := rows.Scan(&functionID, &functionType, &fitnessScore, &totalExecutions,
			&avgLatency, &p99Latency, &successRate, &coldStartRate,
			&errorDist, &bottleneck); err != nil {
			a.logger.WithError(err).Error("Failed to scan DNA profile")
			continue
		}

		fid, _ := uuid.Parse(functionID)
		if fid == uuid.Nil {
			continue
		}

		severity := SeverityWarning
		if fitnessScore < 40 {
			severity = SeverityCritical
		}

		trajectory := TrajectoryDegrading
		if successRate >= 0.95 {
			trajectory = TrajectoryStable
		}

		bottleneckText := "unknown issue"
		if bottleneck.Valid && bottleneck.String != "" {
			bottleneckText = bottleneck.String
		}

		confidence := 0.85
		insights = append(insights, &Insight{
			TenantID:   tenantID,
			Category:   CategoryHealth,
			Severity:   severity,
			Priority:   SeverityWeight(severity) * 10,
			Title:      fmt.Sprintf("Function health degraded: fitness at %.0f/100", fitnessScore),
			Message:    fmt.Sprintf("Your %s function has a fitness score of %.0f/100. The main issue: %s. Success rate is %.1f%% with P99 latency at %.0fms.", functionType, fitnessScore, bottleneckText, successRate*100, p99Latency),
			Summary:    strPtr(fmt.Sprintf("Fitness %.0f — %s", fitnessScore, bottleneckText)),
			FunctionID: &fid,
			InsightData: JSONMap{
				"fitness_score":   fitnessScore,
				"success_rate":    successRate,
				"p99_latency_ms":  p99Latency,
				"cold_start_rate": coldStartRate,
				"total_executions": totalExecutions,
			},
			ActionType: ActionOptimize,
			Trajectory: &trajectory,
			Confidence: &confidence,
			Status:     StatusActive,
			ExpiresAt:  timePtr(time.Now().Add(7 * 24 * time.Hour)),
		})
	}

	return insights, rows.Err()
}
