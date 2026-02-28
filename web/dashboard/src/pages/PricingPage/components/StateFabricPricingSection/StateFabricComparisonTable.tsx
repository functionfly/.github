import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { useScrollAnimation } from "../../hooks";
import { STATE_FABRIC_COMPARISON_ROWS } from "./data";

function getValueColor(value: string) {
  if (value === "Unlimited" || value === "Yes" || value === "Yes + edge" || value === "Dedicated" || value === "Export" || value === "Full") {
    return "text-emerald-400";
  }
  if (value === "—" || value === "None") return "text-text-secondary";
  return "text-text-secondary";
}

export function StateFabricComparisonTable() {
  const { ref, inView } = useScrollAnimation(0.1, false);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5 }}
      className="mt-16"
    >
      <h3 className="text-xl font-bold text-white mb-4 text-center">Compare State Fabric plans</h3>
      <Card className="pricing-state-fabric-comparison border-white/8 bg-white/5 overflow-hidden">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/8">
                  <th className="text-left p-4 text-white font-semibold text-sm">Feature</th>
                  <th className="text-center p-4 text-white font-semibold text-sm">Sandbox</th>
                  <th className="text-center p-4 text-white font-semibold text-sm">Starter</th>
                  <th className="text-center p-4 text-white font-semibold text-sm">Pro</th>
                  <th className="text-center p-4 text-white font-semibold text-sm">Business</th>
                  <th className="text-center p-4 text-white font-semibold text-sm">Enterprise</th>
                </tr>
              </thead>
              <tbody>
                {STATE_FABRIC_COMPARISON_ROWS.map((row, index) => (
                  <motion.tr
                    key={row.feature}
                    className={`border-b border-white/4 ${index % 2 === 0 ? "bg-white/2" : ""}`}
                    initial={{ opacity: 0, x: -10 }}
                    animate={inView ? { opacity: 1, x: 0 } : { opacity: 0, x: -10 }}
                    transition={{ delay: index * 0.05 }}
                  >
                    <td className="p-4 text-white font-medium text-sm">{row.feature}</td>
                    <td className="p-4 text-center text-sm">
                      <span className={getValueColor(row.sandbox)}>{row.sandbox}</span>
                    </td>
                    <td className="p-4 text-center text-sm">
                      <span className={getValueColor(row.starter)}>{row.starter}</span>
                    </td>
                    <td className="p-4 text-center text-sm">
                      <span className={getValueColor(row.pro)}>{row.pro}</span>
                    </td>
                    <td className="p-4 text-center text-sm">
                      <span className={getValueColor(row.business)}>{row.business}</span>
                    </td>
                    <td className="p-4 text-center text-sm">
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
