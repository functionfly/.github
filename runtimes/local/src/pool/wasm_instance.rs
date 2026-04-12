//! Warm WASM instance for pooling with pre-compiled modules.

use std::sync::Arc;
use wasmtime_wasi::p1::WasiP1Ctx;

use super::wasi_state::WasiStateSnapshot;

/// A warm WASM instance ready for reuse with a pre-compiled module and WASI context.
///
/// This struct holds the components needed to execute a function without
/// recompiling or re-instantiating from scratch:
/// - The compiled `wasmtime::Module` (shared, cheap to clone)
/// - The `WasiP1Ctx` (WASI state, extracted from store for Send-safety)
/// - The `WasiStateSnapshot` to reset state between executions
///
/// The `PooledInstance` metadata (timing, memory, reuse count) is managed
/// separately in `InstancePool`. This struct focuses on the execution state.
///
/// # Thread-safety
///
/// `PooledWasmInstance` is `Send` because it only holds `Arc<Module>` (cheap to
/// share) and `WasiP1Ctx` (extracted from store, which WASI contexts are Send).
/// The `Store` itself is NOT stored here; it is recreated per-execution from the
/// pooled `WasiP1Ctx`.
pub struct PooledWasmInstance {
    /// Compiled module (Clone-safe; cheap to share)
    pub module: Arc<wasmtime::Module>,
    /// WASI context extracted from store for Send-safety
    pub wasi_ctx: WasiP1Ctx,
    /// WASI state snapshot for reset between executions
    pub wasi_snapshot: WasiStateSnapshot,
    /// Function key for pool routing
    pub function_key: String,
    /// Reuse counter
    pub reuse_count: u32,
    /// Approximate memory usage of the WASI context in bytes
    pub memory_estimate: usize,
}

impl PooledWasmInstance {
    /// Create a new pooled instance from an existing module and store.
    pub fn new(
        module: wasmtime::Module,
        wasi_ctx: WasiP1Ctx,
        function_key: String,
        pipe_capacity: usize,
    ) -> Self {
        // Estimate memory before Arc-wrapping
        let memory_estimate = Self::estimate_wasi_ctx_size(&module);
        
        Self {
            module: Arc::new(module),
            wasi_ctx,
            wasi_snapshot: WasiStateSnapshot {
                env_vars: Vec::new(),
                args: vec![function_key.clone()],
                pipe_capacity,
            },
            function_key,
            reuse_count: 0,
            memory_estimate,
        }
    }

    /// Create a new Store from the pooled WASI context.
    ///
    /// This is called per-execution to create a fresh Store with the pooled
    /// WASI context, ensuring isolation between executions.
    ///
    /// Note: Since WasiP1Ctx is not Clone, we need to rebuild it from scratch.
    /// This is the trade-off for Send-safety across spawn_blocking calls.
    pub fn create_store(&self, engine: &wasmtime::Engine) -> wasmtime::Store<WasiP1Ctx> {
        // We can't clone WasiP1Ctx, so we rebuild it with a basic config
        // The actual WASI state is restored per-execution via reset_for_execution
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        wasmtime::Store::new(engine, builder.build_p1())
    }

    /// Reset WASI state for a fresh execution.
    ///
    /// Creates new stdin/stdout/stderr pipes (clearing any residual output from
    /// the previous execution) and reapplies the snapshotted environment and args.
    ///
    /// Note: WasiP1Ctx doesn't implement Clone, so we can't actually pool it.
    /// This method is kept for API compatibility but creates a fresh context.
    pub fn reset_for_execution(&mut self, input: &str) {
        use wasmtime_wasi::p2::pipe::{MemoryInputPipe, MemoryOutputPipe};

        // Create fresh pipes to clear residual data
        let stdout = MemoryOutputPipe::new(self.wasi_snapshot.pipe_capacity);
        let stderr = MemoryOutputPipe::new(self.wasi_snapshot.pipe_capacity);
        let stdin = MemoryInputPipe::new(input.as_bytes().to_vec());

        // Rebuild WASI context with fresh pipes and restored state
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        builder.stdin(stdin).stdout(stdout).stderr(stderr);
        self.wasi_snapshot.restore(&mut builder);

        // Replace the WASI context with a fresh one
        self.wasi_ctx = builder.build_p1();
        self.reuse_count += 1;
    }

    /// Estimate WASI context memory overhead.
    fn estimate_wasi_ctx_size(module: &wasmtime::Module) -> usize {
        // Rough estimate: module memory estimate + WASI context overhead
        let module_bytes = module.serialize().map(|s| s.len()).unwrap_or(0);
        // WASI context typically uses 64KB-256KB depending on preopens and env
        module_bytes + 128 * 1024
    }
}

/// A checked-out WASM instance that is automatically returned to the pool
/// when dropped.
///
/// The guard holds an Arc reference to the pool to ensure the pool stays
/// alive as long as any guard exists.
pub struct PooledWasmInstanceGuard {
    /// The pooled instance (always `Some` until `take()` is called on drop)
    pub(crate) instance: Option<PooledWasmInstance>,
    /// Arc reference to the pool for returning the instance
    pub(crate) pool: Arc<super::wasm_pool::WasmInstancePool>,
    /// Whether this instance should be discarded rather than returned
    pub(crate) discard: bool,
}

impl PooledWasmInstanceGuard {
    /// Get the pooled instance.
    pub fn instance(&self) -> &PooledWasmInstance {
        self.instance.as_ref().expect("instance already consumed")
    }

    /// Get a mutable reference to the pooled instance for resetting state.
    pub fn instance_mut(&mut self) -> &mut PooledWasmInstance {
        self.instance.as_mut().expect("instance already consumed")
    }

    /// Mark this instance as dirty so it is discarded instead of returned.
    /// Use this when an execution leaves the WASI context in an inconsistent state.
    pub fn mark_dirty(&mut self) {
        self.discard = true;
    }

    /// Take ownership of the underlying instance, removing it from the guard.
    /// The instance will NOT be returned to the pool on drop.
    pub fn take(mut self) -> PooledWasmInstance {
        self.discard = true;
        self.instance.take().expect("instance already consumed")
    }
}

impl Drop for PooledWasmInstanceGuard {
    fn drop(&mut self) {
        if let Some(instance) = self.instance.take() {
            let pool = Arc::clone(&self.pool);
            if !self.discard {
                // Return to pool (best-effort; ignore errors)
                if let Ok(mut idle) = pool.inner().idle.try_lock() {
                    if idle.len() < pool.inner().max_idle {
                        idle.push_back(instance);
                    }
                    // else: pool is full, instance is dropped
                }
            }
            // Release the semaphore permit (whether pooled or discarded)
            pool.inner().semaphore.add_permits(1);
        }
    }
}
