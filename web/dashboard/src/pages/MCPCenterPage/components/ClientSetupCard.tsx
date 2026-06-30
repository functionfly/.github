/**
 * MCP Center - Client Setup Card
 * Provides install, enable, and disable actions for MCP clients (VS Code, Claude, Cursor, etc.)
 */

import { useState, useCallback } from 'react';
import {
  Copy,
  Check,
  ExternalLink,
  Power,
  ChevronDown,
  ChevronUp,
  Terminal,
  Settings,
  Wifi,
  WifiOff,
  Loader2,
  Globe,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import type { MCPConnection, TransportMode } from '../types';
import { MCP_CLIENT_SETUP, MCP_CLIENT_TYPES } from '../constants';
import { toast } from 'sonner';

interface ClientSetupCardProps {
  connection: MCPConnection;
  apiKey?: string;
  isSelected?: boolean;
  isRefreshing?: boolean;
  onToggle?: (clientType: string, enabled: boolean) => void;
  onClick?: (clientType: string) => void;
  onTest?: (clientType: string) => void;
  isTesting?: boolean;
  testResult?: { success: boolean; message: string; latency_ms: number } | null;
}

function getClientDisplayName(clientType: string): string {
  const meta = MCP_CLIENT_TYPES.find((c) => c.type === clientType);
  return meta?.name ?? clientType;
}

function getClientDescription(clientType: string): string {
  const meta = MCP_CLIENT_TYPES.find((c) => c.type === clientType);
  if (meta) return meta.description;
  switch (clientType) {
    case 'vscode': return 'Copilot Chat + MCP servers';
    case 'claude-desktop': return 'Anthropic Claude with MCP';
    case 'cursor': return 'AI code editor with MCP';
    case 'windsurf': return 'Codeium Windsurf AI IDE';
    default: return 'MCP-compatible client';
  }
}

export function ClientSetupCard({
  connection,
  apiKey,
  isSelected,
  isRefreshing,
  onToggle,
  onClick,
  onTest,
  isTesting,
  testResult,
}: ClientSetupCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [copied, setCopied] = useState(false);
  const [transportMode, setTransportMode] = useState<TransportMode>('stdio');

  const setup = MCP_CLIENT_SETUP[connection.client_type];
  const hasActivity = connection.status === 'active' || connection.status === 'stale';

  const handleCopyConfig = useCallback(async () => {
    if (!setup) return;
    const config = transportMode === 'http'
      ? setup.getConfigHttp(apiKey || '')
      : setup.getConfig(apiKey || '');
    try {
      await navigator.clipboard.writeText(config);
      setCopied(true);
      toast.success('Config copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  }, [setup, apiKey, transportMode]);

  const handleToggle = useCallback(
    (checked: boolean) => {
      onToggle?.(connection.client_type, checked);
    },
    [connection.client_type, onToggle]
  );

  const handleDeepLink = useCallback(() => {
    if (setup?.deepLink) {
      window.open(setup.deepLink, '_blank');
    }
  }, [setup]);

  const getStatusColor = (status: MCPConnection['status']) => {
    switch (status) {
      case 'active':
        return 'bg-emerald-500';
      case 'stale':
        return 'bg-amber-500';
      case 'never':
        return 'bg-gray-400';
    }
  };

  const getStatusLabel = (status: MCPConnection['status']) => {
    switch (status) {
      case 'active':
        return 'Connected';
      case 'stale':
        return 'Stale';
      case 'never':
        return 'Not connected';
    }
  };

  const formatLastConnected = (date: string | null) => {
    if (!date) return 'Never';
    const d = new Date(date);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (hours < 1) return 'Just now';
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return d.toLocaleDateString();
  };

  const currentConfig = transportMode === 'http'
    ? setup?.getConfigHttp(apiKey || '')
    : setup?.getConfig(apiKey || '');

  return (
    <Card
      className={`mcp-client-card group ${isSelected ? 'selected' : ''} ${isRefreshing ? 'opacity-60 pointer-events-none' : ''}`}
      onClick={() => onClick?.(connection.client_type)}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="text-2xl">{connection.client_icon || '🔌'}</div>
            <div>
              <CardTitle className="mcp-client-card-name text-sm font-semibold">
                {getClientDisplayName(connection.client_type)}
              </CardTitle>
              <CardDescription className="text-xs mt-0.5">
                {getClientDescription(connection.client_type)}
              </CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div onClick={(e) => e.stopPropagation()}>
                    <Switch
                      checked={connection.enabled}
                      onCheckedChange={handleToggle}
                    />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  {connection.enabled ? 'Disable MCP connection' : 'Enable MCP connection'}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-3">
        {/* Status + Stats Row */}
        <div className="flex items-center justify-between">
          <Badge
            variant="secondary"
            className={`mcp-status-indicator ${connection.status === 'active' ? 'active' : connection.status === 'stale' ? 'stale' : 'inactive'}`}
          >
            <span className={`inline-block w-2 h-2 rounded-full mr-1.5 ${getStatusColor(connection.status)}`} />
            {getStatusLabel(connection.status)}
          </Badge>
          <span className="text-xs text-muted-foreground">
            {formatLastConnected(connection.last_connected_at)}
          </span>
        </div>

        {/* Stats */}
        <div className="mcp-client-card-stats">
          <div className="mcp-client-card-stat">
            <span className="mcp-client-card-stat-label">Functions</span>
            <span className="mcp-client-card-stat-value">{connection.connected_functions}</span>
          </div>
          <div className="mcp-client-card-stat">
            <span className="mcp-client-card-stat-label">Calls</span>
            <span className="mcp-client-card-stat-value">
              {connection.total_invocations.toLocaleString()}
            </span>
          </div>
        </div>

        {/* Test Result */}
        {testResult && (
          <div className={`flex items-center gap-2 text-xs p-2 rounded ${testResult.success ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
            {testResult.success ? <Wifi className="h-3.5 w-3.5" /> : <WifiOff className="h-3.5 w-3.5" />}
            <span>{testResult.message}</span>
            {testResult.latency_ms > 0 && (
              <span className="ml-auto font-mono">{testResult.latency_ms}ms</span>
            )}
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-2 pt-1">
          {setup?.deepLink && (
            <Button
              variant="default"
              size="sm"
              className="flex-1"
              onClick={(e) => {
                e.stopPropagation();
                handleDeepLink();
              }}
            >
              <Settings className="h-3.5 w-3.5 mr-1.5" />
              Open Settings
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            className="flex-1"
            onClick={(e) => {
              e.stopPropagation();
              onTest?.(connection.client_type);
            }}
            disabled={isTesting}
          >
            {isTesting ? (
              <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
            ) : (
              <Wifi className="h-3.5 w-3.5 mr-1.5" />
            )}
            Test
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="flex-1"
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
          >
            <Terminal className="h-3.5 w-3.5 mr-1.5" />
            Setup
            {expanded ? (
              <ChevronUp className="h-3.5 w-3.5 ml-1" />
            ) : (
              <ChevronDown className="h-3.5 w-3.5 ml-1" />
            )}
          </Button>
        </div>

        {/* Expanded Setup Section */}
        {expanded && setup && (
          <div className="pt-3 border-t border-border space-y-3" onClick={(e) => e.stopPropagation()}>
            {/* Transport Mode Toggle */}
            <div className="flex items-center gap-2">
              <span className="text-xs font-medium text-text-secondary">Transport:</span>
              <div className="flex rounded-md overflow-hidden border border-border">
                <button
                  className={`px-3 py-1 text-xs font-medium transition-colors ${transportMode === 'stdio' ? 'bg-primary text-primary-foreground' : 'bg-muted/50 text-muted-foreground hover:bg-muted'}`}
                  onClick={() => setTransportMode('stdio')}
                >
                  <Terminal className="h-3 w-3 inline mr-1" />
                  stdio
                </button>
                <button
                  className={`px-3 py-1 text-xs font-medium transition-colors ${transportMode === 'http' ? 'bg-primary text-primary-foreground' : 'bg-muted/50 text-muted-foreground hover:bg-muted'}`}
                  onClick={() => setTransportMode('http')}
                >
                  <Globe className="h-3 w-3 inline mr-1" />
                  HTTP
                </button>
              </div>
            </div>

            <div>
              <p className="text-xs font-medium text-text-secondary mb-1.5">
                Add to {setup.configLabel}:
              </p>
              <div className="relative">
                <pre className="bg-muted/50 rounded-md p-3 text-xs overflow-x-auto font-mono leading-relaxed max-h-40 overflow-y-auto">
                  {currentConfig}
                </pre>
                <Button
                  variant="ghost"
                  size="sm"
                  className="absolute top-1.5 right-1.5 h-7 w-7 p-0"
                  onClick={handleCopyConfig}
                >
                  {copied ? (
                    <Check className="h-3.5 w-3.5 text-emerald-500" />
                  ) : (
                    <Copy className="h-3.5 w-3.5" />
                  )}
                </Button>
              </div>
            </div>

            <div className="text-xs text-muted-foreground space-y-1">
              <p>
                <span className="font-medium">Config path:</span>{' '}
                <code className="bg-muted/50 px-1 py-0.5 rounded text-[11px]">
                  {setup.configPath}
                </code>
              </p>
              {transportMode === 'stdio' ? (
                <p>
                  <span className="font-medium">Requires:</span>{' '}
                  <code className="bg-muted/50 px-1 py-0.5 rounded text-[11px]">
                    npx
                  </code>{' '}
                  (Node.js)
                </p>
              ) : (
                <p>
                  <span className="font-medium">Requires:</span>{' '}
                  <code className="bg-muted/50 px-1 py-0.5 rounded text-[11px]">
                    Streamable HTTP
                  </code>{' '}
                  support in client
                </p>
              )}
            </div>

            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="flex-1"
                onClick={handleCopyConfig}
              >
                {copied ? (
                  <>
                    <Check className="h-3.5 w-3.5 mr-1.5 text-emerald-500" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="h-3.5 w-3.5 mr-1.5" />
                    Copy Config
                  </>
                )}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                asChild
              >
                <a href={setup.docsUrl} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                  Docs
                </a>
              </Button>
            </div>

            {!hasActivity && (
              <div className="bg-muted/30 rounded-md p-2.5 text-xs text-muted-foreground">
                <p className="font-medium text-text-secondary mb-0.5">Getting started</p>
                <ol className="list-decimal list-inside space-y-0.5">
                  <li>Copy the config above</li>
                  <li>Paste into {setup.configLabel}</li>
                  <li>Restart {connection.client_type === 'vscode' ? 'VS Code' : 'your editor'}</li>
                  <li>The connection will appear as Active once a tool is called</li>
                </ol>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
