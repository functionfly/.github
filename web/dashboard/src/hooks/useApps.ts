import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { appsApi } from '@/api/apps';
import type { App, CreateAppRequest, Backend, CreateBackendRequest } from '@/types';

// Query keys
export const appKeys = {
  all: ['apps'] as const,
  lists: () => [...appKeys.all, 'list'] as const,
  detail: (id: string) => [...appKeys.all, 'detail', id] as const,
  status: (id: string) => [...appKeys.all, 'status', id] as const,
  backends: (appId: string) => [...appKeys.all, 'backends', appId] as const,
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
    refetchInterval: 30000, // Auto-refresh status every 30s
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
      toast.success('Backend created successfully');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create backend: ${error.message}`);
    },
  });
}

// Get deployment options (for function deploy)
export function useDeployBackendOptions() {
  return useQuery({
    queryKey: appKeys.deployOptions(),
    queryFn: () => appsApi.list().then(({ apps }) => {
      // Fetch backends for all apps and flatten
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
