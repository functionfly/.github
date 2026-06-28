import { AgentMarketplaceView } from '@/components/registry/AgentMarketplaceView';
import { MetaTags } from '@/components/seo/MetaTags';
import { PageGrid } from '@/components/containment';

export function AgentsMarketplacePage() {
  return (
    <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <PageGrid />
      <MetaTags
        title="Agent Marketplace | FunctionFly"
        description="Discover and hire AI agents for code generation, analysis, and more."
        keywords={['AI agents', 'agent marketplace', 'hire agents', 'code generation']}
        url={`${window.location.origin}/marketplace/agents`}
        type="website"
      />

      <div>
        <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, letterSpacing: '-0.005em', color: 'var(--text)' }}>Agent Marketplace</h1>
        <p style={{ fontSize: 14, color: 'var(--text-dim)', marginTop: 'var(--space-2)' }}>
          Discover and hire AI agents for your tasks. Browse by capabilities, pricing, and ratings.
        </p>
      </div>

      <AgentMarketplaceView variant="standalone" />
    </div>
  );
}

export default AgentsMarketplacePage;
