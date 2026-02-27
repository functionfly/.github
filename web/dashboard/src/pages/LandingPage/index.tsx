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
        title="FunctionFly - Deploy Functions Anywhere | Serverless Platform"
        description="Deploy serverless functions to any cloud provider with FunctionFly. Zero-config deployments, instant scaling, and unified developer experience. Start building today."
        keywords={["serverless", "functions", "deployment", "cloud", "devops", "serverless platform", "function as a service"]}
      />

      {/* Structured Data */}
      <LandingPageStructuredData />

      <Navbar variant="landing" />
      <HeroSection />
      <TargetUsersSection />
      <TrustMetricsSection />
      <ProcessStepsSection />
      <IntegrationsSection />
      <SecuritySection />
      <FeaturesSection />
      <PricingSection />
      <InteractiveDemoSection />
      <PerformanceMetricsDashboard />
      <FAQSection />
      <Footer />
    </div>
  );
}
