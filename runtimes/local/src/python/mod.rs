//! Python WASM runtime support.
//!
//! This module provides Python function execution using pre-compiled
//! Micropython WASM modules. Python functions compile to WASM artifacts
//! that execute in the same Wasmtime runtime as Rust/Go functions.

pub mod engine;
pub mod runtime;
