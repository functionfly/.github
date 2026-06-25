import { apiClient } from './client';

export interface UniversityRanking {
  rank: number;
  previous_rank?: number | null;
  rank_delta: number;
  university_id: number;
  slug: string;
  name: string;
  short_name?: string;
  country_code: string;
  state_code?: string;
  city_slug?: string;
  student_count?: number;
  score_raw: number;
  score_per_capita: number;
  active_users: number;
  deployments: number;
  executions_30d: number;
  new_users_30d: number;
  period_end: string;
}

export interface UniversityLeaderboardResponse {
  period_end: string;
  total_ranked: number;
  entries: UniversityRanking[];
  country?: string;
  category: string;
  cache_hit: boolean;
  privacy_min_active_users: number;
}

export interface UniversityDetail {
  slug: string;
  entry: UniversityRanking;
  cache_hit: boolean;
}

export interface UniversityMyResponse {
  source: string;
  university?: {
    id: number;
    slug: string;
    name: string;
    short_name?: string;
    country_code: string;
    state_code?: string;
    city_id?: number;
    student_count: number;
    institution_type: string;
    website?: string;
  };
  ranking?: UniversityRanking;
}

export const universityRankingsApi = {
  async getLeaderboard(params?: {
    country?: string;
    limit?: number;
    category?: string;
  }): Promise<UniversityLeaderboardResponse> {
    const search = new URLSearchParams();
    if (params?.country) search.set('country', params.country);
    if (params?.limit) search.set('limit', String(params.limit));
    if (params?.category) search.set('category', params.category);
    const qs = search.toString();
    return apiClient.get<UniversityLeaderboardResponse>(
      `/v1/university-rankings${qs ? `?${qs}` : ''}`,
    );
  },

  async getDetail(slug: string): Promise<UniversityDetail> {
    return apiClient.get<UniversityDetail>(
      `/v1/university-rankings/${encodeURIComponent(slug)}`,
    );
  },

  async getMy(): Promise<UniversityMyResponse> {
    return apiClient.get<UniversityMyResponse>('/v1/university-rankings/me');
  },

  async resolve(input: string, slug?: string): Promise<{
    university: { id: number; slug: string; name: string };
    source: string;
    ambiguous: boolean;
  }> {
    return apiClient.post('/v1/university-rankings/resolve', { input, slug });
  },

  async setOptOut(optedOut: boolean): Promise<{ opted_out: boolean }> {
    return apiClient.post('/v1/users/me/university-ranking-opt-out', {
      opted_out: optedOut,
    });
  },

  async getOptOut(): Promise<{ opted_out: boolean }> {
    return apiClient.get<{ opted_out: boolean }>(
      '/v1/users/me/university-ranking-opt-out',
    );
  },
};
