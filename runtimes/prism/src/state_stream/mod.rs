//! StateStream Memory Fabric
//!
//! Distributed streaming memory with:
//! - event-sourced state
//! - resumable execution
//! - temporal rollback
//! - memory snapshots
//! - vector-aware state
//! - CRDT synchronization
//! - offline reconciliation
//! - deterministic replay
//!
//! This enables long-running agents, robotic systems, resilient workflows,
//! and collaborative swarms.

mod store;
mod stream;
mod crdt;
mod event;

#[cfg(feature = "state-stream")]
pub mod redis;

pub use store::{StateStore, StateSlice, StateKey, StatePersistence};
pub use stream::{StateStream, StreamHandle};
pub use crdt::{CrdtEngine, CrdtOp, LwwRegister, GCounter, PnCounter};
pub use event::{EventStore, Event, EventType};

#[cfg(feature = "state-stream")]
pub use redis::{RedisStatePersistence, CachedStatePersistence};