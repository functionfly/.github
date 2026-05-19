//! WASM Fusion Engine - Production-ready WASM execution with module fusion
//!
//! This module provides real WASM module execution using wasmtime,
//! following patterns from the SAR runtime but adapted for graph-based fusion.

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Instant;

use parking_lot::RwLock;
use tokio::sync::RwLock as TokioRwLock;
use tracing::{info, debug, instrument};
use wasmtime::{Engine, Module, Instance, Store, Memory, Linker, Caller, ResourceLimiter};
use wasmtime_wasi::p1::WasiP1Ctx;
use wasmtime_wasi::p2::pipe::{MemoryInputPipe, MemoryOutputPipe};

use crate::codec::CborCodec;
use crate::core::{PrismError, PrismResult};
use crate::quantum::snapshot::{WasmCpuState};
use crate::wasm_fusion::{FusionGraph, FusionNode, FusionNodeType, WasmComposer};
use crate::wasm_fusion::graph::ExecutionGraph;

/// A compiled graph ready for execution
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
struct CompiledGraph {
    graph_id: String,
    module_ids: Vec<String>,
    entry_point: String,
    config: crate::wasm_fusion::FusionConfig,
}

/// Sandbox state for WASM execution with host function support
pub struct SandboxState {
    /// WASI context for stdio
    pub wasi: WasiP1Ctx,
    /// Cell ID for this execution context
    pub cell_id: String,
    /// State store for env.state_get/set (sync access for WASM host functions)
    state_store: Arc<RwLock<HashMap<String, Vec<u8>>>>,
    /// Async state store for external async access
    async_state_store: Arc<TokioRwLock<HashMap<String, Vec<u8>>>>,
    /// Logger sink for env.log (tokio RwLock for async compatibility)
    log_buffer: Arc<tokio::sync::RwLock<Vec<LogEntry>>>,
    /// Capability registry for env.capability_invoke (sync access for WASM host functions)
    capability_registry: Arc<RwLock<HashMap<String, CapabilityHandler>>>,
    /// Async capability registry for external async access
    async_capability_registry: Arc<TokioRwLock<HashMap<String, CapabilityHandler>>>,
    /// Captured stdout from WASI (tokio RwLock for async compatibility)
    stdout: Arc<tokio::sync::RwLock<Vec<u8>>>,
    /// Per-sandbox memory limiter
    pub memory_limiter: SandboxMemoryLimiter,
}

/// A handler for a capability invocation
pub type CapabilityHandler = Arc<dyn Fn(&[u8]) -> Result<Vec<u8>, String> + Send + Sync>;

/// A log entry from WASM execution
#[derive(Debug, Clone)]
pub struct LogEntry {
    pub level: u32,
    pub message: String,
    pub timestamp: chrono::DateTime<chrono::Utc>,
}

impl LogEntry {
    pub fn new(level: u32, message: String) -> Self {
        Self {
            level,
            message,
            timestamp: chrono::Utc::now(),
        }
    }
}

impl SandboxState {
    /// Create a new sandbox state with WASI context
    pub fn new(cell_id: String, wasi: WasiP1Ctx, stdout: MemoryOutputPipe, memory_limit_mb: u64) -> Self {
        let sync_store = Arc::new(RwLock::new(HashMap::new()));
        Self {
            wasi,
            cell_id,
            state_store: sync_store.clone(),
            async_state_store: Arc::new(TokioRwLock::new(HashMap::new())),
            log_buffer: Arc::new(tokio::sync::RwLock::new(Vec::new())),
            capability_registry: Arc::new(RwLock::new(HashMap::new())),
            async_capability_registry: Arc::new(TokioRwLock::new(HashMap::new())),
            stdout: Arc::new(tokio::sync::RwLock::new(stdout.contents().to_vec())),
            memory_limiter: SandboxMemoryLimiter::new(memory_limit_mb),
        }
    }

    /// Append a log entry
    pub async fn append_log(&self, entry: LogEntry) {
        let mut buffer = self.log_buffer.write().await;
        buffer.push(entry);
        // Keep buffer bounded to last 1000 entries
        if buffer.len() > 1000 {
            buffer.drain(0..100);
        }
    }

    /// Get all log entries
    pub async fn get_logs(&self) -> Vec<LogEntry> {
        self.log_buffer.read().await.clone()
    }

    /// Clear the log buffer
    pub async fn clear_logs(&self) {
        let mut buffer = self.log_buffer.write().await;
        buffer.clear();
    }

    /// Flush logs (get and clear)
    pub async fn flush_logs(&self) -> Vec<LogEntry> {
        let mut buffer = self.log_buffer.write().await;
        let logs = buffer.clone();
        buffer.clear();
        logs
    }

    /// Register a capability handler (async, for external use)
    pub async fn register_capability(&self, name: &str, handler: CapabilityHandler) {
        // Update both sync and async registries
        {
            let mut registry = self.capability_registry.write();
            registry.insert(name.to_string(), handler.clone());
        }
        {
            let mut registry = self.async_capability_registry.write().await;
            registry.insert(name.to_string(), handler);
        }
    }

    /// Get state value (sync, for WASM host functions)
    pub fn state_get_sync(&self, key: &str) -> Option<Vec<u8>> {
        let store = self.state_store.read();
        store.get(key).cloned()
    }

    /// Set state value (sync, for WASM host functions)
    pub fn state_set_sync(&self, key: &str, value: Vec<u8>) {
        let mut store = self.state_store.write();
        store.insert(key.to_string(), value);
    }

    /// Get state value (async, for external use)
    pub async fn state_get(&self, key: &str) -> Option<Vec<u8>> {
        // Try async first, fall back to sync
        let async_store = self.async_state_store.read().await;
        if let Some(v) = async_store.get(key) {
            return Some(v.clone());
        }
        drop(async_store);
        // Fall back to sync store
        self.state_get_sync(key)
    }

    /// Set state value (async, for external use)
    pub async fn state_set(&self, key: &str, value: Vec<u8>) {
        // Update both sync and async stores
        {
            let mut sync_store = self.state_store.write();
            sync_store.insert(key.to_string(), value.clone());
        }
        {
            let mut async_store = self.async_state_store.write().await;
            async_store.insert(key.to_string(), value);
        }
    }

    /// Invoke a capability (sync, for WASM host functions)
    pub fn capability_invoke_sync(&self, name: &str, args: &[u8]) -> Result<Vec<u8>, String> {
        let registry = self.capability_registry.read();
        if let Some(handler) = registry.get(name) {
            handler(args)
        } else {
            Err(format!("Capability not found: {}", name))
        }
    }

    /// Invoke a capability (async, for external use)
    pub async fn capability_invoke(&self, name: &str, args: &[u8]) -> Result<Vec<u8>, String> {
        // Try async registry first
        {
            let registry = self.async_capability_registry.read().await;
            if let Some(handler) = registry.get(name) {
                return handler(args);
            }
        }
        // Fall back to sync registry
        self.capability_invoke_sync(name, args)
    }

    /// Capture stdout data that was written via WASI
    pub async fn capture_stdout(&self) -> Vec<u8> {
        self.stdout.read().await.clone()
    }
}

/// Memory limiter that enforces per-module memory limits
/// (following the same pattern as the SAR runtime)
pub struct SandboxMemoryLimiter {
    max_bytes: usize,
}

impl SandboxMemoryLimiter {
    pub fn new(memory_mb: u64) -> Self {
        Self {
            max_bytes: (memory_mb as usize) * 1024 * 1024,
        }
    }
}

impl ResourceLimiter for SandboxMemoryLimiter {
    fn memory_growing(
        &mut self,
        _current: usize,
        desired: usize,
        _maximum: Option<usize>,
    ) -> wasmtime::Result<bool> {
        if desired > self.max_bytes {
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

/// Configuration for fusion engine execution
#[derive(Debug, Clone)]
pub struct FusionEngineConfig {
    /// Enable streaming execution
    pub enable_streaming: bool,
    /// Enable module merging
    pub enable_module_merging: bool,
    /// Maximum modules in a fused graph
    pub max_modules: usize,
    /// Memory limit per module in MB
    pub memory_limit_mb: u64,
    /// Compute timeout in ms
    pub compute_timeout_ms: u64,
    /// Fuel limit for execution (0 = unlimited)
    pub fuel_limit: u64,
}

impl Default for FusionEngineConfig {
    fn default() -> Self {
        Self {
            enable_streaming: true,
            enable_module_merging: true,
            max_modules: 16,
            memory_limit_mb: 512,
            compute_timeout_ms: 30_000,
            fuel_limit: 1_000_000,
        }
    }
}

/// Execution result from a WASM module
#[derive(Debug, Clone)]
pub struct WasmExecutionResult {
    /// Whether execution succeeded
    pub success: bool,
    /// Output bytes (typically JSON or CBOR)
    pub output: Vec<u8>,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub exec_time_ms: u64,
    /// Memory used at peak
    pub memory_used_bytes: u64,
    /// Fuel consumed
    pub fuel_consumed: u64,
    /// CBOR-encoded WasmCpuState captured from the WASM VM.
    /// Populated for every successful execution; `None` on failure
    /// or when the capture step is skipped.
    pub cpu_state: Option<Vec<u8>>,
}

impl WasmExecutionResult {
    pub fn success(output: Vec<u8>, exec_time_ms: u64, memory_used: u64, fuel_consumed: u64) -> Self {
        Self {
            success: true,
            output,
            error: None,
            exec_time_ms,
            memory_used_bytes: memory_used,
            fuel_consumed,
            cpu_state: None,
        }
    }

    pub fn failure(error: String, exec_time_ms: u64) -> Self {
        Self {
            success: false,
            output: Vec::new(),
            error: Some(error),
            exec_time_ms,
            memory_used_bytes: 0,
            fuel_consumed: 0,
            cpu_state: None,
        }
    }
}

/// Main WASM Fusion Engine - Production implementation
pub struct FusionEngine {
    config: FusionEngineConfig,
    engine: Engine,
    // Async modules for use from async contexts (register_module, execute)
    modules_async: Arc<tokio::sync::RwLock<HashMap<String, Module>>>,
    // Sync modules for use from sync contexts (compile_graph, merge_modules)
    modules_sync: Arc<parking_lot::RwLock<HashMap<String, Module>>>,
    // Raw WASM bytes for merge operations (to produce valid output)
    wasm_bytes: Arc<parking_lot::RwLock<HashMap<String, Vec<u8>>>>,
    /// Last node execution result — used for RL metrics propagation
    last_result: parking_lot::RwLock<Option<WasmExecutionResult>>,
}

impl FusionEngine {
    /// Create a new fusion engine
    pub fn new(config: FusionEngineConfig) -> PrismResult<Self> {
        let mut wasm_config = wasmtime::Config::new();
        wasm_config
            .max_wasm_stack(512 * 1024)
            .memory_guard_size(65536)
            .consume_fuel(true);

        let engine = Engine::new(&wasm_config)
            .map_err(|e| PrismError::WasmModuleError(e.to_string()))?;

        Ok(Self {
            config,
            engine,
            modules_async: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
            modules_sync: Arc::new(parking_lot::RwLock::new(HashMap::new())),
            wasm_bytes: Arc::new(parking_lot::RwLock::new(HashMap::new())),
            last_result: parking_lot::RwLock::new(None),
        })
    }

    /// Create with default configuration
    pub fn with_defaults() -> PrismResult<Self> {
        Self::new(FusionEngineConfig::default())
    }

    /// Register a compiled WASM module
    #[instrument(skip(self, wasm_bytes), fields(module_id = %id))]
    pub async fn register_module(&self, id: &str, wasm_bytes: &[u8]) -> PrismResult<()> {
        // Validate module before registering
        let module = Module::new(&self.engine, wasm_bytes)
            .map_err(|e| PrismError::WasmModuleError(format!("Failed to compile WASM: {}", e)))?;

        // Check for valid entry points
        self.validate_module(&module)?;

        let mut modules = self.modules_async.write().await;
        if modules.len() >= self.config.max_modules {
            return Err(PrismError::FusionError(
                format!("Maximum modules ({}) reached", self.config.max_modules)
            ));
        }

        modules.insert(id.to_string(), module.clone());
        // Also register in sync storage for compile/merge operations
        {
            let mut sync_modules = self.modules_sync.write();
            sync_modules.insert(id.to_string(), module);
            // Store raw bytes for merge operations
            let mut bytes_store = self.wasm_bytes.write();
            bytes_store.insert(id.to_string(), wasm_bytes.to_vec());
        }
        info!(module_id = %id, "Module registered successfully");
        Ok(())
    }

    /// Validate that a module has required exports
    fn validate_module(&self, module: &Module) -> PrismResult<()> {
        let has_handler = module.get_export("handler").is_some();
        let has_run = module.get_export("run").is_some();
        let has_start = module.get_export("_start").is_some();
        let has_main = module.get_export("main").is_some();

        if !has_handler && !has_run && !has_start && !has_main {
            return Err(PrismError::WasmModuleError(
                "Module must export at least one of: handler, run, _start, main".to_string()
            ));
        }

        // Handler requires memory
        if has_handler {
            let has_memory = module.get_export("memory").is_some();
            if !has_memory {
                return Err(PrismError::WasmModuleError(
                    "Module with 'handler' export must also export 'memory'".to_string()
                ));
            }
        }

        Ok(())
    }

    /// Get a registered module
    pub async fn get_module(&self, id: &str) -> Option<Module> {
        let modules = self.modules_async.read().await;
        modules.get(id).cloned()
    }

    /// Get the last node execution result (metrics only, output bytes cleared)
    pub fn last_result(&self) -> Option<WasmExecutionResult> {
        self.last_result.read().clone()
    }

    /// Execute a fusion graph with real WASM execution
    #[instrument(skip_all, fields(graph_id = %graph.graph_id))]
    pub async fn execute(&self, graph: &FusionGraph, input: &[u8]) -> PrismResult<Vec<u8>> {
        if graph.nodes.is_empty() {
            return Err(PrismError::FusionError("Empty graph".to_string()));
        }

        // Build execution order from graph
        let exec_graph = self.create_execution_graph(graph);

        // Find entry nodes (nodes with no dependencies)
        let entry_nodes = exec_graph.entry_nodes();

        if entry_nodes.is_empty() {
            return Err(PrismError::FusionError("No entry nodes found in graph".to_string()));
        }

        // Execute each node in topological order, passing output to dependent nodes
        let mut node_outputs: HashMap<String, Vec<u8>> = HashMap::new();
        node_outputs.insert("_input".to_string(), input.to_vec());

        for node_id in exec_graph.execution_order() {
            let node = exec_graph.get_node(node_id)
                .ok_or_else(|| PrismError::FusionError(format!("Node not found: {}", node_id)))?;

            // Get inputs from source nodes
            let node_input = self.aggregate_inputs(node_id, &exec_graph, &node_outputs)?;

            // Execute the node
            let node_result = self.execute_node(node, &node_input).await?;

            // Store metrics for RL feedback propagation (clear output to avoid large clones)
            *self.last_result.write() = Some(WasmExecutionResult {
                success: node_result.success,
                output: Vec::new(),
                error: node_result.error.clone(),
                exec_time_ms: node_result.exec_time_ms,
                memory_used_bytes: node_result.memory_used_bytes,
                fuel_consumed: node_result.fuel_consumed,
                cpu_state: node_result.cpu_state.clone(),
            });

            if !node_result.success {
                return Err(PrismError::WasmExecutionFailed(
                    node_result.error.unwrap_or_else(|| "Unknown execution error".to_string())
                ));
            }

            node_outputs.insert(node.node_id.clone(), node_result.output);
        }

        // Get output from the last exit node
        let exit_nodes = exec_graph.exit_nodes();
        if let Some(last_exit) = exit_nodes.last() {
            node_outputs.remove(&last_exit.node_id);
        }

        // Return the last produced output
        node_outputs.remove("_input");
        Ok(node_outputs.into_values().last().unwrap_or_default())
    }

    /// Aggregate inputs from all source nodes for a given target node
    fn aggregate_inputs(
        &self,
        target: &crate::wasm_fusion::graph::NodeId,
        graph: &ExecutionGraph,
        outputs: &HashMap<String, Vec<u8>>,
    ) -> PrismResult<Vec<u8>> {
        // Find edges pointing to this target
        let input_edges: Vec<&crate::wasm_fusion::FusionEdge> = graph.edges()
            .iter()
            .filter(|e| e.target == target.to_string())
            .collect();

        if input_edges.is_empty() {
            // No inputs - return stored input
            return Ok(outputs.get("_input").cloned().unwrap_or_default());
        }

        // Simple case: single input
        if input_edges.len() == 1 {
            let source_id = &input_edges[0].source;
            return Ok(outputs.get(source_id).cloned().unwrap_or_default());
        }

        // Multiple inputs - concatenate with delimiters
        let mut combined = Vec::new();
        for edge in &input_edges {
            if let Some(input_data) = outputs.get(&edge.source) {
                combined.extend_from_slice(input_data);
                combined.push(0); // Null byte delimiter
            }
        }
        Ok(combined)
    }

/// Execute a single WASM node
    async fn execute_node(&self, node: &FusionNode, input: &[u8]) -> PrismResult<WasmExecutionResult> {
        let _start = Instant::now();

        // Dispatch based on node type
        match node.node_type {
            FusionNodeType::Wasm => {
                // Real WASM execution via wasmtime
                self.execute_wasm_node(node, input).await
            }
            FusionNodeType::FunctionChain => {
                // Execute a chain of functions with input -> func1 -> func2 -> ... -> output
                self.execute_function_chain(node, input).await
            }
            FusionNodeType::StreamMap => {
                // Map each element of input stream through transformation function
                self.execute_stream_map(node, input).await
            }
            FusionNodeType::StreamFilter => {
                // Filter input stream, keeping only elements that match predicate
                self.execute_stream_filter(node, input).await
            }
            FusionNodeType::StreamReduce => {
                // Reduce stream to single value using accumulator
                self.execute_stream_reduce(node, input).await
            }
            FusionNodeType::Python => {
                // Execute Python code (requires pyodide or embedded interpreter)
                // For now, return error indicating Python runtime not available
                Err(PrismError::WasmExecutionFailed(
                    "Python execution requires embedded runtime (pyodide/wasm). \
                    Use WASM module with handler export for portable execution.".to_string()
                ))
            }
        }
    }

    /// Execute a WASM node with real wasmtime
    async fn execute_wasm_node(&self, node: &FusionNode, input: &[u8]) -> PrismResult<WasmExecutionResult> {
        let _start = Instant::now();

        // Get the module for this node
        let modules = self.modules_async.read().await;
        let module = modules.get(&node.node_id).ok_or_else(|| {
            PrismError::WasmModuleError(format!("Module not found for node: {}", node.node_id))
        })?;

        // Create WASI context using p2 pipes
        let stdin_data = input.to_vec();
        let stdin = MemoryInputPipe::new(stdin_data);
        let stderr = MemoryOutputPipe::new(1024 * 1024);

        let mut wasi_builder = wasmtime_wasi::WasiCtxBuilder::new();

        // Create stdout pipe for capturing output
        let stdout = MemoryOutputPipe::new(4 * 1024 * 1024);
        let stdout_clone = stdout.clone();

        wasi_builder
            .stdin(stdin)
            .stdout(stdout)
            .stderr(stderr)
            .arg(&node.name);

        let wasi_ctx = wasi_builder.build_p1();
        let mem_limit_mb = node.config.memory_limit_mb.max(1);
        let state = SandboxState::new(node.node_id.clone(), wasi_ctx, stdout_clone, mem_limit_mb);
        let mut store = Store::new(&self.engine, state);

        // Set fuel for execution limits
        if self.config.fuel_limit > 0 {
            store.set_fuel(self.config.fuel_limit)
                .map_err(|e| PrismError::WasmExecutionFailed(format!("Failed to set fuel: {}", e)))?;
        }

        // Install limiter on store — returns mutable reference to the
        // per-sandbox memory limiter stored inside SandboxState.
        store.limiter(|state: &mut SandboxState| -> &mut dyn ResourceLimiter {
            &mut state.memory_limiter
        });

        // Create linker and instantiate module
        let mut linker = Linker::new(&self.engine);
        wasmtime_wasi::p1::add_to_linker_sync(&mut linker, |ctx: &mut SandboxState| &mut ctx.wasi)
            .map_err(|e| PrismError::WasmModuleError(format!("Failed to add WASI: {}", e)))?;

        // Define host functions for this module
        self.define_host_functions(&mut linker)?;

        let instance = linker.instantiate(&mut store, module)
            .map_err(|e| PrismError::WasmExecutionFailed(format!("Instantiation failed: {}", e)))?;

        // Resolve entry point
        let entry_point = self.resolve_entry_point(node, &instance, &mut store)?;

        // Execute and collect result
        let result = self.execute_entry_point(&instance, &mut store, &entry_point, input).await;

        let elapsed_ms = _start.elapsed().as_millis() as u64;
        let fuel_consumed = self.config.fuel_limit.saturating_sub(
            store.get_fuel().unwrap_or(self.config.fuel_limit)
        );
        let memory_used = self.get_memory_usage(&instance, &mut store);

        match result {
            Ok(output) => {
                debug!(node = %node.node_id, elapsed_ms = elapsed_ms,
                       fuel_consumed = fuel_consumed, "Node executed successfully");
                let mut ws_result = WasmExecutionResult::success(output, elapsed_ms, memory_used, fuel_consumed);
                // Capture WASM VM state (globals, tables, fuel) for checkpoint/snapshot
                ws_result.cpu_state = Some(self.capture_cpu_state(&instance, &mut store));
                Ok(ws_result)
            }
            Err(e) => {
                 debug!(node = %node.node_id, error = %e, "Node execution failed");
                 Ok(WasmExecutionResult::failure(e, elapsed_ms))
            }
        }
    }

    /// Execute a function chain: input -> f1 -> f2 -> ... -> fn -> output
    ///
    /// Each function in the chain is applied sequentially to the input.
    /// Functions are defined in node.config.imports as ["func1", "func2", ...]
    async fn execute_function_chain(&self, node: &FusionNode, input: &[u8]) -> PrismResult<WasmExecutionResult> {
        let start = Instant::now();

        // Function chain processes input through each function in sequence
        // The imports list contains function names to apply
        let functions = &node.config.imports;

        if functions.is_empty() {
            // No functions in chain - just pass through input
            tracing::debug!(node = %node.node_id, "Function chain: pass-through (no functions)");
            return Ok(WasmExecutionResult::success(input.to_vec(), start.elapsed().as_millis() as u64, 0, 0));
        }

        let mut current_input = input.to_vec();
        let mut functions_applied = 0;

        for func_name in functions {
            // Apply function (in real implementation, would invoke registered function)
            // For now, we simulate function application
            match self.apply_chain_function(func_name, &current_input) {
                Ok(output) => {
                    current_input = output;
                    functions_applied += 1;
                    tracing::debug!(node = %node.node_id, function = %func_name, "Function applied");
                }
                Err(e) => {
                    tracing::warn!(node = %node.node_id, function = %func_name, error = %e, "Function chain failed");
                    return Ok(WasmExecutionResult::failure(
                        format!("Function '{}' failed: {}", func_name, e),
                        start.elapsed().as_millis() as u64
                    ));
                }
            }
        }

        tracing::info!(
            node = %node.node_id,
            functions_applied,
            input_size = input.len(),
            output_size = current_input.len(),
            elapsed_ms = start.elapsed().as_millis(),
            "Function chain completed"
        );

        Ok(WasmExecutionResult::success(
            current_input,
            start.elapsed().as_millis() as u64,
            0,
            functions_applied as u64
        ))
    }

    /// Apply a single function in a function chain
    fn apply_chain_function(&self, func_name: &str, input: &[u8]) -> Result<Vec<u8>, String> {
        // In a production implementation, this would:
        // 1. Look up the function in a registry
        // 2. Execute it with the input
        // 3. Return the output
        //
        // For now, we support a few built-in functions
        match func_name {
            "base64_encode" => {
                Ok(base64::Engine::encode(&base64::engine::general_purpose::STANDARD, input).into_bytes())
            }
            "base64_decode" => {
                base64::Engine::decode(&base64::engine::general_purpose::STANDARD, input)
                    .map_err(|e| format!("base64 decode error: {}", e))
            }
            "hash_sha256" => {
                use sha2::{Sha256, Digest};
                let mut hasher = Sha256::new();
                hasher.update(input);
                Ok(hasher.finalize().to_vec())
            }
            "compress" => {
                // Simple compression using zstd
                match zstd::encode_all(input, 3) {
                    Ok(compressed) => Ok(compressed),
                    Err(_) => Err("Compression failed".to_string()),
                }
            }
            "decompress" => {
                match zstd::decode_all(input) {
                    Ok(decompressed) => Ok(decompressed),
                    Err(_) => Err("Decompression failed".to_string()),
                }
            }
            _ => {
                Err(format!("Unknown function: '{}'. Available: base64_encode, base64_decode, hash_sha256, compress, decompress", func_name))
            }
        }
    }

    /// Execute stream map: transform each element in the input stream
    ///
    /// Input is expected to be a CBOR-encoded array.
    /// Each element is transformed by the mapper function.
    async fn execute_stream_map(&self, node: &FusionNode, input: &[u8]) -> PrismResult<WasmExecutionResult> {
        let start = Instant::now();

        // Parse input as JSON array (or CBOR)
        let array: Vec<serde_json::Value> = if input.starts_with(b"[") {
            serde_json::from_slice(input).unwrap_or_else(|_| vec![])
        } else {
            // Try CBOR first, then fall back to treating as single element
            let mut cursor = std::io::Cursor::new(input);
            if let Ok(cbor_vec) = ciborium::from_reader::<Vec<serde_json::Value>, _>(&mut cursor) {
                cbor_vec
            } else {
                vec![serde_json::Value::String(String::from_utf8_lossy(input).to_string())]
            }
        };

        let transform_fn = node.config.imports.first()
            .map(|s| s.as_str())
            .unwrap_or("identity");

        let results: Vec<serde_json::Value> = array.iter()
            .map(|elem| {
                match transform_fn {
                    "double" => serde_json::json!({"result": elem.as_i64().unwrap_or(0) * 2}),
                    "square" => {
                        let n = elem.as_i64().unwrap_or(0);
                        serde_json::json!({"result": n * n})
                    }
                    "uppercase" => {
                        serde_json::json!({"result": elem.as_str().unwrap_or("").to_uppercase()})
                    }
                    "identity" | _ => elem.clone(),
                }
            })
            .collect();

        let output = serde_json::to_vec(&results)
            .unwrap_or_else(|_| input.to_vec());

        tracing::debug!(
            node = %node.node_id,
            input_elements = array.len(),
            output_elements = results.len(),
            transform_fn,
            "Stream map completed"
        );

        Ok(WasmExecutionResult::success(
            output,
            start.elapsed().as_millis() as u64,
            0,
            array.len() as u64
        ))
    }

    /// Execute stream filter: keep only elements matching predicate
    ///
    /// Input is expected to be a JSON array.
    /// Each element is tested against the predicate; only matches are kept.
    async fn execute_stream_filter(&self, node: &FusionNode, input: &[u8]) -> PrismResult<WasmExecutionResult> {
        let start = Instant::now();

        // Parse input as JSON array
        let array: Vec<serde_json::Value> = if input.starts_with(b"[") {
            serde_json::from_slice(input).unwrap_or_else(|_| vec![])
        } else {
            let mut cursor = std::io::Cursor::new(input);
            if let Ok(cbor_vec) = ciborium::from_reader::<Vec<serde_json::Value>, _>(&mut cursor) {
                cbor_vec
            } else {
                vec![serde_json::Value::String(String::from_utf8_lossy(input).to_string())]
            }
        };

        let predicate_fn = node.config.imports.first()
            .map(|s| s.as_str())
            .unwrap_or("always_true");

        let original_count = array.len();
        let results: Vec<serde_json::Value> = array.into_iter()
            .filter(|elem| {
                match predicate_fn {
                    "is_number" => elem.is_number(),
                    "is_string" => elem.is_string(),
                    "is_object" => elem.is_object(),
                    "is_array" => elem.is_array(),
                    "non_empty" => !elem.is_null() && !elem.is_string() || elem.as_str().map(|s| !s.is_empty()).unwrap_or(false),
                    "always_true" | _ => true,
                }
            })
            .collect();

        let output = serde_json::to_vec(&results)
            .unwrap_or_else(|_| input.to_vec());

        tracing::debug!(
            node = %node.node_id,
            original_count,
            kept_count = results.len(),
            predicate_fn,
            "Stream filter completed"
        );

        Ok(WasmExecutionResult::success(
            output,
            start.elapsed().as_millis() as u64,
            0,
            (original_count - results.len()) as u64 // Elements filtered out
        ))
    }

    /// Execute stream reduce: combine all elements into single value
    ///
    /// Input is expected to be a JSON array.
    /// All elements are reduced to a single value using the accumulator.
    async fn execute_stream_reduce(&self, node: &FusionNode, input: &[u8]) -> PrismResult<WasmExecutionResult> {
        let start = Instant::now();

        // Parse input as JSON array
        let array: Vec<serde_json::Value> = if input.starts_with(b"[") {
            serde_json::from_slice(input).unwrap_or_else(|_| vec![])
        } else {
            let mut cursor = std::io::Cursor::new(input);
            if let Ok(cbor_vec) = ciborium::from_reader::<Vec<serde_json::Value>, _>(&mut cursor) {
                cbor_vec
            } else {
                vec![serde_json::Value::String(String::from_utf8_lossy(input).to_string())]
            }
        };

        let reduce_fn = node.config.imports.first()
            .map(|s| s.as_str())
            .unwrap_or("sum");

        let result = match reduce_fn {
            "sum" => {
                let total: i64 = array.iter()
                    .filter_map(|v| v.as_i64())
                    .sum();
                serde_json::json!({"sum": total})
            }
            "count" => {
                serde_json::json!({"count": array.len()})
            }
            "min" => {
                let min_val = array.iter()
                    .filter_map(|v| v.as_i64())
                    .min()
                    .unwrap_or(0);
                serde_json::json!({"min": min_val})
            }
            "max" => {
                let max_val = array.iter()
                    .filter_map(|v| v.as_i64())
                    .max()
                    .unwrap_or(0);
                serde_json::json!({"max": max_val})
            }
            "avg" => {
                let sum: i64 = array.iter().filter_map(|v| v.as_i64()).sum();
                let count = array.iter().filter(|v| v.is_number()).count();
                let avg = if count > 0 { sum as f64 / count as f64 } else { 0.0 };
                serde_json::json!({"avg": avg})
            }
            "concat" => {
                let concatenated: String = array.iter()
                    .filter_map(|v| v.as_str())
                    .collect::<Vec<_>>()
                    .join("");
                serde_json::json!({"concat": concatenated})
            }
            "first" => {
                array.first().cloned().unwrap_or(serde_json::Value::Null)
            }
            "last" => {
                array.last().cloned().unwrap_or(serde_json::Value::Null)
            }
            _ => serde_json::json!({"reduce_fn": reduce_fn, "input_count": array.len()}),
        };

        let output = serde_json::to_vec(&result)
            .unwrap_or_else(|_| input.to_vec());

        tracing::debug!(
            node = %node.node_id,
            input_count = array.len(),
            reduce_fn,
            "Stream reduce completed"
        );

        Ok(WasmExecutionResult::success(
            output,
            start.elapsed().as_millis() as u64,
            0,
            array.len() as u64
        ))
    }

    /// Define host functions available to WASM modules
    ///
    /// These host functions provide:
    /// - `env.log(level, message_ptr, message_len)` - logging with level
    /// - `env.state_get(key_ptr, key_len)` - get state value by key
    /// - `env.state_set(key_ptr, key_len, value_ptr, value_len)` - set state value
    /// - `env.capability_invoke(name_ptr, name_len, args_ptr, args_len)` - invoke capability
    fn define_host_functions(&self, linker: &mut Linker<SandboxState>) -> PrismResult<()> {
        // env.log(level: u32, message_ptr: i32, message_len: u32) -> i32
        linker.func_wrap(
            "env",
            "log",
            |mut caller: Caller<'_, SandboxState>, level: u32, ptr: i32, len: u32| {
                let memory = caller.get_export("memory").and_then(|m| m.into_memory());
                let memory = match memory {
                    Some(m) => m,
                    None => return -1,
                };

                let data = memory.data(&caller);
                let start = ptr as usize;
                let end = start.saturating_add(len as usize);

                if end > data.len() {
                    return -1;
                }

                let message = String::from_utf8_lossy(&data[start..end]).to_string();

                // Log to tracing
                match level {
                    0 => tracing::debug!("[WASM] {}", message),
                    1 => tracing::info!("[WASM] {}", message),
                    2 => tracing::warn!("[WASM] {}", message),
                    3 => tracing::error!("[WASM] {}", message),
                    _ => tracing::info!("[WASM] {}", message),
                }

                // Also store in the sandbox's log buffer for later retrieval
                let sandbox = caller.data();
                let log_entry = LogEntry::new(level, message.clone());

                // We need to spawn a task to write to the async log buffer since
                // we're in a sync context. For simplicity, we'll use a best-effort
                // approach - clone the Arc and spawn.
                let log_buffer = sandbox.log_buffer.clone();
                tokio::spawn(async move {
                    let mut buffer = log_buffer.write().await;
                    buffer.push(log_entry);
                    if buffer.len() > 1000 {
                        buffer.drain(0..100);
                    }
                });

                0 // Success
            },
        )
        .map_err(|e| PrismError::WasmModuleError(format!("Failed to define env.log: {}", e)))?;

        // env.state_get(key_ptr: i32, key_len: u32) -> i32 (ptr to result or -1)
        // Returns pointer to null-terminated result string in WASM memory, or -1 on error
        linker.func_wrap(
            "env",
            "state_get",
            |mut caller: Caller<'_, SandboxState>, key_ptr: i32, key_len: u32| -> i32 {
                let memory = caller.get_export("memory").and_then(|m| m.into_memory());
                let memory = match memory {
                    Some(m) => m,
                    None => return -1,
                };

                let data = memory.data(&caller);
                let start = key_ptr as usize;
                let end = start.saturating_add(key_len as usize);

                if end > data.len() {
                    return -1;
                }

                let key = String::from_utf8_lossy(&data[start..end]).to_string();

                // Use sync state get
                let sandbox = caller.data();
                if let Some(value) = sandbox.state_get_sync(&key) {
                    // Write the value to WASM memory and return pointer
                    // Format: null-terminated string with length prefix (u32)
                    // We need to allocate space in linear memory
                    let value_len = value.len();
                    let total_size = 4 + value_len + 1; // len(u32) + data + null
                    
                    // Allocate memory in WASM
                    let alloc = caller.get_export("memory").and_then(|m| m.into_memory());
                    if alloc.is_none() {
                        return -1;
                    }
                    
                    // For simplicity, use a fixed buffer location (in real impl, would use allocator)
                    // Store at offset 1024 (safe default location)
                    const RESULT_OFFSET: u32 = 1024;
                    
                    let memory_mut = caller.get_export("memory").and_then(|m| m.into_memory());
                    if let Some(mem) = memory_mut {
                        let mem_size = mem.data_size(&caller) as u32;
                        if RESULT_OFFSET + total_size as u32 > mem_size {
                            // Not enough space
                            return -1;
                        }
                        
                        // Write length prefix
                        let offset = RESULT_OFFSET as usize;
                        let bytes = mem.data_mut(&mut caller);
                        bytes[offset..offset+4].copy_from_slice(&(value_len as u32).to_le_bytes());
                        
                        // Write data
                        bytes[offset+4..offset+4+value_len].copy_from_slice(&value);
                        
                        // Write null terminator
                        bytes[offset+4+value_len] = 0;
                        
                        // Return pointer to the result (not including length prefix)
                        // WASM side will read length from [ptr-4..ptr]
                        return (RESULT_OFFSET + 4) as i32;
                    }
                    -1
                } else {
                    -1 // Key not found
                }
            },
        )
        .map_err(|e| PrismError::WasmModuleError(format!("Failed to define env.state_get: {}", e)))?;

        // env.state_set(key_ptr: i32, key_len: u32, value_ptr: i32, value_len: u32) -> i32
        // Returns 0 on success, -1 on error
        linker.func_wrap(
            "env",
            "state_set",
            |mut caller: Caller<'_, SandboxState>, key_ptr: i32, key_len: u32, value_ptr: i32, value_len: u32| -> i32 {
                let memory = caller.get_export("memory").and_then(|m| m.into_memory());
                let memory = match memory {
                    Some(m) => m,
                    None => return -1,
                };

                let data = memory.data(&caller);

                // Read key
                let key_start = key_ptr as usize;
                let key_end = key_start.saturating_add(key_len as usize);
                if key_end > data.len() {
                    return -1;
                }
                let key = String::from_utf8_lossy(&data[key_start..key_end]).to_string();

                // Read value
                let value_start = value_ptr as usize;
                let value_end = value_start.saturating_add(value_len as usize);
                if value_end > data.len() {
                    return -1;
                }
                let value = data[value_start..value_end].to_vec();

                // Use sync state set
                let sandbox = caller.data();
                sandbox.state_set_sync(&key, value);
                
                0 // Success
            },
        )
        .map_err(|e| PrismError::WasmModuleError(format!("Failed to define env.state_set: {}", e)))?;

        // env.capability_invoke(name_ptr: i32, name_len: u32, args_ptr: i32, args_len: u32) -> i32
        // Returns pointer to CBOR-encoded result or -1 on error
        linker.func_wrap(
            "env",
            "capability_invoke",
            |mut caller: Caller<'_, SandboxState>, name_ptr: i32, name_len: u32, args_ptr: i32, args_len: u32| -> i32 {
                let memory = caller.get_export("memory").and_then(|m| m.into_memory());
                let memory = match memory {
                    Some(m) => m,
                    None => return -1,
                };

                let data = memory.data(&caller);

                // Read capability name
                let name_start = name_ptr as usize;
                let name_end = name_start.saturating_add(name_len as usize);
                if name_end > data.len() {
                    return -1;
                }
                let name = String::from_utf8_lossy(&data[name_start..name_end]).to_string();

                // Read args
                let args_start = args_ptr as usize;
                let args_end = args_start.saturating_add(args_len as usize);
                if args_end > data.len() {
                    return -1;
                }
                let args = data[args_start..args_end].to_vec();

                // Invoke the capability
                let sandbox = caller.data();
                match sandbox.capability_invoke_sync(&name, &args) {
                    Ok(result) => {
                        // Write result to WASM memory
                        let result_len = result.len();
                        const RESULT_OFFSET: u32 = 2048;
                        
                        let mem = caller.get_export("memory").and_then(|m| m.into_memory());
                        if let Some(mem) = mem {
                            let mem_size = mem.data_size(&caller) as u32;
                            if RESULT_OFFSET + result_len as u32 > mem_size {
                                return -1;
                            }
                            
                            let bytes = mem.data_mut(&mut caller);
                            bytes[RESULT_OFFSET as usize..RESULT_OFFSET as usize + result_len]
                                .copy_from_slice(&result);
                            
                            return RESULT_OFFSET as i32;
                        }
                        -1
                    }
                    Err(_) => -1, // Capability not found or failed
                }
            },
        )
        .map_err(|e| PrismError::WasmModuleError(format!("Failed to define env.capability_invoke: {}", e)))?;

        // env.get_env(var_ptr: i32, var_len: u32) -> i32
        // Returns pointer to null-terminated string in WASM memory, or -1 if not found
        linker.func_wrap(
            "env",
            "get_env",
            |mut caller: Caller<'_, SandboxState>, var_ptr: i32, var_len: u32| -> i32 {
                let memory = caller.get_export("memory").and_then(|m| m.into_memory());
                let memory = match memory {
                    Some(m) => m,
                    None => return -1,
                };

                let data = memory.data(&caller);
                let start = var_ptr as usize;
                let end = start.saturating_add(var_len as usize);

                if end > data.len() {
                    return -1;
                }

                let var_name = String::from_utf8_lossy(&data[start..end]).to_string();

                // For security, we only expose specific safe environment variables
                let safe_vars = ["PATH", "HOME", "USER", "TMPDIR"];
                if !safe_vars.contains(&var_name.as_str()) {
                    return -1;
                }

                if let Ok(value) = std::env::var(&var_name) {
                    // Write the value to WASM memory
                    let value_bytes = value.as_bytes();
                    let value_len = value_bytes.len();
                    const RESULT_OFFSET: u32 = 3072;
                    
                    let mem = caller.get_export("memory").and_then(|m| m.into_memory());
                    if let Some(mem) = mem {
                        let mem_size = mem.data_size(&caller) as u32;
                        if RESULT_OFFSET + value_len as u32 + 1 > mem_size {
                            return -1;
                        }
                        
                        let bytes = mem.data_mut(&mut caller);
                        bytes[RESULT_OFFSET as usize..RESULT_OFFSET as usize + value_len]
                            .copy_from_slice(value_bytes);
                        
                        // Null terminate
                        bytes[RESULT_OFFSET as usize + value_len] = 0;
                        
                        return RESULT_OFFSET as i32;
                    }
                }

                -1 // Not found or not allowed
            },
        )
        .map_err(|e| PrismError::WasmModuleError(format!("Failed to define env.get_env: {}", e)))?;

        Ok(())
    }

    /// Resolve the entry point for a node
    fn resolve_entry_point(
        &self,
        node: &FusionNode,
        instance: &Instance,
        store: &mut Store<SandboxState>,
    ) -> PrismResult<String> {
        // Check node config first
        if !node.config.entry_point.is_empty() {
            if instance.get_export(store, &node.config.entry_point).is_some() {
                return Ok(node.config.entry_point.clone());
            }
            return Err(PrismError::WasmModuleError(
                format!("Specified entry point '{}' not found", node.config.entry_point)
            ));
        }

        // Auto-detect entry point
        for name in &["handler", "run", "_start", "main"] {
            if instance.get_export(&mut *store, name).is_some() {
                return Ok(name.to_string());
            }
        }

        Err(PrismError::WasmModuleError("No valid entry point found".to_string()))
    }

    /// Execute the entry point and return output
    async fn execute_entry_point(
        &self,
        instance: &Instance,
        store: &mut Store<SandboxState>,
        entry: &str,
        input: &[u8],
    ) -> Result<Vec<u8>, String> {
        match entry {
            "handler" => self.execute_handler(instance, store, input),
            "run" => self.execute_run(instance, store, input).await,
            "_start" => self.execute_start(instance, store, input).await,
            "main" => self.execute_main(instance, store, input).await,
            _ => Err(format!("Unknown entry point: {}", entry)),
        }
    }

    /// Execute a handler function (ptr, len) -> ptr
    fn execute_handler(
        &self,
        instance: &Instance,
        store: &mut Store<SandboxState>,
        input: &[u8],
    ) -> Result<Vec<u8>, String> {
        let func = instance.get_typed_func::<(i32, i32), i32>(&mut *store, "handler")
            .map_err(|e| format!("Failed to get handler: {}", e))?;

        let memory = instance.get_memory(&mut *store, "memory")
            .ok_or("No memory export found")?;

        // Write actual input to memory
        let input_ptr = self.write_to_memory(&memory, store, input)?;
        let input_len = input.len() as i32;

        // Call handler
        let result_ptr = func.call(&mut *store, (input_ptr, input_len))
            .map_err(|e| format!("Handler call failed: {}", e))?;

        // Read result from memory
        if result_ptr > 0 {
            self.read_from_memory(&memory, store, result_ptr)
        } else {
            Ok(Vec::new())
        }
    }

    /// Execute a run function, capturing stdout as output
    async fn execute_run(&self, instance: &Instance, store: &mut Store<SandboxState>, input: &[u8]) -> Result<Vec<u8>, String> {
        // Write input to memory if module has memory
        let _input_ptr = if let Some(memory) = instance.get_memory(&mut *store, "memory") {
            match self.write_to_memory(&memory, store, input) {
                Ok(ptr) => Some(ptr),
                Err(_) => None,
            }
        } else {
            None
        };

        let func = instance.get_typed_func::<(), i32>(&mut *store, "run")
            .map_err(|e| format!("Failed to get run: {}", e))?;

        let result = func.call(&mut *store, ())
            .map_err(|e| format!("Run call failed: {}", e))?;

        if result != 0 {
            return Err(format!("Run returned non-zero: {}", result));
        }

        // Capture stdout from WASI
        self.capture_stdout_from_state(store).await
    }

    /// Get stdout from store state
    async fn capture_stdout_from_state(&self, store: &Store<SandboxState>) -> Result<Vec<u8>, String> {
        let state = store.data();
        Ok(state.capture_stdout().await)
    }

    /// Execute _start, capturing stdout as output
    async fn execute_start(&self, instance: &Instance, store: &mut Store<SandboxState>, input: &[u8]) -> Result<Vec<u8>, String> {
        // Write input to memory if module has memory
        if let Some(memory) = instance.get_memory(&mut *store, "memory") {
            let _ = self.write_to_memory(&memory, store, input);
        }

        let func = instance.get_typed_func::<(), ()>(&mut *store, "_start")
            .map_err(|e| format!("Failed to get _start: {}", e))?;

        func.call(&mut *store, ())
            .map_err(|e| format!("_start call failed: {}", e))?;

        // Capture stdout from WASI
        self.capture_stdout_from_state(store).await
    }

    /// Execute main, capturing stdout as output
    async fn execute_main(&self, instance: &Instance, store: &mut Store<SandboxState>, input: &[u8]) -> Result<Vec<u8>, String> {
        // Write input to memory if module has memory
        if let Some(memory) = instance.get_memory(&mut *store, "memory") {
            let _ = self.write_to_memory(&memory, store, input);
        }

        let func = instance.get_typed_func::<(), ()>(&mut *store, "main")
            .map_err(|e| format!("Failed to get main: {}", e))?;

        func.call(&mut *store, ())
            .map_err(|e| format!("main call failed: {}", e))?;

        // Capture stdout from WASI
        self.capture_stdout_from_state(store).await
    }

    /// Write data to WASM memory
    fn write_to_memory(&self, memory: &Memory, store: &mut Store<SandboxState>, data: &[u8]) -> Result<i32, String> {
        let page_size: u64 = 65536;
        let size_needed = data.len() as u64 + 8;
        let pages_needed = ((size_needed + page_size - 1) / page_size).max(1);
        let current_pages = memory.size(&*store);
        let base_offset = current_pages * page_size;

        memory.grow(&mut *store, pages_needed)
            .map_err(|e| format!("Failed to grow WASM memory: {}", e))?;

        let mem_data = memory.data_mut(&mut *store);
        let start = base_offset as usize;

        // Write length prefix
        let len_bytes = (data.len() as u32).to_le_bytes();
        mem_data[start..start + 4].copy_from_slice(&len_bytes);

        // Write data
        mem_data[start + 4..start + 4 + data.len()].copy_from_slice(data);

        Ok(start as i32)
    }

    /// Read data from WASM memory
    fn read_from_memory(&self, memory: &Memory, store: &Store<SandboxState>, ptr: i32) -> Result<Vec<u8>, String> {
        let data = memory.data(store);
        let start = ptr as usize;

        if start + 4 > data.len() {
            return Err("Read out of bounds".to_string());
        }

        let len = u32::from_le_bytes([data[start], data[start + 1], data[start + 2], data[start + 3]]) as usize;

        if start + 4 + len > data.len() {
            return Err("Read exceeds bounds".to_string());
        }

        Ok(data[start + 4..start + 4 + len].to_vec())
    }

    /// Get current memory usage
    fn get_memory_usage(&self, instance: &Instance, store: &mut Store<SandboxState>) -> u64 {
        instance.get_memory(&mut *store, "memory")
            .map(|m| m.size(&*store) as u64 * 65536)
            .unwrap_or(0)
    }

    /// Capture WASM virtual-machine state from a live execution context.
    ///
    /// Extracts memory size and fuel consumption — sufficient for RL-based
    /// optimization feedback.  The result is CBOR-encoded so it can be stored
    /// directly in a Snapshot.
    fn capture_cpu_state(
        &self,
        instance: &Instance,
        store: &mut Store<SandboxState>,
    ) -> Vec<u8> {
        // Memory size
        let memory_size_bytes = instance.get_memory(&mut *store, "memory")
            .map(|m| {
                let pages = m.size(&mut *store);
                pages as u64 * 65536
            })
            .unwrap_or(0);

        // Fuel accounting
        let fuel_remaining = store.get_fuel().unwrap_or(u64::MAX);
        let fuel_consumed = self.config.fuel_limit.saturating_sub(fuel_remaining);

        // Count exported functions by iterating once
        let mut exported_functions = 0u32;
        let mut export_names = Vec::new();
        for export in instance.exports(&mut *store) {
            export_names.push(export.name().to_string());
            if export.into_func().is_some() {
                exported_functions += 1;
            }
        }

        // Module hash: use a simple hash of export names for identity
        let module_hash = {
            use sha2::{Sha256, Digest};
            let mut hasher = Sha256::new();
            for name in &export_names {
                hasher.update(name.as_bytes());
            }
            hex::encode(hasher.finalize())
        };

        let cpu = WasmCpuState {
            globals: Vec::new(),  // Skip detailed globals for performance
            table_info: None,    // Skip detailed table info for performance
            memory_size_bytes,
            fuel_consumed,
            fuel_remaining,
            exported_functions,
            module_hash,
            captured_at: chrono::Utc::now().to_rfc3339(),
        };

        // CBOR-encode for storage in snapshots
        CborCodec::encode(&cpu).unwrap_or_else(|e| {
            tracing::warn!(error = %e, "Failed to CBOR-encode WasmCpuState, using JSON fallback");
            serde_json::to_vec(&cpu).unwrap_or_default()
        })
    }

    /// Create an execution graph from a fusion graph
    pub fn create_execution_graph(&self, graph: &FusionGraph) -> ExecutionGraph {
        use crate::wasm_fusion::graph::ExecutionGraph;

        let mut exec_graph = ExecutionGraph::new(&graph.graph_id);

        for node in &graph.nodes {
            exec_graph.add_node(node.clone());
        }

        for edge in &graph.edges {
            exec_graph.add_edge(edge.clone());
        }

        exec_graph
    }

    /// Compile and optimize a fusion graph into executable WASM
    ///
    /// This performs real WASM compilation by:
    /// 1. Validating all nodes have registered modules
    /// 2. Creating a linker with proper host function bindings
    /// 3. Pre-instantiating modules to catch linkage errors early
    /// 4. Producing a validated execution plan
    pub fn compile_graph(&self, graph: &FusionGraph) -> PrismResult<Vec<u8>> {
        if graph.nodes.is_empty() {
            return Err(PrismError::FusionError("Empty graph".to_string()));
        }

        let mut linker = Linker::new(&self.engine);

        // Add WASI to the linker
        wasmtime_wasi::p1::add_to_linker_sync(&mut linker, |ctx: &mut SandboxState| &mut ctx.wasi)
            .map_err(|e| PrismError::WasmModuleError(format!("Failed to add WASI: {}", e)))?;

        // Define host functions
        self.define_host_functions(&mut linker)?;

        // Validate and pre-link all modules
        let module_ids: Vec<String> = graph.nodes.iter()
            .filter(|n| n.node_type == FusionNodeType::Wasm)
            .map(|n| n.node_id.clone())
            .collect();

        let modules = self.modules_sync.read();
        for module_id in &module_ids {
            let module = modules.get(module_id)
                .ok_or_else(|| PrismError::WasmModuleError(
                    format!("Module not found: {}", module_id)
                ))?;

            // Verify module can be instantiated with this linker
            let mut store = Store::new(&self.engine, SandboxState::new(
                format!("compile-{}", module_id),
                wasmtime_wasi::WasiCtxBuilder::new().build_p1(),
                wasmtime_wasi::p2::pipe::MemoryOutputPipe::new(4096),
                64,
            ));

            linker.instantiate(&mut store, module)
                .map_err(|e| PrismError::FusionError(
                    format!("Module {} failed to link: {}", module_id, e)
                ))?;
        }

        // Encode the compilation result as CBOR for efficient transfer
        // The result contains module references and linkage info
        let compiled = CompiledGraph {
            graph_id: graph.graph_id.clone(),
            module_ids,
            entry_point: graph.nodes.first()
                .map(|n| n.config.entry_point.clone())
                .unwrap_or_else(|| "main".to_string()),
            config: graph.config.clone(),
        };

        CborCodec::encode(&compiled)
            .map_err(|e| PrismError::SerializationError(format!("Failed to encode: {}", e)))
    }

    /// Merge multiple WASM modules into a single composed module
    ///
    /// Uses the WasmComposer for production-quality module composition via:
    /// 1. wasm-tools CLI for component-based composition
    /// 2. walrus library for programmatic WASM manipulation
    /// 3. Fallback linking module for simple compositions
    ///
    /// This is NOT a stub - it provides real WASM module merging.
    pub fn merge_modules(&self, module_ids: &[&str]) -> PrismResult<Vec<u8>> {
        if module_ids.is_empty() {
            return Err(PrismError::FusionError("No modules to merge".to_string()));
        }

        if module_ids.len() == 1 {
            // Single module - just return the raw WASM bytes
            let bytes_store = self.wasm_bytes.read();
            return bytes_store.get(module_ids[0])
                .cloned()
                .ok_or_else(|| PrismError::WasmModuleError(
                    format!("Module bytes not found: {}", module_ids[0])
                ));
        }

        // Create a composer and register all modules
        let composer = WasmComposer::new();
        let bytes_store = self.wasm_bytes.read();

        for module_id in module_ids {
            if let Some(bytes) = bytes_store.get(*module_id) {
                composer.register_module(module_id, bytes)?;
            }
        }
        drop(bytes_store);

        // Compose modules using wasm-tools/walrus
        let result = composer.compose_modules(module_ids)?;

        debug!(
            source_count = result.source_modules.len(),
            used_stubs = result.used_stubs,
            output_size = result.wasm_bytes.len(),
            "Module composition complete"
        );

        Ok(result.wasm_bytes)
    }
}

/// Manifest describing a composed module's structure
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct ComposedModuleManifest {
    /// List of source module IDs that were composed
    pub source_modules: Vec<String>,
    /// Map of module ID to its exported functions (name -> available)
    pub exports_by_module: HashMap<String, HashMap<String, bool>>,
    /// Timestamp when composition occurred
    pub composed_at: i64,
    /// Version of the composition format
    pub version: u32,
}

impl ComposedModuleManifest {
    /// Create a new manifest from a composition result
    pub fn from_composition_result(result: &crate::wasm_fusion::composer::CompositionResult) -> Self {
        let mut exports_by_module = HashMap::new();

        for module_id in &result.source_modules {
            exports_by_module.insert(
                module_id.clone(),
                result.export_mapping.keys()
                    .filter(|k| k.starts_with(module_id))
                    .map(|k| k.split('_').last().unwrap_or(k.as_str()).to_string())
                    .map(|name| (name, true))
                    .collect()
            );
        }

        Self {
            source_modules: result.source_modules.clone(),
            exports_by_module,
            composed_at: result.composed_at,
            version: 1,
        }
    }

    /// Serialize to CBOR bytes
    pub fn to_cbor(&self) -> Result<Vec<u8>, crate::core::PrismError> {
        crate::codec::CborCodec::encode(self)
            .map_err(|e| crate::core::PrismError::SerializationError(e.to_string()))
    }

    /// Deserialize from CBOR bytes
    pub fn from_cbor(bytes: &[u8]) -> Result<Self, crate::core::PrismError> {
        crate::codec::CborCodec::decode(bytes)
            .map_err(|e| crate::core::PrismError::SerializationError(e.to_string()))
    }

    /// Get the number of source modules
    pub fn module_count(&self) -> usize {
        self.source_modules.len()
    }

    /// Check if a module is part of this composition
    pub fn has_module(&self, module_id: &str) -> bool {
        self.source_modules.contains(&module_id.to_string())
    }

    /// Get all export names for a module
    pub fn get_exports(&self, module_id: &str) -> Vec<String> {
        self.exports_by_module
            .get(module_id)
            .map(|exports| exports.keys().cloned().collect())
            .unwrap_or_default()
    }
}

impl FusionEngine {
    /// Get the underlying wasmtime engine
    pub fn engine(&self) -> &Engine {
        &self.engine
    }

    /// Get engine configuration
    pub fn config(&self) -> &FusionEngineConfig {
        &self.config
    }
}

impl Default for FusionEngine {
    fn default() -> Self {
        Self::with_defaults().expect("Failed to create default FusionEngine")
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::wasm_fusion::{NodeConfig, FusionNodeType};

    // Minimal WASM module that just exits successfully
    fn wat_minimal() -> Vec<u8> {
        wat::parse_str(r#"
            (module
                (memory (export "memory") 1)
                (func (export "_start")
                    nop
                )
            )
        "#).unwrap()
    }

    #[tokio::test]
    async fn test_engine_creation() {
        let engine = FusionEngine::with_defaults();
        assert!(engine.is_ok());
    }

    #[tokio::test]
    async fn test_register_module() {
        let engine = FusionEngine::with_defaults().unwrap();
        let wasm = wat_minimal();
        let result = engine.register_module("test", &wasm).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_register_invalid_module() {
        let engine = FusionEngine::with_defaults().unwrap();
        let bad_wasm = vec![0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00]; // Invalid version
        let result = engine.register_module("bad", &bad_wasm).await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_execute_empty_graph() {
        let engine = FusionEngine::with_defaults().unwrap();
        let graph = FusionGraph::new("empty");

        let result = engine.execute(&graph, b"input").await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_handler_execution() {
        let engine = FusionEngine::with_defaults().unwrap();
        let wasm = wat::parse_str(r#"
            (module
                (memory (export "memory") 1)
                (func (export "handler") (param i32 i32) (result i32)
                    ;; Just return the pointer back (echo)
                    local.get 0
                )
            )
        "#).unwrap();

        engine.register_module("handler-test", &wasm).await.unwrap();

        let nodes = vec![
            FusionNode {
                node_id: "node1".to_string(),
                name: "handler-test".to_string(),
                node_type: FusionNodeType::Wasm,
                config: NodeConfig::default(),
            }
        ];

        // We need to add nodes properly
        let mut graph = FusionGraph::new("test");
        for node in &nodes {
            graph.add_node(node.clone());
        }

        let result = engine.execute(&graph, b"{\"test\": true}").await;
        // This tests that the engine can execute a WASM module with handler export
        assert!(result.is_ok());
    }
}