import { useQuery } from '@tanstack/react-query';
import { studioSearchApi } from '@/api/studioSearch';

const SEARCH_KEY = 'studio-search';

export function useStudioSearch(query: string, type?: string, enabled = true) {
  const trimmed = query.trim();
  return useQuery({
    queryKey: [SEARCH_KEY, trimmed, type],
    queryFn: () => studioSearchApi.search({ q: trimmed, type, limit: 30 }),
    enabled: enabled && trimmed.length >= 2,
    staleTime: 1000 * 30,
  });
}
