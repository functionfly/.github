//! Quantum Snapshotting
//!
//! A huge differentiator. Execution cells can:
//! - freeze instantly
//! - serialize full runtime state
//! - migrate to another machine
//! - resume in milliseconds
//!
//! Like VM live migration + game save states + AI memory persistence.
//!
//! Enables:
//! - failover
//! - cost optimization
//! - mobile robotics
//! - intermittent connectivity
//! - edge handoff

pub mod snapshot;
mod migration;
mod checkpoint;
mod compression;

pub use snapshot::{Snapshot, SnapshotManager, SnapshotType, SnapshotMetadata, CompressionAlgorithm};
pub use snapshot::{WasmCpuState, GlobalState, TableState, HandleSnapshot, HandleType};
pub use migration::{MigrationManager, MigrationStrategy, MigrationResult};
pub use checkpoint::{Checkpoint, CheckpointManager};
pub use compression::{compress_snapshot, decompress_snapshot};