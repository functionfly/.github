import { useQuery, useQueryClient } from '@tanstack/react-query';
import { atlasApi, type AtlasSearchRequest } from '@/api/atlas';

export const atlasKeys = {
  all: ['atlas'] as const,
  traces: () => [...atlasKeys.all, 'traces'] as const,
  trace: (runId: string) => [...atlasKeys.all, 'trace', runId] as const,
  graph: (runId: string) => [...atlasKeys.all, 'graph', runId] as const,
  search: (req: AtlasSearchRequest) => [...atlasKeys.all, 'search', req] as const,
  health: () => [...atlasKeys.all, 'health'] as const,
};

export function useAtlasHealth() {
  return useQuery({
    queryKey: atlasKeys.health(),
    queryFn: () => atlasApi.health(),
    staleTime: 1000 * 30,
    retry: false,
  });
}

export function useAtlasTraces(limit = 50) {
  return useQuery({
    queryKey: atlasKeys.traces(),
    queryFn: () => atlasApi.listTraces(limit),
    staleTime: 1000 * 10,
  });
}

export function useAtlasTrace(runId: string) {
  return useQuery({
    queryKey: atlasKeys.trace(runId),
    queryFn: () => atlasApi.getTrace(runId),
    enabled: !!runId,
    staleTime: 1000 * 30,
  });
}

export function useAtlasGraph(runId: string) {
  return useQuery({
    queryKey: atlasKeys.graph(runId),
    queryFn: () => atlasApi.getGraph(runId),
    enabled: !!runId,
    staleTime: 1000 * 30,
  });
}

export function useAtlasSearch(req: AtlasSearchRequest, enabled = true) {
  return useQuery({
    queryKey: atlasKeys.search(req),
    queryFn: () => atlasApi.searchTraces(req),
    enabled: enabled && (!!req.kind || !!req.system_id || !!req.since_ns),
    staleTime: 1000 * 10,
  });
}
