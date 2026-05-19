//! WASM Fusion Engine
//!
//! Dynamic execution graphs with module merging and real WASM execution.
//! - merges multiple WASM modules live
//! - creates dynamic execution graphs
//! - supports streaming function composition
//! - enables runtime patching
//! - supports "fluid execution"

pub mod composer;
pub mod engine;
pub mod fusion;
pub mod graph;
pub mod linker;

pub use composer::{WasmComposer, ComposerConfig, CompositionResult};
pub use engine::{FusionEngine as WasmFusionEngine, FusionEngineConfig, WasmExecutionResult};
pub use fusion::{FusionGraph, FusionNode, FusionEdge, FusionExecutor, FusionConfig, FusionNodeType, NodeConfig, FusionEngine as RealFusionEngine};
pub use graph::{ExecutionGraph, NodeId, NodeResult};
pub use linker::FusionLinker;