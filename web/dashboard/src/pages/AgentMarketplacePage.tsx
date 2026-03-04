import { AgentMarketplace } from "@/components/swarm/AgentMarketplace";
import { Footer } from "@/pages/LandingPage/components/Footer";
import { MetaTags } from '@/components/seo/MetaTags';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';

export function AgentMarketplacePage() {
  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  return (
    <div className="min-h-screen flex flex-col">
      {/* SEO Meta Tags */}
      <MetaTags
        title="AI Agent Marketplace - Deploy & Monetize AI Agents | FunctionFly"
        description="Discover, deploy, and monetize AI agents on FunctionFly. Browse pre-built agents, create custom solutions, and join the AI agent economy."
        keywords={["AI agents", "agent marketplace", "artificial intelligence", "machine learning", "AI deployment", "agent monetization"]}
        url={`${window.location.origin}/agents`}
        type="website"
      />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      <div className="flex-1">
        <AgentMarketplace />
      </div>
      <Footer />
    </div>
  );
}

export default AgentMarketplacePage;
