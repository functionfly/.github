import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { apiKeysApi } from '@/api/apikeys';
import type { APIKey, CreateAPIKeyRequest, APIKeyFilters, RotateAPIKeyRequest } from '@/types/api-key';

// Query keys
export const apiKeyKeys = {
  all: ['api-keys'] as const,
  lists: (filters?: APIKeyFilters) => [...apiKeyKeys.all, 'list', filters] as const,
  detail: (id: string) => [...apiKeyKeys.all, 'detail', id] as const,
  permissions: (id: string) => [...apiKeyKeys.all, 'permissions', id] as const,
  environments: (id: string) => [...apiKeyKeys.all, 'environments', id] as const,
};

// List API keys
export function useAPIKeys(filters?: APIKeyFilters) {
  return useQuery({
    queryKey: apiKeyKeys.lists(filters),
    queryFn: () => apiKeysApi.list(filters),
    staleTime: 1000 * 60,
  });
}

// Get single API key
export function useAPIKey(id: string) {
  return useQuery({
    queryKey: apiKeyKeys.detail(id),
    queryFn: () => apiKeysApi.get(id),
    enabled: !!id,
    staleTime: 1000 * 60,
  });
}

// Create API key
export function useCreateAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAPIKeyRequest) => apiKeysApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.all });
      toast.success('API key created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create API key: ${error.message}`);
    },
  });
}

// Update API key
export function useUpdateAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { name?: string; description?: string } }) =>
      apiKeysApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.lists() });
      toast.success('API key updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update API key: ${error.message}`);
    },
  });
}

// Delete API key
export function useDeleteAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiKeysApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.all });
      toast.success('API key deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete API key: ${error.message}`);
    },
  });
}

// Rotate API key
export function useRotateAPIKey() {
  const queryClient = useQueryClient();

  return useMutation({
    // Pass expires_in_days and reason through to the backend; the backend now
    // understands expires_in_days and applies it as a new ExpiresAt.
    mutationFn: ({ id, expiresInDays, reason }: { id: string; expiresInDays?: number; reason?: RotateAPIKeyRequest['reason'] }) => {
      const payload: RotateAPIKeyRequest = {};
      if (reason) payload.reason = reason;
      if (typeof expiresInDays === 'number' && Number.isFinite(expiresInDays) && expiresInDays > 0) {
        payload.expires_in_days = Math.floor(expiresInDays);
      }
      return apiKeysApi.rotate(id, Object.keys(payload).length > 0 ? payload : undefined);
    },
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.lists() });
      toast.success('API key rotated successfully - save your new key');
    },
    onError: (error: Error) => {
      toast.error(`Failed to rotate API key: ${error.message}`);
    },
  });
}

// List permissions
export function useAPIKeyPermissions(id: string) {
  return useQuery({
    queryKey: apiKeyKeys.permissions(id),
    queryFn: () => apiKeysApi.listPermissions(id),
    enabled: !!id,
    staleTime: 1000 * 60,
  });
}

// Add permission
export function useAddAPIKeyPermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ keyId, resourceType, resourceId, action }: {
      keyId: string;
      resourceType: import('@/types/api-key').ResourceType;
      resourceId: string;
      action: import('@/types/api-key').Permission;
    }) => apiKeysApi.addPermission(keyId, {
      permission: action,
      resource_type: resourceType,
      resource_id: resourceId,
    }),
    onSuccess: (_, { keyId }) => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.permissions(keyId) });
      toast.success('Permission added successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to add permission: ${error.message}`);
    },
  });
}

// Remove permission
export function useRemoveAPIKeyPermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ keyId, permId }: { keyId: string; permId: string }) =>
      apiKeysApi.removePermission(keyId, permId),
    onSuccess: (_, { keyId }) => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.permissions(keyId) });
      toast.success('Permission removed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove permission: ${error.message}`);
    },
  });
}

// List environments
export function useAPIKeyEnvironments(id: string) {
  return useQuery({
    queryKey: apiKeyKeys.environments(id),
    queryFn: () => apiKeysApi.listEnvironments(id),
    enabled: !!id,
    staleTime: 1000 * 60,
  });
}

// Add environment
export function useAddAPIKeyEnvironment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ keyId, name, value }: { keyId: string; name: string; value: string }) =>
      apiKeysApi.addEnvironment(keyId, { environment_id: value, environment_name: name }),
    onSuccess: (_, { keyId }) => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.environments(keyId) });
      toast.success('Environment variable added successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to add environment variable: ${error.message}`);
    },
  });
}

// Remove environment
export function useRemoveAPIKeyEnvironment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ keyId, envId }: { keyId: string; envId: string }) =>
      apiKeysApi.removeEnvironment(keyId, envId),
    onSuccess: (_, { keyId }) => {
      queryClient.invalidateQueries({ queryKey: apiKeyKeys.environments(keyId) });
      toast.success('Environment variable removed successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove environment variable: ${error.message}`);
    },
  });
}
