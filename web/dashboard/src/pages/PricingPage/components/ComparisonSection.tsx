import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { PLANS } from "@/lib/constants";
import { Link } from "react-router-dom";
import { useState } from "react";
import { BottomSheet } from "react-spring-bottom-sheet";
import { Progress } from "@/components/ui/progress";
import { comparisonFeatures } from "../data";
import { useScrollAnimation } from "../hooks";
import toast from "react-hot-toast";

interface ComparisonSectionProps {
  onPlanSelect?: (planId: string) => void;
}

// Comparison Section Component with scroll animations
export function ComparisonSection({ onPlanSelect }: ComparisonSectionProps) {
  const { ref, inView } = useScrollAnimation(0.1, false);
  const [showBottomSheet, setShowBottomSheet] = useState(false);

  const handlePlanSelect = (planId: string) => {
    setShowBottomSheet(false);
    onPlanSelect?.(planId);
    // This will be passed from parent component, but for now we'll use toast
    toast.success(`Selected ${planId} plan!`, {
      duration: 3000,
      style: {
        background: '#1a1a1a',
        color: '#fff',
        border: '1px solid #6366f1',
      },
    });
  };

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.8, ease: "easeOut" }}
      className="pricing-comparison-section mb-20"
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.6, delay: 0.2 }}
        className="text-center mb-12"
      >
        <h2 className="text-3xl font-bold text-white mb-4">Compare Plans</h2>
        <p className="text-text-secondary max-w-2xl mx-auto">
          Detailed comparison of features and limits across all plans
        </p>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
        transition={{ duration: 0.6, delay: 0.4 }}
      >
        <Card className="pricing-comparison-table border-white/8 bg-white/5 overflow-hidden">
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-white/8">
                    <th className="text-left p-6 text-white font-semibold">Features</th>
                    <th className="text-center p-6 text-white font-semibold">Free</th>
                    <th className="text-center p-6 text-white font-semibold">Starter</th>
                    <th className="text-center p-6 text-white font-semibold">Professional</th>
                    <th className="text-center p-6 text-white font-semibold">Enterprise</th>
                  </tr>
                </thead>
                <tbody>
                    {comparisonFeatures.map((item, index) => {
                      const tooltipId = `comparison-${index}`;
                      const getFeatureTooltip = (feature: string) => {
                        const tooltips: { [key: string]: string } = {
                          "Functions": "The number of serverless functions you can deploy",
                          "Providers": "Cloud platforms where you can deploy your functions",
                          "Monthly Requests": "Total function invocations allowed per month",
                          "Custom Domains": "Number of custom domains you can connect",
                          "SLA": "Service Level Agreement uptime guarantee",
                          "Support": "Level of customer support included",
                          "Analytics": "Depth and customization of usage analytics",
                          "Team Collaboration": "Ability to invite team members and collaborate",
                        };
                        return tooltips[feature] || "";
                      };

                      const getProgressValue = (value: string, feature: string) => {
                        if (feature === "Monthly Requests") {
                          const numValue = value.replace(/[^0-9]/g, '');
                          if (value === "Unlimited") return 100;
                          if (numValue) {
                            const numeric = parseInt(numValue);
                            if (feature === "Monthly Requests") {
                              // Scale to show relative values (100K = 10%, 1M = 30%, 10M = 70%, Unlimited = 100%)
                              if (numeric <= 100) return 10;
                              if (numeric <= 1000) return 30;
                              if (numeric <= 10000) return 70;
                              return 100;
                            }
                          }
                        }
                        return 0;
                      };

                      const getValueColor = (value: string) => {
                        if (value === "Unlimited" || value === "All" || value === "Custom") {
                          return "text-emerald-400";
                        }
                        if (value === "None") return "text-red-400";
                        return "text-text-secondary";
                      };

                      return (
                        <motion.tr
                          key={item.feature}
                          className={`border-b border-white/4 ${index % 2 === 0 ? "bg-white/2" : ""}`}
                          initial={{ opacity: 0, x: -20 }}
                          animate={inView ? { opacity: 1, x: 0 } : { opacity: 0, x: -20 }}
                          transition={{ duration: 0.5, delay: 0.6 + index * 0.1 }}
                        >
                          <td
                            className="p-6 text-white font-medium hover:text-[#6366f1] transition-colors cursor-help"
                            data-tooltip-id={tooltipId}
                            data-tooltip-content={getFeatureTooltip(item.feature)}
                          >
                            {item.feature}
                          </td>
                          <td className="p-6 text-center">
                            <div className="flex flex-col items-center gap-2">
                              <span className={getValueColor(item.free)}>{item.free}</span>
                              {getProgressValue(item.free, item.feature) > 0 && (
                                <Progress
                                  value={getProgressValue(item.free, item.feature)}
                                  className="w-16 h-1"
                                />
                              )}
                            </div>
                          </td>
                          <td className="p-6 text-center">
                            <div className="flex flex-col items-center gap-2">
                              <span className={getValueColor(item.starter)}>{item.starter}</span>
                              {getProgressValue(item.starter, item.feature) > 0 && (
                                <Progress
                                  value={getProgressValue(item.starter, item.feature)}
                                  className="w-16 h-1"
                                />
                              )}
                            </div>
                          </td>
                          <td className="p-6 text-center">
                            <div className="flex flex-col items-center gap-2">
                              <span className={getValueColor(item.professional)}>{item.professional}</span>
                              {getProgressValue(item.professional, item.feature) > 0 && (
                                <Progress
                                  value={getProgressValue(item.professional, item.feature)}
                                  className="w-16 h-1"
                                />
                              )}
                            </div>
                          </td>
                          <td className="p-6 text-center">
                            <div className="flex flex-col items-center gap-2">
                              <span className={getValueColor(item.enterprise)}>{item.enterprise}</span>
                              {getProgressValue(item.enterprise, item.feature) > 0 && (
                                <Progress
                                  value={getProgressValue(item.enterprise, item.feature)}
                                  className="w-16 h-1"
                                />
                              )}
                            </div>
                          </td>
                        </motion.tr>
                      );
                    })}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* Bottom Sheet for Detailed Comparison */}
      <BottomSheet
        open={showBottomSheet}
        onDismiss={() => setShowBottomSheet(false)}
        snapPoints={({ maxHeight }) => [maxHeight * 0.8, maxHeight * 0.4]}
        defaultSnap={({ snapPoints }) => snapPoints[1]}
        header={
          <div className="text-center py-4">
            <h3 className="text-xl font-bold text-white">Detailed Plan Comparison</h3>
            <p className="text-text-secondary text-sm">All features and limitations at a glance</p>
          </div>
        }
        className="!bg-black !border-t !border-white/20"
      >
        <div className="p-6 space-y-6">
          {/* Detailed comparison content */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {Object.values(PLANS).map((plan) => (
              <motion.div
                key={plan.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3 }}
                className="bg-white/5 rounded-lg p-4 border border-white/8"
              >
                <h4 className="text-lg font-semibold text-white mb-3">{plan.name}</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-text-secondary">Price:</span>
                    <span className="text-white font-medium">
                      {plan.price === "Custom" ? "Custom" : `$${plan.price}/mo`}
                    </span>
                  </div>
                  {comparisonFeatures.map((feature) => (
                    <div key={feature.feature} className="flex justify-between">
                        <span className="text-text-secondary">{feature.feature}:</span>
                        <span className="text-white">
                          {feature[plan.id.toLowerCase() as keyof typeof feature]}
                        </span>
                    </div>
                  ))}
                </div>
                <Link
                  to={plan.id === "enterprise" ? "/contact" : "/signup"}
                  className="block mt-4"
                  onClick={() => {
                    setShowBottomSheet(false);
                    handlePlanSelect(plan.id);
                  }}
                >
                  <Button
                    variant={plan.id === "free" ? "outline" : "default"}
                    className="w-full"
                    size="sm"
                  >
                    {plan.id === "enterprise"
                      ? "Contact Sales"
                      : plan.id === "free"
                        ? "Start Free"
                        : "Start Trial"}
                  </Button>
                </Link>
              </motion.div>
            ))}
          </div>

          {/* Additional details */}
          <div className="bg-[#6366f1]/5 border border-[#6366f1]/20 rounded-lg p-4">
            <h4 className="text-lg font-semibold text-white mb-2">💡 Pro Tips</h4>
            <ul className="text-text-secondary text-sm space-y-1">
              <li>• All paid plans include a 14-day free trial</li>
              <li>• Enterprise plans are fully customizable</li>
              <li>• Monthly billing with no long-term contracts</li>
              <li>• Cancel anytime, no hidden fees</li>
            </ul>
          </div>
        </div>
      </BottomSheet>
    </motion.div>
  );
}
