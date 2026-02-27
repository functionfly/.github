import { motion } from "framer-motion";
import { ReactNode } from "react";

interface FadeInOnScrollProps {
  children: ReactNode;
  delay?: number;
  className?: string;
}

const fadeInUp = {
  initial: { opacity: 0, y: 30 },
  whileInView: { opacity: 1, y: 0 },
  transition: { duration: 0.6 },
  viewport: { once: true }
};

export function FadeInOnScroll({ children, delay = 0, className = "" }: FadeInOnScrollProps) {
  return (
    <motion.div
      className={className}
      initial={fadeInUp.initial}
      whileInView={fadeInUp.whileInView}
      transition={{ ...fadeInUp.transition, delay }}
      viewport={fadeInUp.viewport}
    >
      {children}
    </motion.div>
  );
}