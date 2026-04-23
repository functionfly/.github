import { environmentService, type Environment } from '@/api/environment';
import { useSidebarStore } from '@/stores/sidebarStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect } from 'react';
import { toast } from 'sonner';

// Re-export Environment type for convenience
export type { Environment };

// Query keys for active environment
export const activeEnvironmentKeys = {
  all: ['active-environment'] as const,
  current: () => [...activeEnvironmentKeys.all, 'current'] as const,
};

/**
 * Hook to fetch and sync the active environment with the backend.
 * This ensures the sidebar environment is persisted across sessions
 * and available for API scoping via the X-Environment header.
 */
export function useActiveEnvironment() {
  const queryClient = useQueryClient();
  const { currentEnvironment, setEnvironment } = useSidebarStore();

  // Query to fetch the active environment from backend
  const { data, isLoading, error } = useQuery({
    queryKey: activeEnvironmentKeys.current(),
    queryFn: () => environmentService.getActiveEnvironment(),
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 30, // 30 minutes
    retry: 2,
  });

  // Mutation to update the active environment
  const mutation = useMutation({
    mutationFn: (env: Environment) => environmentService.setActiveEnvironment(env),
    onSuccess: (response) => {
      // Update local store
      setEnvironment(response.environment);
      // Invalidate queries that depend on environment
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      queryClient.invalidateQueries({ queryKey: ['deployments'] });
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      toast.success(`Switched to ${response.environment} environment`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to switch environment: ${error.message}`);
    },
  });

  // Sync backend environment to local store on initial load
  useEffect(() => {
    if (data?.environment && data.environment !== currentEnvironment) {
      // Backend has a different environment than local store
      // Update local store to match backend (backend is source of truth)
      setEnvironment(data.environment);
    }
  }, [data?.environment, currentEnvironment, setEnvironment]);

  // Wrapper to set environment with backend sync
  const setActiveEnvironmentWithSync = useCallback(
    (environment: Environment) => {
      // Optimistically update local store
      setEnvironment(environment);
      // Sync to backend
      mutation.mutate(environment);
    },
    [setEnvironment, mutation]
  );

  return {
    // Current environment (from local store, synced with backend)
    environment: currentEnvironment,
    // Available environments
    availableEnvironments: data?.available || ['production', 'staging', 'development'],
    // Loading state
    isLoading: isLoading || mutation.isPending,
    // Error state
    error: error || mutation.error,
    // Set environment (with backend sync)
    setEnvironment: setActiveEnvironmentWithSync,
    // Raw mutation for advanced use cases
    mutation,
  };
}

/**
 * Hook to check if an environment selector should be shown.
 * Returns false if the user doesn't have access to multiple environments.
 */
export function useEnvironmentSelectorVisibility() {
  // For now, all users can see the environment selector
  // In the future, this could check tenant plan features
  return {
    showEnvironmentSelector: true,
  };
}

/**
 * Hook to get the environment-aware query key.
 * This ensures queries are automatically invalidated when environment changes.
 */
export function useEnvironmentQueryKey(baseKey: string[]) {
  const { currentEnvironment } = useSidebarStore();
  return [...baseKey, 'env', currentEnvironment];
}
