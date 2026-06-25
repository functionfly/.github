import apiClient from './client';

export interface MissionControlSnapshot {
  id: string;
  tenant_id: string;
  snapshot_date: string;
  total_employees: number;
  active_employees: number;
  new_hires_30d: number;
  departures_30d: number;
  total_projects: number;
  active_projects: number;
  completed_projects_30d: number;
  total_tasks: number;
  completed_tasks_30d: number;
  avg_task_completion_days: number;
  total_learning_hours: number;
  avg_skill_proficiency: number;
  innovation_grants_submitted: number;
  innovation_grants_funded: number;
  pto_days_used_30d: number;
  avg_burnout_risk: number;
}

export const missionControlApi = {
  get: () => apiClient.get<{ snapshot: MissionControlSnapshot }>('/v1/mission-control'),
  refresh: () => apiClient.post('/v1/mission-control/refresh'),
};
