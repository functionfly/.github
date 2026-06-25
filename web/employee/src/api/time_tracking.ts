import apiClient from './client';

export interface TimeEntry {
  id: string;
  employee_id: string;
  project_id?: string;
  task_id?: string;
  date: string;
  hours: number;
  description?: string;
  entry_type: string;
  is_billable: boolean;
  created_at: string;
}

export interface PTORequest {
  id: string;
  employee_id: string;
  pto_type: string;
  start_date: string;
  end_date: string;
  days: number;
  reason?: string;
  status: string;
  notes?: string;
  created_at: string;
}

export const timeTrackingApi = {
  listEntries: (params?: { start_date?: string; end_date?: string; project_id?: string }) =>
    apiClient.get<{ entries: TimeEntry[] }>('/v1/time-entries', { params }),
  createEntry: (data: Partial<TimeEntry>) => apiClient.post<{ entry: TimeEntry }>('/v1/time-entries', data),
  updateEntry: (id: string, data: Partial<TimeEntry>) => apiClient.patch(`/v1/time-entries/${id}`, data),
  deleteEntry: (id: string) => apiClient.delete(`/v1/time-entries/${id}`),
  listPTO: () => apiClient.get<{ requests: PTORequest[] }>('/v1/pto/requests'),
  requestPTO: (data: Partial<PTORequest>) => apiClient.post<{ request: PTORequest }>('/v1/pto/requests', data),
  approvePTO: (id: string) => apiClient.post(`/v1/pto/requests/${id}/approve`),
  rejectPTO: (id: string, notes?: string) => apiClient.post(`/v1/pto/requests/${id}/reject`, { notes }),
};
