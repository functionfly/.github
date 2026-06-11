package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	defaultMaxConcurrent        = 10
	defaultQueryTimeout         = 30 * time.Second
	defaultCircuitBreakerLimit  = 3
	defaultCircuitBreakerTTL    = 5 * time.Minute
	defaultMaxInsightsPerTenant = 50
)

type engineConfig struct {
	maxConcurrent        int
	queryTimeout         time.Duration
	circuitBreakerLimit  int
	circuitBreakerTTL    time.Duration
	maxInsightsPerTenant int
}

func loadEngineConfig() engineConfig {
	cfg := engineConfig{
		maxConcurrent:        defaultMaxConcurrent,
		queryTimeout:         defaultQueryTimeout,
		circuitBreakerLimit:  defaultCircuitBreakerLimit,
		circuitBreakerTTL:    defaultCircuitBreakerTTL,
		maxInsightsPerTenant: defaultMaxInsightsPerTenant,
	}

	if v := os.Getenv("CONSCIOUSNESS_MAX_CONCURRENT"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			cfg.maxConcurrent = n
		}
	}

	return cfg
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

type circuitBreaker struct {
	mu       sync.Mutex
	failures int
	lastFail time.Time
	ttl      time.Duration
	limit    int
}

func newCircuitBreaker(limit int, ttl time.Duration) *circuitBreaker {
	return &circuitBreaker{limit: limit, ttl: ttl}
}

func (cb *circuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.failures >= cb.limit && time.Since(cb.lastFail) < cb.ttl {
		return false
	}
	if time.Since(cb.lastFail) >= cb.ttl {
		cb.failures = 0
	}
	return true
}

func (cb *circuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFail = time.Now()
}

func (cb *circuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.failures > 0 {
		cb.failures--
	}
}

type analyzerCircuitBreakers struct {
	breakers map[string]*circuitBreaker
	mu       sync.RWMutex
}

func newAnalyzerCircuitBreakers(analyzers []Analyzer, limit int, ttl time.Duration) *analyzerCircuitBreakers {
	acb := &analyzerCircuitBreakers{
		breakers: make(map[string]*circuitBreaker),
	}
	for _, a := range analyzers {
		acb.breakers[a.Name()] = newCircuitBreaker(limit, ttl)
	}
	return acb
}

func (acb *analyzerCircuitBreakers) Allow(name string) bool {
	acb.mu.RLock()
	defer acb.mu.RUnlock()
	if cb, ok := acb.breakers[name]; ok {
		return cb.Allow()
	}
	return true
}

func (acb *analyzerCircuitBreakers) RecordFailure(name string) {
	acb.mu.RLock()
	defer acb.mu.RUnlock()
	if cb, ok := acb.breakers[name]; ok {
		cb.RecordFailure()
		circuitBreakerOpenTotal.WithLabelValues(name).Inc()
	}
}

func (acb *analyzerCircuitBreakers) RecordSuccess(name string) {
	acb.mu.RLock()
	defer acb.mu.RUnlock()
	if cb, ok := acb.breakers[name]; ok {
		cb.RecordSuccess()
	}
}

type tenantLock struct {
	mu     sync.Mutex
	active map[uuid.UUID]struct{}
}

func newTenantLock() *tenantLock {
	return &tenantLock{active: make(map[uuid.UUID]struct{})}
}

func (tl *tenantLock) TryLock(tenantID uuid.UUID) bool {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	if _, ok := tl.active[tenantID]; ok {
		return false
	}
	tl.active[tenantID] = struct{}{}
	return true
}

func (tl *tenantLock) Unlock(tenantID uuid.UUID) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	delete(tl.active, tenantID)
}

func (tl *tenantLock) Count() int {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return len(tl.active)
}

type dedupCache struct {
	mu     sync.RWMutex
	items  map[uuid.UUID]map[dedupKey]struct{}
	window time.Duration
}

type dedupKey struct {
	Category   InsightCategory
	FunctionID *uuid.UUID
}

func newDedupCache(window time.Duration) *dedupCache {
	return &dedupCache{
		items:  make(map[uuid.UUID]map[dedupKey]struct{}),
		window: window,
	}
}

func (dc *dedupCache) Load(tenantID uuid.UUID, category InsightCategory, functionID *uuid.UUID) bool {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if tenantItems, ok := dc.items[tenantID]; ok {
		key := dedupKey{Category: category, FunctionID: functionID}
		_, exists := tenantItems[key]
		return exists
	}
	return false
}

func (dc *dedupCache) Store(tenantID uuid.UUID, category InsightCategory, functionID *uuid.UUID) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if _, ok := dc.items[tenantID]; !ok {
		dc.items[tenantID] = make(map[dedupKey]struct{})
	}
	key := dedupKey{Category: category, FunctionID: functionID}
	dc.items[tenantID][key] = struct{}{}
}

func (dc *dedupCache) Invalidate(tenantID uuid.UUID) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	delete(dc.items, tenantID)
}

type Engine struct {
	analyzers       []Analyzer
	repo            *Repository
	scoreComputer   *ScoreComputer
	dispatcher      *NotificationDispatcher
	logger          *logrus.Logger
	config          engineConfig
	circuitBreakers *analyzerCircuitBreakers
	tenantLocks     *tenantLock
	dedupCache      *dedupCache
	stopCh          chan struct{}
	wg              sync.WaitGroup
	mu              sync.Mutex
	running         bool
}

func NewEngine(db *sql.DB, logger *logrus.Logger) *Engine {
	cfg := loadEngineConfig()

	analyzers := []Analyzer{
		NewHealthAnalyzer(db, logger),
		NewCostAnalyzer(db, logger),
		NewScalingAnalyzer(db, logger),
		NewTrafficAnalyzer(db, logger),
		NewRedundancyAnalyzer(db, logger),
		NewMarketplaceAnalyzer(db, logger),
	}

	repo := NewRepository(db, logger)
	scoreComputer := NewScoreComputer(db, logger)
	dispatcher := NewNotificationDispatcher(db, repo, logger)

	engine := &Engine{
		analyzers:       analyzers,
		repo:            repo,
		scoreComputer:   scoreComputer,
		dispatcher:      dispatcher,
		logger:          logger,
		config:          cfg,
		circuitBreakers: newAnalyzerCircuitBreakers(analyzers, cfg.circuitBreakerLimit, cfg.circuitBreakerTTL),
		tenantLocks:     newTenantLock(),
		dedupCache:      newDedupCache(6 * time.Hour),
		stopCh:          make(chan struct{}),
	}

	return engine
}

func (e *Engine) HealthCheck(ctx context.Context) error {
	tables := []string{
		"consciousness_insights",
		"system_awareness_scores",
		"consciousness_preferences",
		"consciousness_delivery_log",
		"function_dna_profiles",
		"cost_allocation_entries",
		"usage_events",
	}

	for _, table := range tables {
		query := fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", table)
		if _, err := e.repo.db.ExecContext(ctx, query); err != nil {
			if isRelationNotExist(err) {
				return fmt.Errorf("required table missing: %s", table)
			}
			return fmt.Errorf("table %s check failed: %w", table, err)
		}
	}

	return nil
}

func (e *Engine) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return nil
	}
	e.mu.Unlock()

	e.logger.Info("Shutting down consciousness engine")

	close(e.stopCh)

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.logger.Info("Consciousness engine shutdown complete")
	case <-ctx.Done():
		e.logger.WithError(ctx.Err()).Warn("Consciousness engine shutdown timed out")
		return ctx.Err()
	}

	e.mu.Lock()
	e.running = false
	e.mu.Unlock()

	return nil
}

func (e *Engine) isRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *Engine) setRunning(v bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = v
}

type AnalysisResult struct {
	TenantID       uuid.UUID
	Insights       []*Insight
	Score          *SystemAwarenessScore
	AnalyzedAt     time.Time
	DurationMs     int64
	AnalyzerErrors map[string]error
}

func (e *Engine) AnalyzeTenant(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) (*AnalysisResult, error) {
	start := time.Now()
	runCtx := contextWithQueryTimeout(ctx, e.config.queryTimeout)

	result := &AnalysisResult{
		TenantID:       tenantID,
		AnalyzedAt:     start,
		AnalyzerErrors: make(map[string]error),
	}

	prefs, err := e.repo.GetPreferences(runCtx, tenantID)
	if err != nil {
		e.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to load preferences, using defaults")
		prefs = DefaultPreferences(tenantID)
	}

	var (
		mu           sync.Mutex
		wg           sync.WaitGroup
		allInsights  []*Insight
		analyzerDone sync.Map
	)

	sem := make(chan struct{}, e.maxConcurrent)

	for _, a := range e.analyzers {
		if !categoryEnabled(string(a.Category()), prefs.EnabledCategories) {
			continue
		}

		if !e.circuitBreakers.Allow(a.Name()) {
			e.logger.WithField("analyzer", a.Name()).Warn("Circuit breaker open, skipping analyzer")
			continue
		}

		wg.Add(1)
		go func(a Analyzer) {
			defer wg.Done()
			if _, loaded := analyzerDone.LoadOrStore(a.Name(), true); loaded {
				return
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-e.stopCh:
				result.AnalyzerErrors[a.Name()] = fmt.Errorf("shutdown in progress")
				return
			case <-runCtx.Done():
				result.AnalyzerErrors[a.Name()] = runCtx.Err()
				return
			default:
			}

			aStart := time.Now()
			insights, err := a.Analyze(runCtx, tenantID, params)
			analyzerDuration.WithLabelValues(a.Name()).Observe(time.Since(aStart).Seconds())

			if err != nil {
				e.circuitBreakers.RecordFailure(a.Name())
				mu.Lock()
				result.AnalyzerErrors[a.Name()] = err
				mu.Unlock()
				analyzerErrorsTotal.WithLabelValues(a.Name()).Inc()
				e.logger.WithError(err).WithField("analyzer", a.Name()).Warn("Analyzer failed")
				return
			}

			e.circuitBreakers.RecordSuccess(a.Name())

			if len(insights) > 0 {
				mu.Lock()
				allInsights = append(allInsights, insights...)
				mu.Unlock()
			}
		}(a)
	}

	wg.Wait()

	existingKeys := e.getDedupKeys(runCtx, tenantID)

	var deduped []*Insight
	for _, insight := range allInsights {
		key := dedupKey{Category: insight.Category, FunctionID: insight.FunctionID}
		if _, exists := existingKeys[key]; exists {
			continue
		}
		if exists, _ := e.dedupCache.Load(tenantID, insight.Category, insight.FunctionID); exists {
			continue
		}

		if insight.ExpiresAt == nil {
			t := time.Now().Add(7 * 24 * time.Hour)
			insight.ExpiresAt = &t
		}
		if insight.Status == "" {
			insight.Status = StatusActive
		}

		if err := e.repo.CreateInsight(runCtx, insight); err != nil {
			e.logger.WithError(err).Warn("Failed to persist insight")
			continue
		}

		e.dedupCache.Store(tenantID, insight.Category, insight.FunctionID)
		deduped = append(deduped, insight)

		insightsCreatedTotal.WithLabelValues(string(insight.Category), string(insight.Severity)).Inc()
	}

	if len(deduped) > e.config.maxInsightsPerTenant {
		sort.Slice(deduped, func(i, j int) bool {
			return deduped[i].Priority > deduped[j].Priority
		})
		deduped = deduped[:e.config.maxInsightsPerTenant]
	}

	result.Insights = deduped

	score, err := e.scoreComputer.Compute(runCtx, tenantID)
	if err != nil {
		e.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to compute score")
	} else {
		previousScore, _ := e.repo.GetScore(runCtx, tenantID)
		if previousScore != nil {
			score.PreviousScore = &previousScore.OverallScore
			trend := computeTrend(previousScore.OverallScore, score.OverallScore)
			score.Trend = &trend
		}

		if err := e.repo.UpsertScore(runCtx, score); err != nil {
			e.logger.WithError(err).Warn("Failed to persist score")
		}
		result.Score = score
	}

	result.DurationMs = time.Since(start).Milliseconds()

	if len(deduped) > 0 {
		prefs, prefsErr := e.repo.GetPreferences(runCtx, tenantID)
		if prefsErr != nil {
			e.logger.WithError(prefsErr).Warn("Failed to load preferences for dispatch")
			prefs = DefaultPreferences(tenantID)
		}

		for _, insight := range deduped {
			channels := e.dispatcher.Dispatch(runCtx, insight, prefs)
			if len(channels) > 0 {
				_ = e.repo.MarkChannelsSent(runCtx, insight.ID, channels)
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

func contextWithQueryTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return contextWithTimeout(ctx, timeout)
}

func contextWithTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithTimeout(ctx, timeout)
}

func (e *Engine) getDedupKeys(ctx context.Context, tenantID uuid.UUID) map[dedupKey]struct{} {
	result := make(map[dedupKey]struct{})

	since := time.Now().Add(-6 * time.Hour)
	query := `
		SELECT category, function_id
		FROM consciousness_insights
		WHERE tenant_id = $1 AND status = 'active' AND created_at > $2`

	rows, err := e.repo.db.QueryContext(ctx, query, tenantID, since)
	if err != nil {
		e.logger.WithError(err).Warn("Dedup batch query failed")
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var functionID *uuid.UUID
		if err := rows.Scan(&category, &functionID); err != nil {
			continue
		}
		key := dedupKey{Category: InsightCategory(category), FunctionID: functionID}
		result[key] = struct{}{}
	}

	return result
}

func (e *Engine) AnalyzeAllTenants(ctx context.Context) error {
	e.setRunning(true)
	defer e.setRunning(false)

	runID := uuid.New().String()

	start := time.Now()
	e.logger.WithField("run_id", runID).Info("Starting consciousness analysis for all tenants")

	tenants, err := e.getConsciousnessTenants(ctx)
	if err != nil {
		return fmt.Errorf("get consciousness tenants: %w", err)
	}

	e.logger.WithField("run_id", runID).WithField("tenant_count", len(tenants)).Info("Starting consciousness analysis for all tenants")

	var wg sync.WaitGroup
	sem := make(chan struct{}, e.maxConcurrent)

	for _, tenantID := range tenants {
		select {
		case <-e.stopCh:
			e.logger.WithField("run_id", runID).Info("Analysis stopped mid-run")
			break
		case <-ctx.Done():
			e.logger.WithField("run_id", runID).WithError(ctx.Err()).Warn("Analysis cancelled")
			return ctx.Err()
		default:
		}

		if !e.tenantLocks.TryLock(tenantID) {
			e.logger.WithField("run_id", runID).WithField("tenant_id", tenantID).Debug("Tenant analysis already in progress, skipping")
			continue
		}

		wg.Add(1)
		go func(tid uuid.UUID) {
			defer wg.Done()
			defer e.tenantLocks.Unlock(tid)
			sem <- struct{}{}
			defer func() { <-sem }()

			concurrencyGauge.Inc()
			defer concurrencyGauge.Dec()

			lookback := e.getLookbackDays(ctx, tid)
			params := AnalysisParams{
				LookbackDays: lookback,
				Since:        time.Now().Add(-time.Duration(lookback) * 24 * time.Hour),
			}

			if _, err := e.AnalyzeTenant(ctx, tid, params); err != nil {
				e.logger.WithError(err).WithField("run_id", runID).WithField("tenant_id", tid).Error("Tenant analysis failed")
			}
		}(tenantID)
	}

	wg.Wait()

	expired, err := e.repo.ExpireOldInsights(ctx)
	if err != nil {
		e.logger.WithError(err).Warn("Failed to expire old insights")
	}
	if expired > 0 {
		e.logger.WithField("expired_count", expired).Info("Expired old insights")
	}

	duration := time.Since(start)
	analysisDuration.WithLabelValues("success").Observe(duration.Seconds())
	tenantsAnalyzedTotal.Add(float64(len(tenants)))

	e.logger.WithFields(logrus.Fields{
		"run_id":       runID,
		"tenant_count": len(tenants),
		"duration_ms":  duration.Milliseconds(),
	}).Info("Consciousness analysis run completed")

	return nil
}

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
