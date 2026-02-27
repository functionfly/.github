import { useState } from "react";
import { motion } from "framer-motion";
import { Footer } from "@/pages/LandingPage/components";
import { MetaTags } from "@/components/seo/MetaTags";
import {
  Navigation,
  PageHeader,
  CTASection,
  IntegrationGrid,
  CategorySection,
  IntegrationFilter
} from "./components";
import {
  integrations,
  integrationCategories,
  getIntegrationsByCategory
} from "./data/integrations";
import { categoryColors } from "./types";

// Helper functions for meta descriptions and keywords
const getCategoryMetaDescription = (category?: string | null) => {
  const baseDescription = "Connect FunctionFly with your favorite platforms, frameworks, and services. Comprehensive integrations for cloud providers, databases, APIs, and monitoring tools.";

  if (!category) return baseDescription;

  const categoryDescriptions = {
    "Cloud Providers": "Deploy to multiple cloud platforms simultaneously with FunctionFly's cloud provider integrations. AWS Lambda, Google Cloud Functions, Vercel, Netlify, and more.",
    "Frameworks": "Native support for popular web frameworks and runtimes. Express.js, Next.js, Fastify, Django, and other frameworks work seamlessly with FunctionFly.",
    "Deployment Platforms": "One-click deployment to your preferred hosting platforms. Railway, Render, Fly.io, Vercel, Netlify, and Heroku integrations included.",
    "Databases": "Connect to popular databases and data stores. PostgreSQL, MongoDB, Redis, and other database integrations for your serverless functions.",
    "APIs & Services": "Integrate with third-party APIs and microservices. Stripe, GitHub, Slack, and other service integrations to extend your functionality.",
    "Monitoring & Analytics": "Track performance, errors, and user behavior with monitoring integrations. Sentry, Datadog, New Relic, and other analytics tools.",
  };

  return categoryDescriptions[category as keyof typeof categoryDescriptions] || baseDescription;
};

const getCategoryKeywords = (category?: string | null) => {
  const baseKeywords = [
    "integrations",
    "api integrations",
    "platform integrations",
    "functionfly integrations",
    "serverless integrations",
    "cloud integrations"
  ];

  if (!category) return baseKeywords;

  const categoryKeywords = {
    "Cloud Providers": ["aws lambda integration", "google cloud functions", "vercel integration", "netlify integration", "cloudflare workers"],
    "Frameworks": ["express.js integration", "next.js integration", "fastify integration", "django integration", "framework integration"],
    "Deployment Platforms": ["railway integration", "render integration", "fly.io integration", "vercel deployment", "netlify deployment"],
    "Databases": ["postgresql integration", "mongodb integration", "redis integration", "database integration", "data store"],
    "APIs & Services": ["stripe integration", "github integration", "slack integration", "api service", "third-party integration"],
    "Monitoring & Analytics": ["sentry integration", "datadog integration", "new relic integration", "monitoring", "analytics integration"],
  };

  return [...baseKeywords, ...(categoryKeywords[category as keyof typeof categoryKeywords] || [])];
};

export function IntegrationsPage() {
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const [expandedIntegrations, setExpandedIntegrations] = useState<Set<string>>(
    new Set(),
  );

  const toggleExpansion = (integrationId: string) => {
    setExpandedIntegrations((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(integrationId)) {
        newSet.delete(integrationId);
      } else {
        newSet.add(integrationId);
      }
      return newSet;
    });
  };

  const filteredIntegrations = activeFilter
    ? integrations.filter((integration) => integration.category === activeFilter)
    : integrations;

  const filteredCategories = activeFilter ? [activeFilter] : integrationCategories.map(cat => cat.name);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative overflow-hidden">
      {/* Background decorative elements */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(99,102,241,0.1),transparent_50%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_70%_80%,rgba(139,92,246,0.1),transparent_50%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[linear-gradient(45deg,transparent_25%,rgba(255,255,255,0.01)_25%,rgba(255,255,255,0.01)_50%,transparent_50%,transparent_75%,rgba(255,255,255,0.01)_75%)] bg-[length:20px_20px] opacity-30 pointer-events-none" />

      {/* SEO Meta Tags */}
      <MetaTags
        title={activeFilter ? `${activeFilter} Integrations | FunctionFly` : "Integrations | FunctionFly - Serverless Function Deployment Platform"}
        description={getCategoryMetaDescription(activeFilter)}
        keywords={getCategoryKeywords(activeFilter)}
        url={activeFilter ? `/integrations?category=${encodeURIComponent(activeFilter)}` : "/integrations"}
      />

      {/* Navigation Bar */}
      <Navigation />

      <div className="relative z-10 max-w-7xl mx-auto px-4 lg:px-6 py-8">
        {/* Page Header */}
        <PageHeader />

        {/* Integration Filters */}
        <div className="mb-16">
          <IntegrationFilter
            categories={integrationCategories}
            integrations={integrations}
            activeFilter={activeFilter}
            onFilterChange={setActiveFilter}
            categoryColors={categoryColors}
          />
        </div>

        {/* All Integrations Grid (when no filter) */}
        {!activeFilter && (
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.4 }}
            className="mb-24"
          >
            <div className="text-center mb-12">
              <div className="inline-flex items-center gap-3 px-4 py-2 rounded-full bg-white/5 border border-white/10 mb-6">
                <div className="w-2 h-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] animate-pulse" />
                <span className="text-sm font-medium text-text-secondary">
                  {integrations.length} Integrations Available
                </span>
              </div>
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
                All Integrations
              </h2>
              <p className="text-text-secondary text-lg max-w-2xl mx-auto">
                Explore our complete integration ecosystem and connect FunctionFly with your favorite platforms
              </p>
            </div>
            <IntegrationGrid
              integrations={filteredIntegrations}
              expandedIntegrations={expandedIntegrations}
              onToggleExpansion={toggleExpansion}
            />
          </motion.div>
        )}

        {/* Category Sections */}
        {activeFilter && filteredCategories.map((category, categoryIndex) => {
          const categoryIntegrations = getIntegrationsByCategory(category);

          return (
            <motion.div
              key={category}
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: categoryIndex * 0.1 }}
              className="mb-24"
            >
              <CategorySection
                category={category}
                categoryIndex={categoryIndex}
                integrations={categoryIntegrations}
                expandedIntegrations={expandedIntegrations}
                onToggleExpansion={toggleExpansion}
              />
            </motion.div>
          );
        })}

        {/* CTA Section */}
        <div className="relative">
          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent blur-3xl -mx-8" />
          <CTASection />
        </div>
      </div>

      <Footer />
    </div>
  );
}