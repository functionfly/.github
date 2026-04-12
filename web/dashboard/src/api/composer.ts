import { apiClient } from './client';
import { z } from 'zod';

// Validation schemas
export const functionGenerationRequestSchema = z.object({
  description: z.string().min(10).max(2000),
  runtime: z.string().default('python'),
  inputs: z.array(z.any()).optional(),
  outputs: z.array(z.any()).optional(),
  constraints: z.string().optional(),
  examples: z.array(z.string()).optional(),
});

export const functionManifestSchema = z.object({
  name: z.string(),
  description: z.string(),
  version: z.string().default('1.0.0'),
  inputs: z.array(z.any()).default([]),
  outputs: z.array(z.any()).default([]),
  runtime: z.string(),
  timeout_seconds: z.number().default(30),
  memory_mb: z.number().default(256),
  capabilities: z.array(z.string()).default([]),
});

export const functionGenerationResultSchema = z.object({
  code: z.string(),
  runtime: z.string(),
  manifest: functionManifestSchema,
  explanation: z.string(),
  suggested_tests: z.array(z.string()).default([]),
  estimated_complexity: z.enum(['simple', 'moderate', 'complex']),
});

export const functionGenerationResponseSchema = z.object({
  success: z.boolean(),
  result: functionGenerationResultSchema.optional(),
  error: z.string().optional(),
  generation_id: z.string(),
  latency_ms: z.number(),
  tokens_used: z.object({
    prompt: z.number().default(0),
    completion: z.number().default(0),
    total: z.number().default(0),
  }).default({ prompt: 0, completion: 0, total: 0 }),
});

export const galleryFunctionSchema = z.object({
  id: z.string(),
  author: z.string(),
  name: z.string(),
  title: z.string(),
  description: z.string(),
  category: z.string().nullable().optional(),
  runtime: z.string(),
  trust_score: z.number().min(0).max(100),
  popularity_score: z.number().default(0),
  remix_count: z.number().default(0),
  like_count: z.number().default(0),
  created_at: z.string(),
  updated_at: z.string(),
});

export const gallerySearchResponseSchema = z.object({
  query: z.string(),
  results: z.array(galleryFunctionSchema),
  total_count: z.number(),
  limit: z.number(),
  offset: z.number(),
});

export const remixResponseSchema = z.object({
  success: z.boolean(),
  message: z.string(),
  remix_id: z.string(),
  new_function_id: z.string().optional(),
  new_author: z.string().optional(),
  new_name: z.string().optional(),
  cost_usd: z.number().optional(),
  new_balance_usd: z.number().optional(),
});

// Types
export type FunctionGenerationRequest = z.infer<typeof functionGenerationRequestSchema>;
export type FunctionGenerationResult = z.infer<typeof functionGenerationResultSchema>;
export type FunctionGenerationResponse = z.infer<typeof functionGenerationResponseSchema>;
export type GalleryFunction = z.infer<typeof galleryFunctionSchema>;
export type GallerySearchResponse = z.infer<typeof gallerySearchResponseSchema>;
export type RemixResponse = z.infer<typeof remixResponseSchema>;

export const composerApi = {
  // Generate a function using AI
  generateFunction: async (data: FunctionGenerationRequest): Promise<FunctionGenerationResponse> => {
    const response = await apiClient.post('/v1/ai/composer/generate', data);
    return functionGenerationResponseSchema.parse(response);
  },

  // Stream function generation (returns EventSource)
  generateFunctionStream: (data: FunctionGenerationRequest): EventSource => {
    const queryParams = new URLSearchParams({
      description: data.description,
      runtime: data.runtime || 'python',
    });
    if (data.constraints) {
      queryParams.append('constraints', data.constraints);
    }
    // EventSource doesn't support headers, so we include a token in the URL for auth
    const token = localStorage.getItem('ff-access-token');
    if (token) {
      queryParams.append('_token', token);
    }
    return new EventSource(
      `${apiClient.getBaseUrl()}/v1/ai/composer/generate/stream?${queryParams.toString()}`
    );
  },
};

export const galleryApi = {
  // Search gallery functions
  search: async (params: {
    query?: string;
    category?: string;
    runtime?: string;
    sort_by?: 'popular' | 'recent' | 'rating' | 'name';
    limit?: number;
    offset?: number;
  }): Promise<GallerySearchResponse> => {
    const searchParams = new URLSearchParams();
    if (params.query) searchParams.append('q', params.query);
    if (params.category) searchParams.append('category', params.category);
    if (params.runtime) searchParams.append('runtime', params.runtime);
    if (params.sort_by) searchParams.append('sort_by', params.sort_by);
    if (params.limit) searchParams.append('limit', params.limit.toString());
    if (params.offset) searchParams.append('offset', params.offset.toString());

    const response = await apiClient.get(`/v1/registry/search?${searchParams.toString()}`);
    return gallerySearchResponseSchema.parse(response);
  },

  // Get trending functions
  getTrending: async (limit = 20): Promise<GalleryFunction[]> => {
    const response = await apiClient.get(`/v1/registry/trending?limit=${limit}`);
    // Handle different response formats - the API might return { functions: [] } or directly []
    const functionList = (response && typeof response === 'object' && 'functions' in response)
      ? (response as { functions: unknown[] }).functions
      : (Array.isArray(response) ? response : []);
    return z.array(galleryFunctionSchema).parse(functionList || []);
  },

  // Get function details
  getFunction: async (author: string, name: string): Promise<GalleryFunction> => {
    const response = await apiClient.get(`/v1/registry/functions/${author}/${name}`);
    return galleryFunctionSchema.parse(response);
  },

  // Remix/fork a function
  remix: async (
    author: string,
    name: string,
    data: {
      new_name?: string;
      customization?: string;
      private_function?: boolean;
    }
  ): Promise<RemixResponse> => {
    const response = await apiClient.post(`/v1/registry/functions/${author}/${name}/remix`, data);
    return remixResponseSchema.parse(response);
  },

  // Like a function
  like: async (author: string, name: string): Promise<{ liked: boolean; like_count: number }> => {
    const response = await apiClient.post(`/v1/registry/functions/${author}/${name}/likes`);
    return z.object({ liked: z.boolean(), like_count: z.number() }).parse(response);
  },

  // Get likes for a function
  getLikes: async (author: string, name: string): Promise<{ like_count: number; liked_by_user: boolean }> => {
    const response = await apiClient.get(`/v1/registry/functions/${author}/${name}/likes`);
    return z.object({ like_count: z.number(), liked_by_user: z.boolean() }).parse(response);
  },

  // Get remix history
  getRemixHistory: async (author: string, name: string): Promise<{
    remix_count: number;
    is_remix: boolean;
    remix_history: Array<{
      remix_id: string;
      source_author: string;
      source_name: string;
      remixed_at: string;
      customization: string;
    }>;
  }> => {
    const response = await apiClient.get(`/v1/registry/functions/${author}/${name}/remix/history`);
    return z.object({
      remix_count: z.number(),
      is_remix: z.boolean(),
      remix_history: z.array(z.object({
        remix_id: z.string(),
        source_author: z.string(),
        source_name: z.string(),
        remixed_at: z.string(),
        customization: z.string(),
      })),
    }).parse(response);
  },

  // Get remix cost and wallet balance
  getRemixCost: async (author: string, name: string): Promise<{
    cost_usd: number;
    balance_usd: number;
    can_remix: boolean;
    required_usd: number;
    function_author: string;
    function_name: string;
    is_own_function: boolean;
  }> => {
    const response = await apiClient.get(`/v1/registry/functions/${author}/${name}/remix/cost`);
    return z.object({
      cost_usd: z.number(),
      balance_usd: z.number(),
      can_remix: z.boolean(),
      required_usd: z.number(),
      function_author: z.string(),
      function_name: z.string(),
      is_own_function: z.boolean(),
    }).parse(response);
  },
};
