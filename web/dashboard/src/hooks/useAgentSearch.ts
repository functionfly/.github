import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { agentSearchApi, type ExecuteSearchRequest } from '@/api/agentSearch';
import { toast } from 'sonner';

// Query keys factory
export const searchKeys = {
  all: ['search'] as const,
  tools: () => [...searchKeys.all, 'tools'] as const,
  stats: (toolName: string) => [...searchKeys.all, 'stats', toolName] as const,
};

export function useSearchTools() {
  return useQuery({
    queryKey: searchKeys.tools(),
    queryFn: () => agentSearchApi.listSearchTools(),
  });
}

export function useSearchStats(toolName: string, since?: string) {
  return useQuery({
    queryKey: searchKeys.stats(toolName),
    queryFn: () => agentSearchApi.getSearchStats(toolName, since),
    enabled: !!toolName,
  });
}

export function useExecuteSearchTool() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: ExecuteSearchRequest) => agentSearchApi.executeSearchTool(req),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: searchKeys.all });
      toast.success(`Search completed: ${data.resultsCount} results`);
    },
    onError: (error: Error) => {
      toast.error(`Search failed: ${error.message}`);
    },
  });
}