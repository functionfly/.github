//! Wasmtime engine for executing WebAssembly functions.

use std::sync::Arc;

use anyhow::Context;
use clap::Parser;
use sha2::{Digest, Sha256};
use tracing::{info, warn};
use wasmtime::*;

use crate::config::Config;
use crate::errors::RuntimeError;
use crate::kv::SharedKVStore;
use crate::logging::StructuredLogger;
use crate::orchestrator_client::{OrchestratorClient, MicroVMExecutionRequest};
use crate::pool::{PooledWasmInstance, PoolManager};
use crate::python::engine::{PythonEngine, PythonSharedState};
use crate::python_pool::PythonRuntimePool;
use crate::wasi::{WasiContext, WasiLinker};

// Import submodules
mod runtime_type;
mod shared_state;
mod aot_cache;
mod memory_limiter;
mod execution;
pub mod graph;
pub mod wasm_cell;
pub mod sar_executor;

// Re-export public types
pub use runtime_type::RuntimeType;
pub use shared_state::SharedState;
pub use aot_cache::{AotCache, AotCacheEntry};
pub use memory_limiter::{FunctionMemoryLimiter, install_memory_limiter, with_limiter, LimiterGuard};
pub use execution::{execute_wasi_sync_inner, execute_wasi_with_module_and_store};
pub use graph::{
    Graph, Node, NodeId, NodeType, Edge, EdgeType, GraphExecutor, GraphExecutionInput,
    GraphExecutionResult, ExecutionContext, ExecutionStatus, ExecutionPriority,
    NodeExecutor, NodeExecutionError, NodeResult, DefaultNodeExecutor,
    LlmTrafficType, MemoryOp, ControlKind, Expr, OptStrategy, RetryPolicy,
};
pub use wasm_cell::{
    WasmCell, WasmCellPool, WasmCellGuard, WasmCellExecutor, CellConfig,
    CellPoolStats, HostFunction,
};
pub use sar_executor::SarNodeExecutor;

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

    /// AOT compilation cache
    aot_cache: AotCache,
    /// Pool manager for warm instance reuse across executions
    pool_manager: std::sync::Mutex<Option<Arc<PoolManager>>>,
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
            aot_cache: AotCache::new(),
            pool_manager: std::sync::Mutex::new(None),
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
                        if let Some(mut guard) = pooled_guard {
                            match guard.execute_sync(&python_code, &input) {
                                Ok(result) => Ok(result),
                                Err(e) => {
                                    // Runtime is already marked dirty by execute_sync
                                    Err(e)
                                }
                            }
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

                    let aot_cache = AotCache::new();
                    let precompiled = aot_cache.load_precompiled(&engine, &cpython_hash, &config)?;

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
                // Try pooled execution first if pooling is enabled
                let function_key = format!("{}@{}", config.function, config.version);

                if let Some(ref pool_manager) = self.pool_manager() {
                    // Attempt to acquire from pool
                    match pool_manager.acquire(&function_key).await {
                        Ok(mut guard) => {
                            // Pooled execution path
                            let engine = self.engine.clone();
                            let timeout_ms = config.timeout_ms;
                            let timeout_duration = std::time::Duration::from_millis(timeout_ms);
                            let input_for_closure = input.to_string();
                            let config_for_closure = config.clone();
                            let wasi_linker = self.wasi_linker.clone();

                            // Clone for fallback since references will be captured
                            let input_fallback = input.to_string();
                            let config_fallback = config.clone();

                            let blocking_task = tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                                let instance = guard.instance_mut();
                                instance.reset_for_execution(&input_for_closure);

                                let mut store = instance.create_store(&engine);

                                // execute_wasi_with_module_and_store applies the hard memory cap
                                // via Store::limiter() using FunctionMemoryLimiter.

                                // Set fuel for CPU-based timeout enforcement (fuel_for_timeout
                                // returns calibrated fuel based on timeout_ms, cpu_ms_limit,
                                // or cpu_fuel_limit). Wall-clock enforcement is done by the
                                // outer tokio::time::timeout() around the spawn_blocking call.
                                let fuel_limit = config_for_closure.fuel_for_timeout();
                                if fuel_limit > 0 {
                                    store.set_fuel(fuel_limit)?;
                                }

                                // Execute using the WASI linker if available
                                if let Some(ref linker) = wasi_linker {
                                    let module = instance.module.as_ref();
                                    crate::engine::execution::execute_wasi_with_module_and_store(
                                        linker,
                                        module,
                                        &mut store,
                                        &config_for_closure,
                                        &instance.wasi_ctx,
                                    )
                                } else {
                                    Err(anyhow::anyhow!("WASI linker not available for pooled execution"))
                                }
                            });

                            // Try the pooled execution, fall back to standard on error
                            match tokio::time::timeout(timeout_duration, blocking_task).await {
                                Ok(Ok(Ok(result))) => Ok(result),
                                Ok(Ok(Err(e))) => {
                                    tracing::debug!("Pooled execution error: {}, using standard path", e);
                                    self.execute_wasm_standard(wasm_bytes, &input_fallback, &config_fallback).await
                                }
                                Ok(Err(e)) => {
                                    tracing::debug!("Pooled execution task error: {}, using standard path", e);
                                    self.execute_wasm_standard(wasm_bytes, &input_fallback, &config_fallback).await
                                }
                                Err(_) => {
                                    tracing::debug!("Pooled execution timed out, using standard path");
                                    self.execute_wasm_standard(wasm_bytes, &input_fallback, &config_fallback).await
                                }
                            }
                        }
                        Err(e) => {
                            tracing::debug!("Could not acquire from pool: {}, using standard path", e);
                            self.execute_wasm_standard(wasm_bytes, input, config).await
                        }
                    }
                } else {
                    self.execute_wasm_standard(wasm_bytes, input, config).await
                }
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
        use crate::python::runtime::PythonRuntime;

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

    /// Set the pool manager for warm instance reuse.
    ///
    /// This should be called once during engine initialization after the
    /// PoolManager is created.
    pub fn set_pool_manager(&self, pool_manager: Arc<PoolManager>) {
        if let Ok(mut guard) = self.pool_manager.lock() {
            *guard = Some(pool_manager);
            tracing::info!("Pool manager attached to WasmEngine");
        }
    }

    /// Get the pool manager if one is attached.
    pub fn pool_manager(&self) -> Option<Arc<PoolManager>> {
        self.pool_manager.lock().ok().and_then(|g| g.clone())
    }

    /// Create a `PooledWasmInstance` from a compiled module and WASI context.
    ///
    /// This is used to pre-warm the pool with compiled instances.
    pub fn create_pooled_instance(
        &self,
        module: Module,
        wasi_ctx: wasmtime_wasi::p1::WasiP1Ctx,
        function_key: String,
    ) -> PooledWasmInstance {
        let pipe_capacity = if self.config.max_output_bytes > 0 {
            self.config.max_output_bytes
        } else {
            1024 * 1024
        };
        PooledWasmInstance::new(module, wasi_ctx, function_key, pipe_capacity)
    }

    /// Execute WASM using the standard non-pooled path.
    ///
    /// This is the fallback when pooled execution is not available or fails.
    async fn execute_wasm_standard(
        &self,
        wasm_bytes: &[u8],
        input: &str,
        config: &Config,
    ) -> anyhow::Result<String> {
        let wasm_bytes = wasm_bytes.to_vec();
        let input = input.to_string();
        let engine = self.engine.clone();
        let config = config.clone();
        let wasi_linker = self.wasi_linker.clone();
        let aot_cache = self.aot_cache.clone();

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
                let precompiled: Option<Module> = aot_cache.load_precompiled(&engine, &wasm_hash, &config)?;

                // If not cached, compile and store
                let precompiled = if precompiled.is_none() && config.aot_cache_enabled {
                    match Module::new(&engine, &wasm_bytes) {
                        Ok(module) => {
                            if let Ok(compiled) = module.serialize() {
                                let size = compiled.len();
                                let counter = aot_cache.counter.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
                                if let Ok(mut cache) = aot_cache.cache.write() {
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

        let timeout_result = tokio::time::timeout(timeout_duration, blocking_task).await;

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

    /// Compile module and create a pooled instance for pre-warming.
    ///
    /// Returns the compiled module (for reuse) and the pooled instance.
    #[allow(dead_code)]
    pub fn compile_and_create_pooled_instance(
        &self,
        wasm_bytes: &[u8],
        function_key: String,
    ) -> anyhow::Result<(Module, PooledWasmInstance)> {
        let module = self.aot_cache.get_or_compile_module(&self.engine, wasm_bytes, &self.config)?;

        // Create WASI context for the pooled instance
        let wasi_ctx = WasiContext::new_with_input(&self.config, function_key.clone(), "")?;
        let pooled = self.create_pooled_instance(
            module.clone(),
            wasi_ctx.ctx,
            function_key,
        );

        Ok((module, pooled))
    }

    /// Pre-warm the pool for a function after successful execution.
    ///
    /// Compiles the module (blocking) and adds it to the pool so subsequent
    /// requests use warm pooled instances instead of paying cold-start cost.
    ///
    /// Called from `execute_function_daemon` after `engine.execute()` succeeds
    /// via the fallback (non-pooled) path.
    pub async fn prewarm_pool(
        &self,
        wasm_bytes: &[u8],
        function_key: &str,
        pool_manager: Arc<PoolManager>,
    ) -> anyhow::Result<()> {
        let wasm_bytes = wasm_bytes.to_vec();
        let function_key_str = function_key.to_string();
        let engine = self.engine.clone();
        let config = self.config.clone();
        let aot_cache = self.aot_cache.clone();

        // Do blocking work (module compilation) in spawn_blocking
        let (module, wasi_ctx) = tokio::task::spawn_blocking(move || {
            // Use AOT cache to get or compile the module
            let module = aot_cache.get_or_compile_module(&engine, &wasm_bytes, &config)?;

            // Create a basic WASI context for the pooled instance
            let wasi_ctx = WasiContext::new_with_input(&config, function_key_str.clone(), "")?;
            Ok::<_, anyhow::Error>((module, wasi_ctx))
        }).await??;

        // Call async prewarm_instance in the async context
        pool_manager.prewarm_instance(&function_key, module, wasi_ctx.ctx).await;
        Ok(())
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

#[cfg(test)]
mod tests {
    use super::*;

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

    #[tokio::test]
    async fn test_execute_minimal_module_without_wasi() {
        let config = Config {
            wasi_enabled: true,  // WASM execution requires WASI linker
            ..Config::default()
        };
        let wasm_bytes = wat::parse_str(r#"
            (module
                (func (export "main"))
            )
        "#).unwrap();

        let engine = make_engine(config.clone());
        let result = engine.execute(&wasm_bytes, "test", &config, None, None).await;
        assert!(result.is_ok(), "WASM execution failed: {:?}", result);
    }

    #[tokio::test]
    async fn test_handler_function_input_marshaling() {
        let config = Config {
            runtime: "wasm".to_string(),
            wasi_enabled: true,
            ..Config::default()
        };

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
        assert!(result.is_ok(), "Handler execution failed: {:?}", result);
    }

    #[test]
    fn test_runtime_type_from_str() {
        assert_eq!(RuntimeType::from_str("wasm"), Some(RuntimeType::Wasm));
        assert_eq!(RuntimeType::from_str("python"), Some(RuntimeType::Python));
        assert_eq!(RuntimeType::from_str("python-wasm"), Some(RuntimeType::PythonWasm));
        assert_eq!(RuntimeType::from_str("python-microvm"), Some(RuntimeType::PythonMicroVM));
        assert_eq!(RuntimeType::from_str("unknown"), None);
        assert_eq!(RuntimeType::from_str(""), None);
    }

    #[test]
    fn test_runtime_type_requires_microvm() {
        assert!(!RuntimeType::Wasm.requires_microvm());
        assert!(!RuntimeType::Python.requires_microvm());
        assert!(!RuntimeType::PythonWasm.requires_microvm());
        assert!(RuntimeType::PythonMicroVM.requires_microvm());
    }

    #[test]
    fn test_runtime_type_display_name() {
        assert_eq!(RuntimeType::Wasm.display_name(), "WebAssembly");
        assert_eq!(RuntimeType::Python.display_name(), "RustPython");
        assert_eq!(RuntimeType::PythonWasm.display_name(), "CPython-WASM");
        assert_eq!(RuntimeType::PythonMicroVM.display_name(), "CPython (MicroVM)");
    }

    #[test]
    fn test_is_epoch_deadline_error() {
        assert!(is_epoch_deadline_error("deadline exceeded"));
        assert!(is_epoch_deadline_error("epoch deadline"));
        assert!(is_epoch_deadline_error("interruption"));
        assert!(!is_epoch_deadline_error("out of memory"));
        assert!(!is_epoch_deadline_error("invalid argument"));
    }

    #[test]
    fn test_convert_timeout_error() {
        let timeout_ms = 5000;

        // Test deadline error conversion
        let deadline_err = anyhow::anyhow!("deadline exceeded");
        let converted = convert_timeout_error(deadline_err, timeout_ms);
        assert!(converted.to_string().contains("5000"));

        // Test non-timeout error passes through
        let other_err = anyhow::anyhow!("some other error");
        let result = convert_timeout_error(other_err, timeout_ms);
        assert!(result.to_string().contains("some other error"));
    }

    #[test]
    fn test_wasm_hash() {
        let bytes = b"hello world";
        let hash1 = WasmEngine::wasm_hash(bytes);
        let hash2 = WasmEngine::wasm_hash(bytes);

        // Same input should produce same hash
        assert_eq!(hash1, hash2);

        // Hash should be 64 hex characters (SHA-256)
        assert_eq!(hash1.len(), 64);

        // Different input should produce different hash
        let different_hash = WasmEngine::wasm_hash(b"different");
        assert_ne!(hash1, different_hash);
    }

    #[test]
    fn test_get_tier_resources() {
        // Test UltraLow tier
        let config_ultra_low = Config {
            tier: "ultra-low".to_string(),
            ..Config::default()
        };
        let (memory, vcpus) = WasmEngine::get_tier_resources(&config_ultra_low);
        assert_eq!(vcpus, 1); // UltraLow uses 1 vCPU
        assert_eq!(memory, 64); // UltraLow gets 64MB max memory per function

        // Test Low tier
        let config_low = Config {
            tier: "low".to_string(),
            ..Config::default()
        };
        let (memory, vcpus) = WasmEngine::get_tier_resources(&config_low);
        assert_eq!(vcpus, 1);
        assert_eq!(memory, 128);

        // Test Medium tier
        let config_medium = Config {
            tier: "medium".to_string(),
            ..Config::default()
        };
        let (memory, vcpus) = WasmEngine::get_tier_resources(&config_medium);
        assert_eq!(vcpus, 8);
        assert_eq!(memory, 256);

        // Test High tier
        let config_high = Config {
            tier: "high".to_string(),
            ..Config::default()
        };
        let (memory, vcpus) = WasmEngine::get_tier_resources(&config_high);
        assert_eq!(vcpus, 16);
        assert_eq!(memory, 512);
    }

    #[test]
    fn test_wasm_engine_wasi_linker_optional() {
        // Test with WASI disabled
        let config_no_wasi = Config {
            wasi_enabled: false,
            ..Config::default()
        };
        let engine = make_engine(config_no_wasi);
        assert!(engine.wasi_linker().is_none());

        // Test with WASI enabled
        let config_with_wasi = Config {
            wasi_enabled: true,
            ..Config::default()
        };
        let engine_with_wasi = make_engine(config_with_wasi);
        assert!(engine_with_wasi.wasi_linker().is_some());
    }
}
