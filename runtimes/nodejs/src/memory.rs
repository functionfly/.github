//! Memory Limiting
//! 
//! This module provides memory limit enforcement for function execution,
//! tracking heap usage and enforcing configurable limits.

use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;

use parking_lot::RwLock;
use tracing::{debug, warn};

use crate::RuntimeError;

/// Manages memory limits for executions
pub struct MemoryLimiter {
    /// Maximum memory in MB
    max_memory_mb: u32,
    
    /// Current memory usage (bytes)
    current_usage: AtomicU64,
    
    /// Peak memory usage (bytes)
    peak_usage: AtomicU64,
    
    /// Total allocations
    total_allocations: AtomicU64,
    
    /// Total deallocations
    total_deallocations: AtomicU64,
    
    /// Number of times limit was exceeded
    limit_exceeded_count: AtomicU64,
    
    /// Active memory tracking per execution
    active_tracking: RwLock<Vec<MemoryTrackEntry>>,
}

/// Memory tracking for a single execution
#[derive(Debug, Clone)]
struct MemoryTrackEntry {
    execution_id: String,
    allocated_bytes: u64,
    peak_bytes: u64,
}

impl MemoryLimiter {
    /// Create a new memory limiter
    pub fn new(max_memory_mb: u32) -> Self {
        let max_bytes = (max_memory_mb as u64) * 1024 * 1024;
        
        Self {
            max_memory_mb,
            current_usage: AtomicU64::new(0),
            peak_usage: AtomicU64::new(0),
            total_allocations: AtomicU64::new(0),
            total_deallocations: AtomicU64::new(0),
            limit_exceeded_count: AtomicU64::new(0),
            active_tracking: RwLock::new(Vec::new()),
        }
    }

    /// Get the maximum memory limit in bytes
    pub fn max_bytes(&self) -> u64 {
        (self.max_memory_mb as u64) * 1024 * 1024
    }

    /// Get the maximum memory limit in MB
    pub fn max_mb(&self) -> u32 {
        self.max_memory_mb
    }

    /// Try to allocate memory for an execution
    /// Returns Ok(()) if allocation is allowed, Err if limit exceeded
    pub fn try_allocate(&self, execution_id: &str, bytes: u64) -> Result<MemoryGuard, RuntimeError> {
        let current = self.current_usage.fetch_add(bytes, Ordering::Relaxed);
        let new_total = current + bytes;
        
        self.total_allocations.fetch_add(bytes, Ordering::Relaxed);
        
        // Check if we've exceeded the limit
        if new_total > self.max_bytes() {
            // Rollback the allocation
            self.current_usage.fetch_sub(bytes, Ordering::Relaxed);
            self.limit_exceeded_count.fetch_add(1, Ordering::Relaxed);
            
            warn!(
                "Memory limit exceeded: {} bytes requested, {} bytes used of {} limit",
                bytes,
                current,
                self.max_bytes()
            );
            
            return Err(RuntimeError::MemoryLimit(format!(
                "Memory limit exceeded: {}MB limit, {}MB used",
                self.max_memory_mb,
                current / (1024 * 1024)
            )));
        }
        
        // Update peak if needed
        let current_peak = self.peak_usage.load(Ordering::Relaxed);
        if new_total > current_peak {
            self.peak_usage.store(new_total, Ordering::Relaxed);
        }
        
        // Track this allocation
        {
            let mut tracking = self.active_tracking.write();
            tracking.push(MemoryTrackEntry {
                execution_id: execution_id.to_string(),
                allocated_bytes: bytes,
                peak_bytes: new_total,
            });
        }
        
        debug!("Allocated {} bytes for execution {}", bytes, execution_id);
        
        Ok(MemoryGuard {
            limiter: Arc::new(self.clone_inner()),
            execution_id: execution_id.to_string(),
            bytes,
        })
    }

    /// Get current memory usage in bytes
    pub fn current_bytes(&self) -> u64 {
        self.current_usage.load(Ordering::Relaxed)
    }

    /// Get current memory usage in MB
    pub fn current_mb(&self) -> f64 {
        self.current_bytes() as f64 / (1024.0 * 1024.0)
    }

    /// Get peak memory usage in bytes
    pub fn peak_bytes(&self) -> u64 {
        self.peak_usage.load(Ordering::Relaxed)
    }

    /// Get peak memory usage in MB
    pub fn peak_mb(&self) -> f64 {
        self.peak_bytes() as f64 / (1024.0 * 1024.0)
    }

    /// Get statistics
    pub fn stats(&self) -> MemoryStats {
        MemoryStats {
            max_memory_mb: self.max_memory_mb,
            current_bytes: self.current_bytes(),
            peak_bytes: self.peak_bytes(),
            total_allocations: self.total_allocations.load(Ordering::Relaxed),
            total_deallocations: self.total_deallocations.load(Ordering::Relaxed),
            limit_exceeded_count: self.limit_exceeded_count.load(Ordering::Relaxed),
        }
    }

    /// Clone inner for MemoryGuard
    fn clone_inner(&self) -> MemoryLimiterInner {
        MemoryLimiterInner {
            max_memory_mb: self.max_memory_mb,
            current_usage: Arc::new(AtomicU64::new(self.current_bytes())),
            peak_usage: Arc::new(AtomicU64::new(self.peak_bytes())),
        }
    }
}

impl Clone for MemoryLimiter {
    fn clone(&self) -> Self {
        Self {
            max_memory_mb: self.max_memory_mb,
            current_usage: AtomicU64::new(self.current_bytes()),
            peak_usage: AtomicU64::new(self.peak_bytes()),
            total_allocations: AtomicU64::new(self.total_allocations.load(Ordering::Relaxed)),
            total_deallocations: AtomicU64::new(self.total_deallocations.load(Ordering::Relaxed)),
            limit_exceeded_count: AtomicU64::new(self.limit_exceeded_count.load(Ordering::Relaxed)),
            active_tracking: RwLock::new(Vec::new()),
        }
    }
}

/// Inner struct for MemoryGuard
#[derive(Clone)]
struct MemoryLimiterInner {
    max_memory_mb: u32,
    current_usage: Arc<AtomicU64>,
    peak_usage: Arc<AtomicU64>,
}

/// Guard that automatically deallocates memory when dropped
pub struct MemoryGuard {
    limiter: Arc<MemoryLimiterInner>,
    execution_id: String,
    bytes: u64,
}

impl MemoryGuard {
    /// Get the allocated bytes
    pub fn bytes(&self) -> u64 {
        self.bytes
    }
}

impl Drop for MemoryGuard {
    fn drop(&mut self) {
        // Deallocate the memory
        let current = self.limiter.current_usage.fetch_sub(self.bytes, Ordering::Relaxed);
        debug!(
            "Deallocated {} bytes for execution {}. Current: {} bytes",
            self.bytes,
            self.execution_id,
            current - self.bytes
        );
    }
}

/// Memory statistics
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct MemoryStats {
    pub max_memory_mb: u32,
    pub current_bytes: u64,
    pub peak_bytes: u64,
    pub total_allocations: u64,
    pub total_deallocations: u64,
    pub limit_exceeded_count: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_memory_limiter_creation() {
        let limiter = MemoryLimiter::new(128);
        let stats = limiter.stats();
        
        assert_eq!(stats.max_memory_mb, 128);
        assert_eq!(stats.current_bytes, 0);
    }

    #[test]
    fn test_memory_allocation() {
        let limiter = MemoryLimiter::new(1); // 1 MB
        let result = limiter.try_allocate("test-1", 1024 * 1024); // 1 MB
        
        assert!(result.is_ok());
        
        let stats = limiter.stats();
        assert_eq!(stats.current_bytes, 1024 * 1024);
    }

    #[test]
    fn test_memory_limit_exceeded() {
        let limiter = MemoryLimiter::new(1); // 1 MB
        let result = limiter.try_allocate("test-1", 2 * 1024 * 1024); // 2 MB
        
        assert!(result.is_err());
        
        let stats = limiter.stats();
        assert_eq!(stats.limit_exceeded_count, 1);
    }
}
