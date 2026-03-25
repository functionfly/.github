/**
 * Legacy full-page marketing (not mounted from App: logged-out "/" redirects to web/site Astro).
 * Sections (Footer, icons, etc.) remain imported by other dashboard pages.
 */
import { Navbar } from '@/components/common/Navbar';
import { PublicAnalytics } from '@/components/common/PublicAnalytics';
import { MetaTags } from '@/components/seo/MetaTags';
import { LandingPageStructuredData } from '@/components/seo/StructuredData';
import { useWebVitals } from '@/hooks/useWebVitals';
import { useAuthStore } from '@/stores/authStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  BuiltForAIAgentsSection,
  FAQSection,
  FeaturesSection,
  Footer,
  HeroSection,
  IntegrationsSection,
  InteractiveDemoSection,
  PerformanceMetricsDashboard,
  PricingSection,
  ProcessStepsSection,
  SecuritySection,
  TargetUsersSection,
  TrustMetricsSection,
} from './components';

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
        title="FunctionFly - The Trust Layer for AI Agents"
        description="Verified, signed, auditable tool discovery with trust scores and a zero-knowledge vault for safe AI agent execution."
        keywords={[
          'trust layer',
          'AI agents',
          'verification',
          'attestations',
          'trust scores',
          'zero-knowledge vault',
          'tool calling',
          'auditable',
        ]}
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
