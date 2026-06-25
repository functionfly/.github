import apiClient from './client';

export interface LivingMemoryEntry {
  id: string;
  tenant_id: string;
  author_id: string;
  title: string;
  body: string;
  memory_type: string;
  project_id?: string;
  tags: string[];
  participants: string[];
  importance: string;
  view_count: number;
  created_at: string;
}

export const livingMemoryApi = {
  search: (query: string, params?: { memory_type?: string }) => apiClient.get<{ results: LivingMemoryEntry[] }>('/v1/memory/search', { params: { q: query, ...params } }),
  list: (params?: { memory_type?: string; project_id?: string }) => apiClient.get<{ entries: LivingMemoryEntry[] }>('/v1/memory', { params }),
  get: (id: string) => apiClient.get<{ entry: LivingMemoryEntry }>(`/v1/memory/${id}`),
  create: (data: Partial<LivingMemoryEntry>) => apiClient.post<{ entry: LivingMemoryEntry }>('/v1/memory', data),
};
