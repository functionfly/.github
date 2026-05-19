//! FunctionFly WasmEdge Runtime Library
//!
//! Production-ready C/C++ and WebAssembly execution runtime with:
//! - WasmEdge WASI 0.2 support for C/C++ (via wasm32-wasi)
//! - WebAssembly sandbox isolation
//! - Resource limits (memory, CPU, time, fuel)
//! - Secure execution mode with syscall filtering
//! - Network and filesystem controls
//! - Comprehensive metrics and observability
//!
//! ## Supported Languages
//!
//! - **C/C++**: Compile with `clang --target=wasm32-wasi` or `emcc`
//! - **Rust**: Compile with `cargo build --target=wasm32-wasi`
//! - **Other WASM targets**: Any WASI 0.2 compatible binary

pub mod config;
pub mod sandbox;
pub mod security;

pub use config::{RuntimeConfig, ExecutionLimits, SecurityPolicy};
pub use sandbox::{Sandbox, SandboxConfig, SandboxResult};
pub use security::{SecurityManager, Permission, PermissionSet};

use once_cell::sync::Lazy;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

/// Initialize tracing for the runtime
pub fn init_tracing() {
    static TRACING: Lazy<()> = Lazy::new(|| {
        let env_filter = EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| EnvFilter::new("info"));

        tracing_subscriber::registry()
            .with(env_filter)
            .with(tracing_subscriber::fmt::layer().with_target(true))
            .init();
    });
    Lazy::force(&TRACING);
}
