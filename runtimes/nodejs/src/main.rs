//! FunctionFly Node.js Runtime Binary
//!
//! Supports two modes:
//!   - CLI mode: execute code directly via --code flag
//!   - Daemon mode: run as HTTP server for the Go orchestrator (--daemon --port)

use std::sync::Arc;
use std::time::Duration;

use clap::{Parser, ValueEnum};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use nodejs_runtime::{
    Runtime, RuntimeConfig, RuntimeVersion, create_runtime, ExecutionInput, ExecutionMetadata,
    wasm_entry,
};

#[derive(Parser, Debug)]
#[command(name = "functionfly-nodejs")]
#[command(about = "FunctionFly Node.js Runtime", long_about = None)]
struct Args {
    /// Runtime version to use
    #[arg(short, long, value_enum, default_value = "node20")]
    runtime: RuntimeArg,

    /// Maximum memory in MB
    #[arg(short, long, default_value = "128")]
    memory: u32,

    /// Maximum timeout in milliseconds
    #[arg(short, long, default_value = "30000")]
    timeout: u64,

    /// Enable network access
    #[arg(long)]
    network: bool,

    /// Code to execute
    #[arg(short, long)]
    code: Option<String>,

    /// Input to pass to the function
    #[arg(short, long, default_value = "\"test\"")]
    input: String,

    /// Execute in REPL mode
    #[arg(short, long)]
    repl: bool,

    /// Verbose logging
    #[arg(short, long)]
    verbose: bool,

    /// Run as HTTP daemon for Go orchestrator
    #[arg(long)]
    daemon: bool,

    /// Port for daemon HTTP server
    #[arg(long, default_value = "9091")]
    port: u16,
}

#[derive(Debug, Clone, ValueEnum)]
enum RuntimeArg {
    Node18,
    Node20,
    Deno,
}

impl From<RuntimeArg> for RuntimeVersion {
    fn from(arg: RuntimeArg) -> Self {
        match arg {
            RuntimeArg::Node18 => RuntimeVersion::Node18,
            RuntimeArg::Node20 => RuntimeVersion::Node20,
            RuntimeArg::Deno => RuntimeVersion::Deno,
        }
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();

    // Set up logging
    let log_level = if args.verbose {
        tracing::Level::DEBUG
    } else {
        tracing::Level::INFO
    };

    tracing_subscriber::registry()
        .with(tracing_subscriber::fmt::layer().with_level(true))
        .with(tracing_subscriber::filter::LevelFilter::from_level(log_level))
        .init();

    tracing::info!("FunctionFly Node.js Runtime starting...");

    // Daemon mode: HTTP server for Go orchestrator
    if args.daemon {
        tracing::info!("Daemon mode — starting HTTP server on port {}", args.port);
        run_daemon(args.port, args.network).await?;
        return Ok(());
    }

    tracing::info!("Runtime: {:?}", args.runtime);
    tracing::info!("Memory limit: {}MB", args.memory);
    tracing::info!("Timeout: {}ms", args.timeout);

    // CLI mode
    let config = RuntimeConfig {
        version: args.runtime.into(),
        max_memory_mb: args.memory,
        max_timeout_ms: args.timeout,
        network_enabled: args.network,
        verbose_logging: args.verbose,
        ..Default::default()
    };

    config.validate()?;
    let runtime = create_runtime(config)?;

    let info = runtime.info();
    tracing::info!("Runtime info: {:?}", info.name);
    tracing::info!("Supported features: {:?}", info.features);

    if let Some(code) = args.code {
        let input = ExecutionInput {
            data: serde_json::Value::String(args.input),
            metadata: ExecutionMetadata::default(),
        };

        tracing::info!("Executing code...");
        let result = runtime.execute(&code, input).await;

        if result.success {
            tracing::info!("✓ Execution successful!");
            tracing::info!("Output: {:?}", result.output);
            tracing::info!("Execution time: {}ms", result.execution_time_ms);
        } else {
            tracing::error!("✗ Execution failed: {:?}", result.error);
            std::process::exit(1);
        }
    } else if args.repl {
        tracing::info!("Starting REPL mode (not implemented)");
    } else {
        tracing::info!("No code provided. Use --code or --daemon");
    }

    Ok(())
}

// ============================================================================
// HTTP Daemon Server
// ============================================================================

use axum::{
    Router, routing::post,
    extract::{State, Json, Path},
    http::StatusCode,
    response::IntoResponse,
};
use tower_http::trace::TraceLayer;
use std::net::SocketAddr;

/// Application state for the daemon HTTP server
#[derive(Clone)]
struct DaemonState {
    network_enabled: bool,
    api_token: Option<String>,
}

/// POST /health — liveness check
async fn health_handler() -> impl IntoResponse {
    Json(serde_json::json!({
        "status": "ok",
        "runtime": "nodejs",
        "version": env!("CARGO_PKG_VERSION"),
    }))
}

/// POST /execute/{function_id}/{version} — run a function
/// Request body: { wasm_binary, wasm_compiled, input, timeout_ms, memory_mb, tenant_id }
/// Response: { result, exec_time_ms, cache_hit }
async fn execute_handler(
    State(state): State<DaemonState>,
    headers: axum::http::HeaderMap,
    Path((function_id, version)): Path<(String, String)>,
    Json(body): Json<serde_json::Value>,
) -> impl IntoResponse {
    // Auth check
    if let Some(ref token) = state.api_token {
        let auth = headers.get("authorization")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");
        if auth != format!("Bearer {}", token) {
            return error_response(StatusCode::UNAUTHORIZED, "unauthorized");
        }
    }
    let wasm_binary = match body.get("wasm_binary").and_then(|v| v.as_str()) {
        Some(b64) => {
            // Decode base64-encoded WASM or source
            let bytes = match base64_decode(b64) {
                Ok(b) => b,
                Err(e) => {
                    return error_response(StatusCode::BAD_REQUEST, e);
                }
            };
            bytes
        }
        None => {
            return error_response(StatusCode::BAD_REQUEST, "missing wasm_binary field");
        }
    };

    // Detect jsDaemonBundle (JSON with source_code field) — this means the
    // Go bundler ran the JS via the daemon at bundle time and stored the result
    // (not the raw source or compiled WASM). We re-execute it here with the real input.
    if let Ok(bundle) = serde_json::from_slice::<jsDaemonBundle>(&wasm_binary) {
        tracing::debug!(
            "daemon bundle detected for function {} v{} — re-executing via QuickJS",
            function_id, version
        );
        let input = body.get("input")
            .and_then(|v| v.as_str())
            .unwrap_or("{}");

        let start = std::time::Instant::now();
        let result = wasm_entry::execute_js(&bundle.source_code, input);
        let exec_time_ms = start.elapsed().as_millis() as u64;

        match result {
            Ok(output) => {
                return (StatusCode::OK, Json(serde_json::json!({
                    "result": output,
                    "exec_time_ms": exec_time_ms,
                    "cache_hit": false,
                })));
            }
            Err(e) => {
                tracing::error!("daemon bundle re-execution error: {}", e);
                return error_response(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string());
            }
        }
    }

    // If it's a WASM binary (starts with \0asm), execute via wasmtime
    if wasm_binary.starts_with(&[0x00, 0x61, 0x73, 0x6D]) {
        let input = body.get("input")
            .and_then(|v| v.as_str())
            .unwrap_or("{}");

        let start = std::time::Instant::now();
        let result = wasm_entry::execute_wasm_binary(&wasm_binary, input);
        let exec_time_ms = start.elapsed().as_millis() as u64;

        match result {
            Ok(output) => {
                tracing::debug!("WASM execution completed in {}ms", exec_time_ms);
                return (StatusCode::OK, Json(serde_json::json!({
                    "result": output,
                    "exec_time_ms": exec_time_ms,
                    "cache_hit": false,
                })));
            }
            Err(e) => {
                tracing::error!("WASM execution error: {}", e);
                return error_response(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string());
            }
        }
    }

    // Otherwise, treat as raw JS source and execute via QuickJS
    let code = match String::from_utf8(wasm_binary.clone()) {
        Ok(s) => s,
        Err(e) => {
            return error_response(StatusCode::BAD_REQUEST, &format!("invalid UTF-8: {}", e));
        }
    };

    let input = body.get("input")
        .and_then(|v| v.as_str())
        .unwrap_or("{}");

    let start = std::time::Instant::now();
    let result = wasm_entry::execute_js(&code, input);
    let exec_time_ms = start.elapsed().as_millis() as u64;

    match result {
        Ok(output) => {
            tracing::debug!("JS execution completed in {}ms", exec_time_ms);
            (StatusCode::OK, Json(serde_json::json!({
                "result": output,
                "exec_time_ms": exec_time_ms,
                "cache_hit": false,
            })))
        }
        Err(e) => {
            tracing::error!("execute error: {}", e);
            error_response(StatusCode::INTERNAL_SERVER_ERROR, &e.to_string())
        }
    }
}

/// Lightweight bundle marker for daemon-pre-executed JS
#[derive(serde::Deserialize)]
struct jsDaemonBundle {
    source_code: String,
}

fn error_response(status: StatusCode, message: &str) -> (StatusCode, Json<serde_json::Value>) {
    (status, Json(serde_json::json!({
        "error": message,
    })))
}

/// Simple base64 decoder
fn base64_decode(input: &str) -> Result<Vec<u8>, &'static str> {
    // Use base64 engine decode
    use base64::Engine;
    base64::engine::general_purpose::STANDARD
        .decode(input)
        .map_err(|_| "invalid base64")
}

async fn run_daemon(port: u16, network_enabled: bool) -> Result<(), Box<dyn std::error::Error>> {
    // Initialize the global JS runtime
    tracing::info!("Initializing daemon JS runtime...");
    wasm_entry::init_daemon()
        .map_err(|e| {
            tracing::error!("Failed to initialize daemon: {}", e);
            format!("failed to init daemon: {}", e)
        })?;
    tracing::info!("Daemon JS runtime initialized successfully");

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

    let state = DaemonState { network_enabled, api_token };


    let app = Router::new()
        .route("/health", post(health_handler).get(health_handler))
        .route("/execute/{function_id}/{version}", post(execute_handler))
        .with_state(state)
        .layer(TraceLayer::new_for_http());


    let addr = SocketAddr::from(([127, 0, 0, 1], port));
    let listener = tokio::net::TcpListener::bind(addr).await?;
    tracing::info!("Daemon HTTP server listening on {}", addr);


    axum::serve(listener, app).await?;

    Ok(())
}
