import { motion } from "framer-motion";
import { Activity, Clock, Globe, Users, Shield, Lock, CheckCircle, Award } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

const targetMetrics = [
  {
    icon: Activity,
    value: "99.99%",
    label: "Target Uptime SLA",
    description: "Enterprise-grade reliability guarantee",
  },
  {
    icon: Clock,
    value: "< 200ms",
    label: "Failover Time",
    description: "Sub-second switching between providers",
  },
  {
    icon: Globe,
    value: "4",
    label: "Cloud Providers",
    description: "Multi-cloud redundancy out of the box",
  },
  {
    icon: Users,
    value: "100+",
    label: "Apps Target",
    description: "Indie SaaS applications we aim to support",
  },
];

const trustBadges = [
  {
    icon: Shield,
    title: "SOC 2 Compliant",
    description: "Security and compliance standards",
  },
  {
    icon: Lock,
    title: "GDPR Compliant",
    description: "European data protection standards",
  },
  {
    icon: CheckCircle,
    title: "ISO 27001",
    description: "Information security management",
  },
  {
    icon: Award,
    title: "99.99% SLA",
    description: "Enterprise-grade reliability guarantee",
  },
];

export function TrustMetricsSection() {
  return (
    <section className="py-20 border-t border-border-subtle gradient-shift-bg trust-metrics-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        {/* Target Metrics */}
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold text-text-primary mb-4">
            Built for reliability from day one
          </h2>
          <p className="text-text-primary text-glow max-w-2xl mx-auto mb-12 text-balance">
            We're committed to enterprise-grade standards that indie SaaS
            founders can trust.
          </p>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
            {targetMetrics.map((metric, index) => (
              <motion.div
                key={metric.label}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                className="text-center"
              >
                <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gradient-to-br from-brand-500/30 to-purple-500/30 border border-brand-500/40 flex items-center justify-center glow">
                  <metric.icon className="w-8 h-8 text-white drop-shadow-lg" />
                </div>
                <div className="text-3xl font-bold text-text-primary mb-1">
                  {metric.value}
                </div>
                <div className="text-sm font-medium text-text-primary mb-1">
                  {metric.label}
                </div>
                <div className="text-xs text-text-primary/80">
                  {metric.description}
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Trust Badges */}
        <div>
          <div className="text-center mb-8">
            <h3 className="text-xl font-semibold text-text-primary mb-2">
              Security & compliance commitments
            </h3>
            <p className="text-text-secondary">
              Industry-leading standards you can count on
            </p>
          </div>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {trustBadges.map((badge, index) => (
              <motion.div
                key={badge.title}
                initial={{ opacity: 0, scale: 0.95 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                className="text-center"
              >
                <Card className="h-full hover:border-brand-500/30 transition-colors card-elevation glass-card shine-effect">
                  <CardContent className="p-4">
                    <div className="w-12 h-12 mx-auto mb-3 rounded-xl bg-gradient-to-br from-success/30 to-success/20 border border-success/40 flex items-center justify-center glow">
                      <badge.icon className="w-6 h-6 text-white drop-shadow-md" />
                    </div>
                    <h4 className="text-sm font-semibold text-text-primary mb-1">
                      {badge.title}
                    </h4>
                    <p className="text-xs text-text-primary/90">
                      {badge.description}
                    </p>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}