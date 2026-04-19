import { EvolutionDashboard } from '@/components/swarm/EvolutionDashboard';
import { useState } from 'react';
import { Helmet } from 'react-helmet-async';
import { useParams } from 'react-router-dom';

export function EvolutionPage() {
  const { slug } = useParams<{ slug: string }>();
  const [selectedAgentId] = useState(slug || 'default-agent');

  return (
    <>
      <Helmet>
        <title>Autonomous Operations — FunctionFly</title>
        <meta
          name="description"
          content="Self-evolving backend graph — AI-optimized checkout and payments"
        />
      </Helmet>
      <EvolutionDashboard agentId={selectedAgentId} />
    </>
  );
}

export default EvolutionPage;
