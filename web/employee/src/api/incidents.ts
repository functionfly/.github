import apiClient from './client';

export interface Incident {
  id: string;
  tenant_id: string;
  title: string;
  description?: string;
  severity: string;
  status: string;
  commander_id?: string;
  project_id?: string;
  detected_at: string;
  acknowledged_at?: string;
  resolved_at?: string;
  closed_at?: string;
  root_cause?: string;
  impact?: string;
  duration_minutes?: number;
  created_at: string;
}

export interface IncidentEvent {
  id: number;
  incident_id: string;
  author_id: string;
  event_type: string;
  body: string;
  created_at: string;
}

export interface Postmortem {
  id: string;
  incident_id: string;
  summary: string;
  root_cause: string;
  contributing_factors?: string;
  what_went_well?: string;
  what_went_wrong?: string;
  action_items: { title: string; assignee_id?: string; due_date?: string; status: string }[];
  lessons_learned?: string;
  status: string;
  created_at: string;
}

export const incidentsApi = {
  list: (params?: { status?: string; severity?: string }) => apiClient.get<{ incidents: Incident[] }>('/v1/incidents', { params }),
  get: (id: string) => apiClient.get<{ incident: Incident }>(`/v1/incidents/${id}`),
  create: (data: Partial<Incident>) => apiClient.post<{ incident: Incident }>('/v1/incidents', data),
  update: (id: string, data: Partial<Incident>) => apiClient.patch(`/v1/incidents/${id}`, data),
  addEvent: (id: string, body: string, eventType?: string) => apiClient.post(`/v1/incidents/${id}/events`, { body, event_type: eventType }),
  listEvents: (id: string) => apiClient.get<{ events: IncidentEvent[] }>(`/v1/incidents/${id}/events`),
  addResponder: (id: string, employeeId: string, role?: string) => apiClient.post(`/v1/incidents/${id}/responders`, { employee_id: employeeId, role }),
  createPostmortem: (id: string, data: Partial<Postmortem>) => apiClient.post(`/v1/incidents/${id}/postmortem`, data),
  getPostmortem: (id: string) => apiClient.get<{ postmortem: Postmortem }>(`/v1/incidents/${id}/postmortem`),
};
