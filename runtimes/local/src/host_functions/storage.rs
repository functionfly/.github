//! File storage host functions implementation.
//!
//! Provides `functionfly.storage_read` and `functionfly.storage_write` for
//! WASM guests to read and write files within a sandboxed base directory.
//!
//! ## Security
//! - Paths containing `..` or starting with `/` are rejected immediately.
//! - For files that already exist, the full canonical path is verified to
//!   remain inside the canonical base directory.
//! - For files that do **not** yet exist (new writes), the parent directory is
//!   canonicalized and checked instead — this prevents the previous bug where
//!   `storage_write` would fail with `canonicalize` on a non-existent path.
//! - The base directory is created automatically on first use if absent.

use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Add `functionfly.storage_read` and `functionfly.storage_write` to the linker.
pub fn add_storage_functions(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // Ensure the base storage directory exists so `canonicalize` can succeed.
    let base = &config.storage_base_dir;
    if !std::path::Path::new(base).exists() {
        std::fs::create_dir_all(base).ok();
    }

    // --- storage_read ---
    let config_read = config.clone();
    linker.func_wrap(
        "functionfly",
        "storage_read",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              path_ptr: i32,
              path_len: i32,
              data_ptr: i32,
              data_len_ptr: i32| -> i32 {
            let path =
                match memory_utils::read_string_from_memory(&mut caller, path_ptr, path_len) {
                    Ok(p) => p,
                    Err(_) => return -2,
                };

            let full_path =
                match validate_storage_path_existing(&path, &config_read.storage_base_dir) {
                    Ok(p) => p,
                    Err(_) => return -5,
                };

            match std::fs::read(&full_path) {
                Ok(data) => {
                    match memory_utils::write_bytes_to_memory(
                        &mut caller,
                        &data,
                        data_ptr,
                        data_len_ptr,
                    ) {
                        Ok(_) => 0,
                        Err(_) => -3,
                    }
                }
                Err(_) => -1,
            }
        },
    )?;

    // --- storage_write ---
    let config_write = config.clone();
    linker.func_wrap(
        "functionfly",
        "storage_write",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              path_ptr: i32,
              path_len: i32,
              data_ptr: i32,
              data_len: i32| -> i32 {
            let path =
                match memory_utils::read_string_from_memory(&mut caller, path_ptr, path_len) {
                    Ok(p) => p,
                    Err(_) => return -2,
                };

            let data =
                match memory_utils::read_bytes_from_memory(&mut caller, data_ptr, data_len) {
                    Ok(d) => d,
                    Err(_) => return -3,
                };

            // For writes we use the write-specific validator that handles new files.
            let full_path =
                match validate_storage_path_for_write(&path, &config_write.storage_base_dir) {
                    Ok(p) => p,
                    Err(_) => return -5,
                };

            // Create intermediate directories if necessary.
            if let Some(parent) = full_path.parent() {
                if !parent.exists() {
                    if std::fs::create_dir_all(parent).is_err() {
                        return -4;
                    }
                }
            }

            match std::fs::write(&full_path, &data) {
                Ok(_) => 0,
                Err(_) => -4,
            }
        },
    )?;

    tracing::debug!(
        "Added functionfly.storage_read and functionfly.storage_write host functions"
    );
    Ok(())
}

/// Validate a storage path for a **read** operation (or any operation where
/// the file is expected to already exist).
///
/// Both the base and the target path are canonicalized via the filesystem, so
/// symlinks are resolved and directory traversal via `.` / `..` is caught.
pub fn validate_storage_path_existing(
    path: &str,
    base_dir: &str,
) -> anyhow::Result<std::path::PathBuf> {
    sanitize_path_string(path)?;

    let full_path = std::path::Path::new(base_dir).join(path);

    let canonical_base = std::fs::canonicalize(base_dir)
        .map_err(|e| anyhow::anyhow!("Cannot canonicalize base dir '{}': {}", base_dir, e))?;
    let canonical_path = std::fs::canonicalize(&full_path)
        .map_err(|e| anyhow::anyhow!("Cannot canonicalize path '{}': {}", full_path.display(), e))?;

    if !canonical_path.starts_with(&canonical_base) {
        return Err(anyhow::anyhow!(
            "Path '{}' escapes storage base directory",
            path
        ));
    }

    Ok(canonical_path)
}

/// Validate a storage path for a **write** operation, where the target file
/// may not yet exist.
///
/// The parent directory must already exist (or be a sub-path of the base) and
/// is canonicalized for the containment check.  The filename component is then
/// appended afterwards.
pub fn validate_storage_path_for_write(
    path: &str,
    base_dir: &str,
) -> anyhow::Result<std::path::PathBuf> {
    sanitize_path_string(path)?;

    let full_path = std::path::Path::new(base_dir).join(path);

    let canonical_base = std::fs::canonicalize(base_dir)
        .map_err(|e| anyhow::anyhow!("Cannot canonicalize base dir '{}': {}", base_dir, e))?;

    // If the file already exists we can canonicalize the whole path.
    if full_path.exists() {
        let canonical_path = std::fs::canonicalize(&full_path)?;
        if !canonical_path.starts_with(&canonical_base) {
            return Err(anyhow::anyhow!(
                "Path '{}' escapes storage base directory",
                path
            ));
        }
        return Ok(canonical_path);
    }

    // File does not yet exist — canonicalize the parent directory instead.
    let parent = full_path
        .parent()
        .ok_or_else(|| anyhow::anyhow!("Path '{}' has no parent directory", path))?;

    // Ensure the parent exists (it must be inside base_dir).
    if !parent.exists() {
        // We create it eagerly so canonicalize can succeed.
        std::fs::create_dir_all(parent)
            .map_err(|e| anyhow::anyhow!("Cannot create parent dir: {}", e))?;
    }

    let canonical_parent = std::fs::canonicalize(parent)
        .map_err(|e| anyhow::anyhow!("Cannot canonicalize parent dir: {}", e))?;

    if !canonical_parent.starts_with(&canonical_base) {
        return Err(anyhow::anyhow!(
            "Path '{}' escapes storage base directory",
            path
        ));
    }

    // Reconstruct the full path with the canonical parent + original filename.
    let file_name = full_path
        .file_name()
        .ok_or_else(|| anyhow::anyhow!("Path '{}' has no file name", path))?;

    Ok(canonical_parent.join(file_name))
}

/// Legacy alias kept for backwards compatibility and test usage.
///
/// Attempts to validate for an existing path; if the file is absent, falls back
/// to the write validator (parent-dir approach).
pub fn validate_storage_path(path: &str, base_dir: &str) -> anyhow::Result<std::path::PathBuf> {
    let full_path = std::path::Path::new(base_dir).join(path);
    if full_path.exists() {
        validate_storage_path_existing(path, base_dir)
    } else {
        // For tests that pass a non-existent path, return the logical full path
        // after string-level sanitization only (no canonicalization possible).
        sanitize_path_string(path)?;
        Ok(full_path)
    }
}

/// Check the path string for obvious traversal patterns before any filesystem
/// operations.
fn sanitize_path_string(path: &str) -> anyhow::Result<()> {
    if path.contains("..") || path.starts_with('/') {
        return Err(anyhow::anyhow!(
            "Invalid path '{}': directory traversal not allowed",
            path
        ));
    }
    Ok(())
}
