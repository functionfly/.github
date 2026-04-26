---
title: C/C++ via WASM Runtime
description: C and C++ WebAssembly runtime environment for FunctionFly functions
---

# C/C++ via WASM Runtime

FunctionFly's C/C++ runtime compiles your code to WebAssembly (WASM) using Emscripten or WASI-SDK for efficient, sandboxed execution.

## Overview

C and C++ functions are compiled to WebAssembly and run in a secure, lightweight WASM runtime. This provides:

- **Fast cold starts** - WASM modules initialize quickly
- **Small footprint** - Compact binary sizes
- **Sandboxed** - Memory-safe, isolated execution
- **Portable** - Same binary runs on any platform
- **Native performance** - Near-native speed for compute-heavy workloads

## Supported Toolchains

| Toolchain | Status | Notes |
|-----------|--------|-------|
| Emscripten | Supported | Recommended; full libc support |
| WASI-SDK | Supported | Lightweight alternative |
| Clang + wasm32 target | Supported | Requires manual sysroot setup |

## Project Structure

```
my-function/
├── src/
│   └── main.c          # Entry point (or main.cpp)
├── functionfly.jsonc   # Function config
└── Makefile            # Optional build script
```

## Function Structure

A C function uses the standard FunctionFly entry points:

```c
#include "functionfly.h"
#include <stdio.h>
#include <string.h>

// Called once at cold start
void init(void) {
    // Optional initialization
}

// Main execution: receives input, returns output
const char* execute(const char* input, int32_t input_len) {
    // Process input (JSON string) and return output (JSON string)
    const char* response = "{\"message\": \"Hello from C!\"}";
    return ff_strdup(response);
}

// Memory management
void* alloc(int32_t size) {
    return malloc(size);
}

void dealloc(void* ptr) {
    free(ptr);
}

// Function metadata
const char* metadata(void) {
    return "{\"name\": \"my-function\", \"runtime\": \"c\", \"version\": \"1.0.0\"}";
}
```

## Example Functions

### HTTP API Handler

```c
#include "functionfly.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

void init(void) {}

const char* execute(const char* input, int32_t input_len) {
    // In a real implementation, parse the JSON input
    // This example returns a simple JSON response
    const char* response =
        "{\"status\": 200, "
        "\"body\": {\"users\": ["
        "{\"id\": \"1\", \"name\": \"Alice\", \"email\": \"alice@example.com\"},"
        "{\"id\": \"2\", \"name\": \"Bob\", \"email\": \"bob@example.com\"}"
        "]}, "
        "\"headers\": {\"Content-Type\": \"application/json\"}}";

    return ff_strdup(response);
}

void* alloc(int32_t size) { return malloc(size); }
void dealloc(void* ptr) { free(ptr); }

const char* metadata(void) {
    return "{\"name\": \"user-api\", \"runtime\": \"c\"}";
}
```

### Webhook Processor

```c
#include "functionfly.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

void init(void) {}

const char* execute(const char* input, int32_t input_len) {
    // Read webhook secret from environment
    char secret_buf[256];
    int32_t secret_len = functionfly_get_env(
        "WEBHOOK_SECRET", 13,
        secret_buf, sizeof(secret_buf)
    );

    if (secret_len <= 0) {
        return ff_strdup("{\"status\": 401, \"body\": {\"error\": \"Missing secret\"}}");
    }

    // In production: verify HMAC signature from headers
    // For now, acknowledge receipt
    return ff_strdup("{\"status\": 200, \"body\": {\"received\": true}}");
}

void* alloc(int32_t size) { return malloc(size); }
void dealloc(void* ptr) { free(ptr); }

const char* metadata(void) {
    return "{\"name\": \"webhook-handler\", \"runtime\": \"c\"}";
}
```

### Data Transformation

```c
#include "functionfly.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

void init(void) {}

const char* execute(const char* input, int32_t input_len) {
    // Use host function for logging
    ff_log("Processing data transformation");

    // In a real implementation, parse input JSON and transform
    // This example returns a transformed structure
    const char* response =
        "{\"status\": 200, "
        "\"body\": {"
        "\"id\": \"123\", "
        "\"name\": \"Alice Smith\", "
        "\"email\": \"alice@example.com\", "
        "\"timestamp\": \"2025-01-15T10:30:00Z\""
        "}}";

    return ff_strdup(response);
}

void* alloc(int32_t size) { return malloc(size); }
void dealloc(void* ptr) { free(ptr); }

const char* metadata(void) {
    return "{\"name\": \"data-transform\", \"runtime\": \"c\"}";
}
```

## Host Functions

The C SDK provides access to FunctionFly host functions:

| Function | Description |
|----------|-------------|
| `functionfly_log(msg, len)` | Log a message |
| `functionfly_get_env(key, key_len, buf, buf_len)` | Get environment variable |
| `functionfly_fetch(url, url_len, buf, buf_len)` | HTTP fetch |
| `functionfly_kv_get(key, key_len, buf, buf_len)` | Read from key-value store |
| `functionfly_kv_set(key, key_len, val, val_len)` | Write to key-value store |
| `functionfly_crypto_hash(algo, algo_len, data, data_len, buf, buf_len)` | Hash data |

## Environment Variables

Access environment variables using the host function:

```c
char buf[1024];
int32_t len = functionfly_get_env("API_KEY", 7, buf, sizeof(buf));
if (len > 0) {
    buf[len] = '\0';
    ff_log(buf);
}
```

## Building for WASM

### Using Emscripten

```bash
# Activate Emscripten SDK
source emsdk/emsdk_env.sh

# Compile to WASM
emcc src/main.c -o function.js \
    -s EXPORTED_FUNCTIONS='["_init", "_execute", "_alloc", "_dealloc", "_metadata", "_malloc", "_free"]' \
    -s EXPORTED_RUNTIME_METHODS='["ccall", "cwrap"]' \
    -O3

# The .wasm file is the deployable artifact
```

### Using WASI-SDK

```bash
# Set WASI-SDK path
export WASI_SDK_PATH=/opt/wasi-sdk

# Compile to WASM
$WASI_SDK_PATH/bin/clang src/main.c -o function.wasm \
    --target=wasm32-wasi \
    -O3 \
    -nostartfiles \
    -Wl,--export=init \
    -Wl,--export=execute \
    -Wl,--export=alloc \
    -Wl,--export=dealloc \
    -Wl,--export=metadata
```

## functionfly.jsonc Configuration

```jsonc
{
  "name": "my-c-function",
  "runtime": "c",
  "wasm": {
    "entrypoint": "execute",
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

## Cold Start

C/C++ WASM functions have very fast cold starts:
- First invocation after deployment: ~5-20ms
- Subsequent invocations: ~1-3ms

## Best Practices

1. **Use release optimizations** - Compile with `-O3` for best performance
2. **Minimize binary size** - Strip debug symbols and use `-Os` if size matters
3. **Handle memory carefully** - WASM has a linear memory model; avoid leaks
4. **Use the SDK header** - Include `functionfly.h` for host function bindings
5. **Return allocated strings** - Use `ff_strdup()` for return values
6. **Validate inputs** - Check input lengths before processing
7. **Use host functions** - Prefer `functionfly_*` over reimplementing functionality
