import { motion } from "framer-motion";
import { ReactNode } from "react";

interface LayerCardProps {
  title: string;
  subtitle?: string;
  icon?: ReactNode;
  className?: string;
  delay?: number;
}

export function LayerCard({ title, subtitle, icon, className = "", delay = 0 }: LayerCardProps) {
  return (
    <motion.div
      className={`bg-bg-secondary p-6 text-center rounded-lg border border-border shadow-sm hover:shadow-md transition-shadow ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.6 }}
      viewport={{ once: true }}
    >
      {icon && (
        <div className="w-12 h-12 rounded-lg bg-slate-100 dark:bg-slate-700 flex items-center justify-center mb-4 mx-auto animate-float">
          {icon}
        </div>
      )}
      <h3 className="text-lg font-semibold text-slate-900 dark:text-white layer-card-title mb-2">{title}</h3>
      {subtitle && (
        <p className="text-sm text-slate-600 dark:text-slate-400 layer-card-subtitle">{subtitle}</p>
      )}
    </motion.div>
  );
}