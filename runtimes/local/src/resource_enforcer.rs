//! Enterprise resource limit enforcement.
//!
//! This module provides advanced resource limit enforcement for enterprise deployments,
//! including dynamic quotas, predictive throttling, and sophisticated resource allocation.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};

use crate::config::Config;
use crate::monitoring::{ResourceMonitor, ExecutionMetrics};

/// Enterprise resource quota types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum QuotaType {
    /// CPU time per minute
    CpuTimePerMinute,
    /// Memory usage in MB
    MemoryUsage,
    /// Concurrent executions
    ConcurrentExecutions,
    /// Total executions per hour
    ExecutionsPerHour,
    /// Network bandwidth per minute (MB)
    BandwidthPerMinute,
}

/// Resource quota definition
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResourceQuota {
    pub quota_type: QuotaType,
    pub limit: f64,
    pub window_seconds: u64,
    pub current_usage: f64,
    #[serde(skip, default = "instant_now")]
    pub last_reset: Instant,
}

fn instant_now() -> Instant {
    Instant::now()
}

/// Enterprise resource enforcer
pub struct ResourceEnforcer {
    /// Resource quotas per function
    quotas: Arc<RwLock<HashMap<String, Vec<ResourceQuota>>>>,
    /// Global resource limits
    global_limits: Arc<RwLock<GlobalResourceLimits>>,
    /// Resource monitor reference
    monitor: Arc<ResourceMonitor>,
    /// Configuration (stored for future use)
    #[allow(dead_code)]
    config: Config,
    /// Enforcement policies
    policies: Arc<RwLock<HashMap<String, EnforcementPolicy>>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GlobalResourceLimits {
    pub max_total_memory_mb: usize,
    pub max_total_cpu_percent: f64,
    pub max_concurrent_functions: usize,
    pub max_bandwidth_mbps: f64,
    // --- Per-tenant limits (P4.1 — noisy-neighbour prevention) ---
    /// Maximum memory (MB) that a single tenant may consume across all their
    /// concurrent function instances.
    pub max_memory_per_tenant_mb: usize,
    /// Maximum number of concurrent function instances for a single tenant.
    pub max_concurrent_per_tenant: usize,
    /// Maximum number of function executions a single tenant may start per
    /// minute (rate-limiting).
    pub max_executions_per_tenant_per_minute: usize,
}

impl Default for GlobalResourceLimits {
    fn default() -> Self {
        Self {
            max_total_memory_mb: 3072,          // 3 GB (leave 1 GB for OS on 4 GB node)
            max_total_cpu_percent: 90.0,
            max_concurrent_functions: 200,
            max_bandwidth_mbps: 100.0,
            max_memory_per_tenant_mb: 512,
            max_concurrent_per_tenant: 20,
            max_executions_per_tenant_per_minute: 600,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnforcementPolicy {
    pub name: String,
    pub throttle_threshold_percent: f64,
    pub block_threshold_percent: f64,
    pub predictive_scaling: bool,
    pub priority: u8, // 0 = lowest, 255 = highest
}

/// Enforcement decision
#[derive(Debug, Clone)]
pub enum EnforcementDecision {
    Allow,
    Throttle(Duration),
    Block(String),
}

impl ResourceEnforcer {
    /// Create a new enterprise resource enforcer
    pub fn new(monitor: Arc<ResourceMonitor>, config: Config) -> Self {
        // Derive sensible per-tenant limits from the budget tier
        let tier_specs = crate::budget::NodeSpecs::for_tier(&config.get_budget_tier());
        let global_limits = GlobalResourceLimits {
            max_total_memory_mb: tier_specs.ram_gb * 1024 * 9 / 10, // 90% of RAM
            max_total_cpu_percent: 80.0,
            max_concurrent_functions: tier_specs.max_concurrent_wasm,
            max_bandwidth_mbps: 100.0,
            // Per-tenant: allow each tenant up to 10% of total capacity by default
            max_memory_per_tenant_mb: (tier_specs.ram_gb * 1024 / 10).max(128),
            max_concurrent_per_tenant: (tier_specs.max_concurrent_wasm / 10).max(5),
            max_executions_per_tenant_per_minute: 600,
        };

        let mut policies = HashMap::new();
        policies.insert("default".to_string(), EnforcementPolicy {
            name: "default".to_string(),
            throttle_threshold_percent: 70.0,
            block_threshold_percent: 90.0,
            predictive_scaling: false,
            priority: 100,
        });

        policies.insert("enterprise".to_string(), EnforcementPolicy {
            name: "enterprise".to_string(),
            throttle_threshold_percent: 80.0,
            block_threshold_percent: 95.0,
            predictive_scaling: true,
            priority: 200,
        });

        Self {
            quotas: Arc::new(RwLock::new(HashMap::new())),
            global_limits: Arc::new(RwLock::new(global_limits)),
            monitor,
            config,
            policies: Arc::new(RwLock::new(policies)),
        }
    }

    /// Get the configuration
    pub fn config(&self) -> &Config {
        &self.config
    }

    /// Get all enforcement policies
    pub async fn policies(&self) -> HashMap<String, EnforcementPolicy> {
        self.policies.read().await.clone()
    }

    /// Get a specific policy by name
    pub async fn get_policy(&self, name: &str) -> Option<EnforcementPolicy> {
        self.policies.read().await.get(name).cloned()
    }

    /// Check if a function execution should be allowed
    pub async fn check_execution_allowed(&self, function_key: &str) -> EnforcementDecision {
        // Check global resource limits first
        if let Some(block_reason) = self.check_global_limits().await {
            return EnforcementDecision::Block(block_reason);
        }

        // Check function-specific quotas
        if let Some(throttle_duration) = self.check_function_quotas(function_key).await {
            return EnforcementDecision::Throttle(throttle_duration);
        }

        // Check predictive throttling using policy-based thresholds
        if let Some(throttle_duration) = self.check_predictive_throttling_with_policy(function_key).await {
            return EnforcementDecision::Throttle(throttle_duration);
        }

        // Fallback: check simple predictive throttling if policy-based didn't trigger
        if let Some(throttle_duration) = self.check_predictive_throttling(function_key).await {
            return EnforcementDecision::Throttle(throttle_duration);
        }

        EnforcementDecision::Allow
    }

    /// Check predictive throttling with policy-based thresholds
    async fn check_predictive_throttling_with_policy(&self, function_key: &str) -> Option<Duration> {
        let policies = self.policies.read().await;

        // Get the applicable policy (use default if no specific policy)
        let policy = policies.get("default").cloned().unwrap_or(EnforcementPolicy {
            name: "default".to_string(),
            throttle_threshold_percent: 70.0,
            block_threshold_percent: 90.0,
            predictive_scaling: false,
            priority: 100,
        });

        // Only throttle if predictive scaling is enabled
        if !policy.predictive_scaling {
            return None;
        }

        // Simple predictive throttling based on recent trends
        let recent_metrics = self.monitor.get_recent_metrics(10).await;

        if recent_metrics.len() < 5 {
            return None; // Not enough data for prediction
        }

        // Check if execution times are trending up
        let avg_execution_time = recent_metrics.iter()
            .map(|m| m.execution_time_ms as f64)
            .sum::<f64>() / recent_metrics.len() as f64;

        let recent_avg = recent_metrics.iter()
            .rev()
            .take(3)
            .map(|m| m.execution_time_ms as f64)
            .sum::<f64>() / 3.0;

        let threshold = policy.throttle_threshold_percent;

        // If recent executions are slower than threshold, throttle
        if recent_avg > avg_execution_time * (1.0 + threshold / 100.0) {
            tracing::debug!(
                "Predictive throttle: recent_avg={:.2}ms > avg={:.2}ms * threshold={:.1}% for {}",
                recent_avg, avg_execution_time, threshold, function_key
            );
            return Some(Duration::from_secs(2));
        }

        None
    }

    /// Record resource usage after execution
    pub async fn record_usage(&self, function_key: &str, metrics: &ExecutionMetrics) {
        let mut quotas = self.quotas.write().await;

        let function_quotas = quotas.entry(function_key.to_string()).or_insert_with(|| {
            vec![
                ResourceQuota {
                    quota_type: QuotaType::CpuTimePerMinute,
                    limit: 60000.0, // 60 seconds per minute
                    window_seconds: 60,
                    current_usage: 0.0,
                    last_reset: Instant::now(),
                },
                ResourceQuota {
                    quota_type: QuotaType::MemoryUsage,
                    limit: 512.0, // 512MB
                    window_seconds: 300, // 5 minutes
                    current_usage: 0.0,
                    last_reset: Instant::now(),
                },
                ResourceQuota {
                    quota_type: QuotaType::ExecutionsPerHour,
                    limit: 1000.0, // 1000 executions per hour
                    window_seconds: 3600,
                    current_usage: 0.0,
                    last_reset: Instant::now(),
                },
            ]
        });

        // Update quotas based on metrics
        for quota in function_quotas.iter_mut() {
            // Reset quota if window has passed
            if quota.last_reset.elapsed().as_secs() >= quota.window_seconds {
                quota.current_usage = 0.0;
                quota.last_reset = Instant::now();
            }

            // Add usage based on quota type
            match quota.quota_type {
                QuotaType::CpuTimePerMinute | QuotaType::ExecutionsPerHour => {
                    quota.current_usage += 1.0; // Count executions
                }
                QuotaType::MemoryUsage => {
                    quota.current_usage = quota.current_usage.max(metrics.memory_used_mb);
                }
                _ => {} // Other quota types not implemented yet
            }
        }
    }

    /// Check global resource limits.
    ///
    /// Previously this compared `functions_served` (total unique functions ever
    /// seen, a monotonically increasing counter) against `max_concurrent_functions`,
    /// which would always eventually trigger a false block. We now sum the actual
    /// `current_concurrent` values from per-function tracking to get a real
    /// concurrent execution count.
    async fn check_global_limits(&self) -> Option<String> {
        let report = self.monitor.generate_report().await;
        let global_limits = self.global_limits.read().await;

        // Check total memory usage
        if report.stats.total_memory_used_mb as usize > global_limits.max_total_memory_mb {
            return Some(format!(
                "Global memory limit exceeded: {}MB > {}MB",
                report.stats.total_memory_used_mb, global_limits.max_total_memory_mb
            ));
        }

        // Check actual concurrent execution count (sum of per-function current_concurrent).
        // This replaces the incorrect use of `functions_served` which is a lifetime total,
        // not a concurrent count.
        let actual_concurrent = self.monitor.get_total_concurrent().await;
        if actual_concurrent > global_limits.max_concurrent_functions {
            return Some(format!(
                "Global concurrent functions limit exceeded: {} > {}",
                actual_concurrent, global_limits.max_concurrent_functions
            ));
        }

        None
    }

    /// Check function-specific quotas
    async fn check_function_quotas(&self, function_key: &str) -> Option<Duration> {
        let quotas = self.quotas.read().await;

        if let Some(function_quotas) = quotas.get(function_key) {
            for quota in function_quotas {
                let usage_percent = (quota.current_usage / quota.limit) * 100.0;

                if usage_percent >= 95.0 {
                    // Block if over 95% of quota
                    return Some(Duration::from_secs(30)); // Throttle for 30 seconds
                } else if usage_percent >= 80.0 {
                    // Throttle if over 80% of quota
                    return Some(Duration::from_secs(5));
                }
            }
        }

        None
    }

    /// Check predictive throttling based on usage patterns
    /// Note: function_key reserved for per-function throttling policies
    async fn check_predictive_throttling(&self, _function_key: &str) -> Option<Duration> {
        // Simple predictive throttling based on recent trends
        // In a real implementation, this would use ML models or time-series analysis

        let recent_metrics = self.monitor.get_recent_metrics(10).await;

        if recent_metrics.len() < 5 {
            return None; // Not enough data for prediction
        }

        // Check if execution times are trending up
        let avg_execution_time = recent_metrics.iter()
            .map(|m| m.execution_time_ms as f64)
            .sum::<f64>() / recent_metrics.len() as f64;

        let recent_avg = recent_metrics.iter()
            .rev()
            .take(3)
            .map(|m| m.execution_time_ms as f64)
            .sum::<f64>() / 3.0;

        // If recent executions are 50% slower than average, throttle
        if recent_avg > avg_execution_time * 1.5 {
            return Some(Duration::from_secs(2));
        }

        None
    }

    /// Update global resource limits (admin function)
    pub async fn update_global_limits(&self, limits: GlobalResourceLimits) {
        let mut global_limits = self.global_limits.write().await;
        *global_limits = limits;
    }

    /// Set resource quotas for a function
    pub async fn set_function_quotas(&self, function_key: String, quotas: Vec<ResourceQuota>) {
        let mut all_quotas = self.quotas.write().await;
        all_quotas.insert(function_key, quotas);
    }

    /// Get current resource usage report
    pub async fn get_resource_report(&self) -> ResourceUsageReport {
        let quotas = self.quotas.read().await;
        let global_limits = self.global_limits.read().await;
        let monitor_stats = self.monitor.generate_report().await;

        ResourceUsageReport {
            global_limits: global_limits.clone(),
            function_quotas: quotas.clone(),
            current_stats: monitor_stats.stats,
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
        }
    }
}

/// Resource usage report for monitoring
#[derive(Debug, Clone)]
pub struct ResourceUsageReport {
    pub global_limits: GlobalResourceLimits,
    #[allow(dead_code)]
    pub function_quotas: HashMap<String, Vec<ResourceQuota>>,
    pub current_stats: crate::monitoring::ResourceStats,
    pub timestamp: u64,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::monitoring::ResourceMonitor;
    use crate::logging::init_structured_logging;
    use std::sync::Arc;

    #[tokio::test]
    async fn test_resource_enforcer_creation() {
        let logger = Arc::new(init_structured_logging(false));
        let monitor = Arc::new(ResourceMonitor::new(Some(logger)));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);
        let report = enforcer.get_resource_report().await;

        assert!(report.function_quotas.is_empty());
        assert_eq!(report.global_limits.max_total_memory_mb, 8192);
    }

    #[tokio::test]
    async fn test_execution_allowed() {
        let logger = Arc::new(init_structured_logging(false));
        let monitor = Arc::new(ResourceMonitor::new(Some(logger)));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);
        let decision = enforcer.check_execution_allowed("test@1.0.0").await;

        match decision {
            EnforcementDecision::Allow => {}, // Expected for empty system
            _ => panic!("Expected Allow decision"),
        }
    }
}
