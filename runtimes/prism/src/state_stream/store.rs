//! State store for StateStream Memory Fabric

use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use crate::codec::{CborCodec, CodecError};
use crate::core::{ValueEncoding, PrismResult};

/// A state key for the state store
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct StateKey {
    pub cell_id: String,
    pub key: String,
    pub namespace: Option<String>,
}

impl StateKey {
    pub fn new(cell_id: &str, key: &str) -> Self {
        Self {
            cell_id: cell_id.to_string(),
            key: key.to_string(),
            namespace: None,
        }
    }

    pub fn with_namespace(mut self, ns: &str) -> Self {
        self.namespace = Some(ns.to_string());
        self
    }

    /// Serialize this key to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize a key from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }

    /// Convert to a CBOR hex string for debugging/logging
    pub fn to_cbor_hex(&self) -> Result<String, CodecError> {
        let bytes = self.to_cbor()?;
        Ok(bytes.iter().map(|b| format!("{:02x}", b)).collect())
    }
}

/// A slice of state in the distributed store
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StateSlice {
    pub key: StateKey,
    pub value: Vec<u8>,
    pub encoding: ValueEncoding,
    pub version: u64,
    pub logical_timestamp: u64,
    pub is_final: bool,
    pub created_at: DateTime<Utc>,
}

impl StateSlice {
    pub fn new(key: StateKey, value: Vec<u8>, encoding: ValueEncoding) -> Self {
        Self {
            key,
            value,
            encoding,
            version: 0,
            logical_timestamp: 0,
            is_final: false,
            created_at: Utc::now(),
        }
    }

    pub fn with_version(mut self, version: u64) -> Self {
        self.version = version;
        self
    }

    /// Serialize this slice to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        CborCodec::encode(self)
    }

    /// Deserialize a slice from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, CodecError> {
        CborCodec::decode(bytes)
    }

    /// Encode the value to CBOR bytes (respects the encoding field)
    /// If encoding is Cbor, wraps in TaggedValue; otherwise converts
    pub fn value_to_cbor(&self) -> Result<Vec<u8>, CodecError> {
        match self.encoding {
            ValueEncoding::Cbor => {
                // Value is already CBOR, wrap in tagged container
                let tagged = crate::codec::TaggedValue::new("state.value", &self.value);
                CborCodec::encode(&tagged)
            }
            ValueEncoding::Json => {
                // Try to parse as JSON and re-encode as CBOR
                if let Ok(json_val) = serde_json::from_slice::<serde_json::Value>(&self.value) {
                    CborCodec::encode(&json_val)
                } else {
                    // Fallback: treat as raw bytes
                    let tagged = crate::codec::TaggedValue::new("state.raw", &self.value);
                    CborCodec::encode(&tagged)
                }
            }
            _ => {
                // Raw, MsgPack, Prost - wrap as raw bytes
                let tagged = crate::codec::TaggedValue::new("state.raw", &self.value);
                CborCodec::encode(&tagged)
            }
        }
    }

    /// Decode CBOR bytes to the value, respecting encoding
    pub fn value_from_cbor(&self, cbor_bytes: &[u8]) -> Result<Vec<u8>, CodecError> {
        match self.encoding {
            ValueEncoding::Cbor => {
                // Unwrap TaggedValue
                let tagged: crate::codec::TaggedValue<Vec<u8>> = CborCodec::decode(cbor_bytes)?;
                Ok(tagged.value)
            }
            ValueEncoding::Json => {
                // Decode as JSON Value and re-serialize as JSON bytes
                let json_val: serde_json::Value = CborCodec::decode(cbor_bytes)?;
                let json_bytes = serde_json::to_vec(&json_val)
                    .map_err(|e| std::io::Error::new(std::io::ErrorKind::InvalidData, e))?;
                Ok(json_bytes)
            }
            _ => {
                // Raw - return as-is
                Ok(cbor_bytes.to_vec())
            }
        }
    }
}

/// In-memory state store with optional persistence
pub struct StateStore {
    /// Main state storage
    state: Arc<RwLock<HashMap<StateKey, StateSlice>>>,
    /// Version map for CRDT-like versioning
    versions: Arc<RwLock<HashMap<StateKey, u64>>>,
    /// Persistence layer (optional)
    persistence: Option<Box<dyn StatePersistence>>,
}

/// Trait for state persistence
pub trait StatePersistence: Send + Sync {
    fn save(&self, key: &StateKey, slice: &StateSlice) -> PrismResult<()>;
    fn load(&self, key: &StateKey) -> PrismResult<Option<StateSlice>>;
    fn delete(&self, key: &StateKey) -> PrismResult<()>;
    fn list_keys(&self, cell_id: &str) -> PrismResult<Vec<StateKey>>;
}

impl StateStore {
    pub fn new() -> Self {
        Self {
            state: Arc::new(RwLock::new(HashMap::new())),
            versions: Arc::new(RwLock::new(HashMap::new())),
            persistence: None,
        }
    }

    pub fn with_persistence<P: StatePersistence + 'static>(persistence: P) -> Self {
        Self {
            state: Arc::new(RwLock::new(HashMap::new())),
            versions: Arc::new(RwLock::new(HashMap::new())),
            persistence: Some(Box::new(persistence)),
        }
    }

    /// Get a value from the store
    pub async fn get(&self, key: &StateKey) -> Option<StateSlice> {
        let state = self.state.read().await;
        state.get(key).cloned()
    }

    /// Set a value in the store
    pub async fn set(&self, slice: StateSlice) -> PrismResult<u64> {
        let mut state = self.state.write().await;
        let mut versions = self.versions.write().await;

        // Increment version
        let key = slice.key.clone();
        let new_version = *versions.entry(key).or_insert(0) + 1;

        let mut slice = slice;
        slice.version = new_version;

        // Persist if enabled
        if let Some(ref persist) = self.persistence {
            persist.save(&slice.key, &slice)?;
        }

        versions.insert(slice.key.clone(), new_version);
        state.insert(slice.key.clone(), slice);

        Ok(new_version)
    }

    /// Delete a value from the store
    pub async fn delete(&self, key: &StateKey) -> bool {
        let mut state = self.state.write().await;
        let mut versions = self.versions.write().await;

        if let Some(ref persist) = self.persistence {
            let _ = persist.delete(key);
        }

        versions.remove(key);
        state.remove(key).is_some()
    }

    /// List all keys for a cell
    pub async fn list_keys(&self, cell_id: &str) -> Vec<StateKey> {
        let state = self.state.read().await;
        state.keys()
            .filter(|k| k.cell_id == cell_id)
            .cloned()
            .collect()
    }

    /// Get the current version of a key
    pub async fn get_version(&self, key: &StateKey) -> u64 {
        let versions = self.versions.read().await;
        versions.get(key).copied().unwrap_or(0)
    }

    /// Clear all state for a cell
    pub async fn clear_cell(&self, cell_id: &str) {
        let mut state = self.state.write().await;
        let mut versions = self.versions.write().await;

        let keys: Vec<_> = state.keys()
            .filter(|k| k.cell_id == cell_id)
            .cloned()
            .collect();

        for key in keys {
            state.remove(&key);
            versions.remove(&key);
        }
    }

    /// Get all state slices for a cell (for snapshotting)
    pub async fn get_all_for_cell(&self, cell_id: &str) -> Vec<StateSlice> {
        let state = self.state.read().await;
        state.values()
            .filter(|s| s.key.cell_id == cell_id)
            .cloned()
            .collect()
    }
}

impl Default for StateStore {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_state_store_basic() {
        let store = StateStore::new();

        let key = StateKey::new("cell-1", "counter");
        let slice = StateSlice::new(key.clone(), b"0".to_vec(), ValueEncoding::Raw);

        let version = store.set(slice).await.unwrap();
        assert_eq!(version, 1);

        let retrieved = store.get(&key).await;
        assert!(retrieved.is_some());
        assert_eq!(retrieved.unwrap().value, b"0");
    }

    #[tokio::test]
    async fn test_state_store_versions() {
        let store = StateStore::new();

        let key = StateKey::new("cell-1", "counter");

        // Set initial value
        let slice1 = StateSlice::new(key.clone(), b"1".to_vec(), ValueEncoding::Raw);
        let v1 = store.set(slice1).await.unwrap();

        // Set new value
        let slice2 = StateSlice::new(key.clone(), b"2".to_vec(), ValueEncoding::Raw);
        let v2 = store.set(slice2).await.unwrap();

        assert_eq!(v1, 1);
        assert_eq!(v2, 2);
        assert_eq!(store.get_version(&key).await, 2);
    }
}