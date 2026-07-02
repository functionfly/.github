import { apiClient } from './client';
import type { UnifiedSearchResponse, MarketplaceSearchParams } from '@/components/marketplace/types';

export interface AgentRating {
  id: string;
  agent_id: string;
  tenant_id: string;
  rating: number;
  review?: string;
  username?: string;
  user_name?: string;
  created_at: string;
  updated_at: string;
}

export const marketplaceUnifiedApi = {
  search: async (params?: MarketplaceSearchParams): Promise<UnifiedSearchResponse> => {
    const searchParams: Record<string, string | number | undefined> = {};
    if (params?.q) searchParams.q = params.q;
    if (params?.type) searchParams.type = params.type;
    if (params?.limit) searchParams.limit = params.limit;
    if (params?.offset) searchParams.offset = params.offset;

    const response = await apiClient.get('/v1/marketplace/search', { params: searchParams });
    return response as UnifiedSearchResponse;
  },

  rateAgent: async (agentId: string, rating: number, review?: string): Promise<{ message: string }> => {
    return apiClient.post(`/v1/marketplace/agents/${agentId}/rate`, { rating, review }) as Promise<{ message: string }>;
  },

  getMyAgentRating: async (agentId: string): Promise<{ rating: AgentRating | null }> => {
    return apiClient.get(`/v1/marketplace/agents/${agentId}/my-rating`) as Promise<{ rating: AgentRating | null }>;
  },

  listAgentRatings: async (agentId: string, limit = 50): Promise<{ ratings: AgentRating[] }> => {
    return apiClient.get(`/v1/marketplace/agents/${agentId}/ratings`, { params: { limit } }) as Promise<{ ratings: AgentRating[] }>;
  },

  rateFunction: async (functionId: string, rating: number, review?: string): Promise<{ message: string }> => {
    return apiClient.post(`/v1/marketplace/functions/${functionId}/rate`, { rating, review }) as Promise<{ message: string }>;
  },

  getMyFunctionRating: async (functionId: string): Promise<{ rating: AgentRating | null }> => {
    return apiClient.get(`/v1/marketplace/functions/${functionId}/my-rating`) as Promise<{ rating: AgentRating | null }>;
  },

  listFunctionRatings: async (functionId: string, limit = 50): Promise<{ ratings: AgentRating[] }> => {
    return apiClient.get(`/v1/marketplace/functions/${functionId}/ratings`, { params: { limit } }) as Promise<{ ratings: AgentRating[] }>;
  },
};
