/**
 * Ghost Mode API Service
 * Connects StudioPage Ghost Mode components to the backend orchestration engine
 * Phase 5.2-5.9 Backend — Wire frontend to Go Ghost Mode engine
 */

import { useState, useCallback } from 'react';

// Types matching Go backend structs
export type GhostPhase = 
  | 'planning' | 'provisioning' | 'building' | 'deploying'
  | 'monitoring' | 'complete' | 'error' | 'paused';

export interface LogEntry {
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'debug';
  message: string;
}

export interface Artifact {
  name: string;
  type: 'schema' | 'code' | 'config' | 'deployment';
  path: string;
  size?: number;
}

export interface TaskState {
  id: string;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed' | 'skipped';
  phase: GhostPhase;
  started_at?: string;
  completed_at?: string;
  duration_ms?: number;
  logs: LogEntry[];
  artifacts?: Artifact[];
  agent_id?: string;
  confidence?: number;
  dependencies?: string[];
  llm_output?: string;
}

export interface BuildState {
  id: string;
  goal: string;
  description: string;
  phase: GhostPhase;
  progress: number;
  tasks: TaskState[];
  started_at: string;
  updated_at: string;
  estimated_completion?: string;
  current_task_id?: string;
  human_approval_required?: boolean;
  approval_type?: 'schema' | 'deployment' | 'pr' | 'infra';
  error?: string;
  artifacts?: Artifact[];
}

export interface CreateBuildRequest {
  goal: string;
  description?: string;
  agent_id?: string;
}

export interface CreateBuildResponse {
  build_id: string;
  status: string;
  message: string;
}

export interface ArchitecturePlan {
  components: ComponentSpec[];
  data_model: EntitySpec[];
  api_design: EndpointSpec[];
  tech_stack: string[];
  dependencies: string[];
  estimated_cost: string;
  risk_factors: string[];
}

export interface ComponentSpec {
  name: string;
  type: 'api' | 'database' | 'cache' | 'queue' | 'worker' | 'frontend';
  description: string;
  technology: string;
}

export interface EntitySpec {
  name: string;
  fields: FieldSpec[];
  indexes?: string[];
  relations?: RelationSpec[];
}

export interface FieldSpec {
  name: string;
  type: string;
  required: boolean;
  default?: string;
}

export interface RelationSpec {
  entity: string;
  type: 'one_to_one' | 'one_to_many' | 'many_to_many';
  via_field?: string;
}

export interface EndpointSpec {
  method: string;
  path: string;
  handler: string;
  auth: boolean;
  input_schema?: Record<string, unknown>;
  output_schema?: Record<string, unknown>;
}

export interface PlanArchitectureRequest {
  goal: string;
  domain?: string;
}

export interface ProvisionDatabaseRequest {
  schema: string;
}

export interface GenerateBackendRequest {
  spec: string;
  language: string;
}

export interface GenerateFrontendRequest {
  spec: string;
  framework: string;
}

export interface DeployRequest {
  build_id: string;
  commit_sha?: string;
  rollback_id?: string;
}

export interface SetupMonitoringRequest {
  services: string[];
  dashboards: string[];
}

export interface ApprovalRequest {
  build_id: string;
  approval_type: string;
  decision: 'approve' | 'reject' | 'revision';
  notes?: string;
}

// API client
const API_BASE = '/api';

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const msg = (body as any)?.error?.message || `HTTP ${res.status}`;
    throw new Error(msg);
  }

  return res.json() as Promise<T>;
}

// Hook for Ghost Mode
export function useGhostMode() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Create a new Ghost Mode build
  const createBuild = useCallback(async (req: CreateBuildRequest): Promise<CreateBuildResponse> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/builds', {
        method: 'POST',
        body: JSON.stringify(req),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // List all builds
  const listBuilds = useCallback(async (): Promise<BuildState[]> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/builds');
      return res.builds || [];
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Get build by ID
  const getBuild = useCallback(async (id: string): Promise<BuildState> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>(`/v1/ghost/builds/${id}`);
      return res.build;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Pause a build
  const pauseBuild = useCallback(async (id: string): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      await apiFetch(`/v1/ghost/builds/${id}/pause`, { method: 'POST' });
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Resume a build
  const resumeBuild = useCallback(async (id: string): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      await apiFetch(`/v1/ghost/builds/${id}/resume`, { method: 'POST' });
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Cancel a build
  const cancelBuild = useCallback(async (id: string): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      await apiFetch(`/v1/ghost/builds/${id}/cancel`, { method: 'POST' });
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Human approval
  const submitApproval = useCallback(async (req: ApprovalRequest): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      await apiFetch(`/v1/ghost/builds/${req.build_id}/approve`, {
        method: 'POST',
        body: JSON.stringify(req),
      });
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Architecture planning
  const planArchitecture = useCallback(async (req: PlanArchitectureRequest): Promise<ArchitecturePlan> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/plan/architecture', {
        method: 'POST',
        body: JSON.stringify(req),
      });
      return res.plan;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Database provisioning
  const provisionDatabase = useCallback(async (req: ProvisionDatabaseRequest): Promise<{ sql: string; migration: string; artifacts: Artifact[] }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/provision/database', {
        method: 'POST',
        body: JSON.stringify(req),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // API provisioning
  const provisionAPI = useCallback(async (spec: string): Promise<{ handlers: string; artifacts: Artifact[] }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/provision/api', {
        method: 'POST',
        body: JSON.stringify({ spec }),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Backend code generation
  const generateBackend = useCallback(async (req: GenerateBackendRequest): Promise<{ code: string; fallback?: boolean }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/generate/backend', {
        method: 'POST',
        body: JSON.stringify(req),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Frontend code generation
  const generateFrontend = useCallback(async (req: GenerateFrontendRequest): Promise<{ code: string }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/generate/frontend', {
        method: 'POST',
        body: JSON.stringify(req),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Deploy to staging
  const deployStaging = useCallback(async (buildId: string): Promise<{ deploy_id: string; url: string }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/deploy/staging', {
        method: 'POST',
        body: JSON.stringify({ build_id: buildId }),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Deploy to production
  const deployProduction = useCallback(async (buildId: string): Promise<{ deploy_id: string; url: string }> => {
    setLoading(true);
    setError(null);
    try {
      const res = await apiFetch<any>('/v1/ghost/deploy/production', {
        method: 'POST',
        body: JSON.stringify({ build_id: buildId }),
      });
      return res;
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Setup monitoring
  const setupMonitoring = useCallback(async (req: SetupMonitoringRequest): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      await apiFetch('/v1/ghost/monitor/setup', {
        method: 'POST',
        body: JSON.stringify(req),
      });
    } catch (e: any) {
      setError(e.message);
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  // Poll build status (for live updates)
  const pollBuild = useCallback(async (id: string): Promise<BuildState> => {
    const res = await apiFetch<any>(`/v1/ghost/builds/${id}`);
    return res.build;
  }, []);

  return {
    loading,
    error,
    createBuild,
    listBuilds,
    getBuild,
    pauseBuild,
    resumeBuild,
    cancelBuild,
    submitApproval,
    planArchitecture,
    provisionDatabase,
    provisionAPI,
    generateBackend,
    generateFrontend,
    deployStaging,
    deployProduction,
    setupMonitoring,
    pollBuild,
  };
}
