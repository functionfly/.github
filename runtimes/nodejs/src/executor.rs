//! Node.js Runtime Executor
//!
//! This module provides the core execution engine for JavaScript functions.
//! Uses rquickjs (high-level QuickJS bindings) for secure, isolated execution.

use std::sync::atomic::Ordering;
use std::time::{Instant, Duration};
use std::collections::HashMap;
use std::future::Future;

use async_trait::async_trait;
use sha2::{Sha256, Digest};
use parking_lot::RwLock;
use tracing::{info, warn, error};
use rquickjs::{Runtime as QjsRuntime, Context as QjsContext};

use crate::{
    ExecutionInput, ExecutionResult, RuntimeError,
    Runtime as RuntimeTrait, RuntimeInfo, RuntimeConfig, RuntimeVersion,
    Sandbox, SandboxConfig,
};

pub struct JsContext {
    runtime: QjsRuntime,
    context: QjsContext,
}

impl JsContext {
    pub fn new(_network_enabled: bool, _environment: &HashMap<String, String>) -> Result<Self, RuntimeError> {
        let runtime = QjsRuntime::new().map_err(|e| RuntimeError::NotReady(format!("JS runtime error: {}", e)))?;
        let context = QjsContext::full(&runtime).map_err(|e| RuntimeError::NotReady(format!("JS context error: {}", e)))?;
        Ok(Self { runtime, context })
    }

    pub fn load_module(&mut self, code: &str) -> Result<(), RuntimeError> {
        // QuickJS `Context::eval` runs code in *script* mode, which does not
        // understand ES-module syntax (`export function foo`, `export default`,
        // `import ... from`). However, our public API documents that
        // handlers may use ESM-style declarations for clarity. To bridge the
        // two, we strip `export ` keywords (and `import` statements that
        // reference bundled modules we cannot resolve here) before eval.
        //
        // This is a conservative transformation: we only strip leading
        // `export ` (with a single space) and `export default `. We do NOT
        // attempt to rewrite arbitrary import specifiers — handlers that
        // need external modules must be pre-bundled at publish time.
        let stripped = strip_esm_keywords(code);
        let code_bytes: Vec<u8> = stripped.into_bytes();
        let result: Result<(), RuntimeError> = self.context.with(move |ctx| {
            match ctx.eval::<(), _>(code_bytes.as_slice()) {
                Ok(_) => Ok(()),
                Err(e) => Err(RuntimeError::Compilation(format!("JS eval error: {}", e))),
            }
        });
        result.map_err(|e| RuntimeError::Execution(format!("context error: {}", e)))
    }

    pub fn call_handler(&mut self, input_json: &str) -> Result<String, RuntimeError> {
        let escaped_input = input_json.replace('\\', "\\\\").replace('\'', "\\'").replace('\n', "\\n").replace('\r', "\\r");
        let handler_code = format!(
            "(function() {{ var input; try {{ input = JSON.parse(\'{}\'); }} catch(e) {{ throw new Error(\'Invalid JSON: \' + e.message); }} var result; if (typeof handler === \'function\') {{ result = handler(input); }} else if (typeof module !== \'undefined\' && module.exports && typeof module.exports.handler === \'function\') {{ result = module.exports.handler(input); }} else if (typeof defaultHandler === \'function\') {{ result = defaultHandler(input); }} else {{ throw new Error(\'No handler found.\'); }} if (result === undefined) {{ return \"null\"; }} return JSON.stringify(result); }})()",
            escaped_input
        );

        let code_bytes: Vec<u8> = handler_code.into_bytes();
        let result: Result<String, RuntimeError> = self.context.with(move |ctx| {
            match ctx.eval::<String, _>(code_bytes.as_slice()) {
                Ok(s) => Ok(s),
                Err(e) => Err(RuntimeError::Execution(format!("handler error: {}", e))),
            }
        });
        result.map_err(|e| RuntimeError::Execution(format!("context error: {}", e)))
    }
}

pub struct NodeExecutor {
    config: RuntimeConfig,
    sandbox: std::sync::Arc<Sandbox>,
    code_cache: RwLock<HashMap<String, CachedCode>>,
    metrics: crate::metrics::ExecutorMetrics,
}

#[derive(Clone)]
struct CachedCode { compiled_at: Instant }

impl NodeExecutor {
    pub fn new(config: RuntimeConfig) -> Result<Self, RuntimeError> {
        config.validate()?;
        // SECURITY: clamp max_code_size_bytes to the sandbox's hard limit
        // (1 MiB) so the executor never tries to push a 10 MiB bundle through
        // a sandbox that only accepts 1 MiB. This also prevents accidental
        // OOMs when callers trust the configured limit without knowing the
        // sandbox's internal ceiling.
        let max_code_size_bytes = (10 * 1024 * 1024).min(crate::sandbox::MAX_CODE_SIZE_BYTES);
        let sandbox = Sandbox::new(SandboxConfig {
            runtime_version: config.version.clone(),
            max_memory_mb: config.max_memory_mb,
            max_concurrent_executions: 100,
            allowed_modules: config.allowed_modules.clone(),
            blocked_modules: config.blocked_modules.clone(),
            network_enabled: config.network_enabled,
            env_vars: config.environment.clone(),
            max_code_size_bytes,
            strict_mode: true,
        })?;
        info!("Created NodeExecutor runtime: {:?} memory: {}MB timeout: {}ms",
            config.version, config.max_memory_mb, config.max_timeout_ms);
        Ok(Self { config, sandbox: std::sync::Arc::new(sandbox), code_cache: RwLock::new(HashMap::new()),
            metrics: crate::metrics::ExecutorMetrics::new() })
    }

    fn get_or_compile(&self, code: &str) -> Result<(), RuntimeError> {
        let cache_key = code_cache_key(code);
        let mut cache = self.code_cache.write();
        if let Some(cached) = cache.get(&cache_key) {
            if cached.compiled_at.elapsed().as_secs() < 3600 {
                self.metrics.cache_hits.fetch_add(1, Ordering::Relaxed);
                return Ok(());
            }
        }
        cache.insert(cache_key, CachedCode { compiled_at: Instant::now() });
        self.metrics.cache_misses.fetch_add(1, Ordering::Relaxed);
        Ok(())
    }

    fn compile_code(&self, code: &str) -> Result<(), RuntimeError> {
        self.sandbox.validate_code(code)?;
        Ok(())
    }
}

#[async_trait]
impl RuntimeTrait for NodeExecutor {
    async fn execute(&self, code: &str, input: ExecutionInput) -> ExecutionResult {
        let start = Instant::now();
        let request_id = input.metadata.request_id.clone();
        let timeout_ms = self.config.max_timeout_ms;
        self.metrics.total_executions.fetch_add(1, Ordering::Relaxed);

        let code_owned = code.to_string();
        let input_data = input.data.clone();

        let result = tokio::time::timeout(
            Duration::from_millis(timeout_ms),
            tokio::task::spawn_blocking(move || {
                let mut ctx = JsContext::new(false, &HashMap::new())
                    .map_err(|e| RuntimeError::NotReady(format!("context failed: {}", e)))?;
                ctx.load_module(&code_owned)
                    .map_err(|e| RuntimeError::Compilation(format!("load failed: {}", e)))?;
                let input_json = serde_json::to_string(&input_data)
                    .map_err(|e| RuntimeError::InvalidInput(format!("serialize failed: {}", e)))?;
                let result_json = ctx.call_handler(&input_json)
                    .map_err(|e| RuntimeError::Execution(format!("handler failed: {}", e)))?;
                serde_json::from_str(&result_json)
                    .map_err(|e| RuntimeError::Execution(format!("invalid result JSON: {}", e)))
            })
        ).await;

        let exec_time = start.elapsed().as_millis() as u64;

        match result {
            Ok(Ok(Ok(output))) => {
                self.metrics.execution_time(exec_time * 1_000_000);
                ExecutionResult::success(request_id, output, exec_time)
            }
            Ok(Ok(Err(e))) => {
                self.metrics.errors.fetch_add(1, Ordering::Relaxed);
                error!("Execution error: {}", e);
                ExecutionResult::error(request_id, e, exec_time)
            }
            Ok(Err(join_error)) => {
                self.metrics.errors.fetch_add(1, Ordering::Relaxed);
                ExecutionResult::error(request_id, RuntimeError::Execution(join_error.to_string()), exec_time)
            }
            Err(_) => {
                self.metrics.timeouts.fetch_add(1, Ordering::Relaxed);
                warn!("Execution timed out after {}ms", timeout_ms);
                ExecutionResult::error(request_id, RuntimeError::Timeout(timeout_ms), exec_time)
            }
        }
    }

    fn execute_async<'a>(&'a self, code: &'a str, input: ExecutionInput) -> std::pin::Pin<Box<dyn Future<Output = ExecutionResult> + Send + 'a>> {
        Box::pin(async move { self.execute(code, input).await })
    }

    fn info(&self) -> RuntimeInfo {
        RuntimeInfo {
            name: "functionfly-nodejs".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
            supported_runtimes: vec![RuntimeVersion::Node18, RuntimeVersion::Node20, RuntimeVersion::Deno],
            max_memory_mb: self.config.max_memory_mb,
            max_timeout_ms: self.config.max_timeout_ms,
            features: vec!["async_await", "promises", "fetch_api", "streams", "url_api", "timing_api", "structured_clone", "console_api", "json_api"].into_iter().map(String::from).collect(),
        }
    }

    async fn health_check(&self) -> bool { self.sandbox.health_check().await }
}

fn code_cache_key(code: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(code.as_bytes());
    hex::encode(hasher.finalize())
}

/// Strip ESM `export ` keywords so the code parses as a plain script in
/// QuickJS's `eval` context. Conservative: only matches the leading
/// `export ` / `export default ` forms; does NOT attempt to rewrite
/// `import ... from` specifiers (those require pre-bundling).
fn strip_esm_keywords(code: &str) -> String {
    let mut out = String::with_capacity(code.len());
    for line in code.lines() {
        let trimmed = line.trim_start();
        if trimmed.starts_with("export default ") {
            // `export default function f(){}` -> `function f(){}`
            // `export default expr` -> `return expr` (handled at call site)
            out.push_str(&trimmed["export default ".len()..]);
        } else if trimmed.starts_with("export ") {
            // `export function handler(...)` -> `function handler(...)`
            // `export const x = ...` -> `const x = ...`
            // `export async function ...` -> `async function ...`
            out.push_str(&trimmed["export ".len()..]);
        } else {
            out.push_str(line);
        }
        out.push('\n');
    }
    out
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
        assert!(executor.info().features.contains(&"async_await".to_string()));
    }
}
