import { EvolutionDashboard } from '@/components/swarm/EvolutionDashboard';
import { MetaTags } from '@/components/seo/MetaTags';
import { useState } from 'react';
import { useParams } from 'react-router-dom';

export function EvolutionPage() {
  const { slug } = useParams<{ slug: string }>();
  const [selectedAgentId] = useState(slug || 'default-agent');

  return (
    <>
      <MetaTags
        title="Autonomous Operations — FunctionFly"
        description="Self-evolving backend graph — AI-optimized checkout and payments"
      />
      <EvolutionDashboard agentId={selectedAgentId} />
    </>
  );
}

export default EvolutionPage;
