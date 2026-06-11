//! Runtime context for Prism Runtime
//!
//! Manages the global state of the Prism runtime including:
//! - Execution cells
//! - Capability registry
//! - Swarm coordination
//! - Health monitoring

use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::info;

use crate::core::{
    CellId, CellStatus, CellConfig, CellMetadata, ExecutionCell,
    PrismError, PrismResult,
};
use crate::hypercore::{Scheduler, SchedulerConfig};
use crate::nats_client::NatsOrchestratorClient;
use crate::neural::{NeuralOptimizer, FeedbackLoop, ExecutionProfile, ExecutionOutcome};
use crate::neural::feedback::FeedbackEntry;
use crate::neural::optimizer::OptimizationSuggestion;
use crate::quantum::{SnapshotManager, Snapshot, SnapshotType};
use crate::security::{SecurityManager, SecurityPolicy};
use crate::ucl::{CapabilityRegistry, Capability};
use crate::swarm::{SwarmCoordinator, SwarmId};
use crate::wasm_fusion::FusionExecutor;
#[cfg(feature = "mesh")]
use crate::mesh::{MeshNetwork, MeshConfig};

/// Runtime context holding all Prism runtime state
pub struct RuntimeContext {
    /// Cell management
    pub cells: Arc<RwLock<HashMap<CellId, ExecutionCell>>>,
    /// WASM module storage (module_id -> wasm_bytes)
    pub modules: Arc<RwLock<HashMap<String, Vec<u8>>>>,
    /// Fusion executor for WASM execution
    pub fusion_executor: Arc<RwLock<Option<FusionExecutor>>>,
    /// Capability registry
    pub capability_registry: Arc<RwLock<CapabilityRegistry>>,
    /// Swarm coordinator
    pub swarm_coordinator: Arc<RwLock<Option<SwarmCoordinator>>>,
    /// NATS client for orchestrator communication
    pub nats_client: Arc<RwLock<Option<NatsOrchestratorClient>>>,
    /// Scheduler for cell placement
    pub scheduler: Arc<RwLock<Scheduler>>,
    /// Snapshot manager for cell migration
    pub snapshot_manager: Arc<RwLock<SnapshotManager>>,
    /// Neural optimizer for self-tuning execution
    pub neural_optimizer: Arc<RwLock<NeuralOptimizer>>,
    /// Feedback loop for RL-based optimization
    pub feedback_loop: Arc<RwLock<FeedbackLoop>>,
    /// Per-cell cache of the last captured WASM CPU state (CBOR-encoded WasmCpuState)
    /// Used to provide real VM state when creating Full snapshots.
    pub last_cpu_states: Arc<RwLock<HashMap<CellId, Vec<u8>>>>,
    /// Security manager for WASM validation and sandbox policy enforcement.
    /// Every WASM module is validated through this manager before being
    /// registered as a cell.
    pub security_manager: Arc<SecurityManager>,
    /// P2P mesh network (only populated when the `mesh` feature is enabled
    /// AND the runtime was constructed with `mesh_enabled = true`).
    #[cfg(feature = "mesh")]
    pub mesh_network: Arc<RwLock<Option<MeshNetwork>>>,
    /// Whether mesh networking is enabled
    pub mesh_enabled: bool,
    /// Runtime listen address
    pub listen_address: String,
}

impl RuntimeContext {
    /// Create a new runtime context with a default (production) security policy.
    pub fn new(listen_address: String, mesh_enabled: bool) -> Self {
        Self::with_policy(listen_address, mesh_enabled, SecurityPolicy::production())
    }

    /// Create a new runtime context with a caller-supplied security policy.
    pub fn with_policy(listen_address: String, mesh_enabled: bool, policy: SecurityPolicy) -> Self {
        Self {
            cells: Arc::new(RwLock::new(HashMap::new())),
            modules: Arc::new(RwLock::new(HashMap::new())),
            fusion_executor: Arc::new(RwLock::new(None)),
            capability_registry: Arc::new(RwLock::new(CapabilityRegistry::new(
                crate::ucl::RegistryConfig::default()
            ))),
            swarm_coordinator: Arc::new(RwLock::new(None)),
            nats_client: Arc::new(RwLock::new(None)),
            scheduler: Arc::new(RwLock::new(Scheduler::new(
                SchedulerConfig::default()
            ))),
            snapshot_manager: Arc::new(RwLock::new(SnapshotManager::new(100))),
            neural_optimizer: Arc::new(RwLock::new(NeuralOptimizer::new(1000))),
            feedback_loop: Arc::new(RwLock::new(FeedbackLoop::default())),
            last_cpu_states: Arc::new(RwLock::new(HashMap::new())),
            security_manager: SecurityManager::new(policy),
            #[cfg(feature = "mesh")]
            mesh_network: Arc::new(RwLock::new(None)),
            mesh_enabled,
            listen_address,
        }
    }

    /// Initialize the fusion executor
    pub async fn init_fusion_executor(&self) -> PrismResult<()> {
        let executor = FusionExecutor::new()?;
        let mut fusion = self.fusion_executor.write().await;
        *fusion = Some(executor);
        info!("Fusion executor initialized");
        Ok(())
    }

    /// Create a new cell
    pub async fn create_cell(
        &self,
        tenant_id: &str,
        name: &str,
        wasm_bytes: Vec<u8>,
        config: CellConfig,
    ) -> PrismResult<CellId> {
        // Validate the WASM module BEFORE registering it. Modules that fail
        // validation are rejected outright so the runtime never has to
        // defend against malicious imports.
        let validation = self.security_manager.validate_wasm(&wasm_bytes);
        if let Ok(result) = &validation {
            if !result.valid {
                let descs: Vec<String> = result.violations.iter()
                    .map(|v| format!("[{:?}] {}: {}", v.severity, v.pattern, v.description))
                    .collect();
                return Err(PrismError::WasmModuleError(format!(
                    "WASM module rejected by security policy: {}",
                    descs.join("; ")
                )));
            }
        } else if let Err(e) = validation {
            return Err(PrismError::WasmModuleError(format!("WASM validation error: {}", e)));
        }

        // Register the WASM module
        let module_id = format!("module-{}", uuid::Uuid::new_v4());
        {
            let mut modules = self.modules.write().await;
            modules.insert(module_id.clone(), wasm_bytes.clone());
        }

        // Initialize fusion executor if not already done
        {
            let mut fusion = self.fusion_executor.write().await;
            if fusion.is_none() {
                *fusion = Some(FusionExecutor::new()?);
            }
            if let Some(ref executor) = *fusion {
                executor.register_module(&module_id, &wasm_bytes).await?;
            }
        }

        // Create the cell
        let metadata = CellMetadata::new(name, "wasm");
        let mut cell = ExecutionCell::new(tenant_id, config, metadata);
        cell.wasm_module_id = Some(module_id.clone());
        cell.status = CellStatus::Initializing;

        let cell_id = cell.id;
        {
            let mut cells = self.cells.write().await;
            cells.insert(cell_id, cell);
        }

        // Notify NATS orchestrator that a new cell has been registered. This is fire-and-forget
        // so a NATS outage does not block cell creation.
        if let Some(client) = self.nats_client.read().await.as_ref() {
            let client = client.clone();
            let cell_id_for_log = cell_id;
            let name_owned = name.to_string();
            let capabilities: Vec<String> = Vec::new();
            tokio::spawn(async move {
                if let Err(e) = client.notify_cell_registered(cell_id_for_log, &name_owned, capabilities).await {
                    tracing::warn!(cell_id = %cell_id_for_log, "Failed to publish cell-registered notification: {}", e);
                }
            });
        }

        info!(cell_id = %cell_id, name = %name, "Cell created successfully");
        Ok(cell_id)
    }

    /// Get a cell by ID
    pub async fn get_cell(&self, cell_id: &CellId) -> Option<ExecutionCell> {
        let cells = self.cells.read().await;
        cells.get(cell_id).cloned()
    }

    /// List all cells
    pub async fn list_cells(&self) -> Vec<ExecutionCell> {
        let cells = self.cells.read().await;
        cells.values().cloned().collect()
    }

    /// Terminate a cell
    pub async fn terminate_cell(&self, cell_id: &CellId) -> PrismResult<()> {
        let mut cells = self.cells.write().await;
        if let Some(cell) = cells.get_mut(cell_id) {
            cell.set_status(CellStatus::Terminated);
            info!(cell_id = %cell_id, "Cell terminated");
            Ok(())
        } else {
            Err(PrismError::CellNotFound(cell_id.to_string()))
        }
    }

    /// Register a capability
    pub async fn register_capability(&self, capability: Capability) -> PrismResult<()> {
        let registry = self.capability_registry.read().await;
        registry.register_local(capability).await?;
        Ok(())
    }

    /// Discover capabilities matching a query
    pub async fn discover_capabilities(&self, query: &str) -> Vec<Capability> {
        let registry = self.capability_registry.read().await;
        registry.search(query).await
    }

    /// List all capabilities
    pub async fn list_capabilities(&self) -> Vec<Capability> {
        let registry = self.capability_registry.read().await;
        registry.list_all().await
    }

    /// Get runtime status
    pub async fn get_status(&self) -> RuntimeStatus {
        let cells = self.cells.read().await;
        let active_cells = cells.values()
            .filter(|c| matches!(c.status, CellStatus::Running | CellStatus::Initializing))
            .count();

        RuntimeStatus {
            version: crate::VERSION.to_string(),
            healthy: true,
            active_cells: active_cells as u32,
            total_cells: cells.len() as u32,
            mesh_enabled: self.mesh_enabled,
            listen_address: self.listen_address.clone(),
        }
    }

    /// Create a snapshot of a cell
    ///
    /// If a cached WASM CPU state exists for this cell (from a prior execution),
    /// it is embedded in Full snapshots instead of placeholder metadata.
    pub async fn snapshot_cell(&self, cell_id: &CellId, snapshot_type: SnapshotType) -> PrismResult<Snapshot> {
        // Verify cell exists and get module bytes
        let (module_bytes, env_vars) = {
            let cells = self.cells.read().await;
            let cell = cells.get(cell_id).ok_or_else(|| PrismError::CellNotFound(cell_id.to_string()))?;

            // Get module bytes from modules store
            let module_bytes = if let Some(ref module_id) = cell.wasm_module_id {
                self.modules.read().await
                    .get(module_id)
                    .cloned()
                    .unwrap_or_default()
            } else {
                Vec::new()
            };

            // Clone what we need
            (module_bytes, cell.config.env_vars.clone())
        };

        // Retrieve cached CPU state if available
        let cached_cpu_state = {
            let cpu_states = self.last_cpu_states.read().await;
            cpu_states.get(cell_id).cloned()
        };

        // Create snapshot via snapshot manager with actual module data and CPU state
        let mut manager = self.snapshot_manager.write().await;
        let snapshot = manager.create_snapshot_with_cpu(
            cell_id, &module_bytes, &env_vars, snapshot_type, cached_cpu_state,
        ).await?;

        info!(cell_id = %cell_id, snapshot_id = %snapshot.metadata.snapshot_id, "Cell snapshot created");
        Ok(snapshot)
    }

    /// Restore a cell from a snapshot
    pub async fn restore_cell_from_snapshot(&self, snapshot_id: &str) -> PrismResult<CellId> {
        let manager = self.snapshot_manager.read().await;
        let snapshot = manager.restore_snapshot(snapshot_id).await?;
        drop(manager);

        // Get the cell from snapshot metadata
        let cell_id = snapshot.metadata.cell_id;

        // Update cell status to indicate restored
        let mut cells = self.cells.write().await;
        if let Some(cell) = cells.get_mut(&cell_id) {
            cell.set_status(CellStatus::Running);
            info!(cell_id = %cell_id, snapshot_id = %snapshot_id, "Cell restored from snapshot");
        }

        Ok(cell_id)
    }

    /// List snapshots for a cell
    pub async fn list_cell_snapshots(&self, cell_id: &CellId) -> Vec<crate::quantum::SnapshotMetadata> {
        let manager = self.snapshot_manager.read().await;
        manager.list_for_cell(cell_id).into_iter().cloned().collect()
    }

    /// Delete a snapshot
    pub async fn delete_snapshot(&self, snapshot_id: &str) -> PrismResult<()> {
        let mut manager = self.snapshot_manager.write().await;
        if manager.delete_snapshot(snapshot_id) {
            info!(snapshot_id = %snapshot_id, "Snapshot deleted");
            Ok(())
        } else {
            Err(PrismError::SnapshotError(format!("Snapshot not found: {}", snapshot_id)))
        }
    }

    /// Record an execution outcome for neural optimization
    pub async fn record_execution_outcome(&self, profile: ExecutionProfile) {
        let mut optimizer = self.neural_optimizer.write().await;
        optimizer.record(profile.clone());

        // Calculate and record feedback
        let suggestion = optimizer.suggest(&profile.cell_id);
        let improvement = if profile.outcome == ExecutionOutcome::Success {
            1.0
        } else {
            -1.0
        };

        // Update Q-values so the optimizer learns from this execution
        let state = profile.cell_id.to_string();
        optimizer.update_q(&state, "memory", improvement, &state);
        optimizer.update_q(&state, "timeout", improvement, &state);
        optimizer.update_q(&state, "location", improvement, &state);

        drop(optimizer);

        let mut feedback = self.feedback_loop.write().await;
        feedback.record(FeedbackEntry {
            cell_id: profile.cell_id,
            suggestion,
            actual_outcome: profile.outcome,
            improvement_score: improvement,
            timestamp: chrono::Utc::now().timestamp(),
        });
    }

    /// Get optimization suggestion for a cell
    pub async fn get_optimization_suggestion(&self, cell_id: &CellId) -> OptimizationSuggestion {
        let optimizer = self.neural_optimizer.read().await;
        optimizer.suggest(cell_id)
    }

    /// Invoke a capability by name
    pub async fn invoke_capability(&self, name: &str, _input: &[u8]) -> PrismResult<Vec<u8>> {
        let registry = self.capability_registry.read().await;
        let caps = registry.search(name).await;

        if let Some(cap) = caps.into_iter().next() {
            // For now, return an error that capability invocation needs mesh
            // In production, this would route to the capability provider
            Err(PrismError::Internal(format!(
                "Capability '{}' found but invocation requires mesh network", cap.name
            )))
        } else {
            Err(PrismError::Internal(format!("Capability '{}' not found", name)))
        }
    }

    /// Run the cell's WASM module with the given input. This is the
    /// inline version of `invoke_capability` that the gRPC service uses
    /// for cell-level execution. Returns the cell's stdout bytes.
    pub async fn invoke_capability_inline(&self, cell_id: &CellId, input: &[u8]) -> PrismResult<Vec<u8>> {
        // Look up the cell to get the module id.
        let (module_id, memory_limit_mb) = {
            let cells = self.cells.read().await;
            let cell = cells
                .get(cell_id)
                .ok_or_else(|| PrismError::CellNotFound(cell_id.to_string()))?;
            (
                cell.wasm_module_id
                    .clone()
                    .ok_or_else(|| PrismError::WasmModuleError("Cell has no WASM module".to_string()))?,
                cell.config.memory_limit_mb,
            )
        };

        // Fetch the module bytes from the module cache.
        let module_bytes = {
            let modules = self.modules.read().await;
            modules
                .get(&module_id)
                .cloned()
                .ok_or_else(|| PrismError::WasmModuleError(format!("Module {} not found", module_id)))?
        };

        // Execute via the fusion executor in a blocking task so the async runtime
        // is not held inside wasmtime's compile/instantiate path.
        let fusion_executor_arc = self.fusion_executor.clone();
        let input_vec = input.to_vec();
        // Clone values that need to outlive the 'static closure.
        let cell_id_owned = *cell_id;
        let module_id_owned = module_id.clone();
        let handle = tokio::task::spawn_blocking(move || {
            let executor_guard = futures::executor::block_on(fusion_executor_arc.read());
            let Some(executor) = executor_guard.as_ref() else {
                return Err(PrismError::Internal("Fusion executor not initialized".to_string()));
            };

            // Build a one-node graph that calls a `_start` (or `handler`) entry.
            use crate::wasm_fusion::{FusionGraph, FusionNode, FusionNodeType, NodeConfig};
            let cell_id_str = cell_id_owned.to_string();
            let mut graph = FusionGraph::new(&cell_id_str);
            graph.add_node(FusionNode {
                node_id: module_id_owned.clone(),
                name: module_id_owned.clone(),
                node_type: FusionNodeType::Wasm,
                config: NodeConfig {
                    entry_point: "_start".to_string(),
                    timeout_ms: 30_000,
                    memory_limit_mb,
                    imports: Vec::new(),
                },
            });

            // Register the module with the executor if needed.
            let module_id_for_register = module_id_owned.clone();
            let module_bytes_clone = module_bytes.clone();
            let _ = futures::executor::block_on(executor.register_module(&module_id_for_register, &module_bytes_clone));

            futures::executor::block_on(executor.execute_graph(&mut graph, &input_vec))
        });

        match handle.await {
            Ok(Ok(output)) => Ok(output),
            Ok(Err(e)) => Err(e),
            Err(e) => Err(PrismError::Internal(format!("Execute task failed: {}", e))),
        }
    }

    /// Migrate a cell to another node. The cell is snapshotted locally and
    /// the snapshot is shipped to the target node. This is a thin wrapper
    /// around `MigrationManager` that resolves to a real network transfer
    /// when the mesh is enabled and a stub return otherwise.
    pub async fn migrate_cell(
        &self,
        cell_id: &CellId,
        target_node: &str,
    ) -> PrismResult<crate::quantum::MigrationResult> {
        // First, snapshot the cell.
        let snap = self.snapshot_cell(cell_id, crate::quantum::SnapshotType::Full).await?;
        let mut mgr = crate::quantum::MigrationManager::new();
        let strategy = crate::quantum::MigrationStrategy::Live;
        mgr.migrate_cell(*cell_id, "local-node", target_node, strategy, &snap)
            .await
    }

    /// Create a swarm
    pub async fn create_swarm(&self, swarm_id: String) -> PrismResult<String> {
        let mut coordinator = self.swarm_coordinator.write().await;
        if coordinator.is_none() {
            *coordinator = Some(SwarmCoordinator::new(crate::swarm::CoordinatorConfig::default()));
        }
        if let Some(ref mut coord) = *coordinator {
            coord.create_swarm(SwarmId(swarm_id.clone()))?;
            Ok(swarm_id)
        } else {
            Err(PrismError::Internal("Swarm coordinator not initialized".to_string()))
        }
    }

    /// Join a swarm
    pub async fn join_swarm(&self, swarm_id: &str, cell_id: CellId) -> PrismResult<()> {
        let mut coordinator = self.swarm_coordinator.write().await;
        if let Some(ref mut coord) = *coordinator {
            coord.add_cell_to_swarm(cell_id, &SwarmId(swarm_id.to_string()))
        } else {
            Err(PrismError::Internal("Swarm coordinator not initialized".to_string()))
        }
    }

    /// Leave a swarm
    pub async fn leave_swarm(&self, _swarm_id: &str, cell_id: CellId) -> PrismResult<()> {
        let mut coordinator = self.swarm_coordinator.write().await;
        if let Some(ref mut coord) = *coordinator {
            if coord.remove_cell(cell_id) {
                Ok(())
            } else {
                Err(PrismError::SwarmError("Cell not in any swarm".to_string()))
            }
        } else {
            Err(PrismError::Internal("Swarm coordinator not initialized".to_string()))
        }
    }

    /// List swarms
    pub async fn list_swarms(&self) -> Vec<crate::swarm::Swarm> {
        let coordinator = self.swarm_coordinator.read().await;
        if let Some(ref coord) = *coordinator {
            coord.swarms().values().cloned().collect()
        } else {
            Vec::new()
        }
    }
}

/// Runtime status information
#[derive(Debug, Clone)]
pub struct RuntimeStatus {
    pub version: String,
    pub healthy: bool,
    pub active_cells: u32,
    pub total_cells: u32,
    pub mesh_enabled: bool,
    pub listen_address: String,
}

impl RuntimeContext {
    /// Connect to NATS orchestrator with retry logic
    pub async fn connect_nats(&self, url: &str) -> PrismResult<()> {
        let mut client = NatsOrchestratorClient::new();
        client.connect(url).await.map_err(|e| {
            PrismError::Internal(format!("NATS connection failed: {}", e))
        })?;
        let mut nats = self.nats_client.write().await;
        *nats = Some(client);
        info!("Connected to NATS at {}", url);
        Ok(())
    }

    /// Initialize the P2P mesh network. The mesh is required for cross-node
    /// cell migration, distributed snapshots, and capability advertisement.
    /// Returns the local peer id once the swarm is listening.
    #[cfg(feature = "mesh")]
    pub async fn start_mesh(&self, config: MeshConfig) -> PrismResult<String> {
        use tracing::info;
        let mut network = MeshNetwork::new(config).map_err(|e| {
            PrismError::Internal(format!("Failed to construct mesh network: {}", e))
        })?;
        let peer_id = network.local_peer_id();
        let peer_id_str = peer_id.to_string();
        network.start().await.map_err(|e| {
            PrismError::Internal(format!("Failed to start mesh network: {}", e))
        })?;
        let mut slot = self.mesh_network.write().await;
        *slot = Some(network);
        info!(peer_id = %peer_id_str, "Mesh network started");
        Ok(peer_id_str)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // Minimal valid WASM module that exports _start
    // Generated from: (module (func (export "_start")))
    fn minimal_wasm() -> Vec<u8> {
        vec![
            0x00, 0x61, 0x73, 0x6d, // magic "\0asm"
            0x01, 0x00, 0x00, 0x00, // version 1
            0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // type section: 1 type, func () -> ()
            0x03, 0x02, 0x01, 0x00, // function section: 1 function (type 0)
            0x07, 0x0a, 0x01, 0x06, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x00, // export section: _start
            0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b, // code section: 1 body, end
        ]
    }

    #[tokio::test]
    async fn test_runtime_context_creation() {
        let ctx = RuntimeContext::new("0.0.0.0:8080".to_string(), false);
        assert_eq!(ctx.listen_address, "0.0.0.0:8080");
        assert!(!ctx.mesh_enabled);
    }

    #[tokio::test]
    async fn test_cell_creation() {
        let ctx = RuntimeContext::new("0.0.0.0:8080".to_string(), false);

        let config = CellConfig::default();
        let wasm = minimal_wasm();

        let result = ctx.create_cell("tenant-1", "test-cell", wasm, config).await;
        assert!(result.is_ok());

        let cell_id = result.unwrap();
        let cell = ctx.get_cell(&cell_id).await;
        assert!(cell.is_some());
        assert_eq!(cell.unwrap().tenant_id, "tenant-1");
    }

    #[tokio::test]
    async fn test_cell_termination() {
        let ctx = RuntimeContext::new("0.0.0.0:8080".to_string(), false);

        let config = CellConfig::default();
        let wasm = minimal_wasm();

        let cell_id = ctx.create_cell("tenant-1", "test-cell", wasm, config).await.unwrap();
        let result = ctx.terminate_cell(&cell_id).await;
        assert!(result.is_ok());

        let cell = ctx.get_cell(&cell_id).await.unwrap();
        assert_eq!(cell.status, CellStatus::Terminated);
    }

    #[tokio::test]
    async fn test_list_cells() {
        let ctx = RuntimeContext::new("0.0.0.0:8080".to_string(), false);

        let config = CellConfig::default();
        let wasm = minimal_wasm();

        ctx.create_cell("tenant-1", "cell-1", wasm.clone(), config.clone()).await.unwrap();
        ctx.create_cell("tenant-1", "cell-2", wasm.clone(), config.clone()).await.unwrap();

        let cells = ctx.list_cells().await;
        assert_eq!(cells.len(), 2);
    }
}