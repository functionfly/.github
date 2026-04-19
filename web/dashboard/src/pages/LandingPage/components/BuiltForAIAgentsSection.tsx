import { motion } from "framer-motion";
import { Bot, DollarSign, Gauge, Users, Shield, Database } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router-dom";

const agentFeatures = [
  {
    icon: DollarSign,
    title: "Per-Agent Cost Tracking",
    description: "Track spending per agent with granular cost attribution. Know exactly what each agent costs to run.",
    color: "#10b981",
    bgGradient: "from-emerald-500/20 to-green-500/10",
    borderColor: "#10b981",
  },
  {
    icon: Gauge,
    title: "Burst Handling",
    description: "Automatically handle sudden traffic spikes without performance degradation. Scale from 50 to 2000+ requests/sec.",
    color: "#3b82f6",
    bgGradient: "from-blue-500/20 to-blue-500/10",
    borderColor: "#3b82f6",
  },
  {
    icon: Shield,
    title: "Budget Enforcement",
    description: "Set spending limits per agent with automatic throttling when limits are reached. Never surprise bills.",
    color: "#f59e0b",
    bgGradient: "from-amber-500/20 to-amber-500/10",
    borderColor: "#f59e0b",
  },
  {
    icon: Users,
    title: "Multi-Agent Coordination",
    description: "Built-in coordination primitives for complex multi-agent workflows and agent-to-agent communication.",
    color: "#8b5cf6",
    bgGradient: "from-violet-500/20 to-violet-500/10",
    borderColor: "#8b5cf6",
  },
  {
    icon: Database,
    title: "State Management",
    description: "Durable state layer with automatic snapshots, deterministic replay, and cost attribution per tool call.",
    color: "#06b6d4",
    bgGradient: "from-cyan-500/20 to-cyan-500/10",
    borderColor: "#06b6d4",
  },
  {
    icon: Bot,
    title: "Agent Runtimes",
    description: "Direct integration with popular agent frameworks. Deploy Python, Node.js, or WASM-based agents.",
    color: "#ec4899",
    bgGradient: "from-pink-500/20 to-pink-500/10",
    borderColor: "#ec4899",
  },
];

export function BuiltForAIAgentsSection() {
  return (
    <section className="built-for-ai-section py-20 border-t border-border-subtle bg-gradient-to-b from-slate-900/50 to-slate-950/50">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-cyan-500/10 to-blue-500/10 border border-cyan-500/20 mb-6">
            <Bot className="w-4 h-4 text-cyan-400" />
            <span className="text-sm font-medium text-cyan-400">AI Agent Infrastructure</span>
          </div>
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
            Built for AI Agents
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto text-lg">
            The complete infrastructure layer for building, scaling, and monetizing AI agents.
            Cost controls, state management, and burst handling—built-in.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
          {agentFeatures.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              viewport={{ once: true }}
            >
              <Card className="h-full bg-gradient-to-br from-white/5 to-white/10 border-white/10 hover:border-white/20 transition-all duration-300 hover:shadow-lg group">
                <CardContent className="p-6">
                  <div
                    className={`w-12 h-12 rounded-lg bg-gradient-to-br ${feature.bgGradient} border flex items-center justify-center mb-4 group-hover:scale-110 transition-transform duration-300`}
                    style={{ borderColor: feature.color }}
                  >
                    <feature.icon className="w-6 h-6" style={{ color: feature.color }} />
                  </div>
                  <h3 className="text-lg font-semibold text-white mb-2">{feature.title}</h3>
                  <p className="text-text-secondary text-sm">{feature.description}</p>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>

        <div className="text-center">
          <Link
            to="/pricing"
            className="inline-flex items-center gap-2 px-6 py-3 rounded-lg bg-gradient-to-r from-cyan-500 to-blue-500 text-white font-semibold hover:from-cyan-600 hover:to-blue-600 transition-all duration-300"
          >
            View Agent Plans
            <Bot className="w-4 h-4" />
          </Link>
          <p className="text-text-secondary text-sm mt-4">
            Starting at $29/month • 14-day free trial • No credit card required
          </p>
        </div>
      </div>
    </section>
  );
}
