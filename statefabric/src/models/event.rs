//! Event model - immutable state change events for append-only storage

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Event types supported by StateFabric
#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum EventType {
    /// Set a key-value pair
    Set,
    /// Delete a key
    Delete,
    /// Create a snapshot
    Snapshot,
    /// Restore from snapshot
    Restore,
    /// Merge/upsert operations
    Merge,
    /// Clear all state
    Clear,
}

impl EventType {
    pub fn as_str(&self) -> &'static str {
        match self {
            EventType::Set => "set",
            EventType::Delete => "delete",
            EventType::Snapshot => "snapshot",
            EventType::Restore => "restore",
            EventType::Merge => "merge",
            EventType::Clear => "clear",
        }
    }

    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "set" => Some(EventType::Set),
            "delete" => Some(EventType::Delete),
            "snapshot" => Some(EventType::Snapshot),
            "restore" => Some(EventType::Restore),
            "merge" => Some(EventType::Merge),
            "clear" => Some(EventType::Clear),
            _ => None,
        }
    }
}

/// Source types for events
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum SourceType {
    /// Event from a function execution
    Function,
    /// Event from direct user action
    User,
    /// Event from system (snapshot, restore)
    System,
    /// Event from trigger
    Trigger,
}

impl SourceType {
    pub fn as_str(&self) -> &'static str {
        match self {
            SourceType::Function => "function",
            SourceType::User => "user",
            SourceType::System => "system",
            SourceType::Trigger => "trigger",
        }
    }
}

/// Represents an immutable event in the state event log
/// This is the core data structure for append-only storage
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    /// Unique event identifier
    pub id: Uuid,

    /// State ID this event belongs to
    pub state_id: Uuid,

    /// Event type
    #[serde(rename = "event_type")]
    pub event_type: EventType,

    /// Key affected (null for state-level events)
    pub key: Option<String>,

    /// Previous value (for diff/rollback)
    pub previous_value: Option<serde_json::Value>,

    /// New value
    pub new_value: Option<serde_json::Value>,

    /// Causation ID (for distributed tracing)
    pub causation_id: Option<Uuid>,

    /// Correlation ID (for distributed tracing)
    pub correlation_id: String,

    /// Source type
    #[serde(rename = "source_type")]
    pub source_type: SourceType,

    /// Source ID (function_id or user_id)
    #[serde(rename = "source_id")]
    pub source_id: String,

    /// Input hash for deterministic verification (Blake3)
    #[serde(rename = "input_hash")]
    pub input_hash: Option<String>,

    /// Output hash for deterministic verification (Blake3)
    #[serde(rename = "output_hash")]
    pub output_hash: Option<String>,

    /// Whether this event is deterministic (replay-safe)
    pub deterministic: bool,

    /// Sequence number (for ordering)
    #[serde(rename = "sequence_num")]
    pub sequence_num: i64,

    /// Timestamp
    pub timestamp: DateTime<Utc>,

    /// Object storage location (for durable storage)
    #[serde(rename = "storage_key")]
    pub storage_key: Option<String>,
}

impl Event {
    /// Create a new event
    pub fn new(
        state_id: Uuid,
        event_type: EventType,
        source_type: SourceType,
        source_id: String,
        correlation_id: Option<String>,
    ) -> Self {
        let now = Utc::now();

        Self {
            id: Uuid::new_v4(),
            state_id,
            event_type,
            key: None,
            previous_value: None,
            new_value: None,
            causation_id: None,
            correlation_id: correlation_id.unwrap_or_else(|| Uuid::new_v4().to_string()),
            source_type,
            source_id,
            input_hash: None,
            output_hash: None,
            deterministic: false,
            sequence_num: 0,
            timestamp: now,
            storage_key: None,
        }
    }

    /// Create a SET event
    pub fn set(
        state_id: Uuid,
        key: String,
        value: serde_json::Value,
        source_type: SourceType,
        source_id: String,
    ) -> Self {
        Self::new(state_id, EventType::Set, source_type, source_id, None)
            .with_key(key)
            .with_new_value(value)
    }

    /// Create a DELETE event
    pub fn delete(
        state_id: Uuid,
        key: String,
        previous_value: Option<serde_json::Value>,
        source_type: SourceType,
        source_id: String,
    ) -> Self {
        Self::new(state_id, EventType::Delete, source_type, source_id, None)
            .with_key(key)
            .with_previous_value(previous_value)
    }

    /// Create a MERGE event
    pub fn merge(
        state_id: Uuid,
        key: String,
        value: serde_json::Value,
        source_type: SourceType,
        source_id: String,
    ) -> Self {
        Self::new(state_id, EventType::Merge, source_type, source_id, None)
            .with_key(key)
            .with_new_value(value)
    }

    // Builder methods
    pub fn with_key(mut self, key: String) -> Self {
        self.key = Some(key);
        self
    }

    pub fn with_new_value(mut self, value: serde_json::Value) -> Self {
        self.new_value = Some(value);
        self
    }

    pub fn with_previous_value(mut self, value: Option<serde_json::Value>) -> Self {
        self.previous_value = value;
        self
    }

    pub fn with_deterministic(mut self, deterministic: bool) -> Self {
        self.deterministic = deterministic;
        self
    }

    pub fn with_sequence(mut self, seq: i64) -> Self {
        self.sequence_num = seq;
        self
    }

    pub fn with_input_hash(mut self, hash: String) -> Self {
        self.input_hash = Some(hash);
        self
    }

    pub fn with_output_hash(mut self, hash: String) -> Self {
        self.output_hash = Some(hash);
        self
    }

    pub fn with_storage_key(mut self, key: String) -> Self {
        self.storage_key = Some(key);
        self
    }

    /// Validate event data integrity and constraints
    /// Returns Ok(()) if valid, Err with message if invalid
    pub fn validate(&self) -> Result<(), String> {
        // Validate state_id is not nil
        if self.state_id == Uuid::nil() {
            return Err("state_id cannot be nil".to_string());
        }

        // Validate key format if present (alphanumeric, dash, underscore, max 256 chars)
        if let Some(ref key) = self.key {
            if key.is_empty() {
                return Err("key cannot be empty if provided".to_string());
            }
            if key.len() > 256 {
                return Err("key exceeds maximum length of 256 characters".to_string());
            }
            if !key.chars().all(|c| c.is_alphanumeric() || c == '-' || c == '_' || c == '.') {
                return Err("key contains invalid characters (allowed: alphanumeric, dash, underscore, dot)".to_string());
            }
        }

        // Validate value sizes (prevent memory exhaustion)
        if let Some(ref value) = self.new_value {
            let value_size = serde_json::to_vec(value)
                .map(|v| v.len())
                .unwrap_or(0);
            if value_size > 10 * 1024 * 1024 {
                // 10MB limit
                return Err("new_value exceeds maximum size of 10MB".to_string());
            }
        }

        if let Some(ref value) = self.previous_value {
            let value_size = serde_json::to_vec(value)
                .map(|v| v.len())
                .unwrap_or(0);
            if value_size > 10 * 1024 * 1024 {
                // 10MB limit
                return Err("previous_value exceeds maximum size of 10MB".to_string());
            }
        }

        // Validate source_id is not empty
        if self.source_id.is_empty() {
            return Err("source_id cannot be empty".to_string());
        }

        // Validate correlation_id is not empty
        if self.correlation_id.is_empty() {
            return Err("correlation_id cannot be empty".to_string());
        }

        // Validate event type constraints
        match self.event_type {
            EventType::Set | EventType::Merge => {
                if self.key.is_none() {
                    return Err("Set/Merge events must have a key".to_string());
                }
                if self.new_value.is_none() {
                    return Err("Set/Merge events must have new_value".to_string());
                }
            }
            EventType::Delete => {
                if self.key.is_none() {
                    return Err("Delete events must have a key".to_string());
                }
            }
            EventType::Clear | EventType::Snapshot | EventType::Restore => {
                // These don't require key/value
            }
        }

        Ok(())
    }
}

/// Event metadata for indexing (stored in PostgreSQL)
/// The actual event data is stored in object storage
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct EventMetadata {
    pub id: Uuid,
    pub state_id: Uuid,
    pub event_type: String,
    pub key: Option<String>,
    pub correlation_id: String,
    pub source_type: String,
    pub source_id: String,
    pub deterministic: bool,
    pub sequence_num: i64,
    pub timestamp: DateTime<Utc>,
    pub storage_key: String,
    pub input_hash: Option<String>,
    pub output_hash: Option<String>,
}
