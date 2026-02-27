import { motion } from "framer-motion";
import { Zap } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

export const PageHeader = () => {
  return (
    <div className="features-header">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <div className="icon-container animate-float">
          <Zap className="w-8 h-8 text-[#6366f1]" />
        </div>
        <h1 className="text-4xl md:text-5xl font-bold text-white mb-4">
          Powerful features for modern developers
        </h1>
        <p className="text-text-secondary max-w-2xl mx-auto text-lg mb-8">
          Everything you need to deploy, scale, and monitor your serverless
          functions with enterprise-grade reliability and developer-friendly
          tools.
        </p>
        <div className="cta-buttons">
          <Link to="/pricing" className="cta-primary">
            View Pricing
          </Link>
          <Link to="/signup" className="cta-secondary">
            Start Free Trial
          </Link>
        </div>
      </motion.div>
    </div>
  );
};