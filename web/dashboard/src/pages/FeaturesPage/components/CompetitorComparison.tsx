import React from "react";
import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Check, X, Minus } from "lucide-react";

interface CompetitorData {
  feature: string;
  functionfly: boolean | string;
  vercel: boolean | string;
  netlify: boolean | string;
  cloudflare: boolean | string;
  aws: boolean | string;
  category?: string;
}

const competitorData: CompetitorData[] = [
  {
    feature: "Multi-Provider Deployment",
    functionfly: true,
    vercel: false,
    netlify: false,
    cloudflare: false,
    aws: false,
    category: "Deployment"
  },
  {
    feature: "Automatic Failover",
    functionfly: true,
    vercel: false,
    netlify: false,
    cloudflare: false,
    aws: false,
    category: "Reliability"
  },
  {
    feature: "Sub-100ms Failover",
    functionfly: true,
    vercel: false,
    netlify: false,
    cloudflare: false,
    aws: false,
    category: "Performance"
  },
  {
    feature: "Predictive Routing",
    functionfly: true,
    vercel: false,
    netlify: false,
    cloudflare: false,
    aws: false,
    category: "Intelligence"
  },
  {
    feature: "Global Edge Network",
    functionfly: "200+ locations",
    vercel: "30+ locations",
    netlify: "200+ locations",
    cloudflare: "200+ locations",
    aws: "25+ regions",
    category: "Infrastructure"
  },
  {
    feature: "Serverless Functions",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: true,
    aws: true,
    category: "Core Features"
  },
  {
    feature: "Custom Domains",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: true,
    aws: true,
    category: "Configuration"
  },
  {
    feature: "Team Collaboration",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: false,
    aws: true,
    category: "Collaboration"
  },
  {
    feature: "Advanced Analytics",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: false,
    aws: true,
    category: "Monitoring"
  },
  {
    feature: "CLI Tools",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: true,
    aws: true,
    category: "Developer Tools"
  },
  {
    feature: "Git Integration",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: false,
    aws: true,
    category: "Developer Tools"
  },
  {
    feature: "API Access",
    functionfly: true,
    vercel: true,
    netlify: true,
    cloudflare: true,
    aws: true,
    category: "Developer Tools"
  }
];

const competitors = [
  { name: "FunctionFly", key: "functionfly", color: "#6366f1", highlight: true },
  { name: "Vercel", key: "vercel", color: "#000000" },
  { name: "Netlify", key: "netlify", color: "#00C46A" },
  { name: "Cloudflare", key: "cloudflare", color: "#F38020" },
  { name: "AWS Lambda", key: "aws", color: "#FF9900" }
];

const categories = [...new Set(competitorData.map(item => item.category).filter(Boolean))];

function getFeatureValue(value: boolean | string) {
  if (typeof value === "boolean") {
    return value ? (
      <Check className="w-5 h-5 text-emerald-400 mx-auto" />
    ) : (
      <X className="w-5 h-5 text-red-400 mx-auto" />
    );
  }
  if (value === "limited" || value === "partial") {
    return <Minus className="w-5 h-5 text-yellow-400 mx-auto" />;
  }
  return <span className="text-sm text-text-secondary">{value}</span>;
}

export function CompetitorComparison() {
  return (
    <motion.section
      initial={{ opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
      className="py-20"
    >
      <div className="text-center mb-16">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <Badge variant="outline" className="mb-4 border-[#6366f1]/30 text-[#6366f1]">
            Why Choose FunctionFly
          </Badge>
          <h2 className="text-4xl font-bold text-white mb-4">
            How we compare to the competition
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto text-lg">
            FunctionFly isn't just another deployment platform. We're the only solution that
            offers true multi-provider deployment with intelligent failover and predictive routing.
          </p>
        </motion.div>
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.6, delay: 0.2 }}
      >
        <Card className="border-white/8 bg-white/5 overflow-hidden">
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full min-w-[800px]">
                <thead>
                  <tr className="border-b border-white/8">
                    <th className="text-left p-6 text-white font-semibold min-w-[200px]">
                      Features
                    </th>
                    {competitors.map((competitor) => (
                      <th
                        key={competitor.key}
                        className="text-center p-6 font-semibold min-w-[120px]"
                        style={{ color: competitor.highlight ? "#6366f1" : "white" }}
                      >
                        <div className="flex flex-col items-center gap-2">
                          <span className={competitor.highlight ? "text-[#6366f1]" : "text-white"}>
                            {competitor.name}
                          </span>
                          {competitor.highlight && (
                            <Badge className="bg-[#6366f1] text-white text-xs px-2 py-0.5">
                              Recommended
                            </Badge>
                          )}
                        </div>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {categories.map((category) => (
                    <React.Fragment key={category}>
                      {/* Category Header */}
                      <tr className="bg-white/2">
                        <td
                          colSpan={competitors.length + 1}
                          className="p-4 text-[#6366f1] font-semibold text-sm uppercase tracking-wide"
                        >
                          {category}
                        </td>
                      </tr>
                      {/* Category Features */}
                      {competitorData
                        .filter((item) => item.category === category)
                        .map((item, index) => (
                          <motion.tr
                            key={item.feature}
                            className={`border-b border-white/4 ${
                              index % 2 === 0 ? "bg-white/1" : ""
                            }`}
                            initial={{ opacity: 0, x: -20 }}
                            whileInView={{ opacity: 1, x: 0 }}
                            viewport={{ once: true }}
                            transition={{ duration: 0.5, delay: index * 0.05 }}
                          >
                            <td className="p-6 text-white font-medium">
                              {item.feature}
                            </td>
                            {competitors.map((competitor) => {
                              const value = item[competitor.key as keyof CompetitorData] as boolean | string;
                              const isHighlighted = competitor.highlight && value === true;

                              return (
                                <td
                                  key={competitor.key}
                                  className={`p-6 text-center ${
                                    isHighlighted ? "bg-[#6366f1]/10" : ""
                                  }`}
                                >
                                  {getFeatureValue(value)}
                                </td>
                              );
                            })}
                          </motion.tr>
                        ))}
                    </React.Fragment>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* Key Differentiators */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.6, delay: 0.4 }}
        className="mt-16 grid md:grid-cols-3 gap-6"
      >
        <Card className="border-[#6366f1]/20 bg-[#6366f1]/5">
          <CardContent className="p-6 text-center">
            <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-[#6366f1]/20 flex items-center justify-center">
              <Check className="w-6 h-6 text-[#6366f1]" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">
              Multi-Provider Magic
            </h3>
            <p className="text-text-secondary text-sm">
              Deploy to multiple providers simultaneously. No vendor lock-in, maximum reliability.
            </p>
          </CardContent>
        </Card>

        <Card className="border-[#6366f1]/20 bg-[#6366f1]/5">
          <CardContent className="p-6 text-center">
            <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-[#6366f1]/20 flex items-center justify-center">
              <Check className="w-6 h-6 text-[#6366f1]" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">
              Intelligent Failover
            </h3>
            <p className="text-text-secondary text-sm">
              Sub-100ms automatic failover ensures your users never experience downtime.
            </p>
          </CardContent>
        </Card>

        <Card className="border-[#6366f1]/20 bg-[#6366f1]/5">
          <CardContent className="p-6 text-center">
            <div className="w-12 h-12 mx-auto mb-4 rounded-xl bg-[#6366f1]/20 flex items-center justify-center">
              <Check className="w-6 h-6 text-[#6366f1]" />
            </div>
            <h3 className="text-lg font-semibold text-white mb-2">
              Predictive Routing
            </h3>
            <p className="text-text-secondary text-sm">
              AI-powered routing predicts and prevents issues before they impact users.
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </motion.section>
  );
}