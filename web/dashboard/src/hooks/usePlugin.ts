import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { pluginsApi, type Plugin, type PluginFilters, type InstallPluginRequest, type UpdateSandboxRequest, type SetPermissionRequest, type RateLimitCheckRequest } from '@/api/plugins';

export { type Plugin, type PluginFilters, type InstallPluginRequest, type UpdateSandboxRequest, type SetPermissionRequest };
export type PluginType = 'code-intelligence' | 'data-visualization' | 'futuristic' | 'marketplace' | 'security' | 'observability' | 'custom';
export type SandboxTier = 'free' | 'basic' | 'pro' | 'enterprise';

export const pluginKeys = {
  all: ['plugins'] as const,
  lists: () => [...pluginKeys.all, 'list'] as const,
  list: (filters: PluginFilters) => [...pluginKeys.lists(), filters] as const,
  details: () => [...pluginKeys.all, 'detail'] as const,
  detail: (id: string) => [...pluginKeys.details(), id] as const,
  sandbox: (id: string) => [...pluginKeys.detail(id), 'sandbox'] as const,
  permissions: (id: string) => [...pluginKeys.detail(id), 'permissions'] as const,
  versions: (id: string) => [...pluginKeys.detail(id), 'versions'] as const,
  telemetry: (id: string, range: string) => [...pluginKeys.detail(id), 'telemetry', range] as const,
};

export function usePlugins(filters?: PluginFilters) {
  return useQuery({
    queryKey: pluginKeys.list(filters || {}),
    queryFn: () => pluginsApi.list(filters),
    staleTime: 1000 * 60,
  });
}

export function usePlugin(pluginId: string) {
  return useQuery({
    queryKey: pluginKeys.detail(pluginId),
    queryFn: () => pluginsApi.get(pluginId),
    enabled: !!pluginId,
  });
}

export function useInstallPlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: InstallPluginRequest) => pluginsApi.install(data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success(`Plugin "${result.plugin.name}" installed`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to install plugin: ${error.message}`);
    },
  });
}

export function useUninstallPlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (pluginId: string) => pluginsApi.uninstall(pluginId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Plugin uninstalled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to uninstall plugin: ${error.message}`);
    },
  });
}

export function useUpdatePlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { pluginId: string; data: Partial<Plugin> }) =>
      pluginsApi.update(params.pluginId, params.data),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      queryClient.setQueryData(pluginKeys.detail(result.plugin.id), { plugin: result.plugin });
      toast.success(`Plugin "${result.plugin.name}" updated`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to update plugin: ${error.message}`);
    },
  });
}

export function useEnablePlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (pluginId: string) => pluginsApi.enable(pluginId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Plugin enabled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to enable plugin: ${error.message}`);
    },
  });
}

export function useDisablePlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (pluginId: string) => pluginsApi.disable(pluginId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Plugin disabled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to disable plugin: ${error.message}`);
    },
  });
}

export function usePausePlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (pluginId: string) => pluginsApi.pause(pluginId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Plugin paused');
    },
    onError: (error: Error) => {
      toast.error(`Failed to pause plugin: ${error.message}`);
    },
  });
}

export function useRollbackPlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { pluginId: string; toVersion?: string }) =>
      pluginsApi.rollback(params.pluginId, params.toVersion),
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success(`Rolled back to version ${result.rolled_back_to}`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to rollback plugin: ${error.message}`);
    },
  });
}

export function useConfigurePlugin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { pluginId: string; config: Record<string, string> }) =>
      pluginsApi.configure(params.pluginId, params.config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Plugin configured');
    },
    onError: (error: Error) => {
      toast.error(`Failed to configure plugin: ${error.message}`);
    },
  });
}

export function usePluginSandbox(pluginId: string) {
  return useQuery({
    queryKey: pluginKeys.sandbox(pluginId),
    queryFn: () => pluginsApi.getSandbox(pluginId),
    enabled: !!pluginId,
  });
}

export function useUpdateSandbox() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { pluginId: string; data: UpdateSandboxRequest }) =>
      pluginsApi.updateSandbox(params.pluginId, params.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Sandbox updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update sandbox: ${error.message}`);
    },
  });
}

export function usePluginPermissions(pluginId: string) {
  return useQuery({
    queryKey: pluginKeys.permissions(pluginId),
    queryFn: () => pluginsApi.getPermissions(pluginId),
    enabled: !!pluginId,
  });
}

export function useSetPermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { pluginId: string; data: SetPermissionRequest }) =>
      pluginsApi.setPermission(params.pluginId, params.data),
    onSuccess: () => {
      toast.success('Permission updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to set permission: ${error.message}`);
    },
  });
}

export function usePluginVersions(pluginId: string) {
  return useQuery({
    queryKey: pluginKeys.versions(pluginId),
    queryFn: () => pluginsApi.listVersions(pluginId),
    enabled: !!pluginId,
  });
}

export function usePluginTelemetry(pluginId: string, timeRange: string = "7d") {
  return useQuery({
    queryKey: pluginKeys.telemetry(pluginId, timeRange),
    queryFn: () => pluginsApi.getTelemetry(pluginId, timeRange),
    enabled: !!pluginId,
    staleTime: 1000 * 60 * 5,
  });
}

export function useSetPluginError() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { pluginId: string; error: string }) =>
      pluginsApi.setError(params.pluginId, params.error),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pluginKeys.all });
      toast.success('Error recorded');
    },
    onError: (error: Error) => {
      toast.error(`Failed to set error: ${error.message}`);
    },
  });
}

export function useCheckRateLimit() {
  return useMutation({
    mutationFn: (data: RateLimitCheckRequest) => pluginsApi.checkRateLimit(data),
    onError: (error: Error) => {
      toast.error(`Rate limit check failed: ${error.message}`);
    },
  });
}

export function useRecordAnalytics() {
  return useMutation({
    mutationFn: (params: { pluginId: string; data: any }) =>
      pluginsApi.recordAnalytics(params.pluginId, params.data),
    onError: (error: Error) => {
      console.error('Failed to record analytics:', error.message);
    },
  });
}