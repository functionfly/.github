//! Python runtime pool for reusing `PythonRuntime` objects across requests.
//!
//! Creating a new `PythonRuntime` (RustPython interpreter) for every request
//! is wasteful because interpreter initialisation is not free.  This module
//! provides a bounded pool of pre-initialised runtimes that are checked out
//! for the duration of a single execution and then returned.
//!
//! # Phase 3 implementation
//!
//! Addresses the gap identified in `plans/SANDBOX_EXECUTION_LAYER.md`:
//! > Python execution via RustPython has no warm instance reuse — **Medium**
//! > The pool only tracks WASM instances; Python `PythonRuntime` objects are
//! > created fresh per request — add a Python runtime pool.
//!
//! # Design
//!
//! * The pool is backed by a `tokio::sync::Semaphore` to bound concurrency and
//!   a `Mutex<VecDeque<PythonRuntime>>` to hold idle runtimes.
//! * When the pool is empty a new runtime is created on demand (up to the
//!   semaphore limit).
//! * Runtimes are returned to the pool after execution unless the pool is full
//!   or the runtime is considered "dirty" (e.g. it raised an unhandled
//!   exception that may have left the interpreter in a bad state).

use std::collections::VecDeque;
use std::sync::Arc;
use tokio::sync::{Mutex, Semaphore};

use crate::python::runtime::{PythonConfig, PythonRuntime};

/// A checked-out Python runtime that is automatically returned to the pool
/// when dropped.
pub struct PooledPythonRuntime {
    /// The actual runtime (always `Some` until `take()` is called on drop).
    runtime: Option<PythonRuntime>,
    /// Reference back to the pool for returning the runtime.
    pool: Arc<PythonRuntimePoolInner>,
    /// Whether this runtime should be discarded rather than returned.
    discard: bool,
}

impl PooledPythonRuntime {
    /// Execute Python code using this pooled runtime.
    pub fn execute_sync(&self, python_code: &str, input: &str) -> anyhow::Result<String> {
        self.runtime
            .as_ref()
            .expect("runtime already consumed")
            .execute_sync(python_code, input)
    }

    /// Mark this runtime as dirty so it is discarded instead of returned.
    pub fn mark_dirty(&mut self) {
        self.discard = true;
    }
}

impl Drop for PooledPythonRuntime {
    fn drop(&mut self) {
        if let Some(runtime) = self.runtime.take() {
            if !self.discard {
                // Return to pool (best-effort; ignore errors)
                let pool = Arc::clone(&self.pool);
                // We are potentially in a sync context so use try_lock
                if let Ok(mut idle) = pool.idle.try_lock() {
                    if idle.len() < pool.max_idle {
                        idle.push_back(runtime);
                    }
                    // else: pool is full, runtime is dropped
                }
                // Release the semaphore permit
                pool.semaphore.add_permits(1);
            } else {
                // Discard: just release the permit
                pool.semaphore.add_permits(1);
            }
        }
    }
}

struct PythonRuntimePoolInner {
    /// Idle runtimes waiting to be checked out.
    idle: Mutex<VecDeque<PythonRuntime>>,
    /// Limits the total number of concurrent runtimes (idle + active).
    semaphore: Semaphore,
    /// Maximum number of idle runtimes to keep.
    max_idle: usize,
    /// Configuration used to create new runtimes.
    config: PythonConfig,
}

/// Bounded pool of `PythonRuntime` instances.
///
/// Clone the `Arc<PythonRuntimePool>` to share across tasks.
pub struct PythonRuntimePool {
    inner: Arc<PythonRuntimePoolInner>,
}

impl PythonRuntimePool {
    /// Create a new pool.
    ///
    /// * `max_concurrent` — maximum number of runtimes that can be active at
    ///   the same time (semaphore limit).
    /// * `max_idle` — maximum number of runtimes to keep warm in the pool.
    /// * `config` — configuration used when creating new runtimes.
    pub fn new(max_concurrent: usize, max_idle: usize, config: PythonConfig) -> Self {
        let max_concurrent = max_concurrent.max(1);
        let max_idle = max_idle.min(max_concurrent);
        Self {
            inner: Arc::new(PythonRuntimePoolInner {
                idle: Mutex::new(VecDeque::new()),
                semaphore: Semaphore::new(max_concurrent),
                max_idle,
                config,
            }),
        }
    }

    /// Acquire a runtime from the pool, waiting if the concurrency limit has
    /// been reached.
    ///
    /// Returns a `PooledPythonRuntime` guard that returns the runtime to the
    /// pool on drop.
    pub async fn acquire(&self) -> anyhow::Result<PooledPythonRuntime> {
        // Wait for a permit (blocks if max_concurrent runtimes are active)
        let _permit = self
            .inner
            .semaphore
            .acquire()
            .await
            .map_err(|_| anyhow::anyhow!("PythonRuntimePool semaphore closed"))?;

        // Forget the permit — we manage it manually in PooledPythonRuntime::drop
        _permit.forget();

        // Try to reuse an idle runtime
        let runtime = {
            let mut idle = self.inner.idle.lock().await;
            idle.pop_front()
        };

        let runtime = match runtime {
            Some(r) => {
                tracing::debug!("PythonRuntimePool: reusing idle runtime");
                r
            }
            None => {
                tracing::debug!("PythonRuntimePool: creating new runtime");
                PythonRuntime::new(self.inner.config.clone())
                    .map_err(|e| anyhow::anyhow!("Failed to create PythonRuntime: {}", e))?
            }
        };

        Ok(PooledPythonRuntime {
            runtime: Some(runtime),
            pool: Arc::clone(&self.inner),
            discard: false,
        })
    }

    /// Return pool statistics.
    pub async fn stats(&self) -> PythonPoolStats {
        let idle = self.inner.idle.lock().await;
        PythonPoolStats {
            idle_count: idle.len(),
            max_idle: self.inner.max_idle,
            available_permits: self.inner.semaphore.available_permits(),
        }
    }
}

/// Statistics about the Python runtime pool.
#[derive(Debug, Clone)]
pub struct PythonPoolStats {
    pub idle_count: usize,
    pub max_idle: usize,
    pub available_permits: usize,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn default_config() -> PythonConfig {
        PythonConfig::default()
    }

    #[tokio::test]
    async fn test_pool_acquire_and_return() {
        let pool = PythonRuntimePool::new(4, 2, default_config());

        {
            let guard = pool.acquire().await.unwrap();
            let stats = pool.stats().await;
            // One permit consumed
            assert_eq!(stats.available_permits, 3);
            drop(guard);
        }

        // After drop, permit should be returned and runtime should be idle
        let stats = pool.stats().await;
        assert_eq!(stats.available_permits, 4);
        assert_eq!(stats.idle_count, 1);
    }

    #[tokio::test]
    async fn test_pool_discard() {
        let pool = PythonRuntimePool::new(4, 2, default_config());

        {
            let mut guard = pool.acquire().await.unwrap();
            guard.mark_dirty();
            drop(guard);
        }

        // Dirty runtime should not be returned to pool
        let stats = pool.stats().await;
        assert_eq!(stats.idle_count, 0);
        assert_eq!(stats.available_permits, 4);
    }

    #[tokio::test]
    async fn test_pool_max_idle_respected() {
        let pool = PythonRuntimePool::new(4, 1, default_config()); // max_idle = 1

        // Acquire and release two runtimes
        let g1 = pool.acquire().await.unwrap();
        let g2 = pool.acquire().await.unwrap();
        drop(g1);
        drop(g2);

        // Only 1 should be kept idle
        let stats = pool.stats().await;
        assert_eq!(stats.idle_count, 1);
    }

    #[tokio::test]
    async fn test_pool_execute_sync() {
        let pool = PythonRuntimePool::new(2, 1, default_config());
        let guard = pool.acquire().await.unwrap();
        // Simple Python that returns a value
        let result = guard.execute_sync("result = 1 + 1", "{}");
        // RustPython returns "None" for exec-mode code without explicit result
        assert!(result.is_ok());
    }
}
