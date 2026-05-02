import { apiClient } from './client';

export interface RegistryFunction {
  id: string;
  author: string;
  name: string;
  title?: string;
  description?: string;
  category?: string;
  tags: string[];
  visibility: string;
  price_per_call: number;
  popularity_score: number;
  reliability_score: number;
  deterministic_score: number;
  latest_version?: string;
  total_ratings: number;
  overall_score: number;
  created_at: string;
  /** Trust score (0-100) - represents overall trust level */
  trust_score?: number;
  /** Trust tier classification */
  trust_tier?: 'critical' | 'high' | 'medium' | 'low' | 'untrusted';
  /** Verification status */
  verification_status?: 'verified' | 'pending' | 'unverified';
}

export interface RegistryFunctionVersion {
  id: string;
  version: string;
  manifest: any;
  source_code?: string; // The actual source code
  runtime: string;
  timeout_ms: number;
  memory_mb: number;
  deterministic: boolean;
  cache_ttl: number;
  published_at: string;
}

export interface RegistrySearchParams {
  query?: string;
  category?: string;
  author?: string;
  visibility?: string;
  tags?: string[];
  limit?: number;
  offset?: number;
}

export interface RegistryExecutionRequest {
  input: any;
  version?: string;
}

export interface RegistryExecutionResponse {
  ok: boolean;
  data: any;
  cached: boolean;
  duration_ms: number;
  version: string;
  execution_id?: string;
}

export interface RegistryRatingRequest {
  overall_score: number;
  reliability_score: number;
  latency_score: number;
  documentation_score: number;
}

export interface RegistryFunctionReview {
  id: string;
  function_id: string;
  user_id: string;
  stars: number; // 1-5
  title: string;
  body: string;
  created_at: string;
  updated_at: string;
}

class RegistryApi {
  // Get list of functions
  async getFunctions(params?: RegistrySearchParams) {
    const queryParams = new URLSearchParams();
    if (params?.query) queryParams.append('q', params.query);
    if (params?.category) queryParams.append('category', params.category);
    if (params?.author) queryParams.append('author', params.author);
    if (params?.visibility) queryParams.append('visibility', params.visibility);
    if (params?.tags) params.tags.forEach((tag) => queryParams.append('tags', tag));
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());

    const url = `/v2/functions${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    return apiClient.get<{ functions: RegistryFunction[] }>(url);
  }

  // Get function details
  async getFunction(author: string, name: string) {
    return apiClient.get<{ function: RegistryFunction; versions: RegistryFunctionVersion[] }>(
      `/v1/functions/${author}/${name}`
    );
  }

  // Get function versions
  async getFunctionVersions(author: string, name: string) {
    return apiClient.get<{ versions: RegistryFunctionVersion[] }>(
      `/v1/functions/${author}/${name}/versions`
    );
  }

  // Get function version source code
  async getFunctionVersionSource(author: string, name: string, version: string = 'latest') {
    const response = await apiClient.get<{ source_code: string; version: string }>(
      `/v1/functions/${author}/${name}/source${version !== 'latest' ? `?version=${version}` : ''}`
    );
    console.log('[getFunctionVersionSource] Response:', response);
    if (!response?.source_code) {
      throw new Error('Source code not available for this function version');
    }
    return response.source_code;
  }

  // Search functions
  async searchFunctions(query: string, category?: string, limit = 50) {
    const params = new URLSearchParams({ q: query, limit: limit.toString() });
    if (category) params.append('category', category);

    return apiClient.get<{ functions: RegistryFunction[] }>(
      `/v1/registry/search?${params.toString()}`
    );
  }

  // Execute function
  async executeFunction(author: string, name: string, request: RegistryExecutionRequest) {
    return apiClient.post<RegistryExecutionResponse>(`/v1/fx/${author}/${name}`, request);
  }

  // Execute function with specific version
  async executeFunctionVersion(
    author: string,
    name: string,
    version: string,
    request: RegistryExecutionRequest
  ) {
    return apiClient.post<RegistryExecutionResponse>(
      `/v1/fx/${author}/${name}@${version}`,
      request
    );
  }

  // Get function stats
  async getFunctionStats(author: string, name: string) {
    return apiClient.get<{
      function_id: string;
      author: string;
      name: string;
      total_calls: number;
      success_rate: number;
      avg_latency_ms: number;
      p95_latency_ms: number;
      reliability_score: number;
      latency_score: number;
      overall_score: number;
      total_ratings: number;
      popularity_score: number;
    }>(`/v1/functions/${author}/${name}/stats`);
  }

  // Submit rating
  async submitRating(author: string, name: string, rating: RegistryRatingRequest) {
    return apiClient.post<{ ok: boolean; message: string }>(
      `/v1/functions/${author}/${name}/rating`,
      rating
    );
  }

  // List reviews (public)
  async listReviews(author: string, name: string, params?: { limit?: number; offset?: number }) {
    const qp = new URLSearchParams();
    if (params?.limit != null) qp.append('limit', String(params.limit));
    if (params?.offset != null) qp.append('offset', String(params.offset));
    const suffix = qp.toString() ? `?${qp.toString()}` : '';
    return apiClient.get<{
      ok: boolean;
      reviews: RegistryFunctionReview[];
      total: number;
      limit: number;
      offset: number;
    }>(`/v1/functions/${author}/${name}/reviews${suffix}`);
  }

  // Submit review (requires auth)
  async submitReview(
    author: string,
    name: string,
    data: { stars: number; title?: string; body?: string }
  ) {
    return apiClient.post<{ ok: boolean; review: RegistryFunctionReview }>(
      `/v1/functions/${author}/${name}/reviews`,
      {
        stars: data.stars,
        title: data.title ?? '',
        body: data.body ?? '',
      }
    );
  }

  // Test function
  async testFunction(author: string, name: string, input: any) {
    return apiClient.post<{ ok: boolean; output: any; duration_ms: number }>(
      `/v1/functions/${author}/${name}/test`,
      { input }
    );
  }

  // Publish function (requires authentication)
  async publishFunction(publishRequest: {
    author: string;
    name: string;
    version: string;
    manifest: any;
    source: { code: string; language?: string };
    readme?: string;
  }) {
    return apiClient.post<{ ok: boolean; function_id: string; version_id: string }>(
      '/v1/registry/publish',
      publishRequest
    );
  }

  // Get replay (public executions)
  async getReplay(executionId: string) {
    return apiClient.get<{
      execution_id: string;
      function: { author: string; name: string; version: string };
      input: any;
      output: any;
      duration_ms: number;
      executed_at: string;
    }>(`/v1/registry/replay/${executionId}`);
  }

  /** GET /v1/functions/{author}/{name}/settings — requires auth, returns function settings (e.g. custom domains). */
  async getFunctionSettings(author: string, name: string) {
    return apiClient.get<FunctionSettingsResponse>(`/v1/functions/${author}/${name}/settings`);
  }

  /** PATCH /v1/functions/{author}/{name}/settings — requires auth, updates settings (e.g. customDomains). */
  async patchFunctionSettings(author: string, name: string, data: { customDomains?: string[] }) {
    return apiClient.patch<FunctionSettingsResponse>(
      `/v1/functions/${author}/${name}/settings`,
      data
    );
  }
}

/** Response shape for GET /v1/functions/{author}/{name}/settings */
export interface FunctionSettingsResponse {
  id: string;
  name: string;
  author: string;
  description: string;
  isPublic: boolean;
  isPublished: boolean;
  allowAnonymousInvoke: boolean;
  corsEnabled: boolean;
  corsOrigins: string[];
  timeout: number;
  memory: number;
  runtime: string;
  providers: string[];
  environmentVariables: Record<string, string>;
  secrets: string[];
  customDomains: string[];
}

export const registryApi = new RegistryApi();
