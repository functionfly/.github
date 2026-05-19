//! Common types for Prism Runtime

use serde::{Deserialize, Serialize};

/// Value encoding for state serialization
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ValueEncoding {
    Json,
    MsgPack,
    Cbor,
    Prost,
    Raw,
}

impl Default for ValueEncoding {
    fn default() -> Self {
        ValueEncoding::Json
    }
}

/// Strategy for state streaming
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum StreamStrategy {
    /// Immediate updates as they happen
    Realtime,
    /// Batch updates for efficiency
    Batch,
    /// Adaptive based on system load
    Adaptive,
}

impl Default for StreamStrategy {
    fn default() -> Self {
        StreamStrategy::Realtime
    }
}

/// Retention policy for state streams
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetentionPolicy {
    /// Maximum age of entries in seconds
    pub max_age_seconds: u32,
    /// Maximum number of items to retain
    pub max_items: u64,
    /// Whether to persist to cold storage
    pub persist_to_cold_storage: bool,
}

impl Default for RetentionPolicy {
    fn default() -> Self {
        Self {
            max_age_seconds: 86400, // 24 hours
            max_items: 100_000,
            persist_to_cold_storage: false,
        }
    }
}

/// Stream configuration for state management
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamConfig {
    /// Stream identifier
    pub stream_id: String,
    /// Streaming strategy
    pub strategy: StreamStrategy,
    /// Buffer size for batching
    pub buffer_size: u32,
    /// Whether stream is resumable
    pub resumable: bool,
    /// Retention policy
    pub retention: RetentionPolicy,
}

impl Default for StreamConfig {
    fn default() -> Self {
        Self {
            stream_id: String::new(),
            strategy: StreamStrategy::Realtime,
            buffer_size: 1000,
            resumable: true,
            retention: RetentionPolicy::default(),
        }
    }
}