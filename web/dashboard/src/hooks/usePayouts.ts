import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  getConnectAccountStatus,
  startConnectOnboarding,
  refreshConnectAccount,
  getPayoutBalance,
  requestPayout,
  listPayoutRequests,
  listPayoutLedger,
  type PayoutRequestResult,
} from '@/api/payouts';

// Query keys
export const payoutKeys = {
  all: ['payouts'] as const,
  connectStatus: () => [...payoutKeys.all, 'connect-status'] as const,
  balance: () => [...payoutKeys.all, 'balance'] as const,
  requests: (params?: { limit?: number; offset?: number }) =>
    [...payoutKeys.all, 'requests', params] as const,
  ledger: (params?: { limit?: number; offset?: number }) =>
    [...payoutKeys.all, 'ledger', params] as const,
};

// Get connect account status
export function useConnectAccountStatus() {
  return useQuery({
    queryKey: payoutKeys.connectStatus(),
    queryFn: getConnectAccountStatus,
    staleTime: 1000 * 60,
  });
}

// Start connect onboarding
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

// Refresh connect account
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

// Get payout balance
export function usePayoutBalance() {
  return useQuery({
    queryKey: payoutKeys.balance(),
    queryFn: getPayoutBalance,
    staleTime: 1000 * 60,
  });
}

// Request payout
export function useRequestPayout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ amountCents, idempotencyKey }: { amountCents: number; idempotencyKey: string }) =>
      requestPayout(amountCents, idempotencyKey),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: payoutKeys.balance() });
      queryClient.invalidateQueries({ queryKey: payoutKeys.requests() });
      toast.success('Payout requested successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to request payout: ${error.message}`);
    },
  });
}

// List payout requests
export function usePayoutRequests(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: payoutKeys.requests(params),
    queryFn: () => listPayoutRequests(params?.limit, params?.offset),
    staleTime: 1000 * 60,
  });
}

// List payout ledger
export function usePayoutLedger(params?: { limit?: number; offset?: number }) {
  return useQuery({
    queryKey: payoutKeys.ledger(params),
    queryFn: () => listPayoutLedger(params?.limit, params?.offset),
    staleTime: 1000 * 60,
  });
}
