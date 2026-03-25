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

	// ============================================
	// Rust-specific WASM metrics
	// ============================================

	// Rust compilation duration histogram - time to compile Rust to WASM
	rustCompilationDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "wasm_rust",
			Subsystem: "compilation",
			Name:      "duration_seconds",
			Help:      "Time to compile Rust to WASM in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
		},
		[]string{"tenant_id", "function_name", "compiler_version"},
	)

	// Rust instance count gauge - number of active Rust WASM instances
	rustInstanceCountGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "wasm_rust",
			Subsystem: "instances",
			Name:      "active_count",
			Help:      "Number of currently active Rust WASM instances",
		},
		[]string{"tenant_id", "function_name"},
	)

	// Rust execution duration histogram - function execution time
	rustExecutionDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "wasm_rust",
			Subsystem: "execution",
			Name:      "duration_seconds",
			Help:      "Rust WASM function execution time in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"tenant_id", "function_name", "status"},
	)

	// Rust memory usage gauge - memory consumption per instance
	rustMemoryUsageGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "wasm_rust",
			Subsystem: "memory",
			Name:      "usage_bytes",
			Help:      "Memory consumption per Rust WASM instance in bytes",
		},
		[]string{"tenant_id", "function_name", "instance_id"},
	)

	// Rust errors counter - compilation and runtime errors
	rustErrorsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "wasm_rust",
			Subsystem: "errors",
			Name:      "total",
			Help:      "Total number of Rust compilation and runtime errors",
		},
		[]string{"tenant_id", "function_name", "error_type"},
	)

	// Rust cold starts counter - cold start frequency
	rustColdStartsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "wasm_rust",
			Subsystem: "lifecycle",
			Name:      "cold_starts_total",
			Help:      "Total number of Rust WASM cold starts",
		},
		[]string{"tenant_id", "function_name"},
	)

	// Rust instance pool metrics
	rustPoolSizeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "wasm_rust",
			Subsystem: "pool",
			Name:      "size",
			Help:      "Total Rust WASM instance pool size",
		},
		[]string{"tenant_id", "function_name"},
	)

	rustPoolAvailableGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "wasm_rust",
			Subsystem: "pool",
			Name:      "available",
			Help:      "Number of available Rust WASM instances in pool",
		},
		[]string{"tenant_id", "function_name"},
	)

	rustPoolHitRateGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "wasm_rust",
			Subsystem: "pool",
			Name:      "hit_rate",
			Help:      "Rust WASM instance pool hit rate as a percentage",
		},
		[]string{"tenant_id", "function_name"},
	)

	// Rust compilation errors counter - specifically for compilation failures
	rustCompilationErrorsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "wasm_rust",
			Subsystem: "compilation",
			Name:      "errors_total",
			Help:      "Total number of Rust compilation errors",
		},
		[]string{"tenant_id", "function_name", "error_category"},
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

// ============================================
// Rust-specific metrics recording methods
// ============================================

// RecordRustCompilation records Rust compilation duration
func RecordRustCompilation(tenantID, functionName, compilerVersion string, duration time.Duration) {
	rustCompilationDurationHistogram.WithLabelValues(tenantID, functionName, compilerVersion).Observe(duration.Seconds())
}

// RecordRustInstanceCount records the number of active Rust WASM instances
func RecordRustInstanceCount(tenantID, functionName string, count int) {
	rustInstanceCountGauge.WithLabelValues(tenantID, functionName).Set(float64(count))
}

// RecordRustExecution records Rust WASM function execution duration
func RecordRustExecution(tenantID, functionName, status string, duration time.Duration) {
	rustExecutionDurationHistogram.WithLabelValues(tenantID, functionName, status).Observe(duration.Seconds())
}

// RecordRustMemoryUsage records memory usage for a Rust WASM instance
func RecordRustMemoryUsage(tenantID, functionName, instanceID string, bytes uint64) {
	rustMemoryUsageGauge.WithLabelValues(tenantID, functionName, instanceID).Set(float64(bytes))
}

// RecordRustError records a Rust compilation or runtime error
func RecordRustError(tenantID, functionName, errorType string) {
	rustErrorsCounter.WithLabelValues(tenantID, functionName, errorType).Inc()
}

// RecordRustColdStart records a Rust WASM cold start
func RecordRustColdStart(tenantID, functionName string) {
	rustColdStartsCounter.WithLabelValues(tenantID, functionName).Inc()
}

// UpdateRustPoolMetrics updates Rust instance pool metrics
func UpdateRustPoolMetrics(tenantID, functionName string, totalSize, availableCount int, hitRate float64) {
	rustPoolSizeGauge.WithLabelValues(tenantID, functionName).Set(float64(totalSize))
	rustPoolAvailableGauge.WithLabelValues(tenantID, functionName).Set(float64(availableCount))
	rustPoolHitRateGauge.WithLabelValues(tenantID, functionName).Set(hitRate)
}

// RecordRustCompilationError records a Rust compilation error
func RecordRustCompilationError(tenantID, functionName, errorCategory string) {
	rustCompilationErrorsCounter.WithLabelValues(tenantID, functionName, errorCategory).Inc()
}

// RustMetricsRecorder provides a dedicated recorder for Rust-specific metrics
type RustMetricsRecorder struct {
	tenantID        string
	functionName    string
	compilerVersion string
}

// NewRustMetricsRecorder creates a new Rust metrics recorder
func NewRustMetricsRecorder(tenantID, functionName, compilerVersion string) *RustMetricsRecorder {
	return &RustMetricsRecorder{
		tenantID:        tenantID,
		functionName:    functionName,
		compilerVersion: compilerVersion,
	}
}

// RecordCompilation records compilation duration
func (r *RustMetricsRecorder) RecordCompilation(duration time.Duration) {
	rustCompilationDurationHistogram.WithLabelValues(r.tenantID, r.functionName, r.compilerVersion).Observe(duration.Seconds())
}

// RecordExecution records execution duration with status
func (r *RustMetricsRecorder) RecordExecution(duration time.Duration, status string) {
	rustExecutionDurationHistogram.WithLabelValues(r.tenantID, r.functionName, status).Observe(duration.Seconds())
}

// RecordMemoryUsage records memory usage for an instance
func (r *RustMetricsRecorder) RecordMemoryUsage(instanceID string, bytes uint64) {
	rustMemoryUsageGauge.WithLabelValues(r.tenantID, r.functionName, instanceID).Set(float64(bytes))
}

// RecordError records an error
func (r *RustMetricsRecorder) RecordError(errorType string) {
	rustErrorsCounter.WithLabelValues(r.tenantID, r.functionName, errorType).Inc()
}

// RecordColdStart records a cold start
func (r *RustMetricsRecorder) RecordColdStart() {
	rustColdStartsCounter.WithLabelValues(r.tenantID, r.functionName).Inc()
}

// UpdatePoolMetrics updates pool metrics
func (r *RustMetricsRecorder) UpdatePoolMetrics(totalSize, availableCount int, hitRate float64) {
	rustPoolSizeGauge.WithLabelValues(r.tenantID, r.functionName).Set(float64(totalSize))
	rustPoolAvailableGauge.WithLabelValues(r.tenantID, r.functionName).Set(float64(availableCount))
	rustPoolHitRateGauge.WithLabelValues(r.tenantID, r.functionName).Set(hitRate)
}

// RecordCompilationError records a compilation error
func (r *RustMetricsRecorder) RecordCompilationError(errorCategory string) {
	rustCompilationErrorsCounter.WithLabelValues(r.tenantID, r.functionName, errorCategory).Inc()
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

