export type MarketplaceItemType = 'agent' | 'extension' | 'function';

export interface UnifiedItem {
  type: MarketplaceItemType;
  id: string;
  name: string;
  description: string;
  icon_url: string | null;
  rating: number;
  install_count: number;
  price: string;
  pricing_model: string;
  tags: string[];
  verified: boolean;
  metadata: Record<string, unknown>;
}

export interface UnifiedSearchResponse {
  items: UnifiedItem[];
  total: number;
  has_more: boolean;
}

export interface MarketplaceSearchParams {
  q?: string;
  type?: MarketplaceItemType | '';
  limit?: number;
  offset?: number;
}

export const ITEM_TYPE_CONFIG: Record<MarketplaceItemType, { label: string; color: string; icon: string }> = {
  agent: { label: 'Agent', color: 'var(--status-ok)', icon: 'Bot' },
  extension: { label: 'Extension', color: '#a78bfa', icon: 'Puzzle' },
  function: { label: 'Function', color: 'var(--foil-a)', icon: 'Code' },
};
