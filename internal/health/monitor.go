package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/aws"
	"github.com/functionfly/functionfly/internal/adapters/cloudflare"
	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/deno"
	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/adapters/functionfly"
	"github.com/functionfly/functionfly/internal/adapters/vercel"
	"github.com/functionfly/functionfly/internal/circuitbreaker"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// cachedResult holds the last health check result for a backend.
type cachedResult struct {
	ok        bool
	degraded  bool
	checkedAt time.Time
}

// backendProbeState tracks consecutive failures for adaptive probe intervals.
type backendProbeState struct {
	consecutiveFailures int
	nextProbeAt         time.Time
}

// Monitor handles backend health monitoring and circuit breaker management
type Monitor struct {
	repo            storage.Repository
	baseProbeInterval time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
	stopOnce        sync.Once
	breakerMgr      *circuitbreaker.Manager
	adapterPool     sync.Map // map[string]common.ProviderAdapter

	// Result cache: short-lived cache to avoid DB queries in the router
	resultCache   sync.Map // map[uuid.UUID]*cachedResult
	cacheTTL      time.Duration

	// Adaptive probe interval tracking
	probeStates   sync.Map // map[uuid.UUID]*backendProbeState

	// Config
	fallbackEnabled   bool
	retentionDays     int
	httpClient        *http.Client
}

// NewMonitor creates a new health monitor with circuit breaker config from environment.
func NewMonitor(repo storage.Repository) *Monitor {
	return NewMonitorWithConfig(repo, nil)
}

// NewMonitorWithConfig creates a new health monitor with custom circuit breaker config.
// If config is nil, environment-based config is used.
func NewMonitorWithConfig(repo storage.Repository, config *circuitbreaker.Config) *Monitor {
	if config == nil {
		cfg := circuitbreaker.ConfigFromEnv()
		asyncPersist := circuitbreaker.NewAsyncPersistence(NewDBPersistence(repo), 1*time.Second)
		cfg.Persistence = asyncPersist
		cfg.OnStateChange = func(key string, from, to circuitbreaker.State) {
			circuitbreaker.MetricsOnStateChange(key, from, to)
			if cfg.Persistence != nil {
				_ = cfg.Persistence.Save(context.Background(), key, &circuitbreaker.StoredState{
					State: int(to),
					Since: time.Now(),
				})
			}
		}
		config = &cfg
	}

	// Load configurable intervals
	baseInterval := envDuration("HEALTH_CHECK_INTERVAL", 5*time.Second)
	cacheTTL := envDuration("HEALTH_CHECK_CACHE_TTL", 3*time.Second)
	retentionDays := envInt("HEALTH_CHECK_RETENTION_DAYS", 7)
	fallbackEnabled := envBool("HEALTH_CHECK_FALLBACK", true)
	probeTimeout := envDuration("HEALTH_CHECK_TIMEOUT", 30*time.Second)

	return &Monitor{
		repo:              repo,
		baseProbeInterval: baseInterval,
		stopChan:          make(chan struct{}),
		breakerMgr:        circuitbreaker.NewManager(*config),
		cacheTTL:          cacheTTL,
		fallbackEnabled:   fallbackEnabled,
		retentionDays:     retentionDays,
		httpClient:        &http.Client{Timeout: probeTimeout},
	}
}

// Start begins the health monitoring loop and the data retention cleanup loop.
func (m *Monitor) Start() {
	logrus.Info("Starting health monitor")

	m.wg.Add(2)
	go m.monitorLoop()
	go m.cleanupLoop()
}

// Stop gracefully stops the health monitoring
func (m *Monitor) Stop() {
	m.stopOnce.Do(func() {
		logrus.Info("Stopping health monitor")
		close(m.stopChan)
		m.wg.Wait()
		logrus.Info("Health monitor stopped")
	})
}

// GetBreakerManager returns the breaker manager for use by other subsystems (e.g., proxy).
func (m *Monitor) GetBreakerManager() *circuitbreaker.Manager {
	return m.breakerMgr
}

// GetCachedResult returns the cached health check result for a backend, if fresh.
// Returns nil if no cached result exists or if the cache has expired.
func (m *Monitor) GetCachedResult(backendID uuid.UUID) *cachedResult {
	v, ok := m.resultCache.Load(backendID)
	if !ok {
		return nil
	}
	cached := v.(*cachedResult)
	if time.Since(cached.checkedAt) > m.cacheTTL {
		return nil
	}
	return cached
}

// monitorLoop runs the continuous health monitoring
func (m *Monitor) monitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.baseProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.probeAllBackends()
		}
	}
}

// cleanupLoop removes old health check records periodically.
func (m *Monitor) cleanupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.runCleanup()
		}
	}
}

func (m *Monitor) runCleanup() {
	if m.retentionDays <= 0 {
		return
	}
	before := time.Now().AddDate(0, 0, -m.retentionDays)
	deleted, err := m.repo.DeleteHealthChecksBefore(context.Background(), before)
	if err != nil {
		logrus.WithError(err).Error("Failed to clean up old health checks")
		return
	}
	if deleted > 0 {
		logrus.WithFields(logrus.Fields{
			"deleted":       deleted,
			"retention_days": m.retentionDays,
		}).Info("Cleaned up old health check records")
	}
}

// probeAllBackends probes all backends for health
func (m *Monitor) probeAllBackends() {
	backends, err := m.getAllBackends(context.Background())
	if err != nil {
		logrus.WithError(err).Error("Failed to get backends for health check")
		return
	}

	now := time.Now()
	var wg sync.WaitGroup
	for _, backend := range backends {
		// Adaptive probe interval: skip backends that haven't reached their next probe time
		if state := m.getProbeState(backend.ID); state != nil && now.Before(state.nextProbeAt) {
			continue
		}

		wg.Add(1)
		go func(b *storage.Backend) {
			defer func() {
				if rec := recover(); rec != nil {
					logrus.WithFields(logrus.Fields{
						"panic":      rec,
						"stack":      string(debug.Stack()),
						"backend_id": b.ID,
						"backend":    b.URL,
					}).Error("Health monitor probeBackend goroutine panicked")
					wg.Done()
				}
			}()
			defer wg.Done()
			m.probeBackend(b)
		}(backend)
	}

	wg.Wait()

	// System edge probe
	if edgeURL := os.Getenv("EDGE_HEALTH_URL"); edgeURL != "" {
		m.probeSystemEdge(edgeURL, os.Getenv("EDGE_HEALTH_SECRET"))
	}
}

// probeSystemEdge probes the configured edge URL and updates Prometheus + in-memory edge stats.
func (m *Monitor) probeSystemEdge(edgeURL, sharedSecret string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	backend := &storage.Backend{
		ID:           uuid.Nil,
		URL:          edgeURL,
		Region:       "system",
		SharedSecret: sharedSecret,
		Provider:     "functionfly-edge",
	}
	adapter := functionfly.NewFunctionFlyAdapter()
	result, err := adapter.HealthCheck(ctx, backend)
	if err != nil {
		monitoring.UpdateEdgeProbeAndMetrics(false, 0, err.Error())
		logrus.WithError(err).WithField("edge_url", edgeURL).Warn("Edge health probe error")
		return
	}
	monitoring.UpdateEdgeProbeAndMetrics(result.OK, result.LatencyMs, result.ErrorMessage)
	if !result.OK {
		logrus.WithFields(logrus.Fields{
			"edge_url": edgeURL,
			"status":   result.StatusCode,
			"error":    result.ErrorMessage,
		}).Warn("Edge health probe failed")
	}
}

// getAllBackends gets all enabled backends
func (m *Monitor) getAllBackends(ctx context.Context) ([]*storage.Backend, error) {
	return m.repo.GetAllEnabledBackends(ctx)
}

// probeBackend probes a single backend for health with fallback chain.
func (m *Monitor) probeBackend(backend *storage.Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adapter := m.getAdapterForProvider(backend.Provider)
	if adapter == nil {
		m.recordHealthCheck(ctx, backend.ID, false, 0, 0, fmt.Sprintf("Unsupported provider '%s'", backend.Provider))
		m.recordBreakerResult(backend.ID, false)
		m.updateProbeState(backend.ID, false)
		return
	}

	// Primary health check (GET /healthz)
	result, err := adapter.HealthCheck(ctx, backend)
	if err != nil {
		m.recordHealthCheck(ctx, backend.ID, false, 0, 0, fmt.Sprintf("Health check error: %v", err))
		m.recordBreakerResult(backend.ID, false)
		m.updateProbeState(backend.ID, false)
		return
	}

	// Fallback chain: if /healthz returned 404, try GET / on the base URL
	if !result.OK && m.fallbackEnabled && result.StatusCode == http.StatusNotFound && backend.Provider != "aws-lambda" {
		fallbackResult := m.probeFallback(ctx, backend)
		if fallbackResult != nil && fallbackResult.OK {
			result = fallbackResult
			result.Degraded = true
		}
	}

	// Parse health response body for structured info (version, uptime)
	if result.OK && result.Version == "" {
		m.parseHealthBody(result)
	}

	// Update result cache
	m.resultCache.Store(backend.ID, &cachedResult{
		ok:        result.OK,
		degraded:  result.Degraded,
		checkedAt: time.Now(),
	})

	// Record results
	healthy := result.OK
	m.recordHealthCheck(ctx, backend.ID, healthy, result.StatusCode, result.LatencyMs, result.ErrorMessage)
	m.recordBreakerResult(backend.ID, healthy)
	m.updateProbeState(backend.ID, healthy)

	if result.Version != "" || result.Degraded {
		logrus.WithFields(logrus.Fields{
			"backend_id": backend.ID,
			"provider":   backend.Provider,
			"region":     result.Region,
			"version":    result.Version,
			"degraded":   result.Degraded,
		}).Debug("Health check completed with provider info")
	}
}

// probeFallback performs a fallback health check by hitting the base URL.
// Used when /healthz returns 404 (user function doesn't implement it).
func (m *Monitor) probeFallback(ctx context.Context, backend *storage.Backend) *common.HealthCheckResult {
	baseURL := strings.TrimSuffix(backend.URL, "/")
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/", nil)
	if err != nil {
		return nil
	}

	start := time.Now()
	resp, err := m.httpClient.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	// Check for provider-specific error pages that indicate "no function deployed".
	// Cloudflare returns HTML pages containing "error code: 1042" (or similar)
	// when the worker doesn't exist. The generic "workers.dev" substring was
	// removed because it false-positived on live workers that simply return 404
	// on the root path.
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if strings.Contains(bodyStr, "error code:") {
		return &common.HealthCheckResult{
			OK:           false,
			StatusCode:   resp.StatusCode,
			LatencyMs:    latencyMs,
			ErrorMessage: "backend returned provider error page (no function deployed)",
		}
	}

	// Any non-5xx response is considered "alive" (degraded healthy)
	if resp.StatusCode < 500 {
		return &common.HealthCheckResult{
			OK:         true,
			StatusCode: resp.StatusCode,
			LatencyMs:  latencyMs,
			Region:     backend.Region,
		}
	}

	return &common.HealthCheckResult{
		OK:         false,
		StatusCode: resp.StatusCode,
		LatencyMs:  latencyMs,
		ErrorMessage: fmt.Sprintf("fallback probe returned %d", resp.StatusCode),
	}
}

// healthBodyResponse represents a structured health check response body.
type healthBodyResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"`
}

// parseHealthBody attempts to parse a structured health response body.
func (m *Monitor) parseHealthBody(result *common.HealthCheckResult) {
	// This is a no-op in the current implementation because we don't
	// read the body from the adapter. The adapters would need to return
	// the body for this to work. This is a placeholder for when adapters
	// are updated to return body content.
	//
	// For now, version extraction happens via response headers in each adapter.
}

// recordHealthCheck records the result of a health check
func (m *Monitor) recordHealthCheck(ctx context.Context, backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) {
	logger := logrus.WithFields(logrus.Fields{
		"backend_id":  backendID,
		"operation":   "health_check",
		"healthy":     ok,
		"status_code": statusCode,
		"latency_ms":  latencyMs,
	})

	err := m.repo.InsertHealthCheck(ctx, backendID, ok, statusCode, latencyMs, errorMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to record health check")
		return
	}

	if ok {
		logger.Debug("Health check passed")
	} else {
		logger.WithField("error", errorMessage).Warn("Health check failed")
	}
}

// recordBreakerResult records a health check result into the shared circuit breaker.
func (m *Monitor) recordBreakerResult(backendID uuid.UUID, healthy bool) {
	breaker := m.breakerMgr.ForBackend(backendID)
	if healthy {
		breaker.RecordSuccess()
	} else {
		breaker.RecordFailure()
	}
}

// getAdapterForProvider returns the appropriate adapter for a provider.
// Adapters are pooled per provider to reuse http.Client connections.
func (m *Monitor) getAdapterForProvider(provider string) common.ProviderAdapter {
	if v, ok := m.adapterPool.Load(provider); ok {
		return v.(common.ProviderAdapter)
	}
	adapter := createAdapter(provider)
	if adapter != nil {
		m.adapterPool.Store(provider, adapter)
	}
	return adapter
}

// getProbeState returns the adaptive probe state for a backend.
func (m *Monitor) getProbeState(backendID uuid.UUID) *backendProbeState {
	v, ok := m.probeStates.Load(backendID)
	if !ok {
		return nil
	}
	return v.(*backendProbeState)
}

// updateProbeState updates the adaptive probe interval based on health check results.
func (m *Monitor) updateProbeState(backendID uuid.UUID, healthy bool) {
	v, _ := m.probeStates.LoadOrStore(backendID, &backendProbeState{})
	state := v.(*backendProbeState)

	if healthy {
		state.consecutiveFailures = 0
		state.nextProbeAt = time.Time{} // reset to immediate
	} else {
		state.consecutiveFailures++
		state.nextProbeAt = time.Now().Add(m.adaptiveInterval(state.consecutiveFailures))
	}
}

// adaptiveInterval returns the probe interval based on consecutive failures.
func (m *Monitor) adaptiveInterval(consecutiveFailures int) time.Duration {
	switch {
	case consecutiveFailures <= 2:
		return m.baseProbeInterval // 5s default
	case consecutiveFailures <= 10:
		return 15 * time.Second
	case consecutiveFailures <= 30:
		return 60 * time.Second
	default:
		return 5 * time.Minute
	}
}

func createAdapter(provider string) common.ProviderAdapter {
	switch provider {
	case "workers":
		return cloudflare.NewCloudflareAdapter()
	case "vercel":
		return vercel.NewVercelAdapter()
	case "fly":
		return fly.NewFlyAdapter()
	case "deno-deploy":
		return deno.NewDenoAdapter()
	case "functionfly-edge", "functionfly":
		return functionfly.NewFunctionFlyAdapter()
	case "aws-lambda":
		return aws.NewAWSAdapter()
	default:
		return nil
	}
}

// env helpers

func envDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultVal
	}
	return d
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func envBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultVal
	}
	return b
}
