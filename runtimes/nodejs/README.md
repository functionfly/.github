# FunctionFly Node.js Runtime

A secure, high-performance JavaScript execution runtime built with Rust, designed for serverless function execution.

## Features

- **QuickJS WASM Engine** - Lightweight JavaScript execution via WebAssembly
- **Sandbox Isolation** - Secure execution with module restrictions
- **Resource Limits** - Configurable memory and CPU limits
- **Timeout Management** - Granular execution timeouts
- **Host Functions** - Built-in support for fetch, console, timers, etc.
- **Metrics Collection** - Prometheus-compatible metrics

## Architecture

```
┌─────────────────────────────────────────┐
│         FunctionFly Node.js Runtime     │
├─────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────┐  │
│  │   Executor  │  │   Sandbox       │  │
│  │             │  │                 │  │
│  │ - Code      │  │ - Isolation     │  │
│  │   Caching   │  │ - Module        │  │
│  │ - QuickJS   │  │   Restrictions  │  │
│  │   WASM      │  │ - Security      │  │
│  └─────────────┘  └─────────────────┘  │
│                                         │
│  ┌─────────────┐  ┌─────────────────┐  │
│  │   Timeout   │  │   Memory        │  │
│  │   Manager   │  │   Limiter       │  │
│  └─────────────┘  └─────────────────┘  │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │      Host Functions            │    │
│  │  - fetch    - console          │    │
│  │  - timers   - crypto (limited)│    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

## Supported Runtimes

| Runtime | Version | Status |
|---------|---------|--------|
| Node.js | 20.x LTS | ✅ Stable |
| Node.js | 18.x LTS | ✅ Stable |
| Deno    | latest  | 🟡 Beta  |

## Installation

```bash
# Build the runtime
cargo build --release

# Run the binary
./target/release/functionfly-nodejs --runtime node20 --code "export function handler(input) { return 'Hello, ' + input + '!'; }" --input "World"
```

## Configuration

```rust
use nodejs_runtime::{RuntimeConfig, create_runtime};

let config = RuntimeConfig {
    version: RuntimeVersion::Node20,
    max_memory_mb: 128,
    max_timeout_ms: 30000,
    network_enabled: false,
    // ... more options
};

let runtime = create_runtime(config)?;
```

## Environment Variables

Functions running in the Node.js runtime have access to:

- `NODE_ENV` - Set to "production" or "development"
- Custom environment variables (if configured)

## Blocked Modules

The following Node.js modules are blocked for security:

- `child_process` - Arbitrary command execution
- `fs` - File system access
- `net`, `tls`, `http`, `https` - Network connections
- `dns` - DNS lookups
- `worker_threads` - Multi-threading

## Development

```bash
# Run tests
cargo test

# Run with logging
RUST_LOG=debug cargo run -- --code "..." --input "..."
```

## License

MIT
