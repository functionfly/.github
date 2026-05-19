//! Security management for WasmEdge runtime
//!
//! Provides permission-based access control, syscall filtering,
//! network security policies, and attack pattern detection.

use crate::config::SecurityPolicy;
use serde::{Deserialize, Serialize};
use std::collections::HashSet;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Individual permission types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Permission {
    /// Allow reading from file system
    Read,
    /// Allow writing to file system
    Write,
    /// Allow network access
    Network,
    /// Allow environment variable access
    EnvAccess,
    /// Allow subprocess spawning
    Spawn,
    /// Allow dynamic code execution (eval)
    DynamicCode,
    /// Allow loading WASM modules
    Wasm,
    /// Allow loading native modules
    NativeModules,
    /// Allow system time access
    SystemTime,
}

impl Permission {
    pub fn as_str(&self) -> &'static str {
        match self {
            Permission::Read => "read",
            Permission::Write => "write",
            Permission::Network => "network",
            Permission::EnvAccess => "env",
            Permission::Spawn => "spawn",
            Permission::DynamicCode => "dynamic-code",
            Permission::Wasm => "wasm",
            Permission::NativeModules => "native-modules",
            Permission::SystemTime => "system-time",
        }
    }
}

/// A set of permissions
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct PermissionSet(HashSet<Permission>);

impl PermissionSet {
    /// Create a new empty permission set
    pub fn new() -> Self {
        Self(HashSet::new())
    }

    /// Create a permission set from a list of permissions
    pub fn from_slice(perms: &[Permission]) -> Self {
        Self(HashSet::from_iter(perms.iter().copied()))
    }

    /// Add a permission to the set
    pub fn insert(&mut self, perm: Permission) {
        self.0.insert(perm);
    }

    /// Check if the set contains a permission
    pub fn contains(&self, perm: Permission) -> bool {
        self.0.contains(&perm)
    }

    /// Get all permissions as a slice
    pub fn as_slice(&self) -> Vec<Permission> {
        self.0.iter().copied().collect()
    }

    /// Returns true if no permissions are granted
    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }
}

/// Security violation record
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityViolation {
    pub function_name: String,
    pub violation_type: ViolationType,
    pub timestamp: u64,
    pub details: String,
    pub severity: Severity,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ViolationType {
    SyscallViolation,
    ResourceExhaustion,
    MemoryViolation,
    NetworkViolation,
    FilesystemViolation,
    TimeAccessViolation,
    CapabilityViolation,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum Severity {
    Low,
    Medium,
    High,
    Critical,
}

/// Attack pattern detection
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AttackPattern {
    pub pattern_type: AttackPatternType,
    pub function_name: String,
    pub occurrences: usize,
    pub first_seen: u64,
    pub last_seen: u64,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum AttackPatternType {
    SyscallFlood,
    MemoryExhaustion,
    PathTraversal,
    CommandInjection,
    ResourceStarvation,
}

/// Security manager for runtime execution
#[derive(Debug, Clone)]
pub struct SecurityManager {
    policy: SecurityPolicy,
    violations: Arc<RwLock<Vec<SecurityViolation>>>,
    #[allow(dead_code)]
    attack_patterns: Arc<RwLock<Vec<AttackPattern>>>,
}

impl SecurityManager {
    /// Create a new security manager from a security policy
    pub fn new(policy: SecurityPolicy) -> Self {
        Self {
            policy,
            violations: Arc::new(RwLock::new(Vec::new())),
            attack_patterns: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Create a security manager with default policy
    pub fn default() -> Self {
        Self::new(SecurityPolicy::default())
    }

    /// Check if a module is allowed to be loaded
    pub fn is_module_allowed(&self, module_name: &str) -> bool {
        // Check blocked modules
        for blocked in &self.policy.blocked_modules {
            if module_name == blocked || module_name.starts_with(&format!("{}/", blocked)) {
                return false;
            }
        }
        true
    }

    /// Check if a host is allowed for network requests
    pub fn is_host_allowed(&self, host: &str) -> bool {
        // Check blocked hosts first (takes precedence)
        for blocked in &self.policy.blocked_hosts {
            if host == blocked || host.ends_with(&format!(".{}", blocked)) {
                return false;
            }
        }

        // If strict whitelist is enabled, check it
        if self.policy.strict_network_whitelist {
            if self.policy.network_whitelist.is_empty() {
                return false; // Empty whitelist with strict mode = deny all
            }
            for allowed in &self.policy.network_whitelist {
                if host == allowed || host.ends_with(&format!(".{}", allowed)) {
                    return true;
                }
            }
            return false;
        }

        // If allowed_hosts is empty, allow all (except blocked)
        if self.policy.allowed_hosts.is_empty() {
            return true;
        }

        // Check allowed hosts
        for allowed in &self.policy.allowed_hosts {
            if host == allowed || host.ends_with(&format!(".{}", allowed)) {
                return true;
            }
        }

        false
    }

    /// Check if an environment variable is accessible
    pub fn is_env_allowed(&self, var_name: &str) -> bool {
        // Check for dangerous env vars
        let upper = var_name.to_uppercase();
        if upper.contains("LD_") || upper.contains("PATH") || upper.contains("SHELL") {
            return false;
        }

        self.policy.env_whitelist.is_empty() || self.policy.env_whitelist.iter().any(|v| v == var_name)
    }

    /// Check if a syscall is allowed
    pub fn is_syscall_allowed(&self, syscall: &str) -> bool {
        if self.policy.allowed_syscalls.is_empty() {
            // Default restricted set
            matches!(
                syscall,
                "fd_write" | "fd_read" | "fd_close" | "fd_seek" | "fd_fdstat_get"
                    | "path_open" | "clock_time_get" | "random_get" | "proc_exit"
            )
        } else {
            self.policy.allowed_syscalls.iter().any(|s| s == syscall)
        }
    }

    /// Check if an environment variable count is within limits
    pub fn is_env_count_allowed(&self, count: usize) -> bool {
        count <= self.policy.max_env_vars
    }

    /// Check if file descriptor count is within limits
    pub fn is_fd_count_allowed(&self, count: usize) -> bool {
        count <= self.policy.max_file_descriptors
    }

    /// Record a security violation
    pub async fn record_violation(&self, function_name: String, violation_type: ViolationType, details: String, severity: Severity) {
        let violation = SecurityViolation {
            function_name,
            violation_type,
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
            details,
            severity: severity.clone(),
        };

        let mut violations = self.violations.write().await;
        violations.push(violation.clone());

        // Keep only recent violations (last 1000)
        if violations.len() > 1000 {
            violations.drain(0..100);
        }

        // Log based on severity
        match severity {
            Severity::High | Severity::Critical => {
                tracing::error!("Security violation: {:?}", violation);
            }
            Severity::Medium => {
                tracing::warn!("Security violation: {:?}", violation);
            }
            _ => {
                tracing::info!("Security violation: {:?}", violation);
            }
        }
    }

    /// Get the security policy
    pub fn policy(&self) -> &SecurityPolicy {
        &self.policy
    }

    /// Calculate permissions for a given set of limits and policy
    pub fn calculate_permissions(&self, allow_net: bool, allow_disk: bool) -> PermissionSet {
        let mut perms = PermissionSet::new();

        if self.policy.sandbox_enabled {
            if allow_disk {
                perms.insert(Permission::Read);
            }
            if allow_net {
                perms.insert(Permission::Network);
            }
        } else {
            // Full permissions when sandbox is disabled
            perms.insert(Permission::Read);
            perms.insert(Permission::Write);
            perms.insert(Permission::Network);
            perms.insert(Permission::EnvAccess);
            perms.insert(Permission::Spawn);
            perms.insert(Permission::Wasm);
            perms.insert(Permission::SystemTime);
        }

        perms
    }

    /// Validate URL against network whitelist
    pub fn validate_url(&self, url: &str) -> Result<(), String> {
        let parsed = url::Url::parse(url)
            .map_err(|e| format!("invalid URL: {}", e))?;

        let host = parsed.host_str()
            .ok_or("URL has no host")?;

        if !self.is_host_allowed(host) {
            return Err(format!("host '{}' is not allowed", host));
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_host_blocking() {
        let manager = SecurityManager::default();
        assert!(!manager.is_host_allowed("169.254.169.254"));
        assert!(!manager.is_host_allowed("metadata.google.internal"));
        assert!(manager.is_host_allowed("api.example.com"));
    }

    #[test]
    fn test_env_blocking() {
        let manager = SecurityManager::default();
        assert!(!manager.is_env_allowed("LD_LIBRARY_PATH"));
        assert!(!manager.is_env_allowed("PATH"));
        assert!(manager.is_env_allowed("DATABASE_URL"));
    }

    #[test]
    fn test_syscall_filtering() {
        let manager = SecurityManager::default();
        assert!(manager.is_syscall_allowed("fd_write"));
        assert!(manager.is_syscall_allowed("random_get"));
        assert!(!manager.is_syscall_allowed("socket")); // Not in default allowed list
    }

    #[test]
    fn test_permission_set() {
        let mut perms = PermissionSet::new();
        perms.insert(Permission::Network);
        perms.insert(Permission::Read);

        assert!(perms.contains(Permission::Network));
        assert!(perms.contains(Permission::Read));
        assert!(!perms.contains(Permission::Write));
    }

    #[tokio::test]
    async fn test_violation_recording() {
        let manager = SecurityManager::default();
        manager.record_violation(
            "test_func".to_string(),
            ViolationType::SyscallViolation,
            "blocked syscall".to_string(),
            Severity::High,
        ).await;

        let violations = manager.violations.read().await;
        assert_eq!(violations.len(), 1);
        assert_eq!(violations[0].severity, Severity::High);
    }
}
