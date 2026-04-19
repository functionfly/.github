---
title: Python SDK
description: Python SDK for FunctionFly
---

# Python SDK

The FunctionFly Python SDK provides a convenient way to interact with the FunctionFly API from Python applications.

## Installation

```bash
pip install functionfly
```

## Quick Start

```python
from functionfly import Client

# Initialize client
client = Client(api_key="your-api-key")

# List functions
functions = client.functions.list()

# Invoke a function
result = client.functions.invoke(
    "my-function",
    {"name": "World"}
)
print(result)  # {"message": "Hello, World!"}
```

## Authentication

```python
from functionfly import Client

# Using API key
client = Client(api_key="ffly_...")

# Using environment variable
import os
client = Client(api_key=os.environ.get("FFLY_API_KEY"))
```

## Managing Functions

### Create a Function

```python
# Create from directory
function = client.functions.create(
    name="my-api",
    runtime="python",
    directory="./my-function"
)

# Deploy
client.functions.deploy(function.id)
```

### List Functions

```python
# List all functions
functions = client.functions.list()

# List with pagination
functions = client.functions.list(limit=10, offset=0)

# Filter by runtime
python_functions = client.functions.list(runtime="python")
```

### Get Function Details

```python
function = client.functions.get("my-function")
print(function.name)
print(function.runtime)
print(function.version)
```

### Update a Function

```python
client.functions.update(
    "my-function",
    description="Updated description",
    environment={"DEBUG": "true"}
)
```

### Delete a Function

```python
client.functions.delete("my-function")
```

## Invoking Functions

### Synchronous Invocation

```python
result = client.functions.invoke(
    "my-function",
    {"key": "value"}
)
```

### Asynchronous Invocation

```python
import asyncio
from functionfly import AsyncClient

async def main():
    client = AsyncClient(api_key="ffly_...")
    result = await client.functions.invoke(
        "my-function",
        {"key": "value"}
    )
    print(result)

asyncio.run(main())
```

### With Custom Headers

```python
result = client.functions.invoke(
    "my-function",
    {"key": "value"},
    headers={
        "Authorization": "Bearer token",
        "X-Custom-Header": "value"
    }
)
```

## Environment Variables & Secrets

```python
# Set environment variables
client.functions.set_env(
    "my-function",
    {"API_URL": "https://api.example.com"}
)

# Set secrets
client.functions.set_secrets(
    "my-function",
    {"API_KEY": "secret-value"}
)

# Get environment variables
env = client.functions.get_env("my-function")
```

## Monitoring

```python
# Get execution logs
logs = client.functions.logs("my-function", limit=100)

# Get metrics
metrics = client.functions.metrics("my-function")
print(metrics.invocations)
print(metrics.errors)
print(metrics.average_duration)

# Check health
health = client.functions.health("my-function")
print(health.status)  # "healthy" or "unhealthy"
```

## Error Handling

```python
from functionfly import FunctionFlyError, NotFoundError, AuthenticationError

try:
    result = client.functions.invoke("nonexistent-function", {})
except NotFoundError:
    print("Function not found")
except AuthenticationError:
    print("Invalid API key")
except FunctionFlyError as e:
    print(f"Error: {e.message}")
```

## Configuration

```python
from functionfly import Client

client = Client(
    api_key="ffly_...",
    base_url="https://api.functionfly.com",
    timeout=30,
    retries=3
)
```

## Advanced Usage

### Batch Invocations

```python
# Invoke multiple functions concurrently
import asyncio
from functionfly import AsyncClient

async def batch_invoke():
    client = AsyncClient(api_key="ffly_...")
    
    tasks = [
        client.functions.invoke("func1", {"id": 1}),
        client.functions.invoke("func2", {"id": 2}),
        client.functions.invoke("func3", {"id": 3}),
    ]
    
    results = await asyncio.gather(*tasks, return_exceptions=True)
    return results

asyncio.run(batch_invoke())
```

### Streaming Results

```python
for chunk in client.functions.invoke_stream("streaming-function", {}):
    print(chunk)
```

## API Reference

### Client

| Method | Description |
|--------|-------------|
| `functions.list()` | List all functions |
| `functions.get(name)` | Get function details |
| `functions.create(...)` | Create a new function |
| `functions.update(name, ...)` | Update function |
| `functions.delete(name)` | Delete function |
| `functions.invoke(name, data)` | Invoke function |
| `functions.deploy(name)` | Deploy function |
| `functions.logs(name)` | Get execution logs |
| `functions.metrics(name)` | Get function metrics |
| `functions.health(name)` | Check function health |
