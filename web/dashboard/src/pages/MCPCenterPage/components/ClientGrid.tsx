/**
 * MCP Center - Client Grid Component
 * Displays AI client connection cards
 */

import { Link } from 'react-router-dom';
import { Loader2, Clock, Activity, ChevronRight } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import type { MCPConnection } from '../types';
import { MCP_CLIENT_TYPES } from '../constants';

interface ClientGridProps {
  connections: MCPConnection[];
  isLoading: boolean;
  onClientClick?: (clientType: string) => void;
}

// Mock connections data generator
const generateMockConnections = (): MCPConnection[] => {
  return MCP_CLIENT_TYPES.map((client) => {
    const hasActivity = Math.random() > 0.3;
    const lastConnected = hasActivity
      ? new Date(Date.now() - Math.random() * 86400000 * 7).toISOString()
      : null;

    return {
      client_type: client.type,
      status: hasActivity
        ? lastConnected && new Date(lastConnected).getTime() > Date.now() - 86400000
          ? 'active'
          : 'stale'
        : 'never',
      enabled: true,
      connected_functions: Math.floor(Math.random() * 10) + 1,
      total_invocations: Math.floor(Math.random() * 2000) + 50,
      last_connected_at: lastConnected,
      avg_latency_ms: Math.floor(Math.random() * 100) + 10,
      connected_function_names: [],
    };
  });
};

export function ClientGrid({ connections, isLoading, onClientClick }: ClientGridProps) {
  // Use mock data if connections is empty
  const displayConnections = connections.length > 0 ? connections : generateMockConnections();

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
        return 'Active';
      case 'stale':
        return 'Stale';
      case 'never':
        return 'Never';
    }
  };

  const formatLastConnected = (date: string | null) => {
    if (!date) return 'No connections yet';
    const d = new Date(date);
    const now = new Date();
    const diff = now.getTime() - d.getTime();
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (hours < 1) return 'Less than an hour ago';
    if (hours < 24) return `${hours} hours ago`;
    if (days < 7) return `${days} days ago`;
    return d.toLocaleDateString();
  };

  const getClientInfo = (clientType: string) => {
    return (
      MCP_CLIENT_TYPES.find((c) => c.type === clientType) || {
        name: clientType,
        description: 'MCP-compatible client',
        icon: '🔌',
      }
    );
  };

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5].map((i) => (
          <Card key={i} className="mcp-client-card">
            <CardHeader>
              <div className="h-6 w-32 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-4 w-24 bg-muted animate-pulse rounded mb-2" />
              <div className="h-4 w-40 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {displayConnections.map((connection) => {
        const clientInfo = getClientInfo(connection.client_type);

        return (
          <Card
            key={connection.client_type}
            className="mcp-client-card"
            onClick={() => onClientClick?.(connection.client_type)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="text-2xl">{clientInfo.icon}</div>
                  <div>
                    <CardTitle className="mcp-client-card-name">{clientInfo.name}</CardTitle>
                    <CardDescription className="text-xs mt-0.5">
                      {clientInfo.description}
                    </CardDescription>
                  </div>
                </div>
                <Badge
                  variant="secondary"
                  className={`mcp-status-indicator ${connection.status === 'active' ? 'active' : connection.status === 'stale' ? 'stale' : 'inactive'}`}
                >
                  {getStatusLabel(connection.status)}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="mcp-client-card-stats">
                <div className="mcp-client-card-stat">
                  <span className="mcp-client-card-stat-label">Functions</span>
                  <span className="mcp-client-card-stat-value">
                    {connection.connected_functions}
                  </span>
                </div>
                <div className="mcp-client-card-stat">
                  <span className="mcp-client-card-stat-label">Invocations</span>
                  <span className="mcp-client-card-stat-value">
                    {connection.total_invocations.toLocaleString()}
                  </span>
                </div>
              </div>

              <div className="pt-2 border-t border-border">
                <p className="text-xs text-muted-foreground">
                  {formatLastConnected(connection.last_connected_at)}
                </p>
              </div>

              <Button variant="ghost" size="sm" className="w-full" asChild>
                <Link to={`/mcp?tab=connections&client=${connection.client_type}`}>
                  View Details <ChevronRight className="h-4 w-4 ml-1" />
                </Link>
              </Button>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

// Connection details panel when a client is selected
interface ConnectionDetailsPanelProps {
  clientType: string;
  connection: MCPConnection | null;
  onClose: () => void;
  onTest?: (clientType: string) => void;
  isTesting?: boolean;
}

export function ConnectionDetailsPanel({
  clientType,
  connection,
  onClose,
  onTest,
  isTesting,
}: ConnectionDetailsPanelProps) {
  const clientInfo = getClientInfo(clientType);

  if (!connection) {
    return (
      <Card className="mcp-connection-panel">
        <CardHeader>
          <CardTitle>{clientInfo.name}</CardTitle>
          <CardDescription>No connection data available</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            This client has not connected to any of your MCP functions yet.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="mcp-connection-panel">
      <CardHeader className="mcp-connection-panel-header">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="text-2xl">{clientInfo.icon}</div>
            <div>
              <CardTitle className="mcp-connection-panel-title">{clientInfo.name}</CardTitle>
              <CardDescription>{clientInfo.description}</CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {onTest && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onTest(clientType)}
                disabled={isTesting}
              >
                {isTesting ? (
                  <Loader2 className="h-3.5 w-3.5 mr-1.5 animate-spin" />
                ) : (
                  <Activity className="h-3.5 w-3.5 mr-1.5" />
                )}
                Test Connection
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="mcp-connection-panel-close"
            >
              Close
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <div className="mcp-instrument">
            <p className="text-sm text-muted-foreground">Connected Functions</p>
            <p className="mcp-metric-card-value">{connection.connected_functions}</p>
          </div>
          <div className="mcp-instrument">
            <p className="text-sm text-muted-foreground">Total Invocations</p>
            <p className="mcp-metric-card-value">{connection.total_invocations.toLocaleString()}</p>
          </div>
          <div className="mcp-instrument">
            <p className="text-sm text-muted-foreground">Avg Latency</p>
            <p className="mcp-metric-card-value">
              {connection.avg_latency_ms > 0 ? `${connection.avg_latency_ms}ms` : '—'}
            </p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="mcp-instrument">
            <p className="text-sm text-muted-foreground">Last Connected</p>
            <p className="text-sm font-medium mt-1">
              {connection.last_connected_at
                ? new Date(connection.last_connected_at).toLocaleString()
                : 'Never'}
            </p>
          </div>
          <div className="mcp-instrument">
            <p className="text-sm text-muted-foreground">Connection Status</p>
            <div className="mt-1">
              <Badge
                variant="secondary"
                className={`mcp-status-indicator ${connection.status === 'active' ? 'active' : connection.status === 'stale' ? 'stale' : 'inactive'}`}
              >
                {connection.status === 'active'
                  ? 'Currently active'
                  : connection.status === 'stale'
                    ? 'Connection stale'
                    : 'Never connected'}
              </Badge>
            </div>
          </div>
        </div>

        {connection.connected_function_names.length > 0 && (
          <div className="mcp-instrument">
            <p className="text-sm text-muted-foreground mb-2">Connected Functions</p>
            <div className="flex flex-wrap gap-1.5">
              {connection.connected_function_names.map((name) => (
                <Badge key={name} variant="outline" className="text-xs font-mono">
                  {name}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function getClientInfo(clientType: string) {
  return (
    MCP_CLIENT_TYPES.find((c) => c.type === clientType) || {
      name: clientType,
      description: 'MCP-compatible client',
      icon: '🔌',
    }
  );
}
