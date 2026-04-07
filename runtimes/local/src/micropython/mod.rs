//! MicroPython WASM runtime support via module linking.
//!
//! This module provides Python function execution using dynamically linked
//! MicroPython WASM modules. It loads `micropython-full.wasm` (1.1MB) at runtime
//! and links it with generated wrapper modules using wasmtime's module linking API.
//!
//! # Architecture
//!
//! The execution flow involves two linked WASM modules:
//!
//! 1. **Wrapper Module** (~10KB) - Generated per-function, provides:
//!    - Shared memory export
//!    - Host function imports (I/O)
//!    - `malloc`/`free` for MicroPython heap
//!    - `mp_js_init`, `mp_js_do_exec` exports
//!
//! 2. **MicroPython Module** (1.1MB) - Pre-built runtime, imports:
//!    - `env.memory` - Shared linear memory
//!    - `env.mp_js_init` - Runtime initialization
//!    - `env.mp_js_do_exec` - Code execution
//!    - `env.malloc`/`env.free` - Memory management
//!
//! # Example Usage
//!
//! ```rust,ignore
//! use runtimes::local::micropython::{MicroPythonExecutor, ExecutorConfig};
//!
//! let config = ExecutorConfig::default();
//! let executor = MicroPythonExecutor::new(config).unwrap();
//!
//! let python_code = r#"
//! def handler(event):
//!     return {"message": "Hello from MicroPython!"}
//! "#;
//!
//! let result = executor.execute(python_code, r#"{"name": "test"}"#).await.unwrap();
//! ```

pub mod errors;
pub mod executor;
pub mod loader;
pub mod memory;
pub mod wrapper;

// Re-export main types for convenience
pub use executor::{ExecutorConfig, MicroPythonExecutor};

/// Version of the MicroPython runtime interface.
pub const MP_INTERFACE_VERSION: &str = "1.0.0";

/// Check if MicroPython WASM is available at the default location.
pub fn is_micropython_available() -> bool {
    let paths = [
        "assets/micropython-full.wasm",
        "internal/bundler/python/micropython-full.wasm",
        "runtimes/local/assets/micropython-full.wasm",
        "./micropython-full.wasm",
    ];

    paths.iter().any(|path| std::path::Path::new(path).exists())
}

/// Get the path to the MicroPython WASM file if it exists.
pub fn find_micropython_wasm() -> Option<String> {
    let paths = [
        "assets/micropython-full.wasm",
        "internal/bundler/python/micropython-full.wasm",
        "runtimes/local/assets/micropython-full.wasm",
        "./micropython-full.wasm",
    ];

    paths
        .iter()
        .find(|path| std::path::Path::new(path).exists())
        .map(|s| s.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_interface_version() {
        assert_eq!(MP_INTERFACE_VERSION, "1.0.0");
    }

    #[test]
    fn test_find_micropython_wasm() {
        // This will only succeed if the WASM file exists
        let path = find_micropython_wasm();
        // Just verify the function doesn't panic
        // The actual result depends on the test environment
        match path {
            Some(p) => assert!(p.contains("micropython")),
            None => {
                // Expected in test environment without WASM file
            }
        }
    }

    #[test]
    fn test_is_micropython_available() {
        // Just verify the function doesn't panic
        let _available = is_micropython_available();
    }
}
