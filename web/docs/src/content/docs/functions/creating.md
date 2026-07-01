---
title: Creating Functions
description: Write and deploy your first serverless function on FunctionFly
sidebar:
  order: 2
---

# Creating Functions

This guide walks through creating, deploying, and invoking a serverless function.

## Initialize a New Function

```bash
# Create with default runtime
ff init my-function

# Specify a runtime
ff init my-function --runtime python
ff init my-function --runtime javascript
ff init my-function --runtime go
```

## Write Your Handler

### Python

```python
# main.py
import json

async def handler(request):
    name = request.get("body", {}).get("name", "World")
    return {
        "status": 200,
        "body": json.dumps({"message": f"Hello, {name}!"}),
        "headers": {"Content-Type": "application/json"}
    }
```

### JavaScript / TypeScript

```javascript
// index.js
export default async function handler(request) {
  const name = request.body?.name || "World";
  return {
    status: 200,
    body: { message: `Hello, ${name}!` },
    headers: { "Content-Type": "application/json" }
  };
}
```

### Go

```go
// main.go
package main

import (
    "encoding/json"
    "net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        name = "World"
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Hello, " + name + "!",
    })
}
```

### Rust (WASM)

```rust
// src/lib.rs
use serde_json::json;

#[no_mangle]
pub extern "C" fn handler(body: &str) -> String {
    let response = json!({ "message": "Hello from Rust!" });
    response.to_string()
}
```

## Deploy

```bash
# Deploy to FunctionFly
ff deploy

# Get the deployment URL
ff info
```

## Invoke

```bash
# Via CLI
ff invoke my-function --data '{"name": "FunctionFly"}'

# Via HTTP
curl https://api.functionfly.com/v1/execute/<function-id> \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"name": "FunctionFly"}'
```

## Next Steps

- [Function Structure](/functions/structure/) — Request/response format
- [Testing Functions](/functions/testing/) — Local dev and Playground
- [Environment Variables](/guides/environment-variables/) — Configuration
- [Deployment Guide](/deployment/) — Deployment options
