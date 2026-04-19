//! Node.js Runtime Executor
//!
//! This module provides the core execution engine for JavaScript functions.
//! It uses QuickJS compiled to WebAssembly for secure, isolated execution.

use std::sync::Arc;
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::{Instant, Duration};
use std::collections::HashMap;
use std::future::Future;

use async_trait::async_trait;
use sha2::{Sha256, Digest};
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

/// JavaScript execution context
/// 
/// Wraps a QuickJS runtime instance for executing JavaScript code.
pub struct JsContext {
    network_enabled: bool,
    environment: HashMap<String, String>,
    id: String,
}

impl JsContext {
    /// Create a new JavaScript context
    pub fn new(
        network_enabled: bool,
        environment: &HashMap<String, String>,
    ) -> Result<Self, RuntimeError> {
        Ok(Self {
            network_enabled,
            environment: environment.clone(),
            id: format!("ctx_{}", uuid::Uuid::new_v4()),
        })
    }

    /// Load a JavaScript module into the context
    pub fn load_module(&mut self, _code: &str) -> Result<(), RuntimeError> {
        // In a real implementation, this would compile and load the JS code
        Ok(())
    }

    /// Call the handler function with the given input
    pub fn call_handler(&mut self, _input_json: &str) -> Result<String, RuntimeError> {
        // In a real implementation, this would execute the handler
        Ok(_input_json.to_string())
    }

    /// Get the context ID
    pub fn id(&self) -> &str {
        &self.id
    }
}

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
    sandbox: Arc<Sandbox>,
    timeout_manager: TimeoutManager,
    memory_limiter: MemoryLimiter,
    code_cache: RwLock<HashMap<String, CachedCode>>,
    metrics: crate::metrics::ExecutorMetrics,
    context_pool: RwLock<Vec<JsContext>>,
}

/// Cached compiled code
#[derive(Clone)]
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
            max_concurrent_executions: 100, // TODO: add to RuntimeConfig
            allowed_modules: config.allowed_modules.clone(),
            blocked_modules: config.blocked_modules.clone(),
            network_enabled: config.network_enabled,
            env_vars: config.environment.clone(),
            max_code_size_bytes: 10 * 1024 * 1024, // 10MB default
            strict_mode: true,
        })?;

        let timeout_manager = TimeoutManager::new(config.max_timeout_ms);
        let memory_limiter = MemoryLimiter::new(config.max_memory_mb);

        // Initialize context pool
        let context_pool = RwLock::new(Vec::new());

        // Wrap sandbox in Arc for sharing
        let sandbox = Arc::new(sandbox);

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
            context_pool,
        })
    }

    /// Acquire a JS context from the pool (or create new one if needed)
    pub fn acquire_context(&self) -> Result<JsContext, RuntimeError> {
        let mut pool = self.context_pool.write();
        if let Some(ctx) = pool.pop() {
            Ok(ctx)
        } else {
            // Create new context
            JsContext::new(self.config.network_enabled, &self.config.environment)
        }
    }

    /// Release a JS context back to the pool
    pub fn release_context(&self, ctx: JsContext) {
        let mut pool = self.context_pool.write();
        // Limit pool size to avoid unbounded growth
        if pool.len() < 10 {
            pool.push(ctx);
        }
    }

    /// Get a cached code entry or compile if not cached
    fn get_or_compile(&self, code: &str) -> Result<CachedCode, RuntimeError> {
        let cache_key = code_cache_key(code);

        // Check cache
        {
            let cache = self.code_cache.read();
            if let Some(cached) = cache.get(&cache_key) {
                // Check if cache is still valid (1 hour)
                if cached.compiled_at.elapsed().as_secs() < 3600 {
                    self.metrics.cache_hits.fetch_add(1, Ordering::Relaxed);
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

        self.metrics.cache_misses.fetch_add(1, Ordering::Relaxed);
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

        // Record execution time (convert ms to ns)
        let exec_time = start.elapsed().as_nanos() as u64;
        self.metrics.execution_time(exec_time);

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
        let timeout_ms = self.config.max_timeout_ms;

        // Record execution start
        self.metrics.total_executions.fetch_add(1, Ordering::Relaxed);

        // Clone data needed for execution
        let code_owned = code.to_string();
        let input_data = input.data.clone();
        let metadata = input.metadata.clone();
        let input = ExecutionInput { data: input_data, metadata };

        // Clone sandbox reference for execution
        let sandbox = self.sandbox.clone();

        // Execute with timeout
        let result = tokio::time::timeout(
            Duration::from_millis(timeout_ms),
            tokio::task::spawn_blocking(move || {
                // Execute in sandbox
                sandbox.execute(&code_owned, &input.data)
            })
        ).await;

        let exec_time = start.elapsed().as_millis() as u64;

        match result {
            Ok(Ok(Ok(output))) => {
                ExecutionResult::success(request_id, output, exec_time)
            }
            Ok(Ok(Err(e))) => {
                self.metrics.errors.fetch_add(1, Ordering::Relaxed);
                error!("Execution error: {}", e);
                ExecutionResult::error(request_id, e, exec_time)
            }
            Ok(Err(panic)) => {
                self.metrics.panics.fetch_add(1, Ordering::Relaxed);
                let panic_msg = panic.to_string();
                error!("Execution panicked: {}", panic_msg);
                ExecutionResult::error(
                    request_id,
                    RuntimeError::Execution(format!("Execution panicked: {}", panic_msg)),
                    exec_time
                )
            }
            Err(_) => {
                self.metrics.timeouts.fetch_add(1, Ordering::Relaxed);
                warn!("Execution timed out after {}ms", timeout_ms);
                ExecutionResult::error(
                    request_id,
                    RuntimeError::Timeout(timeout_ms),
                    exec_time
                )
            }
        }
    }

    fn execute_async<'a>(
        &'a self,
        code: &'a str,
        input: ExecutionInput,
    ) -> std::pin::Pin<Box<dyn Future<Output = ExecutionResult> + Send + 'a>> {
        Box::pin(async move {
            self.execute(code, input).await
        })
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

/// Content-addressable cache key for compiled code (SHA-256).
/// Same source code always yields the same key; different code yields different keys with
/// negligible collision probability.
fn code_cache_key(code: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(code.as_bytes());
    let digest = hasher.finalize();
    hex::encode(digest)
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
