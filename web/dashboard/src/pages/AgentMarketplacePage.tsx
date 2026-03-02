import { AgentMarketplace } from "@/components/swarm/AgentMarketplace";
import { Footer } from "@/pages/LandingPage/components/Footer";

export function AgentMarketplacePage() {
  return (
    <div className="min-h-screen flex flex-col">
      <div className="flex-1">
        <AgentMarketplace />
      </div>
      <Footer />
    </div>
  );
}

export default AgentMarketplacePage;
