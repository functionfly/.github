//! Event store for event sourcing in StateStream Memory Fabric

use std::collections::VecDeque;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::codec::{CborCodec, CodecError};
use crate::core::PrismResult;

/// An event in the event store
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    pub event_id: String,
    pub event_type: EventType,
    pub aggregate_id: String,
    pub payload: Vec<u8>,
    pub metadata: EventMetadata,
    pub timestamp: DateTime<Utc>,
}

impl Event {
    pub fn new(aggregate_id: &str, event_type: EventType, payload: Vec<u8>) -> Self {
        Self {
            event_id: Uuid::new_v4().to_string(),
            event_type,
            aggregate_id: aggregate_id.to_string(),
            payload,
            metadata: EventMetadata::default(),
            timestamp: Utc::now(),
        }
    }

    /// Serialize this event to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize an event from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }

    /// Serialize the payload to CBOR (for storing structured data)
    pub fn payload_to_cbor<T: Serialize>(&mut self, value: &T) -> Result<(), CodecError> {
        self.payload = CborCodec::encode(value)?;
        Ok(())
    }

    /// Deserialize the payload from CBOR
    pub fn payload_from_cbor<T: for<'de> Deserialize<'de>>(&self) -> Result<T, CodecError> {
        CborCodec::decode(&self.payload)
    }
}

/// Type of event
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum EventType {
    CellCreated,
    CellExecuted,
    CellMigrated,
    CellFrozen,
    CellRestored,
    StateUpdated,
    SnapshotCreated,
    SwarmSpawned,
    SwarmCoordination,
}

impl EventType {
    pub fn as_str(&self) -> &'static str {
        match self {
            EventType::CellCreated => "cell.created",
            EventType::CellExecuted => "cell.executed",
            EventType::CellMigrated => "cell.migrated",
            EventType::CellFrozen => "cell.frozen",
            EventType::CellRestored => "cell.restored",
            EventType::StateUpdated => "state.updated",
            EventType::SnapshotCreated => "snapshot.created",
            EventType::SwarmSpawned => "swarm.spawned",
            EventType::SwarmCoordination => "swarm.coordination",
        }
    }
}

/// Metadata attached to an event
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EventMetadata {
    pub correlation_id: Option<String>,
    pub causation_id: Option<String>,
    pub tenant_id: Option<String>,
    pub user_id: Option<String>,
    pub version: u64,
}

impl Default for EventMetadata {
    fn default() -> Self {
        Self {
            correlation_id: None,
            causation_id: None,
            tenant_id: None,
            user_id: None,
            version: 0,
        }
    }
}

/// Event store for event sourcing
pub struct EventStore {
    /// In-memory event storage
    events: VecDeque<Event>,
    /// Maximum events to retain
    max_events: usize,
    /// Current version
    version: u64,
}

impl EventStore {
    pub fn new(max_events: usize) -> Self {
        Self {
            events: VecDeque::new(),
            max_events,
            version: 0,
        }
    }

    /// Append an event to the store
    pub fn append(&mut self, mut event: Event) -> PrismResult<()> {
        self.version += 1;
        event.metadata.version = self.version;

        self.events.push_back(event);

        // Trim if exceeding max
        while self.events.len() > self.max_events {
            self.events.pop_front();
        }

        Ok(())
    }

    /// Get all events for an aggregate
    pub fn get_for_aggregate(&self, aggregate_id: &str) -> Vec<&Event> {
        self.events
            .iter()
            .filter(|e| e.aggregate_id == aggregate_id)
            .collect()
    }

    /// Get events in a version range
    pub fn get_from_version(&self, from_version: u64) -> Vec<&Event> {
        self.events
            .iter()
            .filter(|e| e.metadata.version >= from_version)
            .collect()
    }

    /// Get the current version
    pub fn current_version(&self) -> u64 {
        self.version
    }

    /// Get total event count
    pub fn count(&self) -> usize {
        self.events.len()
    }

    /// Clear all events
    pub fn clear(&mut self) {
        self.events.clear();
        self.version = 0;
    }

    /// Serialize entire event store to CBOR
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        #[derive(Serialize)]
        struct StoreSnapshot<'a> {
            events: &'a VecDeque<Event>,
            version: u64,
            max_events: usize,
        }
        let snapshot = StoreSnapshot {
            events: &self.events,
            version: self.version,
            max_events: self.max_events,
        };
        CborCodec::encode(&snapshot)
    }

    /// Deserialize event store from CBOR
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        #[derive(Deserialize)]
        struct StoreSnapshot {
            events: VecDeque<Event>,
            version: u64,
            max_events: usize,
        }
        let snapshot: StoreSnapshot = CborCodec::decode(bytes)?;
        Ok(Self {
            events: snapshot.events,
            version: snapshot.version,
            max_events: snapshot.max_events,
        })
    }

    /// Export all events as CBOR batch (for backup/replication)
    pub fn export_batch(&self) -> Result<Vec<u8>, CodecError> {
        #[derive(Serialize)]
        struct EventBatch<'a> {
            events: &'a [&'a Event],
            total_version: u64,
        }
        let event_refs: Vec<&Event> = self.events.iter().collect();
        let batch = EventBatch {
            events: &event_refs,
            total_version: self.version,
        };
        CborCodec::encode(&batch)
    }

    /// Import events from CBOR batch (for restore/replication)
    pub fn import_batch(&mut self, bytes: &[u8]) -> Result<usize, CodecError> {
        #[derive(Deserialize)]
        struct EventBatch {
            events: Vec<Event>,
            total_version: u64,
        }
        let batch: EventBatch = CborCodec::decode(bytes)?;
        let count = batch.events.len();
        for event in batch.events {
            self.append(event)?;
        }
        // Only restore version if higher (prevents regression)
        if batch.total_version > self.version {
            self.version = batch.total_version;
        }
        Ok(count)
    }
}

impl Default for EventStore {
    fn default() -> Self {
        Self::new(10_000)
    }
}