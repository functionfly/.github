import { apiClient } from './client';

// ============================================================================
// Pipeline Types
// ============================================================================

export type PipelineStageStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped' | 'waiting';
export type PipelineStatus = 'active' | 'paused' | 'archived';

export interface PipelineTask {
  id: string;
  name: string;
  status: string;
  duration?: number;
  logs?: string[];
  error?: string;
}

export interface PipelineStage {
  id: string;
  name: string;
  status: PipelineStageStatus;
  duration?: number;
  started_at?: number;
  completed_at?: number;
  tasks: PipelineTask[];
}

export interface Pipeline {
  id: string;
  name: string;
  version: string;
  status: PipelineStatus;
  stages: PipelineStage[];
  current_stage_id?: string;
  triggered_by: string;
  triggered_at: number;
  branch: string;
  commit_sha: string;
  source: string;
  tenant_id?: string;
  created_at?: number;
  updated_at?: number;
}

// ============================================================================
// Environment Types
// ============================================================================

export type EnvironmentType = 'production' | 'staging' | 'preview' | 'development';

export interface EnvironmentSecret {
  key: string;
  masked: boolean;
  last_updated: number;
}

export interface Environment {
  id: string;
  name: string;
  type: EnvironmentType;
  color: string;
  variables: Record<string, string>;
  secrets: EnvironmentSecret[];
  replicas: number;
  auto_scale: boolean;
  region?: string;
  tenant_id?: string;
  created_at?: number;
  updated_at?: number;
}

// ============================================================================
// Cloud Region Types
// ============================================================================

export type CloudProvider = 'aws' | 'gcp' | 'azure' | 'custom';

export interface CloudRegionSpec {
  compute?: number;
  memory?: number;
  storage?: number;
  gpu?: boolean;
}

export interface CloudRegion {
  id: string;
  name: string;
  provider: CloudProvider;
  zone: string;
  zone_name: string;
  location: string;
  country: string;
  coordinates: { lat: number; lng: number };
  is_available: boolean;
  is_recommended?: boolean;
  specs?: CloudRegionSpec;
  tenant_id?: string;
}

// ============================================================================
// DevOps Stats
// ============================================================================

export interface DevOpsStats {
  pipelines: number;
  active_pipelines: number;
  success_rate: number;
  avg_cold_start_ms: number;
  environments: number;
  regions: number;
}

// ============================================================================
// API
// ============================================================================

export const studioDevOpsApi = {
  // Stats
  getStats: () =>
    apiClient.get<{ stats: DevOpsStats }>('/v1/studio/devops/stats'),

  // Pipelines
  listPipelines: (params?: { status?: string; limit?: number; offset?: number }) => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set('status', params.status);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    const query = searchParams.toString();
    return apiClient.get<{ pipelines: Pipeline[]; total: number }>(
      `/v1/studio/devops/pipelines${query ? `?${query}` : ''}`
    );
  },

  getPipeline: (id: string) =>
    apiClient.get<{ pipeline: Pipeline }>(`/v1/studio/devops/pipelines/${id}`),

  createPipeline: (data: {
    name: string;
    version?: string;
    branch?: string;
    commit_sha?: string;
    source?: string;
  }) => apiClient.post<{ pipeline: Pipeline }>('/v1/studio/devops/pipelines', data),

  updatePipelineStage: (
    pipelineId: string,
    stageId: string,
    updates: Partial<PipelineStage>
  ) =>
    apiClient.patch<{ pipeline: Pipeline }>(
      `/v1/studio/devops/pipelines/${pipelineId}/stages/${stageId}`,
      updates
    ),

  retryPipelineStage: (pipelineId: string, stageId: string) =>
    apiClient.post<{ pipeline: Pipeline }>(
      `/v1/studio/devops/pipelines/${pipelineId}/stages/${stageId}/retry`,
      {}
    ),

  // Environments
  listEnvironments: (params?: { type?: EnvironmentType }) => {
    const searchParams = new URLSearchParams();
    if (params?.type) searchParams.set('type', params.type);
    const query = searchParams.toString();
    return apiClient.get<{ environments: Environment[] }>(
      `/v1/studio/devops/environments${query ? `?${query}` : ''}`
    );
  },

  getEnvironment: (id: string) =>
    apiClient.get<{ environment: Environment }>(`/v1/studio/devops/environments/${id}`),

  createEnvironment: (data: {
    name: string;
    type: EnvironmentType;
    color?: string;
    variables?: Record<string, string>;
    replicas?: number;
    auto_scale?: boolean;
    region?: string;
  }) => apiClient.post<{ environment: Environment }>('/v1/studio/devops/environments', data),

  updateEnvironment: (id: string, updates: Partial<Environment>) =>
    apiClient.patch<{ environment: Environment }>(
      `/v1/studio/devops/environments/${id}`,
      updates
    ),

  deleteEnvironment: (id: string) =>
    apiClient.delete<{ message: string }>(`/v1/studio/devops/environments/${id}`),

  addEnvironmentVariable: (envId: string, key: string, value: string) =>
    apiClient.post<{ environment: Environment }>(
      `/v1/studio/devops/environments/${envId}/variables`,
      { key, value }
    ),

  addEnvironmentSecret: (envId: string, key: string) =>
    apiClient.post<{ environment: Environment }>(
      `/v1/studio/devops/environments/${envId}/secrets`,
      { key }
    ),

  // Cloud Regions
  listRegions: (params?: { provider?: CloudProvider }) => {
    const searchParams = new URLSearchParams();
    if (params?.provider) searchParams.set('provider', params.provider);
    const query = searchParams.toString();
    return apiClient.get<{ regions: CloudRegion[] }>(
      `/v1/studio/devops/regions${query ? `?${query}` : ''}`
    );
  },

  getRegion: (id: string) =>
    apiClient.get<{ region: CloudRegion }>(`/v1/studio/devops/regions/${id}`),

  createRegion: (data: {
    name: string;
    provider: CloudProvider;
    zone?: string;
    zone_name?: string;
    location?: string;
    country?: string;
    specs?: CloudRegionSpec;
  }) => apiClient.post<{ region: CloudRegion }>('/v1/studio/devops/regions', data),
};