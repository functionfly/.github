import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { stateApi } from "@/api/state";
import type {
  SimpleState,
  StateValue,
  StateHistoryEntry,
  StateSnapshot,
  StatePermission,
  CreateStateRequest,
  UpdateStateRequest,
  PatchStateRequest,
  CreateSnapshotRequest,
  GrantPermissionRequest,
  TimeTravelRequest,
  TimeTravelResponse,
} from "@/types";

// Query keys
export const stateKeys = {
  all: ["state"] as const,
  lists: () => [...stateKeys.all, "list"] as const,
  list: (filters: string) => [...stateKeys.lists(), { filters }] as const,
  details: () => [...stateKeys.all, "detail"] as const,
  detail: (path: string) => [...stateKeys.details(), path] as const,
  value: (path: string) => [...stateKeys.detail(path), "value"] as const,
  history: (path: string) => [...stateKeys.detail(path), "history"] as const,
  snapshots: (path: string) => [...stateKeys.detail(path), "snapshots"] as const,
  permissions: (path: string) => [...stateKeys.detail(path), "permissions"] as const,
};

// Helper to validate path for API calls
function isPathValidForFetch(path: string): boolean {
  return !!path && path !== "new";
}

// List all states
export function useStates(params?: { prefix?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: stateKeys.list(params?.prefix || ""),
    queryFn: () => stateApi.list(params),
  });
}

// Get a single state by path
export function useState(path: string) {
  return useQuery({
    queryKey: stateKeys.detail(path),
    queryFn: () => stateApi.get(path),
    enabled: isPathValidForFetch(path),
  });
}

// Create a state
export function useCreateState() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateStateRequest) => stateApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.lists() });
      toast.success("State created successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to create state: ${error.message}`);
    },
  });
}

// Update a state
export function useUpdateState(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateStateRequest) => stateApi.update(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.detail(path) });
      queryClient.invalidateQueries({ queryKey: stateKeys.lists() });
      toast.success("State updated successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to update state: ${error.message}`);
    },
  });
}

// Delete a state
export function useDeleteState() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (path: string) => stateApi.delete(path),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.lists() });
      toast.success("State deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete state: ${error.message}`);
    },
  });
}

// Get state value
export function useStateValue(path: string) {
  return useQuery({
    queryKey: stateKeys.value(path),
    queryFn: () => stateApi.getValue(path),
    enabled: isPathValidForFetch(path),
  });
}

// Set state value
export function useSetStateValue(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { value: StateValue; ttl?: number }) => stateApi.setValue(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.value(path) });
      queryClient.invalidateQueries({ queryKey: stateKeys.detail(path) });
      toast.success("Value set successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to set value: ${error.message}`);
    },
  });
}

// Patch state value
export function usePatchStateValue(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { operations: PatchStateRequest["operations"] }) =>
      stateApi.patchValue(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.value(path) });
      queryClient.invalidateQueries({ queryKey: stateKeys.detail(path) });
      toast.success("Value patched successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to patch value: ${error.message}`);
    },
  });
}

// Delete state value
export function useDeleteStateValue(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => stateApi.deleteValue(path),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.value(path) });
      queryClient.invalidateQueries({ queryKey: stateKeys.detail(path) });
      toast.success("Value deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete value: ${error.message}`);
    },
  });
}

// Get state history
export function useStateHistory(
  path: string,
  params?: {
    limit?: number;
    offset?: number;
    startTime?: string;
    endTime?: string;
  }
) {
  return useQuery({
    queryKey: stateKeys.history(path),
    queryFn: () => stateApi.getHistory(path, params),
    enabled: isPathValidForFetch(path),
  });
}

// Create snapshot
export function useCreateSnapshot(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateSnapshotRequest) => stateApi.createSnapshot(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.snapshots(path) });
      toast.success("Snapshot created successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to create snapshot: ${error.message}`);
    },
  });
}

// List snapshots
export function useStateSnapshots(path: string) {
  return useQuery({
    queryKey: stateKeys.snapshots(path),
    queryFn: () => stateApi.listSnapshots(path),
    enabled: isPathValidForFetch(path),
  });
}

// Restore snapshot
export function useRestoreSnapshot(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { snapshotId: string }) => stateApi.restoreSnapshot(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.detail(path) });
      queryClient.invalidateQueries({ queryKey: stateKeys.value(path) });
      toast.success("Snapshot restored successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to restore snapshot: ${error.message}`);
    },
  });
}

// Time travel
export function useTimeTravel() {
  return useMutation({
    mutationFn: (data: TimeTravelRequest) => stateApi.timeTravel(data),
    onError: (error: Error) => {
      toast.error(`Time travel failed: ${error.message}`);
    },
  });
}

// Get permissions
export function useStatePermissions(path: string) {
  return useQuery({
    queryKey: stateKeys.permissions(path),
    queryFn: () => stateApi.getPermissions(path),
    enabled: isPathValidForFetch(path),
  });
}

// Grant permission
export function useGrantPermission(path: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: GrantPermissionRequest) => stateApi.grantPermission(path, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: stateKeys.permissions(path) });
      toast.success("Permission granted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to grant permission: ${error.message}`);
    },
  });
}

// Enable encryption for a state path
export function useEnableEncryption(path: string) {
  return useMutation({
    mutationFn: () => stateApi.enableEncryption(path),
    onSuccess: () => {
      toast.success("Encryption enabled");
    },
    onError: (error: Error) => {
      toast.error(`Failed to enable encryption: ${error.message}`);
    },
  });
}

// Get encryption statistics (tenant-level)
export function useEncryptionStats(_path?: string) {
  return useQuery({
    queryKey: ['state', 'encryption', 'stats'] as const,
    queryFn: () => stateApi.getEncryptionStats(),
  });
}

// Migrate encryption
export function useMigrateEncryption(path: string) {
  return useMutation({
    mutationFn: (data: { batch_size?: number; dry_run?: boolean; force_rotate?: boolean }) =>
      stateApi.migrateEncryption(path, data),
    onSuccess: () => {
      toast.success("Encryption migrated");
    },
    onError: (error: Error) => {
      toast.error(`Failed to migrate encryption: ${error.message}`);
    },
  });
}

// Rotate encryption key
export function useRotateEncryptionKey() {
  return useMutation({
    mutationFn: () => stateApi.rotateEncryptionKey(),
    onSuccess: () => {
      toast.success("Encryption key rotated");
    },
    onError: (error: Error) => {
      toast.error(`Failed to rotate encryption key: ${error.message}`);
    },
  });
}
