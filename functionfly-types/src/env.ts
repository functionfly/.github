/**
 * Environment interface for FunctionFly functions
 *
 * Provides access to:
 * - Environment variables
 * - Secrets (decrypted client-side for zero-knowledge vault)
 * - Key-value storage
 * - HTTP fetch capabilities
 */

import type { Request, Response } from "./index";

/**
 * KV (Key-Value) Storage namespace interface
 */
export interface KVNamespace {
  /** Get a value by key */
  get(key: string): Promise<string | null>;
  /** Set a value with optional TTL */
  put(
    key: string,
    value: string,
    options?: { expirationTtl?: number },
  ): Promise<void>;
  /** Delete a key */
  delete(key: string): Promise<void>;
  /** List keys with optional prefix */
  list(options?: {
    prefix?: string;
    limit?: number;
  }): Promise<{ keys: { name: string }[] }>;
}

/**
 * Environment variables and capabilities available to functions
 */
export interface Env extends Record<string, any> {
  // Environment variables (string-indexed)
  [key: string]: any;

  /** Get a secret from the vault (zero-knowledge, decrypted client-side) */
  getSecret?(key: string): Promise<string>;

  /** Get value from key-value store */
  kv?: KVNamespace;

  /** Make HTTP requests */
  fetch?(url: string | Request, init?: RequestInit): Promise<Response>;
}

/**
 * Extended environment with explicit methods
 * This is the preferred interface for new functions
 */
export interface ExtendedEnv {
  /** Get an environment variable */
  get(key: string): string | undefined;

  /** Get an environment variable with a default */
  getOr(key: string, defaultValue: string): string;

  /** Check if an environment variable is set */
  has(key: string): boolean;

  /** Get a secret from the vault */
  getSecret(key: string): Promise<string>;

  /** Get value from key-value store */
  kv: KVNamespace;

  /** Make HTTP requests */
  fetch(url: string | Request, init?: RequestInit): Promise<Response>;
}

/**
 * RequestInit for fetch (subset of standard Web API)
 */
export interface RequestInit {
  method?: string;
  headers?: Record<string, string>;
  body?: string | object;
}
