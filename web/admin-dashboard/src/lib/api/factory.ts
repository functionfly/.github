/**
 * Factory API Client
 * Handles AI Function Factory API communication
 */

import { adminApiClient } from './adminClient';
import { API_ROUTES, factoryRoute } from '@/lib/constants';

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
  metadata?: Record<string, unknown>;
  tags?: string[];
  category?: string;
  confidence_score?: number;
  estimated_value?: number;
  source_url?: string;
  retry_count?: number;
  last_error?: string;
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

export interface FactoryVersion {
  id: string;
  run_id: string;
  opportunity_id: string;
  function_id: string;
  generated_code: string;
  manifest: string;
  model_used: string;
  quality_score: number;
  test_score: number;
  review_required: boolean;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface PublishedFunction {
  id: string;
  author: string;
  name: string;
  version: string;
  runtime: string;
  created_at: string;
  status: string;
}

// Re-export for convenience
export type { PublishedFunction as FactoryPublishedFunction };

export interface FactoryFunctionsResponse {
  versions: FactoryVersion[];
  total_versions: number;
  published_functions?: PublishedFunction[];
  total_published?: number;
  limit: number;
  offset: number;
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
  const response = await adminApiClient.get<FactoryStatus>(API_ROUTES.FACTORY_STATUS);
  return unwrap(response);
}

/**
 * Get factory configuration
 */
export async function getFactoryConfig(): Promise<FactoryConfig> {
  const response = await adminApiClient.get<FactoryConfig>(API_ROUTES.FACTORY_CONFIG);
  return unwrap(response);
}

/**
 * Update factory configuration
 */
export async function updateFactoryConfig(config: Partial<FactoryConfig>): Promise<FactoryConfig> {
  const response = await adminApiClient.patch<FactoryConfig>(API_ROUTES.FACTORY_CONFIG, config);
  return unwrap(response);
}

/**
 * Trigger a manual pipeline run
 */
export async function triggerPipelineRun(): Promise<PipelineRunResponse> {
  const response = await adminApiClient.post<PipelineRunResponse>(API_ROUTES.FACTORY_PIPELINE_RUN);
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
    `${API_ROUTES.FACTORY_REVIEWS_PENDING}${queryString ? `?${queryString}` : ''}`
  );
  return unwrap(response);
}

/**
 * Approve an opportunity
 */
export async function approveOpportunity(id: string): Promise<ReviewDecisionResponse> {
  const response = await adminApiClient.post<ReviewDecisionResponse>(factoryRoute(id, 'approve'));
  return unwrap(response);
}

/**
 * Reject an opportunity with a reason
 */
export async function rejectOpportunity(id: string, reason: string): Promise<ReviewDecisionResponse> {
  const response = await adminApiClient.post<ReviewDecisionResponse>(factoryRoute(id, 'reject'), {
    reason,
  });
  return unwrap(response);
}

/**
 * Get a single opportunity by ID
 */
export async function getOpportunity(id: string): Promise<Opportunity> {
  const response = await adminApiClient.get<Opportunity>(factoryRoute(id));
  return unwrap(response);
}

/**
 * List all opportunities with filtering
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
  const response = await adminApiClient.get<OpportunityListResponse>(
    `${factoryRoute()}${queryString ? `?${queryString}` : ''}`
  );
  return unwrap(response);
}

/**
 * Update opportunity metadata or tags
 */
export async function updateOpportunity(id: string, updates: Partial<Opportunity>): Promise<Opportunity> {
  const response = await adminApiClient.patch<Opportunity>(factoryRoute(id), updates);
  return unwrap(response);
}

/**
 * List factory generated functions
 */
export async function getFactoryFunctions(params?: {
  limit?: number;
  offset?: number;
  include_published?: boolean;
}): Promise<FactoryFunctionsResponse> {
  const searchParams = new URLSearchParams();
  if (params?.limit) searchParams.set('limit', String(params.limit));
  if (params?.offset) searchParams.set('offset', String(params.offset));
  if (params?.include_published !== undefined) searchParams.set('include_published', String(params.include_published));

  const queryString = searchParams.toString();
  const response = await adminApiClient.get<FactoryFunctionsResponse>(
    `${API_ROUTES.FACTORY_FUNCTIONS}${queryString ? `?${queryString}` : ''}`
  );
  return unwrap(response);
}

export interface FactoryVersionCode {
  id: string;
  run_id: string;
  opportunity_id: string;
  function_id: string;
  generated_code: string;
  manifest: string;
  model_used: string;
  quality_score: number;
  test_score: number;
  review_required: boolean;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface FactoryVersionCodeResponse {
  version: FactoryVersionCode;
}

/**
 * Get factory version code (includes generated_code and manifest)
 */
export async function getFactoryVersionCode(versionId: string): Promise<FactoryVersionCode> {
  const response = await adminApiClient.get<FactoryVersionCodeResponse>(
    `${API_ROUTES.FACTORY_VERSION_CODE}/${versionId}/code`
  );
  const version = (response as { data?: { version: FactoryVersionCode } }).data?.version
    ?? (response as any).version
    ?? (response as any).data?.version
    ?? (response as any);
  return version as FactoryVersionCode;
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
  getOpportunity,
  listOpportunities,
  updateOpportunity,
  getFactoryFunctions,
  getFactoryVersionCode,
};

export default factoryApi;
