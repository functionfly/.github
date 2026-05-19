//! Execution profile collection for neural optimization

use std::collections::{HashMap, VecDeque};

use crate::neural::optimizer::{ExecutionProfile, ExecutionFeatures, ExecutionOutcome};
use crate::core::{ExecutionMetrics, CellId};

/// Collects execution profiles
pub struct ProfileCollector {
    profiles: HashMap<CellId, VecDeque<ExecutionProfile>>,
    max_profiles_per_cell: usize,
}

impl ProfileCollector {
    pub fn new(max_profiles_per_cell: usize) -> Self {
        Self {
            profiles: HashMap::new(),
            max_profiles_per_cell,
        }
    }

    /// Record an execution
    pub fn record(
        &mut self,
        cell_id: CellId,
        metrics: ExecutionMetrics,
        features: ExecutionFeatures,
        outcome: ExecutionOutcome,
    ) {
        let profile = ExecutionProfile {
            cell_id,
            metrics,
            features,
            outcome,
        };

        let cell_profiles = self.profiles.entry(cell_id).or_insert_with(VecDeque::new);
        cell_profiles.push_back(profile);

        // Trim
        while cell_profiles.len() > self.max_profiles_per_cell {
            cell_profiles.pop_front();
        }
    }

    /// Get all profiles for a cell
    pub fn get_profiles(&self, cell_id: &CellId) -> Vec<&ExecutionProfile> {
        self.profiles
            .get(cell_id)
            .map(|p| p.iter().collect())
            .unwrap_or_default()
    }

    /// Get average metrics for a cell
    pub fn get_average_metrics(&self, cell_id: &CellId) -> Option<ExecutionMetrics> {
        let profiles = self.profiles.get(cell_id)?;
        if profiles.is_empty() {
            return None;
        }

        let total = profiles.len() as u64;
        let total_duration: u64 = profiles.iter().map(|p| p.metrics.duration_ms).sum();
        let total_memory: u64 = profiles.iter().map(|p| p.metrics.memory_used_bytes).sum();
        let total_cpu: f64 = profiles.iter().map(|p| p.metrics.cpu_usage_percent as f64).sum();
        let avg_duration = total_duration / total;
        let avg_memory = total_memory / total;
        let avg_cpu = (total_cpu / profiles.len() as f64) as f32;

        Some(ExecutionMetrics {
            duration_ms: avg_duration,
            memory_used_bytes: avg_memory,
            cpu_usage_percent: avg_cpu as f64,
            cost_usd: 0.0,
            cache_hit_rate: 0.0,
            bytes_transferred: 0,
            started_at: None,
            completed_at: None,
        })
    }

    /// Get success rate for a cell
    pub fn get_success_rate(&self, cell_id: &CellId) -> f32 {
        let profiles = match self.profiles.get(cell_id) {
            Some(p) => p,
            None => return 0.0,
        };
        if profiles.is_empty() {
            return 0.0;
        }

        let successes = profiles.iter()
            .filter(|p| p.outcome == ExecutionOutcome::Success)
            .count();

        successes as f32 / profiles.len() as f32
    }
}

impl Default for ProfileCollector {
    fn default() -> Self {
        Self::new(100)
    }
}