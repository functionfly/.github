import apiClient from './client';

export interface OrgChartImport {
  id: string;
  file_name: string;
  status: string;
  total_rows: number;
  processed_rows: number;
  error_rows: number;
  errors: { row: number; message: string }[];
  created_at: string;
}

export const orgImportApi = {
  upload: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.post<{ import: OrgChartImport }>('/v1/orgchart/import', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
  },
  getStatus: (id: string) => apiClient.get<{ import: OrgChartImport }>(`/v1/orgchart/import/${id}`),
};
