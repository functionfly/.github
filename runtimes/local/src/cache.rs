//! Result caching for deterministic functions.
//!
//! This module provides two separate caches to avoid the memory overhead of
//! hex-encoding binary data (WASM bytes, packages) into the string-keyed LRU:
//!
//! - `ResultCache` — string results from function executions (LRU, 1000 entries)
//! - `BinaryCache`  — raw binary blobs (WASM modules, packages) stored as `Vec<u8>`
//!   without hex-encoding (LRU, 256 entries)

use lru::LruCache;
use sha2::{Digest, Sha256};
use std::num::NonZeroUsize;
use std::time::{Duration, Instant};

/// Cache for function execution results (string output)
pub struct ResultCache {
    cache: LruCache<String, CachedResult>,
    /// Separate cache for binary blobs (WASM bytes, packages).
    /// Stored as raw `Vec<u8>` to avoid the 2× memory overhead of hex-encoding.
    binary_cache: LruCache<String, CachedBinary>,
    ttl: Duration,
}

/// A cached string result with metadata
pub struct CachedResult {
    pub result: String,
    pub cached_at: Instant,
}

/// A cached binary blob with metadata
#[allow(dead_code)]
pub struct CachedBinary {
    pub data: Vec<u8>,
    pub cached_at: Instant,
}

impl ResultCache {
    /// Create a new result cache.
    ///
    /// String results are stored in a 1000-entry LRU.
    /// Binary blobs (WASM bytes, packages) are stored in a separate 256-entry
    /// LRU as raw `Vec<u8>` — no hex-encoding — to halve memory usage.
    pub fn new(ttl_secs: u64) -> Self {
        Self {
            cache: LruCache::new(NonZeroUsize::new(1000).unwrap()),
            binary_cache: LruCache::new(NonZeroUsize::new(256).unwrap()),
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

    /// Clear all cached results (both string and binary caches)
    pub fn clear(&mut self) {
        self.cache.clear();
        self.binary_cache.clear();
    }

    /// Get cache statistics
    pub fn stats(&self) -> CacheStats {
        CacheStats {
            entries: self.cache.len(),
            binary_entries: self.binary_cache.len(),
            ttl_secs: self.ttl.as_secs(),
        }
    }

    /// Python-specific caching methods.
    ///
    /// Binary data (WASM bytes) is stored in the dedicated `binary_cache` as
    /// raw `Vec<u8>` — no hex-encoding — to avoid the 2× memory overhead of
    /// the previous implementation.
    #[allow(dead_code)]
    /// Get cached Python WASM module by source hash
    pub fn get_python_wasm(&mut self, source_hash: &str) -> Option<Vec<u8>> {
        let key = format!("python_wasm:{}", source_hash);
        if let Some(cached) = self.binary_cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Python WASM cache hit for hash: {}", &source_hash[..16]);
                return Some(cached.data.clone());
            } else {
                self.binary_cache.pop(&key);
            }
        }
        None
    }

    /// Store Python WASM module in cache
    pub fn set_python_wasm(&mut self, source_hash: &str, wasm_bytes: &[u8]) {
        let key = format!("python_wasm:{}", source_hash);
        self.binary_cache.put(
            key,
            CachedBinary {
                data: wasm_bytes.to_vec(),
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

    /// Package caching methods for Enterprise tier.
    ///
    /// Package data is stored in the dedicated `binary_cache` as raw `Vec<u8>`.
    /// Get cached package data by package name and version
    pub fn get_package(&mut self, package_name: &str, version: &str) -> Option<Vec<u8>> {
        let key = format!("package:{}@{}", package_name, version);
        if let Some(cached) = self.binary_cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("Package cache hit for: {}@{}", package_name, version);
                return Some(cached.data.clone());
            } else {
                self.binary_cache.pop(&key);
            }
        }
        None
    }

    /// Store package data in cache
    pub fn set_package(&mut self, package_name: &str, version: &str, package_data: &[u8]) {
        let key = format!("package:{}@{}", package_name, version);
        self.binary_cache.put(
            key,
            CachedBinary {
                data: package_data.to_vec(),
                cached_at: Instant::now(),
            },
        );
        tracing::debug!("Cached package: {}@{}", package_name, version);
    }

    /// Get cached dependency resolution result (string — stays in string cache)
    pub fn get_dependency_resolution(&mut self, requirements_hash: &str) -> Option<String> {
        let key = format!("deps:{}", requirements_hash);
        if let Some(cached) = self.cache.get(&key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!(
                    "Dependency resolution cache hit for hash: {}",
                    &requirements_hash[..16]
                );
                return Some(cached.result.clone());
            } else {
                self.cache.pop(&key);
            }
        }
        None
    }

    /// Store dependency resolution result
    pub fn set_dependency_resolution(
        &mut self,
        requirements_hash: &str,
        resolution_result: String,
    ) {
        let key = format!("deps:{}", requirements_hash);
        self.cache.put(
            key,
            CachedResult {
                result: resolution_result,
                cached_at: Instant::now(),
            },
        );
        tracing::debug!(
            "Cached dependency resolution for hash: {}",
            &requirements_hash[..16]
        );
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

    /// Get cached RustPython/Micropython runtime bytes
    pub fn get_rustpython_runtime(&mut self) -> Option<Vec<u8>> {
        let key = "rustpython_runtime";
        if let Some(cached) = self.binary_cache.get(key) {
            if cached.cached_at.elapsed() < self.ttl {
                tracing::debug!("RustPython runtime cache hit");
                return Some(cached.data.clone());
            } else {
                self.binary_cache.pop(key);
            }
        }
        None
    }

    /// Cache the RustPython/Micropython runtime bytes
    pub fn set_rustpython_runtime(&mut self, wasm_bytes: &[u8]) {
        let key = "rustpython_runtime".to_string();
        self.binary_cache.put(
            key,
            CachedBinary {
                data: wasm_bytes.to_vec(),
                cached_at: Instant::now(),
            },
        );
        tracing::debug!("Cached RustPython runtime ({} bytes)", wasm_bytes.len());
    }
}

/// Cache statistics
#[derive(Debug, Clone)]
pub struct CacheStats {
    pub entries: usize,
    /// Number of entries in the binary blob cache
    pub binary_entries: usize,
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
