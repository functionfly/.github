//! Checkpoint management for Quantum Snapshotting

use std::collections::HashMap;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::codec::{CborCodec, CodecError};
use crate::core::CellId;

/// A checkpoint for resumable execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Checkpoint {
    pub checkpoint_id: String,
    pub cell_id: CellId,
    pub epoch: u64,
    pub timestamp: DateTime<Utc>,
    pub state_hash: String,
    pub metadata: CheckpointMetadata,
}

impl Checkpoint {
    pub fn new(cell_id: CellId, epoch: u64, state_hash: &str) -> Self {
        Self {
            checkpoint_id: Uuid::new_v4().to_string(),
            cell_id,
            epoch,
            timestamp: Utc::now(),
            state_hash: state_hash.to_string(),
            metadata: CheckpointMetadata::default(),
        }
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

impl CheckpointMetadata {
    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }
}

/// Checkpoint metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CheckpointMetadata {
    pub includes_memory: bool,
    pub includes_cpu_state: bool,
    pub is_fresh: bool,
}

impl Default for CheckpointMetadata {
    fn default() -> Self {
        Self {
            includes_memory: true,
            includes_cpu_state: true,
            is_fresh: false,
        }
    }
}

/// Manages checkpoints for resumable execution
pub struct CheckpointManager {
    checkpoints: HashMap<String, Checkpoint>,
    latest_per_cell: HashMap<String, String>,
}

impl CheckpointManager {
    pub fn new() -> Self {
        Self {
            checkpoints: HashMap::new(),
            latest_per_cell: HashMap::new(),
        }
    }

    /// Create a checkpoint
    pub fn create_checkpoint(&mut self, checkpoint: Checkpoint) {
        let cell_id_str = checkpoint.cell_id.to_string();
        self.checkpoints.insert(checkpoint.checkpoint_id.clone(), checkpoint.clone());
        self.latest_per_cell.insert(cell_id_str, checkpoint.checkpoint_id);
    }

    /// Get the latest checkpoint for a cell
    pub fn get_latest(&self, cell_id: &CellId) -> Option<&Checkpoint> {
        self.latest_per_cell.get(&cell_id.to_string())
            .and_then(|id| self.checkpoints.get(id))
    }

    /// Get a specific checkpoint
    pub fn get(&self, checkpoint_id: &str) -> Option<&Checkpoint> {
        self.checkpoints.get(checkpoint_id)
    }

    /// List all checkpoints for a cell
    pub fn list_for_cell(&self, cell_id: &CellId) -> Vec<&Checkpoint> {
        self.checkpoints.values()
            .filter(|c| c.cell_id == *cell_id)
            .collect()
    }

    /// Delete old checkpoints
    pub fn prune(&mut self, cell_id: &CellId, keep_count: usize) {
        // First collect the checkpoint IDs to avoid borrow conflict
        let checkpoint_ids: Vec<_> = self.checkpoints.values()
            .filter(|c| c.cell_id == *cell_id)
            .map(|c| c.checkpoint_id.clone())
            .collect();

        let mut sorted_ids = checkpoint_ids;
        sorted_ids.sort_by(|a, b| {
            // Compare by epoch - we need to look up epochs from checkpoints
            let epoch_a = self.checkpoints.get(a).map(|c| c.epoch).unwrap_or(0);
            let epoch_b = self.checkpoints.get(b).map(|c| c.epoch).unwrap_or(0);
            epoch_b.cmp(&epoch_a)
        });

        for checkpoint_id in sorted_ids.into_iter().skip(keep_count) {
            self.checkpoints.remove(&checkpoint_id);
        }
    }
}

impl Default for CheckpointManager {
    fn default() -> Self {
        Self::new()
    }
}