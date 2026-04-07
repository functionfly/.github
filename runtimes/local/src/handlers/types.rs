//! Shared types for all HTTP handlers.

use axum::response::IntoResponse;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

pub use crate::cache::ResultCache;
pub use crate::config::Config;
pub use crate::engine::WasmEngine;
pub use crate::enterprise_security::EnterpriseSecurityEnforcer;
pub use crate::errors::ErrorRecovery;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::micropython::MicroPythonExecutor;
use crate::monitoring::ResourceMonitor;
use crate::orchestrator_client::OrchestratorClient;
use crate::package::PackageManager;
pub use crate::pool::{InstancePool, PoolManager};
use crate::python_pool::PythonRuntimePool;
use crate::resource_enforcer::ResourceEnforcer;
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
        axum::Json(self).into_response()
    }
}
