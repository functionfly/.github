import { apiClient } from "./client";
import type {
  APIKey,
  APIKeyCreateResponse,
  APIKeyListResponse,
  CreateAPIKeyRequest,
  UpdateAPIKeyRequest,
  RotateAPIKeyRequest,
  APIKeyPermission,
  APIKeyEnvironment,
  AddPermissionRequest,
  AddEnvironmentRequest,
  APIKeyFilters,
} from "@/types/api-key";

export const apiKeysApi = {
  // List all API keys for the current tenant
  list: async (filters?: APIKeyFilters): Promise<APIKeyListResponse> => {
    const response = await apiClient.get<APIKeyListResponse>("/v1/api-keys", { params: filters });
    return response;
  },

  // Get a specific API key
  get: async (id: string): Promise<{ data: APIKey }> => {
    const response = await apiClient.get<{ data: APIKey }>(`/v1/api-keys/${id}`);
    return response;
  },

  // Create a new API key
  create: async (data: CreateAPIKeyRequest): Promise<{ data: APIKeyCreateResponse }> => {
    const response = await apiClient.post<{ data: APIKeyCreateResponse }>("/v1/api-keys", data);
    return response;
  },

  // Update an existing API key
  update: async (id: string, data: UpdateAPIKeyRequest): Promise<{ data: APIKey }> => {
    const response = await apiClient.patch<{ data: APIKey }>(`/v1/api-keys/${id}`, data);
    return response;
  },

  // Delete an API key
  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/v1/api-keys/${id}`);
  },

  // Rotate an API key
  rotate: async (id: string, data?: RotateAPIKeyRequest): Promise<{ data: APIKeyCreateResponse }> => {
    const response = await apiClient.post<{ data: APIKeyCreateResponse }>(`/v1/api-keys/${id}/rotate`, data || {});
    return response;
  },

  // List permissions for an API key
  listPermissions: async (id: string): Promise<{ data: APIKeyPermission[] }> => {
    const response = await apiClient.get<{ data: APIKeyPermission[] }>(`/v1/api-keys/${id}/permissions`);
    return response;
  },

  // Add a permission to an API key
  addPermission: async (id: string, data: AddPermissionRequest): Promise<{ data: APIKeyPermission }> => {
    const response = await apiClient.post<{ data: APIKeyPermission }>(`/v1/api-keys/${id}/permissions`, data);
    return response;
  },

  // Remove a permission from an API key
  removePermission: async (id: string, permId: string): Promise<void> => {
    await apiClient.delete(`/v1/api-keys/${id}/permissions/${permId}`);
  },

  // List environments for an API key
  listEnvironments: async (id: string): Promise<{ data: APIKeyEnvironment[] }> => {
    const response = await apiClient.get<{ data: APIKeyEnvironment[] }>(`/v1/api-keys/${id}/environments`);
    return response;
  },

  // Add an environment to an API key
  addEnvironment: async (id: string, data: AddEnvironmentRequest): Promise<{ data: APIKeyEnvironment }> => {
    const response = await apiClient.post<{ data: APIKeyEnvironment }>(`/v1/api-keys/${id}/environments`, data);
    return response;
  },

  // Remove an environment from an API key
  removeEnvironment: async (id: string, envId: string): Promise<void> => {
    await apiClient.delete(`/v1/api-keys/${id}/environments/${envId}`);
  },
};
