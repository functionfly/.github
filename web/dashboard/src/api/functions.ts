import { apiClient } from "./client";
import type {
  FunctionConfig,
  CreateFunctionRequest,
  UpdateFunctionRequest,
  FunctionDeployment,
  FunctionLog,
  DeployFunctionRequest,
  DeployFunctionResponse,
  TestFunctionRequest,
  TestFunctionResponse
} from "@/types";
import {
  functionConfigSchema,
  createFunctionRequestSchema,
  updateFunctionRequestSchema,
  functionDeploymentSchema,
  functionLogSchema,
  deployFunctionRequestSchema,
  deployFunctionResponseSchema,
  testFunctionRequestSchema,
  testFunctionResponseSchema
} from "../lib/api-validation";
import {
  validateFunctionConfigs,
  validateFunctionDeployments,
  validateFunctionLogs
} from "../lib/validation-utils";

export const functionsApi = {
  // List all functions for the current tenant
  list: async (): Promise<{ functions: FunctionConfig[] }> => {
    const response = await apiClient.get("/v1/functions");
    // Validate the response structure
    if (response && typeof response === 'object' && 'functions' in response && Array.isArray((response as any).functions)) {
      return { functions: validateFunctionConfigs((response as any).functions) as FunctionConfig[] };
    }
    return { functions: [] };
  },

  // Get a specific function
  get: (functionId: string) => apiClient.getValidatedData(functionConfigSchema, `/v1/functions/${functionId}`),

  // Create a new function
  create: async (data: CreateFunctionRequest) => {
    // Validate input data
    const validation = createFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid function creation data: ${validation.error.message}`);
    }
    return apiClient.postValidatedData(functionConfigSchema, "/v1/functions", data);
  },

  // Update an existing function
  update: async (functionId: string, data: UpdateFunctionRequest) => {
    // Validate input data
    const validation = updateFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid function update data: ${validation.error.message}`);
    }
    return apiClient.putValidatedData(functionConfigSchema, `/v1/functions/${functionId}`, data);
  },

  // Delete a function
  delete: (functionId: string) =>
    apiClient.delete(`/v1/functions/${functionId}`),

  // Deploy a function
  deploy: async (data: DeployFunctionRequest) => {
    // Validate input data
    const validation = deployFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid deployment data: ${validation.error.message}`);
    }
    return apiClient.postValidatedData(deployFunctionResponseSchema, "/v1/functions/deploy", data);
  },

  // Test a function
  test: async (data: TestFunctionRequest) => {
    // Validate input data
    const validation = testFunctionRequestSchema.safeParse(data);
    if (!validation.success) {
      throw new Error(`Invalid test data: ${validation.error.message}`);
    }
    return apiClient.postValidatedData(testFunctionResponseSchema, "/v1/functions/test", data);
  },

  // Get function deployments
  getDeployments: async (functionId: string): Promise<{ deployments: FunctionDeployment[] }> => {
    const response = await apiClient.get(`/v1/functions/${functionId}/deployments`);
    // Validate the response structure
    if (response && typeof response === 'object' && 'deployments' in response && Array.isArray((response as any).deployments)) {
      return { deployments: validateFunctionDeployments((response as any).deployments) as FunctionDeployment[] };
    }
    return { deployments: [] };
  },

  // Get function logs
  getLogs: async (functionId: string, params?: { limit?: number; since?: string; level?: string }): Promise<{ logs: FunctionLog[] }> => {
    const response = await apiClient.get(`/v1/functions/${functionId}/logs`, { params });
    // Validate the response structure
    if (response && typeof response === 'object' && 'logs' in response && Array.isArray((response as any).logs)) {
      return { logs: validateFunctionLogs((response as any).logs) as FunctionLog[] };
    }
    return { logs: [] };
  },

  // Get deployment logs
  getDeploymentLogs: async (deploymentId: string): Promise<{ logs: FunctionLog[] }> => {
    const response = await apiClient.get(`/v1/deployments/${deploymentId}/logs`);
    // Validate the response structure
    if (response && typeof response === 'object' && 'logs' in response && Array.isArray((response as any).logs)) {
      return { logs: validateFunctionLogs((response as any).logs) as FunctionLog[] };
    }
    return { logs: [] };
  },

  // Get function analytics/metrics
  getMetrics: (functionId: string, params?: { period?: string; from?: string; to?: string }) =>
    apiClient.get(`/v1/functions/${functionId}/metrics`, { params }),
};