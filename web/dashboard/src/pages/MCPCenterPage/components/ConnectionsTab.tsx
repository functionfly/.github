/**
 * MCP Center - Connections Tab
 * Monitor and manage AI client connections to MCP functions
 */

import { useState, useMemo, useCallback } from 'react';
import { Link2, ExternalLink, Info, RefreshCw, Rocket, Loader2 } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { ClientSetupCard } from './ClientSetupCard';
import { ConnectionDetailsPanel } from './ClientGrid';
import {
  useMCPConnections,
  useRefreshMCPConnections,
  useToggleMCPClient,
  useTestMCPConnection,
} from '../hooks';
import { useAPIKeys } from '@/hooks/useApiKeys';
import { MCP_CLIENT_TYPES } from '../constants';
import type { MCPConnection } from '../types';

const ALL_CLIENT_TYPES = MCP_CLIENT_TYPES.map((c) => c.type);
const PRIMARY_CLIENT_TYPES = MCP_CLIENT_TYPES.filter((c) => c.type !== 'other').map((c) => c.type);

function ensureAllClients(connections: MCPConnection[]): MCPConnection[] {
  const existing = new Map(connections.map((c) => [c.client_type, c]));
  return ALL_CLIENT_TYPES.map((type) => {
    if (existing.has(type)) return existing.get(type)!;
    const meta = MCP_CLIENT_TYPES.find((m) => m.type === type);
    return {
      client_type: type,
      client_icon: meta?.icon,
      status: 'never' as const,
      enabled: true,
      connected_functions: 0,
      total_invocations: 0,
      last_connected_at: null,
      avg_latency_ms: 0,
      connected_function_names: [],
    };
  });
}

type StatusFilter = 'all' | 'active' | 'stale' | 'never';

const STATUS_TABS: { value: StatusFilter; label: string }[] = [
  { value: 'all', label: 'All Clients' },
  { value: 'active', label: 'Active' },
  { value: 'stale', label: 'Stale' },
  { value: 'never', label: 'Not Connected' },
];

export function ConnectionsTab() {
  const [selectedClient, setSelectedClient] = useState<string | null>(null);
  const [activeFilter, setActiveFilter] = useState<StatusFilter>('all');

  const { connections: rawConnections, isLoading } = useMCPConnections();
  const refreshMutation = useRefreshMCPConnections();
  const toggleMutation = useToggleMCPClient();
  const testMutation = useTestMCPConnection();
  const { data: apiKeysData } = useAPIKeys();

  const apiKey = useMemo(() => {
    const keys = apiKeysData?.data ?? [];
    const activeKey = keys.find((k) => k.is_active);
    return activeKey?.key_prefix ? undefined : undefined;
  }, [apiKeysData]);

  const connections = useMemo(
    () => ensureAllClients(rawConnections),
    [rawConnections]
  );

  const selectedConnection = selectedClient
    ? connections.find((c) => c.client_type === selectedClient) || null
    : null;

  const statusCounts = useMemo(() => {
    const counts: Record<StatusFilter, number> = { all: connections.length, active: 0, stale: 0, never: 0 };
    for (const c of connections) {
      if (c.status in counts) counts[c.status as StatusFilter]++;
    }
    return counts;
  }, [connections]);

  const hasAnyActivity = connections.some((c) => c.status === 'active' || c.status === 'stale');

  const filteredConnections = useMemo(() => {
    if (activeFilter === 'all') return connections;
    return connections.filter((c) => c.status === activeFilter);
  }, [connections, activeFilter]);

  const primaryClients = useMemo(
    () => filteredConnections.filter((c) => PRIMARY_CLIENT_TYPES.includes(c.client_type)),
    [filteredConnections]
  );

  const otherClients = useMemo(
    () => filteredConnections.filter((c) => c.client_type === 'other'),
    [filteredConnections]
  );

  const handleFilterChange = useCallback((value: string) => {
    setActiveFilter(value as StatusFilter);
    setSelectedClient(null);
  }, []);

  const handleToggle = useCallback((clientType: string, enabled: boolean) => {
    toggleMutation.mutate({ clientType, enabled });
  }, [toggleMutation]);

  const handleTest = useCallback((clientType: string) => {
    testMutation.mutate(clientType);
  }, [testMutation]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
            <Link2 className="h-5 w-5 text-[var(--status-ok)]" />
            MCP Connections
          </h2>
          <p className="text-sm text-text-secondary mt-1">
            Monitor and manage AI client connections to your MCP functions
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refreshMutation.mutate()}
            disabled={refreshMutation.isPending}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshMutation.isPending ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          <Button variant="outline" size="sm" asChild>
            <a href="https://modelcontextprotocol.io/docs" target="_blank" rel="noopener noreferrer">
              MCP Docs
              <ExternalLink className="h-4 w-4 ml-2" />
            </a>
          </Button>
        </div>
      </div>

      {/* Onboarding Banner — shown when no client has ever connected */}
      {!isLoading && !hasAnyActivity && (
        <Card className="mcp-onboarding-card">
          <CardContent className="py-6">
            <div className="flex items-start gap-4">
              <div className="flex-shrink-0 w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
                <Rocket className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1 space-y-2">
                <h3 className="text-base font-semibold text-text-primary">
                  Connect your first AI client
                </h3>
                <p className="text-sm text-text-secondary leading-relaxed">
                  MCP lets AI tools like VS Code, Cursor, and Claude Desktop call your functions directly.
                  Pick a client below, copy its config, and you&apos;re live in under a minute.
                </p>
                <div className="flex items-center gap-3 pt-1">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <span className="w-5 h-5 rounded-full bg-muted flex items-center justify-center text-[10px] font-bold">1</span>
                    Choose a client
                  </div>
                  <span className="text-muted-foreground">→</span>
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <span className="w-5 h-5 rounded-full bg-muted flex items-center justify-center text-[10px] font-bold">2</span>
                    Copy config
                  </div>
                  <span className="text-muted-foreground">→</span>
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <span className="w-5 h-5 rounded-full bg-muted flex items-center justify-center text-[10px] font-bold">3</span>
                    Restart editor
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Info Banner — shown when user has at least one connection */}
      {!isLoading && hasAnyActivity && (
        <Card className="mcp-info-banner">
          <CardContent className="py-3">
            <div className="flex items-start gap-3">
              <Info className="h-5 w-5 shrink-0 mt-0.5" />
              <div className="text-sm">
                <p className="mcp-info-banner-title">Connect your AI clients</p>
                <p className="mcp-info-banner-text">
                  Enable MCP on any client below to let it call your functions directly.
                  Each card shows connection status and provides the config snippet you need.
                  VS Code uses its built-in MCP support via <code className="bg-muted/50 px-1 rounded text-[11px]">mcp.json</code>.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Client Cards */}
      <Tabs value={activeFilter} onValueChange={handleFilterChange} className="w-full">
        <div className="flex justify-center">
          <TabsList className="mcp-tabs-list">
            {STATUS_TABS.map((tab) => (
              <TabsTrigger key={tab.value} value={tab.value} className="mcp-tab-trigger">
                {tab.label}
                {tab.value !== 'all' && statusCounts[tab.value] > 0 && (
                  <Badge variant="secondary" className="ml-1.5 h-5 min-w-[20px] px-1.5 text-[10px] font-mono">
                    {statusCounts[tab.value]}
                  </Badge>
                )}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>

        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
            {[1, 2, 3, 4, 5].map((i) => (
              <Card key={i} className="mcp-client-card">
                <div className="p-6 space-y-3">
                  <div className="flex items-center gap-3">
                    <Skeleton className="h-10 w-10 rounded" />
                    <div className="space-y-1.5">
                      <Skeleton className="h-4 w-28" />
                      <Skeleton className="h-3 w-40" />
                    </div>
                  </div>
                  <Skeleton className="h-6 w-20" />
                  <div className="flex gap-4">
                    <Skeleton className="h-8 w-16" />
                    <Skeleton className="h-8 w-16" />
                  </div>
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <TabsContent value={activeFilter} className="mt-6">
            {/* Refresh overlay */}
            <div className={`relative ${refreshMutation.isPending ? 'pointer-events-none' : ''}`}>
              {refreshMutation.isPending && (
                <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50 backdrop-blur-sm rounded-lg">
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Refreshing connections...
                  </div>
                </div>
              )}

              {/* Primary clients grid (2x2 or 3-column) */}
              {primaryClients.length > 0 && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {primaryClients.map((conn) => (
                    <ClientSetupCard
                      key={conn.client_type}
                      connection={conn}
                      apiKey={apiKey}
                      isSelected={selectedClient === conn.client_type}
                      isRefreshing={refreshMutation.isPending}
                      onToggle={handleToggle}
                      onClick={setSelectedClient}
                      onTest={handleTest}
                      isTesting={testMutation.isPending && testMutation.variables === conn.client_type}
                      testResult={
                        testMutation.variables === conn.client_type && testMutation.isSuccess
                          ? testMutation.data
                          : null
                      }
                    />
                  ))}
                </div>
              )}

              {/* Other clients — smaller row below */}
              {otherClients.length > 0 && (
                <div className="mt-4">
                  <p className="text-xs font-medium text-muted-foreground mb-3 uppercase tracking-wider">
                    Other clients
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {otherClients.map((conn) => (
                      <ClientSetupCard
                        key={conn.client_type}
                        connection={conn}
                        apiKey={apiKey}
                        isSelected={selectedClient === conn.client_type}
                        isRefreshing={refreshMutation.isPending}
                        onToggle={handleToggle}
                        onClick={setSelectedClient}
                        onTest={handleTest}
                        isTesting={testMutation.isPending && testMutation.variables === conn.client_type}
                        testResult={
                          testMutation.variables === conn.client_type && testMutation.isSuccess
                            ? testMutation.data
                            : null
                        }
                      />
                    ))}
                  </div>
                </div>
              )}

              {/* Empty state for filtered views */}
              {filteredConnections.length === 0 && (
                <div className="text-center py-12 text-muted-foreground text-sm">
                  {activeFilter === 'active' && 'No active connections. Enable a client below to get started.'}
                  {activeFilter === 'stale' && 'No stale connections.'}
                  {activeFilter === 'never' && 'All clients have connected at least once.'}
                </div>
              )}
            </div>
          </TabsContent>
        )}
      </Tabs>

      {/* Selected Client Details */}
      {selectedClient && (
        <div className="mt-6">
          <ConnectionDetailsPanel
            clientType={selectedClient}
            connection={selectedConnection}
            onClose={() => setSelectedClient(null)}
            onTest={handleTest}
            isTesting={testMutation.isPending && testMutation.variables === selectedClient}
          />
        </div>
      )}
    </div>
  );
}
