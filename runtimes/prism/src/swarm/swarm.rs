//! Swarm types and state

use chrono::Utc;
use serde::{Deserialize, Serialize};

use super::coordinator::SwarmId;
use crate::codec::{CborCodec, CodecError};

/// A swarm message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SwarmMessage {
    pub message_id: String,
    pub swarm_id: SwarmId,
    pub sender_id: String,
    pub msg_type: SwarmMessageType,
    pub payload: Vec<u8>,
    pub timestamp: i64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SwarmMessageType {
    Coordination,
    Heartbeat,
    HealthReport,
    Delegation,
    Migration,
    Termination,
}

impl SwarmMessage {
    pub fn new(swarm_id: SwarmId, sender_id: &str, msg_type: SwarmMessageType, payload: Vec<u8>) -> Self {
        Self {
            message_id: uuid::Uuid::new_v4().to_string(),
            swarm_id,
            sender_id: sender_id.to_string(),
            msg_type,
            payload,
            timestamp: Utc::now().timestamp(),
        }
    }

    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }

    /// Serialize to CBOR hex string for logging
    pub fn to_cbor_hex(&self) -> Result<String, CodecError> {
        let bytes = self.to_cbor()?;
        Ok(bytes.iter().map(|b| format!("{:02x}", b)).collect())
    }
}