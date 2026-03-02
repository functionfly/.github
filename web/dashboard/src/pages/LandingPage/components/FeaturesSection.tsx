import { motion } from "framer-motion";
import { ArrowRightLeft, Network, Brain, Database, Zap } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

// Custom illustration component for Predictive Routing
const PredictiveRoutingIllustration = () => (
  <svg
    width="48"
    height="48"
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="w-6 h-6"
  >
    {/* Network nodes */}
    <circle cx="12" cy="12" r="3" fill="var(--success)" />
    <circle cx="36" cy="12" r="3" fill="var(--success)" />
    <circle cx="24" cy="36" r="3" fill="var(--success)" />

    {/* Connection lines with arrows */}
    <path d="M15 12h18" stroke="var(--success)" strokeWidth="2" />
    <path d="M33 9l3 3-3 3" stroke="var(--success)" strokeWidth="2" fill="none" />

    <path d="M27 33V15" stroke="var(--success)" strokeWidth="2" />
    <path d="M24 36l3-3-3-3" stroke="var(--success)" strokeWidth="2" fill="none" />

    <path d="M15 15 21 33" stroke="var(--success)" strokeWidth="2" />
    <path d="M18 30l3 3-3-3" stroke="var(--success)" strokeWidth="2" fill="none" />

    {/* AI brain icon in center */}
    <circle cx="24" cy="24" r="8" fill="var(--success)" opacity="0.2" />
    <path d="M20 24c0-2 1.5-3.5 3.5-3.5s3.5 1.5 3.5 3.5-1.5 3.5-3.5 3.5-3.5-1.5-3.5-3.5z" fill="var(--success)" />
    <path d="M22 22c0-0.5 0.5-1 1-1s1 0.5 1 1-0.5 1-1 1-1-0.5-1-1z" fill="var(--bg-primary)" />
  </svg>
);

const features = [
  {
    icon: ArrowRightLeft,
    title: "Fast Failover",
    description:
      "Sub-second failover between providers. Your users never notice downtime.",
    color: "var(--warning)",
    bgGradient: "from-warning/20 to-warning/10",
    borderColor: "var(--warning)",
  },
  {
    icon: Network,
    title: "No Vendor Lock-in",
    description:
      "Deploy to multiple edge providers simultaneously. Stay flexible.",
    color: "var(--brand-500)",
    bgGradient: "from-brand-500/20 to-brand-600/10",
    borderColor: "var(--brand-500)",
  },
  {
    icon: Brain,
    title: "Predictive Routing",
    description:
      "AI-powered traffic routing based on real-time health metrics.",
    color: "var(--success)",
    bgGradient: "from-success/20 to-success/10",
    borderColor: "var(--success)",
    illustration: PredictiveRoutingIllustration,
  },
  {
    icon: Database,
    title: "State Fabric",
    description:
      "Durable state management with automatic snapshots, deterministic replay, and cost attribution.",
    color: "#06b6d4",
    bgGradient: "from-cyan-500/20 to-cyan-500/10",
    borderColor: "#06b6d4",
  },
  {
    icon: Zap,
    title: "Function Registry",
    description:
      "Publish and discover reusable functions. Built-in marketplace for sharing and monetization.",
    color: "#f59e0b",
    bgGradient: "from-amber-500/20 to-amber-500/10",
    borderColor: "#f59e0b",
  },
];

export function FeaturesSection() {
  return (
    <section className="py-20 border-t border-border-subtle aurora-bg features-section-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <h2 className="features-section-headline text-3xl font-bold text-text-primary mb-4">
            Everything you need to stay online
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto">
            Built for developers who care about reliability without the
            enterprise complexity.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Card
                className="h-full card-elevation glass-card shine-effect"
                style={{
                  borderColor: `${feature.borderColor}1a`,
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = `${feature.borderColor}4d`;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = `${feature.borderColor}1a`;
                }}
              >
                <CardContent className="p-6">
                  <div
                    className="features-section-icon-wrap w-12 h-12 rounded-xl bg-linear-to-br border flex items-center justify-center mb-4 glow hover-glow transition-all duration-300"
                    style={{
                      background: `linear-gradient(135deg, ${feature.color}40, ${feature.color}20)`,
                      borderColor: `${feature.color}50`,
                      color: feature.color,
                    }}
                  >
                    {feature.illustration ? (
                      <feature.illustration />
                    ) : (
                      <feature.icon className="w-6 h-6 drop-shadow-sm text-inherit" />
                    )}
                  </div>
                  <h3 className="text-lg font-semibold text-text-primary mb-2">
                    {feature.title}
                  </h3>
                  <p className="text-text-secondary">{feature.description}</p>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
