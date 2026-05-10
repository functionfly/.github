import { apiClient } from './client';
import type {
  DNAProfile,
  DNAMutation,
  MutationListResponse,
  DNAInsights,
  EnterpriseInsights,
} from '@/types/dna';

export const dnaApi = {
  /** Get the DNA profile for a function */
  getProfile: (functionId: string, type: string = 'registry') =>
    apiClient.get<DNAProfile>(`/v1/functions/${functionId}/dna`, {
      params: { type },
    }),

  /** List mutations for a function */
  listMutations: (
    functionId: string,
    params?: { status?: string; limit?: number; offset?: number }
  ) =>
    apiClient.get<MutationListResponse>(
      `/v1/functions/${functionId}/dna/mutations`,
      { params }
    ),

  /** Get a single mutation with full code details */
  getMutation: (functionId: string, mutationId: string) =>
    apiClient.get<DNAMutation>(
      `/v1/functions/${functionId}/dna/mutations/${mutationId}`
    ),

  /** Accept a proposed variant */
  acceptVariant: (
    functionId: string,
    mutationId: string,
    canaryPercentage: number = 10
  ) =>
    apiClient.post(
      `/v1/functions/${functionId}/dna/variants/${mutationId}/accept`,
      { canary_percentage: canaryPercentage }
    ),

  /** Reject a proposed variant */
  rejectVariant: (functionId: string, mutationId: string, reason?: string) =>
    apiClient.post(
      `/v1/functions/${functionId}/dna/variants/${mutationId}/reject`,
      { reason }
    ),

  /** Rollback a deployed mutation */
  rollbackVariant: (functionId: string, mutationId: string, reason?: string) =>
    apiClient.post(
      `/v1/functions/${functionId}/dna/variants/${mutationId}/rollback`,
      { reason }
    ),

  /** Get time-series DNA insights for a function */
  getInsights: (functionId: string, period: string = '30d') =>
    apiClient.get<DNAInsights>(`/v1/functions/${functionId}/dna/insights`, {
      params: { period },
    }),

  /** Manually trigger DNA analysis */
  triggerAnalysis: (functionId: string, type: string = 'registry') =>
    apiClient.post(`/v1/functions/${functionId}/dna/analyze`, null, {
      params: { type },
    }),

  /** Toggle evolution for a function */
  toggleEvolution: (functionId: string, enabled: boolean, type: string = 'registry') =>
    apiClient.post(`/v1/functions/${functionId}/dna/evolution`, { enabled }, {
      params: { type },
    }),

  /** Get enterprise-wide DNA insights */
  getEnterpriseInsights: (period: string = '30d') =>
    apiClient.get<EnterpriseInsights>('/v1/dna/enterprise/insights', {
      params: { period },
    }),
};
