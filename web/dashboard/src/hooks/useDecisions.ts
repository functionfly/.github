import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { decisionsApi, type TeamDecision, type DecisionStatus, type CreateDecisionRequest } from '@/api/decisions';

// Query keys
export const decisionKeys = {
  all: ['decisions'] as const,
  lists: (teamId: string, params?: { status?: DecisionStatus; tag?: string }) =>
    [...decisionKeys.all, 'list', teamId, params] as const,
  detail: (teamId: string, id: string) => [...decisionKeys.all, 'detail', teamId, id] as const,
  search: (teamId: string, query: string) => [...decisionKeys.all, 'search', teamId, query] as const,
};

// List decisions
export function useDecisions(
  teamId: string,
  params?: { status?: DecisionStatus; tag?: string; limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: decisionKeys.lists(teamId, params),
    queryFn: () => decisionsApi.list(teamId, params),
    enabled: !!teamId,
    staleTime: 1000 * 60,
  });
}

// Get single decision
export function useDecision(teamId: string, decisionId: string) {
  return useQuery({
    queryKey: decisionKeys.detail(teamId, decisionId),
    queryFn: () => decisionsApi.get(teamId, decisionId),
    enabled: !!teamId && !!decisionId,
    staleTime: 1000 * 60,
  });
}

// Create decision
export function useCreateDecision() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, data }: { teamId: string; data: CreateDecisionRequest }) =>
      decisionsApi.create(teamId, data),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: decisionKeys.lists(teamId) });
      toast.success('Decision recorded successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create decision: ${error.message}`);
    },
  });
}

// Update decision
export function useUpdateDecision() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      teamId,
      decisionId,
      data,
    }: {
      teamId: string;
      decisionId: string;
      data: Partial<TeamDecision>;
    }) => decisionsApi.update(teamId, decisionId, data),
    onSuccess: (_, { teamId, decisionId }) => {
      queryClient.invalidateQueries({ queryKey: decisionKeys.detail(teamId, decisionId) });
      queryClient.invalidateQueries({ queryKey: decisionKeys.lists(teamId) });
      toast.success('Decision updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update decision: ${error.message}`);
    },
  });
}

// Delete decision
export function useDeleteDecision() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ teamId, decisionId }: { teamId: string; decisionId: string }) =>
      decisionsApi.delete(teamId, decisionId),
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: decisionKeys.lists(teamId) });
      toast.success('Decision deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete decision: ${error.message}`);
    },
  });
}

// Approve/supersede/deprecate decision
export function useApproveDecision() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      teamId,
      decisionId,
      status,
    }: {
      teamId: string;
      decisionId: string;
      status: 'approved' | 'superseded' | 'deprecated';
    }) => decisionsApi.approve(teamId, decisionId, { status }),
    onSuccess: (_, { teamId, decisionId }) => {
      queryClient.invalidateQueries({ queryKey: decisionKeys.detail(teamId, decisionId) });
      queryClient.invalidateQueries({ queryKey: decisionKeys.lists(teamId) });
      toast.success('Decision status updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update decision status: ${error.message}`);
    },
  });
}

// Search decisions
export function useSearchDecisions(teamId: string, query: string, limit?: number) {
  return useQuery({
    queryKey: decisionKeys.search(teamId, query),
    queryFn: () => decisionsApi.search(teamId, query, limit),
    enabled: !!teamId && !!query,
    staleTime: 1000 * 30,
  });
}
