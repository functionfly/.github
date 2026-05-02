import { apiClient } from "./client";
import type {
  SimpleState,
  StateValue,
  StateHistoryEntry,
  StateSnapshot,
  StatePermission,
  CreateStateRequest,
  UpdateStateRequest,
  PatchStateRequest,
  CreateSnapshotRequest,
  GrantPermissionRequest,
  TimeTravelRequest,
  TimeTravelResponse,
} from "@/types";

export const stateApi = {
  // List states
  list: async (params?: {
    prefix?: string;
    limit?: number;
    offset?: number;
  }): Promise<SimpleState[]> => {
    const queryParams = new URLSearchParams();
    if (params?.prefix) queryParams.append("prefix", params.prefix);
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());

    const data = await apiClient.get<{ states: SimpleState[]; total: number; limit: number; offset: number }>(`/v1/state?${queryParams.toString()}`);
    return data.states;
  },

  // Get state by path
  get: async (path: string): Promise<SimpleState> => {
    return apiClient.get<SimpleState>(`/v1/state/${encodeURIComponent(path)}`);
  },

  // Create state
  create: async (data: CreateStateRequest): Promise<SimpleState> => {
    return apiClient.post<SimpleState>("/v1/state", data);
  },

  // Update state
  update: async (path: string, data: UpdateStateRequest): Promise<SimpleState> => {
    return apiClient.put<SimpleState>(`/v1/state/${encodeURIComponent(path)}`, data);
  },

  // Delete state
  delete: async (path: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/state/${encodeURIComponent(path)}`);
  },

  // Get value
  getValue: async (path: string): Promise<{ value: StateValue }> => {
    return apiClient.get<{ value: StateValue }>(`/v1/state/${encodeURIComponent(path)}/value`);
  },

  // Set value
  setValue: async (
    path: string,
    data: { value: StateValue; ttl?: number }
  ): Promise<{ value: StateValue }> => {
    return apiClient.put<{ value: StateValue }>(`/v1/state/${encodeURIComponent(path)}/value`, data);
  },

  // Patch value
  patchValue: async (
    path: string,
    data: { operations: PatchStateRequest["operations"] }
  ): Promise<{ value: StateValue }> => {
    return apiClient.patch<{ value: StateValue }>(`/v1/state/${encodeURIComponent(path)}/value`, data);
  },

  // Delete value
  deleteValue: async (path: string): Promise<void> => {
    await apiClient.delete<void>(`/v1/state/${encodeURIComponent(path)}/value`);
  },

  // Get history
  getHistory: async (
    path: string,
    params?: {
      limit?: number;
      offset?: number;
      startTime?: string;
      endTime?: string;
    }
  ): Promise<{ history: StateHistoryEntry[]; total: number }> => {
    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.offset) queryParams.append("offset", params.offset.toString());
    if (params?.startTime) queryParams.append("startTime", params.startTime);
    if (params?.endTime) queryParams.append("endTime", params.endTime);

    return apiClient.get<{ history: StateHistoryEntry[]; total: number }>(
      `/v1/state/${encodeURIComponent(path)}/history?${queryParams.toString()}`
    );
  },

  // Create snapshot
  createSnapshot: async (
    path: string,
    data: CreateSnapshotRequest
  ): Promise<StateSnapshot> => {
    return apiClient.post<StateSnapshot>(`/v1/state/${encodeURIComponent(path)}/snapshot`, data);
  },

  // List snapshots
  listSnapshots: async (path: string): Promise<StateSnapshot[]> => {
    return apiClient.get<StateSnapshot[]>(`/v1/state/${encodeURIComponent(path)}/snapshots`);
  },

  // Restore snapshot
  restoreSnapshot: async (
    path: string,
    data: { snapshotId: string }
  ): Promise<SimpleState> => {
    return apiClient.post<SimpleState>(`/v1/state/${encodeURIComponent(path)}/restore`, data);
  },

  // Time travel
  timeTravel: async (data: TimeTravelRequest): Promise<TimeTravelResponse> => {
    const queryParams = new URLSearchParams();
    if (data.timestamp) queryParams.append("timestamp", data.timestamp);
    if (data.version) queryParams.append("version", data.version.toString());

    return apiClient.get<TimeTravelResponse>(
      `/v1/state/${encodeURIComponent(data.path)}/time-travel?${queryParams.toString()}`
    );
  },

  // Get permissions
  getPermissions: async (path: string): Promise<StatePermission[]> => {
    return apiClient.get<StatePermission[]>(`/v1/state/${encodeURIComponent(path)}/permissions`);
  },

  // Grant permission
  grantPermission: async (
    path: string,
    data: GrantPermissionRequest
  ): Promise<StatePermission> => {
    return apiClient.post<StatePermission>(
      `/v1/state/${encodeURIComponent(path)}/permissions`,
      data
    );
  },
};
