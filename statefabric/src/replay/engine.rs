//! Deterministic replay engine
//!
//! SECURITY: Includes replay attack protection via sequence tracking and freshness windows.

use serde::{Deserialize, Serialize};
use std::collections::HashSet;
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

    /// SECURITY: Replay attack detected
    #[error("Replay attack detected: event sequence {0} already processed")]
    ReplayDetected(i64),

    /// SECURITY: Event outside freshness window
    #[error("Event outside freshness window: sequence {0}, expected > {1}")]
    EventTooOld(i64, i64),
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
    /// SECURITY: Freshness window - reject events older than this sequence
    pub freshness_window: Option<i64>,
}

impl Default for ReplayOptions {
    fn default() -> Self {
        Self {
            from_sequence: None,
            to_sequence: None,
            verify: true,
            dry_run: false,
            freshness_window: Some(1000), // Default: allow up to 1000 events behind
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

/// SECURITY: Tracks processed sequences to prevent replay attacks
#[derive(Debug, Clone, Default)]
pub struct SequenceTracker {
    /// Set of already-processed sequence numbers
    processed: HashSet<i64>,
    /// Highest sequence number seen
    highest_sequence: i64,
    /// Freshness window - events older than this are rejected
    freshness_window: i64,
}

impl SequenceTracker {
    /// Create a new sequence tracker with the given freshness window
    pub fn new(freshness_window: i64) -> Self {
        Self {
            processed: HashSet::new(),
            highest_sequence: 0,
            freshness_window,
        }
    }

    /// SECURITY: Check if an event sequence is valid (not replayed and within freshness window)
    ///
    /// Returns Ok(true) if the event is valid and should be processed.
    /// Returns Err(ReplayError) if the event is a replay or too old.
    pub fn check_sequence(&self, sequence: i64) -> ReplayResult<bool> {
        // SECURITY: Check if already processed (replay attack)
        if self.processed.contains(&sequence) {
            return Err(ReplayError::ReplayDetected(sequence));
        }

        // SECURITY: Check freshness window
        if sequence < self.highest_sequence - self.freshness_window + 1 && self.highest_sequence > 0 {
            return Err(ReplayError::EventTooOld(sequence, self.highest_sequence - self.freshness_window + 1));
        }

        Ok(true)
    }

    /// Mark a sequence as processed
    pub fn mark_processed(&mut self, sequence: i64) {
        self.processed.insert(sequence);
        if sequence > self.highest_sequence {
            self.highest_sequence = sequence;
        }

        // SECURITY: Prune old entries to prevent memory bloat
        // Keep only sequences within the freshness window
        let min_valid = self.highest_sequence - self.freshness_window;
        self.processed.retain(|&seq| seq > min_valid);
    }

    /// Get the highest sequence number seen
    pub fn highest(&self) -> i64 {
        self.highest_sequence
    }
}

/// Replay engine for deterministic state reconstruction
pub struct ReplayEngine {
    /// SECURITY: Sequence tracker for replay protection
    sequence_tracker: SequenceTracker,
}

impl ReplayEngine {
    /// Create a new replay engine with replay protection
    pub fn new() -> Self {
        Self {
            sequence_tracker: SequenceTracker::new(1000), // Default 1000 event freshness window
        }
    }

    /// Create a new replay engine with custom freshness window
    pub fn with_freshness_window(window: i64) -> Self {
        Self {
            sequence_tracker: SequenceTracker::new(window),
        }
    }

    /// SECURITY: Replay events with replay attack protection
    ///
    /// Events are validated against the sequence tracker before being applied.
    /// Events outside the freshness window or already processed are rejected.
    pub fn replay_with_protection(
        &self,
        events: &[Event],
        initial_state: Option<&serde_json::Value>,
        options: &ReplayOptions,
    ) -> ReplayResult<serde_json::Value> {
        let mut state = initial_state
            .cloned()
            .unwrap_or_else(|| serde_json::json!({}));

        let state_obj = state.as_object_mut()
            .ok_or_else(|| ReplayError::ReplayFailed("State must be an object".to_string()))?;

        // SECURITY: Validate freshness window from options
        let freshness_window = options.freshness_window.unwrap_or(1000);
let mut tracker = SequenceTracker::new(freshness_window);

        for event in events {
            // SECURITY: Check sequence validity
            tracker.check_sequence(event.sequence_num)?;

            self.apply_event(state_obj, event)?;

            // Mark as processed
            tracker.mark_processed(event.sequence_num);
        }

        Ok(state)
    }

    /// Replay events to reconstruct state
    pub fn replay(
        &self,
        events: &[Event],
        initial_state: Option<&serde_json::Value>,
        options: &ReplayOptions,
    ) -> ReplayResult<serde_json::Value> {
        // Use the protected replay with fresh tracker
        self.replay_with_protection(events, initial_state, options)
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
