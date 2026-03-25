import { AGENT_PLANS } from '@/lib/constants';
import { motion } from 'framer-motion';
import { useScrollAnimation } from '../../hooks';
import { AgentPricingCard } from './AgentPricingCard';

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
      id: 'free',
      name: 'Free',
      tagline: 'Trusted tool discovery baseline',
      price: '$0',
      period: 'pricing',
      description:
        'Start running agents with verified tool discovery, trust scores, and a minimal audit trail.',
      features: [
        'Verified tool discovery (trust_score + trust_level)',
        'Trust policy filters (min score / verification gate)',
        'Audit trail retention: limited',
        'Marketplace access to verified functions',
        'Basic Trust API usage for trust signals',
      ],
      highlighted: false,
      cta: 'Get started',
      href: '/signup',
      priceId: undefined,
    },
    {
      id: AGENT_PLANS.STARTER.id,
      name: AGENT_PLANS.STARTER.name,
      tagline: 'For trust-aware prototyping',
      price: `${AGENT_PLANS.STARTER.price}`,
      period: 'month',
      description:
        'Verification-backed execution with extended audit retention and starter policy templates.',
      features: [
        'Verification + signing gates for trusted tool selection',
        'Audit trail retention: extended (agent logs)',
        'Trust policy templates (starter set)',
        'Usage-based Trust API for trust signals',
        // Preserve the core agent-execution value from the existing plan data
        '$5 daily spend cap • per-agent cost tracking • budget enforcement',
      ],
      highlighted: false,
      cta: 'Start Free Trial',
      href: '/signup',
      priceId: AGENT_PLANS.STARTER.priceId,
    },
    {
      id: AGENT_PLANS.SCALE.id,
      name: AGENT_PLANS.SCALE.name,
      tagline: 'Most Popular • audit-ready agents',
      price: `${AGENT_PLANS.SCALE.price}`,
      period: 'month',
      description:
        'Production-grade trust controls: longer retention, export-ready logs, and stronger trust routing.',
      features: [
        'Verification-backed trusted tool routing',
        'Audit trail retention: 90 days (agent logs)',
        'Policy packs (org templates) for trust constraints',
        'Trust API: higher limits for trust-aware discovery',
        // Keep execution scaling signal
        '500 calls/second burst • 100 concurrent agents',
      ],
      highlighted: true,
      cta: 'Start Free Trial',
      href: '/signup',
      priceId: AGENT_PLANS.SCALE.priceId,
    },
    {
      id: AGENT_PLANS.ENTERPRISE.id,
      name: AGENT_PLANS.ENTERPRISE.name,
      tagline: 'Enterprise',
      price: 'Custom',
      period: 'pricing',
      description:
        'Custom trust controls for compliance: dedicated certification flows, tailored retention, and high-throughput Trust API usage.',
      features: [
        'Custom verification / trust certification support',
        'Audit trail retention: configurable for compliance',
        'Policy packs: custom trust policy templates',
        'Trust API: usage-based access at enterprise scale',
        // Preserve core enterprise agent signal
        'Unlimited tool calls • multi-region deployment • 24/7 phone support',
      ],
      highlighted: false,
      cta: 'Contact Sales',
      href: '/contact',
      priceId: AGENT_PLANS.ENTERPRISE.priceId,
    },
  ];

  return (
    <motion.section
      id="agent-plans"
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.6, ease: 'easeOut' }}
      className="pricing-agent-section mb-24"
    >
      <div className={compact ? 'text-center mb-8' : 'text-center mb-12'}>
        {!compact && (
          <>
            <div className="relative inline-block mb-4">
              <div className="px-4 py-1.5 rounded-full bg-gradient-to-r from-cyan-500/10 to-blue-500/10 border border-cyan-500/20">
                <span className="text-sm font-medium text-cyan-400">
                  Trust-aware Agent Execution
                </span>
              </div>
            </div>
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              Trust-aligned agent tiers
            </h2>
          </>
        )}
        <p className="text-text-secondary max-w-2xl mx-auto text-lg">
          Agents need tools they can trust. These tiers align verification, audit retention, policy
          packs, and Trust API usage with your execution scale.
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
          All tiers include: trust-aware discovery metadata • verification gates • auditable
          execution context (with Trust API access scoped by plan).
        </p>
      </div>
    </motion.section>
  );
}
