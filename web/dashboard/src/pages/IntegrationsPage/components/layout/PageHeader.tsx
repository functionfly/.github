import { motion } from "framer-motion";
import { Puzzle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

const PageHeader = () => {
  return (
    <div className="text-center py-20 md:py-24">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <div className="relative inline-block mb-8">
          <div className="w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br from-[#6366f1]/30 via-[#8b5cf6]/20 to-[#6366f1]/30 border border-[#6366f1]/30 flex items-center justify-center backdrop-blur-sm">
            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
              <Puzzle className="w-8 h-8 text-white" />
            </div>
          </div>
          <div className="absolute -inset-4 bg-gradient-to-r from-[#6366f1]/20 via-[#8b5cf6]/10 to-[#6366f1]/20 rounded-full blur-xl -z-10" />
        </div>

        <h1 className="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-6 leading-tight">
          <span className="bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent">
            Integrations that
          </span>
          <br />
          <span className="bg-gradient-to-r from-[#6366f1] via-[#8b5cf6] to-[#6366f1] bg-clip-text text-transparent">
            power your workflow
          </span>
        </h1>

        <p className="text-text-secondary max-w-3xl mx-auto text-xl md:text-2xl mb-10 leading-relaxed font-light">
          Connect FunctionFly with your favorite platforms, frameworks, and services.
          <br className="hidden md:block" />
          Deploy once, integrate everywhere with our comprehensive integration ecosystem.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <Link to="/signup">
            <Button
              size="lg"
              className="bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90 text-white font-semibold px-8 py-3 shadow-lg shadow-[#6366f1]/25 hover:shadow-[#6366f1]/40 transition-all duration-300 transform hover:scale-105"
            >
              Start Building
            </Button>
          </Link>
          <Link to="/docs/integrations">
            <Button
              size="lg"
              variant="outline"
              className="border-white/20 hover:border-white/40 hover:bg-white/5 text-white font-semibold px-8 py-3 backdrop-blur-sm transition-all duration-300"
            >
              View Documentation
            </Button>
          </Link>
        </div>
      </motion.div>
    </div>
  );
};

export default PageHeader;