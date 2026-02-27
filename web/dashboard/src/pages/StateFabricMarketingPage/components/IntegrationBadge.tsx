import { motion } from "framer-motion";
import { ArrowRight } from "lucide-react";

interface IntegrationBadgeProps {
  title: string;
  description: string;
  icon?: React.ReactNode;
  className?: string;
}

export function IntegrationBadge({
  title,
  description,
  icon,
  className = ""
}: IntegrationBadgeProps) {
  return (
    <motion.div
      className={`bg-bg-secondary p-4 rounded-lg border border-border hover:border-border/80 transition-all duration-300 ${className}`}
      whileHover={{ scale: 1.02, y: -2 }}
      transition={{ type: "spring", stiffness: 300, damping: 20 }}
    >
      <div className="flex items-center gap-3 mb-2">
        {icon && (
          <div className="w-8 h-8 rounded-lg glass-light flex items-center justify-center shrink-0">
            {icon}
          </div>
        )}
        <h4 className="text-lg font-semibold text-slate-900 dark:text-white">
          {title}
        </h4>
        <ArrowRight className="w-4 h-4 text-slate-400 ml-auto" />
      </div>
      <p className="text-slate-600 dark:text-text-secondary text-sm">
        {description}
      </p>
    </motion.div>
  );
}