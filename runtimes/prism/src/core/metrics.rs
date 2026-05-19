//! Execution metrics for Prism Runtime

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Metrics collected during execution
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ExecutionMetrics {
    /// Execution duration in milliseconds
    pub duration_ms: u64,
    /// Peak memory usage in bytes
    pub memory_used_bytes: u64,
    /// Average CPU usage (0.0 to 1.0)
    pub cpu_usage_percent: f64,
    /// Estimated cost in USD
    pub cost_usd: f64,
    /// Cache hit rate (0.0 to 1.0)
    pub cache_hit_rate: f32,
    /// Bytes transferred over network
    pub bytes_transferred: u64,
    /// Timestamp when execution started
    pub started_at: Option<DateTime<Utc>>,
    /// Timestamp when execution completed
    pub completed_at: Option<DateTime<Utc>>,
}

impl ExecutionMetrics {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_duration(mut self, duration_ms: u64) -> Self {
        self.duration_ms = duration_ms;
        self
    }

    pub fn with_memory(mut self, bytes: u64) -> Self {
        self.memory_used_bytes = bytes;
        self
    }

    pub fn with_cpu(mut self, percent: f64) -> Self {
        self.cpu_usage_percent = percent;
        self
    }

    pub fn with_cost(mut self, cost: f64) -> Self {
        self.cost_usd = cost;
        self
    }
}

/// Cost profile for a capability or execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CostProfile {
    /// Cost per function call in USD
    pub per_call_usd: f64,
    /// Cost per MB transferred in USD
    pub per_mb_usd: f64,
    /// Whether this is free
    pub is_free: bool,
}

impl Default for CostProfile {
    fn default() -> Self {
        Self {
            per_call_usd: 0.0,
            per_mb_usd: 0.0,
            is_free: true,
        }
    }
}

/// Performance profile for capabilities
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceProfile {
    /// Average latency in milliseconds
    pub avg_latency_ms: u32,
    /// P99 latency in milliseconds
    pub p99_latency_ms: u32,
    /// Throughput in requests per second
    pub throughput_rps: u32,
    /// Trust score (0.0 to 1.0)
    pub trust_score: f32,
    /// Cost profile
    pub cost: CostProfile,
}

impl Default for PerformanceProfile {
    fn default() -> Self {
        Self {
            avg_latency_ms: 100,
            p99_latency_ms: 500,
            throughput_rps: 1000,
            trust_score: 0.9,
            cost: CostProfile::default(),
        }
    }
}