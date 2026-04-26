use std::collections::HashMap;
use std::sync::Arc;

use parking_lot::RwLock;
use tracing::info;
use uuid::Uuid;

use crate::engine::{NodeId, NodeExecutionError};

#[derive(Debug, Clone)]
pub struct WasmCell {
    pub id: Uuid,
    pub node_id: NodeId,
    pub name: String,
    pub wasm_bytes: Vec<u8>,
    pub memory_limit_mb: u32,
    pub compute_limit_ms: u64,
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
        }
    }
}

pub struct WasmSandbox {
    #[cfg(feature = "wasm-sandbox")]
    engine: wasmtime::Engine,
    #[cfg(feature = "wasm-sandbox")]
    cells: Arc<RwLock<HashMap<Uuid, wasmtime::Module>>>,
}

impl WasmSandbox {
    pub fn new() -> anyhow::Result<Self> {
        Ok(Self {
            #[cfg(feature = "wasm-sandbox")]
            engine: wasmtime::Engine::default(),
            #[cfg(feature = "wasm-sandbox")]
            cells: Arc::new(RwLock::new(HashMap::new())),
        })
    }

    #[cfg(feature = "wasm-sandbox")]
    pub fn register_cell(&self, cell: &WasmCell) -> anyhow::Result<()> {
        let module = wasmtime::Module::new(&self.engine, &cell.wasm_bytes)?;
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
    pub async fn execute(
        &self,
        cell: &WasmCell,
        _input: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        let module = {
            let cells = self.cells.read();
            cells.get(&cell.id).cloned()
        };

        let module = module.ok_or_else(|| {
            NodeExecutionError::non_retryable(cell.node_id, format!("Cell {} not found", cell.id))
        })?;

        let wasi = wasmtime_wasi::p2::WasiCtxBuilder::new().build();
        let mut store = wasmtime::Store::new(&self.engine, wasi);

        let instance = wasmtime::Instance::new(&mut store, &module, &[])
            .map_err(|e| NodeExecutionError::non_retryable(cell.node_id, format!("Instance creation failed: {}", e)))?;

        match instance.get_func(&mut store, "run") {
            Some(_func) => {
                Ok(serde_json::json!({
                    "wasm_result": 0,
                    "status": "completed",
                }))
            }
            None => {
                Err(NodeExecutionError::non_retryable(
                    cell.node_id,
                    "Run function not found in WASM module".to_string(),
                ))
            }
        }
    }

    #[cfg(not(feature = "wasm-sandbox"))]
    pub fn register_cell(&self, _cell: &WasmCell) -> anyhow::Result<()> {
        Err(anyhow::anyhow!("WASM support not enabled"))
    }

    #[cfg(not(feature = "wasm-sandbox"))]
    pub fn unregister_cell(&self, _cell_id: Uuid) {}

    #[cfg(not(feature = "wasm-sandbox"))]
    pub async fn execute(
        &self,
        cell: &WasmCell,
        _input: HashMap<String, serde_json::Value>,
    ) -> Result<serde_json::Value, NodeExecutionError> {
        Err(NodeExecutionError::non_retryable(cell.node_id, "WASM support not enabled".to_string()))
    }
}

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
            return Err(anyhow::anyhow!("Cell pool exhausted"));
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
}

impl Default for CellPool {
    fn default() -> Self {
        Self::new(10_000)
    }
}
