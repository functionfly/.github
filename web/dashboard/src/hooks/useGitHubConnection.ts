import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { githubApi } from '@/api/github';
import { useGitHubStore } from '@/stores/githubStore';

export const githubKeys = {
  all: ['github'] as const,
  connection: () => [...githubKeys.all, 'connection'] as const,
  repos: (params?: Record<string, unknown>) => [...githubKeys.all, 'repos', params] as const,
  repo: (repoId: string) => [...githubKeys.all, 'repo', repoId] as const,
  branches: (repoId: string) => [...githubKeys.all, 'branches', repoId] as const,
  tree: (repoId: string, params?: Record<string, unknown>) => [...githubKeys.all, 'tree', repoId, params] as const,
  scan: (repoId: string) => [...githubKeys.all, 'scan', repoId] as const,
  imports: (params?: Record<string, unknown>) => [...githubKeys.all, 'imports', params] as const,
  importItem: (importId: string) => [...githubKeys.all, 'import', importId] as const,
  syncLogs: (importId: string, params?: Record<string, unknown>) =>
    [...githubKeys.all, 'syncLogs', importId, params] as const,
  templates: () => [...githubKeys.all, 'templates'] as const,
};

export function useGitHubConnection() {
  return useQuery({
    queryKey: githubKeys.connection(),
    queryFn: () => githubApi.getConnection(),
    retry: false,
    staleTime: 0,
    refetchOnMount: 'always',
    refetchOnWindowFocus: true,
  });
}

export function useGitHubConnect() {
  return useMutation({
    mutationFn: () => githubApi.getConnectUrl(),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
    onError: (error: Error) => {
      toast.error(`Failed to connect GitHub: ${error.message}`);
    },
  });
}

export function useGitHubDisconnect() {
  const queryClient = useQueryClient();
  const setConnection = useGitHubStore((s) => s.setConnection);

  return useMutation({
    mutationFn: () => githubApi.disconnect(),
    onSuccess: () => {
      // Use resetQueries so the cached `data` becomes undefined instantly.
      // This guarantees the page immediately switches from "Connected UI" to "NoGitHubConnection".
      queryClient.resetQueries({ queryKey: githubKeys.connection() });
      queryClient.resetQueries({ queryKey: githubKeys.repos() });
      queryClient.resetQueries({ queryKey: ['github', 'imports'] });
      queryClient.resetQueries({ queryKey: githubKeys.templates() });
      setConnection(null);
      toast.success('GitHub account disconnected');
    },
    onError: (error: Error) => {
      toast.error(`Failed to disconnect GitHub: ${error.message}`);
    },
  });
}

export function useGitHubTokenRefresh() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => githubApi.refreshToken(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.connection() });
      toast.success('GitHub token refreshed');
    },
    onError: (error: Error) => {
      toast.error(`Failed to refresh token: ${error.message}`);
    },
  });
}
