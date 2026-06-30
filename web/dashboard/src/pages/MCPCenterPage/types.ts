/**
 * MCP Center - TypeScript Interfaces
 */

export interface MCPSettings {
  function_id?: string;
  enabled: boolean;
  transports: ('streamable-http' | 'stdio')[];
  expose_input_schema: boolean;
  expose_output_schema: boolean;
  tool_name_override: string;
  rate_limit_per_min: number;
  allowlist_origins: string[];
  verified_mcp?: boolean;
  invocation_count?: number;
  last_invoked_at?: string | null;
}

export interface MCPFunction extends MCPSettings {
  id: string;
  author: string;
  name: string;
  status?: 'deployed' | 'draft' | 'archived';
}

export interface MCPAnalytics {
  total_calls: number;
  unique_clients: number;
  avg_latency_ms: number;
  success_rate: number;
  calls_over_time: { time: string; count: number }[];
  client_breakdown: { client: string; count: number }[];
  top_functions: { author: string; name: string; calls: number }[];
  transport_usage: { transport: string; count: number }[];
}

export interface MCPConnection {
  client_type: string;
  client_icon?: string;
  status: 'active' | 'stale' | 'never';
  enabled: boolean;
  connected_functions: number;
  total_invocations: number;
  last_connected_at: string | null;
  avg_latency_ms: number;
  connected_function_names: string[];
}

export type ClientSetupStatus = 'not_installed' | 'installed' | 'enabled' | 'disabled';

export interface ClientSetupInfo {
  type: string;
  name: string;
  description: string;
  icon: string;
  configPath: string;
  setupCommand: string;
  docsUrl: string;
  deepLink?: string;
  configSnippet: string;
}

export interface MCPSettingsGlobal {
  default_transport: 'streamable-http' | 'stdio' | 'both';
  default_rate_limit: number;
  default_expose_input: boolean;
  default_expose_output: boolean;
  auto_add_to_registry: boolean;
  require_verification: boolean;
  public_listing: boolean;
  cors_allowlist: string[];
  rate_limit_multiplier: number;
}

export type MCPFunctionFilter = 'all' | 'enabled' | 'disabled' | 'verified';
export type MCPFunctionSort = 'name' | 'invocations' | 'lastInvoked';
export type TimeRange = '24h' | '7d' | '30d' | '90d';
export type TransportMode = 'stdio' | 'http';

export interface MCPFunctionRow {
  id: string;
  author: string;
  name: string;
  toolName: string;
  enabled: boolean;
  invocationCount: number;
  lastInvoked: string | null;
  verified: boolean;
}