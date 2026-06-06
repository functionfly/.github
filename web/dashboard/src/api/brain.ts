import { apiClient } from "./client";

export interface BrainSignal {
  id: string;
  tenant_id: string;
  connector_slug: string;
  signal_type: string;
  entity_id: string;
  entity_name: string;
  fact: string;
  importance: number;
  source_url: string;
  created_at: string;
  last_seen_at: string;
  ttl_hours: number;
  metadata?: Record<string, unknown>;
}

export interface BrainStats {
  total_signals: number;
  signals_by_type: Record<string, number>;
  signals_by_connector: Record<string, number>;
  oldest_signal?: string;
  newest_signal?: string;
  memory_used: number;
  memory_max: number;
  retention_days: number;
}

export interface BrainComposer {
  id: string;
  tenant_id: string;
  name: string;
  schedule: string;
  signal_filters: SignalFilter[];
  output_format: string;
  actions: ComposerAction[];
  is_active: boolean;
  last_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface SignalFilter {
  connector_slugs: string[];
  signal_types: string[];
  importance_min: number;
  time_window: string;
}

export interface ComposerAction {
  type: string;
  config: Record<string, unknown>;
}

export interface BrainTrigger {
  id: string;
  tenant_id: string;
  agent_id?: string;
  name: string;
  signal_types: string[];
  connector_slugs: string[];
  min_importance: number;
  schedule: string;
  action: string;
  action_config?: Record<string, unknown>;
  is_active: boolean;
  last_fired_at?: string;
  created_at: string;
  updated_at: string;
}

export interface BrainFeedbackRequest {
  signal_id: string;
  helpful: boolean;
  context?: string;
}

export const brainApi = {
  listSignals: async (params?: {
    connector?: string;
    type?: string;
    limit?: number;
    offset?: number;
    sort?: string;
  }): Promise<{ signals: BrainSignal[]; total: number }> => {
    const searchParams = new URLSearchParams();
    if (params?.connector) searchParams.set("connector", params.connector);
    if (params?.type) searchParams.set("type", params.type);
    if (params?.limit) searchParams.set("limit", String(params.limit));
    if (params?.offset) searchParams.set("offset", String(params.offset));
    if (params?.sort) searchParams.set("sort", params.sort);
    const qs = searchParams.toString();
    return apiClient.get(`/v1/brain/signals${qs ? `?${qs}` : ""}`);
  },

  searchSignals: async (
    query: string,
    limit?: number
  ): Promise<{ results: { signal: BrainSignal; score: number }[] }> => {
    const searchParams = new URLSearchParams({ q: query });
    if (limit) searchParams.set("limit", String(limit));
    return apiClient.get(`/v1/brain/signals/search?${searchParams.toString()}`);
  },

  deleteSignal: async (signalId: string): Promise<void> => {
    await apiClient.delete(`/v1/brain/signals/${signalId}`);
  },

  purgeSignals: async (): Promise<void> => {
    await apiClient.post("/v1/brain/signals/purge");
  },

  getStats: async (): Promise<BrainStats> => {
    return apiClient.get("/v1/brain/stats");
  },

  submitFeedback: async (data: BrainFeedbackRequest): Promise<void> => {
    await apiClient.post("/v1/brain/feedback", data);
  },

  // Composers
  listComposers: async (): Promise<BrainComposer[]> => {
    const res = await apiClient.get<{ composers: BrainComposer[] }>(
      "/v1/brain/composers"
    );
    return res.composers;
  },

  createComposer: async (
    data: Omit<BrainComposer, "id" | "tenant_id" | "created_at" | "updated_at">
  ): Promise<BrainComposer> => {
    const res = await apiClient.post<{ composer: BrainComposer }>(
      "/v1/brain/composers",
      data
    );
    return res.composer;
  },

  deleteComposer: async (composerId: string): Promise<void> => {
    await apiClient.delete(`/v1/brain/composers/${composerId}`);
  },

  // Triggers
  listTriggers: async (): Promise<BrainTrigger[]> => {
    const res = await apiClient.get<{ triggers: BrainTrigger[] }>(
      "/v1/brain/triggers"
    );
    return res.triggers;
  },

  createTrigger: async (
    data: Omit<BrainTrigger, "id" | "tenant_id" | "created_at" | "updated_at">
  ): Promise<BrainTrigger> => {
    const res = await apiClient.post<{ trigger: BrainTrigger }>(
      "/v1/brain/triggers",
      data
    );
    return res.trigger;
  },

  deleteTrigger: async (triggerId: string): Promise<void> => {
    await apiClient.delete(`/v1/brain/triggers/${triggerId}`);
  },
};
