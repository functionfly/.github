import apiClient from './client';

export interface ReputationScore {
  id: string;
  employee_id: string;
  category: string;
  score: number;
  rank: number;
  percentile: number;
  components: Record<string, number>;
}

export const reputationApi = {
  get: (employeeId: string) => apiClient.get<{ scores: ReputationScore[] }>(`/v1/reputation/${employeeId}`),
  getLeaderboard: (category: string) => apiClient.get<{ leaderboard: { employee_id: string; ffid: string; score: number; rank: number }[] }>('/v1/reputation/leaderboard', { params: { category } }),
};
