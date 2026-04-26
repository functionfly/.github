//! Wasmtime Memory Limiter
//!
//! Implements Wasmtime's `ResourceLimiter` trait to enforce memory limits
//! at the WASM engine level. This ensures memory.grow returns -1 (OOM)
//! rather than allowing unbounded allocation.
//!
//! # Security Properties
//!
//! - Memory limits are enforced by the Wasmtime engine itself
//! - OOM errors are handled gracefully without panicking the host
//! - Thread-local storage ensures limiter is available for each execution

use wasmtime::ResourceLimiter;

pub struct FunctionMemoryLimiter {
    max_bytes: usize,
    current_bytes: usize,
}

impl FunctionMemoryLimiter {
    pub fn new(memory_mb: u32) -> Self {
        let max_bytes = (memory_mb as usize) * 1024 * 1024;
        Self {
            max_bytes,
            current_bytes: 0,
        }
    }

    pub fn max_bytes(&self) -> usize {
        self.max_bytes
    }

    pub fn current_bytes(&self) -> usize {
        self.current_bytes
    }
}

impl ResourceLimiter for FunctionMemoryLimiter {
    fn memory_growing(
        &mut self,
        _current: usize,
        desired: usize,
        _maximum: Option<usize>,
    ) -> wasmtime::Result<bool> {
        let new_total = self
            .current_bytes
            .checked_add(desired)
            .unwrap_or(usize::MAX);

        if new_total > self.max_bytes {
            tracing::warn!(
                "FunctionMemoryLimiter: denied memory growth of {} bytes (current={}, max={})",
                desired,
                self.current_bytes,
                self.max_bytes
            );
            return Ok(false);
        }

        self.current_bytes = new_total;
        Ok(true)
    }

    fn table_growing(
        &mut self,
        _current: usize,
        desired: usize,
        maximum: Option<usize>,
    ) -> wasmtime::Result<bool> {
        if let Some(max) = maximum {
            if desired > max {
                tracing::warn!(
                    "FunctionMemoryLimiter: denied table growth to {} (maximum is {})",
                    desired,
                    max
                );
                return Ok(false);
            }
        }
        Ok(true)
    }
}

std::thread_local! {
    static MEMORY_LIMITER: std::cell::RefCell<Option<FunctionMemoryLimiter>> =
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
    let limiter = FunctionMemoryLimiter::new(memory_mb);
    MEMORY_LIMITER.with(|cell| {
        *cell.borrow_mut() = Some(limiter);
    });
    LimiterGuard
}

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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_memory_limiter_creation() {
        let limiter = FunctionMemoryLimiter::new(128);
        assert_eq!(limiter.max_bytes(), 128 * 1024 * 1024);
        assert_eq!(limiter.current_bytes(), 0);
    }

    #[test]
    fn test_memory_growing_allowed() {
        let mut limiter = FunctionMemoryLimiter::new(128);
        let result = limiter.memory_growing(0, 1024 * 1024, None);
        assert!(result.is_ok());
        assert!(result.unwrap());
        assert_eq!(limiter.current_bytes(), 1024 * 1024);
    }

    #[test]
    fn test_memory_growing_denied() {
        let mut limiter = FunctionMemoryLimiter::new(1);
        let result = limiter.memory_growing(0, 2 * 1024 * 1024, None);
        assert!(result.is_ok());
        assert!(!result.unwrap());
        assert_eq!(limiter.current_bytes(), 0);
    }
}
