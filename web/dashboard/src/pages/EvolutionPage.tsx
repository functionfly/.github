import { useState } from "react";
import { useParams } from "react-router-dom";
import { EvolutionDashboard } from "@/components/swarm/EvolutionDashboard";

export function EvolutionPage() {
  const { agentId } = useParams<{ agentId: string }>();
  const [selectedAgentId] = useState(agentId || "default-agent");

  return <EvolutionDashboard agentId={selectedAgentId} />;
}

export default EvolutionPage;
