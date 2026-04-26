---
title: Rust SDK
description: Rust SDK for building FunctionFly functions
---

# Rust SDK

The FunctionFly Rust SDK provides a trait-based API for building FunctionFly functions in Rust, compiled to WebAssembly.

## Installation

Add to your `Cargo.toml`:

```toml
[dependencies]
functionfly-sdk = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
```

## Quick Start

```rust
use functionfly_sdk::{Function, Context};

struct MyFunction;

impl Function for MyFunction {
    fn handle(&self, input: &str, _ctx: &Context) -> Result<String, String> {
        Ok(r#"{"message": "Hello from Rust!"}"#.to_string())
    }
}

functionfly_sdk::run!(MyFunction);
```

## Function Trait

All functions implement the `Function` trait:

```rust
pub trait Function {
    /// Process the input and return the output.
    fn handle(&self, input: &str, ctx: &Context) -> Result<String, String>;
}
```

## The `run!` Macro

The `run!` macro registers the required WASM entry points:

```rust
functionfly_sdk::run!(MyFunction);
```

This generates:
- `init()` — Called once at cold start
- `execute(ptr, len)` — Main execution entry point
- `alloc(size)` — Memory allocator
- `dealloc(ptr, size)` — Memory deallocator

## Host Functions

Access FunctionFly platform services:

```rust
use functionfly_sdk::{log, get_env, kv_get, kv_set, fetch};

// Logging
log("Processing request");

// Environment variables
let api_key = get_env("API_KEY").unwrap_or_default();

// Key-value store
kv_set("counter", "42");
let value = kv_get("counter").unwrap_or_default();

// HTTP fetch
let response = fetch("https://api.example.com/data").unwrap_or_default;
```

## Context

The `Context` struct provides execution metadata:

```rust
use functionfly_sdk::Context;

let ctx = Context::new();
// ctx.function_name — Name of the function
// ctx.version — Deployment version
```

## Example Functions

### HTTP API Handler

```rust
use functionfly_sdk::{Function, Context};
use serde::{Deserialize, Serialize};
use serde_json::json;

#[derive(Deserialize)]
struct Request {
    method: String,
    body: Option<serde_json::Value>,
}

#[derive(Serialize)]
struct User {
    id: String,
    name: String,
    email: String,
}

struct UserAPI;

impl Function for UserAPI {
    fn handle(&self, input: &str, _ctx: &Context) -> Result<String, String> {
        let request: Request = serde_json::from_str(input)
            .map_err(|e| format!("Invalid JSON: {}", e))?;

        match request.method.as_str() {
            "GET" => {
                let users = vec![
                    User { id: "1".into(), name: "Alice".into(), email: "alice@example.com".into() },
                    User { id: "2".into(), name: "Bob".into(), email: "bob@example.com".into() },
                ];
                Ok(json!({ "status": 200, "body": { "users": users } }).to_string())
            }
            "POST" => {
                let body = request.body.unwrap_or_default();
                let user = User {
                    id: "3".into(),
                    name: body["name"].as_str().unwrap_or("").into(),
                    email: body["email"].as_str().unwrap_or("").into(),
                };
                Ok(json!({ "status": 201, "body": user }).to_string())
            }
            _ => Ok(r#"{"status": 405, "body": {"error": "Method not allowed"}}"#.into()),
        }
    }
}

functionfly_sdk::run!(UserAPI);
```

### Data Transformation

```rust
use functionfly_sdk::{Function, Context};
use serde::Deserialize;

#[derive(Deserialize)]
struct Input {
    body: InputBody,
}

#[derive(Deserialize)]
struct InputBody {
    id: String,
    first_name: String,
    last_name: String,
    email: String,
    created_at: String,
}

struct Transformer;

impl Function for Transformer {
    fn handle(&self, input: &str, _ctx: &Context) -> Result<String, String> {
        let req: Input = serde_json::from_str(input)
            .map_err(|e| format!("Invalid JSON: {}", e))?;

        let output = serde_json::json!({
            "status": 200,
            "body": {
                "id": req.body.id,
                "name": format!("{} {}", req.body.first_name, req.body.last_name).trim(),
                "email": req.body.email.to_lowercase(),
                "timestamp": req.body.created_at
            }
        });

        Ok(output.to_string())
    }
}

functionfly_sdk::run!(Transformer);
```

## Building for WASM

```bash
# Add the WASM target
rustup target add wasm32-wasip1

# Build
cargo build --target wasm32-wasip1 --release

# Optimize (optional)
wasm-opt -O3 target/wasm32-wasip1/release/my-function.wasm -o optimized.wasm
```

## Error Handling

Return `Result<String, String>` from `handle`:

```rust
impl Function for SafeFunction {
    fn handle(&self, input: &str, ctx: &Context) -> Result<String, String> {
        let request: serde_json::Value = serde_json::from_str(input)
            .map_err(|e| format!("Invalid JSON: {}", e))?;

        let result = process(&request)?;
        Ok(serde_json::json!({ "status": 200, "body": result }).to_string())
    }
}
```

## API Reference

### Traits

| Trait | Method | Signature |
|-------|--------|-----------|
| `Function` | `handle` | `fn handle(&self, input: &str, ctx: &Context) -> Result<String, String>` |

### Host Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `log` | `fn log(msg: &str)` | Log a message |
| `get_env` | `fn get_env(key: &str) -> Option<String>` | Get environment variable |
| `kv_get` | `fn kv_get(key: &str) -> Option<String>` | Read from KV store |
| `kv_set` | `fn kv_set(key: &str, value: &str)` | Write to KV store |
| `fetch` | `fn fetch(url: &str) -> Option<String>` | HTTP fetch |

### Macros

| Macro | Usage | Description |
|-------|-------|-------------|
| `run!` | `functionfly_sdk::run!(MyFunction)` | Register WASM entry points |
