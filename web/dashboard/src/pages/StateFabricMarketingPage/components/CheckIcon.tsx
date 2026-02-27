import { Check } from "lucide-react";
import { motion } from "framer-motion";

interface CheckIconProps {
  className?: string;
  size?: number;
}

export function CheckIcon({ className = "", size = 20 }: CheckIconProps) {
  return (
    <motion.div
      className={`inline-flex items-center justify-center rounded-full bg-green-100 dark:bg-green-900/20 ${className}`}
      initial={{ scale: 0 }}
      animate={{ scale: 1 }}
      transition={{ type: "spring", stiffness: 400, damping: 17 }}
    >
      <Check className={`text-green-600 dark:text-green-400`} size={size} />
    </motion.div>
  );
}