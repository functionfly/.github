/**
 * Context interface for FunctionFly function execution
 *
 * Provides metadata about the current function execution including
 * request information, timing, and platform details.
 */

import type { Response } from "./index";

/**
 * Attestation data returned by ctx.getAttestation()
 */
export interface Attestation {
  /** Attestation public ID (e.g. "att_a1b2c3...") */
  attestation_id: string;
  /** Function the attestation is for */
  function_id: string;
  /** Attestation type (verification, security_scan, code_review, execution, compliance, signature) */
  type: string;
  /** Status (valid, revoked, expired) */
  status: string;
  /** Human-readable title */
  title: string;
  /** Description */
  description?: string;
  /** Detailed results (JSON) */
  results?: Record<string, unknown>;
  /** SHA-256 proof hash */
  proof_hash: string;
  /** Ed25519/ECDSA signature */
  signature?: string;
  /** Public key ID that signed this attestation */
  public_key_id?: string;
  /** SHA-256 of function source code */
  code_hash?: string;
  /** SHA-256 of execution input */
  input_hash?: string;
  /** SHA-256 of execution output */
  output_hash?: string;
  /** ISO timestamp of attestation */
  attested_at: string;
  /** ISO timestamp of expiry (if any) */
  valid_until?: string;
  /** Whether cryptographic integrity is verified */
  is_valid: boolean;
  /** Whether the signature is valid */
  signature_valid: boolean;
}

/**
 * Options for ctx.delegate()
 */
export interface DelegateOptions {
  /** Minimum trust score required for the target function (0-100) */
  min_trust_score?: number;
  /** Minimum trust tier (untrusted, trusted, verified, highly_trusted) */
  min_trust_tier?: string;
  /** Timeout for the delegated call in milliseconds */
  timeout_ms?: number;
  /** Whether to retry on failure */
  retry?: boolean;
  /** Maximum number of retries */
  max_retries?: number;
}

/**
 * Result returned by ctx.delegate()
 */
export interface DelegateResult {
  /** Whether the delegation succeeded */
  success: boolean;
  /** Output from the delegated function */
  output: Record<string, unknown>;
  /** Execution ID of the delegated call */
  execution_id: string;
  /** Execution time in milliseconds */
  execution_time_ms: number;
  /** Function ID that was delegated to */
  function_id: string;
  /** Error message if delegation failed */
  error?: string;
}

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

  /**
   * Retrieve an attestation by ID.
   * Returns the attestation data including cryptographic proof and signature.
   *
   * @param attestationId - The attestation ID (e.g. "att_a1b2c3...")
   * @returns Attestation data, or null if not found
   */
  getAttestation?(attestationId: string): Promise<Attestation | null>;

  /**
   * Delegate execution to another function with trust-aware routing.
   * The platform will select the best instance of the target function
   * based on trust score, proximity, and availability.
   *
   * @param functionId - The target function ID to delegate to
   * @param input - The input to pass to the target function
   * @param options - Optional delegation options (trust requirements, timeout, retry)
   * @returns The result of the delegated execution
   */
  delegate?(
    functionId: string,
    input: Record<string, unknown>,
    options?: DelegateOptions,
  ): Promise<DelegateResult>;
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
