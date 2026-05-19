//! State model - represents the state container for an agent

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Represents a state container for an agent's persistent data
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct State {
    /// Unique identifier for the state
    pub id: Uuid,

    /// Tenant/organization ID for multi-tenancy
    pub tenant_id: Uuid,

    /// Human-readable path (e.g., "agent/counter")
    pub path: String,

    /// Full path including tenant (e.g., "tenant123/agent/counter")
    pub full_path: String,

    /// Current version number (incremented on each change)
    pub current_version: i64,

    /// Storage type: "keyvalue", "document", etc.
    pub storage_type: String,

    /// Current state hash (Blake3)
    pub state_hash: Option<String>,

    /// Size in bytes
    pub size_bytes: i64,

    /// Key count
    pub key_count: i32,

    /// Whether deterministic mode is enabled
    pub deterministic: bool,

    /// Agent ID that owns this state (optional)
    pub agent_id: Option<Uuid>,

    /// Configuration
    pub config: serde_json::Value,

    /// Timestamps
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub last_accessed_at: DateTime<Utc>,
}

impl State {
    /// Create a new state
    pub fn new(tenant_id: Uuid, path: String, agent_id: Option<Uuid>) -> Self {
        let full_path = format!("{}/{}", tenant_id, path);
        let now = Utc::now();

        Self {
            id: Uuid::new_v4(),
            tenant_id,
            path,
            full_path,
            current_version: 0,
            storage_type: "keyvalue".to_string(),
            state_hash: None,
            size_bytes: 0,
            key_count: 0,
            deterministic: false,
            agent_id,
            config: serde_json::json!({}),
            created_at: now,
            updated_at: now,
            last_accessed_at: now,
        }
    }
}

/// State value - individual key-value pair
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StateValue {
    pub key: String,
    pub value: serde_json::Value,
    pub version: i64,
    pub updated_at: DateTime<Utc>,
    pub created_by: Option<String>,
}

/// State key-value entry for storage
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct StateKeyValue {
    pub state_id: Uuid,
    pub key: String,
    pub value: serde_json::Value,
    pub version: i64,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub created_by: Option<String>,
}
