//! Python runtime execution engine.
//!
//! This module provides Python function execution using RustPython VM.
//! Python functions execute directly in the RustPython interpreter for
//! better performance and full Python 3.11+ compatibility.

use anyhow::Context;
use std::sync::Arc;

use crate::config::Config;

use crate::python::runtime::{PythonRuntime, PythonConfig};

// ============================================================================
// Python-specific types
// ============================================================================

/// Python function metadata embedded in the WASM module
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct PythonFunctionMetadata {
    /// Function name
    pub name: String,
    /// Python version used
    pub python_version: String,
    /// Runtime version
    pub runtime_version: String,
    /// Entry point function name
    pub entry_point: String,
    /// Dependencies (packages used)
    pub dependencies: Vec<String>,
    /// Memory requirement in MB
    pub memory_mb: u32,
    /// Whether the function uses network
    pub uses_network: bool,
    /// Whether the function uses filesystem
    pub uses_filesystem: bool,
}

impl Default for PythonFunctionMetadata {
    fn default() -> Self {
        Self {
            name: "main".to_string(),
            python_version: "3.11".to_string(),
            runtime_version: "rustpython-0.4".to_string(),
            entry_point: "handler".to_string(),
            dependencies: vec![],
            memory_mb: 128,
            uses_network: false,
            uses_filesystem: false,
        }
    }
}

/// Python execution result
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct PythonExecutionResult {
    /// Output from the function
    pub output: String,
    /// Whether the execution was successful
    pub success: bool,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub exec_time_ms: u64,
    /// Memory used in bytes
    pub memory_used: u64,
}

impl PythonExecutionResult {
    /// Create a successful result
    pub fn success(output: String, exec_time_ms: u64) -> Self {
        Self {
            output,
            success: true,
            error: None,
            exec_time_ms,
            memory_used: 0,
        }
    }

    /// Create a failed result
    pub fn failure(error: String, exec_time_ms: u64) -> Self {
        Self {
            output: String::new(),
            success: false,
            error: Some(error),
            exec_time_ms,
            memory_used: 0,
        }
    }
}

// ============================================================================
// Python Engine
// ============================================================================

/// Python execution engine for running Python functions using RustPython
pub struct PythonEngine {
    config: Config,
    runtime: PythonRuntime,
}

impl PythonEngine {
    /// Create a new Python engine
    pub fn new(config: Config) -> anyhow::Result<Self> {
        // Create Python runtime configuration
        let python_config = crate::python::runtime::PythonConfig::from(config.clone());

        // Create Python runtime
        let runtime = PythonRuntime::new(python_config)?;

        Ok(Self { config, runtime })
    }


    /// Execute Python code using the RustPython runtime (synchronous)
    pub fn execute_sync(&self, python_code: &str, input: &str) -> anyhow::Result<String> {
        // Execute using the Python runtime with RustPython
        self.runtime.execute_sync(python_code, input)
    }

    /// Execute Python code using the RustPython runtime (async wrapper)
    pub async fn execute(&self, python_code: &str, input: &str) -> anyhow::Result<String> {
        // Execute using the Python runtime with RustPython
        self.runtime.execute(python_code, input).await
    }


    /// Execute with timeout and resource limits
    #[allow(dead_code)]
    pub async fn execute_with_limits(
        &self,
        python_code: &str,
        input: &str,
    ) -> anyhow::Result<String> {
        let timeout_duration = std::time::Duration::from_millis(self.config.timeout_ms);

        tokio::time::timeout(timeout_duration, self.execute(python_code, input))
            .await
            .context("Python execution timeout exceeded")?
    }

    /// Check if the given code is Python code
    pub fn is_python_code(code: &str) -> bool {
        // Simple check for Python code patterns
        code.contains("def ") ||
        code.contains("import ") ||
        code.contains("print(") ||
        code.contains("return ")
    }
}


/// Shared Python engine state
pub struct PythonSharedState {
    pub engine: Arc<PythonEngine>,
    pub config: PythonConfig,
}

impl PythonSharedState {
    pub fn new(engine: PythonEngine, config: PythonConfig) -> Self {
        Self {
            engine: Arc::new(engine),
            config,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_python_code() {
        // Valid Python code
        let python_code = "def hello():\n    return 'world'";
        assert!(PythonEngine::is_python_code(python_code));

        // Non-Python code
        let non_python = "console.log('hello');";
        assert!(!PythonEngine::is_python_code(non_python));
    }


    #[test]
    fn test_metadata_default() {
        let meta = PythonFunctionMetadata::default();
        assert_eq!(meta.name, "main");
        assert_eq!(meta.entry_point, "handler");
        assert_eq!(meta.runtime_version, "rustpython-0.4");
    }

    #[test]
    fn test_execution_result() {
        let success = PythonExecutionResult::success("hello".to_string(), 10);
        assert!(success.success);
        assert_eq!(success.output, "hello");

        let failure = PythonExecutionResult::failure("error".to_string(), 10);
        assert!(!failure.success);
        assert!(failure.error.is_some());
    }

    #[tokio::test]
    async fn test_python_engine_creation() {
        let config = Config {
            port: 8787,
            function: "test".to_string(),
            version: "1.0.0".to_string(),
            wasm: None,
            runtime: "python".to_string(),
            memory_mb: 128,
            timeout_ms: 5000,
            deterministic: false,
            cache_ttl: 3600,
            verbose: false,
            wasi_enabled: true,
            cpu_fuel_limit: 1000000,
            max_cpu_time_ms: 5000,
            enable_monitoring: true,
            hardened_security: true,
            max_concurrent_per_function: 10,
            memory_overhead_percent: 10,
            wasi_dirs: vec![],
            wasi_env: vec![],
            wasi_args: vec![],
            wasi_allow_network: false,
            wasi_allow_time: true,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: vec![],
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: "http://localhost:8080".to_string(),
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tier: "ultra-low".to_string(),
            network_whitelist: vec![],
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
            max_output_bytes: 1024 * 1024,
            max_input_bytes: 1024 * 1024,
            microvm_fallback_allowed: true,
            cors_allow_origin: "*".to_string(),
        };

        let engine = PythonEngine::new(config);
        assert!(engine.is_ok());
    }
}
