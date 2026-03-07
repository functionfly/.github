export interface FunctionFlyConfig {
  apiKey: string;
  baseUrl?: string;
  tenantId?: string;
}

export interface StateOptions {
  path: string;
  key?: string;
}

export interface SetValueOptions {
  value: Record<string, unknown>;
  ttlDays?: number;
}

export interface StateValue {
  id: string;
  state_id: string;
  key: string;
  value: Record<string, unknown>;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface State {
  id: string;
  tenant_id: string;
  name: string;
  full_path: string;
  storage_type: string;
  ttl_days: number;
  max_size_mb: number;
  is_versioned: boolean;
  is_encrypted: boolean;
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

export interface StateListResponse {
  states: State[];
  total: number;
  limit: number;
  offset: number;
}

export interface HistoryEvent {
  id: string;
  state_id: string;
  key: string;
  event_type: string;
  timestamp: string;
  correlation_id: string;
}

export interface HistoryResponse {
  events: HistoryEvent[];
  total: number;
  limit: number;
  offset: number;
}

export interface Snapshot {
  id: string;
  state_id: string;
  version: number;
  label: string;
  created_at: string;
}

export interface SnapshotListResponse {
  snapshots: Snapshot[];
  total: number;
  limit: number;
  offset: number;
}

export interface Permission {
  id: string;
  state_id: string;
  principal_type: string;
  principal_id: string;
  can_read: boolean;
  can_write: boolean;
  can_delete: boolean;
  can_admin: boolean;
  can_trigger: boolean;
  created_at: string;
}

export interface GrantPermissionRequest {
  principal_type: 'user' | 'team' | 'function';
  principal_id: string;
  can_read?: boolean;
  can_write?: boolean;
  can_delete?: boolean;
  can_admin?: boolean;
  can_trigger?: boolean;
}

export interface MemoryOptions {
  agentId: string;
  memoryType?: 'working' | 'longterm' | 'context' | 'episodic';
}

export interface CreateMemoryRequest {
  agent_id: string;
  memory_type: string;
  content: string;
  structured_data?: Record<string, unknown>;
  embedding?: number[];
  importance_score?: number;
  ttl_days?: number;
}

export interface AgentMemory {
  id: string;
  tenant_id: string;
  agent_id: string;
  memory_type: string;
  content: string;
  structured_data: Record<string, unknown>;
  importance_score: number;
  access_count: number;
  ttl_days: number;
  created_at: string;
  updated_at: string;
}

export interface MemoryListResponse {
  memories: AgentMemory[];
  total: number;
  limit: number;
  offset: number;
}

export interface SearchMemoryRequest {
  query?: string;
  embedding?: number[];
  memory_type?: string;
  limit?: number;
  threshold?: number;
}

export interface SearchMemoryResponse {
  memories: AgentMemory[];
  count: number;
}

export interface Trigger {
  id: string;
  tenant_id: string;
  source_state_id: string;
  trigger_type: string;
  key_pattern: string | null;
  condition: Record<string, unknown> | null;
  target_function: string;
  include_previous: boolean;
  include_new: boolean;
  max_invocations_per_minute: number;
  is_active: boolean;
  created_at: string;
}

export interface TimeTravelQuery {
  at: string;
  key?: string;
}

export interface TimeTravelResponse {
  timestamp: string;
  data: Record<string, unknown>;
}

export interface PatchOperation {
  op: 'add' | 'replace' | 'remove' | 'test';
  path: string;
  value?: unknown;
}

export interface PatchValueRequest {
  patch: PatchOperation[];
}

export interface PatchValueResponse {
  value: StateValue;
  previous: Record<string, unknown>;
  applied_ops: number;
}

export interface ApiError {
  error: string;
  message: string;
  status_code: number;
}

export type FetchImpl = typeof fetch;
