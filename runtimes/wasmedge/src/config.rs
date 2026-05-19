//! Runtime configuration for WasmEdge execution

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
    /// Maximum WASM fuel (instructions)
    pub max_fuel: u64,
    /// Maximum number of WebAssembly pages (64KB each)
    pub max_memory_pages: u32,
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
            max_fuel: 10_000_000,          // 10M fuel units
            max_memory_pages: 256,          // 256 * 64KB = 16MB
            allow_disk_io: false,
            allow_net: true,
        }
    }
}

/// Security policy for the runtime
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct SecurityPolicy {
    /// Blocked module patterns (for dynamic linking)
    pub blocked_modules: Vec<String>,
    /// Allowed hosts for network requests (empty = all allowed if allow_net)
    pub allowed_hosts: Vec<String>,
    /// Blocked hosts (takes precedence over allowed_hosts)
    pub blocked_hosts: Vec<String>,
    /// Environment variables to expose (empty = none)
    pub env_whitelist: Vec<String>,
    /// Enable secure sandbox mode
    pub sandbox_enabled: bool,
    /// Allowed syscalls (empty = use default restricted set)
    pub allowed_syscalls: Vec<String>,
    /// Enable syscall auditing
    pub audit_syscalls: bool,
    /// Network whitelist (domains/IPs allowed)
    pub network_whitelist: Vec<String>,
    /// Enable strict network whitelist enforcement
    pub strict_network_whitelist: bool,
    /// Maximum environment variables
    pub max_env_vars: usize,
    /// Maximum file descriptors
    pub max_file_descriptors: usize,
}

impl Default for SecurityPolicy {
    fn default() -> Self {
        Self {
            blocked_modules: vec![],
            allowed_hosts: vec![],
            blocked_hosts: vec![
                "169.254.169.254".to_string(),   // AWS metadata
                "metadata.google.internal".to_string(), // GCP metadata
                "metadata.azure.com".to_string(),
            ],
            env_whitelist: vec![],
            sandbox_enabled: true,
            allowed_syscalls: vec![
                "fd_write".to_string(),
                "fd_read".to_string(),
                "fd_close".to_string(),
                "fd_seek".to_string(),
                "path_open".to_string(),
                "clock_time_get".to_string(),
                "random_get".to_string(),
                "proc_exit".to_string(),
            ],
            audit_syscalls: true,
            network_whitelist: vec![],
            strict_network_whitelist: false,
            max_env_vars: 20,
            max_file_descriptors: 50,
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
        assert_eq!(limits.max_fuel, 10_000_000);
    }

    #[test]
    fn test_default_security_policy() {
        let policy = SecurityPolicy::default();
        assert!(policy.sandbox_enabled);
        assert!(policy.audit_syscalls);
        assert!(policy.allowed_syscalls.contains(&"proc_exit".to_string()));
    }
}
