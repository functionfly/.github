/**
 * FunctionFly TypeScript Type Definitions
 *
 * This package provides TypeScript type definitions for writing functions
 * that run on the FunctionFly platform. It supports multiple runtimes
 * including Bun, Deno, Node.js, and TypeScript WASM.
 *
 * @module functionfly-types
 */

import type { Context } from "./context";
import type { Env } from "./env";

// Re-export all types
export * from "./context";
export * from "./env";

/**
 * HTTP Request object passed to function handlers.
 * This is the standard request format across all FunctionFly runtimes.
 */
export interface Request {
  /** HTTP method (GET, POST, PUT, DELETE, etc.) */
  method: string;
  /** Full URL of the request */
  url: string;
  /** Request headers */
  headers: Record<string, string>;
  /** Request body (raw) */
  body?: any;
  /** Parse body as JSON */
  json: <T = any>() => Promise<T>;
  /** Get body as text */
  text: () => Promise<string>;
  /** Get body as form data (entries compatible with Web API FormData) */
  formData: () => Promise<FormDataLike>;
}

/** Minimal form-data type (compatible with Web API FormData) so types work without DOM lib */
export interface FormDataLike {
  get(key: string): string | null;
  getAll(key: string): string[];
  has(key: string): boolean;
  set(key: string, value: string): void;
  append(key: string, value: string): void;
  delete(key: string): void;
  entries(): IterableIterator<[string, string]>;
  keys(): IterableIterator<string>;
  values(): IterableIterator<string>;
}

/**
 * HTTP Response object returned by function handlers.
 */
export interface Response {
  /** HTTP status code */
  status: number;
  /** Response headers */
  headers: Record<string, string>;
  /** Response body (can be string or object) */
  body: string | object;
}

/**
 * Handler function signature.
 * All FunctionFly functions must export a handler with this signature.
 *
 * @param request - The incoming HTTP request
 * @param env - Environment variables and secret access
 * @param context - Function execution context
 * @returns Response object or Promise resolving to Response
 */
export type Handler = (
  request: Request,
  env: Env,
  context: Context,
) => Promise<Response> | Response;

/**
 * Creates a JSON response with standard headers
 */
export function jsonResponse(
  body: object,
  status = 200,
  headers: Record<string, string> = {},
): Response {
  return {
    status,
    headers: {
      "Content-Type": "application/json",
      ...headers,
    },
    body,
  };
}

/**
 * Creates a text response
 */
export function textResponse(
  body: string,
  status = 200,
  headers: Record<string, string> = {},
): Response {
  return {
    status,
    headers: {
      "Content-Type": "text/plain",
      ...headers,
    },
    body,
  };
}

/**
 * Creates an error response
 */
export function errorResponse(message: string, status = 500): Response {
  return jsonResponse({ error: message }, status);
}

export { Context } from "./context";
export { Env } from "./env";
