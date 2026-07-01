---
title: Function Structure
description: Request and response format, lifecycle, and execution model
sidebar:
  order: 3
---

# Function Structure

Every FunctionFly function receives a request object and returns a response object.

## Request Object

| Field | Type | Description |
|-------|------|-------------|
| `body` | `object` | Parsed request body (JSON) |
| `headers` | `object` | Request headers (lowercase keys) |
| `params` | `object` | URL path and query parameters |
| `method` | `string` | HTTP method (`GET`, `POST`, etc.) |
| `url` | `string` | Full request URL |
| `env` | `object` | Environment variables (non-secret) |

### Example Access

```python
async def handler(request):
    method = request["method"]
    body = request.get("body", {})
    user_agent = request["headers"].get("user-agent", "")
    page = request["params"].get("page", "1")
```

## Response Format

```python
return {
    "status": 200,              # HTTP status code
    "body": {"key": "value"},   # string or dict (auto-serialized to JSON)
    "headers": {                # optional response headers
        "Content-Type": "application/json",
        "X-Request-Id": "abc123"
    }
}
```

### Status Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request |
| `401` | Unauthorized |
| `403` | Forbidden |
| `404` | Not Found |
| `500` | Internal Server Error |

## Execution Model

### Cold Starts

A cold start occurs when a new instance is created to handle a request:

1. Runtime is initialized
2. Your function code is loaded
3. The handler is invoked

Cold start times vary by runtime:

| Runtime | Typical Cold Start |
|---------|--------------------|
| Python | 50–200ms |
| Node.js | 30–100ms |
| Bun | 20–80ms |
| Go | 10–50ms |
| Rust (WASM) | 5–30ms |

### Warm Invocations

Subsequent requests to the same instance reuse the already-loaded runtime, avoiding cold start overhead.

### Concurrency

Each function instance handles one request at a time. FunctionFly automatically scales instances based on traffic.

## File System

Functions have access to a writable `/tmp` directory:

```python
import os

# Write to /tmp
with open("/tmp/cache.json", "w") as f:
    f.write('{"cached": true}')

# Read from /tmp
with open("/tmp/cache.json", "r") as f:
    data = f.read()
```

Files in `/tmp` persist across warm invocations but are not guaranteed across cold starts.

## Dependencies

### Python

Create a `requirements.txt` in your function directory:

```txt
requests>=2.31.0
numpy>=1.24.0
```

### Node.js / Bun

Create a `package.json` in your function directory:

```json
{
  "dependencies": {
    "axios": "^1.6.0",
    "lodash": "^4.17.21"
  }
}
```

### Go

Use standard Go modules (`go.mod`) in your function directory.

## Next Steps

- [Testing Functions](/functions/testing/) — Local dev and Playground
- [Best Practices](/functions/best-practices/) — Performance and reliability
- [Environment Variables](/guides/environment-variables/) — Configuration
- [Error Handling](/guides/error-codes/) — Error codes reference
