//! Ruby Runtime Orchestrator Client
//!
//! NATS-based communication with the FunctionFly orchestrator.
//! Publishes registration, heartbeats, metrics and execution results
//! to the orchestrator via NATS subjects.
//!
//! ## Why `async-nats` and not `nats`
//!
//! The synchronous `nats` crate (v0.26, RUSTSEC-2024-0381) is unmaintained and
//! pulls in an old `rustls 0.22` / `rustls-webpki 0.102` that has multiple
//! outstanding CRL / name-constraint vulnerabilities (RUSTSEC-2026-0049,
//! 0098, 0099, 0104). `async-nats 0.49` uses `rustls 0.23+` with the fixes.
//! We accept the cost of making the public API async because every call site
//! is already inside a `tokio::spawn(async move { ... })` block.

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use tracing::{debug, info, warn};
use uuid::Uuid;

#[cfg(feature = "nats-client")]
use async_nats::Client as NatsClient;
#[cfg(feature = "nats-client")]
use bytes::Bytes;

/// Orchestrator client for Ruby runtime
#[derive(Clone)]
pub struct OrchestratorClient {
    runtime_id: Uuid,
    runtime_type: String,
    nats_url: Option<String>,
    #[cfg(feature = "nats-client")]
    nc: Option<NatsClient>,
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

    /// Connect to NATS. Async because `async_nats::connect` returns a future.
    #[cfg(feature = "nats-client")]
    pub async fn connect(&mut self) -> Result<()> {
        let url = self.nats_url.as_deref().unwrap_or("nats://localhost:4222");
        info!(url = %url, runtime_id = %self.runtime_id, "Connecting to NATS");

        let nc = async_nats::connect(url)
            .await
            .with_context(|| format!("Failed to connect to NATS at {}", url))?;

        self.nc = Some(nc);
        self.connected = true;
        info!(url = %url, "Connected to NATS");
        Ok(())
    }

    #[cfg(not(feature = "nats-client"))]
    pub async fn connect(&mut self) -> Result<()> {
        warn!("NATS client feature not enabled — running in standalone mode");
        Ok(())
    }

    /// Register with orchestrator
    pub async fn register_runtime(&mut self, capabilities: Vec<String>) -> Result<()> {
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
            nc.publish("runtime.registered", Bytes::from(serde_json::to_vec(&msg)?))
                .await
                .context("Failed to publish registration")?;
        }

        self.registered = true;
        Ok(())
    }

    /// Send heartbeat to orchestrator
    pub async fn send_heartbeat(&self, status: &str) -> Result<()> {
        debug!(runtime_id = %self.runtime_id, status = %status, "Sending heartbeat");

        #[cfg(feature = "nats-client")]
        if let Some(ref nc) = self.nc {
            let msg = serde_json::json!({
                "runtime_id": self.runtime_id.to_string(),
                "runtime_type": self.runtime_type,
                "status": status,
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            nc.publish("runtime.heartbeat", Bytes::from(serde_json::to_vec(&msg)?))
                .await
                .context("Failed to publish heartbeat")?;
        }

        Ok(())
    }

    /// Report metrics to orchestrator
    pub async fn report_metrics(
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
            nc.publish("runtime.metrics", Bytes::from(serde_json::to_vec(&msg)?))
                .await
                .context("Failed to publish metrics")?;
        }

        Ok(())
    }

    /// Report an execution result to the orchestrator
    pub async fn report_execution_result(
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
            nc.publish(
                "runtime.execution.result",
                Bytes::from(serde_json::to_vec(&msg)?),
            )
            .await
            .context("Failed to publish execution result")?;
        }

        Ok(())
    }

    /// Send deregistration message and close connection.
    ///
    /// `async_nats::Client` has no `close()` method — when dropped, the
    /// underlying writer task is signaled to drain pending publishes and
    /// terminate. We still issue the deregistration publish first.
    pub async fn disconnect(&mut self) {
        info!(runtime_id = %self.runtime_id, "Disconnecting from NATS");

        #[cfg(feature = "nats-client")]
        if let Some(nc) = self.nc.take() {
            let msg = serde_json::json!({
                "runtime_id": self.runtime_id.to_string(),
                "timestamp": chrono::Utc::now().to_rfc3339(),
            });
            if let Err(e) = nc
                .publish(
                    "runtime.deregistered",
                    Bytes::from(serde_json::to_vec(&msg).unwrap_or_default()),
                )
                .await
            {
                warn!(error = %e, "Failed to publish deregistration");
            }
            // Best-effort flush before dropping the client.
            if let Err(e) = nc.flush().await {
                warn!(error = %e, "Failed to flush NATS client on disconnect");
            }
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
        // We cannot await async work in Drop. Best-effort: drop the client.
        // Outstanding publishes will be queued in the runtime's writer task
        // and flushed on graceful shutdown.
        if self.connected {
            warn!(
                runtime_id = %self.runtime_id,
                "OrchestratorClient dropped while connected; outstanding publishes may not be flushed"
            );
            #[cfg(feature = "nats-client")]
            {
                self.nc.take();
            }
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
}

/// Function execution response (to orchestrator)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionExecutionResponse {
    pub execution_id: String,
    pub runtime_id: String,
    pub success: bool,
    pub output: Option<String>,
    pub error: Option<String>,
    pub execution_time_ms: u64,
    pub timestamp: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_orchestrator_client_creation() {
        let client = OrchestratorClient::new("ruby");
        assert_eq!(client.runtime_type, "ruby");
        assert!(!client.is_connected());
        assert!(!client.is_registered());
    }

    #[tokio::test]
    async fn test_with_nats_url() {
        let client = OrchestratorClient::new("ruby").with_nats_url("nats://localhost:4222");
        assert_eq!(client.nats_url, Some("nats://localhost:4222".to_string()));
    }

    #[tokio::test]
    async fn test_connect_without_feature() {
        // Without nats-client feature, connect should be a no-op success.
        // We avoid actually connecting to NATS in tests (no broker available).
        #[cfg(not(feature = "nats-client"))]
        {
            let mut client =
                OrchestratorClient::new("ruby").with_nats_url("nats://localhost:4222");
            let result = client.connect().await;
            assert!(result.is_ok());
            assert!(!client.is_connected());
        }
        #[cfg(feature = "nats-client")]
        {
            // When the feature is enabled we don't actually open a connection —
            // we just verify the URL setter works and the state transition is
            // observed. The connect() call would require a live NATS broker.
            let mut client =
                OrchestratorClient::new("ruby").with_nats_url("nats://localhost:4222");
            assert!(!client.is_connected());
            // Touch nc field to silence dead-code warnings when nats-client is on
            // and the connect path is exercised in production.
            let _ = &mut client;
        }
    }

    #[tokio::test]
    async fn test_register_runtime_without_nats() {
        // Without nats-client feature, register should succeed without publishing.
        let mut client = OrchestratorClient::new("ruby");
        let result = client.register_runtime(vec!["ruby".to_string()]).await;
        assert!(result.is_ok());
        assert!(client.is_registered());
    }

    #[tokio::test]
    async fn test_send_heartbeat_without_nats() {
        let client = OrchestratorClient::new("ruby");
        let result = client.send_heartbeat("healthy").await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_report_metrics_without_nats() {
        let client = OrchestratorClient::new("ruby");
        let result = client.report_metrics(0.5, 1024, 10).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_report_execution_result_without_nats() {
        let client = OrchestratorClient::new("ruby");
        let result = client
            .report_execution_result("exec-123", true, Some("ok"), None, 100)
            .await;
        assert!(result.is_ok());
    }
}
