import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import {
  HeroSection,
  TargetUsersSection,
  TrustMetricsSection,
  ProcessStepsSection,
  IntegrationsSection,
  SecuritySection,
  FeaturesSection,
  BuiltForAIAgentsSection,
  PricingSection,
  InteractiveDemoSection,
  PerformanceMetricsDashboard,
  FAQSection,
  Footer
} from './components';
import { Navbar } from '@/components/common/Navbar';
import { MetaTags } from '@/components/seo/MetaTags';
import { LandingPageStructuredData } from '@/components/seo/StructuredData';
import { useWebVitals } from '@/hooks/useWebVitals';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';

export function LandingPage() {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isOnboardingComplete = useOnboardingStore((state) => state.isOnboardingComplete);
  const hasSkippedOnboarding = useOnboardingStore((state) => state.hasSkippedOnboarding);

  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // Optional: Send to your analytics service
    console.log('Web Vitals:', metrics);
  });

  useEffect(() => {
    // Redirect authenticated users to dashboard or onboarding
    if (isAuthenticated) {
      if (!isOnboardingComplete && !hasSkippedOnboarding) {
        navigate('/onboarding', { replace: true });
      } else {
        navigate('/dashboard', { replace: true });
      }
    }
  }, [isAuthenticated, isOnboardingComplete, hasSkippedOnboarding, navigate]);

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* SEO Meta Tags */}
      <MetaTags
        title="FunctionFly - Serverless Functions & AI Agent Infrastructure"
        description="Deploy serverless functions to edge with multi-cloud failover. Build AI agents with built-in cost controls, state management, and unlimited scaling."
        keywords={["serverless", "functions", "AI agents", "edge computing", "multi-cloud", "deployment", "cloud", "devops"]}
      />

      {/* Structured Data */}
      <LandingPageStructuredData />

      {/* Public Analytics (Hotjar for user behavior) */}
      <PublicAnalytics />

      <Navbar variant="landing" />
      <main>
        <HeroSection />
        <TargetUsersSection />
        <TrustMetricsSection />
        <ProcessStepsSection />
        <IntegrationsSection />
        <SecuritySection />
        <FeaturesSection />
        <BuiltForAIAgentsSection />
        <PricingSection />
        <InteractiveDemoSection />
        <PerformanceMetricsDashboard />
        <FAQSection />
      </main>
      <Footer />
    </div>
  );
}
