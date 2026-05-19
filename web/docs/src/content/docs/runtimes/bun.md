---
title: Bun Runtime
description: Bun runtime environment for FunctionFly functions
---

FunctionFly's Bun runtime provides a high-performance JavaScript/TypeScript execution environment with native TypeScript support, built-in utilities, and exceptional speed.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| Bun 1.0+ | Supported | Recommended |
| Bun 1.1 | Supported | Latest |

## Runtime Features

The Bun runtime provides:

- **Native TypeScript** - No transpilation step needed
- **Fast cold starts** - Extremely quick initialization
- **Built-in utilities** - `Bun.serve`, `SQLite`, file system APIs
- **Web-standard APIs** - Fetch, WebSocket, crypto
- **Hot reload** - Development mode with instant updates

## Function Structure

A Bun function must export a handler:

```typescript
export default {
  async fetch(request: Request): Promise<Response> {
    return Response.json({
      message: "Hello from Bun!"
    });
  }
};
```

## Request Object

The `Request` object follows web standards:

```typescript
interface Request {
  method: string;       // GET, POST, etc.
  headers: Headers;     // HTTP headers
  url: string;          // Full URL
  path: string;         // URL path
  params: URLSearchParams; // Query parameters
  body: unknown;        // Parsed body (JSON, form, etc.)
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
export default {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    if (request.method === "GET" && path === "/users") {
      const users = [
        { id: "1", name: "Alice", email: "alice@example.com" },
        { id: "2", name: "Bob", email: "bob@example.com" }
      ];
      return Response.json({ users });
    }

    if (request.method === "POST" && path === "/users") {
      const body = await request.json();
      const newUser = {
        id: crypto.randomUUID(),
        ...body
      };
      return Response.json(newUser, { status: 201 });
    }

    return Response.json({ error: "Not found" }, { status: 404 });
  }
};
```

### Webhook Processor

```typescript
import { crypto } from "crypto";

const WEBHOOK_SECRET = process.env.WEBHOOK_SECRET;

export default {
  async fetch(request: Request): Promise<Response> {
    const signature = request.headers.get("x-signature") || "";
    const body = await request.text();

    if (!verifySignature(body, signature, WEBHOOK_SECRET || "")) {
      return Response.json({ error: "Invalid signature" }, { status: 401 });
    }

    const event = JSON.parse(body);
    console.log("Webhook received:", event.type);

    return Response.json({ received: true });
  }
};

function verifySignature(body: string, signature: string, secret: string): boolean {
  const expected = crypto
    .createHmac("sha256", secret)
    .update(body)
    .digest("hex");

  return signature === `sha256=${expected}`;
}
```

### Data Transformation

```typescript
export default {
  async fetch(request: Request): Promise<Response> {
    const body = await request.json();

    const transformed = {
      id: body.id,
      name: `${body.first_name} ${body.last_name}`.trim(),
      email: body.email?.toLowerCase(),
      timestamp: body.created_at
    };

    return Response.json(transformed);
  }
};
```

## Bun-Specific Features

### SQLite Database

```typescript
import { Database } from "bun:sqlite";

const db = new Database(":memory:");
db.run("CREATE TABLE users (id TEXT, name TEXT)");

export default {
  async fetch(request: Request): Promise<Response> {
    db.run("INSERT INTO users VALUES (?, ?)", ["1", "Alice"]);

    const users = db.query("SELECT * FROM users").all();
    return Response.json({ users });
  }
};
```

### File System Access

```typescript
import { readFile, writeFile } from "fs/promises";
import { join } from "path";

export default {
  async fetch(request: Request): Promise<Response> {
    const filePath = join("/tmp", "data.json");

    if (request.method === "POST") {
      const body = await request.json();
      await writeFile(filePath, JSON.stringify(body));
      return Response.json({ saved: true });
    }

    const data = await readFile(filePath, "utf-8");
    return Response.json(JSON.parse(data));
  }
};
```

### Serving Static Files

```typescript
import { fileServer, serve } from "bun";

export default {
  port: 3000,
  fetch(request: Request): Response {
    return fileServer("./public")(request);
  }
};
```

## Environment Variables

Access via `process.env`:

```typescript
const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 1024 MB |
| CPU | 10s | 60s |

## Cold Start

Bun has extremely fast cold starts:

- First invocation: ~10-50ms
- Subsequent invocations: ~1-5ms

## Best Practices

1. **Use native TypeScript** - No transpilation needed
2. **Leverage built-ins** - Bun.serve, SQLite, etc.
3. **Keep functions stateless** - Between invocations
4. **Handle errors gracefully** - Return appropriate status codes
5. **Use `Bun.env` for env vars** - Type-safe access
6. **Minimize dependencies** - Use built-in APIs when possible