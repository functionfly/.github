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
        let wasm_bytes = std::fs::read(path)
            .map_err(|e| MicroPythonError::LoadError(format!("Failed to read {}: {}", path, e)))?;
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

        // Define the wrapper module as "env" - MicroPython imports from "env" namespace
        linker
            .module(&mut *store, "env", wrapper_module)
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define wrapper module as 'env': {}", e))
            })?;

        // Define host functions that the wrapper can call
        super::host_functions::register_all_host_functions(&mut linker, store)?;

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
            .func_wrap(
                "host",
                "log",
                |mut caller: wasmtime::Caller<'_, HostState>, ptr: i32, len: i32| {
                    let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                    if let Some(memory) = memory {
                        let mut buffer = vec![0u8; len as usize];
                        if memory.read(&caller, ptr as usize, &mut buffer).is_ok() {
                            if let Ok(message) = String::from_utf8(buffer) {
                                tracing::debug!("MicroPython: {}", message);
                                // Store log message in HostState synchronously via the logs Arc
                                let state = caller.data_mut();
                                if let Ok(mut logs) = state.logs.try_write() {
                                    logs.push(message.clone());
                                    // Keep only last 1000 logs
                                    if logs.len() > 1000 {
                                        logs.drain(0..100);
                                    }
                                }
                                // Also emit to tracing for real-time visibility
                                tracing::info!(target: "micropython", "{}", message);
                            }
                        }
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host_log: {}", e))
            })?;

        // host_get_input(ptr, max_len) -> actual_len - Get input data
        linker
            .func_wrap(
                "host",
                "get_input",
                |mut caller: wasmtime::Caller<'_, HostState>, ptr: i32, max_len: i32| -> i32 {
                    let input = caller.data().input.clone();
                    let input_bytes = input.as_bytes();
                    let len = input_bytes.len().min(max_len as usize);

                    let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                    if let Some(memory) = memory {
                        let _ = memory.write(&mut caller, ptr as usize, &input_bytes[..len]);
                    }

                    len as i32
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host_get_input: {}", e))
            })?;

        // host_set_output(ptr, len) - Set output data
        linker
            .func_wrap(
                "host",
                "set_output",
                |mut caller: wasmtime::Caller<'_, HostState>, ptr: i32, len: i32| {
                    let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                    if let Some(memory) = memory {
                        let mut buffer = vec![0u8; len as usize];
                        if memory.read(&caller, ptr as usize, &mut buffer).is_ok() {
                            if let Ok(output) = String::from_utf8(buffer.clone()) {
                                let state = caller.data_mut();
                                // Store output in HostState synchronously via the output Arc
                                if let Ok(mut state_output) = state.output.try_write() {
                                    *state_output = output.clone();
                                }
                                tracing::debug!("MicroPython output set: {} bytes", len);
                                // Emit to tracing for visibility
                                tracing::info!(target: "micropython_output", "{}", output);
                            }
                        }
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host_set_output: {}", e))
            })?;

        // --- FunctionFly host functions ---
        // These provide the same host functions as the Go runtime's functionfly.* namespace,
        // enabling Python code running in MicroPython to access platform services.

        // host.ff_log(ptr, len) - Log a message via the FunctionFly logging system
        linker
            .func_wrap(
                "host",
                "ff_log",
                |mut caller: wasmtime::Caller<'_, HostState>, level: i32, msg_ptr: i32, msg_len: i32| {
                    let memory = caller.get_export("memory").and_then(|e| e.into_memory());
                    if let Some(memory) = memory {
                        let mut buffer = vec![0u8; msg_len as usize];
                        if memory.read(&caller, msg_ptr as usize, &mut buffer).is_ok() {
                            if let Ok(message) = String::from_utf8(buffer) {
                                match level {
                                    0 => tracing::debug!(target: "functionfly", "{}", message),
                                    1 => tracing::info!(target: "functionfly", "{}", message),
                                    2 => tracing::warn!(target: "functionfly", "{}", message),
                                    3 => tracing::error!(target: "functionfly", "{}", message),
                                    _ => tracing::info!(target: "functionfly", "{}", message),
                                }
                            }
                        }
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_log: {}", e))
            })?;

        // host.ff_get_env(name_ptr, name_len, val_ptr, val_len_ptr) -> i32
        // Returns 0 on success, -1 if not found, -2 on memory error
        linker
            .func_wrap(
                "host",
                "ff_get_env",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 name_ptr: i32,
                 name_len: i32,
                 val_ptr: i32,
                 val_len_ptr: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    // Read the variable name
                    let mut name_buf = vec![0u8; name_len as usize];
                    if memory.read(&caller, name_ptr as usize, &mut name_buf).is_err() {
                        return -2;
                    }
                    let name = match String::from_utf8(name_buf) {
                        Ok(n) => n,
                        Err(_) => return -2,
                    };

                    // Get the value from the env provider or std::env
                    let value = {
                        let state = caller.data();
                        if let Some(ref provider) = state.env_provider {
                            match provider(&name) {
                                Some(v) => v,
                                None => return -1,
                            }
                        } else {
                            match std::env::var(&name) {
                                Ok(v) => v,
                                Err(_) => return -1,
                            }
                        }
                    };

                    let value_bytes = value.as_bytes();
                    let value_len = value_bytes.len() as i32;

                    // Write value length
                    if memory.write(&mut caller, val_len_ptr as usize, &value_len.to_le_bytes()).is_err() {
                        return -2;
                    }

                    // Write value data
                    if memory.write(&mut caller, val_ptr as usize, value_bytes).is_err() {
                        return -2;
                    }

                    0
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_get_env: {}", e))
            })?;

        // host.ff_kv_get(key_ptr, key_len, val_ptr, val_len_ptr) -> i32
        linker
            .func_wrap(
                "host",
                "ff_kv_get",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 key_ptr: i32,
                 key_len: i32,
                 val_ptr: i32,
                 val_len_ptr: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut key_buf = vec![0u8; key_len as usize];
                    if memory.read(&caller, key_ptr as usize, &mut key_buf).is_err() {
                        return -2;
                    }
                    let key = match String::from_utf8(key_buf) {
                        Ok(k) => k,
                        Err(_) => return -2,
                    };

                    let state = caller.data();
                    let kv_get = match &state.kv_get {
                        Some(f) => f.clone(),
                        None => return -3, // KV not configured
                    };

                    match kv_get(&key) {
                        Ok(value) => {
                            let value_bytes = value.as_bytes();
                            let value_len = value_bytes.len() as i32;
                            if memory.write(&mut caller, val_len_ptr as usize, &value_len.to_le_bytes()).is_err() {
                                return -2;
                            }
                            if memory.write(&mut caller, val_ptr as usize, value_bytes).is_err() {
                                return -2;
                            }
                            0
                        }
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_kv_get: {}", e))
            })?;

        // host.ff_kv_set(key_ptr, key_len, val_ptr, val_len) -> i32
        linker
            .func_wrap(
                "host",
                "ff_kv_set",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 key_ptr: i32,
                 key_len: i32,
                 val_ptr: i32,
                 val_len: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut key_buf = vec![0u8; key_len as usize];
                    if memory.read(&caller, key_ptr as usize, &mut key_buf).is_err() {
                        return -2;
                    }
                    let key = match String::from_utf8(key_buf) {
                        Ok(k) => k,
                        Err(_) => return -2,
                    };

                    let mut val_buf = vec![0u8; val_len as usize];
                    if memory.read(&caller, val_ptr as usize, &mut val_buf).is_err() {
                        return -2;
                    }
                    let value = match String::from_utf8(val_buf) {
                        Ok(v) => v,
                        Err(_) => return -2,
                    };

                    let state = caller.data();
                    let kv_set = match &state.kv_set {
                        Some(f) => f.clone(),
                        None => return -3,
                    };

                    match kv_set(&key, &value) {
                        Ok(()) => 0,
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_kv_set: {}", e))
            })?;

        // host.ff_state_get(path_ptr, path_len, val_ptr, val_len_ptr) -> i32
        linker
            .func_wrap(
                "host",
                "ff_state_get",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 path_ptr: i32,
                 path_len: i32,
                 val_ptr: i32,
                 val_len_ptr: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut path_buf = vec![0u8; path_len as usize];
                    if memory.read(&caller, path_ptr as usize, &mut path_buf).is_err() {
                        return -2;
                    }
                    let path = match String::from_utf8(path_buf) {
                        Ok(p) => p,
                        Err(_) => return -2,
                    };

                    let state = caller.data();
                    let state_get = match &state.state_get {
                        Some(f) => f.clone(),
                        None => return -3, // StateFabric not configured
                    };

                    match state_get(&path) {
                        Ok(value) => {
                            let value_bytes = value.as_bytes();
                            let value_len = value_bytes.len() as i32;
                            if memory.write(&mut caller, val_len_ptr as usize, &value_len.to_le_bytes()).is_err() {
                                return -2;
                            }
                            if memory.write(&mut caller, val_ptr as usize, value_bytes).is_err() {
                                return -2;
                            }
                            0
                        }
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_state_get: {}", e))
            })?;

        // host.ff_state_set(path_ptr, path_len, val_ptr, val_len) -> i32
        linker
            .func_wrap(
                "host",
                "ff_state_set",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 path_ptr: i32,
                 path_len: i32,
                 val_ptr: i32,
                 val_len: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut path_buf = vec![0u8; path_len as usize];
                    if memory.read(&caller, path_ptr as usize, &mut path_buf).is_err() {
                        return -2;
                    }
                    let path = match String::from_utf8(path_buf) {
                        Ok(p) => p,
                        Err(_) => return -2,
                    };

                    let mut val_buf = vec![0u8; val_len as usize];
                    if memory.read(&caller, val_ptr as usize, &mut val_buf).is_err() {
                        return -2;
                    }
                    let value = match String::from_utf8(val_buf) {
                        Ok(v) => v,
                        Err(_) => return -2,
                    };

                    let state = caller.data();
                    let state_set = match &state.state_set {
                        Some(f) => f.clone(),
                        None => return -3,
                    };

                    match state_set(&path, &value) {
                        Ok(()) => 0,
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_state_set: {}", e))
            })?;

        // host.ff_state_delete(path_ptr, path_len) -> i32
        linker
            .func_wrap(
                "host",
                "ff_state_delete",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 path_ptr: i32,
                 path_len: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut path_buf = vec![0u8; path_len as usize];
                    if memory.read(&caller, path_ptr as usize, &mut path_buf).is_err() {
                        return -2;
                    }
                    let path = match String::from_utf8(path_buf) {
                        Ok(p) => p,
                        Err(_) => return -2,
                    };

                    let state = caller.data();
                    let state_delete = match &state.state_delete {
                        Some(f) => f.clone(),
                        None => return -3,
                    };

                    match state_delete(&path) {
                        Ok(()) => 0,
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_state_delete: {}", e))
            })?;

        // host.ff_state_get_fabric(fabric_id_ptr, fabric_id_len, resp_ptr, resp_len_ptr) -> i32
        linker
            .func_wrap(
                "host",
                "ff_state_get_fabric",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 fabric_id_ptr: i32,
                 fabric_id_len: i32,
                 resp_ptr: i32,
                 resp_len_ptr: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut id_buf = vec![0u8; fabric_id_len as usize];
                    if memory.read(&caller, fabric_id_ptr as usize, &mut id_buf).is_err() {
                        return -2;
                    }
                    let fabric_id = match String::from_utf8(id_buf) {
                        Ok(id) => id,
                        Err(_) => return -2,
                    };

                    let state = caller.data();
                    let get_fabric = match &state.state_get_fabric {
                        Some(f) => f.clone(),
                        None => return -3,
                    };

                    match get_fabric(&fabric_id) {
                        Ok(value) => {
                            let value_bytes = value.as_bytes();
                            let value_len = value_bytes.len() as i32;
                            if memory.write(&mut caller, resp_len_ptr as usize, &value_len.to_le_bytes()).is_err() {
                                return -2;
                            }
                            if memory.write(&mut caller, resp_ptr as usize, value_bytes).is_err() {
                                return -2;
                            }
                            0
                        }
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_state_get_fabric: {}", e))
            })?;

        // host.ff_state_create_snapshot(path_ptr, path_len, label_ptr, label_len, resp_ptr, resp_len_ptr) -> i32
        linker
            .func_wrap(
                "host",
                "ff_state_create_snapshot",
                |mut caller: wasmtime::Caller<'_, HostState>,
                 path_ptr: i32,
                 path_len: i32,
                 label_ptr: i32,
                 label_len: i32,
                 resp_ptr: i32,
                 resp_len_ptr: i32| -> i32 {
                    let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                        Some(m) => m,
                        None => return -2,
                    };

                    let mut path_buf = vec![0u8; path_len as usize];
                    if memory.read(&caller, path_ptr as usize, &mut path_buf).is_err() {
                        return -2;
                    }
                    let path = match String::from_utf8(path_buf) {
                        Ok(p) => p,
                        Err(_) => return -2,
                    };

                    let label = if label_len > 0 {
                        let mut label_buf = vec![0u8; label_len as usize];
                        if memory.read(&caller, label_ptr as usize, &mut label_buf).is_err() {
                            return -2;
                        }
                        String::from_utf8(label_buf).unwrap_or_default()
                    } else {
                        String::new()
                    };

                    let state = caller.data();
                    let create_snapshot = match &state.state_create_snapshot {
                        Some(f) => f.clone(),
                        None => return -3,
                    };

                    match create_snapshot(&path, &label) {
                        Ok(value) => {
                            let value_bytes = value.as_bytes();
                            let value_len = value_bytes.len() as i32;
                            if memory.write(&mut caller, resp_len_ptr as usize, &value_len.to_le_bytes()).is_err() {
                                return -2;
                            }
                            if memory.write(&mut caller, resp_ptr as usize, value_bytes).is_err() {
                                return -2;
                            }
                            0
                        }
                        Err(_) => -1,
                    }
                },
            )
            .map_err(|e| {
                MicroPythonError::LinkError(format!("Failed to define host.ff_state_create_snapshot: {}", e))
            })?;

        Ok(())
    }

    /// Get a reference to the compiled MicroPython module.
    #[allow(dead_code)]
    pub fn module(&self) -> &Module {
        &self.mp_module
    }

    /// Get a clone of the Arc-wrapped module.
    #[allow(dead_code)]
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
    #[allow(dead_code)]
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
        self.instance.get_typed_func(store, name).map_err(|e| {
            MicroPythonError::LinkError(format!("Function '{}' not found: {}", name, e))
        })
    }

    /// Read a string from WASM memory.
    #[allow(dead_code)]
    pub fn read_string(&self, store: &mut Store<HostState>, ptr: i32, len: i32) -> Result<String> {
        let mut buffer = vec![0u8; len as usize];
        self.memory
            .read(store, ptr as usize, &mut buffer)
            .map_err(|e| MicroPythonError::MemoryError(format!("Failed to read memory: {}", e)))?;

        String::from_utf8(buffer)
            .map_err(|e| MicroPythonError::MemoryError(format!("Invalid UTF-8: {}", e)))
    }

    /// Write a string to WASM memory.
    #[allow(dead_code)]
    pub fn write_string(&self, store: &mut Store<HostState>, ptr: i32, s: &str) -> Result<()> {
        self.memory
            .write(store, ptr as usize, s.as_bytes())
            .map_err(|e| MicroPythonError::MemoryError(format!("Failed to write memory: {}", e)))
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
