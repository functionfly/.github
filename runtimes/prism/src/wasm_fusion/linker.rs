//! Fusion linker for connecting WASM modules

use std::collections::HashMap;

use wasmtime::{Linker, Module, Store, Instance, Engine};

use crate::core::{PrismError, PrismResult};

/// Linker for connecting WASM modules in a fusion graph
pub struct FusionLinker {
    engine: Engine,
    linker: Linker<()>,
    imports: HashMap<String, wasmtime::Func>,
}

impl FusionLinker {
    pub fn new(engine: &Engine) -> Self {
        Self {
            engine: engine.clone(),
            linker: Linker::new(engine),
            imports: HashMap::new(),
        }
    }

    /// Define a host function that can be imported by WASM modules
    pub fn define_host_function(
        &mut self,
        module: &str,
        name: &str,
        func: wasmtime::Func,
    ) -> PrismResult<()> {
        let mut store = Store::new(&self.engine, ());
        self.linker
            .define(&mut store, module, name, func)
            .map_err(|e| PrismError::WasmModuleError(e.to_string()))?;
        self.imports.insert(format!("{}::{}", module, name), func);
        Ok(())
    }

    /// Link a module using the defined imports
    pub fn link_module(&self, module: &Module) -> PrismResult<Instance> {
        let mut store = Store::new(&self.engine, ());
        self.linker
            .instantiate(&mut store, module)
            .map_err(|e| PrismError::WasmExecutionFailed(e.to_string()))
    }

    /// Get an exported function from a linked instance
    pub fn get_func(&self, instance: &Instance, name: &str) -> PrismResult<wasmtime::Func> {
        let mut store = Store::new(&self.engine, ());
        instance
            .get_func(&mut store, name)
            .ok_or_else(|| PrismError::WasmExecutionFailed(format!("Function not found: {}", name)))
    }
}

impl Default for FusionLinker {
    fn default() -> Self {
        // The Default impl cannot construct a valid FusionLinker because the
        // underlying wasmtime::Engine must be configured. Build a fresh
        // Engine with default settings so the linker is still usable for
        // test/utility code that just needs *some* linker.
        Self::new(&wasmtime::Engine::default())
    }
}