import { motion } from "framer-motion";
import { useScrollAnimation } from "../../hooks";
import { STATE_FABRIC_PLANS } from "./data";
import { StateFabricPricingCard } from "./StateFabricPricingCard";
import { StateFabricComparisonTable } from "./StateFabricComparisonTable";
import { StateFabricAddOnsSection } from "./StateFabricAddOnsSection";

interface StateFabricPricingSectionProps {
  /** When true, use a compact header (e.g. when shown inside pricing tabs). */
  compact?: boolean;
  /** Callback when a plan is selected for upgrade/checkout */
  onPlanSelect?: (planId: string, priceId?: string) => void;
}

/**
 * State Fabric pricing section for the main /pricing page.
 * Kept as a separate component for easy updates and reuse.
 */
export function StateFabricPricingSection({ compact, onPlanSelect }: StateFabricPricingSectionProps) {
  const { ref, inView } = useScrollAnimation(0.1, false);

  return (
    <motion.section
      id="state-fabric"
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.6, ease: "easeOut" }}
      className="pricing-state-fabric-section"
    >
      <div className={compact ? "text-center mb-8" : "text-center mb-12"}>
        {!compact && (
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">State Fabric</h2>
        )}
        <p className="text-text-secondary max-w-2xl mx-auto text-lg">
          Stateful capabilities for serverless. Start free, scale with usage—from Sandbox to
          Enterprise.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-6 max-w-7xl mx-auto pt-6">
        {STATE_FABRIC_PLANS.map((plan, index) => (
          <StateFabricPricingCard 
            key={plan.id} 
            plan={plan} 
            index={index} 
            inView={inView}
            onPlanSelect={onPlanSelect}
          />
        ))}
      </div>

      <StateFabricComparisonTable />
      <StateFabricAddOnsSection />
    </motion.section>
  );
}
