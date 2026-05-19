//! Ruby Runtime Configuration
//!
//! Defines execution limits, security policies, and runtime settings.

use serde::{Deserialize, Serialize};
use std::time::Duration;

/// Execution limits for Ruby code
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionLimits {
    /// Maximum memory in MB
    pub max_memory_mb: u64,
    /// Maximum CPU time in seconds
    pub max_cpu_time_secs: u64,
    /// Maximum wall clock time in seconds
    pub max_wall_time_secs: u64,
    /// Maximum output size in bytes
    pub max_output_bytes: usize,
    /// Maximum stack depth
    pub max_stack_depth: u32,
    /// Maximum number of allocated objects
    pub max_allocations: u64,
}

impl Default for ExecutionLimits {
    fn default() -> Self {
        Self {
            max_memory_mb: 256,
            max_cpu_time_secs: 30,
            max_wall_time_secs: 60,
            max_output_bytes: 1024 * 1024, // 1MB
            max_stack_depth: 1024,
            max_allocations: 1_000_000,
        }
    }
}

/// Security policy for Ruby execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityPolicy {
    /// Enable sandbox isolation
    pub sandbox_enabled: bool,
    /// Enable seccomp filtering (Linux only)
    pub enable_seccomp: bool,
    /// Enable landlock restrictions (Linux only)
    pub enable_landlock: bool,
    /// Allow filesystem access
    pub allow_filesystem: bool,
    /// Allow network access
    pub allow_network: bool,
    /// Allowed directories for file operations
    pub allowed_dirs: Vec<String>,
    /// Blocked system calls
    pub blocked_syscalls: Vec<String>,
    /// Enable code sanitization
    pub sanitize_code: bool,
    /// Maximum_require_depth for Ruby requires
    pub max_require_depth: u32,
}

impl Default for SecurityPolicy {
    fn default() -> Self {
        Self {
            sandbox_enabled: true,
            enable_seccomp: false,
            enable_landlock: false,
            allow_filesystem: false,
            allow_network: false,
            allowed_dirs: vec![],
            blocked_syscalls: vec![],
            sanitize_code: true,
            max_require_depth: 3,
        }
    }
}

/// Ruby engine configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RubyConfig {
    /// Enable $SAFE taint checking
    pub enable_safe: bool,
    /// Maximum constant nesting depth
    pub max_constant_depth: u32,
    /// Maximum method call depth
    pub max_method_depth: u32,
    /// Enable verbose error messages
    pub verbose_errors: bool,
    /// Default encoding
    pub default_encoding: String,
}

impl Default for RubyConfig {
    fn default() -> Self {
        Self {
            enable_safe: true,
            max_constant_depth: 64,
            max_method_depth: 256,
            verbose_errors: true,
            default_encoding: "UTF-8".to_string(),
        }
    }
}

/// Runtime configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RuntimeConfig {
    /// Execution limits
    pub limits: ExecutionLimits,
    /// Security policy
    pub security: SecurityPolicy,
    /// Ruby-specific configuration
    pub ruby: RubyConfig,
    /// Enable sandbox mode
    pub use_sandbox: bool,
    /// Maximum concurrent executions
    pub max_concurrent: usize,
    /// Default timeout for executions
    pub default_timeout: Duration,
}

impl Default for RuntimeConfig {
    fn default() -> Self {
        Self {
            limits: ExecutionLimits::default(),
            security: SecurityPolicy::default(),
            ruby: RubyConfig::default(),
            use_sandbox: true,
            max_concurrent: 100,
            default_timeout: Duration::from_secs(30),
        }
    }
}

impl RuntimeConfig {
    /// Validate the configuration
    pub fn validate(&self) -> Result<(), String> {
        if self.limits.max_memory_mb == 0 {
            return Err("max_memory_mb must be > 0".to_string());
        }
        if self.limits.max_cpu_time_secs == 0 {
            return Err("max_cpu_time_secs must be > 0".to_string());
        }
        if self.limits.max_wall_time_secs < self.limits.max_cpu_time_secs {
            return Err("max_wall_time_secs must be >= max_cpu_time_secs".to_string());
        }
        if self.max_concurrent == 0 {
            return Err("max_concurrent must be > 0".to_string());
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = RuntimeConfig::default();
        config.validate().expect("default config should be valid");
    }

    #[test]
    fn test_validation() {
        let mut config = RuntimeConfig::default();
        config.limits.max_memory_mb = 0;
        assert!(config.validate().is_err());

        let mut config = RuntimeConfig::default();
        config.limits.max_wall_time_secs = 0;
        assert!(config.validate().is_err());
    }
}