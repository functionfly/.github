import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

const CTASection = () => {
  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6 }}
      className="py-24 md:py-32 text-center relative"
    >
      <div className="max-w-4xl mx-auto px-6">
        <div className="mb-6">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-[#6366f1]/20 mb-6">
            <div className="w-2 h-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] animate-pulse" />
            <span className="text-sm font-medium text-[#6366f1]">Ready to get started?</span>
          </div>
        </div>

        <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold text-white mb-6 leading-tight">
          <span className="bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent">
            Ready to
          </span>
          <br />
          <span className="bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] bg-clip-text text-transparent">
            integrate?
          </span>
        </h2>

        <p className="text-text-secondary text-xl md:text-2xl mb-12 max-w-3xl mx-auto leading-relaxed font-light">
          Connect FunctionFly with your existing tools and platforms.
          Start building with our comprehensive integration ecosystem today.
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link to="/signup">
            <Button
              size="lg"
              className="bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90 text-white font-semibold px-10 py-4 text-lg shadow-lg shadow-[#6366f1]/25 hover:shadow-[#6366f1]/40 transition-all duration-300 transform hover:scale-105"
            >
              Start Building
            </Button>
          </Link>
          <Link to="/docs/integrations">
            <Button
              size="lg"
              variant="outline"
              className="border-white/30 hover:border-white/50 hover:bg-white/10 text-white font-semibold px-10 py-4 text-lg backdrop-blur-sm transition-all duration-300"
            >
              Integration Docs
            </Button>
          </Link>
        </div>
      </div>
    </motion.section>
  );
};

export default CTASection;