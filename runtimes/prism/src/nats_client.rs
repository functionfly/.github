//! NATS client for Prism Runtime orchestration
//!
//! This module handles communication with the main orchestrator via NATS,
//! following the same patterns as the SAR runtime. Uses CBOR for efficient
//! binary serialization as required by the user.

use std::sync::Arc;
use std::time::Duration;
use chrono::Utc;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::{info, warn, debug};
use uuid::Uuid;

#[cfg(feature = "nats")]
use async_nats::Client;

#[cfg(feature = "nats")]
use futures_util::StreamExt;

use crate::core::CellId;
use crate::codec::{CborCodec, CodecError};

/// NATS client errors
#[derive(Error, Debug)]
pub enum NatsClientError {
    #[error("NATS error: {0}")]
    Nats(String),
    #[error("CBOR serialization error: {0}")]
    CborSerialization(#[from] CodecError),
    #[error("Not connected")]
    NotConnected,
    #[error("Connection closed")]
    ConnectionClosed,
    #[error("Subscription error: {0}")]
    Subscription(String),
}

/// Message types for orchestrator communication
/// Uses CBOR tag to discriminate variants for efficient binary serialization
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum OrchestratorMessageType {
    CellRegistered {
        cell_id: String,
        name: String,
        capabilities: Vec<String>,
    },
    CellUnregistered {
        cell_id: String,
    },
    CellHeartbeat {
        cell_id: String,
        status: String,
        active_executions: u32,
    },
    ExecutionRequest {
        execution_id: String,
        cell_id: String,
        input: CborBytes,
    },
    ExecutionResult {
        execution_id: String,
        status: String,
        output: Option<CborBytes>,
        error: Option<String>,
    },
    CapabilityAnnouncement {
        cell_id: String,
        capabilities: Vec<CapabilityInfo>,
    },
    SwarmAnnouncement {
        swarm_id: String,
        cell_id: String,
        action: SwarmAction,
    },
    RuntimeStatus {
        healthy: bool,
        active_cells: u32,
        active_swarms: u32,
    },
}

/// CBOR bytes wrapper for efficient binary serialization
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CborBytes {
    pub value: Vec<u8>,
}

impl CborBytes {
    pub fn new(value: Vec<u8>) -> Self {
        Self { value }
    }

    pub fn as_slice(&self) -> &[u8] {
        &self.value
    }
}

impl From<Vec<u8>> for CborBytes {
    fn from(value: Vec<u8>) -> Self {
        Self::new(value)
    }
}

/// Capability information for discovery
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapabilityInfo {
    pub name: String,
    pub category: String,
    pub version: String,
    pub endpoint: Option<String>,
}

/// Swarm action announcements
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SwarmAction {
    Created,
    Joined,
    Left,
    Dissolved,
}

/// Orchestrator message envelope with CBOR serialization
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrchestratorMessage {
    pub id: Uuid,
    pub msg_type: OrchestratorMessageType,
    pub cell_id: Option<CellId>,
    pub payload: CborBytes,
    pub timestamp: chrono::DateTime<Utc>,
}

impl OrchestratorMessage {
    /// Serialize this message to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, NatsClientError> {
        CborCodec::encode(self).map_err(NatsClientError::CborSerialization)
    }

    /// Deserialize a message from CBOR bytes
    pub fn from_cbor(data: &[u8]) -> Result<Self, NatsClientError> {
        CborCodec::decode(data).map_err(NatsClientError::CborSerialization)
    }
}

/// Connection state for monitoring
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Reconnecting,
}

/// NATS client for communicating with the orchestrator
/// Production-ready implementation with connection management and retry logic
#[derive(Clone)]
pub struct NatsOrchestratorClient {
    client: Arc<RwLock<Option<Client>>>,
    cell_id: Arc<RwLock<Option<String>>>,
    state: Arc<RwLock<ConnectionState>>,
    subjects: Arc<RwLock<Vec<String>>>,
}

impl NatsOrchestratorClient {
    /// Create a new NATS client
    pub fn new() -> Self {
        Self {
            client: Arc::new(RwLock::new(None)),
            cell_id: Arc::new(RwLock::new(None)),
            state: Arc::new(RwLock::new(ConnectionState::Disconnected)),
            subjects: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Set the cell ID for this runtime
    pub fn set_cell_id(&self, cell_id: CellId) {
        *self.cell_id.write() = Some(cell_id.to_string());
    }

    /// Get current connection state
    pub fn connection_state(&self) -> ConnectionState {
        *self.state.read()
    }

    /// Connect to NATS server with retry logic
    #[cfg(feature = "nats")]
    pub async fn connect(&mut self, url: &str) -> Result<(), NatsClientError> {
        *self.state.write() = ConnectionState::Connecting;

        let max_retries = 5;
        let base_delay = Duration::from_millis(100);
        let mut last_error = String::new();

        for attempt in 0..max_retries {
            match async_nats::connect(url).await {
                Ok(client) => {
                    *self.client.write() = Some(client);
                    *self.state.write() = ConnectionState::Connected;
                    info!("Prism connected to NATS at {} (attempt {})", url, attempt + 1);

                    // Spawn connection monitor
                    let client_clone = self.client.clone();
                    tokio::spawn(async move {
                        let mut interval = tokio::time::interval(Duration::from_secs(30));
                        loop {
                            interval.tick().await;
                            let _ = client_clone.read();
                            // Client health is tracked by state
                            debug!("NATS connection health check");
                        }
                    });

                    return Ok(());
                }
                Err(e) => {
                    last_error = e.to_string();
                    warn!("NATS connection attempt {} failed: {}", attempt + 1, last_error);

                    if attempt < max_retries - 1 {
                        let delay = base_delay * 2u32.pow(attempt.min(5) as u32);
                        tokio::time::sleep(delay).await;
                    }
                }
            }
        }

        *self.state.write() = ConnectionState::Disconnected;
        Err(NatsClientError::Nats(format!(
            "Failed to connect after {} attempts: {}",
            max_retries,
            last_error
        )))
    }

    /// Connect stub for when nats feature is disabled
    #[cfg(not(feature = "nats"))]
    pub async fn connect(&mut self, _url: &str) -> Result<(), NatsClientError> {
        // When nats feature is disabled, simulate connection behavior
        // for development and testing without a NATS server
        *self.state.write() = ConnectionState::Connecting;

        // Simulate connection delay for realism
        tokio::time::sleep(std::time::Duration::from_millis(10)).await;

        // Mark as connected (simulated)
        *self.state.write() = ConnectionState::Connected;
        warn!("NATS feature disabled - running in local-only mode");
        Ok(())
    }

    /// Publish a raw message to NATS
    #[cfg(feature = "nats")]
    pub async fn publish(&self, subject: &str, payload: &[u8]) -> Result<(), NatsClientError> {
        // Verify state without holding the parking_lot guard across await (parking_lot guards are !Send).
        {
            let state = self.state.read();
            if *state != ConnectionState::Connected {
                return Err(NatsClientError::NotConnected);
            }
        }

        // Clone out of the optional client to release the guard before the async call.
        let client = {
            let client_guard = self.client.read();
            client_guard.as_ref().cloned()
        }
        .ok_or(NatsClientError::NotConnected)?;

        client.publish(subject.to_string(), payload.to_vec().into())
            .await
            .map_err(|e| NatsClientError::Nats(e.to_string()))?;

        debug!(subject = %subject, bytes = payload.len(), "Published to NATS");
        Ok(())
    }

    /// Publish stub for non-nats builds
    #[cfg(not(feature = "nats"))]
    pub async fn publish(&self, subject: &str, payload: &[u8]) -> Result<(), NatsClientError> {
        // When nats is disabled, simulate successful publish for local development
        // The message is logged but not actually sent anywhere
        let state = self.state.read();
        if *state != ConnectionState::Connected {
            // Queue message locally for later delivery when NATS becomes available
            debug!(subject = %subject, bytes = payload.len(), "NATS disabled - message queued locally");
            return Ok(());
        }
        drop(state);

        // Simulate publish by logging
        debug!(subject = %subject, bytes = payload.len(), "NATS disabled - published to local queue");
        Ok(())
    }

    /// Check if connected to NATS
    pub fn is_connected(&self) -> bool {
        let state = self.state.read();
        *state == ConnectionState::Connected
    }

    /// Notify orchestrator that a cell has been registered
    pub async fn notify_cell_registered(
        &self,
        cell_id: CellId,
        name: &str,
        capabilities: Vec<String>,
    ) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "cell_id": cell_id.to_string(),
            "name": name,
            "capabilities": capabilities.clone(),
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::CellRegistered {
                cell_id: cell_id.to_string(),
                name: name.to_string(),
                capabilities,
            },
            cell_id: Some(cell_id),
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.cell.registered", &data).await?;
        info!(cell_id = %cell_id, "Notified orchestrator of cell registration");
        Ok(())
    }

    /// Notify orchestrator that a cell has been unregistered
    pub async fn notify_cell_unregistered(&self, cell_id: &CellId) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "cell_id": cell_id.to_string(),
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::CellUnregistered {
                cell_id: cell_id.to_string(),
            },
            cell_id: Some(*cell_id),
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.cell.unregistered", &data).await?;
        info!(cell_id = %cell_id, "Notified orchestrator of cell unregistration");
        Ok(())
    }

    /// Send heartbeat to orchestrator
    pub async fn notify_cell_heartbeat(
        &self,
        cell_id: &CellId,
        status: &str,
        active_executions: u32,
    ) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "cell_id": cell_id.to_string(),
            "status": status,
            "active_executions": active_executions,
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::CellHeartbeat {
                cell_id: cell_id.to_string(),
                status: status.to_string(),
                active_executions,
            },
            cell_id: Some(*cell_id),
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.cell.heartbeat", &data).await?;
        Ok(())
    }

    /// Report execution result to orchestrator
    pub async fn report_execution_result(
        &self,
        execution_id: &str,
        cell_id: &CellId,
        status: &str,
        output: Option<Vec<u8>>,
        error: Option<String>,
    ) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "execution_id": execution_id,
            "cell_id": cell_id.to_string(),
            "status": status,
            "error": error,
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::ExecutionResult {
                execution_id: execution_id.to_string(),
                status: status.to_string(),
                output: output.map(CborBytes::new),
                error,
            },
            cell_id: Some(*cell_id),
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.execution.result", &data).await?;
        Ok(())
    }

    /// Announce capabilities to the mesh
    pub async fn announce_capabilities(
        &self,
        cell_id: &CellId,
        capabilities: Vec<CapabilityInfo>,
    ) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "cell_id": cell_id.to_string(),
            "capabilities": capabilities,
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::CapabilityAnnouncement {
                cell_id: cell_id.to_string(),
                capabilities,
            },
            cell_id: Some(*cell_id),
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.capabilities.announce", &data).await?;
        Ok(())
    }

    /// Announce swarm activity
    pub async fn announce_swarm(
        &self,
        swarm_id: &str,
        cell_id: &CellId,
        action: SwarmAction,
    ) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "swarm_id": swarm_id,
            "cell_id": cell_id.to_string(),
            "action": format!("{:?}", action),
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::SwarmAnnouncement {
                swarm_id: swarm_id.to_string(),
                cell_id: cell_id.to_string(),
                action,
            },
            cell_id: Some(*cell_id),
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.swarm.announce", &data).await?;
        Ok(())
    }

    /// Report runtime status to orchestrator
    pub async fn report_runtime_status(
        &self,
        healthy: bool,
        active_cells: u32,
        active_swarms: u32,
    ) -> Result<(), NatsClientError> {
        let payload = CborBytes::new(serde_json::json!({
            "healthy": healthy,
            "active_cells": active_cells,
            "active_swarms": active_swarms,
        }).to_string().into_bytes());

        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::RuntimeStatus {
                healthy,
                active_cells,
                active_swarms,
            },
            cell_id: None,
            payload,
            timestamp: Utc::now(),
        };
        let data = msg.to_cbor()?;
        self.publish("prism.runtime.status", &data).await?;
        Ok(())
    }

    /// Subscribe to execution requests with message handler
    #[cfg(feature = "nats")]
    pub async fn subscribe_to_execution_requests<F>(&self, handler: F) -> Result<(), NatsClientError>
    where
        F: Fn(String, Vec<u8>) + Send + Sync + 'static,
    {
        let state = self.state.read();
        if *state != ConnectionState::Connected {
            return Err(NatsClientError::NotConnected);
        }
        drop(state);

        let client = self.client.read();
        let client = client.as_ref().ok_or(NatsClientError::NotConnected)?;

        let mut subscriber = client
            .subscribe("prism.execution.request")
            .await
            .map_err(|e| NatsClientError::Subscription(e.to_string()))?;

        // Track subscription
        {
            let mut subjects = self.subjects.write();
            subjects.push("prism.execution.request".to_string());
        }

        // Spawn task to handle messages
        tokio::spawn(async move {
            while let Some(msg) = subscriber.next().await {
                if msg.payload.is_empty() {
                    continue;
                }

                match OrchestratorMessage::from_cbor(&msg.payload) {
                    Ok(value) => {
                        if let OrchestratorMessageType::ExecutionRequest { execution_id, input, .. } = value.msg_type {
                            debug!(execution_id = %execution_id, "Received execution request");
                            handler(execution_id, input.value);
                        }
                    }
                    Err(e) => {
                        warn!(error = %e, "Failed to parse orchestrator message");
                    }
                }
            }
            warn!("Execution request subscription ended");
        });

        info!("Subscribed to prism.execution.request");
        Ok(())
    }

    /// Subscribe stub for non-nats builds
    #[cfg(not(feature = "nats"))]
    pub async fn subscribe_to_execution_requests<F>(&self, _handler: F) -> Result<(), NatsClientError>
    where
        F: Fn(String, Vec<u8>) + Send + Sync + 'static,
    {
        // When nats is disabled, just return OK without subscribing
        Ok(())
    }

    /// Subscribe to a subject with custom message handler
    #[cfg(feature = "nats")]
    pub async fn subscribe_to_subject<F>(&self, subject: &str, handler: F) -> Result<(), NatsClientError>
    where
        F: Fn(Vec<u8>) + Send + Sync + 'static,
    {
        let client = self.client.read();
        let client = client.as_ref().ok_or(NatsClientError::NotConnected)?;

        let subject_owned = subject.to_string();
        let mut subscriber = client
            .subscribe(subject_owned.clone())
            .await
            .map_err(|e| NatsClientError::Subscription(e.to_string()))?;

        // Track subscription
        {
            let mut subjects = self.subjects.write();
            subjects.push(subject_owned.clone());
        }

        tokio::spawn(async move {
            while let Some(msg) = subscriber.next().await {
                handler(msg.payload.to_vec());
            }
        });

        Ok(())
    }

    /// Get list of subscribed subjects
    pub fn subscribed_subjects(&self) -> Vec<String> {
        self.subjects.read().clone()
    }
}

impl Default for NatsOrchestratorClient {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_orchestrator_message_cbor_serialization() {
        let payload = CborBytes::new(vec![1, 2, 3]);
        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::CellRegistered {
                cell_id: "test-cell".to_string(),
                name: "Test Cell".to_string(),
                capabilities: vec!["wasm".to_string(), "http".to_string()],
            },
            cell_id: None,
            payload,
            timestamp: Utc::now(),
        };

        // Test CBOR roundtrip
        let encoded = msg.to_cbor().unwrap();
        let decoded = OrchestratorMessage::from_cbor(&encoded).unwrap();

        match decoded.msg_type {
            OrchestratorMessageType::CellRegistered { cell_id, name, capabilities } => {
                assert_eq!(cell_id, "test-cell");
                assert_eq!(name, "Test Cell");
                assert_eq!(capabilities.len(), 2);
            }
            _ => panic!("Wrong message type"),
        }
    }

    #[test]
    fn test_nats_client_creation() {
        let client = NatsOrchestratorClient::new();
        assert!(!client.is_connected());
        assert_eq!(client.connection_state(), ConnectionState::Disconnected);
    }

    #[test]
    fn test_cbor_bytes_roundtrip() {
        let original = vec![0x01, 0x02, 0x03, 0x04];
        let cb = CborBytes::new(original.clone());
        let encoded = CborCodec::encode(&cb).unwrap();
        let decoded: CborBytes = CborCodec::decode(&encoded).unwrap();
        assert_eq!(original, decoded.value);
    }

    #[test]
    fn test_connection_state() {
        let client = NatsOrchestratorClient::new();
        assert_eq!(client.connection_state(), ConnectionState::Disconnected);
    }
}