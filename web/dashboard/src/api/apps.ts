import type { App, AppStatus, AppAnalyticsResponse, Backend, CreateAppRequest, CreateBackendRequest, UpdateAppRequest } from '@/types';
import { apiClient } from './client';

/** Row for POST /v1/functions/deploy (backend_id) — aggregated from all tenant apps */
export interface DeployBackendOption {
  id: string;
  appId: string;
  appName: string;
  provider: string;
  region: string;
}

export const appsApi = {
  list: () => apiClient.get<{ apps: App[] }>('/v1/apps'),

  get: (appId: string) => apiClient.get<App>(`/v1/apps/${appId}`),

  create: (data: CreateAppRequest) => apiClient.post<App>('/v1/apps', data),

  update: (appId: string, data: UpdateAppRequest) => apiClient.patch<App>(`/v1/apps/${appId}`, data),

  delete: (appId: string) => apiClient.delete<void>(`/v1/apps/${appId}`),

  getStatus: (appId: string) => apiClient.get<AppStatus>(`/v1/apps/${appId}/status`),

  listBackends: (appId: string) =>
    apiClient.get<{ backends: Backend[] }>(`/v1/apps/${appId}/backends`),

  createBackend: (appId: string, data: CreateBackendRequest) =>
    apiClient.post<Backend>(`/v1/apps/${appId}/backends`, data),

  deleteBackend: (appId: string, backendId: string) =>
    apiClient.delete<void>(`/v1/apps/${appId}/backends/${backendId}`),

  updateBackend: (appId: string, backendId: string, data: { enabled?: boolean; priority?: number }) =>
    apiClient.patch<Backend>(`/v1/apps/${appId}/backends/${backendId}`, data),

  getRoute: (appId: string, params?: { clientRegion?: string; method?: string }) =>
    apiClient.get(`/v1/apps/${appId}/route`, { params }),

  getAnalytics: (appId: string, days = 7) =>
    apiClient.get<AppAnalyticsResponse>(`/v1/apps/${appId}/analytics?days=${days}`),

  listSecrets: (appId: string) =>
    apiClient.get<{ secrets: Array<{ key: string; value?: string }> }>(`/v1/apps/${appId}/secrets`),

  setSecrets: (appId: string, data: { provider: string; secrets: Record<string, string> }) =>
    apiClient.post<void>(`/v1/apps/${appId}/secrets`, data),
};

export async function fetchDeployBackendOptions(): Promise<DeployBackendOption[]> {
  const { apps } = await appsApi.list();
  const out: DeployBackendOption[] = [];
  for (const app of apps) {
    const { backends } = await appsApi.listBackends(app.id);
    for (const b of backends) {
      const raw = b as unknown as Record<string, unknown>;
      const id = typeof raw.id === 'string' ? raw.id : null;
      if (!id) continue;
      out.push({
        id,
        appId: typeof raw.app_id === 'string' ? raw.app_id : app.id,
        appName: app.name,
        provider: String(raw.provider ?? ''),
        region: String(raw.region ?? ''),
      });
    }
  }
  return out;
}
