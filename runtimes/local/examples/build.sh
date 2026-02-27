#!/bin/bash

# Build the WASI example for FunctionFly
# Requires: cargo, wasm32-wasi target

set -e

echo "Building WASI example for FunctionFly..."

# Add the WASI target if not present
rustup target add wasm32-wasi

# Build the example
cargo build --target wasm32-wasi --release --example wasi-example

# Copy to a convenient location
cp target/wasm32-wasi/release/examples/wasi_example.wasm .

echo "Built wasi_example.wasm"
echo ""
echo "Run with FunctionFly:"
echo "functionfly-local --wasm wasi_example.wasm --wasi-enabled --wasi-dirs /tmp:/tmp:rw"