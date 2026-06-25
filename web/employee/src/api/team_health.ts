import apiClient from './client';

export interface TeamHealthMetric {
  id: string;
  department_id?: number;
  metric_date: string;
  workload_score: number;
  burnout_risk: number;
  velocity_score: number;
  collaboration_score: number;
  knowledge_sharing_score: number;
  pto_utilization_pct: number;
  avg_overtime_hours: number;
  headcount: number;
}

export const teamHealthApi = {
  get: (params?: { department_id?: number }) => apiClient.get<{ metrics: TeamHealthMetric[] }>('/v1/team-health', { params }),
  getBurnoutRisk: () => apiClient.get<{ at_risk: { employee_id: string; ffid: string; risk_score: number }[] }>('/v1/team-health/burnout-risk'),
};
