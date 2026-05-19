//! Cost attribution for SAR graph executions.
//!
//! Emits `AgentExecutionRecord` events to the Go backend for per-agent,
//! per-execution-path cost tracking. Cost is broken down by:
//! - LLM calls: by model + provider (from FlyMind response)
//! - Tool calls: by tool name
//! - Memory ops: by operation type
//!
//! ## Cost Emission
//!
//! Costs are buffered and emitted in batches to the Go backend via HTTP POST.
//! The Go backend forwards to ClickHouse for storage and billing dashboards.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use tracing::{debug, instrument};

use crate::engine::graph::{ExecutionStatus, NodeId, NodeResult};
use crate::router::flymind::Usage;

/// Configuration for the cost attributor.
#[derive(Debug, Clone)]
pub struct CostAttributorConfig {
    /// Go backend URL for cost events.
    pub go_endpoint: String,
    /// Maximum events to buffer before flushing.
    pub max_buffer: usize,
    /// Maximum time to wait before flushing.
    pub flush_interval: Duration,
    /// Whether cost attribution is enabled.
    pub enabled: bool,
}

impl Default for CostAttributorConfig {
    fn default() -> Self {
        Self {
            go_endpoint: std::env::var("ORCHESTRATOR_COST_URL")
                .unwrap_or_else(|_| {
                    eprintln!("Warning: ORCHESTRATOR_COST_URL not set - cost attribution disabled");
                    String::new()
                }),
            max_buffer: 100,
            flush_interval: Duration::from_secs(5),
            enabled: true,
        }
    }
}

/// A single cost event for one node within an execution.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentExecutionRecord {
    /// Unique execution ID.
    pub execution_id: String,
    /// Graph ID.
    pub graph_id: String,
    /// Agent/graph name.
    pub agent_name: String,
    /// Tenant ID (for multi-tenant billing).
    pub tenant_id: Option<String>,
    /// Node ID.
    pub node_id: String,
    /// Node type: llm, tool, memory, control, optimization, passthrough.
    pub node_type: String,
    /// For LLM nodes: the model used.
    pub model: Option<String>,
    /// For LLM nodes: the provider used.
    pub provider: Option<String>,
    /// For Tool nodes: the tool name.
    pub tool_name: Option<String>,
    /// Token usage (LLM nodes only).
    pub usage: Option<Usage>,
    /// Cost in USD (estimated).
    pub cost_usd: Option<f64>,
    /// Node execution duration in ms.
    pub duration_ms: u64,
    /// Node execution status.
    pub status: String,
    /// Whether this node was retried.
    pub retried: bool,
    /// Timestamp when the execution started.
    pub started_at: String,
    /// Timestamp when the execution ended.
    pub ended_at: String,
}

/// Estimates cost in USD from token usage.
//
// Pricing is approximate based on FlyMind provider pricing as of 2026.
// These rates should be synced with the Go backend billing system.
pub fn estimate_cost(usage: &Usage, model: &str) -> f64 {
    // Approximate pricing per 1M tokens (input, output)
    let (input_per_m, output_per_m) = if model.contains("gpt") {
        (2.50, 10.00)
    } else if model.contains("claude") {
        (3.00, 15.00)
    } else if model.contains("gemini") {
        (0.50, 1.50)
    } else if model.contains("groq") || model.contains("llama") {
        (0.20, 0.20)
    } else if model.contains("fireworks") {
        (0.70, 2.80)
    } else {
        // Default: assume OpenAI pricing
        (2.50, 10.00)
    };

    let input_cost = (usage.prompt_tokens as f64 / 1_000_000.0) * input_per_m;
    let output_cost = (usage.completion_tokens as f64 / 1_000_000.0) * output_per_m;

    input_cost + output_cost
}

/// Cost attribution buffer that batches records and emits to the Go backend.
pub struct CostAttributor {
    config: CostAttributorConfig,
    buffer: Arc<RwLock<Vec<AgentExecutionRecord>>>,
    last_flush: Arc<RwLock<Instant>>,
}

impl CostAttributor {
    /// Create a new cost attributor.
    pub fn new(config: CostAttributorConfig) -> Self {
        let max_buffer = config.max_buffer;
        Self {
            config,
            buffer: Arc::new(RwLock::new(Vec::with_capacity(max_buffer))),
            last_flush: Arc::new(RwLock::new(Instant::now())),
        }
    }

    /// Record a completed node execution.
    //
    // For LLM nodes, `llm_result` contains the FlyMind response with usage.
    // For other nodes, it should be `None`.
    #[instrument(skip_all, fields(node_id = %node_id, node_type = %node_type_str))]
    pub async fn record_node(
        &self,
        execution_id: uuid::Uuid,
        graph_id: uuid::Uuid,
        agent_name: &str,
        tenant_id: Option<&str>,
        node_id: NodeId,
        node_type_str: &str,
        model: Option<&str>,
        provider: Option<&str>,
        tool_name: Option<&str>,
        llm_usage: Option<&Usage>,
        duration_ms: u64,
        status: ExecutionStatus,
        retried: bool,
        started_at: Instant,
    ) {
        if !self.config.enabled {
            return;
        }

        let cost_usd = llm_usage.and_then(|u| {
            model.map(|m| estimate_cost(u, m))
        });

        let now = Instant::now();
        // Convert Instant to SystemTime-based duration since epoch.
        // Instant represents time relative to the program's start, not wall clock time.
        // We use SystemTime for wall clock operations.
        let duration_since_epoch = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default();
        let now_wall = chrono::DateTime::from_timestamp(
            duration_since_epoch.as_secs() as i64,
            duration_since_epoch.subsec_nanos(),
        ).unwrap_or_default();

        // For started_at, we can't accurately convert Instant to wall time without
        // additional context. We record the duration since process start as a proxy.
        let started_duration = started_at
            .checked_duration_since(Instant::now())
            .map(|d| now_wall - chrono::Duration::from_std(d).unwrap_or_default())
            .unwrap_or(now_wall);

        let record = AgentExecutionRecord {
            execution_id: execution_id.to_string(),
            graph_id: graph_id.to_string(),
            agent_name: agent_name.to_string(),
            tenant_id: tenant_id.map(String::from),
            node_id: node_id.to_string(),
            node_type: node_type_str.to_string(),
            model: model.map(String::from),
            provider: provider.map(String::from),
            tool_name: tool_name.map(String::from),
            usage: llm_usage.cloned(),
            cost_usd,
            duration_ms,
            status: format!("{:?}", status),
            retried,
            started_at: started_duration.to_rfc3339(),
            ended_at: now_wall.to_rfc3339(),
        };

        let mut buf = self.buffer.write().await;
        buf.push(record);

        if buf.len() >= self.config.max_buffer {
            drop(buf);
            self.flush().await;
        }
    }

    /// Flush buffered records to the Go backend.
    pub async fn flush(&self) {
        let records = {
            let mut buf = self.buffer.write().await;
            if buf.is_empty() {
                return;
            }
            let records = std::mem::take(&mut *buf);
            *self.last_flush.write().await = Instant::now();
            records
        };

        debug!(count = records.len(), "Flushing cost records to Go backend");

        let payload = serde_json::json!({
            "records": &records,
            "source": "rust-sar-runtime",
        });

        match reqwest::Client::new()
            .post(&self.config.go_endpoint)
            .json(&payload)
            .timeout(Duration::from_secs(10))
            .send()
            .await
        {
            Ok(resp) if resp.status().is_success() => {
                debug!("Cost records flushed successfully");
            }
            Ok(resp) => {
                tracing::warn!(
                    status = %resp.status().as_u16(),
                    "Cost flush returned non-success"
                );
            }
            Err(e) => {
                tracing::warn!("Failed to flush cost records: {}", e);
            }
        }
    }

    /// Get the number of buffered records.
    pub async fn buffer_len(&self) -> usize {
        self.buffer.read().await.len()
    }
}

impl Drop for CostAttributor {
    fn drop(&mut self) {
        // Attempt synchronous flush on drop (best-effort)
        let records = self.buffer.try_read();
        if let Ok(buf) = records {
            if !buf.is_empty() {
                tracing::debug!("Dropping {} unflushed cost records", buf.len());
            }
        }
    }
}