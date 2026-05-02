mod memory_limiter;
mod host_functions;

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Instant;

use parking_lot::RwLock;
use serde::{Deserialize, Serialize};
use tracing::{info, warn, debug, instrument};
use uuid::Uuid;
use wasmtime::Store;

use crate::engine::{NodeId, NodeExecutionError};

use self::memory_limiter::{install_memory_limiter, with_limiter};

// ---------------------------------------------------------------------------
// Store data
// ---------------------------------------------------------------------------

pub struct SandboxState {
    pub wasi: wasmtime_wasi::p1::WasiP1Ctx,
}

// ---------------------------------------------------------------------------
// Sandbox config
// ---------------------------------------------------------------------------

#[derive(Clone, Debug)]
pub struct SandboxConfig {
    pub allowed_env_vars: Vec<String>,
}

impl Default for SandboxConfig {
    fn default() -> Self {
        Self {
            allowed_env_vars: vec![
                "PATH".to_string(),
                "HOME".to_string(),
                "PWD".to_string(),
                "TMPDIR".to_string(),
            ],
        }
    }
}

// ---------------------------------------------------------------------------
// Execution result
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionResult {
    pub output: serde_json::Value,
    pub success: bool,
    pub error: Option<String>,
    pub exec_time_ms: u64,
    pub memory_used_bytes: u64,
    pub fuel_consumed: u64,
}

impl ExecutionResult {
    pub fn success(output: serde_json::Value, exec_time_ms: u64, memory_used_bytes: u64, fuel_consumed: u64) -> Self {
        Self {
            output,
            success: true,
            error: None,
            exec_time_ms,
            memory_used_bytes,
            fuel_consumed,
        }
    }

    pub fn failure(error: String, exec_time_ms: u64) -> Self {
        Self {
            output: serde_json::Value::Null,
            success: false,
            error: Some(error),
            exec_time_ms,
            memory_used_bytes: 0,
            fuel_consumed: 0,
        }
    }
}

// ---------------------------------------------------------------------------
// WasmCell
// ---------------------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct WasmCell {
    pub id: Uuid,
    pub node_id: NodeId,
    pub name: String,
    pub wasm_bytes: Vec<u8>,
    pub memory_limit_mb: u32,
    pub compute_limit_ms: u64,
    pub fuel_limit: u64,
    pub entry_point: Option<String>,
}

impl WasmCell {
    pub fn new(name: String, wasm_bytes: Vec<u8>) -> Self {
        Self {
            id: Uuid::new_v4(),
            node_id: NodeId(Uuid::nil()),
            name,
            wasm_bytes,
            memory_limit_mb: 64,
            compute_limit_ms: 5000,
            fuel_limit: 1_000_000,
            entry_point: None,
        }
    }

    pub fn with_memory_limit(mut self, mb: u32) -> Self {
        self.memory_limit_mb = mb;
        self
    }

    pub fn with_compute_limit(mut self, ms: u64) -> Self {
        self.compute_limit_ms = ms;
        self
    }

    pub fn with_fuel_limit(mut self, fuel: u64) -> Self {
        self.fuel_limit = fuel;
        self
    }

    pub fn with_entry_point(mut self, entry: String) -> Self {
        self.entry_point = Some(entry);
        self
    }
}

// ---------------------------------------------------------------------------
// WasmSandbox
// ---------------------------------------------------------------------------

pub struct WasmSandbox {
    #[cfg(feature = "wasm-sandbox")]
    engine: wasmtime::Engine,
    #[cfg(feature = "wasm-sandbox")]
    linker: wasmtime::Linker<SandboxState>,
    #[cfg(feature = "wasm-sandbox")]
    cells: Arc<RwLock<HashMap<Uuid, wasmtime::Module>>>,
}

impl WasmSandbox {
    pub fn new() -> anyhow::Result<Self> {
        Self::with_config(SandboxConfig::default())
    }

    #[cfg(feature = "wasm-sandbox")]
    pub fn with_config(config: SandboxConfig) -> anyhow::Result<Self> {
        let mut engine_config = wasmtime::Config::new();
        engine_config.consume_fuel(true);
        let engine = wasmtime::Engine::new(&engine_config)?;

        let mut linker = wasmtime::Linker::<SandboxState>::new(&engine);
        wasmtime_wasi::p1::add_to_linker_sync(&mut linker, |ctx| &mut ctx.wasi)?;
        host_functions::register_host_functions(&mut linker, &config.allowed_env_vars)?;

        Ok(Self {
            engine,
            linker,
            cells: Arc::new(RwLock::new(HashMap::new())),
        })
    }

    #[cfg(feature = "wasm-sandbox")]
    pub fn register_cell(&self, cell: &WasmCell) -> anyhow::Result<()> {
        let module = wasmtime::Module::new(&self.engine, &cell.wasm_bytes)?;
        Self::validate_module(&module)?;
        let mut cells = self.cells.write();
        cells.insert(cell.id, module);
        info!(cell_id = %cell.id, name = %cell.name, "WASM cell registered");
        Ok(())
    }

    #[cfg(feature = "wasm-sandbox")]
    pub fn unregister_cell(&self, cell_id: Uuid) {
        let mut cells = self.cells.write();
        cells.remove(&cell_id);
    }

    #[cfg(feature = "wasm-sandbox")]
    fn validate_module(module: &wasmtime::Module) -> anyhow::Result<()> {
        let has_handler = module.get_export("handler").is_some();
        let has_run = module.get_export("run").is_some();
        let has_start = module.get_export("_start").is_some();
        let has_main = module.get_export("main").is_some();

        if !has_handler && !has_run && !has_start && !has_main {
            return Err(anyhow::anyhow!(
                "WASM module must export at least one of: handler, run, _start, main"
            ));
        }

        if has_handler {
            let has_memory = module.get_export("memory").is_some();
            if !has_memory {
                return Err(anyhow::anyhow!(
                    "WASM module with 'handler' export must also export 'memory'"
                ));
            }
        }

        Ok(())
    }

    #[cfg(feature = "wasm-sandbox")]
    #[instrument(skip_all, fields(cell_id = %cell.id, cell_name = %cell.name))]
    pub async fn execute(
        &self,
        cell: &WasmCell,
        input: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let result = self.execute_with_metrics(cell, input).await?;
        if result.success {
            Ok(result.output)
        } else {
            Err(NodeExecutionError::non_retryable(
                cell.node_id,
                result.error.unwrap_or_else(|| "Unknown execution error".to_string()),
            ))
        }
    }

    #[cfg(feature = "wasm-sandbox")]
    pub async fn execute_with_metrics(
        &self,
        cell: &WasmCell,
        input: HashMap<String, serde_json::Value>,
    ) -> Result<ExecutionResult, NodeExecutionError> {
        let module = {
            let cells = self.cells.read();
            cells.get(&cell.id).cloned()
        };

        let module = module.ok_or_else(|| {
            NodeExecutionError::non_retryable(cell.node_id, format!("Cell {} not registered", cell.id))
        })?;

        let input_json = serde_json::to_string(&input).unwrap_or_else(|_| "{}".to_string());
        let compute_limit = cell.compute_limit_ms;
        let cell_id = cell.id;

        let start = Instant::now();
        let result = tokio::time::timeout(
            std::time::Duration::from_millis(compute_limit),
            async { self.execute_inner(&module, cell, &input_json) },
        )
        .await;

        let elapsed_ms = start.elapsed().as_millis() as u64;

        match result {
            Ok(Ok(exec_result)) => Ok(exec_result),
            Ok(Err(e)) => {
                warn!(cell_id = %cell_id, error = %e, "WASM execution failed");
                Ok(ExecutionResult::failure(e, elapsed_ms))
            }
            Err(_) => {
                warn!(cell_id = %cell_id, timeout_ms = compute_limit, "WASM execution timed out");
                Ok(ExecutionResult::failure(
                    format!("Execution timed out after {}ms", compute_limit),
                    elapsed_ms,
                ))
            }
        }
    }

    #[cfg(feature = "wasm-sandbox")]
    fn execute_inner(
        &self,
        module: &wasmtime::Module,
        cell: &WasmCell,
        input_json: &str,
    ) -> Result<ExecutionResult, String> {
        use wasmtime_wasi::p2::pipe::{MemoryInputPipe, MemoryOutputPipe};

        let start = Instant::now();

        let stdin = MemoryInputPipe::new(input_json.as_bytes().to_vec());
        let stdout = MemoryOutputPipe::new(4 * 1024 * 1024);
        let stderr = MemoryOutputPipe::new(1024 * 1024);

        let mut builder = wasmtime_wasi::WasiCtxBuilder::new();
        builder
            .stdin(stdin)
            .stdout(stdout.clone())
            .stderr(stderr.clone())
            .arg(&cell.name);
        let wasi_ctx = builder.build_p1();

        let state = SandboxState { wasi: wasi_ctx };
        let mut store = Store::new(&self.engine, state);

        let _limiter_guard = install_memory_limiter(cell.memory_limit_mb);
        store.limiter(|_data| unsafe { with_limiter(|l| l) });

        if cell.fuel_limit > 0 {
            store.set_fuel(cell.fuel_limit).map_err(|e| format!("Failed to set fuel: {}", e))?;
        }

        let instance = self
            .linker
            .instantiate(&mut store, module)
            .map_err(|e| format!("WASM instantiation failed: {}", e))?;

        let entry = self.resolve_entry_point(cell, &instance, &mut store)?;

        let result_ptr: Option<i32> = match entry.as_str() {
            "handler" => {
                let func = instance
                    .get_typed_func::<(i32, i32), i32>(&mut store, "handler")
                    .map_err(|e| format!("Failed to get handler function: {}", e))?;

                let memory = instance
                    .get_memory(&mut store, "memory")
                    .ok_or("No memory export found")?;

                let input_ptr = Self::write_to_wasm_memory(&memory, &mut store, input_json.as_bytes())?;
                let input_len = input_json.len() as i32;

                let ptr = func
                    .call(&mut store, (input_ptr, input_len))
                    .map_err(|e| format!("Handler execution failed: {}", e))?;

                Some(ptr)
            }
            "run" => {
                let func = instance
                    .get_typed_func::<(), i32>(&mut store, "run")
                    .map_err(|e| format!("Failed to get run function: {}", e))?;

                let ret = func
                    .call(&mut store, ())
                    .map_err(|e| format!("Run execution failed: {}", e))?;

                if ret != 0 {
                    let stderr_bytes = stderr.contents();
                    let stderr_str = String::from_utf8_lossy(&stderr_bytes);
                    return Err(format!("Run returned non-zero exit code {}: {}", ret, stderr_str));
                }
                None
            }
            "_start" => {
                let func = instance
                    .get_typed_func::<(), ()>(&mut store, "_start")
                    .map_err(|e| format!("Failed to get _start function: {}", e))?;

                func.call(&mut store, ())
                    .map_err(|e| format!("_start execution failed: {}", e))?;
                None
            }
            "main" => {
                let func = instance
                    .get_typed_func::<(), ()>(&mut store, "main")
                    .map_err(|e| format!("Failed to get main function: {}", e))?;

                func.call(&mut store, ())
                    .map_err(|e| format!("Main execution failed: {}", e))?;
                None
            }
            other => {
                return Err(format!("Unsupported entry point: {}", other));
            }
        };

        let elapsed_ms = start.elapsed().as_millis() as u64;
        let fuel_consumed = cell.fuel_limit.saturating_sub(store.get_fuel().unwrap_or(0));
        let memory_used = Self::get_memory_usage(&instance, &mut store);

        if let Some(ptr) = result_ptr {
            if ptr > 0 {
                if let Some(memory) = instance.get_memory(&mut store, "memory") {
                    match Self::read_handler_result(&memory, &store, ptr) {
                        Ok(output_str) if !output_str.is_empty() => {
                            let output = Self::parse_output(&output_str);
                            return Ok(ExecutionResult::success(output, elapsed_ms, memory_used, fuel_consumed));
                        }
                        Ok(_) => {}
                        Err(e) => debug!("Could not read handler result from memory: {}", e),
                    }
                }
            } else if ptr < 0 {
                let stderr_bytes = stderr.contents();
                let stderr_str = String::from_utf8_lossy(&stderr_bytes);
                return Err(format!(
                    "Handler returned error indicator ({}). stderr: {}",
                    ptr, stderr_str
                ));
            }
        }

        let stdout_bytes = stdout.contents();
        let stderr_bytes = stderr.contents();

        if !stdout_bytes.is_empty() {
            let output_str = String::from_utf8_lossy(&stdout_bytes);
            let output = Self::parse_output(&output_str);
            Ok(ExecutionResult::success(output, elapsed_ms, memory_used, fuel_consumed))
        } else if !stderr_bytes.is_empty() {
            let stderr_str = String::from_utf8_lossy(&stderr_bytes);
            Err(format!("WASM stderr: {}", stderr_str))
        } else {
            Ok(ExecutionResult::success(
                serde_json::Value::Null,
                elapsed_ms,
                memory_used,
                fuel_consumed,
            ))
        }
    }

    #[cfg(feature = "wasm-sandbox")]
    fn resolve_entry_point(
        &self,
        cell: &WasmCell,
        instance: &wasmtime::Instance,
        store: &mut wasmtime::Store<SandboxState>,
    ) -> Result<String, String> {
        if let Some(ref entry) = cell.entry_point {
            if instance.get_export(&mut *store, entry).is_some() {
                return Ok(entry.clone());
            }
            return Err(format!("Specified entry point '{}' not found in module", entry));
        }

        for name in &["handler", "run", "_start", "main"] {
            if instance.get_export(&mut *store, name).is_some() {
                return Ok(name.to_string());
            }
        }

        Err("No valid entry point found in WASM module".to_string())
    }

    #[cfg(feature = "wasm-sandbox")]
    fn write_to_wasm_memory(
        memory: &wasmtime::Memory,
        store: &mut Store<SandboxState>,
        data: &[u8],
    ) -> Result<i32, String> {
        let page_size: u64 = 65536;
        let size_needed = data.len() as u64 + 8;
        let pages_needed = ((size_needed + page_size - 1) / page_size).max(1);
        let current_pages = memory.size(&*store);
        let base_offset = current_pages * page_size;

        memory
            .grow(&mut *store, pages_needed)
            .map_err(|e| format!("Failed to grow WASM memory: {}", e))?;

        let mem_data = memory.data_mut(&mut *store);
        let start = base_offset as usize;

        if start + 4 + data.len() > mem_data.len() {
            return Err("WASM memory allocation out of bounds".to_string());
        }

        let len_bytes = (data.len() as u32).to_le_bytes();
        mem_data[start..start + 4].copy_from_slice(&len_bytes);
        mem_data[start + 4..start + 4 + data.len()].copy_from_slice(data);

        Ok(start as i32)
    }

    #[cfg(feature = "wasm-sandbox")]
    fn read_handler_result(
        memory: &wasmtime::Memory,
        store: &Store<SandboxState>,
        ptr: i32,
    ) -> Result<String, String> {
        let data = memory.data(store);
        let start = ptr as usize;

        if start + 4 > data.len() {
            return Err("Result pointer out of bounds".to_string());
        }

        let len = u32::from_le_bytes([data[start], data[start + 1], data[start + 2], data[start + 3]]) as usize;

        if start + 4 + len > data.len() {
            return Err("Result data exceeds memory bounds".to_string());
        }

        String::from_utf8(data[start + 4..start + 4 + len].to_vec())
            .map_err(|e| format!("Result is not valid UTF-8: {}", e))
    }

    #[cfg(feature = "wasm-sandbox")]
    fn parse_output(raw: &str) -> serde_json::Value {
        let trimmed = raw.trim();
        if trimmed.is_empty() {
            return serde_json::Value::Null;
        }
        serde_json::from_str(trimmed).unwrap_or_else(|_| serde_json::Value::String(trimmed.to_string()))
    }

    #[cfg(feature = "wasm-sandbox")]
    fn get_memory_usage(instance: &wasmtime::Instance, store: &mut wasmtime::Store<SandboxState>) -> u64 {
        instance
            .get_memory(&mut *store, "memory")
            .map(|m| m.size(&*store) * 65536)
            .unwrap_or(0)
    }
}

// ---------------------------------------------------------------------------
// Non-feature stubs
// ---------------------------------------------------------------------------

#[cfg(not(feature = "wasm-sandbox"))]
impl WasmSandbox {
    pub fn new() -> anyhow::Result<Self> {
        Ok(Self {})
    }

    pub fn with_config(_config: SandboxConfig) -> anyhow::Result<Self> {
        Ok(Self {})
    }

    pub fn register_cell(&self, _cell: &WasmCell) -> anyhow::Result<()> {
        Err(anyhow::anyhow!("WASM support not enabled"))
    }

    pub fn unregister_cell(&self, _cell_id: Uuid) {}

    pub async fn execute(
        &self,
        cell: &WasmCell,
        _input: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        Err(NodeExecutionError::non_retryable(
            cell.node_id,
            "WASM support not enabled".to_string(),
        ))
    }

    pub async fn execute_with_metrics(
        &self,
        cell: &WasmCell,
        _input: HashMap<String, serde_json::Value>,
    ) -> Result<ExecutionResult, NodeExecutionError> {
        Err(NodeExecutionError::non_retryable(
            cell.node_id,
            "WASM support not enabled".to_string(),
        ))
    }
}

// ---------------------------------------------------------------------------
// CellPool
// ---------------------------------------------------------------------------

pub struct CellPool {
    cells: Arc<RwLock<HashMap<Uuid, WasmCell>>>,
    max_cells: usize,
}

impl CellPool {
    pub fn new(max_cells: usize) -> Self {
        Self {
            cells: Arc::new(RwLock::new(HashMap::new())),
            max_cells,
        }
    }

    pub fn register(&self, cell: WasmCell) -> anyhow::Result<()> {
        let mut cells = self.cells.write();
        if cells.len() >= self.max_cells {
            return Err(anyhow::anyhow!("Cell pool exhausted (max {})", self.max_cells));
        }
        cells.insert(cell.id, cell);
        Ok(())
    }

    pub fn unregister(&self, cell_id: Uuid) {
        let mut cells = self.cells.write();
        cells.remove(&cell_id);
    }

    pub fn get(&self, cell_id: Uuid) -> Option<WasmCell> {
        self.cells.read().get(&cell_id).cloned()
    }

    pub fn len(&self) -> usize {
        self.cells.read().len()
    }

    pub fn is_empty(&self) -> bool {
        self.cells.read().is_empty()
    }
}

impl Default for CellPool {
    fn default() -> Self {
        Self::new(10_000)
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    fn wat_minimal() -> Vec<u8> {
        br#"
            (module
                (memory (export "memory") 1)
                (func (export "_start")
                    nop
                )
            )
        "#.to_vec()
    }

    fn wat_hello_stdout() -> Vec<u8> {
        br#"
            (module
                (memory (export "memory") 1)
                (import "wasi_snapshot_preview1" "fd_write"
                    (func $fd_write (param i32 i32 i32 i32) (result i32)))
                (func (export "_start")
                    (i32.store (i32.const 0) (i32.const 8))
                    (i32.store (i32.const 4) (i32.const 5))
                    (memory.copy (i32.const 8) (i32.const 100) (i32.const 5))
                    (i32.store8 (i32.const 100) (i32.const 72))
                    (i32.store8 (i32.const 101) (i32.const 101))
                    (i32.store8 (i32.const 102) (i32.const 108))
                    (i32.store8 (i32.const 103) (i32.const 108))
                    (i32.store8 (i32.const 104) (i32.const 111))
                    (drop (call $fd_write (i32.const 1) (i32.const 0) (i32.const 1) (i32.const 32)))
                )
            )
        "#.to_vec()
    }

    fn wat_run_export() -> Vec<u8> {
        br#"
            (module
                (memory (export "memory") 1)
                (func (export "run") (result i32)
                    i32.const 0
                )
            )
        "#.to_vec()
    }

    fn wat_handler_echo() -> Vec<u8> {
        br#"
            (module
                (memory (export "memory") 1)
                (global $heap (mut i32) (i32.const 1024))
                (func (export "handler") (param $ptr i32) (param $len i32) (result i32)
                    (local $out_ptr i32)
                    (local $out_len i32)
                    (local.set $out_ptr (global.get $heap))
                    (local.set $out_len (i32.add (local.get $len) (i32.const 4)))
                    (global.set $heap (i32.add (local.get $out_ptr) (local.get $out_len)))
                    (i32.store (local.get $out_ptr) (local.get $len))
                    (memory.copy
                        (i32.add (local.get $out_ptr) (i32.const 4))
                        (i32.add (local.get $ptr) (i32.const 4))
                        (local.get $len))
                    (local.get $out_ptr)
                )
            )
        "#.to_vec()
    }

    #[test]
    fn test_sandbox_creation() {
        let sandbox = WasmSandbox::new();
        assert!(sandbox.is_ok());
    }

    #[test]
    fn test_register_and_validate() {
        let sandbox = WasmSandbox::new().unwrap();
        let cell = WasmCell::new("test".to_string(), wat_minimal());
        assert!(sandbox.register_cell(&cell).is_ok());
        sandbox.unregister_cell(cell.id);
    }

    #[test]
    fn test_reject_no_entry_point() {
        let sandbox = WasmSandbox::new().unwrap();
        let bad_wat = br#"
            (module
                (memory (export "memory") 1)
            )
        "#;
        let cell = WasmCell::new("bad".to_string(), bad_wat.to_vec());
        assert!(sandbox.register_cell(&cell).is_err());
    }

    #[test]
    fn test_reject_handler_without_memory() {
        let sandbox = WasmSandbox::new().unwrap();
        let bad_wat = br#"
            (module
                (func (export "handler") (param i32 i32) (result i32)
                    i32.const 0)
            )
        "#;
        let cell = WasmCell::new("no-mem".to_string(), bad_wat.to_vec());
        assert!(sandbox.register_cell(&cell).is_err());
    }

    #[tokio::test]
    async fn test_execute_minimal() {
        let sandbox = WasmSandbox::new().unwrap();
        let cell = WasmCell::new("minimal".to_string(), wat_minimal());
        sandbox.register_cell(&cell).unwrap();

        let input = HashMap::new();
        let result = sandbox.execute(&cell, input).await;
        assert!(result.is_ok(), "execute failed: {:?}", result.err());
    }

    #[tokio::test]
    async fn test_execute_run_export() {
        let sandbox = WasmSandbox::new().unwrap();
        let cell = WasmCell::new("run-test".to_string(), wat_run_export());
        sandbox.register_cell(&cell).unwrap();

        let input = HashMap::new();
        let result = sandbox.execute(&cell, input).await;
        assert!(result.is_ok(), "execute failed: {:?}", result.err());
    }

    #[tokio::test]
    async fn test_execute_handler_echo() {
        let sandbox = WasmSandbox::new().unwrap();
        let cell = WasmCell::new("echo".to_string(), wat_handler_echo());
        sandbox.register_cell(&cell).unwrap();

        let mut input = HashMap::new();
        input.insert("key".to_string(), serde_json::json!("value"));
        let result = sandbox.execute_with_metrics(&cell, input).await;
        assert!(result.is_ok(), "execute_with_metrics failed: {:?}", result.err());

        let exec = result.unwrap();
        assert!(exec.success, "execution not successful: {:?}", exec.error);
        assert!(exec.exec_time_ms < 5000, "execution took too long: {}ms", exec.exec_time_ms);
    }

    #[tokio::test]
    async fn test_memory_limit_enforced() {
        let sandbox = WasmSandbox::new().unwrap();
        let alloc_wat = br#"
            (module
                (memory (export "memory") 1 16)
                (func (export "_start")
                    (block $ok
                        (br_if $ok (i32.ge_s (memory.grow (i32.const 16)) (i32.const 0)))
                        (unreachable)
                    )
                )
            )
        "#;
        let cell = WasmCell::new("oom".to_string(), alloc_wat.to_vec())
            .with_memory_limit(1);
        sandbox.register_cell(&cell).unwrap();

        let result = sandbox.execute_with_metrics(&cell, HashMap::new()).await;
        assert!(result.is_ok());
        let exec = result.unwrap();
        assert!(!exec.success, "expected memory limit to cause failure");
        assert!(exec.error.is_some());
    }

    #[tokio::test]
    async fn test_fuel_limit_enforced() {
        let sandbox = WasmSandbox::new().unwrap();
        let infinite_wat = br#"
            (module
                (memory (export "memory") 1)
                (func (export "_start")
                    (loop $inf
                        br $inf)
                )
            )
        "#;
        let cell = WasmCell::new("infinite".to_string(), infinite_wat.to_vec())
            .with_fuel_limit(1000)
            .with_compute_limit(1000);
        sandbox.register_cell(&cell).unwrap();

        let result = sandbox.execute_with_metrics(&cell, HashMap::new()).await;
        assert!(result.is_ok());
        let exec = result.unwrap();
        assert!(!exec.success, "expected fuel limit to cause failure");
    }

    #[tokio::test]
    async fn test_unregister_prevents_execution() {
        let sandbox = WasmSandbox::new().unwrap();
        let cell = WasmCell::new("temp".to_string(), wat_minimal());
        sandbox.register_cell(&cell).unwrap();
        sandbox.unregister_cell(cell.id);

        let result = sandbox.execute(&cell, HashMap::new()).await;
        assert!(result.is_err());
    }

    #[test]
    fn test_cell_pool() {
        let pool = CellPool::new(2);
        let cell1 = WasmCell::new("a".to_string(), vec![]);
        let cell2 = WasmCell::new("b".to_string(), vec![]);
        let cell3 = WasmCell::new("c".to_string(), vec![]);

        assert!(pool.register(cell1).is_ok());
        assert!(pool.register(cell2).is_ok());
        assert!(pool.register(cell3).is_err());
        assert_eq!(pool.len(), 2);
    }

    #[test]
    fn test_cell_builder() {
        let cell = WasmCell::new("test".to_string(), vec![0])
            .with_memory_limit(128)
            .with_compute_limit(10000)
            .with_fuel_limit(5_000_000)
            .with_entry_point("custom_run".to_string());

        assert_eq!(cell.memory_limit_mb, 128);
        assert_eq!(cell.compute_limit_ms, 10000);
        assert_eq!(cell.fuel_limit, 5_000_000);
        assert_eq!(cell.entry_point.as_deref(), Some("custom_run"));
    }
}
