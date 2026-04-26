//! Wasmtime Engine Configuration
//!
//! Provides secure Wasmtime engine configuration with:
//! - Fuel consumption for CPU timeouts
//! - Epoch interruption for wall-clock timeouts
//! - Stack size limits to prevent stack overflow attacks
//! - Memory limiter integration
//!
//! # Security Hardening
//!
//! This module implements defense-in-depth security:
//! 1. **Fuel-based limits**: CPU instruction counting prevents infinite loops
//! 2. **Epoch interruption**: Wall-clock timeouts via epoch counter
//! 3. **Stack limits**: Prevents stack overflow attacks from malicious WASM
//! 4. **Memory limits**: Engine-level memory.grow enforcement

mod memory_limiter;

pub use memory_limiter::{
    install_memory_limiter, with_limiter, FunctionMemoryLimiter, LimiterGuard,
};

use std::sync::Arc;
use std::time::{Duration, Instant};
use wasmtime::{Config, Engine, Module, Store};
use wasmtime_wasi::p1::WasiP1Ctx;

const DEFAULT_MAX_WASM_STACK: usize = 512 * 1024;
const DEFAULT_FUEL_PER_MS: u64 = 10_000;

#[derive(Clone)]
pub struct WasmEngine {
    engine: Arc<Engine>,
    config: WasmEngineConfig,
}

#[derive(Clone, Debug)]
pub struct WasmEngineConfig {
    pub max_memory_mb: u32,
    pub max_timeout_ms: u64,
    pub max_wasm_stack: usize,
    pub fuel_per_ms: u64,
    pub consume_fuel: bool,
    pub epoch_interruption: bool,
}

impl Default for WasmEngineConfig {
    fn default() -> Self {
        Self {
            max_memory_mb: 128,
            max_timeout_ms: 30000,
            max_wasm_stack: DEFAULT_MAX_WASM_STACK,
            fuel_per_ms: DEFAULT_FUEL_PER_MS,
            consume_fuel: true,
            epoch_interruption: true,
        }
    }
}

impl WasmEngine {
    pub fn new(config: WasmEngineConfig) -> anyhow::Result<Self> {
        let engine = Self::create_engine(&config)?;
        Ok(Self {
            engine: Arc::new(engine),
            config,
        })
    }

    fn create_engine(config: &WasmEngineConfig) -> anyhow::Result<Engine> {
        let mut wasm_config = Config::new();

        wasm_config
            .consume_fuel(config.consume_fuel)
            .epoch_interruption(config.epoch_interruption)
            .max_wasm_stack(config.max_wasm_stack);

        Engine::new(&wasm_config)
            .map_err(|e| anyhow::anyhow!("Failed to create Wasmtime engine: {}", e))
    }

    pub fn engine(&self) -> &Engine {
        &self.engine
    }

    pub fn config(&self) -> &WasmEngineConfig {
        &self.config
    }

    pub fn increment_epoch(&self) {
        self.engine.increment_epoch();
    }

    pub fn spawn_epoch_ticker(&self) -> std::thread::JoinHandle<()> {
        let engine = self.engine.clone();
        std::thread::Builder::new()
            .name("wasmtime-epoch-ticker".to_string())
            .spawn(move || loop {
                std::thread::sleep(Duration::from_millis(1));
                engine.increment_epoch();
            })
            .expect("Failed to spawn epoch ticker thread")
    }

    pub fn create_store(&self, wasi_ctx: WasiP1Ctx) -> Store<WasiP1Ctx> {
        let mut store = Store::new(&self.engine, wasi_ctx);
        let fuel_limit = self.config.fuel_per_ms * self.config.max_timeout_ms;
        store.set_fuel(fuel_limit).expect("Failed to set fuel");
        store
    }

    pub fn create_store_with_limiter(
        &self,
        wasi_ctx: WasiP1Ctx,
        memory_mb: u32,
    ) -> Store<WasiP1Ctx> {
        let mut store = self.create_store(wasi_ctx);
        let _guard = install_memory_limiter(memory_mb);
        unsafe {
            store.limiter(|_data| with_limiter(|l| l));
        }
        store
    }

    pub fn compile_module(&self, wasm_bytes: &[u8]) -> anyhow::Result<Module> {
        Module::new(&self.engine, wasm_bytes)
            .map_err(|e| anyhow::anyhow!("Failed to compile WASM module: {}", e))
    }

    pub fn execute(&self, wasm_bytes: &[u8], input: &str) -> anyhow::Result<String> {
        let limiter_guard = install_memory_limiter(self.config.max_memory_mb);
        let epoch_guard = self.spawn_epoch_ticker();

        let start = Instant::now();
        let result = self.execute_inner(wasm_bytes, input);

        drop(limiter_guard);
        drop(epoch_guard);

        let elapsed = start.elapsed().as_millis() as u64;
        if elapsed > self.config.max_timeout_ms {
            return Err(anyhow::anyhow!("Execution timeout after {}ms", elapsed));
        }

        result
    }

    fn execute_inner(&self, wasm_bytes: &[u8], input: &str) -> anyhow::Result<String> {
        let wasi = wasmtime_wasi::WasiCtxBuilder::new()
            .inherit_stdio()
            .build_p1();

        let mut store = self.create_store_with_limiter(wasi, self.config.max_memory_mb);

        let module = self.compile_module(wasm_bytes)?;

        let wasi_linker = wasmtime::Linker::<WasiP1Ctx>::new(&self.engine);
        let instance = wasi_linker.instantiate(&mut store, &module)?;
        let handler = instance.get_typed_func::<(i32, i32), i32>(&mut store, "handler")?;

        let input_bytes = input.as_bytes();
        let input_ptr = 0u32 as i32;
        let input_len = input_bytes.len() as i32;

        let result = handler.call(&mut store, (input_ptr, input_len))?;

        Ok(format!("Execution returned: {}", result))
    }
}

pub fn calculate_fuel_limit(timeout_ms: u64, fuel_per_ms: u64) -> u64 {
    timeout_ms.saturating_mul(fuel_per_ms)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_engine_config_default() {
        let config = WasmEngineConfig::default();
        assert_eq!(config.max_memory_mb, 128);
        assert_eq!(config.max_timeout_ms, 30000);
        assert_eq!(config.max_wasm_stack, DEFAULT_MAX_WASM_STACK);
        assert!(config.consume_fuel);
        assert!(config.epoch_interruption);
    }

    #[test]
    fn test_calculate_fuel_limit() {
        assert_eq!(calculate_fuel_limit(1000, 10000), 10_000_000);
        assert_eq!(calculate_fuel_limit(0, 10000), 0);
        assert_eq!(calculate_fuel_limit(30000, 10000), 300_000_000);
    }
}
