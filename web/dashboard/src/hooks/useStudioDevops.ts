import {
  studioDevOpsApi,
  type CloudRegion,
  type Environment,
  type Pipeline,
  type PipelineStage
} from '@/api/studioDevops';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

const DEVOPS_STATS_KEY = 'studio-devops-stats';
const PIPELINES_KEY = 'studio-devops-pipelines';
const ENVIRONMENTS_KEY = 'studio-devops-environments';
const REGIONS_KEY = 'studio-devops-regions';

export function useStudioDevOps() {
  const queryClient = useQueryClient();

  // Stats query
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: [DEVOPS_STATS_KEY],
    queryFn: async () => {
      const res = await studioDevOpsApi.getStats();
      return res.stats;
    },
    staleTime: 1000 * 30, // 30 seconds
  });

  // Pipelines query
  const {
    data: pipelines = [],
    isLoading: pipelinesLoading,
    refetch: refetchPipelines,
  } = useQuery({
    queryKey: [PIPELINES_KEY],
    queryFn: async () => {
      const res = await studioDevOpsApi.listPipelines({ limit: 50 });
      return res.pipelines;
    },
    staleTime: 1000 * 60, // 1 minute
  });

  // Environments query
  const {
    data: environments = [],
    isLoading: environmentsLoading,
    refetch: refetchEnvironments,
  } = useQuery({
    queryKey: [ENVIRONMENTS_KEY],
    queryFn: async () => {
      const res = await studioDevOpsApi.listEnvironments();
      return res.environments;
    },
    staleTime: 1000 * 60,
  });

  // Regions query
  const {
    data: regions = [],
    isLoading: regionsLoading,
    refetch: refetchRegions,
  } = useQuery({
    queryKey: [REGIONS_KEY],
    queryFn: async () => {
      const res = await studioDevOpsApi.listRegions();
      return res.regions;
    },
    staleTime: 1000 * 60,
  });

  // Create pipeline mutation
  const createPipelineMutation = useMutation({
    mutationFn: (data: {
      name: string;
      version?: string;
      branch?: string;
      commit_sha?: string;
      source?: string;
    }) => studioDevOpsApi.createPipeline(data),
    onSuccess: (data) => {
      queryClient.setQueryData([PIPELINES_KEY], (old: Pipeline[] = []) => [
        data.pipeline,
        ...old,
      ]);
      queryClient.invalidateQueries({ queryKey: [DEVOPS_STATS_KEY] });
    },
  });

  // Update pipeline stage mutation
  const updateStageMutation = useMutation({
    mutationFn: ({
      pipelineId,
      stageId,
      updates,
    }: {
      pipelineId: string;
      stageId: string;
      updates: Partial<PipelineStage>;
    }) => studioDevOpsApi.updatePipelineStage(pipelineId, stageId, updates),
    onSuccess: (data) => {
      queryClient.setQueryData([PIPELINES_KEY], (old: Pipeline[] = []) =>
        old.map((p) => (p.id === data.pipeline.id ? data.pipeline : p))
      );
    },
  });

  // Retry pipeline stage mutation
  const retryStageMutation = useMutation({
    mutationFn: ({ pipelineId, stageId }: { pipelineId: string; stageId: string }) =>
      studioDevOpsApi.retryPipelineStage(pipelineId, stageId),
    onSuccess: (data) => {
      queryClient.setQueryData([PIPELINES_KEY], (old: Pipeline[] = []) =>
        old.map((p) => (p.id === data.pipeline.id ? data.pipeline : p))
      );
    },
  });

  // Create environment mutation
  const createEnvironmentMutation = useMutation({
    mutationFn: (data: {
      name: string;
      type: 'production' | 'staging' | 'preview' | 'development';
      color?: string;
      variables?: Record<string, string>;
      replicas?: number;
      auto_scale?: boolean;
      region?: string;
    }) => studioDevOpsApi.createEnvironment(data),
    onSuccess: (data) => {
      queryClient.setQueryData([ENVIRONMENTS_KEY], (old: Environment[] = []) => [
        data.environment,
        ...old,
      ]);
      queryClient.invalidateQueries({ queryKey: [DEVOPS_STATS_KEY] });
    },
  });

  // Update environment mutation
  const updateEnvironmentMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: Partial<Environment> }) =>
      studioDevOpsApi.updateEnvironment(id, updates),
    onSuccess: (data) => {
      queryClient.setQueryData([ENVIRONMENTS_KEY], (old: Environment[] = []) =>
        old.map((e) => (e.id === data.environment.id ? data.environment : e))
      );
    },
  });

  // Delete environment mutation
  const deleteEnvironmentMutation = useMutation({
    mutationFn: (id: string) => studioDevOpsApi.deleteEnvironment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [ENVIRONMENTS_KEY] });
      queryClient.invalidateQueries({ queryKey: [DEVOPS_STATS_KEY] });
    },
  });

  // Add environment variable mutation
  const addVariableMutation = useMutation({
    mutationFn: ({ envId, key, value }: { envId: string; key: string; value: string }) =>
      studioDevOpsApi.addEnvironmentVariable(envId, key, value),
    onSuccess: (data) => {
      queryClient.setQueryData([ENVIRONMENTS_KEY], (old: Environment[] = []) =>
        old.map((e) => (e.id === data.environment.id ? data.environment : e))
      );
    },
  });

  // Add environment secret mutation
  const addSecretMutation = useMutation({
    mutationFn: ({ envId, key }: { envId: string; key: string }) =>
      studioDevOpsApi.addEnvironmentSecret(envId, key),
    onSuccess: (data) => {
      queryClient.setQueryData([ENVIRONMENTS_KEY], (old: Environment[] = []) =>
        old.map((e) => (e.id === data.environment.id ? data.environment : e))
      );
    },
  });

  // Create region mutation
  const createRegionMutation = useMutation({
    mutationFn: studioDevOpsApi.createRegion,
    onSuccess: (data) => {
      queryClient.setQueryData([REGIONS_KEY], (old: CloudRegion[] = []) => [
        data.region,
        ...old,
      ]);
      queryClient.invalidateQueries({ queryKey: [DEVOPS_STATS_KEY] });
    },
  });

  // Action callbacks
  const createPipeline = useCallback(
    async (data: {
      name: string;
      version?: string;
      branch?: string;
      commit_sha?: string;
      source?: string;
    }) => {
      await createPipelineMutation.mutateAsync(data);
    },
    [createPipelineMutation]
  );

  const updateStage = useCallback(
    async (pipelineId: string, stageId: string, updates: Partial<PipelineStage>) => {
      await updateStageMutation.mutateAsync({ pipelineId, stageId, updates });
    },
    [updateStageMutation]
  );

  const retryStage = useCallback(
    async (pipelineId: string, stageId: string) => {
      await retryStageMutation.mutateAsync({ pipelineId, stageId });
    },
    [retryStageMutation]
  );

  const createEnvironment = useCallback(
    async (data: {
      name: string;
      type: 'production' | 'staging' | 'preview' | 'development';
      color?: string;
      variables?: Record<string, string>;
      replicas?: number;
      auto_scale?: boolean;
      region?: string;
    }) => {
      await createEnvironmentMutation.mutateAsync(data);
    },
    [createEnvironmentMutation]
  );

  const updateEnvironment = useCallback(
    async (id: string, updates: Partial<Environment>) => {
      await updateEnvironmentMutation.mutateAsync({ id, updates });
    },
    [updateEnvironmentMutation]
  );

  const deleteEnvironment = useCallback(
    async (id: string) => {
      await deleteEnvironmentMutation.mutateAsync(id);
    },
    [deleteEnvironmentMutation]
  );

  const addVariable = useCallback(
    async (envId: string, key: string, value: string) => {
      await addVariableMutation.mutateAsync({ envId, key, value });
    },
    [addVariableMutation]
  );

  const addSecret = useCallback(
    async (envId: string, key: string) => {
      await addSecretMutation.mutateAsync({ envId, key });
    },
    [addSecretMutation]
  );

  const createRegion = useCallback(
    async (data: {
      name: string;
      provider: 'aws' | 'gcp' | 'azure' | 'custom';
      zone?: string;
      zone_name?: string;
      location?: string;
      country?: string;
      specs?: { compute?: number; memory?: number; storage?: number; gpu?: boolean };
    }) => {
      await createRegionMutation.mutateAsync(data);
    },
    [createRegionMutation]
  );

  const refreshAll = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: [DEVOPS_STATS_KEY] });
    queryClient.invalidateQueries({ queryKey: [PIPELINES_KEY] });
    queryClient.invalidateQueries({ queryKey: [ENVIRONMENTS_KEY] });
    queryClient.invalidateQueries({ queryKey: [REGIONS_KEY] });
  }, [queryClient]);

  return {
    // Data
    stats,
    pipelines,
    environments,
    regions,

    // Loading states
    isLoadingStats: statsLoading,
    isLoadingPipelines: pipelinesLoading,
    isLoadingEnvironments: environmentsLoading,
    isLoadingRegions: regionsLoading,
    isLoading:
      statsLoading || pipelinesLoading || environmentsLoading || regionsLoading,

    // Actions
    createPipeline,
    updateStage,
    retryStage,
    createEnvironment,
    updateEnvironment,
    deleteEnvironment,
    addVariable,
    addSecret,
    createRegion,
    refreshAll,

    // Mutation states
    isCreatingPipeline: createPipelineMutation.isPending,
    isUpdatingStage: updateStageMutation.isPending,
    isRetryingStage: retryStageMutation.isPending,
    isCreatingEnvironment: createEnvironmentMutation.isPending,
    isUpdatingEnvironment: updateEnvironmentMutation.isPending,
    isDeletingEnvironment: deleteEnvironmentMutation.isPending,
    isAddingVariable: addVariableMutation.isPending,
    isAddingSecret: addSecretMutation.isPending,
    isCreatingRegion: createRegionMutation.isPending,
  };
}