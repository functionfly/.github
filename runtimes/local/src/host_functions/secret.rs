//! Secrets host function implementation.
//!
//! Provides `functionfly.get_secret` — a scoped, read-only secrets API for WASM guests.
//!
//! Secret resolution order (later entries override earlier ones):
//!   1. Optional JSON secrets file (path from `--secrets-file` config flag)
//!   2. Environment variables with the `SECRET_` prefix (prefix is stripped)
//!
//! Every access is audited via tracing at INFO level so access is always visible
//! in structured logs without exposing the value itself.
//!
//! The function namespace (function_name@version) is included in every log record
//! to enable per-function audit queries.

use std::collections::HashMap;
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::config::Config;

use super::memory_utils;

/// Add `functionfly.get_secret` to the linker.
///
/// Signature:
/// ```text
/// get_secret(key_ptr: i32, key_len: i32, value_ptr: i32, value_len_ptr: i32) -> i32
/// ```
/// Returns:
///   `0`  — success, value written to WASM memory
///   `-1` — key not found
///   `-2` — invalid key (memory read error)
///   `-3` — memory write error
pub fn add_get_secret_function(
    config: Config,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    let secrets = build_secret_map(&config);
    let function_key = config.function_key();

    linker.func_wrap(
        "functionfly",
        "get_secret",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              key_ptr: i32,
              key_len: i32,
              value_ptr: i32,
              value_len_ptr: i32| -> i32 {
            let key = match memory_utils::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2,
            };

            // Audit every access — value is never logged
            tracing::info!(
                function = %function_key,
                secret_key = %key,
                "functionfly.get_secret accessed"
            );

            match secrets.get(&key) {
                None => {
                    tracing::warn!(
                        function = %function_key,
                        key = %key,
                        "functionfly.get_secret: key not found"
                    );
                    -1
                }
                Some(value) => {
                    match memory_utils::write_string_to_memory(
                        &mut caller,
                        value,
                        value_ptr,
                        value_len_ptr,
                    ) {
                        Ok(_) => 0,
                        Err(_) => -3,
                    }
                }
            }
        },
    )?;

    tracing::debug!("Added functionfly.get_secret host function");
    Ok(())
}

/// Build the complete secrets map for a function instance.
///
/// Resolution order: secrets file first, then `SECRET_*` env vars (override).
pub fn build_secret_map(config: &Config) -> HashMap<String, String> {
    let mut secrets = HashMap::new();

    if let Some(path) = &config.secrets_file {
        match load_secrets_file(path) {
            Ok(file_secrets) => {
                secrets.extend(file_secrets);
                tracing::debug!("Loaded {} secrets from file '{}'", secrets.len(), path);
            }
            Err(e) => {
                tracing::warn!("Failed to load secrets file '{}': {}", path, e);
            }
        }
    }

    // Environment variables with SECRET_ prefix override file entries
    let mut env_count = 0usize;
    for (key, value) in std::env::vars() {
        if let Some(secret_key) = key.strip_prefix("SECRET_") {
            secrets.insert(secret_key.to_string(), value);
            env_count += 1;
        }
    }
    if env_count > 0 {
        tracing::debug!("Loaded {} secrets from SECRET_* environment variables", env_count);
    }

    secrets
}

fn load_secrets_file(path: &str) -> anyhow::Result<HashMap<String, String>> {
    let content = std::fs::read_to_string(path)
        .map_err(|e| anyhow::anyhow!("Cannot read secrets file '{}': {}", path, e))?;
    let map: HashMap<String, String> = serde_json::from_str(&content)
        .map_err(|e| anyhow::anyhow!("Invalid JSON in secrets file '{}': {}", path, e))?;
    Ok(map)
}
