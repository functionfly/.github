# Runtime Specification

This document specifies the supported runtimes for FunctionFly functions.

## Supported Runtimes

| Runtime | Version | Language | Status |
|---------|---------|----------|--------|
| Bun | 1.x | TypeScript/JavaScript | ✅ Supported |
| Node.js | 18 | JavaScript | ✅ Supported |
| Node.js | 20 | JavaScript | ✅ Supported |
| Python | 3.11 | Python | ✅ Supported |
| Python | 3.12 | Python | ✅ Supported |
| Deno | 1.x | TypeScript/JavaScript | ✅ Supported |

## Runtime Selection

Specify the runtime in your `functionfly.jsonc`:

```jsonc
{
  "runtime": "bun"
}
```

## Bun Runtime

Bun is the recommended runtime for TypeScript functions. It provides:

- Native TypeScript support (no transpilation needed)
- Fast cold starts (~5ms)
- Native npm package support
- Built-in JSON support

### Bun Configuration

```jsonc
{
  "runtime": "bun",
  "memory": "256MB",
  "timeout": "30s"
}
```

### Bun Environment Variables

- `BUN_ENV` - Set to "production" in production

## Node.js Runtime

Node.js 20 is recommended for JavaScript functions that need compatibility with npm packages.

### Node.js Configuration

```jsonc
{
  "runtime": "node20",
  "memory": "256MB",
  "timeout": "30s"
}
```

## Deno Runtime

Deno provides secure execution with built-in TypeScript support.

### Deno Configuration

```jsonc
{
  "runtime": "deno",
  "memory": "256MB",
  "timeout": "30s"
}
```

### Deno Permissions

Deno functions run with limited permissions by default. Use the manifest to request permissions:

```jsonc
{
  "runtime": "deno",
  "permissions": {
    "net": true,
    "env": ["API_KEY"]
  }
}
```

## Python Runtime

Python 3.12 is recommended for data processing and ML functions.

### Python Configuration

```jsonc
{
  "runtime": "python3.12",
  "memory": "512MB",
  "timeout": "60s"
}
```

## Capabilities

Functions can declare capabilities they need:

```jsonc
{
  "capabilities": ["kv", "http", "scheduled"]
}
```

| Capability | Description |
|------------|-------------|
| `kv` | Key-value storage access |
| `http` | Make outbound HTTP requests |
| `scheduled` | Support cron triggers |
| `websocket` | WebSocket connections |

## Memory and Timeout

Configure resource limits:

```jsonc
{
  "memory": "256MB",
  "timeout": "30s"
}
```

Default limits:

- Memory: 128MB
- Timeout: 5s

Maximum limits:

- Memory: 2048MB
- Timeout: 300s
