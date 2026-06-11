import { motion } from "framer-motion";
import { Zap, Shield, Globe, Rocket, CheckCircle } from "lucide-react";
import { Card } from "@/components/ui/card";
import { HelpTooltip } from "@/components/ui/help-tooltip";

export function WelcomeStep() {
  const benefits = [
    {
      icon: Globe,
      title: "Multi-Provider Deployment",
      description: "Deploy across Cloudflare, Vercel, and Fly.io simultaneously",
      tooltip: "FunctionFly automatically distributes your functions across multiple cloud providers, ensuring better performance and reliability."
    },
    {
      icon: Shield,
      title: "Automatic Failover",
      description: "Zero-downtime failover if any provider experiences issues",
      tooltip: "If one provider goes down, traffic automatically routes to healthy providers. Your functions stay online 24/7."
    },
    {
      icon: Rocket,
      title: "Edge Computing",
      description: "Functions run closer to users for faster response times",
      tooltip: "Edge functions execute at global locations near your users, reducing latency and improving performance."
    },
    {
      icon: Zap,
      title: "Simple Setup",
      description: "Connect providers and deploy in minutes",
      tooltip: "Get started quickly with our guided onboarding. We'll help you connect providers and deploy your first function."
    }
  ];

  const steps = [
    "Connect your cloud providers",
    "Deploy your first function",
    "Test automatic failover",
    "Go live with confidence"
  ];

  return (
    <div className="space-y-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="w-20 h-20 bg-gradient-to-br from-aviation-amber to-aviation-amber-glow rounded-full flex items-center justify-center mx-auto shadow-lg shadow-aviation-amber-dim">
          <Zap className="w-10 h-10 text-aviation-bg-primary" fill="currentColor" />
        </div>
        <p className="text-lg text-aviation-text-secondary max-w-2xl mx-auto leading-relaxed font-mono">
          Deploy serverless functions with automatic failover across multiple cloud providers.
          Built for reliability, speed, and simplicity.
        </p>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="grid grid-cols-1 md:grid-cols-2 gap-4"
      >
        {benefits.map((benefit, index) => (
          <Card key={index} className="aviation-instrument p-4 hover:border-aviation-amber-dim transition-all">
            <div className="flex items-start gap-3">
              <div className="w-10 h-10 bg-aviation-amber-subtle rounded-lg flex items-center justify-center flex-shrink-0">
                <benefit.icon className="w-5 h-5 text-aviation-amber" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-mono font-semibold text-aviation-text-primary">{benefit.title}</h3>
                  <HelpTooltip content={benefit.tooltip} />
                </div>
                <p className="text-sm text-aviation-text-secondary font-mono">{benefit.description}</p>
              </div>
            </div>
          </Card>
        ))}
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="space-y-4"
      >
        <h3 className="text-xl font-mono font-semibold text-aviation-text-primary text-center">
          What to Expect
        </h3>
        <Card className="aviation-panel p-6">
          <div className="space-y-3">
            {steps.map((step, index) => (
              <div key={index} className="flex items-center gap-3">
                <div className="w-6 h-6 bg-aviation-amber-subtle rounded-full flex items-center justify-center flex-shrink-0">
                  <span className="text-xs font-mono font-bold text-aviation-amber">{index + 1}</span>
                </div>
                <span className="text-aviation-text-primary font-mono">{step}</span>
              </div>
            ))}
          </div>
        </Card>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
        className="text-center space-y-4"
      >
        <div className="bg-aviation-green-dim border border-aviation-green/30 rounded-lg p-4 inline-block">
          <div className="flex items-center gap-2 text-aviation-green">
            <CheckCircle className="w-5 h-5" />
            <span className="font-mono font-medium">Ready to get started?</span>
          </div>
          <p className="text-sm text-aviation-text-secondary mt-1 font-mono">
            This will take about 5 minutes to complete.
          </p>
        </div>
        <p className="text-aviation-text-secondary font-mono">
          Click <strong>Next</strong> to begin connecting your first provider.
        </p>
      </motion.div>
    </div>
  );
}