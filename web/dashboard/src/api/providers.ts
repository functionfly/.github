import type { ConnectedProvider, ConnectProviderRequest, ConnectProviderResponse } from '@/types';
import { apiClient } from './client';

/** One row per provider slug (newest `connectedAt` wins). Use after fetch to hide legacy duplicate DB rows. */
export function dedupeConnectedProvidersBySlug(list: ConnectedProvider[]): ConnectedProvider[] {
  const bySlug = new Map<string, ConnectedProvider>();
  for (const p of list) {
    const key = p.name.trim().toLowerCase();
    const prev = bySlug.get(key);
    if (!prev || new Date(p.connectedAt).getTime() > new Date(prev.connectedAt).getTime()) {
      bySlug.set(key, p);
    }
  }
  return Array.from(bySlug.values());
}

export const providersApi = {
  async getConnectedProviders(): Promise<ConnectedProvider[]> {
    const raw = await apiClient.get<ConnectedProvider[]>('/v1/providers');
    return dedupeConnectedProvidersBySlug(raw);
  },

  async connectProvider(request: ConnectProviderRequest): Promise<ConnectProviderResponse> {
    return apiClient.post<ConnectProviderResponse>('/v1/providers/connect', {
      providerId: request.providerId,
      apiKey: request.apiKey,
    });
  },

  async disconnectProvider(providerId: string): Promise<void> {
    return apiClient.delete(`/v1/providers/${providerId}`);
  },

  async testConnection(providerId: string): Promise<{ success: boolean; message?: string }> {
    return apiClient.post(`/v1/providers/${providerId}/test`);
  },
};
