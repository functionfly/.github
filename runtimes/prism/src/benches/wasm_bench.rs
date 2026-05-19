//! WASM execution benchmarks
//!
//! Tests performance of WASM module compilation, registration, and execution.

use criterion::{black_box, criterion_group, Criterion};
use std::sync::Arc;
use crate::wasm_fusion::{FusionEngine, FusionEngineConfig};
use crate::wasm_fusion::{FusionGraph, FusionNode, FusionNodeType, NodeConfig};

/// Minimal valid WASM module that exports _start
/// Generated from: (module (func (export "_start")))
fn minimal_wasm() -> Vec<u8> {
    vec![
        0x00, 0x61, 0x73, 0x6d, // magic "\0asm"
        0x01, 0x00, 0x00, 0x00, // version 1
        0x01, 0x04, 0x01, 0x60, 0x00, 0x00, // type section: 1 type, func () -> ()
        0x03, 0x02, 0x01, 0x00, // function section: 1 function (type 0)
        0x07, 0x0a, 0x01, 0x06, 0x5f, 0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x00, // export section: _start
        0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b, // code section: 1 body, end
    ]
}

/// WASM module with memory export (required for handler export)
fn wasm_with_handler() -> Vec<u8> {
    vec![
        0x00, 0x61, 0x73, 0x6d, // magic
        0x01, 0x00, 0x00, 0x00, // version
        // Type section
        0x01, 0x07, 0x02, 0x60, 0x00, 0x00, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, // func () -> (), func (i32, i32) -> i32
        // Function section
        0x03, 0x02, 0x01, 0x00, 0x01, // 2 functions: type 0, type 1
        // Memory section
        0x05, 0x03, 0x01, 0x00, 0x00, // memory: 1 page, no max
        // Export section
        0x07, 0x13, 0x03, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x00, 0x02, // "memory" export (memory)
        0x07, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x00, 0x01, // "handler" export (func 1)
        0x0a, 0x09, 0x01, 0x07, 0x00, 0x20, 0x00, 0x20, 0x01, 0x6a, 0x0b, 0x0b, // code: handler gets (i32, i32) and returns i32
    ]
}

/// WASM module that performs computation
fn wasm_computation_module() -> Vec<u8> {
    vec![
        0x00, 0x61, 0x73, 0x6d,
        0x01, 0x00, 0x00, 0x00,
        // Type section - two functions
        0x01, 0x0b, 0x03, 0x60, 0x00, 0x00, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f,
        // Function section
        0x03, 0x02, 0x01, 0x00, 0x01,
        // Memory section
        0x05, 0x03, 0x01, 0x00, 0x00,
        // Export section
        0x07, 0x13, 0x03,
        0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x00, 0x02, // memory
        0x07, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72, 0x00, 0x01, // handler
        // Code section - handler: (i32, i32) -> i32
        // Get memory[input_ptr..input_ptr+input_len], sum bytes, write to output
        0x0a, 0x09, 0x01, 0x07, 0x00,
        0x20, 0x00, // local.get 0 (ptr)
        0x20, 0x01, // local.get 1 (len)
        0x20, 0x00, // local.get 0 (ptr)
        0x20, 0x01, // local.get 1 (len)
        0x02, 0x40, // block for loop
        0x03, 0x40, // loop
        0x20, 0x02, // local.get 2 (i)
        0x20, 0x03, // local.get 3 (sum)
        0x20, 0x02, // local.get 2 (i)
        0x2f, 0x00, // i32.load8_u (load byte)
        0x6c, // i32.or
        0x21, 0x03, // local.tee sum
        0x20, 0x02, // local.get 2 (i)
        0x20, 0x01, // local.get 1 (len)
        0x46, // i32.eqz - exit if i == len
        0x04, 0x40, // br_if 1
        0x20, 0x02, // local.get 2 (i)
        0x20, 0x02, // local.get 2 (i)
        0x20, 0x03, // local.get 3 (sum)
        0x0b, 0x0b, // end block / loop
        0x0b, // end
    ]
}

pub fn bench_fusion_engine_creation(c: &mut Criterion) {
    c.bench_function("fusion_engine_creation", |b| {
        b.iter(|| {
            let config = FusionEngineConfig::default();
            black_box(FusionEngine::new(config).unwrap())
        });
    });
}

pub fn bench_fusion_engine_with_custom_config(c: &mut Criterion) {
    c.bench_function("fusion_engine_custom_config", |b| {
        b.iter(|| {
            let config = FusionEngineConfig {
                enable_streaming: true,
                enable_module_merging: true,
                max_modules: 64,
                memory_limit_mb: 1024,
                compute_timeout_ms: 60_000,
                fuel_limit: 10_000_000,
            };
            black_box(FusionEngine::new(config).unwrap())
        });
    });
}

pub async fn async_bench_module_registration() -> impl Fn(&mut Criterion) {
    move |c: &mut Criterion| {
        let wasm = minimal_wasm();
        let wasm_clone = wasm.clone();

        c.bench_function("module_registration", |b| {
            let runtime = tokio::runtime::Runtime::new().unwrap();
            b.to_async(&runtime).iter(|| async {
                let engine = FusionEngine::new(FusionEngineConfig::default()).unwrap();
                engine.register_module("test-module", &wasm_clone).await
            });
        });
    }
}

pub fn bench_wasm_module_validation(c: &mut Criterion) {
    let wasm = minimal_wasm();

    c.bench_function("wasm_module_validation", |b| {
        b.iter(|| {
            let config = FusionEngineConfig::default();
            let engine = FusionEngine::new(config).unwrap();
            let config = wasmtime::Config::new();
            let engine_obj = wasmtime::Engine::new(&config).unwrap();
            let _module = wasmtime::Module::new(&engine_obj, &wasm);
        });
    });
}

pub fn bench_wasm_bytes_creation(c: &mut Criterion) {
    c.bench_function("wasm_bytes_creation", |b| {
        b.iter(|| {
            black_box(minimal_wasm())
        });
    });

    c.bench_function("wasm_with_handler_creation", |b| {
        b.iter(|| {
            black_box(wasm_with_handler())
        });
    });
}

pub fn bench_fusion_graph_creation(c: &mut Criterion) {
    c.bench_function("fusion_graph_creation_empty", |b| {
        b.iter(|| {
            black_box(FusionGraph::new("bench-graph".to_string()))
        });
    });
}

pub fn bench_fusion_node_creation(c: &mut Criterion) {
    c.bench_function("fusion_node_creation", |b| {
        b.iter(|| {
            let node = FusionNode::new(
                "bench-node".to_string(),
                FusionNodeType::Wasm,
                NodeConfig::default(),
            );
            black_box(node)
        });
    });
}

pub fn bench_fusion_graph_add_node(c: &mut Criterion) {
    c.bench_function("fusion_graph_add_node", |b| {
        b.iter(|| {
            let mut graph = FusionGraph::new("bench-graph".to_string());
            for i in 0..10 {
                let node = FusionNode::new(
                    format!("node-{}", i),
                    FusionNodeType::Wasm,
                    NodeConfig::default(),
                );
                graph.add_node(node);
            }
            black_box(graph)
        });
    });
}

pub fn bench_node_config_defaults(c: &mut Criterion) {
    c.bench_function("node_config_default", |b| {
        b.iter(|| {
            black_box(NodeConfig::default())
        });
    });
}

pub fn bench_fusion_node_type_variants(c: &mut Criterion) {
    c.bench_function("fusion_node_type_variants", |b| {
        b.iter(|| {
            black_box(FusionNodeType::Wasm);
            black_box(FusionNodeType::FunctionChain);
            black_box(FusionNodeType::StreamMap);
            black_box(FusionNodeType::StreamFilter);
            black_box(FusionNodeType::StreamReduce);
        });
    });
}

pub fn bench_stream_operations(c: &mut Criterion) {
    c.bench_function("stream_map_creation", |b| {
        b.iter(|| {
            let node = FusionNode::new(
                "stream-map".to_string(),
                FusionNodeType::StreamMap,
                NodeConfig {
                    imports: vec!["double".to_string()],
                    ..Default::default()
                },
            );
            black_box(node)
        });
    });

    c.bench_function("stream_filter_creation", |b| {
        b.iter(|| {
            let node = FusionNode::new(
                "stream-filter".to_string(),
                FusionNodeType::StreamFilter,
                NodeConfig {
                    imports: vec!["is_number".to_string()],
                    ..Default::default()
                },
            );
            black_box(node)
        });
    });

    c.bench_function("stream_reduce_creation", |b| {
        b.iter(|| {
            let node = FusionNode::new(
                "stream-reduce".to_string(),
                FusionNodeType::StreamReduce,
                NodeConfig {
                    imports: vec!["sum".to_string()],
                    ..Default::default()
                },
            );
            black_box(node)
        });
    });
}

pub fn bench_function_chain_creation(c: &mut Criterion) {
    c.bench_function("function_chain_creation", |b| {
        b.iter(|| {
            let node = FusionNode::new(
                "func-chain".to_string(),
                FusionNodeType::FunctionChain,
                NodeConfig {
                    imports: vec![
                        "base64_encode".to_string(),
                        "hash_sha256".to_string(),
                        "compress".to_string(),
                    ],
                    ..Default::default()
                },
            );
            black_box(node)
        });
    });
}

pub fn wasm_benchmarks() -> impl Fn(&mut Criterion) {
    move |c: &mut Criterion| {
        bench_fusion_engine_creation(c);
        bench_fusion_engine_with_custom_config(c);
        bench_wasm_module_validation(c);
        bench_wasm_bytes_creation(c);
        bench_fusion_graph_creation(c);
        bench_fusion_node_creation(c);
        bench_fusion_graph_add_node(c);
        bench_node_config_defaults(c);
        bench_fusion_node_type_variants(c);
        bench_stream_operations(c);
        bench_function_chain_creation(c);
    }
}