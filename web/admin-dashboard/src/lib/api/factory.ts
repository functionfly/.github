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

/**
 * Get factory status including totals and latest run
 */
export async function getFactoryStatus(): Promise<FactoryStatus> {
  const response = await adminApiClient.get<FactoryStatus>('/factory/status');
  return response.data;
}

/**
 * Get factory configuration
 */
export async function getFactoryConfig(): Promise<FactoryConfig> {
  const response = await adminApiClient.get<FactoryConfig>('/factory/config');
  return response.data;
}

/**
 * Update factory configuration
 */
export async function updateFactoryConfig(config: Partial<FactoryConfig>): Promise<FactoryConfig> {
  const response = await adminApiClient.patch<FactoryConfig>('/factory/config', config);
  return response.data;
}

/**
 * Trigger a manual pipeline run
 */
export async function triggerPipelineRun(): Promise<PipelineRunResponse> {
  const response = await adminApiClient.post<PipelineRunResponse>('/factory/pipeline/run');
  return response.data;
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
  return response.data;
}

/**
 * Approve an opportunity
 */
export async function approveOpportunity(id: string): Promise<ReviewDecisionResponse> {
  const response = await adminApiClient.post<ReviewDecisionResponse>(`/factory/opportunities/${id}/approve`);
  return response.data;
}

/**
 * Reject an opportunity with a reason
 */
export async function rejectOpportunity(id: string, reason: string): Promise<ReviewDecisionResponse> {
  const response = await adminApiClient.post<ReviewDecisionResponse>(`/factory/opportunities/${id}/reject`, {
    reason,
  });
  return response.data;
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
