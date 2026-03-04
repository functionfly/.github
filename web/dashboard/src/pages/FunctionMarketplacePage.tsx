import { FunctionMarketplace } from "@/components/swarm/FunctionMarketplace";
import { Footer } from "@/pages/LandingPage/components/Footer";
import { MetaTags } from '@/components/seo/MetaTags';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';

export function FunctionMarketplacePage() {
  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  return (
    <div className="min-h-screen flex flex-col">
      {/* SEO Meta Tags */}
      <MetaTags
        title="Function Marketplace - Serverless Functions Library | FunctionFly"
        description="Browse and deploy pre-built serverless functions from our extensive marketplace. From API integrations to data processing, find the perfect function for your project."
        keywords={["function marketplace", "serverless functions", "function library", "API integrations", "data processing", "cloud functions"]}
        url={`${window.location.origin}/functions`}
        type="website"
      />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      <div className="flex-1">
        <FunctionMarketplace />
      </div>
      <Footer />
    </div>
  );
}

export default FunctionMarketplacePage;
