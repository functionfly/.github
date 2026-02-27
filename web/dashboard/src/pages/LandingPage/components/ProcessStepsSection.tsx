import { motion } from "framer-motion";
import { Rocket, Activity, RefreshCw, CheckCircle2, AlertTriangle, TrendingUp, ArrowRight, Check } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

const processSteps = [
  {
    step: 1,
    icon: Rocket,
    title: "Deploy",
    description:
      "Deploy your functions to multiple cloud providers simultaneously with a single command.",
    details: "Zero-config multi-cloud deployment",
  },
  {
    step: 2,
    icon: Activity,
    title: "Monitor",
    description:
      "Real-time health monitoring across all providers with predictive failure detection.",
    details: "AI-powered health metrics",
  },
  {
    step: 3,
    icon: RefreshCw,
    title: "Failover",
    description:
      "Automatic traffic rerouting to healthy providers in under 200ms when issues are detected.",
    details: "Sub-second switching",
  },
  {
    step: 4,
    icon: CheckCircle2,
    title: "Recovery",
    description:
      "Seamless recovery with no user impact. Detailed incident reports for continuous improvement.",
    details: "Zero-downtime guarantee",
  },
];

const beforeAfterData = {
  before: {
    icon: AlertTriangle,
    title: "Current Reality",
    description:
      "Manual failover takes hours, customers notice outages, and recovery is stressful",
    points: [
      "Hours to detect and respond to outages",
      "Customers experience downtime",
      "Complex manual intervention required",
      "Lost revenue and trust",
    ],
  },
  after: {
    icon: TrendingUp,
    title: "With FunctionFly",
    description:
      "Automatic failover happens in seconds, users never notice, you focus on building",
    points: [
      "Sub-200ms automatic failover",
      "Zero user-facing downtime",
      "Set-and-forget reliability",
      "Focus on product development",
    ],
  },
};

export function ProcessStepsSection() {
  return (
    <section className="py-20 border-t border-white/8 mesh-gradient-bg process-steps-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold text-text-primary mb-4" style={{ color: 'var(--text-primary)', fontWeight: 800 }}>
            How FunctionFly works
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto" style={{ color: 'var(--text-secondary)' }}>
            Four simple steps to eliminate downtime worries forever.
          </p>
        </div>

        {/* Process Steps */}
        <div className="relative mb-20">
          {/* Connection Line */}
          <div className="hidden lg:block absolute top-24 left-0 right-0 h-0.5 bg-linear-to-r from-transparent via-[#6366f1]/30 to-transparent" />

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
            {processSteps.map((step, index) => (
              <motion.div
                key={step.step}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                className="relative"
                style={{ willChange: 'transform, opacity' }}
              >
                <Card className="h-full hover:border-[#6366f1]/30 transition-colors card-elevation glass-card shine-effect">
                  <CardContent className="p-6 text-center">
                    {/* Step Number */}
                    <div className="w-8 h-8 mx-auto mb-4 rounded-full bg-linear-to-r from-[#6366f1] to-[#8b5cf6] flex items-center justify-center text-white font-bold text-sm">
                      {step.step}
                    </div>

                    {/* Icon */}
                    <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-linear-to-br from-[#6366f1]/30 to-[#8b5cf6]/30 border border-[#6366f1]/40 flex items-center justify-center glow">
                      <step.icon className="w-8 h-8 text-white drop-shadow-lg" />
                    </div>

                    {/* Content */}
                    <h3 className="text-xl font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 700 }}>
                      {step.title}
                    </h3>
                    <p className="text-text-secondary mb-3">
                      {step.description}
                    </p>
                    <div className="text-xs text-[#6366f1] font-medium">
                      {step.details}
                    </div>
                  </CardContent>
                </Card>

                {/* Arrow for desktop */}
                {index < processSteps.length - 1 && (
                  <div className="hidden lg:block absolute top-24 -right-6 z-10">
                    <ArrowRight className="w-6 h-6 text-[#6366f1]/60" />
                  </div>
                )}
              </motion.div>
            ))}
          </div>
        </div>

        {/* Before/After Comparison */}
        <div className="grid md:grid-cols-2 gap-8">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <Card className="h-full border-red-500/20 bg-red-500/5 card-elevation glass-card shine-effect">
              <CardContent className="p-8">
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-12 h-12 rounded-xl bg-red-500/20 border border-red-500/20 flex items-center justify-center">
                    <beforeAfterData.before.icon className="w-6 h-6 text-red-400" />
                  </div>
                  <h3 className="text-xl font-semibold text-text-primary">
                    {beforeAfterData.before.title}
                  </h3>
                </div>

                <p className="text-text-secondary mb-6">
                  {beforeAfterData.before.description}
                </p>

                <ul className="space-y-3">
                  {beforeAfterData.before.points.map((point, index) => (
                    <li key={index} className="flex items-start gap-3">
                      <div className="w-5 h-5 rounded-full bg-red-500/20 flex items-center justify-center mt-0.5">
                        <div className="w-2 h-2 rounded-full bg-red-400" />
                      </div>
                      <span className="text-text-secondary text-sm">
                        {point}
                      </span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <Card className="h-full border-emerald-500/20 bg-emerald-500/5 card-elevation glass-card shine-effect">
              <CardContent className="p-8">
                <div className="flex items-center gap-3 mb-6">
                  <div className="w-12 h-12 rounded-xl bg-emerald-500/20 border border-emerald-500/20 flex items-center justify-center">
                    <beforeAfterData.after.icon className="w-6 h-6 text-emerald-400" />
                  </div>
                  <h3 className="text-xl font-semibold text-text-primary">
                    {beforeAfterData.after.title}
                  </h3>
                </div>

                <p className="text-text-secondary mb-6">
                  {beforeAfterData.after.description}
                </p>

                <ul className="space-y-3">
                  {beforeAfterData.after.points.map((point, index) => (
                    <li key={index} className="flex items-start gap-3">
                      <div className="w-5 h-5 rounded-full bg-emerald-500/20 flex items-center justify-center mt-0.5">
                        <Check className="w-3 h-3 text-emerald-400" />
                      </div>
                      <span className="text-text-secondary text-sm">
                        {point}
                      </span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </motion.div>
        </div>
      </div>
    </section>
  );
}