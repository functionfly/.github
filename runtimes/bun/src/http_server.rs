//! HTTP server for Bun runtime
//!
//! Production-ready HTTP API with health checks, metrics,
//! and secure execution endpoints.

use crate::config::RuntimeConfig;
use crate::execution::{ExecutionRequest, ExecutionResponse};
use crate::metrics::RuntimeMetrics;
use crate::Executor;
use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::net::TcpListener;
use tracing::{info, warn};

/// Application state
#[derive(Clone)]
pub struct AppState {
    pub executor: Arc<Executor>,
    pub config: RuntimeConfig,
    pub started_at: Instant,
}

/// Health check response
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub runtime: String,
    pub version: String,
    pub uptime_secs: u64,
}

/// Metrics response
#[derive(Debug, Serialize)]
pub struct MetricsResponse {
    pub metrics: RuntimeMetrics,
    pub uptime_secs: u64,
}

/// Error response
#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: String,
}

/// Execution request from HTTP
#[derive(Debug, Deserialize)]
pub struct ExecuteRequest {
    pub code: String,
    pub input: Option<serde_json::Value>,
    pub timeout_ms: Option<u64>,
}

/// Execution response for HTTP
#[derive(Debug, Serialize)]
pub struct ExecuteResponse {
    pub id: String,
    pub success: bool,
    pub output: Option<serde_json::Value>,
    pub error: Option<String>,
    pub execution_time_ms: u64,
}

/// Create the Axum router
pub fn create_app(state: AppState) -> Router {
    Router::new()
        .route("/health", get(health_handler))
        .route("/metrics", get(metrics_handler))
        .route("/execute", post(execute_handler))
        .route("/execute/sync", post(execute_sync_handler))
        .with_state(state)
}

/// Health check handler
async fn health_handler(State(state): State<AppState>) -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        runtime: "bun".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        uptime_secs: state.started_at.elapsed().as_secs(),
    })
}

/// Metrics handler
async fn metrics_handler(State(state): State<AppState>) -> Json<MetricsResponse> {
    let metrics = state.executor.metrics().await;
    Json(MetricsResponse {
        metrics,
        uptime_secs: state.started_at.elapsed().as_secs(),
    })
}

/// Async execution handler
async fn execute_handler(
    State(state): State<AppState>,
    Json(req): Json<ExecuteRequest>,
) -> Result<Json<ExecuteResponse>, (StatusCode, Json<ErrorResponse>)> {
    let request = ExecutionRequest {
        id: uuid::Uuid::new_v4(),
        code: req.code,
        entry: None,
        input: req.input,
        timeout: req.timeout_ms.map(Duration::from_millis),
        limits: None,
    };

    let response = state.executor.execute(request).await.map_err(|e| {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: e.to_string(),
                code: "EXECUTION_ERROR".to_string(),
            }),
        )
    })?;

    Ok(Json(ExecuteResponse {
        id: response.id.to_string(),
        success: response.success,
        output: response.output,
        error: response.error,
        execution_time_ms: response.execution_time_ms,
    }))
}

/// Synchronous execution handler with shorter timeout
async fn execute_sync_handler(
    State(state): State<AppState>,
    Json(req): Json<ExecuteRequest>,
) -> Result<Json<ExecuteResponse>, (StatusCode, Json<ErrorResponse>)> {
    let timeout = req.timeout_ms.unwrap_or(5000).min(30000); // Max 30s

    let request = ExecutionRequest {
        id: uuid::Uuid::new_v4(),
        code: req.code,
        entry: None,
        input: req.input,
        timeout: Some(Duration::from_millis(timeout)),
        limits: None,
    };

    let result = tokio::time::timeout(
        Duration::from_millis(timeout + 1000),
        state.executor.execute(request),
    )
    .await;

    let response = match result {
        Ok(Ok(r)) => r,
        Ok(Err(e)) => {
            return Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: e.to_string(),
                    code: "EXECUTION_ERROR".to_string(),
                }),
            ))
        }
        Err(_) => {
            return Err((
                StatusCode::REQUEST_TIMEOUT,
                Json(ErrorResponse {
                    error: "Execution timed out".to_string(),
                    code: "TIMEOUT".to_string(),
                }),
            ))
        }
    };

    Ok(Json(ExecuteResponse {
        id: response.id.to_string(),
        success: response.success,
        output: response.output,
        error: response.error,
        execution_time_ms: response.execution_time_ms,
    }))
}

/// Run the HTTP server
pub async fn run_server(config: RuntimeConfig, port: u16) -> anyhow::Result<()> {
    let executor = Arc::new(Executor::new(config.clone()));
    let state = AppState {
        executor,
        config,
        started_at: Instant::now(),
    };

    let app = create_app(state);

    let addr: SocketAddr = format!("127.0.0.1:{}", port).parse()?;
    let listener = TcpListener::bind(addr).await?;

    info!(port = port, "Bun runtime HTTP server started");

    axum::serve(listener, app).await?;

    Ok(())
}