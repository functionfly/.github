//! Wasmtime `ResourceLimiter` implementation for hard-capping linear memory growth.
//!
//! This module implements the `wasmtime::ResourceLimiter` trait so that each
//! `Store` can enforce a per-function memory ceiling at the WebAssembly level,
//! not just at the monitoring layer.  When a guest tries to grow its linear
//! memory beyond the configured limit the growth is denied and the guest
//! receives a `MemoryOutOfBounds` trap rather than silently consuming host RAM.
//!
//! # Phase 1 implementation
//!
//! Addresses the gap identified in `plans/SANDBOX_EXECUTION_LAYER.md`:
//! > Memory limit is declared but not enforced at the OS level for Tier 1
//! > **High** — Use Wasmtime's `Store::limiter()` API with a `ResourceLimiter`
//! > implementation that hard-caps linear memory growth.

use wasmtime::ResourceLimiter;

/// Per-execution resource limiter that hard-caps WebAssembly linear memory.
///
/// Attach this to a `Store` via `store.limiter(|state| &mut state.limiter)` or
/// by storing it alongside the WASI context and calling
/// `store.limiter(|ctx| &mut ctx.resource_limiter)`.
///
/// Because `Store::limiter` requires the limiter to be embedded in the store
/// data, `WasmResourceLimiter` is designed to be stored inside the store's
/// data type (e.g. alongside `WasiP1Ctx`).
#[derive(Debug, Clone)]
pub struct WasmResourceLimiter {
    /// Maximum number of WebAssembly memory pages (1 page = 64 KiB).
    max_memory_pages: u64,
    /// Maximum number of table elements.
    max_table_elements: u32,
    /// Current number of allocated memory pages (tracked for reporting).
    current_pages: u64,
    /// Whether to log memory growth attempts.
    verbose: bool,
}

impl WasmResourceLimiter {
    /// Create a new limiter from a memory limit expressed in **megabytes**.
    ///
    /// The limit is converted to WebAssembly pages (64 KiB each).
    pub fn new(memory_limit_mb: u32) -> Self {
        // 1 MiB = 16 pages (each page is 64 KiB)
        let max_memory_pages = (memory_limit_mb as u64 * 1024 * 1024) / 65536;
        Self {
            max_memory_pages,
            max_table_elements: 10_000,
            current_pages: 0,
            verbose: false,
        }
    }

    /// Enable verbose logging of memory growth attempts.
    pub fn with_verbose(mut self, verbose: bool) -> Self {
        self.verbose = verbose;
        self
    }

    /// Return the current memory usage in bytes.
    pub fn current_memory_bytes(&self) -> u64 {
        self.current_pages * 65536
    }

    /// Return the memory limit in bytes.
    pub fn memory_limit_bytes(&self) -> u64 {
        self.max_memory_pages * 65536
    }
}

impl ResourceLimiter for WasmResourceLimiter {
    /// Called before a WebAssembly linear memory is grown.
    ///
    /// Returns `Ok(true)` to allow the growth, `Ok(false)` to deny it (the
    /// guest will receive a `MemoryOutOfBounds` trap), or `Err(e)` to
    /// propagate a host error.
    fn memory_growing(
        &mut self,
        current: usize,
        desired: usize,
        maximum: Option<usize>,
    ) -> anyhow::Result<bool> {
        let desired_pages = (desired as u64 + 65535) / 65536; // round up to pages

        if desired_pages > self.max_memory_pages {
            if self.verbose {
                tracing::warn!(
                    current_bytes = current,
                    desired_bytes = desired,
                    limit_bytes = self.memory_limit_bytes(),
                    "WasmResourceLimiter: memory growth denied (would exceed limit)"
                );
            } else {
                tracing::debug!(
                    "WasmResourceLimiter: memory growth denied ({} -> {} bytes, limit {} bytes)",
                    current,
                    desired,
                    self.memory_limit_bytes()
                );
            }
            return Ok(false);
        }

        // Check against the module-declared maximum if present
        if let Some(max) = maximum {
            let max_pages = (max as u64 + 65535) / 65536;
            if max_pages > self.max_memory_pages {
                tracing::debug!(
                    "WasmResourceLimiter: module maximum ({} pages) exceeds host limit ({} pages)",
                    max_pages,
                    self.max_memory_pages
                );
                // Still allow growth up to our limit; the module max is advisory
            }
        }

        self.current_pages = desired_pages;
        Ok(true)
    }

    /// Called before a WebAssembly table is grown.
    fn table_growing(
        &mut self,
        current: usize,
        desired: usize,
        maximum: Option<usize>,
    ) -> anyhow::Result<bool> {
        if desired > self.max_table_elements as usize {
            tracing::debug!(
                "WasmResourceLimiter: table growth denied ({} -> {} elements, limit {})",
                current,
                desired,
                self.max_table_elements
            );
            return Ok(false);
        }
        let _ = maximum; // advisory only
        Ok(true)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_limiter_allows_within_limit() {
        let mut limiter = WasmResourceLimiter::new(128); // 128 MB
        // 64 MB growth should be allowed
        let result = limiter.memory_growing(0, 64 * 1024 * 1024, None);
        assert!(result.is_ok());
        assert!(result.unwrap());
    }

    #[test]
    fn test_limiter_denies_over_limit() {
        let mut limiter = WasmResourceLimiter::new(64); // 64 MB
        // 128 MB growth should be denied
        let result = limiter.memory_growing(0, 128 * 1024 * 1024, None);
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    #[test]
    fn test_limiter_page_conversion() {
        let limiter = WasmResourceLimiter::new(64); // 64 MB
        assert_eq!(limiter.memory_limit_bytes(), 64 * 1024 * 1024);
    }

    #[test]
    fn test_table_limiter_allows_within_limit() {
        let mut limiter = WasmResourceLimiter::new(128);
        let result = limiter.table_growing(0, 1000, None);
        assert!(result.is_ok());
        assert!(result.unwrap());
    }

    #[test]
    fn test_table_limiter_denies_over_limit() {
        let mut limiter = WasmResourceLimiter::new(128);
        let result = limiter.table_growing(0, 100_000, None);
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }
}
