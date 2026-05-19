/**
 * Function execution module for FunctionFly Agent SDK
 */

import { calculateBackoffDelay, DEFAULT_RETRY_CONFIG } from "./trust.js";
import type { ExecuteRequest, ExecuteResponse, RetryConfig } from "./types.js";

export interface ExecutionClient {
  /**
   * Execute a function by ID
   */
  execute(
    functionId: string,
    request: ExecuteRequest,
  ): Promise<ExecuteResponse>;

  /**
   * Execute a function by author and name
   */
  executeByName(
    author: string,
    name: string,
    request: ExecuteRequest,
  ): Promise<ExecuteResponse>;
}

/**
 * Creates an abort controller with timeout
 */
export function createTimeoutController(
  timeoutMs?: number,
): AbortController | undefined {
  if (!timeoutMs || timeoutMs <= 0) {
    return undefined;
  }

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  // Clean up timeout on abort
  controller.signal.addEventListener("abort", () => clearTimeout(timeoutId));

  return controller;
}

/**
 * Determines if an error is retryable
 */
export function isRetryableError(error: Error, config: RetryConfig): boolean {
  // Timeout errors
  if (error.name === "AbortError" || error.name === "TimeoutError") {
    return config.retryOnTimeout;
  }

  // Network errors
  if (
    error.message.includes("network") ||
    error.message.includes("ECONNREFUSED")
  ) {
    return true;
  }

  // 5xx errors (server errors)
  if (
    error.message.includes("500") ||
    error.message.includes("502") ||
    error.message.includes("503")
  ) {
    return config.retryOnError;
  }

  // 429 Too Many Requests
  if (error.message.includes("429")) {
    return true;
  }

  return config.retryOnError;
}

/**
 * Parses execution response from API
 */
export function parseExecutionResponse(
  functionId: string,
  response: Record<string, unknown>,
): ExecuteResponse {
  return {
    success: Boolean(response.success ?? true),
    output: (response.output ?? response.result ?? {}) as Record<
      string,
      unknown
    >,
    executionId: String(response.execution_id || response.executionId || ""),
    executionTimeMs: Number(
      response.execution_time_ms || response.executionTimeMs || 0,
    ),
    functionId,
    functionVersion: String(
      response.function_version || response.functionVersion || "v1",
    ),
    error: response.error ? String(response.error) : undefined,
    errorCode: response.error_code ? String(response.error_code) : undefined,
  };
}

/**
 * Sleep utility for retry delays
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Executes a function with retry logic
 */
export async function executeWithRetry(
  client: ExecutionClient,
  functionId: string,
  request: ExecuteRequest,
  retryConfig: RetryConfig = DEFAULT_RETRY_CONFIG,
): Promise<ExecuteResponse> {
  let lastError: Error | undefined;

  for (let attempt = 0; attempt <= retryConfig.maxRetries; attempt++) {
    try {
      const response = await client.execute(functionId, {
        ...request,
        timeout: request.timeout || retryConfig.maxDelayMs,
      });

      if (response.success) {
        return response;
      }

      // If execution failed but didn't throw, check if retryable
      if (
        response.error &&
        !isRetryableError(new Error(response.error), retryConfig)
      ) {
        return response;
      }

      lastError = new Error(response.error || "Execution failed");
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));

      // Check if we should retry
      if (
        attempt < retryConfig.maxRetries &&
        isRetryableError(lastError, retryConfig)
      ) {
        const delay = calculateBackoffDelay(
          attempt,
          retryConfig.baseDelayMs,
          retryConfig.maxDelayMs,
          retryConfig.backoffMultiplier,
        );
        await sleep(delay);
        continue;
      }
    }

    // If this wasn't the last attempt, wait before retrying
    if (attempt < retryConfig.maxRetries) {
      const delay = calculateBackoffDelay(
        attempt,
        retryConfig.baseDelayMs,
        retryConfig.maxDelayMs,
        retryConfig.backoffMultiplier,
      );
      await sleep(delay);
    }
  }

  throw lastError || new Error("Execution failed after retries");
}
