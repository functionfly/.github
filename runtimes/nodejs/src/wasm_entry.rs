//! WASM Entry Points
//!
//! Exposes the WASM ABI expected by Go's JavyRuntime so this binary can
//! serve as the JS execution backend for the Go orchestrator in daemon mode.
//!
//! In daemon mode, the orchestrator sends HTTP requests to this binary's HTTP
//! server. The wasm_entry module provides the JS execution primitives that
//! wrap around QuickJS (via rquickjs) and Wasmtime for WASM execution.
//!
//! # Wasmtime Security Hardening
//!
//! When the `wasmtime` feature is enabled, WASM binaries are executed with:
//! - **Memory limits**: Enforced at the engine level via ResourceLimiter
//! - **CPU limits**: Fuel consumption tracking prevents infinite loops
//! - **Wall-clock timeouts**: Epoch interruption for deadline enforcement
//! - **Stack limits**: Max stack size prevents stack overflow attacks

#[cfg(feature = "wasmtime")]
use std::sync::Arc;

use once_cell::sync::Lazy;
use std::sync::RwLock;
use tracing::{error, info, warn};

use crate::executor::{JsContext, NodeExecutor};
use crate::RuntimeConfig;
use crate::RuntimeError;

#[cfg(feature = "wasmtime")]
use crate::engine::{WasmEngine, WasmEngineConfig};

static EXECUTOR: Lazy<RwLock<Option<NodeExecutor>>> = Lazy::new(|| RwLock::new(None));
static LAST_ERROR: Lazy<RwLock<String>> = Lazy::new(|| RwLock::new(String::new()));

#[cfg(feature = "wasmtime")]
static WASM_ENGINE: Lazy<RwLock<Option<Arc<WasmEngine>>>> = Lazy::new(|| RwLock::new(None));

fn set_error(err: &str) {
    if let Ok(mut e) = LAST_ERROR.write() {
        *e = err.to_string();
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

#[no_mangle]
pub extern "C" fn init() -> i32 {
    info!("WASM init called (daemon mode)");
    let config = RuntimeConfig::default();
    match NodeExecutor::new(config) {
        Ok(executor) => {
            if let Ok(mut exec_guard) = EXECUTOR.write() {
                *exec_guard = Some(executor);
            }
            #[cfg(feature = "wasmtime")]
            {
                init_wasm_engine();
            }
            info!("WASM init successful");
            0
        }
        Err(e) => {
            set_error(&format!("failed to create executor: {}", e));
            error!("WASM init failed: {}", e);
            -1
        }
    }
}

#[cfg(feature = "wasmtime")]
fn init_wasm_engine() {
    let engine_config = WasmEngineConfig {
        max_memory_mb: 128,
        max_timeout_ms: 30000,
        max_wasm_stack: 512 * 1024,
        fuel_per_ms: 10_000,
        consume_fuel: true,
        epoch_interruption: true,
    };

    match WasmEngine::new(engine_config) {
        Ok(engine) => {
            if let Ok(mut guard) = WASM_ENGINE.write() {
                *guard = Some(Arc::new(engine));
            }
            info!("Wasmtime engine initialized successfully");
        }
        Err(e) => {
            error!("Failed to initialize Wasmtime engine: {}", e);
        }
    }
}

#[no_mangle]
pub extern "C" fn execute(_input_ptr: i32, _input_len: i32) -> i32 {
    info!("WASM execute called");
    0
}

#[no_mangle]
pub extern "C" fn load_code(_code_ptr: i32, _code_len: i32) -> i32 {
    info!("WASM load_code called");
    0
}

#[no_mangle]
pub extern "C" fn alloc(size: i32) -> i32 {
    if size <= 0 || size > (64 * 1024 * 1024) {
        0
    } else {
        size
    }
}

#[no_mangle]
pub extern "C" fn dealloc(_ptr: i32, _size: i32) {}

#[no_mangle]
pub extern "C" fn memory() -> i32 {
    0
}

#[no_mangle]
pub extern "C" fn get_last_error() -> i32 {
    if get_error().is_empty() {
        0
    } else {
        1
    }
}

#[no_mangle]
pub extern "C" fn dealloc_error(_ptr: i32) {
    clear_error();
}

pub fn execute_js(code: &str, input_json: &str) -> Result<String, RuntimeError> {
    let needs_init = EXECUTOR.read().map(|g| g.is_none()).unwrap_or(true);
    if needs_init {
        let executor = NodeExecutor::new(RuntimeConfig::default())
            .map_err(|e| RuntimeError::NotReady(e.to_string()))?;
        if let Ok(mut guard) = EXECUTOR.write() {
            *guard = Some(executor);
        }
    }

    let guard = EXECUTOR
        .read()
        .map_err(|_| RuntimeError::NotReady("lock poisoned".to_string()))?;
    let _executor = guard
        .as_ref()
        .ok_or_else(|| RuntimeError::NotReady("not initialized".to_string()))?;

    let mut ctx = JsContext::new(false, &std::collections::HashMap::new())
        .map_err(|e| RuntimeError::Execution(format!("context error: {}", e)))?;

    ctx.load_module(code)
        .map_err(|e| RuntimeError::Compilation(format!("load failed: {}", e)))?;
    ctx.call_handler(input_json)
        .map_err(|e| RuntimeError::Execution(e.to_string()))
}

pub fn init_daemon() -> Result<(), RuntimeError> {
    let executor = NodeExecutor::new(RuntimeConfig::default())?;
    if let Ok(mut guard) = EXECUTOR.write() {
        *guard = Some(executor);
    }

    #[cfg(feature = "wasmtime")]
    {
        init_wasm_engine();
    }

    info!("Daemon JS runtime initialized");
    Ok(())
}

#[cfg(feature = "wasmtime")]
pub fn execute_wasm_binary(wasm_bytes: &[u8], input_json: &str) -> Result<String, RuntimeError> {
    let engine = WASM_ENGINE
        .read()
        .map_err(|_| RuntimeError::NotReady("WASM engine lock poisoned".to_string()))?
        .clone()
        .ok_or_else(|| RuntimeError::NotReady("WASM engine not initialized".to_string()))?;

    let timeout_ms = 30000u64;
    let _deadline = std::time::Instant::now() + std::time::Duration::from_millis(timeout_ms);

    let epoch_handle = engine.spawn_epoch_ticker();

    let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        engine.execute(wasm_bytes, input_json)
    }));

    drop(epoch_handle);

    match result {
        Ok(Ok(output)) => Ok(output),
        Ok(Err(e)) => {
            let msg = e.to_string();
            if msg.contains("deadline") || msg.contains("epoch") || msg.contains("fuel") {
                warn!("WASM execution timeout/fuel limit: {}", msg);
                Err(RuntimeError::Timeout(timeout_ms))
            } else if msg.contains("memory") || msg.contains("Memory") {
                warn!("WASM memory limit exceeded: {}", msg);
                Err(RuntimeError::MemoryLimit(msg))
            } else {
                error!("WASM execution error: {}", msg);
                Err(RuntimeError::Execution(msg))
            }
        }
        Err(panic_info) => {
            let msg = if let Some(s) = panic_info.downcast_ref::<&str>() {
                s.to_string()
            } else if let Some(s) = panic_info.downcast_ref::<String>() {
                s.clone()
            } else {
                "Unknown panic".to_string()
            };
            error!("WASM execution panicked: {}", msg);
            Err(RuntimeError::Execution(format!("WASM panic: {}", msg)))
        }
    }
}

#[cfg(not(feature = "wasmtime"))]
pub fn execute_wasm_binary(_wasm_bytes: &[u8], _input_json: &str) -> Result<String, RuntimeError> {
    Err(RuntimeError::Execution(
        "WASM binary execution requires wasmtime feature (compile with --features wasmtime)"
            .to_string(),
    ))
}
