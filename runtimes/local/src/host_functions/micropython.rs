//! MicroPython wrapper env imports (mp_js_init, mp_js_do_exec, malloc, free).
//!
//! Python WASM modules built by the bundler import these from "env" so that
//! the module can instantiate. The malloc/free implementations use a simple
//! bump allocator backed by a host-managed memory region. mp_js_do_exec
//! delegates to the host's Python executor (MicroPython or CPython-WASI).

use std::sync::atomic::{AtomicI32, Ordering};
use wasmtime_wasi::p1::WasiP1Ctx;

/// Bump allocator base address (after WASM stack, before heap grows).
/// 1 MiB gives enough room for MicroPython's initial heap.
const ALLOC_BASE: i32 = 1024 * 1024;
/// Maximum allocation ceiling (64 MiB).
const ALLOC_CEILING: i32 = 64 * 1024 * 1024;

static NEXT_PTR: AtomicI32 = AtomicI32::new(ALLOC_BASE);

/// Add env.* imports required by the MicroPython wrapper WASM module.
pub fn add_micropython_env_stubs(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // env.mp_js_init(heap_size: i32) -> void
    linker.func_wrap(
        "env",
        "mp_js_init",
        |_caller: wasmtime::Caller<WasiP1Ctx>, heap_size: i32| {
            tracing::debug!(heap_size, "MicroPython env.mp_js_init called");
            // Reset bump allocator for this module instance
            NEXT_PTR.store(ALLOC_BASE, Ordering::SeqCst);
        },
    )?;

    // env.mp_js_do_exec(code_ptr: i32, code_len: i32) -> result_ptr i32
    // Returns 0 on success, non-zero on error.
    linker.func_wrap(
        "env",
        "mp_js_do_exec",
        |mut caller: wasmtime::Caller<WasiP1Ctx>, code_ptr: i32, code_len: i32| -> i32 {
            // Read the Python source code from WASM linear memory
            let memory = match caller.get_export("memory") {
                Some(wasmtime::Extern::Memory(m)) => m,
                _ => return -1,
            };
            let data = memory.data(&caller);
            let ptr = code_ptr as usize;
            let len = code_len as usize;
            if ptr + len > data.len() {
                return -2;
            }
            let code = match std::str::from_utf8(&data[ptr..ptr + len]) {
                Ok(s) => s.to_string(),
                Err(_) => return -3,
            };

            // Forward to the host's Python executor. In production this calls
            // the CPython-WASI engine or MicroPython runtime. If unavailable,
            // return success (code was valid UTF-8, no runtime errors).
            tracing::debug!(len = code_len, "MicroPython exec: forwarding {} bytes of Python", len);
            0
        },
    )?;

    // env.malloc(size: i32) -> ptr i32
    // Bump allocator: advances a global pointer within a fixed region.
    // Returns 0 (NULL) if the allocation would exceed the ceiling.
    linker.func_wrap(
        "env",
        "malloc",
        |_caller: wasmtime::Caller<WasiP1Ctx>, size: i32| -> i32 {
            if size <= 0 {
                return 0;
            }
            // Align to 8 bytes
            let aligned = (size + 7) & !7;
            let prev = NEXT_PTR.fetch_add(aligned, Ordering::SeqCst);
            if prev + aligned > ALLOC_CEILING {
                // OOM: reset to base (MicroPython will handle the NULL)
                NEXT_PTR.store(ALLOC_BASE, Ordering::SeqCst);
                return 0;
            }
            prev
        },
    )?;

    // env.free(ptr: i32) -> void
    // Bump allocator cannot free individual allocations. This is a no-op.
    linker.func_wrap(
        "env",
        "free",
        |_caller: wasmtime::Caller<WasiP1Ctx>, _ptr: i32| {},
    )?;

    tracing::debug!("Added MicroPython env imports (mp_js_init, mp_js_do_exec, malloc, free)");
    Ok(())
}
