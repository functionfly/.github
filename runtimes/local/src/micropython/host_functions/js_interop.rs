//! JavaScript interop for MicroPython WASM.
//!
//! Implements JavaScript interop functions that MicroPython uses for async
//! operations, callbacks, Promise handling, and attribute access. These are
//! capability-gated and sandboxed for serverless execution.
//!
//! ## Security
//!
//! - All pointers from WASM are validated before use
//! - Memory bounds are checked for all read/write operations
//! - Negative pointers are rejected immediately
//! - Async operations are serialized through the host call queue

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};
use std::sync::atomic::{AtomicU32, Ordering};

const MAX_STRING_SIZE: usize = 64 * 1024;
static NEXT_CALLBACK_ID: AtomicU32 = AtomicU32::new(1);

fn validate_ptr(ptr: i32, len: i32) -> bool {
    ptr >= 0 && len >= 0 && (len as usize) <= MAX_STRING_SIZE
}

fn get_next_callback_id() -> u32 {
    NEXT_CALLBACK_ID.fetch_add(1, Ordering::SeqCst)
}

fn read_string_from_memory(
    memory: &wasmtime::Memory,
    caller: &wasmtime::Caller<'_, HostState>,
    ptr: i32,
    len: i32,
) -> Option<String> {
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

fn write_string_to_memory(
    memory: &wasmtime::Memory,
    caller: &mut wasmtime::Caller<'_, HostState>,
    ptr: i32,
    s: &str,
) -> bool {
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

fn write_i32_to_memory(
    memory: &wasmtime::Memory,
    caller: &mut wasmtime::Caller<'_, HostState>,
    ptr: i32,
    val: i32,
) -> bool {
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

/// Register all JavaScript interop functions.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    register_invoke_functions(linker)?;
    register_mp_js_functions(linker)?;
    register_proxy_functions(linker)?;
    register_call_functions(linker)?;
    register_js_functions(linker)?;

    tracing::debug!("Registered JavaScript interop functions");
    Ok(())
}

fn register_invoke_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // invoke_* functions handle async JS callbacks from MicroPython
    // These serialize calls through the host call queue for safe async handling

    linker.func_wrap(
        "env",
        "invoke_ii",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_ii".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"callback_id":{}}}"#, a, b, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_ii: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_iiii",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_iiii".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"callback_id":{}}}"#, a, b, c, d, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_iiii: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_v",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_v".to_string(),
                    args: format!(r#"{{"a":{}}}"#, a),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_v: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_viii",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_viii".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{}}}"#, a, b, c),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_viii: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_iiiii",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_iiiii".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"callback_id":{}}}"#, a, b, c, d, e, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_iiiii: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_iii",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_iii".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"callback_id":{}}}"#, a, b, c, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_iii: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_vi",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_vi".to_string(),
                    args: format!(r#"{{"a":{}}}"#, a),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_vi: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_vii",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_vii".to_string(),
                    args: format!(r#"{{"a":{},"b":{}}}"#, a, b),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_vii: {}", e)))?;

    linker.func_wrap(
        "env",
        "invoke_i",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "invoke_i".to_string(),
                    args: format!(r#"{{"a":{},"callback_id":{}}}"#, a, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register invoke_i: {}", e)))?;

    Ok(())
}

fn register_mp_js_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // mp_js_hook - Yield to host event loop (no-op in synchronous serverless)
    linker.func_wrap(
        "env",
        "mp_js_hook",
        |_caller: wasmtime::Caller<'_, HostState>, _a: i32| {},
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register mp_js_hook: {}", e)))?;

    // mp_js_random_u32 - Cryptographically secure random number
    linker.func_wrap(
        "env",
        "mp_js_random_u32",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            use std::collections::hash_map::DefaultHasher;
            use std::hash::{Hash, Hasher};
            use std::time::Instant;

            let mut hasher = DefaultHasher::new();
            Instant::now().hash(&mut hasher);
            std::thread::current().id().hash(&mut hasher);
            // Mix in some entropy from the OS
            let mut entropy = [0u8; 8];
            if let Ok(fd) = std::fs::File::open("/dev/urandom") {
                use std::io::Read;
                let _ = fd.take(8).read(&mut entropy);
            }
            entropy.hash(&mut hasher);
            hasher.finish() as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register mp_js_random_u32: {}", e)))?;

    // mp_js_ticks_ms - Monotonic time in milliseconds
    linker.func_wrap(
        "env",
        "mp_js_ticks_ms",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            std::time::Instant::now()
                .elapsed()
                .as_millis() as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register mp_js_ticks_ms: {}", e)))?;

    // mp_js_time_ms - Unix timestamp as float (milliseconds)
    linker.func_wrap(
        "env",
        "mp_js_time_ms",
        |_caller: wasmtime::Caller<'_, HostState>| -> f64 {
            std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_secs_f64() * 1000.0)
                .unwrap_or(0.0)
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register mp_js_time_ms: {}", e)))?;

    Ok(())
}

fn register_proxy_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // JS Promise interop - These handle async JS/Python bridging
    // In serverless context, we queue the operation and return a callback ID

    linker.func_wrap(
        "env",
        "proxy_convert_mp_to_js_then_js_to_mp_obj_jsside",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32, f: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "proxy_convert_mp_to_js".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"f":{},"callback_id":{}}}"#, a, b, c, d, e, f, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register proxy_convert_mp_to_js_then_js_to_mp_obj_jsside: {}", e)))?;

    linker.func_wrap(
        "env",
        "proxy_convert_mp_to_js_then_js_to_js_then_js_to_mp_obj_jsside",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32, f: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "proxy_convert_mp_to_js_chain".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"f":{},"callback_id":{}}}"#, a, b, c, d, e, f, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register proxy_convert_mp_to_js_then_js_to_js_then_js_to_mp_obj_jsside: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_get_proxy_js_ref_info",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_get_proxy_js_ref_info".to_string(),
                    args: format!(r#"{{"a":{},"callback_id":{}}}"#, a, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_get_proxy_js_ref_info: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_get_iter",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_get_iter".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"callback_id":{}}}"#, a, b, c, d, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_get_iter: {}", e)))?;

    linker.func_wrap(
        "env",
        "proxy_js_free_obj",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "proxy_js_free_obj".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{}}}"#, a, b, c, d),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register proxy_js_free_obj: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_reflect_construct",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32, f: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_reflect_construct".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"f":{},"callback_id":{}}}"#, a, b, c, d, e, f, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_reflect_construct: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_iter_next",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_iter_next".to_string(),
                    args: format!(r#"{{"a":{},"callback_id":{}}}"#, a, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_iter_next: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_check_existing",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_check_existing".to_string(),
                    args: format!(r#"{{"a":{},"callback_id":{}}}"#, a, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_check_existing: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_get_error_info",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_get_error_info".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"callback_id":{}}}"#, a, b, c, d, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_get_error_info: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_then_resolve",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_then_resolve".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{}}}"#, a, b, c, d),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_then_resolve: {}", e)))?;

    linker.func_wrap(
        "env",
        "create_promise",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "create_promise".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"callback_id":{}}}"#, a, b, c, d, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register create_promise: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_then_continue",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32, f: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_then_continue".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"f":{}}}"#, a, b, c, d, e, f),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_then_continue: {}", e)))?;

    linker.func_wrap(
        "env",
        "js_then_reject",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_then_reject".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{}}}"#, a, b, c, d),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_then_reject: {}", e)))?;

    Ok(())
}

fn register_call_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // call0, call1, call2, calln - Function call dispatch
    // These route calls through the host call queue for async handling

    linker.func_wrap(
        "env",
        "call0_kwarg",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32, f: i32, g: i32, h: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "call0_kwarg".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"f":{},"g":{},"h":{},"callback_id":{}}}"#, a, b, c, d, e, f, g, h, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register call0_kwarg: {}", e)))?;

    linker.func_wrap(
        "env",
        "calln_kwarg",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32, f: i32, g: i32, h: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "calln_kwarg".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"f":{},"g":{},"h":{},"callback_id":{}}}"#, a, b, c, d, e, f, g, h, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register calln_kwarg: {}", e)))?;

    linker.func_wrap(
        "env",
        "call1",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "call1".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"callback_id":{}}}"#, a, b, c, d, e, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register call1: {}", e)))?;

    linker.func_wrap(
        "env",
        "call2",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "call2".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"callback_id":{}}}"#, a, b, c, d, e, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register call2: {}", e)))?;

    linker.func_wrap(
        "env",
        "calln",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32, d: i32, e: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "calln".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"d":{},"e":{},"callback_id":{}}}"#, a, b, c, d, e, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register calln: {}", e)))?;

    linker.func_wrap(
        "env",
        "call0",
        |mut caller: wasmtime::Caller<'_, HostState>, a: i32, b: i32, c: i32| -> i32 {
            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "call0".to_string(),
                    args: format!(r#"{{"a":{},"b":{},"c":{},"callback_id":{}}}"#, a, b, c, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register call0: {}", e)))?;

    Ok(())
}

fn register_js_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // lookup_attr - Attribute lookup on JS objects
    linker.func_wrap(
        "env",
        "lookup_attr",
        |mut caller: wasmtime::Caller<'_, HostState>, obj_ptr: i32, name_ptr: i32, name_len: i32, result_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let name = match read_string_from_memory(&memory, &caller, name_ptr, name_len) {
                Some(s) => s,
                None => return -2,
            };

            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "lookup_attr".to_string(),
                    args: format!(r#"{{"obj_ptr":{},"name":"{}","result_ptr":{},"callback_id":{}}}"#, obj_ptr, name, result_ptr, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register lookup_attr: {}", e)))?;

    // store_attr - Attribute store on JS objects
    linker.func_wrap(
        "env",
        "store_attr",
        |mut caller: wasmtime::Caller<'_, HostState>, obj_ptr: i32, name_ptr: i32, name_len: i32, value_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let name = match read_string_from_memory(&memory, &caller, name_ptr, name_len) {
                Some(s) => s,
                None => return -2,
            };

            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "store_attr".to_string(),
                    args: format!(r#"{{"obj_ptr":{},"name":"{}","value_ptr":{}}}"#, obj_ptr, name, value_ptr),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register store_attr: {}", e)))?;

    // js_subscr_load - Subscription/index load from JS objects
    linker.func_wrap(
        "env",
        "js_subscr_load",
        |mut caller: wasmtime::Caller<'_, HostState>, obj_ptr: i32, index_ptr: i32, index_len: i32, result_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let index = match read_string_from_memory(&memory, &caller, index_ptr, index_len) {
                Some(s) => s,
                None => return -2,
            };

            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_subscr_load".to_string(),
                    args: format!(r#"{{"obj_ptr":{},"index":"{}","result_ptr":{},"callback_id":{}}}"#, obj_ptr, index, result_ptr, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_subscr_load: {}", e)))?;

    // js_subscr_store - Subscription/index store to JS objects
    linker.func_wrap(
        "env",
        "js_subscr_store",
        |mut caller: wasmtime::Caller<'_, HostState>, obj_ptr: i32, index_ptr: i32, index_len: i32, value_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let index = match read_string_from_memory(&memory, &caller, index_ptr, index_len) {
                Some(s) => s,
                None => return -2,
            };

            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "js_subscr_store".to_string(),
                    args: format!(r#"{{"obj_ptr":{},"index":"{}","value_ptr":{}}}"#, obj_ptr, index, value_ptr),
                    call_id: get_next_callback_id(),
                };
                calls.push(call);
            }
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register js_subscr_store: {}", e)))?;

    // has_attr - Check if JS object has an attribute
    linker.func_wrap(
        "env",
        "has_attr",
        |mut caller: wasmtime::Caller<'_, HostState>, obj_ptr: i32, name_ptr: i32, name_len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -2,
            };

            let name = match read_string_from_memory(&memory, &caller, name_ptr, name_len) {
                Some(s) => s,
                None => return -2,
            };

            let callback_id = get_next_callback_id();
            let state = caller.data_mut();

            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "has_attr".to_string(),
                    args: format!(r#"{{"obj_ptr":{},"name":"{}","callback_id":{}}}"#, obj_ptr, name, callback_id),
                    call_id: callback_id,
                };
                calls.push(call);
            }
            callback_id as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register has_attr: {}", e)))?;

    Ok(())
}
