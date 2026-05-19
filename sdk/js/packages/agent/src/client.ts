/**
 * Main Agent Client for FunctionFly SDK
 */

import {
  buildSearchQuery,
  transformFunctionInfo,
  transformTrustScore,
} from "./discovery.js";
import {
  createTimeoutController,
  parseExecutionResponse,
} from "./execution.js";
import {
  DEFAULT_RETRY_CONFIG,
  meetsTrustThreshold,
  selectBestFallback,
  sortByTrustScore,
} from "./trust.js";
import type {
  AgentClientConfig,
  ExecuteRequest,
  ExecuteResponse,
  ExecutionOptions,
  FallbackFunction,
  FunctionInfo,
  FunctionSearchOptions,
  PaginatedResponse,
  RetryConfig,
  TrustScoreResponse,
} from "./types.js";
import {
  AgentSDKError,
  ExecutionWithFallbackError,
  FunctionNotFoundError,
  TrustScoreTooLowError,
} from "./types.js";

const DEFAULT_BASE_URL = "https://api.functionfly.com";

/**
 * AgentClient is the main entry point for the FunctionFly Agent SDK.
 * It provides function discovery, trust-aware routing, and automatic retry with fallback.
 */
export class AgentClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly appKey?: string;
  private readonly defaultTimeout: number;
  private readonly retryConfig: RetryConfig;

  constructor(config: AgentClientConfig) {
    this.apiKey = config.apiKey;
    this.baseUrl = config.baseUrl || DEFAULT_BASE_URL;
    this.appKey = config.appKey;
    this.defaultTimeout = config.timeout || 30000;
    this.retryConfig = { ...DEFAULT_RETRY_CONFIG, ...config.retryConfig };
  }

  /**
   * Makes an authenticated request to the FunctionFly API
   */
  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Authorization: `Bearer ${this.apiKey}`,
      ...((options.headers as Record<string, string>) || {}),
    };

    if (this.appKey) {
      headers["X-App-Key"] = this.appKey;
    }

    const controller = createTimeoutController(this.defaultTimeout);

    try {
      const response = await fetch(url, {
        ...options,
        headers,
        signal: controller?.signal,
      });

      if (!response.ok) {
        const errorBody = await response.text().catch(() => "");
        throw new AgentSDKError(
          `API request failed: ${response.status} ${response.statusText} - ${errorBody}`,
          "API_ERROR",
          response.status,
        );
      }

      return (await response.json()) as T;
    } catch (error) {
      if (error instanceof AgentSDKError) {
        throw error;
      }
      if (error instanceof Error && error.name === "AbortError") {
        throw new AgentSDKError("Request timeout", "TIMEOUT", 408);
      }
      throw new AgentSDKError(
        error instanceof Error ? error.message : String(error),
        "NETWORK_ERROR",
      );
    }
  }

  // ============================================
  // Discovery Methods
  // ============================================

  /**
   * Search for functions in the registry
   */
  async searchFunctions(
    options: FunctionSearchOptions,
  ): Promise<PaginatedResponse<FunctionInfo>> {
    const query = buildSearchQuery(options);
    const endpoint = `/v1/registry/functions/search?${query.toString()}`;

    const response = await this.request<Record<string, unknown>>(endpoint);

    const functions = Array.isArray(response.functions)
      ? response.functions.map((f: Record<string, unknown>) =>
          transformFunctionInfo(f),
        )
      : [];

    const total = Number(response.total || 0);
    const limit = Number(response.limit || options.limit || 20);
    const offset = Number(response.offset || options.offset || 0);

    return {
      data: functions,
      total,
      limit,
      offset,
      hasMore: offset + functions.length < total,
    };
  }

  /**
   * List all functions with optional filters
   */
  async listFunctions(
    options: Partial<FunctionSearchOptions> = {},
  ): Promise<PaginatedResponse<FunctionInfo>> {
    const query = buildSearchQuery(options);
    const endpoint = `/v1/registry/functions?${query.toString()}`;

    const response = await this.request<Record<string, unknown>>(endpoint);

    const functions = Array.isArray(response.functions)
      ? response.functions.map((f: Record<string, unknown>) =>
          transformFunctionInfo(f),
        )
      : [];

    const total = Number(response.total || 0);
    const limit = Number(response.limit || options.limit || 20);
    const offset = Number(response.offset || options.offset || 0);

    return {
      data: functions,
      total,
      limit,
      offset,
      hasMore: offset + functions.length < total,
    };
  }

  /**
   * Get a specific function by author and name
   */
  async getFunction(author: string, name: string): Promise<FunctionInfo> {
    const endpoint = `/v1/registry/functions/${encodeURIComponent(author)}/${encodeURIComponent(name)}`;

    try {
      const response = await this.request<Record<string, unknown>>(endpoint);
      return transformFunctionInfo(response);
    } catch (error) {
      if (error instanceof AgentSDKError && error.statusCode === 404) {
        throw new FunctionNotFoundError(`${author}/${name}`);
      }
      throw error;
    }
  }

  /**
   * Get a specific function by ID
   */
  async getFunctionById(functionId: string): Promise<FunctionInfo> {
    // First we need to find the function to get author/name
    // This is a simplified approach - in production you might have a direct ID endpoint
    const searchResult = await this.searchFunctions({ limit: 100 });
    const found = searchResult.data.find((f) => f.id === functionId);

    if (!found) {
      throw new FunctionNotFoundError(functionId);
    }

    return found;
  }

  /**
   * Get trust score for a function
   */
  async getTrustScore(functionId: string): Promise<TrustScoreResponse> {
    const endpoint = `/v1/registry/functions/${functionId}/trust`;

    try {
      const response = await this.request<Record<string, unknown>>(endpoint);
      return transformTrustScore(response);
    } catch (error) {
      if (error instanceof AgentSDKError && error.statusCode === 404) {
        throw new FunctionNotFoundError(functionId);
      }
      throw error;
    }
  }

  /**
   * List functions by category
   */
  async listByCategory(
    category: string,
    limit: number = 20,
    offset: number = 0,
  ): Promise<PaginatedResponse<FunctionInfo>> {
    return this.listFunctions({ category, limit, offset });
  }

  /**
   * List functions by author
   */
  async listByAuthor(
    author: string,
    limit: number = 20,
    offset: number = 0,
  ): Promise<PaginatedResponse<FunctionInfo>> {
    return this.listFunctions({ author, limit, offset });
  }

  /**
   * Find similar functions
   */
  async findSimilar(
    author: string,
    name: string,
    limit: number = 5,
  ): Promise<FunctionInfo[]> {
    const endpoint = `/v1/registry/functions/${encodeURIComponent(author)}/${encodeURIComponent(name)}/similar?limit=${limit}`;

    const response = await this.request<Record<string, unknown>>(endpoint);

    const similar = Array.isArray(response.similar)
      ? response.similar.map((f: Record<string, unknown>) =>
          transformFunctionInfo(f),
        )
      : [];

    return similar;
  }

  /**
   * Discover functions suitable for an agent task
   */
  async discoverFunctions(
    options: {
      task?: string;
      category?: string;
      minTrustScore?: number;
      limit?: number;
    } = {},
  ): Promise<FunctionInfo[]> {
    const searchOptions: FunctionSearchOptions = {
      query: options.task,
      category: options.category,
      minTrustScore: options.minTrustScore,
      limit: options.limit || 50,
    };

    const result = await this.searchFunctions(searchOptions);

    // Filter and sort by trust score
    let functions = result.data;

    if (options.minTrustScore) {
      functions = functions.filter(
        (f) => f.trustScore >= options.minTrustScore!,
      );
    }

    return sortByTrustScore(functions);
  }

  // ============================================
  // Execution Methods
  // ============================================

  /**
   * Execute a function by ID
   */
  async execute(
    functionId: string,
    request: ExecuteRequest,
  ): Promise<ExecuteResponse> {
    const endpoint = `/v1/functions/${functionId}/execute`;

    try {
      const response = await this.request<Record<string, unknown>>(endpoint, {
        method: "POST",
        body: JSON.stringify({
          input: request.input,
          context: request.context,
        }),
      });

      return parseExecutionResponse(functionId, response);
    } catch (error) {
      if (error instanceof AgentSDKError && error.statusCode === 404) {
        throw new FunctionNotFoundError(functionId);
      }
      throw error;
    }
  }

  /**
   * Execute a function by author and name
   */
  async executeByName(
    author: string,
    name: string,
    request: ExecuteRequest,
  ): Promise<ExecuteResponse> {
    const fn = await this.getFunction(author, name);
    return this.execute(fn.id, request);
  }

  /**
   * Execute with automatic retry and trust-aware fallback
   */
  async executeWithRetry(
    functionId: string,
    input: Record<string, unknown>,
    options: ExecutionOptions = { enableFallback: true },
  ): Promise<ExecuteResponse> {
    const retryConfig: RetryConfig = {
      ...this.retryConfig,
      ...options.retry,
    };

    const attemptedFunctions: FallbackFunction[] = [];
    let lastError: Error | undefined;

    // First, check trust score if minimum is specified
    if (options.minTrustScore || options.preferredTrustTier) {
      try {
        const trust = await this.getTrustScore(functionId);
        if (
          !meetsTrustThreshold(
            trust.trustScore,
            options.minTrustScore,
            options.preferredTrustTier,
          )
        ) {
          throw new TrustScoreTooLowError(
            functionId,
            trust.trustScore,
            options.minTrustScore || 0,
          );
        }
      } catch (error) {
        if (error instanceof TrustScoreTooLowError) {
          throw error;
        }
        // If we can't get trust score, proceed with execution
      }
    }

    // Get function info for fallback tracking
    let currentFunction: FunctionInfo | undefined;
    try {
      currentFunction = await this.getFunctionById(functionId);
    } catch {
      // Function info not critical for execution
    }

    // Attempt execution with retry
    for (let attempt = 0; attempt <= retryConfig.maxRetries; attempt++) {
      try {
        const response = await this.execute(functionId, { input });

        if (response.success) {
          return response;
        }

        // Execution returned unsuccessful result
        lastError = new Error(response.error || "Execution failed");

        if (attempt < retryConfig.maxRetries) {
          await this.delay(this.calculateBackoff(attempt, retryConfig));
        }
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        // Track attempted function
        if (currentFunction) {
          attemptedFunctions.push({
            functionId: currentFunction.id,
            author: currentFunction.author,
            name: currentFunction.name,
            trustScore: currentFunction.trustScore,
            reason: lastError.message,
          });
        }

        // Check if we should try fallback
        if (options.enableFallback && attempt >= retryConfig.maxRetries - 1) {
          const fallback = await this.findFallbackFunction(
            functionId,
            input,
            options,
          );
          if (fallback) {
            currentFunction = fallback;
            functionId = fallback.id;
            attemptedFunctions.push({
              functionId: fallback.id,
              author: fallback.author,
              name: fallback.name,
              trustScore: fallback.trustScore,
              reason: `Fallback from previous failure`,
            });
            continue;
          }
        }

        // Check if error is retryable
        if (
          attempt < retryConfig.maxRetries &&
          this.isRetryableError(lastError)
        ) {
          await this.delay(this.calculateBackoff(attempt, retryConfig));
          continue;
        }

        break;
      }
    }

    throw new ExecutionWithFallbackError(
      `Execution failed after ${attemptedFunctions.length} attempt(s): ${lastError?.message || "Unknown error"}`,
      attemptedFunctions,
      lastError,
    );
  }

  /**
   * Find a fallback function when primary fails
   */
  private async findFallbackFunction(
    excludeFunctionId: string,
    input: Record<string, unknown>,
    options: ExecutionOptions,
  ): Promise<FunctionInfo | undefined> {
    // Find similar functions with same category and higher trust score
    if (!options.minTrustScore && !options.preferredTrustTier) {
      return undefined;
    }

    // Search for functions with same category
    const searchResult = await this.searchFunctions({
      limit: 20,
      minTrustScore: options.minTrustScore || 50,
    });

    const candidates = searchResult.data.filter(
      (f) =>
        f.id !== excludeFunctionId &&
        f.trustScore > (options.minTrustScore || 0),
    );

    return selectBestFallback(candidates, [excludeFunctionId]);
  }

  /**
   * Calculate backoff delay
   */
  private calculateBackoff(attempt: number, config: RetryConfig): number {
    const delay = Math.min(
      config.baseDelayMs * Math.pow(config.backoffMultiplier, attempt),
      config.maxDelayMs,
    );
    // Add jitter
    return delay * (0.5 + Math.random() * 0.5);
  }

  /**
   * Check if error is retryable
   */
  private isRetryableError(error: Error): boolean {
    if (error.name === "AbortError" || error.name === "TimeoutError") {
      return this.retryConfig.retryOnTimeout;
    }
    if (
      error.message.includes("network") ||
      error.message.includes("ECONNREFUSED")
    ) {
      return true;
    }
    if (
      error.message.includes("429") ||
      error.message.includes("500") ||
      error.message.includes("503")
    ) {
      return this.retryConfig.retryOnError;
    }
    return this.retryConfig.retryOnError;
  }

  /**
   * Delay utility
   */
  private delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  // ============================================
  // Tool Builder (for Agent Frameworks)
  // ============================================

  /**
   * Build a tool definition for agent frameworks
   */
  buildTool(function_: FunctionInfo): AgentTool {
    return {
      name: `${function_.author}_${function_.name}`.replace(
        /[^a-zA-Z0-9_]/g,
        "_",
      ),
      description:
        function_.description ||
        `${function_.name} function by ${function_.author}`,
      parameters: function_.inputs || {},
      trustScore: function_.trustScore,
      trustLevel: function_.trustLevel,
      verified: function_.verified,
      functionId: function_.id,
      execute: (input: Record<string, unknown>) =>
        this.execute(function_.id, { input }),
    };
  }

  /**
   * Build multiple tools for agent frameworks
   */
  async buildTools(options: FunctionSearchOptions = {}): Promise<AgentTool[]> {
    const result = await this.searchFunctions({
      ...options,
      limit: options.limit || 50,
    });
    return result.data.map((f) => this.buildTool(f));
  }
}

/**
 * Agent tool definition (compatible with LangChain, AutoGen, CrewAI)
 */
export interface AgentTool {
  name: string;
  description: string;
  parameters: Record<string, unknown>;
  trustScore: number;
  trustLevel: string;
  verified: boolean;
  functionId: string;
  execute: (input: Record<string, unknown>) => Promise<ExecuteResponse>;
}

export type {
  AgentClientConfig,
  ExecuteRequest,
  ExecuteResponse,
  FunctionInfo,
  FunctionSearchOptions,
};
