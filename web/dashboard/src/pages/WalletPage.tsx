import { WalletDashboard } from '@/components/swarm/WalletDashboard';
import { useParams } from 'react-router-dom';

export function WalletPage() {
  const { agentId } = useParams<{ agentId: string }>();
  const resolvedAgentId = agentId ?? 'default-agent';

  return <WalletDashboard agentId={resolvedAgentId} />;
}

export default WalletPage;
