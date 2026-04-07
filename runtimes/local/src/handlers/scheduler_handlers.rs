//! Scheduler-related handlers (Phase 4: bin-packing scheduler).

use axum::{extract::{State, Path}, response::IntoResponse, Json};
use std::sync::Arc;

use super::types::AppState;
use crate::scheduler::{NodeCapacity, SchedulingDecision, SchedulingRequest};

/// Scheduler status handler.
pub async fn scheduler_status(State(state): State<Arc<AppState>>) -> axum::response::Response {
    let scheduler_info = if let Some(ref sched) = state.scheduler {
        let stats = sched.stats().await;
        let capacities = sched.node_capacities().await;
        Some(serde_json::json!({
            "scheduler_stats": stats,
            "node_capacities": capacities,
            "total_nodes": stats.total_nodes,
            "healthy_nodes": stats.healthy_nodes,
            "active_executions": stats.total_active_executions,
            "average_utilisation": stats.average_utilisation
        }))
    } else {
        None
    };

    Json(serde_json::json!({
        "scheduler": scheduler_info,
        "enterprise_enabled": state.config.enterprise_enabled,
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

/// Scheduling simulation endpoint — demonstrates SchedulingRequest and NodeCapacity usage.
pub async fn scheduling_simulate(State(state): State<Arc<AppState>>) -> axum::response::Response {
    // Use SchedulingRequest::from_resources for convenience
    let request = SchedulingRequest::from_resources(256, 1);

    // If scheduler is enabled, use it to get actual scheduling decisions
    let scheduling_decisions: Vec<serde_json::Value> = if let Some(ref sched) = state.scheduler {
        // Register example nodes if they don't exist
        let node1 = NodeCapacity::new("sim-node-1", 4000, 8192, 20);
        let node2 = NodeCapacity::new("sim-node-2", 2000, 4096, 10);
        sched.upsert_node(node1).await;
        sched.upsert_node(node2).await;

        // Use the actual scheduler's schedule method to get SchedulingDecision
        let mut decisions: Vec<serde_json::Value> = Vec::new();

        // Simulate scheduling with preferred node
        let mut req_with_preferred = request.clone();
        req_with_preferred.preferred_node = Some("sim-node-1".to_string());
        if let Some(SchedulingDecision { node_id, utilisation_score }) = sched.schedule(&req_with_preferred).await {
            decisions.push(serde_json::json!({
                "request": {
                    "cpu_millicores": req_with_preferred.cpu_millicores,
                    "memory_mb": req_with_preferred.memory_mb,
                    "preferred_node": req_with_preferred.preferred_node
                },
                "decision": {
                    "node_id": node_id,
                    "utilisation_score": utilisation_score
                }
            }));
        }

        // Simulate scheduling without preferred node (let scheduler decide)
        if let Some(SchedulingDecision { node_id, utilisation_score }) = sched.schedule(&request).await {
            decisions.push(serde_json::json!({
                "request": {
                    "cpu_millicores": request.cpu_millicores,
                    "memory_mb": request.memory_mb,
                    "preferred_node": request.preferred_node
                },
                "decision": {
                    "node_id": node_id,
                    "utilisation_score": utilisation_score
                }
            }));
        }

        decisions
    } else {
        Vec::new()
    };

    // Create example node capacities for display
    let node1 = NodeCapacity::new("sim-node-1", 4000, 8192, 20);
    let node2 = NodeCapacity::new("sim-node-2", 2000, 4096, 10);

    // Calculate if request fits
    let fits_node1 = node1.can_fit(&request);
    let fits_node2 = node2.can_fit(&request);

    Json(serde_json::json!({
        "scheduling_request": {
            "cpu_millicores": request.cpu_millicores,
            "memory_mb": request.memory_mb,
            "preferred_node": request.preferred_node,
            "created_via": "SchedulingRequest::from_resources"
        },
        "scheduling_decisions": scheduling_decisions,
        "nodes": [
            {
                "node_id": node1.node_id,
                "total_cpu_millicores": node1.total_cpu_millicores,
                "total_memory_mb": node1.total_memory_mb,
                "max_executions": node1.max_executions,
                "available_cpu_millicores": node1.available_cpu_millicores(),
                "available_memory_mb": node1.available_memory_mb(),
                "utilisation_score": node1.utilisation_score(),
                "can_fit_request": fits_node1
            },
            {
                "node_id": node2.node_id,
                "total_cpu_millicores": node2.total_cpu_millicores,
                "total_memory_mb": node2.total_memory_mb,
                "max_executions": node2.max_executions,
                "available_cpu_millicores": node2.available_cpu_millicores(),
                "available_memory_mb": node2.available_memory_mb(),
                "utilisation_score": node2.utilisation_score(),
                "can_fit_request": fits_node2
            }
        ],
        "timestamp": std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
    }))
    .into_response()
}

/// Mark a node as unhealthy so it stops receiving new requests.
pub async fn scheduler_mark_unhealthy(
    State(state): State<Arc<AppState>>,
    Path(node_id): Path<String>,
) -> axum::response::Response {
    if let Some(ref sched) = state.scheduler {
        sched.mark_unhealthy(&node_id).await;
        Json(serde_json::json!({
            "success": true,
            "node_id": node_id,
            "status": "marked_unhealthy"
        }))
        .into_response()
    } else {
        Json(serde_json::json!({
            "success": false,
            "error": "Scheduler not enabled"
        }))
        .into_response()
    }
}

/// Mark a node as healthy again.
pub async fn scheduler_mark_healthy(
    State(state): State<Arc<AppState>>,
    Path(node_id): Path<String>,
) -> axum::response::Response {
    if let Some(ref sched) = state.scheduler {
        sched.mark_healthy(&node_id).await;
        Json(serde_json::json!({
            "success": true,
            "node_id": node_id,
            "status": "marked_healthy"
        }))
        .into_response()
    } else {
        Json(serde_json::json!({
            "success": false,
            "error": "Scheduler not enabled"
        }))
        .into_response()
    }
}

/// Remove a node from the scheduler.
pub async fn scheduler_remove_node(
    State(state): State<Arc<AppState>>,
    Path(node_id): Path<String>,
) -> axum::response::Response {
    if let Some(ref sched) = state.scheduler {
        sched.remove_node(&node_id).await;
        Json(serde_json::json!({
            "success": true,
            "node_id": node_id,
            "status": "removed"
        }))
        .into_response()
    } else {
        Json(serde_json::json!({
            "success": false,
            "error": "Scheduler not enabled"
        }))
        .into_response()
    }
}
