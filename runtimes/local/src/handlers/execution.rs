//! Function execution handlers.

use axum::{extract::Json, response::IntoResponse, extract::State};
use std::sync::Arc;
use uuid::Uuid;

use super::types::{AppState, ErrorResponse, ExecuteRequest, ExecuteResponse};
use crate::errors::{RuntimeError, RecoveryStrategy};
use crate::logging::{CorrelationId, RequestContext};
use crate::scheduler::SchedulingRequest;

// ---------------------------------------------------------------------------
// Main execution handler
// ---------------------------------------------------------------------------

/// Execute a function with structured logging and error handling.
pub async fn execute_function(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<ExecuteRequest>,
) -> axum::response::Response {
    let correlation_id = state.logger.generate_correlation_id().await;
    let request_context = RequestContext::new()
        .with_function(&state.config.function, &state.config.version);

    state.logger.log_with_correlation(
        crate::logging::LogLevel::Info,
        format!("Starting function execution: {}", state.config.function),
        &correlation_id,
    );

    let start = std::time::Instant::now();
    let input = payload.input.unwrap_or_default();

    // Use tenant_id for KV store namespace isolation
    let tenant_id = payload.tenant_id.unwrap_or_else(|| "default".to_string());

    state.logger.log_with_correlation(
        crate::logging::LogLevel::Debug,
        format!("Executing function for tenant: {}", tenant_id),
        &correlation_id,
    );

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
        }
        .into_response();
    }

    // Check resource enforcement policies before execution
    if let Some(ref enforcer) = state.resource_enforcer {
        let function_key = format!("{}@{}", state.config.function, state.config.version);
        let decision = enforcer.check_execution_allowed(&function_key).await;
        match decision {
            crate::resource_enforcer::EnforcementDecision::Block(reason) => {
                state.logger.log_with_correlation(
                    crate::logging::LogLevel::Warn,
                    format!("Execution blocked by resource policy: {}", reason),
                    &correlation_id,
                );
                return ErrorResponse {
                    error: format!("Execution blocked: {}", reason),
                    correlation_id: Some(correlation_id.to_string()),
                    recovery_suggestions: vec!["Resource policy violation".to_string()],
                }
                .into_response();
            }
            crate::resource_enforcer::EnforcementDecision::Throttle(duration) => {
                state.logger.log_with_correlation(
                    crate::logging::LogLevel::Warn,
                    format!("Execution throttled for {:?}: ", duration),
                    &correlation_id,
                );
                tokio::time::sleep(duration).await;
            }
            crate::resource_enforcer::EnforcementDecision::Allow => {}
        }
    }

    let instance_id = Uuid::new_v4().to_string();
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
            let execution_result = execute_with_error_handling(
                &state, &input, &correlation_id, &request_context,
            )
            .await;

            match execution_result {
                Ok(output) => {
                    cache.set(&input, output.clone());
                    result = output;
                }
                Err(error) => {
                    // Attempt error recovery if available
                    if let Some(ref recovery) = state.error_recovery {
                        let strategy = recovery.get_recovery_strategy(&error);
                        match strategy {
                            RecoveryStrategy::Retry { attempts, delay_ms } => {
                                state.logger.log_with_correlation(
                                    crate::logging::LogLevel::Warn,
                                    format!(
                                        "Execution failed, attempting {} retries with {}ms delay",
                                        attempts, delay_ms
                                    ),
                                    &correlation_id,
                                );
                                let recovered = recovery.execute_recovery(
                                    &strategy,
                                    || async {
                                        execute_with_error_handling(
                                            &state, &input, &correlation_id, &request_context,
                                        )
                                        .await
                                    },
                                )
                                .await;
                                match recovered {
                                    Ok(output) => {
                                        state.logger.log_with_correlation(
                                            crate::logging::LogLevel::Info,
                                            format!("Error recovery succeeded after {} attempts", attempts),
                                            &correlation_id,
                                        );
                                        result = output;
                                    }
                                    Err(final_error) => {
                                        state.logger.log_error(&correlation_id, &final_error);
                                        return ErrorResponse {
                                            error: final_error.message.clone(),
                                            correlation_id: Some(correlation_id.to_string()),
                                            recovery_suggestions: final_error.recovery_suggestions.clone(),
                                        }
                                        .into_response();
                                    }
                                }
                            }
                            _ => {
                                // Non-retryable error, fall through to normal error handling
                                state.logger.log_error(&correlation_id, &error);
                                return ErrorResponse {
                                    error: error.message.clone(),
                                    correlation_id: Some(correlation_id.to_string()),
                                    recovery_suggestions: error.recovery_suggestions.clone(),
                                }
                                .into_response();
                            }
                        }
                    } else {
                        state.logger.log_error(&correlation_id, &error);
                        return ErrorResponse {
                            error: error.message.clone(),
                            correlation_id: Some(correlation_id.to_string()),
                            recovery_suggestions: error.recovery_suggestions.clone(),
                        }
                        .into_response();
                    }
                }
            }
        }
    } else {
        match execute_with_error_handling(&state, &input, &correlation_id, &request_context).await {
            Ok(output) => result = output,
            Err(error) => {
                // Attempt error recovery if available
                if let Some(ref recovery) = state.error_recovery {
                    let strategy = recovery.get_recovery_strategy(&error);
                    match strategy {
                        RecoveryStrategy::Retry { attempts, delay_ms } => {
                            state.logger.log_with_correlation(
                                crate::logging::LogLevel::Warn,
                                format!(
                                    "Execution failed, attempting {} retries with {}ms delay",
                                    attempts, delay_ms
                                ),
                                &correlation_id,
                            );
                            let recovered = recovery.execute_recovery(
                                &strategy,
                                || async {
                                    execute_with_error_handling(
                                        &state, &input, &correlation_id, &request_context,
                                    )
                                    .await
                                },
                            )
                            .await;
                            match recovered {
                                Ok(output) => {
                                    state.logger.log_with_correlation(
                                        crate::logging::LogLevel::Info,
                                        format!("Error recovery succeeded after {} attempts", attempts),
                                        &correlation_id,
                                    );
                                    result = output;
                                }
                                Err(final_error) => {
                                    state.logger.log_error(&correlation_id, &final_error);
                                    return ErrorResponse {
                                        error: final_error.message.clone(),
                                        correlation_id: Some(correlation_id.to_string()),
                                        recovery_suggestions: final_error.recovery_suggestions.clone(),
                                    }
                                    .into_response();
                                }
                            }
                        }
                        _ => {
                            state.logger.log_error(&correlation_id, &error);
                            return ErrorResponse {
                                error: error.message.clone(),
                                correlation_id: Some(correlation_id.to_string()),
                                recovery_suggestions: error.recovery_suggestions.clone(),
                            }
                            .into_response();
                        }
                    }
                } else {
                    state.logger.log_error(&correlation_id, &error);
                    return ErrorResponse {
                        error: error.message.clone(),
                        correlation_id: Some(correlation_id.to_string()),
                        recovery_suggestions: error.recovery_suggestions.clone(),
                    }
                    .into_response();
                }
            }
        }
    }

    let exec_time = start.elapsed().as_millis() as u64;

    // Record resource usage with resource enforcer after successful execution
    if let Some(ref enforcer) = state.resource_enforcer {
        let function_key = format!("{}@{}", state.config.function, state.config.version);
        let metrics = crate::monitoring::ExecutionMetrics {
            function_name: state.config.function.clone(),
            function_version: state.config.version.clone(),
            execution_time_ms: exec_time,
            cpu_fuel_used: 0,
            memory_used_mb: state.config.memory_mb as f64,
            peak_memory_mb: state.config.memory_mb as f64,
            cache_hit,
            cold_start: false,
            error_occurred: false,
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
        };
        enforcer.record_usage(&function_key, &metrics).await;
    }

    // Record execution metrics with monitoring system
    let monitoring_metrics = crate::monitoring::ExecutionMetrics {
        function_name: state.config.function.clone(),
        function_version: state.config.version.clone(),
        execution_time_ms: exec_time,
        cpu_fuel_used: 0,
        memory_used_mb: state.config.memory_mb as f64,
        peak_memory_mb: state.config.memory_mb as f64,
        cache_hit,
        cold_start: false,
        error_occurred: false,
        timestamp: std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs(),
    };
    state.monitor.record_execution(monitoring_metrics).await;

    state.logger.log_function_execution(
        &correlation_id,
        &state.config.function,
        exec_time,
        true,
        cache_hit,
    );

    Json(ExecuteResponse {
        result,
        exec_time_ms: exec_time,
        cache_hit,
        instance_id,
        function: state.config.function.clone(),
        version: state.config.version.clone(),
    })
    .into_response()
}

// ---------------------------------------------------------------------------
// Core execution logic
// ---------------------------------------------------------------------------

/// Execute function with comprehensive pre-flight checks and error handling.
async fn execute_with_error_handling(
    state: &AppState,
    input: &str,
    correlation_id: &CorrelationId,
    request_context: &RequestContext,
) -> Result<String, RuntimeError> {
    // Load WASM bytes
    let wasm_bytes = match &state.config.wasm {
        Some(wasm_path) => std::fs::read(wasm_path).map_err(|_e| {
            RuntimeError::file_not_found(wasm_path)
                .with_correlation_id(correlation_id.to_string())
                .with_context(crate::errors::ErrorContext {
                    function_name: Some(request_context.function_name.clone().unwrap_or_default()),
                    function_version: Some(request_context.function_version.clone().unwrap_or_default()),
                    input_size: Some(input.len()),
                    ..Default::default()
                })
        })?,
        None => {
            return Err(
                RuntimeError::config_error("No WebAssembly module specified")
                    .with_correlation_id(correlation_id.to_string()),
            )
        }
    };

    let function_key = format!("{}@{}", state.config.function, state.config.version);

    // YARA scan before execution
    if let Some(ref scanner) = state.yara_scanner {
        if let Err(e) = scanner.scan_or_block(&wasm_bytes).await {
            tracing::warn!("WASM artifact blocked by YARA scanner for {}: {}", function_key, e);
            return Err(RuntimeError::security_violation(e.to_string())
                .with_correlation_id(correlation_id.to_string()));
        }
        tracing::debug!("YARA scan passed for {}", function_key);
    }

    // Enterprise security checks
    if let Some(ref enterprise_security) = state.enterprise_security {
        match enterprise_security.validate_input(input, &function_key).await {
            crate::enterprise_security::ValidationResult::Invalid(reason) => {
                return Err(RuntimeError::security_violation(reason)
                    .with_correlation_id(correlation_id.to_string()));
            }
            crate::enterprise_security::ValidationResult::Suspicious(reason) => {
                tracing::warn!("Suspicious input detected for {}: {}", function_key, reason);
                enterprise_security
                    .log_audit_entry(&function_key, "input_validation", "suspicious", &reason, None, None)
                    .await;
            }
            crate::enterprise_security::ValidationResult::Valid => {}
        }

        if !enterprise_security.check_rate_limit(&function_key, &function_key).await {
            return Err(RuntimeError::rate_limit_exceeded()
                .with_correlation_id(correlation_id.to_string()));
        }

        // Validate sandboxing with declared capabilities
        let capabilities: Vec<String> = state.config.capabilities
            .split(',')
            .filter(|s| !s.is_empty())
            .map(|s| s.trim().to_string())
            .collect();
        if !capabilities.is_empty() {
            if let crate::enterprise_security::ValidationResult::Invalid(reason) = enterprise_security.validate_sandboxing(&function_key, &capabilities).await {
                return Err(RuntimeError::security_violation(format!("Sandboxing violation: {}", reason))
                    .with_correlation_id(correlation_id.to_string()));
            }
        }

        enterprise_security
            .log_audit_entry(&function_key, "execution", "attempted", "Function execution started", None, None)
            .await;
    }

    // Resource limit checks
    if let Err(limit_error) = state
        .monitor
        .check_limits(&state.config.function, &state.config.version)
        .await
    {
        return Err(RuntimeError::resource_limit(limit_error)
            .with_correlation_id(correlation_id.to_string()));
    }

    // Enterprise resource enforcement
    if let Some(ref enforcer) = state.resource_enforcer {
        match enforcer.check_execution_allowed(&function_key).await {
            crate::resource_enforcer::EnforcementDecision::Allow => {}
            crate::resource_enforcer::EnforcementDecision::Throttle(duration) => {
                tokio::time::sleep(duration).await;
                tracing::warn!("Function {} throttled for {:?}", function_key, duration);
            }
            crate::resource_enforcer::EnforcementDecision::Block(reason) => {
                return Err(RuntimeError::resource_limit(reason)
                    .with_correlation_id(correlation_id.to_string()));
            }
        }
    }

    // Security violation block check
    if state
        .security_monitor
        .should_block_function(&state.config.function)
        .await
    {
        state.security_monitor
            .record_violation(
                state.config.function.clone(),
                crate::security::ViolationType::ResourceExhaustion,
                "Function blocked due to excessive security violations".to_string(),
                crate::security::Severity::Critical,
            )
            .await;

        return Err(RuntimeError::security_violation("Function blocked due to security violations")
            .with_correlation_id(correlation_id.to_string()));
    }

    state.monitor
        .increment_concurrent(&state.config.function, &state.config.version)
        .await;

    // Record execution start with scheduler if available
    let scheduling_req = SchedulingRequest::from_resources(
        state.config.memory_mb,
        1, // vCPUs
    );
    if let Some(ref sched) = state.scheduler {
        sched.record_execution_start("local-node", &scheduling_req).await;
    }

    let timeout_duration = std::time::Duration::from_millis(state.config.timeout_ms);
    let execution_future = state.engine.execute(
        &wasm_bytes,
        input,
        &state.config,
        state.python_pool.clone(),
        state.micropython_executor.clone(),
    );

    let result = match tokio::time::timeout(timeout_duration, execution_future).await {
        Ok(result) => match result {
            Ok(output) => {
                state.monitor
                    .decrement_concurrent(&state.config.function, &state.config.version)
                    .await;
                // Record execution end with scheduler
                if let Some(ref sched) = state.scheduler {
                    sched.record_execution_end("local-node", &scheduling_req).await;
                }
                Ok(output)
            }
            Err(e) => {
                state.monitor
                    .decrement_concurrent(&state.config.function, &state.config.version)
                    .await;
                // Record execution end with scheduler
                if let Some(ref sched) = state.scheduler {
                    sched.record_execution_end("local-node", &scheduling_req).await;
                }
                Err(RuntimeError::wasm_execution(e.to_string())
                    .with_correlation_id(correlation_id.to_string())
                    .with_context(crate::errors::ErrorContext {
                        function_name: Some(request_context.function_name.clone().unwrap_or_default()),
                        function_version: Some(request_context.function_version.clone().unwrap_or_default()),
                        input_size: Some(input.len()),
                        execution_time: Some(request_context.elapsed()),
                        ..Default::default()
                    }))
            }
        },
        Err(_) => {
            state.monitor
                .decrement_concurrent(&state.config.function, &state.config.version)
                .await;
            // Record execution end with scheduler (for timeout case too)
            if let Some(ref sched) = state.scheduler {
                sched.record_execution_end("local-node", &scheduling_req).await;
            }
            Err(RuntimeError::timeout(state.config.timeout_ms)
                .with_correlation_id(correlation_id.to_string())
                .with_context(crate::errors::ErrorContext {
                    function_name: Some(request_context.function_name.clone().unwrap_or_default()),
                    function_version: Some(request_context.function_version.clone().unwrap_or_default()),
                    input_size: Some(input.len()),
                    execution_time: Some(std::time::Duration::from_millis(state.config.timeout_ms)),
                    ..Default::default()
                }))
        }
    };

    result
}
