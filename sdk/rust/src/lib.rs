//! FunctionFly Rust SDK
//!
//! Build serverless functions in Rust targeting WASM (wasm32-wasip1).
//!
//! # Usage
//!
//! ```rust,no_run
//! use functionfly_sdk::{Function, Context};
//!
//! struct MyFunction;
//!
//! impl Function for MyFunction {
//!     fn handle(&self, input: &str, _ctx: &Context) -> Result<String, String> {
//!         Ok(format!("{{\"message\": \"Hello from Rust!\"}}"))
//!     }
//! }
//!
//! functionfly_sdk::run!(MyFunction);
//! ```

mod context;
mod host_functions;
mod runtime;

pub use context::Context;
pub use host_functions::{log, get_env, kv_get, kv_set, fetch};

/// Trait that all FunctionFly functions must implement.
pub trait Function {
    /// Process the input and return the output.
    fn handle(&self, input: &str, ctx: &Context) -> Result<String, String>;
}

/// Register a function as the WASM entry point.
#[macro_export]
macro_rules! run {
    ($func:expr) => {
        #[no_mangle]
        pub extern "C" fn init() {
            // Initialization (if needed)
        }

        #[no_mangle]
        pub extern "C" fn execute(ptr: *const u8, len: usize) -> *const u8 {
            let input = unsafe {
                let slice = std::slice::from_raw_parts(ptr, len);
                std::str::from_utf8_unchecked(slice)
            };
            let ctx = functionfly_sdk::Context::new();
            match $func.handle(input, &ctx) {
                Ok(output) => {
                    let boxed = output.into_bytes().leak();
                    boxed.as_ptr()
                }
                Err(e) => {
                    let err_json = format!("{{\"error\": {{\"code\": \"RUNTIME_ERROR\", \"message\": \"{}\"}}}}", e);
                    let boxed = err_json.into_bytes().leak();
                    boxed.as_ptr()
                }
            }
        }

        #[no_mangle]
        pub extern "C" fn alloc(size: usize) -> *mut u8 {
            let mut buf = Vec::with_capacity(size);
            let ptr = buf.as_mut_ptr();
            std::mem::forget(buf);
            ptr
        }

        #[no_mangle]
        pub extern "C" fn dealloc(ptr: *mut u8, size: usize) {
            unsafe {
                let _ = Vec::from_raw_parts(ptr, size, size);
            }
        }
    };
}
