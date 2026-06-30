---
title: Writing Functions
description: How to write and structure functions in FunctionFly
---

FunctionFly supports multiple languages and runtime environments for your serverless functions.

## Supported Runtimes

| Runtime | Version | Description |
|---------|---------|-------------|
| Python | 3.11+ | Python functions with async support |
| Node.js | 18+ | JavaScript/TypeScript functions |
| Local/WASM | Latest | WASM-based execution |

## Testing Functions

Use the [Function Playground](/guides/playground/) to test functions interactively:

- Execute with custom inputs
- View streaming responses
- Compare outputs across versions
- Debug with detailed execution history

## Function Structure

### Python

```python
# main.py
async def handler(request):
    """
    Function handler
    Args:
        request: Request object with body, headers, params
    Returns:
        dict with status, body, headers
    """
    name = request.get("body", {}).get("name", "World")
    
    return {
        "status": 200,
        "body": {"message": f"Hello, {name}!"},
        "headers": {"Content-Type": "application/json"}
    }
```

### Node.js

```javascript
// main.js
export default async function(req) {
  const name = req.body?.name || "World";
  
  return {
    status: 200,
    body: { message: `Hello, ${name}!` },
    headers: { "Content-Type": "application/json" }
  };
}
```

## Request Object

| Field | Type | Description |
|-------|------|-------------|
| `body` | object | Request body (parsed JSON) |
| `headers` | object | Request headers |
| `params` | object | URL parameters |
| `method` | string | HTTP method |
| `url` | string | Full URL |

## Response Format

```python
return {
    "status": 200,           # HTTP status code
    "body": "string|object", # Response body
    "headers": {             # Optional headers
        "Content-Type": "application/json"
    }
}
```

## Dependencies

Add dependencies using `requirements.txt` (Python) or `package.json` (Node.js):

```txt
# requirements.txt
requests>=2.31.0
numpy>=1.24.0
```

```json
// package.json
{
  "dependencies": {
    "axios": "^1.6.0"
  }
}
```

## Environment Variables

Access environment variables in your function:

```python
import os

api_key = os.environ.get("API_KEY")
```

## Working with Files

```python
# Read file from /tmp directory
with open("/tmp/data.json", "r") as f:
    data = f.read()
```

## Error Handling

```python
async def handler(request):
    try:
        # Your code here
        result = await process_data(request["body"])
        return {"status": 200, "body": result}
    except Exception as e:
        return {
            "status": 500,
            "body": {"error": str(e)}
        }
```

## Examples

See the [examples](/docs/examples) directory for more function templates:

- HTTP API handlers
- Webhook processors
- Data transformations
- AI/ML inference
