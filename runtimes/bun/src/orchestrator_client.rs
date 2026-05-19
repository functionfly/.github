//! Orchestrator client for Bun runtime
//!
//! Handles runtime registration, function dispatch, heartbeat, and metrics
//! reporting to the FunctionFly orchestrator via NATS.

use crate::config::{ExecutionLimits, RuntimeConfig};
use crate::sandbox::{Sandbox, SandboxConfig, SandboxResult};
use crate::security::SecurityManager;
use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use uuid::Uuid;

#[cfg(feature = "nats-client")]
use nats::Connection;

#[cfg(feature = "nats-client")]
use serde_json::json;

/// Message types for orchestrator communication
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum OrchestratorMessageType {
    RuntimeRegistered { runtime_id: String, runtime_type: String },
    RuntimeDeregistered { runtime_id: String },
    RuntimeHeartbeat { runtime_id: String, status: String },
    FunctionExecutionRequest { execution_id: String, function_id: String, input: serde_json::Value },
    FunctionExecutionResult { execution_id: String, status: String, output: Option<serde_json::Value>, error: Option<String> },
    RuntimeMetrics { runtime_id: String, cpu_percent: f64, memory_bytes: u64, total_executions: u64 },
}

/// Message envelope for orchestrator communication
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrchestratorMessage {
    pub id: Uuid,
    pub msg_type: OrchestratorMessageType,
    pub runtime_id: Option<String>,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl OrchestratorMessage {
    pub fn new(msg_type: OrchestratorMessageType) -> Self {
        Self {
            id: Uuid::new_v4(),
            msg_type,
            runtime_id: None,
            timestamp: chrono::Utc::now(),
        }
    }

    pub fn with_runtime_id(mut self, runtime_id: &str) -> Self {
        self.runtime_id = Some(runtime_id.to_string());
        self
    }
}

/// Orchestrator client for runtime-orchestrator communication
pub struct OrchestratorClient {
    runtime_id: String,
    runtime_type: String,
    nats_url: Option<String>,
    #[cfg(feature = "nats-client")]
    connection: Option<Connection>,
    registered: bool,
}

impl OrchestratorClient {
    /// Create a new orchestrator client
    pub fn new(runtime_type: &str) -> Self {
        Self {
            runtime_id: format!("bun-{}", Uuid::new_v4().to_string().split('-').next().unwrap()),
            runtime_type: runtime_type.to_string(),
            nats_url: None,
            #[cfg(feature = "nats-client")]
            connection: None,
            registered: false,
        }
    }

    /// Set the NATS URL for connection
    pub fn with_nats_url(mut self, url: &str) -> Self {
        self.nats_url = Some(url.to_string());
        self
    }

    /// Connect to NATS
    #[cfg(feature = "nats-client")]
    pub fn connect(&mut self) -> Result<()> {
        if let Some(ref url) = self.nats_url {
            let conn = nats::connect(url)
                .map_err(|e| anyhow!("failed to connect to NATS: {}", e))?;
            self.connection = Some(conn);
            Ok(())
        } else {
            Err(anyhow!("no NATS URL configured"))
        }
    }

    #[cfg(not(feature = "nats-client"))]
    pub fn connect(&mut self) -> Result<()> {
        Ok(())
    }

    /// Check if client is connected
    pub fn is_connected(&self) -> bool {
        #[cfg(feature = "nats-client")]
        return self.connection.is_some();
        #[cfg(not(feature = "nats-client"))]
        return false;
    }

    /// Register this runtime with the orchestrator
    #[cfg(feature = "nats-client")]
    pub fn register_runtime(&mut self, functions: Vec<String>) -> Result<()> {
        let connection = self.connection.as_ref()
            .ok_or_else(|| anyhow!("not connected to NATS"))?;

        let msg = OrchestratorMessage::new(
            OrchestratorMessageType::RuntimeRegistered {
                runtime_id: self.runtime_id.clone(),
                runtime_type: self.runtime_type.clone(),
            },
        )
        .with_runtime_id(&self.runtime_id);

        let payload = serde_json::to_vec(&msg)?;
        connection.publish("runtime.registered", &payload)
            .map_err(|e| anyhow!("failed to publish registration: {}", e))?;

        // Also publish available functions
        for func in functions {
            let func_msg = json!({
                "type": "FunctionAvailable",
                "runtime_id": self.runtime_id,
                "function": func,
            });
            let func_payload = serde_json::to_vec(&func_msg)?;
            connection.publish("runtime.functions.available", &func_payload)
                .map_err(|e| anyhow!("failed to publish function availability: {}", e))?;
        }

        self.registered = true;
        Ok(())
    }

    #[cfg(not(feature = "nats-client"))]
    pub fn register_runtime(&mut self, _functions: Vec<String>) -> Result<()> {
        self.registered = true;
        Ok(())
    }

    /// Send heartbeat to orchestrator
    #[cfg(feature = "nats-client")]
    pub fn send_heartbeat(&self, status: &str) -> Result<()> {
        let connection = self.connection.as_ref()
            .ok_or_else(|| anyhow!("not connected to NATS"))?;

        let msg = OrchestratorMessage::new(
            OrchestratorMessageType::RuntimeHeartbeat {
                runtime_id: self.runtime_id.clone(),
                status: status.to_string(),
            },
        )
        .with_runtime_id(&self.runtime_id);

        let payload = serde_json::to_vec(&msg)?;
        connection.publish("runtime.heartbeat", &payload)?;
        Ok(())
    }

    #[cfg(not(feature = "nats-client"))]
    pub fn send_heartbeat(&self, _status: &str) -> Result<()> {
        Ok(())
    }

    /// Report metrics to orchestrator
    #[cfg(feature = "nats-client")]
    pub fn report_metrics(&self, cpu_percent: f64, memory_bytes: u64, total_executions: u64) -> Result<()> {
        let connection = self.connection.as_ref()
            .ok_or_else(|| anyhow!("not connected to NATS"))?;

        let msg = OrchestratorMessage::new(
            OrchestratorMessageType::RuntimeMetrics {
                runtime_id: self.runtime_id.clone(),
                cpu_percent,
                memory_bytes,
                total_executions,
            },
        )
        .with_runtime_id(&self.runtime_id);

        let payload = serde_json::to_vec(&msg)?;
        connection.publish("runtime.metrics", &payload)?;
        Ok(())
    }

    #[cfg(not(feature = "nats-client"))]
    pub fn report_metrics(&self, _cpu_percent: f64, _memory_bytes: u64, _total_executions: u64) -> Result<()> {
        Ok(())
    }

    /// Deregister this runtime from the orchestrator
    #[cfg(feature = "nats-client")]
    pub fn deregister_runtime(&self) -> Result<()> {
        let connection = self.connection.as_ref()
            .ok_or_else(|| anyhow!("not connected to NATS"))?;

        let msg = OrchestratorMessage::new(
            OrchestratorMessageType::RuntimeDeregistered {
                runtime_id: self.runtime_id.clone(),
            },
        )
        .with_runtime_id(&self.runtime_id);

        let payload = serde_json::to_vec(&msg)?;
        connection.publish("runtime.deregistered", &payload)?;
        Ok(())
    }

    #[cfg(not(feature = "nats-client"))]
    pub fn deregister_runtime(&self) -> Result<()> {
        Ok(())
    }

    /// Get the runtime ID
    pub fn runtime_id(&self) -> &str {
        &self.runtime_id
    }

    /// Check if registered
    pub fn is_registered(&self) -> bool {
        self.registered
    }
}

impl Default for OrchestratorClient {
    fn default() -> Self {
        Self::new("bun")
    }
}

/// Function execution request from orchestrator
#[derive(Debug, Clone, Deserialize)]
pub struct FunctionExecutionRequest {
    pub execution_id: String,
    pub function_id: String,
    pub code: String,
    pub input: Option<serde_json::Value>,
    pub timeout_ms: Option<u64>,
    pub memory_mb: Option<u64>,
}

/// Result of function execution
#[derive(Debug, Clone, Serialize)]
pub struct FunctionExecutionResponse {
    pub execution_id: String,
    pub success: bool,
    pub output: Option<serde_json::Value>,
    pub error: Option<String>,
    pub execution_time_ms: u64,
    pub memory_used_mb: Option<u64>,
}

impl FunctionExecutionResponse {
    pub fn success(execution_id: &str, output: serde_json::Value, execution_time_ms: u64) -> Self {
        Self {
            execution_id: execution_id.to_string(),
            success: true,
            output: Some(output),
            error: None,
            execution_time_ms,
            memory_used_mb: None,
        }
    }

    pub fn error(execution_id: &str, error: String, execution_time_ms: u64) -> Self {
        Self {
            execution_id: execution_id.to_string(),
            success: false,
            output: None,
            error: Some(error),
            execution_time_ms,
            memory_used_mb: None,
        }
    }
}