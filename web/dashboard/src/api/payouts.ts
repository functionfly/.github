import { apiClient } from './client';

// ==================== Types ====================

export interface ConnectAccountStatus {
  has_account: boolean;
  account_id?: string;
  status: string;
  payouts_enabled: boolean;
  bank_name?: string;
  bank_last4?: string;
  country?: string;
  onboarding_url?: string;
  needs_onboarding: boolean;
}

export interface OnboardingResult {
  account_id: string;
  onboarding_url: string;
  status: string;
}

export interface PayoutBalance {
  user_id: string;
  available_balance_usd: number;
  pending_balance_usd: number;
  total_earnings_usd: number;
  total_paid_out_usd: number;
}

export interface PayoutRequest {
  id: string;
  amount_cents: number;
  currency: string;
  status: string;
  stripe_transfer_id?: string;
  stripe_payout_id?: string;
  idempotency_key: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;
}

export interface PayoutRequestResult {
  payout_request_id: string;
  amount_cents: number;
  currency: string;
  status: string;
}

export interface PayoutLedgerEntry {
  id: string;
  user_id: string;
  entry_type: string;
  amount_cents: number;
  currency: string;
  reference_type?: string;
  reference_id?: string;
  balance_after_cents: number;
  description?: string;
  created_at: string;
}

export interface PayoutRequestsResponse {
  requests: PayoutRequest[];
  total: number;
  limit: number;
  offset: number;
}

export interface PayoutLedgerResponse {
  entries: PayoutLedgerEntry[];
  total: number;
  limit: number;
  offset: number;
}

// ==================== API Functions ====================

/**
 * Get the current user's Stripe Connect account status.
 */
export async function getConnectAccountStatus(): Promise<ConnectAccountStatus> {
  return apiClient.get<ConnectAccountStatus>('/v1/payouts/connect-account');
}

/**
 * Start Stripe Connect onboarding. Returns an onboarding URL.
 */
export async function startConnectOnboarding(): Promise<OnboardingResult> {
  return apiClient.post<OnboardingResult>('/v1/payouts/connect-account', {});
}

/**
 * Refresh the connect account status from Stripe.
 */
export async function refreshConnectAccount(): Promise<ConnectAccountStatus> {
  return apiClient.post<ConnectAccountStatus>('/v1/payouts/connect-account/refresh', {});
}

/**
 * Get the user's payout balance breakdown.
 */
export async function getPayoutBalance(): Promise<PayoutBalance> {
  return apiClient.get<PayoutBalance>('/v1/payouts/balance');
}

/**
 * Request a payout.
 * @param amountCents Amount in USD cents
 * @param idempotencyKey Unique key to prevent duplicate requests
 */
export async function requestPayout(
  amountCents: number,
  idempotencyKey: string
): Promise<PayoutRequestResult> {
  return apiClient.post<PayoutRequestResult>('/v1/payouts/request', {
    amount_cents: amountCents,
    idempotency_key: idempotencyKey,
  });
}

/**
 * List payout requests.
 */
export async function listPayoutRequests(
  limit = 20,
  offset = 0
): Promise<PayoutRequestsResponse> {
  return apiClient.get<PayoutRequestsResponse>('/v1/payouts/requests', {
    params: { limit, offset },
  });
}

/**
 * List payout ledger entries.
 */
export async function listPayoutLedger(
  limit = 50,
  offset = 0
): Promise<PayoutLedgerResponse> {
  return apiClient.get<PayoutLedgerResponse>('/v1/payouts/ledger', {
    params: { limit, offset },
  });
}
