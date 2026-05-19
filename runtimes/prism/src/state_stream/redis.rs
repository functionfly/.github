//! Redis-backed state persistence for StateStream Memory Fabric
//!
//! Provides a `StatePersistence` implementation that stores state slices
//! in Redis for distributed, durable state.  Requires the `state-stream`
//! Cargo feature (`cargo build --features state-stream`).

use redis::aio::ConnectionManager;
use redis::AsyncCommands;
use tokio::sync::Mutex;
use tracing::debug;

use crate::codec::CborCodec;
use crate::core::PrismResult;
use super::store::{StateKey, StatePersistence, StateSlice};

/// Redis-backed persistence layer for the StateStream store.
///
/// State slices are stored as CBOR-encoded byte blobs keyed by
/// `prism:state:{cell_id}:{key}`.  A companion Redis SET
/// (`prism:keys:{cell_id}`) tracks all keys per cell for efficient
/// enumeration.
pub struct RedisStatePersistence {
    conn: Mutex<ConnectionManager>,
    prefix: String,
}

impl RedisStatePersistence {
    /// Create a new Redis persistence layer.
    ///
    /// `redis_url` is a standard Redis URL, e.g. `redis://localhost:6379`.
    pub async fn new(redis_url: &str) -> PrismResult<Self> {
        let client = redis::Client::open(redis_url)
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis client error: {}", e)
            ))?;

        let conn = ConnectionManager::new(client)
            .await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis connection error: {}", e)
            ))?;

        debug!(url = %redis_url, "Redis state persistence connected");
        Ok(Self {
            conn: Mutex::new(conn),
            prefix: "prism:state".to_string(),
        })
    }

    /// Create with a custom key prefix (useful for multi-tenant isolation).
    pub async fn with_prefix(redis_url: &str, prefix: &str) -> PrismResult<Self> {
        let mut s = Self::new(redis_url).await?;
        s.prefix = prefix.to_string();
        Ok(s)
    }

    /// Compute the Redis key for a state entry.
    fn entry_key(&self, key: &StateKey) -> String {
        match &key.namespace {
            Some(ns) => format!("{}:{}:{}:{}", self.prefix, key.cell_id, ns, key.key),
            None => format!("{}:{}:{}", self.prefix, key.cell_id, key.key),
        }
    }

    /// Compute the Redis SET key that tracks all keys for a cell.
    fn index_key(&self, cell_id: &str) -> String {
        format!("{}:keys:{}", self.prefix, cell_id)
    }

    // ── async implementations (called through block_in_place) ────────

    async fn save_impl(&self, key: &StateKey, slice: &StateSlice) -> PrismResult<()> {
        let redis_key = self.entry_key(key);
        let index = self.index_key(&key.cell_id);

        // Serialize the entire StateSlice to CBOR
        let encoded: Vec<u8> = CborCodec::encode(slice)
            .map_err(|e| crate::core::PrismError::Internal(
                format!("CBOR encode error: {}", e)
            ))?;

        let mut conn = self.conn.lock().await;

        // Store the slice blob
        conn.set::<_, _, ()>(&redis_key, encoded.as_slice())
            .await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis SET error: {}", e)
            ))?;

        // Add to the per-cell key index
        conn.sadd::<_, _, ()>(&index, &key.key)
            .await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis SADD error: {}", e)
            ))?;

        debug!(key = %redis_key, "State persisted to Redis");
        Ok(())
    }

    async fn load_impl(&self, key: &StateKey) -> PrismResult<Option<StateSlice>> {
        let redis_key = self.entry_key(key);
        let mut conn = self.conn.lock().await;

        let data: Option<Vec<u8>> = conn.get(&redis_key).await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis GET error: {}", e)
            ))?;

        match data {
            Some(bytes) => {
                let slice: StateSlice = CborCodec::decode(&bytes)
                    .map_err(|e| crate::core::PrismError::Internal(
                        format!("CBOR decode error: {}", e)
                    ))?;
                Ok(Some(slice))
            }
            None => Ok(None),
        }
    }

    async fn delete_impl(&self, key: &StateKey) -> PrismResult<()> {
        let redis_key = self.entry_key(key);
        let index = self.index_key(&key.cell_id);
        let mut conn = self.conn.lock().await;

        conn.del::<_, ()>(&redis_key).await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis DEL error: {}", e)
            ))?;

        conn.srem::<_, _, ()>(&index, &key.key).await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis SREM error: {}", e)
            ))?;

        debug!(key = %redis_key, "State deleted from Redis");
        Ok(())
    }

    async fn list_keys_impl(&self, cell_id: &str) -> PrismResult<Vec<StateKey>> {
        let index = self.index_key(cell_id);
        let mut conn = self.conn.lock().await;

        let members: Vec<String> = conn.smembers(&index).await
            .map_err(|e| crate::core::PrismError::Internal(
                format!("Redis SMEMBERS error: {}", e)
            ))?;

        Ok(members.into_iter().map(|k| StateKey::new(cell_id, &k)).collect())
    }
}

// ── Synchronous trait implementation ─────────────────────────────────
// The StatePersistence trait methods are synchronous, but Redis I/O is
// async.  We bridge the two with `tokio::task::block_in_place` which
// temporarily converts the current tokio worker to a blocking thread,
// allowing us to `.block_on()` the async operations without deadlocking.

impl StatePersistence for RedisStatePersistence {
    fn save(&self, key: &StateKey, slice: &StateSlice) -> PrismResult<()> {
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(self.save_impl(key, slice))
        })
    }

    fn load(&self, key: &StateKey) -> PrismResult<Option<StateSlice>> {
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(self.load_impl(key))
        })
    }

    fn delete(&self, key: &StateKey) -> PrismResult<()> {
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(self.delete_impl(key))
        })
    }

    fn list_keys(&self, cell_id: &str) -> PrismResult<Vec<StateKey>> {
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(self.list_keys_impl(cell_id))
        })
    }
}

// ── LRU Cache Layer ─────────────────────────────────────────────────
// Optional hot-cache in front of Redis.  Gates on the `state-stream`
// feature (which enables both `redis` and `lru`).

/// A read-through cache that sits in front of a `StatePersistence`
/// backend, keeping the N most recently accessed entries in an LRU
/// in-memory cache to reduce Redis round-trips.
pub struct CachedStatePersistence<P: StatePersistence> {
    inner: P,
    cache: Mutex<lru::LruCache<String, StateSlice>>,
}

impl<P: StatePersistence> CachedStatePersistence<P> {
    pub fn new(inner: P, capacity: usize) -> Self {
        Self {
            inner,
            cache: Mutex::new(lru::LruCache::new(
                std::num::NonZeroUsize::new(capacity.max(1)).unwrap(),
            )),
        }
    }

    fn cache_key(key: &StateKey) -> String {
        match &key.namespace {
            Some(ns) => format!("{}:{}:{}", key.cell_id, ns, key.key),
            None => format!("{}:{}", key.cell_id, key.key),
        }
    }
}

impl<P: StatePersistence> StatePersistence for CachedStatePersistence<P> {
    fn save(&self, key: &StateKey, slice: &StateSlice) -> PrismResult<()> {
        self.inner.save(key, slice)?;
        // Update cache
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(async {
                let mut cache = self.cache.lock().await;
                cache.put(Self::cache_key(key), slice.clone());
            })
        });
        Ok(())
    }

    fn load(&self, key: &StateKey) -> PrismResult<Option<StateSlice>> {
        // Check cache first
        let cached = tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(async {
                let mut cache = self.cache.lock().await;
                cache.get(&Self::cache_key(key)).cloned()
            })
        });
        if let Some(slice) = cached {
            return Ok(Some(slice));
        }

        // Cache miss — fetch from inner
        let result = self.inner.load(key)?;
        if let Some(ref slice) = result {
            tokio::task::block_in_place(|| {
                tokio::runtime::Handle::current().block_on(async {
                    let mut cache = self.cache.lock().await;
                    cache.put(Self::cache_key(key), slice.clone());
                })
            });
        }
        Ok(result)
    }

    fn delete(&self, key: &StateKey) -> PrismResult<()> {
        self.inner.delete(key)?;
        tokio::task::block_in_place(|| {
            tokio::runtime::Handle::current().block_on(async {
                let mut cache = self.cache.lock().await;
                cache.pop(&Self::cache_key(key));
            })
        });
        Ok(())
    }

    fn list_keys(&self, cell_id: &str) -> PrismResult<Vec<StateKey>> {
        self.inner.list_keys(cell_id)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::ValueEncoding;

    // These tests require a running Redis instance.
    // Run with: cargo test --features state-stream redis::

    #[tokio::test]
    #[ignore] // Remove @ignore when Redis is available
    async fn test_redis_save_and_load() {
        let redis_url = std::env::var("REDIS_URL")
            .unwrap_or_else(|_| "redis://localhost:6379".to_string());

        let persistence = RedisStatePersistence::new(&redis_url).await.unwrap();
        let key = StateKey::new("test-cell", "counter");
        let slice = StateSlice::new(key.clone(), b"42".to_vec(), ValueEncoding::Raw);

        persistence.save(&key, &slice).unwrap();

        let loaded = persistence.load(&key).unwrap();
        assert!(loaded.is_some());
        let loaded = loaded.unwrap();
        assert_eq!(loaded.value, b"42");

        // Cleanup
        persistence.delete(&key).unwrap();
        assert!(persistence.load(&key).unwrap().is_none());
    }

    #[tokio::test]
    #[ignore] // Remove @ignore when Redis is available
    async fn test_redis_list_keys() {
        let redis_url = std::env::var("REDIS_URL")
            .unwrap_or_else(|_| "redis://localhost:6379".to_string());

        let persistence = RedisStatePersistence::new(&redis_url).await.unwrap();

        let k1 = StateKey::new("cell-list", "a");
        let k2 = StateKey::new("cell-list", "b");
        let s1 = StateSlice::new(k1.clone(), b"1".to_vec(), ValueEncoding::Raw);
        let s2 = StateSlice::new(k2.clone(), b"2".to_vec(), ValueEncoding::Raw);

        persistence.save(&k1, &s1).unwrap();
        persistence.save(&k2, &s2).unwrap();

        let keys = persistence.list_keys("cell-list").unwrap();
        assert!(keys.len() >= 2);

        // Cleanup
        persistence.delete(&k1).unwrap();
        persistence.delete(&k2).unwrap();
    }

    #[tokio::test]
    #[ignore]
    async fn test_cached_persistence() {
        let redis_url = std::env::var("REDIS_URL")
            .unwrap_or_else(|_| "redis://localhost:6379".to_string());

        let inner = RedisStatePersistence::new(&redis_url).await.unwrap();
        let cached = CachedStatePersistence::new(inner, 100);

        let key = StateKey::new("cache-cell", "x");
        let slice = StateSlice::new(key.clone(), b"hello".to_vec(), ValueEncoding::Raw);

        cached.save(&key, &slice).unwrap();

        // First load: cache miss → Redis
        let r1 = cached.load(&key).unwrap().unwrap();
        assert_eq!(r1.value, b"hello");

        // Second load: cache hit
        let r2 = cached.load(&key).unwrap().unwrap();
        assert_eq!(r2.value, b"hello");

        // Cleanup
        cached.delete(&key).unwrap();
    }
}
