//! Key-value storage host functions implementation.
//!
//! # Phase 2 — per-execution KV namespace
//!
//! All keys are automatically prefixed with `{namespace}:` where `namespace`
//! is `{tenant_id}:{function_name}`.  This prevents cross-tenant data leakage
//! when multiple functions share the same in-memory KV store.
//!
//! The namespace is passed in at linker construction time via
//! `add_kv_functions_namespaced`.  The legacy `add_kv_functions` wrapper
//! passes an empty namespace for backward compatibility.

use wasmtime_wasi::p1::WasiP1Ctx;

use crate::kv::SharedKVStore;

use super::memory_utils;

/// Add KV functions with an explicit namespace prefix.
///
/// All keys written/read by the WASM guest are transparently prefixed with
/// `{namespace}:` so that different tenants/functions cannot interfere with
/// each other's data.
///
/// Pass `namespace = ""` to disable namespacing (backward-compatible behaviour).
pub fn add_kv_functions_namespaced(
    kv_store: SharedKVStore,
    namespace: String,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    let ns_get = namespace.clone();
    let ns_set = namespace.clone();

    // functionfly.kv_get(key_ptr: i32, key_len: i32, value_ptr: i32, value_len_ptr: i32) -> i32
    // Returns 0 on success, -1 if key not found, other negative values on error
    let kv_store_get = kv_store.clone();
    linker.func_wrap(
        "functionfly",
        "kv_get",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              key_ptr: i32,
              key_len: i32,
              value_ptr: i32,
              value_len_ptr: i32| -> i32 {
            // Get the key from WASM memory
            let raw_key = match memory_utils::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2, // Invalid key
            };

            // Apply namespace prefix
            let namespaced_key = if ns_get.is_empty() {
                raw_key
            } else {
                format!("{}:{}", ns_get, raw_key)
            };

            // Get value from KV store
            let value = match tokio::task::block_in_place(|| {
                tokio::runtime::Handle::current().block_on(async {
                    let mut store = kv_store_get.write().await;
                    store.get(&namespaced_key)
                })
            }) {
                Some(v) => v,
                None => return -1, // Key not found
            };

            // Write value back to WASM memory
            match memory_utils::write_string_to_memory(&mut caller, &value, value_ptr, value_len_ptr) {
                Ok(_) => 0, // Success
                Err(_) => -3, // Memory write error
            }
        },
    )?;

    // functionfly.kv_set(key_ptr: i32, key_len: i32, value_ptr: i32, value_len: i32, ttl_seconds: i32) -> i32
    // ttl_seconds = -1 means no TTL, 0 means delete
    let kv_store_set = kv_store.clone();
    linker.func_wrap(
        "functionfly",
        "kv_set",
        move |mut caller: wasmtime::Caller<WasiP1Ctx>,
              key_ptr: i32,
              key_len: i32,
              value_ptr: i32,
              value_len: i32,
              ttl_seconds: i32| -> i32 {
            // Get the key from WASM memory
            let raw_key = match memory_utils::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2, // Invalid key
            };

            // Apply namespace prefix
            let namespaced_key = if ns_set.is_empty() {
                raw_key
            } else {
                format!("{}:{}", ns_set, raw_key)
            };

            // Get the value from WASM memory
            let value = match memory_utils::read_string_from_memory(&mut caller, value_ptr, value_len) {
                Ok(v) => v,
                Err(_) => return -3, // Invalid value
            };

            // Handle TTL
            let ttl_seconds_opt = if ttl_seconds == -1 {
                None
            } else if ttl_seconds == 0 {
                // Delete operation
                let deleted = tokio::task::block_in_place(|| {
                    tokio::runtime::Handle::current().block_on(async {
                        let mut store = kv_store_set.write().await;
                        store.delete(&namespaced_key)
                    })
                });
                return if deleted { 0 } else { -4 }; // Delete success/failure
            } else if ttl_seconds > 0 {
                Some(ttl_seconds as u64)
            } else {
                return -5; // Invalid TTL
            };

            // Set value in KV store
            let result = tokio::task::block_in_place(|| {
                tokio::runtime::Handle::current().block_on(async {
                    let mut store = kv_store_set.write().await;
                    store.set(namespaced_key, value, ttl_seconds_opt)
                })
            });

            match result {
                Ok(_) => 0, // Success
                Err(_) => -6, // Storage error
            }
        },
    )?;

    tracing::debug!(
        "Added functionfly.kv_get and functionfly.kv_set host functions (namespace='{}')",
        namespace
    );
    Ok(())
}

/// Add KV functions without namespacing (backward-compatible wrapper).
pub fn add_kv_functions(
    kv_store: SharedKVStore,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    add_kv_functions_namespaced(kv_store, String::new(), linker)
}
