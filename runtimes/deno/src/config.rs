//! Runtime configuration for Deno execution

use serde::{Deserialize, Serialize};
use std::time::Duration;

/// Main runtime configuration
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct RuntimeConfig {
    /// Execution limits
    pub limits: ExecutionLimits,
    /// Security policy
    pub security: SecurityPolicy,
    /// Enable WASM sandbox
    pub use_sandbox: bool,
    /// Maximum concurrent executions
    pub max_concurrent: usize,
    /// Execution timeout
    pub default_timeout: Duration,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            limits: ExecutionLimits::default(),
            security: SecurityPolicy::default(),
            use_sandbox: true,
            max_concurrent: 100,
            default_timeout: Duration::from_secs(30),
        }
    }
}

/// Execution limits for a single function execution
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionLimits {
    /// Maximum memory in MB
    pub max_memory_mb: u64,
    /// Maximum CPU time in seconds
    pub max_cpu_time_secs: u64,
    /// Maximum wall time (including I/O)
    pub max_wall_time_secs: u64,
    /// Maximum output size in bytes
    pub max_output_bytes: usize,
    /// Maximum number of modules that can be loaded
    pub max_modules: usize,
    /// Maximum total network request size
    pub max_network_request_bytes: usize,
    /// Enable disk I/O (sandboxed)
    pub allow_disk_io: bool,
    /// Allow network access
    pub allow_net: bool,
}

impl Default for ExecutionLimits {
    fn default() -> Self {
        Self {
            max_memory_mb: 512,
            max_cpu_time_secs: 10,
            max_wall_time_secs: 30,
            max_output_bytes: 1024 * 1024, // 1MB
            max_modules: 50,
            max_network_request_bytes: 10 * 1024 * 1024, // 10MB
            allow_disk_io: false,
            allow_net: true,
        }
    }
}

/// Security policy for the runtime
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SecurityPolicy {
    /// Blocked module patterns (e.g., "fs", "child_process", "unsafe-eval")
    pub blocked_modules: Vec<String>,
    /// Allowed hosts for network requests (empty = all allowed if allow_net)
    pub allowed_hosts: Vec<String>,
    /// Blocked hosts (takes precedence over allowed_hosts)
    pub blocked_hosts: Vec<String>,
    /// Environment variables to expose (empty = none)
    pub env_whitelist: Vec<String>,
    /// Enable secure sandbox mode
    pub sandbox_enabled: bool,
    /// Allow eval() and similar unsafe operations
    pub allow_eval: bool,
    /// Allow dynamic code generation
    pub allow_dynamic_code: bool,
    /// Maximum depth for module imports
    pub max_module_depth: usize,
}

impl Default for SecurityPolicy {
    fn default() -> Self {
        Self {
            blocked_modules: vec![
                "child_process".to_string(),
                "fs".to_string(),
                "net".to_string(),
                "tls".to_string(),
                "crypto".to_string(),
                "worker_threads".to_string(),
                "cluster".to_string(),
                "async_hooks".to_string(),
            ],
            allowed_hosts: vec![],
            blocked_hosts: vec![
                "169.254.169.254".to_string(), // AWS metadata
                "metadata.google.internal".to_string(), // GCP metadata
            ],
            env_whitelist: vec![],
            sandbox_enabled: true,
            allow_eval: false,
            allow_dynamic_code: false,
            max_module_depth: 10,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = RuntimeConfig::default();
        assert!(config.use_sandbox);
        assert_eq!(config.max_concurrent, 100);
    }

    #[test]
    fn test_default_limits() {
        let limits = ExecutionLimits::default();
        assert_eq!(limits.max_memory_mb, 512);
        assert_eq!(limits.max_cpu_time_secs, 10);
    }

    #[test]
    fn test_default_security_policy() {
        let policy = SecurityPolicy::default();
        assert!(policy.sandbox_enabled);
        assert!(!policy.allow_eval);
        assert!(policy.blocked_modules.contains(&"child_process".to_string()));
    }
}