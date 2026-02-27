//! MicroPython execution orchestrator.
//!
//! This module provides the high-level interface for executing Python code
//! using the linked MicroPython WASM runtime.

use super::errors::{MicroPythonError, Result};
use super::loader::{LinkedInstance, MicroPythonLoader};
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
        // Generate wrapper module with embedded Python code
        let wrapper_wasm = self.wrapper_gen.generate(python_code)?;
        let wrapper_module = Module::new(&self.engine, &wrapper_wasm).map_err(|e| {
            MicroPythonError::WrapperError(format!("Failed to compile wrapper: {}", e))
        })?;

        // Create store with host state
        let host_state = HostState::new(input);
        let mut store = Store::new(&self.engine, host_state);

        // Set fuel limit for execution (prevents infinite loops)
        let fuel_limit = self.config.timeout_ms * 1000; // Convert ms to fuel units
        store.set_fuel(fuel_limit).map_err(|e| {
            MicroPythonError::ExecutionError(-1)
        })?;

        // Create linked instance
        let linked = self.loader.create_linked_instance(&mut store, &wrapper_module)?;

        // Call mp_js_init to initialize the MicroPython runtime
        let init_func = linked
            .get_typed_func::<i32, i32>(&mut store, "mp_js_init")
            .map_err(|e| MicroPythonError::LinkError(format!("mp_js_init not found: {}", e)))?;

        let heap_size = (self.config.heap_size_kb * 1024) as i32;
        let init_result = init_func.call(&mut store, heap_size).map_err(|e| {
            MicroPythonError::ExecutionError(-1)
        })?;

        if init_result != 0 {
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
            .map_err(|e| MicroPythonError::ExecutionError(-1))?;

        if exec_result != 0 {
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
    pub fn execute_sync(&self, python_code: &str, input: &str) -> Result<String> {
        EXECUTOR_RUNTIME.block_on(self.execute(python_code, input))
    }

    /// Get the executor configuration.
    pub fn config(&self) -> &ExecutorConfig {
        &self.config
    }

    /// Update the executor configuration.
    pub fn set_config(&mut self, config: ExecutorConfig) {
        self.config = config;
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
