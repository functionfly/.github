use std::sync::Arc;
use chrono::Utc;
use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::info;
use uuid::Uuid;

#[cfg(feature = "nats-events")]
use nats::Connection;

use crate::core::AgentId;
use crate::events::Event;

#[derive(Error, Debug)]
pub enum NatsClientError {
    #[error("NATS error: {0}")]
    Nats(String),
    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),
    #[error("Not connected")]
    NotConnected,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrchestratorMessage {
    pub id: Uuid,
    pub msg_type: OrchestratorMessageType,
    pub agent_id: Option<AgentId>,
    pub payload: serde_json::Value,
    pub timestamp: chrono::DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum OrchestratorMessageType {
    AgentRegistered { name: String },
    AgentUnregistered { agent_id: String },
    AgentHeartbeat { agent_id: String, status: String },
    ExecutionRequest { execution_id: String, agent_id: String, input: serde_json::Value },
    ExecutionResult { execution_id: String, status: String, output: Option<serde_json::Value> },
    RuntimeStatus { healthy: bool, active_agents: u32 },
}

pub struct NatsOrchestratorClient {
    connection: Arc<RwLock<Option<Connection>>>,
}

impl NatsOrchestratorClient {
    pub fn new() -> Self {
        Self {
            connection: Arc::new(RwLock::new(None)),
        }
    }

    #[cfg(feature = "nats-events")]
    pub fn connect(&mut self, url: &str) -> Result<(), NatsClientError> {
        let conn = nats::connect(url)
            .map_err(|e| NatsClientError::Nats(e.to_string()))?;
        *self.connection.write() = Some(conn);
        info!("Connected to NATS at {}", url);
        Ok(())
    }

    #[cfg(not(feature = "nats-events"))]
    pub fn connect(&mut self, _url: &str) -> Result<(), NatsClientError> {
        Ok(())
    }

    #[cfg(feature = "nats-events")]
    pub fn publish(&self, subject: &str, payload: &[u8]) -> Result<(), NatsClientError> {
        let conn = self.connection.read();
        let conn = conn.as_ref().ok_or(NatsClientError::NotConnected)?;
        conn.publish(subject, payload)
            .map_err(|e| NatsClientError::Nats(e.to_string()))?;
        Ok(())
    }

    #[cfg(not(feature = "nats-events"))]
    pub fn publish(&self, _subject: &str, _payload: &[u8]) -> Result<(), NatsClientError> {
        Ok(())
    }

    pub fn notify_agent_registered(&self, agent_id: AgentId, name: &str) -> Result<(), NatsClientError> {
        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::AgentRegistered { name: name.to_string() },
            agent_id: Some(agent_id),
            payload: serde_json::json!({ "agent_id": agent_id.to_string(), "name": name }),
            timestamp: Utc::now(),
        };
        let data = serde_json::to_vec(&msg)?;
        self.publish("orchestrator.agent.registered", &data)?;
        info!(agent_id = %agent_id, "Notified orchestrator of agent registration");
        Ok(())
    }

    pub fn notify_agent_unregistered(&self, agent_id: &AgentId) -> Result<(), NatsClientError> {
        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::AgentUnregistered { agent_id: agent_id.to_string() },
            agent_id: Some(*agent_id),
            payload: serde_json::json!({ "agent_id": agent_id.to_string() }),
            timestamp: Utc::now(),
        };
        let data = serde_json::to_vec(&msg)?;
        self.publish("orchestrator.agent.unregistered", &data)?;
        info!(agent_id = %agent_id, "Notified orchestrator of agent unregistration");
        Ok(())
    }

    pub fn notify_agent_heartbeat(&self, agent_id: &AgentId, status: &str) -> Result<(), NatsClientError> {
        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::AgentHeartbeat {
                agent_id: agent_id.to_string(),
                status: status.to_string()
            },
            agent_id: Some(*agent_id),
            payload: serde_json::json!({ "agent_id": agent_id.to_string(), "status": status }),
            timestamp: Utc::now(),
        };
        let data = serde_json::to_vec(&msg)?;
        self.publish("orchestrator.agent.heartbeat", &data)?;
        Ok(())
    }

    #[cfg(feature = "nats-events")]
    pub fn start_event_listener<F>(&self, handler: F) -> Result<(), NatsClientError>
    where
        F: Fn(Event) + Send + Sync + 'static,
    {
        let conn = self.connection.read();
        let conn = conn.as_ref().ok_or(NatsClientError::NotConnected)?;

        let subscription = conn.subscribe("execution.request")
            .map_err(|e| NatsClientError::Nats(e.to_string()))?;

        std::thread::spawn(move || {
            for msg in subscription.messages() {
                let data: Result<serde_json::Value, _> = serde_json::from_slice(msg.data.as_ref());
                if let Ok(data) = data {
                    if let Some(input) = data.get("input") {
                        if let Some(execution_id) = data.get("execution_id").and_then(|v| v.as_str()) {
                            handler(Event::new(
                                crate::events::EventSource::Api,
                                "execution.request".to_string(),
                                serde_json::json!({
                                    "execution_id": execution_id,
                                    "input": input,
                                }),
                            ));
                        }
                    }
                }
            }
        });

        Ok(())
    }

    #[cfg(not(feature = "nats-events"))]
    pub fn start_event_listener<F>(&self, _handler: F) -> Result<(), NatsClientError>
    where
        F: Fn(Event) + Send + Sync + 'static,
    {
        Ok(())
    }

    pub fn report_runtime_status(&self, healthy: bool, active_agents: u32) -> Result<(), NatsClientError> {
        let msg = OrchestratorMessage {
            id: Uuid::new_v4(),
            msg_type: OrchestratorMessageType::RuntimeStatus { healthy, active_agents },
            agent_id: None,
            payload: serde_json::json!({
                "healthy": healthy,
                "active_agents": active_agents,
            }),
            timestamp: Utc::now(),
        };
        let data = serde_json::to_vec(&msg)?;
        self.publish("orchestrator.runtime.status", &data)?;
        Ok(())
    }
}

impl Default for NatsOrchestratorClient {
    fn default() -> Self {
        Self::new()
    }
}