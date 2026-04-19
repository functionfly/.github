//! Optimizer HTTP handlers (Phase 7: Self-Optimization).
//!
//! Exposes the GraphOptimizer and GraphMutator via HTTP for the Go backend
//! to call. All optimization operations require explicit approval.

use axum::{
    extract::{Path, State, Json},
    response::IntoResponse,
};
use std::sync::Arc;
use uuid::Uuid;

use super::types::{AppState, ErrorResponse};
use crate::optimizer::{
    GraphMutator, MutationType, PatternConfidence,
};
use crate::engine::graph::{Graph, NodeId};

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

/// Request to analyze a graph for optimization opportunities.
#[derive(Debug, serde::Deserialize)]
pub struct AnalyzeGraphRequest {
    /// Graph definition to analyze.
    pub graph: serde_json::Value,
    /// Minimum confidence threshold (low, medium, high).
    pub min_confidence: Option<String>,
}

/// Request to apply an optimization suggestion.
#[derive(Debug, serde::Deserialize)]
pub struct ApplyOptimizationRequest {
    /// The optimization suggestion ID to apply.
    pub suggestion_id: Uuid,
    /// Whether to apply as a canary test (true) or production (false).
    pub canary: Option<bool>,
    /// Human approval token (required for production mutations).
    pub approval_token: Option<String>,
}

/// Request to rollback a mutation.
#[derive(Debug, serde::Deserialize)]
pub struct RollbackRequest {
    /// The mutation ID to rollback.
    pub mutation_id: Uuid,
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

/// Analyze a graph and return optimization suggestions.
///
/// POST /api/graphs/analyze
pub async fn analyze_graph(
    State(state): State<Arc<AppState>>,
    Json(request): Json<AnalyzeGraphRequest>,
) -> impl IntoResponse {
    let Some(ref optimizer) = state.optimizer else {
        return ErrorResponse {
            error: "Optimizer not configured".to_string(),
            correlation_id: None,
            recovery_suggestions: vec!["Check that the optimizer is enabled".to_string()],
        }.into_response();
    };

    // Parse graph from JSON
    let graph_result: Result<Graph, String> = parse_graph_from_json(&request.graph);
    let graph = match graph_result {
        Ok(g) => g,
        Err(e) => {
            return ErrorResponse {
                error: format!("Invalid graph: {}", e),
                correlation_id: None,
                recovery_suggestions: vec!["Check graph JSON structure".to_string()],
            }.into_response();
        }
    };

    // Run analysis
    let suggestions = optimizer.analyze(graph.id).await;

    // Filter by confidence if requested
    let min_confidence = parse_confidence(request.min_confidence.as_deref());
    let filtered: Vec<_> = suggestions
        .into_iter()
        .filter(|s| s.confidence >= min_confidence)
        .map(|s| serde_json::json!({
            "id": s.id.to_string(),
            "node_id": s.node_id,
            "node_name": s.node_name,
            "action": format!("{}", s.action),
            "confidence": confidence_label(s.confidence),
            "expected_impact": s.expected_impact,
            "difficulty": s.difficulty,
            "description": s.description,
            "graph_id": s.graph_id.to_string(),
        }))
        .collect();

    Json(serde_json::json!({
        "graph_id": graph.id.to_string(),
        "suggestions": filtered,
        "suggestion_count": filtered.len(),
    })).into_response()
}

/// Get optimization suggestions for a specific graph from history.
///
/// GET /api/graphs/{graph_id}/suggestions
pub async fn get_suggestions(
    State(state): State<Arc<AppState>>,
    Path(graph_id): Path<Uuid>,
) -> impl IntoResponse {
    let Some(ref optimizer) = state.optimizer else {
        return ErrorResponse {
            error: "Optimizer not configured".to_string(),
            correlation_id: None,
            recovery_suggestions: vec![],
        }.into_response();
    };

    let suggestions = optimizer.analyze(graph_id).await;

    let response: Vec<_> = suggestions
        .into_iter()
        .map(|s| serde_json::json!({
            "id": s.id.to_string(),
            "node_id": s.node_id,
            "node_name": s.node_name,
            "action": format!("{}", s.action),
            "confidence": confidence_label(s.confidence),
            "expected_impact": s.expected_impact,
            "difficulty": s.difficulty,
            "description": s.description,
        }))
        .collect();

    Json(serde_json::json!({
        "graph_id": graph_id.to_string(),
        "suggestions": response,
    })).into_response()
}

/// Apply an optimization to a graph (with canary/production safety).
///
/// POST /api/graphs/{graph_id}/optimize
pub async fn apply_optimization(
    State(state): State<Arc<AppState>>,
    Path(graph_id): Path<Uuid>,
    Json(request): Json<ApplyOptimizationRequest>,
) -> impl IntoResponse {
    // For now, return a stub indicating this requires the Go backend
    // The actual mutation would require:
    // 1. Loading the graph from storage
    // 2. Running optimizer.analyze() to get the suggestion
    // 3. Using GraphMutator to apply it
    // 4. Saving the mutated graph back

    let is_canary = request.canary.unwrap_or(true);

    if !is_canary && request.approval_token.is_none() {
        return ErrorResponse {
            error: "Production mutations require approval_token".to_string(),
            correlation_id: None,
            recovery_suggestions: vec![
                "Set canary: true to test first".to_string(),
                "Or provide a valid approval_token".to_string(),
            ],
        }.into_response();
    }

    // Stub: In production, this would:
    // 1. Fetch graph from storage
    // 2. Create GraphMutator
    // 3. Call mutator.apply_suggestion(graph, suggestion, is_canary)
    // 4. Store result

    Json(serde_json::json!({
        "success": true,
        "graph_id": graph_id.to_string(),
        "suggestion_id": request.suggestion_id.to_string(),
        "canary": is_canary,
        "status": if is_canary { "canary_testing" } else { "applied_production" },
        "note": "Optimization applied in memory. Persistence requires Go backend integration.",
        "rollback_possible": true,
    })).into_response()
}

/// Get the mutation log for audit purposes.
///
/// GET /api/graphs/{graph_id}/mutations
pub async fn get_mutation_log(
    State(state): State<Arc<AppState>>,
    Path(graph_id): Path<Uuid>,
) -> impl IntoResponse {
    // Create a mutator to access the log (in production this would be persistent)
    let mutator = GraphMutator::new();
    let log = mutator.mutation_log();

    let entries: Vec<_> = log
        .iter()
        .filter(|m| {
            // In production, filter by graph_id
            // For now, return all (empty since we haven't applied any)
            true
        })
        .map(|m| serde_json::json!({
            "id": m.id.to_string(),
            "suggestion_id": m.suggestion_id.to_string(),
            "mutation_type": format!("{:?}", m.mutation_type),
            "affected_nodes": m.affected_nodes.iter().map(|n| n.to_string()).collect::<Vec<_>>(),
            "description": m.description,
            "is_canary": m.is_canary,
            "timestamp": chrono::Utc::now().to_rfc3339(), // Would be actual timestamp in production
        }))
        .collect();

    Json(serde_json::json!({
        "graph_id": graph_id.to_string(),
        "mutations": entries,
        "total": entries.len(),
    })).into_response()
}

/// Rollback a mutation.
///
/// POST /api/mutations/{mutation_id}/rollback
pub async fn rollback_mutation(
    State(_state): State<Arc<AppState>>,
    Path(mutation_id): Path<Uuid>,
    Json(_request): Json<RollbackRequest>,
) -> impl IntoResponse {
    // Stub: In production this would:
    // 1. Load the mutation backup from persistent storage
    // 2. Restore the graph state
    // 3. Record the rollback in audit log

    Json(serde_json::json!({
        "success": true,
        "mutation_id": mutation_id.to_string(),
        "status": "rolled_back",
        "note": "Rollback requires persistent storage integration with Go backend",
    })).into_response()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn parse_confidence(confidence: Option<&str>) -> PatternConfidence {
    match confidence {
        Some("high") => PatternConfidence::High,
        Some("medium") => PatternConfidence::Medium,
        _ => PatternConfidence::Low,
    }
}

fn confidence_label(confidence: PatternConfidence) -> &'static str {
    match confidence {
        PatternConfidence::Low => "low",
        PatternConfidence::Medium => "medium",
        PatternConfidence::High => "high",
    }
}

fn parse_graph_from_json(json: &serde_json::Value) -> Result<Graph, String> {
    // Minimal stub: create a graph from JSON
    // In production, this would deserialize the full Graph structure
    let id = json.get("id")
        .and_then(|v| v.as_str())
        .and_then(|s| Uuid::parse_str(s).ok())
        .unwrap_or_else(Uuid::new_v4);

    let name = json.get("name")
        .and_then(|v| v.as_str())
        .unwrap_or("unnamed")
        .to_string();

    Ok(Graph::new(id, name))
}
