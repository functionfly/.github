export interface WarMatch {
  id: number;
  war_id: number;
  round: string;
  position: number;
  metro_a_id: number;
  metro_a_slug: string;
  metro_a_name: string;
  metro_a_country: string;
  metro_b_id: number;
  metro_b_slug: string;
  metro_b_name: string;
  metro_b_country: string;
  score_a: number;
  score_b: number;
  active_users_a: number;
  active_users_b: number;
  winner_metro_id?: number;
  decided_at?: string;
}

export interface War {
  id: number;
  slug: string;
  name: string;
  season: string;
  status: 'scheduled' | 'active' | 'complete' | 'cancelled';
  round: string;
  starts_at: string;
  ends_at: string;
  champion_metro_id?: number;
  total_matches: number;
  total_active_users: number;
  created_at: string;
  updated_at: string;
  champion_slug?: string;
  champion_name?: string;
  champion_country?: string;
  quarterfinals?: WarMatch[];
  semifinals?: WarMatch[];
  final?: WarMatch;
}

export interface WarsResponse {
  wars: War[];
}

export interface WarResponse {
  war?: War;
}

export interface WarChampion {
  war_slug: string;
  war_season: string;
  war_ends_at: string;
  metro_id: number;
  metro_slug: string;
  metro_name: string;
  country_code: string;
}

export interface ChampionsResponse {
  champions: WarChampion[];
}
