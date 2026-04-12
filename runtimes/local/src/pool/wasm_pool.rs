//! Bounded pool of compiled WASM instances per function.
//!
//! This pool provides true warm-instance reuse by:
//! 1. Storing pre-compiled `Module` objects (avoiding ~100ms compilation cost)
//! 2. Storing pre-initialized `WasiP1Ctx` (avoiding ~1ms WASI setup cost)
//! 3. Creating fresh `Store` and `Instance` per execution for isolation

use std::collections::VecDeque;
use std::sync::Arc;

use tokio::sync::{Mutex, Semaphore};

use super::stats::WasmPoolStats;
use super::wasm_instance::{PooledWasmInstance, PooledWasmInstanceGuard};

/// Inner pool state (thread-safe via Mutex)
pub struct WasmInstancePoolInner {
    /// Idle instances waiting to be checked out
    pub idle: Mutex<VecDeque<PooledWasmInstance>>,
    /// Limits the total number of concurrent instances (idle + active)
    pub semaphore: Semaphore,
    /// Maximum number of idle instances to keep warm in the pool
    pub max_idle: usize,
    /// Function key this pool is for
    pub function_key: String,
    /// Concurrency limit per function
    pub max_concurrent: usize,
}

/// Bounded pool of compiled WASM instances per function.
///
/// Clone the `Arc<WasmInstancePool>` to share across tasks.
///
/// # Example
/// ```
/// let pool = Arc::new(WasmInstancePool::new("fn@1.0.0", 8, 4));
/// let guard = pool.acquire().await?;
/// let mut instance = guard.instance_mut();
/// instance.reset_for_execution(&input);
/// // Use instance.module and instance.create_store() to execute
/// drop(guard); // Automatically returns to pool
/// ```
pub struct WasmInstancePool {
    inner: Arc<WasmInstancePoolInner>,
}

unsafe impl Send for WasmInstancePool {}
unsafe impl Sync for WasmInstancePool {}

impl WasmInstancePool {
    /// Create a new pool for a specific function.
    ///
    /// * `function_key` - The function this pool is for (e.g. "fn@1.0.0")
    /// * `max_concurrent` - Maximum number of concurrent executions (semaphore limit)
    /// * `max_idle` - Maximum number of idle instances to keep warm
    pub fn new(function_key: String, max_concurrent: usize, max_idle: usize) -> Self {
        let max_concurrent = max_concurrent.max(1);
        let max_idle = max_idle.min(max_concurrent);
        Self {
            inner: Arc::new(WasmInstancePoolInner {
                idle: Mutex::new(VecDeque::new()),
                semaphore: Semaphore::new(max_concurrent),
                max_idle,
                function_key,
                max_concurrent,
            }),
        }
    }

    /// Create a pool with the same limits for a new function key.
    ///
    /// Useful when creating pools for different functions while maintaining
    /// consistent per-function resource limits.
    pub fn for_function(&self, function_key: String) -> Self {
        Self::new(
            function_key,
            self.inner.max_concurrent,
            self.inner.max_idle,
        )
    }

    /// Get access to the inner state (for guard's Drop impl).
    pub(crate) fn inner(&self) -> &Arc<WasmInstancePoolInner> {
        &self.inner
    }

    /// Acquire an instance from the pool, waiting if the concurrency limit
    /// has been reached.
    ///
    /// Returns a `PooledWasmInstanceGuard` that returns the instance to the
    /// pool on drop.
    ///
    /// Note: This method takes `Arc<Self>` so the guard can hold an Arc reference
    /// to keep the pool alive.
    pub async fn acquire(pool: Arc<Self>) -> anyhow::Result<PooledWasmInstanceGuard> {
        // Wait for a permit (blocks if max_concurrent executions are active)
        let _permit = pool
            .inner
            .semaphore
            .acquire()
            .await
            .map_err(|_| anyhow::anyhow!("WasmInstancePool semaphore closed"))?;

        // Forget the permit — we manage it manually in PooledWasmInstanceGuard::drop
        _permit.forget();

        // Try to reuse an idle instance
        let instance = {
            let mut idle = pool.inner.idle.lock().await;
            idle.pop_front()
        };

        let instance = match instance {
            Some(i) => {
                tracing::debug!(
                    "WasmInstancePool[{}]: reusing idle instance (reuse_count={})",
                    pool.inner.function_key,
                    i.reuse_count
                );
                i
            }
            None => {
                tracing::debug!(
                    "WasmInstancePool[{}]: no idle instance available",
                    pool.inner.function_key
                );
                return Err(anyhow::anyhow!(
                    "No pooled instance available for {}. Pool is empty and must be pre-warmed via prewarm_instance().",
                    pool.inner.function_key
                ));
            }
        };

        Ok(PooledWasmInstanceGuard {
            instance: Some(instance),
            pool,
            discard: false,
        })
    }

    /// Pre-warm the pool by adding a compiled instance.
    ///
    /// Call this during startup or when you anticipate traffic to avoid
    /// cold-start latency on the first request.
    pub async fn prewarm(&self, instance: PooledWasmInstance) {
        if let Ok(mut idle) = self.inner.idle.try_lock() {
            if idle.len() < self.inner.max_idle {
                idle.push_back(instance);
                tracing::debug!(
                    "WasmInstancePool[{}]: pre-warmed instance",
                    self.inner.function_key
                );
            }
        }
    }

    /// Pre-warm with a module and WASI context directly.
    ///
    /// This is a convenience method that creates the `PooledWasmInstance` for you.
    pub async fn prewarm_with(
        &self,
        module: wasmtime::Module,
        wasi_ctx: wasmtime_wasi::p1::WasiP1Ctx,
        pipe_capacity: usize,
    ) {
        let instance = PooledWasmInstance::new(
            module,
            wasi_ctx,
            self.inner.function_key.clone(),
            pipe_capacity,
        );
        self.prewarm(instance).await;
    }

    /// Return pool statistics.
    pub async fn stats(&self) -> WasmPoolStats {
        let idle = self.inner.idle.lock().await;
        WasmPoolStats {
            idle_count: idle.len(),
            max_idle: self.inner.max_idle,
            available_permits: self.inner.semaphore.available_permits(),
            max_concurrent: self.inner.max_concurrent,
            function_key: self.inner.function_key.clone(),
        }
    }

    /// Check if the pool has any warm instances.
    pub async fn is_warmed(&self) -> bool {
        let idle = self.inner.idle.lock().await;
        !idle.is_empty()
    }

    /// Get the function key for this pool.
    pub fn function_key(&self) -> &str {
        &self.inner.function_key
    }
}

impl WasmInstancePool {
    /// Blocking version of stats() for use from non-async contexts.
    pub fn blocking_stats(&self) -> WasmPoolStats {
        // Use try_lock to avoid blocking; return zeros if lock is held
        match self.inner.idle.try_lock() {
            Ok(idle) => WasmPoolStats {
                idle_count: idle.len(),
                max_idle: self.inner.max_idle,
                available_permits: self.inner.semaphore.available_permits(),
                max_concurrent: self.inner.max_concurrent,
                function_key: self.inner.function_key.clone(),
            },
            Err(_) => WasmPoolStats {
                idle_count: 0,
                max_idle: self.inner.max_idle,
                available_permits: 0,
                max_concurrent: self.inner.max_concurrent,
                function_key: self.inner.function_key.clone(),
            },
        }
    }

    /// Check if warmed without awaiting.
    pub fn is_warmed_blocking(&self) -> bool {
        self.inner.idle.try_lock().is_ok_and(|idle| !idle.is_empty())
    }
}
