import { motion } from "framer-motion";
import { ArrowRight } from "lucide-react";

interface AnimatedFlowArrowProps {
  delay?: number;
  className?: string;
}

export function AnimatedFlowArrow({ delay = 0, className = "" }: AnimatedFlowArrowProps) {
  return (
    <motion.div
      className={`flex items-center justify-center ${className}`}
      initial={{ opacity: 0, scale: 0.8 }}
      whileInView={{ opacity: 1, scale: 1 }}
      transition={{ delay, duration: 0.5 }}
      viewport={{ once: true }}
    >
      <motion.div
        animate={{
          x: [0, 10, 0],
        }}
        transition={{
          duration: 2,
          repeat: Infinity,
          ease: "easeInOut"
        }}
      >
        <ArrowRight className="w-8 h-8 text-blue-500" />
      </motion.div>
    </motion.div>
  );
}