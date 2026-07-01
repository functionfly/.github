import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { appsApi } from '@/api/apps';
import type { App, AppAnalyticsResponse, CreateAppRequest, UpdateAppRequest, Backend, CreateBackendRequest } from '@/types';

// Query keys
export const appKeys = {
  all: ['apps'] as const,
  lists: () => [...appKeys.all, 'list'] as const,
  detail: (id: string) => [...appKeys.all, 'detail', id] as const,
  status: (id: string) => [...appKeys.all, 'status', id] as const,
  analytics: (id: string, days?: number) => [...appKeys.all, 'analytics', id, days] as const,
  backends: (appId: string) => [...appKeys.all, 'backends', appId] as const,
  secrets: (appId: string) => [...appKeys.all, 'secrets', appId] as const,
  route: (appId: string) => [...appKeys.all, 'route', appId] as const,
  deployOptions: () => [...appKeys.all, 'deploy-options'] as const,
};

// List apps
export function useApps() {
  return useQuery({
    queryKey: appKeys.lists(),
    queryFn: () => appsApi.list(),
    staleTime: 1000 * 60,
  });
}

// Get app
export function useApp(appId: string) {
  return useQuery({
    queryKey: appKeys.detail(appId),
    queryFn: () => appsApi.get(appId),
    enabled: !!appId,
    staleTime: 1000 * 60,
  });
}

// Get app status
export function useAppStatus(appId: string) {
  return useQuery({
    queryKey: appKeys.status(appId),
    queryFn: () => appsApi.getStatus(appId),
    enabled: !!appId,
    staleTime: 1000 * 30,
    refetchInterval: 30000,
  });
}

// Get app analytics
export function useAppAnalytics(appId: string, days = 7) {
  return useQuery<AppAnalyticsResponse>({
    queryKey: appKeys.analytics(appId, days),
    queryFn: () => appsApi.getAnalytics(appId, days),
    enabled: !!appId,
    staleTime: 1000 * 30,
    refetchInterval: 30000,
  });
}

// Create app
export function useCreateApp() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAppRequest) => appsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: appKeys.lists() });
      toast.success('App created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create app: ${error.message}`);
    },
  });
}

// Update app
export function useUpdateApp(appId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateAppRequest) => appsApi.update(appId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: appKeys.detail(appId) });
      queryClient.invalidateQueries({ queryKey: appKeys.status(appId) });
      queryClient.invalidateQueries({ queryKey: appKeys.lists() });
      toast.success('App updated successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update app: ${error.message}`);
    },
  });
}

// Delete app
export function useDeleteApp() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (appId: string) => appsApi.delete(appId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: appKeys.lists() });
      toast.success('App deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete app: ${error.message}`);
    },
  });
}

// List backends
export function useBackends(appId: string) {
  return useQuery({
    queryKey: appKeys.backends(appId),
    queryFn: () => appsApi.listBackends(appId),
    enabled: !!appId,
    staleTime: 1000 * 60,
  });
}

// Create backend
export function useCreateBackend() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      appId,
      data,
    }: {
      appId: string;
      data: CreateBackendRequest;
    }) => appsApi.createBackend(appId, data),
    onSuccess: (_, { appId }) => {
      queryClient.invalidateQueries({ queryKey: appKeys.backends(appId) });
      queryClient.invalidateQueries({ queryKey: appKeys.status(appId) });
      toast.success('Backend created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create backend: ${error.message}`);
    },
  });
}

// Delete backend
export function useDeleteBackend(appId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (backendId: string) => appsApi.deleteBackend(appId, backendId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: appKeys.backends(appId) });
      queryClient.invalidateQueries({ queryKey: appKeys.status(appId) });
      toast.success('Backend removed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to remove backend: ${error.message}`);
    },
  });
}

// List app secrets (provider-level, e.g. Fly.io)
export function useAppSecrets(appId: string) {
  return useQuery({
    queryKey: appKeys.secrets(appId),
    queryFn: () => appsApi.listSecrets(appId),
    enabled: !!appId,
    staleTime: 1000 * 60,
  });
}

// Get deployment options (for function deploy)
export function useDeployBackendOptions() {
  return useQuery({
    queryKey: appKeys.deployOptions(),
    queryFn: () => appsApi.list().then(({ apps }) => {
      return Promise.all(
        apps.map(async (app) => {
          const { backends } = await appsApi.listBackends(app.id);
          return backends.map((b) => ({
            id: b.id,
            appId: app.id,
            appName: app.name,
            provider: b.provider || '',
            region: b.region || '',
          }));
        })
      ).then((results) => results.flat());
    }),
    staleTime: 1000 * 60,
  });
}
