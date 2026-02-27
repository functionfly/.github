//! Result caching for deterministic functions.

use anyhow::Result;
use lru::LruCache;
use sha2::{Digest, Sha256};
use std::num::NonZeroUsize;
use std::time::{Duration, Instant};

/// Cache for function execution results
pub struct ResultCache {
    cache: LruCache<String, CachedResult>,
    ttl: Duration,
}

/// A cached result with metadata
pub struct CachedResult {
    pub result: String,
    pub cached_at: Instant,
}

impl ResultCache {
    /// Create a new result cache
    pub fn new(ttl_secs: u64) -> Self {
        Self {
            cache: LruCache::new(NonZeroUsize::new(1000).unwrap()), // Max 1000 entries
            ttl: Duration::from_secs(ttl_secs),
        }
    }

    /// Generate cache key from input
    pub fn generate_key(input: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(input.as_bytes());
        hex::encode(hasher.finalize())
    }

    /// Get cached result if available and not expired
    pub fn get(&mut self, input: &str) -> Option<String> {
        let key = Self::generate_key(input);

        if let Some(cached) = self.cache.get(&key) {
            // Check if expired
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Cache hit for key: {}", &key[..16]);
                return Some(cached.result.clone());
            } else {
                // Remove expired entry
                self.cache.pop(&key);
            }
        }

        None
    }

    /// Store result in cache
    pub fn set(&mut self, input: &str, result: String) {
        let key = Self::generate_key(input);

        self.cache.put(
            key,
            CachedResult {
                result,
                cached_at: Instant::now(),
            },
        );

        tracing::debug!("Cached result");
    }

    /// Check if caching is enabled
    pub fn is_enabled(&self) -> bool {
        self.ttl.as_secs() > 0
    }

    /// Clear all cached results
    pub fn clear(&mut self) {
        self.cache.clear();
    }

    /// Get cache statistics
    pub fn stats(&self) -> CacheStats {
        CacheStats {
            entries: self.cache.len(),
            ttl_secs: self.ttl.as_secs(),
        }
    }

    /// Python-specific caching methods
    /// Get cached Python WASM module by source hash
    pub fn get_python_wasm(&mut self, source_hash: &str) -> Option<Vec<u8>> {
        let key = format!("python_wasm:{}", source_hash);

        if let Some(cached) = self.cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Python WASM cache hit for hash: {}", &source_hash[..16]);
                // Parse the cached WASM bytes (stored as string for simplicity)
                if let Ok(wasm_bytes) = hex::decode(&cached.result) {
                    return Some(wasm_bytes);
                }
            } else {
                self.cache.pop(&key);
            }
        }

        None
    }

    /// Store Python WASM module in cache
    pub fn set_python_wasm(&mut self, source_hash: &str, wasm_bytes: &[u8]) {
        let key = format!("python_wasm:{}", source_hash);
        let hex_bytes = hex::encode(wasm_bytes);

        self.cache.put(
            key,
            CachedResult {
                result: hex_bytes,
                cached_at: Instant::now(),
            },
        );

        tracing::debug!("Cached Python WASM module for hash: {}", &source_hash[..16]);
    }

    /// Generate hash for Python source code
    pub fn hash_python_source(source_code: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(b"python_source:");
        hasher.update(source_code.as_bytes());
        hex::encode(hasher.finalize())
    }

    /// Package caching methods for Enterprise tier

    /// Get cached package data by package name and version
    pub fn get_package(&mut self, package_name: &str, version: &str) -> Option<Vec<u8>> {
        let key = format!("package:{}@{}", package_name, version);

        if let Some(cached) = self.cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Package cache hit for: {}@{}", package_name, version);
                // Parse the cached package data (stored as hex-encoded bytes)
                if let Ok(package_data) = hex::decode(&cached.result) {
                    return Some(package_data);
                }
            } else {
                self.cache.pop(&key);
            }
        }

        None
    }

    /// Store package data in cache
    pub fn set_package(&mut self, package_name: &str, version: &str, package_data: &[u8]) {
        let key = format!("package:{}@{}", package_name, version);
        let hex_data = hex::encode(package_data);

        self.cache.put(
            key,
            CachedResult {
                result: hex_data,
                cached_at: Instant::now(),
            },
        );

        tracing::debug!("Cached package: {}@{}", package_name, version);
    }

    /// Get cached dependency resolution result
    pub fn get_dependency_resolution(&mut self, requirements_hash: &str) -> Option<String> {
        let key = format!("deps:{}", requirements_hash);

        if let Some(cached) = self.cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Dependency resolution cache hit for hash: {}", &requirements_hash[..16]);
                return Some(cached.result.clone());
            } else {
                self.cache.pop(&key);
            }
        }

        None
    }

    /// Store dependency resolution result
    pub fn set_dependency_resolution(&mut self, requirements_hash: &str, resolution_result: String) {
        let key = format!("deps:{}", requirements_hash);

        self.cache.put(
            key,
            CachedResult {
                result: resolution_result,
                cached_at: Instant::now(),
            },
        );

        tracing::debug!("Cached dependency resolution for hash: {}", &requirements_hash[..16]);
    }

    /// Generate hash for package requirements
    pub fn hash_requirements(requirements: &[String]) -> String {
        let mut hasher = Sha256::new();
        hasher.update(b"requirements:");
        for req in requirements {
            hasher.update(req.as_bytes());
            hasher.update(b"\n");
        }
        hex::encode(hasher.finalize())
    }

    /// Get cached Micropython runtime
    pub fn get_rustpython_runtime(&mut self) -> Option<Vec<u8>> {
        let key = "rustpython_runtime".to_string();

        if let Some(cached) = self.cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Micropython runtime cache hit");
                // Parse the cached WASM bytes (stored as hex string)
                if let Ok(wasm_bytes) = hex::decode(&cached.result) {
                    return Some(wasm_bytes);
                }
            } else {
                self.cache.pop(&key);
            }
        }

        None
    }

    /// Cache the Micropython runtime
    pub fn set_rustpython_runtime(&mut self, wasm_bytes: &[u8]) {
        let key = "rustpython_runtime".to_string();
        let hex_bytes = hex::encode(wasm_bytes);

        self.cache.put(
            key,
            CachedResult {
                result: hex_bytes,
                cached_at: Instant::now(),
            },
        );

        tracing::debug!("Cached Micropython runtime ({} bytes)", wasm_bytes.len());
    }
}

/// Cache statistics
#[derive(Debug, Clone)]
pub struct CacheStats {
    pub entries: usize,
    pub ttl_secs: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cache_key_generation() {
        let key1 = ResultCache::generate_key("hello");
        let key2 = ResultCache::generate_key("hello");
        let key3 = ResultCache::generate_key("world");

        assert_eq!(key1, key2);
        assert_ne!(key1, key3);
    }

    #[test]
    fn test_cache_set_get() {
        let mut cache = ResultCache::new(3600);

        cache.set("hello", "world".to_string());

        let result = cache.get("hello");
        assert_eq!(result, Some("world".to_string()));
    }

    #[test]
    fn test_cache_miss() {
        let mut cache = ResultCache::new(3600);

        let result = cache.get("nonexistent");
        assert_eq!(result, None);
    }
}
