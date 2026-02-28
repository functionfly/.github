import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { stateFabricApi } from "@/api/stateFabric";
import type {
  StateFabric,
  StateFabricStore,
  Pipeline,
  EventLog,
  Snapshot,
  ReplaySession,
  CreateStateFabricRequest,
  UpdateStateFabricRequest,
  CreatePipelineRequest,
  UpdatePipelineRequest,
  CreateStoreRequest,
  StateFabricMetrics,
  StateFabricTrigger,
  CreateTriggerRequest,
} from "@/types";

// Query keys
export const stateFabricKeys = {
  all: ["state-fabrics"] as const,
  lists: () => [...stateFabricKeys.all, "list"] as const,
  list: (filters: string) => [...stateFabricKeys.lists(), { filters }] as const,
  details: () => [...stateFabricKeys.all, "detail"] as const,
  detail: (id: string) => [...stateFabricKeys.details(), id] as const,
  metrics: (id: string) => [...stateFabricKeys.detail(id), "metrics"] as const,
  stores: (fabricId: string) => [...stateFabricKeys.detail(fabricId), "stores"] as const,
  pipelines: (fabricId: string) => [...stateFabricKeys.detail(fabricId), "pipelines"] as const,
  events: (fabricId: string, filters?: Record<string, any>) =>
    [...stateFabricKeys.detail(fabricId), "events", filters] as const,
  snapshots: (fabricId: string) => [...stateFabricKeys.detail(fabricId), "snapshots"] as const,
  replays: (fabricId: string) => [...stateFabricKeys.detail(fabricId), "replays"] as const,
  triggers: (fabricId?: string) => [...stateFabricKeys.all, "triggers", fabricId] as const,
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
  return !!id && id !== "new";
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
      toast.success("State fabric updated successfully");
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
      toast.success("State fabric deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete state fabric: ${error.message}`);
    },
  });
}

// Get state fabric metrics
export function useStateFabricMetrics(id: string, timeRange?: string) {
  return useQuery({
    queryKey: stateFabricKeys.metrics(id),
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
      toast.success("Store created successfully");
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
      toast.success("Store deleted successfully");
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
      toast.success("Pipeline created successfully");
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
      toast.success("Pipeline updated successfully");
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
      toast.success("Pipeline deleted successfully");
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
      toast.success("Pipeline executed successfully");
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
      toast.success("Snapshot created successfully");
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
      toast.success("Snapshot deleted successfully");
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
      toast.success("Replay session created");
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
    queryFn: () => stateFabricApi.listTriggers({ state: fabricId }),
    enabled: !!fabricId,
  });
}

// Create a trigger
export function useCreateTrigger() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTriggerRequest) => stateFabricApi.createTrigger(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.triggers() });
      toast.success("Trigger created successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to create trigger: ${error.message}`);
    },
  });
}

// Delete a trigger
export function useDeleteTrigger() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (triggerId: string) => stateFabricApi.deleteTrigger(triggerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateFabricKeys.triggers() });
      toast.success("Trigger deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete trigger: ${error.message}`);
    },
  });
}
