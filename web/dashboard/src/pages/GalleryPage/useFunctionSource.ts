import { useQuery } from '@tanstack/react-query';
import { registryApi } from '@/api/registry';

export function useFunctionSource(author: string | undefined, name: string | undefined) {
  return useQuery({
    queryKey: ['function-source', author, name],
    enabled: !!author && !!name,
    staleTime: 5 * 60 * 1000,
    queryFn: () => registryApi.getFunctionVersionSource(author!, name!),
  });
}

export function useFunctionStats(author: string | undefined, name: string | undefined) {
  return useQuery({
    queryKey: ['function-stats', author, name],
    enabled: !!author && !!name,
    staleTime: 2 * 60 * 1000,
    queryFn: () => registryApi.getFunctionStats(author!, name!),
  });
}
