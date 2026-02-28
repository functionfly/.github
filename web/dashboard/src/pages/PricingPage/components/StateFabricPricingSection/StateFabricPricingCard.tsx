import { motion } from "framer-motion";
import { Check, Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router-dom";
import type { StateFabricPlan } from "./data";

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(" ");
}

interface StateFabricPricingCardProps {
  plan: StateFabricPlan;
  index: number;
  inView: boolean;
}

export function StateFabricPricingCard({ plan, index, inView }: StateFabricPricingCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
      transition={{ duration: 0.5, delay: index * 0.05 }}
    >
      <Card
        className={cn(
          "pricing-state-fabric-card h-full relative overflow-hidden transition-all duration-300",
          "bg-gradient-to-br from-white/5 to-white/10 backdrop-blur-sm",
          "border border-white/10 hover:border-white/20",
          "hover:shadow-2xl hover:shadow-[#6366f1]/10",
          plan.highlighted && "border-[#6366f1]/50 ring-1 ring-[#6366f1]/20"
        )}
      >
        {plan.highlighted && (
          <div className="pricing-state-fabric-badge absolute -top-3 left-1/2 -translate-x-1/2 z-10">
            <span className="px-3 py-1.5 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] text-white text-xs font-semibold flex items-center gap-1.5 shadow-lg shadow-[#6366f1]/25">
              <Star className="w-3.5 h-3.5 fill-current" />
              Most Popular
            </span>
          </div>
        )}
        <CardContent className="p-5 md:p-6 relative z-10">
          <div className="mb-3">
            <h3 className="text-xl font-bold text-white">{plan.name}</h3>
            {plan.tagline && (
              <p className="text-text-secondary text-sm mt-0.5">{plan.tagline}</p>
            )}
          </div>
          <p className="text-text-secondary text-sm mb-4 leading-snug">{plan.description}</p>
          <div className="flex items-baseline gap-1.5 mb-4">
            <span className="pricing-state-fabric-price text-3xl font-bold text-white">{plan.price}</span>
            {plan.period && (
              <span className="text-text-secondary text-sm">/{plan.period}</span>
            )}
          </div>
          <ul className="space-y-2.5 mb-4">
            {plan.features.map((feature, i) => (
              <motion.li
                key={feature}
                className="flex items-start gap-2.5"
                initial={{ opacity: 0, x: -8 }}
                animate={inView ? { opacity: 1, x: 0 } : { opacity: 0, x: -8 }}
                transition={{ delay: 0.1 + index * 0.05 + i * 0.03 }}
              >
                <div className="w-5 h-5 rounded-full bg-gradient-to-br from-emerald-500/20 to-green-500/20 border border-emerald-500/30 flex items-center justify-center mt-0.5 shrink-0">
                  <Check className="w-3 h-3 text-emerald-400" />
                </div>
                <span className="text-text-secondary text-sm leading-snug">{feature}</span>
              </motion.li>
            ))}
          </ul>
          {plan.addOns && (
            <p className="text-text-secondary/80 text-xs mb-4 border-t border-white/5 pt-3">
              Add-ons: {plan.addOns}
            </p>
          )}
          <Link to={plan.href}>
            <Button
              variant={plan.highlighted ? "default" : "outline"}
              size="lg"
              className={cn(
                "w-full py-3 text-sm font-semibold",
                plan.highlighted &&
                  "bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90",
                !plan.highlighted && "border-white/30 hover:bg-white/10"
              )}
            >
              {plan.cta}
            </Button>
          </Link>
        </CardContent>
      </Card>
    </motion.div>
  );
}
