//! WASM Entry Points
//!
//! Exposes the WASM ABI expected by Go's JavyRuntime so this binary can
//! serve as the JS execution backend for the Go orchestrator in daemon mode.
//!
//! In daemon mode, the orchestrator sends HTTP requests to this binary's HTTP
//! server. The wasm_entry module provides the JS execution primitives that
//! wrap around QuickJS (via quickjs-wasm-rs).

use once_cell::sync::Lazy;
use std::sync::RwLock;
use tracing::{error, info, warn};

use crate::executor::NodeExecutor;
use crate::RuntimeConfig;
use crate::RuntimeError;

// ============================================================================
// Global JS state (shared across all calls in a single-process daemon)
// ============================================================================

static EXECUTOR: Lazy<RwLock<Option<NodeExecutor>>> = Lazy::new(|| RwLock::new(None));
static LAST_ERROR: Lazy<RwLock<String>> = Lazy::new(|| RwLock::new(String::new()));

fn set_error(err: &str) {
    match LAST_ERROR.write() {
        Ok(mut e) => *e = err.to_string(),
        Err(_) => eprintln!("Failed to set error (lock poisoned): {}", err),
    }
}

fn get_error() -> String {
    LAST_ERROR.read().map(|e| e.clone()).unwrap_or_default()
}

fn clear_error() {
    if let Ok(mut e) = LAST_ERROR.write() {
        *e = String::new();
    }
}

// ============================================================================
// WASM ABI Exports
//
// These functions follow the WASM i32/i64 ABI used by Go's JavyRuntime:
//   init()       -> i32  (0=success, -1=error)
//   execute(ptr,len) -> i32  (0=success, -1=error)
//   load_code(ptr,len) -> i32  (0=success, -1=error)
//   alloc(size)   -> i32  (pointer to allocated memory)
//   dealloc(ptr,size) -> ()
//   memory()      -> i32  (memory base address, 0 for non-WASM builds)
//   get_last_error() -> i32  (0=no error, pointer to error string)
//   dealloc_error(ptr) -> ()
// ============================================================================

/// Initialize the QuickJS runtime.
/// Must be called before any other exports.
/// Returns 0 on success, -1 on error.
#[no_mangle]
pub extern "C" fn init() -> i32 {
    info!("WASM init called (daemon mode)");

    let config = RuntimeConfig::default();

    match NodeExecutor::new(config) {
        Ok(executor) => {
            match EXECUTOR.write() {
                Ok(mut exec_guard) => *exec_guard = Some(executor),
                Err(_) => {
                    set_error("executor lock poisoned");
                    error!("WASM init failed: executor lock poisoned");
                    return -1;
                }
            }

            info!("WASM init successful");
            return 0;
        }
        Err(e) => {
            set_error(&format!("failed to create executor: {}", e));
            error!("WASM init failed: {}", e);
            return -1;
        }
    }
}

/// Execute the loaded JavaScript module with the given input JSON pointer/length.
/// In daemon mode, input_json is passed directly (not from WASM memory).
/// Returns 0 on success, -1 on error.
#[no_mangle]
pub extern "C" fn execute(input_ptr: i32, input_len: i32) -> i32 {
    info!(
        "WASM execute called ({} bytes at offset {})",
        input_len, input_ptr
    );

    // In a real WASM build, we'd read from linear memory at input_ptr.
    // In daemon mode, the Go side calls execute_js() directly instead.
    // This export exists for ABI compatibility.
    0
}

/// Load JavaScript source code into the runtime.
/// code_ptr / code_len: WASM memory pointer to source bytes (0 in daemon mode).
/// Returns 0 on success, -1 on error.
#[no_mangle]
pub extern "C" fn load_code(code_ptr: i32, code_len: i32) -> i32 {
    info!("WASM load_code called ({} bytes)", code_len);
    // In daemon mode, code is loaded via the HTTP API, not via WASM memory.
    // This export exists for ABI compatibility.
    0
}

/// Allocate memory in WASM linear memory.
/// size: bytes to allocate.
/// Returns: pointer to allocated block (0 on failure).
/// In daemon mode, memory is managed by the host process.
#[no_mangle]
pub extern "C" fn alloc(size: i32) -> i32 {
    if size <= 0 || size > (64 * 1024 * 1024) {
        return 0;
    }
    // In daemon mode, return a non-zero sentinel.
    // The Go side doesn't actually dereference this.
    size
}

/// Deallocate memory previously allocated by alloc().
#[no_mangle]
pub extern "C" fn dealloc(_ptr: i32, _size: i32) {
    // In daemon mode, memory is managed by the host process.
}

/// Return the base address of WASM linear memory.
/// Always returns 0 in daemon (non-WASM) mode.
#[no_mangle]
pub extern "C" fn memory() -> i32 {
    0
}

/// Get the last error as a pointer to a null-terminated string.
/// Returns 0 if there is no error.
#[no_mangle]
pub extern "C" fn get_last_error() -> i32 {
    if get_error().is_empty() {
        0
    } else {
        1 // fake pointer — daemon mode uses get_error_string() directly
    }
}

/// Free the last error string (no-op in daemon mode).
#[no_mangle]
pub extern "C" fn dealloc_error(_ptr: i32) {
    clear_error();
}

// ============================================================================
// Daemon-mode helpers (called by the HTTP server in main.rs)
// ============================================================================

/// Execute JavaScript code directly from the daemon HTTP handler.
pub fn execute_js(code: &str, input_json: &str) -> Result<String, RuntimeError> {
    // Ensure executor is initialized
    let needs_init = match EXECUTOR.read() {
        Ok(exec_guard) => exec_guard.is_none(),
        Err(_) => return Err(RuntimeError::NotReady("executor lock poisoned".to_string())),
    };

    if needs_init {
        let config = RuntimeConfig::default();
        let executor =
            NodeExecutor::new(config).map_err(|e| RuntimeError::NotReady(e.to_string()))?;
        match EXECUTOR.write() {
            Ok(mut exec_guard) => *exec_guard = Some(executor),
            Err(_) => return Err(RuntimeError::NotReady("executor lock poisoned".to_string())),
        }
    }

    // Acquire context while holding the executor lock
    let mut ctx = {
        match EXECUTOR.read() {
            Ok(exec_guard) => {
                let executor = exec_guard.as_ref().ok_or_else(|| {
                    RuntimeError::NotReady("executor not initialized".to_string())
                })?;
                executor.acquire_context().map_err(|e| {
                    RuntimeError::Execution(format!("failed to acquire JS context: {}", e))
                })?
            }
            Err(_) => return Err(RuntimeError::NotReady("executor lock poisoned".to_string())),
        }
    };

    // Load module
    ctx.load_module(code)
        .map_err(|e| RuntimeError::Compilation(format!("failed to load JS module: {}", e)))?;

    // Call handler
    let result = ctx
        .call_handler(input_json)
        .map_err(|e| RuntimeError::Execution(e.to_string()))?;

    // Return context to pool
    match EXECUTOR.read() {
        Ok(exec_guard) => {
            if let Some(executor) = exec_guard.as_ref() {
                executor.release_context(ctx);
            }
        }
        Err(_) => {}
    }

    Ok(result)
}

/// Initialize the global state for daemon (non-WASM) mode.
pub fn init_daemon() -> Result<(), RuntimeError> {
    let config = RuntimeConfig::default();

    let executor = NodeExecutor::new(config)?;
    match EXECUTOR.write() {
        Ok(mut exec_guard) => *exec_guard = Some(executor),
        Err(_) => return Err(RuntimeError::NotReady("executor lock poisoned".to_string())),
    }

    info!("Daemon JS runtime initialized");
    Ok(())
}

/// Execute a compiled WASM binary via wasmtime (Javy-compiled .wasm files).
/// This is called from the daemon HTTP handler when the binary is a real
/// WebAssembly module (starts with \0asm magic bytes).
pub fn execute_wasm_binary(wasm_bytes: &[u8], _input_json: &str) -> Result<String, RuntimeError> {
    // Use wasmtime to instantiate and run the compiled WASM
    use wasmtime::{Config, Engine, Linker, Module, Store};

    let mut config = Config::new();
    config.wasm_backtrace(true);

    let engine = Engine::new(&config)
        .map_err(|e| RuntimeError::Execution(format!("wasmtime engine error: {}", e)))?;

    let module = Module::from_binary(&engine, wasm_bytes)
        .map_err(|e| RuntimeError::Compilation(format!("failed to compile WASM: {}", e)))?;

    // Define a simple state type for the store
    struct WasmState;

    let mut store = Store::new(&engine, WasmState);

    let mut linker = Linker::<WasmState>::new(&engine);
    let _ = linker.allow_shadowing(true);

    // Register functionfly host functions with proper store type
    let _ = linker.func_wrap(
        "functionfly",
        "console_log",
        move |mut caller: wasmtime::Caller<'_, WasmState>, ptr: i32, len: i32| {
            if let Some(extern_item) = caller.get_export("memory") {
                if let Some(memory) = extern_item.into_memory() {
                    let data = memory.data(&caller);
                    let start = ptr as usize;
                    let end = (ptr + len) as usize;
                    if end <= data.len() {
                        let bytes = &data[start..end];
                        print!("{}", String::from_utf8_lossy(bytes));
                    }
                }
            }
        },
    );
    let _ = linker.func_wrap(
        "functionfly",
        "console_error",
        move |mut caller: wasmtime::Caller<'_, WasmState>, ptr: i32, len: i32| {
            if let Some(extern_item) = caller.get_export("memory") {
                if let Some(memory) = extern_item.into_memory() {
                    let data = memory.data(&caller);
                    let start = ptr as usize;
                    let end = (ptr + len) as usize;
                    if end <= data.len() {
                        let bytes = &data[start..end];
                        eprintln!("[error] {}", String::from_utf8_lossy(bytes));
                    }
                }
            }
        },
    );
    let _ = linker.func_wrap("functionfly", "time_ms", || {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_millis() as f64
    });

    // Instantiate with linker
    let instance = linker
        .instantiate(&mut store, &module)
        .map_err(|e| RuntimeError::Execution(format!("WASM instance error: {}", e)))?;

    // Call init if exported
    if let Ok(init_fn) = instance.get_typed_func::<(), ()>(&mut store, "init") {
        let _ = init_fn.call(&mut store, ());
    }

    // Call execute
    if let Ok(execute_fn) = instance.get_typed_func::<(), i32>(&mut store, "execute") {
        // Call with no args for now (input already baked into the module at compile time)
        let result = execute_fn
            .call(&mut store, ())
            .map_err(|e| RuntimeError::Execution(format!("execute failed: {}", e)))?;

        return Ok(format!("{{\"result_ptr\":{}}}", result));
    }

    Ok("{\"ok\":true}".to_string())
}
