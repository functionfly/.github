//! Error types for Prism Runtime

use thiserror::Error;

/// Result type alias for Prism operations
pub type PrismResult<T> = Result<T, PrismError>;

/// Main error type for Prism Runtime
#[derive(Error, Debug)]
pub enum PrismError {
    #[error("Cell not found: {0}")]
    CellNotFound(String),

    #[error("Cell already exists: {0}")]
    CellAlreadyExists(String),

    #[error("Cell in invalid state: {0} (current: {1})")]
    InvalidCellState(String, String),

    #[error("WASM execution failed: {0}")]
    WasmExecutionFailed(String),

    #[error("WASM module error: {0}")]
    WasmModuleError(String),

    #[error("Module not found: {0}")]
    ModuleNotFound(String),

    #[error("Fusion graph error: {0}")]
    FusionError(String),

    #[error("Scheduler error: {0}")]
    SchedulerError(String),

    #[error("Capability not found: {0}")]
    CapabilityNotFound(String),

    #[error("Capability error: {0}")]
    CapabilityError(String),

    #[error("State stream error: {0}")]
    StateStreamError(String),

    #[error("Snapshot error: {0}")]
    SnapshotError(String),

    #[error("Migration failed: {0}")]
    MigrationFailed(String),

    #[error("Mesh networking error: {0}")]
    MeshError(String),

    #[error("Swarm coordination error: {0}")]
    SwarmError(String),

    #[error("Neural optimization error: {0}")]
    NeuralError(String),

    #[error("Timeout: {0}ms exceeded")]
    Timeout(u32),

    #[error("Resource exhausted: {0}")]
    ResourceExhausted(String),

    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    #[error("Serialization error: {0}")]
    SerializationError(String),

    #[error("Permission denied: {0}")]
    PermissionDenied(String),

    #[error("Tenant isolation violation: {0}")]
    TenantViolation(String),

    #[error("Internal error: {0}")]
    Internal(String),
}

impl From<std::io::Error> for PrismError {
    fn from(err: std::io::Error) -> Self {
        PrismError::Internal(format!("IO error: {}", err))
    }
}

impl From<wasmtime::Error> for PrismError {
    fn from(err: wasmtime::Error) -> Self {
        PrismError::WasmExecutionFailed(err.to_string())
    }
}

impl From<serde_json::Error> for PrismError {
    fn from(err: serde_json::Error) -> Self {
        PrismError::SerializationError(err.to_string())
    }
}

impl From<prost::EncodeError> for PrismError {
    fn from(err: prost::EncodeError) -> Self {
        PrismError::SerializationError(err.to_string())
    }
}

impl From<prost::DecodeError> for PrismError {
    fn from(err: prost::DecodeError) -> Self {
        PrismError::SerializationError(err.to_string())
    }
}