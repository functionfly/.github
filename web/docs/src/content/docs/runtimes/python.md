---
title: Python Runtime
description: Python runtime environment for FunctionFly functions
---

# Python Runtime

FunctionFly's Python runtime provides a robust environment for running Python-based serverless functions.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| Python 3.11 | Supported | Recommended |
| Python 3.12 | Supported | Latest |

## Function Structure

A Python function must expose a `handler` function:

```python
async def handler(request: dict) -> dict:
    """
    Main entry point for the function.
    
    Args:
        request: A dictionary containing request data
        
    Returns:
        A dictionary with status, body, and optional headers
    """
    return {
        "status": 200,
        "body": {"message": "Hello, World!"},
        "headers": {"Content-Type": "application/json"}
    }
```

## Request Object

The `request` dictionary contains:

| Field | Type | Description |
|-------|------|-------------|
| `body` | `any` | Parsed request body (JSON or raw) |
| `headers` | `dict[str, str]` | HTTP headers |
| `params` | `dict[str, str]` | URL query parameters |
| `method` | `str` | HTTP method (GET, POST, etc.) |
| `url` | `str` | Full request URL |
| `path` | `str` | URL path |

## Response Format

Return a dictionary with:

```python
return {
    "status": 200,              # Required: HTTP status code
    "body": "response data",    # Required: Any JSON-serializable data
    "headers": {                  # Optional: Response headers
        "Content-Type": "application/json"
    }
}
```

## Dependencies

### requirements.txt

Include a `requirements.txt` file:

```txt
requests>=2.31.0
pydantic>=2.0.0
httpx>=0.25.0
```

### Popular Packages

The following packages are pre-installed:

- `json`, `asyncio`, `os`, `sys` (standard library)
- `requests` - HTTP requests
- `httpx` - Async HTTP client
- `pydantic` - Data validation

## Example Functions

### HTTP API Handler

```python
from typing import Any

async def handler(request: dict) -> dict:
    method = request.get("method", "GET")
    
    if method == "GET":
        return {
            "status": 200,
            "body": {"users": ["alice", "bob"]}
        }
    elif method == "POST":
        body = request.get("body", {})
        new_user = body.get("name")
        return {
            "status": 201,
            "body": {"created": new_user}
        }
    
    return {
        "status": 405,
        "body": {"error": "Method not allowed"}
    }
```

### Webhook Processor

```python
import hmac
import hashlib
import os

async def handler(request: dict) -> dict:
    # Verify signature
    signature = request["headers"].get("X-Signature", "")
    secret = os.environ.get("WEBHOOK_SECRET", "")
    
    body = str(request.get("body", ""))
    expected = hmac.new(
        secret.encode(),
        body.encode(),
        hashlib.sha256
    ).hexdigest()
    
    if not hmac.compare_digest(signature, f"sha256={expected}"):
        return {"status": 401, "body": {"error": "Invalid signature"}}
    
    # Process webhook
    event = request.get("body", {})
    await process_event(event)
    
    return {"status": 200, "body": {"received": True}}

async def process_event(event: dict):
    # Your event processing logic
    pass
```

### Data Transformation

```python
import json

async def handler(request: dict) -> dict:
    data = request.get("body", {})
    
    # Transform data
    transformed = {
        "id": data.get("id"),
        "name": f"{data.get('first_name', '')} {data.get('last_name', '')}".strip(),
        "email": data.get("email", "").lower(),
        "timestamp": data.get("created_at")
    }
    
    return {
        "status": 200,
        "body": transformed
    }
```

## Environment Variables

Access environment variables using `os.environ`:

```python
import os

api_key = os.environ.get("API_KEY")
db_url = os.environ.get("DATABASE_URL")
debug = os.environ.get("DEBUG", "false").lower() == "true"
```

## File System

The `/tmp` directory is available for temporary file storage:

```python
import json

# Write to temp file
with open("/tmp/data.json", "w") as f:
    json.dump({"key": "value"}, f)

# Read from temp file
with open("/tmp/data.json", "r") as f:
    data = json.load(f)
```

Note: Files in `/tmp` are ephemeral and may not persist between invocations.

## Error Handling

```python
import traceback

async def handler(request: dict) -> dict:
    try:
        result = await process_request(request)
        return {
            "status": 200,
            "body": result
        }
    except ValueError as e:
        return {
            "status": 400,
            "body": {"error": str(e)}
        }
    except Exception as e:
        # Log full traceback
        print(traceback.format_exc())
        return {
            "status": 500,
            "body": {"error": "Internal server error"}
        }

async def process_request(request: dict):
    # Your processing logic
    pass
```

## Async Support

The Python runtime fully supports `async`/`await`:

```python
import asyncio
import httpx

async def handler(request: dict) -> dict:
    async with httpx.AsyncClient() as client:
        response = await client.get("https://api.example.com/data")
        data = response.json()
    
    return {
        "status": 200,
        "body": data
    }
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 2048 MB |
| CPU | 1 vCPU | 4 vCPU |

Configure in `functionfly.jsonc`:

```jsonc
{
  "limits": {
    "timeout": 60,
    "memory": 512
  }
}
```

## Cold Start

Python functions may experience cold starts:
- First invocation after deployment: ~500-1000ms
- Subsequent invocations: ~10-50ms
- Keep functions warm with scheduled invocations

## Best Practices

1. **Use async I/O** for network requests
2. **Import only what you need** to reduce cold start time
3. **Use environment variables** for configuration
4. **Handle exceptions** gracefully
5. **Keep functions small** and focused
6. **Cache expensive operations** where possible
