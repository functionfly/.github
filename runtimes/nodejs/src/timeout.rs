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
    active_guards: AtomicU64,
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
            active_guards: AtomicU64::new(0),
        }
    }

    /// Start tracking a new execution
    /// The returned TimeoutGuard automatically cleans up when dropped,
    /// but for better control, call `guard.finish()` before dropping.
    pub fn start_execution(&self, execution_id: &str, timeout_ms: Option<u64>) -> TimeoutGuard {
        let timeout = timeout_ms.unwrap_or(self.max_timeout_ms);

        // Clone counters for the guard's inner
        let total_timeouts = Arc::new(AtomicU64::new(self.total_timeouts.load(Ordering::Relaxed)));
        let cancelled_timeouts = Arc::new(AtomicU64::new(self.cancelled_timeouts.load(Ordering::Relaxed)));
        let active_guards = Arc::new(AtomicU64::new(self.active_guards.load(Ordering::Relaxed)));

        // Increment counters
        self.total_timeouts.fetch_add(1, Ordering::Relaxed);
        self.active_guards.fetch_add(1, Ordering::Relaxed);

        // Create a callback for cleanup that the guard will call on drop
        let exec_id = execution_id.to_string();
        let manager = Arc::new(self.clone_inner());
        let cleanup_callback = Arc::new(move || {
            // Remove from active_timeouts and decrement active_guards
            let mut active = manager.active_timeouts.write();
            active.retain(|e| e.execution_id != exec_id);
            manager.active_guards.fetch_sub(1, Ordering::Relaxed);
        });

        TimeoutGuard {
            inner: Arc::new(TimeoutGuardInner {
                max_timeout_ms: self.max_timeout_ms,
                execution_id: execution_id.to_string(),
                started_at: Instant::now(),
                timeout_ms: timeout,
                cancelled: AtomicBool::new(false),
                total_timeouts,
                cancelled_timeouts,
                active_guards,
                cleanup_callback: Some(cleanup_callback),
            }),
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
        self.active_guards.fetch_sub(1, Ordering::Relaxed);
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
            active_timeouts: RwLock::new(Vec::new()), // Empty in cloned inner
            total_timeouts: Arc::new(AtomicU64::new(self.total_timeouts.load(Ordering::Relaxed))),
            cancelled_timeouts: Arc::new(AtomicU64::new(self.cancelled_timeouts.load(Ordering::Relaxed))),
            active_guards: Arc::new(AtomicU64::new(self.active_guards.load(Ordering::Relaxed))),
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
            active_guards: AtomicU64::new(self.active_guards.load(Ordering::Relaxed)),
        }
    }
}

/// Inner struct shared between TimeoutManager and TimeoutGuard
struct TimeoutManagerInner {
    max_timeout_ms: u64,
    active_timeouts: RwLock<Vec<TimeoutEntry>>,
    total_timeouts: Arc<AtomicU64>,
    cancelled_timeouts: Arc<AtomicU64>,
    active_guards: Arc<AtomicU64>,
}

/// Inner data for TimeoutGuard
struct TimeoutGuardInner {
    max_timeout_ms: u64,
    execution_id: String,
    started_at: Instant,
    timeout_ms: u64,
    cancelled: AtomicBool,
    total_timeouts: Arc<AtomicU64>,
    cancelled_timeouts: Arc<AtomicU64>,
    active_guards: Arc<AtomicU64>,
    cleanup_callback: Option<Arc<dyn Fn() + Send + Sync>>,
}

/// Guard that automatically handles timeout cleanup
pub struct TimeoutGuard {
    inner: Arc<TimeoutGuardInner>,
}

impl TimeoutGuard {
    /// Check if this execution has timed out
    pub fn is_timeout(&self) -> bool {
        if self.inner.cancelled.load(Ordering::Relaxed) {
            return true;
        }

        let elapsed = self.inner.started_at.elapsed().as_millis() as u64;
        elapsed > self.inner.timeout_ms
    }

    /// Get remaining time in milliseconds
    pub fn remaining_ms(&self) -> u64 {
        let elapsed = self.inner.started_at.elapsed().as_millis() as u64;
        self.inner.timeout_ms.saturating_sub(elapsed)
    }

    /// Cancel this timeout
    pub fn cancel(&self) {
        self.inner.cancelled.store(true, Ordering::Relaxed);
        self.inner.cancelled_timeouts.fetch_add(1, Ordering::Relaxed);
        self.inner.active_guards.fetch_sub(1, Ordering::Relaxed);
        // Invoke cleanup to remove from active_timeouts
        if let Some(ref cb) = self.inner.cleanup_callback {
            cb();
        }
    }

    /// Mark execution as finished and clean up manager tracking
    /// Call this when execution completes normally (before dropping the guard)
    pub fn finish(mut self) {
        // Invoke cleanup to remove from active_timeouts
        if let Some(ref cb) = self.inner.cleanup_callback {
            cb();
        }
        debug!("Execution {} finished", self.inner.execution_id);
    }
}

impl Drop for TimeoutGuard {
    fn drop(&mut self) {
        // Only invoke cleanup if not cancelled (cancel handles its own cleanup)
        if !self.inner.cancelled.load(Ordering::Relaxed) {
            self.inner.active_guards.fetch_sub(1, Ordering::Relaxed);
            if let Some(ref cb) = self.inner.cleanup_callback {
                cb();
            }
        }
        debug!("TimeoutGuard for {} dropped", self.inner.execution_id);
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

    #[test]
    fn test_finish_execution() {
        let manager = TimeoutManager::new(5000);

        let guard1 = manager.start_execution("exec-1", None);
        let guard2 = manager.start_execution("exec-2", None);

        let stats = manager.stats();
        assert_eq!(stats.active_count, 2);

        // Call finish on guard1
        guard1.finish();

        let stats = manager.stats();
        assert_eq!(stats.active_count, 1);

        // Drop guard2 without calling finish
        drop(guard2);

        let stats = manager.stats();
        assert_eq!(stats.active_count, 0);
    }
}
