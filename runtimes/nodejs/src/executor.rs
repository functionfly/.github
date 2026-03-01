//! Node.js Runtime Executor
//! 
//! This module provides the core execution engine for JavaScript functions.
//! It uses QuickJS compiled to WebAssembly for secure, isolated execution.

use std::sync::Arc;
use std::time::Instant;
use std::collections::HashMap;

use async_trait::async_trait;
use parking_lot::RwLock;
use tracing::{info, warn, error, instrument};
use uuid::Uuid;

use crate::{
    ExecutionInput, ExecutionResult, ExecutionMetadata, RuntimeError,
    Runtime, RuntimeInfo, RuntimeConfig, RuntimeVersion,
    Sandbox, SandboxConfig,
    TimeoutManager,
    MemoryLimiter,
    host_functions,
};

/// Node.js Runtime Executor
/// 
/// This is the main entry point for executing JavaScript functions.
/// It handles:
/// - Code compilation and caching
/// - Execution isolation via sandboxing
/// - Timeout and memory management
/// - Metrics collection
pub struct NodeExecutor {
    config: RuntimeConfig,
    sandbox: Sandbox,
    timeout_manager: TimeoutManager,
    memory_limiter: MemoryLimiter,
    code_cache: RwLock<HashMap<String, CachedCode>>,
    metrics: crate::metrics::ExecutorMetrics,
}

/// Cached compiled code
struct CachedCode {
    compiled_at: Instant,
    // In a real implementation, this would hold the compiled WASM
    _bytecode: Vec<u8>,
}

impl NodeExecutor {
    /// Create a new NodeExecutor with the given configuration
    pub fn new(config: RuntimeConfig) -> Result<Self, RuntimeError> {
        // Validate configuration
        config.validate()?;
        
        let sandbox = Sandbox::new(SandboxConfig {
            runtime_version: config.version.clone(),
            max_memory_mb: config.max_memory_mb,
            allowed_modules: config.allowed_modules.clone(),
            blocked_modules: config.blocked_modules.clone(),
            network_enabled: config.network_enabled,
            env_vars: config.environment.clone(),
        })?;
        
        let timeout_manager = TimeoutManager::new(config.max_timeout_ms);
        let memory_limiter = MemoryLimiter::new(config.max_memory_mb);
        
        info!(
            "Created NodeExecutor with runtime: {:?}, max_memory: {}MB, max_timeout: {}ms",
            config.version,
            config.max_memory_mb,
            config.max_timeout_ms
        );
        
        Ok(Self {
            config,
            sandbox,
            timeout_manager,
            memory_limiter,
            code_cache: RwLock::new(HashMap::new()),
            metrics: crate::metrics::ExecutorMetrics::new(),
        })
    }

    /// Get a cached code entry or compile if not cached
    fn get_or_compile(&self, code: &str) -> Result<CachedCode, RuntimeError> {
        // Simple hash for cache key (in production, use proper hashing)
        let cache_key = format!("{:x}", md5_hash(code));
        
        // Check cache
        {
            let cache = self.code_cache.read();
            if let Some(cached) = cache.get(&cache_key) {
                // Check if cache is still valid (1 hour)
                if cached.compiled_at.elapsed().as_secs() < 3600 {
                    self.metrics.cache_hits.inc();
                    return Ok(cached.clone());
                }
            }
        }
        
        // Compile code (in a real implementation, this would compile to WASM)
        let compiled = self.compile_code(code)?;
        
        // Cache the compiled code
        let cached = CachedCode {
            compiled_at: Instant::now(),
            _bytecode: compiled,
        };
        
        {
            let mut cache = self.code_cache.write();
            cache.insert(cache_key, cached.clone());
        }
        
        self.metrics.cache_misses.inc();
        Ok(cached)
    }

    /// Compile JavaScript code to the target format
    fn compile_code(&self, code: &str) -> Result<Vec<u8>, RuntimeError> {
        // Validate code for security issues
        self.sandbox.validate_code(code)?;
        
        // In a real implementation, this would:
        // 1. Parse the JavaScript
        // 2. Transform for WASM compatibility
        // 3. Compile to WASM
        // 4. Return the compiled bytecode
        
        // For now, we just return the source as "bytecode"
        // This is a placeholder for the actual compilation
        Ok(code.as_bytes().to_vec())
    }

    /// Execute the compiled code with the given input
    fn execute_internal(
        &self,
        code: &str,
        input: &ExecutionInput,
    ) -> Result<serde_json::Value, RuntimeError> {
        let start = Instant::now();
        
        // Get or compile the code
        let _cached = self.get_or_compile(code)?;
        
        // Execute in sandbox
        let result = self.sandbox.execute(code, &input.data)?;
        
        // Record execution time
        let exec_time = start.elapsed().as_millis() as u64;
        self.metrics.execution_time.observe(exec_time as f64);
        
        info!(
            "Executed function {} in {}ms",
            input.metadata.request_id,
            exec_time
        );
        
        Ok(result)
    }
}

#[async_trait]
impl Runtime for NodeExecutor {
    #[instrument(skip(self, input))]
    async fn execute(
        &self,
        code: &str,
        input: ExecutionInput,
    ) -> ExecutionResult {
        let start = Instant::now();
        let request_id = input.metadata.request_id.clone();
        
        // Record execution start
        self.metrics.total_executions.inc();
        
        // Execute with timeout
        let result = tokio::time::timeout(
            Duration::from_millis(self.config.max_timeout_ms),
            tokio::task::spawn_blocking({
                let code = code.to_string();
                let input = input.clone();
                move || {
                    self.execute_internal(&code, &input)
                }
            })
        ).await;
        
        let exec_time = start.elapsed().as_millis() as u64;
        
        match result {
            Ok(Ok(Ok(output))) => {
                ExecutionResult::success(request_id, output, exec_time)
            }
            Ok(Ok(Err(e))) => {
                self.metrics.errors.inc();
                error!("Execution error: {}", e);
                ExecutionResult::error(request_id, e, exec_time)
            }
            Ok(Err(panic)) => {
                self.metrics.panics.inc();
                error!("Execution panicked: {}", panic);
                ExecutionResult::error(
                    request_id,
                    RuntimeError::Execution(format!("Execution panicked: {}", panic)),
                    exec_time
                )
            }
            Err(_) => {
                self.metrics.timeouts.inc();
                warn!("Execution timed out after {}ms", self.config.max_timeout_ms);
                ExecutionResult::error(
                    request_id,
                    RuntimeError::Timeout(self.config.max_timeout_ms),
                    exec_time
                )
            }
        }
    }

    fn execute_async(
        &self,
        code: &str,
        input: ExecutionInput,
    ) -> std::pin::Pin<Box<dyn Future<Output = ExecutionResult> + Send + '_>> {
        Box::pin(self.execute(code, input))
    }

    fn info(&self) -> RuntimeInfo {
        RuntimeInfo {
            name: "functionfly-nodejs".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
            supported_runtimes: vec![
                RuntimeVersion::Node18,
                RuntimeVersion::Node20,
                RuntimeVersion::Deno,
            ],
            max_memory_mb: self.config.max_memory_mb,
            max_timeout_ms: self.config.max_timeout_ms,
            features: vec![
                "async_await".to_string(),
                "promises".to_string(),
                "fetch_api".to_string(),
                "streams".to_string(),
                "url_api".to_string(),
                "timing_api".to_string(),
                "structured_clone".to_string(),
                "console_api".to_string(),
                "json_api".to_string(),
                "crypto_api".to_string(),
            ],
        }
    }

    async fn health_check(&self) -> bool {
        // Check if sandbox is responsive
        self.sandbox.health_check().await
    }
}

// Simple hash function for cache keys
fn md5_hash(input: &str) -> u64 {
    use std::hash::{Hash, Hasher};
    let mut hasher = std::collections::hash_map::DefaultHasher::new();
    input.hash(&mut hasher);
    hasher.finish()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_executor_creation() {
        let config = RuntimeConfig::default();
        let executor = NodeExecutor::new(config);
        assert!(executor.is_ok());
    }

    #[test]
    fn test_runtime_info() {
        let config = RuntimeConfig::default();
        let executor = NodeExecutor::new(config).unwrap();
        let info = executor.info();
        
        assert!(info.features.contains(&"async_await".to_string()));
    }
}
