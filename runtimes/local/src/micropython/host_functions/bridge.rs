//! FunctionFly Python Bridge for MicroPython WASM.
//!
//! Provides the `_functionfly` Python module that bridges Python code
//! running in MicroPython to FunctionFly platform services via host
//! function calls through shared memory.
//!
//! ## Security
//!
//! - All pointer parameters are validated before use
//! - Memory bounds are checked for all read/write operations
//! - Negative pointers are rejected immediately
//! - Output sizes are validated against configured limits

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};

const MAX_OUTPUT_SIZE: usize = 64 * 1024;

/// Register all FunctionFly Python bridge host functions.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    register_ff_log(linker)?;
    register_ff_get_env(linker)?;
    register_ff_kv_get(linker)?;
    register_ff_kv_set(linker)?;
    register_ff_state_get(linker)?;
    register_ff_state_set(linker)?;
    register_ff_state_delete(linker)?;
    register_ff_state_get_fabric(linker)?;
    register_ff_state_create_snapshot(linker)?;

    tracing::debug!("Registered FunctionFly Python bridge functions");
    Ok(())
}

fn validate_memory(memory: &wasmtime::Memory, caller: &wasmtime::Caller<'_, HostState>, ptr: i32, len: i32) -> bool {
    if ptr < 0 || len < 0 {
        return false;
    }
    let mem_size = memory.data_size(caller);
    (ptr as usize) + (len as usize) <= mem_size
}

fn read_string_from_memory(memory: &wasmtime::Memory, caller: &wasmtime::Caller<'_, HostState>, ptr: i32, len: i32) -> Option<String> {
    if ptr < 0 || len < 0 {
        return None;
    }
    let ptr = ptr as usize;
    let len = len as usize;
    let mem_size = memory.data_size(caller);
    if ptr + len > mem_size {
        return None;
    }
    let mut buf = vec![0u8; len];
    if memory.read(caller, ptr, &mut buf).is_err() {
        return None;
    }
    String::from_utf8(buf).ok()
}

fn write_string_to_memory(memory: &wasmtime::Memory, caller: &mut wasmtime::Caller<'_, HostState>, ptr: i32, s: &str) -> bool {
    if ptr < 0 {
        return false;
    }
    let ptr = ptr as usize;
    let mem_size = memory.data_size(&*caller);
    if ptr + s.len() > mem_size {
        return false;
    }
    memory.write(&mut *caller, ptr, s.as_bytes()).is_ok()
}

fn write_i32_to_memory(memory: &wasmtime::Memory, caller: &mut wasmtime::Caller<'_, HostState>, ptr: i32, val: i32) -> bool {
    if ptr < 0 {
        return false;
    }
    let ptr = ptr as usize;
    let mem_size = memory.data_size(&*caller);
    if ptr + 4 > mem_size {
        return false;
    }
    let bytes = val.to_le_bytes();
    memory.write(&mut *caller, ptr, &bytes).is_ok()
}

fn register_ff_log(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_log",
        |mut caller: wasmtime::Caller<'_, HostState>, level: i32, msg_ptr: i32, msg_len: i32| {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return,
            };

            let message = match read_string_from_memory(&memory, &caller, msg_ptr, msg_len) {
                Some(s) => s,
                None => return,
            };

            match level {
                0 => tracing::debug!(target: "functionfly", "{}", message),
                1 => tracing::info!(target: "functionfly", "{}", message),
                2 => tracing::warn!(target: "functionfly", "{}", message),
                3 => tracing::error!(target: "functionfly", "{}", message),
                _ => tracing::info!(target: "functionfly", "{}", message),
            }

            if let Ok(mut logs) = caller.data().logs.try_write() {
                logs.push(message);
                if logs.len() > 1000 {
                    logs.drain(0..100);
                }
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_log: {}", e)))?;
    Ok(())
}

fn register_ff_get_env(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_get_env",
        |mut caller: wasmtime::Caller<'_, HostState>, name_ptr: i32, name_len: i32, val_ptr: i32, val_len_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let name = match read_string_from_memory(&memory, &caller, name_ptr, name_len) {
                Some(s) => s,
                None => return -2,
            };

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

            if value.len() > MAX_OUTPUT_SIZE {
                tracing::warn!("ff_get_env: value too large ({} bytes)", value.len());
                return -1;
            }

            if !write_string_to_memory(&memory, &mut caller, val_ptr, &value) {
                return -2;
            }

            if !write_i32_to_memory(&memory, &mut caller, val_len_ptr, value.len() as i32) {
                return -2;
            }

            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_get_env: {}", e)))?;
    Ok(())
}

fn register_ff_kv_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_kv_get",
        |mut caller: wasmtime::Caller<'_, HostState>, key_ptr: i32, key_len: i32, val_ptr: i32, val_len_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let key = match read_string_from_memory(&memory, &caller, key_ptr, key_len) {
                Some(s) => s,
                None => return -2,
            };

            let state = caller.data();
            let kv_get = match &state.kv_get {
                Some(f) => f.clone(),
                None => return -3,
            };

            match kv_get(&key) {
                Ok(value) => {
                    if value.len() > MAX_OUTPUT_SIZE {
                        return -1;
                    }
                    if !write_string_to_memory(&memory, &mut caller, val_ptr, &value) {
                        return -2;
                    }
                    if !write_i32_to_memory(&memory, &mut caller, val_len_ptr, value.len() as i32) {
                        return -2;
                    }
                    0
                }
                Err(_) => -1,
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_kv_get: {}", e)))?;
    Ok(())
}

fn register_ff_kv_set(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_kv_set",
        |mut caller: wasmtime::Caller<'_, HostState>, key_ptr: i32, key_len: i32, val_ptr: i32, val_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let key = match read_string_from_memory(&memory, &caller, key_ptr, key_len) {
                Some(s) => s,
                None => return -2,
            };

            let value = match read_string_from_memory(&memory, &caller, val_ptr, val_len) {
                Some(s) => s,
                None => return -2,
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
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_kv_set: {}", e)))?;
    Ok(())
}

fn register_ff_state_get(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_state_get",
        |mut caller: wasmtime::Caller<'_, HostState>, path_ptr: i32, path_len: i32, val_ptr: i32, val_len_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let path = match read_string_from_memory(&memory, &caller, path_ptr, path_len) {
                Some(s) => s,
                None => return -2,
            };

            let state = caller.data();
            let state_get = match &state.state_get {
                Some(f) => f.clone(),
                None => return -3,
            };

            match state_get(&path) {
                Ok(value) => {
                    if value.len() > MAX_OUTPUT_SIZE {
                        return -1;
                    }
                    if !write_string_to_memory(&memory, &mut caller, val_ptr, &value) {
                        return -2;
                    }
                    if !write_i32_to_memory(&memory, &mut caller, val_len_ptr, value.len() as i32) {
                        return -2;
                    }
                    0
                }
                Err(_) => -1,
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_state_get: {}", e)))?;
    Ok(())
}

fn register_ff_state_set(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_state_set",
        |mut caller: wasmtime::Caller<'_, HostState>, path_ptr: i32, path_len: i32, val_ptr: i32, val_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let path = match read_string_from_memory(&memory, &caller, path_ptr, path_len) {
                Some(s) => s,
                None => return -2,
            };

            let value = match read_string_from_memory(&memory, &caller, val_ptr, val_len) {
                Some(s) => s,
                None => return -2,
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
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_state_set: {}", e)))?;
    Ok(())
}

fn register_ff_state_delete(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_state_delete",
        |mut caller: wasmtime::Caller<'_, HostState>, path_ptr: i32, path_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let path = match read_string_from_memory(&memory, &caller, path_ptr, path_len) {
                Some(s) => s,
                None => return -2,
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
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_state_delete: {}", e)))?;
    Ok(())
}

fn register_ff_state_get_fabric(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_state_get_fabric",
        |mut caller: wasmtime::Caller<'_, HostState>, fabric_id_ptr: i32, fabric_id_len: i32, resp_ptr: i32, resp_len_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let fabric_id = match read_string_from_memory(&memory, &caller, fabric_id_ptr, fabric_id_len) {
                Some(s) => s,
                None => return -2,
            };

            let state = caller.data();
            let get_fabric = match &state.state_get_fabric {
                Some(f) => f.clone(),
                None => return -3,
            };

            match get_fabric(&fabric_id) {
                Ok(value) => {
                    if value.len() > MAX_OUTPUT_SIZE {
                        return -1;
                    }
                    if !write_string_to_memory(&memory, &mut caller, resp_ptr, &value) {
                        return -2;
                    }
                    if !write_i32_to_memory(&memory, &mut caller, resp_len_ptr, value.len() as i32) {
                        return -2;
                    }
                    0
                }
                Err(_) => -1,
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_state_get_fabric: {}", e)))?;
    Ok(())
}

fn register_ff_state_create_snapshot(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    linker.func_wrap(
        "env",
        "ff_state_create_snapshot",
        |mut caller: wasmtime::Caller<'_, HostState>, path_ptr: i32, path_len: i32, label_ptr: i32, label_len: i32, resp_ptr: i32, resp_len_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let path = match read_string_from_memory(&memory, &caller, path_ptr, path_len) {
                Some(s) => s,
                None => return -2,
            };

            let label = if label_len > 0 {
                match read_string_from_memory(&memory, &caller, label_ptr, label_len) {
                    Some(s) => s,
                    None => return -2,
                }
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
                    if value.len() > MAX_OUTPUT_SIZE {
                        return -1;
                    }
                    if !write_string_to_memory(&memory, &mut caller, resp_ptr, &value) {
                        return -2;
                    }
                    if !write_i32_to_memory(&memory, &mut caller, resp_len_ptr, value.len() as i32) {
                        return -2;
                    }
                    0
                }
                Err(_) => -1,
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register ff_state_create_snapshot: {}", e)))?;
    Ok(())
}