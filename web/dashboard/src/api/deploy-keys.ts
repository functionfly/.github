import { apiClient } from "./client";

export interface DeployKey {
  id: string;
  name: string;
  public_key: string;
  fingerprint: string;
  created_at: string;
  expires_at?: string;
  last_used_at?: string;
  created_by: string;
}

export interface DeployKeyListResponse {
  deploy_keys: DeployKey[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface DeployKeyCreateRequest {
  name: string;
  public_key: string;
}

export const deployKeysApi = {
  list: async (): Promise<DeployKeyListResponse> => {
    const response = await apiClient.get<DeployKeyListResponse>("/v1/deploy-keys");
    return response;
  },

  get: async (id: string): Promise<{ data: DeployKey }> => {
    const response = await apiClient.get<{ data: DeployKey }>(`/v1/deploy-keys/${id}`);
    return response;
  },

  create: async (data: DeployKeyCreateRequest): Promise<{ data: DeployKey }> => {
    const response = await apiClient.post<{ data: DeployKey }>("/v1/deploy-keys", data);
    return response;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/v1/deploy-keys/${id}`);
  },

  verify: async (id: string): Promise<{ valid: boolean; fingerprint: string }> => {
    const response = await apiClient.post<{ valid: boolean; fingerprint: string }>(`/v1/deploy-keys/${id}/verify`);
    return response;
  },
};