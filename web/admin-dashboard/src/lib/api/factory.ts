/**
 * Factory API Client
 * Handles AI Function Factory API communication
 */

import { adminApiClient } from './adminClient';

// ============================================================================
// Types matching backend DTOs
// ============================================================================

export interface FactoryConfig {
  agent_id: string;
  discovery_batch_size: number;
  minimum_quality_score: number;
  minimum_test_score: number;
  require_all_tests_pass: boolean;
  auto_publish: boolean;
  max_opportunities_per_run: number;
  retry_attempts: number;
  retry_backoff_ms: number;
  schedule_enabled?: boolean;
  schedule_cron?: string;
  schedule_timezone?: string;
  notification_webhook_url?: string;
  rate_limit_per_hour?: number;
  max_concurrent_runs?: number;
  dry_run_mode?: boolean;
  discovery_sources?: string[];
  feature_flags?: Record<string, boolean>;
  approval_required_above_quality?: number;
  approval_required_above_test?: number;
  log_level?: string;
  notify_on_failure?: boolean;
  notify_on_review_required?: boolean;
  discovery_cooldown_minutes?: number;
  max_versions_per_function?: number;
}

export interface FactoryTotals {
  runs: number;
  versions: number;
  published: number;
  opportunities: number;
  autopublish: boolean;
  quality_minimum: number;
  test_minimum: number;
}

export interface FactoryRun {
  id: string;
  status: 'running' | 'completed' | 'failed';
  started_at: string;
  completed_at?: string;
  opportunities_found: number;
  opportunities_approved: number;
  opportunities_rejected: number;
  functions_created: number;
  functions_failed: number;
  error?: string;
}

export interface FactoryStatus {
  agent_id: string;
  config: FactoryConfig;
  totals: FactoryTotals;
  latest_run?: FactoryRun;
}

export interface Opportunity {
  id: string;
  source: string;
  status: string;
  review_status: string;
  quality_score?: number;
  test_score?: number;
  title: string;
  description: string;
  created_at: string;
  updated_at: string;
  review_requested_at?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  review_decision?: 'approved' | 'rejected';
  review_reason?: string;
}

export interface OpportunityListResponse {
  opportunities: Opportunity[];
  total: number;
  limit: number;
  offset: number;
}

export interface PendingReview {
  id: string;
  source: string;
  status: string;
  review_status: string;
  quality_score?: number;
  test_score?: number;
  title: string;
  description: string;
  created_at: string;
  review_requested_at?: string;
}

export interface PendingReviewListResponse {
  reviews: PendingReview[];
  total: number;
  limit: number;
  offset: number;
}

export interface PipelineRunResponse {
  run: FactoryRun;
  error?: string;
}

export interface ReviewDecisionResponse {
  message: string;
  id: string;
}

// ============================================================================
// Factory API Functions
// ============================================================================

/** Unwrap admin API response (backend may send payload at root or under .data) */
function unwrap<T>(response: { data?: T } | T): T {
  return (response as { data?: T }).data !== undefined
    ? (response as { data: T }).data
    : (response as T);
}

/**
 * Get factory status including totals and latest run
 */
export async function getFactoryStatus(): Promise<FactoryStatus> {
  const response = await adminApiClient.get<FactoryStatus>('/factory/status');
  return unwrap(response);
}

/**
 * Get factory configuration
 */
export async function getFactoryConfig(): Promise<FactoryConfig> {
  const response = await adminApiClient.get<FactoryConfig>('/factory/config');
  return unwrap(response);
}

/**
 * Update factory configuration
 */
export async function updateFactoryConfig(config: Partial<FactoryConfig>): Promise<FactoryConfig> {
  const response = await adminApiClient.patch<FactoryConfig>('/factory/config', config);
  return unwrap(response);
}

/**
 * Trigger a manual pipeline run
 */
export async function triggerPipelineRun(): Promise<PipelineRunResponse> {
  const response = await adminApiClient.post<PipelineRunResponse>('/factory/pipeline/run');
  return unwrap(response);
}

/**
 * List pending reviews
 */
export async function listPendingReviews(params?: {
  source?: string;
  limit?: number;
  offset?: number;
}): Promise<PendingReviewListResponse> {
  const searchParams = new URLSearchParams();
  if (params?.source) searchParams.set('source', params.source);
  if (params?.limit) searchParams.set('limit', String(params.limit));
  if (params?.offset) searchParams.set('offset', String(params.offset));

  const queryString = searchParams.toString();
  const response = await adminApiClient.get<PendingReviewListResponse>(
    `/factory/reviews/pending${queryString ? `?${queryString}` : ''}`
  );
  return unwrap(response);
}

/**
 * Approve an opportunity
 */
export async function approveOpportunity(id: string): Promise<ReviewDecisionResponse> {
  const response = await adminApiClient.post<ReviewDecisionResponse>(`/factory/opportunities/${id}/approve`);
  return unwrap(response);
}

/**
 * Reject an opportunity with a reason
 */
export async function rejectOpportunity(id: string, reason: string): Promise<ReviewDecisionResponse> {
  const response = await adminApiClient.post<ReviewDecisionResponse>(`/factory/opportunities/${id}/reject`, {
    reason,
  });
  return unwrap(response);
}

// ============================================================================
// Export all functions
// ============================================================================

export const factoryApi = {
  getStatus: getFactoryStatus,
  getConfig: getFactoryConfig,
  updateConfig: updateFactoryConfig,
  triggerPipelineRun,
  listPendingReviews,
  approveOpportunity,
  rejectOpportunity,
};

export default factoryApi;
