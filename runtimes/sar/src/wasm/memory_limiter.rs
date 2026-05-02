use wasmtime::ResourceLimiter;

pub struct SandboxMemoryLimiter {
    max_bytes: usize,
}

impl SandboxMemoryLimiter {
    pub fn new(memory_mb: u32) -> Self {
        Self {
            max_bytes: (memory_mb as usize) * 1024 * 1024,
        }
    }
}

impl ResourceLimiter for SandboxMemoryLimiter {
    fn memory_growing(
        &mut self,
        _current: usize,
        desired: usize,
        _maximum: Option<usize>,
    ) -> wasmtime::Result<bool> {
        if desired > self.max_bytes {
            tracing::warn!(
                desired_bytes = desired,
                limit_bytes = self.max_bytes,
                "SandboxMemoryLimiter: denied memory growth"
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

std::thread_local! {
    static MEMORY_LIMITER: std::cell::RefCell<Option<SandboxMemoryLimiter>> =
        const { std::cell::RefCell::new(None) };
}

pub struct LimiterGuard;

impl Drop for LimiterGuard {
    fn drop(&mut self) {
        MEMORY_LIMITER.with(|cell| {
            *cell.borrow_mut() = None;
        });
    }
}

pub fn install_memory_limiter(memory_mb: u32) -> LimiterGuard {
    let limiter = SandboxMemoryLimiter::new(memory_mb);
    MEMORY_LIMITER.with(|cell| {
        *cell.borrow_mut() = Some(limiter);
    });
    LimiterGuard
}

/// # Safety
///
/// Caller must ensure a `LimiterGuard` is alive (limiter installed). The
/// returned reference has a `'static` lifetime due to wasmtime's API
/// requirements but is actually backed by thread-local storage scoped to the
/// guard's lifetime.
pub unsafe fn with_limiter<F, R>(f: F) -> R
where
    F: FnOnce(&'static mut dyn ResourceLimiter) -> R,
{
    MEMORY_LIMITER.with(|cell| {
        let ptr = cell.as_ptr();
        let limiter: &'static mut dyn ResourceLimiter = unsafe { &mut *ptr }
            .as_mut()
            .map(|l| -> &mut dyn ResourceLimiter { l })
            .expect("MEMORY_LIMITER must be set before Store is used");
        f(limiter)
    })
}
