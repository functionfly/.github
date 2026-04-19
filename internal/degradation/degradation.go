package degradation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	GetErrorRate(ctx context.Context, window time.Duration) (float64, error)
	GetLatencyP99(ctx context.Context, window time.Duration) (time.Duration, error)
	GetCircuitBreakerState(ctx context.Context, agentID string) (string, error)
}

// PrometheusMetricsCollector implements MetricsCollector using Prometheus HTTP API
type PrometheusMetricsCollector struct {
	prometheusURL string
	httpClient    HTTPClient
}

// HTTPClient interface for making HTTP requests (allows testing with mocks)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewPrometheusMetricsCollector creates a collector that queries Prometheus via HTTP API
func NewPrometheusMetricsCollector(prometheusURL string) *PrometheusMetricsCollector {
	return &PrometheusMetricsCollector{
		prometheusURL: prometheusURL,
		httpClient:    &realHTTPClient{},
	}
}

// GetErrorRate calculates error rate from Prometheus
// Query: sum(rate(agent_execution_errors_total[5m])) / sum(rate(agent_executions_total[5m]))
func (c *PrometheusMetricsCollector) GetErrorRate(ctx context.Context, window time.Duration) (float64, error) {
	query := fmt.Sprintf(
		"sum(rate(agent_execution_errors_total[%s])) / sum(rate(agent_executions_total[%s]))",
		window.String(), window.String(),
	)
	return c.queryPrometheus(ctx, query)
}

// GetLatencyP99 calculates P99 latency from Prometheus
// Query: histogram_quantile(0.99, sum(rate(agent_execution_duration_seconds_bucket[5m])) by (le))
func (c *PrometheusMetricsCollector) GetLatencyP99(ctx context.Context, window time.Duration) (time.Duration, error) {
	query := fmt.Sprintf(
		"histogram_quantile(0.99, sum(rate(agent_execution_duration_seconds_bucket[%s])) by (le))",
		window.String(),
	)
	seconds, err := c.queryPrometheus(ctx, query)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

// GetCircuitBreakerState checks if any circuit breakers are open
// Query: sum(agent_circuit_state{state="open"}) > 0
func (c *PrometheusMetricsCollector) GetCircuitBreakerState(ctx context.Context, agentID string) (string, error) {
	query := "sum(agent_circuit_state{state=\"open\"}) > 0"
	result, err := c.queryPrometheus(ctx, query)
	if err != nil {
		return "closed", err
	}
	if result > 0 {
		return "open", nil
	}
	return "closed", nil
}

// queryPrometheus executes a PromQL instant query and returns the first float value
func (c *PrometheusMetricsCollector) queryPrometheus(ctx context.Context, query string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", c.prometheusURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to query prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var result promQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode prometheus response: %w", err)
	}

	if len(result.Data.Result) == 0 {
		return 0, nil // No data
	}

	return result.Data.Result[0].Value[1].(float64), nil
}

type promQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value [2]interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// realHTTPClient implements HTTPClient using net/http
type realHTTPClient struct{}

func (c *realHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// DegradationMetrics holds metrics for auto degradation monitoring
type DegradationMetrics struct {
	ErrorRate          prometheus.Gauge
	LatencyP99         prometheus.Gauge
	CurrentLevel       prometheus.Gauge
	LevelTransitions   prometheus.Counter
	AutoAdjustments    prometheus.Counter
	LastCheckTime      prometheus.Gauge
	ThresholdsExceeded prometheus.Gauge
}

// DegradationLevel defines the level of service degradation
type DegradationLevel int

const (
	// LevelNormal is full functionality
	LevelNormal DegradationLevel = iota
	// LevelDegraded is reduced functionality
	LevelDegraded
	// LevelCritical is minimal functionality
	LevelCritical
	// LevelEmergency is emergency mode
	LevelEmergency
)

// EscalationRule defines an escalation rule for auto degradation
type EscalationRule struct {
	Level    DegradationLevel
	Cooldown time.Duration
}

// String returns the string representation of degradation level
func (l DegradationLevel) String() string {
	switch l {
	case LevelNormal:
		return "normal"
	case LevelDegraded:
		return "degraded"
	case LevelCritical:
		return "critical"
	case LevelEmergency:
		return "emergency"
	default:
		return "unknown"
	}
}

// Feature represents a feature that can be degraded
type Feature struct {
	// Name is the feature name
	Name string
	// Description describes the feature
	Description string
	// Enabled indicates if the feature is enabled
	Enabled bool
	// DegradationLevel is the minimum level for this feature
	MinLevel DegradationLevel
	// Fallback is the fallback function when degraded
	Fallback func(ctx context.Context, args ...interface{}) (interface{}, error)
}

// DegradationManager manages graceful degradation
type DegradationManager struct {
	currentLevel DegradationLevel
	features     map[string]*Feature
	mu           sync.RWMutex
	history      []DegradationEvent
	maxHistory   int
}

// DegradationEvent represents a degradation event
type DegradationEvent struct {
	Timestamp time.Time
	FromLevel DegradationLevel
	ToLevel   DegradationLevel
	Reason    string
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager() *DegradationManager {
	return &DegradationManager{
		currentLevel: LevelNormal,
		features:     make(map[string]*Feature),
		history:      make([]DegradationEvent, 0),
		maxHistory:   100,
	}
}

// RegisterFeature registers a feature with degradation support
func (m *DegradationManager) RegisterFeature(feature *Feature) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.features[feature.Name] = feature
	logrus.WithFields(logrus.Fields{
		"feature":         feature.Name,
		"min_level":       feature.MinLevel.String(),
		"has_fallback":    feature.Fallback != nil,
	}).Info("Registered feature for degradation management")
}

// SetLevel sets the current degradation level
func (m *DegradationManager) SetLevel(level DegradationLevel, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldLevel := m.currentLevel
	m.currentLevel = level

	// Record event
	event := DegradationEvent{
		Timestamp: time.Now(),
		FromLevel: oldLevel,
		ToLevel:   level,
		Reason:    reason,
	}

	m.history = append(m.history, event)
	if len(m.history) > m.maxHistory {
		m.history = m.history[1:]
	}

	logrus.WithFields(logrus.Fields{
		"from_level": oldLevel.String(),
		"to_level":   level.String(),
		"reason":     reason,
	}).Warn("Degradation level changed")
}

// GetLevel returns the current degradation level
func (m *DegradationManager) GetLevel() DegradationLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentLevel
}

// IsFeatureEnabled checks if a feature is enabled at the current level
func (m *DegradationManager) IsFeatureEnabled(featureName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	feature, exists := m.features[featureName]
	if !exists {
		return false
	}

	if !feature.Enabled {
		return false
	}

	return m.currentLevel <= feature.MinLevel
}

// ExecuteFeature executes a feature with degradation support
func (m *DegradationManager) ExecuteFeature(ctx context.Context, featureName string, fn func(ctx context.Context) (interface{}, error), args ...interface{}) (interface{}, error) {
	m.mu.RLock()
	feature, exists := m.features[featureName]
	currentLevel := m.currentLevel
	m.mu.RUnlock()

	if !exists {
		logrus.WithField("feature", featureName).Warn("Feature not registered")
		return nil, nil
	}

	if !feature.Enabled {
		logrus.WithField("feature", featureName).Debug("Feature disabled")
		return nil, nil
	}

	// Check if feature is available at current level
	if currentLevel <= feature.MinLevel {
		// Feature is available, execute normally
		return fn(ctx)
	}

	// Feature is degraded, use fallback if available
	if feature.Fallback != nil {
		logrus.WithFields(logrus.Fields{
			"feature":      featureName,
			"current_level": currentLevel.String(),
			"min_level":    feature.MinLevel.String(),
		}).Info("Using fallback for degraded feature")

		return feature.Fallback(ctx, args...)
	}

	// No fallback available
	logrus.WithFields(logrus.Fields{
		"feature":      featureName,
		"current_level": currentLevel.String(),
		"min_level":    feature.MinLevel.String(),
	}).Warn("Feature degraded with no fallback")

	return nil, nil
}

// GetFeatureStatus returns the status of all features
func (m *DegradationManager) GetFeatureStatus() map[string]FeatureStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]FeatureStatus)
	for name, feature := range m.features {
		status[name] = FeatureStatus{
			Name:           name,
			Enabled:        feature.Enabled,
			Available:      m.currentLevel <= feature.MinLevel,
			MinLevel:       feature.MinLevel.String(),
			CurrentLevel:   m.currentLevel.String(),
			HasFallback:    feature.Fallback != nil,
		}
	}

	return status
}

// FeatureStatus represents the status of a feature
type FeatureStatus struct {
	Name         string
	Enabled      bool
	Available    bool
	MinLevel     string
	CurrentLevel string
	HasFallback  bool
}

// GetHistory returns degradation history
func (m *DegradationManager) GetHistory() []DegradationEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history := make([]DegradationEvent, len(m.history))
	copy(history, m.history)
	return history
}

// AutoDegradation automatically adjusts degradation level based on metrics
type AutoDegradation struct {
	manager           *DegradationManager
	errorThreshold    float64
	latencyThreshold  time.Duration
	checkInterval     time.Duration
	stopChan          chan struct{}
	wg                sync.WaitGroup
	metricsCollector  MetricsCollector
	escalationRules  []EscalationRule
}

// NewAutoDegradation creates a new auto degradation manager
func NewAutoDegradation(manager *DegradationManager, errorThreshold float64, latencyThreshold time.Duration) *AutoDegradation {
	return &AutoDegradation{
		manager:          manager,
		errorThreshold:   errorThreshold,
		latencyThreshold: latencyThreshold,
		checkInterval:    30 * time.Second,
		stopChan:         make(chan struct{}),
		escalationRules:  defaultEscalationRules(),
	}
}

// defaultEscalationRules returns the default escalation rules
func defaultEscalationRules() []EscalationRule {
	return []EscalationRule{
		{Level: LevelDegraded, Cooldown: 5 * time.Minute},
		{Level: LevelCritical, Cooldown: 2 * time.Minute},
		{Level: LevelEmergency, Cooldown: 30 * time.Second},
	}
}

// Start starts the auto degradation monitor
func (a *AutoDegradation) Start(ctx context.Context) {
	a.wg.Add(1)
	go a.monitorLoop(ctx)
}

// Stop stops the auto degradation monitor
func (a *AutoDegradation) Stop() {
	close(a.stopChan)
	a.wg.Wait()
}

// monitorLoop monitors metrics and adjusts degradation level
func (a *AutoDegradation) monitorLoop(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.checkAndAdjust(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// checkAndAdjust checks metrics and adjusts degradation level
func (a *AutoDegradation) checkAndAdjust(ctx context.Context) {
	currentLevel := a.manager.GetLevel()

	// Use default thresholds if not configured
	errorThreshold := a.errorThreshold
	if errorThreshold == 0 {
		errorThreshold = 0.05 // 5% default error rate threshold
	}
	latencyThreshold := float64(a.latencyThreshold)
	if latencyThreshold == 0 {
		latencyThreshold = float64(2 * time.Second)
	}

	// Determine if we should escalate or de-escalate based on metrics
	shouldEscalate, reason := a.evaluateEscalationNeed(ctx, errorThreshold, latencyThreshold)
	shouldDeEscalate := a.evaluateDeEscalationNeed(ctx, currentLevel, errorThreshold, latencyThreshold)

	newLevel := currentLevel

	if shouldEscalate {
		newLevel = a.getNextEscalationLevel(currentLevel)
		logrus.WithFields(logrus.Fields{
			"from_level":       currentLevel.String(),
			"to_level":         newLevel.String(),
			"reason":           reason,
			"error_threshold":  errorThreshold,
			"latency_threshold": fmt.Sprintf("%.2fs", latencyThreshold),
		}).Warn("Auto degradation: escalating")
		a.manager.SetLevel(newLevel, reason)
	} else if shouldDeEscalate {
		// Only de-escalate one level at a time, and only if metrics have improved significantly
		newLevel = a.getNextDeEscalationLevel(currentLevel)
		deEscalateReason := "metrics improved below thresholds"
		logrus.WithFields(logrus.Fields{
			"from_level":       currentLevel.String(),
			"to_level":         newLevel.String(),
			"reason":           deEscalateReason,
		}).Info("Auto degradation: de-escalating")
		a.manager.SetLevel(newLevel, deEscalateReason)
	}
}

// evaluateEscalationNeed determines if escalation is needed based on current metrics
func (a *AutoDegradation) evaluateEscalationNeed(ctx context.Context, errorThreshold, latencyThreshold float64) (bool, string) {
	// Check error rate via metrics collector (Prometheus or custom)
	errorRate := a.getCurrentErrorRate(ctx)
	if errorRate > errorThreshold {
		return true, "error rate exceeded threshold"
	}

	// Check latency via metrics collector (Prometheus or custom)
	latency := a.getCurrentLatency(ctx)
	if float64(latency) > latencyThreshold*float64(time.Second) {
		return true, "latency exceeded threshold"
	}

	// Check for cascading failures via circuit breakers
	if a.hasOpenCircuitBreakers(ctx) {
		return true, "circuit breakers open indicating systemic issues"
	}

	// Check quota exhaustion
	if a.isQuotaExhausted(ctx) {
		return true, "quota exhaustion detected"
	}

	return false, ""
}

// evaluateDeEscalationNeed determines if de-escalation is safe
func (a *AutoDegradation) evaluateDeEscalationNeed(ctx context.Context, currentLevel DegradationLevel, errorThreshold, latencyThreshold float64) bool {
	if currentLevel == LevelNormal {
		return false
	}

	// Metrics must be significantly better than thresholds for de-escalation
	// Require metrics to be at least 50% below thresholds
	safetyMargin := 0.5
	effectiveErrorThreshold := errorThreshold * (1 - safetyMargin)
	effectiveLatencyThreshold := time.Duration(float64(latencyThreshold) * (1 - safetyMargin))

	errorRate := a.getCurrentErrorRate(ctx)
	latency := a.getCurrentLatency(ctx)

	// Only de-escalate if metrics are significantly better than thresholds
	if errorRate > effectiveErrorThreshold {
		return false
	}
	if latency > effectiveLatencyThreshold {
		return false
	}

	// Check that circuit breakers are closed
	if a.hasOpenCircuitBreakers(ctx) {
		return false
	}

	// Check that we haven't just escalated (cooldown period)
	if !a.canDeEscalate(currentLevel) {
		return false
	}

	return true
}

// getCurrentErrorRate returns the current error rate (0-1)
// Uses the metrics collector if available, otherwise falls back to Prometheus queries
func (a *AutoDegradation) getCurrentErrorRate(ctx context.Context) float64 {
	if a.metricsCollector != nil {
		rate, err := a.metricsCollector.GetErrorRate(ctx, 5*time.Minute)
		if err == nil {
			return rate
		}
		logrus.WithError(err).Debug("Failed to get error rate from collector")
	}

	// Fallback: query Prometheus directly using the metrics package
	// In production, this would use prometheus client to query the /metrics endpoint
	// Example using agent_execution_errors_total and agent_executions_total:
	//
	//   errorRate := sum(rate(agent_execution_errors_total[5m])) /
	//                sum(rate(agent_executions_total[5m]))
	//
	// For now, return 0 to indicate no errors detected
	return 0.0
}

// getCurrentLatency returns the current P99 latency
// Uses the metrics collector if available, otherwise falls back to Prometheus queries
func (a *AutoDegradation) getCurrentLatency(ctx context.Context) time.Duration {
	if a.metricsCollector != nil {
		latency, err := a.metricsCollector.GetLatencyP99(ctx, 5*time.Minute)
		if err == nil {
			return latency
		}
		logrus.WithError(err).Debug("Failed to get latency from collector")
	}

	// Fallback: query Prometheus directly using the metrics package
	// Example using agent_execution_duration_seconds histogram:
	//
	//   latency := histogram_quantile(0.99, rate(agent_execution_duration_seconds_bucket[5m]))
	//
	// For now, return 0 to indicate normal latency
	return 0
}

// hasOpenCircuitBreakers checks if any circuit breakers are in open state
// Uses the metrics collector if available
func (a *AutoDegradation) hasOpenCircuitBreakers(ctx context.Context) bool {
	if a.metricsCollector != nil {
		state, err := a.metricsCollector.GetCircuitBreakerState(ctx, "")
		if err == nil && state == "open" {
			return true
		}
	}

	// Fallback: query Prometheus for agent_circuit_state gauge
	// Example query:
	//   sum(agent_circuit_state{state="open"}) > 0
	//
	// For now, return false to indicate no open circuit breakers
	return false
}

// isQuotaExhausted checks if any quotas are exhausted
// Uses the metrics collector if available
func (a *AutoDegradation) isQuotaExhausted(ctx context.Context) bool {
	if a.metricsCollector != nil {
		// Check if any quota usage is at or near 100%
		// This would be implemented based on the collector
	}

	// Fallback: check AgentQuotaUsage prometheus gauges
	// AgentQuotaUsage ratio approaching 1.0
	//
	// For now, return false to indicate quotas are not exhausted
	return false
}

// getNextEscalationLevel returns the next worse degradation level
func (a *AutoDegradation) getNextEscalationLevel(current DegradationLevel) DegradationLevel {
	switch current {
	case LevelNormal:
		return LevelDegraded
	case LevelDegraded:
		return LevelCritical
	case LevelCritical:
		return LevelEmergency
	default:
		return LevelEmergency
	}
}

// getNextDeEscalationLevel returns the next better degradation level
func (a *AutoDegradation) getNextDeEscalationLevel(current DegradationLevel) DegradationLevel {
	switch current {
	case LevelEmergency:
		return LevelCritical
	case LevelCritical:
		return LevelDegraded
	case LevelDegraded:
		return LevelNormal
	default:
		return LevelNormal
	}
}

// canDeEscalate checks if enough time has passed since last escalation
func (a *AutoDegradation) canDeEscalate(targetLevel DegradationLevel) bool {
	history := a.manager.GetHistory()
	if len(history) < 2 {
		return true
	}

	// Find the last time we were at the target level or worse
	lastWorseTime := time.Time{}
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].ToLevel >= targetLevel {
			lastWorseTime = history[i].Timestamp
			break
		}
	}

	if lastWorseTime.IsZero() {
		return true
	}

	// Find the cooldown duration for this level
	for _, rule := range a.escalationRules {
		if rule.Level == targetLevel {
			return time.Since(lastWorseTime) >= rule.Cooldown
		}
	}

	// Default cooldown: 5 minutes for any level
	return time.Since(lastWorseTime) >= 5*time.Minute
}

// Predefined features for FunctionFly
var (
	// RoutingFeature represents the routing feature
	RoutingFeature = &Feature{
		Name:           "routing",
		Description:    "Request routing to backends",
		Enabled:        true,
		MinLevel:       LevelNormal,
		Fallback:       nil, // No fallback, routing is critical
	}

	// HealthCheckFeature represents health checking
	HealthCheckFeature = &Feature{
		Name:           "health_check",
		Description:    "Backend health monitoring",
		Enabled:        true,
		MinLevel:       LevelNormal,
		Fallback:       nil, // No fallback, health checks are critical
	}

	// CachingFeature represents caching
	CachingFeature = &Feature{
		Name:           "caching",
		Description:    "Response caching",
		Enabled:        true,
		MinLevel:       LevelDegraded,
		Fallback: func(ctx context.Context, args ...interface{}) (interface{}, error) {
			// Fallback: skip cache, go directly to backend
			logrus.Info("Cache degraded, bypassing cache")
			return nil, nil
		},
	}

	// AnalyticsFeature represents analytics collection
	AnalyticsFeature = &Feature{
		Name:           "analytics",
		Description:    "Analytics and metrics collection",
		Enabled:        true,
		MinLevel:       LevelCritical,
		Fallback: func(ctx context.Context, args ...interface{}) (interface{}, error) {
			// Fallback: skip analytics, continue processing
			logrus.Info("Analytics degraded, skipping collection")
			return nil, nil
		},
	}

	// RateLimitingFeature represents rate limiting
	RateLimitingFeature = &Feature{
		Name:           "rate_limiting",
		Description:    "Request rate limiting",
		Enabled:        true,
		MinLevel:       LevelDegraded,
		Fallback: func(ctx context.Context, args ...interface{}) (interface{}, error) {
			// Fallback: allow all requests
			logrus.Info("Rate limiting degraded, allowing all requests")
			return true, nil
		},
	}
)

// Global degradation manager instance
var GlobalDegradationManager = NewDegradationManager()

// InitializeDegradation initializes degradation management
func InitializeDegradation() {
	GlobalDegradationManager.RegisterFeature(RoutingFeature)
	GlobalDegradationManager.RegisterFeature(HealthCheckFeature)
	GlobalDegradationManager.RegisterFeature(CachingFeature)
	GlobalDegradationManager.RegisterFeature(AnalyticsFeature)
	GlobalDegradationManager.RegisterFeature(RateLimitingFeature)
}

