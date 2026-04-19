//! Priority-queue scheduler for SAR graph executions.
//!
//! Schedules `PendingExecution` items into priority queues and executes them
//! with per-tenant rate limiting and backpressure when queues are saturated.
//!
//! ## Priority Mapping
//!
//! | Tier        | Queue   |
//! |------------|---------|
//! | Critical   | Critical (4) |
//! | Paid High  | High (3)    |
//! | Normal     | Normal (2)   |
//! | Free/Low   | Low (1)      |
//!
//! ## Rate Limiting
//!
//! Per-tenant rate limits are enforced via Redis counters. When Redis is
//! unavailable (no `memory` feature), falls back to an in-process sliding-window
//! counter. The rate limit state is keyed by `rate_limit:{tenant_id}`.
//!
//! ## Backpressure
//!
//! When the critical queue exceeds `max_critical_depth`, new critical-priority
//! executions are rejected with `503 Service Unavailable` until the queue drains.

pub mod agent_scheduler;
pub mod worker;

pub use agent_scheduler::{
    AgentScheduler, PriorityLevel, RateLimitConfig,
    SchedulerBackpressure, TenantRateLimiter,
};
pub use worker::{
    JobTracker, JobStatus, TrackedJob, spawn_scheduler_workers,
    execute_queued_graph, QueuedGraphExecution,
};
