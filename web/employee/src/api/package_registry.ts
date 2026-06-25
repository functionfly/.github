import apiClient from './client';

export interface Package {
  id: string;
  name: string;
  scope?: string;
  description?: string;
  registry_type: string;
  latest_version?: string;
  total_downloads: number;
  is_internal: boolean;
  repository_url?: string;
  published_at?: string;
}

export interface PackageVersion {
  id: string;
  version: string;
  description?: string;
  downloads: number;
  published_at: string;
}

export const packageRegistryApi = {
  list: (params?: { registry_type?: string }) => apiClient.get<{ packages: Package[] }>('/v1/packages', { params }),
  get: (id: string) => apiClient.get<{ pkg: Package }>(`/v1/packages/${id}`),
  publish: (data: Partial<Package>) => apiClient.post<{ pkg: Package }>('/v1/packages', data),
  listVersions: (id: string) => apiClient.get<{ versions: PackageVersion[] }>(`/v1/packages/${id}/versions`),
};
