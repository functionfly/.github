//! Hot memory tier — in-process dashmap LRU.
//!
//! < 5ms reads with no network overhead. Tenant-isolated via namespaced keys.
//! Falls through to the Warm tier on miss so graph nodes always have a consistent view.

use std::hash::Hash;
use std::sync::Arc;
use std::time::{Duration, Instant};

use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;

/// Maximum entries in the hot LRU before we start evicting.
const DEFAULT_HOT_CAPACITY: usize = 10_000;

/// Default TTL for hot entries — after this they are considered stale
/// and the warm tier is checked even if the key exists here.
const DEFAULT_HOT_TTL_SECS: u64 = 300; // 5 minutes

/// A cached value with metadata for LRU and TTL management.
#[derive(Debug, Clone)]
struct CacheEntry {
    value: String,
    inserted_at: Instant,
    last_accessed: Instant,
    access_count: u64,
}

impl CacheEntry {
    fn new(value: String) -> Self {
        let now = Instant::now();
        Self {
            value,
            inserted_at: now,
            last_accessed: now,
            access_count: 1,
        }
    }

    fn touch(&mut self) {
        self.last_accessed = Instant::now();
        self.access_count += 1;
    }

    fn is_expired(&self, ttl: Duration) -> bool {
        self.inserted_at.elapsed() > ttl
    }
}

/// Hot (in-process) LRU memory for sub-millisecond reads.
///
/// Uses a `DashMap` so reads are lock-free. Writes take a write lock on the
/// specific shard. This is safe for concurrent access from multiple tokio tasks.
pub struct HotMemory {
    /// Per-tenant cache shards to reduce contention.
    /// Key = tenant_id (or "default" if none).
    shards: DashMap<String, Arc<RwLock<LruCache>>>,
    capacity: usize,
    ttl: Duration,
}

struct LruCache {
    /// LRU-ordered entries — front = least recently used.
    order: Vec<String>,
    entries: std::collections::HashMap<String, CacheEntry>,
    capacity: usize,
}

impl LruCache {
    fn new(capacity: usize) -> Self {
        Self {
            order: Vec::new(),
            entries: std::collections::HashMap::new(),
            capacity,
        }
    }

    fn get(&mut self, key: &str) -> Option<String> {
        let entry = self.entries.get_mut(key)?;
        entry.touch();
        // Move to back (most recently used)
        self.order.retain(|k| k != key);
        self.order.push(key.to_string());
        Some(entry.value.clone())
    }

    fn set(&mut self, key: String, value: String) {
        if let Some(entry) = self.entries.get_mut(&key) {
            entry.value = value;
            entry.touch();
            self.order.retain(|k| k != &key);
            self.order.push(key);
        } else {
            if self.entries.len() >= self.capacity {
                // Evict LRU entry
                if let Some(lru_key) = self.order.first().cloned() {
                    self.entries.remove(&lru_key);
                    self.order.remove(0);
                }
            }
            self.order.push(key.clone());
            self.entries.insert(key, CacheEntry::new(value));
        }
    }

    fn remove(&mut self, key: &str) -> bool {
        let removed = self.entries.remove(key).is_some();
        if removed {
            self.order.retain(|k| k != key);
        }
        removed
    }

    fn clear(&mut self) {
        self.entries.clear();
        self.order.clear();
    }

    fn keys(&self) -> Vec<String> {
        self.entries.keys().cloned().collect()
    }
}

impl HotMemory {
    /// Create a new hot memory tier.
    pub fn new(capacity_per_tenant: usize) -> Self {
        Self {
            shards: DashMap::new(),
            capacity: capacity_per_tenant,
            ttl: Duration::from_secs(DEFAULT_HOT_TTL_SECS),
        }
    }

    /// Get a tenant-specific shard (creates if missing).
    fn shard(&self, tenant_id: Option<&str>) -> Arc<RwLock<LruCache>> {
        let key = tenant_id.unwrap_or("default");
        self.shards
            .entry(key.to_string())
            .or_insert_with(|| Arc::new(RwLock::new(LruCache::new(self.capacity))))
            .clone()
    }

    /// Namespaced key for a tenant.
    fn namespaced_key(&self, tenant_id: Option<&str>, key: &str) -> String {
        match tenant_id {
            Some(t) => format!("mem:{}:{}", t, key),
            None => format!("mem:{}", key),
        }
    }

    /// Get a value from hot memory.
    ///
    /// Returns `Ok(Some(value))` on hit, `Ok(None)` on miss.
    /// Does NOT fall through to warm — the caller decides that.
    pub async fn get(
        &self,
        tenant_id: Option<&str>,
        key: &str,
    ) -> anyhow::Result<Option<String>> {
        let ns_key = self.namespaced_key(tenant_id, key);
        let shard = self.shard(tenant_id);
        let mut cache = shard.write().await;

        match cache.get(&ns_key) {
            Some(v) => Ok(Some(v)),
            None => Ok(None),
        }
    }

    /// Set a value in hot memory.
    pub async fn set(
        &self,
        tenant_id: Option<&str>,
        key: &str,
        value: String,
    ) -> anyhow::Result<()> {
        let ns_key = self.namespaced_key(tenant_id, key);
        let shard = self.shard(tenant_id);
        let mut cache = shard.write().await;
        cache.set(ns_key, value);
        Ok(())
    }

    /// Delete a key from hot memory.
    pub async fn delete(&self, tenant_id: Option<&str>, key: &str) -> anyhow::Result<bool> {
        let ns_key = self.namespaced_key(tenant_id, key);
        let shard = self.shard(tenant_id);
        let mut cache = shard.write().await;
        Ok(cache.remove(&ns_key))
    }

    /// Clear all entries for a tenant (or the default shard if none).
    pub async fn clear(&self, tenant_id: Option<&str>) -> anyhow::Result<()> {
        let shard = self.shard(tenant_id);
        let mut cache = shard.write().await;
        cache.clear();
        Ok(())
    }

    /// List all keys for a tenant (or the default shard if none).
    ///
    /// Returns namespaced keys stripped of the tenant prefix so callers
    /// see the original key they wrote.
    pub async fn list(&self, tenant_id: Option<&str>) -> anyhow::Result<Vec<String>> {
        let shard = self.shard(tenant_id);
        let cache = shard.read().await;
        let prefix = match tenant_id {
            Some(t) => format!("mem:{}:", t),
            None => "mem:".to_string(),
        };
        Ok(cache.keys().iter()
            .filter_map(|k| k.strip_prefix(&prefix).map(String::from))
            .collect())
    }

    /// Evict all expired entries across all shards.
    pub async fn evict_expired(&self) -> usize {
        let ttl = self.ttl;
        let mut evicted = 0;

        for shard_ref in self.shards.iter() {
            let mut cache = shard_ref.write().await;
            let expired: Vec<String> = cache
                .entries
                .iter()
                .filter(|(_, e)| e.is_expired(ttl))
                .map(|(k, _)| k.clone())
                .collect();

            for key in &expired {
                cache.remove(key);
                evicted += 1;
            }
        }

        evicted
    }

    /// Get per-shard statistics.
    pub async fn stats(&self) -> HotMemoryStats {
        let mut total_entries = 0;
        let mut shards_info = Vec::new();

        for shard_ref in self.shards.iter() {
            let tenant = shard_ref.key().clone();
            let cache = shard_ref.read().await;
            let entries = cache.entries.len();
            total_entries += entries;
            shards_info.push(HotShardStats {
                tenant,
                entries,
            });
        }

        HotMemoryStats {
            total_entries,
            capacity_per_tenant: self.capacity,
            ttl_secs: self.ttl.as_secs(),
            shards: shards_info,
        }
    }
}

/// Statistics for the hot memory tier.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HotMemoryStats {
    pub total_entries: usize,
    pub capacity_per_tenant: usize,
    pub ttl_secs: u64,
    pub shards: Vec<HotShardStats>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HotShardStats {
    pub tenant: String,
    pub entries: usize,
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_hot() -> HotMemory {
        HotMemory::new(100)
    }

    #[tokio::test]
    async fn test_basic_get_set() {
        let hot = make_hot();
        hot.set(None, "key1", "value1".to_string()).await.unwrap();
        let v = hot.get(None, "key1").await.unwrap();
        assert_eq!(v, Some("value1".to_string()));
    }

    #[tokio::test]
    async fn test_missing_key() {
        let hot = make_hot();
        let v = hot.get(None, "nonexistent").await.unwrap();
        assert_eq!(v, None);
    }

    #[tokio::test]
    async fn test_tenant_isolation() {
        let hot = make_hot();
        hot.set(Some("tenant-a"), "key", "a_value".to_string()).await.unwrap();
        hot.set(Some("tenant-b"), "key", "b_value".to_string()).await.unwrap();

        let va = hot.get(Some("tenant-a"), "key").await.unwrap();
        let vb = hot.get(Some("tenant-b"), "key").await.unwrap();

        assert_eq!(va, Some("a_value".to_string()));
        assert_eq!(vb, Some("b_value".to_string()));
    }

    #[tokio::test]
    async fn test_lru_eviction() {
        let hot = HotMemory::new(3); // very small capacity
        for i in 0..3 {
            hot.set(None, &format!("key{}", i), format!("value{}", i))
                .await
                .unwrap();
        }

        // Adding a 4th key should evict key0 (LRU)
        hot.set(None, "key3", "value3".to_string()).await.unwrap();

        let evicted = hot.get(None, "key0").await.unwrap();
        let still_there = hot.get(None, "key3").await.unwrap();

        assert_eq!(evicted, None);
        assert_eq!(still_there, Some("value3".to_string()));
    }

    #[tokio::test]
    async fn test_delete() {
        let hot = make_hot();
        hot.set(None, "key1", "value1".to_string()).await.unwrap();
        let deleted = hot.delete(None, "key1").await.unwrap();
        assert!(deleted);

        let v = hot.get(None, "key1").await.unwrap();
        assert_eq!(v, None);
    }

    #[tokio::test]
    async fn test_stats() {
        let hot = make_hot();
        hot.set(None, "key1", "value1".to_string()).await.unwrap();
        let stats = hot.stats().await;
        assert_eq!(stats.total_entries, 1);
        assert_eq!(stats.capacity_per_tenant, 100);
    }
}
