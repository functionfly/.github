//! Metrics collection for Kotlin/JVM runtime
//!
//! Provides execution metrics, performance monitoring, and observability.

use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;
use std::time::Instant;

/// Runtime metrics for monitoring and observability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuntimeMetrics {
    /// Total number of executions
    pub total_executions: u64,
    /// Number of successful executions
    pub successful_executions: u64,
    /// Number of failed executions
    pub failed_executions: u64,
    /// Number of timed out executions
    pub timed_out_executions: u64,
    /// Number of security violations blocked
    pub security_violations: u64,
    /// Average execution time in milliseconds
    pub avg_execution_time_ms: f64,
    /// Minimum execution time in milliseconds
    pub min_execution_time_ms: u64,
    /// Maximum execution time in milliseconds
    pub max_execution_time_ms: u64,
    /// Total CPU time consumed (ms)
    pub total_cpu_time_ms: u64,
    /// Current memory usage in MB
    pub current_memory_mb: f64,
    /// Peak memory usage in MB
    pub peak_memory_mb: f64,
    /// Number of currently active executions
    pub active_executions: u64,
    /// Peak concurrent executions
    pub peak_concurrent: u64,
    /// Uptime in seconds
    pub uptime_secs: f64,
    /// Timestamp of last execution
    pub last_execution_ts: Option<i64>,
    /// Average code size in bytes
    pub avg_code_size_bytes: u64,
    /// Maximum code size in bytes
    pub max_code_size_bytes: u64,
    /// Total code size samples
    pub code_size_samples: u64,
}

impl Default for RuntimeMetrics {
    fn default() -> Self {
        Self {
            total_executions: 0,
            successful_executions: 0,
            failed_executions: 0,
            timed_out_executions: 0,
            security_violations: 0,
            avg_execution_time_ms: 0.0,
            min_execution_time_ms: u64::MAX,
            max_execution_time_ms: 0,
            total_cpu_time_ms: 0,
            current_memory_mb: 0.0,
            peak_memory_mb: 0.0,
            active_executions: 0,
            peak_concurrent: 0,
            uptime_secs: 0.0,
            last_execution_ts: None,
            avg_code_size_bytes: 0,
            max_code_size_bytes: 0,
            code_size_samples: 0,
        }
    }
}

/// Internal metrics state
#[derive(Debug, Default)]
struct MetricsState {
    total_executions: u64,
    successful_executions: u64,
    failed_executions: u64,
    timed_out_executions: u64,
    security_violations: u64,
    total_cpu_time_ms: u64,
    peak_memory_mb: u64,
    active_executions: u64,
    peak_concurrent: u64,
    min_execution_time_ms: u64,
    max_execution_time_ms: u64,
    total_code_size_bytes: u64,
    max_code_size_bytes: u64,
    code_size_samples: u64,
}

/// Metrics collector for tracking runtime performance
pub struct MetricsCollector {
    start_time: Instant,
    metrics: Arc<RwLock<MetricsState>>,
}

impl MetricsCollector {
    /// Create a new metrics collector
    pub fn new(_runtime_name: impl Into<String>) -> Self {
        Self {
            start_time: Instant::now(),
            metrics: Arc::new(RwLock::new(MetricsState::default())),
        }
    }

    /// Record an execution result
    pub async fn record_execution(&self, execution_time_ms: u64, memory_mb: u64) {
        let mut state = self.metrics.write().await;

        state.total_executions += 1;
        state.successful_executions += 1;
        state.total_cpu_time_ms += execution_time_ms;

        // Update execution time stats
        if execution_time_ms < state.min_execution_time_ms {
            state.min_execution_time_ms = execution_time_ms;
        }
        if execution_time_ms > state.max_execution_time_ms {
            state.max_execution_time_ms = execution_time_ms;
        }

        // Update memory peak
        if memory_mb > state.peak_memory_mb {
            state.peak_memory_mb = memory_mb;
        }

        // Decrement active
        if state.active_executions > 0 {
            state.active_executions -= 1;
        }
    }

    /// Record a failed execution
    pub async fn record_failure(&self, execution_time_ms: u64) {
        let mut state = self.metrics.write().await;

        state.total_executions += 1;
        state.failed_executions += 1;
        state.total_cpu_time_ms += execution_time_ms;

        if state.active_executions > 0 {
            state.active_executions -= 1;
        }
    }

    /// Record a timeout
    pub async fn record_timeout(&self) {
        let mut state = self.metrics.write().await;

        state.total_executions += 1;
        state.timed_out_executions += 1;

        if state.active_executions > 0 {
            state.active_executions -= 1;
        }
    }

    /// Record a security violation
    pub async fn record_security_violation(&self) {
        let mut state = self.metrics.write().await;
        state.security_violations += 1;
    }

    /// Update current memory usage
    pub async fn update_memory(&self, memory_mb: u64) {
        let mut state = self.metrics.write().await;
        if memory_mb > state.peak_memory_mb {
            state.peak_memory_mb = memory_mb;
        }
    }

    /// Increment active executions
    pub async fn start_execution(&self) {
        let mut state = self.metrics.write().await;
        state.active_executions += 1;
        if state.active_executions > state.peak_concurrent {
            state.peak_concurrent = state.active_executions;
        }
    }

    /// Decrement active executions
    pub async fn end_execution(&self) {
        let mut state = self.metrics.write().await;
        if state.active_executions > 0 {
            state.active_executions -= 1;
        }
    }

    /// Record code size for histogram tracking
    pub async fn record_code_size(&self, size_bytes: usize) {
        let mut state = self.metrics.write().await;
        state.total_code_size_bytes += size_bytes as u64;
        state.code_size_samples += 1;
        if size_bytes as u64 > state.max_code_size_bytes {
            state.max_code_size_bytes = size_bytes as u64;
        }
    }

    /// Get current metrics snapshot
    pub async fn get_metrics(&self) -> RuntimeMetrics {
        let state = self.metrics.read().await;
        let uptime = self.start_time.elapsed().as_secs_f64();
        let total = state.total_executions.max(1);

        RuntimeMetrics {
            total_executions: state.total_executions,
            successful_executions: state.successful_executions,
            failed_executions: state.failed_executions,
            timed_out_executions: state.timed_out_executions,
            security_violations: state.security_violations,
            avg_execution_time_ms: state.total_cpu_time_ms as f64 / total as f64,
            min_execution_time_ms: state.min_execution_time_ms,
            max_execution_time_ms: state.max_execution_time_ms,
            total_cpu_time_ms: state.total_cpu_time_ms,
            current_memory_mb: state.peak_memory_mb as f64,
            peak_memory_mb: state.peak_memory_mb as f64,
            active_executions: state.active_executions,
            peak_concurrent: state.peak_concurrent,
            uptime_secs: uptime,
            last_execution_ts: None,
            avg_code_size_bytes: if state.code_size_samples > 0 {
                state.total_code_size_bytes / state.code_size_samples
            } else {
                0
            },
            max_code_size_bytes: state.max_code_size_bytes,
            code_size_samples: state.code_size_samples,
        }
    }

    /// Encode metrics in Prometheus text format
    pub fn encode_prometheus(&self) -> String {
        // We need sync access, so we use blocking get_metrics pattern
        // In production, this would use a channel or atomic values
        let metrics = RuntimeMetrics::default();
        let mut output = String::new();

        output.push_str(&format!(concat!(
            "# HELP kotlin_runtime_total_executions Total number of code executions\n",
            "# TYPE kotlin_runtime_total_executions counter\n",
            "kotlin_runtime_total_executions {}\n",
            "# HELP kotlin_runtime_successful_executions Number of successful executions\n",
            "# TYPE kotlin_runtime_successful_executions counter\n",
            "kotlin_runtime_successful_executions {}\n",
            "# HELP kotlin_runtime_failed_executions Number of failed executions\n",
            "# TYPE kotlin_runtime_failed_executions counter\n",
            "kotlin_runtime_failed_executions {}\n",
            "# HELP kotlin_runtime_timed_out_executions Number of timed out executions\n",
            "# TYPE kotlin_runtime_timed_out_executions counter\n",
            "kotlin_runtime_timed_out_executions {}\n",
            "# HELP kotlin_runtime_security_violations_total Number of security violations blocked\n",
            "# TYPE kotlin_runtime_security_violations_total counter\n",
            "kotlin_runtime_security_violations_total {}\n",
            "# HELP kotlin_runtime_current_memory_mb Current memory usage in MB\n",
            "# TYPE kotlin_runtime_current_memory_mb gauge\n",
            "kotlin_runtime_current_memory_mb {:.2}\n",
            "# HELP kotlin_runtime_peak_memory_mb Peak memory usage in MB\n",
            "# TYPE kotlin_runtime_peak_memory_mb gauge\n",
            "kotlin_runtime_peak_memory_mb {:.2}\n",
            "# HELP kotlin_runtime_currently_executing Number of currently executing functions\n",
            "# TYPE kotlin_runtime_currently_executing gauge\n",
            "kotlin_runtime_currently_executing {}\n",
            "# HELP kotlin_runtime_uptime_seconds Runtime uptime in seconds\n",
            "# TYPE kotlin_runtime_uptime_seconds gauge\n",
            "kotlin_runtime_uptime_seconds {:.2}\n",
            "# HELP kotlin_runtime_avg_execution_time_ms Average execution duration in milliseconds\n",
            "# TYPE kotlin_runtime_avg_execution_time_ms gauge\n",
            "kotlin_runtime_avg_execution_time_ms {:.2}\n",
            "# HELP kotlin_runtime_peak_concurrent Peak concurrent executions\n",
            "# TYPE kotlin_runtime_peak_concurrent gauge\n",
            "kotlin_runtime_peak_concurrent {}\n",
        ),
            metrics.total_executions,
            metrics.successful_executions,
            metrics.failed_executions,
            metrics.timed_out_executions,
            metrics.security_violations,
            metrics.current_memory_mb,
            metrics.peak_memory_mb,
            metrics.active_executions,
            metrics.uptime_secs,
            metrics.avg_execution_time_ms,
            metrics.peak_concurrent,
        ));

        output
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_metrics_creation() {
        let collector = MetricsCollector::new("test-runtime");
        let metrics = collector.get_metrics().await;

        assert_eq!(metrics.total_executions, 0);
        assert_eq!(metrics.successful_executions, 0);
    }

    #[tokio::test]
    async fn test_record_execution() {
        let collector = MetricsCollector::new("test-runtime");
        collector.record_execution(100, 50).await;
        collector.record_execution(200, 60).await;

        let metrics = collector.get_metrics().await;
        assert_eq!(metrics.total_executions, 2);
        assert_eq!(metrics.successful_executions, 2);
    }

    #[tokio::test]
    async fn test_record_failure() {
        let collector = MetricsCollector::new("test-runtime");
        collector.record_failure(50).await;

        let metrics = collector.get_metrics().await;
        assert_eq!(metrics.total_executions, 1);
        assert_eq!(metrics.failed_executions, 1);
    }

    #[tokio::test]
    async fn test_record_timeout() {
        let collector = MetricsCollector::new("test-runtime");
        collector.record_timeout().await;

        let metrics = collector.get_metrics().await;
        assert_eq!(metrics.timed_out_executions, 1);
    }

    #[tokio::test]
    async fn test_peak_memory() {
        let collector = MetricsCollector::new("test-runtime");
        collector.update_memory(100).await;
        collector.update_memory(150).await;
        collector.update_memory(120).await;

        let metrics = collector.get_metrics().await;
        assert_eq!(metrics.peak_memory_mb, 150.0);
    }
}