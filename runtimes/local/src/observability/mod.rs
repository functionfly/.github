//! Observability — OpenTelemetry tracing, cost attribution, trace replay.
//!
//! This module provides production-grade observability for the SAR runtime:
//!
//! ## Tracing (`tracing` + OpenTelemetry)
//!
//! Each graph node execution is wrapped in a span with attributes:
//! - `node_id`, `node_name`, `node_type`
//! - `graph_id`, `execution_id`
//! - `tenant_id`
//!
//! Spans are exported to an OTLP endpoint (e.g. Jaeger, Grafana Tempo,
//! Honeycomb). When no endpoint is configured, spans are logged to stdout
//! in structured JSON format.
//!
//! ## Cost Attribution
//!
//! `CostAttributor` tracks per-agent, per-execution-path cost by emitting
//! `AgentExecutionRecord` events to the Go backend via HTTP. Cost is
//! broken down by node type (LLM, Tool, Memory) and by model/provider
//! for LLM nodes.
//!
//! ## Metrics
//!
//! Prometheus metrics are exposed via `/metrics` endpoint (already implemented
//! in `handlers/mod.rs`). This module focuses on distributed tracing.

pub mod tracing;
pub mod cost;

pub use tracing::{init_tracing};
pub use cost::{CostAttributor, CostAttributorConfig, AgentExecutionRecord};
