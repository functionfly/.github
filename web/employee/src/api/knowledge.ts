import apiClient from './client';

export interface KnowledgeArticle {
  id: string;
  tenant_id: string;
  title: string;
  slug: string;
  body: string;
  category?: string;
  tags?: string[];
  author_id: string;
  status: string;
  view_count: number;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

export const knowledgeApi = {
  list: (params?: { category?: string; status?: string; search?: string; limit?: number; offset?: number }) =>
    apiClient.get<{ articles: KnowledgeArticle[]; total: number }>('/v1/knowledge', { params }),

  get: (slug: string) =>
    apiClient.get<{ article: KnowledgeArticle }>(`/v1/knowledge/${slug}`),

  create: (data: Partial<KnowledgeArticle>) =>
    apiClient.post<{ article: KnowledgeArticle }>('/v1/knowledge', data),

  update: (id: string, data: Partial<KnowledgeArticle>) =>
    apiClient.patch<{ article: KnowledgeArticle }>(`/v1/knowledge/${id}`, data),

  search: (query: string) =>
    apiClient.get<{ articles: KnowledgeArticle[] }>('/v1/knowledge/search', { params: { q: query } }),
};
