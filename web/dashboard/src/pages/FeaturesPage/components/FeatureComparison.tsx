import { motion } from "framer-motion";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Link } from "react-router-dom";
import { Check, X, Star, Zap } from "lucide-react";

interface PlanFeature {
  name: string;
  free: boolean | string;
  pro: boolean | string;
  description?: string;
}

const planFeatures: PlanFeature[] = [
  {
    name: "Deployments per month",
    free: "100",
    pro: "Unlimited",
    description: "Number of function deployments"
  },
  {
    name: "Multi-provider deployment",
    free: false,
    pro: true,
    description: "Deploy to multiple edge providers simultaneously"
  },
  {
    name: "Global edge locations",
    free: "5 regions",
    pro: "200+ locations",
    description: "Number of edge locations available"
  },
  {
    name: "Predictive routing (AI)",
    free: false,
    pro: true,
    description: "AI-powered traffic optimization and issue prevention"
  },
  {
    name: "Custom domains",
    free: false,
    pro: true,
    description: "Use your own domain names"
  },
  {
    name: "Team members",
    free: "3",
    pro: "Unlimited",
    description: "Number of team members you can invite"
  },
  {
    name: "Advanced analytics",
    free: "Basic metrics",
    pro: "Full analytics suite",
    description: "Dashboard and reporting capabilities"
  },
  {
    name: "Priority support",
    free: false,
    pro: true,
    description: "24/7 priority customer support"
  },
  {
    name: "SLA guarantee",
    free: "99.5%",
    pro: "99.9%",
    description: "Uptime service level agreement"
  },
  {
    name: "Environment variables",
    free: "10",
    pro: "Unlimited",
    description: "Number of environment variables"
  }
];

export function FeatureComparison() {
  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6 }}
      className="py-20 border-t border-white/8"
    >
      <div className="text-center mb-16">
        <h2 className="text-3xl font-bold text-white mb-4">
          Choose the right plan for your needs
        </h2>
        <p className="text-text-secondary max-w-2xl mx-auto">
          Compare our Free and Pro plans to see which features are right for your project.
          Start free and upgrade anytime.
        </p>
      </div>

      <div className="max-w-6xl mx-auto">
        {/* Plan Cards Header */}
        <div className="grid md:grid-cols-2 gap-8 mb-8">
          <Card className="border-white/10 bg-white/5">
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-xl text-white mb-2">Free Tier</CardTitle>
              <div className="text-3xl font-bold text-white mb-1">$0</div>
              <p className="text-sm text-text-secondary">Perfect for getting started</p>
            </CardHeader>
            <CardContent className="text-center">
              <Link to="/signup">
                <Button variant="outline" className="w-full">
                  Get Started Free
                </Button>
              </Link>
            </CardContent>
          </Card>

          <Card className="border-[#6366f1]/50 bg-linear-to-br from-[#6366f1]/10 to-[#8b5cf6]/10 relative">
            <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
              <Badge className="bg-[#6366f1] text-white px-3 py-1">
                <Star className="w-3 h-3 mr-1" />
                Most Popular
              </Badge>
            </div>
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-xl text-white mb-2">Pro Plan</CardTitle>
              <div className="text-3xl font-bold text-white mb-1">$29<span className="text-lg font-normal">/month</span></div>
              <p className="text-sm text-text-secondary">For growing teams and production apps</p>
            </CardHeader>
            <CardContent className="text-center">
              <Link to="/signup?plan=pro">
                <Button className="w-full bg-[#6366f1] hover:bg-[#6366f1]/90">
                  <Zap className="w-4 h-4 mr-2" />
                  Upgrade to Pro
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>

        {/* Feature Comparison Table */}
        <Card className="border-white/10">
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left p-6 text-white font-semibold">Feature</th>
                    <th className="text-center p-6 text-white font-semibold">Free</th>
                    <th className="text-center p-6 text-white font-semibold">Pro</th>
                  </tr>
                </thead>
                <tbody>
                  {planFeatures.map((feature, index) => (
                    <motion.tr
                      key={feature.name}
                      initial={{ opacity: 0, x: -20 }}
                      whileInView={{ opacity: 1, x: 0 }}
                      viewport={{ once: true }}
                      transition={{ duration: 0.5, delay: index * 0.05 }}
                      className="border-b border-white/5 hover:bg-white/5 transition-colors"
                    >
                      <td className="p-6">
                        <div>
                          <div className="font-medium text-white mb-1">{feature.name}</div>
                          {feature.description && (
                            <div className="text-sm text-text-secondary">{feature.description}</div>
                          )}
                        </div>
                      </td>
                      <td className="p-6 text-center">
                        {typeof feature.free === 'boolean' ? (
                          feature.free ? (
                            <Check className="w-5 h-5 text-green-500 mx-auto" />
                          ) : (
                            <X className="w-5 h-5 text-red-500 mx-auto" />
                          )
                        ) : (
                          <span className="text-white font-medium">{feature.free}</span>
                        )}
                      </td>
                      <td className="p-6 text-center">
                        {typeof feature.pro === 'boolean' ? (
                          feature.pro ? (
                            <Check className="w-5 h-5 text-green-500 mx-auto" />
                          ) : (
                            <X className="w-5 h-5 text-red-500 mx-auto" />
                          )
                        ) : (
                          <span className="text-white font-medium">{feature.pro}</span>
                        )}
                      </td>
                    </motion.tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>

        {/* Upgrade CTA */}
        <div className="text-center mt-12">
          <div className="max-w-md mx-auto">
            <h3 className="text-xl font-semibold text-white mb-4">
              Ready to unlock the full potential?
            </h3>
            <p className="text-text-secondary mb-6">
              Upgrade to Pro and get access to all features, unlimited usage, and priority support.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link to="/pricing">
                <Button variant="outline">
                  Compare All Plans
                </Button>
              </Link>
              <Link to="/signup?plan=pro">
                <Button className="bg-[#6366f1] hover:bg-[#6366f1]/90">
                  Start Pro Trial
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </motion.section>
  );
}