//! Security hardening and isolation enhancements for the sandbox.
//!
//! This module implements additional security measures beyond basic WASI isolation,
//! including attack surface reduction, syscall filtering, and resource isolation.

use std::collections::HashSet;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Security profile for function execution
#[derive(Debug, Clone)]
pub struct SecurityProfile {
    /// Allowed syscalls (empty means deny all except essential)
    pub allowed_syscalls: HashSet<String>,
    /// Maximum file descriptors per function
    pub max_file_descriptors: usize,
    /// Allow network access
    pub allow_network: bool,
    /// Allow filesystem access
    pub allow_filesystem: bool,
    /// Maximum environment variables
    pub max_env_vars: usize,
    /// Disable dangerous syscalls
    pub disable_dangerous_syscalls: bool,
    /// Enable syscall auditing
    pub audit_syscalls: bool,
    /// Network whitelist (allowed domains/IPs)
    pub network_whitelist: HashSet<String>,
    /// Enable strict network whitelist enforcement
    pub strict_network_whitelist: bool,
}

/// Security monitor for tracking and preventing attacks
pub struct SecurityMonitor {
    /// Security profiles per function
    profiles: Arc<RwLock<std::collections::HashMap<String, SecurityProfile>>>,
    /// Global security violations
    violations: Arc<RwLock<Vec<SecurityViolation>>>,
    /// Attack patterns detected
    attack_patterns: Arc<RwLock<Vec<AttackPattern>>>,
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SecurityViolation {
    pub function_name: String,
    pub violation_type: ViolationType,
    pub timestamp: u64,
    pub details: String,
    pub severity: Severity,
}

#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub enum ViolationType {
    SyscallViolation,
    ResourceExhaustion,
    MemoryViolation,
    NetworkViolation,
    FilesystemViolation,
    TimeAccessViolation,
    CapabilityViolation, // Function attempted to use undeclared capability
}

#[derive(Debug, Clone, PartialEq, serde::Serialize, serde::Deserialize)]
pub enum Severity {
    Low,
    Medium,
    High,
    Critical,
}

#[derive(Debug, Clone)]
pub struct AttackPattern {
    pub pattern_type: AttackPatternType,
    pub function_name: String,
    pub occurrences: usize,
    pub first_seen: u64,
    pub last_seen: u64,
}

#[derive(Debug, Clone, PartialEq)]
pub enum AttackPatternType {
    SyscallFlood,
    MemoryExhaustion,
    PathTraversal,
    CommandInjection,
    ResourceStarvation,
}

impl AttackPatternType {
    /// Get a human-readable name for the pattern type
    pub fn debug_name(&self) -> &'static str {
        match self {
            AttackPatternType::SyscallFlood => "SyscallFlood",
            AttackPatternType::MemoryExhaustion => "MemoryExhaustion",
            AttackPatternType::PathTraversal => "PathTraversal",
            AttackPatternType::CommandInjection => "CommandInjection",
            AttackPatternType::ResourceStarvation => "ResourceStarvation",
        }
    }
}

impl SecurityMonitor {
    /// Create a new security monitor
    pub fn new() -> Self {
        Self {
            profiles: Arc::new(RwLock::new(std::collections::HashMap::new())),
            violations: Arc::new(RwLock::new(Vec::new())),
            attack_patterns: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Create hardened security profile for ultra-secure execution
    pub fn create_hardened_profile() -> SecurityProfile {
        let mut allowed_syscalls = HashSet::new();

        // Essential WASI syscalls only
        allowed_syscalls.insert("fd_write".to_string());
        allowed_syscalls.insert("fd_read".to_string());
        allowed_syscalls.insert("fd_close".to_string());
        allowed_syscalls.insert("fd_seek".to_string());
        allowed_syscalls.insert("path_open".to_string()); // Limited filesystem
        allowed_syscalls.insert("clock_time_get".to_string()); // Controlled time access
        allowed_syscalls.insert("random_get".to_string());
        allowed_syscalls.insert("proc_exit".to_string());

        SecurityProfile {
            allowed_syscalls,
            max_file_descriptors: 10, // Very limited
            allow_network: false,
            allow_filesystem: false, // No host filesystem access
            max_env_vars: 5, // Minimal environment
            disable_dangerous_syscalls: true,
            audit_syscalls: true,
            network_whitelist: HashSet::new(), // No network access in hardened mode
            strict_network_whitelist: true,
        }
    }

    /// Create standard security profile with reasonable defaults
    pub fn create_standard_profile() -> SecurityProfile {
        let mut allowed_syscalls = HashSet::new();

        // Standard WASI syscalls
        let standard_calls = [
            "fd_write", "fd_read", "fd_close", "fd_seek", "fd_fdstat_get",
            "path_open", "path_create_directory", "path_remove_directory",
            "clock_time_get", "random_get", "proc_exit", "fd_readdir",
            "path_filestat_get", "fd_filestat_get", "path_unlink_file",
        ];

        for syscall in &standard_calls {
            allowed_syscalls.insert(syscall.to_string());
        }

        SecurityProfile {
            allowed_syscalls,
            max_file_descriptors: 50,
            allow_network: false, // Can be enabled per function
            allow_filesystem: true, // Controlled access
            max_env_vars: 20,
            disable_dangerous_syscalls: true,
            audit_syscalls: false, // Optional
            network_whitelist: HashSet::new(), // Empty by default, populated from config
            strict_network_whitelist: false, // Not strict by default
        }
    }

    /// Create enterprise security profile with network whitelist support
    pub fn create_enterprise_profile(network_whitelist: Vec<String>) -> SecurityProfile {
        let mut profile = Self::create_standard_profile();
        profile.allow_network = true;
        profile.network_whitelist = network_whitelist.into_iter().collect();
        profile.strict_network_whitelist = true;
        profile.audit_syscalls = true; // Enable auditing for enterprise
        profile
    }

    /// Register security profile for a function
    pub async fn register_profile(&self, function_key: String, profile: SecurityProfile) {
        let mut profiles = self.profiles.write().await;
        profiles.insert(function_key, profile);
    }

    /// Check if syscall is allowed for function
    pub async fn is_syscall_allowed(&self, function_key: &str, syscall: &str) -> bool {
        let profiles = self.profiles.read().await;
        if let Some(profile) = profiles.get(function_key) {
            profile.allowed_syscalls.contains(syscall) || !profile.disable_dangerous_syscalls
        } else {
            // Default to standard profile
            let standard = Self::create_standard_profile();
            standard.allowed_syscalls.contains(syscall) || !standard.disable_dangerous_syscalls
        }
    }

    /// Check if network request is allowed for function
    pub async fn is_network_allowed(&self, function_key: &str, url: &str) -> bool {
        let profiles = self.profiles.read().await;
        if let Some(profile) = profiles.get(function_key) {
            // If network is not allowed at all, deny
            if !profile.allow_network {
                return false;
            }

            // If strict whitelist is disabled, allow all
            if !profile.strict_network_whitelist {
                return true;
            }

            // Check if the URL matches any whitelist entry
            self.check_url_against_whitelist(url, &profile.network_whitelist)
        } else {
            // Default behavior: allow network but no strict whitelist
            true
        }
    }

    /// Check if URL matches network whitelist
    fn check_url_against_whitelist(&self, url: &str, whitelist: &HashSet<String>) -> bool {
        if whitelist.is_empty() {
            return false; // Empty whitelist means deny all when strict
        }

        // Parse the URL to extract domain/IP
        if let Ok(parsed_url) = url::Url::parse(url) {
            let host = parsed_url.host_str().unwrap_or("");

            // Check exact matches
            if whitelist.contains(host) {
                return true;
            }

            // Check wildcard patterns (*.domain.com)
            for pattern in whitelist {
                if let Some(domain_suffix) = pattern.strip_prefix("*.") {
                    if host.ends_with(domain_suffix) {
                        return true;
                    }
                }
            }

            false
        } else {
            false // Invalid URL
        }
    }

    /// Record security violation
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

        // Check for attack patterns
        self.detect_attack_patterns(&violation).await;

        // Log security events
        match severity {
            Severity::High | Severity::Critical => {
                tracing::error!("Security violation detected: {:?}", violation);
            }
            Severity::Medium => {
                tracing::warn!("Security violation detected: {:?}", violation);
            }
            _ => {
                tracing::info!("Security violation detected: {:?}", violation);
            }
        }
    }

    /// Detect attack patterns
    async fn detect_attack_patterns(&self, violation: &SecurityViolation) {
        let mut patterns = self.attack_patterns.write().await;

        // Simple pattern detection - can be enhanced with ML
        let pattern_type = match violation.violation_type {
            ViolationType::SyscallViolation => AttackPatternType::SyscallFlood,
            ViolationType::MemoryViolation => AttackPatternType::MemoryExhaustion,
            ViolationType::FilesystemViolation => AttackPatternType::PathTraversal,
            ViolationType::NetworkViolation => AttackPatternType::CommandInjection,
            ViolationType::ResourceExhaustion => AttackPatternType::ResourceStarvation,
            _ => return, // No pattern for other violations
        };

        // Check if we already have this pattern
        for pattern in patterns.iter_mut() {
            if pattern.pattern_type == pattern_type && pattern.function_name == violation.function_name {
                pattern.occurrences += 1;
                pattern.last_seen = violation.timestamp;

                // Alert on high occurrence patterns - use debug_name for readable output
                if pattern.occurrences > 10 {
                    tracing::error!(
                        "Attack pattern detected: type={}, function={}, occurrences={}, first_seen={}",
                        pattern.pattern_type.debug_name(),
                        pattern.function_name,
                        pattern.occurrences,
                        pattern.first_seen
                    );
                }
                return;
            }
        }

        // New pattern
        let new_pattern = AttackPattern {
            pattern_type,
            function_name: violation.function_name.clone(),
            occurrences: 1,
            first_seen: violation.timestamp,
            last_seen: violation.timestamp,
        };

        patterns.push(new_pattern);
    }

    /// Get security report
    pub async fn get_security_report(&self) -> SecurityReport {
        let violations = self.violations.read().await;
        let patterns = self.attack_patterns.read().await;

        let total_violations = violations.len();
        let critical_violations = violations.iter().filter(|v| matches!(v.severity, Severity::Critical)).count();
        let high_violations = violations.iter().filter(|v| matches!(v.severity, Severity::High)).count();

        SecurityReport {
            total_violations,
            critical_violations,
            high_violations,
            active_attack_patterns: patterns.len(),
            recent_violations: violations.iter().rev().take(10).cloned().collect(),
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs(),
        }
    }

    /// Check if function should be blocked due to security violations
    pub async fn should_block_function(&self, function_name: &str) -> bool {
        let violations = self.violations.read().await;

        // Count recent violations (last hour)
        let one_hour_ago = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs() - 3600;

        let recent_critical = violations.iter()
            .filter(|v| v.function_name == function_name)
            .filter(|v| v.timestamp > one_hour_ago)
            .filter(|v| matches!(v.severity, Severity::Critical))
            .count();

        // Block if more than 5 critical violations in the last hour
        recent_critical > 5
    }

    /// Get the security profile for a function
    pub async fn get_profile(&self, function_key: &str) -> SecurityProfile {
        let profiles = self.profiles.read().await;
        profiles.get(function_key).cloned().unwrap_or_else(Self::create_standard_profile)
    }

    /// Check if filesystem access is allowed for a function
    pub async fn is_filesystem_allowed(&self, function_key: &str) -> bool {
        let profile = self.get_profile(function_key).await;
        profile.allow_filesystem
    }

    /// Check if the number of open file descriptors is within limits
    pub async fn check_file_descriptor_limit(&self, function_key: &str, current_count: usize) -> bool {
        let profile = self.get_profile(function_key).await;
        if current_count >= profile.max_file_descriptors {
            tracing::warn!(
                "File descriptor limit exceeded for {}: {} >= {}",
                function_key, current_count, profile.max_file_descriptors
            );
            return false;
        }
        true
    }

    /// Check if the number of environment variables is within limits
    pub async fn check_env_vars_limit(&self, function_key: &str, env_var_count: usize) -> bool {
        let profile = self.get_profile(function_key).await;
        if env_var_count > profile.max_env_vars {
            tracing::warn!(
                "Environment variable limit exceeded for {}: {} > {}",
                function_key, env_var_count, profile.max_env_vars
            );
            return false;
        }
        true
    }

    /// Get all attack patterns for a function
    pub async fn get_attack_patterns(&self, function_name: &str) -> Vec<AttackPattern> {
        let patterns = self.attack_patterns.read().await;
        patterns.iter()
            .filter(|p| p.function_name == function_name)
            .cloned()
            .collect()
    }

    /// Get attack pattern with first_seen timestamp for forensics
    pub async fn get_attack_pattern_with_history(&self, function_name: &str, pattern_type: AttackPatternType) -> Option<AttackPattern> {
        let patterns = self.attack_patterns.read().await;
        patterns.iter()
            .filter(|p| p.function_name == function_name && p.pattern_type == pattern_type)
            .cloned()
            .inspect(|p| {
                // Use first_seen to calculate pattern age
                let age_seconds = std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap_or_default()
                    .as_secs()
                    .saturating_sub(p.first_seen);
                tracing::debug!(
                    "Attack pattern '{}' for {} first seen {} seconds ago, {} occurrences",
                    p.pattern_type.debug_name(), p.function_name, age_seconds, p.occurrences
                );
            })
            .collect::<Vec<_>>()
            .into_iter().next()
    }
}

/// Security report for monitoring
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SecurityReport {
    pub total_violations: usize,
    pub critical_violations: usize,
    pub high_violations: usize,
    pub active_attack_patterns: usize,
    pub recent_violations: Vec<SecurityViolation>,
    pub timestamp: u64,
}

/// Isolation utilities for enhanced sandboxing
pub struct IsolationUtils;

impl IsolationUtils {
    /// Create isolated temporary directory for function execution
    pub fn create_isolated_temp_dir(function_id: &str) -> anyhow::Result<std::path::PathBuf> {
        let base_temp = std::env::temp_dir();
        let isolated_path = base_temp.join(format!("functionfly-{}-{}", function_id, uuid::Uuid::new_v4()));

        std::fs::create_dir_all(&isolated_path)?;
        Ok(isolated_path)
    }

    /// Clean up isolated directory after execution
    pub fn cleanup_isolated_dir(path: &std::path::Path) -> std::io::Result<()> {
        if path.exists() && path.starts_with(std::env::temp_dir()) {
            std::fs::remove_dir_all(path)?;
        }
        Ok(())
    }

    /// Validate that a path is safe (no path traversal)
    pub fn is_safe_path(base_path: &std::path::Path, requested_path: &str) -> bool {
        let requested = std::path::Path::new(requested_path);

        // Reject paths with null bytes
        if requested_path.contains('\0') {
            return false;
        }

        // Reject paths with obvious traversal attempts before canonicalization
        if requested_path.contains("..\\") || requested_path.contains("../") {
            // Only reject if the resolved path would escape base_path
            // Some legitimate paths may contain .. but still resolve safely
        }

        // Try to resolve the base path first
        let resolved_base = std::fs::canonicalize(base_path).unwrap_or_else(|_| base_path.to_path_buf());

        // For the requested path, we must canonicalize to resolve symlinks and ..
        // If the file doesn't exist yet, we check the parent directory
        let resolved_requested = if requested.exists() {
            match std::fs::canonicalize(requested) {
                Ok(p) => p,
                Err(_) => return false, // Cannot resolve path — reject for safety
            }
        } else {
            // File doesn't exist — resolve parent directory and append filename
            if let Some(parent) = requested.parent() {
                let resolved_parent = match std::fs::canonicalize(parent) {
                    Ok(p) => p,
                    Err(_) => return false, // Cannot resolve parent — reject for safety
                };
                if let Some(file_name) = requested.file_name() {
                    resolved_parent.join(file_name)
                } else {
                    return false; // No filename component
                }
            } else {
                return false; // No parent directory
            }
        };

        // Verify the resolved path is within the base path
        resolved_requested.starts_with(&resolved_base)
    }

    /// Sanitize environment variables to prevent injection
    pub fn sanitize_env_vars(env_vars: &[(String, String)]) -> Vec<(String, String)> {
        env_vars.iter().filter_map(|(key, value)| {
            // Remove dangerous environment variables
            if key.to_uppercase().contains("LD_") ||
               key.to_uppercase().contains("PATH") ||
               key.to_uppercase().contains("SHELL") {
                None
            } else {
                // Sanitize values (basic filtering)
                let sanitized_value = value.chars()
                    .filter(|c| c.is_alphanumeric() || *c == '_' || *c == '-' || *c == '.' || *c == '/')
                    .collect::<String>();
                Some((key.clone(), sanitized_value))
            }
        }).collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_security_monitor() {
        let monitor = SecurityMonitor::new();

        // Test syscall checking
        monitor.register_profile("test@1.0.0".to_string(), SecurityMonitor::create_hardened_profile()).await;

        assert!(monitor.is_syscall_allowed("test@1.0.0", "fd_write").await);
        assert!(!monitor.is_syscall_allowed("test@1.0.0", "socket").await);
    }

    #[test]
    fn test_isolation_utils() {
        // Test path safety - use existing paths for reliable testing
        let base = std::path::Path::new("/tmp");
        assert!(IsolationUtils::is_safe_path(base, "/tmp/test.txt"));
        assert!(!IsolationUtils::is_safe_path(base, "/etc/passwd"));
    }

    #[test]
    fn test_env_var_sanitization() {
        let env_vars = vec![
            ("SAFE_VAR".to_string(), "safe_value".to_string()),
            ("LD_LIBRARY_PATH".to_string(), "/dangerous/path".to_string()),
            ("PATH".to_string(), "/bin:/usr/bin".to_string()),
        ];

        let sanitized = IsolationUtils::sanitize_env_vars(&env_vars);
        assert_eq!(sanitized.len(), 1);
        assert_eq!(sanitized[0].0, "SAFE_VAR");
    }
}
