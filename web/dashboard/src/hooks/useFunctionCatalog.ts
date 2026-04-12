import { useQuery } from '@tanstack/react-query';
import { registryApi, type RegistryFunction } from '@/api/registry';
import type { FunctionCatalogItem } from '@/types/frg';

// Query keys for function catalog
export const catalogKeys = {
  all: ['function-catalog'] as const,
  lists: () => [...catalogKeys.all, 'list'] as const,
  list: (filters: CatalogFilters) => [...catalogKeys.lists(), { filters }] as const,
  search: (query: string) => [...catalogKeys.all, 'search', query] as const,
  byCategory: (category: string) => [...catalogKeys.all, 'category', category] as const,
};

export interface CatalogFilters {
  category?: string;
  search?: string;
  author?: string;
  limit?: number;
  offset?: number;
}

// Transform RegistryFunction to FunctionCatalogItem
// This bridges the API response to the UI component's expected type
function transformRegistryFunction(fn: RegistryFunction): FunctionCatalogItem {
  // Map category colors for UI consistency
  const categoryColors: Record<string, string> = {
    api: '#6366f1',      // Indigo
    data: '#10b981',     // Emerald
    text: '#3b82f6',     // Blue
    image: '#8b5cf6',    // Violet
    video: '#ef4444',    // Red
    audio: '#f59e0b',    // Amber
    code: '#14b8a6',     // Teal
    ml: '#ec4899',       // Pink
    default: '#6b7280',  // Gray
  };

  // Map category to icon type
  const categoryIcons: Record<string, string> = {
    api: 'globe',
    data: 'database',
    text: 'file-text',
    image: 'image',
    video: 'video',
    audio: 'music',
    code: 'code',
    ml: 'cpu',
    default: 'code',
  };

  const category = fn.category || 'default';

  return {
    id: fn.id,
    author: fn.author,
    name: fn.name,
    version: fn.latest_version || '1.0.0',
    description: fn.description || `${fn.name} - ${fn.author}`,
    category: category,
    tags: fn.tags || [],
    // Schema info not available in list view - will be fetched when needed
    inputSchema: { type: 'object', properties: {} },
    outputSchema: { type: 'object', properties: {} },
    trustScore: fn.trust_score ?? fn.overall_score ?? 4.0,
    usageCount: fn.popularity_score ? Math.round(fn.popularity_score * 1000) : 0,
    avgExecutionTimeMs: 0, // Not available in registry list response
    icon: categoryIcons[category] || categoryIcons.default,
    color: categoryColors[category] || categoryColors.default,
  };
}

/**
 * Hook to fetch function catalog with filters
 * Fetches from the registry API and transforms to FunctionCatalogItem format
 */
export function useFunctionCatalog(filters?: CatalogFilters) {
  return useQuery({
    queryKey: catalogKeys.list(filters || {}),
    queryFn: async () => {
      const response = await registryApi.getFunctions({
        category: filters?.category,
        query: filters?.search,
        author: filters?.author,
        limit: filters?.limit ?? 50,
        offset: filters?.offset ?? 0,
      });

      // Transform API response to UI format
      const functions = response.functions || [];
      return functions.map(transformRegistryFunction);
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 10, // 10 minutes
    enabled: filters !== undefined,
  });
}

/**
 * Hook to search functions by query string
 */
export function useFunctionSearch(query: string, category?: string) {
  return useQuery({
    queryKey: catalogKeys.search(query),
    queryFn: async () => {
      if (!query.trim()) {
        return [];
      }
      const response = await registryApi.searchFunctions(query, category, 50);
      const functions = response.functions || [];
      return functions.map(transformRegistryFunction);
    },
    enabled: query.trim().length > 0,
    staleTime: 1000 * 60 * 2, // 2 minutes
  });
}

/**
 * Hook to get a single function's details with full schemas
 */
export function useFunctionDetail(author: string, name: string) {
  return useQuery({
    queryKey: ['function-detail', author, name],
    queryFn: async () => {
      const response = await registryApi.getFunction(author, name);
      const fn = response.function;

      // Get the latest version for schema info
      const latestVersion = response.versions?.[0];

      const baseTransform = transformRegistryFunction(fn);

      return {
        ...baseTransform,
        // Override with version-specific details if available
        inputSchema: latestVersion?.manifest?.input_schema || baseTransform.inputSchema,
        outputSchema: latestVersion?.manifest?.output_schema || baseTransform.outputSchema,
        version: latestVersion?.version || baseTransform.version,
      } as FunctionCatalogItem;
    },
    enabled: !!author && !!name,
    staleTime: 1000 * 60 * 5,
  });
}

/**
 * Hook to get recently used functions (for "Recently Used" tab)
 * Currently uses popular functions as a proxy
 */
export function useRecentFunctions(limit = 10) {
  return useQuery({
    queryKey: catalogKeys.lists(),
    queryFn: async () => {
      const response = await registryApi.getFunctions({
        limit,
        offset: 0,
      });
      // Sort by popularity as a proxy for "recent" until we have usage history
      const functions = response.functions || [];
      return functions
        .sort((a, b) => (b.popularity_score ?? 0) - (a.popularity_score ?? 0))
        .slice(0, limit)
        .map(transformRegistryFunction);
    },
    staleTime: 1000 * 60 * 5,
  });
}
