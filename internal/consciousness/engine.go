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

const (
	DefaultAnalysisTimeout = 5 * time.Minute
	DefaultDedupWindow     = 6 * time.Hour
	DefaultInsightExpiry   = 7 * 24 * time.Hour
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

// NewEngineWithConfig creates a new consciousness engine with custom configuration.
func NewEngineWithConfig(db *sql.DB, logger *logrus.Logger, maxConcurrent int) *Engine {
	engine := NewEngine(db, logger)
	engine.maxConcurrent = maxConcurrent
	return engine
}

// AnalyzeResult holds the output of an analysis run for one tenant.
type AnalysisResult struct {
	TenantID       uuid.UUID
	Insights       []*Insight
	Score          *SystemAwarenessScore
	AnalyzedAt     time.Time
	DurationMs     int64
	AnalyzerErrors map[string]error
}

// analyzerResult holds the result from a single analyzer goroutine.
type analyzerResult struct {
	insights []*Insight
	name     string
	err      error
}

// AnalyzeTenant runs all analyzers for a single tenant and produces insights.
func (e *Engine) AnalyzeTenant(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) (*AnalysisResult, error) {
	return e.AnalyzeTenantWithTimeout(ctx, tenantID, params, DefaultAnalysisTimeout)
}

// AnalyzeTenantWithTimeout runs all analyzers with a custom timeout.
func (e *Engine) AnalyzeTenantWithTimeout(ctx context.Context, tenantID uuid.UUID, params AnalysisParams, timeout time.Duration) (*AnalysisResult, error) {
	analysisCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	result := &AnalysisResult{
		TenantID:       tenantID,
		AnalyzedAt:     start,
		AnalyzerErrors: make(map[string]error),
	}

	prefs, err := e.repo.GetPreferences(analysisCtx, tenantID)
	if err != nil {
		e.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to load preferences, using defaults")
		prefs = DefaultPreferences(tenantID)
	}

	resultsCh := make(chan analyzerResult, len(e.analyzers))
	sem := make(chan struct{}, e.maxConcurrent)

	for _, a := range e.analyzers {
		if !categoryEnabled(string(a.Category()), prefs.EnabledCategories) {
			continue
		}

		select {
		case sem <- struct{}{}:
		case <-analysisCtx.Done():
			result.AnalyzerErrors[a.Name()] = analysisCtx.Err()
			continue
		}

		go func(a Analyzer) {
			defer func() { <-sem }()

			analyzerCtx, cancel := context.WithTimeout(analysisCtx, 30*time.Second)
			defer cancel()

			insights, err := a.Analyze(analyzerCtx, tenantID, params)
			resultsCh <- analyzerResult{
				insights: insights,
				name:     a.Name(),
				err:      err,
			}
		}(a)
	}

	for i := 0; i < len(e.analyzers); i++ {
		select {
		case res := <-resultsCh:
			if res.err != nil {
				result.AnalyzerErrors[res.name] = res.err
				e.logger.WithError(res.err).WithField("analyzer", res.name).Warn("Analyzer failed")
			} else if len(res.insights) > 0 {
				result.Insights = append(result.Insights, res.insights...)
			}
		case <-analysisCtx.Done():
			e.logger.WithError(analysisCtx.Err()).Warn("Analysis context cancelled while waiting for results")
			break
		}
	}
	close(resultsCh)

	if analysisCtx.Err() != nil {
		e.logger.WithError(analysisCtx.Err()).WithField("tenant_id", tenantID).Warn("Analysis timed out")
	}

	var deduped []*Insight
	for _, insight := range result.Insights {
		exists, err := e.repo.HasRecentInsight(analysisCtx, tenantID, insight.Category, insight.FunctionID, DefaultDedupWindow)
		if err != nil {
			e.logger.WithError(err).Warn("Dedup check failed")
		}
		if exists {
			continue
		}

		if insight.ExpiresAt == nil {
			t := time.Now().Add(DefaultInsightExpiry)
			insight.ExpiresAt = &t
		}
		if insight.Status == "" {
			insight.Status = StatusActive
		}

		if err := e.repo.CreateInsight(analysisCtx, insight); err != nil {
			e.logger.WithError(err).Warn("Failed to persist insight")
			continue
		}
		deduped = append(deduped, insight)
	}

	result.Insights = deduped

	score, err := e.scoreComputer.Compute(analysisCtx, tenantID)
	if err != nil {
		e.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to compute score")
	} else {
		previousScore, _ := e.repo.GetScore(analysisCtx, tenantID)
		if previousScore != nil {
			score.PreviousScore = &previousScore.OverallScore
			trend := computeTrend(previousScore.OverallScore, score.OverallScore)
			score.Trend = &trend
		}

		if err := e.repo.UpsertScore(analysisCtx, score); err != nil {
			e.logger.WithError(err).Warn("Failed to persist score")
		}
		result.Score = score
	}

	result.DurationMs = time.Since(start).Milliseconds()

	if len(deduped) > 0 {
		prefs, prefsErr := e.repo.GetPreferences(analysisCtx, tenantID)
		if prefsErr != nil {
			e.logger.WithError(prefsErr).Warn("Failed to load preferences for dispatch")
			prefs = DefaultPreferences(tenantID)
		}

		for _, insight := range deduped {
			channels := e.dispatcher.Dispatch(analysisCtx, insight, prefs)
			if len(channels) > 0 {
				_ = e.repo.MarkChannelsSent(analysisCtx, insight.ID, channels)
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
			e.logger.WithError(err).Error("Failed to scan tenant UUID")
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

// GetSchedulerStatus returns the current scheduler status for health checks.
func (e *Engine) GetSchedulerStatus() map[string]interface{} {
	return map[string]interface{}{
		"max_concurrent": e.maxConcurrent,
		"analyzer_count": len(e.analyzers),
	}
}
