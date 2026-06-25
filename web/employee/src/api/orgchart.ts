import apiClient from './client';

export interface OrgChartEmployee {
  id: string;
  user_id: string;
  ffid: string;
  name: string;
  avatar_url?: string;
  department_id?: number;
  department_name?: string;
  manager_id?: string;
  job_title?: string;
  status: string;
}

export interface Goal {
  id: string;
  title: string;
  description?: string;
  level: string;
  progress: number;
  visibility: string;
  parent_id?: string;
  owner_id: string;
}

export const orgchartApi = {
  getOrgChart: () =>
    apiClient.get<{ employees: OrgChartEmployee[] }>('/v1/orgchart'),

  getDirectReports: (employeeId: string) =>
    apiClient.get<{ reports: OrgChartEmployee[] }>(`/v1/orgchart/${employeeId}/reports`),

  getGoals: () =>
    apiClient.get<{ goals: Goal[] }>('/v1/goals'),

  createGoal: (data: Partial<Goal>) =>
    apiClient.post<{ goal: Goal }>('/v1/goals', data),
};
