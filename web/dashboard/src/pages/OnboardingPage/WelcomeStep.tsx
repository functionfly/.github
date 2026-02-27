import { motion } from "framer-motion";
import { Zap, Shield, Globe, Rocket, ArrowRight, CheckCircle } from "lucide-react";
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
      {/* Hero Section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="w-20 h-20 bg-linear-to-br from-[#6366f1] to-[#8b5cf6] rounded-full flex items-center justify-center mx-auto">
          <Zap className="w-10 h-10 text-white" fill="currentColor" />
        </div>
        <h2 className="text-3xl font-bold gradient-text">
          Welcome to FunctionFly
        </h2>
        <p className="text-lg text-text-secondary max-w-2xl mx-auto leading-relaxed">
          Deploy serverless functions with automatic failover across multiple cloud providers.
          Built for reliability, speed, and simplicity.
        </p>
      </motion.div>

      {/* Key Benefits */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="grid grid-cols-1 md:grid-cols-2 gap-4"
      >
        {benefits.map((benefit, index) => (
          <Card key={index} className="card p-4 hover:border-[#6366f1]/50 transition-colors">
            <div className="flex items-start gap-3">
              <div className="w-10 h-10 bg-[#6366f1]/20 rounded-lg flex items-center justify-center flex-shrink-0">
                <benefit.icon className="w-5 h-5 text-[#6366f1]" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-medium text-text-primary">{benefit.title}</h3>
                  <HelpTooltip content={benefit.tooltip} />
                </div>
                <p className="text-sm text-text-secondary">{benefit.description}</p>
              </div>
            </div>
          </Card>
        ))}
      </motion.div>

      {/* What to Expect */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        className="space-y-4"
      >
        <h3 className="text-xl font-semibold text-text-primary text-center">
          What to Expect
        </h3>
        <Card className="card p-6">
          <div className="space-y-3">
            {steps.map((step, index) => (
              <div key={index} className="flex items-center gap-3">
                <div className="w-6 h-6 bg-[#6366f1]/20 rounded-full flex items-center justify-center flex-shrink-0">
                  <span className="text-xs font-medium text-[#6366f1]">{index + 1}</span>
                </div>
                <span className="text-text-primary">{step}</span>
              </div>
            ))}
          </div>
        </Card>
      </motion.div>

      {/* Call to Action */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
        className="text-center space-y-4"
      >
        <div className="bg-green-500/10 border border-green-500/20 rounded-lg p-4 inline-block">
          <div className="flex items-center gap-2 text-green-400">
            <CheckCircle className="w-5 h-5" />
            <span className="font-medium">Ready to get started?</span>
          </div>
          <p className="text-sm text-text-secondary mt-1">
            This will take about 5 minutes to complete.
          </p>
        </div>
        <p className="text-text-secondary">
          Click <strong>Next</strong> to begin connecting your first provider.
        </p>
      </motion.div>
    </div>
  );
}