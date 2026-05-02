import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { githubApi } from '@/api/github';
import { githubKeys } from '@/hooks/useGitHubConnection';
import type { UpdateSyncRequest, ListSyncLogsParams } from '@/types/github';

export function useUpdateSync(importId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateSyncRequest) => githubApi.updateSync(importId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.importItem(importId) });
      queryClient.invalidateQueries({ queryKey: githubKeys.imports() });
      toast.success('Sync settings updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update sync: ${error.message}`);
    },
  });
}

export function useSyncLogs(importId: string, params?: ListSyncLogsParams) {
  return useQuery({
    queryKey: githubKeys.syncLogs(importId, params as Record<string, unknown>),
    queryFn: () => githubApi.getSyncLogs(importId, params),
    enabled: !!importId,
    staleTime: 1000 * 30,
  });
}
