import { apiClient } from './client';

// ==================== Types ====================

export interface MicroVMExecution {
  id: string;
  tenant_id: string;
  function_id: string;
  function_version: string;
  execution_id: string;
  started_at: string;
  completed_at: string | null;
  duration_ms: number;
  memory_mb: number;
  vcpus: number;
  status: 'running' | 'completed' | 'failed' | 'timeout';
  outcome: string | null;
  error_message: string | null;
  network_allowed: boolean;
  packages_cached: boolean;
  created_at: string;
}

export interface MicroVMStats {
  total_executions: number;
  running_vms: number;
  avg_duration_ms: number;
  success_rate: number;
  total_compute_seconds: number;
}

export interface MicroVMQuota {
  tenant_id: string;
  max_concurrent_vms: number;
  max_memory_mb: number;
  max_vcpus: number;
  max_timeout_ms: number;
  current_compute_usage: number;
  current_memory_usage: number;
  updated_at: string;
}

export interface MicroVMBillingRecord {
  id: string;
  tenant_id: string;
  billing_period: string;
  total_executions: number;
  total_compute_seconds: number;
  total_memory_seconds: number;
  avg_memory_mb: number;
  avg_vcpus: number;
  base_fee_cents: number;
  compute_charge_cents: number;
  memory_charge_cents: number;
  total_charge_cents: number;
  created_at: string;
  updated_at: string;
}

export interface MicroVMLimits {
  max_micro_vms: number;
  max_concurrent_vms: number; // Maximum concurrent VMs (0 = use max_micro_vms as limit)
  default_memory_mb: number;
  max_memory_mb: number;
  default_vcpu: number;
  max_vcpu: number;
  default_timeout: number;
  max_timeout: number;
  included_budget?: number; // Included compute budget in cents (for MicroVM Enterprise)
}

export interface MicroVMAuditLog {
  id: string;
  tenant_id: string;
  user_id: string | null;
  action: string;
  resource_type: string;
  resource_id: string | null;
  details: Record<string, unknown>;
  ip_address: string | null;
  user_agent: string | null;
  created_at: string;
}

export interface MicroVMUsageResponse {
  stats: MicroVMStats;
  quota: MicroVMQuota | null;
  executions: MicroVMExecution[];
  current_plan: string;
}

export interface MicroVMBillingResponse {
  current: MicroVMBillingRecord | null;
  history: MicroVMBillingRecord[];
  limits: MicroVMLimits;
}

export interface MicroVMAuditResponse {
  logs: MicroVMAuditLog[];
  total: number;
}

// ==================== API Functions ====================

export async function getMicroVMUsage(): Promise<MicroVMUsageResponse> {
  return apiClient.get<MicroVMUsageResponse>('/v1/microvm/usage');
}

export async function getMicroVMQuota(): Promise<MicroVMQuota> {
  return apiClient.get<MicroVMQuota>('/v1/microvm/quota');
}

export async function getMicroVMBilling(): Promise<MicroVMBillingResponse> {
  return apiClient.get<MicroVMBillingResponse>('/v1/microvm/billing');
}

export async function aggregateMicroVMBilling(): Promise<MicroVMBillingRecord> {
  return apiClient.post<MicroVMBillingRecord>('/v1/microvm/billing/aggregate');
}

export async function getMicroVMAudit(params?: {
  limit?: number;
  offset?: number;
}): Promise<MicroVMAuditResponse> {
  return apiClient.get<MicroVMAuditResponse>('/v1/microvm/audit', { params });
}

export async function createMicroVMExecution(params: {
  function_id: string;
  function_version: string;
  execution_id: string;
  memory_mb: number;
  vcpus: number;
  network_allowed: boolean;
  packages_cached: boolean;
}): Promise<MicroVMExecution> {
  return apiClient.post<MicroVMExecution>('/v1/microvm/executions', params);
}

export async function updateMicroVMExecution(
  id: string,
  params: {
    status: 'completed' | 'failed' | 'timeout';
    outcome?: string;
    error_message?: string;
    duration_ms: number;
  }
): Promise<void> {
  await apiClient.patch(`/v1/microvm/executions/${id}`, params);
}
