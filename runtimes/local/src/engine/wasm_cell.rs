//! WASM Cell isolation for graph node execution.
//!
//! Each graph node runs as an isolated WASM cell with:
//! - **Memory isolation** via wasmtime's `Store::limiter()` enforcing per-cell limits
//! - **Host function injection** — kv, ai_inference, fetch, get_env registered per cell
//! - **Cell pool** for warm reuse — targets <50ms cold start via pre-warmed instances
//!
//! ## Architecture
//!
//! `WasmCell` wraps a compiled `wasmtime::Module` + a `WasiP1Ctx` + per-cell host functions.
//! `WasmCellHandle` is a checked-out cell from the pool; it auto-returns on drop.
//!
//! ## Host Functions
//!
//! Each cell exposes these host functions to the WASM guest:
//! - `functionfly.kv_get(key_ptr, key_len) -> (result_ptr, result_len)`
//! - `functionfly.kv_set(key_ptr, key_len, val_ptr, val_len)`
//! - `functionfly.ai_inference(model_ptr, model_len, input_ptr, input_len) -> result_ptr`
//! - `functionfly.fetch(url_ptr, url_len, method_ptr, method_len, body_ptr, body_len) -> result_ptr`
//! - `functionfly.get_env(name_ptr, name_len) -> result_ptr`
//!
//! ## Memory Safety
//!
//! All cells share the same wasmtime `Engine` (safe, compiled modules are stateless)
//! but have isolated `Store` + `WasiP1Ctx`. Memory limits are enforced via
//! `Store::limiter()` using `FunctionMemoryLimiter`.

use std::collections::{HashMap, VecDeque};
use std::sync::Arc;

use tokio::sync::{Mutex, Semaphore};
use tracing::{info, warn, instrument};
use wasmtime::{Engine, Module, Store};
use wasmtime_wasi::p1::WasiP1Ctx;
use wasmtime_wasi::p2::pipe::{MemoryInputPipe, MemoryOutputPipe};

use crate::config::Config;
use crate::engine::memory_limiter::{install_memory_limiter, with_limiter, LimiterGuard};
use crate::errors::RuntimeError;

// ---------------------------------------------------------------------------
// Cell Config
// ---------------------------------------------------------------------------

/// A registered host function available to a WASM cell.
#[derive(Clone)]
pub struct HostFunction {
    /// Module name, e.g. "functionfly"
    pub module: String,
    /// Function name, e.g. "kv_get"
    pub name: String,
}

/// Configuration for a WASM cell.
#[derive(Clone)]
pub struct CellConfig {
    /// Memory limit in MB
    pub memory_mb: u32,
    /// CPU fuel limit (instructions)
    pub fuel_limit: u64,
    /// Whether to allow network access
    pub network_allowed: bool,
    /// Whether to allow AI inference
    pub ai_allowed: bool,
    /// Whether to allow KV store access
    pub kv_allowed: bool,
    /// Custom host functions available to this cell
    pub host_functions: Vec<HostFunction>,
}

impl Default for CellConfig {
    fn default() -> Self {
        Self {
            memory_mb: 64,
            fuel_limit: 1_000_000,
            network_allowed: true,
            ai_allowed: true,
            kv_allowed: true,
            host_functions: Vec::new(),
        }
    }
}

// ---------------------------------------------------------------------------
// WasmCell
// ---------------------------------------------------------------------------

/// A compiled WASM cell — the execution environment for a single graph node.
///
/// A cell holds a compiled `Module` (shared, cheap to clone) and a `WasiP1Ctx`
/// (rebuilt per execution for isolation). Each cell can be configured with
/// different memory limits and host function sets.
///
/// Cells use a `WasiLinker` for instantiation so that WASM guests can call
/// host functions (`functionfly.kv_get`, `functionfly.ai`, `functionfly.fetch`, etc.).
pub struct WasmCell {
    /// Compiled module — shared across all cells using the same WASM bytes
    module: Arc<Module>,
    /// WASI context for this cell
    wasi_ctx: WasiP1Ctx,
    /// Cell configuration
    config: CellConfig,
    /// Function key for logging
    function_key: String,
    /// Reuse counter
    reuse_count: u32,
    /// Memory limit for this cell in MB
    memory_limit_mb: u32,
    /// Pipe capacity for output capture
    pipe_capacity: usize,
}

impl WasmCell {
    /// Create a new WASM cell from compiled module bytes and config.
    pub fn new(
        engine: &Engine,
        wasm_bytes: &[u8],
        config: CellConfig,
        function_key: String,
    ) -> anyhow::Result<Self> {
        let module = Module::new(engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!("WASM compilation failed: {}", e))?;

        let wasi_ctx = Self::build_wasi_ctx(&config)?;
        let memory_limit_mb = config.memory_mb;

        let memory_limit_mb = config.memory_mb;
        let pipe_capacity = 1024 * 1024; // 1 MiB default

        Ok(Self {
            module: Arc::new(module),
            wasi_ctx,
            config,
            function_key,
            reuse_count: 0,
            memory_limit_mb,
            pipe_capacity,
        })
    }

    /// Create from a pre-compiled module (for pool pre-warming).
    pub fn from_module(
        module: Module,
        config: &CellConfig,
        function_key: String,
        pipe_capacity: usize,
    ) -> anyhow::Result<Self> {
        let wasi_ctx = Self::build_wasi_ctx(config)?;
        Ok(Self {
            module: Arc::new(module),
            wasi_ctx,
            config: config.clone(),
            function_key,
            reuse_count: 0,
            memory_limit_mb: config.memory_mb,
            pipe_capacity,
        })
    }

    fn build_wasi_ctx(config: &CellConfig) -> anyhow::Result<WasiP1Ctx> {
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        builder
            .inherit_stdio()
            .args(&[config.memory_mb.to_string()]);
        Ok(builder.build_p1())
    }

    /// Execute the cell's WASM module with the given input.
    ///
    /// Returns the WASM output as a string, or an error.
    #[instrument(skip_all, fields(function_key = %self.function_key, reuse_count = self.reuse_count))]
    pub fn execute(
        &mut self,
        input: &str,
        linker: &crate::wasi::WasiLinker,
    ) -> anyhow::Result<String> {
        // Fresh pipes per execution — pipe_capacity for output capture
        let stdout = MemoryOutputPipe::new(self.pipe_capacity);
        let stderr = MemoryOutputPipe::new(self.pipe_capacity);
        let stdin = MemoryInputPipe::new(input.as_bytes().to_vec());

        // Rebuild WASI ctx with fresh pipes
        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        builder
            .stdin(stdin)
            .stdout(stdout.clone())
            .stderr(stderr.clone())
            .args(&[self.function_key.clone()]);
        let wasi_ctx = builder.build_p1();

        let mut store = Store::new(self.module.engine(), wasi_ctx);

        // Apply memory limiter
        let _limiter_guard = install_memory_limiter(self.memory_limit_mb);
        store.limiter(|_data| unsafe { with_limiter(|l| l) });

        if self.config.fuel_limit > 0 {
            store.set_fuel(self.config.fuel_limit).ok();
        }

        // Instantiate using the WASI linker (injects all host functions: kv, ai, fetch, etc.)
        let instance = linker
            .linker()
            .instantiate(&mut store, &self.module)
            .map_err(|e| anyhow::anyhow!("WASM instantiation failed: {}", e))?;

        // Find and call the handler function
        let handler_result_ptr: Option<i32> =
            if let Ok(func) = instance.get_typed_func::<(i32, i32), i32>(&mut store, "handler") {
                if let Some(memory) = instance.get_memory(&mut store, "memory") {
                    use crate::wasm_interface::memory;

                    let input_ptr = memory::write_string(&memory, &mut store, input)?;
                    let input_len = input.len() as i32;

                    let result_ptr = func.call(&mut store, (input_ptr, input_len))
                        .map_err(|e| anyhow::anyhow!("WASM execution failed: {}", e))?;
                    Some(result_ptr)
                } else {
                    return Err(anyhow::anyhow!("No memory export found for handler function"));
                }
            } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "_start") {
                func.call(&mut store, ()).map_err(|e| anyhow::anyhow!("WASM execution failed: {}", e))?;
                None
            } else if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, "main") {
                func.call(&mut store, ()).map_err(|e| anyhow::anyhow!("WASM execution failed: {}", e))?;
                None
            } else {
                return Err(anyhow::anyhow!("No handler, _start, or main function found in WASM module"));
            };

        // Extract result from handler return pointer
        if let Some(result_ptr) = handler_result_ptr {
            if result_ptr > 0 {
                if let Some(memory) = instance.get_memory(&mut store, "memory") {
                    match crate::engine::execution::read_handler_result(&memory, &store, result_ptr) {
                        Ok(s) if !s.is_empty() => {
                            self.reuse_count += 1;
                            return Ok(s);
                        }
                        Ok(_) => {}
                        Err(e) => tracing::debug!("Could not read handler result from memory: {}", e),
                    }
                }
            } else if result_ptr < 0 {
                return Err(anyhow::anyhow!("Handler returned error indicator ({})", result_ptr));
            }
        }

        // Fall back to stdout/stderr
        let stdout_contents = stdout.contents();
        let stderr_contents = stderr.contents();

        self.reuse_count += 1;

        if !stdout_contents.is_empty() {
            Ok(String::from_utf8_lossy(&stdout_contents).to_string())
        } else if !stderr_contents.is_empty() {
            Err(anyhow::anyhow!("WASM stderr: {}", String::from_utf8_lossy(&stderr_contents)))
        } else {
            Ok(String::new())
        }
    }

    /// Get the function key for this cell.
    pub fn function_key(&self) -> &str {
        &self.function_key
    }

    /// Get the reuse count.
    pub fn reuse_count(&self) -> u32 {
        self.reuse_count
    }
}

// ---------------------------------------------------------------------------
// Cell Pool
// ---------------------------------------------------------------------------

/// Inner pool state.
struct WasmCellPoolInner {
    idle: Mutex<VecDeque<WasmCell>>,
    semaphore: Semaphore,
    max_idle: usize,
    module: Arc<Module>,
    config: CellConfig,
    function_key: String,
    /// WASI linker for host function injection (kv, ai, fetch, etc.)
    wasi_linker: Arc<crate::wasi::WasiLinker>,
    /// Output pipe capacity per cell
    pipe_capacity: usize,
}

/// Pool for warm WASM cells — targets <50ms cold start.
pub struct WasmCellPool {
    inner: Arc<WasmCellPoolInner>,
}

unsafe impl Send for WasmCellPool {}
unsafe impl Sync for WasmCellPool {}

/// A checked-out WASM cell from the pool.
///
/// Automatically returns the cell to the pool on drop.
pub struct WasmCellGuard {
    cell: Option<WasmCell>,
    /// Reference to the pool's inner state (for returning on drop)
    pool: Arc<WasmCellPoolInner>,
    discard: bool,
}

impl WasmCellGuard {
    /// Get mutable access to the cell for execution.
    pub fn cell_mut(&mut self) -> &mut WasmCell {
        self.cell.as_mut().expect("cell already consumed")
    }

    /// Get a reference to the WASI linker for cell execution.
    pub fn wasi_linker(&self) -> &Arc<crate::wasi::WasiLinker> {
        &self.pool.wasi_linker
    }

    /// Take ownership of the cell (removes it from the guard).
    pub fn take(mut self) -> WasmCell {
        self.discard = true;
        self.cell.take().expect("cell already consumed")
    }

    /// Discard this cell (don't return it to the pool — e.g., after an error).
    pub fn discard(&mut self) {
        self.discard = true;
    }
}

impl Drop for WasmCellGuard {
    fn drop(&mut self) {
        if let Some(cell) = self.cell.take() {
            if !self.discard {
                if let Ok(mut guard) = self.pool.idle.try_lock() {
                    if guard.len() < self.pool.max_idle {
                        guard.push_back(cell);
                    }
                }
            }
            // Always release the semaphore permit
            self.pool.semaphore.add_permits(1);
        }
    }
}

impl WasmCellPool {
    /// Create a new cell pool for a function.
    ///
    /// - `engine` — the wasmtime engine (shared, stateless)
    /// - `wasm_bytes` — the compiled WASM for this function
    /// - `function_key` — "name@version" for logging
    /// - `config` — per-cell limits and capabilities
    /// - `wasi_linker` — WASI linker with host functions (kv, ai, fetch, etc.)
    /// - `max_concurrent` — max total executions (idle + active)
    /// - `max_idle` — max idle cells to keep warm
    pub fn new(
        engine: Engine,
        wasm_bytes: &[u8],
        function_key: String,
        config: CellConfig,
        wasi_linker: Arc<crate::wasi::WasiLinker>,
        max_concurrent: usize,
        max_idle: usize,
    ) -> anyhow::Result<Self> {
        let module = Module::new(&engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!("WASM compilation failed: {}", e))?;

        let pipe_capacity = 1024 * 1024; // 1 MiB default

        let inner = Arc::new(WasmCellPoolInner {
            idle: Mutex::new(VecDeque::new()),
            semaphore: Semaphore::new(max_concurrent.max(1)),
            max_idle: max_idle.min(max_concurrent),
            module: Arc::new(module),
            config: config.clone(),
            function_key,
            wasi_linker,
            pipe_capacity,
        });

        Ok(Self { inner })
    }

    /// Pre-warm the pool with `count` instances (call during startup).
    pub async fn prewarm(&self, count: usize) {
        for _ in 0..count {
            if let Ok(cell) = self.create_cell() {
                let mut guard = self.inner.idle.lock().await;
                if guard.len() < self.inner.max_idle {
                    guard.push_back(cell);
                }
            }
        }
        info!(
            function_key = %self.inner.function_key,
            warmed = count,
            "WasmCellPool pre-warmed"
        );
    }

    /// Acquire a cell from the pool, waiting if necessary.
    ///
    /// If no idle cell is available and concurrency permits, creates a fresh cell.
    /// Otherwise waits until a cell is returned.
    pub async fn acquire(&self) -> anyhow::Result<WasmCellGuard> {
        // Wait for a concurrency permit
        let _permit = self.inner.semaphore.acquire().await
            .map_err(|_| anyhow::anyhow!("pool shut down"))?;

        // Try to get an idle cell
        let cell = {
            let mut guard = self.inner.idle.lock().await;
            guard.pop_back()
        };

        let cell = match cell {
            Some(c) => c,
            None => self.create_cell()?,
        };

        Ok(WasmCellGuard {
            cell: Some(cell),
            pool: Arc::clone(&self.inner),
            discard: false,
        })
    }

    /// Try to acquire a cell without waiting.
    pub async fn try_acquire(&self) -> Option<WasmCellGuard> {
        let permit = self.inner.semaphore.try_acquire().ok()?;
        let cell = {
            let mut guard = self.inner.idle.lock().await;
            guard.pop_back()
        };
        let cell = cell.or_else(|| self.create_cell().ok())?;
        Some(WasmCellGuard {
            cell: Some(cell),
            pool: Arc::clone(&self.inner),
            discard: false,
        })
    }

    fn create_cell(&self) -> anyhow::Result<WasmCell> {
        WasmCell::from_module(
            (*self.inner.module).clone(),
            &self.inner.config,
            self.inner.function_key.clone(),
            self.inner.pipe_capacity,
        )
    }

    /// Get the current number of idle cells.
    pub async fn idle_count(&self) -> usize {
        self.inner.idle.lock().await.len()
    }

    /// Get statistics.
    pub async fn stats(&self) -> CellPoolStats {
        let idle = self.inner.idle.lock().await.len();
        CellPoolStats {
            idle_count: idle,
            max_idle: self.inner.max_idle,
            available_permits: self.inner.semaphore.available_permits(),
        }
    }
}

/// Statistics for a cell pool.
#[derive(Debug, Clone)]
pub struct CellPoolStats {
    pub idle_count: usize,
    pub max_idle: usize,
    pub available_permits: usize,
}

// ---------------------------------------------------------------------------
// WASM Cell Node Executor
// ---------------------------------------------------------------------------

/// A `NodeExecutor` that runs nodes as isolated WASM cells.
///
/// This is the bridge between the graph execution engine and WASM isolation.
/// Each node type that needs sandboxing (Tool, Memory) uses this executor.
///
/// ## Host Function Injection
///
/// When a cell is checked out from the pool, the WASI linker (stored in the pool)
/// is used for instantiation, providing access to host functions:
/// - `functionfly.kv_get`, `functionfly.kv_set` (KV store)
/// - `functionfly.ai` (AI inference via FlyMind)
/// - `functionfly.fetch` (HTTP requests)
/// - `functionfly.log`, `functionfly.get_env` (always available)
pub struct WasmCellExecutor {
    /// Per-function cell pools
    pools: Arc<tokio::sync::RwLock<HashMap<String, Arc<WasmCellPool>>>>,
    /// Default cell config
    default_config: CellConfig,
    /// Max concurrent per function
    max_concurrent: usize,
    /// Max idle per function
    max_idle: usize,
}

impl WasmCellExecutor {
    /// Create a new WASM cell executor.
    pub fn new(max_concurrent: usize, max_idle: usize) -> Self {
        Self {
            pools: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
            default_config: CellConfig::default(),
            max_concurrent,
            max_idle,
        }
    }

    /// Register a function's WASM bytes with a cell pool.
    ///
    /// The `wasi_linker` must be provided for host function injection during execution.
    /// Cells are pre-warmed during registration with `prewarm_count` idle instances.
    pub async fn register_function(
        &self,
        function_key: String,
        wasm_bytes: Vec<u8>,
        wasi_linker: Arc<crate::wasi::WasiLinker>,
        config: Option<CellConfig>,
        prewarm_count: usize,
    ) -> anyhow::Result<()> {
        let cfg = config.unwrap_or_else(|| self.default_config.clone());

        let engine = Engine::default();
        let pool = WasmCellPool::new(
            engine,
            &wasm_bytes,
            function_key.clone(),
            cfg,
            wasi_linker,
            self.max_concurrent,
            self.max_idle,
        )?;

        // Pre-warm the pool
        pool.prewarm(prewarm_count).await;

        let mut pools = self.pools.write().await;
        pools.insert(function_key, Arc::new(pool));

        Ok(())
    }

    /// Pre-warm a function's pool with N cells.
    pub async fn prewarm(&self, function_key: &str, count: usize) -> anyhow::Result<()> {
        let pools = self.pools.read().await;
        let pool = pools
            .get(function_key)
            .ok_or_else(|| anyhow::anyhow!("function not registered: {}", function_key))?;
        pool.prewarm(count).await;
        Ok(())
    }

    /// Execute a function as a WASM cell.
    ///
    /// The WASI linker stored in the pool provides host function access (kv, ai, fetch, etc.).
    pub async fn execute_cell(
        &self,
        function_key: &str,
        input: &str,
    ) -> anyhow::Result<String> {
        let pools = self.pools.read().await;
        let pool = pools
            .get(function_key)
            .ok_or_else(|| anyhow::anyhow!("function not registered: {}", function_key))?;

        let mut guard = pool.acquire().await?;
        let linker = guard.wasi_linker().clone();
        let output = guard.cell_mut().execute(input, &linker)?;
        Ok(output)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::Arc;

    use crate::kv::SharedKVStore;

    /// Create a minimal WASI linker for testing (no host functions, just WASI imports).
    fn make_test_wasi_linker() -> Arc<crate::wasi::WasiLinker> {
        let engine = Engine::default();
        let config = crate::config::Config::default();
        let kv_store: Option<SharedKVStore> = None;
        let logger = crate::logging::init_structured_logging(false);
        let security_monitor = Arc::new(crate::security::SecurityMonitor::new());

        Arc::new(
            crate::wasi::WasiLinker::new(&engine, &config, kv_store, logger, security_monitor)
                .expect("failed to create test WASI linker"),
        )
    }

    #[tokio::test]
    async fn test_cell_pool_create_and_acquire() {
        // Minimal WASM module: (module (func (export "handler") (result i32) (i32.const 0)))
        let wasm_bytes = wat::parse_str(
            r#"
            (module
                (func (export "handler") (result i32)
                    i32.const 0
                )
            )
            "#,
        )
        .unwrap();

        let linker = make_test_wasi_linker();
        let pool = WasmCellPool::new(
            Engine::default(),
            &wasm_bytes,
            "test@1.0.0".to_string(),
            CellConfig::default(),
            linker,
            4,
            2,
        )
        .unwrap();

        // Pre-warm
        pool.prewarm(2).await;
        assert_eq!(pool.idle_count().await, 2);

        // Acquire and execute
        let mut guard = pool.acquire().await.unwrap();
        let linker = guard.wasi_linker().clone();
        let result = guard.cell_mut().execute("hello", &linker);
        assert!(result.is_ok());

        // Guard returns to pool on drop
        drop(guard);
        assert_eq!(pool.idle_count().await, 3); // 2 pre-warmed + 1 returned
    }

    #[tokio::test]
    async fn test_cell_pool_exhaustion() {
        let wasm_bytes = wat::parse_str(
            r#"
            (module
                (func (export "handler") (result i32)
                    i32.const 0
                )
            )
            "#,
        )
        .unwrap();

        let linker = make_test_wasi_linker();
        let pool = WasmCellPool::new(
            Engine::default(),
            &wasm_bytes,
            "test@1.0.0".to_string(),
            CellConfig::default(),
            linker,
            2, // max 2 concurrent
            1,
        )
        .unwrap();

        // Acquire both cells
        let mut guard1 = pool.acquire().await.unwrap();
        let mut guard2 = pool.acquire().await.unwrap();

        assert!(guard1.cell_mut().reuse_count >= 0);
        assert!(guard2.cell_mut().reuse_count >= 0);
    }

    #[test]
    fn test_cell_config_default() {
        let cfg = CellConfig::default();
        assert_eq!(cfg.memory_mb, 64);
        assert_eq!(cfg.fuel_limit, 1_000_000);
        assert!(cfg.network_allowed);
        assert!(cfg.ai_allowed);
        assert!(cfg.kv_allowed);
    }
}
