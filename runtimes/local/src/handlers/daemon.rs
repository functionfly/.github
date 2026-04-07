//! Daemon-mode execution handler.

use axum::{extract::{Json, Path, State}, response::IntoResponse};
use std::sync::Arc;
use uuid::Uuid;

use super::types::{AppState, DaemonExecuteRequest, ErrorResponse, ExecuteResponse};

/// Execute a function via the daemon endpoint (Phase 3.2).
///
/// This handler is used by SandboxClient when the runtime runs in daemon mode.
/// It receives the WASM binary (base64) and per-request resource limits,
/// allowing a single runtime process to serve multiple functions with
/// independent memory/CPU/timeout constraints.
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

    let start = std::time::Instant::now();

    // Execute using the engine with per-request config
    let result = match state
        .engine
        .execute(&wasm_bytes, &payload.input, &exec_config, state.python_pool.clone(), state.micropython_executor.clone())
        .await
    {
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
