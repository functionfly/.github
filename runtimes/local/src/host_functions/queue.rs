//! Queue host functions implementation.
//!
//! Provides per-named-queue FIFO push/pop operations for WASM guests.
//!
//! ## Security
//! - Queues are namespaced by `function_key` (`name@version`) so two
//!   different functions cannot share queues even if they pick the same name.
//! - Maximum number of messages per queue and maximum number of distinct
//!   queues are both capped by config to prevent unbounded memory growth.
//!
//! ## Host functions
//! ```text
//! queue_push(name_ptr, name_len, msg_ptr, msg_len) -> i32
//!   0   = success
//!  -1   = queue full (max_len reached)
//!  -2   = invalid queue name (memory read error)
//!  -3   = invalid message (memory read error)
//!  -4   = too many distinct queues (max_queues reached)
//!
//! queue_pop(name_ptr, name_len, out_ptr, out_len_ptr) -> i32
//!   0   = success, message written to WASM memory
//!  -1   = queue empty
//!  -2   = invalid queue name (memory read error)
//!  -3   = memory write error
//! ```

use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use tokio::sync::RwLock;
use wasmtime_wasi::p1::WasiP1Ctx;

use super::memory_utils;

/// Shared queue store: scoped_queue_key → VecDeque<message>.
///
/// The scoped key has the form `{namespace}:{queue_name}` where `namespace`
/// is the function's `name@version` key.
pub type SharedQueueStore = Arc<RwLock<HashMap<String, VecDeque<String>>>>;

pub fn new_queue_store() -> SharedQueueStore {
    Arc::new(RwLock::new(HashMap::new()))
}

/// Add `functionfly.queue_push` and `functionfly.queue_pop` to the linker.
pub fn add_queue_functions(
    store: SharedQueueStore,
    namespace: String,
    max_len: usize,
    max_queues: usize,
    linker: &mut wasmtime::Linker<WasiP1Ctx>,
) -> anyhow::Result<()> {
    // --- queue_push ---
    {
        let store_push = store.clone();
        let ns_push = namespace.clone();

        linker.func_wrap(
            "functionfly",
            "queue_push",
            move |mut caller: wasmtime::Caller<WasiP1Ctx>,
                  name_ptr: i32,
                  name_len: i32,
                  msg_ptr: i32,
                  msg_len: i32| -> i32 {
                let name =
                    match memory_utils::read_string_from_memory(&mut caller, name_ptr, name_len) {
                        Ok(n) => n,
                        Err(_) => return -2,
                    };
                let msg =
                    match memory_utils::read_string_from_memory(&mut caller, msg_ptr, msg_len) {
                        Ok(m) => m,
                        Err(_) => return -3,
                    };

                let key = format!("{}:{}", ns_push, name);

                let result: Result<(), i32> = tokio::task::block_in_place(|| {
                    tokio::runtime::Handle::current().block_on(async {
                        let mut queues = store_push.write().await;

                        // Enforce max distinct queues
                        if !queues.contains_key(&key) && queues.len() >= max_queues {
                            tracing::warn!(
                                namespace = %ns_push,
                                "queue_push: max_queues ({}) reached",
                                max_queues
                            );
                            return Err(-4);
                        }

                        let queue = queues.entry(key.clone()).or_insert_with(VecDeque::new);

                        if queue.len() >= max_len {
                            tracing::warn!(
                                namespace = %ns_push,
                                queue = %name,
                                "queue_push: queue full (max_len={})",
                                max_len
                            );
                            return Err(-1);
                        }

                        queue.push_back(msg);
                        Ok(())
                    })
                });

                match result {
                    Ok(()) => 0,
                    Err(code) => code,
                }
            },
        )?;
    }

    // --- queue_pop ---
    {
        let store_pop = store.clone();
        let ns_pop = namespace.clone();

        linker.func_wrap(
            "functionfly",
            "queue_pop",
            move |mut caller: wasmtime::Caller<WasiP1Ctx>,
                  name_ptr: i32,
                  name_len: i32,
                  out_ptr: i32,
                  out_len_ptr: i32| -> i32 {
                let name =
                    match memory_utils::read_string_from_memory(&mut caller, name_ptr, name_len) {
                        Ok(n) => n,
                        Err(_) => return -2,
                    };

                let key = format!("{}:{}", ns_pop, name);

                let msg: Option<String> = tokio::task::block_in_place(|| {
                    tokio::runtime::Handle::current()
                        .block_on(async { store_pop.write().await.get_mut(&key).and_then(|q| q.pop_front()) })
                });

                match msg {
                    None => -1,
                    Some(m) => {
                        match memory_utils::write_string_to_memory(
                            &mut caller,
                            &m,
                            out_ptr,
                            out_len_ptr,
                        ) {
                            Ok(_) => 0,
                            Err(_) => -3,
                        }
                    }
                }
            },
        )?;
    }

    tracing::debug!(
        "Added functionfly.queue_push / queue_pop host functions (namespace='{}', max_len={}, max_queues={})",
        namespace,
        max_len,
        max_queues,
    );
    Ok(())
}
