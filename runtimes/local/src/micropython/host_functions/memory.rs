//! Memory allocation for MicroPython WASM.
//!
//! Implements a bump allocator for the MicroPython heap. The allocator
//! tracks allocations to support `free`, using a simple free list for
//! reclaimed memory.

use crate::micropython::memory::HostState;
use crate::micropython::errors::MicroPythonError;
use wasmtime::{Linker, Store};
use std::sync::Mutex;

struct BumpAllocator {
    next: usize,
    end: usize,
    free_list: Vec<(usize, usize)>,
}

impl BumpAllocator {
    fn new(start: usize, size: usize) -> Self {
        Self {
            next: start,
            end: start + size,
            free_list: Vec::new(),
        }
    }

    fn alloc(&mut self, size: usize) -> Option<usize> {
        // Align to 8 bytes
        let aligned_size = (size + 7) & !7;

        // Check free list first
        if let Some(pos) = self.free_list.iter().position(|&(s, _)| s >= aligned_size) {
            let (block_size, ptr) = self.free_list.remove(pos);
            // Split the block if there's remaining space
            let remaining = block_size - aligned_size;
            if remaining >= 16 {
                self.free_list.push((remaining, ptr + aligned_size));
            }
            return Some(ptr);
        }

        // Allocate from the bump pointer
        if self.next + aligned_size <= self.end {
            let ptr = self.next;
            self.next += aligned_size;
            Some(ptr)
        } else {
            None // Out of memory
        }
    }

    fn free(&mut self, ptr: usize, size: usize) {
        // Add to free list for potential reuse
        // In a real allocator, we'd merge adjacent free blocks
        let aligned_size = (size + 7) & !7;
        self.free_list.push((aligned_size, ptr));
    }

    fn reset(&mut self) {
        self.next = self.next - (self.next - self.free_list.first().map(|&(_, p)| p).unwrap_or(self.next));
        self.free_list.clear();
    }
}

// Global bump allocator state (protected by mutex for thread safety)
static ALLOCATOR: Mutex<Option<BumpAllocator>> = Mutex::new(None);

/// Register malloc and free host functions.
pub fn register(linker: &mut Linker<HostState>, _store: &mut Store<HostState>) -> Result<(), MicroPythonError> {
    // env.malloc(size: i32) -> ptr i32
    linker.func_wrap(
        "env",
        "malloc",
        |_caller: wasmtime::Caller<'_, HostState>, size: i32| -> i32 {
            if size <= 0 {
                return 0;
            }

            let mut guard = match ALLOCATOR.lock() {
                Ok(g) => g,
                Err(e) => {
                    tracing::error!("malloc: failed to lock allocator: {}", e);
                    return 0;
                }
            };

            match guard.as_mut() {
                Some(allocator) => {
                    match allocator.alloc(size as usize) {
                        Some(ptr) => {
                            tracing::trace!("malloc({}) = {}", size, ptr);
                            ptr as i32
                        }
                        None => {
                            tracing::warn!("malloc({}): out of memory", size);
                            0
                        }
                    }
                }
                None => {
                    tracing::error!("malloc: allocator not initialized");
                    0
                }
            }
        },
    ).map_err(|e| {
        MicroPythonError::LinkError(format!("Failed to register malloc: {}", e))
    })?;

    // env.free(ptr: i32) -> void
    linker.func_wrap(
        "env",
        "free",
        |_caller: wasmtime::Caller<'_, HostState>, ptr: i32, size: i32| {
            if ptr <= 0 || size <= 0 {
                return;
            }

            let mut guard = match ALLOCATOR.lock() {
                Ok(g) => g,
                Err(e) => {
                    tracing::error!("free: failed to lock allocator: {}", e);
                    return;
                }
            };

            if let Some(allocator) = guard.as_mut() {
                allocator.free(ptr as usize, size as usize);
                tracing::trace!("free({}, {})", ptr, size);
            }
        },
    ).map_err(|e| {
        MicroPythonError::LinkError(format!("Failed to register free: {}", e))
    })?;

    tracing::debug!("Registered malloc and free");
    Ok(())
}

/// Initialize the bump allocator with the given memory region.
pub fn init_allocator(dynamic_base: u32, heap_size: u32) {
    let mut guard = ALLOCATOR.lock().unwrap();
    *guard = Some(BumpAllocator::new(dynamic_base as usize, heap_size as usize));
    tracing::debug!("Initialized bump allocator at 0x{:x} with {} bytes", dynamic_base, heap_size);
}

/// Reset the allocator (called on mp_js_init).
#[allow(dead_code)]
pub fn reset_allocator() {
    let mut guard = ALLOCATOR.lock().unwrap();
    if let Some(ref mut allocator) = *guard {
        allocator.reset();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bump_allocator() {
        let mut allocator = BumpAllocator::new(0x1000, 1024);

        let ptr1 = allocator.alloc(100).unwrap();
        assert!(ptr1 >= 0x1000);

        let ptr2 = allocator.alloc(100).unwrap();
        assert!(ptr2 > ptr1);

        // Free and reallocate
        allocator.free(ptr1, 100);
        let ptr3 = allocator.alloc(100).unwrap();
        assert!(ptr3 >= 0x1000);
    }

    #[test]
    fn test_bump_allocator_oom() {
        let mut allocator = BumpAllocator::new(0x1000, 100);

        assert!(allocator.alloc(50).is_some());
        assert!(allocator.alloc(60).is_none()); // Not enough space
        assert!(allocator.alloc(100).is_none());
    }
}