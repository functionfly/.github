import { apiClient } from "./client";
import type {
  StateFabric,
  StateFabricStore,
  Pipeline,
  EventLog,
  Snapshot,
  ReplaySession,
  CreateStateFabricRequest,
  UpdateStateFabricRequest,
  CreatePipelineRequest,
  UpdatePipelineRequest,
  CreateStoreRequest,
  StateFabricMetrics,
} from "@/types";

export const stateFabricApi = {
  // State Fabric CRUD
  list: async (): Promise<StateFabric[]> => {
    return apiClient.get<StateFabric[]>("/v1/state-fabrics");
  },

  get: async (id: string): Promise<StateFabric> => {
    return apiClient.get<StateFabric>(`/v1/state-fabrics/${id}`);
  },

  create: async (data: CreateStateFabricRequest): Promise<StateFabric> => {
    return apiClient.post<StateFabric>("/v1/state-fabrics", data);
  },

  update: async (id: string, data: UpdateStateFabricRequest): Promise<StateFabric> => {
    return apiClient.patch<StateFabric>(`/v1/state-fabrics/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/state-fabrics/${id}`);
  },

  // Metrics
  getMetrics: async (id: string, timeRange?: string): Promise<StateFabricMetrics> => {
    const params = timeRange ? `?timeRange=${timeRange}` : "";
    return apiClient.get<StateFabricMetrics>(`/v1/state-fabrics/${id}/metrics${params}`);
  },

  // Stores
  listStores: async (fabricId: string): Promise<StateFabricStore[]> => {
    return apiClient.get<StateFabricStore[]>(`/v1/state-fabrics/${fabricId}/stores`);
  },

  createStore: async (fabricId: string, data: CreateStoreRequest): Promise<StateFabricStore> => {
    return apiClient.post<StateFabricStore>(`/v1/state-fabrics/${fabricId}/stores`, data);
  },

  deleteStore: async (fabricId: string, storeId: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/state-fabrics/${fabricId}/stores/${storeId}`);
  },

  // Pipelines
  listPipelines: async (fabricId: string): Promise<Pipeline[]> => {
    return apiClient.get<Pipeline[]>(`/v1/state-fabrics/${fabricId}/pipelines`);
  },

  createPipeline: async (fabricId: string, data: CreatePipelineRequest): Promise<Pipeline> => {
    return apiClient.post<Pipeline>(`/v1/state-fabrics/${fabricId}/pipelines`, data);
  },

  updatePipeline: async (
    fabricId: string,
    pipelineId: string,
    data: UpdatePipelineRequest
  ): Promise<Pipeline> => {
    return apiClient.patch<Pipeline>(
      `/v1/state-fabrics/${fabricId}/pipelines/${pipelineId}`,
      data
    );
  },

  deletePipeline: async (fabricId: string, pipelineId: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/state-fabrics/${fabricId}/pipelines/${pipelineId}`);
  },

  executePipeline: async (
    fabricId: string,
    pipelineId: string,
    input: Record<string, any>
  ): Promise<{ executionId: string; status: string }> => {
    return apiClient.post<{ executionId: string; status: string }>(
      `/v1/state-fabrics/${fabricId}/pipelines/${pipelineId}/execute`,
      input
    );
  },

  // Event Logs
  listEventLogs: async (
    fabricId: string,
    params?: {
      storeId?: string;
      eventType?: string;
      startTime?: string;
      endTime?: string;
      limit?: number;
      offset?: number;
    }
  ): Promise<{ events: EventLog[]; total: number }> => {
    const queryParams = new URLSearchParams();
    if (params?.storeId) queryParams.append("storeId", params.storeId);
    if (params?.eventType) queryParams.append("eventType", params.eventType);
    if (params?.startTime) queryParams.append("startTime", params.startTime);
    if (params?.endTime) queryParams.append("endTime", params.endTime);
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());

    return apiClient.get<{ events: EventLog[]; total: number }>(
      `/v1/state-fabrics/${fabricId}/events?${queryParams.toString()}`
    );
  },

  // Snapshots
  listSnapshots: async (fabricId: string, storeId?: string): Promise<Snapshot[]> => {
    const params = storeId ? `?storeId=${storeId}` : "";
    return apiClient.get<Snapshot[]>(`/v1/state-fabrics/${fabricId}/snapshots${params}`);
  },

  createSnapshot: async (
    fabricId: string,
    data: {
      name: string;
      description?: string;
      storeId?: string;
    }
  ): Promise<Snapshot> => {
    return apiClient.post<Snapshot>(`/v1/state-fabrics/${fabricId}/snapshots`, data);
  },

  deleteSnapshot: async (fabricId: string, snapshotId: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/state-fabrics/${fabricId}/snapshots/${snapshotId}`);
  },

  // Replay
  createReplay: async (
    fabricId: string,
    data: {
      snapshotId?: string;
      startEventId?: string;
      endEventId?: string;
    }
  ): Promise<ReplaySession> => {
    return apiClient.post<ReplaySession>(`/v1/state-fabrics/${fabricId}/replays`, data);
  },

  getReplay: async (fabricId: string, replayId: string): Promise<ReplaySession> => {
    return apiClient.get<ReplaySession>(`/v1/state-fabrics/${fabricId}/replays/${replayId}`);
  },

  listReplays: async (fabricId: string): Promise<ReplaySession[]> => {
    return apiClient.get<ReplaySession[]>(`/v1/state-fabrics/${fabricId}/replays`);
  },
};

// Admin API for State Fabric management
export const adminStateFabricApi = {
  listAll: async (params?: {
    tenantId?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ fabrics: StateFabric[]; total: number }> => {
    const queryParams = new URLSearchParams();
    if (params?.tenantId) queryParams.append("tenantId", params.tenantId);
    if (params?.status) queryParams.append("status", params.status);
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());

    return apiClient.get<{ fabrics: StateFabric[]; total: number }>(
      `/v1/admin/state-fabrics?${queryParams.toString()}`
    );
  },

  getStats: async (): Promise<{
    totalFabrics: number;
    activeFabrics: number;
    totalStores: number;
    totalPipelines: number;
    totalEvents: number;
    storageUsed: number;
  }> => {
    return apiClient.get<{
      totalFabrics: number;
      activeFabrics: number;
      totalStores: number;
      totalPipelines: number;
      totalEvents: number;
      storageUsed: number;
    }>("/v1/admin/state-fabrics/stats");
  },

  suspendFabric: async (fabricId: string, reason: string): Promise<void> => {
    await apiClient.post<void>(`/v1/admin/state-fabrics/${fabricId}/suspend`, { reason });
  },

  resumeFabric: async (fabricId: string): Promise<void> => {
    await apiClient.post<void>(`/v1/admin/state-fabrics/${fabricId}/resume`, {});
  },
};
