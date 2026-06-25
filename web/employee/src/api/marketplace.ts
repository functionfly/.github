import apiClient from './client';

export interface MarketplaceOpportunity {
  id: string;
  tenant_id: string;
  posted_by: string;
  department_id?: number;
  title: string;
  description: string;
  opportunity_type: string;
  skills_required: string[];
  hours_per_week?: number;
  duration_weeks?: number;
  is_remote: boolean;
  status: string;
  max_applicants?: number;
  created_at: string;
}

export interface MarketplaceApplication {
  id: string;
  opportunity_id: string;
  applicant_id: string;
  message?: string;
  status: string;
  created_at: string;
}

export const marketplaceApi = {
  list: (params?: { status?: string; type?: string }) => apiClient.get<{ opportunities: MarketplaceOpportunity[] }>('/v1/marketplace/opportunities', { params }),
  get: (id: string) => apiClient.get<{ opportunity: MarketplaceOpportunity }>(`/v1/marketplace/opportunities/${id}`),
  create: (data: Partial<MarketplaceOpportunity>) => apiClient.post<{ opportunity: MarketplaceOpportunity }>('/v1/marketplace/opportunities', data),
  apply: (id: string, message?: string) => apiClient.post(`/v1/marketplace/opportunities/${id}/apply`, { message }),
  reviewApplication: (id: string, action: 'accept' | 'reject') => apiClient.post(`/v1/marketplace/applications/${id}/review`, { action }),
};
