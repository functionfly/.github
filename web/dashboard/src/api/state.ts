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

  // Enable encryption
  enableEncryption: async (path: string): Promise<{ enabled: boolean }> => {
    return apiClient.post<{ enabled: boolean }>(
      `/v1/state/${encodeURIComponent(path)}/encryption`,
      {}
    );
  },

  // Get encryption statistics (tenant-level, no path needed)
  getEncryptionStats: async (): Promise<{
    total_states: number;
    encrypted_states: number;
    unencrypted_states: number;
    total_values: number;
    encrypted_values: number;
    unencrypted_values: number;
    encryption_enabled: boolean;
  }> => {
    return apiClient.get<{
      total_states: number;
      encrypted_states: number;
      unencrypted_states: number;
      total_values: number;
      encrypted_values: number;
      unencrypted_values: number;
      encryption_enabled: boolean;
    }>(`/v1/state/encryption-stats`);
  },

  // Migrate encryption (encrypt existing values)
  migrateEncryption: async (
    path: string,
    data: { state_id?: string; batch_size?: number; dry_run?: boolean; force_rotate?: boolean }
  ): Promise<{
    states_processed: number;
    values_encrypted: number;
    values_skipped: number;
    errors: string[];
    completed: boolean;
  }> => {
    return apiClient.post<{
      states_processed: number;
      values_encrypted: number;
      values_skipped: number;
      errors: string[];
      completed: boolean;
    }>(`/v1/state/encrypt`, { state_id: path, ...data });
  },

  // Rotate encryption key
  rotateEncryptionKey: async (): Promise<{ rotated: boolean; keyId: string }> => {
    return apiClient.post<{ rotated: boolean; keyId: string }>(
      `/v1/state/rotate-key`,
      {}
    );
  },
};
