/**
 * Function discovery module for FunctionFly Agent SDK
 */

import { parseTrustTier } from "./trust.js";
import type {
  FunctionInfo,
  FunctionSearchOptions,
  PaginatedResponse,
  TrustScoreResponse,
} from "./types.js";

export interface DiscoveryClient {
  /**
   * Search for functions in the registry
   */
  searchFunctions(
    options: FunctionSearchOptions,
  ): Promise<PaginatedResponse<FunctionInfo>>;

  /**
   * Get a specific function by author and name
   */
  getFunction(author: string, name: string): Promise<FunctionInfo>;

  /**
   * Get a specific function by ID
   */
  getFunctionById(functionId: string): Promise<FunctionInfo>;

  /**
   * Get trust score for a function
   */
  getTrustScore(functionId: string): Promise<TrustScoreResponse>;

  /**
   * List functions by category
   */
  listByCategory(
    category: string,
    limit?: number,
    offset?: number,
  ): Promise<PaginatedResponse<FunctionInfo>>;

  /**
   * List functions by author
   */
  listByAuthor(
    author: string,
    limit?: number,
    offset?: number,
  ): Promise<PaginatedResponse<FunctionInfo>>;

  /**
   * Find similar functions
   */
  findSimilar(
    author: string,
    name: string,
    limit?: number,
  ): Promise<FunctionInfo[]>;
}

/**
 * Transforms API response to FunctionInfo
 */
export function transformFunctionInfo(
  apiResponse: Record<string, unknown>,
): FunctionInfo {
  return {
    id: String(apiResponse.id || ""),
    author: String(apiResponse.author || ""),
    name: String(apiResponse.name || ""),
    version: String(
      apiResponse.version || apiResponse.latest_version || "v1.0.0",
    ),
    description: String(apiResponse.description || ""),
    category: String(apiResponse.category || ""),
    tags: Array.isArray(apiResponse.tags) ? apiResponse.tags.map(String) : [],
    runtime: String(apiResponse.runtime || "wasm"),
    trustScore: Number(apiResponse.trust_score || apiResponse.trustScore || 0),
    trustLevel: String(
      apiResponse.trust_level || apiResponse.trustLevel || "unknown",
    ),
    verified: Boolean(apiResponse.verified || apiResponse.is_verified || false),
    successRate: Number(
      apiResponse.success_rate || apiResponse.successRate || 0,
    ),
    reliability: Number(
      apiResponse.reliability || apiResponse.reliability_score || 0,
    ),
    deterministic: Boolean(apiResponse.deterministic || false),
    inputs: (apiResponse.inputs || apiResponse.input_schema || {}) as Record<
      string,
      unknown
    >,
    outputs: (apiResponse.outputs || apiResponse.output_schema || {}) as Record<
      string,
      unknown
    >,
    createdAt: String(
      apiResponse.created_at ||
        apiResponse.createdAt ||
        new Date().toISOString(),
    ),
    updatedAt: String(
      apiResponse.updated_at ||
        apiResponse.updatedAt ||
        new Date().toISOString(),
    ),
  };
}

/**
 * Transforms API trust score response
 */
export function transformTrustScore(
  apiResponse: Record<string, unknown>,
): TrustScoreResponse {
  const components = (apiResponse.components || {}) as Record<string, number>;
  const metrics = (apiResponse.metrics || {}) as Record<string, number>;
  const diversity = (apiResponse.diversity || {}) as Record<string, number>;

  return {
    functionId: String(apiResponse.function_id || apiResponse.functionId || ""),
    trustScore: Number(apiResponse.trust_score || apiResponse.trustScore || 0),
    trustTier: parseTrustTier(
      String(apiResponse.trust_tier || apiResponse.trustTier || "unknown"),
    ),
    isVerified: Boolean(
      apiResponse.is_verified || apiResponse.isVerified || false,
    ),
    verificationLevel: String(
      apiResponse.verification_level || apiResponse.verificationLevel || "",
    ),
    lastUpdated: String(
      apiResponse.last_updated ||
        apiResponse.lastUpdated ||
        new Date().toISOString(),
    ),
    windowStart: String(
      apiResponse.window_start || apiResponse.windowStart || "",
    ),
    windowEnd: String(apiResponse.window_end || apiResponse.windowEnd || ""),
    components: {
      reliability: components.reliability || 0,
      latency: components.latency || 0,
      errorRate: components.error_rate || components.errorRate || 0,
      userRating: components.user_rating || components.userRating || 0,
      verification: components.verification || 0,
    },
    metrics: {
      totalCalls: metrics.total_calls || metrics.totalCalls || 0,
      successRate: metrics.success_rate || metrics.successRate || 0,
      p50LatencyMs: metrics.p50_latency_ms || metrics.p50LatencyMs || 0,
      p95LatencyMs: metrics.p95_latency_ms || metrics.p95LatencyMs || 0,
      p99LatencyMs: metrics.p99_latency_ms || metrics.p99LatencyMs || 0,
      errorRate: metrics.error_rate || metrics.errorRate || 0,
      timeoutRate: metrics.timeout_rate || metrics.timeoutRate || 0,
    },
    diversity: {
      consumers: diversity.consumers || 0,
      tenants: diversity.tenants || 0,
      users: diversity.users || 0,
    },
  };
}

/**
 * Builds query string for function search
 */
export function buildSearchQuery(
  options: FunctionSearchOptions,
): URLSearchParams {
  const params = new URLSearchParams();

  if (options.query) {
    params.set("q", options.query);
  }
  if (options.category) {
    params.set("category", options.category);
  }
  if (options.runtime) {
    params.set("runtime", options.runtime);
  }
  if (options.minTrustScore !== undefined) {
    params.set("min_rating", String(options.minTrustScore / 100)); // API uses 0-1 scale
  }
  if (options.limit !== undefined) {
    params.set("limit", String(options.limit));
  }
  if (options.offset !== undefined) {
    params.set("offset", String(options.offset));
  }
  if (options.author) {
    params.set("author", options.author);
  }
  if (options.tags && options.tags.length > 0) {
    params.set("tags", options.tags.join(","));
  }

  return params;
}
