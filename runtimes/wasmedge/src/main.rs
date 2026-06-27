//! FunctionFly WasmEdge Runtime - Main binary
//!
//! Production-ready C/C++ and WebAssembly execution runtime with:
//! - WasmEdge WASI 0.2 support for C/C++
//! - Orchestrator communication via NATS
//! - Comprehensive security controls
//! - Resource limits and metrics

use functionfly_wasmedge_runtime::{
    init_tracing,
    Sandbox, SandboxConfig, SecurityManager,
};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tracing::info;

#[derive(Debug, Clone)]
struct Args {
    port: u16,
    max_concurrent: usize,
    max_memory_mb: u64,
    max_fuel: u64,
    max_execution_time_secs: u64,
    sandbox_enabled: bool,
    #[allow(dead_code)]
    nats_url: Option<String>,
    #[allow(dead_code)]
    orchestrator_url: Option<String>,
    working_dir: Option<String>,
}

impl Default for Args {
    fn default() -> Self {
        Self {
            port: 8092,
            max_concurrent: 100,
            max_memory_mb: 512,
            max_fuel: 10_000_000,
            max_execution_time_secs: 30,
            sandbox_enabled: true,
            nats_url: std::env::var("NATS_URL").ok(),
            orchestrator_url: std::env::var("ORCHESTRATOR_URL").ok(),
            working_dir: std::env::var("WORKING_DIR").ok(),
        }
    }
}

impl Args {
    fn from_env() -> Self {
        let is_production = std::env::var("ENVIRONMENT")
            .map(|v| v.eq_ignore_ascii_case("production"))
            .unwrap_or(false);

        // In production, sandbox is always enabled regardless of env var
        let sandbox_enabled = if is_production {
            true
        } else {
            std::env::var("SANDBOX_ENABLED")
                .unwrap_or_else(|_| "true".to_string())
                .to_lowercase() != "false"
        };

        Self {
            port: std::env::var("PORT")
                .unwrap_or_else(|_| "8092".to_string())
                .parse()
                .unwrap_or(8092),
            max_concurrent: std::env::var("MAX_CONCURRENT")
                .unwrap_or_else(|_| "100".to_string())
                .parse()
                .unwrap_or(100),
            max_memory_mb: std::env::var("MAX_MEMORY_MB")
                .unwrap_or_else(|_| "512".to_string())
                .parse()
                .unwrap_or(512),
            max_fuel: std::env::var("MAX_FUEL")
                .unwrap_or_else(|_| "10000000".to_string())
                .parse()
                .unwrap_or(10_000_000),
            max_execution_time_secs: std::env::var("MAX_EXECUTION_TIME_SECS")
                .unwrap_or_else(|_| "30".to_string())
                .parse()
                .unwrap_or(30),
            sandbox_enabled,
            nats_url: std::env::var("NATS_URL").ok(),
            orchestrator_url: std::env::var("ORCHESTRATOR_URL").ok(),
            working_dir: std::env::var("WORKING_DIR").ok(),
        }
    }
}

/// Runtime state
struct RuntimeState {
    sandbox: Arc<Sandbox>,
    total_executions: Arc<std::sync::atomic::AtomicU64>,
    successful_executions: Arc<std::sync::atomic::AtomicU64>,
    failed_executions: Arc<std::sync::atomic::AtomicU64>,
    api_token: Option<String>,
}

impl RuntimeState {
    async fn execute_wasm(
        &self,
        execution_id: &str,
        wasm_bytes: &[u8],
        timeout: Duration,
    ) -> Result<serde_json::Value, String> {
        let start = Instant::now();

        // Execute via sandbox
        let result = self.sandbox.execute(wasm_bytes, timeout)
            .await
            .map_err(|e| e.to_string())?;

        let execution_time_ms = start.elapsed().as_millis() as u64;

        if result.success {
            self.total_executions.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
            self.successful_executions.fetch_add(1, std::sync::atomic::Ordering::Relaxed);

            Ok(serde_json::json!({
                "execution_id": execution_id,
                "success": true,
                "output": result.output,
                "execution_time_ms": execution_time_ms,
                "memory_used_mb": result.memory_used_mb,
                "fuel_consumed": result.fuel_consumed,
            }))
        } else {
            self.total_executions.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
            self.failed_executions.fetch_add(1, std::sync::atomic::Ordering::Relaxed);

            Err(serde_json::json!({
                "execution_id": execution_id,
                "success": false,
                "error": result.error,
                "termination_reason": result.termination_reason,
                "execution_time_ms": execution_time_ms,
            }).to_string())
        }
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    init_tracing();

    let args = Args::from_env();

    info!(
        version = env!("CARGO_PKG_VERSION"),
        port = args.port,
        max_concurrent = args.max_concurrent,
        max_memory_mb = args.max_memory_mb,
        max_fuel = args.max_fuel,
        sandbox_enabled = args.sandbox_enabled,
        "Starting FunctionFly WasmEdge Runtime"
    );

    let api_token = std::env::var("RUNTIME_API_TOKEN").ok().filter(|t| !t.is_empty());
    let is_production = std::env::var("ENVIRONMENT")
        .map(|v| v.eq_ignore_ascii_case("production"))
        .unwrap_or(false);

    if api_token.is_none() {
        if is_production {
            tracing::error!(
                "RUNTIME_API_TOKEN is not set in production. \
                 The /execute endpoint is UNAUTHENTICATED. Set the token and restart."
            );
        } else {
            tracing::warn!("RUNTIME_API_TOKEN not set — /execute endpoint is unauthenticated (dev mode)");
        }
    }

    let limits = functionfly_wasmedge_runtime::config::ExecutionLimits {
        max_memory_mb: args.max_memory_mb,
        max_cpu_time_secs: args.max_execution_time_secs,
        max_wall_time_secs: args.max_execution_time_secs,
        max_fuel: args.max_fuel,
        ..Default::default()
    };

    let security = Arc::new(SecurityManager::default());

    let sandbox_config = SandboxConfig {
        enable_fuel_metering: true,
        working_dir: args.working_dir,
        ..Default::default()
    };

    let sandbox = Arc::new(Sandbox::new(
        sandbox_config,
        limits,
        security,
    ));

    let state = RuntimeState {
        sandbox,
        total_executions: Arc::new(std::sync::atomic::AtomicU64::new(0)),
        successful_executions: Arc::new(std::sync::atomic::AtomicU64::new(0)),
        failed_executions: Arc::new(std::sync::atomic::AtomicU64::new(0)),
        api_token,
    };

    // Run the HTTP server
    run_server_with_state(args.port, state).await;

    Ok(())
}

/// Run server with runtime state
async fn run_server_with_state(port: u16, state: RuntimeState) {
    use axum::{Router, Json, routing::{get, post}, extract::State};
    use std::net::SocketAddr;

    #[derive(Clone)]
    struct AppState {
        inner: Arc<RuntimeState>,
    }

    async fn health_handler() -> Json<serde_json::Value> {
        Json(serde_json::json!({
            "status": "healthy",
            "runtime": "wasmedge",
            "version": env!("CARGO_PKG_VERSION"),
        }))
    }

    async fn execute_handler(
        State(state): State<AppState>,
        headers: axum::http::HeaderMap,
        Json(req): Json<serde_json::Value>,
    ) -> Result<Json<serde_json::Value>, (axum::http::StatusCode, Json<serde_json::Value>)> {
        // Auth check
        if let Some(ref token) = state.inner.api_token {
            let auth = headers.get("authorization")
                .and_then(|v| v.to_str().ok())
                .unwrap_or("");
            if auth != format!("Bearer {}", token) {
                return Err((axum::http::StatusCode::UNAUTHORIZED, Json(serde_json::json!({"error": "unauthorized"}))));
            }
        }

        let execution_id = req.get("execution_id")
            .and_then(|v| v.as_str())
            .unwrap_or("unknown");

        // Get WASM bytes (base64 encoded or raw)
        let wasm_input = req.get("wasm")
            .or_else(|| req.get("code"))
            .ok_or_else(|| (axum::http::StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "wasm code required"}))))?;

        let wasm_bytes = if let Some(b64) = wasm_input.as_str() {
            // Decode base64 if it's a string
            base64::Engine::decode(&base64::engine::general_purpose::STANDARD, b64)
                .map_err(|e| (axum::http::StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": format!("invalid base64: {}", e)}))))?
        } else if let Some(arr) = wasm_input.as_array() {
            // Accept as array of bytes
            arr.iter().map(|v| v.as_u64().unwrap_or(0) as u8).collect()
        } else {
            return Err((axum::http::StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "wasm must be base64 or byte array"}))));
        };

        let timeout_ms = req.get("timeout_ms")
            .and_then(|v| v.as_u64())
            .unwrap_or(30000);

        match state.inner.execute_wasm(
            execution_id,
            &wasm_bytes,
            Duration::from_millis(timeout_ms),
        ).await {
            Ok(result) => Ok(Json(result)),
            Err(e) => {
                // Parse error JSON if it's a serde_json error
                if let Ok(err_json) = serde_json::from_str::<serde_json::Value>(&e) {
                    Err((axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(err_json)))
                } else {
                    Err((axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({
                        "execution_id": execution_id,
                        "success": false,
                        "error": e,
                    }))))
                }
            }
        }
    }

    async fn metrics_handler(State(state): State<AppState>) -> Json<serde_json::Value> {
        let total = state.inner.total_executions.load(std::sync::atomic::Ordering::Relaxed);
        let successful = state.inner.successful_executions.load(std::sync::atomic::Ordering::Relaxed);
        let failed = state.inner.failed_executions.load(std::sync::atomic::Ordering::Relaxed);

        Json(serde_json::json!({
            "total_executions": total,
            "successful_executions": successful,
            "failed_executions": failed,
            "runtime": "wasmedge",
        }))
    }

    async fn ready_handler() -> Json<serde_json::Value> {
        Json(serde_json::json!({
            "ready": true,
            "runtime": "wasmedge",
        }))
    }

    let app_state = AppState { inner: Arc::new(state) };

    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        .route("/metrics", get(metrics_handler))
        .route("/execute", post(execute_handler))
        .with_state(app_state);

    let addr: SocketAddr = format!("0.0.0.0:{}", port).parse().expect("invalid address");
    let listener = tokio::net::TcpListener::bind(addr).await.expect("failed to bind");

    info!(port = port, "WasmEdge Runtime HTTP server started");
    axum::serve(listener, app).await.expect("server error");
}
