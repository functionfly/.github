import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { API_BASE_URL } from '@/lib/constants';

export interface Extension {
  id: string;
  name: string;
  version: string;
  description: string;
  author: { name: string };
  status: 'enabled' | 'disabled' | 'error';
  permissions: string[];
  hooks: string[];
  size: number;
  installedAt: string;
  updatedAt: string;
  category: string;
  error?: string;
}

export interface ExtensionConfig {
  key: string;
  value: string;
}

async function extensionFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
}

export function useExtensions() {
  return useQuery({
    queryKey: ['extensions'],
    queryFn: () => extensionFetch<{ extensions: Extension[] }>('/v1/extensions'),
    staleTime: 1000 * 60,
  });
}

export function useInstallExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (extensionId: string) =>
      extensionFetch<{ ok: boolean }>(`/v1/extensions/${extensionId}/install`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['extensions'] });
      toast.success('Extension installed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to install extension: ${error.message}`);
    },
  });
}

export function useUninstallExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (extensionId: string) =>
      extensionFetch<{ ok: boolean }>(`/v1/extensions/${extensionId}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['extensions'] });
      toast.success('Extension uninstalled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to uninstall extension: ${error.message}`);
    },
  });
}

export function useEnableExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (extensionId: string) =>
      extensionFetch<{ ok: boolean }>(`/v1/extensions/${extensionId}/enable`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['extensions'] });
      toast.success('Extension enabled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to enable extension: ${error.message}`);
    },
  });
}

export function useDisableExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (extensionId: string) =>
      extensionFetch<{ ok: boolean }>(`/v1/extensions/${extensionId}/disable`, {
        method: 'POST',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['extensions'] });
      toast.success('Extension disabled');
    },
    onError: (error: Error) => {
      toast.error(`Failed to disable extension: ${error.message}`);
    },
  });
}

export function useConfigureExtension() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { extensionId: string; config: ExtensionConfig[] }) =>
      extensionFetch<{ ok: boolean }>(`/v1/extensions/${params.extensionId}/config`, {
        method: 'PUT',
        body: JSON.stringify({ config: params.config }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['extensions'] });
      toast.success('Extension configured');
    },
    onError: (error: Error) => {
      toast.error(`Failed to configure extension: ${error.message}`);
    },
  });
}

export function useExtensionHooks() {
  return useQuery({
    queryKey: ['extensions', 'hooks'],
    queryFn: () => extensionFetch<{ hooks: Array<{ id: string; name: string; description: string; extensionId: string; events: string[]; enabled: boolean }> }>('/v1/extensions/hooks'),
    staleTime: 1000 * 60,
  });
}