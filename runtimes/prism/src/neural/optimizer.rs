//! Neural Execution Optimization - RL-based self-optimization

use std::collections::{HashMap, VecDeque};
use serde::{Deserialize, Serialize};

use crate::core::{ExecutionMetrics, CellId};

/// An execution profile for learning
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionProfile {
    pub cell_id: CellId,
    pub metrics: ExecutionMetrics,
    pub features: ExecutionFeatures,
    pub outcome: ExecutionOutcome,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionFeatures {
    pub input_size_bytes: u64,
    pub memory_limit_mb: u64,
    pub vcpus: u32,
    pub gpu_used: bool,
    pub execution_location: String,
    pub time_of_day: f32, // 0-24 hours
    pub day_of_week: u8,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ExecutionOutcome {
    Success,
    Timeout,
    OOM,
    Error,
}

/// Optimization policy derived from learning
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizationPolicy {
    pub memory_multiplier: f32,
    pub timeout_multiplier: f32,
    pub preferred_location: String,
    pub cache_enabled: bool,
    pub parallel_execution: bool,
}

impl Default for OptimizationPolicy {
    fn default() -> Self {
        Self {
            memory_multiplier: 1.0,
            timeout_multiplier: 1.0,
            preferred_location: "cloud".to_string(),
            cache_enabled: true,
            parallel_execution: false,
        }
    }
}

/// Neural optimizer that learns from execution patterns
pub struct NeuralOptimizer {
    /// Sliding window of recent execution profiles
    history: VecDeque<ExecutionProfile>,
    /// Maximum history size
    max_history: usize,
    /// Current policy
    policy: OptimizationPolicy,
    /// Q-table for reinforcement learning
    q_table: HashMap<String, f32>,
    /// Learning rate
    learning_rate: f32,
    /// Discount factor
    discount_factor: f32,
}

impl NeuralOptimizer {
    pub fn new(max_history: usize) -> Self {
        Self {
            history: VecDeque::new(),
            max_history,
            policy: OptimizationPolicy::default(),
            q_table: HashMap::new(),
            learning_rate: 0.1,
            discount_factor: 0.9,
        }
    }

    /// Record an execution outcome
    pub fn record(&mut self, profile: ExecutionProfile) {
        self.history.push_back(profile);

        // Trim history
        while self.history.len() > self.max_history {
            self.history.pop_front();
        }

        // Update policy based on new data
        self.update_policy();
    }

    /// Get optimization suggestions for a cell
    pub fn suggest(&self, cell_id: &CellId) -> OptimizationSuggestion {
        let state = self.get_state_key(cell_id);

        // Get Q-values for different actions
        let memory_action = self.q_table.get(&format!("{}:memory", state)).cloned().unwrap_or(1.0);
        let timeout_action = self.q_table.get(&format!("{}:timeout", state)).cloned().unwrap_or(1.0);
        let location_action = self.q_table.get(&format!("{}:location", state)).cloned().unwrap_or(0.5);

        // Calculate confidence based on history
        let confidence = self.calculate_confidence();

        OptimizationSuggestion {
            cell_id: *cell_id,
            suggested_memory_mb: (128.0 * memory_action) as u64,
            suggested_timeout_ms: (30000.0 * timeout_action) as u64,
            suggested_location: if location_action > 0.5 { "edge".to_string() } else { "cloud".to_string() },
            cache_recommended: self.policy.cache_enabled,
            confidence,
        }
    }

    /// Calculate confidence based on execution history
    fn calculate_confidence(&self) -> f32 {
        let history_len = self.history.len();
        if history_len == 0 {
            return 0.3; // No history, low confidence
        }

        // Confidence increases with more history, up to a point
        let history_factor = (history_len as f32 / 100.0).min(1.0);

        // Calculate consistency based on outcome distribution
        let success_count = self.history.iter()
            .filter(|p| p.outcome == ExecutionOutcome::Success)
            .count();
        let success_rate = success_count as f32 / history_len as f32;

        // High consistency (all successes or all failures) increases confidence
        // Variance in outcomes reduces confidence
        let failure_count = history_len - success_count;
        let variance = if failure_count > 0 && success_count > 0 {
            // Binary variance: p*(1-p) where p = success_rate
            success_rate * (1.0 - success_rate)
        } else {
            0.0 // Perfect consistency
        };

        // Base confidence from history length + adjustment for consistency
        let base = 0.3 + (0.5 * history_factor);
        let consistency_adjustment = (0.2 * (1.0 - variance * 4.0)).max(-0.2);

        (base + consistency_adjustment).min(0.95).max(0.1)
    }

    /// Get the current policy
    pub fn get_policy(&self) -> &OptimizationPolicy {
        &self.policy
    }

    /// Update policy based on execution history
    fn update_policy(&mut self) {
        // Calculate success rate
        let total = self.history.len();
        if total == 0 {
            return;
        }

        let successes = self.history.iter()
            .filter(|p| p.outcome == ExecutionOutcome::Success)
            .count();

        let success_rate = successes as f32 / total as f32;

        // Adjust memory multiplier based on OOM occurrences
        let oom_count = self.history.iter()
            .filter(|p| p.outcome == ExecutionOutcome::OOM)
            .count();

        if oom_count > 0 {
            self.policy.memory_multiplier *= 1.1;
        } else if success_rate > 0.95 {
            self.policy.memory_multiplier *= 0.95;
        }

        // Adjust timeout multiplier based on timeouts
        let timeout_count = self.history.iter()
            .filter(|p| p.outcome == ExecutionOutcome::Timeout)
            .count();

        if timeout_count > 0 {
            self.policy.timeout_multiplier *= 1.2;
        } else if success_rate > 0.95 {
            self.policy.timeout_multiplier *= 0.95;
        }

        // Normalize multipliers
        self.policy.memory_multiplier = self.policy.memory_multiplier.max(0.5).min(4.0);
        self.policy.timeout_multiplier = self.policy.timeout_multiplier.max(0.5).min(4.0);
    }

    /// Get a state key for Q-learning
    fn get_state_key(&self, cell_id: &CellId) -> String {
        cell_id.to_string()
    }

    /// Update Q-value based on reward
    pub fn update_q(&mut self, state: &str, action: &str, reward: f32, next_state: &str) {
        let current_q = self.q_table.get(&format!("{}:{}", state, action)).cloned().unwrap_or(0.0);

        // Get max Q-value for next state
        let max_next_q = ["memory", "timeout", "location"]
            .iter()
            .filter_map(|a| self.q_table.get(&format!("{}:{}", next_state, a)))
            .fold(0.0f32, |max, &q| max.max(q));

        // Q-learning update
        let new_q = current_q + self.learning_rate * (reward + self.discount_factor * max_next_q - current_q);
        self.q_table.insert(format!("{}:{}", state, action), new_q);
    }
}

impl Default for NeuralOptimizer {
    fn default() -> Self {
        Self::new(1000)
    }
}

/// An optimization suggestion
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizationSuggestion {
    pub cell_id: CellId,
    pub suggested_memory_mb: u64,
    pub suggested_timeout_ms: u64,
    pub suggested_location: String,
    pub cache_recommended: bool,
    pub confidence: f32,
}