/**
 * MCP Center - Unified Dashboard Page
 *
 * A comprehensive dashboard for managing Model Context Protocol (MCP) integration
 * across all functions. Provides registry management, connection monitoring,
 * analytics, and global settings.
 */

import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { usePageTitle } from '@/hooks';
import { Zap, Link2, BarChart3, Settings } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { RegistryTab, ConnectionsTab, AnalyticsTab, SettingsTab } from './components';

export function MCPCenterPage() {
  usePageTitle('MCP Center');
  const [searchParams, setSearchParams] = useSearchParams();
  const defaultTab = searchParams.get('tab') || 'registry';

  const [activeTab, setActiveTab] = useState(defaultTab);

  const handleTabChange = (tab: string) => {
    setActiveTab(tab);
    setSearchParams({ tab }, { replace: true });
  };

  return (
    <div className="mcp-center-container space-y-6">
      <div className="mcp-aviation-grid-bg" />
      {/* Page Header */}
      <div className="mcp-header">
        <h1 className="mcp-header-title">
          <Zap className="h-6 w-6" />
          MCP Center
        </h1>
        <p className="mcp-header-subtitle">
          Manage Model Context Protocol integration across your functions
        </p>
      </div>

      {/* Tab Navigation */}
      <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full mcp-tabs">
        <div className="flex justify-center">
          <TabsList className="mcp-tabs-list">
            <TabsTrigger value="registry" className="mcp-tab-trigger">
              <Zap className="w-4 h-4" />
              Registry
            </TabsTrigger>
            <TabsTrigger value="connections" className="mcp-tab-trigger">
              <Link2 className="w-4 h-4" />
              Connections
            </TabsTrigger>
            <TabsTrigger value="analytics" className="mcp-tab-trigger">
              <BarChart3 className="w-4 h-4" />
              Analytics
            </TabsTrigger>
            <TabsTrigger value="settings" className="mcp-tab-trigger">
              <Settings className="w-4 h-4" />
              Settings
            </TabsTrigger>
          </TabsList>
        </div>

        {/* Registry Tab */}
        <TabsContent value="registry" className="mt-6">
          <RegistryTab />
        </TabsContent>

        {/* Connections Tab */}
        <TabsContent value="connections" className="mt-6">
          <ConnectionsTab />
        </TabsContent>

        {/* Analytics Tab */}
        <TabsContent value="analytics" className="mt-6">
          <AnalyticsTab />
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="mt-6">
          <SettingsTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default MCPCenterPage;
