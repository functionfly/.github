package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Engine is the main orchestrator that runs all analyzers, computes the
// System Awareness Score, and dispatches insights to notification channels.
type Engine struct {
	analyzers     []Analyzer
	repo          *Repository
	scoreComputer *ScoreComputer
	dispatcher    *NotificationDispatcher
	logger        *logrus.Logger
	maxConcurrent int
}

// NewEngine creates a new consciousness engine.
func NewEngine(db *sql.DB, logger *logrus.Logger) *Engine {
	repo := NewRepository(db, logger)
	scoreComputer := NewScoreComputer(db, logger)
	dispatcher := NewNotificationDispatcher(db, repo, logger)

	return &Engine{
		analyzers: []Analyzer{
			NewHealthAnalyzer(db, logger),
			NewCostAnalyzer(db, logger),
			NewScalingAnalyzer(db, logger),
			NewTrafficAnalyzer(db, logger),
			NewRedundancyAnalyzer(db, logger),
			NewMarketplaceAnalyzer(db, logger),
		},
		repo:          repo,
		scoreComputer: scoreComputer,
		dispatcher:    dispatcher,
		logger:        logger,
		maxConcurrent: 10,
	}
}

// AnalysisResult holds the output of an analysis run for one tenant.
type AnalysisResult struct {
	TenantID       uuid.UUID
	Insights       []*Insight
	Score          *SystemAwarenessScore
	AnalyzedAt     time.Time
	DurationMs     int64
	AnalyzerErrors map[string]error
}

// AnalyzeTenant runs all analyzers for a single tenant and produces insights.
func (e *Engine) AnalyzeTenant(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) (*AnalysisResult, error) {
	start := time.Now()
	result := &AnalysisResult{
		TenantID:       tenantID,
		AnalyzedAt:     start,
		AnalyzerErrors: make(map[string]error),
	}

	// Load preferences for category filtering
	prefs, err := e.repo.GetPreferences(ctx, tenantID)
	if err != nil {
		e.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to load preferences, using defaults")
		prefs = DefaultPreferences(tenantID)
	}

	// Run analyzers concurrently
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		allInsights []*Insight
	)

	sem := make(chan struct{}, e.maxConcurrent)

	for _, a := range e.analyzers {
		// Skip analyzers whose category is disabled in preferences
		if !categoryEnabled(string(a.Category()), prefs.EnabledCategories) {
			continue
		}

		wg.Add(1)
		go func(a Analyzer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			insights, err := a.Analyze(ctx, tenantID, params)
			if err != nil {
				mu.Lock()
				result.AnalyzerErrors[a.Name()] = err
				mu.Unlock()
				e.logger.WithError(err).WithField("analyzer", a.Name()).Warn("Analyzer failed")
				return
			}

			if len(insights) > 0 {
				mu.Lock()
				allInsights = append(allInsights, insights...)
				mu.Unlock()
			}
		}(a)
	}

	wg.Wait()

	// Deduplicate insights (don't re-emit similar insights within the dedup window)
	var deduped []*Insight
	for _, insight := range allInsights {
		exists, err := e.repo.HasRecentInsight(ctx, tenantID, insight.Category, insight.FunctionID, 6*time.Hour)
		if err != nil {
			e.logger.WithError(err).Warn("Dedup check failed")
		}
		if exists {
			continue
		}

		// Set expiry if not set
		if insight.ExpiresAt == nil {
			t := time.Now().Add(7 * 24 * time.Hour)
			insight.ExpiresAt = &t
		}
		if insight.Status == "" {
			insight.Status = StatusActive
		}

		// Persist
		if err := e.repo.CreateInsight(ctx, insight); err != nil {
			e.logger.WithError(err).Warn("Failed to persist insight")
			continue
		}
		deduped = append(deduped, insight)
	}

	result.Insights = deduped

	// Compute awareness score
	score, err := e.scoreComputer.Compute(ctx, tenantID)
	if err != nil {
		e.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to compute score")
	} else {
		// Get previous score for trend detection
		previousScore, _ := e.repo.GetScore(ctx, tenantID)
		if previousScore != nil {
			score.PreviousScore = &previousScore.OverallScore
			trend := computeTrend(previousScore.OverallScore, score.OverallScore)
			score.Trend = &trend
		}

		if err := e.repo.UpsertScore(ctx, score); err != nil {
			e.logger.WithError(err).Warn("Failed to persist score")
		}
		result.Score = score
	}

	result.DurationMs = time.Since(start).Milliseconds()

	// Dispatch new insights through notification channels
	if len(deduped) > 0 {
		prefs, prefsErr := e.repo.GetPreferences(ctx, tenantID)
		if prefsErr != nil {
			e.logger.WithError(prefsErr).Warn("Failed to load preferences for dispatch")
			prefs = DefaultPreferences(tenantID)
		}

		for _, insight := range deduped {
			channels := e.dispatcher.Dispatch(ctx, insight, prefs)
			if len(channels) > 0 {
				_ = e.repo.MarkChannelsSent(ctx, insight.ID, channels)
			}
		}
	}

	e.logger.WithFields(logrus.Fields{
		"tenant_id":       tenantID,
		"insights":        len(deduped),
		"analyzer_errors": len(result.AnalyzerErrors),
		"duration_ms":     result.DurationMs,
	}).Info("Consciousness analysis completed")

	return result, nil
}

// AnalyzeAllTenants runs analysis for all tenants with consciousness enabled.
func (e *Engine) AnalyzeAllTenants(ctx context.Context) error {
	// Get all tenants with consciousness feature enabled
	tenants, err := e.getConsciousnessTenants(ctx)
	if err != nil {
		return fmt.Errorf("get consciousness tenants: %w", err)
	}

	e.logger.WithField("tenant_count", len(tenants)).Info("Starting consciousness analysis for all tenants")

	var wg sync.WaitGroup
	sem := make(chan struct{}, e.maxConcurrent)

	for _, tenantID := range tenants {
		wg.Add(1)
		go func(tid uuid.UUID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Determine lookback based on plan
			lookback := e.getLookbackDays(ctx, tid)
			params := AnalysisParams{
				LookbackDays: lookback,
				Since:        time.Now().Add(-time.Duration(lookback) * 24 * time.Hour),
			}

			if _, err := e.AnalyzeTenant(ctx, tid, params); err != nil {
				e.logger.WithError(err).WithField("tenant_id", tid).Error("Tenant analysis failed")
			}
		}(tenantID)
	}

	wg.Wait()

	// Expire old insights
	expired, err := e.repo.ExpireOldInsights(ctx)
	if err != nil {
		e.logger.WithError(err).Warn("Failed to expire old insights")
	}
	if expired > 0 {
		e.logger.WithField("expired_count", expired).Info("Expired old insights")
	}

	return nil
}

// getConsciousnessTenants returns tenant IDs that have consciousness enabled.
func (e *Engine) getConsciousnessTenants(ctx context.Context) ([]uuid.UUID, error) {
	query := `
		SELECT id FROM tenants
		WHERE plan IN ('professional', 'enterprise', 'agent_enterprise')
		AND status = 'active'`

	rows, err := e.repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		tenants = append(tenants, id)
	}
	return tenants, rows.Err()
}

// getLookbackDays returns the analysis lookback window for a tenant's plan.
func (e *Engine) getLookbackDays(ctx context.Context, tenantID uuid.UUID) int {
	var plan string
	if err := e.repo.db.QueryRowContext(ctx, "SELECT plan FROM tenants WHERE id = $1", tenantID).Scan(&plan); err != nil {
		return 7
	}
	switch plan {
	case "professional":
		return 7
	case "enterprise":
		return 30
	case "agent_enterprise":
		return 90
	default:
		return 7
	}
}

// computeTrend determines the trend direction.
func computeTrend(previous, current float64) string {
	diff := current - previous
	switch {
	case diff > 3:
		return "improving"
	case diff < -3:
		return "declining"
	default:
		return "stable"
	}
}
