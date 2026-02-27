import { motion } from "framer-motion";
import { Zap, Shield, Users } from "lucide-react";
import { useScrollAnimation } from "../hooks";

// Why Choose Us Section with scroll animations
export function WhyChooseUsSection() {
  const { ref, inView } = useScrollAnimation(0.1, false);

  const features = [
    {
      icon: Zap,
      title: "Lightning Fast",
      description: "Sub-millisecond failover switching ensures your users never experience downtime."
    },
    {
      icon: Shield,
      title: "99.99% Uptime",
      description: "Enterprise-grade reliability with automatic recovery and comprehensive monitoring."
    },
    {
      icon: Users,
      title: "Developer First",
      description: "Simple setup, intuitive dashboard, and APIs designed by developers for developers."
    }
  ];

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.8, ease: "easeOut" }}
      className="mb-24 relative"
    >
      {/* Section background */}
      <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent blur-3xl -mx-8 rounded-3xl" />

      <div className="relative">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-[#6366f1]/10 to-[#8b5cf6]/10 border border-[#6366f1]/20 mb-6">
            <div className="w-2 h-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] animate-pulse" />
            <span className="text-sm font-medium text-[#6366f1]">Why Choose Us</span>
          </div>

          <h2 className="text-4xl md:text-5xl font-bold text-white mb-6">
            <span className="bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent">
              Why Choose
            </span>
            <br />
            <span className="bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] bg-clip-text text-transparent">
              FunctionFly?
            </span>
          </h2>

          <p className="text-text-secondary max-w-3xl mx-auto text-xl leading-relaxed">
            Built by engineers who've been through serverless outages. We know what matters most for your production applications.
          </p>
        </motion.div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        {features.map((feature, index) => {
          const colors = [
            { bg: 'from-[#6366f1]/20 to-[#8b5cf6]/10', border: 'border-[#6366f1]/30', icon: 'text-[#6366f1]' },
            { bg: 'from-[#10b981]/20 to-[#06b6d4]/10', border: 'border-[#10b981]/30', icon: 'text-[#10b981]' },
            { bg: 'from-[#f59e0b]/20 to-[#ef4444]/10', border: 'border-[#f59e0b]/30', icon: 'text-[#f59e0b]' },
          ];

          const colorScheme = colors[index % colors.length];

          return (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 30, scale: 0.9 }}
              animate={inView ? { opacity: 1, y: 0, scale: 1 } : { opacity: 0, y: 30, scale: 0.9 }}
              transition={{
                duration: 0.6,
                delay: 0.4 + index * 0.15,
                ease: [0.25, 0.46, 0.45, 0.94]
              }}
              className="text-center group"
            >
              <div className="relative mb-6">
                <motion.div
                  className={`w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br ${colorScheme.bg} border ${colorScheme.border} flex items-center justify-center backdrop-blur-sm shadow-lg`}
                  whileHover={{ scale: 1.15, rotate: 5 }}
                  transition={{ type: "spring", stiffness: 300 }}
                >
                  <div className={`w-16 h-16 rounded-2xl bg-gradient-to-br ${colorScheme.bg} flex items-center justify-center`}>
                    <feature.icon className={`w-8 h-8 ${colorScheme.icon}`} />
                  </div>
                </motion.div>
                <div className={`absolute inset-0 bg-gradient-to-br ${colorScheme.bg} rounded-3xl blur-xl opacity-50 -z-10 group-hover:opacity-75 transition-opacity duration-300`} />
              </div>

              <motion.h3
                className="text-2xl font-bold text-white mb-4 group-hover:text-white/90 transition-colors"
                initial={{ opacity: 0 }}
                animate={inView ? { opacity: 1 } : { opacity: 0 }}
                transition={{ delay: 0.6 + index * 0.15 }}
              >
                {feature.title}
              </motion.h3>

              <motion.p
                className="text-text-secondary text-base leading-relaxed group-hover:text-text-secondary/90 transition-colors"
                initial={{ opacity: 0 }}
                animate={inView ? { opacity: 1 } : { opacity: 0 }}
                transition={{ delay: 0.7 + index * 0.15 }}
              >
                {feature.description}
              </motion.p>
            </motion.div>
          );
        })}
      </div>
      </div>
    </motion.div>
  );
}