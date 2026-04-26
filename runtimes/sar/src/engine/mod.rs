use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use std::time::Duration;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tracing::{info, warn, instrument};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct NodeId(pub Uuid);

impl std::fmt::Display for NodeId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    pub id: NodeId,
    pub name: String,
    pub node_type: NodeType,
    pub timeout_ms: u64,
    pub retry: RetryPolicy,
    pub input_schema: Option<String>,
    pub output_schema: Option<String>,
    pub metadata: HashMap<String, String>,
}

impl Node {
    pub fn new(id: NodeId, name: String, node_type: NodeType) -> Node {
        Node {
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

    pub fn with_timeout(mut self, timeout_ms: u64) -> Node {
        self.timeout_ms = timeout_ms;
        self
    }

    pub fn with_retry(mut self, retry: RetryPolicy) -> Node {
        self.retry = retry;
        self
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum NodeType {
    LLM {
        model: Option<String>,
        prompt: String,
        temperature: f32,
        max_tokens: Option<u32>,
        traffic_type: LlmTrafficType,
    },
    Tool {
        name: String,
        params: serde_json::Value,
    },
    Memory {
        operation: MemoryOp,
        key: String,
    },
    Control {
        kind: ControlKind,
        condition: Expr,
    },
    Optimization {
        strategy: OptStrategy,
    },
    Action {
        connector: String,
        action: String,
        params: serde_json::Value,
    },
    Passthrough,
}

impl NodeType {
    pub fn llm(prompt: String) -> NodeType {
        NodeType::LLM {
            model: None,
            prompt,
            temperature: 0.7,
            max_tokens: None,
            traffic_type: LlmTrafficType::General,
        }
    }

    pub fn tool(name: String, params: serde_json::Value) -> NodeType {
        NodeType::Tool { name, params }
    }

    pub fn memory(operation: MemoryOp, key: String) -> NodeType {
        NodeType::Memory { operation, key }
    }

    pub fn control_if(condition: Expr) -> NodeType {
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
            Expr::Var(name) => ctx.get(name).and_then(|v| v.as_bool()).unwrap_or(false),
        }
    }
}

impl ExprValue {
    pub fn eval(&self, ctx: &HashMap<String, serde_json::Value>) -> serde_json::Value {
        match self {
            ExprValue::Var(name) => ctx.get(name).cloned().unwrap_or(serde_json::Value::Null),
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetryPolicy {
    pub max_attempts: u32,
    pub initial_delay_ms: u64,
    pub max_delay_ms: u64,
    pub backoff_multiplier: f64,
}

impl Default for RetryPolicy {
    fn default() -> Self {
        Self {
            max_attempts: 3,
            initial_delay_ms: 100,
            max_delay_ms: 10_000,
            backoff_multiplier: Edge::default_backoff(),
        }
    }
}

impl RetryPolicy {
    pub fn no_retries() -> Self {
        Self { max_attempts: 0, ..Default::default() }
    }

    pub fn delay_for(&self, attempt: u32) -> Duration {
        let delay = self.initial_delay_ms as f64
            * self.backoff_multiplier.powi(attempt as i32 - 1);
        let delay = delay.min(self.max_delay_ms as f64);
        Duration::from_millis(delay as u64)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Edge {
    pub id: Uuid,
    pub source: NodeId,
    pub target: NodeId,
    pub edge_type: EdgeType,
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

    fn default_backoff() -> f64 {
        2.0
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum EdgeType {
    DataFlow,
    Trigger,
    Dependency,
}

impl Default for EdgeType {
    fn default() -> Self {
        EdgeType::DataFlow
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
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

    pub fn detect_cycle(&self, source: NodeId, target: NodeId) -> bool {
        let mut visited = HashSet::new();
        let mut queue = vec![target];
        while let Some(current) = queue.pop() {
            if current == source {
                return true;
            }
            if visited.insert(current) {
                for edge in &self.edges {
                    if edge.source == current {
                        queue.push(edge.target);
                    }
                }
            }
        }
        false
    }

    pub fn topological_order(&self) -> Option<Vec<NodeId>> {
        let mut in_degree: HashMap<NodeId, usize> = self.nodes.keys().map(|&id| (id, 0)).collect();
        let mut adj: HashMap<NodeId, Vec<NodeId>> = self.nodes.keys().map(|&id| (id, Vec::new())).collect();

        for edge in &self.edges {
            adj.entry(edge.source).or_default().push(edge.target);
            *in_degree.entry(edge.target).or_insert(0) += 1;
        }

        let mut queue: Vec<NodeId> = in_degree.iter()
            .filter(|&(_, &deg)| deg == 0)
            .map(|(&id, _)| id)
            .collect();
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

        if result.len() != self.nodes.len() { None } else { Some(result) }
    }

    pub fn upstream_of(&self, node_id: NodeId) -> Vec<NodeId> {
        self.edges.iter().filter(|e| e.target == node_id).map(|e| e.source).collect()
    }

    pub fn downstream_of(&self, node_id: NodeId) -> Vec<NodeId> {
        self.edges.iter().filter(|e| e.source == node_id).map(|e| e.target).collect()
    }

    pub fn ready_nodes(&self, completed: &HashSet<NodeId>, running: &HashSet<NodeId>) -> Vec<NodeId> {
        self.nodes.keys()
            .filter(|&&id| {
                !completed.contains(&id)
                    && !running.contains(&id)
                    && self.upstream_of(id).iter().all(|u| completed.contains(u))
            })
            .copied()
            .collect()
    }
}

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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphExecutionInput {
    pub graph_id: Uuid,
    pub initial_input: HashMap<String, serde_json::Value>,
    pub tenant_id: Option<String>,
}

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
        self.total_duration_ms = self.started_at
            .and_then(|s| self.completed_at.map(|c| (c - s).num_milliseconds() as u64));
    }

    pub fn mark_failed(&mut self, err: String) {
        self.status = ExecutionStatus::Failed;
        self.error = Some(err);
        self.completed_at = Some(chrono::Utc::now());
        self.total_duration_ms = self.started_at
            .and_then(|s| self.completed_at.map(|c| (c - s).num_milliseconds() as u64));
    }
}

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

#[derive(Debug, Clone)]
pub struct NodeExecutionError {
    pub node_id: NodeId,
    pub message: String,
    pub retryable: bool,
}

impl NodeExecutionError {
    pub fn new(node_id: NodeId, message: String) -> Self {
        Self { node_id, message, retryable: true }
    }

    pub fn non_retryable(node_id: NodeId, message: String) -> Self {
        Self { node_id, message, retryable: false }
    }
}

impl std::fmt::Display for NodeExecutionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "Node {}: {}", self.node_id, self.message)
    }
}

impl std::error::Error for NodeExecutionError {}

#[derive(Debug, Clone)]
pub struct ExecutionContext {
    pub execution_id: Uuid,
    pub graph_id: Uuid,
    pub tenant_id: Option<String>,
    pub shared: Arc<RwLock<HashMap<NodeId, NodeResult>>>,
}

impl ExecutionContext {
    pub fn new(execution_id: Uuid, tenant_id: Option<String>) -> Self {
        Self {
            execution_id,
            graph_id: Uuid::nil(),
            tenant_id,
            shared: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn upstream_output(&self, node_id: NodeId) -> Option<serde_json::Value> {
        let results = self.shared.read();
        results.get(&node_id).and_then(|r| r.output.clone())
    }

    pub async fn store_output(&self, node_id: NodeId, result: NodeResult) {
        let mut results = self.shared.write();
        results.insert(node_id, result);
    }
}

pub trait NodeExecutor: Send + Sync {
    fn execute_node(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        ctx: &ExecutionContext,
    ) -> impl std::future::Future<Output = Result<serde_json::Value, NodeExecutionError>> + Send;
}

pub struct GraphExecutor<E> {
    executor: E,
}

impl<E: NodeExecutor> GraphExecutor<E> {
    pub fn new(executor: E) -> Self {
        Self { executor }
    }

    #[instrument(skip_all, fields(graph_id = %graph.id, execution_id = %input.graph_id))]
    pub async fn execute(
        &self,
        graph: &Graph,
        input: GraphExecutionInput,
        ctx: Arc<ExecutionContext>,
    ) -> GraphExecutionResult {
        let mut result = GraphExecutionResult::new(input.graph_id, graph.id);
        result.mark_running();

        let Some(order) = graph.topological_order() else {
            result.mark_failed("Graph contains a cycle".to_string());
            return result;
        };

        info!(graph_id = %graph.id, node_count = graph.nodes.len(), "Starting graph execution");

        let completed: Arc<RwLock<HashSet<NodeId>>> = Arc::new(RwLock::new(HashSet::new()));
        let node_outputs: Arc<RwLock<HashMap<NodeId, serde_json::Value>>> =
            Arc::new(RwLock::new(HashMap::new()));

        {
            let mut outputs = node_outputs.write();
            for (k, v) in &input.initial_input {
                let mut input_obj = serde_json::Map::new();
                input_obj.insert("input".to_string(), serde_json::Value::String(k.clone()));
                input_obj.insert("value".to_string(), v.clone());
                outputs.insert(NodeId(Uuid::nil()), serde_json::Value::Object(input_obj));
            }
        }

        let dataflow_edges: HashSet<(NodeId, NodeId)> = graph.edges.iter()
            .filter(|e| e.edge_type == EdgeType::DataFlow)
            .map(|e| (e.source, e.target))
            .collect();

        for &node_id in &order {
            let upstream = graph.upstream_of(node_id);
            loop {
                let done = { upstream.iter().all(|u| completed.read().contains(u)) };
                if done { break; }
                tokio::time::sleep(Duration::from_millis(1)).await;
            }

            let Some(node) = graph.nodes.get(&node_id) else { continue; };

            let node_input = {
                let outputs = node_outputs.read();
                let mut input_map = serde_json::Map::new();
                for &upstream_id in &upstream {
                    if let Some(output) = outputs.get(&upstream_id) {
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

            let node_start = std::time::Instant::now();
            let node_result = self.execute_with_retry(node, node_input, &ctx).await;
            let node_duration_ms = node_start.elapsed().as_millis() as u64;

            {
                let mut outputs = node_outputs.write();
                match &node_result {
                    Ok(output) => { outputs.insert(node_id, output.clone()); }
                    Err(_) => {
                        outputs.insert(node_id, serde_json::json!({
                            "__error": "node_failed",
                            "node_id": node_id.to_string(),
                        }));
                    }
                }
            }

            {
                completed.write().insert(node_id);
            }

            {
                let mut nr = ctx.shared.write();
                nr.insert(node_id, NodeResult {
                    node_id,
                    output: node_result.as_ref().ok().cloned(),
                    error: node_result.as_ref().err().map(|e| e.message.clone()),
                    duration_ms: node_duration_ms,
                    attempts: 1,
                    status: if node_result.is_ok() { ExecutionStatus::Completed } else { ExecutionStatus::Failed },
                });
            }

            if let Err(ref err) = node_result {
                result.mark_failed(format!("Node {} failed: {}", node.name, err));
                result.node_results = ctx.shared.read().clone();
                return result;
            }
        }

        let final_output = {
            let outputs = node_outputs.read();
            order.last().and_then(|id| outputs.get(id)).cloned()
                .unwrap_or(serde_json::Value::Object(serde_json::Map::new()))
        };

        let final_map = match final_output {
            serde_json::Value::Object(obj) => obj.into_iter().collect(),
            _ => HashMap::new(),
        };

        result.mark_completed(final_map);
        result.node_results = ctx.shared.read().clone();

        info!(graph_id = %graph.id, status = ?result.status, "Graph execution complete");
        result
    }

    async fn execute_with_retry(
        &self,
        node: &Node,
        input: serde_json::Value,
        ctx: &Arc<ExecutionContext>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let mut attempt = 0;
        let mut last_err = None;

        loop {
            let node_input = match &input {
                serde_json::Value::Object(m) => m.clone().into_iter().collect(),
                _ => HashMap::new(),
            };

            let result = tokio::time::timeout(
                Duration::from_millis(node.timeout_ms),
                self.executor.execute_node(node, node_input, ctx),
            ).await;

            match result {
                Ok(Ok(output)) => return Ok(output),
                Ok(Err(err)) => {
                    warn!(node_id = %node.id, attempt, error = %err.message, "Node attempt failed");
                    last_err = Some(err.clone());
                    if !err.retryable || attempt >= node.retry.max_attempts { break; }
                    attempt += 1;
                    tokio::time::sleep(node.retry.delay_for(attempt)).await;
                }
                Err(_) => {
                    let err = NodeExecutionError::new(node.id, format!("Timed out after {}ms", node.timeout_ms));
                    warn!(node_id = %node.id, timeout_ms = node.timeout_ms, "Node timed out");
                    last_err = Some(err);
                    if attempt >= node.retry.max_attempts { break; }
                    attempt += 1;
                    tokio::time::sleep(node.retry.delay_for(attempt)).await;
                }
            }
        }

        Err(last_err.unwrap_or_else(|| NodeExecutionError::new(node.id, "Node execution failed".to_string())))
    }
}

pub struct DefaultNodeExecutor;

impl NodeExecutor for DefaultNodeExecutor {
    async fn execute_node(
        &self,
        node: &Node,
        input: HashMap<String, serde_json::Value>,
        _ctx: &ExecutionContext,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        match &node.node_type {
            NodeType::LLM { prompt, .. } => {
                Ok(serde_json::json!({
                    "response": format!("[LLM stub] prompt: {}", prompt),
                    "model": "stub",
                }))
            }
            NodeType::Tool { name, params } => {
                Ok(serde_json::json!({
                    "tool": name,
                    "result": format!("[Tool stub] {}", name),
                    "params": params,
                }))
            }
            NodeType::Memory { operation, key } => {
                Ok(serde_json::json!({
                    "memory_operation": format!("{:?}", operation),
                    "key": key,
                }))
            }
            NodeType::Control { kind, .. } => {
                Ok(serde_json::json!({
                    "control": format!("{:?}", kind),
                    "condition_met": true,
                }))
            }
            NodeType::Optimization { strategy } => {
                Ok(serde_json::json!({
                    "optimization": format!("{:?}", strategy),
                    "suggestions": [],
                }))
            }
            NodeType::Action { connector, action, params } => {
                Ok(serde_json::json!({
                    "connector": connector,
                    "action": action,
                    "result": format!("[Action stub] {}::{}", connector, action),
                    "params": params,
                }))
            }
            NodeType::Passthrough => {
                Ok(serde_json::Value::Object(input.into_iter().collect()))
            }
        }
    }
}

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

    #[test]
    fn test_topological_order_chain() {
        let graph = make_chain_graph();
        let order = graph.topological_order().expect("no cycle");
        assert_eq!(order.len(), 3);
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
        let executor = GraphExecutor::new(DefaultNodeExecutor);
        let ctx = Arc::new(ExecutionContext::new(Uuid::new_v4(), None));
        let input = GraphExecutionInput {
            graph_id: graph.id,
            initial_input: [("x".to_string(), serde_json::json!("hello"))].into(),
            tenant_id: None,
        };

        let result = executor.execute(&graph, input, ctx).await;
        assert_eq!(result.status, ExecutionStatus::Completed);
        assert!(result.error.is_none());
    }

    #[test]
    fn test_retry_policy_delay() {
        let policy = RetryPolicy::default();
        assert_eq!(policy.delay_for(1), Duration::from_millis(100));
        assert_eq!(policy.delay_for(2), Duration::from_millis(200));
    }
}
