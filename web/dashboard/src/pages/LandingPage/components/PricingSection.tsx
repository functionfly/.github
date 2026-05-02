import { motion } from "framer-motion";
import { Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { PLANS } from "@/lib/constants";
import { Link } from "react-router-dom";

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(" ");
}

export function PricingSection() {
  return (
    <section id="pricing" className="py-20 border-t border-border-subtle mesh-gradient-bg pricing-section-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold text-text-primary mb-4" style={{ color: 'var(--text-primary)', fontWeight: 800 }}>
            Simple, transparent pricing
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto">
            Start free, scale as you grow. No hidden fees.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 max-w-7xl mx-auto">
          {Object.values(PLANS).map((plan, index) => (
            <motion.div
              key={plan.id}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Card
                className={cn(
                  "h-full card-elevation glass-card shine-effect",
                  plan.id === "professional" &&
                    "gradient-border-static relative glow",
                )}
              >
                {plan.id === "professional" && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                    <span className="px-3 py-1 rounded-full bg-gradient-to-r from-brand-500 to-purple-500 text-white text-xs font-medium">
                      Most Popular
                    </span>
                  </div>
                )}
                <CardContent className="p-6">
                  <div className="mb-6">
                    <h3 className="text-xl font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 700 }}>
                      {plan.name}
                    </h3>
                    <p className="text-text-secondary text-sm mb-4">
                      {plan.description}
                    </p>
                    <div className="flex items-baseline gap-1">
                      <span className="text-4xl font-bold text-text-primary" style={{ color: 'var(--text-primary)' }}>
                        {`$${(plan as { price: number }).price}`}
                      </span>
                      {((plan as { price: number }).price) > 0 && (
                        <span className="text-text-secondary">/month</span>
                      )}
                    </div>
                  </div>

                  <ul className="space-y-3 mb-6">
                    {plan.features.map((feature) => (
                      <li
                        key={feature}
                        className="flex items-center gap-3 text-sm"
                      >
                        <div className="w-5 h-5 rounded-full bg-success/20 flex items-center justify-center">
                          <Check className="w-3 h-3 text-success" />
                        </div>
                        <span className="text-text-secondary">{feature}</span>
                      </li>
                    ))}
                  </ul>

                  <Link
                    to={plan.id === "enterprise" ? "/contact" : "/signup"}
                    className="block"
                  >
                    <Button
                      variant={plan.id === "free" ? "outline" : "default"}
                      className="w-full hover-lift"
                    >
                      {plan.id === "enterprise"
                        ? "Contact Sales"
                        : plan.id === "free"
                          ? "Start Free"
                          : "Get Started"}
                    </Button>
                  </Link>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}