//! MicroPython execution orchestrator.
//!
//! This module provides the high-level interface for executing Python code
//! using the linked MicroPython WASM runtime.

use super::errors::{ExecutionErrorCode, MicroPythonError, Result};
use super::loader::MicroPythonLoader;
use super::memory::{HostState, MemoryLayout};
use super::wrapper::{WrapperConfig, WrapperGenerator};
use once_cell::sync::Lazy;
use std::sync::Arc;
use wasmtime::{Engine, Module, Store};

/// Shared runtime for synchronous execution. Avoids creating a new runtime per call.
static EXECUTOR_RUNTIME: Lazy<Arc<tokio::runtime::Runtime>> = Lazy::new(|| {
    Arc::new(
        tokio::runtime::Runtime::new().expect("MicroPython executor runtime"),
    )
});

/// Bootstrap code injected before user Python code to provide the _functionfly module.
/// This module bridges Python calls to WASM host functions via shared memory.
const _FUNCTIONFLY_BOOTSTRAP: &str = r#"
import sys as _sys
class _FunctionFly:
    """WASM host function bridge - injected by FunctionFly runtime."""
    _available = False
    def _try_init(self):
        try:
            import _functionfly as _ff
            self._ff = _ff
            self._available = True
        except ImportError:
            self._available = False
    def __init__(self):
        self._try_init()
    def state_get(self, path):
        if self._available: return self._ff.state_get(path)
        return None
    def state_set(self, path, value):
        if self._available: return self._ff.state_set(path, value)
        return False
    def state_delete(self, path):
        if self._available: return self._ff.state_delete(path)
        return False
    def state_get_fabric(self, fabric_id):
        if self._available: return self._ff.state_get_fabric(fabric_id)
        return None
    def state_create_snapshot(self, path, label=''):
        if self._available: return self._ff.state_create_snapshot(path, label)
        return None
    def get_env(self, name):
        if self._available: return self._ff.get_env(name)
        import os; return os.environ.get(name)
    def kv_get(self, key):
        if self._available: return self._ff.kv_get(key)
        return None
    def kv_set(self, key, value):
        if self._available: return self._ff.kv_set(key, value)
        return False
    def log(self, level, message):
        if self._available: self._ff.log(level, message)
    def is_available(self):
        return self._available
_functionfly = _FunctionFly()
_sys.modules['_functionfly'] = _functionfly
"#;

/// Configuration for MicroPython execution.
#[derive(Debug, Clone)]
pub struct ExecutorConfig {
    /// Path to the micropython-full.wasm file
    pub mp_wasm_path: String,
    /// Heap size in KB for MicroPython
    pub heap_size_kb: u32,
    /// Execution timeout in milliseconds
    pub timeout_ms: u64,
    /// Enable debug logging
    pub debug: bool,
}

impl Default for ExecutorConfig {
    fn default() -> Self {
        Self {
            mp_wasm_path: "assets/micropython-full.wasm".to_string(),
            heap_size_kb: 512, // 512KB default heap
            timeout_ms: 5000,  // 5 second default timeout
            debug: false,
        }
    }
}

/// Executor for Python code using linked MicroPython WASM.
pub struct MicroPythonExecutor {
    /// The MicroPython WASM loader
    loader: Arc<MicroPythonLoader>,
    /// Wrapper module generator
    wrapper_gen: WrapperGenerator,
    /// Execution configuration
    config: ExecutorConfig,
    /// Wasmtime engine
    engine: Engine,
}

impl MicroPythonExecutor {
    /// Create a new MicroPython executor.
    pub fn new(config: ExecutorConfig) -> Result<Self> {
        let engine = Engine::default();

        // Load the MicroPython WASM module
        let loader = if std::path::Path::new(&config.mp_wasm_path).exists() {
            Arc::new(MicroPythonLoader::from_file(&engine, &config.mp_wasm_path)?)
        } else {
            // Try alternative paths
            let alt_paths = [
                "internal/bundler/python/micropython-full.wasm",
                "runtimes/local/assets/micropython-full.wasm",
                "./micropython-full.wasm",
            ];

            let mut loader = None;
            for path in &alt_paths {
                if std::path::Path::new(path).exists() {
                    loader = Some(Arc::new(MicroPythonLoader::from_file(&engine, path)?));
                    break;
                }
            }

            loader.ok_or_else(|| {
                MicroPythonError::LoadError(format!(
                    "MicroPython WASM not found at {} or standard locations",
                    config.mp_wasm_path
                ))
            })?
        };

        // Create wrapper generator with appropriate memory layout
        let memory_layout = MemoryLayout::with_heap_size(config.heap_size_kb);
        let wrapper_config = WrapperConfig {
            memory: memory_layout,
            debug: config.debug,
            max_output_size: 64 * 1024,
        };
        let wrapper_gen = WrapperGenerator::with_config(wrapper_config);

        Ok(Self {
            loader,
            wrapper_gen,
            config,
            engine,
        })
    }

    /// Create executor with explicit engine and pre-loaded loader.
    #[allow(dead_code)]
    pub fn with_loader(engine: Engine, loader: Arc<MicroPythonLoader>, config: ExecutorConfig) -> Self {
        let memory_layout = MemoryLayout::with_heap_size(config.heap_size_kb);
        let wrapper_config = WrapperConfig {
            memory: memory_layout,
            debug: config.debug,
            max_output_size: 64 * 1024,
        };
        let wrapper_gen = WrapperGenerator::with_config(wrapper_config);

        Self {
            loader,
            wrapper_gen,
            config,
            engine,
        }
    }

    /// Execute Python code with the given input.
    ///
    /// # Arguments
    /// * `python_code` - The Python source code to execute
    /// * `input` - JSON input data passed to the handler function
    ///
    /// # Returns
    /// The output from the Python execution as a string.
    pub async fn execute(&self, python_code: &str, input: &str) -> Result<String> {
        self.execute_with_wrapper(python_code, input).await
    }

    /// Internal execution with full wrapper generation and error handling.
    async fn execute_with_wrapper(&self, python_code: &str, input: &str) -> Result<String> {
        // Prepend _functionfly bootstrap module to make host functions
        // available via `from _functionfly import state_get` etc.
        let bootstrap = _FUNCTIONFLY_BOOTSTRAP;
        let wrapped_code = format!("{}\n{}", bootstrap, python_code);

        // Generate wrapper module with embedded Python code
        let wrapper_wasm = self.wrapper_gen.generate(&wrapped_code)?;
        let wrapper_module = Module::new(&self.engine, &wrapper_wasm).map_err(|e| {
            MicroPythonError::WrapperError(format!("Failed to compile wrapper: {}", e))
        })?;

        // Create store with host state
        let host_state = HostState::new(input);
        let mut store = Store::new(&self.engine, host_state);

        // Set fuel limit for execution (prevents infinite loops)
        let fuel_limit = self.config.timeout_ms * 1000; // Convert ms to fuel units
        store.set_fuel(fuel_limit).map_err(|_e| {
            MicroPythonError::ExecutionError(-1)
        })?;

        // Create linked instance
        let linked = self.loader.create_linked_instance(&mut store, &wrapper_module)?;

        // Call mp_js_init to initialize the MicroPython runtime
        let init_func = linked
            .get_typed_func::<i32, i32>(&mut store, "mp_js_init")
            .map_err(|e| MicroPythonError::LinkError(format!("mp_js_init not found: {}", e)))?;

        let heap_size = (self.config.heap_size_kb * 1024) as i32;
        let init_result = init_func.call(&mut store, heap_size).map_err(|_e| {
            MicroPythonError::ExecutionError(-1)
        })?;

        if init_result != 0 {
            let code = ExecutionErrorCode::from_i32(init_result);
            tracing::error!(error_code = ?code, description = code.description(), "MicroPython init failed");
            return Err(MicroPythonError::ExecutionError(init_result));
        }

        // Get code location from wrapper
        let code_base = self.wrapper_gen.config().memory.code_buffer_base;
        let code_len = python_code.len() as i32;

        // Call mp_js_do_exec to execute the Python code
        let exec_func = linked
            .get_typed_func::<(i32, i32), i32>(&mut store, "mp_js_do_exec")
            .map_err(|e| MicroPythonError::LinkError(format!("mp_js_do_exec not found: {}", e)))?;

        let exec_result = exec_func
            .call(&mut store, (code_base as i32, code_len))
            .map_err(|e| {
                let err_str = e.to_string();
                if err_str.contains("fuel") || err_str.contains("out of fuel") || err_str.contains("fuel was not consumed") {
                    tracing::error!("MicroPython execution timed out (fuel exhausted)");
                    MicroPythonError::TimeoutError
                } else {
                    MicroPythonError::ExecutionError(-1)
                }
            })?;

        if exec_result != 0 {
            let code = ExecutionErrorCode::from_i32(exec_result);
            tracing::error!(error_code = ?code, description = code.description(), "MicroPython execution failed");
            return Err(MicroPythonError::ExecutionError(exec_result));
        }

        // Read output from shared linear memory (output buffer) first
        let layout = &self.wrapper_gen.config().memory;
        let output_base = layout.output_buffer_base;
        let output_size = layout.output_buffer_size;
        let mut buf = vec![0u8; output_size as usize];
        linked
            .memory()
            .read(&store, output_base as usize, &mut buf)
            .map_err(|e| MicroPythonError::MemoryError(format!("Failed to read output buffer: {}", e)))?;
        let len = buf.iter().position(|&b| b == 0).unwrap_or(buf.len());
        let output_from_memory = String::from_utf8_lossy(&buf[..len])
            .trim_end_matches('\0')
            .to_string();

        // Use shared-memory output if non-empty, otherwise host state (e.g. from host_set_output)
        let output = if output_from_memory.is_empty() {
            store.data().get_output().await
        } else {
            output_from_memory
        };

        Ok(output)
    }

    /// Execute Python code synchronously.
    ///
    /// This is useful for testing and blocking contexts. Uses a shared runtime
    /// so a new Tokio runtime is not created on every call.
    #[allow(dead_code)]
    pub fn execute_sync(&self, python_code: &str, input: &str) -> Result<String> {
        EXECUTOR_RUNTIME.block_on(self.execute(python_code, input))
    }

    /// Get the executor configuration.
    pub fn config(&self) -> &ExecutorConfig {
        &self.config
    }

    /// Update the executor configuration.
    #[allow(dead_code)]
    pub fn set_config(&mut self, config: ExecutorConfig) {
        self.config = config;
    }

    /// Get the loader reference
    #[allow(dead_code)]
    pub fn loader(&self) -> &Arc<MicroPythonLoader> {
        &self.loader
    }

    pub fn is_ready(&self) -> bool {
        true
    }
}

impl MicroPythonExecutor {
    /// Execute with a specific Python code string.
    /// This is the primary method for executing MicroPython code.
    pub fn execute_with_code(&self, python_code: &str, input: &str) -> anyhow::Result<String> {
        let runtime = tokio::runtime::Runtime::new()
            .map_err(|e| anyhow::anyhow!("Failed to create runtime: {}", e))?;
        runtime.block_on(self.execute(python_code, input))
            .map_err(|e| anyhow::anyhow!("MicroPython execution failed: {}", e))
    }
}

impl std::fmt::Debug for MicroPythonExecutor {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("MicroPythonExecutor")
            .field("config", &self.config)
            .finish_non_exhaustive()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_executor_config_default() {
        let config = ExecutorConfig::default();
        assert_eq!(config.heap_size_kb, 512);
        assert_eq!(config.timeout_ms, 5000);
        assert!(!config.debug);
    }

    #[test]
    fn test_executor_new_without_wasm() {
        // This should fail because we don't have the WASM file in test environment
        let config = ExecutorConfig::default();
        let result = MicroPythonExecutor::new(config);
        // Expected to fail since micropython-full.wasm doesn't exist in test environment
        assert!(result.is_err());
    }

    #[test]
    fn test_executor_debug() {
        let config = ExecutorConfig::default();
        // Can't create executor without WASM, so just test the config
        assert_eq!(config.mp_wasm_path, "assets/micropython-full.wasm");
    }
}
