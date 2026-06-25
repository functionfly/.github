import { apiClient } from './client';

export interface Ambassador {
  metro_id: number;
  metro_slug?: string;
  metro_name?: string;
  country_code?: string;
  state_code?: string;
  city_slug?: string;
  user_id: string;
  name: string;
  profile_number?: number;
  promoted_at: string;
  source: string;
}

export interface AmbassadorsListResponse {
  total: number;
  entries: Ambassador[];
  country?: string;
  privacy_min_active_users: number;
}

export interface AmbassadorResponse {
  metro_slug: string;
  metro_name: string;
  country_code: string;
  ambassador: Ambassador;
}

export const ambassadorsApi = {
  async list(country?: string, limit = 200): Promise<AmbassadorsListResponse> {
    const search = new URLSearchParams();
    if (country) search.set('country', country);
    if (limit) search.set('limit', String(limit));
    const qs = search.toString();
    return apiClient.get<AmbassadorsListResponse>(
      `/v1/city-rankings/ambassadors${qs ? `?${qs}` : ''}`,
    );
  },

  async getForMetro(slug: string): Promise<AmbassadorResponse> {
    return apiClient.get<AmbassadorResponse>(
      `/v1/city-rankings/${encodeURIComponent(slug)}/ambassador`,
    );
  },
};
