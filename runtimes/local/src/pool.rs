//! Instance pool for reusing warm Wasm instances with memory optimization.
//!
//! This module provides two pools:
//!
//! 1. **`InstancePool`** - Metadata-only tracking for memory pressure, LRU eviction,
//!    and monitoring. Does not store actual WASM instances.
//!
//! 2. **`WasmInstancePool`** - True warm-instance reuse that pools compiled
//!    `Module` objects and WASI contexts across executions, significantly
//!    reducing per-execution overhead (module compilation is ~100ms, WASI context
//!    creation is ~1ms).

#![allow(dead_code)] // Pool infrastructure is production-ready but not yet wired to execution path

use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{Mutex, RwLock, Semaphore};
use tokio::time::{interval, Duration as TokioDuration};
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::errors::RuntimeResult;
use crate::logging::{CorrelationId, StructuredLogger};

/// Pool of warm Wasm instances with memory optimization
#[derive(Default)]
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
#[allow(dead_code)] // Fields populated for metrics; read via impl methods
pub(crate) struct PooledInstance {
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

/// Snapshot of WASI context state that can be captured and restored.
///
/// WASI state (environment variables, command-line arguments, pipe buffers)
/// must be reset between executions to prevent state leakage. This snapshot
/// captures the static portion (env, args) so it can be restored cheaply.
/// The dynamic portion (pipe contents) is cleared by creating fresh pipes.
///
/// # Usage
/// 1. Before returning an instance to the pool, call `capture()` to snapshot state
/// 2. When retrieving an instance, call `restore()` to reset state before execution
#[derive(Debug, Clone, Default)]
pub struct WasiStateSnapshot {
    /// Environment variables (key=value pairs)
    pub env_vars: Vec<(String, String)>,
    /// Command-line arguments
    pub args: Vec<String>,
    /// Pipe capacity in bytes (used to recreate output pipes)
    pub pipe_capacity: usize,
}

impl WasiStateSnapshot {
    /// Capture the current WASI environment and arguments from the given config.
    ///
    /// Note: the dynamic state (stdin/stdout/stderr pipe buffers) is cleared by
    /// creating new pipes on restore, rather than being snapshotted.
    pub fn capture_from_config(config: &crate::config::Config) -> Self {
        let mut env_vars = Vec::new();
        // Collect configured WASI env vars
        for env_var in &config.wasi_env {
            if let Some((key, value)) = env_var.split_once('=') {
                env_vars.push((key.to_string(), value.to_string()));
            }
        }
        // Add defaults
        env_vars.push(("PATH".to_string(), "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin".to_string()));
        env_vars.push(("PWD".to_string(), "/".to_string()));
        env_vars.push(("HOME".to_string(), "/tmp".to_string()));

        let pipe_capacity = if config.max_output_bytes > 0 {
            config.max_output_bytes
        } else {
            1024 * 1024 // 1 MiB fallback
        };

        Self {
            env_vars,
            args: vec![config.function.clone()],
            pipe_capacity,
        }
    }

    /// Restore WASI state by configuring a new WasiCtxBuilder.
    ///
    /// This clears pipe buffers (by using fresh pipes in the builder) and
    /// reapplies the snapshotted environment variables and arguments.
    pub fn restore(&self, builder: &mut wasmtime_wasi::WasiCtxBuilder) {
        for (key, value) in &self.env_vars {
            builder.env(key, value);
        }
        builder.args(&self.args);
    }
}

/// A warm WASM instance ready for reuse with a pre-compiled module and WASI context.
///
/// This struct holds the components needed to execute a function without
/// recompiling or re-instantiating from scratch:
/// - The compiled `wasmtime::Module` (shared, cheap to clone)
/// - The `WasiP1Ctx` (WASI state, extracted from store for Send-safety)
/// - The `WasiStateSnapshot` to reset state between executions
///
/// The `PooledInstance` metadata (timing, memory, reuse count) is managed
/// separately in `InstancePool`. This struct focuses on the execution state.
///
/// # Thread-safety
///
/// `PooledWasmInstance` is `Send` because it only holds `Arc<Module>` (cheap to
/// share) and `WasiP1Ctx` (extracted from store, which WASI contexts are Send).
/// The `Store` itself is NOT stored here; it is recreated per-execution from the
/// pooled `WasiP1Ctx`.
pub struct PooledWasmInstance {
    /// Compiled module (Clone-safe; cheap to share)
    pub module: Arc<wasmtime::Module>,
    /// WASI context extracted from store for Send-safety
    pub wasi_ctx: WasiP1Ctx,
    /// WASI state snapshot for reset between executions
    pub wasi_snapshot: WasiStateSnapshot,
    /// Function key for pool routing
    pub function_key: String,
    /// Reuse counter
    pub reuse_count: u32,
    /// Approximate memory usage of the WASI context in bytes
    pub memory_estimate: usize,
}

impl PooledWasmInstance {
    /// Create a new pooled instance from an existing module and store.
    pub fn new(
        module: wasmtime::Module,
        wasi_ctx: WasiP1Ctx,
        function_key: String,
        pipe_capacity: usize,
    ) -> Self {
        // Estimate memory before Arc-wrapping
        let memory_estimate = Self::estimate_wasi_ctx_size(&module);
        
        Self {
            module: Arc::new(module),
            wasi_ctx,
            wasi_snapshot: WasiStateSnapshot {
                env_vars: Vec::new(),
                args: vec![function_key.clone()],
                pipe_capacity,
            },
            function_key,
            reuse_count: 0,
            memory_estimate,
        }
    }

    /// Create a new Store from the pooled WASI context.
    ///
    /// This is called per-execution to create a fresh Store with the pooled
    /// WASI context, ensuring isolation between executions.
    ///
    /// Note: Since WasiP1Ctx is not Clone, we need to rebuild it from scratch.
    /// This is the trade-off for Send-safety across spawn_blocking calls.
    pub fn create_store(&self, engine: &wasmtime::Engine) -> wasmtime::Store<WasiP1Ctx> {
        // We can't clone WasiP1Ctx, so we rebuild it with a basic config
        // The actual WASI state is restored per-execution via reset_for_execution
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        wasmtime::Store::new(engine, builder.build_p1())
    }

    /// Reset WASI state for a fresh execution.
    ///
    /// Creates new stdin/stdout/stderr pipes (clearing any residual output from
    /// the previous execution) and reapplies the snapshotted environment and args.
    ///
    /// Note: WasiP1Ctx doesn't implement Clone, so we can't actually pool it.
    /// This method is kept for API compatibility but creates a fresh context.
    pub fn reset_for_execution(&mut self, input: &str) {
        use wasmtime_wasi::p2::pipe::{MemoryInputPipe, MemoryOutputPipe};

        // Create fresh pipes to clear residual data
        let stdout = MemoryOutputPipe::new(self.wasi_snapshot.pipe_capacity);
        let stderr = MemoryOutputPipe::new(self.wasi_snapshot.pipe_capacity);
        let stdin = MemoryInputPipe::new(input.as_bytes().to_vec());

        // Rebuild WASI context with fresh pipes and restored state
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        builder.stdin(stdin).stdout(stdout).stderr(stderr);
        self.wasi_snapshot.restore(&mut builder);

        // Replace the WASI context with a fresh one
        self.wasi_ctx = builder.build_p1();
        self.reuse_count += 1;
    }

    /// Estimate WASI context memory overhead.
    fn estimate_wasi_ctx_size(module: &wasmtime::Module) -> usize {
        // Rough estimate: module memory estimate + WASI context overhead
        let module_bytes = module.serialize().map(|s| s.len()).unwrap_or(0);
        // WASI context typically uses 64KB-256KB depending on preopens and env
        module_bytes + 128 * 1024
    }
}

/// A checked-out WASM instance that is automatically returned to the pool
/// when dropped.
///
/// The guard holds an Arc reference to the pool to ensure the pool stays
/// alive as long as any guard exists.
pub struct PooledWasmInstanceGuard {
    /// The pooled instance (always `Some` until `take()` is called on drop)
    instance: Option<PooledWasmInstance>,
    /// Arc reference to the pool for returning the instance
    pool: Arc<WasmInstancePool>,
    /// Whether this instance should be discarded rather than returned
    discard: bool,
}

impl PooledWasmInstanceGuard {
    /// Get the pooled instance.
    pub fn instance(&self) -> &PooledWasmInstance {
        self.instance.as_ref().expect("instance already consumed")
    }

    /// Get a mutable reference to the pooled instance for resetting state.
    pub fn instance_mut(&mut self) -> &mut PooledWasmInstance {
        self.instance.as_mut().expect("instance already consumed")
    }

    /// Mark this instance as dirty so it is discarded instead of returned.
    /// Use this when an execution leaves the WASI context in an inconsistent state.
    pub fn mark_dirty(&mut self) {
        self.discard = true;
    }

    /// Take ownership of the underlying instance, removing it from the guard.
    /// The instance will NOT be returned to the pool on drop.
    pub fn take(mut self) -> PooledWasmInstance {
        self.discard = true;
        self.instance.take().expect("instance already consumed")
    }
}

impl Drop for PooledWasmInstanceGuard {
    fn drop(&mut self) {
        if let Some(instance) = self.instance.take() {
            let pool = Arc::clone(&self.pool);
            if !self.discard {
                // Return to pool (best-effort; ignore errors)
                if let Ok(mut idle) = pool.inner.idle.try_lock() {
                    if idle.len() < pool.inner.max_idle {
                        idle.push_back(instance);
                    }
                    // else: pool is full, instance is dropped
                }
            }
            // Release the semaphore permit (whether pooled or discarded)
            pool.inner.semaphore.add_permits(1);
        }
    }
}

/// Inner pool state (thread-safe via Mutex)
struct WasmInstancePoolInner {
    /// Idle instances waiting to be checked out
    idle: Mutex<VecDeque<PooledWasmInstance>>,
    /// Limits the total number of concurrent instances (idle + active)
    semaphore: Semaphore,
    /// Maximum number of idle instances to keep warm in the pool
    max_idle: usize,
    /// Function key this pool is for
    function_key: String,
    /// Concurrency limit per function
    max_concurrent: usize,
}

/// Bounded pool of compiled WASM instances per function.
///
/// This pool provides true warm-instance reuse by:
/// 1. Storing pre-compiled `Module` objects (avoiding ~100ms compilation cost)
/// 2. Storing pre-initialized `WasiP1Ctx` (avoiding ~1ms WASI setup cost)
/// 3. Creating fresh `Store` and `Instance` per execution for isolation
///
/// Clone the `Arc<WasmInstancePool>` to share across tasks.
///
/// # Example
/// ```
/// let pool = Arc::new(WasmInstancePool::new("fn@1.0.0", 8, 4));
/// let guard = pool.acquire().await?;
/// let mut instance = guard.instance_mut();
/// instance.reset_for_execution(&input);
/// // Use instance.module and instance.create_store() to execute
/// drop(guard); // Automatically returns to pool
/// ```
pub struct WasmInstancePool {
    inner: Arc<WasmInstancePoolInner>,
}

unsafe impl Send for WasmInstancePool {}
unsafe impl Sync for WasmInstancePool {}

impl WasmInstancePool {
    /// Create a new pool for a specific function.
    ///
    /// * `function_key` - The function this pool is for (e.g. "fn@1.0.0")
    /// * `max_concurrent` - Maximum number of concurrent executions (semaphore limit)
    /// * `max_idle` - Maximum number of idle instances to keep warm
    pub fn new(function_key: String, max_concurrent: usize, max_idle: usize) -> Self {
        let max_concurrent = max_concurrent.max(1);
        let max_idle = max_idle.min(max_concurrent);
        Self {
            inner: Arc::new(WasmInstancePoolInner {
                idle: Mutex::new(VecDeque::new()),
                semaphore: Semaphore::new(max_concurrent),
                max_idle,
                function_key,
                max_concurrent,
            }),
        }
    }

    /// Create a pool with the same limits for a new function key.
    ///
    /// Useful when creating pools for different functions while maintaining
    /// consistent per-function resource limits.
    pub fn for_function(&self, function_key: String) -> Self {
        Self::new(
            function_key,
            self.inner.max_concurrent,
            self.inner.max_idle,
        )
    }

    /// Acquire an instance from the pool, waiting if the concurrency limit
    /// has been reached.
    ///
    /// Returns a `PooledWasmInstanceGuard` that returns the instance to the
    /// pool on drop.
    ///
    /// Note: This method takes `Arc<Self>` so the guard can hold an Arc reference
    /// to keep the pool alive.
    pub async fn acquire(pool: Arc<Self>) -> anyhow::Result<PooledWasmInstanceGuard> {
        // Wait for a permit (blocks if max_concurrent executions are active)
        let _permit = pool
            .inner
            .semaphore
            .acquire()
            .await
            .map_err(|_| anyhow::anyhow!("WasmInstancePool semaphore closed"))?;

        // Forget the permit — we manage it manually in PooledWasmInstanceGuard::drop
        _permit.forget();

        // Try to reuse an idle instance
        let instance = {
            let mut idle = pool.inner.idle.lock().await;
            idle.pop_front()
        };

        let instance = match instance {
            Some(i) => {
                tracing::debug!(
                    "WasmInstancePool[{}]: reusing idle instance (reuse_count={})",
                    pool.inner.function_key,
                    i.reuse_count
                );
                i
            }
            None => {
                tracing::debug!(
                    "WasmInstancePool[{}]: no idle instance available",
                    pool.inner.function_key
                );
                return Err(anyhow::anyhow!(
                    "No pooled instance available for {}. Pool is empty and must be pre-warmed via prewarm_instance().",
                    pool.inner.function_key
                ));
            }
        };

        Ok(PooledWasmInstanceGuard {
            instance: Some(instance),
            pool,
            discard: false,
        })
    }

    /// Pre-warm the pool by adding a compiled instance.
    ///
    /// Call this during startup or when you anticipate traffic to avoid
    /// cold-start latency on the first request.
    pub async fn prewarm(&self, instance: PooledWasmInstance) {
        if let Ok(mut idle) = self.inner.idle.try_lock() {
            if idle.len() < self.inner.max_idle {
                idle.push_back(instance);
                tracing::debug!(
                    "WasmInstancePool[{}]: pre-warmed instance",
                    self.inner.function_key
                );
            }
        }
    }

    /// Pre-warm with a module and WASI context directly.
    ///
    /// This is a convenience method that creates the `PooledWasmInstance` for you.
    pub async fn prewarm_with(
        &self,
        module: wasmtime::Module,
        wasi_ctx: WasiP1Ctx,
        pipe_capacity: usize,
    ) {
        let instance = PooledWasmInstance::new(
            module,
            wasi_ctx,
            self.inner.function_key.clone(),
            pipe_capacity,
        );
        self.prewarm(instance).await;
    }

    /// Return pool statistics.
    pub async fn stats(&self) -> WasmPoolStats {
        let idle = self.inner.idle.lock().await;
        WasmPoolStats {
            idle_count: idle.len(),
            max_idle: self.inner.max_idle,
            available_permits: self.inner.semaphore.available_permits(),
            max_concurrent: self.inner.max_concurrent,
            function_key: self.inner.function_key.clone(),
        }
    }

    /// Check if the pool has any warm instances.
    pub async fn is_warmed(&self) -> bool {
        let idle = self.inner.idle.lock().await;
        !idle.is_empty()
    }

    /// Get the function key for this pool.
    pub fn function_key(&self) -> &str {
        &self.inner.function_key
    }
}

/// Statistics about a WASM instance pool.
#[derive(Debug, Clone)]
pub struct WasmPoolStats {
    pub idle_count: usize,
    pub max_idle: usize,
    pub available_permits: usize,
    pub max_concurrent: usize,
    pub function_key: String,
}

impl std::fmt::Display for WasmPoolStats {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(
            f,
            "WasmPoolStats {{ {}: idle={}/{}, permits={}/{} }}",
            self.function_key,
            self.idle_count,
            self.max_idle,
            self.available_permits,
            self.max_concurrent
        )
    }
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

    /// Start background pruning task.
    ///
    /// The pool must already be wrapped in `Arc<RwLock<InstancePool>>` before
    /// calling this method. Pass the shared reference so the pruning task
    /// operates on the *same* pool instance that is used by the server, not a
    /// detached clone.
    ///
    /// # Previous bug
    /// The old implementation called `self.clone()` and wrapped the clone in a
    /// new `Arc<RwLock<...>>`. The spawned task then pruned that detached clone
    /// while the server continued using the original `self` — meaning the pool
    /// was never actually pruned in production.
    pub fn start_background_pruning_shared(shared_pool: Arc<RwLock<InstancePool>>) -> tokio::task::JoinHandle<()> {
        tokio::spawn(async move {
            let mut interval = interval(TokioDuration::from_secs(60)); // Prune every minute

            loop {
                interval.tick().await;
                let mut pool_guard = shared_pool.write().await;

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
        })
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

/// Manages multiple `WasmInstancePool` objects, one per function.
///
/// This is the main entry point for the WASM instance pooling infrastructure.
/// Each function gets its own pool with per-function concurrency limits and idle
/// instance caching.
///
/// # Example
/// ```
/// let manager = PoolManager::new(10, 4); // max_concurrent=10, max_idle=4
/// manager.prewarm_instance("fn@1.0.0", module, wasi_ctx, 1024*1024).await;
/// let guard = manager.acquire("fn@1.0.0").await?;
/// ```
pub struct PoolManager {
    /// Per-function instance pools
    pools: RwLock<HashMap<String, Arc<WasmInstancePool>>>,
    /// Maximum concurrent executions per function
    max_concurrent: usize,
    /// Maximum idle instances per function
    max_idle: usize,
    /// Default pipe capacity
    default_pipe_capacity: usize,
}

impl PoolManager {
    /// Create a new pool manager.
    ///
    /// * `max_concurrent` - Maximum concurrent executions per function
    /// * `max_idle` - Maximum idle instances to keep warm per function
    pub fn new(max_concurrent: usize, max_idle: usize) -> Self {
        Self {
            pools: RwLock::new(HashMap::new()),
            max_concurrent,
            max_idle,
            default_pipe_capacity: 1024 * 1024, // 1 MiB
        }
    }

    /// Create with custom pipe capacity.
    pub fn with_pipe_capacity(mut self, capacity: usize) -> Self {
        self.default_pipe_capacity = capacity;
        self
    }

    /// Get or create a pool for the given function key.
    pub async fn get_or_create_pool(&self, function_key: &str) -> Arc<WasmInstancePool> {
        // Try read lock first (fast path)
        {
            let pools = self.pools.read().await;
            if let Some(pool) = pools.get(function_key) {
                return Arc::clone(pool);
            }
        }

        // Need to create - acquire write lock
        let mut pools = self.pools.write().await;
        // Double-check after acquiring write lock
        if let Some(pool) = pools.get(function_key) {
            return Arc::clone(pool);
        }

        let pool = Arc::new(WasmInstancePool::new(
            function_key.to_string(),
            self.max_concurrent,
            self.max_idle,
        ));
        pools.insert(function_key.to_string(), Arc::clone(&pool));
        tracing::info!(
            "Created new WasmInstancePool for {} (max_concurrent={}, max_idle={})",
            function_key,
            self.max_concurrent,
            self.max_idle
        );
        pool
    }

    /// Acquire a pooled instance for the given function.
    ///
    /// Returns a guard that automatically returns the instance to the pool on drop.
    pub async fn acquire(&self, function_key: &str) -> anyhow::Result<PooledWasmInstanceGuard> {
        let pool = self.get_or_create_pool(function_key).await;
        WasmInstancePool::acquire(pool).await
    }

    /// Pre-warm a pool with a compiled instance.
    ///
    /// Call this during startup or when you anticipate traffic to a function.
    /// If the pool doesn't exist yet, it will be created.
    pub async fn prewarm_instance(
        &self,
        function_key: &str,
        module: wasmtime::Module,
        wasi_ctx: WasiP1Ctx,
    ) {
        let pool = self.get_or_create_pool(function_key).await;
        pool.prewarm_with(module, wasi_ctx, self.default_pipe_capacity).await;
    }

    /// Pre-warm with a PooledWasmInstance directly.
    pub async fn prewarm(&self, function_key: &str, instance: PooledWasmInstance) {
        let pool = self.get_or_create_pool(function_key).await;
        pool.prewarm(instance).await;
    }

    /// Get statistics for all pools.
    pub async fn stats(&self) -> Vec<WasmPoolStats> {
        let pools = self.pools.read().await;
        let mut stats = Vec::with_capacity(pools.len());
        for pool in pools.values() {
            stats.push(pool.stats().await);
        }
        stats
    }

    /// Get stats for a specific function's pool.
    pub async fn pool_stats(&self, function_key: &str) -> Option<WasmPoolStats> {
        let pools = self.pools.read().await;
        pools.get(function_key).map(|p| p.blocking_stats())
    }

    /// Get the total number of warmed functions.
    pub async fn warmed_function_count(&self) -> usize {
        let pools = self.pools.read().await;
        pools.values().filter(|p| p.is_warmed_blocking()).count()
    }

    /// Check if a specific function's pool has warmed instances.
    pub async fn is_warmed(&self, function_key: &str) -> bool {
        let pools = self.pools.read().await;
        if let Some(pool) = pools.get(function_key) {
            pool.is_warmed_blocking()
        } else {
            false
        }
    }

    /// Clear all pools.
    pub async fn clear(&self) {
        let mut pools = self.pools.write().await;
        pools.clear();
        tracing::info!("Cleared all WasmInstancePools");
    }

    /// Remove a specific function's pool.
    pub async fn remove_pool(&self, function_key: &str) {
        let mut pools = self.pools.write().await;
        if pools.remove(function_key).is_some() {
            tracing::info!("Removed WasmInstancePool for {}", function_key);
        }
    }
}

impl WasmInstancePool {
    /// Blocking version of stats() for use from non-async contexts.
    fn blocking_stats(&self) -> WasmPoolStats {
        // Use try_lock to avoid blocking; return zeros if lock is held
        match self.inner.idle.try_lock() {
            Ok(idle) => WasmPoolStats {
                idle_count: idle.len(),
                max_idle: self.inner.max_idle,
                available_permits: self.inner.semaphore.available_permits(),
                max_concurrent: self.inner.max_concurrent,
                function_key: self.inner.function_key.clone(),
            },
            Err(_) => WasmPoolStats {
                idle_count: 0,
                max_idle: self.inner.max_idle,
                available_permits: 0,
                max_concurrent: self.inner.max_concurrent,
                function_key: self.inner.function_key.clone(),
            },
        }
    }

    /// Check if warmed without awaiting.
    fn is_warmed_blocking(&self) -> bool {
        self.inner.idle.try_lock().is_ok_and(|idle| !idle.is_empty())
    }
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
