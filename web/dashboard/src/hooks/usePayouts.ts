import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  getConnectAccountStatus,
  startConnectOnboarding,
  refreshConnectAccount,
  getPayoutBalance,
  requestPayout,
  cancelPayout,
  listPayoutRequests,
  listPayoutLedger,
  getPayoutSchedule,
  updatePayoutSchedule,
  type PayoutRequestWithFee,
  type PayoutSchedulePreference,
} from '@/api/payouts';

export const payoutKeys = {
  all: ['payouts'] as const,
  connectStatus: () => [...payoutKeys.all, 'connect-status'] as const,
  balance: () => [...payoutKeys.all, 'balance'] as const,
  requests: (params?: { limit?: number; offset?: number }) =>
    [...payoutKeys.all, 'requests', params] as const,
  ledger: (params?: { limit?: number; offset?: number }) =>
    [...payoutKeys.all, 'ledger', params] as const,
  schedule: () => [...payoutKeys.all, 'schedule'] as const,
};

export function useConnectAccountStatus() {
  return useQuery({
    queryKey: payoutKeys.connectStatus(),
    queryFn: getConnectAccountStatus,
    staleTime: 1000 * 60,
    retry: false,
  });
}

export function useStartConnectOnboarding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: startConnectOnboarding,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: payoutKeys.connectStatus() });
    },
    onError: (error: Error) => {
      toast.error(`Failed to start onboarding: ${error.message}`);
    },
  });
}

export function useRefreshConnectAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: refreshConnectAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: payoutKeys.connectStatus() });
      toast.success('Account status refreshed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to refresh account: ${error.message}`);
    },
  });
}

export function usePayoutBalance() {
  return useQuery({
    queryKey: payoutKeys.balance(),
    queryFn: getPayoutBalance,
    staleTime: 1000 * 60,
    retry: false,
  });
}

export function useRequestPayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      amountCents,
      idempotencyKey,
      feeType,
    }: {
      amountCents: number;
      idempotencyKey: string;
      feeType?: string;
    }) => requestPayout(amountCents, idempotencyKey, feeType),
    onSuccess: (data: PayoutRequestWithFee) => {
      const netUsd = (data.fee.net_amount_cents / 100).toFixed(2);
      const feeUsd = (data.fee.fee_amount_cents / 100).toFixed(2);
      const desc =
        data.fee.fee_amount_cents > 0
          ? `$${netUsd} after $${feeUsd} fee — ${data.payout.status}`
          : `$${netUsd} — ${data.payout.status}`;
      toast.success('Payout requested', { description: desc });
      queryClient.invalidateQueries({ queryKey: payoutKeys.balance() });
      queryClient.invalidateQueries({ queryKey: payoutKeys.requests() });
      queryClient.invalidateQueries({ queryKey: payoutKeys.ledger() });
    },
    onError: (error: Error) => {
      toast.error(`Payout failed: ${error.message}`);
    },
  });
}

export function useCancelPayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ payoutId, reason }: { payoutId: string; reason?: string }) =>
      cancelPayout(payoutId, reason),
    onSuccess: () => {
      toast.success('Payout cancelled — funds returned to balance');
      queryClient.invalidateQueries({ queryKey: payoutKeys.balance() });
      queryClient.invalidateQueries({ queryKey: payoutKeys.requests() });
      queryClient.invalidateQueries({ queryKey: payoutKeys.ledger() });
    },
    onError: (error: Error) => {
      toast.error(`Cancel failed: ${error.message}`);
    },
  });
}

export function usePayoutRequests(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: payoutKeys.requests(params),
    queryFn: () => listPayoutRequests(params?.limit, params?.offset),
    staleTime: 1000 * 60,
  });
}

export function usePayoutLedger(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: payoutKeys.ledger(params),
    queryFn: () => listPayoutLedger(params?.limit, params?.offset),
    staleTime: 1000 * 60,
  });
}

export function usePayoutSchedule() {
  return useQuery({
    queryKey: payoutKeys.schedule(),
    queryFn: getPayoutSchedule,
    staleTime: 1000 * 60,
  });
}

export function useUpdatePayoutSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (pref: Omit<PayoutSchedulePreference, 'last_auto_payout_at' | 'next_scheduled_at'>) =>
      updatePayoutSchedule(pref),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: payoutKeys.schedule() });
      toast.success('Payout schedule updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update schedule: ${error.message}`);
    },
  });
}
