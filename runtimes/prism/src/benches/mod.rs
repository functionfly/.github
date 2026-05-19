//! Benchmarking utilities for Prism Runtime
//!
//! This module provides pure benchmark functions (without criterion)
//! that can be used by the criterion harness in benches/benchmark.rs

pub mod cell_bench;
pub mod wasm_bench;
pub mod scheduler_bench;
pub mod crdt_bench;
pub mod neural_bench;