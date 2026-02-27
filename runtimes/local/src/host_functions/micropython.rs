//! MicroPython wrapper env imports (mp_js_init, mp_js_do_exec, malloc, free).
//!
//! Python WASM modules built by the bundler import these from "env" so that
//! the module can instantiate. Stub implementations allow instantiation;
//! full execution would require a real MicroPython runtime or host integration.

use wasmtime_wasi::p1::WasiP1Ctx;

/// Add env.* stubs required by the MicroPython wrapper WASM module.
/// Without these, instantiation fails with "unknown import".
pub fn add_micropython_env_stubs(
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // env.mp_js_init(heap_size: i32) -> void
    linker.func_wrap(
        "env",
        "mp_js_init",
        |_caller: wasmtime::Caller<WasiP1Ctx>, _heap_size: i32| {},
    )?;

    // env.mp_js_do_exec(code_ptr: i32, code_len: i32) -> result_ptr i32
    linker.func_wrap(
        "env",
        "mp_js_do_exec",
        |_caller: wasmtime::Caller<WasiP1Ctx>, _code_ptr: i32, _code_len: i32| -> i32 { 0 },
    )?;

    // env.malloc(size: i32) -> ptr i32 (bump allocator stub: return 0 = "no memory" for now)
    linker.func_wrap(
        "env",
        "malloc",
        |_caller: wasmtime::Caller<WasiP1Ctx>, _size: i32| -> i32 { 0 },
    )?;

    // env.free(ptr: i32) -> void
    linker.func_wrap(
        "env",
        "free",
        |_caller: wasmtime::Caller<WasiP1Ctx>, _ptr: i32| {},
    )?;

    tracing::debug!("Added MicroPython env stubs (mp_js_init, mp_js_do_exec, malloc, free)");
    Ok(())
}
