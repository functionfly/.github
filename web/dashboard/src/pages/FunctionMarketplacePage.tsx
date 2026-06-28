import { BrowseFunctionsView } from '@/components/registry/BrowseFunctionsView';
import { MetaTags } from '@/components/seo/MetaTags';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import { usePageTitle } from '@/hooks';
import { PageGrid } from '@/components/containment';

export function FunctionMarketplacePage() {
  usePageTitle('Discover Functions');

  useWebVitals((metrics) => {
    console.log('Web Vitals:', metrics);
  });

  return (
    <div className="min-h-screen flex flex-col">
      <PageGrid />
      <MetaTags
        title="Discover Functions | FunctionFly"
        description="Discover and deploy serverless functions. Browse the registry, deploy instantly, or try live in the playground."
        keywords={["function registry", "serverless functions", "discover functions", "deploy functions"]}
        url={`${window.location.origin}/dashboard`}
        type="website"
      />
      <PublicAnalytics />
      <BrowseFunctionsView variant="dashboard" />
    </div>
  );
}

export default FunctionMarketplacePage;
