//! Key-value store for function execution.
//!
//! This module provides a simple in-memory key-value store that functions can use
//! when they declare the "kv" capability. The store supports basic get/set operations
//! with optional TTL (time-to-live) for automatic expiration.
//!
//! ## Expiry cleanup strategy
//!
//! Previously `cleanup_expired()` was called on every `get()` and `set()`, making
//! each operation O(n) in the number of entries. The store now uses a background
//! task (started via `KVStore::start_background_cleanup`) that runs every 30 seconds
//! to remove expired entries. Individual `get()` / `set()` calls still check
//! expiry for the specific key they touch, but no longer scan the entire store.

use anyhow::Result;
use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;

/// A stored key-value pair with optional expiration
#[derive(Debug, Clone)]
pub struct KVEntry {
    pub value: String,
    pub expires_at: Option<Instant>,
    pub last_accessed: Instant,
}

/// In-memory key-value store
pub struct KVStore {
    /// The actual key-value storage
    store: HashMap<String, KVEntry>,
    /// Access order for LRU eviction (most recently used at the back)
    access_order: VecDeque<String>,
    /// Maximum number of entries to prevent unbounded growth
    max_entries: usize,
}

impl KVStore {
    /// Create a new KV store with a maximum number of entries
    pub fn new(max_entries: usize) -> Self {
        Self {
            store: HashMap::new(),
            access_order: VecDeque::new(),
            max_entries,
        }
    }

    /// Get a value by key, returning None if not found or expired.
    ///
    /// Only checks expiry for the requested key — does not scan the entire store.
    /// Full expired-entry cleanup is handled by the background task started via
    /// `SharedKVStore::start_background_cleanup`.
    pub fn get(&mut self, key: &str) -> Option<String> {
        if let Some(entry) = self.store.get_mut(key) {
            // Check if expired
            if let Some(expires_at) = entry.expires_at {
                if expires_at <= Instant::now() {
                    // Remove expired entry
                    self.store.remove(key);
                    // Also remove from access order
                    self.access_order.retain(|k| k != key);
                    return None;
                }
            }

            // Update access timestamp and move to back of access order
            entry.last_accessed = Instant::now();
            self.access_order.retain(|k| k != key);
            self.access_order.push_back(key.to_string());

            Some(entry.value.clone())
        } else {
            None
        }
    }

    /// Set a value by key with optional TTL in seconds.
    ///
    /// Does not scan the entire store for expired entries on every call.
    /// Expired-entry cleanup is handled by the background task.
    pub fn set(&mut self, key: String, value: String, ttl_seconds: Option<u64>) -> Result<()> {
        // Check if we're at capacity and need to evict
        if self.store.len() >= self.max_entries && !self.store.contains_key(&key) {
            // LRU eviction: remove the least recently used entry
            while self.store.len() >= self.max_entries {
                if let Some(lru_key) = self.access_order.pop_front() {
                    if self.store.remove(&lru_key).is_some() {
                        tracing::debug!("Evicted LRU KV entry '{}' due to capacity limit", lru_key);
                        break; // Successfully evicted one entry
                    }
                    // If the key wasn't in store (already removed), continue to next
                } else {
                    // No more keys to evict, break to avoid infinite loop
                    break;
                }
            }
        }

        let expires_at = ttl_seconds.map(|ttl| Instant::now() + Duration::from_secs(ttl));
        let now = Instant::now();

        // Update access order: remove existing position and add to back (most recent)
        self.access_order.retain(|k| k != &key);
        self.access_order.push_back(key.clone());

        self.store.insert(
            key.clone(),
            KVEntry {
                value,
                expires_at,
                last_accessed: now,
            },
        );

        tracing::debug!("Set KV entry '{}' with TTL {:?}", key, ttl_seconds);
        Ok(())
    }

    /// Delete a key-value pair
    pub fn delete(&mut self, key: &str) -> bool {
        let existed = self.store.remove(key).is_some();
        if existed {
            tracing::debug!("Deleted KV entry '{}'", key);
        }
        true // Always return true for compatibility with some KV APIs
    }

    /// Check if a key exists and is not expired
    pub fn exists(&mut self, key: &str) -> bool {
        self.get(key).is_some()
    }

    /// Get all keys (for debugging/admin purposes)
    pub fn keys(&self) -> Vec<String> {
        self.store.keys().cloned().collect()
    }

    /// Get statistics about the KV store
    pub fn stats(&self) -> KVStats {
        let total_entries = self.store.len();
        let expired_count = self.store
            .values()
            .filter(|entry| {
                entry.expires_at
                    .map(|expires| expires <= Instant::now())
                    .unwrap_or(false)
            })
            .count();

        KVStats {
            total_entries,
            expired_entries: expired_count,
            max_entries: self.max_entries,
        }
    }

    /// Clean up expired entries
    fn cleanup_expired(&mut self) {
        let now = Instant::now();
        let expired_keys: Vec<String> = self.store
            .iter()
            .filter_map(|(key, entry)| {
                if let Some(expires_at) = entry.expires_at {
                    if expires_at <= now {
                        Some(key.clone())
                    } else {
                        None
                    }
                } else {
                    None
                }
            })
            .collect();

        for key in expired_keys {
            self.store.remove(&key);
            // Also remove from access order
            self.access_order.retain(|k| k != &key);
        }
    }

    /// Clear all entries
    pub fn clear(&mut self) {
        self.store.clear();
        self.access_order.clear();
        tracing::debug!("Cleared all KV entries");
    }
}

/// Thread-safe wrapper for KVStore
pub type SharedKVStore = Arc<RwLock<KVStore>>;

/// Start a background task that periodically removes expired KV entries.
///
/// This replaces the previous approach of calling `cleanup_expired()` on every
/// `get()` and `set()` (which was O(n) per operation). The background task runs
/// every 30 seconds and is the only place that scans the full store.
///
/// Returns the `JoinHandle` so the caller can abort the task on shutdown.
pub fn start_background_cleanup(store: SharedKVStore) -> tokio::task::JoinHandle<()> {
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(tokio::time::Duration::from_secs(30));
        loop {
            interval.tick().await;
            let mut kv = store.write().await;
            let before = kv.store.len();
            kv.cleanup_expired();
            let after = kv.store.len();
            if before != after {
                tracing::debug!("KV background cleanup: removed {} expired entries", before - after);
            }
        }
    })
}

/// Statistics about the KV store
#[derive(Debug, Clone)]
pub struct KVStats {
    pub total_entries: usize,
    pub expired_entries: usize,
    pub max_entries: usize,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;
    use std::time::Duration;

    #[test]
    fn test_kv_basic_operations() {
        let mut store = KVStore::new(100);

        // Test set and get
        store.set("key1".to_string(), "value1".to_string(), None).unwrap();
        assert_eq!(store.get("key1"), Some("value1".to_string()));

        // Test non-existent key
        assert_eq!(store.get("nonexistent"), None);

        // Test overwrite
        store.set("key1".to_string(), "value2".to_string(), None).unwrap();
        assert_eq!(store.get("key1"), Some("value2".to_string()));
    }

    #[test]
    fn test_kv_ttl() {
        let mut store = KVStore::new(100);

        // Set with short TTL
        store.set("ttl_key".to_string(), "ttl_value".to_string(), Some(1)).unwrap();
        assert_eq!(store.get("ttl_key"), Some("ttl_value".to_string()));

        // Wait for expiration
        thread::sleep(Duration::from_secs(2));
        assert_eq!(store.get("ttl_key"), None);
    }

    #[test]
    fn test_kv_delete() {
        let mut store = KVStore::new(100);

        store.set("delete_key".to_string(), "delete_value".to_string(), None).unwrap();
        assert_eq!(store.get("delete_key"), Some("delete_value".to_string()));

        store.delete("delete_key");
        assert_eq!(store.get("delete_key"), None);
    }

    #[test]
    fn test_kv_exists() {
        let mut store = KVStore::new(100);

        assert!(!store.exists("exists_key"));

        store.set("exists_key".to_string(), "exists_value".to_string(), None).unwrap();
        assert!(store.exists("exists_key"));

        store.delete("exists_key");
        assert!(!store.exists("exists_key"));
    }

    #[test]
    fn test_kv_capacity_limit() {
        let mut store = KVStore::new(2); // Very small capacity for testing

        // Fill up the store
        store.set("key1".to_string(), "value1".to_string(), None).unwrap();
        store.set("key2".to_string(), "value2".to_string(), None).unwrap();

        // Adding a third should evict the first (LRU eviction)
        store.set("key3".to_string(), "value3".to_string(), None).unwrap();

        // We should still have exactly max_entries entries
        // One of the original entries should be evicted
        let key1_exists = store.get("key1").is_some();
        let key2_exists = store.get("key2").is_some();
        let key3_exists = store.get("key3").is_some();

        assert!(key3_exists, "Newly added key3 should exist");
        // Exactly one of key1 or key2 should be evicted (LRU behavior)
        assert!((key1_exists && !key2_exists) || (!key1_exists && key2_exists),
                "Exactly one of the original keys should be evicted");
    }

    #[test]
    fn test_kv_stats() {
        let mut store = KVStore::new(100);

        store.set("key1".to_string(), "value1".to_string(), None).unwrap();
        store.set("key2".to_string(), "value2".to_string(), Some(1)).unwrap(); // Will expire quickly

        let stats = store.stats();
        assert_eq!(stats.total_entries, 2);
        assert_eq!(stats.max_entries, 100);

        // Wait for expiration and check again
        thread::sleep(Duration::from_secs(2));
        let stats_after = store.stats();
        assert_eq!(stats_after.expired_entries, 1); // One expired entry
    }

    #[test]
    fn test_kv_clear() {
        let mut store = KVStore::new(100);

        store.set("key1".to_string(), "value1".to_string(), None).unwrap();
        store.set("key2".to_string(), "value2".to_string(), None).unwrap();

        assert_eq!(store.stats().total_entries, 2);

        store.clear();
        assert_eq!(store.stats().total_entries, 0);
    }
}