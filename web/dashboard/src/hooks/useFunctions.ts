import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { functionsApi, type FunctionConfig, type CreateFunctionRequest, type UpdateFunctionRequest, type DeployFunctionRequest } from '@/api/functions';

// Query keys
export const functionKeys = {
  all: ['functions'] as const,
  lists: () => [...functionKeys.all, 'list'] as const,
  detail: (id: string) => [...functionKeys.all, 'detail', id] as const,
  deployments: (id: string) => [...functionKeys.all, 'deployments', id] as const,
  logs: (id: string, params?: { limit?: number; since?: string; level?: string }) =>
    [...functionKeys.all, 'logs', id, params] as const,
  metrics: (id: string, params?: { period?: string; from?: string; to?: string }) =>
    [...functionKeys.all, 'metrics', id, params] as const,
  trustScore: (id: string) => [...functionKeys.all, 'trust-score', id] as const,
  deploymentLogs: (deploymentId: string) => [...functionKeys.all, 'deployment-logs', deploymentId] as const,
};

// List functions
export function useFunctions() {
  return useQuery({
    queryKey: functionKeys.lists(),
    queryFn: () => functionsApi.list(),
    staleTime: 1000 * 60,
  });
}

// Get function
export function useFunction(functionId: string) {
  return useQuery({
    queryKey: functionKeys.detail(functionId),
    queryFn: () => functionsApi.get(functionId),
    enabled: !!functionId,
    staleTime: 1000 * 60,
  });
}

// Create function
export function useCreateFunction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateFunctionRequest) => functionsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: functionKeys.lists() });
      toast.success('Function created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create function: ${error.message}`);
    },
  });
}

// Update function
export function useUpdateFunction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ functionId, data }: { functionId: string; data: UpdateFunctionRequest }) =>
      functionsApi.update(functionId, data),
    onSuccess: (_, { functionId }) => {
      queryClient.invalidateQueries({ queryKey: functionKeys.detail(functionId) });
      queryClient.invalidateQueries({ queryKey: functionKeys.lists() });
      toast.success('Function updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update function: ${error.message}`);
    },
  });
}

// Delete function
export function useDeleteFunction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (functionId: string) => functionsApi.delete(functionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: functionKeys.lists() });
      toast.success('Function deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete function: ${error.message}`);
    },
  });
}

// Deploy function
export function useDeployFunction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: DeployFunctionRequest) => functionsApi.deploy(data),
    onSuccess: (_, data) => {
      queryClient.invalidateQueries({ queryKey: functionKeys.deployments(data.functionId) });
      toast.success('Function deployed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to deploy function: ${error.message}`);
    },
  });
}

// Get function deployments
export function useFunctionDeployments(functionId: string) {
  return useQuery({
    queryKey: functionKeys.deployments(functionId),
    queryFn: () => functionsApi.getDeployments(functionId),
    enabled: !!functionId,
    staleTime: 1000 * 60,
  });
}

// Get function logs
export function useFunctionLogs(
  functionId: string,
  params?: { limit?: number; since?: string; level?: string }
) {
  return useQuery({
    queryKey: functionKeys.logs(functionId, params),
    queryFn: () => functionsApi.getLogs(functionId, params),
    enabled: !!functionId,
    staleTime: 1000 * 30,
    refetchInterval: 30000, // Refresh logs every 30s
  });
}

// Get function metrics
export function useFunctionMetrics(
  functionId: string,
  params?: { period?: string; from?: string; to?: string }
) {
  return useQuery({
    queryKey: functionKeys.metrics(functionId, params),
    queryFn: () => functionsApi.getMetrics(functionId, params),
    enabled: !!functionId,
    staleTime: 1000 * 60,
  });
}

// Get function trust score
export function useFunctionTrustScore(functionId: string) {
  return useQuery({
    queryKey: functionKeys.trustScore(functionId),
    queryFn: () => functionsApi.getTrustScore(functionId),
    enabled: !!functionId,
    staleTime: 1000 * 60 * 5,
  });
}

// Get deployment logs
export function useDeploymentLogs(deploymentId: string) {
  return useQuery({
    queryKey: functionKeys.deploymentLogs(deploymentId),
    queryFn: () => functionsApi.getDeploymentLogs(deploymentId),
    enabled: !!deploymentId,
    staleTime: 1000 * 30,
  });
}

// Test function
export function useTestFunction() {
  return useMutation({
    mutationFn: ({ functionId, input }: { functionId: string; input: unknown }) =>
      functionsApi.test({ functionId, input }),
    onError: (error: Error) => {
      toast.error(`Test failed: ${error.message}`);
    },
  });
}
