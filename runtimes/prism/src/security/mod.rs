//! Prism Runtime Security
//!
//! Security policies, permissions, code validation, and sandbox enforcement.

use std::path::Path;
use parking_lot::RwLock;
use regex::Regex;
use sha2::{Sha256, Digest};

pub mod isolation;

#[cfg(feature = "secure-sandbox")]
use wasmparser;

pub use isolation::{
    SecurityAuditor, ExecutionSecurityContext, IsolatedExecutionResult,
    SecurityEvent, SecurityEventType, SecuritySeverity, EnclaveType, EnclaveAttestation,
    EnforceResourceLimits,
};

/// Permission types for execution
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum Permission {
    FileRead,
    FileWrite,
    Network,
    Subprocess,
    NativeExtensions,
    Environment,
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

/// WASM validation result
#[derive(Debug, Clone)]
pub struct WasmValidationResult {
    pub valid: bool,
    pub sanitized: Option<Vec<u8>>,
    pub violations: Vec<WasmViolation>,
    pub code_hash: String,
}

/// WASM violation details
#[derive(Debug, Clone)]
pub struct WasmViolation {
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

/// Security policy for Prism runtime
#[derive(Debug, Clone)]
pub struct SecurityPolicy {
    /// Enable sandbox mode
    pub sandbox_enabled: bool,
    /// Allow filesystem access
    pub allow_filesystem: bool,
    /// Allow network access
    pub allow_network: bool,
    /// Enable seccomp filtering
    pub enable_seccomp: bool,
    /// Enable landlock restrictions
    pub enable_landlock: bool,
    /// Directories to allow access to (empty = none)
    pub allowed_dirs: Vec<String>,
    /// Syscalls to block
    pub blocked_syscalls: Vec<String>,
    /// Require enclave for execution
    pub require_enclave: bool,
    /// Memory limit in bytes
    pub memory_limit_bytes: u64,
    /// CPU time limit in seconds
    pub cpu_time_limit_secs: u64,
}

impl Default for SecurityPolicy {
    fn default() -> Self {
        Self {
            sandbox_enabled: true,
            allow_filesystem: false,
            allow_network: false,
            enable_seccomp: false,
            enable_landlock: false,
            allowed_dirs: vec![],
            blocked_syscalls: vec![
                "ptrace".to_string(),
                "mount".to_string(),
                "umount2".to_string(),
                "syslog".to_string(),
                "init_module".to_string(),
                "delete_module".to_string(),
                "create_module".to_string(),
                "socket".to_string(),
                "capget".to_string(),
                "capset".to_string(),
            ],
            require_enclave: false,
            memory_limit_bytes: 256 * 1024 * 1024, // 256MB
            cpu_time_limit_secs: 30,
        }
    }
}

/// Production-grade security policy with all security features enabled
impl SecurityPolicy {
    pub fn production() -> Self {
        Self {
            sandbox_enabled: true,
            allow_filesystem: false,
            allow_network: false,
            enable_seccomp: true,
            enable_landlock: true,
            allowed_dirs: vec![],
            blocked_syscalls: vec![
                "ptrace".to_string(),
                "mount".to_string(),
                "umount2".to_string(),
                "syslog".to_string(),
                "init_module".to_string(),
                "delete_module".to_string(),
                "create_module".to_string(),
                "socket".to_string(),
                "capget".to_string(),
                "capset".to_string(),
                "process_vm_readv".to_string(),
                "process_vm_writev".to_string(),
                "personality".to_string(),
                "sysinfo".to_string(),
            ],
            require_enclave: false, // Set to true if running in a TEE environment
            memory_limit_bytes: 256 * 1024 * 1024,
            cpu_time_limit_secs: 30,
        }
    }

    pub fn permissive() -> Self {
        Self {
            sandbox_enabled: false,
            allow_filesystem: true,
            allow_network: true,
            enable_seccomp: false,
            enable_landlock: false,
            allowed_dirs: vec!["/tmp".to_string(), "/var/tmp".to_string()],
            blocked_syscalls: vec![],
            require_enclave: false,
            memory_limit_bytes: 512 * 1024 * 1024,
            cpu_time_limit_secs: 300,
        }
    }
}

/// Security manager for Prism runtime
#[derive(Clone)]
pub struct SecurityManager {
    policy: SecurityPolicy,
    permissions: std::sync::Arc<RwLock<PermissionSet>>,
    sandbox_enabled: bool,
    auditor: std::sync::Arc<SecurityAuditor>,
    #[allow(dead_code)]
    /// Regex for path traversal detection
    path_traversal_regex: Regex,
    #[allow(dead_code)]
    /// Regex for code injection patterns
    injection_pattern_regex: Regex,
}

impl SecurityManager {
    /// Create a new security manager from policy
    pub fn new(policy: SecurityPolicy) -> std::sync::Arc<Self> {
        let auditor = std::sync::Arc::new(SecurityAuditor::new());
        let sandbox_enabled = policy.sandbox_enabled;
        let permissions = Self::compute_permissions(&policy);

        let path_traversal_regex = Regex::new(r"(?:\.\.[/\\])+").unwrap();
        let injection_pattern_regex = Regex::new(r#"(?:['"].*\$|\$\{|`[^`]*`)"#).unwrap();

        std::sync::Arc::new(Self {
            policy,
            permissions: std::sync::Arc::new(RwLock::new(permissions)),
            sandbox_enabled,
            auditor,
            path_traversal_regex,
            injection_pattern_regex,
        })
    }

    /// Get the security auditor
    pub fn auditor(&self) -> std::sync::Arc<SecurityAuditor> {
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

    /// Validate WASM module with deep bytecode analysis
    pub fn validate_wasm(&self, wasm_bytes: &[u8]) -> Result<WasmValidationResult, SecurityError> {
        let mut violations = Vec::new();
        let sanitized = wasm_bytes.to_vec();

        // Calculate code hash
        let mut hasher = Sha256::new();
        hasher.update(wasm_bytes);
        let result = hasher.finalize();
        let code_hash = hex::encode(result);

        // Validate WASM magic number
        if wasm_bytes.len() < 4 {
            violations.push(WasmViolation {
                pattern: "wasm_header".to_string(),
                severity: ViolationSeverity::Critical,
                description: "WASM module too short".to_string(),
                position: Some((0, wasm_bytes.len())),
            });
        } else if &wasm_bytes[0..4] != b"\0asm" {
            violations.push(WasmViolation {
                pattern: "wasm_magic".to_string(),
                severity: ViolationSeverity::Critical,
                description: "Invalid WASM magic number".to_string(),
                position: Some((0, 4)),
            });
        }

        // Deep WASM bytecode validation using wasmparser
        #[cfg(feature = "secure-sandbox")]
        {
            self.validate_wasm_with_parser(wasm_bytes, &mut violations);
        }

        #[cfg(not(feature = "secure-sandbox"))]
        {
            // Without secure-sandbox, only check dangerous imports by name patterns
            self.check_dangerous_imports_by_pattern(wasm_bytes, &mut violations);
        }

        let has_critical = violations.iter().any(|v| v.severity == ViolationSeverity::Critical);
        let has_high = violations.iter().any(|v| v.severity == ViolationSeverity::High);

        let valid = !has_critical && !has_high;

        if !valid {
            self.auditor.log_blocked_code(&code_hash, "WASM validation failed", None);
        }

        Ok(WasmValidationResult {
            valid,
            sanitized: if valid { Some(sanitized) } else { None },
            violations,
            code_hash,
        })
    }

    #[cfg(feature = "secure-sandbox")]
    fn validate_wasm_with_parser(&self, wasm_bytes: &[u8], violations: &mut Vec<WasmViolation>) {
        use wasmparser::Parser;

        let parser = Parser::new(0);
        for payload in parser.parse_all(wasm_bytes) {
            match payload {
                Ok(wasmparser::Payload::Version { num, encoding: _, range: _ }) => {
                    if num != 1 && num != 2 {
                        violations.push(WasmViolation {
                            pattern: "wasm_version".to_string(),
                            severity: ViolationSeverity::High,
                            description: format!("Unsupported WASM version: {}", num),
                            position: None,
                        });
                    }
                }
                Ok(wasmparser::Payload::ImportSection(s)) => {
                    for i in s {
                        if let Ok(import) = i {
                            let module = import.module;
                            let field = import.name;

                            let dangerous = match (module, field) {
                                ("env", "ptrace") => ("Process introspection via ptrace", ViolationSeverity::Critical),
                                ("env", "mount") => ("Filesystem mounting", ViolationSeverity::Critical),
                                ("env", "syslog") => ("System logging", ViolationSeverity::Critical),
                                ("env", "init_module") => ("Kernel module loading", ViolationSeverity::Critical),
                                ("env", "delete_module") => ("Kernel module unloading", ViolationSeverity::Critical),
                                ("env", "create_module") => ("Kernel module creation", ViolationSeverity::Critical),
                                ("env", "socket") => ("Raw socket creation", ViolationSeverity::High),
                                ("env", "capget") => ("Capability reading", ViolationSeverity::High),
                                ("env", "capset") => ("Capability setting", ViolationSeverity::High),
                                ("wasi_snapshot_preview1", "fd_write") => {
                                    ("File descriptor write via WASI", ViolationSeverity::Low)
                                }
                                ("env", "abort") => ("Process abort", ViolationSeverity::Medium),
                                ("env", "exit") => ("Process exit", ViolationSeverity::Medium),
                                ("env", "raise") => ("Signal raise", ViolationSeverity::Medium),
                                _ => continue,
                            };

                            violations.push(WasmViolation {
                                pattern: format!("{}.{}", module, field),
                                severity: dangerous.1,
                                description: dangerous.0.to_string(),
                                position: None,
                            });
                        }
                    }
                }
                Err(e) => {
                    violations.push(WasmViolation {
                        pattern: "parse_error".to_string(),
                        severity: ViolationSeverity::Critical,
                        description: format!("WASM parse error: {}", e),
                        position: None,
                    });
                }
                _ => {}
            }
        }
    }

    #[cfg(not(feature = "secure-sandbox"))]
    fn check_dangerous_imports_by_pattern(&self, wasm_bytes: &[u8], _violations: &mut Vec<WasmViolation>) {
        // Lightweight check without wasmparser - just scan for dangerous strings
        let data = String::from_utf8_lossy(wasm_bytes);

        // These are rough heuristics - not real parsing
        let dangerous = [
            ("ptrace", "Process introspection"),
            ("mount", "Filesystem mounting"),
            ("syslog", "System logging"),
        ];

        for (pattern, desc) in dangerous {
            if data.contains(&format!("\"env\"\0\"{}\"", pattern)) {
                // Potential dangerous import - but without wasmparser we can't be sure
                // Log a warning-level violation since we can't confirm
                _violations.push(WasmViolation {
                    pattern: pattern.to_string(),
                    severity: ViolationSeverity::Warning,
                    description: format!("Potential {} detected (enable secure-sandbox for deep analysis)", desc),
                    position: None,
                });
            }
        }
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
    pub fn create_execution_context(&self) -> ExecutionSecurityContext {
        let mut ctx = ExecutionSecurityContext::new(
            self.policy.sandbox_enabled,
            self.policy.enable_seccomp,
            self.policy.enable_landlock,
            self.policy.memory_limit_bytes,
            self.policy.cpu_time_limit_secs,
            self.policy.allowed_dirs.iter().map(Path::new).map(|p| p.to_path_buf()).collect(),
            self.policy.allow_network,
        );

        ctx.enclave_type = ExecutionSecurityContext::detect_enclave();
        ctx.require_enclave = self.policy.require_enclave;

        ctx
    }

    /// Verify enclave attestation
    pub fn verify_enclave_attestation(&self, ctx: &ExecutionSecurityContext) -> Result<EnclaveAttestation, SecurityError> {
        let attestation = ctx.get_attestation()
            .map_err(|e| SecurityError::AttestationError(e))?;

        if self.policy.require_enclave && attestation.enclave_type == EnclaveType::None {
            self.auditor.log_enclave_attestation_failed("Enclave required but not available");
            return Err(SecurityError::EnclaveNotAvailable);
        }

        if self.policy.require_enclave && !attestation.is_attested {
            self.auditor.log_enclave_attestation_failed("Enclave not attested");
            return Err(SecurityError::AttestationError("Attestation failed".to_string()));
        }

        Ok(attestation)
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

    #[error("enclave not available")]
    EnclaveNotAvailable,

    #[error("attestation error: {0}")]
    AttestationError(String),
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_policy() -> SecurityPolicy {
        SecurityPolicy {
            sandbox_enabled: true,
            allow_filesystem: false,
            allow_network: false,
            enable_seccomp: true,
            enable_landlock: true,
            ..Default::default()
        }
    }

    #[test]
    fn test_wasm_validation_accepts_valid_wasm() {
        let manager = SecurityManager::new(test_policy());
        // Valid WASM magic number
        let valid_wasm = b"\0asm\x01\x00\x00\x00";
        let result = manager.validate_wasm(valid_wasm);
        assert!(result.is_ok());
        let validation = result.unwrap();
        assert!(validation.valid);
    }

    #[test]
    fn test_wasm_validation_rejects_invalid_magic() {
        let manager = SecurityManager::new(test_policy());
        let invalid_wasm = b"\x00\x00\x00\x00\x01\x00\x00\x00";
        let result = manager.validate_wasm(invalid_wasm);
        assert!(result.is_ok());
        let validation = result.unwrap();
        assert!(!validation.valid);
        assert!(!validation.violations.is_empty());
    }

    #[test]
    fn test_enclave_detection() {
        let manager = SecurityManager::new(test_policy());
        let ctx = manager.create_execution_context();
        let enclave = ctx.enclave_type;
        // Just verify it returns a valid type
        assert!(matches!(enclave, EnclaveType::None | EnclaveType::Sgx | EnclaveType::Sev | EnclaveType::TrustZone | EnclaveType::Tee));
    }

    #[test]
    fn test_code_hash_calculation() {
        let manager = SecurityManager::new(test_policy());
        let wasm1 = b"\0asm\x01\x00\x00\x00";
        let wasm2 = b"\0asm\x01\x00\x00\x01";

        let result1 = manager.validate_wasm(wasm1).unwrap();
        let result2 = manager.validate_wasm(wasm2).unwrap();

        assert_ne!(result1.code_hash, result2.code_hash);
    }
}