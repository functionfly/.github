//! HTTP handlers for the local runtime.

use axum::{
    extract::{Json, State},
    response::IntoResponse,
};
use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex};
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::cache::ResultCache;
use crate::config::Config;
use crate::engine::WasmEngine;
use crate::errors::RuntimeError;
use crate::budget::{BudgetOptimizer, BudgetTier, FunctionProfile};
use crate::kv::SharedKVStore;
use crate::logging::{CorrelationId, RequestContext, StructuredLogger};
use crate::monitoring::{ExecutionMetrics, ResourceMonitor};
use crate::enterprise_security::EnterpriseSecurityEnforcer;
use crate::package::PackageManager;
use crate::resource_enforcer::ResourceEnforcer;
use crate::security::{SecurityMonitor, ViolationType, Severity};

/// Application state
#[derive(Clone)]
pub struct AppState {
    pub engine: Arc<crate::engine::WasmEngine>,
    pub pool: Arc<RwLock<crate::pool::InstancePool>>,
    pub cache: Arc<RwLock<ResultCache>>,
    pub kv: SharedKVStore,
    pub config: Config,
    pub logger: StructuredLogger,
    pub monitor: Arc<ResourceMonitor>,
    pub security_monitor: Arc<SecurityMonitor>,
    pub package_manager: Option<Arc<PackageManager>>,
    pub resource_enforcer: Option<Arc<ResourceEnforcer>>,
    pub enterprise_security: Option<Arc<EnterpriseSecurityEnforcer>>,
    pub orchestrator_client: Option<Arc<crate::orchestrator_client::OrchestratorClient>>,
}

/// Execute request
#[derive(Debug, Deserialize)]
pub struct ExecuteRequest {
    /// Input to the function
    pub input: Option<String>,
}

/// Execute response
#[derive(Debug, Serialize)]
pub struct ExecuteResponse {
    /// Function output
    pub result: String,
    /// Execution time in milliseconds
    pub exec_time_ms: u64,
    /// Whether the result was served from cache
    pub cache_hit: bool,
    /// Instance ID
    pub instance_id: String,
    /// Function name
    pub function: String,
    /// Function version
    pub version: String,
}

/// Health check response
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
}

/// Error response
#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub correlation_id: Option<String>,
    pub recovery_suggestions: Vec<String>,
}

impl IntoResponse for ErrorResponse {
    fn into_response(self) -> axum::response::Response {
        let status = axum::http::StatusCode::INTERNAL_SERVER_ERROR;
        (status, axum::Json(self)).into_response()
    }
}

/// Execute a function with structured logging and error handling
pub async fn execute_function(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<ExecuteRequest>,
) -> axum::response::Response {
    // Generate correlation ID for this request
    let correlation_id = state.logger.generate_correlation_id().await;
    let request_context = RequestContext::new(correlation_id.clone())
        .with_function(&state.config.function, &state.config.version);

    state.logger.log_with_correlation(
        crate::logging::LogLevel::Info,
        format!("Starting function execution: {}", state.config.function),
        &correlation_id,
    );

    let start = std::time::Instant::now();
    let input = payload.input.unwrap_or_default();

    // Reject inputs that exceed the configured maximum size before doing any
    // WASM work.  This prevents memory pressure from oversized payloads.
    if state.config.max_input_bytes > 0 && input.len() > state.config.max_input_bytes {
        return ErrorResponse {
            error: format!(
                "Input size {} bytes exceeds maximum allowed size of {} bytes",
                input.len(),
                state.config.max_input_bytes
            ),
            correlation_id: Some(correlation_id.to_string()),
            recovery_suggestions: vec![
                "Reduce the size of the input payload".to_string(),
                format!("Maximum allowed input size is {} bytes", state.config.max_input_bytes),
            ],
        }.into_response();
    }
    let instance_id = Uuid::new_v4().to_string();

    // Check cache for deterministic functions
    let cache_hit = state.config.deterministic;
    let result: String;

    if cache_hit {
        let mut cache = state.cache.write().await;
        if let Some(cached) = cache.get(&input) {
            state.logger.log_with_correlation(
                crate::logging::LogLevel::Debug,
                "Cache hit for function execution",
                &correlation_id,
            );
            result = cached;
        } else {
            // Execute function using WebAssembly
            let execution_result = execute_with_error_handling(
                &state,
                &input,
                &correlation_id,
                &request_context,
            ).await;

            match execution_result {
                Ok(output) => {
                    // Cache result if deterministic
                    cache.set(&input, output.clone());
                    result = output;
                }
                Err(error) => {
                    state.logger.log_error(&correlation_id, &error);
                    return ErrorResponse {
                        error: error.message.clone(),
                        correlation_id: Some(correlation_id.to_string()),
                        recovery_suggestions: error.recovery_suggestions.clone(),
                    }.into_response();
                }
            }
        }
    } else {
        // Execute function without caching
        match execute_with_error_handling(
            &state,
            &input,
            &correlation_id,
            &request_context,
        ).await {
            Ok(output) => result = output,
                Err(error) => {
                    state.logger.log_error(&correlation_id, &error);
                    return ErrorResponse {
                        error: error.message.clone(),
                        correlation_id: Some(correlation_id.to_string()),
                        recovery_suggestions: error.recovery_suggestions.clone(),
                    }.into_response();
                }
        }
    }

    let exec_time = start.elapsed().as_millis() as u64;

    // Log function execution with structured logging
    state.logger.log_function_execution(
        &correlation_id,
        &state.config.function,
        exec_time,
        true, // success
        cache_hit,
    );

    Json(ExecuteResponse {
        result,
        exec_time_ms: exec_time,
        cache_hit,
        instance_id,
        function: state.config.function.clone(),
        version: state.config.version.clone(),
    }).into_response()
}

/// Execute function with comprehensive error handling
async fn execute_with_error_handling(
    state: &AppState,
    input: &str,
    correlation_id: &CorrelationId,
    request_context: &RequestContext,
) -> Result<String, RuntimeError> {
    // Load WASM bytes
    let wasm_bytes = match &state.config.wasm {
        Some(wasm_path) => std::fs::read(wasm_path)
            .map_err(|e| RuntimeError::file_not_found(wasm_path)
                .with_correlation_id(correlation_id.to_string())
                .with_context(crate::errors::ErrorContext {
                    function_name: Some(request_context.function_name.clone().unwrap_or_default()),
                    function_version: Some(request_context.function_version.clone().unwrap_or_default()),
                    input_size: Some(input.len()),
                    ..Default::default()
                }))?,
        None => return Err(RuntimeError::config_error("No WebAssembly module specified")
            .with_correlation_id(correlation_id.to_string())),
    };

    // Enterprise security checks
    let function_key = format!("{}@{}", state.config.function, state.config.version);

    if let Some(ref enterprise_security) = state.enterprise_security {
        // Validate input
        match enterprise_security.validate_input(input, &function_key).await {
            crate::enterprise_security::ValidationResult::Invalid(reason) => {
                return Err(RuntimeError::security_violation(reason)
                    .with_correlation_id(correlation_id.to_string()));
            }
            crate::enterprise_security::ValidationResult::Suspicious(reason) => {
                // Log suspicious input but allow execution
                tracing::warn!("Suspicious input detected for {}: {}", function_key, reason);
                enterprise_security.log_audit_entry(&function_key, "input_validation", "suspicious", &reason, None, None).await;
            }
            crate::enterprise_security::ValidationResult::Valid => {}
        }

        // Check rate limiting (use function key as identifier for now)
        if !enterprise_security.check_rate_limit(&function_key, &function_key).await {
            return Err(RuntimeError::resource_limit("Rate limit exceeded")
                .with_correlation_id(correlation_id.to_string()));
        }

        // Log execution attempt
        enterprise_security.log_audit_entry(&function_key, "execution", "attempted", "Function execution started", None, None).await;
    }

    // Check resource limits before execution
    if let Err(limit_error) = state.monitor.check_limits(&state.config.function, &state.config.version).await {
        return Err(RuntimeError::resource_limit(limit_error)
            .with_correlation_id(correlation_id.to_string()));
    }

    // Check enterprise resource enforcement
    if let Some(ref enforcer) = state.resource_enforcer {
        let function_key = format!("{}@{}", state.config.function, state.config.version);
        match enforcer.check_execution_allowed(&function_key).await {
            crate::resource_enforcer::EnforcementDecision::Allow => {
                // Proceed with execution
            }
            crate::resource_enforcer::EnforcementDecision::Throttle(duration) => {
                tokio::time::sleep(duration).await;
                // Log throttling
                tracing::warn!("Function {} throttled for {:?}", function_key, duration);
            }
            crate::resource_enforcer::EnforcementDecision::Block(reason) => {
                return Err(RuntimeError::resource_limit(reason)
                    .with_correlation_id(correlation_id.to_string()));
            }
        }
    }

    // Check security status - block functions with too many violations
    if state.security_monitor.should_block_function(&state.config.function).await {
        state.security_monitor.record_violation(
            state.config.function.clone(),
            ViolationType::ResourceExhaustion,
            "Function blocked due to excessive security violations".to_string(),
            Severity::Critical,
        ).await;

        return Err(RuntimeError::security_violation("Function blocked due to security violations")
            .with_correlation_id(correlation_id.to_string()));
    }

    // Increment concurrent counter
    state.monitor.increment_concurrent(&state.config.function, &state.config.version).await;

    // Execute with timeout
    let timeout_duration = std::time::Duration::from_millis(state.config.timeout_ms);
    let execution_future = state.engine.execute(&wasm_bytes, input, &state.config);

    let result = match tokio::time::timeout(timeout_duration, execution_future).await {
        Ok(result) => match result {
            Ok(output) => {
                // Decrement concurrent counter on success
                state.monitor.decrement_concurrent(&state.config.function, &state.config.version).await;
                Ok(output)
            },
            Err(e) => {
                // Decrement concurrent counter on error
                state.monitor.decrement_concurrent(&state.config.function, &state.config.version).await;

                let error = RuntimeError::wasm_execution(e.to_string())
                    .with_correlation_id(correlation_id.to_string())
                    .with_context(crate::errors::ErrorContext {
                        function_name: Some(request_context.function_name.clone().unwrap_or_default()),
                        function_version: Some(request_context.function_version.clone().unwrap_or_default()),
                        input_size: Some(input.len()),
                        execution_time: Some(request_context.elapsed()),
                        ..Default::default()
                    });
                Err(error)
            }
        },
        Err(_) => {
            // Decrement concurrent counter on timeout
            state.monitor.decrement_concurrent(&state.config.function, &state.config.version).await;

            let error = RuntimeError::timeout(state.config.timeout_ms)
                .with_correlation_id(correlation_id.to_string())
                .with_context(crate::errors::ErrorContext {
                    function_name: Some(request_context.function_name.clone().unwrap_or_default()),
                    function_version: Some(request_context.function_version.clone().unwrap_or_default()),
                    input_size: Some(input.len()),
                    execution_time: Some(std::time::Duration::from_millis(state.config.timeout_ms)),
                    ..Default::default()
                });
            Err(error)
        }
    };

    result
}

/// Health check handler
pub async fn health_check() -> axum::Json<HealthResponse> {
    axum::Json(HealthResponse {
        status: "healthy".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
    })
}

/// Readiness check handler
pub async fn ready_check() -> axum::Json<HealthResponse> {
    axum::Json(HealthResponse {
        status: "ready".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
    })
}

/// Monitoring statistics handler
pub async fn monitoring_stats(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    let report = state.monitor.generate_report().await;

    // Get orchestrator stats if available
    let orchestrator_stats = if let Some(client) = &state.orchestrator_client {
        match client.get_stats().await {
            Ok(stats) => Some(serde_json::json!({
                "active_vms": stats.active_vms,
                "warm_vms": stats.warm_vms,
                "max_vms": stats.max_vms
            })),
            Err(_) => None,
        }
    } else {
        None
    };

    let mut response = serde_json::json!({
        "stats": report.stats,
        "performance": {
            "p50_execution_time_ms": report.p50_execution_time_ms,
            "p95_execution_time_ms": report.p95_execution_time_ms,
            "p99_execution_time_ms": report.p99_execution_time_ms,
            "recent_executions": report.recent_executions
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    });

    if let Some(stats) = orchestrator_stats {
        response["orchestrator"] = stats;
    }

    Json(response).into_response()
}

/// Budget analysis handler
pub async fn budget_analysis(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    // Get real execution statistics from monitoring data
    let stats = state.monitor.get_stats().await;
    let avg_execution_time = if stats.total_executions > 0 {
        stats.average_execution_time_ms as u64
    } else {
        100 // Fallback estimate when no executions have been recorded
    };

    // Create function profile based on current config and monitoring data
    let function_profile = FunctionProfile {
        runtime: state.config.runtime.clone(),
        avg_execution_time_ms: avg_execution_time,
        avg_memory_mb: state.config.memory_mb as f64,
        cold_start_time_ms: 50, // Estimate
        requests_per_minute: 10.0, // Estimate
        cache_hit_rate: if state.config.deterministic { 0.7 } else { 0.1 },
    };

    // Use ultra-low tier optimizer (our target)
    let optimizer = BudgetOptimizer::for_tier(&BudgetTier::UltraLow);
    let recommendations = optimizer.generate_recommendations(&vec![function_profile]);

    Json(serde_json::json!({
        "budget_analysis": recommendations,
        "current_config": {
            "memory_mb": state.config.memory_mb,
            "timeout_ms": state.config.timeout_ms,
            "deterministic": state.config.deterministic,
            "cpu_fuel_limit": state.config.cpu_fuel_limit,
            "max_concurrent_per_function": state.config.max_concurrent_per_function
        },
        "optimization_suggestions": recommendations.suggestions,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    })).into_response()
}

/// Security status handler
pub async fn security_status(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    let report = state.security_monitor.get_security_report().await;

    Json(serde_json::json!({
        "security_report": report,
        "hardened_mode": state.config.hardened_security,
        "isolation_level": if state.config.hardened_security { "maximum" } else { "standard" },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    })).into_response()
}

/// KV store status handler
pub async fn kv_status(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    let stats = {
        let store = state.kv.read().await;
        store.stats()
    };

    Json(serde_json::json!({
        "kv_stats": {
            "total_entries": stats.total_entries,
            "expired_entries": stats.expired_entries,
            "max_entries": stats.max_entries,
            "has_kv_capability": crate::capability::Capabilities::from_string(&state.config.capabilities).can_kv()
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    })).into_response()
}

/// Webhook capability status handler
pub async fn webhook_status(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    let capabilities = crate::capability::Capabilities::from_string(&state.config.capabilities);

    Json(serde_json::json!({
        "webhook_capability": {
            "enabled": capabilities.can_webhook(),
            "description": "Allows functions to send HTTP requests to external webhook endpoints",
            "supported_methods": ["GET", "POST", "PUT", "PATCH", "DELETE"],
            "timeout_seconds": 30,
            "max_payload_size_kb": 1024
        },
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    })).into_response()
}

/// Orchestrator status handler
pub async fn orchestrator_status(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    let orchestrator_info = if let Some(client) = &state.orchestrator_client {
        let health = client.ping().await;
        let stats = if health {
            match client.get_stats().await {
                Ok(stats) => Some(serde_json::json!({
                    "active_vms": stats.active_vms,
                    "warm_vms": stats.warm_vms,
                    "max_vms": stats.max_vms
                })),
                Err(_) => None,
            }
        } else {
            None
        };

        serde_json::json!({
            "available": health,
            "url": client.url(),
            "stats": stats,
            "enterprise_enabled": state.config.enterprise_enabled,
            "tier": state.config.tier,
            "supports_microvm": state.config.supports_microvm()
        })
    } else {
        serde_json::json!({
            "available": false,
            "enterprise_enabled": state.config.enterprise_enabled,
            "tier": state.config.tier,
            "supports_microvm": false
        })
    };

    Json(serde_json::json!({
        "orchestrator": orchestrator_info,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    })).into_response()
}

/// Prometheus metrics handler.
///
/// Exposes runtime metrics in Prometheus text format at `/metrics`.
/// This enables unified observability with the Go backend which already
/// exports Prometheus metrics via `prometheus/client_golang`.
///
/// Metrics exported:
/// - `functionfly_executions_total` — total function executions
/// - `functionfly_execution_time_ms_avg` — average execution time
/// - `functionfly_cache_hit_rate` — cache hit rate (0.0–100.0)
/// - `functionfly_error_rate` — error rate (0.0–100.0)
/// - `functionfly_concurrent_executions` — current concurrent executions
pub async fn prometheus_metrics(
    State(state): State<Arc<AppState>>,
) -> axum::response::Response {
    let stats = state.monitor.get_stats().await;
    let concurrent = state.monitor.get_total_concurrent().await;

    // Build Prometheus text format manually to avoid adding a heavy registry
    // dependency. For production use, consider using the `prometheus` crate's
    // `TextEncoder` with registered metrics.
    let mut output = String::new();

    output.push_str("# HELP functionfly_executions_total Total number of function executions\n");
    output.push_str("# TYPE functionfly_executions_total counter\n");
    output.push_str(&format!("functionfly_executions_total {}\n\n", stats.total_executions));

    output.push_str("# HELP functionfly_execution_time_ms_avg Average execution time in milliseconds\n");
    output.push_str("# TYPE functionfly_execution_time_ms_avg gauge\n");
    output.push_str(&format!("functionfly_execution_time_ms_avg {:.2}\n\n", stats.average_execution_time_ms));

    output.push_str("# HELP functionfly_cache_hit_rate Cache hit rate percentage (0-100)\n");
    output.push_str("# TYPE functionfly_cache_hit_rate gauge\n");
    output.push_str(&format!("functionfly_cache_hit_rate {:.2}\n\n", stats.cache_hit_rate));

    output.push_str("# HELP functionfly_error_rate Error rate percentage (0-100)\n");
    output.push_str("# TYPE functionfly_error_rate gauge\n");
    output.push_str(&format!("functionfly_error_rate {:.2}\n\n", stats.error_rate));

    output.push_str("# HELP functionfly_concurrent_executions Current number of concurrent executions\n");
    output.push_str("# TYPE functionfly_concurrent_executions gauge\n");
    output.push_str(&format!("functionfly_concurrent_executions {}\n\n", concurrent));

    output.push_str("# HELP functionfly_peak_memory_mb Peak memory usage in MB\n");
    output.push_str("# TYPE functionfly_peak_memory_mb gauge\n");
    output.push_str(&format!("functionfly_peak_memory_mb {:.2}\n\n", stats.peak_memory_usage_mb));

    output.push_str("# HELP functionfly_uptime_seconds Runtime uptime in seconds\n");
    output.push_str("# TYPE functionfly_uptime_seconds counter\n");
    output.push_str(&format!("functionfly_uptime_seconds {}\n", stats.uptime_seconds));

    axum::response::Response::builder()
        .status(200)
        .header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
        .body(axum::body::Body::from(output))
        .unwrap_or_else(|_| axum::Json(serde_json::json!({"error": "metrics unavailable"})).into_response())
}
