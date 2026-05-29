import { apiClient } from './client';

export interface StudioSearchResult {
  id: string;
  type: 'graph' | 'node' | 'plugin' | 'setting' | 'doc';
  title: string;
  description: string;
  path?: string;
  relevance: number;
  recent?: boolean;
}

export interface StudioSearchResponse {
  results: StudioSearchResult[];
  query: string;
  total: number;
}

export const studioSearchApi = {
  search: (params: { q: string; type?: string; limit?: number }) => {
    const search = new URLSearchParams({ q: params.q });
    if (params.type) search.set('type', params.type);
    if (params.limit !== undefined) search.set('limit', String(params.limit));
    return apiClient.get<StudioSearchResponse>(`/v1/studio/search?${search.toString()}`);
  },
};
