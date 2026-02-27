import { X } from "lucide-react";
import { motion } from "framer-motion";

interface CrossIconProps {
  className?: string;
  size?: number;
}

export function CrossIcon({ className = "", size = 20 }: CrossIconProps) {
  return (
    <motion.div
      className={`inline-flex items-center justify-center rounded-full bg-red-100 dark:bg-red-900/20 ${className}`}
      initial={{ scale: 0 }}
      animate={{ scale: 1 }}
      transition={{ type: "spring", stiffness: 400, damping: 17 }}
    >
      <X className={`text-red-600 dark:text-red-400`} size={size} />
    </motion.div>
  );
}