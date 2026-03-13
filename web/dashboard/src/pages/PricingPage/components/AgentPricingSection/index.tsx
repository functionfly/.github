import { motion } from "framer-motion";
import { useScrollAnimation } from "../../hooks";
import { AGENT_PLANS } from "@/lib/constants";
import { AgentPricingCard } from "./AgentPricingCard";

interface AgentPricingSectionProps {
  /** When true, use a compact header (e.g. when shown inside pricing tabs). */
  compact?: boolean;
  /** Callback when a plan is selected for upgrade/checkout */
  onPlanSelect?: (planId: string, priceId?: string) => void;
}

/**
 * Agent pricing section for the main /pricing page.
 * Highlights Agent Execution Plans (AEP) for AI agent infrastructure.
 */
export function AgentPricingSection({ compact, onPlanSelect }: AgentPricingSectionProps) {
  const { ref, inView } = useScrollAnimation(0.1, false);

  const agentPlans = [
    {
      id: AGENT_PLANS.STARTER.id,
      name: AGENT_PLANS.STARTER.name,
      tagline: "For prototypes",
      price: `${AGENT_PLANS.STARTER.price}`,
      period: "month",
      description: AGENT_PLANS.STARTER.description,
      features: [...AGENT_PLANS.STARTER.features],
      highlighted: false,
      cta: "Start Free Trial",
      href: "/signup",
      priceId: AGENT_PLANS.STARTER.priceId,
    },
    {
      id: AGENT_PLANS.SCALE.id,
      name: AGENT_PLANS.SCALE.name,
      tagline: "Most Popular",
      price: `${AGENT_PLANS.SCALE.price}`,
      period: "month",
      description: AGENT_PLANS.SCALE.description,
      features: [...AGENT_PLANS.SCALE.features],
      highlighted: true,
      cta: "Start Free Trial",
      href: "/signup",
      priceId: AGENT_PLANS.SCALE.priceId,
    },
    {
      id: AGENT_PLANS.PRO.id,
      name: AGENT_PLANS.PRO.name,
      tagline: "Production",
      price: `${AGENT_PLANS.PRO.price}`,
      period: "month",
      description: AGENT_PLANS.PRO.description,
      features: [...AGENT_PLANS.PRO.features],
      highlighted: false,
      cta: "Start Free Trial",
      href: "/signup",
      priceId: AGENT_PLANS.PRO.priceId,
    },
    {
      id: AGENT_PLANS.ENTERPRISE.id,
      name: AGENT_PLANS.ENTERPRISE.name,
      tagline: "Enterprise",
      price: "Custom",
      period: "pricing",
      description: AGENT_PLANS.ENTERPRISE.description,
      features: [...AGENT_PLANS.ENTERPRISE.features],
      highlighted: false,
      cta: "Contact Sales",
      href: "/contact",
      priceId: AGENT_PLANS.ENTERPRISE.priceId,
    },
  ];

  return (
    <motion.section
      id="agent-plans"
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.6, ease: "easeOut" }}
      className="pricing-agent-section mb-24"
    >
      <div className={compact ? "text-center mb-8" : "text-center mb-12"}>
        {!compact && (
          <>
            <div className="relative inline-block mb-4">
              <div className="px-4 py-1.5 rounded-full bg-gradient-to-r from-cyan-500/10 to-blue-500/10 border border-cyan-500/20">
                <span className="text-sm font-medium text-cyan-400">AI Agent Infrastructure</span>
              </div>
            </div>
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              Agent Execution Plans
            </h2>
          </>
        )}
        <p className="text-text-secondary max-w-2xl mx-auto text-lg">
          Scale your AI agents with built-in cost tracking, budget enforcement, and burst handling.
          Start free, scale with usage.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto pt-6">
        {agentPlans.map((plan, index) => (
          <AgentPricingCard 
            key={plan.id} 
            plan={plan} 
            index={index} 
            inView={inView}
            onPlanSelect={onPlanSelect}
          />
        ))}
      </div>

      {/* Feature comparison note */}
      <div className="mt-12 text-center">
        <p className="text-text-secondary text-sm">
          All plans include: Per-agent cost attribution • Automatic budget enforcement •
          Real-time spend monitoring • API access
        </p>
      </div>
    </motion.section>
  );
}
