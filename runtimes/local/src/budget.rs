//! Budget optimization utilities for low-cost node deployments.
//!
//! This module provides utilities to optimize FunctionFly for $5-10 monthly
//! node budgets by calculating optimal resource allocation and providing
//! recommendations for cost-effective scaling.

use serde::{Deserialize, Serialize};

/// Budget tier definitions based on monthly cost
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum BudgetTier {
    /// $5-10/month nodes (target tier)
    UltraLow,
    /// $10-20/month nodes
    Low,
    /// $20-50/month nodes
    Medium,
    /// $50+/month nodes
    High,
}

/// Hardware specifications for different budget tiers
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeSpecs {
    pub monthly_cost: f64,
    pub vcpu: usize,
    pub ram_gb: usize,
    pub storage_gb: usize,
    pub bandwidth_gbps: f64,
    /// Maximum number of concurrent Wasm function instances on this node.
    pub max_concurrent_wasm: usize,
    /// Maximum memory per individual function instance in MB.
    pub max_memory_per_fn_mb: usize,
    /// AOT compilation cache budget in MB.
    pub aot_cache_mb: usize,
    /// Whether this tier supports Firecracker MicroVM execution.
    pub supports_firecracker: bool,
}

impl NodeSpecs {
    /// Get recommended node specs for a budget tier
    pub fn for_tier(tier: &BudgetTier) -> Self {
        match tier {
            BudgetTier::UltraLow => NodeSpecs {
                monthly_cost: 7.5, // Average of $5-10
                vcpu: 2,
                ram_gb: 4,
                storage_gb: 75,
                bandwidth_gbps: 0.2,
                max_concurrent_wasm: 200,
                max_memory_per_fn_mb: 64,
                aot_cache_mb: 256,
                supports_firecracker: false,
            },
            BudgetTier::Low => NodeSpecs {
                monthly_cost: 15.0,
                vcpu: 4,
                ram_gb: 8,
                storage_gb: 150,
                bandwidth_gbps: 0.4,
                max_concurrent_wasm: 400,
                max_memory_per_fn_mb: 128,
                aot_cache_mb: 512,
                supports_firecracker: false,
            },
            BudgetTier::Medium => NodeSpecs {
                monthly_cost: 35.0,
                vcpu: 8,
                ram_gb: 16,
                storage_gb: 300,
                bandwidth_gbps: 1.0,
                max_concurrent_wasm: 800,
                max_memory_per_fn_mb: 256,
                aot_cache_mb: 1024,
                supports_firecracker: true,
            },
            BudgetTier::High => NodeSpecs {
                monthly_cost: 75.0,
                vcpu: 16,
                ram_gb: 32,
                storage_gb: 500,
                bandwidth_gbps: 2.0,
                max_concurrent_wasm: 2000,
                max_memory_per_fn_mb: 512,
                aot_cache_mb: 2048,
                supports_firecracker: true,
            },
        }
    }
}

/// Function resource profile based on runtime and usage patterns
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionProfile {
    pub runtime: String,
    pub avg_execution_time_ms: u64,
    pub avg_memory_mb: f64,
    pub cold_start_time_ms: u64,
    pub requests_per_minute: f64,
    pub cache_hit_rate: f64,
}

impl Default for FunctionProfile {
    fn default() -> Self {
        Self {
            runtime: "wasm".to_string(),
            avg_execution_time_ms: 100,
            avg_memory_mb: 32.0,
            cold_start_time_ms: 50,
            requests_per_minute: 10.0,
            cache_hit_rate: 0.3, // 30% cache hit rate
        }
    }
}

/// Cost optimization calculator
pub struct BudgetOptimizer {
    node_specs: NodeSpecs,
}

impl BudgetOptimizer {
    /// Create optimizer for specific node specs
    pub fn new(node_specs: NodeSpecs) -> Self {
        Self { node_specs }
    }

    /// Create optimizer for budget tier
    pub fn for_tier(tier: &BudgetTier) -> Self {
        Self::new(NodeSpecs::for_tier(tier))
    }

    /// Calculate maximum concurrent functions for given profile
    pub fn max_concurrent_functions(&self, profile: &FunctionProfile) -> usize {
        let memory_capacity = (self.node_specs.ram_gb as f64 * 1024.0) - 512.0; // Reserve 512MB for system
        let memory_per_function = profile.avg_memory_mb * 1.5; // 50% overhead for safety
        let memory_based_limit = (memory_capacity / memory_per_function) as usize;

        // CPU-based limit (rough estimate: 10ms execution time = 100 concurrent)
        let cpu_based_limit = ((self.node_specs.vcpu as f64 * 100.0)
            / (profile.avg_execution_time_ms as f64 / 10.0)) as usize;

        memory_based_limit.min(cpu_based_limit).max(1)
    }

    /// Calculate optimal instance pool size
    pub fn optimal_pool_size(&self, profile: &FunctionProfile, function_count: usize) -> usize {
        let max_concurrent = self.max_concurrent_functions(profile);
        let pool_per_function = (max_concurrent / function_count).clamp(1, 5); // 1-5 instances per function

        // Consider cold start impact
        let cold_start_penalty =
            profile.cold_start_time_ms as f64 / profile.avg_execution_time_ms as f64;
        let adjusted_pool = (pool_per_function as f64 * (1.0 + cold_start_penalty * 0.1)) as usize;

        adjusted_pool.min(max_concurrent)
    }

    /// Calculate cost per 1000 executions
    pub fn cost_per_thousand_executions(&self, profile: &FunctionProfile) -> f64 {
        let daily_executions = profile.requests_per_minute * 60.0 * 24.0;
        let monthly_executions = daily_executions * 30.0;
        let executions_per_thousand = monthly_executions / 1000.0;

        self.node_specs.monthly_cost / executions_per_thousand
    }

    /// Generate budget optimization recommendations
    pub fn generate_recommendations(&self, functions: &[FunctionProfile]) -> BudgetRecommendations {
        let total_functions = functions.len();
        let avg_profile = self.average_profile(functions);

        let max_concurrent = self.max_concurrent_functions(&avg_profile);
        let optimal_pool = self.optimal_pool_size(&avg_profile, total_functions);
        let cost_per_1k = self.cost_per_thousand_executions(&avg_profile);

        // Calculate efficiency metrics
        let memory_efficiency = self.calculate_memory_efficiency(&avg_profile);
        let cpu_efficiency = self.calculate_cpu_efficiency(&avg_profile);

        BudgetRecommendations {
            tier: self.detect_tier(),
            max_concurrent_functions: max_concurrent,
            recommended_pool_size: optimal_pool,
            cost_per_1000_executions: cost_per_1k,
            memory_efficiency_percent: memory_efficiency,
            cpu_efficiency_percent: cpu_efficiency,
            suggestions: self.generate_suggestions(&avg_profile, memory_efficiency, cpu_efficiency),
            node_specs: self.node_specs.clone(),
        }
    }

    /// Calculate average function profile
    fn average_profile(&self, functions: &[FunctionProfile]) -> FunctionProfile {
        if functions.is_empty() {
            return FunctionProfile::default();
        }

        let total_execution_time: u64 = functions.iter().map(|f| f.avg_execution_time_ms).sum();
        let total_memory: f64 = functions.iter().map(|f| f.avg_memory_mb).sum();
        let total_cold_start: u64 = functions.iter().map(|f| f.cold_start_time_ms).sum();
        let total_rpm: f64 = functions.iter().map(|f| f.requests_per_minute).sum();
        let total_cache_hit: f64 = functions.iter().map(|f| f.cache_hit_rate).sum();

        let count = functions.len() as f64;
        FunctionProfile {
            runtime: "mixed".to_string(),
            avg_execution_time_ms: (total_execution_time as f64 / count) as u64,
            avg_memory_mb: total_memory / count,
            cold_start_time_ms: (total_cold_start as f64 / count) as u64,
            requests_per_minute: total_rpm / count,
            cache_hit_rate: total_cache_hit / count,
        }
    }

    /// Calculate memory efficiency percentage
    fn calculate_memory_efficiency(&self, profile: &FunctionProfile) -> f64 {
        let memory_capacity = self.node_specs.ram_gb as f64 * 1024.0; // Convert to MB
        let system_reserved = 512.0; // MB
        let available_memory = memory_capacity - system_reserved;

        // Estimate memory overhead (WASM, pooling, etc.)
        let memory_overhead = profile.avg_memory_mb * 0.3; // 30% overhead
        let effective_memory_per_function = profile.avg_memory_mb + memory_overhead;

        let max_functions = available_memory / effective_memory_per_function;
        let efficiency = (profile.requests_per_minute / max_functions).min(1.0) * 100.0;

        efficiency.min(100.0)
    }

    /// Calculate CPU efficiency percentage
    fn calculate_cpu_efficiency(&self, profile: &FunctionProfile) -> f64 {
        let cpu_capacity_milliseconds = self.node_specs.vcpu as f64 * 1000.0; // 1000ms per second per vCPU
        let function_cpu_time =
            profile.avg_execution_time_ms as f64 * profile.requests_per_minute * 60.0; // per minute

        let utilization = function_cpu_time / cpu_capacity_milliseconds;
        (utilization * 100.0).min(100.0)
    }

    /// Generate optimization suggestions
    fn generate_suggestions(
        &self,
        profile: &FunctionProfile,
        memory_eff: f64,
        cpu_eff: f64,
    ) -> Vec<String> {
        let mut suggestions = Vec::new();

        if memory_eff > 80.0 {
            suggestions.push("High memory utilization - consider increasing cache hit rates or optimizing function memory usage".to_string());
        }

        if cpu_eff > 80.0 {
            suggestions.push("High CPU utilization - consider optimizing function execution time or increasing instance pooling".to_string());
        }

        if profile.cache_hit_rate < 0.5 {
            suggestions.push(
                "Low cache hit rate detected - enable deterministic caching for better performance"
                    .to_string(),
            );
        }

        if profile.avg_memory_mb > 64.0 {
            suggestions.push("Functions using high memory - consider optimizing memory usage or increasing node RAM".to_string());
        }

        if profile.avg_execution_time_ms > 500 {
            suggestions.push(
                "Slow function execution detected - optimize code or consider different runtime"
                    .to_string(),
            );
        }

        if suggestions.is_empty() {
            suggestions.push("Configuration looks optimal for current workload".to_string());
        }

        suggestions
    }

    /// Detect current budget tier
    fn detect_tier(&self) -> BudgetTier {
        if self.node_specs.monthly_cost <= 10.0 {
            BudgetTier::UltraLow
        } else if self.node_specs.monthly_cost <= 20.0 {
            BudgetTier::Low
        } else if self.node_specs.monthly_cost <= 50.0 {
            BudgetTier::Medium
        } else {
            BudgetTier::High
        }
    }
}

/// Budget optimization recommendations
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BudgetRecommendations {
    pub tier: BudgetTier,
    pub max_concurrent_functions: usize,
    pub recommended_pool_size: usize,
    pub cost_per_1000_executions: f64,
    pub memory_efficiency_percent: f64,
    pub cpu_efficiency_percent: f64,
    pub suggestions: Vec<String>,
    pub node_specs: NodeSpecs,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ultra_low_tier_specs() {
        let specs = NodeSpecs::for_tier(&BudgetTier::UltraLow);
        assert_eq!(specs.vcpu, 2);
        assert_eq!(specs.ram_gb, 4);
        assert!(specs.monthly_cost >= 5.0 && specs.monthly_cost <= 10.0);
    }

    #[test]
    fn test_budget_optimizer() {
        let optimizer = BudgetOptimizer::for_tier(&BudgetTier::UltraLow);
        let profile = FunctionProfile::default();

        let max_concurrent = optimizer.max_concurrent_functions(&profile);
        assert!(max_concurrent > 0);

        let recommendations = optimizer.generate_recommendations(&vec![profile]);
        assert!(!recommendations.suggestions.is_empty());
    }
}
