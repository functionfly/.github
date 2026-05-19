---
title: StateFabric & Edge State
description: Durable state management for stateless functions at the edge
---

StateFabric provides durable, distributed state management for your serverless functions. It enables stateful patterns like shopping carts, session management, and user preferences while maintaining the scalability benefits of stateless architectures.

## Overview

StateFabric consists of two main components:

1. **StateFabric (Backend)**: Full-featured state management via HTTP API with rich querying, history, and snapshots
2. **Edge State**: High-performance state access for functions running at the edge (WASM runtime) via direct host function calls

## Key Features

- **Durable Storage**: Automatic persistence with replication
- **Snapshots**: Point-in-time state backups
- **History Tracking**: Full audit trail of state changes
- **Time Travel**: Query state at any point in time
- **Permissions**: Fine-grained access control
- **Edge Optimized**: Low-latency access for edge functions

## Use Cases

- **Shopping Carts**: Maintain cart state across page views
- **User Sessions**: Track user activity and preferences
- **Rate Limiting**: Distributed request throttling
- **Configuration**: Dynamic application settings
- **Counters & Analytics**: Distributed counting and metrics
- **Workflow State**: Multi-step process coordination

## Creating a StateFabric

### Using the Dashboard

1. Navigate to **StateFabric** in the sidebar
2. Click **Create Fabric**
3. Choose a name and region
4. Select storage type (durable, ephemeral, or cached)

### Using the API

```bash
curl -X POST https://api.functionfly.com/v1/state-fabrics \
  -H "Authorization: Bearer $FFLY_API_KEY" \
  -d '{
    "name": "my-app-state",
    "type": "custom",
    "settings": {
      "storage_type": "durable",
      "region": "us-east-1"
    }
  }'
```

## Using State from Edge Functions

Edge functions running in the WASM runtime can access StateFabric through the Edge State SDK, which provides low-latency state operations via direct host function calls.

### Python SDK

```python
from flypy import edge_state

# Auto-detects WASM environment and uses host functions
state = edge_state.EdgeStateClient(fabric_id="my-fabric")

# Get value with default
cart = state.get("carts/user123", default={"items": []})

# Set value
state.set("carts/user123", {
    "items": [{"id": 1, "name": "Widget", "price": 9.99}],
    "total": 9.99
})

# Delete value
state.delete("carts/user123")

# Create snapshot for backup
snapshot = state.snapshot("carts/user123", label="pre-checkout")
```

### Decorator Pattern

```python
from flypy import edge_state

manager = edge_state.EdgeStateManager(fabric_id="my-fabric")

@manager.state("preferences", key="user_id", write=True)
def get_preferences(user_id: str):
    """Auto-persisted state with decorator."""
    return {
        "theme": "light",
        "notifications": True,
        "language": "en"
    }

# Usage automatically reads/writes state
prefs = get_preferences("user123")  # Gets or creates
```

### JavaScript/TypeScript SDK

```typescript
import { edgeState } from '@functionfly/sdk';

const state = edgeState.createClient({ fabricId: 'my-fabric' });

// Get value
const cart = await state.get('carts/user123', { default: { items: [] } });

// Set value
await state.set('carts/user123', { items: [...] });

// Delete
await state.delete('carts/user123');
```

## Path Structure

State paths follow a hierarchical structure:

```
{tenant_id}/{fabric_id}/{key}
```

Examples:
- `my-tenant/my-fabric/carts/user123`
- `my-fabric/carts/user123` (uses default tenant)
- `carts/user123` (uses default tenant and fabric)

## State Operations

### Reading State

```python
# Simple get
value = state.get("key")

# With default
value = state.get("key", default={})

# Get all values with prefix
all_carts = state.get_all("carts/")
```

### Writing State

```python
# Set value
state.set("key", {"data": "value"})

# Set with metadata
state.set("key", value, metadata={
    "version": "1.0",
    "source": "checkout-flow"
})
```

### Deleting State

```python
# Delete specific key
state.delete("carts/user123")

# Delete all with prefix (use carefully)
state.delete_all("carts/temp-")
```

## Snapshots

Snapshots capture a point-in-time view of state:

```python
# Create snapshot
snapshot = state.snapshot("carts/user123", label="before-checkout")

# List snapshots for a key
snapshots = state.list_snapshots("carts/user123")

# Restore from snapshot
state.restore("carts/user123", snapshot_id="snap_123")
```

## History & Time Travel

Query state history and restore previous versions:

```python
# Get history
history = state.get_history("carts/user123", limit=50)

# Time travel - get state at specific time
past_state = state.time_travel(
    "carts/user123",
    timestamp="2024-01-15T10:30:00Z"
)

# Get specific version
version_5 = state.time_travel("carts/user123", version=5)
```

## Permissions

Control access to state:

```python
# Grant read access
state.grant_permission(
    path="carts/user123",
    principal_type="user",
    principal_id="user_456",
    can_read=True
)

# Grant write access
state.grant_permission(
    path="carts/user123",
    principal_type="service",
    principal_id="checkout-service",
    can_read=True,
    can_write=True
)
```

## Triggers

Execute functions on state changes:

```python
# Create trigger
state.create_trigger(
    state_path="carts/*",
    trigger_type="on_set",
    target_function="update-inventory",
    key_pattern="carts/*"
)

# Trigger on specific conditions
state.create_trigger(
    state_path="carts/*",
    trigger_type="on_change",
    target_function="send-notification",
    condition={
        "field": "total",
        "operator": "gt",
        "value": 100
    }
)
```

## Best Practices

### Design Patterns

1. **Namespace Your Keys**: Use prefixes like `carts/`, `sessions/`, `config/`
2. **Keep Values Small**: Under 1MB for best performance
3. **Use Appropriate Storage**:
   - `durable`: For critical data (carts, orders)
   - `ephemeral`: For temporary data (sessions, cache)
   - `cached`: For read-heavy data (config, catalogs)

### Error Handling

```python
from flypy.edge_state import EdgeStateNotFoundError, EdgeStatePermissionError

try:
    value = state.get("carts/user123")
except EdgeStateNotFoundError:
    # Initialize default
    value = {"items": []}
    state.set("carts/user123", value)
except EdgeStatePermissionError:
    # Handle permission error
    return {"error": "Access denied"}
```

### Performance Tips

1. **Use Edge State for Edge Functions**: Lower latency than HTTP API
2. **Batch Operations**: Group multiple reads/writes when possible
3. **Cache Locally**: Store frequently accessed data in function memory
4. **Use Snapshots Wisely**: Don't create snapshots on every write

## Storage Types

| Type | Durability | Latency | Use Case |
|------|-----------|---------|----------|
| `durable` | High | Medium | Carts, orders, user data |
| `ephemeral` | Low | Low | Sessions, temp data |
| `cached` | Medium | Very Low | Config, catalogs, read-heavy |

## Pricing

StateFabric usage is billed based on:
- **Storage**: GB-months of stored data
- **Operations**: Number of read/write operations
- **Snapshots**: Storage for snapshots
- **Transfer**: Data transfer between regions

See [Pricing](/pricing/) for detailed rates.

## Limits

| Limit | Value |
|-------|-------|
| Max key length | 256 characters |
| Max value size | 1 MB |
| Max fabric count | Per plan limits |
| Max stores per fabric | 10 |
| Snapshot retention | 30 days |
| History retention | 90 days |

## Troubleshooting

### Common Issues

**State not persisting**
- Check fabric status is `active`
- Verify storage type is not `ephemeral` (if durability needed)
- Check permissions

**High latency**
- Use Edge State for edge functions
- Check region matches function deployment
- Consider `cached` storage type

**Permission denied**
- Verify API key has state access
- Check state permissions
- Ensure fabric belongs to correct tenant

### Debug Mode

```python
# Enable debug logging
import logging
logging.basicConfig(level=logging.DEBUG)

# Test connection
state = edge_state.EdgeStateClient()
info = state.get_fabric_info()
print(f"Connected to: {info}")
```

## Migration from External State

### From Redis

```python
# Before (Redis)
import redis
r = redis.Redis()
cart = r.get("cart:user123")

# After (StateFabric)
from flypy import edge_state
state = edge_state.EdgeStateClient()
cart = state.get("carts/user123")
```

### From Database

```python
# Before (Database)
cart = db.query("SELECT * FROM carts WHERE user_id = %s", user_id)

# After (StateFabric)
cart = state.get(f"carts/{user_id}")
```

## API Reference

See the [REST API documentation](/api-reference/) for complete StateFabric endpoints.

## Next Steps

- [Python SDK](/sdks/python/) - Detailed SDK reference
- [JavaScript SDK](/sdks/javascript/) - JavaScript/TypeScript SDK
- [REST API](/api-reference/) - Direct API access
- [Pricing](/pricing/) - Pricing details