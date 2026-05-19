//! Ruby Runtime Execution Engine
//!
//! Core execution engine for Ruby code with async support and tenant isolation.

use crate::config::{ExecutionLimits, RuntimeConfig};
use crate::metrics::MetricsCollector;
use crate::sandbox::{Sandbox, SandboxConfig, SandboxResult};
use crate::security::SecurityManager;
use anyhow::Result;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::{Duration, Instant};
use parking_lot::RwLock;
use tracing::{debug, info, warn};
use uuid::Uuid;

/// Tenant context for execution isolation
#[derive(Debug, Clone)]
pub struct TenantContext {
    pub tenant_id: String,
    pub rate_limit_rps: Option<u64>,
    pub max_concurrent: Option<usize>,
    pub allowed_dirs: Option<Vec<String>>,
    pub memory_limit_mb: Option<u64>,
    pub timeout_ms: Option<u64>,
}

impl TenantContext {
    pub fn new(tenant_id: String) -> Self {
        Self {
            tenant_id,
            rate_limit_rps: None,
            max_concurrent: None,
            allowed_dirs: None,
            memory_limit_mb: None,
            timeout_ms: None,
        }
    }

    pub fn with_rate_limit(mut self, rps: u64) -> Self {
        self.rate_limit_rps = Some(rps);
        self
    }

    pub fn with_max_concurrent(mut self, max: usize) -> Self {
        self.max_concurrent = Some(max);
        self
    }

    pub fn with_allowed_dirs(mut self, dirs: Vec<String>) -> Self {
        self.allowed_dirs = Some(dirs);
        self
    }

    pub fn with_memory_limit(mut self, mb: u64) -> Self {
        self.memory_limit_mb = Some(mb);
        self
    }

    pub fn with_timeout(mut self, ms: u64) -> Self {
        self.timeout_ms = Some(ms);
        self
    }
}

/// Execution request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionRequest {
    /// Unique execution ID
    pub execution_id: String,
    /// Ruby code to execute
    pub code: String,
    /// Input data (JSON)
    pub input: Option<serde_json::Value>,
    /// Timeout in milliseconds
    pub timeout_ms: Option<u64>,
    /// Tenant ID for isolation
    pub tenant_id: Option<String>,
}

impl ExecutionRequest {
    /// Get timeout with default
    pub fn timeout(&self) -> Duration {
        Duration::from_millis(self.timeout_ms.unwrap_or(30000))
    }

    /// Get tenant context
    pub fn tenant_context(&self) -> Option<TenantContext> {
        self.tenant_id.as_ref().map(|id| TenantContext::new(id.clone()))
    }

    /// Get code hash for audit
    pub fn code_hash(&self) -> String {
        use sha2::{Sha256, Digest};
        let mut hasher = Sha256::new();
        hasher.update(self.code.as_bytes());
        hex::encode(hasher.finalize())
    }
}

/// Execution response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionResponse {
    /// Execution ID
    pub execution_id: String,
    /// Success flag
    pub success: bool,
    /// Output data (JSON string)
    pub output: Option<String>,
    /// Error message
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB
    pub memory_used_mb: Option<f64>,
    /// Cache hit flag
    pub cache_hit: bool,
    /// Code hash for audit
    pub code_hash: Option<String>,
}

impl ExecutionResponse {
    pub fn success(
        execution_id: String,
        output: String,
        execution_time_ms: u64,
        code_hash: Option<String>,
    ) -> Self {
        Self {
            execution_id,
            success: true,
            output: Some(output),
            error: None,
            execution_time_ms,
            memory_used_mb: None,
            cache_hit: false,
            code_hash,
        }
    }

    pub fn error(execution_id: String, error: String, execution_time_ms: u64) -> Self {
        Self {
            execution_id,
            success: false,
            output: None,
            error: Some(error),
            execution_time_ms,
            memory_used_mb: None,
            cache_hit: false,
            code_hash: None,
        }
    }
}

/// Executor trait for Ruby code execution
#[async_trait]
pub trait Executor: Send + Sync {
    /// Execute Ruby code
    async fn execute(&self, request: ExecutionRequest) -> ExecutionResponse;
}

/// Default executor implementation with tenant isolation
pub struct DefaultExecutor {
    sandbox: Arc<Sandbox>,
    limits: ExecutionLimits,
    metrics: Arc<MetricsCollector>,
    active_count: Arc<RwLock<usize>>,
    max_concurrent: usize,
}

impl DefaultExecutor {
    /// Create a new executor
    pub fn new(
        config: RuntimeConfig,
        security: Arc<SecurityManager>,
        metrics: Arc<MetricsCollector>,
    ) -> Arc<Self> {
        let sandbox_config = SandboxConfig {
            isolate_process: config.security.sandbox_enabled,
            strict_mode: config.security.sanitize_code,
            ..Default::default()
        };

        let sandbox = Sandbox::new(
            sandbox_config,
            config.limits.clone(),
            security,
            metrics.clone(),
        );

        Arc::new(Self {
            sandbox,
            limits: config.limits,
            metrics,
            active_count: Arc::new(RwLock::new(0)),
            max_concurrent: config.max_concurrent,
        })
    }

    /// Execute a request with tenant context
    async fn do_execute(&self, request: ExecutionRequest, timeout: Duration) -> ExecutionResponse {
        let exec_id = request.execution_id.clone();
        let code_hash = request.code_hash();
        let _tenant_id = request.tenant_id.clone();
        let start = Instant::now();

        // Check concurrent limit
        {
            let count = self.active_count.read();
            if *count >= self.max_concurrent {
                let elapsed = start.elapsed().as_millis() as u64;
                return ExecutionResponse::error(
                    exec_id,
                    "too many concurrent executions".to_string(),
                    elapsed,
                );
            }
        }

        // Increment active count
        {
            let mut guard = self.active_count.write();
            *guard += 1;
        }
        self.metrics.inc_active();
        self.metrics.inc_executions();

        // Execute with sandbox
        let result = self.sandbox.execute(&request.code, timeout).await;

        // Decrement active count
        {
            let mut guard = self.active_count.write();
            *guard = guard.saturating_sub(1);
        }
        self.metrics.dec_active();

        let elapsed_ms = start.elapsed().as_millis() as u64;

        match result {
            Ok(sandbox_result) => {
                if sandbox_result.success {
                    ExecutionResponse {
                        execution_id: exec_id,
                        success: true,
                        output: sandbox_result.output,
                        error: None,
                        execution_time_ms: elapsed_ms,
                        memory_used_mb: sandbox_result.memory_used_mb,
                        cache_hit: false,
                        code_hash: Some(code_hash),
                    }
                } else {
                    ExecutionResponse::error(
                        exec_id,
                        sandbox_result.error.unwrap_or_else(|| "unknown error".to_string()),
                        elapsed_ms,
                    )
                }
            }
            Err(e) => {
                self.metrics.record_failure();
                ExecutionResponse::error(exec_id, e.to_string(), elapsed_ms)
            }
        }
    }
}

#[async_trait]
impl Executor for DefaultExecutor {
    async fn execute(&self, request: ExecutionRequest) -> ExecutionResponse {
        let timeout = request.timeout();
        self.do_execute(request, timeout).await
    }
}

/// Execution statistics
#[derive(Debug, Clone, Default)]
pub struct ExecutionStats {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub failed_executions: u64,
    pub active_executions: usize,
    pub total_execution_time_ms: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_execution_request_timeout() {
        let request = ExecutionRequest {
            execution_id: "test-1".to_string(),
            code: "puts 'hello'".to_string(),
            input: None,
            timeout_ms: None,
            tenant_id: None,
        };
        assert_eq!(request.timeout(), Duration::from_millis(30000));
    }

    #[test]
    fn test_execution_request_custom_timeout() {
        let request = ExecutionRequest {
            execution_id: "test-1".to_string(),
            code: "puts 'hello'".to_string(),
            input: None,
            timeout_ms: Some(5000),
            tenant_id: None,
        };
        assert_eq!(request.timeout(), Duration::from_millis(5000));
    }

    #[test]
    fn test_tenant_context() {
        let ctx = TenantContext::new("tenant-1".to_string())
            .with_rate_limit(100)
            .with_max_concurrent(50)
            .with_memory_limit(512);

        assert_eq!(ctx.tenant_id, "tenant-1");
        assert_eq!(ctx.rate_limit_rps, Some(100));
        assert_eq!(ctx.max_concurrent, Some(50));
        assert_eq!(ctx.memory_limit_mb, Some(512));
    }

    #[test]
    fn test_code_hash() {
        let request = ExecutionRequest {
            execution_id: "test-1".to_string(),
            code: "puts 'hello'".to_string(),
            input: None,
            timeout_ms: None,
            tenant_id: None,
        };

        let hash1 = request.code_hash();
        let hash2 = request.code_hash();
        assert_eq!(hash1, hash2);

        let request2 = ExecutionRequest {
            execution_id: "test-2".to_string(),
            code: "puts 'world'".to_string(),
            input: None,
            timeout_ms: None,
            tenant_id: None,
        };
        let hash3 = request2.code_hash();
        assert_ne!(hash1, hash3);
    }
}