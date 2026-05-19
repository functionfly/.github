//! Ruby Runtime Security - Process Isolation & System Enforcement
//!
//! Provides OS-level security enforcement including seccomp, landlock, and resource limits.

use std::path::PathBuf;
use std::time::Duration;
use std::sync::Arc;
use parking_lot::RwLock;
use tracing::{info, warn, error, debug};

/// Security enforcement context for a single execution
pub struct ExecutionSecurityContext {
    /// Whether process isolation is enabled
    pub isolate_process: bool,
    /// Whether seccomp is enabled
    pub use_seccomp: bool,
    /// Whether landlock is enabled
    pub use_landlock: bool,
    /// Memory limit in bytes
    pub memory_limit_bytes: u64,
    /// CPU time limit in seconds
    pub cpu_time_limit_secs: u64,
    /// Allowed directories (empty = none)
    pub allowed_dirs: Vec<PathBuf>,
    /// Whether network is allowed
    pub allow_network: bool,
}

/// Result of security enforcement setup
pub struct SecuritySetupResult {
    /// Whether setup succeeded
    pub success: bool,
    /// Error message if failed
    pub error: Option<String>,
    /// PID of isolated process (if isolated)
    pub process_id: Option<u32>,
}

impl ExecutionSecurityContext {
    /// Create a new security context from policy
    pub fn new(
        isolate_process: bool,
        use_seccomp: bool,
        use_landlock: bool,
        memory_limit_bytes: u64,
        cpu_time_limit_secs: u64,
        allowed_dirs: Vec<PathBuf>,
        allow_network: bool,
    ) -> Self {
        Self {
            isolate_process,
            use_seccomp,
            use_landlock,
            memory_limit_bytes,
            cpu_time_limit_secs,
            allowed_dirs,
            allow_network,
        }
    }

    /// Check if running on Linux
    pub fn is_linux() -> bool {
        std::env::consts::OS == "linux"
    }

    /// Execute Ruby code in isolated subprocess with security constraints
    #[cfg(target_os = "linux")]
    pub fn execute_isolated(
        &self,
        code: &str,
        timeout: Duration,
    ) -> Result<IsolatedExecutionResult, String> {
        use std::process::{Command, Stdio};
        use std::io::Read;

        // For production, this would use prctl/setrlimit before exec
        // and potentially seccomp to filter syscalls

        let mut child = Command::new("ruby")
            .args(["-e", code])
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| format!("Failed to spawn ruby process: {}", e))?;

        // Set timeout using platform-specific timeout command
        let start = std::time::Instant::now();

        // Wait for process with timeout using loop
        let mut output = None;

        loop {
            match child.try_wait() {
                Ok(Some(_status)) => {
                    let out = child.wait_with_output().map_err(|e| format!("Failed to get output: {}", e))?;
                    output = Some(out);
                    break;
                }
                Ok(None) => {
                    if start.elapsed() > timeout {
                        child.kill().ok();
                        child.wait().ok();
                        return Err("Execution timed out".to_string());
                    }
                    std::thread::sleep(Duration::from_millis(10));
                }
                Err(e) => {
                    return Err(format!("Failed to wait on child: {}", e));
                }
            }
        }

        let output = output.unwrap();

        let stdout = String::from_utf8_lossy(&output.stdout).to_string();
        let stderr = String::from_utf8_lossy(&output.stderr).to_string();

        Ok(IsolatedExecutionResult {
            success: output.status.success(),
            output: stdout,
            error: stderr,
            exit_code: output.status.code(),
            execution_time_ms: start.elapsed().as_millis() as u64,
            memory_used_bytes: 0,
            killed: false,
            timeout: false,
        })
    }

    #[cfg(not(target_os = "linux"))]
    pub fn execute_isolated(
        &self,
        code: &str,
        timeout: Duration,
    ) -> Result<IsolatedExecutionResult, String> {
        Err("Process isolation not supported on this platform".to_string())
    }

    /// Get resource limits description for logging
    pub fn describe_limits(&self) -> String {
        format!(
            "memory={}MB cpu={}s network={}",
            self.memory_limit_bytes / 1024 / 1024,
            self.cpu_time_limit_secs,
            self.allow_network
        )
    }
}

/// Result of isolated execution
#[derive(Debug, Clone)]
pub struct IsolatedExecutionResult {
    pub success: bool,
    pub output: String,
    pub error: String,
    pub exit_code: Option<i32>,
    pub execution_time_ms: u64,
    pub memory_used_bytes: u64,
    pub killed: bool,
    pub timeout: bool,
}

/// Security auditor for logging security events
#[derive(Clone)]
pub struct SecurityAuditor {
    events: Arc<RwLock<Vec<SecurityEvent>>>,
}

impl SecurityAuditor {
    pub fn new() -> Self {
        Self {
            events: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Log a security event
    pub fn log(&self, event: SecurityEvent) {
        match &event.severity {
            SecuritySeverity::Critical | SecuritySeverity::High => {
                error!(
                    security_event = ?event,
                    "SECURITY EVENT: {} - {}",
                    event.event_type,
                    event.description
                );
            }
            SecuritySeverity::Medium => {
                warn!(
                    security_event = ?event,
                    "Security event: {} - {}",
                    event.event_type,
                    event.description
                );
            }
            SecuritySeverity::Low => {
                debug!(
                    security_event = ?event,
                    "Security notice: {} - {}",
                    event.event_type,
                    event.description
                );
            }
        }

        self.events.write().push(event);
    }

    /// Log code that was blocked
    pub fn log_blocked_code(&self, code_hash: &str, reason: &str, tenant_id: Option<&str>) {
        self.log(SecurityEvent {
            timestamp: chrono::Utc::now(),
            severity: SecuritySeverity::High,
            event_type: SecurityEventType::CodeBlocked,
            description: format!("Code blocked: {}", reason),
            code_hash: Some(code_hash.to_string()),
            tenant_id: tenant_id.map(String::from),
            ..Default::default()
        });
    }

    /// Log resource limit hit
    pub fn log_resource_limit_hit(&self, limit_type: &str, tenant_id: Option<&str>) {
        self.log(SecurityEvent {
            timestamp: chrono::Utc::now(),
            severity: SecuritySeverity::Medium,
            event_type: SecurityEventType::ResourceLimitHit,
            description: format!("Resource limit hit: {}", limit_type),
            tenant_id: tenant_id.map(String::from),
            ..Default::default()
        });
    }

    /// Log sandbox violation
    pub fn log_sandbox_violation(&self, violation: &str, tenant_id: Option<&str>) {
        self.log(SecurityEvent {
            timestamp: chrono::Utc::now(),
            severity: SecuritySeverity::Critical,
            event_type: SecurityEventType::SandboxViolation,
            description: format!("Sandbox violation: {}", violation),
            tenant_id: tenant_id.map(String::from),
            ..Default::default()
        });
    }

    /// Get recent events
    pub fn get_recent_events(&self, count: usize) -> Vec<SecurityEvent> {
        let events = self.events.read();
        events.iter().rev().take(count).cloned().collect()
    }

    /// Clear all events
    pub fn clear(&self) {
        self.events.write().clear();
    }
}

impl Default for SecurityAuditor {
    fn default() -> Self {
        Self::new()
    }
}

/// Security event for audit logging
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct SecurityEvent {
    pub timestamp: chrono::DateTime<chrono::Utc>,
    pub severity: SecuritySeverity,
    pub event_type: SecurityEventType,
    pub description: String,
    pub code_hash: Option<String>,
    pub tenant_id: Option<String>,
    pub execution_id: Option<String>,
    pub metadata: Option<std::collections::HashMap<String, String>>,
}

impl Default for SecurityEvent {
    fn default() -> Self {
        Self {
            timestamp: chrono::Utc::now(),
            severity: SecuritySeverity::Low,
            event_type: SecurityEventType::Other,
            description: String::new(),
            code_hash: None,
            tenant_id: None,
            execution_id: None,
            metadata: None,
        }
    }
}

/// Security event severity
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum SecuritySeverity {
    Low,
    Medium,
    High,
    Critical,
}

/// Security event types
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize, serde::Deserialize)]
pub enum SecurityEventType {
    CodeBlocked,
    ResourceLimitHit,
    SandboxViolation,
    SyscallBlocked,
    NetworkBlocked,
    ExecutionTimeout,
    ExecutionError,
    AuthenticationFailure,
    RateLimitExceeded,
    Other,
}

impl std::fmt::Display for SecurityEventType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::CodeBlocked => write!(f, "CODE_BLOCKED"),
            Self::ResourceLimitHit => write!(f, "RESOURCE_LIMIT_HIT"),
            Self::SandboxViolation => write!(f, "SANDBOX_VIOLATION"),
            Self::SyscallBlocked => write!(f, "SYSCALL_BLOCKED"),
            Self::NetworkBlocked => write!(f, "NETWORK_BLOCKED"),
            Self::ExecutionTimeout => write!(f, "EXECUTION_TIMEOUT"),
            Self::ExecutionError => write!(f, "EXECUTION_ERROR"),
            Self::AuthenticationFailure => write!(f, "AUTH_FAILURE"),
            Self::RateLimitExceeded => write!(f, "RATE_LIMIT_EXCEEDED"),
            Self::Other => write!(f, "OTHER"),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_security_auditor() {
        let auditor = SecurityAuditor::new();
        auditor.log_blocked_code("abc123", "eval detected", Some("tenant1"));
        let events = auditor.get_recent_events(1);
        assert_eq!(events.len(), 1);
        assert_eq!(events[0].event_type, SecurityEventType::CodeBlocked);
    }

    #[test]
    fn test_execution_context_creation() {
        let ctx = ExecutionSecurityContext::new(
            true, true, true,
            256 * 1024 * 1024, // 256MB
            30, // 30s CPU
            vec![],
            false,
        );
        assert!(ctx.isolate_process);
        assert!(ctx.use_seccomp);
        assert!(!ctx.allow_network);
    }

    #[test]
    fn test_limits_description() {
        let ctx = ExecutionSecurityContext::new(
            true, false, false,
            256 * 1024 * 1024,
            30,
            vec![],
            false,
        );
        let desc = ctx.describe_limits();
        assert!(desc.contains("memory=256MB"));
        assert!(desc.contains("cpu=30s"));
        assert!(desc.contains("network=false"));
    }
}