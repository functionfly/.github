import { apiClient } from "./client";
import type {
  ConnectedProvider,
  ConnectProviderRequest,
  ConnectProviderResponse,
} from "@/types";

export const providersApi = {
  async getConnectedProviders(): Promise<ConnectedProvider[]> {
    return apiClient.get<ConnectedProvider[]>("/v1/providers");
  },

  async connectProvider(
    request: ConnectProviderRequest
  ): Promise<ConnectProviderResponse> {
    return apiClient.post<ConnectProviderResponse>("/v1/providers/connect", request);
  },

  async disconnectProvider(providerId: string): Promise<void> {
    return apiClient.delete(`/v1/providers/${providerId}`);
  },

  async testConnection(providerId: string): Promise<{ success: boolean; message?: string }> {
    return apiClient.post(`/v1/providers/${providerId}/test`);
  },
};