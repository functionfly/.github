import apiClient from './client';

export interface CareerPath {
  id: string;
  tenant_id: string;
  title: string;
  track: string;
  level: number;
  description?: string;
  requirements: { skills?: string[]; years_exp?: number; certs?: string[] };
  salary_range_min_cents?: number;
  salary_range_max_cents?: number;
  next_path_id?: string;
  is_active: boolean;
}

export interface CareerProgress {
  id: number;
  employee_id: string;
  career_path_id: string;
  status: string;
  gap_analysis: { missing_skills?: string[]; missing_certs?: string[] };
  target_date?: string;
}

export const careerApi = {
  listPaths: (params?: { track?: string }) => apiClient.get<{ paths: CareerPath[] }>('/v1/career/paths', { params }),
  getPath: (id: string) => apiClient.get<{ path: CareerPath }>(`/v1/career/paths/${id}`),
  getMyProgress: () => apiClient.get<{ progress: CareerProgress[] }>('/v1/career/progress'),
  setTarget: (careerPathId: string, targetDate?: string) => apiClient.post('/v1/career/progress', { career_path_id: careerPathId, target_date: targetDate }),
  getGapAnalysis: (pathId: string) => apiClient.get<{ gap_analysis: { missing_skills: string[]; missing_certs: string[]; match_pct: number } }>(`/v1/career/paths/${pathId}/gap-analysis`),
};
