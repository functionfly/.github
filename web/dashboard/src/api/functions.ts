import type {
  CreateFunctionRequest,
  DeployFunctionRequest,
  DeployFunctionResponse,
  FunctionConfig,
  FunctionDeployment,
  FunctionLog,
  FunctionTrustScore,
  TestFunctionRequest,
  TestFunctionResponse,
  UpdateFunctionRequest,
} from '@/types';
import {
  createFunctionRequestSchema,
  deployFunctionRequestSchema,
  deployFunctionResponseSchema,
  functionConfigSchema,
  testFunctionRequestSchema,
  testFunctionResponseSchema,
  updateFunctionRequestSchema,
} from '../lib/api-validation';
import {
  validateFunctionConfigs,
  validateFunctionDeployments,
  validateFunctionLogs,
} from '../lib/validation-utils';
import { apiClient } from './client';

export const functionsApi = {
  list: async (): Promise<{ functions: FunctionConfig[] }> => {
    const response = await apiClient.get('/v2/functions/mine');
    if (
      response &&
      typeof response === 'object' &&
      'functions' in response &&
      Array.isArray((response as any).functions)
    ) {
      const funcs = (response as any).functions.map((f: any) => {
        if (!f || typeof f !== 'object') return null;
        return {
          id: f.id,
          name: f.name,
          version: f.version || f.latest_version?.split('@')[0] || '1.0.0',
          status: f.version ? 'deployed' as const : 'draft' as const,
          providers: [],
          region: f.region || 'us-east-1',
          code: f.code || '',
          envVars: [],
          tenantId: f.tenant_id || f.tenantId || '',
          createdAt: f.created_at || f.createdAt || new Date().toISOString(),
          updatedAt: f.updated_at || f.updatedAt || new Date().toISOString(),
          trustScore: f.trust_score || f.trustScore,
        };
      }).filter(Boolean);
      return { functions: funcs };
    }
    return { functions: [] };
  },

  // Get a specific function
  get: (functionId: string): Promise<FunctionConfig> =>
    apiClient.getValidatedData<FunctionConfig>(functionConfigSchema, `/v1/functions/${functionId}`),

  // Create a new function
  create: async (data: CreateFunctionRequest): Promise<FunctionConfig> => {
    // Validate input data
    const validation = createFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid function creation data: ${validation.error.message}`);
    }
    const payload = {
      name: data.name,
      code: data.code,
      providers: data.providers,
      region: data.region,
      env_vars:
        data.envVars?.map((e) => ({
          key: e.key,
          value: e.value,
          is_secret: e.isSecret,
        })) ?? [],
    };
    return apiClient.postValidatedData<FunctionConfig>(
      functionConfigSchema,
      '/v1/functions',
      payload
    );
  },

  // Update an existing function
  update: async (functionId: string, data: UpdateFunctionRequest): Promise<FunctionConfig> => {
    // Validate input data
    const validation = updateFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid function update data: ${validation.error.message}`);
    }
    const payload: Record<string, unknown> = {};
    if (data.name !== undefined) payload.name = data.name;
    if (data.code !== undefined) payload.code = data.code;
    if (data.providers !== undefined) payload.providers = data.providers;
    if (data.region !== undefined) payload.region = data.region;
    if (data.envVars !== undefined) {
      payload.env_vars = data.envVars.map((e) => ({
        key: e.key,
        value: e.value,
        is_secret: e.isSecret,
      }));
    }
    return apiClient.putValidatedData<FunctionConfig>(
      functionConfigSchema,
      `/v1/functions/${functionId}`,
      payload
    );
  },

  // Delete a function
  delete: (functionId: string) => apiClient.delete(`/v1/functions/${functionId}`),

  // Deploy a function
  deploy: async (data: DeployFunctionRequest): Promise<DeployFunctionResponse> => {
    // Validate input data
    const validation = deployFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid deployment data: ${validation.error.message}`);
    }
    const payload: Record<string, unknown> = {
      function_id: data.functionId,
      backend_id: data.backendId,
    };
    if (data.version) payload.version = data.version;
    if (data.environment) payload.environment = data.environment;
    return apiClient.postValidatedData<DeployFunctionResponse>(
      deployFunctionResponseSchema,
      '/v1/functions/deploy',
      payload
    );
  },

  // Test a function
  test: async (data: TestFunctionRequest): Promise<TestFunctionResponse> => {
    // Validate input data
    const validation = testFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid test data: ${validation.error.message}`);
    }
    return apiClient.postValidatedData<TestFunctionResponse>(
      testFunctionResponseSchema,
      '/v1/functions/test',
      data
    );
  },

  // Get function deployments
  getDeployments: async (functionId: string): Promise<{ deployments: FunctionDeployment[] }> => {
    const response = await apiClient.get(`/v1/functions/${functionId}/deployments`);
    // Validate the response structure
    if (
      response &&
      typeof response === 'object' &&
      'deployments' in response &&
      Array.isArray((response as any).deployments)
    ) {
      return {
        deployments: validateFunctionDeployments(
          (response as any).deployments
        ) as FunctionDeployment[],
      };
    }
    return { deployments: [] };
  },

  // Get function logs
  getLogs: async (
    functionId: string,
    params?: { limit?: number; since?: string; level?: string }
  ): Promise<{ logs: FunctionLog[] }> => {
    const response = await apiClient.get(`/v1/functions/${functionId}/logs`, { params });
    // Validate the response structure
    if (
      response &&
      typeof response === 'object' &&
      'logs' in response &&
      Array.isArray((response as any).logs)
    ) {
      return { logs: validateFunctionLogs((response as any).logs) as FunctionLog[] };
    }
    return { logs: [] };
  },

  // Get deployment logs
  getDeploymentLogs: async (deploymentId: string): Promise<{ logs: FunctionLog[] }> => {
    const response = await apiClient.get(`/v1/deployments/${deploymentId}/logs`);
    // Validate the response structure
    if (
      response &&
      typeof response === 'object' &&
      'logs' in response &&
      Array.isArray((response as any).logs)
    ) {
      return { logs: validateFunctionLogs((response as any).logs) as FunctionLog[] };
    }
    return { logs: [] };
  },

  // Get function analytics/metrics
  getMetrics: (functionId: string, params?: { period?: string; from?: string; to?: string }) =>
    apiClient.get(`/v1/functions/${functionId}/metrics`, { params }),

  // Get function trust score with full component breakdown
  getTrustScore: async (functionId: string): Promise<FunctionTrustScore> => {
    const response = await apiClient.get(`/v1/functions/${functionId}/trust`);
    if (response && typeof response === 'object' && 'trust_score' in response) {
      const r = response as any;
      return {
        trustScore: typeof r.trust_score === 'number' ? r.trust_score : 0,
        trustTier: r.trust_tier ?? 'untrusted',
        isVerified: r.is_verified ?? false,
        verificationLevel: r.verification_level ?? 'none',
        components: {
          reliability: r.components?.reliability ?? 0,
          latency: r.components?.latency ?? 0,
          errorRate: r.components?.error_rate ?? 0,
          userRating: r.components?.user_rating ?? 0,
          verification: r.components?.verification ?? 0,
        },
        metrics: {
          totalCalls: r.metrics?.total_calls ?? 0,
          successRate: r.metrics?.success_rate ?? 0,
          p50LatencyMs: r.metrics?.p50_latency_ms ?? 0,
          p95LatencyMs: r.metrics?.p95_latency_ms ?? 0,
          p99LatencyMs: r.metrics?.p99_latency_ms ?? 0,
          errorRate: r.metrics?.error_rate ?? 0,
          timeoutRate: r.metrics?.timeout_rate ?? 0,
        },
      };
    }
    return { trustScore: 0, trustTier: 'untrusted', isVerified: false, verificationLevel: 'none', components: { reliability: 0, latency: 0, errorRate: 0, userRating: 0, verification: 0 }, metrics: { totalCalls: 0, successRate: 0, p50LatencyMs: 0, p95LatencyMs: 0, p99LatencyMs: 0, errorRate: 0, timeoutRate: 0 } };
  },

  // Parse code to extract functions
  parseCode: async (code: string, forceLanguage?: string): Promise<{
    language: string;
    confidence: number;
    functions: Array<{
      id: string;
      name: string;
      language: string;
      signature: string;
      parameters: Array<{ name: string; type?: string; has_default: boolean }>;
      return_type?: string;
      docstring?: string;
      code: string;
      start_line: number;
      end_line: number;
    }>;
    raw_code_length: number;
  }> => {
    return apiClient.post('/v1/functions/parse', {
      code,
      force_language: forceLanguage,
    });
  },

  // Create functions from parsed code
  createFromCode: async (data: {
    functions: Array<{ name: string; code: string; language: string }>;
    visibility: 'private' | 'public';
    providers?: string[];
    region?: string;
  }): Promise<{
    created: Array<{ id: string; name: string; status: string }>;
    failed?: Array<{ name: string; error: string }>;
  }> => {
    return apiClient.post('/v1/functions/from-code', data);
  },
};
