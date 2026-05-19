# @functionfly/flypy

FunctionFly JavaScript/TypeScript SDK for building deterministic edge functions.

## Installation

### Using Bun (Recommended)

```bash
# Install bun if you haven't already
curl -fsSL https://bun.sh/install | bash

# Install dependencies
bun install

# Build the SDK
bun run build
```

### Using npm

```bash
npm install
npm run build
```

## Features

- ✅ **StateFabric Client**: Full TypeScript client for durable state management
- ✅ **Type Safety**: Comprehensive TypeScript types for all operations
- ✅ **Error Handling**: Specialized error classes for different failure modes
- ✅ **Convenience Methods**: Simplified state operations with StateManager
- 🔄 **Function Compilation**: Framework for deterministic function compilation (coming soon)

## Usage

```typescript
import { fly } from '@functionfly/flypy';

// Create a deterministic function
export const handler = fly.function({
  name: 'calculate-total',
  deterministic: true,
  idempotent: true,
  cache_ttl: 3600
})((event: any) => {
  const { items, tax_rate = 0.08 } = event;

  const subtotal = items.reduce((sum: number, item: any) =>
    sum + (item.price * item.quantity), 0
  );

  const tax = subtotal * tax_rate;
  const total = subtotal + tax;

  return { subtotal, tax, total };
});
```

## API Reference

### fly.function(options)

Decorator for creating FunctionFly functions.

**Options:**
- `name` (string): Function name
- `deterministic` (boolean): Whether function is deterministic (default: true)
- `idempotent` (boolean): Whether function is idempotent (default: false)
- `cache_ttl` (number): Cache TTL in seconds
- `max_execution_time` (number): Max execution time in milliseconds

## Development

### Scripts

- `bun run build` - Compile TypeScript
- `bun run test` - Run tests with Vitest
- `bun run dev` - Watch mode for development

### Testing

```bash
bun run test
```

## Requirements

- Node.js 18+ or Bun
- TypeScript 5.6+

## License

MIT