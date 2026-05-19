//! Swarm protocol for inter-cell communication

use super::{SwarmMessage, SwarmMessageType};
use crate::codec::{CborCodec, CodecError};
use crate::core::PrismResult;

/// Swarm protocol handler
pub struct SwarmProtocol;

impl SwarmProtocol {
    /// Create a coordination message
    pub fn coordination_message(swarm_id: &str, sender_id: &str, payload: Vec<u8>) -> SwarmMessage {
        SwarmMessage::new(
            super::SwarmId::new(swarm_id),
            sender_id,
            SwarmMessageType::Coordination,
            payload,
        )
    }

    /// Create a heartbeat message
    pub fn heartbeat(swarm_id: &str, sender_id: &str) -> SwarmMessage {
        SwarmMessage::new(
            super::SwarmId::new(swarm_id),
            sender_id,
            SwarmMessageType::Heartbeat,
            vec![],
        )
    }

    /// Create a health report message
    pub fn health_report(swarm_id: &str, sender_id: &str, health_json: Vec<u8>) -> SwarmMessage {
        SwarmMessage::new(
            super::SwarmId::new(swarm_id),
            sender_id,
            SwarmMessageType::HealthReport,
            health_json,
        )
    }

    /// Parse a swarm message from CBOR bytes
    pub fn parse_message(data: &[u8]) -> PrismResult<SwarmMessage> {
        CborCodec::decode(data)
            .map_err(|e| crate::core::PrismError::SerializationError(e.to_string()))
    }

    /// Serialize a swarm message to CBOR bytes
    pub fn serialize_message(msg: &SwarmMessage) -> PrismResult<Vec<u8>> {
        CborCodec::encode(msg)
            .map_err(|e| crate::core::PrismError::SerializationError(e.to_string()))
    }

    /// Parse a swarm message from CBOR hex string (for debugging)
    pub fn parse_message_hex(hex: &str) -> PrismResult<SwarmMessage> {
        let bytes: Vec<u8> = hex
            .as_bytes()
            .chunks(2)
            .map(|chunk| {
                let s = std::str::from_utf8(chunk).unwrap();
                u8::from_str_radix(s, 16).unwrap()
            })
            .collect();
        Self::parse_message(&bytes)
    }

    /// Serialize a swarm message to CBOR hex string (for debugging/logging)
    pub fn serialize_message_hex(msg: &SwarmMessage) -> Result<String, CodecError> {
        let bytes = CborCodec::encode(msg)?;
        Ok(bytes.iter().map(|b| format!("{:02x}", b)).collect())
    }
}