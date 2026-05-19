//! Core types for Prism Runtime
//!
//! This module defines the fundamental types used throughout Prism,
//! including Adaptive Execution Cells (AECs), execution metrics,
//! and common error types.

mod cell;
mod error;
mod metrics;
mod types;

pub use cell::{ExecutionCell, CellId, CellStatus, CellConfig, CellResources, CellMetadata, PlacementHint, ExecutionLocation, ExecutionTarget};
pub use error::{PrismError, PrismResult};
pub use metrics::{ExecutionMetrics, CostProfile, PerformanceProfile};
pub use types::{ValueEncoding, StreamStrategy, StreamConfig, RetentionPolicy};