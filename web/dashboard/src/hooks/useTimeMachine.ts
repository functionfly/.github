import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import * as timeMachineApi from '@/api/timeMachine';
import type { CreateReplayRequest } from '@/api/timeMachine';

const ACTIVE_STATUSES = new Set([
  'pending',
  'scanning',
  'replaying',
  'diffing',
  'reconciling',
]);

export const timeMachineKeys = {
  all: ['time-machine'] as const,
  replays: () => [...timeMachineKeys.all, 'replays'] as const,
  replayList: (params?: Record<string, unknown>) =>
    [...timeMachineKeys.replays(), params] as const,
  replay: (id: string) => [...timeMachineKeys.replays(), id] as const,
  items: (id: string, params?: Record<string, unknown>) =>
    [...timeMachineKeys.replay(id), 'items', params] as const,
  diffSummary: (id: string) => [...timeMachineKeys.replay(id), 'diff-summary'] as const,
  reconciliations: (id: string) =>
    [...timeMachineKeys.replay(id), 'reconciliations'] as const,
  auditCertificate: (id: string) =>
    [...timeMachineKeys.replay(id), 'audit-certificate'] as const,
  limits: () => [...timeMachineKeys.all, 'limits'] as const,
};

export function useReplays(params?: {
  function_id?: string;
  status?: string;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: timeMachineKeys.replayList(params as Record<string, unknown>),
    queryFn: () => timeMachineApi.listReplays(params),
    staleTime: 1000 * 30,
  });
}

export function useReplay(id: string) {
  return useQuery({
    queryKey: timeMachineKeys.replay(id),
    queryFn: () => timeMachineApi.getReplay(id),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status && ACTIVE_STATUSES.has(status) ? 5000 : false;
    },
  });
}

export function useCreateReplay() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: CreateReplayRequest) => timeMachineApi.createReplay(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: timeMachineKeys.replays() });
      toast.success('Replay job created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create replay: ${error.message}`);
    },
  });
}

export function useCancelReplay() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => timeMachineApi.cancelReplay(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: timeMachineKeys.replay(id) });
      queryClient.invalidateQueries({ queryKey: timeMachineKeys.replays() });
      toast.success('Replay cancelled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to cancel replay: ${error.message}`);
    },
  });
}

export function useReplayItems(
  replayId: string,
  params?: { limit?: number; offset?: number; diff_type?: string }
) {
  return useQuery({
    queryKey: timeMachineKeys.items(replayId, params as Record<string, unknown>),
    queryFn: () => timeMachineApi.getReplayItems(replayId, params),
    enabled: !!replayId,
    staleTime: 1000 * 30,
  });
}

export function useReplayItem(replayId: string, itemId: string) {
  return useQuery({
    queryKey: [...timeMachineKeys.replay(replayId), 'items', itemId],
    queryFn: () => timeMachineApi.getReplayItem(replayId, itemId),
    enabled: !!replayId && !!itemId,
  });
}

export function useDiffSummary(replayId: string) {
  return useQuery({
    queryKey: timeMachineKeys.diffSummary(replayId),
    queryFn: () => timeMachineApi.getDiffSummary(replayId),
    enabled: !!replayId,
  });
}

export function useStartReconciliation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, mode }: { id: string; mode: 'dry_run' | 'live' }) =>
      timeMachineApi.startReconciliation(id, mode),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: timeMachineKeys.replay(id) });
      queryClient.invalidateQueries({ queryKey: timeMachineKeys.reconciliations(id) });
      toast.success('Reconciliation started');
    },
    onError: (error: Error) => {
      toast.error(`Failed to start reconciliation: ${error.message}`);
    },
  });
}

export function useReconciliations(
  replayId: string,
  params?: { limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: timeMachineKeys.reconciliations(replayId),
    queryFn: () => timeMachineApi.listReconciliations(replayId, params),
    enabled: !!replayId,
    staleTime: 1000 * 30,
  });
}

export function useAuditCertificate(replayId: string) {
  return useQuery({
    queryKey: timeMachineKeys.auditCertificate(replayId),
    queryFn: () => timeMachineApi.getAuditCertificate(replayId),
    enabled: !!replayId,
  });
}

export function useTimeMachineLimits() {
  return useQuery({
    queryKey: timeMachineKeys.limits(),
    queryFn: () => timeMachineApi.getTimeMachineLimits(),
    staleTime: 1000 * 60 * 5,
  });
}
