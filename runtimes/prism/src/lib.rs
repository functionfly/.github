//! FunctionFly Prism Runtime
//!
//! Universal Adaptive WASM Execution Fabric for AI-native execution,
//! autonomous agents, robotics, edge systems, and cross-language functions.
//!
//! # Core Vision
//!
//! Prism is not "just another WASM runtime." It's a distributed AI-native
//! universal WASM execution fabric that:
//! - understands AI workflows
//! - dynamically adapts execution
//! - coordinates distributed intelligence
//! - streams state between agents/functions
//! - executes across cloud, browser, robotics, edge, and local systems
//! - treats tools/functions as living distributed capabilities
//!
//! # Architecture
//!
//! The runtime consists of several core layers:
//!
//! - **HyperCore Scheduler**: AI compute traffic controller for intelligent placement
//! - **WASM Fusion Engine**: Dynamic execution graphs with module merging
//! - **Universal Capability Layer (UCL)**: DNS for AI capabilities
//! - **StateStream Memory Fabric**: Distributed streaming memory with CRDT
//! - **Quantum Snapshotting**: Live migration and state serialization
//! - **Mesh Networking**: P2P capability mesh with libp2p/QUIC
//! - **Neural Execution Optimization**: RL-based self-optimization
//! - **Autonomous Function Swarms**: Self-healing, coordinated function clusters

pub mod core;
pub mod hypercore;
pub mod wasm_fusion;
pub mod ucl;
pub mod state_stream;
pub mod quantum;
pub mod mesh;
pub mod neural;
pub mod swarm;
pub mod cli;
pub mod nats_client;
pub mod codec;
pub mod proto; // Protobuf generated types
pub mod runtime;
pub mod security;

#[cfg(test)]
pub mod integration_tests;

// Re-export commonly used types
pub use core::{
    CellId, CellStatus, CellConfig, CellResources, CellMetadata,
    ExecutionMetrics, ExecutionCell, ExecutionTarget, PlacementHint,
};
pub use hypercore::{Scheduler, PlacementDecision, ScheduleRequest, ScheduleResponse};
pub use wasm_fusion::{FusionGraph, FusionNode, FusionNodeType, FusionEdge, FusionExecutor, NodeConfig, WasmComposer, ComposerConfig, CompositionResult};
pub use ucl::{Capability, CapabilityRegistry, CapabilityDiscovery, CapabilityCategory};
pub use state_stream::{StateStore, StateSlice};
pub use core::StreamConfig;
pub use quantum::{Snapshot, SnapshotManager, MigrationManager};
pub use mesh::{MeshNetwork, Peer, PeerId, ConnectionState, RoutingTable};
pub use neural::{NeuralOptimizer, ExecutionProfile, ProfileCollector};
pub use swarm::{SwarmCoordinator, SwarmId, SwarmState, SwarmHealth, SwarmCommand};
pub use nats_client::{NatsOrchestratorClient, OrchestratorMessage, OrchestratorMessageType, CapabilityInfo};
pub use codec::{CborCodec, CodecError, TaggedValue, CborBytes, CborString};
pub use runtime::{RuntimeContext, RuntimeStatus};
pub use security::{
    SecurityManager, SecurityPolicy, SecurityAuditor, SecurityEvent, SecurityEventType,
    ExecutionSecurityContext, EnclaveType, EnclaveAttestation, Permission, PermissionSet,
    WasmValidationResult, WasmViolation, ViolationSeverity, SecurityError,
};

/// Version of the Prism runtime
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version_exists() {
        assert!(!VERSION.is_empty());
    }
}