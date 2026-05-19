//! Autonomous Function Swarms
//!
//! Functions can:
//! - spawn sub-functions
//! - negotiate resources
//! - delegate workloads
//! - form temporary clusters
//! - self-heal
//!
//! Like a multi-agent system where each function is a capable agent.

mod commands;
mod coordinator;
mod swarm;
mod protocol;
mod health;

pub use commands::SwarmCommand;
pub use coordinator::{SwarmCoordinator, CoordinatorConfig, Swarm, SwarmId, SwarmState, SwarmHealth};
pub use swarm::{SwarmMessage, SwarmMessageType};
pub use protocol::SwarmProtocol;
pub use health::{HealthMonitor, HealthCheck, SelfHealAction};