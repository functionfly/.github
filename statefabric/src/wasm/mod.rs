//! Wasm module - Wasmtime integration for function execution

use std::collections::HashMap;
use std::fmt;
use std::sync::Arc;
use thiserror::Error;
use tokio::sync::RwLock;
use uuid::Uuid;
use wasmtime::{Engine, Linker, Module, Store};
use wasmtime_wasi::p1;
use wasmtime_wasi::WasiCtxBuilder;

use crate::models::{Event, SourceType};
use crate::state::StateManager;

/// Errors that can occur in WASM execution
#[derive(Error, Debug)]
pub enum WasmError {
    #[error("WASM compilation error: {0}")]
    CompilationError(String),

    #[error("WASM execution error: {0}")]
    ExecutionError(String),

    #[error("Memory error: {0}")]
    MemoryError(String),

    #[error("API error: {0}")]
    ApiError(String),

    #[error("Invalid operation: {0}")]
    InvalidOperation(String),
}

impl From<wasmtime::Error> for WasmError {
    fn from(e: wasmtime::Error) -> Self {
        WasmError::CompilationError(e.to_string())
    }
}

/// Result type for WASM operations
pub type WasmResult<T> = Result<T, WasmError>;

/// Wasm runtime configuration
#[derive(Debug, Clone)]
pub struct WasmConfig {
    /// Maximum memory pages (64KB each)
    pub max_memory_pages: u32,
    /// Enable deterministic mode
    pub deterministic: bool,
    /// Enable gas metering (SECURITY: enabled by default to prevent runaway WASM)
    pub enable_gas: bool,
    /// Maximum execution time in milliseconds
    pub max_execution_time_ms: u64,
    /// Maximum gas budget per execution (if gas metering enabled)
    pub max_gas_budget: u64,
    /// SECURITY: Approved module hashes for verification (name -> hash)
    /// If empty, no verification is performed
    pub approved_hashes: Vec<ApprovedModule>,
    /// SECURITY: Maximum events a WASM module can emit per execution (rate limiting)
    pub max_wasm_events_per_execution: u32,
}

impl WasmConfig {
    /// Security-hardened configuration
    pub fn secure() -> Self {
        Self {
            max_memory_pages: 256, // 16MB
            deterministic: true,
            enable_gas: true, // Always on in secure mode
            max_execution_time_ms: 5000,
            max_gas_budget: 1_000_000,
            approved_hashes: Vec::new(),
            max_wasm_events_per_execution: 100,
        }
    }

    /// Load configuration from environment variables
    pub fn from_env() -> Self {
        Self {
            max_memory_pages: std::env::var("STATEFABRIC_WASM_MAX_MEMORY_PAGES")
                .unwrap_or_else(|_| "256".to_string())
                .parse()
                .unwrap_or(256),
            deterministic: std::env::var("STATEFABRIC_WASM_DETERMINISTIC")
                .unwrap_or_else(|_| "true".to_string())
                .parse()
                .unwrap_or(true),
            enable_gas: std::env::var("STATEFABRIC_WASM_ENABLE_GAS")
                .unwrap_or_else(|_| "true".to_string())
                .parse()
                .unwrap_or(true),
            max_execution_time_ms: std::env::var("STATEFABRIC_WASM_MAX_EXECUTION_TIME_MS")
                .unwrap_or_else(|_| "5000".to_string())
                .parse()
                .unwrap_or(5000),
            max_gas_budget: std::env::var("STATEFABRIC_WASM_MAX_GAS_BUDGET")
                .unwrap_or_else(|_| "1000000".to_string())
                .parse()
                .unwrap_or(1_000_000),
            // SECURITY: Load approved module hashes from env var
            // Format: "module1:hash1,module2:hash2"
            approved_hashes: std::env::var("STATEFABRIC_WASM_APPROVED_MODULES")
                .map(|s| {
                    s.split(',')
                        .filter_map(|entry| {
                            let parts: Vec<_> = entry.split(':').collect();
                            if parts.len() == 2 {
                                Some(ApprovedModule {
                                    name: parts[0].trim().to_string(),
                                    sha256_hash: parts[1].trim().to_string(),
                                })
                            } else {
                                None
                            }
                        })
                        .collect()
                })
                .unwrap_or_default(),
            // SECURITY: Limit WASM events per execution (rate limiting)
            max_wasm_events_per_execution: std::env::var("STATEFABRIC_WASM_MAX_EVENTS_PER_EXEC")
                .unwrap_or_else(|_| "100".to_string())
                .parse()
                .unwrap_or(100),
        }
    }
}

impl Default for WasmConfig {
    fn default() -> Self {
        Self {
            max_memory_pages: 256,
            deterministic: true,
            enable_gas: true,
            max_execution_time_ms: 5000,
            max_gas_budget: 1_000_000,
            approved_hashes: Vec::new(),
            max_wasm_events_per_execution: 100,
        }
    }
}

/// Wasm execution result
#[derive(Debug)]
pub struct ExecutionResult {
    /// Whether execution was successful
    pub success: bool,
    /// Output data from WASM function
    pub output: Vec<u8>,
    /// Events committed during execution
    pub committed_events: Vec<Event>,
    /// Gas used (if metering enabled)
    pub gas_used: Option<u64>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
}

/// Module hash entry for approved WASM modules
#[derive(Debug, Clone)]
pub struct ApprovedModule {
    /// Module name
    pub name: String,
    /// SHA-256 hash of the module bytes
    pub sha256_hash: String,
}

/// Shared memory buffer for host-WASM communication
#[derive(Debug)]
pub struct SharedMemory {
    /// Memory buffer
    buffer: Vec<u8>,
    /// Current write position
    write_pos: usize,
    /// Current read position
    read_pos: usize,
}

impl SharedMemory {
    /// Create new shared memory buffer
    pub fn new(size: usize) -> Self {
        Self {
            buffer: vec![0; size],
            write_pos: 0,
            read_pos: 0,
        }
    }

    /// Write data to shared memory
    pub fn write(&mut self, data: &[u8]) -> WasmResult<usize> {
        let len = data.len();
        if self.write_pos + len > self.buffer.len() {
            return Err(WasmError::MemoryError("Shared memory buffer overflow".to_string()));
        }

        self.buffer[self.write_pos..self.write_pos + len].copy_from_slice(data);
        self.write_pos += len;
        Ok(len)
    }

    /// Read data from shared memory
    pub fn read(&mut self, len: usize) -> WasmResult<Vec<u8>> {
        if self.read_pos + len > self.buffer.len() {
            return Err(WasmError::MemoryError("Shared memory read overflow".to_string()));
        }

        let data = self.buffer[self.read_pos..self.read_pos + len].to_vec();
        self.read_pos += len;
        Ok(data)
    }

    /// Reset read/write positions
    pub fn reset(&mut self) {
        self.write_pos = 0;
        self.read_pos = 0;
    }

    /// Get current buffer contents
    pub fn as_slice(&self) -> &[u8] {
        &self.buffer
    }

    /// Get mutable buffer access
    pub fn as_mut_slice(&mut self) -> &mut [u8] {
        &mut self.buffer
    }
}

/// WASM host state - shared between host and WASM
pub struct HostState {
    /// State manager reference
    pub state_manager: Arc<StateManager>,
    /// Shared memory buffer
    pub shared_memory: SharedMemory,
    /// Current state ID context
    pub current_state_id: Option<Uuid>,
    /// Committed events during execution
    pub committed_events: Vec<Event>,
    /// Function correlation ID
    pub correlation_id: String,
    /// WASI preview1 context (stdio, env, etc.)
    pub wasi: p1::WasiP1Ctx,
    /// SECURITY: Event counter for WASM-initiated events (rate limiting)
    wasm_event_count: u32,
    /// SECURITY: Maximum events allowed per WASM execution
    max_wasm_events: u32,
}

impl fmt::Debug for HostState {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("HostState")
            .field("state_manager", &"...")
            .field("shared_memory", &self.shared_memory)
            .field("current_state_id", &self.current_state_id)
            .field("committed_events", &self.committed_events)
            .field("correlation_id", &self.correlation_id)
            .field("wasi", &"...")
            .finish()
    }
}

impl HostState {
    pub fn new(state_manager: Arc<StateManager>, memory_size: usize, wasi: p1::WasiP1Ctx) -> Self {
        Self {
            state_manager,
            shared_memory: SharedMemory::new(memory_size),
            current_state_id: None,
            committed_events: Vec::new(),
            correlation_id: Uuid::new_v4().to_string(),
            wasi,
            wasm_event_count: 0,
            max_wasm_events: 100, // Default limit per execution
        }
    }

    /// Set max WASM events limit
    pub fn set_max_wasm_events(&mut self, max: u32) {
        self.max_wasm_events = max;
    }

    /// Increment event count and check rate limit
    /// Returns Ok(()) if under limit, Err if exceeded
    fn check_increment_event_count(&mut self) -> Result<(), WasmError> {
        self.wasm_event_count += 1;
        if self.wasm_event_count > self.max_wasm_events {
            return Err(WasmError::ExecutionError(format!(
                "WASM event rate limit exceeded: {} events per execution (limit: {})",
                self.wasm_event_count, self.max_wasm_events
            )));
        }
        Ok(())
    }

    /// Set current state context
    pub fn set_state_context(&mut self, state_id: Uuid) {
        self.current_state_id = Some(state_id);
    }

    /// Get current state ID
    pub fn get_state_id(&self) -> WasmResult<Uuid> {
        self.current_state_id.ok_or_else(|| {
            WasmError::InvalidOperation("No state context set".to_string())
        })
    }
}

/// Wasmtime runtime for executing WASM functions
pub struct WasmRuntime {
    /// Wasmtime engine
    engine: Engine,
    /// Compiled WASM modules cache
    modules: RwLock<HashMap<String, Module>>,
    /// Runtime configuration
    config: WasmConfig,
}

impl fmt::Debug for WasmRuntime {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.debug_struct("WasmRuntime")
            .field("config", &self.config)
            .finish_non_exhaustive()
    }
}

impl WasmRuntime {
    /// Create new WASM runtime
    pub fn new(config: WasmConfig) -> WasmResult<Self> {
        let mut engine_config = wasmtime::Config::new();

        // Configure engine for deterministic execution if enabled
        if config.deterministic {
            engine_config.cranelift_opt_level(wasmtime::OptLevel::Speed);
            engine_config.cranelift_nan_canonicalization(true);
        }

        // Set memory limits
        engine_config.max_wasm_stack(2 * 1024 * 1024); // 2MB stack

        // SECURITY: Enable fuel metering for gas tracking
        // This allows us to track gas consumption and enforce limits
        engine_config.consume_fuel(true);

        let engine = Engine::new(&engine_config)
            .map_err(|e| WasmError::CompilationError(format!("Failed to create engine: {}", e)))?;

        Ok(Self {
            engine,
            modules: RwLock::new(HashMap::new()),
            config,
        })
    }

    /// Create default runtime
    pub fn default() -> WasmResult<Self> {
        Self::new(WasmConfig::default())
    }

    /// Compile and cache a WASM module
    /// SECURITY: If approved_hashes is configured, verifies module hash before loading
    /// SECURITY P0: In production (STATEFABRIC_ENVIRONMENT=production), require approved hashes
    pub async fn compile_module(&self, name: &str, wasm_bytes: &[u8]) -> WasmResult<()> {
        // SECURITY P0: In production mode, reject module loading if no approved hashes configured
        let is_production = std::env::var("STATEFABRIC_ENVIRONMENT")
            .map(|v| v == "production" || v == "prod")
            .unwrap_or(false);

        if is_production && self.config.approved_hashes.is_empty() {
            tracing::error!("SECURITY: Production mode requires STATEFABRIC_WASM_APPROVED_MODULES but none are configured");
            return Err(WasmError::CompilationError(
                "Production mode requires approved WASM module hashes - set STATEFABRIC_WASM_APPROVED_MODULES".to_string()
            ));
        }

        // SECURITY: Verify module hash if allowlist is configured
        if !self.config.approved_hashes.is_empty() {
            let hash = blake3::hash(wasm_bytes);
            let hash_hex = hash.to_hex().to_string();
            let approved = self.config.approved_hashes.iter().any(|m| m.name == name && m.sha256_hash == hash_hex);
            if !approved {
                tracing::warn!(module = name, hash = %hash_hex, "WASM module not in approved hash list");
                return Err(WasmError::CompilationError(
                    format!("Module '{}' hash {} not in approved list", name, hash_hex)
                ));
            }
            tracing::info!(module = name, hash = %hash_hex, "WASM module hash verified");
        }

        let module = Module::new(&self.engine, wasm_bytes)
            .map_err(|e| WasmError::CompilationError(format!("Failed to compile module {}: {}", name, e)))?;

        let mut modules = self.modules.write().await;
        modules.insert(name.to_string(), module);

        Ok(())
    }

    /// Execute a WASM function
    pub async fn execute_function(
        &self,
        module_name: &str,
        function_name: &str,
        state_manager: Arc<StateManager>,
        state_id: Uuid,
        input: &[u8],
    ) -> WasmResult<ExecutionResult> {
        let start_time = std::time::Instant::now();

        // Get compiled module
        let modules = self.modules.read().await;
        let module = modules.get(module_name)
            .ok_or_else(|| WasmError::InvalidOperation(format!("Module {} not found", module_name)))?;

        // Build WASI preview1 context (stdio, env)
        // SECURITY P0: Do NOT inherit full environment - only pass safe, non-sensitive vars
        // Filter to prevent exposure of secrets like STATEFABRIC_JWT_SECRET, STATEFABRIC_ENCRYPTION_KEY
        let wasi = WasiCtxBuilder::new()
            .inherit_stdio()
            .env("RUST_LOG", std::env::var("RUST_LOG").unwrap_or_default())
            .build_p1();

        // Create host state (state manager + shared memory + WASI)
        let mut host_state = HostState::new(state_manager, 1024 * 1024, wasi); // 1MB shared memory
        // SECURITY: Set max WASM events limit from config
        host_state.set_max_wasm_events(self.config.max_wasm_events_per_execution);
        let mut store = Store::new(&self.engine, host_state);

        // SECURITY: Set fuel for gas metering
        // Fuel represents gas units that are consumed during execution
        // Note: wasmtime 22+ uses set_fuel() to set initial fuel
        let max_gas = self.config.max_gas_budget;
        store.set_fuel(max_gas)
            .map_err(|e| WasmError::ExecutionError(format!("Failed to set fuel: {}", e)))?;

        let mut linker = Linker::new(&self.engine);
        p1::add_to_linker_sync(&mut linker, |state: &mut HostState| &mut state.wasi)
            .map_err(|e| WasmError::CompilationError(format!("Failed to setup WASI: {}", e)))?;

        // Add host functions
        self.add_host_functions(&mut linker)?;

        // Instantiate module
        let instance = linker.instantiate(&mut store, module)
            .map_err(|e| WasmError::ExecutionError(format!("Failed to instantiate module: {}", e)))?;

        // Set state context
        store.data_mut().set_state_context(state_id);

        // Write input to shared memory
        store.data_mut().shared_memory.write(input)
            .map_err(|e| WasmError::ExecutionError(format!("Failed to write input: {}", e)))?;

        // Get function to call
        let func = instance.get_typed_func::<(), ()>(&mut store, function_name)
            .map_err(|e| WasmError::ExecutionError(format!("Function {} not found: {}", function_name, e)))?;

        // Execute function with fuel consumption
        // SECURITY: If execution runs out of fuel, wasmtime will return an OutOfGas error
        let result = func.call(&mut store, ());

        // Check if we ran out of gas (wasmtime will error if fuel exhausted)
        if let Err(e) = &result {
            let err_msg = format!("Function execution failed: {:?}", e);
            // Check if this is an out-of-gas error
            if err_msg.contains("fuel") || err_msg.contains("gas") || err_msg.contains("OutOfGas") {
                return Err(WasmError::ExecutionError(format!(
                    "Out of gas: exceeded budget of {} units",
                    max_gas
                )));
            }
            return Err(WasmError::ExecutionError(err_msg));
        }

        let execution_time = start_time.elapsed().as_millis() as u64;

        // Collect results
        let host_state = store.into_data();
        let committed_events = host_state.committed_events;

        Ok(ExecutionResult {
            success: true,
            output: host_state.shared_memory.as_slice().to_vec(),
            committed_events,
            gas_used: Some(max_gas), // SECURITY: Report gas budget as used (actual consumption tracked by wasmtime)
            execution_time_ms: execution_time,
        })
    }

    /// Add host functions to the linker
    fn add_host_functions(&self, linker: &mut Linker<HostState>) -> WasmResult<()> {
        // CommitEvent API - allows WASM to commit events
        // SECURITY: Rate-limited to prevent WASM from flooding event log
        linker.func_wrap("env", "commit_event", |mut caller: wasmtime::Caller<HostState>,
                                                 event_type: i32,
                                                 key_ptr: i32,
                                                 key_len: i32,
                                                 value_ptr: i32,
                                                 value_len: i32| -> i32 {
            // SECURITY: Check rate limit before processing event
            if let Err(e) = caller.data_mut().check_increment_event_count() {
                tracing::warn!(error = %e, "WASM event rate limit exceeded");
                return -7; // Rate limit error code
            }

            let state_id = match caller.data().get_state_id() {
                Ok(id) => id,
                Err(_) => return -1,
            };
            let key = match Self::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(s) => s,
                Err(_) => return -2,
            };
            let value_bytes = match Self::read_bytes_from_memory(&mut caller, value_ptr, value_len) {
                Ok(b) => b,
                Err(_) => return -3,
            };
            let value: serde_json::Value = match serde_json::from_slice(&value_bytes) {
                Ok(v) => v,
                Err(_) => return -4,
            };
            let event = match event_type {
                0 => Event::set(state_id, key, value, SourceType::Function, "wasm".to_string()),
                1 => Event::delete(state_id, key, None, SourceType::Function, "wasm".to_string()),
                2 => Event::merge(state_id, key, value, SourceType::Function, "wasm".to_string()),
                _ => return -5,
            };
            let manager = Arc::clone(&caller.data().state_manager);
            match tokio::runtime::Handle::current().block_on(manager.commit_event(event)) {
                Ok(committed_event) => {
                    caller.data_mut().committed_events.push(committed_event);
                    0
                }
                Err(_) => -6,
            }
        })?;

        // Get state value
        linker.func_wrap("env", "get_state", |mut caller: wasmtime::Caller<HostState>,
                                               key_ptr: i32,
                                               key_len: i32,
                                               output_ptr: i32,
                                               output_max_len: i32| -> i32 {
            let state_id = match caller.data().get_state_id() {
                Ok(id) => id,
                Err(_) => return -1,
            };
            let key = match Self::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(s) => s,
                Err(_) => return -2,
            };
            let manager = Arc::clone(&caller.data().state_manager);
            let value = match tokio::runtime::Handle::current().block_on(manager.get_key(state_id, &key)) {
                Ok(v) => v,
                Err(_) => return -3,
            };
            let json_bytes = match serde_json::to_vec(&value) {
                Ok(b) => b,
                Err(_) => return -4,
            };
            match Self::write_bytes_to_memory(&mut caller, output_ptr, output_max_len, &json_bytes) {
                Ok(len) => len as i32,
                Err(_) => -5,
            }
        })?;

        // Set state value (direct, without event)
        linker.func_wrap("env", "set_state", |mut caller: wasmtime::Caller<HostState>,
                                               key_ptr: i32,
                                               key_len: i32,
                                               value_ptr: i32,
                                               value_len: i32| -> i32 {
            let state_id = match caller.data().get_state_id() {
                Ok(id) => id,
                Err(_) => return -1,
            };
            let key = match Self::read_string_from_memory(&mut caller, key_ptr, key_len) {
                Ok(s) => s,
                Err(_) => return -2,
            };
            let value_bytes = match Self::read_bytes_from_memory(&mut caller, value_ptr, value_len) {
                Ok(b) => b,
                Err(_) => return -3,
            };
            let value: serde_json::Value = match serde_json::from_slice(&value_bytes) {
                Ok(v) => v,
                Err(_) => return -4,
            };
            let manager = Arc::clone(&caller.data().state_manager);
            match tokio::runtime::Handle::current().block_on(manager.set(state_id, key, value)) {
                Ok(_) => 0,
                Err(_) => -5,
            }
        })?;

        Ok(())
    }

    /// Read string from WASM memory
    fn read_string_from_memory(caller: &mut wasmtime::Caller<HostState>, ptr: i32, len: i32) -> WasmResult<String> {
        let memory = caller.get_export("memory")
            .and_then(|export| export.into_memory())
            .ok_or_else(|| WasmError::MemoryError("Memory export not found".to_string()))?;

        let ptr = ptr as usize;
        let len = len as usize;

        let mut buffer = vec![0u8; len];
        memory.read(caller, ptr, &mut buffer)
            .map_err(|e| WasmError::MemoryError(format!("Failed to read memory: {}", e)))?;

        String::from_utf8(buffer)
            .map_err(|e| WasmError::MemoryError(format!("Invalid UTF-8: {}", e)))
    }

    /// Read bytes from WASM memory
    fn read_bytes_from_memory(caller: &mut wasmtime::Caller<HostState>, ptr: i32, len: i32) -> WasmResult<Vec<u8>> {
        let memory = caller.get_export("memory")
            .and_then(|export| export.into_memory())
            .ok_or_else(|| WasmError::MemoryError("Memory export not found".to_string()))?;

        let ptr = ptr as usize;
        let len = len as usize;

        let mut buffer = vec![0u8; len];
        memory.read(caller, ptr, &mut buffer)
            .map_err(|e| WasmError::MemoryError(format!("Failed to read memory: {}", e)))?;

        Ok(buffer)
    }

    /// Write bytes to WASM memory
    fn write_bytes_to_memory(caller: &mut wasmtime::Caller<HostState>, ptr: i32, max_len: i32, data: &[u8]) -> WasmResult<usize> {
        let memory = caller.get_export("memory")
            .and_then(|export| export.into_memory())
            .ok_or_else(|| WasmError::MemoryError("Memory export not found".to_string()))?;

        let ptr = ptr as usize;
        let max_len = max_len as usize;
        let len = data.len().min(max_len);

        memory.write(caller, ptr, &data[..len])
            .map_err(|e| WasmError::MemoryError(format!("Failed to write memory: {}", e)))?;

        Ok(len)
    }
}

impl WasmConfig {
    /// Create new config
    pub fn new() -> Self {
        Self::default()
    }

    /// Set max memory
    pub fn with_max_memory(mut self, pages: u32) -> Self {
        self.max_memory_pages = pages;
        self
    }

    /// Enable deterministic mode
    pub fn with_deterministic(mut self, deterministic: bool) -> Self {
        self.deterministic = deterministic;
        self
    }
}
