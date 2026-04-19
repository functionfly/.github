//! Function execution handlers.

use axum::{extract::Json, response::IntoResponse, extract::State};
use std::sync::Arc;
use uuid::Uuid;

use super::types::{AppState, ChunkedExecuteRequest, ChunkedExecuteResponse, ChunkedCompleteResponse, ErrorResponse, ExecuteRequest, ExecuteResponse};
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

// ---------------------------------------------------------------------------
// Chunked WASM execution handler (streaming / memory-efficient processing)
// ---------------------------------------------------------------------------

/// Maximum number of chunks per chunked execution session.
const MAX_CHUNKED_CHUNKS: u32 = 1000;
/// Maximum total WASM size across all chunks (50 MB).
const MAX_CHUNKED_TOTAL_WASM_BYTES: usize = 50 * 1024 * 1024;

/// In-memory buffer for accumulating WASM chunks.
struct ChunkedSession {
    chunks: Vec<Vec<u8>>,
    total_size: usize,
    input: Option<String>,
    timeout_ms: u64,
    memory_mb: Option<u32>,
    function: Option<String>,
    version: Option<String>,
    tenant_id: Option<String>,
}

impl ChunkedSession {
    fn new() -> Self {
        Self {
            chunks: Vec::new(),
            total_size: 0,
            input: None,
            timeout_ms: 30_000,
            memory_mb: None,
            function: None,
            version: None,
            tenant_id: None,
        }
    }

    fn add_chunk(&mut self, index: u32, data: Vec<u8>, is_last: bool) -> Result<(), &'static str> {
        if index as usize != self.chunks.len() {
            return Err("Chunk index out of order");
        }
        if self.total_size + data.len() > MAX_CHUNKED_TOTAL_WASM_BYTES {
            return Err("Total WASM size exceeds 50 MB limit");
        }
        if self.chunks.len() as u32 >= MAX_CHUNKED_CHUNKS {
            return Err("Too many chunks");
        }
        self.total_size += data.len();
        self.chunks.push(data);
        if is_last {
            // No-op: session stays open; caller should stop sending
        }
        Ok(())
    }

    fn is_complete(&self) -> bool {
        // Session is considered complete when at least one chunk has been received.
        !self.chunks.is_empty()
    }

    fn combine_wasm(&self) -> Vec<u8> {
        let mut combined = Vec::with_capacity(self.total_size);
        for chunk in &self.chunks {
            combined.extend_from_slice(chunk);
        }
        combined
    }
}

/// Base64 decode helper (mirrors daemon.rs).
fn base64_decode(input: &str) -> Result<Vec<u8>, String> {
    let input = input.replace(['\n', '\r', ' '], "");
    let mut output = Vec::new();
    let mut buf: u32 = 0;
    let mut bits: u32 = 0;

    for ch in input.bytes() {
        let val = match ch {
            b'A'..=b'Z' => (ch - b'A') as u32,
            b'a'..=b'z' => (ch - b'a' + 26) as u32,
            b'0'..=b'9' => (ch - b'0' + 52) as u32,
            b'+' => 62,
            b'/' => 63,
            b'=' => break,
            _ => return Err(format!("Invalid base64 character: {}", ch as char)),
        };
        buf = (buf << 6) | val;
        bits += 6;
        if bits >= 8 {
            bits -= 8;
            output.push((buf >> bits) as u8);
            buf &= (1 << bits) - 1;
        }
    }
    Ok(output)
}

/// Execute a function via streamed WASM chunks.
///
/// Protocol:
/// 1. Client sends N POST requests to /execute/chunked with ChunkedExecuteRequest
/// 2. Server accumulates chunks in memory (no persistence across restarts)
/// 3. After each chunk, server responds with ChunkedExecuteResponse (partial_output, etc.)
/// 4. After is_last=true, server executes accumulated WASM and returns ChunkedCompleteResponse
///
/// Only one active session per runtime process (suitable for low-concurrency scenarios).
pub async fn execute_chunked(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<ChunkedExecuteRequest>,
) -> axum::response::Response {
    use std::collections::HashMap;
    use std::sync::Arc;
    use tokio::sync::Mutex;

    // Session storage: keyed by a fixed sentinel (singleton per process).
    // For multi-tenant / multi-function scenarios a per-tenant key would be needed.
    static SESSION: Mutex<Option<ChunkedSession>> = Mutex::const_new(None);

    let correlation_id = state.logger.generate_correlation_id().await;

    // --- First chunk: initialise session ---
    {
        let mut session_guard = SESSION.lock().await;
        if session_guard.is_none() {
            let mut sess = ChunkedSession::new();
            sess.input = payload.input.clone();
            sess.timeout_ms = payload.timeout_ms.unwrap_or(30_000);
            sess.memory_mb = payload.memory_mb;
            sess.function = payload.function.clone();
            sess.version = payload.version.clone();
            sess.tenant_id = payload.tenant_id.clone();
            *session_guard = Some(sess);
            tracing::debug!(correlation_id = %correlation_id, "Chunked session started");
        }
    }

    // --- Decode this chunk ---
    let chunk_bytes = match base64_decode(&payload.chunk_data) {
        Ok(b) => b,
        Err(e) => {
            return ErrorResponse {
                error: format!("Failed to decode chunk: {}", e),
                correlation_id: Some(correlation_id.to_string()),
                recovery_suggestions: vec!["Ensure chunk_data is valid base64".to_string()],
            }
            .into_response();
        }
    };

    // --- Accumulate chunk ---
    {
        let mut session_guard = SESSION.lock().await;
        let session = session_guard.as_mut().expect("Session must be initialised");
        if let Err(msg) = session.add_chunk(payload.chunk_index, chunk_bytes, payload.is_last) {
            // Reset on error
            *session_guard = None;
            return ErrorResponse {
                error: msg.to_string(),
                correlation_id: Some(correlation_id.to_string()),
                recovery_suggestions: vec![],
            }
            .into_response();
        }
    }

    // --- Return per-chunk acknowledgment ---
    if !payload.is_last {
        return Json(ChunkedExecuteResponse {
            chunk_index: payload.chunk_index,
            is_last: false,
            partial_output: String::new(),
            exec_time_ms: 0,
            done: false,
        })
        .into_response();
    }

    // --- Last chunk received: execute ---
    let (wasm_bytes, input, timeout_ms, memory_mb, function, version, tenant_id) = {
        let mut session_guard = SESSION.lock().await;
        let _session = session_guard.as_mut().expect("Session must be initialised");
        let sess = std::mem::replace(
            &mut *session_guard,
            None,
        );
        let mut session = sess.expect("Session must be present");
        (session.combine_wasm(), session.input.unwrap_or_default(), session.timeout_ms, session.memory_mb, session.function, session.version, session.tenant_id)
    };

    let start = std::time::Instant::now();

    // Build config for execution
    let mut exec_config = state.config.clone();
    if let Some(f) = function {
        exec_config.function = f;
    }
    if let Some(v) = version {
        exec_config.version = v;
    }
    if let Some(ms) = Some(timeout_ms) {
        exec_config.timeout_ms = ms;
    }
    if let Some(mb) = memory_mb {
        exec_config.memory_mb = mb;
    }
    if let Some(tid) = tenant_id {
        exec_config.tenant_id = Some(tid);
    }

    let result = state
        .engine
        .execute(
            &wasm_bytes,
            &input,
            &exec_config,
            state.python_pool.clone(),
            state.micropython_executor.clone(),
        )
        .await;

    let exec_time_ms = start.elapsed().as_millis() as u64;

    match result {
        Ok(output) => {
            // Log execution
            state.logger.log_function_execution(
                &correlation_id,
                &exec_config.function,
                exec_time_ms,
                true,
                false,
            );

            Json(ChunkedCompleteResponse {
                result: output,
                total_exec_time_ms: exec_time_ms,
                chunks_processed: ((wasm_bytes.len() + 65535) / 65536) as u32,
                cache_hit: false,
            })
            .into_response()
        }
        Err(e) => ErrorResponse {
            error: format!("Execution failed: {}", e),
            correlation_id: Some(correlation_id.to_string()),
            recovery_suggestions: vec![
                "Verify the WASM binary is valid".to_string(),
                "Check that resource limits are sufficient".to_string(),
            ],
        }
        .into_response(),
    }
}
