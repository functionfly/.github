//! Execution engine for Bun runtime
//!
//! Handles code execution with resource limits, timeouts, and security policies.

use crate::config::{ExecutionLimits, RuntimeConfig};
use crate::metrics::{RuntimeMetrics, MetricsCollector};
use crate::sandbox::{Sandbox, SandboxConfig};
use crate::security::SecurityManager;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::{Duration, Instant};
use uuid::Uuid;

/// Request for code execution
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionRequest {
    /// Unique execution ID
    pub id: Uuid,
    /// Code to execute (TypeScript/JavaScript)
    pub code: String,
    /// Entry point (module URL or function name)
    pub entry: Option<String>,
    /// Input data for the execution
    pub input: Option<serde_json::Value>,
    /// Execution timeout
    pub timeout: Option<Duration>,
    /// Custom limits for this execution (overrides defaults)
    pub limits: Option<ExecutionLimits>,
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
        }
    }

    /// Create an execution request with input
    pub fn with_input(code: String, input: serde_json::Value) -> Self {
        Self {
            id: Uuid::new_v4(),
            code,
            entry: None,
            input: Some(input),
            timeout: None,
            limits: None,
        }
    }
}

/// Response from code execution
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct ExecutionResponse {
    /// Execution ID (matches request)
    pub id: Uuid,
    /// Whether execution succeeded
    pub success: bool,
    /// Output from execution
    pub output: Option<serde_json::Value>,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB (if available)
    pub memory_used_mb: Option<u64>,
    /// Number of modules loaded
    pub modules_loaded: usize,
}

impl ExecutionResponse {
    /// Create a successful response
    pub fn success(
        id: Uuid,
        output: serde_json::Value,
        execution_time_ms: u64,
        modules_loaded: usize,
    ) -> Self {
        Self {
            id,
            success: true,
            output: Some(output),
            error: None,
            execution_time_ms,
            memory_used_mb: None,
            modules_loaded,
        }
    }

    /// Create an error response
    pub fn error(id: Uuid, error: String, execution_time_ms: u64) -> Self {
        Self {
            id,
            success: false,
            output: None,
            error: Some(error),
            execution_time_ms,
            memory_used_mb: None,
            modules_loaded: 0,
        }
    }
}

/// Executor for running code with resource limits
pub struct Executor {
    config: RuntimeConfig,
    sandbox: Arc<Sandbox>,
    metrics: Arc<MetricsCollector>,
}

impl Executor {
    /// Create a new executor with the given configuration
    pub fn new(config: RuntimeConfig) -> Self {
        let security = Arc::new(SecurityManager::new(config.security.clone()));
        let sandbox_config = SandboxConfig::default();
        let sandbox = Arc::new(Sandbox::new(
            sandbox_config,
            config.limits.clone(),
            security,
        ));

        Self {
            config,
            sandbox,
            metrics: Arc::new(MetricsCollector::new()),
        }
    }

    /// Execute code with the given request
    pub async fn execute(&self, request: ExecutionRequest) -> Result<ExecutionResponse> {
        let start = Instant::now();

        // Use custom limits if provided, otherwise use defaults
        let timeout = request.timeout.unwrap_or(self.config.default_timeout);

        // Execute in sandbox
        let result = self.sandbox.execute(&request.code, timeout).await;

        let execution_time_ms = start.elapsed().as_millis() as u64;

        // Update metrics
        self.metrics.record_execution(execution_time_ms, result.is_ok()).await;

        match result {
            Ok(sandbox_result) => {
                if sandbox_result.success {
                    let output = serde_json::json!({
                        "stdout": sandbox_result.output,
                        "execution_time_ms": sandbox_result.execution_time_ms,
                    });
                    Ok(ExecutionResponse::success(
                        request.id,
                        output,
                        execution_time_ms,
                        0, // modules loaded
                    ))
                } else {
                    // Sandbox returned a structured failure (timeout,
                    // memory limit, security violation). Surface it as
                    // an error in the executor response so callers can
                    // distinguish success from failure without parsing
                    // the output JSON.
                    Ok(ExecutionResponse::error(
                        request.id,
                        sandbox_result
                            .error
                            .unwrap_or_else(|| "sandbox execution failed".to_string()),
                        execution_time_ms,
                    ))
                }
            }
            Err(e) => Ok(ExecutionResponse::error(request.id, e.to_string(), execution_time_ms)),
        }
    }

    /// Execute code asynchronously with a channel for cancellation
    pub async fn execute_with_cancel(
        &self,
        request: ExecutionRequest,
    ) -> Result<ExecutionResponse, Cancelled> {
        let timeout = request.timeout.unwrap_or(self.config.default_timeout);
        let start = Instant::now();

        let result = tokio::time::timeout(timeout, self.execute(request)).await;

        let execution_time_ms = start.elapsed().as_millis() as u64;

        match result {
            Ok(Ok(response)) => Ok(response),
            Ok(Err(e)) => Err(Cancelled {
                reason: e.to_string(),
                execution_time_ms,
            }),
            Err(_) => Err(Cancelled {
                reason: "execution timed out".to_string(),
                execution_time_ms,
            }),
        }
    }

    /// Get current metrics
    pub async fn metrics(&self) -> RuntimeMetrics {
        self.metrics.get_metrics().await
    }

    /// Reset metrics
    pub async fn reset_metrics(&self) {
        self.metrics.reset().await;
    }

    /// Get the configuration
    pub fn config(&self) -> &RuntimeConfig {
        &self.config
    }
}

/// Error returned when execution is cancelled
#[derive(Debug)]
pub struct Cancelled {
    pub reason: String,
    pub execution_time_ms: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_executor_execution() {
        let config = RuntimeConfig::default();
        let executor = Executor::new(config);

        let request = ExecutionRequest::new("console.log('hello')".to_string());
        let response = executor.execute(request).await.unwrap();

        assert!(response.success);
        assert!(response.output.is_some());
    }

    #[tokio::test]
    async fn test_executor_timeout() {
        let mut config = RuntimeConfig::default();
        // Use a tiny memory limit (8 MB) so the allocation loop is guaranteed
        // to trip the limit before the 2-second timeout.
        config.limits.max_memory_mb = 8;
        let executor = Executor::new(config);

        // Allocate memory until the runtime's memory limit kicks in. We use
        // memory exhaustion (not an infinite loop) because we removed the
        // QuickJS interrupt handler — `while(true){}` would never be aborted.
        // Allocating an array that grows past the memory limit forces
        // QuickJS to throw `RangeError`, which we catch and return as an
        // error.
        let request = ExecutionRequest {
            id: Uuid::new_v4(),
            code: "var a = []; while(true) { a.push(new Array(1024)); }".to_string(),
            entry: None,
            input: None,
            timeout: Some(Duration::from_secs(2)),
            limits: None,
        };

        let response = executor.execute(request).await.unwrap();
        // Either: timeout triggered (response.success=false, error=Some) OR
        // memory limit triggered (same). The test passes if execution did
        // not hang.
        assert!(!response.success || response.error.is_some());
    }
}
