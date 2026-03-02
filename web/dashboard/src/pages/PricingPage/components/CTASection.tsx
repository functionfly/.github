import { motion } from "framer-motion";
import { HeadphonesIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router-dom";
import { useScrollAnimation } from "../hooks";

interface CTASectionProps {
  onPlanSelect?: (planId: string) => void;
}

// Call to Action Section with scroll animations
export function CTASection({ onPlanSelect }: CTASectionProps) {
  const { ref, inView } = useScrollAnimation(0.1, false);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 50, scale: 0.9 }}
      animate={inView ? { opacity: 1, y: 0, scale: 1 } : { opacity: 0, y: 50, scale: 0.9 }}
      transition={{
        duration: 0.8,
        ease: [0.25, 0.46, 0.45, 0.94]
      }}
      className="text-center pb-20 relative"
    >
      {/* Background gradient */}
      <div className="absolute inset-0 bg-gradient-to-r from-[#6366f1]/10 via-transparent to-[#8b5cf6]/10 blur-3xl -mx-8 rounded-3xl" />

      <div className="relative max-w-4xl mx-auto">
        <Card className="pricing-cta-section border border-white/20 bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl shadow-2xl shadow-[#6366f1]/10 overflow-hidden">
          {/* Card background gradient overlay */}
          <div className="absolute inset-0 bg-gradient-to-br from-[#6366f1]/5 via-transparent to-[#8b5cf6]/5" />

          <CardContent className="p-12 relative z-10">
            <motion.div
              initial={{ scale: 0, rotate: -180 }}
              animate={inView ? { scale: 1, rotate: 0 } : { scale: 0, rotate: -180 }}
              transition={{ duration: 0.8, delay: 0.3, type: "spring", stiffness: 200 }}
              className="relative mb-8"
            >
              <div className="w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br from-[#6366f1]/30 to-[#8b5cf6]/20 border border-[#6366f1]/30 flex items-center justify-center backdrop-blur-sm">
                <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                  <HeadphonesIcon className="w-9 h-9 text-white" />
                </div>
              </div>
              <div className="absolute inset-0 bg-gradient-to-br from-[#6366f1]/20 to-[#8b5cf6]/10 rounded-3xl blur-xl opacity-60 -z-10" />
            </motion.div>

            <motion.h3
              className="pricing-cta-heading text-4xl md:text-5xl font-bold text-white mb-6"
              initial={{ opacity: 0, y: 20 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
              transition={{ duration: 0.6, delay: 0.5 }}
            >
              <span className="pricing-cta-heading-line1 bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent">
                Need help choosing
              </span>
              <br />
              <span className="pricing-cta-heading-line2 bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] bg-clip-text text-transparent">
                a plan?
              </span>
            </motion.h3>

            <motion.p
              className="text-text-secondary text-xl mb-10 max-w-2xl mx-auto leading-relaxed"
              initial={{ opacity: 0, y: 20 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
              transition={{ duration: 0.6, delay: 0.6 }}
            >
              Our team is here to help you find the perfect plan for your needs.
              Schedule a call or chat with us to discuss your requirements.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row gap-6 justify-center"
              initial={{ opacity: 0, y: 20 }}
              animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
              transition={{ duration: 0.6, delay: 0.7 }}
            >
              <Link to="/contact" onClick={() => onPlanSelect?.("enterprise")}>
                <Button
                  size="lg"
                  className="bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90 text-white font-semibold px-8 py-4 shadow-lg shadow-[#6366f1]/25 hover:shadow-[#6366f1]/40 transition-all duration-300 transform hover:scale-105"
                >
                  Contact Sales
                </Button>
              </Link>
              <Link to="/docs">
                <Button
                  size="lg"
                  variant="outline"
                  className="pricing-cta-docs-btn border-white/30 hover:border-white/50 hover:bg-white/10 text-white font-semibold px-8 py-4 backdrop-blur-sm transition-all duration-300"
                >
                  View Documentation
                </Button>
              </Link>
            </motion.div>
          </CardContent>
        </Card>
      </div>
    </motion.div>
  );
}
