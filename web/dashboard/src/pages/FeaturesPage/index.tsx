import { useState } from "react";
import { Footer } from "@/pages/LandingPage/components";
import { CompetitorComparison } from "./components/CompetitorComparison";
import { FeatureComparison } from "./components/FeatureComparison";
import { FeatureFilter } from "./components/FeatureFilter";
import { InteractiveFeatureDemos } from "./components/InteractiveFeatureDemos";
import { MetaTags } from "@/components/seo/MetaTags";
import {
  Navigation,
  PageHeader,
  CTASection,
  StructuredData,
  FeatureGrid,
  CategorySection
} from "./components";
import { features, categories } from "./data/features";
import { categoryColors } from "./types";

// Helper functions for meta descriptions and keywords
const getCategoryMetaDescription = (category?: string | null) => {
  const baseDescription = "Discover powerful features for modern developers: multi-provider deployment, intelligent failover, predictive routing, and advanced analytics.";

  if (!category) return baseDescription;

  const categoryDescriptions = {
    "Reliability": "Ensure your applications stay online with fast failover, intelligent routing, and zero-downtime deployments across multiple providers.",
    "Deployment": "Deploy serverless functions to multiple edge providers simultaneously with unified APIs, zero vendor lock-in, and global distribution.",
    "Intelligence": "AI-powered traffic routing and predictive analytics that prevent issues before they impact users, with machine learning algorithms.",
    "Developer Tools": "Seamless deployment experience with Git integration, CLI tools, comprehensive APIs, and TypeScript/JavaScript support.",
    "Collaboration": "Team collaboration features with role-based permissions, member invites, audit logs, and secure access control.",
    "Monitoring": "Comprehensive analytics with custom dashboards, real-time metrics, performance insights, and detailed deployment monitoring.",
    "Configuration": "Flexible configuration options including environments, custom domains, environment variables, and granular deployment settings.",
    "Infrastructure": "Global edge network with 200+ locations worldwide for optimal performance, reduced latency, and enterprise-grade reliability.",
  };

  return categoryDescriptions[category as keyof typeof categoryDescriptions] || baseDescription;
};

const getCategoryKeywords = (category?: string | null) => {
  const baseKeywords = [
    "serverless features",
    "multi-provider deployment",
    "intelligent failover",
    "predictive routing",
    "serverless analytics",
    "deployment platform",
    "function deployment",
  ];

  if (!category) return baseKeywords;

  const categoryKeywords = {
    "Reliability": ["fast failover", "zero downtime", "high availability", "provider redundancy", "automatic failover"],
    "Deployment": ["multi-provider deployment", "vercel deployment", "netlify deployment", "fly.io deployment", "cloudflare deployment"],
    "Intelligence": ["AI routing", "predictive analytics", "machine learning", "smart routing", "traffic optimization"],
    "Developer Tools": ["developer experience", "git integration", "cli tools", "api support", "typescript support"],
    "Collaboration": ["team collaboration", "role permissions", "access control", "audit logs", "member invites"],
    "Monitoring": ["analytics dashboard", "real-time metrics", "performance monitoring", "deployment insights", "custom dashboards"],
    "Configuration": ["environment management", "custom domains", "configuration as code", "deployment settings", "environment variables"],
    "Infrastructure": ["global edge network", "edge locations", "low latency", "cdn", "geographic distribution"],
  };

  return [...baseKeywords, ...(categoryKeywords[category as keyof typeof categoryKeywords] || [])];
};

export function FeaturesPage() {
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const [expandedFeatures, setExpandedFeatures] = useState<string[]>([]);

  const toggleFeatureExpansion = (featureTitle: string) => {
    setExpandedFeatures((prev) => {
      const isCurrentlyExpanded = prev.includes(featureTitle);
      if (isCurrentlyExpanded) {
        return prev.filter(title => title !== featureTitle);
      } else {
        return [...prev, featureTitle];
      }
    });
  };

  const filteredFeatures = activeFilter
    ? features.filter((f) => f.category === activeFilter)
    : features;

  const filteredCategories = activeFilter ? [activeFilter] : categories;

  return (
    <div className="features-page">
      {/* SEO Meta Tags */}
      <MetaTags
        title={activeFilter ? `${activeFilter} Features | FunctionFly` : "Features | FunctionFly - Serverless Function Deployment Platform"}
        description={getCategoryMetaDescription(activeFilter)}
        keywords={getCategoryKeywords(activeFilter)}
        url={activeFilter ? `/features?category=${encodeURIComponent(activeFilter)}` : "/features"}
      />

      {/* Structured Data for SEO */}
      <StructuredData />

      {/* Navigation Bar */}
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        {/* Page Header */}
        <PageHeader />

        {/* Feature Filters */}
        <FeatureFilter
          categories={categories}
          features={features}
          activeFilter={activeFilter}
          onFilterChange={setActiveFilter}
          categoryColors={categoryColors}
        />

        {/* Features Grid */}
        <FeatureGrid
          features={filteredFeatures}
          expandedFeatures={expandedFeatures}
          onToggleFeatureExpansion={toggleFeatureExpansion}
        />

        {/* Interactive Feature Demos */}
        <div id="interactive-demos" className="interactive-features">
          <InteractiveFeatureDemos />
        </div>

        {/* Competitor Comparison */}
        <CompetitorComparison />

        {/* Feature Comparison */}
        <FeatureComparison />

        {/* Category Sections */}
        {filteredCategories.map((category, categoryIndex) => {
          const categoryFeatures = features.filter(
            (f) => f.category === category,
          );

          return (
            <CategorySection
              key={category}
              category={category}
              categoryIndex={categoryIndex}
              features={categoryFeatures}
            />
          );
        })}

        {/* CTA Section */}
        <CTASection />
      </div>

      <Footer />
    </div>
  );
}
