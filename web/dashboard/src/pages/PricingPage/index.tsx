import { motion } from "framer-motion";
import { ArrowLeft, Check, Star, CreditCard, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { PLANS } from "@/lib/constants";
import { Link } from "react-router-dom";
import { Footer } from "@/pages/LandingPage/components";
import { useState } from "react";
import { Tooltip } from "react-tooltip";
import Confetti from "react-confetti";
import { useWindowSize } from "react-use";
import toast, { Toaster } from "react-hot-toast";
import { useCardGestures, useScrollAnimation } from "./hooks";
import { ComparisonSection } from "./components/ComparisonSection";
import { StateFabricPricingSection } from "./components/StateFabricPricingSection";
import { AgentPricingSection } from "./components/AgentPricingSection";
import { WhyChooseUsSection } from "./components/WhyChooseUsSection";
import { FAQSection } from "./components/FAQSection";
import { CTASection } from "./components/CTASection";
import { FeatureCarousel } from "./components/FeatureCarousel";
import { MetaTags } from "@/components/seo/MetaTags";
import { PricingPageStructuredData } from "@/components/seo/StructuredData";

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(" ");
}

export function PricingPage() {
  const { width, height } = useWindowSize();
  const [showConfetti, setShowConfetti] = useState(false);

  const handlePlanSelect = (planId: string) => {
    setShowConfetti(true);
    // Hide confetti after 3 seconds
    setTimeout(() => setShowConfetti(false), 3000);

    // Show success toast
    const planNames: { [key: string]: string } = {
      free: "Free Plan",
      starter: "Starter Plan",
      professional: "Professional Plan",
      enterprise: "Enterprise Plan"
    };

    toast.success(
      `Great choice! You're about to start with the ${planNames[planId] || planId}.`,
      {
        duration: 4000,
        style: {
          background: '#1a1a1a',
          color: '#fff',
          border: '1px solid #6366f1',
        },
        icon: '🚀',
      }
    );
  };

  return (
    <>
      {showConfetti && (
        <Confetti
          width={width}
          height={height}
          recycle={false}
          numberOfPieces={200}
          gravity={0.3}
          colors={['#6366f1', '#8b5cf6', '#06b6d4', '#10b981', '#f59e0b']}
        />
      )}

      {/* SEO Meta Tags */}
      <MetaTags
        title="Pricing | FunctionFly - Serverless Function Deployment Pricing"
        description="Simple, transparent pricing for serverless function deployment. Start free, scale as you grow. 14-day free trial on all paid plans. No hidden fees, no surprise charges."
        keywords={["serverless pricing", "function deployment pricing", "serverless plans", "function as a service pricing", "cloud function pricing", "serverless hosting pricing"]}
        url="/pricing"
      />

      {/* Structured Data */}
      <PricingPageStructuredData />

      <div className="pricing-page min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative overflow-hidden">
        {/* Background decorative elements */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_30%,rgba(99,102,241,0.08),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_80%_70%,rgba(139,92,246,0.08),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(6,182,212,0.05),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[linear-gradient(45deg,transparent_25%,rgba(255,255,255,0.005)_25%,rgba(255,255,255,0.005)_50%,transparent_50%,transparent_75%,rgba(255,255,255,0.005)_75%)] bg-[length:20px_20px] opacity-30 pointer-events-none" />

        {/* Navigation Bar */}
        <nav className="border-b border-white/10 bg-black/30 backdrop-blur-md sticky top-0 z-50 relative overflow-hidden">
          {/* Background gradient overlay */}
          <div className="absolute inset-0 bg-gradient-to-r from-[#6366f1]/5 via-transparent to-[#8b5cf6]/5" />
          <div className="relative max-w-7xl mx-auto px-4 lg:px-6">
            <div className="flex items-center justify-between h-16">
              <Link to="/" className="flex items-center gap-2 text-white hover:text-[#6366f1] transition-all duration-300 group">
                <div className="p-1 rounded-lg bg-white/5 group-hover:bg-[#6366f1]/10 transition-colors">
                  <ArrowLeft className="w-4 h-4" />
                </div>
                <span className="font-medium">Back to Home</span>
              </Link>
              <div className="flex items-center gap-3">
                <div className="w-2 h-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] animate-pulse" />
                <h1 className="text-xl font-bold bg-gradient-to-r from-white to-text-secondary bg-clip-text text-transparent">
                  Pricing
                </h1>
              </div>
              <div className="w-24" /> {/* Spacer for centering */}
            </div>
          </div>
        </nav>

        <div className="relative z-10 max-w-7xl mx-auto px-4 lg:px-6 py-8">
          {/* Page Header */}
          <div className="text-center py-20 md:py-24">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              <div className="relative inline-block mb-8">
                <div className="w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br from-[#6366f1]/30 via-[#8b5cf6]/20 to-[#06b6d4]/20 border border-[#6366f1]/30 flex items-center justify-center backdrop-blur-sm">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                    <CreditCard className="w-8 h-8 text-white" />
                  </div>
                </div>
                <div className="absolute -inset-4 bg-gradient-to-r from-[#6366f1]/20 via-[#8b5cf6]/10 to-[#06b6d4]/20 rounded-full blur-xl -z-10" />
              </div>

              <h1 className="text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-6 leading-tight">
                <span className="bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent">
                  Simple, transparent
                </span>
                <br />
                <span className="bg-gradient-to-r from-[#6366f1] via-[#8b5cf6] to-[#06b6d4] bg-clip-text text-transparent">
                  pricing
                </span>
              </h1>

              <p className="text-text-secondary max-w-3xl mx-auto text-xl md:text-2xl mb-10 leading-relaxed font-light">
                Start free, scale as you grow. No hidden fees, no surprise charges.
                <br className="hidden md:block" />
                Choose the plan that fits your needs.
              </p>

              <div className="flex items-center justify-center gap-3 mb-4">
                <div className="flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-green-500/10 to-emerald-500/10 border border-green-500/20">
                  <Shield className="w-4 h-4 text-green-400" />
                  <span className="text-sm font-medium text-green-400">14-day free trial on all paid plans</span>
                </div>
              </div>

              <div className="flex items-center justify-center gap-2 text-sm text-text-secondary">
                <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
                <span>No setup fees • Cancel anytime • Enterprise support available</span>
              </div>
            </motion.div>
          </div>

          {/* Pricing Cards */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            className="relative mb-24"
          >
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent blur-3xl -mx-8 rounded-3xl" />
            <div className="relative grid md:grid-cols-2 lg:grid-cols-4 gap-8 max-w-7xl mx-auto">
            {Object.values(PLANS).map((plan, index) => {
              const { ref, inView } = useScrollAnimation(0.2, false);
              const gestures = useCardGestures(plan.name);

              return (
                <motion.div
                  key={plan.id}
                  ref={ref}
                  {...gestures.bind()}
                  initial={{ opacity: 0, y: 30, scale: 0.95 }}
                  animate={inView ? { opacity: 1, y: 0, scale: 1 } : { opacity: 0, y: 30, scale: 0.95 }}
                  transition={{
                    duration: 0.6,
                    delay: index * 0.1,
                    ease: [0.25, 0.46, 0.45, 0.94]
                  }}
                  style={gestures.style}
                  className="transition-shadow duration-300"
                  onMouseEnter={() => setTimeout(() => {}, 0)} // Force re-render for hover effects
                  onMouseLeave={() => setTimeout(() => {}, 0)}
                >
                  <Card
                    className={cn(
                      "pricing-plan-card h-full relative overflow-hidden transition-all duration-300 group cursor-pointer",
                      "bg-gradient-to-br from-white/5 to-white/10 backdrop-blur-sm",
                      "border border-white/10 hover:border-white/20",
                      "hover:shadow-2xl hover:shadow-[#6366f1]/10",
                      plan.id === "professional" &&
                        "border-[#6366f1]/50 ring-1 ring-[#6366f1]/20 hover:ring-[#6366f1]/40"
                    )}
                  >
                    {/* Card background gradient overlay */}
                    <div className="absolute inset-0 bg-gradient-to-br from-transparent via-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                    {plan.id === "professional" && (
                      <div className="absolute -top-4 left-1/2 -translate-x-1/2 z-10">
                        <motion.div
                          initial={{ scale: 0, opacity: 0 }}
                          animate={{ scale: 1, opacity: 1 }}
                          transition={{ delay: 0.5 + index * 0.1 }}
                          className="relative"
                        >
                          <span className="px-4 py-2 rounded-full bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] text-white text-sm font-semibold flex items-center gap-2 shadow-lg shadow-[#6366f1]/25">
                            <Star className="w-4 h-4 fill-current animate-pulse" />
                            Most Popular
                          </span>
                          <div className="absolute inset-0 bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] rounded-full blur-lg opacity-50 -z-10" />
                        </motion.div>
                      </div>
                    )}

                    <CardContent className="p-8 relative z-10">
                      <div className="mb-8">
                        <motion.div
                          initial={{ opacity: 0, y: 10 }}
                          animate={{ opacity: 1, y: 0 }}
                          transition={{ delay: 0.2 + index * 0.1 }}
                        >
                          <h3 className="text-2xl font-bold text-white mb-3 group-hover:text-[#6366f1] transition-colors">
                            {plan.name}
                          </h3>
                          <p className="text-text-secondary text-base mb-6 leading-relaxed">
                            {plan.description}
                          </p>
                        </motion.div>

                        <motion.div
                          initial={{ opacity: 0, scale: 0.8 }}
                          animate={{ opacity: 1, scale: 1 }}
                          transition={{ delay: 0.3 + index * 0.1 }}
                          className="mb-4"
                        >
                          <div className="flex items-baseline gap-2">
                            <span className="text-5xl md:text-6xl font-bold bg-gradient-to-r from-white to-text-secondary bg-clip-text text-transparent">
                              {plan.price === "Custom"
                                ? "Custom"
                                : `$${plan.price}`}
                            </span>
                            {plan.price !== "Custom" && (
                              <span className="text-text-secondary text-lg">/month</span>
                            )}
                          </div>
                          {plan.id !== "free" && plan.id !== "enterprise" && (
                            <div className="flex items-center gap-2 mt-2">
                              <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
                              <p className="text-green-400 text-sm font-medium">
                                Billed monthly, cancel anytime
                              </p>
                            </div>
                          )}
                        </motion.div>
                      </div>

                      <motion.ul
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ delay: 0.4 + index * 0.1 }}
                        className="space-y-4 mb-8"
                      >
                        {plan.features.map((feature, featureIndex) => {
                          const tooltipId = `${plan.id}-feature-${featureIndex}`;
                          const getTooltipContent = (feature: string) => {
                            const tooltips: { [key: string]: string } = {
                              "1 function": "Deploy a single serverless function",
                              "5 functions": "Deploy up to 5 serverless functions",
                              "25 functions": "Deploy up to 25 serverless functions",
                              "Unlimited functions": "Deploy unlimited serverless functions",
                              "2 providers": "Deploy to Vercel or Netlify",
                              "3 providers": "Deploy to Vercel, Netlify, or Fly.io",
                              "5 providers": "Deploy to all supported providers",
                              "All providers": "Deploy to all supported providers",
                              "100K requests/month": "100,000 function invocations per month",
                              "1M requests/month": "1 million function invocations per month",
                              "10M requests/month": "10 million function invocations per month",
                              "Unlimited requests": "Unlimited function invocations",
                              "1 custom domain": "Connect one custom domain",
                              "5 custom domains": "Connect up to 5 custom domains",
                              "Unlimited custom domains": "Connect unlimited custom domains",
                              "Email support": "Get help via email during business hours",
                              "Priority support": "24/7 priority email and chat support",
                              "Dedicated support": "Dedicated account manager and phone support",
                              "Basic analytics": "View basic usage metrics and logs",
                              "Advanced analytics": "Detailed analytics with custom dashboards",
                              "Custom analytics": "White-labeled analytics with custom integrations",
                              "Team collaboration": "Invite team members to collaborate",
                            };
                            return tooltips[feature] || "";
                          };

                          return (
                            <motion.li
                              key={feature}
                              initial={{ opacity: 0, x: -10 }}
                              animate={{ opacity: 1, x: 0 }}
                              transition={{ delay: 0.5 + index * 0.1 + featureIndex * 0.05 }}
                              className="flex items-start gap-4 group"
                              data-tooltip-id={tooltipId}
                              data-tooltip-content={getTooltipContent(feature)}
                            >
                              <div className="w-6 h-6 rounded-full bg-gradient-to-br from-emerald-500/20 to-green-500/20 border border-emerald-500/30 flex items-center justify-center mt-0.5 group-hover:scale-110 transition-transform duration-200">
                                <Check className="w-3.5 h-3.5 text-emerald-400" />
                              </div>
                              <span className="text-text-secondary group-hover:text-white transition-colors cursor-help text-base leading-relaxed">
                                {feature}
                              </span>
                            </motion.li>
                          );
                        })}
                      </motion.ul>

                      <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.6 + index * 0.1 }}
                      >
                        <Link
                          to={plan.id === "enterprise" ? "/contact" : "/signup"}
                          className="block group"
                          onClick={() => handlePlanSelect(plan.id)}
                        >
                          <Button
                            variant={plan.id === "free" ? "outline" : "default"}
                            size="lg"
                            className={cn(
                              "w-full py-4 text-base font-semibold transition-all duration-300 transform hover:scale-105",
                              plan.id === "free" && "border-2 border-white/30 hover:border-white/50 hover:bg-white/10",
                              plan.id === "professional" && "bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90 shadow-lg shadow-[#6366f1]/25 hover:shadow-[#6366f1]/40",
                              plan.id !== "free" && plan.id !== "professional" && "bg-gradient-to-r from-white/10 to-white/5 hover:from-white/20 hover:to-white/10 border border-white/20"
                            )}
                          >
                            {plan.id === "enterprise"
                              ? "Contact Sales"
                              : plan.id === "free"
                                ? "Start Free"
                                : "Start Free Trial"}
                          </Button>
                        </Link>
                      </motion.div>
                    </CardContent>
                  </Card>
                </motion.div>
              );
            })}
            </div>
          </motion.div>

          {/* State Fabric pricing – separate section for easy management */}
          <StateFabricPricingSection />

          {/* Agent Execution Plans – for AI agent infrastructure */}
          <AgentPricingSection />

          {/* Additional Sections */}
          <FeatureCarousel />
          <WhyChooseUsSection />
          <ComparisonSection onPlanSelect={handlePlanSelect} />
          <FAQSection />
          <CTASection onPlanSelect={handlePlanSelect} />
        </div>

        <Footer />

        {/* Global Tooltip */}
        <Tooltip
          place="top"
          className="!bg-black !text-white !border !border-white/20 !rounded-lg !text-sm !max-w-xs"
          clickable={false}
          noArrow={false}
          offset={10}
        />

        {/* Toast Notifications */}
        <Toaster
          position="bottom-right"
          toastOptions={{
            style: {
              background: '#1a1a1a',
              color: '#fff',
              border: '1px solid #6366f1',
            },
          }}
        />
      </div>
    </>
  );
}
