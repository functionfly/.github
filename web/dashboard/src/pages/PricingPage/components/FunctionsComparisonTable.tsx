import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { comparisonFeatures } from "../data";
import { useScrollAnimation } from "../hooks";

function getValueColor(value: string) {
  if (value === "Unlimited" || value === "All" || value === "Custom") return "text-emerald-400";
  if (value === "None") return "text-red-400";
  return "text-text-secondary";
}

/**
 * Condensed comparison table for Function Deployment plans only.
 * Shown under the function plan cards to avoid a separate "Compare Plans" section.
 */
export function FunctionsComparisonTable() {
  const { ref, inView } = useScrollAnimation(0.1, false);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5 }}
      className="mt-12 pricing-comparison-section"
    >
      <p className="text-center text-text-secondary text-base mb-5">
        Detailed comparison of features and limits
      </p>
      <Card className="pricing-comparison-table border-white/8 bg-white/5 overflow-hidden">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/8">
                  <th className="text-left p-5 text-white font-semibold text-base">Feature</th>
                  <th className="text-center p-5 text-white font-semibold text-base">Free</th>
                  <th className="text-center p-5 text-white font-semibold text-base">Starter</th>
                  <th className="text-center p-5 text-white font-semibold text-base">Professional</th>
                  <th className="text-center p-5 text-white font-semibold text-base">Enterprise</th>
                </tr>
              </thead>
              <tbody>
                {comparisonFeatures.map((row, index) => (
                  <motion.tr
                    key={row.feature}
                    className={`border-b border-white/4 ${index % 2 === 0 ? "bg-white/2" : ""}`}
                    initial={{ opacity: 0, x: -10 }}
                    animate={inView ? { opacity: 1, x: 0 } : { opacity: 0, x: -10 }}
                    transition={{ delay: index * 0.03 }}
                  >
                    <td className="p-5 text-white font-medium text-base">{row.feature}</td>
                    <td className="p-5 text-center text-base">
                      <span className={getValueColor(row.free)}>{row.free}</span>
                    </td>
                    <td className="p-5 text-center text-base">
                      <span className={getValueColor(row.starter)}>{row.starter}</span>
                    </td>
                    <td className="p-5 text-center text-base">
                      <span className={getValueColor(row.professional)}>{row.professional}</span>
                    </td>
                    <td className="p-5 text-center text-base">
                      <span className={getValueColor(row.enterprise)}>{row.enterprise}</span>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
