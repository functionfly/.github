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

export interface PayoutFeeInfo {
  gross_amount_cents: number;
  fee_amount_cents: number;
  net_amount_cents: number;
  fee_type: string;
  fee_rate: number;
}

export interface PayoutRequestWithFee {
  payout: PayoutRequestResult;
  fee: PayoutFeeInfo;
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

export interface PayoutSchedulePreference {
  schedule_enabled: boolean;
  frequency: 'weekly' | 'biweekly' | 'monthly';
  minimum_amount_cents: number;
  day_of_week?: number;
  day_of_month?: number;
  last_auto_payout_at?: string;
  next_scheduled_at?: string;
}

// ==================== API Functions ====================

export async function getConnectAccountStatus(): Promise<ConnectAccountStatus> {
  return apiClient.get<ConnectAccountStatus>('/v1/payouts/connect-account');
}

export async function startConnectOnboarding(): Promise<OnboardingResult> {
  return apiClient.post<OnboardingResult>('/v1/payouts/connect-account', {});
}

export async function refreshConnectAccount(): Promise<ConnectAccountStatus> {
  return apiClient.post<ConnectAccountStatus>('/v1/payouts/connect-account/refresh', {});
}

export async function getPayoutBalance(): Promise<PayoutBalance> {
  return apiClient.get<PayoutBalance>('/v1/payouts/balance');
}

export async function requestPayout(
  amountCents: number,
  idempotencyKey: string,
  feeType = 'standard'
): Promise<PayoutRequestWithFee> {
  return apiClient.post<PayoutRequestWithFee>('/v1/payouts/request', {
    amount_cents: amountCents,
    idempotency_key: idempotencyKey,
    fee_type: feeType,
  });
}

export async function cancelPayout(
  payoutId: string,
  reason?: string
): Promise<{ success: boolean; payout_id: string; cancelled_at: string }> {
  return apiClient.post(`/v1/payouts/${payoutId}/cancel`, { reason: reason ?? '' });
}

export async function listPayoutRequests(
  limit = 20,
  offset = 0
): Promise<PayoutRequestsResponse> {
  return apiClient.get<PayoutRequestsResponse>('/v1/payouts/requests', {
    params: { limit, offset },
  });
}

export async function listPayoutLedger(
  limit = 50,
  offset = 0
): Promise<PayoutLedgerResponse> {
  return apiClient.get<PayoutLedgerResponse>('/v1/payouts/ledger', {
    params: { limit, offset },
  });
}

export async function getPayoutSchedule(): Promise<PayoutSchedulePreference> {
  return apiClient.get<PayoutSchedulePreference>('/v1/payouts/schedule');
}

export async function updatePayoutSchedule(
  pref: Omit<PayoutSchedulePreference, 'last_auto_payout_at' | 'next_scheduled_at'>
): Promise<{ success: boolean; preference: PayoutSchedulePreference }> {
  return apiClient.put('/v1/payouts/schedule', pref);
}
