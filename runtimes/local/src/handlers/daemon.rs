//! Daemon-mode execution handler.

use axum::{extract::{Json, Path, State}, response::IntoResponse};
use std::sync::Arc;
use uuid::Uuid;

use super::types::{AppState, DaemonExecuteRequest, ErrorResponse, ExecuteResponse};

/// Execute a function via the daemon endpoint (Phase 3.2).
///
/// Pool-aware execution strategy:
/// 1. engine.execute() handles pool acquisition internally (fast path if pool is warm).
/// 2. If the pool is empty (cold start), engine.execute() falls back to execute_wasm_standard
///    which compiles the module.
/// 3. After a successful cold-start execution, we pre-warm the pool with the compiled
///    module so subsequent requests use warm pooled instances.
pub async fn execute_function_daemon(
    State(state): State<Arc<AppState>>,
    Path((function_id, version)): Path<(String, String)>,
    Json(payload): Json<DaemonExecuteRequest>,
) -> axum::response::Response {
    let correlation_id = state.logger.generate_correlation_id().await;

    // Use pre-compiled WASM if available, otherwise decode the WASM binary
    let wasm_bytes = if let Some(ref compiled) = payload.wasm_compiled {
        match base64_decode(compiled) {
            Ok(bytes) => {
                state.logger.log_with_correlation(
                    crate::logging::LogLevel::Debug,
                    "Using pre-compiled WASM bytes",
                    &correlation_id,
                );
                bytes
            }
            Err(e) => {
                return ErrorResponse {
                    error: format!("Failed to decode wasm_compiled: {}", e),
                    correlation_id: Some(correlation_id.to_string()),
                    recovery_suggestions: vec!["Ensure wasm_compiled is valid base64".to_string()],
                }
                .into_response()
            }
        }
    } else {
        match base64_decode(&payload.wasm_binary) {
            Ok(bytes) => bytes,
            Err(e) => {
                return ErrorResponse {
                    error: format!("Failed to decode wasm_binary: {}", e),
                    correlation_id: Some(correlation_id.to_string()),
                    recovery_suggestions: vec!["Ensure wasm_binary is valid base64".to_string()],
                }
                .into_response()
            }
        }
    };

    // Build a temporary config with per-request resource limits
    let mut exec_config = state.config.clone();
    // Use payload fields to override URL path parameters if provided
    exec_config.function = payload.function.clone().unwrap_or(function_id.clone());
    exec_config.version = payload.version.clone().unwrap_or(version.clone());
    if let Some(ms) = payload.timeout_ms {
        exec_config.timeout_ms = ms;
    }
    if let Some(mb) = payload.memory_mb {
        exec_config.memory_mb = mb;
    }
    if let Some(ref tid) = payload.tenant_id {
        exec_config.tenant_id = Some(tid.clone());
    }
    // Store execution context for downstream use
    let execution_context = payload.context.clone();
    let function_key = format!("{}@{}", exec_config.function, exec_config.version);

    let start = std::time::Instant::now();

    // Pool-aware execution:
    // 1. engine.execute() handles pool acquisition internally (fast path if pool is warm).
    // 2. If the pool is empty (cold start), engine.execute() falls back to execute_wasm_standard.
    // 3. After a successful cold-start execution, we pre-warm the pool so the next request is fast.
    let is_pool_warm = if let Some(ref pm) = state.wasm_pool {
        pm.is_warmed(&function_key).await
    } else {
        false
    };

    let result = state
        .engine
        .execute(&wasm_bytes, &payload.input, &exec_config, state.python_pool.clone(), state.micropython_executor.clone())
        .await;

    // Pre-warm the pool after a cold start (successful fallback execution)
    // The pre-warm is fire-and-forget at the pool level — we don't await it.
    if result.is_ok() && !is_pool_warm {
        if let Some(ref pool_manager) = state.wasm_pool {
            let wasm_bytes_for_prewarm = wasm_bytes.clone();
            let function_key_for_prewarm = function_key.clone();
            let pool_manager_clone = pool_manager.clone();
            let engine = state.engine.clone();
            tracing::debug!(function = %function_key_for_prewarm, "Pre-warming pool after cold start");
            tokio::spawn(async move {
                if let Err(e) = engine.prewarm_pool(&wasm_bytes_for_prewarm, &function_key_for_prewarm, pool_manager_clone).await {
                    tracing::warn!(function = %function_key_for_prewarm, "Failed to pre-warm pool: {}", e);
                }
            });
        }
    }

    // Handle result or error
    let result = match result {
        Ok(output) => output,
        Err(e) => {
            return ErrorResponse {
                error: format!("Execution failed: {}", e),
                correlation_id: Some(correlation_id.to_string()),
                recovery_suggestions: vec![
                    "Check that the WASM binary is valid".to_string(),
                    "Verify resource limits are sufficient".to_string(),
                ],
            }
            .into_response()
        }
    };

    let exec_time = start.elapsed().as_millis() as u64;

    state.logger.log_function_execution(
        &correlation_id,
        &exec_config.function,
        exec_time,
        true,
        false,
    );

    // Log context if provided (for debugging/tracing)
    if let Some(ref ctx) = execution_context {
        tracing::debug!(correlation_id = %correlation_id, context = ?ctx, "Execution context provided");
    }

    Json(ExecuteResponse {
        result,
        exec_time_ms: exec_time,
        cache_hit: false,
        instance_id: Uuid::new_v4().to_string(),
        function: exec_config.function.clone(),
        version: exec_config.version.clone(),
    })
    .into_response()
}

/// Simple base64 decoder (standard alphabet, no padding enforcement).
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
            b'=' => break, // padding
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
