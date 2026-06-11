import { apiClient } from './client';

// ==================== Types ====================

export interface TrustAPITierPricing {
  tier: string;
  monthly_price_usd: number;
  included_requests: number;
  overage_price_per_1000: number;
  has_overage_billing: boolean;
  rate_limit_per_minute: number;
  rate_limit_per_day: number;
  monthly_request_limit: number;
  description: string;
}

export interface TrustAPITierPricingResponse {
  tiers: TrustAPITierPricing[];
}

export interface TrustAPIBillingStatus {
  partner_id: string;
  tier: string;
  billing_status: string;
  is_founder_mode: boolean;
  monthly_price_usd: number;
  included_requests: number;
  current_usage: number;
  remaining_requests: number;
  overage_requests: number;
  overage_charge_usd: number;
  billing_period_start: string;
  billing_period_end: string;
  is_hard_limit: boolean;
  founder_mode_started_at?: string;
  founder_mode_ends_at?: string;
  usage_threshold?: number;
}

export interface TrustAPICheckoutRequest {
  tier: string;
  success_url?: string;
  cancel_url?: string;
}

export interface TrustAPICheckoutResponse {
  session_id: string;
  url: string;
  status: string;
}

export interface TrustAPIUsageRecord {
  date: string;
  requests: number;
  cost_usd: number;
}

export interface TrustAPIUsageReport {
  partner_id: string;
  tier: string;
  period_start: string;
  period_end: string;
  total_requests: number;
  included_requests: number;
  overage_requests: number;
  overage_cost_usd: number;
  total_cost_usd: number;
  records: TrustAPIUsageRecord[];
}

export interface TrustAPIInvoice {
  id: string;
  partner_id: string;
  stripe_invoice_id: string | null;
  amount: number;
  currency: string;
  status: string;
  invoice_date: string | null;
  due_date: string | null;
  invoice_pdf: string | null;
  hosted_invoice_url: string | null;
  created_at: string;
}

export interface TrustAPIInvoicesResponse {
  invoices: TrustAPIInvoice[];
  limit: number;
  offset: number;
  total: number;
}

export interface TrustAPIFounderModeRequest {
  mode_type: 'time_based' | 'revenue_based' | 'hybrid';
  free_days: number;
  mrr_threshold: number;
  success_url: string;
  cancel_url: string;
}

export interface TrustAPIFounderModeStatus {
  id: string;
  partner_id: string;
  mode_type: string;
  status: string;
  started_at: string;
  ends_at?: string;
  free_days: number;
  mrr_threshold_cents: number;
  days_remaining: number;
}

// ==================== Helper Functions ====================

export function getTrustAPIErrorMessage(e: unknown, defaultMsg = 'An error occurred'): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string; message?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.data?.message) return res.data.message;
    if (res?.status === 401) return 'Unauthorized. Please log in again.';
    if (res?.status === 403) return 'Access denied.';
    if (res?.status === 404) return 'Resource not found.';
    if (res?.status === 503) return 'Billing is not configured. Contact support.';
  }
  return defaultMsg;
}

// ==================== API Functions ====================

/**
 * Get all Trust API tier pricing.
 * GET /v1/partners/tiers
 */
export async function getTrustAPITierPricing(): Promise<TrustAPITierPricingResponse> {
  return apiClient.get<TrustAPITierPricingResponse>('/v1/partners/tiers');
}

/**
 * Get current partner's billing status.
 * GET /v1/partners/{partner_id}/billing
 */
export async function getTrustAPIBillingStatus(): Promise<TrustAPIBillingStatus> {
  return apiClient.get<TrustAPIBillingStatus>('/v1/partners/me/billing');
}

/**
 * Create a Stripe checkout session for tier upgrade.
 * POST /v1/partners/{partner_id}/billing/checkout
 */
export async function createTrustAPICheckout(
  tier: string,
  successUrl?: string,
  cancelUrl?: string
): Promise<TrustAPICheckoutResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  const body: TrustAPICheckoutRequest = {
    tier,
  };
  if (successUrl) body.success_url = successUrl;
  if (cancelUrl) body.cancel_url = cancelUrl;

  return apiClient.post<TrustAPICheckoutResponse>(
    '/v1/partners/me/billing/checkout',
    body,
    { headers }
  );
}

/**
 * Get partner's usage report.
 * GET /v1/partners/{partner_id}/billing/usage
 */
export async function getTrustAPIUsageReport(): Promise<TrustAPIUsageReport> {
  return apiClient.get<TrustAPIUsageReport>('/v1/partners/me/billing/usage');
}

/**
 * Get partner's invoices.
 * GET /v1/partners/{partner_id}/billing/invoices
 */
export async function getTrustAPIInvoices(
  limit = 10,
  offset = 0
): Promise<TrustAPIInvoicesResponse> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
  });
  return apiClient.get<TrustAPIInvoicesResponse>(`/v1/partners/me/billing/invoices?${params}`);
}

/**
 * Enroll in founder mode.
 * POST /v1/partners/{partner_id}/founder
 */
export async function enrollTrustAPIFounderMode(
  data: TrustAPIFounderModeRequest
): Promise<TrustAPIFounderModeStatus> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<TrustAPIFounderModeStatus>(
    '/v1/partners/me/founder',
    data,
    { headers }
  );
}

// ==================== Tier Display Helpers ====================

export const TRUST_API_TIER_LABELS: Record<string, string> = {
  developer: 'Developer',
  payg: 'Pay-as-you-Go',
  startup: 'Startup',
  business: 'Business',
  enterprise: 'Enterprise',
};

export const TRUST_API_TIER_COLORS: Record<string, string> = {
  developer: '#6B7280',
  payg: '#3B82F6',
  startup: '#10B981',
  business: '#F59E0B',
  enterprise: '#8B5CF6',
};

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatNumber(num: number): string {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(1)}M`;
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(0)}K`;
  }
  return num.toString();
}

export function getUsagePercentage(current: number, included: number): number {
  if (included === 0) return 0;
  return Math.min((current / included) * 100, 100);
}