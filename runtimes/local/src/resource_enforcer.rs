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
        let recent_metrics = self.monitor.get_recent_metrics(20).await;

        if recent_metrics.len() < 5 {
            return None;
        }

        let execution_times: Vec<f64> = recent_metrics.iter()
            .map(|m| m.execution_time_ms as f64)
            .collect();

        let timestamps: Vec<f64> = recent_metrics.iter()
            .map(|m| m.timestamp as f64)
            .collect();

        let n = execution_times.len() as f64;

        // Exponential Moving Average (alpha=0.3 weights recent data more)
        let alpha = 0.3;
        let mut ema = execution_times[0];
        for &et in &execution_times[1..] {
            ema = alpha * et + (1.0 - alpha) * ema;
        }

        // Linear regression for trend slope
        let sum_t = timestamps.iter().sum::<f64>();
        let sum_et = execution_times.iter().sum::<f64>();
        let sum_t_et: f64 = timestamps.iter().zip(execution_times.iter())
            .map(|(t, et)| t * et).sum::<f64>();
        let sum_t_sq: f64 = timestamps.iter().map(|t| t * t).sum::<f64>();

        let slope = if (n * sum_t_sq - sum_t * sum_t).abs() > f64::EPSILON {
            (n * sum_t_et - sum_t * sum_et) / (n * sum_t_sq - sum_t * sum_t)
        } else {
            0.0
        };

        // Normalize slope to per-millisecond for readability
        let time_range = timestamps.last().copied().unwrap_or(0.0) - timestamps.first().copied().unwrap_or(0.0);
        let slope_per_ms = if time_range > 0.0 { slope / time_range } else { 0.0 };

        // Standard deviation for anomaly detection
        let mean = sum_et / n;
        let variance = execution_times.iter()
            .map(|et| (et - mean).powi(2)).sum::<f64>() / n;
        let std_dev = variance.sqrt();

        // Throttle condition 1: Strong upward trend with elevated EMA
        if slope_per_ms > 0.001 && ema > mean * 1.2 {
            return Some(Duration::from_secs(5));
        }

        // Throttle condition 2: Anomaly — recent spike beyond 2σ
        let recent_avg = execution_times.iter().rev().take(3).sum::<f64>() / 3.0;
        if recent_avg > mean + 2.0 * std_dev {
            return Some(Duration::from_secs(3));
        }

        // Throttle condition 3: Sustained degradation (recent avg >> overall avg)
        let overall_avg = execution_times.iter().sum::<f64>() / execution_times.len() as f64;
        if recent_avg > overall_avg * 1.5 {
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
        let monitor = Arc::new(ResourceMonitor::new(Some(logger), None));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);
        let report = enforcer.get_resource_report().await;

        assert!(report.function_quotas.is_empty());
        assert_eq!(report.global_limits.max_total_memory_mb, 3686);
    }

    #[tokio::test]
    async fn test_execution_allowed() {
        let logger = Arc::new(init_structured_logging(false));
        let monitor = Arc::new(ResourceMonitor::new(Some(logger), None));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);
        let decision = enforcer.check_execution_allowed("test@1.0.0").await;

        match decision {
            EnforcementDecision::Allow => {}, // Expected for empty system
            _ => panic!("Expected Allow decision"),
        }
    }

    #[tokio::test]
    async fn test_enforcement_decision_display() {
        // Test that decision variants can be cloned and compared
        let allow = EnforcementDecision::Allow;
        let _throttle = EnforcementDecision::Throttle(Duration::from_secs(5));
        let block = EnforcementDecision::Block("test reason".to_string());

        // Clone should work
        let allow_clone = allow.clone();
        assert!(matches!(allow_clone, EnforcementDecision::Allow));

        let block_clone = block.clone();
        assert!(matches!(block_clone, EnforcementDecision::Block(ref msg) if msg == "test reason"));
    }

    #[tokio::test]
    async fn test_global_resource_limits_default() {
        let limits = GlobalResourceLimits::default();
        assert_eq!(limits.max_total_memory_mb, 3072);
        assert_eq!(limits.max_total_cpu_percent, 90.0);
        assert_eq!(limits.max_concurrent_functions, 200);
        assert_eq!(limits.max_memory_per_tenant_mb, 512);
        assert_eq!(limits.max_concurrent_per_tenant, 20);
        assert_eq!(limits.max_executions_per_tenant_per_minute, 600);
    }

    #[tokio::test]
    async fn test_resource_quota_creation() {
        let quota = ResourceQuota {
            quota_type: QuotaType::CpuTimePerMinute,
            limit: 60000.0,
            window_seconds: 60,
            current_usage: 0.0,
            last_reset: Instant::now(),
        };

        assert!(matches!(quota.quota_type, QuotaType::CpuTimePerMinute));
        assert_eq!(quota.limit, 60000.0);
        assert!(quota.current_usage < quota.limit);
    }

    #[tokio::test]
    async fn test_enforcement_policy_creation() {
        let policy = EnforcementPolicy {
            name: "test_policy".to_string(),
            throttle_threshold_percent: 75.0,
            block_threshold_percent: 90.0,
            predictive_scaling: true,
            priority: 150,
        };

        assert_eq!(policy.name, "test_policy");
        assert_eq!(policy.throttle_threshold_percent, 75.0);
        assert!(policy.predictive_scaling);
        assert_eq!(policy.priority, 150);
    }

    #[tokio::test]
    async fn test_quota_type_variants() {
        // Test all quota types can be created and used
        let cpu_quota = QuotaType::CpuTimePerMinute;
        let mem_quota = QuotaType::MemoryUsage;
        let concurrent_quota = QuotaType::ConcurrentExecutions;
        let hourly_quota = QuotaType::ExecutionsPerHour;
        let bw_quota = QuotaType::BandwidthPerMinute;

        // Verify they are different variants
        assert_ne!(format!("{:?}", cpu_quota), format!("{:?}", mem_quota));
        assert_ne!(format!("{:?}", concurrent_quota), format!("{:?}", hourly_quota));
        let _ = bw_quota;
    }

    #[tokio::test]
    async fn test_update_global_limits() {
        let logger = Arc::new(init_structured_logging(false));
        let monitor = Arc::new(ResourceMonitor::new(Some(logger), None));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);

        // Update limits
        let new_limits = GlobalResourceLimits {
            max_total_memory_mb: 4096,
            max_total_cpu_percent: 85.0,
            max_concurrent_functions: 150,
            max_bandwidth_mbps: 50.0,
            max_memory_per_tenant_mb: 256,
            max_concurrent_per_tenant: 10,
            max_executions_per_tenant_per_minute: 300,
        };

        enforcer.update_global_limits(new_limits.clone()).await;

        let report = enforcer.get_resource_report().await;
        assert_eq!(report.global_limits.max_total_memory_mb, 4096);
        assert_eq!(report.global_limits.max_total_cpu_percent, 85.0);
        assert_eq!(report.global_limits.max_concurrent_functions, 150);
    }

    #[tokio::test]
    async fn test_set_and_get_function_quotas() {
        let logger = Arc::new(init_structured_logging(false));
        let monitor = Arc::new(ResourceMonitor::new(Some(logger), None));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);

        let quotas = vec![
            ResourceQuota {
                quota_type: QuotaType::CpuTimePerMinute,
                limit: 30000.0,
                window_seconds: 60,
                current_usage: 0.0,
                last_reset: Instant::now(),
            },
        ];

        enforcer.set_function_quotas("test-fn@1.0.0".to_string(), quotas).await;

        let report = enforcer.get_resource_report().await;
        assert!(report.function_quotas.contains_key("test-fn@1.0.0"));
    }

    #[tokio::test]
    async fn test_record_usage_updates_quotas() {
        let logger = Arc::new(init_structured_logging(false));
        let monitor = Arc::new(ResourceMonitor::new(Some(logger), None));
        let config = Config::default();

        let enforcer = ResourceEnforcer::new(monitor, config);

        // Record usage for a function
        let metrics = ExecutionMetrics {
            function_name: "test".to_string(),
            function_version: "1.0.0".to_string(),
            execution_time_ms: 100,
            cpu_fuel_used: 50000,
            memory_used_mb: 10.5,
            peak_memory_mb: 15.2,
            cache_hit: false,
            cold_start: false,
            error_occurred: false,
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
        };

        enforcer.record_usage("test@1.0.0", &metrics).await;

        // Verify quotas were created for this function
        let report = enforcer.get_resource_report().await;
        assert!(report.function_quotas.contains_key("test@1.0.0"));
    }
}
