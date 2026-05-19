//! Mesh protocols for capability advertisement and state sync

use serde::{Deserialize, Serialize};

/// Capability advertisement message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapabilityAdvert {
    pub peer_id: String,
    pub capabilities: Vec<CapabilityInfo>,
    pub timestamp: i64,
}

/// Information about a capability
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapabilityInfo {
    pub name: String,
    pub category: String,
    pub performance_score: f32,
    pub trust_score: f32,
    pub is_local: bool,
    pub endpoint: Option<String>,
}

/// State synchronization message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StateSync {
    pub sync_id: String,
    pub peer_id: String,
    pub state_type: StateSyncType,
    pub payload: Vec<u8>,
    pub version: u64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum StateSyncType {
    Full,
    Delta,
    Request,
    Response,
}

/// Protocol handler for mesh communications
pub struct MeshProtocol;

impl MeshProtocol {
    /// Create a capability advertisement message
    pub fn capability_advert(peer_id: &str, capabilities: Vec<CapabilityInfo>) -> CapabilityAdvert {
        CapabilityAdvert {
            peer_id: peer_id.to_string(),
            capabilities,
            timestamp: chrono::Utc::now().timestamp(),
        }
    }

    /// Create a state sync message
    pub fn state_sync(sync_id: &str, peer_id: &str, sync_type: StateSyncType, payload: Vec<u8>, version: u64) -> StateSync {
        StateSync {
            sync_id: sync_id.to_string(),
            peer_id: peer_id.to_string(),
            state_type: sync_type,
            payload,
            version,
        }
    }
}