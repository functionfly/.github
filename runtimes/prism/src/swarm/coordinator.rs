//! Swarm coordinator for autonomous function coordination

use std::collections::{HashMap, HashSet};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::codec::{CborCodec, CodecError};
use crate::core::{CellId, PrismError, PrismResult};
use crate::swarm::commands::SwarmCommand;

/// A swarm of coordinated cells
#[derive(Debug, Clone)]
pub struct Swarm {
    pub swarm_id: SwarmId,
    pub cells: HashSet<CellId>,
    pub peer_nodes: HashMap<String, String>, // node_id -> address
    pub state: SwarmState,
    pub created_at: DateTime<Utc>,
}

impl Swarm {
    pub fn new(swarm_id: SwarmId) -> Self {
        Self {
            swarm_id,
            cells: HashSet::new(),
            peer_nodes: HashMap::new(),
            state: SwarmState::default(),
            created_at: Utc::now(),
        }
    }

    pub fn add_cell(&mut self, cell_id: CellId) {
        self.cells.insert(cell_id);
    }

    pub fn remove_cell(&mut self, cell_id: CellId) {
        self.cells.remove(&cell_id);
    }

    pub fn cell_count(&self) -> usize {
        self.cells.len()
    }
}

/// Unique identifier for a swarm
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct SwarmId(pub String);

impl SwarmId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }

    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }
}

impl std::fmt::Display for SwarmId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// State of a swarm
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SwarmState {
    pub active_cells: Vec<String>,
    pub completed_cells: Vec<String>,
    pub health: SwarmHealth,
}

impl SwarmState {
    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }
}

/// Health status of a swarm
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SwarmHealth {
    pub is_healthy: bool,
    pub active_count: u32,
    pub total_count: u32,
    pub failed_cells: Vec<String>,
}

impl SwarmHealth {
    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }

    /// Get health percentage
    pub fn health_percentage(&self) -> f32 {
        if self.total_count == 0 {
            1.0
        } else {
            self.active_count as f32 / self.total_count as f32
        }
    }
}

/// Configuration for swarm coordinator
#[derive(Debug, Clone)]
pub struct CoordinatorConfig {
    pub max_swarm_size: usize,
    pub health_check_interval_secs: u64,
    pub self_heal_enabled: bool,
    pub self_heal_threshold: f32,
}

impl Default for CoordinatorConfig {
    fn default() -> Self {
        Self {
            max_swarm_size: 100,
            health_check_interval_secs: 30,
            self_heal_enabled: true,
            self_heal_threshold: 0.7,
        }
    }
}

/// Main swarm coordinator
pub struct SwarmCoordinator {
    config: CoordinatorConfig,
    swarms: HashMap<SwarmId, Swarm>,
    cells_to_swarms: HashMap<CellId, SwarmId>,
}

impl SwarmCoordinator {
    pub fn new(config: CoordinatorConfig) -> Self {
        Self {
            config,
            swarms: HashMap::new(),
            cells_to_swarms: HashMap::new(),
        }
    }

    /// Create a new swarm
    pub fn create_swarm(&mut self, swarm_id: SwarmId) -> PrismResult<Swarm> {
        if self.swarms.contains_key(&swarm_id) {
            return Err(PrismError::SwarmError(format!("Swarm {} already exists", swarm_id)));
        }

        let swarm = Swarm::new(swarm_id.clone());
        self.swarms.insert(swarm_id.clone(), swarm.clone());
        Ok(swarm)
    }

    /// Add a cell to a swarm
    pub fn add_cell_to_swarm(&mut self, cell_id: CellId, swarm_id: &SwarmId) -> PrismResult<()> {
        let swarm = self.swarms.get_mut(swarm_id)
            .ok_or_else(|| PrismError::SwarmError(format!("Swarm {} not found", swarm_id)))?;

        if swarm.cell_count() >= self.config.max_swarm_size {
            return Err(PrismError::SwarmError("Swarm at max capacity".to_string()));
        }

        swarm.add_cell(cell_id);
        self.cells_to_swarms.insert(cell_id, swarm_id.clone());
        Ok(())
    }

    /// Remove a cell from its swarm
    pub fn remove_cell(&mut self, cell_id: CellId) -> bool {
        if let Some(swarm_id) = self.cells_to_swarms.remove(&cell_id) {
            if let Some(swarm) = self.swarms.get_mut(&swarm_id) {
                swarm.remove_cell(cell_id);
                return true;
            }
        }
        false
    }

    /// Get all swarms (for CLI display)
    pub fn swarms(&self) -> &HashMap<SwarmId, Swarm> {
        &self.swarms
    }

    /// Get a swarm by ID
    pub fn get_swarm(&self, swarm_id: &SwarmId) -> Option<&Swarm> {
        self.swarms.get(swarm_id)
    }

    /// Get the swarm for a cell
    pub fn get_swarm_for_cell(&self, cell_id: &CellId) -> Option<&Swarm> {
        self.cells_to_swarms.get(cell_id)
            .and_then(|id| self.swarms.get(id))
    }

    /// Update swarm health
    pub fn update_health(&mut self, swarm_id: &SwarmId, health: SwarmHealth) {
        if let Some(swarm) = self.swarms.get_mut(swarm_id) {
            swarm.state.health = health;
        }
    }

    /// Check swarm health and trigger self-heal if needed
    ///
    /// Analyzes the swarm's health state and generates appropriate healing actions
    /// based on the type and severity of issues detected.
    pub fn check_and_heal(&self, swarm_id: &SwarmId) -> PrismResult<Vec<SwarmCommand>> {
        let swarm = self.swarms.get(swarm_id)
            .ok_or_else(|| PrismError::SwarmError(format!("Swarm {} not found", swarm_id)))?;

        let health_ratio = swarm.state.health.active_count as f32 / swarm.state.health.total_count as f32;
        let healthy_threshold = self.config.self_heal_threshold;
        let critical_threshold = 0.5; // Below this is critical

        let mut commands = Vec::new();

        // Analyze health and determine appropriate actions
        if health_ratio < critical_threshold {
            // Critical: swarm is severely degraded
            tracing::warn!(swarm_id = %swarm_id, health_ratio, active = swarm.state.health.active_count, total = swarm.state.health.total_count, "Critical health - initiating emergency repair");

            // Count how many replacements needed
            let needed_replacements = swarm.state.health.failed_cells.len().max(1);

            for _ in 0..needed_replacements {
                commands.push(SwarmCommand::SpawnReplacement);
            }

            // Notify peers about the critical situation
            commands.push(SwarmCommand::NotifyPeers);
            commands.push(SwarmCommand::RequestStateSync);

            // If there's a coordinator cell, trigger it
            if swarm.state.health.active_count > 0 {
                commands.push(SwarmCommand::ElectLeader);
            }

        } else if health_ratio < healthy_threshold {
            // Warning: swarm is below healthy threshold but not critical
            tracing::info!(swarm_id = %swarm_id, health_ratio, "Swarm health degraded - initiating self-heal");

            // Determine number of replacements needed
            let active = swarm.state.health.active_count as i32;
            let total = swarm.state.health.total_count as i32;
            let target_active = (total as f32 * healthy_threshold) as i32;
            let replacements_needed = (target_active - active).max(1) as usize;

            for _ in 0..replacements_needed {
                commands.push(SwarmCommand::SpawnReplacement);
            }

            // Notify peers about the healing process
            commands.push(SwarmCommand::NotifyPeers);

            // Check if we need to redistribute load
            if !swarm.state.health.failed_cells.is_empty() {
                commands.push(SwarmCommand::RedistributeLoad);
            }

        } else {
            // Healthy: swarm is operating normally
            tracing::debug!(swarm_id = %swarm_id, health_ratio, "Swarm health nominal");

            // Still perform preventive maintenance
            if swarm.state.health.active_count < 5 {
                // Low cell count - consider growing the swarm
                commands.push(SwarmCommand::SpawnReplacement);
            }

            // If cells are near capacity, trigger capacity planning
            let avg_load = self.estimate_swarm_load(swarm);
            if avg_load > 0.8 {
                commands.push(SwarmCommand::ScaleUp);
            }
        }

        // Store the health check result for this swarm
        let health_commands = commands.clone();

        tracing::info!(
            swarm_id = %swarm_id,
            health_ratio,
            commands_generated = health_commands.len(),
            command_types = ?health_commands,
            "Health check complete"
        );

        Ok(commands)
    }

    /// Estimate the average load of the swarm based on active cells
    fn estimate_swarm_load(&self, swarm: &Swarm) -> f32 {
        if swarm.state.health.total_count == 0 {
            return 0.0;
        }

        // Estimate load based on active vs total ratio and failed cells
        let active_ratio = swarm.state.health.active_count as f32 / swarm.state.health.total_count as f32;
        let failure_penalty = swarm.state.health.failed_cells.len() as f32 * 0.1;

        (active_ratio - failure_penalty).max(0.0).min(1.0)
    }
}