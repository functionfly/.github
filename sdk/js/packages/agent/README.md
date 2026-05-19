# FunctionFly Agent SDK

Core SDK for AI agent integrations with FunctionFly function registry.

## Features

- **Function Discovery**: Search and filter functions by category, trust score, and metadata
- **Trust-Aware Routing**: Prefer higher trust score functions for critical tasks
- **Automatic Retry with Fallback**: If a function fails, automatically try the next trusted alternative
- **Execution with Context**: Execute functions with optional context and timeout control

## Installation

```bash
npm install @functionfly/agent
# or
yarn add @functionfly/agent
# or
bun add @functionfly/agent
```

## Quick Start

```typescript
import { AgentClient } from '@functionfly/agent';

const client = new AgentClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.functionfly.com', // optional, defaults to production
});

// Discover functions
const functions = await client.discoverFunctions({
  category: 'data-processing',
  minTrustScore: 80,
});

// Execute with automatic retry
const result = await client.execute('function-id', { input: { data: 'test' } });

// Execute with fallback
const resultWithFallback = await client.executeWithRetry('function-id', { data: 'test' }, {
  enableFallback: true,
  minTrustScore: 70,
});
```

## API Reference

### AgentClient

#### Constructor

```typescript
const client = new AgentClient({
  apiKey: string;
  baseUrl?: string;      // defaults to 'https://api.functionfly.com'
  appKey?: string;       // optional app-scoped key
  timeout?: number;      // request timeout in ms, default 30000
  retryConfig?: Partial<RetryConfig>;
});
```

#### Discovery Methods

```typescript
// Search functions
await client.searchFunctions({
  query?: string;
  category?: string;
  runtime?: string;
  minTrustScore?: number;
  limit?: number;
  offset?: number;
  author?: string;
  tags?: string[];
});

// Discover functions for agent task
await client.discoverFunctions({
  task?: string;
  category?: string;
  minTrustScore?: number;
  limit?: number;
});

// Get function by author/name
await client.getFunction('author', 'function-name');

// Get function by ID
await client.getFunctionById('function-id');

// Get trust score
await client.getTrustScore('function-id');

// List by category
await client.listByCategory('data-processing', 20, 0);

// List by author
await client.listByAuthor('functionfly', 20, 0);

// Find similar functions
await client.findSimilar('author', 'function-name', 5);
```

#### Execution Methods

```typescript
// Execute function
await client.execute('function-id', { input: { data: 'test' } });

// Execute by author/name
await client.executeByName('author', 'name', { input: { data: 'test' } });

// Execute with retry and fallback
await client.executeWithRetry('function-id', { data: 'test' }, {
  enableFallback: boolean;
  minTrustScore?: number;
  preferredTrustTier?: TrustTier;
  retry?: Partial<RetryConfig>;
});
```

### Trust Tiers

Functions are assigned trust tiers based on their trust scores:

- `highly_trusted`: Score ≥ 90 AND verified
- `verified`: Score ≥ 70
- `trusted`: Score ≥ 50
- `untrusted`: Score < 50

### Retry Configuration

```typescript
interface RetryConfig {
  maxRetries: number;        // default: 3
  baseDelayMs: number;       // default: 1000
  maxDelayMs: number;         // default: 10000
  backoffMultiplier: number; // default: 2
  retryOnTimeout: boolean;    // default: true
  retryOnError: boolean;      // default: true
}
```

## Error Handling

The SDK provides specific error types:

```typescript
import { 
  AgentSDKError,
  FunctionNotFoundError,
  TrustScoreTooLowError,
  ExecutionWithFallbackError 
} from '@functionfly/agent';

try {
  await client.executeWithRetry('function-id', input, options);
} catch (error) {
  if (error instanceof FunctionNotFoundError) {
    // Handle function not found
  } else if (error instanceof TrustScoreTooLowError) {
    // Handle trust score below minimum
  } else if (error instanceof ExecutionWithFallbackError) {
    // All fallbacks exhausted, access attempted functions
    console.log(error.attemptedFunctions);
  }
}
```

## Building Tools for Agent Frameworks

```typescript
// Build a tool from a function
const tool = client.buildTool(functionInfo);

// Build multiple tools
const tools = await client.buildTools({
  category: 'data-processing',
  minTrustScore: 70,
});
```

## License

MIT
