//! Environment variables host function implementation

use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Add the functionfly.get_env function for environment variables
pub fn add_get_env_function(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // functionfly.get_env(key_ptr: i32, key_len: i32, value_ptr: i32, value_len_ptr: i32) -> i32
    // Returns 0 on success, -1 if key not found, other negative values on error
    linker.func_wrap(
        "functionfly",
        "get_env",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              key_ptr: i32,
              key_len: i32,
              value_ptr: i32,
              value_len_ptr: i32| -> i32 {
            // Get the key from WASM memory
            let key = match memory_utils::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2, // Invalid key
            };

            // Get environment variable value
            let value = if let Ok(v) = std::env::var(&key) {
                v
            } else {
                // Check config environment variables as fallback
                match config.wasi_env.iter()
                    .find(|env_var| env_var.starts_with(&format!("{}=", key)))
                    .and_then(|env_var| env_var.split_once('=').map(|(_, v)| v.to_string())) {
                    Some(v) => v,
                    None => return -1, // Key not found
                }
            };

            // Write value back to WASM memory
            match memory_utils::write_string_to_memory(&mut caller, &value, value_ptr, value_len_ptr) {
                Ok(_) => 0, // Success
                Err(_) => -3, // Memory write error
            }
        },
    )?;

    tracing::debug!("Added functionfly.get_env host function");
    Ok(())
}