//! HTTP server for Kotlin/JVM runtime
//!
//! Provides REST API endpoints for code execution, health checks, and metrics.

use crate::config::RuntimeConfig;
use crate::execution::{ExecutionRequest, ExecutionResponse};
use crate::metrics::MetricsCollector;
use crate::Executor;
use anyhow::Result;
use axum::{
    body::Body,
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::trace::TraceLayer;
use tracing;

/// Application state shared across handlers
#[derive(Clone)]
pub struct AppState {
    pub executor: Arc<Executor>,
    pub metrics: Arc<MetricsCollector>,
    pub config: RuntimeConfig,
    pub api_token: Option<String>,
}

impl AppState {
    pub fn new(executor: Executor, metrics: MetricsCollector, config: RuntimeConfig) -> Self {
        Self {
            executor: Arc::new(executor),
            metrics: Arc::new(metrics),
            config,
            api_token: None,
        }
    }

    pub fn with_auth(mut self, token: Option<String>) -> Self {
        self.api_token = token;
        self
    }
}

/// Health check response
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
    pub uptime_secs: f64,
}

/// Readiness check response
#[derive(Debug, Serialize)]
pub struct ReadinessResponse {
    pub ready: bool,
    pub sandbox_available: bool,
    pub currently_executing: u64,
}

/// Metrics response (JSON)
#[derive(Debug, Serialize)]
pub struct MetricsResponse {
    pub metrics: crate::metrics::RuntimeMetrics,
}

/// Validation request
#[derive(Debug, Deserialize)]
pub struct ValidateRequest {
    pub code: String,
}

/// Validation response
#[derive(Debug, Serialize)]
pub struct ValidateResponse {
    pub valid: bool,
    pub errors: Vec<String>,
    pub warnings: Vec<String>,
}

/// Error response
#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: u16,
}

impl ErrorResponse {
    pub fn new(error: impl Into<String>, code: u16) -> Self {
        Self {
            error: error.into(),
            code,
        }
    }
}

impl IntoResponse for ErrorResponse {
    fn into_response(self) -> Response {
        (StatusCode::from_u16(self.code).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR), Json(self)).into_response()
    }
}

/// Health check endpoint
async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        uptime_secs: 0.0, // Will be filled by caller
    })
}

/// Readiness check endpoint
async fn ready_handler(State(state): State<AppState>) -> Json<ReadinessResponse> {
    let metrics = state.metrics.get_metrics().await;
    Json(ReadinessResponse {
        ready: true,
        sandbox_available: metrics.active_executions < state.config.max_concurrent as u64,
        currently_executing: metrics.active_executions,
    })
}

/// Metrics endpoint (JSON)
async fn metrics_handler(State(state): State<AppState>) -> Json<MetricsResponse> {
    Json(MetricsResponse {
        metrics: state.metrics.get_metrics().await,
    })
}

/// Metrics endpoint (Prometheus text format)
async fn metrics_prometheus_handler(State(state): State<AppState>) -> impl IntoResponse {
    // Note: encode_prometheus is sync, but it uses default metrics
    // In production, this should be async with proper state access
    let metrics_text = state.metrics.encode_prometheus();
    Response::builder()
        .status(200)
        .header("Content-Type", "text/plain; version=0.0.4")
        .body(Body::from(metrics_text))
        .unwrap()
}

/// Execute code endpoint
async fn execute_handler(
    State(state): State<AppState>,
    headers: axum::http::HeaderMap,
    Json(request): Json<ExecutionRequest>,
) -> Result<Json<ExecutionResponse>, ErrorResponse> {
    // Auth check
    if let Some(ref token) = state.api_token {
        let auth = headers.get("authorization")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");
        if auth != format!("Bearer {}", token) {
            return Err(ErrorResponse::new("unauthorized", 401));
        }
    }

    // Record start
    state.metrics.start_execution().await;

    // Execute code
    let response = state.executor.execute(request).await;

    // Record end
    state.metrics.end_execution().await;

    // Record metrics based on result
    if response.success {
        state.metrics.record_execution(response.execution_time_ms, response.memory_used_mb.unwrap_or(0) as u64).await;
    } else if response.termination_reason.as_deref() == Some("timeout") {
        state.metrics.record_timeout().await;
    } else {
        state.metrics.record_failure(response.execution_time_ms).await;
    }

    Ok(Json(response))
}

/// Validate code endpoint
async fn validate_handler(
    State(state): State<AppState>,
    Json(request): Json<ValidateRequest>,
) -> Json<ValidateResponse> {
    let result = state.executor.validate(&request.code);

    Json(ValidateResponse {
        valid: result.valid,
        errors: result.errors,
        warnings: result.warnings,
    })
}

/// Root handler
async fn root_handler() -> &'static str {
    "FunctionFly Kotlin Runtime\n=======================\n\nVersion: 0.1.0\n\nEndpoints:\n  GET  /health      - Health check\n  GET  /ready       - Readiness check\n  GET  /metrics     - Metrics (JSON)\n  GET  /metrics/prom - Metrics (Prometheus)\n  POST /execute     - Execute Kotlin code\n  POST /validate    - Validate Kotlin code\n"
}

/// Create the Axum router
pub fn create_app(state: AppState) -> Router {
    Router::new()
        .route("/", get(root_handler))
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        .route("/metrics", get(metrics_handler))
        .route("/metrics/prom", get(metrics_prometheus_handler))
        .route("/execute", post(execute_handler))
        .route("/validate", post(validate_handler))
        .with_state(state)
        .layer(TraceLayer::new_for_http())
}

/// Run the HTTP server
pub async fn run_server(
    addr: SocketAddr,
    executor: Executor,
    metrics: MetricsCollector,
    config: RuntimeConfig,
) -> Result<()> {
    let state = AppState::new(executor, metrics, config);

    let app = create_app(state);

    tracing::info!("Starting Kotlin runtime HTTP server on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}

/// Create a server with graceful shutdown
pub async fn run_server_with_shutdown(
    addr: SocketAddr,
    executor: Executor,
    metrics: MetricsCollector,
    config: RuntimeConfig,
    shutdown_signal: impl std::future::Future<Output = ()> + Send + 'static,
) -> Result<()> {
    let state = AppState::new(executor, metrics, config);
    let app = create_app(state);

    tracing::info!("Starting Kotlin runtime HTTP server on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;

    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal)
        .await?;

    tracing::info!("Kotlin runtime HTTP server stopped");
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_health_response_serialization() {
        let response = HealthResponse {
            status: "healthy".to_string(),
            version: "0.1.0".to_string(),
            uptime_secs: 100.0,
        };

        let json = serde_json::to_string(&response).unwrap();
        assert!(json.contains("healthy"));
        assert!(json.contains("0.1.0"));
    }

    #[test]
    fn test_error_response() {
        let error = ErrorResponse::new("something went wrong", 500);
        assert_eq!(error.code, 500);
        assert_eq!(error.error, "something went wrong");
    }

    #[tokio::test]
    async fn test_ready_handler() {
        let metrics = Arc::new(MetricsCollector::new("test".to_string()));
        let executor = Executor::with_defaults(metrics.clone()).unwrap();
        let state = AppState::new(executor, (*metrics).clone(), RuntimeConfig::default());

        let response = ready_handler(State(state)).await;
        assert!(response.ready);
    }
}