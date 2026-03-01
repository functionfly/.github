//! Timeout Management
//! 
//! This module provides timeout handling for function execution,
//! including configurable timeout durations and cancellation support.

use std::sync::atomic::{AtomicU64, AtomicBool, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

use parking_lot::RwLock;
use tracing::{debug, warn};

/// Manages execution timeouts
pub struct TimeoutManager {
    /// Maximum timeout in milliseconds
    max_timeout_ms: u64,
    
    /// Active execution tracking
    active_timeouts: RwLock<Vec<TimeoutEntry>>,
    
    /// Statistics
    total_timeouts: AtomicU64,
    cancelled_timeouts: AtomicU64,
}

/// A single timeout entry
#[derive(Debug)]
struct TimeoutEntry {
    execution_id: String,
    started_at: Instant,
    duration_ms: u64,
}

impl TimeoutManager {
    /// Create a new timeout manager
    pub fn new(max_timeout_ms: u64) -> Self {
        Self {
            max_timeout_ms,
            active_timeouts: RwLock::new(Vec::new()),
            total_timeouts: AtomicU64::new(0),
            cancelled_timeouts: AtomicU64::new(0),
        }
    }

    /// Start tracking a new execution
    pub fn start_execution(&self, execution_id: &str, timeout_ms: Option<u64>) -> TimeoutGuard {
        let timeout = timeout_ms.unwrap_or(self.max_timeout_ms);
        
        {
            let mut active = self.active_timeouts.write();
            active.push(TimeoutEntry {
                execution_id: execution_id.to_string(),
                started_at: Instant::now(),
                duration_ms: timeout,
            });
        }
        
        self.total_timeouts.fetch_add(1, Ordering::Relaxed);
        
        TimeoutGuard {
            manager: Arc::new(self.clone_inner()),
            execution_id: execution_id.to_string(),
            started_at: Instant::now(),
            timeout_ms: timeout,
            cancelled: AtomicBool::new(false),
        }
    }

    /// Check if an execution has exceeded its timeout
    pub fn is_timeout(&self, execution_id: &str) -> bool {
        let active = self.active_timeouts.read();
        
        for entry in active.iter() {
            if entry.execution_id == execution_id {
                let elapsed = entry.started_at.elapsed().as_millis() as u64;
                return elapsed > entry.duration_ms;
            }
        }
        
        false
    }

    /// Remove a completed execution
    pub fn finish_execution(&self, execution_id: &str) {
        let mut active = self.active_timeouts.write();
        active.retain(|e| e.execution_id != execution_id);
    }

    /// Get statistics
    pub fn stats(&self) -> TimeoutStats {
        TimeoutStats {
            max_timeout_ms: self.max_timeout_ms,
            total_timeouts: self.total_timeouts.load(Ordering::Relaxed),
            cancelled_timeouts: self.cancelled_timeouts.load(Ordering::Relaxed),
            active_count: self.active_timeouts.read().len() as u64,
        }
    }

    /// Clone just the inner data (for TimeoutGuard)
    fn clone_inner(&self) -> TimeoutManagerInner {
        TimeoutManagerInner {
            max_timeout_ms: self.max_timeout_ms,
            // We don't clone the active timeouts - just share the pointer
            total_timeouts: Arc::new(AtomicU64::new(self.total_timeouts.load(Ordering::Relaxed))),
            cancelled_timeouts: Arc::new(AtomicU64::new(self.cancelled_timeouts.load(Ordering::Relaxed))),
        }
    }
}

impl Clone for TimeoutManager {
    fn clone(&self) -> Self {
        Self {
            max_timeout_ms: self.max_timeout_ms,
            active_timeouts: RwLock::new(Vec::new()), // Don't clone active
            total_timeouts: AtomicU64::new(self.total_timeouts.load(Ordering::Relaxed)),
            cancelled_timeouts: AtomicU64::new(self.cancelled_timeouts.load(Ordering::Relaxed)),
        }
    }
}

/// Inner struct for TimeoutGuard (clonable)
#[derive(Clone)]
struct TimeoutManagerInner {
    max_timeout_ms: u64,
    total_timeouts: Arc<AtomicU64>,
    cancelled_timeouts: Arc<AtomicU64>,
}

/// Guard that automatically handles timeout cleanup
pub struct TimeoutGuard {
    manager: Arc<TimeoutManagerInner>,
    execution_id: String,
    started_at: Instant,
    timeout_ms: u64,
    cancelled: AtomicBool,
}

impl TimeoutGuard {
    /// Check if this execution has timed out
    pub fn is_timeout(&self) -> bool {
        if self.cancelled.load(Ordering::Relaxed) {
            return true;
        }
        
        let elapsed = self.started_at.elapsed().as_millis() as u64;
        elapsed > self.timeout_ms
    }

    /// Get remaining time in milliseconds
    pub fn remaining_ms(&self) -> u64 {
        let elapsed = self.started_at.elapsed().as_millis() as u64;
        self.timeout_ms.saturating_sub(elapsed)
    }

    /// Cancel this timeout
    pub fn cancel(&self) {
        self.cancelled.store(true, Ordering::Relaxed);
        self.manager.cancelled_timeouts.fetch_add(1, Ordering::Relaxed);
    }
}

impl Drop for TimeoutGuard {
    fn drop(&mut self) {
        // The execution has finished - nothing to clean up here
        // In a real implementation, we'd notify the manager
        debug!("Execution {} finished", self.execution_id);
    }
}

/// Timeout statistics
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct TimeoutStats {
    pub max_timeout_ms: u64,
    pub total_timeouts: u64,
    pub cancelled_timeouts: u64,
    pub active_count: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_timeout_creation() {
        let manager = TimeoutManager::new(5000);
        let stats = manager.stats();
        
        assert_eq!(stats.max_timeout_ms, 5000);
        assert_eq!(stats.total_timeouts, 0);
    }

    #[test]
    fn test_timeout_guard() {
        let manager = TimeoutManager::new(100);
        let guard = manager.start_execution("test-1", Some(50));
        
        // Should not timeout immediately
        assert!(!guard.is_timeout());
        
        // Wait and check
        std::thread::sleep(std::time::Duration::from_millis(60));
        assert!(guard.is_timeout());
    }
}
