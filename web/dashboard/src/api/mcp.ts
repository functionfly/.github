/**
 * MCP API - Model Context Protocol endpoints
 */

import { apiClient } from './client';
import type { MCPSettings, MCPSettingsGlobal, MCPAnalytics, MCPConnection } from '@/pages/MCPCenterPage/types';

export interface MCPFunctionWithSettings {
  id: string;
  author: string;
  name: string;
  status: string;
  mcp: MCPSettings;
}

export interface MCPFunctionListResponse {
  functions: MCPFunctionWithSettings[];
}

export interface MCPAnalyticsResponse extends MCPAnalytics {}

export interface MCPConnectionsResponse {
  connections: MCPConnection[];
}

export interface MCPConnectionTestResponse {
  success: boolean;
  message: string;
  latency_ms: number;
}

export interface MCPSettingsResponse extends MCPSettingsGlobal {}

export const mcpApi = {
  /**
   * List all functions with MCP settings
   * GET /v1/functions/mcp
   */
  listFunctions: async (): Promise<MCPFunctionListResponse> => {
    const res = await apiClient.get<MCPFunctionListResponse>('/v1/functions/mcp');
    return res ?? { functions: [] };
  },

  /**
   * Get MCP settings for a specific function
   * GET /v1/functions/:id/mcp
   */
  getFunctionSettings: async (functionId: string): Promise<MCPSettings> => {
    return apiClient.get<MCPSettings>(`/v1/functions/${functionId}/mcp`);
  },

  /**
   * Update MCP settings for a function
   * PATCH /v1/functions/:id/mcp
   */
  updateFunctionSettings: async (
    functionId: string,
    settings: Partial<MCPSettings>
  ): Promise<MCPSettings> => {
    return apiClient.patch<MCPSettings>(`/v1/functions/${functionId}/mcp`, settings);
  },

  /**
   * Get MCP analytics
   * GET /v1/mcp/analytics
   */
  getAnalytics: async (params?: { days?: number }): Promise<MCPAnalyticsResponse> => {
    const query = params?.days ? `?days=${params.days}` : '';
    const res = await apiClient.get<MCPAnalyticsResponse>(`/v1/mcp/analytics${query}`);
    return res ?? {
      total_calls: 0,
      unique_clients: 0,
      avg_latency_ms: 0,
      success_rate: 0,
      calls_over_time: [],
      client_breakdown: [],
      top_functions: [],
      transport_usage: [],
    };
  },

  /**
   * Get MCP connections
   * GET /v1/mcp/connections
   */
  getConnections: async (): Promise<MCPConnectionsResponse> => {
    const res = await apiClient.get<MCPConnectionsResponse>('/v1/mcp/connections');
    return res ?? { connections: [] };
  },

  /**
   * Get global MCP settings
   * GET /v1/mcp/settings
   */
  getSettings: async (): Promise<MCPSettingsResponse> => {
    return apiClient.get<MCPSettingsResponse>('/v1/mcp/settings');
  },

  /**
   * Update global MCP settings
   * PATCH /v1/mcp/settings
   */
  updateSettings: async (settings: Partial<MCPSettingsGlobal>): Promise<MCPSettingsResponse> => {
    return apiClient.patch<MCPSettingsResponse>('/v1/mcp/settings', settings);
  },

  /**
   * Toggle a client connection enabled/disabled
   * PATCH /v1/mcp/connections/:clientType
   */
  toggleConnection: async (clientType: string, enabled: boolean): Promise<{ enabled: boolean }> => {
    return apiClient.patch<{ enabled: boolean }>(`/v1/mcp/connections/${clientType}`, { enabled });
  },

  /**
   * Test MCP connection health
   * POST /v1/mcp/connections/:clientType/test
   */
  testConnection: async (clientType: string): Promise<MCPConnectionTestResponse> => {
    return apiClient.post<MCPConnectionTestResponse>(`/v1/mcp/connections/${clientType}/test`);
  },
};