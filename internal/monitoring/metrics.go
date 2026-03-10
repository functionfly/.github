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
