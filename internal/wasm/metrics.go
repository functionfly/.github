// Package wasm provides WebAssembly runtime support for FunctionFly
// This file contains Prometheus metrics implementation
package wasm

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Metrics namespace and subsystem
	metricsNamespace = "wasm"
	metricsSubsystem = "execution"

	// Execution duration histogram
	executionDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "duration_seconds",
			Help:      "WASM execution duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"runtime", "status", "tenant_id"},
	)

	// Instance pool size gauge
	instancePoolSizeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "instance_pool",
			Name:      "size",
			Help:      "Total instance pool size",
		},
		[]string{"tenant_id", "runtime"},
	)

	// Instance pool available gauge
	instancePoolAvailableGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "instance_pool",
			Name:      "available",
			Help:      "Number of available instances in pool",
		},
		[]string{"tenant_id", "runtime"},
	)

	// Cold starts counter
	coldStartsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "cold_starts_total",
			Help:      "Total number of cold starts (new instances created)",
		},
		[]string{"runtime", "tenant_id"},
	)

	// Memory usage gauge
	memoryUsageGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes",
		},
		[]string{"runtime", "tenant_id", "instance_id"},
	)

	// Execution errors counter
	executionErrorsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "errors_total",
			Help:      "Total number of execution errors",
		},
		[]string{"runtime", "error_type", "tenant_id"},
	)

	// Input size histogram
	inputSizeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "input_size_bytes",
			Help:      "WASM input size in bytes",
			Buckets:   []float64{1024, 4096, 16384, 65536, 262144, 1048576, 4194304},
		},
		[]string{"runtime", "tenant_id"},
	)

	// Output size histogram
	outputSizeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "output_size_bytes",
			Help:      "WASM output size in bytes",
			Buckets:   []float64{1024, 4096, 16384, 65536, 262144, 1048576, 4194304},
		},
		[]string{"runtime", "tenant_id"},
	)

	// Pool hit rate gauge
	poolHitRateGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "instance_pool",
			Name:      "hit_rate",
			Help:      "Pool hit rate as a percentage",
		},
		[]string{"tenant_id", "runtime"},
	)

	// Active instances gauge
	activeInstancesGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "instance_pool",
			Name:      "active_instances",
			Help:      "Number of currently active instances",
		},
		[]string{"tenant_id", "runtime"},
	)

	// Deterministic execution counter
	deterministicExecutionsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "deterministic_total",
			Help:      "Total number of deterministic executions",
		},
		[]string{"runtime", "tenant_id"},
	)

	// Streaming executions counter
	streamingExecutionsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "streaming_total",
			Help:      "Total number of streaming executions",
		},
		[]string{"runtime", "tenant_id"},
	)

	// Global metrics mutex for thread-safe updates
	metricsMu sync.RWMutex
)

// MetricsRecorder provides methods to record WASM execution metrics
type MetricsRecorder struct {
	runtime  string
	tenantID string
	startTime time.Time
}

// NewMetricsRecorder creates a new metrics recorder
func NewMetricsRecorder(runtime, tenantID string) *MetricsRecorder {
	return &MetricsRecorder{
		runtime:  runtime,
		tenantID: tenantID,
	}
}

// RecordExecution records an execution with duration and status
func (m *MetricsRecorder) RecordExecution(duration time.Duration, status string) {
	executionDurationHistogram.WithLabelValues(m.runtime, status, m.tenantID).Observe(duration.Seconds())
}

// RecordInputSize records input size
func (m *MetricsRecorder) RecordInputSize(size int) {
	inputSizeHistogram.WithLabelValues(m.runtime, m.tenantID).Observe(float64(size))
}

// RecordOutputSize records output size
func (m *MetricsRecorder) RecordOutputSize(size int) {
	outputSizeHistogram.WithLabelValues(m.runtime, m.tenantID).Observe(float64(size))
}

// RecordError records an execution error
func (m *MetricsRecorder) RecordError(errorType string) {
	executionErrorsCounter.WithLabelValues(m.runtime, errorType, m.tenantID).Inc()
}

// RecordColdStart records a cold start
func (m *MetricsRecorder) RecordColdStart() {
	coldStartsCounter.WithLabelValues(m.runtime, m.tenantID).Inc()
}

// RecordMemoryUsage records memory usage for an instance
func (m *MetricsRecorder) RecordMemoryUsage(instanceID string, bytes uint64) {
	memoryUsageGauge.WithLabelValues(m.runtime, m.tenantID, instanceID).Set(float64(bytes))
}

// RecordDeterministicExecution records a deterministic execution
func (m *MetricsRecorder) RecordDeterministicExecution() {
	deterministicExecutionsCounter.WithLabelValues(m.runtime, m.tenantID).Inc()
}

// RecordStreamingExecution records a streaming execution
func (m *MetricsRecorder) RecordStreamingExecution() {
	streamingExecutionsCounter.WithLabelValues(m.runtime, m.tenantID).Inc()
}

// UpdatePoolMetrics updates pool-related metrics
func UpdatePoolMetrics(tenantID, runtime string, totalSize, availableCount, activeCount int, hitRate float64) {
	instancePoolSizeGauge.WithLabelValues(tenantID, runtime).Set(float64(totalSize))
	instancePoolAvailableGauge.WithLabelValues(tenantID, runtime).Set(float64(availableCount))
	activeInstancesGauge.WithLabelValues(tenantID, runtime).Set(float64(activeCount))
	poolHitRateGauge.WithLabelValues(tenantID, runtime).Set(hitRate)
}

// RecordExecutionWithSizes records execution metrics with input/output sizes
func (m *MetricsRecorder) RecordExecutionWithSizes(duration time.Duration, status string, inputSize, outputSize int) {
	m.RecordExecution(duration, status)
	if inputSize > 0 {
		m.RecordInputSize(inputSize)
	}
	if outputSize > 0 {
		m.RecordOutputSize(outputSize)
	}
}

// WasmMetricsCollector implements prometheus.Collector for custom metrics
type WasmMetricsCollector struct {
	pool *InstancePool
}

// NewWasmMetricsCollector creates a new metrics collector
func NewWasmMetricsCollector(pool *InstancePool) *WasmMetricsCollector {
	return &WasmMetricsCollector{pool: pool}
}

// Describe implements prometheus.Collector
func (c *WasmMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	// Already registered via promauto
}

// Collect implements prometheus.Collector
func (c *WasmMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	if c.pool == nil || c.pool.metrics == nil {
		return
	}

	// Collect pool metrics
	metrics := c.pool.metrics

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(metricsNamespace, "instance_pool", "hits_total"),
			"Counter",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(metrics.GetHits()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(metricsNamespace, "instance_pool", "misses_total"),
			"Counter",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(metrics.GetMisses()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(metricsNamespace, "instance_pool", "cold_starts_total"),
			"Counter",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(metrics.GetColdStarts()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(metricsNamespace, "instance_pool", "evictions_total"),
			"Counter",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(metrics.GetEvictions()),
	)
}

// ParseErrorType parses an error message into a categorized error type
func ParseErrorType(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	switch {
	case containsStr(errStr, "timeout"):
		return "timeout"
	case containsStr(errStr, "memory"):
		return "memory_limit"
	case containsStr(errStr, "input"):
		return "invalid_input"
	case containsStr(errStr, "instantiate"):
		return "instantiation_error"
	case containsStr(errStr, "execution"):
		return "execution_error"
	default:
		return "unknown"
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsAtStr(s, substr))
}

func containsAtStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Int64ToFloat converts int64 to float64 safely
func Int64ToFloat(val int64) float64 {
	return float64(val)
}

// Uint64ToFloat converts uint64 to float64 safely
func Uint64ToFloat(val uint64) float64 {
	return float64(val)
}

// StringToFloat64 safely converts string to float64
func StringToFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
