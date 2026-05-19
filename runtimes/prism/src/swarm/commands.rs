//! Swarm commands module
//!
//! Commands that can be sent to a swarm for coordination and self-healing

/// Commands that can be sent to a swarm
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SwarmCommand {
    /// Spawn a replacement cell to replace a failed one
    SpawnReplacement,
    /// Notify peer nodes about swarm state changes
    NotifyPeers,
    /// Migrate a cell to another node for load balancing or fault tolerance
    MigrateCell,
    /// Terminate a cell that is stuck or consuming resources
    TerminateCell,
    /// Coordinate work among cells in the swarm
    CoordinateWork,
    /// Request state synchronization from peers (used after cell failure)
    RequestStateSync,
    /// Elect a new leader cell when current leader fails
    ElectLeader,
    /// Redistribute workload across healthy cells
    RedistributeLoad,
    /// Scale up swarm capacity by spawning additional cells
    ScaleUp,
    /// Scale down swarm capacity by removing idle cells
    ScaleDown,
}

impl SwarmCommand {
    /// Get the priority level of this command (lower = higher priority)
    pub fn priority(&self) -> u8 {
        match self {
            SwarmCommand::ElectLeader => 1,
            SwarmCommand::RequestStateSync => 2,
            SwarmCommand::SpawnReplacement => 3,
            SwarmCommand::RedistributeLoad => 4,
            SwarmCommand::NotifyPeers => 5,
            SwarmCommand::MigrateCell => 6,
            SwarmCommand::ScaleUp => 7,
            SwarmCommand::CoordinateWork => 8,
            SwarmCommand::ScaleDown => 9,
            SwarmCommand::TerminateCell => 10,
        }
    }

    /// Check if this command requires quorum (needs majority agreement)
    pub fn requires_quorum(&self) -> bool {
        matches!(
            self,
            SwarmCommand::ElectLeader
                | SwarmCommand::MigrateCell
                | SwarmCommand::RedistributeLoad
                | SwarmCommand::ScaleDown
        )
    }

    /// Check if this command affects multiple cells
    pub fn is_multi_cell(&self) -> bool {
        matches!(
            self,
            SwarmCommand::RedistributeLoad
                | SwarmCommand::ScaleUp
                | SwarmCommand::ScaleDown
                | SwarmCommand::NotifyPeers
        )
    }
}