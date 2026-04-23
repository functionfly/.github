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

## Edge State (StateFabric for Edge Functions)

The Edge State client provides low-latency state access for functions running at the edge (WASM runtime). Unlike the standard StateClient which uses HTTP API calls, EdgeStateClient uses direct WASM host function calls for optimal performance.

### Auto-Detection

The EdgeStateClient automatically detects whether it's running in a WASM environment or locally:
- **In WASM (edge)**: Uses direct host function calls
- **Locally**: Falls back to HTTP API calls

### Basic Usage

```python
from flypy import edge_state

# Initialize client (auto-detects environment)
state = edge_state.EdgeStateClient(fabric_id="my-fabric")

# Get state value
cart = state.get("carts/user123", default={"items": []})

# Set state value
state.set("carts/user123", {"items": [{"id": 1, "name": "Widget"}]})

# Delete state value
state.delete("carts/user123")

# Create snapshot
snapshot = state.snapshot("carts/user123", label="pre-checkout")
```

### Using Convenience Functions

```python
from flypy import edge_state

# Get with default
config = edge_state.get("config/app", default={"theme": "light"})

# Set value
edge_state.set("config/app", {"theme": "dark", "language": "en"})

# Delete
edge_state.delete("config/app")
```

### Decorator Pattern

```python
from flypy import edge_state

manager = edge_state.EdgeStateManager(fabric_id="my-fabric")

@manager.state("preferences", key="user_id", write=True)
def get_user_preferences(user_id: str):
    """Get user preferences, creating defaults if not exists."""
    return {
        "theme": "light",
        "notifications": True,
        "language": "en"
    }

# Usage - automatically reads/writes state
prefs = get_user_preferences("user123")
```

### Edge State Manager

For more control, use the `EdgeStateManager` class:

```python
from flypy import edge_state

manager = edge_state.EdgeStateManager(
    fabric_id="my-fabric",
    tenant_id="my-tenant"
)

# Direct operations
manager.get("carts", "user123")
manager.set("carts", "user123", {"items": []})
manager.delete("carts", "user123")
```

### Path Formats

State paths support multiple formats:
- `"key"` - Simple key (uses default tenant/fabric)
- `"fabric_id/key"` - Fabric + key
- `"tenant_id/fabric_id/key"` - Full path

### Environment Configuration

Set these environment variables for automatic configuration:

```bash
export FLYPY_STATE_FABRIC_ID="my-fabric"
export FLYPY_TENANT_ID="my-tenant"
export FLYPY_API_KEY="ffly_..."  # For local fallback mode
```

### Error Handling

```python
from flypy.edge_state import EdgeStateError, EdgeStateNotFoundError

try:
    value = state.get("carts/user123")
except EdgeStateNotFoundError:
    # State doesn't exist
    value = {"items": []}
except EdgeStateError as e:
    # Other state errors
    print(f"State error: {e}")
```

### Example: Shopping Cart at Edge

```python
from flypy import edge_state

def shopping_cart_handler(event):
    state = edge_state.EdgeStateClient()
    user_id = event["user_id"]
    action = event["action"]
    
    cart_key = f"carts/{user_id}"
    
    if action == "get":
        return state.get(cart_key, default={"items": []})
    
    elif action == "add":
        cart = state.get(cart_key, default={"items": []})
        cart["items"].append(event["item"])
        state.set(cart_key, cart)
        return cart
    
    elif action == "checkout":
        # Create pre-checkout snapshot
        cart = state.get(cart_key)
        snapshot = state.snapshot(cart_key, label=f"checkout-{user_id}")
        return {"cart": cart, "snapshot": snapshot}
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

### Edge State

| Class/Function | Description |
|----------------|-------------|
| `EdgeStateClient` | Client for edge state operations |
| `EdgeStateManager` | Manager with decorator support |
| `edge_state.get()` | Convenience function to get state |
| `edge_state.set()` | Convenience function to set state |
| `edge_state.delete()` | Convenience function to delete state |
| `edge_state.snapshot()` | Convenience function to create snapshot |
