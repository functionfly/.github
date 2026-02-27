import { motion } from "framer-motion";
import { ArrowRight, Database, Network, Zap, Shield, Clock, Globe, Layers, CheckCircle, BarChart3, Users, Lock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Link } from "react-router-dom";
import { Footer } from "@/pages/LandingPage/components";
import { Navbar } from "@/components/common/Navbar";
import { MetaTags } from "@/components/seo/MetaTags";
import { HeroSection } from "./components/HeroSection";
import { ProblemSection } from "./components/ProblemSection";
import { WhatIsStateFabricSection } from "./components/WhatIsStateFabricSection";
import { ArchitectureVisualizationSection } from "./components/ArchitectureVisualizationSection";
import { CoreCapabilitiesSection } from "./components/CoreCapabilitiesSection";
import { BuiltForAIAgentsSection } from "./components/BuiltForAIAgentsSection";
import { DeterministicReplaySection } from "./components/DeterministicReplaySection";
import { ComparisonSection } from "./components/ComparisonSection";
import { PricingCTASection } from "./components/PricingCTASection";
import { FAQSection } from "./components/FAQSection";
import { useState } from "react";
import "@/styles/index.css";

const fadeInUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.6 }
};

const stagger = {
  animate: {
    transition: {
      staggerChildren: 0.1
    }
  }
};

const features = [
  {
    icon: <Database className="w-8 h-8 text-blue-500" />,
    title: "State Management",
    description: "Unified state management across distributed systems with automatic synchronization and conflict resolution."
  },
  {
    icon: <Network className="w-8 h-8 text-green-500" />,
    title: "Data Orchestration",
    description: "Orchestrate complex data flows between services with visual pipelines and real-time monitoring."
  },
  {
    icon: <Zap className="w-8 h-8 text-yellow-500" />,
    title: "High Performance",
    description: "Sub-millisecond latency with global edge caching and intelligent data routing."
  },
  {
    icon: <Shield className="w-8 h-8 text-red-500" />,
    title: "Enterprise Security",
    description: "End-to-end encryption, audit trails, and compliance with SOC 2, GDPR, and HIPAA."
  },
  {
    icon: <Clock className="w-8 h-8 text-purple-500" />,
    title: "Real-time Sync",
    description: "Instant synchronization across all connected applications and services."
  },
  {
    icon: <Globe className="w-8 h-8 text-indigo-500" />,
    title: "Global Scale",
    description: "Deploy state fabrics globally with automatic failover and load balancing."
  }
];

const useCases = [
  {
    title: "E-commerce Platforms",
    description: "Manage product catalogs, shopping carts, and user sessions across global storefronts.",
    icon: <Layers className="w-6 h-6" />,
    metrics: ["99.9% uptime", "50ms avg latency", "10M+ concurrent users"]
  },
  {
    title: "Financial Services",
    description: "Handle transaction processing, account balances, and real-time trading data.",
    icon: <BarChart3 className="w-6 h-6" />,
    metrics: ["99.99% uptime", "10ms avg latency", "100B+ daily transactions"]
  },
  {
    title: "Social Platforms",
    description: "Scale user interactions, feeds, and real-time messaging for millions of users.",
    icon: <Users className="w-6 h-6" />,
    metrics: ["99.95% uptime", "25ms avg latency", "1B+ daily interactions"]
  }
];

const benefits = [
  "Zero-configuration setup",
  "Automatic scaling",
  "Multi-cloud support",
  "Developer-friendly APIs",
  "Built-in monitoring",
  "Enterprise support"
];

export function StateFabricMarketingPage() {
  const [activeTab, setActiveTab] = useState(0);

  return (
    <div className="min-h-screen bg-bg-primary state-fabric-page">
      {/* SEO Meta Tags */}
      <MetaTags
        title="State Fabric - Distributed State Management & Data Orchestration"
        description="Build scalable applications with State Fabric. Distributed state management, data orchestration, and real-time synchronization across global applications. Enterprise-grade performance and security."
        keywords={["state management", "data orchestration", "distributed systems", "real-time sync", "global state", "data fabric", "state synchronization"]}
        url="/state-fabric"
      />

      <Navbar variant="landing" />

      <HeroSection />

      <ProblemSection />

      <WhatIsStateFabricSection />

      <ArchitectureVisualizationSection />

      <CoreCapabilitiesSection />

      <BuiltForAIAgentsSection />

      <DeterministicReplaySection />

      <ComparisonSection />

      <PricingCTASection />

      <FAQSection />

      {/* Features Grid */}
      <section className="py-20 bg-bg-secondary/50 features-section-enhanced">
        <div className="container mx-auto px-4">
          <motion.div
            className="text-center mb-16"
            initial="initial"
            whileInView="animate"
            viewport={{ once: true }}
            variants={stagger}
          >
            <motion.h2 variants={fadeInUp} className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-4 animate-fade-in-up">
              Powerful Features for Modern Applications
            </motion.h2>
            <motion.p variants={fadeInUp} className="text-xl text-slate-600 dark:text-text-secondary max-w-2xl mx-auto animate-fade-in-up animate-delay-100">
              Everything you need to build distributed, real-time applications at scale.
            </motion.p>
          </motion.div>

          <motion.div
            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8"
            initial="initial"
            whileInView="animate"
            viewport={{ once: true }}
            variants={stagger}
          >
            {features.map((feature, index) => (
              <motion.div key={index} variants={fadeInUp}>
                <Card className="glass-card h-full hover-lift animate-fade-in-up">
                  <CardHeader>
                    <div className="w-12 h-12 rounded-lg glass-light flex items-center justify-center mb-4 animate-float">
                      {feature.icon}
                    </div>
                    <CardTitle className="text-xl text-slate-900 dark:text-white">{feature.title}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-slate-600 dark:text-text-secondary">{feature.description}</p>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </motion.div>
        </div>
      </section>

      {/* Use Cases */}
      <section className="py-20 target-users-enhanced">
        <div className="container mx-auto px-4">
          <motion.div
            className="text-center mb-16"
            initial="initial"
            whileInView="animate"
            viewport={{ once: true }}
            variants={stagger}
          >
            <motion.h2 variants={fadeInUp} className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-4 animate-fade-in-up">
              Trusted by Industry Leaders
            </motion.h2>
            <motion.p variants={fadeInUp} className="target-users-subtitle text-xl text-slate-600 dark:text-text-secondary max-w-2xl mx-auto animate-fade-in-up animate-delay-100">
              See how organizations are transforming their applications with State Fabric.
            </motion.p>
          </motion.div>

          <motion.div
            className="grid grid-cols-1 lg:grid-cols-3 gap-8"
            initial="initial"
            whileInView="animate"
            viewport={{ once: true }}
            variants={stagger}
          >
            {useCases.map((useCase, index) => (
              <motion.div key={index} variants={fadeInUp}>
                <Card className="glass-card h-full hover-lift animate-fade-in-up animate-stagger">
                  <CardHeader>
                    <div className="flex items-center gap-3 mb-4">
                      <div className="w-10 h-10 rounded-lg glass-light flex items-center justify-center animate-pulse-scale">
                        {useCase.icon}
                      </div>
                      <CardTitle className="text-xl text-slate-900 dark:text-white">{useCase.title}</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <p className="text-slate-600 dark:text-text-secondary mb-6">{useCase.description}</p>
                    <div className="space-y-2">
                      {useCase.metrics.map((metric, metricIndex) => (
                        <div key={metricIndex} className="flex items-center gap-2 text-sm animate-fade-in-up">
                          <CheckCircle className="w-4 h-4 text-green-400 animate-bounce" />
                          <span className="text-slate-600 dark:text-text-secondary">{metric}</span>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </motion.div>
        </div>
      </section>

      {/* Benefits Section */}
      <section className="py-20 bg-bg-secondary/50 benefits-section-enhanced">
        <div className="container mx-auto px-4">
          <div className="max-w-4xl mx-auto">
            <motion.div
              className="text-center mb-16"
              initial="initial"
              whileInView="animate"
              viewport={{ once: true }}
              variants={stagger}
            >
              <motion.h2 variants={fadeInUp} className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-4 animate-fade-in-up">
                Why Choose State Fabric?
              </motion.h2>
              <motion.p variants={fadeInUp} className="text-xl text-slate-600 dark:text-text-secondary animate-fade-in-up animate-delay-100">
                Built for developers, trusted by enterprises.
              </motion.p>
            </motion.div>

            <motion.div
              className="grid grid-cols-1 md:grid-cols-2 gap-6"
              initial="initial"
              whileInView="animate"
              viewport={{ once: true }}
              variants={stagger}
            >
              {benefits.map((benefit, index) => (
                <motion.div key={index} variants={fadeInUp}>
                  <Card className="glass-card hover-lift animate-fade-in-up animate-stagger">
                    <CardContent className="p-6">
                      <div className="flex items-center gap-3">
                        <CheckCircle className="w-5 h-5 text-green-400 shrink-0 animate-bounce" />
                        <span className="text-slate-900 dark:text-white font-medium">{benefit}</span>
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              ))}
            </motion.div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 gradient-shift-bg">
        <div className="container mx-auto px-4">
          <motion.div
            className="max-w-4xl mx-auto text-center"
            initial="initial"
            whileInView="animate"
            viewport={{ once: true }}
            variants={stagger}
          >
            <motion.h2 variants={fadeInUp} className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-4 animate-fade-in-up">
              Ready to Transform Your Applications?
            </motion.h2>
            <motion.p variants={fadeInUp} className="text-xl text-slate-600 dark:text-text-secondary mb-8 max-w-2xl mx-auto animate-fade-in-up animate-delay-100">
              Join thousands of developers building the next generation of distributed applications with State Fabric.
            </motion.p>

            <motion.div variants={fadeInUp} className="flex flex-col sm:flex-row gap-4 justify-center animate-fade-in-up animate-delay-200">
              <Button size="lg" className="hero-primary-button gap-2 glow-lg animate-pulse-scale">
                Get Started Free
                <ArrowRight className="w-4 h-4" />
              </Button>
              <Button variant="outline" size="lg" className="hero-outline-button">
                Schedule Demo
              </Button>
            </motion.div>

            <motion.div variants={fadeInUp} className="mt-8 text-sm cta-disclaimer animate-fade-in-up animate-delay-300">
              <Lock className="w-4 h-4 inline mr-2 animate-bounce" />
              Free tier includes 100GB storage and 1M operations/month. No credit card required.
            </motion.div>
          </motion.div>
        </div>
      </section>

      <Footer />
    </div>
  );
}