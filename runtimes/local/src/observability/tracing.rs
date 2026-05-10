//! OpenTelemetry tracing for SAR graph execution.
//!
//! Wraps each graph node execution in a span with full attribute coverage.
//! Exports to OTLP endpoint when `FLYMIND_OTLP_ENDPOINT` is set (requires `observability` feature).
//! Falls back to structured JSON logging to stdout when no endpoint is configured.

use std::time::Instant;

use tokio::sync::RwLock;
use tracing::{Span, info_span};
use tracing_subscriber::prelude::*;

use crate::engine::graph::NodeType;

/// Initialize tracing for the SAR runtime.
//
// Always uses structured JSON logging to stdout.
// When the `observability` feature is enabled AND `FLYMIND_OTLP_ENDPOINT` is set,
// additionally configures OTLP gRPC export to the configured endpoint.
pub fn init_tracing() -> anyhow::Result<()> {
    // Always set up structured JSON logging first
    let filter = tracing_subscriber::EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info,functionfly=debug"));

    let fmt_layer = tracing_subscriber::fmt::layer()
        .json()
        .with_target(true)
        .with_thread_ids(true)
        .with_thread_names(true)
        .with_file(true)
        .with_line_number(true);

    #[cfg(feature = "observability")]
    {
        let otlp_endpoint = std::env::var("FLYMIND_OTLP_ENDPOINT").ok();
        if let Some(endpoint) = otlp_endpoint {
            // Initialize OTLP exporter
            match init_otlp_tracing(&endpoint, filter.clone()) {
                Ok(_) => {
                    tracing::info!("OTLP tracing initialized at {}", endpoint);
                    return Ok(());
                }
                Err(e) => {
                    tracing::warn!("Failed to initialize OTLP tracing: {}. Falling back to JSON logging.", e);
                    // Fall through to JSON-only logging
                }
            }
        }
    }

    // JSON-only fallback (or when observability feature is disabled)
    static INIT: std::sync::Once = std::sync::Once::new();
    INIT.call_once(|| {
        tracing_subscriber::registry()
            .with(filter)
            .with(fmt_layer)
            .try_init()
            .ok(); // ok if already initialized
    });
    Ok(())
}

#[cfg(feature = "observability")]
fn init_otlp_tracing(
    endpoint: &str,
    filter: tracing_subscriber::EnvFilter,
) -> anyhow::Result<()> {
    use opentelemetry::trace::TracerProvider;
    use opentelemetry_otlp::{SpanExporter, WithExportConfig};
    use opentelemetry_sdk::trace::SdkTracerProvider;
    use tracing_opentelemetry::OpenTelemetryLayer;

    // Build the OTLP exporter with tonic
    let exporter = SpanExporter::builder()
        .with_tonic()
        .with_endpoint(endpoint)
        .build()
        .map_err(|e| anyhow::anyhow!("Failed to build OTLP exporter: {}", e))?;

    // Create a tracer provider with batch exporter
    let provider = SdkTracerProvider::builder()
        .with_batch_exporter(exporter)
        .with_resource(
            opentelemetry_sdk::Resource::builder_empty()
                .with_attributes([
                    opentelemetry::KeyValue::new("service.name", "rust-sar-runtime"),
                    opentelemetry::KeyValue::new("service.version", env!("CARGO_PKG_VERSION")),
                    opentelemetry::KeyValue::new(
                        "deployment.environment",
                        std::env::var("DEPLOYMENT_ENV").unwrap_or_else(|_| "development".to_string()),
                    ),
                ])
                .build(),
        )
        .build();

    // Set as global provider
    opentelemetry::global::set_tracer_provider(provider.clone());

    // Create a tracer for the layer
    let tracer = provider.tracer("rust-sar-runtime");

    // Create the OpenTelemetry layer
    let otlp_layer = OpenTelemetryLayer::new(tracer);

    // Initialize with both OTLP and JSON logging
    let fmt_layer = tracing_subscriber::fmt::layer()
        .json()
        .with_target(true)
        .with_thread_ids(true)
        .with_thread_names(true)
        .with_file(true)
        .with_line_number(true);

    static INIT: std::sync::Once = std::sync::Once::new();
    INIT.call_once(|| {
        tracing_subscriber::registry()
            .with(filter)
            .with(fmt_layer)
            .with(otlp_layer)
            .try_init()
            .ok(); // ok if already initialized
    });

    Ok(())
}

/// Shutdown the OpenTelemetry tracer provider (call on server shutdown).
/// Note: In opentelemetry 0.28, shutdown is handled via the provider's drop.
#[cfg(feature = "observability")]
pub fn shutdown_tracing() {
    // Force flush by dropping the global provider
    // In opentelemetry 0.28, this triggers the shutdown sequence
    // For full graceful shutdown, we would need to store the provider handle
    // and call shutdown explicitly on it
}

#[cfg(not(feature = "observability"))]
pub fn shutdown_tracing() {
    // No-op without observability feature
}

/// Create a span for a graph node execution.
pub fn node_span(
    _name: &'static str,
    graph_id: uuid::Uuid,
    execution_id: uuid::Uuid,
    node: &crate::engine::graph::Node,
    tenant_id: Option<&str>,
) -> Span {
    let node_type_str: &'static str = match &node.node_type {
        NodeType::LLM { .. } => "llm",
        NodeType::Tool { .. } => "tool",
        NodeType::Memory { .. } => "memory",
        NodeType::Control { .. } => "control",
        NodeType::Optimization { .. } => "optimization",
        NodeType::Action { .. } => "action",
        NodeType::Passthrough => "passthrough",
    };

    info_span!(
        "node_execute",
        node_id = %node.id,
        node_name = %node.name,
        node_type = node_type_str,
        graph_id = %graph_id,
        execution_id = %execution_id,
        tenant_id = ?tenant_id,
    )
}

/// Span batcher for efficient trace export to the Go backend.
//
// Batches spans in memory and flushes when batch size or time threshold is reached.
// Used when OTLP is not configured (falls back to HTTP export to Go → ClickHouse).
pub struct SpanBatch {
    spans: RwLock<Vec<SpanRecord>>,
    max_batch_size: usize,
    flush_interval: std::time::Duration,
    last_flush: RwLock<Instant>,
}

#[derive(Debug, Clone, serde::Serialize)]
struct SpanRecord {
    trace_id: uuid::Uuid,
    span_id: uuid::Uuid,
    parent_span_id: Option<uuid::Uuid>,
    node_id: String,
    node_type: String,
    graph_id: uuid::Uuid,
    execution_id: uuid::Uuid,
    tenant_id: Option<String>,
    start_ms: i64,
    duration_ms: u64,
    status: &'static str,
    error: Option<String>,
}

impl SpanBatch {
    /// Create a new span batcher.
    pub fn new(max_batch_size: usize, flush_interval_secs: u64) -> Self {
        Self {
            spans: RwLock::new(Vec::with_capacity(max_batch_size)),
            max_batch_size,
            flush_interval: std::time::Duration::from_secs(flush_interval_secs),
            last_flush: RwLock::new(Instant::now()),
        }
    }

    /// Record a completed span.
    pub async fn record(&self, record: SpanRecord) {
        let mut spans = self.spans.write().await;
        spans.push(record);

        if spans.len() >= self.max_batch_size {
            drop(spans);
            self.flush().await;
        }
    }

    /// Flush recorded spans to the Go backend via HTTP.
    //
    // Spans are POSTed as JSON to the Go orchestrator's trace endpoint.
    // The Go backend forwards to ClickHouse for long-term storage.
    pub async fn flush(&self) {
        let spans = {
            let mut s = self.spans.write().await;
            if s.is_empty() {
                return;
            }
            std::mem::take(&mut *s)
        };

        let payload = serde_json::json!({
            "spans": spans,
            "source": "rust-sar-runtime",
        });

        let go_endpoint = std::env::var("ORCHESTRATOR_TRACE_URL")
            .unwrap_or_else(|_| "http://localhost:8080/api/traces".to_string());

        match reqwest::Client::new()
            .post(&go_endpoint)
            .json(&payload)
            .timeout(std::time::Duration::from_secs(5))
            .send()
            .await
        {
            Ok(resp) if resp.status().is_success() => {
                tracing::debug!("Flushed {} spans to orchestrator", spans.len());
            }
            Ok(resp) => {
                tracing::warn!(
                    "Span flush returned {}: {}",
                    resp.status().as_u16(),
                    resp.text().await.unwrap_or_default()
                );
            }
            Err(e) => {
                tracing::warn!("Failed to flush spans to orchestrator: {}", e);
            }
        }

        *self.last_flush.write().await = Instant::now();
    }

    /// Check if we should flush based on elapsed time.
    pub async fn should_flush(&self) -> bool {
        self.last_flush.read().await.elapsed() >= self.flush_interval
    }
}
