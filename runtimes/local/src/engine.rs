//! Wasmtime engine for executing WebAssembly functions.

use anyhow::Context;
use clap::Parser;
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{info, warn};
use wasmtime::*;
use wasmtime_wasi::p1::WasiP1Ctx;

use crate::cache::ResultCache;
use crate::config::Config;
use crate::errors::RuntimeError;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::monitoring::ResourceMonitor;
use crate::orchestrator_client::{OrchestratorClient, MicroVMExecutionRequest};
use crate::pool::{InstancePool, PoolManager, PooledWasmInstance};
use crate::python::engine::{PythonEngine, PythonSharedState};
use crate::python::runtime::PythonRuntime;
use crate::python_pool::PythonRuntimePool;
use crate::wasi::{WasiContext, WasiLinker};

// serde_json is used in the PythonWasm execution branch to build the augmented
// input payload that carries the Python source code to the CPython-WASM binary.
#[allow(unused_imports)]
use serde_json as _;

/// Check if an error message indicates an epoch deadline (timeout) failure.
/// Wasmtime uses epoch-based interruption for wall-clock timeouts.
fn is_epoch_deadline_error(msg: &str) -> bool {
    msg.contains("deadline") || msg.contains("epoch") || msg.contains("interruption")
}

/// Convert epoch deadline errors to RuntimeError::timeout() for consistent error handling.
/// This ensures timeout errors are properly categorized as ErrorKind::TimeoutExceeded.
fn convert_timeout_error(err: anyhow::Error, timeout_ms: u64) -> anyhow::Error {
    let msg = err.to_string();
    if is_epoch_deadline_error(&msg) {
        return anyhow::anyhow!(RuntimeError::timeout(timeout_ms));
    }
    err
}

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
    /// WASM instance pool manager for warm-instance reuse
    pub wasm_pool: Option<Arc<PoolManager>>,
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

        // Create shared Python state for RustPython fallback execution (before WasmEngine)
        let python_shared_state = match PythonSharedState::new(config.clone().into()) {
            Ok(state) => {
                tracing::info!("Shared Python engine initialized");
                Some(Arc::new(state))
            }
            Err(e) => {
                tracing::warn!("Failed to create shared Python engine: {}. \
                    RustPython fallback will create engines per-request.", e);
                None
            }
        };

        // Create WASM engine with logger, orchestrator client, and python shared state
        let engine = match WasmEngine::with_config(
            config.clone(),
            None,
            logger.clone(),
            orchestrator_client.clone(),
            security_monitor,
            python_shared_state.clone(),
        ) {
            Ok(e) => e,
            Err(e) => {
                tracing::error!("Failed to create WASM engine: {}", e);
                panic!("Failed to create WASM engine: {}", e);
            }
        };

        // Create WASM pool manager if enabled
        let wasm_pool = if config.wasm_pool_enabled {
            Some(Arc::new(PoolManager::new(
                config.wasm_pool_max_concurrent,
                config.wasm_pool_max_idle,
            )))
        } else {
            None
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
            wasm_pool,
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
    #[allow(dead_code)]
    kv_store: Option<SharedKVStore>,
    #[allow(dead_code)]
    logger: StructuredLogger,
    orchestrator_client: Option<Arc<OrchestratorClient>>,
    #[allow(dead_code)]
    security_monitor: Arc<crate::security::SecurityMonitor>,
    /// Shared Python engine for RustPython fallback execution
    python_shared_state: Option<Arc<PythonSharedState>>,

    /// AOT compilation cache: wasm_hash → compiled bytes.
    aot_cache: Arc<std::sync::RwLock<HashMap<String, AotCacheEntry>>>,
    /// Monotonic counter for LRU eviction ordering.
    aot_counter: Arc<std::sync::atomic::AtomicU64>,
}

impl WasmEngine {
    /// Create a new Wasm engine
    #[allow(dead_code)]
    pub fn new(logger: StructuredLogger, security_monitor: Arc<crate::security::SecurityMonitor>) -> anyhow::Result<Self> {
        let config = Config::parse();
        Self::with_config(config, None, logger, None, security_monitor, None)
    }

    /// Create engine with explicit config
    pub fn with_config(
        config: Config,
        kv_store: Option<SharedKVStore>,
        logger: StructuredLogger,
        orchestrator_client: Option<Arc<OrchestratorClient>>,
        security_monitor: Arc<crate::security::SecurityMonitor>,
        python_shared_state: Option<Arc<PythonSharedState>>,
    ) -> anyhow::Result<Self> {
        // Configure Wasmtime
        let mut wasm_config = wasmtime::Config::new();
        wasm_config
            .consume_fuel(true)
            .epoch_interruption(true)
            .max_wasm_stack(512 * 1024); // 512KB stack

        let engine = Engine::new(&wasm_config)
            .map_err(|e| anyhow::anyhow!("Failed to create Wasmtime engine: {}", e))?;

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

        Ok(Self {
            engine,
            config,
            wasi_linker,
            kv_store,
            logger,
            orchestrator_client,
            security_monitor,
            python_shared_state,
            aot_cache: Arc::new(std::sync::RwLock::new(HashMap::new())),
            aot_counter: Arc::new(std::sync::atomic::AtomicU64::new(0)),
        })
    }

    /// Compute a SHA-256 hash of WASM bytes for use as a module cache key.
    ///
    /// This hash is used for AOT compilation cache keys and module identification.
    #[allow(dead_code)]
    pub fn wasm_hash(wasm_bytes: &[u8]) -> String {
        let mut hasher = Sha256::new();
        hasher.update(wasm_bytes);
        hex::encode(hasher.finalize())
    }

    /// Get the underlying Wasmtime engine
    pub fn engine(&self) -> &Engine {
        &self.engine
    }

    /// Get the WASI linker if available
    pub fn wasi_linker(&self) -> Option<&Arc<crate::wasi::WasiLinker>> {
        self.wasi_linker.as_ref()
    }

    /// Execute a function with the given input
    ///
    /// If `python_pool` is provided and the runtime is Python, the pooled
    /// interpreter will be used for efficient reuse.
    /// Otherwise a fresh interpreter is created per call.
    ///
    /// If `micropython_executor` is provided and the runtime is Python with
    /// MicroPython enabled, the MicroPython executor will be used for WASM-based
    /// Python execution with module linking support.
    pub async fn execute(
        &self,
        wasm_bytes: &[u8],
        input: &str,
        config: &Config,
        python_pool: Option<Arc<PythonRuntimePool>>,
        micropython_executor: Option<Arc<crate::micropython::MicroPythonExecutor>>,
    ) -> anyhow::Result<String> {
        // Detect runtime type
        let runtime_type = self.detect_runtime_type(wasm_bytes);
        tracing::debug!("Executing function with runtime: {}", runtime_type.display_name());

        match runtime_type {
            RuntimeType::Python => {
                // Execute Python synchronously in a blocking task to avoid Send issues
                let wasm_bytes = wasm_bytes.to_vec();
                let input = input.to_string();
                let config = config.clone();
                let python_pool = python_pool.clone();
                let micropython_executor = micropython_executor.clone();

                let timeout_ms = config.timeout_ms;
                let timeout_duration = std::time::Duration::from_millis(timeout_ms);

                // Try MicroPython executor first if available (uses WASM module linking)
                if let Some(ref mp_exec) = micropython_executor {
                    tracing::debug!("Using MicroPython executor for Python execution");

                    let mp_exec_clone = mp_exec.clone();
                    let blocking_task = tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                        let python_code = String::from_utf8_lossy(&wasm_bytes);
                        mp_exec_clone.execute_with_code(&python_code, &input)
                    });

                    let timeout_future: tokio::time::Timeout<tokio::task::JoinHandle<anyhow::Result<String>>> = tokio::time::timeout(timeout_duration, blocking_task);
                    let timeout_result = timeout_future.await;

                    match timeout_result {
                        Ok(join_result) => {
                            match join_result {
                                Ok(Ok(value)) => Ok::<String, anyhow::Error>(value),
                                Ok(Err(e)) => Err::<String, anyhow::Error>(anyhow::anyhow!("MicroPython execution failed: {}", e)),
                                Err(e) => Err::<String, anyhow::Error>(anyhow::anyhow!("MicroPython blocking task join error: {}", e)),
                            }
                        }
                        Err(_) => Err::<String, anyhow::Error>(anyhow::anyhow!(RuntimeError::timeout(timeout_ms))),
                    }
                } else {
                    // Fall back to RustPython via pool, shared state, or fresh interpreter
                    // If pool is available, acquire before spawning blocking task
                    let pooled_guard = if let Some(ref pool) = python_pool {
                        match pool.acquire().await {
                            Ok(guard) => Some(guard),
                            Err(e) => {
                                tracing::warn!("Failed to acquire Python runtime from pool: {}", e);
                                None
                            }
                        }
                    } else {
                        None
                    };

                    // Clone shared Python state for use in blocking task
                    let python_shared_state = self.python_shared_state.clone();

                    let blocking_task = tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                        // For Python execution, treat the wasm_bytes as direct Python source code
                        let python_code = String::from_utf8_lossy(&wasm_bytes);

                        // Use pooled interpreter if available, otherwise use shared state, otherwise create fresh one
                        if let Some(guard) = pooled_guard {
                            guard.execute_sync(&python_code, &input)
                        } else if let Some(ref state) = python_shared_state {
                            state.execute_sync(&python_code, &input)
                        } else {
                            let engine_config = config.clone();
                            let engine = PythonEngine::new(engine_config)?;
                            engine.execute_sync(&python_code, &input)
                        }
                    });

                    let timeout_future: tokio::time::Timeout<tokio::task::JoinHandle<anyhow::Result<String>>> = tokio::time::timeout(timeout_duration, blocking_task);
                    let timeout_result = timeout_future.await;

                    match timeout_result {
                        Ok(join_result) => {
                            match join_result {
                                Ok(Ok(value)) => Ok::<String, anyhow::Error>(value),
                                Ok(Err(e)) => Err::<String, anyhow::Error>(convert_timeout_error(e, timeout_ms)),
                                Err(e) => Err::<String, anyhow::Error>(anyhow::anyhow!("Python blocking task join error: {}", e)),
                            }
                        }
                        Err(_) => Err::<String, anyhow::Error>(anyhow::anyhow!(RuntimeError::timeout(timeout_ms))),
                    }
                }
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
                // aot_counter is reserved for tracking AOT compilation counts if needed later
                let _aot_counter = self.aot_counter.clone();

                let timeout_ms = config.timeout_ms;
                let timeout_duration = std::time::Duration::from_millis(timeout_ms);

                let blocking_task = tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
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
                            cache.get(&cpython_hash).and_then(|e| {
                                unsafe { Module::deserialize(&engine, &e.compiled) }.ok()
                            })
                        } else {
                            None
                        }
                    } else {
                        None
                    };

                    if let Some(ref linker) = wasi_linker {
                        // Log linker configuration for debugging/auditing
                        let function_key = format!("{}@{}", config.function, config.version);
                        linker.log_configuration(&function_key);

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
                });

                let timeout_future: tokio::time::Timeout<tokio::task::JoinHandle<anyhow::Result<String>>> = tokio::time::timeout(timeout_duration, blocking_task);
                let timeout_result = timeout_future.await;
                
                let output: anyhow::Result<String> = match timeout_result {
                    Ok(join_result) => {
                        match join_result {
                            Ok(Ok(value)) => Ok::<String, anyhow::Error>(value),
                            Ok(Err(e)) => Err::<String, anyhow::Error>(convert_timeout_error(e, timeout_ms)),
                            Err(e) => Err::<String, anyhow::Error>(anyhow::anyhow!("CPython-WASM blocking task join error: {}", e)),
                        }
                    }
                    Err(_) => Err::<String, anyhow::Error>(anyhow::anyhow!(RuntimeError::timeout(timeout_ms))),
                };
                Ok::<anyhow::Result<String>, anyhow::Error>(output).context("Failed to execute CPython-WASM in blocking task")?
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

                let timeout_ms = config.timeout_ms;
                let timeout_duration = std::time::Duration::from_millis(timeout_ms);

                let blocking_task = tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
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
                                cache.get(&wasm_hash).and_then(|e| {
                                    unsafe { Module::deserialize(&engine, &e.compiled) }.ok()
                                })
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
                });

                let timeout_future: tokio::time::Timeout<tokio::task::JoinHandle<anyhow::Result<String>>> = tokio::time::timeout(timeout_duration, blocking_task);
                let timeout_result = timeout_future.await;
                
                let output: anyhow::Result<String> = match timeout_result {
                    Ok(join_result) => {
                        match join_result {
                            Ok(Ok(value)) => Ok::<String, anyhow::Error>(value),
                            Ok(Err(e)) => Err::<String, anyhow::Error>(convert_timeout_error(e, timeout_ms)),
                            Err(e) => Err::<String, anyhow::Error>(anyhow::anyhow!("WASM blocking task join error: {}", e)),
                        }
                    }
                    Err(_) => Err::<String, anyhow::Error>(anyhow::anyhow!(RuntimeError::timeout(timeout_ms))),
                };
                Ok::<anyhow::Result<String>, anyhow::Error>(output).context("Failed to execute WASM in blocking task")?
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
                            let python_code = String::from_utf8_lossy(wasm_bytes);
                            if let Some(ref state) = self.python_shared_state {
                                state.execute_sync(&python_code, input)
                            } else {
                                let engine = PythonEngine::new(config.clone())?;
                                engine.execute_sync(&python_code, input)
                            }
                        } else {
                            // Execute using MicroVM orchestrator with tier-based resource allocation
                            let (memory_mb, vcpus) = Self::get_tier_resources(config);
                            // Deduplicate packages and whitelist entries before forwarding.
                            let packages: Vec<String> = {
                                let mut seen = std::collections::HashSet::new();
                                config.python_packages.iter()
                                    .filter(|p| !p.trim().is_empty() && seen.insert(p.trim().to_lowercase()))
                                    .map(|p| p.trim().to_string())
                                    .collect()
                            };
                            let network_whitelist: Vec<String> = {
                                let mut seen = std::collections::HashSet::new();
                                config.network_whitelist.iter()
                                    .filter(|h| !h.trim().is_empty() && seen.insert(h.trim().to_lowercase()))
                                    .map(|h| h.trim().to_string())
                                    .collect()
                            };
                            let request = MicroVMExecutionRequest {
                                code: String::from_utf8_lossy(wasm_bytes).to_string(),
                                input: input.to_string(),
                                handler: "handler".to_string(),
                                packages,
                                memory_mb,
                                vcpus,
                                timeout_ms: config.timeout_ms,
                                // tenant_id comes from the CLI flag; Go sets it to the real UUID
                                tenant_id: config.tenant_id.clone().unwrap_or_else(|| config.function.clone()),
                                network_whitelist,
                                strict_network_whitelist: config.strict_network_whitelist,
                                package_caching_enabled: config.package_caching_enabled,
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
                                    let python_code = String::from_utf8_lossy(wasm_bytes);
                                    if let Some(ref state) = self.python_shared_state {
                                        state.execute_sync(&python_code, input)
                                    } else {
                                        let engine = PythonEngine::new(config.clone())?;
                                        engine.execute_sync(&python_code, input)
                                    }
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
                        let python_code = String::from_utf8_lossy(wasm_bytes);
                        if let Some(ref state) = self.python_shared_state {
                            state.execute_sync(&python_code, input)
                        } else {
                            let engine = PythonEngine::new(config.clone())?;
                            engine.execute_sync(&python_code, input)
                        }
                    }
                }
            }
        }
    }

    /// Detect the runtime type of a WASM module
    pub fn detect_runtime_type(&self, wasm_bytes: &[u8]) -> RuntimeType {
        // Check if it's a Python WASM module
        if PythonRuntime::is_python_code(wasm_bytes) {
            // Use RuntimeType::from_str for explicit runtime selection from config
            if let Some(runtime) = RuntimeType::from_str(&self.config.runtime) {
                // Explicit manifest/runtime selection from config
                if runtime.requires_microvm() && self.orchestrator_client.is_some() {
                    return runtime;
                }
                if runtime == RuntimeType::PythonWasm && self.config.supports_cpython_wasm() {
                    return runtime;
                }
                if runtime == RuntimeType::Python {
                    return runtime;
                }
            }
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
            .map_err(|e| anyhow::anyhow!("AOT: failed to compile Wasm module: {}", e))?;

        // Serialize to portable compiled bytes
        let compiled = module.serialize()
            .map_err(|e| anyhow::anyhow!("AOT: failed to serialize compiled module: {}", e))?;

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
                        .map_err(|e| anyhow::anyhow!("AOT: failed to deserialize cached module: {}", e))?;
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
            .map_err(|e| anyhow::anyhow!("AOT: failed to deserialize freshly compiled module: {}", e))?;
        Ok(module)
    }

    /// Get resource allocation based on budget tier
    fn get_tier_resources(config: &Config) -> (u32, u32) {
        use crate::budget::{BudgetTier, NodeSpecs};

        let tier = config.get_budget_tier();
        let specs = NodeSpecs::for_tier(&tier);

        // Allocate resources based on tier using NodeSpecs
        match tier {
            BudgetTier::UltraLow => (specs.max_memory_per_fn_mb as u32, 1),
            BudgetTier::Low => (specs.max_memory_per_fn_mb as u32, 1),
            BudgetTier::Medium => (specs.max_memory_per_fn_mb as u32, specs.vcpu as u32),
            BudgetTier::High => (specs.max_memory_per_fn_mb as u32, specs.vcpu as u32),
        }
    }

    /// Create a `PooledWasmInstance` from a compiled module and WASI context.
    ///
    /// This is used to pre-warm the pool with compiled instances.
    pub fn create_pooled_instance(
        &self,
        module: Module,
        wasi_ctx: WasiP1Ctx,
        function_key: String,
    ) -> PooledWasmInstance {
        let pipe_capacity = if self.config.max_output_bytes > 0 {
            self.config.max_output_bytes
        } else {
            1024 * 1024
        };
        PooledWasmInstance::new(module, wasi_ctx, function_key, pipe_capacity)
    }

    /// Compile module and create a pooled instance for pre-warming.
    ///
    /// Returns the compiled module (for reuse) and the pooled instance.
    #[allow(dead_code)]
    pub fn compile_and_create_pooled_instance(
        &self,
        wasm_bytes: &[u8],
        function_key: String,
    ) -> anyhow::Result<(Module, PooledWasmInstance)> {
        let module = self.get_or_compile_module(wasm_bytes)?;

        // Create WASI context for the pooled instance
        let wasi_ctx = WasiContext::new_with_input(&self.config, function_key.clone(), "")?;
        let pooled = self.create_pooled_instance(
            module.clone(),
            wasi_ctx.ctx,
            function_key,
        );

        Ok((module, pooled))
    }

    /// Get a WASI context for the given function key.
    ///
    /// This creates a fresh WASI context. For pooled execution, use the
    /// pool's `PooledWasmInstanceGuard` instead.
    pub fn create_wasi_context(
        &self,
        function_key: &str,
        input: &str,
    ) -> anyhow::Result<WasiContext> {
        WasiContext::new_with_input(&self.config, function_key.to_string(), input)
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
    ) -> wasmtime::Result<bool> {
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
    ) -> wasmtime::Result<bool> {
        Ok(true)
    }
}
// Thread-local storage for the per-execution memory limiter.
//
// Because `Store<WasiP1Ctx>` uses an opaque data type that we cannot extend,
// we store the limiter in a thread-local and hand out `&mut` references to it
// through `store.limiter()`.  Each `spawn_blocking` task gets its own OS
// thread, so this is safe for concurrent execution.
std::thread_local! {
    static MEMORY_LIMITER: std::cell::RefCell<Option<FunctionMemoryLimiter>> =
        const { std::cell::RefCell::new(None) };
}

/// Guard that clears the thread-local limiter on drop.
struct LimiterGuard;

impl Drop for LimiterGuard {
    fn drop(&mut self) {
        MEMORY_LIMITER.with(|cell| {
            *cell.borrow_mut() = None;
        });
    }
}

/// Install the memory limiter for the current thread and return a guard that
/// clears it when the execution is done.
fn install_memory_limiter(memory_mb: u32) -> LimiterGuard {
    let max_bytes = (memory_mb as usize) * 1024 * 1024;
    MEMORY_LIMITER.with(|cell| {
        *cell.borrow_mut() = Some(FunctionMemoryLimiter { max_bytes });
    });
    LimiterGuard
}

/// Synchronous WASI execution function for use in spawn_blocking.
/// Accepts an optional pre-compiled module (from the AOT cache) to skip re-compilation.
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

    // Install the hard memory limiter via the thread-local mechanism.
    // The `LimiterGuard` clears the thread-local on drop so subsequent
    // executions on the same thread start clean.
    let _limiter_guard = install_memory_limiter(config.memory_mb);
    store.limiter(|_data| {
        MEMORY_LIMITER.with(|cell| {
            // Safety: the limiter is always Some while the guard is alive, and
            // Wasmtime only calls this closure while the Store (and thus the
            // guard on the same stack frame) is alive.
            let ptr = cell.as_ptr();
            unsafe { &mut *ptr }
                .as_mut()
                .expect("MEMORY_LIMITER must be set before Store is used")
        })
    });

    // Calibrated fuel metering: prefer timeout_ms × fuel_per_ms, else cpu_ms_limit × fuel_per_ms, else cpu_fuel_limit.
    let fuel_limit = if config.fuel_for_timeout() > 0 {
        config.fuel_for_timeout()
    } else if config.cpu_ms_limit > 0 && config.fuel_per_ms > 0 {
        config.cpu_ms_limit.saturating_mul(config.fuel_per_ms)
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

    // Compile module (or use pre-compiled AOT module)
    let module = if let Some(m) = precompiled {
        m
    } else {
        Module::new(engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!(RuntimeError::wasm_compilation(e.to_string())))?
    };

    // Instantiate module with WASI
    let instance = linker.linker().instantiate(&mut store, &module)
        .map_err(|e| anyhow::anyhow!(RuntimeError::wasm_instantiation(e.to_string())))?;

    // Execute the function. Prefer handler when we have input so Python WASM (and other
    // handler-based modules) receive input and can return a value via memory; otherwise try _start/main.
    let handler_result_ptr: Option<i32> = if let Ok(func) = instance.get_typed_func::<(i32, i32), i32>(&mut store, "handler") {
        if let Some(memory) = instance.get_memory(&mut store, "memory") {
            use crate::wasm_interface::memory;

            let input_ptr = memory::write_string(&memory, &mut store, input)?;
            let input_len = input.len() as i32;

            let result_ptr = func.call(&mut store, (input_ptr, input_len)).map_err(|e| {
                anyhow::anyhow!(RuntimeError::wasm_execution(format!("Handler function failed: {}", e)))
            })?;
            tracing::info!("Handler function returned: {}", result_ptr);
            Some(result_ptr)
        } else {
            return Err(anyhow::anyhow!("No memory export found for handler function"));
        }
    } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "_start") {
        func.call(&mut store, ()).map_err(|e| anyhow::anyhow!(RuntimeError::wasm_execution(format!("_start function failed: {}", e))))?;
        None
    } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "main") {
        func.call(&mut store, ()).map_err(|e| anyhow::anyhow!(RuntimeError::wasm_execution(format!("main function failed: {}", e))))?;
        None
    } else {
        return Err(anyhow::anyhow!(RuntimeError::function_not_found("handler, _start, or main")));
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

    // Fall back to stdout/stderr.
    // Phase 2.2: detect truncation and return an explicit error instead of silently
    // dropping bytes.  `MemoryOutputPipe` stops accepting bytes once its capacity is
    // reached; we detect this by comparing the pipe's byte count to the configured
    // limit.
    let stdout = stdout_pipe.contents();
    let stderr = stderr_pipe.contents();

    drop(store);
    let _ = instance;
    drop(module);

    let pipe_capacity = if config.max_output_bytes > 0 {
        config.max_output_bytes
    } else {
        1024 * 1024
    };
    if stdout.len() >= pipe_capacity {
        return Err(anyhow::anyhow!(
            "Function output was truncated: stdout reached the {} byte limit. \
             Increase --max-output-bytes to capture more output.",
            pipe_capacity
        ));
    }
    if stderr.len() >= pipe_capacity {
        return Err(anyhow::anyhow!(
            "Function output was truncated: stderr reached the {} byte limit. \
             Increase --max-output-bytes to capture more output.",
            pipe_capacity
        ));
    }

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

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::runtime::Runtime;

    /// Helper: create a `WasmEngine` from a `Config`.
    fn make_engine(config: Config) -> WasmEngine {
        let logger = crate::logging::init_structured_logging(false);
        let security_monitor = Arc::new(crate::security::SecurityMonitor::new());
        WasmEngine::with_config(config, None, logger, None, security_monitor, None).unwrap()
    }

    #[test]
    fn test_wasm_engine_creation_without_wasi() {
        let config = Config {
            wasi_enabled: false,
            ..Config::default()
        };
        let engine = make_engine(config);
        assert!(engine.wasi_linker.is_none());
    }

    #[test]
    fn test_wasm_engine_creation_with_wasi() {
        let config = Config {
            wasi_enabled: true,
            ..Config::default()
        };
        let engine = make_engine(config);
        assert!(engine.wasi_linker.is_some());
    }

    #[test]
    fn test_execute_minimal_module_without_wasi() {
        let rt = Runtime::new().unwrap();
        rt.block_on(async {
            let config = Config {
                wasi_enabled: false,
                ..Config::default()
            };
            let wasm_bytes = wat::parse_str(r#"
                (module
                    (func (export "main"))
                )
            "#).unwrap();

            let engine = make_engine(config.clone());
            let result = engine.execute(&wasm_bytes, "test", &config, None, None).await;
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

            let engine = make_engine(config.clone());
            let test_input = "hello world";
            let result = engine.execute(&wasm_bytes, test_input, &config, None, None).await;
            assert!(result.is_ok());
        });
    }
}
