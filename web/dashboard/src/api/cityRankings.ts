import { apiClient } from './client';

export interface CityRankingEntry {
  rank: number;
  previous_rank: number;
  rank_delta: number;
  metro_slug: string;
  metro_name: string;
  country_code: string;
  population: number;
  score_raw: number;
  score_per_capita: number;
  active_users: number;
  deployments: number;
  executions_30d: number;
  founder_earnings_cents: number;
  new_users_30d: number;
  period_end: string;
}

export interface CityLeaderboardResponse {
  period_end: string;
  total_ranked: number;
  entries: CityRankingEntry[];
  country?: string;
  cache_hit: boolean;
  category?: string;
}

export interface CityMetroDetail {
  current: CityRankingEntry | null;
  history: CityRankingEntry[];
  not_ranked: boolean;
  privacy_min_active_users: number;
  period_end: string;
  cache_hit: boolean;
}

export interface CityMoversResponse {
  direction: 'gainers' | 'losers';
  period_end: string;
  entries: CityRankingEntry[];
}

export interface MyCityResponse {
  has_city: boolean;
  metro?: CityRankingEntry;
  opted_out: boolean;
}

export interface CityOptOutResponse {
  opted_out: boolean;
}

export interface UniversityInMetro {
  university_id: number;
  slug: string;
  name: string;
  short_name?: string;
  country_code: string;
  state_code?: string;
  rank: number;
  score_per_capita: number;
  active_users: number;
  student_count?: number;
}

export interface MetroUniversitiesResponse {
  metro_slug: string;
  entries: UniversityInMetro[];
  total: number;
}

export const cityRankingsApi = {
  async getLeaderboard(params?: {
    country?: string;
    limit?: number;
  }): Promise<CityLeaderboardResponse> {
    const search = new URLSearchParams();
    if (params?.country) search.set('country', params.country);
    if (params?.limit) search.set('limit', String(params.limit));
    const qs = search.toString();
    return apiClient.get<CityLeaderboardResponse>(`/v1/city-rankings${qs ? `?${qs}` : ''}`);
  },

  async getMetro(slug: string): Promise<CityMetroDetail> {
    return apiClient.get<CityMetroDetail>(`/v1/city-rankings/${encodeURIComponent(slug)}`);
  },

  async getMovers(
    direction: 'gainers' | 'losers' = 'gainers',
    limit = 25
  ): Promise<CityMoversResponse> {
    return apiClient.get<CityMoversResponse>(
      `/v1/city-rankings/movers?direction=${direction}&limit=${limit}`
    );
  },

  async getMyCity(): Promise<MyCityResponse> {
    return apiClient.get<MyCityResponse>('/v1/city-rankings/me');
  },

  async getOptOut(): Promise<CityOptOutResponse> {
    return apiClient.get<CityOptOutResponse>('/v1/users/me/city-ranking-opt-out');
  },

  async setOptOut(optedOut: boolean): Promise<CityOptOutResponse> {
    return apiClient.post<CityOptOutResponse>('/v1/users/me/city-ranking-opt-out', {
      opted_out: optedOut,
    });
  },

  async getUniversitiesForMetro(slug: string): Promise<MetroUniversitiesResponse> {
    return apiClient.get<MetroUniversitiesResponse>(
      `/v1/city-rankings/${encodeURIComponent(slug)}/universities`
    );
  },
};
