//! WASM Fusion Engine - Dynamic execution graphs with module merging
//!
//! This module provides the FusionGraph, FusionNode, and FusionEdge types
//! for defining execution graphs. The actual execution is handled by the
//! FusionEngine in engine.rs.

use std::sync::Arc;
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::core::{CellId, ExecutionMetrics, PrismResult};
use crate::neural::ExecutionOutcome;

/// Callback type for execution metrics - enables RL feedback loop integration
pub type ExecutionMetricsCallback = Arc<dyn Fn(CellId, ExecutionMetrics, ExecutionOutcome) + Send + Sync>;

// Re-export engine types for backwards compatibility
pub use crate::wasm_fusion::engine::{FusionEngine, FusionEngineConfig};

/// A fusion graph defining how modules are linked and executed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FusionGraph {
    pub graph_id: String,
    pub nodes: Vec<FusionNode>,
    pub edges: Vec<FusionEdge>,
    pub config: FusionConfig,
}

impl FusionGraph {
    pub fn new(_name: impl Into<String>) -> Self {
        Self {
            graph_id: Uuid::new_v4().to_string(),
            nodes: Vec::new(),
            edges: Vec::new(),
            config: FusionConfig::default(),
        }
    }

    pub fn add_node(&mut self, node: FusionNode) -> &str {
        self.nodes.push(node);
        &self.nodes.last().unwrap().node_id
    }

    pub fn add_edge(&mut self, edge: FusionEdge) {
        self.edges.push(edge);
    }
}

/// A node in a fusion graph
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FusionNode {
    pub node_id: String,
    pub name: String,
    pub node_type: FusionNodeType,
    pub config: NodeConfig,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum FusionNodeType {
    Wasm,
    Python,
    FunctionChain,
    StreamMap,
    StreamFilter,
    StreamReduce,
}

/// Configuration for a fusion node
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeConfig {
    pub entry_point: String,
    pub timeout_ms: u32,
    pub memory_limit_mb: u64,
    pub imports: Vec<String>,
}

impl Default for NodeConfig {
    fn default() -> Self {
        Self {
            entry_point: "main".to_string(),
            timeout_ms: 30_000,
            memory_limit_mb: 128,
            imports: Vec::new(),
        }
    }
}

/// An edge connecting two fusion nodes
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FusionEdge {
    pub edge_id: String,
    pub source: String,
    pub target: String,
    pub edge_type: EdgeType,
    pub output_mapping: Option<String>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum EdgeType {
    DataFlow,
    ControlFlow,
    Stream,
}

/// Configuration for the fusion engine
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FusionConfig {
    pub enable_streaming: bool,
    pub enable_module_merging: bool,
    pub enable_live_patch: bool,
    pub max_fusion_modules: usize,
}

impl Default for FusionConfig {
    fn default() -> Self {
        Self {
            enable_streaming: true,
            enable_module_merging: true,
            enable_live_patch: false,
            max_fusion_modules: 16,
        }
    }
}

/// Executor for fusion graphs - wrapper that delegates to real engine
#[derive(Clone)]
pub struct FusionExecutor {
    engine: Arc<RwLock<FusionEngine>>,
    metrics_callback: Arc<RwLock<Option<ExecutionMetricsCallback>>>,
    /// Last execution metrics for RL feedback loop (exec_time_ms, memory_used_bytes, fuel_consumed, success)
    last_metrics: Arc<RwLock<Option<ExecutionSnapshot>>>,
}

/// Lightweight snapshot of execution metrics for RL feedback propagation
#[derive(Debug, Clone)]
pub struct ExecutionSnapshot {
    pub exec_time_ms: u64,
    pub memory_used_bytes: u64,
    pub fuel_consumed: u64,
    pub success: bool,
    pub error: Option<String>,
    /// CBOR-encoded WasmCpuState captured during execution
    pub cpu_state: Option<Vec<u8>>,
}

impl FusionExecutor {
    pub fn new() -> PrismResult<Self> {
        Ok(Self {
            engine: Arc::new(RwLock::new(FusionEngine::with_defaults()?)),
            metrics_callback: Arc::new(RwLock::new(None)),
            last_metrics: Arc::new(RwLock::new(None)),
        })
    }

    pub async fn execute_graph(
        &self,
        graph: &FusionGraph,
        input: &[u8],
    ) -> PrismResult<Vec<u8>> {
        let engine = self.engine.read().await;
        let result = engine.execute(graph, input).await;

        // Collect metrics from the engine's last node execution
        if let Some(last) = engine.last_result() {
            let snapshot = ExecutionSnapshot {
                exec_time_ms: last.exec_time_ms,
                memory_used_bytes: last.memory_used_bytes,
                fuel_consumed: last.fuel_consumed,
                success: last.success,
                error: last.error,
                cpu_state: last.cpu_state,
            };

            // Store for later retrieval by the HTTP handler
            let mut guard = self.last_metrics.write().await;
            *guard = Some(snapshot.clone());

            // Fire the RL metrics callback if registered
            let cb = self.metrics_callback.read().await;
            if let Some(ref callback) = *cb {
                let metrics = ExecutionMetrics {
                    duration_ms: snapshot.exec_time_ms,
                    memory_used_bytes: snapshot.memory_used_bytes,
                    ..Default::default()
                };
                let outcome = if snapshot.success {
                    ExecutionOutcome::Success
                } else {
                    ExecutionOutcome::Error
                };
                // Parse cell_id from graph_id (set to cell_id_str in the HTTP handler)
                if let Ok(uuid) = uuid::Uuid::parse_str(&graph.graph_id) {
                    callback(CellId::from_uuid(uuid), metrics, outcome);
                }
            }
        }

        result
    }

    /// Register a WASM module for use in fusion graphs
    pub async fn register_module(&self, id: &str, wasm_bytes: &[u8]) -> PrismResult<()> {
        let engine = self.engine.write().await;
        engine.register_module(id, wasm_bytes).await
    }

    /// Set callback for execution metrics (enables RL feedback loop)
    pub fn set_metrics_callback(&self, callback: ExecutionMetricsCallback) {
        let cb = self.metrics_callback.clone();
        let mut guard = futures::executor::block_on(cb.write());
        *guard = Some(callback);
    }

    /// Clear the metrics callback
    pub fn clear_metrics_callback(&self) {
        let mut guard = futures::executor::block_on(self.metrics_callback.write());
        *guard = None;
    }

    /// Take the last execution metrics snapshot (consumes the stored value)
    pub async fn take_last_metrics(&self) -> Option<ExecutionSnapshot> {
        self.last_metrics.write().await.take()
    }

    /// Get the underlying engine (read lock)
    pub async fn engine(&self) -> tokio::sync::RwLockReadGuard<'_, FusionEngine> {
        self.engine.read().await
    }
}

impl Default for FusionExecutor {
    fn default() -> Self {
        Self::new().expect("Failed to create default FusionExecutor")
    }
}