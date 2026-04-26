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
    ///
    /// Marks the runtime dirty if execution fails, ensuring it is discarded
    /// rather than returned to the pool where it could corrupt subsequent runs.
    pub fn execute_sync(&mut self, python_code: &str, input: &str) -> anyhow::Result<String> {
        match self.runtime.as_ref().expect("runtime already consumed").execute_sync(python_code, input) {
            Ok(result) => {
                self.pool.reuse_count.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
                if self.pool.reuse_count.load(std::sync::atomic::Ordering::Relaxed) > self.pool.max_reuse {
                    tracing::debug!("Python runtime exceeded max reuse count ({}/{}), marking dirty",
                        self.pool.reuse_count.load(std::sync::atomic::Ordering::Relaxed), self.pool.max_reuse);
                    self.mark_dirty();
                }
                Ok(result)
            }
            Err(e) => {
                tracing::warn!("Python execution failed, marking runtime dirty: {}", e);
                self.mark_dirty();
                Err(e)
            }
        }
    }

    /// Mark this runtime as dirty so it is discarded instead of returned.
    #[allow(dead_code)]
    pub fn mark_dirty(&mut self) {
        self.discard = true;
    }

    /// Check if this runtime is marked for discard.
    #[allow(dead_code)]
    pub fn is_dirty(&self) -> bool {
        self.discard
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
                self.pool.semaphore.add_permits(1);
            } else {
                // Discard: just release the permit
                self.pool.semaphore.add_permits(1);
            }
        }
    }
}

struct PythonRuntimePoolInner {
    idle: Mutex<VecDeque<PythonRuntime>>,
    semaphore: Semaphore,
    max_idle: usize,
    config: PythonConfig,
    max_reuse: usize,
    reuse_count: std::sync::atomic::AtomicUsize,
}

/// Bounded pool of `PythonRuntime` instances.
///
/// Clone the `Arc<PythonRuntimePool>` to share across tasks.
pub struct PythonRuntimePool {
    inner: Arc<PythonRuntimePoolInner>,
}

impl PythonRuntimePool {
    pub fn new(max_concurrent: usize, max_idle: usize, max_reuse: usize, config: PythonConfig) -> Self {
        let semaphore = Semaphore::new(max_concurrent);
        let idle = Mutex::new(VecDeque::with_capacity(max_idle));
        Self {
            inner: Arc::new(PythonRuntimePoolInner {
                idle,
                semaphore,
                max_idle,
                config,
                max_reuse,
                reuse_count: std::sync::atomic::AtomicUsize::new(0),
            }),
        }
    }

    pub async fn acquire(&self) -> anyhow::Result<PooledPythonRuntime> {
        let _permit = self.inner.semaphore.acquire().await?;

        let mut idle = self.inner.idle.lock().await;
        let runtime = if let Some(runtime) = idle.pop_front() {
            runtime
        } else {
            drop(idle);
            PythonRuntime::new(self.inner.config.clone())?
        };

        Ok(PooledPythonRuntime {
            runtime: Some(runtime),
            pool: Arc::clone(&self.inner),
            discard: false,
        })
    }

    pub async fn stats(&self) -> PoolStats {
        let idle = self.inner.idle.lock().await;
        let active_count = self.inner.semaphore.available_permits();
        PoolStats {
            max_concurrent: self.inner.semaphore.available_permits() + idle.len(),
            max_idle: self.inner.max_idle,
            idle_count: idle.len(),
            active_count: active_count.saturating_sub(self.inner.semaphore.available_permits().min(active_count)),
            reuse_count: self.inner.reuse_count.load(std::sync::atomic::Ordering::Relaxed),
            max_reuse: self.inner.max_reuse,
        }
    }

    pub fn reuse_count(&self) -> usize {
        self.inner.reuse_count.load(std::sync::atomic::Ordering::Relaxed)
    }
}

impl Clone for PythonRuntimePool {
    fn clone(&self) -> Self {
        Self {
            inner: Arc::clone(&self.inner),
        }
    }
}

/// Statistics about the pool state.
#[derive(Debug, Clone, Copy)]
pub struct PoolStats {
    pub max_concurrent: usize,
    pub max_idle: usize,
    pub idle_count: usize,
    pub active_count: usize,
    pub reuse_count: usize,
    pub max_reuse: usize,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::python::runtime::PythonConfig;

    fn default_config() -> PythonConfig {
        PythonConfig::default()
    }

    #[tokio::test]
    async fn test_pool_creation() {
        let pool = PythonRuntimePool::new(4, 2, 100, default_config());
        let stats = pool.stats().await;
        assert_eq!(stats.max_concurrent, 4);
        assert_eq!(stats.max_idle, 2);
        assert_eq!(stats.idle_count, 0);
        assert_eq!(stats.active_count, 0);
    }

    #[tokio::test]
    async fn test_pool_acquire_release() {
        let pool = PythonRuntimePool::new(2, 2, 100, default_config());

        let guard = pool.acquire().await.unwrap();
        let stats = pool.stats().await;
        assert_eq!(stats.active_count, 1);

        drop(guard);

        tokio::time::sleep(tokio::time::Duration::from_millis(10)).await;

        let stats = pool.stats().await;
        assert_eq!(stats.active_count, 0);
        assert_eq!(stats.idle_count, 1);
    }

    #[tokio::test]
    async fn test_pool_max_idle_respected() {
        let pool = PythonRuntimePool::new(4, 1, 100, default_config());

        let g1 = pool.acquire().await.unwrap();
        let g2 = pool.acquire().await.unwrap();
        drop(g1);
        drop(g2);

        tokio::time::sleep(tokio::time::Duration::from_millis(10)).await;

        let stats = pool.stats().await;
        assert_eq!(stats.idle_count, 1);
    }

    #[tokio::test]
    async fn test_pool_execute_sync() {
        let pool = PythonRuntimePool::new(2, 1, 100, default_config());
        let mut guard = pool.acquire().await.unwrap();
        let _result = guard.execute_sync("x = 1 + 1", "{}");
    }

    #[tokio::test]
    async fn test_dirty_runtime_discarded() {
        let pool = PythonRuntimePool::new(2, 1, 100, default_config());
        let mut guard = pool.acquire().await.unwrap();

        guard.mark_dirty();
        drop(guard);

        tokio::time::sleep(tokio::time::Duration::from_millis(10)).await;

        let stats = pool.stats().await;
        assert_eq!(stats.idle_count, 0);
    }

    #[tokio::test]
    async fn test_max_reuse_respected() {
        let pool = PythonRuntimePool::new(2, 2, 3, default_config());

        for i in 0..3 {
            let mut guard = pool.acquire().await.unwrap();
            let result = guard.execute_sync(&format!("x = {}", i), "{}");
            if i < 3 {
                assert!(!guard.is_dirty());
            }
            drop(guard);
            tokio::time::sleep(tokio::time::Duration::from_millis(10)).await;
        }

        let stats = pool.stats().await;
        assert_eq!(stats.reuse_count, 3);
    }
}
