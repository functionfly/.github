import { motion } from "framer-motion";
import { ReactNode } from "react";

interface HoverGlowCardProps {
  children: ReactNode;
  className?: string;
}

export function HoverGlowCard({ children, className = "" }: HoverGlowCardProps) {
  return (
    <motion.div
      className={`bg-bg-secondary p-6 rounded-lg border border-border shadow-sm hover:shadow-lg hover:border-border/80 transition-all duration-300 ${className}`}
      whileHover={{ scale: 1.02 }}
      transition={{ type: "spring", stiffness: 300, damping: 20 }}
    >
      {children}
    </motion.div>
  );
}