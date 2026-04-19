//! Multi-tier memory layer for the SAR runtime.
//!
//! ## Architecture
//!
//! Memory flows through three tiers based on access patterns and age:
//!
//! ```text
//!   ┌─────────────────────────────────────────────────────┐
//!   │  Graph Node executes (MemoryOp::Read/Write)          │
//!   └────────────────┬────────────────────────────────────┘
//!                    │
//!                    ▼
//!   ┌─────────────────────────────────────────────────────┐
//!   │  HOT  — dashmap LRU (< 5ms read, no network)       │
//!   │  Tenant-isolated, in-process. Falls through to Warm. │
//!   └────────────────┬────────────────────────────────────┘
//!                    │ miss / explicit flush
//!                    ▼
//!   ┌─────────────────────────────────────────────────────┐
//!   │  WARM — Redis (redis-rs, shared across instances)  │
//!   │  Reuses Go backend Redis key schema.                │
//!   │  Falls through to Cold on miss.                    │
//!   └────────────────┬────────────────────────────────────┘
//!                    │ miss / cold query
//!                    ▼
//!   ┌─────────────────────────────────────────────────────┐
//!   │  COLD — PostgreSQL + pgvector (semantic recall)     │
//!   │  sqlx for queries, optional vector similarity.      │
//!   │  Stores execution history, long-term facts.        │
//!   └─────────────────────────────────────────────────────┘
//! ```
//!
//! ## Key Schema (follows Go backend patterns)
//!
//! - Hot: `mem:{tenant}:{key}` — in-process, no prefix sharing needed
//! - Warm Redis: `{prefix}:{tenant}:{key}` — matches `internal/cache/redis.go`
//!   - Execution state: `exec:{tenant}:{exec_id}`
//!   - Agent memory: `mem:{tenant}:{agent_id}:{key}`
//!   - Vector search: `vec:{tenant}:{index}:{id}`
//!
//! ## State Graph Memory
//!
//! Per-graph execution history is tracked for Phase 7 (optimizer):
//! - Decision log: which branches were taken
//! - Execution path: topological node order
//! - Success/failure rates per node
//! - Average latency per node

pub mod hot;
pub mod warm;
pub mod cold;
pub mod state;
pub mod layer;

pub use hot::HotMemory;
pub use warm::WarmMemory;
pub use cold::ColdMemory;
pub use state::{StateGraphMemory, NodeMetrics, ExecutionDecision};
pub use layer::MemoryLayer;
