# Runtimes Directory

This directory contains references to external runtime engines used by FunctionFly.

## SAR (Stateful Agent Runtime)

**Location:** https://github.com/functionfly/sar

SAR is a separate Rust-based runtime service that executes AI agent workflows. It runs independently on port 8082 and is called by the orchestrator via HTTP/gRPC.

### Quick Start

```bash
# Clone SAR to a separate location
git clone https://github.com/functionfly/sar /path/to/sar

# Build and run
cd /path/to/sar
cargo build --release
./target/release/functionfly-sar --api-port 8082
```

### Configuration

```bash
export SAR_API_KEY="your-secure-api-key"
export NATS_URL="nats://localhost:4222"
export DATABASE_URL="postgres://user:pass@localhost/sar"
```

### Docker

```bash
cd /path/to/sar
docker build -t functionfly-sar:latest .
docker run -p 8082:8082 \
  -e SAR_API_KEY="your-key" \
  functionfly-sar:latest
```

### Security

See [SAR Security Documentation](https://github.com/functionfly/sar/blob/main/docs/CONTAINERIZATION.md) for:
- Container security settings
- TLS configuration
- Audit logging setup
- Production deployment checklist

## Other Runtimes

| Runtime | Status | Purpose |
|---------|--------|---------|
| `nodejs/` | Active | Node.js function execution |
| `deno/` | Active | Deno function execution |
| `python.wasm` | Active | Python WASM execution |
| `microvm/` | Active | MicroVM sandbox |
| `kotlin/` | Active | Kotlin/JVM execution |
| `ruby/` | Active | Ruby execution |
| `wasmedge/` | Active | WasmEdge runtime |
| `prism/` | Active | Prism runtime |

## Adding a New Runtime

1. Create runtime in a separate repository
2. Add entry to `internal/api/handlers/runtime/handlers.go`
3. Document configuration requirements