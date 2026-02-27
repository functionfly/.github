import { motion } from "framer-motion";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

const fadeInUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.6 }
};

export function PrimaryCTA() {
  return (
    <Button size="lg" className="hero-primary-button gap-2 glow">
      Explore the Architecture
      <ArrowRight className="w-4 h-4" />
    </Button>
  );
}