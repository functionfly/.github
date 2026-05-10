import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { dnaApi } from '@/api/dna';

function handleDNAError(error: Error, action: string): void {
  const message = error?.message || '';
  if (message.includes('402') || message.includes('insufficient')) {
    toast.error(`Insufficient credits to ${action}. Please top up your wallet.`);
  } else if (message.includes('409') || message.includes('already actioned')) {
    toast.error('This mutation has already been actioned.');
  } else if (message.includes('404') || message.includes('expired')) {
    toast.error('This mutation has expired or is no longer available.');
  } else {
    toast.error(`Failed to ${action}: ${message}`);
  }
}

// ──────────────────────────────────────────────────────────────────────────────
// Query Keys
// ──────────────────────────────────────────────────────────────────────────────

export const dnaKeys = {
  all: ['dna'] as const,
  profile: (functionId: string) => [...dnaKeys.all, 'profile', functionId] as const,
  mutations: (functionId: string, params?: { status?: string; limit?: number; offset?: number }) =>
    [...dnaKeys.all, 'mutations', functionId, params] as const,
  mutation: (functionId: string, mutationId: string) =>
    [...dnaKeys.all, 'mutation', functionId, mutationId] as const,
  insights: (functionId: string, period?: string) =>
    [...dnaKeys.all, 'insights', functionId, period] as const,
  enterpriseInsights: (period?: string) =>
    [...dnaKeys.all, 'enterprise-insights', period] as const,
};

// ──────────────────────────────────────────────────────────────────────────────
// Queries
// ──────────────────────────────────────────────────────────────────────────────

export function useDNAProfile(functionId: string, type?: string) {
  return useQuery({
    queryKey: dnaKeys.profile(functionId),
    queryFn: () => dnaApi.getProfile(functionId, type),
    enabled: !!functionId,
    staleTime: 1000 * 60 * 2,
  });
}

export function useDNAMutations(
  functionId: string,
  params?: { status?: string; limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: dnaKeys.mutations(functionId, params),
    queryFn: () => dnaApi.listMutations(functionId, params),
    enabled: !!functionId,
    staleTime: 1000 * 30,
  });
}

export function useDNAMutation(functionId: string, mutationId: string) {
  return useQuery({
    queryKey: dnaKeys.mutation(functionId, mutationId),
    queryFn: () => dnaApi.getMutation(functionId, mutationId),
    enabled: !!functionId && !!mutationId,
    staleTime: 1000 * 60 * 5,
  });
}

export function useDNAInsights(functionId: string, period?: string) {
  return useQuery({
    queryKey: dnaKeys.insights(functionId, period),
    queryFn: () => dnaApi.getInsights(functionId, period),
    enabled: !!functionId,
    staleTime: 1000 * 60 * 5,
  });
}

export function useEnterpriseDNAInsights(period?: string) {
  return useQuery({
    queryKey: dnaKeys.enterpriseInsights(period),
    queryFn: () => dnaApi.getEnterpriseInsights(period),
    staleTime: 1000 * 60 * 10,
  });
}

// ──────────────────────────────────────────────────────────────────────────────
// Mutations
// ──────────────────────────────────────────────────────────────────────────────

export function useAcceptDNAVariant(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      mutationId,
      canaryPercentage = 10,
    }: {
      mutationId: string;
      canaryPercentage?: number;
    }) => dnaApi.acceptVariant(functionId, mutationId, canaryPercentage),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: dnaKeys.profile(functionId) });
      queryClient.invalidateQueries({ queryKey: dnaKeys.mutations(functionId) });
      toast.success('Variant accepted — deploying via canary');
    },
    onError: (error: Error) => {
      handleDNAError(error, 'accept variant');
    },
  });
}

export function useRejectDNAVariant(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ mutationId, reason }: { mutationId: string; reason?: string }) =>
      dnaApi.rejectVariant(functionId, mutationId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: dnaKeys.mutations(functionId) });
      toast.success('Variant rejected');
    },
    onError: (error: Error) => {
      handleDNAError(error, 'reject variant');
    },
  });
}

export function useRollbackDNAVariant(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ mutationId, reason }: { mutationId: string; reason?: string }) =>
      dnaApi.rollbackVariant(functionId, mutationId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: dnaKeys.profile(functionId) });
      queryClient.invalidateQueries({ queryKey: dnaKeys.mutations(functionId) });
      toast.success('Variant rolled back — canary cancelled');
    },
    onError: (error: Error) => {
      handleDNAError(error, 'rollback variant');
    },
  });
}

export function useTriggerDNAAnalysis(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => dnaApi.triggerAnalysis(functionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: dnaKeys.profile(functionId) });
      toast.success('DNA analysis queued');
    },
    onError: (error: Error) => {
      handleDNAError(error, 'trigger analysis');
    },
  });
}

export function useToggleDNAEvolution(functionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (enabled: boolean) => dnaApi.toggleEvolution(functionId, enabled),
    onSuccess: (_, enabled) => {
      queryClient.invalidateQueries({ queryKey: dnaKeys.profile(functionId) });
      toast.success(enabled ? 'Evolution enabled' : 'Evolution paused');
    },
    onError: (error: Error) => {
      handleDNAError(error, 'toggle evolution');
    },
  });
}
