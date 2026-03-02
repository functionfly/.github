import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useSpringCarousel } from "react-spring-carousel";
import { useScrollAnimation } from "../hooks";
import toast from "react-hot-toast";

// Feature Carousel Component
export function FeatureCarousel() {
  const { ref, inView } = useScrollAnimation(0.1, false);

  const features = [
    {
      title: "Multi-Provider Deployment",
      description: "Deploy to Vercel, Netlify, Fly.io, and more with a single command. No vendor lock-in.",
      icon: "🚀",
      color: "from-[#6366f1] to-[#8b5cf6]"
    },
    {
      title: "Sub-Millisecond Failover",
      description: "Automatic switching between providers ensures zero downtime for your users.",
      icon: "⚡",
      color: "from-[#06b6d4] to-[#0891b2]"
    },
    {
      title: "Real-time Analytics",
      description: "Monitor performance across all providers with detailed insights and alerts.",
      icon: "📊",
      color: "from-[#10b981] to-[#059669]"
    },
    {
      title: "Enterprise Security",
      description: "Bank-level encryption, compliance certifications, and advanced access controls.",
      icon: "🔒",
      color: "from-[#f59e0b] to-[#d97706]"
    },
    {
      title: "Developer Experience",
      description: "Intuitive CLI, comprehensive APIs, and excellent documentation for seamless integration.",
      icon: "👨‍💻",
      color: "from-[#ef4444] to-[#dc2626]"
    },
    {
      title: "Global CDN",
      description: "Automatic edge deployment ensures your functions run close to your users worldwide.",
      icon: "🌍",
      color: "from-[#8b5cf6] to-[#7c3aed]"
    }
  ];

  const { carouselFragment, slideToPrevItem, slideToNextItem, useListenToCustomEvent } = useSpringCarousel({
    itemsPerSlide: 1,
    withLoop: true,
    items: features.map((feature, index) => ({
      id: `feature-${index}`,
      renderItem: (
        <div className="flex items-center justify-center p-8 h-full">
          <Card className="pricing-feature-carousel-card w-full max-w-md border-white/10 bg-white/5 backdrop-blur-sm">
            <CardContent className="pt-8 pb-8">
              <motion.div
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.6 }}
                className="text-center flex flex-col items-center justify-center"
              >
                <div className={`w-20 h-20 mb-6 rounded-2xl bg-linear-to-br ${feature.color} border border-white/20 flex items-center justify-center text-4xl`}>
                  {feature.icon}
                </div>
                <h3 className="text-2xl font-bold text-white mb-4">{feature.title}</h3>
                <p className="text-text-secondary leading-relaxed">{feature.description}</p>
              </motion.div>
            </CardContent>
          </Card>
        </div>
      ),
    })),
  });

  useListenToCustomEvent((event) => {
    if (event.eventName === "onSlideStartChange") {
      toast(`Feature: ${features[event.nextItem.id.split('-')[1]].title}`, {
        duration: 1500,
        style: {
          background: '#1a1a1a',
          color: '#fff',
          border: '1px solid #6366f1',
        },
      });
    }
  });

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.8, ease: "easeOut" }}
      className="pricing-feature-carousel mb-20"
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.6, delay: 0.2 }}
        className="text-center mb-12"
      >
        <h2 className="text-3xl font-bold text-white mb-4">Powerful Features</h2>
        <p className="text-text-secondary max-w-2xl mx-auto">
          Discover what makes FunctionFly the most reliable serverless platform
        </p>
      </motion.div>

      <div className="relative max-w-4xl mx-auto">
        <div className="overflow-hidden rounded-2xl bg-white/5 border border-white/8">
          {carouselFragment}
        </div>

        <div className="flex justify-center gap-4 mt-6">
          <Button
            variant="outline"
            size="sm"
            onClick={slideToPrevItem}
            className="pricing-carousel-nav-btn border-white/20 hover:bg-white/10 text-text-primary"
            aria-label="Previous feature"
          >
            ← Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={slideToNextItem}
            className="pricing-carousel-nav-btn border-white/20 hover:bg-white/10 text-text-primary"
            aria-label="Next feature"
          >
            Next →
          </Button>
        </div>
      </div>
    </motion.div>
  );
}
