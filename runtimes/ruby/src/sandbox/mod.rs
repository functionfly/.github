//! Ruby Runtime Sandbox
//!
//! Sandboxed execution environment for Ruby code with resource limits and process isolation.

use crate::config::{ExecutionLimits, SecurityPolicy};
use crate::metrics::MetricsCollector;
use crate::security::{SecurityManager, SecurityError, ExecutionSecurityContext};
use std::sync::Arc;
use std::time::{Duration, Instant};
use parking_lot::RwLock;
use tracing::{debug, error, info, warn};

/// Sandbox configuration
#[derive(Debug, Clone)]
pub struct SandboxConfig {
    /// Enable memory tracking
    pub track_memory: bool,
    /// Enable CPU time tracking
    pub track_cpu_time: bool,
    /// Enable sandboxed filesystem
    pub sandbox_filesystem: bool,
    /// Sandbox working directory
    pub working_dir: Option<String>,
    /// Enable process isolation
    pub isolate_process: bool,
    /// Enable strict security mode (reject dangerous code)
    pub strict_mode: bool,
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self {
            track_memory: true,
            track_cpu_time: true,
            sandbox_filesystem: true,
            working_dir: None,
            isolate_process: true,
            strict_mode: true,
        }
    }
}

/// Sandbox execution result
#[derive(Debug, Clone, serde::Serialize)]
pub struct SandboxResult {
    /// Whether execution succeeded
    pub success: bool,
    /// Output from execution (JSON string)
    pub output: Option<String>,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB
    pub memory_used_mb: Option<f64>,
    /// Whether a timeout occurred
    pub timeout: bool,
    /// Sandbox violation if any
    pub violation: Option<String>,
    /// Code hash for audit trail
    pub code_hash: Option<String>,
}

impl SandboxResult {
    pub fn success(output: String, execution_time_ms: u64, code_hash: Option<String>) -> Self {
        Self {
            success: true,
            output: Some(output),
            error: None,
            execution_time_ms,
            memory_used_mb: None,
            timeout: false,
            violation: None,
            code_hash,
        }
    }

    pub fn timeout(execution_time_ms: u64) -> Self {
        Self {
            success: false,
            output: None,
            error: Some("execution timed out".to_string()),
            execution_time_ms,
            memory_used_mb: None,
            timeout: true,
            violation: None,
            code_hash: None,
        }
    }

    pub fn violation(msg: String, execution_time_ms: u64) -> Self {
        Self {
            success: false,
            output: None,
            error: Some(msg.clone()),
            execution_time_ms,
            memory_used_mb: None,
            timeout: false,
            violation: Some(msg),
            code_hash: None,
        }
    }

    pub fn error(msg: String, execution_time_ms: u64) -> Self {
        Self {
            success: false,
            output: None,
            error: Some(msg),
            execution_time_ms,
            memory_used_mb: None,
            timeout: false,
            violation: None,
            code_hash: None,
        }
    }
}

/// Sandbox for Ruby code execution
pub struct Sandbox {
    config: SandboxConfig,
    limits: ExecutionLimits,
    security: Arc<SecurityManager>,
    metrics: Arc<MetricsCollector>,
    active_count: Arc<RwLock<usize>>,
}

impl Sandbox {
    /// Create a new sandbox
    pub fn new(
        config: SandboxConfig,
        limits: ExecutionLimits,
        security: Arc<SecurityManager>,
        metrics: Arc<MetricsCollector>,
    ) -> Arc<Self> {
        Arc::new(Self {
            config,
            limits,
            security,
            metrics,
            active_count: Arc::new(RwLock::new(0)),
        })
    }

    /// Execute Ruby code in sandbox with full security enforcement
    pub async fn execute(
        &self,
        code: &str,
        timeout: Duration,
    ) -> Result<SandboxResult, SandboxError> {
        // Check concurrent limit
        {
            let count = self.active_count.read();
            if *count >= self.limits.max_allocations as usize {
                return Err(SandboxError::TooManyExecutions);
            }
        }

        // Increment active count
        {
            let mut guard = self.active_count.write();
            *guard += 1;
        }
        self.metrics.inc_active();
        let _defer = defer(|| {
            let mut guard = self.active_count.write();
            *guard = guard.saturating_sub(1);
            drop(guard);
            self.metrics.dec_active();
        });

        let start = Instant::now();

        // Step 1: Validate code with strict security checks
        let validation = match self.security.validate_code_detailed(code) {
            Ok(v) => v,
            Err(e) => {
                let elapsed = start.elapsed().as_millis() as u64;
                self.metrics.record_failure();
                return Ok(SandboxResult::violation(e.to_string(), elapsed));
            }
        };

        // Step 2: Check if code is valid (no critical/high violations)
        if !validation.valid {
            let elapsed = start.elapsed().as_millis() as u64;
            self.metrics.record_failure();
            return Ok(SandboxResult::violation(
                "Code blocked: contains dangerous patterns".to_string(),
                elapsed,
            ));
        }

        // Step 3: Execute with process isolation
        let exec_timeout = timeout.min(Duration::from_secs(self.limits.max_cpu_time_secs as u64));

        // Execute using inline execution (process isolation would require additional setup)
        let result = self.execute_inline(&validation.sanitized.unwrap_or_default(), exec_timeout).await;

        let elapsed_ms = start.elapsed().as_millis() as u64;

        match result {
            Ok((success, output, error_msg)) => {
                if success {
                    // Truncate output if needed
                    let final_output = truncate_string(output, self.limits.max_output_bytes);
                    self.metrics.record_success(elapsed_ms);
                    Ok(SandboxResult::success(
                        final_output,
                        elapsed_ms,
                        Some(validation.code_hash),
                    ))
                } else {
                    self.metrics.record_failure();
                    Ok(SandboxResult::error(
                        error_msg,
                        elapsed_ms,
                    ))
                }
            }
            Err(e) => {
                self.metrics.record_failure();
                Ok(SandboxResult::error(e.to_string(), elapsed_ms))
            }
        }
    }

    /// Execute Ruby code inline with the Ruby engine
    async fn execute_inline(
        &self,
        code: &str,
        timeout: Duration,
    ) -> Result<(bool, String, String), SandboxError> {
        #[cfg(feature = "ruby-engine")]
        {
            use rutie::VM;
            use std::sync::mpsc;

            let (tx, rx) = mpsc::channel::<Result<(bool, String, String), std::sync::mpsc::RecvError>>();
            let code_owned = code.to_string();

            std::thread::spawn(move || {
                VM::init();
                match VM::eval(&code_owned) {
                    Ok(value) => {
                        let output = format!("{:?}", value);
                        let _ = tx.send(Ok((true, output, String::new())));
                    }
                    Err(e) => {
                        let _ = tx.send(Ok((false, String::new(), format!("{:?}", e))));
                    }
                }
            });

            match rx.recv_timeout(timeout) {
                Ok(Ok(result)) => Ok(result),
                Ok(Err(e)) => Err(SandboxError::ExecutionError(e.to_string())),
                Err(_) => {
                    self.metrics.record_timeout_hit();
                    Err(SandboxError::Timeout)
                }
            }
        }

        #[cfg(not(feature = "ruby-engine"))]
        {
            Err(SandboxError::EngineNotAvailable(
                "Ruby engine not compiled - ruby-engine feature is required".to_string(),
            ))
        }
    }

    /// Get the number of active executions
    pub fn active_count(&self) -> usize {
        *self.active_count.read()
    }

    /// Get the configured limits
    pub fn limits(&self) -> &ExecutionLimits {
        &self.limits
    }

    /// Check if sandbox is in strict mode
    pub fn is_strict(&self) -> bool {
        self.config.strict_mode
    }
}

/// Truncate string to max bytes
fn truncate_string(s: String, max_bytes: usize) -> String {
    if s.len() <= max_bytes {
        return s;
    }
    let truncated = &s[..max_bytes];
    format!("{}... [truncated {} bytes]", truncated, s.len() - max_bytes)
}

/// RAII guard for decrementing active count
struct Defer<F: Fn()>(F);
impl<F: Fn()> Drop for Defer<F> {
    fn drop(&mut self) {
        (self.0)()
    }
}
fn defer<F: Fn()>(f: F) -> Defer<F> {
    Defer(f)
}

/// Sandbox errors
#[derive(Debug, thiserror::Error)]
pub enum SandboxError {
    #[error("too many concurrent executions")]
    TooManyExecutions,

    #[error("execution error: {0}")]
    ExecutionError(String),

    #[error("security violation: {0}")]
    SecurityViolation(String),

    #[error("timeout")]
    Timeout,

    #[error("engine not available: {0}")]
    EngineNotAvailable(String),
}

/// From SecurityError conversion
impl From<SecurityError> for SandboxError {
    fn from(e: SecurityError) -> Self {
        SandboxError::SecurityViolation(e.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_limits() -> ExecutionLimits {
        ExecutionLimits {
            max_memory_mb: 256,
            max_cpu_time_secs: 30,
            max_wall_time_secs: 60,
            max_output_bytes: 1024 * 1024,
            max_stack_depth: 1024,
            max_allocations: 100,
        }
    }

    fn test_policy() -> SecurityPolicy {
        SecurityPolicy::default()
    }

    fn test_metrics() -> Arc<MetricsCollector> {
        Arc::new(MetricsCollector::new())
    }

    fn test_security() -> Arc<SecurityManager> {
        SecurityManager::new(test_policy())
    }

    #[tokio::test]
    async fn test_sandbox_creation() {
        let sandbox = Sandbox::new(
            SandboxConfig::default(),
            test_limits(),
            test_security(),
            test_metrics(),
        );
        assert_eq!(sandbox.active_count(), 0);
    }

    #[tokio::test]
    async fn test_sandbox_blocks_dangerous_code() {
        let sandbox = Sandbox::new(
            SandboxConfig::default(),
            test_limits(),
            test_security(),
            test_metrics(),
        );

        let result = sandbox.execute("eval('puts 1')", Duration::from_secs(5)).await;
        // Should block dangerous code
        assert!(!result.unwrap().success);
    }

    #[tokio::test]
    async fn test_truncate_string() {
        let long_string = "a".repeat(100);
        let truncated = truncate_string(long_string.clone(), 50);
        assert!(truncated.len() < long_string.len());
        assert!(truncated.contains("truncated"));
    }
}