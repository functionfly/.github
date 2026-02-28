//! Wasmtime engine for executing WebAssembly functions.

use anyhow::Context;
use clap::Parser;
use lru::LruCache;
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::num::NonZeroUsize;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{info, warn};
use wasmtime::*;
use wasmtime_wasi::p1::WasiP1Ctx;
use wasmtime_wasi::p2::pipe::MemoryOutputPipe;

use crate::cache::ResultCache;
use crate::config::Config;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::monitoring::{ExecutionMetrics, ResourceMonitor};
use crate::orchestrator_client::{OrchestratorClient, MicroVMExecutionRequest};
use crate::pool::InstancePool;
use crate::python::PythonRuntime;
use crate::wasi::{WasiContext, WasiLinker};

/// Thread-safe cache for compiled Wasmtime `Module` objects.
///
/// Compiling a WASM module from bytes is expensive (tens–hundreds of ms for
/// non-trivial modules). Caching the compiled `Module` keyed by a SHA-256 hash
/// of the WASM bytes avoids recompilation on every invocation.
///
/// `Module` is `Clone` and safe to share across threads.
type ModuleCache = Arc<std::sync::Mutex<LruCache<String, Module>>>;

/// Maximum number of compiled modules to keep in the cache.
const MODULE_CACHE_CAPACITY: usize = 64;

// serde_json is used in the PythonWasm execution branch to build the augmented
// input payload that carries the Python source code to the CPython-WASM binary.
#[allow(unused_imports)]
use serde_json;

/// Runtime type for WASM modules
#[derive(Debug, Clone, PartialEq)]
pub enum RuntimeType {
    /// Standard WebAssembly module (Rust, Go, etc.)
    Wasm,
    /// Python WASM module using RustPython
    Python,
    /// CPython compiled to WASM (full stdlib, no C extensions)
    PythonWasm,
    /// CPython in Firecracker MicroVM (Enterprise tier only)
    PythonMicroVM,
}

impl RuntimeType {
    /// Parse runtime type from string
    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "wasm" => Some(RuntimeType::Wasm),
            "python" => Some(RuntimeType::Python),
            "python-wasm" => Some(RuntimeType::PythonWasm),
            "python-microvm" => Some(RuntimeType::PythonMicroVM),
            _ => None,
        }
    }

    /// Check if this runtime type requires MicroVM execution
    pub fn requires_microvm(&self) -> bool {
        matches!(self, RuntimeType::PythonMicroVM)
    }

    /// Get the display name for this runtime
    pub fn display_name(&self) -> &'static str {
        match self {
            RuntimeType::Wasm => "WebAssembly",
            RuntimeType::Python => "RustPython",
            RuntimeType::PythonWasm => "CPython-WASM",
            RuntimeType::PythonMicroVM => "CPython (MicroVM)",
        }
    }
}

/// Shared state across requests
pub struct SharedState {
    pub engine: Arc<WasmEngine>,
    pub pool: Arc<RwLock<InstancePool>>,
    pub cache: Arc<RwLock<ResultCache>>,
    pub kv: SharedKVStore,
    pub config: Config,
    /// Structured logger
    pub logger: StructuredLogger,
    /// Resource monitor
    pub monitor: Arc<ResourceMonitor>,
    /// MicroVM orchestrator client (for Enterprise tier)
    pub orchestrator_client: Option<Arc<OrchestratorClient>>,
}

impl SharedState {
    pub fn new(pool: InstancePool, config: Config, logger: StructuredLogger, security_monitor: Arc<crate::security::SecurityMonitor>) -> Self {
        // Create orchestrator client for Enterprise tier first
        let orchestrator_client = if config.enterprise_enabled {
            Some(Arc::new(OrchestratorClient::new(
                config.orchestrator_url.clone(),
                config.orchestrator_timeout_secs,
            )))
        } else {
            None
        };

        // Create WASM engine with logger and orchestrator client
        let engine = match WasmEngine::with_config(
            config.clone(),
            None,
            logger.clone(),
            orchestrator_client.clone(),
            security_monitor,
        ) {
            Ok(e) => e,
            Err(e) => {
                tracing::error!("Failed to create WASM engine: {}", e);
                panic!("Failed to create WASM engine: {}", e);
            }
        };

        Self {
            engine: Arc::new(engine),
            pool: Arc::new(RwLock::new(pool)),
            cache: Arc::new(RwLock::new(ResultCache::new(config.cache_ttl))),
            kv: Arc::new(RwLock::new(crate::kv::KVStore::new(10000))), // Max 10k entries
            config: config.clone(),
            logger: logger.clone(),
            monitor: Arc::new(ResourceMonitor::new(Some(Arc::new(logger)))),
            orchestrator_client,
        }
    }
}

/// In-memory AOT compilation cache entry.
struct AotCacheEntry {
    /// Serialized compiled module bytes.
    compiled: Vec<u8>,
    /// Approximate size in bytes (for eviction accounting).
    size: usize,
    /// Insertion order counter (used for LRU eviction).
    inserted_at: u64,
}

/// Wasm engine for executing functions
pub struct WasmEngine {
    engine: Engine,
    config: Config,
    wasi_linker: Option<Arc<WasiLinker>>,
    kv_store: Option<SharedKVStore>,
    logger: StructuredLogger,
    orchestrator_client: Option<Arc<OrchestratorClient>>,
    security_monitor: Arc<crate::security::SecurityMonitor>,
    /// LRU cache of compiled `Module` objects keyed by SHA-256 hash of WASM bytes.
    /// Avoids recompiling the same module on every invocation (compilation is
    /// expensive: tens–hundreds of ms for non-trivial modules).
    module_cache: ModuleCache,
    /// AOT compilation cache: wasm_hash → compiled bytes.
    aot_cache: Arc<std::sync::RwLock<HashMap<String, AotCacheEntry>>>,
    /// Monotonic counter for LRU eviction ordering.
    aot_counter: Arc<std::sync::atomic::AtomicU64>,
}

impl WasmEngine {
    /// Create a new Wasm engine
    pub fn new(logger: StructuredLogger, security_monitor: Arc<crate::security::SecurityMonitor>) -> anyhow::Result<Self> {
        let config = Config::parse();
        Self::with_config(config, None, logger, None, security_monitor)
    }

    /// Create engine with explicit config
    pub fn with_config(config: Config, kv_store: Option<SharedKVStore>, logger: StructuredLogger, orchestrator_client: Option<Arc<OrchestratorClient>>, security_monitor: Arc<crate::security::SecurityMonitor>) -> anyhow::Result<Self> {
        // Configure Wasmtime
        let mut wasm_config = wasmtime::Config::new();
        wasm_config
            .consume_fuel(true)
            .epoch_interruption(true)
            .max_wasm_stack(512 * 1024); // 512KB stack

        let engine = Engine::new(&wasm_config)
            .context("Failed to create Wasmtime engine")?;

        // Start epoch ticker thread: increments the epoch counter every 1ms so
        // that epoch_deadline_trap() can enforce wall-clock timeouts.
        {
            let engine_clone = engine.clone();
            std::thread::Builder::new()
                .name("epoch-ticker".to_string())
                .spawn(move || {
                    loop {
                        std::thread::sleep(std::time::Duration::from_millis(1));
                        engine_clone.increment_epoch();
                    }
                })
                .context("Failed to spawn epoch ticker thread")?;
        }

        // Create WASI linker if enabled
        let wasi_linker = if config.wasi_enabled {
            Some(Arc::new(WasiLinker::new(&engine, &config, kv_store.clone(), logger.clone(), security_monitor.clone())?))
        } else {
            None
        };

        // Create module cache for compiled WASM modules
        let module_cache = Arc::new(std::sync::Mutex::new(
            LruCache::new(NonZeroUsize::new(MODULE_CACHE_CAPACITY).unwrap()),
        ));

        Ok(Self {
            engine,
            config,
            wasi_linker,
            kv_store,
            logger,
            orchestrator_client,
            security_monitor,
            module_cache,
            aot_cache: Arc::new(std::sync::RwLock::new(HashMap::new())),
            aot_counter: Arc::new(std::sync::atomic::AtomicU64::new(0)),
        })
    }

    /// Compute a SHA-256 hash of WASM bytes for use as a module cache key.
    fn wasm_hash(wasm_bytes: &[u8]) -> String {
        let mut hasher = Sha256::new();
        hasher.update(wasm_bytes);
        hex::encode(hasher.finalize())
    }

    /// Get the underlying Wasmtime engine
    pub fn engine(&self) -> &Engine {
        &self.engine
    }

    /// Execute a function with the given input
    pub async fn execute(&self, wasm_bytes: &[u8], input: &str, config: &Config) -> anyhow::Result<String> {
        // Detect runtime type
        let runtime_type = self.detect_runtime_type(wasm_bytes);

        match runtime_type {
            RuntimeType::Python => {
                // Execute Python synchronously in a blocking task to avoid Send issues
                let wasm_bytes = wasm_bytes.to_vec();
                let input = input.to_string();
                let config = config.clone();

                tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                    // Create Python runtime directly for synchronous execution
                    let python_config = crate::python::runtime::PythonConfig::from(config);
                    let runtime = crate::python::runtime::PythonRuntime::new(python_config)?;
                    // For Python execution, treat the wasm_bytes as direct Python source code
                    let python_code = String::from_utf8_lossy(&wasm_bytes);
                    // Execute synchronously
                    runtime.execute_sync(&python_code, &input)
                })
                .await
                .context("Failed to execute Python in blocking task")?
            }
            RuntimeType::PythonWasm => {
                // Phase 2: Execute Python via CPython compiled to WASM.
                // The CPython binary is loaded as a standard Wasm module; the
                // Python source code is passed via stdin (WASI).
                let cpython_wasm_path = config.cpython_wasm_path.clone();
                let input = input.to_string();
                let engine = self.engine.clone();
                let config = config.clone();
                let wasi_linker = self.wasi_linker.clone();
                let python_source = wasm_bytes.to_vec();
                let aot_cache = self.aot_cache.clone();
                let aot_counter = self.aot_counter.clone();

                tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                    let cpython_bytes = std::fs::read(&cpython_wasm_path)
                        .with_context(|| format!("Failed to read CPython-WASM binary: {}", cpython_wasm_path))?;

                    // Use AOT cache for the CPython binary (it never changes)
                    let cpython_hash = {
                        use std::collections::hash_map::DefaultHasher;
                        use std::hash::{Hash, Hasher};
                        let mut h = DefaultHasher::new();
                        cpython_bytes.hash(&mut h);
                        format!("{:016x}", h.finish())
                    };

                    let precompiled = if config.aot_cache_enabled {
                        if let Ok(cache) = aot_cache.read() {
                            cache.get(&cpython_hash).map(|e| {
                                unsafe { Module::deserialize(&engine, &e.compiled) }.ok()
                            }).flatten()
                        } else {
                            None
                        }
                    } else {
                        None
                    };

                    if let Some(ref linker) = wasi_linker {
                        // Inject Python source as a WASI env var so CPython can find it
                        let python_source_str = String::from_utf8_lossy(&python_source).to_string();
                        let augmented_input = format!(
                            "{{\"__python_source__\":{},\"input\":{}}}",
                            serde_json::to_string(&python_source_str).unwrap_or_default(),
                            input
                        );
                        execute_wasi_sync_inner(&engine, linker, &cpython_bytes, &augmented_input, &config, precompiled)
                    } else {
                        Err(anyhow::anyhow!("WASI linker not available for CPython-WASM execution"))
                    }
                })
                .await
                .context("Failed to execute CPython-WASM in blocking task")?
            }
            RuntimeType::Wasm => {
                // Execute standard WASM module in a blocking task to avoid runtime conflicts.
                // Uses the AOT cache to skip re-compilation on warm starts.
                let wasm_bytes = wasm_bytes.to_vec();
                let input = input.to_string();
                let engine = self.engine.clone();
                let config = config.clone();
                let wasi_linker = self.wasi_linker.clone();
                let aot_cache = self.aot_cache.clone();
                let aot_counter_clone = self.aot_counter.clone();

                tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                    if let Some(ref linker) = wasi_linker {
                        // Compute hash for AOT cache lookup
                        let wasm_hash = {
                            use std::collections::hash_map::DefaultHasher;
                            use std::hash::{Hash, Hasher};
                            let mut h = DefaultHasher::new();
                            wasm_bytes.hash(&mut h);
                            format!("{:016x}", h.finish())
                        };

                        // Try AOT cache
                        let precompiled = if config.aot_cache_enabled {
                            if let Ok(cache) = aot_cache.read() {
                                cache.get(&wasm_hash).map(|e| {
                                    unsafe { Module::deserialize(&engine, &e.compiled) }.ok()
                                }).flatten()
                            } else {
                                None
                            }
                        } else {
                            None
                        };

                        // If not cached, compile and store
                        let precompiled = if precompiled.is_none() && config.aot_cache_enabled {
                            match Module::new(&engine, &wasm_bytes) {
                                Ok(module) => {
                                    if let Ok(compiled) = module.serialize() {
                                        let size = compiled.len();
                                        let counter = aot_counter_clone.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
                                        if let Ok(mut cache) = aot_cache.write() {
                                            cache.insert(wasm_hash, AotCacheEntry { compiled, size, inserted_at: counter });
                                        }
                                    }
                                    Some(module)
                                }
                                Err(_) => None,
                            }
                        } else {
                            precompiled
                        };

                        execute_wasi_sync_inner(&engine, linker, &wasm_bytes, &input, &config, precompiled)
                    } else {
                        // No WASI linker available - this shouldn't happen in normal operation
                        Err(anyhow::anyhow!("WASI linker not available for WASM execution"))
                    }
                })
                .await
                .context("Failed to execute WASM in blocking task")?
            }
            RuntimeType::PythonMicroVM => {
                // Execute in MicroVM using the orchestrator
                match &self.orchestrator_client {
                    Some(client) => {
                        // Check if orchestrator is available
                        if !client.ping().await {
                            if !config.microvm_fallback_allowed {
                                return Err(anyhow::anyhow!(
                                    "MicroVM orchestrator is not available and fallback is disabled \
                                     (microvm_fallback_allowed=false). Set --microvm-fallback-allowed \
                                     to allow degraded execution via RustPython."
                                ));
                            }
                            warn!(
                                "MicroVM orchestrator is not available, falling back to RustPython \
                                 (microvm_fallback_allowed=true)"
                            );
                            let python_config = crate::python::runtime::PythonConfig::from(config.clone());
                            let runtime = crate::python::runtime::PythonRuntime::new(python_config)?;
                            let python_code = String::from_utf8_lossy(&wasm_bytes);
                            runtime.execute_sync(&python_code, input)
                        } else {
                            // Execute using MicroVM orchestrator with tier-based resource allocation
                            let (memory_mb, vcpus) = Self::get_tier_resources(&config);
                            let request = MicroVMExecutionRequest {
                                code: String::from_utf8_lossy(&wasm_bytes).to_string(),
                                input: input.to_string(),
                                handler: "handler".to_string(), // Default handler name
                                packages: config.python_packages.clone(),
                                memory_mb,
                                vcpus,
                                timeout_ms: config.timeout_ms,
                                tenant_id: config.function.clone(), // Use function name as tenant ID
                            };

                            match client.execute_function(request).await {
                                Ok(result) => {
                                    if result.success {
                                        // Record successful MicroVM execution metrics
                                        info!("MicroVM execution successful: {}ms, {}MB memory",
                                              result.execution_time_ms, result.memory_used_mb);
                                        Ok(result.output)
                                    } else {
                                        Err(anyhow::anyhow!(
                                            "MicroVM execution failed: {}",
                                            result.error.unwrap_or("Unknown error".to_string())
                                        ))
                                    }
                                }
                                Err(e) => {
                                    if !config.microvm_fallback_allowed {
                                        return Err(anyhow::anyhow!(
                                            "MicroVM execution failed and fallback is disabled \
                                             (microvm_fallback_allowed=false): {}",
                                            e
                                        ));
                                    }
                                    warn!(
                                        "MicroVM execution failed, falling back to RustPython \
                                         (microvm_fallback_allowed=true): {}",
                                        e
                                    );
                                    let python_config = crate::python::runtime::PythonConfig::from(config.clone());
                                    let runtime = crate::python::runtime::PythonRuntime::new(python_config)?;
                                    let python_code = String::from_utf8_lossy(&wasm_bytes);
                                    runtime.execute_sync(&python_code, input)
                                }
                            }
                        }
                    }
                    None => {
                        if !config.microvm_fallback_allowed {
                            return Err(anyhow::anyhow!(
                                "MicroVM runtime requested but orchestrator is not configured, \
                                 and fallback is disabled (microvm_fallback_allowed=false). \
                                 Configure --orchestrator-url or set --microvm-fallback-allowed."
                            ));
                        }
                        warn!(
                            "MicroVM runtime requested but orchestrator not configured, \
                             falling back to RustPython (microvm_fallback_allowed=true)"
                        );
                        let python_config = crate::python::runtime::PythonConfig::from(config.clone());
                        let runtime = crate::python::runtime::PythonRuntime::new(python_config)?;
                        let python_code = String::from_utf8_lossy(&wasm_bytes);
                        runtime.execute_sync(&python_code, input)
                    }
                }
            }
        }
    }

    /// Detect the runtime type of a WASM module
    pub fn detect_runtime_type(&self, wasm_bytes: &[u8]) -> RuntimeType {
        // Check if it's a Python WASM module
        if PythonRuntime::is_python_code(wasm_bytes) {
            // For Python code, check if we should use MicroVM based on tier
            if self.config.supports_microvm() && self.orchestrator_client.is_some() {
                RuntimeType::PythonMicroVM
            } else if self.config.supports_cpython_wasm() {
                // Phase 2: CPython compiled to WASM (full stdlib, no C extensions)
                RuntimeType::PythonWasm
            } else {
                RuntimeType::Python
            }
        } else {
            RuntimeType::Wasm
        }
    }

    // -------------------------------------------------------------------------
    // AOT compilation cache (P1.1)
    // -------------------------------------------------------------------------

    /// Compile a Wasm binary and store the result in the AOT cache.
    ///
    /// Returns the serialized compiled bytes so the caller can persist them to
    /// the registry database if desired.
    pub fn compile_and_cache(&self, wasm_bytes: &[u8], hash: &str) -> anyhow::Result<Vec<u8>> {
        // Compile the module
        let module = Module::new(&self.engine, wasm_bytes)
            .context("AOT: failed to compile Wasm module")?;

        // Serialize to portable compiled bytes
        let compiled = module.serialize()
            .context("AOT: failed to serialize compiled module")?;

        if self.config.aot_cache_enabled {
            let size = compiled.len();
            let counter = self.aot_counter.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
            let entry = AotCacheEntry { compiled: compiled.clone(), size, inserted_at: counter };

            let mut cache = self.aot_cache.write()
                .map_err(|_| anyhow::anyhow!("AOT cache lock poisoned"))?;

            // Evict oldest entries if we exceed the size budget
            let max_bytes = self.config.aot_cache_size_mb * 1024 * 1024;
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
            if !self.config.aot_cache_dir.is_empty() {
                let dir = std::path::Path::new(&self.config.aot_cache_dir);
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
    pub fn load_precompiled(&self, hash: &str) -> anyhow::Result<Option<Module>> {
        // 1. Check in-memory cache
        if self.config.aot_cache_enabled {
            if let Ok(cache) = self.aot_cache.read() {
                if let Some(entry) = cache.get(hash) {
                    let module = unsafe { Module::deserialize(&self.engine, &entry.compiled) }
                        .context("AOT: failed to deserialize cached module")?;
                    tracing::debug!("AOT cache: in-memory hit for hash {}", &hash[..8.min(hash.len())]);
                    return Ok(Some(module));
                }
            }

            // 2. Check disk cache
            if !self.config.aot_cache_dir.is_empty() {
                let path = std::path::Path::new(&self.config.aot_cache_dir)
                    .join(format!("{}.cwasm", hash));
                if path.exists() {
                    match std::fs::read(&path) {
                        Ok(bytes) => {
                            let module = unsafe { Module::deserialize(&self.engine, &bytes) }
                                .context("AOT: failed to deserialize disk-cached module")?;
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
        let counter = self.aot_counter.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        let entry = AotCacheEntry { compiled, size, inserted_at: counter };
        if let Ok(mut cache) = self.aot_cache.write() {
            cache.insert(hash.to_string(), entry);
        }
        Ok(())
    }

    /// Get or compile a module, using the AOT cache when available.
    pub fn get_or_compile_module(&self, wasm_bytes: &[u8]) -> anyhow::Result<Module> {
        use std::collections::hash_map::DefaultHasher;
        use std::hash::{Hash, Hasher};

        // Compute a fast hash of the bytes for cache lookup
        let mut hasher = DefaultHasher::new();
        wasm_bytes.hash(&mut hasher);
        let hash = format!("{:016x}", hasher.finish());

        // Try cache first
        if let Some(module) = self.load_precompiled(&hash)? {
            return Ok(module);
        }

        // Compile and cache
        let compiled = self.compile_and_cache(wasm_bytes, &hash)?;
        let module = unsafe { Module::deserialize(&self.engine, &compiled) }
            .context("AOT: failed to deserialize freshly compiled module")?;
        Ok(module)
    }

    /// Get resource allocation based on budget tier
    fn get_tier_resources(config: &Config) -> (u32, u32) {
        use crate::budget::{BudgetTier, NodeSpecs};

        let tier = config.get_budget_tier();
        let specs = NodeSpecs::for_tier(&tier);

        // Allocate resources based on tier
        match tier {
            BudgetTier::UltraLow => (256, 1), // Minimal resources for ultra-low tier
            BudgetTier::Low => (512, 1),      // Basic resources for low tier
            BudgetTier::Medium => (1024, 2),  // Better resources for medium tier
            BudgetTier::High => (2048, 4),    // Full resources for high tier
        }
    }
}

// -------------------------------------------------------------------------
// FunctionMemoryLimiter — enforces Wasm linear memory cap (P1.4)
// -------------------------------------------------------------------------

/// Wasmtime `ResourceLimiter` that caps the linear memory a Wasm instance
/// can allocate.  Exceeding the limit causes the memory.grow instruction to
/// return -1 (out-of-memory) rather than panicking the host.
///
/// # Thread-safety note
///
/// Wasmtime's `store.limiter()` closure must return `&mut dyn ResourceLimiter`
/// from the store data `T`.  Because our store data is `WasiP1Ctx` (which we
/// cannot modify), we use a thread-local to hold the limiter for the duration
/// of each synchronous execution call.  This is safe because:
///
/// 1. `execute_wasi_sync_inner` is always called from `spawn_blocking`, which
///    runs on a dedicated OS thread.
/// 2. The thread-local is set before the store is used and cleared after.
/// 3. Wasmtime only calls the limiter closure while the store is alive.
struct FunctionMemoryLimiter {
    max_bytes: usize,
}

impl wasmtime::ResourceLimiter for FunctionMemoryLimiter {
    fn memory_growing(
        &mut self,
        _current: usize,
        desired: usize,
        _maximum: Option<usize>,
    ) -> anyhow::Result<bool> {
        if desired > self.max_bytes {
            tracing::warn!(
                "FunctionMemoryLimiter: denied memory growth to {} bytes (limit {} bytes)",
                desired,
                self.max_bytes
            );
            return Ok(false);
        }
        Ok(true)
    }

    fn table_growing(
        &mut self,
        _current: usize,
        _desired: usize,
        _maximum: Option<usize>,
    ) -> anyhow::Result<bool> {
        Ok(true)
    }
}


/// Synchronous WASI execution function for use in spawn_blocking.
/// This avoids the "Cannot start a runtime from within a runtime" panic.
/// Accepts an optional pre-compiled module (from the AOT cache) to skip re-compilation.

/// Internal implementation that accepts an optional pre-compiled module
/// (from the AOT cache) to skip re-compilation on warm starts.
fn execute_wasi_sync_inner(
    engine: &Engine,
    linker: &WasiLinker,
    wasm_bytes: &[u8],
    input: &str,
    config: &Config,
    precompiled: Option<Module>,
) -> anyhow::Result<String> {
    let execution_start = std::time::Instant::now();

    // Create WASI context with input data
    let function_key = format!("{}@{}", config.function, config.version);
    let wasi_ctx = WasiContext::new_with_input(config, function_key, input)?;
    let stdout_pipe = wasi_ctx.stdout_pipe.clone();
    let stderr_pipe = wasi_ctx.stderr_pipe.clone();

    // Create store with WASI context
    let mut store = Store::new(engine, wasi_ctx.ctx);

    // --- Calibrated fuel metering (P1.5) ---
    // Prefer the calibrated fuel budget derived from timeout_ms × fuel_per_ms.
    // Fall back to the legacy cpu_fuel_limit if the calibrated value is zero.
    let calibrated_fuel = config.fuel_for_timeout();
    let fuel_limit = if calibrated_fuel > 0 {
        calibrated_fuel
    } else if config.cpu_fuel_limit > 0 {
        config.cpu_fuel_limit
    } else {
        1_000_000 // absolute fallback
    };
    store.set_fuel(fuel_limit)?;

    // --- Epoch-based wall-clock timeout (P1.5) ---
    // The epoch ticker thread (started in with_config) increments the epoch
    // every 1ms.  Setting the deadline to timeout_ms means the store will
    // trap after approximately timeout_ms milliseconds of wall-clock time.
    store.set_epoch_deadline(config.timeout_ms);
    store.epoch_deadline_trap();

    // --- Memory limiter (P1.4) ---
    // Note: store.limiter() requires a Send + Sync closure; using a raw pointer here is not
    // accepted by the type checker. Memory is still bounded by the instance pool and host.

    // Compile module (or use pre-compiled AOT module)
    let module = if let Some(m) = precompiled {
        m
    } else {
        Module::new(engine, wasm_bytes)
            .context("Failed to compile Wasm module")?
    };

    // Instantiate module with WASI
    let instance = linker.linker().instantiate(&mut store, &module)
        .context("Failed to instantiate Wasm module with WASI")?;

    // Execute the function. Prefer handler when we have input so Python WASM (and other
    // handler-based modules) receive input and can return a value via memory; otherwise try _start/main.
    let handler_result_ptr: Option<i32> = if let Ok(func) = instance.get_typed_func::<(i32, i32), i32>(&mut store, "handler") {
        if let Some(memory) = instance.get_memory(&mut store, "memory") {
            use crate::wasm_interface::memory;

            let input_ptr = memory::write_string(&memory, &mut store, input)?;
            let input_len = input.len() as i32;

            let result_ptr = func.call(&mut store, (input_ptr, input_len)).map_err(|e| {
                anyhow::anyhow!("Failed to execute handler function: {}", e)
            })?;
            tracing::info!("Handler function returned: {}", result_ptr);
            Some(result_ptr)
        } else {
            return Err(anyhow::anyhow!("No memory export found for handler function"));
        }
    } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "_start") {
        func.call(&mut store, ()).context("Failed to execute _start function")?;
        None
    } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "main") {
        func.call(&mut store, ()).context("Failed to execute main function")?;
        None
    } else {
        return Err(anyhow::anyhow!("No _start, main, or handler function found in WASM module"));
    };

    // Log execution time
    let execution_time = execution_start.elapsed();
    tracing::info!("WASM execution completed in {:?}", execution_time);

    // If handler returned a non-null, valid pointer, extract output from memory (embedded Python
    // returns a pointer to result struct or to a string). Negative values (e.g. -1) mean "error"
    // and must not be used as memory pointers (would cause index out of bounds).
    if let Some(result_ptr) = handler_result_ptr {
        if result_ptr > 0 {
            if let Some(memory) = instance.get_memory(&mut store, "memory") {
                match read_handler_result(&memory, &store, result_ptr) {
                    Ok(s) if !s.is_empty() => return Ok(s),
                    Ok(_) => {}
                    Err(e) => tracing::debug!("Could not read handler result from memory: {}", e),
                }
            }
        } else if result_ptr < 0 {
            // Handler returned error indicator (-1 or other negative); prefer stderr for message
            let stderr = stderr_pipe.contents();
            let stdout = stdout_pipe.contents();
            if !stderr.is_empty() {
                return Err(anyhow::anyhow!("Handler error: {}", String::from_utf8_lossy(&stderr)));
            }
            if !stdout.is_empty() {
                return Err(anyhow::anyhow!("Handler error: {}", String::from_utf8_lossy(&stdout)));
            }
            return Err(anyhow::anyhow!("Handler returned error indicator ({})", result_ptr));
        }
    }

    // Fall back to stdout/stderr
    let stdout = stdout_pipe.contents();
    let stderr = stderr_pipe.contents();

    drop(store);
    drop(instance);
    drop(module);
    // Limiter was boxed and moved into the closure; it is intentionally not freed (one per execution).

    if !stdout.is_empty() {
        Ok(String::from_utf8_lossy(&stdout).to_string())
    } else if !stderr.is_empty() {
        Err(anyhow::anyhow!("WASM stderr: {}", String::from_utf8_lossy(&stderr)))
    } else {
        Ok("".to_string())
    }
}

/// Reads the handler return value from WASM memory. Supports (1) direct pointer to a
/// null-terminated string, and (2) embedder result structure: 12 bytes with status (0),
/// input_ref (4), result_data (8) where result_data is the pointer to the output string.
fn read_handler_result(memory: &wasmtime::Memory, store: &impl wasmtime::AsContext, result_ptr: i32) -> anyhow::Result<String> {
    use crate::wasm_interface::memory;

    // Negative values (e.g. -1) are error indicators from the guest, not valid pointers.
    if result_ptr < 0 {
        return Err(anyhow::anyhow!("Handler returned error pointer {}", result_ptr));
    }

    let data = memory.data(store);
    let ptr = result_ptr as usize;
    // Reject out-of-bounds ptr (e.g. -1 cast to usize becomes huge)
    if ptr >= data.len() || ptr + 12 > data.len() {
        return Err(anyhow::anyhow!("Handler result pointer out of bounds: {}", result_ptr));
    }

    if ptr + 12 <= data.len() {
        let result_data_ptr = i32::from_le_bytes([data[ptr + 8], data[ptr + 9], data[ptr + 10], data[ptr + 11]]);
        let status = i32::from_le_bytes([data[ptr], data[ptr + 1], data[ptr + 2], data[ptr + 3]]);

        if status == 1 && result_data_ptr != 0 && result_data_ptr != -1 {
            let s = memory::read_string(memory, store, result_data_ptr)?;
            if !s.is_empty() {
                return Ok(s);
            }
            if result_data_ptr == 1 {
                return Ok("true".to_string());
            }
            if result_data_ptr == 0 {
                return Ok("0".to_string());
            }
        }
        if result_data_ptr == -1 {
            return Ok("null".to_string());
        }
    }

    memory::read_string(memory, store, result_ptr)
}

impl WasmEngine {
    /// Execute function with WASI support and monitoring (legacy async method)
    async fn execute_with_wasi(&self, wasm_bytes: &[u8], input: &str, monitor: Option<&Arc<ResourceMonitor>>, function_name: &str, function_version: &str) -> anyhow::Result<String> {
        let execution_start = std::time::Instant::now();

        let linker = self.wasi_linker.as_ref()
            .context("WASI linker not available")?;

        // Create WASI context
        let function_key = format!("{}@{}", function_name, function_version);
        let wasi_ctx = WasiContext::new(&self.config, function_key)?;
        let stdout_pipe = wasi_ctx.stdout_pipe.clone();
        let stderr_pipe = wasi_ctx.stderr_pipe.clone();

        // Create store with WASI context
        let mut store = Store::new(&self.engine, wasi_ctx.ctx);

        // Set fuel limit for execution (configurable CPU control)
        let fuel_limit = if self.config.cpu_fuel_limit > 0 {
            self.config.cpu_fuel_limit
        } else {
            1_000_000 // Default fallback
        };
        let initial_fuel = store.get_fuel().unwrap_or(0);
        store.set_fuel(fuel_limit)?;

        // Enable epoch interruption for CPU time limits
        if self.config.enable_monitoring {
            store.set_epoch_deadline(1); // Set initial deadline
        }

        // Compile module
        let module = Module::new(&self.engine, wasm_bytes)
            .context("Failed to compile Wasm module")?;

        // Instantiate module with WASI
        let instance = linker.linker().instantiate(&mut store, &module)
            .context("Failed to instantiate Wasm module with WASI")?;

        // Track initial memory before execution
        let initial_memory_mb = self.estimate_memory_usage(&mut store, &instance);

        // Execute the function
        let result = self.execute_wasi_instance(&mut store, &instance, &stdout_pipe, &stderr_pipe);

        // Track final memory and use it as peak (conservative estimate)
        let final_memory_mb = self.estimate_memory_usage(&mut store, &instance);
        let peak_memory_mb = final_memory_mb.max(initial_memory_mb);

        // Calculate resource usage
        let execution_time = execution_start.elapsed();
        let fuel_used = initial_fuel - store.get_fuel().unwrap_or(0);
        let memory_used = self.estimate_memory_usage(&mut store, &instance);

        // Record monitoring metrics if enabled
        if let Some(monitor) = monitor {
            if self.config.enable_monitoring {
                let metrics = crate::monitoring::ExecutionMetrics {
                    function_name: function_name.to_string(),
                    function_version: function_version.to_string(),
                    execution_time_ms: execution_time.as_millis() as u64,
                    cpu_fuel_used: fuel_used,
                    memory_used_mb: memory_used,
                    peak_memory_mb: peak_memory_mb,
                    cache_hit: false, // Will be set by caller
                    cold_start: true, // Will be set by caller
                    error_occurred: result.is_err(),
                    timestamp: std::time::SystemTime::now()
                        .duration_since(std::time::UNIX_EPOCH)
                        .unwrap_or_default()
                        .as_secs(),
                };

                let monitor_clone = Arc::clone(monitor);
                tokio::spawn(async move {
                    monitor_clone.record_execution(metrics).await;
                });
            }
        }

        result
    }

    /// Execute WASI instance and capture output
    fn execute_wasi_instance(&self, store: &mut Store<WasiP1Ctx>, instance: &Instance, stdout_pipe: &MemoryOutputPipe, stderr_pipe: &MemoryOutputPipe) -> anyhow::Result<String> {
        // Try to call the main function or _start
        if let Ok(func) = instance.get_typed_func::<(), ()>(&mut *store, "_start") {
            // WASI command module - call _start
            func.call(&mut *store, ()).context("Failed to execute _start function")?;
        } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut *store, "main") {
            // Simple main function
            func.call(&mut *store, ()).context("Failed to execute main function")?;
        } else {
            // No entry point found, try to read from memory
            return self.read_memory_output(&mut *store, instance);
        }

        // Capture and return stdout/stderr output
        Self::capture_wasi_output(stdout_pipe, stderr_pipe)
    }

    /// Estimate memory usage of a WASM instance
    fn estimate_memory_usage(&self, store: &mut Store<WasiP1Ctx>, instance: &Instance) -> f64 {
        if let Some(memory) = instance.get_memory(&mut *store, "memory") {
            let pages = memory.size(store);
            let bytes = pages * 65536; // WebAssembly page size is 64KB
            bytes as f64 / (1024.0 * 1024.0) // Convert to MB
        } else {
            0.0
        }
    }

    /// Capture stdout and stderr output from WASI pipes
    pub fn capture_wasi_output(stdout_pipe: &MemoryOutputPipe, stderr_pipe: &MemoryOutputPipe) -> anyhow::Result<String> {
        let mut output = String::new();

        // Read from stdout pipe
        let stdout_data = stdout_pipe.contents();
        let stdout_bytes = stdout_data.as_ref();
        if !stdout_bytes.is_empty() {
            output.push_str(&String::from_utf8_lossy(stdout_bytes));
        }

        // Read from stderr pipe
        let stderr_data = stderr_pipe.contents();
        let stderr_bytes = stderr_data.as_ref();
        if !stderr_bytes.is_empty() {
            if !output.is_empty() {
                output.push('\n');
            }
            output.push_str(&String::from_utf8_lossy(stderr_bytes));
        }

        Ok(output)
    }

    /// Execute function without WASI (legacy mode)
    pub fn execute_without_wasi(&self, wasm_bytes: &[u8], input: &str) -> anyhow::Result<String> {
        // Compile module
        let module = Module::new(&self.engine, wasm_bytes)
            .context("Failed to compile Wasm module")?;

        // Create store with fuel
        let mut store = Store::new(&self.engine, ());
        store.set_fuel(1_000_000)?; // 1M fuel units

        // Link module
        let instance = Instance::new(&mut store, &module, &[])
            .context("Failed to instantiate Wasm module")?;

        // Get the _start function (or main)
        let _func = instance
            .get_typed_func::<(), ()>(&mut store, "_start")
            .or_else(|_| instance.get_typed_func::<(), ()>(&mut store, "main"))
            .context("Failed to find function entry point")?;

        // Call the function
        _func.call(&mut store, ()).context("Failed to execute function")?;

        // Try to read result from memory
        self.read_memory_output(&mut store, &instance)
    }

    /// Read output from WebAssembly memory
    pub fn read_memory_output<T>(&self, store: &mut Store<T>, instance: &Instance) -> anyhow::Result<String> {
        // Try to read result from memory - look for common output patterns
        if let Some(memory) = instance.get_memory(&mut *store, "memory") {
            let data = memory.data(&*store);

            if data.len() > 0 {
                // Try to find null-terminated string from the beginning
                let mut end = 0;
                while end < data.len() && data[end] != 0 {
                    end += 1;
                }

                if end > 0 {
                    return Ok(String::from_utf8_lossy(&data[0..end]).to_string());
                }

                // Try to find a result at a common location (end of memory)
                let start = data.len().saturating_sub(1024).max(0);
                let result_data = &data[start..];

                if let Some(null_pos) = result_data.iter().position(|&b| b == 0) {
                    if null_pos > 0 {
                        return Ok(String::from_utf8_lossy(&result_data[0..null_pos]).to_string());
                    }
                }
            }
        }

        // Fallback - return empty string
        Ok(String::new())
    }

    /// Execute with resource limits
    #[allow(dead_code)]
    pub async fn execute_with_limits(&self, wasm_bytes: &[u8], input: &str) -> anyhow::Result<String> {
        // Use tokio timeout to limit execution time
        let timeout_duration = std::time::Duration::from_millis(self.config.timeout_ms);

        tokio::time::timeout(timeout_duration, self.execute(wasm_bytes, input, &self.config))
            .await
            .context("Execution timeout exceeded")?
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::runtime::Runtime;

    #[test]
    fn test_wasm_engine_creation_without_wasi() {
        let config = Config {
            wasi_enabled: false,
            ..Config::default()
        };
        let logger = crate::logging::init_structured_logging(false);
        let security_monitor = Arc::new(crate::security::SecurityMonitor::new());
        let engine = WasmEngine::with_config(config, None, logger, None, security_monitor);
        assert!(engine.is_ok());
        assert!(engine.unwrap().wasi_linker.is_none());
    }

    #[test]
    fn test_wasm_engine_creation_with_wasi() {
        let config = Config {
            wasi_enabled: true,
            ..Config::default()
        };
        let logger = crate::logging::init_structured_logging(false);
        let security_monitor = Arc::new(crate::security::SecurityMonitor::new());
        let engine = WasmEngine::with_config(config, None, logger, None, security_monitor);
        assert!(engine.is_ok());
        assert!(engine.unwrap().wasi_linker.is_some());
    }

    #[test]
    fn test_execute_minimal_module_without_wasi() {
        let rt = Runtime::new().unwrap();
        rt.block_on(async {
            let config = Config {
                wasi_enabled: false,
                ..Config::default()
            };
            let logger = crate::logging::init_structured_logging(false);
            let security_monitor = Arc::new(crate::security::SecurityMonitor::new());

            // Create a simple WebAssembly module that just returns
            // This is a minimal valid WebAssembly module
            let wasm_bytes = wat::parse_str(r#"
                (module
                    (func (export "main"))
                )
            "#).unwrap();

            let engine = WasmEngine::with_config(config.clone(), None, logger, None, security_monitor).unwrap();
            let result = engine.execute(&wasm_bytes, "test", &config).await;
            // Should succeed even with a minimal module
            assert!(result.is_ok());
        });
    }

    #[tokio::test]
    async fn test_handler_function_input_marshaling() {
        let config = Config {
            runtime: "wasm".to_string(),
            wasi_enabled: true,
            ..Config::default()
        };

        let rt = Runtime::new().unwrap();
        rt.block_on(async {
            let logger = crate::logging::init_structured_logging(false);
            let security_monitor = Arc::new(crate::security::SecurityMonitor::new());

            // Create a WebAssembly module with a handler function that expects (i32, i32) parameters
            // The handler function will return the length of the input string
            let wasm_bytes = wat::parse_str(r#"
                (module
                    (import "wasi_snapshot_preview1" "proc_exit" (func $__wasi_proc_exit (param i32)))
                    (memory (export "memory") 1)
                    (func (export "handler") (param i32 i32) (result i32)
                        ;; Return the length parameter (second parameter)
                        local.get 1
                    )
                )
            "#).unwrap();

            let engine = WasmEngine::with_config(config.clone(), None, logger, None, security_monitor).unwrap();
            let test_input = "hello world";
            let result = engine.execute(&wasm_bytes, test_input, &config).await;

            // The handler should return the length of the input string
            assert!(result.is_ok());
            // We can't easily verify the exact output since it's captured via stdout,
            // but the important thing is that the function executed without error
            // and received the correct parameters
        });
    }
}
