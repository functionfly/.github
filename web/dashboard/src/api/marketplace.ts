import { apiClient } from './client';

export interface Extension {
  id: string;
  creator_id: string;
  plugin_id?: string;
  name: string;
  version: string;
  description: string;
  category: string;
  icon_url?: string;
  homepage_url?: string;
  screenshots?: string[];
  manifest?: Record<string, unknown>;
  manifest_url?: string;
  signature?: string;
  verified: boolean;
  status: string;
  featured: boolean;
  install_count: number;
  rating_average: number;
  rating_count: number;
  trust_score: number;
  sandbox_score: number;
  security_score: number;
  runtime_score: number;
  compatibility?: Record<string, unknown>;
  tags?: string[];
  changelog?: string;
  release_notes?: string;
  created_at?: string;
  updated_at?: string;
  published_at?: string;
}

export interface CategoryCount {
  category: string;
  count: number;
}

export interface MarketplaceFilters {
  category?: string;
  status?: string;
  search?: string;
  featured?: boolean;
  tags?: string[];
  sort?: "trending" | "top_rated" | "newest" | "most_installed";
  limit?: number;
  offset?: number;
}

export interface CreateExtensionRequest {
  name: string;
  version: string;
  description?: string;
  category?: string;
  icon_url?: string;
  manifest?: Record<string, unknown>;
  tags?: string[];
}

export interface UpdateExtensionRequest {
  name?: string;
  version?: string;
  description?: string;
  category?: string;
  icon_url?: string;
  manifest?: Record<string, unknown>;
  status?: string;
  featured?: boolean;
  tags?: string[];
}

export interface Rating {
  id: string;
  extension_id: string;
  tenant_id: string;
  rating: number;
  review?: string;
  created_at: string;
  updated_at: string;
}

export interface ExtensionUpdate {
  installed_plugin_id: string;
  installed_version: string;
  extension_id: string;
  latest_version: string;
  changelog: string;
  manifest?: Record<string, unknown>;
}

export interface SandboxConfig {
  id: string;
  name: string;
  tier: 'free' | 'basic' | 'pro' | 'enterprise';
  memoryLimit: number;
  timeout: number;
  maxEndpoints: number;
  price: number;
}

export interface InstalledPluginInfo {
  id: string;
  name: string;
  version: string;
}

export const marketplaceApi = {
  list: async (filters?: MarketplaceFilters): Promise<{ extensions: Extension[] }> => {
    const params: Record<string, string | number | boolean | undefined> = {};
    if (filters?.category) params.category = filters.category;
    if (filters?.status) params.status = filters.status;
    if (filters?.search) params.search = filters.search;
    if (filters?.featured !== undefined) params.featured = filters.featured;
    if (filters?.tags && filters.tags.length > 0) params.tags = filters.tags.join(',');
    if (filters?.sort) params.sort = filters.sort;
    if (filters?.limit) params.limit = filters.limit;
    if (filters?.offset) params.offset = filters.offset;

    const response = await apiClient.get('/marketplace/extensions', { params });
    if (response && typeof response === 'object' && 'extensions' in response) {
      return response as { extensions: Extension[] };
    }
    return { extensions: [] };
  },

  get: async (extensionId: string): Promise<{ extension: Extension }> => {
    const response = await apiClient.get(`/marketplace/extensions/${extensionId}`);
    return response as { extension: Extension };
  },

  create: async (data: CreateExtensionRequest): Promise<{ extension: Extension }> => {
    const response = await apiClient.post('/marketplace/extensions', data);
    return response as { extension: Extension };
  },

  update: async (extensionId: string, data: UpdateExtensionRequest): Promise<{ extension: Extension }> => {
    const response = await apiClient.put(`/marketplace/extensions/${extensionId}`, data);
    return response as { extension: Extension };
  },

  delete: async (extensionId: string): Promise<{ message: string }> => {
    const response = await apiClient.delete(`/marketplace/extensions/${extensionId}`);
    return response as { message: string };
  },

  install: async (extensionId: string): Promise<{ message: string; extension: Extension; extension_id: string; plugin_manifest?: Record<string, unknown> }> => {
    const response = await apiClient.post(`/marketplace/extensions/${extensionId}/install`);
    return response as { message: string; extension: Extension; extension_id: string; plugin_manifest?: Record<string, unknown> };
  },

  rate: async (extensionId: string, rating: number, review?: string): Promise<{ message: string; rating: Rating }> => {
    const response = await apiClient.post(`/marketplace/extensions/${extensionId}/rate`, { rating, review: review || '' });
    return response as { message: string; rating: Rating };
  },

  getMyRating: async (extensionId: string): Promise<{ rating: Rating | null }> => {
    const response = await apiClient.get(`/marketplace/extensions/${extensionId}/my-rating`);
    return response as { rating: Rating | null };
  },

  listRatings: async (extensionId: string, limit: number = 50): Promise<{ ratings: Rating[] }> => {
    const response = await apiClient.get(`/marketplace/extensions/${extensionId}/ratings`, { params: { limit } });
    return response as { ratings: Rating[] };
  },

  checkUpdates: async (installed: InstalledPluginInfo[]): Promise<{ updates: ExtensionUpdate[] }> => {
    const response = await apiClient.post('/marketplace/check-updates', installed);
    return response as { updates: ExtensionUpdate[] };
  },

  getCategories: async (): Promise<{ categories: CategoryCount[] }> => {
    const response = await apiClient.get('/marketplace/categories');
    return response as { categories: CategoryCount[] };
  },

  getInstallCounts: async (extensionIds: string[]): Promise<{ install_counts: Record<string, number> }> => {
    const ids = extensionIds.join(',');
    const response = await apiClient.get('/marketplace/install-counts', { params: { ids } });
    return response as { install_counts: Record<string, number> };
  },
};