import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { githubApi } from '@/api/github';
import { githubKeys } from '@/hooks/useGitHubConnection';
import type { ListReposParams, GetTreeParams, ScanRepoOptions } from '@/types/github';

export function useGitHubRepos(params?: ListReposParams) {
  return useQuery({
    queryKey: githubKeys.repos(params as Record<string, unknown>),
    queryFn: () => githubApi.listRepos(params),
    staleTime: 1000 * 60 * 2,
  });
}

export function useGitHubRepo(repoId: string) {
  return useQuery({
    queryKey: githubKeys.repo(repoId),
    queryFn: () => githubApi.getRepo(repoId),
    enabled: !!repoId,
    staleTime: 1000 * 60,
  });
}

export function useRefreshGitHubRepos() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => githubApi.refreshRepos(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.repos() });
      toast.success(`Synced ${data.refreshed} repositories`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to refresh repositories: ${error.message}`);
    },
  });
}

export function useScanGitHubRepo(repoId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (options?: ScanRepoOptions) => githubApi.scanRepo(repoId, options),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.repo(repoId) });
      queryClient.invalidateQueries({ queryKey: githubKeys.repos() });
      queryClient.setQueryData(githubKeys.scan(repoId), data);
      toast.success(`Found ${data.functions.length} functions`);
    },
    onError: (error: Error) => {
      toast.error(`Scan failed: ${error.message}`);
    },
  });
}

export function useGitHubBranches(repoId: string) {
  return useQuery({
    queryKey: githubKeys.branches(repoId),
    queryFn: () => githubApi.listBranches(repoId),
    enabled: !!repoId,
    staleTime: 1000 * 60 * 5,
  });
}

export function useGitHubTree(repoId: string, params?: GetTreeParams) {
  return useQuery({
    queryKey: githubKeys.tree(repoId, params as Record<string, unknown>),
    queryFn: () => githubApi.getTree(repoId, params),
    enabled: !!repoId,
    staleTime: 1000 * 60,
  });
}
