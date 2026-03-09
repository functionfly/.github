import { apiClient } from './client';

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

export interface FactoryVersion {
  id: string;
  function_name: string;
  function_description: string;
  version: string;
  quality_score: number;
  test_score: number;
  status: string;
  created_at: string;
  published_at?: string;
}

export interface PublishedFunction {
  id: string;
  function_name: string;
  function_description: string;
  version: string;
  runtime: string;
  provider: string;
  created_at: string;
  published_at: string;
  execution_count: number;
}

export interface FunctionsListResponse {
  versions: FactoryVersion[];
  total_versions: number;
  published_functions?: PublishedFunction[];
  total_published?: number;
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

export interface RejectRequest {
  reason: string;
}

// ============================================================================
// API Functions
// ============================================================================

/**
 * Get factory status including totals and latest run
 */
export async function getFactoryStatus(): Promise<FactoryStatus> {
  return apiClient.get<FactoryStatus>('/factory/status');
}

/**
 * Get factory configuration
 */
export async function getFactoryConfig(): Promise<FactoryConfig> {
  return apiClient.get<FactoryConfig>('/factory/config');
}

/**
 * Update factory configuration
 */
export async function updateFactoryConfig(config: Partial<FactoryConfig>): Promise<FactoryConfig> {
  return apiClient.put<FactoryConfig>('/factory/config', config);
}

/**
 * Trigger a manual pipeline run
 */
export async function triggerPipelineRun(): Promise<PipelineRunResponse> {
  return apiClient.post<PipelineRunResponse>('/factory/pipeline/run');
}

/**
 * List opportunities with optional filters
 */
export async function listOpportunities(params?: {
  status?: string;
  source?: string;
  limit?: number;
  offset?: number;
}): Promise<OpportunityListResponse> {
  const searchParams = new URLSearchParams();
  if (params?.status) searchParams.set('status', params.status);
  if (params?.source) searchParams.set('source', params.source);
  if (params?.limit) searchParams.set('limit', String(params.limit));
  if (params?.offset) searchParams.set('offset', String(params.offset));

  const queryString = searchParams.toString();
  return apiClient.get<OpportunityListResponse>(
    `/factory/opportunities${queryString ? `?${queryString}` : ''}`
  );
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
  return apiClient.get<PendingReviewListResponse>(
    `/factory/reviews/pending${queryString ? `?${queryString}` : ''}`
  );
}

/**
 * Get opportunity details
 */
export async function getOpportunity(id: string): Promise<Opportunity> {
  return apiClient.get<Opportunity>(`/factory/opportunities/${id}`);
}

/**
 * Approve an opportunity
 */
export async function approveOpportunity(id: string): Promise<ReviewDecisionResponse> {
  return apiClient.post<ReviewDecisionResponse>(`/factory/opportunities/${id}/approve`);
}

/**
 * Reject an opportunity with a reason
 */
export async function rejectOpportunity(id: string, reason: string): Promise<ReviewDecisionResponse> {
  return apiClient.post<ReviewDecisionResponse>(`/factory/opportunities/${id}/reject`, {
    reason,
  } as RejectRequest);
}

/**
 * List factory functions (versions and published functions)
 */
export async function listFactoryFunctions(params?: {
  include_published?: boolean;
  limit?: number;
  offset?: number;
}): Promise<FunctionsListResponse> {
  const searchParams = new URLSearchParams();
  if (params?.include_published !== undefined) {
    searchParams.set('include_published', String(params.include_published));
  }
  if (params?.limit) searchParams.set('limit', String(params.limit));
  if (params?.offset) searchParams.set('offset', String(params.offset));

  const queryString = searchParams.toString();
  return apiClient.get<FunctionsListResponse>(
    `/factory/functions${queryString ? `?${queryString}` : ''}`
  );
}

// ============================================================================
// Export all types and API
// ============================================================================

export const factoryApi = {
  getStatus: getFactoryStatus,
  getConfig: getFactoryConfig,
  updateConfig: updateFactoryConfig,
  triggerPipelineRun,
  listOpportunities,
  listPendingReviews,
  getOpportunity,
  approveOpportunity,
  rejectOpportunity,
  listFunctions: listFactoryFunctions,
};

export default factoryApi;
