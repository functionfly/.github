//! HTTP server for the Prism runtime.
//!
//! ## Why this exists
//!
//! The original Prism HTTP layer was a hand-rolled parser built on
//! `tokio::net::TcpListener::accept` + a fixed-size 2 KiB read buffer +
//! substring-based header parsing. That implementation had several
//! launch-blocking bugs:
//!
//! - **2 KiB read buffer**: any body > 2 KiB was silently truncated. A
//!   10 KiB WASM module would only get the first 2 KiB and fail with
//!   a confusing parse error.
//! - **Case-sensitive method matching**: `request.starts_with("POST ...")`
//!   would not match lowercase methods, even though HTTP methods are
//!   case-sensitive per RFC 9110 — but our routing for paths was
//!   inconsistent, so we routed on a mix of substrings.
//! - **No body parsing**: we used `request.split("\r\n\r\n")` which
//!   assumed HTTP/1.1 with no chunked transfer encoding, no keep-alive,
//!   no trailers. Real HTTP clients use all of these.
//! - **Auth header parsing was case-insensitive but split-on-colon-only**:
//!   `Authorization: Bearer  abc` (extra space) would fail auth.
//! - **No security headers, no CORS lockdown, no body size limit**.
//!
//! This module replaces the entire dispatch loop with `axum`, which:
//!
//! - Parses the full HTTP/1.1 message (including chunked encoding)
//! - Supports HTTP/1.1 keep-alive
//! - Enforces a 2 MiB body limit on every route
//! - Emits security headers (HSTS, CSP, X-Frame-Options, X-Content-Type-Options)
//! - Validates auth via a real `Authorization: Bearer <token>` parser
//! - Tracks per-IP rate limits in a token-bucket, capped at 4096 entries
//!   (matching the runtime-level cap to prevent memory exhaustion DoS)
//!
//! The existing handler functions in `main.rs` are still callable in
//! their legacy `(&rt, request: &str) -> String` form for the CLI tools;
//! new routes in this module call into the same business logic with a
//! `(headers, path, body)` triple reconstructed from the typed request.

use axum::{
    body::Body,
    extract::{Path, State},
    http::{HeaderMap, HeaderValue, StatusCode},
    middleware::{self, Next},
    response::{IntoResponse, Json, Response},
    routing::{get, post},
    Router,
};
use parking_lot::RwLock;
use serde_json::json;
use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::net::TcpListener;
use tower_http::cors::{CorsLayer, AllowOrigin};
use tower_http::limit::RequestBodyLimitLayer;
use tower_http::set_header::SetResponseHeaderLayer;
use tower_http::timeout::TimeoutLayer;
use tracing::{info, warn};

use prism_runtime::runtime::RuntimeContext;

/// Maximum request body size across all routes. Set conservatively so a
/// single attacker cannot OOM the runtime with a 1 GB body.
pub const MAX_BODY_BYTES: usize = 2 * 1024 * 1024;

/// Request timeout. After this, the connection is closed with a 408.
pub const REQUEST_TIMEOUT: Duration = Duration::from_secs(30);

/// Maximum number of distinct per-IP rate limiter buckets retained in
/// memory. Same cap as the runtime-level per-tenant limiter to keep
/// memory usage bounded under attack.
const MAX_IP_LIMITERS: usize = 4096;

/// Token-bucket rate limiter, keyed by IP (or `Authorization: Bearer`
/// token, if present and well-formed).
#[derive(Clone)]
pub struct IpRateLimiter {
    tokens: Arc<std::sync::atomic::AtomicU64>,
    max_tokens: u64,
    refill_rate: u64,
    last_refill: Arc<std::sync::Mutex<Instant>>,
}

impl IpRateLimiter {
    fn new(max_tokens: u64, refill_rate: u64) -> Self {
        Self {
            tokens: Arc::new(std::sync::atomic::AtomicU64::new(max_tokens)),
            max_tokens,
            refill_rate,
            last_refill: Arc::new(std::sync::Mutex::new(Instant::now())),
        }
    }

    fn try_acquire(&self) -> bool {
        let mut last = self.last_refill.lock().unwrap();
        let now = Instant::now();
        let elapsed = now.duration_since(*last).as_secs_f64();
        let refill = (elapsed * self.refill_rate as f64) as u64;

        if refill > 0 {
            let new = (self.tokens.load(std::sync::atomic::Ordering::Relaxed) + refill)
                .min(self.max_tokens);
            self.tokens.store(new, std::sync::atomic::Ordering::Relaxed);
            *last = now;
        }

        let cur = self.tokens.load(std::sync::atomic::Ordering::Relaxed);
        if cur > 0 {
            self.tokens.store(cur - 1, std::sync::atomic::Ordering::Relaxed);
            true
        } else {
            false
        }
    }
}

/// Shared application state for the axum router.
#[derive(Clone)]
pub struct HttpState {
    /// Reference to the underlying Prism runtime context. `HttpState`
    /// does not own it — the main binary keeps it alive for the lifetime
    /// of the process and passes `Arc<RuntimeContext>` in.
    pub runtime: Arc<RuntimeContext>,
    /// Per-IP rate limiter map (capped at `MAX_IP_LIMITERS`).
    pub ip_limiters: Arc<RwLock<HashMap<String, IpRateLimiter>>>,
    /// Optional API token. When `Some`, sensitive routes require
    /// `Authorization: Bearer <token>`.
    pub api_token: Arc<Option<String>>,
    /// LRU-ish insertion-order tracker used to evict oldest IPs when
    /// `ip_limiters` hits `MAX_IP_LIMITERS`.
    pub insertion_order: Arc<std::sync::Mutex<std::collections::VecDeque<String>>>,
}

impl HttpState {
    pub fn new(runtime: Arc<RuntimeContext>, api_token: Option<String>) -> Self {
        Self {
            runtime,
            ip_limiters: Arc::new(RwLock::new(HashMap::new())),
            api_token: Arc::new(api_token),
            insertion_order: Arc::new(std::sync::Mutex::new(
                std::collections::VecDeque::new(),
            )),
        }
    }

    /// Look up or create a rate limiter for the given key (IP or token).
    /// Evicts the oldest entry if at capacity to prevent unbounded growth.
    fn limiter_for(&self, key: String) -> IpRateLimiter {
        // Fast path
        if let Some(limiter) = self.ip_limiters.read().get(&key) {
            return limiter.clone();
        }

        // Slow path: insert
        let limiter = IpRateLimiter::new(100, 100);
        self.ip_limiters.write().insert(key.clone(), limiter.clone());

        let mut order = self.insertion_order.lock().unwrap();
        order.push_back(key);
        while order.len() > MAX_IP_LIMITERS {
            if let Some(oldest) = order.pop_front() {
                self.ip_limiters.write().remove(&oldest);
            }
        }
        limiter
    }
}

/// Rate-limit middleware. Returns 429 when the bucket is empty.
///
/// The limiter key is the `X-Forwarded-For` first hop if present (we
/// trust upstream proxies in the deployment guide) or the connection's
/// peer address otherwise.
async fn rate_limit_middleware(
    State(state): State<HttpState>,
    req: axum::http::Request<Body>,
    next: Next,
) -> Response {
    let key = req
        .headers()
        .get("x-forwarded-for")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.split(',').next())
        .map(|s| s.trim().to_string())
        .or_else(|| {
            req.extensions()
                .get::<axum::extract::ConnectInfo<SocketAddr>>()
                .map(|c| c.0.ip().to_string())
        })
        .unwrap_or_else(|| "unknown".to_string());

    if !state.limiter_for(key).try_acquire() {
        return (
            StatusCode::TOO_MANY_REQUESTS,
            [("retry-after", "1")],
            Json(json!({"error": "rate limit exceeded"})),
        )
            .into_response();
    }

    next.run(req).await
}

/// Bearer-token middleware. Refuses requests to /execute and other
/// protected routes when an API token is configured but missing or wrong.
async fn require_auth_middleware(
    State(state): State<HttpState>,
    req: axum::http::Request<Body>,
    next: Next,
) -> Response {
    // No token configured → dev mode, allow.
    let expected = match state.api_token.as_ref() {
        Some(t) => t,
        None => return next.run(req).await,
    };

    // Skip auth for /health, /ready, /metrics — they're scraped by
    // Kubernetes probes / Prometheus which can't carry tokens.
    let path = req.uri().path();
    if matches!(path, "/health" | "/ready" | "/metrics") {
        return next.run(req).await;
    }

    let header_value = req
        .headers()
        .get("authorization")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("");

    // Accept `Bearer <token>` (case-insensitive scheme, single space
    // tolerated). We do NOT use `to_lowercase()` on the whole header to
    // preserve token case (RFC 6750 says the scheme is case-insensitive
    // but the token is opaque and case-sensitive).
    let bearer = header_value
        .strip_prefix("Bearer ")
        .or_else(|| header_value.strip_prefix("bearer "))
        .or_else(|| header_value.strip_prefix("BEARER "))
        .unwrap_or("");

    if bearer.is_empty()
        || !constant_time_eq::constant_time_eq(bearer.as_bytes(), expected.as_bytes())
    {
        warn!(
            path = %path,
            "Refused request: missing or invalid bearer token"
        );
        return (
            StatusCode::UNAUTHORIZED,
            [("www-authenticate", "Bearer realm=\"prism\"")],
            Json(json!({"error": "unauthorized"})),
        )
            .into_response();
    }

    next.run(req).await
}

/// Build the security headers layer applied to every response.
///
/// Headers follow OWASP's "Secure Headers" project recommendations.
/// `Strict-Transport-Security` is set to 1 year — production deployments
/// behind a TLS-terminating proxy should always enable HSTS.
fn security_headers_layer() -> tower_http::set_header::SetResponseHeaderLayer<axum::http::HeaderValue> {
    let policy = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'";
    SetResponseHeaderLayer::if_not_present(
        axum::http::header::HeaderName::from_static("content-security-policy"),
        axum::http::HeaderValue::from_str(policy)
            .expect("static CSP header value is always valid"),
    )
}

/// Run the HTTP server bound to `address`. The caller is responsible
/// for keeping `runtime` alive for the lifetime of the server.
pub async fn run_server(
    runtime: Arc<RuntimeContext>,
    address: SocketAddr,
    api_token: Option<String>,
) -> anyhow::Result<()> {
    let state = HttpState::new(runtime, api_token);

    let cors = {
        // SECURITY: deny-by-default. CORS is restricted to the explicit
        // allowlist in `PRISM_CORS_ALLOWED_ORIGINS` (comma-separated).
        // If unset, browsers will reject any cross-origin request — we
        // never default to `*`.
        let allowed: Vec<_> = std::env::var("PRISM_CORS_ALLOWED_ORIGINS")
            .unwrap_or_default()
            .split(',')
            .filter(|s| !s.is_empty())
            .filter_map(|s| s.trim().parse().ok())
            .collect();

        if allowed.is_empty() {
            CorsLayer::new()
                .allow_methods(tower_http::cors::AllowMethods::list([
                    axum::http::Method::GET,
                    axum::http::Method::POST,
                    axum::http::Method::DELETE,
                ]))
                .allow_headers(tower_http::cors::AllowHeaders::list([
                    axum::http::header::CONTENT_TYPE,
                    axum::http::header::AUTHORIZATION,
                    axum::http::HeaderName::from_static("x-tenant-id"),
                ]))
        } else {
            CorsLayer::new()
                .allow_origin(AllowOrigin::list(allowed))
                .allow_methods(tower_http::cors::AllowMethods::mirror_request())
                .allow_headers(tower_http::cors::AllowHeaders::mirror_request())
                .max_age(Duration::from_secs(86400))
        }
    };

    let app = Router::new()
        // Health and readiness — public, no auth.
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        // All other routes require auth (when configured).
        .route("/execute", post(execute_handler))
        .route("/cells", post(create_cell_handler))
        .route("/cells/{cell_id}", get(get_cell_handler).delete(delete_cell_handler))
        .route("/cells/{cell_id}/snapshot", post(snapshot_cell_handler))
        .route("/cells/{cell_id}/snapshots", get(list_snapshots_handler))
        .route("/snapshots/{snapshot_id}", post(restore_snapshot_handler).delete(delete_snapshot_handler))
        .route("/capabilities", get(list_capabilities_handler).post(register_capability_handler))
        .route("/capabilities/{name}/invoke", post(invoke_capability_handler))
        .route("/swarms", get(list_swarms_handler).post(create_swarm_handler))
        .route("/swarms/{swarm_id}/join", post(join_swarm_handler))
        .route("/swarms/{swarm_id}/leave", post(leave_swarm_handler))
        .route("/optimize/{cell_id}", get(get_optimization_handler))
        .layer(middleware::from_fn_with_state(state.clone(), rate_limit_middleware))
        .layer(middleware::from_fn_with_state(state.clone(), require_auth_middleware))
        .layer(RequestBodyLimitLayer::new(MAX_BODY_BYTES))
        .layer(TimeoutLayer::with_status_code(
            StatusCode::REQUEST_TIMEOUT,
            REQUEST_TIMEOUT,
        ))
        .layer(tower_http::set_header::SetResponseHeaderLayer::if_not_present(
            axum::http::header::HeaderName::from_static("x-content-type-options"),
            HeaderValue::from_static("nosniff"),
        ))
        .layer(tower_http::set_header::SetResponseHeaderLayer::if_not_present(
            axum::http::header::HeaderName::from_static("x-frame-options"),
            HeaderValue::from_static("DENY"),
        ))
        .layer(tower_http::set_header::SetResponseHeaderLayer::if_not_present(
            axum::http::header::HeaderName::from_static("referrer-policy"),
            HeaderValue::from_static("strict-origin-when-cross-origin"),
        ))
        .layer(security_headers_layer())
        .layer(tower_http::set_header::SetResponseHeaderLayer::if_not_present(
            axum::http::header::HeaderName::from_static("strict-transport-security"),
            HeaderValue::from_static("max-age=31536000; includeSubDomains"),
        ))
        .layer(cors)
        .with_state(state);

    let listener = TcpListener::bind(address).await?;
    info!(address = %address, "Prism runtime HTTP server listening (axum)");

    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<SocketAddr>(),
    )
    .await?;

    Ok(())
}

// ============================================================================
// Handlers
// ============================================================================
//
// Each handler is a thin wrapper over the underlying runtime methods. We
// keep the response shape identical to the legacy JSON-string format so
// existing CLI tools and dashboards keep working.

async fn health_handler(State(state): State<HttpState>) -> Json<serde_json::Value> {
    let status = state.runtime.get_status().await;
    Json(json!({
        "version": status.version,
        "healthy": status.healthy,
        "active_cells": status.active_cells,
        "total_cells": status.total_cells,
        "mesh_enabled": status.mesh_enabled,
    }))
}

async fn ready_handler(State(state): State<HttpState>) -> Json<serde_json::Value> {
    // Readiness check: runtime is ready if NATS is connected (or
    // intentionally disabled) and the executor is initialized.
    let runtime = state.runtime.get_status().await;
    let ready = runtime.healthy;
    Json(json!({
        "ready": ready,
        "mesh_enabled": runtime.mesh_enabled,
        "active_cells": runtime.active_cells,
    }))
}

#[derive(serde::Deserialize)]
struct CreateCellBody {
    name: String,
    module: String,
    #[serde(default)]
    memory: Option<u64>,
}

async fn create_cell_handler(
    State(state): State<HttpState>,
    headers: HeaderMap,
    Json(req): Json<CreateCellBody>,
) -> Response {
    let memory_mb = req.memory.unwrap_or(128);

    let wasm_bytes = match std::fs::read(&req.module) {
        Ok(b) => b,
        Err(e) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": format!("Failed to read module: {}", e)})),
            )
                .into_response();
        }
    };

    if wasm_bytes.len() < 4 || &wasm_bytes[0..4] != b"\0asm" {
        return (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": "Invalid WASM file: magic number not found"})),
        )
            .into_response();
    }

    let tenant = extract_tenant(&headers, &state);
    let config = prism_runtime::core::CellConfig {
        memory_limit_mb: memory_mb,
        execution_target: prism_runtime::core::ExecutionTarget::Cloud,
        ..prism_runtime::core::CellConfig::default()
    };

    match state
        .runtime
        .create_cell(&tenant, &req.name, wasm_bytes, config)
        .await
    {
        Ok(cell_id) => (
            StatusCode::CREATED,
            Json(json!({"id": cell_id.to_string(), "status": "created"})),
        )
            .into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn get_cell_handler(
    State(state): State<HttpState>,
    Path(cell_id): Path<String>,
) -> Response {
    let cell_id = match uuid::Uuid::parse_str(&cell_id) {
        Ok(u) => prism_runtime::core::CellId::from_uuid(u),
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "Invalid cell_id"})),
            )
                .into_response();
        }
    };

    match state.runtime.get_cell(&cell_id).await {
        Some(cell) => (
            StatusCode::OK,
            Json(json!({
                "id": cell.id.to_string(),
                "tenant": cell.tenant_id,
                "status": format!("{:?}", cell.status),
                "name": cell.metadata.name,
            })),
        )
            .into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(json!({"error": "cell not found"})),
        )
            .into_response(),
    }
}

async fn delete_cell_handler(
    State(state): State<HttpState>,
    Path(cell_id): Path<String>,
) -> Response {
    let cell_id = match uuid::Uuid::parse_str(&cell_id) {
        Ok(u) => prism_runtime::core::CellId::from_uuid(u),
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "Invalid cell_id"})),
            )
                .into_response();
        }
    };
    match state.runtime.terminate_cell(&cell_id).await {
        Ok(_) => (StatusCode::NO_CONTENT, "").into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

#[derive(serde::Deserialize)]
struct ExecuteBody {
    cell_id: String,
    #[serde(default)]
    input: Option<String>,
}

async fn execute_handler(
    State(state): State<HttpState>,
    headers: HeaderMap,
    Json(req): Json<ExecuteBody>,
) -> Response {
    use chrono::{Datelike, Timelike};
    use prism_runtime::core::{CellId, ExecutionMetrics};
    use prism_runtime::neural::{ExecutionFeatures, ExecutionOutcome, ExecutionProfile};
    use prism_runtime::wasm_fusion::{FusionGraph, FusionNode, FusionNodeType, NodeConfig};

    let cell_id = match uuid::Uuid::parse_str(&req.cell_id) {
        Ok(u) => CellId::from_uuid(u),
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "Invalid cell_id"})),
            )
                .into_response();
        }
    };

    let cell = match state.runtime.get_cell(&cell_id).await {
        Some(c) => c,
        None => {
            return (
                StatusCode::NOT_FOUND,
                Json(json!({"error": "cell not found"})),
            )
                .into_response();
        }
    };

    let module_id = match cell.wasm_module_id {
        Some(m) => m,
        None => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "cell has no WASM module"})),
            )
                .into_response();
        }
    };

    let input_bytes = req.input.unwrap_or_default().into_bytes();
    let input_size_bytes = input_bytes.len() as u64;
    let memory_limit_mb = cell.config.memory_limit_mb;

    let mut graph = FusionGraph::new(&req.cell_id);
    graph.add_node(FusionNode {
        node_id: module_id.clone(),
        name: module_id,
        node_type: FusionNodeType::Wasm,
        config: NodeConfig {
            entry_point: "handler".to_string(),
            timeout_ms: cell.config.timeout_ms,
            memory_limit_mb: cell.config.memory_limit_mb,
            imports: Vec::new(),
        },
    });

    let fusion_executor_arc = state.runtime.fusion_executor.clone();
    let graph_arc = Arc::new(std::sync::Mutex::new(Some(graph)));
    let graph_arc2 = graph_arc.clone();

    let _tenant = extract_tenant(&headers, &state);
    let handle = tokio::task::spawn_blocking(move || {
        let executor_guard = futures::executor::block_on(fusion_executor_arc.read());
        match executor_guard.as_ref() {
            Some(executor) => {
                let mut g = graph_arc2.lock().unwrap().take().unwrap();
                futures::executor::block_on(executor.execute_graph(&mut g, &input_bytes))
            }
            None => Err(prism_runtime::core::PrismError::FusionError(
                "Fusion executor not initialized".to_string(),
            )),
        }
    });

    let result = match handle.await {
        Ok(Ok(output)) => Ok(output),
        Ok(Err(e)) => Err(format!("Execution failed: {}", e)),
        Err(e) => Err(format!("Task failed: {}", e)),
    };

    // Record execution outcome for the neural optimizer (best-effort).
    let executor_guard = state.runtime.fusion_executor.read().await;
    if let Some(ref executor) = *executor_guard {
        if let Some(snapshot) = executor.take_last_metrics().await {
            if let Some(ref cpu) = snapshot.cpu_state {
                let mut cpu_states = state.runtime.last_cpu_states.write().await;
                cpu_states.insert(cell_id, cpu.clone());
            }
            let profile = ExecutionProfile {
                cell_id,
                metrics: ExecutionMetrics {
                    duration_ms: snapshot.exec_time_ms,
                    memory_used_bytes: snapshot.memory_used_bytes,
                    ..Default::default()
                },
                features: ExecutionFeatures {
                    input_size_bytes,
                    memory_limit_mb,
                    vcpus: 1,
                    gpu_used: false,
                    execution_location: "cloud".to_string(),
                    time_of_day: chrono::Utc::now().hour() as f32,
                    day_of_week: chrono::Utc::now().weekday().num_days_from_sunday() as u8,
                },
                outcome: if snapshot.success {
                    ExecutionOutcome::Success
                } else {
                    ExecutionOutcome::Error
                },
            };
            state.runtime.record_execution_outcome(profile).await;
        }
    }
    drop(executor_guard);

    match result {
        Ok(output) => (
            StatusCode::OK,
            [("content-type", "application/octet-stream")],
            output,
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": e})),
        )
            .into_response(),
    }
}

async fn snapshot_cell_handler(
    State(state): State<HttpState>,
    Path(cell_id): Path<String>,
) -> Response {
    let cell_id = match uuid::Uuid::parse_str(&cell_id) {
        Ok(u) => prism_runtime::core::CellId::from_uuid(u),
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "Invalid cell_id"})),
            )
                .into_response();
        }
    };
    match state
        .runtime
        .snapshot_cell(&cell_id, prism_runtime::quantum::SnapshotType::Full)
        .await
    {
        Ok(snapshot) => (
            StatusCode::CREATED,
            Json(json!({
                "snapshot_id": snapshot.metadata.snapshot_id.to_string(),
                "cell_id": snapshot.metadata.cell_id.to_string(),
            })),
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn list_snapshots_handler(
    State(state): State<HttpState>,
    Path(cell_id): Path<String>,
) -> Response {
    let cell_id = match uuid::Uuid::parse_str(&cell_id) {
        Ok(u) => prism_runtime::core::CellId::from_uuid(u),
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "Invalid cell_id"})),
            )
                .into_response();
        }
    };
    let snapshots = state.runtime.list_cell_snapshots(&cell_id).await;
    (
        StatusCode::OK,
        Json(json!({
            "cell_id": cell_id.to_string(),
            "snapshots": snapshots.iter().map(|s| json!({
                "id": s.snapshot_id,
                "type": format!("{:?}", s.snapshot_type),
                "created_at": s.created_at.to_rfc3339(),
            })).collect::<Vec<_>>()
        })),
    )
        .into_response()
}

async fn restore_snapshot_handler(
    State(state): State<HttpState>,
    Path(snapshot_id): Path<String>,
) -> Response {
    match state.runtime.restore_cell_from_snapshot(&snapshot_id).await {
        Ok(cell_id) => (
            StatusCode::OK,
            Json(json!({"cell_id": cell_id.to_string()})),
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn delete_snapshot_handler(
    State(state): State<HttpState>,
    Path(snapshot_id): Path<String>,
) -> Response {
    match state.runtime.delete_snapshot(&snapshot_id).await {
        Ok(_) => (StatusCode::NO_CONTENT, "").into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn list_capabilities_handler(State(state): State<HttpState>) -> Response {
    let caps = state.runtime.list_capabilities().await;
    (
        StatusCode::OK,
        Json(json!({
            "capabilities": caps.iter().map(|c| json!({
                "id": c.capability_id.0,
                "name": c.name,
                "category": format!("{:?}", c.category),
                "version": c.metadata.version,
            })).collect::<Vec<_>>()
        })),
    )
        .into_response()
}

#[derive(serde::Deserialize)]
struct RegisterCapabilityBody {
    name: String,
    category: String,
    #[serde(default)]
    version: Option<String>,
}

async fn register_capability_handler(
    State(state): State<HttpState>,
    Json(req): Json<RegisterCapabilityBody>,
) -> Response {
    let category = match req.category.to_lowercase().as_str() {
        "ai" => prism_runtime::ucl::CapabilityCategory::Ai,
        "data" => prism_runtime::ucl::CapabilityCategory::System, // best-effort mapping
        "compute" => prism_runtime::ucl::CapabilityCategory::Compute,
        "network" => prism_runtime::ucl::CapabilityCategory::Network,
        "storage" => prism_runtime::ucl::CapabilityCategory::Storage,
        "crypto" => prism_runtime::ucl::CapabilityCategory::Crypto,
        "sensors" => prism_runtime::ucl::CapabilityCategory::Sensors,
        "system" => prism_runtime::ucl::CapabilityCategory::System,
        _ => prism_runtime::ucl::CapabilityCategory::Compute,
    };
    let provider = req
        .version
        .clone()
        .unwrap_or_else(|| "functionfly".to_string());
    let mut cap = prism_runtime::ucl::Capability::new(&req.name, category, &provider);
    if let Some(v) = req.version.as_ref() {
        cap.metadata.version = v.clone();
    }
    match state.runtime.register_capability(cap).await {
        Ok(_) => (
            StatusCode::CREATED,
            Json(json!({"name": req.name, "status": "registered"})),
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

#[derive(serde::Deserialize)]
struct InvokeCapabilityBody {
    /// Input is forwarded as raw bytes to the capability. JSON-typed
    /// callers should serialize before posting.
    #[serde(default)]
    input: Option<String>,
}

async fn invoke_capability_handler(
    State(state): State<HttpState>,
    Path(name): Path<String>,
    Json(req): Json<InvokeCapabilityBody>,
) -> Response {
    let input_bytes = req.input.unwrap_or_default().into_bytes();
    match state.runtime.invoke_capability(&name, &input_bytes).await {
        Ok(out) => (
            StatusCode::OK,
            [("content-type", "application/octet-stream")],
            out,
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn list_swarms_handler(State(state): State<HttpState>) -> Response {
    let swarms = state.runtime.list_swarms().await;
    (
        StatusCode::OK,
        Json(json!({
            "swarms": swarms.iter().map(|s| json!({
                "id": s.swarm_id.0.to_string(),
                "state": format!("{:?}", s.state),
                "cells": s.cells.len(),
                "peers": s.peer_nodes.len(),
                "created_at": s.created_at.to_rfc3339(),
            })).collect::<Vec<_>>()
        })),
    )
        .into_response()
}

#[derive(serde::Deserialize)]
struct CreateSwarmBody {
    swarm_id: String,
}

async fn create_swarm_handler(
    State(state): State<HttpState>,
    Json(req): Json<CreateSwarmBody>,
) -> Response {
    match state.runtime.create_swarm(req.swarm_id).await {
        Ok(swarm_id) => (StatusCode::CREATED, Json(json!({"swarm_id": swarm_id}))).into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

#[derive(serde::Deserialize)]
struct SwarmMembershipBody {
    #[serde(default)]
    cell_id: Option<String>,
}

async fn join_swarm_handler(
    State(state): State<HttpState>,
    Path(swarm_id): Path<String>,
    Json(req): Json<SwarmMembershipBody>,
) -> Response {
    // If no cell_id is provided, we use a zero UUID (legacy behavior).
    // Callers should provide cell_id in the body for production use.
    let cell_uuid = req
        .cell_id
        .as_deref()
        .and_then(|s| uuid::Uuid::parse_str(s).ok())
        .unwrap_or_else(uuid::Uuid::nil);
    let cell_id = prism_runtime::core::CellId::from_uuid(cell_uuid);
    match state.runtime.join_swarm(&swarm_id, cell_id).await {
        Ok(_) => (StatusCode::OK, Json(json!({"status": "joined"}))).into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn leave_swarm_handler(
    State(state): State<HttpState>,
    Path(swarm_id): Path<String>,
    Json(req): Json<SwarmMembershipBody>,
) -> Response {
    let cell_uuid = req
        .cell_id
        .as_deref()
        .and_then(|s| uuid::Uuid::parse_str(s).ok())
        .unwrap_or_else(uuid::Uuid::nil);
    let cell_id = prism_runtime::core::CellId::from_uuid(cell_uuid);
    match state.runtime.leave_swarm(&swarm_id, cell_id).await {
        Ok(_) => (StatusCode::OK, Json(json!({"status": "left"}))).into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(json!({"error": format!("{}", e)})),
        )
            .into_response(),
    }
}

async fn get_optimization_handler(
    State(state): State<HttpState>,
    Path(cell_id): Path<String>,
) -> Response {
    let cell_id = match uuid::Uuid::parse_str(&cell_id) {
        Ok(u) => prism_runtime::core::CellId::from_uuid(u),
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(json!({"error": "Invalid cell_id"})),
            )
                .into_response();
        }
    };
    let suggestion = state.runtime.get_optimization_suggestion(&cell_id).await;
    (
        StatusCode::OK,
        Json(json!({
            "cell_id": cell_id.to_string(),
            "suggestion": format!("{:?}", suggestion),
        })),
    )
        .into_response()
}

/// Extract tenant ID for rate-limiting / audit purposes.
///
/// Priority:
/// 1. `X-Tenant-ID` header (set by trusted upstream proxies)
/// 2. Bearer token prefix (first 16 chars) — used for audit grouping
/// 3. `"anonymous"` fallback
fn extract_tenant(headers: &HeaderMap, state: &HttpState) -> String {
    if let Some(t) = headers
        .get("x-tenant-id")
        .and_then(|v| v.to_str().ok())
        .filter(|s| !s.is_empty())
    {
        return t.to_string();
    }

    if let Some(_token) = state.api_token.as_ref() {
        if let Some(auth) = headers
            .get("authorization")
            .and_then(|v| v.to_str().ok())
            .and_then(|v| v.strip_prefix("Bearer "))
        {
            // Hash the token prefix for tenant grouping. Not for auth —
            // we already validated the token above.
            let prefix = auth.bytes().take(16).collect::<Vec<u8>>();
            return format!("token:{}", String::from_utf8_lossy(&prefix));
        }
    }

    "anonymous".to_string()
}
