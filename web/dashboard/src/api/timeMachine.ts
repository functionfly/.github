import { apiClient } from "./client";

export interface TimeMachineLimits {
  replay_window_hours: number;
  max_executions_per_replay: number;
  max_concurrent_replays: number;
  data_retention_days: number;
  auto_reconciliation: boolean;
  live_reconciliation: boolean;
  audit_certificates: boolean;
  replay_scheduling: boolean;
  full_diff_reports: boolean;
  incident_insurance: boolean;
  unlimited: boolean;
}

export interface ReplayJob {
  id: string;
  tenant_id: string;
  user_id: string;
  function_id: string;
  window_start: string;
  window_end: string;
  target_version_id: string;
  target_version: string;
  max_executions: number;
  reconciliation_mode: string;
  auto_reconcile: boolean;
  status: string;
  progress_percent: number;
  current_phase: string | null;
  error_message: string | null;
  total_executions_found: number;
  total_executions_replayed: number;
  total_executions_changed: number;
  total_executions_failed: number;
  reason: string;
  incident_url: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateReplayRequest {
  function_id: string;
  window_start: string;
  window_end: string;
  target_version_id: string;
  reason: string;
  incident_url?: string;
  reconciliation_mode?: string;
  max_executions?: number;
}

export interface ReplayItem {
  id: string;
  replay_id: string;
  original_execution_id: string;
  original_input: unknown;
  original_output: unknown;
  original_version: string;
  original_duration_ms: number;
  original_timestamp: string;
  original_meg_root_hash: string | null;
  original_certificate_id: string | null;
  new_output: unknown | null;
  new_duration_ms: number | null;
  new_meg_root_hash: string | null;
  new_status_code: number | null;
  output_changed: boolean | null;
  diff_type: string | null;
  diff_summary: string | null;
  diff_detail: unknown | null;
  reconciliation_status: string;
  reconciliation_actions: unknown | null;
  replay_error: string | null;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface DiffSummary {
  replay_id: string;
  total_executions: number;
  identical: number;
  changed: number;
  failed: number;
  breakdown: {
    identical: number;
    minor: number;
    major: number;
    breaking: number;
    error: number;
  };
}

export interface AuditCertificate {
  id: string;
  replay_id: string;
  certificate_id: string;
  cert_json: unknown;
  cert_hash: string;
  previous_cert_hash: string | null;
  merkle_root: string | null;
  signature: string | null;
  compliance_frameworks: string[];
  retention_policy: string;
  anchored: boolean;
  created_at: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface Reconciliation {
  id: string;
  replay_id: string;
  replay_item_id: string;
  action_type: string;
  target_resource: string;
  old_value: unknown;
  new_value: unknown;
  status: string;
  applied_at: string | null;
  error_message: string | null;
  dry_run: boolean;
  reversible: boolean;
  created_at: string;
}

export async function createReplay(req: CreateReplayRequest): Promise<ReplayJob> {
  return apiClient.post<ReplayJob>('/v1/time-machine/replays', req);
}

export async function listReplays(params?: {
  function_id?: string;
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<PaginatedResponse<ReplayJob>> {
  return apiClient.get<PaginatedResponse<ReplayJob>>('/v1/time-machine/replays', { params });
}

export async function getReplay(id: string): Promise<ReplayJob> {
  return apiClient.get<ReplayJob>(`/v1/time-machine/replays/${id}`);
}

export async function cancelReplay(id: string): Promise<void> {
  return apiClient.delete(`/v1/time-machine/replays/${id}`);
}

export async function getReplayItems(
  id: string,
  params?: { limit?: number; offset?: number; diff_type?: string }
): Promise<PaginatedResponse<ReplayItem>> {
  return apiClient.get<PaginatedResponse<ReplayItem>>(
    `/v1/time-machine/replays/${id}/items`,
    { params }
  );
}

export async function getReplayItem(replayId: string, itemId: string): Promise<ReplayItem> {
  return apiClient.get<ReplayItem>(`/v1/time-machine/replays/${replayId}/items/${itemId}`);
}

export async function getDiffSummary(id: string): Promise<DiffSummary> {
  return apiClient.get<DiffSummary>(`/v1/time-machine/replays/${id}/diff-summary`);
}

export async function startReconciliation(
  id: string,
  options?: { dry_run?: boolean }
): Promise<void> {
  return apiClient.post(`/v1/time-machine/replays/${id}/reconcile`, { dry_run: options?.dry_run ?? true });
}

export async function listReconciliations(
  id: string,
  params?: { limit?: number; offset?: number }
): Promise<PaginatedResponse<Reconciliation>> {
  return apiClient.get<PaginatedResponse<Reconciliation>>(
    `/v1/time-machine/replays/${id}/reconciliations`,
    { params }
  );
}

export async function getAuditCertificate(id: string): Promise<AuditCertificate> {
  return apiClient.get<AuditCertificate>(`/v1/time-machine/replays/${id}/audit-certificate`);
}

export async function getTimeMachineLimits(): Promise<TimeMachineLimits> {
  return apiClient.get<TimeMachineLimits>('/v1/time-machine/limits');
}
