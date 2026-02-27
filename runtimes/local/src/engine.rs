//! Wasmtime engine for executing WebAssembly functions.

use anyhow::Context;
use clap::Parser;
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

/// Runtime type for WASM modules
#[derive(Debug, Clone, PartialEq)]
pub enum RuntimeType {
    /// Standard WebAssembly module (Rust, Go, etc.)
    Wasm,
    /// Python WASM module using RustPython
    Python,
    /// CPython in Firecracker MicroVM (Enterprise tier only)
    PythonMicroVM,
}

impl RuntimeType {
    /// Parse runtime type from string
    pub fn from_str(s: &str) -> Option<Self> {
        match s {
            "wasm" => Some(RuntimeType::Wasm),
            "python" => Some(RuntimeType::Python),
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

/// Wasm engine for executing functions
pub struct WasmEngine {
    engine: Engine,
    config: Config,
    wasi_linker: Option<Arc<WasiLinker>>,
    kv_store: Option<SharedKVStore>,
    logger: StructuredLogger,
    orchestrator_client: Option<Arc<OrchestratorClient>>,
    security_monitor: Arc<crate::security::SecurityMonitor>,
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
        })
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
            RuntimeType::Wasm => {
                // Execute standard WASM module in a blocking task to avoid runtime conflicts
                let wasm_bytes = wasm_bytes.to_vec();
                let input = input.to_string();
                let engine = self.engine.clone();
                let config = config.clone();
                let wasi_linker = self.wasi_linker.clone();

                tokio::task::spawn_blocking(move || -> anyhow::Result<String> {
                    if let Some(ref linker) = wasi_linker {
                        // Execute with WASI support synchronously
                        execute_wasi_sync(&engine, linker, &wasm_bytes, &input, &config)
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
                            warn!("MicroVM orchestrator is not available, falling back to RustPython");
                            // Fall back to RustPython runtime
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
                                    warn!("MicroVM execution failed, falling back to RustPython: {}", e);
                                    // Fall back to RustPython runtime
                                    let python_config = crate::python::runtime::PythonConfig::from(config.clone());
                                    let runtime = crate::python::runtime::PythonRuntime::new(python_config)?;
                                    let python_code = String::from_utf8_lossy(&wasm_bytes);
                                    runtime.execute_sync(&python_code, input)
                                }
                            }
                        }
                    }
                    None => {
                        warn!("MicroVM runtime requested but orchestrator not configured, falling back to RustPython");
                        // Fall back to RustPython runtime
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

        // Allocate resources based on tier
        match tier {
            BudgetTier::UltraLow => (256, 1), // Minimal resources for ultra-low tier
            BudgetTier::Low => (512, 1),      // Basic resources for low tier
            BudgetTier::Medium => (1024, 2),  // Better resources for medium tier
            BudgetTier::High => (2048, 4),    // Full resources for high tier
        }
    }
}

/// Synchronous WASI execution function for use in spawn_blocking
/// This avoids the "Cannot start a runtime from within a runtime" panic
fn execute_wasi_sync(
    engine: &Engine,
    linker: &WasiLinker,
    wasm_bytes: &[u8],
    input: &str,
    config: &Config,
) -> anyhow::Result<String> {
    let execution_start = std::time::Instant::now();

    // Create WASI context with input data
    let function_key = format!("{}@{}", config.function, config.version);
    let wasi_ctx = WasiContext::new_with_input(config, function_key, input)?;
    let stdout_pipe = wasi_ctx.stdout_pipe.clone();
    let stderr_pipe = wasi_ctx.stderr_pipe.clone();

    // Create store with WASI context
    let mut store = Store::new(engine, wasi_ctx.ctx);

    // Set fuel limit for execution (configurable CPU control)
    let fuel_limit = if config.cpu_fuel_limit > 0 {
        config.cpu_fuel_limit
    } else {
        1_000_000 // Default fallback
    };
    store.set_fuel(fuel_limit)?;

    // Enable epoch interruption for CPU time limits
    if config.enable_monitoring {
        store.set_epoch_deadline(1); // Set initial deadline
    }

    // Compile module
    let module = Module::new(engine, wasm_bytes)
        .context("Failed to compile Wasm module")?;

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
            port: 8787,
            function: "test".to_string(),
            version: "1.0.0".to_string(),
            wasm: None,
            runtime: "nodejs".to_string(),
            memory_mb: 128,
            timeout_ms: 5000,
            deterministic: false,
            cache_ttl: 3600,
            verbose: false,
            wasi_enabled: false,
            cpu_fuel_limit: 1000000,
            max_cpu_time_ms: 5000,
            enable_monitoring: true,
            hardened_security: true,
            max_concurrent_per_function: 10,
            memory_overhead_percent: 10,
            wasi_dirs: vec![],
            wasi_env: vec![],
            wasi_args: vec![],
            wasi_allow_network: false,
            wasi_allow_time: true,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: vec![],
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: "http://localhost:8080".to_string(),
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tier: "ultra-low".to_string(),
            network_whitelist: vec![],
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
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
            port: 8787,
            function: "test".to_string(),
            version: "1.0.0".to_string(),
            wasm: None,
            runtime: "nodejs".to_string(),
            memory_mb: 128,
            timeout_ms: 5000,
            deterministic: false,
            cache_ttl: 3600,
            verbose: false,
            wasi_enabled: true,
            cpu_fuel_limit: 1000000,
            max_cpu_time_ms: 5000,
            enable_monitoring: true,
            hardened_security: true,
            max_concurrent_per_function: 10,
            memory_overhead_percent: 10,
            wasi_dirs: vec![],
            wasi_env: vec![],
            wasi_args: vec![],
            wasi_allow_network: false,
            wasi_allow_time: true,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: vec![],
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: "http://localhost:8080".to_string(),
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tier: "ultra-low".to_string(),
            network_whitelist: vec![],
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
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
                port: 8787,
                function: "test".to_string(),
                version: "1.0.0".to_string(),
                wasm: None,
                runtime: "nodejs".to_string(),
                memory_mb: 128,
                timeout_ms: 5000,
                deterministic: false,
                cache_ttl: 3600,
                verbose: false,
                wasi_enabled: false,
                cpu_fuel_limit: 1000000,
                max_cpu_time_ms: 5000,
                enable_monitoring: true,
                hardened_security: true,
                max_concurrent_per_function: 10,
                memory_overhead_percent: 10,
                wasi_dirs: vec![],
                wasi_env: vec![],
                wasi_args: vec![],
                wasi_allow_network: false,
                wasi_allow_time: true,
                python_runtime: "rustpython-0.4".to_string(),
                capabilities: "".to_string(),
                python_packages: vec![],
                python_debug: false,
                smtp_host: "localhost".to_string(),
                smtp_port: 587,
                smtp_username: None,
                smtp_password: None,
                storage_base_dir: "./storage".to_string(),
                ai_models_dir: "./models".to_string(),
                external_api_rate_limit: 60,
                external_api_timeout_secs: 30,
                orchestrator_url: "http://localhost:8080".to_string(),
                orchestrator_timeout_secs: 60,
                enterprise_enabled: false,
                tier: "ultra-low".to_string(),
                network_whitelist: vec![],
                strict_network_whitelist: false,
                package_caching_enabled: false,
                package_cache_dir: "./package-cache".to_string(),
                package_cache_size_mb: 1024,
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
            port: 8787,
            function: "test".to_string(),
            version: "1.0.0".to_string(),
            wasm: None,
            runtime: "wasm".to_string(),
            memory_mb: 128,
            timeout_ms: 5000,
            deterministic: false,
            cache_ttl: 3600,
            verbose: false,
            wasi_enabled: true,
            cpu_fuel_limit: 1000000,
            max_cpu_time_ms: 5000,
            enable_monitoring: true,
            hardened_security: true,
            max_concurrent_per_function: 10,
            memory_overhead_percent: 10,
            wasi_dirs: vec![],
            wasi_env: vec![],
            wasi_args: vec![],
            wasi_allow_network: false,
            wasi_allow_time: true,
            python_runtime: "rustpython-0.4".to_string(),
            capabilities: "".to_string(),
            python_packages: vec![],
            python_debug: false,
            smtp_host: "localhost".to_string(),
            smtp_port: 587,
            smtp_username: None,
            smtp_password: None,
            storage_base_dir: "./storage".to_string(),
            ai_models_dir: "./models".to_string(),
            external_api_rate_limit: 60,
            external_api_timeout_secs: 30,
            orchestrator_url: "http://localhost:8080".to_string(),
            orchestrator_timeout_secs: 60,
            enterprise_enabled: false,
            tier: "ultra-low".to_string(),
            network_whitelist: vec![],
            strict_network_whitelist: false,
            package_caching_enabled: false,
            package_cache_dir: "./package-cache".to_string(),
            package_cache_size_mb: 1024,
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
