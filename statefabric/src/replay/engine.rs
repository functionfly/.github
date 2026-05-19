//! Deterministic replay engine

use serde::{Deserialize, Serialize};
use thiserror::Error;
use uuid::Uuid;

use crate::models::{Event, EventType};

/// Errors that can occur during replay
#[derive(Error, Debug)]
pub enum ReplayError {
    #[error("State not found: {0}")]
    StateNotFound(Uuid),

    #[error("Event not found: {0}")]
    EventNotFound(Uuid),

    #[error("Invalid sequence: {0}")]
    InvalidSequence(String),

    #[error("Hash mismatch: {0}")]
    HashMismatch(String),

    #[error("Replay failed: {0}")]
    ReplayFailed(String),
}

/// Result type for replay operations
pub type ReplayResult<T> = Result<T, ReplayError>;

/// Replay options
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReplayOptions {
    /// Start sequence number (inclusive)
    pub from_sequence: Option<i64>,
    /// End sequence number (inclusive)
    pub to_sequence: Option<i64>,
    /// Whether to verify hashes
    pub verify: bool,
    /// Whether to apply side effects (false for dry-run)
    pub dry_run: bool,
}

impl Default for ReplayOptions {
    fn default() -> Self {
        Self {
            from_sequence: None,
            to_sequence: None,
            verify: true,
            dry_run: false,
        }
    }
}

/// Replay status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReplayStatus {
    /// State ID
    pub state_id: Uuid,
    /// Starting sequence
    pub from_sequence: i64,
    /// Ending sequence
    pub to_sequence: i64,
    /// Events replayed
    pub events_replayed: i64,
    /// Final hash
    pub final_hash: String,
    /// Whether verification passed
    pub verified: bool,
}

/// Replay engine for deterministic state reconstruction
pub struct ReplayEngine;

impl ReplayEngine {
    /// Create a new replay engine
    pub fn new() -> Self {
        Self
    }

    /// Replay events to reconstruct state
    pub fn replay(
        &self,
        events: &[Event],
        initial_state: Option<&serde_json::Value>,
        _options: &ReplayOptions,
    ) -> ReplayResult<serde_json::Value> {
        let mut state = initial_state
            .cloned()
            .unwrap_or_else(|| serde_json::json!({}));

        let state_obj = state.as_object_mut()
            .ok_or_else(|| ReplayError::ReplayFailed("State must be an object".to_string()))?;

        for event in events {
            self.apply_event(state_obj, event)?;
        }

        Ok(state)
    }

    /// Apply a single event to state
    fn apply_event(
        &self,
        state: &mut serde_json::Map<String, serde_json::Value>,
        event: &Event,
    ) -> ReplayResult<()> {
        match event.event_type {
            EventType::Set => {
                if let (Some(key), Some(value)) = (&event.key, &event.new_value) {
                    state.insert(key.clone(), value.clone());
                }
            }
            EventType::Delete => {
                if let Some(key) = &event.key {
                    state.remove(key);
                }
            }
            EventType::Merge => {
                if let Some(value) = &event.new_value {
                    if let Some(obj) = value.as_object() {
                        for (k, v) in obj {
                            state.insert(k.clone(), v.clone());
                        }
                    }
                }
            }
            EventType::Clear => {
                state.clear();
            }
            EventType::Snapshot | EventType::Restore => {
                // These don't modify state directly in replay
                // Snapshot is handled separately
            }
        }

        Ok(())
    }

    /// Verify event chain integrity
    pub fn verify_chain(
        &self,
        events: &[Event],
        expected_hash: &str,
    ) -> ReplayResult<bool> {
        use crate::replay::compute_event_chain_hash;

        let actual_hash = compute_event_chain_hash(events);

        if expected_hash != &actual_hash {
            return Err(ReplayError::HashMismatch(format!(
                "Expected {}, got {}",
                expected_hash, actual_hash
            )));
        }

        Ok(true)
    }

    /// Get replay window (events between two sequence numbers)
    pub fn get_window(
        &self,
        from_sequence: i64,
        to_sequence: i64,
    ) -> ReplayResult<(i64, i64)> {
        if from_sequence > to_sequence {
            return Err(ReplayError::InvalidSequence(
                "from_sequence must be <= to_sequence".to_string(),
            ));
        }

        Ok((from_sequence, to_sequence))
    }
}

impl Default for ReplayEngine {
    fn default() -> Self {
        Self::new()
    }
}

/// Snapshot-based replay (faster than full replay)
pub struct SnapshotReplay;

impl SnapshotReplay {
    /// Create a new snapshot replay engine
    pub fn new() -> Self {
        Self
    }

    /// Replay from snapshot + delta events
    pub fn replay_from_snapshot(
        &self,
        snapshot: &serde_json::Value,
        delta_events: &[Event],
        _options: &ReplayOptions,
    ) -> ReplayResult<serde_json::Value> {
        let mut state = snapshot.clone();

        let state_obj = state.as_object_mut()
            .ok_or_else(|| ReplayError::ReplayFailed("Snapshot must be an object".to_string()))?;

        let engine = ReplayEngine::new();

        for event in delta_events {
            engine.apply_event(state_obj, event)?;
        }

        Ok(state)
    }
}

impl Default for SnapshotReplay {
    fn default() -> Self {
        Self::new()
    }
}
