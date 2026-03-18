# TypeScript Getting Started Guide

This guide will help you write TypeScript functions for the FunctionFly platform.

## Overview

FunctionFly supports TypeScript functions running on Bun, Node.js, and Deno runtimes. TypeScript functions provide:
- **Type safety** with full TypeScript type definitions
- **IDE support** with autocomplete and inline documentation
- **Modern JavaScript** features (async/await, modules, etc.)
- **Native performance** with Bun runtime

## Quick Start

### 1. Create a Basic Function

Create a new directory for your function and add the manifest file:

```jsonc
// functionfly.jsonc
{
  "name": "my-function",
  "version": "1.0.0",
  "runtime": "bun",
  "entry": "main.ts",
  "handler": "handler",
  "memory": "128MB",
  "timeout": "5s"
}
```

### 2. Write Your Function

```typescript
// main.ts
import type { Handler, Request, Env, Context, Response } from "./types/functionfly";

const handler: Handler = async (
  request: Request,
  env: Env,
  context: Context
): Promise<Response> => {
  // Parse request body
  const body = await request.json<{ name?: string }>();
  
  // Access environment variables
  const greeting = env.GREETING ?? "Hello";
  
  // Return response
  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: {
      message: `${greeting}, ${body.name ?? "World"}!`
    }
  };
};

export { handler };
export default handler;
```

### 3. Add TypeScript Configuration

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noEmit": true,
    "skipLibCheck": true
  }
}
```

### 4. Add Type Definitions

Create a `types/functionfly.d.ts` file in your function directory:

```typescript
interface Request {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: any;
  json: <T = any>() => Promise<T>;
  text: () => Promise<string>;
  formData: () => Promise<FormData>;
}

interface Response {
  status: number;
  headers: Record<string, string>;
  body: string | object;
}

interface KVNamespace {
  get(key: string): Promise<string | null>;
  put(key: string, value: string, options?: { expirationTtl?: number }): Promise<void>;
  delete(key: string): Promise<void>;
  list(options?: { prefix?: string; limit?: number }): Promise<{ keys: { name: string }[] }>;
}

interface Env {
  [key: string]: string;
}

interface Context {
  request: Request;
  env: Env;
  kv: KVNamespace;
  waitUntil(promise: Promise<any>): void;
  next(): Promise<Response>;
}

type Handler = (
  request: Request,
  env: Env,
  context: Context
) => Promise<Response> | Response;
```

Or reference the central types from `web/dashboard/src/types/functionfly.d.ts`.

## Type Definitions

### Request

The `Request` object contains all information about the incoming HTTP request:

```typescript
interface Request {
  method: string;      // GET, POST, PUT, DELETE, etc.
  url: string;         // Full URL with query parameters
  headers: Record<string, string>;
  body: any;
  
  // Parse body as JSON
  json: <T>() => Promise<T>;
  
  // Get body as plain text
  text: () => Promise<string>;
  
  // Parse body as FormData
  formData: () => Promise<FormData>;
}
```

### Response

The `Response` object returned by your function:

```typescript
interface Response {
  status: number;      // HTTP status code (200, 404, 500, etc.)
  headers: Record<string, string>;
  body: string | object;
}
```

### Context

The `Context` object provides access to platform features:

```typescript
interface Context {
  request: Request;
  env: Env;
  kv: KVNamespace;
  
  // Schedule background work
  waitUntil(promise: Promise<any>): void;
  
  // Call next middleware
  next(): Promise<Response>;
}
```

### KV Storage

Functions can use KV storage for persistent data:

```typescript
// Get a value
const value = await context.kv.get("my-key");

// Store a value
await context.kv.put("my-key", "my-value");

// Store with expiration (TTL in seconds)
await context.kv.put("cache-key", "data", { expirationTtl: 3600 });

// Delete a value
await context.kv.delete("my-key");

// List keys
const result = await context.kv.list({ prefix: "user:" });
```

### Environment Variables

Access secrets and configuration:

```typescript
// Access environment variable
const apiKey = context.env.API_KEY;

// With fallback
const logLevel = context.env.LOG_LEVEL ?? "info";
```

## Examples

See the [`examples/typescript/`](../examples/typescript/) directory for complete examples:

- [`hello-world/`](examples/typescript/hello-world/) - Basic function with types
- [`kv-store/`](examples/typescript/kv-store/) - Using KV storage
- [`http-api/`](examples/typescript/http-api/) - RESTful API with routing

## Runtimes

### Bun (Recommended)

Fastest runtime with native TypeScript support:

```json
{
  "runtime": "bun"
}
```

### Node.js

```json
{
  "runtime": "node20"
}
```

### Deno

```json
{
  "runtime": "deno"
}
```

## Best Practices

### 1. Always Type Your Inputs

```typescript
interface MyInput {
  name: string;
  age?: number;
}

const body = await request.json<MyInput>();
```

### 2. Use Error Handling

```typescript
try {
  // Your logic
} catch (error) {
  return {
    status: 500,
    headers: { "Content-Type": "application/json" },
    body: { error: error.message }
  };
}
```

### 3. Validate Input

```typescript
if (!body.name) {
  return {
    status: 400,
    headers: { "Content-Type": "application/json" },
    body: { error: "Name is required" }
  };
}
```

### 4. Use Async/Await

```typescript
const handler: Handler = async (request, env, context) => {
  const data = await someAsyncOperation();
  return { status: 200, headers: {}, body: data };
};
```

## IDE Support

### VS Code Extensions

Install recommended extensions:
- TypeScript and JavaScript Language Features
- ESLint
- Prettier
- Deno (for Deno runtime functions)

### Configuration

The `.vscode/settings.json` in this project provides optimal TypeScript settings.

## Troubleshooting

### Type Errors

If you see type errors, ensure your `tsconfig.json` has:
```json
{
  "compilerOptions": {
    "skipLibCheck": true,
    "strict": true
  }
}
```

### Module Not Found

Ensure your `tsconfig.json` has:
```json
{
  "compilerOptions": {
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true
  }
}
```

## Next Steps

- Learn about [KV Storage](kv-store.md)
- Build [REST APIs](http-api.md)
- Explore [Scheduled Functions](scheduled-functions.md)
- Read the [Runtime Specification](RUNTIME_SPEC.md)
