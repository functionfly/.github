//! Shared types for all HTTP handlers.

use axum::response::IntoResponse;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

use crate::agent_scheduler::agent_scheduler::AgentScheduler;
use crate::agent_scheduler::worker::JobTracker;
pub use crate::cache::ResultCache;
pub use crate::config::Config;
use crate::engine::WasmCellExecutor;
pub use crate::engine::WasmEngine;
use crate::enterprise_security::EnterpriseSecurityEnforcer;
use crate::errors::ErrorRecovery;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::memory::layer::MemoryLayer;
use crate::micropython::MicroPythonExecutor;
use crate::monitoring::ResourceMonitor;
use crate::observability::cost::{CostAttributor, CostAttributorConfig};
use crate::optimizer::optimizer::GraphOptimizer;
use crate::orchestrator_client::OrchestratorClient;
use crate::package::PackageManager;
pub use crate::pool::{InstancePool, PoolManager};
use crate::python_pool::PythonRuntimePool;
use crate::resource_enforcer::ResourceEnforcer;
use crate::router::flymind::FlyMindClient;
use crate::scheduler::BinPackingScheduler;
use crate::security::SecurityMonitor;
use crate::shutdown::{ResourceManager, ShutdownCoordinator};
use crate::yara_scanner::YaraScanner;

// ---------------------------------------------------------------------------
// Application State
// ---------------------------------------------------------------------------

/// Application state shared across all handlers.
#[derive(Clone)]
pub struct AppState {
    pub engine: Arc<WasmEngine>,
    pub pool: Arc<tokio::sync::RwLock<InstancePool>>,
    pub cache: Arc<tokio::sync::RwLock<ResultCache>>,
    pub kv: SharedKVStore,
    pub config: Config,
    pub logger: StructuredLogger,
    pub monitor: Arc<ResourceMonitor>,
    pub security_monitor: Arc<SecurityMonitor>,
    pub package_manager: Option<Arc<PackageManager>>,
    pub resource_enforcer: Option<Arc<ResourceEnforcer>>,
    pub enterprise_security: Option<Arc<EnterpriseSecurityEnforcer>>,
    pub orchestrator_client: Option<Arc<OrchestratorClient>>,
    pub yara_scanner: Option<Arc<YaraScanner>>,
    pub scheduler: Option<Arc<BinPackingScheduler>>,
    pub micropython_executor: Option<Arc<MicroPythonExecutor>>,
    pub shutdown_coordinator: Option<Arc<tokio::sync::RwLock<ShutdownCoordinator>>>,
    pub resource_manager: Option<Arc<ResourceManager>>,
    pub error_recovery: Option<Arc<ErrorRecovery>>,
    pub python_pool: Option<Arc<PythonRuntimePool>>,
    pub wasm_pool: Option<Arc<PoolManager>>,
    /// WASM cell executor for graph node isolation (Phase 2).
    /// Each Tool/Memory graph node runs as an isolated WASM cell with
    /// host function injection (kv, ai, fetch) and memory limits.
    pub wasm_cell_executor: Option<Arc<WasmCellExecutor>>,
    /// Multi-tier memory layer (Phase 3) — hot (dashmap LRU) + warm (Redis) + cold (PostgreSQL).
    /// Shared across all graph executions for Memory node operations.
    pub memory_layer: Option<Arc<MemoryLayer>>,
    /// FlyMind model router client (Phase 4) — routes LLM nodes to Python ai-service on port 8081.
    /// Reuses the existing 9-provider routing instead of rebuilding in Rust.
    pub flymind_client: Option<Arc<FlyMindClient>>,
    /// Integrated SAR node executor — wires MemoryLayer + FlyMindClient + WasmCellExecutor
    /// for real graph node execution (replaces DefaultNodeExecutor stubs).
    pub sar_executor: Option<Arc<crate::engine::sar_executor::SarNodeExecutor>>,
    /// Agent scheduler (Phase 5) — priority queues, per-tenant rate limiting, backpressure.
    /// Enqueues graph executions for async processing when the system is under load.
    pub agent_scheduler: Option<Arc<tokio::sync::RwLock<AgentScheduler>>>,
    /// Job tracker for scheduled async executions — tracks queued/processing/completed jobs.
    pub job_tracker: Option<Arc<JobTracker>>,
    /// Self-optimization engine (Phase 7) — analyzes execution history, emits suggestions to Go backend.
    /// Never auto-mutates production graphs; suggestions require manual approval.
    pub optimizer: Option<Arc<GraphOptimizer>>,
    /// Cost attributor (Phase 6) — tracks per-agent, per-execution-path cost.
    /// Emits `AgentExecutionRecord` events to the Go backend for billing.
    pub cost_attributor: Option<Arc<CostAttributor>>,
}

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

/// Execute request payload.
#[derive(Debug, Deserialize)]
pub struct ExecuteRequest {
    /// Input to the function.
    pub input: Option<String>,
    /// Tenant ID for KV namespace isolation (optional; falls back to config).
    pub tenant_id: Option<String>,
}

/// Daemon execution request (from SandboxClient).
#[derive(Debug, Deserialize)]
pub struct DaemonExecuteRequest {
    /// Base64-encoded WASM binary.
    pub wasm_binary: String,
    /// Base64-encoded AOT-compiled module bytes (optional, avoids JIT compilation).
    pub wasm_compiled: Option<String>,
    /// Input to the function.
    pub input: String,
    /// Timeout in milliseconds.
    pub timeout_ms: Option<u64>,
    /// Memory limit in MB.
    pub memory_mb: Option<u32>,
    /// Function name override.
    pub function: Option<String>,
    /// Function version override.
    pub version: Option<String>,
    /// Tenant ID for KV namespace isolation.
    pub tenant_id: Option<String>,
    /// Execution context (JSON).
    pub context: Option<serde_json::Value>,
}

/// Chunked execution request payload — streams WASM chunks for memory-efficient processing.
#[derive(Debug, Deserialize)]
pub struct ChunkedExecuteRequest {
    /// Total number of chunks being sent.
    pub total_chunks: u32,
    /// Chunk index (0-based).
    pub chunk_index: u32,
    /// Base64-encoded chunk data.
    pub chunk_data: String,
    /// Whether this is the last chunk.
    pub is_last: bool,
    /// Input to the function (sent with first chunk).
    pub input: Option<String>,
    /// Timeout in milliseconds.
    pub timeout_ms: Option<u64>,
    /// Memory limit in MB.
    pub memory_mb: Option<u32>,
    /// Function name override.
    pub function: Option<String>,
    /// Function version override.
    pub version: Option<String>,
    /// Tenant ID for KV namespace isolation.
    pub tenant_id: Option<String>,
}

/// Chunked execution response — sent after each chunk is processed.
#[derive(Debug, Serialize)]
pub struct ChunkedExecuteResponse {
    /// Chunk index this response corresponds to.
    pub chunk_index: u32,
    /// Whether this is the last chunk.
    pub is_last: bool,
    /// Partial output accumulated so far.
    pub partial_output: String,
    /// Execution time in milliseconds.
    pub exec_time_ms: u64,
    /// Whether execution is complete.
    pub done: bool,
}

/// Chunked execution completion response — sent when all chunks are processed.
#[derive(Debug, Serialize)]
pub struct ChunkedCompleteResponse {
    /// Final result string.
    pub result: String,
    /// Total execution time in milliseconds.
    pub total_exec_time_ms: u64,
    /// Number of chunks processed.
    pub chunks_processed: u32,
    /// Whether the result was served from cache.
    pub cache_hit: bool,
}

/// Health check response.
#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
}

/// Execute response payload.
#[derive(Debug, Serialize)]
pub struct ExecuteResponse {
    pub result: String,
    pub exec_time_ms: u64,
    pub cache_hit: bool,
    pub instance_id: String,
    pub function: String,
    pub version: String,
}

/// Standard error response shape.
#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub correlation_id: Option<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub recovery_suggestions: Vec<String>,
}

impl IntoResponse for ErrorResponse {
    fn into_response(self) -> axum::response::Response {
        (axum::http::StatusCode::INTERNAL_SERVER_ERROR, axum::Json(self)).into_response()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_execute_request_deserialization() {
        let json = r#"{"input": "test input", "tenant_id": "tenant-123"}"#;
        let request: ExecuteRequest = serde_json::from_str(json).unwrap();
        assert_eq!(request.input, Some("test input".to_string()));
        assert_eq!(request.tenant_id, Some("tenant-123".to_string()));
    }

    #[test]
    fn test_execute_request_minimal() {
        let json = r#"{}"#;
        let request: ExecuteRequest = serde_json::from_str(json).unwrap();
        assert_eq!(request.input, None);
        assert_eq!(request.tenant_id, None);
    }

    #[test]
    fn test_daemon_execute_request_deserialization() {
        let json = r#"{
            "wasm_binary": "aGVsbG8=",
            "input": "test input",
            "timeout_ms": 5000,
            "memory_mb": 256
        }"#;
        let request: DaemonExecuteRequest = serde_json::from_str(json).unwrap();
        assert_eq!(request.wasm_binary, "aGVsbG8=");
        assert_eq!(request.input, "test input");
        assert_eq!(request.timeout_ms, Some(5000));
        assert_eq!(request.memory_mb, Some(256));
    }

    #[test]
    fn test_health_response_serialization() {
        let response = HealthResponse {
            status: "healthy".to_string(),
            version: "1.0.0".to_string(),
        };
        let json = serde_json::to_string(&response).unwrap();
        assert!(json.contains("healthy"));
        assert!(json.contains("1.0.0"));
    }

    #[test]
    fn test_execute_response_serialization() {
        let response = ExecuteResponse {
            result: "success".to_string(),
            exec_time_ms: 100,
            cache_hit: true,
            instance_id: "inst-123".to_string(),
            function: "my-function".to_string(),
            version: "1.0.0".to_string(),
        };
        let json = serde_json::to_string(&response).unwrap();
        assert!(json.contains("success"));
        assert!(json.contains("100"));
        assert!(json.contains("inst-123"));
    }

    #[test]
    fn test_error_response_serialization() {
        let response = ErrorResponse {
            error: "Something went wrong".to_string(),
            correlation_id: Some("corr-456".to_string()),
            recovery_suggestions: vec!["Try again".to_string(), "Check logs".to_string()],
        };
        let json = serde_json::to_string(&response).unwrap();
        assert!(json.contains("Something went wrong"));
        assert!(json.contains("corr-456"));
        assert!(json.contains("Try again"));
    }

    #[test]
    fn test_error_response_optional_fields() {
        let response = ErrorResponse {
            error: "Simple error".to_string(),
            correlation_id: None,
            recovery_suggestions: vec![],
        };
        let json = serde_json::to_string(&response).unwrap();
        // correlation_id should be skipped when None
        assert!(!json.contains("correlation_id"));
        // recovery_suggestions should be skipped when empty
        assert!(!json.contains("recovery_suggestions"));
    }

    #[test]
    fn test_error_response_into_response() {
        let response = ErrorResponse {
            error: "test error".to_string(),
            correlation_id: None,
            recovery_suggestions: vec![],
        };
        let axum_response = response.into_response();
        assert_eq!(axum_response.status(), axum::http::StatusCode::OK);
    }
}
