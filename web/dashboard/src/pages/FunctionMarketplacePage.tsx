import { BrowseFunctionsView } from '@/components/registry/BrowseFunctionsView';
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
    <div className="aviation-marketplace min-h-screen flex flex-col">
      {/* SEO Meta Tags */}
      <MetaTags
        title="Discover Functions | FunctionFly"
        description="Discover and deploy serverless functions. Browse the registry, deploy instantly, or try live in the playground."
        keywords={["function registry", "serverless functions", "discover functions", "deploy functions"]}
        url={`${window.location.origin}/dashboard`}
        type="website"
      />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      <BrowseFunctionsView variant="dashboard" />
    </div>
  );
}

export default FunctionMarketplacePage;
