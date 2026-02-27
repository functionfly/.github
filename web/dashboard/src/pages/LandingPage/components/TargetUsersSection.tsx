import { motion } from "framer-motion";
import { Check, Sparkles } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

const targetUsers = [
  {
    icon: "User" as const,
    title: "Indie SaaS Founder",
    subtitle: "Deploying your first app",
    description: "You built an amazing product and now need to ensure it stays online 24/7. FunctionFly gives you enterprise-grade reliability without the enterprise price tag.",
    benefits: [
      "Zero-downtime deployments",
      "Automatic scaling and failover",
      "Focus on building, not infrastructure",
      "Pay only for what you use",
    ],
    scenario: "Your app goes viral on Product Hunt - FunctionFly handles the traffic spike automatically.",
    gradient: "from-blue-500/20 to-cyan-500/20",
    borderColor: "border-blue-500/20",
  },
  {
    icon: "Building" as const,
    title: "Growing Startup",
    subtitle: "Needing high availability",
    description: "Your user base is growing rapidly, and downtime costs you customers and revenue. Multi-cloud redundancy ensures your app stays online when others fail.",
    benefits: [
      "99.99% uptime SLA",
      "Sub-200ms failover time",
      "Multi-region deployment",
      "Predictive failure detection",
    ],
    scenario: "During a major cloud provider outage, your app seamlessly switches to backup providers - users never notice.",
    gradient: "from-purple-500/20 to-pink-500/20",
    borderColor: "border-purple-500/20",
  },
  {
    icon: "Building2" as const,
    title: "Enterprise",
    subtitle: "Wanting multi-cloud strategy",
    description: "Your organization requires vendor diversity and compliance across multiple cloud providers. FunctionFly provides the flexibility and control you need.",
    benefits: [
      "Multi-cloud deployment",
      "Advanced compliance features",
      "Enterprise security standards",
      "Dedicated support and SLAs",
    ],
    scenario: "Comply with regulatory requirements by distributing workloads across AWS, GCP, and Azure simultaneously.",
    gradient: "from-emerald-500/20 to-teal-500/20",
    borderColor: "border-emerald-500/20",
  },
];

const realWorldScenarios = [
  {
    icon: "ShoppingCart" as const,
    title: "E-commerce Black Friday",
    description: "Handle massive traffic spikes during peak shopping seasons without service degradation.",
    impact: "Zero lost sales during high-traffic periods",
  },
  {
    icon: "Lightning" as const,
    title: "Social Media Virality",
    description: "Your content goes viral - FunctionFly automatically scales across providers to handle the load.",
    impact: "Seamless scaling during unexpected traffic surges",
  },
  {
    icon: "Target" as const,
    title: "Product Launches",
    description: "Launch new features or products with confidence, knowing failover is automatic.",
    impact: "Reliable performance during critical business moments",
  },
  {
    icon: "Globe" as const,
    title: "Global Expansion",
    description: "Deploy to multiple regions simultaneously for global users with consistent performance.",
    impact: "Worldwide availability with local performance",
  },
];

const iconComponents = {
  User: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
    </svg>
  ),
  Building: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
    </svg>
  ),
  Building2: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
    </svg>
  ),
  ShoppingCart: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4m0 0L7 13m0 0l-1.1 5H19M7 13v8a2 2 0 002 2h10a2 2 0 002-2v-3" />
    </svg>
  ),
  Lightning: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
    </svg>
  ),
  Target: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
  Globe: ({ className }: { className?: string }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
};

export function TargetUsersSection() {
  return (
    <section className="py-20 border-t border-border-subtle gradient-shift-bg target-users-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold text-white mb-4">
            Who benefits most from FunctionFly
          </h2>
          <p className="text-text-primary text-glow max-w-2xl mx-auto text-balance target-users-subtitle">
            From solo founders to enterprise teams, FunctionFly provides the reliability
            and flexibility you need to focus on building great products.
          </p>
        </div>

        {/* Target Users */}
        <div className="grid md:grid-cols-3 gap-8 mb-20">
          {targetUsers.map((user, index) => {
            const IconComponent = iconComponents[user.icon];
            return (
              <motion.div
                key={user.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
              >
                <Card className={`h-full hover:border-[#6366f1]/30 transition-colors card-elevation glass-card shine-effect ${user.borderColor}`}>
                  <CardContent className="p-8">
                    <div className="flex items-center gap-4 mb-6">
                      <div className={`w-14 h-14 rounded-2xl bg-linear-to-br ${user.gradient} border border-white/20 flex items-center justify-center glow hover-glow`}>
                        <IconComponent className="w-7 h-7 text-white drop-shadow-lg" />
                      </div>
                      <div>
                        <h3 className="text-xl font-semibold text-white mb-1">
                          {user.title}
                        </h3>
                        <p className="text-sm text-text-secondary">
                          {user.subtitle}
                        </p>
                      </div>
                    </div>

                    <p className="text-text-secondary mb-6">
                      {user.description}
                    </p>

                    <div className="mb-6">
                      <h4 className="text-sm font-semibold text-white mb-3">Key Benefits:</h4>
                      <ul className="space-y-2">
                        {user.benefits.map((benefit, benefitIndex) => (
                          <li key={benefitIndex} className="flex items-start gap-3">
                            <div className="w-4 h-4 rounded-full bg-emerald-500/20 flex items-center justify-center mt-0.5">
                              <Check className="w-2.5 h-2.5 text-emerald-400" />
                            </div>
                            <span className="text-sm text-text-secondary">
                              {benefit}
                            </span>
                          </li>
                        ))}
                      </ul>
                    </div>

                    <div className="pt-4 border-t border-white/8">
                      <div className="flex items-start gap-3">
                        <Sparkles className="w-4 h-4 text-[#6366f1] mt-0.5" />
                        <div>
                          <p className="text-sm font-medium text-[#6366f1] mb-1">Real Scenario:</p>
                          <p className="text-sm text-text-secondary">
                            {user.scenario}
                          </p>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            );
          })}
        </div>

        {/* Real-World Scenarios */}
        <div>
          <div className="text-center mb-12">
            <h3 className="text-2xl font-bold text-white mb-4">
              Real-world scenarios where FunctionFly shines
            </h3>
            <p className="text-text-secondary max-w-2xl mx-auto">
              See how FunctionFly handles the toughest challenges in production environments.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {realWorldScenarios.map((scenario, index) => {
              const IconComponent = iconComponents[scenario.icon];
              return (
                <motion.div
                  key={scenario.title}
                  initial={{ opacity: 0, scale: 0.95 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                >
                  <Card className="h-full hover:border-[#6366f1]/30 transition-colors card-elevation glass-card shine-effect">
                    <CardContent className="p-6 text-center">
                      <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-linear-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 border border-[#6366f1]/20 flex items-center justify-center">
                        <IconComponent className="w-6 h-6 text-[#6366f1]" />
                      </div>
                      <h4 className="text-lg font-semibold text-white mb-2">
                        {scenario.title}
                      </h4>
                      <p className="text-text-secondary text-sm mb-4">
                        {scenario.description}
                      </p>
                      <div className="pt-4 border-t border-white/8">
                        <p className="text-xs text-[#6366f1] font-medium">
                          {scenario.impact}
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}