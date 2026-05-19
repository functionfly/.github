/**
 * Core types for FunctionFly Agent SDK
 */

/**
 * Trust tiers for functions
 */
export enum TrustTier {
  HighlyTrusted = "highly_trusted",
  Verified = "verified",
  Trusted = "trusted",
  Untrusted = "untrusted",
  Unknown = "unknown",
}

/**
 * Trust score components
 */
export interface TrustComponents {
  reliability: number;
  latency: number;
  errorRate: number;
  userRating: number;
  verification: number;
}

/**
 * Trust metrics
 */
export interface TrustMetrics {
  totalCalls: number;
  successRate: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  errorRate: number;
  timeoutRate: number;
}

/**
 * Trust diversity metrics
 */
export interface TrustDiversity {
  consumers: number;
  tenants: number;
  users: number;
}

/**
 * Trust score response from API
 */
export interface TrustScoreResponse {
  functionId: string;
  trustScore: number;
  trustTier: TrustTier;
  isVerified: boolean;
  verificationLevel: string;
  lastUpdated: string;
  windowStart: string;
  windowEnd: string;
  components: TrustComponents;
  metrics: TrustMetrics;
  diversity: TrustDiversity;
}

/**
 * Function metadata from registry
 */
export interface FunctionInfo {
  id: string;
  author: string;
  name: string;
  version: string;
  description: string;
  category: string;
  tags: string[];
  runtime: string;
  trustScore: number;
  trustLevel: string;
  verified: boolean;
  successRate: number;
  reliability: number;
  deterministic: boolean;
  inputs: Record<string, unknown>;
  outputs: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

/**
 * Function search/filter options
 */
export interface FunctionSearchOptions {
  query?: string;
  category?: string;
  runtime?: string;
  minTrustScore?: number;
  limit?: number;
  offset?: number;
  author?: string;
  tags?: string[];
}

/**
 * Paginated list response
 */
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
  hasMore: boolean;
}

/**
 * Function execution request
 */
export interface ExecuteRequest {
  input: Record<string, unknown>;
  context?: Record<string, unknown>;
  timeout?: number;
}

/**
 * Function execution response
 */
export interface ExecuteResponse {
  success: boolean;
  output: Record<string, unknown>;
  executionId: string;
  executionTimeMs: number;
  functionId: string;
  functionVersion: string;
  error?: string;
  errorCode?: string;
}

/**
 * Fallback function for retry
 */
export interface FallbackFunction {
  functionId: string;
  author: string;
  name: string;
  trustScore: number;
  reason?: string;
}

/**
 * Retry configuration
 */
export interface RetryConfig {
  maxRetries: number;
  baseDelayMs: number;
  maxDelayMs: number;
  backoffMultiplier: number;
  retryOnTimeout: boolean;
  retryOnError: boolean;
}

/**
 * Execution options with fallback
 */
export interface ExecutionOptions {
  enableFallback: boolean;
  minTrustScore?: number;
  preferredTrustTier?: TrustTier;
  retry?: Partial<RetryConfig>;
  timeout?: number;
}

/**
 * Client configuration
 */
export interface AgentClientConfig {
  apiKey: string;
  baseUrl?: string;
  appKey?: string;
  timeout?: number;
  retryConfig?: Partial<RetryConfig>;
}

/**
 * Agent SDK error
 */
export class AgentSDKError extends Error {
  constructor(
    message: string,
    public code: string,
    public statusCode?: number,
    public functionId?: string,
  ) {
    super(message);
    this.name = "AgentSDKError";
  }
}

/**
 * Function not found error
 */
export class FunctionNotFoundError extends AgentSDKError {
  constructor(functionId: string) {
    super(
      `Function not found: ${functionId}`,
      "FUNCTION_NOT_FOUND",
      404,
      functionId,
    );
    this.name = "FunctionNotFoundError";
  }
}

/**
 * Trust score too low error
 */
export class TrustScoreTooLowError extends AgentSDKError {
  constructor(functionId: string, trustScore: number, minTrustScore: number) {
    super(
      `Trust score ${trustScore} is below minimum ${minTrustScore} for function ${functionId}`,
      "TRUST_SCORE_TOO_LOW",
      400,
      functionId,
    );
    this.name = "TrustScoreTooLowError";
  }
}

/**
 * Execution error with fallback info
 */
export class ExecutionWithFallbackError extends AgentSDKError {
  constructor(
    message: string,
    public attemptedFunctions: FallbackFunction[],
    public lastError?: Error,
  ) {
    super(message, "EXECUTION_WITH_FALLBACK_FAILED");
    this.name = "ExecutionWithFallbackError";
  }
}
