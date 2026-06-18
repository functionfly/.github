package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RoutingDecisions = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "wasm_pool",
		Subsystem: "router",
		Name:      "routing_decisions_total",
		Help:      "Total routing decisions by target and reason.",
	}, []string{"decision", "reason"})

	ClientLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "wasm_pool",
		Subsystem: "client",
		Name:      "latency_seconds",
		Help:      "Client-side execute latency by target and runtime.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"target", "runtime"})

	BreakerStateGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "wasm_pool",
		Subsystem: "breaker",
		Name:      "state",
		Help:      "Circuit breaker state (0=closed, 1=open, 2=half_open).",
	}, []string{"state"})

	DryRunDivergences = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "wasm_pool",
		Subsystem: "dry_run",
		Name:      "divergences_total",
		Help:      "Dry-run divergences between local and external results, by field.",
	}, []string{"field"})
)

// init seeds the breaker gauge so dashboards can read it before any traffic.
func init() {
	for _, s := range []BreakerState{BreakerClosed, BreakerOpen, BreakerHalfOpen} {
		BreakerStateGauge.WithLabelValues(s.String()).Set(0)
	}
	BreakerStateGauge.WithLabelValues(BreakerClosed.String()).Set(1)
}
