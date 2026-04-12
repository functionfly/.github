// Package monitoring provides registry-specific metrics and monitoring for FunctionFly.
package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

// Constants for metric collection and thresholds
const (
	// DefaultMetricsCollectionInterval is the default interval for collecting metrics from DB
	DefaultMetricsCollectionInterval = 30 * time.Second

	// SlowQueryThreshold defines what constitutes a slow query
	SlowQueryThreshold = 1 * time.Second

	// ActivePublisherWindow is the time window for considering a publisher "active"
	ActivePublisherWindow = 30 * 24 * time.Hour // 30 days

	// Trust score tiers
	trustTierCritical  = "critical"
	trustTierStandard  = "standard"
	trustTierCommunity = "community"
)

// RegistryMetrics holds all registry-specific Prometheus metrics
type RegistryMetrics struct {
	// Function counts
	TotalFunctions    prometheus.Gauge
	VerifiedFunctions prometheus.Gauge
	FunctionsByTier   *prometheus.GaugeVec

	// Execution metrics
	ExecutionDuration prometheus.Histogram
	ExecutionTotal    *prometheus.CounterVec
	ExecutionErrors   *prometheus.CounterVec

	// Trust and verification
	AverageTrustScore  prometheus.Gauge
	TrustScoreDist     *prometheus.HistogramVec
	VerificationRate   prometheus.Gauge

	// Registry operations
	RegistryQueryDuration prometheus.Histogram
	RegistryCacheHits     prometheus.Counter
	RegistryCacheMisses   prometheus.Counter

	// User metrics
	ActivePublishers prometheus.Gauge
	TotalWallets     prometheus.Gauge

	// Backup metrics
	LastBackupTime   prometheus.Gauge
	BackupDuration   prometheus.Histogram
	BackupSize       prometheus.Gauge
	BackupFailures   prometheus.Counter

	// Database metrics
	DBConnections    prometheus.Gauge
	DBQueryDuration  prometheus.Histogram
	SlowQueries      prometheus.Counter

	// Performance indicators
	CacheHitRate     prometheus.Gauge
	RegistryLatency  prometheus.Histogram

	registry prometheus.Registerer
	mu       sync.RWMutex

	// Graceful shutdown support
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewRegistryMetrics creates and initializes registry metrics
func NewRegistryMetrics() *RegistryMetrics {
	return &RegistryMetrics{
		TotalFunctions: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "functions_total",
			Help:      "Total number of functions in the registry",
		}),
		VerifiedFunctions: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "functions_verified_total",
			Help:      "Number of verified functions in the registry",
		}),
		FunctionsByTier: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "functions_by_tier",
			Help:      "Number of functions by trust tier",
		}, []string{"tier"}),
		ExecutionDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "execution_duration_seconds",
			Help:      "Duration of function executions",
			Buckets:   prometheus.DefBuckets,
		}),
		ExecutionTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "execution_total",
			Help:      "Total number of function executions",
		}, []string{"status", "function_id"}),
		ExecutionErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "execution_errors_total",
			Help:      "Total number of execution errors by type",
		}, []string{"error_type"}),
		AverageTrustScore: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "trust_score_average",
			Help:      "Average trust score across all functions",
		}),
		TrustScoreDist: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "trust_score_distribution",
			Help:      "Distribution of trust scores",
			Buckets:   []float64{0, 20, 40, 60, 80, 90, 95, 100},
		}, []string{"tier"}),
		VerificationRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "verification_rate",
			Help:      "Percentage of functions that are verified",
		}),
		RegistryQueryDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "query_duration_seconds",
			Help:      "Duration of registry queries",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		RegistryCacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "cache_hits_total",
			Help:      "Total number of registry cache hits",
		}),
		RegistryCacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "cache_misses_total",
			Help:      "Total number of registry cache misses",
		}),
		ActivePublishers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "publishers_active",
			Help:      "Number of active function publishers (published in last 30 days)",
		}),
		TotalWallets: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "wallets_total",
			Help:      "Total number of user wallets",
		}),
		LastBackupTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "backup",
			Name:      "last_success_timestamp",
			Help:      "Unix timestamp of last successful backup",
		}),
		BackupDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "backup",
			Name:      "duration_seconds",
			Help:      "Duration of backup operations",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
		}),
		BackupSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "backup",
			Name:      "size_bytes",
			Help:      "Size of last backup in bytes",
		}),
		BackupFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "backup",
			Name:      "failures_total",
			Help:      "Total number of backup failures",
		}),
		DBConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "database",
			Name:      "connections_active",
			Help:      "Number of active database connections",
		}),
		DBQueryDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "database",
			Name:      "query_duration_seconds",
			Help:      "Duration of database queries",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		SlowQueries: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "functionfly",
			Subsystem: "database",
			Name:      "slow_queries_total",
			Help:      "Total number of slow queries (>1s)",
		}),
		CacheHitRate: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "functionfly",
			Subsystem: "cache",
			Name:      "hit_rate_percent",
			Help:      "Cache hit rate percentage",
		}),
		RegistryLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "functionfly",
			Subsystem: "registry",
			Name:      "api_latency_seconds",
			Help:      "Registry API latency",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		stopCh: make(chan struct{}),
	}
}

// Register registers all metrics with the provided registry
func (m *RegistryMetrics) Register(r prometheus.Registerer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.registry = r

	metrics := []prometheus.Collector{
		m.TotalFunctions,
		m.VerifiedFunctions,
		m.FunctionsByTier,
		m.ExecutionDuration,
		m.ExecutionTotal,
		m.ExecutionErrors,
		m.AverageTrustScore,
		m.TrustScoreDist,
		m.VerificationRate,
		m.RegistryQueryDuration,
		m.RegistryCacheHits,
		m.RegistryCacheMisses,
		m.ActivePublishers,
		m.TotalWallets,
		m.LastBackupTime,
		m.BackupDuration,
		m.BackupSize,
		m.BackupFailures,
		m.DBConnections,
		m.DBQueryDuration,
		m.SlowQueries,
		m.CacheHitRate,
		m.RegistryLatency,
	}

	for _, metric := range metrics {
		if err := r.Register(metric); err != nil {
			// Handle already registered error gracefully
			if areRegistered, ok := err.(prometheus.AlreadyRegisteredError); ok {
				log.Printf("[RegistryMetrics] Metric already registered: %v", areRegistered.ExistingCollector)
				continue
			}
			return fmt.Errorf("failed to register metric: %w", err)
		}
	}

	return nil
}

// Unregister removes all metrics from the registry
func (m *RegistryMetrics) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.registry == nil {
		return
	}

	m.registry.Unregister(m.TotalFunctions)
	m.registry.Unregister(m.VerifiedFunctions)
	m.registry.Unregister(m.FunctionsByTier)
	m.registry.Unregister(m.ExecutionDuration)
	m.registry.Unregister(m.ExecutionTotal)
	m.registry.Unregister(m.ExecutionErrors)
	m.registry.Unregister(m.AverageTrustScore)
	m.registry.Unregister(m.TrustScoreDist)
	m.registry.Unregister(m.VerificationRate)
	m.registry.Unregister(m.RegistryQueryDuration)
	m.registry.Unregister(m.RegistryCacheHits)
	m.registry.Unregister(m.RegistryCacheMisses)
	m.registry.Unregister(m.ActivePublishers)
	m.registry.Unregister(m.TotalWallets)
	m.registry.Unregister(m.LastBackupTime)
	m.registry.Unregister(m.BackupDuration)
	m.registry.Unregister(m.BackupSize)
	m.registry.Unregister(m.BackupFailures)
	m.registry.Unregister(m.DBConnections)
	m.registry.Unregister(m.DBQueryDuration)
	m.registry.Unregister(m.SlowQueries)
	m.registry.Unregister(m.CacheHitRate)
	m.registry.Unregister(m.RegistryLatency)
}

// Stop gracefully stops the metrics collection goroutine
func (m *RegistryMetrics) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// getErrorType categorizes errors to prevent high cardinality label values
func getErrorType(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()

	// Categorize common error types to prevent high cardinality
	switch {
	case contains(errStr, "connection") || contains(errStr, "dial"):
		return "connection_error"
	case contains(errStr, "timeout") || contains(errStr, "deadline"):
		return "timeout"
	case contains(errStr, "not found") || contains(errStr, "404"):
		return "not_found"
	case contains(errStr, "unauthorized") || contains(errStr, "401"):
		return "unauthorized"
	case contains(errStr, "forbidden") || contains(errStr, "403"):
		return "forbidden"
	case contains(errStr, "rate limit") || contains(errStr, "429"):
		return "rate_limited"
	case contains(errStr, "validation") || contains(errStr, "invalid"):
		return "validation_error"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RecordExecution records a function execution
func (m *RegistryMetrics) RecordExecution(functionID string, duration time.Duration, err error) {
	m.ExecutionDuration.Observe(duration.Seconds())

	status := "success"
	if err != nil {
		status = "error"
		errorType := getErrorType(err)
		m.ExecutionErrors.WithLabelValues(errorType).Inc()
	}

	m.ExecutionTotal.WithLabelValues(status, functionID).Inc()
}

// RecordQuery records a registry query
func (m *RegistryMetrics) RecordQuery(duration time.Duration, cached bool) {
	m.RegistryQueryDuration.Observe(duration.Seconds())

	if cached {
		m.RegistryCacheHits.Inc()
	} else {
		m.RegistryCacheMisses.Inc()
	}

	// Update cache hit rate (with mutex protection)
	m.updateCacheHitRate()
}

// RecordDBQuery records a database query
func (m *RegistryMetrics) RecordDBQuery(duration time.Duration) {
	m.DBQueryDuration.Observe(duration.Seconds())

	if duration > SlowQueryThreshold {
		m.SlowQueries.Inc()
	}
}

// RecordBackup records a backup operation
func (m *RegistryMetrics) RecordBackup(duration time.Duration, sizeBytes int64, err error) {
	m.BackupDuration.Observe(duration.Seconds())

	if err != nil {
		m.BackupFailures.Inc()
	} else {
		m.LastBackupTime.Set(float64(time.Now().Unix()))
		m.BackupSize.Set(float64(sizeBytes))
	}
}

// UpdateFunctionsCount updates function count metrics from database
func (m *RegistryMetrics) UpdateFunctionsCount(ctx context.Context, db *sql.DB) error {
	// Total functions
	var total int64
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM registry_functions").Scan(&total)
	if err != nil {
		return fmt.Errorf("failed to count functions: %w", err)
	}

	m.mu.Lock()
	m.TotalFunctions.Set(float64(total))
	m.mu.Unlock()

	// Verified functions
	var verified int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM registry_functions WHERE verified = true").Scan(&verified)
	if err != nil {
		return fmt.Errorf("failed to count verified functions: %w", err)
	}

	m.mu.Lock()
	m.VerifiedFunctions.Set(float64(verified))
	m.mu.Unlock()

	// Verification rate
	if total > 0 {
		m.mu.Lock()
		m.VerificationRate.Set(float64(verified) / float64(total) * 100)
		m.mu.Unlock()
	}

	return nil
}

// UpdateTrustMetrics updates trust score metrics
func (m *RegistryMetrics) UpdateTrustMetrics(ctx context.Context, db *sql.DB) error {
	// Average trust score
	var avgScore float64
	err := db.QueryRowContext(ctx, "SELECT COALESCE(AVG(trust_score), 0) FROM registry_functions").Scan(&avgScore)
	if err != nil {
		return fmt.Errorf("failed to get average trust score: %w", err)
	}

	m.mu.Lock()
	m.AverageTrustScore.Set(avgScore)
	m.mu.Unlock()

	// Functions by tier - using safe parameterized queries
	tierQueries := []struct {
		tier  string
		query string
	}{
		{trustTierCritical, "SELECT COUNT(*) FROM registry_functions WHERE trust_score >= 95 AND verified = true"},
		{trustTierStandard, "SELECT COUNT(*) FROM registry_functions WHERE trust_score >= 80 AND trust_score < 95"},
		{trustTierCommunity, "SELECT COUNT(*) FROM registry_functions WHERE trust_score < 80 OR (verified = false AND trust_score IS NOT NULL)"},
	}

	for _, tq := range tierQueries {
		var count int64
		if err := db.QueryRowContext(ctx, tq.query).Scan(&count); err != nil {
			log.Printf("[RegistryMetrics] Failed to count functions by tier %s: %v", tq.tier, err)
			continue
		}

		m.mu.Lock()
		m.FunctionsByTier.WithLabelValues(tq.tier).Set(float64(count))
		m.mu.Unlock()
	}

	return nil
}

// UpdatePublisherMetrics updates publisher-related metrics
func (m *RegistryMetrics) UpdatePublisherMetrics(ctx context.Context, db *sql.DB) error {
	// Active publishers (published in last 30 days)
	var publishers int64
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT author)
		FROM registry_functions
		WHERE updated_at > NOW() - INTERVAL '30 days'
	`).Scan(&publishers)
	if err != nil {
		return fmt.Errorf("failed to count publishers: %w", err)
	}

	m.mu.Lock()
	m.ActivePublishers.Set(float64(publishers))
	m.mu.Unlock()

	// Total wallets
	var wallets int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM registry_user_wallets").Scan(&wallets)
	if err != nil {
		return fmt.Errorf("failed to count wallets: %w", err)
	}

	m.mu.Lock()
	m.TotalWallets.Set(float64(wallets))
	m.mu.Unlock()

	return nil
}

// UpdateDBConnections updates database connection metrics
func (m *RegistryMetrics) UpdateDBConnections(db *sql.DB) {
	stats := db.Stats()

	m.mu.Lock()
	m.DBConnections.Set(float64(stats.OpenConnections))
	m.mu.Unlock()
}

// updateCacheHitRate recalculates cache hit rate
func (m *RegistryMetrics) updateCacheHitRate() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hits := getCounterValue(m.RegistryCacheHits)
	misses := getCounterValue(m.RegistryCacheMisses)
	total := hits + misses

	if total > 0 {
		rate := (hits / total) * 100
		m.CacheHitRate.Set(rate)
	}
}

// getCounterValue extracts the current value from a counter
func getCounterValue(c prometheus.Counter) float64 {
	ch := make(chan prometheus.Metric, 1)
	c.Collect(ch)
	close(ch)

	for metric := range ch {
		dto := &io_prometheus_client.Metric{}
		if err := metric.Write(dto); err == nil && dto.Counter != nil && dto.Counter.Value != nil {
			return *dto.Counter.Value
		}
	}
	return 0
}

// StartMetricsCollection starts periodic metrics collection from database
func (m *RegistryMetrics) StartMetricsCollection(ctx context.Context, db *sql.DB, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultMetricsCollectionInterval
	}

	ticker := time.NewTicker(interval)
	m.wg.Add(1)

	go func() {
		defer m.wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[RegistryMetrics] Stopping metrics collection due to context cancellation")
				return
			case <-m.stopCh:
				log.Println("[RegistryMetrics] Stopping metrics collection due to stop signal")
				return
			case <-ticker.C:
				if err := m.UpdateFunctionsCount(ctx, db); err != nil {
					log.Printf("[RegistryMetrics] Failed to update function counts: %v", err)
				}
				if err := m.UpdateTrustMetrics(ctx, db); err != nil {
					log.Printf("[RegistryMetrics] Failed to update trust metrics: %v", err)
				}
				if err := m.UpdatePublisherMetrics(ctx, db); err != nil {
					log.Printf("[RegistryMetrics] Failed to update publisher metrics: %v", err)
				}
				m.UpdateDBConnections(db)
			}
		}
	}()
}

// MetricsHandler returns an HTTP handler for metrics endpoint
func (m *RegistryMetrics) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// RecordAPICall records an API call with latency
func (m *RegistryMetrics) RecordAPICall(duration time.Duration) {
	m.RegistryLatency.Observe(duration.Seconds())
}

// HealthCheck performs a health check and returns status
func (m *RegistryMetrics) HealthCheck(ctx context.Context, db *sql.DB) map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Check database with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		health["status"] = "unhealthy"
		health["database"] = map[string]string{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		health["database"] = map[string]string{
			"status": "connected",
		}
	}

	// Get current metrics snapshot
	health["metrics"] = map[string]interface{}{
		"total_functions":    getGaugeValue(m.TotalFunctions),
		"verified_functions": getGaugeValue(m.VerifiedFunctions),
		"avg_trust_score":    getGaugeValue(m.AverageTrustScore),
		"active_publishers":  getGaugeValue(m.ActivePublishers),
		"cache_hit_rate":     getGaugeValue(m.CacheHitRate),
	}

	return health
}

// getGaugeValue extracts the current value from a gauge
func getGaugeValue(g prometheus.Gauge) float64 {
	ch := make(chan prometheus.Metric, 1)
	g.Collect(ch)
	close(ch)

	for metric := range ch {
		dto := &io_prometheus_client.Metric{}
		if err := metric.Write(dto); err == nil && dto.Gauge != nil && dto.Gauge.Value != nil {
			return *dto.Gauge.Value
		}
	}
	return 0
}

// Snapshot returns a snapshot of current metric values
func (m *RegistryMetrics) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"functions": map[string]interface{}{
			"total":     getGaugeValue(m.TotalFunctions),
			"verified":  getGaugeValue(m.VerifiedFunctions),
			"by_tier": map[string]float64{
				"critical":  getGaugeValue(m.FunctionsByTier.WithLabelValues(trustTierCritical)),
				"standard":  getGaugeValue(m.FunctionsByTier.WithLabelValues(trustTierStandard)),
				"community": getGaugeValue(m.FunctionsByTier.WithLabelValues(trustTierCommunity)),
			},
		},
		"execution": map[string]interface{}{
			"duration_p50": 0.0, // Could be computed from histogram
		},
		"trust": map[string]interface{}{
			"average": getGaugeValue(m.AverageTrustScore),
			"rate":    getGaugeValue(m.VerificationRate),
		},
		"backup": map[string]interface{}{
			"last_time":  getGaugeValue(m.LastBackupTime),
			"failures":   getCounterValue(m.BackupFailures),
		},
		"database": map[string]interface{}{
			"connections": getGaugeValue(m.DBConnections),
			"slow_queries": getCounterValue(m.SlowQueries),
		},
		"cache": map[string]interface{}{
			"hit_rate": getGaugeValue(m.CacheHitRate),
			"hits":     getCounterValue(m.RegistryCacheHits),
			"misses":   getCounterValue(m.RegistryCacheMisses),
		},
	}
}
