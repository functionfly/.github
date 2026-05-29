import { apiClient } from './client';

export interface ModelCatalogItem {
  id: string;
  display_name: string;
  provider: string;
  capabilities?: string[];
  tier?: string;
  context_window?: number;
  cost_hint?: string;
  provider_available?: boolean;
}

export interface ModelSelection {
  provider: string;
  model_id: string;
}

export interface TenantAIPreferences {
  profile: 'fast' | 'balanced' | 'premium' | 'custom';
  global_default?: ModelSelection;
  use_same_model_everywhere: boolean;
  defaults: Record<string, ModelSelection>;
  enabled_models: ModelSelection[];
  allow_user_overrides: boolean;
  routing_strategy: 'quality_first' | 'balanced' | 'cost_optimized' | 'cost_first';
}

export const aiModelsApi = {
  getCatalog: async (capability?: string, feature?: string): Promise<ModelCatalogItem[]> => {
    const params = new URLSearchParams();
    if (capability) params.set('capability', capability);
    if (feature) params.set('feature', feature);
    const query = params.toString();
    const response = await apiClient.get(`/v1/ai/models/catalog${query ? `?${query}` : ''}`);
    return (response as { models?: ModelCatalogItem[] }).models || [];
  },

  getPreferences: async (): Promise<TenantAIPreferences> => {
    return apiClient.get('/v1/ai/models/preferences') as Promise<TenantAIPreferences>;
  },

  updatePreferences: async (payload: TenantAIPreferences): Promise<TenantAIPreferences> => {
    return apiClient.put('/v1/ai/models/preferences', payload) as Promise<TenantAIPreferences>;
  },

  refreshCatalog: async (): Promise<void> => {
    await apiClient.post('/v1/ai/models/catalog/refresh', {});
  },

  updateMyOverrides: async (overrides: Record<string, ModelSelection>): Promise<void> => {
    await apiClient.patch('/v1/users/me/settings/ai', { overrides });
  },
};
