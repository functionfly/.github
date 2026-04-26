---
title: WASM & WebAssembly
description: Universal WASM runtime pipeline for polyglot serverless functions
---

# WASM & WebAssembly

FunctionFly's universal WASM runtime pipeline lets you write serverless functions in any supported language and compiles them to WebAssembly for portable, secure, edge-native execution.

## Overview

FunctionFly accepts source code in Rust, Go, C/C++, Ruby, Kotlin, Swift, JavaScript, TypeScript, and Python. Languages with mature WASM toolchains compile directly to `.wasm` binaries. Others run inside embedded interpreters compiled to WASM. All functions execute in the same secure wasmtime-based runtime.

### How It Works

```
Source Code → Compiler/Toolchain → .wasm Binary → wasmtime Runtime → Response
```

1. **Upload** — Submit your function source code via API, dashboard, or CLI
2. **Compile** — FunctionFly detects the runtime and invokes the appropriate compiler
3. **Optimize** — The resulting WASM binary is optimized for size and cold start
4. **Execute** — The binary runs in a sandboxed wasmtime instance with host function access

## Supported Languages

| Language | Compilation | Runtime ID | Status |
|----------|-------------|------------|--------|
| Rust | Direct (`cargo build --target wasm32-wasip1`) | `rust-wasm` | Stable |
| Go | Direct (`GOOS=wasip1 GOARCH=wasm go build`) | `go-wasm` | Stable |
| C | Emscripten or WASI-SDK | `c` | Stable |
| C++ | Emscripten or WASI-SDK | `cpp` | Stable |
| Ruby | mruby interpreter embedded in WASM | `ruby-wasm` | Beta |
| Kotlin | Kotlin/WASM compiler | `kotlin-wasm` | Beta |
| Swift | SwiftWasm toolchain | `swift-wasm` | Experimental |
| JavaScript | Javy (JS → WASM) | `js-wasm` | Stable |
| TypeScript | Javy via JS compilation | `ts-wasm` | Stable |
| Python | Python interpreter in WASM | `python-wasm` | Stable |

## Compilation Strategies

### Direct Compilation

Rust, Go, and C/C++ compile directly to WASM with no interpreter overhead:

- **Rust** — `cargo build --target wasm32-wasip1 --release`
- **Go** — `GOOS=wasip1 GOARCH=wasm go build -o function.wasm .`
- **C/C++** — `emcc` (Emscripten) or `clang --target=wasm32-wasi` (WASI-SDK)

These produce the smallest binaries and fastest cold starts.

### Interpreter-Embedded

Ruby runs inside an mruby interpreter compiled to WASM:

- Source code is interpreted at runtime inside the WASM sandbox
- No gems — only mruby's built-in standard library
- Moderate cold start due to interpreter initialization

### Hybrid

Kotlin supports two paths:

- **Kotlin/WASM** (recommended) — Direct compilation via JetBrains' Kotlin/WASM compiler
- **Kotlin → JS → Javy** (fallback) — Compile Kotlin to JavaScript, then to WASM via Javy

## Security Model

All WASM functions run in wasmtime's sandboxed execution environment:

- **Memory isolation** — Each function has its own linear memory
- **Capability-based** — Host functions are explicitly granted (network, filesystem, etc.)
- **No filesystem access** — Only `/tmp` is writable
- **Resource limits** — Timeout, memory, and CPU caps enforced per invocation
- **No network by default** — Outbound HTTP requires explicit capability grant

## Host Functions

WASM functions can access platform services through host function bindings:

| Capability | Description | Languages |
|------------|-------------|-----------|
| `log` | Write to execution logs | All |
| `get_env` | Read environment variables | All |
| `fetch` | Outbound HTTP requests | All |
| `kv_get` / `kv_set` | Key-value store | All |
| `ai` | AI inference | All |
| `crypto_hash` | Cryptographic hashing | All |

## Performance

| Metric | Direct (Rust/Go/C) | Interpreter (Ruby) | Hybrid (Kotlin/Swift) |
|--------|---------------------|--------------------|-----------------------|
| Cold start | 1–50 ms | 50–150 ms | 30–120 ms |
| Execution | Near-native | 2–5× slower | 1.5–3× slower |
| Binary size | 50 KB – 2 MB | 2–5 MB | 2–10 MB |
| Memory usage | 10–50 MB | 50–128 MB | 64–256 MB |

## Choosing a Language

| If you need... | Use |
|----------------|-----|
| Maximum performance and small binaries | **Rust** |
| Familiar syntax with good WASM support | **Go** or **C** |
| Quick prototyping with minimal setup | **JavaScript** or **Python** |
| Type safety with modern tooling | **Kotlin** or **Swift** |
| Ruby ecosystem compatibility | **Ruby** (mruby) |
| Low-level control over memory | **C** or **Rust** |

## Toolchain Detection

FunctionFly auto-detects available toolchains on the build system:

```bash
# Check available toolchains
curl https://api.functionfly.com/v1/runtimes/toolchains
```

Response:

```json
{
  "toolchains": {
    "rust": { "available": true, "version": "1.78.0" },
    "go": { "available": true, "version": "1.22.0" },
    "emscripten": { "available": true, "version": "3.1.50" },
    "wasi_sdk": { "available": true, "version": "21.0" },
    "mruby": { "available": true, "version": "3.2.0" },
    "kotlin_wasm": { "available": true, "version": "2.1.0" },
    "swiftwasm": { "available": false },
    "javy": { "available": true, "version": "3.0.0" }
  }
}
```

## SDKs

Each supported language has a dedicated SDK for building functions:

| SDK | Language | Package |
|-----|----------|---------|
| [C SDK](/sdks/c/) | C/C++ | Header file (`functionfly.h`) |
| [Go SDK](/sdks/go/) | Go | `go get github.com/functionfly/sdk-go` |
| [Rust SDK](/sdks/rust/) | Rust | `functionfly-sdk` crate |
| [Ruby SDK](/sdks/ruby/) | Ruby | Included in project |
| [Kotlin SDK](/sdks/kotlin/) | Kotlin | Gradle dependency |
| [Swift SDK](/sdks/swift/) | Swift | Swift package |
| [JavaScript SDK](/sdks/javascript/) | JS/TS | `npm install @functionfly/sdk` |
| [Python SDK](/sdks/python/) | Python | `pip install functionfly-sdk` |

## Configuration

All WASM functions use `functionfly.jsonc` for configuration:

```jsonc
{
  "name": "my-function",
  "runtime": "rust-wasm",
  "wasm": {
    "entrypoint": "execute",
    "wasi": true
  },
  "limits": {
    "timeout": 30,
    "memory": 128
  },
  "capabilities": {
    "network": ["api.example.com"],
    "kv": true,
    "ai": false
  }
}
```

### Runtime Values

Use the `runtime` field to specify the compilation target:

| Value | Language | Toolchain |
|-------|----------|-----------|
| `rust-wasm` | Rust | cargo + wasm32-wasip1 |
| `go-wasm` | Go | go build with WASM |
| `c` | C | Emscripten / WASI-SDK |
| `cpp` | C++ | Emscripten / WASI-SDK |
| `ruby-wasm` | Ruby | mruby interpreter |
| `kotlin-wasm` | Kotlin | Kotlin/WASM compiler |
| `swift-wasm` | Swift | SwiftWasm |
| `js-wasm` | JavaScript | Javy |
| `ts-wasm` | TypeScript | Javy via JS |
| `python-wasm` | Python | Python WASM interpreter |

## Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 128 MB | 1024 MB |
| CPU | 1 vCPU | 4 vCPU |
| Binary size | — | 50 MB |
| `/tmp` storage | — | 512 MB |
