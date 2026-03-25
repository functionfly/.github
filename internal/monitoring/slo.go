package monitoring

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SLOConfig defines Service Level Objective configuration
type SLOConfig struct {
	// Name is the SLO name
	Name string
	// Target is the target percentage (0-100)
	Target float64
	// Window is the time window for SLO calculation
	Window time.Duration
	// Description describes what this SLO measures
	Description string
}

// SLI represents a Service Level Indicator
type SLI struct {
	// Name is the SLI name
	Name string
	// Value is the current value
	Value float64
	// Timestamp is when this was measured
	Timestamp time.Time
	// Labels are additional labels
	Labels map[string]string
}

// SLOTracker tracks SLO compliance
type SLOTracker struct {
	config     SLOConfig
	sliGauge   prometheus.Gauge
	errorBudget prometheus.Gauge
	compliance prometheus.Gauge
	mu         sync.RWMutex
	window     time.Time
	total      int64
	successes  int64
}

// NewSLOTracker creates a new SLO tracker
func NewSLOTracker(config SLOConfig) *SLOTracker {
	labels := prometheus.Labels{
		"slo_name": config.Name,
	}

	tracker := &SLOTracker{
		config: config,
		window: time.Now(),
		sliGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "functionfly_sli_current",
			Help:        config.Description,
			ConstLabels: labels,
		}),
		errorBudget: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "functionfly_slo_error_budget_remaining",
			Help:        "Remaining error budget percentage",
			ConstLabels: labels,
		}),
		compliance: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "functionfly_slo_compliance",
			Help:        "SLO compliance percentage",
			ConstLabels: labels,
		}),
	}

	return tracker
}

// RecordSuccess records a successful operation
func (t *SLOTracker) RecordSuccess() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.total++
	t.successes++
	t.updateMetrics()
}

// RecordFailure records a failed operation
func (t *SLOTracker) RecordFailure() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.total++
	t.updateMetrics()
}

// updateMetrics updates Prometheus metrics
func (t *SLOTracker) updateMetrics() {
	if t.total == 0 {
		return
	}

	// Calculate current SLI
	compliance := float64(t.successes) / float64(t.total) * 100
	t.compliance.Set(compliance)

	// Calculate error budget
	errorBudget := 100 - compliance
	t.errorBudget.Set(errorBudget)

	// Update SLI gauge
	t.sliGauge.Set(compliance)
}

// GetCompliance returns current SLO compliance
func (t *SLOTracker) GetCompliance() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.total == 0 {
		return 100.0
	}

	return float64(t.successes) / float64(t.total) * 100
}

// GetErrorBudget returns remaining error budget
func (t *SLOTracker) GetErrorBudget() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.total == 0 {
		return 100 - t.config.Target
	}

	compliance := float64(t.successes) / float64(t.total) * 100
	return 100 - compliance
}

// Reset resets the tracker for a new window
func (t *SLOTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.window = time.Now()
	t.total = 0
	t.successes = 0
	t.updateMetrics()
}

// SLOManager manages multiple SLO trackers
type SLOManager struct {
	trackers map[string]*SLOTracker
	mu       sync.RWMutex
}

// NewSLOManager creates a new SLO manager
func NewSLOManager() *SLOManager {
	return &SLOManager{
		trackers: make(map[string]*SLOTracker),
	}
}

// RegisterSLO registers a new SLO
func (m *SLOManager) RegisterSLO(config SLOConfig) *SLOTracker {
	m.mu.Lock()
	defer m.mu.Unlock()

	tracker := NewSLOTracker(config)
	m.trackers[config.Name] = tracker
	return tracker
}

// GetTracker returns a tracker by name
func (m *SLOManager) GetTracker(name string) *SLOTracker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.trackers[name]
}

// GetAllCompliance returns compliance for all SLOs
func (m *SLOManager) GetAllCompliance() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64)
	for name, tracker := range m.trackers {
		result[name] = tracker.GetCompliance()
	}

	return result
}

// Predefined SLO configurations
var (
	// AvailabilitySLO tracks API availability
	AvailabilitySLO = SLOConfig{
		Name:        "api_availability",
		Target:      99.9,
		Window:      30 * 24 * time.Hour, // 30 days
		Description: "API availability percentage",
	}

	// LatencySLO tracks P95 latency
	LatencySLO = SLOConfig{
		Name:        "p95_latency",
		Target:      99.0,
		Window:      30 * 24 * time.Hour,
		Description: "P95 latency under 500ms",
	}

	// ErrorRateSLO tracks error rate
	ErrorRateSLO = SLOConfig{
		Name:        "error_rate",
		Target:      99.5,
		Window:      30 * 24 * time.Hour,
		Description: "Error rate below 0.5%",
	}

	// RoutingSuccessSLO tracks routing success rate
	RoutingSuccessSLO = SLOConfig{
		Name:        "routing_success",
		Target:      99.9,
		Window:      30 * 24 * time.Hour,
		Description: "Routing decision success rate",
	}

	// HealthCheckSLO tracks health check success rate
	HealthCheckSLO = SLOConfig{
		Name:        "health_check_success",
		Target:      99.5,
		Window:      30 * 24 * time.Hour,
		Description: "Health check success rate",
	}
)

// Global SLO manager instance
var GlobalSLOManager = NewSLOManager()

// InitializeSLOs initializes all predefined SLOs
func InitializeSLOs() {
	GlobalSLOManager.RegisterSLO(AvailabilitySLO)
	GlobalSLOManager.RegisterSLO(LatencySLO)
	GlobalSLOManager.RegisterSLO(ErrorRateSLO)
	GlobalSLOManager.RegisterSLO(RoutingSuccessSLO)
	GlobalSLOManager.RegisterSLO(HealthCheckSLO)
}

// RecordAvailability records an availability event
func RecordAvailability(success bool) {
	tracker := GlobalSLOManager.GetTracker("api_availability")
	if tracker != nil {
		if success {
			tracker.RecordSuccess()
		} else {
			tracker.RecordFailure()
		}
	}
}

// RecordLatency records a latency event
func RecordLatency(latencyMs int, thresholdMs int) {
	tracker := GlobalSLOManager.GetTracker("p95_latency")
	if tracker != nil {
		if latencyMs <= thresholdMs {
			tracker.RecordSuccess()
		} else {
			tracker.RecordFailure()
		}
	}
}

// RecordErrorRate records an error rate event
func RecordErrorRate(success bool) {
	tracker := GlobalSLOManager.GetTracker("error_rate")
	if tracker != nil {
		if success {
			tracker.RecordSuccess()
		} else {
			tracker.RecordFailure()
		}
	}
}

// RecordRoutingSuccess records a routing success event
func RecordRoutingSuccess(success bool) {
	tracker := GlobalSLOManager.GetTracker("routing_success")
	if tracker != nil {
		if success {
			tracker.RecordSuccess()
		} else {
			tracker.RecordFailure()
		}
	}
}

// RecordHealthCheckSuccess records a health check success event
func RecordHealthCheckSuccess(success bool) {
	tracker := GlobalSLOManager.GetTracker("health_check_success")
	if tracker != nil {
		if success {
			tracker.RecordSuccess()
		} else {
			tracker.RecordFailure()
		}
	}
}
