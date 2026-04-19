---
title: JavaScript/TypeScript SDK
description: JavaScript and TypeScript SDK for FunctionFly
---

# JavaScript/TypeScript SDK

The FunctionFly JavaScript SDK provides a convenient way to interact with the FunctionFly API from Node.js and browser applications.

## Installation

```bash
npm install @functionfly/sdk
# or
yarn add @functionfly/sdk
# or
bun add @functionfly/sdk
```

## Quick Start

```typescript
import { Client } from "@functionfly/sdk";

// Initialize client
const client = new Client({
  apiKey: "your-api-key"
});

// List functions
const functions = await client.functions.list();

// Invoke a function
const result = await client.functions.invoke("my-function", {
  name: "World"
});
console.log(result); // { message: "Hello, World!" }
```

## Authentication

```typescript
import { Client } from "@functionfly/sdk";

// Using API key
const client = new Client({ apiKey: "ffly_..." });

// Using environment variable (Node.js)
const client = new Client({
  apiKey: process.env.FFLY_API_KEY!
});
```

## TypeScript Support

The SDK is written in TypeScript and includes type definitions.

```typescript
import { Client, Function, InvocationResult } from "@functionfly/sdk";

// Typed invocation
interface RequestData {
  name: string;
}

interface ResponseData {
  message: string;
}

const result = await client.functions.invoke<RequestData, ResponseData>(
  "my-function",
  { name: "World" }
);
// result is typed as ResponseData
```

## Managing Functions

### Create a Function

```typescript
// Create from directory
const function = await client.functions.create({
  name: "my-api",
  runtime: "nodejs",
  directory: "./my-function"
});

// Deploy
await client.functions.deploy(function.id);
```

### List Functions

```typescript
// List all functions
const functions = await client.functions.list();

// List with pagination
const functions = await client.functions.list({
  limit: 10,
  offset: 0
});

// Filter by runtime
const nodeFunctions = await client.functions.list({
  runtime: "nodejs"
});
```

### Get Function Details

```typescript
const func = await client.functions.get("my-function");
console.log(func.name);
console.log(func.runtime);
console.log(func.version);
```

### Update a Function

```typescript
await client.functions.update("my-function", {
  description: "Updated description",
  environment: { DEBUG: "true" }
});
```

### Delete a Function

```typescript
await client.functions.delete("my-function");
```

## Invoking Functions

### Basic Invocation

```typescript
const result = await client.functions.invoke("my-function", {
  key: "value"
});
```

### With Custom Headers

```typescript
const result = await client.functions.invoke(
  "my-function",
  { key: "value" },
  {
    headers: {
      Authorization: "Bearer token",
      "X-Custom-Header": "value"
    }
  }
);
```

## Environment Variables & Secrets

```typescript
// Set environment variables
await client.functions.setEnv("my-function", {
  API_URL: "https://api.example.com"
});

// Set secrets
await client.functions.setSecrets("my-function", {
  API_KEY: "secret-value"
});

// Get environment variables
const env = await client.functions.getEnv("my-function");
```

## Monitoring

```typescript
// Get execution logs
const logs = await client.functions.logs("my-function", { limit: 100 });

// Get metrics
const metrics = await client.functions.metrics("my-function");
console.log(metrics.invocations);
console.log(metrics.errors);
console.log(metrics.averageDuration);

// Check health
const health = await client.functions.health("my-function");
console.log(health.status); // "healthy" or "unhealthy"
```

## Error Handling

```typescript
import { FunctionFlyError, NotFoundError, AuthenticationError } from "@functionfly/sdk";

try {
  const result = await client.functions.invoke("nonexistent-function", {});
} catch (error) {
  if (error instanceof NotFoundError) {
    console.log("Function not found");
  } else if (error instanceof AuthenticationError) {
    console.log("Invalid API key");
  } else if (error instanceof FunctionFlyError) {
    console.log(`Error: ${error.message}`);
  }
}
```

## Configuration

```typescript
import { Client } from "@functionfly/sdk";

const client = new Client({
  apiKey: "ffly_...",
  baseURL: "https://api.functionfly.com",
  timeout: 30000,
  retries: 3
});
```

## Browser Usage

The SDK works in browser environments with some limitations:

```html
<script type="module">
  import { Client } from "@functionfly/sdk";
  
  const client = new Client({
    apiKey: "your-api-key" // Use read-only key for browser
  });
  
  const result = await client.functions.invoke("public-function", {
    name: "World"
  });
</script>
```

### CORS Considerations

When using the SDK in browsers:
- Use read-only API keys for public functions
- Configure CORS on your functions for browser access
- Admin operations require Node.js environment

## Advanced Usage

### Batch Invocations

```typescript
// Invoke multiple functions concurrently
const results = await Promise.all([
  client.functions.invoke("func1", { id: 1 }),
  client.functions.invoke("func2", { id: 2 }),
  client.functions.invoke("func3", { id: 3 })
]);
```

### With AbortController

```typescript
const controller = new AbortController();

// Cancel after 5 seconds
setTimeout(() => controller.abort(), 5000);

try {
  const result = await client.functions.invoke(
    "my-function",
    { key: "value" },
    { signal: controller.signal }
  );
} catch (error) {
  if (error.name === "AbortError") {
    console.log("Request was aborted");
  }
}
```

### Streaming (Node.js only)

```typescript
for await (const chunk of client.functions.invokeStream(
  "streaming-function",
  {}
)) {
  console.log(chunk);
}
```

## React/Next.js Integration

```tsx
import { useState, useEffect } from "react";
import { Client } from "@functionfly/sdk";

const client = new Client({
  apiKey: process.env.NEXT_PUBLIC_FLY_API_KEY!
});

export function useFunction(name: string) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const invoke = async (payload: any) => {
    setLoading(true);
    try {
      const result = await client.functions.invoke(name, payload);
      setData(result);
      return result;
    } catch (err) {
      setError(err);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { invoke, data, loading, error };
}
```

## API Reference

### Client Configuration

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `apiKey` | `string` | Required | Your API key |
| `baseURL` | `string` | `https://api.functionfly.com` | API base URL |
| `timeout` | `number` | `30000` | Request timeout (ms) |
| `retries` | `number` | `3` | Number of retries |

### Functions API

| Method | Description |
|--------|-------------|
| `list(options?)` | List all functions |
| `get(name)` | Get function details |
| `create(options)` | Create a new function |
| `update(name, options)` | Update function |
| `delete(name)` | Delete function |
| `invoke(name, data, options?)` | Invoke function |
| `deploy(name)` | Deploy function |
| `logs(name, options?)` | Get execution logs |
| `metrics(name)` | Get function metrics |
| `health(name)` | Check function health |
