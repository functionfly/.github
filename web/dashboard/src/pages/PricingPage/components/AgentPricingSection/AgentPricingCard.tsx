import { motion } from "framer-motion";
import { Check, Star, Bot } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router-dom";

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(" ");
}

interface AgentPlan {
  id: string;
  name: string;
  tagline?: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  highlighted: boolean;
  cta: string;
  href: string;
}

interface AgentPricingCardProps {
  plan: AgentPlan;
  index: number;
  inView: boolean;
}

export function AgentPricingCard({ plan, index, inView }: AgentPricingCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 30 }}
      transition={{ duration: 0.5, delay: index * 0.05 }}
    >
      <Card
        className={cn(
          "pricing-agent-card h-full relative overflow-hidden transition-all duration-300",
          "bg-gradient-to-br from-cyan-950/30 to-blue-950/30 backdrop-blur-sm",
          "border border-cyan-500/20 hover:border-cyan-500/40",
          "hover:shadow-2xl hover:shadow-cyan-500/10",
          plan.highlighted && "border-cyan-500/50 ring-2 ring-cyan-500/20"
        )}
      >
        {plan.highlighted && (
          <div className="pricing-agent-badge absolute -top-3 left-1/2 -translate-x-1/2 z-10">
            <span className="px-3 py-1.5 rounded-full bg-gradient-to-r from-cyan-500 to-blue-500 text-white text-xs font-semibold flex items-center gap-1.5 shadow-lg shadow-cyan-500/25">
              <Star className="w-3.5 h-3.5 fill-current" />
              Most Popular
            </span>
          </div>
        )}
        
        {/* Bot icon header */}
        <div className="absolute top-0 right-0 w-24 h-24 opacity-10">
          <Bot className="w-full h-full text-cyan-400" />
        </div>
        
        <CardContent className="p-5 md:p-6 relative z-10">
          <div className="mb-3">
            <div className="flex items-center gap-2 mb-1">
              <Bot className="w-5 h-5 text-cyan-400" />
              <h3 className="text-xl font-bold text-white">{plan.name}</h3>
            </div>
            {plan.tagline && (
              <p className="text-cyan-300/70 text-sm">{plan.tagline}</p>
            )}
          </div>
          <p className="text-text-secondary text-sm mb-4 leading-snug">{plan.description}</p>
          <div className="flex items-baseline gap-1.5 mb-4">
            <span className="pricing-agent-price text-3xl font-bold text-white">{plan.price}</span>
            {plan.period !== "pricing" && (
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
                <div className="w-5 h-5 rounded-full bg-gradient-to-br from-cyan-500/20 to-blue-500/20 border border-cyan-500/30 flex items-center justify-center mt-0.5 shrink-0">
                  <Check className="w-3 h-3 text-cyan-400" />
                </div>
                <span className="text-text-secondary text-sm leading-snug">{feature}</span>
              </motion.li>
            ))}
          </ul>
          <Link to={plan.href}>
            <Button
              variant={plan.highlighted ? "default" : "outline"}
              size="lg"
              className={cn(
                "w-full py-3 text-sm font-semibold",
                plan.highlighted &&
                  "bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-600 hover:to-blue-600",
                !plan.highlighted && "border-cyan-500/30 text-cyan-400 hover:bg-cyan-500/10"
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
