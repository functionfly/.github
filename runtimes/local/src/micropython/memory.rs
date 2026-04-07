//! Shared memory management for MicroPython WASM execution.

use std::sync::Arc;
use tokio::sync::RwLock;

/// Memory layout for shared linear memory between wrapper and MicroPython.
///
/// The memory is organized as follows:
/// ```
/// 0x00000 - 0x0FFFF: Wrapper static data (64KB)
/// 0x10000 - 0x1FFFF: MicroPython stack (64KB)
/// 0x20000 - 0x9FFFF: MicroPython heap (512KB)
/// 0xA0000 - 0xDFFFF: User code buffer (256KB)
/// 0xE0000 - 0xEFFFF: Output buffer (64KB)
/// 0xF0000+: Dynamic allocation area
/// ```
#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct MemoryLayout {
    /// Base address of wrapper static data
    pub wrapper_data_base: u32,
    /// Size of wrapper static data
    pub wrapper_data_size: u32,

    /// Base address of MicroPython stack
    pub stack_base: u32,
    /// Size of MicroPython stack
    pub stack_size: u32,

    /// Base address of MicroPython heap
    pub heap_base: u32,
    /// Size of MicroPython heap
    pub heap_size: u32,

    /// Base address of user code buffer
    pub code_buffer_base: u32,
    /// Size of user code buffer
    pub code_buffer_size: u32,

    /// Base address of output buffer
    pub output_buffer_base: u32,
    /// Size of output buffer
    pub output_buffer_size: u32,

    /// Base address of dynamic allocation area
    pub dynamic_base: u32,
}

impl Default for MemoryLayout {
    fn default() -> Self {
        Self {
            wrapper_data_base: 0x00000,
            wrapper_data_size: 0x10000, // 64KB

            stack_base: 0x10000,
            stack_size: 0x10000, // 64KB

            heap_base: 0x20000,
            heap_size: 0x80000, // 512KB

            code_buffer_base: 0xA0000,
            code_buffer_size: 0x40000, // 256KB

            output_buffer_base: 0xE0000,
            output_buffer_size: 0x10000, // 64KB

            dynamic_base: 0xF0000,
        }
    }
}

impl MemoryLayout {
    /// Create a new memory layout with custom heap size.
    pub fn with_heap_size(heap_size_kb: u32) -> Self {
        let mut layout = Self::default();
        layout.heap_size = heap_size_kb * 1024;
        // Adjust subsequent regions
        layout.code_buffer_base = layout.heap_base + layout.heap_size;
        layout.output_buffer_base = layout.code_buffer_base + layout.code_buffer_size;
        layout.dynamic_base = layout.output_buffer_base + layout.output_buffer_size;
        layout
    }

    /// Get the total initial memory size in bytes.
    pub fn total_initial_size(&self) -> u32 {
        // Round up to nearest 64KB (WASM page size)
        let total = self.dynamic_base + 0x10000; // Add some padding
        total.div_ceil(0x10000) * 0x10000
    }

    /// Get the number of WASM pages (64KB each) needed.
    pub fn initial_pages(&self) -> u32 {
        self.total_initial_size() / 0x10000
    }

    /// Check if an address is within the code buffer.
    #[allow(dead_code)]
    pub fn is_code_buffer(&self, addr: u32) -> bool {
        addr >= self.code_buffer_base
            && addr < self.code_buffer_base + self.code_buffer_size
    }

    /// Check if an address is within the output buffer.
    #[allow(dead_code)]
    pub fn is_output_buffer(&self, addr: u32) -> bool {
        addr >= self.output_buffer_base
            && addr < self.output_buffer_base + self.output_buffer_size
    }
}

/// Memory manager for shared linear memory.
pub struct MemoryManager {
    #[allow(dead_code)]
    layout: MemoryLayout,
    #[allow(dead_code)]
    alloc_ptr: Arc<RwLock<u32>>,
}

impl MemoryManager {
    /// Create a new memory manager with the given layout.
    pub fn new(layout: MemoryLayout) -> Self {
        let dynamic_base = layout.dynamic_base;
        Self {
            layout,
            alloc_ptr: Arc::new(RwLock::new(dynamic_base)),
        }
    }

    /// Get the memory layout.
    #[allow(dead_code)]
    pub fn layout(&self) -> &MemoryLayout {
        &self.layout
    }

    /// Allocate memory from the dynamic area (bump allocator).
    /// For production use, this should be replaced with a real allocator.
    #[allow(dead_code)]
    pub async fn allocate(&self, size: u32) -> Option<u32> {
        let mut ptr = self.alloc_ptr.write().await;
        let aligned_size = (size + 7) & !7; // Align to 8 bytes
        let current = *ptr;
        *ptr += aligned_size;
        Some(current)
    }

    /// Get the base address for user code.
    #[allow(dead_code)]
    pub fn code_base(&self) -> u32 {
        self.layout.code_buffer_base
    }

    /// Get the maximum size for user code.
    #[allow(dead_code)]
    pub fn code_size(&self) -> u32 {
        self.layout.code_buffer_size
    }

    /// Get the base address for output.
    #[allow(dead_code)]
    pub fn output_base(&self) -> u32 {
        self.layout.output_buffer_base
    }

    /// Get the maximum size for output.
    #[allow(dead_code)]
    pub fn output_size(&self) -> u32 {
        self.layout.output_buffer_size
    }
}

impl Default for MemoryManager {
    fn default() -> Self {
        Self::new(MemoryLayout::default())
    }
}

/// Host state passed to the WASM store during execution.
pub struct HostState {
    /// Input data (JSON string)
    pub input: String,
    /// Output data (captured from execution)
    pub output: Arc<RwLock<String>>,
    /// Memory manager
    #[allow(dead_code)]
    pub memory: MemoryManager,
    /// Execution logs
    pub logs: Arc<RwLock<Vec<String>>>,
}

impl HostState {
    /// Create new host state with input data.
    pub fn new(input: impl Into<String>) -> Self {
        Self {
            input: input.into(),
            output: Arc::new(RwLock::new(String::new())),
            memory: MemoryManager::default(),
            logs: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Log a message.
    #[allow(dead_code)]
    pub async fn log(&self, message: String) {
        self.logs.write().await.push(message);
    }

    /// Set the output.
    #[allow(dead_code)]
    pub async fn set_output(&self, output: String) {
        *self.output.write().await = output;
    }

    /// Get the output.
    pub async fn get_output(&self) -> String {
        self.output.read().await.clone()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_memory_layout() {
        let layout = MemoryLayout::default();
        assert_eq!(layout.wrapper_data_base, 0x00000);
        assert_eq!(layout.stack_base, 0x10000);
        assert_eq!(layout.heap_base, 0x20000);
        assert_eq!(layout.code_buffer_base, 0xA0000);
        assert_eq!(layout.output_buffer_base, 0xE0000);
        assert_eq!(layout.dynamic_base, 0xF0000);
    }

    #[test]
    fn test_memory_layout_with_heap_size() {
        let layout = MemoryLayout::with_heap_size(1024); // 1MB heap
        assert_eq!(layout.heap_size, 1024 * 1024);
        assert_eq!(layout.code_buffer_base, 0x20000 + 0x100000); // heap_base + heap_size
    }

    #[test]
    fn test_initial_pages() {
        let layout = MemoryLayout::default();
        let pages = layout.initial_pages();
        assert!(pages >= 16); // At least 1MB
    }

    #[tokio::test]
    async fn test_memory_manager_allocate() {
        let manager = MemoryManager::default();
        let ptr1 = manager.allocate(100).await.unwrap();
        let ptr2 = manager.allocate(100).await.unwrap();
        assert_eq!(ptr2, ptr1 + 104); // 100 bytes aligned to 8 = 104
    }

    #[tokio::test]
    async fn test_host_state() {
        let state = HostState::new(r#"{"test": "value"}"#);
        assert_eq!(state.input, r#"{"test": "value"}"#);

        state.set_output("result".to_string()).await;
        assert_eq!(state.get_output().await, "result");
    }
}
