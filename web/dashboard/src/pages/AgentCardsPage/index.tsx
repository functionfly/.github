/**
 * AgentCardsPage — Browse A2A agent cards, see trust scores, install.
 * Sourced from GET /v1/a2a/agents/cards.
 */

import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Users, Search, Shield, Globe } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { AgentCardBrowser, PublishCardForm } from './components';

export function AgentCardsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const defaultTab = searchParams.get('tab') || 'browse';
  const [activeTab, setActiveTab] = useState(defaultTab);

  const handleTabChange = (tab: string) => {
    setActiveTab(tab);
    setSearchParams({ tab }, { replace: true });
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-text-primary tracking-tight flex items-center gap-2">
          <Users className="h-6 w-6 text-emerald-500" />
          Agent Cards
        </h1>
        <p className="text-text-secondary mt-1">
          Browse and publish A2A agent cards for the Agent-to-Agent Protocol
        </p>
      </div>

      {/* Tab Navigation */}
      <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
        <div className="flex justify-center">
          <TabsList className="h-12 p-1.5 rounded-xl bg-bg-secondary/80 border border-border-subtle backdrop-blur-sm gap-1">
            <TabsTrigger
              value="browse"
              className="relative px-5 py-2.5 text-sm font-medium rounded-lg transition-all duration-200 data-[state=active]:bg-card data-[state=active]:text-text-primary data-[state=active]:shadow-sm data-[state=active]:border-border-subtle flex items-center gap-2"
            >
              <Globe className="w-4 h-4" />
              Browse
            </TabsTrigger>
            <TabsTrigger
              value="publish"
              className="relative px-5 py-2.5 text-sm font-medium rounded-lg transition-all duration-200 data-[state=active]:bg-card data-[state=active]:text-text-primary data-[state=active]:shadow-sm data-[state=active]:border-border-subtle flex items-center gap-2"
            >
              <Shield className="w-4 h-4" />
              Publish
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="browse" className="mt-6">
          <AgentCardBrowser />
        </TabsContent>
        <TabsContent value="publish" className="mt-6">
          <PublishCardForm />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default AgentCardsPage;
