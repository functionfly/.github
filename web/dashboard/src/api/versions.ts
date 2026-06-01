import { apiClient } from './client';

export type FunctionVersionStatus = 'draft' | 'published' | 'deprecated' | 'archived' | string;

export interface RegistryFunctionVersion {
  id: string;
  functionId: string;
  version: string;
  status: FunctionVersionStatus;
  publishedAt?: string;
  isLatest: boolean;
  isStable: boolean;
  deprecation?: {
    reason?: string;
    migrationGuide?: string;
    replacedBy?: string;
  };
}

export interface VersionDiffResponse {
  fromVersion: string;
  toVersion: string;
  changes: Array<{
    field: string;
    fromValue: string;
    toValue: string;
    changeType: string;
  }>;
  breakingChanges?: string[];
  isBreaking?: boolean;
  summary?: { added: number; removed: number; modified: number };
}

export interface RollbackRecord {
  id: string;
  function_id: string;
  from_version: string;
  to_version: string;
  strategy: string;
  status: string;
  initiated_at: string;
  completed_at?: string;
}

export interface CanaryConfig {
  id: string;
  function_id: string;
  version: string;
  traffic_percent: number;
  auto_promote: boolean;
  promote_threshold?: number;
  promote_window?: number;
  status?: string;
}

const registryBase = (functionId: string) => `/v1/functions/${functionId}`;

export const versionsApi = {
  list(functionId: string, status?: string) {
    const q = status ? `?status=${encodeURIComponent(status)}` : '';
    return apiClient.get<{ versions: RegistryFunctionVersion[] }>(
      `${registryBase(functionId)}/versions${q}`
    );
  },

  get(functionId: string, version: string) {
    return apiClient.get<RegistryFunctionVersion>(
      `${registryBase(functionId)}/versions/${encodeURIComponent(version)}`
    );
  },

  publish(
    functionId: string,
    version: string,
    body?: { setAsLatest?: boolean; setAsStable?: boolean }
  ) {
    return apiClient.post<RegistryFunctionVersion>(
      `${registryBase(functionId)}/versions/${encodeURIComponent(version)}/publish`,
      {
        version,
        setAsLatest: body?.setAsLatest ?? true,
        setAsStable: body?.setAsStable ?? false,
      }
    );
  },

  archive(functionId: string, version: string, reason?: string) {
    return apiClient.post<{ version: string; status: string; archivedAt?: string }>(
      `${registryBase(functionId)}/versions/${encodeURIComponent(version)}/archive`,
      { reason: reason ?? '' }
    );
  },

  deprecate(
    functionId: string,
    version: string,
    body: { reason: string; replacedBy?: string; migrationGuide?: string; gracePeriodDays?: number }
  ) {
    return apiClient.post<Record<string, unknown>>(
      `${registryBase(functionId)}/versions/${encodeURIComponent(version)}/deprecate`,
      body
    );
  },

  setAlias(functionId: string, version: string, alias: 'latest' | 'stable') {
    return apiClient.post<{ alias: string; version: string }>(
      `${registryBase(functionId)}/versions/${encodeURIComponent(version)}/alias/${alias}`,
      {}
    );
  },

  rollbackToVersion(functionId: string, version: string, strategy = 'immediate') {
    return apiClient.post<{
      rollbackId: string;
      fromVersion: string;
      toVersion: string;
      status: string;
    }>(`${registryBase(functionId)}/versions/${encodeURIComponent(version)}/rollback`, {
      toVersion: version,
      strategy,
    });
  },

  rollbackLatest(functionId: string, strategy = 'immediate') {
    return apiClient.post<{
      rollbackId: string;
      fromVersion: string;
      toVersion: string;
      status: string;
    }>(`${registryBase(functionId)}/rollback`, { strategy });
  },

  rollbackHistory(functionId: string, limit = 20) {
    return apiClient.get<{ rollbacks: RollbackRecord[] }>(
      `${registryBase(functionId)}/rollbacks?limit=${limit}`
    );
  },

  compare(functionId: string, v1: string, v2: string) {
    return apiClient.get<VersionDiffResponse>(
      `${registryBase(functionId)}/versions/compare?v1=${encodeURIComponent(v1)}&v2=${encodeURIComponent(v2)}`
    );
  },

  createChangelog(
    functionId: string,
    version: string,
    body: {
      changeType: string;
      changeCategory: string;
      description: string;
      breakingChanges?: string[];
      migrationSteps?: string[];
    }
  ) {
    return apiClient.post<Record<string, unknown>>(
      `${registryBase(functionId)}/versions/${encodeURIComponent(version)}/changelog`,
      body
    );
  },
};

export const registryCanaryApi = {
  get(author: string, name: string) {
    return apiClient.get<CanaryConfig | { message?: string }>(
      `/functions/${author}/${name}/canary`
    );
  },

  create(
    author: string,
    name: string,
    body: {
      version: string;
      traffic_percent: number;
      auto_promote?: boolean;
      promote_threshold?: number;
      promote_window?: number;
    }
  ) {
    return apiClient.post<CanaryConfig>(`/functions/${author}/${name}/canary`, body);
  },

  update(
    author: string,
    name: string,
    body: Partial<{
      traffic_percent: number;
      auto_promote: boolean;
      promote_threshold: number;
      promote_window: number;
    }>
  ) {
    return apiClient.patch<CanaryConfig>(`/functions/${author}/${name}/canary`, body);
  },

  cancel(author: string, name: string) {
    return apiClient.delete<void>(`/functions/${author}/${name}/canary`);
  },

  promote(author: string, name: string) {
    return apiClient.post<CanaryConfig>(`/functions/${author}/${name}/canary/promote`, {});
  },

  rollback(author: string, name: string) {
    return apiClient.post<CanaryConfig>(`/functions/${author}/${name}/canary/rollback`, {});
  },
};
