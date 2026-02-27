import { motion } from "framer-motion";
import { CheckIcon } from "./CheckIcon";
import { CrossIcon } from "./CrossIcon";

interface ComparisonItem {
  feature: string;
  diy: boolean | string;
  stateFabric: boolean | string;
}

interface ComparisonTableProps {
  items: ComparisonItem[];
  className?: string;
}

export function ComparisonTable({ items, className = "" }: ComparisonTableProps) {
  return (
    <motion.div
      className={`bg-bg-secondary rounded-lg border border-border overflow-hidden ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      viewport={{ once: true }}
    >
      <div className="grid grid-cols-3 gap-0">
        {/* Header */}
        <div className="bg-bg-primary p-4 border-b border-border">
          <h3 className="font-semibold text-slate-900 dark:text-white">Feature</h3>
        </div>
        <div className="bg-red-50 dark:bg-red-900/10 p-4 border-b border-border border-l">
          <h3 className="font-semibold text-red-700 dark:text-red-300">DIY Infra</h3>
        </div>
        <div className="bg-green-50 dark:bg-green-900/10 p-4 border-b border-border border-l">
          <h3 className="font-semibold text-green-700 dark:text-green-300">StateFabric</h3>
        </div>

        {/* Rows */}
        {items.map((item, index) => (
          <motion.div
            key={index}
            className="contents"
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.4, delay: index * 0.1 }}
            viewport={{ once: true }}
          >
            <div className={`p-4 border-b border-border ${index % 2 === 0 ? 'bg-bg-primary/50' : ''}`}>
              <span className="text-slate-700 dark:text-slate-300 font-medium">
                {item.feature}
              </span>
            </div>
            <div className={`p-4 border-b border-border border-l ${index % 2 === 0 ? 'bg-red-50/50 dark:bg-red-900/5' : ''}`}>
              {typeof item.diy === 'boolean' ? (
                item.diy ? (
                  <CheckIcon size={16} />
                ) : (
                  <CrossIcon size={16} />
                )
              ) : (
                <span className="text-slate-600 dark:text-slate-400 text-sm">
                  {item.diy}
                </span>
              )}
            </div>
            <div className={`p-4 border-b border-border border-l ${index % 2 === 0 ? 'bg-green-50/50 dark:bg-green-900/5' : ''}`}>
              {typeof item.stateFabric === 'boolean' ? (
                item.stateFabric ? (
                  <CheckIcon size={16} />
                ) : (
                  <CrossIcon size={16} />
                )
              ) : (
                <span className="text-slate-600 dark:text-slate-400 text-sm">
                  {item.stateFabric}
                </span>
              )}
            </div>
          </motion.div>
        ))}
      </div>
    </motion.div>
  );
}