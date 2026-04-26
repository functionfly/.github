---
title: C SDK
description: C SDK for building FunctionFly functions
---

# C SDK

The FunctionFly C SDK provides a header-only library for building FunctionFly functions in C, compiled to WebAssembly.

## Installation

Copy `functionfly.h` into your project:

```bash
curl -O https://raw.githubusercontent.com/functionfly/sdk-c/main/functionfly.h
```

Or include it from the FunctionFly SDK repository:

```
my-function/
├── functionfly.h       # SDK header
├── src/
│   └── main.c
└── functionfly.jsonc
```

## Quick Start

```c
#include "functionfly.h"

void init(void) {
    // Optional: called once at cold start
}

const char* execute(const char* input, int32_t input_len) {
    return ff_strdup("{\"message\": \"Hello from C!\"}");
}

void* alloc(int32_t size) { return malloc(size); }
void dealloc(void* ptr) { free(ptr); }

const char* metadata(void) {
    return "{\"name\": \"my-function\", \"runtime\": \"c\"}";
}
```

## Entry Points

Every C function must export these entry points:

| Function | Signature | Description |
|----------|-----------|-------------|
| `init` | `void init(void)` | Called once at cold start |
| `execute` | `const char* execute(const char* input, int32_t input_len)` | Main execution; returns allocated string |
| `alloc` | `void* alloc(int32_t size)` | Memory allocator for runtime |
| `dealloc` | `void dealloc(void* ptr)` | Memory deallocator for runtime |
| `metadata` | `const char* metadata(void)` | Returns JSON metadata about the function |

## Host Functions

Access FunctionFly platform services:

```c
// Logging
ff_log("Processing request");
functionfly_log(msg, strlen(msg));

// Environment variables
char buf[256];
int32_t len = functionfly_get_env("API_KEY", 7, buf, sizeof(buf));

// HTTP fetch
char response[4096];
int32_t resp_len = functionfly_fetch(url, strlen(url), response, sizeof(response));

// Key-value store
functionfly_kv_set("key", 3, "value", 5);
int32_t val_len = functionfly_kv_get("key", 3, buf, sizeof(buf));

// Crypto
char hash[64];
functionfly_crypto_hash("sha256", 6, data, data_len, hash, sizeof(hash));
```

## Utility Functions

The SDK provides inline helpers:

| Function | Description |
|----------|-------------|
| `ff_strdup(s)` | Duplicate a string (allocates memory) |
| `ff_string_len(s)` | Get string length as `int32_t` |
| `ff_log(msg)` | Log a C string (convenience wrapper) |

## Export Macro

The `FF_EXPORT` macro handles platform-specific export:

- **Emscripten**: Uses `EMSCRIPTEN_KEEPALIVE` to prevent dead-code elimination
- **WASI-SDK**: Uses `__attribute__((visibility("default")))`

## Building

### With Emscripten

```bash
source emsdk/emsdk_env.sh
emcc src/main.c -o function.js \
    -s EXPORTED_FUNCTIONS='["_init","_execute","_alloc","_dealloc","_metadata","_malloc","_free"]' \
    -s EXPORTED_RUNTIME_METHODS='["ccall","cwrap"]' \
    -O3
```

### With WASI-SDK

```bash
$WASI_SDK_PATH/bin/clang src/main.c -o function.wasm \
    --target=wasm32-wasi -O3 -nostartfiles \
    -Wl,--export=init -Wl,--export=execute \
    -Wl,--export=alloc -Wl,--export=dealloc -Wl,--export=metadata
```

## C++ Support

The header is C++ compatible:

```cpp
#include "functionfly.h"

extern "C" {
    void init(void) {}
    const char* execute(const char* input, int32_t input_len) {
        return ff_strdup("{\"message\": \"Hello from C++!\"}");
    }
    void* alloc(int32_t size) { return malloc(size); }
    void dealloc(void* ptr) { free(ptr); }
    const char* metadata(void) {
        return "{\"name\": \"my-cpp-function\", \"runtime\": \"cpp\"}";
    }
}
```

## API Reference

### Host Function Signatures

| Function | Parameters | Returns |
|----------|-----------|---------|
| `functionfly_log` | `(const char* msg, int32_t len)` | `void` |
| `functionfly_get_env` | `(const char* key, int32_t key_len, char* buf, int32_t buf_len)` | `int32_t` (bytes written) |
| `functionfly_fetch` | `(const char* url, int32_t url_len, char* buf, int32_t buf_len)` | `int32_t` (bytes written) |
| `functionfly_kv_get` | `(const char* key, int32_t key_len, char* buf, int32_t buf_len)` | `int32_t` (bytes written) |
| `functionfly_kv_set` | `(const char* key, int32_t key_len, const char* val, int32_t val_len)` | `int32_t` (0 = success) |
| `functionfly_crypto_hash` | `(const char* algo, int32_t algo_len, const char* data, int32_t data_len, char* buf, int32_t buf_len)` | `int32_t` (bytes written) |
