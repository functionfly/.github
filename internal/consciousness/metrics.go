package consciousness

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricsNamespace = "functionfly"
	metricsSubsystem = "consciousness"

	analysisDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "analysis_duration_seconds",
			Help:      "Duration of consciousness analysis runs",
			Buckets:   []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"status"},
	)

	tenantsAnalyzedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "tenants_analyzed_total",
			Help:      "Total number of tenants analyzed",
		},
	)

	insightsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "insights_created_total",
			Help:      "Total insights created by category and severity",
		},
		[]string{"category", "severity"},
	)

	analyzerErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "analyzer_errors_total",
			Help:      "Total analyzer errors by analyzer name",
		},
		[]string{"analyzer"},
	)

	analyzerDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "analyzer_duration_seconds",
			Help:      "Duration of individual analyzer runs",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"analyzer"},
	)

	dispatchTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "dispatch_total",
			Help:      "Total notification dispatches by channel and status",
		},
		[]string{"channel", "status"},
	)

	dispatchRetryTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "dispatch_retry_total",
			Help:      "Total dispatch retry attempts",
		},
	)

	dispatchDeadLetterTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "dispatch_dead_letter_total",
			Help:      "Total dispatches sent to dead letter queue",
		},
	)

	activeInsightsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "active_insights_gauge",
			Help:      "Current number of active insights by category and severity",
		},
		[]string{"category", "severity"},
	)

	scoreGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "score_gauge",
			Help:      "Current system awareness score components",
		},
		[]string{"tenant_id", "component"},
	)

	concurrencyGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "analysis_concurrency",
			Help:      "Current number of concurrent tenant analyses",
		},
	)

	circuitBreakerOpenTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      "circuit_breaker_open_total",
			Help:      "Total times an analyzer circuit breaker opened",
		},
		[]string{"analyzer"},
	)

	schedulerRunTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: "scheduler",
			Name:      "run_total",
			Help:      "Total scheduler runs by status",
		},
		[]string{"status"},
	)

	schedulerConfigInvalid = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: "scheduler",
			Name:      "config_invalid",
			Help:      "Whether consciousness scheduler config is invalid (1=invalid, 0=valid)",
		},
	)
)
