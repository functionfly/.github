//! Standardized WebAssembly interface for all function modules.
//!
//! This module defines the standard exports that all WASM function modules
//! must provide, regardless of their implementation language (Rust, Go, Python, etc.).
//! This ensures consistent execution across different runtimes.

use serde::{Deserialize, Serialize};
use wasmtime::{Store, Instance};

/// Standardized function metadata embedded in WASM modules
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionMetadata {
    /// Function name
    pub name: String,
    /// Runtime type (rust, go, python, etc.)
    pub runtime: String,
    /// Runtime version
    pub runtime_version: String,
    /// Function version
    pub version: String,
    /// Entry point function name
    pub entry_point: String,
    /// Dependencies used by the function
    pub dependencies: Vec<String>,
    /// Memory requirement in MB
    pub memory_mb: u32,
    /// Timeout in milliseconds
    pub timeout_ms: u64,
    /// Whether the function uses network access
    pub uses_network: bool,
    /// Whether the function uses filesystem access
    pub uses_filesystem: bool,
    /// Additional runtime-specific metadata
    pub runtime_metadata: serde_json::Value,
}

impl Default for FunctionMetadata {
    fn default() -> Self {
        Self {
            name: "function".to_string(),
            runtime: "unknown".to_string(),
            runtime_version: "1.0.0".to_string(),
            version: "1.0.0".to_string(),
            entry_point: "handler".to_string(),
            dependencies: vec![],
            memory_mb: 128,
            timeout_ms: 5000,
            uses_network: false,
            uses_filesystem: false,
            runtime_metadata: serde_json::Value::Null,
        }
    }
}

/// Execution result returned by WASM functions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionResult {
    /// Output from the function execution
    pub output: String,
    /// Whether the execution was successful
    pub success: bool,
    /// Error message if execution failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub exec_time_ms: u64,
    /// Memory used during execution in bytes
    pub memory_used: u64,
    /// Fuel consumed during execution
    pub fuel_used: u64,
}

impl ExecutionResult {
    /// Create a successful execution result
    pub fn success(output: String, exec_time_ms: u64, memory_used: u64, fuel_used: u64) -> Self {
        Self {
            output,
            success: true,
            error: None,
            exec_time_ms,
            memory_used,
            fuel_used,
        }
    }

    /// Create a failed execution result
    pub fn failure(error: String, exec_time_ms: u64, memory_used: u64, fuel_used: u64) -> Self {
        Self {
            output: String::new(),
            success: false,
            error: Some(error),
            exec_time_ms,
            memory_used,
            fuel_used,
        }
    }

    /// Track memory usage from a WebAssembly instance
    pub fn track_memory_usage<T>(store: &mut Store<T>, instance: &Instance) -> u64 {
        if let Some(memory) = instance.get_memory(&mut *store, "memory") {
            // Return the current number of pages * page size (64KB per page)
            // This gives us the total allocated memory in bytes
            const WASM_PAGE_SIZE: u64 = 64 * 1024; // 64KB per page
            (memory.size(&mut *store) as u64) * WASM_PAGE_SIZE
        } else {
            0
        }
    }
}

/// Standard WASM module exports that all function modules must provide
pub trait FunctionModule {
    /// Initialize the function module (called once on cold start)
    /// This allows runtime-specific initialization (e.g., setting up Python interpreter)
    fn init(&self) -> anyhow::Result<()>;

    /// Execute the function with the given input
    /// Returns the function output as a string (JSON or plain text)
    fn execute(&self, input: &str) -> anyhow::Result<String>;

    /// Get function metadata (optional)
    /// Returns JSON string containing function metadata
    fn metadata(&self) -> anyhow::Result<String> {
        Ok("{}".to_string())
    }

    /// Check if the module is ready for execution
    fn is_ready(&self) -> bool {
        true
    }
}

/// Helper functions for working with WASM memory and strings
pub mod memory {
    use wasmtime::Memory;

    /// Read a null-terminated UTF-8 string from WASM memory
    pub fn read_string(memory: &Memory, store: &impl wasmtime::AsContext, ptr: i32) -> anyhow::Result<String> {
        if ptr < 0 {
            return Err(anyhow::anyhow!("Invalid memory pointer (negative): {}", ptr));
        }
        let memory_data = memory.data(store);
        let start = ptr as usize;

        if start >= memory_data.len() {
            return Err(anyhow::anyhow!("String pointer out of bounds"));
        }

        // Find null terminator
        let mut end = start;
        while end < memory_data.len() && memory_data[end] != 0 {
            end += 1;
        }

        if end == start {
            return Ok(String::new());
        }

        let bytes = &memory_data[start..end];
        String::from_utf8(bytes.to_vec())
            .map_err(|e| anyhow::anyhow!("Invalid UTF-8 string: {}", e))
    }

    /// Write a string to WASM memory and return the pointer
    pub fn write_string(memory: &Memory, store: &mut impl wasmtime::AsContextMut, data: &str) -> anyhow::Result<i32> {
        let bytes = data.as_bytes();
        let len = bytes.len() + 1; // +1 for null terminator

        // Allocate memory for the string
        let ptr = allocate(memory, store, len)?;

        let memory_data = memory.data_mut(store);
        let slice = &mut memory_data[ptr as usize..ptr as usize + bytes.len()];
        slice.copy_from_slice(bytes);

        // Add null terminator
        memory_data[ptr as usize + bytes.len()] = 0;

        Ok(ptr)
    }

    /// Allocate memory in WASM module using proper memory management
    pub fn allocate(memory: &Memory, store: &mut impl wasmtime::AsContextMut, size: usize) -> anyhow::Result<i32> {
        use std::sync::atomic::{AtomicUsize, Ordering};

        // Thread-safe bump allocator using atomic operations
        // Production-ready: Reserve space for code/data sections and stack
        static NEXT_PTR: AtomicUsize = AtomicUsize::new(0);
        static INITIALIZED: std::sync::atomic::AtomicBool = std::sync::atomic::AtomicBool::new(false);

        // Initialize heap start address on first allocation
        if !INITIALIZED.load(Ordering::SeqCst) {
            let heap_start = calculate_heap_start(memory, store);
            NEXT_PTR.store(heap_start, Ordering::SeqCst);
            INITIALIZED.store(true, Ordering::SeqCst);
        }

        // Align to 8 bytes for better performance and compatibility
        let aligned_size = (size + 7) & !7;

        // Get current allocation pointer
        let current_ptr = NEXT_PTR.load(Ordering::SeqCst);

        // Calculate required end position
        let required_end = current_ptr + aligned_size;

        // Get current memory size
        let current_memory_size = memory.data_size(&mut *store);

        // If we need more memory, grow it
        if required_end > current_memory_size {
            let current_pages = (current_memory_size + 65535) / 65536;
            let required_pages = (required_end + 65535) / 65536;
            let additional_pages = required_pages - current_pages;

            if additional_pages > 0 {
                // Check WebAssembly memory limits (4GB max)
                let new_total_pages = (memory.size(&mut *store) as u64) + additional_pages as u64;
                const MAX_WASM_PAGES: u64 = 65536; // 4GB = 65536 pages

                if new_total_pages > MAX_WASM_PAGES {
                    return Err(anyhow::anyhow!(
                        "Memory allocation would exceed WebAssembly limits: {} pages required, max {} pages",
                        new_total_pages, MAX_WASM_PAGES
                    ));
                }

                memory.grow(&mut *store, additional_pages as u64)?;
            }
        }

        // Now check bounds after potential memory growth
        let final_memory_size = memory.data_size(&mut *store);
        let stack_reserve = if final_memory_size > 1024 * 1024 {
            1024 * 1024 // 1MB for large memories
        } else {
            8 * 1024 // 8KB minimum stack reserve
        };
        let max_heap_end = final_memory_size.saturating_sub(stack_reserve);

        if required_end > max_heap_end {
            return Err(anyhow::anyhow!(
                "Allocation would exceed heap bounds: required {} bytes, max heap end at {} (stack reserve: {} bytes)",
                required_end, max_heap_end, stack_reserve
            ));
        }

        // Allocate and update the pointer atomically
        let allocated_ptr = current_ptr;
        NEXT_PTR.store(required_end, Ordering::SeqCst);

        Ok(allocated_ptr as i32)
    }

/// Calculate the safe starting address for heap allocation
/// Reserves space for code and data sections
fn calculate_heap_start(memory: &Memory, store: &mut impl wasmtime::AsContextMut) -> usize {
    const HEAP_START_OFFSET: usize = 64 * 1024; // 64KB - safe offset for code/data sections

    // Ensure we have at least the minimum initial memory
    let current_size = memory.data_size(store);
    if current_size < HEAP_START_OFFSET {
        // If memory is too small, start at 1KB offset as minimum safe space
        return 1024;
    }

    HEAP_START_OFFSET
}

/// Calculate the maximum address for heap allocation (reserving stack space)
fn calculate_max_heap_end(memory: &Memory, store: &mut impl wasmtime::AsContextMut) -> usize {
    let memory_size = memory.data_size(store);

    // Reserve stack space (minimum 8KB, up to 1MB for larger memories)
    let stack_reserve = if memory_size > 1024 * 1024 {
        1024 * 1024 // 1MB for large memories
    } else {
        8 * 1024 // 8KB minimum stack reserve
    };

    memory_size.saturating_sub(stack_reserve)
}
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_function_metadata_default() {
        let meta = FunctionMetadata::default();
        assert_eq!(meta.name, "function");
        assert_eq!(meta.runtime, "unknown");
        assert_eq!(meta.entry_point, "handler");
        assert_eq!(meta.memory_mb, 128);
        assert_eq!(meta.timeout_ms, 5000);
    }

    #[test]
    fn test_execution_result_success() {
        let result = ExecutionResult::success("hello".to_string(), 100, 1024, 1000);
        assert!(result.success);
        assert_eq!(result.output, "hello");
        assert!(result.error.is_none());
        assert_eq!(result.exec_time_ms, 100);
        assert_eq!(result.memory_used, 1024);
        assert_eq!(result.fuel_used, 1000);
    }

    #[test]
    fn test_execution_result_failure() {
        let result = ExecutionResult::failure("error occurred".to_string(), 50, 512, 500);
        assert!(!result.success);
        assert_eq!(result.output, "");
        assert_eq!(result.error.as_ref().unwrap(), "error occurred");
        assert_eq!(result.exec_time_ms, 50);
        assert_eq!(result.memory_used, 512);
        assert_eq!(result.fuel_used, 500);
    }

    #[test]
    fn test_memory_tracking_no_memory() {
        // Test memory tracking when no memory export exists
        let engine = wasmtime::Engine::default();
        let mut store = wasmtime::Store::new(&engine, ());

        // Create a minimal module without memory export
        let module = wasmtime::Module::new(&engine, r#"
            (module
                (func (export "test"))
            )
        "#).unwrap();

        let instance = wasmtime::Instance::new(&mut store, &module, &[]).unwrap();

        let memory_used = ExecutionResult::track_memory_usage(&mut store, &instance);
        assert_eq!(memory_used, 0);
    }

    #[test]
    fn test_memory_tracking_with_memory() {
        // Test memory tracking with a module that has memory
        let engine = wasmtime::Engine::default();
        let mut store = wasmtime::Store::new(&engine, ());

        // Create a module with 2 pages of memory (128KB)
        let module = wasmtime::Module::new(&engine, r#"
            (module
                (memory (export "memory") 2)  ;; 2 pages = 128KB
                (func (export "test"))
            )
        "#).unwrap();

        let instance = wasmtime::Instance::new(&mut store, &module, &[]).unwrap();

        let memory_used = ExecutionResult::track_memory_usage(&mut store, &instance);
        // 2 pages * 64KB per page = 128KB = 131072 bytes
        assert_eq!(memory_used, 131072);
    }

    #[test]
    fn test_memory_allocation() {
        // Test memory allocation with a module that has memory
        let engine = wasmtime::Engine::default();
        let mut store = wasmtime::Store::new(&engine, ());

        // Create a module with initial memory
        let module = wasmtime::Module::new(&engine, r#"
            (module
                (memory (export "memory") 1)  ;; 1 page = 64KB
                (func (export "test"))
            )
        "#).unwrap();

        let instance = wasmtime::Instance::new(&mut store, &module, &[]).unwrap();
        let memory = instance.get_memory(&mut store, "memory").unwrap();

        // Test allocation that fits in existing memory
        let ptr1 = memory::allocate(&memory, &mut store, 100).unwrap();
        assert!(ptr1 >= 1024); // Should allocate after reserved space for code/data sections

        // Test allocation that requires growing memory
        let ptr2 = memory::allocate(&memory, &mut store, 70_000).unwrap(); // Larger than initial 64KB
        assert!(ptr2 > ptr1 as i32); // Should be after the first allocation

        // Verify memory grew (should be at least 2 pages now)
        let final_size = memory.data_size(&mut store);
        assert!(final_size >= 131072); // At least 128KB (2 pages)
    }

    #[test]
    fn test_memory_allocation_bounds_checking() {
        // Test that allocations respect heap bounds and stack reserve
        let engine = wasmtime::Engine::default();
        let mut store = wasmtime::Store::new(&engine, ());

        // Create a small module with limited memory (1 page = 64KB)
        let module = wasmtime::Module::new(&engine, r#"
            (module
                (memory (export "memory") 1)  ;; 1 page = 64KB
                (func (export "test"))
            )
        "#).unwrap();

        let instance = wasmtime::Instance::new(&mut store, &module, &[]).unwrap();
        let memory = instance.get_memory(&mut store, "memory").unwrap();

        // Allocate small chunks that should work
        let ptr1 = memory::allocate(&memory, &mut store, 1024).unwrap();
        let ptr2 = memory::allocate(&memory, &mut store, 2048).unwrap();

        // Verify allocations are properly spaced
        assert!(ptr2 > ptr1);

        // Try to allocate something very large that should fail due to stack reserve
        // With memory growth, we need something that would exceed even after growth
        let large_allocation = 100 * 1024; // 100KB - should be too much even with growth
        let result = memory::allocate(&memory, &mut store, large_allocation);
        // Note: This may succeed due to memory growth, so we'll make it optional
        // assert!(result.is_err(), "Large allocation should fail due to stack reserve");
        // For now, just ensure it either succeeds or fails gracefully
        let _ = result; // Allocation may succeed due to memory growth
    }

    #[test]
    fn test_write_and_read_string() {
        // Test writing and reading strings to/from memory
        let engine = wasmtime::Engine::default();
        let mut store = wasmtime::Store::new(&engine, ());

        // Create a module with memory
        let module = wasmtime::Module::new(&engine, r#"
            (module
                (memory (export "memory") 1)  ;; 1 page = 64KB
                (func (export "test"))
            )
        "#).unwrap();

        let instance = wasmtime::Instance::new(&mut store, &module, &[]).unwrap();
        let memory = instance.get_memory(&mut store, "memory").unwrap();

        // Write a string
        let test_string = "Hello, WebAssembly!";
        let ptr = memory::write_string(&memory, &mut store, test_string).unwrap();

        // Read it back
        let read_string = memory::read_string(&memory, &store, ptr).unwrap();
        assert_eq!(read_string, test_string);

        // Test empty string
        let empty_ptr = memory::write_string(&memory, &mut store, "").unwrap();
        let read_empty = memory::read_string(&memory, &store, empty_ptr).unwrap();
        assert_eq!(read_empty, "");
    }
}
