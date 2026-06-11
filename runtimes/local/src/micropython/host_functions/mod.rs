//! Host functions for MicroPython WASM runtime.
//!
//! This module provides the implementation of all env.* imports that
//! MicroPython.wasm requires. These are organized by functional area:
//!
//! - `exec.rs` - Core execution: mp_js_init, mp_js_do_exec
//! - `memory.rs` - Memory allocation: malloc, free
//! - `js_interop.rs` - JavaScript interop: invoke_*, mp_js_*, proxy_*, call*
//! - `syscalls.rs` - Syscall stubs: __syscall_*
//! - `streaming.rs` - Chunked I/O: streaming_*
//! - `wasi.rs` - WASI imports: wasi_snapshot_preview1.fd_*
//! - `emscripten.rs` - Emscripten compatibility: emscripten_*
//! - `bridge.rs` - FunctionFly Python Bridge: ff_*

pub mod exec;
pub mod memory;
pub mod js_interop;
pub mod syscalls;
pub mod streaming;
pub mod wasi;
pub mod emscripten;
pub mod bridge;

use super::errors::MicroPythonError;
use super::memory::HostState;
use wasmtime::{Linker, Store};

/// Register all MicroPython host functions with the linker.
pub fn register_all_host_functions(
    linker: &mut Linker<HostState>,
    store: &mut Store<HostState>,
) -> Result<(), MicroPythonError> {
    exec::register(linker, store)?;
    memory::register(linker, store)?;
    js_interop::register(linker, store)?;
    syscalls::register(linker, store)?;
    streaming::register(linker, store)?;
    wasi::register(linker, store)?;
    emscripten::register(linker, store)?;
    bridge::register(linker, store)?;

    tracing::debug!("Registered all MicroPython host functions");
    Ok(())
}