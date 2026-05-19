//! HyperCore metrics

use serde::{Deserialize, Serialize};

/// Metrics collected by the HyperCore scheduler
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SchedulerMetrics {
    pub total_nodes: u32,
    pub available_nodes: u32,
    pub active_placements: u32,
    pub total_cpu_cores: u32,
    pub total_memory_bytes: u64,
    pub total_gpu_count: u32,
}

impl SchedulerMetrics {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_nodes(mut self, count: u32) -> Self {
        self.total_nodes = count;
        self
    }
}