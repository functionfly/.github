//! Prometheus metrics exporter.

use axum::{extract::State, response::IntoResponse};
use std::sync::Arc;

use super::types::AppState;

/// Prometheus metrics handler.
///
/// Exposes runtime metrics in Prometheus text format at `/metrics`.
/// This enables unified observability with the Go backend which already
/// exports Prometheus metrics via `prometheus/client_golang`.
///
/// Metrics exported:
/// - `functionfly_executions_total` — total function executions
/// - `functionfly_execution_time_ms_avg` — average execution time
/// - `functionfly_cache_hit_rate` — cache hit rate (0.0–100.0)
/// - `functionfly_error_rate` — error rate (0.0–100.0)
/// - `functionfly_concurrent_executions` — current concurrent executions
pub async fn prometheus_metrics(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let stats = state.monitor.get_stats().await;
    let concurrent = state.monitor.get_total_concurrent().await;

    let mut output = String::new();

    output.push_str("# HELP functionfly_executions_total Total number of function executions\n");
    output.push_str("# TYPE functionfly_executions_total counter\n");
    output.push_str(&format!(
        "functionfly_executions_total {}\n\n",
        stats.total_executions
    ));

    output.push_str("# HELP functionfly_execution_time_ms_avg Average execution time in milliseconds\n");
    output.push_str("# TYPE functionfly_execution_time_ms_avg gauge\n");
    output.push_str(&format!(
        "functionfly_execution_time_ms_avg {:.2}\n\n",
        stats.average_execution_time_ms
    ));

    output.push_str("# HELP functionfly_cache_hit_rate Cache hit rate percentage (0-100)\n");
    output.push_str("# TYPE functionfly_cache_hit_rate gauge\n");
    output.push_str(&format!(
        "functionfly_cache_hit_rate {:.2}\n\n",
        stats.cache_hit_rate
    ));

    output.push_str("# HELP functionfly_error_rate Error rate percentage (0-100)\n");
    output.push_str("# TYPE functionfly_error_rate gauge\n");
    output.push_str(&format!(
        "functionfly_error_rate {:.2}\n\n",
        stats.error_rate
    ));

    output.push_str("# HELP functionfly_concurrent_executions Current number of concurrent executions\n");
    output.push_str("# TYPE functionfly_concurrent_executions gauge\n");
    output.push_str(&format!(
        "functionfly_concurrent_executions {}\n\n",
        concurrent
    ));

    output.push_str("# HELP functionfly_peak_memory_mb Peak memory usage in MB\n");
    output.push_str("# TYPE functionfly_peak_memory_mb gauge\n");
    output.push_str(&format!(
        "functionfly_peak_memory_mb {:.2}\n\n",
        stats.peak_memory_usage_mb
    ));

    output.push_str("# HELP functionfly_uptime_seconds Runtime uptime in seconds\n");
    output.push_str("# TYPE functionfly_uptime_seconds counter\n");
    output.push_str(&format!(
        "functionfly_uptime_seconds {}",
        stats.uptime_seconds
    ));

    axum::response::Response::builder()
        .status(200)
        .header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
        .body(axum::body::Body::from(output))
        .unwrap_or_else(|_| {
            axum::Json(serde_json::json!({ "error": "metrics unavailable" })).into_response()
        })
}
