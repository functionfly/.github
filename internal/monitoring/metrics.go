package monitoring

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	registry = prometheus.NewRegistry()
)

// Custom metrics for FunctionFly performance monitoring
var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "endpoint", "status_code", "backend_provider", "region"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "backend_provider", "region"},
	)

	// Routing metrics
	routingDecisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_routing_decisions_total",
			Help: "Total number of routing decisions made",
		},
		[]string{"decision_type", "backend_provider", "region", "plan"},
	)

	routingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_routing_latency_seconds",
			Help:    "Time taken to make routing decisions",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"backend_count", "circuit_breaker_state"},
	)

	backendHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_backend_health_status",
			Help: "Current health status of backends (1=healthy, 0=unhealthy)",
		},
		[]string{"backend_id", "provider", "region"},
	)

	ewmaLatencyScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_backend_ewma_latency_score",
			Help: "Exponentially weighted moving average latency score for backends",
		},
		[]string{"backend_id", "provider", "region"},
	)

	// Circuit breaker metrics
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_circuit_breaker_state",
			Help: "Current state of circuit breakers (0=closed, 1=open, 2=half-open)",
		},
		[]string{"backend_id", "provider", "region"},
	)

	circuitBreakerTransitionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_circuit_breaker_transitions_total",
			Help: "Total number of circuit breaker state transitions",
		},
		[]string{"backend_id", "from_state", "to_state"},
	)

	// Uptime metrics
	uptimeRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_uptime_ratio",
			Help: "Uptime ratio for components and providers (0.0 to 1.0)",
		},
		[]string{"component", "provider"},
	)

	cacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_cache_size",
			Help: "Current size of caches",
		},
		[]string{"cache_type"},
	)

	redisConnectionPoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_redis_connection_pool_size",
			Help: "Size of Redis connection pools",
		},
		[]string{"pool_name"},
	)

	// Database metrics
	dbConnectionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_db_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"db_name"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"query_type", "table_name"},
	)

	dbTransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_db_transaction_duration_seconds",
			Help:    "Database transaction duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"transaction_type"},
	)

	// Function execution metrics
	functionExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_function_execution_duration_seconds",
			Help:    "Function execution duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"function_id", "runtime", "provider", "region"},
	)

	functionExecutionErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_function_execution_errors_total",
			Help: "Total number of function execution errors",
		},
		[]string{"function_id", "runtime", "provider", "region", "error_type"},
	)

	functionInvocationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_function_invocations_total",
			Help: "Total number of function invocations",
		},
		[]string{"function_id", "runtime", "provider", "region", "trigger"},
	)

	// Resource usage metrics
	memoryUsageBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
		[]string{"service", "type"}, // heap, stack, system
	)

	goroutinesCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_goroutines_count",
			Help: "Current number of goroutines",
		},
		[]string{"service"},
	)

	// Queue and async processing metrics
	queueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_queue_size",
			Help: "Current size of processing queues",
		},
		[]string{"queue_name"},
	)

	queueProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_queue_processing_duration_seconds",
			Help:    "Time spent processing queue items",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"queue_name", "operation"},
	)

	// Alert and incident metrics
	alertsTriggeredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_alerts_triggered_total",
			Help: "Total number of alerts triggered",
		},
		[]string{"alert_type", "severity", "component"},
	)

	activeIncidents = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_active_incidents",
			Help: "Number of currently active incidents",
		},
		[]string{"severity", "component"},
	)

	// Business metrics
	userSessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_user_sessions_active",
			Help: "Number of currently active user sessions",
		},
	)

	functionDeploymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_function_deployments_total",
			Help: "Total number of function deployments",
		},
		[]string{"status", "provider", "region"},
	)

	// Performance test metrics (for load testing)
	loadTestActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_load_test_active",
			Help: "Whether a load test is currently running (1=yes, 0=no)",
		},
	)

	loadTestVirtualUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_load_test_virtual_users",
			Help: "Number of virtual users in current load test",
		},
	)

	// Local runtime metrics
	localRuntimeMemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_local_runtime_memory_usage_bytes",
			Help: "Memory usage of local runtimes",
		},
		[]string{"runtime_type", "memory_type"}, // vms, rss, data
	)

	localRuntimeCPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_local_runtime_cpu_usage_percent",
			Help: "CPU usage percentage of local runtimes",
		},
		[]string{"runtime_type"},
	)

	localRuntimeRequestThroughput = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_local_runtime_request_throughput_per_second",
			Help: "Request throughput of local runtimes",
		},
		[]string{"runtime_type"},
	)

	localRuntimeActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_local_runtime_active_connections",
			Help: "Number of active connections for local runtimes",
		},
		[]string{"runtime_type", "port"},
	)

	localRuntimeExecutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_local_runtime_execution_total",
			Help: "Total number of function executions in local runtimes",
		},
		[]string{"runtime_type", "status", "function_name"},
	)

	localRuntimeExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_local_runtime_execution_duration_seconds",
			Help:    "Execution duration of functions in local runtimes",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"runtime_type", "function_name"},
	)

	// State and state fabric metrics
	// State operation metrics (per tenant)
	stateOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_state_operations_total",
			Help: "Total number of state operations",
		},
		[]string{"tenant_id", "operation_type", "status"},
	)

	stateOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_state_operation_duration_seconds",
			Help:    "State operation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"tenant_id", "operation_type"},
	)

	stateStorageSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_state_storage_size_bytes",
			Help: "Current storage size for state in bytes",
		},
		[]string{"tenant_id", "state_path"},
	)

	// State fabric metrics
	stateFabricOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_state_fabric_operations_total",
			Help: "Total number of state fabric operations",
		},
		[]string{"tenant_id", "fabric_id", "operation_type", "status"},
	)

	stateFabricOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_state_fabric_operation_duration_seconds",
			Help:    "State fabric operation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"tenant_id", "fabric_id", "operation_type"},
	)

	stateFabricActiveCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_state_fabric_active_count",
			Help: "Number of active state fabrics",
		},
		[]string{"tenant_id"},
	)

	stateFabricStoreCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_state_fabric_stores_count",
			Help: "Number of stores in a state fabric",
		},
		[]string{"tenant_id", "fabric_id"},
	)

	stateFabricPipelineExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_state_fabric_pipeline_executions_total",
			Help: "Total number of state fabric pipeline executions",
		},
		[]string{"tenant_id", "fabric_id", "pipeline_id", "status"},
	)

	stateFabricPipelineDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_state_fabric_pipeline_duration_seconds",
			Help:    "State fabric pipeline execution duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"tenant_id", "fabric_id", "pipeline_id"},
	)

	// Trigger execution metrics
	triggerExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_trigger_executions_total",
			Help: "Total number of trigger executions",
		},
		[]string{"tenant_id", "state_path", "trigger_type", "status"},
	)

	triggerExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_trigger_execution_duration_seconds",
			Help:    "Trigger execution duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"tenant_id", "state_path", "trigger_type"},
	)

	triggerErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_trigger_errors_total",
			Help: "Total number of trigger execution errors",
		},
		[]string{"tenant_id", "state_path", "trigger_type", "error_type"},
	)

	// FRG Graph execution metrics
	frgGraphExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_frg_graph_executions_total",
			Help: "Total number of FRG graph executions",
		},
		[]string{"tenant_id", "graph_id", "operation_type", "status"},
	)

	frgGraphExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_frg_graph_execution_duration_seconds",
			Help:    "FRG graph execution duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"tenant_id", "graph_id"},
	)

	frgGraphActiveCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_frg_graph_active_count",
			Help: "Number of active FRG graph executions",
		},
		[]string{"tenant_id"},
	)

	// FRG Quota metrics
	frgQuotaExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_frg_quota_exceeded_total",
			Help: "Total number of FRG graph executions blocked due to quota exceeded",
		},
		[]string{"tenant_id"},
	)

	frgQuotaUsagePercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_frg_quota_usage_percent",
			Help: "Current FRG quota usage percentage",
		},
		[]string{"tenant_id"},
	)

	frgWebhookSignatureFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_frg_webhook_signature_failures_total",
			Help: "Total number of FRG webhook signature verification failures",
		},
		[]string{"reason"},
	)

	frgGraphCreationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_frg_graph_creation_total",
			Help: "Total number of FRG graph definitions created",
		},
		[]string{"tenant_id", "visibility", "status"},
	)

	frgGraphNodesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_frg_graph_nodes_total",
			Help: "Total number of nodes across all FRG graphs",
		},
		[]string{"tenant_id"},
	)

	// Event and snapshot metrics
	stateEventCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_state_events_total",
			Help: "Total number of state events",
		},
		[]string{"tenant_id", "fabric_id", "event_type"},
	)

	stateSnapshotCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_state_snapshots_count",
			Help: "Number of state snapshots",
		},
		[]string{"tenant_id", "fabric_id"},
	)

	localRuntimeErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_local_runtime_errors_total",
			Help: "Total number of errors in local runtimes",
		},
		[]string{"runtime_type", "error_type", "function_name"},
	)

	// Edge (edge.functionfly.com) monitoring metrics
	edgeHealthStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_edge_health_status",
			Help: "Edge health from system probe (1=healthy, 0=unhealthy)",
		},
	)
	edgeProbeLatencySeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_edge_probe_latency_seconds",
			Help: "Edge health probe latency in seconds",
		},
	)
	edgeProbeErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_edge_probe_errors_total",
			Help: "Total number of edge health probe failures",
		},
	)
	edgeRequestsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_edge_requests_total",
			Help: "Total number of requests routed to FunctionFly Edge",
		},
	)
	edgeUptimeRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "functionfly_edge_uptime_ratio",
			Help: "Edge uptime ratio (0.0 to 1.0) from recent probes",
		},
	)

	// Billing and payment metrics
	billingCheckoutCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_checkout_created_total",
			Help: "Total number of checkout sessions created",
		},
		[]string{"checkout_type"}, // subscription, wallet_topup, addon, agent_credits
	)

	billingCheckoutCompletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_checkout_completed_total",
			Help: "Total number of successful checkout completions",
		},
		[]string{"checkout_type"},
	)

	billingCheckoutFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_checkout_failed_total",
			Help: "Total number of failed checkout attempts",
		},
		[]string{"checkout_type", "failure_reason"},
	)

	billingCheckoutDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_billing_checkout_duration_seconds",
			Help:    "Time from checkout creation to completion",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200}, // 1min to 2hrs
		},
		[]string{"checkout_type"},
	)

	billingWebhookReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_webhook_received_total",
			Help: "Total number of webhooks received from Stripe",
		},
		[]string{"event_type", "status"}, // status: success, error, ignored
	)

	billingWebhookProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_billing_webhook_processing_duration_seconds",
			Help:    "Time to process a webhook event",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"event_type"},
	)

	billingPaymentFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_payment_failed_total",
			Help: "Total number of failed payments",
		},
		[]string{"failure_type", "decline_code"},
	)

	billingSubscriptionStatusChangesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_subscription_status_changes_total",
			Help: "Total number of subscription status changes",
		},
		[]string{"from_status", "to_status"},
	)

	billingWalletBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_billing_wallet_balance_usd",
			Help: "Current wallet balance in USD",
		},
		[]string{"user_id", "tenant_id"},
	)

	billingPortalSessionsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "functionfly_billing_portal_sessions_created_total",
			Help: "Total number of customer portal sessions created",
		},
	)

	// Billing operational readiness alerts
	billingWebhookLatencyExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_webhook_latency_exceeded_total",
			Help: "Total number of times webhook processing exceeded threshold (5s)",
		},
		[]string{"event_type"},
	)

	billingWebhookSignatureFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_webhook_signature_failures_total",
			Help: "Total number of webhook signature verification failures",
		},
		[]string{"reason"},
	)

	billingDunningRetriesStuck = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_billing_dunning_retries_stuck",
			Help: "Number of dunning retries stuck (not processing on schedule)",
		},
		[]string{"retry_status"},
	)

	billingInvoiceGenerationFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_invoice_generation_failures_total",
			Help: "Total number of invoice generation failures",
		},
		[]string{"failure_reason"},
	)

	billingStripeAPIErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_stripe_api_errors_total",
			Help: "Total number of Stripe API errors",
		},
		[]string{"operation", "error_code"},
	)

	billingAlertTriggered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_billing_alert_triggered_total",
			Help: "Total number of billing alerts triggered",
		},
		[]string{"alert_type", "severity"},
	)

	// Execution log retention metrics
	executionLogCleanupRecordsDeleted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_execution_log_cleanup_records_deleted_total",
			Help: "Total number of execution log records deleted by cleanup",
		},
		[]string{"table_name"},
	)

	executionLogCleanupDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_execution_log_cleanup_duration_seconds",
			Help:    "Duration of execution log cleanup operations",
			Buckets: []float64{1.0, 5.0, 10.0, 30.0, 60.0, 120.0, 300.0},
		},
		[]string{"result"}, // success, error
	)

	executionLogCleanupErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_execution_log_cleanup_errors_total",
			Help: "Total number of execution log cleanup errors",
		},
		[]string{"table_name", "error_type"},
	)

	executionLogRetentionAge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_execution_log_retention_age_days",
			Help: "Configured retention age in days for each table",
		},
		[]string{"table_name"},
	)

	executionLogTableRecords = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_execution_log_table_records",
			Help: "Number of records in execution log tables (from stats queries)",
		},
		[]string{"table_name", "age_range"}, // age_range: total, older_than_30d, older_than_90d, older_than_365d
	)

	// Team Memory metrics (memory extraction and agent context)
	teamMemoryExtractionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_team_memory_extractions_total",
			Help: "Total number of memory extraction attempts from conversations",
		},
		[]string{"team_id", "memory_type", "status"}, // status: success, failed, rejected
	)

	teamMemoryExtractionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_team_memory_extraction_duration_seconds",
			Help:    "Time spent extracting memories from a conversation",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"team_id", "source"}, // source: ai_service, fallback, manual
	)

	teamMemoryExtractionConfidence = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_team_memory_extraction_confidence",
			Help:    "Confidence score distribution for extracted memories",
			Buckets: []float64{0.5, 0.6, 0.7, 0.8, 0.85, 0.9, 0.95, 1.0},
		},
		[]string{"team_id", "memory_type"},
	)

	teamMemoriesCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_team_memories_created_total",
			Help: "Total number of team memories created",
		},
		[]string{"team_id", "memory_type", "source"}, // source: auto_extraction, manual, template
	)

	teamMemoryContextInjectionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_team_memory_context_injections_total",
			Help: "Total number of times team memory context was injected into agent prompts",
		},
		[]string{"team_id"},
	)

	teamMemoryContextInjectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_team_memory_context_injection_duration_seconds",
			Help:    "Time spent building and injecting team memory context",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"team_id"},
	)

	teamMemorySearchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "functionfly_team_memory_search_duration_seconds",
			Help:    "Time spent searching team memories (vector similarity)",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"team_id", "search_type"}, // search_type: vector, text, hybrid
	)

	teamMemoriesActiveGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "functionfly_team_memories_active",
			Help: "Current number of active (non-expired) team memories",
		},
		[]string{"team_id", "memory_type", "validated"}, // validated: true, false
	)

	teamMemoryCacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_team_memory_cache_hits_total",
			Help: "Total number of team memory context cache hits",
		},
		[]string{"team_id"},
	)

	teamMemoryCacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "functionfly_team_memory_cache_misses_total",
			Help: "Total number of team memory context cache misses",
		},
		[]string{"team_id"},
	)
)

// Metric recording functions

// RecordHTTPMetrics records HTTP request metrics
func RecordHTTPMetrics(method, endpoint, statusCode, backendProvider, region string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, backendProvider, region).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint, backendProvider, region).Observe(duration.Seconds())
}

// RecordRoutingDecision records routing decision metrics
func RecordRoutingDecision(decisionType, backendProvider, region, plan string) {
	routingDecisionsTotal.WithLabelValues(decisionType, backendProvider, region, plan).Inc()
}

// RecordRoutingLatency records routing latency metrics
func RecordRoutingLatency(backendCount int, circuitBreakerState string, duration time.Duration) {
	routingLatency.WithLabelValues(string(rune(backendCount+'0')), circuitBreakerState).Observe(duration.Seconds())
}

// UpdateBackendHealth updates backend health status
func UpdateBackendHealth(backendID, provider, region string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	backendHealthStatus.WithLabelValues(backendID, provider, region).Set(value)
}

// UpdateEWMALatencyScore updates EWMA latency score for a backend
func UpdateEWMALatencyScore(backendID, provider, region string, score float64) {
	ewmaLatencyScore.WithLabelValues(backendID, provider, region).Set(score)
}

// UpdateCircuitBreakerState updates circuit breaker state
func UpdateCircuitBreakerState(backendID, provider, region string, state int) {
	circuitBreakerState.WithLabelValues(backendID, provider, region).Set(float64(state))
}

// RecordCircuitBreakerTransition records circuit breaker state transitions
func RecordCircuitBreakerTransition(backendID, fromState, toState string) {
	circuitBreakerTransitionsTotal.WithLabelValues(backendID, fromState, toState).Inc()
}

// UpdateUptimeRatio updates uptime ratio for components and providers
func UpdateUptimeRatio(component, provider string, ratio float64) {
	uptimeRatio.WithLabelValues(component, provider).Set(ratio)
}

// UpdateCacheSize updates cache size metrics
func UpdateCacheSize(cacheType string, size int) {
	cacheSize.WithLabelValues(cacheType).Set(float64(size))
}

// UpdateRedisPoolSize updates Redis connection pool size
func UpdateRedisPoolSize(poolName string, size int) {
	redisConnectionPoolSize.WithLabelValues(poolName).Set(float64(size))
}

// UpdateDBConnections updates database connection metrics
func UpdateDBConnections(dbName string, activeConnections int) {
	dbConnectionsActive.WithLabelValues(dbName).Set(float64(activeConnections))
}

// RecordDBQueryDuration records database query duration
func RecordDBQueryDuration(queryType, tableName string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(queryType, tableName).Observe(duration.Seconds())
}

// RecordDBTransactionDuration records database transaction duration
func RecordDBTransactionDuration(transactionType string, duration time.Duration) {
	dbTransactionDuration.WithLabelValues(transactionType).Observe(duration.Seconds())
}

// RecordFunctionExecution records function execution metrics
func RecordFunctionExecution(functionID, runtime, provider, region string, duration time.Duration, success bool) {
	functionExecutionDuration.WithLabelValues(functionID, runtime, provider, region).Observe(duration.Seconds())

	if !success {
		functionExecutionErrorsTotal.WithLabelValues(functionID, runtime, provider, region, "execution_error").Inc()
	}
}

// RecordFunctionInvocation records function invocation metrics
func RecordFunctionInvocation(functionID, runtime, provider, region, trigger string) {
	functionInvocationsTotal.WithLabelValues(functionID, runtime, provider, region, trigger).Inc()
}

// UpdateMemoryUsage updates memory usage metrics
func UpdateMemoryUsage(service, memType string, bytes int64) {
	memoryUsageBytes.WithLabelValues(service, memType).Set(float64(bytes))
}

// UpdateGoroutinesCount updates goroutine count metrics
func UpdateGoroutinesCount(service string, count int) {
	goroutinesCount.WithLabelValues(service).Set(float64(count))
}

// UpdateQueueSize updates queue size metrics
func UpdateQueueSize(queueName string, size int) {
	queueSize.WithLabelValues(queueName).Set(float64(size))
}

// RecordQueueProcessingDuration records queue processing duration
func RecordQueueProcessingDuration(queueName, operation string, duration time.Duration) {
	queueProcessingDuration.WithLabelValues(queueName, operation).Observe(duration.Seconds())
}

// RecordAlert records alert metrics
func RecordAlert(alertType, severity, component string) {
	alertsTriggeredTotal.WithLabelValues(alertType, severity, component).Inc()
}

// UpdateActiveIncidents updates active incident count
func UpdateActiveIncidents(severity, component string, count int) {
	activeIncidents.WithLabelValues(severity, component).Set(float64(count))
}

// UpdateUserSessions updates active user session count
func UpdateUserSessions(count int) {
	userSessionsActive.Set(float64(count))
}

// RecordFunctionDeployment records function deployment metrics
func RecordFunctionDeployment(status, provider, region string) {
	functionDeploymentsTotal.WithLabelValues(status, provider, region).Inc()
}

// UpdateLoadTestStatus updates load test status metrics
func UpdateLoadTestStatus(active bool, virtualUsers int) {
	activeValue := 0.0
	if active {
		activeValue = 1.0
	}
	loadTestActive.Set(activeValue)
	loadTestVirtualUsers.Set(float64(virtualUsers))
}

// Local runtime metric recording functions

// RecordLocalRuntimeMemoryUsage records memory usage for local runtimes
func RecordLocalRuntimeMemoryUsage(runtimeType string, vms, rss, data uint64) {
	localRuntimeMemoryUsage.WithLabelValues(runtimeType, "vms").Set(float64(vms))
	localRuntimeMemoryUsage.WithLabelValues(runtimeType, "rss").Set(float64(rss))
	localRuntimeMemoryUsage.WithLabelValues(runtimeType, "data").Set(float64(data))
}

// RecordLocalRuntimeCPUUsage records CPU usage for local runtimes
func RecordLocalRuntimeCPUUsage(runtimeType string, cpuPercent float64) {
	localRuntimeCPUUsage.WithLabelValues(runtimeType).Set(cpuPercent)
}

// RecordLocalRuntimeRequestThroughput records request throughput for local runtimes
func RecordLocalRuntimeRequestThroughput(runtimeType string, throughput float64) {
	localRuntimeRequestThroughput.WithLabelValues(runtimeType).Set(throughput)
}

// RecordLocalRuntimeActiveConnections records active connections for local runtimes
func RecordLocalRuntimeActiveConnections(runtimeType string, port int, connections int) {
	localRuntimeActiveConnections.WithLabelValues(runtimeType, fmt.Sprintf("%d", port)).Set(float64(connections))
}

// RecordLocalRuntimeExecution records function execution metrics for local runtimes
func RecordLocalRuntimeExecution(runtimeType, status, functionName string) {
	localRuntimeExecutionTotal.WithLabelValues(runtimeType, status, functionName).Inc()
}

// RecordLocalRuntimeExecutionDuration records execution duration for local runtimes
func RecordLocalRuntimeExecutionDuration(runtimeType, functionName string, duration time.Duration) {
	localRuntimeExecutionDuration.WithLabelValues(runtimeType, functionName).Observe(duration.Seconds())
}

// RecordLocalRuntimeError records errors for local runtimes
func RecordLocalRuntimeError(runtimeType, errorType, functionName string) {
	localRuntimeErrorsTotal.WithLabelValues(runtimeType, errorType, functionName).Inc()
}

// State operation metrics recording functions

// RecordStateOperation records state operation metrics
func RecordStateOperation(tenantID, operationType, status string) {
	stateOperationsTotal.WithLabelValues(tenantID, operationType, status).Inc()
}

// RecordStateOperationDuration records state operation duration
func RecordStateOperationDuration(tenantID, operationType string, duration time.Duration) {
	stateOperationDuration.WithLabelValues(tenantID, operationType).Observe(duration.Seconds())
}

// UpdateStateStorageSize updates state storage size metrics
func UpdateStateStorageSize(tenantID, statePath string, sizeBytes int64) {
	stateStorageSize.WithLabelValues(tenantID, statePath).Set(float64(sizeBytes))
}

// State fabric metrics recording functions

// RecordStateFabricOperation records state fabric operation metrics
func RecordStateFabricOperation(tenantID, fabricID, operationType, status string) {
	stateFabricOperationsTotal.WithLabelValues(tenantID, fabricID, operationType, status).Inc()
}

// RecordStateFabricOperationDuration records state fabric operation duration
func RecordStateFabricOperationDuration(tenantID, fabricID, operationType string, duration time.Duration) {
	stateFabricOperationDuration.WithLabelValues(tenantID, fabricID, operationType).Observe(duration.Seconds())
}

// UpdateStateFabricActiveCount updates the count of active state fabrics
func UpdateStateFabricActiveCount(tenantID string, count int) {
	stateFabricActiveCount.WithLabelValues(tenantID).Set(float64(count))
}

// UpdateStateFabricStoreCount updates the count of stores in a state fabric
func UpdateStateFabricStoreCount(tenantID, fabricID string, count int) {
	stateFabricStoreCount.WithLabelValues(tenantID, fabricID).Set(float64(count))
}

// RecordStateFabricPipelineExecution records pipeline execution metrics
func RecordStateFabricPipelineExecution(tenantID, fabricID, pipelineID, status string) {
	stateFabricPipelineExecutionsTotal.WithLabelValues(tenantID, fabricID, pipelineID, status).Inc()
}

// RecordStateFabricPipelineDuration records pipeline execution duration
func RecordStateFabricPipelineDuration(tenantID, fabricID, pipelineID string, duration time.Duration) {
	stateFabricPipelineDuration.WithLabelValues(tenantID, fabricID, pipelineID).Observe(duration.Seconds())
}

// Trigger execution metrics recording functions

// RecordTriggerExecution records trigger execution metrics
func RecordTriggerExecution(tenantID, statePath, triggerType, status string) {
	triggerExecutionsTotal.WithLabelValues(tenantID, statePath, triggerType, status).Inc()
}

// RecordTriggerExecutionDuration records trigger execution duration
func RecordTriggerExecutionDuration(tenantID, statePath, triggerType string, duration time.Duration) {
	triggerExecutionDuration.WithLabelValues(tenantID, statePath, triggerType).Observe(duration.Seconds())
}

// RecordTriggerError records trigger error metrics
func RecordTriggerError(tenantID, statePath, triggerType, errorType string) {
	triggerErrorsTotal.WithLabelValues(tenantID, statePath, triggerType, errorType).Inc()
}

// FRG Graph execution metrics recording functions

// RecordFRGGraphExecution records a FRG graph execution event
func RecordFRGGraphExecution(tenantID, graphID, operationType, status string) {
	frgGraphExecutionsTotal.WithLabelValues(tenantID, graphID, operationType, status).Inc()
}

// RecordFRGGraphExecutionDuration records the duration of a FRG graph execution
func RecordFRGGraphExecutionDuration(tenantID, graphID string, duration time.Duration) {
	frgGraphExecutionDuration.WithLabelValues(tenantID, graphID).Observe(duration.Seconds())
}

// RecordFRGGraphActiveIncrement increments the active graph execution count
func RecordFRGGraphActiveIncrement(tenantID string) {
	frgGraphActiveCount.WithLabelValues(tenantID).Inc()
}

// RecordFRGGraphActiveDecrement decrements the active graph execution count
func RecordFRGGraphActiveDecrement(tenantID string) {
	frgGraphActiveCount.WithLabelValues(tenantID).Dec()
}

// RecordFRGQuotaExceeded records when a FRG graph execution is blocked due to quota
func RecordFRGQuotaExceeded(tenantID string) {
	frgQuotaExceededTotal.WithLabelValues(tenantID).Inc()
}

// RecordFRGQuotaUsagePercent records the current quota usage percentage
func RecordFRGQuotaUsagePercent(tenantID string, percent float64) {
	frgQuotaUsagePercent.WithLabelValues(tenantID).Set(percent)
}

// RecordFRGWebhookSignatureFailure records a FRG webhook signature verification failure
func RecordFRGWebhookSignatureFailure(reason string) {
	frgWebhookSignatureFailures.WithLabelValues(reason).Inc()
}

// RecordFRGGraphCreation records a FRG graph creation event
func RecordFRGGraphCreation(tenantID, visibility, status string) {
	frgGraphCreationTotal.WithLabelValues(tenantID, visibility, status).Inc()
}

// UpdateFRGGraphNodesTotal updates the total node count across all FRG graphs
func UpdateFRGGraphNodesTotal(tenantID string, count int) {
	frgGraphNodesTotal.WithLabelValues(tenantID).Set(float64(count))
}

// Edge metrics (edge.functionfly.com)

// UpdateEdgeProbeResult updates edge health, latency, and error metrics from a probe.
func UpdateEdgeProbeResult(ok bool, latencyMs int, errorMessage string) {
	if ok {
		edgeHealthStatus.Set(1)
		edgeProbeLatencySeconds.Set(float64(latencyMs) / 1000.0)
	} else {
		edgeHealthStatus.Set(0)
		edgeProbeLatencySeconds.Set(float64(latencyMs) / 1000.0)
		edgeProbeErrorsTotal.Inc()
	}
}

// RecordEdgeRequest increments the edge request counter (call when routing to edge backend).
func RecordEdgeRequest() {
	edgeRequestsTotal.Inc()
}

// UpdateEdgeUptimeRatio sets the edge uptime ratio gauge (0.0 to 1.0).
func UpdateEdgeUptimeRatio(ratio float64) {
	edgeUptimeRatio.Set(ratio)
}

// State event and snapshot metrics recording functions

// RecordStateEvent records state event metrics
func RecordStateEvent(tenantID, fabricID, eventType string) {
	stateEventCount.WithLabelValues(tenantID, fabricID, eventType).Inc()
}

// UpdateStateSnapshotCount updates the count of state snapshots
func UpdateStateSnapshotCount(tenantID, fabricID string, count int) {
	stateSnapshotCount.WithLabelValues(tenantID, fabricID).Set(float64(count))
}

// HTTPMetricsMiddleware returns a middleware that records HTTP metrics
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get route variables from Gorilla mux
		vars := mux.Vars(r)
		route := mux.CurrentRoute(r)

		// Extract method, endpoint, and other info
		method := r.Method
		endpoint := r.URL.Path
		backendProvider := vars["provider"]
		if backendProvider == "" {
			backendProvider = "unknown"
		}
		region := vars["region"]
		if region == "" {
			region = "unknown"
		}

		// If we have route information, use it for better endpoint naming
		if route != nil {
			if path, err := route.GetPathTemplate(); err == nil {
				endpoint = path
			}
		}

		// WebSocket upgrade requires the raw ResponseWriter (http.Hijacker); do not wrap
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			RecordHTTPMetrics(method, endpoint, "101", backendProvider, region, duration)
			return
		}

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		statusCode := strconv.Itoa(rw.statusCode)
		RecordHTTPMetrics(method, endpoint, statusCode, backendProvider, region, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Billing metrics recording functions

// RecordBillingCheckoutCreated records when a checkout session is created
func RecordBillingCheckoutCreated(checkoutType string) {
	billingCheckoutCreatedTotal.WithLabelValues(checkoutType).Inc()
}

// RecordBillingCheckoutCompleted records when a checkout session is completed
func RecordBillingCheckoutCompleted(checkoutType string) {
	billingCheckoutCompletedTotal.WithLabelValues(checkoutType).Inc()
}

// RecordBillingCheckoutFailed records when a checkout fails
func RecordBillingCheckoutFailed(checkoutType, failureReason string) {
	billingCheckoutFailedTotal.WithLabelValues(checkoutType, failureReason).Inc()
}

// RecordBillingCheckoutDuration records the duration from checkout creation to completion
func RecordBillingCheckoutDuration(checkoutType string, duration time.Duration) {
	billingCheckoutDuration.WithLabelValues(checkoutType).Observe(duration.Seconds())
}

// RecordBillingWebhookReceived records when a webhook is received
func RecordBillingWebhookReceived(eventType, status string) {
	billingWebhookReceivedTotal.WithLabelValues(eventType, status).Inc()
}

// RecordBillingWebhookProcessingDuration records webhook processing time
func RecordBillingWebhookProcessingDuration(eventType string, duration time.Duration) {
	billingWebhookProcessingDuration.WithLabelValues(eventType).Observe(duration.Seconds())
}

// RecordBillingPaymentFailed records a failed payment
func RecordBillingPaymentFailed(failureType, declineCode string) {
	billingPaymentFailedTotal.WithLabelValues(failureType, declineCode).Inc()
}

// RecordBillingSubscriptionStatusChange records a subscription status change
func RecordBillingSubscriptionStatusChange(fromStatus, toStatus string) {
	billingSubscriptionStatusChangesTotal.WithLabelValues(fromStatus, toStatus).Inc()
}

// UpdateBillingWalletBalance updates the wallet balance gauge
func UpdateBillingWalletBalance(userID, tenantID string, balanceUSD float64) {
	billingWalletBalance.WithLabelValues(userID, tenantID).Set(balanceUSD)
}

// RecordBillingPortalSessionCreated records when a portal session is created
func RecordBillingPortalSessionCreated() {
	billingPortalSessionsCreatedTotal.Inc()
}

// RecordStripeEventProcessed records when a Stripe event is successfully processed
func RecordStripeEventProcessed(eventType string) {
	billingWebhookReceivedTotal.WithLabelValues(eventType, "processed").Inc()
}

// Execution log retention metrics recording functions

// RecordExecutionLogCleanupDeleted records the number of records deleted during cleanup
func RecordExecutionLogCleanupDeleted(tableName string, count int64) {
	executionLogCleanupRecordsDeleted.WithLabelValues(tableName).Add(float64(count))
}

// RecordExecutionLogCleanupDuration records the duration of cleanup operations
func RecordExecutionLogCleanupDuration(result string, duration time.Duration) {
	executionLogCleanupDuration.WithLabelValues(result).Observe(duration.Seconds())
}

// RecordExecutionLogCleanupError records errors during cleanup
func RecordExecutionLogCleanupError(tableName, errorType string) {
	executionLogCleanupErrors.WithLabelValues(tableName, errorType).Inc()
}

// UpdateExecutionLogRetentionAge updates the configured retention age metric
func UpdateExecutionLogRetentionAge(tableName string, days int) {
	executionLogRetentionAge.WithLabelValues(tableName).Set(float64(days))
}

// UpdateExecutionLogTableRecords updates the record count metrics for a table
func UpdateExecutionLogTableRecords(tableName, ageRange string, count int64) {
	executionLogTableRecords.WithLabelValues(tableName, ageRange).Set(float64(count))
}

// Team Memory metrics recording functions

// RecordTeamMemoryExtraction records a memory extraction attempt
func RecordTeamMemoryExtraction(teamID, memoryType, status string) {
	teamMemoryExtractionsTotal.WithLabelValues(teamID, memoryType, status).Inc()
}

// RecordTeamMemoryExtractionDuration records the time spent extracting memories
func RecordTeamMemoryExtractionDuration(teamID, source string, duration time.Duration) {
	teamMemoryExtractionDuration.WithLabelValues(teamID, source).Observe(duration.Seconds())
}

// RecordTeamMemoryExtractionConfidence records the confidence score of an extraction
func RecordTeamMemoryExtractionConfidence(teamID, memoryType string, confidence float64) {
	teamMemoryExtractionConfidence.WithLabelValues(teamID, memoryType).Observe(confidence)
}

// RecordTeamMemoryCreated records when a team memory is created
func RecordTeamMemoryCreated(teamID, memoryType, source string) {
	teamMemoriesCreatedTotal.WithLabelValues(teamID, memoryType, source).Inc()
}

// RecordTeamMemoryContextInjection records a context injection to an agent
func RecordTeamMemoryContextInjection(teamID string) {
	teamMemoryContextInjectionsTotal.WithLabelValues(teamID).Inc()
}

// RecordTeamMemoryContextInjectionDuration records the duration of building/injecting context
func RecordTeamMemoryContextInjectionDuration(teamID string, duration time.Duration) {
	teamMemoryContextInjectionDuration.WithLabelValues(teamID).Observe(duration.Seconds())
}

// RecordTeamMemorySearchDuration records the duration of memory search operations
func RecordTeamMemorySearchDuration(teamID, searchType string, duration time.Duration) {
	teamMemorySearchDuration.WithLabelValues(teamID, searchType).Observe(duration.Seconds())
}

// UpdateTeamMemoriesActiveGauge updates the count of active memories
func UpdateTeamMemoriesActiveGauge(teamID, memoryType, validated string, count int) {
	teamMemoriesActiveGauge.WithLabelValues(teamID, memoryType, validated).Set(float64(count))
}

// RecordTeamMemoryCacheHit records a cache hit for team memory context
func RecordTeamMemoryCacheHit(teamID string) {
	teamMemoryCacheHitsTotal.WithLabelValues(teamID).Inc()
}

// RecordTeamMemoryCacheMiss records a cache miss for team memory context
func RecordTeamMemoryCacheMiss(teamID string) {
	teamMemoryCacheMissesTotal.WithLabelValues(teamID).Inc()
}

// Billing Alert Recording Functions

// RecordBillingWebhookLatencyExceeded records when webhook processing exceeds threshold (5s)
func RecordBillingWebhookLatencyExceeded(eventType string) {
	billingWebhookLatencyExceeded.WithLabelValues(eventType).Inc()
}

// RecordBillingWebhookSignatureFailure records a webhook signature verification failure
func RecordBillingWebhookSignatureFailure(reason string) {
	billingWebhookSignatureFailures.WithLabelValues(reason).Inc()
}

// UpdateBillingDunningRetriesStuck updates the gauge for stuck dunning retries
func UpdateBillingDunningRetriesStuck(retryStatus string, count int) {
	billingDunningRetriesStuck.WithLabelValues(retryStatus).Set(float64(count))
}

// RecordBillingInvoiceGenerationFailure records an invoice generation failure
func RecordBillingInvoiceGenerationFailure(failureReason string) {
	billingInvoiceGenerationFailures.WithLabelValues(failureReason).Inc()
}

// RecordBillingStripeAPIError records a Stripe API error
func RecordBillingStripeAPIError(operation, errorCode string) {
	billingStripeAPIErrors.WithLabelValues(operation, errorCode).Inc()
}

// RecordBillingAlertTriggered records when a billing alert is triggered
func RecordBillingAlertTriggered(alertType, severity string) {
	billingAlertTriggered.WithLabelValues(alertType, severity).Inc()
}
