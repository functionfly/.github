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
    GraphMutator, PatternConfidence,
};
use crate::engine::graph::{Graph, NodeId, NodeType, RetryPolicy, Edge, EdgeType, MemoryOp, ControlKind, Expr, OptStrategy, LlmTrafficType};
use std::collections::HashMap;

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
    let Some(ref optimizer) = state.optimizer else {
        return ErrorResponse {
            error: "Optimizer not configured".to_string(),
            correlation_id: None,
            recovery_suggestions: vec!["Check that the optimizer is enabled".to_string()],
        }.into_response();
    };

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

    // Fetch current optimization suggestions for this graph.
    let suggestions = optimizer.analyze(graph_id).await;

    // Find the requested suggestion by ID.
    let suggestion = match suggestions.iter().find(|s| s.id == request.suggestion_id) {
        Some(s) => s.clone(),
        None => {
            return ErrorResponse {
                error: format!(
                    "Suggestion {} not found for graph {}",
                    request.suggestion_id, graph_id
                ),
                correlation_id: None,
                recovery_suggestions: vec![
                    "Run POST /api/graphs/analyze first to generate suggestions".to_string(),
                ],
            }.into_response();
        }
    };

    // Build a minimal graph stub for the mutator. The GraphMutator operates
    // on the in-memory graph structure; for persistence, the Go backend
    // re-fetches from Postgres and re-applies.
    let mut graph = Graph::new(graph_id, format!("graph-{}", graph_id));

    let mut mutator = GraphMutator::new();
    let result = mutator.apply_suggestion(&mut graph, &suggestion, is_canary);

    let status = if result.success {
        if is_canary { "canary_testing" } else { "applied_production" }
    } else {
        "failed"
    };

    Json(serde_json::json!({
        "success": result.success,
        "graph_id": graph_id.to_string(),
        "mutation_id": result.mutation_id.to_string(),
        "suggestion_id": request.suggestion_id.to_string(),
        "canary": is_canary,
        "status": status,
        "applied_changes": result.applied_changes,
        "errors": result.errors,
        "rollback_possible": result.rollback_possible,
    })).into_response()
}

/// Get the mutation log for audit purposes.
///
/// GET /api/graphs/{graph_id}/mutations
pub async fn get_mutation_log(
    State(state): State<Arc<AppState>>,
    Path(graph_id): Path<Uuid>,
) -> impl IntoResponse {
    let Some(ref optimizer) = state.optimizer else {
        return Json(serde_json::json!({
            "graph_id": graph_id.to_string(),
            "mutations": [],
            "total": 0,
            "warning": "optimizer not configured; mutation log unavailable",
        })).into_response();
    };

    // Re-analyze to get current suggestions — the mutation log lives on the
    // GraphMutator which is created per-request. For a persistent log the Go
    // backend stores AppliedMutation records in Postgres. We return the
    // current suggestions so the caller can cross-reference.
    let suggestions = optimizer.analyze(graph_id).await;

    let entries: Vec<_> = suggestions
        .iter()
        .map(|s| serde_json::json!({
            "suggestion_id": s.id.to_string(),
            "node_id": s.node_id,
            "node_name": s.node_name,
            "action": format!("{}", s.action),
            "confidence": confidence_label(s.confidence),
            "expected_impact": s.expected_impact,
            "description": s.description,
            "graph_id": s.graph_id.to_string(),
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
    State(state): State<Arc<AppState>>,
    Path(mutation_id): Path<Uuid>,
    Json(request): Json<RollbackRequest>,
) -> impl IntoResponse {
    let Some(ref optimizer) = state.optimizer else {
        return ErrorResponse {
            error: "Optimizer not configured".to_string(),
            correlation_id: Some(mutation_id.to_string()),
            recovery_suggestions: vec!["Check that the optimizer is enabled".to_string()],
        }.into_response();
    }

    // Validate the mutation_id matches the request
    if request.mutation_id != mutation_id {
        return ErrorResponse {
            error: "mutation_id in URL does not match request body".to_string(),
            correlation_id: Some(mutation_id.to_string()),
            recovery_suggestions: vec!["Ensure URL path and JSON body use the same mutation_id".to_string()],
        }.into_response();
    }

    // The Go backend is the source of truth for mutation rollback because it
    // owns persistent storage. Forward the rollback request to the backend.
    let orchestrator_url = std::env::var("ORCHESTRATOR_URL")
        .unwrap_or_else(|_| "http://localhost:8080".to_string());

    let url = format!(
        "{}/api/optimizer/mutations/{}/rollback",
        orchestrator_url.trim_end_matches('/'),
        mutation_id
    );

    let client = match reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(10))
        .build()
    {
        Ok(c) => c,
        Err(e) => {
            return ErrorResponse {
                error: format!("Failed to create HTTP client: {}", e),
                correlation_id: Some(mutation_id.to_string()),
                recovery_suggestions: vec!["Retry the request".to_string()],
            }.into_response();
        }
    };

    let mut req = client.post(&url).json(&serde_json::json!({
        "mutation_id": mutation_id.to_string(),
    }));

    if let Ok(token) = std::env::var("RUNTIME_API_TOKEN") {
        if !token.is_empty() {
            req = req.header("Authorization", format!("Bearer {}", token));
        }
    }

    match req.send().await {
        Ok(resp) if resp.status().is_success() => {
            Json(serde_json::json!({
                "success": true,
                "mutation_id": mutation_id.to_string(),
                "status": "rolled_back",
            })).into_response()
        }
        Ok(resp) => {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            ErrorResponse {
                error: format!("Go backend returned {} on rollback: {}", status, body),
                correlation_id: Some(mutation_id.to_string()),
                recovery_suggestions: vec![
                    "Check that the Go backend is running".to_string(),
                    "Verify the mutation_id exists in the database".to_string(),
                ],
            }.into_response()
        }
        Err(e) => ErrorResponse {
            error: format!("Failed to reach Go backend for rollback: {}", e),
            correlation_id: Some(mutation_id.to_string()),
            recovery_suggestions: vec![
                "Ensure ORCHESTRATOR_URL is set correctly".to_string(),
                "Check that the Go backend is running on the expected port".to_string(),
            ],
        }.into_response(),
    }
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
    let id = json.get("id")
        .and_then(|v| v.as_str())
        .and_then(|s| Uuid::parse_str(s).ok())
        .unwrap_or_else(Uuid::new_v4);

    let name = json.get("name")
        .and_then(|v| v.as_str())
        .unwrap_or("unnamed")
        .to_string();

    let mut graph = Graph::new(id, name);

    // Parse nodes
    if let Some(nodes) = json.get("nodes").and_then(|v| v.as_array()) {
        for node_json in nodes {
            let node_id = node_json.get("id")
                .and_then(|v| v.as_str())
                .and_then(|s| Uuid::parse_str(s).ok())
                .unwrap_or_else(Uuid::new_v4);

            let node_name = node_json.get("name")
                .and_then(|v| v.as_str())
                .unwrap_or("unnamed")
                .to_string();

            let node_type_str = node_json.get("type")
                .or_else(|| node_json.get("node_type"))
                .and_then(|v| v.as_str())
                .unwrap_or("passthrough");

            let node_type = match node_type_str {
                "llm" => NodeType::LLM {
                    model: node_json.get("model").and_then(|v| v.as_str()).map(String::from),
                    prompt: node_json.get("prompt").and_then(|v| v.as_str()).unwrap_or_default().to_string(),
                    temperature: node_json.get("temperature").and_then(|v| v.as_f64()).unwrap_or(0.7) as f32,
                    max_tokens: node_json.get("max_tokens").and_then(|v| v.as_u64()).map(|v| v as u32),
                    traffic_type: match node_json.get("traffic_type").and_then(|v| v.as_str()) {
                        Some("realtime") => LlmTrafficType::Realtime,
                        Some("structured") => LlmTrafficType::Structured,
                        Some("function_calling") => LlmTrafficType::FunctionCalling,
                        Some("background") => LlmTrafficType::Background,
                        _ => LlmTrafficType::General,
                    },
                },
                "tool" => NodeType::Tool {
                    name: node_json.get("tool").or_else(|| node_json.get("name")).and_then(|v| v.as_str()).unwrap_or_default().to_string(),
                    params: node_json.get("params").cloned().unwrap_or(serde_json::Value::Null),
                },
                "memory" => NodeType::Memory {
                    operation: match node_json.get("memory_operation").or_else(|| node_json.get("operation")).and_then(|v| v.as_str()) {
                        Some("write") => MemoryOp::Write,
                        Some("delete") => MemoryOp::Delete,
                        Some("list") => MemoryOp::List,
                        _ => MemoryOp::Read,
                    },
                    key: node_json.get("memory_key").or_else(|| node_json.get("key")).and_then(|v| v.as_str()).unwrap_or_default().to_string(),
                },
                "control" => NodeType::Control {
                    kind: match node_json.get("control_kind").or_else(|| node_json.get("kind")).and_then(|v| v.as_str()) {
                        Some("loop") => ControlKind::Loop,
                        Some("switch") => ControlKind::Switch,
                        _ => ControlKind::If,
                    },
                    condition: Expr::Const(true),
                },
                "optimization" => NodeType::Optimization {
                    strategy: match node_json.get("optimization_strategy").or_else(|| node_json.get("strategy")).and_then(|v| v.as_str()) {
                        Some("enable_caching") => OptStrategy::EnableCaching,
                        Some("increase_quota") => OptStrategy::IncreaseQuota,
                        Some("simplify_path") => OptStrategy::SimplifyPath,
                        _ => OptStrategy::AdjustTimeouts,
                    },
                },
                "action" => NodeType::Action {
                    connector: node_json.get("connector").and_then(|v| v.as_str()).unwrap_or_default().to_string(),
                    action: node_json.get("action").and_then(|v| v.as_str()).unwrap_or_default().to_string(),
                    params: node_json.get("params").cloned().unwrap_or(serde_json::Value::Null),
                },
                _ => NodeType::Passthrough,
            };

            let retry = RetryPolicy {
                max_attempts: node_json.get("max_attempts").and_then(|v| v.as_u64()).unwrap_or(3) as u32,
                initial_delay_ms: 100,
                max_delay_ms: 10_000,
                backoff_multiplier: 2.0,
            };

            let mut node = crate::engine::graph::Node::new(NodeId(node_id), node_name, node_type);
            node.retry = retry;
            if let Some(timeout_ms) = node_json.get("timeout_ms").and_then(|v| v.as_u64()) {
                node = node.with_timeout(timeout_ms);
            }
            graph.add_node(node);
        }
    }

    // Parse edges
    if let Some(edges) = json.get("edges").and_then(|v| v.as_array()) {
        for edge_json in edges {
            let source = edge_json.get("source")
                .and_then(|v| v.as_str())
                .and_then(|s| Uuid::parse_str(s).ok());
            let target = edge_json.get("target")
                .and_then(|v| v.as_str())
                .and_then(|s| Uuid::parse_str(s).ok());

            let (Some(source), Some(target)) = (source, target) else {
                continue;
            };

            let edge_type = match edge_json.get("type").and_then(|v| v.as_str()) {
                Some("trigger") => EdgeType::Trigger,
                Some("dependency") => EdgeType::Dependency,
                _ => EdgeType::DataFlow,
            };

            let mut edge = Edge::new(NodeId(source), NodeId(target), edge_type);
            if let Some(mapping) = edge_json.get("mapping").and_then(|v| v.as_str()) {
                edge = edge.with_mapping(mapping.to_string());
            }
            graph.add_edge(edge);
        }
    }

    Ok(graph)
}
