//! Graph execution HTTP handler.
//!
//! Exposes the DAG executor via HTTP for the Go orchestrator to call.
//! This is the sidecar interface: Go sends a graph + input, Rust runs the execution.

use std::sync::Arc;

use axum::{
    extract::{Extension, Json, Path},
    response::IntoResponse,
    routing::{get, post},
    Router,
};
use tracing::{info, warn};
use uuid::Uuid;

use crate::agent_scheduler::{
    PriorityLevel,
    worker::{JobStatus as WorkerJobStatus, QueuedGraphExecution},
};
use crate::engine::graph::{
    DefaultNodeExecutor, Edge, EdgeType, ExecutionPriority, Graph, GraphExecutionInput,
    GraphExecutor, Node, NodeId, NodeType, RetryPolicy,
};
use crate::handlers::types::{AppState, ErrorResponse};

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

/// Request to execute a graph.
#[derive(Debug, serde::Deserialize)]
pub struct ExecuteGraphRequest {
    /// Unique ID for this execution.
    pub execution_id: Option<Uuid>,
    /// Graph definition (nodes + edges).
    pub graph: GraphDefinition,
    /// Initial input to the graph.
    pub initial_input: Option<std::collections::HashMap<String, serde_json::Value>>,
    /// Tenant ID for memory tier isolation.
    pub tenant_id: Option<String>,
    /// Priority (controls scheduler queue position).
    pub priority: Option<String>,
}

#[derive(Debug, serde::Deserialize)]
pub struct GraphDefinition {
    pub id: Uuid,
    pub name: String,
    pub nodes: Vec<NodeDefinition>,
    pub edges: Vec<EdgeDefinition>,
}

#[derive(Debug, serde::Deserialize)]
pub struct NodeDefinition {
    pub id: Uuid,
    pub name: String,
    #[serde(rename = "type")]
    pub node_type: String,
    pub timeout_ms: Option<u64>,
    pub max_attempts: Option<u32>,
    pub model: Option<String>,
    pub prompt: Option<String>,
    pub temperature: Option<f32>,
    pub max_tokens: Option<u32>,
    pub traffic_type: Option<String>,
    pub tool: Option<String>,
    pub params: Option<serde_json::Value>,
    pub memory_operation: Option<String>,
    pub memory_key: Option<String>,
    pub control_kind: Option<String>,
    pub optimization_strategy: Option<String>,
    /// Action connector name (e.g., "stripe", "resend")
    pub connector: Option<String>,
    /// Action name to execute on the connector
    pub action: Option<String>,
}

#[derive(Debug, serde::Deserialize)]
pub struct EdgeDefinition {
    pub source: Uuid,
    pub target: Uuid,
    #[serde(rename = "type")]
    pub edge_type: Option<String>,
    pub mapping: Option<String>,
}

/// Response from a graph execution.
#[derive(Debug, serde::Serialize)]
pub struct ExecuteGraphResponse {
    pub execution_id: Uuid,
    pub status: String,
    pub output: Option<serde_json::Value>,
    pub error: Option<String>,
    pub node_results: Vec<NodeResultDto>,
    pub duration_ms: Option<u64>,
}

#[derive(Debug, serde::Serialize)]
pub struct NodeResultDto {
    pub node_id: Uuid,
    pub output: Option<serde_json::Value>,
    pub error: Option<String>,
    pub duration_ms: u64,
    pub attempts: u32,
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

/// Execute a graph via the DAG executor (synchronous/inline execution).
///
/// For scheduled async execution, use `/execute/graph/scheduled`.
pub async fn execute_graph(
    Extension(state): Extension<Arc<AppState>>,
    Json(request): Json<ExecuteGraphRequest>,
) -> impl IntoResponse {
    let execution_id = request.execution_id.unwrap_or_else(Uuid::new_v4);

    info!(
        execution_id = %execution_id,
        graph_id = %request.graph.id,
        node_count = request.graph.nodes.len(),
        "Graph execution request received (inline)"
    );

    // Check scheduler backpressure if available (Phase 5).
    // This prevents the system from being overwhelmed during high load.
    if let Some(ref scheduler) = state.agent_scheduler {
        let sched = scheduler.read().await;
        let bp = sched.check_backpressure();
        if bp.is_rejected() {
            return (axum::http::StatusCode::SERVICE_UNAVAILABLE, Json(serde_json::json!({
                "error": format!("System overloaded: {}", bp.label()),
                "retry_after_secs": 30,
            }))).into_response();
        }
    }

    // Build the Graph struct from the request.
    let graph = match build_graph(&request.graph) {
        Ok(g) => g,
        Err(e) => {
            return ErrorResponse {
                error: format!("Invalid graph definition: {}", e),
                correlation_id: Some(execution_id.to_string()),
                recovery_suggestions: vec![],
            }.into_response();
        }
    };

    // Build the initial input.
    let initial_input = request.initial_input.unwrap_or_default();

    let input = GraphExecutionInput {
        graph_id: graph.id,
        initial_input,
        tenant_id: request.tenant_id.clone(),
    };

    // Create execution context with graph_id for cost attribution (Phase 6).
    let ctx = Arc::new(crate::engine::graph::ExecutionContext::with_graph_id(
        execution_id,
        graph.id,
        request.tenant_id.clone(),
    ));

    // Execute the graph using the integrated SAR executor (if available), or fall back to stubs.
    // Enable cost attribution when CostAttributor is configured (Phase 6: Observability).
    let result = if let Some(ref sar) = state.sar_executor {
        info!("Using SarNodeExecutor with real service implementations");
        let executor = if let Some(ref cost_attr) = state.cost_attributor {
            GraphExecutor::with_cost_attributor(sar.as_ref(), cost_attr.clone())
        } else {
            GraphExecutor::new(sar.as_ref())
        };
        executor.execute(&graph, input, ctx).await
    } else {
        info!("Using DefaultNodeExecutor with FlyMind + HotMemory + Action connectors");
        let executor = GraphExecutor::new(DefaultNodeExecutor::new());
        executor.execute(&graph, input, ctx).await
    };

    // Build response.
    let response = ExecuteGraphResponse {
        execution_id,
        status: format!("{:?}", result.status),
        output: result.output.map(|m| serde_json::Value::Object(m.into_iter().collect())),
        error: result.error,
        node_results: result
            .node_results
            .into_iter()
            .map(|(node_id, nr)| NodeResultDto {
                node_id: node_id.0,
                output: nr.output,
                error: nr.error,
                duration_ms: nr.duration_ms,
                attempts: nr.attempts,
            })
            .collect(),
        duration_ms: result.total_duration_ms,
    };

    info!(
        execution_id = %execution_id,
        status = %response.status,
        duration_ms = ?response.duration_ms,
        "Graph execution completed"
    );

    (axum::http::StatusCode::OK, Json(response)).into_response()
}

// ---------------------------------------------------------------------------
// Graph builder helpers
// ---------------------------------------------------------------------------

fn build_graph(def: &GraphDefinition) -> Result<Graph, String> {
    let mut graph = Graph::new(def.id, def.name.clone());

    // First pass: add all nodes.
    for node_def in &def.nodes {
        let node = build_node(node_def).map_err(|e| e.to_string())?;
        graph.add_node(node);
    }

    // Second pass: add all edges (after all nodes exist so we can validate).
    for edge_def in &def.edges {
        let source = NodeId(edge_def.source);
        let target = NodeId(edge_def.target);

        // Validate nodes exist.
        if !graph.nodes.contains_key(&source) {
            return Err(format!("Edge references unknown source node: {}", source));
        }
        if !graph.nodes.contains_key(&target) {
            return Err(format!("Edge references unknown target node: {}", target));
        }

        let edge_type = match edge_def.edge_type.as_deref() {
            Some("trigger") => EdgeType::Trigger,
            Some("dependency") => EdgeType::Dependency,
            _ => EdgeType::DataFlow,
        };

        let mut edge = Edge::new(source, target, edge_type);
        if let Some(mapping) = &edge_def.mapping {
            edge = edge.with_mapping(mapping.clone());
        }
        graph.add_edge(edge);
    }

    // Validate no cycles.
    // We check all edges added don't create a cycle by re-using detect_cycle.
    for edge in &graph.edges {
        if graph.detect_cycle(edge.source, edge.target) {
            return Err(format!(
                "Adding edge {} -> {} would create a cycle",
                edge.source, edge.target
            ));
        }
    }

    Ok(graph)
}

fn build_node(def: &NodeDefinition) -> Result<Node, String> {
    let node_type = match def.node_type.as_str() {
        "llm" => NodeType::LLM {
            model: def.model.clone(),
            prompt: def.prompt.clone().unwrap_or_default(),
            temperature: def.temperature.unwrap_or(0.7),
            max_tokens: def.max_tokens,
            traffic_type: match def.traffic_type.as_deref() {
                Some("realtime") => crate::engine::graph::LlmTrafficType::Realtime,
                Some("structured") => crate::engine::graph::LlmTrafficType::Structured,
                Some("function_calling") => crate::engine::graph::LlmTrafficType::FunctionCalling,
                Some("background") => crate::engine::graph::LlmTrafficType::Background,
                _ => crate::engine::graph::LlmTrafficType::General,
            },
        },
        "tool" => NodeType::Tool {
            name: def.tool.clone().unwrap_or_default(),
            params: def.params.clone().unwrap_or(serde_json::Value::Null),
        },
        "memory" => {
            let op = match def.memory_operation.as_deref() {
                Some("write") => crate::engine::graph::MemoryOp::Write,
                Some("delete") => crate::engine::graph::MemoryOp::Delete,
                Some("list") => crate::engine::graph::MemoryOp::List,
                _ => crate::engine::graph::MemoryOp::Read,
            };
            NodeType::Memory {
                operation: op,
                key: def.memory_key.clone().unwrap_or_default(),
            }
        }
        "control" => NodeType::Control {
            kind: match def.control_kind.as_deref() {
                Some("loop") => crate::engine::graph::ControlKind::Loop,
                Some("switch") => crate::engine::graph::ControlKind::Switch,
                _ => crate::engine::graph::ControlKind::If,
            },
            condition: crate::engine::graph::Expr::Const(true),
        },
        "optimization" => NodeType::Optimization {
            strategy: match def.optimization_strategy.as_deref() {
                Some("enable_caching") => crate::engine::graph::OptStrategy::EnableCaching,
                Some("increase_quota") => crate::engine::graph::OptStrategy::IncreaseQuota,
                Some("simplify_path") => crate::engine::graph::OptStrategy::SimplifyPath,
                _ => crate::engine::graph::OptStrategy::AdjustTimeouts,
            },
        },
        "action" => NodeType::Action {
            connector: def.connector.clone().unwrap_or_default(),
            action: def.action.clone().unwrap_or_default(),
            params: def.params.clone().unwrap_or(serde_json::Value::Null),
        },
        _ => NodeType::Passthrough,
    };

    let retry = RetryPolicy {
        max_attempts: def.max_attempts.unwrap_or(3),
        initial_delay_ms: 100,
        max_delay_ms: 10_000,
        backoff_multiplier: 2.0,
    };

    let mut node = Node::new(NodeId(def.id), def.name.clone(), node_type);
    node.retry = retry;
    if let Some(timeout_ms) = def.timeout_ms {
        node = node.with_timeout(timeout_ms);
    }

    Ok(node)
}

// ---------------------------------------------------------------------------
// Scheduled Async Execution Endpoint
// ---------------------------------------------------------------------------

/// Request to enqueue a graph for async execution.
#[derive(Debug, serde::Deserialize)]
pub struct ScheduleGraphRequest {
    #[serde(flatten)]
    pub execute_request: ExecuteGraphRequest,
    /// Priority for queue placement (defaults to Normal).
    pub priority: Option<String>,
}

/// Response from a scheduled graph execution.
#[derive(Debug, serde::Serialize)]
pub struct ScheduleGraphResponse {
    pub job_id: Uuid,
    pub status: String,
    pub queued_at: String,
    pub estimated_start: Option<String>,
    pub queue_position: Option<usize>,
}

/// Response from a job status query.
#[derive(Debug, serde::Serialize)]
pub struct JobStatusResponse {
    pub job_id: Uuid,
    pub graph_id: Uuid,
    pub status: String,
    pub queued_at: String,
    pub started_at: Option<String>,
    pub completed_at: Option<String>,
    pub duration_ms: Option<u64>,
    pub queue_wait_ms: Option<u64>,
    pub result: Option<serde_json::Value>,
    pub error: Option<String>,
}

/// Enqueue a graph for async execution via the scheduler.
pub async fn schedule_graph(
    Extension(state): Extension<Arc<AppState>>,
    Json(request): Json<ScheduleGraphRequest>,
) -> impl IntoResponse {
    let execution_id = request.execute_request.execution_id.unwrap_or_else(Uuid::new_v4);
    let priority = parse_priority(request.priority.as_deref());

    // Build the graph
    let graph = match build_graph(&request.execute_request.graph) {
        Ok(g) => g,
        Err(e) => {
            return ErrorResponse {
                error: format!("Invalid graph definition: {}", e),
                correlation_id: Some(execution_id.to_string()),
                recovery_suggestions: vec!["Check graph JSON structure".to_string()],
            }.into_response();
        }
    };

    // Check if we have a scheduler
    let Some(ref scheduler) = state.agent_scheduler else {
        return ErrorResponse {
            error: "Agent scheduler not configured".to_string(),
            correlation_id: Some(execution_id.to_string()),
            recovery_suggestions: vec!["Use /execute/graph for inline execution".to_string()],
        }.into_response();
    };

    // Check backpressure
    {
        let sched = scheduler.read().await;
        let bp = sched.check_backpressure();
        if bp.is_rejected() {
            return (axum::http::StatusCode::SERVICE_UNAVAILABLE, Json(serde_json::json!({
                "error": format!("System overloaded: {}", bp.label()),
                "retry_after_secs": 30,
            }))).into_response();
        }
    }

    // Clone initial_input for reuse
    let initial_input = request.execute_request.initial_input.clone().unwrap_or_default();

    // Create queued execution with full graph definition
    // Map ExecutionPriority to PriorityLevel for the scheduler
    let priority_level = match priority {
        ExecutionPriority::Critical => PriorityLevel::Critical,
        ExecutionPriority::High => PriorityLevel::High,
        ExecutionPriority::Normal => PriorityLevel::Normal,
        ExecutionPriority::Low => PriorityLevel::Low,
    };
    let exec = QueuedGraphExecution::new(
        graph,
        GraphExecutionInput {
            graph_id: request.execute_request.graph.id,
            initial_input: initial_input.clone(),
            tenant_id: request.execute_request.tenant_id.clone(),
        },
        priority_level,
        request.execute_request.tenant_id.clone(),
    );

    let job_id = exec.id;
    let _queued_at = std::time::Instant::now();

    // Track the job
    if let Some(ref tracker) = state.job_tracker {
        tracker.track(job_id, request.execute_request.graph.id).await;
    }

    // Enqueue the job with full graph definition
    // The scheduler now accepts QueuedGraphExecution which includes the complete graph
    {
        let mut sched = scheduler.write().await;
        match sched.enqueue(exec).await {
            Ok(enqueued_id) => {
                info!(job_id = %enqueued_id, "Graph execution enqueued with full definition");
            }
            Err(e) => {
                warn!(job_id = %job_id, error = %e, "Failed to enqueue graph execution");
                return ErrorResponse {
                    error: format!("Failed to enqueue: {}", e),
                    correlation_id: Some(execution_id.to_string()),
                    recovery_suggestions: vec!["Try again later".to_string()],
                }.into_response();
            }
        }
    }

    // Build response
    let response = ScheduleGraphResponse {
        job_id,
        status: "queued".to_string(),
        queued_at: chrono::Utc::now().to_rfc3339(),
        estimated_start: None,
        queue_position: None,
    };

    (axum::http::StatusCode::ACCEPTED, Json(response)).into_response()
}

/// Get the status of a scheduled job.
pub async fn get_job_status(
    Extension(state): Extension<Arc<AppState>>,
    Path(job_id): Path<Uuid>,
) -> impl IntoResponse {
    let Some(ref tracker) = state.job_tracker else {
        return ErrorResponse {
            error: "Job tracking not configured".to_string(),
            correlation_id: Some(job_id.to_string()),
            recovery_suggestions: vec!["Job tracker not initialized".to_string()],
        }.into_response();
    };

    let job = tracker.get_job(job_id).await;

    match job {
        Some(j) => {
            let status_str = match j.status {
                WorkerJobStatus::Queued => "queued",
                WorkerJobStatus::Processing => "processing",
                WorkerJobStatus::Completed => "completed",
                WorkerJobStatus::Failed => "failed",
                WorkerJobStatus::Cancelled => "cancelled",
            }.to_string();

            let response = JobStatusResponse {
                job_id: j.id,
                graph_id: j.graph_id,
                status: status_str,
                queued_at: format!("{:?}", j.queued_at),
                started_at: j.started_at.map(|i| format!("{:?}", i)),
                completed_at: j.completed_at.map(|i| format!("{:?}", i)),
                duration_ms: j.completed_at.zip(j.started_at).map(|(c, s)| {
                    c.duration_since(s).as_millis() as u64
                }),
                queue_wait_ms: j.started_at.map(|s| s.duration_since(j.queued_at).as_millis() as u64),
                result: j.result.as_ref().map(|r| {
                    serde_json::json!({
                        "status": format!("{:?}", r.status),
                        "output": r.output.as_ref().map(|o| serde_json::Value::Object(o.clone().into_iter().collect())),
                    })
                }),
                error: j.error,
            };
            (axum::http::StatusCode::OK, Json(response)).into_response()
        }
        None => {
            ErrorResponse {
                error: format!("Job {} not found", job_id),
                correlation_id: Some(job_id.to_string()),
                recovery_suggestions: vec!["Job may have expired or ID is incorrect".to_string()],
            }.into_response()
        }
    }
}

/// Parse priority string into ExecutionPriority.
fn parse_priority(priority: Option<&str>) -> ExecutionPriority {
    match priority {
        Some("critical") => ExecutionPriority::Critical,
        Some("high") => ExecutionPriority::High,
        Some("low") => ExecutionPriority::Low,
        _ => ExecutionPriority::Normal,
    }
}

// ---------------------------------------------------------------------------
// Router registration helper
// ---------------------------------------------------------------------------

/// Add graph execution routes to the router.
pub fn routes() -> Router {
    Router::new()
        .route("/execute/graph", post(execute_graph))
        .route("/execute/graph/scheduled", post(schedule_graph))
        .route("/execute/graph/status/{job_id}", get(get_job_status))
}
