//! Ruby Runtime Security
//!
//! Security policies, permissions, code validation, and sandbox enforcement.

use crate::config::SecurityPolicy;
use anyhow::Result;
use std::sync::Arc;
use std::path::Path;
use parking_lot::RwLock;
use regex::Regex;
use tracing::{error, debug};
use sha2::{Sha256, Digest};

pub mod isolation;

pub use isolation::{SecurityAuditor, ExecutionSecurityContext, IsolatedExecutionResult, SecurityEvent, SecurityEventType, SecuritySeverity};

/// Permission types for Ruby execution
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum Permission {
    FileRead,
    FileWrite,
    Network,
    Subprocess,
    NativeExtensions,
    Environment,
    Eval,
    Syscalls,
}

impl Permission {
    pub fn name(&self) -> &'static str {
        match self {
            Permission::FileRead => "file_read",
            Permission::FileWrite => "file_write",
            Permission::Network => "network",
            Permission::Subprocess => "subprocess",
            Permission::NativeExtensions => "native_extensions",
            Permission::Environment => "environment",
            Permission::Eval => "eval",
            Permission::Syscalls => "syscalls",
        }
    }
}

/// Set of permissions
#[derive(Debug, Clone, Default)]
pub struct PermissionSet(Vec<Permission>);

impl PermissionSet {
    pub fn new() -> Self {
        Self(vec![])
    }

    pub fn permissive() -> Self {
        Self(vec![
            Permission::FileRead,
            Permission::FileWrite,
            Permission::Network,
            Permission::Subprocess,
            Permission::NativeExtensions,
            Permission::Environment,
            Permission::Eval,
            Permission::Syscalls,
        ])
    }

    pub fn restrictive() -> Self {
        Self(vec![])
    }

    pub fn add(&mut self, perm: Permission) {
        if !self.0.contains(&perm) {
            self.0.push(perm);
        }
    }

    pub fn contains(&self, perm: &Permission) -> bool {
        self.0.contains(perm)
    }

    pub fn get_all(&self) -> &[Permission] {
        &self.0
    }
}

/// Code validation result
#[derive(Debug, Clone)]
pub struct CodeValidationResult {
    pub valid: bool,
    pub sanitized: Option<String>,
    pub violations: Vec<CodeViolation>,
    pub code_hash: String,
}

/// Code violation details
#[derive(Debug, Clone)]
pub struct CodeViolation {
    pub pattern: String,
    pub severity: ViolationSeverity,
    pub description: String,
    pub position: Option<(usize, usize)>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ViolationSeverity {
    Critical,
    High,
    Medium,
    Low,
    Warning,
}

/// Security manager for Ruby runtime
#[derive(Clone)]
pub struct SecurityManager {
    policy: SecurityPolicy,
    permissions: Arc<RwLock<PermissionSet>>,
    sandbox_enabled: bool,
    auditor: Arc<SecurityAuditor>,
    /// Compiled patterns for code validation
    dangerous_patterns: Vec<DangerousPattern>,
    /// Regex for path traversal detection
    path_traversal_regex: Regex,
    /// Regex for code injection patterns
    injection_pattern_regex: Regex,
}

#[derive(Clone)]
struct DangerousPattern {
    pattern: &'static str,
    severity: ViolationSeverity,
    description: &'static str,
    replacement: Option<&'static str>,
}

impl SecurityManager {
    /// Create a new security manager from policy
    pub fn new(policy: SecurityPolicy) -> Arc<Self> {
        let auditor = Arc::new(SecurityAuditor::new());
        let sandbox_enabled = policy.sandbox_enabled;
        let permissions = Self::compute_permissions(&policy);

        let dangerous_patterns = vec![
            // Process creation - CRITICAL
            DangerousPattern {
                pattern: r#"\beval\s*"#,
                severity: ViolationSeverity::Critical,
                description: "eval() allows arbitrary code execution",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bexec\s*"#,
                severity: ViolationSeverity::Critical,
                description: "exec() replaces current process",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bsystem\s*"#,
                severity: ViolationSeverity::High,
                description: "system() executes shell commands",
                replacement: Some("puts 'system() is disabled'"),
            },
            DangerousPattern {
                pattern: r#"`"#,
                severity: ViolationSeverity::Critical,
                description: "Backtick executes shell commands",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bspawn\s*"#,
                severity: ViolationSeverity::High,
                description: "spawn() creates new processes",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bfork\s*"#,
                severity: ViolationSeverity::Critical,
                description: "fork() creates process copy",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bProcess\.(spawn|detach|kill|wait)\s*"#,
                severity: ViolationSeverity::High,
                description: "Process methods can manipulate other processes",
                replacement: None,
            },
            // File system - HIGH
            DangerousPattern {
                pattern: r#"\bFile\.delete\s*"#,
                severity: ViolationSeverity::High,
                description: "File.delete() removes files",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bFile\.rename\s*"#,
                severity: ViolationSeverity::High,
                description: "File.rename() moves/overwrites files",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bDir\.rmdir\s*"#,
                severity: ViolationSeverity::High,
                description: "Dir.rmdir() removes directories",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bDir\.unlink\s*"#,
                severity: ViolationSeverity::High,
                description: "Dir.unlink() removes directories",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bFile\.chmod\s*"#,
                severity: ViolationSeverity::Medium,
                description: "File.chmod() changes permissions",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bFile\.chown\s*"#,
                severity: ViolationSeverity::Medium,
                description: "File.chown() changes ownership",
                replacement: None,
            },
            // Network - HIGH
            DangerousPattern {
                pattern: r#"\bSocket\.open\s*"#,
                severity: ViolationSeverity::High,
                description: "Socket.open() creates network connections",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bNet::HTTP\.start\s*"#,
                severity: ViolationSeverity::High,
                description: "HTTP connections to remote servers",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bOpenSSL::SSL"#,
                severity: ViolationSeverity::High,
                description: "SSL/TLS connections",
                replacement: None,
            },
            // Environment - MEDIUM
            DangerousPattern {
                pattern: r#"\bENV\s*\[="#,
                severity: ViolationSeverity::Medium,
                description: "Environment variable access",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bargv\s*"#,
                severity: ViolationSeverity::Medium,
                description: "Command line arguments",
                replacement: None,
            },
            // Kernel/Proc - CRITICAL
            DangerousPattern {
                pattern: r#"/proc/"#,
                severity: ViolationSeverity::Critical,
                description: "Access to /proc filesystem",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\breadlink\s*"#,
                severity: ViolationSeverity::High,
                description: "readlink() can read sensitive symlinks",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bsymlink\s*"#,
                severity: ViolationSeverity::High,
                description: "symlink() creates symbolic links",
                replacement: None,
            },
            // Path traversal - CRITICAL
            DangerousPattern {
                pattern: r#"\.\.\/"#,
                severity: ViolationSeverity::Critical,
                description: "Path traversal attempt",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\.\.\\"#,
                severity: ViolationSeverity::Critical,
                description: "Windows path traversal attempt",
                replacement: None,
            },
            // Null byte injection
            // Note: regex crate doesn't support \0 as a backreference escape,
            // so we use \\x00 which matches a literal NUL byte (0x00).
            DangerousPattern {
                pattern: r#"\x00"#,
                severity: ViolationSeverity::Critical,
                description: "Null byte injection",
                replacement: None,
            },
            // Marshal load - CRITICAL (arbitrary object deserialization)
            DangerousPattern {
                pattern: r#"\bMarshal\.load\s*"#,
                severity: ViolationSeverity::Critical,
                description: "Marshal.load() can deserialize arbitrary objects",
                replacement: None,
            },
            DangerousPattern {
                pattern: r#"\bMarshal\.dump\s*"#,
                severity: ViolationSeverity::High,
                description: "Marshal.dump() serializes objects",
                replacement: None,
            },
            // YAML load - HIGH
            DangerousPattern {
                pattern: r#"\bYAML\.load\s*"#,
                severity: ViolationSeverity::High,
                description: "YAML.load() can execute arbitrary code",
                replacement: Some("YAML.safe_load()"),
            },
            DangerousPattern {
                pattern: r#"\bPsych\.load\s*"#,
                severity: ViolationSeverity::High,
                description: "Psych.load() is the same as YAML.load()",
                replacement: Some("Psych.safe_load()"),
            },
            // Nokogiri XXE prevention
            DangerousPattern {
                pattern: r#"\bNokogiri::HTML\.parse\s*\(.*fetch"#,
                severity: ViolationSeverity::High,
                description: "Potential XXE in Nokogiri",
                replacement: None,
            },
            // Ractor - potential DoS
            DangerousPattern {
                pattern: r#"\bRactor\.new\s*"#,
                severity: ViolationSeverity::Medium,
                description: "Ractor creates new actors (potential DoS)",
                replacement: None,
            },
            // Thread - potential DoS
            DangerousPattern {
                pattern: r#"\bThread\.new\s*"#,
                severity: ViolationSeverity::Medium,
                description: "Thread creation (potential DoS)",
                replacement: None,
            },
        ];

        let path_traversal_regex = Regex::new(r#"(?:\.\.[/\\])+"#).unwrap();
        let injection_pattern_regex = Regex::new(r#"(?:['"`].*\$|\$\{|`[^`]*`)"#).unwrap();

        Arc::new(Self {
            policy,
            permissions: Arc::new(RwLock::new(permissions)),
            sandbox_enabled,
            auditor,
            dangerous_patterns,
            path_traversal_regex,
            injection_pattern_regex,
        })
    }

    /// Get the security auditor
    pub fn auditor(&self) -> Arc<SecurityAuditor> {
        self.auditor.clone()
    }

    /// Compute permissions from security policy
    fn compute_permissions(policy: &SecurityPolicy) -> PermissionSet {
        let mut perms = PermissionSet::new();

        if policy.allow_filesystem {
            perms.add(Permission::FileRead);
            perms.add(Permission::FileWrite);
        }

        if policy.allow_network {
            perms.add(Permission::Network);
        }

        perms
    }

    /// Check if a permission is allowed
    pub fn check_permission(&self, perm: &Permission) -> bool {
        if !self.sandbox_enabled {
            return true;
        }
        self.permissions.read().contains(perm)
    }

    /// Check if file access is allowed
    pub fn can_read_file(&self, path: &Path) -> bool {
        if !self.sandbox_enabled {
            return true;
        }

        if self.policy.allowed_dirs.is_empty() {
            return false;
        }

        for dir in &self.policy.allowed_dirs {
            let allowed = Path::new(dir);
            if path.starts_with(allowed) {
                return true;
            }
        }

        false
    }

    /// Check if network access is allowed
    pub fn can_access_network(&self) -> bool {
        if !self.sandbox_enabled {
            return true;
        }
        self.check_permission(&Permission::Network)
    }

    /// Validate and sanitize code - returns error if code is dangerous
    pub fn validate_code(&self, code: &str) -> Result<String, SecurityError> {
        if !self.policy.sanitize_code {
            return Ok(code.to_string());
        }

        let validation = self.validate_code_detailed(code)?;

        if !validation.valid {
            let violations: Vec<String> = validation
                .violations
                .iter()
                .map(|v| format!("{}: {}", v.pattern, v.description))
                .collect();

            return Err(SecurityError::DangerousPattern(format!(
                "Code blocked due to security violations: {}",
                violations.join("; ")
            )));
        }

        Ok(validation.sanitized.unwrap_or_else(|| code.to_string()))
    }

    /// Validate code and return detailed results
    pub fn validate_code_detailed(&self, code: &str) -> Result<CodeValidationResult, SecurityError> {
        let mut violations = Vec::new();
        let mut sanitized = code.to_string();

        // Calculate code hash
        let mut hasher = Sha256::new();
        hasher.update(code.as_bytes());
        let result = hasher.finalize();
        let code_hash = hex::encode(result);

        // Check each dangerous pattern
        for pattern in &self.dangerous_patterns {
            let regex = Regex::new(pattern.pattern)
                .map_err(|e| SecurityError::ValidationError(e.to_string()))?;

            if regex.is_match(code) {
                // Find matches for position
                let matches: Vec<_> = regex.find_iter(code).collect();
                let positions: Vec<(usize, usize)> = matches
                    .iter()
                    .map(|m| (m.start(), m.end()))
                    .collect();

                violations.push(CodeViolation {
                    pattern: pattern.pattern.to_string(),
                    severity: pattern.severity,
                    description: pattern.description.to_string(),
                    position: positions.first().copied(),
                });

                // Apply replacement if available
                if let Some(replacement) = pattern.replacement {
                    sanitized = regex.replace_all(&sanitized, replacement).to_string();
                }
            }
        }

        // Check for path traversal
        if self.path_traversal_regex.is_match(code) {
            violations.push(CodeViolation {
                pattern: r"\.\.[/\\]".to_string(),
                severity: ViolationSeverity::Critical,
                description: "Path traversal detected".to_string(),
                position: None,
            });
        }

        // Check for code injection
        if self.injection_pattern_regex.is_match(code) {
            violations.push(CodeViolation {
                pattern: r#"['"].*\$|\$\{|`[^`]*`"#.to_string(),
                severity: ViolationSeverity::High,
                description: "Potential code injection via string interpolation".to_string(),
                position: None,
            });
        }

        // Check recursion depth via simple heuristic
        let mut depth: i32 = 0;
        let mut max_depth: i32 = 0;
        for ch in code.chars() {
            match ch {
                '{' | '(' | '[' => {
                    depth += 1;
                    max_depth = max_depth.max(depth);
                }
                '}' | ')' | ']' => {
                    depth = depth.saturating_sub(1);
                }
                _ => {}
            }
        }

        if max_depth > self.policy.max_require_depth as i32 {
            violations.push(CodeViolation {
                pattern: "excessive_nesting".to_string(),
                severity: ViolationSeverity::Medium,
                description: format!(
                    "Excessive nesting depth ({}) exceeds limit ({})",
                    max_depth, self.policy.max_require_depth
                ),
                position: None,
            });
        }

        // Determine if code is valid
        let has_critical = violations.iter().any(|v| v.severity == ViolationSeverity::Critical);
        let has_high = violations.iter().any(|v| v.severity == ViolationSeverity::High);

        let valid = !has_critical && !has_high;

        if !valid {
            // Log the security event
            let reasons: Vec<&str> = violations
                .iter()
                .filter(|v| v.severity == ViolationSeverity::Critical || v.severity == ViolationSeverity::High)
                .map(|v| v.description.as_str())
                .collect();

            self.auditor.log_blocked_code(&code_hash, &reasons.join("; "), None);
        }

        Ok(CodeValidationResult {
            valid,
            sanitized: if valid { Some(sanitized) } else { None },
            violations,
            code_hash,
        })
    }

    /// Legacy sanitize_code for backward compatibility
    pub fn sanitize_code(&self, code: &str) -> Result<String, SecurityError> {
        self.validate_code(code)
    }

    /// Check if a syscall is allowed
    pub fn is_syscall_allowed(&self, syscall: &str) -> bool {
        if !self.sandbox_enabled {
            return true;
        }

        if self.policy.blocked_syscalls.contains(&syscall.to_string()) {
            return false;
        }

        true
    }

    /// Get current policy
    pub fn policy(&self) -> &SecurityPolicy {
        &self.policy
    }

    /// Check if sandbox is enabled
    pub fn is_sandbox_enabled(&self) -> bool {
        self.sandbox_enabled
    }

    /// Create execution security context for this policy
    pub fn create_execution_context(&self, memory_limit_bytes: u64, cpu_time_secs: u64) -> ExecutionSecurityContext {
        ExecutionSecurityContext::new(
            self.policy.sandbox_enabled,
            self.policy.enable_seccomp,
            self.policy.enable_landlock,
            memory_limit_bytes,
            cpu_time_secs,
            self.policy.allowed_dirs.iter().map(Path::new).map(|p| p.to_path_buf()).collect(),
            self.policy.allow_network,
        )
    }
}

/// Security error types
#[derive(Debug, thiserror::Error)]
pub enum SecurityError {
    #[error("permission denied: {0}")]
    PermissionDenied(String),

    #[error("dangerous pattern detected: {0}")]
    DangerousPattern(String),

    #[error("syscall blocked: {0}")]
    SyscallBlocked(String),

    #[error("path not allowed: {0}")]
    PathNotAllowed(String),

    #[error("validation error: {0}")]
    ValidationError(String),
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_policy() -> SecurityPolicy {
        SecurityPolicy {
            sandbox_enabled: true,
            allow_filesystem: false,
            allow_network: false,
            sanitize_code: true,
            ..Default::default()
        }
    }

    #[test]
    fn test_restrictive_permissions() {
        let manager = SecurityManager::new(test_policy());
        assert!(!manager.can_read_file(Path::new("/etc/passwd")));
    }

    #[test]
    fn test_code_validation_allows_safe_code() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code("puts 'hello world'");
        assert!(result.is_ok());
    }

    #[test]
    fn test_code_validation_blocks_eval() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code("eval('puts 1')");
        assert!(result.is_err());
        let err = result.unwrap_err();
        assert!(matches!(err, SecurityError::DangerousPattern(_)));
    }

    #[test]
    fn test_code_validation_blocks_exec() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code("exec('ls')");
        assert!(result.is_err());
    }

    #[test]
    fn test_code_validation_blocks_system() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code("system('ls')");
        assert!(result.is_err());
    }

    #[test]
    fn test_code_validation_blocks_path_traversal() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code("File.read('../../../etc/passwd')");
        assert!(result.is_err());
    }

    #[test]
    fn test_code_validation_blocks_backtick() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code("`ls -la`");
        assert!(result.is_err());
    }

    #[test]
    fn test_code_validation_allows_safe_file_operations() {
        let mut policy = test_policy();
        policy.allow_filesystem = true;
        let manager = SecurityManager::new(policy);

        // Should allow reading safe files when filesystem is permitted.
        // `/safe/file.txt` does not contain `../`, so it must be allowed.
        let result = manager.validate_code("File.read('/safe/file.txt')");
        assert!(result.is_ok(), "expected safe file path to be allowed, got: {:?}", result);
    }

    #[test]
    fn test_code_validation_detailed() {
        let manager = SecurityManager::new(test_policy());
        let result = manager.validate_code_detailed("puts 'hello'").unwrap();
        assert!(result.valid);
        assert!(result.violations.is_empty());
    }

    #[test]
    fn test_code_hash_calculation() {
        let manager = SecurityManager::new(test_policy());
        let result1 = manager.validate_code_detailed("puts 'hello'").unwrap();
        let result2 = manager.validate_code_detailed("puts 'hello'").unwrap();
        let result3 = manager.validate_code_detailed("puts 'world'").unwrap();

        assert_eq!(result1.code_hash, result2.code_hash);
        assert_ne!(result1.code_hash, result3.code_hash);
    }
}
