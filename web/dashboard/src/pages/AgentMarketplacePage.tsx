import { AgentMarketplace } from "@/components/swarm/AgentMarketplace";
import { Footer } from "@/pages/LandingPage/components/Footer";
import { MetaTags } from '@/components/seo/MetaTags';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import { useTranslation } from 'react-i18next';

export function AgentMarketplacePage() {
  const { t } = useTranslation();
  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  return (
    <div className="min-h-screen flex flex-col">
      {/* SEO Meta Tags */}
      <MetaTags
        title={t('agentMarketplace.metaTitle')}
        description={t('agentMarketplace.metaDescription')}
        keywords={t('agentMarketplace.metaKeywords', { returnObjects: true }) as string[]}
        url={`${window.location.origin}/marketplace/agents`}
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
