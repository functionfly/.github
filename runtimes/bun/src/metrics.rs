//! Metrics collection for Bun runtime
//!
//! Provides execution metrics, performance monitoring, and observability.

use serde::{Deserialize, Serialize};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use tokio::sync::RwLock;

/// Runtime metrics for monitoring and observability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuntimeMetrics {
    /// Total number of executions
    pub total_executions: u64,
    /// Number of successful executions
    pub successful_executions: u64,
    /// Number of failed executions
    pub failed_executions: u64,
    /// Average execution time in milliseconds
    pub avg_execution_time_ms: f64,
    /// Minimum execution time in milliseconds
    pub min_execution_time_ms: u64,
    /// Maximum execution time in milliseconds
    pub max_execution_time_ms: u64,
    /// Total CPU time consumed (ms)
    pub total_cpu_time_ms: u64,
    /// Number of currently active executions
    pub active_executions: u64,
    /// Peak concurrent executions
    pub peak_concurrent: u64,
    /// Timestamp of last execution
    pub last_execution_ts: Option<i64>,
}

impl Default for RuntimeMetrics {
    fn default() -> Self {
        Self {
            total_executions: 0,
            successful_executions: 0,
            failed_executions: 0,
            avg_execution_time_ms: 0.0,
            min_execution_time_ms: u64::MAX,
            max_execution_time_ms: 0,
            total_cpu_time_ms: 0,
            active_executions: 0,
            peak_concurrent: 0,
            last_execution_ts: None,
        }
    }
}

/// Metrics collector for tracking runtime performance
pub struct MetricsCollector {
    metrics: Arc<RwLock<RuntimeMetrics>>,
}

impl MetricsCollector {
    /// Create a new metrics collector
    pub fn new() -> Self {
        Self {
            metrics: Arc::new(RwLock::new(RuntimeMetrics::default())),
        }
    }

    /// Record an execution result
    pub async fn record_execution(&self, execution_time_ms: u64, success: bool) {
        let mut metrics = self.metrics.write().await;

        metrics.total_executions += 1;
        if success {
            metrics.successful_executions += 1;
        } else {
            metrics.failed_executions += 1;
        }

        // Update execution times
        if execution_time_ms < metrics.min_execution_time_ms {
            metrics.min_execution_time_ms = execution_time_ms;
        }
        if execution_time_ms > metrics.max_execution_time_ms {
            metrics.max_execution_time_ms = execution_time_ms;
        }

        // Calculate running average
        let total = metrics.total_executions as f64;
        let current_avg = metrics.avg_execution_time_ms;
        metrics.avg_execution_time_ms = ((total - 1.0) * current_avg + execution_time_ms as f64) / total;

        metrics.total_cpu_time_ms += execution_time_ms;
        metrics.last_execution_ts = Some(chrono::Utc::now().timestamp());
    }

    /// Increment active executions
    pub async fn increment_active(&self) {
        let mut metrics = self.metrics.write().await;
        metrics.active_executions += 1;
        if metrics.active_executions > metrics.peak_concurrent {
            metrics.peak_concurrent = metrics.active_executions;
        }
    }

    /// Decrement active executions
    pub async fn decrement_active(&self) {
        let mut metrics = self.metrics.write().await;
        if metrics.active_executions > 0 {
            metrics.active_executions -= 1;
        }
    }

    /// Get current metrics snapshot
    pub async fn get_metrics(&self) -> RuntimeMetrics {
        self.metrics.read().await.clone()
    }

    /// Reset all metrics
    pub async fn reset(&self) {
        *self.metrics.write().await = RuntimeMetrics::default();
    }
}

impl Default for MetricsCollector {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_record_execution() {
        let collector = MetricsCollector::new();
        collector.record_execution(100, true).await;
        collector.record_execution(200, true).await;
        collector.record_execution(50, false).await;

        let metrics = collector.get_metrics().await;
        assert_eq!(metrics.total_executions, 3);
        assert_eq!(metrics.successful_executions, 2);
        assert_eq!(metrics.failed_executions, 1);
        assert_eq!(metrics.min_execution_time_ms, 50);
        assert_eq!(metrics.max_execution_time_ms, 200);
    }

    #[tokio::test]
    async fn test_concurrent_tracking() {
        let collector = MetricsCollector::new();
        collector.increment_active().await;
        collector.increment_active().await;
        collector.decrement_active().await;

        let metrics = collector.get_metrics().await;
        assert_eq!(metrics.active_executions, 1);
        assert_eq!(metrics.peak_concurrent, 2);
    }
}