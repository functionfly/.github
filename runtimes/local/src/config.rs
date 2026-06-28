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
    /// Default: true (required for CPython-WASI stdlib imports)
    #[arg(long, default_value = "true")]
    pub wasi_allow_time: bool,

    /// Shutdown timeout in seconds (for graceful shutdown)
    #[arg(long, default_value = "30")]
    pub shutdown_timeout_secs: u64,

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

    /// SMTP use TLS (for port 465 or explicit TLS)
    #[arg(long, default_value = "false")]
    pub smtp_use_tls: bool,

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

    /// MicroVM orchestrator URL (for Enterprise tier; must be set explicitly)
    #[arg(long)]
    pub orchestrator_url: String,

    /// Tenant UUID forwarded to the MicroVM orchestrator for isolation and billing.
    /// Go sets this via --tenant-id when starting the local runtime for an enterprise tenant.
    #[arg(long)]
    pub tenant_id: Option<String>,

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

    /// CORS allowed origin for the HTTP server.
    /// Use "*" or leave empty to allow all origins (default, suitable for local dev).
    /// Set to a specific origin (e.g. "https://app.example.com") for production.
    #[arg(long, default_value = "*")]
    pub cors_allow_origin: String,

    /// Enable AOT (Ahead-of-Time) module compilation cache.
    /// When enabled, Wasm modules are compiled once and the result is cached
    /// in memory (and optionally on disk) for near-instant subsequent loads.
    #[arg(long, default_value = "true")]
    pub aot_cache_enabled: bool,

    /// Directory for persisting AOT-compiled modules to disk.
    /// Empty string means in-memory cache only (no disk persistence).
    #[arg(long, default_value = "")]
    pub aot_cache_dir: String,

    /// Maximum size of the AOT cache in MB.
    /// Oldest entries are evicted when this limit is reached.
    #[arg(long, default_value = "512")]
    pub aot_cache_size_mb: usize,

    /// Maximum number of compiled modules to keep in the in-memory LRU cache.
    #[arg(long, default_value = "64")]
    pub aot_cache_memory_capacity: usize,

    /// Fuel units per millisecond of CPU time (calibration constant).
    /// Used to convert timeout_ms into a fuel budget for Wasmtime's fuel
    /// metering.  Typical value on modern hardware: ~10_000_000 fuel/ms.
    #[arg(long, default_value = "10000000")]
    pub fuel_per_ms: u64,

    /// CPU time limit in milliseconds (0 = use cpu_fuel_limit). When set, overrides fuel via fuel_per_ms.
    #[arg(long, default_value = "0")]
    pub cpu_ms_limit: u64,

    /// Use CPython compiled to WASM instead of RustPython for Python functions.
    #[arg(long, default_value = "true")]
    pub use_cpython_wasm: bool,

    /// Path to the CPython-WASM binary used when use_cpython_wasm is true.
    #[arg(long, default_value = "./runtimes/cpython.wasm")]
    pub cpython_wasm_path: String,

    /// Path to the CPython-WASI stdlib directory (contains lib/python3.xx/ or python3xx.zip).
    #[arg(long, default_value = "./runtimes/cpython-wasi/lib")]
    pub cpython_stdlib_path: String,

    /// Enable persistent daemon mode.
    #[arg(long, default_value = "false")]
    pub daemon_mode: bool,

    /// Enable YARA scanning of WASM artifacts before execution.
    #[arg(long, default_value = "false")]
    pub yara_scan_enabled: bool,

    /// URL of the YARA service (must be set explicitly when yara_scan_enabled=true)
    #[arg(long)]
    pub yara_service_url: String,

    /// YARA scanner timeout in seconds.
    #[arg(long, default_value = "5")]
    pub yara_timeout_secs: u64,

    /// If true, allow execution when the YARA service is unreachable (fail-open).
    #[arg(long, default_value = "true")]
    pub yara_fail_open: bool,

    /// Maximum number of concurrent Python (RustPython) runtimes.
    #[arg(long, default_value = "8")]
    pub python_pool_max_concurrent: usize,

    /// Maximum number of idle Python runtimes to keep warm in the pool.
    #[arg(long, default_value = "4")]
    pub python_pool_max_idle: usize,

    /// Maximum number of times a Python runtime can be reused before being recycled.
    #[arg(long, default_value = "100")]
    pub python_pool_max_reuse: usize,

    /// Path to a JSON file containing secrets (key-value pairs).
    /// Environment variables prefixed with SECRET_ take precedence over file entries.
    #[arg(long)]
    pub secrets_file: Option<String>,

    /// Maximum number of messages per named queue (queue capability).
    #[arg(long, default_value = "1000")]
    pub queue_max_len: usize,

    /// Maximum number of distinct named queues allowed per function instance.
    #[arg(long, default_value = "16")]
    pub queue_max_queues: usize,

    /// Enable seccomp-BPF syscall filtering after initialization.
    /// When true, only syscalls in the runtime's allowlist are permitted.
    /// Includes architecture validation, NO_NEW_PRIVS, and CLONE_NEWUSER
    /// restriction.
    /// Default: false (opt-in for now; will default to true in production).
    #[arg(long, default_value = "false")]
    pub enable_seccomp: bool,

    /// When seccomp is enabled, use KILL_PROCESS as the default action for
    /// disallowed syscalls (instead of ENOSYS).  This terminates the process
    /// immediately on any syscall not in the allowlist.
    /// Also controls whether filter installation failure is fatal.
    /// Default: false for backward compatibility; required for production.
    #[arg(long, default_value = "false")]
    pub seccomp_strict: bool,

    /// When seccomp is enabled, log disallowed syscalls via the kernel audit
    /// subsystem before returning ENOSYS.  Use this to discover which syscalls
    /// the runtime actually invokes before switching to --seccomp-strict.
    /// Ignored when --seccomp-strict is true (strict mode logs and kills).
    /// Default: false.
    #[arg(long, default_value = "false")]
    pub seccomp_monitor: bool,

    /// Enable Linux network namespace isolation.
    /// When true, the runtime process is moved to a new network namespace with
    /// only the loopback interface available. Egress must be explicitly allowed
    /// via --network-whitelist iptables rules.
    /// Default: false (opt-in; requires CAP_NET_ADMIN).
    #[arg(long, default_value = "false")]
    pub enable_net_ns: bool,

    /// When network namespace isolation is enabled, fail hard if it cannot be applied.
    /// Requires CAP_NET_ADMIN capability. Set to true for production environments
    /// where network isolation is mandatory.
    /// Default: false for backward compatibility; recommended true for production.
    #[arg(long, default_value = "false")]
    pub netns_strict: bool,

    /// Enable WASM instance pooling for warm-instance reuse.
    /// When enabled, compiled modules and WASI contexts are pooled per function,
    /// reducing per-execution overhead (~1ms savings per invocation).
    /// Default: true for production.
    #[arg(long, default_value = "true")]
    pub wasm_pool_enabled: bool,

    /// Maximum concurrent WASM executions per function when pooling is enabled.
    /// Each concurrent execution holds a slot in the pool's semaphore.
    #[arg(long, default_value = "10")]
    pub wasm_pool_max_concurrent: usize,

    /// Maximum idle WASM instances to keep warm per function.
    /// Idle instances consume memory but enable sub-millisecond cold-start.
    #[arg(long, default_value = "4")]
    pub wasm_pool_max_idle: usize,

    /// Number of instances to pre-warm per function on startup.
    /// Set to 0 to disable pre-warming.
    #[arg(long, default_value = "2")]
    pub wasm_pool_prewarm_count: usize,
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
                eprintln!(
                    "Warning: Unknown tier '{}', defaulting to UltraLow",
                    self.tier
                );
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
                self.memory_mb,
                self.tier,
                tier_specs.ram_gb * 1024
            ));
        }

        Ok(())
    }

    /// Check if this tier supports MicroVM execution
    pub fn supports_microvm(&self) -> bool {
        self.enterprise_enabled
            && matches!(
                self.get_budget_tier(),
                BudgetTier::High | BudgetTier::Medium
            )
    }

    /// Check if CPython-WASM is enabled and available
    pub fn supports_cpython_wasm(&self) -> bool {
        self.use_cpython_wasm && !self.cpython_wasm_path.is_empty()
    }

    /// Compute the fuel budget for a given timeout in milliseconds.
    /// Uses the calibrated fuel_per_ms constant.
    pub fn fuel_for_timeout(&self) -> u64 {
        self.timeout_ms.saturating_mul(self.fuel_per_ms)
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
            shutdown_timeout_secs: 30,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: Vec::new(),
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            smtp_use_tls: false,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: String::new(), // Must be set explicitly
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tenant_id: None,
            tier: "ultra-low".to_string(),
            network_whitelist: Vec::new(),
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
            max_output_bytes: 1024 * 1024, // 1 MiB
            max_input_bytes: 1024 * 1024,  // 1 MiB
            microvm_fallback_allowed: true,
            cors_allow_origin: "*".to_string(),
            aot_cache_enabled: true,
            aot_cache_dir: "".to_string(),
            aot_cache_size_mb: 512,
            aot_cache_memory_capacity: 64,
            fuel_per_ms: 10_000_000,
            cpu_ms_limit: 0,
            use_cpython_wasm: true,
            cpython_wasm_path: "./runtimes/cpython.wasm".to_string(),
            cpython_stdlib_path: "./runtimes/cpython-wasi/lib".to_string(),
            daemon_mode: false,
            yara_scan_enabled: false,
            yara_service_url: String::new(), // Must be set explicitly when yara_scan_enabled=true
            yara_timeout_secs: 5,
            yara_fail_open: true,
            python_pool_max_concurrent: 8,
            python_pool_max_idle: 4,
            python_pool_max_reuse: 100,
            secrets_file: None,
            queue_max_len: 1000,
            queue_max_queues: 16,
            enable_seccomp: false,
            seccomp_strict: false,
            seccomp_monitor: false,
            enable_net_ns: false,
            netns_strict: false,
            wasm_pool_enabled: true,
            wasm_pool_max_concurrent: 10,
            wasm_pool_max_idle: 4,
            wasm_pool_prewarm_count: 2,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_function_key() {
        // Use Config::default() and override only the fields under test.
        // This avoids brittle tests that break when new fields are added.
        let config = Config {
            function: "slugify".to_string(),
            version: "1.0.0".to_string(),
            ..Config::default()
        };

        assert_eq!(config.function_key(), "slugify@1.0.0");
    }
}
