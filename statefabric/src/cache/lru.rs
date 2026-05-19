//! LRU Cache implementation

use std::collections::HashMap;
use std::time::Duration;

/// LRU Cache entry
#[derive(Debug)]
struct CacheEntry<T> {
    value: T,
    access_order: u64,
}

/// LRU Cache - simple in-memory cache with eviction
#[derive(Debug)]
pub struct LruCache<K, V> {
    /// Maximum number of entries
    max_size: usize,
    /// Hash map for O(1) lookup
    map: HashMap<K, CacheEntry<V>>,
    /// Current access order counter
    counter: u64,
    /// Default TTL
    default_ttl: Option<Duration>,
}

impl<K, V> LruCache<K, V>
where
    K: std::hash::Hash + Clone + Eq,
{
    /// Create a new LRU cache
    pub fn new(max_size: usize) -> Self {
        Self {
            max_size,
            map: HashMap::new(),
            counter: 0,
            default_ttl: None,
        }
    }

    /// Create a new LRU cache with default TTL
    pub fn new_with_ttl(max_size: usize, ttl: Duration) -> Self {
        Self {
            max_size,
            map: HashMap::new(),
            counter: 0,
            default_ttl: Some(ttl),
        }
    }

    /// Get a value from the cache
    pub fn get(&mut self, key: &K) -> Option<&V> {
        if let Some(entry) = self.map.get_mut(key) {
            self.counter += 1;
            entry.access_order = self.counter;
            Some(&entry.value)
        } else {
            None
        }
    }

    /// Put a value into the cache
    pub fn put(&mut self, key: K, value: V) {
        // If key exists, update it
        if let Some(entry) = self.map.get_mut(&key) {
            self.counter += 1;
            entry.access_order = self.counter;
            entry.value = value;
            return;
        }

        // Evict if at capacity
        if self.map.len() >= self.max_size {
            self.evict_lru();
        }

        // Insert new entry
        self.counter += 1;
        self.map.insert(key, CacheEntry {
            value,
            access_order: self.counter,
        });
    }

    /// Remove a value from the cache
    pub fn remove(&mut self, key: &K) -> Option<V> {
        self.map.remove(key).map(|e| e.value)
    }

    /// Check if key exists
    pub fn contains(&self, key: &K) -> bool {
        self.map.contains_key(key)
    }

    /// Get current size
    pub fn len(&self) -> usize {
        self.map.len()
    }

    /// Check if empty
    pub fn is_empty(&self) -> bool {
        self.map.is_empty()
    }

    /// Clear the cache
    pub fn clear(&mut self) {
        self.map.clear();
        self.counter = 0;
    }

    /// Evict least recently used entry
    fn evict_lru(&mut self) {
        if let Some(lru_key) = self.map
            .iter()
            .min_by_key(|(_, v)| v.access_order)
            .map(|(k, _)| k.clone())
        {
            self.map.remove(&lru_key);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_lru_cache() {
        let mut cache = LruCache::new(2);

        cache.put("a", 1);
        cache.put("b", 2);

        assert_eq!(cache.get(&"a"), Some(&1));
        assert_eq!(cache.get(&"b"), Some(&2));

        // This should evict "b" since "a" was recently accessed
        cache.put("c", 3);

        assert_eq!(cache.get(&"a"), Some(&1));
        assert_eq!(cache.get(&"b"), None);
        assert_eq!(cache.get(&"c"), Some(&3));
    }
}
