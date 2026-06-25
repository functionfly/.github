import apiClient from './client';

export interface FeatureFlag {
  id: string;
  tenant_id: string;
  key: string;
  name: string;
  description?: string;
  flag_type: string;
  is_enabled: boolean;
  rollout_pct: number;
  variants: Record<string, number>;
  target_audience: Record<string, unknown>;
  created_at: string;
}

export const featureFlagsApi = {
  list: () => apiClient.get<{ flags: FeatureFlag[] }>('/v1/feature-flags'),
  get: (id: string) => apiClient.get<{ flag: FeatureFlag }>(`/v1/feature-flags/${id}`),
  create: (data: Partial<FeatureFlag>) => apiClient.post<{ flag: FeatureFlag }>('/v1/feature-flags', data),
  update: (id: string, data: Partial<FeatureFlag>) => apiClient.patch(`/v1/feature-flags/${id}`, data),
  evaluate: (key: string) => apiClient.get<{ enabled: boolean; variant?: string }>(`/v1/feature-flags/${key}/evaluate`),
};
