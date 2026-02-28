import { motion } from "framer-motion";
import { Zap, Shield, Brain, BarChart3 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { useScrollAnimation } from "../../hooks";
import { STATE_FABRIC_ADDONS } from "./data";

const ADDON_ICONS = [Zap, Shield, Brain, BarChart3];

export function StateFabricAddOnsSection() {
  const { ref, inView } = useScrollAnimation(0.1, false);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5 }}
      className="mt-16"
    >
      <h3 className="text-xl font-bold text-white mb-4 text-center">Optional add-ons</h3>
      <p className="text-text-secondary text-sm text-center max-w-2xl mx-auto mb-6">
        Enhance any plan with performance, security, or analytics add-ons.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 max-w-6xl mx-auto">
        {STATE_FABRIC_ADDONS.map((addon, index) => {
          const Icon = ADDON_ICONS[index % ADDON_ICONS.length];
          return (
            <motion.div
              key={addon.name}
              initial={{ opacity: 0, y: 10 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
              transition={{ delay: index * 0.08 }}
            >
              <Card className="pricing-state-fabric-addon border-white/8 bg-white/5 h-full">
                <CardContent className="p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Icon className="w-5 h-5 text-[#6366f1] shrink-0" />
                    <h4 className="text-base font-semibold text-white">{addon.name}</h4>
                  </div>
                  <div className="flex items-baseline gap-1 text-white font-bold mb-1">
                    <span className="pricing-state-fabric-price">{addon.price}</span>
                    <span className="text-text-secondary text-sm font-normal">{addon.period}</span>
                  </div>
                  <p className="text-text-secondary text-sm">{addon.description}</p>
                </CardContent>
              </Card>
            </motion.div>
          );
        })}
      </div>
    </motion.div>
  );
}
