//! Node representation in the HyperCore scheduler

use std::collections::HashSet;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::hypercore::placement::NodeId;
use crate::core::ExecutionLocation;

/// Node status
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum NodeStatus {
    Available,
    Busy,
    Unavailable,
    Draining,
}

/// Resources available on a node
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct NodeResources {
    pub vcpus: u32,
    pub memory_bytes: u64,
    pub gpu_count: u32,
    pub gpu_memory_mb: u32,
}

impl NodeResources {
    pub fn new(vcpus: u32, memory_bytes: u64) -> Self {
        Self {
            vcpus,
            memory_bytes,
            gpu_count: 0,
            gpu_memory_mb: 0,
        }
    }

    pub fn with_gpu(mut self, gpu_count: u32, gpu_memory_mb: u32) -> Self {
        self.gpu_count = gpu_count;
        self.gpu_memory_mb = gpu_memory_mb;
        self
    }
}

/// Static information about a node
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeInfo {
    pub node_id: String,
    pub node_type: String,  // "cloud", "edge", "browser"
    pub location: ExecutionLocation,
    pub regions: Vec<String>,
    pub resources: NodeResources,
    pub max_concurrent: u32,
    pub advertised_capabilities: Vec<String>,
    pub version: String,
}

/// A node in the distributed scheduler
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    pub id: NodeId,
    pub info: NodeInfo,
    pub status: NodeStatus,
    pub active_cells: HashSet<String>,
    pub registered_at: DateTime<Utc>,
    pub last_heartbeat: DateTime<Utc>,
}

impl Node {
    pub fn new(
        node_id: impl Into<String>,
        location: ExecutionLocation,
        resources: NodeResources,
    ) -> Self {
        let node_id_str = node_id.into();
        Self {
            id: NodeId::new(&node_id_str),
            info: NodeInfo {
                node_id: node_id_str.clone(),
                node_type: match location {
                    ExecutionLocation::Cloud => "cloud".to_string(),
                    ExecutionLocation::Edge => "edge".to_string(),
                    ExecutionLocation::Browser => "browser".to_string(),
                    ExecutionLocation::Robotic => "robotic".to_string(),
                    ExecutionLocation::Mobile => "mobile".to_string(),
                    ExecutionLocation::IoT => "iot".to_string(),
                },
                location,
                regions: Vec::new(),
                resources,
                max_concurrent: 1000,
                advertised_capabilities: Vec::new(),
                version: "1.0.0".to_string(),
            },
            status: NodeStatus::Available,
            active_cells: HashSet::new(),
            registered_at: Utc::now(),
            last_heartbeat: Utc::now(),
        }
    }

    /// Check if node is available for scheduling
    pub fn is_available(&self) -> bool {
        self.status == NodeStatus::Available
            && self.active_cells.len() < self.info.max_concurrent as usize
    }

    /// Add an active cell to this node
    pub fn add_cell(&mut self, cell_id: &str) {
        self.active_cells.insert(cell_id.to_string());
    }

    /// Remove an active cell from this node
    pub fn remove_cell(&mut self, cell_id: &str) {
        self.active_cells.remove(cell_id);
    }

    /// Get current load factor (0.0 to 1.0)
    pub fn load_factor(&self) -> f32 {
        self.active_cells.len() as f32 / self.info.max_concurrent as f32
    }

    /// Update heartbeat timestamp
    pub fn heartbeat(&mut self) {
        self.last_heartbeat = Utc::now();
    }
}