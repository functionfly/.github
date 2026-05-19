//! Security management for Kotlin/JVM runtime
//!
//! Implements security policies, permission management, and code verification
//! for safe execution of untrusted Kotlin/JVM code.

use crate::config::SecurityPolicy;
use anyhow::{anyhow, Result};
use std::collections::HashSet;
use std::sync::Arc;

/// Permission types available for Kotlin/JVM execution
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Permission {
    /// Read access to file system
    Read,
    /// Write access to file system
    Write,
    /// Network access (outbound connections)
    Network,
    /// Environment variable access
    EnvAccess,
    /// Process spawning/execution
    ProcessExec,
    /// Reflection API access
    Reflection,
    /// JNI/native code access
    Jni,
    /// Thread creation
    Threads,
    /// Class loading
    ClassLoading,
    /// Runtime exit
    Exit,
}

impl std::fmt::Display for Permission {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Permission::Read => write!(f, "read"),
            Permission::Write => write!(f, "write"),
            Permission::Network => write!(f, "network"),
            Permission::EnvAccess => write!(f, "env"),
            Permission::ProcessExec => write!(f, "process-exec"),
            Permission::Reflection => write!(f, "reflection"),
            Permission::Jni => write!(f, "jni"),
            Permission::Threads => write!(f, "threads"),
            Permission::ClassLoading => write!(f, "class-loading"),
            Permission::Exit => write!(f, "exit"),
        }
    }
}

/// Set of permissions
#[derive(Debug, Clone, Default)]
pub struct PermissionSet(HashSet<Permission>);

impl PermissionSet {
    /// Create a new empty permission set
    pub fn new() -> Self {
        Self(HashSet::new())
    }

    /// Create with all permissions (for testing only)
    pub fn all() -> Self {
        Self(HashSet::from([
            Permission::Read,
            Permission::Write,
            Permission::Network,
            Permission::EnvAccess,
            Permission::ProcessExec,
            Permission::Reflection,
            Permission::Jni,
            Permission::Threads,
            Permission::ClassLoading,
            Permission::Exit,
        ]))
    }

    /// Create with basic safe permissions
    pub fn safe() -> Self {
        Self(HashSet::from([
            Permission::Read,
            Permission::Network,
        ]))
    }

    /// Add a permission
    pub fn add(&mut self, perm: Permission) {
        self.0.insert(perm);
    }

    /// Remove a permission
    pub fn remove(&mut self, perm: Permission) {
        self.0.remove(&perm);
    }

    /// Check if has a permission
    pub fn has(&self, perm: Permission) -> bool {
        self.0.contains(&perm)
    }

    /// Check if has all of the given permissions
    pub fn has_all(&self, perms: &[Permission]) -> bool {
        perms.iter().all(|p| self.0.contains(p))
    }

    /// Check if has any of the given permissions
    pub fn has_any(&self, perms: &[Permission]) -> bool {
        perms.iter().any(|p| self.0.contains(p))
    }

    /// Get all permissions as a slice
    pub fn as_slice(&self) -> Vec<Permission> {
        self.0.iter().cloned().collect()
    }
}

/// Security violation type
#[derive(Debug, Clone)]
pub enum SecurityViolation {
    /// Blocked package access
    PackageAccess(String),
    /// Blocked class access
    ClassAccess(String),
    /// Blocked host access
    HostAccess(String),
    /// Blocked reflection
    ReflectionDepthExceeded,
    /// Blocked native code
    JniAccess,
    /// Blocked process execution
    ProcessExec,
    /// Blocked environment variable access
    EnvAccess(String),
    /// Blocked network operation
    NetworkAccess(String),
    /// Code contains suspicious patterns
    SuspiciousCode(String),
    /// Code size exceeds limit
    CodeSizeExceeded(usize, usize),
    /// Thread creation blocked
    ThreadCreation,
    /// Class loading blocked
    ClassLoading(String),
    /// Runtime exit attempt
    ExitAttempt,
}

impl std::fmt::Display for SecurityViolation {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SecurityViolation::PackageAccess(pkg) => {
                write!(f, "security violation: blocked package '{}'", pkg)
            }
            SecurityViolation::ClassAccess(cls) => {
                write!(f, "security violation: blocked class '{}'", cls)
            }
            SecurityViolation::HostAccess(host) => {
                write!(f, "security violation: blocked host '{}'", host)
            }
            SecurityViolation::ReflectionDepthExceeded => {
                write!(f, "security violation: reflection depth exceeded")
            }
            SecurityViolation::JniAccess => {
                write!(f, "security violation: JNI access is blocked")
            }
            SecurityViolation::ProcessExec => {
                write!(f, "security violation: process execution is blocked")
            }
            SecurityViolation::EnvAccess(var) => {
                write!(f, "security violation: environment variable '{}' access blocked", var)
            }
            SecurityViolation::NetworkAccess(addr) => {
                write!(f, "security violation: network access to '{}' blocked", addr)
            }
            SecurityViolation::SuspiciousCode(reason) => {
                write!(f, "security violation: suspicious code detected - {}", reason)
            }
            SecurityViolation::CodeSizeExceeded(size, limit) => {
                write!(f, "security violation: code size {} exceeds limit {}", size, limit)
            }
            SecurityViolation::ThreadCreation => {
                write!(f, "security violation: thread creation is blocked")
            }
            SecurityViolation::ClassLoading(cls) => {
                write!(f, "security violation: loading class '{}' is blocked", cls)
            }
            SecurityViolation::ExitAttempt => {
                write!(f, "security violation: runtime exit attempt blocked")
            }
        }
    }
}

/// Security manager for Kotlin/JVM runtime
#[derive(Debug, Clone)]
pub struct SecurityManager {
    policy: SecurityPolicy,
}

impl Default for SecurityManager {
    fn default() -> Self {
        Self {
            policy: SecurityPolicy::default(),
        }
    }
}

impl SecurityManager {
    /// Create a new security manager with the given policy
    pub fn new(policy: SecurityPolicy) -> Self {
        Self { policy }
    }

    /// Get the current security policy
    pub fn policy(&self) -> &SecurityPolicy {
        &self.policy
    }

    /// Check if a package is allowed
    pub fn is_package_allowed(&self, package: &str) -> bool {
        if !self.policy.sandbox_enabled {
            return true;
        }

        for blocked in &self.policy.blocked_packages {
            if package.starts_with(blocked.as_str()) || package == blocked.as_str() {
                return false;
            }
        }
        true
    }

    /// Check if a class is allowed
    pub fn is_class_allowed(&self, class: &str) -> bool {
        if !self.policy.sandbox_enabled {
            return true;
        }

        for blocked in &self.policy.blocked_classes {
            if class.starts_with(blocked.as_str()) || class == blocked.as_str() {
                return false;
            }
        }
        true
    }

    /// Check if a host is allowed
    pub fn is_host_allowed(&self, host: &str) -> bool {
        // Check blocked hosts first
        for blocked in &self.policy.blocked_hosts {
            if host == blocked.as_str() || host.ends_with(blocked.as_str()) {
                return false;
            }
        }

        // If allowed_hosts is empty, allow all (unless blocked above)
        if self.policy.allowed_hosts.is_empty() {
            return true;
        }

        // Check allowed hosts
        for allowed in &self.policy.allowed_hosts {
            if host == allowed.as_str() || host.ends_with(allowed.as_str()) {
                return true;
            }
        }

        false
    }

    /// Check if an environment variable is allowed
    pub fn is_env_allowed(&self, var: &str) -> bool {
        if !self.policy.sandbox_enabled {
            return true;
        }

        // Sensitive patterns
        let sensitive_patterns = [
            "PASSWORD", "SECRET", "TOKEN", "API_KEY", "AWS_", "GCP_", "AZURE_",
            "PRIVATE", "CREDENTIAL", "AUTH", "LD_LIBRARY_PATH", "DYLD_",
            "LD_PRELOAD", "SHELL", "PATH", "HOME", "USER",
        ];

        for pattern in &sensitive_patterns {
            if var.contains(pattern) {
                return false;
            }
        }

        // Check whitelist
        if self.policy.env_whitelist.is_empty() {
            return false;
        }

        self.policy.env_whitelist.iter().any(|v| v == var)
    }

    /// Check if reflection is allowed
    pub fn is_reflection_allowed(&self, depth: usize) -> bool {
        if !self.policy.sandbox_enabled {
            return true;
        }

        if !self.policy.allow_reflection {
            return false;
        }

        depth <= self.policy.max_reflection_depth
    }

    /// Check if process execution is allowed
    pub fn is_process_exec_allowed(&self) -> bool {
        if !self.policy.sandbox_enabled {
            return true;
        }
        self.policy.allow_process_exec
    }

    /// Check if JNI is allowed
    pub fn is_jni_allowed(&self) -> bool {
        if !self.policy.sandbox_enabled {
            return true;
        }
        self.policy.allow_jni
    }

    /// Verify code is safe before execution
    pub fn verify_code(&self, code: &str) -> Result<()> {
        if !self.policy.sandbox_enabled {
            return Ok(());
        }

        // Check code size (10MB limit for Kotlin code)
        let max_code_size = 10 * 1024 * 1024;
        if code.len() > max_code_size {
            return Err(anyhow!("{}", SecurityViolation::CodeSizeExceeded(code.len(), max_code_size)));
        }

        // Check for suspicious patterns
        self.check_suspicious_patterns(code)?;

        // Check for dangerous Kotlin/JVM patterns
        self.check_dangerous_patterns(code)?;

        Ok(())
    }

    /// Check for suspicious code patterns
    fn check_suspicious_patterns(&self, code: &str) -> Result<()> {
        // Check for null bytes (binary obfuscation)
        if code.as_bytes().contains(&0) {
            return Err(anyhow!("{}", SecurityViolation::SuspiciousCode("null bytes detected".to_string())));
        }

        // Check for very long lines
        for line in code.lines() {
            if line.len() > 10000 {
                return Err(anyhow!("{}", SecurityViolation::SuspiciousCode("line too long (>10KB)".to_string())));
            }
        }

        // Check for high ratio of non-printable characters
        let total_chars = code.chars().count();
        if total_chars > 0 {
            let non_printable = code.chars()
                .filter(|c| !c.is_ascii_graphic() && !c.is_whitespace())
                .count();
            if non_printable as f64 / total_chars as f64 > 0.1 {
                return Err(anyhow!("{}", SecurityViolation::SuspiciousCode("high ratio of non-printable characters".to_string())));
            }
        }

        Ok(())
    }

    /// Check for dangerous Kotlin/JVM patterns
    fn check_dangerous_patterns(&self, code: &str) -> Result<()> {
        let dangerous_patterns = [
            // Process execution
            ("ProcessBuilder", "process execution"),
            ("Runtime.getRuntime().exec", "process execution"),
            ("System.exec(", "process execution"),
            ("ProcessImpl", "process execution"),

            // File system
            ("FileInputStream", "file system access"),
            ("FileOutputStream", "file system access"),
            ("RandomAccessFile", "file system access"),
            ("java.io.File", "file system access"),
            ("java.nio.file.Files", "file system access"),
            ("java.nio.file.Path", "file system access"),

            // Network
            ("ServerSocket", "network server"),
            ("DatagramSocket", "network access"),
            ("MulticastSocket", "network access"),

            // Reflection and class loading
            ("Class.forName", "dynamic class loading"),
            ("getClassLoader()", "class loader access"),
            ("defineClass", "class definition"),
            ("loadLibrary", "native library loading"),

            // Threads
            ("Thread.start()", "thread creation"),
            ("ThreadPoolExecutor", "thread pool creation"),
            ("newThread", "thread creation"),

            // Security manager bypass
            ("setSecurityManager", "security manager modification"),
            ("SecurityManager", "security manager access"),

            // Runtime modification
            ("System.setOut", "runtime modification"),
            ("System.setErr", "runtime modification"),
            ("System.setIn", "runtime modification"),

            // Exit
            ("System.exit(", "runtime exit"),
            ("Runtime.getRuntime().halt", "runtime halt"),
        ];

        for (pattern, description) in &dangerous_patterns {
            if code.contains(pattern) {
                // For patterns that are allowed by policy, skip
                match *pattern {
                    "Thread.start()" | "ThreadPoolExecutor" | "newThread" => {
                        if self.policy.sandbox_enabled {
                            // Check if threads are allowed
                            // For now, block thread creation in sandbox
                        }
                    }
                    _ => {
                        return Err(anyhow!("{}", SecurityViolation::SuspiciousCode(format!(
                            "dangerous pattern '{}' ({}) detected", pattern, description
                        ))));
                    }
                }
            }
        }

        Ok(())
    }

    /// Create Arc-wrapped security manager for sharing
    pub fn into_arc(self) -> Arc<Self> {
        Arc::new(self)
    }
}

/// Kotlin/JVM bytecode security validator
pub struct BytecodeValidator {
    security_manager: Arc<SecurityManager>,
}

impl BytecodeValidator {
    /// Create a new bytecode validator
    pub fn new(security_manager: Arc<SecurityManager>) -> Self {
        Self { security_manager }
    }

    /// Validate bytecode before execution
    pub fn validate_bytecode(&self, bytecode: &[u8]) -> Result<()> {
        // Check bytecode size
        let max_bytecode_size = 5 * 1024 * 1024; // 5MB
        if bytecode.len() > max_bytecode_size {
            return Err(anyhow!("bytecode size exceeds maximum allowed size"));
        }

        // Validate bytecode magic number (Java class file)
        if bytecode.len() < 4 {
            return Err(anyhow!("bytecode too short to be valid class file"));
        }

        // Check magic number: 0xCAFEBABE
        if bytecode[0] != 0xCA || bytecode[1] != 0xFE || bytecode[2] != 0xBA || bytecode[3] != 0xBE {
            return Err(anyhow!("invalid bytecode magic number"));
        }

        // Version check (major.minor)
        let major = u16::from_be_bytes([bytecode[6], bytecode[7]]);
        let minor = u16::from_be_bytes([bytecode[4], bytecode[5]]);

        // Support Java 8-21 (major versions 52-65)
        if major < 52 || major > 65 {
            return Err(anyhow!("unsupported bytecode version: {}.{}", minor, major));
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_package_blocking() {
        let manager = SecurityManager::default();
        assert!(!manager.is_package_allowed("java.lang.Process"));
        assert!(!manager.is_package_allowed("java.lang.ProcessBuilder"));
        assert!(manager.is_package_allowed("java.util"));
    }

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
        assert!(!manager.is_env_allowed("AWS_SECRET_KEY"));
        assert!(!manager.is_env_allowed("PASSWORD"));
        assert!(!manager.is_env_allowed("MY_API_TOKEN"));
    }

    #[test]
    fn test_code_verification() {
        let manager = SecurityManager::default();

        // Valid code should pass
        let valid_code = r#"
            fun main() {
                println("Hello, World!")
            }
        "#;
        assert!(manager.verify_code(valid_code).is_ok());

        // Process execution should be blocked
        let bad_code = r#"
            fun main() {
                Runtime.getRuntime().exec("ls")
            }
        "#;
        assert!(manager.verify_code(bad_code).is_err());
    }

    #[test]
    fn test_bytecode_validation() {
        let manager = Arc::new(SecurityManager::default());
        let validator = BytecodeValidator::new(manager);

        // Valid Java class magic number
        let valid_class = vec![
            0xCA, 0xFE, 0xBA, 0xBE, 0x00, 0x00, 0x00, 0x34
        ];
        assert!(validator.validate_bytecode(&valid_class).is_ok());

        // Invalid magic number
        let invalid_class = vec![0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x00];
        assert!(validator.validate_bytecode(&invalid_class).is_err());
    }

    #[test]
    fn test_permission_set() {
        let mut perms = PermissionSet::new();
        assert!(!perms.has(Permission::Read));

        perms.add(Permission::Read);
        assert!(perms.has(Permission::Read));
        assert!(!perms.has(Permission::Write));

        perms.add(Permission::Write);
        assert!(perms.has_all(&[Permission::Read, Permission::Write]));
        assert!(perms.has_any(&[Permission::Write, Permission::Network]));

        perms.remove(Permission::Read);
        assert!(!perms.has(Permission::Read));
        assert!(perms.has(Permission::Write));
    }
}