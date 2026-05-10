package consciousness

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ScoreComputer computes the System Awareness Score (0-100).
type ScoreComputer struct {
	db      *sql.DB
	logger  *logrus.Logger
	weights ScoreWeights
}

// NewScoreComputer creates a new score computer.
func NewScoreComputer(db *sql.DB, logger *logrus.Logger) *ScoreComputer {
	return &ScoreComputer{
		db:      db,
		logger:  logger,
		weights: DefaultScoreWeights(),
	}
}

// Compute calculates the System Awareness Score for a tenant.
func (sc *ScoreComputer) Compute(ctx context.Context, tenantID uuid.UUID) (*SystemAwarenessScore, error) {
	score := &SystemAwarenessScore{
		TenantID:   tenantID,
		ComputedAt: time.Now(),
	}

	// Health: Based on DNA fitness scores
	healthScore, funcCount, err := sc.computeHealthScore(ctx, tenantID)
	if err != nil {
		sc.logger.WithError(err).Warn("Failed to compute health score")
	}
	score.HealthScore = healthScore
	score.FunctionsAnalyzed = funcCount

	// Efficiency: Based on cost per execution trends
	efficiencyScore, err := sc.computeEfficiencyScore(ctx, tenantID)
	if err != nil {
		sc.logger.WithError(err).Warn("Failed to compute efficiency score")
	}
	score.EfficiencyScore = efficiencyScore

	// Scalability: Based on headroom before plan limits
	scalabilityScore, err := sc.computeScalabilityScore(ctx, tenantID)
	if err != nil {
		sc.logger.WithError(err).Warn("Failed to compute scalability score")
	}
	score.ScalabilityScore = scalabilityScore

	// Reliability: Based on success rates and error patterns
	reliabilityScore, err := sc.computeReliabilityScore(ctx, tenantID)
	if err != nil {
		sc.logger.WithError(err).Warn("Failed to compute reliability score")
	}
	score.ReliabilityScore = reliabilityScore

	// Optimization: Default to 70 (neutral) — enhanced when marketplace analyzer is wired
	score.OptimizationScore = 70

	// Count active insights
	active, critical, err := sc.countInsights(ctx, tenantID)
	if err == nil {
		score.ActiveInsights = active
		score.CriticalInsights = critical
	}

	// Weighted overall score
	score.OverallScore = score.HealthScore*sc.weights.Health +
		score.EfficiencyScore*sc.weights.Efficiency +
		score.ScalabilityScore*sc.weights.Scalability +
		score.ReliabilityScore*sc.weights.Reliability +
		score.OptimizationScore*sc.weights.Optimization

	// Clamp to 0-100
	if score.OverallScore > 100 {
		score.OverallScore = 100
	}
	if score.OverallScore < 0 {
		score.OverallScore = 0
	}

	return score, nil
}

// computeHealthScore returns 0-100 based on average DNA fitness scores.
func (sc *ScoreComputer) computeHealthScore(ctx context.Context, tenantID uuid.UUID) (float64, int, error) {
	query := `
		SELECT COALESCE(AVG(fitness_score), 75), COUNT(*)
		FROM function_dna_profiles
		WHERE tenant_id = $1 AND total_executions > 10`

	var avgFitness float64
	var count int
	err := sc.db.QueryRowContext(ctx, query, tenantID).Scan(&avgFitness, &count)
	if err == sql.ErrNoRows {
		return 75, 0, nil // Default for tenants with no data
	}
	return avgFitness, count, err
}

// computeEfficiencyScore returns 0-100 based on cost trends.
func (sc *ScoreComputer) computeEfficiencyScore(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	// Compare current week cost vs previous week
	query := `
		WITH current_week AS (
			SELECT COALESCE(SUM(total_cost_cents), 0) as cost
			FROM cost_allocation_entries
			WHERE tenant_id = $1 AND timestamp > NOW() - INTERVAL '7 days'
		),
		previous_week AS (
			SELECT COALESCE(SUM(total_cost_cents), 0) as cost
			FROM cost_allocation_entries
			WHERE tenant_id = $1
			AND timestamp > NOW() - INTERVAL '14 days'
			AND timestamp <= NOW() - INTERVAL '7 days'
		)
		SELECT c.cost, p.cost FROM current_week c, previous_week p`

	var currentCost, previousCost int64
	err := sc.db.QueryRowContext(ctx, query, tenantID).Scan(&currentCost, &previousCost)
	if err != nil && err != sql.ErrNoRows {
		return 70, err
	}

	if previousCost == 0 {
		return 75, nil // No baseline
	}

	// Lower cost = higher score; cost increase = lower score
	ratio := float64(currentCost) / float64(previousCost)
	switch {
	case ratio <= 0.8:
		return 95, nil // Costs decreasing
	case ratio <= 1.0:
		return 85, nil // Costs stable/decreasing
	case ratio <= 1.2:
		return 70, nil // Slight increase
	case ratio <= 1.5:
		return 55, nil // Moderate increase
	default:
		return 35, nil // Significant increase
	}
}

// computeScalabilityScore returns 0-100 based on headroom before limits.
func (sc *ScoreComputer) computeScalabilityScore(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	var plan string
	if err := sc.db.QueryRowContext(ctx, "SELECT plan FROM tenants WHERE id = $1", tenantID).Scan(&plan); err != nil {
		return 70, err
	}

	limit := planRequestLimit(plan)
	if limit <= 0 {
		return 95, nil // Unlimited plan
	}

	var usage int64
	err := sc.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM usage_events WHERE tenant_id = $1 AND created_at > DATE_TRUNC('month', NOW())",
		tenantID).Scan(&usage)
	if err != nil && err != sql.ErrNoRows {
		return 70, err
	}

	headroom := 1.0 - (float64(usage) / float64(limit))
	switch {
	case headroom > 0.7:
		return 95, nil
	case headroom > 0.5:
		return 80, nil
	case headroom > 0.3:
		return 60, nil
	case headroom > 0.1:
		return 40, nil
	default:
		return 20, nil
	}
}

// computeReliabilityScore returns 0-100 based on success rates.
func (sc *ScoreComputer) computeReliabilityScore(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(AVG(success_rate * 100), 90)
		FROM function_dna_profiles
		WHERE tenant_id = $1 AND total_executions > 10`

	var avgSuccessRate float64
	err := sc.db.QueryRowContext(ctx, query, tenantID).Scan(&avgSuccessRate)
	if err == sql.ErrNoRows {
		return 85, nil
	}
	return avgSuccessRate, err
}

// countInsights returns active and critical insight counts.
func (sc *ScoreComputer) countInsights(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'active'),
			COUNT(*) FILTER (WHERE status = 'active' AND severity = 'critical')
		FROM consciousness_insights WHERE tenant_id = $1`
	var active, critical int
	err := sc.db.QueryRowContext(ctx, query, tenantID).Scan(&active, &critical)
	return active, critical, err
}
