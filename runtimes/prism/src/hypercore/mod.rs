//! HyperCore Scheduler
//!
//! An AI-aware distributed scheduler that decides:
//! - where functions execute (cloud, edge, browser, etc.)
//! - GPU vs CPU placement
//! - memory optimization
//! - latency-aware placement
//! - cost-aware execution
//! - AI-model affinity routing
//!
//! Think of it as the "AI compute traffic controller" for Prism.

mod scheduler;
mod placement;
mod node;
mod metrics;

pub use scheduler::{Scheduler, SchedulerConfig, ScheduleRequest, ScheduleResponse};
pub use placement::{PlacementEngine, PlacementDecision, PlacementScore, NodeId};
pub use node::{Node, NodeResources, NodeStatus, NodeInfo};
pub use metrics::SchedulerMetrics;