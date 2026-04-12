//! Memory limiter for WASM execution.

use wasmtime::ResourceLimiter;

/// Wasmtime `ResourceLimiter` that caps the linear memory a Wasm instance
/// can allocate.  Exceeding the limit causes the memory.grow instruction to
/// return -1 (out-of-memory) rather than panicking the host.
///
/// # Thread-safety note
///
/// Wasmtime's `store.limiter()` closure must return `&mut dyn ResourceLimiter`
/// from the store data `T`.  Because our store data is `WasiP1Ctx` (which we
/// cannot modify), we use a thread-local to hold the limiter for the duration
/// of each synchronous execution call.  This is safe because:
///
/// 1. `execute_wasi_sync_inner` is always called from `spawn_blocking`, which
///    runs on a dedicated OS thread.
/// 2. The thread-local is set before the store is used and cleared after.
/// 3. Wasmtime only calls the limiter closure while the store is alive.
pub struct FunctionMemoryLimiter {
    max_bytes: usize,
}

impl FunctionMemoryLimiter {
    /// Create a new memory limiter with the specified limit in megabytes.
    pub fn new(memory_mb: u32) -> Self {
        let max_bytes = (memory_mb as usize) * 1024 * 1024;
        Self { max_bytes }
    }
}

impl ResourceLimiter for FunctionMemoryLimiter {
    fn memory_growing(
        &mut self,
        _current: usize,
        desired: usize,
        _maximum: Option<usize>,
    ) -> wasmtime::Result<bool> {
        if desired > self.max_bytes {
            tracing::warn!(
                "FunctionMemoryLimiter: denied memory growth to {} bytes (limit {} bytes)",
                desired,
                self.max_bytes
            );
            return Ok(false);
        }
        Ok(true)
    }

    fn table_growing(
        &mut self,
        _current: usize,
        _desired: usize,
        _maximum: Option<usize>,
    ) -> wasmtime::Result<bool> {
        Ok(true)
    }
}

// Thread-local storage for the per-execution memory limiter.
//
// Because `Store<WasiP1Ctx>` uses an opaque data type that we cannot extend,
// we store the limiter in a thread-local and hand out `&mut` references to it
// through `store.limiter()`.  Each `spawn_blocking` task gets its own OS
// thread, so this is safe for concurrent execution.
std::thread_local! {
    static MEMORY_LIMITER: std::cell::RefCell<Option<FunctionMemoryLimiter>> =
        const { std::cell::RefCell::new(None) };
}

/// Guard that clears the thread-local limiter on drop.
pub struct LimiterGuard;

impl Drop for LimiterGuard {
    fn drop(&mut self) {
        MEMORY_LIMITER.with(|cell| {
            *cell.borrow_mut() = None;
        });
    }
}

/// Install the memory limiter for the current thread and return a guard that
/// clears it when the execution is done.
pub fn install_memory_limiter(memory_mb: u32) -> LimiterGuard {
    let limiter = FunctionMemoryLimiter::new(memory_mb);
    MEMORY_LIMITER.with(|cell| {
        *cell.borrow_mut() = Some(limiter);
    });
    LimiterGuard
}

/// Access the thread-local memory limiter.
///
/// # Safety
/// This should only be called while a limiter guard is active (i.e., between
/// `install_memory_limiter` and when the guard is dropped). The returned reference
/// has a 'static lifetime due to Wasmtime's API requirements, but it's actually
/// backed by thread-local storage that lives for the duration of the task.
pub unsafe fn with_limiter<F, R>(f: F) -> R
where
    F: FnOnce(&'static mut dyn ResourceLimiter) -> R,
{
    MEMORY_LIMITER.with(|cell| {
        let ptr = cell.as_ptr();
        // Safety: the limiter is always Some while the guard is alive
        let limiter: &'static mut dyn ResourceLimiter = unsafe { &mut *ptr }
            .as_mut()
            .map(|l| -> &mut dyn ResourceLimiter { l })
            .expect("MEMORY_LIMITER must be set before Store is used");
        f(limiter)
    })
}
