# @functionfly/edge-sdk

TypeScript SDK for FunctionFly State Fabric - optimized for edge runtimes (Cloudflare Workers, Vercel Edge, Deno Deploy, etc.)

## Installation

```bash
npm install @functionfly/edge-sdk
```

## Quick Start

```typescript
import { createClient, state } from '@functionfly/edge-sdk';

const client = createClient({
  apiKey: process.env.FUNCTIONFLY_API_KEY!,
  tenantId: process.env.FUNCTIONFLY_TENANT_ID,
});

// Or use the state helper for cleaner API
const cart = state(client, 'acme/cart');
await cart.set({ items: [], total: 0 });
```

## Usage Examples

### Basic State Operations

```typescript
import { createClient, state } from '@functionfly/edge-sdk';

const client = createClient({ apiKey: 'your-api-key' });

// Get state value
const user = await client.getValue('acme/users', 'user-123');
console.log(user.value);

// Set state value
await client.setValue('acme/users', { name: 'John', email: 'john@example.com' }, 'user-123');

// Delete state value
await client.deleteValue('acme/users', 'user-123');
```

### Using the State Helper

```typescript
import { state } from '@functionfly/edge-sdk';

const cart = state(client, 'acme/cart');

// Add item to cart
await cart.set({ items: [{ id: 'prod-1', qty: 1 }] }, 'user-123');

// Get cart
const cartData = await cart.get('user-123');

// Update cart
await cart.patch([
  { op: 'add', path: '/items/0/qty', value: 2 }
], 'user-123');

// Cart history
const history = await cart.history('user-123');
```

### Transactions

```typescript
// Get state, modify, and set atomically
const current = await client.getValue('acme/inventory', 'product-456');
const updated = { ...current.value, stock: current.value.stock - 1 };
await client.setValue('acme/inventory', updated, 'product-456');
```

### Time Travel

```typescript
// Query state at a specific point in time
const oldState = await client.timeTravel(
  'acme/users',
  '2024-01-15T10:00:00Z',
  'user-123'
);
console.log(oldState.data);
```

### Snapshots

```typescript
// Create a snapshot before a major operation
await client.createSnapshot('acme/database', 'before-migration');

// List snapshots
const snapshots = await client.listSnapshots('acme/database');

// Restore from snapshot
await client.restoreSnapshot('acme/database', 42);
```

### Permissions

```typescript
// Grant read access to a team
await client.grantPermission('acme/sensitive-data', {
  principal_type: 'team',
  principal_id: 'team-123',
  can_read: true,
  can_write: false,
  can_delete: false,
});

// List permissions
const permissions = await client.getPermissions('acme/sensitive-data');
```

### Agent Memory

```typescript
// Store memory with embedding
await client.createMemory({
  agent_id: 'assistant-001',
  memory_type: 'longterm',
  content: 'User prefers dark mode and email notifications',
  embedding: [0.1, 0.2, 0.3, /* ... */],
  importance_score: 0.8,
  ttl_days: 90,
});

// Search similar memories
const similar = await client.searchMemories({
  agent_id: 'assistant-001',
  memory_type: 'longterm',
  embedding: [0.1, 0.2, 0.3, /* ... */],
  limit: 5,
  threshold: 0.7,
});
```

### Triggers

```typescript
// Create a trigger on state changes
await client.createTrigger({
  state_path: 'acme/orders',
  trigger_type: 'on_write',
  key_pattern: 'order/*',
  target_function: 'acme/order-processor:v1',
  include_previous: true,
  include_new: true,
  is_active: true,
});
```

## Configuration

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `apiKey` | string | Yes | Your FunctionFly API key |
| `baseUrl` | string | No | API base URL (defaults to `https://api.functionfly.com`) |
| `tenantId` | string | No | Tenant ID for multi-tenant setups |

## Edge Runtime Compatibility

This SDK is designed to work seamlessly in edge environments:

- **Cloudflare Workers** - Uses native `fetch`
- **Vercel Edge** - Uses Edge Runtime compatible APIs
- **Deno Deploy** - Uses Deno's built-in fetch
- **Bun** - Uses Bun's fetch implementation

## API Reference

### Client Methods

- `listStates(limit, offset)` - List all states
- `createState(state)` - Create a new state
- `getState(path)` - Get state metadata
- `updateState(path, updates)` - Update state
- `deleteState(path)` - Delete a state
- `getValue(path, key)` - Get state value
- `setValue(path, value, key, ttlDays)` - Set state value
- `patchValue(path, patch, key)` - Patch state value
- `deleteValue(path, key)` - Delete state value
- `getHistory(path, key, limit, offset)` - Get event history
- `createSnapshot(path, label)` - Create snapshot
- `listSnapshots(path, limit, offset)` - List snapshots
- `restoreSnapshot(path, version)` - Restore snapshot
- `timeTravel(path, timestamp, key)` - Query historical state
- `getPermissions(path)` - List permissions
- `grantPermission(path, permission)` - Grant permission
- `listTriggers(statePath, limit, offset)` - List triggers
- `createTrigger(trigger)` - Create trigger
- `deleteTrigger(id)` - Delete trigger
- `createMemory(memory)` - Create memory
- `getMemory(id)` - Get memory
- `updateMemory(id, updates)` - Update memory
- `deleteMemory(id)` - Delete memory
- `listMemories(agentId, type, limit, offset)` - List memories
- `searchMemories(request)` - Search memories

## TypeScript Support

This SDK is written in TypeScript and provides full type definitions out of the box.

```typescript
import type { State, StateValue, AgentMemory } from '@functionfly/edge-sdk';
```

## License

MIT
