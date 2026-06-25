import apiClient from './client';

export interface LifecycleEvent {
  id: string;
  employee_id: string;
  event_type: string;
  payload: Record<string, unknown>;
  notes?: string;
  created_at: string;
}

export interface LifecycleWorkflow {
  id: string;
  name: string;
  description?: string;
  trigger_event: string;
  steps: { title: string; assignee_role: string; due_days: number; description: string }[];
  is_active: boolean;
}

export interface LifecycleWorkflowInstance {
  id: string;
  workflow_id: string;
  employee_id: string;
  status: string;
  current_step: number;
  steps_status: { step_idx: number; status: string; completed_by?: string; completed_at?: string; notes?: string }[];
  started_at: string;
  completed_at?: string;
}

export const lifecycleApi = {
  listEvents: (params?: { employee_id?: string; event_type?: string }) => apiClient.get<{ events: LifecycleEvent[] }>('/v1/lifecycle/events', { params }),
  listWorkflows: () => apiClient.get<{ workflows: LifecycleWorkflow[] }>('/v1/lifecycle/workflows'),
  createWorkflow: (data: Partial<LifecycleWorkflow>) => apiClient.post<{ workflow: LifecycleWorkflow }>('/v1/lifecycle/workflows', data),
  getInstance: (id: string) => apiClient.get<{ instance: LifecycleWorkflowInstance }>(`/v1/lifecycle/instances/${id}`),
  completeStep: (instanceId: string, stepIdx: number, notes?: string) => apiClient.post(`/v1/lifecycle/instances/${instanceId}/steps/${stepIdx}/complete`, { notes }),
};
