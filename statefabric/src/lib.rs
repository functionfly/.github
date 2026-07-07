//! StateFabric - Deterministic state management for AI agents
//!
//! Key features:
//! - Append-only event storage
//! - Snapshot-based state management
//! - Deterministic replay with Blake3 hashing
//! - Object storage for durability
//! - Wasmtime integration for function execution
//! - Multi-tenant isolation
//! - JWT/API key authentication
//! - Rate limiting
//! - Encryption at rest (AES-256-GCM)

pub mod api;
pub mod cache;
pub mod crypto;  // AES-256-GCM encryption for object storage
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

// Re-export crypto utilities
pub use crypto::{ObjectEncryptor, generate_key};
