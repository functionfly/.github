import { motion } from "framer-motion";
import { Database } from "lucide-react";

const fadeInUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.6 }
};

export function GradientHeadline() {
  return (
    <>
      <motion.div
        variants={fadeInUp}
        className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white/90 dark:bg-slate-900/90 border border-blue-300 dark:border-blue-600 text-slate-800 dark:text-white text-sm font-semibold mb-8 animate-pulse-scale glow-sm backdrop-blur-sm shadow-lg"
      >
        <Database className="w-4 h-4 text-blue-600 dark:text-blue-400" />
        StateFabric
      </motion.div>

      <motion.h1
        variants={fadeInUp}
        className="hero-title-main text-4xl lg:text-6xl font-bold mb-6 animate-fade-in-up text-slate-900 dark:text-white"
      >
        Deterministic State for Autonomous Systems
      </motion.h1>
    </>
  );
}