//! Metrics Collection
//! 
//! This module provides metrics collection for the Node.js runtime.

use std::sync::atomic::{AtomicU64, Ordering};

/// Executor metrics
pub struct ExecutorMetrics {
    /// Total number of executions
    pub total_executions: AtomicU64,
    
    /// Number of successful executions
    pub successful_executions: AtomicU64,
    
    /// Number of failed executions
    pub errors: AtomicU64,
    
    /// Number of timeouts
    pub timeouts: AtomicU64,
    
    /// Number of panics
    pub panics: AtomicU64,
    
    /// Cache hits
    pub cache_hits: AtomicU64,
    
    /// Cache misses
    pub cache_misses: AtomicU64,
    
    /// Total execution time (nanoseconds)
    pub total_execution_time_ns: AtomicU64,
}

impl ExecutorMetrics {
    /// Create new metrics
    pub fn new() -> Self {
        Self {
            total_executions: AtomicU64::new(0),
            successful_executions: AtomicU64::new(0),
            errors: AtomicU64::new(0),
            timeouts: AtomicU64::new(0),
            panics: AtomicU64::new(0),
            cache_hits: AtomicU64::new(0),
            cache_misses: AtomicU64::new(0),
            total_execution_time_ns: AtomicU64::new(0),
        }
    }

    /// Record execution time
    pub fn execution_time(&self, ns: u64) {
        self.total_execution_time_ns.fetch_add(ns, Ordering::Relaxed);
    }

    /// Get metrics snapshot
    pub fn snapshot(&self) -> ExecutorMetricsSnapshot {
        ExecutorMetricsSnapshot {
            total_executions: self.total_executions.load(Ordering::Relaxed),
            successful_executions: self.successful_executions.load(Ordering::Relaxed),
            errors: self.errors.load(Ordering::Relaxed),
            timeouts: self.timeouts.load(Ordering::Relaxed),
            panics: self.panics.load(Ordering::Relaxed),
            cache_hits: self.cache_hits.load(Ordering::Relaxed),
            cache_misses: self.cache_misses.load(Ordering::Relaxed),
            total_execution_time_ns: self.total_execution_time_ns.load(Ordering::Relaxed),
        }
    }
}

impl Default for ExecutorMetrics {
    fn default() -> Self {
        Self::new()
    }
}

/// Metrics snapshot
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ExecutorMetricsSnapshot {
    pub total_executions: u64,
    pub successful_executions: u64,
    pub errors: u64,
    pub timeouts: u64,
    pub panics: u64,
    pub cache_hits: u64,
    pub cache_misses: u64,
    pub total_execution_time_ns: u64,
}

impl ExecutorMetricsSnapshot {
    /// Get success rate as a percentage
    pub fn success_rate(&self) -> f64 {
        if self.total_executions == 0 {
            return 100.0;
        }
        
        let successful = self.total_executions 
            - self.errors.load(Ordering::Relaxed) 
            - self.timeouts.load(Ordering::Relaxed);
        
        (successful as f64 / self.total_executions as f64) * 100.0
    }

    /// Get average execution time in milliseconds
    pub fn avg_execution_time_ms(&self) -> f64 {
        if self.total_executions == 0 {
            return 0.0;
        }
        
        (self.total_execution_time_ns as f64 / 1_000_000.0) / self.total_executions as f64
    }

    /// Get cache hit rate as a percentage
    pub fn cache_hit_rate(&self) -> f64 {
        let total = self.cache_hits + self.cache_misses;
        if total == 0 {
            return 0.0;
        }
        
        (self.cache_hits as f64 / total as f64) * 100.0
    }
}
