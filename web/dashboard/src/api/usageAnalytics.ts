import { apiClient } from './client';

// ==================== Types ====================

export interface CostSummary {
  tenant_id: string;
  tenant_name?: string;
  period_start: string;
  period_end: string;
  total_executions: number;
  unique_functions: number;
  total_cost_cents: number;
  total_cost_usd: number;
  previous_period_cost_usd?: number;
  cost_trend_percent?: number;
  cost_breakdown: {
    execution: number;
    compute: number;
    platform_fee: number;
    data_transfer: number;
  };
  function_summaries: FunctionCostSummary[];
  daily_breakdown: DailyCostBreakdown[];
}

export interface FunctionCostSummary {
  function_id: string;
  function_name: string;
  function_author: string;
  total_executions: number;
  success_executions: number;
  error_executions: number;
  cached_executions: number;
  total_duration_ms: number;
  avg_duration_ms: number;
  total_cost_cents: number;
  total_cost_usd: number;
  avg_cost_cents: number;
  avg_cost_usd: number;
  cost_breakdown: {
    execution: number;
    compute: number;
    platform_fee: number;
    data_transfer: number;
  };
}

export interface DailyCostBreakdown {
  date: string;
  executions: number;
  cost_cents: number;
  cost_usd: number;
}

export interface RegionCostBreakdown {
  [region: string]: {
    total_executions: number;
    total_cost_cents: number;
    total_cost_usd: number;
    cost_breakdown: {
      execution: number;
      compute: number;
      platform_fee: number;
      data_transfer: number;
    };
  };
}

export interface CostAllocationEntry {
  id: string;
  function_id: string;
  function_name: string;
  function_author: string;
  execution_id: string;
  execution_outcome: 'success' | 'error';
  cached: boolean;
  duration_ms: number;
  cpu_time_ms: number;
  memory_used_mb: number;
  wall_time_ms: number;
  execution_cost_usd: number;
  compute_cost_usd: number;
  platform_fee_usd: number;
  data_transfer_usd: number;
  total_cost_usd: number;
  region: string;
  timestamp: string;
  tags: Record<string, string> | null;
  metadata: Record<string, unknown> | null;
}

export interface CostEntriesResponse {
  entries: CostAllocationEntry[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
  tenant_id: string;
}

export interface UsageForecast {
  tenant_id: string;
  period_start: string;
  period_end: string;
  current_usage: {
    executions: number;
    compute_ms: number;
    estimated_cost_usd: number;
  };
  forecast: {
    projected_executions: number;
    projected_compute_ms: number;
    projected_cost_usd: number;
    projected_overrun: boolean;
  };
  predicted_monthly_cost_usd: number;
  confidence: number;
  trend: 'increasing' | 'decreasing' | 'stable';
  recommendations: string[];
}

export interface UsageAlert {
  id: string;
  tenant_id: string;
  alert_type: 'spend_cap' | 'usage_threshold' | 'forecast_overrun';
  threshold_value: number;
  current_value: number;
  status: 'active' | 'triggered' | 'resolved';
  created_at: string;
  triggered_at?: string;
  resolved_at?: string;
}

export interface SpendCap {
  tenant_id: string;
  enabled: boolean;
  spend_cap_usd: number;
  cap_amount_usd: number;
  current_spend_usd: number;
  percentage_used: number;
  action: 'notify' | 'throttle' | 'block';
}

export interface UsageTrend {
  period: string;
  metric: string;
  current_value: number;
  previous_value: number;
  change_percent: number;
  trend_direction: 'up' | 'down' | 'flat';
}

// ==================== Helper Functions ====================

export function formatCostUsd(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100);
}

export function formatNumber(value: number): string {
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(1)}M`;
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(1)}K`;
  }
  return value.toLocaleString();
}

export function getErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (
      e as { response?: { data?: { error?: string; message?: string }; status?: number } }
    ).response;
    if (res?.data?.error) return res.data.error;
    if (res?.data?.message) return res.data.message;
    if (res?.status === 503) return 'Service temporarily unavailable. Please try again later.';
    if (res?.status === 429) return 'Rate limit exceeded. Please wait a moment.';
  }
  return 'Failed to load analytics data. Please try again.';
}

// ==================== API Functions ====================

/**
 * Get comprehensive cost summary for the tenant.
 * Returns total costs, execution counts, and breakdowns by function and day.
 */
export async function getCostSummary(startDate?: string, endDate?: string): Promise<CostSummary> {
  const params = new URLSearchParams();
  if (startDate) params.set('start', startDate);
  if (endDate) params.set('end', endDate);

  const queryString = params.toString();
  const url = queryString ? `/v1/costs/summary?${queryString}` : '/v1/costs/summary';

  return apiClient.get<CostSummary>(url);
}

/**
 * Get cost breakdown by function.
 * Returns aggregated costs and execution stats per function.
 */
export async function getCostByFunction(
  startDate?: string,
  endDate?: string
): Promise<{
  tenant_id: string;
  period_start: string;
  period_end: string;
  function_count: number;
  functions: FunctionCostSummary[];
}> {
  const params = new URLSearchParams();
  if (startDate) params.set('start', startDate);
  if (endDate) params.set('end', endDate);

  const queryString = params.toString();
  const url = queryString ? `/v1/costs/by-function?${queryString}` : '/v1/costs/by-function';

  return apiClient.get(url);
}

/**
 * Get cost breakdown by time period (daily).
 * Returns daily execution counts and costs.
 */
export async function getCostByPeriod(
  startDate?: string,
  endDate?: string,
  granularity: 'day' | 'hour' = 'day'
): Promise<{
  tenant_id: string;
  period_start: string;
  period_end: string;
  daily_breakdown: DailyCostBreakdown[];
}> {
  const params = new URLSearchParams();
  if (startDate) params.set('start', startDate);
  if (endDate) params.set('end', endDate);
  params.set('granularity', granularity);

  return apiClient.get(`/v1/costs/by-period?${params}`);
}

/**
 * Get cost breakdown by region.
 * Returns aggregated costs per region.
 */
export async function getCostByRegion(
  startDate?: string,
  endDate?: string
): Promise<{
  tenant_id: string;
  period_start: string;
  period_end: string;
  regions: RegionCostBreakdown;
}> {
  const params = new URLSearchParams();
  if (startDate) params.set('start', startDate);
  if (endDate) params.set('end', endDate);

  const queryString = params.toString();
  const url = queryString ? `/v1/costs/by-region?${queryString}` : '/v1/costs/by-region';

  return apiClient.get(url);
}

/**
 * Get detailed cost allocation entries.
 * Returns paginated list of individual execution cost records.
 */
export async function getCostEntries(
  options: {
    startDate?: string;
    endDate?: string;
    functionId?: string;
    outcome?: 'success' | 'error';
    cached?: boolean;
    region?: string;
    limit?: number;
    offset?: number;
  } = {}
): Promise<CostEntriesResponse> {
  const params = new URLSearchParams();

  if (options.startDate) params.set('start', options.startDate);
  if (options.endDate) params.set('end', options.endDate);
  if (options.functionId) params.set('function_id', options.functionId);
  if (options.outcome) params.set('outcome', options.outcome);
  if (options.cached !== undefined) params.set('cached', String(options.cached));
  if (options.region) params.set('region', options.region);
  if (options.limit) params.set('limit', String(options.limit));
  if (options.offset) params.set('offset', String(options.offset));

  return apiClient.get(`/v1/costs/entries?${params}`);
}

/**
 * Get usage forecast.
 * Returns projected usage and costs for the current period.
 */
export async function getUsageForecast(): Promise<UsageForecast> {
  return apiClient.get('/v1/usage/forecast');
}

/**
 * Get usage forecast by specific metric type.
 */
export async function getUsageForecastByType(
  type: 'executions' | 'compute' | 'cost'
): Promise<UsageForecast> {
  return apiClient.get(`/v1/usage/forecast/${type}`);
}

/**
 * Refresh the usage forecast.
 */
export async function refreshUsageForecast(): Promise<{ message: string }> {
  return apiClient.post('/v1/usage/forecast/refresh');
}

/**
 * List configured usage alerts.
 */
export async function listUsageAlerts(): Promise<{ alerts: UsageAlert[] }> {
  return apiClient.get('/v1/usage/alerts');
}

/**
 * Create a new usage alert.
 */
export async function createUsageAlert(alert: {
  alert_type: 'spend_cap' | 'usage_threshold' | 'forecast_overrun';
  threshold_value: number;
  action?: 'notify' | 'throttle' | 'block';
}): Promise<UsageAlert> {
  return apiClient.post('/v1/usage/alerts', alert);
}

/**
 * Update an existing usage alert.
 */
export async function updateUsageAlert(
  alertId: string,
  updates: Partial<Omit<UsageAlert, 'id' | 'tenant_id'>>
): Promise<UsageAlert> {
  return apiClient.put(`/v1/usage/alerts/${alertId}`, updates);
}

/**
 * Delete a usage alert.
 */
export async function deleteUsageAlert(alertId: string): Promise<void> {
  await apiClient.delete(`/v1/usage/alerts/${alertId}`);
}

/**
 * Get spend cap configuration.
 */
export async function getSpendCap(): Promise<SpendCap> {
  return apiClient.get('/v1/usage/spend-cap');
}

/**
 * Update spend cap configuration.
 */
export async function updateSpendCap(config: {
  enabled: boolean;
  cap_amount_usd: number;
  action: 'notify' | 'throttle' | 'block';
}): Promise<SpendCap> {
  return apiClient.put('/v1/usage/spend-cap', config);
}

/**
 * Get usage trends over time.
 */
export async function getUsageTrends(days: number = 30): Promise<{ trends: UsageTrend[] }> {
  return apiClient.get(`/v1/usage/trends?days=${days}`);
}

/**
 * Get alert history.
 */
export async function getAlertHistory(
  limit: number = 20,
  offset: number = 0
): Promise<{
  alerts: Array<UsageAlert & { triggered_at: string; resolved_at?: string }>;
  total: number;
  limit: number;
  offset: number;
}> {
  return apiClient.get(`/v1/usage/alerts/history?limit=${limit}&offset=${offset}`);
}
