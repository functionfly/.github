//! State graph memory — tracks per-node execution history for Phase 7 optimizer.
//!
//! Stores decision logs, execution paths, success/failure rates, and latency
//! histograms per node. The optimizer (Phase 7) reads this to drive threshold-based
//! graph rewriting.
//!
//! ## What Gets Tracked
//!
//! - **Per-node metrics**: call count, success count, failure count, avg latency
//! - **Decision log**: which branches were taken (Control nodes)
//! - **Execution path**: topological node order for each run
//! - **Pattern history**: detected sequences (e.g. "timeout_rate > 0.1")

use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant};

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;

/// Number of historical executions to retain per node for metrics.
const DEFAULT_HISTORY_LEN: usize = 1000;

/// Per-node execution metrics — updated after each execution.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct NodeMetrics {
    pub node_id: String,
    /// Optional human-readable name (set from graph metadata).
    pub node_name: Option<String>,
    pub call_count: u64,
    pub success_count: u64,
    pub failure_count: u64,
    pub total_latency_ms: u64,
    /// Rolling latencies for percentile calculation.
    latency_history: VecDeque<u64>,
    /// Number of times this node triggered a retry.
    pub retry_count: u64,
    /// Number of times this node timed out.
    pub timeout_count: u64,
    /// Current configured timeout for this node (ms).
    pub timeout_ms: u64,
}

impl NodeMetrics {
    pub fn success_rate(&self) -> f64 {
        if self.call_count == 0 {
            return 1.0;
        }
        self.success_count as f64 / self.call_count as f64
    }

    pub fn failure_rate(&self) -> f64 {
        if self.call_count == 0 {
            return 0.0;
        }
        self.failure_count as f64 / self.call_count as f64
    }

    pub fn timeout_rate(&self) -> f64 {
        if self.call_count == 0 {
            return 0.0;
        }
        self.timeout_count as f64 / self.call_count as f64
    }

    pub fn avg_latency_ms(&self) -> f64 {
        if self.call_count == 0 {
            return 0.0;
        }
        self.total_latency_ms as f64 / self.call_count as f64
    }

    pub fn p95_latency_ms(&self) -> f64 {
        if self.latency_history.is_empty() {
            return 0.0;
        }
        let mut sorted: Vec<_> = self.latency_history.iter().copied().collect();
        sorted.sort_unstable();
        let idx = (sorted.len() as f64 * 0.95).ceil() as usize;
        sorted[idx.min(sorted.len() - 1)] as f64
    }

    pub fn p50_latency_ms(&self) -> f64 {
        if self.latency_history.is_empty() {
            return 0.0;
        }
        let mut sorted: Vec<_> = self.latency_history.iter().copied().collect();
        sorted.sort_unstable();
        let idx = (sorted.len() as f64 * 0.50).floor() as usize;
        sorted[idx.min(sorted.len() - 1)] as f64
    }

    /// p99 latency in milliseconds.
    pub fn p99_latency_ms(&self) -> f64 {
        if self.latency_history.is_empty() {
            return 0.0;
        }
        let mut sorted: Vec<_> = self.latency_history.iter().copied().collect();
        sorted.sort_unstable();
        let idx = ((sorted.len() as f64 * 0.99).ceil() as usize).min(sorted.len() - 1);
        sorted[idx] as f64
    }

    /// Ratio of p99 to p50 latency — high values indicate high variance.
    pub fn latency_variance_ratio(&self) -> f64 {
        let p50 = self.p50_latency_ms();
        let p99 = self.p99_latency_ms();
        if p50 == 0.0 {
            return 0.0;
        }
        p99 / p50
    }

    /// Record a successful execution.
    pub fn record_success(&mut self, latency_ms: u64) {
        self.call_count += 1;
        self.success_count += 1;
        self.total_latency_ms += latency_ms;
        if self.latency_history.len() >= 100 {
            self.latency_history.pop_front();
        }
        self.latency_history.push_back(latency_ms);
    }

    /// Record a failed execution.
    pub fn record_failure(&mut self, latency_ms: u64) {
        self.call_count += 1;
        self.failure_count += 1;
        self.total_latency_ms += latency_ms;
        if self.latency_history.len() >= 100 {
            self.latency_history.pop_front();
        }
        self.latency_history.push_back(latency_ms);
    }

    /// Record a timeout.
    pub fn record_timeout(&mut self) {
        self.call_count += 1;
        self.timeout_count += 1;
    }

    /// Record a retry attempt.
    pub fn record_retry(&mut self) {
        self.retry_count += 1;
    }
}

/// A single decision recorded during graph execution.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionDecision {
    pub execution_id: uuid::Uuid,
    pub node_id: String,
    pub decision_type: DecisionType,
    pub detail: String,
    pub latency_ms: u64,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum DecisionType {
    BranchTaken,
    LoopIterated,
    NodeCompleted,
    NodeFailed,
    RetryTriggered,
    OptimizationApplied,
}

impl std::fmt::Display for DecisionType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            DecisionType::BranchTaken => write!(f, "branch_taken"),
            DecisionType::LoopIterated => write!(f, "loop_iterated"),
            DecisionType::NodeCompleted => write!(f, "node_completed"),
            DecisionType::NodeFailed => write!(f, "node_failed"),
            DecisionType::RetryTriggered => write!(f, "retry_triggered"),
            DecisionType::OptimizationApplied => write!(f, "optimization_applied"),
        }
    }
}

/// A complete execution record for a single graph run.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GraphExecutionRecord {
    pub execution_id: uuid::Uuid,
    pub graph_id: uuid::Uuid,
    pub tenant_id: Option<String>,
    pub status: String,
    pub started_at: chrono::DateTime<chrono::Utc>,
    pub completed_at: Option<chrono::DateTime<chrono::Utc>>,
    pub duration_ms: Option<u64>,
    /// Topological execution order for this run.
    pub execution_path: Vec<String>,
    /// Decision log for this execution.
    pub decisions: Vec<ExecutionDecision>,
    pub node_results: HashMap<String, NodeResultSummary>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeResultSummary {
    pub status: String,
    pub latency_ms: u64,
    pub attempts: u32,
    pub error: Option<String>,
}

/// State graph memory — tracks metrics, decisions, and execution history.
///
/// This is the source of truth for the Phase 7 self-optimization engine.
pub struct StateGraphMemory {
    /// Per-node rolling metrics.
    node_metrics: RwLock<HashMap<String, NodeMetrics>>,
    /// Per-graph execution history (last N executions per tenant+graph).
    execution_history: RwLock<HashMap<(String, String), VecDeque<GraphExecutionRecord>>>,
    /// Global decision log (last N decisions).
    decision_log: RwLock<VecDeque<ExecutionDecision>>,
    /// Maximum history entries to retain per node.
    max_history: usize,
    /// Maximum execution records to retain per (tenant, graph) key.
    max_executions: usize,
    /// Maximum decisions to retain in the global log.
    max_decisions: usize,
}

impl StateGraphMemory {
    pub fn new() -> Self {
        Self {
            node_metrics: RwLock::new(HashMap::new()),
            execution_history: RwLock::new(HashMap::new()),
            decision_log: RwLock::new(VecDeque::new()),
            max_history: DEFAULT_HISTORY_LEN,
            max_executions: 100,
            max_decisions: 10_000,
        }
    }

    /// Record a successful node execution.
    pub async fn record_node_success(&self, node_id: &str, latency_ms: u64) {
        let mut metrics_map = self.node_metrics.write().await;
        let metrics = metrics_map
            .entry(node_id.to_string())
            .or_insert_with(|| NodeMetrics {
                node_id: node_id.to_string(),
                ..Default::default()
            });
        metrics.record_success(latency_ms);
    }

    /// Record a failed node execution.
    pub async fn record_node_failure(&self, node_id: &str, latency_ms: u64) {
        let mut metrics_map = self.node_metrics.write().await;
        let metrics = metrics_map
            .entry(node_id.to_string())
            .or_insert_with(|| NodeMetrics {
                node_id: node_id.to_string(),
                ..Default::default()
            });
        metrics.record_failure(latency_ms);
    }

    /// Record a node timeout.
    pub async fn record_node_timeout(&self, node_id: &str) {
        let mut metrics_map = self.node_metrics.write().await;
        let metrics = metrics_map
            .entry(node_id.to_string())
            .or_insert_with(|| NodeMetrics {
                node_id: node_id.to_string(),
                ..Default::default()
            });
        metrics.record_timeout();
    }

    /// Record a retry attempt.
    pub async fn record_retry(&self, node_id: &str) {
        let mut metrics_map = self.node_metrics.write().await;
        let metrics = metrics_map
            .entry(node_id.to_string())
            .or_insert_with(|| NodeMetrics {
                node_id: node_id.to_string(),
                ..Default::default()
            });
        metrics.record_retry();
    }

    /// Record an execution decision.
    pub async fn record_decision(&self, decision: ExecutionDecision) {
        // Update global decision log
        {
            let mut log = self.decision_log.write().await;
            log.push_back(decision.clone());
            while log.len() > self.max_decisions {
                log.pop_front();
            }
        }

        // Update execution history if we have an execution_id
        if decision.execution_id != uuid::Uuid::nil() {
            let key = (
                decision.timestamp.to_rfc3339(),
                decision.node_id.clone(),
            );
            // Decision log is global; per-execution decisions are stored
            // in the GraphExecutionRecord pushed at the end of execution.
            let _ = key;
        }
    }

    /// Start recording an execution (creates a new record).
    pub async fn start_execution(
        &self,
        execution_id: uuid::Uuid,
        graph_id: uuid::Uuid,
        tenant_id: Option<&str>,
        status: &str,
    ) {
        let record = GraphExecutionRecord {
            execution_id,
            graph_id,
            tenant_id: tenant_id.map(String::from),
            status: status.to_string(),
            started_at: chrono::Utc::now(),
            completed_at: None,
            duration_ms: None,
            execution_path: Vec::new(),
            decisions: Vec::new(),
            node_results: HashMap::new(),
        };

        let tenant = tenant_id.unwrap_or("default");
        let mut history = self.execution_history.write().await;
        let deque = history
            .entry((tenant.to_string(), graph_id.to_string()))
            .or_insert_with(VecDeque::new);
        deque.push_back(record);
        while deque.len() > self.max_executions {
            deque.pop_front();
        }
    }

    /// Finalize an execution record after completion.
    pub async fn finish_execution(
        &self,
        execution_id: uuid::Uuid,
        status: &str,
        execution_path: Vec<String>,
        node_results: HashMap<String, NodeResultSummary>,
    ) {
        let mut history = self.execution_history.write().await;
        for (_, deque) in history.iter_mut() {
            for record in deque.iter_mut() {
                if record.execution_id == execution_id {
                    record.status = status.to_string();
                    record.completed_at = Some(chrono::Utc::now());
                    record.duration_ms = record.completed_at.map(|c| {
                        (c - record.started_at).num_milliseconds() as u64
                    });
                    record.execution_path = execution_path.clone();
                    record.node_results = node_results.clone();
                    break;
                }
            }
        }
    }

    /// Get metrics for a specific node.
    pub async fn get_node_metrics(&self, node_id: &str) -> Option<NodeMetrics> {
        let metrics_map = self.node_metrics.read().await;
        metrics_map.get(node_id).cloned()
    }

    /// Get all node metrics (for the optimizer).
    pub async fn all_node_metrics(&self) -> HashMap<String, NodeMetrics> {
        let metrics_map = self.node_metrics.read().await;
        metrics_map.clone()
    }

    /// Alias for `all_node_metrics()` — used by the optimizer.
    pub async fn get_all_metrics(&self) -> HashMap<String, NodeMetrics> {
        self.all_node_metrics().await
    }

    /// Get the configured timeout for a node.
    pub async fn get_node_timeout(&self, node_id: uuid::Uuid) -> u64 {
        let metrics_map = self.node_metrics.read().await;
        metrics_map
            .get(&node_id.to_string())
            .map(|m| m.timeout_ms)
            .unwrap_or(30_000) // default 30s
    }

    /// Update the timeout for a node (called when optimizer applies a suggestion).
    pub async fn set_node_timeout(&self, node_id: uuid::Uuid, timeout_ms: u64) {
        let mut metrics_map = self.node_metrics.write().await;
        let entry = metrics_map
            .entry(node_id.to_string())
            .or_insert_with(|| NodeMetrics {
                node_id: node_id.to_string(),
                ..Default::default()
            });
        entry.timeout_ms = timeout_ms;
    }

    /// Get recent execution records for a graph (any tenant).
    pub async fn get_execution_history(&self, graph_id: uuid::Uuid) -> Vec<GraphExecutionRecord> {
        let history = self.execution_history.read().await;
        let mut results = Vec::new();
        for ((_tenant, g_id), deque) in history.iter() {
            if *g_id == graph_id.to_string() {
                results.extend(deque.iter().cloned());
            }
        }
        results
    }

    /// Get the global decision log.
    pub async fn decision_log(&self) -> Vec<ExecutionDecision> {
        let log = self.decision_log.read().await;
        log.iter().cloned().collect()
    }

    /// Get patterns useful for Phase 7 optimizer.
    ///
    /// Analyzes recent execution history and returns detected patterns:
    /// - `timeout_rate > 0.1` → suggest timeout adjustment
    /// - `success_rate > 0.95 && call_count > 100` → suggest node consolidation
    /// - `avg_latency > 2x median` → suggest parallel execution
    pub async fn detect_patterns(&self) -> Vec<DetectedPattern> {
        let mut patterns = Vec::new();
        let metrics_map = self.node_metrics.read().await;

        for (node_id, metrics) in metrics_map.iter() {
            if metrics.timeout_rate() > 0.1 {
                patterns.push(DetectedPattern {
                    node_id: node_id.clone(),
                    pattern_type: PatternType::HighTimeoutRate,
                    confidence: (metrics.timeout_rate() * 10.0).min(0.9),
                    detail: format!(
                        "timeout_rate={:.2}, call_count={}",
                        metrics.timeout_rate(),
                        metrics.call_count
                    ),
                });
            }

            if metrics.success_rate() > 0.95 && metrics.call_count > 100 {
                patterns.push(DetectedPattern {
                    node_id: node_id.clone(),
                    pattern_type: PatternType::StableHighSuccess,
                    confidence: 0.75,
                    detail: format!(
                        "success_rate={:.2}, call_count={}",
                        metrics.success_rate(),
                        metrics.call_count
                    ),
                });
            }

            let avg = metrics.avg_latency_ms();
            let p95 = metrics.p95_latency_ms();
            if avg > 0.0 && p95 / avg > 3.0 {
                patterns.push(DetectedPattern {
                    node_id: node_id.clone(),
                    pattern_type: PatternType::HighLatencyVariance,
                    confidence: 0.7,
                    detail: format!(
                        "avg={:.0}ms, p95={:.0}ms, ratio={:.1}x",
                        avg,
                        p95,
                        p95 / avg
                    ),
                });
            }
        }

        patterns
    }
}

/// A detected optimization pattern.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DetectedPattern {
    pub node_id: String,
    pub pattern_type: PatternType,
    pub confidence: f64,
    pub detail: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PatternType {
    HighTimeoutRate,
    StableHighSuccess,
    HighLatencyVariance,
    LowSuccessRate,
}

impl std::fmt::Display for PatternType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            PatternType::HighTimeoutRate => write!(f, "high_timeout_rate"),
            PatternType::StableHighSuccess => write!(f, "stable_high_success"),
            PatternType::HighLatencyVariance => write!(f, "high_latency_variance"),
            PatternType::LowSuccessRate => write!(f, "low_success_rate"),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_state() -> StateGraphMemory {
        StateGraphMemory::new()
    }

    #[tokio::test]
    async fn test_record_success() {
        let state = make_state();
        state.record_node_success("node-a", 50).await;
        let metrics = state.get_node_metrics("node-a").await.unwrap();
        assert_eq!(metrics.call_count, 1);
        assert_eq!(metrics.success_count, 1);
        assert_eq!(metrics.failure_count, 0);
    }

    #[tokio::test]
    async fn test_record_failure() {
        let state = make_state();
        state.record_node_failure("node-b", 30).await;
        let metrics = state.get_node_metrics("node-b").await.unwrap();
        assert_eq!(metrics.call_count, 1);
        assert_eq!(metrics.failure_count, 1);
        assert_eq!(metrics.success_rate(), 0.0);
    }

    #[tokio::test]
    async fn test_success_rate() {
        let state = make_state();
        for _ in 0..9 {
            state.record_node_success("node-c", 10).await;
        }
        state.record_node_failure("node-c", 20).await;

        let metrics = state.get_node_metrics("node-c").await.unwrap();
        assert_eq!(metrics.success_rate(), 0.9);
        assert_eq!(metrics.failure_rate(), 0.1);
    }

    #[tokio::test]
    async fn test_timeout_rate() {
        let state = make_state();
        for _ in 0..9 {
            state.record_node_timeout("node-d").await;
        }
        state.record_node_success("node-d", 50).await;

        let metrics = state.get_node_metrics("node-d").await.unwrap();
        assert!((metrics.timeout_rate() - 0.9).abs() < 0.001);
    }

    #[tokio::test]
    async fn test_decision_log() {
        let state = make_state();
        let decision = ExecutionDecision {
            execution_id: uuid::Uuid::new_v4(),
            node_id: "node-e".to_string(),
            decision_type: DecisionType::BranchTaken,
            detail: "taken=true".to_string(),
            latency_ms: 10,
            timestamp: chrono::Utc::now(),
        };
        state.record_decision(decision).await;
        let log = state.decision_log().await;
        assert_eq!(log.len(), 1);
    }

    #[tokio::test]
    async fn test_detect_patterns() {
        let state = make_state();

        // Node with high timeout rate (9/10 = 90% timeout rate → pattern detected)
        for _ in 0..9 {
            state.record_node_timeout("slow-node").await;
        }
        state.record_node_success("slow-node", 100).await;

        let patterns = state.detect_patterns().await;
        let timeout_patterns: Vec<_> = patterns
            .iter()
            .filter(|p| p.pattern_type == PatternType::HighTimeoutRate)
            .collect();

        assert!(!timeout_patterns.is_empty());
        assert_eq!(timeout_patterns[0].node_id, "slow-node");
    }

    #[tokio::test]
    async fn test_execution_history() {
        let state = make_state();
        let exec_id = uuid::Uuid::new_v4();
        let graph_id = uuid::Uuid::new_v4();

        state
            .start_execution(exec_id, graph_id, Some("tenant-x"), "running")
            .await;

        let history = state.get_execution_history(graph_id).await;
        assert_eq!(history.len(), 1);
        assert_eq!(history[0].status, "running");
    }
}