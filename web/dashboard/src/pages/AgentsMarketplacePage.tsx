import { AgentMarketplaceView } from '@/components/registry/AgentMarketplaceView';
import { MetaTags } from '@/components/seo/MetaTags';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import '@/styles/aviation-marketplace.css';

export function AgentsMarketplacePage() {
  useWebVitals((metrics) => {
    console.log('Web Vitals:', metrics);
  });

  return (
    <div className="aviation-marketplace min-h-screen flex flex-col">
      <MetaTags
        title="Agent Marketplace | FunctionFly"
        description="Discover and hire AI agents for code generation, analysis, and more. Browse worker, manager, and infrastructure agents."
        keywords={['AI agents', 'agent marketplace', 'hire agents', 'code generation', 'analysis agents']}
        url={`${window.location.origin}/marketplace/agents`}
        type="website"
      />

      <PublicAnalytics />

      <div className="container mx-auto py-8 px-4">
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2">Agent Marketplace</h1>
          <p className="text-text-muted">
            Discover and hire AI agents for your tasks. Browse by capabilities, pricing, and ratings.
          </p>
        </div>

        <AgentMarketplaceView variant="standalone" />
      </div>
    </div>
  );
}

export default AgentsMarketplacePage;