//! Core cell types for Adaptive Execution Cells (AECs)
//!
//! An AEC is a self-describing, portable, AI-aware, hot-swappable,
//! state-streaming WASM cell that serves as the fundamental unit
//! of execution in Prism.

use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Unique identifier for an ExecutionCell
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct CellId(pub Uuid);

impl CellId {
    pub fn new() -> Self {
        Self(Uuid::new_v4())
    }

    pub fn from_uuid(uuid: Uuid) -> Self {
        Self(uuid)
    }

    pub fn as_uuid(&self) -> Uuid {
        self.0
    }

    pub fn nil() -> Self {
        Self(Uuid::nil())
    }
}

impl Default for CellId {
    fn default() -> Self {
        Self::new()
    }
}

impl std::fmt::Display for CellId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Status of an ExecutionCell
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[repr(u8)]
pub enum CellStatus {
    Pending = 0,
    Initializing = 1,
    Running = 2,
    Waiting = 3,
    Migrating = 4,
    Frozen = 5,
    Failed = 6,
    Terminated = 7,
}

impl Default for CellStatus {
    fn default() -> Self {
        CellStatus::Pending
    }
}

/// Configuration for an ExecutionCell
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CellConfig {
    /// Memory limit in megabytes
    pub memory_limit_mb: u64,
    /// Execution timeout in milliseconds
    pub timeout_ms: u32,
    /// Maximum number of concurrent instances
    pub max_instances: u32,
    /// Whether to enable strict isolation
    pub isolation_enabled: bool,
    /// Required capabilities (e.g., "ai:inference", "storage:read")
    pub capabilities: Vec<String>,
    /// Environment variables
    pub env_vars: HashMap<String, String>,
    /// Preferred execution target
    pub execution_target: ExecutionTarget,
    /// Scheduling hints
    pub placement_hint: Option<PlacementHint>,
}

impl Default for CellConfig {
    fn default() -> Self {
        Self {
            memory_limit_mb: 128,
            timeout_ms: 30_000,
            max_instances: 1,
            isolation_enabled: true,
            capabilities: Vec::new(),
            env_vars: HashMap::new(),
            execution_target: ExecutionTarget::Cloud,
            placement_hint: None,
        }
    }
}

/// Execution target location
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ExecutionTarget {
    Cloud,
    Edge,
    Browser,
    Robotic,
    Mobile,
    IoT,
}

impl Default for ExecutionTarget {
    fn default() -> Self {
        ExecutionTarget::Cloud
    }
}

/// Resources allocated to a cell
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CellResources {
    pub vcpus: u32,
    pub memory_bytes: u64,
    pub gpu_required: bool,
    pub gpu_memory_mb: u32,
    pub location: ExecutionLocation,
    pub cost_weight: f32,
}

impl Default for CellResources {
    fn default() -> Self {
        Self {
            vcpus: 1,
            memory_bytes: 128 * 1024 * 1024, // 128 MB
            gpu_required: false,
            gpu_memory_mb: 0,
            location: ExecutionLocation::Cloud,
            cost_weight: 1.0,
        }
    }
}

/// Execution location
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ExecutionLocation {
    Cloud,
    Edge,
    Browser,
    Robotic,
    Mobile,
    IoT,
}

impl Default for ExecutionLocation {
    fn default() -> Self {
        ExecutionLocation::Cloud
    }
}

/// Metadata about a cell
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CellMetadata {
    pub name: String,
    pub version: String,
    pub runtime: String,
    pub tags: HashMap<String, String>,
    pub languages: Vec<String>,
    pub description: String,
    pub created_at: DateTime<Utc>,
    pub last_executed_at: Option<DateTime<Utc>>,
    pub expires_at: Option<DateTime<Utc>>,
}

impl CellMetadata {
    pub fn new(name: &str, runtime: &str) -> Self {
        Self {
            name: name.to_string(),
            version: "1.0.0".to_string(),
            runtime: runtime.to_string(),
            tags: HashMap::new(),
            languages: Vec::new(),
            description: String::new(),
            created_at: Utc::now(),
            last_executed_at: None,
            expires_at: None,
        }
    }
}

/// Placement hint for the scheduler
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlacementHint {
    pub preferred_location: ExecutionLocation,
    pub preferred_regions: Vec<String>,
    pub latency_sensitivity: f32,  // 0.0 to 1.0
    pub cost_sensitivity: f32,     // 0.0 to 1.0
    pub gpu_required: bool,
    pub model_affinity: Option<String>,  // For AI model routing
}

impl Default for PlacementHint {
    fn default() -> Self {
        Self {
            preferred_location: ExecutionLocation::Cloud,
            preferred_regions: Vec::new(),
            latency_sensitivity: 0.5,
            cost_sensitivity: 0.5,
            gpu_required: false,
            model_affinity: None,
        }
    }
}

/// A complete Adaptive Execution Cell
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionCell {
    pub id: CellId,
    pub tenant_id: String,
    pub status: CellStatus,
    pub config: CellConfig,
    pub resources: CellResources,
    pub metadata: CellMetadata,
    pub wasm_module_id: Option<String>,
    pub serialized_state: Option<Vec<u8>>,
    pub checkpoint_epoch: u64,
}

impl ExecutionCell {
    pub fn new(tenant_id: &str, config: CellConfig, metadata: CellMetadata) -> Self {
        Self {
            id: CellId::new(),
            tenant_id: tenant_id.to_string(),
            status: CellStatus::Pending,
            config,
            resources: CellResources::default(),
            metadata,
            wasm_module_id: None,
            serialized_state: None,
            checkpoint_epoch: 0,
        }
    }

    pub fn set_status(&mut self, status: CellStatus) {
        self.status = status;
    }

    pub fn is_running(&self) -> bool {
        matches!(self.status, CellStatus::Running)
    }

    pub fn can_migrate(&self) -> bool {
        matches!(
            self.status,
            CellStatus::Running | CellStatus::Waiting | CellStatus::Frozen
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cell_id_generation() {
        let id = CellId::new();
        assert!(!id.as_uuid().is_nil());
    }

    #[test]
    fn test_cell_creation() {
        let config = CellConfig::default();
        let metadata = CellMetadata::new("test-cell", "wasm");
        let cell = ExecutionCell::new("tenant-1", config, metadata);

        assert_eq!(cell.tenant_id, "tenant-1");
        assert_eq!(cell.status, CellStatus::Pending);
        assert_eq!(cell.metadata.name, "test-cell");
    }

    #[test]
    fn test_cell_can_migrate() {
        let config = CellConfig::default();
        let metadata = CellMetadata::new("test", "wasm");
        let mut cell = ExecutionCell::new("tenant", config, metadata);

        cell.status = CellStatus::Running;
        assert!(cell.can_migrate());

        cell.status = CellStatus::Pending;
        assert!(!cell.can_migrate());
    }
}