//! Manages multiple `WasmInstancePool` objects, one per function.
//!
//! This is the main entry point for the WASM instance pooling infrastructure.
//! Each function gets its own pool with per-function concurrency limits and idle
//! instance caching.

use std::collections::HashMap;
use std::sync::Arc;

use tokio::sync::RwLock;

use super::stats::WasmPoolStats;
use super::wasm_instance::{PooledWasmInstance, PooledWasmInstanceGuard};
use super::wasm_pool::WasmInstancePool;

/// Manages multiple `WasmInstancePool` objects, one per function.
///
/// # Example
/// ```
/// let manager = PoolManager::new(10, 4); // max_concurrent=10, max_idle=4
/// manager.prewarm_instance("fn@1.0.0", module, wasi_ctx, 1024*1024).await;
/// let guard = manager.acquire("fn@1.0.0").await?;
/// ```
pub struct PoolManager {
    /// Per-function instance pools
    pools: RwLock<HashMap<String, Arc<WasmInstancePool>>>,
    /// Maximum concurrent executions per function
    max_concurrent: usize,
    /// Maximum idle instances per function
    max_idle: usize,
    /// Default pipe capacity
    default_pipe_capacity: usize,
}

impl PoolManager {
    /// Create a new pool manager.
    ///
    /// * `max_concurrent` - Maximum concurrent executions per function
    /// * `max_idle` - Maximum idle instances to keep warm per function
    pub fn new(max_concurrent: usize, max_idle: usize) -> Self {
        Self {
            pools: RwLock::new(HashMap::new()),
            max_concurrent,
            max_idle,
            default_pipe_capacity: 1024 * 1024, // 1 MiB
        }
    }

    /// Create with custom pipe capacity.
    pub fn with_pipe_capacity(mut self, capacity: usize) -> Self {
        self.default_pipe_capacity = capacity;
        self
    }

    /// Get or create a pool for the given function key.
    pub async fn get_or_create_pool(&self, function_key: &str) -> Arc<WasmInstancePool> {
        // Try read lock first (fast path)
        {
            let pools = self.pools.read().await;
            if let Some(pool) = pools.get(function_key) {
                return Arc::clone(pool);
            }
        }

        // Need to create - acquire write lock
        let mut pools = self.pools.write().await;
        // Double-check after acquiring write lock
        if let Some(pool) = pools.get(function_key) {
            return Arc::clone(pool);
        }

        let pool = Arc::new(WasmInstancePool::new(
            function_key.to_string(),
            self.max_concurrent,
            self.max_idle,
        ));
        pools.insert(function_key.to_string(), Arc::clone(&pool));
        tracing::info!(
            "Created new WasmInstancePool for {} (max_concurrent={}, max_idle={})",
            function_key,
            self.max_concurrent,
            self.max_idle
        );
        pool
    }

    /// Acquire a pooled instance for the given function.
    ///
    /// Returns a guard that automatically returns the instance to the pool on drop.
    pub async fn acquire(&self, function_key: &str) -> anyhow::Result<PooledWasmInstanceGuard> {
        let pool = self.get_or_create_pool(function_key).await;
        WasmInstancePool::acquire(pool).await
    }

    /// Pre-warm a pool with a compiled instance.
    ///
    /// Call this during startup or when you anticipate traffic to a function.
    /// If the pool doesn't exist yet, it will be created.
    pub async fn prewarm_instance(
        &self,
        function_key: &str,
        module: wasmtime::Module,
        wasi_ctx: wasmtime_wasi::p1::WasiP1Ctx,
    ) {
        let pool = self.get_or_create_pool(function_key).await;
        pool.prewarm_with(module, wasi_ctx, self.default_pipe_capacity).await;
    }

    /// Pre-warm with a PooledWasmInstance directly.
    pub async fn prewarm(&self, function_key: &str, instance: PooledWasmInstance) {
        let pool = self.get_or_create_pool(function_key).await;
        pool.prewarm(instance).await;
    }

    /// Get statistics for all pools.
    pub async fn stats(&self) -> Vec<WasmPoolStats> {
        let pools = self.pools.read().await;
        let mut stats = Vec::with_capacity(pools.len());
        for pool in pools.values() {
            stats.push(pool.stats().await);
        }
        stats
    }

    /// Get stats for a specific function's pool.
    pub async fn pool_stats(&self, function_key: &str) -> Option<WasmPoolStats> {
        let pools = self.pools.read().await;
        pools.get(function_key).map(|p| p.blocking_stats())
    }

    /// Get the total number of warmed functions.
    pub async fn warmed_function_count(&self) -> usize {
        let pools = self.pools.read().await;
        pools.values().filter(|p| p.is_warmed_blocking()).count()
    }

    /// Check if a specific function's pool has warmed instances.
    pub async fn is_warmed(&self, function_key: &str) -> bool {
        let pools = self.pools.read().await;
        if let Some(pool) = pools.get(function_key) {
            pool.is_warmed_blocking()
        } else {
            false
        }
    }

    /// Clear all pools.
    pub async fn clear(&self) {
        let mut pools = self.pools.write().await;
        pools.clear();
        tracing::info!("Cleared all WasmInstancePools");
    }

    /// Remove a specific function's pool.
    pub async fn remove_pool(&self, function_key: &str) {
        let mut pools = self.pools.write().await;
        if pools.remove(function_key).is_some() {
            tracing::info!("Removed WasmInstancePool for {}", function_key);
        }
    }
}
