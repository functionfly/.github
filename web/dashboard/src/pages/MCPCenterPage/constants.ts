/**
 * MCP Center - Constants
 */

import type { TimeRange } from './types';

export const TIME_RANGES: { value: TimeRange; label: string; days: number }[] = [
  { value: '24h', label: '24 hours', days: 1 },
  { value: '7d', label: '7 days', days: 7 },
  { value: '30d', label: '30 days', days: 30 },
  { value: '90d', label: '90 days', days: 90 },
];

export const MCP_FUNCTION_FILTERS = [
  { value: 'all', label: 'All Functions' },
  { value: 'enabled', label: 'Enabled' },
  { value: 'disabled', label: 'Disabled' },
  { value: 'verified', label: 'Verified' },
] as const;

export const MCP_FUNCTION_SORTS = [
  { value: 'name', label: 'Name' },
  { value: 'invocations', label: 'Invocations' },
  { value: 'lastInvoked', label: 'Last Invoked' },
] as const;

export const DEFAULT_MCP_SETTINGS: {
  default_transport: 'streamable-http' | 'stdio' | 'both';
  default_rate_limit: number;
  default_expose_input: boolean;
  default_expose_output: boolean;
  auto_add_to_registry: boolean;
  require_verification: boolean;
  public_listing: boolean;
  cors_allowlist: string[];
  rate_limit_multiplier: number;
} = {
  default_transport: 'streamable-http',
  default_rate_limit: 60,
  default_expose_input: true,
  default_expose_output: false,
  auto_add_to_registry: false,
  require_verification: false,
  public_listing: false,
  cors_allowlist: [],
  rate_limit_multiplier: 1,
};

export const MCP_CLIENT_TYPES = [
  {
    type: 'claude-desktop',
    name: 'Claude Desktop',
    description: 'Anthropic Claude with MCP support',
    icon: '🤖',
  },
  {
    type: 'cursor',
    name: 'Cursor',
    description: 'AI-powered code editor with MCP',
    icon: '📐',
  },
  {
    type: 'vscode',
    name: 'VS Code',
    description: 'Microsoft VS Code with MCP extension',
    icon: '💻',
  },
  {
    type: 'windsurf',
    name: 'Windsurf',
    description: 'Codeium Windsurf AI IDE',
    icon: '🌊',
  },
  {
    type: 'other',
    name: 'Other Clients',
    description: 'Other MCP-compatible AI tools',
    icon: '🔌',
  },
] as const;