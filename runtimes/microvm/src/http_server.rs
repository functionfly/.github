//! HTTP API server for the MicroVM Orchestrator
//!
//! Exposes /execute, /health, /stats, and /metrics endpoints for the local runtime to invoke.
//!
//! **Security**: When `FUNCTIONFLY_MICROVM_API_TOKEN` is set, every request to `/execute` and
//! `/stats` must carry `Authorization: Bearer <token>`. `/health` and `/metrics` are open.

use axum::{
    body::Body,
    extract::State,
    http::{header, Request, StatusCode},
    middleware::{self, Next},
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::Instant;
use tokio::sync::RwLock;
use tracing::{error, info, warn};

use crate::orchestrator::{ExecutionRequest, MicroVMOrchestrator, OrchestratorStats};

/// Counters for basic observability (Prometheus-compatible text at /metrics).
#[derive(Default)]
pub struct HttpMetrics {
    pub execute_total: AtomicU64,
    pub execute_ok: AtomicU64,
    pub execute_err: AtomicU64,
    /// Cumulative wall-time across all executions (nanoseconds).
    pub execute_duration_ns: AtomicU64,
    /// Requests rejected by the pool-exhausted guard.
    pub pool_exhausted_total: AtomicU64,
    /// Failed Firecracker spawns.
    pub fc_spawn_failures: AtomicU64,
}

/// Shared application state
#[derive(Clone)]
pub struct AppState {
    pub orchestrator: Arc<RwLock<MicroVMOrchestrator>>,
    pub metrics: Arc<HttpMetrics>,
    /// Static bearer token required on sensitive endpoints. `None` means no auth (dev only).
    pub api_token: Option<Arc<str>>,
}

/// Execution request (matches OrchestratorClient::MicroVMExecutionRequest)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecuteRequest {
    pub code: String,
    pub input: String,
    pub handler: String,
    #[serde(default)]
    pub packages: Vec<String>,
    pub memory_mb: u32,
    pub vcpus: u32,
    pub timeout_ms: u64,
    pub tenant_id: String,
    /// Allowed outbound hostnames for the guest. Empty = no network.
    #[serde(default)]
    pub network_whitelist: Vec<String>,
    /// Reject connections to non-whitelisted hosts rather than just logging.
    #[serde(default)]
    pub strict_network_whitelist: bool,
    /// Enable per-tenant package caching inside the VM.
    #[serde(default)]
    pub package_caching_enabled: bool,
}

/// Execution response (matches OrchestratorClient::MicroVMExecutionResult)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecuteResponse {
    pub output: String,
    pub success: bool,
    pub error: Option<String>,
    pub execution_time_ms: u64,
    pub memory_used_mb: u32,
    pub stats: Option<OrchestratorStats>,
}

/// Bearer-token middleware. 401 when a token is configured and the request does not match.
async fn require_token(
    State(state): State<AppState>,
    req: Request<Body>,
    next: Next,
) -> Response {
    if let Some(expected) = &state.api_token {
        let bearer = req
            .headers()
            .get(header::AUTHORIZATION)
            .and_then(|v| v.to_str().ok())
            .and_then(|v| v.strip_prefix("Bearer "));
        match bearer {
            Some(tok) if tok == expected.as_ref() => {}
            _ => {
                warn!("Unauthorized /execute request (bad or missing bearer token)");
                return (StatusCode::UNAUTHORIZED, "Unauthorized").into_response();
            }
        }
    }
    next.run(req).await
}

/// Build the HTTP router.
///
/// `/health` and `/metrics` are unauthenticated (scraped by Prometheus).
/// `/execute` and `/stats` require a bearer token when `FUNCTIONFLY_MICROVM_API_TOKEN` is set.
pub fn router(state: AppState) -> Router {
    let protected = Router::new()
        .route("/execute", post(handle_execute))
        .route("/stats", get(handle_stats))
        .layer(middleware::from_fn_with_state(state.clone(), require_token));

    Router::new()
        .merge(protected)
        .route("/health", get(handle_health))
        .route("/metrics", get(handle_metrics))
        .with_state(state)
}

/// Handle POST /execute
async fn handle_execute(
    State(state): State<AppState>,
    Json(req): Json<ExecuteRequest>,
) -> (StatusCode, Json<ExecuteResponse>) {
    if req.tenant_id.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(ExecuteResponse {
                output: String::new(),
                success: false,
                error: Some("tenant_id is required".into()),
                execution_time_ms: 0,
                memory_used_mb: 0,
                stats: None,
            }),
        );
    }

    info!(
        tenant = %req.tenant_id,
        handler = %req.handler,
        memory_mb = req.memory_mb,
        vcpus = req.vcpus,
        "Execute request"
    );

    let exec_req = ExecutionRequest {
        code: req.code,
        input: req.input,
        handler: req.handler,
        packages: req.packages,
        memory_mb: req.memory_mb,
        vcpus: req.vcpus,
        timeout_ms: req.timeout_ms,
        network_whitelist: req.network_whitelist,
        strict_network_whitelist: req.strict_network_whitelist,
        package_caching_enabled: req.package_caching_enabled,
    };

    state.metrics.execute_total.fetch_add(1, Ordering::Relaxed);
    let wall = Instant::now();

    let result = {
        let mut orch = state.orchestrator.write().await;
        orch.execute(&req.tenant_id, exec_req).await
    };

    let elapsed_ns = wall.elapsed().as_nanos() as u64;
    state.metrics.execute_duration_ns.fetch_add(elapsed_ns, Ordering::Relaxed);

    match result {
        Ok(exec_result) => {
            if exec_result.success {
                state.metrics.execute_ok.fetch_add(1, Ordering::Relaxed);
            } else {
                state.metrics.execute_err.fetch_add(1, Ordering::Relaxed);
            }
            let resp = ExecuteResponse {
                output: exec_result.output,
                success: exec_result.success,
                error: exec_result.error,
                execution_time_ms: exec_result.execution_time_ms,
                memory_used_mb: exec_result.memory_used_mb,
                stats: Some(state.orchestrator.read().await.stats()),
            };
            (StatusCode::OK, Json(resp))
        }
        Err(e) => {
            let msg = e.to_string();
            error!(error = %msg, "Execution failed");
            if msg.contains("exhausted") {
                state.metrics.pool_exhausted_total.fetch_add(1, Ordering::Relaxed);
            }
            if msg.contains("spawn") || msg.contains("Firecracker") {
                state.metrics.fc_spawn_failures.fetch_add(1, Ordering::Relaxed);
            }
            state.metrics.execute_err.fetch_add(1, Ordering::Relaxed);
            let resp = ExecuteResponse {
                output: String::new(),
                success: false,
                error: Some(msg),
                execution_time_ms: 0,
                memory_used_mb: 0,
                stats: Some(state.orchestrator.read().await.stats()),
            };
            (StatusCode::INTERNAL_SERVER_ERROR, Json(resp))
        }
    }
}

/// Handle GET /health
async fn handle_health() -> StatusCode {
    StatusCode::OK
}

/// Handle GET /stats
async fn handle_stats(State(state): State<AppState>) -> Json<OrchestratorStats> {
    let orch = state.orchestrator.read().await;
    Json(orch.stats())
}

/// Handle GET /metrics (Prometheus text exposition format 0.0.4)
async fn handle_metrics(State(state): State<AppState>) -> impl IntoResponse {
    let t   = state.metrics.execute_total.load(Ordering::Relaxed);
    let ok  = state.metrics.execute_ok.load(Ordering::Relaxed);
    let err = state.metrics.execute_err.load(Ordering::Relaxed);
    let dur_ns = state.metrics.execute_duration_ns.load(Ordering::Relaxed);
    let dur_s  = dur_ns as f64 / 1_000_000_000.0;
    let pool_ex = state.metrics.pool_exhausted_total.load(Ordering::Relaxed);
    let fc_fail = state.metrics.fc_spawn_failures.load(Ordering::Relaxed);

    let stats = state.orchestrator.read().await.stats();

    let body = format!(
"# HELP functionfly_microvm_execute_total Total /execute invocations\n\
# TYPE functionfly_microvm_execute_total counter\n\
functionfly_microvm_execute_total {t}\n\
# HELP functionfly_microvm_execute_ok Successful executions\n\
# TYPE functionfly_microvm_execute_ok counter\n\
functionfly_microvm_execute_ok {ok}\n\
# HELP functionfly_microvm_execute_err Failed or unsuccessful executions\n\
# TYPE functionfly_microvm_execute_err counter\n\
functionfly_microvm_execute_err {err}\n\
# HELP functionfly_microvm_execute_duration_seconds_total Cumulative wall-time across executions\n\
# TYPE functionfly_microvm_execute_duration_seconds_total counter\n\
functionfly_microvm_execute_duration_seconds_total {dur_s:.6}\n\
# HELP functionfly_microvm_pool_exhausted_total Requests rejected due to pool exhaustion\n\
# TYPE functionfly_microvm_pool_exhausted_total counter\n\
functionfly_microvm_pool_exhausted_total {pool_ex}\n\
# HELP functionfly_microvm_fc_spawn_failures_total Failed Firecracker spawn attempts\n\
# TYPE functionfly_microvm_fc_spawn_failures_total counter\n\
functionfly_microvm_fc_spawn_failures_total {fc_fail}\n\
# HELP functionfly_microvm_active_vms Currently executing VMs\n\
# TYPE functionfly_microvm_active_vms gauge\n\
functionfly_microvm_active_vms {}\n\
# HELP functionfly_microvm_warm_vms VMs in warm pool\n\
# TYPE functionfly_microvm_warm_vms gauge\n\
functionfly_microvm_warm_vms {}\n\
# HELP functionfly_microvm_max_vms Configured VM capacity\n\
# TYPE functionfly_microvm_max_vms gauge\n\
functionfly_microvm_max_vms {}\n",
        stats.active_vms, stats.warm_vms, stats.max_vms
    );
    ([(axum::http::header::CONTENT_TYPE, "text/plain; version=0.0.4")], body)
}
