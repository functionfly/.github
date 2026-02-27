//! Key-value storage host functions implementation

use wasmtime_wasi::p1::WasiP1Ctx;

use crate::kv::SharedKVStore;

use super::memory_utils;

/// Add KV functions (get and set)
pub fn add_kv_functions(
    kv_store: SharedKVStore,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
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
            let key = match memory_utils::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2, // Invalid key
            };

            // Get value from KV store
            let value = match tokio::task::block_in_place(|| {
                tokio::runtime::Handle::current().block_on(async {
                    let mut store = kv_store_get.write().await;
                    store.get(&key)
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
            let key = match memory_utils::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(k) => k,
                Err(_) => return -2, // Invalid key
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
                        store.delete(&key)
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
                    store.set(key, value, ttl_seconds_opt)
                })
            });

            match result {
                Ok(_) => 0, // Success
                Err(_) => -6, // Storage error
            }
        },
    )?;

    tracing::debug!("Added functionfly.kv_get and functionfly.kv_set host functions");
    Ok(())
}