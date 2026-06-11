/**
 * MCP Center - Connections Tab
 * Monitor AI client connections to MCP functions
 */

import { useState } from 'react';
import { Link2, ExternalLink, Info } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ClientGrid, ConnectionDetailsPanel } from '../components';
import type { MCPConnection } from '../types';

export function ConnectionsTab() {
  const [selectedClient, setSelectedClient] = useState<string | null>(null);

  // Mock connections for demonstration
  const mockConnections: MCPConnection[] = [
    {
      client_type: 'claude-desktop',
      status: 'active',
      connected_functions: 5,
      total_invocations: 1250,
      last_connected_at: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
    },
    {
      client_type: 'cursor',
      status: 'active',
      connected_functions: 3,
      total_invocations: 890,
      last_connected_at: new Date(Date.now() - 7200000).toISOString(), // 2 hours ago
    },
    {
      client_type: 'vscode',
      status: 'stale',
      connected_functions: 2,
      total_invocations: 156,
      last_connected_at: new Date(Date.now() - 86400000 * 3).toISOString(), // 3 days ago
    },
    {
      client_type: 'windsurf',
      status: 'never',
      connected_functions: 0,
      total_invocations: 0,
      last_connected_at: null,
    },
    {
      client_type: 'other',
      status: 'stale',
      connected_functions: 1,
      total_invocations: 42,
      last_connected_at: new Date(Date.now() - 86400000 * 5).toISOString(), // 5 days ago
    },
  ];

  const selectedConnection = selectedClient
    ? mockConnections.find((c) => c.client_type === selectedClient) || null
    : null;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-text-primary flex items-center gap-2">
            <Link2 className="h-5 w-5 text-brand-500" />
            MCP Connections
          </h2>
          <p className="text-sm text-text-secondary mt-1">
            Monitor which AI clients are connecting to your MCP functions
          </p>
        </div>
        <Button variant="outline" asChild>
          <a href="https://modelcontextprotocol.io/docs" target="_blank" rel="noopener noreferrer">
            MCP Documentation
            <ExternalLink className="h-4 w-4 ml-2" />
          </a>
        </Button>
      </div>

      {/* Info Banner */}
      <Card className="mcp-info-banner">
        <CardContent className="py-3">
          <div className="flex items-start gap-3">
            <Info className="h-5 w-5" />
            <div className="text-sm">
              <p className="mcp-info-banner-title">What are MCP connections?</p>
              <p className="mcp-info-banner-text">
                When an AI client (like Claude Desktop or Cursor) connects to your MCP functions, it
                creates a connection that allows the AI to call your functions directly. Monitor
                these connections to understand which clients are using your functions.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Client Grid */}
      <Tabs defaultValue="all" className="w-full">
        <div className="flex justify-center">
          <TabsList className="mcp-tabs-list">
            <TabsTrigger value="all" className="mcp-tab-trigger">
              All Clients
            </TabsTrigger>
            <TabsTrigger value="active" className="mcp-tab-trigger">
              Active
            </TabsTrigger>
            <TabsTrigger value="stale" className="mcp-tab-trigger">
              Stale
            </TabsTrigger>
            <TabsTrigger value="never" className="mcp-tab-trigger">
              Never Connected
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="all" className="mt-6">
          <ClientGrid
            connections={mockConnections}
            isLoading={false}
            onClientClick={setSelectedClient}
          />
        </TabsContent>

        <TabsContent value="active" className="mt-6">
          <ClientGrid
            connections={mockConnections.filter((c) => c.status === 'active')}
            isLoading={false}
            onClientClick={setSelectedClient}
          />
        </TabsContent>

        <TabsContent value="stale" className="mt-6">
          <ClientGrid
            connections={mockConnections.filter((c) => c.status === 'stale')}
            isLoading={false}
            onClientClick={setSelectedClient}
          />
        </TabsContent>

        <TabsContent value="never" className="mt-6">
          <ClientGrid
            connections={mockConnections.filter((c) => c.status === 'never')}
            isLoading={false}
            onClientClick={setSelectedClient}
          />
        </TabsContent>
      </Tabs>

      {/* Selected Client Details */}
      {selectedClient && (
        <div className="mt-6">
          <ConnectionDetailsPanel
            clientType={selectedClient}
            connection={selectedConnection}
            onClose={() => setSelectedClient(null)}
          />
        </div>
      )}
    </div>
  );
}
