//! FunctionFly Bun Runtime - Main binary
//!
//! Secure production-ready JavaScript/TypeScript execution runtime with
//! orchestrator communication via NATS.

use functionfly_bun_runtime::{
    init_tracing, config::RuntimeConfig,
    OrchestratorClient, Sandbox, SandboxConfig, SecurityManager,
    MetricsCollector, ExecutionLimits,
};
use parking_lot::RwLock;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::time::interval;
use tracing::{info, warn};

#[derive(Debug, Clone)]
struct Args {
    port: u16,
    max_concurrent: usize,
    max_memory_mb: u64,
    max_execution_time_secs: u64,
    sandbox_enabled: bool,
    nats_url: Option<String>,
    orchestrator_url: Option<String>,
    api_token: Option<String>,
    is_production: bool,
}

impl Default for Args {
    fn default() -> Self {
        Self {
            port: 8091,
            max_concurrent: 100,
            max_memory_mb: 512,
            max_execution_time_secs: 30,
            sandbox_enabled: true,
            nats_url: std::env::var("NATS_URL").ok(),
            orchestrator_url: std::env::var("ORCHESTRATOR_URL").ok(),
            api_token: None,
            is_production: false,
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
                .unwrap_or_else(|_| "8091".to_string())
                .parse()
                .unwrap_or(8091),
            max_concurrent: std::env::var("MAX_CONCURRENT")
                .unwrap_or_else(|_| "100".to_string())
                .parse()
                .unwrap_or(100),
            max_memory_mb: std::env::var("MAX_MEMORY_MB")
                .unwrap_or_else(|_| "512".to_string())
                .parse()
                .unwrap_or(512),
            max_execution_time_secs: std::env::var("MAX_EXECUTION_TIME_SECS")
                .unwrap_or_else(|_| "30".to_string())
                .parse()
                .unwrap_or(30),
            sandbox_enabled,
            nats_url: std::env::var("NATS_URL").ok(),
            orchestrator_url: std::env::var("ORCHESTRATOR_URL").ok(),
            api_token: std::env::var("RUNTIME_API_TOKEN").ok().filter(|t| !t.is_empty()),
            is_production,
        }
    }
}

/// Runtime state with orchestrator client
struct RuntimeState {
    sandbox: Arc<Sandbox>,
    orchestrator: Arc<RwLock<OrchestratorClient>>,
    metrics: Arc<MetricsCollector>,
    total_executions: Arc<std::sync::atomic::AtomicU64>,
    api_token: Option<String>,
}

impl RuntimeState {
    async fn execute_function(
        &self,
        execution_id: &str,
        code: &str,
        _input: Option<serde_json::Value>,
        timeout: Duration,
    ) -> Result<serde_json::Value, String> {
        let start = Instant::now();

        // Execute via sandbox
        let result = self.sandbox.execute(code, timeout).await
            .map_err(|e| e.to_string())?;

        if result.success {
            let execution_time_ms = start.elapsed().as_millis() as u64;
            self.total_executions.fetch_add(1, std::sync::atomic::Ordering::Relaxed);

            Ok(serde_json::json!({
                "execution_id": execution_id,
                "success": true,
                "output": result.output,
                "execution_time_ms": execution_time_ms,
                "memory_used_mb": result.memory_used_mb,
            }))
        } else {
            Err(result.error.unwrap_or_else(|| "unknown error".to_string()))
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
        sandbox_enabled = args.sandbox_enabled,
        nats_url = ?args.nats_url,
        orchestrator_url = ?args.orchestrator_url,
        "Starting FunctionFly Bun Runtime"
    );

    if args.api_token.is_none() {
        if args.is_production {
            tracing::error!(
                "RUNTIME_API_TOKEN is not set in production. \
                 The /execute endpoint is UNAUTHENTICATED. Set the token and restart."
            );
        } else {
            warn!("RUNTIME_API_TOKEN not set — /execute endpoint is unauthenticated (dev mode)");
        }
    }

    let config = RuntimeConfig {
        limits: ExecutionLimits {
            max_memory_mb: args.max_memory_mb,
            max_cpu_time_secs: args.max_execution_time_secs,
            max_wall_time_secs: args.max_execution_time_secs,
            ..Default::default()
        },
        security: functionfly_bun_runtime::config::SecurityPolicy {
            sandbox_enabled: args.sandbox_enabled,
            ..Default::default()
        },
        use_sandbox: args.sandbox_enabled,
        max_concurrent: args.max_concurrent,
        default_timeout: Duration::from_secs(args.max_execution_time_secs),
    };

    // Create security manager and sandbox
    let security = Arc::new(SecurityManager::new(config.security.clone()));
    let sandbox_config = SandboxConfig::default();
    let sandbox = Arc::new(Sandbox::new(
        sandbox_config,
        config.limits.clone(),
        security,
    ));

    // Create orchestrator client
    let mut orchestrator = OrchestratorClient::new("bun");
    if let Some(ref nats_url) = args.nats_url {
        orchestrator = orchestrator.with_nats_url(nats_url);
        if let Err(e) = orchestrator.connect() {
            warn!(error = %e, "Failed to connect to NATS, running in standalone mode");
        } else {
            // Register with orchestrator
            if let Err(e) = orchestrator.register_runtime(vec![]) {
                warn!(error = %e, "Failed to register with orchestrator");
            } else {
                info!(runtime_id = %orchestrator.runtime_id(), "Registered with orchestrator");
            }
        }
    }

    let orchestrator = Arc::new(RwLock::new(orchestrator));
    let metrics = Arc::new(MetricsCollector::new());
    let total_executions = Arc::new(std::sync::atomic::AtomicU64::new(0));

    let state = RuntimeState {
        sandbox,
        orchestrator,
        metrics,
        total_executions,
        api_token: args.api_token.clone(),
    };

    // Start background tasks
    let orchestrator_for_heartbeat = state.orchestrator.clone();
    let orchestrator_for_metrics = state.orchestrator.clone();
    let total_exec_clone = state.total_executions.clone();

    // Heartbeat task
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(30));
        loop {
            interval.tick().await;

            let client = orchestrator_for_heartbeat.read();
            if client.is_registered() {
                if let Err(e) = client.send_heartbeat("healthy") {
                    warn!(error = %e, "Failed to send heartbeat");
                }
            }
        }
    });

    // Metrics reporting task
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(60));
        loop {
            interval.tick().await;

            let client = orchestrator_for_metrics.read();
            if client.is_registered() {
                let total = total_exec_clone.load(std::sync::atomic::Ordering::Relaxed);
                if let Err(e) = client.report_metrics(0.0, 0, total) {
                    warn!(error = %e, "Failed to report metrics");
                }
            }
        }
    });

    // Run the HTTP server
    run_server_with_state(config, args.port, state).await?;

    Ok(())
}

/// Run server with additional runtime state
async fn run_server_with_state(
    _config: RuntimeConfig,
    port: u16,
    state: RuntimeState,
) -> anyhow::Result<()> {
    use axum::{Router, Json, routing::{get, post}, extract::State};
    use std::net::SocketAddr;

    #[derive(Clone)]
    struct AppState {
        inner: Arc<RuntimeState>,
    }

    async fn health_handler() -> Json<serde_json::Value> {
        Json(serde_json::json!({
            "status": "healthy",
            "runtime": "bun",
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
        let code = req.get("code")
            .and_then(|v| v.as_str())
            .ok_or_else(|| (axum::http::StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "code required"}))))?;
        let input = req.get("input").cloned();
        let timeout_ms = req.get("timeout_ms")
            .and_then(|v| v.as_u64())
            .unwrap_or(30000);

        match state.inner.execute_function(
            execution_id,
            code,
            input,
            Duration::from_millis(timeout_ms),
        ).await {
            Ok(result) => Ok(Json(result)),
            Err(e) => Err((axum::http::StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({
                "execution_id": execution_id,
                "success": false,
                "error": e,
            })))),
        }
    }

    async fn metrics_handler(State(state): State<AppState>) -> Json<serde_json::Value> {
        let total = state.inner.total_executions.load(std::sync::atomic::Ordering::Relaxed);
        Json(serde_json::json!({
            "total_executions": total,
            "runtime": "bun",
        }))
    }

    async fn ready_handler(State(state): State<AppState>) -> Json<serde_json::Value> {
        let orchestrator = state.inner.orchestrator.read();
        Json(serde_json::json!({
            "ready": true,
            "registered": orchestrator.is_registered(),
            "nats_connected": orchestrator.is_connected(),
        }))
    }

    let app_state = AppState { inner: Arc::new(state) };

    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        .route("/metrics", get(metrics_handler))
        .route("/execute", post(execute_handler))
        .with_state(app_state);

    let addr: SocketAddr = format!("0.0.0.0:{}", port).parse()?;
    let listener = tokio::net::TcpListener::bind(addr).await?;

    info!(port = port, "Bun Runtime HTTP server started");
    axum::serve(listener, app).await?;

    Ok(())
}