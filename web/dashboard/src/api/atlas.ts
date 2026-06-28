import { apiClient } from './client';
import { API_URLS } from '@/lib/api-urls';

export interface AtlasEvent {
  event_id: string;
  run_id: string;
  timestamp_ns: number;
  sequence: number;
  system_id: string;
  target_system_id?: string;
  kind: 'input' | 'decision' | 'action' | 'result' | 'error';
  parent?: string;
  payload: Record<string, unknown>;
}

export interface AtlasRunRecord {
  tenant_id: string;
  run_id: string;
  labels: Record<string, unknown>;
  first_event_ts_ns?: number;
  last_event_ts_ns?: number;
  event_count: number;
  created_at_ns: number;
}

export interface AtlasGraphNode {
  id: string;
  kind: string;
  system_id: string;
  sequence: number;
  timestamp_ns: number;
  parent?: string;
}

export interface AtlasGraphEdge {
  from: string;
  to: string;
  relation: string;
}

export interface AtlasGraphResponse {
  nodes: AtlasGraphNode[];
  edges: AtlasGraphEdge[];
}

export interface AtlasTraceResponse {
  run: AtlasRunRecord;
  events: AtlasEvent[];
  graph?: AtlasGraphResponse;
}

export interface AtlasListTracesResponse {
  runs: AtlasRunRecord[];
  count: number;
}

export interface AtlasSearchRequest {
  kind?: string;
  system_id?: string;
  since_ns?: number;
  until_ns?: number;
  payload_matches?: { path: string; equals: unknown }[];
  limit?: number;
}

export interface AtlasSearchResponse {
  events: AtlasEvent[];
  scanned_runs: number;
  truncated: boolean;
}

export interface AtlasHealthResponse {
  status: string;
  message?: string;
}

class AtlasApi {
  async getTrace(runId: string): Promise<AtlasTraceResponse> {
    return apiClient.get<AtlasTraceResponse>(API_URLS.atlas.traces.get(runId));
  }

  async listTraces(limit = 50, after?: string): Promise<AtlasListTracesResponse> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (after) params.set('after', after);
    return apiClient.get<AtlasListTracesResponse>(`${API_URLS.atlas.traces.list()}?${params}`);
  }

  async searchTraces(req: AtlasSearchRequest): Promise<AtlasSearchResponse> {
    return apiClient.post<AtlasSearchResponse>(API_URLS.atlas.traces.search(), req);
  }

  async getGraph(runId: string): Promise<AtlasGraphResponse> {
    return apiClient.get<AtlasGraphResponse>(API_URLS.atlas.traces.graph(runId));
  }

  async health(): Promise<AtlasHealthResponse> {
    return apiClient.get<AtlasHealthResponse>(API_URLS.atlas.traces.health());
  }
}

export const atlasApi = new AtlasApi();
