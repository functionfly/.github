import { apiClient } from './client';

// ============================================================================
// Types
// ============================================================================

export interface WebSearchResult {
  title: string;
  url: string;
  snippet: string;
  source: string;
  publishedDate?: string;
}

export interface NewsArticle {
  title: string;
  url: string;
  source: string;
  publishedAt: string;
  description: string;
  imageUrl?: string;
}

export interface DocsDocument {
  title: string;
  url: string;
  source: string;
  snippet: string;
  relevanceScore: number;
}

export interface FundingRound {
  round: string;
  amount: string;
  date: string;
}

export interface CompanyInfo {
  name: string;
  domain: string;
  description?: string;
  foundedYear?: number;
  hq?: { city?: string; country?: string };
  size?: string;
  industry?: string;
  funding?: { totalRaised?: string; rounds?: FundingRound[] };
  technologies?: string[];
  news?: NewsArticle[];
}

export interface SearchToolsResult {
  web: WebSearchResult[];
  news: NewsArticle[];
  docs: DocsDocument[];
  company: CompanyInfo;
}

export interface ExecuteSearchRequest {
  toolName: 'search.web' | 'search.news' | 'search.docs' | 'search.company';
  parameters: Record<string, unknown>;
  enableCache?: boolean;
}

export interface ExecuteSearchResponse {
  ok: boolean;
  result: unknown;
  cached: boolean;
  executionTimeMs: number;
  creditsUsed: number;
  resultsCount: number;
}

export interface SearchToolDefinition {
  name: string;
  category: string;
  description: string;
  parameters_schema: Record<string, unknown>;
  cost_per_call: number;
  cost_per_result: number;
  enabled: boolean;
  timeout_ms: number;
}

export interface ListSearchToolsResponse {
  ok: boolean;
  tools: SearchToolDefinition[];
  count: number;
}

export interface SearchStats {
  toolName: string;
  totalCalls: number;
  totalResults: number;
  totalCredits: number;
  avgExecutionTimeMs: number;
  cacheHitRate: number;
  period: string;
}

// ============================================================================
// API
// ============================================================================

export const agentSearchApi = {
  /**
   * Execute a search tool.
   * POST /v1/agent/tools/search/execute
   */
  executeSearchTool: async (req: ExecuteSearchRequest) => {
    const response = await apiClient.post<ExecuteSearchResponse>(
      '/v1/agent/tools/search/execute',
      req
    );
    return response;
  },

  /**
   * List all available search tools.
   * GET /v1/agent/tools/search
   */
  listSearchTools: async () => {
    const response = await apiClient.get<ListSearchToolsResponse>('/v1/agent/tools/search');
    return response;
  },

  /**
   * Get search statistics for a tool.
   * GET /v1/agent/tools/search/stats
   */
  getSearchStats: async (toolName: string, since?: string) => {
    const response = await apiClient.get<{
      ok: boolean;
      stats: SearchStats;
    }>('/v1/agent/tools/search/stats', {
      params: { tool_name: toolName, since },
    });
    return response;
  },
};