//! File storage host functions implementation

use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Add storage functions (read and write files)
pub fn add_storage_functions(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // functionfly.storage_read(path_ptr: i32, path_len: i32, data_ptr: i32, data_len_ptr: i32) -> i32
    // Returns 0 on success, -1 if file not found, other negative values on error
    let config_read = config.clone();
    linker.func_wrap(
        "functionfly",
        "storage_read",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              path_ptr: i32,
              path_len: i32,
              data_ptr: i32,
              data_len_ptr: i32| -> i32 {
            // Get file path from WASM memory
            let path = match memory_utils::read_string_from_memory(&mut caller, path_ptr, path_len) {
                Ok(p) => p,
                Err(_) => return -2, // Invalid path
            };

            // Validate and construct full path
            let full_path = match validate_storage_path(&path, &config_read.storage_base_dir) {
                Ok(p) => p,
                Err(_) => return -5, // Invalid path
            };

            // Read file
            let result = std::fs::read(&full_path);

            match result {
                Ok(data) => {
                    // Write data back to WASM memory
                    match memory_utils::write_bytes_to_memory(&mut caller, &data, data_ptr, data_len_ptr) {
                        Ok(_) => 0, // Success
                        Err(_) => -3, // Memory write error
                    }
                }
                Err(_) => -1, // File not found or read error
            }
        },
    )?;

    // functionfly.storage_write(path_ptr: i32, path_len: i32, data_ptr: i32, data_len: i32) -> i32
    // Returns 0 on success, negative values on error
    let config_write = config.clone();
    linker.func_wrap(
        "functionfly",
        "storage_write",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              path_ptr: i32,
              path_len: i32,
              data_ptr: i32,
              data_len: i32| -> i32 {
            // Get file path from WASM memory
            let path = match memory_utils::read_string_from_memory(&mut caller, path_ptr, path_len) {
                Ok(p) => p,
                Err(_) => return -2, // Invalid path
            };

            // Get data from WASM memory
            let data = match memory_utils::read_bytes_from_memory(&mut caller, data_ptr, data_len) {
                Ok(d) => d,
                Err(_) => return -3, // Invalid data
            };

            // Validate and construct full path
            let full_path = match validate_storage_path(&path, &config_write.storage_base_dir) {
                Ok(p) => p,
                Err(_) => return -5, // Invalid path
            };

            // Write file
            let result = std::fs::write(&full_path, &data);

            match result {
                Ok(_) => 0, // Success
                Err(_) => -4, // Write error
            }
        },
    )?;

    tracing::debug!("Added functionfly.storage_read and functionfly.storage_write host functions");
    Ok(())
}

/// Validate storage path to prevent directory traversal
pub fn validate_storage_path(path: &str, base_dir: &str) -> anyhow::Result<std::path::PathBuf> {
    // Prevent directory traversal attacks
    if path.contains("..") || path.starts_with('/') {
        return Err(anyhow::anyhow!("Invalid path: directory traversal not allowed"));
    }

    // Construct full path within base directory
    let full_path = std::path::Path::new(base_dir).join(path);

    // Ensure the path is still within the base directory
    let canonical_base = std::fs::canonicalize(base_dir)?;
    let canonical_path = std::fs::canonicalize(&full_path)?;

    if !canonical_path.starts_with(&canonical_base) {
        return Err(anyhow::anyhow!("Invalid path: outside of storage directory"));
    }

    Ok(full_path)
}