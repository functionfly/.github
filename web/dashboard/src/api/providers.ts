import {
  connectProviderRequestSchema,
  connectProviderResponseSchema,
  connectedProviderSchema,
  testConnectionResponseSchema,
} from '@/lib/api-validation';
import type { ConnectProviderRequest, ConnectProviderResponse, ConnectedProvider, ProviderMaintenanceStatus } from '@/types';
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

// Sanitize provider ID to prevent injection
function sanitizeProviderId(id: string): string {
  return id.replace(/[^a-z0-9-]/g, '');
}

// Sanitize API key input (basic sanitization, actual validation happens in schema)
function sanitizeApiKey(key: string): string {
  return key.trim().replace(/\s+/g, '');
}

export const providersApi = {
  async getConnectedProviders(): Promise<ConnectedProvider[]> {
    const raw = await apiClient.get<ConnectedProvider[]>('/v1/providers');
    const deduped = dedupeConnectedProvidersBySlug(raw);
    // Validate each provider with Zod schema
    return deduped.filter((p) => {
      const result = connectedProviderSchema.safeParse(p);
      if (!result.success) {
        console.warn('Provider validation failed:', result.error);
      }
      return result.success;
    });
  },

  async getProviderCredentials(): Promise<Array<ConnectedProvider & { maskedApiKey: string }>> {
    return apiClient.get('/v1/providers/credentials');
  },

  async connectProvider(request: ConnectProviderRequest): Promise<ConnectProviderResponse> {
    // Sanitize inputs
    const sanitizedRequest = {
      providerId: sanitizeProviderId(request.providerId),
      apiKey: request.apiKey ? sanitizeApiKey(request.apiKey) : '',
    };

    // Validate with Zod before sending
    const validation = connectProviderRequestSchema.safeParse(sanitizedRequest);
    if (!validation.success) {
      const errors = validation.error.issues.map((e) => e.message).join(', ');
      throw new Error(`Invalid request: ${errors}`);
    }

    return apiClient.postValidatedData(
      connectProviderResponseSchema,
      '/v1/providers/connect',
      sanitizedRequest
    );
  },

  async rotateKey(providerId: string, newApiKey: string): Promise<ConnectProviderResponse> {
    const sanitizedId = sanitizeProviderId(providerId);
    if (!sanitizedId) {
      throw new Error('Invalid provider ID');
    }
    const sanitizedKey = newApiKey ? sanitizeApiKey(newApiKey) : '';

    const response = await apiClient.postValidatedData(
      connectProviderResponseSchema,
      `/v1/providers/${sanitizedId}/rotate`,
      { apiKey: sanitizedKey }
    );
    return response;
  },

  async disconnectProvider(providerId: string): Promise<void> {
    const sanitizedId = sanitizeProviderId(providerId);
    if (!sanitizedId) {
      throw new Error('Invalid provider ID');
    }
    return apiClient.delete(`/v1/providers/${sanitizedId}`);
  },

  async testConnection(providerId: string): Promise<{ success: boolean; message?: string }> {
    const sanitizedId = sanitizeProviderId(providerId);
    if (!sanitizedId) {
      throw new Error('Invalid provider ID');
    }
    return apiClient.postValidatedData(
      testConnectionResponseSchema,
      `/v1/providers/${sanitizedId}/test`
    );
  },

  async getProviderMaintenanceStatus(): Promise<Record<string, ProviderMaintenanceStatus>> {
    return apiClient.get('/v1/providers/status');
  },

  async discoverResources(provider: string, apiKey: string): Promise<Array<{ id: string; name: string; url: string }>> {
    const resp = await apiClient.get<{ resources: Array<{ id: string; name: string; url: string }> }>(
      `/v1/providers/discover?provider=${encodeURIComponent(provider)}`,
      { headers: { 'X-Provider-Key': apiKey } },
    );
    return resp.resources ?? [];
  },
};
