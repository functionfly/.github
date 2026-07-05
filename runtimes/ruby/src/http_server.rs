//! HTTP server for Ruby runtime
//!
//! Production-ready HTTP API with health checks, metrics,
//! circuit breaker, rate limiting, and secure execution endpoints.

use crate::config::RuntimeConfig;
use crate::execution::{ExecutionRequest, ExecutionResponse};
use crate::metrics::MetricsCollector;
use crate::orchestrator_client::OrchestratorClient;
use crate::security::{SecurityAuditor, SecurityEvent, SecurityEventType};
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::{Duration, Instant};
use std::sync::atomic::{AtomicU64, Ordering};
use tokio::net::TcpListener;
use tracing::{info, warn, debug};

/// Application state
#[derive(Clone)]
pub struct AppState {
    pub executor: Arc<dyn crate::execution::Executor>,
    pub metrics: Arc<MetricsCollector>,
    pub orchestrator: Arc<RwLock<OrchestratorClient>>,
    pub config: RuntimeConfig,
    pub security_auditor: Arc<SecurityAuditor>,
    pub started_at: Instant,
    pub circuit_breaker: Arc<CircuitBreaker>,
    pub api_token: Option<String>,
}

/// Circuit breaker state
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CircuitState {
    Closed,   // Normal operation
    Open,     // Failing, rejecting requests
    HalfOpen, // Testing if recovery is possible
}

/// Circuit breaker for handling cascading failures
pub struct CircuitBreaker {
    state: AtomicU64,
    failure_count: AtomicU64,
    last_failure_time: RwLock<Instant>,
    threshold: u64,
    recovery_timeout: Duration,
}

impl CircuitBreaker {
    pub fn new(threshold: u64, recovery_timeout_secs: u64) -> Arc<Self> {
        Arc::new(Self {
            state: AtomicU64::new(0),
            failure_count: AtomicU64::new(0),
            last_failure_time: RwLock::new(Instant::now()),
            threshold,
            recovery_timeout: Duration::from_secs(recovery_timeout_secs),
        })
    }

    pub fn state(&self) -> CircuitState {
        match self.state.load(Ordering::SeqCst) {
            0 => CircuitState::Closed,
            1 => CircuitState::Open,
            2 => CircuitState::HalfOpen,
            _ => CircuitState::Closed,
        }
    }

    pub fn record_success(&self) {
        self.failure_count.store(0, Ordering::SeqCst);
        self.state.store(0, Ordering::SeqCst);
    }

    pub fn record_failure(&self) {
        let failures = self.failure_count.fetch_add(1, Ordering::SeqCst) + 1;
        *self.last_failure_time.write() = Instant::now();

        if failures >= self.threshold {
            self.state.store(1, Ordering::SeqCst); // Open
        }
    }

    pub fn is_allowed(&self) -> bool {
        match self.state() {
            CircuitState::Closed => true,
            CircuitState::Open => {
                let elapsed = self.last_failure_time.read().elapsed();
                if elapsed >= self.recovery_timeout {
                    self.state.store(2, Ordering::SeqCst); // HalfOpen
                    true
                } else {
                    false
                }
            }
            CircuitState::HalfOpen => true,
        }
    }
}

/// Health check response
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub runtime: String,
    pub version: String,
    pub uptime_secs: u64,
    pub circuit_breaker: String,
}

/// Metrics response
#[derive(Debug, Serialize)]
pub struct MetricsResponse {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub failed_executions: u64,
    pub active_executions: usize,
    pub uptime_secs: u64,
}

/// Error response
#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: String,
    pub retry_after_ms: Option<u64>,
}

/// Execution request from HTTP
#[derive(Debug, Deserialize)]
pub struct ExecuteRequest {
    pub code: String,
    pub input: Option<serde_json::Value>,
    pub timeout_ms: Option<u64>,
    pub tenant_id: Option<String>,
}

/// Execution response for HTTP
#[derive(Debug, Serialize)]
pub struct ExecuteResponse {
    pub execution_id: String,
    pub success: bool,
    pub output: Option<String>,
    pub error: Option<String>,
    pub execution_time_ms: u64,
    pub memory_used_mb: Option<f64>,
}

/// Maximum request body size (1MB)
const MAX_REQUEST_BODY_SIZE: usize = 1024 * 1024;

/// Maximum code size (256KB)
const MAX_CODE_SIZE: usize = 256 * 1024;

/// Request validator
pub struct RequestValidator;

impl RequestValidator {
    pub fn validate(req: &ExecuteRequest) -> Result<(), ValidationError> {
        if req.code.len() > MAX_CODE_SIZE {
            return Err(ValidationError::CodeTooLarge {
                size: req.code.len(),
                max: MAX_CODE_SIZE,
            });
        }

        if req.code.trim().is_empty() {
            return Err(ValidationError::EmptyCode);
        }

        if req.code.contains('\0') {
            return Err(ValidationError::InvalidCharacters);
        }

        if let Some(timeout_ms) = req.timeout_ms {
            if timeout_ms == 0 {
                return Err(ValidationError::InvalidTimeout { value: timeout_ms });
            }
            if timeout_ms > 300_000 {
                return Err(ValidationError::TimeoutTooLarge { value: timeout_ms });
            }
        }

        Ok(())
    }
}

#[derive(Debug, thiserror::Error)]
pub enum ValidationError {
    #[error("code too large: {size} bytes (max {max})")]
    CodeTooLarge { size: usize, max: usize },

    #[error("code is empty")]
    EmptyCode,

    #[error("code contains invalid characters")]
    InvalidCharacters,

    #[error("invalid timeout: {value}ms")]
    InvalidTimeout { value: u64 },

    #[error("timeout too large: {value}ms (max 300000ms)")]
    TimeoutTooLarge { value: u64 },
}

/// Security headers middleware. See bun/src/main.rs for full rationale.
pub async fn security_headers_middleware(
    req: axum::http::Request<axum::body::Body>,
    next: axum::middleware::Next,
) -> axum::response::Response {
    let mut response = next.run(req).await;
    let headers = response.headers_mut();
    headers.insert(
        axum::http::header::HeaderName::from_static("x-content-type-options"),
        axum::http::HeaderValue::from_static("nosniff"),
    );
    headers.insert(
        axum::http::header::HeaderName::from_static("x-frame-options"),
        axum::http::HeaderValue::from_static("DENY"),
    );
    headers.insert(
        axum::http::header::HeaderName::from_static("referrer-policy"),
        axum::http::HeaderValue::from_static("strict-origin-when-cross-origin"),
    );
    headers.insert(
        axum::http::header::HeaderName::from_static("content-security-policy"),
        axum::http::HeaderValue::from_static(
            "default-src 'none'; frame-ancestors 'none'; base-uri 'none'",
        ),
    );
    headers.insert(
        axum::http::header::HeaderName::from_static("strict-transport-security"),
        axum::http::HeaderValue::from_static("max-age=31536000; includeSubDomains"),
    );
    response
}

/// Create the Axum router
pub fn create_app(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        .route("/metrics", get(metrics_handler))
        .route("/execute", post(execute_handler))
        .route("/execute/{function_id}/{version}", post(execute_versioned_handler))
        .route("/shutdown", post(shutdown_handler))
        .with_state(state)
        .layer(axum::middleware::from_fn(security_headers_middleware))
}

/// Health check handler
async fn health_handler(State(state): State<AppState>) -> Json<HealthResponse> {
    let cb_state = format!("{:?}", state.circuit_breaker.state());
    Json(HealthResponse {
        status: "healthy".to_string(),
        runtime: "ruby".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        uptime_secs: state.started_at.elapsed().as_secs(),
        circuit_breaker: cb_state,
    })
}

/// Readiness check handler
async fn ready_handler(State(state): State<AppState>) -> Json<serde_json::Value> {
    let orchestrator = state.orchestrator.read();
    let cb_state = format!("{:?}", state.circuit_breaker.state());
    Json(serde_json::json!({
        "ready": true,
        "registered": orchestrator.is_registered(),
        "nats_connected": orchestrator.is_connected(),
        "circuit_breaker": cb_state,
    }))
}

/// Metrics handler
async fn metrics_handler(State(state): State<AppState>) -> impl IntoResponse {
    let summary = state.metrics.summary();
    let cb_state = format!("{:?}", state.circuit_breaker.state());

    let body = format!(
        "# HELP ruby_runtime_executions_total Total number of executions\n\
         # TYPE ruby_runtime_executions_total counter\n\
         ruby_runtime_executions_total {}\n\
         # HELP ruby_runtime_successful_executions Total successful executions\n\
         # TYPE ruby_runtime_successful_executions counter\n\
         ruby_runtime_successful_executions {}\n\
         # HELP ruby_runtime_failed_executions Total failed executions\n\
         # TYPE ruby_runtime_failed_executions counter\n\
         ruby_runtime_failed_executions {}\n\
         # HELP ruby_runtime_active_executions Currently active executions\n\
         # TYPE ruby_runtime_active_executions gauge\n\
         ruby_runtime_active_executions {}\n\
         # HELP ruby_runtime_circuit_breaker_state Circuit breaker state\n\
         # TYPE ruby_runtime_circuit_breaker_state gauge\n\
         ruby_runtime_circuit_breaker_state {}\n",
        summary.total_executions,
        summary.successful_executions,
        summary.failed_executions,
        summary.active_executions,
        cb_state
    );

    ([(axum::http::header::CONTENT_TYPE, "text/plain; charset=utf-8")], body)
}

/// Async execution handler
async fn execute_handler(
    State(state): State<AppState>,
    headers: axum::http::HeaderMap,
    Json(req): Json<ExecuteRequest>,
) -> Result<impl IntoResponse, (StatusCode, Json<ErrorResponse>)> {
    // Auth check
    if let Some(ref token) = state.api_token {
        let auth = headers.get("authorization")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");
        if !constant_time_eq::constant_time_eq(auth.as_bytes(), format!("Bearer {}", token).as_bytes()) {
            return Err((
                StatusCode::UNAUTHORIZED,
                Json(ErrorResponse {
                    error: "unauthorized".to_string(),
                    code: "UNAUTHORIZED".to_string(),
                    retry_after_ms: None,
                }),
            ));
        }
    }

    // Check circuit breaker
    if !state.circuit_breaker.is_allowed() {
        state.metrics.record_rate_limited();
        return Err((
            StatusCode::SERVICE_UNAVAILABLE,
            Json(ErrorResponse {
                error: "Service temporarily unavailable due to high error rate".to_string(),
                code: "CIRCUIT_OPEN".to_string(),
                retry_after_ms: Some(5000),
            }),
        ));
    }

    // Validate request
    if let Err(e) = RequestValidator::validate(&req) {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: e.to_string(),
                code: "VALIDATION_ERROR".to_string(),
                retry_after_ms: None,
            }),
        ));
    }

    let execution_id = uuid::Uuid::new_v4().to_string();

    let exec_request = ExecutionRequest {
        execution_id: execution_id.clone(),
        code: req.code,
        input: req.input,
        timeout_ms: req.timeout_ms,
        tenant_id: req.tenant_id.clone(),
    };

    debug!(execution_id = %execution_id, tenant_id = ?req.tenant_id, "Executing Ruby code");

    let start = Instant::now();
    let response = state.executor.execute(exec_request).await;
    let elapsed_ms = start.elapsed().as_millis() as u64;

    if response.success {
        state.circuit_breaker.record_success();
        Ok(Json(ExecuteResponse {
            execution_id,
            success: true,
            output: response.output,
            error: None,
            execution_time_ms: elapsed_ms,
            memory_used_mb: response.memory_used_mb,
        }))
    } else {
        state.circuit_breaker.record_failure();
        state.security_auditor.log(SecurityEvent {
            timestamp: chrono::Utc::now(),
            severity: crate::security::SecuritySeverity::Medium,
            event_type: SecurityEventType::ExecutionError,
            description: response.error.clone().unwrap_or_default(),
            execution_id: Some(execution_id.clone()),
            tenant_id: req.tenant_id,
            ..Default::default()
        });
        Err((
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: response.error.unwrap_or_else(|| "unknown error".to_string()),
                code: "EXECUTION_ERROR".to_string(),
                retry_after_ms: None,
            }),
        ))
    }
}

/// Versioned execution handler
async fn execute_versioned_handler(
    State(state): State<AppState>,
    headers: axum::http::HeaderMap,
    Path((_function_id, _version)): Path<(String, String)>,
    Json(req): Json<ExecuteRequest>,
) -> Result<impl IntoResponse, (StatusCode, Json<ErrorResponse>)> {
    // Auth check
    if let Some(ref token) = state.api_token {
        let auth = headers.get("authorization")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");
        if !constant_time_eq::constant_time_eq(auth.as_bytes(), format!("Bearer {}", token).as_bytes()) {
            return Err((
                StatusCode::UNAUTHORIZED,
                Json(ErrorResponse {
                    error: "unauthorized".to_string(),
                    code: "UNAUTHORIZED".to_string(),
                    retry_after_ms: None,
                }),
            ));
        }
    }

    // Validate request
    if let Err(e) = RequestValidator::validate(&req) {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: e.to_string(),
                code: "VALIDATION_ERROR".to_string(),
                retry_after_ms: None,
            }),
        ));
    }

    let execution_id = uuid::Uuid::new_v4().to_string();

    let exec_request = ExecutionRequest {
        execution_id: execution_id.clone(),
        code: req.code,
        input: req.input,
        timeout_ms: req.timeout_ms,
        tenant_id: req.tenant_id,
    };

    let start = Instant::now();
    let response = state.executor.execute(exec_request).await;
    let elapsed_ms = start.elapsed().as_millis() as u64;

    if response.success {
        state.circuit_breaker.record_success();
        Ok(Json(serde_json::json!({
            "result": response.output,
            "exec_time_ms": elapsed_ms,
            "cache_hit": false,
        })))
    } else {
        state.circuit_breaker.record_failure();
        Err((
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: response.error.unwrap_or_else(|| "unknown error".to_string()),
                code: "EXECUTION_ERROR".to_string(),
                retry_after_ms: None,
            }),
        ))
    }
}

/// Shutdown handler
async fn shutdown_handler(
    State(_state): State<AppState>,
    Json(_req): Json<ShutdownRequest>,
) -> Json<serde_json::Value> {
    // In production, this would trigger graceful shutdown
    Json(serde_json::json!({
        "ok": true,
        "message": "Shutdown endpoint received"
    }))
}

#[derive(Debug, Deserialize)]
struct ShutdownRequest {
    grace_period_seconds: Option<u64>,
}

/// Run the HTTP server
pub async fn run_server(
    config: RuntimeConfig,
    port: u16,
    executor: Arc<dyn crate::execution::Executor>,
    metrics: Arc<MetricsCollector>,
    orchestrator: Arc<RwLock<OrchestratorClient>>,
    security_auditor: Arc<SecurityAuditor>,
    api_token: Option<String>,
) -> anyhow::Result<()> {
    let state = AppState {
        executor,
        metrics,
        orchestrator,
        config,
        security_auditor,
        started_at: Instant::now(),
        circuit_breaker: CircuitBreaker::new(10, 30),
        api_token,
    };

    let app = create_app(state)
        // Limit request body to 1 MiB. Ruby code is small; without this cap,
        // a single 100 MB payload could exhaust the process before the
        // executor even sees it.
        .layer(axum::extract::DefaultBodyLimit::max(1 * 1024 * 1024));

    let addr: SocketAddr = format!("127.0.0.1:{}", port).parse()?;
    let listener = TcpListener::bind(addr).await?;

    info!(port = port, "Ruby runtime HTTP server started");

    axum::serve(listener, app).await?;

    Ok(())
}

static START_TIME: std::sync::OnceLock<Instant> = std::sync::OnceLock::new();

pub fn get_uptime_seconds() -> u64 {
    START_TIME.get().map(|t| t.elapsed().as_secs()).unwrap_or(0)
}

pub fn init_start_time() {
    let _ = START_TIME.set(Instant::now());
}
