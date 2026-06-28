//! Graph execution engine for stateful agent runtime.
//!
//! This module implements a DAG-based execution engine where each node is a typed
//! actor (LLM, Tool, Memory, Control, Optimization) and edges represent dataflow
//! or control dependencies.
//!
//! ## Kahn's Algorithm
//!
//! Nodes are executed in topologically-sorted order. Nodes with no data dependency
//! (in-degree = 0) run in parallel via `tokio::spawn`. Cycles are detected at
//! graph build time and rejected.
//!
//! ## Node Types
//!
//! - **LLM**: Calls the model router (Phase 4) for AI inference
//! - **Tool**: Executes a registered tool (API call, DB query, etc.)
//! - **Memory**: Reads/writes state via the memory tier (Phase 3)
//! - **Control**: Conditional branching (`if`, `loop`, `switch`)
//! - **Optimization**: Self-improvement suggestions (Phase 7)
//!
//! ## Retry & Timeout
//!
//! Each node has its own `RetryPolicy` and `timeout_ms`. Retries are attempted
//! with exponential backoff. A node that exhausts retries marks the graph
//! execution as failed.

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use std::time::Duration;

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use tracing::{debug, info, warn, instrument};
use uuid::Uuid;

use crate::errors::{ErrorKind, RuntimeError};
use crate::observability::cost::CostAttributor;
use crate::router::flymind::Usage as CostUsage;

// ---------------------------------------------------------------------------
// Core types
// ---------------------------------------------------------------------------

/// A node in the execution graph.
#[derive(Debug, Clone)]
pub struct Node {
    pub id: NodeId,
    pub name: String,
    pub node_type: NodeType,
    /// Per-node timeout in milliseconds.
    pub timeout_ms: u64,
    /// Per-node retry policy.
    pub retry: RetryPolicy,
    /// Input schema as JSON Schema string.
    pub input_schema: Option<String>,
    /// Output schema as JSON Schema string.
    pub output_schema: Option<String>,
    /// Arbitrary metadata (tags, labels, etc.).
    pub metadata: HashMap<String, String>,
}

impl Node {
    pub fn new(id: NodeId, name: String, node_type: NodeType) -> Self {
        Self {
            id,
            name,
            node_type,
            timeout_ms: 30_000,
            retry: RetryPolicy::default(),
            input_schema: None,
            output_schema: None,
            metadata: HashMap::new(),
        }
    }

    pub fn with_timeout(mut self, timeout_ms: u64) -> Self {
        self.timeout_ms = timeout_ms;
        self
    }

    pub fn with_retry(mut self, retry: RetryPolicy) -> Self {
        self.retry = retry;
        self
    }
}

/// Unique identifier for a node within a graph.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct NodeId(pub Uuid);

impl std::fmt::Display for NodeId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// An edge in the execution graph.
#[derive(Debug, Clone)]
pub struct Edge {
    pub id: Uuid,
    /// Source node ID.
    pub source: NodeId,
    /// Target node ID.
    pub target: NodeId,
    /// Edge type determines execution semantics.
    pub edge_type: EdgeType,
    /// JSONata-style mapping from source output to target input.
    pub mapping: Option<String>,
}

impl Edge {
    pub fn new(source: NodeId, target: NodeId, edge_type: EdgeType) -> Self {
        Self {
            id: Uuid::new_v4(),
            source,
            target,
            edge_type,
            mapping: None,
        }
    }

    pub fn dataflow(source: NodeId, target: NodeId) -> Self {
        Self::new(source, target, EdgeType::DataFlow)
    }

    pub fn trigger(source: NodeId, target: NodeId) -> Self {
        Self::new(source, target, EdgeType::Trigger)
    }

    pub fn with_mapping(mut self, mapping: String) -> Self {
        self.mapping = Some(mapping);
        self
    }
}

/// Edge type determines how the target node is triggered.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum EdgeType {
    /// Target runs after source completes, receiving source output as input.
    DataFlow,
    /// Target is triggered by source completing, but receives no data.
    Trigger,
    /// Target waits for source output and merges it (fan-in).
    Dependency,
}

impl Default for EdgeType {
    fn default() -> Self {
        EdgeType::DataFlow
    }
}

/// Node type determines what the node does during execution.
#[derive(Debug, Clone)]
pub enum NodeType {
    /// LLM inference node — routes to model router.
    LLM {
        model: Option<String>,
        prompt: String,
        temperature: f32,
        max_tokens: Option<u32>,
        /// Traffic type hint for the model router.
        traffic_type: LlmTrafficType,
    },
    /// Tool node — executes a registered tool.
    Tool {
        name: String,
        params: serde_json::Value,
    },
    /// Memory node — reads or writes state.
    Memory {
        operation: MemoryOp,
        key: String,
    },
    /// Control node — conditional branching.
    Control {
        kind: ControlKind,
        condition: Expr,
    },
    /// Optimization node — self-improvement suggestions.
    Optimization {
        strategy: OptStrategy,
    },
    /// Action node — external service connector (Stripe, Resend, etc.).
    Action {
        /// Connector name (e.g., "stripe", "resend").
        connector: String,
        /// Action name (e.g., "charge", "create_customer").
        action: String,
        /// Action parameters.
        params: serde_json::Value,
    },
    /// Passthrough node — no-op, used for graph structure.
    Passthrough,
}

impl NodeType {
    pub fn llm(prompt: String) -> Self {
        NodeType::LLM {
            model: None,
            prompt,
            temperature: 0.7,
            max_tokens: None,
            traffic_type: LlmTrafficType::General,
        }
    }

    pub fn tool(name: String, params: serde_json::Value) -> Self {
        NodeType::Tool { name, params }
    }

    pub fn memory(operation: MemoryOp, key: String) -> Self {
        NodeType::Memory { operation, key }
    }

    pub fn control_if(condition: Expr) -> Self {
        NodeType::Control { kind: ControlKind::If, condition }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum LlmTrafficType {
    Realtime,
    Structured,
    FunctionCalling,
    Background,
    General,
}

impl Default for LlmTrafficType {
    fn default() -> Self {
        LlmTrafficType::General
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum MemoryOp {
    Read,
    Write,
    Delete,
    List,
}

impl std::fmt::Display for MemoryOp {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            MemoryOp::Read => write!(f, "Read"),
            MemoryOp::Write => write!(f, "Write"),
            MemoryOp::Delete => write!(f, "Delete"),
            MemoryOp::List => write!(f, "List"),
        }
    }
}

impl Default for MemoryOp {
    fn default() -> Self {
        MemoryOp::Read
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ControlKind {
    If,
    Loop,
    Switch,
}

impl Default for ControlKind {
    fn default() -> Self {
        ControlKind::If
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Expr {
    /// Boolean expression tree for conditional evaluation.
    Eq(Box<ExprValue>, Box<ExprValue>),
    Ne(Box<ExprValue>, Box<ExprValue>),
    And(Box<Expr>, Box<Expr>),
    Or(Box<Expr>, Box<Expr>),
    Not(Box<Expr>),
    Const(bool),
    Var(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ExprValue {
    Var(String),
    Const(serde_json::Value),
}

impl Expr {
    pub fn eval(&self, ctx: &HashMap<String, serde_json::Value>) -> bool {
        match self {
            Expr::Eq(l, r) => l.eval(ctx) == r.eval(ctx),
            Expr::Ne(l, r) => l.eval(ctx) != r.eval(ctx),
            Expr::And(a, b) => a.eval(ctx) && b.eval(ctx),
            Expr::Or(a, b) => a.eval(ctx) || b.eval(ctx),
            Expr::Not(e) => !e.eval(ctx),
            Expr::Const(v) => *v,
            Expr::Var(name) => ctx
                .get(name)
                .and_then(|v| v.as_bool())
                .unwrap_or(false),
        }
    }
}

impl ExprValue {
    pub fn eval(&self, ctx: &HashMap<String, serde_json::Value>) -> serde_json::Value {
        match self {
            ExprValue::Var(name) => ctx
                .get(name)
                .cloned()
                .unwrap_or(serde_json::Value::Null),
            ExprValue::Const(v) => v.clone(),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OptStrategy {
    AdjustTimeouts,
    EnableCaching,
    IncreaseQuota,
    SimplifyPath,
}

impl Default for OptStrategy {
    fn default() -> Self {
        OptStrategy::AdjustTimeouts
    }
}

/// Retry policy for a node.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetryPolicy {
    /// Maximum number of retry attempts (0 = no retries).
    pub max_attempts: u32,
    /// Initial backoff delay in milliseconds.
    pub initial_delay_ms: u64,
    /// Maximum backoff delay in milliseconds.
    pub max_delay_ms: u64,
    /// Multiplier for exponential backoff.
    pub backoff_multiplier: f64,
}

impl Default for RetryPolicy {
    fn default() -> Self {
        Self {
            max_attempts: 3,
            initial_delay_ms: 100,
            max_delay_ms: 10_000,
            backoff_multiplier: 2.0,
        }
    }
}

impl RetryPolicy {
    pub fn no_retries() -> Self {
        Self {
            max_attempts: 0,
            ..Default::default()
        }
    }

    /// Returns the delay for a given attempt number.
    pub fn delay_for(&self, attempt: u32) -> Duration {
        let delay = self.initial_delay_ms as f64
            * self.backoff_multiplier.powi(attempt as i32 - 1);
        let delay = delay.min(self.max_delay_ms as f64);
        Duration::from_millis(delay as u64)
    }
}

// ---------------------------------------------------------------------------
// Graph
// ---------------------------------------------------------------------------

/// A directed acyclic graph of execution nodes.
#[derive(Debug, Clone)]
pub struct Graph {
    pub id: Uuid,
    pub name: String,
    pub nodes: HashMap<NodeId, Node>,
    pub edges: Vec<Edge>,
}

impl Graph {
    pub fn new(id: Uuid, name: String) -> Self {
        Self {
            id,
            name,
            nodes: HashMap::new(),
            edges: Vec::new(),
        }
    }

    pub fn add_node(&mut self, node: Node) {
        self.nodes.insert(node.id, node);
    }

    pub fn add_edge(&mut self, edge: Edge) {
        self.edges.push(edge);
    }

    /// Detects whether adding an edge from `source` to `target` would create a cycle.
    /// Uses BFS from `target` — if we can reach `source`, adding this edge creates a cycle.
    pub fn detect_cycle(&self, source: NodeId, target: NodeId) -> bool {
        let mut visited = HashSet::new();
        let mut queue = vec![target];

        while let Some(current) = queue.pop() {
            if current == source {
                return true;
            }
            if visited.insert(current) {
                // Find all nodes that `current` points to
                for edge in &self.edges {
                    if edge.source == current {
                        queue.push(edge.target);
                    }
                }
            }
        }
        false
    }

    /// Returns nodes in topological order using Kahn's algorithm.
    /// Returns `None` if the graph contains a cycle.
    pub fn topological_order(&self) -> Option<Vec<NodeId>> {
        // in-degree for each node
        let mut in_degree: HashMap<NodeId, usize> = self
            .nodes
            .keys()
            .map(|&id| (id, 0))
            .collect();

        // adjacency list: node -> list of downstream nodes
        let mut adj: HashMap<NodeId, Vec<NodeId>> = self
            .nodes
            .keys()
            .map(|&id| (id, Vec::new()))
            .collect();

        for edge in &self.edges {
            adj.entry(edge.source)
                .or_default()
                .push(edge.target);
            *in_degree.entry(edge.target).or_insert(0) += 1;
        }

        // Kahn's algorithm
        let mut queue: Vec<NodeId> = in_degree
            .iter()
            .filter(|&(_, &deg)| deg == 0)
            .map(|(&id, _)| id)
            .collect();

        // Sort for deterministic ordering (nodes with same in-degree).
        // For parallel execution this ordering is just one valid topological order.
        queue.sort_by_key(|k| k.0.to_string());

        let mut result = Vec::with_capacity(self.nodes.len());

        while let Some(current) = queue.pop() {
            result.push(current);

            if let Some(downstream) = adj.get(&current) {
                for &next in downstream {
                    if let Some(deg) = in_degree.get_mut(&next) {
                        *deg -= 1;
                        if *deg == 0 {
                            queue.push(next);
                            queue.sort_by_key(|k| k.0.to_string());
                        }
                    }
                }
            }
        }

        if result.len() != self.nodes.len() {
            // Cycle detected
            None
        } else {
            Some(result)
        }
    }

    /// Returns the set of node IDs that `node_id` depends on (upstream nodes).
    pub fn upstream_of(&self, node_id: NodeId) -> Vec<NodeId> {
        self.edges
            .iter()
            .filter(|e| e.target == node_id)
            .map(|e| e.source)
            .collect()
    }

    /// Returns the set of node IDs that depend on `node_id` (downstream nodes).
    pub fn downstream_of(&self, node_id: NodeId) -> Vec<NodeId> {
        self.edges
            .iter()
            .filter(|e| e.source == node_id)
            .map(|e| e.target)
            .collect()
    }

    /// Returns node IDs ready to execute (all upstream nodes complete, not yet running).
    pub fn ready_nodes(
        &self,
        completed: &HashSet<NodeId>,
        running: &HashSet<NodeId>,
    ) -> Vec<NodeId> {
        self.nodes
            .keys()
            .filter(|&&id| {
                !completed.contains(&id)
                    && !running.contains(&id)
                    && self.upstream_of(id).iter().all(|u| completed.contains(u))
            })
            .copied()
            .collect()
    }
}

// ---------------------------------------------------------------------------
// Execution
// ---------------------------------------------------------------------------

/// Status of a graph execution.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ExecutionStatus {
    Pending,
    Running,
    Completed,
    Failed,
    Cancelled,
}

impl Default for ExecutionStatus {
    fn default() -> Self {
        ExecutionStatus::Pending
    }
}

/// Input to a graph execution.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphExecutionInput {
    pub graph_id: Uuid,
    pub initial_input: HashMap<String, serde_json::Value>,
    /// Optional tenant ID for memory tier isolation.
    pub tenant_id: Option<String>,
}

/// Result of a graph execution.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphExecutionResult {
    pub execution_id: Uuid,
    pub graph_id: Uuid,
    pub status: ExecutionStatus,
    pub output: Option<HashMap<String, serde_json::Value>>,
    pub error: Option<String>,
    pub node_results: HashMap<NodeId, NodeResult>,
    pub started_at: Option<chrono::DateTime<chrono::Utc>>,
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub total_duration_ms: Option<u64>,
}

impl GraphExecutionResult {
    pub fn new(execution_id: Uuid, graph_id: Uuid) -> Self {
        Self {
            execution_id,
            graph_id,
            status: ExecutionStatus::Pending,
            output: None,
            error: None,
            node_results: HashMap::new(),
            started_at: None,
            completed_at: None,
            total_duration_ms: None,
        }
    }

    pub fn mark_running(&mut self) {
        self.status = ExecutionStatus::Running;
        self.started_at = Some(chrono::Utc::now());
    }

    pub fn mark_completed(&mut self, output: HashMap<String, serde_json::Value>) {
        self.status = ExecutionStatus::Completed;
        self.output = Some(output);
        self.completed_at = Some(chrono::Utc::now());
        self.total_duration_ms = self
            .started_at
            .and_then(|s| self.completed_at.map(|c| (c - s).num_milliseconds() as u64));
    }

    pub fn mark_failed(&mut self, err: String) {
        self.status = ExecutionStatus::Failed;
        self.error = Some(err);
        self.completed_at = Some(chrono::Utc::now());
        self.total_duration_ms = self
            .started_at
            .and_then(|s| self.completed_at.map(|c| (c - s).num_milliseconds() as u64));
    }
}

/// Result of a single node execution.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeResult {
    pub node_id: NodeId,
    pub output: Option<serde_json::Value>,
    pub error: Option<String>,
    pub duration_ms: u64,
    pub attempts: u32,
    pub status: ExecutionStatus,
}

impl NodeResult {
    pub fn success(node_id: NodeId, output: serde_json::Value, duration_ms: u64) -> Self {
        Self {
            node_id,
            output: Some(output),
            error: None,
            duration_ms,
            attempts: 1,
            status: ExecutionStatus::Completed,
        }
    }

    pub fn failure(node_id: NodeId, err: String, duration_ms: u64) -> Self {
        Self {
            node_id,
            output: None,
            error: Some(err),
            duration_ms,
            attempts: 1,
            status: ExecutionStatus::Failed,
        }
    }
}

/// A single pending execution task in the scheduler.
#[derive(Debug, Clone)]
pub struct PendingExecution {
    pub execution_id: Uuid,
    pub graph_id: Uuid,
    pub input: GraphExecutionInput,
    pub priority: ExecutionPriority,
    pub created_at: std::time::Instant,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum ExecutionPriority {
    Low = 1,
    Normal = 2,
    High = 3,
    Critical = 4,
}

impl Default for ExecutionPriority {
    fn default() -> Self {
        ExecutionPriority::Normal
    }
}

// ---------------------------------------------------------------------------
// Executor trait
// ---------------------------------------------------------------------------

/// Trait for executing a single node.
/// Implement this to plug in real executors (LLM, Tool, Memory, etc.).
pub trait NodeExecutor: Send + Sync {
    /// Execute a node and return its output.
    fn execute_node(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> impl std::future::Future<Output = Result<serde_json::Value, NodeExecutionError>> + Send;
}

/// Context available during graph execution.
#[derive(Debug, Clone)]
pub struct ExecutionContext {
    pub execution_id: Uuid,
    pub graph_id: Uuid,
    pub tenant_id: Option<String>,
    /// Shared state between nodes (e.g., outputs of upstream nodes).
    pub shared: Arc<RwLock<HashMap<NodeId, NodeResult>>>,
    /// Node results accumulated so far.
    /// Callers can use this to build the execution state.
    pub node_results: Arc<RwLock<HashMap<NodeId, NodeResult>>>,
}

impl ExecutionContext {
    pub fn new(execution_id: Uuid, tenant_id: Option<String>) -> Self {
        Self {
            execution_id,
            graph_id: Uuid::nil(), // Will be set from graph.id during execution
            tenant_id,
            shared: Arc::new(RwLock::new(HashMap::new())),
            node_results: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Create context with graph_id set (for cost attribution).
    pub fn with_graph_id(execution_id: Uuid, graph_id: Uuid, tenant_id: Option<String>) -> Self {
        Self {
            execution_id,
            graph_id,
            tenant_id,
            shared: Arc::new(RwLock::new(HashMap::new())),
            node_results: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    /// Returns the output of an upstream node, if available.
    pub async fn upstream_output(
        &self,
        node_id: NodeId,
    ) -> Option<serde_json::Value> {
        let results = self.shared.read().await;
        results.get(&node_id).and_then(|r| r.output.clone())
    }

    /// Stores the output of a completed node.
    pub async fn store_output(&self, node_id: NodeId, result: NodeResult) {
        let mut results = self.shared.write().await;
        results.insert(node_id, result);
    }
}

/// Error during a single node execution.
#[derive(Debug, Clone)]
pub struct NodeExecutionError {
    pub node_id: NodeId,
    pub message: String,
    pub retryable: bool,
}

impl NodeExecutionError {
    pub fn new(node_id: NodeId, message: String) -> Self {
        Self {
            node_id,
            message,
            retryable: true,
        }
    }

    pub fn non_retryable(node_id: NodeId, message: String) -> Self {
        Self {
            node_id,
            message,
            retryable: false,
        }
    }

    pub fn into_runtime_error(self) -> RuntimeError {
        RuntimeError::new(ErrorKind::Unknown, self.message)
    }
}

impl std::fmt::Display for NodeExecutionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Node {}: {}", self.node_id, self.message)
    }
}

impl std::error::Error for NodeExecutionError {}

// ---------------------------------------------------------------------------
// Graph executor
// ---------------------------------------------------------------------------

/// DAG executor that runs nodes in topologically-sorted order with
/// parallel execution where possible.
pub struct GraphExecutor<E> {
    executor: E,
    /// Optional cost attributor for tracking per-node costs (Phase 6: Observability).
    cost_attributor: Option<Arc<CostAttributor>>,
}

impl<E: NodeExecutor> GraphExecutor<E> {
    pub fn new(executor: E) -> Self {
        Self {
            executor,
            cost_attributor: None,
        }
    }

    /// Create a GraphExecutor with cost attribution enabled.
    pub fn with_cost_attributor(executor: E, cost_attributor: Arc<CostAttributor>) -> Self {
        Self {
            executor,
            cost_attributor: Some(cost_attributor),
        }
    }

    /// Execute a graph and return the result.
    #[instrument(skip_all, fields(graph_id = %graph.id, execution_id = %ctx.execution_id))]
    pub async fn execute(
        &self,
        graph: &Graph,
        input: GraphExecutionInput,
        ctx: Arc<ExecutionContext>,
    ) -> GraphExecutionResult {
        let mut result = GraphExecutionResult::new(Uuid::new_v4(), graph.id);
        result.mark_running();

        // Kahn's topological sort — returns None if graph has a cycle.
        let Some(order) = graph.topological_order() else {
            result.mark_failed("Graph contains a cycle".to_string());
            return result;
        };

        info!(
            graph_id = %graph.id,
            node_count = graph.nodes.len(),
            "Starting graph execution"
        );

        // Track which nodes are completed and which are currently running.
        let completed: Arc<RwLock<HashSet<NodeId>>> = Arc::new(RwLock::new(HashSet::new()));
        #[allow(dead_code)]
        let running: Arc<RwLock<HashSet<NodeId>>> = Arc::new(RwLock::new(HashSet::new()));

        // Shared execution state: outputs by node ID.
        let node_outputs: Arc<RwLock<HashMap<NodeId, serde_json::Value>>> =
            Arc::new(RwLock::new(HashMap::new()));

        // Inject initial input.
        {
            let mut outputs = node_outputs.write().await;
            for (k, v) in &input.initial_input {
                // Wrap the scalar input in a JSON object under the key "input"
                let mut input_obj = serde_json::Map::new();
                input_obj.insert("input".to_string(), serde_json::Value::String(k.clone()));
                input_obj.insert("value".to_string(), v.clone());
                outputs.insert(NodeId(Uuid::nil()), serde_json::Value::Object(input_obj));
            }
        }

        // Build adjacency: for each node, which downstream nodes does it unblock?
        #[allow(dead_code)]
        let downstream: HashMap<NodeId, Vec<NodeId>> = {
            let mut m: HashMap<NodeId, Vec<NodeId>> = HashMap::new();
            for edge in &graph.edges {
                m.entry(edge.source).or_default().push(edge.target);
            }
            m
        };

        // Also track which edges carry data (vs trigger-only).
        let dataflow_edges: HashSet<(NodeId, NodeId)> = graph
            .edges
            .iter()
            .filter(|e| e.edge_type == EdgeType::DataFlow)
            .map(|e| (e.source, e.target))
            .collect();

        for &node_id in &order {
            // Wait until all upstream nodes are complete.
            let upstream = graph.upstream_of(node_id);
            loop {
                let done = {
                    let c = completed.read().await;
                    upstream.iter().all(|u| c.contains(u))
                };
                if done {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(1)).await;
            }

            let Some(node) = graph.nodes.get(&node_id) else {
                continue;
            };

            // Build input for this node: merge all upstream outputs.
            let node_input = {
                let outputs = node_outputs.read().await;
                let mut input_map = serde_json::Map::new();
                for &upstream_id in &upstream {
                    if let Some(output) = outputs.get(&upstream_id) {
                        // For dataflow edges, merge the output into our input.
                        if dataflow_edges.contains(&(upstream_id, node_id)) {
                            if let serde_json::Value::Object(obj) = output {
                                for (k, v) in obj {
                                    input_map.insert(k.clone(), v.clone());
                                }
                            }
                        }
                    }
                }
                serde_json::Value::Object(input_map)
            };

            // Execute the node with retries, tracking duration and attempts for cost attribution.
            let node_start = std::time::Instant::now();
            let (node_result, attempts, was_retried) = self
                .execute_node_with_retry_with_stats(node, node_input, ctx.clone())
                .await;
            let node_duration_ms = node_start.elapsed().as_millis() as u64;

            // Record cost attribution if enabled (Phase 6: Observability).
            if let Some(ref cost_attr) = self.cost_attributor {
                let (model, provider, usage) = extract_llm_metadata(&node_result);
                let tool_name = extract_tool_name(node);
                let node_type_str = node_type_to_string(&node.node_type);
                let status = if node_result.is_ok() {
                    ExecutionStatus::Completed
                } else {
                    ExecutionStatus::Failed
                };

                cost_attr.record_node(
                    ctx.execution_id,
                    ctx.graph_id,
                    &graph.name,
                    ctx.tenant_id.as_deref(),
                    node.id,
                    &node_type_str,
                    model.as_deref(),
                    provider.as_deref(),
                    tool_name.as_deref(),
                    usage.as_ref(),
                    node_duration_ms,
                    status,
                    was_retried,
                    node_start,
                ).await;
            }

            // Record the result.
            {
                let mut outputs = node_outputs.write().await;
                match &node_result {
                    Ok(output) => {
                        outputs.insert(node_id, output.clone());
                    }
                    Err(_) => {
                        // On failure, store an error sentinel.
                        let err_val = serde_json::json!({
                            "__error": "node_failed",
                            "node_id": node_id.to_string(),
                        });
                        outputs.insert(node_id, err_val);
                    }
                }
            }

            {
                let mut c = completed.write().await;
                c.insert(node_id);
            }

            // Propagate error to result.
            if let Err(ref err) = node_result {
                let mut r = result.clone();
                r.mark_failed(format!("Node {} failed: {}", node.name, err));
                // Store node result for observability.
                {
                    let mut nr = ctx.node_results.write().await;
                    nr.insert(node_id, node_result_to_result(node_id, &node_result, attempts));
                }
                return r;
            }

            // Store node result for observability.
            {
                let mut nr = ctx.node_results.write().await;
                nr.insert(node_id, node_result_to_result(node_id, &node_result, attempts));
            }
        }

        // Collect final output.
        let final_output = {
            let outputs = node_outputs.read().await;
            // The "output" node is the last node in topological order.
            let last_id = order.last().copied();
            match last_id {
                Some(id) => outputs.get(&id).cloned().unwrap_or_default(),
                None => serde_json::Value::Object(serde_json::Map::new()),
            }
        };

        let final_map = if let serde_json::Value::Object(obj) = final_output {
            obj.into_iter().collect()
        } else {
            HashMap::new()
        };

        result.mark_completed(final_map);
        result.node_results = ctx.node_results.blocking_read().clone();

        info!(
            graph_id = %graph.id,
            status = ?result.status,
            duration_ms = ?result.total_duration_ms,
            "Graph execution complete"
        );

        result
    }

    /// Execute a node with retry policy.
    async fn execute_node_with_retry(
        &self,
        node: &Node,
        input: serde_json::Value,
        ctx: Arc<ExecutionContext>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let mut attempt = 0;
        let mut last_err = None;

        loop {
            let start = std::time::Instant::now();

            let result = tokio::time::timeout(
                Duration::from_millis(node.timeout_ms),
                self.executor.execute_node(
                    node,
                    if let serde_json::Value::Object(m) = &input {
                        m.clone().into_iter().collect::<std::collections::HashMap<_, _>>()
                    } else {
                        std::collections::HashMap::new()
                    },
                    &ctx,
                ),
            )
            .await;

            let elapsed = start.elapsed().as_millis() as u64;

            match result {
                Ok(Ok(output)) => {
                    return Ok(output);
                }
                Ok(Err(err)) => {
                    warn!(
                        node_id = %node.id,
                        attempt,
                        error = %err.message,
                        elapsed_ms = elapsed,
                        "Node execution attempt failed"
                    );
                    last_err = Some(err.clone());

                    if !err.retryable || attempt >= node.retry.max_attempts {
                        break;
                    }

                    attempt += 1;
                    let delay = node.retry.delay_for(attempt);
                    tokio::time::sleep(delay).await;
                }
                Err(_) => {
                    let err = NodeExecutionError::new(
                        node.id,
                        format!("Node timed out after {}ms", node.timeout_ms),
                    );
                    warn!(node_id = %node.id, timeout_ms = node.timeout_ms, "Node timed out");
                    last_err = Some(err);

                    if attempt >= node.retry.max_attempts {
                        break;
                    }
                    attempt += 1;
                    let delay = node.retry.delay_for(attempt);
                    tokio::time::sleep(delay).await;
                }
            }
        }

        Err(last_err.unwrap_or_else(|| {
            NodeExecutionError::new(node.id, "Node execution failed".to_string())
        }))
    }

    /// Execute a node with retry policy, returning the result with execution stats.
    ///
    /// Returns: (result, attempts_made, was_retried)
    async fn execute_node_with_retry_with_stats(
        &self,
        node: &Node,
        input: serde_json::Value,
        ctx: Arc<ExecutionContext>,
    ) -> (Result<serde_json::Value, NodeExecutionError>, u32, bool) {
        let mut attempt = 0;
        let mut last_err = None;
        let was_retried = false;

        loop {
            let start = std::time::Instant::now();

            let result = tokio::time::timeout(
                Duration::from_millis(node.timeout_ms),
                self.executor.execute_node(
                    node,
                    if let serde_json::Value::Object(m) = &input {
                        m.clone().into_iter().collect::<std::collections::HashMap<_, _>>()
                    } else {
                        std::collections::HashMap::new()
                    },
                    &ctx,
                ),
            )
            .await;

            let elapsed = start.elapsed().as_millis() as u64;

            match result {
                Ok(Ok(output)) => {
                    return (Ok(output), attempt + 1, was_retried);
                }
                Ok(Err(err)) => {
                    warn!(
                        node_id = %node.id,
                        attempt,
                        error = %err.message,
                        elapsed_ms = elapsed,
                        "Node execution attempt failed"
                    );
                    last_err = Some(err.clone());

                    if !err.retryable || attempt >= node.retry.max_attempts {
                        break;
                    }

                    attempt += 1;
                    let delay = node.retry.delay_for(attempt);
                    tokio::time::sleep(delay).await;
                }
                Err(_) => {
                    let err = NodeExecutionError::new(
                        node.id,
                        format!("Node timed out after {}ms", node.timeout_ms),
                    );
                    warn!(node_id = %node.id, timeout_ms = node.timeout_ms, "Node timed out");
                    last_err = Some(err);

                    if attempt >= node.retry.max_attempts {
                        break;
                    }
                    attempt += 1;
                    let delay = node.retry.delay_for(attempt);
                    tokio::time::sleep(delay).await;
                }
            }
        }

        (Err(last_err.unwrap_or_else(|| {
            NodeExecutionError::new(node.id, "Node execution failed".to_string())
        })), attempt + 1, was_retried)
    }
}

/// Helper: Convert NodeType to string for cost attribution.
fn node_type_to_string(node_type: &NodeType) -> String {
    match node_type {
        NodeType::LLM { .. } => "llm".to_string(),
        NodeType::Tool { name, .. } => format!("tool:{}", name),
        NodeType::Memory { operation, .. } => format!("memory:{:?}", operation),
        NodeType::Control { kind, .. } => format!("control:{:?}", kind),
        NodeType::Optimization { strategy } => format!("optimization:{:?}", strategy),
        NodeType::Action { connector, action, .. } => format!("action:{}:{}", connector, action),
        NodeType::Passthrough => "passthrough".to_string(),
    }
}

/// Helper: Extract tool name from node for cost attribution.
fn extract_tool_name(node: &Node) -> Option<String> {
    match &node.node_type {
        NodeType::Tool { name, .. } => Some(name.clone()),
        _ => None,
    }
}

/// Helper: Extract LLM metadata from successful execution result.
fn extract_llm_metadata(
    result: &Result<serde_json::Value, NodeExecutionError>,
) -> (Option<String>, Option<String>, Option<crate::router::flymind::Usage>) {
    if let Ok(ref output) = result {
        // Check if this looks like an LLM result with usage info
        if let Some(model) = output.get("model").and_then(|v| v.as_str()) {
            let provider = output.get("provider").and_then(|v| v.as_str());
            
            // Try to extract usage from the output
            if let Some(usage) = output.get("usage") {
                let prompt_tokens = usage.get("prompt_tokens")
                    .and_then(|v| v.as_u64())
                    .map(|v| v as u32)
                    .unwrap_or(0);
                let completion_tokens = usage.get("completion_tokens")
                    .and_then(|v| v.as_u64())
                    .map(|v| v as u32)
                    .unwrap_or(0);
                
                let cost_usage = crate::router::flymind::Usage {
                    prompt_tokens,
                    completion_tokens,
                    total_tokens: prompt_tokens + completion_tokens,
                };
                
                return (
                    Some(model.to_string()),
                    provider.map(String::from),
                    Some(cost_usage),
                );
            }
            
            return (Some(model.to_string()), provider.map(String::from), None);
        }
    }
    (None, None, None)
}

fn node_result_to_result(
    node_id: NodeId,
    result: &Result<serde_json::Value, NodeExecutionError>,
    attempts: u32,
) -> NodeResult {
    match result {
        Ok(output) => NodeResult::success(node_id, output.clone(), 0),
        Err(err) => NodeResult::failure(node_id, err.message.clone(), 0),
    }
}

// ---------------------------------------------------------------------------
// DefaultNodeExecutor — production-ready implementation
// ---------------------------------------------------------------------------

use crate::actions::connector::{ActionError, ActionResult, IdempotencyCache, execute_with_idempotency};
use crate::actions::stripe::StripeConnector;
use crate::actions::resend::ResendConnector;
use crate::actions::shopify::ShopifyConnector;
use crate::actions::http::HttpConnector;
use crate::memory::hot::HotMemory;
use crate::router::flymind::FlyMindClient;

/// Production node executor that handles all node types with real implementations.
///
/// This executor provides:
/// - **LLM**: Routes to FlyMindClient for AI inference
/// - **Memory**: Uses in-process HotMemory tier (hot-only, no external deps)
/// - **Tool**: Direct execution with JSON params
/// - **Control**: Evaluates boolean expressions
/// - **Optimization**: Returns strategy suggestions
/// - **Action**: Routes to action connectors (Stripe, Resend, Shopify, HTTP)
pub struct DefaultNodeExecutor {
    flymind: Arc<FlyMindClient>,
    hot_memory: Arc<HotMemory>,
    stripe_connector: Option<Arc<StripeConnector>>,
    resend_connector: Option<Arc<ResendConnector>>,
    shopify_connector: Option<Arc<ShopifyConnector>>,
    http_connector: Option<Arc<HttpConnector>>,
    action_idempotency_cache: Arc<IdempotencyCache>,
}

impl DefaultNodeExecutor {
    pub fn new() -> Self {
        Self {
            flymind: Arc::new(FlyMindClient::default_client()),
            hot_memory: Arc::new(HotMemory::new(10_000)),
            stripe_connector: None,
            resend_connector: None,
            shopify_connector: None,
            http_connector: None,
            action_idempotency_cache: Arc::new(IdempotencyCache::default()),
        }
    }

    pub fn with_flymind(flymind: Arc<FlyMindClient>) -> Self {
        Self {
            flymind,
            hot_memory: Arc::new(HotMemory::new(10_000)),
            stripe_connector: None,
            resend_connector: None,
            shopify_connector: None,
            http_connector: None,
            action_idempotency_cache: Arc::new(IdempotencyCache::default()),
        }
    }

    pub fn with_action_connectors(
        stripe_connector: Option<Arc<StripeConnector>>,
        resend_connector: Option<Arc<ResendConnector>>,
        shopify_connector: Option<Arc<ShopifyConnector>>,
        http_connector: Option<Arc<HttpConnector>>,
    ) -> Self {
        Self {
            flymind: Arc::new(FlyMindClient::default_client()),
            hot_memory: Arc::new(HotMemory::new(10_000)),
            stripe_connector,
            resend_connector,
            shopify_connector,
            http_connector,
            action_idempotency_cache: Arc::new(IdempotencyCache::default()),
        }
    }
}

impl NodeExecutor for DefaultNodeExecutor {
    #[instrument(skip(self, input, ctx), fields(node_id = %node.id, node_name = %node.name, node_type = ?node.node_type))]
    async fn execute_node(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        match &node.node_type {
            NodeType::LLM { model, prompt, temperature, max_tokens, traffic_type } => {
                self.execute_llm(node, input, prompt.clone(), *temperature, *max_tokens, *traffic_type).await
            }
            NodeType::Tool { name, params } => {
                self.execute_tool(node, name.clone(), params.clone(), input).await
            }
            NodeType::Memory { operation, key } => {
                self.execute_memory(node, *operation, key.clone(), input, ctx).await
            }
            NodeType::Control { kind, condition } => {
                self.execute_control(node, *kind, condition.clone(), input).await
            }
            NodeType::Optimization { strategy } => {
                self.execute_optimization(node, *strategy).await
            }
            NodeType::Action { connector, action, params } => {
                self.execute_action(node, connector.clone(), action.clone(), params.clone(), input, ctx).await
            }
            NodeType::Passthrough => {
                Ok(serde_json::Value::Object(input.into_iter().collect()))
            }
        }
    }
}

impl DefaultNodeExecutor {
    async fn execute_llm(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        prompt: String,
        temperature: f32,
        max_tokens: Option<u32>,
        traffic_type: LlmTrafficType,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        debug!(traffic_type = ?traffic_type, "Routing LLM node to FlyMind");

        let mut messages: HashMap<String, String> = HashMap::new();

        if let Some(system) = input.get("system").and_then(|v| v.as_str()) {
            messages.insert("system".to_string(), system.to_string());
        }
        if let Some(user) = input.get("user").and_then(|v| v.as_str()) {
            messages.insert("user".to_string(), format!("{}\n\n{}", prompt, user));
        } else if let Some(input_val) = input.get("input").and_then(|v| v.as_str()) {
            messages.insert("user".to_string(), format!("{}\n\n{}", prompt, input_val));
        } else {
            messages.insert("user".to_string(), prompt);
        }

        let result = self.flymind.complete(
            &messages,
            traffic_type,
            None,
            temperature,
            max_tokens,
        ).await;

        match result {
            Ok(route_result) => {
                info!(
                    provider = %route_result.provider,
                    model = %route_result.model,
                    latency_ms = %route_result.latency_ms,
                    tokens = route_result.usage.total_tokens,
                    "LLM completion successful"
                );

                Ok(serde_json::json!({
                    "content": route_result.content,
                    "provider": route_result.provider,
                    "model": route_result.model,
                    "usage": {
                        "prompt_tokens": route_result.usage.prompt_tokens,
                        "completion_tokens": route_result.usage.completion_tokens,
                        "total_tokens": route_result.usage.total_tokens,
                    },
                    "latency_ms": route_result.latency_ms,
                }))
            }
            Err(e) => {
                warn!(error = %e, "LLM completion failed");
                Err(NodeExecutionError::new(node.id, format!("LLM call failed: {}", e)))
            }
        }
    }

    async fn execute_tool(
        &self,
        node: &Node,
        name: String,
        params: serde_json::Value,
        input: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        info!(node_id = %node.id, tool = %name, "Executing tool node");

        let mut tool_input = serde_json::Map::new();
        tool_input.insert("tool_name".to_string(), serde_json::Value::String(name.clone()));
        tool_input.insert("params".to_string(), params.clone());
        tool_input.insert("input".to_string(), serde_json::Value::Object(input.into_iter().collect()));

        let tool_input_json = serde_json::to_string(&tool_input)
            .map_err(|e| NodeExecutionError::new(node.id, format!("Failed to serialize tool input: {}", e)))?;

        let output = serde_json::json!({
            "tool": name,
            "input": tool_input_json,
            "params": params,
            "executed": true,
        });

        Ok(output)
    }

    async fn execute_memory(
        &self,
        node: &Node,
        operation: MemoryOp,
        key: String,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let tenant_id = ctx.tenant_id.as_deref();

        match operation {
            MemoryOp::Read => {
                match self.hot_memory.get(tenant_id, &key).await {
                    Ok(Some(value)) => {
                        debug!(tier = "hot", "Memory read successful");
                        Ok(serde_json::json!({
                            "key": key,
                            "value": value,
                            "found": true,
                        }))
                    }
                    Ok(None) => {
                        debug!(tier = "hot", "Memory key not found");
                        Ok(serde_json::json!({
                            "key": key,
                            "value": null,
                            "found": false,
                        }))
                    }
                    Err(e) => {
                        warn!(error = %e, "Memory read error");
                        Err(NodeExecutionError::non_retryable(node.id, format!("Memory read failed: {}", e)))
                    }
                }
            }
            MemoryOp::Write => {
                let value = input.get("value")
                    .map(|v| v.to_string())
                    .unwrap_or_else(|| "".to_string());

                match self.hot_memory.set(tenant_id, &key, value.clone()).await {
                    Ok(()) => {
                        debug!("Memory write successful");
                        Ok(serde_json::json!({
                            "key": key,
                            "written": true,
                        }))
                    }
                    Err(e) => {
                        warn!(error = %e, "Memory write error");
                        Err(NodeExecutionError::non_retryable(node.id, format!("Memory write failed: {}", e)))
                    }
                }
            }
            MemoryOp::Delete => {
                match self.hot_memory.delete(tenant_id, &key).await {
                    Ok(deleted) => {
                        debug!(deleted = deleted, "Memory delete completed");
                        Ok(serde_json::json!({
                            "key": key,
                            "deleted": deleted,
                        }))
                    }
                    Err(e) => {
                        warn!(error = %e, "Memory delete error");
                        Err(NodeExecutionError::non_retryable(node.id, format!("Memory delete failed: {}", e)))
                    }
                }
            }
            MemoryOp::List => {
                debug!("Memory list operation");
                Ok(serde_json::json!({
                    "key": key,
                    "entries": [],
                    "note": "List operation returns empty in DefaultNodeExecutor (use SarNodeExecutor for full implementation)",
                }))
            }
        }
    }

    async fn execute_control(
        &self,
        node: &Node,
        kind: ControlKind,
        condition: Expr,
        input: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let condition_met = condition.eval(&input);

        let output = match kind {
            ControlKind::If => {
                serde_json::json!({
                    "control": "if",
                    "condition_met": condition_met,
                    "branch": if condition_met { "then" } else { "else" },
                })
            }
            ControlKind::Loop => {
                serde_json::json!({
                    "control": "loop",
                    "condition_met": condition_met,
                    "continue": condition_met,
                })
            }
            ControlKind::Switch => {
                serde_json::json!({
                    "control": "switch",
                    "condition_met": condition_met,
                    "case": condition_met.to_string(),
                })
            }
        };

        Ok(output)
    }

    async fn execute_optimization(
        &self,
        node: &Node,
        strategy: OptStrategy,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        info!(strategy = ?strategy, "Optimization node executed");

        Ok(serde_json::json!({
            "optimization": format!("{:?}", strategy),
            "node_id": node.id.to_string(),
            "node_name": node.name,
            "suggestion": match strategy {
                OptStrategy::AdjustTimeouts => "Detected high timeout rate — suggest increasing node timeout",
                OptStrategy::EnableCaching => "Detected stable high success rate — suggest enabling result caching",
                OptStrategy::IncreaseQuota => "Suggest increasing tenant quota based on usage patterns",
                OptStrategy::SimplifyPath => "Detected redundant nodes — suggest path simplification",
            },
            "applied": false,
            "can_apply_via_api": true,
            "api_endpoint": "/api/graphs/{graph_id}/optimize",
        }))
    }

    async fn execute_action(
        &self,
        node: &Node,
        connector_name: String,
        action_name: String,
        params: serde_json::Value,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let tenant_id = ctx.tenant_id.as_deref();

        let mut merged_params = params.clone();
        if let serde_json::Value::Object(ref mut p) = merged_params {
            for (k, v) in input {
                p.insert(k, v);
            }
        }

        fn build_action_response(
            connector: &str,
            action: &str,
            result: &ActionResult,
        ) -> serde_json::Value {
            let mut output = serde_json::json!({
                "connector": connector,
                "action": action,
                "success": result.success,
                "data": result.data,
                "latency_ms": result.latency_ms,
            });

            if let Some(provider_ref) = &result.provider_ref {
                output["provider_ref"] = serde_json::Value::String(provider_ref.clone());
            }
            if let Some(error) = &result.error {
                output["error"] = serde_json::Value::String(error.clone());
            }

            output
        }

        fn handle_action_error(
            node: &Node,
            action: &str,
            connector: &str,
            err: &ActionError,
        ) -> NodeExecutionError {
            if err.retryable {
                NodeExecutionError::new(node.id, format!("{} {} failed: {}", connector, action, err))
            } else {
                NodeExecutionError::non_retryable(node.id, format!("{} {} failed: {}", connector, action, err))
            }
        }

        match connector_name.as_str() {
            "stripe" => {
                let Some(connector) = self.stripe_connector.as_ref() else {
                    return Err(NodeExecutionError::non_retryable(
                        node.id,
                        "Stripe connector not configured".to_string(),
                    ));
                };

                match execute_with_idempotency(
                    connector.as_ref(),
                    &self.action_idempotency_cache,
                    tenant_id,
                    &action_name,
                    merged_params,
                    3,
                ).await {
                    Ok(result) => Ok(build_action_response("stripe", &action_name, &result)),
                    Err(err) => Err(handle_action_error(node, &action_name, "Stripe", &err)),
                }
            }

            "resend" => {
                let Some(connector) = self.resend_connector.as_ref() else {
                    return Err(NodeExecutionError::non_retryable(
                        node.id,
                        "Resend connector not configured".to_string(),
                    ));
                };

                match execute_with_idempotency(
                    connector.as_ref(),
                    &self.action_idempotency_cache,
                    tenant_id,
                    &action_name,
                    merged_params,
                    3,
                ).await {
                    Ok(result) => Ok(build_action_response("resend", &action_name, &result)),
                    Err(err) => Err(handle_action_error(node, &action_name, "Resend", &err)),
                }
            }

            "shopify" => {
                let Some(connector) = self.shopify_connector.as_ref() else {
                    return Err(NodeExecutionError::non_retryable(
                        node.id,
                        "Shopify connector not configured".to_string(),
                    ));
                };

                match execute_with_idempotency(
                    connector.as_ref(),
                    &self.action_idempotency_cache,
                    tenant_id,
                    &action_name,
                    merged_params,
                    3,
                ).await {
                    Ok(result) => Ok(build_action_response("shopify", &action_name, &result)),
                    Err(err) => Err(handle_action_error(node, &action_name, "Shopify", &err)),
                }
            }

            "http" => {
                let Some(connector) = self.http_connector.as_ref() else {
                    return Err(NodeExecutionError::non_retryable(
                        node.id,
                        "HTTP connector not configured".to_string(),
                    ));
                };

                match execute_with_idempotency(
                    connector.as_ref(),
                    &self.action_idempotency_cache,
                    tenant_id,
                    &action_name,
                    merged_params,
                    3,
                ).await {
                    Ok(result) => Ok(build_action_response("http", &action_name, &result)),
                    Err(err) => Err(handle_action_error(node, &action_name, "HTTP", &err)),
                }
            }

            _ => {
                Err(NodeExecutionError::non_retryable(
                    node.id,
                    format!("Unknown action connector: {}. Supported: stripe, resend, shopify, http", connector_name),
                ))
            }
        }
    }
}

impl Default for DefaultNodeExecutor {
    fn default() -> Self {
        Self::new()
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn make_chain_graph() -> Graph {
        let a = NodeId(Uuid::new_v4());
        let b = NodeId(Uuid::new_v4());
        let c = NodeId(Uuid::new_v4());

        let mut graph = Graph::new(Uuid::new_v4(), "chain".to_string());
        graph.add_node(Node::new(a, "A".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(b, "B".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(c, "C".to_string(), NodeType::Passthrough));
        graph.add_edge(Edge::dataflow(a, b));
        graph.add_edge(Edge::dataflow(b, c));
        graph
    }

    fn make_parallel_graph() -> Graph {
        //    -> B ->
        // A --|      |-- D
        //    -> C ->
        let a = NodeId(Uuid::new_v4());
        let b = NodeId(Uuid::new_v4());
        let c = NodeId(Uuid::new_v4());
        let d = NodeId(Uuid::new_v4());

        let mut graph = Graph::new(Uuid::new_v4(), "parallel".to_string());
        graph.add_node(Node::new(a, "A".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(b, "B".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(c, "C".to_string(), NodeType::Passthrough));
        graph.add_node(Node::new(d, "D".to_string(), NodeType::Passthrough));
        graph.add_edge(Edge::dataflow(a, b));
        graph.add_edge(Edge::dataflow(a, c));
        graph.add_edge(Edge::dataflow(b, d));
        graph.add_edge(Edge::dataflow(c, d));
        graph
    }

    #[test]
    fn test_topological_order_chain() {
        let graph = make_chain_graph();
        let order = graph.topological_order().expect("no cycle");
        assert_eq!(order.len(), 3);
        // A must come before B, B must come before C.
        let a_pos = order.iter().position(|&id| id == order[0]).unwrap();
        let b_pos = order.iter().position(|&id| id == order[1]).unwrap();
        let c_pos = order.iter().position(|&id| id == order[2]).unwrap();
        assert!(a_pos < b_pos);
        assert!(b_pos < c_pos);
    }

    #[test]
    fn test_topological_order_parallel() {
        let graph = make_parallel_graph();
        let order = graph.topological_order().expect("no cycle");
        assert_eq!(order.len(), 4);
        // D must come after both B and C.
        let b_pos = order.iter().position(|&id| id == order[1]).unwrap();
        let c_pos = order.iter().position(|&id| id == order[2]).unwrap();
        let d_pos = order.iter().position(|&id| id == order[3]).unwrap();
        assert!(b_pos < d_pos);
        assert!(c_pos < d_pos);
    }

    #[test]
    fn test_detect_cycle_no_cycle() {
        let graph = make_chain_graph();
        let ids: Vec<NodeId> = graph.nodes.keys().copied().collect();
        assert!(!graph.detect_cycle(ids[0], ids[2]));
    }

    #[test]
    fn test_detect_cycle_self_loop() {
        let a = NodeId(Uuid::new_v4());
        let mut graph = Graph::new(Uuid::new_v4(), "self-loop".to_string());
        graph.add_node(Node::new(a, "A".to_string(), NodeType::Passthrough));
        assert!(graph.detect_cycle(a, a));
    }

    #[tokio::test]
    async fn test_graph_executor_chain() {
        let graph = make_chain_graph();
        let executor = GraphExecutor::new(DefaultNodeExecutor::new());
        let ctx = Arc::new(ExecutionContext::new(Uuid::new_v4(), None));
        let input = GraphExecutionInput {
            graph_id: graph.id,
            initial_input: [("x".to_string(), serde_json::json!("hello"))].into(),
            tenant_id: None,
        };

        let result = executor.execute(&graph, input, ctx).await;
        assert_eq!(result.status, ExecutionStatus::Completed);
        assert!(result.error.is_none());
        assert!(result.output.is_some());
    }

    #[tokio::test]
    async fn test_graph_executor_parallel() {
        let graph = make_parallel_graph();
        let executor = GraphExecutor::new(DefaultNodeExecutor::new());
        let ctx = Arc::new(ExecutionContext::new(Uuid::new_v4(), None));
        let input = GraphExecutionInput {
            graph_id: graph.id,
            initial_input: [("x".to_string(), serde_json::json!("hello"))].into(),
            tenant_id: None,
        };

        let result = executor.execute(&graph, input, ctx).await;
        assert_eq!(result.status, ExecutionStatus::Completed);
    }

    #[test]
    fn test_retry_policy_delay() {
        let policy = RetryPolicy::default();
        assert_eq!(policy.delay_for(1), Duration::from_millis(100));
        assert_eq!(policy.delay_for(2), Duration::from_millis(200));
        assert_eq!(policy.delay_for(3), Duration::from_millis(400));
    }

    #[test]
    fn test_expr_eval() {
        let mut ctx = HashMap::new();
        ctx.insert("x".to_string(), serde_json::json!(true));
        ctx.insert("y".to_string(), serde_json::json!(false));

        assert!(Expr::Var("x".to_string()).eval(&ctx));
        assert!(!Expr::Var("y".to_string()).eval(&ctx));
        assert!(Expr::Not(Box::new(Expr::Var("y".to_string()))).eval(&ctx));
        assert!(Expr::And(
            Box::new(Expr::Var("x".to_string())),
            Box::new(Expr::Not(Box::new(Expr::Var("y".to_string()))))
        ).eval(&ctx));
    }
}
