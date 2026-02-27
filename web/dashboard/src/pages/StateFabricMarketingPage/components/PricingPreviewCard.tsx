import { motion } from "framer-motion";
import { Check } from "lucide-react";

interface PricingPreviewCardProps {
  title: string;
  price: string;
  period?: string;
  description: string;
  features: string[];
  highlighted?: boolean;
  className?: string;
}

export function PricingPreviewCard({
  title,
  price,
  period = "month",
  description,
  features,
  highlighted = false,
  className = ""
}: PricingPreviewCardProps) {
  return (
    <motion.div
      className={`
        bg-bg-secondary rounded-xl p-6 border transition-all duration-300
        ${highlighted
          ? 'border-blue-500 shadow-lg shadow-blue-500/20'
          : 'border-border hover:border-border/80'
        }
        ${className}
      `}
      whileHover={{ scale: highlighted ? 1.02 : 1.01, y: -2 }}
      transition={{ type: "spring", stiffness: 300, damping: 20 }}
    >
      {highlighted && (
        <div className="bg-blue-500 text-white text-xs font-semibold px-3 py-1 rounded-full inline-block mb-4">
          Most Popular
        </div>
      )}

      <div className="mb-4">
        <h3 className="text-xl font-bold text-slate-900 dark:text-white mb-2">
          {title}
        </h3>
        <div className="flex items-baseline gap-1">
          <span className="text-3xl font-bold text-slate-900 dark:text-white">
            {price}
          </span>
          {period && (
            <span className="text-slate-600 dark:text-text-secondary">
              /{period}
            </span>
          )}
        </div>
        <p className="text-slate-600 dark:text-text-secondary text-sm mt-2">
          {description}
        </p>
      </div>

      <ul className="space-y-3">
        {features.map((feature, index) => (
          <motion.li
            key={index}
            className="flex items-start gap-3"
            initial={{ opacity: 0, x: -10 }}
            whileInView={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.3, delay: index * 0.1 }}
            viewport={{ once: true }}
          >
            <Check className="w-5 h-5 text-green-500 shrink-0 mt-0.5" />
            <span className="text-slate-700 dark:text-slate-300 text-sm">
              {feature}
            </span>
          </motion.li>
        ))}
      </ul>
    </motion.div>
  );
}