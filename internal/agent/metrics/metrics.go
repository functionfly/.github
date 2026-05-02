package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// AgentExecutions tracks total agent executions
	AgentExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_executions_total",
			Help: "Total agent executions",
		},
		[]string{"agent_id", "function_uri", "outcome"},
	)

	// AgentExecutionDuration tracks agent execution duration
	AgentExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agent_execution_duration_seconds",
			Help:    "Agent execution duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"agent_id", "function_uri"},
	)

	// AgentExecutionErrors tracks agent execution errors
	AgentExecutionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_execution_errors_total",
			Help: "Total agent execution errors",
		},
		[]string{"agent_id", "function_uri", "error_code"},
	)

	// AgentQuotaUsage tracks agent quota usage ratio
	AgentQuotaUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_quota_usage_ratio",
			Help: "Agent quota usage ratio (0-1)",
		},
		[]string{"agent_id", "quota_type"},
	)

	// AgentQuotaViolations tracks total quota violations
	AgentQuotaViolations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_quota_violations_total",
			Help: "Total quota violations",
		},
		[]string{"agent_id", "quota_type"},
	)

	// AgentPolicyViolations tracks total policy violations
	AgentPolicyViolations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_policy_violations_total",
			Help: "Total policy violations",
		},
		[]string{"agent_id", "violation_code"},
	)

	// AgentCostUSD tracks total cost in USD
	AgentCostUSD = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_cost_usd_total",
			Help: "Total cost in USD",
		},
		[]string{"agent_id", "function_uri"},
	)

	// AgentConcurrencyActive tracks active concurrent executions
	AgentConcurrencyActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_concurrency_active",
			Help: "Active concurrent executions",
		},
		[]string{"agent_id"},
	)

	// AgentConcurrencyLimit tracks concurrency limit
	AgentConcurrencyLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_concurrency_limit",
			Help: "Concurrency limit",
		},
		[]string{"agent_id"},
	)

	// AgentCircuitState tracks circuit breaker state
	AgentCircuitState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_circuit_state",
			Help: "Agent circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"agent_id"},
	)

	// AgentCircuitTransitions tracks circuit breaker state transitions
	AgentCircuitTransitions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_circuit_transitions_total",
			Help: "Total circuit breaker state transitions",
		},
		[]string{"agent_id", "from", "to"},
	)

	// AgentRetryAttempts tracks retry attempts
	AgentRetryAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_retry_attempts_total",
			Help: "Total retry attempts",
		},
		[]string{"agent_id", "function_uri"},
	)

	// AgentRetrySuccesses tracks successful retries
	AgentRetrySuccesses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_retry_successes_total",
			Help: "Total successful retries",
		},
		[]string{"agent_id", "function_uri"},
	)

	// AgentDeadLetterTotal tracks dead letter queue entries
	AgentDeadLetterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_dead_letter_total",
			Help: "Total dead letter entries created",
		},
		[]string{"agent_id", "function_uri", "error_code"},
	)

	// AgentDeadLetterPending tracks pending dead letter entries
	AgentDeadLetterPending = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_dead_letter_pending",
			Help: "Current pending dead letter entries",
		},
		[]string{"agent_id"},
	)

	// AgentDeadLetterRetryStorm tracks potential retry storms
	AgentDeadLetterRetryStorm = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_dead_letter_retry_storm_total",
			Help: "Dead letter entries that triggered retry storm alerts",
		},
		[]string{"agent_id", "function_uri"},
	)

	// AgentDeadLetterAttemptsHistogram tracks distribution of attempts before dead letter
	AgentDeadLetterAttemptsHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agent_dead_letter_attempts_histogram",
			Help:    "Distribution of attempts before dead letter",
			Buckets: []float64{1, 2, 3, 4, 5, 10, 20, 50},
		},
		[]string{"agent_id", "function_uri"},
	)

	// AgentDeadLetterRetryOutcome tracks retry outcomes for dead letter entries
	AgentDeadLetterRetryOutcome = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_dead_letter_retry_outcome_total",
			Help: "Dead letter retry outcomes",
		},
		[]string{"agent_id", "outcome"},
	)
)

func init() {
	prometheus.MustRegister(
		AgentExecutions,
		AgentExecutionDuration,
		AgentExecutionErrors,
		AgentQuotaUsage,
		AgentQuotaViolations,
		AgentPolicyViolations,
		AgentCostUSD,
		AgentConcurrencyActive,
		AgentConcurrencyLimit,
		AgentCircuitState,
		AgentCircuitTransitions,
		AgentRetryAttempts,
		AgentRetrySuccesses,
		AgentDeadLetterTotal,
		AgentDeadLetterPending,
		AgentDeadLetterRetryStorm,
		AgentDeadLetterAttemptsHistogram,
		AgentDeadLetterRetryOutcome,
	)
}
