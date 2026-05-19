//! StateFabric - Deterministic state management for AI agents
//!
//! Key features:
//! - Append-only event storage
//! - Snapshot-based state management
//! - Deterministic replay with Blake3 hashing
//! - Object storage for durability
//! - Wasmtime integration for function execution

pub mod api;
pub mod cache;
pub mod models;
pub mod replay;
pub mod state;
pub mod storage;
pub mod wasm;

pub use anyhow::Result;
pub use thiserror::Error;

// Re-export commonly used types
pub use models::{Event, Snapshot, State};
pub use state::StateManager;
