import { PROVIDERS } from '@/lib/constants';

// Extended provider metadata for enhanced UI
export const PROVIDER_CAPABILITIES: Record<string, string[]> = {
  workers: ['edge', 'serverless', 'wasm', 'global', 'lowLatency', 'freeTier'],
  vercel: ['edge', 'serverless', 'global', 'freeTier'],
  fly: ['containers', 'serverless', 'global', 'lowLatency'],
  deno: ['edge', 'serverless', 'wasm', 'global', 'freeTier'],
  'functionfly-edge': ['edge', 'serverless', 'security', 'global', 'lowLatency'],
};

export type ProviderCapability =
  | 'edge'
  | 'serverless'
  | 'containers'
  | 'freeTier'
  | 'lowLatency'
  | 'security'
  | 'global'
  | 'wasm';

export interface ProviderComparison {
  coldStartMs: number;
  pricingTier: 'Free' | 'Low' | 'Medium' | 'High';
  regionCount: string;
  bestFor: string;
}

export const PROVIDER_COMPARISON: Record<string, ProviderComparison> = {
  workers: {
    coldStartMs: 0,
    pricingTier: 'Free',
    regionCount: '300+',
    bestFor: 'Global edge functions with zero cold starts and WebAssembly support',
  },
  vercel: {
    coldStartMs: 100,
    pricingTier: 'Free',
    regionCount: '100+',
    bestFor: 'Next.js and frontend-focused deployments with edge caching',
  },
  fly: {
    coldStartMs: 500,
    pricingTier: 'Medium',
    regionCount: '35+',
    bestFor: 'Container workloads and persistent apps requiring full VMs',
  },
  deno: {
    coldStartMs: 50,
    pricingTier: 'Free',
    regionCount: '35+',
    bestFor: 'Modern Type-first edge functions with native TypeScript support',
  },
  'functionfly-edge': {
    coldStartMs: 0,
    pricingTier: 'Free',
    regionCount: '3',
    bestFor: 'Quick deployments without external accounts - fully managed',
  },
};

// Region mapping to continents for visual grouping
const REGION_CONTINENT_MAP: Record<string, string> = {
  // Americas
  iad: 'americas',
  iad1: 'americas',
  'us-east4': 'americas',
  'us-east-1': 'americas',
  lax: 'americas',
  'us-west2': 'americas',
  'us-west-2': 'americas',
  sfo1: 'americas',
  ord: 'americas',
  // Europe
  lhr: 'europe',
  lhr1: 'americas',
  fra: 'europe',
  fra1: 'europe',
  'europe-west1': 'europe',
  'eu-central-1': 'europe',
  // Asia Pacific
  sin: 'asia',
  'asia-northeast1': 'asia',
  nrt: 'asia',
  tyo: 'asia',
  hkg: 'asia',
  // Oceania
  syd: 'oceania',
};

export function getRegionContinent(region: string): string {
  return REGION_CONTINENT_MAP[region] || 'americas';
}

export function getProviderRegionGroup(providerId: string, region: string): string {
  // Some providers have unique region codes
  return getRegionContinent(region);
}

// Region metadata for display
export const PROVIDER_REGION_META: Record<
  string,
  { name: string; location: string; continent: string }
> = {
  iad: { name: 'IAD', location: 'Washington, DC', continent: 'americas' },
  iad1: { name: 'IAD1', location: 'Washington, DC', continent: 'americas' },
  lax: { name: 'LAX', location: 'Los Angeles', continent: 'americas' },
  ord: { name: 'ORD', location: 'Chicago', continent: 'americas' },
  sfo1: { name: 'SFO1', location: 'San Francisco', continent: 'americas' },
  'us-east4': { name: 'US East', location: 'Virginia', continent: 'americas' },
  'us-west2': { name: 'US West', location: 'Oregon', continent: 'americas' },
  'us-east-1': { name: 'US East', location: 'Virginia', continent: 'americas' },
  'us-west-2': { name: 'US West', location: 'Oregon', continent: 'americas' },
  lhr: { name: 'LHR', location: 'London', continent: 'europe' },
  lhr1: { name: 'LHR1', location: 'London', continent: 'europe' },
  fra: { name: 'FRA', location: 'Frankfurt', continent: 'europe' },
  fra1: { name: 'FRA1', location: 'Frankfurt', continent: 'europe' },
  'europe-west1': { name: 'EU West', location: 'Belgium', continent: 'europe' },
  'eu-central-1': { name: 'EU Central', location: 'Frankfurt', continent: 'europe' },
  sin: { name: 'SIN', location: 'Singapore', continent: 'asia' },
  syd: { name: 'SYD', location: 'Sydney', continent: 'oceania' },
  hkg: { name: 'HKG', location: 'Hong Kong', continent: 'asia' },
  tyo: { name: 'TYO', location: 'Tokyo', continent: 'asia' },
  nrt: { name: 'NRT', location: 'Tokyo', continent: 'asia' },
  'asia-northeast1': { name: 'Asia NE', location: 'Tokyo', continent: 'asia' },
};

// Provider configuration type (mirrors PROVIDERS structure but as a type)
export interface ProviderConfig {
  id: string;
  name: string;
  color: string;
  icon: string;
  regions: readonly string[];
  description?: string;
  isManaged?: boolean;
}

// Convert PROVIDERS constant to array for easier iteration
export function getAllProviderConfigs(): ProviderConfig[] {
  return Object.values(PROVIDERS);
}

// Find provider config by ID
export function getProviderConfig(id: string): ProviderConfig | undefined {
  return Object.values(PROVIDERS).find((p) => p.id === id);
}
