package atlas

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ObservabilityRunsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "atlas_observability_runs_created_total",
			Help: "Total observability runs created",
		},
		[]string{"tenant_id", "agent_type", "status"},
	)

	ObservabilityEventsRecorded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "atlas_observability_events_recorded_total",
			Help: "Total observability events recorded",
		},
		[]string{"tenant_id", "event_kind"},
	)

	ObservabilityEventSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "atlas_observability_event_size_bytes",
			Help:    "Size of observability events in bytes",
			Buckets: prometheus.ExponentialBuckets(64, 2, 12),
		},
		[]string{"event_kind"},
	)

	ObservabilityRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "atlas_observability_run_duration_seconds",
			Help:    "Observability run duration in seconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12),
		},
		[]string{"agent_type"},
	)

	ObservabilityCostRecorded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "atlas_observability_cost_usd_total",
			Help: "Total cost recorded in observability events (USD)",
		},
		[]string{"tenant_id", "agent_type"},
	)

	ObservabilityTokensRecorded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "atlas_observability_tokens_total",
			Help: "Total tokens recorded in observability events",
		},
		[]string{"tenant_id", "token_type"},
	)

	ObservabilityErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "atlas_observability_errors_total",
			Help: "Total observability errors",
		},
		[]string{"tenant_id", "error_type"},
	)

	ObservabilitySamplingDecisions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "atlas_observability_sampling_decisions_total",
			Help: "Total sampling decisions",
		},
		[]string{"tenant_id", "decision"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(ObservabilityRunsCreated)
	prometheus.MustRegister(ObservabilityEventsRecorded)
	prometheus.MustRegister(ObservabilityEventSize)
	prometheus.MustRegister(ObservabilityRunDuration)
	prometheus.MustRegister(ObservabilityCostRecorded)
	prometheus.MustRegister(ObservabilityTokensRecorded)
	prometheus.MustRegister(ObservabilityErrors)
	prometheus.MustRegister(ObservabilitySamplingDecisions)
}
