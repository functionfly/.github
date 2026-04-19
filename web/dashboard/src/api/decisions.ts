import { apiClient } from './client';

export type DecisionStatus = 'pending' | 'approved' | 'superseded' | 'deprecated';

export interface TeamDecision {
  id: string;
  team_id: string;
  title: string;
  description?: string;
  rationale?: string;
  outcome?: string;
  alternatives?: string[];
  tags?: string[];
  created_by: string;
  created_at: string;
  updated_at: string;
  status: DecisionStatus;
  approved_by?: string;
  approved_at?: string;
  source_type: 'manual' | 'ai_extracted' | 'imported';
  source_id?: string;
  importance_score: number;
}

export interface CreateDecisionRequest {
  title: string;
  description?: string;
  rationale?: string;
  outcome?: string;
  alternatives?: string[];
  tags?: string[];
  importance_score?: number;
}

export interface UpdateDecisionRequest {
  title?: string;
  description?: string;
  rationale?: string;
  outcome?: string;
  alternatives?: string[];
  tags?: string[];
  status?: DecisionStatus;
  importance_score?: number;
}

export interface ApproveDecisionRequest {
  status: 'approved' | 'superseded' | 'deprecated';
}

export interface ListDecisionsResponse {
  decisions: TeamDecision[];
  total_count: number;
  limit: number;
  offset: number;
}

export interface SearchDecisionsResponse {
  decisions: TeamDecision[];
  query: string;
  count: number;
}

interface ListDecisionsParams {
  status?: DecisionStatus;
  tag?: string;
  limit?: number;
  offset?: number;
}

export const decisionsApi = {
  // Create a new decision
  create: async (teamId: string, data: CreateDecisionRequest): Promise<TeamDecision> => {
    const response = await apiClient.post<TeamDecision>(`/v1/teams/${teamId}/decisions`, data);
    return response;
  },

  // List decisions for a team
  list: async (teamId: string, params?: ListDecisionsParams): Promise<ListDecisionsResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set('status', params.status);
    if (params?.tag) searchParams.set('tag', params.tag);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.offset) searchParams.set('offset', params.offset.toString());

    const queryString = searchParams.toString();
    const url = `/v1/teams/${teamId}/decisions${queryString ? `?${queryString}` : ''}`;

    const response = await apiClient.get<ListDecisionsResponse>(url);
    return response ?? { decisions: [], total_count: 0, limit: 20, offset: 0 };
  },

  // Get a specific decision
  get: async (teamId: string, decisionId: string): Promise<TeamDecision> => {
    const response = await apiClient.get<TeamDecision>(
      `/v1/teams/${teamId}/decisions/${decisionId}`
    );
    return response;
  },

  // Update a decision
  update: async (
    teamId: string,
    decisionId: string,
    data: UpdateDecisionRequest
  ): Promise<TeamDecision> => {
    const response = await apiClient.put<TeamDecision>(
      `/v1/teams/${teamId}/decisions/${decisionId}`,
      data
    );
    return response;
  },

  // Delete a decision
  delete: async (teamId: string, decisionId: string): Promise<void> => {
    await apiClient.delete(`/v1/teams/${teamId}/decisions/${decisionId}`);
  },

  // Approve/supersede/deprecate a decision
  approve: async (
    teamId: string,
    decisionId: string,
    data: ApproveDecisionRequest
  ): Promise<TeamDecision> => {
    const response = await apiClient.post<TeamDecision>(
      `/v1/teams/${teamId}/decisions/${decisionId}/approve`,
      data
    );
    return response;
  },

  // Search decisions
  search: async (
    teamId: string,
    query: string,
    limit?: number
  ): Promise<SearchDecisionsResponse> => {
    const searchParams = new URLSearchParams({ q: query });
    if (limit) searchParams.set('limit', limit.toString());

    const response = await apiClient.get<SearchDecisionsResponse>(
      `/v1/teams/${teamId}/decisions/search?${searchParams.toString()}`
    );
    return response ?? { decisions: [], query, count: 0 };
  },
};
