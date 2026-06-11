package dna

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricsNamespace = "dna"
	metricsSubsystem = "analysis"

	analysisQueueDepthGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "queue",
			Name:      "depth",
			Help:      "Current number of analyses waiting in the queue",
		},
	)

	mutationsProposedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "mutations",
			Name:      "proposed_total",
			Help:      "Total number of mutations proposed",
		},
		[]string{"function_id", "tenant_id", "mutation_type"},
	)

	mutationsAcceptedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "mutations",
			Name:      "accepted_total",
			Help:      "Total number of mutations accepted",
		},
		[]string{"function_id", "tenant_id"},
	)

	mutationsRejectedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "mutations",
			Name:      "rejected_total",
			Help:      "Total number of mutations rejected",
		},
		[]string{"function_id", "tenant_id"},
	)

	analysisDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "duration_seconds",
			Help:      "DNA analysis duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
		},
		[]string{"function_id", "status"},
	)

	aiServiceCallsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "ai",
			Name:      "calls_total",
			Help:      "Total number of AI service calls",
		},
		[]string{"status"},
	)

	aiResponseTimeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "ai",
			Name:      "response_time_seconds",
			Help:      "AI service response time in seconds",
			Buckets:   []float64{0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{},
	)

	mutationValidationDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "validation",
			Name:      "duration_seconds",
			Help:      "Mutation validation duration in seconds",
			Buckets:   []float64{1, 5, 10, 30, 60, 120},
		},
		[]string{"status"},
	)

	executionMetricsInsertedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "execution_metrics",
			Name:      "inserted_total",
			Help:      "Total number of execution metrics records inserted",
		},
	)

	executionRecordingFailuresCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "execution_metrics",
			Name:      "recording_failures_total",
			Help:      "Total number of execution metric recording failures",
		},
	)

	dnaProfilesGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "profiles",
			Name:      "total",
			Help:      "Total number of DNA profiles",
		},
	)

	activeAnalysisWorkersGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "active_workers",
			Help:      "Number of currently active analysis workers",
		},
	)

	canaryDeploymentsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "canary",
			Name:      "deployments_total",
			Help:      "Total number of canary deployments triggered",
		},
		[]string{"function_id", "tenant_id", "status"},
	)

	rollbacksCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "rollbacks",
			Name:      "total",
			Help:      "Total number of mutation rollbacks",
		},
		[]string{"function_id", "tenant_id", "reason"},
	)

	circuitBreakerStateGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "circuit_breaker",
			Name:      "state",
			Help:      "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{},
	)

	circuitBreakerFailuresGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "circuit_breaker",
			Name:      "consecutive_failures",
			Help:      "Current consecutive failure count for early warning before breaker trips",
		},
	)

	circuitBreakerSuccessesGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "circuit_breaker",
			Name:      "consecutive_successes",
			Help:      "Current consecutive success count (useful during half-open recovery monitoring)",
		},
	)

	rateLimiterCountGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "rate_limiter",
			Name:      "entries",
			Help:      "Number of entries in the rate limiter map",
		},
	)

	partitionMaintenanceDurationHistogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "partition",
			Name:      "maintenance_duration_seconds",
			Help:      "Partition maintenance job duration in seconds",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300},
		},
	)

	insightsAggregationDurationHistogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: "insights",
			Name:      "aggregation_duration_seconds",
			Help:      "Insights aggregation job duration in seconds",
			Buckets:   []float64{10, 30, 60, 120, 300, 600},
		},
	)

	insightsAggregationTenantsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "insights",
			Name:      "tenants_processed",
			Help:      "Number of tenants processed in last insights aggregation",
		},
	)

	metricsMu sync.RWMutex
)

func RecordMutationProposed(functionID, tenantID, mutationType string) {
	mutationsProposedCounter.WithLabelValues(functionID, tenantID, mutationType).Inc()
}

func RecordMutationAccepted(functionID, tenantID string) {
	mutationsAcceptedCounter.WithLabelValues(functionID, tenantID).Inc()
}

func RecordMutationRejected(functionID, tenantID string) {
	mutationsRejectedCounter.WithLabelValues(functionID, tenantID).Inc()
}

func RecordAnalysisDuration(functionID, status string, duration time.Duration) {
	analysisDurationHistogram.WithLabelValues(functionID, status).Observe(duration.Seconds())
}

func RecordAIServiceCall(status string) {
	aiServiceCallsCounter.WithLabelValues(status).Inc()
}

func RecordAIResponseTime(duration time.Duration) {
	aiResponseTimeHistogram.WithLabelValues().Observe(duration.Seconds())
}

func RecordMutationValidation(status string, duration time.Duration) {
	mutationValidationDurationHistogram.WithLabelValues(status).Observe(duration.Seconds())
}

func RecordExecutionMetricsInserted() {
	executionMetricsInsertedCounter.Inc()
}

func RecordExecutionRecordingFailure() {
	executionRecordingFailuresCounter.Inc()
}

func SetQueueDepth(depth float64) {
	analysisQueueDepthGauge.Set(depth)
}

func SetDNAProfilesTotal(count float64) {
	dnaProfilesGauge.Set(count)
}

func SetActiveWorkers(count float64) {
	activeAnalysisWorkersGauge.Set(count)
}

func RecordCanaryDeployment(functionID, tenantID, status string) {
	canaryDeploymentsCounter.WithLabelValues(functionID, tenantID, status).Inc()
}

func RecordRollback(functionID, tenantID, reason string) {
	rollbacksCounter.WithLabelValues(functionID, tenantID, reason).Inc()
}

func SetCircuitBreakerState(state float64) {
	circuitBreakerStateGauge.WithLabelValues().Set(state)
}

func SetCircuitBreakerFailures(count float64) {
	circuitBreakerFailuresGauge.Set(count)
}

func SetCircuitBreakerSuccesses(count float64) {
	circuitBreakerSuccessesGauge.Set(count)
}

func SetRateLimiterEntries(count float64) {
	rateLimiterCountGauge.Set(count)
}

func RecordPartitionMaintenance(duration time.Duration) {
	partitionMaintenanceDurationHistogram.Observe(duration.Seconds())
}

func RecordInsightsAggregation(duration time.Duration, tenantsProcessed int) {
	insightsAggregationDurationHistogram.Observe(duration.Seconds())
	insightsAggregationTenantsGauge.Set(float64(tenantsProcessed))
}

type MetricsRecorder struct {
	functionID string
	tenantID   string
	startTime  time.Time
}

func NewMetricsRecorder(functionID, tenantID string) *MetricsRecorder {
	return &MetricsRecorder{
		functionID: functionID,
		tenantID:   tenantID,
		startTime:  time.Now(),
	}
}

func (m *MetricsRecorder) RecordAnalysis(status string) {
	duration := time.Since(m.startTime)
	RecordAnalysisDuration(m.functionID, status, duration)
}

func (m *MetricsRecorder) RecordMutationProposed(mutationType string) {
	RecordMutationProposed(m.functionID, m.tenantID, mutationType)
}

func (m *MetricsRecorder) RecordMutationAccepted() {
	RecordMutationAccepted(m.functionID, m.tenantID)
}

func (m *MetricsRecorder) RecordMutationRejected() {
	RecordMutationRejected(m.functionID, m.tenantID)
}

func (m *MetricsRecorder) RecordCanaryDeployment(status string) {
	RecordCanaryDeployment(m.functionID, m.tenantID, status)
}

func (m *MetricsRecorder) RecordRollback(reason string) {
	RecordRollback(m.functionID, m.tenantID, reason)
}
