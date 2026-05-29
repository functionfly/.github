import type { ModelCatalogItem, ModelSelection, TenantAIPreferences } from '@/api/aiModels';

export function modelKey(provider: string, modelId: string): string {
  return `${provider}:${modelId}`;
}

export function isModelEnabled(
  enabledModels: ModelSelection[],
  provider: string,
  modelId: string
): boolean {
  if (enabledModels.length === 0) return true;
  return enabledModels.some((m) => m.provider === provider && m.model_id === modelId);
}

export function enabledModelCount(
  enabledModels: ModelSelection[],
  catalog: ModelCatalogItem[]
): number {
  if (enabledModels.length === 0) return catalog.length;
  return catalog.filter((m) => isModelEnabled(enabledModels, m.provider, m.id)).length;
}

export function toggleModelAllowlist(
  enabledModels: ModelSelection[],
  catalog: ModelCatalogItem[],
  provider: string,
  modelId: string,
  enabled: boolean
): ModelSelection[] {
  if (enabledModels.length === 0) {
    if (!enabled) {
      return catalog
        .filter((m) => !(m.provider === provider && m.id === modelId))
        .map((m) => ({ provider: m.provider, model_id: m.id }));
    }
    return [];
  }

  if (enabled) {
    if (isModelEnabled(enabledModels, provider, modelId)) return enabledModels;
    return [...enabledModels, { provider, model_id: modelId }];
  }

  return enabledModels.filter((m) => !(m.provider === provider && m.model_id === modelId));
}

export function enableAllModels(catalog: ModelCatalogItem[]): ModelSelection[] {
  return catalog.map((m) => ({ provider: m.provider, model_id: m.id }));
}

export function clearAllowlist(): ModelSelection[] {
  return [];
}

export type CatalogSortField = 'name' | 'provider' | 'tier' | 'cost' | 'enabled' | 'availability';

export type CatalogSortOrder = 'asc' | 'desc';

const TIER_SORT_ORDER: Record<string, number> = {
  frontier: 0,
  reasoning: 1,
  fast: 2,
  code: 3,
  embedding: 4,
  local: 5,
  balanced: 6,
};

const COST_SORT_ORDER: Record<string, number> = {
  free: 0,
  $: 1,
  $$: 2,
  $$$: 3,
  varies: 4,
};

export function sortCatalog(
  items: ModelCatalogItem[],
  field: CatalogSortField,
  order: CatalogSortOrder,
  enabledModels: ModelSelection[]
): ModelCatalogItem[] {
  const dir = order === 'asc' ? 1 : -1;
  const byName = (a: ModelCatalogItem, b: ModelCatalogItem) =>
    a.display_name.localeCompare(b.display_name, undefined, { sensitivity: 'base' });

  return [...items].sort((a, b) => {
    let cmp = 0;
    switch (field) {
      case 'name':
        cmp = byName(a, b);
        break;
      case 'provider':
        cmp =
          a.provider.localeCompare(b.provider, undefined, { sensitivity: 'base' }) || byName(a, b);
        break;
      case 'tier':
        cmp =
          (TIER_SORT_ORDER[a.tier ?? 'balanced'] ?? 99) -
            (TIER_SORT_ORDER[b.tier ?? 'balanced'] ?? 99) || byName(a, b);
        break;
      case 'cost': {
        const costA = COST_SORT_ORDER[a.cost_hint ?? ''] ?? 99;
        const costB = COST_SORT_ORDER[b.cost_hint ?? ''] ?? 99;
        cmp = costA - costB || (a.cost_hint ?? '').localeCompare(b.cost_hint ?? '');
        break;
      }
      case 'enabled': {
        const enA = isModelEnabled(enabledModels, a.provider, a.id) ? 1 : 0;
        const enB = isModelEnabled(enabledModels, b.provider, b.id) ? 1 : 0;
        cmp = enA - enB || byName(a, b);
        break;
      }
      case 'availability': {
        const avA = a.provider_available !== false ? 1 : 0;
        const avB = b.provider_available !== false ? 1 : 0;
        cmp = avA - avB || byName(a, b);
        break;
      }
    }
    return cmp * dir;
  });
}

export function formatModelLabel(selection?: ModelSelection): string {
  if (!selection?.model_id) return 'Not set';
  const shortId = selection.model_id.includes('/')
    ? selection.model_id.split('/').pop()!
    : selection.model_id;
  return selection.provider ? `${selection.provider} · ${shortId}` : shortId;
}

export function preferencesEqual(a: TenantAIPreferences, b: TenantAIPreferences): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export const PROFILE_DEFAULTS: Record<
  'fast' | 'balanced' | 'premium',
  Record<string, ModelSelection>
> = {
  fast: {
    composer: { provider: 'groq', model_id: 'llama-4-scout-17b-16e-instruct' },
    frg: { provider: 'groq', model_id: 'llama-3.3-70b-versatile' },
    agent: { provider: 'groq', model_id: 'llama-4-scout-17b-16e-instruct' },
  },
  balanced: {
    composer: { provider: 'openrouter', model_id: 'anthropic/claude-sonnet-4.6' },
    frg: { provider: 'openrouter', model_id: 'anthropic/claude-sonnet-4.6' },
    agent: { provider: 'openrouter', model_id: 'anthropic/claude-sonnet-4.6' },
  },
  premium: {
    composer: { provider: 'openrouter', model_id: 'anthropic/claude-sonnet-4.6' },
    frg: { provider: 'openrouter', model_id: 'anthropic/claude-sonnet-4.6' },
    dna: { provider: 'openrouter', model_id: 'google/gemini-3.1-pro' },
    chat: { provider: 'openrouter', model_id: 'google/gemini-3.1-pro' },
    support: { provider: 'openrouter', model_id: 'anthropic/claude-haiku-4' },
    embeddings: { provider: 'openai', model_id: 'text-embedding-3-large' },
    agent: { provider: 'openrouter', model_id: 'anthropic/claude-sonnet-4.6' },
  },
};

export function expandProfileDefaults(
  profile: 'fast' | 'balanced' | 'premium'
): Record<string, ModelSelection> {
  return { ...PROFILE_DEFAULTS[profile] };
}

export const PROFILE_OPTIONS = [
  {
    id: 'fast' as const,
    label: 'Fast',
    description: 'Lower latency and cost — great for drafts and iteration.',
  },
  {
    id: 'balanced' as const,
    label: 'Balanced',
    description: 'Recommended default — quality and cost in balance.',
  },
  {
    id: 'premium' as const,
    label: 'Premium',
    description: 'Highest quality models for production codegen.',
  },
];

export const FEATURE_DEFAULTS = [
  { key: 'composer', label: 'AI Composer', capability: 'code' },
  { key: 'frg', label: 'FRG Assistant', capability: 'code' },
] as const;

export const USER_OVERRIDE_FEATURES = [
  { key: 'composer', label: 'AI Composer', capability: 'code' },
  { key: 'frg', label: 'FRG Assistant', capability: 'code' },
] as const;
