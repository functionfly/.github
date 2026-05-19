//! Ruby Runtime Metrics
//!
//! Metrics collection for Ruby execution runtime using atomic counters.

use std::sync::Arc;
use std::sync::atomic::{AtomicU64, Ordering};

/// Runtime metrics collector using atomic counters
#[derive(Clone)]
pub struct MetricsCollector {
    // Execution metrics
    total_executions: Arc<AtomicU64>,
    successful_executions: Arc<AtomicU64>,
    failed_executions: Arc<AtomicU64>,
    total_execution_time_ms: Arc<AtomicU64>,
    active_executions: Arc<AtomicU64>,
    // Memory metrics
    current_memory_bytes: Arc<AtomicU64>,
    peak_memory_bytes: Arc<AtomicU64>,
    // Sandbox metrics
    sandbox_violations: Arc<AtomicU64>,
    code_sanitizations: Arc<AtomicU64>,
    // Resource limit metrics
    memory_limit_hits: Arc<AtomicU64>,
    timeout_hits: Arc<AtomicU64>,
    stack_overflow_hits: Arc<AtomicU64>,
    // Request metrics
    total_requests: Arc<AtomicU64>,
    rate_limited_requests: Arc<AtomicU64>,
}

impl MetricsCollector {
    /// Create a new metrics collector
    pub fn new() -> Self {
        Self {
            total_executions: Arc::new(AtomicU64::new(0)),
            successful_executions: Arc::new(AtomicU64::new(0)),
            failed_executions: Arc::new(AtomicU64::new(0)),
            total_execution_time_ms: Arc::new(AtomicU64::new(0)),
            active_executions: Arc::new(AtomicU64::new(0)),
            current_memory_bytes: Arc::new(AtomicU64::new(0)),
            peak_memory_bytes: Arc::new(AtomicU64::new(0)),
            sandbox_violations: Arc::new(AtomicU64::new(0)),
            code_sanitizations: Arc::new(AtomicU64::new(0)),
            memory_limit_hits: Arc::new(AtomicU64::new(0)),
            timeout_hits: Arc::new(AtomicU64::new(0)),
            stack_overflow_hits: Arc::new(AtomicU64::new(0)),
            total_requests: Arc::new(AtomicU64::new(0)),
            rate_limited_requests: Arc::new(AtomicU64::new(0)),
        }
    }

    /// Increment total executions
    pub fn inc_executions(&self) {
        self.total_executions.fetch_add(1, Ordering::Relaxed);
    }

    /// Record successful execution
    pub fn record_success(&self, execution_time_ms: u64) {
        self.successful_executions.fetch_add(1, Ordering::Relaxed);
        self.total_execution_time_ms.fetch_add(execution_time_ms, Ordering::Relaxed);
    }

    /// Record failed execution
    pub fn record_failure(&self) {
        self.failed_executions.fetch_add(1, Ordering::Relaxed);
    }

    /// Update active executions count
    pub fn set_active_executions(&self, count: usize) {
        self.active_executions.store(count as u64, Ordering::Relaxed);
    }

    /// Increment active executions
    pub fn inc_active(&self) {
        self.active_executions.fetch_add(1, Ordering::Relaxed);
    }

    /// Decrement active executions
    pub fn dec_active(&self) {
        self.active_executions.fetch_sub(1, Ordering::Relaxed);
    }

    /// Update current memory usage
    pub fn set_current_memory(&self, bytes: u64) {
        self.current_memory_bytes.store(bytes, Ordering::Relaxed);
        // Update peak if necessary
        let current_peak = self.peak_memory_bytes.load(Ordering::Relaxed);
        if bytes > current_peak {
            self.peak_memory_bytes.store(bytes, Ordering::Relaxed);
        }
    }

    /// Record sandbox violation
    pub fn record_sandbox_violation(&self) {
        self.sandbox_violations.fetch_add(1, Ordering::Relaxed);
    }

    /// Record code sanitization
    pub fn record_sanitization(&self) {
        self.code_sanitizations.fetch_add(1, Ordering::Relaxed);
    }

    /// Record memory limit hit
    pub fn record_memory_limit_hit(&self) {
        self.memory_limit_hits.fetch_add(1, Ordering::Relaxed);
    }

    /// Record timeout hit
    pub fn record_timeout_hit(&self) {
        self.timeout_hits.fetch_add(1, Ordering::Relaxed);
    }

    /// Record stack overflow hit
    pub fn record_stack_overflow(&self) {
        self.stack_overflow_hits.fetch_add(1, Ordering::Relaxed);
    }

    /// Increment total requests
    pub fn inc_requests(&self) {
        self.total_requests.fetch_add(1, Ordering::Relaxed);
    }

    /// Record rate limited request
    pub fn record_rate_limited(&self) {
        self.rate_limited_requests.fetch_add(1, Ordering::Relaxed);
    }

    /// Get current metrics as Prometheus format
    pub fn gather(&self) -> String {
        let summary = self.summary();
        format!(
            "# HELP ruby_runtime_executions_total Total number of executions\n\
             # TYPE ruby_runtime_executions_total counter\n\
             ruby_runtime_executions_total {}\n\
             # HELP ruby_runtime_successful_executions Total successful executions\n\
             # TYPE ruby_runtime_successful_executions counter\n\
             ruby_runtime_successful_executions {}\n\
             # HELP ruby_runtime_failed_executions Total failed executions\n\
             # TYPE ruby_runtime_failed_executions counter\n\
             ruby_runtime_failed_executions {}\n\
             # HELP ruby_runtime_active_executions Currently active executions\n\
             # TYPE ruby_runtime_active_executions gauge\n\
             ruby_runtime_active_executions {}\n\
             # HELP ruby_runtime_memory_bytes Current memory usage in bytes\n\
             # TYPE ruby_runtime_memory_bytes gauge\n\
             ruby_runtime_memory_bytes {}\n\
             # HELP ruby_runtime_peak_memory_bytes Peak memory usage in bytes\n\
             # TYPE ruby_runtime_peak_memory_bytes gauge\n\
             ruby_runtime_peak_memory_bytes {}\n",
            summary.total_executions,
            summary.successful_executions,
            summary.failed_executions,
            summary.active_executions,
            summary.current_memory_bytes,
            summary.peak_memory_bytes
        )
    }

    /// Get runtime metrics summary
    pub fn summary(&self) -> RuntimeMetrics {
        RuntimeMetrics {
            total_executions: self.total_executions.load(Ordering::Relaxed),
            successful_executions: self.successful_executions.load(Ordering::Relaxed),
            failed_executions: self.failed_executions.load(Ordering::Relaxed),
            active_executions: self.active_executions.load(Ordering::Relaxed) as usize,
            current_memory_bytes: self.current_memory_bytes.load(Ordering::Relaxed),
            peak_memory_bytes: self.peak_memory_bytes.load(Ordering::Relaxed),
            sandbox_violations: self.sandbox_violations.load(Ordering::Relaxed),
            code_sanitizations: self.code_sanitizations.load(Ordering::Relaxed),
            memory_limit_hits: self.memory_limit_hits.load(Ordering::Relaxed),
            timeout_hits: self.timeout_hits.load(Ordering::Relaxed),
            stack_overflow_hits: self.stack_overflow_hits.load(Ordering::Relaxed),
            total_requests: self.total_requests.load(Ordering::Relaxed),
            rate_limited_requests: self.rate_limited_requests.load(Ordering::Relaxed),
        }
    }
}

impl Default for MetricsCollector {
    fn default() -> Self {
        Self::new()
    }
}

/// Runtime metrics summary
#[derive(Debug, Clone, serde::Serialize)]
pub struct RuntimeMetrics {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub failed_executions: u64,
    pub active_executions: usize,
    pub current_memory_bytes: u64,
    pub peak_memory_bytes: u64,
    pub sandbox_violations: u64,
    pub code_sanitizations: u64,
    pub memory_limit_hits: u64,
    pub timeout_hits: u64,
    pub stack_overflow_hits: u64,
    pub total_requests: u64,
    pub rate_limited_requests: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_metrics_collector() {
        let metrics = MetricsCollector::new();
        metrics.inc_executions();
        metrics.record_success(100);
        assert_eq!(metrics.summary().total_executions, 1);
        assert_eq!(metrics.summary().successful_executions, 1);
    }
}