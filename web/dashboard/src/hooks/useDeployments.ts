import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { deploymentsApi, type Deployment, type DeployRequest } from '@/api/deployments';

// Query keys
export const deploymentKeys = {
  all: ['deployments'] as const,
  lists: (appId: string) => [...deploymentKeys.all, 'list', appId] as const,
  detail: (id: string) => [...deploymentKeys.all, 'detail', id] as const,
};

// List deployments for an app
export function useDeployments(appId: string, limit?: number) {
  return useQuery({
    queryKey: deploymentKeys.lists(appId),
    queryFn: () => deploymentsApi.list(appId, limit),
    enabled: !!appId,
    staleTime: 1000 * 60,
  });
}

// Get deployment
export function useDeployment(deploymentId: string) {
  return useQuery({
    queryKey: deploymentKeys.detail(deploymentId),
    queryFn: () => deploymentsApi.get(deploymentId),
    enabled: !!deploymentId,
    staleTime: 1000 * 60,
  });
}

// Deploy function
export function useDeploy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: DeployRequest) => deploymentsApi.deploy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: deploymentKeys.all });
      toast.success('Deployment started successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to deploy: ${error.message}`);
    },
  });
}

// Rollback deployment
export function useRollbackDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (deploymentId: string) => deploymentsApi.rollback(deploymentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: deploymentKeys.all });
      toast.success('Rollback initiated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to rollback: ${error.message}`);
    },
  });
}
