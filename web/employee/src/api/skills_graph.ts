import apiClient from './client';

export interface SkillGraphNode {
  id: string;
  skill_name: string;
  category: string;
  total_employees: number;
  avg_proficiency: number;
  demand_score: number;
  supply_score: number;
  gap_score: number;
  trending: boolean;
}

export const skillsGraphApi = {
  get: (params?: { category?: string }) => apiClient.get<{ skills: SkillGraphNode[] }>('/v1/skills-graph', { params }),
  getGap: () => apiClient.get<{ gaps: SkillGraphNode[] }>('/v1/skills-graph/gaps'),
};
