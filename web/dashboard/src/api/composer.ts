import { apiClient } from './client';
import { tokenVault } from '@/utils/token-vault';
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

// Refinement request schema
export const functionRefinementRequestSchema = z.object({
  generation_id: z.string(),
  modification_request: z.string().min(5).max(1000),
  preserve_structure: z.boolean().default(true),
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

// Refinement response schema (same structure as generation)
export const functionRefinementResponseSchema = functionGenerationResponseSchema.extend({
  parent_generation_id: z.string(),
  refinement_id: z.string(),
});

export const galleryFunctionSchema = z.object({
  id: z.string().default(''),
  author: z.string().default(''),
  name: z.string().default(''),
  title: z.string().optional(),
  description: z.string().optional(),
  category: z.string().nullable().optional(),
  runtime: z.string().default(''),
  trust_score: z.number().min(0).max(100).optional().default(0),
  popularity_score: z.number().optional().default(0),
  remix_count: z.number().optional().default(0),
  like_count: z.number().optional().default(0),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
  // Additional fields from backend
  version: z.string().optional(),
  tags: z.array(z.string()).optional(),
  trust_level: z.string().optional(),
  success_rate: z.number().optional(),
  reliability: z.number().optional(),
}).passthrough(); // Allow additional fields from backend without validation errors

// Helper to parse array of functions with individual item fallback
const parseFunctionsArray = (data: unknown): z.infer<typeof galleryFunctionSchema>[] => {
  if (!Array.isArray(data)) return [];
  
  return data.filter((item): item is z.infer<typeof galleryFunctionSchema> => {
    // Basic check - item must be object with at least id
    if (typeof item !== 'object' || item === null) return false;
    return 'id' in item && 'name' in item && 'author' in item;
  }).map(item => {
    // Apply defaults for missing fields
    return {
      id: String((item as Record<string, unknown>).id || ''),
      author: String((item as Record<string, unknown>).author || ''),
      name: String((item as Record<string, unknown>).name || ''),
      title: (item as Record<string, unknown>).title as string | undefined,
      description: (item as Record<string, unknown>).description as string | undefined,
      category: (item as Record<string, unknown>).category as string | null | undefined,
      runtime: String((item as Record<string, unknown>).runtime || ''),
      trust_score: Number((item as Record<string, unknown>).trust_score) || 0,
      popularity_score: Number((item as Record<string, unknown>).popularity_score) || 0,
      remix_count: Number((item as Record<string, unknown>).remix_count) || 0,
      like_count: Number((item as Record<string, unknown>).like_count) || 0,
      created_at: (item as Record<string, unknown>).created_at as string | undefined,
      updated_at: (item as Record<string, unknown>).updated_at as string | undefined,
      version: (item as Record<string, unknown>).version as string | undefined,
      tags: (item as Record<string, unknown>).tags as string[] | undefined,
      trust_level: (item as Record<string, unknown>).trust_level as string | undefined,
      success_rate: (item as Record<string, unknown>).success_rate as number | undefined,
      reliability: (item as Record<string, unknown>).reliability as number | undefined,
    };
  });
};

export const gallerySearchResponseSchema = z.object({
  // Backend returns these fields
  Functions: z.array(z.any()).optional(),
  functions: z.array(z.any()).optional(),
  Total: z.number().optional(),
  total: z.number().optional(),
  total_count: z.number().optional(),
  Limit: z.number().optional(),
  limit: z.number().optional(),
  Offset: z.number().optional(),
  offset: z.number().optional(),
  // Legacy fields that might still be returned
  query: z.string().optional().default(''),
  results: z.array(z.any()).optional(),
}).transform((data) => ({
  // Normalize to standard fields with safe parsing
  results: parseFunctionsArray(data.results || data.functions || data.Functions || []),
  total_count: data.total_count || data.total || data.Total || 0,
  limit: data.limit || data.Limit || 50,
  offset: data.offset || data.Offset || 0,
  query: data.query || '',
}));

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
export type FunctionRefinementRequest = z.infer<typeof functionRefinementRequestSchema>;
export type FunctionGenerationResult = z.infer<typeof functionGenerationResultSchema>;
export type FunctionGenerationResponse = z.infer<typeof functionGenerationResponseSchema>;
export type FunctionRefinementResponse = z.infer<typeof functionRefinementResponseSchema>;
export type GalleryFunction = z.infer<typeof galleryFunctionSchema>;
export type GallerySearchResponse = z.infer<typeof gallerySearchResponseSchema>;
export type RemixResponse = z.infer<typeof remixResponseSchema>;

// Runtime to Monaco language mapping
export const RUNTIME_MONACO_LANG: Record<string, string> = {
  python: 'python',
  'python-light': 'python',
  nodejs: 'javascript',
  typescript: 'typescript',
  go: 'go',
  rust: 'rust',
  deno: 'typescript',
  bun: 'typescript',
};

export const composerApi = {
  // Generate a function using AI
  generateFunction: async (data: FunctionGenerationRequest): Promise<FunctionGenerationResponse> => {
    const response = await apiClient.post('/v1/ai/composer/generate', data);
    return functionGenerationResponseSchema.parse(response);
  },

  // Stream function generation (returns EventSource)
  generateFunctionStream: async (data: FunctionGenerationRequest): Promise<EventSource> => {
    const queryParams = new URLSearchParams({
      description: data.description,
      runtime: data.runtime || 'python',
    });
    if (data.constraints) {
      queryParams.append('constraints', data.constraints);
    }
    // EventSource doesn't support headers, so we include a token in the URL for auth
    await tokenVault.initialize();
    const token = await tokenVault.getAccessToken();
    if (token) {
      queryParams.append('_token', token);
    }
    return new EventSource(
      `${apiClient.getBaseUrl()}/v1/ai/composer/generate/stream?${queryParams.toString()}`
    );
  },

  // Refine/improve a previously generated function
  refineFunction: async (data: FunctionRefinementRequest): Promise<FunctionRefinementResponse> => {
    const response = await apiClient.post('/v1/ai/composer/refine', data);
    return functionRefinementResponseSchema.parse(response);
  },

  // Stream refinement (returns EventSource)
  refineFunctionStream: async (data: FunctionRefinementRequest): Promise<EventSource> => {
    const queryParams = new URLSearchParams({
      generation_id: data.generation_id,
      modification_request: data.modification_request,
      preserve_structure: data.preserve_structure?.toString() ?? 'true',
    });
    // EventSource doesn't support headers, so we include a token in the URL for auth
    await tokenVault.initialize();
    const token = await tokenVault.getAccessToken();
    if (token) {
      queryParams.append('_token', token);
    }
    return new EventSource(
      `${apiClient.getBaseUrl()}/v1/ai/composer/refine/stream?${queryParams.toString()}`
    );
  },

  // Get generation history for a session
  getGenerationHistory: async (sessionId: string): Promise<{
    generations: Array<{
      generation_id: string;
      description: string;
      runtime: string;
      created_at: string;
      parent_generation_id?: string;
    }>;
  }> => {
    const response = await apiClient.get(`/v1/ai/composer/history?session=${sessionId}`);
    return z.object({
      generations: z.array(z.object({
        generation_id: z.string(),
        description: z.string(),
        runtime: z.string(),
        created_at: z.string(),
        parent_generation_id: z.string().optional(),
      })),
    }).parse(response);
  },
};

// Future AI Expansion API - Placeholder for upcoming features
export const aiExpansionApi = {
  // Get status of all AI namespace features
  getStatus: async (): Promise<{
    namespace: string;
    version: string;
    features: Record<string, {
      path: string;
      status: string;
      description: string;
      estimated_release?: string;
      endpoints: string[];
    }>;
    message: string;
  }> => {
    const response = await apiClient.get('/v1/ai/status');
    return z.object({
      namespace: z.string(),
      version: z.string(),
      features: z.record(z.string(), z.object({
        path: z.string(),
        status: z.string(),
        description: z.string(),
        estimated_release: z.string().optional(),
        endpoints: z.array(z.string()),
      })),
      message: z.string(),
    }).parse(response) as { namespace: string; version: string; features: Record<string, { path: string; status: string; description: string; estimated_release?: string; endpoints: string[]; }>; message: string; };
  },

  // AI Chat - RESERVED FOR FUTURE (Q3 2026)
  chat: {
    // Placeholder for conversational AI interface
    sendMessage: async (message: string): Promise<{
      feature: string;
      status: string;
      message: string;
      alternative: string;
    }> => {
      const response = await apiClient.post('/v1/ai/chat/message', { message });
      return z.object({
        feature: z.string(),
        status: z.string(),
        message: z.string(),
        alternative: z.string(),
      }).parse(response);
    },
  },

  // AI Suggest - RESERVED FOR FUTURE (Q4 2026)
  suggest: {
    // Placeholder for code suggestions/completions
    getCompletions: async (codeContext: string): Promise<{
      feature: string;
      status: string;
      message: string;
    }> => {
      const response = await apiClient.post('/v1/ai/suggest/completions', { code_context: codeContext });
      return z.object({
        feature: z.string(),
        status: z.string(),
        message: z.string(),
      }).parse(response);
    },

    // Placeholder for fix suggestions
    getFixes: async (errorMessage: string): Promise<{
      feature: string;
      status: string;
      message: string;
      alternative: string;
    }> => {
      const response = await apiClient.post('/v1/ai/suggest/fixes', { error_message: errorMessage });
      return z.object({
        feature: z.string(),
        status: z.string(),
        message: z.string(),
        alternative: z.string(),
      }).parse(response);
    },

    // Get suggestion service status
    getStatus: async (): Promise<{
      service: string;
      status: string;
      features: string[];
      estimated_release: string;
      message: string;
    }> => {
      const response = await apiClient.get('/v1/ai/suggest/status');
      return z.object({
        service: z.string(),
        status: z.string(),
        features: z.array(z.string()),
        estimated_release: z.string(),
        message: z.string(),
      }).parse(response);
    },
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
    console.log('[Gallery API] Raw response:', response);
    
    // Schema handles transformation internally, but we need to ensure
    // the raw response has the expected shape before parsing
    const responseObj = response as Record<string, unknown>;
    const responseWithDefaults = {
      ...responseObj,
      query: params.query || '',
    };
    
    // Use safeParse to handle validation errors gracefully
    const parseResult = gallerySearchResponseSchema.safeParse(responseWithDefaults);
    
    if (!parseResult.success) {
      console.error('[Gallery API] Validation error:', parseResult.error.issues);
      // Try to extract functions anyway from raw response using safe parsing
      const rawFunctions = (responseObj.functions || responseObj.Functions || []) as unknown[];
      const fallbackFunctions = parseFunctionsArray(rawFunctions);
      console.log('[Gallery API] Using fallback with', fallbackFunctions.length, 'functions (raw had', rawFunctions.length, ')');
      return {
        results: fallbackFunctions,
        total_count: (responseObj.total || responseObj.Total || fallbackFunctions.length) as number,
        limit: params.limit || 50,
        offset: params.offset || 0,
        query: params.query || '',
      };
    }
    
    console.log('[Gallery API] Parsed:', parseResult.data);
    return parseResult.data;
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
    const response = await apiClient.get(`/v1/functions/${author}/${name}`);
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
    const response = await apiClient.post(`/v1/functions/${author}/${name}/remix`, data);
    return remixResponseSchema.parse(response);
  },

  // Like a function
  like: async (author: string, name: string): Promise<{ liked: boolean; like_count: number }> => {
    const response = await apiClient.post(`/v1/functions/${author}/${name}/likes`);
    return z.object({ liked: z.boolean(), like_count: z.number() }).parse(response);
  },

  // Get likes for a function
  getLikes: async (author: string, name: string): Promise<{ like_count: number; liked_by_user: boolean }> => {
    const response = await apiClient.get(`/v1/functions/${author}/${name}/likes`);
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
    const response = await apiClient.get(`/v1/functions/${author}/${name}/remix/history`);
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
    const response = await apiClient.get(`/v1/functions/${author}/${name}/remix/cost`);
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
