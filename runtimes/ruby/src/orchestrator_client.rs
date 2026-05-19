//! Ruby Runtime Orchestrator Client
//!
//! NATS-based communication with the FunctionFly orchestrator.
//! Publishes registration, heartbeats, metrics and execution results
//! to the orchestrator via NATS subjects.

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use tracing::{debug, error, info, warn};
use uuid::Uuid;

/// Orchestrator client for Ruby runtime
#[derive(Clone)]
pub struct OrchestratorClient {
    runtime_id: Uuid,
    runtime_type: String,
    nats_url: Option<String>,
    #[cfg(feature = "nats-client")]
    nc: Option<nats::Connection>,
    connected: bool,
    registered: bool,
}

impl OrchestratorClient {
    /// Create a new orchestrator client
    pub fn new(runtime_type: &str) -> Self {
        Self {
            runtime_id: Uuid::new_v4(),
            runtime_type: runtime_type.to_string(),
            nats_url: None,
            #[cfg(feature = "nats-client")]
            nc: None,
            connected: false,
            registered: false,
        }
    }

    /// Set NATS URL
    pub fn with_nats_url(mut self, url: &str) -> Self {
        self.nats_url = Some(url.to_string());
        self
    }

    /// Connect to NATS
    #[cfg(feature = "nats-client")]
    pub fn connect(&mut self) -> Result<()> {
        let url = self.nats_url.as_deref().unwrap_or("nats://localhost:4222");
        info!(url = %url, runtime_id = %self.runtime_id, "Connecting to NATS");

        let nc = nats::connect(url)
            .with_context(|| format!("Failed to connect to NATS at {}", url))?;

        self.nc = Some(nc);
        self.connected = true;
        info!(url = %url, "Connected to NATS");
        Ok(())
    }

    #[cfg(not(feature = "nats-client"))]
    pub fn connect(&mut self) -> Result<()> {
        warn!("NATS client feature not enabled — running in standalone mode");
        Ok(())
    }

    /// Register with orchestrator
    pub fn register_runtime(&mut self, capabilities: Vec<String>) -> Result<()> {
        info!(
            runtime_id = %self.runtime_id,
            runtime_type = %self.runtime_type,
            ?capabilities,
            "Registering with orchestrator"
        );

        #[cfg(feature = "nats-client")]
        if let Some(ref nc) = self.nc {
            let msg = serde_json::json!({
                "runtime_id": self.runtime_id.to_string(),
                "runtime_type": self.runtime_type,
                "capabilities": capabilities,
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            nc.publish("runtime.registered", serde_json::to_vec(&msg)?)
                .context("Failed to publish registration")?;
        }

        self.registered = true;
        Ok(())
    }

    /// Send heartbeat to orchestrator
    pub fn send_heartbeat(&self, status: &str) -> Result<()> {
        debug!(runtime_id = %self.runtime_id, status = %status, "Sending heartbeat");

        #[cfg(feature = "nats-client")]
        if let Some(ref nc) = self.nc {
            let msg = serde_json::json!({
                "runtime_id": self.runtime_id.to_string(),
                "runtime_type": self.runtime_type,
                "status": status,
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            nc.publish("runtime.heartbeat", serde_json::to_vec(&msg)?)
                .context("Failed to publish heartbeat")?;
        }

        Ok(())
    }

    /// Report metrics to orchestrator
    pub fn report_metrics(
        &self,
        cpu_usage: f64,
        memory_bytes: u64,
        total_executions: u64,
    ) -> Result<()> {
        debug!(
            runtime_id = %self.runtime_id,
            cpu_usage, memory_bytes, total_executions,
            "Reporting metrics"
        );

        #[cfg(feature = "nats-client")]
        if let Some(ref nc) = self.nc {
            let msg = serde_json::json!({
                "runtime_id": self.runtime_id.to_string(),
                "runtime_type": self.runtime_type,
                "cpu_usage": cpu_usage,
                "memory_bytes": memory_bytes,
                "total_executions": total_executions,
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            nc.publish("runtime.metrics", serde_json::to_vec(&msg)?)
                .context("Failed to publish metrics")?;
        }

        Ok(())
    }

    /// Report an execution result to the orchestrator
    pub fn report_execution_result(
        &self,
        execution_id: &str,
        success: bool,
        output: Option<&str>,
        error: Option<&str>,
        execution_time_ms: u64,
    ) -> Result<()> {
        #[cfg(feature = "nats-client")]
        if let Some(ref nc) = self.nc {
            let msg = serde_json::json!({
                "execution_id": execution_id,
                "runtime_id": self.runtime_id.to_string(),
                "success": success,
                "output": output,
                "error": error,
                "execution_time_ms": execution_time_ms,
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            nc.publish("runtime.execution.result", serde_json::to_vec(&msg)?)
                .context("Failed to publish execution result")?;
        }

        Ok(())
    }

    /// Send deregistration message and close connection
    pub fn disconnect(&mut self) {
        info!(runtime_id = %self.runtime_id, "Disconnecting from NATS");

        #[cfg(feature = "nats-client")]
        if let Some(ref nc) = self.nc {
            let msg = serde_json::json!({
                "runtime_id": self.runtime_id.to_string(),
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            if let Err(e) = nc.publish("runtime.deregistered", serde_json::to_vec(&msg).unwrap_or_default()) {
                warn!(error = %e, "Failed to publish deregistration");
            }
            nc.close();
            self.nc = None;
        }

        self.connected = false;
        self.registered = false;
    }

    /// Get runtime ID
    pub fn runtime_id(&self) -> Uuid {
        self.runtime_id
    }

    /// Check if connected
    pub fn is_connected(&self) -> bool {
        self.connected
    }

    /// Check if registered
    pub fn is_registered(&self) -> bool {
        self.registered
    }
}

impl Drop for OrchestratorClient {
    fn drop(&mut self) {
        if self.connected {
            self.disconnect();
        }
    }
}

/// Function execution request (from orchestrator)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionExecutionRequest {
    pub execution_id: String,
    pub function_id: String,
    pub version: String,
    pub code: String,
    pub input: Option<serde_json::Value>,
    pub timeout_ms: Option<u64>,
    pub tenant_id: Option<String>,
}

/// Function execution response (to orchestrator)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionExecutionResponse {
    pub execution_id: String,
    pub success: bool,
    pub output: Option<String>,
    pub error: Option<String>,
    pub execution_time_ms: u64,
    pub memory_used_mb: Option<f64>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_orchestrator_client_creation() {
        let client = OrchestratorClient::new("ruby");
        assert!(!client.is_connected());
        assert!(!client.is_registered());
    }

    #[test]
    fn test_orchestrator_client_with_nats() {
        let client = OrchestratorClient::new("ruby").with_nats_url("nats://localhost:4222");
        assert_eq!(client.nats_url, Some("nats://localhost:4222".to_string()));
    }
}
