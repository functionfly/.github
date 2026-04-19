---
title: Rust via WASM Runtime
description: Rust WebAssembly runtime environment for FunctionFly functions
---

# Rust via WASM Runtime

FunctionFly's Rust runtime compiles your Rust code to WebAssembly (WASM) for efficient, sandboxed execution.

## Overview

Rust functions are compiled to WebAssembly and run in a secure, lightweight WASM runtime. This provides:

- **Fast cold starts** - WASM modules initialize quickly
- **Small footprint** - Compact binary sizes
- **Sandboxed** - Memory-safe, isolated execution
- **Portable** - Same binary runs on any platform

## Supported Toolchains

| Toolchain | Status | Notes |
|-----------|--------|-------|
| wasm32-wasi | Supported | Recommended |
| wasm32-unknown-unknown | Supported | Limited std support |

## Project Structure

```
my-function/
├── Cargo.toml        # Rust manifest
├── src/
│   └── main.rs       # Entry point
└── functionfly.jsonc # Function config
```

## Cargo.toml

```toml
[package]
name = "my-function"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

[profile.release]
opt-level = 3
lto = true
strip = true
```

## Function Structure

A Rust function uses the WASI interface:

```rust
use serde::{Deserialize, Serialize};
use std::io::{self, Read, Write};

#[derive(Deserialize)]
struct Request {
    body: serde_json::Value,
    headers: std::collections::HashMap<String, String>,
    params: std::collections::HashMap<String, String>,
    method: String,
    url: String,
    path: String,
}

#[derive(Serialize)]
struct Response {
    status: u16,
    body: serde_json::Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    headers: Option<std::collections::HashMap<String, String>>,
}

fn main() {
    // Read request from stdin
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    
    let request: Request = serde_json::from_str(&input).unwrap();
    
    // Process and respond
    let response = handler(request);
    
    // Write response to stdout
    let output = serde_json::to_string(&response).unwrap();
    io::stdout().write_all(output.as_bytes()).unwrap();
}

fn handler(request: Request) -> Response {
    Response {
        status: 200,
        body: serde_json::json!({"message": "Hello, World!"}),
        headers: Some({
            let mut h = std::collections::HashMap::new();
            h.insert("Content-Type".to_string(), "application/json".to_string());
            h
        }),
    }
}
```

## Example Functions

### HTTP API Handler

```rust
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::collections::HashMap;
use std::io::{self, Read, Write};

#[derive(Deserialize)]
struct Request {
    body: serde_json::Value,
    headers: HashMap<String, String>,
    method: String,
}

#[derive(Serialize)]
struct Response {
    status: u16,
    body: serde_json::Value,
    headers: Option<HashMap<String, String>>,
}

#[derive(Serialize)]
struct User {
    id: String,
    name: String,
    email: String,
}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    
    let request: Request = serde_json::from_str(&input).unwrap();
    let response = handler(request);
    
    let output = serde_json::to_string(&response).unwrap();
    io::stdout().write_all(output.as_bytes()).unwrap();
}

fn handler(request: Request) -> Response {
    let mut headers = HashMap::new();
    headers.insert("Content-Type".to_string(), "application/json".to_string());
    
    match request.method.as_str() {
        "GET" => {
            let users = vec![
                User { id: "1".to_string(), name: "Alice".to_string(), email: "alice@example.com".to_string() },
                User { id: "2".to_string(), name: "Bob".to_string(), email: "bob@example.com".to_string() },
            ];
            
            Response {
                status: 200,
                body: json!({"users": users}),
                headers: Some(headers),
            }
        }
        
        "POST" => {
            let name = request.body["name"].as_str().unwrap_or("");
            let email = request.body["email"].as_str().unwrap_or("");
            
            let user = User {
                id: "3".to_string(),
                name: name.to_string(),
                email: email.to_string(),
            };
            
            Response {
                status: 201,
                body: json!(user),
                headers: Some(headers),
            }
        }
        
        _ => Response {
            status: 405,
            body: json!({"error": "Method not allowed"}),
            headers: Some(headers),
        }
    }
}
```

### Webhook Processor

```rust
use hmac::{Hmac, Mac};
use serde::{Deserialize, Serialize};
use serde_json::json;
use sha2::Sha256;
use std::collections::HashMap;
use std::env;
use std::io::{self, Read, Write};

type HmacSha256 = Hmac<Sha256>;

#[derive(Deserialize)]
struct Request {
    body: serde_json::Value,
    headers: HashMap<String, String>,
}

#[derive(Serialize)]
struct Response {
    status: u16,
    body: serde_json::Value,
}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    
    let request: Request = serde_json::from_str(&input).unwrap();
    let response = handler(request);
    
    let output = serde_json::to_string(&response).unwrap();
    io::stdout().write_all(output.as_bytes()).unwrap();
}

fn handler(request: Request) -> Response {
    let secret = env::var("WEBHOOK_SECRET").unwrap_or_default();
    let signature = request.headers.get("x-signature").cloned().unwrap_or_default();
    
    if !verify_signature(&request.body, &signature, &secret) {
        return Response {
            status: 401,
            body: json!({"error": "Invalid signature"}),
        };
    }
    
    // Process webhook (async in real implementation)
    Response {
        status: 200,
        body: json!({"received": true}),
    }
}

fn verify_signature(body: &serde_json::Value, signature: &str, secret: &str) -> bool {
    if !signature.starts_with("sha256=") {
        return false;
    }
    
    let body_str = body.to_string();
    
    let mut mac = HmacSha256::new_from_slice(secret.as_bytes()).unwrap();
    mac.update(body_str.as_bytes());
    let result = mac.finalize();
    let expected = hex::encode(result.into_bytes());
    
    let provided = &signature[7..]; // Remove "sha256=" prefix
    
    // Constant-time comparison
    expected.as_bytes().len() == provided.as_bytes().len()
        && expected.as_bytes().iter().zip(provided.as_bytes()).all(|(a, b)| a == b)
}
```

### Data Transformation

```rust
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::collections::HashMap;
use std::io::{self, Read, Write};

#[derive(Deserialize)]
struct Request {
    body: InputData,
}

#[derive(Deserialize)]
struct InputData {
    id: String,
    first_name: String,
    last_name: String,
    email: String,
    created_at: String,
}

#[derive(Serialize)]
struct OutputData {
    id: String,
    name: String,
    email: String,
    timestamp: String,
}

#[derive(Serialize)]
struct Response {
    status: u16,
    body: OutputData,
    headers: Option<HashMap<String, String>>,
}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap();
    
    let request: Request = serde_json::from_str(&input).unwrap();
    
    let transformed = OutputData {
        id: request.body.id,
        name: format!("{} {}", request.body.first_name, request.body.last_name).trim().to_string(),
        email: request.body.email.to_lowercase(),
        timestamp: request.body.created_at,
    };
    
    let mut headers = HashMap::new();
    headers.insert("Content-Type".to_string(), "application/json".to_string());
    
    let response = Response {
        status: 200,
        body: transformed,
        headers: Some(headers),
    };
    
    let output = serde_json::to_string(&response).unwrap();
    io::stdout().write_all(output.as_bytes()).unwrap();
}
```

## Environment Variables

Access environment variables using `std::env`:

```rust
use std::env;

fn main() {
    let api_key = env::var("API_KEY").unwrap_or_default();
    let db_url = env::var("DATABASE_URL").unwrap_or_default();
    let debug = env::var("DEBUG").map(|v| v == "true").unwrap_or(false);
    
    // Use variables
    println!("API Key present: {}", !api_key.is_empty());
}
```

## File System

The `/tmp` directory is available:

```rust
use std::fs;
use std::path::Path;

fn write_to_temp(filename: &str, data: &str) -> std::io::Result<()> {
    let path = Path::new("/tmp").join(filename);
    fs::write(path, data)
}

fn read_from_temp(filename: &str) -> std::io::Result<String> {
    let path = Path::new("/tmp").join(filename);
    fs::read_to_string(path)
}
```

## Error Handling

```rust
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::io::{self, Read, Write};

#[derive(Deserialize)]
struct Request {
    body: serde_json::Value,
}

#[derive(Serialize)]
struct Response {
    status: u16,
    body: serde_json::Value,
}

fn main() {
    let mut input = String::new();
    if let Err(e) = io::stdin().read_to_string(&mut input) {
        let response = Response {
            status: 400,
            body: json!({"error": format!("Failed to read input: {}", e)}),
        };
        let _ = writeln!(io::stdout(), "{}", serde_json::to_string(&response).unwrap());
        return;
    }
    
    let request: Request = match serde_json::from_str(&input) {
        Ok(req) => req,
        Err(e) => {
            let response = Response {
                status: 400,
                body: json!({"error": format!("Invalid JSON: {}", e)}),
            };
            let _ = writeln!(io::stdout(), "{}", serde_json::to_string(&response).unwrap());
            return;
        }
    };
    
    let response = handler(request);
    let output = serde_json::to_string(&response).unwrap();
    io::stdout().write_all(output.as_bytes()).unwrap();
}

fn handler(request: Request) -> Response {
    match process_request(&request.body) {
        Ok(result) => Response {
            status: 200,
            body: result,
        },
        Err(e) => Response {
            status: 500,
            body: json!({"error": e}),
        },
    }
}

fn process_request(body: &serde_json::Value) -> Result<serde_json::Value, String> {
    // Processing logic
    Ok(json!({"status": "ok"}))
}
```

## Building for WASM

```bash
# Build for WASI target
cargo build --target wasm32-wasi --release

# Or use cargo wasi plugin
cargo wasi build --release

# Optimize with wasm-opt (optional)
wasm-opt -O3 target/wasm32-wasi/release/my-function.wasm -o optimized.wasm
```

## functionfly.jsonc Configuration

```jsonc
{
  "name": "my-rust-function",
  "runtime": "rust-wasm",
  "wasm": {
    "entrypoint": "main",
    "wasi": true
  },
  "limits": {
    "timeout": 30,
    "memory": 128
  }
}
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 128 MB | 1024 MB |
| CPU | 1 vCPU | 4 vCPU |

Rust/WASM functions typically use less memory than native runtimes.

## Cold Start

Rust/WASM functions have very fast cold starts:
- First invocation after deployment: ~10-50ms
- Subsequent invocations: ~1-5ms

## Best Practices

1. **Use release profile** - Optimize for size and speed
2. **Enable LTO** - Link-time optimization for smaller binaries
3. **Minimize dependencies** - Each crate adds to binary size
4. **Use `strip = true`** - Remove debug symbols
5. **Prefer `serde_json::Value`** - For flexible JSON handling
6. **Handle all errors** - Return appropriate status codes
7. **Use constant-time comparison** - For security-sensitive operations
