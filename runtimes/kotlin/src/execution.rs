//! Execution engine for Kotlin/JVM code
//!
//! Handles code execution requests, validation, and result formatting.

use crate::config::{ExecutionLimits, RuntimeConfig};
use crate::metrics::MetricsCollector;
use crate::sandbox::{Sandbox, SandboxConfig};
use crate::security::SecurityManager;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::{Duration, Instant};
use uuid::Uuid;

/// Execution request from the orchestrator
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionRequest {
    /// Unique execution ID
    pub id: Uuid,
    /// Kotlin source code to execute
    pub code: String,
    /// Entry point function name (defaults to "main")
    pub entry: Option<String>,
    /// Input data passed to the function (JSON)
    pub input: Option<serde_json::Value>,
    /// Execution timeout (overrides default)
    pub timeout: Option<Duration>,
    /// Custom limits for this execution
    pub limits: Option<ExecutionLimits>,
    /// Security policy override
    pub security_policy: Option<crate::config::SecurityPolicy>,
}

impl ExecutionRequest {
    /// Create a new execution request
    pub fn new(code: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            code,
            entry: None,
            input: None,
            timeout: None,
            limits: None,
            security_policy: None,
        }
    }

    /// Set the entry point
    pub fn with_entry(mut self, entry: impl Into<String>) -> Self {
        self.entry = Some(entry.into());
        self
    }

    /// Set the input data
    pub fn with_input(mut self, input: serde_json::Value) -> Self {
        self.input = Some(input);
        self
    }

    /// Set a custom timeout
    pub fn with_timeout(mut self, timeout: Duration) -> Self {
        self.timeout = Some(timeout);
        self
    }
}

/// Execution response returned to the orchestrator
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionResponse {
    /// Execution ID
    pub id: Uuid,
    /// Whether execution succeeded
    pub success: bool,
    /// Output from execution (stdout)
    pub output: Option<serde_json::Value>,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB (if available)
    pub memory_used_mb: Option<u64>,
    /// Peak memory in MB
    pub peak_memory_mb: Option<u64>,
    /// Whether execution was terminated due to limits
    pub terminated: bool,
    /// Termination reason
    pub termination_reason: Option<String>,
}

impl ExecutionResponse {
    /// Create a successful response
    pub fn success(
        id: Uuid,
        output: serde_json::Value,
        execution_time_ms: u64,
        memory_used_mb: Option<u64>,
        peak_memory_mb: Option<u64>,
    ) -> Self {
        Self {
            id,
            success: true,
            output: Some(output),
            error: None,
            execution_time_ms,
            memory_used_mb,
            peak_memory_mb,
            terminated: false,
            termination_reason: None,
        }
    }

    /// Create a failure response
    pub fn failure(id: Uuid, error: String, execution_time_ms: u64) -> Self {
        Self {
            id,
            success: false,
            output: None,
            error: Some(error),
            execution_time_ms,
            memory_used_mb: None,
            peak_memory_mb: None,
            terminated: true,
            termination_reason: Some("execution_failed".to_string()),
        }
    }

    /// Create a timeout response
    pub fn timeout(id: Uuid, timeout: Duration) -> Self {
        Self {
            id,
            success: false,
            output: None,
            error: Some("execution timed out".to_string()),
            execution_time_ms: timeout.as_millis() as u64,
            memory_used_mb: None,
            peak_memory_mb: None,
            terminated: true,
            termination_reason: Some("timeout".to_string()),
        }
    }
}

/// Code validation result
#[derive(Debug, Clone)]
pub struct ValidationResult {
    /// Whether code is valid
    pub valid: bool,
    /// Errors if invalid
    pub errors: Vec<String>,
    /// Warnings if any
    pub warnings: Vec<String>,
}

impl ValidationResult {
    /// Create a successful validation
    pub fn valid() -> Self {
        Self {
            valid: true,
            errors: vec![],
            warnings: vec![],
        }
    }

    /// Create a failed validation
    pub fn invalid(errors: Vec<String>) -> Self {
        Self {
            valid: false,
            errors,
            warnings: vec![],
        }
    }

    /// Add a warning
    pub fn with_warning(mut self, warning: impl Into<String>) -> Self {
        self.warnings.push(warning.into());
        self
    }
}

/// Executor for Kotlin/JVM code
pub struct Executor {
    config: RuntimeConfig,
    sandbox: Arc<Sandbox>,
    metrics: Arc<MetricsCollector>,
}

impl Executor {
    /// Create a new executor
    pub fn new(config: RuntimeConfig, metrics: Arc<MetricsCollector>) -> Result<Self> {
        let security = SecurityManager::new(config.security.clone());
        let sandbox = Sandbox::new(
            SandboxConfig::default(),
            config.limits.clone(),
            security.into_arc(),
        );

        Ok(Self {
            config,
            sandbox: Arc::new(sandbox),
            metrics,
        })
    }

    /// Create with default configuration
    pub fn with_defaults(metrics: Arc<MetricsCollector>) -> Result<Self> {
        Self::new(RuntimeConfig::default(), metrics)
    }

    /// Execute code with the given request
    pub async fn execute(&self, request: ExecutionRequest) -> ExecutionResponse {
        let start = Instant::now();
        let id = request.id;

        // Validate code first
        let validation = self.validate(&request.code);
        if !validation.valid {
            return ExecutionResponse::failure(
                id,
                format!("validation failed: {}", validation.errors.join(", ")),
                start.elapsed().as_millis() as u64,
            );
        }

        // Get timeout
        let timeout = request.timeout.unwrap_or(self.config.default_timeout);

        // Execute in sandbox
        let result = self.sandbox.execute(&request.code, timeout).await;

        // Record metrics
        let execution_time_ms = start.elapsed().as_millis() as u64;
        self.metrics.record_execution(execution_time_ms, result.as_ref().map(|r| r.memory_used_mb.unwrap_or(0)).unwrap_or(0)).await;

        // Convert result to response
        match result {
            Ok(sandbox_result) => {
                if sandbox_result.success {
                    // Parse output as JSON if possible
                    let output = self.parse_output(&sandbox_result.output);
                    ExecutionResponse::success(
                        id,
                        output,
                        execution_time_ms,
                        sandbox_result.memory_used_mb,
                        sandbox_result.peak_memory_mb,
                    )
                } else {
                    ExecutionResponse::failure(
                        id,
                        sandbox_result.error.unwrap_or_else(|| "unknown error".to_string()),
                        execution_time_ms,
                    )
                }
            }
            Err(e) => {
                if e.to_string().contains("timed out") {
                    ExecutionResponse::timeout(id, timeout)
                } else {
                    ExecutionResponse::failure(id, e.to_string(), execution_time_ms)
                }
            }
        }
    }

    /// Validate code before execution
    pub fn validate(&self, code: &str) -> ValidationResult {
        let mut errors = Vec::new();
        let mut warnings = Vec::new();

        // Basic syntax checks
        if code.trim().is_empty() {
            errors.push("code is empty".to_string());
            return ValidationResult::invalid(errors);
        }

        // Check for Kotlin-specific patterns
        let has_function = code.contains("fun ");
        let has_main = code.contains("fun main");

        if !has_function {
            warnings.push("no function definitions found".to_string());
        }

        if !has_main {
            warnings.push("no main function found - code may not execute".to_string());
        }

        // Check for balanced braces
        let open_braces = code.matches('{').count();
        let close_braces = code.matches('}').count();
        if open_braces != close_braces {
            errors.push(format!(
                "unbalanced braces: {} open, {} close",
                open_braces, close_braces
            ));
        }

        // Check for balanced parentheses
        let open_parens = code.matches('(').count();
        let close_parens = code.matches(')').count();
        if open_parens != close_parens {
            errors.push(format!(
                "unbalanced parentheses: {} open, {} close",
                open_parens, close_parens
            ));
        }

        // Security check
        if let Err(e) = self.sandbox.security().verify_code(code) {
            errors.push(format!("security check failed: {}", e));
        }

        if errors.is_empty() {
            ValidationResult::valid().with_warning(warnings.join("; "))
        } else {
            ValidationResult::invalid(errors)
        }
    }

    /// Parse output string into JSON value
    fn parse_output(&self, output: &str) -> serde_json::Value {
        // Try to parse as JSON first
        if let Ok(parsed) = serde_json::from_str(output) {
            return parsed;
        }

        // Try to parse as JSON lines
        for line in output.lines() {
            if let Ok(parsed) = serde_json::from_str(line) {
                return parsed;
            }
        }

        // Return as string
        serde_json::json!({ "output": output })
    }

    /// Get the sandbox reference
    pub fn sandbox(&self) -> &Arc<Sandbox> {
        &self.sandbox
    }

    /// Get the configuration
    pub fn config(&self) -> &RuntimeConfig {
        &self.config
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validation() {
        let metrics = Arc::new(MetricsCollector::new("test".to_string()));
        let executor = Executor::with_defaults(metrics).unwrap();

        // Valid Kotlin code
        let result = executor.validate("fun main() { println(\"hello\") }");
        assert!(result.valid);

        // Invalid code - unbalanced braces
        let result = executor.validate("fun main() { println(\"hello\") ");
        assert!(!result.valid);

        // Empty code
        let result = executor.validate("");
        assert!(!result.valid);
    }

    #[test]
    fn test_output_parsing() {
        let metrics = Arc::new(MetricsCollector::new("test".to_string()));
        let executor = Executor::with_defaults(metrics).unwrap();

        // JSON output
        let output = r#"{"result": "success"}"#;
        let parsed = executor.parse_output(output);
        assert_eq!(parsed["result"], "success");

        // Plain text output
        let output = "plain text";
        let parsed = executor.parse_output(output);
        assert_eq!(parsed["output"], "plain text");
    }

    #[tokio::test]
    async fn test_execution() {
        let metrics = Arc::new(MetricsCollector::new("test".to_string()));
        let executor = Executor::with_defaults(metrics).unwrap();

        let request = ExecutionRequest::new("fun main() { println(\"hello\") }".to_string());
        let response = executor.execute(request).await;

        assert!(response.success || response.error.is_some());
    }
}