---
title: Functions
description: Serverless functions on FunctionFly
---

# Functions

FunctionFly Functions are serverless compute units that run on a global edge network. Write in your language of choice and deploy in seconds.

## Key Features

- **Multi-language support** — Python, JavaScript, TypeScript, Go, Rust, Ruby, Kotlin, and more
- **Edge execution** — Deploy to 300+ locations worldwide
- **Auto-scaling** — Scale from zero to thousands of concurrent invocations
- **Sub-millisecond cold starts** — Optimized runtimes with WebAssembly support
- **Built-in observability** — Logs, metrics, and tracing out of the box

## Supported Runtimes

| Runtime | Version | Type |
|---------|---------|------|
| Python | 3.11+ | Interpreted |
| JavaScript / Node.js | 18+ | Interpreted |
| TypeScript | 5+ | Compiled |
| Bun | Latest | Interpreted |
| Deno | Latest | Interpreted |
| Go | 1.21+ | Compiled |
| Rust | Latest | WASM |
| Ruby | 3+ | Interpreted |
| Kotlin | Latest | WASM |
| Swift | Latest | WASM |
| C/C++ | Latest | WASM |

## Quick Example

```python
async def handler(request):
    name = request.get("body", {}).get("name", "World")
    return {
        "status": 200,
        "body": {"message": f"Hello, {name}!"},
        "headers": {"Content-Type": "application/json"}
    }
```

## How It Works

1. **Write** your function in any supported language
2. **Deploy** via CLI, dashboard, or API
3. **Invoke** via HTTP, webhooks, or scheduled triggers
4. **Monitor** with built-in analytics and logging

## Next Steps

- [Creating Functions](/functions/creating/) — Write and deploy your first function
- [Function Structure](/functions/structure/) — Request/response format and lifecycle
- [Testing Functions](/functions/testing/) — Local dev and the Playground
- [Best Practices](/functions/best-practices/) — Performance, security, and reliability tips
- [CLI Reference](/cli/) — Full CLI documentation
