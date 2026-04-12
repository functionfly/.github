//! Instance metadata tracking for the legacy pool.

use std::time::Instant;

/// A pooled Wasm instance with memory tracking (metadata-only).
///
/// This is used by the legacy `InstancePool` for metadata tracking only.
/// The actual WASM instances are managed by `PooledWasmInstance` in the
/// warm-instance pool.
#[derive(Clone, Debug)]
pub struct PooledInstance {
    /// When the instance was created
    pub created_at: Instant,
    /// When the instance was last used
    pub last_used: Instant,
    /// Instance ID for tracking
    pub instance_id: String,
    /// Estimated memory usage in bytes
    pub memory_usage: usize,
    /// Number of times this instance has been reused
    pub reuse_count: u32,
    /// Function key this instance is associated with
    pub function_key: String,
}

impl PooledInstance {
    /// Create a new test instance for testing purposes.
    #[cfg(test)]
    pub fn create_test(instance_id: &str, memory_usage: usize) -> Self {
        Self {
            created_at: Instant::now(),
            last_used: Instant::now(),
            instance_id: instance_id.to_string(),
            memory_usage,
            reuse_count: 0,
            function_key: "test".to_string(),
        }
    }
}
