/**
 * FRG (Function Runtime Graph) API Client
 * Handles all FRG-related API calls
 */

import { apiClient } from './client';
import type {
  GraphDefinition,
  GraphInstance,
  ExecutionResult,
  GraphEvent,
  AISuggestion,
  GraphOptimizationSuggestion,
  VersionComparison,
  TestCase,
  GraphNodeRef,
  GraphEdgeDefinition,
} from '@/types/frg';

// API response types
export interface GraphListResponse {
  graphs: GraphDefinition[];
  total?: number;
}

export interface GraphVersionsResponse {
  versions: GraphVersion[];
}

export interface GraphVersion {
  version: string;
  status: 'published' | 'draft' | 'wip';
  changelog?: string;
  author?: string;
  createdAt: string;
}

export interface ExecuteGraphRequest {
  input?: Record<string, unknown>;
}

export interface ExecuteGraphResponse {
  instanceId: string;
  status: string;
  output?: unknown;
  result?: unknown; // Legacy field name from old backend
  error?: string;
  durationMs?: number;
  nodeResults?: Record<string, unknown>;
}

export interface AIComposeRequest {
  prompt: string;
  requirements?: string[];
}

export interface AIComposeResponse {
  success: boolean;
  graph?: GraphDefinition;
  explanation?: Record<string, unknown>;
  confidence: number;
  generationId?: string;
  latencyMs?: number;
  suggestions?: string[];
  error?: string;
}

export interface SemanticSearchResponse {
  graphs: GraphDefinition[];
}

// Helper to convert snake_case from Go to camelCase in responses
function normalizeGraph(def: Record<string, unknown>): GraphDefinition {
  return {
    id: def.id as string,
    author: def.author as string,
    name: def.name as string,
    version: def.version as string,
    fullName: def.full_name as string || `${def.author}/${def.name}@${def.version}`,
    nodeRefs: (def.node_refs ? JSON.parse(def.node_refs as string) : []) as GraphNodeRef[],
    edges: (def.edges ? JSON.parse(def.edges as string) : []) as GraphEdgeDefinition[],
    executionMode: (def.execution_mode as string || 'sync') as GraphDefinition['executionMode'],
    triggerConfig: def.trigger_config ? JSON.parse(def.trigger_config as string) : undefined,
    inputSchema: def.input_schema ? JSON.parse(def.input_schema as string) : undefined,
    outputSchema: def.output_schema ? JSON.parse(def.output_schema as string) : undefined,
    aiDescription: def.ai_description as string | undefined,
    compositionScore: (def.composition_score as number) || 0,
    trustScore: (def.trust_score as number) || 0,
    deterministic: (def.deterministic as boolean) || false,
    forkedFromAuthor: def.forked_from_author as string | undefined,
    forkedFromName: def.forked_from_name as string | undefined,
    forkedFromVersion: def.forked_from_version as string | undefined,
    tenantId: def.tenant_id as string | undefined,
    ownerUserId: def.owner_user_id as string | undefined,
    visibility: (def.visibility as string || 'public') as GraphDefinition['visibility'],
    pricingType: (def.pricing_type as string || 'free') as GraphDefinition['pricingType'],
    basePrice: (def.base_price as number) || 0,
    revenueShare: (def.revenue_share as number) || 80,
    createdAt: def.created_at as string,
    updatedAt: def.updated_at as string,
    publishedAt: def.published_at as string | undefined,
  };
}

export const frgApi = {
  // ==================== Graph Definitions ====================

  /**
   * List all graphs visible to the current user
   */
  listGraphs: async (params?: {
    author?: string;
    visibility?: string;
    executionMode?: string;
  }): Promise<GraphListResponse> => {
    const queryParams = new URLSearchParams();
    if (params?.author) queryParams.set('author', params.author);
    if (params?.visibility) queryParams.set('visibility', params.visibility);
    if (params?.executionMode) queryParams.set('execution_mode', params.executionMode);

    const query = queryParams.toString();
    const url = `/frg/graphs${query ? `?${query}` : ''}`;

    const response = await apiClient.get<Record<string, unknown>[]>(url);
    const graphs = Array.isArray(response)
      ? response.map((g) => normalizeGraph(g))
      : [];

    return { graphs };
  },

  /**
   * Get a specific graph by author/name
   */
  getGraph: async (
    author: string,
    name: string,
    version?: string
  ): Promise<GraphDefinition> => {
    const url = version
      ? `/frg/graphs/${author}/${name}?version=${version}`
      : `/frg/graphs/${author}/${name}`;

    const response = await apiClient.get<Record<string, unknown>>(url);
    return normalizeGraph(response);
  },

  /**
   * Create a new graph
   */
  createGraph: async (data: {
    name: string;
    description?: string;
    executionMode?: string;
    nodes: GraphNodeRef[];
    edges: GraphEdgeDefinition[];
    inputSchema?: Record<string, unknown>;
    outputSchema?: Record<string, unknown>;
    triggerConfig?: Record<string, unknown>;
    visibility?: string;
    basePrice?: number;
  }): Promise<GraphDefinition> => {
    const payload = {
      name: data.name,
      description: data.description,
      execution_mode: data.executionMode || 'sync',
      nodes: data.nodes,
      edges: data.edges,
      input_schema: data.inputSchema,
      output_schema: data.outputSchema,
      trigger_config: data.triggerConfig,
      visibility: data.visibility || 'private',
      base_price: data.basePrice,
    };

    const response = await apiClient.post<Record<string, unknown>>('/frg/graphs', payload);
    return normalizeGraph(response);
  },

  /**
   * Update a graph (only if not published)
   */
  updateGraph: async (
    author: string,
    name: string,
    updates: {
      name?: string;
      nodes?: GraphNodeRef[];
      edges?: GraphEdgeDefinition[];
      inputSchema?: Record<string, unknown>;
      outputSchema?: Record<string, unknown>;
    }
  ): Promise<GraphDefinition> => {
    const payload = {
      name: updates.name,
      nodes: updates.nodes,
      edges: updates.edges,
      input_schema: updates.inputSchema,
      output_schema: updates.outputSchema,
    };

    const response = await apiClient.put<Record<string, unknown>>(
      `/frg/graphs/${author}/${name}`,
      payload
    );
    return normalizeGraph(response);
  },

  /**
   * Delete a graph (only if not published)
   */
  deleteGraph: async (author: string, name: string): Promise<void> => {
    await apiClient.delete(`/frg/graphs/${author}/${name}`);
  },

  /**
   * Publish a graph version
   */
  publishGraph: async (author: string, name: string): Promise<{ status: string }> => {
    return apiClient.post<{ status: string }>(`/frg/graphs/${author}/${name}/publish`, {});
  },

  /**
   * Fork/remix an existing graph
   */
  remixGraph: async (
    author: string,
    name: string,
    newName: string
  ): Promise<GraphDefinition> => {
    const response = await apiClient.post<Record<string, unknown>>(
      `/frg/graphs/${author}/${name}/remix`,
      { new_name: newName }
    );
    return normalizeGraph(response);
  },

  // ==================== Graph Versions ====================

  /**
   * List all versions of a graph
   */
  listVersions: async (author: string, name: string): Promise<GraphVersionsResponse> => {
    // The backend doesn't have a dedicated versions endpoint,
    // so we fetch the graph which includes version info
    try {
      const def = await frgApi.getGraph(author, name);
      return {
        versions: [
          {
            version: def.version,
            status: def.publishedAt ? 'published' : 'draft',
            author: def.author,
            createdAt: def.createdAt,
          },
        ],
      };
    } catch {
      return { versions: [] };
    }
  },

  // ==================== Graph Execution ====================

  /**
   * Execute a graph
   */
  executeGraph: async (
    author: string,
    name: string,
    input?: Record<string, unknown>,
    version?: string
  ): Promise<ExecuteGraphResponse> => {
    const url = version
      ? `/gx/${author}/${name}@${version}`
      : `/gx/${author}/${name}`;

    const payload = { input: input || {} };

    return apiClient.post<ExecuteGraphResponse>(url, payload);
  },

  /**
   * Execute a single function node
   */
  executeNode: async (
    author: string,
    name: string,
    version: string = 'latest',
    input?: Record<string, unknown>
  ): Promise<ExecuteGraphResponse> => {
    let url = `/v1/functions/${author}/${name}/test`;
    if (version && version !== 'latest') {
      url += `?version=${version}`;
    }
    console.log('[executeNode] Calling:', url, 'author:', author, 'name:', name, 'version:', version);
    const payload = { input: input || {} };
    // Handle the simplified test response format from HandleTest
    const response = await apiClient.post(url, payload);
    const rawData = (response as any).data ?? response;
    // Map to ExecuteGraphResponse format
    return {
      instanceId: rawData.instanceId || rawData.function?.name || '',
      status: rawData.status || rawData.message || 'completed',
      output: rawData.output ?? rawData.result ?? rawData,
      error: rawData.error,
      durationMs: rawData.durationMs,
      nodeResults: rawData.nodeResults,
    };
  },

  /**
   * Get instance status
   */
  getInstanceStatus: async (instanceId: string): Promise<GraphInstance> => {
    return apiClient.get<GraphInstance>(`/frg/instances/${instanceId}`);
  },

  /**
   * Stop a running instance
   */
  stopInstance: async (instanceId: string): Promise<{ status: string }> => {
    return apiClient.post<{ status: string }>(`/frg/instances/${instanceId}/stop`, {});
  },

  /**
   * List instances for a graph
   */
  listInstances: async (
    author: string,
    name: string,
    status?: string
  ): Promise<GraphInstance[]> => {
    const url = `/frg/graphs/${author}/${name}/instances${status ? `?status=${status}` : ''}`;
    return apiClient.get<GraphInstance[]>(url);
  },

  // ==================== AI Composition ====================

  /**
   * Generate a graph using AI
   */
  aiCompose: async (data: AIComposeRequest): Promise<AIComposeResponse> => {
    return apiClient.post<AIComposeResponse>('/frg/compose', data);
  },

  /**
   * Semantic search for graphs
   */
  semanticSearch: async (query: string): Promise<SemanticSearchResponse> => {
    const response = await apiClient.get<Record<string, unknown>[]>(
      `/frg/discover?q=${encodeURIComponent(query)}`
    );
    return {
      graphs: Array.isArray(response)
        ? response.map((g) => normalizeGraph(g))
        : [],
    };
  },

  /**
   * Get optimization suggestions for a graph
   */
  getOptimizations: async (
    author: string,
    name: string
  ): Promise<GraphOptimizationSuggestion[]> => {
    return apiClient.get<GraphOptimizationSuggestion[]>(
      `/frg/graphs/${author}/${name}/optimizations`
    );
  },

  /**
   * Generate a function using AI (for adding to graphs)
   */
  generateFunction: async (data: {
    author: string;
    name: string;
    description: string;
    runtime?: string;
  }): Promise<{ success: boolean; functionId?: string; error?: string }> => {
    return apiClient.post<{ success: boolean; functionId?: string; error?: string }>(
      '/frg/functions/generate',
      data
    );
  },
};