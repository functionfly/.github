//! Integration tests for Prism Runtime core systems
//!
//! These tests verify the interaction between multiple runtime components.

#[cfg(test)]
mod tests {
    use crate::core::{CellId, CellStatus, CellConfig, CellMetadata, ExecutionCell, ExecutionTarget};

    // === Cell Lifecycle Integration Tests ===

    #[tokio::test]
    async fn test_cell_status_transitions() {
        let config = CellConfig::default();
        let metadata = CellMetadata::new("test-cell", "wasm");
        let mut cell = ExecutionCell::new("tenant-1", config, metadata);

        // Initial state
        assert_eq!(cell.status, CellStatus::Pending);

        // Transition to initializing
        cell.set_status(CellStatus::Initializing);
        assert_eq!(cell.status, CellStatus::Initializing);

        // Transition to running
        cell.set_status(CellStatus::Running);
        assert_eq!(cell.status, CellStatus::Running);
        assert!(cell.is_running());

        // Transition to waiting
        cell.set_status(CellStatus::Waiting);
        assert!(cell.can_migrate());

        // Transition to frozen
        cell.set_status(CellStatus::Frozen);
        assert!(cell.can_migrate());

        // Transition to migrating (while already migrating, can't start another)
        cell.set_status(CellStatus::Migrating);
        assert!(!cell.can_migrate(), "Cell that is already migrating cannot start another migration");

        // Transition to terminated
        cell.set_status(CellStatus::Terminated);
        assert!(!cell.can_migrate());
    }

    #[tokio::test]
    async fn test_cell_migration_eligibility() {
        let config = CellConfig::default();
        let metadata = CellMetadata::new("migrate-test", "wasm");
        let cell = ExecutionCell::new("tenant-1", config, metadata);

        // Cells that can migrate
        let migratable = [
            CellStatus::Running,
            CellStatus::Waiting,
            CellStatus::Frozen,
        ];

        for status in migratable {
            let mut c = cell.clone();
            c.set_status(status);
            assert!(c.can_migrate(), "Status {:?} should be migratable", status);
        }

        // Cells that cannot migrate
        let non_migratable = [
            CellStatus::Pending,
            CellStatus::Initializing,
            CellStatus::Migrating,
            CellStatus::Failed,
            CellStatus::Terminated,
        ];

        for status in non_migratable {
            let mut c = cell.clone();
            c.set_status(status);
            assert!(!c.can_migrate(), "Status {:?} should not be migratable", status);
        }
    }

    #[tokio::test]
    async fn test_cell_execution_target_variants() {
        let config = CellConfig {
            execution_target: ExecutionTarget::Edge,
            ..Default::default()
        };
        let metadata = CellMetadata::new("edge-cell", "wasm");
        let cell = ExecutionCell::new("tenant-1", config, metadata);

        assert!(matches!(cell.config.execution_target, ExecutionTarget::Edge));
    }

    #[tokio::test]
    async fn test_cell_id_uniqueness() {
        let id1 = CellId::new();
        let id2 = CellId::new();
        assert_ne!(id1, id2);
    }

    #[tokio::test]
    async fn test_cell_clone_preserves_id() {
        let config = CellConfig::default();
        let metadata = CellMetadata::new("clone-test", "wasm");
        let cell = ExecutionCell::new("tenant-1", config, metadata);
        let cell_clone = cell.clone();

        assert_eq!(cell.id, cell_clone.id);
        assert_eq!(cell.tenant_id, cell_clone.tenant_id);
    }

    // === CellConfig Tests ===

    #[tokio::test]
    async fn test_cell_config_defaults() {
        let config = CellConfig::default();

        assert_eq!(config.memory_limit_mb, 128);
        assert_eq!(config.timeout_ms, 30_000);
        assert_eq!(config.max_instances, 1);
        assert!(config.isolation_enabled);
        assert!(config.capabilities.is_empty());
        assert!(config.env_vars.is_empty());
    }

    #[tokio::test]
    async fn test_cell_config_custom_values() {
        let config = CellConfig {
            memory_limit_mb: 512,
            timeout_ms: 60_000,
            max_instances: 4,
            isolation_enabled: false,
            capabilities: vec!["gpu".to_string(), "inference".to_string()],
            env_vars: vec![("KEY".to_string(), "value".to_string())].into_iter().collect(),
            execution_target: ExecutionTarget::Cloud,
            placement_hint: None,
        };

        assert_eq!(config.memory_limit_mb, 512);
        assert_eq!(config.timeout_ms, 60_000);
        assert_eq!(config.max_instances, 4);
        assert!(!config.isolation_enabled);
        assert_eq!(config.capabilities.len(), 2);
    }

    // === RuntimeContext Tests ===

    #[tokio::test]
    async fn test_runtime_initialization() {
        use crate::runtime::RuntimeContext;

        let runtime = RuntimeContext::new("127.0.0.1:0".to_string(), false);

        assert!(!runtime.mesh_enabled);
        assert_eq!(runtime.listen_address, "127.0.0.1:0");
    }

    #[tokio::test]
    async fn test_runtime_with_mesh_enabled() {
        use crate::runtime::RuntimeContext;

        let runtime = RuntimeContext::new("0.0.0.0:8080".to_string(), true);

        assert!(runtime.mesh_enabled);
    }

    #[tokio::test]
    async fn test_runtime_status() {
        use crate::runtime::RuntimeContext;

        let runtime = RuntimeContext::new("127.0.0.1:0".to_string(), false);

        let status = runtime.get_status().await;
        assert_eq!(status.version, crate::VERSION);
        assert!(status.healthy);
        assert_eq!(status.total_cells, 0);
        assert_eq!(status.active_cells, 0);
        assert!(!status.mesh_enabled);
    }

    // === Swarm Coordinator Tests ===

    #[tokio::test]
    async fn test_swarm_coordinator_creation() {
        use crate::swarm::{SwarmCoordinator, CoordinatorConfig};

        let coordinator = SwarmCoordinator::new(CoordinatorConfig::default());

        // Should start with no swarms
        let swarms = coordinator.swarms();
        assert!(swarms.is_empty());
    }

    #[tokio::test]
    async fn test_swarm_cell_management() {
        use crate::swarm::{SwarmCoordinator, SwarmId, CoordinatorConfig};

        let mut coordinator = SwarmCoordinator::new(CoordinatorConfig::default());
        let swarm_id = SwarmId::new("test-swarm-2".to_string());

        coordinator.create_swarm(swarm_id.clone()).unwrap();

        // Create a cell
        let cell_id = CellId::new();

        // Add cell to swarm
        let result = coordinator.add_cell_to_swarm(cell_id, &swarm_id);
        assert!(result.is_ok());

        // Verify cell is in swarm
        let swarm = coordinator.get_swarm(&swarm_id).unwrap();
        assert!(swarm.cells.contains(&cell_id));
    }

    #[tokio::test]
    async fn test_swarm_removal() {
        use crate::swarm::{SwarmCoordinator, SwarmId, CoordinatorConfig};

        let mut coordinator = SwarmCoordinator::new(CoordinatorConfig::default());
        let swarm_id = SwarmId::new("test-swarm-3".to_string());

        coordinator.create_swarm(swarm_id.clone()).unwrap();

        let cell_id = CellId::new();
        coordinator.add_cell_to_swarm(cell_id, &swarm_id).unwrap();

        // Remove cell
        let result = coordinator.remove_cell(cell_id);
        assert!(result, "Should remove cell successfully");

        // Verify cell is gone
        let swarm = coordinator.get_swarm(&swarm_id).unwrap();
        assert!(!swarm.cells.contains(&cell_id));
    }

    // === Codec Tests ===

    #[tokio::test]
    async fn test_cbor_bytes_roundtrip() {
        use crate::codec::{CborCodec, CborBytes};

        let original = b"hello world".to_vec();
        let encoded = CborCodec::encode(&CborBytes::new(original.clone())).unwrap();
        let decoded = CborCodec::decode::<CborBytes>(&encoded).unwrap();
        assert_eq!(decoded.as_ref(), original.as_slice());
    }

    #[tokio::test]
    async fn test_cbor_string_roundtrip() {
        use crate::codec::{CborCodec, CborString};

        let original = "test string with unicode".to_string();
        let encoded = CborCodec::encode(&CborString::new(original.clone())).unwrap();
        let decoded = CborCodec::decode::<CborString>(&encoded).unwrap();
        assert_eq!(decoded.as_ref(), original);
    }

    #[tokio::test]
    async fn test_tagged_value_encoding() {
        use crate::codec::{CborCodec, TaggedValue};

        let value = TaggedValue::new("my-type", vec![1u8, 2, 3]);
        let encoded = CborCodec::encode(&value).unwrap();
        let decoded: TaggedValue<Vec<u8>> = CborCodec::decode(&encoded).unwrap();
        assert_eq!(decoded.tag, "my-type");
        assert_eq!(decoded.value, vec![1, 2, 3]);
    }

    // === Scheduler Tests ===

    #[tokio::test]
    async fn test_scheduler_node_registration() {
        use crate::hypercore::{Scheduler, SchedulerConfig, Node, NodeResources};
        use crate::core::ExecutionLocation;

        let scheduler = Scheduler::new(SchedulerConfig::default());

        let node = Node::new(
            "node-1",
            ExecutionLocation::Cloud,
            NodeResources::new(4, 8192),
        );

        let result = scheduler.register_node(node).await;
        assert!(result.is_ok());

        let nodes = scheduler.list_nodes().await;
        assert_eq!(nodes.len(), 1);
    }

    #[tokio::test]
    async fn test_scheduler_placement_decision() {
        use crate::hypercore::{Scheduler, SchedulerConfig, ScheduleRequest, Node, NodeResources};
        use crate::core::{ExecutionLocation, PlacementHint};

        let scheduler = Scheduler::new(SchedulerConfig::default());

        // Register a node
        let node = Node::new(
            "cloud-node",
            ExecutionLocation::Cloud,
            NodeResources::new(4, 8192),
        );
        scheduler.register_node(node).await.unwrap();

        // Create schedule request
        let request = ScheduleRequest {
            cell_id: CellId::new(),
            required_vcpus: 2,
            required_memory: 4096,
            placement_hint: Some(PlacementHint::default()),
        };

        // Schedule the cell
        let response = scheduler.schedule(request).await;
        assert!(response.is_ok(), "Should schedule successfully");

        let decision = response.unwrap().decision;
        assert_eq!(decision.location, ExecutionLocation::Cloud);
    }

    #[tokio::test]
    async fn test_scheduler_stats_tracking() {
        use crate::hypercore::{Scheduler, SchedulerConfig, ScheduleRequest, Node, NodeResources};
        use crate::core::ExecutionLocation;

        let scheduler = Scheduler::new(SchedulerConfig::default());

        let node = Node::new(
            "stats-node",
            ExecutionLocation::Cloud,
            NodeResources::new(4, 8192),
        );
        scheduler.register_node(node).await.unwrap();

        // Make some schedules
        for _ in 0..3 {
            let request = ScheduleRequest {
                cell_id: CellId::new(),
                required_vcpus: 1,
                required_memory: 1024,
                placement_hint: None,
            };
            let _ = scheduler.schedule(request).await;
        }

        // Check stats
        let stats = scheduler.get_stats().await;
        assert_eq!(stats.total_scheduled, 3);
    }

    // === StateStore Tests ===

    #[tokio::test]
    async fn test_state_store_creation() {
        use crate::state_stream::StateStore;
        use crate::core::ValueEncoding;

        let store = StateStore::new();
        let key = crate::state_stream::StateKey::new("cell1", "key1");
        let slice = crate::state_stream::StateSlice::new(key, b"value".to_vec(), ValueEncoding::Raw);

        // Set via StateSlice API
        let version = store.set(slice).await.unwrap();
        assert_eq!(version, 1);
    }

    // === Connection State Tests ===

    #[test]
    fn test_connection_state_variants() {
        use crate::mesh::ConnectionState;

        // Just verify the enum variants exist
        let states = [
            ConnectionState::Disconnected,
            ConnectionState::Connecting,
            ConnectionState::Connected,
        ];

        for state in states {
            let desc = format!("{:?}", state);
            assert!(!desc.is_empty());
        }
    }

    // === Neural Optimizer Type Tests ===

    #[test]
    fn test_neural_optimizer_types_exist() {
        // Verify types exist and can be instantiated
        use crate::neural::{ExecutionProfile, ExecutionOutcome, ExecutionFeatures};
        use crate::core::ExecutionMetrics;

        let cell_id = CellId::new();
        let _profile = ExecutionProfile {
            cell_id,
            metrics: ExecutionMetrics::default(),
            features: ExecutionFeatures {
                input_size_bytes: 512,
                memory_limit_mb: 128,
                vcpus: 1,
                gpu_used: false,
                execution_location: "cloud".to_string(),
                time_of_day: 10.0,
                day_of_week: 1,
            },
            outcome: ExecutionOutcome::Success,
        };
    }

    #[test]
    fn test_execution_outcome_variants() {
        use crate::neural::ExecutionOutcome;

        // Verify outcome variants exist
        let outcomes = [
            ExecutionOutcome::Success,
            ExecutionOutcome::Error,
            ExecutionOutcome::Timeout,
            ExecutionOutcome::OOM,
        ];

        for outcome in outcomes {
            let desc = format!("{:?}", outcome);
            assert!(!desc.is_empty());
        }
    }

    // === Cell Metadata Tests ===

    #[tokio::test]
    async fn test_cell_metadata_creation() {
        let metadata = CellMetadata::new("test-cell", "wasm");

        assert_eq!(metadata.name, "test-cell");
        assert_eq!(metadata.runtime, "wasm");
        assert_eq!(metadata.version, "1.0.0");
    }

    // === Placement Hint Tests ===

    #[test]
    fn test_placement_hint_defaults() {
        use crate::core::PlacementHint;

        let hint = PlacementHint::default();

        assert_eq!(hint.latency_sensitivity, 0.5);
        assert_eq!(hint.cost_sensitivity, 0.5);
        assert!(!hint.gpu_required);
        assert!(hint.model_affinity.is_none());
    }

    // === CRDT Merge Tests ===

    #[tokio::test]
    async fn test_crdt_engine_exists() {
        use crate::state_stream::CrdtEngine;

        let _engine = CrdtEngine::new("test-node".to_string());
    }

    // === Cell Resources Tests ===

    #[test]
    fn test_cell_resources_defaults() {
        use crate::core::CellResources;

        let resources = CellResources::default();

        assert_eq!(resources.vcpus, 1);
        assert_eq!(resources.memory_bytes, 128 * 1024 * 1024);
        assert!(!resources.gpu_required);
        assert_eq!(resources.gpu_memory_mb, 0);
        assert_eq!(resources.cost_weight, 1.0);
    }
}