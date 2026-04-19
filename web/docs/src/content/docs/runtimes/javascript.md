---
title: JavaScript Runtime
description: JavaScript runtime environment for FunctionFly functions
---

# JavaScript Runtime

FunctionFly's JavaScript runtime provides a modern environment for running JavaScript-based serverless functions.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| Node.js 18 | Supported | LTS |
| Node.js 20 | Supported | Recommended |
| Node.js 21 | Supported | Latest |

## Function Structure

A JavaScript function must export a handler function:

```javascript
// main.js
export default async function handler(request) {
    return {
        status: 200,
        body: { message: "Hello, World!" },
        headers: { "Content-Type": "application/json" }
    };
}
```

## Request Object

The `request` object contains:

| Property | Type | Description |
|----------|------|-------------|
| `body` | `any` | Parsed request body |
| `headers` | `Object` | HTTP headers (lowercase keys) |
| `params` | `Object` | URL query parameters |
| `method` | `string` | HTTP method |
| `url` | `string` | Full request URL |
| `path` | `string` | URL path |

## Response Format

Return an object with:

```javascript
return {
    status: 200,              // Required: HTTP status code
    body: "response data",    // Required: Any JSON-serializable data
    headers: {                // Optional: Response headers
        "Content-Type": "application/json"
    }
};
```

## Dependencies

### package.json

Include a `package.json` file:

```json
{
  "dependencies": {
    "axios": "^1.6.0",
    "zod": "^3.22.0"
  }
}
```

### Popular Packages

The following packages are pre-installed:

- Global: `fetch`, `console`, `Buffer`, `process`
- `axios` - HTTP client
- `node-fetch` - Fetch polyfill

## Module Support

### ES Modules (Recommended)

```javascript
// main.js
import { myHelper } from "./helpers.js";

export default async function handler(request) {
    const result = await myHelper(request.body);
    return { status: 200, body: result };
}
```

### CommonJS

```javascript
// main.js
const { myHelper } = require("./helpers");

module.exports = async function handler(request) {
    const result = await myHelper(request.body);
    return { status: 200, body: result };
};
```

## Example Functions

### HTTP API Handler

```javascript
export default async function handler(request) {
    const { method } = request;
    
    switch (method) {
        case "GET":
            return {
                status: 200,
                body: { users: ["alice", "bob"] }
            };
        
        case "POST": {
            const { name } = request.body || {};
            return {
                status: 201,
                body: { created: name }
            };
        }
        
        default:
            return {
                status: 405,
                body: { error: "Method not allowed" }
            };
    }
}
```

### Webhook Processor

```javascript
import crypto from "crypto";

export default async function handler(request) {
    // Verify signature
    const signature = request.headers["x-signature"] || "";
    const secret = process.env.WEBHOOK_SECRET || "";
    
    const body = JSON.stringify(request.body);
    const expected = crypto
        .createHmac("sha256", secret)
        .update(body)
        .digest("hex");
    
    const isValid = crypto.timingSafeEqual(
        Buffer.from(signature),
        Buffer.from(`sha256=${expected}`)
    );
    
    if (!isValid) {
        return {
            status: 401,
            body: { error: "Invalid signature" }
        };
    }
    
    // Process webhook
    await processEvent(request.body);
    
    return {
        status: 200,
        body: { received: true }
    };
}

async function processEvent(event) {
    // Your event processing logic
}
```

### Data Transformation

```javascript
export default async function handler(request) {
    const data = request.body || {};
    
    // Transform data
    const transformed = {
        id: data.id,
        name: `${data.first_name || ""} ${data.last_name || ""}`.trim(),
        email: (data.email || "").toLowerCase(),
        timestamp: data.created_at
    };
    
    return {
        status: 200,
        body: transformed
    };
}
```

### Using Fetch

```javascript
export default async function handler(request) {
    const response = await fetch("https://api.example.com/data", {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${process.env.API_KEY}`
        }
    });
    
    const data = await response.json();
    
    return {
        status: 200,
        body: data
    };
}
```

## Environment Variables

Access environment variables via `process.env`:

```javascript
const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
const debug = process.env.DEBUG === "true";
```

## File System

The `/tmp` directory is available for temporary file storage:

```javascript
import fs from "fs/promises";

// Write to temp file
await fs.writeFile("/tmp/data.json", JSON.stringify({ key: "value" }));

// Read from temp file
const data = JSON.parse(await fs.readFile("/tmp/data.json", "utf-8"));
```

Note: Files in `/tmp` are ephemeral and may not persist between invocations.

## Error Handling

```javascript
export default async function handler(request) {
    try {
        const result = await processRequest(request);
        return {
            status: 200,
            body: result
        };
    } catch (error) {
        console.error("Error:", error);
        
        if (error.name === "ValidationError") {
            return {
                status: 400,
                body: { error: error.message }
            };
        }
        
        return {
            status: 500,
            body: { error: "Internal server error" }
        };
    }
}

async function processRequest(request) {
    // Your processing logic
}
```

## Async/Await

The JavaScript runtime fully supports async/await:

```javascript
export default async function handler(request) {
    const [users, posts] = await Promise.all([
        fetch("https://api.example.com/users").then(r => r.json()),
        fetch("https://api.example.com/posts").then(r => r.json())
    ]);
    
    return {
        status: 200,
        body: { users, posts }
    };
}
```

## Timeout and Limits

| Resource | Default | Maximum |
|----------|---------|---------|
| Timeout | 30s | 300s (5 min) |
| Memory | 256 MB | 2048 MB |
| CPU | 1 vCPU | 4 vCPU |

Configure in `functionfly.jsonc`:

```jsonc
{
  "limits": {
    "timeout": 60,
    "memory": 512
  }
}
```

## Cold Start

JavaScript functions may experience cold starts:
- First invocation after deployment: ~100-300ms
- Subsequent invocations: ~5-20ms
- Keep functions warm with scheduled invocations

## Best Practices

1. **Use async/await** for asynchronous operations
2. **Use native fetch** when available (Node.js 18+)
3. **Minimize dependencies** to reduce cold start time
4. **Use environment variables** for configuration
5. **Handle errors** gracefully with try/catch
6. **Use ES modules** for better tree-shaking
7. **Avoid global state** between invocations
