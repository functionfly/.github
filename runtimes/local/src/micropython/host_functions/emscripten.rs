//! Emscripten compatibility functions for MicroPython WASM.
//!
//! Provides Emscripten-specific functions that MicroPython and compiled
//! code may call. In serverless context, memory growth is denied and
//! threading is not supported.
//!
//! ## Security
//!
//! - Dynamic memory growth is denied to prevent exhaustion
//! - Threading operations are denied (single-threaded serverless)
//! - Async operations are serialized through the host call queue
//! - All pointer arguments are validated before use

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};

fn validate_memory_ptr(memory: &wasmtime::Memory, caller: &wasmtime::Caller<'_, HostState>, ptr: i32, len: i32) -> bool {
    if ptr < 0 || len < 0 {
        return false;
    }
    let mem_size = memory.data_size(caller);
    (ptr as usize) + (len as usize) <= mem_size
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

/// Register all Emscripten compatibility functions.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    register_memory_functions(linker)?;
    register_time_functions(linker)?;
    register_async_functions(linker)?;
    register_process_functions(linker)?;
    register_console_functions(linker)?;
    register_thread_functions(linker)?;

    tracing::debug!("Registered Emscripten compatibility functions");
    Ok(())
}

fn register_memory_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // emscripten_memcpy(dst: i32, src: i32, len: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_memcpy",
        |mut caller: wasmtime::Caller<'_, HostState>, dst: i32, src: i32, len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            if dst < 0 || src < 0 || len < 0 {
                tracing::warn!(target: "emscripten", "emscripten_memcpy: invalid pointer/len");
                return -1;
            }

            let mem_size = memory.data_size(&caller);
            let dst = dst as usize;
            let src = src as usize;
            let len = len as usize;

            if dst + len > mem_size || src + len > mem_size {
                tracing::warn!(target: "emscripten", "emscripten_memcpy: out of bounds");
                return -1;
            }

            let mut buf = vec![0u8; len];
            if let Err(e) = memory.read(&caller, src, &mut buf) {
                tracing::error!(target: "emscripten", "emscripten_memcpy: read failed: {}", e);
                return -1;
            }

            if let Err(e) = memory.write(&mut caller, dst, &buf) {
                tracing::error!(target: "emscripten", "emscripten_memcpy: write failed: {}", e);
                return -1;
            }

            dst as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_memcpy: {}", e)))?;

    // emscripten_memset(dst: i32, val: i32, len: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_memset",
        |mut caller: wasmtime::Caller<'_, HostState>, dst: i32, val: i32, len: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            if dst < 0 || len < 0 {
                return -1;
            }

            let mem_size = memory.data_size(&caller);
            let dst = dst as usize;
            let len = len as usize;

            if dst + len > mem_size {
                return -1;
            }

            let val = (val & 0xFF) as u8;
            let buf = vec![val; len];

            if let Err(e) = memory.write(&mut caller, dst, &buf) {
                tracing::error!(target: "emscripten", "emscripten_memset: write failed: {}", e);
                return -1;
            }

            dst as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_memset: {}", e)))?;

    // emscripten_memalign(ptr_ptr: i32, align: i32, size: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_memalign",
        |mut caller: wasmtime::Caller<'_, HostState>, ptr_ptr: i32, align: i32, size: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            if ptr_ptr < 0 || align < 0 || size < 0 {
                return -1;
            }

            // Alignment must be power of 2
            if align & (align - 1) != 0 {
                tracing::warn!(target: "emscripten", "emscripten_memalign: invalid alignment {}", align);
                return -1;
            }

            let mem_size = memory.data_size(&caller);
            if (ptr_ptr as usize) + 4 > mem_size {
                return -1;
            }

            // In serverless, we don't actually allocate - just return a pointer
            // The guest should use the stack or existing allocations
            // Return 0 to indicate "use current break" (like sbrk)
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_memalign: {}", e)))?;

    // emscripten_get_heap_size() -> i32
    linker.func_wrap(
        "env",
        "emscripten_get_heap_size",
        |mut caller: wasmtime::Caller<'_, HostState>| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return 0,
            };

            (memory.data_size(&caller) / (64 * 1024)) as i32
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_get_heap_size: {}", e)))?;

    // emscripten_get_now() -> f64
    linker.func_wrap(
        "env",
        "emscripten_get_now",
        |_caller: wasmtime::Caller<'_, HostState>| -> f64 {
            std::time::Instant::now()
                .elapsed()
                .as_secs_f64() * 1000.0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_get_now: {}", e)))?;

    // emscripten_resize_heap(size: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_resize_heap",
        |_caller: wasmtime::Caller<'_, HostState>, _size: i32| -> i32 {
            tracing::warn!(target: "emscripten", "emscripten_resize_heap: denied - dynamic memory growth not allowed in serverless");
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_resize_heap: {}", e)))?;

    // emscripten_scan_registers(callback: i32)
    linker.func_wrap(
        "env",
        "emscripten_scan_registers",
        |_caller: wasmtime::Caller<'_, HostState>, _ptr: i32| {},
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_scan_registers: {}", e)))?;

    Ok(())
}

fn register_time_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // emscripten_date_now() -> f64
    linker.func_wrap(
        "env",
        "emscripten_date_now",
        |_caller: wasmtime::Caller<'_, HostState>| -> f64 {
            std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_secs_f64() * 1000.0)
                .unwrap_or(0.0)
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_date_now: {}", e)))?;

    // emscripten_get_minimum_heap_size() -> i32
    linker.func_wrap(
        "env",
        "emscripten_get_minimum_heap_size",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            512 * 1024 // 512KB minimum
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_get_minimum_heap_size: {}", e)))?;

    // emscripten_get_expected_maximum_heap_size() -> i32
    linker.func_wrap(
        "env",
        "emscripten_get_expected_maximum_heap_size",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            512 * 1024 // 512KB expected max in serverless
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_get_expected_maximum_heap_size: {}", e)))?;

    Ok(())
}

fn register_async_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // emscripten_async_call(func: i32, arg: i32, delay: i32)
    linker.func_wrap(
        "env",
        "emscripten_async_call",
        |mut caller: wasmtime::Caller<'_, HostState>, func: i32, arg: i32, _delay: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "emscripten_async_call".to_string(),
                    args: format!(r#"{{"func":{},"arg":{}}}"#, func, arg),
                    call_id: state.next_call_id.fetch_add(1, std::sync::atomic::Ordering::SeqCst),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_async_call: {}", e)))?;

    // emscripten_sync_run_in_main_thread(func: i32, arg: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_sync_run_in_main_thread",
        |mut caller: wasmtime::Caller<'_, HostState>, func: i32, arg: i32| -> i32 {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "emscripten_sync_run_in_main_thread".to_string(),
                    args: format!(r#"{{"func":{},"arg":{}}}"#, func, arg),
                    call_id: state.next_call_id.fetch_add(1, std::sync::atomic::Ordering::SeqCst),
                };
                calls.push(call);
            }
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_sync_run_in_main_thread: {}", e)))?;

    // emscripten_run_in_main_thread(func: i32, arg: i32, flags: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_run_in_main_thread",
        |mut caller: wasmtime::Caller<'_, HostState>, func: i32, arg: i32, _flags: i32| -> i32 {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "emscripten_run_in_main_thread".to_string(),
                    args: format!(r#"{{"func":{},"arg":{}}}"#, func, arg),
                    call_id: state.next_call_id.fetch_add(1, std::sync::atomic::Ordering::SeqCst),
                };
                calls.push(call);
            }
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_run_in_main_thread: {}", e)))?;

    // emscripten_async_run_in_main_thread(func: i32, arg: i32, flags: i32)
    linker.func_wrap(
        "env",
        "emscripten_async_run_in_main_thread",
        |mut caller: wasmtime::Caller<'_, HostState>, func: i32, arg: i32, _flags: i32| {
            let state = caller.data_mut();
            if let Ok(mut calls) = state.pending_calls.try_write() {
                let call = crate::micropython::memory::HostFunctionCall {
                    function: "emscripten_async_run_in_main_thread".to_string(),
                    args: format!(r#"{{"func":{},"arg":{}}}"#, func, arg),
                    call_id: state.next_call_id.fetch_add(1, std::sync::atomic::Ordering::SeqCst),
                };
                calls.push(call);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_async_run_in_main_thread: {}", e)))?;

    Ok(())
}

fn register_process_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // _emscripten_throw_longjmp - Called for longjmp-based exceptions
    linker.func_wrap(
        "env",
        "_emscripten_throw_longjmp",
        |_caller: wasmtime::Caller<'_, HostState>| {
            tracing::debug!(target: "emscripten", "_emscripten_throw_longjmp: called");
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register _emscripten_throw_longjmp: {}", e)))?;

    // emscripten_exit_with_live_runtime
    linker.func_wrap(
        "env",
        "emscripten_exit_with_live_runtime",
        |_caller: wasmtime::Caller<'_, HostState>| {},
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_exit_with_live_runtime: {}", e)))?;

    // emscripten_force_exit(code: i32)
    linker.func_wrap(
        "env",
        "emscripten_force_exit",
        |_caller: wasmtime::Caller<'_, HostState>, code: i32| {
            tracing::info!(target: "emscripten", "emscripten_force_exit: code={}", code);
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_force_exit: {}", e)))?;

    // emscripten_get_now wrapper for profiling
    linker.func_wrap(
        "env",
        "emscripten_performance_now",
        |_caller: wasmtime::Caller<'_, HostState>| -> f64 {
            std::time::Instant::now()
                .elapsed()
                .as_secs_f64() * 1000.0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_performance_now: {}", e)))?;

    // emscripten_get_previous_tagged_computation_time(tag: i32) -> f64
    linker.func_wrap(
        "env",
        "emscripten_get_previous_tagged_computation_time",
        |_caller: wasmtime::Caller<'_, HostState>, _tag: i32| -> f64 {
            0.0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_get_previous_tagged_computation_time: {}", e)))?;

    // emscripten_enter_soft_realtime() -> i32
    linker.func_wrap(
        "env",
        "emscripten_enter_soft_realtime",
        |_caller: wasmtime::Caller<'_, HostState>, _memory_handle: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_enter_soft_realtime: {}", e)))?;

    // emscripten_exit_soft_realtime()
    linker.func_wrap(
        "env",
        "emscripten_exit_soft_realtime",
        |_caller: wasmtime::Caller<'_, HostState>| {},
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_exit_soft_realtime: {}", e)))?;

    Ok(())
}

fn register_console_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // emscripten_out(str: i32)
    linker.func_wrap(
        "env",
        "emscripten_out",
        |mut caller: wasmtime::Caller<'_, HostState>, str_ptr: i32| {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return,
            };

            // Read string until null terminator
            let mem_size = memory.data_size(&caller);
            let str_ptr = str_ptr as usize;

            let mut buf = Vec::new();
            for i in str_ptr..mem_size.min(str_ptr + 4096) {
                let mut byte = [0u8; 1];
                if memory.read(&caller, i, &mut byte).is_err() {
                    break;
                }
                if byte[0] == 0 {
                    break;
                }
                buf.push(byte[0]);
            }

            if let Ok(s) = String::from_utf8(buf) {
                tracing::info!(target: "emscripten_out", "{}", s);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_out: {}", e)))?;

    // emscripten_err(str: i32)
    linker.func_wrap(
        "env",
        "emscripten_err",
        |mut caller: wasmtime::Caller<'_, HostState>, str_ptr: i32| {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return,
            };

            let mem_size = memory.data_size(&caller);
            let str_ptr = str_ptr as usize;

            let mut buf = Vec::new();
            for i in str_ptr..mem_size.min(str_ptr + 4096) {
                let mut byte = [0u8; 1];
                if memory.read(&caller, i, &mut byte).is_err() {
                    break;
                }
                if byte[0] == 0 {
                    break;
                }
                buf.push(byte[0]);
            }

            if let Ok(s) = String::from_utf8(buf) {
                tracing::error!(target: "emscripten_err", "{}", s);
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_err: {}", e)))?;

    // emscripten_print(str: i32, from_javascript: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_print",
        |mut caller: wasmtime::Caller<'_, HostState>, str_ptr: i32, _from_javascript: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let mem_size = memory.data_size(&caller);
            let str_ptr = str_ptr as usize;

            let mut buf = Vec::new();
            for i in str_ptr..mem_size.min(str_ptr + 4096) {
                let mut byte = [0u8; 1];
                if memory.read(&caller, i, &mut byte).is_err() {
                    break;
                }
                if byte[0] == 0 {
                    break;
                }
                buf.push(byte[0]);
            }

            if let Ok(s) = String::from_utf8(buf) {
                tracing::info!(target: "emscripten_print", "{}", s);
                s.len() as i32
            } else {
                -1
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_print: {}", e)))?;

    // emscripten_log(level: i32, str_ptr: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_log",
        |mut caller: wasmtime::Caller<'_, HostState>, level: i32, str_ptr: i32| -> i32 {
            let memory = match caller.get_export("memory").and_then(|e| e.into_memory()) {
                Some(m) => m,
                None => return -1,
            };

            let mem_size = memory.data_size(&caller);
            let str_ptr = str_ptr as usize;

            let mut buf = Vec::new();
            for i in str_ptr..mem_size.min(str_ptr + 4096) {
                let mut byte = [0u8; 1];
                if memory.read(&caller, i, &mut byte).is_err() {
                    break;
                }
                if byte[0] == 0 {
                    break;
                }
                buf.push(byte[0]);
            }

            if let Ok(s) = String::from_utf8(buf) {
                match level {
                    0 => tracing::debug!(target: "emscripten", "{}", s),
                    1 => tracing::info!(target: "emscripten", "{}", s),
                    2 => tracing::warn!(target: "emscripten", "{}", s),
                    3 => tracing::error!(target: "emscripten", "{}", s),
                    _ => tracing::info!(target: "emscripten", "{}", s),
                }
                s.len() as i32
            } else {
                -1
            }
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_log: {}", e)))?;

    Ok(())
}

fn register_thread_functions(linker: &mut Linker<HostState>) -> Result<(), MicroPythonError> {
    // emscripten_has_threading_support() -> i32
    linker.func_wrap(
        "env",
        "emscripten_has_threading_support",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            0 // No threading in serverless
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_has_threading_support: {}", e)))?;

    // emscripten_is_main_browser_thread() -> i32
    linker.func_wrap(
        "env",
        "emscripten_is_main_browser_thread",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            1 // Always main thread in serverless
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_is_main_browser_thread: {}", e)))?;

    // emscripten_main_browser_thread_id() -> i32
    linker.func_wrap(
        "env",
        "emscripten_main_browser_thread_id",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_main_browser_thread_id: {}", e)))?;

    // emscripten_threads_init()
    linker.func_wrap(
        "env",
        "emscripten_threads_init",
        |_caller: wasmtime::Caller<'_, HostState>| {},
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_threads_init: {}", e)))?;

    // emscripten_thread_init(thread_id: i32, stack_size: i32)
    linker.func_wrap(
        "env",
        "emscripten_thread_init",
        |_caller: wasmtime::Caller<'_, HostState>, thread_id: i32, _stack_size: i32| {
            tracing::debug!(target: "emscripten", "emscripten_thread_init: thread_id={} (threading not supported)", thread_id);
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_thread_init: {}", e)))?;

    // emscripten_thread_join(thread_id: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_thread_join",
        |_caller: wasmtime::Caller<'_, HostState>, thread_id: i32| -> i32 {
            tracing::warn!(target: "emscripten", "emscripten_thread_join: denied - threading not supported (thread {})", thread_id);
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_thread_join: {}", e)))?;

    // emscripten_thread_terminate(thread_id: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_thread_terminate",
        |_caller: wasmtime::Caller<'_, HostState>, thread_id: i32| -> i32 {
            tracing::warn!(target: "emscripten", "emscripten_thread_terminate: denied - threading not supported (thread {})", thread_id);
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_thread_terminate: {}", e)))?;

    // emscripten_fetch_initialize() -> i32
    linker.func_wrap(
        "env",
        "emscripten_fetch_initialize",
        |_caller: wasmtime::Caller<'_, HostState>| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_fetch_initialize: {}", e)))?;

    // emscripten_fetch_create(data_ptr: i32, data_len: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_fetch_create",
        |_caller: wasmtime::Caller<'_, HostState>, _data_ptr: i32, _data_len: i32| -> i32 {
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_fetch_create: {}", e)))?;

    // emscripten_fetch_send(fetch_handle: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_fetch_send",
        |_caller: wasmtime::Caller<'_, HostState>, _fetch_handle: i32| -> i32 {
            -1
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_fetch_send: {}", e)))?;

    // emscripten_fetch_close(fetch_handle: i32) -> i32
    linker.func_wrap(
        "env",
        "emscripten_fetch_close",
        |_caller: wasmtime::Caller<'_, HostState>, _fetch_handle: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_fetch_close: {}", e)))?;

    // emscripten_get_now wrapper for threading timing
    linker.func_wrap(
        "env",
        "emscripten_thread_time",
        |_caller: wasmtime::Caller<'_, HostState>, _thread_id: i32| -> i32 {
            0
        },
    ).map_err(|e| MicroPythonError::LinkError(format!("Failed to register emscripten_thread_time: {}", e)))?;

    Ok(())
}