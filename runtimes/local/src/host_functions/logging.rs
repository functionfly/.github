//! Logging host function implementation

use wasmtime_wasi::p1::WasiP1Ctx;

use crate::logging::StructuredLogger;

use super::memory_utils;

/// Add the functionfly.log function for structured logging
pub fn add_log_function(
    logger: StructuredLogger,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // functionfly.log(level_ptr: i32, level_len: i32, message_ptr: i32, message_len: i32) -> i32
    // Returns 0 on success, negative values on error
    linker.func_wrap(
        "functionfly",
        "log",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              level_ptr: i32,
              level_len: i32,
              message_ptr: i32,
              message_len: i32| -> i32 {
            // Get the log level from WASM memory
            let level = match memory_utils::read_string_from_memory(&mut caller, level_ptr, level_len) {
                Ok(l) => l,
                Err(_) => return -1, // Invalid level
            };

            // Get the message from WASM memory
            let message = match memory_utils::read_string_from_memory(&mut caller, message_ptr, message_len) {
                Ok(m) => m,
                Err(_) => return -2, // Invalid message
            };

            // Convert log level string to enum
            let log_level = match level.to_lowercase().as_str() {
                "debug" => crate::logging::LogLevel::Debug,
                "info" => crate::logging::LogLevel::Info,
                "warn" => crate::logging::LogLevel::Warn,
                "error" => crate::logging::LogLevel::Error,
                _ => return -3, // Invalid log level
            };

            // Log the message (blocking operation in async context)
            let logger_clone = logger.clone();
            let result: Result<(), anyhow::Error> = tokio::task::block_in_place(|| {
                tokio::runtime::Handle::current().block_on(async {
                    let correlation_id = logger_clone.generate_correlation_id().await;
                    logger_clone.log_with_correlation(log_level, message, &correlation_id);
                    Ok(())
                })
            });

            match result {
                Ok(_) => 0, // Success
                Err(_) => -4, // Logging error
            }
        },
    )?;

    tracing::debug!("Added functionfly.log host function");
    Ok(())
}