//! State manager - core state operations

use std::collections::HashMap;
use std::fmt;
use std::sync::Arc;
use thiserror::Error;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::cache::{RedisCache, RedisConfig, LruCache};
use crate::models::{Event, EventType, Snapshot, SourceType, CreateSnapshotRequest, RestoreSnapshotRequest, SnapshotMetadata, EventMetadata};
use crate::replay::compute_state_hash;
use crate::storage::{ObjectStore, PostgresSnapshotRepository, PostgresEventRepository, PostgresStateRepository};
use crate::wasm::{WasmRuntime, WasmConfig, ExecutionResult};

/// Errors that can occur in state management
#[derive(Error, Debug)]
pub enum StateError {
    #[error("State not found: {0}")]
    NotFound(Uuid),

    #[error("Key not found: {0}")]
    KeyNotFound(String),

    #[error("Invalid operation: {0}")]
    InvalidOperation(String),

    #[error("Storage error: {0}")]
    StorageError(String),
}

/// Result type for state operations
pub type StateResult<T> = Result<T, StateError>;

/// State manager with multi-layer caching (Phase 1 + Phase 2)
pub struct StateManager {
    /// Phase 1: In-memory LRU cache for hot data
    lru_cache: Arc<RwLock<LruCache<Uuid, serde_json::Value>>>,
    /// Phase 2: Redis cache for distributed caching (optional)
    redis_cache: Option<Arc<RedisCache>>,
    /// Sequence numbers per state
    sequences: Arc<RwLock<HashMap<Uuid, i64>>>,
    /// Object storage for snapshots
    object_store: Option<Box<dyn ObjectStore + Send + Sync>>,
    /// PostgreSQL snapshot repository
    snapshot_repo: Option<PostgresSnapshotRepository>,
    /// PostgreSQL event repository
    event_repo: Option<PostgresEventRepository>,
    /// PostgreSQL state repository for tenant verification
    state_repo: Option<PostgresStateRepository>,
    /// WASM runtime for function execution
    wasm_runtime: Option<WasmRuntime>,
}

impl fmt::Debug for StateManager {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("StateManager")
            .field("lru_cache", &self.lru_cache)
            .field("redis_cache", &self.redis_cache)
            .field("sequences", &self.sequences)
            .field("object_store", &self.object_store.as_ref().map(|_| "..."))
            .field("snapshot_repo", &self.snapshot_repo)
            .field("event_repo", &self.event_repo)
            .field("state_repo", &self.state_repo)
            .field("wasm_runtime", &self.wasm_runtime)
            .finish()
    }
}

impl StateManager {
    /// Create a new state manager with default LRU cache (Phase 1)
    pub fn new() -> Self {
        Self {
            lru_cache: Arc::new(RwLock::new(LruCache::new(1000))), // Default 1000 entries
            redis_cache: None,
            sequences: Arc::new(RwLock::new(HashMap::new())),
            object_store: None,
            snapshot_repo: None,
            event_repo: None,
            state_repo: None,
            wasm_runtime: None,
        }
    }

    /// Create a new state manager with custom LRU cache size
    pub fn with_lru_cache_size(max_size: usize) -> Self {
        Self {
            lru_cache: Arc::new(RwLock::new(LruCache::new(max_size))),
            redis_cache: None,
            sequences: Arc::new(RwLock::new(HashMap::new())),
            object_store: None,
            snapshot_repo: None,
            event_repo: None,
            state_repo: None,
            wasm_runtime: None,
        }
    }

    /// Create a new state manager with Redis cache (Phase 2)
    pub async fn with_redis_cache(config: RedisConfig) -> Result<Self, Box<dyn std::error::Error + Send + Sync>> {
        let redis_cache = Some(Arc::new(RedisCache::new(config).await?));

        Ok(Self {
            lru_cache: Arc::new(RwLock::new(LruCache::new(1000))),
            redis_cache,
            sequences: Arc::new(RwLock::new(HashMap::new())),
            object_store: None,
            snapshot_repo: None,
            event_repo: None,
            state_repo: None,
            wasm_runtime: None,
        })
    }

    /// Create a new state manager with storage
    pub fn with_storage(
        object_store: Box<dyn ObjectStore + Send + Sync>,
        snapshot_repo: PostgresSnapshotRepository,
        event_repo: PostgresEventRepository,
        state_repo: PostgresStateRepository,
    ) -> Self {
        Self {
            lru_cache: Arc::new(RwLock::new(LruCache::new(1000))),
            redis_cache: None,
            sequences: Arc::new(RwLock::new(HashMap::new())),
            object_store: Some(object_store),
            snapshot_repo: Some(snapshot_repo),
            event_repo: Some(event_repo),
            state_repo: Some(state_repo),
            wasm_runtime: None,
        }
    }

    /// Create a new state manager with storage and WASM runtime
    pub fn with_wasm(
        object_store: Box<dyn ObjectStore + Send + Sync>,
        snapshot_repo: PostgresSnapshotRepository,
        event_repo: PostgresEventRepository,
        state_repo: PostgresStateRepository,
        wasm_config: WasmConfig,
    ) -> StateResult<Self> {
        let wasm_runtime = Some(WasmRuntime::new(wasm_config)
            .map_err(|e| StateError::StorageError(format!("Failed to create WASM runtime: {}", e)))?);

        Ok(Self {
            lru_cache: Arc::new(RwLock::new(LruCache::new(1000))),
            redis_cache: None,
            sequences: Arc::new(RwLock::new(HashMap::new())),
            object_store: Some(object_store),
            snapshot_repo: Some(snapshot_repo),
            event_repo: Some(event_repo),
            state_repo: Some(state_repo),
            wasm_runtime,
        })
    }

    /// Get current state (with multi-layer caching)
    pub async fn get(&self, state_id: Uuid) -> StateResult<serde_json::Value> {
        // Phase 1: Check LRU cache first
        let mut lru_cache = self.lru_cache.write().await;
        if let Some(state) = lru_cache.get(&state_id) {
            return Ok(state.clone());
        }
        drop(lru_cache);

        // Phase 2: Check Redis cache if available
        if let Some(redis_cache) = &self.redis_cache {
            if let Ok(Some(cached_entry)) = redis_cache.get_state(&state_id).await {
                // Verify cache is still valid (basic TTL check)
                let now = std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .map_err(|_| StateError::StorageError("System time before epoch".to_string()))?
                    .as_secs();

                if now < (cached_entry.cached_at + cached_entry.ttl) {
                    // Update LRU cache with Redis data
                    let mut lru_cache = self.lru_cache.write().await;
                    lru_cache.put(state_id, cached_entry.data.clone());
                    return Ok(cached_entry.data);
                } else {
                    // Cache expired, remove it
                    let _ = redis_cache.delete_state(&state_id).await;
                }
            }
        }

        // Cache miss - return empty state
        Ok(serde_json::json!({}))
    }

    /// Set a value in state (with multi-layer cache updates)
    pub async fn set(
        &self,
        state_id: Uuid,
        key: String,
        value: serde_json::Value,
    ) -> StateResult<serde_json::Value> {
        // Get previous value for event
        let previous_value = self.get_key(state_id, &key).await.ok();

        // Get current full state for cache updates
        let mut current_state = self.get(state_id).await?;
        let state_obj = current_state.as_object_mut()
            .ok_or_else(|| StateError::InvalidOperation("State must be an object".to_string()))?;

        // Update the state object
        state_obj.insert(key.clone(), value.clone());

        // Phase 1: Update LRU cache
        let mut lru_cache = self.lru_cache.write().await;
        lru_cache.put(state_id, current_state.clone());

        // Phase 2: Update Redis cache if available
        if let Some(redis_cache) = &self.redis_cache {
            let version = format!("{}", std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs());
            if let Err(e) = redis_cache.set_state(&state_id, current_state.clone(), version).await {
                tracing::warn!(error = %e, state_id = %state_id, "Failed to update Redis cache");
                // Continue execution - Redis failure shouldn't block operations
            }
        }

        // Drop cache locks before committing event
        drop(lru_cache);

        // Create and commit event
        let event = Event::set(state_id, key, value.clone(), SourceType::User, "api".to_string())
            .with_previous_value(previous_value);

        self.commit_event(event).await?;

        Ok(value)
    }

    /// Get a specific key (with multi-layer caching)
    pub async fn get_key(&self, state_id: Uuid, key: &str) -> StateResult<serde_json::Value> {
        let state = self.get(state_id).await?;
        state.get(key)
            .cloned()
            .ok_or_else(|| StateError::KeyNotFound(key.to_string()))
    }

    /// Delete a key (with multi-layer cache updates)
    pub async fn delete(&self, state_id: Uuid, key: &str) -> StateResult<()> {
        // Get previous value for event
        let previous_value = self.get_key(state_id, key).await.ok();

        // Get current full state for cache updates
        let mut current_state = self.get(state_id).await?;
        if let Some(obj) = current_state.as_object_mut() {
            obj.remove(key);
        }

        // Phase 1: Update LRU cache
        let mut lru_cache = self.lru_cache.write().await;
        lru_cache.put(state_id, current_state.clone());

        // Phase 2: Update Redis cache if available
        if let Some(redis_cache) = &self.redis_cache {
            let version = format!("{}", std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs());
            if let Err(e) = redis_cache.set_state(&state_id, current_state.clone(), version).await {
                tracing::warn!(error = %e, state_id = %state_id, key = %key, "Failed to update Redis cache during delete");
            }
        }

        // Drop cache locks before committing event
        drop(lru_cache);

        // Create and commit event
        let event = Event::delete(state_id, key.to_string(), previous_value, SourceType::User, "api".to_string());
        self.commit_event(event).await?;

        Ok(())
    }

    /// Compute current state hash
    pub async fn hash(&self, state_id: Uuid) -> StateResult<String> {
        let state = self.get(state_id).await?;
        Ok(compute_state_hash(&state))
    }

    /// Get next sequence number
    pub async fn next_sequence(&self, state_id: Uuid) -> StateResult<i64> {
        let mut sequences = self.sequences.write().await;

        let seq = sequences.entry(state_id).or_insert(0);
        *seq += 1;

        Ok(*seq)
    }

    /// Load state from snapshot (for replay) - updates both caches
    pub async fn load_snapshot(&self, state_id: Uuid, snapshot: serde_json::Value) {
        // Phase 1: Update LRU cache
        let mut lru_cache = self.lru_cache.write().await;
        lru_cache.put(state_id, snapshot.clone());

        // Phase 2: Update Redis cache if available
        if let Some(redis_cache) = &self.redis_cache {
            let version = format!("snapshot_{}", std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs());
            if let Err(e) = redis_cache.set_state(&state_id, snapshot, version).await {
                tracing::warn!(error = %e, state_id = %state_id, "Failed to update Redis cache during snapshot load");
            }
        }
    }

    /// Clear state (with multi-layer cache updates)
    pub async fn clear(&self, state_id: Uuid) -> StateResult<()> {
        // Get previous state for event
        let previous_state = self.get(state_id).await.ok();

        let empty_state = serde_json::json!({});

        // Phase 1: Update LRU cache
        let mut lru_cache = self.lru_cache.write().await;
        lru_cache.put(state_id, empty_state.clone());

        // Phase 2: Update Redis cache if available
        if let Some(redis_cache) = &self.redis_cache {
            let version = format!("clear_{}", std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs());
            if let Err(e) = redis_cache.set_state(&state_id, empty_state.clone(), version).await {
                tracing::warn!(error = %e, state_id = %state_id, "Failed to update Redis cache during clear");
            }
        }

        // Drop cache locks before committing event
        drop(lru_cache);

        // Create and commit event
        let event = Event::new(state_id, EventType::Clear, SourceType::User, "api".to_string(), None)
            .with_previous_value(previous_state);

        self.commit_event(event).await?;

        Ok(())
    }

    /// Get all keys (with multi-layer caching)
    pub async fn keys(&self, state_id: Uuid) -> StateResult<Vec<String>> {
        let state = self.get(state_id).await?;
        let keys: Vec<String> = state.as_object()
            .map(|obj| obj.keys().cloned().collect())
            .unwrap_or_default();

        Ok(keys)
    }

    /// Get state size in bytes (with multi-layer caching)
    pub async fn size(&self, state_id: Uuid) -> StateResult<i64> {
        let state = self.get(state_id).await?;
        let bytes = serde_json::to_vec(&state)
            .map(|v| v.len() as i64)
            .unwrap_or(0);

        Ok(bytes)
    }

    /// Create a snapshot of the current state
    pub async fn create_snapshot(
        &self,
        state_id: Uuid,
        request: CreateSnapshotRequest,
    ) -> StateResult<Snapshot> {
        let object_store = self.object_store.as_ref()
            .ok_or_else(|| StateError::StorageError("Object store not configured".to_string()))?;
        let snapshot_repo = self.snapshot_repo.as_ref()
            .ok_or_else(|| StateError::StorageError("Snapshot repository not configured".to_string()))?;

        // Get current state
        let state_data = self.get(state_id).await?;

        // Check if we should force snapshot or if state has changed
        if !request.force.unwrap_or(false) {
            // Check if state has actually changed since last snapshot
            if let Ok(Some(latest_snapshot)) = snapshot_repo.get_latest(state_id).await {
                if let Some(previous_checksum) = &latest_snapshot.checksum {
                    let current_hash = compute_state_hash(&state_data);
                    if current_hash == *previous_checksum {
                        // State hasn't changed, don't create duplicate snapshot
                        return Err(StateError::InvalidOperation(
                            "State has not changed since last snapshot. Use force=true to create anyway.".to_string()
                        ));
                    }
                }
            }
            // If no previous snapshot or no checksum, proceed with creating snapshot
        }

        // Get next snapshot version
        let snapshot_version = snapshot_repo.get_next_version(state_id)
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to get next version: {}", e)))?;

        // Determine first sequence for this snapshot
        let first_sequence = if snapshot_version == 1 {
            // First snapshot - start from sequence 1
            1i64
        } else {
            // Subsequent snapshot - start from after the last snapshot
            let latest_snapshot = snapshot_repo.get_latest(state_id)
                .await
                .map_err(|e| StateError::StorageError(format!("Failed to get latest snapshot: {}", e)))?;

            if let Some(prev_snapshot) = latest_snapshot {
                prev_snapshot.last_sequence + 1
            } else {
                // Should not happen if version > 1, but handle gracefully
                1i64
            }
        };

        // Get current sequence number
        let last_sequence = self.next_sequence(state_id).await? - 1;

        // Get the root event ID (first event in this snapshot's range)
        let actual_root_event_id = if let Some(event_repo) = &self.event_repo {
            // Get the first event in the sequence range
            let events = event_repo.get_range(state_id, first_sequence, first_sequence)
                .await
                .map_err(|e| StateError::StorageError(format!("Failed to get root event: {}", e)))?;

            events.first().map(|e| e.id).unwrap_or_else(|| Uuid::new_v4())
        } else {
            // Fallback if no event repository available
            Uuid::new_v4()
        };

        // Compute state hash for checksum
        let state_hash = compute_state_hash(&state_data);

        // Create snapshot
        let mut snapshot = Snapshot::new(
            state_id,
            snapshot_version,
            state_data.clone(),
            first_sequence,
            last_sequence,
            actual_root_event_id,
        );

        // Set checksum
        snapshot = snapshot.with_checksum(state_hash);

        if let Some(label) = &request.label {
            snapshot = snapshot.with_label(label.clone());
        }

        // Generate storage key and store snapshot data
        let storage_key = object_store.snapshot_key(&state_id, &snapshot.id);
        let snapshot_json = serde_json::to_vec(&snapshot)
            .map_err(|e| StateError::StorageError(format!("Failed to serialize snapshot: {}", e)))?;

        object_store.put(&storage_key, &snapshot_json, Some("application/json"))
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to store snapshot: {}", e)))?;

        // Update snapshot with storage key
        snapshot = snapshot.with_storage_key(storage_key.clone());

        // Create metadata for database
        let metadata = SnapshotMetadata {
            id: snapshot.id,
            state_id,
            snapshot_version,
            label: snapshot.label.clone(),
            key_count: snapshot.key_count,
            size_bytes: snapshot.size_bytes,
            first_sequence: snapshot.first_sequence,
            last_sequence: snapshot.last_sequence,
            root_event_id: snapshot.root_event_id,
            is_compressed: snapshot.is_compressed,
            compression_algo: snapshot.compression_algo.clone(),
            checksum: snapshot.checksum.clone(),
            storage_key,
            created_at: snapshot.created_at,
        };

        // Store metadata in database
        snapshot_repo.insert_metadata(&metadata)
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to store snapshot metadata: {}", e)))?;

        // Phase 2: Cache hot snapshot in Redis for fast access
        if let Err(e) = self.cache_hot_snapshot(state_id, snapshot_version, state_data.clone()).await {
            tracing::warn!(error = %e, state_id = %state_id, snapshot_version = %snapshot_version, "Failed to cache hot snapshot");
            // Don't fail the operation for caching issues
        }

        Ok(snapshot)
    }

    /// Restore state from a snapshot (with hot snapshot caching)
    pub async fn restore_snapshot(
        &self,
        state_id: Uuid,
        request: RestoreSnapshotRequest,
    ) -> StateResult<()> {
        let object_store = self.object_store.as_ref()
            .ok_or_else(|| StateError::StorageError("Object store not configured".to_string()))?;
        let snapshot_repo = self.snapshot_repo.as_ref()
            .ok_or_else(|| StateError::StorageError("Snapshot repository not configured".to_string()))?;

        // Get snapshot metadata by version
        let metadata = snapshot_repo.get_by_version(state_id, request.snapshot_version)
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to get snapshot metadata: {}", e)))?
            .ok_or_else(|| StateError::NotFound(state_id))?;

        // Phase 2: Try to get snapshot data from Redis cache first (hot snapshot)
        let snapshot = if let Ok(Some(cached_data)) = self.get_hot_snapshot(state_id, request.snapshot_version).await {
            // Deserialize from cached data
            serde_json::from_value(cached_data)
                .map_err(|e| StateError::StorageError(format!("Failed to deserialize cached snapshot: {}", e)))?
        } else {
            // Fallback to object storage
            let snapshot_data = object_store.get(&metadata.storage_key)
                .await
                .map_err(|e| StateError::StorageError(format!("Failed to retrieve snapshot: {}", e)))?;

            let snapshot: Snapshot = serde_json::from_slice(&snapshot_data)
                .map_err(|e| StateError::StorageError(format!("Failed to deserialize snapshot: {}", e)))?;

            // Cache this snapshot for future hot access
            if let Ok(snapshot_value) = serde_json::to_value(&snapshot) {
                if let Err(e) = self.cache_hot_snapshot(state_id, request.snapshot_version, snapshot_value).await {
                    tracing::warn!(error = %e, state_id = %state_id, snapshot_version = %request.snapshot_version, "Failed to cache hot snapshot during restore");
                }
            }

            snapshot
        };

        // Load snapshot state into cache
        self.load_snapshot(state_id, snapshot.state_data).await;

        // Handle keep_subsequent flag
        if request.keep_subsequent.unwrap_or(false) {
            // Replay events that occurred after the snapshot
            if let Some(event_repo) = &self.event_repo {
                let from_seq = snapshot.last_sequence + 1;
                let to_seq = self.next_sequence(state_id).await? - 1;

                if from_seq <= to_seq {
                    // Get events in the range after the snapshot
                    let events_metadata = event_repo.get_range(state_id, from_seq, to_seq)
                        .await
                        .map_err(|e| StateError::StorageError(format!("Failed to get subsequent events: {}", e)))?;

                    // Replay each event to rebuild the state
                    for event_meta in events_metadata {
                        // Get the full event data from object storage
                        let storage_key = &event_meta.storage_key;
                        if !storage_key.is_empty() {
                            let event_data = object_store.get(storage_key)
                                .await
                                .map_err(|e| StateError::StorageError(format!("Failed to retrieve event {}: {}", event_meta.id, e)))?;

                            let event: Event = serde_json::from_slice(&event_data)
                                .map_err(|e| StateError::StorageError(format!("Failed to deserialize event {}: {}", event_meta.id, e)))?;

                            // Apply the event to the current state
                            let mut current_state = self.get(state_id).await.unwrap_or_else(|_| serde_json::json!({}));

                            match event.event_type {
                                EventType::Set => {
                                    if let (Some(key), Some(value)) = (&event.key, &event.new_value) {
                                        if let Some(obj) = current_state.as_object_mut() {
                                            obj.insert(key.clone(), value.clone());
                                        } else {
                                            // If state is not an object, initialize as object
                                            current_state = serde_json::json!({key.clone(): value.clone()});
                                        }
                                    }
                                }
                                EventType::Delete => {
                                    if let Some(key) = &event.key {
                                        if let Some(obj) = current_state.as_object_mut() {
                                            obj.remove(key);
                                        }
                                        // If state is not an object, no action needed
                                    }
                                }
                                EventType::Merge => {
                                    if let (Some(key), Some(value)) = (&event.key, &event.new_value) {
                                        if let Some(obj) = current_state.as_object_mut() {
                                            // Merge operation: deep merge the new value with existing value
                                            if let Some(existing_value) = obj.get(key) {
                                                // Perform a deep merge - implement simple object merge
                                                if existing_value.is_object() && value.is_object() {
                                                    let mut merged = existing_value.as_object()
                                                        .ok_or_else(|| StateError::InvalidOperation("Expected object value".to_string()))?
                                                        .clone();
                                                    if let Some(new_obj) = value.as_object() {
                                                        for (k, v) in new_obj {
                                                            merged.insert(k.clone(), v.clone());
                                                        }
                                                    }
                                                    obj.insert(key.clone(), serde_json::Value::Object(merged));
                                                } else {
                                                    // Not objects, just replace
                                                    obj.insert(key.clone(), value.clone());
                                                }
                                            } else {
                                                // Key doesn't exist, treat as set
                                                obj.insert(key.clone(), value.clone());
                                            }
                                        } else {
                                            // If state is not an object, treat as set
                                            current_state = serde_json::json!({key.clone(): value.clone()});
                                        }
                                    }
                                }
                                EventType::Clear => {
                                    // Clear all state - replace with empty object
                                    current_state = serde_json::json!({});
                                }
                                EventType::Snapshot | EventType::Restore => {
                                    // Metadata operations that don't affect state data
                                    // No action needed for replay
                                }
                            }

                            // Update cache with the new state
                            let mut lru_cache = self.lru_cache.write().await;
                            lru_cache.put(state_id, current_state);
                        }
                    }

                    // Update sequence to current
                    let mut sequences = self.sequences.write().await;
                    sequences.insert(state_id, to_seq + 1);
                } else {
                    // No subsequent events, just set sequence after snapshot
                    let mut sequences = self.sequences.write().await;
                    sequences.insert(state_id, snapshot.last_sequence + 1);
                }
            } else {
                // No event repository available, fall back to just setting sequence
                let mut sequences = self.sequences.write().await;
                sequences.insert(state_id, snapshot.last_sequence + 1);
            }
        } else {
            // Don't keep subsequent events, just reset sequence after snapshot
            let mut sequences = self.sequences.write().await;
            sequences.insert(state_id, snapshot.last_sequence + 1);
        }

        Ok(())
    }

    /// Get a specific snapshot by version
    pub async fn get_snapshot(&self, state_id: Uuid, version: i64) -> StateResult<Option<SnapshotMetadata>> {
        let snapshot_repo = self.snapshot_repo.as_ref()
            .ok_or_else(|| StateError::StorageError("Snapshot repository not configured".to_string()))?;

        let snapshot = snapshot_repo.get_by_version(state_id, version)
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to get snapshot: {}", e)))?;

        Ok(snapshot)
    }

    /// List snapshots for a state
    pub async fn list_snapshots(&self, state_id: Uuid) -> StateResult<Vec<SnapshotMetadata>> {
        let snapshot_repo = self.snapshot_repo.as_ref()
            .ok_or_else(|| StateError::StorageError("Snapshot repository not configured".to_string()))?;

        let snapshots = snapshot_repo.get_all(state_id)
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to get snapshots: {}", e)))?;

        Ok(snapshots)
    }

    /// Commit an event to object storage and PostgreSQL
    pub async fn commit_event(&self, mut event: Event) -> StateResult<Event> {
        let object_store = self.object_store.as_ref()
            .ok_or_else(|| StateError::StorageError("Object store not configured".to_string()))?;
        let event_repo = self.event_repo.as_ref()
            .ok_or_else(|| StateError::StorageError("Event repository not configured".to_string()))?;

        // Get next sequence number and update the event
        let sequence = self.next_sequence(event.state_id).await?;
        event.sequence_num = sequence;

        // Generate storage key for the event
        let storage_key = object_store.event_key(&event.state_id, &event.id);
        event.storage_key = Some(storage_key.clone());

        // Store event data in object storage
        let event_json = serde_json::to_vec(&event)
            .map_err(|e| StateError::StorageError(format!("Failed to serialize event: {}", e)))?;

        object_store.put(&storage_key, &event_json, Some("application/json"))
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to store event: {}", e)))?;

        // Create metadata for PostgreSQL
        let metadata = EventMetadata {
            id: event.id,
            state_id: event.state_id,
            event_type: event.event_type.as_str().to_string(),
            key: event.key.clone(),
            correlation_id: event.correlation_id.clone(),
            source_type: event.source_type.as_str().to_string(),
            source_id: event.source_id.clone(),
            deterministic: event.deterministic,
            sequence_num: event.sequence_num,
            timestamp: event.timestamp,
            storage_key,
            input_hash: event.input_hash.clone(),
            output_hash: event.output_hash.clone(),
        };

        // Store metadata in PostgreSQL
        event_repo.insert_metadata(&metadata)
            .await
            .map_err(|e| StateError::StorageError(format!("Failed to store event metadata: {}", e)))?;

        Ok(event)
    }

    /// Commit multiple events in batch (for atomic operations)
    pub async fn commit_events(&self, events: Vec<Event>) -> StateResult<Vec<Event>> {
        if events.is_empty() {
            return Ok(Vec::new());
        }

        let object_store = self.object_store.as_ref()
            .ok_or_else(|| StateError::StorageError("Object store not configured".to_string()))?;
        let event_repo = self.event_repo.as_ref()
            .ok_or_else(|| StateError::StorageError("Event repository not configured".to_string()))?;

        let state_id = events[0].state_id;
        let mut committed_events = Vec::new();

        // Verify all events belong to the same state
        for event in &events {
            if event.state_id != state_id {
                return Err(StateError::InvalidOperation(
                    "All events in batch must belong to the same state".to_string()
                ));
            }
        }

        // Get starting sequence number
        let mut current_sequence = self.next_sequence(state_id).await?;

        for mut event in events {
            // Assign sequence number
            event.sequence_num = current_sequence;
            current_sequence += 1;

            // Generate storage key
            let storage_key = object_store.event_key(&event.state_id, &event.id);
            event.storage_key = Some(storage_key.clone());

            // Store event data
            let event_json = serde_json::to_vec(&event)
                .map_err(|e| StateError::StorageError(format!("Failed to serialize event: {}", e)))?;

            object_store.put(&storage_key, &event_json, Some("application/json"))
                .await
                .map_err(|e| StateError::StorageError(format!("Failed to store event {}: {}", event.id, e)))?;

            // Create metadata
            let metadata = EventMetadata {
                id: event.id,
                state_id: event.state_id,
                event_type: event.event_type.as_str().to_string(),
                key: event.key.clone(),
                correlation_id: event.correlation_id.clone(),
                source_type: event.source_type.as_str().to_string(),
                source_id: event.source_id.clone(),
                deterministic: event.deterministic,
                sequence_num: event.sequence_num,
                timestamp: event.timestamp,
                storage_key,
                input_hash: event.input_hash.clone(),
                output_hash: event.output_hash.clone(),
            };

            // Store metadata
            event_repo.insert_metadata(&metadata)
                .await
                .map_err(|e| StateError::StorageError(format!("Failed to store event metadata {}: {}", event.id, e)))?;

            committed_events.push(event);
        }

        Ok(committed_events)
    }

    /// Merge a value into an existing key (deep merge for objects) - with multi-layer caching
    pub async fn merge(
        &self,
        state_id: Uuid,
        key: String,
        value: serde_json::Value,
    ) -> StateResult<serde_json::Value> {
        // Get previous value for event
        let previous_value = self.get_key(state_id, &key).await.ok();

        // Get current full state for cache updates
        let mut current_state = self.get(state_id).await?;
        let state_obj = current_state.as_object_mut()
            .ok_or_else(|| StateError::InvalidOperation("State must be an object".to_string()))?;

        let merged_value = if let Some(existing_value) = state_obj.get(&key) {
            // Perform deep merge
            if existing_value.is_object() && value.is_object() {
                let mut merged = existing_value.as_object().unwrap().clone();
                if let Some(new_obj) = value.as_object() {
                    for (k, v) in new_obj {
                        merged.insert(k.clone(), v.clone());
                    }
                }
                serde_json::Value::Object(merged)
            } else {
                // Not objects, just replace
                value.clone()
            }
        } else {
            // Key doesn't exist, treat as set
            value.clone()
        };

        // Update the state
        state_obj.insert(key.clone(), merged_value.clone());

        // Phase 1: Update LRU cache
        let mut lru_cache = self.lru_cache.write().await;
        lru_cache.put(state_id, current_state.clone());

        // Phase 2: Update Redis cache if available
        if let Some(redis_cache) = &self.redis_cache {
            let version = format!("merge_{}", std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs());
            if let Err(e) = redis_cache.set_state(&state_id, current_state.clone(), version).await {
                tracing::warn!(error = %e, state_id = %state_id, key = %key, "Failed to update Redis cache during merge");
            }
        }

        // Drop cache locks before committing event
        drop(lru_cache);

        // Create and commit event
        let event = Event::merge(state_id, key, merged_value.clone(), SourceType::User, "api".to_string())
            .with_previous_value(previous_value);

        self.commit_event(event).await?;

        Ok(merged_value)
    }

    /// Load a WASM module into the runtime
    pub async fn load_wasm_module(&self, name: &str, wasm_bytes: &[u8]) -> StateResult<()> {
        let runtime = self.wasm_runtime.as_ref()
            .ok_or_else(|| StateError::StorageError("WASM runtime not configured".to_string()))?;

        runtime.compile_module(name, wasm_bytes).await
            .map_err(|e| StateError::StorageError(format!("Failed to load WASM module: {}", e)))
    }

    /// Execute a WASM function (call with `Arc::clone(&state_manager).execute_wasm_function(...)`)
    pub async fn execute_wasm_function(
        self: Arc<Self>,
        module_name: &str,
        function_name: &str,
        state_id: Uuid,
        input: &[u8],
    ) -> StateResult<ExecutionResult> {
        let runtime = self.wasm_runtime.as_ref()
            .ok_or_else(|| StateError::StorageError("WASM runtime not configured".to_string()))?;

        runtime.execute_function(module_name, function_name, Arc::clone(&self), state_id, input).await
            .map_err(|e| StateError::StorageError(format!("WASM execution failed: {}", e)))
    }

    /// Check if WASM runtime is available
    pub fn has_wasm_runtime(&self) -> bool {
        self.wasm_runtime.is_some()
    }

    // ===== REDIS PHASE 2 METHODS =====

    /// Check if Redis cache is available
    pub fn has_redis_cache(&self) -> bool {
        self.redis_cache.is_some()
    }

    /// Get Redis cache statistics
    pub async fn get_redis_stats(&self) -> StateResult<Option<HashMap<String, u64>>> {
        if let Some(redis_cache) = &self.redis_cache {
            match redis_cache.get_stats().await {
                Ok(stats) => Ok(Some(stats)),
                Err(e) => Err(StateError::StorageError(format!("Redis stats error: {}", e))),
            }
        } else {
            Ok(None)
        }
    }

    /// Check rate limit for an identifier
    pub async fn check_rate_limit(&self, identifier: &str, max_requests: u32, window_seconds: u64) -> StateResult<bool> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.check_rate_limit(identifier, max_requests, window_seconds)
                .await
                .map_err(|e| StateError::StorageError(format!("Rate limit check error: {}", e)))
        } else {
            // No Redis - allow all requests (no rate limiting)
            Ok(true)
        }
    }

    /// Get rate limit status
    pub async fn get_rate_limit_status(&self, identifier: &str, window_seconds: u64) -> StateResult<Option<crate::cache::RateLimitInfo>> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.get_rate_limit_status(identifier, window_seconds)
                .await
                .map_err(|e| StateError::StorageError(format!("Rate limit status error: {}", e)))
        } else {
            Ok(None)
        }
    }

    /// Set active agent state
    pub async fn set_agent_state(&self, agent_state: crate::cache::ActiveAgentState) -> StateResult<()> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.set_agent_state(agent_state)
                .await
                .map_err(|e| StateError::StorageError(format!("Agent state set error: {}", e)))
        } else {
            Err(StateError::StorageError("Redis cache not configured for agent state".to_string()))
        }
    }

    /// Get active agent state
    pub async fn get_agent_state(&self, agent_id: &str) -> StateResult<Option<crate::cache::ActiveAgentState>> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.get_agent_state(agent_id)
                .await
                .map_err(|e| StateError::StorageError(format!("Agent state get error: {}", e)))
        } else {
            Ok(None)
        }
    }

    /// Update agent activity timestamp
    pub async fn update_agent_activity(&self, agent_id: &str) -> StateResult<()> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.update_agent_activity(agent_id)
                .await
                .map_err(|e| StateError::StorageError(format!("Agent activity update error: {}", e)))
        } else {
            Err(StateError::StorageError("Redis cache not configured for agent state".to_string()))
        }
    }

    /// Remove agent state
    pub async fn remove_agent_state(&self, agent_id: &str) -> StateResult<()> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.remove_agent_state(agent_id)
                .await
                .map_err(|e| StateError::StorageError(format!("Agent state remove error: {}", e)))
        } else {
            Err(StateError::StorageError("Redis cache not configured for agent state".to_string()))
        }
    }

    /// Get all active agents
    pub async fn get_all_active_agents(&self) -> StateResult<Vec<crate::cache::ActiveAgentState>> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.get_all_active_agents()
                .await
                .map_err(|e| StateError::StorageError(format!("Active agents list error: {}", e)))
        } else {
            Ok(Vec::new())
        }
    }

    /// Cache a hot snapshot for fast access
    pub async fn cache_hot_snapshot(&self, state_id: Uuid, version: i64, snapshot_data: serde_json::Value) -> StateResult<()> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.set_snapshot(&state_id, version, snapshot_data)
                .await
                .map_err(|e| StateError::StorageError(format!("Hot snapshot cache error: {}", e)))
        } else {
            // No Redis - silently succeed (no hot caching)
            Ok(())
        }
    }

    /// Get cached hot snapshot
    pub async fn get_hot_snapshot(&self, state_id: Uuid, version: i64) -> StateResult<Option<serde_json::Value>> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.get_snapshot(&state_id, version)
                .await
                .map_err(|e| StateError::StorageError(format!("Hot snapshot get error: {}", e)))
        } else {
            Ok(None)
        }
    }

    /// Set metadata in Redis cache
    pub async fn set_cache_metadata(&self, key: &str, value: serde_json::Value, ttl_seconds: Option<u64>) -> StateResult<()> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.set_metadata(key, value, ttl_seconds)
                .await
                .map_err(|e| StateError::StorageError(format!("Metadata set error: {}", e)))
        } else {
            Err(StateError::StorageError("Redis cache not configured for metadata".to_string()))
        }
    }

    /// Get metadata from Redis cache
    pub async fn get_cache_metadata(&self, key: &str) -> StateResult<Option<serde_json::Value>> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.get_metadata(key)
                .await
                .map_err(|e| StateError::StorageError(format!("Metadata get error: {}", e)))
        } else {
            Ok(None)
        }
    }

    /// Health check for Redis cache
    pub async fn redis_health_check(&self) -> StateResult<bool> {
        if let Some(redis_cache) = &self.redis_cache {
            redis_cache.health_check()
                .await
                .map_err(|e| StateError::StorageError(format!("Redis health check error: {}", e)))
        } else {
            Ok(false)
        }
    }

    // ===== TENANT ISOLATION METHODS =====

    /// Set a value in state with tenant isolation
    pub async fn set_with_tenant(
        &self,
        state_id: Uuid,
        key: String,
        value: serde_json::Value,
        tenant_id: Uuid,
    ) -> StateResult<serde_json::Value> {
        // Validate tenant ownership via state_repo (queries DB)
        self.validate_tenant(state_id, tenant_id).await?;
        self.set(state_id, key, value).await
    }

    /// Delete a key with tenant isolation
    pub async fn delete_with_tenant(
        &self,
        state_id: Uuid,
        key: &str,
        tenant_id: Uuid,
    ) -> StateResult<()> {
        self.validate_tenant(state_id, tenant_id).await?;
        self.delete(state_id, key).await
    }

    /// Merge a value with tenant isolation
    pub async fn merge_with_tenant(
        &self,
        state_id: Uuid,
        key: String,
        value: serde_json::Value,
        tenant_id: Uuid,
    ) -> StateResult<serde_json::Value> {
        self.validate_tenant(state_id, tenant_id).await?;
        self.merge(state_id, key, value).await
    }

    /// Clear state with tenant isolation
    pub async fn clear_with_tenant(
        &self,
        state_id: Uuid,
        tenant_id: Uuid,
    ) -> StateResult<()> {
        self.validate_tenant(state_id, tenant_id).await?;
        self.clear(state_id).await
    }

    /// Create snapshot with tenant isolation
    pub async fn create_snapshot_with_tenant(
        &self,
        state_id: Uuid,
        request: CreateSnapshotRequest,
        tenant_id: Uuid,
    ) -> StateResult<Snapshot> {
        self.validate_tenant(state_id, tenant_id).await?;
        self.create_snapshot(state_id, request).await
    }

    /// Restore snapshot with tenant isolation
    pub async fn restore_snapshot_with_tenant(
        &self,
        state_id: Uuid,
        request: RestoreSnapshotRequest,
        tenant_id: Uuid,
    ) -> StateResult<()> {
        self.validate_tenant(state_id, tenant_id).await?;
        self.restore_snapshot(state_id, request).await
    }

    /// SECURITY: Verify tenant ownership of a state
    ///
    /// Queries the database to verify that the given state_id belongs to the
    /// specified tenant. This prevents cross-tenant data access attacks.
    ///
    /// Returns Ok(true) if tenant owns the state, Ok(false) if not,
    /// or an error if verification could not be completed.
    pub async fn verify_tenant_ownership(&self, state_id: Uuid, tenant_id: Uuid) -> StateResult<bool> {
        if tenant_id == Uuid::nil() {
            tracing::warn!("verify_tenant_ownership called with nil tenant_id");
            return Ok(false);
        }

        if let Some(state_repo) = &self.state_repo {
            match state_repo.get_by_id(state_id).await {
                Ok(Some(state)) => {
                    if state.tenant_id == tenant_id {
                        Ok(true)
                    } else {
                        tracing::warn!(
                            "Tenant {} attempted to access state {} owned by tenant {}",
                            tenant_id,
                            state_id,
                            state.tenant_id
                        );
                        Ok(false)
                    }
                }
                Ok(None) => {
                    tracing::debug!("State {} not found for tenant verification", state_id);
                    Ok(false)
                }
                Err(e) => {
                    tracing::error!("Failed to verify tenant ownership via state repo: {}", e);
                    Err(StateError::StorageError(
                        "Could not verify tenant ownership".to_string(),
                    ))
                }
            }
        } else {
            tracing::warn!("No state_repo configured - cannot verify tenant ownership");
            Err(StateError::StorageError(
                "State repository not configured - tenant verification unavailable".to_string(),
            ))
        }
    }

    /// Validate that the given tenant has access to the state
    async fn validate_tenant(&self, state_id: Uuid, tenant_id: Uuid) -> StateResult<()> {
        if tenant_id == Uuid::nil() {
            return Err(StateError::InvalidOperation(
                "Invalid tenant: tenant_id cannot be nil".to_string(),
            ));
        }

        // SECURITY: Actually verify tenant ownership
        match self.verify_tenant_ownership(state_id, tenant_id).await {
            Ok(true) => Ok(()),
            Ok(false) => Err(StateError::InvalidOperation(
                "Tenant does not own this state".to_string(),
            )),
            Err(e) => Err(e),
        }
    }

}



impl Default for StateManager {
    fn default() -> Self {
        Self::new()
    }
}

/// Create an event from a state operation
pub fn create_set_event(
    state_id: Uuid,
    key: String,
    value: serde_json::Value,
    source_type: SourceType,
    source_id: String,
    sequence: i64,
) -> Event {
    Event::set(state_id, key, value, source_type, source_id)
        .with_sequence(sequence)
}

pub fn create_delete_event(
    state_id: Uuid,
    key: String,
    previous_value: Option<serde_json::Value>,
    source_type: SourceType,
    source_id: String,
    sequence: i64,
) -> Event {
    Event::delete(state_id, key, previous_value, source_type, source_id)
        .with_sequence(sequence)
}
