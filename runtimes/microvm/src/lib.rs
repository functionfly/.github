//! FunctionFly MicroVM Orchestrator
//!
//! This library provides CPython execution inside Firecracker microVMs
//! for Enterprise tier customers.

pub mod executor;
pub mod firecracker;
pub mod orchestrator;
pub mod vsock;

// Re-export main types for convenience
pub use orchestrator::{MicroVMOrchestrator, ExecutionRequest, ExecutionResult, OrchestratorStats};
pub use firecracker::{FirecrackerClient, VMConfig, VMInstance, VMState};
pub use vsock::{VsockClient, VsockServer};