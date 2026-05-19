---
title: Deno Runtime
description: Deno runtime environment for FunctionFly functions
---

FunctionFly's Deno runtime provides a secure, TypeScript-first execution environment with built-in TypeScript support, web-standard APIs, and fine-grained security permissions.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| Deno 1.38+ | Supported | Recommended |
| Deno 2.0 | Supported | Latest |

## Runtime Features

The Deno runtime provides:

- **Native TypeScript** - No config or transpilation needed
- **Web-standard APIs** - Fetch, WebSocket, Crypto, URL
- **Fine-grained permissions** - Explicit access control
- **Built-in testing** - Deno test framework
- **Standard library** - Path, IO, encoding utilities

## Function Structure

A Deno function must export a handler:

```typescript
import type { Handler } from "@functionfly/runtime";

export const handler: Handler = async (request) => {
  return Response.json({
    message: "Hello from Deno!"
  });
};
```

## Request Object

The `Request` follows web standards:

```typescript
interface Request {
  method: string;           // GET, POST, etc.
  headers: Headers;         // HTTP headers
  url: string;              // Full URL
  path: string;             // URL path
  params: URLSearchParams;  // Query parameters
  body: unknown;            // Parsed body
}
```

## Response Format

Return a standard `Response`:

```typescript
return Response.json({
  status: 200,
  body: { message: "Hello" }
});

// Or with full control
return new Response(JSON.stringify(data), {
  status: 201,
  headers: { "Content-Type": "application/json" }
});
```

## Example Functions

### HTTP API Handler

```typescript
import type { Handler } from "@functionfly/runtime";

export const handler: Handler = async (req) => {
  const url = new URL(req.url);
  const path = url.pathname;

  if (req.method === "GET" && path === "/items") {
    const items = [
      { id: "1", name: "Item A", price: 9.99 },
      { id: "2", name: "Item B", price: 14.99 }
    ];
    return Response.json({ items });
  }

  if (req.method === "POST" && path === "/items") {
    const body = await req.json();
    const newItem = {
      id: crypto.randomUUID(),
      ...body
    };
    return Response.json(newItem, { status: 201 });
  }

  return Response.json({ error: "Not found" }, { status: 404 });
};
```

### Webhook Processor

```typescript
import { createHmac } from "crypto";
import type { Handler } from "@functionfly/runtime";

const WEBHOOK_SECRET = Deno.env.get("WEBHOOK_SECRET") || "";

export const handler: Handler = async (req) => {
  const signature = req.headers.get("x-signature") || "";
  const body = await req.text();

  if (!verifySignature(body, signature, WEBHOOK_SECRET)) {
    return Response.json({ error: "Invalid signature" }, { status: 401 });
  }

  const event = JSON.parse(body);
  console.log("Webhook event:", event.type);

  // Process the webhook
  await processWebhook(event);

  return Response.json({ received: true });
};

function verifySignature(body: string, signature: string, secret: string): boolean {
  const expected = createHmac("sha256", secret)
    .update(body)
    .digest("hex");

  return signature === `sha256=${expected}`;
}

async function processWebhook(event: Record<string, unknown>) {
  // Your webhook processing logic
}
```

### Data Transformation

```typescript
import type { Handler } from "@functionfly/runtime";

interface InputData {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  created_at: string;
}

export const handler: Handler = async (req) => {
  const body = await req.json() as InputData;

  const transformed = {
    id: body.id,
    name: `${body.first_name} ${body.last_name}`.trim(),
    email: body.email.toLowerCase(),
    timestamp: body.created_at
  };

  return Response.json(transformed);
};
```

## Deno-Specific Features

### KV Database

```typescript
import { serve } from "@functionfly/runtime";

// Use Deno KV for persistence
const kv = await Deno.openKv();

export const handler: Handler = async (req) => {
  const url = new URL(req.url);
  const path = url.pathname;

  if (req.method === "GET" && path === "/counter") {
    const count = await kv.get(["counter"]);
    return Response.json({ count: count.value ?? 0 });
  }

  if (req.method === "POST" && path === "/counter") {
    const current = await kv.get(["counter"]);
    const newCount = (current.value as number || 0) + 1;
    await kv.set(["counter"], newCount);
    return Response.json({ count: newCount });
  }

  return Response.json({ error: "Not found" }, { status: 404 });
};
```

### Fetch with Deno

```typescript
import type { Handler } from "@functionfly/runtime";

export const handler: Handler = async (req) => {
  // Make external API calls
  const response = await fetch("https://api.example.com/data", {
    headers: {
      "Authorization": `Bearer ${Deno.env.get("API_KEY")}`
    }
  });

  const data = await response.json();
  return Response.json(data);
};
```

### File System (Read Only)

```typescript
import { readFile } from "@functionfly/runtime";

export const handler: Handler = async (req) => {
  // Read static files from allowed paths
  const config = await readFile("./config.json", "utf-8");
  const settings = JSON.parse(config);

  return Response.json(settings);
};
```

## Permissions

The Deno runtime uses fine-grained permissions:

| Permission | Default | Description |
|------------|---------|-------------|
| `--allow-net` | Enabled | Network access (configurable hosts) |
| `--allow-env` | Enabled | Environment variables (whitelist) |
| `--allow-read` | Limited | File read (configurable paths) |
| `--allow-write` | Disabled | File write (disabled by default) |

## Environment Variables

Access via `Deno.env`:

```typescript
const apiKey = Deno.env.get("API_KEY");
const dbUrl = Deno.env.get("DATABASE_URL");
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 1024 MB |
| CPU | 10s | 60s |

## Cold Start

Deno has fast cold starts:

- First invocation: ~20-100ms
- Subsequent invocations: ~2-10ms

## Best Practices

1. **Use native TypeScript** - No transpilation
2. **Leverage Deno KV** - Built-in key-value store
3. **Handle permissions explicitly** - Request only what you need
4. **Use web-standard APIs** - Fetch, Response, Headers, etc.
5. **Keep functions stateless** - Use KV for persistence
6. **Handle errors gracefully** - Return appropriate status codes
7. **Minimize dependencies** - Use standard library when possible