import type {
  CreateFunctionRequest,
  DeployFunctionRequest,
  DeployFunctionResponse,
  FunctionConfig,
  FunctionDeployment,
  FunctionLog,
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
  // List all functions for the current tenant
  list: async (): Promise<{ functions: FunctionConfig[] }> => {
    const response = await apiClient.get('/v1/functions');
    // Validate the response structure
    if (
      response &&
      typeof response === 'object' &&
      'functions' in response &&
      Array.isArray((response as any).functions)
    ) {
      return {
        functions: validateFunctionConfigs((response as any).functions) as FunctionConfig[],
      };
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

  // Get function trust score
  getTrustScore: async (functionId: string): Promise<{ trustScore: number }> => {
    const response = await apiClient.get(`/v1/functions/${functionId}/trust`);
    // Trust score from API is 0-100 scale
    if (
      response &&
      typeof response === 'object' &&
      'trust_score' in response &&
      typeof (response as any).trust_score === 'number'
    ) {
      return { trustScore: (response as any).trust_score };
    }
    return { trustScore: 0 };
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
