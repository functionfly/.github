/**
 * FunctionFly TypeScript Type Definitions
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

interface Response {
  status: number;
  headers: Record<string, string>;
  body: string | object;
}

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
  context: Context,
) => Promise<Response> | Response;

export type { Context, Env, Handler, KVNamespace, Request, Response };
