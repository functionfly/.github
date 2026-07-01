---
title: Best Practices
description: Performance, security, and reliability tips for FunctionFly functions
sidebar:
  order: 5
---

# Function Best Practices

Guidelines for writing reliable, performant, and secure functions.

## Performance

### Minimize Cold Start Time

- **Keep dependencies lean** — Only import what you need
- **Avoid heavy initialization** — Move setup code into the handler only when necessary
- **Use lightweight runtimes** — Bun and Go have the fastest cold starts
- **Consider WASM** — Rust and C compiled to WASM start in under 30ms

### Connection Pooling

Reuse database and HTTP connections across invocations:

```python
import httpx

# Initialize outside handler — reused across warm invocations
client = httpx.AsyncClient()

async def handler(request):
    response = await client.get("https://api.example.com/data")
    return {"status": 200, "body": response.json()}
```

### Caching

Use `/tmp` for ephemeral caching between warm invocations:

```python
import json, os, time

CACHE_FILE = "/tmp/cache.json"
CACHE_TTL = 300  # 5 minutes

def get_cached():
    if os.path.exists(CACHE_FILE):
        with open(CACHE_FILE) as f:
            data = json.load(f)
        if time.time() - data["ts"] < CACHE_TTL:
            return data["value"]
    return None

def set_cached(value):
    with open(CACHE_FILE, "w") as f:
        json.dump({"ts": time.time(), "value": value}, f)
```

## Security

### Validate Input

Always validate and sanitize request input:

```python
async def handler(request):
    body = request.get("body", {})

    if not isinstance(body.get("email"), str):
        return {"status": 400, "body": {"error": "Invalid email"}}

    email = body["email"].strip().lower()
    if "@" not in email:
        return {"status": 400, "body": {"error": "Invalid email format"}}
```

### Use Secrets Vault

Never hardcode secrets. Use the [Secrets Vault](/secrets-vault/) for API keys and credentials:

```python
import os

# Secrets are injected as environment variables
api_key = os.environ["API_KEY"]  # Set via Secrets Vault
```

### Limit Response Size

Avoid returning large payloads:

```python
MAX_BODY_SIZE = 1_000_000  # 1MB

async def handler(request):
    result = fetch_data()
    body = json.dumps(result)

    if len(body) > MAX_BODY_SIZE:
        return {"status": 413, "body": {"error": "Response too large"}}

    return {"status": 200, "body": result}
```

## Reliability

### Timeouts

Set appropriate timeouts for external calls:

```python
import httpx

async def handler(request):
    async with httpx.AsyncClient(timeout=10.0) as client:
        try:
            response = await client.get("https://api.example.com/data")
            return {"status": 200, "body": response.json()}
        except httpx.TimeoutException:
            return {"status": 504, "body": {"error": "Upstream timeout"}}
```

### Error Handling

Return structured errors and log details:

```python
import logging

logger = logging.getLogger(__name__)

async def handler(request):
    try:
        result = await process(request["body"])
        return {"status": 200, "body": result}
    except ValueError as e:
        logger.warning(f"Validation error: {e}")
        return {"status": 400, "body": {"error": str(e)}}
    except Exception as e:
        logger.error(f"Unexpected error: {e}", exc_info=True)
        return {"status": 500, "body": {"error": "Internal server error"}}
```

### Idempotency

Design handlers to be idempotent where possible — retries should not cause duplicate side effects:

```python
async def handler(request):
    idempotency_key = request["headers"].get("idempotency-key")

    if idempotency_key and already_processed(idempotency_key):
        return get_previous_response(idempotency_key)

    result = await do_work(request["body"])

    if idempotency_key:
        store_result(idempotency_key, result)

    return {"status": 200, "body": result}
```

## Monitoring

- **Set up alerts** — Use [Monitoring & Observability](/guides/monitoring/) for error rate and latency alerts
- **Track metrics** — Monitor invocation count, duration, and error rate
- **Use structured logging** — Log JSON for easier parsing and querying

## Next Steps

- [Monitoring & Observability](/guides/monitoring/) — Production monitoring
- [Secrets Vault](/secrets-vault/) — Managing secrets
- [Rate Limiting](/guides/rate-limiting/) — Protecting your functions
- [Function Webhooks](/function-webhooks/) — Event-driven invocations
