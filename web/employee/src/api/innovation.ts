import apiClient from './client';

export interface InnovationGrant {
  id: string;
  tenant_id: string;
  proposer_id: string;
  title: string;
  description: string;
  category: string;
  requested_amount_cents?: number;
  status: string;
  feasibility_score?: number;
  votes_for: number;
  votes_against: number;
  created_at: string;
}

export interface InnovationVote {
  id: number;
  grant_id: string;
  voter_id: string;
  vote: boolean;
  comment?: string;
  created_at: string;
}

export const innovationApi = {
  list: (params?: { status?: string }) => apiClient.get<{ grants: InnovationGrant[] }>('/v1/innovation/grants', { params }),
  get: (id: string) => apiClient.get<{ grant: InnovationGrant }>(`/v1/innovation/grants/${id}`),
  create: (data: Partial<InnovationGrant>) => apiClient.post<{ grant: InnovationGrant }>('/v1/innovation/grants', data),
  submit: (id: string) => apiClient.post(`/v1/innovation/grants/${id}/submit`),
  vote: (id: string, vote: boolean, comment?: string) => apiClient.post(`/v1/innovation/grants/${id}/vote`, { vote, comment }),
  review: (id: string, action: 'approve' | 'reject', reason?: string) => apiClient.post(`/v1/innovation/grants/${id}/review`, { action, reason }),
};
