import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  factoryApi,
  type FactoryStatus,
  type FactoryConfig,
  type PendingReviewListResponse,
  type OpportunityListResponse,
  type FunctionsListResponse,
} from '@/api/factory';

// Query keys
export const factoryKeys = {
  all: ['factory'] as const,
  status: () => [...factoryKeys.all, 'status'] as const,
  config: () => [...factoryKeys.all, 'config'] as const,
  opportunities: (params?: { status?: string; source?: string; limit?: number; offset?: number }) =>
    [...factoryKeys.all, 'opportunities', params] as const,
  pendingReviews: (params?: { source?: string; limit?: number; offset?: number }) =>
    [...factoryKeys.all, 'pendingReviews', params] as const,
  functions: (params?: { include_published?: boolean; limit?: number; offset?: number }) =>
    [...factoryKeys.all, 'functions', params] as const,
};

// Get factory status
export function useFactoryStatus(options?: { refetchInterval?: number }) {
  return useQuery({
    queryKey: factoryKeys.status(),
    queryFn: async () => {
      const data = await factoryApi.getStatus();
      return data ?? null;
    },
    refetchInterval: options?.refetchInterval ?? 30000, // Default 30 seconds
    staleTime: 1000 * 15, // 15 seconds
  });
}

// Get factory configuration
export function useFactoryConfig() {
  return useQuery({
    queryKey: factoryKeys.config(),
    queryFn: () => factoryApi.getConfig(),
    staleTime: 1000 * 60, // 1 minute
  });
}

// Update factory configuration
export function useUpdateFactoryConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (config: Partial<FactoryConfig>) => factoryApi.updateConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: factoryKeys.config() });
      queryClient.invalidateQueries({ queryKey: factoryKeys.status() });
      toast.success('Factory configuration updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update configuration: ${error.message}`);
    },
  });
}

// Trigger pipeline run
export function useTriggerPipelineRun() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => factoryApi.triggerPipelineRun(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: factoryKeys.status() });
      toast.success('Pipeline run initiated', {
        description: data.run?.id ? `Run ID: ${data.run.id}` : 'Pipeline started successfully',
      });
    },
    onError: (error: Error) => {
      toast.error('Failed to start pipeline', {
        description: error.message || 'An error occurred',
      });
    },
  });
}

// List opportunities
export function useFactoryOpportunities(
  params?: { status?: string; source?: string; limit?: number; offset?: number }
) {
  return useQuery({
    queryKey: factoryKeys.opportunities(params),
    queryFn: async () => {
      const data = await factoryApi.listOpportunities(params);
      return data ?? { opportunities: [], total: 0, limit: 20, offset: 0 };
    },
    staleTime: 1000 * 30, // 30 seconds
  });
}

// List pending reviews
export function usePendingReviews(
  params?: { source?: string; limit?: number; offset?: number },
  options?: { refetchInterval?: number }
) {
  return useQuery({
    queryKey: factoryKeys.pendingReviews(params),
    queryFn: async () => {
      const data = await factoryApi.listPendingReviews(params);
      return data ?? { reviews: [], total: 0, limit: 20, offset: 0 };
    },
    refetchInterval: options?.refetchInterval ?? 15000, // Default 15 seconds
    staleTime: 1000 * 10, // 10 seconds
  });
}

// List factory functions
export function useFactoryFunctions(params?: { include_published?: boolean; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: factoryKeys.functions(params),
    queryFn: async () => {
      const data = await factoryApi.listFunctions(params);
      return data ?? { versions: [], total_versions: 0, limit: 20, offset: 0 };
    },
    staleTime: 1000 * 60,
  });
}

// Approve opportunity
export function useApproveOpportunity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => factoryApi.approveOpportunity(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: factoryKeys.pendingReviews() });
      queryClient.invalidateQueries({ queryKey: factoryKeys.status() });
      queryClient.invalidateQueries({ queryKey: factoryKeys.opportunities() });
      toast.success('Opportunity approved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to approve opportunity: ${error.message}`);
    },
  });
}

// Reject opportunity
export function useRejectOpportunity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) =>
      factoryApi.rejectOpportunity(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: factoryKeys.pendingReviews() });
      queryClient.invalidateQueries({ queryKey: factoryKeys.status() });
      queryClient.invalidateQueries({ queryKey: factoryKeys.opportunities() });
      toast.success('Opportunity rejected');
    },
    onError: (error: Error) => {
      toast.error(`Failed to reject opportunity: ${error.message}`);
    },
  });
}
