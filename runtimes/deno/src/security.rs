//! Security management for Deno runtime
//!
//! Provides permission-based access control, module restrictions,
//! and network security policies.

use crate::config::SecurityPolicy;
use serde::{Deserialize, Serialize};
use std::collections::HashSet;

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

/// Security manager for runtime execution
#[derive(Debug, Clone)]
pub struct SecurityManager {
    policy: SecurityPolicy,
}

impl SecurityManager {
    /// Create a new security manager from a security policy
    pub fn new(policy: SecurityPolicy) -> Self {
        Self { policy }
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
        self.policy.env_whitelist.is_empty() || self.policy.env_whitelist.iter().any(|v| v == var_name)
    }

    /// Check if eval/dynamic code is allowed
    pub fn is_dynamic_code_allowed(&self) -> bool {
        !self.policy.sandbox_enabled || self.policy.allow_eval || self.policy.allow_dynamic_code
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
            if self.policy.allow_eval {
                perms.insert(Permission::DynamicCode);
            }
            perms.insert(Permission::Wasm);
        }

        perms
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_module_blocking() {
        let manager = SecurityManager::default();
        assert!(!manager.is_module_allowed("child_process"));
        assert!(!manager.is_module_allowed("fs"));
        assert!(manager.is_module_allowed("fetch")); // Allowed module
    }

    #[test]
    fn test_host_blocking() {
        let manager = SecurityManager::default();
        assert!(!manager.is_host_allowed("169.254.169.254"));
        assert!(!manager.is_host_allowed("metadata.google.internal"));
        assert!(manager.is_host_allowed("api.example.com"));
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
}