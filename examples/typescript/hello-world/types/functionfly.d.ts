/**
 * FunctionFly TypeScript Type Definitions
 *
 * These types provide type safety and IDE autocomplete for writing
 * TypeScript functions that run on the FunctionFly platform.
 */

/**
 * HTTP Request object passed to function handlers.
 */
interface Request {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: any;
  json: <T = any>() => Promise<T>;
  text: () => Promise<string>;
  formData: () => Promise<FormData>;
}

/**
 * HTTP Response object returned by function handlers.
 */
interface Response {
  status: number;
  headers: Record<string, string>;
  body: string | object;
}

/**
 * KV (Key-Value) Storage namespace.
 */
interface KVNamespace {
  get(key: string): Promise<string | null>;
  put(
    key: string,
    value: string,
    options?: { expirationTtl?: number },
  ): Promise<void>;
  delete(key: string): Promise<void>;
  list(options?: {
    prefix?: string;
    limit?: number;
  }): Promise<{ keys: { name: string }[] }>;
}

/**
 * Environment variables available to the function.
 */
interface Env {
  [key: string]: string;
}

/**
 * Function execution context.
 */
interface Context {
  request: Request;
  env: Env;
  kv: KVNamespace;
  waitUntil(promise: Promise<any>): void;
  next(): Promise<Response>;
}

/**
 * Basic handler function signature.
 */
type Handler = (
  request: Request,
  env: Env,
  context: Context,
) => Promise<Response> | Response;

export type { Request, Response, KVNamespace, Env, Context, Handler };
