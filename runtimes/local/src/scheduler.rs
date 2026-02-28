//! Bin-packing scheduler for routing function invocations to the least-loaded node.
//!
//! This module implements a lightweight scheduler that tracks per-node resource
//! utilisation and routes new invocations to the node with the most headroom,
//! preventing hot-spots and enabling cost-efficient packing of functions onto
//! a small number of hosts.
//!
//! # Phase 4 implementation
//!
//! Addresses the gap identified in `plans/SANDBOX_EXECUTION_LAYER.md`:
//! > No bin-packing / scheduling across multiple nodes — **Medium**
//! > Add a lightweight scheduler that tracks per-node capacity and routes new
//! > invocations to the least-loaded node.
//!
//! # Algorithm
//!
//! The scheduler uses a **Best-Fit Decreasing** heuristic:
//! 1. Sort candidate nodes by available capacity (descending).
//! 2. Pick the first node that can satisfy the request's resource requirements.
//! 3. If no node has enough capacity, return `None` (caller should queue or
//!    reject the request).
//!
//! Node capacity is tracked as `(cpu_used_percent, memory_used_mb)` and is
//! updated after each execution via `record_execution`.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};

/// Snapshot of a node's current resource utilisation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeCapacity {
    /// Unique node identifier (e.g. hostname or IP:port).
    pub node_id: String,
    /// Total CPU capacity in millicores (1000 = 1 vCPU).
    pub total_cpu_millicores: u32,
    /// Currently used CPU in millicores.
    pub used_cpu_millicores: u32,
    /// Total RAM in MB.
    pub total_memory_mb: u32,
    /// Currently used RAM in MB.
    pub used_memory_mb: u32,
    /// Number of active function executions on this node.
    pub active_executions: u32,
    /// Maximum concurrent executions this node supports.
    pub max_executions: u32,
    /// Whether the node is healthy and accepting requests.
    pub healthy: bool,
    /// Last time this node's capacity was updated.
    #[serde(skip)]
    pub last_updated: Option<Instant>,
}

impl NodeCapacity {
    /// Create a new node capacity record.
    pub fn new(
        node_id: impl Into<String>,
        total_cpu_millicores: u32,
        total_memory_mb: u32,
        max_executions: u32,
    ) -> Self {
        Self {
            node_id: node_id.into(),
            total_cpu_millicores,
            used_cpu_millicores: 0,
            total_memory_mb,
            used_memory_mb: 0,
            active_executions: 0,
            max_executions,
            healthy: true,
            last_updated: Some(Instant::now()),
        }
    }

    /// Available CPU in millicores.
    pub fn available_cpu_millicores(&self) -> u32 {
        self.total_cpu_millicores.saturating_sub(self.used_cpu_millicores)
    }

    /// Available RAM in MB.
    pub fn available_memory_mb(&self) -> u32 {
        self.total_memory_mb.saturating_sub(self.used_memory_mb)
    }

    /// CPU utilisation as a fraction [0.0, 1.0].
    pub fn cpu_utilisation(&self) -> f64 {
        if self.total_cpu_millicores == 0 {
            return 1.0;
        }
        self.used_cpu_millicores as f64 / self.total_cpu_millicores as f64
    }

    /// Memory utilisation as a fraction [0.0, 1.0].
    pub fn memory_utilisation(&self) -> f64 {
        if self.total_memory_mb == 0 {
            return 1.0;
        }
        self.used_memory_mb as f64 / self.total_memory_mb as f64
    }

    /// Combined utilisation score (higher = more loaded).
    ///
    /// Weighted average: 60% CPU + 40% memory.
    pub fn utilisation_score(&self) -> f64 {
        0.6 * self.cpu_utilisation() + 0.4 * self.memory_utilisation()
    }

    /// Whether this node can satisfy the given resource request.
    pub fn can_fit(&self, req: &SchedulingRequest) -> bool {
        self.healthy
            && self.active_executions < self.max_executions
            && self.available_cpu_millicores() >= req.cpu_millicores
            && self.available_memory_mb() >= req.memory_mb
    }
}

/// Resource requirements for a single function invocation.
#[derive(Debug, Clone)]
pub struct SchedulingRequest {
    /// CPU requirement in millicores.
    pub cpu_millicores: u32,
    /// Memory requirement in MB.
    pub memory_mb: u32,
    /// Preferred node (if any); scheduler will try this node first.
    pub preferred_node: Option<String>,
}

impl SchedulingRequest {
    /// Create a scheduling request from a memory limit and vCPU count.
    pub fn from_resources(memory_mb: u32, vcpus: u32) -> Self {
        Self {
            cpu_millicores: vcpus * 1000,
            memory_mb,
            preferred_node: None,
        }
    }
}

/// Scheduling decision returned by the scheduler.
#[derive(Debug, Clone)]
pub struct SchedulingDecision {
    /// The selected node ID.
    pub node_id: String,
    /// Utilisation score of the selected node at decision time.
    pub utilisation_score: f64,
}

/// Bin-packing scheduler.
///
/// Wrap in `Arc<BinPackingScheduler>` and share across request handlers.
pub struct BinPackingScheduler {
    /// Per-node capacity, keyed by node ID.
    nodes: Arc<RwLock<HashMap<String, NodeCapacity>>>,
    /// How long before a node's capacity record is considered stale.
    stale_threshold: Duration,
}

impl BinPackingScheduler {
    /// Create a new scheduler.
    pub fn new(stale_threshold_secs: u64) -> Self {
        Self {
            nodes: Arc::new(RwLock::new(HashMap::new())),
            stale_threshold: Duration::from_secs(stale_threshold_secs),
        }
    }

    /// Register or update a node's capacity.
    pub async fn upsert_node(&self, capacity: NodeCapacity) {
        let mut nodes = self.nodes.write().await;
        nodes.insert(capacity.node_id.clone(), capacity);
    }

    /// Remove a node from the scheduler (e.g. when it goes offline).
    pub async fn remove_node(&self, node_id: &str) {
        let mut nodes = self.nodes.write().await;
        nodes.remove(node_id);
    }

    /// Mark a node as unhealthy so it stops receiving new requests.
    pub async fn mark_unhealthy(&self, node_id: &str) {
        let mut nodes = self.nodes.write().await;
        if let Some(node) = nodes.get_mut(node_id) {
            node.healthy = false;
            tracing::warn!("BinPackingScheduler: node {} marked unhealthy", node_id);
        }
    }

    /// Mark a node as healthy again.
    pub async fn mark_healthy(&self, node_id: &str) {
        let mut nodes = self.nodes.write().await;
        if let Some(node) = nodes.get_mut(node_id) {
            node.healthy = true;
            tracing::info!("BinPackingScheduler: node {} marked healthy", node_id);
        }
    }

    /// Select the best node for the given request using Best-Fit Decreasing.
    ///
    /// Returns `None` if no node can satisfy the request.
    pub async fn schedule(&self, req: &SchedulingRequest) -> Option<SchedulingDecision> {
        let nodes = self.nodes.read().await;

        // Filter to healthy nodes that can fit the request and are not stale
        let mut candidates: Vec<&NodeCapacity> = nodes
            .values()
            .filter(|n| {
                // Check staleness
                if let Some(last_updated) = n.last_updated {
                    if last_updated.elapsed() > self.stale_threshold {
                        tracing::debug!(
                            "BinPackingScheduler: skipping stale node {}",
                            n.node_id
                        );
                        return false;
                    }
                }
                n.can_fit(req)
            })
            .collect();

        if candidates.is_empty() {
            tracing::warn!(
                "BinPackingScheduler: no node can satisfy request (cpu={} mem={}MB)",
                req.cpu_millicores,
                req.memory_mb
            );
            return None;
        }

        // Check preferred node first
        if let Some(ref preferred) = req.preferred_node {
            if let Some(node) = candidates.iter().find(|n| &n.node_id == preferred) {
                return Some(SchedulingDecision {
                    node_id: node.node_id.clone(),
                    utilisation_score: node.utilisation_score(),
                });
            }
        }

        // Sort by utilisation score ascending (least loaded first) — Best-Fit
        candidates.sort_by(|a, b| {
            a.utilisation_score()
                .partial_cmp(&b.utilisation_score())
                .unwrap_or(std::cmp::Ordering::Equal)
        });

        let selected = candidates[0];
        tracing::debug!(
            "BinPackingScheduler: selected node {} (score={:.2})",
            selected.node_id,
            selected.utilisation_score()
        );

        Some(SchedulingDecision {
            node_id: selected.node_id.clone(),
            utilisation_score: selected.utilisation_score(),
        })
    }

    /// Record that an execution has started on a node (increases used resources).
    pub async fn record_execution_start(&self, node_id: &str, req: &SchedulingRequest) {
        let mut nodes = self.nodes.write().await;
        if let Some(node) = nodes.get_mut(node_id) {
            node.used_cpu_millicores = node
                .used_cpu_millicores
                .saturating_add(req.cpu_millicores);
            node.used_memory_mb = node.used_memory_mb.saturating_add(req.memory_mb);
            node.active_executions += 1;
            node.last_updated = Some(Instant::now());
        }
    }

    /// Record that an execution has finished on a node (decreases used resources).
    pub async fn record_execution_end(&self, node_id: &str, req: &SchedulingRequest) {
        let mut nodes = self.nodes.write().await;
        if let Some(node) = nodes.get_mut(node_id) {
            node.used_cpu_millicores = node
                .used_cpu_millicores
                .saturating_sub(req.cpu_millicores);
            node.used_memory_mb = node.used_memory_mb.saturating_sub(req.memory_mb);
            node.active_executions = node.active_executions.saturating_sub(1);
            node.last_updated = Some(Instant::now());
        }
    }

    /// Return a snapshot of all node capacities.
    pub async fn node_capacities(&self) -> Vec<NodeCapacity> {
        let nodes = self.nodes.read().await;
        nodes.values().cloned().collect()
    }

    /// Return scheduler statistics.
    pub async fn stats(&self) -> SchedulerStats {
        let nodes = self.nodes.read().await;
        let total_nodes = nodes.len();
        let healthy_nodes = nodes.values().filter(|n| n.healthy).count();
        let total_active: u32 = nodes.values().map(|n| n.active_executions).sum();
        let avg_utilisation = if total_nodes > 0 {
            nodes.values().map(|n| n.utilisation_score()).sum::<f64>() / total_nodes as f64
        } else {
            0.0
        };

        SchedulerStats {
            total_nodes,
            healthy_nodes,
            total_active_executions: total_active,
            average_utilisation: avg_utilisation,
        }
    }
}

/// Scheduler statistics.
#[derive(Debug, Clone)]
pub struct SchedulerStats {
    pub total_nodes: usize,
    pub healthy_nodes: usize,
    pub total_active_executions: u32,
    pub average_utilisation: f64,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_node(id: &str, cpu: u32, mem: u32, max_exec: u32) -> NodeCapacity {
        NodeCapacity::new(id, cpu, mem, max_exec)
    }

    #[tokio::test]
    async fn test_schedule_selects_least_loaded() {
        let sched = BinPackingScheduler::new(60);

        let mut n1 = make_node("node-1", 4000, 8192, 20);
        n1.used_cpu_millicores = 3000; // 75% CPU
        n1.used_memory_mb = 4096;     // 50% mem

        let mut n2 = make_node("node-2", 4000, 8192, 20);
        n2.used_cpu_millicores = 1000; // 25% CPU
        n2.used_memory_mb = 2048;     // 25% mem

        sched.upsert_node(n1).await;
        sched.upsert_node(n2).await;

        let req = SchedulingRequest::from_resources(512, 1);
        let decision = sched.schedule(&req).await.unwrap();

        // node-2 is less loaded
        assert_eq!(decision.node_id, "node-2");
    }

    #[tokio::test]
    async fn test_schedule_returns_none_when_no_capacity() {
        let sched = BinPackingScheduler::new(60);

        let mut n1 = make_node("node-1", 1000, 512, 5);
        n1.used_cpu_millicores = 1000; // fully loaded
        n1.used_memory_mb = 512;

        sched.upsert_node(n1).await;

        let req = SchedulingRequest::from_resources(256, 1);
        let decision = sched.schedule(&req).await;
        assert!(decision.is_none());
    }

    #[tokio::test]
    async fn test_record_execution_updates_capacity() {
        let sched = BinPackingScheduler::new(60);
        sched.upsert_node(make_node("node-1", 4000, 8192, 20)).await;

        let req = SchedulingRequest::from_resources(1024, 2);
        sched.record_execution_start("node-1", &req).await;

        let nodes = sched.node_capacities().await;
        let node = nodes.iter().find(|n| n.node_id == "node-1").unwrap();
        assert_eq!(node.used_cpu_millicores, 2000);
        assert_eq!(node.used_memory_mb, 1024);
        assert_eq!(node.active_executions, 1);

        sched.record_execution_end("node-1", &req).await;
        let nodes = sched.node_capacities().await;
        let node = nodes.iter().find(|n| n.node_id == "node-1").unwrap();
        assert_eq!(node.active_executions, 0);
    }

    #[tokio::test]
    async fn test_preferred_node_respected() {
        let sched = BinPackingScheduler::new(60);
        sched.upsert_node(make_node("node-1", 4000, 8192, 20)).await;
        sched.upsert_node(make_node("node-2", 4000, 8192, 20)).await;

        let req = SchedulingRequest {
            cpu_millicores: 500,
            memory_mb: 256,
            preferred_node: Some("node-1".to_string()),
        };

        let decision = sched.schedule(&req).await.unwrap();
        assert_eq!(decision.node_id, "node-1");
    }

    #[tokio::test]
    async fn test_unhealthy_node_excluded() {
        let sched = BinPackingScheduler::new(60);
        sched.upsert_node(make_node("node-1", 4000, 8192, 20)).await;
        sched.upsert_node(make_node("node-2", 4000, 8192, 20)).await;
        sched.mark_unhealthy("node-1").await;

        let req = SchedulingRequest::from_resources(256, 1);
        let decision = sched.schedule(&req).await.unwrap();
        assert_eq!(decision.node_id, "node-2");
    }
}
