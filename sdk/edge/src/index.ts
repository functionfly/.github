import {
  FunctionFlyConfig,
  StateOptions,
  SetValueOptions,
  StateValue,
  State,
  StateListResponse,
  HistoryResponse,
  SnapshotListResponse,
  Permission,
  GrantPermissionRequest,
  MemoryOptions,
  CreateMemoryRequest,
  AgentMemory,
  MemoryListResponse,
  SearchMemoryRequest,
  SearchMemoryResponse,
  Trigger,
  TimeTravelQuery,
  TimeTravelResponse,
  PatchValueRequest,
  PatchValueResponse,
  FetchImpl,
} from './types';

const DEFAULT_BASE_URL = 'https://api.functionfly.com';

export class FunctionFlyClient {
  private readonly apiKey: string;
  private readonly baseUrl: string;
  private readonly tenantId?: string;
  private readonly fetch: FetchImpl;

  constructor(config: FunctionFlyConfig, fetchImpl?: FetchImpl) {
    this.apiKey = config.apiKey;
    this.baseUrl = config.baseUrl || DEFAULT_BASE_URL;
    this.tenantId = config.tenantId;
    this.fetch = fetchImpl || globalThis.fetch.bind(globalThis);
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
    queryParams?: Record<string, string>
  ): Promise<T> {
    const url = new URL(`${this.baseUrl}/v1${path}`);
    
    if (queryParams) {
      Object.entries(queryParams).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }

    const headers: HeadersInit = {
      'Authorization': `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json',
    };

    if (this.tenantId) {
      headers['X-Tenant-ID'] = this.tenantId;
    }

    const options: RequestInit = {
      method,
      headers,
    };

    if (body) {
      options.body = JSON.stringify(body);
    }

    const response = await this.fetch(url.toString(), options);

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error', message: response.statusText }));
      throw new Error(`API Error: ${error.error || error.message} (${response.status})`);
    }

    return response.json();
  }

  // State Operations
  
  async listStates(limit = 20, offset = 0): Promise<StateListResponse> {
    return this.request<StateListResponse>('GET', '/state', undefined, {
      limit: String(limit),
      offset: String(offset),
    });
  }

  async createState(state: Partial<State>): Promise<State> {
    return this.request<State>('POST', '/state', state);
  }

  async getState(path: string): Promise<State> {
    return this.request<State>('GET', `/state/${encodeURIComponent(path)}`);
  }

  async updateState(path: string, updates: Partial<State>): Promise<State> {
    return this.request<State>('PUT', `/state/${encodeURIComponent(path)}`, updates);
  }

  async deleteState(path: string): Promise<void> {
    return this.request<void>('DELETE', `/state/${encodeURIComponent(path)}`);
  }

  async getValue(path: string, key?: string): Promise<StateValue> {
    const queryParams = key ? { key } : undefined;
    return this.request<StateValue>('GET', `/state/${encodeURIComponent(path)}/value`, undefined, queryParams);
  }

  async setValue(path: string, value: Record<string, unknown>, key?: string, ttlDays?: number): Promise<StateValue> {
    const queryParams: Record<string, string> = {};
    if (key) queryParams.key = key;
    
    return this.request<StateValue>('PUT', `/state/${encodeURIComponent(path)}/value`, {
      value,
      ...(ttlDays && { expires_in_days: ttlDays }),
    }, queryParams);
  }

  async patchValue(path: string, patch: PatchValueRequest['patch'], key?: string): Promise<PatchValueResponse> {
    const queryParams = key ? { key } : undefined;
    return this.request<PatchValueResponse>('PATCH', `/state/${encodeURIComponent(path)}/value`, { patch }, queryParams);
  }

  async deleteValue(path: string, key?: string): Promise<void> {
    const queryParams = key ? { key } : undefined;
    return this.request<void>('DELETE', `/state/${encodeURIComponent(path)}/value`, undefined, queryParams);
  }

  async getHistory(path: string, key?: string, limit = 50, offset = 0): Promise<HistoryResponse> {
    const queryParams: Record<string, string> = { limit: String(limit), offset: String(offset) };
    if (key) queryParams.key = key;
    return this.request<HistoryResponse>('GET', `/state/${encodeURIComponent(path)}/history`, undefined, queryParams);
  }

  async createSnapshot(path: string, label?: string): Promise<Snapshot> {
    return this.request<Snapshot>('POST', `/state/${encodeURIComponent(path)}/snapshot`, { label });
  }

  async listSnapshots(path: string, limit = 20, offset = 0): Promise<SnapshotListResponse> {
    return this.request<SnapshotListResponse>('GET', `/state/${encodeURIComponent(path)}/snapshots`, undefined, {
      limit: String(limit),
      offset: String(offset),
    });
  }

  async restoreSnapshot(path: string, snapshotVersion: number): Promise<void> {
    return this.request<void>('POST', `/state/${encodeURIComponent(path)}/restore`, { snapshot_version: snapshotVersion });
  }

  async timeTravel(path: string, timestamp: string, key?: string): Promise<TimeTravelResponse> {
    const queryParams: Record<string, string> = { at: timestamp };
    if (key) queryParams.key = key;
    return this.request<TimeTravelResponse>('GET', `/state/${encodeURIComponent(path)}/time-travel`, undefined, queryParams);
  }

  // Permission Operations

  async getPermissions(path: string): Promise<Permission[]> {
    return this.request<Permission[]>('GET', `/state/${encodeURIComponent(path)}/permissions`);
  }

  async grantPermission(path: string, permission: GrantPermissionRequest): Promise<Permission> {
    return this.request<Permission>('POST', `/state/${encodeURIComponent(path)}/permissions`, permission);
  }

  // Trigger Operations

  async listTriggers(statePath?: string, limit = 20, offset = 0): Promise<{ triggers: Trigger[]; total: number }> {
    const queryParams: Record<string, string> = { limit: String(limit), offset: String(offset) };
    if (statePath) queryParams.state = statePath;
    return this.request<{ triggers: Trigger[]; total: number }>('GET', '/triggers', undefined, queryParams);
  }

  async createTrigger(trigger: {
    state_path: string;
    trigger_type: string;
    key_pattern?: string;
    target_function: string;
    include_previous?: boolean;
    include_new?: boolean;
    max_invocations_per_minute?: number;
    is_active?: boolean;
  }): Promise<Trigger> {
    return this.request<Trigger>('POST', '/triggers', trigger);
  }

  async deleteTrigger(triggerId: string): Promise<void> {
    return this.request<void>('DELETE', `/triggers/${encodeURIComponent(triggerId)}`);
  }

  // Memory Operations

  async createMemory(memory: CreateMemoryRequest): Promise<AgentMemory> {
    return this.request<AgentMemory>('POST', '/memories', memory);
  }

  async getMemory(memoryId: string): Promise<AgentMemory> {
    return this.request<AgentMemory>('GET', `/memories/${encodeURIComponent(memoryId)}`);
  }

  async updateMemory(memoryId: string, updates: Partial<CreateMemoryRequest>): Promise<AgentMemory> {
    return this.request<AgentMemory>('PATCH', `/memories/${encodeURIComponent(memoryId)}`, updates);
  }

  async deleteMemory(memoryId: string): Promise<void> {
    return this.request<void>('DELETE', `/memories/${encodeURIComponent(memoryId)}`);
  }

  async listMemories(agentId?: string, memoryType?: string, limit = 20, offset = 0): Promise<MemoryListResponse> {
    const queryParams: Record<string, string> = { limit: String(limit), offset: String(offset) };
    if (agentId) queryParams.agent_id = agentId;
    if (memoryType) queryParams.memory_type = memoryType;
    return this.request<MemoryListResponse>('GET', '/memories', undefined, queryParams);
  }

  async searchMemories(request: SearchMemoryRequest): Promise<SearchMemoryResponse> {
    return this.request<SearchMemoryResponse>('POST', '/memories/search', request);
  }
}

// Convenience class for simpler state operations
export class StateClient {
  private readonly client: FunctionFlyClient;
  private readonly path: string;

  constructor(client: FunctionFlyClient, path: string) {
    this.client = client;
    this.path = path;
  }

  async get(key?: string): Promise<StateValue> {
    return this.client.getValue(this.path, key);
  }

  async set(value: Record<string, unknown>, key?: string, ttlDays?: number): Promise<StateValue> {
    return this.client.setValue(this.path, value, key, ttlDays);
  }

  async patch(patch: PatchValueRequest['patch'], key?: string): Promise<PatchValueResponse> {
    return this.client.patchValue(this.path, patch, key);
  }

  async delete(key?: string): Promise<void> {
    return this.client.deleteValue(this.path, key);
  }

  async history(key?: string, limit = 50, offset = 0): Promise<HistoryResponse> {
    return this.client.getHistory(this.path, key, limit, offset);
  }

  async snapshot(label?: string): Promise<Snapshot> {
    return this.client.createSnapshot(this.path, label);
  }

  async snapshots(limit = 20, offset = 0): Promise<SnapshotListResponse> {
    return this.client.listSnapshots(this.path, limit, offset);
  }

  async restore(snapshotVersion: number): Promise<void> {
    return this.client.restoreSnapshot(this.path, snapshotVersion);
  }

  async at(timestamp: string, key?: string): Promise<TimeTravelResponse> {
    return this.client.timeTravel(this.path, timestamp, key);
  }

  async permissions(): Promise<Permission[]> {
    return this.client.getPermissions(this.path);
  }

  async grantPermission(permission: GrantPermissionRequest): Promise<Permission> {
    return this.client.grantPermission(this.path, permission);
  }
}

// Factory function for creating state clients
export function createClient(config: FunctionFlyConfig): FunctionFlyClient {
  return new FunctionFlyClient(config);
}

export function state(client: FunctionFlyClient, path: string): StateClient {
  return new StateClient(client, path);
}

// Default export
export default {
  FunctionFlyClient,
  StateClient,
  createClient,
  state,
};
