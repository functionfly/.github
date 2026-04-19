//! Synchronous WASI execution function and result reading.

use wasmtime::{Engine, Module, Store};

use crate::config::Config;
use crate::errors::RuntimeError;
use crate::wasi::{WasiContext, WasiLinker};

use super::memory_limiter::{install_memory_limiter, with_limiter, LimiterGuard};

/// Synchronous WASI execution function for use in spawn_blocking.
/// Accepts an optional pre-compiled module (from the AOT cache) to skip re-compilation.
pub fn execute_wasi_sync_inner(
    engine: &Engine,
    linker: &WasiLinker,
    wasm_bytes: &[u8],
    input: &str,
    config: &Config,
    precompiled: Option<Module>,
) -> anyhow::Result<String> {
    let execution_start = std::time::Instant::now();

    // Create WASI context with input data
    let function_key = format!("{}@{}", config.function, config.version);
    let wasi_ctx = WasiContext::new_with_input(config, function_key, input)?;
    let stdout_pipe = wasi_ctx.stdout_pipe.clone();
    let stderr_pipe = wasi_ctx.stderr_pipe.clone();

    // Create store with WASI context
    let mut store = Store::new(engine, wasi_ctx.ctx);

    // Install the hard memory limiter via the thread-local mechanism.
    // The `LimiterGuard` clears the thread-local on drop so subsequent
    // executions on the same thread start clean.
    let _limiter_guard = install_memory_limiter(config.memory_mb);
    // Safety: the limiter is valid for the lifetime of the guard. The 'static
    // lifetime is a lie that Wasmtime requires for the trait object, but this
    // is safe because spawn_blocking runs each task on its own OS thread.
    store.limiter(|_data| unsafe { with_limiter(|l| l) });

    // Calibrated fuel metering: prefer timeout_ms × fuel_per_ms, else cpu_ms_limit × fuel_per_ms, else cpu_fuel_limit.
    let fuel_limit = if config.fuel_for_timeout() > 0 {
        config.fuel_for_timeout()
    } else if config.cpu_ms_limit > 0 && config.fuel_per_ms > 0 {
        config.cpu_ms_limit.saturating_mul(config.fuel_per_ms)
    } else if config.cpu_fuel_limit > 0 {
        config.cpu_fuel_limit
    } else {
        1_000_000 // absolute fallback
    };
    store.set_fuel(fuel_limit)?;

    // Compile module (or use pre-compiled AOT module)
    let module = if let Some(m) = precompiled {
        m
    } else {
        Module::new(engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!(RuntimeError::wasm_compilation(e.to_string())))?
    };

    // Instantiate module with WASI
    let instance = linker
        .linker()
        .instantiate(&mut store, &module)
        .map_err(|e| anyhow::anyhow!(RuntimeError::wasm_instantiation(e.to_string())))?;

    // Execute the function. Prefer handler when we have input so Python WASM (and other
    // handler-based modules) receive input and can return a value via memory; otherwise try _start/main.
    let handler_result_ptr: Option<i32> =
        if let Ok(func) = instance.get_typed_func::<(i32, i32), i32>(&mut store, "handler") {
            if let Some(memory) = instance.get_memory(&mut store, "memory") {
                use crate::wasm_interface::memory;

                let input_ptr = memory::write_string(&memory, &mut store, input)?;
                let input_len = input.len() as i32;

                let result_ptr = func.call(&mut store, (input_ptr, input_len)).map_err(|e| {
                    anyhow::anyhow!(RuntimeError::wasm_execution(format!(
                        "Handler function failed: {}",
                        e
                    )))
                })?;
                tracing::info!("Handler function returned: {}", result_ptr);
                Some(result_ptr)
            } else {
                return Err(anyhow::anyhow!(
                    "No memory export found for handler function"
                ));
            }
        } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "_start") {
            func.call(&mut store, ()).map_err(|e| {
                anyhow::anyhow!(RuntimeError::wasm_execution(format!(
                    "_start function failed: {}",
                    e
                )))
            })?;
            None
        } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "main") {
            func.call(&mut store, ()).map_err(|e| {
                anyhow::anyhow!(RuntimeError::wasm_execution(format!(
                    "main function failed: {}",
                    e
                )))
            })?;
            None
        } else {
            return Err(anyhow::anyhow!(RuntimeError::function_not_found(
                "handler, _start, or main"
            )));
        };

    // Log execution time
    let execution_time = execution_start.elapsed();
    tracing::info!("WASM execution completed in {:?}", execution_time);

    // If handler returned a non-null, valid pointer, extract output from memory (embedded Python
    // returns a pointer to result struct or to a string). Negative values (e.g. -1) mean "error"
    // and must not be used as memory pointers (would cause index out of bounds).
    if let Some(result_ptr) = handler_result_ptr {
        if result_ptr > 0 {
            if let Some(memory) = instance.get_memory(&mut store, "memory") {
                match read_handler_result(&memory, &store, result_ptr) {
                    Ok(s) if !s.is_empty() => return Ok(s),
                    Ok(_) => {}
                    Err(e) => tracing::debug!("Could not read handler result from memory: {}", e),
                }
            }
        } else if result_ptr < 0 {
            // Handler returned error indicator (-1 or other negative); prefer stderr for message
            let stderr = stderr_pipe.contents();
            let stdout = stdout_pipe.contents();
            if !stderr.is_empty() {
                return Err(anyhow::anyhow!(
                    "Handler error: {}",
                    String::from_utf8_lossy(&stderr)
                ));
            }
            if !stdout.is_empty() {
                return Err(anyhow::anyhow!(
                    "Handler error: {}",
                    String::from_utf8_lossy(&stdout)
                ));
            }
            return Err(anyhow::anyhow!(
                "Handler returned error indicator ({})",
                result_ptr
            ));
        }
    }

    // Fall back to stdout/stderr.
    // Phase 2.2: detect truncation and return an explicit error instead of silently
    // dropping bytes.  `MemoryOutputPipe` stops accepting bytes once its capacity is
    // reached; we detect this by comparing the pipe's byte count to the configured
    // limit.
    let stdout = stdout_pipe.contents();
    let stderr = stderr_pipe.contents();

    drop(store);
    let _ = instance;
    drop(module);

    let pipe_capacity = if config.max_output_bytes > 0 {
        config.max_output_bytes
    } else {
        1024 * 1024
    };
    if stdout.len() >= pipe_capacity {
        return Err(anyhow::anyhow!(
            "Function output was truncated: stdout reached the {} byte limit. \
             Increase --max-output-bytes to capture more output.",
            pipe_capacity
        ));
    }
    if stderr.len() >= pipe_capacity {
        return Err(anyhow::anyhow!(
            "Function output was truncated: stderr reached the {} byte limit. \
             Increase --max-output-bytes to capture more output.",
            pipe_capacity
        ));
    }

    if !stdout.is_empty() {
        Ok(String::from_utf8_lossy(&stdout).to_string())
    } else if !stderr.is_empty() {
        Err(anyhow::anyhow!(
            "WASM stderr: {}",
            String::from_utf8_lossy(&stderr)
        ))
    } else {
        Ok("".to_string())
    }
}

/// Execute WASI with a pre-compiled module and an existing store.
///
/// This is used by the pooled execution path to reuse compiled modules and WASI contexts.
///
/// Note: For pooled execution, output capture is handled by the caller via
/// `PooledWasmInstance::reset_for_execution()` which sets up fresh pipes. The
/// handler result is extracted from memory return pointer when available.
///
/// Memory growth is hard-capped via Wasmtime's `Store::limiter()` API using
/// `FunctionMemoryLimiter`, enforcing the limit declared in `config.memory_mb`.
pub fn execute_wasi_with_module_and_store(
    linker: &WasiLinker,
    module: &Module,
    store: &mut Store<wasmtime_wasi::p1::WasiP1Ctx>,
    config: &Config,
    _wasi_ctx: &wasmtime_wasi::p1::WasiP1Ctx,
) -> anyhow::Result<String> {
    // Apply hard memory cap via Store::limiter() — enforces config.memory_mb
    // at the Wasm linear-memory level so memory.grow returns -1 (not OOM panic).
    let _limiter_guard = install_memory_limiter(config.memory_mb);
    store.limiter(|_data| unsafe { with_limiter(|l| l) });
    // Instantiate module with WASI
    let instance = linker
        .linker()
        .instantiate(&mut *store, module)
        .map_err(|e| anyhow::anyhow!(RuntimeError::wasm_instantiation(e.to_string())))?;

    // Execute the function
    let handler_result_ptr: Option<i32> =
        if let Ok(func) = instance.get_typed_func::<(i32, i32), i32>(&mut *store, "handler") {
            if let Some(memory) = instance.get_memory(&mut *store, "memory") {
                use crate::wasm_interface::memory;

                // For pooled execution, we need to write the input to memory
                // The input was already set via reset_for_execution which updates stdin
                // But for handler-based modules, we need to pass input via memory
                let input = ""; // Input is passed via stdin in pooled execution
                let input_ptr = memory::write_string(&memory, &mut *store, input)?;
                let input_len = 0i32;

                let result_ptr = func
                    .call(&mut *store, (input_ptr, input_len))
                    .map_err(|e| {
                        anyhow::anyhow!(RuntimeError::wasm_execution(format!(
                            "Handler function failed: {}",
                            e
                        )))
                    })?;
                tracing::info!("Pooled handler function returned: {}", result_ptr);
                Some(result_ptr)
            } else {
                return Err(anyhow::anyhow!(
                    "No memory export found for handler function"
                ));
            }
        } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut *store, "_start") {
            func.call(&mut *store, ()).map_err(|e| {
                anyhow::anyhow!(RuntimeError::wasm_execution(format!(
                    "_start function failed: {}",
                    e
                )))
            })?;
            None
        } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut *store, "main") {
            func.call(&mut *store, ()).map_err(|e| {
                anyhow::anyhow!(RuntimeError::wasm_execution(format!(
                    "main function failed: {}",
                    e
                )))
            })?;
            None
        } else {
            return Err(anyhow::anyhow!(RuntimeError::function_not_found(
                "handler, _start, or main"
            )));
        };

    // If handler returned a non-null, valid pointer, extract output from memory
    if let Some(result_ptr) = handler_result_ptr {
        if result_ptr > 0 {
            if let Some(memory) = instance.get_memory(&mut *store, "memory") {
                match read_handler_result(&memory, &*store, result_ptr) {
                    Ok(s) if !s.is_empty() => return Ok(s),
                    Ok(_) => {}
                    Err(e) => tracing::debug!("Could not read handler result from memory: {}", e),
                }
            }
        } else if result_ptr < 0 {
            return Err(anyhow::anyhow!(
                "Handler returned error indicator ({})",
                result_ptr
            ));
        }
    }

    // Note: For pooled execution with pipes, we can't easily extract stdout/stderr
    // here because the WASI context owns the pipes. The caller handles output capture.
    Ok("".to_string())
}

/// Reads the handler return value from WASM memory. Supports (1) direct pointer to a
/// null-terminated string, and (2) embedder result structure: 12 bytes with status (0),
/// input_ref (4), result_data (8) where result_data is the pointer to the output string.
pub fn read_handler_result(
    memory: &wasmtime::Memory,
    store: &impl wasmtime::AsContext,
    result_ptr: i32,
) -> anyhow::Result<String> {
    use crate::wasm_interface::memory;

    // Negative values (e.g. -1) are error indicators from the guest, not valid pointers.
    if result_ptr < 0 {
        return Err(anyhow::anyhow!(
            "Handler returned error pointer {}",
            result_ptr
        ));
    }

    let data = memory.data(store);
    let ptr = result_ptr as usize;
    // Reject out-of-bounds ptr (e.g. -1 cast to usize becomes huge)
    if ptr >= data.len() || ptr + 12 > data.len() {
        return Err(anyhow::anyhow!(
            "Handler result pointer out of bounds: {}",
            result_ptr
        ));
    }

    if ptr + 12 <= data.len() {
        let result_data_ptr =
            i32::from_le_bytes([data[ptr + 8], data[ptr + 9], data[ptr + 10], data[ptr + 11]]);
        let status = i32::from_le_bytes([data[ptr], data[ptr + 1], data[ptr + 2], data[ptr + 3]]);

        if status == 1 && result_data_ptr != 0 && result_data_ptr != -1 {
            let s = memory::read_string(memory, store, result_data_ptr)?;
            if !s.is_empty() {
                return Ok(s);
            }
            if result_data_ptr == 1 {
                return Ok("true".to_string());
            }
            if result_data_ptr == 0 {
                return Ok("0".to_string());
            }
        }
        if result_data_ptr == -1 {
            return Ok("null".to_string());
        }
    }

    memory::read_string(memory, store, result_ptr)
}
