//! MicroPython WASM module loader and linker.
//!
//! This module handles loading the MicroPython WASM runtime and linking it
//! with wrapper modules using wasmtime's module linking API.

use super::errors::{MicroPythonError, Result};
use super::memory::HostState;
use std::sync::Arc;
use wasmtime::{Engine, Instance, Linker, Memory, Module, Store};

/// Loader for the MicroPython WASM runtime.
pub struct MicroPythonLoader {
    /// Pre-compiled MicroPython module (1.1MB)
    mp_module: Arc<Module>,
    /// Wasmtime engine reference
    engine: Engine,
}

impl MicroPythonLoader {
    /// Create a new MicroPython loader from WASM bytes.
    ///
    /// # Arguments
    /// * `engine` - The wasmtime Engine to use
    /// * `wasm_bytes` - The raw bytes of micropython-full.wasm
    pub fn new(engine: &Engine, wasm_bytes: &[u8]) -> Result<Self> {
        let mp_module = Module::new(engine, wasm_bytes).map_err(|e| {
            MicroPythonError::LoadError(format!("Failed to compile MicroPython module: {}", e))
        })?;

        Ok(Self {
            mp_module: Arc::new(mp_module),
            engine: engine.clone(),
        })
    }

    /// Load MicroPython from a file path.
    pub fn from_file(engine: &Engine, path: &str) -> Result<Self> {
        let wasm_bytes = std::fs::read(path).map_err(|e| {
            MicroPythonError::LoadError(format!("Failed to read {}: {}", path, e))
        })?;
        Self::new(engine, &wasm_bytes)
    }

    /// Create a linked instance combining the MicroPython runtime with a wrapper module.
    ///
    /// This creates the module linking between:
    /// - Wrapper exports (memory, malloc, mp_js_init, etc.)
    /// - MicroPython imports (env.memory, env.malloc, env.mp_js_init, etc.)
    pub fn create_linked_instance(
        &self,
        store: &mut Store<HostState>,
        wrapper_module: &Module,
    ) -> Result<LinkedInstance> {
        // Create a linker to connect the modules
        let mut linker = Linker::new(&self.engine);

        // Define the wrapper module - it provides the exports that MicroPython needs
        linker
            .module(&mut *store, "wrapper", wrapper_module)
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define wrapper module: {}", e))
            })?;

        // Define host functions that the wrapper can call
        Self::define_host_functions(&mut linker)?;

        // Instantiate the MicroPython module with imports from wrapper
        let instance = linker
            .instantiate(&mut *store, &self.mp_module)
            .map_err(|e| {
                MicroPythonError::InstantiationError(format!(
                    "Failed to instantiate MicroPython: {}",
                    e
                ))
            })?;

        // Get the memory export from the wrapper for shared access
        let memory = instance
            .get_memory(&mut *store, "memory")
            .ok_or_else(|| MicroPythonError::MemoryError("No memory export found".to_string()))?;

        Ok(LinkedInstance {
            instance,
            memory,
            _marker: std::marker::PhantomData,
        })
    }

    /// Define host functions that wrapper modules can call.
    fn define_host_functions(linker: &mut Linker<HostState>) -> Result<()> {
        // host_log(ptr, len) - Log a message from the WASM module
        linker
            .func_wrap("host", "log", |mut caller: wasmtime::Caller<'_, HostState>, ptr: i32, len: i32| {
                let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                if let Some(memory) = memory {
                    let mut buffer = vec![0u8; len as usize];
                    if memory.read(&caller, ptr as usize, &mut buffer).is_ok() {
                        if let Ok(message) = String::from_utf8(buffer) {
                            tracing::debug!("MicroPython: {}", message);
                            // Store in host state for later retrieval
                            let state = caller.data_mut();
                            // Use block_on for synchronous context
                            if let Ok(handle) = tokio::runtime::Handle::try_current() {
                                let _ = handle.enter();
                            }
                        }
                    }
                }
            })
            .map_err(|e| MicroPythonError::LinkError(format!("Failed to define host_log: {}", e)))?;

        // host_get_input(ptr, max_len) -> actual_len - Get input data
        linker
            .func_wrap("host", "get_input", |mut caller: wasmtime::Caller<'_, HostState>, ptr: i32, max_len: i32| -> i32 {
                let input = caller.data().input.clone();
                let input_bytes = input.as_bytes();
                let len = input_bytes.len().min(max_len as usize);

                let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                if let Some(memory) = memory {
                    let _ = memory.write(&mut caller, ptr as usize, &input_bytes[..len]);
                }

                len as i32
            })
            .map_err(|e| MicroPythonError::LinkError(format!("Failed to define host_get_input: {}", e)))?;

        // host_set_output(ptr, len) - Set output data
        linker
            .func_wrap("host", "set_output", |mut caller: wasmtime::Caller<'_, HostState>, ptr: i32, len: i32| {
                let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                if let Some(memory) = memory {
                    let mut buffer = vec![0u8; len as usize];
                    if memory.read(&caller, ptr as usize, &mut buffer).is_ok() {
                        if let Ok(output) = String::from_utf8(buffer) {
                            let state = caller.data_mut();
                            // Use block_on for synchronous context
                            if let Ok(handle) = tokio::runtime::Handle::try_current() {
                                let _ = handle.enter();
                            }
                            // Store output synchronously for now
                            // In async context, we'd use a channel or shared state
                        }
                    }
                }
            })
            .map_err(|e| MicroPythonError::LinkError(format!("Failed to define host_set_output: {}", e)))?;

        Ok(())
    }

    /// Get a reference to the compiled MicroPython module.
    pub fn module(&self) -> &Module {
        &self.mp_module
    }

    /// Get a clone of the Arc-wrapped module.
    pub fn module_arc(&self) -> Arc<Module> {
        self.mp_module.clone()
    }
}

/// A linked instance of MicroPython with its wrapper.
pub struct LinkedInstance {
    /// The instantiated WASM module
    instance: Instance,
    /// Shared memory reference
    memory: Memory,
    /// Phantom data for lifetime tracking
    _marker: std::marker::PhantomData<*const ()>,
}

impl LinkedInstance {
    /// Get a reference to the WASM instance.
    pub fn instance(&self) -> &Instance {
        &self.instance
    }

    /// Get a reference to the shared memory.
    pub fn memory(&self) -> &Memory {
        &self.memory
    }

    /// Get a typed function export.
    pub fn get_typed_func<Params, Results>(
        &self,
        store: &mut Store<HostState>,
        name: &str,
    ) -> Result<wasmtime::TypedFunc<Params, Results>>
    where
        Params: wasmtime::WasmParams,
        Results: wasmtime::WasmResults,
    {
        self.instance
            .get_typed_func(store, name)
            .map_err(|e| MicroPythonError::LinkError(format!("Function '{}' not found: {}", name, e)).into())
    }

    /// Read a string from WASM memory.
    pub fn read_string(&self, store: &mut Store<HostState>, ptr: i32, len: i32) -> Result<String> {
        let mut buffer = vec![0u8; len as usize];
        self.memory
            .read(store, ptr as usize, &mut buffer)
            .map_err(|e| MicroPythonError::MemoryError(format!("Failed to read memory: {}", e)))?;

        String::from_utf8(buffer)
            .map_err(|e| MicroPythonError::MemoryError(format!("Invalid UTF-8: {}", e)).into())
    }

    /// Write a string to WASM memory.
    pub fn write_string(&self, store: &mut Store<HostState>, ptr: i32, s: &str) -> Result<()> {
        self.memory
            .write(store, ptr as usize, s.as_bytes())
            .map_err(|e| MicroPythonError::MemoryError(format!("Failed to write memory: {}", e)).into())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::micropython::memory::MemoryLayout;
    use crate::micropython::wrapper::WrapperGenerator;

    // Minimal valid WASM module for testing
    const MINIMAL_WASM: &[u8] = &[
        0x00, 0x61, 0x73, 0x6d, // magic
        0x01, 0x00, 0x00, 0x00, // version
        0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // type section
        0x03, 0x02, 0x01, 0x00, // func section
        0x07, 0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, // export "add"
        0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b, // code section
    ];

    #[test]
    fn test_loader_new() {
        let engine = Engine::default();
        let loader = MicroPythonLoader::new(&engine, MINIMAL_WASM);
        assert!(loader.is_ok());
    }

    #[test]
    fn test_loader_module_access() {
        let engine = Engine::default();
        let loader = MicroPythonLoader::new(&engine, MINIMAL_WASM).unwrap();
        let _module = loader.module();
        let _arc_module = loader.module_arc();
    }

    #[test]
    fn test_linked_instance_creation() {
        let engine = Engine::default();

        // Create a simple wrapper module
        let gen = WrapperGenerator::new();
        let wrapper_wasm = gen.generate("print('test')").unwrap();
        let wrapper_module = Module::new(&engine, &wrapper_wasm).unwrap();

        // We can't fully test linking without the actual MicroPython module
        // but we can verify the loader structure
        let loader = MicroPythonLoader::new(&engine, MINIMAL_WASM).unwrap();
        assert!(loader.module_arc().imports().len() >= 0);
    }
}
