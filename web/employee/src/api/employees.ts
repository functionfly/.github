import apiClient from './client';

export interface Employee {
  id: string;
  user_id: string;
  tenant_id: string;
  employee_number: string;
  ffid: string;
  department_id?: number;
  manager_id?: string;
  hire_date?: string;
  employment_type: string;
  clearance_level: string;
  work_location?: string;
  office_location?: string;
  timezone?: string;
  bio?: string;
  pronouns?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface ListEmployeesOpts {
  department_id?: number;
  status?: string;
  search?: string;
  limit?: number;
  offset?: number;
}

export const employeesApi = {
  list: (opts?: ListEmployeesOpts) =>
    apiClient.get<{ employees: Employee[]; total: number }>('/v1/employees', { params: opts }),

  get: (id: string) =>
    apiClient.get<{ employee: Employee }>(`/v1/employees/${id}`),

  getByFFID: (ffid: string) =>
    apiClient.get<{ employee: Employee }>(`/v1/employees/ffid/${ffid}`),

  create: (data: Partial<Employee>) =>
    apiClient.post<{ employee: Employee }>('/v1/employees', data),

  update: (id: string, data: Partial<Employee>) =>
    apiClient.patch<{ employee: Employee }>(`/v1/employees/${id}`, data),
};
