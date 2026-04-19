//! AOT (Ahead-of-Time) compilation cache for WASM modules.

use std::collections::HashMap;
use std::sync::Arc;
use std::sync::atomic::{AtomicU64, Ordering};

use wasmtime::Module;

/// In-memory AOT compilation cache entry.
pub struct AotCacheEntry {
    /// Serialized compiled module bytes.
    pub compiled: Vec<u8>,
    /// Approximate size in bytes (for eviction accounting).
    pub size: usize,
    /// Insertion order counter (used for LRU eviction).
    pub inserted_at: u64,
}

/// AOT compilation cache for WASM modules.
pub struct AotCache {
    /// AOT compilation cache: wasm_hash → compiled bytes.
    pub cache: Arc<std::sync::RwLock<HashMap<String, AotCacheEntry>>>,
    /// Monotonic counter for LRU eviction ordering.
    pub counter: Arc<AtomicU64>,
}

impl AotCache {
    /// Create a new empty AOT cache.
    pub fn new() -> Self {
        Self {
            cache: Arc::new(std::sync::RwLock::new(HashMap::new())),
            counter: Arc::new(AtomicU64::new(0)),
        }
    }

    /// Compile a Wasm binary and store the result in the AOT cache.
    ///
    /// Returns the serialized compiled bytes so the caller can persist them to
    /// the registry database if desired.
    pub fn compile_and_cache(
        &self,
        engine: &wasmtime::Engine,
        wasm_bytes: &[u8],
        hash: &str,
        config: &crate::config::Config,
    ) -> anyhow::Result<Vec<u8>> {
        // Compile the module
        let module = Module::new(engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!("AOT: failed to compile Wasm module: {}", e))?;

        // Serialize to portable compiled bytes
        let compiled = module.serialize()
            .map_err(|e| anyhow::anyhow!("AOT: failed to serialize compiled module: {}", e))?;

        if config.aot_cache_enabled {
            let size = compiled.len();
            let counter = self.counter.fetch_add(1, Ordering::Relaxed);
            let entry = AotCacheEntry { compiled: compiled.clone(), size, inserted_at: counter };

            let mut cache = self.cache.write()
                .map_err(|_| anyhow::anyhow!("AOT cache lock poisoned"))?;

            // Evict oldest entries if we exceed the size budget
            let max_bytes = config.aot_cache_size_mb * 1024 * 1024;
            let current_bytes: usize = cache.values().map(|e| e.size).sum();
            if current_bytes + size > max_bytes {
                // Find and remove the oldest entry
                if let Some(oldest_key) = cache.iter()
                    .min_by_key(|(_, e)| e.inserted_at)
                    .map(|(k, _)| k.clone())
                {
                    cache.remove(&oldest_key);
                    tracing::debug!("AOT cache: evicted entry {}", &oldest_key[..8.min(oldest_key.len())]);
                }
            }

            cache.insert(hash.to_string(), entry);
            tracing::debug!("AOT cache: stored compiled module for hash {}", &hash[..8.min(hash.len())]);

            // Optionally persist to disk
            if !config.aot_cache_dir.is_empty() {
                let dir = std::path::Path::new(&config.aot_cache_dir);
                if let Err(e) = std::fs::create_dir_all(dir) {
                    tracing::warn!("AOT cache: failed to create cache dir: {}", e);
                } else {
                    let path = dir.join(format!("{}.cwasm", hash));
                    if let Err(e) = std::fs::write(&path, &compiled) {
                        tracing::warn!("AOT cache: failed to write {}: {}", path.display(), e);
                    }
                }
            }
        }

        Ok(compiled)
    }

    /// Load a precompiled module from the AOT cache (memory or disk).
    ///
    /// Returns `None` if the hash is not cached.
    ///
    /// # Safety
    /// The compiled bytes must have been produced by `compile_and_cache` on
    /// the same Wasmtime engine configuration.  Bytes from untrusted sources
    /// must NOT be passed here.
    pub fn load_precompiled(
        &self,
        engine: &wasmtime::Engine,
        hash: &str,
        config: &crate::config::Config,
    ) -> anyhow::Result<Option<Module>> {
        // 1. Check in-memory cache
        if config.aot_cache_enabled {
            if let Ok(cache) = self.cache.read() {
                if let Some(entry) = cache.get(hash) {
                    let module = unsafe { Module::deserialize(engine, &entry.compiled) }
                        .map_err(|e| anyhow::anyhow!("AOT: failed to deserialize cached module: {}", e))?;
                    tracing::debug!("AOT cache: in-memory hit for hash {}", &hash[..8.min(hash.len())]);
                    return Ok(Some(module));
                }
            }

            // 2. Check disk cache
            if !config.aot_cache_dir.is_empty() {
                let path = std::path::Path::new(&config.aot_cache_dir)
                    .join(format!("{}.cwasm", hash));
                if path.exists() {
                    match std::fs::read(&path) {
                        Ok(bytes) => {
                            let module = unsafe { Module::deserialize(engine, &bytes) }
                                .map_err(|e| anyhow::anyhow!("AOT: failed to deserialize disk-cached module: {}", e))?;
                            tracing::debug!("AOT cache: disk hit for hash {}", &hash[..8.min(hash.len())]);
                            // Warm the in-memory cache
                            let _ = self.compile_and_cache_precompiled(hash, bytes);
                            return Ok(Some(module));
                        }
                        Err(e) => {
                            tracing::warn!("AOT cache: failed to read disk cache {}: {}", path.display(), e);
                        }
                    }
                }
            }
        }

        Ok(None)
    }

    /// Store already-compiled bytes in the in-memory cache (used when loading
    /// from disk to warm the memory cache).
    fn compile_and_cache_precompiled(&self, hash: &str, compiled: Vec<u8>) -> anyhow::Result<()> {
        let size = compiled.len();
        let counter = self.counter.fetch_add(1, Ordering::Relaxed);
        let entry = AotCacheEntry { compiled, size, inserted_at: counter };
        if let Ok(mut cache) = self.cache.write() {
            cache.insert(hash.to_string(), entry);
        }
        Ok(())
    }

    /// Get or compile a module, using the AOT cache when available.
    pub fn get_or_compile_module(
        &self,
        engine: &wasmtime::Engine,
        wasm_bytes: &[u8],
        config: &crate::config::Config,
    ) -> anyhow::Result<Module> {
        use std::collections::hash_map::DefaultHasher;
        use std::hash::{Hash, Hasher};

        // Compute a fast hash of the bytes for cache lookup
        let mut hasher = DefaultHasher::new();
        wasm_bytes.hash(&mut hasher);
        let hash = format!("{:016x}", hasher.finish());

        // Try cache first
        if let Some(module) = self.load_precompiled(engine, &hash, config)? {
            return Ok(module);
        }

        // Compile and cache
        let compiled = self.compile_and_cache(engine, wasm_bytes, &hash, config)?;
        let module = unsafe { Module::deserialize(engine, &compiled) }
            .map_err(|e| anyhow::anyhow!("AOT: failed to deserialize freshly compiled module: {}", e))?;
        Ok(module)
    }
}

impl Default for AotCache {
    fn default() -> Self {
        Self::new()
    }
}

impl Clone for AotCache {
    fn clone(&self) -> Self {
        Self {
            cache: Arc::clone(&self.cache),
            counter: Arc::clone(&self.counter),
        }
    }
}
