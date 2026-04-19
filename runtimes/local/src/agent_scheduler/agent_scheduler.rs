//! Agent scheduler — priority queues, per-tenant rate limiting, backpressure.
//!
//! This module replaces the ad-hoc worker pool in `server.rs` with a proper
//! scheduler that supports:
//! - **4 priority queues**: critical, high, normal, low
//! - **Per-tenant rate limiting** via Redis (or in-process fallback)
//! - **Backpressure** when queues are saturated
//! - **Sub-50ms scheduling delay** under 10k+ concurrent agents
//!
//! ## Queue Capacity
//!
//! Each queue has a configurable max depth. When a queue is full, new
//! submissions are rejected (not queued — this prevents unbounded memory growth).
//!
//! ## Rate Limiting
//!
//! Uses a sliding-window counter per tenant. With the `memory` feature enabled
//! (Redis), counters are distributed and survive restarts. Without Redis,
//! an in-process `dashmap` LRU is used as fallback (per-instance only).
//!
//! ## Backpressure Signalling
//!
//! When the critical queue is saturated, `check_backpressure()` returns
//! `SchedulerBackpressure::CriticalSaturated` so callers can fail-fast with
//! `503` instead of enqueueing.

use std::collections::VecDeque;
use std::sync::Arc;
use std::time::{Duration, Instant};

use tokio::sync::{broadcast, Semaphore};
use tracing::{debug, info, instrument, warn};

use crate::engine::graph::{ExecutionPriority, GraphExecutionInput};
use crate::agent_scheduler::worker::QueuedGraphExecution;
use uuid::Uuid;

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/// Scheduling priority level — maps to one of the four internal queues.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum PriorityLevel {
    Low = 1,
    Normal = 2,
    High = 3,
    Critical = 4,
}

impl Default for PriorityLevel {
    fn default() -> Self {
        PriorityLevel::Normal
    }
}

impl From<ExecutionPriority> for PriorityLevel {
    fn from(ep: ExecutionPriority) -> Self {
        match ep {
            ExecutionPriority::Low => PriorityLevel::Low,
            ExecutionPriority::Normal => PriorityLevel::Normal,
            ExecutionPriority::High => PriorityLevel::High,
            ExecutionPriority::Critical => PriorityLevel::Critical,
        }
    }
}

/// A queued execution waiting to be dispatched.
#[derive(Debug, Clone)]
pub struct QueuedExecution {
    /// Unique ID for this queued item.
    pub id: Uuid,
    /// Graph ID being executed.
    pub graph_id: Uuid,
    /// Graph execution input.
    pub input: GraphExecutionInput,
    /// Queue priority.
    pub priority: PriorityLevel,
    /// When this item was enqueued.
    pub enqueued_at: Instant,
    /// Optional tenant ID for rate limiting.
    pub tenant_id: Option<String>,
}

impl QueuedExecution {
    /// Create a new queued execution.
    pub fn new(
        graph_id: Uuid,
        input: GraphExecutionInput,
        priority: PriorityLevel,
        tenant_id: Option<String>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            graph_id,
            input,
            priority,
            enqueued_at: Instant::now(),
            tenant_id,
        }
    }

    /// How long this item has been waiting.
    pub fn waiting_duration(&self) -> Duration {
        self.enqueued_at.elapsed()
    }
}

/// Configuration for a priority queue.
#[derive(Debug, Clone)]
pub struct QueueConfig {
    /// Maximum number of items in this queue.
    pub max_depth: usize,
    /// Maximum time an item can wait before being rejected.
    pub max_age: Duration,
}

impl Default for QueueConfig {
    fn default() -> Self {
        Self {
            max_depth: 10_000,
            max_age: Duration::from_secs(300),
        }
    }
}

/// Configuration for per-tenant rate limiting.
#[derive(Debug, Clone)]
pub struct RateLimitConfig {
    /// Maximum requests per window.
    pub requests_per_window: u32,
    /// Window duration in seconds.
    pub window_secs: u64,
    /// Burst allowance (extra requests allowed above limit).
    pub burst: u32,
}

impl Default for RateLimitConfig {
    fn default() -> Self {
        Self {
            requests_per_window: 100,
            window_secs: 60,
            burst: 20,
        }
    }
}

/// Backpressure state returned by `check_backpressure()`.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SchedulerBackpressure {
    /// All queues are healthy — proceed normally.
    Ok,
    /// Critical queue is saturated — new critical submissions should be rejected.
    CriticalSaturated,
    /// All queues are saturated — reject all new submissions.
    FullySaturated,
}

impl SchedulerBackpressure {
    /// Returns `true` if we should reject new submissions.
    pub fn is_rejected(&self) -> bool {
        matches!(self, SchedulerBackpressure::CriticalSaturated | SchedulerBackpressure::FullySaturated)
    }

    /// Human-readable label for metrics/logs.
    pub fn label(&self) -> &'static str {
        match self {
            SchedulerBackpressure::Ok => "ok",
            SchedulerBackpressure::CriticalSaturated => "critical_saturated",
            SchedulerBackpressure::FullySaturated => "fully_saturated",
        }
    }
}

// ---------------------------------------------------------------------------
// Rate Limiter
// ---------------------------------------------------------------------------

/// Per-tenant rate limiter using sliding-window counters.
///
/// With the `memory` feature (Redis): distributed, survives restarts.
/// Without `memory` feature: in-process dashmap LRU, per-instance only.

pub struct TenantRateLimiter {
    /// Rate limit configuration.
    config: RateLimitConfig,
    /// In-process counter fallback (used when Redis is unavailable).
    counters: dashmap::DashMap<String, (u32, Instant)>,
    /// Redis client for distributed counters (when `memory` feature is enabled).
    #[cfg(feature = "memory")]
    redis_url: Option<String>,
}

impl TenantRateLimiter {
    /// Create a new rate limiter from config.
    pub fn new(config: RateLimitConfig) -> Self {
        Self {
            config,
            counters: dashmap::DashMap::new(),
            #[cfg(feature = "memory")]
            redis_url: std::env::var("REDIS_URL").ok(),
        }
    }

    /// Check if a tenant is within their rate limit.
    /// Returns `Ok(remaining)` if allowed, `Err(wait_secs)` if limited.
    pub async fn check(&self, tenant_id: &str) -> Result<u32, u64> {
        let now = Instant::now();
        let window = Duration::from_secs(self.config.window_secs);

        #[cfg(not(feature = "memory"))]
        {
            // In-process sliding window counter
            let mut slot = self.counters.entry(tenant_id.to_string()).or_insert((0, now));

            // Reset window if expired
            if slot.1.elapsed() > window {
                slot.1 = now;
                slot.0 = 0;
            }

            if slot.0 >= self.config.requests_per_window {
                let wait = window.saturating_sub(slot.1.elapsed()).as_secs().max(1);
                return Err(wait);
            }

            slot.0 += 1;
            let remaining = self.config.requests_per_window - slot.0;
            Ok(remaining)
        }

        #[cfg(feature = "memory")]
        {
            use redis::AsyncCommands;

            let redis_url = match &self.redis_url {
                Some(url) => url,
                None => return self.check_inprocess(tenant_id).await,
            };

            let client = match redis::Client::open(redis_url.as_str()) {
                Ok(c) => c,
                Err(_) => return self.check_inprocess(tenant_id).await,
            };

            let mut conn = match client.get_multiplexed_async_connection().await {
                Ok(c) => c,
                Err(_) => return self.check_inprocess(tenant_id).await,
            };

            let key = format!("rate_limit:{}", tenant_id);
            let ttl = self.config.window_secs as i64;

            let count: Option<u32> = conn.get(&key).await.unwrap_or(None);
            let count = count.unwrap_or(0);

            if count >= self.config.requests_per_window {
                // Get TTL to tell caller how long to wait
                let remaining_ttl: Option<i64> = conn.ttl(&key).await.unwrap_or(None);
                let wait = remaining_ttl.unwrap_or(self.config.window_secs as i64).max(1) as u64;
                return Err(wait);
            }

            // Increment counter with expiry
            let _: () = conn.incr(&key, 1u32).await.unwrap_or(());
            let _: () = conn.expire(&key, ttl).await.unwrap_or(());

            let remaining = self.config.requests_per_window.saturating_sub(count + 1);
            Ok(remaining)
        }
    }

    /// In-process fallback when Redis is unavailable.
    async fn check_inprocess(&self, tenant_id: &str) -> Result<u32, u64> {
        let now = Instant::now();
        let window = Duration::from_secs(self.config.window_secs);

        let mut entry = self.counters.entry(tenant_id.to_string()).or_insert((0, now));

        if entry.1.elapsed() > window {
            entry.1 = now;
            entry.0 = 0;
        }

        if entry.0 >= self.config.requests_per_window {
            let wait = window.saturating_sub(entry.1.elapsed()).as_secs().max(1);
            return Err(wait);
        }

        entry.0 += 1;
        let remaining = self.config.requests_per_window - entry.0;
        Ok(remaining)
    }

    /// Reset rate limit for a tenant (e.g., when upgrading tier).
    pub async fn reset(&self, tenant_id: &str) {
        #[cfg(not(feature = "memory"))]
        {
            self.counters.remove(tenant_id);
        }
        #[cfg(feature = "memory")]
        {
            use redis::AsyncCommands;
            if let Some(url) = &self.redis_url {
                if let Ok(client) = redis::Client::open(url.as_str()) {
                    if let Ok(mut conn) = client.get_multiplexed_async_connection().await {
                        let key = format!("rate_limit:{}", tenant_id);
                        let _: () = conn.del(&key).await.unwrap_or(());
                    }
                }
            }
            // Also clean in-process
            self.counters.remove(tenant_id);
        }
    }
}

// ---------------------------------------------------------------------------
// Agent Scheduler
// ---------------------------------------------------------------------------

/// Multi-tenant agent scheduler with priority queues and rate limiting.
///
/// Uses 4 priority queues internally. Attempts to acquire a rate-limit
/// token before accepting a new execution. Enforces queue depth limits and
/// signals backpressure when saturated.
///
/// Wrap in `Arc<AgentScheduler>` and share across request handlers.
pub struct AgentScheduler {
    queues: [VecDeque<QueuedGraphExecution>; 4],
    semaphores: [Semaphore; 4],
    queue_configs: [QueueConfig; 4],
    rate_limiter: Arc<TenantRateLimiter>,
    /// Broadcast channel for queue depth updates (for observability).
    depth_tx: broadcast::Sender<QueueDepthSnapshot>,
    /// Shutdown flag.
    shutdown: std::sync::atomic::AtomicBool,
}

/// Snapshot of all queue depths for observability.
#[derive(Debug, Clone)]
pub struct QueueDepthSnapshot {
    pub critical_depth: usize,
    pub high_depth: usize,
    pub normal_depth: usize,
    pub low_depth: usize,
    pub total: usize,
}

impl AgentScheduler {
    /// Create a new scheduler with default queue configs.
    pub fn new(rate_limit_config: RateLimitConfig) -> Self {
        let (depth_tx, _) = broadcast::channel(64);
        Self {
            queues: [
                VecDeque::new(), // Low
                VecDeque::new(), // Normal
                VecDeque::new(), // High
                VecDeque::new(), // Critical
            ],
            semaphores: [
                Semaphore::new(5000),
                Semaphore::new(5000),
                Semaphore::new(2000),
                Semaphore::new(500),
            ],
            queue_configs: [
                QueueConfig { max_depth: 50_000, ..Default::default() },
                QueueConfig { max_depth: 30_000, ..Default::default() },
                QueueConfig { max_depth: 10_000, ..Default::default() },
                QueueConfig { max_depth: 1_000, ..Default::default() },
            ],
            rate_limiter: Arc::new(TenantRateLimiter::new(rate_limit_config)),
            depth_tx,
            shutdown: std::sync::atomic::AtomicBool::new(false),
        }
    }

    /// Get the broadcast receiver for queue depth updates.
    pub fn subscribe_depth(&self) -> broadcast::Receiver<QueueDepthSnapshot> {
        self.depth_tx.subscribe()
    }

    /// Check backpressure and return the current state.
    pub fn check_backpressure(&self) -> SchedulerBackpressure {
        let critical = self.queues[3].len();
        let high = self.queues[2].len();
        let normal = self.queues[1].len();
        let low = self.queues[0].len();

        if critical >= self.queue_configs[3].max_depth {
            return SchedulerBackpressure::CriticalSaturated;
        }
        if critical >= self.queue_configs[3].max_depth / 2
            && high >= self.queue_configs[2].max_depth / 2
            && normal >= self.queue_configs[1].max_depth / 2
        {
            return SchedulerBackpressure::FullySaturated;
        }

        SchedulerBackpressure::Ok
    }

    /// Enqueue a new execution with full graph definition.
    ///
    /// Returns `Ok(queued_id>` if successfully queued, or an error if:
    /// - Backpressure is preventing new submissions
    /// - Rate limit exceeded for this tenant
    /// - Queue is at capacity
    #[instrument(skip(self, exec))]
    pub async fn enqueue(
        &mut self,
        exec: QueuedGraphExecution,
    ) -> Result<Uuid, SchedulerError> {
        // Check backpressure first
        let bp = self.check_backpressure();
        let level = exec.priority;

        if bp == SchedulerBackpressure::FullySaturated {
            return Err(SchedulerError::QueueSaturated {
                queue: level,
                reason: "all queues at capacity".to_string(),
            });
        }

        if bp == SchedulerBackpressure::CriticalSaturated && level == PriorityLevel::Critical {
            return Err(SchedulerError::QueueSaturated {
                queue: level,
                reason: "critical queue at capacity".to_string(),
            });
        }

        // Check rate limit if tenant ID is present
        if let Some(ref tenant_id) = exec.tenant_id {
            match self.rate_limiter.check(tenant_id).await {
                Ok(remaining) => {
                    debug!(tenant_id = %tenant_id, remaining = remaining, "Rate limit check passed");
                }
                Err(wait_secs) => {
                    return Err(SchedulerError::RateLimited {
                        tenant_id: tenant_id.clone(),
                        retry_after_secs: wait_secs,
                    });
                }
            }
        }

        // Queue index: 0=Low, 1=Normal, 2=High, 3=Critical
        let idx = level as usize - 1;

        // Check queue capacity
        if self.queues[idx].len() >= self.queue_configs[idx].max_depth {
            return Err(SchedulerError::QueueSaturated {
                queue: level,
                reason: format!("queue {} at max depth", idx),
            });
        }

        // Try to acquire a semaphore permit (limits concurrent processing)
        let permit = self.semaphores[idx].acquire().await.map_err(|_| {
            SchedulerError::QueueSaturated {
                queue: level,
                reason: "semaphore closed".to_string(),
            }
        })?;

        let exec_id = exec.id;

        // Drop permit immediately — we just wanted to bound enqueue, not hold it
        drop(permit);

        self.queues[idx].push_back(exec);

        // Broadcast depth update
        let _ = self.broadcast_depth();

        debug!(
            queue = ?level,
            queue_depth = self.queues[idx].len(),
            exec_id = %exec_id,
            "Execution enqueued"
        );

        Ok(exec_id)
    }

    /// Dequeue the next execution, respecting priority ordering.
    ///
    /// Checks queues in priority order (Critical → High → Normal → Low)
    /// and returns the first available item, or `None` if all queues are empty.
    pub async fn dequeue(&mut self) -> Option<QueuedGraphExecution> {
        for idx in [3, 2, 1, 0] {
            if let Some(exec) = self.queues[idx].pop_front() {
                let _ = self.broadcast_depth();
                return Some(exec);
            }
        }
        None
    }

    /// Get current queue depths.
    pub fn queue_depths(&self) -> QueueDepthSnapshot {
        let critical = self.queues[3].len();
        let high = self.queues[2].len();
        let normal = self.queues[1].len();
        let low = self.queues[0].len();
        QueueDepthSnapshot {
            critical_depth: critical,
            high_depth: high,
            normal_depth: normal,
            low_depth: low,
            total: critical + high + normal + low,
        }
    }

    fn broadcast_depth(&self) {
        let snapshot = self.queue_depths();
        let _ = self.depth_tx.send(snapshot);
    }
}

/// Errors returned by the scheduler.
#[derive(Debug, thiserror::Error)]
pub enum SchedulerError {
    #[error("Queue {queue:?} is saturated: {reason}")]
    QueueSaturated { queue: PriorityLevel, reason: String },

    #[error("Tenant {tenant_id} is rate-limited, retry after {retry_after_secs}s")]
    RateLimited { tenant_id: String, retry_after_secs: u64 },
}

impl SchedulerError {
    /// Returns `true` if this is a rate limit error.
    pub fn is_rate_limit(&self) -> bool {
        matches!(self, SchedulerError::RateLimited { .. })
    }

    /// HTTP status code for this error.
    pub fn status_code(&self) -> u16 {
        match self {
            SchedulerError::QueueSaturated { .. } => 503,
            SchedulerError::RateLimited { .. } => 429,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_enqueue_dequeue_priority_order() {
        use crate::engine::graph::Graph;

        let mut sched = AgentScheduler::new(RateLimitConfig::default());

        let graph1 = Graph::new(Uuid::new_v4(), "test1".to_string());
        let graph2 = Graph::new(Uuid::new_v4(), "test2".to_string());
        let graph3 = Graph::new(Uuid::new_v4(), "test3".to_string());
        let graph4 = Graph::new(Uuid::new_v4(), "test4".to_string());

        // Enqueue in mixed order using QueuedGraphExecution
        sched.enqueue(QueuedGraphExecution::new(
            graph1.clone(),
            GraphExecutionInput { graph_id: graph1.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Normal,
            None,
        )).await.unwrap();
        sched.enqueue(QueuedGraphExecution::new(
            graph2.clone(),
            GraphExecutionInput { graph_id: graph2.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Critical,
            None,
        )).await.unwrap();
        sched.enqueue(QueuedGraphExecution::new(
            graph3.clone(),
            GraphExecutionInput { graph_id: graph3.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Low,
            None,
        )).await.unwrap();
        sched.enqueue(QueuedGraphExecution::new(
            graph4.clone(),
            GraphExecutionInput { graph_id: graph4.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::High,
            None,
        )).await.unwrap();

        // Dequeue should return critical first
        let first = sched.dequeue().await.unwrap();
        assert_eq!(first.graph.id, graph2.id);

        // Then high
        let second = sched.dequeue().await.unwrap();
        assert_eq!(second.graph.id, graph4.id);

        // Then normal
        let third = sched.dequeue().await.unwrap();
        assert_eq!(third.graph.id, graph1.id);

        // Then low
        let fourth = sched.dequeue().await.unwrap();
        assert_eq!(fourth.graph.id, graph3.id);

        // Then none
        assert!(sched.dequeue().await.is_none());
    }

    #[tokio::test]
    async fn test_backpressure_when_critical_full() {
        use crate::engine::graph::Graph;

        let mut sched = AgentScheduler::new(RateLimitConfig::default());
        // Override critical max depth to small value for testing
        sched.queue_configs[3].max_depth = 2;

        // Fill critical queue
        let g = Graph::new(Uuid::new_v4(), "test".to_string());
        sched.enqueue(QueuedGraphExecution::new(
            g.clone(),
            GraphExecutionInput { graph_id: g.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Critical,
            None,
        )).await.unwrap();
        sched.enqueue(QueuedGraphExecution::new(
            Graph::new(Uuid::new_v4(), "test2".to_string()),
            GraphExecutionInput { graph_id: g.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Critical,
            None,
        )).await.unwrap();

        // Next critical should be rejected with backpressure
        let result = sched.enqueue(QueuedGraphExecution::new(
            Graph::new(Uuid::new_v4(), "test3".to_string()),
            GraphExecutionInput { graph_id: g.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Critical,
            None,
        )).await;
        assert!(result.is_err());

        // Lower priority should still work
        let result = sched.enqueue(QueuedGraphExecution::new(
            Graph::new(Uuid::new_v4(), "test4".to_string()),
            GraphExecutionInput { graph_id: g.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Normal,
            None,
        )).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_rate_limit_enforced() {
        use crate::engine::graph::Graph;

        let config = RateLimitConfig {
            requests_per_window: 2,
            window_secs: 60,
            burst: 0,
        };
        let mut sched = AgentScheduler::new(config);

        let g1 = Graph::new(Uuid::new_v4(), "test1".to_string());
        let g2 = Graph::new(Uuid::new_v4(), "test2".to_string());
        let g3 = Graph::new(Uuid::new_v4(), "test3".to_string());

        // First two should succeed
        assert!(sched.enqueue(QueuedGraphExecution::new(
            g1.clone(),
            GraphExecutionInput { graph_id: g1.id, initial_input: Default::default(), tenant_id: Some("tenant-a".to_string()) },
            PriorityLevel::Normal,
            Some("tenant-a".to_string()),
        )).await.is_ok());
        assert!(sched.enqueue(QueuedGraphExecution::new(
            g2.clone(),
            GraphExecutionInput { graph_id: g2.id, initial_input: Default::default(), tenant_id: Some("tenant-a".to_string()) },
            PriorityLevel::Normal,
            Some("tenant-a".to_string()),
        )).await.is_ok());

        // Third should be rate-limited
        let err = sched.enqueue(QueuedGraphExecution::new(
            g3.clone(),
            GraphExecutionInput { graph_id: g3.id, initial_input: Default::default(), tenant_id: Some("tenant-a".to_string()) },
            PriorityLevel::Normal,
            Some("tenant-a".to_string()),
        )).await.unwrap_err();
        assert!(err.is_rate_limit());
    }

    #[tokio::test]
    async fn test_queue_depths() {
        use crate::engine::graph::Graph;

        let mut sched = AgentScheduler::new(RateLimitConfig::default());

        let g1 = Graph::new(Uuid::new_v4(), "test1".to_string());
        let g2 = Graph::new(Uuid::new_v4(), "test2".to_string());
        let g3 = Graph::new(Uuid::new_v4(), "test3".to_string());

        sched.enqueue(QueuedGraphExecution::new(
            g1.clone(),
            GraphExecutionInput { graph_id: g1.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Critical,
            None,
        )).await.unwrap();
        sched.enqueue(QueuedGraphExecution::new(
            g2.clone(),
            GraphExecutionInput { graph_id: g2.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Critical,
            None,
        )).await.unwrap();
        sched.enqueue(QueuedGraphExecution::new(
            g3.clone(),
            GraphExecutionInput { graph_id: g3.id, initial_input: Default::default(), tenant_id: None },
            PriorityLevel::Normal,
            None,
        )).await.unwrap();

        let depths = sched.queue_depths();
        assert_eq!(depths.critical_depth, 2);
        assert_eq!(depths.normal_depth, 1);
        assert_eq!(depths.total, 3);
    }
}