/**
 * FunctionFly TypeScript Type Definitions
 *
 * These types provide type safety and IDE autocomplete for writing
 * TypeScript functions that run on the FunctionFly platform.
 *
 * @packageDocumentation
 * @version 1.0.0
 */

/**
 * HTTP Request object passed to function handlers.
 * Contains all information about the incoming HTTP request.
 *
 * @example
 * ```typescript
 * async function handler(request: Request) {
 *   const method = request.method;
 *   const url = request.url;
 *   const headers = request.headers;
 *
 *   // Parse JSON body
 *   const data = await request.json<{ name: string }>();
 *
 *   // Or get raw text
 *   const text = await request.text();
 * }
 * ```
 */
interface Request {
  /** HTTP method (GET, POST, PUT, DELETE, etc.) */
  method: string;

  /** Full URL of the request including query parameters */
  url: string;

  /** Request headers as key-value pairs */
  headers: Record<string, string>;

  /** Raw request body (any type) */
  body: any;

  /**
   * Parse the request body as JSON and return typed data
   * @returns Promise resolving to parsed JSON data
   */
  json: <T = any>() => Promise<T>;

  /**
   * Get the request body as a plain text string
   * @returns Promise resolving to body text
   */
  text: () => Promise<string>;

  /**
   * Parse the request body as FormData
   * @returns Promise resolving to FormData object
   */
  formData: () => Promise<FormData>;
}

/**
 * HTTP Response object returned by function handlers.
 * Functions must return a Response object to send data back to the client.
 *
 * @example
 * ```typescript
 * // Simple JSON response
 * return {
 *   status: 200,
 *   headers: { "Content-Type": "application/json" },
 *   body: { message: "Hello, World!" }
 * };
 *
 * // Text response
 * return {
 *   status: 200,
 *   headers: { "Content-Type": "text/plain" },
 *   body: "Hello, World!"
 * };
 * ```
 */
interface Response {
  /** HTTP status code (200, 404, 500, etc.) */
  status: number;

  /** Response headers as key-value pairs */
  headers: Record<string, string>;

  /** Response body - can be string or object (will be JSON serialized) */
  body: string | object;
}

/**
 * KV (Key-Value) Storage namespace for persistent data storage.
 * Provides synchronous-looking async API for storing and retrieving data.
 *
 * @example
 * ```typescript
 * // Store a value
 * await context.kv.put("user:123", JSON.stringify({ name: "Alice" }));
 *
 * // Retrieve a value
 * const data = await context.kv.get("user:123");
 * if (data) {
 *   const user = JSON.parse(data);
 * }
 *
 * // Delete a value
 * await context.kv.delete("user:123");
 *
 * // List all keys with prefix
 * const result = await context.kv.list({ prefix: "user:" });
 * for (const key of result.keys) {
 *   console.log(key.name);
 * }
 * ```
 */
interface KVNamespace {
  /**
   * Get a value by key
   * @param key - The key to retrieve
   * @returns Promise resolving to the value or null if not found
   */
  get(key: string): Promise<string | null>;

  /**
   * Store a value with optional expiration
   * @param key - The key to store
   * @param value - The value to store (will be converted to string)
   * @param options - Optional configuration (expirationTtl in seconds)
   */
  put(key: string, value: string, options?: {
    /** Time-to-live in seconds before the key expires */
    expirationTtl?: number;
  }): Promise<void>;

  /**
   * Delete a key from the KV store
   * @param key - The key to delete
   */
  delete(key: string): Promise<void>;

  /**
   * List all keys in the KV namespace
   * @param options - Optional configuration
   * @returns Promise resolving to list of keys
   */
  list(options?: {
    /** Only return keys that start with this prefix */
    prefix?: string;
    /** Maximum number of keys to return (default: 1000) */
    limit?: number;
  }): Promise<{
    /** Array of keys in the namespace */
    keys: Array<{ name: string }>;
  }>;
}

/**
 * Environment variables available to the function.
 * Contains user-defined secrets and configuration values.
 *
 * @example
 * ```typescript
 * // Access environment variables
 * const apiKey = context.env.API_KEY;
 * const databaseUrl = context.env.DATABASE_URL;
 *
 * // Type-safe access with optional fallback
 * const logLevel = context.env.LOG_LEVEL ?? "info";
 * ```
 */
interface Env {
  /**
   * User-defined secrets and configuration.
   * Access via dot notation: env.VAR_NAME
   */
  [key: string]: string;
}

/**
 * Function execution context containing request, environment, and utilities.
 * This is the main object passed to function handlers.
 *
 * @example
 * ```typescript
 * async function handler(context: Context) {
 *   // Access request data
 *   const method = context.request.method;
 *   const body = await context.request.json();
 *
 *   // Access environment variables
 *   const apiKey = context.env.API_KEY;
 *
 *   // Use KV storage
 *   await context.kv.put("cacheKey", JSON.stringify({ data: "value" }));
 *
 *   // Schedule background work
 *   context.waitUntil(someAsyncTask());
 *
 *   // For middleware: call next handler
 *   return context.next();
 * }
 * ```
 */
interface Context {
  /** The incoming HTTP request */
  request: Request;

  /** Environment variables and secrets */
  env: Env;

  /** KV Storage namespace for persistent data */
  kv: KVNamespace;

  /**
   * Schedule a promise to run in the background.
   * The platform will wait for this promise to resolve before
   * terminating the function, but won't block the response.
   * @param promise - The promise to run in the background
   */
  waitUntil(promise: Promise<any>): void;

  /**
   * Call the next handler in the middleware chain.
   * Only used in middleware functions.
   * @returns Promise resolving to the next handler's response
   */
  next(): Promise<Response>;
}

/**
 * Basic handler function signature.
 * This is the simplest form of a FunctionFly function handler.
 *
 * @example
 * ```typescript
 * const handler: Handler = async (request, env) => {
 *   return {
 *     status: 200,
 *     headers: { "Content-Type": "application/json" },
 *     body: { message: "Hello from FunctionFly!" }
 *   };
 * };
 * ```
 *
 * @param request - The incoming HTTP request
 * @param env - Environment variables
 * @param context - Additional context (KV storage, waitUntil, etc.)
 * @returns Response object or Promise resolving to Response
 */
type Handler = (
  request: Request,
  env: Env,
  context: Context
) => Promise<Response> | Response;

/**
 * HTTP handler with full context object.
 * This is the recommended handler type for most functions.
 *
 * @example
 * ```typescript
 * const httpHandler: HTTPHandler = async (context) => {
 *   const { method, url } = context.request;
 *
 *   if (method === "GET" && url === "/health") {
 *     return {
 *       status: 200,
 *       headers: { "Content-Type": "application/json" },
 *       body: { status: "healthy" }
 *     };
 *   }
 *
 *   return {
 *     status: 404,
 *     headers: { "Content-Type": "application/json" },
 *     body: { error: "Not found" }
 *   };
 * };
 * ```
 *
 * @param context - Full function context with request, env, kv, etc.
 * @returns Promise resolving to Response
 */
type HTTPHandler = (context: Context) => Promise<Response>;

/**
 * Scheduled handler for cron-triggered functions.
 * Functions triggered by schedules receive this event object.
 *
 * @example
 * ```typescript
 * const scheduledHandler: ScheduledHandler = async (event, env) => {
 *   console.log(`Scheduled task ran at ${event.timestamp}`);
 *   console.log(`Cron expression: ${event.cron}`);
 *
 *   // Perform cleanup, send reports, etc.
 *   await performDailyTask();
 * };
 * ```
 *
 * @param event - Scheduled event containing cron info and timestamp
 * @param env - Environment variables
 */
type ScheduledHandler = (
  event: {
    /** Cron expression that triggered this execution */
    cron: string;
    /** Unix timestamp of execution */
    timestamp: number;
  },
  env: Env
) => Promise<void>;

/**
 * WebSocket handler for real-time connections.
 *
 * @example
 * ```typescript
 * const wsHandler: WebSocketHandler = async (context, socket) => {
 *   // Handle incoming messages
 *   socket.onmessage = async (event) => {
 *     const data = JSON.parse(event.data);
 *     // Process message...
 *     socket.send(JSON.stringify({ echo: data }));
 *   };
 *
 *   // Keep connection alive
 *   socket.send(JSON.stringify({ type: "connected" }));
 * };
 * ```
 */
type WebSocketHandler = (
  context: Context,
  socket: WebSocket
) => Promise<void>;

/**
 * Middleware handler function.
 * Middleware can modify requests before they reach the main handler,
 * or modify responses before they're sent back to the client.
 *
 * @example
 * ```typescript
 * const authMiddleware: Middleware = async (context) => {
 *   const token = context.request.headers["authorization"];
 *
 *   if (!token) {
 *     return {
 *       status: 401,
 *       headers: { "Content-Type": "application/json" },
 *       body: { error: "Missing authorization token" }
 *     };
 *   }
 *
 *   // Verify token and add user to context
 *   // Then call next handler
 *   return context.next();
 * };
 * ```
 */
type Middleware = (context: Context) => Promise<Response>;

/**
 * Edge-specific types for Cloudflare Workers compatibility
 */
declare namespace Edge {
  /**
   * Cloudflare Workers style Request (subset)
   */
  interface RequestInit {
    method?: string;
    headers?: Record<string, string>;
    body?: string | FormData;
  }

  /**
   * Cache API for edge caching
   */
  interface Cache {
    match(request: Request): Promise<Response | undefined>;
    put(request: Request, response: Response): Promise<void>;
    delete(request: Request): Promise<boolean>;
  }

  /**
   * Global cache instance
   */
  const caches: {
    open(cacheName: string): Promise<Cache>;
  };
}

/**
 * Type exports for convenience
 */
export type {
  Request,
  Response,
  KVNamespace,
  Env,
  Context,
  Handler,
  HTTPHandler,
  ScheduledHandler,
  WebSocketHandler,
  Middleware
};

/**
 * Module declaration for global types
 */
export {};
