# FunctionFly WasmEdge Runtime

Production-ready C/C++ and WebAssembly execution runtime for FunctionFly with full WasmEdge SDK integration.

## Features

- **WasmEdge SDK 0.14**: Full embedded WASM execution via SDK (not subprocess)
- **WasmEdge WASI 0.2 Support**: Execute C/C++ and other WASI-compatible languages
- **WebAssembly Sandboxing**: Memory-safe execution with resource limits
- **Fuel Metering**: Prevents infinite loops with instruction counting
- **Security Hardening**: Syscall filtering, network controls, path traversal prevention
- **Resource Limits**: Memory, CPU time, wall time, and fuel limits
- **NATS Integration**: Orchestrator communication support

## Supported Languages

### C/C++ (Primary)

Compile your C/C++ code to WebAssembly using:

```bash
# Using clang
clang --target=wasm32-wasi -o function.wasm function.c -lc

# Using emscripten (for POSIX compatibility)
emcc function.c -o function.js -s WASI=1
```

### Rust

```bash
rustup target add wasm32-wasi
cargo build --target wasm32-wasi --release
```

### Other Languages

Any language that compiles to WASI 0.2 compatible WebAssembly:
- Go (via TinyGo or Go 1.21+ with wasip1)
- Kotlin/Wasm (experimental)
- AssemblyScript
- Grain

## Quick Start

### Prerequisites

- Rust 1.70+
- WasmEdge SDK (optional, for native execution)
- NATS server (optional, for orchestrator integration)

### Build

```bash
cargo build --release
```

### Run

```bash
# With default settings
./target/release/functionfly-wasmedge-runtime

# With custom settings
WASMEDGE_PORT=8092 \
MAX_MEMORY_MB=1024 \
MAX_FUEL=50000000 \
./target/release/functionfly-wasmedge-runtime
```

### Docker

```bash
docker build -t functionfly-wasmedge:latest -f Dockerfile .
docker run -p 8092:8092 functionfly-wasmedge:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8092` |
| `MAX_CONCURRENT` | Max concurrent executions | `100` |
| `MAX_MEMORY_MB` | Max memory per execution (MB) | `512` |
| `MAX_FUEL` | Max fuel units per execution | `10_000_000` |
| `MAX_EXECUTION_TIME_SECS` | Max execution time (seconds) | `30` |
| `SANDBOX_ENABLED` | Enable WASM sandbox | `true` |
| `NATS_URL` | NATS server URL | None |
| `WORKING_DIR` | Sandbox working directory | None |

### API Endpoints

#### Health Check

```bash
curl http://localhost:8092/health
```

Response:
```json
{
  "status": "healthy",
  "runtime": "wasmedge",
  "version": "0.1.0"
}
```

#### Execute WASM

```bash
curl -X POST http://localhost:8092/execute \
  -H "Content-Type: application/json" \
  -d '{
    "execution_id": "func-123",
    "wasm": "$(cat function.wasm | base64)",
    "timeout_ms": 30000
  }'
```

Response:
```json
{
  "execution_id": "func-123",
  "success": true,
  "output": "[WasmEdge] Executed successfully. Fuel consumed: 1234",
  "execution_time_ms": 45,
  "memory_used_mb": 2,
  "fuel_consumed": 1234
}
```

#### Metrics

```bash
curl http://localhost:8092/metrics
```

Response:
```json
{
  "total_executions": 1523,
  "successful_executions": 1500,
  "failed_executions": 23,
  "runtime": "wasmedge"
}
```

## Security

### Default Security Policy

The runtime enforces a restrictive security policy by default:

- **Syscall Filtering**: Only essential WASI syscalls allowed
  - `fd_write`, `fd_read`, `fd_close`, `fd_seek`
  - `path_open`, `clock_time_get`, `random_get`, `proc_exit`
- **Network Controls**: Metadata endpoints (AWS, GCP, Azure) are blocked by default
- **Filesystem Isolation**: Only preopened directories accessible
- **Environment Sanitization**: Dangerous env vars (LD_*, PATH, SHELL) are blocked
- **Fuel Metering**: Prevents infinite loops with instruction counting

### Production Deployment Checklist

- [ ] Review and adjust syscall whitelist for your C/C++ code requirements
- [ ] Configure network whitelist (`ALLOWED_HOSTS`) for your API endpoints
- [ ] Set appropriate memory limits (`MAX_MEMORY_MB`) for your workload
- [ ] Review fuel limits (`MAX_FUEL`) for complex C/C++ functions
- [ ] Enable syscall auditing in production (`AUDIT_SYSCALLS=true`)
- [ ] Test with production workloads before deployment

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    WasmEdge Runtime                          │
├─────────────────────────────────────────────────────────────┤
│  HTTP API (Axum)  │  Metrics  │  Orchestrator (NATS)       │
├───────────────────┴───────────┴────────────────────────────┤
│                    Sandbox Engine                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Fuel      │  │   Memory    │  │   Syscall   │        │
│  │   Meter     │  │   Limiter   │  │   Filter    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│               WasmEdge WASI 0.2 Runtime                      │
│  ┌─────────────────────────────────────────────────┐        │
│  │         C/C++ WASM Execution Environment        │        │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## Example: C Function

### Source (hello.c)

```c
#include <stdio.h>

int main() {
    printf("Hello from WasmEdge!\n");
    return 0;
}
```

### Compile

```bash
clang --target=wasm32-wasi -o hello.wasm hello.c -lc
```

### Execute

```bash
curl -X POST http://localhost:8092/execute \
  -H "Content-Type: application/json" \
  -d '{
    "execution_id": "hello-1",
    "wasm": "'$(base64 -w0 hello.wasm)'"
  }'
```

## Troubleshooting

### "WASM binary too short"

The WASM binary is invalid or corrupted. Verify it was compiled correctly:

```bash
# Check WASM magic number
hexdump -C function.wasm | head -n 1
# Should show: 00 61 73 6d 01 00 00 00

# Validate with wasm-validate (if installed)
wasm-validate function.wasm
```

### "Fuel limit exceeded"

Your program used too many instructions. Increase the limit:

```bash
MAX_FUEL=50000000 ./functionfly-wasmedge-runtime
```

Or optimize your code to be more efficient.

### "Memory limit exceeded"

Increase the memory limit:

```bash
MAX_MEMORY_MB=1024 ./functionfly-wasmedge-runtime
```

## License

MIT
