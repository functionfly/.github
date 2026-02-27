import { motion } from "framer-motion";

interface MetricBadgeProps {
  value: string | number;
  label: string;
  className?: string;
  variant?: "default" | "success" | "warning" | "info";
}

const variants = {
  default: "metric-badge-default bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800",
  success: "metric-badge-success bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 border-green-200 dark:border-green-800",
  warning: "metric-badge-warning bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-300 border-yellow-200 dark:border-yellow-800",
  info: "metric-badge-info bg-purple-50 dark:bg-purple-900/20 text-purple-700 dark:text-purple-300 border-purple-200 dark:border-purple-800",
};

export function MetricBadge({
  value,
  label,
  className = "",
  variant = "default"
}: MetricBadgeProps) {
  return (
    <motion.div
      className={`inline-flex items-center gap-2 px-3 py-1 rounded-full border text-sm font-medium ${variants[variant]} ${className}`}
      whileHover={{ scale: 1.02 }}
      transition={{ type: "spring", stiffness: 300, damping: 20 }}
    >
      <span className="font-semibold">{value}</span>
      <span className="opacity-90">{label}</span>
    </motion.div>
  );
}