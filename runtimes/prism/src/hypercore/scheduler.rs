//! HyperCore Scheduler - AI-aware distributed scheduler
//!
//! The scheduler makes intelligent decisions about where and how to execute
//! Adaptive Execution Cells (AECs) based on:
//! - Resource availability
//! - Latency requirements
//! - Cost optimization
//! - GPU/CPU affinity
//! - AI model routing

use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{debug, warn, info};

use crate::core::{
    CellId,
    PlacementHint, PrismError, PrismResult,
};
use crate::hypercore::{Node, PlacementDecision};
use crate::hypercore::placement::NodeId;

/// Main HyperCore scheduler
pub struct Scheduler {
    config: SchedulerConfig,
    nodes: Arc<RwLock<HashMap<NodeId, Node>>>,
    placements: Arc<RwLock<HashMap<CellId, PlacementDecision>>>,
    /// History of placement decisions for analytics
    placement_history: Arc<RwLock<Vec<PlacementRecord>>>,
    /// Metrics about scheduling decisions
    schedule_stats: Arc<RwLock<ScheduleStats>>,
}

/// A record of a placement decision for auditing and analytics
#[derive(Debug, Clone)]
pub struct PlacementRecord {
    pub cell_id: CellId,
    pub node_id: NodeId,
    pub decision_time: chrono::DateTime<chrono::Utc>,
    pub scheduling_reason: String,
    pub score: u64,
}

/// Scheduling statistics
#[derive(Debug, Clone, Default)]
pub struct ScheduleStats {
    pub total_scheduled: u64,
    pub total_failed: u64,
    pub by_location: std::collections::HashMap<String, u64>,
}

/// Configuration for the scheduler
#[derive(Debug, Clone)]
pub struct SchedulerConfig {
    /// Maximum concurrent executions
    pub max_concurrent: usize,
    /// Maximum queue size
    pub max_queue_size: usize,
    /// Rate limit per second
    pub rate_limit: u64,
}

impl Default for SchedulerConfig {
    fn default() -> Self {
        Self {
            max_concurrent: 64,
            max_queue_size: 1024,
            rate_limit: 1000,
        }
    }
}

impl Scheduler {
    /// Create a new scheduler with default configuration
    pub fn new(config: SchedulerConfig) -> Self {
        Self {
            config,
            nodes: Arc::new(RwLock::new(HashMap::new())),
            placements: Arc::new(RwLock::new(HashMap::new())),
            placement_history: Arc::new(RwLock::new(Vec::new())),
            schedule_stats: Arc::new(RwLock::new(ScheduleStats::default())),
        }
    }

    /// Register a node with the scheduler
    pub async fn register_node(&self, node: Node) -> PrismResult<()> {
        let mut nodes = self.nodes.write().await;
        if nodes.len() >= self.config.max_queue_size {
            return Err(PrismError::SchedulerError("Maximum nodes reached".to_string()));
        }
        debug!(node_id = node.id.as_str(), "Registering node");
        nodes.insert(node.id.clone(), node);
        info!(count = nodes.len(), "Node registered successfully");
        Ok(())
    }

    /// Schedule a cell for execution
    pub async fn schedule(&self, request: ScheduleRequest) -> PrismResult<ScheduleResponse> {
        let nodes = self.nodes.read().await;

        if nodes.is_empty() {
            warn!("No nodes available for scheduling");
            return Err(PrismError::SchedulerError("No available nodes".to_string()));
        }

        // Find best node based on placement criteria
        let best_node = nodes.values()
            .filter(|n| self.is_node_eligible(n, &request))
            .min_by_key(|n| self.calculate_score(n, &request))
            .cloned();

        match best_node {
            Some(node) => {
                let score = self.calculate_score(&node, &request);
                let decision = PlacementDecision {
                    cell_id: request.cell_id,
                    node_id: node.id.clone(),
                    location: node.info.location,
                    config_overrides: HashMap::new(),
                    scheduling_reason: "best_score".to_string(),
                };

                // Record the placement decision
                {
                    let mut placements = self.placements.write().await;
                    placements.insert(request.cell_id, decision.clone());
                }

                // Record in history and update stats
                {
                    let mut history = self.placement_history.write().await;
                    history.push(PlacementRecord {
                        cell_id: request.cell_id,
                        node_id: node.id.clone(),
                        decision_time: chrono::Utc::now(),
                        scheduling_reason: decision.scheduling_reason.clone(),
                        score,
                    });

                    // Keep history bounded to last 10000 decisions
                    if history.len() > 10000 {
                        history.drain(0..1000);
                    }
                }

                {
                    let mut stats = self.schedule_stats.write().await;
                    stats.total_scheduled += 1;
                    let location_key = format!("{:?}", node.info.location);
                    *stats.by_location.entry(location_key).or_insert(0) += 1;
                }

                debug!(cell_id = %request.cell_id, node_id = %node.id.as_str(), score = score, "Scheduling decision");
                Ok(ScheduleResponse {
                    decision,
                })
            }
            None => {
                warn!(cell_id = %request.cell_id, "No eligible node found");
                Err(PrismError::SchedulerError("No eligible node found".to_string()))
            }
        }
    }

    /// Check if a node is eligible for scheduling
    fn is_node_eligible(&self, node: &Node, request: &ScheduleRequest) -> bool {
        // Check resource requirements
        node.info.resources.vcpus >= request.required_vcpus
            && node.info.resources.memory_bytes >= request.required_memory
            && node.is_available()
    }

    /// Calculate placement score for a node (lower is better)
    fn calculate_score(&self, node: &Node, request: &ScheduleRequest) -> u64 {
        let resource_score = node.info.resources.vcpus.saturating_sub(request.required_vcpus) as u64
            + node.info.resources.memory_bytes.saturating_sub(request.required_memory) / (1024 * 1024);

        // Apply location preference
        let location_penalty = match &request.placement_hint {
            Some(hint) if hint.preferred_location == node.info.location => 0,
            None => 0,
            _ => 100,
        };

        resource_score + location_penalty
    }

    /// Get all registered nodes
    pub async fn list_nodes(&self) -> Vec<Node> {
        let nodes = self.nodes.read().await;
        nodes.values().cloned().collect()
    }

    /// Get scheduler config
    pub fn config(&self) -> &SchedulerConfig {
        &self.config
    }

    /// Get all placement decisions
    pub async fn get_placements(&self) -> Vec<(CellId, PlacementDecision)> {
        let placements = self.placements.read().await;
        placements.iter()
            .map(|(cell_id, decision)| (*cell_id, decision.clone()))
            .collect()
    }

    /// Get placement for a specific cell
    pub async fn get_placement(&self, cell_id: &CellId) -> Option<PlacementDecision> {
        let placements = self.placements.read().await;
        placements.get(cell_id).cloned()
    }

    /// Get recent placement history
    pub async fn get_recent_placements(&self, limit: usize) -> Vec<PlacementRecord> {
        let history = self.placement_history.read().await;
        history.iter().rev().take(limit).cloned().collect()
    }

    /// Get scheduling statistics
    pub async fn get_stats(&self) -> ScheduleStats {
        let stats = self.schedule_stats.read().await;
        stats.clone()
    }

    /// Cancel (remove) a placement for a cell
    pub async fn cancel_placement(&self, cell_id: &CellId) -> bool {
        let mut placements = self.placements.write().await;
        placements.remove(cell_id).is_some()
    }
}

/// Request for scheduling a cell
#[derive(Debug, Clone)]
pub struct ScheduleRequest {
    pub cell_id: CellId,
    pub required_vcpus: u32,
    pub required_memory: u64,
    pub placement_hint: Option<PlacementHint>,
}

/// Response from the scheduler
#[derive(Debug, Clone)]
pub struct ScheduleResponse {
    pub decision: PlacementDecision,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::{CellId, ExecutionLocation};
    use crate::hypercore::NodeResources;

    fn create_test_node(id: &str, vcpus: u32, memory: u64) -> Node {
        Node::new(
            id,
            ExecutionLocation::Cloud,
            NodeResources::new(vcpus, memory),
        )
    }

    #[test]
    fn test_scheduler_creation() {
        let config = SchedulerConfig::default();
        let scheduler = Scheduler::new(config);
        assert_eq!(scheduler.config.max_concurrent, 64);
    }

    #[tokio::test]
    async fn test_node_registration() {
        let scheduler = Scheduler::new(SchedulerConfig::default());
        let node = create_test_node("node-1", 4, 8192);

        assert!(scheduler.register_node(node).await.is_ok());
        assert_eq!(scheduler.list_nodes().await.len(), 1);
    }

    #[tokio::test]
    async fn test_empty_scheduler() {
        let scheduler = Scheduler::new(SchedulerConfig::default());
        let request = ScheduleRequest {
            cell_id: CellId::new(),
            required_vcpus: 2,
            required_memory: 4096,
            placement_hint: None,
        };

        assert!(scheduler.schedule(request).await.is_err());
    }
}