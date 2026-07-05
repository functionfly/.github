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
  /** DNA generation (0 = no DNA data) */
  dna_generation?: number;
  /** DNA fitness score (0-100) */
  dna_fitness_score?: number;
  /** Total DNA mutations */
  dna_total_mutations?: number;
  /** Total executions tracked by DNA */
  dna_total_executions?: number;
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
  username?: string;
  user_name?: string;
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

  // Get registry functions owned by the authenticated user
  async getMyFunctions(params?: { limit?: number; offset?: number }) {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    if (params?.offset) queryParams.append('offset', params.offset.toString());

    const url = `/v2/functions/mine${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
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
    source: { code: string; language?: string; wasmBinary?: string };
    readme?: string;
    conflictStrategy?: 'error' | 'overwrite' | 'create_new';
    changelog?: {
      category: string;
      title: string;
      description: string;
      changes: unknown[];
    };
  }) {
    const params = new URLSearchParams();
    if (publishRequest.conflictStrategy) {
      params.set('conflict_strategy', publishRequest.conflictStrategy);
    }
    const url = `/v1/registry/publish${params.toString() ? `?${params.toString()}` : ''}`;
    return apiClient.post<{ ok: boolean; function_id: string; version_id: string }>(
      url,
      publishRequest
    );
  }

  // Request a presigned URL for direct browser → object-storage upload of a
  // single artifact. The dashboard calls this first when handling large
  // files, then PUTs to presignedUploadResponse.url, then includes the
  // returned key + content_hash in the publish payload.
  async presignArtifactUpload(req: {
    kind: 'wasm' | 'source' | 'readme' | 'code';
    content_hash: string;
    content_type?: string;
    ext?: string;
  }): Promise<{
    key: string;
    url: string;
    method: 'PUT';
    content_type: string;
    max_bytes: number;
    expires_at: string;
  }> {
    return apiClient.post('/v1/registry/publish/presign', req);
  }

  // Upload a Blob to a presigned URL with progress callbacks. Throws if the
  // upload fails or the response status is outside 2xx.
  async uploadToPresignedUrl(
    url: string,
    body: Blob,
    contentType: string,
    onProgress?: (loaded: number, total: number) => void
  ): Promise<void> {
    // XHR gives us upload progress events; fetch() does not.
    if (typeof XMLHttpRequest !== 'undefined' && onProgress) {
      await new Promise<void>((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        xhr.open('PUT', url, true);
        xhr.setRequestHeader('Content-Type', contentType);
        xhr.upload.onprogress = (evt) => {
          if (evt.lengthComputable) onProgress(evt.loaded, evt.total);
        };
        xhr.onload = () => {
          if (xhr.status >= 200 && xhr.status < 300) resolve();
          else reject(new Error(`upload failed: ${xhr.status} ${xhr.statusText}`));
        };
        xhr.onerror = () => reject(new Error('upload network error'));
        xhr.send(body);
      });
      return;
    }
    const res = await fetch(url, { method: 'PUT', body, headers: { 'Content-Type': contentType } });
    if (!res.ok) {
      throw new Error(`upload failed: ${res.status} ${res.statusText}`);
    }
  }

  // Request a short-lived signed download URL for a stored artifact.
  async presignArtifactDownload(key: string, ttlSeconds = 300): Promise<{ url: string; expires_at: string }> {
    return apiClient.post('/v1/artifacts/download', { key, ttl_seconds: ttlSeconds });
  }

  // Higher-level publish flow: for each non-trivial file, presign → upload to
  // R2 → publish with the resulting keys. Falls back to JSON publish when the
  // artifact store is unavailable (server returns 503) or for tiny payloads
  // below the PRESIGN_THRESHOLD.
  async publishFunctionViaPresigned(opts: {
    author: string;
    name: string;
    version: string;
    manifest: any;
    source?: { code: string; language?: string; codeBlob?: Blob };
    readme?: string;
    wasm?: { binaryBase64?: string; blob?: Blob };
    conflictStrategy?: 'error' | 'overwrite' | 'create_new';
    changelog?: {
      category: string;
      title: string;
      description: string;
      changes: unknown[];
    };
    onProgress?: (stage: 'wasm' | 'source' | 'readme' | 'publish', loaded: number, total: number) => void;
  }): Promise<{ ok: boolean; function_id: string; version_id: string }> {
    const PRESIGN_THRESHOLD = 256 * 1024;

    const sourceBlob = opts.source?.codeBlob;
    const readmeBlob = opts.readme ? new Blob([opts.readme], { type: 'text/markdown' }) : undefined;
    const wasmBlob = opts.wasm?.blob;

    const largeSource = sourceBlob && sourceBlob.size > PRESIGN_THRESHOLD;
    const largeWasm = wasmBlob && wasmBlob.size > PRESIGN_THRESHOLD;
    const largeReadme = readmeBlob && readmeBlob.size > PRESIGN_THRESHOLD;

    let sourceKey: string | undefined;
    let sourceHash: string | undefined;
    let wasmKey: string | undefined;
    let wasmHash: string | undefined;
    let readmeKey: string | undefined;

    // Helper: SHA-256 hex via Web Crypto, returns null on unsupported env.
    async function sha256Hex(blob: Blob): Promise<string> {
      if (typeof crypto === 'undefined' || !crypto.subtle) return '';
      const buf = await blob.arrayBuffer();
      const digest = await crypto.subtle.digest('SHA-256', buf);
      const bytes = new Uint8Array(digest);
      let hex = '';
      for (let i = 0; i < bytes.length; i++) {
        hex += bytes[i].toString(16).padStart(2, '0');
      }
      return hex;
    }

    if (largeSource && sourceBlob) {
      sourceHash = await sha256Hex(sourceBlob);
      const presign = await this.presignArtifactUpload({
        kind: 'source',
        content_hash: sourceHash,
        content_type: sourceBlob.type || 'text/plain; charset=utf-8',
      });
      await this.uploadToPresignedUrl(presign.url, sourceBlob, presign.content_type, (l, t) =>
        opts.onProgress?.('source', l, t)
      );
      sourceKey = presign.key;
    }

    if (largeWasm && wasmBlob) {
      wasmHash = await sha256Hex(wasmBlob);
      const presign = await this.presignArtifactUpload({
        kind: 'wasm',
        content_hash: wasmHash,
        content_type: 'application/wasm',
      });
      await this.uploadToPresignedUrl(presign.url, wasmBlob, presign.content_type, (l, t) =>
        opts.onProgress?.('wasm', l, t)
      );
      wasmKey = presign.key;
    }

    if (largeReadme && readmeBlob) {
      const readmeHash = await sha256Hex(readmeBlob);
      const presign = await this.presignArtifactUpload({
        kind: 'readme',
        content_hash: readmeHash,
        content_type: 'text/markdown; charset=utf-8',
      });
      await this.uploadToPresignedUrl(presign.url, readmeBlob, presign.content_type, (l, t) =>
        opts.onProgress?.('readme', l, t)
      );
      readmeKey = presign.key;
    }

    const publishBody: any = {
      author: opts.author,
      name: opts.name,
      version: opts.version,
      manifest: opts.manifest,
      source: sourceBlob
        ? {
            code: largeSource ? '' : (opts.source?.code ?? ''),
            language: opts.source?.language,
            storage_backend: largeSource ? 'r2' : undefined,
            storage_key: sourceKey,
            content_hash: sourceHash,
            presigned_upload_complete: !!sourceKey,
            wasmBinary: opts.wasm?.binaryBase64,
          }
        : undefined,
      readme: opts.readme,
      wasm:
        largeWasm
          ? {
              storage_key: wasmKey,
              content_hash: wasmHash,
              presigned_upload_complete: true,
            }
          : undefined,
      readme_storage: largeReadme
        ? { storage_key: readmeKey, presigned_upload_complete: true }
        : undefined,
      changelog: opts.changelog,
    };

    return this.publishFunction(publishBody);
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
