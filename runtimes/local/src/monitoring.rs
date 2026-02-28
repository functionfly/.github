//! Resource monitoring and performance profiling for the sandbox.
//!
//! This module provides comprehensive monitoring of function executions,
//! resource usage, and performance metrics to optimize for budget constraints.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};

use crate::logging::StructuredLogger;

/// Performance metrics for a single function execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionMetrics {
    pub function_name: String,
    pub function_version: String,
    pub execution_time_ms: u64,
    pub cpu_fuel_used: u64,
    pub memory_used_mb: f64,
    pub peak_memory_mb: f64,
    pub cache_hit: bool,
    pub cold_start: bool,
    pub error_occurred: bool,
    pub timestamp: u64,
}

/// Resource usage statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResourceStats {
    pub total_executions: u64,
    pub total_execution_time_ms: u64,
    pub average_execution_time_ms: f64,
    pub peak_memory_usage_mb: f64,
    pub total_memory_used_mb: f64,
    pub cache_hit_rate: f64,
    pub error_rate: f64,
    pub functions_served: usize,
    pub uptime_seconds: u64,
}

/// Per-function resource limits and tracking
#[derive(Debug, Clone)]
pub struct FunctionLimits {
    pub max_memory_mb: u32,
    pub max_cpu_time_ms: u64,
    pub max_concurrent: usize,
    pub current_concurrent: usize,
    pub total_executions: u64,
    pub last_execution: Instant,
}

/// Comprehensive monitoring system
pub struct ResourceMonitor {
    /// Global execution metrics
    metrics: Arc<RwLock<Vec<ExecutionMetrics>>>,
    /// Per-function resource limits and tracking
    function_limits: Arc<RwLock<HashMap<String, FunctionLimits>>>,
    /// Resource usage statistics
    stats: Arc<RwLock<ResourceStats>>,
    /// Start time for uptime calculation
    start_time: Instant,
    /// Logger for structured logging
    logger: Option<Arc<StructuredLogger>>,
    /// Maximum metrics to retain in memory
    max_metrics_retained: usize,
}

impl ResourceMonitor {
    /// Create a new resource monitor
    pub fn new(logger: Option<Arc<StructuredLogger>>) -> Self {
        Self {
            metrics: Arc::new(RwLock::new(Vec::new())),
            function_limits: Arc::new(RwLock::new(HashMap::new())),
            stats: Arc::new(RwLock::new(ResourceStats {
                total_executions: 0,
                total_execution_time_ms: 0,
                average_execution_time_ms: 0.0,
                peak_memory_usage_mb: 0.0,
                total_memory_used_mb: 0.0,
                cache_hit_rate: 0.0,
                error_rate: 0.0,
                functions_served: 0,
                uptime_seconds: 0,
            })),
            start_time: Instant::now(),
            logger,
            max_metrics_retained: 10000, // Retain last 10k metrics
        }
    }

    /// Record execution metrics
    pub async fn record_execution(&self, metrics: ExecutionMetrics) {
        // Add to metrics history
        {
            let mut metrics_guard = self.metrics.write().await;
            metrics_guard.push(metrics.clone());

            // Maintain size limit
            if metrics_guard.len() > self.max_metrics_retained {
                let excess = metrics_guard.len() - self.max_metrics_retained;
                metrics_guard.drain(0..excess);
            }
        }

        // Update global statistics
        {
            let mut stats = self.stats.write().await;
            stats.total_executions += 1;
            stats.total_execution_time_ms += metrics.execution_time_ms;
            stats.average_execution_time_ms = stats.total_execution_time_ms as f64 / stats.total_executions as f64;
            stats.total_memory_used_mb += metrics.memory_used_mb;

            if metrics.peak_memory_mb > stats.peak_memory_usage_mb {
                stats.peak_memory_usage_mb = metrics.peak_memory_mb;
            }

            // Update uptime
            stats.uptime_seconds = self.start_time.elapsed().as_secs();
        }

        // Update function-specific tracking
        self.update_function_tracking(&metrics).await;

        // Log significant events
        if let Some(ref logger) = self.logger {
            if metrics.execution_time_ms > 1000 { // Log slow executions
                let correlation_id = logger.generate_correlation_id().await;
                logger.log_with_correlation(
                    crate::logging::LogLevel::Warn,
                    format!(
                        "Slow execution: {}@{} took {}ms",
                        metrics.function_name, metrics.function_version, metrics.execution_time_ms
                    ),
                    &correlation_id,
                );
            }

            if metrics.memory_used_mb > 100.0 { // Log high memory usage
                let correlation_id = logger.generate_correlation_id().await;
                logger.log_with_correlation(
                    crate::logging::LogLevel::Info,
                    format!(
                        "High memory usage: {}@{} used {:.2}MB",
                        metrics.function_name, metrics.function_version, metrics.memory_used_mb
                    ),
                    &correlation_id,
                );
            }
        }
    }

    /// Update per-function tracking
    async fn update_function_tracking(&self, metrics: &ExecutionMetrics) {
        let function_key = format!("{}@{}", metrics.function_name, metrics.function_version);

        let mut limits = self.function_limits.write().await;
        let function_limit = limits.entry(function_key.clone()).or_insert_with(|| FunctionLimits {
            max_memory_mb: 128, // Default
            max_cpu_time_ms: 5000, // Default
            max_concurrent: 10, // Default
            current_concurrent: 0,
            total_executions: 0,
            last_execution: Instant::now(),
        });

        function_limit.total_executions += 1;
        function_limit.last_execution = Instant::now();
    }

    /// Check if function execution should be allowed based on limits
    pub async fn check_limits(&self, function_name: &str, function_version: &str) -> Result<(), String> {
        let function_key = format!("{}@{}", function_name, function_version);

        let limits = self.function_limits.read().await;
        if let Some(function_limit) = limits.get(&function_key) {
            if function_limit.current_concurrent >= function_limit.max_concurrent {
                return Err(format!(
                    "Function {}@{} exceeded concurrent execution limit ({})",
                    function_name, function_version, function_limit.max_concurrent
                ));
            }
        }

        Ok(())
    }

    /// Increment concurrent execution count
    pub async fn increment_concurrent(&self, function_name: &str, function_version: &str) {
        let function_key = format!("{}@{}", function_name, function_version);
        let mut limits = self.function_limits.write().await;
        if let Some(limit) = limits.get_mut(&function_key) {
            limit.current_concurrent += 1;
        }
    }

    /// Decrement concurrent execution count
    pub async fn decrement_concurrent(&self, function_name: &str, function_version: &str) {
        let function_key = format!("{}@{}", function_name, function_version);
        let mut limits = self.function_limits.write().await;
        if let Some(limit) = limits.get_mut(&function_key) {
            if limit.current_concurrent > 0 {
                limit.current_concurrent -= 1;
            }
        }
    }

    /// Get the total number of currently executing functions across all tracked functions.
    ///
    /// This sums `current_concurrent` from all per-function `FunctionLimits` entries,
    /// providing an accurate real-time concurrent execution count (as opposed to
    /// `functions_served` which is a lifetime total).
    pub async fn get_total_concurrent(&self) -> usize {
        let limits = self.function_limits.read().await;
        limits.values().map(|l| l.current_concurrent).sum()
    }

    /// Get current resource statistics
    pub async fn get_stats(&self) -> ResourceStats {
        let mut stats = self.stats.read().await.clone();

        // Calculate cache hit rate and error rate from recent metrics
        let metrics = self.metrics.read().await;
        if !metrics.is_empty() {
            let recent_count = metrics.len().min(100) as f64; // Last 100 executions
            let recent_metrics = &metrics[metrics.len().saturating_sub(100)..];

            let cache_hits = recent_metrics.iter().filter(|m| m.cache_hit).count() as f64;
            let errors = recent_metrics.iter().filter(|m| m.error_occurred).count() as f64;

            stats.cache_hit_rate = (cache_hits / recent_count) * 100.0;
            stats.error_rate = (errors / recent_count) * 100.0;
        }

        // Count unique functions served
        let functions = self.function_limits.read().await;
        stats.functions_served = functions.len();

        stats
    }

    /// Get recent execution metrics
    pub async fn get_recent_metrics(&self, limit: usize) -> Vec<ExecutionMetrics> {
        let metrics = self.metrics.read().await;
        let start = metrics.len().saturating_sub(limit);
        metrics[start..].to_vec()
    }

    /// Generate performance report
    pub async fn generate_report(&self) -> PerformanceReport {
        let stats = self.get_stats().await;
        let recent_metrics = self.get_recent_metrics(100).await;

        // Calculate percentiles
        let mut execution_times: Vec<u64> = recent_metrics.iter().map(|m| m.execution_time_ms).collect();
        execution_times.sort();

        let p50 = percentile(&execution_times, 50.0);
        let p95 = percentile(&execution_times, 95.0);
        let p99 = percentile(&execution_times, 99.0);

        PerformanceReport {
            stats,
            p50_execution_time_ms: p50,
            p95_execution_time_ms: p95,
            p99_execution_time_ms: p99,
            recent_executions: recent_metrics.len(),
        }
    }

    /// Clean up old metrics to prevent memory bloat.
    ///
    /// Removes metrics whose Unix timestamp (`m.timestamp`) is older than 1 hour.
    /// Previously this used `Instant` arithmetic which is semantically incorrect
    /// because `Instant` represents a monotonic clock value, not an absolute time.
    /// We now compare against `SystemTime` (Unix epoch) consistently.
    pub async fn cleanup(&self) {
        let mut metrics = self.metrics.write().await;

        // Calculate the cutoff as a Unix timestamp (seconds since epoch).
        let one_hour_ago = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
            .saturating_sub(3600);

        // Retain only metrics recorded within the last hour.
        metrics.retain(|m| m.timestamp >= one_hour_ago);
    }
}

/// Performance report with percentiles
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceReport {
    pub stats: ResourceStats,
    pub p50_execution_time_ms: u64,
    pub p95_execution_time_ms: u64,
    pub p99_execution_time_ms: u64,
    pub recent_executions: usize,
}

/// Calculate percentile from sorted vector
fn percentile(sorted_data: &[u64], percentile: f64) -> u64 {
    if sorted_data.is_empty() {
        return 0;
    }

    let index = ((percentile / 100.0) * (sorted_data.len() - 1) as f64) as usize;
    sorted_data[index]
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::runtime::Runtime;

    #[test]
    fn test_monitor_creation() {
        let monitor = ResourceMonitor::new(None);
        assert!(monitor.start_time.elapsed() < Duration::from_millis(100));
    }

    #[tokio::test]
    async fn test_record_execution() {
        let monitor = ResourceMonitor::new(None);

        let metrics = ExecutionMetrics {
            function_name: "test".to_string(),
            function_version: "1.0.0".to_string(),
            execution_time_ms: 100,
            cpu_fuel_used: 50000,
            memory_used_mb: 10.5,
            peak_memory_mb: 15.2,
            cache_hit: false,
            cold_start: true,
            error_occurred: false,
            timestamp: 1234567890,
        };

        monitor.record_execution(metrics).await;

        let stats = monitor.get_stats().await;
        assert_eq!(stats.total_executions, 1);
        assert_eq!(stats.total_execution_time_ms, 100);
        assert_eq!(stats.average_execution_time_ms, 100.0);
    }

    #[test]
    fn test_percentile_calculation() {
        let data = vec![10, 20, 30, 40, 50];
        // For 50th percentile (median) of 5 values, should be the 3rd value (index 2)
        assert_eq!(percentile(&data, 50.0), 30);
        // For 95th percentile, using the formula (p/100)*(n-1), we get index 3 (0-based)
        // which gives the 4th value (40)
        assert_eq!(percentile(&data, 95.0), 40);
    }
}