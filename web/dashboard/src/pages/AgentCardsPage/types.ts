/**
 * AgentCardsPage — TypeScript interfaces for A2A agent cards.
 */

export interface AgentCard {
  id: string;
  name: string;
  description: string;
  url: string;
  version: string;
  protocol_version: string;
  capabilities: string[];
  skills: AgentSkill[];
  auth_schemes: string[];
  input_modes: string[];
  output_modes: string[];
  trust_score: number;
  peer_jwks_url?: string;
  published_at: string;
  updated_at: string;
}

export interface AgentSkill {
  id: string;
  description: string;
}

export interface AgentCardListResponse {
  cards: AgentCard[];
  total: number;
}

export interface PublishCardRequest {
  id: string;
  name: string;
  description: string;
  url: string;
  version?: string;
  protocol_version?: string;
  capabilities?: string[];
  skills?: AgentSkill[];
  auth_schemes?: string[];
  input_modes?: string[];
  output_modes?: string[];
  peer_jwks_url?: string;
}
