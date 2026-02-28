//! AOT (Ahead-of-Time) compiled module cache.
//!
//! Wasmtime compiles WebAssembly modules to native code on first use.  Without
//! caching this compilation happens on every process restart, adding tens to
//! hundreds of milliseconds of cold-start latency.
//!
//! This module persists compiled modules to disk using Wasmtime's
//! `Module::serialize()` / `unsafe Module::deserialize_file()` API so that
//! subsequent loads skip JIT compilation entirely.
//!
//! # Phase 1 implementation
//!
//! Addresses the gap identified in `plans/SANDBOX_EXECUTION_LAYER.md`:
//! > Module compilation is not AOT-cached to disk — **High** (perf)
//! > Use Wasmtime's `Module::serialize()` / `Module::deserialize_file()` to
//! > persist compiled modules to disk; eliminates JIT cost on restart.
//!
//! # Security note
//!
//! Deserializing a compiled module is **unsafe** because a maliciously crafted
//! `.cwasm` file could exploit the Wasmtime runtime.  The cache directory
//! **must** be writable only by the `functionfly-local` process.  The file
//! name is derived from a SHA-256 hash of the raw WASM bytes so that a
//! modified WASM binary always produces a cache miss.

use std::path::{Path, PathBuf};
use std::sync::Arc;
use tokio::sync::RwLock;
use sha2::{Digest, Sha256};
use wasmtime::{Engine, Module};

/// In-memory + on-disk AOT module cache.
///
/// Thread-safe: wrap in `Arc<ModuleCache>` and share across request handlers.
pub struct ModuleCache {
    /// Directory where `.cwasm` files are stored.
    cache_dir: PathBuf,
    /// In-memory index: wasm_hash → compiled module.
    ///
    /// Keeping deserialized modules in memory avoids repeated disk I/O for
    /// hot functions.
    memory_cache: Arc<RwLock<lru::LruCache<String, Module>>>,
    /// Whether the cache is enabled.
    enabled: bool,
}

impl ModuleCache {
    /// Create a new module cache backed by `cache_dir`.
    ///
    /// `memory_capacity` controls how many compiled modules are kept in RAM.
    /// A value of 64 is a reasonable default for most deployments.
    pub fn new(cache_dir: impl Into<PathBuf>, memory_capacity: usize, enabled: bool) -> Self {
        let cache_dir = cache_dir.into();
        if enabled {
            if let Err(e) = std::fs::create_dir_all(&cache_dir) {
                tracing::warn!("ModuleCache: could not create cache dir {:?}: {}", cache_dir, e);
            }
        }
        Self {
            cache_dir,
            memory_cache: Arc::new(RwLock::new(lru::LruCache::new(
                std::num::NonZeroUsize::new(memory_capacity.max(1)).unwrap(),
            ))),
            enabled,
        }
    }

    /// Compute the cache key (hex-encoded SHA-256 of the raw WASM bytes).
    pub fn cache_key(wasm_bytes: &[u8]) -> String {
        let mut hasher = Sha256::new();
        hasher.update(wasm_bytes);
        hex::encode(hasher.finalize())
    }

    /// Return the on-disk path for a given cache key.
    fn disk_path(&self, key: &str) -> PathBuf {
        self.cache_dir.join(format!("{}.cwasm", key))
    }

    /// Try to load a compiled module from cache (memory first, then disk).
    ///
    /// Returns `None` on a cache miss or if the cache is disabled.
    ///
    /// # Safety
    ///
    /// Deserializing from disk is `unsafe` per Wasmtime's API contract.  We
    /// accept this risk because the cache directory is controlled by the
    /// runtime process.
    pub async fn get(&self, engine: &Engine, wasm_bytes: &[u8]) -> Option<Module> {
        if !self.enabled {
            return None;
        }

        let key = Self::cache_key(wasm_bytes);

        // 1. Check in-memory LRU cache first (fast path)
        {
            let mut mem = self.memory_cache.write().await;
            if let Some(module) = mem.get(&key) {
                tracing::debug!("ModuleCache: memory hit for key {}", &key[..8]);
                return Some(module.clone());
            }
        }

        // 2. Check on-disk cache
        let path = self.disk_path(&key);
        if path.exists() {
            match unsafe { Module::deserialize_file(engine, &path) } {
                Ok(module) => {
                    tracing::debug!("ModuleCache: disk hit for key {}", &key[..8]);
                    // Promote to memory cache
                    let mut mem = self.memory_cache.write().await;
                    mem.put(key, module.clone());
                    return Some(module);
                }
                Err(e) => {
                    // Corrupted or incompatible cache file — remove it so the
                    // next request recompiles and writes a fresh entry.
                    tracing::warn!(
                        "ModuleCache: failed to deserialize {:?}: {} — removing stale entry",
                        path,
                        e
                    );
                    let _ = std::fs::remove_file(&path);
                }
            }
        }

        None
    }

    /// Compile `wasm_bytes` with `engine`, store the result in cache, and
    /// return the compiled `Module`.
    pub async fn get_or_compile(&self, engine: &Engine, wasm_bytes: &[u8]) -> anyhow::Result<Module> {
        // Try cache first
        if let Some(module) = self.get(engine, wasm_bytes).await {
            return Ok(module);
        }

        // Cache miss — compile
        tracing::debug!("ModuleCache: compiling module ({} bytes)", wasm_bytes.len());
        let module = Module::new(engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!("Failed to compile Wasm module: {}", e))?;

        // Persist to cache
        self.put(engine, wasm_bytes, &module).await;

        Ok(module)
    }

    /// Store a compiled module in the cache (memory + disk).
    pub async fn put(&self, _engine: &Engine, wasm_bytes: &[u8], module: &Module) {
        if !self.enabled {
            return;
        }

        let key = Self::cache_key(wasm_bytes);

        // Write to memory cache
        {
            let mut mem = self.memory_cache.write().await;
            mem.put(key.clone(), module.clone());
        }

        // Write to disk cache (best-effort; failures are logged but not fatal)
        let path = self.disk_path(&key);
        match module.serialize() {
            Ok(bytes) => {
                if let Err(e) = std::fs::write(&path, &bytes) {
                    tracing::warn!("ModuleCache: failed to write {:?}: {}", path, e);
                } else {
                    tracing::debug!(
                        "ModuleCache: wrote {} bytes to {:?}",
                        bytes.len(),
                        path
                    );
                }
            }
            Err(e) => {
                tracing::warn!("ModuleCache: failed to serialize module: {}", e);
            }
        }
    }

    /// Return cache statistics.
    pub async fn stats(&self) -> ModuleCacheStats {
        let mem = self.memory_cache.read().await;
        let disk_entries = if self.enabled {
            std::fs::read_dir(&self.cache_dir)
                .map(|rd| rd.filter_map(|e| e.ok()).count())
                .unwrap_or(0)
        } else {
            0
        };

        ModuleCacheStats {
            memory_entries: mem.len(),
            disk_entries,
            cache_dir: self.cache_dir.clone(),
            enabled: self.enabled,
        }
    }

    /// Evict all entries from the in-memory cache (disk entries are kept).
    pub async fn clear_memory(&self) {
        let mut mem = self.memory_cache.write().await;
        mem.clear();
    }

    /// Remove all on-disk cache files.  Use with caution.
    pub fn clear_disk(&self) {
        if let Ok(rd) = std::fs::read_dir(&self.cache_dir) {
            for entry in rd.flatten() {
                let path = entry.path();
                if path.extension().map(|e| e == "cwasm").unwrap_or(false) {
                    let _ = std::fs::remove_file(&path);
                }
            }
        }
    }
}

/// Statistics about the module cache.
#[derive(Debug, Clone)]
pub struct ModuleCacheStats {
    pub memory_entries: usize,
    pub disk_entries: usize,
    pub cache_dir: PathBuf,
    pub enabled: bool,
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;

    fn make_engine() -> Engine {
        let mut config = wasmtime::Config::new();
        config.consume_fuel(true);
        Engine::new(&config).unwrap()
    }

    fn minimal_wasm() -> Vec<u8> {
        wat::parse_str("(module (func (export \"main\")))").unwrap()
    }

    #[tokio::test]
    async fn test_cache_miss_then_hit() {
        let dir = TempDir::new().unwrap();
        let cache = ModuleCache::new(dir.path(), 8, true);
        let engine = make_engine();
        let wasm = minimal_wasm();

        // First call: cache miss, compiles
        let m1 = cache.get_or_compile(&engine, &wasm).await.unwrap();

        // Second call: should hit memory cache
        let m2 = cache.get_or_compile(&engine, &wasm).await.unwrap();

        // Both modules should be valid (we can't compare Module equality directly)
        let stats = cache.stats().await;
        assert_eq!(stats.memory_entries, 1);
        assert_eq!(stats.disk_entries, 1);
        drop(m1);
        drop(m2);
    }

    #[tokio::test]
    async fn test_cache_disabled() {
        let dir = TempDir::new().unwrap();
        let cache = ModuleCache::new(dir.path(), 8, false);
        let engine = make_engine();
        let wasm = minimal_wasm();

        let _m = cache.get_or_compile(&engine, &wasm).await.unwrap();

        let stats = cache.stats().await;
        assert_eq!(stats.memory_entries, 0);
        assert_eq!(stats.disk_entries, 0);
    }

    #[test]
    fn test_cache_key_deterministic() {
        let wasm = minimal_wasm();
        let k1 = ModuleCache::cache_key(&wasm);
        let k2 = ModuleCache::cache_key(&wasm);
        assert_eq!(k1, k2);
    }

    #[test]
    fn test_cache_key_differs_for_different_wasm() {
        let w1 = wat::parse_str("(module (func (export \"a\")))").unwrap();
        let w2 = wat::parse_str("(module (func (export \"b\")))").unwrap();
        assert_ne!(ModuleCache::cache_key(&w1), ModuleCache::cache_key(&w2));
    }
}
