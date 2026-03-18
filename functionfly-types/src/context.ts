/**
 * Context interface for FunctionFly function execution
 *
 * Provides metadata about the current function execution including
 * request information, timing, and platform details.
 */

import type { Response } from "./index";

/**
 * Function execution context
 */
export interface Context {
  /** Unique identifier for this request */
  requestId: string;

  /** Timestamp when execution started (Unix epoch in milliseconds) */
  startTime: number;

  /** Maximum execution timeout in milliseconds */
  executionTimeout: number;

  /** Region where the function is executing */
  region: string;

  /** Tenant/workspace ID */
  tenantId: string;

  /** Function version */
  version?: string;

  /** Function name */
  functionName?: string;

  /** Platform-specific waitUntil for async operations */
  waitUntil?(promise: Promise<any>): void;

  /** Platform-specific next() for middleware chaining */
  next?(): Promise<Response>;
}

/**
 * WaitUntilOptions for background tasks
 */
export interface WaitUntilOptions {
  /** Label for the background task */
  label?: string;
}

/**
 * Extended context with helper methods
 */
export interface ExtendedContext extends Context {
  /** Get execution duration so far */
  getDuration(): number;

  /** Check if execution is about to timeout */
  isTimeoutImminent(thresholdMs?: number): boolean;

  /** Wait for background tasks to complete */
  waitUntil(promise: Promise<any>, options?: WaitUntilOptions): void;
}
