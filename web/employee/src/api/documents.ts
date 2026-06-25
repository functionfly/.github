import apiClient from './client';

export interface Document {
  id: string;
  tenant_id: string;
  author_id: string;
  title: string;
  body?: string;
  doc_type: string;
  category?: string;
  tags?: string[];
  is_template: boolean;
  status: string;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export const documentsApi = {
  list: (params?: { doc_type?: string; category?: string; status?: string }) => apiClient.get<{ documents: Document[] }>('/v1/documents', { params }),
  get: (id: string) => apiClient.get<{ document: Document }>(`/v1/documents/${id}`),
  create: (data: Partial<Document>) => apiClient.post<{ document: Document }>('/v1/documents', data),
  update: (id: string, data: Partial<Document>) => apiClient.patch(`/v1/documents/${id}`, data),
  share: (id: string, employeeId: string, permission?: string) => apiClient.post(`/v1/documents/${id}/share`, { employee_id: employeeId, permission }),
  listTemplates: () => apiClient.get<{ documents: Document[] }>('/v1/documents/templates'),
};
