import { apiClient } from './client';

export interface Plugin {
  id: string;
  tenant_id: string;
  manifest: Record<string, unknown>;
  plugin_type: PluginType;
  name: string;
  version: string;
  description?: string;
  author_name: string;
  author_email?: string;
  author_website?: string;
  category: string;
  status: PluginStatus;
  icon_url?: string;
  homepage_url?: string;
  repository_url?: string;
  license?: string;
  size_bytes: number;
  signature?: string;
  verified: boolean;
  config: Record<string, string>;
  metadata: Record<string, unknown>;
  installed_at: string;
  updated_at: string;
  enabled_at?: string;
  error?: string;
}

export type PluginType = 'ui' | 'graph' | 'ai_tool' | 'runtime' | 'infrastructure' | 'marketplace';
export type PluginStatus = 'enabled' | 'disabled' | 'error' | 'paused';
export type SandboxTier = 'wasm' | 'worker' | 'gvisor' | 'microvm' | 'enterprise';

export interface PluginSandbox {
  id: string;
  plugin_id: string;
  tier: SandboxTier;
  cpu_limit: number;
  memory_limit_mb: number;
  timeout_seconds: number;
  network_isolated: boolean;
  filesystem_scope: string;
  max_instances: number;
  env_vars?: Record<string, string>;
  allowed_domains?: string[];
  blocked_domains?: string[];
  rate_limit_rpm?: number;
  created_at: string;
  updated_at: string;
}

export interface PluginPermission {
  id: string;
  plugin_id: string;
  permission_type: string;
  permission_action: string;
  resource?: string;
  granted: boolean;
  granted_at?: string;
  granted_by?: string;
  expires_at?: string;
  created_at: string;
}

export interface PluginVersion {
  id: string;
  plugin_id: string;
  version: string;
  changelog?: string;
  manifest: Record<string, unknown>;
  size_bytes: number;
  signature?: string;
  release_at: string;
  created_at: string;
}

export interface InstallPluginRequest {
  manifest: Record<string, unknown>;
  plugin_type: PluginType;
  name: string;
  version: string;
  description?: string;
  author_name: string;
  author_email?: string;
  author_website?: string;
  category?: string;
  icon_url?: string;
  homepage_url?: string;
  repository_url?: string;
  license?: string;
  size_bytes?: number;
  signature?: string;
  sandbox_tier?: SandboxTier;
  sandbox_config?: SandboxConfig;
}

export interface SandboxConfig {
  cpu_limit?: number;
  memory_limit_mb?: number;
  timeout_seconds?: number;
  allowed_domains?: string[];
  blocked_domains?: string[];
}

export interface UpdateSandboxRequest {
  tier?: SandboxTier;
  cpu_limit?: number;
  memory_limit_mb?: number;
  timeout_seconds?: number;
  allowed_domains?: string[];
  blocked_domains?: string[];
  rate_limit_rpm?: number;
}

export interface SetPermissionRequest {
  permission_type: string;
  permission_action: string;
  resource?: string;
  granted: boolean;
  expires_at?: string;
}

export interface PluginFilters {
  type?: PluginType;
  status?: PluginStatus;
  category?: string;
  search?: string;
  limit?: number;
  offset?: number;
}

export interface TelemetryData {
  executions: number;
  errors: number;
  error_rate: number;
  avg_latency_ms: number;
  cpu_usage_seconds: number;
  avg_memory_usage_mb: number;
  network_bytes: number;
  previous_executions: number;
  latency_trend: "up" | "down" | "stable";
  executions_trend: "up" | "down" | "stable";
}

export interface RateLimitCheckRequest {
  ip: string;
  endpoint: string;
  user_id?: string;
  limit?: number;
  window_sec?: number;
}

export interface RateLimitCheckResponse {
  enabled: boolean;
  message?: string;
  allowed?: boolean;
  remaining?: number;
  reset_at?: number;
  limit?: number;
}

export const pluginsApi = {
  list: async (filters?: PluginFilters): Promise<{ plugins: Plugin[] }> => {
    const params: Record<string, string | number | undefined> = {};
    if (filters?.type) params.type = filters.type;
    if (filters?.status) params.status = filters.status;
    if (filters?.category) params.category = filters.category;
    if (filters?.search) params.search = filters.search;
    if (filters?.limit) params.limit = filters.limit;
    if (filters?.offset) params.offset = filters.offset;

    const response = await apiClient.get('/plugins', { params });
    if (response && typeof response === 'object' && 'plugins' in response) {
      return response as { plugins: Plugin[] };
    }
    return { plugins: [] };
  },

  get: async (pluginId: string): Promise<{ plugin: Plugin }> => {
    const response = await apiClient.get(`/plugins/${pluginId}`);
    return response as { plugin: Plugin };
  },

  install: async (data: InstallPluginRequest): Promise<{ plugin: Plugin }> => {
    const response = await apiClient.post('/plugins', data);
    return response as { plugin: Plugin };
  },

  update: async (pluginId: string, data: Partial<Plugin>): Promise<{ plugin: Plugin }> => {
    const response = await apiClient.put(`/plugins/${pluginId}`, data);
    return response as { plugin: Plugin };
  },

  uninstall: async (pluginId: string): Promise<{ message: string }> => {
    const response = await apiClient.delete(`/plugins/${pluginId}`);
    return response as { message: string };
  },

  enable: async (pluginId: string): Promise<{ message: string }> => {
    const response = await apiClient.post(`/plugins/${pluginId}/enable`);
    return response as { message: string };
  },

  disable: async (pluginId: string): Promise<{ message: string }> => {
    const response = await apiClient.post(`/plugins/${pluginId}/disable`);
    return response as { message: string };
  },

  pause: async (pluginId: string): Promise<{ message: string }> => {
    const response = await apiClient.post(`/plugins/${pluginId}/pause`);
    return response as { message: string };
  },

  rollback: async (pluginId: string, toVersion?: string): Promise<{ plugin: Plugin; rolled_back_to: string }> => {
    const params: Record<string, string> = {};
    if (toVersion) params.to_version = toVersion;
    const response = await apiClient.post(`/plugins/${pluginId}/rollback`, null, { params });
    return response as { plugin: Plugin; rolled_back_to: string };
  },

  configure: async (pluginId: string, config: Record<string, string>): Promise<{ message: string }> => {
    const response = await apiClient.put(`/plugins/${pluginId}/config`, { config });
    return response as { message: string };
  },

  getSandbox: async (pluginId: string): Promise<{ sandbox: PluginSandbox | null }> => {
    const response = await apiClient.get(`/plugins/${pluginId}/sandbox`);
    return response as { sandbox: PluginSandbox | null };
  },

  updateSandbox: async (pluginId: string, data: UpdateSandboxRequest): Promise<{ sandbox: PluginSandbox }> => {
    const response = await apiClient.put(`/plugins/${pluginId}/sandbox`, data);
    return response as { sandbox: PluginSandbox };
  },

  getPermissions: async (pluginId: string): Promise<{ permissions: PluginPermission[] }> => {
    const response = await apiClient.get(`/plugins/${pluginId}/permissions`);
    return response as { permissions: PluginPermission[] };
  },

  setPermission: async (pluginId: string, data: SetPermissionRequest): Promise<{ message: string }> => {
    const response = await apiClient.post(`/plugins/${pluginId}/permissions`, data);
    return response as { message: string };
  },

  listVersions: async (pluginId: string): Promise<{ versions: PluginVersion[] }> => {
    const response = await apiClient.get(`/plugins/${pluginId}/versions`);
    return response as { versions: PluginVersion[] };
  },

  getTelemetry: async (pluginId: string, timeRange: string = "7d"): Promise<{ telemetry: TelemetryData }> => {
    const response = await apiClient.get(`/plugins/${pluginId}/telemetry`, { params: { range: timeRange } });
    return response as { telemetry: TelemetryData };
  },

  setError: async (pluginId: string, error: string): Promise<{ message: string }> => {
    const response = await apiClient.post(`/plugins/${pluginId}/error`, { error });
    return response as { message: string };
  },

  checkRateLimit: async (data: RateLimitCheckRequest): Promise<RateLimitCheckResponse> => {
    const response = await apiClient.post('/plugins/check-rate-limit', data);
    return response as RateLimitCheckResponse;
  },

  recordAnalytics: async (pluginId: string, data: Partial<TelemetryData> & { event_type?: string; period_start?: string; period_end?: string }): Promise<{ message: string }> => {
    const response = await apiClient.post(`/plugins/${pluginId}/analytics`, data);
    return response as { message: string };
  },
};