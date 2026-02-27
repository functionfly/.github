import { motion } from "framer-motion";
import { AnimatedBackgroundGrid } from "./AnimatedBackgroundGrid";
import { GradientHeadline } from "./GradientHeadline";
import { PrimaryCTA } from "./PrimaryCTA";
import { SecondaryCTA } from "./SecondaryCTA";

const fadeInUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.6 }
};

const stagger = {
  animate: {
    transition: {
      staggerChildren: 0.1
    }
  }
};

export function HeroSection() {
  return (
    <AnimatedBackgroundGrid>
      <motion.div
        className="text-center max-w-4xl mx-auto"
        initial="initial"
        animate="animate"
        variants={stagger}
      >
        <GradientHeadline />

        <motion.p
          variants={fadeInUp}
          className="hero-subtitle text-xl mb-8 max-w-2xl mx-auto animate-fade-in-up animate-delay-200 text-slate-700 dark:text-text-secondary"
        >
          A durable, replayable state layer designed for AI agents, serverless functions, and distributed workflows.
          <br />
          <span className="font-semibold text-slate-900 dark:text-white">No race conditions. No lost memory. No state drift.</span>
        </motion.p>

        <motion.div variants={fadeInUp} className="flex flex-col sm:flex-row gap-4 justify-center animate-fade-in-up animate-delay-300">
          <PrimaryCTA />
          <SecondaryCTA />
        </motion.div>
      </motion.div>
    </AnimatedBackgroundGrid>
  );
}