//! Orchestrator client for Kotlin/JVM runtime
//!
//! Handles communication with the FunctionFly orchestrator via NATS
//! for registration, heartbeat, and function execution messages.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use uuid::Uuid;

/// NATS configuration
#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct NatsConfig {
    /// NATS server URL
    pub url: String,
    /// Runtime identifier
    pub runtime_id: String,
    /// Runtime type
    pub runtime_type: String,
    /// Runtime version
    pub runtime_version: String,
}

impl Default for NatsConfig {
    fn default() -> Self {
        Self {
            url: std::env::var("NATS_URL").unwrap_or_else(|_| "nats://localhost:4222".to_string()),
            runtime_id: Uuid::new_v4().to_string(),
            runtime_type: "kotlin".to_string(),
            runtime_version: env!("CARGO_PKG_VERSION").to_string(),
        }
    }
}

/// Message types for orchestrator communication
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", content = "data")]
pub enum OrchestratorMessage {
    /// Runtime registration
    RuntimeRegistered {
        runtime_id: String,
        runtime_type: String,
        version: String,
        capabilities: Vec<String>,
    },
    /// Runtime heartbeat
    RuntimeHeartbeat {
        runtime_id: String,
        status: RuntimeStatus,
    },
    /// Runtime metrics report
    RuntimeMetrics {
        runtime_id: String,
        total_executions: u64,
        successful_executions: u64,
        failed_executions: u64,
        avg_execution_time_ms: f64,
        current_memory_mb: f64,
    },
    /// Function execution request
    FunctionExecutionRequest {
        execution_id: Uuid,
        function_id: String,
        code: String,
        input: Option<serde_json::Value>,
        timeout_ms: Option<u64>,
    },
    /// Function execution result
    FunctionExecutionResult {
        execution_id: Uuid,
        success: bool,
        output: Option<serde_json::Value>,
        error: Option<String>,
        execution_time_ms: u64,
    },
    /// Runtime deregistration
    RuntimeDeregistered {
        runtime_id: String,
        reason: String,
    },
}

/// Runtime status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RuntimeStatus {
    /// Running and ready
    Ready,
    /// Running but busy
    Busy,
    /// Shutting down
    ShuttingDown,
    /// Error state
    Error(String),
}

/// Orchestrator client for NATS communication
pub struct OrchestratorClient {
    config: NatsConfig,
    nats_conn: Option<Arc<RwLock<nats::Connection>>>,
    subscriptions: Arc<RwLock<Vec<nats::Subscription>>>,
}

impl OrchestratorClient {
    /// Create a new orchestrator client
    pub fn new(config: NatsConfig) -> Self {
        Self {
            config,
            nats_conn: None,
            subscriptions: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Create with default configuration
    pub fn with_defaults() -> Self {
        Self::new(NatsConfig::default())
    }

    /// Connect to NATS
    pub async fn connect(&mut self) -> Result<()> {
        let nats_url = &self.config.url;

        let opts = nats::Options::new()
            .with_name(&self.config.runtime_id);

        let conn = opts.connect(nats_url)?;

        self.nats_conn = Some(Arc::new(RwLock::new(conn)));
        tracing::info!("Connected to NATS at {}", nats_url);

        Ok(())
    }

    /// Check if connected to NATS
    pub fn is_connected(&self) -> bool {
        self.nats_conn.is_some()
    }

    /// Publish a message to a subject
    pub async fn publish(&self, subject: &str, msg: &OrchestratorMessage) -> Result<()> {
        let conn = self.nats_conn.as_ref()
            .ok_or_else(|| anyhow::anyhow!("not connected to NATS"))?;

        let conn = conn.read().await;
        let payload = serde_json::to_vec(msg)?;

        conn.publish(subject, &payload)?;

        Ok(())
    }

    /// Subscribe to a subject
    pub async fn subscribe(&mut self, subject: &str) -> Result<()> {
        let conn = self.nats_conn.as_ref()
            .ok_or_else(|| anyhow::anyhow!("not connected to NATS"))?;

        let conn = conn.read().await;
        let sub = conn.subscribe(subject)?;

        let mut subs = self.subscriptions.write().await;
        subs.push(sub);

        Ok(())
    }

    /// Register this runtime with the orchestrator
    pub async fn register(&self) -> Result<()> {
        let capabilities = vec![
            "kotlin".to_string(),
            "jvm".to_string(),
            "secure-execution".to_string(),
            "resource-limits".to_string(),
        ];

        let msg = OrchestratorMessage::RuntimeRegistered {
            runtime_id: self.config.runtime_id.clone(),
            runtime_type: self.config.runtime_type.clone(),
            version: self.config.runtime_version.clone(),
            capabilities,
        };

        self.publish("runtime.registered", &msg).await?;
        tracing::info!("Registered as runtime {}", self.config.runtime_id);

        Ok(())
    }

    /// Send heartbeat to orchestrator
    pub async fn heartbeat(&self, status: RuntimeStatus) -> Result<()> {
        let msg = OrchestratorMessage::RuntimeHeartbeat {
            runtime_id: self.config.runtime_id.clone(),
            status,
        };

        self.publish("runtime.heartbeat", &msg).await?;
        Ok(())
    }

    /// Report metrics to orchestrator
    pub async fn report_metrics(
        &self,
        total_executions: u64,
        successful_executions: u64,
        failed_executions: u64,
        avg_execution_time_ms: f64,
        current_memory_mb: f64,
    ) -> Result<()> {
        let msg = OrchestratorMessage::RuntimeMetrics {
            runtime_id: self.config.runtime_id.clone(),
            total_executions,
            successful_executions,
            failed_executions,
            avg_execution_time_ms,
            current_memory_mb,
        };

        self.publish("runtime.metrics", &msg).await?;
        Ok(())
    }

    /// Send execution result to orchestrator
    pub async fn send_execution_result(
        &self,
        execution_id: Uuid,
        success: bool,
        output: Option<serde_json::Value>,
        error: Option<String>,
        execution_time_ms: u64,
    ) -> Result<()> {
        let msg = OrchestratorMessage::FunctionExecutionResult {
            execution_id,
            success,
            output,
            error,
            execution_time_ms,
        };

        self.publish("runtime.execution.result", &msg).await?;
        Ok(())
    }

    /// Deregister this runtime
    pub async fn deregister(&self, reason: String) -> Result<()> {
        let msg = OrchestratorMessage::RuntimeDeregistered {
            runtime_id: self.config.runtime_id.clone(),
            reason,
        };

        self.publish("runtime.deregistered", &msg).await?;
        tracing::info!("Deregistered runtime {}", self.config.runtime_id);

        Ok(())
    }

    /// Get the runtime ID
    pub fn runtime_id(&self) -> &str {
        &self.config.runtime_id
    }

    /// Get the NATS configuration
    pub fn config(&self) -> &NatsConfig {
        &self.config
    }
}

/// Start heartbeat loop
pub async fn start_heartbeat_loop(
    client: Arc<RwLock<OrchestratorClient>>,
    interval_secs: u64,
) -> Result<()> {
    let interval = Duration::from_secs(interval_secs);

    let mut interval_timer = tokio::time::interval(interval);

    loop {
        interval_timer.tick().await;

        let client = client.read().await;
        if !client.is_connected() {
            continue;
        }

        let status = RuntimeStatus::Ready;
        if let Err(e) = client.heartbeat(status).await {
            tracing::warn!("Failed to send heartbeat: {}", e);
        }
    }
}

/// Start metrics reporting loop
pub async fn start_metrics_loop(
    client: Arc<RwLock<OrchestratorClient>>,
    metrics: Arc<crate::metrics::MetricsCollector>,
    interval_secs: u64,
) -> Result<()> {
    let interval = Duration::from_secs(interval_secs);

    let mut interval_timer = tokio::time::interval(interval);

    loop {
        interval_timer.tick().await;

        let client = client.read().await;
        if !client.is_connected() {
            continue;
        }

        let runtime_metrics = metrics.get_metrics().await;

        if let Err(e) = client.report_metrics(
            runtime_metrics.total_executions,
            runtime_metrics.successful_executions,
            runtime_metrics.failed_executions,
            runtime_metrics.avg_execution_time_ms,
            runtime_metrics.current_memory_mb,
        ).await {
            tracing::warn!("Failed to report metrics: {}", e);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_nats_config_default() {
        let config = NatsConfig::default();
        assert_eq!(config.runtime_type, "kotlin");
        assert!(!config.runtime_id.is_empty());
    }

    #[test]
    fn test_orchestrator_message_serialization() {
        let msg = OrchestratorMessage::RuntimeRegistered {
            runtime_id: "test-123".to_string(),
            runtime_type: "kotlin".to_string(),
            version: "0.1.0".to_string(),
            capabilities: vec!["kotlin".to_string()],
        };

        let json = serde_json::to_string(&msg).unwrap();
        assert!(json.contains("RuntimeRegistered"));
        assert!(json.contains("test-123"));
    }

    #[test]
    fn test_execution_result_message() {
        let msg = OrchestratorMessage::FunctionExecutionResult {
            execution_id: Uuid::new_v4(),
            success: true,
            output: Some(serde_json::json!({"result": "ok"})),
            error: None,
            execution_time_ms: 100,
        };

        let json = serde_json::to_string(&msg).unwrap();
        assert!(json.contains("FunctionExecutionResult"));
        assert!(json.contains("success"));
    }
}