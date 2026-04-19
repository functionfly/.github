//! Model router — wires the Rust SAR runtime to the Python FlyMind service.
//!
//! This module provides a client for calling the FlyMind AI service running on port 8081.
//! FlyMind handles provider selection, fallback chains, and cost tracking internally;
//! Rust just passes traffic type, messages, and timeout budget.
//!
//! ## FlyMind API
//!
//! - `POST /api/complete` — synchronous completion (JSON)
//! - `POST /api/stream` — streaming completion (SSE)
//! - `POST /api/route/decide` — ML-based edge routing
//!
//! ## Traffic Types
//!
//! Traffic is classified by the Rust graph node's `traffic_type` field, which maps
//! to FlyMind's `TrafficType` enum:
//! - `Realtime` → Groq (lowest latency)
//! - `Structured` → Fireworks (best structured output)
//! - `FunctionCalling` → Fireworks (FireAttention engine)
//! - `Background` → DeepInfra (cost-optimized)
//! - `General` → Fireworks (default)

pub mod flymind;

pub use flymind::{FlyMindClient, FlyMindConfig, LlmTrafficType, RouteResult};
