//! Pool statistics and metrics.

/// Statistics about a WASM instance pool.
#[derive(Debug, Clone)]
pub struct WasmPoolStats {
    pub idle_count: usize,
    pub max_idle: usize,
    pub available_permits: usize,
    pub max_concurrent: usize,
    pub function_key: String,
}

impl std::fmt::Display for WasmPoolStats {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "WasmPoolStats {{ {}: idle={}/{}, permits={}/{} }}",
            self.function_key,
            self.idle_count,
            self.max_idle,
            self.available_permits,
            self.max_concurrent
        )
    }
}

/// Legacy pool statistics with memory information.
#[derive(Debug, Clone)]
pub struct PoolStats {
    pub total_instances: usize,
    pub functions_in_pool: usize,
    pub max_per_function: usize,
    pub idle_timeout_secs: u64,
    pub current_memory_usage_mb: f64,
    pub max_memory_usage_mb: f64,
    pub memory_pressure_percent: f64,
}
