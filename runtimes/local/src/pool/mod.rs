//! Instance pool for reusing warm Wasm instances with memory optimization.
//!
//! This module provides two pools:
//!
//! 1. **`InstancePool`** (legacy_pool) - Metadata-only tracking for memory pressure, LRU eviction,
//!    and monitoring. Does not store actual WASM instances.
//!
//! 2. **`WasmInstancePool`** (wasm_pool) - True warm-instance reuse that pools compiled
//!    `Module` objects and WASI contexts across executions, significantly
//!    reducing per-execution overhead (module compilation is ~100ms, WASI context
//!    creation is ~1ms).
//!
//! The **`PoolManager`** (manager) is the main entry point that manages multiple
//! `WasmInstancePool` objects, one per function.

mod instance;
mod legacy_pool;
mod manager;
mod stats;
mod wasm_instance;
mod wasm_pool;
mod wasi_state;

// Re-export the main public types
pub use instance::PooledInstance;
pub use legacy_pool::InstancePool;
pub use manager::PoolManager;
pub use stats::{PoolStats, WasmPoolStats};
pub use wasm_instance::{PooledWasmInstance, PooledWasmInstanceGuard};
pub use wasm_pool::WasmInstancePool;
pub use wasi_state::WasiStateSnapshot;
