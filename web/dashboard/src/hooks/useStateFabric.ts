import { stateFabricApi } from '@/api/stateFabric';
import type {
  CreatePipelineRequest,
  CreateStateFabricRequest,
  CreateStoreRequest,
  CreateTriggerRequest,
  UpdatePipelineRequest,
  UpdateStateFabricRequest,
} from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

// Query keys
export const stateFabricKeys = {
  all: ['state-fabrics'] as const,
  lists: () => [...stateFabricKeys.all, 'list'] as const,
  list: (filters: string) => [...stateFabricKeys.lists(), { filters }] as const,
  details: () => [...stateFabricKeys.all, 'detail'] as const,
  detail: (id: string) => [...stateFabricKeys.details(), id] as const,
  metrics: (id: string, timeRange?: string) =>
    [...stateFabricKeys.detail(id), 'metrics', timeRange ?? 'default'] as const,
  stores: (fabricId: string) => [...stateFabricKeys.detail(fabricId), 'stores'] as const,
  pipelines: (fabricId: string) => [...stateFabricKeys.detail(fabricId), 'pipelines'] as const,
  events: (fabricId: string, filters?: Record<string, any>) =>
    [...stateFabricKeys.detail(fabricId), 'events', filters] as const,
  snapshots: (fabricId: string) => [...stateFabricKeys.detail(fabricId), 'snapshots'] as const,
  replays: (fabricId: string) => [...stateFabricKeys.detail(fabricId), 'replays'] as const,
  triggers: (fabricId?: string) => [...stateFabricKeys.all, 'triggers', fabricId] as const,
  featureFlags: () => [...stateFabricKeys.all, 'feature-flags'] as const,
};

// List all state fabrics
export function useStateFabrics() {
  return useQuery({
    queryKey: stateFabricKeys.lists(),
    queryFn: () => stateFabricApi.list(),
  });
}

// True when we should skip API calls (e.g. "new" is not a valid UUID)
function isFabricIdValidForFetch(id: string): boolean {
  return !!id && id !== 'new';
}

// Get a single state fabric
export function useStateFabric(id: string) {
  return useQuery({
    queryKey: stateFabricKeys.detail(id),
    queryFn: () => stateFabricApi.get(id),
    enabled: isFabricIdValidForFetch(id),
  });
}

// Create a state fabric
export function useCreateStateFabric() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateStateFabricRequest) => stateFabricApi.create(data),
    onSuccess: (created) => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.lists() });
      toast.success(`"${created.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to create state fabric: ${error.message}`);
    },
  });
}

// Update a state fabric
export function useUpdateStateFabric(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateStateFabricRequest) => stateFabricApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.lists() });
      toast.success('State fabric updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update state fabric: ${error.message}`);
    },
  });
}

// Delete a state fabric
export function useDeleteStateFabric() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => stateFabricApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.lists() });
      toast.success('State fabric deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete state fabric: ${error.message}`);
    },
  });
}

// Get state fabric metrics (optionally scoped by timeRange for charts)
export function useStateFabricMetrics(id: string, timeRange?: string) {
  return useQuery({
    queryKey: stateFabricKeys.metrics(id, timeRange),
    queryFn: () => stateFabricApi.getMetrics(id, timeRange),
    enabled: isFabricIdValidForFetch(id),
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

// List stores for a fabric
export function useStateFabricStores(fabricId: string) {
  return useQuery({
    queryKey: stateFabricKeys.stores(fabricId),
    queryFn: () => stateFabricApi.listStores(fabricId),
    enabled: isFabricIdValidForFetch(fabricId),
  });
}

// Create a store
export function useCreateStore(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateStoreRequest) => stateFabricApi.createStore(fabricId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.stores(fabricId) });
      toast.success('Store created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create store: ${error.message}`);
    },
  });
}

// Delete a store
export function useDeleteStore(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (storeId: string) => stateFabricApi.deleteStore(fabricId, storeId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.stores(fabricId) });
      toast.success('Store deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete store: ${error.message}`);
    },
  });
}

// List pipelines for a fabric
export function useStateFabricPipelines(fabricId: string) {
  return useQuery({
    queryKey: stateFabricKeys.pipelines(fabricId),
    queryFn: () => stateFabricApi.listPipelines(fabricId),
    enabled: isFabricIdValidForFetch(fabricId),
  });
}

// Create a pipeline
export function useCreatePipeline(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreatePipelineRequest) => stateFabricApi.createPipeline(fabricId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.pipelines(fabricId) });
      toast.success('Pipeline created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create pipeline: ${error.message}`);
    },
  });
}

// Update a pipeline
export function useUpdatePipeline(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ pipelineId, data }: { pipelineId: string; data: UpdatePipelineRequest }) =>
      stateFabricApi.updatePipeline(fabricId, pipelineId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.pipelines(fabricId) });
      toast.success('Pipeline updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update pipeline: ${error.message}`);
    },
  });
}

// Delete a pipeline
export function useDeletePipeline(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (pipelineId: string) => stateFabricApi.deletePipeline(fabricId, pipelineId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.pipelines(fabricId) });
      toast.success('Pipeline deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete pipeline: ${error.message}`);
    },
  });
}

// Execute a pipeline
export function useExecutePipeline(fabricId: string, pipelineId: string) {
  return useMutation({
    mutationFn: (input: Record<string, any>) =>
      stateFabricApi.executePipeline(fabricId, pipelineId, input),
    onSuccess: () => {
      toast.success('Pipeline executed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to execute pipeline: ${error.message}`);
    },
  });
}

// List event logs
export function useStateFabricEventLogs(
  fabricId: string,
  params?: {
    storeId?: string;
    eventType?: string;
    startTime?: string;
    endTime?: string;
    limit?: number;
    offset?: number;
  }
) {
  return useQuery({
    queryKey: stateFabricKeys.events(fabricId, params),
    queryFn: () => stateFabricApi.listEventLogs(fabricId, params),
    enabled: isFabricIdValidForFetch(fabricId),
  });
}

// List snapshots
export function useStateFabricSnapshots(fabricId: string, storeId?: string) {
  return useQuery({
    queryKey: stateFabricKeys.snapshots(fabricId),
    queryFn: () => stateFabricApi.listSnapshots(fabricId, storeId),
    enabled: isFabricIdValidForFetch(fabricId),
  });
}

// Create a snapshot
export function useCreateSnapshot(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; description?: string; storeId?: string }) =>
      stateFabricApi.createSnapshot(fabricId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.snapshots(fabricId) });
      toast.success('Snapshot created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create snapshot: ${error.message}`);
    },
  });
}

// Delete a snapshot
export function useDeleteSnapshot(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (snapshotId: string) => stateFabricApi.deleteSnapshot(fabricId, snapshotId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.snapshots(fabricId) });
      toast.success('Snapshot deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete snapshot: ${error.message}`);
    },
  });
}

// List replays
export function useStateFabricReplays(fabricId: string) {
  return useQuery({
    queryKey: stateFabricKeys.replays(fabricId),
    queryFn: () => stateFabricApi.listReplays(fabricId),
    enabled: isFabricIdValidForFetch(fabricId),
  });
}

// Create a replay session
export function useCreateReplay(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { snapshotId?: string; startEventId?: string; endEventId?: string }) =>
      stateFabricApi.createReplay(fabricId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.replays(fabricId) });
      toast.success('Replay session created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create replay: ${error.message}`);
    },
  });
}

// List triggers for a fabric
export function useStateFabricTriggers(fabricId?: string) {
  return useQuery({
    queryKey: stateFabricKeys.triggers(fabricId),
    queryFn: () => stateFabricApi.listTriggers(fabricId!),
    enabled: !!fabricId,
  });
}

// Create a trigger
export function useCreateTrigger(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTriggerRequest) => stateFabricApi.createTrigger(fabricId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.triggers() });
      toast.success('Trigger created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create trigger: ${error.message}`);
    },
  });
}

// Delete a trigger
export function useDeleteTrigger(fabricId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (triggerId: string) => stateFabricApi.deleteTrigger(fabricId, triggerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.triggers() });
      toast.success('Trigger deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete trigger: ${error.message}`);
    },
  });
}

// Fabric Keys - list keys in a store
export function useFabricKeys(fabricId: string, params?: { prefix?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: [...stateFabricKeys.detail(fabricId), 'keys', params] as const,
    queryFn: async () => {
      const stores = await stateFabricApi.listStores(fabricId);
      const store = stores[0];
      if (!store) return { keys: [], total: 0, statePath: '' };
      return {
        keys: [],
        total: 0,
        statePath: `${fabricId}/${store.id}`,
      };
    },
    enabled: isFabricIdValidForFetch(fabricId),
  });
}

// Get a single replay
export function useStateFabricReplay(fabricId: string, replayId: string) {
  return useQuery({
    queryKey: [...stateFabricKeys.replays(fabricId), replayId] as const,
    queryFn: () => stateFabricApi.getReplay(fabricId, replayId),
    enabled: isFabricIdValidForFetch(fabricId) && !!replayId,
  });
}

export interface ReplayProgressEvent {
  progress: number;
  eventsReplayed: number;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  error?: string;
  completed?: boolean;
}

interface UseReplayProgressStreamOptions {
  enabled?: boolean;
  onProgress?: (event: ReplayProgressEvent) => void;
  onError?: (error: Error) => void;
  onComplete?: (event: ReplayProgressEvent) => void;
  maxRetries?: number;
  retryDelay?: number;
}

export function useReplayProgressStream(
  fabricId: string,
  replayId: string,
  options: UseReplayProgressStreamOptions = {}
) {
  const {
    enabled = true,
    onProgress,
    onError,
    onComplete,
    maxRetries = 3,
    retryDelay = 2000,
  } = options;

  const queryClient = useQueryClient();
  const eventSourceRef = { current: null as EventSource | null };
  const retryCountRef = { current: 0 };
  const isConnectedRef = { current: false };

  const cleanup = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
      isConnectedRef.current = false;
    }
  };

  const connect = () => {
    if (!enabled || !fabricId || !replayId) return;

    const userToken = localStorage.getItem('auth_token');
    if (!userToken) {
      onError?.(new Error('Authentication required'));
      return;
    }

    cleanup();

    const url = `/v1/state-fabrics/${fabricId}/replays/${replayId}/progress?token=${encodeURIComponent(userToken)}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;
    isConnectedRef.current = true;

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as ReplayProgressEvent;
        retryCountRef.current = 0;

        if (data.error) {
          onError?.(new Error(data.error));
          return;
        }

        onProgress?.(data);

        if (data.completed || data.status === 'completed' || data.status === 'failed' || data.status === 'cancelled') {
          onComplete?.(data);
          cleanup();
        }
      } catch (err) {
        console.error('Failed to parse replay progress event:', err);
      }
    };

    eventSource.onerror = (err) => {
      console.error('Replay progress EventSource error:', err);
      isConnectedRef.current = false;

      if (eventSource.readyState === EventSource.CLOSED) {
        if (retryCountRef.current < maxRetries) {
          retryCountRef.current++;
          setTimeout(connect, retryDelay * retryCountRef.current);
        } else {
          onError?.(new Error('Connection lost and max retries exceeded'));
        }
      }
    };

    eventSource.onopen = () => {
      retryCountRef.current = 0;
      isConnectedRef.current = true;
    };
  };

  useMutation({
    mutationFn: async () => {
      connect();
      return { connected: true };
    },
  });

  return {
    connect,
    disconnect: cleanup,
    isConnected: isConnectedRef.current,
  };
}

export function useReplayProgress(fabricId: string, replayId: string) {
  const queryClient = useQueryClient();

  const { data: replay } = useStateFabricReplay(fabricId, replayId);
  const isActive = replay?.status === 'pending' || replay?.status === 'running';

  useReplayProgressStream(fabricId, replayId, {
    enabled: isActive,
    onProgress: (event) => {
      queryClient.setQueryData([...stateFabricKeys.replays(fabricId), replayId], (old: any) => {
        if (!old) return old;
        return {
          ...old,
          progress: event.progress,
          eventsReplayed: event.eventsReplayed,
          status: event.status,
        };
      });
    },
    onComplete: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.replays(fabricId) });
    },
  });

  return replay;
}

export function useStateFabricFeatureFlags() {
  return useQuery({
    queryKey: stateFabricKeys.featureFlags(),
    queryFn: () => stateFabricApi.getFeatureFlags(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useIsReplayStreamingEnabled() {
  const { data: flags } = useStateFabricFeatureFlags();
  return flags?.replay_progress_streaming ?? false;
}
