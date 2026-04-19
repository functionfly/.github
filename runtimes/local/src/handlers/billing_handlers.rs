//! Billing and cost attribution HTTP handlers (Phase 6: Observability).
//!
//! Exposes cost tracking data and allows manual flush of cost records
//! to the Go backend.

use axum::{
    extract::{Query, State},
    response::IntoResponse,
    Json,
};
use std::sync::Arc;

use super::types::{AppState, ErrorResponse};

/// Query parameters for cost record queries.
#[derive(Debug, serde::Deserialize)]
pub struct CostQuery {
    /// Tenant ID to filter by.
    pub tenant_id: Option<String>,
    /// Graph ID to filter by.
    pub graph_id: Option<String>,
    /// Maximum records to return.
    pub limit: Option<usize>,
}

/// Get current cost statistics.
///
/// GET /api/billing/costs
pub async fn get_costs(
    State(state): State<Arc<AppState>>,
    Query(query): Query<CostQuery>,
) -> impl IntoResponse {
    let Some(ref cost_attributor) = state.cost_attributor else {
        return ErrorResponse {
            error: "Cost attribution not configured".to_string(),
            correlation_id: None,
            recovery_suggestions: vec!["Check CostAttributor initialization".to_string()],
        }.into_response();
    };

    let buffer_len = cost_attributor.buffer_len().await;

    // Return current buffer stats (actual records require persistent storage)
    Json(serde_json::json!({
        "buffered_records": buffer_len,
        "tenant_filter": query.tenant_id,
        "graph_filter": query.graph_id,
        "note": "Detailed records are flushed to Go backend. Query ClickHouse for historical data.",
        "endpoints": {
            "flush": "/api/billing/costs/flush",
            "estimate": "/api/billing/estimate",
        }
    })).into_response()
}

/// Manually flush buffered cost records to the Go backend.
///
/// POST /api/billing/costs/flush
pub async fn flush_costs(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let Some(ref cost_attributor) = state.cost_attributor else {
        return ErrorResponse {
            error: "Cost attribution not configured".to_string(),
            correlation_id: None,
            recovery_suggestions: vec![],
        }.into_response();
    };

    // Trigger flush
    cost_attributor.flush().await;

    let remaining = cost_attributor.buffer_len().await;

    Json(serde_json::json!({
        "success": true,
        "status": "flushed",
        "remaining_buffer": remaining,
        "destination": std::env::var("ORCHESTRATOR_COST_URL")
            .unwrap_or_else(|_| "http://localhost:8080/api/costs".to_string()),
    })).into_response()
}

/// Estimate cost for a hypothetical execution.
///
/// GET /api/billing/estimate?model=gpt-4&prompt_tokens=1000&completion_tokens=500
pub async fn estimate_cost(
    Query(params): Query<CostEstimateParams>,
) -> impl IntoResponse {
    use crate::observability::cost::estimate_cost;
    use crate::router::flymind::Usage;

    let model = params.model.unwrap_or_else(|| "gpt-4".to_string());
    let prompt_tokens = params.prompt_tokens.unwrap_or(0);
    let completion_tokens = params.completion_tokens.unwrap_or(0);

    let usage = Usage {
        prompt_tokens,
        completion_tokens,
        total_tokens: prompt_tokens + completion_tokens,
    };

    let cost_usd = estimate_cost(&usage, &model);

    Json(serde_json::json!({
        "model": model,
        "prompt_tokens": prompt_tokens,
        "completion_tokens": completion_tokens,
        "total_tokens": usage.total_tokens,
        "estimated_cost_usd": cost_usd,
        "pricing_note": "Approximate pricing based on FlyMind provider rates. Actual cost may vary.",
    })).into_response()
}

/// Query parameters for cost estimation.
#[derive(Debug, serde::Deserialize)]
pub struct CostEstimateParams {
    pub model: Option<String>,
    pub prompt_tokens: Option<u32>,
    pub completion_tokens: Option<u32>,
}
