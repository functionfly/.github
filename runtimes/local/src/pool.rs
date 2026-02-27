//! Instance pool for reusing warm Wasm instances with memory optimization.

use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use tokio::time::{interval, Duration as TokioDuration};

use crate::errors::{RuntimeError, RuntimeResult};
use crate::logging::{CorrelationId, StructuredLogger};

/// Pool of warm Wasm instances with memory optimization
pub struct InstancePool {
    /// Pooled instances per function key
    instances: HashMap<String, VecDeque<PooledInstance>>,
    /// Maximum instances per function
    max_per_function: usize,
    /// Idle timeout before recycling
    idle_timeout: Duration,
    /// Maximum total instances
    max_total: usize,
    /// Memory pressure threshold (percentage of system memory)
    memory_pressure_threshold: f64,
    /// Current total memory usage estimate
    current_memory_usage: usize,
    /// Maximum memory usage allowed
    max_memory_usage: usize,
    /// Instance reuse limit before forced recycling
    max_reuse_count: u32,
    /// Background pruning task handle
    _pruning_task: Option<Arc<tokio::task::JoinHandle<()>>>,
    /// Logger for structured logging
    logger: Option<Arc<StructuredLogger>>,
}

/// A pooled Wasm instance with memory tracking
#[derive(Clone)]
struct PooledInstance {
    /// When the instance was created
    created_at: Instant,
    /// When the instance was last used
    last_used: Instant,
    /// Instance ID for tracking
    instance_id: String,
    /// Estimated memory usage in bytes
    memory_usage: usize,
    /// Number of times this instance has been reused
    reuse_count: u32,
    /// Function key this instance is associated with
    function_key: String,
}

impl InstancePool {
    /// Create a new instance pool with memory optimization
    pub fn new(max_per_function: usize, idle_timeout_secs: u64) -> Self {
        Self::with_memory_limits(
            max_per_function,
            idle_timeout_secs,
            128 * 1024 * 1024, // 128MB default memory limit
            80.0, // 80% memory pressure threshold
        )
    }

    /// Create instance pool with explicit memory limits
    pub fn with_memory_limits(
        max_per_function: usize,
        idle_timeout_secs: u64,
        max_memory_mb: usize,
        memory_pressure_threshold: f64,
    ) -> Self {
        Self {
            instances: HashMap::new(),
            max_per_function,
            idle_timeout: Duration::from_secs(idle_timeout_secs),
            max_total: max_per_function * 10, // Allow some overflow
            memory_pressure_threshold,
            current_memory_usage: 0,
            max_memory_usage: max_memory_mb * 1024 * 1024, // Convert MB to bytes
            max_reuse_count: 100, // Default reuse limit
            _pruning_task: None,
            logger: None,
        }
    }

    /// Create instance pool with logger
    pub fn with_logger(mut self, logger: Arc<StructuredLogger>) -> Self {
        self.logger = Some(logger);
        self
    }

    /// Start background pruning task
    pub fn start_background_pruning(&mut self) {
        let pool = Arc::new(RwLock::new(self.clone()));
        let pruning_task = tokio::spawn(async move {
            let mut interval = interval(TokioDuration::from_secs(60)); // Prune every minute

            loop {
                interval.tick().await;
                let mut pool_guard = pool.write().await;

        // Extract logger reference to avoid borrow checker issues
        let logger_option = pool_guard.logger.clone();

        if let Some(logger) = logger_option {
            let correlation_id = logger.generate_correlation_id().await;
            let pruned = pool_guard.prune_with_memory_optimization(&correlation_id).await;
            if pruned > 0 {
                let stats = pool_guard.stats();
                logger.log_pool_stats(
                    &correlation_id,
                    stats.total_instances,
                    stats.functions_in_pool,
                    pruned,
                );
            }
        } else {
            let _ = pool_guard.prune_with_memory_optimization_simple().await;
        }
            }
        });

        // Store the task handle for proper lifecycle management
        self._pruning_task = Some(Arc::new(pruning_task));
    }

    /// Get an instance from the pool, if available
    pub async fn get(&mut self, function_key: &str) -> RuntimeResult<Option<PooledInstance>> {
        if let Some(queue) = self.instances.get_mut(function_key) {
        if let Some(mut instance) = queue.pop_front() {
            // Check if instance should be recycled due to reuse limit BEFORE using it
            if instance.reuse_count >= self.max_reuse_count {
                tracing::debug!(
                    "Instance {} reached reuse limit ({}), recycling",
                    instance.instance_id,
                    self.max_reuse_count
                );
                // Don't return this instance, it will be dropped
                return Ok(None);
            }

            // Update last used time and reuse count
            instance.last_used = Instant::now();
            instance.reuse_count += 1;

                // Update memory tracking
                self.current_memory_usage -= instance.memory_usage;

                if let Some(ref logger) = self.logger {
                    let correlation_id = logger.generate_correlation_id().await;
                    logger.log_with_correlation(
                        crate::logging::LogLevel::Debug,
                        format!("Retrieved instance {} from pool for {}", instance.instance_id, function_key),
                        &correlation_id,
                    );
                } else {
                    tracing::debug!(
                        "Got instance {} from pool for {}",
                        instance.instance_id,
                        function_key
                    );
                }

                return Ok(Some(instance));
            }
        }
        Ok(None)
    }

    /// Return an instance to the pool with memory optimization
    pub async fn return_instance(&mut self, function_key: String, mut instance: PooledInstance) -> RuntimeResult<()> {
        // Update instance metadata
        instance.function_key = function_key.clone();

        // Check memory pressure first
        if self.is_under_memory_pressure() {
            if let Some(ref logger) = self.logger {
                let correlation_id = logger.generate_correlation_id().await;
                logger.log_with_correlation(
                    crate::logging::LogLevel::Warn,
                    "Memory pressure detected, not returning instance to pool",
                    &correlation_id,
                );
            }
            return Ok(());
        }

        // Check if we've reached the limit for this function
        let queue = self.instances.entry(function_key.clone()).or_default();

        if queue.len() >= self.max_per_function {
            if let Some(ref logger) = self.logger {
                let correlation_id = logger.generate_correlation_id().await;
                logger.log_with_correlation(
                    crate::logging::LogLevel::Debug,
                    format!("Pool full for function {}, discarding instance", function_key),
                    &correlation_id,
                );
            } else {
                tracing::debug!("Pool full for function, discarding instance");
            }
            return Ok(());
        }

        // Check if instance is still valid (not too old)
        if instance.last_used.elapsed() > self.idle_timeout {
            if let Some(ref logger) = self.logger {
                let correlation_id = logger.generate_correlation_id().await;
                logger.log_with_correlation(
                    crate::logging::LogLevel::Debug,
                    "Instance expired, not returning to pool",
                    &correlation_id,
                );
            } else {
                tracing::debug!("Instance expired, not returning to pool");
            }
            return Ok(());
        }

        // Update memory tracking
        self.current_memory_usage += instance.memory_usage;

        queue.push_back(instance);

        if let Some(ref logger) = self.logger {
            let correlation_id = logger.generate_correlation_id().await;
            logger.log_with_correlation(
                crate::logging::LogLevel::Debug,
                format!("Returned instance to pool for {}", function_key),
                &correlation_id,
            );
        } else {
            tracing::debug!("Returned instance to pool for {}", function_key);
        }

        Ok(())
    }


    /// Advanced pruning with memory optimization
    pub async fn prune_with_memory_optimization(&mut self, correlation_id: &CorrelationId) -> usize {
        let mut removed = 0;
        let mut memory_freed = 0;
        let has_logger = self.logger.is_some();

        // Prune expired instances
        for (function_key, queue) in self.instances.iter_mut() {
            let original_len = queue.len();
            let original_memory = queue.iter().map(|i| i.memory_usage).sum::<usize>();

            // Collect instances to log before modifying the queue
            let mut instances_to_log = Vec::new();
            if has_logger {
                for instance in queue.iter() {
                    let is_expired = instance.last_used.elapsed() > self.idle_timeout;
                    let is_over_reused = instance.reuse_count >= self.max_reuse_count;
                    if is_expired || is_over_reused {
                        instances_to_log.push((instance.instance_id.clone(), is_expired, is_over_reused));
                    }
                }
            }

            // Remove expired instances and those exceeding reuse limits
            queue.retain(|instance| {
                let is_expired = instance.last_used.elapsed() > self.idle_timeout;
                let is_over_reused = instance.reuse_count >= self.max_reuse_count;
                !(is_expired || is_over_reused)
            });

            let new_len = queue.len();
            let new_memory = queue.iter().map(|i| i.memory_usage).sum::<usize>();
            removed += original_len - new_len;
            memory_freed += original_memory - new_memory;

            // Log the pruning after modifying the queue
            if has_logger && original_len != new_len {
                if let Some(ref logger) = self.logger {
                    // Log individual instance pruning
                    for (instance_id, is_expired, is_over_reused) in instances_to_log {
                        logger.log_with_correlation(
                            crate::logging::LogLevel::Debug,
                            format!(
                                "Pruning instance {}: expired={}, over_reused={}",
                                instance_id, is_expired, is_over_reused
                            ),
                            correlation_id,
                        );
                    }

                    // Log summary for this function
                    logger.log_with_correlation(
                        crate::logging::LogLevel::Info,
                        format!("Pruned {} instances for function {}", original_len - new_len, function_key),
                        correlation_id,
                    );
                }
            }
        }

        // Clean up empty queues
        self.instances.retain(|_, queue| !queue.is_empty());

        // Update memory tracking
        self.current_memory_usage -= memory_freed;

        // If still under memory pressure, prune least recently used instances
        if self.is_under_memory_pressure() {
            removed += self.prune_lru_instances(correlation_id).await;
        }

        if removed > 0 {
            if let Some(ref logger) = self.logger {
                logger.log_with_correlation(
                    crate::logging::LogLevel::Info,
                    format!("Pruned {} total instances, freed {:.2}MB", removed, memory_freed as f64 / 1024.0 / 1024.0),
                    correlation_id,
                );
            } else {
                tracing::info!("Pruned {} expired instances", removed);
            }
        }

        removed
    }

    /// Simple pruning for backward compatibility (async version)
    pub async fn prune_with_memory_optimization_simple(&mut self) -> usize {
        let correlation_id = if let Some(ref logger) = self.logger {
            logger.generate_correlation_id().await
        } else {
            CorrelationId::new("prune_simple".to_string())
        };

        self.prune_with_memory_optimization(&correlation_id).await
    }

    /// Prune least recently used instances when under memory pressure
    async fn prune_lru_instances(&mut self, correlation_id: &CorrelationId) -> usize {
        let mut removed = 0;
        let mut memory_freed = 0;

        // Calculate target memory usage (70% of max)
        let target_memory = (self.max_memory_usage as f64 * 0.7) as usize;
        let mut instances_to_remove = Vec::new();

        // Collect all instances with their last used time
        for (function_key, queue) in &self.instances {
            for instance in queue {
                instances_to_remove.push((
                    function_key.clone(),
                    instance.instance_id.clone(),
                    instance.last_used,
                    instance.memory_usage,
                ));
            }
        }

        // Sort by last used time (oldest first)
        instances_to_remove.sort_by(|a, b| a.2.cmp(&b.2));

        // Remove oldest instances until we're below target memory
        for (function_key, instance_id, _, mem_usage) in instances_to_remove {
            if self.current_memory_usage <= target_memory {
                break;
            }

            if let Some(queue) = self.instances.get_mut(&function_key) {
                queue.retain(|instance| instance.instance_id != instance_id);

                self.current_memory_usage -= mem_usage;
                memory_freed += mem_usage;
                removed += 1;

                if let Some(ref logger) = self.logger {
                    logger.log_with_correlation(
                        crate::logging::LogLevel::Warn,
                        format!("LRU pruning instance {} from function {}", instance_id, function_key),
                        correlation_id,
                    );
                }
            }
        }

        // Clean up empty queues
        self.instances.retain(|_, queue| !queue.is_empty());

        if memory_freed > 0 {
            if let Some(ref logger) = self.logger {
                logger.log_with_correlation(
                    crate::logging::LogLevel::Warn,
                    format!("LRU pruning freed {:.2}MB memory", memory_freed as f64 / 1024.0 / 1024.0),
                    correlation_id,
                );
            }
        }

        removed
    }

    /// Check if pool is under memory pressure
    fn is_under_memory_pressure(&self) -> bool {
        let memory_usage_percent = (self.current_memory_usage as f64 / self.max_memory_usage as f64) * 100.0;
        memory_usage_percent >= self.memory_pressure_threshold
    }

    /// Get pool statistics with memory information
    pub fn stats(&self) -> PoolStats {
        let total_instances: usize = self.instances.values().map(|q| q.len()).sum();
        let memory_usage_mb = self.current_memory_usage as f64 / 1024.0 / 1024.0;
        let max_memory_mb = self.max_memory_usage as f64 / 1024.0 / 1024.0;
        let memory_pressure_percent = (self.current_memory_usage as f64 / self.max_memory_usage as f64) * 100.0;

        PoolStats {
            total_instances,
            functions_in_pool: self.instances.len(),
            max_per_function: self.max_per_function,
            idle_timeout_secs: self.idle_timeout.as_secs(),
            current_memory_usage_mb: memory_usage_mb,
            max_memory_usage_mb: max_memory_mb,
            memory_pressure_percent,
        }
    }

    /// Clear all instances
    pub fn clear(&mut self) {
        self.instances.clear();
        tracing::info!("Cleared instance pool");
    }

    /// Check if pool has capacity for new instance (considering memory limits)
    pub fn has_capacity(&self, estimated_memory_usage: usize) -> bool {
        let total_instances: usize = self.instances.values().map(|q| q.len()).sum();

        // Check instance count limit
        if total_instances >= self.max_total {
            return false;
        }

        // Check memory limit
        if self.current_memory_usage + estimated_memory_usage > self.max_memory_usage {
            return false;
        }

        true
    }

    /// Legacy capacity check (for backward compatibility)
    pub fn has_capacity_legacy(&self) -> bool {
        self.has_capacity(0) // Assume no memory usage for legacy calls
    }

    /// Create a new pooled instance for testing
    #[cfg(test)]
    pub fn create_test_instance(instance_id: &str, memory_usage: usize) -> PooledInstance {
        PooledInstance {
            created_at: Instant::now(),
            last_used: Instant::now(),
            instance_id: instance_id.to_string(),
            memory_usage,
            reuse_count: 0,
            function_key: "test".to_string(),
        }
    }
}

impl Clone for InstancePool {
    fn clone(&self) -> Self {
        Self {
            instances: self.instances.clone(),
            max_per_function: self.max_per_function,
            idle_timeout: self.idle_timeout,
            max_total: self.max_total,
            memory_pressure_threshold: self.memory_pressure_threshold,
            current_memory_usage: self.current_memory_usage,
            max_memory_usage: self.max_memory_usage,
            max_reuse_count: self.max_reuse_count,
            _pruning_task: None, // Don't clone the task handle
            logger: self.logger.clone(),
        }
    }
}

/// Pool statistics
#[derive(Debug, Clone)]
pub struct PoolStats {
    pub total_instances: usize,
    pub functions_in_pool: usize,
    pub max_per_function: usize,
    pub idle_timeout_secs: u64,
    pub current_memory_usage_mb: f64,
    pub max_memory_usage_mb: f64,
    pub memory_pressure_percent: f64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_pool_basic() {
        let mut pool = InstancePool::new(5, 60);

        // Should be empty initially
        let instance = pool.get("test@1.0.0").await.unwrap();
        assert!(instance.is_none());

        // Return an instance
        pool.return_instance(
            "test@1.0.0".to_string(),
            InstancePool::create_test_instance("test-1", 1024 * 1024), // 1MB
        ).await.unwrap();

        // Should get it back
        let instance = pool.get("test@1.0.0").await.unwrap();
        assert!(instance.is_some());
        assert_eq!(instance.unwrap().instance_id, "test-1");
    }

    #[tokio::test]
    async fn test_pool_limit() {
        let mut pool = InstancePool::new(2, 60);

        // Add 3 instances
        for i in 0..3 {
            pool.return_instance(
                "test@1.0.0".to_string(),
                InstancePool::create_test_instance(&format!("test-{}", i), 1024 * 1024), // 1MB each
            ).await.unwrap();
        }

        // Should only have 2
        let stats = pool.stats();
        assert_eq!(stats.total_instances, 2);
    }

    #[tokio::test]
    async fn test_memory_pressure() {
        // Create pool with very low memory limit (1MB)
        let mut pool = InstancePool::with_memory_limits(10, 60, 1, 80.0);

        // Add instance that uses 0.5MB
        pool.return_instance(
            "test@1.0.0".to_string(),
            InstancePool::create_test_instance("test-1", 512 * 1024), // 0.5MB
        ).await.unwrap();

        // Should have capacity for another 0.5MB instance
        assert!(pool.has_capacity(512 * 1024));

        // But not for a 1MB instance
        assert!(!pool.has_capacity(1024 * 1024));

        let stats = pool.stats();
        assert!(stats.current_memory_usage_mb > 0.0);
    }

    #[tokio::test]
    async fn test_reuse_limit() {
        let mut pool = InstancePool::new(5, 60);
        pool.max_reuse_count = 2; // Set low reuse limit for testing

        let mut instance = InstancePool::create_test_instance("test-1", 1024);
        instance.reuse_count = 1; // One use already

        // Return instance (reuse count will be checked on get)
        pool.return_instance("test@1.0.0".to_string(), instance).await.unwrap();

        // First get should succeed
        let instance = pool.get("test@1.0.0").await.unwrap();
        assert!(instance.is_some());

        // Return it again
        pool.return_instance("test@1.0.0".to_string(), instance.unwrap()).await.unwrap();

        // Second get should return None due to reuse limit
        let instance = pool.get("test@1.0.0").await.unwrap();
        assert!(instance.is_none());
    }
}
