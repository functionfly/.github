//! Configuration for the local runtime.

use clap::Parser;

use crate::budget::BudgetTier;

/// Runtime configuration
#[derive(Parser, Debug, Clone)]
#[command(name = "functionfly-local")]
#[command(about = "FunctionFly local development runtime", long_about = None)]
pub struct Config {
    /// Port to listen on
    #[arg(short, long, default_value = "8787")]
    pub port: u16,

    /// Function name
    #[arg(short, long, default_value = "function")]
    pub function: String,

    /// Function version
    #[arg(short, long, default_value = "1.0.0")]
    pub version: String,

    /// Wasm file to load
    #[arg(short, long)]
    pub wasm: Option<String>,

    /// Runtime type (nodejs, python)
    #[arg(long, default_value = "nodejs")]
    pub runtime: String,

    /// Memory limit in MB
    #[arg(long, default_value = "128")]
    pub memory_mb: u32,

    /// Timeout in milliseconds
    #[arg(long, default_value = "5000")]
    pub timeout_ms: u64,

    /// Enable deterministic caching
    #[arg(long, default_value = "false")]
    pub deterministic: bool,

    /// Cache TTL in seconds
    #[arg(long, default_value = "3600")]
    pub cache_ttl: u64,

    /// Enable verbose logging
    #[arg(long, default_value = "false")]
    pub verbose: bool,

    /// Enable WASI support
    #[arg(long, default_value = "true")]
    pub wasi_enabled: bool,

    /// CPU fuel limit per execution (0 = unlimited, use with caution)
    #[arg(long, default_value = "1000000")]
    pub cpu_fuel_limit: u64,

    /// Maximum CPU time per function execution in milliseconds
    #[arg(long, default_value = "5000")]
    pub max_cpu_time_ms: u64,

    /// Enable resource monitoring and profiling
    #[arg(long, default_value = "true")]
    pub enable_monitoring: bool,

    /// Enable attack surface reduction (stricter isolation)
    #[arg(long, default_value = "true")]
    pub hardened_security: bool,

    /// Maximum concurrent executions per function
    #[arg(long, default_value = "10")]
    pub max_concurrent_per_function: usize,

    /// Memory overhead allowance (percentage added to declared memory)
    #[arg(long, default_value = "10")]
    pub memory_overhead_percent: u8,

    /// Preopened directories for WASI filesystem access (format: host_path:wasm_path:permissions)
    #[arg(long)]
    pub wasi_dirs: Vec<String>,

    /// Environment variables to expose to WASI (format: KEY=VALUE)
    #[arg(long)]
    pub wasi_env: Vec<String>,

    /// Command line arguments to pass to WASI module
    #[arg(long)]
    pub wasi_args: Vec<String>,

    /// Allow network access in WASI
    #[arg(long, default_value = "false")]
    pub wasi_allow_network: bool,

    /// Allow system time access in WASI
    #[arg(long, default_value = "true")]
    pub wasi_allow_time: bool,

    /// Python runtime version (for Python functions)
    #[arg(long, default_value = "rustpython-0.4")]
    pub python_runtime: String,

    /// Capabilities declared in function manifest (determines what bindings are exposed)
    /// Format: capability1,capability2,...
    /// Example: fetch:read,crypto,cache:write
    #[arg(long, default_value = "")]
    pub capabilities: String,

    /// Python stdlib packages to include
    #[arg(long)]
    pub python_packages: Vec<String>,

    /// Enable Python debugging output
    #[arg(long, default_value = "false")]
    pub python_debug: bool,

    /// SMTP server hostname for email capability
    #[arg(long, default_value = "localhost")]
    pub smtp_host: String,

    /// SMTP server port for email capability
    #[arg(long, default_value = "587")]
    pub smtp_port: u16,

    /// SMTP username for email capability
    #[arg(long)]
    pub smtp_username: Option<String>,

    /// SMTP password for email capability
    #[arg(long)]
    pub smtp_password: Option<String>,

    /// Storage base directory for file operations
    #[arg(long, default_value = "./storage")]
    pub storage_base_dir: String,

    /// AI models directory
    #[arg(long, default_value = "./models")]
    pub ai_models_dir: String,

    /// External API rate limit (requests per minute)
    #[arg(long, default_value = "60")]
    pub external_api_rate_limit: u32,

    /// External API timeout in seconds
    #[arg(long, default_value = "30")]
    pub external_api_timeout_secs: u64,

    /// MicroVM orchestrator URL (for Enterprise tier)
    #[arg(long, default_value = "http://localhost:8080")]
    pub orchestrator_url: String,

    /// MicroVM orchestrator timeout in seconds
    #[arg(long, default_value = "60")]
    pub orchestrator_timeout_secs: u64,

    /// Enable Enterprise features (MicroVM support)
    #[arg(long, default_value = "false")]
    pub enterprise_enabled: bool,

    /// Budget tier (ultra-low, low, medium, high)
    #[arg(long, default_value = "ultra-low")]
    pub tier: String,

    /// Network whitelist (comma-separated list of allowed domains/IPs)
    /// Example: api.example.com,1.2.3.4,*.trusted.com
    #[arg(long)]
    pub network_whitelist: Vec<String>,

    /// Enable strict network whitelist enforcement (Enterprise feature)
    #[arg(long, default_value = "false")]
    pub strict_network_whitelist: bool,

    /// Enable package caching (Enterprise feature)
    #[arg(long, default_value = "false")]
    pub package_caching_enabled: bool,

    /// Package cache directory
    #[arg(long, default_value = "./package-cache")]
    pub package_cache_dir: String,

    /// Maximum package cache size in MB
    #[arg(long, default_value = "1024")]
    pub package_cache_size_mb: usize,

    /// Maximum size of the stdout/stderr capture pipe in bytes.
    /// Functions that produce output larger than this will have their output
    /// silently truncated.  Defaults to 1 MiB.
    #[arg(long, default_value = "1048576")]
    pub max_output_bytes: usize,

    /// Maximum allowed input size in bytes.
    /// Requests with a body larger than this are rejected before execution.
    /// Defaults to 1 MiB (1048576 bytes).
    #[arg(long, default_value = "1048576")]
    pub max_input_bytes: usize,

    /// Allow silent fallback from MicroVM to RustPython when the orchestrator
    /// is unavailable.  When false, execution fails fast instead of silently
    /// degrading to a different Python runtime.
    #[arg(long, default_value = "true")]
    pub microvm_fallback_allowed: bool,

    // -------------------------------------------------------------------------
    // Phase 1: AOT module cache
    // -------------------------------------------------------------------------

    /// Enable AOT (Ahead-of-Time) compiled module cache.
    ///
    /// When enabled, compiled Wasmtime modules are serialised to disk so that
    /// subsequent process restarts skip JIT compilation.
    #[arg(long, default_value = "true")]
    pub aot_cache_enabled: bool,

    /// Directory where AOT-compiled `.cwasm` files are stored.
    #[arg(long, default_value = "./module-cache")]
    pub aot_cache_dir: String,

    /// Maximum number of compiled modules to keep in the in-memory LRU cache.
    #[arg(long, default_value = "64")]
    pub aot_cache_memory_capacity: usize,

    // -------------------------------------------------------------------------
    // Phase 1: CPU millisecond limit (fuel calibration abstraction)
    // -------------------------------------------------------------------------

    /// CPU time limit expressed in **milliseconds** of wall-clock CPU time.
    ///
    /// When non-zero this overrides `cpu_fuel_limit`: the runtime converts the
    /// millisecond budget to Wasmtime fuel units using the `fuel_per_ms`
    /// calibration constant.  Set to 0 to use `cpu_fuel_limit` directly.
    #[arg(long, default_value = "0")]
    pub cpu_ms_limit: u64,

    /// Fuel units consumed per millisecond of CPU time on this hardware class.
    ///
    /// Run the calibration benchmark (`functionfly-local --calibrate`) to
    /// determine the correct value for your deployment.  The default of 200_000
    /// is a conservative estimate suitable for most cloud VMs.
    #[arg(long, default_value = "200000")]
    pub fuel_per_ms: u64,

    // -------------------------------------------------------------------------
    // Phase 2: YARA scanner
    // -------------------------------------------------------------------------

    /// Enable YARA scanning of WASM artifacts before execution.
    #[arg(long, default_value = "false")]
    pub yara_scan_enabled: bool,

    /// URL of the YARA service (e.g. `http://localhost:5000`).
    #[arg(long, default_value = "http://localhost:5000")]
    pub yara_service_url: String,

    /// YARA scanner timeout in seconds.
    #[arg(long, default_value = "5")]
    pub yara_timeout_secs: u64,

    /// If true, allow execution when the YARA service is unreachable (fail-open).
    /// If false, block execution when the service is unreachable (fail-closed).
    #[arg(long, default_value = "true")]
    pub yara_fail_open: bool,

    // -------------------------------------------------------------------------
    // Phase 3: Python runtime pool
    // -------------------------------------------------------------------------

    /// Maximum number of concurrent Python (RustPython) runtimes.
    #[arg(long, default_value = "8")]
    pub python_pool_max_concurrent: usize,

    /// Maximum number of idle Python runtimes to keep warm in the pool.
    #[arg(long, default_value = "4")]
    pub python_pool_max_idle: usize,
}

impl Config {
    /// Get the function key (name@version)
    pub fn function_key(&self) -> String {
        format!("{}@{}", self.function, self.version)
    }

    /// Get the budget tier enum
    pub fn get_budget_tier(&self) -> BudgetTier {
        match self.tier.to_lowercase().as_str() {
            "ultra-low" | "ultralow" => BudgetTier::UltraLow,
            "low" => BudgetTier::Low,
            "medium" => BudgetTier::Medium,
            "high" => BudgetTier::High,
            _ => {
                eprintln!("Warning: Unknown tier '{}', defaulting to UltraLow", self.tier);
                BudgetTier::UltraLow
            }
        }
    }

    /// Validate configuration consistency
    pub fn validate(&self) -> Result<(), String> {
        // Enterprise features require appropriate tier
        if self.enterprise_enabled {
            match self.get_budget_tier() {
                BudgetTier::UltraLow | BudgetTier::Low => {
                    return Err(format!(
                        "Enterprise features (MicroVM) are not available for {} tier. Upgrade to Medium or High tier.",
                        self.tier
                    ));
                }
                BudgetTier::Medium | BudgetTier::High => {
                    // Enterprise features allowed for these tiers
                }
            }
        }

        // Validate resource limits based on tier
        let tier_specs = crate::budget::NodeSpecs::for_tier(&self.get_budget_tier());
        if self.memory_mb as usize > tier_specs.ram_gb * 1024 {
            return Err(format!(
                "Memory limit {}MB exceeds {} tier maximum of {}MB",
                self.memory_mb, self.tier, tier_specs.ram_gb * 1024
            ));
        }

        Ok(())
    }

    /// Check if this tier supports MicroVM execution
    pub fn supports_microvm(&self) -> bool {
        self.enterprise_enabled && matches!(self.get_budget_tier(), BudgetTier::High | BudgetTier::Medium)
    }
}

impl Default for Config {
    fn default() -> Self {
        Self {
            port: 8787,
            function: "function".to_string(),
            version: "1.0.0".to_string(),
            wasm: None,
            runtime: "nodejs".to_string(),
            memory_mb: 128,
            timeout_ms: 5000,
            deterministic: false,
            cache_ttl: 3600,
            verbose: false,
            wasi_enabled: true,
            cpu_fuel_limit: 1000000,
            max_cpu_time_ms: 5000,
            enable_monitoring: true,
            hardened_security: true,
            max_concurrent_per_function: 10,
            memory_overhead_percent: 10,
            wasi_dirs: Vec::new(),
            wasi_env: Vec::new(),
            wasi_args: Vec::new(),
            wasi_allow_network: false,
            wasi_allow_time: true,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: Vec::new(),
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: "http://localhost:8080".to_string(),
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tier: "ultra-low".to_string(),
            network_whitelist: Vec::new(),
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
            max_output_bytes: 1024 * 1024,   // 1 MiB
            max_input_bytes: 1024 * 1024,    // 1 MiB
            microvm_fallback_allowed: true,
            // Phase 1: AOT cache
            aot_cache_enabled: true,
            aot_cache_dir: "./module-cache".to_string(),
            aot_cache_memory_capacity: 64,
            // Phase 1: CPU ms limit
            cpu_ms_limit: 0,
            fuel_per_ms: 200_000,
            // Phase 2: YARA scanner
            yara_scan_enabled: false,
            yara_service_url: "http://localhost:5000".to_string(),
            yara_timeout_secs: 5,
            yara_fail_open: true,
            // Phase 3: Python pool
            python_pool_max_concurrent: 8,
            python_pool_max_idle: 4,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_function_key() {
        let config = Config {
            port: 8787,
            function: "slugify".to_string(),
            version: "1.0.0".to_string(),
            wasm: None,
            runtime: "nodejs".to_string(),
            memory_mb: 128,
            timeout_ms: 5000,
            deterministic: false,
            cache_ttl: 3600,
            verbose: false,
            wasi_enabled: true,
            cpu_fuel_limit: 1000000,
            max_cpu_time_ms: 5000,
            enable_monitoring: true,
            hardened_security: true,
            max_concurrent_per_function: 10,
            memory_overhead_percent: 10,
            wasi_dirs: Vec::new(),
            wasi_env: Vec::new(),
            wasi_args: Vec::new(),
            wasi_allow_network: false,
            wasi_allow_time: true,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: Vec::new(),
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: "http://localhost:8080".to_string(),
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tier: "ultra-low".to_string(),
            network_whitelist: Vec::new(),
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
            max_output_bytes: 1024 * 1024,
            max_input_bytes: 1024 * 1024,
            microvm_fallback_allowed: true,
            aot_cache_enabled: true,
            aot_cache_dir: "./module-cache".to_string(),
            aot_cache_memory_capacity: 64,
            cpu_ms_limit: 0,
            fuel_per_ms: 200_000,
            yara_scan_enabled: false,
            yara_service_url: "http://localhost:5000".to_string(),
            yara_timeout_secs: 5,
            yara_fail_open: true,
            python_pool_max_concurrent: 8,
            python_pool_max_idle: 4,
        };

        assert_eq!(config.function_key(), "slugify@1.0.0");
    }
}
