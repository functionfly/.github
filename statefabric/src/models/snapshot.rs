//! Snapshot model - point-in-time state snapshots

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Represents a point-in-time snapshot of state
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Snapshot {
    /// Unique snapshot identifier
    pub id: Uuid,

    /// State ID this snapshot belongs to
    pub state_id: Uuid,

    /// Snapshot version number
    pub snapshot_version: i64,

    /// Optional human-readable label
    pub label: Option<String>,

    /// The state data at this snapshot point
    #[serde(rename = "state_data")]
    pub state_data: serde_json::Value,

    /// Size in bytes
    pub size_bytes: i64,

    /// Number of keys in this snapshot
    pub key_count: i32,

    /// First sequence number covered by this snapshot
    pub first_sequence: i64,

    /// Last sequence number covered by this snapshot
    pub last_sequence: i64,

    /// Root event ID (first event in this snapshot)
    pub root_event_id: Uuid,

    /// Whether the data is compressed
    pub is_compressed: bool,

    /// Compression algorithm used
    pub compression_algo: Option<String>,

    /// Checksum hash (Blake3)
    pub checksum: Option<String>,

    /// Object storage location
    pub storage_key: Option<String>,

    /// Creation timestamp
    pub created_at: DateTime<Utc>,
}

impl Snapshot {
    /// Create a new snapshot
    pub fn new(
        state_id: Uuid,
        snapshot_version: i64,
        state_data: serde_json::Value,
        first_sequence: i64,
        last_sequence: i64,
        root_event_id: Uuid,
    ) -> Self {
        let key_count = state_data.as_object()
            .map(|m| m.len() as i32)
            .unwrap_or(0);

        let size_bytes = serde_json::to_vec(&state_data)
            .map(|v| v.len() as i64)
            .unwrap_or(0);

        Self {
            id: Uuid::new_v4(),
            state_id,
            snapshot_version,
            label: None,
            state_data,
            size_bytes,
            key_count,
            first_sequence,
            last_sequence,
            root_event_id,
            is_compressed: false,
            compression_algo: None,
            checksum: None,
            storage_key: None,
            created_at: Utc::now(),
        }
    }

    /// Add a label to the snapshot
    pub fn with_label(mut self, label: String) -> Self {
        self.label = Some(label);
        self
    }

    /// Set the checksum
    pub fn with_checksum(mut self, checksum: String) -> Self {
        self.checksum = Some(checksum);
        self
    }

    /// Set compression info
    pub fn with_compression(mut self, algo: &str) -> Self {
        self.is_compressed = true;
        self.compression_algo = Some(algo.to_string());
        self
    }

    /// Set storage key
    pub fn with_storage_key(mut self, key: String) -> Self {
        self.storage_key = Some(key);
        self
    }
}

/// Snapshot metadata for indexing (stored in PostgreSQL)
/// The actual snapshot data is stored in object storage
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct SnapshotMetadata {
    pub id: Uuid,
    pub state_id: Uuid,
    pub snapshot_version: i64,
    pub label: Option<String>,
    pub key_count: i32,
    pub size_bytes: i64,
    pub first_sequence: i64,
    pub last_sequence: i64,
    pub root_event_id: Uuid,
    pub is_compressed: bool,
    pub compression_algo: Option<String>,
    pub checksum: Option<String>,
    pub storage_key: String,
    pub created_at: DateTime<Utc>,
}

/// Request to create a snapshot
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateSnapshotRequest {
    /// Optional label for the snapshot
    pub label: Option<String>,
    /// Whether to force snapshot even if no changes
    pub force: Option<bool>,
}

/// Request to restore from a snapshot
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RestoreSnapshotRequest {
    /// Snapshot version to restore from
    pub snapshot_version: i64,
    /// Whether to keep events after the snapshot
    pub keep_subsequent: Option<bool>,
}
