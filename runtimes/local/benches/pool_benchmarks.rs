//! Pool-aware execution benchmarks.
//!
//! Run with: `cargo bench --package functionfly-local --bench pool_benchmarks`
//!
//! These benchmarks measure:
//! - Cold start: execution when pool is empty (falls back to engine.execute)
//! - Warm execution: execution when pool has pre-warmed instances
//! - Pre-warm overhead: cost of compiling and pre-warming the pool
//! - AOT cache hit vs miss: compilation time savings

use std::hint::black_box;
use std::sync::Arc;
use wasmtime::Module;

use criterion::{criterion_group, criterion_main};

use functionfly_local::config::Config;
use functionfly_local::engine::{AotCache, WasmEngine};
use functionfly_local::pool::{PoolManager, PooledWasmInstance, WasmInstancePool};
use functionfly_local::wasi::WasiContext;

// =============================================================================
// Benchmark Configuration
// =============================================================================

fn create_benchmark_config() -> Config {
    Config {
        aot_cache_enabled: true,
        aot_cache_size_mb: 128,
        aot_cache_dir: "".to_string(),
        function: "bench".to_string(),
        version: "1.0.0".to_string(),
        wasi_env: vec![],
        max_output_bytes: 1024 * 1024,
        memory_mb: 128,
        timeout_ms: 5000,
        wasm_pool_enabled: true,
        wasm_pool_max_concurrent: 10,
        wasm_pool_max_idle: 4,
        wasm_pool_prewarm_count: 1,
        ..Config::default()
    }
}

/// Minimal WASM module (empty with just a start function)
fn create_minimal_wasm_module() -> Vec<u8> {
    vec![
        0x00, 0x61, 0x73, 0x6d, // magic
        0x01, 0x00, 0x00, 0x00, // version
        0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // type section
        0x03, 0x02, 0x01, 0x00, // func section
        0x07, 0x07, 0x01, 0x03, 0x61, 0x64, 0x64, 0x00, 0x00, // export "add"
        0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b, // code section
    ]
}

/// WASM module with handler function that returns input
fn create_handler_wasm_module() -> Vec<u8> {
    // This module exports a "handler" function that accepts (i32, i32) ptr/len
    wat::parse_str(r#"
        (module
            (memory (export "memory") 1)
            (func (export "handler") (param i32 i32) (result i32)
                ;; Just return the length (input length)
                local.get 1
            )
            (func (export "_start"))
        )
    "#).unwrap()
}

/// Creates a WASM module of approximately target size (for testing large modules)
fn create_wasm_module_with_size(target_bytes: usize) -> Vec<u8> {
    let base = create_minimal_wasm_module();
    let padding_needed = target_bytes.saturating_sub(base.len());
    if padding_needed == 0 {
        return base;
    }

    let mut module = base;
    let custom_name = "p";
    let name_len = 1;
    let section_overhead = 3 + name_len;
    let data_per_section = 65535;

    let mut remaining = padding_needed;
    while remaining > 0 {
        let data_len = std::cmp::min(remaining.saturating_sub(section_overhead), data_per_section);
        if data_len == 0 {
            break;
        }
        module.push(0x00);
        module.push((name_len + data_len + 1) as u8);
        module.push(name_len as u8);
        module.extend_from_slice(custom_name.as_bytes());
        module.extend(vec![0x00; data_len]);
        remaining = remaining.saturating_sub(section_overhead + data_len);
    }
    module
}

// =============================================================================
// AOT Cache Benchmarks
// =============================================================================

fn bench_aot_cache(c: &mut criterion::Criterion) {
    let mut group = c.benchmark_group("aot_cache");

    let config = create_benchmark_config();
    let engine = wasmtime::Engine::default();
    let aot_cache = AotCache::new();
    let wasm_bytes = create_minimal_wasm_module();

    // First call - cold cache
    group.bench_function("compile_cold", |b| {
        // Create fresh engine for each run to avoid engine-level caching
        let fresh_engine = wasmtime::Engine::default();
        b.iter(|| {
            let result = aot_cache.get_or_compile_module(
                black_box(&fresh_engine),
                black_box(&wasm_bytes),
                black_box(&config),
            );
            black_box(result)
        });
    });

    // Subsequent calls - warm cache
    // First compile to populate cache
    let _ = aot_cache.get_or_compile_module(&engine, &wasm_bytes, &config);

    group.bench_function("compile_warm", |b| {
        b.iter(|| {
            let result = aot_cache.get_or_compile_module(
                black_box(&engine),
                black_box(&wasm_bytes),
                black_box(&config),
            );
            black_box(result)
        });
    });

    group.finish();
}

// =============================================================================
// Module Compilation Benchmarks
// =============================================================================

fn bench_module_compilation(c: &mut criterion::Criterion) {
    let mut group = c.benchmark_group("module_compilation");

    let engine = wasmtime::Engine::default();
    let wasm_bytes = create_minimal_wasm_module();

    // Fresh compilation (no caching)
    group.bench_function("fresh_compile", |b| {
        b.iter(|| {
            let result = Module::new(black_box(&engine), black_box(wasm_bytes.as_slice()));
            black_box(result)
        });
    });

    // AOT cached compilation
    let config = create_benchmark_config();
    let aot_cache = AotCache::new();
    let _ = aot_cache.get_or_compile_module(&engine, &wasm_bytes, &config);

    group.bench_function("aot_cached_compile", |b| {
        b.iter(|| {
            let result = aot_cache.get_or_compile_module(
                black_box(&engine),
                black_box(&wasm_bytes),
                black_box(&config),
            );
            black_box(result)
        });
    });

    // Different module sizes
    for size in [1_000, 10_000, 100_000, 1_000_000].iter() {
        let _wasm_bytes = create_wasm_module_with_size(*size);
        group.bench_with_input(criterion::BenchmarkId::from_parameter(size), size, |b, &wasm_size| {
            let wasm_bytes = create_wasm_module_with_size(wasm_size);
            b.iter(|| {
                let result = aot_cache.get_or_compile_module(
                    black_box(&engine),
                    black_box(&wasm_bytes),
                    black_box(&config),
                );
                black_box(result)
            });
        });
    }

    group.finish();
}

// =============================================================================
// Pool Benchmarks
// =============================================================================

fn bench_pool_operations(c: &mut criterion::Criterion) {
    let mut group = c.benchmark_group("pool_operations");

    // Pre-warm overhead - measure instance creation time
    group.bench_function("prewarm_overhead", |b| {
        let engine = wasmtime::Engine::default();
        let wasm_bytes = create_handler_wasm_module();
        b.iter(|| {
            let module = wasmtime::Module::new(&engine, &wasm_bytes).unwrap();
            let config = create_benchmark_config();
            let wasi_ctx = WasiContext::new_with_input(&config, "bench@1.0.0".to_string(), "").unwrap();
            let instance = PooledWasmInstance::new(
                module,
                wasi_ctx.ctx,
                "bench@1.0.0".to_string(),
                1024 * 1024,
            );
            black_box(instance)
        });
    });

    // Pool creation overhead
    group.bench_function("pool_creation", |b| {
        b.iter(|| {
            let pool = WasmInstancePool::new("bench@1.0.0".to_string(), 10, 4);
            black_box(pool)
        });
    });

    // Pool stats collection (non-blocking)
    let pool = Arc::new(WasmInstancePool::new("bench@1.0.0".to_string(), 10, 4));
    group.bench_function("pool_stats_blocking", |b| {
        b.iter(|| {
            let stats = pool.blocking_stats();
            black_box(stats)
        });
    });

    // Pool is_warmed check
    group.bench_function("pool_is_warmed_blocking", |b| {
        b.iter(|| {
            let warmed = pool.is_warmed_blocking();
            black_box(warmed)
        });
    });

    group.finish();
}

// =============================================================================
// End-to-End Execution Benchmarks
// =============================================================================

fn bench_execution(c: &mut criterion::Criterion) {
    let mut group = c.benchmark_group("execution");

    let rt = tokio::runtime::Runtime::new().unwrap();
    let logger = functionfly_local::logging::init_structured_logging(false);
    let security_monitor = Arc::new(functionfly_local::security::SecurityMonitor::new());
    let config = create_benchmark_config();

    let engine = rt.block_on(async {
        WasmEngine::with_config(
            config.clone(),
            None,
            logger,
            None,
            security_monitor,
            None,
        ).expect("Failed to create engine")
    });

    let wasm_bytes = create_handler_wasm_module();
    let pool_manager = Arc::new(PoolManager::new(10, 4));
    engine.set_pool_manager(pool_manager.clone());

    // Cold execution (pool empty) - measures full compilation + execution
    group.bench_function("cold_execution", |b| {
        // Clear the pool before each run
        let pool = Arc::new(PoolManager::new(10, 4));
        engine.set_pool_manager(pool.clone());
        let wasm_bytes = wasm_bytes.clone();
        let config = config.clone();

        b.iter(|| {
            let result = rt.block_on(engine.execute(
                black_box(&wasm_bytes),
                black_box("test input"),
                black_box(&config),
                black_box(None),
                black_box(None),
            ));
            black_box(result)
        });
    });

    // Module compilation benchmark (proxy for cold execution cost)
    group.bench_function("module_compile", |b| {
        let engine = wasmtime::Engine::default();
        b.iter(|| {
            let result = Module::new(&engine, black_box(&wasm_bytes));
            black_box(result)
        });
    });

    group.finish();
}

// =============================================================================
// WASI Context Creation Benchmarks
// =============================================================================

fn bench_wasi_context(c: &mut criterion::Criterion) {
    let mut group = c.benchmark_group("wasi_context");

    let config = create_benchmark_config();
    let function_key = "bench@1.0.0".to_string();

    group.bench_function("create_context_empty_input", |b| {
        b.iter(|| {
            let result = WasiContext::new_with_input(
                black_box(&config),
                function_key.clone(),
                "",
            );
            black_box(result)
        });
    });

    group.bench_function("create_context_with_input", |b| {
        let input = "x".repeat(1000);
        b.iter(|| {
            let result = WasiContext::new_with_input(
                black_box(&config),
                function_key.clone(),
                &input,
            );
            black_box(result)
        });
    });

    group.finish();
}

// =============================================================================
// Run All Benchmarks
// =============================================================================

criterion_group!(
    name = benches;
    config = criterion::Criterion::default();
    targets = bench_aot_cache, bench_module_compilation, bench_pool_operations, bench_execution, bench_wasi_context
);
criterion_main!(benches);