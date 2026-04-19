---
title: TypeScript Runtime
description: TypeScript runtime environment for FunctionFly functions
---

# TypeScript Runtime

FunctionFly's TypeScript runtime compiles your TypeScript code to JavaScript and executes it in the Node.js environment.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| TypeScript 5.0 | Supported | Recommended |
| TypeScript 5.3 | Supported | Latest |

## Function Structure

A TypeScript function must export a typed handler:

```typescript
// main.ts
import { Request, Response } from "@functionfly/types";

export default async function handler(request: Request): Promise<Response> {
    return {
        status: 200,
        body: { message: "Hello, World!" },
        headers: { "Content-Type": "application/json" }
    };
}
```

## Type Definitions

### Request Type

```typescript
interface Request {
    body: any;
    headers: Record<string, string>;
    params: Record<string, string>;
    method: string;
    url: string;
    path: string;
}
```

### Response Type

```typescript
interface Response {
    status: number;
    body: any;
    headers?: Record<string, string>;
}
```

## Project Structure

```
my-function/
├── main.ts           # Entry point
├── types.ts          # Custom types
├── utils.ts          # Helper functions
├── package.json      # Dependencies
└── tsconfig.json     # TypeScript config
```

## tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "node",
    "esModuleInterop": true,
    "strict": true,
    "skipLibCheck": true,
    "outDir": "./dist"
  },
  "include": ["*.ts"],
  "exclude": ["node_modules"]
}
```

## Dependencies

### package.json

```json
{
  "dependencies": {
    "@types/node": "^20.0.0",
    "zod": "^3.22.0"
  }
}
```

## Typed Request/Response

### Strict Typing

```typescript
// types.ts
interface UserRequest {
    name: string;
    email: string;
}

interface UserResponse {
    id: string;
    name: string;
    email: string;
    createdAt: string;
}

// main.ts
import { Request, Response } from "@functionfly/types";
import { UserRequest, UserResponse } from "./types";

export default async function handler(
    request: Request<UserRequest>
): Promise<Response<UserResponse>> {
    const { name, email } = request.body;
    
    const user: UserResponse = {
        id: crypto.randomUUID(),
        name,
        email,
        createdAt: new Date().toISOString()
    };
    
    return {
        status: 201,
        body: user
    };
}
```

### Validation with Zod

```typescript
import { z } from "zod";
import { Request, Response } from "@functionfly/types";

const UserSchema = z.object({
    name: z.string().min(1),
    email: z.string().email(),
    age: z.number().min(0).optional()
});

type UserInput = z.infer<typeof UserSchema>;

export default async function handler(
    request: Request
): Promise<Response> {
    const result = UserSchema.safeParse(request.body);
    
    if (!result.success) {
        return {
            status: 400,
            body: {
                error: "Validation failed",
                issues: result.error.issues
            }
        };
    }
    
    const user = result.data;
    
    return {
        status: 200,
        body: {
            message: `Hello, ${user.name}!`,
            email: user.email
        }
    };
}
```

## Example Functions

### HTTP API Handler with Types

```typescript
import { Request, Response } from "@functionfly/types";

interface CreateUserInput {
    name: string;
    email: string;
}

interface User {
    id: string;
    name: string;
    email: string;
}

export default async function handler(
    request: Request
): Promise<Response> {
    const { method } = request;
    
    switch (method) {
        case "GET": {
            const users: User[] = [
                { id: "1", name: "Alice", email: "alice@example.com" },
                { id: "2", name: "Bob", email: "bob@example.com" }
            ];
            
            return {
                status: 200,
                body: { users }
            };
        }
        
        case "POST": {
            const body = request.body as CreateUserInput;
            const newUser: User = {
                id: crypto.randomUUID(),
                name: body.name,
                email: body.email
            };
            
            return {
                status: 201,
                body: newUser
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

### Generic Handler Helper

```typescript
import { Request, Response } from "@functionfly/types";

interface Handler<TInput, TOutput> {
    (request: Request<TInput>): Promise<Response<TOutput>>;
}

function createHandler<TInput, TOutput>(
    fn: (input: TInput) => Promise<TOutput>
): Handler<TInput, TOutput> {
    return async (request): Promise<Response<TOutput>> => {
        try {
            const output = await fn(request.body);
            return {
                status: 200,
                body: output
            };
        } catch (error) {
            return {
                status: 500,
                body: { error: (error as Error).message } as unknown as TOutput
            };
        }
    };
}

// Usage
interface GreetInput {
    name: string;
}

interface GreetOutput {
    message: string;
}

const greetHandler = createHandler<GreetInput, GreetOutput>(
    async (input) => ({
        message: `Hello, ${input.name}!`
    })
);

export default greetHandler;
```

### Middleware Pattern

```typescript
import { Request, Response } from "@functionfly/types";

type Middleware = (
    request: Request,
    next: () => Promise<Response>
) => Promise<Response>;

function withAuth(middleware: Middleware): Middleware {
    return async (request, next) => {
        const token = request.headers["authorization"];
        
        if (!token) {
            return {
                status: 401,
                body: { error: "Unauthorized" }
            };
        }
        
        return next();
    };
}

function withLogging(middleware: Middleware): Middleware {
    return async (request, next) => {
        console.log(`[${new Date().toISOString()}] ${request.method} ${request.path}`);
        const response = await next();
        console.log(`[${new Date().toISOString()}] Response: ${response.status}`);
        return response;
    };
}

// Compose middleware
function compose(...middlewares: Middleware[]) {
    return async (request: Request): Promise<Response> => {
        let index = 0;
        
        async function next(): Promise<Response> {
            if (index >= middlewares.length) {
                return { status: 404, body: { error: "Not found" } };
            }
            
            const middleware = middlewares[index++];
            return middleware(request, next);
        }
        
        return next();
    };
}

// Main handler
async function mainHandler(request: Request): Promise<Response> {
    return {
        status: 200,
        body: { message: "Hello, World!" }
    };
}

// Export composed handler
export default compose(
    withLogging,
    withAuth
)(mainHandler);
```

## Environment Variables

Access environment variables with proper typing:

```typescript
declare global {
    namespace NodeJS {
        interface ProcessEnv {
            API_KEY: string;
            DATABASE_URL: string;
            DEBUG?: string;
        }
    }
}

const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
const debug = process.env.DEBUG === "true";
```

## File System

```typescript
import { promises as fs } from "fs";

async function readTempFile(filename: string): Promise<string> {
    return fs.readFile(`/tmp/${filename}`, "utf-8");
}

async function writeTempFile(filename: string, data: string): Promise<void> {
    await fs.writeFile(`/tmp/${filename}`, data);
}
```

## Error Handling with Types

```typescript
import { Request, Response } from "@functionfly/types";

class FunctionError extends Error {
    constructor(
        message: string,
        public status: number
    ) {
        super(message);
        this.name = "FunctionError";
    }
}

class ValidationError extends FunctionError {
    constructor(message: string) {
        super(message, 400);
        this.name = "ValidationError";
    }
}

class NotFoundError extends FunctionError {
    constructor(message: string) {
        super(message, 404);
        this.name = "NotFoundError";
    }
}

async function processRequest(request: Request): Promise<any> {
    // Your logic here
    throw new NotFoundError("Resource not found");
}

export default async function handler(
    request: Request
): Promise<Response> {
    try {
        const result = await processRequest(request);
        return {
            status: 200,
            body: result
        };
    } catch (error) {
        if (error instanceof FunctionError) {
            return {
                status: error.status,
                body: { error: error.message }
            };
        }
        
        console.error("Unexpected error:", error);
        return {
            status: 500,
            body: { error: "Internal server error" }
        };
    }
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
  "runtime": "typescript",
  "limits": {
    "timeout": 60,
    "memory": 512
  }
}
```

## Best Practices

1. **Use strict mode** in `tsconfig.json`
2. **Define interfaces** for request/response types
3. **Validate inputs** with Zod or similar libraries
4. **Use type guards** for runtime safety
5. **Export default** the handler function
6. **Avoid `any`** - use proper types
7. **Use const assertions** for literal types
